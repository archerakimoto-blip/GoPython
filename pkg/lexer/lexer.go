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
	LT_EQ    = "<="
	GT_EQ    = ">="
	PLUS_EQ  = "+="
	MINUS_EQ = "-="
	MUL_EQ   = "*="
	DIV_EQ   = "/="
	OR_EQ    = "|="
	AND_EQ   = "&="
	XOR_EQ   = "^="
	LT_LT    = "<<"
	GT_GT    = ">>"
	LT_LT_EQ = "<<="
	GT_GT_EQ = ">>="

	EQ     = "=="
	NOT_EQ = "!="

	COMMA     = ","
	COLON     = ":"
	SEMICOLON = ";"
	DOT       = "."
	ARROW     = "->"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	LBRACKET = "["
	RBRACKET = "]"

	BITOR    = "|"
	BITAND   = "&"
	BITXOR   = "^"
	BITNOT   = "~"

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
	NOT_IN  = "NOT_IN"
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
	IMPORT   = "IMPORT"
	FROM     = "FROM"
	IS       = "IS"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	IS_NOT   = "IS_NOT"
	AT       = "@"
	ASYNC    = "ASYNC"
	AWAIT    = "AWAIT"

	INDENT = "INDENT"
	DEDENT = "DEDENT"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"def":     FUNCTION,
	"let":     LET,
	"true":    TRUE,
	"false":   FALSE,
	"if":      IF,
	"else":    ELSE,
	"return":  RETURN,
	"while":   WHILE,
	"for":     FOR,
	"in":      IN,
	"and":     AND,
	"or":      OR,
	"not":     NOT,
	"None":    NONE,
	"class":   CLASS,
	"lambda":  LAMBDA,
	"try":     TRY,
	"except":  EXCEPT,
	"finally": FINALLY,
	"raise":   RAISE,
	"as":      AS,
	"with":    WITH,
	"yield":   YIELD,
	"pass":    PASS,
	"import":  IMPORT,
	"from":    FROM,
	"is":      IS,
	"async":   ASYNC,
	"await":   AWAIT,
}

type Lexer struct {
	input             string
	position          int
	readPosition      int
	ch                byte
	prevNonWhiteCh    byte
	justSkippedNewline bool
	indentStack       []int
	currentIndent     int
	previousIndent    int
	pendingTokens     []Token
	expectIndent      bool
	lastKeyword       string
	prevTokenType     TokenType
}

