package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/data/gltf"
)

type ModelAssetBuilder struct {
	Asset
	gltfProvider GLTFProvider
}

func (b *ModelAssetBuilder) FromGLTF(gltfProvider GLTFProvider) *ModelAssetBuilder {
	b.gltfProvider = gltfProvider
	return b
}

func (b *ModelAssetBuilder) Build() error {
	gltfDoc, err := b.gltfProvider.GLTF()
	if err != nil {
		return fmt.Errorf("failed to get gltf document: %w", err)
	}

	model := &asset.Model{
		Nodes:  make([]asset.Node, len(gltfDoc.Nodes)),
		Meshes: make([]asset.Mesh, len(gltfDoc.Meshes)),
	}

	for n := range model.Nodes {
		model.Nodes[n].ParentIndex = -1
	}
	for n, gltfNode := range gltfDoc.Nodes {
		node := &model.Nodes[n]
		node.Name = gltfNode.Name

		for _, j := range gltfNode.Children {
			model.Nodes[j].ParentIndex = int16(n)
		}

		// TODO: Translation, Rotation, Scale (TRS)

		node.Matrix = [16]float32{
			1.0, 0.0, 0.0, 0.0,
			0.0, 1.0, 0.0, 0.0,
			0.0, 0.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 1.0,
		}

		if gltfNode.Mesh == nil {
			return fmt.Errorf("node %q is missing a mesh", gltfNode.Name)
		}
		node.MeshIndex = uint16(*gltfNode.Mesh)
	}

	for m := range gltfDoc.Meshes {
		gltfMesh := gltfDoc.FindMesh(m)
		mesh := &model.Meshes[m]
		mesh.Name = gltfMesh.Name
		meshConstructor := NewMeshConstructor()

		mesh.SubMeshes = make([]asset.SubMesh, len(gltfMesh.Primitives))
		for s := range gltfMesh.Primitives {
			gltfPrimitive := gltfMesh.FindPrimitive(s)
			subMesh := &mesh.SubMeshes[s]
			subMeshConstructor := meshConstructor.SubMesh()

			switch gltfPrimitive.FindMode() {
			case gltf.ModePoints:
				subMesh.Primitive = asset.PrimitiveTriangles
			case gltf.ModeLines:
				subMesh.Primitive = asset.PrimitiveLines
			case gltf.ModeLineLoop:
				subMesh.Primitive = asset.PrimitiveLineLoop
			case gltf.ModeLineStrip:
				subMesh.Primitive = asset.PrimitiveLineStrip
			case gltf.ModeTriangles:
				subMesh.Primitive = asset.PrimitiveTriangles
			case gltf.ModeTriangleStrip:
				subMesh.Primitive = asset.PrimitiveTriangleStrip
			case gltf.ModeTriangleFan:
				subMesh.Primitive = asset.PrimitiveTriangleFan
			}

			gltfMaterial := gltfPrimitive.FindMaterial()

			subMesh.Material.Type = "pbr"
			subMesh.Material.BackfaceCulling = !gltfMaterial.DoubleSided
			subMesh.Material.Metalness = gltfMaterial.FindMetallic()
			subMesh.Material.Color = [4]float32{1.0, 0.0, 0.0, 1.0}
			if color, found := gltfMaterial.FindBaseColor(); found {
				subMesh.Material.Color = color
			} else {
				subMesh.Material.Color = [4]float32{1.0, 1.0, 1.0, 1.0}
			}
			if name, found := gltfMaterial.FindColorTexture(); found {
				subMesh.Material.ColorTexture = name
			}
			subMesh.Material.Roughness = gltfMaterial.FindRoughness()
			if name, found := gltfMaterial.FindRoughnessTexture(); found {
				subMesh.Material.RoughnessTexture = name
			}
			if name, scale, found := gltfMaterial.FindNormalTexture(); found {
				subMesh.Material.NormalScale = scale
				subMesh.Material.NormalTexture = name
			}

			indexCount := gltfPrimitive.FindIndexCount()
			subMesh.IndexCount = uint32(indexCount)
			subMesh.IndexOffset = uint32(subMeshConstructor.IndexOffsetBytes())

			for i := 0; i < indexCount; i++ {
				index := gltfPrimitive.FindIndex(i)
				vertex := Vertex{
					Coord:     gltfPrimitive.FindCoord(index),
					Normal:    gltfPrimitive.FindNormal(index),
					Tangent:   gltfPrimitive.FindTangent(index),
					TexCoord0: gltfPrimitive.FindTexCoord0(index),
					Color0:    gltfPrimitive.FindColor0(index),
				}
				vertexIndex := meshConstructor.AddVertex(vertex)
				subMeshConstructor.AddIndex(vertexIndex)
			}
		}

		vertexMask := VertexMask{
			HasCoord:     gltfMesh.HasAttribute(gltf.AttributePosition),
			HasNormal:    gltfMesh.HasAttribute(gltf.AttributeNormal),
			HasTangent:   gltfMesh.HasAttribute(gltf.AttributeTangent),
			HasTexCoord0: gltfMesh.HasAttribute(gltf.AttributeTexCoord0),
			HasColor0:    gltfMesh.HasAttribute(gltf.AttributeColor0),
		}
		mesh.VertexData = meshConstructor.VertexData(vertexMask)
		vertexLayout := vertexMask.Layout()
		mesh.VertexLayout = asset.VertexLayout{
			CoordOffset:    int16(vertexLayout.CoordOffset),
			CoordStride:    int16(vertexLayout.Stride),
			NormalOffset:   int16(vertexLayout.NormalOffset),
			NormalStride:   int16(vertexLayout.Stride),
			TangentOffset:  int16(vertexLayout.TangentOffset),
			TangentStride:  int16(vertexLayout.Stride),
			TexCoordOffset: int16(vertexLayout.TexCoord0Offset),
			TexCoordStride: int16(vertexLayout.Stride),
			ColorOffset:    int16(vertexLayout.Color0Offset),
			ColorStride:    int16(vertexLayout.Stride),
		}
		mesh.IndexData = meshConstructor.IndexData()
	}

	file, err := b.CreateFile()
	if err != nil {
		return err
	}
	defer file.Close()

	if err := asset.EncodeModel(file, model); err != nil {
		return fmt.Errorf("failed to encode model: %w", err)
	}
	return nil
}

