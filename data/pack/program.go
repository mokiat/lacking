package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
)

type ProgramAssetBuilder struct {
	Asset
	vertexShaderProvider   ShaderResourceProvider
	fragmentShaderProvider ShaderResourceProvider
}

func (b *ProgramAssetBuilder) WithVertexShader(shaderProvider ShaderResourceProvider) *ProgramAssetBuilder {
	b.vertexShaderProvider = shaderProvider
	return b
}

func (b *ProgramAssetBuilder) WithFragmentShader(shaderProvider ShaderResourceProvider) *ProgramAssetBuilder {
	b.fragmentShaderProvider = shaderProvider
	return b
}

func (b *ProgramAssetBuilder) Build() error {
	vertexShader, err := b.vertexShaderProvider.Shader()
	if err != nil {
		return fmt.Errorf("failed to get vertex shader: %w", err)
	}

	fragmentShader, err := b.fragmentShaderProvider.Shader()
	if err != nil {
		return fmt.Errorf("failed to get fragment shader: %w", err)
	}

	program := &asset.Program{
		VertexSourceCode:   vertexShader,
		FragmentSourceCode: fragmentShader,
	}

	file, err := b.CreateFile()
	if err != nil {
		return err
	}
	defer file.Close()

	if err := asset.EncodeProgram(file, program); err != nil {
		return fmt.Errorf("failed to encode program: %w", err)
	}
	return nil
}
