package opengl

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewProgram() *Program {
	return &Program{}
}

type Program struct {
	id uint32
}

func (p *Program) ID() uint32 {
	return p.id
}

func (p *Program) Allocate(info ProgramAllocateInfo) error {
	p.id = gl.CreateProgram()
	if p.id == 0 {
		return fmt.Errorf("failed to allocate program")
	}

	if info.VertexShader != nil {
		gl.AttachShader(p.id, info.VertexShader.ID())
		defer gl.DetachShader(p.id, info.VertexShader.ID())
	}
	if info.FragmentShader != nil {
		gl.AttachShader(p.id, info.FragmentShader.ID())
		defer gl.DetachShader(p.id, info.FragmentShader.ID())
	}

	gl.LinkProgram(p.id)
	if !p.isLinkSuccessful() {
		return fmt.Errorf("failed to link program: %s", p.getInfoLog())
	}
	return nil
}

func (p *Program) UniformLocation(name string) int32 {
	result := gl.GetUniformLocation(p.id, gl.Str(name+"\x00"))
	runtime.KeepAlive(name) // TODO: This probably does nothing, since Str takes a concatenated string
	return result
}

func (p *Program) Release() error {
	gl.DeleteProgram(p.id)
	p.id = 0
	return nil
}

func (p *Program) isLinkSuccessful() bool {
	var status int32
	gl.GetProgramiv(p.id, gl.LINK_STATUS, &status)
	return status != gl.FALSE
}

func (p *Program) getInfoLog() string {
	var logLength int32
	gl.GetProgramiv(p.id, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetProgramInfoLog(p.id, logLength, nil, gl.Str(log))
	return log
}

type ProgramAllocateInfo struct {
	VertexShader   *Shader
	FragmentShader *Shader
}
