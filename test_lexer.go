package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	input := `f = lambda x: x + 1
print(f(5))`
	
	l := lexer.New(input)
	
	for {
		tok := l.NextToken()
		fmt.Printf("Token: %s, Literal: '%s'\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
