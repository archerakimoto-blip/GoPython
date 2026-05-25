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
	
	// Print each instruction with its opcode name
	ops := map[byte]string{
		0: "OpConstant", 1: "OpAdd", 2: "OpSub", 3: "OpMul", 4: "OpDiv",
		5: "OpPop", 6: "OpPopExec", 7: "OpTrue", 8: "OpFalse",
		17: "OpGetGlobal", 18: "OpSetGlobal", 19: "OpGetLocal",
		20: "OpSetLocal", 21: "OpCreateClass", 22: "OpGetAttribute",
		25: "OpCall", 26: "OpReturnValue", 27: "OpReturn",
	}
	
	for i := 0; i < len(code.Instructions); i++ {
		b := code.Instructions[i]
		if name, ok := ops[b]; ok {
			fmt.Printf("  %d: %s", i, name)
			if name == "OpCall" || name == "OpGetAttribute" || name == "OpGetGlobal" || name == "OpSetGlobal" || name == "OpConstant" {
				fmt.Printf(" operand=%d", code.Instructions[i+1])
				i++
			}
			fmt.Println()
		} else {
			fmt.Printf("  %d: Op%d\n", i, b)
		}
	}
}
