
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	input := "[1,2,3]"
	fmt.Println("Testing input:", input)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println("Parser error:", err)
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

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		fmt.Println("VM error:", err)
		return
	}

	stackTop := machine.LastPoppedStackElem()
	if stackTop != nil {
		fmt.Println("Result:", stackTop.Inspect())
	}
}

