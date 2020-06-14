package graphics

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Program struct {
	ID               uint32
	VertexShaderID   uint32
	FragmentShaderID uint32

	FBColor0TextureLocation int32
	FBColor1TextureLocation int32
	FBDepthTextureLocation  int32

	ProjectionMatrixLocation int32
	ModelMatrixLocation      int32
	ViewMatrixLocation       int32
	CameraMatrixLocation     int32

	MetalnessLocation            int32
	MetalnessTwoDTextureLocation int32
	RoughnessLocation            int32
	RoughnessTwoDTextureLocation int32
	AlbedoColorLocation          int32
	AlbedoTwoDTextureLocation    int32
	AlbedoCubeTextureLocation    int32
	NormalScaleLocation          int32
	NormalTwoDTextureLocation    int32
}

type ProgramData struct {
	VertexShaderSourceCode   string
	FragmentShaderSourceCode string
}

func (p *Program) Allocate(data ProgramData) error {
	p.ID = gl.CreateProgram()

	p.VertexShaderID = gl.CreateShader(gl.VERTEX_SHADER)
	setShaderSourceCode(p.VertexShaderID, data.VertexShaderSourceCode)
	gl.CompileShader(p.VertexShaderID)
	if getShaderCompileStatus(p.VertexShaderID) == gl.FALSE {
		log := getShaderLog(p.VertexShaderID)
		return fmt.Errorf("failed to compile vertex shader: %s", log)
	}
	gl.AttachShader(p.ID, p.VertexShaderID)

	p.FragmentShaderID = gl.CreateShader(gl.FRAGMENT_SHADER)
	setShaderSourceCode(p.FragmentShaderID, data.FragmentShaderSourceCode)
	gl.CompileShader(p.FragmentShaderID)
	if getShaderCompileStatus(p.FragmentShaderID) == gl.FALSE {
		log := getShaderLog(p.FragmentShaderID)
		return fmt.Errorf("failed to compile fragment shader: %s", log)
	}
	gl.AttachShader(p.ID, p.FragmentShaderID)

	gl.LinkProgram(p.ID)
	if getProgramLinkStatus(p.ID) == gl.FALSE {
		log := getProgramLog(p.ID)
		return fmt.Errorf("failed to link program: %s", log)
	}

	p.FBColor0TextureLocation = getUniformLocation(p.ID, "fbColor0TextureIn")
	p.FBColor1TextureLocation = getUniformLocation(p.ID, "fbColor1TextureIn")
	p.FBDepthTextureLocation = getUniformLocation(p.ID, "fbDepthTextureIn")

	p.ProjectionMatrixLocation = getUniformLocation(p.ID, "projectionMatrixIn")
	p.ModelMatrixLocation = getUniformLocation(p.ID, "modelMatrixIn")
	p.ViewMatrixLocation = getUniformLocation(p.ID, "viewMatrixIn")
	p.CameraMatrixLocation = getUniformLocation(p.ID, "cameraMatrixIn")

	p.MetalnessLocation = getUniformLocation(p.ID, "metalnessIn")
	p.MetalnessTwoDTextureLocation = getUniformLocation(p.ID, "metalnessTwoDTextureIn")
	p.RoughnessLocation = getUniformLocation(p.ID, "roughnessIn")
	p.RoughnessTwoDTextureLocation = getUniformLocation(p.ID, "roughnessTwoDTextureIn")
	p.AlbedoColorLocation = getUniformLocation(p.ID, "albedoColorIn")
	p.AlbedoTwoDTextureLocation = getUniformLocation(p.ID, "albedoTwoDTextureIn")
	p.AlbedoCubeTextureLocation = getUniformLocation(p.ID, "albedoCubeTextureIn")
	p.NormalScaleLocation = getUniformLocation(p.ID, "normalScaleIn")
	p.NormalTwoDTextureLocation = getUniformLocation(p.ID, "normalTwoDTextureIn")
	return nil
}

func (p *Program) Release() error {
	gl.DeleteProgram(p.ID)
	gl.DeleteShader(p.VertexShaderID)
	gl.DeleteShader(p.FragmentShaderID)
	p.ID = 0
	p.VertexShaderID = 0
	p.FragmentShaderID = 0
	return nil
}

func getUniformLocation(id uint32, name string) int32 {
	return gl.GetUniformLocation(id, gl.Str(name+"\x00"))
}

func setShaderSourceCode(id uint32, sourceCode string) {
	sources, free := gl.Strs(sourceCode + "\x00")
	defer free()
	gl.ShaderSource(id, 1, sources, nil)
}

func getShaderCompileStatus(shader uint32) int32 {
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	return status
}

func getShaderLog(shader uint32) string {
	var logLength int32
	gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
	return log
}

func getProgramLinkStatus(program uint32) int32 {
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	return status
}

func getProgramLog(program uint32) string {
	var logLength int32
	gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
	return log
}
