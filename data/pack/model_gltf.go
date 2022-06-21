package pack

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/gltfutil"
	"github.com/qmuntal/gltf"
)

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
	gltfDoc, err := gltf.Open(a.uri)
	if err != nil {
		return fmt.Errorf("failed to parse gltf model %q: %w", a.uri, err)
	}

	a.model = &Model{}

	// build meshes
	meshFromIndex := make(map[uint32]*Mesh)

	for i, gltfMesh := range gltfDoc.Meshes {
		mesh := &Mesh{
			Name:      gltfMesh.Name,
			SubMeshes: make([]SubMesh, len(gltfMesh.Primitives)),
		}
		meshFromIndex[uint32(i)] = mesh
		a.model.Meshes = append(a.model.Meshes, mesh)
		indexFromVertex := make(map[Vertex]int)

		for j, gltfPrimitive := range gltfMesh.Primitives {
			if gltfutil.HasAttribute(gltfPrimitive, gltf.POSITION) {
				mesh.VertexLayout.HasCoords = true
			}
			if gltfutil.HasAttribute(gltfPrimitive, gltf.NORMAL) {
				mesh.VertexLayout.HasNormals = true
			}
			if gltfutil.HasAttribute(gltfPrimitive, gltf.TANGENT) {
				mesh.VertexLayout.HasTangents = true
			}
			if gltfutil.HasAttribute(gltfPrimitive, gltf.TEXCOORD_0) {
				mesh.VertexLayout.HasTexCoords = true
			}
			if gltfutil.HasAttribute(gltfPrimitive, gltf.COLOR_0) {
				mesh.VertexLayout.HasColors = true
			}

			subMesh := SubMesh{}
			subMesh.IndexOffset = len(mesh.Indices)
			subMesh.IndexCount = gltfutil.IndexCount(gltfDoc, gltfPrimitive)

			for k := 0; k < subMesh.IndexCount; k++ {
				gltfIndex := gltfutil.Index(gltfDoc, gltfPrimitive, k)
				vertex := Vertex{
					Coord:    gltfutil.Coord(gltfDoc, gltfPrimitive, gltfIndex),
					Normal:   gltfutil.Normal(gltfDoc, gltfPrimitive, gltfIndex),
					Tangent:  gltfutil.Tangent(gltfDoc, gltfPrimitive, gltfIndex),
					TexCoord: gltfutil.TexCoord0(gltfDoc, gltfPrimitive, gltfIndex),
					Color:    gltfutil.Color0(gltfDoc, gltfPrimitive, gltfIndex),
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

			material := Material{
				Type:                     "pbr",
				BackfaceCulling:          true,
				AlphaTesting:             false,
				AlphaThreshold:           0.5,
				Blending:                 false,
				Color:                    sprec.NewVec4(1.0, 1.0, 1.0, 1.0),
				ColorTexture:             "",
				Metallic:                 1.0,
				Roughness:                1.0,
				MetallicRoughnessTexture: "",
				NormalScale:              1.0,
				NormalTexture:            "",
			}
			if gltfMaterial := gltfutil.PrimitiveMaterial(gltfDoc, gltfPrimitive); gltfMaterial != nil {
				material.BackfaceCulling = !gltfMaterial.DoubleSided
				material.AlphaTesting = gltfMaterial.AlphaMode == gltf.AlphaMask
				material.AlphaThreshold = gltfMaterial.AlphaCutoffOrDefault()
				material.Blending = gltfMaterial.AlphaMode == gltf.AlphaBlend
				if gltfPBR := gltfMaterial.PBRMetallicRoughness; gltfPBR != nil {
					material.Color = gltfutil.BaseColor(gltfPBR)
					material.Metallic = gltfPBR.MetallicFactorOrDefault()
					material.Roughness = gltfPBR.RoughnessFactorOrDefault()
					material.ColorTexture = gltfutil.ColorTexture(gltfDoc, gltfPBR)
					material.MetallicRoughnessTexture = gltfutil.MetallicRoughnessTexture(gltfDoc, gltfPBR)
				}
				material.NormalTexture, material.NormalScale = gltfutil.NormalTexture(gltfDoc, gltfMaterial)
			}
			subMesh.Material = material

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

		if gltfNode.Matrix != gltf.DefaultMatrix {
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
	for _, node := range gltfutil.RootNodes(gltfDoc) {
		a.model.RootNodes = append(a.model.RootNodes, visitNode(node))
	}

	return nil
}
