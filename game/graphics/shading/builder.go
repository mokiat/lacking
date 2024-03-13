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

func (b *Builder) Variable(index VariableIndex) VariableDetails {
	return b.variables[index]
}

func (b *Builder) Parameters(operation OperationDetails) []VariableIndex {
	offset := operation.ParamsOffset
	count := operation.ParamsCount
	return b.parameters[offset : offset+count]
}

func (b *Builder) Operations() []OperationDetails {
	return b.operations
}

func (b *Builder) Value(value float32) Vec1Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeFloat32Value,
		Source: VariableSourceCode,
	})
	offset, count := b.addParameters(varIndex)
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationDefineVec1,
	})
	return Vec1Variable(varIndex)
}

func (b *Builder) Vec1() Vec1Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec1,
		Source: VariableSourceCode,
	})
	offset, count := b.addParameters(varIndex)
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationDefineVec1,
	})
	return Vec1Variable(varIndex)
}

func (b *Builder) Vec2() Vec2Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec2,
		Source: VariableSourceCode,
	})
	offset, count := b.addParameters(varIndex)
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationDefineVec2,
	})
	return Vec2Variable(varIndex)
}

func (b *Builder) Vec3() Vec3Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec3,
		Source: VariableSourceCode,
	})
	offset, count := b.addParameters(varIndex)
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationDefineVec3,
	})
	return Vec3Variable(varIndex)
}

func (b *Builder) Vec4() Vec4Variable {
	varIndex := b.createVariable(VariableDetails{
		Type:   VariableTypeVec4,
		Source: VariableSourceCode,
	})
	offset, count := b.addParameters(varIndex)
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationDefineVec4,
	})
	return Vec4Variable(varIndex)
}

func (b *Builder) UniformVec1() Vec1Variable {
	return Vec1Variable(b.createVariable(VariableDetails{
		Type:   VariableTypeVec1,
		Source: VariableSourceUniform,
	}))
}

func (b *Builder) UniformVec2() Vec2Variable {
	return Vec2Variable(b.createVariable(VariableDetails{
		Type:   VariableTypeVec2,
		Source: VariableSourceUniform,
	}))
}

func (b *Builder) UniformVec3() Vec3Variable {
	return Vec3Variable(b.createVariable(VariableDetails{
		Type:   VariableTypeVec3,
		Source: VariableSourceUniform,
	}))
}

func (b *Builder) UniformVec4() Vec4Variable {
	return Vec4Variable(b.createVariable(VariableDetails{
		Type:   VariableTypeVec4,
		Source: VariableSourceUniform,
	}))
}

func (b *Builder) AssignVec1(target Vec1Variable, x Vec1Variable) {
	offset, count := b.addParameters(VariableIndex(target), VariableIndex(x))
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationTypeAssignVec1,
	})
}

func (b *Builder) AssignVec2(target Vec2Variable, x Vec1Variable, y Vec1Variable) {
	offset, count := b.addParameters(VariableIndex(target), VariableIndex(x), VariableIndex(y))
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationTypeAssignVec2,
	})
}

func (b *Builder) AssignVec3(target Vec3Variable, x Vec1Variable, y Vec1Variable, z Vec1Variable) {
	offset, count := b.addParameters(VariableIndex(target), VariableIndex(x), VariableIndex(y), VariableIndex(z))
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationTypeAssignVec3,
	})
}

func (b *Builder) AssignVec4(target Vec4Variable, x Vec1Variable, y Vec1Variable, z Vec1Variable, w Vec1Variable) {
	offset, count := b.addParameters(VariableIndex(target), VariableIndex(x), VariableIndex(y), VariableIndex(z), VariableIndex(w))
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationTypeAssignVec4,
	})
}

func (b *Builder) ForwardOutputColor(color Vec4Variable) {
	offset, count := b.addParameters(VariableIndex(color))
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationTypeForwardOutputColor,
	})
}

func (b *Builder) ForwardAlphaDiscard(alpha Vec1Variable, threshold Vec1Variable) {
	offset, count := b.addParameters(VariableIndex(alpha), VariableIndex(threshold))
	b.createOperation(OperationDetails{
		ParamsOffset: offset,
		ParamsCount:  count,
		Type:         OperationTypeForwardAlphaDiscard,
	})
}

func (b *Builder) createVariable(details VariableDetails) VariableIndex {
	index := VariableIndex(len(b.variables))
	b.variables = append(b.variables, details)
	return index
}

func (b *Builder) addParameters(params ...VariableIndex) (uint32, uint32) {
	offset := len(b.parameters)
	count := len(params)
	b.parameters = append(b.parameters, params...)
	return uint32(offset), uint32(count)
}

func (b *Builder) createOperation(details OperationDetails) {
	b.operations = append(b.operations, details)
}
