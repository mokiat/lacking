package lsl

import (
	"fmt"
	"strconv"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/graphics/lsl/internal"
)

func Parse(source string) (*Shader, error) {
	// FIXME: Switch with Parse2
	return &Shader{
		Declarations: []Declaration{
			&UniformBlockDeclaration{
				Fields: []Field{
					{
						Name: "color",
						Type: TypeNameVec4,
					},
				},
			},

			&FunctionDeclaration{
				Name: "#fragment",
				Body: []Statement{
					&Assignment{
						Target: "#color",
						Expression: &Identifier{
							Name: "color",
						},
					},
				},
			},
		},
	}, nil
}

func Parse2(source string) (*Shader, error) {
	return NewParser(source).ParseShader()
}

func NewParser(source string) *Parser {
	tokenizer := internal.NewTokenizer(source)
	return &Parser{
		tokenizer: tokenizer,
		token:     tokenizer.Next(),
	}
}

type Parser struct {
	tokenizer *internal.Tokenizer
	token     internal.Token
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
		case token.IsSpecificIdentifier("#uniform"):
			decl, err := p.ParseUniformBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing uniform block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier("#varying"):
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

func (p *Parser) peekToken() internal.Token {
	return p.token
}

func (p *Parser) nextToken() internal.Token {
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

func (p *Parser) ParseUniformBlock() (*UniformBlockDeclaration, error) {
	uniformToken := p.nextToken()
	if !uniformToken.IsSpecificIdentifier("#uniform") {
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
	if !varyingToken.IsSpecificIdentifier("#varying") {
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

func (p *Parser) ParseFunction() (*FunctionDeclaration, error) {
	var decl FunctionDeclaration
	if err := p.parseFunctionHeader(&decl); err != nil {
		return nil, fmt.Errorf("error parsing function header: %w", err)
	}
	if err := p.parseFunctionBody(&decl); err != nil {
		return nil, fmt.Errorf("error parsing function body: %w", err)
	}
	if err := p.ParseBlockEnd(); err != nil {
		return nil, fmt.Errorf("error parsing function footer: %w", err)
	}
	return &decl, nil
}

func (p *Parser) parseFunctionHeader(decl *FunctionDeclaration) error {
	funcToken := p.nextToken()
	if !funcToken.IsSpecificIdentifier("func") {
		return fmt.Errorf("expected func keyword")
	}

	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return fmt.Errorf("expected function name identifier")
	}
	decl.Name = nameToken.Value

	paramOpeningBracketToken := p.nextToken()
	if !paramOpeningBracketToken.IsSpecificOperator("(") {
		return fmt.Errorf("expected opening bracket")
	}

	inputParams, err := p.ParseNamedParameterList()
	if err != nil {
		return fmt.Errorf("error parsing input params: %w", err)
	}
	decl.Inputs = inputParams

	paramClosingBracketToken := p.nextToken()
	if !paramClosingBracketToken.IsSpecificOperator(")") {
		return fmt.Errorf("expected closing bracket")
	}

	nextToken := p.peekToken()
	if !nextToken.IsSpecificOperator("{") {
		switch {
		case nextToken.IsSpecificOperator("("):
			resultOpeningBracketToken := p.nextToken()
			if !resultOpeningBracketToken.IsSpecificOperator("(") {
				return fmt.Errorf("expected opening bracket")
			}
			outputParams, err := p.ParseUnnamedParameterList()
			if err != nil {
				return fmt.Errorf("error parsing unnamed params: %w", err)
			}
			decl.Outputs = outputParams
			resultClosingBracketToken := p.nextToken()
			if !resultClosingBracketToken.IsSpecificOperator(")") {
				return fmt.Errorf("expected closing bracket")
			}
		case nextToken.IsIdentifier():
			typeToken := p.nextToken()
			decl.Outputs = []Field{
				{
					Type: typeToken.Value,
				},
			}
		default:
			return fmt.Errorf("unexpected token: %v", nextToken)
		}
	}

	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator("{") {
		return fmt.Errorf("expected opening bracket")
	}

	if err := p.ParseOptionalRemainder(); err != nil {
		return fmt.Errorf("error parsing end of line: %w", err)
	}
	return nil
}

func (p *Parser) parseFunctionBody(decl *FunctionDeclaration) error {
	token := p.peekToken()
	for !token.IsSpecificOperator("}") {
		switch {
		case token.IsNewLine():
			if err := p.ParseNewLine(); err != nil {
				return fmt.Errorf("error parsing new line: %w", err)
			}
		case token.IsComment():
			if err := p.ParseComment(); err != nil {
				return fmt.Errorf("error parsing comment: %w", err)
			}
		case token.IsIdentifier():
			statement, err := p.parseStatement()
			if err != nil {
				return fmt.Errorf("error parsing statement: %w", err)
			}
			decl.Body = append(decl.Body, statement)
		default:
			return fmt.Errorf("unexpected token: %v", token)
		}
		token = p.peekToken()
	}
	return nil
}

func (p *Parser) parseStatement() (Statement, error) {
	token := p.peekToken()
	switch {
	case token.IsSpecificIdentifier("var"):
		return p.parseVariableDeclaration()
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
		expr, err := p.parseExpression()
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

func (p *Parser) parseImperativeStatement() (Statement, error) {
	// assignToken := p.peekToken()
	// if assignToken.IsSpecificOperator("=") {
	// 	return p.parseAssignment()
	// }
	// return nil, fmt.Errorf("unexpected token: %v", assignToken)
	panic("TODO")
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

func (p *Parser) parseExpression() (Expression, error) {
	values := ds.NewStack[Expression](2)
	operators := ds.NewStack[string](1)

	expectValue := true
	expectMoreTokens := true
	for expectMoreTokens {
		if expectValue {
			value, err := p.parseValueExpression()
			if err != nil {
				return nil, fmt.Errorf("error parsing value expression: %w", err)
			}
			values.Push(value)

			nextToken := p.peekToken()
			if nextToken.IsNewLine() {
				if err := p.ParseNewLine(); err != nil {
					return nil, fmt.Errorf("error parsing new line: %w", err)
				}
				expectMoreTokens = false
			}
		} else {
			operatorToken := p.nextToken()
			if !operatorToken.IsOperator() {
				return nil, fmt.Errorf("expected operator")
			}
			if operatorToken.IsSpecificOperator(")") {
				break
			}
			if operatorToken.IsSpecificOperator(",") {
				break
			}

			newOperator := operatorToken.Value
			newOperatorPrio := operatorPriority(newOperator)

			if !operators.IsEmpty() {
				oldOperator := operators.Peek()
				oldOperatorPrio := operatorPriority(oldOperator)

				for oldOperatorPrio > newOperatorPrio {
					right := values.Pop()
					left := values.Pop()
					operator := operators.Pop()
					values.Push(&BinaryExpression{
						Left:     left,
						Operator: operator,
						Right:    right,
					})
					if operators.IsEmpty() {
						break
					}
					oldOperator = operators.Peek()
					oldOperatorPrio = operatorPriority(oldOperator)
				}
			}

			operators.Push(newOperator)

			nextToken := p.peekToken()
			switch {
			case nextToken.IsComment():
				if err := p.ParseComment(); err != nil {
					return nil, fmt.Errorf("error parsing comment: %w", err)
				}
			case nextToken.IsNewLine():
				if err := p.ParseNewLine(); err != nil {
					return nil, fmt.Errorf("error parsing new line: %w", err)
				}
			}
		}
		expectValue = !expectValue
	}

	for values.Size() > 1 {
		right := values.Pop()
		left := values.Pop()
		if operators.IsEmpty() {
			return nil, fmt.Errorf("no operator found for binary expression")
		}
		operator := operators.Pop()
		values.Push(&BinaryExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
		})
	}
	if values.IsEmpty() {
		return nil, fmt.Errorf("no value expressions found")
	}
	return values.Pop(), nil
}

func (p *Parser) parseValueExpression() (Expression, error) {
	token := p.peekToken()
	switch {
	case token.IsOperator():
		operatorToken := p.nextToken()
		if !operatorToken.IsUnaryOperator() {
			return nil, fmt.Errorf("expected unary operator")
		}
		expr, err := p.parseValueExpression()
		if err != nil {
			return nil, fmt.Errorf("error parsing value expression: %w", err)
		}
		return &UnaryExpression{
			Operator: operatorToken.Value,
			Operand:  expr,
		}, nil

	case token.IsNumber():
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
	case token.IsIdentifier():
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

	default:
		return nil, fmt.Errorf("unexpected token: %v", token)
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
			arg, err := p.parseExpression()
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
