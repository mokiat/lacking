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
	_isExpression()
}

// Declaration represents a top-level construct in a shader.
//
// Example:
//
//	func hello() {}
type Declaration interface {
	_isDeclaration()
}

// Statement represents a code line in a function.
type Statement interface {
	_isStatement()
}

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

func (s *Shader) FindFunction(name string) (*FunctionDeclaration, bool) {
	for _, decl := range s.Declarations {
		if fn, ok := decl.(*FunctionDeclaration); ok && fn.Name == name {
			return fn, true
		}
	}
	return nil, false
}

type TextureBlockDeclaration struct {
	Fields []Field
}

func (*TextureBlockDeclaration) _isDeclaration() {}

type UniformBlockDeclaration struct {
	Fields []Field
}

func (*UniformBlockDeclaration) _isDeclaration() {}

type VaryingBlockDeclaration struct {
	Fields []Field
}

func (*VaryingBlockDeclaration) _isDeclaration() {}

type FunctionDeclaration struct {
	Name    string
	Inputs  []Field
	Outputs []Field
	Body    []Statement
}

func (*FunctionDeclaration) _isDeclaration() {}

type VariableDeclaration struct {
	Name       string
	Type       string
	Assignment Expression
}

func (*VariableDeclaration) _isStatement() {}

type FunctionCall struct {
	Name      string
	Arguments []Expression
}

func (*FunctionCall) _isExpression() {}

func (*FunctionCall) _isStatement() {}

type Assignment struct {
	Operator   string
	Target     Expression
	Expression Expression
}

func (*Assignment) _isStatement() {}

type Conditional struct {
	Condition Expression
	Then      []Statement
	ElseIf    *Conditional
	Else      []Statement
}

func (*Conditional) _isStatement() {}

type Discard struct{}

func (*Discard) _isStatement() {}

type IntLiteral struct {
	Value int64
}

func (*IntLiteral) _isExpression() {}

type FloatLiteral struct {
	Value float64
}

func (*FloatLiteral) _isExpression() {}

// ExpressionGroup represents a paren enclosed expression.
type ExpressionGroup struct {
	Expression Expression
}

func (*ExpressionGroup) _isExpression() {}

// Identifier represents a reference to a variable or a function.
type Identifier struct {
	Name string
}

func (*Identifier) _isExpression() {}

// FieldIDentifier represents a reference to a field of a structure.
type FieldIdentifier struct {
	ObjName   string // TODO: Identifier, or maybe even expression?
	FieldName string // TODO: Identifier
}

func (*FieldIdentifier) _isExpression() {}

type Field struct {
	Name string
	Type string
}

type UnaryExpression struct {
	Operator string
	Operand  Expression
}

func (*UnaryExpression) _isExpression() {}

type BinaryExpression struct {
	Operator string
	Left     Expression
	Right    Expression
}

func (*BinaryExpression) _isExpression() {}
