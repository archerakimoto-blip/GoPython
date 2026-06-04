package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

func main() {
	code := `
try:
    raise ExceptionGroup("eg", [TypeError("err1"), ValueError("err2")])
except* TypeError as e:
    print(e)
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

	bc := comp.Bytecode()
	fmt.Println("=== Constants ===")
	for i, c := range bc.Constants {
		switch v := c.(type) {
		case *compiler.CompiledFunction:
			fmt.Printf("[%d] CompiledFunction (params=%d, locals=%d)\n", i, v.NumParameters, v.NumLocals)
		case *objects.String:
			fmt.Printf("[%d] String(%q)\n", i, v.Value)
		case *objects.Integer:
			fmt.Printf("[%d] Integer(%d)\n", i, v.Value)
		case *objects.Builtin:
			fmt.Printf("[%d] Builtin(%s)\n", i, v.Name)
		default:
			fmt.Printf("[%d] %T\n", i, c)
		}
	}
}
