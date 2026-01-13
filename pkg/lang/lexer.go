package lang

import (
	"errors"
	"fmt"
	"os"
)

type TokenKind uint8

const (
	// Statements
	TokenIfStatement TokenKind = iota
	TokenElseStatement
	TokenFunctionDeclaration
	TokenImportStatement
	TokenExportStatement
	TokenForStatement
	TokenVarDeclaration
	TokenReturnStatement
	TokenStructDeclaration
	TokenAsStatement
	TokenRangeStatement
	TokenWhileStatement

	// Values
	TokenTrue
	TokenFalse
	TokenString
	TokenNumber
	TokenIdentifier

	// Types
	TokenKindTypeNumber
	TokenKindTypeString
	TokenKindTypeBool

	// Symbols
	TokenColon
	TokenSemiColon
	TokenNewLine
	TokenComma
	TokenLeftBracket
	TokenRightBracket
	TokenLeftBrace
	TokenRightBrace
	TokenLeftSquareBracket
	TokenRightSquareBracket
	TokenAsterisk
	TokenPlus
	TokenDash
	TokenForwardSlash
	TokenAmpersand
	TokenBar
	TokenApostrophe
	TokenExclamationMark
	TokenEquals
	TokenGreaterThan
	TokenLessThan
	TokenPeriod

	TokenEOF
)

type Token struct {
	Kind    TokenKind
	Literal string
	Line    int
}

type Lexer struct {
	content     string
	cursor      int
	currentLine int
}

func NewLexer(content string) *Lexer {
	return &Lexer{content, 0, 1}
}

func (l *Lexer) Next() (Token, error) {
	currentStr := ""
	for l.cursor < len(l.content) {
		// Need to use a substring instead of an index to get a value of type string
		char := l.content[l.cursor : l.cursor+1]
		l.cursor++

		if char == " " || char == "\t" || char == "\r" {
			continue
		}

		// Check character is a valid token
		if token, err := getCharTokenKind(char); err == nil {
			if token == TokenNewLine {
				l.currentLine++
			}
			return Token{
				Kind:    token,
				Literal: char,
				Line:    l.currentLine,
			}, nil
		}

		// If the character is a quotation mark, it's the beginning of a string
		if char == "\"" {
			strContent, err := l.readString()
			if err != nil {
				return Token{}, err
			}
			return Token{
				Kind:    TokenString,
				Literal: strContent,
				Line:    l.currentLine,
			}, nil
		}

		currentStr += char
		var endOfToken bool
		if l.cursor >= len(l.content) {
			endOfToken = true
		} else {
			nextChar := l.content[l.cursor : l.cursor+1]
			// Check if the next character terminates a token
			if nextChar == "" || nextChar == " " || nextChar == "\n" || nextChar == "\r" || nextChar == "\t" || nextChar == "\"" {
				endOfToken = true
			} else if _, err := getCharTokenKind(nextChar); err == nil {
				endOfToken = true
			}
		}

		if endOfToken {
			return Token{
				Kind:    getLiteralTokenKind(currentStr),
				Literal: currentStr,
				Line:    l.currentLine,
			}, nil
		}
	}

	return Token{
		Kind:    TokenEOF,
		Literal: "EOF",
		Line:    l.currentLine,
	}, nil
}

// Returns the contents of the next token without progressing the cursor
func (l *Lexer) Peek() (Token, error) {
	originalPos := l.cursor
	originalLine := l.currentLine
	token, err := l.Next()
	l.cursor = originalPos
	l.currentLine = originalLine
	return token, err
}

func (l *Lexer) PeekOrExit() Token {
	token, err := l.Peek()
	if err != nil {
		fmt.Println(err, fmt.Sprint(l.currentLine)+":"+fmt.Sprint(l.cursor))
		os.Exit(1)
	}
	return token
}

func (l *Lexer) NextOrExit() Token {
	token, err := l.Next()
	if err != nil {
		fmt.Println(err, fmt.Sprint(l.currentLine)+":"+fmt.Sprint(l.cursor))
		os.Exit(1)
	}
	return token
}

