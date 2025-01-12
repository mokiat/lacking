package lsl

import (
	"fmt"
	"strconv"

	"github.com/mokiat/gog/ds"
)

func MustParse(source string) *Shader {
	shader, err := Parse(source)
	if err != nil {
		panic(err)
	}
	return shader
}

func Parse(source string) (*Shader, error) {
	return NewParser(source).ParseShader()
}

func NewParser(source string) *Parser {
	tokenizer := NewTokenizer(source)
	return &Parser{
		tokenizer: tokenizer,
		token:     tokenizer.Next(),
	}
}

type Parser struct {
	tokenizer *Tokenizer
	token     Token
}

func (p *Parser) ParseShader() (*Shader, error) {
	var shader Shader
	token := p.peekToken()
	for !token.IsEOF() {
		switch {
		case token.IsNewLine():
			if err := p.ParseNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}
		case token.IsComment():
			if err := p.ParseComment(); err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}
		case token.IsSpecificIdentifier("textures"):
			decl, err := p.ParseTextureBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing texture block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier("uniforms"):
			decl, err := p.ParseUniformBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing uniform block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier("varyings"):
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

func (p *Parser) peekToken() Token {
	return p.token
}

func (p *Parser) nextToken() Token {
	token := p.token
	p.token = p.tokenizer.Next()
	return token
}

func (p *Parser) ParseNewLine() error {
	token := p.nextToken()
	if !token.IsNewLine() {
		return fmt.Errorf("expected new line")
	}
	return nil
}

func (p *Parser) ParseComment() error {
	commentToken := p.nextToken()
	if !commentToken.IsComment() {
		return fmt.Errorf("expected comment token")
	}
	nextToken := p.peekToken()
	if nextToken.IsNewLine() {
		if err := p.ParseNewLine(); err != nil {
			return fmt.Errorf("error parsing new line after comment: %w", err)
		}
	}
	return nil
}

func (p *Parser) ParseOptionalRemainder() error {
	token := p.peekToken()
	switch {
	case token.IsEOF():
		return nil
	case token.IsNewLine():
		return p.ParseNewLine()
	case token.IsComment():
		return p.ParseComment()
	default:
		return fmt.Errorf("expected new line or comment")
	}
}

func (p *Parser) ParseBlockStart() error {
	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator("{") {
		return fmt.Errorf("expected opening bracket")
	}
	if err := p.ParseOptionalRemainder(); err != nil {
		return fmt.Errorf("error parsing end of line: %w", err)
	}
	return nil
}

func (p *Parser) ParseBlockEnd() error {
	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator("}") {
		return fmt.Errorf("expected closing bracket")
	}
	if err := p.ParseOptionalRemainder(); err != nil {
		return fmt.Errorf("error parsing end of line: %w", err)
	}
	return nil
}

func (p *Parser) ParseNamedParameterList() ([]Field, error) {
	var params []Field

	for {
		token := p.peekToken()
		switch {
		case token.IsEOF():
			return params, nil

		case token.IsNewLine():
			if err := p.ParseNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}

		case token.IsComment():
			if err := p.ParseComment(); err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}

		case token.IsIdentifier():
			nameToken := p.nextToken()
			if !nameToken.IsIdentifier() {
				return nil, fmt.Errorf("expected name identifier")
			}
			typeToken := p.nextToken()
			if !typeToken.IsIdentifier() {
				return nil, fmt.Errorf("expected type identifier")
			}
			params = append(params, Field{
				Name: nameToken.Value,
				Type: typeToken.Value,
			})
			nextToken := p.peekToken()
			switch {
			case nextToken.IsEOF():
				return params, nil
			case nextToken.IsOperator():
				if !nextToken.IsSpecificOperator(",") {
					return params, nil
				}
				p.nextToken() // consume the comma
			default:
				return nil, fmt.Errorf("unexpected token: %v", nextToken)
			}

		case token.IsOperator():
			if token.IsSpecificOperator(",") {
				return nil, fmt.Errorf("unexpected token comma separator: %v", token)
			}
			return params, nil

		default:
			return nil, fmt.Errorf("unexpected token: %v", token)
		}
	}
}

