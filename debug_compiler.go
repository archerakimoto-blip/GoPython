
package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug_compiler.go <file.py>")
		return
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		return
	}

	fmt.Println("=== Parsed Program ===")
	for _, stmt := range program.Statements {
		fmt.Printf("Stmt: %#v\n", stmt)
		if es, ok := stmt.(*ast.ExpressionStatement); ok {
			fmt.Printf("Expr: %#v\n", es.Expression)
		}
	}

	fmt.Println("\n=== Compiling ===")
	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("Compiler error: %v\n", err)
		return
	}

	bytecode := c.Bytecode()
	fmt.Printf("Compiled constants: %#v\n", bytecode.Constants)
	fmt.Printf("Compiled instructions: %v\n", bytecode.Instructions)

	fmt.Println("\n=== Running ===")
	machine := vm.New(bytecode)
	if err := machine.Run(); err != nil {
		fmt.Printf("VM error: %v\n", err)
		return
	}

	if machine.LastPopped != nil {
		fmt.Printf("Last popped: %v\n", machine.LastPopped.Inspect())
	}
}
