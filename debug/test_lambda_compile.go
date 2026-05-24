package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := "lambda x: x + 1"
	fmt.Println("Testing input:", input)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("  ", err)
		}
		return
	}
	
	fmt.Println("Parsed program:", program.String())
	
	desugared := desugar.Desugar(program)
	fmt.Println("Desugared program:", desugared.String())
	
	c := compiler.New()
	err := c.Compile(desugared)
	if err != nil {
		fmt.Println("Compiler error:", err)
		return
	}
	
	fmt.Println("Compilation successful!")
	
	fmt.Println("Constants:")
	for i, obj := range c.Bytecode().Constants {
		fmt.Printf("  %d: %s (Type: %T)\n", i, obj.Inspect(), obj)
	}
}
