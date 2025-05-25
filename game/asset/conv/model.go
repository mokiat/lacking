package conv

import (
	"fmt"
	"slices"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/asset/conv/hierarchyconv"
	"github.com/mokiat/lacking/game/asset/dto/animationdto"
	"github.com/mokiat/lacking/game/asset/dto/backgrounddto"
	"github.com/mokiat/lacking/game/asset/dto/hierarchydto"
	"github.com/mokiat/lacking/game/asset/dto/lightingdto"
	"github.com/mokiat/lacking/game/asset/dto/meshdto"
	"github.com/mokiat/lacking/game/asset/dto/physicsdto"
	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/x448/float16"
)

func NewConverter(model *mdl.Model) *Converter {
	return &Converter{
		model: model,

		convertedNodes:           make(map[*mdl.Node]uint32),
		convertedArmatures:       make(map[*mdl.Armature]uint32),
		convertedShaders:         make(map[*mdl.Shader]uint32),
		convertedTextures:        make(map[*mdl.Texture]uint32),
		convertedMaterials:       make(map[*mdl.Material]uint32),
		convertedGeometries:      make(map[*mdl.Geometry]uint32),
		convertedMeshDefinitions: make(map[*mdl.MeshDefinition]uint32),
		convertedBodyMaterials:   make(map[*mdl.BodyMaterial]uint32),
		convertedBodyDefinitions: make(map[*mdl.BodyDefinition]uint32),
	}
}

type Converter struct {
	model *mdl.Model

	convertedNodes map[*mdl.Node]uint32

	assetArmatures     []meshdto.Armature
	convertedArmatures map[*mdl.Armature]uint32

	assetShaders     []shadingdto.Shader
	convertedShaders map[*mdl.Shader]uint32

	assetTextures     []shadingdto.Texture
	convertedTextures map[*mdl.Texture]uint32

	assetMaterials     []shadingdto.Material
	convertedMaterials map[*mdl.Material]uint32

	assetGeometries     []meshdto.Geometry
	convertedGeometries map[*mdl.Geometry]uint32

	assetMeshDefinitions     []meshdto.MeshDefinition
	convertedMeshDefinitions map[*mdl.MeshDefinition]uint32

	assetBodyMaterials     []physicsdto.BodyMaterial
	convertedBodyMaterials map[*mdl.BodyMaterial]uint32

	assetBodyDefinitions     []physicsdto.BodyDefinition
	convertedBodyDefinitions map[*mdl.BodyDefinition]uint32
}

func (c *Converter) Convert() (asset.Model, error) {
	return c.convertModel(c.model)
}

func (c *Converter) convertModel(s *mdl.Model) (asset.Model, error) {
	var (
		assetMeshes            []meshdto.Mesh
		assetBodies            []physicsdto.Body
		assetAmbientLights     []lightingdto.AmbientLight
		assetPointLights       []lightingdto.PointLight
		assetSpotLights        []lightingdto.SpotLight
		assetDirectionalLights []lightingdto.DirectionalLight
		assetSkies             []backgrounddto.Sky
	)

	// First nodes pass, so that all nodes are tracked, otherwise
	// armature resolution will fail.
	for i, node := range s.NodesIter() {
		c.convertedNodes[node] = uint32(i)
	}

	for i, node := range s.NodesIter() {
		switch source := node.Source().(type) {
		case *mdl.Body:
			assetBody, err := c.convertBody(uint32(i), source)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting body %q: %w", node.Name(), err)
			}
			assetBodies = append(assetBodies, assetBody)
		}
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
		case *mdl.Sky:
			assetSky, err := c.convertSky(uint32(i), target)
			if err != nil {
				return asset.Model{}, fmt.Errorf("error converting sky %q: %w", node.Name(), err)
			}
			assetSkies = append(assetSkies, assetSky)
		}
	}

	assetAnimations := make([]animationdto.Animation, len(c.model.Animations()))
	for i, animation := range c.model.Animations() {
		assetAnimations[i] = c.convertAnimation(animation)
	}

	return asset.Model{
		HierarchyChunkHolder: hierarchydto.HierarchyChunkHolder{
			HierarchyChunk: hierarchyconv.CreateHierarchyChunk(c.model),
		},
		AnimationChunkHolder: animationdto.AnimationChunkHolder{
			AnimationChunk: &animationdto.AnimationChunk{
				Animations: assetAnimations,
			},
		},
		ShadingChunkHolder: shadingdto.ShadingChunkHolder{
			ShadingChunk: &shadingdto.ShadingChunk{
				Shaders:   c.assetShaders,
				Textures:  c.assetTextures,
				Materials: c.assetMaterials,
			},
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
			PhysicsChunk: &physicsdto.PhysicsChunk{
				BodyMaterials:   c.assetBodyMaterials,
				BodyDefinitions: c.assetBodyDefinitions,
				Bodies:          assetBodies,
			},
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
			BackgroundChunk: &backgrounddto.BackgroundChunk{
				Skies: assetSkies,
			},
		},
	}, nil
}

