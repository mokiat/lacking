package pack

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/util/gltfutil"
	"github.com/mokiat/lacking/util/resource"
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
	locator resource.ReadLocator
	uri     string
	model   *Model
}

func (a *OpenGLTFResourceAction) Describe() string {
	return fmt.Sprintf("open_gltf_resource(%q)", a.uri)
}

func (a *OpenGLTFResourceAction) Model() *Model {
	if a.model == nil {
		panic("reading data from unprocessed action")
	}
	return a.model
}

func (a *OpenGLTFResourceAction) Run() error {
	in, err := a.locator.ReadResource(a.uri)
	if err != nil {
		return fmt.Errorf("failed to open model resource %q: %w", a.uri, err)
	}
	defer in.Close()

	gltfDoc := new(gltf.Document)
	if err := gltf.NewDecoder(in).Decode(gltfDoc); err != nil {
		return fmt.Errorf("failed to parse gltf model %q: %w", a.uri, err)
	}

	a.model = &Model{}

	imagesFromIndex := make(map[uint32]*Image)
	for i, gltfImage := range gltfDoc.Images {
		img, err := a.openImage(gltfDoc, gltfImage, a.locator)
		if err != nil {
			return fmt.Errorf("error loading image: %w", err)
		}
		a.model.Textures = append(a.model.Textures, img)
		imagesFromIndex[uint32(i)] = img
	}

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
			ColorTexture:             nil,
			Metallic:                 1.0,
			Roughness:                1.0,
			MetallicRoughnessTexture: nil,
			NormalScale:              1.0,
			NormalTexture:            nil,
		}
		if gltfPBR := gltfMaterial.PBRMetallicRoughness; gltfPBR != nil {
			material.Color = gltfutil.BaseColor(gltfPBR)
			material.Metallic = gltfPBR.MetallicFactorOrDefault()
			material.Roughness = gltfPBR.RoughnessFactorOrDefault()
			if texIndex := gltfutil.ColorTextureIndex(gltfDoc, gltfPBR); texIndex != nil {
				material.ColorTexture = &TextureRef{
					TextureIndex: int(*texIndex),
				}
			}
			if texIndex := gltfutil.MetallicRoughnessTextureIndex(gltfDoc, gltfPBR); texIndex != nil {
				material.MetallicRoughnessTexture = &TextureRef{
					TextureIndex: int(*texIndex),
				}
			}
		}
		if texIndex, texScale := gltfutil.NormalTextureIndexScale(gltfDoc, gltfMaterial); texIndex != nil {
			material.NormalTexture = &TextureRef{
				TextureIndex: int(*texIndex),
			}
			material.NormalScale = texScale
		}
		a.model.Materials = append(a.model.Materials, material)
		materialFromIndex[uint32(i)] = material
	}

	// build mesh definitions
	meshDefinitionFromIndex := make(map[uint32]*MeshDefinition)
	for i, gltfMesh := range gltfDoc.Meshes {
		mesh := &MeshDefinition{
			Name:         gltfMesh.Name,
			Fragments:    make([]MeshFragment, len(gltfMesh.Primitives)),
			HasCollision: !gltfutil.IsMeshCollisionDisabled(gltfMesh),
		}
		meshDefinitionFromIndex[uint32(i)] = mesh
		a.model.MeshDefinitions = append(a.model.MeshDefinitions, mesh)
		indexFromVertex := make(map[Vertex]int)

		for j, gltfPrimitive := range gltfMesh.Primitives {
			indexOffset := len(mesh.Indices) // this needs to happen first
			gltfIndices, err := gltfutil.Indices(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading indices: %w", err)
			}
			gltfCoords, err := gltfutil.Coords(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading coords: %w", err)
			}
			gltfNormals, err := gltfutil.Normals(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading normals: %w", err)
			}
			gltfTangents, err := gltfutil.Tangents(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading tangents: %w", err)
			}
			gltfTexCoords, err := gltfutil.TexCoord0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading tex coords: %w", err)
			}
			gltfColors, err := gltfutil.Color0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading colors: %w", err)
			}
			gltfWeights, err := gltfutil.Weight0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading weights: %w", err)
			}
			gltfJoints, err := gltfutil.Joint0s(gltfDoc, gltfPrimitive)
			if err != nil {
				return fmt.Errorf("error reading joints: %w", err)
			}

			if gltfCoords != nil {
				mesh.VertexLayout.HasCoords = true
			}
			if gltfNormals != nil {
				mesh.VertexLayout.HasNormals = true
			}
			if gltfTangents != nil {
				mesh.VertexLayout.HasTangents = true
			}
			if gltfTexCoords != nil {
				mesh.VertexLayout.HasTexCoords = true
			}
			if gltfColors != nil {
				mesh.VertexLayout.HasColors = true
			}
			if gltfWeights != nil {
				mesh.VertexLayout.HasWeights = true
			}
			if gltfJoints != nil {
				mesh.VertexLayout.HasJoints = true
			}

			for _, gltfIndex := range gltfIndices {
				var vertex Vertex
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
					mesh.Indices = append(mesh.Indices, index)
				} else {
					index = len(mesh.Vertices)
					mesh.Vertices = append(mesh.Vertices, vertex)
					mesh.Indices = append(mesh.Indices, index)
					indexFromVertex[vertex] = index
				}
			}

			var fragment MeshFragment
			fragment.IndexOffset = indexOffset
			fragment.IndexCount = len(gltfIndices)
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

	// prepare armatures
	armatureDefinitionFromIndex := make(map[uint32]*Armature)
	for i, gltfSkin := range gltfDoc.Skins {
		armature := &Armature{
			Joints: make([]Joint, len(gltfSkin.Joints)),
		}
		armatureDefinitionFromIndex[uint32(i)] = armature
		a.model.Armatures = append(a.model.Armatures, armature)
	}

	// build nodes
	nodeFromIndex := make(map[uint32]*Node)
	var visitNode func(nodeIndex uint32) *Node
	visitNode = func(nodeIndex uint32) *Node {
		gltfNode := gltfDoc.Nodes[nodeIndex]
		node := &Node{
			Name:        gltfNode.Name,
			Translation: dprec.ZeroVec3(),
			Rotation:    dprec.IdentityQuat(),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		nodeFromIndex[nodeIndex] = node

		if gltfNode.Matrix != gltf.DefaultMatrix {
			matrix := stod.Mat4(sprec.ColumnMajorArrayToMat4(gltfNode.Matrix))
			translation, rotation, scale := matrix.TRS()
			node.Translation = translation
			node.Rotation = rotation
			node.Scale = scale
		} else {
			node.Translation = dprec.NewVec3(
				float64(gltfNode.Translation[0]),
				float64(gltfNode.Translation[1]),
				float64(gltfNode.Translation[2]),
			)
			node.Rotation = dprec.NewQuat(
				float64(gltfNode.Rotation[3]),
				float64(gltfNode.Rotation[0]),
				float64(gltfNode.Rotation[1]),
				float64(gltfNode.Rotation[2]),
			)
			node.Scale = dprec.NewVec3(
				float64(gltfNode.Scale[0]),
				float64(gltfNode.Scale[1]),
				float64(gltfNode.Scale[2]),
			)
		}

		if gltfNode.Mesh != nil {
			meshDefinition := meshDefinitionFromIndex[*gltfNode.Mesh]
			meshInstance := &MeshInstance{
				Name:         gltfNode.Name,
				Node:         node,
				Definition:   meshDefinition,
				HasCollision: !gltfutil.IsCollisionDisabled(gltfNode) && meshDefinition.HasCollision,
			}
			if gltfNode.Skin != nil {
				meshInstance.Armature = armatureDefinitionFromIndex[*gltfNode.Skin]
			}
			if gltfutil.IsCollisionDisabled(gltfNode) {
				meshInstance.HasCollision = false
			}
			a.model.MeshInstances = append(a.model.MeshInstances, meshInstance)
		}
		for _, childID := range gltfNode.Children {
			node.Children = append(node.Children, visitNode(childID))
		}
		return node
	}
	for _, nodeIndex := range gltfutil.RootNodeIndices(gltfDoc) {
		a.model.RootNodes = append(a.model.RootNodes, visitNode(nodeIndex))
	}

	// finalize armatures (now that all nodes are available)
	for i, gltfSkin := range gltfDoc.Skins {
		armature := a.model.Armatures[i]
		for j, joint := range gltfSkin.Joints {
			armature.Joints[j].Node = nodeFromIndex[joint]
			armature.Joints[j].InverseBindMatrix = gltfutil.InverseBindMatrix(gltfDoc, gltfSkin, j)
		}
	}

	// prepare animations
	for _, gltfAnimation := range gltfDoc.Animations {
		bindingFromNodeIndex := make(map[uint32]*AnimationBinding)
		animation := &Animation{
			Name: gltfAnimation.Name,
		}
		for _, gltfChannel := range gltfAnimation.Channels {
			nodeRef := gltfChannel.Target.Node
			if nodeRef == nil {
				log.Warn("Channel does not reference a node")
				continue
			}
			samplerRef := gltfChannel.Sampler
			if samplerRef == nil {
				log.Warn("Channel does not reference a sampler")
				continue
			}
			binding := bindingFromNodeIndex[*nodeRef]
			if binding == nil {
				binding = &AnimationBinding{
					Node: nodeFromIndex[*nodeRef],
				}
				animation.Bindings = append(animation.Bindings, binding)
				bindingFromNodeIndex[*nodeRef] = binding
			}

			gltfSampler := gltfAnimation.Samplers[*samplerRef]
			if gltfSampler.Interpolation != gltf.InterpolationLinear {
				log.Warn("Unsupported animation interpolation - results may be wrong")
			}

			timestamps := gltfutil.AnimationKeyframes(gltfDoc, gltfSampler)
			if len(timestamps) > 0 {
				if timestamps[0] < animation.StartTime {
					animation.StartTime = timestamps[0]
				}
				if timestamps[len(timestamps)-1] > animation.EndTime {
					animation.EndTime = timestamps[len(timestamps)-1]
				}
			}

			switch gltfChannel.Target.Path {
			case gltf.TRSTranslation:
				translations := gltfutil.AnimationTranslations(gltfDoc, gltfSampler)
				if len(translations) != len(timestamps) {
					log.Error("Translations do not match number of keyframes")
					continue
				}
				binding.TranslationKeyframes = make([]TranslationKeyframe, len(timestamps))
				for i := 0; i < len(timestamps); i++ {
					binding.TranslationKeyframes[i] = TranslationKeyframe{
						Timestamp:   timestamps[i],
						Translation: translations[i],
					}
				}

			case gltf.TRSRotation:
				rotations := gltfutil.AnimationRotations(gltfDoc, gltfSampler)
				if len(rotations) != len(timestamps) {
					log.Error("Rotations do not match number of keyframes")
					continue
				}
				binding.RotationKeyframes = make([]RotationKeyframe, len(timestamps))
				for i := 0; i < len(timestamps); i++ {
					binding.RotationKeyframes[i] = RotationKeyframe{
						Timestamp: timestamps[i],
						Rotation:  rotations[i],
					}
				}

			case gltf.TRSScale:
				scales := gltfutil.AnimationScales(gltfDoc, gltfSampler)
				if len(scales) != len(timestamps) {
					log.Error("Scales do not match number of keyframes")
					continue
				}
				binding.ScaleKeyframes = make([]ScaleKeyframe, len(timestamps))
				for i := 0; i < len(timestamps); i++ {
					binding.ScaleKeyframes[i] = ScaleKeyframe{
						Timestamp: timestamps[i],
						Scale:     scales[i],
					}
				}

			default:
				log.Warn("Channel has unsupported path")
			}
		}
		a.model.Animations = append(a.model.Animations, animation)
	}
	return nil
}

func (a *OpenGLTFResourceAction) openImage(doc *gltf.Document, img *gltf.Image, locator resource.ReadLocator) (*Image, error) {
	var content []byte
	if img.BufferView != nil {
		content = gltfutil.BufferViewData(doc, *img.BufferView)
	} else {
		in, err := locator.ReadResource(path.Join(path.Dir(a.uri), img.URI))
		if err != nil {
			return nil, fmt.Errorf("error opening resource: %w", err)
		}
		content, err = io.ReadAll(in)
		if err != nil {
			return nil, fmt.Errorf("error reading resource: %w", err)
		}
	}
	return ParseImageResource(bytes.NewReader(content))
}