func (p *Parser) ParseUnnamedParameterList() ([]Field, error) {
	var params []Field

	for {
		token := p.peekToken()
		switch {
		case token.IsEOF():
			return params, nil

		case token.IsNewLine():
			if err := p.ParseNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}

		case token.IsComment():
			if err := p.ParseComment(); err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}

		case token.IsIdentifier():
			typeToken := p.nextToken()
			if !typeToken.IsIdentifier() {
				return nil, fmt.Errorf("expected type identifier")
			}
			params = append(params, Field{
				Type: typeToken.Value,
			})
			nextToken := p.peekToken()
			switch {
			case nextToken.IsEOF():
				return params, nil
			case nextToken.IsOperator():
				if !nextToken.IsSpecificOperator(",") {
					return params, nil
				}
				p.nextToken() // consume the comma
			default:
				return nil, fmt.Errorf("unexpected token: %v", nextToken)
			}

		case token.IsOperator():
			if token.IsSpecificOperator(",") {
				return nil, fmt.Errorf("unexpected token comma separator: %v", token)
			}
			return params, nil

		default:
			return nil, fmt.Errorf("unexpected token: %v", token)
		}
	}
}

func (p *Parser) ParseTextureBlock() (*TextureBlockDeclaration, error) {
	uniformToken := p.nextToken()
	if !uniformToken.IsSpecificIdentifier("textures") {
		return nil, fmt.Errorf("expected uniform keyword")
	}
	if err := p.ParseBlockStart(); err != nil {
		return nil, fmt.Errorf("error parsing uniform block start: %w", err)
	}
	fields, err := p.ParseNamedParameterList()
	if err != nil {
		return nil, fmt.Errorf("error parsing varying block fields: %w", err)
	}
	if err := p.ParseBlockEnd(); err != nil {
		return nil, fmt.Errorf("error parsing uniform block end: %w", err)
	}
	return &TextureBlockDeclaration{
		Fields: fields,
	}, nil
}

func (p *Parser) ParseUniformBlock() (*UniformBlockDeclaration, error) {
	uniformToken := p.nextToken()
	if !uniformToken.IsSpecificIdentifier("uniforms") {
		return nil, fmt.Errorf("expected uniform keyword")
	}
	if err := p.ParseBlockStart(); err != nil {
		return nil, fmt.Errorf("error parsing uniform block start: %w", err)
	}
	fields, err := p.ParseNamedParameterList()
	if err != nil {
		return nil, fmt.Errorf("error parsing varying block fields: %w", err)
	}
	if err := p.ParseBlockEnd(); err != nil {
		return nil, fmt.Errorf("error parsing uniform block end: %w", err)
	}
	return &UniformBlockDeclaration{
		Fields: fields,
	}, nil
}

func (p *Parser) ParseVaryingBlock() (*VaryingBlockDeclaration, error) {
	varyingToken := p.nextToken()
	if !varyingToken.IsSpecificIdentifier("varyings") {
		return nil, fmt.Errorf("expected varying keyword")
	}
	if err := p.ParseBlockStart(); err != nil {
		return nil, fmt.Errorf("error parsing varying block start: %w", err)
	}
	fields, err := p.ParseNamedParameterList()
	if err != nil {
		return nil, fmt.Errorf("error parsing varying block fields: %w", err)
	}
	if err := p.ParseBlockEnd(); err != nil {
		return nil, fmt.Errorf("error parsing varying block end: %w", err)
	}
	return &VaryingBlockDeclaration{
		Fields: fields,
	}, nil
}

func (p *Parser) ParseExpression() (Expression, error) {
	valStack := ds.NewStack[Expression](2)
	opStack := ds.NewStack[string](1)

	value, err := p.parseExpressionValue()
	if err != nil {
		return nil, fmt.Errorf("error parsing value expression: %w", err)
	}
	valStack.Push(value)

	nextToken := p.peekToken()
	for nextToken.IsBinaryOperator() {
		operatorToken := p.nextToken()
		if !operatorToken.IsBinaryOperator() {
			return nil, fmt.Errorf("expected binary operator")
		}

		operator := operatorToken.Value
		operatorPrio := operatorPriority(operator)
		for !opStack.IsEmpty() {
			oldOperator := opStack.Peek()
			oldOperatorPrio := operatorPriority(oldOperator)
			if oldOperatorPrio <= operatorPrio {
				break
			}
			opStack.Pop() // pop it
			right := valStack.Pop()
			left := valStack.Pop()
			valStack.Push(&BinaryExpression{
				Operator: oldOperator,
				Left:     left,
				Right:    right,
			})
		}
		opStack.Push(operator)

		nextToken = p.peekToken()
		switch {
		case nextToken.IsNewLine():
			if err := p.ParseNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}
		case nextToken.IsComment():
			if err := p.ParseComment(); err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}
		}

		valueToken, err := p.parseExpressionValue()
		if err != nil {
			return nil, fmt.Errorf("error parsing value expression: %w", err)
		}
		valStack.Push(valueToken)

		nextToken = p.peekToken()
	}

	for valStack.Size() > 1 {
		right := valStack.Pop()
		left := valStack.Pop()
		if opStack.IsEmpty() {
			return nil, fmt.Errorf("no operator found for binary expression")
		}
		operator := opStack.Pop()
		valStack.Push(&BinaryExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
		})
	}
	if valStack.IsEmpty() {
		return nil, fmt.Errorf("no value expressions found")
	}
	return valStack.Pop(), nil
}

