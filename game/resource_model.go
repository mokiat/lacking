package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) loadModel(resource *asset.Resource) (*ModelDefinition, error) {
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

func (s *ResourceSet) openResource(resource *asset.Resource) async.Promise[asset.Model] {
	promise := async.NewPromise[asset.Model]()
	s.ioWorker.Schedule(func() {
		assetModel, err := resource.OpenContent()
		if err != nil {
			promise.Fail(err)
		} else {
			promise.Deliver(assetModel)
		}
	})
	return promise
}

func (s *ResourceSet) convertModel(assetModel asset.Model) (*ModelDefinition, error) {
	nodes := make([]nodeDefinition, len(assetModel.Nodes))
	for i, assetNode := range assetModel.Nodes {
		nodes[i] = s.convertNode(assetNode)
	}

	animationPromises := make([]async.Promise[*AnimationDefinition], len(assetModel.Animations))
	for i, assetAnimation := range assetModel.Animations {
		animationPromises[i] = s.convertAnimation(assetAnimation)
	}
	animations, err := async.WaitPromises(animationPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert animations: %w", err)
	}

	armatures := make([]armatureDefinition, len(assetModel.Armatures))
	for i, assetArmature := range assetModel.Armatures {
		armatures[i] = s.convertArmature(assetArmature)
	}

	// TODO: Convert cameras

	shaderPromises := make([]async.Promise[*graphics.Shader], len(assetModel.Shaders))
	for i, assetShader := range assetModel.Shaders {
		shaderPromises[i] = s.convertShader(assetShader)
	}
	shaders, err := async.WaitPromises(shaderPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shaders: %w", err)
	}

	texturePromises := make([]async.Promise[render.Texture], len(assetModel.Textures))
	for i, assetTexture := range assetModel.Textures {
		texturePromises[i] = s.convertTexture(assetTexture)
	}
	textures, err := async.WaitPromises(texturePromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert textures: %w", err)
	}

	materialPromises := make([]async.Promise[*graphics.Material], len(assetModel.Materials))
	for i, assetMaterial := range assetModel.Materials {
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

	meshGeometryPromises := make([]async.Promise[*graphics.MeshGeometry], len(assetModel.Geometries))
	for i, assetGeometry := range assetModel.Geometries {
		meshGeometryPromises[i] = s.convertMeshGeometry(assetGeometry)
	}
	meshGeometries, err := async.WaitPromises(meshGeometryPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mesh geometries: %w", err)
	}

	meshDefinitionPromises := make([]async.Promise[*graphics.MeshDefinition], len(assetModel.MeshDefinitions))
	for i, assetMeshDefinition := range assetModel.MeshDefinitions {
		meshDefinitionPromises[i] = s.convertMeshDefinition(
			meshGeometries,
			materials,
			assetMeshDefinition,
		)
	}
	meshDefinitions, err := async.WaitPromises(meshDefinitionPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert mesh definitions: %w", err)
	}

	meshes := make([]meshInstance, len(assetModel.Meshes))
	for i, assetMesh := range assetModel.Meshes {
		meshes[i] = s.convertMeshInstance(assetMesh)
	}

	bodyMaterialPromises := make([]async.Promise[*physics.Material], len(assetModel.BodyMaterials))
	for i, assetBodyMaterial := range assetModel.BodyMaterials {
		bodyMaterialPromises[i] = s.convertBodyMaterial(assetBodyMaterial)
	}
	bodyMaterials, err := async.WaitPromises(bodyMaterialPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert body materials: %w", err)
	}

	bodyDefinitionPromises := make([]async.Promise[*physics.BodyDefinition], len(assetModel.BodyDefinitions))
	for i, assetBodyDefinition := range assetModel.BodyDefinitions {
		bodyDefinitionPromises[i] = s.convertBodyDefinition(
			bodyMaterials,
			assetBodyDefinition,
		)
	}
	bodyDefinitions, err := async.WaitPromises(bodyDefinitionPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert body definitions: %w", err)
	}

	bodies := make([]bodyInstance, len(assetModel.Bodies))
	for i, assetBody := range assetModel.Bodies {
		bodies[i] = s.convertBody(assetBody)
	}

	ambientLights := make([]ambientLightInstance, len(assetModel.AmbientLights))
	for i, assetAmbientLight := range assetModel.AmbientLights {
		ambientLights[i] = s.convertAmbientLight(
			textures,
			assetAmbientLight,
		)
	}

	pointLights := make([]pointLightInstance, len(assetModel.PointLights))
	for i, assetPointLight := range assetModel.PointLights {
		pointLights[i] = s.convertPointLight(assetPointLight)
	}

	spotLights := make([]spotLightInstance, len(assetModel.SpotLights))
	for i, assetSpotLight := range assetModel.SpotLights {
		spotLights[i] = s.convertSpotLight(assetSpotLight)
	}

	directionalLights := make([]directionalLightInstance, len(assetModel.DirectionalLights))
	for i, assetDirectionalLight := range assetModel.DirectionalLights {
		directionalLights[i] = s.convertDirectionalLight(assetDirectionalLight)
	}

	skyDefinitionPromises := make([]async.Promise[*graphics.SkyDefinition], len(assetModel.Skies))
	for i, assetSky := range assetModel.Skies {
		skyDefinitionPromises[i] = s.convertSkyDefinition(
			materials,
			assetSky,
		)
	}
	skyDefinitions, err := async.WaitPromises(skyDefinitionPromises...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sky definitions: %w", err)
	}

	skies := make([]skyInstance, len(assetModel.Skies))
	for i, assetSky := range assetModel.Skies {
		skies[i] = s.convertSky(i, assetSky)
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
