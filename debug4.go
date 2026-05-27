package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := "(lambda x: x + 1)(5)"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) !=0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
	}
	fmt.Printf("Original AST: %#v\n", program)

	program = desugar.Desugar(program)
	fmt.Printf("Desugared AST: %#v\n", program)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	bc := comp.Bytecode()
	fmt.Println("=== Compiler Output ===")
	fmt.Printf("Constants: len=%d\n", len(bc.Constants))
	for i, cnst := range bc.Constants {
		if cf, ok := cnst.(*compiler.CompiledFunction); ok {
			fmt.Printf("\n  [%d] CompiledFunction\n", i)
			fmt.Printf("    NumLocals: %v\n", cf.NumLocals)
			fmt.Printf("    NumParams: %v\n", cf.NumParameters)
			fmt.Printf("    Instructions len: %v bytes\n", len(cf.Instructions))
			for j, b := range cf.Instructions {
				fmt.Printf("    %02d: 0x%02x (%3d)\n", j, b, b)
			}
		} else {
			fmt.Printf("  [%d] %T: %v\n", i, cnst, cnst)
		}
	}
	fmt.Printf("\nMain Instructions len: %v bytes\n", len(bc.Instructions))
	for i, b := range bc.Instructions {
		fmt.Printf("  %02d: 0x%02x (%3d)\n", i, b, b)
	}

}
