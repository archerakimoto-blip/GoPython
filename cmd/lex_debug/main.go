package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	code := `try:
    raise ExceptionGroup("eg", [TypeError("err1"), ValueError("err2")])
except* TypeError as e:
    _caught = e
`
	l := lexer.New(code)
	for {
		tok := l.NextToken()
		fmt.Printf("Type=%q Literal=%q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
