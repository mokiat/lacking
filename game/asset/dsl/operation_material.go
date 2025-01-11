package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// Clear creates an operation that clears all data from the target.
func Clear() Operation {
	type clearable interface {
		Clear()
	}
	return FuncOperation(
		// apply function
		func(target any) error {
			clearable, ok := target.(clearable)
			if !ok {
				return fmt.Errorf("target %T is not a clearable", target)
			}
			clearable.Clear()
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("clear")
		},
	)
}

// AddGeometryPass creates an operation that adds a new geometry pass
// to the target.
func AddGeometryPass(passProvider Provider[*mdl.MaterialPass]) Operation {
	type passContainer interface {
		AddGeometryPass(*mdl.MaterialPass)
	}
	return FuncOperation(
		// apply function
		func(target any) error {
			pass, err := passProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting pass: %w", err)
			}
			container, ok := target.(passContainer)
			if !ok {
				return fmt.Errorf("target %T is not a geometry pass container", target)
			}
			container.AddGeometryPass(pass)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("add-geometry-pass", passProvider)
		},
	)
}

// AddForwardPass creates an operation that adds a new forward pass
// to the target.
func AddForwardPass(passProvider Provider[*mdl.MaterialPass]) Operation {
	type passContainer interface {
		AddForwardPass(*mdl.MaterialPass)
	}
	return FuncOperation(
		// apply function
		func(target any) error {
			pass, err := passProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting pass: %w", err)
			}
			container, ok := target.(passContainer)
			if !ok {
				return fmt.Errorf("target %T is not a forward pass container", target)
			}
			container.AddForwardPass(pass)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("add-forward-pass", passProvider)
		},
	)
}

// SetShader creates an operation that sets the shader of the target.
func SetShader(shaderProvider Provider[*mdl.Shader]) Operation {
	type shaderHolder interface {
		SetShader(*mdl.Shader)
	}
	return FuncOperation(
		// apply function
		func(target any) error {
			shader, err := shaderProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting shader: %w", err)
			}
			holder, ok := target.(shaderHolder)
			if !ok {
				return fmt.Errorf("target %T is not a shader holder", target)
			}
			holder.SetShader(shader)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-shader", shaderProvider)
		},
	)
}

// SetCulling creates an operation that sets the culling mode of the target.
func SetCulling(modeProvider Provider[mdl.CullMode]) Operation {
	type cullingSetter interface {
		SetCulling(mdl.CullMode)
	}
	return FuncOperation(
		// apply function
		func(target any) error {
			culling, err := modeProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting culling mode: %w", err)
			}
			setter, ok := target.(cullingSetter)
			if !ok {
				return fmt.Errorf("target %T is not a culling setter", target)
			}
			setter.SetCulling(culling)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-culling", modeProvider)
		},
	)
}
