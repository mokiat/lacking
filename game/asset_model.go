package game

import (
	"errors"
	"fmt"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/render"
)

// ModelTemplate represents a template for a model that can be instantiated
// in a Scene.
type ModelTemplate struct {
	Recordings      IdentifiableList[*animation.Clip]
	Shaders         IdentifiableList[*graphics.Shader]
	Textures        IdentifiableList[render.Texture]
	Materials       IdentifiableList[*graphics.Material]
	MeshGeometries  IdentifiableList[*graphics.MeshGeometry]
	MeshDefinitions IdentifiableList[*graphics.MeshDefinition]

	Nodes             IdentifiableList[NodeTemplate]
	Terrains          IdentifiableList[TerrainTemplate]
	Armatures         IdentifiableList[ArmatureTemplate]
	Meshes            IdentifiableList[MeshTemplate]
	AmbientLights     IdentifiableList[AmbientLightTemplate]
	PointLights       IdentifiableList[PointLightTemplate]
	SpotLights        IdentifiableList[SpotLightTemplate]
	DirectionalLights IdentifiableList[DirectionalLightTemplate]
	SkyTemplates      IdentifiableList[SkyTemplate]
}

func (t *ModelTemplate) FindRecording(name string) *animation.Clip {
	for _, recording := range t.Recordings.Iter() {
		if recording.Name() == name {
			return recording
		}
	}
	return nil
}

// LoadModelTemplate resolves a model template from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadModelTemplate(loader *AssetLoader, assetModel dto.Model) (*ModelTemplate, error) {
	restTransforms := restTransformsOf(assetModel.HierarchyChunk.Nodes)
	recordings, err := LoadAnimationRecordings(loader, assetModel.AnimationChunk.Animations, restTransforms)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve animation recordings: %w", err)
	}

	shaders, err := LoadShaders(loader, assetModel.ShadingChunk.Shaders)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve shaders: %w", err)
	}

	textures, err := LoadTextures(loader, assetModel.ShadingChunk.Textures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve textures: %w", err)
	}

	materials, err := LoadMaterials(loader, assetModel.ShadingChunk.Materials, shaders, textures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve materials: %w", err)
	}

	meshGeometries, err := LoadMeshGeometries(loader, assetModel.MeshChunk.Geometries)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh geometries: %w", err)
	}

	meshDefinitions, err := LoadMeshDefinitions(loader, assetModel.MeshChunk.MeshDefinitions, meshGeometries, materials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh definitions: %w", err)
	}

	nodes, err := LoadNodeTemplates(loader, assetModel.HierarchyChunk.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node templates: %w", err)
	}

	terrains, err := LoadPhysicsTerrainTemplates(loader, assetModel.PhysicsChunk.Terrains)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve physics terrain templates: %w", err)
	}

	armatures, err := LoadArmatureTemplates(loader, assetModel.MeshChunk.Armatures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve armature templates: %w", err)
	}

	meshes, err := LoadMeshTemplates(loader, assetModel.MeshChunk.Meshes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh templates: %w", err)
	}

	ambientLights, err := LoadAmbientLightTemplates(loader, assetModel.LightingChunk.AmbientLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ambient light templates: %w", err)
	}

	pointLights, err := LoadPointLightTemplates(loader, assetModel.LightingChunk.PointLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve point light templates: %w", err)
	}

	spotLights, err := LoadSpotLightTemplates(loader, assetModel.LightingChunk.SpotLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve spot light templates: %w", err)
	}

	directionalLights, err := LoadDirectionalLightTemplates(loader, assetModel.LightingChunk.DirectionalLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directional light templates: %w", err)
	}

	// TODO: Convert cameras

	skyTemplates, err := LoadSkyTemplates(loader, assetModel.BackgroundChunk.Skies, materials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve sky templates: %w", err)
	}

	return &ModelTemplate{
		Recordings:      recordings,
		Shaders:         shaders,
		Textures:        textures,
		Materials:       materials,
		MeshGeometries:  meshGeometries,
		MeshDefinitions: meshDefinitions,

		Nodes:             nodes,
		Terrains:          terrains,
		Armatures:         armatures,
		Meshes:            meshes,
		AmbientLights:     ambientLights,
		PointLights:       pointLights,
		SpotLights:        spotLights,
		DirectionalLights: directionalLights,
		SkyTemplates:      skyTemplates,
	}, nil
}

