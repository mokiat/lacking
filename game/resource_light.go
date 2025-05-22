package game

import (
	"github.com/mokiat/lacking/game/asset/lightingdto"
	"github.com/mokiat/lacking/render"
)

func (s *ResourceSet) convertAmbientLight(textures []render.Texture, assetLight lightingdto.AmbientLight) ambientLightInstance {
	return ambientLightInstance{
		nodeIndex:         int(assetLight.NodeIndex),
		reflectionTexture: textures[assetLight.ReflectionTextureIndex],
		refractionTexture: textures[assetLight.RefractionTextureIndex],
		castShadow:        assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertPointLight(assetLight lightingdto.PointLight) pointLightInstance {
	return pointLightInstance{
		nodeIndex:    int(assetLight.NodeIndex),
		emitColor:    assetLight.EmitColor,
		emitDistance: assetLight.EmitDistance,
		castShadow:   assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertSpotLight(assetLight lightingdto.SpotLight) spotLightInstance {
	return spotLightInstance{
		nodeIndex:      int(assetLight.NodeIndex),
		emitColor:      assetLight.EmitColor,
		emitDistance:   assetLight.EmitDistance,
		emitAngleOuter: assetLight.EmitAngleOuter,
		emitAngleInner: assetLight.EmitAngleInner,
		castShadow:     assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertDirectionalLight(assetLight lightingdto.DirectionalLight) directionalLightInstance {
	return directionalLightInstance{
		nodeIndex:  int(assetLight.NodeIndex),
		emitColor:  assetLight.EmitColor,
		castShadow: assetLight.CastShadow,
	}
}
