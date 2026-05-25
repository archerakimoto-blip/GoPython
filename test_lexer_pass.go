package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	input := `class Dog: { pass }`
	l := lexer.New(input)
	
	for {
		tok := l.NextToken()
		fmt.Printf("Token: Type=%s, Literal=%s\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
