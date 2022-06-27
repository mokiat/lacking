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

	// build materials
	materialFromIndex := make(map[uint32]*Material)
	for i, gltfMaterial := range gltfDoc.Materials {
		material := &Material{
			Name:                     gltfMaterial.Name,
			BackfaceCulling:          !gltfMaterial.DoubleSided,
			AlphaTesting:             gltfMaterial.AlphaMode == gltf.AlphaMask,
			AlphaThreshold:           gltfMaterial.AlphaCutoffOrDefault(),
			Blending:                 gltfMaterial.AlphaMode == gltf.AlphaBlend,
			Color:                    sprec.NewVec4(1.0, 1.0, 1.0, 1.0),
			ColorTexture:             "",
			Metallic:                 1.0,
			Roughness:                1.0,
			MetallicRoughnessTexture: "",
			NormalScale:              1.0,
			NormalTexture:            "",
		}
		if gltfPBR := gltfMaterial.PBRMetallicRoughness; gltfPBR != nil {
			material.Color = gltfutil.BaseColor(gltfPBR)
			material.Metallic = gltfPBR.MetallicFactorOrDefault()
			material.Roughness = gltfPBR.RoughnessFactorOrDefault()
			material.ColorTexture = gltfutil.ColorTexture(gltfDoc, gltfPBR)
			material.MetallicRoughnessTexture = gltfutil.MetallicRoughnessTexture(gltfDoc, gltfPBR)
		}
		material.NormalTexture, material.NormalScale = gltfutil.NormalTexture(gltfDoc, gltfMaterial)

		a.model.Materials = append(a.model.Materials, material)
		materialFromIndex[uint32(i)] = material
	}

	// build mesh definitions
	meshDefinitionFromIndex := make(map[uint32]*MeshDefinition)
	for i, gltfMesh := range gltfDoc.Meshes {
		mesh := &MeshDefinition{
			Name:      gltfMesh.Name,
			Fragments: make([]MeshFragment, len(gltfMesh.Primitives)),
		}
		meshDefinitionFromIndex[uint32(i)] = mesh
		a.model.MeshDefinitions = append(a.model.MeshDefinitions, mesh)
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
			if gltfutil.HasAttribute(gltfPrimitive, gltf.WEIGHTS_0) {
				mesh.VertexLayout.HasWeights = true
			}
			if gltfutil.HasAttribute(gltfPrimitive, gltf.JOINTS_0) {
				mesh.VertexLayout.HasJoints = true
			}

			fragment := MeshFragment{}
			fragment.IndexOffset = len(mesh.Indices)
			fragment.IndexCount = gltfutil.IndexCount(gltfDoc, gltfPrimitive)

			for k := 0; k < fragment.IndexCount; k++ {
				gltfIndex := gltfutil.Index(gltfDoc, gltfPrimitive, k)
				vertex := Vertex{
					Coord:    gltfutil.Coord(gltfDoc, gltfPrimitive, gltfIndex),
					Normal:   gltfutil.Normal(gltfDoc, gltfPrimitive, gltfIndex),
					Tangent:  gltfutil.Tangent(gltfDoc, gltfPrimitive, gltfIndex),
					TexCoord: gltfutil.TexCoord0(gltfDoc, gltfPrimitive, gltfIndex),
					Color:    gltfutil.Color0(gltfDoc, gltfPrimitive, gltfIndex),
					Weights:  gltfutil.Weights0(gltfDoc, gltfPrimitive, gltfIndex),
					Joints:   gltfutil.Joints0(gltfDoc, gltfPrimitive, gltfIndex),
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
				fragment.Primitive = PrimitivePoints
			case gltf.PrimitiveLines:
				fragment.Primitive = PrimitiveLines
			case gltf.PrimitiveLineLoop:
				fragment.Primitive = PrimitiveLineLoop
			case gltf.PrimitiveLineStrip:
				fragment.Primitive = PrimitiveLineStrip
			case gltf.PrimitiveTriangles:
				fragment.Primitive = PrimitiveTriangles
			case gltf.PrimitiveTriangleStrip:
				fragment.Primitive = PrimitiveTriangleStrip
			case gltf.PrimitiveTriangleFan:
				fragment.Primitive = PrimitiveTriangleFan
			default:
				fragment.Primitive = PrimitiveTriangles
			}

			if gltfPrimitive.Material != nil {
				fragment.Material = materialFromIndex[*gltfPrimitive.Material]
			}

			mesh.Fragments[j] = fragment
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
			meshInstance := &MeshInstance{
				Name:       gltfNode.Name,
				Node:       node,
				Definition: meshDefinitionFromIndex[*gltfNode.Mesh],
			}
			a.model.MeshInstances = append(a.model.MeshInstances, meshInstance)
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
