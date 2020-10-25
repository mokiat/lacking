package pack

type ProgramProvider interface {
	Program() *Program
}

type Program struct {
	VertexShader   *Shader
	FragmentShader *Shader
}

type BuildProgramOption func(a *BuildProgramAction)

func WithVertexShader(shader ShaderProvider) BuildProgramOption {
	return func(a *BuildProgramAction) {
		a.vertexShaderProvider = shader
	}
}

func WithFragmentShader(shader ShaderProvider) BuildProgramOption {
	return func(a *BuildProgramAction) {
		a.fragmentShaderProvider = shader
	}
}

type BuildProgramAction struct {
	vertexShaderProvider   ShaderProvider
	fragmentShaderProvider ShaderProvider
	program                *Program
}

func (a *BuildProgramAction) Describe() string {
	return "build_program()"
}

func (a *BuildProgramAction) Program() *Program {
	if a.program == nil {
		panic("reading data from unprocessed action")
	}
	return a.program
}

func (a *BuildProgramAction) Run() error {
	a.program = &Program{
		VertexShader:   a.vertexShaderProvider.Shader(),
		FragmentShader: a.fragmentShaderProvider.Shader(),
	}
	return nil
}
