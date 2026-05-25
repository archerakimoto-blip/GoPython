package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `f = lambda x: x + 1
print(f(5))`
	
	l := lexer.New(input)
	
	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Token: %s, Literal: '%s'\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
	
	fmt.Println("\n=== Parsing AST ===")
	l2 := lexer.New(input)
	p2 := parser.New(l2)
	program := p2.ParseProgram()
	
	if len(p2.Errors()) != 0 {
		for _, err := range p2.Errors() {
			fmt.Printf("Parser error: %s\n", err)
		}
	}
	
	fmt.Printf("Number of statements: %d\n", len(program.Statements))
	fmt.Printf("AST: %s\n", program.String())
	
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %s\n", i, stmt.String())
	}
}
