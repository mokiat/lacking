package pack

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/data/asset"
)

type SaveModelAssetAction struct {
	locator       AssetLocator
	uri           string
	modelProvider ModelProvider
}

func (a *SaveModelAssetAction) Describe() string {
	return fmt.Sprintf("save_model_asset(uri: %q)", a.uri)
}

func (a *SaveModelAssetAction) Run() error {
	model := a.modelProvider.Model()

	nodeMapping := make(map[*Node]int)
	meshMapping := make(map[*Mesh]int)

	modelAsset := &asset.Model{}

	// build meshes
	modelAsset.Meshes = make([]asset.Mesh, len(model.Meshes))
	for i, mesh := range model.Meshes {
		meshMapping[mesh] = i
		modelAsset.Meshes[i] = meshToAssetMesh(mesh)
	}

	// build nodes
	var visitNode func(parent, node *Node)
	visitNode = func(parent, node *Node) {
		nodeMapping[node] = len(nodeMapping)

		nodeAsset := asset.Node{
			Name:   node.Name,
			Matrix: matrixToArray(node.Matrix()),
		}
		if parentIndex, ok := nodeMapping[parent]; ok {
			nodeAsset.ParentIndex = int16(parentIndex)
		} else {
			nodeAsset.ParentIndex = int16(-1)
		}
		if meshIndex, ok := meshMapping[node.Mesh]; ok {
			nodeAsset.MeshIndex = int16(meshIndex)
		} else {
			nodeAsset.MeshIndex = int16(-1)
		}

		modelAsset.Nodes = append(modelAsset.Nodes, nodeAsset)

		for _, child := range node.Children {
			visitNode(node, child)
		}
	}
	for _, node := range model.RootNodes {
		visitNode(nil, node)
	}

	out, err := a.locator.Create(a.uri)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := asset.EncodeModel(out, modelAsset); err != nil {
		return fmt.Errorf("failed to encode asset: %w", err)
	}
	return nil
}

