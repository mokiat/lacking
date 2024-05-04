package dsl

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/debug/log"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/util/gltfutil"
	"github.com/qmuntal/gltf"
	lightspunctual "github.com/qmuntal/gltf/ext/lightspuntual"
)

// CreateModel creates a new model with the specified name and operations.
func CreateModel(name string, operations ...Operation) Provider[*mdl.Model] {
	if _, ok := modelProviders[name]; ok {
		panic(fmt.Sprintf("provider for model %q already exists", name))
	}

	modelProviders[name] = OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Model, error) {
			var model mdl.Model
			model.SetName(name)
			for _, operation := range operations {
				if err := operation.Apply(&model); err != nil {
					return nil, err
				}
			}
			return &model, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-model", name, operations)
		},
	))

	return modelProviders[name]
}

// OpenGLTFModel creates a new model provider that loads a model from the
// specified path.
func OpenGLTFModel(path string) Provider[*mdl.Model] {
	return FuncProvider(
		// get function
		func() (*mdl.Model, error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open model file %q: %w", path, err)
			}
			defer file.Close()

			model, err := parseGLTFResource(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse gltf model: %w", err)
			}
			return model, nil
		},

		// digest function
		func() ([]byte, error) {
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file %q: %w", path, err)
			}
			return CreateDigest("opengl-gltf-model", path, info.ModTime())
		},
	)
}

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

func parseGLTFResource(in io.Reader) (*mdl.Model, error) {
	gltfDoc := new(gltf.Document)
	if err := gltf.NewDecoder(in).Decode(gltfDoc); err != nil {
		return nil, fmt.Errorf("failed to parse gltf model: %w", err)
	}
	return BuildModelResource(gltfDoc)
}

