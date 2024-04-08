package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

func CreateAmbientLight(name string, reflectionTextureProvider, refractionTextureProvider Provider[*mdl.Texture], operations ...Operation) Provider[mdl.Node] {
	get := func() (mdl.Node, error) {
		reflectionTexture, err := reflectionTextureProvider.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get reflection texture: %w", err)
		}

		refractionTexture, err := refractionTextureProvider.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get refraction texture: %w", err)
		}

		var light mdl.AmbientLight
		light.SetName(name)
		light.SetTranslation(dprec.ZeroVec3())
		light.SetRotation(dprec.IdentityQuat())
		light.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
		light.SetReflectionTexture(reflectionTexture)
		light.SetRefractionTexture(refractionTexture)
		light.SetCastShadow(false)
		for _, operation := range operations {
			if err := operation.Apply(&light); err != nil {
				return nil, err
			}
		}
		return &light, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("point-light", name, operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}

func CreatePointLight(name string, operations ...Operation) Provider[mdl.Node] {
	get := func() (mdl.Node, error) {
		var light mdl.PointLight
		light.SetName(name)
		light.SetTranslation(dprec.ZeroVec3())
		light.SetRotation(dprec.IdentityQuat())
		light.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
		light.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
		light.SetEmitDistance(10.0)
		light.SetCastShadow(false)
		for _, operation := range operations {
			if err := operation.Apply(&light); err != nil {
				return nil, err
			}
		}
		return &light, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("point-light", name, operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}

func CreateSpotLight(name string, operations ...Operation) Provider[mdl.Node] {
	get := func() (mdl.Node, error) {
		var light mdl.SpotLight
		light.SetName(name)
		light.SetTranslation(dprec.ZeroVec3())
		light.SetRotation(dprec.IdentityQuat())
		light.SetScale(dprec.NewVec3(1.0, 1.0, 1.0))
		light.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
		light.SetEmitDistance(10.0)
		light.SetEmitAngleOuter(dprec.Degrees(90.0))
		light.SetEmitAngleInner(dprec.Degrees(60.0))
		light.SetCastShadow(false)
		for _, operation := range operations {
			if err := operation.Apply(&light); err != nil {
				return nil, err
			}
		}
		return &light, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("spot-light", name, operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
