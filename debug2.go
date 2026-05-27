package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	inputCode := "x + 1"
	l := lexer.New(inputCode)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
	}
	fmt.Printf("AST: %#v\n", program)

	c := compiler.New()
	// Enter a scope and define "x"!
	c.symbolTable = compiler.NewEnclosedSymbolTable(c.symbolTable)
	c.symbolTable.Define("x")

	err := c.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
	}

	bc := c.Bytecode()
	fmt.Printf("Constants: %#v\n", bc.Constants)
	fmt.Printf("Instructions: %#v\n", bc.Instructions)
	for i, b := range bc.Instructions {
		fmt.Printf("Instruction %d: 0x%x (%d)\n", i, b, b)
	}

}
