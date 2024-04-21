package game

import (
	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
)

func (s *ResourceSet) convertAmbientLight(textures []render.Texture, assetLight asset.AmbientLight) ambientLightInstance {
	return ambientLightInstance{
		nodeIndex:         int(assetLight.NodeIndex),
		reflectionTexture: textures[assetLight.ReflectionTextureIndex],
		refractionTexture: textures[assetLight.RefractionTextureIndex],
		castShadow:        assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertPointLight(assetLight asset.PointLight) pointLightInstance {
	return pointLightInstance{
		nodeIndex:    int(assetLight.NodeIndex),
		emitColor:    assetLight.EmitColor,
		emitDistance: assetLight.EmitDistance,
		castShadow:   assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertSpotLight(assetLight asset.SpotLight) spotLightInstance {
	return spotLightInstance{
		nodeIndex:      int(assetLight.NodeIndex),
		emitColor:      assetLight.EmitColor,
		emitDistance:   assetLight.EmitDistance,
		emitAngleOuter: assetLight.EmitAngleOuter,
		emitAngleInner: assetLight.EmitAngleInner,
		castShadow:     assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertDirectionalLight(assetLight asset.DirectionalLight) directionalLightInstance {
	return directionalLightInstance{
		nodeIndex:  int(assetLight.NodeIndex),
		emitColor:  assetLight.EmitColor,
		castShadow: assetLight.CastShadow,
	}
}
