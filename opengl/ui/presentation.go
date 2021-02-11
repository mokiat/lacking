package ui

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/opengl"
)

type presentation struct {
	program                  *opengl.Program
	projectionMatrixLocation int32
	textureLocation          int32
}

func newPresentation(vShader, fShader string) (*presentation, error) {
	vertexShader := opengl.NewShader()
	vertexShaderInfo := opengl.ShaderAllocateInfo{
		ShaderType: gl.VERTEX_SHADER,
		SourceCode: vShader,
	}
	if err := vertexShader.Allocate(vertexShaderInfo); err != nil {
		return nil, err
	}
	fragmentShader := opengl.NewShader()
	fragmentShaderInfo := opengl.ShaderAllocateInfo{
		ShaderType: gl.FRAGMENT_SHADER,
		SourceCode: fShader,
	}
	if err := fragmentShader.Allocate(fragmentShaderInfo); err != nil {
		return nil, err
	}

	programInfo := opengl.ProgramAllocateInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
	}
	program := opengl.NewProgram()
	if err := program.Allocate(programInfo); err != nil {
		return nil, err
	}

	if err := vertexShader.Release(); err != nil {
		return nil, err
	}
	if err := fragmentShader.Release(); err != nil {
		return nil, err
	}

	return &presentation{
		program:                  program,
		projectionMatrixLocation: program.UniformLocation("projectionMatrixIn"),
		textureLocation:          program.UniformLocation("textureIn"),
	}, nil
}

func releasePresentation(p *presentation) error {
	if err := p.program.Release(); err != nil {
		return err
	}
	return nil
}

func newSolidPresentation() (*presentation, error) {
	vertexShaderSource := opengl.NewShaderSourceBuilder(solidShapeVertexShaderTemplate)
	fragmentShaderSource := opengl.NewShaderSourceBuilder(solidShapeFragmentShaderTemplate)
	return newPresentation(vertexShaderSource.Build(), fragmentShaderSource.Build())
}

func newImagePresentation() (*presentation, error) {
	vertexShaderSource := opengl.NewShaderSourceBuilder(imageShapeVertexShaderTemplate)
	fragmentShaderSource := opengl.NewShaderSourceBuilder(imageShapeFragmentShaderTemplate)
	return newPresentation(vertexShaderSource.Build(), fragmentShaderSource.Build())
}
