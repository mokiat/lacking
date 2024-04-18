package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	newasset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/render"
)

type ModelDefinition struct {
	nodes                     []nodeDefinition
	armatures                 []armatureDefinition
	animations                []*AnimationDefinition
	textures                  []render.Texture
	materialDefinitions       []*graphics.Material
	meshDefinitions           []*graphics.MeshDefinition
	meshInstances             []meshInstance
	bodyDefinitions           []*physics.BodyDefinition
	bodyInstances             []bodyInstance
	pointLightInstances       []newasset.PointLight
	spotLightInstances        []newasset.SpotLight
	directionalLightInstances []newasset.DirectionalLight
}

func (d *ModelDefinition) FindAnimation(name string) *AnimationDefinition {
	for _, def := range d.animations {
		if def.name == name {
			return def
		}
	}
	return nil
}

type nodeDefinition struct {
	ParentIndex int
	Name        string
	Position    dprec.Vec3
	Rotation    dprec.Quat
	Scale       dprec.Vec3
}

type meshInstance struct {
	NodeIndex       int
	ArmatureIndex   int
	DefinitionIndex int
}

type bodyInstance struct {
	NodeIndex       int
	DefinitionIndex int
}

type armatureDefinition struct {
	Joints []armatureJoint
}

func (d armatureDefinition) InverseBindMatrices() []sprec.Mat4 {
	result := make([]sprec.Mat4, len(d.Joints))
	for i, joint := range d.Joints {
		result[i] = joint.InverseBindMatrix
	}
	return result
}

type armatureJoint struct {
	NodeIndex         int
	InverseBindMatrix sprec.Mat4
}

