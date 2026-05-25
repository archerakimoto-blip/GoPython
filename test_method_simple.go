package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	input := `class Dog: {
    def bark(self): {
        return "Woof!"
    }
}
d = Dog()
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
	v := vm.New(code)
	err = v.Run()
	if err != nil {
		fmt.Printf("VM error: %s\n", err)
		return
	}
	
	fmt.Println("Dog instance created")
	
	// Now test method call separately
	input2 := `d.bark()`
	l2 := lexer.New(input2)
	p2 := parser.New(l2)
	program2 := p2.ParseProgram()
	
	if len(p2.Errors()) != 0 {
		fmt.Println("Parser errors for method call:")
		for _, msg := range p2.Errors() {
			fmt.Printf("  %s\n", msg)
		}
		return
	}
	
	fmt.Println("Method call parsed successfully")
	fmt.Println("Method call type:", program2.Statements[0])
}
