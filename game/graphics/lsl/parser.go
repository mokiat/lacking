package lsl

import (
	"fmt"
	"strconv"

	"github.com/mokiat/gog/ds"
)

// ParseError is an error that occurs during parsing.
type ParseError struct {

	// Pos is the position in the source code where the error occurred.
	Pos Position

	// Message is the error message.
	Message string
}

// Error returns the error message.
func (e *ParseError) Error() string {
	return fmt.Sprintf("shader source code error %q at position %s", e.Message, e.Pos)
}

// Parse parses the given LSL source code and returns a shader AST object.
func Parse(source string) (*Shader, error) {
	return NewParser(source).ParseShader()
}

// NewParser creates a new LSL parser for the given source code.
func NewParser(source string) *Parser {
	tokenizer := NewTokenizer(source)
	return &Parser{
		tokenizer: tokenizer,
		token:     tokenizer.Next(),
	}
}

// Parser is responsible for parsing LSL source code into a shader AST object.
type Parser struct {
	tokenizer *Tokenizer
	token     Token
}

// ParseFieldGroup parses a block containing field declarations such as one
// used in textures/uniforms/varyings blocks or struct declarations.
//
// Example:
//
//	(
//		color vec3
//	)
func (p *Parser) ParseFieldGroup(opening, closing string) ([]Field, error) {
	var fields []Field

	openingToken := p.nextToken()
	if !openingToken.IsSpecificOperator(opening) {
		return nil, &ParseError{
			Pos:     openingToken.Pos,
			Message: "expected an opening bracket",
		}
	}

	token := p.peekToken()
	for !token.IsSpecificOperator(closing) {
		switch {
		case token.IsError():
			return nil, &ParseError{
				Pos:     token.Pos,
				Message: fmt.Sprintf("tokenization error: %s", token.Value),
			}

		case token.IsNewLine():
			if err := p.consumeNewLine(); err != nil {
				return nil, err
			}

		case token.IsComment():
			if err := p.consumeComment(); err != nil {
				return nil, err
			}

		case token.IsIdentifier():
			field, err := p.parseField()
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)

		default:
			return nil, &ParseError{
				Pos:     token.Pos,
				Message: "expected a name identifier or end of list",
			}
		}
		token = p.peekToken()
	}

	closingToken := p.nextToken()
	if !closingToken.IsSpecificOperator(closing) {
		return nil, &ParseError{
			Pos:     closingToken.Pos,
			Message: "expected a closing bracket",
		}
	}

	if err := p.consumeRemainingLine(); err != nil {
		return nil, err
	}

	return fields, nil
}

// ParseTextureBlock parses a block containing texture fields.
//
// Example:
//
//	texture (
//		color sampler2D
//	)
func (p *Parser) ParseTextureBlock() (*TextureBlockDeclaration, error) {
	uniformToken := p.nextToken()
	if !uniformToken.IsSpecificIdentifier(KeywordTexture) {
		return nil, &ParseError{
			Pos:     uniformToken.Pos,
			Message: fmt.Sprintf("expected %q keyword", KeywordTexture),
		}
	}
	fields, err := p.parseFieldBlock(GroupStart, GroupEnd)
	if err != nil {
		return nil, err
	}
	return &TextureBlockDeclaration{
		Pos:    uniformToken.Pos,
		Fields: fields,
	}, nil
}

// ParseUniformBlock parses a block containing uniform fields.
//
// Example:
//
//	uniform (
//		color vec4
//	)
func (p *Parser) ParseUniformBlock() (*UniformBlockDeclaration, error) {
	uniformToken := p.nextToken()
	if !uniformToken.IsSpecificIdentifier(KeywordUniform) {
		return nil, &ParseError{
			Pos:     uniformToken.Pos,
			Message: fmt.Sprintf("expected %q keyword", KeywordUniform),
		}
	}
	fields, err := p.parseFieldBlock(GroupStart, GroupEnd)
	if err != nil {
		return nil, err
	}
	return &UniformBlockDeclaration{
		Pos:    uniformToken.Pos,
		Fields: fields,
	}, nil
}