func BuildModelResource(gltfDoc *gltf.Document) (*mdl.Model, error) {
	model := &mdl.Model{}

	// build images
	imagesFromIndex := make(map[uint32]*mdl.Image)
	for i, gltfImage := range gltfDoc.Images {
		img, err := openGLTFImage(gltfDoc, gltfImage)
		if err != nil {
			return nil, fmt.Errorf("error loading image: %w", err)
		}
		imagesFromIndex[uint32(i)] = img
	}

	// build textures
	texturesFromIndex := make(map[uint32]*mdl.Texture)
	for i, img := range imagesFromIndex {
		texture := &mdl.Texture{}
		texture.SetKind(mdl.TextureKind2D)
		texture.SetFormat(mdl.TextureFormatRGBA8)
		texture.Resize(img.Width(), img.Height())
		texture.SetLayerImage(0, img)
		texturesFromIndex[i] = texture
	}

	// build samplers
	samplersFromIndex := make(map[uint32]*mdl.Sampler)
	for i, gltfTexture := range gltfDoc.Textures {
		sampler := &mdl.Sampler{}
		if gltfTexture.Sampler != nil {
			gltfSampler := gltfDoc.Samplers[*gltfTexture.Sampler]
			switch gltfSampler.WrapS {
			case gltf.WrapRepeat:
				sampler.SetWrapMode(mdl.WrapModeRepeat)
			case gltf.WrapClampToEdge:
				sampler.SetWrapMode(mdl.WrapModeClamp)
			case gltf.WrapMirroredRepeat:
				sampler.SetWrapMode(mdl.WrapModeMirroredRepeat)
			default:
				sampler.SetWrapMode(mdl.WrapModeClamp)
				log.Warn("Unsupported texture wrap mode: %v", gltfSampler.WrapS)
			}
			switch gltfSampler.MinFilter {
			case gltf.MinNearest:
				sampler.SetFilterMode(mdl.FilterModeNearest)
			case gltf.MinLinear:
				sampler.SetFilterMode(mdl.FilterModeLinear)
			case gltf.MinNearestMipMapNearest:
				sampler.SetMipmapping(true)
				sampler.SetFilterMode(mdl.FilterModeNearest)
			case gltf.MinLinearMipMapNearest:
				sampler.SetMipmapping(true)
				sampler.SetFilterMode(mdl.FilterModeLinear)
			case gltf.MinNearestMipMapLinear:
				sampler.SetMipmapping(true)
				sampler.SetFilterMode(mdl.FilterModeNearest)
			case gltf.MinLinearMipMapLinear:
				sampler.SetMipmapping(true)
				sampler.SetFilterMode(mdl.FilterModeLinear)
			default:
				sampler.SetFilterMode(mdl.FilterModeLinear)
				log.Warn("Unsupported texture min filter mode: %v", gltfSampler.MinFilter)
			}
		} else {
			sampler.SetFilterMode(mdl.FilterModeLinear)
			sampler.SetWrapMode(mdl.WrapModeRepeat)
			sampler.SetMipmapping(true)
		}
		if gltfTexture.Source != nil {
			sampler.SetTexture(texturesFromIndex[*gltfTexture.Source])
		} else {
			return nil, fmt.Errorf("texture source not set")
		}
		samplersFromIndex[uint32(i)] = sampler
	}

	// build materials
	materialFromIndex := make(map[uint32]*mdl.Material)
	for i, gltfMaterial := range gltfDoc.Materials {
		var (
			color          sprec.Vec4
			metallic       float32
			roughness      float32
			normalScale    float32
			alphaThreshold float32

			colorTextureIndex             *uint32
			metallicRoughnessTextureIndex *uint32
			normalTextureIndex            *uint32
		)

		if gltfPBR := gltfMaterial.PBRMetallicRoughness; gltfPBR != nil {
			color = gltfutil.BaseColor(gltfPBR)
			metallic = float32(gltfPBR.MetallicFactorOrDefault())
			roughness = float32(gltfPBR.RoughnessFactorOrDefault())
			if texIndex := gltfutil.ColorTextureIndex(gltfDoc, gltfPBR); texIndex != nil {
				colorTextureIndex = texIndex
			}
			if texIndex := gltfutil.MetallicRoughnessTextureIndex(gltfDoc, gltfPBR); texIndex != nil {
				metallicRoughnessTextureIndex = texIndex
			}
		} else {
			color = sprec.NewVec4(1.0, 1.0, 1.0, 1.0)
			metallic = 1.0
			roughness = 1.0
		}

		alphaThreshold = float32(gltfMaterial.AlphaCutoffOrDefault())

		if texIndex, texScale := gltfutil.NormalTextureIndexScale(gltfDoc, gltfMaterial); texIndex != nil {
			normalTextureIndex = texIndex
			normalScale = texScale
		} else {
			normalScale = 1.0
		}

		geometryShader := mdl.NewShader(mdl.ShaderTypeGeometry)
		geometryShader.SetSourceCode(createPBRShader(pbrShaderConfig{
			hasColorTexture:             colorTextureIndex != nil,
			hasMetallicRoughnessTexture: metallicRoughnessTextureIndex != nil,
			hasNormalTexture:            normalTextureIndex != nil,
			hasAlphaTesting:             gltfMaterial.AlphaMode == gltf.AlphaMask,
		}))

		geometryPass := mdl.NewMaterialPass()
		geometryPass.SetLayer(0)
		if gltfMaterial.DoubleSided {
			geometryPass.SetCulling(mdl.CullModeNone)
		} else {
			geometryPass.SetCulling(mdl.CullModeBack)
		}
		geometryPass.SetFrontFace(mdl.FaceOrientationCCW)
		geometryPass.SetDepthTest(true)
		geometryPass.SetDepthWrite(true)
		geometryPass.SetDepthComparison(mdl.ComparisonLessOrEqual)
		geometryPass.SetBlending(false) // if gltfMaterial.AlphaMode == gltf.AlphaBlend, use forward pass somehow
		geometryPass.SetShader(geometryShader)

		shadowShader := mdl.NewShader(mdl.ShaderTypeGeometry)
		shadowShader.SetSourceCode(``)

		shadowPass := mdl.NewMaterialPass()
		shadowPass.SetLayer(0)
		if gltfMaterial.DoubleSided {
			shadowPass.SetCulling(mdl.CullModeNone)
		} else {
			shadowPass.SetCulling(mdl.CullModeBack)
		}
		shadowPass.SetFrontFace(mdl.FaceOrientationCCW)
		shadowPass.SetDepthTest(true)
		shadowPass.SetDepthWrite(true)
		shadowPass.SetDepthComparison(mdl.ComparisonLessOrEqual)
		shadowPass.SetBlending(false) // if gltfMaterial.AlphaMode == gltf.AlphaBlend, use forward pass somehow
		shadowPass.SetShader(shadowShader)

		material := &mdl.Material{}
		material.SetName(gltfMaterial.Name)
		material.SetMetadata(gltfutil.Properties(gltfMaterial.Extras))
		material.AddGeometryPass(geometryPass)
		material.AddShadowPass(shadowPass)
		material.SetProperty("color", color)
		material.SetProperty("metallic", metallic)
		material.SetProperty("roughness", roughness)
		material.SetProperty("normalScale", normalScale)
		material.SetProperty("alphaThreshold", alphaThreshold)
		if colorTextureIndex != nil {
			material.SetSampler("colorSampler", samplersFromIndex[*colorTextureIndex])
		}
		if metallicRoughnessTextureIndex != nil {
			material.SetSampler("metallicRoughnessSampler", samplersFromIndex[*metallicRoughnessTextureIndex])
		}
		if normalTextureIndex != nil {
			material.SetSampler("normalSampler", samplersFromIndex[*normalTextureIndex])
		}

		materialFromIndex[uint32(i)] = material
	}

	// build mesh definitions
	meshDefinitionFromIndex := make(map[uint32]*mdl.MeshDefinition)
	for i, gltfMesh := range gltfDoc.Meshes {
		geometry := &mdl.Geometry{}
		geometry.SetName(gltfMesh.Name)
		geometry.SetMetadata(gltfutil.Properties(gltfMesh.Extras))

		meshDefinition := &mdl.MeshDefinition{}
		meshDefinition.SetName(gltfMesh.Name)
		meshDefinition.SetGeometry(geometry)

		indexFromVertex := make(map[mdl.Vertex]int)

		for _, gltfPrimitive := range gltfMesh.Primitives {
			indexOffset := geometry.IndexOffset() // this needs to happen first

			gltfIndices, err := gltfutil.Indices(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading indices: %w", err)
			}
			gltfCoords, err := gltfutil.Coords(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading coords: %w", err)
			}
			gltfNormals, err := gltfutil.Normals(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading normals: %w", err)
			}
			gltfTangents, err := gltfutil.Tangents(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading tangents: %w", err)
			}
			gltfTexCoords, err := gltfutil.TexCoord0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading tex coords: %w", err)
			}
			gltfColors, err := gltfutil.Color0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading colors: %w", err)
			}
			gltfWeights, err := gltfutil.Weight0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading weights: %w", err)
			}
			gltfJoints, err := gltfutil.Joint0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return nil, fmt.Errorf("error reading joints: %w", err)
			}

			geometryFormat := geometry.Format()
			if gltfCoords != nil {
				geometryFormat |= mdl.VertexFormatCoord
			}
			if gltfNormals != nil {
				geometryFormat |= mdl.VertexFormatNormal
			}
			if gltfTangents != nil {
				geometryFormat |= mdl.VertexFormatTangent
			}
			if gltfTexCoords != nil {
				geometryFormat |= mdl.VertexFormatTexCoord
			}
			if gltfColors != nil {
				geometryFormat |= mdl.VertexFormatColor
			}
			if gltfWeights != nil {
				geometryFormat |= mdl.VertexFormatWeights
			}
			if gltfJoints != nil {
				geometryFormat |= mdl.VertexFormatJoints
			}
			geometry.SetFormat(geometryFormat)

			for _, gltfIndex := range gltfIndices {
				var vertex mdl.Vertex
				if gltfCoords != nil {
					vertex.Coord = gltfCoords[gltfIndex]
				}
				if gltfNormals != nil {
					vertex.Normal = gltfNormals[gltfIndex]
				}
				if gltfTangents != nil {
					vertex.Tangent = gltfTangents[gltfIndex]
				}
				if gltfTexCoords != nil {
					vertex.TexCoord = gltfTexCoords[gltfIndex]
				}
				if gltfColors != nil {
					vertex.Color = gltfColors[gltfIndex]
				}
				if gltfWeights != nil {
					vertex.Weights = gltfWeights[gltfIndex]
				}
				if gltfJoints != nil {
					vertex.Joints = gltfJoints[gltfIndex]
				}

				if index, ok := indexFromVertex[vertex]; ok {
					geometry.AddIndex(index)
				} else {
					index = geometry.VertexOffset()
					geometry.AddVertex(vertex)
					geometry.AddIndex(index)
					indexFromVertex[vertex] = index
				}
			}

			fragment := &mdl.Fragment{}
			fragment.SetIndexOffset(indexOffset)
			fragment.SetIndexCount(len(gltfIndices))
			switch gltfPrimitive.Mode {
			case gltf.PrimitivePoints:
				fragment.SetTopology(mdl.TopologyPoints)
			case gltf.PrimitiveLines:
				fragment.SetTopology(mdl.TopologyLineList)
			case gltf.PrimitiveLineStrip:
				fragment.SetTopology(mdl.TopologyLineStrip)
			case gltf.PrimitiveTriangles:
				fragment.SetTopology(mdl.TopologyTriangleList)
			case gltf.PrimitiveTriangleStrip:
				fragment.SetTopology(mdl.TopologyTriangleStrip)
			default:
				return nil, fmt.Errorf("unsupported primitive mode %d", gltfPrimitive.Mode)
			}
			if gltfPrimitive.Material != nil {
				gltfMaterial := gltfDoc.Materials[*gltfPrimitive.Material]
				fragment.SetName(gltfMaterial.Name)
				meshDefinition.BindMaterial(gltfMaterial.Name, materialFromIndex[*gltfPrimitive.Material])
			} else {
				return nil, fmt.Errorf("missing material for primitive")
			}
			geometry.AddFragment(fragment)
		}

		meshDefinition.SetName(gltfMesh.Name)
		meshDefinition.SetGeometry(geometry)
		meshDefinitionFromIndex[uint32(i)] = meshDefinition
	}

	// prepare armatures
	armatureFromIndex := make(map[uint32]*mdl.Armature)
	for i := range gltfDoc.Skins {
		armatureFromIndex[uint32(i)] = mdl.NewArmature()
	}

	createPointLight := func(gltfLight *lightspunctual.Light) *mdl.PointLight {
		emitColor := dprec.ArrayToVec3(gltfLight.ColorOrDefault())
		emitColor = dprec.Vec3Prod(emitColor, gltfLight.IntensityOrDefault())
		emitDistance := gog.ValueOf(gltfLight.Range, 10.0)

		light := mdl.NewPointLight()
		light.SetEmitColor(emitColor)
		light.SetEmitDistance(emitDistance)
		light.SetCastShadow(false)
		return light
	}

	createSpotLight := func(gltfLight *lightspunctual.Light) *mdl.SpotLight {
		emitColor := dprec.ArrayToVec3(gltfLight.ColorOrDefault())
		emitColor = dprec.Vec3Prod(emitColor, float64(gltfLight.IntensityOrDefault()))
		emitDistance := gog.ValueOf(gltfLight.Range, 10.0)
		emitInnerConeAngle := dprec.Radians(gltfLight.Spot.InnerConeAngle)
		emitOuterConeAngle := dprec.Radians(gltfLight.Spot.OuterConeAngleOrDefault())

		light := mdl.NewSpotLight()
		light.SetEmitColor(emitColor)
		light.SetEmitDistance(emitDistance)
		light.SetEmitAngleInner(emitInnerConeAngle)
		light.SetEmitAngleOuter(emitOuterConeAngle)
		light.SetCastShadow(false)
		return light
	}

	createDirectionalLight := func(gltfLight *lightspunctual.Light) *mdl.DirectionalLight {
		emitColor := dprec.ArrayToVec3(gltfLight.ColorOrDefault())
		emitColor = dprec.Vec3Prod(emitColor, gltfLight.IntensityOrDefault())

		light := mdl.NewDirectionalLight()
		light.SetEmitColor(emitColor)
		light.SetCastShadow(false)
		return light
	}

	createLight := func(gltfNode *gltf.Node) any {
		gltfDocLights := gltfDoc.Extensions[lightspunctual.ExtensionName].(lightspunctual.Lights)
		gltfNodeLight := gltfNode.Extensions[lightspunctual.ExtensionName].(lightspunctual.LightIndex)

		gltfLight := gltfDocLights[gltfNodeLight]
		switch gltfLight.Type {
		case lightspunctual.TypePoint:
			return createPointLight(gltfLight)
		case lightspunctual.TypeSpot:
			return createSpotLight(gltfLight)
		case lightspunctual.TypeDirectional:
			return createDirectionalLight(gltfLight)
		default:
			// FIXME: Return an error
			panic(fmt.Errorf("unsupported light type %q", gltfLight.Type))
		}
	}

	createMesh := func(gltfNode *gltf.Node) *mdl.Mesh {
		mesh := mdl.NewMesh()
		mesh.SetDefinition(meshDefinitionFromIndex[*gltfNode.Mesh])
		if gltfNode.Skin != nil {
			mesh.SetArmature(armatureFromIndex[*gltfNode.Skin])
		}
		return mesh
	}

	// ensure unique node names
	nodeNames := ds.NewSet[string](0)
	for i, gltfNode := range gltfDoc.Nodes {
		if nodeNames.Contains(gltfNode.Name) {
			gltfNode.Name = fmt.Sprintf("%s_%d", gltfNode.Name, i)
		}
		nodeNames.Add(gltfNode.Name)
	}

	// build nodes
	nodeFromIndex := make(map[uint32]*mdl.Node)
	var visitNode func(nodeIndex uint32) *mdl.Node
	visitNode = func(nodeIndex uint32) *mdl.Node {
		gltfNode := gltfDoc.Nodes[nodeIndex]

		node := mdl.NewNode(gltfNode.Name)
		node.SetMetadata(gltfutil.Properties(gltfNode.Extras))

		switch {
		case gltfNodeHasMesh(gltfNode):
			node.SetTarget(createMesh(gltfNode))
		case gltfNodeHasLight(gltfNode):
			node.SetTarget(createLight(gltfNode))
		}

		if gltfNode.MatrixOrDefault() != gltf.DefaultMatrix {
			matrix := dprec.ColumnMajorArrayToMat4(gltfNode.Matrix)
			translation, rotation, scale := matrix.TRS()
			node.SetTranslation(translation)
			node.SetRotation(rotation)
			node.SetScale(scale)
		} else {
			node.SetTranslation(dprec.NewVec3(
				float64(gltfNode.Translation[0]),
				float64(gltfNode.Translation[1]),
				float64(gltfNode.Translation[2]),
			))
			node.SetRotation(dprec.NewQuat(
				float64(gltfNode.Rotation[3]),
				float64(gltfNode.Rotation[0]),
				float64(gltfNode.Rotation[1]),
				float64(gltfNode.Rotation[2]),
			))
			node.SetScale(dprec.NewVec3(
				float64(gltfNode.Scale[0]),
				float64(gltfNode.Scale[1]),
				float64(gltfNode.Scale[2]),
			))
		}

		nodeFromIndex[nodeIndex] = node
		for _, childID := range gltfNode.Children {
			node.AddNode(visitNode(childID))
		}
		return node
	}
	for _, nodeIndex := range gltfutil.RootNodeIndices(gltfDoc) {
		model.AddNode(visitNode(nodeIndex))
	}

	// finalize armatures (now that all nodes are available)
	for i, gltfSkin := range gltfDoc.Skins {
		armature := armatureFromIndex[uint32(i)]
		for j, gltfJoint := range gltfSkin.Joints {
			joint := mdl.NewJoint()
			joint.SetNode(nodeFromIndex[gltfJoint])
			joint.SetInverseBindMatrix(gltfutil.InverseBindMatrix(gltfDoc, gltfSkin, j))
			armature.AddJoint(joint)
		}
	}

	// prepare animations
	for _, gltfAnimation := range gltfDoc.Animations {
		bindingFromNodeIndex := make(map[uint32]*mdl.AnimationBinding)
		animation := &mdl.Animation{}
		animation.SetName(gltfAnimation.Name)
		for _, gltfChannel := range gltfAnimation.Channels {
			nodeRef := gltfChannel.Target.Node
			if nodeRef == nil {
				log.Warn("Channel does not reference a node!")
				continue
			}
			samplerRef := gltfChannel.Sampler
			if samplerRef == nil {
				log.Warn("Channel does not reference a sampler!")
				continue
			}

			binding := bindingFromNodeIndex[*nodeRef]
			if binding == nil {
				binding = &mdl.AnimationBinding{}
				binding.SetNodeName(nodeFromIndex[*nodeRef].Name())
				animation.AddBinding(binding)
				bindingFromNodeIndex[*nodeRef] = binding
			}

			gltfSampler := gltfAnimation.Samplers[*samplerRef]
			if gltfSampler.Interpolation != gltf.InterpolationLinear {
				log.Warn("Unsupported animation interpolation - results may be wrong!")
			}

			timestamps := gltfutil.AnimationKeyframes(gltfDoc, gltfSampler)
			if len(timestamps) > 0 {
				if timestamps[0] < animation.StartTime() {
					animation.SetStartTime(timestamps[0])
				}
				if timestamps[len(timestamps)-1] > animation.EndTime() {
					animation.SetEndTime(timestamps[len(timestamps)-1])
				}
			}

			switch gltfChannel.Target.Path {
			case gltf.TRSTranslation:
				translations := gltfutil.AnimationTranslations(gltfDoc, gltfSampler)
				if len(translations) != len(timestamps) {
					log.Error("Translations do not match number of keyframes")
					continue
				}
				for i := 0; i < len(timestamps); i++ {
					binding.AddTranslationKeyframe(mdl.TranslationKeyframe{
						Timestamp: timestamps[i],
						Value:     translations[i],
					})
				}

			case gltf.TRSRotation:
				rotations := gltfutil.AnimationRotations(gltfDoc, gltfSampler)
				if len(rotations) != len(timestamps) {
					log.Error("Rotations do not match number of keyframes")
					continue
				}
				for i := 0; i < len(timestamps); i++ {
					binding.AddRotationKeyframe(mdl.RotationKeyframe{
						Timestamp: timestamps[i],
						Value:     rotations[i],
					})
				}

			case gltf.TRSScale:
				scales := gltfutil.AnimationScales(gltfDoc, gltfSampler)
				if len(scales) != len(timestamps) {
					log.Error("Scales do not match number of keyframes")
					continue
				}
				for i := 0; i < len(timestamps); i++ {
					binding.AddScaleKeyframe(mdl.ScaleKeyframe{
						Timestamp: timestamps[i],
						Value:     scales[i],
					})
				}

			default:
				log.Warn("Channel has unsupported path: %s", gltfChannel.Target.Path)
			}
		}
		model.AddAnimation(animation)
	}
	return model, nil
}

