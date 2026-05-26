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

	fmt.Println("\n--- Parser AST with debugging ---")
	l := lexer.New(input)
	p := parser.New(l)
	
	// Monkey patch to add debug output
	originalNextToken := p.nextToken
	p.nextToken = func() {
		fmt.Printf("  [DBG] nextToken: cur=%s(%q), peek=%s(%q) -> ", 
			p.curToken.Type, p.curToken.Literal,
			p.peekToken.Type, p.peekToken.Literal)
		originalNextToken()
		fmt.Printf("cur=%s(%q), peek=%s(%q)\n", 
			p.curToken.Type, p.curToken.Literal,
			p.peekToken.Type, p.peekToken.Literal)
	}

	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
		return
	}

	fmt.Printf("\nProgram has %d statements\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("\nStatement %d: %T\n", i, stmt)
		if clsStmt, ok := stmt.(*ast.ClassStatement); ok {
			fmt.Printf("  Class: %s\n", clsStmt.Name.Value)
			fmt.Printf("  Methods: %d\n", len(clsStmt.Methods))
			for _, m := range clsStmt.Methods {
				fmt.Printf("    - %s\n", m.Name)
			}
		}
	}
}