// UnloadModelTemplate unloads the given model template from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadModelTemplate(loader *AssetLoader, template *ModelTemplate) error {
	return errors.Join(
		UnloadRecordings(loader, template.Recordings),
		UnloadShaders(loader, template.Shaders),
		UnloadTextures(loader, template.Textures),
		UnloadMaterials(loader, template.Materials),
		UnloadMeshGeometries(loader, template.MeshGeometries),
		UnloadMeshDefinitions(loader, template.MeshDefinitions),

		UnloadNodeTemplates(loader, template.Nodes),
		UnloadArmatureTemplates(loader, template.Armatures),
		UnloadMeshTemplates(loader, template.Meshes),
		UnloadAmbientLightTemplates(loader, template.AmbientLights),
		UnloadPointLightTemplates(loader, template.PointLights),
		UnloadSpotLightTemplates(loader, template.SpotLights),
		UnloadDirectionalLightTemplates(loader, template.DirectionalLights),
		UnloadSkyTemplates(loader, template.SkyTemplates),
	)
}

// ModelInfo contains the information necessary to place a Model
// instance into a Scene.
type ModelInfo struct {

	// Template specifies the template from which this instance will
	// be created.
	Template *ModelTemplate

	// Name specifies the name of this instance.
	Name opt.T[string]

	// SubTreeNode specifies the name of the root node of the model to use, in
	// which case a wrapper root node will not be created. The selected root node
	// will be renamed to Name if it is specified.
	SubTreeNode opt.T[string]

	// Position is used to specify a location for the model instance.
	Position opt.T[dprec.Vec3]

	// Rotation is used to specify a rotation for the model instance.
	Rotation opt.T[dprec.Quat]

	// Scale is used to specify a scale for the model instance.
	Scale opt.T[dprec.Vec3]

	// IsDynamic determines whether the model can be repositioned once
	// placed in the Scene - whether it should be added to the scene hierarchy.
	//
	// TODO: Base this on individual node flags.
	IsDynamic bool

	// DiscardHierarchy will remove the model's hierarchy from the scene
	// after instantiation, leaving only the graphical and physical
	// representations.
	//
	// This is applicable only for static models.
	DiscardHierarchy bool
}

// Model represents an instance of a ModelTemplate in a Scene.
type Model struct {
	scene      *Scene
	root       hierarchy.NodeID
	nodes      IdentifiableList[hierarchy.NodeID]
	recordings []*animation.Clip
	skeleton   *animation.Skeleton
}

// Skeleton returns the skeleton that resolves this model's node names to bone
// indices and holds their rest transforms. It is used when constructing
// animation node trees for this model.
func (m *Model) Skeleton() *animation.Skeleton {
	return m.skeleton
}

// Root returns the root node of the model hierarchy.
func (m *Model) Root() hierarchy.NodeID {
	return m.root
}

func (m *Model) Nodes() IdentifiableList[hierarchy.NodeID] {
	return m.nodes
}

// FindNode is a convenience method that searches for a node
// by its name in the model hierarchy.
func (m *Model) FindNode(name string) hierarchy.NodeID {
	return m.scene.Hierarchy().Nodes().FindNodeInSubtree(m.root, name)
}

func (m *Model) Recordings() []*animation.Clip {
	return m.recordings
}

func (m *Model) FindRecording(name string) *animation.Clip {
	for _, clip := range m.recordings {
		if clip.Name() == name {
			return clip
		}
	}
	return nil
}

func (m *Model) AnimatedNodes() []hierarchy.NodeID {
	result := ds.EmptySet[hierarchy.NodeID]()
	for _, clip := range m.recordings {
		for nodeName := range clip.ChannelNamesIter() {
			if nodeID := m.FindNode(nodeName); nodeID != hierarchy.NilNodeID {
				result.Add(nodeID)
			}
		}
	}
	return result.Items()
}

// BindAnimation binds the animated nodes of the model (see
// [Model.AnimatedNodes]) to the bones of the specified animation node tree
// through the [Scene.ArmatureBindingSet]. The tree is not advanced by this;
// use [Scene.PlayAnimation] for that.
func (m *Model) BindAnimation(root animation.Node) {
	m.bindAnimationNodes(root, m.AnimatedNodes())
}

// BindAnimationSubtree binds all nodes in the subtree of the node with the
// specified name to the bones of the specified animation node tree through
// the [Scene.ArmatureBindingSet]. The tree is not advanced by this; use
// [Scene.PlayAnimation] for that.
func (m *Model) BindAnimationSubtree(root animation.Node, nodeName string) {
	subtreeNodeID := m.FindNode(nodeName)
	var animatedNodeIDs []hierarchy.NodeID
	for nodeID := range m.scene.Hierarchy().Nodes().WalkSubtreeIter(subtreeNodeID) {
		animatedNodeIDs = append(animatedNodeIDs, nodeID)
	}
	m.bindAnimationNodes(root, animatedNodeIDs)
}

