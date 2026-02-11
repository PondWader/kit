package tokens

type Token struct {
	Kind    TokenKind
	Literal string
}

func (t Token) String() string {
	if t.Literal == "" {
		return t.Kind.String()
	}
	kindString := t.Kind.String()
	if kindString[0] == '"' {
		return `"` + t.Literal + `"`
	}
	return `"` + t.Literal + `" (` + t.Kind.String() + ")"
}

type TokenKind uint32

const (
	// Default type state
	TokenKindIllegalState = iota

	// JavaScript Keywords
	TokenKindBreak
	TokenKindCase
	TokenKindCatch
	TokenKindClass
	TokenKindConst
	TokenKindContinue
	TokenKindDebugger
	TokenKindDefault
	TokenKindDelete
	TokenKindDo
	TokenKindElse
	TokenKindEnum
	TokenKindExport
	TokenKindExtends
	TokenKindFalse
	TokenKindFinally
	TokenKindFor
	TokenKindFunction
	TokenKindIf
	TokenKindImport
	TokenKindIn
	TokenKindInstanceof
	TokenKindLet
	TokenKindNew
	TokenKindNull
	TokenKindReturn
	TokenKindSuper
	TokenKindSwitch
	TokenKindThis
	TokenKindThrow
	TokenKindTrue
	TokenKindTry
	TokenKindTypeof
	TokenKindUndefined
	TokenKindVar
	TokenKindVoid
	TokenKindWhile
	TokenKindWith
	TokenKindYield
	TokenKindAsync
	TokenKindAwait
	TokenKindOf
	TokenKindStatic
	TokenKindFrom

	// Protected Keywords
	TokenKindAbstract
	TokenKindAny
	TokenKindAs
	TokenKindAsserts
	TokenKindBoolean
	TokenKindConstructor
	TokenKindDeclare
	TokenKindGet
	TokenKindImplements
	TokenKindInfer
	TokenKindInterface
	TokenKindIs
	TokenKindKeyof
	TokenKindModule
	TokenKindNamespace
	TokenKindNever
	TokenKindKeywordNumber
	TokenKindObject
	TokenKindPackage
	TokenKindPrivate
	TokenKindProtected
	TokenKindPublic
	TokenKindReadonly
	TokenKindRequire
	TokenKindSet
	TokenKindString
	TokenKindSymbol
	TokenKindType
	TokenKindUnique
	TokenKindUnknown
	TokenKindUsing

	// Operators
	TokenKindPlus
	TokenKindMinus
	TokenKindMultiply
	TokenKindDivide
	TokenKindModulo
	TokenKindExponent
	TokenKindIncrement
	TokenKindDecrement
	TokenKindNotEquals
	TokenKindEquals
	TokenKindLessThan
	TokenKindLessThanOrEqual
	TokenKindGreaterThan
	TokenKindGreaterThanOrEqual
	TokenKindLogicalAnd
	TokenKindLogicalOr
	TokenKindLogicalNot
	TokenKindBitwiseAnd
	TokenKindBitwiseOr
	TokenKindBitwiseXor
	TokenKindBitwiseNot
	TokenKindLeftShift
	TokenKindRightShift
	TokenKindUnsignedRightShift
	TokenKindAssign
	TokenKindPlusAssign
	TokenKindMinusAssign
	TokenKindMultiplyAssign
	TokenKindDivideAssign
	TokenKindModuloAssign
	TokenKindExponentAssign
	TokenKindLeftShiftAssign
	TokenKindRightShiftAssign
	TokenKindUnsignedRightShiftAssign
	TokenKindBitwiseAndAssign
	TokenKindBitwiseOrAssign
	TokenKindBitwiseXorAssign
	TokenKindNullishCoalescing
	TokenKindOptionalChaining
	TokenKindConditional
	TokenKindColon
	TokenKindComma
	TokenKindSemicolon
	TokenKindArrow
	TokenKindRest

	// Punctuation and Delimiters
	TokenKindLeftParen
	TokenKindRightParen
	TokenKindLeftSquareBracket
	TokenKindRightSquareBracket
	TokenKindLeftBrace
	TokenKindRightBrace
	TokenKindDot
	TokenKindBacktick
	TokenKindDoubleQuote
	TokenKindSingleQuote
	TokenKindHash

	// Literals
	TokenKindIdentifier
	TokenKindNumberLiteral
	TokenKindStringLiteral
	TokenKindTemplateLiteral
	TokenKindRegexLiteral
	TokenKindBooleanLiteral
	TokenKindNullLiteral
	TokenKindUndefinedLiteral

	// Special
	TokenKindEOF
	TokenKindNewLine
	TokenKindWhitespace
	TokenKindSingleLineComment
	TokenKindMultiLineComment
	TokenKindUnknownToken
)