// ParseVaryingBlock parses a block containing varying fields.
//
// Example:
//
//	varying (
//		color vec3
//	)
func (p *Parser) ParseVaryingBlock() (*VaryingBlockDeclaration, error) {
	varyingToken := p.nextToken()
	if !varyingToken.IsSpecificIdentifier(KeywordVarying) {
		return nil, &ParseError{
			Pos:     varyingToken.Pos,
			Message: fmt.Sprintf("expected %q keyword", KeywordVarying),
		}
	}
	fields, err := p.parseFieldBlock(GroupStart, GroupEnd)
	if err != nil {
		return nil, err
	}
	return &VaryingBlockDeclaration{
		Pos:    varyingToken.Pos,
		Fields: fields,
	}, nil
}

// ParseTypeDeclaration parses a type declaration.
//
// Example:
//
//	type MyType struct {
//		color vec3
//	}
func (p *Parser) ParseTypeDeclaration() (Declaration, error) {
	typeToken := p.nextToken()
	if !typeToken.IsSpecificIdentifier(KeywordType) {
		return nil, &ParseError{
			Pos:     typeToken.Pos,
			Message: fmt.Sprintf("expected %q keyword", KeywordType),
		}
	}

	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return nil, &ParseError{
			Pos:     nameToken.Pos,
			Message: "expected a name identifier",
		}
	}

	structToken := p.nextToken()
	if !structToken.IsSpecificIdentifier(KeywordStruct) {
		return nil, &ParseError{
			Pos:     structToken.Pos,
			Message: fmt.Sprintf("expected %q keyword", KeywordStruct),
		}
	}

	fields, err := p.parseFieldBlock(BlockStart, BlockEnd)
	if err != nil {
		return nil, err
	}

	return &StructTypeDeclaration{
		Pos:    typeToken.Pos,
		Name:   nameToken.Value,
		Fields: fields,
	}, nil
}

// ParseExpression uses the Shunting Yard algorithm to parse an expression and
// return its AST form.
//
// Example:
//
//	1 + 2 * 3
func (p *Parser) ParseExpression() (Expression, error) {
	valStack := ds.NewStack[Expression](2)
	opStack := ds.NewStack[string](1)

	firstExpressionToken := p.peekToken()

	value, err := p.parseExpressionValue()
	if err != nil {
		return nil, err
	}
	valStack.Push(value)

	nextToken := p.peekToken()
	for nextToken.IsBinaryOperator() {
		operatorToken := p.nextToken()
		operator := operatorToken.Value
		operatorPrec := operatorPrecedence(operator)
		for !opStack.IsEmpty() {
			prevOperator := opStack.Peek()
			prevOperatorPrec := operatorPrecedence(prevOperator)
			if prevOperatorPrec < operatorPrec {
				break
			}
			opStack.Pop() // pop it
			rightValue := valStack.Pop()
			leftValue := valStack.Pop()
			valStack.Push(&BinaryExpression{
				Operator: prevOperator,
				Left:     leftValue,
				Right:    rightValue,
			})
		}
		opStack.Push(operator)

		nextToken = p.peekToken()
		switch {
		case nextToken.IsNewLine():
			if err := p.consumeNewLine(); err != nil {
				return nil, err
			}
		case nextToken.IsComment():
			if err := p.consumeComment(); err != nil {
				return nil, err
			}
		}

		valueToken, err := p.parseExpressionValue()
		if err != nil {
			return nil, err
		}
		valStack.Push(valueToken)

		nextToken = p.peekToken()
	}

	for valStack.Size() > 1 {
		rightValue := valStack.Pop()
		leftValue := valStack.Pop()
		if opStack.IsEmpty() {
			// This should not really happen, as long as binary expressions are
			// properly maintained.
			return nil, &ParseError{
				Pos:     firstExpressionToken.Pos,
				Message: "expression has insufficient operators",
			}
		}
		operator := opStack.Pop()
		valStack.Push(&BinaryExpression{
			Operator: operator,
			Left:     leftValue,
			Right:    rightValue,
		})
	}
	if valStack.IsEmpty() {
		// This should not really happen, as long as binary expressions are
		// properly maintained.
		return nil, &ParseError{
			Pos:     firstExpressionToken.Pos,
			Message: "expression is empty or malformed",
		}
	}
	return valStack.Pop(), nil
}

