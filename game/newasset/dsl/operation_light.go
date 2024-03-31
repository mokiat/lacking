package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/model"
)

func SetEmitColor(color dprec.Vec3) Operation {
	apply := func(target any) error {
		emitter, ok := target.(model.ColorEmitter)
		if !ok {
			return fmt.Errorf("target %T is not a color emitter", target)
		}
		emitter.SetEmitColor(color)
		return nil
	}

	digest := func() ([]byte, error) {
		return digestItems("set-emit-color", color)
	}

	return FuncOperation(apply, digest)
}

func SetEmitDistance(distance float64) Operation {
	apply := func(target any) error {
		emitter, ok := target.(model.DistanceEmitter)
		if !ok {
			return fmt.Errorf("target %T is not a distance emitter", target)
		}
		emitter.SetEmitDistance(distance)
		return nil
	}

	digest := func() ([]byte, error) {
		return digestItems("set-emit-distance", distance)
	}

	return FuncOperation(apply, digest)
}
