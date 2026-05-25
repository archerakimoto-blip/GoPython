package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/ast"
)

func main() {
	input := `d.bark()`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("  %s\n", msg)
		}
		return
	}
	
	fmt.Printf("Program has %d statements\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
		if es, ok := stmt.(*ast.ExpressionStatement); ok {
			fmt.Printf("  Expression: %T\n", es.Expression)
		}
	}
}
