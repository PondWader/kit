package tokens

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var (
	ErrInvalidIdentifier    = errors.New("invalid identifier")
	ErrUnexpectedSymbol     = errors.New("character is not an expected symbol")
	ErrInvalidNumberLiteral = errors.New("invalid number literal")
)

func NewLexer(input io.Reader) *Lexer {
	return &Lexer{
		r:           bufio.NewReader(input),
		currentLine: 1,
	}
}

type Lexer struct {
	r *bufio.Reader

	nextTokenAvailable   bool
	nextToken            Token
	currentLine          int
	emitTypescriptTokens bool

	State int
}

func (l *Lexer) SetEmitTypeScriptTokens(enabled bool) {
	l.emitTypescriptTokens = enabled
}

func (l *Lexer) Next() (Token, error) {
	l.State++

	if l.nextTokenAvailable {
		l.nextTokenAvailable = false
		return l.nextToken, nil
	}

	var sb strings.Builder

	for {
		ch, _, err := l.r.ReadRune()
		if err != nil {
			// Handle EOF while building a literal
			if errors.Is(err, io.EOF) {
				if sb.Len() > 0 {
					return l.getTextToken(sb.String())
				}
				return Token{Kind: TokenKindEOF}, nil
			}
			return Token{}, err
		}

		// Detect number literals at the start of a token
		if sb.Len() == 0 {
			// Case 1: Digit starts a number
			if ch >= '0' && ch <= '9' {
				return l.scanNumber(ch)
			}

			// Case 2: Dot might start a number (.456)
			if ch == '.' {
				// Peek ahead to see if it's followed by a digit
				nextCh, _, peekErr := l.r.ReadRune()
				if peekErr == nil {
					if nextCh >= '0' && nextCh <= '9' {
						// It's a number starting with .
						l.r.UnreadRune() // unread the digit
						return l.scanNumber(ch)
					}
					l.r.UnreadRune() // unread whatever we peeked
				}
				// Not a number, fall through to symbol processing
			}
		}

		if _, ok := textToSymbolToken[string(ch)]; ok {
			if sb.Len() > 0 {
				l.r.UnreadRune()
				return l.getTextToken(sb.String())
			}

			return l.getSymbolToken(ch)
		}

		// Character is not a token, add to literal builder
		sb.WriteRune(ch)
	}
}

var textToSymbolToken = map[string]TokenKind{
	"=":    TokenKindAssign,
	"==":   TokenKindLooseEquals,
	"===":  TokenKindStrictEquals,
	"->":   TokenKindArrow,
	"!":    TokenKindLogicalNot,
	"!=":   TokenKindLooseNotEquals,
	"!==":  TokenKindStrictNotEquals,
	"<":    TokenKindLessThan,
	"<=":   TokenKindLessThanOrEqual,
	"<<":   TokenKindLeftShift,
	"<<=":  TokenKindLeftShiftAssign,
	">":    TokenKindGreaterThan,
	">=":   TokenKindGreaterThanOrEqual,
	">>":   TokenKindRightShift,
	">>=":  TokenKindRightShiftAssign,
	">>>":  TokenKindUnsignedRightShift,
	">>>=": TokenKindUnsignedRightShiftAssign,
	"+":    TokenKindPlus,
	"++":   TokenKindIncrement,
	"+=":   TokenKindPlusAssign,
	"-":    TokenKindMinus,
	"--":   TokenKindDecrement,
	"-=":   TokenKindMinusAssign,
	"*":    TokenKindMultiply,
	"**":   TokenKindExponent,
	"*=":   TokenKindMultiplyAssign,
	"**=":  TokenKindExponentAssign,
	"/":    TokenKindDivide,
	"/=":   TokenKindDivideAssign,
	"%":    TokenKindModulo,
	"%=":   TokenKindModuloAssign,
	"&":    TokenKindBitwiseAnd,
	"&&":   TokenKindLogicalAnd,
	"&=":   TokenKindBitwiseAndAssign,
	"|":    TokenKindBitwiseOr,
	"||":   TokenKindLogicalOr,
	"|=":   TokenKindBitwiseOrAssign,
	"^":    TokenKindBitwiseXor,
	"^=":   TokenKindBitwiseXorAssign,
	"~":    TokenKindBitwiseNot,
	"?":    TokenKindConditional,
	"??":   TokenKindNullishCoalescing,
	"?.":   TokenKindOptionalChaining,
	".":    TokenKindDot,
	":":    TokenKindColon,
	";":    TokenKindSemicolon,
	",":    TokenKindComma,
	"(":    TokenKindLeftParen,
	")":    TokenKindRightParen,
	"[":    TokenKindLeftSquareBracket,
	"]":    TokenKindRightSquareBracket,
	"{":    TokenKindLeftBrace,
	"}":    TokenKindRightBrace,
	"'":    TokenKindSingleQuote,
	"#":    TokenKindHash,
	`"`:    TokenKindDoubleQuote,
	"`":    TokenKindBacktick,
	" ":    TokenKindWhitespace,
	"\r":   TokenKindWhitespace,
	"\t":   TokenKindWhitespace,
	"\n":   TokenKindNewLine,
}

