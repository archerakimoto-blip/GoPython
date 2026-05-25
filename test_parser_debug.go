package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := "def test(): { pass }"
	l := lexer.New(input)
	p := parser.New(l)
	
	program := p.ParseProgram()
	
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("  %s\n", msg)
		}
	}
	
	fmt.Printf("Program has %d statements\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T - %s\n", i, stmt, stmt.String())
	}
}
