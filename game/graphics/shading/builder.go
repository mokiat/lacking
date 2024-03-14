package shading

type GenericBuilderFunc func(builder *Builder)

func NewBuilder() *Builder {
	return &Builder{}
}

type Builder struct {
	variables  []VariableDetails
	parameters []VariableIndex
	operations []OperationDetails
}

func (b *Builder) Uniforms() []VariableIndex {
	var result []VariableIndex
	for i, variable := range b.variables {
		if variable.Source == VariableSourceUniform {
			result = append(result, VariableIndex(i))
		}
	}
	return result
}

func (b *Builder) Varyings() []VariableIndex {
	var result []VariableIndex
	for i, variable := range b.variables {
		if variable.Source == VariableSourceVarying {
			result = append(result, VariableIndex(i))
		}
	}
	return result
}

func (b *Builder) Variable(index VariableIndex) VariableDetails {
	return b.variables[index]
}

func (b *Builder) Parameters(pRange ParamRange) []VariableIndex {
	return b.parameters[pRange.Offset : pRange.Offset+pRange.Count]
}

func (b *Builder) InputParameters(operation OperationDetails) []VariableIndex {
	return b.Parameters(operation.InputParams)
}

func (b *Builder) OutputParameters(operation OperationDetails) []VariableIndex {
	return b.Parameters(operation.OutputParams)
}

func (b *Builder) Operations() []OperationDetails {
	return b.operations
}

func (b *Builder) Value(value float32) Vec1Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeFloat32Value,
		Source: VariableSourceCode,
	})
	b.createOperation(OperationDetails{
		OutputParams: b.addParameters(varIndex),
		Type:         OperationDefineVec1,
	})
	return Vec1Variable{
		VariableIndex: varIndex,
	}
}

func (b *Builder) Vec1() Vec1Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec1,
		Source: VariableSourceCode,
	})
	b.createOperation(OperationDetails{
		OutputParams: b.addParameters(varIndex),
		Type:         OperationDefineVec1,
	})
	return Vec1Variable{
		VariableIndex: varIndex,
	}
}

func (b *Builder) Vec2() Vec2Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec2,
		Source: VariableSourceCode,
	})
	b.createOperation(OperationDetails{
		OutputParams: b.addParameters(varIndex),
		Type:         OperationDefineVec2,
	})
	return Vec2Variable{
		VariableIndex: varIndex,
	}
}

func (b *Builder) Vec3() Vec3Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec3,
		Source: VariableSourceCode,
	})
	b.createOperation(OperationDetails{
		OutputParams: b.addParameters(varIndex),
		Type:         OperationDefineVec3,
	})
	return Vec3Variable{
		VariableIndex: varIndex,
	}
}

func (b *Builder) Vec4() Vec4Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec4,
		Source: VariableSourceCode,
	})
	b.createOperation(OperationDetails{
		OutputParams: b.addParameters(varIndex),
		Type:         OperationDefineVec4,
	})
	return Vec4Variable{
		VariableIndex: varIndex,
	}
}

func (b *Builder) UniformVec1() Vec1Variable {
	return Vec1Variable{
		VariableIndex: b.createVariable(VariableDetails{
			Type:   VariableTypeVec1,
			Source: VariableSourceUniform,
		}),
	}
}

func (b *Builder) UniformVec2() Vec2Variable {
	return Vec2Variable{
		VariableIndex: b.createVariable(VariableDetails{
			Type:   VariableTypeVec2,
			Source: VariableSourceUniform,
		}),
	}
}

func (b *Builder) UniformVec3() Vec3Variable {
	return Vec3Variable{
		VariableIndex: b.createVariable(VariableDetails{
			Type:   VariableTypeVec3,
			Source: VariableSourceUniform,
		}),
	}
}

func (b *Builder) UniformVec4() Vec4Variable {
	return Vec4Variable{
		VariableIndex: b.createVariable(VariableDetails{
			Type:   VariableTypeVec4,
			Source: VariableSourceUniform,
		}),
	}
}

func (b *Builder) AssignVec1(target Vec1Variable, x Vec1Variable) {
	b.createOperation(OperationDetails{
		InputParams:  b.addParameters(x.Index()),
		OutputParams: b.addParameters(target.Index()),
		Type:         OperationTypeAssignVec1,
	})

}

func (b *Builder) AssignVec2(target Vec2Variable, x Vec1Variable, y Vec1Variable) {
	b.createOperation(OperationDetails{
		InputParams:  b.addParameters(x.Index(), y.Index()),
		OutputParams: b.addParameters(target.Index()),
		Type:         OperationTypeAssignVec2,
	})
}

func (b *Builder) AssignVec3(target Vec3Variable, x Vec1Variable, y Vec1Variable, z Vec1Variable) {
	b.createOperation(OperationDetails{
		InputParams:  b.addParameters(x.Index(), y.Index(), z.Index()),
		OutputParams: b.addParameters(target.Index()),
		Type:         OperationTypeAssignVec3,
	})
}

func (b *Builder) AssignVec4(target Vec4Variable, x Vec1Variable, y Vec1Variable, z Vec1Variable, w Vec1Variable) {
	b.createOperation(OperationDetails{
		InputParams:  b.addParameters(x.Index(), y.Index(), z.Index(), w.Index()),
		OutputParams: b.addParameters(target.Index()),
		Type:         OperationTypeAssignVec4,
	})
}

func (b *Builder) ForwardOutputColor(color Vec4Variable) {
	b.createOperation(OperationDetails{
		InputParams: b.addParameters(color.Index()),
		Type:        OperationTypeForwardOutputColor,
	})
}

func (b *Builder) ForwardAlphaDiscard(alpha Vec1Variable, threshold Vec1Variable) {
	b.createOperation(OperationDetails{
		InputParams: b.addParameters(alpha.Index(), threshold.Index()),
		Type:        OperationTypeForwardAlphaDiscard,
	})
}

func (b *Builder) createVariable(details VariableDetails) VariableIndex {
	index := VariableIndex(len(b.variables))
	b.variables = append(b.variables, details)
	return index
}

func (b *Builder) addParameters(params ...VariableIndex) ParamRange {
	offset := len(b.parameters)
	count := len(params)
	b.parameters = append(b.parameters, params...)
	return ParamRange{
		Offset: uint32(offset),
		Count:  uint32(count),
	}
}

func (b *Builder) createOperation(details OperationDetails) {
	b.operations = append(b.operations, details)
}
