package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	code := `
def f():
    x = 10
    return x

print(f())
`
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:", p.Errors())
		return
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Println("Compile error:", err)
		return
	}

	bytecode := comp.Bytecode()

	for i, c := range bytecode.Constants {
		if fn, ok := c.(*compiler.CompiledFunction); ok {
			fmt.Printf("Function [%d] (params=%d, locals=%d):\n", i, fn.NumParameters, fn.NumLocals)
			fmt.Printf("  Raw bytes (%d): ", len(fn.Instructions))
			for _, b := range fn.Instructions {
				fmt.Printf("%02x ", b)
			}
			fmt.Println()
		}
	}
}
