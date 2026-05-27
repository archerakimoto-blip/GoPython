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
	input := "(lambda x: x + 1)(5)"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}

	program = desugar.Desugar(program)

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	bc := c.Bytecode()
	fmt.Printf("Constants:\n")
	for i, cnst := range bc.Constants {
		fmt.Printf("  %d: %T, %#v\n", i, cnst, cnst)
		if fn, ok := cnst.(*compiler.CompiledFunction); ok {
			fmt.Printf("    fn.Instructions: %#v\n", fn.Instructions)
			fmt.Printf("    fn.Instructions bytes: %v\n", []byte(fn.Instructions))
			for j, b := range fn.Instructions {
				fmt.Printf("    %d: 0x%x (%d)\n", j, b, b)
			}
			fmt.Printf("    fn.NumLocals: %d\n", fn.NumLocals)
			fmt.Printf("    fn.NumParameters: %d\n", fn.NumParameters)
		}
	}
	fmt.Printf("Main Instructions: %#v\n", bc.Instructions)
	fmt.Printf("Main Instructions as bytes: %v\n", []byte(bc.Instructions))
	for i, b := range bc.Instructions {
		fmt.Printf("%d: 0x%x (%d)\n", i, b, b)
	}

	machine := vm.New(bc)
	err = machine.Run()
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
	}

	result := machine.LastPoppedStackElem()
	fmt.Printf("Result: %#v\n", result)
}
