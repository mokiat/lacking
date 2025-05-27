package lightingconv

import (
	"github.com/mokiat/lacking/game/asset/dto/lightingdto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

type Source interface {
	AllAmbientLightPlacements() []mdl.Placed[*mdl.AmbientLight]
	AllPointLightPlacements() []mdl.Placed[*mdl.PointLight]
	AllSpotLightPlacements() []mdl.Placed[*mdl.SpotLight]
	AllDirectionalLightPlacements() []mdl.Placed[*mdl.DirectionalLight]
}

func CreateLightingChunk(src Source) (*lightingdto.LightingChunk, error) {
	allAmbientLightPlacements := src.AllAmbientLightPlacements()
	dtoAmbientLights := make([]lightingdto.AmbientLight, len(allAmbientLightPlacements))
	for i, placement := range allAmbientLightPlacements {
		dtoAmbientLights[i] = convertAmbientLight(placement.Node, placement.Value)
	}

	allPointLightPlacements := src.AllPointLightPlacements()
	dtoPointLights := make([]lightingdto.PointLight, len(allPointLightPlacements))
	for i, placement := range allPointLightPlacements {
		dtoPointLights[i] = convertPointLight(placement.Node, placement.Value)
	}

	allSpotLightPlacements := src.AllSpotLightPlacements()
	dtoSpotLights := make([]lightingdto.SpotLight, len(allSpotLightPlacements))
	for i, placement := range allSpotLightPlacements {
		dtoSpotLights[i] = convertSpotLight(placement.Node, placement.Value)
	}

	allDirectionalLightPlacements := src.AllDirectionalLightPlacements()
	dtoDirectionalLights := make([]lightingdto.DirectionalLight, len(allDirectionalLightPlacements))
	for i, placement := range allDirectionalLightPlacements {
		dtoDirectionalLights[i] = convertDirectionalLight(placement.Node, placement.Value)
	}

	return &lightingdto.LightingChunk{
		AmbientLights:     dtoAmbientLights,
		PointLights:       dtoPointLights,
		SpotLights:        dtoSpotLights,
		DirectionalLights: dtoDirectionalLights,
	}, nil
}

func convertAmbientLight(node *mdl.Node, light *mdl.AmbientLight) lightingdto.AmbientLight {
	return lightingdto.AmbientLight{
		ID:                  light.ID(),
		NodeID:              node.ID(),
		ReflectionTextureID: light.ReflectionTexture().ID(),
		RefractionTextureID: light.RefractionTexture().ID(),
		CastShadow:          light.CastShadow(),
	}
}

func convertPointLight(node *mdl.Node, light *mdl.PointLight) lightingdto.PointLight {
	return lightingdto.PointLight{
		ID:           light.ID(),
		NodeID:       node.ID(),
		EmitColor:    light.EmitColor(),
		EmitDistance: light.EmitDistance(),
		CastShadow:   light.CastShadow(),
	}
}

func convertSpotLight(node *mdl.Node, light *mdl.SpotLight) lightingdto.SpotLight {
	return lightingdto.SpotLight{
		ID:             light.ID(),
		NodeID:         node.ID(),
		EmitColor:      light.EmitColor(),
		EmitDistance:   light.EmitDistance(),
		EmitAngleOuter: light.EmitAngleOuter(),
		EmitAngleInner: light.EmitAngleInner(),
		CastShadow:     light.CastShadow(),
	}
}

func convertDirectionalLight(node *mdl.Node, light *mdl.DirectionalLight) lightingdto.DirectionalLight {
	return lightingdto.DirectionalLight{
		ID:         light.ID(),
		NodeID:     node.ID(),
		EmitColor:  light.EmitColor(),
		CastShadow: light.CastShadow(),
	}
}
