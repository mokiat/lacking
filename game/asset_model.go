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
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/render"
)

// ModelTemplate represents a template for a model that can be instantiated
// in a Scene.
type ModelTemplate struct {
	Recordings      IdentifiableList[*animation.Recording]
	Shaders         IdentifiableList[*graphics.Shader]
	Textures        IdentifiableList[render.Texture]
	Materials       IdentifiableList[*graphics.Material]
	BodyMaterials   IdentifiableList[*physics.Material]
	BodyDefinitions IdentifiableList[*physics.BodyDefinition]
	MeshGeometries  IdentifiableList[*graphics.MeshGeometry]
	MeshDefinitions IdentifiableList[*graphics.MeshDefinition]

	Nodes             IdentifiableList[NodeTemplate]
	Bodies            IdentifiableList[BodyTemplate]
	Armatures         IdentifiableList[ArmatureTemplate]
	Meshes            IdentifiableList[MeshTemplate]
	AmbientLights     IdentifiableList[AmbientLightTemplate]
	PointLights       IdentifiableList[PointLightTemplate]
	SpotLights        IdentifiableList[SpotLightTemplate]
	DirectionalLights IdentifiableList[DirectionalLightTemplate]
	SkyTemplates      IdentifiableList[SkyTemplate]
}

func (t *ModelTemplate) FindRecording(name string) *animation.Recording {
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
	recordings, err := LoadAnimationRecordings(loader, assetModel.AnimationChunk.Animations)
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

	bodyMaterials, err := LoadPhysicsMaterials(loader, assetModel.PhysicsChunk.BodyMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve body materials: %w", err)
	}

	bodyDefinitions, err := LoadPhysicsBodyDefinitions(loader, assetModel.PhysicsChunk.BodyDefinitions, bodyMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve body definitions: %w", err)
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

	bodies, err := LoadPhysicsBodyTemplates(loader, assetModel.PhysicsChunk.Bodies, bodyDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve physics body templates: %w", err)
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
		BodyMaterials:   bodyMaterials,
		BodyDefinitions: bodyDefinitions,
		MeshGeometries:  meshGeometries,
		MeshDefinitions: meshDefinitions,

		Nodes:             nodes,
		Bodies:            bodies,
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
		UnloadPhysicsMaterials(loader, template.BodyMaterials),
		UnloadPhysicsBodyDefinitions(loader, template.BodyDefinitions),
		UnloadMeshGeometries(loader, template.MeshGeometries),
		UnloadMeshDefinitions(loader, template.MeshDefinitions),

		UnloadNodeTemplates(loader, template.Nodes),
		UnloadPhysicsBodyTemplates(loader, template.Bodies),
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
}

// Model represents an instance of a ModelTemplate in a Scene.
type Model struct {
	root       *hierarchy.Node
	nodes      IdentifiableList[*hierarchy.Node]
	recordings []*animation.Recording
}

// Root returns the root node of the model hierarchy.
func (m *Model) Root() *hierarchy.Node {
	return m.root
}

func (m *Model) Nodes() IdentifiableList[*hierarchy.Node] {
	return m.nodes
}

// FindNode is a convenience method that searches for a node
// by its name in the model hierarchy.
func (m *Model) FindNode(name string) *hierarchy.Node {
	return m.root.FindNode(name)
}

func (m *Model) Recordings() []*animation.Recording {
	return m.recordings
}

func (m *Model) FindRecording(name string) *animation.Recording {
	for _, animation := range m.recordings {
		if animation.Name() == name {
			return animation
		}
	}
	return nil
}

func (m *Model) AnimatedNodes() []*hierarchy.Node {
	result := ds.NewSet[*hierarchy.Node](0)
	for _, animation := range m.recordings {
		for nodeName := range animation.BoundNodesIter() {
			if node := m.FindNode(nodeName); node != nil {
				result.Add(node)
			}
		}
	}
	return result.Items()
}

func (m *Model) BindAnimation(root animation.Node) *animation.Player {
	animatedNodes := m.AnimatedNodes()
	return m.bindAnimationNodes(root, animatedNodes)
}

func (m *Model) BindAnimationSubtree(root animation.Node, nodeName string) *animation.Player {
	var animatedNodes []*hierarchy.Node
	subtreeNode := m.FindNode(nodeName)
	hierarchy.EachNode(subtreeNode, func(node *hierarchy.Node) {
		animatedNodes = append(animatedNodes, node)
	})
	return m.bindAnimationNodes(root, animatedNodes)
}

func (m *Model) bindAnimationNodes(root animation.Node, nodes []*hierarchy.Node) *animation.Player {
	player := animation.NewPlayer(root)
	for _, node := range nodes {
		node.SetSource(AnimationNodeSource{
			Player: player,
		})
	}
	return player
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
		AttachToScene: opt.V(info.IsDynamic),
	}

	hierarchyInstance := InstantiateHierarchy(scene, hierarchyInfo)
	modelNode := hierarchyInstance.RootNode
	nodes := hierarchyInstance.Nodes

	definition := info.Template
	textures := definition.Textures
	recordings := definition.Recordings
	meshDefinitions := definition.MeshDefinitions

	for template := range definition.Bodies.Values() {
		if nodes.HasID(template.NodeID) {
			if info.IsDynamic {
				InstantiatePhysicsBodyTemplateDynamic(scene, template, nodes)
			} else {
				InstantiatePhysicsBodyTemplateStatic(scene, template, nodes)
			}
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

	modelNode.ResetDelta()
	modelNode.ApplyFromSource(true)
	modelNode.ApplyToTarget(true)

	return &Model{
		root:       modelNode,
		nodes:      nodes,
		recordings: recordings.ValuesList(),
	}
}

func (s *Scene) InstantiateModel(info ModelInfo) *Model {
	return InstantiateModel(s, info)
}
