package dsl

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/model"
)

func CreatePointLight(name string, operations ...Operation) Provider[model.Node] {
	get := func() (model.Node, error) {
		var light model.PointLight
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

func CreateSpotLight(name string, operations ...Operation) Provider[model.Node] {
	get := func() (model.Node, error) {
		var light model.SpotLight
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