// getOperatorToken parses single and multi-character operators
func (l *Lexer) getSymbolToken(firstChar rune) (Token, error) {
	if firstChar == '\n' {
		l.currentLine++
	}

	// Check for comments when we see '/'
	if firstChar == '/' {
		nextCh, _, err := l.r.ReadRune()
		if err != nil && !errors.Is(err, io.EOF) {
			return Token{}, err
		}

		if err == nil {
			// Single-line comment
			if nextCh == '/' {
				return l.scanSingleLineComment()
			}
			// Multi-line comment
			if nextCh == '*' {
				return l.scanMultiLineComment()
			}
			// Not a comment, unread the character and continue with operator parsing
			l.r.UnreadRune()
		}
	}

	str := string(firstChar)
	for {
		if _, ok := textToSymbolToken[str]; !ok {
			if len(str) > 1 {
				str = str[0 : len(str)-1]
				tokenKind := textToSymbolToken[str]
				l.r.UnreadRune()
				return Token{Kind: tokenKind, Literal: str}, nil
			}

			return Token{}, fmt.Errorf("%w: %s", ErrUnexpectedSymbol, str)
		}

		ch, _, err := l.r.ReadRune()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return Token{}, err
			}
			return Token{Kind: textToSymbolToken[str], Literal: str}, nil
		}
		str += string(ch)
	}
}

