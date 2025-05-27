package conv

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/asset/conv/animationconv"
	"github.com/mokiat/lacking/game/asset/conv/backgroundconv"
	"github.com/mokiat/lacking/game/asset/conv/hierarchyconv"
	"github.com/mokiat/lacking/game/asset/conv/physicsconv"
	"github.com/mokiat/lacking/game/asset/conv/shadingconv"
	"github.com/mokiat/lacking/game/asset/dto/animationdto"
	"github.com/mokiat/lacking/game/asset/dto/backgrounddto"
	"github.com/mokiat/lacking/game/asset/dto/hierarchydto"
	"github.com/mokiat/lacking/game/asset/dto/lightingdto"
	"github.com/mokiat/lacking/game/asset/dto/meshdto"
	"github.com/mokiat/lacking/game/asset/dto/physicsdto"
	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/x448/float16"
)

func NewConverter(model *mdl.Model) *Converter {
	return &Converter{
		model: model,

		convertedNodes:           make(map[*mdl.Node]uint32),
		convertedArmatures:       make(map[*mdl.Armature]uint32),
		convertedGeometries:      make(map[*mdl.Geometry]uint32),
		convertedMeshDefinitions: make(map[*mdl.MeshDefinition]uint32),
	}
}

type Converter struct {
	model *mdl.Model

	convertedNodes map[*mdl.Node]uint32

	assetArmatures     []meshdto.Armature
	convertedArmatures map[*mdl.Armature]uint32

	assetGeometries     []meshdto.Geometry
	convertedGeometries map[*mdl.Geometry]uint32

	assetMeshDefinitions     []meshdto.MeshDefinition
	convertedMeshDefinitions map[*mdl.MeshDefinition]uint32
}

func (c *Converter) Convert() (asset.Model, error) {
	return c.convertModel(c.model)
}

func (c *Converter) convertModel(s *mdl.Model) (asset.Model, error) {
	var (
		assetMeshes            []meshdto.Mesh
		assetAmbientLights     []lightingdto.AmbientLight
		assetPointLights       []lightingdto.PointLight
		assetSpotLights        []lightingdto.SpotLight
		assetDirectionalLights []lightingdto.DirectionalLight
	)

	// First nodes pass, so that all nodes are tracked, otherwise
	// armature resolution will fail.
	for i, node := range s.NodesIter() {
		c.convertedNodes[node] = uint32(i)
	}

	for i, node := range s.NodesIter() {
		switch target := node.Target().(type) {
		case *mdl.Mesh:
			assetMesh, err := c.convertMesh(uint32(i), target)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting mesh %q: %w", node.Name(), err)
			}
			assetMeshes = append(assetMeshes, assetMesh)
		case *mdl.AmbientLight:
			ambientLightAsset, err := c.convertAmbientLight(uint32(i), target)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting ambient light %q: %w", node.Name(), err)
			}
			assetAmbientLights = append(assetAmbientLights, ambientLightAsset)
		case *mdl.PointLight:
			pointLightAsset := c.convertPointLight(uint32(i), target)
			assetPointLights = append(assetPointLights, pointLightAsset)
		case *mdl.SpotLight:
			spotLightAsset := c.convertSpotLight(uint32(i), target)
			assetSpotLights = append(assetSpotLights, spotLightAsset)
		case *mdl.DirectionalLight:
			directionalLightAsset := c.convertDirectionalLight(uint32(i), target)
			assetDirectionalLights = append(assetDirectionalLights, directionalLightAsset)
		}
	}

	return asset.Model{
		HierarchyChunkHolder: hierarchydto.HierarchyChunkHolder{
			HierarchyChunk: hierarchyconv.CreateHierarchyChunk(c.model),
		},
		AnimationChunkHolder: animationdto.AnimationChunkHolder{
			AnimationChunk: animationconv.CreateAnimationChunk(c.model),
		},
		ShadingChunkHolder: shadingdto.ShadingChunkHolder{
			ShadingChunk: gog.Must(shadingconv.CreateShadingChunk(c.model)),
		},
		MeshChunkHolder: meshdto.MeshChunkHolder{
			MeshChunk: &meshdto.MeshChunk{
				Armatures:       c.assetArmatures,
				Geometries:      c.assetGeometries,
				MeshDefinitions: c.assetMeshDefinitions,
				Meshes:          assetMeshes,
			},
		},
		PhysicsChunkHolder: physicsdto.PhysicsChunkHolder{
			PhysicsChunk: physicsconv.CreatePhysicsChunk(c.model),
		},
		LightingChunkHolder: lightingdto.LightingChunkHolder{
			LightingChunk: &lightingdto.LightingChunk{
				AmbientLights:     assetAmbientLights,
				PointLights:       assetPointLights,
				SpotLights:        assetSpotLights,
				DirectionalLights: assetDirectionalLights,
			},
		},
		BackgroundChunkHolder: backgrounddto.BackgroundChunkHolder{
			BackgroundChunk: gog.Must(backgroundconv.CreateBackgroundChunk(c.model)),
		},
	}, nil
}

