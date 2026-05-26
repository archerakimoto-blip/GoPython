package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `class Animal:
    def speak(self):
        return "Animal sound"
`
	fmt.Println("Input code:")
	fmt.Println(input)

	fmt.Println("\n--- Lexer tokens ---")
	l := lexer.New(input)
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %-10s Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}

	fmt.Println("\n--- Parser AST ---")
	l2 := lexer.New(input)
	p := parser.New(l2)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
	}

	fmt.Printf("\nProgram has %d statements\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("\nStatement %d: %T\n", i, stmt)
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			fmt.Printf("  Expression: %T\n", exprStmt.Expression)
		} else {
			fmt.Printf("  Type: %T\n", stmt)
		}
	}
}