func (r *ResourceSet) loadModel(resource asset.Resource) (*ModelDefinition, error) {
	modelAsset := new(asset.Model)
	ioTask := func() error {
		return resource.ReadContent(modelAsset)
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	return r.allocateModel(modelAsset)
}

func (r *ResourceSet) allocateModel(modelAsset *asset.Model) (*ModelDefinition, error) {
	nodes := make([]nodeDefinition, len(modelAsset.Nodes))
	for i, nodeAsset := range modelAsset.Nodes {
		nodes[i] = nodeDefinition{
			ParentIndex: int(nodeAsset.ParentIndex),
			Name:        nodeAsset.Name,
			Position:    nodeAsset.Translation,
			Rotation:    nodeAsset.Rotation,
			Scale:       nodeAsset.Scale,
		}
	}

	armatures := make([]armatureDefinition, len(modelAsset.Armatures))
	for i, armatureAsset := range modelAsset.Armatures {
		joints := make([]armatureJoint, len(armatureAsset.Joints))
		for j, jointAsset := range armatureAsset.Joints {
			joints[j] = armatureJoint{
				NodeIndex:         int(jointAsset.NodeIndex),
				InverseBindMatrix: jointAsset.InverseBindMatrix,
			}
		}
		armatures[i] = armatureDefinition{
			Joints: joints,
		}
	}

	animations := make([]*AnimationDefinition, len(modelAsset.Animations))
	for i, animationAsset := range modelAsset.Animations {
		bindings := make([]AnimationBindingDefinitionInfo, len(animationAsset.Bindings))
		for j, assetBinding := range animationAsset.Bindings {
			translationKeyframes := make([]Keyframe[dprec.Vec3], len(assetBinding.TranslationKeyframes))
			for k, keyframe := range assetBinding.TranslationKeyframes {
				translationKeyframes[k] = Keyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Translation,
				}
			}
			rotationKeyframes := make([]Keyframe[dprec.Quat], len(assetBinding.RotationKeyframes))
			for k, keyframe := range assetBinding.RotationKeyframes {
				rotationKeyframes[k] = Keyframe[dprec.Quat]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Rotation,
				}
			}
			scaleKeyframes := make([]Keyframe[dprec.Vec3], len(assetBinding.ScaleKeyframes))
			for k, keyframe := range assetBinding.ScaleKeyframes {
				scaleKeyframes[k] = Keyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Scale,
				}
			}
			bindings[j] = AnimationBindingDefinitionInfo{
				NodeIndex:            int(assetBinding.NodeIndex),
				NodeName:             assetBinding.NodeName,
				TranslationKeyframes: translationKeyframes,
				RotationKeyframes:    rotationKeyframes,
				ScaleKeyframes:       scaleKeyframes,
			}
		}
		r.gfxWorker.ScheduleVoid(func() {
			animations[i] = r.engine.CreateAnimationDefinition(AnimationDefinitionInfo{
				Name:      animationAsset.Name,
				StartTime: animationAsset.StartTime,
				EndTime:   animationAsset.EndTime,
				Bindings:  bindings,
			})
		}).Wait()
	}

	bodyMaterials := make([]*physics.Material, len(modelAsset.BodyMaterials))
	for i, materialAsset := range modelAsset.BodyMaterials {
		physicsEngine := r.engine.Physics()
		bodyMaterials[i] = physicsEngine.CreateMaterial(physics.MaterialInfo{
			FrictionCoefficient:    materialAsset.FrictionCoefficient,
			RestitutionCoefficient: materialAsset.RestitutionCoefficient,
		})
	}

	bodyDefinitions := make([]*physics.BodyDefinition, len(modelAsset.BodyDefinitions))
	for i, definitionAsset := range modelAsset.BodyDefinitions {
		material := bodyMaterials[definitionAsset.MaterialIndex]
		physicsEngine := r.engine.Physics()
		r.gfxWorker.ScheduleVoid(func() {
			bodyDefinitions[i] = physicsEngine.CreateBodyDefinition(physics.BodyDefinitionInfo{
				Mass:                   definitionAsset.Mass,
				MomentOfInertia:        definitionAsset.MomentOfInertia,
				FrictionCoefficient:    material.FrictionCoefficient(),
				RestitutionCoefficient: material.RestitutionCoefficient(),
				DragFactor:             definitionAsset.DragFactor,
				AngularDragFactor:      definitionAsset.AngularDragFactor,
				AerodynamicShapes:      nil, // TODO
				CollisionSpheres:       r.constructCollisionSpheres(definitionAsset),
				CollisionBoxes:         r.constructCollisionBoxes(definitionAsset),
				CollisionMeshes:        r.constructCollisionMeshes(definitionAsset),
			})
		}).Wait()
	}

	bodyInstances := make([]bodyInstance, len(modelAsset.BodyInstances))
	for i, instanceAsset := range modelAsset.BodyInstances {
		bodyInstances[i] = bodyInstance{
			NodeIndex:       int(instanceAsset.NodeIndex),
			DefinitionIndex: int(instanceAsset.BodyDefinitionIndex),
		}
	}

	textures := make([]render.Texture, len(modelAsset.Textures))
	for i, textureAsset := range modelAsset.Textures {
		textures[i] = r.allocateTexture(textureAsset)
	}

	gfxEngine := r.engine.Graphics()

	geometryShaders := make([]*graphics.GeometryShader, len(modelAsset.GeometryShaders))
	for i, shaderAsset := range modelAsset.GeometryShaders {
		var shader *graphics.GeometryShader
		r.gfxWorker.ScheduleVoid(func() {
			shader = gfxEngine.CreateGeometryShader(graphics.ShaderInfo{
				SourceCode: shaderAsset.SourceCode,
			})
		}).Wait()
		geometryShaders[i] = shader
	}

	shadowShaders := make([]*graphics.ShadowShader, len(modelAsset.ShadowShaders))
	for i, shaderAsset := range modelAsset.ShadowShaders {
		var shader *graphics.ShadowShader
		r.gfxWorker.ScheduleVoid(func() {
			shader = gfxEngine.CreateShadowShader(graphics.ShaderInfo{
				SourceCode: shaderAsset.SourceCode,
			})
		}).Wait()
		shadowShaders[i] = shader
	}

	forwardShaders := make([]*graphics.ForwardShader, len(modelAsset.ForwardShaders))
	for i, shaderAsset := range modelAsset.ForwardShaders {
		var shader *graphics.ForwardShader
		r.gfxWorker.ScheduleVoid(func() {
			shader = gfxEngine.CreateForwardShader(graphics.ShaderInfo{
				SourceCode: shaderAsset.SourceCode,
			})
		}).Wait()
		forwardShaders[i] = shader
	}

	materialDefinitions := make([]*graphics.Material, len(modelAsset.Materials))
	for i, materialAsset := range modelAsset.Materials {
		materialInfo := graphics.MaterialInfo{
			Name: materialAsset.Name,
			GeometryPasses: gog.Map(materialAsset.GeometryPasses, func(pass newasset.GeometryPass) graphics.GeometryRenderPassInfo {
				return graphics.GeometryRenderPassInfo{
					Layer:           pass.Layer,
					Culling:         opt.V(resolveCullMode(pass.Culling)),
					FrontFace:       opt.V(resolveFaceOrientation(pass.FrontFace)),
					DepthTest:       opt.V(pass.DepthTest),
					DepthWrite:      opt.V(pass.DepthWrite),
					DepthComparison: opt.V(resolveComparison(pass.DepthComparison)),
					Shader:          geometryShaders[pass.ShaderIndex],
				}
			}),
			ShadowPasses: gog.Map(materialAsset.ShadowPasses, func(pass newasset.ShadowPass) graphics.ShadowRenderPassInfo {
				return graphics.ShadowRenderPassInfo{
					Culling:   opt.V(resolveCullMode(pass.Culling)),
					FrontFace: opt.V(resolveFaceOrientation(pass.FrontFace)),
					Shader:    shadowShaders[pass.ShaderIndex],
				}
			}),
			ForwardPasses: gog.Map(materialAsset.ForwardPasses, func(pass newasset.ForwardPass) graphics.ForwardRenderPassInfo {
				return graphics.ForwardRenderPassInfo{
					Layer:           pass.Layer,
					Culling:         opt.V(resolveCullMode(pass.Culling)),
					FrontFace:       opt.V(resolveFaceOrientation(pass.FrontFace)),
					DepthTest:       opt.V(pass.DepthTest),
					DepthWrite:      opt.V(pass.DepthWrite),
					DepthComparison: opt.V(resolveComparison(pass.DepthComparison)),
					Blending:        opt.V(pass.Blending),
					Shader:          forwardShaders[pass.ShaderIndex],
				}
			}),
		}

		var material *graphics.Material
		r.gfxWorker.ScheduleVoid(func() {
			material = gfxEngine.CreateMaterial(materialInfo)
			for _, binding := range materialAsset.Textures {
				texture := textures[binding.TextureIndex]
				material.SetTexture(binding.BindingName, texture)
			}
			for _, binding := range materialAsset.Textures {
				sampler := r.renderAPI.CreateSampler(render.SamplerInfo{
					Wrapping:   resolveWrapMode(binding.Wrapping),
					Filtering:  resolveFiltering(binding.Filtering),
					Mipmapping: binding.Mipmapping,
				})
				material.SetSampler(binding.BindingName, sampler)
			}
			for _, binding := range materialAsset.Properties {
				material.SetProperty(binding.BindingName, binding.Data)
			}
		}).Wait()
		materialDefinitions[i] = material
	}

	meshGeometries := make([]*graphics.MeshGeometry, len(modelAsset.MeshDefinitions))
	for i, assetGeometry := range modelAsset.Geometries {
		meshFragmentsInfo := make([]graphics.MeshGeometryFragmentInfo, len(assetGeometry.Fragments))
		for j, assetFragment := range assetGeometry.Fragments {
			meshFragmentsInfo[j] = graphics.MeshGeometryFragmentInfo{
				Name:            assetFragment.Name,
				Topology:        resolveTopology(assetFragment.Topology),
				IndexByteOffset: assetFragment.IndexByteOffset,
				IndexCount:      assetFragment.IndexCount,
			}
		}

		meshGeometryInfo := graphics.MeshGeometryInfo{
			VertexBuffers: gog.Map(assetGeometry.VertexBuffers, func(buffer newasset.VertexBuffer) graphics.MeshGeometryVertexBuffer {
				return graphics.MeshGeometryVertexBuffer{
					ByteStride: buffer.Stride,
					Data:       buffer.Data,
				}
			}),
			VertexFormat: resolveVertexFormat(assetGeometry.VertexLayout),
			IndexBuffer: graphics.MeshGeometryIndexBuffer{
				Data:   assetGeometry.IndexBuffer.Data,
				Format: resolveIndexFormat(assetGeometry.IndexBuffer.IndexLayout),
			},
			Fragments:            meshFragmentsInfo,
			BoundingSphereRadius: assetGeometry.BoundingSphereRadius,
		}

		gfxEngine := r.engine.Graphics()
		r.gfxWorker.ScheduleVoid(func() {
			meshGeometries[i] = gfxEngine.CreateMeshGeometry(meshGeometryInfo)
		}).Wait()
	}

	meshDefinitions := make([]*graphics.MeshDefinition, len(modelAsset.MeshDefinitions))
	for i, assetDefinition := range modelAsset.MeshDefinitions {
		gfxEngine := r.engine.Graphics()
		r.gfxWorker.ScheduleVoid(func() {
			var materials []*graphics.Material
			for _, assetBinding := range assetDefinition.MaterialBindings {
				materials = append(materials, materialDefinitions[assetBinding.MaterialIndex])
			}
			meshDefinitions[i] = gfxEngine.CreateMeshDefinition(graphics.MeshDefinitionInfo{
				Geometry:  meshGeometries[i],
				Materials: materials,
			})
		}).Wait()
	}

	meshInstances := make([]meshInstance, len(modelAsset.MeshInstances))
	for i, instanceAsset := range modelAsset.MeshInstances {
		meshInstances[i] = meshInstance{
			NodeIndex:       int(instanceAsset.NodeIndex),
			ArmatureIndex:   int(instanceAsset.ArmatureIndex),
			DefinitionIndex: int(instanceAsset.MeshDefinitionIndex),
		}
	}

	return &ModelDefinition{
		nodes:                     nodes,
		armatures:                 armatures,
		animations:                animations,
		textures:                  textures,
		materialDefinitions:       materialDefinitions,
		meshDefinitions:           meshDefinitions,
		meshInstances:             meshInstances,
		bodyDefinitions:           bodyDefinitions,
		bodyInstances:             bodyInstances,
		pointLightInstances:       modelAsset.PointLights,
		spotLightInstances:        modelAsset.SpotLights,
		directionalLightInstances: modelAsset.DirectionalLights,
	}, nil
}