func openGLTFImage(doc *gltf.Document, img *gltf.Image) (*mdl.Image, error) {
	var content []byte
	if img.BufferView != nil {
		content = gltfutil.BufferViewData(doc, *img.BufferView)
	} else {
		// TODO: Add support for external images
		// in, err := locator.ReadResource(path.Join(path.Dir(a.uri), img.URI))
		// if err != nil {
		// 	return nil, fmt.Errorf("error opening resource: %w", err)
		// }
		// content, err = io.ReadAll(in)
		// if err != nil {
		// 	return nil, fmt.Errorf("error reading resource: %w", err)
		// }
		return nil, fmt.Errorf("external images not supported right now")
	}

	result, err := mdl.ParseImage(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	return result, nil
}

type pbrShaderConfig struct {
	hasColorTexture             bool
	hasMetallicRoughnessTexture bool
	hasNormalTexture            bool
	hasAlphaTesting             bool
}

func createPBRShader(cfg pbrShaderConfig) string {
	var sourceCode string

	var textureLines string
	if cfg.hasColorTexture {
		textureLines += "  colorSampler sampler2D,\n"
	}
	if cfg.hasMetallicRoughnessTexture {
		textureLines += "  metallicRoughnessSampler sampler2D,\n"
	}
	if cfg.hasNormalTexture {
		textureLines += "  normalSampler sampler2D,\n"
	}
	if textureLines != "" {
		sourceCode += "textures {\n" + textureLines + "}\n"
	}

	sourceCode += `
		uniforms {
			color vec4,
			metallic float,
			roughness float,
			normalScale float,
			alphaThreshold float,
		}
	`
	sourceCode += `
		func #fragment() {
	`

	if cfg.hasColorTexture {
		sourceCode += `
			#color = sample(colorSampler, #vertexUV)
		`
	} else {
		sourceCode += `
			#color = color
		`
	}

	sourceCode += `
		#color *= #vertexColor
	`

	if cfg.hasAlphaTesting {
		sourceCode += `
				if #color.a < alphaThreshold {
					discard
				}
		`
	}

	sourceCode += `
			#metallic = metallic
			#roughness = roughness
		}
	`

	return sourceCode
}

func gltfNodeHasMesh(node *gltf.Node) bool {
	return node.Mesh != nil
}

func gltfNodeHasLight(node *gltf.Node) bool {
	if node.Extensions == nil {
		return false
	}

	ext, ok := node.Extensions[lightspunctual.ExtensionName]
	if !ok {
		return false
	}
	_, ok = ext.(lightspunctual.LightIndex)
	return ok
}
