package pack

import (
	"fmt"
	"io"
	"path"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/qmuntal/gltf"
)

const (
	AttributePosition  = "POSITION"
	AttributeNormal    = "NORMAL"
	AttributeTangent   = "TANGENT"
	AttributeTexCoord0 = "TEXCOORD_0"
	AttributeColor0    = "COLOR_0"
)

var (
	emptyMatrix      = [16]float32{}
	emptyTranslation = [3]float32{}
	emptyRotation    = [4]float32{}
	emptyScale       = [3]float32{}
)

type OpenGLTFResourceAction struct {
	locator ResourceLocator
	uri     string
	model   *Model
}

func (a *OpenGLTFResourceAction) Describe() string {
	return fmt.Sprintf("open_gltf_resource(uri: %q)", a.uri)
}

func (a *OpenGLTFResourceAction) Model() *Model {
	if a.model == nil {
		panic("reading data from unprocessed action")
	}
	return a.model
}

func (a *OpenGLTFResourceAction) Run() error {
	rawGLTF, err := gltf.Open(a.uri)
	if err != nil {
		return fmt.Errorf("failed to parse gltf model %q: %w", a.uri, err)
	}
	gltfDoc := GLTFDocument{rawGLTF}

	a.model = &Model{}

	// build meshes
	meshMapping := make(map[uint32]*Mesh)

	for i := range gltfDoc.Meshes {
		gltfMesh := gltfDoc.FindMesh(i)
		mesh := &Mesh{
			Name:        gltfMesh.Name,
			SubMeshes:   make([]SubMesh, len(gltfMesh.Primitives)),
			VertexCount: 0,
			IndexCount:  0,
		}

		for j := range gltfMesh.Primitives {
			gltfPrimitive := gltfMesh.FindPrimitive(j)

			subMesh := SubMesh{}
			subMesh.IndexOffset = len(mesh.Indices)
			subMesh.IndexCount = int(gltfPrimitive.FindIndexCount())

			for k := 0; k < subMesh.IndexCount; k++ {
				gltfIndex := gltfPrimitive.FindIndex(k)
				coord := gltfPrimitive.FindCoord(gltfIndex)
				normal := gltfPrimitive.FindNormal(gltfIndex)
				tangent := gltfPrimitive.FindTangent(gltfIndex)
				texCoord := gltfPrimitive.FindTexCoord0(gltfIndex)
				color := gltfPrimitive.FindColor0(gltfIndex)

				// find same vertex
				matchingIndex := -1
				for l := 0; l < len(mesh.Coords); l++ { // FIXME: Coords might be nil
					isMatching := true
					if gltfPrimitive.HasAttribute(AttributePosition) {
						isMatching = isMatching && (mesh.Coords[l] == coord)
					}
					if gltfPrimitive.HasAttribute(AttributeNormal) {
						isMatching = isMatching && (mesh.Normals[l] == normal)
					}
					if gltfPrimitive.HasAttribute(AttributeTangent) {
						isMatching = isMatching && (mesh.Tangents[l] == tangent)
					}
					if gltfPrimitive.HasAttribute(AttributeTexCoord0) {
						isMatching = isMatching && (mesh.TexCoords[l] == texCoord)
					}
					if gltfPrimitive.HasAttribute(AttributeColor0) {
						isMatching = isMatching && (mesh.Colors[l] == color)
					}
					if isMatching {
						matchingIndex = l
						break
					}
				}

				if matchingIndex != -1 {
					mesh.Indices = append(mesh.Indices, matchingIndex)
				} else {
					mesh.VertexCount++
					if gltfPrimitive.HasAttribute(AttributePosition) {
						mesh.Coords = append(mesh.Coords, coord)
					}
					if gltfPrimitive.HasAttribute(AttributeNormal) {
						mesh.Normals = append(mesh.Normals, normal)
					}
					if gltfPrimitive.HasAttribute(AttributeTangent) {
						mesh.Tangents = append(mesh.Tangents, tangent)
					}
					if gltfPrimitive.HasAttribute(AttributeTexCoord0) {
						mesh.TexCoords = append(mesh.TexCoords, texCoord)
					}
					if gltfPrimitive.HasAttribute(AttributeColor0) {
						mesh.Colors = append(mesh.Colors, color)
					}
					mesh.Indices = append(mesh.Indices, mesh.VertexCount-1)
				}
				mesh.IndexCount++
			}

			switch gltfPrimitive.FindMode() {
			case gltf.PrimitivePoints:
				subMesh.Primitive = PrimitivePoints
			case gltf.PrimitiveLines:
				subMesh.Primitive = PrimitiveLines
			case gltf.PrimitiveLineLoop:
				subMesh.Primitive = PrimitiveLineLoop
			case gltf.PrimitiveLineStrip:
				subMesh.Primitive = PrimitiveLineStrip
			case gltf.PrimitiveTriangles:
				subMesh.Primitive = PrimitiveTriangles
			case gltf.PrimitiveTriangleStrip:
				subMesh.Primitive = PrimitiveTriangleStrip
			case gltf.PrimitiveTriangleFan:
				subMesh.Primitive = PrimitiveTriangleFan
			default:
				subMesh.Primitive = PrimitiveTriangles
			}

			gltfMaterial := gltfPrimitive.FindMaterial()
			subMesh.Material.Type = "pbr"
			subMesh.Material.BackfaceCulling = !gltfMaterial.DoubleSided
			subMesh.Material.AlphaTesting = true  // TODO
			subMesh.Material.AlphaThreshold = 0.5 // TODO
			subMesh.Material.Metalness = gltfMaterial.FindMetallic()
			// TODO: Metalness texture
			subMesh.Material.Roughness = gltfMaterial.FindRoughness()
			if name, found := gltfMaterial.FindRoughnessTexture(); found {
				subMesh.Material.RoughnessTexture = name
			}
			if color, found := gltfMaterial.FindBaseColor(); found {
				subMesh.Material.Color = sprec.NewVec4(color[0], color[1], color[2], color[3])
			} else {
				subMesh.Material.Color = sprec.NewVec4(1.0, 1.0, 1.0, 1.0)
			}
			if name, found := gltfMaterial.FindColorTexture(); found {
				subMesh.Material.ColorTexture = name
			}
			if name, scale, found := gltfMaterial.FindNormalTexture(); found {
				subMesh.Material.NormalScale = scale
				subMesh.Material.NormalTexture = name
			}

			mesh.SubMeshes[j] = subMesh
		}

		meshMapping[uint32(i)] = mesh
		a.model.Meshes = append(a.model.Meshes, mesh)
	}

	// build nodes
	var visitNode func(gltfNode *gltf.Node) *Node
	visitNode = func(gltfNode *gltf.Node) *Node {
		node := &Node{
			Name:        gltfNode.Name,
			Translation: sprec.ZeroVec3(),
			Rotation:    sprec.IdentityQuat(),
			Scale:       sprec.NewVec3(1.0, 1.0, 1.0),
		}

		if gltfNode.Matrix != emptyMatrix {
			matrix := sprec.ColumnMajorArrayMat4(gltfNode.Matrix)
			node.Translation = matrix.Translation()
			node.Scale = matrix.Scale()
			node.Rotation = matrix.RotationQuat()
		} else {
			// TODO: Fix these; they should not be empty if matrix is empty
			// and in theory the desired scale may be zero,zero,zero.
			if gltfNode.Translation != emptyTranslation {
				node.Translation = sprec.NewVec3(
					gltfNode.Translation[0],
					gltfNode.Translation[1],
					gltfNode.Translation[2],
				)
			} else {
				node.Translation = sprec.ZeroVec3()
			}
			if gltfNode.Rotation != emptyRotation {
				node.Rotation = sprec.NewQuat(
					gltfNode.Rotation[3],
					gltfNode.Rotation[0],
					gltfNode.Rotation[1],
					gltfNode.Rotation[2],
				)
			} else {
				node.Rotation = sprec.IdentityQuat()
			}
			if gltfNode.Scale != emptyScale {
				node.Scale = sprec.NewVec3(
					gltfNode.Scale[0],
					gltfNode.Scale[1],
					gltfNode.Scale[2],
				)
			} else {
				node.Scale = sprec.NewVec3(1.0, 1.0, 1.0)
			}
		}

		if gltfNode.Mesh != nil {
			node.Mesh = meshMapping[*gltfNode.Mesh]
		}
		for _, childID := range gltfNode.Children {
			node.Children = append(node.Children, visitNode(gltfDoc.Nodes[childID]))
		}
		return node
	}
	for _, node := range gltfDoc.RootNodes() {
		a.model.RootNodes = append(a.model.RootNodes, visitNode(node))
	}

	return nil
}

