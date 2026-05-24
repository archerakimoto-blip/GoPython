
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `arr[1:3]`

	l := lexer.New(input)

	fmt.Println("=== Lexer Output ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %v, Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}

	l = lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	fmt.Println("\n=== Parser Errors ===")
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Printf("- %s\n", err)
		}
	}
	fmt.Println("\n=== Program ===")
	fmt.Println(program.String())
}
