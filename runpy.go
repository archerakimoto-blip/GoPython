package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.py>\n", os.Args[0])
		os.Exit(1)
	}

	codeBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	code := string(codeBytes)
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:", p.Errors())
		return
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Println("Compile error:", err)
		return
	}

	bytecode := comp.Bytecode()

	machine := vm.New(bytecode)
	err = machine.Run()
	if err != nil {
		fmt.Println("VM Error:", err)
		fmt.Println("Stack:", machine.GetSP())
		for i := 0; i < machine.GetSP(); i++ {
			obj := machine.GetStack(i)
			if obj != nil {
				fmt.Printf("  [%d] %T: %v\n", i, obj, obj)
			} else {
				fmt.Printf("  [%d] nil\n", i)
			}
		}
		os.Exit(1)
	}
}