func (p *Parser) ParseFunction() (*FunctionDeclaration, error) {
	var decl FunctionDeclaration

	funcToken := p.nextToken()
	if !funcToken.IsSpecificIdentifier("func") {
		return nil, fmt.Errorf("expected func keyword")
	}

	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return nil, fmt.Errorf("expected function name identifier")
	}
	decl.Name = nameToken.Value

	paramBracketStart := p.nextToken()
	if !paramBracketStart.IsSpecificOperator("(") {
		return nil, fmt.Errorf("expected opening bracket")
	}

	inputParams, err := p.ParseNamedParameterList()
	if err != nil {
		return nil, fmt.Errorf("error parsing input params: %w", err)
	}
	decl.Inputs = inputParams

	paramBracketEnd := p.nextToken()
	if !paramBracketEnd.IsSpecificOperator(")") {
		return nil, fmt.Errorf("expected closing bracket")
	}

	nextToken := p.peekToken()
	if nextToken.IsSpecificOperator("(") {
		resultBracketStart := p.nextToken()
		if !resultBracketStart.IsSpecificOperator("(") {
			return nil, fmt.Errorf("expected opening bracket")
		}

		outputParams, err := p.ParseUnnamedParameterList()
		if err != nil {
			return nil, fmt.Errorf("error parsing output params: %w", err)
		}
		decl.Outputs = outputParams

		resultBracketEnd := p.nextToken()
		if !resultBracketEnd.IsSpecificOperator(")") {
			return nil, fmt.Errorf("expected closing bracket")
		}
	}

	if err := p.ParseBlockStart(); err != nil {
		return nil, fmt.Errorf("error parsing function block start: %w", err)
	}

	statements, err := p.ParseStatementList()
	if err != nil {
		return nil, fmt.Errorf("error parsing function body: %w", err)
	}
	decl.Body = statements

	if err := p.ParseBlockEnd(); err != nil {
		return nil, fmt.Errorf("error parsing function footer: %w", err)
	}
	return &decl, nil
}

func (p *Parser) ParseStatementList() ([]Statement, error) {
	var statements []Statement
	for {
		token := p.peekToken()
		switch {
		case token.IsNewLine():
			if err := p.ParseNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}
		case token.IsComment():
			if err := p.ParseComment(); err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}
		case token.IsIdentifier():
			statement, err := p.ParseStatement()
			if err != nil {
				return nil, fmt.Errorf("error parsing statement: %w", err)
			}
			statements = append(statements, statement)
		default:
			return statements, nil
		}
	}
}

func (p *Parser) ParseStatement() (Statement, error) {
	token := p.peekToken()
	switch {
	case token.IsSpecificIdentifier("var"):
		return p.parseVariableDeclaration()
	case token.IsSpecificIdentifier("if"):
		return p.parseConditionalStatement()
	case token.IsSpecificIdentifier("discard"):
		return p.parseDiscardStatement()
	case token.IsIdentifier():
		return p.parseImperativeStatement()
	default:
		return nil, fmt.Errorf("unexpected token: %v", token)
	}
}

func (p *Parser) parseVariableDeclaration() (*VariableDeclaration, error) {
	var decl VariableDeclaration
	varToken := p.nextToken()
	if !varToken.IsSpecificIdentifier("var") {
		return nil, fmt.Errorf("expected var keyword")
	}
	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return nil, fmt.Errorf("expected identifier")
	}
	decl.Name = nameToken.Value
	typeToken := p.nextToken()
	if !typeToken.IsIdentifier() {
		return nil, fmt.Errorf("expected identifier")
	}
	decl.Type = typeToken.Value

	nextToken := p.peekToken()
	if nextToken.IsAssignmentOperator() {
		assignToken := p.nextToken()
		if !assignToken.IsOperator() {
			return nil, fmt.Errorf("expected operator")
		}
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, fmt.Errorf("error parsing expression: %w", err)
		}
		decl.Assignment = expr
		if err := p.ParseOptionalRemainder(); err != nil {
			return nil, fmt.Errorf("error parsing end of line: %w", err)
		}
	} else {
		if err := p.ParseOptionalRemainder(); err != nil {
			return nil, fmt.Errorf("error parsing end of line: %w", err)
		}
	}
	return &decl, nil
}

