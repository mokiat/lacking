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

	FBAlbedoTextureLocation int32
	FBNormalTextureLocation int32
	FBDepthTextureLocation  int32

	ProjectionMatrixLocation  int32
	ModelMatrixLocation       int32
	ViewMatrixLocation        int32
	InverseViewMatrixLocation int32

	DiffuseColorLocation   int32
	DiffuseTextureLocation int32
	SkyboxTextureLocation  int32
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

	p.FBAlbedoTextureLocation = gl.GetUniformLocation(p.ID, gl.Str("fbAlbedoTextureIn"+"\x00"))
	p.FBNormalTextureLocation = gl.GetUniformLocation(p.ID, gl.Str("fbNormalTextureIn"+"\x00"))
	p.FBDepthTextureLocation = gl.GetUniformLocation(p.ID, gl.Str("fbDepthTextureIn"+"\x00"))

	p.ProjectionMatrixLocation = gl.GetUniformLocation(p.ID, gl.Str("projectionMatrixIn"+"\x00"))
	p.ModelMatrixLocation = gl.GetUniformLocation(p.ID, gl.Str("modelMatrixIn"+"\x00"))
	p.ViewMatrixLocation = gl.GetUniformLocation(p.ID, gl.Str("viewMatrixIn"+"\x00"))
	p.InverseViewMatrixLocation = gl.GetUniformLocation(p.ID, gl.Str("inverseViewMatrixIn"+"\x00"))

	p.DiffuseColorLocation = gl.GetUniformLocation(p.ID, gl.Str("diffuseColorIn"+"\x00"))
	p.DiffuseTextureLocation = gl.GetUniformLocation(p.ID, gl.Str("diffuseTextureIn"+"\x00"))
	p.SkyboxTextureLocation = gl.GetUniformLocation(p.ID, gl.Str("skyboxTextureIn"+"\x00"))
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
