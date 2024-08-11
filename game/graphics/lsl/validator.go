package lsl

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mokiat/gog"
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

		variables: make(map[string]string),
	}
}

type Validator struct {
	shader *Shader
	schema Schema

	variables map[string]string
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

	varyingBlocks := v.shader.VaryingBlocks()
	if len(varyingBlocks) > 1 {
		return errors.New("multiple varying blocks not allowed")
	}
	if len(varyingBlocks) == 1 {
		if err := v.validateVaryingBlock(varyingBlocks[0]); err != nil {
			return err
		}
	}

	functions := v.shader.Functions()
	vertexFunctions := gog.Select(functions, func(fn *FunctionDeclaration) bool {
		return fn.Name == "#vertex"
	})
	if len(vertexFunctions) > 1 {
		return errors.New("multiple #vertex functions not allowed")
	}
	if len(vertexFunctions) == 1 {
		if err := v.validateFunction(vertexFunctions[0]); err != nil {
			return err
		}
	}

	fragmentFunctions := gog.Select(functions, func(fn *FunctionDeclaration) bool {
		return fn.Name == "#fragment"
	})
	if len(fragmentFunctions) > 1 {
		return errors.New("multiple #fragment functions not allowed")
	}
	if len(fragmentFunctions) == 1 {
		if err := v.validateFunction(fragmentFunctions[0]); err != nil {
			return err
		}
	}

	hasCustomFunctions := slices.ContainsFunc(functions, func(fn *FunctionDeclaration) bool {
		return (fn.Name != "#vertex") && (fn.Name != "#fragment")
	})
	if hasCustomFunctions {
		return errors.New("custom functions not supported yet")
	}

	return nil
}

func (v *Validator) validateTextureBlock(block *TextureBlockDeclaration) error {
	for _, field := range block.Fields {
		if strings.HasPrefix(field.Name, "#") {
			return fmt.Errorf("field %q cannot start with #", field.Name)
		}
		if _, ok := v.variables[field.Name]; ok {
			return fmt.Errorf("field %q already declared", field.Name)
		} else {
			v.variables[field.Name] = field.Type
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
		if _, ok := v.variables[field.Name]; ok {
			return fmt.Errorf("field %q already declared", field.Name)
		} else {
			v.variables[field.Name] = field.Type
		}
		if !v.schema.IsAllowedUniformType(field.Type) {
			return fmt.Errorf("type %q not allowed in uniform block", field.Type)
		}
	}
	return nil
}

func (v *Validator) validateVaryingBlock(block *VaryingBlockDeclaration) error {
	for _, field := range block.Fields {
		if strings.HasPrefix(field.Name, "#") {
			return fmt.Errorf("field %q cannot start with #", field.Name)
		}
		if _, ok := v.variables[field.Name]; ok {
			return fmt.Errorf("field %q already declared", field.Name)
		} else {
			v.variables[field.Name] = field.Type
		}
		if !v.schema.IsAllowedVaryingType(field.Type) {
			return fmt.Errorf("type %q not allowed in varying block", field.Type)
		}
	}
	return nil
}

func (v *Validator) validateFunction(function *FunctionDeclaration) error {
	for _, stmt := range function.Body {
		if err := v.validateStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validateStatement(stmt Statement) error {
	switch stmt := stmt.(type) {
	case *VariableDeclaration:
		return v.validateVariableDeclaration(stmt)
	}
	return nil
}

func (v *Validator) validateVariableDeclaration(decl *VariableDeclaration) error {
	if strings.HasPrefix(decl.Name, "#") {
		return fmt.Errorf("variable %q cannot start with #", decl.Name)
	}
	if _, ok := v.variables[decl.Name]; ok {
		return fmt.Errorf("variable %q already declared", decl.Name)
	} else {
		v.variables[decl.Name] = decl.Type
	}
	if !v.schema.IsAllowedVariableType(decl.Type) {
		return fmt.Errorf("type %q not allowed in variable declaration", decl.Type)
	}
	return nil
}