// ParseArgumentBlock parses a block containing argument declarations such as
// one used in function invocations.
//
// Example:
//
//	(10.5, ^66, 2+3)
func (p *Parser) ParseArgumentBlock() ([]Expression, error) {
	var args []Expression

	openingToken := p.nextToken()
	if !openingToken.IsSpecificOperator(GroupStart) {
		return nil, &ParseError{
			Pos:     openingToken.Pos,
			Message: "expected an opening bracket",
		}
	}

	token := p.peekToken()
	for !token.IsSpecificOperator(GroupEnd) {
		switch {
		case token.IsError():
			return nil, &ParseError{
				Pos:     token.Pos,
				Message: fmt.Sprintf("tokenization error: %s", token.Value),
			}

		case token.IsNewLine():
			if err := p.consumeNewLine(); err != nil {
				return nil, err
			}

		case token.IsComment():
			if err := p.consumeComment(); err != nil {
				return nil, err
			}

		default:
			arg, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			nextToken := p.peekToken()
			switch {
			case nextToken.IsSpecificOperator(SeparatorComma):
				p.nextToken()
			case nextToken.IsSpecificOperator(GroupEnd):
				// Do nothing.
			default:
				return nil, &ParseError{
					Pos:     nextToken.Pos,
					Message: "expected a comma or a closing bracket",
				}
			}
		}
		token = p.peekToken()
	}

	closingToken := p.nextToken()
	if !closingToken.IsSpecificOperator(GroupEnd) {
		return nil, &ParseError{
			Pos:     closingToken.Pos,
			Message: "expected a closing bracket",
		}
	}
	return args, nil
}

// ParseStatement parses a single imperative statement from the source code.
func (p *Parser) ParseStatement() (Statement, error) {
	token := p.peekToken()
	switch {
	case token.IsSpecificIdentifier(KeywordDiscard):
		return p.parseDiscardStatement()
	case token.IsSpecificIdentifier(KeywordVar):
		return p.parseVariableDeclaration()
	case token.IsSpecificIdentifier(KeywordIf):
		return p.parseConditionalStatement()
	case token.IsIdentifier():
		return p.parseImperativeStatement()
	default:
		return nil, &ParseError{
			Pos:     token.Pos,
			Message: "expected a statement",
		}
	}
}

// ParseStatementList parses a list of statements from the source code.
func (p *Parser) ParseStatementList() (StatementList, error) {
	var statements []Statement
	for {
		token := p.peekToken()
		switch {
		case token.IsError():
			return nil, &ParseError{
				Pos:     token.Pos,
				Message: fmt.Sprintf("tokenization error: %s", token.Value),
			}
		case token.IsNewLine():
			if err := p.consumeNewLine(); err != nil {
				return nil, err
			}
		case token.IsComment():
			if err := p.consumeComment(); err != nil {
				return nil, err
			}
		case token.IsIdentifier():
			statement, err := p.ParseStatement()
			if err != nil {
				return nil, err
			}
			statements = append(statements, statement)
		default:
			return statements, nil
		}
	}
}

func (p *Parser) ParseFunction() (*FunctionDeclaration, error) {
	var decl FunctionDeclaration

	funcToken := p.nextToken()
	if !funcToken.IsSpecificIdentifier("func") {
		return nil, fmt.Errorf("expected func keyword")
	}
	decl.Pos = funcToken.Pos

	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return nil, fmt.Errorf("expected function name identifier")
	}
	decl.Name = nameToken.Value

	paramBracketStart := p.nextToken()
	if !paramBracketStart.IsSpecificOperator("(") {
		return nil, fmt.Errorf("expected opening bracket")
	}

	inputParams, err := p.parseParameterList()
	if err != nil {
		return nil, fmt.Errorf("error parsing input params: %w", err)
	}
	decl.Inputs = inputParams

	paramBracketEnd := p.nextToken()
	if !paramBracketEnd.IsSpecificOperator(")") {
		return nil, fmt.Errorf("expected closing bracket")
	}

	nextToken := p.peekToken()
	if nextToken.IsIdentifier() {
		outputTypeToken := p.nextToken()
		decl.OutputType = outputTypeToken.Value
	}

	if err := p.consumeBlockStart(); err != nil {
		return nil, fmt.Errorf("error parsing function block start: %w", err)
	}

	statements, err := p.ParseStatementList()
	if err != nil {
		return nil, fmt.Errorf("error parsing function body: %w", err)
	}
	decl.Body = statements

	if err := p.consumeBlockEnd(true); err != nil {
		return nil, fmt.Errorf("error parsing function footer: %w", err)
	}
	return &decl, nil
}

