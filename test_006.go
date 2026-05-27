
package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	code, err := os.ReadFile("tests/benchmarks/006_string.py")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	l := lexer.New(string(code))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, e := range p.Errors() {
			fmt.Printf("- %s\n", e)
		}
		return
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	machine := vm.New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}
}
