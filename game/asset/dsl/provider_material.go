package dsl

import "github.com/mokiat/lacking/game/asset/mdl"

func CreateMaterial(name string, opts ...Operation) Provider[*mdl.Material] {
	panic("TODO")
}

// CreateMaterialPass creates a provider that will create a material pass
// with the specified options.
func CreateMaterialPass(opts ...Operation) Provider[*mdl.MaterialPass] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.MaterialPass, error) {
			pass := mdl.NewMaterialPass()
			for _, opt := range opts {
				if err := opt.Apply(pass); err != nil {
					return nil, err
				}
			}
			return pass, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-material-pass", opts)
		},
	))
}

// CreateShader creates a provider that will create a shader with the
// specified type and source code.
func CreateShader(shaderType mdl.ShaderType, code string) Provider[*mdl.Shader] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Shader, error) {
			shader := mdl.NewShader(shaderType)
			shader.SetSourceCode(code)
			return shader, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-shader", uint8(shaderType), code)
		},
	))
}