// ParseShader parses the LSL source code and returns a shader AST object.
// If the source code is invalid, an error is returned.
func (p *Parser) ParseShader() (*Shader, error) {
	var shader Shader
	token := p.peekToken()
	for !token.IsEOF() {
		switch {
		case token.IsError():
			return nil, fmt.Errorf("error token: %v", token)
		case token.IsNewLine():
			if err := p.consumeNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}
		case token.IsComment():
			if err := p.consumeComment(); err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}
		case token.IsSpecificIdentifier(KeywordType):
			decl, err := p.ParseTypeDeclaration()
			if err != nil {
				return nil, err
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier(KeywordTexture):
			decl, err := p.ParseTextureBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing texture block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier(KeywordUniform):
			decl, err := p.ParseUniformBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing uniform block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier(KeywordVarying):
			decl, err := p.ParseVaryingBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing varying block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier("func"):
			decl, err := p.ParseFunction()
			if err != nil {
				return nil, fmt.Errorf("error parsing function: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		default:
			return nil, fmt.Errorf("unexpected token: %v", token)
		}
		token = p.peekToken()
	}
	return &shader, nil
}

// peekToken returns the next token without consuming it.
func (p *Parser) peekToken() Token {
	return p.token
}

// nextToken returns the next token and consumes it.
func (p *Parser) nextToken() Token {
	token := p.token
	p.token = p.tokenizer.Next()
	return token
}

// consumeNewLine expects that a new line token follows and consumes it.
func (p *Parser) consumeNewLine() error {
	token := p.nextToken()
	if !token.IsNewLine() {
		return &ParseError{
			Pos:     token.Pos,
			Message: "expected a new line token",
		}
	}
	return nil
}

// consumeComment expects that a comment token follows and consumes it together
// with a new line token if it follows.
func (p *Parser) consumeComment() error {
	commentToken := p.nextToken()
	if !commentToken.IsComment() {
		return &ParseError{
			Pos:     commentToken.Pos,
			Message: "expected a comment token",
		}
	}
	nextToken := p.peekToken()
	if nextToken.IsNewLine() {
		if err := p.consumeNewLine(); err != nil {
			return err
		}
	}
	return nil
}

// consumeRemainingLine consumes the remainder of the line, including any
// new line or comment tokens. It expects that anything to follow is
// non-vital (comments, new lines) and can be ignored.
func (p *Parser) consumeRemainingLine() error {
	token := p.peekToken()
	switch {
	case token.IsError():
		return &ParseError{
			Pos:     token.Pos,
			Message: fmt.Sprintf("tokenization error: %s", token.Value),
		}
	case token.IsEOF():
		return nil
	case token.IsNewLine():
		return p.consumeNewLine()
	case token.IsComment():
		return p.consumeComment()
	default:
		return &ParseError{
			Pos:     token.Pos,
			Message: "expected a comment, new line or end of file",
		}
	}
}

// consumeBlockStart expects the next token to follow to be an opening bracket
// and consumes it. It also consumes the remaining line.
func (p *Parser) consumeBlockStart() error {
	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator(BlockStart) {
		return &ParseError{
			Pos:     bracketToken.Pos,
			Message: "expected an opening bracket",
		}
	}
	if err := p.consumeRemainingLine(); err != nil {
		return err
	}
	return nil
}

// consumeBlockEnd expects the next token to follow to be a closing bracket
// and consumes it. It consumes the remaining line depending if requested.
func (p *Parser) consumeBlockEnd(consumeRemainder bool) error {
	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator(BlockEnd) {
		return &ParseError{
			Pos:     bracketToken.Pos,
			Message: "expected a closing bracket",
		}
	}
	if consumeRemainder {
		if err := p.consumeRemainingLine(); err != nil {
			return err
		}
	}
	return nil
}

// parseField parses a single field declaration.
//
// Example:
//
//	color vec3
func (p *Parser) parseField() (Field, error) {
	nameToken := p.nextToken()
	typeToken := p.nextToken()
	if !typeToken.IsIdentifier() {
		return Field{}, &ParseError{
			Pos:     typeToken.Pos,
			Message: "expected a type identifier",
		}
	}
	if err := p.consumeRemainingLine(); err != nil {
		return Field{}, err
	}
	return Field{
		Pos:  nameToken.Pos,
		Name: nameToken.Value,
		Type: typeToken.Value,
	}, nil
}

// parseFieldBlock parses a block containing one or more field declarations
// with a user-defined opening and closing brackets.
//
// Example:
//
//	o
//		color vec3
//		uv vec2
//	c
func (p *Parser) parseFieldBlock(opening, closing string) ([]Field, error) {
	var fields []Field

	nextToken := p.peekToken()
	switch {
	case nextToken.IsError():
		return nil, &ParseError{
			Pos:     nextToken.Pos,
			Message: fmt.Sprintf("tokenization error: %s", nextToken.Value),
		}

	case nextToken.IsSpecificOperator(opening):
		var err error
		fields, err = p.ParseFieldGroup(opening, closing)
		if err != nil {
			return nil, err
		}

	case nextToken.IsIdentifier():
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = []Field{field}

	default:
		return nil, &ParseError{
			Pos:     nextToken.Pos,
			Message: "expected a name identifier or an opening bracket",
		}
	}

	return fields, nil
}

// parseParameterList parses a list of parameter name and type pairs.
func (p *Parser) parseParameterList() ([]Field, error) {
	var params []Field

	for {
		token := p.peekToken()
		switch {
		case token.IsError():
			return nil, &ParseError{
				Pos:     token.Pos,
				Message: fmt.Sprintf("tokenization error: %s", token.Value),
			}

		case token.IsEOF():
			return params, nil

		case token.IsNewLine():
			if err := p.consumeNewLine(); err != nil {
				return nil, err
			}

		case token.IsComment():
			if err := p.consumeComment(); err != nil {
				return nil, err
			}

		case token.IsOperator():
			if token.IsSpecificOperator(SeparatorComma) {
				return nil, &ParseError{
					Pos:     token.Pos,
					Message: "unexpected comma",
				}
			}
			return params, nil // end it here

		case token.IsIdentifier():
			nameToken := p.nextToken()
			if !nameToken.IsIdentifier() {
				return nil, &ParseError{
					Pos:     nameToken.Pos,
					Message: "expected a name identifier",
				}
			}
			typeToken := p.nextToken()
			if !typeToken.IsIdentifier() {
				return nil, &ParseError{
					Pos:     typeToken.Pos,
					Message: "expected a type identifier",
				}
			}
			params = append(params, Field{
				Pos:  nameToken.Pos,
				Name: nameToken.Value,
				Type: typeToken.Value,
			})
			nextToken := p.peekToken()
			switch {
			case nextToken.IsError():
				return nil, &ParseError{
					Pos:     nextToken.Pos,
					Message: "tokenization error",
				}
			case nextToken.IsEOF():
				return params, nil
			case nextToken.IsOperator():
				if !nextToken.IsSpecificOperator(SeparatorComma) {
					return params, nil
				}
				p.nextToken() // consume the comma
			default:
				return nil, &ParseError{
					Pos:     nextToken.Pos,
					Message: "expected a comma or end of list",
				}
			}

		default:
			return nil, &ParseError{
				Pos:     token.Pos,
				Message: "expected a name identifier or end of list",
			}
		}
	}
}

// parseNumericExpression expects that the next token is a non-negative
// number of type int or float.
//
// Example:
//
//	15.3
func (p *Parser) parseNumericExpression() (Expression, error) {
	numberToken := p.nextToken()
	if !numberToken.IsNumber() {
		return nil, &ParseError{
			Pos:     numberToken.Pos,
			Message: "expected a number token",
		}
	}

	intValue, err := strconv.ParseInt(numberToken.Value, 10, 64)
	if err == nil {
		return &IntLiteral{
			Pos:   numberToken.Pos,
			Value: intValue,
		}, nil
	}

	floatValue, err := strconv.ParseFloat(numberToken.Value, 64)
	if err == nil {
		return &FloatLiteral{
			Pos:   numberToken.Pos,
			Value: floatValue,
		}, nil
	}

	return nil, &ParseError{
		Pos:     numberToken.Pos,
		Message: "expected a valid number",
	}
}

// parseUnaryExpression expects that the next token is a unary operator
// and parses the expression that follows it.
//
// Example:
//
//	^35
func (p *Parser) parseUnaryExpression() (Expression, error) {
	operatorToken := p.nextToken()
	if !operatorToken.IsUnaryOperator() {
		return nil, &ParseError{
			Pos:     operatorToken.Pos,
			Message: "expected a unary operator",
		}
	}

	expr, err := p.parseExpressionValue()
	if err != nil {
		return nil, err
	}

	return &UnaryExpression{
		Pos:      operatorToken.Pos,
		Operator: operatorToken.Value,
		Operand:  expr,
	}, nil
}

// parseIdentifierExpression expects that the next token is an identifier
// and parses the expression that follows it. It also handles field access
// and function calls.
//
// Example:
//
//	color.r
func (p *Parser) parseIdentifierExpression() (Expression, error) {
	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return nil, &ParseError{
			Pos:     nameToken.Pos,
			Message: "expected an identifier",
		}
	}

	var result Expression
	result = &Identifier{
		Pos:  nameToken.Pos,
		Name: nameToken.Value,
	}

	for {
		nextToken := p.peekToken()

		switch {
		case nextToken.IsSpecificOperator(AccessOperatorDot):
			p.nextToken() // consume the dot
			fieldToken := p.nextToken()
			if !fieldToken.IsIdentifier() {
				return nil, &ParseError{
					Pos:     fieldToken.Pos,
					Message: "expected an identifier",
				}
			}
			result = &FieldIdentifier{
				Owner: result,
				Field: Identifier{
					Pos:  fieldToken.Pos,
					Name: fieldToken.Value,
				},
			}

		case nextToken.IsSpecificOperator(GroupStart):
			args, err := p.ParseArgumentBlock()
			if err != nil {
				return nil, err
			}
			result = &FunctionCall{
				Owner:     result,
				Arguments: args,
			}

		default:
			return result, nil
		}
	}
}

