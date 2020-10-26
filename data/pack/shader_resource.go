package pack

import (
	"fmt"
	"io/ioutil"
)

type OpenShaderResourceAction struct {
	locator ResourceLocator
	uri     string
	shader  *Shader
}

func (a *OpenShaderResourceAction) Describe() string {
	return fmt.Sprintf("open_shader_resource(uri: %q)", a.uri)
}

func (a *OpenShaderResourceAction) Shader() *Shader {
	if a.shader == nil {
		panic("reading data from unprocessed action")
	}
	return a.shader
}

func (a *OpenShaderResourceAction) Run() error {
	in, err := a.locator.Open(a.uri)
	if err != nil {
		return err
	}
	defer in.Close()

	data, err := ioutil.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read shader content: %w", err)
	}

	a.shader = &Shader{
		Source: string(data),
	}
	return nil
}