func (r *ResourceSet) releaseModel(model *ModelDefinition) {
	for _, texture := range model.textures {
		texture.Release()
	}
	model.textures = nil
}

func (r *ResourceSet) constructCollisionSpheres(bodyDef newasset.BodyDefinition) []collision.Sphere {
	result := make([]collision.Sphere, len(bodyDef.CollisionSpheres))
	for i, collisionSphereAsset := range bodyDef.CollisionSpheres {
		result[i] = collision.NewSphere(
			collisionSphereAsset.Translation,
			collisionSphereAsset.Radius,
		)
	}
	return result
}

func (r *ResourceSet) constructCollisionBoxes(bodyDef newasset.BodyDefinition) []collision.Box {
	result := make([]collision.Box, len(bodyDef.CollisionBoxes))
	for i, collisionBoxAsset := range bodyDef.CollisionBoxes {
		result[i] = collision.NewBox(
			collisionBoxAsset.Translation,
			collisionBoxAsset.Rotation,
			dprec.NewVec3(collisionBoxAsset.Width, collisionBoxAsset.Height, collisionBoxAsset.Lenght),
		)
	}
	return result
}

func (r *ResourceSet) constructCollisionMeshes(bodyDef newasset.BodyDefinition) []collision.Mesh {
	result := make([]collision.Mesh, len(bodyDef.CollisionMeshes))
	for i, collisionMeshAsset := range bodyDef.CollisionMeshes {
		transform := collision.TRTransform(collisionMeshAsset.Translation, collisionMeshAsset.Rotation)
		triangles := make([]collision.Triangle, len(collisionMeshAsset.Triangles))
		for j, triangleAsset := range collisionMeshAsset.Triangles {
			template := collision.NewTriangle(
				triangleAsset.A,
				triangleAsset.B,
				triangleAsset.C,
			)
			triangles[j].Replace(template, transform)
		}
		result[i] = collision.NewMesh(triangles)
	}
	return result
}