func (l *Lexer) getTextToken(literal string) (Token, error) {
	switch literal {
	// JavaScript Keywords
	case "break":
		return Token{Kind: TokenKindBreak, Literal: literal}, nil
	case "case":
		return Token{Kind: TokenKindCase, Literal: literal}, nil
	case "catch":
		return Token{Kind: TokenKindCatch, Literal: literal}, nil
	case "class":
		return Token{Kind: TokenKindClass, Literal: literal}, nil
	case "const":
		return Token{Kind: TokenKindConst, Literal: literal}, nil
	case "continue":
		return Token{Kind: TokenKindContinue, Literal: literal}, nil
	case "debugger":
		return Token{Kind: TokenKindDebugger, Literal: literal}, nil
	case "default":
		return Token{Kind: TokenKindDefault, Literal: literal}, nil
	case "delete":
		return Token{Kind: TokenKindDelete, Literal: literal}, nil
	case "do":
		return Token{Kind: TokenKindDo, Literal: literal}, nil
	case "else":
		return Token{Kind: TokenKindElse, Literal: literal}, nil
	case "enum":
		return Token{Kind: TokenKindEnum, Literal: literal}, nil
	case "export":
		return Token{Kind: TokenKindExport, Literal: literal}, nil
	case "extends":
		return Token{Kind: TokenKindExtends, Literal: literal}, nil
	case "false":
		return Token{Kind: TokenKindFalse, Literal: literal}, nil
	case "finally":
		return Token{Kind: TokenKindFinally, Literal: literal}, nil
	case "for":
		return Token{Kind: TokenKindFor, Literal: literal}, nil
	case "fn":
		return Token{Kind: TokenKindFunction, Literal: literal}, nil
	case "if":
		return Token{Kind: TokenKindIf, Literal: literal}, nil
	case "import":
		return Token{Kind: TokenKindImport, Literal: literal}, nil
	case "in":
		return Token{Kind: TokenKindIn, Literal: literal}, nil
	case "instanceof":
		return Token{Kind: TokenKindInstanceof, Literal: literal}, nil
	case "let":
		return Token{Kind: TokenKindLet, Literal: literal}, nil
	case "new":
		return Token{Kind: TokenKindNew, Literal: literal}, nil
	case "null":
		return Token{Kind: TokenKindNull, Literal: literal}, nil
	case "return":
		return Token{Kind: TokenKindReturn, Literal: literal}, nil
	case "super":
		return Token{Kind: TokenKindSuper, Literal: literal}, nil
	case "switch":
		return Token{Kind: TokenKindSwitch, Literal: literal}, nil
	case "this":
		return Token{Kind: TokenKindThis, Literal: literal}, nil
	case "throw":
		return Token{Kind: TokenKindThrow, Literal: literal}, nil
	case "true":
		return Token{Kind: TokenKindTrue, Literal: literal}, nil
	case "try":
		return Token{Kind: TokenKindTry, Literal: literal}, nil
	case "typeof":
		return Token{Kind: TokenKindTypeof, Literal: literal}, nil
	case "undefined":
		return Token{Kind: TokenKindUndefined, Literal: literal}, nil
	case "var":
		return Token{Kind: TokenKindVar, Literal: literal}, nil
	case "void":
		return Token{Kind: TokenKindVoid, Literal: literal}, nil
	case "while":
		return Token{Kind: TokenKindWhile, Literal: literal}, nil
	case "with":
		return Token{Kind: TokenKindWith, Literal: literal}, nil
	case "yield":
		return Token{Kind: TokenKindYield, Literal: literal}, nil
	case "async":
		return Token{Kind: TokenKindAsync, Literal: literal}, nil
	case "await":
		return Token{Kind: TokenKindAwait, Literal: literal}, nil
	case "of":
		return Token{Kind: TokenKindOf, Literal: literal}, nil
	case "static":
		return Token{Kind: TokenKindStatic, Literal: literal}, nil
	case "from":
		return Token{Kind: TokenKindFrom, Literal: literal}, nil
	}

	if l.emitTypescriptTokens {
		switch literal {
		// TypeScript Keywords
		case "abstract":
			return Token{Kind: TokenKindAbstract, Literal: literal}, nil
		case "any":
			return Token{Kind: TokenKindAny, Literal: literal}, nil
		case "as":
			return Token{Kind: TokenKindAs, Literal: literal}, nil
		case "asserts":
			return Token{Kind: TokenKindAsserts, Literal: literal}, nil
		case "boolean":
			return Token{Kind: TokenKindBoolean, Literal: literal}, nil
		case "constructor":
			return Token{Kind: TokenKindConstructor, Literal: literal}, nil
		case "declare":
			return Token{Kind: TokenKindDeclare, Literal: literal}, nil
		case "get":
			return Token{Kind: TokenKindGet, Literal: literal}, nil
		case "implements":
			return Token{Kind: TokenKindImplements, Literal: literal}, nil
		case "infer":
			return Token{Kind: TokenKindInfer, Literal: literal}, nil
		case "interface":
			return Token{Kind: TokenKindInterface, Literal: literal}, nil
		case "is":
			return Token{Kind: TokenKindIs, Literal: literal}, nil
		case "keyof":
			return Token{Kind: TokenKindKeyof, Literal: literal}, nil
		case "module":
			return Token{Kind: TokenKindModule, Literal: literal}, nil
		case "namespace":
			return Token{Kind: TokenKindNamespace, Literal: literal}, nil
		case "never":
			return Token{Kind: TokenKindNever, Literal: literal}, nil
		case "number":
			return Token{Kind: TokenKindKeywordNumber, Literal: literal}, nil
		case "object":
			return Token{Kind: TokenKindObject, Literal: literal}, nil
		case "package":
			return Token{Kind: TokenKindPackage, Literal: literal}, nil
		case "private":
			return Token{Kind: TokenKindPrivate, Literal: literal}, nil
		case "protected":
			return Token{Kind: TokenKindProtected, Literal: literal}, nil
		case "public":
			return Token{Kind: TokenKindPublic, Literal: literal}, nil
		case "readonly":
			return Token{Kind: TokenKindReadonly, Literal: literal}, nil
		case "require":
			return Token{Kind: TokenKindRequire, Literal: literal}, nil
		case "set":
			return Token{Kind: TokenKindSet, Literal: literal}, nil
		case "string":
			return Token{Kind: TokenKindString, Literal: literal}, nil
		case "symbol":
			return Token{Kind: TokenKindSymbol, Literal: literal}, nil
		case "type":
			return Token{Kind: TokenKindType, Literal: literal}, nil
		case "unique":
			return Token{Kind: TokenKindUnique, Literal: literal}, nil
		case "unknown":
			return Token{Kind: TokenKindUnknown, Literal: literal}, nil
		case "using":
			return Token{Kind: TokenKindUsing, Literal: literal}, nil
		}
	}

	if !isValidIdentifier(literal) {
		return Token{}, fmt.Errorf("%w: %s", ErrInvalidIdentifier, literal)
	}

	return Token{Kind: TokenKindIdentifier, Literal: literal}, nil
}