func (l *Lexer) readString() (string, error) {
	currentStr := ""
	escapedChar := false
	for l.cursor < len(l.content) {
		char := l.content[l.cursor : l.cursor+1]
		l.cursor++

		if char == "\\" && !escapedChar {
			escapedChar = true
			continue
		}
		if escapedChar {
			switch char {
			case "\"":
				return currentStr, nil
			case "n":
				currentStr += "\n"
			case "t":
				currentStr += "\t"
			case "r":
				currentStr += "\r"
			default:
				currentStr += char
			}
			escapedChar = false
		} else {
			if char == "\"" {
				return currentStr, nil
			}
			currentStr += char
		}
		if char == "\n" {
			return "", errors.New("unexpected newline while reading string literal")
		}
	}

	return "", errors.New("reached EOF without string finishing")
}

func (l *Lexer) GetCurrentLine() int {
	return l.currentLine
}

func (l *Lexer) SetCurrentLine(line int) {
	l.currentLine = line
}

func (l *Lexer) GetCursor() int {
	return l.cursor
}

func (l *Lexer) SetCursor(cursor int) {
	l.cursor = cursor
}

// Moves the cursor back to the start of the previously read token so it will be read at the next call of Next().
// Only the last read token should be passed to Unread.
func (l *Lexer) Unread(token Token) {
	if token.Kind == TokenEOF {
		return
	}
	l.cursor -= len(token.Literal)
	if token.Kind == TokenString {
		l.cursor -= 2 // Account for quotation marks on either side
	} else if token.Kind == TokenNewLine {
		l.currentLine--
	}
}

func (l *Lexer) SavePos() LexerPos {
	return LexerPos{l.cursor, l.currentLine, l}
}

// Stores the position of a lexer
type LexerPos struct {
	Cursor int
	Line   int
	lexer  *Lexer
}

func (pos LexerPos) GoTo() (undo func()) {
	originalLine := pos.lexer.GetCurrentLine()
	originalCursor := pos.lexer.GetCursor()

	pos.lexer.SetCurrentLine(pos.Line)
	pos.lexer.SetCurrentLine(pos.Cursor)

	return func() {
		pos.lexer.SetCurrentLine(originalLine)
		pos.lexer.SetCursor(originalCursor)
	}
}

func getCharTokenKind(char string) (TokenKind, error) {
	switch char {
	case ":":
		return TokenColon, nil
	case ";":
		return TokenSemiColon, nil
	case "\n":
		return TokenNewLine, nil
	case ",":
		return TokenComma, nil
	case "(":
		return TokenLeftBracket, nil
	case ")":
		return TokenRightBracket, nil
	case "{":
		return TokenLeftBrace, nil
	case "}":
		return TokenRightBrace, nil
	case "[":
		return TokenLeftSquareBracket, nil
	case "]":
		return TokenRightSquareBracket, nil
	case "*":
		return TokenAsterisk, nil
	case "+":
		return TokenPlus, nil
	case "-":
		return TokenDash, nil
	case "/":
		return TokenForwardSlash, nil
	case "&":
		return TokenAmpersand, nil
	case "|":
		return TokenBar, nil
	case "'":
		return TokenApostrophe, nil
	case "!":
		return TokenExclamationMark, nil
	case "=":
		return TokenEquals, nil
	case ">":
		return TokenGreaterThan, nil
	case "<":
		return TokenLessThan, nil
	case ".":
		return TokenPeriod, nil
	}
	return 0, errors.New("char provided is not a valid token")
}

func getLiteralTokenKind(literal string) TokenKind {
	switch literal {
	case "true":
		return TokenTrue
	case "false":
		return TokenFalse
	}

	switch literal {
	// Statements
	case "if":
		return TokenIfStatement
	case "else":
		return TokenElseStatement
	case "fn":
		return TokenFunctionDeclaration
	case "import":
		return TokenImportStatement
	case "export":
		return TokenExportStatement
	case "for":
		return TokenForStatement
	case "var":
		return TokenVarDeclaration
	case "return":
		return TokenReturnStatement
	case "struct":
		return TokenStructDeclaration
	case "as":
		return TokenAsStatement
	case "while":
		return TokenWhileStatement
	case "range":
		return TokenRangeStatement

	// Types
	case "number":
		return TokenKindTypeNumber
	case "string":
		return TokenKindTypeString
	case "bool":
		return TokenKindTypeBool
	}

	for _, char := range literal {
		// 0-9 are between char code 48 and 57
		if char < 48 || char > 57 {
			return TokenIdentifier
		}
	}

	return TokenNumber
}
