
package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run debug_compiler.go <filename>")
		return
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("-", err)
		}
		return
	}

	program = desugar.Desugar(program)

	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		fmt.Println("Compiler error:", err)
	} else {
		fmt.Println("Compilation successful!")
	}
}
