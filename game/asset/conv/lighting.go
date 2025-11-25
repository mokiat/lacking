package conv

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
)

type LightingSource interface {
	AllAmbientLightPlacements() []mdl.Placed[*mdl.AmbientLight]
	AllPointLightPlacements() []mdl.Placed[*mdl.PointLight]
	AllSpotLightPlacements() []mdl.Placed[*mdl.SpotLight]
	AllDirectionalLightPlacements() []mdl.Placed[*mdl.DirectionalLight]
}

func NewLightingConverter() *LightingConverter {
	return &LightingConverter{}
}

type LightingConverter struct{}

func (c *LightingConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(LightingSource)
	if !ok {
		return nil
	}
	chunk, err := c.CreateLightingChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(dto.LightingChunkID, chunk))
	return nil
}

func (c *LightingConverter) CreateLightingChunk(src LightingSource) (*dto.LightingChunk, error) {
	allAmbientLightPlacements := src.AllAmbientLightPlacements()
	dtoAmbientLights := make([]dto.AmbientLight, len(allAmbientLightPlacements))
	for i, placement := range allAmbientLightPlacements {
		dtoAmbientLights[i] = c.convertAmbientLight(placement.Node, placement.Value)
	}

	allPointLightPlacements := src.AllPointLightPlacements()
	dtoPointLights := make([]dto.PointLight, len(allPointLightPlacements))
	for i, placement := range allPointLightPlacements {
		dtoPointLights[i] = c.convertPointLight(placement.Node, placement.Value)
	}

	allSpotLightPlacements := src.AllSpotLightPlacements()
	dtoSpotLights := make([]dto.SpotLight, len(allSpotLightPlacements))
	for i, placement := range allSpotLightPlacements {
		dtoSpotLights[i] = c.convertSpotLight(placement.Node, placement.Value)
	}

	allDirectionalLightPlacements := src.AllDirectionalLightPlacements()
	dtoDirectionalLights := make([]dto.DirectionalLight, len(allDirectionalLightPlacements))
	for i, placement := range allDirectionalLightPlacements {
		dtoDirectionalLights[i] = c.convertDirectionalLight(placement.Node, placement.Value)
	}

	return &dto.LightingChunk{
		AmbientLights:     dtoAmbientLights,
		PointLights:       dtoPointLights,
		SpotLights:        dtoSpotLights,
		DirectionalLights: dtoDirectionalLights,
	}, nil
}

func (c *LightingConverter) convertAmbientLight(node *mdl.Node, light *mdl.AmbientLight) dto.AmbientLight {
	return dto.AmbientLight{
		ID:                  light.ID(),
		NodeID:              node.ID(),
		ReflectionTextureID: light.ReflectionTexture().ID(),
		RefractionTextureID: light.RefractionTexture().ID(),
		CastShadow:          light.CastShadow(),
	}
}

func (c *LightingConverter) convertPointLight(node *mdl.Node, light *mdl.PointLight) dto.PointLight {
	return dto.PointLight{
		ID:           light.ID(),
		NodeID:       node.ID(),
		EmitColor:    light.EmitColor(),
		EmitDistance: light.EmitDistance(),
		CastShadow:   light.CastShadow(),
	}
}

func (c *LightingConverter) convertSpotLight(node *mdl.Node, light *mdl.SpotLight) dto.SpotLight {
	return dto.SpotLight{
		ID:             light.ID(),
		NodeID:         node.ID(),
		EmitColor:      light.EmitColor(),
		EmitDistance:   light.EmitDistance(),
		EmitAngleOuter: light.EmitAngleOuter(),
		EmitAngleInner: light.EmitAngleInner(),
		CastShadow:     light.CastShadow(),
	}
}

func (c *LightingConverter) convertDirectionalLight(node *mdl.Node, light *mdl.DirectionalLight) dto.DirectionalLight {
	return dto.DirectionalLight{
		ID:         light.ID(),
		NodeID:     node.ID(),
		EmitColor:  light.EmitColor(),
		CastShadow: light.CastShadow(),
	}
}
