package lsl

const (
	// TypeNameBool is the name of the boolean type.
	TypeNameBool = "bool"

	// TypeNameInt is the name of the signed 32bit integer type.
	TypeNameInt = "int"

	// TypeNameUint is the name of the unsigned 32bit integer type.
	TypeNameUint = "uint"

	// TypeNameFloat is the name of the 32bit floating point type.
	TypeNameFloat = "float"

	// TypeNameVec2 is the name of the 32bit floating point 2D vector type.
	TypeNameVec2 = "vec2"

	// TypeNameVec3 is the name of the 32bit floating point 3D vector type.
	TypeNameVec3 = "vec3"

	// TypeNameVec4 is the name of the 32bit floating point 4D vector type.
	TypeNameVec4 = "vec4"

	// TypeNameBVec2 is the name of the boolean 2D vector type.
	TypeNameBVec2 = "bvec2"

	// TypeNameBVec3 is the name of the boolean 3D vector type.
	TypeNameBVec3 = "bvec3"

	// TypeNameBVec4 is the name of the boolean 4D vector type.
	TypeNameBVec4 = "bvec4"

	// TypeNameIVec2 is the name of the signed 32bit integer 2D vector type.
	TypeNameIVec2 = "ivec2"

	// TypeNameIVec3 is the name of the signed 32bit integer 3D vector type.
	TypeNameIVec3 = "ivec3"

	// TypeNameIVec4 is the name of the signed 32bit integer 4D vector type.
	TypeNameIVec4 = "ivec4"

	// TypeNameUVec2 is the name of the unsigned 32bit integer 2D vector type.
	TypeNameUVec2 = "uvec2"

	// TypeNameUVec3 is the name of the unsigned 32bit integer 3D vector type.
	TypeNameUVec3 = "uvec3"

	// TypeNameUVec4 is the name of the unsigned 32bit integer 4D vector type.
	TypeNameUVec4 = "uvec4"

	// TypeNameMat2 is the name of the 2x2 matrix type.
	TypeNameMat2 = "mat2"

	// TypeNameMat3 is the name of the 3x3 matrix type.
	TypeNameMat3 = "mat3"

	// TypeNameMat4 is the name of the 4x4 matrix type.
	TypeNameMat4 = "mat4"

	// TypeNameSampler2D is the name of the 2D sampler type.
	TypeNameSampler2D = "sampler2D"

	// TypeNameSamplerCube is the name of the cube sampler type.
	TypeNameSamplerCube = "samplerCube"
)

// Expression represents a sequence of tokens that can be evaluated to a value.
//
// Example:
//
//	1 + sin(10)
type Expression interface {

	// GetPos returns the position of the expression in the source code.
	GetPos() Position

	_isExpression()
}

// Declaration represents a top-level construct in a shader.
//
// Example:
//
//	func hello() {
//	}
type Declaration interface {

	// GetPos returns the position of the declaration in the source code.
	GetPos() Position

	_isDeclaration()
}

// Statement represents a code line in a function.
//
// Example:
//
//	a := 5.0
type Statement interface {

	// GetPos returns the position of the statement in the source code.
	GetPos() Position

	_isStatement()
}

// Shader represents a complete shader program.
type Shader struct {
	Declarations []Declaration
}

