package lsl

import (
	"errors"
	"fmt"
	"strings"
)

type UsageScope uint8

const (
	UsageScopeNone         UsageScope = 0
	UsageScopeVertexShader UsageScope = 1 << iota
	UsageScopeFragmentShader
)

func Validate(shader *Shader, schema Schema) error {
	return NewValidator(shader, schema).Validate()
}

func NewValidator(shader *Shader, schema Schema) *Validator {
	return &Validator{
		shader: shader,
		schema: schema,
	}
}

type Validator struct {
	shader *Shader
	schema Schema
}

func (v *Validator) Validate() error {
	textureBlocks := v.shader.TextureBlocks()
	if len(textureBlocks) > 1 {
		return errors.New("multiple texture blocks not allowed")
	}
	if len(textureBlocks) == 1 {
		if err := v.validateTextureBlock(textureBlocks[0]); err != nil {
			return err
		}
	}

	uniformBlocks := v.shader.UniformBlocks()
	if len(uniformBlocks) > 1 {
		return errors.New("multiple uniform blocks not allowed")
	}
	if len(uniformBlocks) == 1 {
		if err := v.validateUniformBlock(uniformBlocks[0]); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateTextureBlock(block *TextureBlockDeclaration) error {
	for _, field := range block.Fields {
		if strings.HasPrefix(field.Name, "#") {
			return fmt.Errorf("field %q cannot start with #", field.Name)
		}
		if !v.schema.IsAllowedTextureType(field.Type) {
			return fmt.Errorf("type %q not allowed in texture block", field.Type)
		}
	}
	return nil
}

func (v *Validator) validateUniformBlock(block *UniformBlockDeclaration) error {
	for _, field := range block.Fields {
		if strings.HasPrefix(field.Name, "#") {
			return fmt.Errorf("field %q cannot start with #", field.Name)
		}
		if !v.schema.IsAllowedUniformType(field.Type) {
			return fmt.Errorf("type %q not allowed in uniform block", field.Type)
		}
	}
	return nil
}
