package pack

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/qmuntal/gltf"
)

const (
	attributePosition  = "POSITION"
	attributeNormal    = "NORMAL"
	attributeTangent   = "TANGENT"
	attributeTexCoord0 = "TEXCOORD_0"
	attributeColor0    = "COLOR_0"
)

var (
	emptyMatrix    = [16]float32{}
	identityMatrix = [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
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

	// NOTE: glTF allows a sub-mesh to use totally different
	// mesh vertices and indices. It may even reuse part of the
	// attributes but use dedicated buffers for the remaining ones.
	//
	// Since we don't support that and our mesh model has a shared
	// vertex data with sub-meshes only having index offsets and counts,
	// we need to reindex the data.
	//
	// This acts also as a form of optimization where if the glTF has
	// additional attributes that we don't care about but that result in
	// mesh partitioning, we would be getting rid of the unnecessary
	// partitioning.

	// build meshes
	meshFromIndex := make(map[uint32]*Mesh)

	for i, gltfMesh := range gltfDoc.GetMeshes() {
		mesh := &Mesh{
			Name:      gltfMesh.Name,
			SubMeshes: make([]SubMesh, len(gltfMesh.Primitives)),
		}
		meshFromIndex[uint32(i)] = mesh
		a.model.Meshes = append(a.model.Meshes, mesh)
		indexFromVertex := make(map[Vertex]int)

		for j, gltfPrimitive := range gltfMesh.GetPrimitives() {
			subMesh := SubMesh{}
			subMesh.IndexOffset = len(mesh.Indices)
			subMesh.IndexCount = int(gltfPrimitive.FindIndexCount())

			if gltfPrimitive.HasAttribute(attributePosition) {
				mesh.VertexLayout.HasCoords = true
			}
			if gltfPrimitive.HasAttribute(attributeNormal) {
				mesh.VertexLayout.HasNormals = true
			}
			if gltfPrimitive.HasAttribute(attributeTangent) {
				mesh.VertexLayout.HasTangents = true
			}
			if gltfPrimitive.HasAttribute(attributeTexCoord0) {
				mesh.VertexLayout.HasTexCoords = true
			}
			if gltfPrimitive.HasAttribute(attributeColor0) {
				mesh.VertexLayout.HasColors = true
			}

			for k := 0; k < subMesh.IndexCount; k++ {
				gltfIndex := gltfPrimitive.FindIndex(k)
				vertex := Vertex{
					Coord:    gltfPrimitive.FindCoord(gltfIndex),
					Normal:   gltfPrimitive.FindNormal(gltfIndex),
					Tangent:  gltfPrimitive.FindTangent(gltfIndex),
					TexCoord: gltfPrimitive.FindTexCoord0(gltfIndex),
					Color:    gltfPrimitive.FindColor0(gltfIndex),
				}

				if index, ok := indexFromVertex[vertex]; ok {
					mesh.Indices = append(mesh.Indices, index)
				} else {
					index = len(mesh.Vertices)
					mesh.Vertices = append(mesh.Vertices, vertex)
					mesh.Indices = append(mesh.Indices, index)
					indexFromVertex[vertex] = index
				}
			}

			switch gltfPrimitive.Mode {
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

		if gltfNode.Matrix != emptyMatrix && gltfNode.Matrix != identityMatrix {
			matrix := sprec.ColumnMajorArrayMat4(gltfNode.Matrix)
			node.Translation = matrix.Translation()
			node.Scale = matrix.Scale()
			node.Rotation = matrix.RotationQuat()
		} else {
			node.Translation = sprec.NewVec3(
				gltfNode.Translation[0],
				gltfNode.Translation[1],
				gltfNode.Translation[2],
			)
			node.Rotation = sprec.NewQuat(
				gltfNode.Rotation[3],
				gltfNode.Rotation[0],
				gltfNode.Rotation[1],
				gltfNode.Rotation[2],
			)
			node.Scale = sprec.NewVec3(
				gltfNode.Scale[0],
				gltfNode.Scale[1],
				gltfNode.Scale[2],
			)
		}

		if gltfNode.Mesh != nil {
			node.Mesh = meshFromIndex[*gltfNode.Mesh]
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

func (d GLTFDocument) GetMeshes() []GLTFMesh {
	result := make([]GLTFMesh, len(d.Meshes))
	for i, mesh := range d.Meshes {
		result[i] = GLTFMesh{
			doc:  d,
			Mesh: mesh,
		}
	}
	return result
}

type GLTFMesh struct {
	doc GLTFDocument
	*gltf.Mesh
}

func (m GLTFMesh) GetPrimitives() []GLTFPrimitive {
	result := make([]GLTFPrimitive, len(m.Primitives))
	for i, primitive := range m.Primitives {
		result[i] = GLTFPrimitive{
			doc:       m.doc,
			Primitive: primitive,
		}
	}
	return result
}

type GLTFPrimitive struct {
	doc GLTFDocument
	*gltf.Primitive
}

func (p GLTFPrimitive) HasAttribute(name string) bool {
	_, ok := p.Attributes[name]
	return ok
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
	if !p.HasAttribute(attributePosition) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[attributePosition]]
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
	if !p.HasAttribute(attributeNormal) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[attributeNormal]]
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
	if !p.HasAttribute(attributeTangent) {
		return sprec.ZeroVec3()
	}
	accessor := p.doc.Accessors[p.Attributes[attributeTangent]]
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
	if !p.HasAttribute(attributeTexCoord0) {
		return sprec.ZeroVec2()
	}
	accessor := p.doc.Accessors[p.Attributes[attributeTexCoord0]]
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
	if !p.HasAttribute(attributeColor0) {
		return sprec.ZeroVec4()
	}
	accessor := p.doc.Accessors[p.Attributes[attributeColor0]]
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