func (s *Shader) TextureBlocks() []*TextureBlockDeclaration {
	var blocks []*TextureBlockDeclaration
	for _, decl := range s.Declarations {
		if block, ok := decl.(*TextureBlockDeclaration); ok {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func (s *Shader) FindTextureBlock() (*TextureBlockDeclaration, bool) {
	for _, decl := range s.Declarations {
		if block, ok := decl.(*TextureBlockDeclaration); ok {
			return block, true
		}
	}
	return nil, false
}

func (s *Shader) UniformBlocks() []*UniformBlockDeclaration {
	var blocks []*UniformBlockDeclaration
	for _, decl := range s.Declarations {
		if block, ok := decl.(*UniformBlockDeclaration); ok {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func (s *Shader) FindUniformBlock() (*UniformBlockDeclaration, bool) {
	for _, decl := range s.Declarations {
		if block, ok := decl.(*UniformBlockDeclaration); ok {
			return block, true
		}
	}
	return nil, false
}

func (s *Shader) VaryingBlocks() []*VaryingBlockDeclaration {
	var blocks []*VaryingBlockDeclaration
	for _, decl := range s.Declarations {
		if block, ok := decl.(*VaryingBlockDeclaration); ok {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func (s *Shader) FindVaryingBlock() (*VaryingBlockDeclaration, bool) {
	for _, decl := range s.Declarations {
		if block, ok := decl.(*VaryingBlockDeclaration); ok {
			return block, true
		}
	}
	return nil, false
}

func (s *Shader) Functions() []*FunctionDeclaration {
	var functions []*FunctionDeclaration
	for _, decl := range s.Declarations {
		if fn, ok := decl.(*FunctionDeclaration); ok {
			functions = append(functions, fn)
		}
	}
	return functions
}

func (s *Shader) FindFunction(name string) (*FunctionDeclaration, bool) {
	for _, decl := range s.Declarations {
		if fn, ok := decl.(*FunctionDeclaration); ok && fn.Name == name {
			return fn, true
		}
	}
	return nil, false
}

// TextureBlockDeclaration represents a texture block declaration.
type TextureBlockDeclaration struct {
	Pos    Position
	Fields []Field
}

func (t *TextureBlockDeclaration) GetPos() Position {
	return t.Pos
}

func (*TextureBlockDeclaration) _isDeclaration() {}

// UniformBlockDeclaration represents a uniform block declaration.
type UniformBlockDeclaration struct {
	Pos    Position
	Fields []Field
}

func (u *UniformBlockDeclaration) GetPos() Position {
	return u.Pos
}

func (*UniformBlockDeclaration) _isDeclaration() {}

// VaryingBlockDeclaration represents a varying block declaration.
type VaryingBlockDeclaration struct {
	Pos    Position
	Fields []Field
}

func (v *VaryingBlockDeclaration) GetPos() Position {
	return v.Pos
}

func (*VaryingBlockDeclaration) _isDeclaration() {}

// StructTypeDeclaration represents a structure type declaration.
type StructTypeDeclaration struct {
	Pos    Position
	Name   string
	Fields []Field
}

func (s *StructTypeDeclaration) GetPos() Position {
	return s.Pos
}

func (*StructTypeDeclaration) _isDeclaration() {}

// FunctionDeclaration represents a function declaration.
type FunctionDeclaration struct {
	Pos        Position
	Name       string
	Inputs     []Field
	OutputType string
	Body       StatementList
}

func (f *FunctionDeclaration) GetPos() Position {
	return f.Pos
}

func (*FunctionDeclaration) _isDeclaration() {}

// StatementList represents a list of statements.
type StatementList []Statement

func (l StatementList) GetPos() Position {
	if len(l) == 0 {
		return At(0, 0)
	}
	return l[0].GetPos()
}

func (StatementList) _isStatement() {}

// VariableDeclaration represents a variable declaration.
type VariableDeclaration struct {
	Pos        Position
	Name       string
	Type       string
	Assignment Expression
}

func (v *VariableDeclaration) GetPos() Position {
	return v.Pos
}

func (*VariableDeclaration) _isStatement() {}

// FunctionCall represents a function call.
type FunctionCall struct {
	Owner     Expression
	Arguments []Expression
}

func (f *FunctionCall) GetPos() Position {
	return f.Owner.GetPos()
}

func (*FunctionCall) _isExpression() {}

func (*FunctionCall) _isStatement() {}

// Assignment represents an assignment statement.
type Assignment struct {
	Target     Expression
	Expression Expression
	Operator   string
}

func (a *Assignment) GetPos() Position {
	return a.Target.GetPos()
}

func (*Assignment) _isStatement() {}

// Conditional represents a conditional statement.
type Conditional struct {
	Pos       Position
	Condition Expression
	Then      StatementList
	Else      Statement
}

func (c *Conditional) GetPos() Position {
	return c.Pos
}

func (*Conditional) _isStatement() {}

// Return represents a return statement.
type Return struct {
	Pos        Position
	Expression Expression
}

func (r *Return) GetPos() Position {
	return r.Pos
}

func (*Return) _isStatement() {}

// Discard represents a statement that does nothing.
type Discard struct {
	Pos Position
}

func (d *Discard) GetPos() Position {
	return d.Pos
}

func (*Discard) _isStatement() {}

// BoolLiteral represents a boolean literal.
type BoolLiteral struct {
	Pos   Position
	Value bool
}

func (b *BoolLiteral) GetPos() Position {
	return b.Pos
}

func (*BoolLiteral) _isExpression() {}

// IntLiteral represents an integer literal.
type IntLiteral struct {
	Pos   Position
	Value int64
}

func (i *IntLiteral) GetPos() Position {
	return i.Pos
}

func (*IntLiteral) _isExpression() {}

// FloatLiteral represents a floating point literal.
type FloatLiteral struct {
	Pos   Position
	Value float64
}

func (f *FloatLiteral) GetPos() Position {
	return f.Pos
}

func (*FloatLiteral) _isExpression() {}

// Identifier represents a reference to a variable or a function.
type Identifier struct {
	Pos  Position
	Name string
}

func (i *Identifier) GetPos() Position {
	return i.Pos
}

func (*Identifier) _isExpression() {}

// FieldIdentifier represents a reference to a field of a structure.
type FieldIdentifier struct {
	Owner Expression
	Field Identifier
}

func (f *FieldIdentifier) GetPos() Position {
	return f.Owner.GetPos()
}

func (*FieldIdentifier) _isExpression() {}

// UnaryExpression represents a unary operation.
type UnaryExpression struct {
	Pos      Position
	Operator string
	Operand  Expression
}

func (u *UnaryExpression) GetPos() Position {
	return u.Pos
}

func (*UnaryExpression) _isExpression() {}

// BinaryExpression represents a binary operation.
type BinaryExpression struct {
	Operator string
	Left     Expression
	Right    Expression
}

func (b *BinaryExpression) GetPos() Position {
	return b.Left.GetPos()
}

func (*BinaryExpression) _isExpression() {}

// Field represents a field in a block or in a parameter list.
type Field struct {
	Pos  Position
	Name string
	Type string
}
