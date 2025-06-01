package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/storage/chunked"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) loadModel(asyncEngine *AsyncEngine, resource *chunked.Asset) (*ModelDefinition, error) {
	assetModelPromise := s.openResource(resource)
	assetModel, err := assetModelPromise.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	return s.convertModel(asyncEngine, assetModel)
}

func (s *ResourceSet) freeModel(model *ModelDefinition) {
	s.gfxWorker.Schedule(func() {
		for skyTemplate := range model.skyTemplates.Values() {
			definition := skyTemplate.Definition
			definition.Delete()
		}
	})
	s.gfxWorker.Schedule(func() {
		for _, meshDefinition := range model.meshDefinitions {
			meshDefinition.Delete()
		}
	})
	s.gfxWorker.Schedule(func() {
		for _, meshGeometry := range model.meshGeometries {
			meshGeometry.Delete()
		}
	})
	s.gfxWorker.Schedule(func() {
		for texture := range model.textures.Values() {
			texture.Release()
		}
	})
}

func (s *ResourceSet) openResource(resource *chunked.Asset) async.Promise[dto.Model] {
	promise := async.NewPromise[dto.Model]()
	s.ioWorker.Schedule(func() {
		var assetModel dto.Model
		if err := resource.Read(&assetModel); err != nil {
			promise.Fail(err)
		} else {
			promise.Deliver(assetModel)
		}
	})
	return promise
}