type gltfLocator struct {
	locator ResourceLocator
	uri     string
}

func (l gltfLocator) Open() (io.ReadCloser, error) {
	return l.locator.Open(l.uri)
}

func (l gltfLocator) OpenRelative(uri string) (io.ReadCloser, error) {
	return l.locator.Open(path.Join(path.Dir(l.uri), uri))
}

type GLTFDocument struct {
	*gltf.Document
}

func (d GLTFDocument) RootNodes() []*gltf.Node {
	childrenIDs := make(map[uint32]struct{})
	for _, node := range d.Nodes {
		for _, childID := range node.Children {
			childrenIDs[childID] = struct{}{}
		}
	}
	var result []*gltf.Node
	for id, node := range d.Nodes {
		if _, ok := childrenIDs[uint32(id)]; !ok {
			result = append(result, node)
		}
	}
	return result
}

func (d GLTFDocument) FindMesh(index int) GLTFMesh {
	return GLTFMesh{
		doc:  d,
		Mesh: d.Meshes[index],
	}
}

type GLTFMesh struct {
	doc GLTFDocument
	*gltf.Mesh
}

func (m GLTFMesh) FindPrimitive(index int) GLTFPrimitive {
	return GLTFPrimitive{
		doc:       m.doc,
		Primitive: m.Primitives[index],
	}
}