func (p *Parser) parseConditionalStatement() (*Conditional, error) {
	var statement Conditional

	ifToken := p.nextToken()
	if !ifToken.IsSpecificIdentifier("if") {
		return nil, fmt.Errorf("expected if keyword")
	}

	conditionExpression, err := p.ParseExpression()
	if err != nil {
		return nil, fmt.Errorf("error parsing condition expression: %w", err)
	}
	statement.Condition = conditionExpression

	if err := p.ParseBlockStart(); err != nil {
		return nil, fmt.Errorf("error parsing block start: %w", err)
	}

	thenStatements, err := p.ParseStatementList()
	if err != nil {
		return nil, fmt.Errorf("error parsing function body: %w", err)
	}
	statement.Then = thenStatements

	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator("}") {
		return nil, fmt.Errorf("expected closing bracket")
	}

	nextToken := p.peekToken()
	if nextToken.IsSpecificIdentifier("else") {
		_ = p.nextToken() // consume else token
		nextToken = p.peekToken()
		if nextToken.IsSpecificIdentifier("if") {
			elseIfConditional, err := p.parseConditionalStatement()
			if err != nil {
				return nil, fmt.Errorf("error parsing else if conditional: %w", err)
			}
			statement.ElseIf = elseIfConditional
		} else {
			if err := p.ParseBlockStart(); err != nil {
				return nil, fmt.Errorf("error parsing block start: %w", err)
			}

			elseStatements, err := p.ParseStatementList()
			if err != nil {
				return nil, fmt.Errorf("error parsing function body: %w", err)
			}
			statement.Else = elseStatements

			if err := p.ParseBlockEnd(); err != nil {
				return nil, fmt.Errorf("error parsing block end: %w", err)
			}
		}
	} else {
		if err := p.ParseOptionalRemainder(); err != nil {
			return nil, fmt.Errorf("error parsing end of line: %w", err)
		}
	}

	return &statement, nil
}

func (p *Parser) parseDiscardStatement() (*Discard, error) {
	discardToken := p.nextToken()
	if !discardToken.IsSpecificIdentifier("discard") {
		return nil, fmt.Errorf("expected discard keyword")
	}
	if err := p.ParseOptionalRemainder(); err != nil {
		return nil, fmt.Errorf("error parsing end of line: %w", err)
	}
	return &Discard{}, nil
}

func (p *Parser) parseImperativeStatement() (Statement, error) {
	identifierToken := p.nextToken()
	if !identifierToken.IsIdentifier() {
		return nil, fmt.Errorf("expected identifier")
	}

	var target Expression
	nextToken := p.peekToken()
	if nextToken.IsSpecificOperator(".") {
		p.nextToken() // consume the dot
		fieldToken := p.nextToken()
		if !fieldToken.IsIdentifier() {
			return nil, fmt.Errorf("expected identifier")
		}
		target = &FieldIdentifier{
			ObjName:   identifierToken.Value,
			FieldName: fieldToken.Value,
		}
		nextToken = p.peekToken()
	} else {
		target = &Identifier{
			Name: identifierToken.Value,
		}
	}

	switch {
	case nextToken.IsSpecificOperator("("):
		openingToken := p.nextToken()
		if !openingToken.IsSpecificOperator("(") {
			return nil, fmt.Errorf("expected opening bracket")
		}
		fields, err := p.parseArguments()
		if err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}
		closingToken := p.nextToken()
		if !closingToken.IsSpecificOperator(")") {
			return nil, fmt.Errorf("expected closing bracket")
		}
		if err := p.ParseOptionalRemainder(); err != nil {
			return nil, fmt.Errorf("error parsing end of line: %w", err)
		}
		return &FunctionCall{
			Name:      identifierToken.Value,
			Arguments: fields,
		}, nil

	case nextToken.IsAssignmentOperator():
		operatorToken := p.nextToken()
		if !operatorToken.IsAssignmentOperator() {
			return nil, fmt.Errorf("expected assignment operator")
		}
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, fmt.Errorf("error parsing expression: %w", err)
		}
		if err := p.ParseOptionalRemainder(); err != nil {
			return nil, fmt.Errorf("error parsing end of line: %w", err)
		}
		return &Assignment{
			Target:     target,
			Operator:   operatorToken.Value,
			Expression: expr,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected token: %v", nextToken)
	}
}

