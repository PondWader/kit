package lang

import (
	"strings"

	"github.com/PondWader/kit/pkg/lang/tokens"
	"github.com/PondWader/kit/pkg/lang/values"
)

func resolveEscapedCharacter(char rune) rune {
	switch char {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'v':
		return '\v'
	case '\\':
		return '\\'
	case '\'':
		return '\''
	default:
		return char
	}
}

func (p *parser) parsePureString() (string, error) {
	if _, err := p.expectToken(tokens.TokenKindDoubleQuote); err != nil {
		return "", err
	}

	var sb strings.Builder
	for {
		char, _, err := p.r.ReadRune()
		if err != nil {
			return "", err
		}
		if char == '"' {
			return sb.String(), nil
		}
		sb.WriteRune(char)
	}
}

func (p *parser) parseString(quotation rune) (n NodeString, err error) {
	var isEscaped bool
	var isDollarPrefixed bool
	var sb strings.Builder

	for {
		char, _, err := p.r.ReadRune()
		if err != nil || char == '\n' {
			return n, ErrUnterminatedString
		}

		if char == '\\' {
			isEscaped = !isEscaped
			if isEscaped {
				continue
			}
		}

		if isEscaped {
			resolvedChar := resolveEscapedCharacter(char)
			sb.WriteRune(resolvedChar)
			isEscaped = false
			isDollarPrefixed = false
			continue
		}

		if isDollarPrefixed && char == '{' {
			if sb.Len() > 0 {
				n.Parts = append(n.Parts, NodeLiteral{
					Value: values.String(sb.String()).Val(),
				})
				sb.Reset()
			}

			expr, err := p.parseExpression()
			if err != nil {
				return n, err
			}
			n.Parts = append(n.Parts, expr)

			_, err = p.expectToken(tokens.TokenKindRightBrace)
			if err != nil {
				return n, err
			}

			isDollarPrefixed = false
			continue
		} else if isDollarPrefixed {
			sb.WriteRune('$')
		}

		if char == quotation {
			if isDollarPrefixed {
				sb.WriteRune('$')
			}
			if sb.Len() > 0 {
				n.Parts = append(n.Parts, NodeLiteral{
					Value: values.String(sb.String()).Val(),
				})
			}
			break
		} else if char == '$' {
			isDollarPrefixed = true
		} else {
			sb.WriteRune(char)
			isDollarPrefixed = false
		}
	}

	return n, nil
}
