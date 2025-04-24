package lsl

// Validate validates the provided shader against the provided schema.
func Validate(shader *Shader, schema Schema) error {
	// return NewValidator(shader, schema).Validate()
	return nil
}

// // NewValidator creates a new validator for the provided shader and schema.
// func NewValidator(shader *Shader, schema Schema) *Validator {
// 	return &Validator{
// 		shader: shader,
// 		schema: schema,
// 	}
// }

// // Validator is a shader validator.
// type Validator struct {
// 	shader *Shader
// 	schema Schema
// }

// // ValidateFunctionCallExpression validates a function call.
// func (v *Validator) ValidateFunctionCallExpression(ctx *ValidationContext, call *FunctionCall) (ResolvedType, error) {
// 	identifier, ok := call.Owner.(*Identifier)
// 	if !ok {
// 		return ResolvedType{}, &ValidateError{
// 			Pos:     call.GetPos(),
// 			Message: "only global functions can be called",
// 		}
// 	}

// 	if !ctx.IsKnownFunction(identifier.Name) {
// 		return ResolvedType{}, &ValidateError{
// 			Pos:     call.GetPos(),
// 			Message: "function not found",
// 		}
// 	}

// 	argumentTypes := make([]ResolvedType, len(call.Arguments))
// 	for i, argument := range call.Arguments {
// 		resolvedType, err := v.ValidateExpression(ctx, argument)
// 		if err != nil {
// 			return ResolvedType{}, err
// 		}
// 		argumentTypes[i] = resolvedType
// 	}

// 	// TODO: Use a digest of the argument types to find the function overload.
// 	returnType, ok := ctx.FunctionReturnType(identifier.Name, argumentTypes)
// 	if !ok {
// 		return ResolvedType{}, &ValidateError{
// 			Pos:     call.GetPos(),
// 			Message: "matching function overload not found",
// 		}
// 	}

// 	return returnType, nil
// }

// // ValidateIdentifierExpression validates an identifier and returns its resolved
// // type.
// func (v *Validator) ValidateIdentifierExpression(ctx *ValidationContext, identifier *Identifier) (ResolvedType, error) {
// 	identifierType, ok := ctx.IdentifierType(identifier.Name)
// 	if !ok {
// 		return ResolvedType{}, &ValidateError{
// 			Pos:     identifier.Pos,
// 			Message: "unknown identifier",
// 		}
// 	}
// 	return identifierType, nil
// }

// // ValidateFieldIdentifierExpression validates a field identifier and returns
// // its resolved type.
// func (v *Validator) ValidateFieldIdentifierExpression(ctx *ValidationContext, fieldIdentifier *FieldIdentifier) (ResolvedType, error) {
// 	identifierType, err := v.ValidateExpression(ctx, fieldIdentifier.Owner)
// 	if err != nil {
// 		return ResolvedType{}, err
// 	}

// 	fieldType, ok := ctx.FieldType(identifierType, fieldIdentifier.Field.Name)
// 	if !ok {
// 		return ResolvedType{}, &ValidateError{
// 			Pos:     fieldIdentifier.Field.Pos,
// 			Message: "unknown field",
// 		}
// 	}

// 	return fieldType, nil
// }

// // ValidateUnaryExpression validates a unary expression and returns its resolved
// // type.
// func (v *Validator) ValidateUnaryExpression(ctx *ValidationContext, unary *UnaryExpression) (ResolvedType, error) {
// 	resolvedType, err := v.ValidateExpression(ctx, unary.Operand)
// 	if err != nil {
// 		return ResolvedType{}, err
// 	}

// 	returnType, ok := ctx.UnaryOperationReturnType(resolvedType, unary.Operator)
// 	if !ok {
// 		return ResolvedType{}, &ValidateError{
// 			Pos:     unary.Pos,
// 			Message: "invalid unary operation",
// 		}
// 	}

// 	return returnType, nil
// }

// // ValidateBinaryExpression validates a binary expression and returns its
// // resolved type.
// func (v *Validator) ValidateBinaryExpression(ctx *ValidationContext, binary *BinaryExpression) (ResolvedType, error) {
// 	leftType, err := v.ValidateExpression(ctx, binary.Left)
// 	if err != nil {
// 		return ResolvedType{}, err
// 	}

