package tokens

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// scanNumber scans a complete number literal starting with firstChar
func (l *Lexer) scanNumber(firstChar rune) (Token, error) {
	var sb strings.Builder
	sb.WriteRune(firstChar)

	// Handle different number formats based on prefix
	if firstChar == '0' {
		// Peek next character to check for hex/binary/octal prefix
		ch, _, err := l.r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Just "0" - valid number
				return Token{Kind: TokenKindNumberLiteral, Literal: "0"}, nil
			}
			return Token{}, err
		}

		// Check for prefixes: 0x, 0b, 0o
		switch ch {
		case 'x', 'X':
			// Hexadecimal
			sb.WriteRune(ch)
			return l.scanPrefixedNumber(&sb, func(r rune) bool {
				return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == '_'
			})
		case 'b', 'B':
			// Binary
			sb.WriteRune(ch)
			return l.scanPrefixedNumber(&sb, func(r rune) bool {
				return r == '0' || r == '1' || r == '_'
			})
		case 'o', 'O':
			// Octal
			sb.WriteRune(ch)
			return l.scanPrefixedNumber(&sb, func(r rune) bool {
				return (r >= '0' && r <= '7') || r == '_'
			})
		}

		// Not a prefix, unread and continue as decimal
		l.r.UnreadRune()
	}

	// Decimal number (possibly with decimal point and/or exponent)
	return l.scanDecimalNumber(&sb)
}

// scanPrefixedNumber scans hex/binary/octal numbers after the prefix
func (l *Lexer) scanPrefixedNumber(sb *strings.Builder, isValidDigit func(rune) bool) (Token, error) {
	hasDigits := false

	for {
		ch, _, err := l.r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return Token{}, err
		}

		if isValidDigit(ch) {
			if ch != '_' {
				hasDigits = true
			}
			sb.WriteRune(ch)
		} else if ch == 'n' && hasDigits {
			// BigInt suffix
			sb.WriteRune(ch)
			break
		} else {
			// End of number
			l.r.UnreadRune()
			break
		}
	}

	literal := sb.String()
	if !isValidNumber(literal) {
		return Token{}, fmt.Errorf("%w: %s", ErrInvalidNumberLiteral, literal)
	}
	return Token{Kind: TokenKindNumberLiteral, Literal: literal}, nil
}

// scanDecimalNumber scans decimal numbers (with optional decimal point and exponent)
func (l *Lexer) scanDecimalNumber(sb *strings.Builder) (Token, error) {
	// Scan integer part (already have first digit or '.' in sb)
	firstChar := []rune(sb.String())[0]

	// If we started with '.', check that we're followed by a digit
	if firstChar == '.' {
		ch, _, err := l.r.ReadRune()
		if err != nil || !(ch >= '0' && ch <= '9') {
			// Not a number, just a dot
			if err == nil {
				l.r.UnreadRune()
			}
			return Token{}, fmt.Errorf("%w: %s", ErrInvalidNumberLiteral, sb.String())
		}
		sb.WriteRune(ch)
	}

	// Read integer/fractional digits
	for {
		ch, _, err := l.r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return Token{}, err
		}

		if (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		} else if ch == '.' && firstChar != '.' {
			// Potential decimal point - peek ahead to check if followed by digit or EOF
			nextCh, _, peekErr := l.r.ReadRune()
			if peekErr == nil {
				if nextCh >= '0' && nextCh <= '9' {
					// It's a decimal point followed by digits
					sb.WriteRune(ch)
					sb.WriteRune(nextCh)
					// Continue reading fractional part
					continue
				}
				// Not followed by digit - the dot is not part of this number
				// Unread the peeked character and save the dot for next token
				l.r.UnreadRune() // unread the peeked character
			}
			// Save the dot as a token for the next call to Next()
			l.Unread(Token{Kind: TokenKindDot, Literal: "."})
			break
		} else if ch == 'e' || ch == 'E' {
			// Exponent notation
			sb.WriteRune(ch)
			if err := l.scanExponent(sb); err != nil {
				return Token{}, err
			}
			break
		} else if ch == 'n' {
			// BigInt suffix
			sb.WriteRune(ch)
			break
		} else {
			// End of number
			l.r.UnreadRune()
			break
		}
	}

	literal := sb.String()
	if !isValidNumber(literal) {
		return Token{}, fmt.Errorf("%w: %s", ErrInvalidNumberLiteral, literal)
	}
	return Token{Kind: TokenKindNumberLiteral, Literal: literal}, nil
}

