package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
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
	
	// Print bytecode with opcode names
	ops := map[byte]string{
		0: "OpConstant", 1: "OpPop", 2: "OpDupTop", 3: "OpAdd", 4: "OpSub",
		5: "OpDiv", 6: "OpTrue", 7: "OpFalse", 8: "OpEqual", 9: "OpNotEqual",
		10: "OpGreaterThan", 11: "OpLessThan", 12: "OpMinus", 13: "OpBang", 14: "OpJump",
		15: "OpJumpNotTruthy", 16: "OpNull", 17: "OpGetGlobal", 18: "OpSetGlobal",
		19: "OpGetLocal", 20: "OpSetLocal", 21: "OpArray", 22: "OpHash", 23: "OpSet",
		24: "OpIndex", 25: "OpSlice", 26: "OpCall", 27: "OpReturnValue", 28: "OpReturn",
		29: "OpGetLocal", 30: "OpSetLocal", 31: "OpGetFree", 32: "OpClosure",
	}
	
	fmt.Printf("Bytecode Instructions (%d):\n", len(code.Instructions))
	for i := 0; i < len(code.Instructions); {
		b := code.Instructions[i]
		name := ops[b]
		if name == "" {
			name = fmt.Sprintf("Op%d", b)
		}
		fmt.Printf("  [%d] %s", i, name)
		switch name {
		case "OpConstant", "OpGetGlobal", "OpSetGlobal", "OpGetLocal", "OpSetLocal", "OpGetFree", "OpClosure":
			fmt.Printf(" %d", code.Instructions[i+1])
			i += 2
		case "OpCall":
			fmt.Printf(" %d", code.Instructions[i+1])
			i += 2
		default:
			fmt.Println()
			i++
		}
	}
	
	// Check Class methods
	for _, c := range code.Constants {
		if cls, ok := c.(*objects.Class); ok {
			fmt.Printf("\nClass: %s\n", cls.Name)
			for name, method := range cls.Methods {
				if fn, ok := method.(*compiler.CompiledFunction); ok {
					fmt.Printf("  Method %s instructions: %v\n", name, fn.Instructions)
				}
			}
		}
	}
}
