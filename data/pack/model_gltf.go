package pack

import (
	"fmt"
	"io"
	"path"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/data/gltf"
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
	rawGLTF, err := gltf.Parse(gltfLocator{
		locator: a.locator,
		uri:     a.uri,
	})
	if err != nil {
		return fmt.Errorf("failed to parse gltf model %q: %w", a.uri, err)
	}
	gltfDoc := GLTFDocument{rawGLTF}

	a.model = &Model{}

	// build meshes
	meshMapping := make(map[int]*Mesh)

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
			subMesh.IndexCount = gltfPrimitive.FindIndexCount()

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
					if gltfPrimitive.HasAttribute(gltf.AttributePosition) {
						isMatching = isMatching && (mesh.Coords[l] == coord)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeNormal) {
						isMatching = isMatching && (mesh.Normals[l] == normal)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeTangent) {
						isMatching = isMatching && (mesh.Tangents[l] == tangent)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeTexCoord0) {
						isMatching = isMatching && (mesh.TexCoords[l] == texCoord)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeColor0) {
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
					if gltfPrimitive.HasAttribute(gltf.AttributePosition) {
						mesh.Coords = append(mesh.Coords, coord)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeNormal) {
						mesh.Normals = append(mesh.Normals, normal)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeTangent) {
						mesh.Tangents = append(mesh.Tangents, tangent)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeTexCoord0) {
						mesh.TexCoords = append(mesh.TexCoords, texCoord)
					}
					if gltfPrimitive.HasAttribute(gltf.AttributeColor0) {
						mesh.Colors = append(mesh.Colors, color)
					}
					mesh.Indices = append(mesh.Indices, mesh.VertexCount-1)
				}
				mesh.IndexCount++
			}

			switch gltfPrimitive.FindMode() {
			case gltf.ModePoints:
				subMesh.Primitive = PrimitivePoints
			case gltf.ModeLines:
				subMesh.Primitive = PrimitiveLines
			case gltf.ModeLineLoop:
				subMesh.Primitive = PrimitiveLineLoop
			case gltf.ModeLineStrip:
				subMesh.Primitive = PrimitiveLineStrip
			case gltf.ModeTriangles:
				subMesh.Primitive = PrimitiveTriangles
			case gltf.ModeTriangleStrip:
				subMesh.Primitive = PrimitiveTriangleStrip
			case gltf.ModeTriangleFan:
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

		meshMapping[i] = mesh
		a.model.Meshes = append(a.model.Meshes, mesh)
	}

	// build nodes
	var visitNode func(gltfNode gltf.Node) *Node
	visitNode = func(gltfNode gltf.Node) *Node {
		node := &Node{
			Name:        gltfNode.Name,
			Translation: sprec.ZeroVec3(),             // TODO
			Rotation:    sprec.IdentityQuat(),         // TODO
			Scale:       sprec.NewVec3(1.0, 1.0, 1.0), // TODO
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

func (d GLTFDocument) RootNodes() []gltf.Node {
	childrenIDs := make(map[int]struct{})
	for _, node := range d.Nodes {
		for _, childID := range node.Children {
			childrenIDs[childID] = struct{}{}
		}
	}
	var result []gltf.Node
	for id, node := range d.Nodes {
		if _, ok := childrenIDs[id]; !ok {
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
	gltf.Mesh
}

func (m GLTFMesh) FindPrimitive(index int) GLTFPrimitive {
	return GLTFPrimitive{
		doc:       m.doc,
		Primitive: m.Primitives[index],
	}
}

type GLTFPrimitive struct {
	doc GLTFDocument
	gltf.Primitive
}

func (p GLTFPrimitive) FindMode() int {
	if p.Mode == nil {
		return gltf.ModeTriangles
	}
	return *p.Mode
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

func (p GLTFPrimitive) FindIndexCount() int {
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
	case gltf.ComponentTypeUnsignedShort:
		return int(buffer.UInt16(2 * index))
	default:
		panic(fmt.Errorf("unsupported index component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindCoord(index int) sprec.Vec3 {
	if !p.HasAttribute(gltf.AttributePosition) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributePosition]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec3()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec3 {
		panic(fmt.Errorf("unsupported coord type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
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
	if !p.HasAttribute(gltf.AttributeNormal) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeNormal]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec3()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec3 {
		panic(fmt.Errorf("unsupported normal type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
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
	if !p.HasAttribute(gltf.AttributeTangent) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeTangent]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec3()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec3 {
		panic(fmt.Errorf("unsupported tangent type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
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
	if !p.HasAttribute(gltf.AttributeTexCoord0) {
		return sprec.ZeroVec2()
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeTexCoord0]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec2()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec2 {
		panic(fmt.Errorf("unsupported tex coord type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
		return sprec.NewVec2(
			buffer.Float32(2*4*index+4*0),
			1.0-buffer.Float32(2*4*index+4*1), // fix tex coord orientation
		)
	default:
		panic(fmt.Errorf("unsupported tex coord component type %d", accessor.ComponentType))
	}
}

func (p GLTFPrimitive) FindColor0(index int) sprec.Vec4 {
	if !p.HasAttribute(gltf.AttributeColor0) {
		return sprec.ZeroVec4()
	}
	accessor := p.doc.Accessors[p.Attributes[gltf.AttributeColor0]]
	if accessor.BufferView == nil {
		return sprec.ZeroVec4()
	}
	bufferView := p.doc.BufferViews[*accessor.BufferView]
	buffer := data.Buffer(p.doc.Buffers[bufferView.Buffer].Data[bufferView.ByteOffset:])
	if accessor.Type != gltf.TypeVec4 {
		panic(fmt.Errorf("unsupported color type %s", accessor.Type))
	}
	switch accessor.ComponentType {
	case gltf.ComponentTypeFloat:
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
	gltf.Material
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

func (m GLTFMaterial) FindBaseColor() (gltf.Color, bool) {
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
	texture := m.doc.Textures[m.NormalTexture.Index]
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