func (c *Converter) convertMesh(nodeIndex uint32, mesh *mdl.Mesh) (meshdto.Mesh, error) {
	meshDefinitionIndex, err := c.convertMeshDefinition(mesh.Definition())
	if err != nil {
		return meshdto.Mesh{}, fmt.Errorf("error converting mesh definition: %w", err)
	}

	var armatureIndex = meshdto.UnspecifiedArmatureIndex
	if mesh.Armature() != nil {
		assetArmatureIndex, err := c.convertArmature(mesh.Armature())
		if err != nil {
			return meshdto.Mesh{}, fmt.Errorf("error converting armature: %w", err)
		}
		armatureIndex = int32(assetArmatureIndex)
	}

	return meshdto.Mesh{
		NodeIndex:           nodeIndex,
		MeshDefinitionIndex: meshDefinitionIndex,
		ArmatureIndex:       armatureIndex,
	}, nil
}

func (c *Converter) convertGeometry(geometry *mdl.Geometry) (uint32, error) {
	if index, ok := c.convertedGeometries[geometry]; ok {
		return index, nil
	}

	const (
		sizeUnsignedByte  = 1
		sizeUnsignedShort = 2
		sizeUnsignedInt   = 4
		sizeHalfFloat     = 2
		sizeFloat         = 4
	)

	var (
		stride              uint32
		coordBufferIndex    int32
		coordOffset         uint32
		normalBufferIndex   int32
		normalOffset        uint32
		tangentBufferIndex  int32
		tangentOffset       uint32
		texCoordBufferIndex int32
		texCoordOffset      uint32
		colorBufferIndex    int32
		colorOffset         uint32
		weightsBufferIndex  int32
		weightsOffset       uint32
		jointsBufferIndex   int32
		jointsOffset        uint32
	)

	layout := geometry.Format()
	if layout&mdl.VertexFormatCoord != 0 {
		coordBufferIndex = 0
		coordOffset = stride
		stride += 3 * sizeFloat
	} else {
		coordBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatNormal != 0 {
		normalBufferIndex = 0
		normalOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		normalBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatTangent != 0 {
		tangentBufferIndex = 0
		tangentOffset = stride
		stride += 3 * sizeHalfFloat
		stride += sizeHalfFloat // due to alignment requirements
	} else {
		tangentBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatTexCoord != 0 {
		texCoordBufferIndex = 0
		texCoordOffset = stride
		stride += 2 * sizeHalfFloat
	} else {
		texCoordBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatColor != 0 {
		colorBufferIndex = 0
		colorOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		colorBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatWeights != 0 {
		weightsBufferIndex = 0
		weightsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		weightsBufferIndex = meshdto.UnspecifiedBufferIndex
	}
	if layout&mdl.VertexFormatJoints != 0 {
		jointsBufferIndex = 0
		jointsOffset = stride
		stride += 4 * sizeUnsignedByte
	} else {
		jointsBufferIndex = meshdto.UnspecifiedBufferIndex
	}

	vertexData := gblob.LittleEndianBlock(make([]byte, len(geometry.Vertices())*int(stride)))
	if layout&mdl.VertexFormatCoord != 0 {
		offset := int(coordOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetFloat32(offset+0*sizeFloat, vertex.Coord.X)
			vertexData.SetFloat32(offset+1*sizeFloat, vertex.Coord.Y)
			vertexData.SetFloat32(offset+2*sizeFloat, vertex.Coord.Z)
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatNormal != 0 {
		offset := int(normalOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Normal.Z).Bits())
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatTangent != 0 {
		offset := int(tangentOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Y).Bits())
			vertexData.SetUint16(offset+2*sizeHalfFloat, float16.Fromfloat32(vertex.Tangent.Z).Bits())
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatTexCoord != 0 {
		offset := int(texCoordOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint16(offset+0*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.X).Bits())
			vertexData.SetUint16(offset+1*sizeHalfFloat, float16.Fromfloat32(vertex.TexCoord.Y).Bits())
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatColor != 0 {
		offset := int(colorOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Color.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Color.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Color.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Color.W*255.0))
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatWeights != 0 {
		offset := int(weightsOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Weights.X*255.0))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Weights.Y*255.0))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Weights.Z*255.0))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Weights.W*255.0))
			offset += int(stride)
		}
	}
	if layout&mdl.VertexFormatJoints != 0 {
		offset := int(jointsOffset)
		for _, vertex := range geometry.Vertices() {
			vertexData.SetUint8(offset+0*sizeUnsignedByte, uint8(vertex.Joints[0]))
			vertexData.SetUint8(offset+1*sizeUnsignedByte, uint8(vertex.Joints[1]))
			vertexData.SetUint8(offset+2*sizeUnsignedByte, uint8(vertex.Joints[2]))
			vertexData.SetUint8(offset+3*sizeUnsignedByte, uint8(vertex.Joints[3]))
			offset += int(stride)
		}
	}

	var (
		indexLayout meshdto.IndexLayout
		indexData   gblob.LittleEndianBlock
		indexSize   int
	)
	if len(geometry.Vertices()) >= 0xFFFF {
		indexSize = sizeUnsignedInt
		indexLayout = meshdto.IndexLayoutUint32
		indexData = gblob.LittleEndianBlock(make([]byte, len(geometry.Indices())*sizeUnsignedInt))
		for i, index := range geometry.Indices() {
			indexData.SetUint32(i*sizeUnsignedInt, uint32(index))
		}
	} else {
		indexSize = sizeUnsignedShort
		indexLayout = meshdto.IndexLayoutUint16
		indexData = gblob.LittleEndianBlock(make([]byte, len(geometry.Indices())*sizeUnsignedShort))
		for i, index := range geometry.Indices() {
			indexData.SetUint16(i*sizeUnsignedShort, uint16(index))
		}
	}

	assetFragments := make([]meshdto.Fragment, 0, len(geometry.Fragments()))
	for _, fragment := range geometry.Fragments() {
		assetFragments = append(assetFragments, meshdto.Fragment{
			Name:            fragment.Name(),
			Topology:        fragment.Topology(),
			IndexByteOffset: uint32(fragment.IndexOffset() * indexSize),
			IndexCount:      uint32(fragment.IndexCount()),
		})
	}

	var boundingSphereRadius float64
	for _, vertex := range geometry.Vertices() {
		boundingSphereRadius = dprec.Max(
			boundingSphereRadius,
			float64(vertex.Coord.Length()),
		)
	}

	assetGeometry := meshdto.Geometry{
		VertexBuffers: []meshdto.VertexBuffer{
			{
				Stride: stride,
				Data:   vertexData,
			},
		},
		VertexLayout: meshdto.VertexLayout{
			Coord: meshdto.VertexAttribute{
				BufferIndex: coordBufferIndex,
				ByteOffset:  coordOffset,
				Format:      meshdto.VertexAttributeFormatRGB32F,
			},
			Normal: meshdto.VertexAttribute{
				BufferIndex: normalBufferIndex,
				ByteOffset:  normalOffset,
				Format:      meshdto.VertexAttributeFormatRGB16F,
			},
			Tangent: meshdto.VertexAttribute{
				BufferIndex: tangentBufferIndex,
				ByteOffset:  tangentOffset,
				Format:      meshdto.VertexAttributeFormatRGB16F,
			},
			TexCoord: meshdto.VertexAttribute{
				BufferIndex: texCoordBufferIndex,
				ByteOffset:  texCoordOffset,
				Format:      meshdto.VertexAttributeFormatRG16F,
			},
			Color: meshdto.VertexAttribute{
				BufferIndex: colorBufferIndex,
				ByteOffset:  colorOffset,
				Format:      meshdto.VertexAttributeFormatRGBA8UN,
			},
			Weights: meshdto.VertexAttribute{
				BufferIndex: weightsBufferIndex,
				ByteOffset:  weightsOffset,
				Format:      meshdto.VertexAttributeFormatRGBA8UN,
			},
			Joints: meshdto.VertexAttribute{
				BufferIndex: jointsBufferIndex,
				ByteOffset:  jointsOffset,
				Format:      meshdto.VertexAttributeFormatRGBA8IU,
			},
		},
		IndexBuffer: meshdto.IndexBuffer{
			IndexLayout: indexLayout,
			Data:        indexData,
		},
		Fragments:            assetFragments,
		BoundingSphereRadius: boundingSphereRadius,
		MinDistance:          geometry.MinDistance(),
		MaxDistance:          geometry.MaxDistance(),
		MaxCascade:           uint8(geometry.MaxCascade()),
	}

	index := uint32(len(c.assetGeometries))
	c.assetGeometries = append(c.assetGeometries, assetGeometry)
	c.convertedGeometries[geometry] = index
	return index, nil
}