// 	rightType, err := v.ValidateExpression(ctx, binary.Right)
// 	if err != nil {
// 		return ResolvedType{}, err
// 	}

// 	returnType, ok := ctx.BinaryOperationReturnType(leftType, rightType, binary.Operator)
// 	if !ok {
// 		return ResolvedType{}, &ValidateError{
// 			Pos:     binary.GetPos(),
// 			Message: "invalid binary operation",
// 		}
// 	}

// 	return returnType, nil
// }

// // ValidateExpression validates an expression and returns its resolved type.
// func (v *Validator) ValidateExpression(ctx *ValidationContext, expression Expression) (ResolvedType, error) {
// 	switch expr := expression.(type) {
// 	case *FunctionCall:
// 		return v.ValidateFunctionCallExpression(ctx, expr)
// 	case *BoolLiteral:
// 		return BuiltInResolvedType(TypeNameBool, UsageScopeAll), nil
// 	case *IntLiteral:
// 		return BuiltInResolvedType(TypeNameInt, UsageScopeAll), nil
// 	case *FloatLiteral:
// 		return BuiltInResolvedType(TypeNameFloat, UsageScopeAll), nil
// 	case *Identifier:
// 		return v.ValidateIdentifierExpression(ctx, expr)
// 	case *FieldIdentifier:
// 		return v.ValidateFieldIdentifierExpression(ctx, expr)
// 	case *UnaryExpression:
// 		return v.ValidateUnaryExpression(ctx, expr)
// 	case *BinaryExpression:
// 		return v.ValidateBinaryExpression(ctx, expr)
// 	default:
// 		return ResolvedType{}, fmt.Errorf("unsupported expression type %T", expr)
// 	}
// }

// // ValidateVariableDeclarationStatement validates a variable declaration.
// func (v *Validator) ValidateVariableDeclarationStatement(ctx *ValidationContext, decl *VariableDeclaration) error {
// 	if strings.HasPrefix(decl.Name, "#") {
// 		return &ValidateError{
// 			Pos:     decl.Pos,
// 			Message: "variable name cannot start with #",
// 		}
// 	}

// 	if ok := ctx.IsKnownIdentifier(decl.Name); ok {
// 		return &ValidateError{
// 			Pos:     decl.Pos,
// 			Message: "variable name already in use",
// 		}
// 	}

// 	var declarationType opt.T[ResolvedType]
// 	if decl.Type != "" {
// 		resolvedType, ok := ctx.TypeByName(decl.Type)
// 		if !ok {
// 			return &ValidateError{
// 				Pos:     decl.Pos,
// 				Message: "varaible has unknown type",
// 			}
// 		}
// 		declarationType = opt.V(resolvedType)
// 	}

// 	var assignmentType opt.T[ResolvedType]
// 	if decl.Assignment != nil {
// 		resolvedType, err := v.ValidateExpression(ctx, decl.Assignment)
// 		if err != nil {
// 			return err
// 		}
// 		assignmentType = opt.V(resolvedType)
// 	}

// 	if !declarationType.Specified && !assignmentType.Specified {
// 		// Note: This should not occur in practice, since the parser should catch
// 		// it. But we'll keep it here for safety.
// 		return &ValidateError{
// 			Pos:     decl.Pos,
// 			Message: "variable must have a declared type or use an assignment",
// 		}
// 	}

// 	if declarationType.Specified && assignmentType.Specified {
// 		if declarationType.Value != assignmentType.Value {
// 			return &ValidateError{
// 				Pos:     decl.Pos,
// 				Message: "variable type does not match assignment expression type",
// 			}
// 		}
// 	}

// 	// TODO: Figure out what the idea behind this is.
// 	// if !v.schema.IsAllowedVariableType(decl.Type) {
// 	// 	return fmt.Errorf("type %q not allowed in variable declaration", decl.Type)
// 	// }

// 	if declarationType.Specified {
// 		ctx.RegisterIdentifier(decl.Name, declarationType.Value)
// 	} else {
// 		ctx.RegisterIdentifier(decl.Name, assignmentType.Value)
// 	}
// 	return nil
// }