type GLTFPrimitive struct {
	doc GLTFDocument
	*gltf.Primitive
}

func (p GLTFPrimitive) HasAttribute(name string) bool {
	_, ok := p.Attributes[name]
	return ok
}

func (p GLTFPrimitive) FindMode() gltf.PrimitiveMode {
	return p.Mode
}

func (p GLTFPrimitive) FindMaterial() GLTFMaterial {
	if p.Material == nil {
		panic("primitive lacks material")
	}
	return GLTFMaterial{
		doc:      p.doc,
		Material: p.doc.Materials[*p.Material],
	}
}

func (p GLTFPrimitive) FindIndexCount() uint32 {
	if p.Indices == nil {
		panic("missing indices: unsupported")
	}
	accessor := p.doc.Accessors[*p.Indices]
	return accessor.Count
}

func (p GLTFPrimitive) FindIndex(index int) int {
	if p.Indices == nil {
		return index
	}
	accessor := p.doc.Accessors[*p.Indices]
	if accessor.BufferView == nil {
		return index
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	switch accessor.ComponentType {
	case gltf.ComponentUshort:
		return int(buffer.UInt16(2 * index))
	case gltf.ComponentUint:
		return int(buffer.UInt32(4 * index))
	default:
		panic(fmt.Errorf("unsupported index component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindCoord(index int) sprec.Vec3 {
	if !p.HasAttribute(AttributePosition) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[AttributePosition]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec3()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec3 {
		panic(fmt.Errorf("unsupported coord type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec3(
			buffer.Float32(3*4*index+4*0),
			buffer.Float32(3*4*index+4*1),
			buffer.Float32(3*4*index+4*2),
		)
	default:
		panic(fmt.Errorf("unsupported coord component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindNormal(index int) sprec.Vec3 {
	if !p.HasAttribute(AttributeNormal) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[AttributeNormal]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec3()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec3 {
		panic(fmt.Errorf("unsupported normal type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec3(
			buffer.Float32(3*4*index+4*0),
			buffer.Float32(3*4*index+4*1),
			buffer.Float32(3*4*index+4*2),
		)
	default:
		panic(fmt.Errorf("unsupported normal component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindTangent(index int) sprec.Vec3 {
	if !p.HasAttribute(AttributeTangent) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[AttributeTangent]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec3()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec3 {
		panic(fmt.Errorf("unsupported tangent type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec3(
			buffer.Float32(3*4*index+4*0),
			buffer.Float32(3*4*index+4*1),
			buffer.Float32(3*4*index+4*2),
		)
	default:
		panic(fmt.Errorf("unsupported tangent component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindTexCoord0(index int) sprec.Vec2 {
	if !p.HasAttribute(AttributeTexCoord0) {
		return sprec.ZeroVec2()
	}
	accessor := p.doc.Accessors[p.Attributes[AttributeTexCoord0]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec2()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec2 {
		panic(fmt.Errorf("unsupported tex coord type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec2(
			buffer.Float32(2*4*index+4*0),
			1.0-buffer.Float32(2*4*index+4*1), // fix tex coord orientation
		)
	default:
		panic(fmt.Errorf("unsupported tex coord component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindColor0(index int) sprec.Vec4 {
	if !p.HasAttribute(AttributeColor0) {
		return sprec.ZeroVec4()
	}
	accessor := p.doc.Accessors[p.Attributes[AttributeColor0]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec4()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.AccessorVec4 {
		panic(fmt.Errorf("unsupported color type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentFloat:
		return sprec.NewVec4(
			buffer.Float32(4*4*index+4*0),
			buffer.Float32(4*4*index+4*1),
			buffer.Float32(4*4*index+4*2),
			buffer.Float32(4*4*index+4*3),
		)
	default:
		panic(fmt.Errorf("unsupported color component type %d", accessor.ComponentType))
	}
}

type GLTFMaterial struct {
	doc GLTFDocument
	*gltf.Material
}

func (m GLTFMaterial) FindMetallic() float32 {
	if m.PBRMetallicRoughness == nil {
		panic("material lacks pbr metallic roughness")
	}
	if m.PBRMetallicRoughness.MetallicFactor == nil {
		return 1.0
	}
	return *m.PBRMetallicRoughness.MetallicFactor
}

func (m GLTFMaterial) FindRoughness() float32 {
	if m.PBRMetallicRoughness == nil {
		panic("material lacks pbr metallic roughness")
	}
	if m.PBRMetallicRoughness.RoughnessFactor == nil {
		return 1.0
	}
	return *m.PBRMetallicRoughness.RoughnessFactor
}

func (m GLTFMaterial) FindBaseColor() ([4]float32, bool) {
	if m.PBRMetallicRoughness == nil {
		return [4]float32{}, false
	}
	if m.PBRMetallicRoughness.BaseColorFactor == nil {
		return [4]float32{}, false
	}
	return *m.PBRMetallicRoughness.BaseColorFactor, true
}

func (m GLTFMaterial) FindColorTexture() (string, bool) {
	if m.PBRMetallicRoughness == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.BaseColorTexture == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.BaseColorTexture.TexCoord > 0 {
		panic("mesh material uses multiple uv coords")
	}
	texture := m.doc.Textures[m.PBRMetallicRoughness.BaseColorTexture.Index]
	if texture.Source == nil {
		panic("texture lacks a source")
	}
	image := m.doc.Images[*texture.Source]
	return image.Name, true
}

func (m GLTFMaterial) FindRoughnessTexture() (string, bool) {
	if m.PBRMetallicRoughness == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.MetallicRoughnessTexture == nil {
		return "", false
	}
	if m.PBRMetallicRoughness.MetallicRoughnessTexture.TexCoord > 0 {
		panic("mesh material uses multiple uv coords")
	}
	texture := m.doc.Textures[m.PBRMetallicRoughness.MetallicRoughnessTexture.Index]
	if texture.Source == nil {
		panic("texture lacks a source")
	}
	image := m.doc.Images[*texture.Source]
	return image.Name, true
}

func (m GLTFMaterial) FindNormalTexture() (string, float32, bool) {
	if m.NormalTexture == nil {
		return "", 0.0, false
	}
	if m.NormalTexture.TexCoord > 0 {
		panic("mesh material uses multiple uv coords")
	}
	if m.NormalTexture.Index == nil {
		panic("missign texture index")
	}
	texture := m.doc.Textures[*m.NormalTexture.Index]
	if texture.Source == nil {
		panic("texture lacks a source")
	}
	image := m.doc.Images[*texture.Source]
	scale := float32(1.0)
	if m.NormalTexture.Scale != nil {
		scale = *m.NormalTexture.Scale
	}
	return image.Name, scale, true
}
