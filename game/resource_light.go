package game

import (
	"github.com/mokiat/lacking/game/asset/dto"
)

func (s *ResourceSet) convertAmbientLight(assetLight dto.AmbientLight) ambientLightInstance {
	return ambientLightInstance{
		nodeID:              assetLight.NodeID,
		reflectionTextureID: assetLight.ReflectionTextureID,
		refractionTextureID: assetLight.RefractionTextureID,
		castShadow:          assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertPointLight(assetLight dto.PointLight) pointLightInstance {
	return pointLightInstance{
		nodeID:       assetLight.NodeID,
		emitColor:    assetLight.EmitColor,
		emitDistance: assetLight.EmitDistance,
		castShadow:   assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertSpotLight(assetLight dto.SpotLight) spotLightInstance {
	return spotLightInstance{
		nodeID:         assetLight.NodeID,
		emitColor:      assetLight.EmitColor,
		emitDistance:   assetLight.EmitDistance,
		emitAngleOuter: assetLight.EmitAngleOuter,
		emitAngleInner: assetLight.EmitAngleInner,
		castShadow:     assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertDirectionalLight(assetLight dto.DirectionalLight) directionalLightInstance {
	return directionalLightInstance{
		nodeID:     assetLight.NodeID,
		emitColor:  assetLight.EmitColor,
		castShadow: assetLight.CastShadow,
	}
}