// // ValidateFunctionCallStatement validates a function call.
// func (v *Validator) ValidateFunctionCallStatement(ctx *ValidationContext, call *FunctionCall) error {
// 	_, err := v.ValidateFunctionCallExpression(ctx, call)
// 	return err
// }

// // ValidateAssignmentStatement validates an assignment statement.
// func (v *Validator) ValidateAssignmentStatement(ctx *ValidationContext, assignment *Assignment) error {
// 	targetType, err := v.ValidateExpression(ctx, assignment.Target)
// 	if err != nil {
// 		return err
// 	}

// 	if targetType.ReadOnly {
// 		return &ValidateError{
// 			Pos:     assignment.GetPos(),
// 			Message: "cannot assign to read-only identifier",
// 		}
// 	}

// 	assignmentType, err := v.ValidateExpression(ctx, assignment.Expression)
// 	if err != nil {
// 		return err
// 	}

// 	if !assignmentType.NonCopy {
// 		return &ValidateError{
// 			Pos:     assignment.GetPos(),
// 			Message: "the assignment expression cannot be copied",
// 		}
// 	}

// 	if !ctx.CanAssign(targetType, assignmentType, assignment.Operator) {
// 		return &ValidateError{
// 			Pos:     assignment.GetPos(),
// 			Message: "cannot assign to target type with specified operator and expression type",
// 		}
// 	}

// 	return nil
// }

// // ValidateConditionalStatement validates a conditional statement.
// func (v *Validator) ValidateConditionalStatement(ctx *ValidationContext, conditional *Conditional) error {
// 	return nil // TODO
// }

// // ValidateDiscardStatement validates a discard statement.
// func (v *Validator) ValidateDiscardStatement(ctx *ValidationContext, discard *Discard) error {
// 	ctx.ReduceScope(UsageScopeFragmentShader)
// 	return nil
// }

// // ValidateStatement validates a statement.
// func (v *Validator) ValidateStatement(ctx *ValidationContext, statement Statement) error {
// 	switch stmt := statement.(type) {
// 	case *VariableDeclaration:
// 		return v.ValidateVariableDeclarationStatement(ctx, stmt)
// 	case *FunctionCall:
// 		return v.ValidateFunctionCallStatement(ctx, stmt)
// 	case *Assignment:
// 		return v.ValidateAssignmentStatement(ctx, stmt)
// 	case *Conditional:
// 		return v.ValidateConditionalStatement(ctx, stmt)
// 	case *Discard:
// 		return v.ValidateDiscardStatement(ctx, stmt)
// 	default:
// 		return fmt.Errorf("unsupported statement type %T", stmt)
// 	}
// }

// func (v *Validator) ValidateDeclarationShallow(declaration Declaration) error {
// 	switch decl := declaration.(type) {
// 	default:
// 		return fmt.Errorf("unsupported declaration type %T", decl)
// 	}
// }

// // Validate validates the shader.
// func (v *Validator) Validate() error {
// 	for _, decl := range v.shader.Declarations {
// 		if err := v.ValidateDeclarationShallow(decl); err != nil {
// 			return err
// 		}
// 	}