func (c *Converter) convertMeshDefinition(definition *mdl.MeshDefinition) (uint32, error) {
	if index, ok := c.convertedMeshDefinitions[definition]; ok {
		return index, nil
	}

	geometryIndex, err := c.convertGeometry(definition.Geometry())
	if err != nil {
		return 0, fmt.Errorf("error converting geometry: %w", err)
	}
	geometry := c.assetGeometries[geometryIndex]

	var materialBindings []meshdto.MaterialBinding
	for i, fragment := range geometry.Fragments {
		material, ok := definition.MaterialBindings()[fragment.Name]
		if !ok {
			continue // likely invisible fragment.
		}
		materialBindings = append(materialBindings, meshdto.MaterialBinding{
			FragmentIndex: uint32(i),
			MaterialID:    material.ID(),
		})
	}

	assetDefinition := meshdto.MeshDefinition{
		GeometryIndex:    geometryIndex,
		MaterialBindings: materialBindings,
	}

	index := uint32(len(c.assetMeshDefinitions))
	c.assetMeshDefinitions = append(c.assetMeshDefinitions, assetDefinition)
	c.convertedMeshDefinitions[definition] = index
	return index, nil
}

func (c *Converter) convertArmature(armature *mdl.Armature) (uint32, error) {
	if index, ok := c.convertedArmatures[armature]; ok {
		return index, nil
	}

	assetArmature := meshdto.Armature{
		Joints: gog.Map(armature.Joints(), func(joint *mdl.Joint) meshdto.Joint {
			return meshdto.Joint{
				NodeIndex:         c.convertedNodes[joint.Node()],
				InverseBindMatrix: joint.InverseBindMatrix(),
			}
		}),
	}

	index := uint32(len(c.assetArmatures))
	c.assetArmatures = append(c.assetArmatures, assetArmature)
	c.convertedArmatures[armature] = index
	return index, nil
}

