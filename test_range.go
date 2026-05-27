package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	code := "print(range(10))"

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}

	fmt.Println("AST after parse:")
	printNode(program, 0)

	program = desugar.Desugar(program)

	fmt.Println("\nAST after desugar:")
	printNode(program, 0)

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	bytecode := comp.Bytecode()
	fmt.Println("\nConstants:")
	for i, c := range bytecode.Constants {
		fmt.Printf("%d: %T %v\n", i, c, c)
	}
}

func printNode(node interface{}, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	switch n := node.(type) {
	case interface{ String() string }:
		fmt.Printf("%s%s\n", prefix, n.String())
	}
}