func operatorPriority(operator string) int {
	// TODO: Add more...
	switch operator {
	case "<<", ">>":
		return 3
	case "*", "/":
		return 2
	case "+", "-":
		return 1
	default:
		return 0
	}
}

func (p *Parser) parseExpressionValue() (Expression, error) {
	token := p.peekToken()
	switch {
	case token.IsSpecificOperator("("):
		return p.parseExpressionGroup()

	case token.IsUnaryOperator():
		return p.parseUnaryExpression()

	case token.IsNumber():
		return p.parseNumericExpression()

	case token.IsIdentifier():
		return p.parseIdentifierExpression()

	default:
		return nil, fmt.Errorf("unexpected token: %v", token)
	}
}

func (p *Parser) parseExpressionGroup() (Expression, error) {
	openingToken := p.nextToken()
	if !openingToken.IsSpecificOperator("(") {
		return nil, fmt.Errorf("expected opening bracket")
	}
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, fmt.Errorf("error parsing expression: %w", err)
	}
	closingToken := p.nextToken()
	if !closingToken.IsSpecificOperator(")") {
		return nil, fmt.Errorf("expected closing bracket")
	}
	return expr, nil
}

func (p *Parser) parseUnaryExpression() (Expression, error) {
	operatorToken := p.nextToken()
	if !operatorToken.IsUnaryOperator() {
		return nil, fmt.Errorf("expected unary operator")
	}
	expr, err := p.parseExpressionValue()
	if err != nil {
		return nil, fmt.Errorf("error parsing value expression: %w", err)
	}
	return &UnaryExpression{
		Operator: operatorToken.Value,
		Operand:  expr,
	}, nil
}

func (p *Parser) parseNumericExpression() (Expression, error) {
	numberToken := p.nextToken()
	intValue, err := strconv.ParseInt(numberToken.Value, 10, 64)
	if err == nil {
		return &IntLiteral{
			Value: intValue,
		}, nil
	}
	floatValue, err := strconv.ParseFloat(numberToken.Value, 64)
	if err == nil {
		return &FloatLiteral{
			Value: floatValue,
		}, nil
	}
	return nil, fmt.Errorf("error parsing number: %w", err)
}

func (p *Parser) parseIdentifierExpression() (Expression, error) {
	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return nil, fmt.Errorf("expected identifier")
	}

	nextToken := p.peekToken()
	switch {
	case nextToken.IsSpecificOperator("("):
		openingToken := p.nextToken()
		if !openingToken.IsSpecificOperator("(") {
			return nil, fmt.Errorf("expected opening bracket")
		}

		args, err := p.parseArguments()
		if err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}

		closingToken := p.nextToken()
		if !closingToken.IsSpecificOperator(")") {
			return nil, fmt.Errorf("expected closing bracket")
		}

		return &FunctionCall{
			Name:      nameToken.Value,
			Arguments: args,
		}, nil

	case nextToken.IsSpecificOperator("."):
		dotToken := p.nextToken()
		if !dotToken.IsOperator() {
			return nil, fmt.Errorf("expected dot operator")
		}
		fieldToken := p.nextToken()
		if !fieldToken.IsIdentifier() {
			return nil, fmt.Errorf("expected identifier")
		}
		return &FieldIdentifier{
			ObjName:   nameToken.Value,
			FieldName: fieldToken.Value,
		}, nil

	default:
		return &Identifier{
			Name: nameToken.Value,
		}, nil
	}
}

func (p *Parser) parseArguments() ([]Expression, error) {
	var args []Expression
	token := p.peekToken()
	for !token.IsSpecificOperator(")") {
		switch {
		case token.IsNewLine():
			if err := p.ParseNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}
		case token.IsComment():
			if err := p.ParseComment(); err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}
		default:
			arg, err := p.ParseExpression()
			if err != nil {
				return nil, fmt.Errorf("error parsing expression: %w", err)
			}
			args = append(args, arg)

			nextToken := p.peekToken()
			switch {
			case nextToken.IsSpecificOperator(","):
				p.nextToken()
			case nextToken.IsSpecificOperator(")"):
				// Do nothing
			default:
				return nil, fmt.Errorf("unexpected token: %v", nextToken)
			}
		}
		token = p.peekToken()
	}
	return args, nil
}