func New(input string) *Lexer {
	l := &Lexer{
		input:         input,
		indentStack:   []int{0},
		currentIndent: 0,
		pendingTokens: []Token{},
		expectIndent:  false,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		l.justSkippedNewline = false
		l.prevNonWhiteCh = 0
		l.prevTokenType = tok.Type
		return tok
	}

	l.skipWhitespace()

	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		l.justSkippedNewline = false
		l.prevNonWhiteCh = 0
		l.prevTokenType = tok.Type
		return tok
	}

	// Skip comments (# to end of line)
	for l.ch == '#' {
		for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
			l.readChar()
		}
		l.skipWhitespace()
		if len(l.pendingTokens) > 0 {
			tok := l.pendingTokens[0]
			l.pendingTokens = l.pendingTokens[1:]
			l.justSkippedNewline = false
			l.prevNonWhiteCh = 0
			return tok
		}
	}

	var tok Token

	if l.justSkippedNewline && l.prevNonWhiteCh == ':' {
		newIndent := l.currentIndent
		if newIndent > l.indentStack[len(l.indentStack)-1] {
			l.indentStack = append(l.indentStack, newIndent)
			l.justSkippedNewline = false
			l.prevTokenType = INDENT
			return Token{Type: INDENT, Literal: "INDENT"}
		}
	}

	// Check if previous token was a statement terminator (return, pass, yield, raise, break, continue)
	// In these cases, don't insert automatic semicolon
	isStatementTerminator := l.prevTokenType == RETURN || l.prevTokenType == PASS ||
		l.prevTokenType == YIELD || l.prevTokenType == RAISE ||
		l.prevTokenType == BREAK || l.prevTokenType == CONTINUE

	if l.justSkippedNewline && l.prevNonWhiteCh != ':' && l.prevNonWhiteCh != '@' && l.prevNonWhiteCh != 0 && !isStatementTerminator {
		if isIdentifierChar(l.prevNonWhiteCh) || l.prevNonWhiteCh == ')' || l.prevNonWhiteCh == ']' || l.prevNonWhiteCh == '}' || l.prevNonWhiteCh == '"' || (l.prevNonWhiteCh >= '0' && l.prevNonWhiteCh <= '9') {
			if l.ch != '@' {
				l.justSkippedNewline = false
				// Reset prevTokenType when inserting semicolon
				l.prevTokenType = SEMICOLON
				return Token{Type: SEMICOLON, Literal: ";"}
			}
		}
	}

	// Reset lastKeyword after checking
	l.lastKeyword = ""

	if l.ch == 0 {
		for len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.pendingTokens = append(l.pendingTokens, Token{Type: DEDENT, Literal: "DEDENT"})
		}
		if len(l.pendingTokens) > 0 {
			tok := l.pendingTokens[0]
			l.pendingTokens = l.pendingTokens[1:]
			l.prevTokenType = tok.Type
			return tok
		}
		l.prevTokenType = EOF
		return Token{Type: EOF, Literal: ""}
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
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch)}
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
	case '|':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: OR_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(BITOR, l.ch)
		}
	case '&':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: AND_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(BITAND, l.ch)
		}
	case '^':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: XOR_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(BITXOR, l.ch)
		}
	case '~':
		tok = newToken(BITNOT, l.ch)
	case '<':
		if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = Token{Type: LT_LT_EQ, Literal: string(ch) + "<="}
			} else {
				tok = Token{Type: LT_LT, Literal: string(ch) + string(l.ch)}
			}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(LT, l.ch)
		}
	case '>':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = Token{Type: GT_GT_EQ, Literal: string(ch) + ">="}
			} else {
				tok = Token{Type: GT_GT, Literal: string(ch) + string(l.ch)}
			}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(GT, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(BANG, l.ch)
		}
	case ';':
		tok = newToken(SEMICOLON, l.ch)
	case ':':
		tok = newToken(COLON, l.ch)
	case '.':
		tok = newToken(DOT, l.ch)
	case '@':
		tok = newToken(AT, l.ch)
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
		// Check if this is a triple-quoted string
		if l.peekChar() == '"' && len(l.input) > l.readPosition+1 && l.input[l.readPosition+1] == '"' {
			// It's a triple-quoted string!
			l.readChar() // consume second "
			l.readChar() // consume third "
			tok.Type = STRING
			tok.Literal = l.readTripleQuotedString()
		} else {
			// Regular double-quoted string
			tok.Type = STRING
			tok.Literal = l.readString()
		}
	case '\'':
		// Check if this is a triple-quoted single quote string
		if l.peekChar() == '\'' && len(l.input) > l.readPosition+1 && l.input[l.readPosition+1] == '\'' {
			// It's a triple-quoted single quote string!
			l.readChar() // consume second '
			l.readChar() // consume third '
			tok.Type = STRING
			tok.Literal = l.readTripleQuotedSingleQuoteString()
		} else {
			// Regular single-quoted string
			tok.Type = STRING
			tok.Literal = l.readSingleQuotedString()
		}
	case 'f':
		if l.peekChar() == '"' {
			l.readChar()
			// Check if this is a triple-quoted f-string
			if l.peekChar() == '"' && len(l.input) > l.readPosition+1 && l.input[l.readPosition+1] == '"' {
				l.readChar() // second quote
				l.readChar() // third quote
				tok.Type = FSTRING
				tok.Literal = l.readTripleQuotedString()
				return tok
			} else {
				// Regular double-quoted f-string
				tok.Type = FSTRING
				tok.Literal = l.readString()
				l.readChar()
				return tok
			}
		} else if l.peekChar() == '\'' {
			l.readChar()
			// Check if this is a triple-quoted single quote f-string
			if l.peekChar() == '\'' && len(l.input) > l.readPosition+1 && l.input[l.readPosition+1] == '\'' {
				l.readChar() // second '
				l.readChar() // third '
				tok.Type = FSTRING
				tok.Literal = l.readTripleQuotedSingleQuoteString()
				return tok
			} else {
				// Regular single-quoted f-string
				tok.Type = FSTRING
				tok.Literal = l.readSingleQuotedString()
				l.readChar()
				return tok
			}
		}
		tok.Literal = l.readIdentifier()
		tok.Type = lookupIdent(tok.Literal)
		if tok.Type != IDENT {
			l.lastKeyword = tok.Literal
		} else {
			l.lastKeyword = ""
		}
		return tok
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			
			if tok.Type == NOT && len(l.pendingTokens) == 0 {
				// Try to see if next is "in"
				startPos := l.position
				startReadPos := l.readPosition
				startCh := l.ch
				
				// Read next token
				l.skipWhitespace()
				if isLetter(l.ch) {
					nextIdent := l.readIdentifier()
					if nextIdent == "in" && (l.ch == 0 || l.ch == ')' || l.ch == ']' || l.ch == '}' || l.ch == ',' || l.ch == ';' || l.ch == ':' || l.ch == '\n' || l.ch == '\r' || l.ch == ' ' || l.ch == '\t') {
						l.lastKeyword = ""
						l.prevTokenType = NOT_IN
						return Token{Type: NOT_IN, Literal: "not in"}
					}
				}
				
				// If not, reset to original state
				l.position = startPos
				l.readPosition = startReadPos
				l.ch = startCh
			}
			
			if tok.Type != IDENT {
				l.lastKeyword = tok.Literal
			} else {
				l.lastKeyword = ""
			}
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

