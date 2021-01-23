package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/mokiat/lacking/opengl"
)

type Program struct {
	Program *opengl.Program

	FBColor0TextureLocation int32
	FBColor1TextureLocation int32
	FBDepthTextureLocation  int32

	ProjectionMatrixLocation int32
	ModelMatrixLocation      int32
	ViewMatrixLocation       int32
	CameraMatrixLocation     int32
	LightDirectionWSLocation int32
	ExposureLocation         int32

	MetalnessLocation                int32
	MetalnessTwoDTextureLocation     int32
	RoughnessLocation                int32
	RoughnessTwoDTextureLocation     int32
	AlbedoColorLocation              int32
	AlbedoTwoDTextureLocation        int32
	AlbedoCubeTextureLocation        int32
	AmbientReflectionTextureLocation int32
	AmbientRefractionTextureLocation int32
	NormalScaleLocation              int32
	NormalTwoDTextureLocation        int32
}

func (p *Program) ID() uint32 {
	return p.Program.ID()
}

type ProgramData struct {
	VertexShaderSourceCode   string
	FragmentShaderSourceCode string
}

func (p *Program) Allocate(data ProgramData) error {
	vertexShader := opengl.NewShader()
	vertexShaderInfo := opengl.ShaderAllocateInfo{
		ShaderType: gl.VERTEX_SHADER,
		SourceCode: data.VertexShaderSourceCode,
	}
	if err := vertexShader.Allocate(vertexShaderInfo); err != nil {
		return fmt.Errorf("failed to allocate vertex shader: %w", err)
	}

	fragmentShader := opengl.NewShader()
	fragmentShaderInfo := opengl.ShaderAllocateInfo{
		ShaderType: gl.FRAGMENT_SHADER,
		SourceCode: data.FragmentShaderSourceCode,
	}
	if err := fragmentShader.Allocate(fragmentShaderInfo); err != nil {
		return fmt.Errorf("failed to allocate fragment shader: %w", err)
	}

	programInfo := opengl.ProgramAllocateInfo{
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
	}

	p.Program = opengl.NewProgram()
	if err := p.Program.Allocate(programInfo); err != nil {
		return fmt.Errorf("failed to allocate program: %w", err)
	}

	if err := fragmentShader.Release(); err != nil {
		return fmt.Errorf("failed to release fragment shader: %w", err)
	}
	if err := vertexShader.Release(); err != nil {
		return fmt.Errorf("failed to release vertex shader: %w", err)
	}

	p.FBColor0TextureLocation = p.Program.UniformLocation("fbColor0TextureIn")
	p.FBColor1TextureLocation = p.Program.UniformLocation("fbColor1TextureIn")
	p.FBDepthTextureLocation = p.Program.UniformLocation("fbDepthTextureIn")

	p.ProjectionMatrixLocation = p.Program.UniformLocation("projectionMatrixIn")
	p.ModelMatrixLocation = p.Program.UniformLocation("modelMatrixIn")
	p.ViewMatrixLocation = p.Program.UniformLocation("viewMatrixIn")
	p.CameraMatrixLocation = p.Program.UniformLocation("cameraMatrixIn")
	p.LightDirectionWSLocation = p.Program.UniformLocation("lightDirectionWSIn")
	p.ExposureLocation = p.Program.UniformLocation("exposureIn")

	p.MetalnessLocation = p.Program.UniformLocation("metalnessIn")
	p.MetalnessTwoDTextureLocation = p.Program.UniformLocation("metalnessTwoDTextureIn")
	p.RoughnessLocation = p.Program.UniformLocation("roughnessIn")
	p.RoughnessTwoDTextureLocation = p.Program.UniformLocation("roughnessTwoDTextureIn")
	p.AlbedoColorLocation = p.Program.UniformLocation("albedoColorIn")
	p.AlbedoTwoDTextureLocation = p.Program.UniformLocation("albedoTwoDTextureIn")
	p.AlbedoCubeTextureLocation = p.Program.UniformLocation("albedoCubeTextureIn")
	p.AmbientReflectionTextureLocation = p.Program.UniformLocation("ambientReflectionTextureIn")
	p.AmbientRefractionTextureLocation = p.Program.UniformLocation("ambientRefractionTextureIn")
	p.NormalScaleLocation = p.Program.UniformLocation("normalScaleIn")
	p.NormalTwoDTextureLocation = p.Program.UniformLocation("normalTwoDTextureIn")
	return nil
}

func (p *Program) Release() error {
	return p.Program.Release()
}