func isValidIdentifier(str string) bool {
	if len(str) == 0 {
		return false
	}

	// First character must be a letter, underscore, or dollar sign
	firstRune := rune(str[0])
	if !unicode.IsLetter(firstRune) && firstRune != '_' && firstRune != '$' {
		return false
	}

	// Remaining characters can be letters, digits, underscores, or dollar signs
	for _, char := range str[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '$' {
			return false
		}
	}

	return true
}

func (l *Lexer) GetLine() int {
	return l.currentLine
}

func (l *Lexer) AddLine() {
	l.currentLine++
}

func (l *Lexer) Unread(token Token) {
	if l.nextTokenAvailable {
		panic("a token has already been unread that has not been consumed")
	}
	l.nextTokenAvailable = true
	l.nextToken = token
	l.State--
}

// scanSingleLineComment scans a single-line comment starting with //
func (l *Lexer) scanSingleLineComment() (Token, error) {
	var sb strings.Builder
	sb.WriteString("//")

	for {
		ch, _, err := l.r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return Token{Kind: TokenKindSingleLineComment, Literal: sb.String()}, nil
			}
			return Token{}, err
		}

		if ch == '\n' {
			l.r.UnreadRune() // Unread the newline so it can be processed normally
			return Token{Kind: TokenKindSingleLineComment, Literal: sb.String()}, nil
		}

		sb.WriteRune(ch)
	}
}

// scanMultiLineComment scans a multi-line comment starting with /*
func (l *Lexer) scanMultiLineComment() (Token, error) {
	var sb strings.Builder
	sb.WriteString("/*")

	for {
		ch, _, err := l.r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return Token{}, errors.New("unterminated multi-line comment")
			}
			return Token{}, err
		}

		if ch == '\n' {
			l.currentLine++
		}

		sb.WriteRune(ch)

		if ch == '*' {
			nextCh, _, err := l.r.ReadRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return Token{}, errors.New("unterminated multi-line comment")
				}
				return Token{}, err
			}

			if nextCh == '/' {
				sb.WriteRune(nextCh)
				return Token{Kind: TokenKindMultiLineComment, Literal: sb.String()}, nil
			}

			if nextCh == '\n' {
				l.currentLine++
			}

			sb.WriteRune(nextCh)
		}
	}
}
