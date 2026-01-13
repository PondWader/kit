package lang

import (
	"bufio"
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
	ErrExportNotAtTopLevel       = errors.New("all export statements must be declared at the top level of the module")
	ErrUnexpectedToken           = errors.New("unexpected token encountered")
	ErrUnterminatedString        = errors.New("unterminated string literal")
	ErrExportMustHaveDeclaration = errors.New("an export statement must be followed by a declaration")
	ErrAssignmentNotAllowed      = errors.New("assignment not allowed")
	ErrCallAtTopLevel            = errors.New("functions cannot be called at the top level of the program")
	ErrMissingLambdaArg          = errors.New("missing lamdba arg name")
)

func fmtUnexpectedToken(expected []tokens.TokenKind, got tokens.Token) error {
	if len(expected) == 0 {
		return fmt.Errorf("%w: got %s", ErrUnexpectedToken, got)
	}
	return fmt.Errorf("%w: got %s but expected %s", ErrUnexpectedToken, got, join(expected, " or "))
}

func Parse(r io.Reader) ([]Node, error) {
	br := bufio.NewReader(r)
	l := tokens.NewLexer(br)
	p := parser{0, l, br, -1}
	return p.parseProgram()
}

type parser struct {
	blockDepth   int
	l            *tokens.Lexer
	r            *bufio.Reader
	newLineState int
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
	if token.Kind == tokens.TokenKindNewLine || token.Kind == tokens.TokenKindWhitespace {
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
	if token.Kind == tokens.TokenKindNewLine || token.Kind == tokens.TokenKindSemicolon || token.Kind == tokens.TokenKindWhitespace {
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
	case tokens.TokenKindFunction:
		return p.parseFunction()
	case tokens.TokenKindReturn:
		return p.parseReturn()
	default:
		return p.parseExpressionFromToken(tok)
	}
}

func (p *parser) parseExport() (n NodeExport, err error) {
	if p.blockDepth != 0 {
		return n, ErrExportNotAtTopLevel
	}

	ident, err := p.expectToken(tokens.TokenKindIdentifier, tokens.TokenKindFunction)
	if err != nil {
		return n, err
	}
	node, err := p.parseStatementFromToken(ident)
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

func (p *parser) parseReturn() (n NodeReturn, err error) {
	n.Val, err = p.parseExpression()
	return n, err
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
	case tokens.TokenKindDoubleQuote:
		node, err = p.parseString('"')
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
		case tokens.TokenKindLeftParen:
			node, err = p.parseCallExpression(node)
		case tokens.TokenKindArrow:
			node, err = p.parseLambda(node)
		default:
			p.l.Unread(next)
			return node, nil
		}

		if err != nil {
			return nil, err
		}
	}
}

func (p *parser) parseCallExpression(fn Node) (n NodeCall, err error) {
	if p.blockDepth == 0 {
		return n, ErrCallAtTopLevel
	}

	n.Fn = fn

	tok, err := p.next()
	if err != nil {
		return n, err
	} else if tok.Kind != tokens.TokenKindRightParen {
		n.Arg, err = p.parseExpressionFromToken(tok)
		if err != nil {
			return n, err
		}
		_, err := p.expectToken(tokens.TokenKindRightParen)
		if err != nil {
			return n, err
		}
	}
	return
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

func (p *parser) parseLambda(arg Node) (n NodeFunction, err error) {
	ident, ok := arg.(NodeIdentifier)
	if !ok {
		return n, ErrMissingLambdaArg
	}
	n.ArgName = ident.Ident
	n.Body, err = p.parseExpression()
	return
}

func (p *parser) parseFunction() (Node, error) {
	// Optional function name, if so it's a declaration otherwise it's an anonymous function
	start, err := p.expectToken(tokens.TokenKindIdentifier, tokens.TokenKindLeftParen)
	if err != nil {
		return nil, err
	}
	if start.Kind == tokens.TokenKindIdentifier {
		if _, err = p.expectToken(tokens.TokenKindLeftParen); err != nil {
			return nil, err
		}
	}

	var fn NodeFunction
	tok, err := p.expectToken(tokens.TokenKindIdentifier, tokens.TokenKindRightParen)
	if err != nil {
		return nil, err
	} else if tok.Kind == tokens.TokenKindIdentifier {
		fn.ArgName = tok.Literal
		if _, err := p.expectToken(tokens.TokenKindRightParen); err != nil {
			return nil, err
		}
	}

	tok, err = p.expectToken(tokens.TokenKindLeftBrace, tokens.TokenKindArrow)
	if err != nil {
		return nil, err
	} else if tok.Kind == tokens.TokenKindLeftBrace {
		b, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		b.IsFunctionBody = true
		fn.Body = b
	} else {
		// Although it's not actually a block, we want to allow function calls
		p.blockDepth++
		b, err := p.parseExpression()
		p.blockDepth--
		if err != nil {
			return nil, err
		}
		fn.Body = b
	}

	if start.Kind == tokens.TokenKindIdentifier {
		return NodeDeclaration{Name: start.Literal, Value: fn}, nil
	}
	return fn, nil
}

func (p *parser) parseBlock() (NodeBlock, error) {
	p.blockDepth++
	defer func() {
		p.blockDepth--
	}()

	b := NodeBlock{
		Body: make([]Node, 0, 16),
	}

	for {
		first, err := p.nextStatementToken()
		if err != nil {
			return b, err
		}
		if first.Kind == tokens.TokenKindRightBrace {
			return b, err
		}

		stmt, err := p.parseStatementFromToken(first)
		if err != nil {
			return b, err
		}
		b.Body = append(b.Body, stmt)

		if p.newLineState == p.l.State {
			continue
		}
		next, err := p.expectToken(tokens.TokenKindRightBrace, tokens.TokenKindEOF, tokens.TokenKindNewLine, tokens.TokenKindSemicolon)
		if err != nil {
			return b, err
		}
		if next.Kind == tokens.TokenKindRightBrace {
			return b, err
		}
	}
}

// Next returns the next token from the lexer, skipping comments and recording new lines
func (p *parser) next() (tokens.Token, error) {
	for {
		inNewLine := p.newLineState == p.l.State
		token, err := p.l.Next()
		if err != nil {
			return token, err
		}
		if token.Kind == tokens.TokenKindSingleLineComment || token.Kind == tokens.TokenKindMultiLineComment {
			continue
		} else if token.Kind == tokens.TokenKindNewLine || (token.Kind == tokens.TokenKindWhitespace && inNewLine) {
			p.newLineState = p.l.State
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
		if token.Kind != tokens.TokenKindWhitespace && token.Kind != tokens.TokenKindNewLine {
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
