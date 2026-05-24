
package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	input := `arr = [1, 2, 3, 4, 5]; arr[1:3]`

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
	fmt.Println("\n=== Desugared Program ===")
	fmt.Println(program.String())

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		fmt.Println("Compiler error:", err)
		return
	}

	bytecode := c.Bytecode()
	fmt.Println("\n=== Constants ===")
	for i, obj := range bytecode.Constants {
		fmt.Printf("%d: %v\n", i, obj.Inspect())
	}

	fmt.Println("\n=== Instructions ===")
	fmt.Printf("%v\n", bytecode.Instructions)

	fmt.Println("\n=== VM Execution ===")
	machine := vm.New(bytecode)
	err = machine.Run()
	if err != nil {
		fmt.Println("VM error:", err)
		return
	}
	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		fmt.Println("Result:", lastPopped.Inspect())
	}
}