// 	// textureBlocks := v.shader.TextureBlocks()
// 	// if len(textureBlocks) > 1 {
// 	// 	return errors.New("multiple texture blocks not allowed")
// 	// }
// 	// if len(textureBlocks) == 1 {
// 	// 	if err := v.validateTextureBlock(textureBlocks[0]); err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	// uniformBlocks := v.shader.UniformBlocks()
// 	// if len(uniformBlocks) > 1 {
// 	// 	return errors.New("multiple uniform blocks not allowed")
// 	// }
// 	// if len(uniformBlocks) == 1 {
// 	// 	if err := v.validateUniformBlock(uniformBlocks[0]); err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	// varyingBlocks := v.shader.VaryingBlocks()
// 	// if len(varyingBlocks) > 1 {
// 	// 	return errors.New("multiple varying blocks not allowed")
// 	// }
// 	// if len(varyingBlocks) == 1 {
// 	// 	if err := v.validateVaryingBlock(varyingBlocks[0]); err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	// functions := v.shader.Functions()
// 	// vertexFunctions := gog.Select(functions, func(fn *FunctionDeclaration) bool {
// 	// 	return fn.Name == "#vertex"
// 	// })
// 	// if len(vertexFunctions) > 1 {
// 	// 	return errors.New("multiple #vertex functions not allowed")
// 	// }
// 	// if len(vertexFunctions) == 1 {
// 	// 	if err := v.validateFunction(vertexFunctions[0]); err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	// fragmentFunctions := gog.Select(functions, func(fn *FunctionDeclaration) bool {
// 	// 	return fn.Name == "#fragment"
// 	// })
// 	// if len(fragmentFunctions) > 1 {
// 	// 	return errors.New("multiple #fragment functions not allowed")
// 	// }
// 	// if len(fragmentFunctions) == 1 {
// 	// 	if err := v.validateFunction(fragmentFunctions[0]); err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	// hasCustomFunctions := slices.ContainsFunc(functions, func(fn *FunctionDeclaration) bool {
// 	// 	return (fn.Name != "#vertex") && (fn.Name != "#fragment")
// 	// })
// 	// if hasCustomFunctions {
// 	// 	return errors.New("custom functions not supported yet")
// 	// }

// 	return nil
// }

// func (v *Validator) validateShader() error {
// 	ctx := ValidationContext{}

// 	// TODO: Regoster functions from schema.
// 	for _, function := range v.shader.Functions() {
// 		if err := ctx.RegisterFunction(function); err != nil {
// 			return fmt.Errorf("failed to register function %q: %w", function.Name, err)
// 		}
// 	}

// 	panic("TODO")
// }

// // func (v *Validator) validateTextureBlock(block *TextureBlockDeclaration) error {
// // 	for _, field := range block.Fields {
// // 		if strings.HasPrefix(field.Name, "#") {
// // 			return fmt.Errorf("field %q cannot start with #", field.Name)
// // 		}
// // 		if _, ok := v.variables[field.Name]; ok {
// // 			return fmt.Errorf("field %q already declared", field.Name)
// // 		} else {
// // 			v.variables[field.Name] = field.Type
// // 		}
// // 		if !v.schema.IsAllowedTextureType(field.Type) {
// // 			return fmt.Errorf("type %q not allowed in texture block", field.Type)
// // 		}
// // 	}
// // 	return nil
// // }

// // func (v *Validator) validateUniformBlock(block *UniformBlockDeclaration) error {
// // 	for _, field := range block.Fields {
// // 		if strings.HasPrefix(field.Name, "#") {
// // 			return fmt.Errorf("field %q cannot start with #", field.Name)
// // 		}
// // 		if _, ok := v.variables[field.Name]; ok {
// // 			return fmt.Errorf("field %q already declared", field.Name)
// // 		} else {
// // 			v.variables[field.Name] = field.Type
// // 		}
// // 		if !v.schema.IsAllowedUniformType(field.Type) {
// // 			return fmt.Errorf("type %q not allowed in uniform block", field.Type)
// // 		}
// // 	}
// // 	return nil
// // }

// // func (v *Validator) validateVaryingBlock(block *VaryingBlockDeclaration) error {
// // 	for _, field := range block.Fields {
// // 		if strings.HasPrefix(field.Name, "#") {
// // 			return fmt.Errorf("field %q cannot start with #", field.Name)
// // 		}
// // 		if _, ok := v.variables[field.Name]; ok {
// // 			return fmt.Errorf("field %q already declared", field.Name)
// // 		} else {
// // 			v.variables[field.Name] = field.Type
// // 		}
// // 		if !v.schema.IsAllowedVaryingType(field.Type) {
// // 			return fmt.Errorf("type %q not allowed in varying block", field.Type)
// // 		}
// // 	}
// // 	return nil
// // }

// // func (v *Validator) validateFunction(function *FunctionDeclaration) error {
// // 	for _, stmt := range function.Body {
// // 		if err := v.validateStatement(stmt); err != nil {
// // 			return err
// // 		}
// // 	}
// // 	return nil
// // }

