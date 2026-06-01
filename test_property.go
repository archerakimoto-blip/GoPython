package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	input := `class Test:
    def __init__(self):
        self._x = 42

    @property
    def x(self):
        return self._x

    @x.setter
    def x(self, value):
        self._x = value

t = Test()
print(t.x)
t.x = 100
print(t.x)
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}

	fmt.Printf("Before desugar: %v\n", program.String())

	desugared := desugar.Desugar(program)

	fmt.Printf("\nAfter desugar: %v\n", desugared.String())

	c := compiler.New()
	err := c.Compile(desugared)
	if err != nil {
		fmt.Printf("Compiler error: %v\n", err)
		return
	}

	machine := vm.New(c.Bytecode(), c.Constants())
	err = machine.Run()
	if err != nil {
		fmt.Printf("VM error: %v\n", err)
	}

	fmt.Printf("VM last popped: %v\n", machine.LastPopped())
}
