package pack

import (
	"fmt"
	"hash"
	"sync"
)

type ProgramProvider interface {
	Program(ctx *Context) (*Program, error)
	Digest(hasher hash.Hash) error
}

type Program struct {
	VertexShader   *Shader
	FragmentShader *Shader
}

func BuildProgram() *BuildProgramAction {
	return &BuildProgramAction{}
}

var _ ProgramProvider = (*BuildProgramAction)(nil)

type BuildProgramAction struct {
	vertexShaderProvider   ShaderProvider
	fragmentShaderProvider ShaderProvider

	resultMutex  sync.Mutex
	resultDigest []byte
	result       *Program
}

func (a *BuildProgramAction) WithVertexShader(shaderProvider ShaderProvider) *BuildProgramAction {
	a.vertexShaderProvider = shaderProvider
	return a
}

func (a *BuildProgramAction) WithFragmentShader(shaderProvider ShaderProvider) *BuildProgramAction {
	a.fragmentShaderProvider = shaderProvider
	return a
}

func (a *BuildProgramAction) Describe() string {
	return "build_program()"
}

func (a *BuildProgramAction) Digest(hasher hash.Hash) error {
	params := HashableParams{}
	if a.vertexShaderProvider != nil {
		params["vertex_shader"] = a.vertexShaderProvider
	}
	if a.fragmentShaderProvider != nil {
		params["fragment_shader"] = a.fragmentShaderProvider
	}
	return WriteCompositeDigest(hasher, "build_program", params)
}

func (a *BuildProgramAction) Program(ctx *Context) (*Program, error) {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	a.resultMutex.Lock()
	defer a.resultMutex.Unlock()

	digest, err := CalculateDigest(a)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}
	if EqualDigests(digest, a.resultDigest) {
		return a.result, nil
	}

	result, err := a.run(ctx)
	if err != nil {
		return nil, err
	}

	a.result = result
	a.resultDigest = digest
	return result, nil
}

func (a *BuildProgramAction) run(ctx *Context) (*Program, error) {
	vertexShader, err := a.vertexShaderProvider.Shader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vertex shader: %w", err)
	}
	fragmentShader, err := a.fragmentShaderProvider.Shader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get fragment shader: %w", err)
	}
	return &Program{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
	}, nil
}
