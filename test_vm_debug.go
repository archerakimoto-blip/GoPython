package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	input := `f = lambda x: x + 1
print(f(5))`
	
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		fmt.Printf("Compile error: %v\n", err)
		return
	}
	
	fmt.Println("=== Raw Instructions ===")
	instructions := c.Bytecode().Instructions
	for i := 0; i < len(instructions); i++ {
		fmt.Printf("%04d: 0x%02x\n", i, instructions[i])
	}
	
	fmt.Println("\n=== Decoded Instructions ===")
	for i := 0; i < len(instructions); {
		op := compiler.Opcode(instructions[i])
		fmt.Printf("%04d: Opcode %d (%s)", i, op, op)
		
		if op == compiler.OpConstant || op == compiler.OpGetGlobal || op == compiler.OpSetGlobal ||
			op == compiler.OpArray || op == compiler.OpHash || op == compiler.OpSet ||
			op == compiler.OpJump || op == compiler.OpJumpNotTruthy {
			if i+2 < len(instructions) {
				// Little-endian interpretation
				operand := int(uint16(instructions[i+2])<<8 | uint16(instructions[i+1]))
				fmt.Printf(" | Operand: 0x%04x (%d)", operand, operand)
			}
			i += 3
		} else if op == compiler.OpCall || op == compiler.OpGetLocal || op == compiler.OpSetLocal {
			if i+1 < len(instructions) {
				fmt.Printf(" | Operand: %d", instructions[i+1])
			}
			i += 2
		} else {
			i += 1
		}
		fmt.Println()
	}
	
	fmt.Println("\n=== Running VM ===")
	vm := vm.New(c.Bytecode())
	err = vm.Run()
	if err != nil {
		fmt.Printf("VM error: %v\n", err)
	}
}