// parseExpressionGroup expects the next token to be an opening bracket
// and parses the expression that follows it.
//
// Example:
//
//	(a + b)
func (p *Parser) parseExpressionGroup() (Expression, error) {
	openingToken := p.nextToken()
	if !openingToken.IsSpecificOperator(GroupStart) {
		return nil, &ParseError{
			Pos:     openingToken.Pos,
			Message: "expected an opening bracket",
		}
	}
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	closingToken := p.nextToken()
	if !closingToken.IsSpecificOperator(GroupEnd) {
		return nil, &ParseError{
			Pos:     closingToken.Pos,
			Message: "expected a closing bracket",
		}
	}
	return expr, nil
}

func (p *Parser) parseExpressionValue() (Expression, error) {
	token := p.peekToken()
	switch {
	case token.IsError():
		return nil, &ParseError{
			Pos:     token.Pos,
			Message: fmt.Sprintf("tokenization error: %s", token.Value),
		}

	case token.IsSpecificOperator(GroupStart):
		return p.parseExpressionGroup()

	case token.IsUnaryOperator():
		return p.parseUnaryExpression()

	case token.IsNumber():
		return p.parseNumericExpression()

	case token.IsSpecificIdentifier(KeywordTrue):
		nameToken := p.nextToken()
		return &BoolLiteral{
			Pos:   nameToken.Pos,
			Value: true,
		}, nil

	case token.IsSpecificIdentifier(KeywordFalse):
		nameToken := p.nextToken()
		return &BoolLiteral{
			Pos:   nameToken.Pos,
			Value: false,
		}, nil

	case token.IsIdentifier():
		return p.parseIdentifierExpression()

	default:
		return nil, &ParseError{
			Pos:     token.Pos,
			Message: "expected an expression value",
		}
	}
}