type Vertex struct {
	Coord     [3]float32
	Normal    [3]float32
	Tangent   [3]float32
	TexCoord0 [2]float32
	Color0    [4]float32
}

type VertexLayout struct {
	CoordSize       int
	CoordOffset     int
	NormalSize      int
	NormalOffset    int
	TangentSize     int
	TangentOffset   int
	TexCoord0Size   int
	TexCoord0Offset int
	Color0Size      int
	Color0Offset    int
	Stride          int
}

type VertexMask struct {
	HasCoord     bool
	HasNormal    bool
	HasTangent   bool
	HasTexCoord0 bool
	HasColor0    bool
}

func (m VertexMask) Layout() VertexLayout {
	layout := VertexLayout{
		CoordSize:       3 * 4,
		CoordOffset:     -1,
		NormalSize:      3 * 4,
		NormalOffset:    -1,
		TangentSize:     3 * 4,
		TangentOffset:   -1,
		TexCoord0Size:   2 * 4,
		TexCoord0Offset: -1,
		Color0Size:      4 * 4,
		Color0Offset:    -1,
	}

	layout.Stride = 0
	if m.HasCoord {
		layout.CoordOffset = layout.Stride
		layout.Stride += layout.CoordSize
	}
	if m.HasNormal {
		layout.NormalOffset = layout.Stride
		layout.Stride += layout.NormalSize
	}
	if m.HasTangent {
		layout.TangentOffset = layout.Stride
		layout.Stride += layout.TangentSize
	}
	if m.HasTexCoord0 {
		layout.TexCoord0Offset = layout.Stride
		layout.Stride += layout.TexCoord0Size
	}
	if m.HasColor0 {
		layout.Color0Offset = layout.Stride
		layout.Stride += layout.Color0Size
	}
	return layout
}

func NewMeshConstructor() *MeshConstructor {
	return &MeshConstructor{
		vertices: make(map[Vertex]int),
	}
}

type MeshConstructor struct {
	vertices map[Vertex]int
	indices  []int
}

func (c *MeshConstructor) AddVertex(vertex Vertex) int {
	if index, found := c.vertices[vertex]; found {
		return index
	}
	index := len(c.vertices)
	c.vertices[vertex] = index
	return index
}

func (c *MeshConstructor) VertexData(mask VertexMask) []byte {
	layout := mask.Layout()

	buffer := data.Buffer(make([]byte, len(c.vertices)*layout.Stride))
	for vertex, index := range c.vertices {
		if mask.HasCoord {
			buffer.SetFloat32(index*layout.Stride+layout.CoordOffset+4*0, vertex.Coord[0])
			buffer.SetFloat32(index*layout.Stride+layout.CoordOffset+4*1, vertex.Coord[1])
			buffer.SetFloat32(index*layout.Stride+layout.CoordOffset+4*2, vertex.Coord[2])
		}
		if mask.HasNormal {
			buffer.SetFloat32(index*layout.Stride+layout.NormalOffset+4*0, vertex.Normal[0])
			buffer.SetFloat32(index*layout.Stride+layout.NormalOffset+4*1, vertex.Normal[1])
			buffer.SetFloat32(index*layout.Stride+layout.NormalOffset+4*2, vertex.Normal[2])
		}
		if mask.HasTangent {
			buffer.SetFloat32(index*layout.Stride+layout.TangentOffset+4*0, vertex.Tangent[0])
			buffer.SetFloat32(index*layout.Stride+layout.TangentOffset+4*1, vertex.Tangent[1])
			buffer.SetFloat32(index*layout.Stride+layout.TangentOffset+4*2, vertex.Tangent[2])
		}
		if mask.HasTexCoord0 {
			buffer.SetFloat32(index*layout.Stride+layout.TexCoord0Offset+4*0, vertex.TexCoord0[0])
			buffer.SetFloat32(index*layout.Stride+layout.TexCoord0Offset+4*1, vertex.TexCoord0[1])
		}
		if mask.HasColor0 {
			buffer.SetFloat32(index*layout.Stride+layout.Color0Offset+4*0, vertex.Color0[0])
			buffer.SetFloat32(index*layout.Stride+layout.Color0Offset+4*1, vertex.Color0[1])
			buffer.SetFloat32(index*layout.Stride+layout.Color0Offset+4*2, vertex.Color0[2])
			buffer.SetFloat32(index*layout.Stride+layout.Color0Offset+4*3, vertex.Color0[3])
		}
	}
	return buffer
}

func (c *MeshConstructor) IndexData() []byte {
	buffer := data.Buffer(make([]byte, len(c.indices)*2))
	for i, index := range c.indices {
		buffer.SetUInt16(i*2, uint16(index))
	}
	return buffer
}

func (c *MeshConstructor) SubMesh() *SubMeshConstructor {
	return &SubMeshConstructor{
		meshConst: c,
	}
}

type SubMeshConstructor struct {
	meshConst *MeshConstructor
}

func (c *SubMeshConstructor) IndexOffsetBytes() int {
	return len(c.meshConst.indices) * 2
}

func (c *SubMeshConstructor) AddIndex(index int) {
	c.meshConst.indices = append(c.meshConst.indices, index)
}
