package lang

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/PondWader/kit/pkg/lang/tokens"
	"github.com/PondWader/kit/pkg/lang/values"
)

var (
	ErrImportNotAtTopLevel          = errors.New("all import statements must be declared at the top level of the module")
	ErrUnexpectedToken              = errors.New("unexpected token encountered")
	ErrUnterminatedString           = errors.New("unterminated string literal")
	ErrExpectedDeclarationStatement = errors.New("expected a declaration statement")
	ErrExportMustHaveDeclaration    = errors.New("an export statement must be followed by a declaration")
	ErrAssignmentNotAllowed         = errors.New("assignment not allowed")
)

func fmtUnexpectedToken(expected []tokens.TokenKind, got tokens.Token) error {
	if len(expected) == 0 {
		return fmt.Errorf("%w: got %s", ErrUnexpectedToken, got)
	}
	return fmt.Errorf("%w: got %s but expected %s", ErrUnexpectedToken, got, join(expected, " or "))
}

func Parse(r io.Reader) ([]Node, error) {
	l := tokens.NewLexer(r)
	p := parser{l}
	return p.parseProgram()
}

type parser struct {
	l *tokens.Lexer
}

func (p *parser) expectToken(kind ...tokens.TokenKind) (tokens.Token, error) {
	token, err := p.next()
	if err != nil {
		return token, err
	}
	if slices.Contains(kind, token.Kind) {
		return token, nil
	}

	// Ignore white space and new lines since we don't want to be sensitive to it
	if token.Kind == tokens.TokenKindNewline || token.Kind == tokens.TokenKindWhitespace {
		return p.expectToken(kind...)
	}

	return token, fmtUnexpectedToken(kind, token)
}

func (p *parser) parseProgram() ([]Node, error) {
	prog := make([]Node, 0, 32)
	for {
		n, err := p.parseStatement()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return prog, nil
			}
			return nil, err
		}
		prog = append(prog, n)
	}
}

func (p *parser) parseStatement() (Node, error) {
	tok, err := p.nextStatementToken()
	if err != nil {
		return nil, err
	}

	return p.parseStatementFromToken(tok)
}

// Returns the next token that could form a valid statemen (skips new lines, white space and semicolons).
// Returns an io.EOF error upon reaching an EOF token.
func (p *parser) nextStatementToken() (tokens.Token, error) {
	token, err := p.next()
	if err != nil {
		return token, err
	}
	if token.Kind == tokens.TokenKindNewline || token.Kind == tokens.TokenKindSemicolon || token.Kind == tokens.TokenKindWhitespace {
		return p.nextStatementToken()
	}
	if token.Kind == tokens.TokenKindEOF {
		return token, io.EOF
	}
	return token, err
}

func (p *parser) parseStatementFromToken(tok tokens.Token) (n Node, err error) {
	defer func() {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
	}()

	switch tok.Kind {
	case tokens.TokenKindExport:
		return p.parseExport()
	default:
		return p.parseExpressionFromToken(tok)
	}
}

func (p *parser) parseExport() (n NodeExport, err error) {
	ident, err := p.expectToken(tokens.TokenKindIdentifier)
	if err != nil {
		return n, err
	}
	node, err := p.parseExpressionFromToken(ident)
	if err != nil {
		return n, err
	}
	decl, ok := node.(NodeDeclaration)
	if !ok {
		return n, ErrExportMustHaveDeclaration
	}
	n.Decl = decl
	return
}

func (p *parser) parseExpression() (Node, error) {
	token, err := p.nextAfterWhiteSpace()
	if err != nil {
		return nil, err
	}
	return p.parseExpressionFromToken(token)
}

func (p *parser) parseExpressionFromToken(tok tokens.Token) (Node, error) {
	var node Node
	var err error
	switch tok.Kind {
	case tokens.TokenKindIdentifier:
		node = NodeIdentifier{Ident: tok.Literal}
	case tokens.TokenKindNumberLiteral:
		var num float64
		num, err = strconv.ParseFloat(tok.Literal, 64)
		node = NodeLiteral{Value: values.Of(num)}
	default:
		return nil, fmtUnexpectedToken(nil, tok)
	}

	if err != nil {
		return node, err
	}
	return p.parseOperation(node)
}

func (p *parser) parseOperation(node Node) (Node, error) {
	for {
		next, err := p.nextAfterWhiteSpace()
		if err != nil {
			return nil, err
		}

		switch next.Kind {
		case tokens.TokenKindAssign:
			node, err = p.parseAssignment(node)
		case tokens.TokenKindDot:
			node, err = p.parseKeyAccess(node)
		default:
			p.l.Unread(next)
			return node, nil
		}

		if err != nil {
			return nil, err
		}
	}
}

func (p *parser) parseAssignment(node Node) (Node, error) {
	right, err := p.parseExpression()
	if err != nil {
		return NodeDeclaration{}, err
	}

	switch node := node.(type) {
	case NodeIdentifier:
		return NodeDeclaration{
			Name:  node.Ident,
			Value: right,
		}, nil
	default:
		return nil, ErrAssignmentNotAllowed
	}
}

func (p *parser) parseKeyAccess(node Node) (n NodeKeyAccess, err error) {
	// Check if accessing a private field (#property)
	next, err := p.expectToken(tokens.TokenKindIdentifier)
	if err != nil {
		return n, err
	}

	n.Val = node
	n.Key = next.Literal
	return
}

// Next returns the next token from the lexer, skipping comments
func (p *parser) next() (tokens.Token, error) {
	for {
		token, err := p.l.Next()
		if err != nil {
			return token, err
		}
		if token.Kind == tokens.TokenKindSingleLineComment || token.Kind == tokens.TokenKindMultiLineComment {
			continue
		}
		return token, nil
	}
}

func (p *parser) nextAfterWhiteSpace() (tokens.Token, error) {
	for {
		token, err := p.next()
		if err != nil {
			return token, err
		}
		if token.Kind != tokens.TokenKindWhitespace {
			return token, err
		}
	}
}

func join[T fmt.Stringer](items []T, sep string) string {
	elems := make([]string, len(items))
	for i, item := range items {
		elems[i] = item.String()
	}
	return strings.Join(elems, sep)
}