// parseDiscardStatement expects the next token to be a discard keyword
// and consumes it. It also consumes the remaining line.
//
// Example:
//
//	discard
func (p *Parser) parseDiscardStatement() (*Discard, error) {
	discardToken := p.nextToken()
	if !discardToken.IsSpecificIdentifier("discard") {
		return nil, &ParseError{
			Pos:     discardToken.Pos,
			Message: "expected discard keyword",
		}
	}
	if err := p.consumeRemainingLine(); err != nil {
		return nil, err
	}
	return &Discard{
		Pos: discardToken.Pos,
	}, nil
}

// parseVariableDeclaration expects the next token to be a var keyword
// and parses a variable declaration.
//
// Example:
//
//	var color vec3 = vec3(1.0, 0.0, 0.0)
func (p *Parser) parseVariableDeclaration() (*VariableDeclaration, error) {
	var decl VariableDeclaration

	varToken := p.nextToken()
	if !varToken.IsSpecificIdentifier(KeywordVar) {
		return nil, &ParseError{
			Pos:     varToken.Pos,
			Message: "expected var keyword",
		}
	}
	decl.Pos = varToken.Pos

	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return nil, &ParseError{
			Pos:     nameToken.Pos,
			Message: "expected a name identifier",
		}
	}
	decl.Name = nameToken.Value

	nextToken := p.peekToken()
	switch {
	case nextToken.IsIdentifier():
		typeToken := p.nextToken()
		decl.Type = typeToken.Value

		nextToken = p.peekToken()
		if nextToken.IsSpecificOperator(AssignmentOperatorEq) {
			p.nextToken() // consume the assignment operator
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			decl.Assignment = expr
		}

	case nextToken.IsSpecificOperator(AssignmentOperatorEq):
		p.nextToken() // consume the assignment operator
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		decl.Assignment = expr

	default:
		return nil, &ParseError{
			Pos:     nextToken.Pos,
			Message: "expected a type identifier or an assignment operator",
		}
	}

	if err := p.consumeRemainingLine(); err != nil {
		return nil, err
	}

	return &decl, nil
}

