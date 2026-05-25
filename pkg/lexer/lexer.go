package lexer

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"
	FSTRING = "FSTRING"

	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	PLUS_EQ  = "+="
	MINUS_EQ = "-="
	MUL_EQ   = "*="
	DIV_EQ   = "/="

	EQ     = "=="
	NOT_EQ = "!="

	COMMA     = ","
	COLON     = ":"
	SEMICOLON = ";"
	DOT       = "."

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	LBRACKET = "["
	RBRACKET = "]"

	FUNCTION = "DEF"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	WHILE    = "WHILE"
	FOR      = "FOR"
	IN       = "IN"
	AND      = "AND"
	OR       = "OR"
	NOT      = "NOT"
	NONE     = "NONE"
	CLASS    = "CLASS"
	LAMBDA   = "LAMBDA"
	TRY      = "TRY"
	EXCEPT   = "EXCEPT"
	FINALLY  = "FINALLY"
	RAISE    = "RAISE"
	AS       = "AS"
	WITH     = "WITH"
	YIELD    = "YIELD"
	PASS     = "PASS"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"def":    FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"while":  WHILE,
	"for":    FOR,
	"in":     IN,
	"and":    AND,
	"or":     OR,
	"not":    NOT,
	"None":   NONE,
	"class":  CLASS,
	"lambda": LAMBDA,
	"try":    TRY,
	"except": EXCEPT,
	"finally": FINALLY,
	"raise": RAISE,
	"as": AS,
	"with":   WITH,
	"yield":  YIELD,
	"pass":   PASS,
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	prevNonWhiteCh byte // previous non-whitespace character
	justSkippedNewline bool // whether we just skipped a newline character
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	// 如果刚刚跳过了换行符，并且前一个 token 适合结束语句，就返回分号
	if l.justSkippedNewline {
		// 检查前一个非空白字符，判断是否应该插入分号
		// 注意：不在冒号后插入分号，因为冒号后面通常是缩进块
		if l.prevNonWhiteCh != ':' && (isIdentifierChar(l.prevNonWhiteCh) || l.prevNonWhiteCh == ')' || l.prevNonWhiteCh == ']' || l.prevNonWhiteCh == '}' || l.prevNonWhiteCh == '"' || l.prevNonWhiteCh == '0' || l.prevNonWhiteCh == '1' || l.prevNonWhiteCh == '2' || l.prevNonWhiteCh == '3' || l.prevNonWhiteCh == '4' || l.prevNonWhiteCh == '5' || l.prevNonWhiteCh == '6' || l.prevNonWhiteCh == '7' || l.prevNonWhiteCh == '8' || l.prevNonWhiteCh == '9') {
			l.justSkippedNewline = false // we will use this
			return Token{Type: SEMICOLON, Literal: ";"}
		}
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(ASSIGN, l.ch)
		}
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: PLUS_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(PLUS, l.ch)
		}
	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MINUS_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(MINUS, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MUL_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(ASTERISK, l.ch)
		}
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: DIV_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(SLASH, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(BANG, l.ch)
		}
	case '<':
		tok = newToken(LT, l.ch)
	case '>':
		tok = newToken(GT, l.ch)
	case ';':
		tok = newToken(SEMICOLON, l.ch)
	case ':':
		tok = newToken(COLON, l.ch)
	case '.':
		tok = newToken(DOT, l.ch)
	case ',':
		tok = newToken(COMMA, l.ch)
	case '(':
		tok = newToken(LPAREN, l.ch)
	case ')':
		tok = newToken(RPAREN, l.ch)
	case '{':
		tok = newToken(LBRACE, l.ch)
	case '}':
		tok = newToken(RBRACE, l.ch)
	case '[':
		tok = newToken(LBRACKET, l.ch)
	case ']':
		tok = newToken(RBRACKET, l.ch)
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
	case 'f':
		if l.peekChar() == '"' {
			// f"..."
			l.readChar() // skip 'f'
			tok.Type = FSTRING
			tok.Literal = l.readString()
			l.readChar() // skip closing "
			return tok
		}
		// normal identifier starting with f
		tok.Literal = l.readIdentifier()
		tok.Type = lookupIdent(tok.Literal)
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type, tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readChar() {
	// Save the current character before moving to the next one
	if l.ch != ' ' && l.ch != '\t' && l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.prevNonWhiteCh = l.ch
	}
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() (TokenType, string) {
	position := l.position
	isFloat := false
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	if isFloat {
		return FLOAT, l.input[position:l.position]
	}
	return INT, l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	l.justSkippedNewline = false
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' || l.ch == '\r' {
			l.justSkippedNewline = true
		}
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isIdentifierChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

func newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}

