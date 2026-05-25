package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	input := `class Animal:{def speak(self): return "Animal"}`
	
	l := lexer.New(input)
	
	for {
		tok := l.NextToken()
		fmt.Printf("Token: %s (%s)\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
