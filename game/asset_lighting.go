package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/render"
)

type AmbientLightTemplate struct {
	NodeID              uint32
	ReflectionTextureID uint32
	RefractionTextureID uint32
	CastShadow          bool
}

func (l *AssetLoader) ResolveAmbientLightTemplate(assetLight dto.AmbientLight) (Identifiable[AmbientLightTemplate], error) {
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

func (l *AssetLoader) ResolveAmbientLightTemplates(assetLights []dto.AmbientLight) (IdentifiableList[AmbientLightTemplate], error) {
	templates := make(IdentifiableList[AmbientLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := l.ResolveAmbientLightTemplate(assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

func (s *Scene) InstantiateAmbientLightTemplate(template AmbientLightTemplate, nodes IdentifiableList[*hierarchy.Node], textures IdentifiableList[render.Texture]) *graphics.AmbientLight {
	node, ok := nodes.FindByID(template.NodeID)
	if !ok {
		return nil
	}
	reflectionTexture, ok := textures.FindByID(template.ReflectionTextureID)
	if !ok {
		return nil
	}
	refractionTexture, ok := textures.FindByID(template.RefractionTextureID)
	if !ok {
		return nil
	}
	info := AmbientLightInfo{
		ReflectionTexture: reflectionTexture,
		RefractionTexture: refractionTexture,
		OuterRadius:       opt.Unspecified[float64](),
		InnerRadius:       opt.Unspecified[float64](),
		CastShadow:        opt.V(template.CastShadow),
	}
	return s.PlaceAmbientLight(node, info)
}

type PointLightTemplate struct {
	NodeID       uint32
	EmitColor    dprec.Vec3
	EmitDistance float64
	CastShadow   bool
}

func (l *AssetLoader) ResolvePointLightTemplate(assetLight dto.PointLight) (Identifiable[PointLightTemplate], error) {
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

func (l *AssetLoader) ResolvePointLightTemplates(assetLights []dto.PointLight) (IdentifiableList[PointLightTemplate], error) {
	templates := make(IdentifiableList[PointLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := l.ResolvePointLightTemplate(assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

func (s *Scene) InstantiatePointLightTemplate(template PointLightTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.PointLight {
	node, ok := nodes.FindByID(template.NodeID)
	if !ok {
		return nil
	}
	info := PointLightInfo{
		EmitColor:    opt.V(template.EmitColor),
		EmitDistance: opt.V(template.EmitDistance),
		CastShadow:   opt.V(template.CastShadow),
	}
	return s.PlacePointLight(node, info)
}

type SpotLightTemplate struct {
	NodeID         uint32
	EmitColor      dprec.Vec3
	EmitDistance   float64
	EmitAngleOuter dprec.Angle
	EmitAngleInner dprec.Angle
	CastShadow     bool
}

func (l *AssetLoader) ResolveSpotLightTemplate(assetLight dto.SpotLight) (Identifiable[SpotLightTemplate], error) {
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

func (l *AssetLoader) ResolveSpotLightTemplates(assetLights []dto.SpotLight) (IdentifiableList[SpotLightTemplate], error) {
	templates := make(IdentifiableList[SpotLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := l.ResolveSpotLightTemplate(assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

func (s *Scene) InstantiateSpotLightTemplate(template SpotLightTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.SpotLight {
	node, ok := nodes.FindByID(template.NodeID)
	if !ok {
		return nil
	}
	info := SpotLightInfo{
		EmitColor:          opt.V(template.EmitColor),
		EmitDistance:       opt.V(template.EmitDistance),
		EmitOuterConeAngle: opt.V(template.EmitAngleOuter),
		EmitInnerConeAngle: opt.V(template.EmitAngleInner),
		CastShadow:         opt.V(template.CastShadow),
	}
	return s.PlaceSpotLight(node, info)
}

type DirectionalLightTemplate struct {
	NodeID     uint32
	EmitColor  dprec.Vec3
	CastShadow bool
}

func (l *AssetLoader) ResolveDirectionalLightTemplate(assetLight dto.DirectionalLight) (Identifiable[DirectionalLightTemplate], error) {
	return Identifiable[DirectionalLightTemplate]{
		ID: assetLight.ID,
		Value: DirectionalLightTemplate{
			NodeID:     assetLight.NodeID,
			EmitColor:  assetLight.EmitColor,
			CastShadow: assetLight.CastShadow,
		},
	}, nil
}

func (l *AssetLoader) ResolveDirectionalLightTemplates(assetLights []dto.DirectionalLight) (IdentifiableList[DirectionalLightTemplate], error) {
	templates := make(IdentifiableList[DirectionalLightTemplate], len(assetLights))
	for i, assetLight := range assetLights {
		template, err := l.ResolveDirectionalLightTemplate(assetLight)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

func (s *Scene) InstantiateDirectionalLightTemplate(template DirectionalLightTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.DirectionalLight {
	node, ok := nodes.FindByID(template.NodeID)
	if !ok {
		return nil
	}
	info := DirectionalLightInfo{
		EmitColor:  opt.V(template.EmitColor),
		CastShadow: opt.V(template.CastShadow),
	}
	return s.PlaceDirectionalLight(node, info)
}
