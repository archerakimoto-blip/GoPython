package lexer

import (
	"testing"
)

// helper: collect all tokens from a lexer
func lexAll(input string) []Token {
	l := New(input)
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}

// ===================== 1. Token Type Constants =====================

func TestTokenConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      TokenType
		expected TokenType
	}{
		{"ILLEGAL", ILLEGAL, "ILLEGAL"},
		{"EOF", EOF, "EOF"},
		{"IDENT", IDENT, "IDENT"},
		{"INT", INT, "INT"},
		{"FLOAT", FLOAT, "FLOAT"},
		{"COMPLEX", COMPLEX, "COMPLEX"},
		{"STRING", STRING, "STRING"},
		{"FSTRING", FSTRING, "FSTRING"},
		{"BYTESTRING", BYTESTRING, "BYTESTRING"},
		{"ASSIGN", ASSIGN, "="},
		{"PLUS", PLUS, "+"},
		{"MINUS", MINUS, "-"},
		{"BANG", BANG, "!"},
		{"ASTERISK", ASTERISK, "*"},
		{"SLASH", SLASH, "/"},
		{"PERCENT", PERCENT, "%"},
		{"LT", LT, "<"},
		{"GT", GT, ">"},
		{"PIPE", PIPE, "|"},
		{"AMPERSAND", AMPERSAND, "&"},
		{"CARET", CARET, "^"},
		{"EQ", EQ, "=="},
		{"NOT_EQ", NOT_EQ, "!="},
		{"COMMA", COMMA, ","},
		{"COLON", COLON, ":"},
		{"SEMICOLON", SEMICOLON, ";"},
		{"DOT", DOT, "."},
		{"AT", AT, "@"},
		{"LPAREN", LPAREN, "("},
		{"RPAREN", RPAREN, ")"},
		{"LBRACE", LBRACE, "{"},
		{"RBRACE", RBRACE, "}"},
		{"LBRACKET", LBRACKET, "["},
		{"RBRACKET", RBRACKET, "]"},
		{"FUNCTION", FUNCTION, "DEF"},
		{"LET", LET, "LET"},
		{"TRUE", TRUE, "TRUE"},
		{"FALSE", FALSE, "FALSE"},
		{"IF", IF, "IF"},
		{"ELIF", ELIF, "ELIF"},
		{"ELSE", ELSE, "ELSE"},
		{"RETURN", RETURN, "RETURN"},
		{"WHILE", WHILE, "WHILE"},
		{"FOR", FOR, "FOR"},
		{"IN", IN, "IN"},
		{"AND", AND, "AND"},
		{"OR", OR, "OR"},
		{"NOT", NOT, "NOT"},
		{"NONE", NONE, "NONE"},
		{"CLASS", CLASS, "CLASS"},
		{"LAMBDA", LAMBDA, "LAMBDA"},
		{"TRY", TRY, "TRY"},
		{"EXCEPT", EXCEPT, "EXCEPT"},
		{"FINALLY", FINALLY, "FINALLY"},
		{"RAISE", RAISE, "RAISE"},
		{"AS", AS, "AS"},
		{"WITH", WITH, "WITH"},
		{"YIELD", YIELD, "YIELD"},
		{"PASS", PASS, "PASS"},
		{"IMPORT", IMPORT, "IMPORT"},
		{"FROM", FROM, "FROM"},
		{"ASYNC", ASYNC, "ASYNC"},
		{"AWAIT", AWAIT, "AWAIT"},
		{"WALRUS", WALRUS, "WALRUS"},
		{"GLOBAL", GLOBAL, "GLOBAL"},
		{"NONLOCAL", NONLOCAL, "NONLOCAL"},
		{"RETURN_TYPE", RETURN_TYPE, "RETURN_TYPE"},
		{"DEL", DEL, "DEL"},
		{"MATCH", MATCH, "MATCH"},
		{"CASE", CASE, "CASE"},
		{"INDENT", INDENT, "INDENT"},
		{"DEDENT", DEDENT, "DEDENT"},
		{"PLUS_EQ", PLUS_EQ, "+="},
		{"MINUS_EQ", MINUS_EQ, "-="},
		{"MUL_EQ", MUL_EQ, "*="},
		{"DIV_EQ", DIV_EQ, "/="},
		{"PERCENT_EQ", PERCENT_EQ, "%="},
		{"FLOOR_DIV", FLOOR_DIV, "//"},
		{"FLOOR_DIV_EQ", FLOOR_DIV_EQ, "//="},
		{"POWER", POWER, "**"},
		{"POWER_EQ", POWER_EQ, "**="},
		{"LSHIFT", LSHIFT, "<<"},
		{"RSHIFT", RSHIFT, ">>"},
		{"LSHIFT_EQ", LSHIFT_EQ, "<<="},
		{"RSHIFT_EQ", RSHIFT_EQ, ">>="},
		{"PIPE_EQ", PIPE_EQ, "|="},
		{"AMPERSAND_EQ", AMPERSAND_EQ, "&="},
		{"CARET_EQ", CARET_EQ, "^="},
		{"VAR_ARGS", VAR_ARGS, "VAR_ARGS"},
		{"KW_ARGS", KW_ARGS, "KW_ARGS"},
		{"ELLIPSIS", ELLIPSIS, "ELLIPSIS"},
	}
	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("token constant %s: got %q, want %q", tt.name, tt.got, tt.expected)
		}
	}
}

