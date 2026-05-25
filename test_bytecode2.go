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
b = d.bark()
print(b)
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
	
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}
	
	code := comp.Bytecode()
	fmt.Printf("Bytecode Instructions: %v\n", code.Instructions)
	
	// Print all bytes
	for i, b := range code.Instructions {
		fmt.Printf("  [%d] = %d (0x%x)\n", i, b, b)
	}
}