func (c *Converter) convertAnimation(animation *mdl.Animation) animationdto.Animation {
	assetAnimation := animationdto.Animation{
		Name:      animation.Name(),
		StartTime: animation.StartTime(),
		EndTime:   animation.EndTime(),
		Bindings:  make([]animationdto.AnimationBinding, len(animation.Bindings())),
	}
	for i, binding := range animation.Bindings() {
		translationKeyframes := make([]animationdto.AnimationKeyframe[dprec.Vec3], len(binding.TranslationKeyframes()))
		for j, keyframe := range binding.TranslationKeyframes() {
			translationKeyframes[j] = animationdto.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		rotationKeyframes := make([]animationdto.AnimationKeyframe[dprec.Quat], len(binding.RotationKeyframes()))
		for j, keyframe := range binding.RotationKeyframes() {
			rotationKeyframes[j] = animationdto.AnimationKeyframe[dprec.Quat]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		scaleKeyframes := make([]animationdto.AnimationKeyframe[dprec.Vec3], len(binding.ScaleKeyframes()))
		for j, keyframe := range binding.ScaleKeyframes() {
			scaleKeyframes[j] = animationdto.AnimationKeyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		assetAnimation.Bindings[i] = animationdto.AnimationBinding{
			NodeName:             binding.NodeName(),
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}
	return assetAnimation
}

func (c *Converter) convertMaterialPass(pass *mdl.MaterialPass) (shadingdto.MaterialPass, error) {
	shaderIndex, err := c.convertShader(pass.Shader())
	if err != nil {
		return shadingdto.MaterialPass{}, fmt.Errorf("error converting shader: %w", err)
	}
	return shadingdto.MaterialPass{
		Layer:           int32(pass.Layer()),
		Culling:         pass.Culling(),
		FrontFace:       pass.FrontFace(),
		DepthTest:       pass.DepthTest(),
		DepthWrite:      pass.DepthWrite(),
		DepthComparison: pass.DepthComparison(),
		Blending:        pass.Blending(),
		ShaderIndex:     shaderIndex,
	}, nil
}

func (c *Converter) convertMaterial(material *mdl.Material) (uint32, error) {
	if index, ok := c.convertedMaterials[material]; ok {
		return index, nil
	}

	textures, err := c.convertSamplers(material.Samplers())
	if err != nil {
		return 0, fmt.Errorf("error converting samplers: %w", err)
	}

	properties, err := c.convertProperties(material.Properties())
	if err != nil {
		return 0, fmt.Errorf("error converting properties: %w", err)
	}

	assetMaterial := shadingdto.Material{
		Name:                 material.Name(),
		Textures:             textures,
		Properties:           properties,
		GeometryPasses:       make([]shadingdto.MaterialPass, len(material.GeometryPasses())),
		ShadowPasses:         make([]shadingdto.MaterialPass, len(material.ShadowPasses())),
		ForwardPasses:        make([]shadingdto.MaterialPass, len(material.ForwardPasses())),
		SkyPasses:            make([]shadingdto.MaterialPass, len(material.SkyPasses())),
		PostprocessingPasses: make([]shadingdto.MaterialPass, len(material.PostprocessingPasses())),
	}
	for i, pass := range material.GeometryPasses() {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.GeometryPasses[i] = assetPass
	}
	for i, pass := range material.ShadowPasses() {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.ShadowPasses[i] = assetPass
	}
	for i, pass := range material.ForwardPasses() {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.ForwardPasses[i] = assetPass
	}
	for i, pass := range material.SkyPasses() {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.SkyPasses[i] = assetPass
	}
	for i, pass := range material.PostprocessingPasses() {
		assetPass, err := c.convertMaterialPass(pass)
		if err != nil {
			return 0, fmt.Errorf("error converting material pass: %w", err)
		}
		assetMaterial.PostprocessingPasses[i] = assetPass
	}

	index := uint32(len(c.assetMaterials))
	c.assetMaterials = append(c.assetMaterials, assetMaterial)
	c.convertedMaterials[material] = index
	return index, nil
}

func (c *Converter) convertBodyMaterial(material *mdl.BodyMaterial) (uint32, error) {
	if index, ok := c.convertedBodyMaterials[material]; ok {
		return index, nil
	}

	assetMaterial := physicsdto.BodyMaterial{
		FrictionCoefficient:    material.FrictionCoefficient(),
		RestitutionCoefficient: material.RestitutionCoefficient(),
	}

	index := uint32(len(c.assetBodyMaterials))
	c.assetBodyMaterials = append(c.assetBodyMaterials, assetMaterial)
	c.convertedBodyMaterials[material] = index
	return index, nil
}

func (c *Converter) convertBodyDefinition(definition *mdl.BodyDefinition) (uint32, error) {
	if index, ok := c.convertedBodyDefinitions[definition]; ok {
		return index, nil
	}

	materialIndex, err := c.convertBodyMaterial(definition.Material())
	if err != nil {
		return 0, fmt.Errorf("error converting body material: %w", err)
	}

	assetDefinition := physicsdto.BodyDefinition{
		MaterialIndex:     materialIndex,
		Mass:              definition.Mass(),
		MomentOfInertia:   definition.MomentOfInertia(),
		DragFactor:        definition.DragFactor(),
		AngularDragFactor: definition.AngularDragFactor(),
		CollisionBoxes: gog.Map(definition.CollisionBoxes(), func(box *mdl.CollisionBox) physicsdto.CollisionBox {
			return physicsdto.CollisionBox{
				Translation: box.Translation(),
				Rotation:    box.Rotation(),
				Width:       box.Width(),
				Height:      box.Height(),
				Length:      box.Length(),
			}
		}),
		CollisionSpheres: gog.Map(definition.CollisionSpheres(), func(sphere *mdl.CollisionSphere) physicsdto.CollisionSphere {
			return physicsdto.CollisionSphere{
				Translation: sphere.Translation(),
				Radius:      sphere.Radius(),
			}
		}),
		CollisionMeshes: gog.Map(definition.CollisionMeshes(), func(mesh *mdl.CollisionMesh) physicsdto.CollisionMesh {
			return physicsdto.CollisionMesh{
				Translation: mesh.Translation(),
				Rotation:    mesh.Rotation(),
				Triangles: gog.Map(mesh.Triangles(), func(triangle mdl.CollisionTriangle) physicsdto.CollisionTriangle {
					return physicsdto.CollisionTriangle{
						A: triangle.A,
						B: triangle.B,
						C: triangle.C,
					}
				}),
			}
		}),
	}

	index := uint32(len(c.assetBodyDefinitions))
	c.assetBodyDefinitions = append(c.assetBodyDefinitions, assetDefinition)
	c.convertedBodyDefinitions[definition] = index
	return index, nil
}

func (c *Converter) convertBody(nodeIndex uint32, body *mdl.Body) (physicsdto.Body, error) {
	bodyDefinitionIndex, err := c.convertBodyDefinition(body.Definition())
	if err != nil {
		return physicsdto.Body{}, fmt.Errorf("error converting body definition: %w", err)
	}
	return physicsdto.Body{
		NodeIndex:           nodeIndex,
		BodyDefinitionIndex: bodyDefinitionIndex,
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
		materialIndex, err := c.convertMaterial(material)
		if err != nil {
			return 0, fmt.Errorf("error converting material: %w", err)
		}
		materialBindings = append(materialBindings, meshdto.MaterialBinding{
			FragmentIndex: uint32(i),
			MaterialIndex: materialIndex,
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
	reflectionTextureIndex, err := c.convertTexture(light.ReflectionTexture())
	if err != nil {
		return lightingdto.AmbientLight{}, fmt.Errorf("error converting reflection texture: %w", err)
	}

	refractionTextureIndex, err := c.convertTexture(light.RefractionTexture())
	if err != nil {
		return lightingdto.AmbientLight{}, fmt.Errorf("error converting refraction texture: %w", err)
	}

	return lightingdto.AmbientLight{
		NodeIndex:              nodeIndex,
		ReflectionTextureIndex: reflectionTextureIndex,
		RefractionTextureIndex: refractionTextureIndex,
		CastShadow:             light.CastShadow(),
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

func (c *Converter) convertSky(nodeIndex uint32, sky *mdl.Sky) (backgrounddto.Sky, error) {
	materialIndex, err := c.convertMaterial(sky.Material())
	if err != nil {
		return backgrounddto.Sky{}, fmt.Errorf("error converting material: %w", err)
	}

	assetSky := backgrounddto.Sky{
		NodeIndex:     nodeIndex,
		MaterialIndex: materialIndex,
	}
	return assetSky, nil
}

func (c *Converter) convertShader(shader *mdl.Shader) (uint32, error) {
	if index, ok := c.convertedShaders[shader]; ok {
		return index, nil
	}
	ast, err := lsl.Parse(shader.SourceCode())
	if err != nil {
		return 0, fmt.Errorf("error parsing shader: %w", err)
	}
	var schema lsl.Schema
	switch shader.ShaderType() {
	case mdl.ShaderTypeGeometry:
		schema = lsl.GeometrySchema()
	case mdl.ShaderTypeShadow:
		schema = lsl.ShadowSchema()
	case mdl.ShaderTypeForward:
		schema = lsl.ForwardSchema()
	case mdl.ShaderTypeSky:
		schema = lsl.SkySchema()
	case mdl.ShaderTypePostprocess:
		schema = lsl.PostprocessSchema()
	default:
		schema = lsl.DefaultSchema()
	}
	if err := lsl.Validate(ast, schema); err != nil {
		return 0, fmt.Errorf("error validating shader: %w", err)
	}
	shaderIndex := uint32(len(c.assetShaders))
	assetShader := shadingdto.Shader{
		ShaderType: shader.ShaderType(),
		SourceCode: shader.SourceCode(),
	}
	c.convertedShaders[shader] = shaderIndex
	c.assetShaders = append(c.assetShaders, assetShader)
	return shaderIndex, nil
}

func (c *Converter) convertSamplers(samplers map[string]*mdl.Sampler) ([]shadingdto.TextureBinding, error) {
	bindings := make([]shadingdto.TextureBinding, 0, len(samplers))
	for name, sampler := range samplers {
		textureIndex, err := c.convertTexture(sampler.Texture())
		if err != nil {
			return nil, fmt.Errorf("error converting texture: %w", err)
		}
		bindings = append(bindings, shadingdto.TextureBinding{
			BindingName:  name,
			TextureIndex: textureIndex,
			Wrapping:     sampler.WrapMode(),
			Filtering:    sampler.FilterMode(),
			Mipmapping:   sampler.Mipmapping(),
		})
	}
	return bindings, nil
}

func isLikelyLinearSpace(format mdl.TextureFormat) bool {
	linearFormats := []mdl.TextureFormat{
		mdl.TextureFormatRGBA16F,
		mdl.TextureFormatRGBA32F,
	}
	return slices.Contains(linearFormats, format)
}

func (c *Converter) convertTexture(texture *mdl.Texture) (uint32, error) {
	if index, ok := c.convertedTextures[texture]; ok {
		return index, nil
	}

	var flags shadingdto.TextureFlag
	switch texture.Kind() {
	case mdl.TextureKind2D:
		flags = shadingdto.TextureFlag2D
	case mdl.TextureKind2DArray:
		flags = shadingdto.TextureFlag2DArray
	case mdl.TextureKind3D:
		flags = shadingdto.TextureFlag3D
	case mdl.TextureKindCube:
		flags = shadingdto.TextureFlagCubeMap
	default:
		return 0, fmt.Errorf("unsupported texture kind %d", texture.Kind())
	}
	if isLikelyLinearSpace(texture.Format()) || texture.Linear() {
		flags |= shadingdto.TextureFlagLinearSpace
	}
	if texture.GenerateMipmaps() {
		flags |= shadingdto.TextureFlagMipmapping
	}
	assetTexture := shadingdto.Texture{
		Format: texture.Format(),
		Flags:  flags,
		MipmapLayers: gog.Map(texture.MipmapLayers(), func(mipLayer mdl.MipmapLayer) shadingdto.MipmapLayer {
			return shadingdto.MipmapLayer{
				Width:  uint32(mipLayer.Width()),
				Height: uint32(mipLayer.Height()),
				Depth:  uint32(mipLayer.Depth()),
				Layers: gog.Map(mipLayer.Layers(), func(layer mdl.TextureLayer) shadingdto.TextureLayer {
					return shadingdto.TextureLayer{
						Data: layer.Data(),
					}
				}),
			}
		}),
	}

	index := uint32(len(c.assetTextures))
	c.assetTextures = append(c.assetTextures, assetTexture)
	c.convertedTextures[texture] = index
	return index, nil
}

func (c *Converter) convertProperties(properties map[string]interface{}) ([]shadingdto.PropertyBinding, error) {
	bindings := make([]shadingdto.PropertyBinding, 0, len(properties))
	for name, value := range properties {
		var data gblob.LittleEndianBlock
		switch value := value.(type) {
		case float32:
			data = make(gblob.LittleEndianBlock, 4)
			data.SetFloat32(0, value)
		case sprec.Vec2:
			data = make(gblob.LittleEndianBlock, 8)
			data.SetFloat32(0, value.X)
			data.SetFloat32(4, value.Y)
		case sprec.Vec3:
			data = make(gblob.LittleEndianBlock, 12)
			data.SetFloat32(0, value.X)
			data.SetFloat32(4, value.Y)
			data.SetFloat32(8, value.Z)
		case sprec.Vec4:
			data = make(gblob.LittleEndianBlock, 16)
			data.SetFloat32(0, value.X)
			data.SetFloat32(4, value.Y)
			data.SetFloat32(8, value.Z)
			data.SetFloat32(12, value.W)
		default:
			return nil, fmt.Errorf("unsupported property type %T", value)
		}
		bindings = append(bindings, shadingdto.PropertyBinding{
			BindingName: name,
			Data:        data,
		})
	}
	return bindings, nil
}
