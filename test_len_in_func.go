
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
	test := `
def f(n):
    a = range(1, n+1)
    b = len(a)
    return b
print(f(10))
`

	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors())>0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}

	program = desugar.Desugar(program)
	c := compiler.New()

	if err := c.Compile(program); err != nil {
		fmt.Printf("Compile err: %v\n", err)
		return
	}

	fmt.Printf("Bytecode constants:\n")
	bc := c.Bytecode()
	for i, cons := range bc.Constants {
		fmt.Printf("  #%d: Type=%T  Val=%v\n", i, cons, cons)
	}

	fmt.Printf("=== Now running ===\n")
	v := vm.New(bc)
	if err := v.Run(); err != nil {
		fmt.Printf("Run err: %v\n", err)
	}
}