func (l *Lexer) Peek2Token() Token {
	if len(l.pendingTokens) >= 2 {
		return l.pendingTokens[1]
	}
	if len(l.pendingTokens) == 1 {
		tok2 := l.NextToken()
		l.pendingTokens = append(l.pendingTokens, tok2)
		return tok2
	}
	tok1 := l.NextToken()
	tok2 := l.NextToken()
	l.pendingTokens = append(l.pendingTokens, tok1, tok2)
	return tok2
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdentifierChar(l.ch) {
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

func (l *Lexer) readTripleQuotedString() string {
	position := l.position // since we already consumed 3 quotes
	for {
		l.readChar()
		if l.ch == '"' && l.peekChar() == '"' && len(l.input) > l.readPosition+1 && l.input[l.readPosition+1] == '"' {
			// Found closing triple quote
			l.readChar()
			l.readChar()
			break
		}
		if l.ch == 0 {
			break // end of input
		}
	}
	return l.input[position:l.position-2] // subtract 2 because we read 2 extra characters
}

func (l *Lexer) readSingleQuotedString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readTripleQuotedSingleQuoteString() string {
	position := l.position // since we already consumed 3 quotes
	for {
		l.readChar()
		if l.ch == '\'' && l.peekChar() == '\'' && len(l.input) > l.readPosition+1 && l.input[l.readPosition+1] == '\'' {
			// Found closing triple single quote
			l.readChar()
			l.readChar()
			break
		}
		if l.ch == 0 {
			break // end of input
		}
	}
	return l.input[position:l.position-2] // subtract 2 because we read 2 extra characters
}

func (l *Lexer) skipWhitespace() {
	l.previousIndent = l.currentIndent
	l.currentIndent = 0

	if l.ch != '\n' && l.ch != '\r' {
		for l.ch == ' ' || l.ch == '\t' {
			l.readChar()
		}
		return
	}

	for l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.justSkippedNewline = true
		}
		l.readChar()
	}

	l.currentIndent = 0
	for l.ch == ' ' || l.ch == '\t' {
		if l.ch == '\t' {
			l.currentIndent += 4
		} else {
			l.currentIndent += 1
		}
		l.readChar()
	}

	if l.justSkippedNewline && l.currentIndent < l.indentStack[len(l.indentStack)-1] {
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > l.currentIndent {
			if l.ch == '@' && l.previousIndent == l.indentStack[len(l.indentStack)-1] {
				break
			}
			l.pendingTokens = append(l.pendingTokens, Token{Type: DEDENT, Literal: "DEDENT"})
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
		}
		l.justSkippedNewline = false
		l.prevNonWhiteCh = 0
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

func (l *Lexer) PushToken(tok Token) {
	// Push token to front of pending tokens so that NextToken() returns it next!
	l.pendingTokens = append([]Token{tok}, l.pendingTokens...)
}

func (l *Lexer) LastKeyword() string {
	return l.lastKeyword
}

func newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}