func (c *Converter) convertAmbientLight(nodeIndex uint32, light *mdl.AmbientLight) (lightingdto.AmbientLight, error) {
	return lightingdto.AmbientLight{
		NodeIndex:           nodeIndex,
		ReflectionTextureID: light.ReflectionTexture().ID(),
		RefractionTextureID: light.RefractionTexture().ID(),
		CastShadow:          light.CastShadow(),
	}, nil
}

func (c *Converter) convertPointLight(nodeIndex uint32, light *mdl.PointLight) lightingdto.PointLight {
	return lightingdto.PointLight{
		NodeIndex:    nodeIndex,
		EmitColor:    light.EmitColor(),
		EmitDistance: light.EmitDistance(),
		CastShadow:   light.CastShadow(),
	}
}

func (c *Converter) convertSpotLight(nodeIndex uint32, light *mdl.SpotLight) lightingdto.SpotLight {
	return lightingdto.SpotLight{
		NodeIndex:      nodeIndex,
		EmitColor:      light.EmitColor(),
		EmitDistance:   light.EmitDistance(),
		EmitAngleOuter: light.EmitAngleOuter(),
		EmitAngleInner: light.EmitAngleInner(),
		CastShadow:     light.CastShadow(),
	}
}

func (c *Converter) convertDirectionalLight(nodeIndex uint32, light *mdl.DirectionalLight) lightingdto.DirectionalLight {
	return lightingdto.DirectionalLight{
		NodeIndex:  nodeIndex,
		EmitColor:  light.EmitColor(),
		CastShadow: light.CastShadow(),
	}
}
