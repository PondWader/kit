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
	ErrImportNotAtTopLevel       = errors.New("all import statements must be declared at the top level of the module")
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
	case tokens.TokenKindImport:
		return p.parseImport()
	case tokens.TokenKindFunction:
		return p.parseFunction()
	case tokens.TokenKindReturn:
		return p.parseReturn()
	case tokens.TokenKindIf:
		return p.parseIf()
	case tokens.TokenKindFor:
		return p.parseFor()
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

func (p *parser) parseImport() (n NodeImport, err error) {
	if p.blockDepth != 0 {
		return n, ErrImportNotAtTopLevel
	}

	for {
		modName, err := p.parsePureString()
		if err != nil {
			return n, err
		}
		n.Modules = append(n.Modules, modName)

		tok, err := p.nextAfterWhiteSpace()
		if err != nil {
			return n, err
		}
		if tok.Kind != tokens.TokenKindComma {
			p.l.Unread(tok)
			return n, nil
		}
	}
}

func (p *parser) parseReturn() (n NodeReturn, err error) {
	n.Val, err = p.parseExpression()
	return n, err
}

func (p *parser) parseIf() (n NodeIf, err error) {
	n.Condition, err = p.parseExpression()
	if err != nil {
		return
	}

	next, err := p.nextAfterWhiteSpace()
	if err != nil {
		return n, err
	} else if next.Kind == tokens.TokenKindLeftBrace {
		n.Body, err = p.parseBlock()
		if err != nil {
			return n, err
		}
	} else {
		n.Body, err = p.parseExpressionFromToken(next)
		if err != nil {
			return n, err
		}
	}

	// Parse else if it exists
	next, err = p.nextAfterWhiteSpace()
	if err != nil {
		return n, err
	} else if next.Kind == tokens.TokenKindElse {
		next, err = p.nextAfterWhiteSpace()

		switch next.Kind {
		case tokens.TokenKindLeftBrace:
			n.Else, err = p.parseBlock()
		case tokens.TokenKindIf:
			n.Else, err = p.parseIf()
		default:
			n.Else, err = p.parseExpressionFromToken(next)
		}
		if err != nil {
			return n, err
		}
	} else {
		p.l.Unread(next)
	}
	return
}

func (p *parser) parseFor() (n NodeForInLoop, err error) {
	ident, err := p.expectToken(tokens.TokenKindIdentifier)
	if err != nil {
		return n, err
	}
	if _, err = p.expectToken(tokens.TokenKindIn); err != nil {
		return n, err
	}
	n.Var = ident.Literal

	iter, err := p.parseExpression()
	if err != nil {
		return n, err
	}
	n.Iterable = iter

	if _, err = p.expectToken(tokens.TokenKindLeftBrace); err != nil {
		return n, err
	}
	b, err := p.parseBlock()
	if err != nil {
		return n, err
	}
	n.Body = b
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
	return p.parseExpressionFromTokenPrec(tok, 0)
}

func (p *parser) parseExpressionFromTokenPrec(tok tokens.Token, minPrec int) (Node, error) {
	var node Node
	var err error
	switch tok.Kind {
	case tokens.TokenKindIdentifier:
		node = NodeIdentifier{Ident: tok.Literal}
	case tokens.TokenKindNumberLiteral:
		var num float64
		num, err = parseNumberLiteral(tok.Literal)
		node = NodeLiteral{Value: values.Of(num)}
	case tokens.TokenKindDoubleQuote:
		node, err = p.parseString('"')
	case tokens.TokenKindLeftSquareBracket:
		node, err = p.parseList()
	case tokens.TokenKindLeftBrace:
		node, err = p.parseObject()
	case tokens.TokenKindTrue:
		node = NodeLiteral{Value: values.Of(true)}
	case tokens.TokenKindFalse:
		node = NodeLiteral{Value: values.Of(false)}
	default:
		return nil, fmtUnexpectedToken(nil, tok)
	}

	if err != nil {
		return node, err
	}
	return p.parseOperation(node, minPrec)
}

