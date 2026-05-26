package main

import (
	"fmt"
)

type TokenType string

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	STRING TokenType = "STRING"

	ASSIGN TokenType = "="
	PLUS   TokenType = "+"
	MINUS  TokenType = "-"
	EQ     TokenType = "=="
	NOT_EQ TokenType = "!="
	LT     TokenType = "<"
	GT     TokenType = ">"

	COMMA     TokenType = ","
	COLON     TokenType = ":"
	SEMICOLON TokenType = ";"

	LPAREN TokenType = "("
	RPAREN TokenType = ")"
	LBRACE TokenType = "{"
	RBRACE TokenType = "}"

	FUNCTION TokenType = "DEF"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	RETURN   TokenType = "RETURN"
	CLASS    TokenType = "CLASS"

	INDENT TokenType = "INDENT"
	DEDENT TokenType = "DEDENT"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"def":    FUNCTION,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"class":  CLASS,
}

type Lexer struct {
	input             string
	position          int
	readPosition      int
	ch                byte
	indentStack       []int
	currentIndent     int
	pendingTokens     []Token
}

func New(input string) *Lexer {
	l := &Lexer{
		input:         input,
		indentStack:   []int{0},
		currentIndent: 0,
		pendingTokens: []Token{},
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() Token {
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	l.skipWhitespace()

	var tok Token

	switch l.ch {
	case 0:
		for len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			l.pendingTokens = append(l.pendingTokens, Token{Type: DEDENT, Literal: "DEDENT"})
		}
		if len(l.pendingTokens) > 0 {
			tok := l.pendingTokens[0]
			l.pendingTokens = l.pendingTokens[1:]
			return tok
		}
		tok.Type = EOF
		tok.Literal = ""
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(PLUS, l.ch)
	case '-':
		tok = newToken(MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	case '<':
		tok = newToken(LT, l.ch)
	case '>':
		tok = newToken(GT, l.ch)
	case ';':
		tok = newToken(SEMICOLON, l.ch)
	case ':':
		tok = newToken(COLON, l.ch)
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
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	l.currentIndent = 0

	// 跳过换行符和前面的空格，计算新的缩进级别
	if l.ch == '\n' || l.ch == '\r' {
		// 跳过换行符
		for l.ch == '\n' || l.ch == '\r' {
			l.readChar()
		}

		// 计算缩进（空格或 tab 的数量）
		l.currentIndent = 0
		for l.ch == ' ' || l.ch == '\t' {
			if l.ch == '\t' {
				l.currentIndent += 4
			} else {
				l.currentIndent += 1
			}
			l.readChar()
		}

		// 处理缩进变化
		lastIndent := l.indentStack[len(l.indentStack)-1]
		if l.currentIndent > lastIndent {
			l.indentStack = append(l.indentStack, l.currentIndent)
			l.pendingTokens = append(l.pendingTokens, Token{Type: INDENT, Literal: "INDENT"})
		} else if l.currentIndent < lastIndent {
			for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > l.currentIndent {
				l.pendingTokens = append(l.pendingTokens, Token{Type: DEDENT, Literal: "DEDENT"})
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
			}
		}
	} else {
		// 没有换行符，直接跳过当前行的空格
		for l.ch == ' ' || l.ch == '\t' {
			l.readChar()
		}
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
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

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
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

func main() {
	input := `class Animal:
    def speak(self):
        return "Animal sound"

def add(x, y):
    return x + y

result = add(3, 5)
print(result)
`

	fmt.Println("Test Input:")
	fmt.Println(input)

	l := New(input)

	fmt.Println("\nTokens:")
	for i := 0; i < 100; i++ {
		tok := l.NextToken()
		fmt.Printf("Token %2d: Type: %-10s Literal: %q\n", i, tok.Type, tok.Literal)
		if tok.Type == EOF {
			break
		}
	}
}
