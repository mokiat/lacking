package lsl

import (
	"errors"
	"fmt"
	"io"

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
	return newParser(source).Parse()
}

func newParser(source string) *parser {
	tokenizer := internal.NewTokenizer(source)
	return &parser{
		tokenizer: tokenizer,
		token:     tokenizer.Next(),
	}
}

type parser struct {
	tokenizer *internal.Tokenizer
	token     internal.Token
}

func (p *parser) Parse() (*Shader, error) {
	shader, err := p.parseShader()
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("error parsing top-level declarations: %w", err)
	}
	return shader, nil
}

func (p *parser) parseShader() (*Shader, error) {
	var shader Shader
	token := p.peekToken()
	for !token.IsEOF() {
		switch {
		case token.IsNewLine():
			if err := p.parseNewLine(); err != nil {
				return nil, fmt.Errorf("error parsing new line: %w", err)
			}
		case token.IsComment():
			decl, err := p.parseComment()
			if err != nil {
				return nil, fmt.Errorf("error parsing comment: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier("uniform"):
			decl, err := p.parseUniformBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing uniform block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		case token.IsSpecificIdentifier("varying"):
			decl, err := p.parseVaryingBlock()
			if err != nil {
				return nil, fmt.Errorf("error parsing varying block: %w", err)
			}
			shader.Declarations = append(shader.Declarations, decl)
		default:
			return nil, fmt.Errorf("unexpected token: %v", token)
		}
		token = p.peekToken()
	}
	return &shader, nil
}

func (p *parser) peekToken() internal.Token {
	return p.token
}

func (p *parser) nextToken() internal.Token {
	token := p.token
	p.token = p.tokenizer.Next()
	return token
}

func (p *parser) parseNewLine() error {
	token := p.nextToken()
	if !token.IsEOF() && !token.IsNewLine() {
		return fmt.Errorf("expected new line")
	}
	return nil
}

func (p *parser) parseEndOfLine() error {
	token := p.peekToken()
	switch {
	case token.IsEOF():
		return p.parseNewLine()
	case token.IsNewLine():
		return p.parseNewLine()
	case token.IsComment():
		_, err := p.parseComment()
		return err
	default:
		return fmt.Errorf("expected new line or comment")
	}
}

func (p *parser) parseComment() (*CommentDeclaration, error) {
	commentToken := p.nextToken()
	if !commentToken.IsComment() {
		return nil, fmt.Errorf("expected comment token")
	}
	if err := p.parseNewLine(); err != nil {
		return nil, fmt.Errorf("error parsing new line after comment: %w", err)
	}
	return &CommentDeclaration{
		Comment: commentToken.Value,
	}, nil
}

func (p *parser) parseUniformBlock() (*UniformBlockDeclaration, error) {
	if err := p.parseUniformBlockHeader(); err != nil {
		return nil, fmt.Errorf("error parsing uniform block header: %w", err)
	}
	decl, err := p.parseUniformBlockBody()
	if err != nil {
		return nil, fmt.Errorf("error parsing uniform block body: %w", err)
	}
	if err := p.parseUniformBlockFooter(); err != nil {
		return nil, fmt.Errorf("error parsing uniform block footer: %w", err)
	}
	return decl, nil
}

func (p *parser) parseUniformBlockHeader() error {
	uniformToken := p.nextToken()
	if !uniformToken.IsSpecificIdentifier("uniform") {
		return fmt.Errorf("expected uniform keyword")
	}
	bracketToken := p.nextToken()
	if !bracketToken.IsOperator() || bracketToken.Value != "{" {
		return fmt.Errorf("expected opening bracket")
	}
	if err := p.parseEndOfLine(); err != nil {
		return fmt.Errorf("error parsing end of line: %w", err)
	}
	return nil
}

func (p *parser) parseUniformBlockBody() (*UniformBlockDeclaration, error) {
	var decl UniformBlockDeclaration
	token := p.peekToken()
	for !token.IsSpecificOperator("}") {
		field, err := p.parseField()
		if err != nil {
			return nil, fmt.Errorf("error parsing field: %w", err)
		}
		decl.Fields = append(decl.Fields, field)
		token = p.peekToken()
	}
	return &decl, nil
}

func (p *parser) parseUniformBlockFooter() error {
	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator("}") {
		return fmt.Errorf("expected closing bracket")
	}
	if err := p.parseEndOfLine(); err != nil {
		return fmt.Errorf("error parsing end of line: %w", err)
	}
	return nil
}

func (p *parser) parseVaryingBlock() (*VaryingBlockDeclaration, error) {
	if err := p.parseVaryingBlockHeader(); err != nil {
		return nil, fmt.Errorf("error parsing varying block header: %w", err)
	}
	decl, err := p.parseVaryingBlockBody()
	if err != nil {
		return nil, fmt.Errorf("error parsing varying block body: %w", err)
	}
	if err := p.parseVaryingBlockFooter(); err != nil {
		return nil, fmt.Errorf("error parsing varying block footer: %w", err)
	}
	return decl, nil
}

func (p *parser) parseVaryingBlockHeader() error {
	varyingToken := p.nextToken()
	if !varyingToken.IsSpecificIdentifier("varying") {
		return fmt.Errorf("expected varying keyword")
	}
	bracketToken := p.nextToken()
	if !bracketToken.IsOperator() || bracketToken.Value != "{" {
		return fmt.Errorf("expected opening bracket")
	}
	if err := p.parseEndOfLine(); err != nil {
		return fmt.Errorf("error parsing end of line: %w", err)
	}
	return nil
}

func (p *parser) parseVaryingBlockBody() (*VaryingBlockDeclaration, error) {
	var decl VaryingBlockDeclaration
	token := p.peekToken()
	for !token.IsSpecificOperator("}") {
		field, err := p.parseField()
		if err != nil {
			return nil, fmt.Errorf("error parsing field: %w", err)
		}
		decl.Fields = append(decl.Fields, field)
		token = p.peekToken()
	}
	return &decl, nil
}

func (p *parser) parseVaryingBlockFooter() error {
	bracketToken := p.nextToken()
	if !bracketToken.IsSpecificOperator("}") {
		return fmt.Errorf("expected closing bracket")
	}
	if err := p.parseEndOfLine(); err != nil {
		return fmt.Errorf("error parsing end of line: %w", err)
	}
	return nil
}

func (p *parser) parseField() (Field, error) {
	nameToken := p.nextToken()
	if !nameToken.IsIdentifier() {
		return Field{}, fmt.Errorf("expected name identifier")
	}
	typeToken := p.nextToken()
	if !typeToken.IsIdentifier() {
		return Field{}, fmt.Errorf("expected type identifier")
	}
	if err := p.parseEndOfLine(); err != nil {
		return Field{}, fmt.Errorf("error parsing end of line: %w", err)
	}
	return Field{
		Name: nameToken.Value,
		Type: typeToken.Value,
	}, nil
}