func resolveVertexFormat(layout newasset.VertexLayout) graphics.MeshGeometryVertexFormat {
	var result graphics.MeshGeometryVertexFormat
	if attrib := layout.Coord; attrib.BufferIndex != newasset.UnspecifiedBufferIndex {
		result.Coord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Normal; attrib.BufferIndex != newasset.UnspecifiedBufferIndex {
		result.Normal = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Tangent; attrib.BufferIndex != newasset.UnspecifiedBufferIndex {
		result.Tangent = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.TexCoord; attrib.BufferIndex != newasset.UnspecifiedBufferIndex {
		result.TexCoord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Color; attrib.BufferIndex != newasset.UnspecifiedBufferIndex {
		result.Color = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Weights; attrib.BufferIndex != newasset.UnspecifiedBufferIndex {
		result.Weights = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Joints; attrib.BufferIndex != newasset.UnspecifiedBufferIndex {
		result.Joints = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      resolveVertexAttributeFormat(attrib.Format),
		})
	}
	return result
}

func resolveVertexAttributeFormat(format newasset.VertexAttributeFormat) render.VertexAttributeFormat {
	switch format {
	case newasset.VertexAttributeFormatRGBA32F:
		return render.VertexAttributeFormatRGBA32F
	case newasset.VertexAttributeFormatRGB32F:
		return render.VertexAttributeFormatRGB32F
	case newasset.VertexAttributeFormatRG32F:
		return render.VertexAttributeFormatRG32F
	case newasset.VertexAttributeFormatR32F:
		return render.VertexAttributeFormatR32F

	case newasset.VertexAttributeFormatRGBA16F:
		return render.VertexAttributeFormatRGBA16F
	case newasset.VertexAttributeFormatRGB16F:
		return render.VertexAttributeFormatRGB16F
	case newasset.VertexAttributeFormatRG16F:
		return render.VertexAttributeFormatRG16F
	case newasset.VertexAttributeFormatR16F:
		return render.VertexAttributeFormatR16F

	case newasset.VertexAttributeFormatRGBA16S:
		return render.VertexAttributeFormatRGBA16S
	case newasset.VertexAttributeFormatRGB16S:
		return render.VertexAttributeFormatRGB16S
	case newasset.VertexAttributeFormatRG16S:
		return render.VertexAttributeFormatRG16S
	case newasset.VertexAttributeFormatR16S:
		return render.VertexAttributeFormatR16S

	case newasset.VertexAttributeFormatRGBA16SN:
		return render.VertexAttributeFormatRGBA16SN
	case newasset.VertexAttributeFormatRGB16SN:
		return render.VertexAttributeFormatRGB16SN
	case newasset.VertexAttributeFormatRG16SN:
		return render.VertexAttributeFormatRG16SN
	case newasset.VertexAttributeFormatR16SN:
		return render.VertexAttributeFormatR16SN

	case newasset.VertexAttributeFormatRGBA16U:
		return render.VertexAttributeFormatRGBA16U
	case newasset.VertexAttributeFormatRGB16U:
		return render.VertexAttributeFormatRGB16U
	case newasset.VertexAttributeFormatRG16U:
		return render.VertexAttributeFormatRG16U
	case newasset.VertexAttributeFormatR16U:
		return render.VertexAttributeFormatR16U

	case newasset.VertexAttributeFormatRGBA16UN:
		return render.VertexAttributeFormatRGBA16UN
	case newasset.VertexAttributeFormatRGB16UN:
		return render.VertexAttributeFormatRGB16UN
	case newasset.VertexAttributeFormatRG16UN:
		return render.VertexAttributeFormatRG16UN
	case newasset.VertexAttributeFormatR16UN:
		return render.VertexAttributeFormatR16UN

	case newasset.VertexAttributeFormatRGBA8S:
		return render.VertexAttributeFormatRGBA8S
	case newasset.VertexAttributeFormatRGB8S:
		return render.VertexAttributeFormatRGB8S
	case newasset.VertexAttributeFormatRG8S:
		return render.VertexAttributeFormatRG8S
	case newasset.VertexAttributeFormatR8S:
		return render.VertexAttributeFormatR8S

	case newasset.VertexAttributeFormatRGBA8SN:
		return render.VertexAttributeFormatRGBA8SN
	case newasset.VertexAttributeFormatRGB8SN:
		return render.VertexAttributeFormatRGB8SN
	case newasset.VertexAttributeFormatRG8SN:
		return render.VertexAttributeFormatRG8SN
	case newasset.VertexAttributeFormatR8SN:
		return render.VertexAttributeFormatR8SN

	case newasset.VertexAttributeFormatRGBA8U:
		return render.VertexAttributeFormatRGBA8U
	case newasset.VertexAttributeFormatRGB8U:
		return render.VertexAttributeFormatRGB8U
	case newasset.VertexAttributeFormatRG8U:
		return render.VertexAttributeFormatRG8U
	case newasset.VertexAttributeFormatR8U:
		return render.VertexAttributeFormatR8U

	case newasset.VertexAttributeFormatRGBA8UN:
		return render.VertexAttributeFormatRGBA8UN
	case newasset.VertexAttributeFormatRGB8UN:
		return render.VertexAttributeFormatRGB8UN
	case newasset.VertexAttributeFormatRG8UN:
		return render.VertexAttributeFormatRG8UN
	case newasset.VertexAttributeFormatR8UN:
		return render.VertexAttributeFormatR8UN

	case newasset.VertexAttributeFormatRGBA8IU:
		return render.VertexAttributeFormatRGBA8IU
	case newasset.VertexAttributeFormatRGB8IU:
		return render.VertexAttributeFormatRGB8IU
	case newasset.VertexAttributeFormatRG8IU:
		return render.VertexAttributeFormatRG8IU
	case newasset.VertexAttributeFormatR8IU:
		return render.VertexAttributeFormatR8IU

	default:
		panic(fmt.Errorf("unsupported vertex attribute format: %d", format))
	}
}

func resolveIndexFormat(layout newasset.IndexLayout) render.IndexFormat {
	switch layout {
	case newasset.IndexLayoutUint16:
		return render.IndexFormatUnsignedU16
	case newasset.IndexLayoutUint32:
		return render.IndexFormatUnsignedU32
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}

func resolveTopology(primitive newasset.Topology) render.Topology {
	switch primitive {
	case newasset.TopologyPoints:
		return render.TopologyPoints
	case newasset.TopologyLineList:
		return render.TopologyLineList
	case newasset.TopologyLineStrip:
		return render.TopologyLineStrip
	case newasset.TopologyTriangleList:
		return render.TopologyTriangleList
	case newasset.TopologyTriangleStrip:
		return render.TopologyTriangleStrip
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}

// ModelInfo contains the information necessary to place a Model
// instance into a Scene.
type ModelInfo struct {
	// Name specifies the name of this instance. This should not be
	// confused with the name of the definition.
	Name string

	// Definition specifies the template from which this instance will
	// be created.
	Definition *ModelDefinition

	// Position is used to specify a location for the model instance.
	Position dprec.Vec3

	// Rotation is used to specify a rotation for the model instance.
	Rotation dprec.Quat

	// Scale is used to specify a scale for the model instance.
	Scale dprec.Vec3

	// IsDynamic determines whether the model can be repositioned once
	// placed in the Scene.
	// (i.e. whether it should be added to the scene hierarchy)
	IsDynamic bool

	// PrepareAnimations indicates whether animation definitions should be
	// instantiated for this model.
	PrepareAnimations bool
}

type Model struct {
	definition *ModelDefinition
	root       *hierarchy.Node

	nodes         []*hierarchy.Node
	armatures     []*graphics.Armature
	animations    []*Animation
	bodyInstances []physics.Body
}

func (m *Model) Root() *hierarchy.Node {
	return m.root
}

func (m *Model) FindNode(name string) *hierarchy.Node {
	return m.root.FindNode(name)
}

func (m *Model) BodyInstances() []physics.Body {
	return m.bodyInstances
}

func (m *Model) Animations() []*Animation {
	return m.animations
}

func (m *Model) FindAnimation(name string) *Animation {
	for _, animation := range m.animations {
		if animation.name == name {
			return animation
		}
	}
	return nil
}
