package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
)

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