func meshToAssetMesh(mesh *Mesh) asset.Mesh {
	var (
		coordOffset    int16
		normalOffset   int16
		tangentOffset  int16
		texCoordOffset int16
		colorOffset    int16
		stride         int16
	)

	stride = 0
	if mesh.Coords != nil {
		coordOffset = stride
		stride += 3 * 4
	} else {
		coordOffset = asset.UnspecifiedOffset
	}
	if mesh.Normals != nil {
		normalOffset = stride
		stride += 3 * 4
	} else {
		normalOffset = asset.UnspecifiedOffset
	}
	if mesh.Tangents != nil {
		tangentOffset = stride
		stride += 3 * 4
	} else {
		tangentOffset = asset.UnspecifiedOffset
	}
	if mesh.TexCoords != nil {
		texCoordOffset = stride
		stride += 2 * 4
	} else {
		texCoordOffset = asset.UnspecifiedOffset
	}
	if mesh.Colors != nil {
		colorOffset = stride
		stride += 4 * 4
	} else {
		colorOffset = asset.UnspecifiedOffset
	}

	vertexData := data.Buffer(make([]byte, mesh.VertexCount*int(stride)))
	if mesh.Coords != nil {
		for j, coord := range mesh.Coords {
			vertexData.SetFloat32(int(coordOffset)+j*int(stride)+0, coord.X)
			vertexData.SetFloat32(int(coordOffset)+j*int(stride)+4, coord.Y)
			vertexData.SetFloat32(int(coordOffset)+j*int(stride)+8, coord.Z)
		}
	}
	if mesh.Normals != nil {
		for j, normal := range mesh.Normals {
			vertexData.SetFloat32(int(normalOffset)+j*int(stride)+0, normal.X)
			vertexData.SetFloat32(int(normalOffset)+j*int(stride)+4, normal.Y)
			vertexData.SetFloat32(int(normalOffset)+j*int(stride)+8, normal.Z)
		}
	}
	if mesh.Tangents != nil {
		for j, tangent := range mesh.Tangents {
			vertexData.SetFloat32(int(tangentOffset)+j*int(stride)+0, tangent.X)
			vertexData.SetFloat32(int(tangentOffset)+j*int(stride)+4, tangent.Y)
			vertexData.SetFloat32(int(tangentOffset)+j*int(stride)+8, tangent.Z)
		}
	}
	if mesh.TexCoords != nil {
		for j, texCoord := range mesh.TexCoords {
			vertexData.SetFloat32(int(texCoordOffset)+j*int(stride)+0, texCoord.X)
			vertexData.SetFloat32(int(texCoordOffset)+j*int(stride)+4, texCoord.Y)
		}
	}
	if mesh.Colors != nil {
		for j, color := range mesh.Colors {
			vertexData.SetFloat32(int(colorOffset)+j*int(stride)+0, color.X)
			vertexData.SetFloat32(int(colorOffset)+j*int(stride)+4, color.Y)
			vertexData.SetFloat32(int(colorOffset)+j*int(stride)+8, color.Z)
		}
	}

	indexData := data.Buffer(make([]byte, mesh.IndexCount*2))
	for j, index := range mesh.Indices {
		indexData.SetUInt16(j*2, uint16(index))
	}

	meshAsset := asset.Mesh{
		Name:       mesh.Name,
		VertexData: vertexData,
		VertexLayout: asset.VertexLayout{
			CoordOffset:    coordOffset,
			CoordStride:    stride,
			NormalOffset:   normalOffset,
			NormalStride:   stride,
			TangentOffset:  tangentOffset,
			TangentStride:  stride,
			TexCoordOffset: texCoordOffset,
			TexCoordStride: stride,
			ColorOffset:    colorOffset,
			ColorStride:    stride,
		},
		IndexData: indexData,
		SubMeshes: make([]asset.SubMesh, len(mesh.SubMeshes)),
	}
	for j, subMesh := range mesh.SubMeshes {
		subMeshAsset := asset.SubMesh{
			IndexCount:  uint32(subMesh.IndexCount),
			IndexOffset: uint32(subMesh.IndexOffset * 2),
			Material: asset.Material{
				Type:             subMesh.Material.Type,
				BackfaceCulling:  subMesh.Material.BackfaceCulling,
				AlphaTesting:     subMesh.Material.AlphaTesting,
				AlphaThreshold:   subMesh.Material.AlphaThreshold,
				Metalness:        subMesh.Material.Metalness,
				MetalnessTexture: subMesh.Material.MetalnessTexture,
				Roughness:        subMesh.Material.Roughness,
				RoughnessTexture: subMesh.Material.RoughnessTexture,
				Color: [4]float32{
					subMesh.Material.Color.X,
					subMesh.Material.Color.Y,
					subMesh.Material.Color.Z,
					subMesh.Material.Color.W,
				},
				ColorTexture:  subMesh.Material.ColorTexture,
				NormalScale:   subMesh.Material.NormalScale,
				NormalTexture: subMesh.Material.NormalTexture,
			},
		}
		switch subMesh.Primitive {
		case PrimitivePoints:
			subMeshAsset.Primitive = asset.PrimitivePoints
		case PrimitiveLines:
			subMeshAsset.Primitive = asset.PrimitiveLines
		case PrimitiveLineStrip:
			subMeshAsset.Primitive = asset.PrimitiveLineStrip
		case PrimitiveLineLoop:
			subMeshAsset.Primitive = asset.PrimitiveLineLoop
		case PrimitiveTriangles:
			subMeshAsset.Primitive = asset.PrimitiveTriangles
		case PrimitiveTriangleStrip:
			subMeshAsset.Primitive = asset.PrimitiveTriangleStrip
		case PrimitiveTriangleFan:
			subMeshAsset.Primitive = asset.PrimitiveTriangleFan
		default:
			panic(fmt.Errorf("unsupported primitive type: %d", subMesh.Primitive))
		}
		meshAsset.SubMeshes[j] = subMeshAsset
	}
	return meshAsset
}

// TODO: Move to math lib
func matrixToArray(matrix sprec.Mat4) [16]float32 {
	var result [16]float32
	result[0] = matrix.M11
	result[1] = matrix.M21
	result[2] = matrix.M31
	result[3] = matrix.M41

	result[4] = matrix.M12
	result[5] = matrix.M22
	result[6] = matrix.M32
	result[7] = matrix.M42

	result[8] = matrix.M13
	result[9] = matrix.M23
	result[10] = matrix.M33
	result[11] = matrix.M43

	result[12] = matrix.M14
	result[13] = matrix.M24
	result[14] = matrix.M34
	result[15] = matrix.M44
	return result
}
