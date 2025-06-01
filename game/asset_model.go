package game

import (
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

func (l *AssetLoader) ResolveModelTemplate(assetModel dto.Model) (*ModelTemplate, error) {
	recordings, err := l.ResolveAnimationRecordings(assetModel.AnimationChunk.Animations)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve animation recordings: %w", err)
	}

	shaders, err := l.ResolveShaders(assetModel.ShadingChunk.Shaders)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve shaders: %w", err)
	}

	textures, err := l.ResolveTextures(assetModel.ShadingChunk.Textures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve textures: %w", err)
	}

	materials, err := l.ResolveMaterials(assetModel.ShadingChunk.Materials, shaders, textures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve materials: %w", err)
	}

	bodyMaterials, err := l.ResolvePhysicsMaterials(assetModel.PhysicsChunk.BodyMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve body materials: %w", err)
	}

	bodyDefinitions, err := l.ResolvePhysicsBodyDefinitions(assetModel.PhysicsChunk.BodyDefinitions, bodyMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve body definitions: %w", err)
	}

	meshGeometries, err := l.ResolveMeshGeometries(assetModel.MeshChunk.Geometries)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh geometries: %w", err)
	}

	meshDefinitions, err := l.ResolveMeshDefinitions(assetModel.MeshChunk.MeshDefinitions, meshGeometries, materials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh definitions: %w", err)
	}

	nodes, err := l.ResolveNodeTemplates(assetModel.HierarchyChunk.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node templates: %w", err)
	}

	bodies, err := l.ResolvePhysicsBodyTemplates(assetModel.PhysicsChunk.Bodies, bodyDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve physics body templates: %w", err)
	}

	armatures, err := l.ResolveArmatureTemplates(assetModel.MeshChunk.Armatures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve armature templates: %w", err)
	}

	meshes, err := l.ResolveMeshTemplates(assetModel.MeshChunk.Meshes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh templates: %w", err)
	}

	ambientLights, err := l.ResolveAmbientLightTemplates(assetModel.LightingChunk.AmbientLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ambient light templates: %w", err)
	}

	pointLights, err := l.ResolvePointLightTemplates(assetModel.LightingChunk.PointLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve point light templates: %w", err)
	}

	spotLights, err := l.ResolveSpotLightTemplates(assetModel.LightingChunk.SpotLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve spot light templates: %w", err)
	}

	directionalLights, err := l.ResolveDirectionalLightTemplates(assetModel.LightingChunk.DirectionalLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directional light templates: %w", err)
	}

	// TODO: Convert cameras

	skyTemplates, err := l.ResolveSkyTemplates(assetModel.BackgroundChunk.Skies, materials)
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
	recordings []*animation.Recording
}

// Root returns the root node of the model hierarchy.
func (m *Model) Root() *hierarchy.Node {
	return m.root
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
		for nodeName := range animation.BoundNodes() {
			if node := m.FindNode(nodeName); node != nil {
				result.Add(node)
			}
		}
	}
	return result.Items()
}

func (m *Model) BindAnimationSource(source animation.Source) {
	for _, node := range m.AnimatedNodes() {
		node.SetSource(AnimationNodeSource{
			Source: source,
		})
	}
}

func (s *Scene) InstantiateModel(info ModelInfo) *Model {
	hierarchyInfo := HierarchyInfo{
		NodeTemplates: info.Template.Nodes,
		Name:          info.Name,
		Position:      info.Position,
		Rotation:      info.Rotation,
		Scale:         info.Scale,
		SubTreeNode:   info.SubTreeNode,
		AttachToScene: opt.V(info.IsDynamic),
	}

	hierarchyInstance := s.InstantiateHierarchy(hierarchyInfo)
	modelNode := hierarchyInstance.RootNode
	nodes := hierarchyInstance.Nodes

	definition := info.Template
	textures := definition.Textures
	recordings := definition.Recordings
	meshDefinitions := definition.MeshDefinitions

	for template := range definition.Bodies.Values() {
		if nodes.HasID(template.NodeID) {
			if info.IsDynamic {
				s.InstantiatePhysicsBodyTemplateDynamic(template, nodes)
			} else {
				s.InstantiatePhysicsBodyTemplateStatic(template, nodes)
			}
		}
	}

	armatures := make(IdentifiableList[*graphics.Armature], 0, len(definition.Armatures))
	for id, template := range definition.Armatures.Iter() {
		armature := s.InstantiateArmatureTemplate(template, nodes)
		armatures = append(armatures, Identifiable[*graphics.Armature]{
			ID:    id,
			Value: armature,
		})
	}

	for template := range definition.Meshes.Values() {
		if nodes.HasID(template.NodeID) {
			if info.IsDynamic {
				s.InstantiateMeshTemplateDynamic(template, nodes, meshDefinitions, armatures)
			} else {
				s.InstantiateMeshTemplateStatic(template, nodes, meshDefinitions, armatures)
			}
		}
	}

	for template := range definition.AmbientLights.Values() {
		if nodes.HasID(template.NodeID) {
			s.InstantiateAmbientLightTemplate(template, nodes, textures)
		}
	}
	for template := range definition.PointLights.Values() {
		if nodes.HasID(template.NodeID) {
			s.InstantiatePointLightTemplate(template, nodes)
		}
	}
	for template := range definition.SpotLights.Values() {
		if nodes.HasID(template.NodeID) {
			s.InstantiateSpotLightTemplate(template, nodes)
		}
	}
	for template := range definition.DirectionalLights.Values() {
		if nodes.HasID(template.NodeID) {
			s.InstantiateDirectionalLightTemplate(template, nodes)
		}
	}
	for template := range definition.SkyTemplates.Values() {
		if nodes.HasID(template.NodeID) {
			s.InstantiateSkyTemplate(template, nodes)
		}
	}

	modelNode.ApplyFromSource(true)
	modelNode.ApplyToTarget(true)

	return &Model{
		root:       modelNode,
		recordings: recordings.ValuesList(),
	}
}