// scanExponent scans the exponent part of a number (after 'e' or 'E')
func (l *Lexer) scanExponent(sb *strings.Builder) error {
	// Optional sign
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return fmt.Errorf("%w: incomplete exponent", ErrInvalidNumberLiteral)
	}

	if ch == '+' || ch == '-' {
		sb.WriteRune(ch)
		ch, _, err = l.r.ReadRune()
		if err != nil {
			return fmt.Errorf("%w: incomplete exponent", ErrInvalidNumberLiteral)
		}
	}

	// Must have at least one digit
	if !(ch >= '0' && ch <= '9') {
		return fmt.Errorf("%w: exponent requires digits", ErrInvalidNumberLiteral)
	}
	sb.WriteRune(ch)

	// Read remaining exponent digits
	for {
		ch, _, err := l.r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		} else {
			l.r.UnreadRune()
			break
		}
	}

	return nil
}

func isValidNumber(str string) bool {
	if len(str) == 0 {
		return false
	}

	// Check for BigInt suffix (123n)
	hasBigIntSuffix := false
	if str[len(str)-1] == 'n' {
		hasBigIntSuffix = true
		str = str[:len(str)-1]
		if len(str) == 0 {
			return false
		}
	}

	// Hexadecimal: 0x1F, 0X1F
	if len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
		if len(str) == 2 {
			return false // just "0x" is invalid
		}
		for _, ch := range str[2:] {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') || ch == '_') {
				return false
			}
		}
		return true
	}

	// Binary: 0b1010, 0B1010
	if len(str) >= 2 && str[0] == '0' && (str[1] == 'b' || str[1] == 'B') {
		if len(str) == 2 {
			return false // just "0b" is invalid
		}
		for _, ch := range str[2:] {
			if ch != '0' && ch != '1' && ch != '_' {
				return false
			}
		}
		return true
	}

	// Octal: 0o17, 0O17
	if len(str) >= 2 && str[0] == '0' && (str[1] == 'o' || str[1] == 'O') {
		if len(str) == 2 {
			return false // just "0o" is invalid
		}
		for _, ch := range str[2:] {
			if (ch < '0' || ch > '7') && ch != '_' {
				return false
			}
		}
		return true
	}

	// Decimal number (possibly with exponent and/or decimal point)
	// Valid: 123, 123.456, .456, 123., 1e10, 1.23e-10, 1_000_000
	i := 0
	hasDigits := false
	dotCount := 0

	// Leading digits before decimal point (with optional underscores)
	for i < len(str) && ((str[i] >= '0' && str[i] <= '9') || str[i] == '_') {
		if str[i] != '_' {
			hasDigits = true
		}
		i++
	}

	// Optional decimal point
	if i < len(str) && str[i] == '.' {
		dotCount++
		i++
		// Digits after decimal point
		for i < len(str) && ((str[i] >= '0' && str[i] <= '9') || str[i] == '_') {
			if str[i] != '_' {
				hasDigits = true
			}
			i++
		}
	}

	// BigInt cannot have decimal point or exponent
	if hasBigIntSuffix && dotCount > 0 {
		return false
	}

	// Must have at least one digit
	if !hasDigits {
		return false
	}

	// Optional exponent (e10, e+10, e-10, E10, etc.)
	if i < len(str) && (str[i] == 'e' || str[i] == 'E') {
		if hasBigIntSuffix {
			return false // BigInt cannot have exponent
		}
		i++
		// Optional sign
		if i < len(str) && (str[i] == '+' || str[i] == '-') {
			i++
		}
		// Must have at least one digit in exponent
		expStart := i
		for i < len(str) && ((str[i] >= '0' && str[i] <= '9') || str[i] == '_') {
			i++
		}
		if i == expStart {
			return false // no digits in exponent
		}
	}

	// Should have consumed entire string
	return i == len(str)
}
