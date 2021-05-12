package opengl

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewShader() *Shader {
	return &Shader{}
}

type Shader struct {
	id uint32
}

func (s *Shader) ID() uint32 {
	return s.id
}

func (s *Shader) Allocate(info ShaderAllocateInfo) {
	if s.id != 0 {
		panic(fmt.Errorf("shader already allocated"))
	}
	s.id = gl.CreateShader(info.ShaderType)
	if s.id == 0 {
		panic(fmt.Errorf("failed to allocate shader"))
	}
	s.setSourceCode(info.SourceCode)
	gl.CompileShader(s.id)
	if !s.isCompileSuccessful() {
		panic(fmt.Errorf("failed to compile shader: %s", s.getInfoLog()))
	}
}

func (s *Shader) Release() {
	if s.id == 0 {
		panic(fmt.Errorf("shader already released"))
	}
	gl.DeleteShader(s.id)
	s.id = 0
}

func (s *Shader) setSourceCode(code string) {
	sources, free := gl.Strs(code + "\x00")
	defer free()
	gl.ShaderSource(s.id, 1, sources, nil)
}

func (s *Shader) isCompileSuccessful() bool {
	var status int32
	gl.GetShaderiv(s.id, gl.COMPILE_STATUS, &status)
	return status != gl.FALSE
}

func (s *Shader) getInfoLog() string {
	var logLength int32
	gl.GetShaderiv(s.id, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetShaderInfoLog(s.id, logLength, nil, gl.Str(log))
	runtime.KeepAlive(log)
	return log
}

type ShaderAllocateInfo struct {
	ShaderType uint32
	SourceCode string
}

func NewShaderSourceBuilder(template string) *ShaderSourceBuilder {
	return &ShaderSourceBuilder{
		version:  "460",
		features: []string{},
		template: template,
	}
}

type ShaderSourceBuilder struct {
	version  string
	features []string
	template string
}

func (b *ShaderSourceBuilder) SetVersion(version string) {
	b.version = version
}

func (b *ShaderSourceBuilder) AddFeature(feature string) {
	b.features = append(b.features, feature)
}

func (b *ShaderSourceBuilder) Build() string {
	var builder strings.Builder
	builder.WriteString("#version ")
	builder.WriteString(b.version)
	builder.WriteRune('\n')
	for _, feature := range b.features {
		builder.WriteString("#define ")
		builder.WriteString(feature)
		builder.WriteRune('\n')
	}
	builder.WriteString(b.template)
	builder.WriteRune('\n')
	return builder.String()
}
