package preset

import (
	"bytes"
	"fmt"
)

func NewShaderBuilder(template string) *ShaderBuilder {
	return &ShaderBuilder{
		version:  "410",
		features: []string{},
		template: template,
	}
}

type ShaderBuilder struct {
	version  string
	features []string
	template string
}

func (b *ShaderBuilder) SetVersion(version string) {
	b.version = version
}

func (b *ShaderBuilder) AddFeature(feature string) {
	b.features = append(b.features, feature)
}

func (b *ShaderBuilder) Build() string {
	buffer := &bytes.Buffer{}
	fmt.Fprint(buffer, "#version ")
	fmt.Fprintln(buffer, b.version)
	fmt.Fprintln(buffer)
	for _, feature := range b.features {
		fmt.Fprint(buffer, "#define ")
		fmt.Fprintln(buffer, feature)
	}
	fmt.Fprintln(buffer, b.template)
	return buffer.String()
}
