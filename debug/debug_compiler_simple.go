
package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `[1,2,3,4,5][1:3]`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, e := range p.Errors() {
			fmt.Println(e)
		}
		return
	}

	fmt.Println("=== Program ===")
	fmt.Println(program.String())

	program = desugar.Desugar(program)
	fmt.Println("\n=== Desugared ===")
	fmt.Println(program.String())

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		fmt.Println("Compiler err:", err)
		return
	}

	bc := c.Bytecode()
	fmt.Println("\n=== Constants ===")
	for i, o := range bc.Constants {
		fmt.Printf("%d: %s\n", i, o.Inspect())
	}
	fmt.Println("\n=== Instructions ===")
	for _, b := range bc.Instructions {
		fmt.Printf("%d ", b)
	}
	fmt.Println()
}