// parseConditionalStatement expects the next token to be an if keyword
// and parses a conditional statement.
//
// Example:
//
//	if (a > b) {
//		// then block
//	} else if (a > c) {
//		// else if block
//	} else {
//		// else block
//	}
func (p *Parser) parseConditionalStatement() (*Conditional, error) {
	var statement Conditional

	ifToken := p.nextToken()
	if !ifToken.IsSpecificIdentifier(KeywordIf) {
		return nil, &ParseError{
			Pos:     ifToken.Pos,
			Message: fmt.Sprintf("expected %q keyword", KeywordIf),
		}
	}
	statement.Pos = ifToken.Pos

	conditionExpression, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	statement.Condition = conditionExpression

	if err := p.consumeBlockStart(); err != nil {
		return nil, err
	}

	thenStatements, err := p.ParseStatementList()
	if err != nil {
		return nil, err
	}
	statement.Then = thenStatements

	if err := p.consumeBlockEnd(false); err != nil {
		return nil, err
	}

	nextToken := p.peekToken()
	if nextToken.IsSpecificIdentifier(KeywordElse) {
		p.nextToken() // consume else token
		nextToken = p.peekToken()
		if nextToken.IsSpecificIdentifier(KeywordIf) {
			elseIfConditional, err := p.parseConditionalStatement()
			if err != nil {
				return nil, err
			}
			statement.Else = elseIfConditional
		} else {
			if err := p.consumeBlockStart(); err != nil {
				return nil, err
			}
			elseStatements, err := p.ParseStatementList()
			if err != nil {
				return nil, err
			}
			statement.Else = elseStatements
			if err := p.consumeBlockEnd(true); err != nil {
				return nil, err
			}
		}
	} else {
		if err := p.consumeRemainingLine(); err != nil {
			return nil, err
		}
	}

	return &statement, nil
}

// parseImperativeStatement parses an imperative statement, which can be a
// function call or a variable assignment.
//
// Example:
//
//	color.r = 1.0
func (p *Parser) parseImperativeStatement() (Statement, error) {
	target, err := p.parseIdentifierExpression()
	if err != nil {
		return nil, err
	}

	switch target := target.(type) {
	case *FunctionCall:
		if err := p.consumeRemainingLine(); err != nil {
			return nil, err
		}
		return target, nil

	case *Identifier:
		nextToken := p.peekToken()
		switch {
		case nextToken.IsSpecificOperator(AssignmentOperatorAuto):
			p.nextToken() // consume the assignment operator
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.consumeRemainingLine(); err != nil {
				return nil, err
			}
			return &VariableDeclaration{
				Pos:        target.Pos,
				Name:       target.Name,
				Type:       "", // auto-assignment
				Assignment: expr,
			}, nil

		case nextToken.IsAssignmentOperator():
			operatorToken := p.nextToken()
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.consumeRemainingLine(); err != nil {
				return nil, err
			}
			return &Assignment{
				Target:     target,
				Expression: expr,
				Operator:   operatorToken.Value,
			}, nil

		default:
			return nil, &ParseError{
				Pos:     nextToken.Pos,
				Message: "expected an assignment operator",
			}
		}

	default:
		operatorToken := p.nextToken()
		if !operatorToken.IsAssignmentOperator() {
			return nil, &ParseError{
				Pos:     operatorToken.Pos,
				Message: "expected an assignment operator",
			}
		}
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.consumeRemainingLine(); err != nil {
			return nil, err
		}
		return &Assignment{
			Target:     target,
			Expression: expr,
			Operator:   operatorToken.Value,
		}, nil
	}
}
