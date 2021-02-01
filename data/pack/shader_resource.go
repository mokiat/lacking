package pack

import (
	"fmt"
	"hash"
	"io/ioutil"
	"sync"
)

func OpenShaderResource(uri string) *OpenShaderResourceAction {
	return &OpenShaderResourceAction{
		uri: uri,
	}
}

var _ ShaderProvider = (*OpenShaderResourceAction)(nil)

type OpenShaderResourceAction struct {
	uri string

	resultMutex  sync.Mutex
	resultDigest []byte
	result       *Shader
}

func (a *OpenShaderResourceAction) Describe() string {
	return fmt.Sprintf("open_shader_resource(uri: %q)", a.uri)
}

func (a *OpenShaderResourceAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "open_shader_resource", HashableParams{
		"uri": a.uri,
	})
}

func (a *OpenShaderResourceAction) Shader(ctx *Context) (*Shader, error) {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	a.resultMutex.Lock()
	defer a.resultMutex.Unlock()

	digest, err := CalculateDigest(a)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate digest: %w", err)
	}
	if EqualDigests(digest, a.resultDigest) {
		return a.result, nil
	}

	result, err := a.run(ctx)
	if err != nil {
		return nil, err
	}

	a.result = result
	a.resultDigest = digest
	return result, nil
}

func (a *OpenShaderResourceAction) run(ctx *Context) (*Shader, error) {
	var shader *Shader
	readShader := func(storage Storage) error {
		in, err := storage.OpenResource(a.uri)
		if err != nil {
			return err
		}
		defer in.Close()

		data, err := ioutil.ReadAll(in)
		if err != nil {
			return fmt.Errorf("failed to read shader content: %w", err)
		}
		shader = &Shader{
			Source: string(data),
		}
		return nil
	}

	if err := ctx.IO(readShader); err != nil {
		return nil, err
	}
	return shader, nil
}
