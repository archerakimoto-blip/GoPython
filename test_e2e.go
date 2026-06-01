package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/vm"
)

func runTest(name, input string) {
	fmt.Printf("=== %s ===\n", name)
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Println("  Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("    %s\n", err)
		}
		return
	}

	desugared := desugar.Desugar(program)
	c := compiler.New()
	err := c.Compile(desugared)
	if err != nil {
		fmt.Printf("  Compile error: %v\n", err)
		return
	}

	bytecode := c.Bytecode()
	machine := vm.New(bytecode)
	result := machine.Run()
	if result != nil {
		fmt.Printf("  VM error: %v\n", result)
		return
	}
	fmt.Printf("  Result: ok\n")
}

func main() {
	runTest("Default param", `def f(x=10):
    return x
f()`)

	runTest("Default override", `def f(x=10):
    return x
f(5)`)

	runTest("Multiple defaults", `def f(a, b=5, c=10):
    return a + b + c
f(1)`)

	runTest("Multiple defaults partial", `def f(a, b=5, c=10):
    return a + b + c
f(1, 2)`)

	runTest("Keyword-only", `def f(a, *, b):
    return a + b
f(1, b=2)`)

	runTest("Default + kwonly", `def f(a, b=10, *, c=20):
    return a + b + c
f(1)`)

	runTest("VarArgs", `def f(*args):
    return args
f(1, 2, 3)`)

	runTest("KwArgs", `def f(**kwargs):
    return kwargs
f(a=1, b=2)`)
}