func (m *Model) bindAnimationNodes(root animation.Node, nodeIDs []hierarchy.NodeID) {
	nodes := m.scene.Hierarchy().Nodes()
	for _, nodeID := range nodeIDs {
		bone, ok := m.skeleton.BoneIndex(nodes.Handle(nodeID).Name())
		if !ok {
			continue
		}
		m.scene.armatureBindingSet.Bind(nodeID, AnimationSource{
			Node: root,
			Bone: bone,
		})
	}
}

// InstantiateModel instantiates a model in the given scene based on the
// provided info.
func InstantiateModel(scene *Scene, info ModelInfo) *Model {
	hierarchyInfo := HierarchyInfo{
		NodeTemplates: info.Template.Nodes,
		Name:          info.Name,
		Position:      info.Position,
		Rotation:      info.Rotation,
		Scale:         info.Scale,
		SubTreeNode:   info.SubTreeNode,
	}

	hierarchyInstance := InstantiateHierarchy(scene, hierarchyInfo)
	modelNode := hierarchyInstance.RootNode
	nodes := hierarchyInstance.Nodes

	definition := info.Template
	textures := definition.Textures
	recordings := definition.Recordings
	meshDefinitions := definition.MeshDefinitions

	for template := range definition.Terrains.Values() {
		if nodes.HasID(template.NodeID) {
			InstantiatePhysicsTerrainTemplate(scene, template, nodes)
		}
	}

	armatures := make(IdentifiableList[*graphics.Armature], 0, len(definition.Armatures))
	for id, template := range definition.Armatures.Iter() {
		armature := InstantiateArmatureTemplate(scene, template, nodes)
		armatures = append(armatures, Identifiable[*graphics.Armature]{
			ID:    id,
			Value: armature,
		})
	}

	for template := range definition.Meshes.Values() {
		if nodes.HasID(template.NodeID) {
			if info.IsDynamic {
				InstantiateMeshTemplateDynamic(scene, template, nodes, meshDefinitions, armatures)
			} else {
				InstantiateMeshTemplateStatic(scene, template, nodes, meshDefinitions, armatures)
			}
		}
	}

	for template := range definition.AmbientLights.Values() {
		if nodes.HasID(template.NodeID) {
			InstantiateAmbientLightTemplate(scene, template, nodes, textures)
		}
	}
	for template := range definition.PointLights.Values() {
		if nodes.HasID(template.NodeID) {
			InstantiatePointLightTemplate(scene, template, nodes)
		}
	}
	for template := range definition.SpotLights.Values() {
		if nodes.HasID(template.NodeID) {
			InstantiateSpotLightTemplate(scene, template, nodes)
		}
	}
	for template := range definition.DirectionalLights.Values() {
		if nodes.HasID(template.NodeID) {
			InstantiateDirectionalLightTemplate(scene, template, nodes)
		}
	}
	for template := range definition.SkyTemplates.Values() {
		if nodes.HasID(template.NodeID) {
			InstantiateSkyTemplate(scene, template, nodes)
		}
	}

	scene.Hierarchy().Nodes().Snap(modelNode)
	scene.Hierarchy().ApplySourceToNode(modelNode, true)
	scene.Hierarchy().ApplyTargetFromNode(modelNode, true)
	scene.Hierarchy().ApplyInterpolationFromNode(modelNode, 1.0, true)

	// Build the animation skeleton from the model's nodes, so that animation
	// node trees can resolve bone names to indices and fall back to the rest
	// transforms of the nodes. This must happen before any hierarchy discard
	// below.
	bones := make([]animation.SkeletonBone, 0, len(nodes))
	for nodeID := range nodes.Values() {
		node := scene.Hierarchy().Nodes().Handle(nodeID)
		bones = append(bones, animation.SkeletonBone{
			Name: node.Name(),
			RestTransform: animation.Transform{
				Translation: node.Position(),
				Rotation:    node.Rotation(),
				Scale:       node.Scale(),
			},
		})
	}
	skeleton := animation.NewSkeleton(bones)

	if !info.IsDynamic && info.DiscardHierarchy {
		for node := range nodes.Values() {
			scene.AmbientLightBindingSet().Unbind(node)
			scene.PointLightBindingSet().Unbind(node)
			scene.SpotLightBindingSet().Unbind(node)
			scene.DirectionalLightBindingSet().Unbind(node)
			scene.SkyBindingSet().Unbind(node)
		}
		scene.Hierarchy().Nodes().Delete(modelNode)
	}

	return &Model{
		scene:      scene,
		root:       modelNode,
		nodes:      nodes,
		recordings: recordings.ValuesList(),
		skeleton:   skeleton,
	}
}

func (s *Scene) InstantiateModel(info ModelInfo) *Model {
	return InstantiateModel(s, info)
}
