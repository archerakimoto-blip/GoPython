package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `class Dog: {
    def bark(self): {
        return "Woof!"
    }
}
d = Dog()
print(d)
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("  %s\n", msg)
		}
		return
	}
	
	fmt.Printf("Program has %d statements\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
	}
	
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}
	
	code := comp.Bytecode()
	fmt.Printf("Bytecode Instructions: %v\n", code.Instructions)
	fmt.Printf("Bytecode Constants: %d\n", len(code.Constants))
	for i, c := range code.Constants {
		fmt.Printf("  Constant %d: %T = %v\n", i, c, c.Inspect())
	}
}
