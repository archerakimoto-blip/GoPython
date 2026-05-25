package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
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
	
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}
	
	code := comp.Bytecode()
	fmt.Printf("Bytecode Instructions: %v\n", code.Instructions)
	
	v := vm.New(code)
	err = v.Run()
	if err != nil {
		fmt.Printf("VM error: %s\n", err)
		return
	}
	
	// Access globals through reflection or just check LastPoppedStackElem
	lastPopped := v.LastPoppedStackElem()
	if lastPopped != nil {
		fmt.Printf("Result: %s\n", lastPopped.Inspect())
	}
}