// // func (v *Validator) validateStatement(stmt Statement) error {
// // 	switch stmt := stmt.(type) {
// // 	case *VariableDeclaration:
// // 		return v.validateVariableDeclaration(stmt)
// // 	}
// // 	return nil
// // }

// // func (v *Validator) validateVariableDeclaration(decl *VariableDeclaration) error {
// // 	if strings.HasPrefix(decl.Name, "#") {
// // 		return fmt.Errorf("variable %q cannot start with #", decl.Name)
// // 	}
// // 	if _, ok := v.variables[decl.Name]; ok {
// // 		return fmt.Errorf("variable %q already declared", decl.Name)
// // 	} else {
// // 		v.variables[decl.Name] = decl.Type
// // 	}
// // 	if !v.schema.IsAllowedVariableType(decl.Type) {
// // 		return fmt.Errorf("type %q not allowed in variable declaration", decl.Type)
// // 	}
// // 	return nil
// // }

// type ValidationContext struct {
// 	// identifiers map[string]Object
// }

// func (c *ValidationContext) ReduceScope(scope UsageScope) {
// 	// TODO
// }

// func (c *ValidationContext) IsKnownIdentifier(name string) bool {
// 	return false // FIXME
// }

// func (c *ValidationContext) IdentifierType(name string) (ResolvedType, bool) {
// 	// identifier, ok := c.identifiers[name]
// 	// if !ok {
// 	// 	return ResolvedType{}, false
// 	// }
// 	// return identifier.Type, true
// 	return ResolvedType{}, false // FIXME
// }

// func (c *ValidationContext) FieldType(identifierType ResolvedType, fieldName string) (ResolvedType, bool) {
// 	// TODO
// 	return ResolvedType{}, false
// }

// func (c *ValidationContext) RegisterIdentifier(name string, typ ResolvedType) {
// 	// TODO
// }

// func (c *ValidationContext) TypeByName(name string) (ResolvedType, bool) {
// 	return ResolvedType{}, false // FIXME
// }

// func (c *ValidationContext) IsKnownFunction(name string) bool {
// 	return false // FIXME
// }

// func (c *ValidationContext) FunctionReturnType(name string, params []ResolvedType) (ResolvedType, bool) {
// 	return ResolvedType{}, false // FIXME
// }

// func (c *ValidationContext) CanAssign(targetType, assignmentType ResolvedType, operator string) bool {
// 	return false // FIXME
// }

// func (c *ValidationContext) UnaryOperationReturnType(operand ResolvedType, operator string) (ResolvedType, bool) {
// 	return ResolvedType{}, false // FIXME
// }

// func (c *ValidationContext) BinaryOperationReturnType(left, right ResolvedType, operator string) (ResolvedType, bool) {
// 	return ResolvedType{}, false // FIXME
// }

// func (c *ValidationContext) Push() {

// }

// func (c *ValidationContext) Pop() {
// }

// func (c *ValidationContext) RegisterFunction(function *FunctionDeclaration) error {
// 	// name := function.Name
// 	// identifier, ok := c.identifiers[name]

// 	// if ok {
// 	// 	if identifier.Type != "function" {
// 	// 		return fmt.Errorf("identifier %q is not a function", name)
// 	// 	}
// 	// 	mapsha
// 	// 	if _, ok := identifier.Overloads[[32]Field{}]; ok {
// 	// 		return fmt.Errorf("function %q already registered", name)
// 	// 	}
// 	// } else {
// 	// 	c.identifiers[name] = Object{
// 	// 		Name:      name,
// 	// 		Type:      "function",
// 	// 		Overloads: make(map[[32]Field]*FunctionDeclaration),
// 	// 	}
// 	// }

// 	return nil // FIXME
// }

// // ValidateError is an error that occurs during validation.
// type ValidateError struct {

// 	// Pos is the position in the source code where the error occurred.
// 	Pos Position

// 	// Message is the error message.
// 	Message string
// }

// // Error returns the error message.
// func (e *ValidateError) Error() string {
// 	return fmt.Sprintf("shader code error %s at position %s", e.Message, e.Pos)
// }

// type Object struct {
// 	Name      string
// 	Type      string
// 	Overloads map[uint32]*FunctionDeclaration
// }
