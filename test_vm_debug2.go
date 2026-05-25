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
	v := vm.New(code)
	err = v.Run()
	if err != nil {
		fmt.Printf("VM error: %s\n", err)
		return
	}
	
	fmt.Println("VM completed without error")
	lastPopped := v.LastPoppedStackElem()
	if lastPopped != nil {
		fmt.Printf("Last popped: %s\n", lastPopped.Inspect())
	}
}