func (tk TokenKind) String() string {
	switch tk {
	case TokenKindIllegalState:
		return "ILLEGAL_TOKEN_STATE"
	// JavaScript Keywords
	case TokenKindBreak:
		return `"break"`
	case TokenKindCase:
		return `"case"`
	case TokenKindCatch:
		return `"catch"`
	case TokenKindClass:
		return `"class"`
	case TokenKindConst:
		return `"const"`
	case TokenKindContinue:
		return `"continue"`
	case TokenKindDebugger:
		return `"debugger"`
	case TokenKindDefault:
		return `"default"`
	case TokenKindDelete:
		return `"delete"`
	case TokenKindDo:
		return `"do"`
	case TokenKindElse:
		return `"else"`
	case TokenKindEnum:
		return `"enum"`
	case TokenKindExport:
		return `"export"`
	case TokenKindExtends:
		return `"extends"`
	case TokenKindFalse:
		return `"false"`
	case TokenKindFinally:
		return `"finally"`
	case TokenKindFor:
		return `"for"`
	case TokenKindFunction:
		return `"function"`
	case TokenKindIf:
		return `"if"`
	case TokenKindImport:
		return `"import"`
	case TokenKindIn:
		return `"in"`
	case TokenKindInstanceof:
		return `"instanceof"`
	case TokenKindLet:
		return `"let"`
	case TokenKindNew:
		return `"new"`
	case TokenKindNull:
		return `"null"`
	case TokenKindReturn:
		return `"return"`
	case TokenKindSuper:
		return `"super"`
	case TokenKindSwitch:
		return `"switch"`
	case TokenKindThis:
		return `"this"`
	case TokenKindThrow:
		return `"throw"`
	case TokenKindTrue:
		return `"true"`
	case TokenKindTry:
		return `"try"`
	case TokenKindTypeof:
		return `"typeof"`
	case TokenKindUndefined:
		return `"undefined"`
	case TokenKindVar:
		return `"var"`
	case TokenKindVoid:
		return `"void"`
	case TokenKindWhile:
		return `"while"`
	case TokenKindWith:
		return `"with"`
	case TokenKindYield:
		return `"yield"`
	case TokenKindAsync:
		return `"async"`
	case TokenKindAwait:
		return `"await"`
	case TokenKindOf:
		return `"of"`
	case TokenKindStatic:
		return `"static"`
	case TokenKindFrom:
		return `"from"`

	// TypeScript Keywords
	case TokenKindAbstract:
		return `"abstract"`
	case TokenKindAny:
		return `"any"`
	case TokenKindAs:
		return `"as"`
	case TokenKindAsserts:
		return `"asserts"`
	case TokenKindBoolean:
		return `"boolean"`
	case TokenKindConstructor:
		return `"constructor"`
	case TokenKindDeclare:
		return `"declare"`
	case TokenKindGet:
		return `"get"`
	case TokenKindImplements:
		return `"implements"`
	case TokenKindInfer:
		return `"infer"`
	case TokenKindInterface:
		return `"interface"`
	case TokenKindIs:
		return `"is"`
	case TokenKindKeyof:
		return `"keyof"`
	case TokenKindModule:
		return `"module"`
	case TokenKindNamespace:
		return `"namespace"`
	case TokenKindNever:
		return `"never"`
	case TokenKindKeywordNumber:
		return `"number"`
	case TokenKindObject:
		return `"object"`
	case TokenKindPackage:
		return `"package"`
	case TokenKindPrivate:
		return `"private"`
	case TokenKindProtected:
		return `"protected"`
	case TokenKindPublic:
		return `"public"`
	case TokenKindReadonly:
		return `"readonly"`
	case TokenKindRequire:
		return `"require"`
	case TokenKindSet:
		return `"set"`
	case TokenKindString:
		return `"string"`
	case TokenKindSymbol:
		return `"symbol"`
	case TokenKindType:
		return `"type"`
	case TokenKindUnique:
		return `"unique"`
	case TokenKindUnknown:
		return `"unknown"`
	case TokenKindUsing:
		return `"using"`

	// Operators
	case TokenKindPlus:
		return `"+"`
	case TokenKindMinus:
		return `"-"`
	case TokenKindMultiply:
		return `"*"`
	case TokenKindDivide:
		return `"/"`
	case TokenKindModulo:
		return `"%"`
	case TokenKindExponent:
		return `"**"`
	case TokenKindIncrement:
		return `"++"`
	case TokenKindDecrement:
		return `"--"`
	case TokenKindNotEquals:
		return `"!="`
	case TokenKindEquals:
		return `"=="`
	case TokenKindLessThan:
		return `"<"`
	case TokenKindLessThanOrEqual:
		return `"<="`
	case TokenKindGreaterThan:
		return `">"`
	case TokenKindGreaterThanOrEqual:
		return `">="`
	case TokenKindLogicalAnd:
		return `"&&"`
	case TokenKindLogicalOr:
		return `"||"`
	case TokenKindLogicalNot:
		return `"!"`
	case TokenKindBitwiseAnd:
		return `"&"`
	case TokenKindBitwiseOr:
		return `"|"`
	case TokenKindBitwiseXor:
		return `"^"`
	case TokenKindBitwiseNot:
		return `"~"`
	case TokenKindLeftShift:
		return `"<<"`
	case TokenKindRightShift:
		return `">>"`
	case TokenKindUnsignedRightShift:
		return `">>>"`
	case TokenKindAssign:
		return `"="`
	case TokenKindPlusAssign:
		return `"+="`
	case TokenKindMinusAssign:
		return `"-="`
	case TokenKindMultiplyAssign:
		return `"*="`
	case TokenKindDivideAssign:
		return `"/="`
	case TokenKindModuloAssign:
		return `"%="`
	case TokenKindExponentAssign:
		return `"**="`
	case TokenKindLeftShiftAssign:
		return `"<<="`
	case TokenKindRightShiftAssign:
		return `">>="`
	case TokenKindUnsignedRightShiftAssign:
		return `">>>="`
	case TokenKindBitwiseAndAssign:
		return `"&="`
	case TokenKindBitwiseOrAssign:
		return `"|="`
	case TokenKindBitwiseXorAssign:
		return `"^="`
	case TokenKindNullishCoalescing:
		return `"??"`
	case TokenKindOptionalChaining:
		return `"?."`
	case TokenKindConditional:
		return `"?"`
	case TokenKindColon:
		return `":"`
	case TokenKindComma:
		return `","`
	case TokenKindSemicolon:
		return `";"`
	case TokenKindArrow:
		return `"->"`
	case TokenKindRest:
		return `"..."`

	// Punctuation and Delimiters
	case TokenKindLeftParen:
		return `"("`
	case TokenKindRightParen:
		return `")"`
	case TokenKindLeftSquareBracket:
		return `"["`
	case TokenKindRightSquareBracket:
		return `"]"`
	case TokenKindLeftBrace:
		return `"{"`
	case TokenKindRightBrace:
		return `"}"`
	case TokenKindDot:
		return `"."`
	case TokenKindBacktick:
		return "\"`\""
	case TokenKindDoubleQuote:
		return `"\""`
	case TokenKindSingleQuote:
		return `"'"`
	case TokenKindHash:
		return `"#"`

	// Literals
	case TokenKindIdentifier:
		return "IDENTIFIER"
	case TokenKindNumberLiteral:
		return "NUMBER_LITERAL"
	case TokenKindStringLiteral:
		return "STRING_LITERAL"
	case TokenKindTemplateLiteral:
		return "TEMPLATE_LITERAL"
	case TokenKindRegexLiteral:
		return "REGEX_LITERAL"
	case TokenKindBooleanLiteral:
		return "BOOLEAN_LITERAL"
	case TokenKindNullLiteral:
		return "NULL_LITERAL"
	case TokenKindUndefinedLiteral:
		return "UNDEFINED_LITERAL"

	// Special
	case TokenKindEOF:
		return "EOF"
	case TokenKindNewLine:
		return "NEWLINE"
	case TokenKindWhitespace:
		return "WHITESPACE"
	case TokenKindSingleLineComment:
		return "SINGLE_LINE_COMMENT"
	case TokenKindMultiLineComment:
		return "MULTI_LINE_COMMENT"
	case TokenKindUnknownToken:
		return "UNKNOWN"

	default:
		return "INVALID"
	}
}
