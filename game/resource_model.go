package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/storage/chunked"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) loadModel(resource *chunked.Asset) (*ModelDefinition, error) {
	assetModelPromise := s.openResource(resource)
	assetModel, err := assetModelPromise.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	return s.convertModel(assetModel)
}

func (s *ResourceSet) freeModel(model *ModelDefinition) {
	s.gfxWorker.Schedule(func() {
		for _, skyDefinition := range model.skyDefinitions {
			skyDefinition.Delete()
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
		for _, texture := range model.textures {
			texture.Release()
		}
	})
}

func (s *ResourceSet) openResource(resource *chunked.Asset) async.Promise[asset.Model] {
	promise := async.NewPromise[asset.Model]()
	s.ioWorker.Schedule(func() {
		var assetModel asset.Model
		if err := resource.Read(&assetModel); err != nil {
			promise.Fail(err)
		} else {
			promise.Deliver(assetModel)
		}
	})
	return promise
}

func (s *ResourceSet) convertModel(assetModel asset.Model) (*ModelDefinition, error) {
	nodes := make([]nodeDefinition, len(assetModel.HierarchyChunk.Nodes))
	for i, assetNode := range assetModel.HierarchyChunk.Nodes {
		nodes[i] = s.convertNode(assetNode)
	}
	nodeIndexByID := make(map[uint32]int, len(assetModel.HierarchyChunk.Nodes))
	for i, assetNode := range assetModel.HierarchyChunk.Nodes {
		nodeIndexByID[assetNode.ID] = i
	}

	animationPromises := make([]async.Promise[*AnimationDefinition], len(assetModel.AnimationChunk.Animations))
	for i, assetAnimation := range assetModel.AnimationChunk.Animations {
		animationPromises[i] = s.convertAnimation(assetAnimation)
	}
	animations, err := async.WaitPromises(animationPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert animations: %w", err)
	}

	armatures := make([]armatureDefinition, len(assetModel.MeshChunk.Armatures))
	for i, assetArmature := range assetModel.MeshChunk.Armatures {
		armatures[i] = s.convertArmature(
			nodeIndexByID,
			assetArmature,
		)
	}
	armatureIndexByID := make(map[uint32]int, len(assetModel.MeshChunk.Armatures))
	for i, assetArmature := range assetModel.MeshChunk.Armatures {
		armatureIndexByID[assetArmature.ID] = i
	}

	// TODO: Convert cameras

	shaderPromises := make([]async.Promise[*graphics.Shader], len(assetModel.ShadingChunk.Shaders))
	for i, assetShader := range assetModel.ShadingChunk.Shaders {
		shaderPromises[i] = s.convertShader(assetShader)
	}
	shaders, err := async.WaitPromises(shaderPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shaders: %w", err)
	}
	shaderByID := make(map[uint32]*graphics.Shader, len(assetModel.ShadingChunk.Shaders))
	for i, assetShader := range assetModel.ShadingChunk.Shaders {
		shaderByID[assetShader.ID] = shaders[i]
	}

	texturePromises := make([]async.Promise[render.Texture], len(assetModel.ShadingChunk.Textures))
	for i, assetTexture := range assetModel.ShadingChunk.Textures {
		texturePromises[i] = s.convertTexture(assetTexture)
	}
	textures, err := async.WaitPromises(texturePromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert textures: %w", err)
	}
	textureByID := make(map[uint32]render.Texture, len(assetModel.ShadingChunk.Textures))
	for i, assetTexture := range assetModel.ShadingChunk.Textures {
		textureByID[assetTexture.ID] = textures[i]
	}

	materialPromises := make([]async.Promise[*graphics.Material], len(assetModel.ShadingChunk.Materials))
	for i, assetMaterial := range assetModel.ShadingChunk.Materials {
		materialPromises[i] = s.convertMaterial(
			shaderByID,
			textureByID,
			assetMaterial,
		)
	}
	materials, err := async.WaitPromises(materialPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert materials: %w", err)
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
			nodeIndexByID,
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
		bodies[i] = s.convertBody(nodeIndexByID, bodyDefinitionIndexByID, assetBody)
	}

	ambientLights := make([]ambientLightInstance, len(assetModel.LightingChunk.AmbientLights))
	for i, assetAmbientLight := range assetModel.LightingChunk.AmbientLights {
		ambientLights[i] = s.convertAmbientLight(
			nodeIndexByID,
			textureByID,
			assetAmbientLight,
		)
	}

	pointLights := make([]pointLightInstance, len(assetModel.LightingChunk.PointLights))
	for i, assetPointLight := range assetModel.LightingChunk.PointLights {
		pointLights[i] = s.convertPointLight(
			nodeIndexByID,
			assetPointLight,
		)
	}

	spotLights := make([]spotLightInstance, len(assetModel.LightingChunk.SpotLights))
	for i, assetSpotLight := range assetModel.LightingChunk.SpotLights {
		spotLights[i] = s.convertSpotLight(
			nodeIndexByID,
			assetSpotLight,
		)
	}

	directionalLights := make([]directionalLightInstance, len(assetModel.LightingChunk.DirectionalLights))
	for i, assetDirectionalLight := range assetModel.LightingChunk.DirectionalLights {
		directionalLights[i] = s.convertDirectionalLight(
			nodeIndexByID,
			assetDirectionalLight,
		)
	}

	skyDefinitionPromises := make([]async.Promise[*graphics.SkyDefinition], len(assetModel.BackgroundChunk.Skies))
	for i, assetSky := range assetModel.BackgroundChunk.Skies {
		skyDefinitionPromises[i] = s.convertSkyDefinition(
			materialByID,
			assetSky,
		)
	}
	skyDefinitions, err := async.WaitPromises(skyDefinitionPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sky definitions: %w", err)
	}

	skies := make([]skyInstance, len(assetModel.BackgroundChunk.Skies))
	for i, assetSky := range assetModel.BackgroundChunk.Skies {
		skies[i] = s.convertSky(nodeIndexByID, i, assetSky)
	}

	return &ModelDefinition{
		nodes:             nodes,
		animations:        animations,
		armatures:         armatures,
		shaders:           shaders,
		textures:          textures,
		materials:         materials,
		meshGeometries:    meshGeometries,
		meshDefinitions:   meshDefinitions,
		meshes:            meshes,
		bodyMaterials:     bodyMaterials,
		bodyDefinitions:   bodyDefinitions,
		bodies:            bodies,
		ambientLights:     ambientLights,
		pointLights:       pointLights,
		spotLights:        spotLights,
		directionalLights: directionalLights,
		skyDefinitions:    skyDefinitions,
		skies:             skies,
	}, nil
}
