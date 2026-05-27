package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := "_result = 1 + 2"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}
	program = desugar.Desugar(program)
	fmt.Printf("AST after desugar: %#v\n", program)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	bc := comp.Bytecode()
	fmt.Println("=== Compiler Output ===")
	fmt.Printf("Constants (%d):\n", len(bc.Constants))
	for i, cnst := range bc.Constants {
		fmt.Printf("  [%d] %T: %v\n", i, cnst, cnst)
	}
	fmt.Printf("\nInstructions (%d bytes):\n", len(bc.Instructions))
	for i := 0; i < len(bc.Instructions); i++ {
		fmt.Printf("  %d: 0x%02x (%3d)\n", i, bc.Instructions[i], bc.Instructions[i])
	}
}
