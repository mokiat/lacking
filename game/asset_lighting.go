package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/render"
)

// AmbientLightTemplate represents a template for an ambient light in the scene.
type AmbientLightTemplate struct {
	NodeID              uint32
	ReflectionTextureID uint32
	RefractionTextureID uint32
	CastShadow          bool
}

// LoadAmbientLightTemplate loads an ambient light template from the given
// asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadAmbientLightTemplate(loader *AssetLoader, assetLight dto.AmbientLight) (Identifiable[AmbientLightTemplate], error) {
	return Identifiable[AmbientLightTemplate]{
		ID: assetLight.ID,
		Value: AmbientLightTemplate{
			NodeID:              assetLight.NodeID,
			ReflectionTextureID: assetLight.ReflectionTextureID,
			RefractionTextureID: assetLight.RefractionTextureID,
			CastShadow:          assetLight.CastShadow,
		},
	}, nil
}

// LoadAmbientLightTemplates loads a list of ambient light templates from the
// given asset ambient lights.
//
// This is a blocking operation and should be called from a worker thread.
func LoadAmbientLightTemplates(loader *AssetLoader, assetLights []dto.AmbientLight) (IdentifiableList[AmbientLightTemplate], error) {
	templates := make(IdentifiableList[AmbientLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := LoadAmbientLightTemplate(loader, assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

// UnloadAmbientLightTemplate unloads an ambient light template from the asset
// loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadAmbientLightTemplate(loader *AssetLoader, idLight Identifiable[AmbientLightTemplate]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadAmbientLightTemplates unloads a list of ambient light templates from
// the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadAmbientLightTemplates(loader *AssetLoader, idLights IdentifiableList[AmbientLightTemplate]) error {
	for _, idLight := range idLights {
		if err := UnloadAmbientLightTemplate(loader, idLight); err != nil {
			return err
		}
	}
	return nil
}

// InstantiateAmbientLightTemplate creates an ambient light in the scene based
// on the provided template.
//
// This operation needs to be called from the main thread.
func InstantiateAmbientLightTemplate(scene *Scene, template AmbientLightTemplate, nodes IdentifiableList[*hierarchy.Node], textures IdentifiableList[render.Texture]) *graphics.AmbientLight {
	node := nodes.GetByID(template.NodeID)
	reflectionTexture := textures.GetByID(template.ReflectionTextureID)
	refractionTexture := textures.GetByID(template.RefractionTextureID)
	info := AmbientLightInfo{
		ReflectionTexture: reflectionTexture,
		RefractionTexture: refractionTexture,
		OuterRadius:       opt.Unspecified[float64](),
		InnerRadius:       opt.Unspecified[float64](),
		CastShadow:        opt.V(template.CastShadow),
	}
	return scene.PlaceAmbientLight(node, info)
}

// PointLightTemplate represents a template for a point light in the scene.
type PointLightTemplate struct {
	NodeID       uint32
	EmitColor    dprec.Vec3
	EmitDistance float64
	CastShadow   bool
}

// LoadPointLightTemplate loads a point light template from the given asset
// data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPointLightTemplate(loader *AssetLoader, assetLight dto.PointLight) (Identifiable[PointLightTemplate], error) {
	return Identifiable[PointLightTemplate]{
		ID: assetLight.ID,
		Value: PointLightTemplate{
			NodeID:       assetLight.NodeID,
			EmitColor:    assetLight.EmitColor,
			EmitDistance: assetLight.EmitDistance,
			CastShadow:   assetLight.CastShadow,
		},
	}, nil
}

// LoadPointLightTemplates loads a list of point light templates from the given
// asset point lights.
//
// This is a blocking operation and should be called from a worker thread.
func LoadPointLightTemplates(loader *AssetLoader, assetLights []dto.PointLight) (IdentifiableList[PointLightTemplate], error) {
	templates := make(IdentifiableList[PointLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := LoadPointLightTemplate(loader, assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

// UnloadPointLightTemplate unloads a point light template from the asset
// loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPointLightTemplate(loader *AssetLoader, idLight Identifiable[PointLightTemplate]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadPointLightTemplates unloads a list of point light templates from the
// asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadPointLightTemplates(loader *AssetLoader, idLights IdentifiableList[PointLightTemplate]) error {
	for _, idLight := range idLights {
		if err := UnloadPointLightTemplate(loader, idLight); err != nil {
			return err
		}
	}
	return nil
}

// InstantiatePointLightTemplate creates a point light in the scene based on
// the provided template.
//
// This operation needs to be called from the main thread.
func InstantiatePointLightTemplate(scene *Scene, template PointLightTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.PointLight {
	node := nodes.GetByID(template.NodeID)
	info := PointLightInfo{
		EmitColor:    opt.V(template.EmitColor),
		EmitDistance: opt.V(template.EmitDistance),
		CastShadow:   opt.V(template.CastShadow),
	}
	return scene.PlacePointLight(node, info)
}

// SpotLightTemplate represents a template for a spot light in the scene.
type SpotLightTemplate struct {
	NodeID         uint32
	EmitColor      dprec.Vec3
	EmitDistance   float64
	EmitAngleOuter dprec.Angle
	EmitAngleInner dprec.Angle
	CastShadow     bool
}

// LoadSpotLightTemplate loads a spot light template from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadSpotLightTemplate(loader *AssetLoader, assetLight dto.SpotLight) (Identifiable[SpotLightTemplate], error) {
	return Identifiable[SpotLightTemplate]{
		ID: assetLight.ID,
		Value: SpotLightTemplate{
			NodeID:         assetLight.NodeID,
			EmitColor:      assetLight.EmitColor,
			EmitDistance:   assetLight.EmitDistance,
			EmitAngleOuter: assetLight.EmitAngleOuter,
			EmitAngleInner: assetLight.EmitAngleInner,
			CastShadow:     assetLight.CastShadow,
		},
	}, nil
}

// LoadSpotLightTemplates loads a list of spot light templates from the given
// asset spot lights.
//
// This is a blocking operation and should be called from a worker thread.
func LoadSpotLightTemplates(loader *AssetLoader, assetLights []dto.SpotLight) (IdentifiableList[SpotLightTemplate], error) {
	templates := make(IdentifiableList[SpotLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := LoadSpotLightTemplate(loader, assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

// UnloadSpotLightTemplate unloads a spot light template from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadSpotLightTemplate(loader *AssetLoader, idLight Identifiable[SpotLightTemplate]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadSpotLightTemplates unloads a list of spot light templates from the
// asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadSpotLightTemplates(loader *AssetLoader, idLights IdentifiableList[SpotLightTemplate]) error {
	for _, idLight := range idLights {
		if err := UnloadSpotLightTemplate(loader, idLight); err != nil {
			return err
		}
	}
	return nil
}

// InstantiateSpotLightTemplate creates a spot light in the scene based on the
// provided template.
//
// This operation needs to be called from the main thread.
func InstantiateSpotLightTemplate(scene *Scene, template SpotLightTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.SpotLight {
	node := nodes.GetByID(template.NodeID)
	info := SpotLightInfo{
		EmitColor:          opt.V(template.EmitColor),
		EmitDistance:       opt.V(template.EmitDistance),
		EmitOuterConeAngle: opt.V(template.EmitAngleOuter),
		EmitInnerConeAngle: opt.V(template.EmitAngleInner),
		CastShadow:         opt.V(template.CastShadow),
	}
	return scene.PlaceSpotLight(node, info)
}

// DirectionalLightTemplate represents a template for a directional light in
// the scene.
type DirectionalLightTemplate struct {
	NodeID     uint32
	EmitColor  dprec.Vec3
	CastShadow bool
}

// LoadDirectionalLightTemplate loads a directional light template from the
// given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadDirectionalLightTemplate(loader *AssetLoader, assetLight dto.DirectionalLight) (Identifiable[DirectionalLightTemplate], error) {
	return Identifiable[DirectionalLightTemplate]{
		ID: assetLight.ID,
		Value: DirectionalLightTemplate{
			NodeID:     assetLight.NodeID,
			EmitColor:  assetLight.EmitColor,
			CastShadow: assetLight.CastShadow,
		},
	}, nil
}

// LoadDirectionalLightTemplates loads a list of directional light templates
// from the given asset directional lights.
//
// This is a blocking operation and should be called from a worker thread.
func LoadDirectionalLightTemplates(loader *AssetLoader, assetLights []dto.DirectionalLight) (IdentifiableList[DirectionalLightTemplate], error) {
	templates := make(IdentifiableList[DirectionalLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := LoadDirectionalLightTemplate(loader, assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

// UnloadDirectionalLightTemplate unloads a directional light template from the
// asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadDirectionalLightTemplate(loader *AssetLoader, idLight Identifiable[DirectionalLightTemplate]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadDirectionalLightTemplates unloads a list of directional light
// templates from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadDirectionalLightTemplates(loader *AssetLoader, idLights IdentifiableList[DirectionalLightTemplate]) error {
	for _, idLight := range idLights {
		if err := UnloadDirectionalLightTemplate(loader, idLight); err != nil {
			return err
		}
	}
	return nil
}

// InstantiateDirectionalLightTemplate creates a directional light in the scene
// based on the provided template.
//
// This operation needs to be called from the main thread.
func InstantiateDirectionalLightTemplate(scene *Scene, template DirectionalLightTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.DirectionalLight {
	node := nodes.GetByID(template.NodeID)
	info := DirectionalLightInfo{
		EmitColor:  opt.V(template.EmitColor),
		CastShadow: opt.V(template.CastShadow),
	}
	return scene.PlaceDirectionalLight(node, info)
}