func precedenceOf(kind tokens.TokenKind) int {
	switch kind {
	case tokens.TokenKindLogicalOr:
		return 1
	case tokens.TokenKindLogicalAnd:
		return 2
	case tokens.TokenKindEquals, tokens.TokenKindNotEquals:
		return 3
	case tokens.TokenKindLessThan, tokens.TokenKindLessThanOrEqual,
		tokens.TokenKindGreaterThan, tokens.TokenKindGreaterThanOrEqual:
		return 4
	default:
		return 0
	}
}

func (p *parser) parseOperation(node Node, minPrec int) (Node, error) {
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
		case tokens.TokenKindEquals, tokens.TokenKindNotEquals, tokens.TokenKindLessThan, tokens.TokenKindLessThanOrEqual,
			tokens.TokenKindGreaterThan, tokens.TokenKindGreaterThanOrEqual, tokens.TokenKindLogicalAnd, tokens.TokenKindLogicalOr:
			prec := precedenceOf(next.Kind)
			if prec < minPrec {
				p.l.Unread(next)
				return node, nil
			}
			node, err = p.parseBinaryOp(node, next.Kind, prec)
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

func (p *parser) parseList() (n NodeList, err error) {
	for {
		tok, err := p.nextAfterWhiteSpace()
		if err != nil {
			return n, err
		} else if tok.Kind == tokens.TokenKindRightSquareBracket {
			return n, nil
		}

		el, err := p.parseExpressionFromToken(tok)
		if err != nil {
			return n, err
		}
		n.Elements = append(n.Elements, el)

		tok, err = p.expectToken(tokens.TokenKindRightSquareBracket, tokens.TokenKindComma)
		if err != nil {
			return n, err
		} else if tok.Kind == tokens.TokenKindRightSquareBracket {
			return n, nil
		}
	}
}

func (p *parser) parseObject() (NodeObject, error) {
	b, err := p.parseBlock()
	if err != nil {
		return NodeObject{}, err
	}
	return NodeObject{Body: b.Body}, nil
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

func (p *parser) parseBinaryOp(left Node, op tokens.TokenKind, prec int) (Node, error) {
	tok, err := p.nextAfterWhiteSpace()
	if err != nil {
		return nil, err
	}
	right, err := p.parseExpressionFromTokenPrec(tok, prec+1)
	if err != nil {
		return nil, err
	}

	switch op {
	case tokens.TokenKindEquals, tokens.TokenKindNotEquals:
		var n Node = NodeEquals{left, right}
		if op == tokens.TokenKindNotEquals {
			n = NodeNot{Inner: n}
		}
		return n, nil
	case tokens.TokenKindLessThan, tokens.TokenKindLessThanOrEqual:
		return NodeNumberComparison{Target: -1, Left: left, Right: right, Inclusive: op == tokens.TokenKindLessThanOrEqual}, nil
	case tokens.TokenKindGreaterThan, tokens.TokenKindGreaterThanOrEqual:
		return NodeNumberComparison{Target: 1, Left: left, Right: right, Inclusive: op == tokens.TokenKindGreaterThanOrEqual}, nil
	case tokens.TokenKindLogicalAnd:
		return NodeLogicalOp{Left: left, Right: right, Op: LogicalOpAnd}, nil
	case tokens.TokenKindLogicalOr:
		return NodeLogicalOp{Left: left, Right: right, Op: LogicalOpOr}, nil
	default:
		return nil, errors.New("unexpected binary operation: " + op.String())
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

func parseNumberLiteral(lit string) (float64, error) {
	// Strip underscore separators
	cleaned := strings.ReplaceAll(lit, "_", "")

	// Strip BigInt suffix
	if len(cleaned) > 0 && cleaned[len(cleaned)-1] == 'n' {
		cleaned = cleaned[:len(cleaned)-1]
	}

	// Handle prefixed integer formats
	if len(cleaned) > 2 && cleaned[0] == '0' {
		switch cleaned[1] {
		case 'x', 'X':
			n, err := strconv.ParseInt(cleaned[2:], 16, 64)
			return float64(n), err
		case 'b', 'B':
			n, err := strconv.ParseInt(cleaned[2:], 2, 64)
			return float64(n), err
		case 'o', 'O':
			n, err := strconv.ParseInt(cleaned[2:], 8, 64)
			return float64(n), err
		}
	}

	return strconv.ParseFloat(cleaned, 64)
}