// ===================== 2. Empty Input & EOF =====================

func TestEmptyInput(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != EOF {
		t.Errorf("empty input: got %q, want EOF", tok.Type)
	}
	if tok.Literal != "" {
		t.Errorf("empty input literal: got %q, want empty", tok.Literal)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	l := New("   \t  ")
	tok := l.NextToken()
	if tok.Type != EOF {
		t.Errorf("whitespace only: got %q, want EOF", tok.Type)
	}
}

// ===================== 3. Identifiers & Keywords =====================

func TestIdentifiers(t *testing.T) {
	input := "foobar _bar foo123 __init__"
	expected := []Token{
		{IDENT, "foobar"},
		{IDENT, "_bar"},
		{IDENT, "foo123"},
		{IDENT, "__init__"},
		{EOF, ""},
	}
	for i, tok := range lexAll(input) {
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("ident[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"def", FUNCTION},
		{"let", LET},
		{"true", TRUE},
		{"false", FALSE},
		{"if", IF},
		{"elif", ELIF},
		{"else", ELSE},
		{"return", RETURN},
		{"while", WHILE},
		{"for", FOR},
		{"in", IN},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"None", NONE},
		{"class", CLASS},
		{"lambda", LAMBDA},
		{"try", TRY},
		{"except", EXCEPT},
		{"finally", FINALLY},
		{"raise", RAISE},
		{"as", AS},
		{"with", WITH},
		{"yield", YIELD},
		{"pass", PASS},
		{"import", IMPORT},
		{"from", FROM},
		{"async", ASYNC},
		{"await", AWAIT},
		{"global", GLOBAL},
		{"nonlocal", NONLOCAL},
		{"del", DEL},
		{"match", MATCH},
		{"case", CASE},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected {
			t.Errorf("keyword %q: got %q, want %q", tt.input, tok.Type, tt.expected)
		}
	}
}

// ===================== 4. Number Literals =====================

func TestIntegerLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{"0", Token{INT, "0"}},
		{"42", Token{INT, "42"}},
		{"123456789", Token{INT, "123456789"}},
		{"1_000_000", Token{INT, "1000000"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("int %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

func TestFloatLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{"3.14", Token{FLOAT, "3.14"}},
		{"0.5", Token{FLOAT, "0.5"}},
		{"1_000.5_00", Token{FLOAT, "1000.500"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("float %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

func TestHexLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{"0xFF", Token{INT, "0xFF"}},
		{"0x1a2b", Token{INT, "0x1a2b"}},
		{"0XAB", Token{INT, "0XAB"}},
		{"0xFF_FF", Token{INT, "0xFFFF"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("hex %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

func TestBinaryLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{"0b1010", Token{INT, "0b1010"}},
		{"0B1101", Token{INT, "0B1101"}},
		{"0b10_10", Token{INT, "0b1010"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("binary %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

func TestOctalLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{"0o777", Token{INT, "0777"}},
		{"0O123", Token{INT, "0123"}},
		{"0o7_7_7", Token{INT, "0777"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("octal %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

func TestComplexLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{"1j", Token{COMPLEX, "1j"}},
		{"3.14j", Token{COMPLEX, "3.14j"}},
		{"5J", Token{COMPLEX, "5J"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("complex %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

// ===================== 5. String Literals =====================

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{`"hello"`, Token{STRING, "hello"}},
		{`'world'`, Token{STRING, "world"}},
		{`""`, Token{STRING, ""}},
		{`''`, Token{STRING, ""}},
		{`"hello world"`, Token{STRING, "hello world"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("string %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

func TestByteStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{`b"hello"`, Token{BYTESTRING, "hello"}},
		{`b'world'`, Token{BYTESTRING, "world"}},
		{`b""`, Token{BYTESTRING, ""}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("bytestring %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

func TestFStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{`f"hello"`, Token{FSTRING, "hello"}},
		{`f'world'`, Token{FSTRING, "world"}},
		{`f""`, Token{FSTRING, ""}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("fstring %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

// ===================== 6. Operators =====================

func TestOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{"+", []Token{{PLUS, "+"}, {EOF, ""}}},
		{"-", []Token{{MINUS, "-"}, {EOF, ""}}},
		{"*", []Token{{ASTERISK, "*"}, {EOF, ""}}},
		{"/", []Token{{SLASH, "/"}, {EOF, ""}}},
		{"%", []Token{{PERCENT, "%"}, {EOF, ""}}},
		{"**", []Token{{POWER, "**"}, {EOF, ""}}},
		{"//", []Token{{FLOOR_DIV, "//"}, {EOF, ""}}},
		{"<<", []Token{{LSHIFT, "<<"}, {EOF, ""}}},
		{">>", []Token{{RSHIFT, ">>"}, {EOF, ""}}},
		{"&", []Token{{AMPERSAND, "&"}, {EOF, ""}}},
		{"|", []Token{{PIPE, "|"}, {EOF, ""}}},
		{"^", []Token{{CARET, "^"}, {EOF, ""}}},
		{"==", []Token{{EQ, "=="}, {EOF, ""}}},
		{"!=", []Token{{NOT_EQ, "!="}, {EOF, ""}}},
		{"<", []Token{{LT, "<"}, {EOF, ""}}},
		{">", []Token{{GT, ">"}, {EOF, ""}}},
		{"=", []Token{{ASSIGN, "="}, {EOF, ""}}},
		{"+=", []Token{{PLUS_EQ, "+="}, {EOF, ""}}},
		{"-=", []Token{{MINUS_EQ, "-="}, {EOF, ""}}},
		{"*=", []Token{{MUL_EQ, "*="}, {EOF, ""}}},
		{"/=", []Token{{DIV_EQ, "/="}, {EOF, ""}}},
		{"//=", []Token{{FLOOR_DIV_EQ, "/=="}, {EOF, ""}}},
		{"%=", []Token{{PERCENT_EQ, "%="}, {EOF, ""}}},
		{"**=", []Token{{POWER_EQ, "*=="}, {EOF, ""}}},
		{"&=", []Token{{AMPERSAND_EQ, "&="}, {EOF, ""}}},
		{"|=", []Token{{PIPE_EQ, "|="}, {EOF, ""}}},
		{"^=", []Token{{CARET_EQ, "^="}, {EOF, ""}}},
		{"<<=", []Token{{LSHIFT_EQ, "<=="}, {EOF, ""}}},
		{">>=", []Token{{RSHIFT_EQ, ">=="}, {EOF, ""}}},
		{":=", []Token{{WALRUS, ":="}, {EOF, ""}}},
		{"->", []Token{{RETURN_TYPE, "->"}, {EOF, ""}}},
		{"...", []Token{{ELLIPSIS, "..."}, {EOF, ""}}},
	}
	for _, tt := range tests {
		tokens := lexAll(tt.input)
		for i, tok := range tokens {
			if i >= len(tt.expected) {
				t.Errorf("operator %q: extra token (%q, %q)", tt.input, tok.Type, tok.Literal)
				break
			}
			if tok.Type != tt.expected[i].Type || tok.Literal != tt.expected[i].Literal {
				t.Errorf("operator %q[%d]: got (%q, %q), want (%q, %q)", tt.input, i, tok.Type, tok.Literal, tt.expected[i].Type, tt.expected[i].Literal)
			}
		}
	}
}

func TestBangAlone(t *testing.T) {
	l := New("!")
	tok := l.NextToken()
	if tok.Type != BANG {
		t.Errorf("bang alone: got %q, want BANG", tok.Type)
	}
}

// ===================== 7. Delimiters =====================

func TestDelimiters(t *testing.T) {
	input := "()[]{},:.;@"
	expected := []Token{
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{COMMA, ","},
		{COLON, ":"},
		{DOT, "."},
		{SEMICOLON, ";"},
		{AT, "@"},
		{EOF, ""},
	}
	tokens := lexAll(input)
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("delimiter[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("delimiter[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 8. Comments =====================

func TestSingleLineComment(t *testing.T) {
	input := "x # this is a comment\ny"
	tokens := lexAll(input)
	// x, SEMICOLON (newline after identifier), y, EOF
	expected := []Token{
		{IDENT, "x"},
		{SEMICOLON, ";"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("comment[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("comment[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestCommentOnlyLine(t *testing.T) {
	input := "# just a comment\n42"
	tokens := lexAll(input)
	// Comment text sets prevNonWhiteCh, so newline after comment produces SEMICOLON
	expected := []Token{
		{SEMICOLON, ";"},
		{INT, "42"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("comment_only[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("comment_only[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 9. INDENT / DEDENT =====================

func TestIndentDedent(t *testing.T) {
	input := "if x:\n    y\nz"
	tokens := lexAll(input)
	expected := []Token{
		{IF, "if"},
		{IDENT, "x"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "y"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{IDENT, "z"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("indent[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("indent[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestMultipleIndents(t *testing.T) {
	input := "if a:\n    if b:\n        c\n    d\ne"
	tokens := lexAll(input)
	expected := []Token{
		{IF, "if"},
		{IDENT, "a"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IF, "if"},
		{IDENT, "b"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "c"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{IDENT, "d"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{IDENT, "e"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("multi_indent[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("multi_indent[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestDedentAtEOF(t *testing.T) {
	input := "if x:\n    y"
	tokens := lexAll(input)
	expected := []Token{
		{IF, "if"},
		{IDENT, "x"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "y"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("dedent_eof[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("dedent_eof[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestTabIndent(t *testing.T) {
	input := "if x:\n\ty"
	tokens := lexAll(input)
	expected := []Token{
		{IF, "if"},
		{IDENT, "x"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "y"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("tab_indent[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("tab_indent[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 10. Newline => SEMICOLON =====================

func TestNewlineSemicolon(t *testing.T) {
	input := "x\ny"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{SEMICOLON, ";"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("nl_semi[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("nl_semi[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestNewlineAfterRParen(t *testing.T) {
	input := "foo()\nbar"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "foo"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{SEMICOLON, ";"},
		{IDENT, "bar"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("nl_rparen[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("nl_rparen[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestNewlineAfterRBracket(t *testing.T) {
	input := "x[0]\ny"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{LBRACKET, "["},
		{INT, "0"},
		{RBRACKET, "]"},
		{SEMICOLON, ";"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("nl_rbracket[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("nl_rbracket[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestNewlineAfterRBrace(t *testing.T) {
	input := "{}\nx"
	tokens := lexAll(input)
	expected := []Token{
		{LBRACE, "{"},
		{RBRACE, "}"},
		{SEMICOLON, ";"},
		{IDENT, "x"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("nl_rbrace[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("nl_rbrace[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestNewlineAfterString(t *testing.T) {
	input := "\"hello\"\nx"
	tokens := lexAll(input)
	expected := []Token{
		{STRING, "hello"},
		{SEMICOLON, ";"},
		{IDENT, "x"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("nl_string[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("nl_string[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestNewlineAfterNumber(t *testing.T) {
	input := "42\nx"
	tokens := lexAll(input)
	expected := []Token{
		{INT, "42"},
		{SEMICOLON, ";"},
		{IDENT, "x"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("nl_number[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("nl_number[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 11. ILLEGAL tokens =====================

func TestIllegalCharacter(t *testing.T) {
	l := New("$")
	tok := l.NextToken()
	if tok.Type != ILLEGAL {
		t.Errorf("illegal char: got %q, want ILLEGAL", tok.Type)
	}
}

// ===================== 12. Complex Expressions =====================

func TestComplexExpression(t *testing.T) {
	input := `x = 1 + 2 * 3`
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{ASSIGN, "="},
		{INT, "1"},
		{PLUS, "+"},
		{INT, "2"},
		{ASTERISK, "*"},
		{INT, "3"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("complex_expr[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("complex_expr[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestFunctionDef(t *testing.T) {
	input := "def foo(x, y):\n    return x + y"
	tokens := lexAll(input)
	expected := []Token{
		{FUNCTION, "def"},
		{IDENT, "foo"},
		{LPAREN, "("},
		{IDENT, "x"},
		{COMMA, ","},
		{IDENT, "y"},
		{RPAREN, ")"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{RETURN, "return"},
		{IDENT, "x"},
		{PLUS, "+"},
		{IDENT, "y"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("func_def[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("func_def[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestClassDef(t *testing.T) {
	input := "class Foo:\n    pass"
	tokens := lexAll(input)
	expected := []Token{
		{CLASS, "class"},
		{IDENT, "Foo"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{PASS, "pass"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("class_def[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("class_def[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestIfElifElse(t *testing.T) {
	input := "if x:\n    y\nelif z:\n    w\nelse:\n    v"
	tokens := lexAll(input)
	expected := []Token{
		{IF, "if"},
		{IDENT, "x"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "y"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{ELIF, "elif"},
		{IDENT, "z"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "w"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{ELSE, "else"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "v"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("if_elif_else[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("if_elif_else[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestTryExceptFinally(t *testing.T) {
	input := "try:\n    x\nexcept:\n    y\nfinally:\n    z"
	tokens := lexAll(input)
	expected := []Token{
		{TRY, "try"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "x"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{EXCEPT, "except"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "y"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{FINALLY, "finally"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "z"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("try_except[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("try_except[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestWithAs(t *testing.T) {
	input := "with x as y:\n    pass"
	tokens := lexAll(input)
	expected := []Token{
		{WITH, "with"},
		{IDENT, "x"},
		{AS, "as"},
		{IDENT, "y"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{PASS, "pass"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("with_as[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("with_as[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestImportFrom(t *testing.T) {
	input := "from x import y"
	tokens := lexAll(input)
	expected := []Token{
		{FROM, "from"},
		{IDENT, "x"},
		{IMPORT, "import"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("import_from[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("import_from[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestAsyncAwait(t *testing.T) {
	input := "async def foo():\n    await bar()"
	tokens := lexAll(input)
	expected := []Token{
		{ASYNC, "async"},
		{FUNCTION, "def"},
		{IDENT, "foo"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{AWAIT, "await"},
		{IDENT, "bar"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("async_await[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("async_await[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestLambda(t *testing.T) {
	input := "lambda x: x + 1"
	tokens := lexAll(input)
	expected := []Token{
		{LAMBDA, "lambda"},
		{IDENT, "x"},
		{COLON, ":"},
		{IDENT, "x"},
		{PLUS, "+"},
		{INT, "1"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("lambda[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("lambda[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestMatchCase(t *testing.T) {
	input := "match x:\n    case 1:\n        y"
	tokens := lexAll(input)
	expected := []Token{
		{MATCH, "match"},
		{IDENT, "x"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{CASE, "case"},
		{INT, "1"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "y"},
		{DEDENT, "DEDENT"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("match_case[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("match_case[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestGlobalNonlocal(t *testing.T) {
	input := "global x\nnonlocal y"
	tokens := lexAll(input)
	expected := []Token{
		{GLOBAL, "global"},
		{IDENT, "x"},
		{SEMICOLON, ";"},
		{NONLOCAL, "nonlocal"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("global_nonlocal[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("global_nonlocal[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestYieldRaiseDel(t *testing.T) {
	input := "yield x\nraise e\ndel y"
	tokens := lexAll(input)
	expected := []Token{
		{YIELD, "yield"},
		{IDENT, "x"},
		{SEMICOLON, ";"},
		{RAISE, "raise"},
		{IDENT, "e"},
		{SEMICOLON, ";"},
		{DEL, "del"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("yield_raise_del[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("yield_raise_del[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestAndOrNotIs(t *testing.T) {
	input := "x and y or not z"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{AND, "and"},
		{IDENT, "y"},
		{OR, "or"},
		{NOT, "not"},
		{IDENT, "z"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("and_or_not[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("and_or_not[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestTrueFalseNone(t *testing.T) {
	input := "true false None"
	tokens := lexAll(input)
	expected := []Token{
		{TRUE, "true"},
		{FALSE, "false"},
		{NONE, "None"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("true_false_none[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("true_false_none[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestForInWhile(t *testing.T) {
	input := "for x in y:\n    pass\nwhile true:\n    pass"
	tokens := lexAll(input)
	expected := []Token{
		{FOR, "for"},
		{IDENT, "x"},
		{IN, "in"},
		{IDENT, "y"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{PASS, "pass"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{WHILE, "while"},
		{TRUE, "true"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{PASS, "pass"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("for_in_while[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("for_in_while[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestAssertBreakContinue(t *testing.T) {
	input := "assert x\nbreak\ncontinue"
	tokens := lexAll(input)
	// "assert" is not a keyword in this lexer, so it's IDENT
	expected := []Token{
		{IDENT, "assert"},
		{IDENT, "x"},
		{SEMICOLON, ";"},
		// "break" and "continue" are not keywords either
		{IDENT, "break"},
		{SEMICOLON, ";"},
		{IDENT, "continue"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("assert_break[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("assert_break[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 13. f-string / b-string prefix with identifier =====================

func TestFNotFString(t *testing.T) {
	// "foo" starts with 'f' but next char is not a quote, so it's an identifier
	l := New("foo")
	tok := l.NextToken()
	if tok.Type != IDENT {
		t.Errorf("f_not_fstring: got %q, want IDENT", tok.Type)
	}
	if tok.Literal != "foo" {
		t.Errorf("f_not_fstring literal: got %q, want 'foo'", tok.Literal)
	}
}

func TestBNotByteString(t *testing.T) {
	// "bar" starts with 'b' but next char is not a quote, so it's an identifier
	l := New("bar")
	tok := l.NextToken()
	if tok.Type != IDENT {
		t.Errorf("b_not_bytestring: got %q, want IDENT", tok.Type)
	}
	if tok.Literal != "bar" {
		t.Errorf("b_not_bytestring literal: got %q, want 'bar'", tok.Literal)
	}
}

// ===================== 14. Dot edge cases =====================

func TestDotDot(t *testing.T) {
	// ".." in this lexer: first '.' sees next is '.', reads it, then checks if third is '.'
	// Since third is not '.', it produces a single DOT token (consuming both dots)
	l := New("..")
	tok1 := l.NextToken()
	if tok1.Type != DOT {
		t.Errorf("dotdot[0]: got %q, want DOT", tok1.Type)
	}
	tok2 := l.NextToken()
	if tok2.Type != EOF {
		t.Errorf("dotdot[1]: got %q, want EOF", tok2.Type)
	}
}

// ===================== 15. Multiple DEDENT at once =====================

func TestMultipleDedent(t *testing.T) {
	input := "if a:\n    if b:\n        c\nz"
	tokens := lexAll(input)
	expected := []Token{
		{IF, "if"},
		{IDENT, "a"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IF, "if"},
		{IDENT, "b"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IDENT, "c"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{DEDENT, "DEDENT"},
		{IDENT, "z"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("multi_dedent[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("multi_dedent[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 16. Walrus operator =====================

func TestWalrusOperator(t *testing.T) {
	l := New(":=")
	tok := l.NextToken()
	if tok.Type != WALRUS {
		t.Errorf("walrus: got %q, want WALRUS", tok.Type)
	}
	if tok.Literal != ":=" {
		t.Errorf("walrus literal: got %q, want ':='", tok.Literal)
	}
}

// ===================== 17. Return type arrow =====================

func TestReturnTypeArrow(t *testing.T) {
	l := New("->")
	tok := l.NextToken()
	if tok.Type != RETURN_TYPE {
		t.Errorf("return_type: got %q, want RETURN_TYPE", tok.Type)
	}
	if tok.Literal != "->" {
		t.Errorf("return_type literal: got %q, want '->'", tok.Literal)
	}
}

// ===================== 18. Ellipsis =====================

func TestEllipsis(t *testing.T) {
	l := New("...")
	tok := l.NextToken()
	if tok.Type != ELLIPSIS {
		t.Errorf("ellipsis: got %q, want ELLIPSIS", tok.Type)
	}
	if tok.Literal != "..." {
		t.Errorf("ellipsis literal: got %q, want '...'", tok.Literal)
	}
}

// ===================== 19. String with escape-like content =====================

func TestStringWithSpecialChars(t *testing.T) {
	input := `"hello\nworld"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("string_special: got %q, want STRING", tok.Type)
	}
	// The lexer reads raw between quotes; escape sequences are not processed
	if tok.Literal != `hello\nworld` {
		t.Errorf("string_special literal: got %q, want 'hello\\nworld'", tok.Literal)
	}
}

// ===================== 20. Unterminated string =====================

func TestUnterminatedString(t *testing.T) {
	input := `"hello`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("unterminated_string: got %q, want STRING", tok.Type)
	}
	if tok.Literal != "hello" {
		t.Errorf("unterminated_string literal: got %q, want 'hello'", tok.Literal)
	}
}

// ===================== 21. Number followed by dot (not float) =====================

func TestNumberThenDot(t *testing.T) {
	// "42." where the dot is followed by a non-digit — should be INT then DOT
	input := "42.x"
	tokens := lexAll(input)
	expected := []Token{
		{INT, "42"},
		{DOT, "."},
		{IDENT, "x"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("num_dot[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("num_dot[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 22. Multiple statements on one line =====================

func TestMultipleStatementsOnOneLine(t *testing.T) {
	input := "x = 1; y = 2"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{ASSIGN, "="},
		{INT, "1"},
		{SEMICOLON, ";"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{INT, "2"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("multi_stmt[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("multi_stmt[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 23. Decorator =====================

func TestDecorator(t *testing.T) {
	input := "@decorator"
	tokens := lexAll(input)
	expected := []Token{
		{AT, "@"},
		{IDENT, "decorator"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("decorator[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("decorator[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 24. CRLF handling =====================

func TestCRLF(t *testing.T) {
	input := "x\r\ny"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{SEMICOLON, ";"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("crlf[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("crlf[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 25. Minus followed by non-eq non-gt =====================

func TestMinusAlone(t *testing.T) {
	l := New("-")
	tok := l.NextToken()
	if tok.Type != MINUS {
		t.Errorf("minus_alone: got %q, want MINUS", tok.Type)
	}
}

// ===================== 26. LT not followed by < =====================

func TestLTAlone(t *testing.T) {
	l := New("< ")
	tok := l.NextToken()
	if tok.Type != LT {
		t.Errorf("lt_alone: got %q, want LT", tok.Type)
	}
}

// ===================== 27. GT not followed by > =====================

func TestGTAlone(t *testing.T) {
	l := New("> ")
	tok := l.NextToken()
	if tok.Type != GT {
		t.Errorf("gt_alone: got %q, want GT", tok.Type)
	}
}

// ===================== 28. LookupIdent for non-keyword =====================

func TestLookupIdentNonKeyword(t *testing.T) {
	tokType := lookupIdent("foobar")
	if tokType != IDENT {
		t.Errorf("lookup_non_keyword: got %q, want IDENT", tokType)
	}
}

// ===================== 29. stripUnderscores no underscores =====================

func TestStripUnderscoresNoUnderscores(t *testing.T) {
	result := stripUnderscores("12345")
	if result != "12345" {
		t.Errorf("strip_no_underscores: got %q, want '12345'", result)
	}
}

// ===================== 30. Full program-like input =====================

func TestFullProgram(t *testing.T) {
	input := `def factorial(n):
    if n < 2:
        return 1
    return n * factorial(n - 1)`
	tokens := lexAll(input)
	expected := []Token{
		{FUNCTION, "def"},
		{IDENT, "factorial"},
		{LPAREN, "("},
		{IDENT, "n"},
		{RPAREN, ")"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{IF, "if"},
		{IDENT, "n"},
		{LT, "<"},
		{INT, "2"},
		{COLON, ":"},
		{INDENT, "INDENT"},
		{RETURN, "return"},
		{INT, "1"},
		{SEMICOLON, ";"},
		{DEDENT, "DEDENT"},
		{RETURN, "return"},
		{IDENT, "n"},
		{ASTERISK, "*"},
		{IDENT, "factorial"},
		{LPAREN, "("},
		{IDENT, "n"},
		{MINUS, "-"},
		{INT, "1"},
		{RPAREN, ")"},
		{DEDENT, "DEDENT"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("full_program[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("full_program[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 31. Newline after colon (no indent — same level) =====================

func TestNewlineAfterColonNoIndent(t *testing.T) {
	// After colon, if indent doesn't increase, no INDENT token
	input := "if x:\ny"
	tokens := lexAll(input)
	expected := []Token{
		{IF, "if"},
		{IDENT, "x"},
		{COLON, ":"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("colon_no_indent[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("colon_no_indent[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 32. Comment at end of file =====================

func TestCommentAtEOF(t *testing.T) {
	input := "42 # comment"
	tokens := lexAll(input)
	expected := []Token{
		{INT, "42"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("comment_eof[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("comment_eof[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 33. Newline not after identifier-like =====================

func TestNewlineAfterColon(t *testing.T) {
	// Newline after colon should NOT produce SEMICOLON
	input := "if x:\n    y"
	tokens := lexAll(input)
	for _, tok := range tokens {
		if tok.Type == SEMICOLON {
			t.Errorf("newline_after_colon: unexpected SEMICOLON token")
		}
	}
}

// ===================== 34. b-string with double quotes =====================

func TestByteStringDoubleQuote(t *testing.T) {
	l := New(`b"test"`)
	tok := l.NextToken()
	if tok.Type != BYTESTRING {
		t.Errorf("bs_dquote: got %q, want BYTESTRING", tok.Type)
	}
	if tok.Literal != "test" {
		t.Errorf("bs_dquote literal: got %q, want 'test'", tok.Literal)
	}
}

// ===================== 35. f-string with single quotes =====================

func TestFStringSingleQuote(t *testing.T) {
	l := New(`f'test'`)
	tok := l.NextToken()
	if tok.Type != FSTRING {
		t.Errorf("fs_squote: got %q, want FSTRING", tok.Type)
	}
	if tok.Literal != "test" {
		t.Errorf("fs_squote literal: got %q, want 'test'", tok.Literal)
	}
}

// ===================== 36. b-string with single quotes =====================

func TestByteStringSingleQuote(t *testing.T) {
	l := New(`b'test'`)
	tok := l.NextToken()
	if tok.Type != BYTESTRING {
		t.Errorf("bs_squote: got %q, want BYTESTRING", tok.Type)
	}
	if tok.Literal != "test" {
		t.Errorf("bs_squote literal: got %q, want 'test'", tok.Literal)
	}
}

// ===================== 37. Operator combinations =====================

func TestOperatorCombinations(t *testing.T) {
	input := "x += y -= z *= w /= v"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{PLUS_EQ, "+="},
		{IDENT, "y"},
		{MINUS_EQ, "-="},
		{IDENT, "z"},
		{MUL_EQ, "*="},
		{IDENT, "w"},
		{DIV_EQ, "/="},
		{IDENT, "v"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("op_combo[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("op_combo[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestBitwiseCompoundAssign(t *testing.T) {
	input := "x &= y |= z ^= w"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{AMPERSAND_EQ, "&="},
		{IDENT, "y"},
		{PIPE_EQ, "|="},
		{IDENT, "z"},
		{CARET_EQ, "^="},
		{IDENT, "w"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("bitwise_compound[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("bitwise_compound[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

func TestShiftCompoundAssign(t *testing.T) {
	input := "x <<= y >>= z"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{LSHIFT_EQ, "<=="},
		{IDENT, "y"},
		{RSHIFT_EQ, ">=="},
		{IDENT, "z"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("shift_compound[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("shift_compound[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 38. Single-quoted string =====================

func TestSingleQuotedString(t *testing.T) {
	l := New(`'hello'`)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("single_quote_string: got %q, want STRING", tok.Type)
	}
	if tok.Literal != "hello" {
		t.Errorf("single_quote_string literal: got %q, want 'hello'", tok.Literal)
	}
}

// ===================== 39. Multiple tokens on same line =====================

func TestMultipleTokensSameLine(t *testing.T) {
	input := "x + y * z"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{PLUS, "+"},
		{IDENT, "y"},
		{ASTERISK, "*"},
		{IDENT, "z"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("same_line[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("same_line[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 40. Number with underscore in different positions =====================

func TestNumberUnderscores(t *testing.T) {
	tests := []struct {
		input    string
		expected Token
	}{
		{"1_2", Token{INT, "12"}},
		{"1_000", Token{INT, "1000"}},
		{"0x1_AB", Token{INT, "0x1AB"}},
		{"0b1_0_1", Token{INT, "0b101"}},
		{"0o7_7", Token{INT, "077"}},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected.Type || tok.Literal != tt.expected.Literal {
			t.Errorf("num_underscore %q: got (%q, %q), want (%q, %q)", tt.input, tok.Type, tok.Literal, tt.expected.Type, tt.expected.Literal)
		}
	}
}

// ===================== 41. Float with underscore =====================

func TestFloatUnderscore(t *testing.T) {
	l := New("1_000.5_00")
	tok := l.NextToken()
	if tok.Type != FLOAT {
		t.Errorf("float_underscore: got %q, want FLOAT", tok.Type)
	}
	if tok.Literal != "1000.500" {
		t.Errorf("float_underscore literal: got %q, want '1000.500'", tok.Literal)
	}
}

// ===================== 42. Complex with underscore =====================

func TestComplexUnderscore(t *testing.T) {
	l := New("1_000j")
	tok := l.NextToken()
	if tok.Type != COMPLEX {
		t.Errorf("complex_underscore: got %q, want COMPLEX", tok.Type)
	}
	if tok.Literal != "1000j" {
		t.Errorf("complex_underscore literal: got %q, want '1000j'", tok.Literal)
	}
}

// ===================== 43. Pending tokens (DEDENT queue) =====================

func TestPendingTokensDedentQueue(t *testing.T) {
	// When multiple DEDENTs are queued, they should be returned one by one
	input := "if a:\n    if b:\n        if c:\n            d\nz"
	tokens := lexAll(input)
	// Count DEDENT tokens
	dedentCount := 0
	for _, tok := range tokens {
		if tok.Type == DEDENT {
			dedentCount++
		}
	}
	if dedentCount != 3 {
		t.Errorf("pending_dedent: got %d DEDENTs, want 3", dedentCount)
	}
}

// ===================== 44. readString (double-quote only) =====================

func TestReadString(t *testing.T) {
	// readString is used internally; test via the public API with double quotes
	l := New(`"abc"`)
	tok := l.NextToken()
	if tok.Type != STRING || tok.Literal != "abc" {
		t.Errorf("read_string: got (%q, %q), want (STRING, 'abc')", tok.Type, tok.Literal)
	}
}

// ===================== 45. Newline after non-identifier-like char (no SEMICOLON) =====================

func TestNewlineAfterOperator(t *testing.T) {
	// After '+', newline should NOT produce SEMICOLON
	input := "x +\n    y"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{PLUS, "+"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("nl_after_op[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("nl_after_op[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 46. isLetter, isDigit, isIdentifierChar coverage =====================

func TestIsLetter(t *testing.T) {
	tests := []struct {
		ch       byte
		expected bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'_', true},
		{'0', false},
		{'$', false},
	}
	for _, tt := range tests {
		result := isLetter(tt.ch)
		if result != tt.expected {
			t.Errorf("isLetter(%q): got %v, want %v", tt.ch, result, tt.expected)
		}
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		ch       byte
		expected bool
	}{
		{'0', true},
		{'9', true},
		{'a', false},
		{' ', false},
	}
	for _, tt := range tests {
		result := isDigit(tt.ch)
		if result != tt.expected {
			t.Errorf("isDigit(%q): got %v, want %v", tt.ch, result, tt.expected)
		}
	}
}

func TestIsHexDigit(t *testing.T) {
	tests := []struct {
		ch       byte
		expected bool
	}{
		{'0', true},
		{'9', true},
		{'a', true},
		{'f', true},
		{'A', true},
		{'F', true},
		{'g', false},
		{'G', false},
	}
	for _, tt := range tests {
		result := isHexDigit(tt.ch)
		if result != tt.expected {
			t.Errorf("isHexDigit(%q): got %v, want %v", tt.ch, result, tt.expected)
		}
	}
}

func TestIsBinaryDigit(t *testing.T) {
	tests := []struct {
		ch       byte
		expected bool
	}{
		{'0', true},
		{'1', true},
		{'2', false},
		{'a', false},
	}
	for _, tt := range tests {
		result := isBinaryDigit(tt.ch)
		if result != tt.expected {
			t.Errorf("isBinaryDigit(%q): got %v, want %v", tt.ch, result, tt.expected)
		}
	}
}

func TestIsOctalDigit(t *testing.T) {
	tests := []struct {
		ch       byte
		expected bool
	}{
		{'0', true},
		{'7', true},
		{'8', false},
		{'a', false},
	}
	for _, tt := range tests {
		result := isOctalDigit(tt.ch)
		if result != tt.expected {
			t.Errorf("isOctalDigit(%q): got %v, want %v", tt.ch, result, tt.expected)
		}
	}
}

func TestIsIdentifierChar(t *testing.T) {
	tests := []struct {
		ch       byte
		expected bool
	}{
		{'a', true},
		{'0', true},
		{'_', true},
		{'$', false},
		{' ', false},
	}
	for _, tt := range tests {
		result := isIdentifierChar(tt.ch)
		if result != tt.expected {
			t.Errorf("isIdentifierChar(%q): got %v, want %v", tt.ch, result, tt.expected)
		}
	}
}

// ===================== 47. newToken helper =====================

func TestNewToken(t *testing.T) {
	tok := newToken(IDENT, 'x')
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("newToken: got (%q, %q), want (IDENT, 'x')", tok.Type, tok.Literal)
	}
}

// ===================== 48. Multiple consecutive newlines =====================

func TestMultipleNewlines(t *testing.T) {
	input := "x\n\n\ny"
	tokens := lexAll(input)
	expected := []Token{
		{IDENT, "x"},
		{SEMICOLON, ";"},
		{IDENT, "y"},
		{EOF, ""},
	}
	for i, tok := range tokens {
		if i >= len(expected) {
			t.Errorf("multi_nl[%d]: extra token (%q, %q)", i, tok.Type, tok.Literal)
			break
		}
		if tok.Type != expected[i].Type || tok.Literal != expected[i].Literal {
			t.Errorf("multi_nl[%d]: got (%q, %q), want (%q, %q)", i, tok.Type, tok.Literal, expected[i].Type, expected[i].Literal)
		}
	}
}

// ===================== 49. String at EOF (no closing quote) =====================

func TestStringAtEOF(t *testing.T) {
	l := New(`"abc`)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("string_eof: got %q, want STRING", tok.Type)
	}
	if tok.Literal != "abc" {
		t.Errorf("string_eof literal: got %q, want 'abc'", tok.Literal)
	}
}

// ===================== 50. Single-quoted unterminated string =====================

func TestSingleQuotedUnterminatedString(t *testing.T) {
	l := New(`'abc`)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("squote_unterminated: got %q, want STRING", tok.Type)
	}
	if tok.Literal != "abc" {
		t.Errorf("squote_unterminated literal: got %q, want 'abc'", tok.Literal)
	}
}