func (s *ResourceSet) convertModel(asyncEngine *AsyncEngine, assetModel dto.Model) (*ModelDefinition, error) {
	// TODO: Figure out how to better get this working and extensible.
	loader := &AssetLoader{
		resourceSet: s,
		asyncEngine: asyncEngine,
	}

	recordings, err := loader.ResolveAnimationRecordings(assetModel.AnimationChunk.Animations)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve animation recordings: %w", err)
	}

	shaders, err := loader.ResolveShaders(assetModel.ShadingChunk.Shaders)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve shaders: %w", err)
	}

	armatures := make([]armatureDefinition, len(assetModel.MeshChunk.Armatures))
	for i, assetArmature := range assetModel.MeshChunk.Armatures {
		armatures[i] = s.convertArmature(assetArmature)
	}
	armatureIndexByID := make(map[uint32]int, len(assetModel.MeshChunk.Armatures))
	for i, assetArmature := range assetModel.MeshChunk.Armatures {
		armatureIndexByID[assetArmature.ID] = i
	}

	// TODO: Convert cameras

	textures, err := loader.ResolveTextures(assetModel.ShadingChunk.Textures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve textures: %w", err)
	}

	materialPromises := make([]async.Promise[*graphics.Material], len(assetModel.ShadingChunk.Materials))
	for i, assetMaterial := range assetModel.ShadingChunk.Materials {
		materialPromises[i] = s.convertMaterial(
			shaders,
			textures,
			assetMaterial,
		)
	}
	materials, err := async.WaitPromises(materialPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert materials: %w", err)
	}
	identifiableMaterials := make(IdentifiableList[*graphics.Material], len(assetModel.ShadingChunk.Materials))
	for i, assetMaterial := range assetModel.ShadingChunk.Materials {
		identifiableMaterials[i] = Identifiable[*graphics.Material]{
			ID:    assetMaterial.ID,
			Value: materials[i],
		}
	}

	materialByID := make(map[uint32]*graphics.Material, len(assetModel.ShadingChunk.Materials))
	for i, assetMaterial := range assetModel.ShadingChunk.Materials {
		materialByID[assetMaterial.ID] = materials[i]
	}

	meshGeometryPromises := make([]async.Promise[*graphics.MeshGeometry], len(assetModel.MeshChunk.Geometries))
	for i, assetGeometry := range assetModel.MeshChunk.Geometries {
		meshGeometryPromises[i] = s.convertMeshGeometry(assetGeometry)
	}
	meshGeometries, err := async.WaitPromises(meshGeometryPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mesh geometries: %w", err)
	}
	meshGeometryByID := make(map[uint32]*graphics.MeshGeometry, len(assetModel.MeshChunk.Geometries))
	for i, assetGeometry := range assetModel.MeshChunk.Geometries {
		meshGeometryByID[assetGeometry.ID] = meshGeometries[i]
	}

	meshDefinitionPromises := make([]async.Promise[*graphics.MeshDefinition], len(assetModel.MeshChunk.MeshDefinitions))
	for i, assetMeshDefinition := range assetModel.MeshChunk.MeshDefinitions {
		meshDefinitionPromises[i] = s.convertMeshDefinition(
			meshGeometryByID,
			materialByID,
			assetMeshDefinition,
		)
	}
	meshDefinitions, err := async.WaitPromises(meshDefinitionPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mesh definitions: %w", err)
	}
	meshDefinitionIndexByID := make(map[uint32]int, len(assetModel.MeshChunk.MeshDefinitions))
	for i, assetMeshDefinition := range assetModel.MeshChunk.MeshDefinitions {
		meshDefinitionIndexByID[assetMeshDefinition.ID] = i
	}

	meshes := make([]meshInstance, len(assetModel.MeshChunk.Meshes))
	for i, assetMesh := range assetModel.MeshChunk.Meshes {
		meshes[i] = s.convertMeshInstance(
			armatureIndexByID,
			meshDefinitionIndexByID,
			assetMesh,
		)
	}

	bodyMaterialPromises := make([]async.Promise[*physics.Material], len(assetModel.PhysicsChunk.BodyMaterials))
	for i, assetBodyMaterial := range assetModel.PhysicsChunk.BodyMaterials {
		bodyMaterialPromises[i] = s.convertBodyMaterial(assetBodyMaterial)
	}
	bodyMaterials, err := async.WaitPromises(bodyMaterialPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert body materials: %w", err)
	}
	bodyMaterialByID := make(map[uint32]*physics.Material, len(bodyMaterials))
	for i, assetBodyMaterial := range assetModel.PhysicsChunk.BodyMaterials {
		bodyMaterialByID[assetBodyMaterial.ID] = bodyMaterials[i]
	}

	bodyDefinitionPromises := make([]async.Promise[*physics.BodyDefinition], len(assetModel.PhysicsChunk.BodyDefinitions))
	for i, assetBodyDefinition := range assetModel.PhysicsChunk.BodyDefinitions {
		bodyDefinitionPromises[i] = s.convertBodyDefinition(
			bodyMaterialByID,
			assetBodyDefinition,
		)
	}
	bodyDefinitions, err := async.WaitPromises(bodyDefinitionPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert body definitions: %w", err)
	}
	bodyDefinitionIndexByID := make(map[uint32]int, len(bodyDefinitions))
	for i, assetBodyDefinition := range assetModel.PhysicsChunk.BodyDefinitions {
		bodyDefinitionIndexByID[assetBodyDefinition.ID] = i
	}

	bodies := make([]bodyInstance, len(assetModel.PhysicsChunk.Bodies))
	for i, assetBody := range assetModel.PhysicsChunk.Bodies {
		bodies[i] = s.convertBody(bodyDefinitionIndexByID, assetBody)
	}

	nodes, err := loader.ResolveNodeTemplates(assetModel.HierarchyChunk.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node templates: %w", err)
	}

	ambientLights, err := loader.ResolveAmbientLightTemplates(assetModel.LightingChunk.AmbientLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ambient light templates: %w", err)
	}

	pointLights, err := loader.ResolvePointLightTemplates(assetModel.LightingChunk.PointLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve point light templates: %w", err)
	}

	spotLights, err := loader.ResolveSpotLightTemplates(assetModel.LightingChunk.SpotLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve spot light templates: %w", err)
	}

	directionalLights, err := loader.ResolveDirectionalLightTemplates(assetModel.LightingChunk.DirectionalLights)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directional light templates: %w", err)
	}

	skyTemplates, err := loader.ResolveSkyTemplates(assetModel.BackgroundChunk.Skies, identifiableMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve sky templates: %w", err)
	}

	return &ModelDefinition{
		recordings: recordings,
		shaders:    shaders,
		textures:   textures,

		armatures:       armatures,
		materials:       materials,
		meshGeometries:  meshGeometries,
		meshDefinitions: meshDefinitions,
		meshes:          meshes,
		bodyMaterials:   bodyMaterials,
		bodyDefinitions: bodyDefinitions,
		bodies:          bodies,

		nodes:             nodes,
		ambientLights:     ambientLights,
		pointLights:       pointLights,
		spotLights:        spotLights,
		directionalLights: directionalLights,
		skyTemplates:      skyTemplates,
	}, nil
}
