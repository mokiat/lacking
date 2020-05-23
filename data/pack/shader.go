package pack

import (
	"fmt"
	"io/ioutil"
)

type ShaderResourceProvider interface {
	Shader() (string, error)
}

type ShaderResourceFile struct {
	Resource
}

func (f *ShaderResourceFile) Shader() (string, error) {
	data, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return "", fmt.Errorf("failed to read shader file: %w", err)
	}
	return string(data), nil
}
