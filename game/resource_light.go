package game

import (
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/render"
)

func (s *ResourceSet) convertAmbientLight(nodes map[uint32]int, textures map[uint32]render.Texture, assetLight dto.AmbientLight) ambientLightInstance {
	nodeIndex := nodes[assetLight.NodeID]
	return ambientLightInstance{
		nodeIndex:         nodeIndex,
		reflectionTexture: textures[assetLight.ReflectionTextureID],
		refractionTexture: textures[assetLight.RefractionTextureID],
		castShadow:        assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertPointLight(nodes map[uint32]int, assetLight dto.PointLight) pointLightInstance {
	nodeIndex := nodes[assetLight.NodeID]
	return pointLightInstance{
		nodeIndex:    nodeIndex,
		emitColor:    assetLight.EmitColor,
		emitDistance: assetLight.EmitDistance,
		castShadow:   assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertSpotLight(nodes map[uint32]int, assetLight dto.SpotLight) spotLightInstance {
	nodeIndex := nodes[assetLight.NodeID]
	return spotLightInstance{
		nodeIndex:      nodeIndex,
		emitColor:      assetLight.EmitColor,
		emitDistance:   assetLight.EmitDistance,
		emitAngleOuter: assetLight.EmitAngleOuter,
		emitAngleInner: assetLight.EmitAngleInner,
		castShadow:     assetLight.CastShadow,
	}
}

func (s *ResourceSet) convertDirectionalLight(nodes map[uint32]int, assetLight dto.DirectionalLight) directionalLightInstance {
	nodeIndex := nodes[assetLight.NodeID]
	return directionalLightInstance{
		nodeIndex:  nodeIndex,
		emitColor:  assetLight.EmitColor,
		castShadow: assetLight.CastShadow,
	}
}
