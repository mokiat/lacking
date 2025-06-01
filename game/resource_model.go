package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/dto"
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
		for meshDefinition := range model.meshDefinitions.Values() {
			meshDefinition.Delete()
		}
	})
	s.gfxWorker.Schedule(func() {
		for meshGeometry := range model.meshGeometries.Values() {
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

	textures, err := loader.ResolveTextures(assetModel.ShadingChunk.Textures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve textures: %w", err)
	}

	materials, err := loader.ResolveMaterials(assetModel.ShadingChunk.Materials, shaders, textures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve materials: %w", err)
	}

	bodyMaterials, err := loader.ResolvePhysicsMaterials(assetModel.PhysicsChunk.BodyMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve body materials: %w", err)
	}

	bodyDefinitions, err := loader.ResolvePhysicsBodyDefinitions(assetModel.PhysicsChunk.BodyDefinitions, bodyMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve body definitions: %w", err)
	}

	meshGeometries, err := loader.ResolveMeshGeometries(assetModel.MeshChunk.Geometries)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh geometries: %w", err)
	}

	meshDefinitions, err := loader.ResolveMeshDefinitions(assetModel.MeshChunk.MeshDefinitions, meshGeometries, materials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh definitions: %w", err)
	}

	nodes, err := loader.ResolveNodeTemplates(assetModel.HierarchyChunk.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node templates: %w", err)
	}

	bodies, err := loader.ResolvePhysicsBodyTemplates(assetModel.PhysicsChunk.Bodies, bodyDefinitions)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve physics body templates: %w", err)
	}

	armatures, err := loader.ResolveArmatureTemplates(assetModel.MeshChunk.Armatures)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve armature templates: %w", err)
	}

	meshes, err := loader.ResolveMeshTemplates(assetModel.MeshChunk.Meshes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mesh templates: %w", err)
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

	// TODO: Convert cameras

	skyTemplates, err := loader.ResolveSkyTemplates(assetModel.BackgroundChunk.Skies, materials)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve sky templates: %w", err)
	}

	return &ModelDefinition{
		recordings:      recordings,
		shaders:         shaders,
		textures:        textures,
		materials:       materials,
		bodyMaterials:   bodyMaterials,
		bodyDefinitions: bodyDefinitions,
		meshGeometries:  meshGeometries,
		meshDefinitions: meshDefinitions,

		meshes: meshes,

		nodes:             nodes,
		bodies:            bodies,
		armatures:         armatures,
		ambientLights:     ambientLights,
		pointLights:       pointLights,
		spotLights:        spotLights,
		directionalLights: directionalLights,
		skyTemplates:      skyTemplates,
	}, nil
}
