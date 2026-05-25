package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/compiler"
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
	
	fmt.Println("=== Instructions ===")
	instructions := c.Bytecode().Instructions
	for i := 0; i < len(instructions); i++ {
		op := compiler.Opcode(instructions[i])
		fmt.Printf("%04d: Opcode %d (0x%02x)", i, op, instructions[i])
		
		if op == compiler.OpConstant || op == compiler.OpGetGlobal || op == compiler.OpSetGlobal ||
			op == compiler.OpArray || op == compiler.OpHash || op == compiler.OpSet ||
			op == compiler.OpJump || op == compiler.OpJumpNotTruthy {
			if i+2 < len(instructions) {
				// Show raw bytes
				fmt.Printf(" | Raw bytes: [0x%02x, 0x%02x]", instructions[i+1], instructions[i+2])
				// Little-endian interpretation (how VM reads it)
				operandLE := int(uint16(instructions[i+2])<<8 | uint16(instructions[i+1]))
				fmt.Printf(" | Little-endian: 0x%04x (%d)", operandLE, operandLE)
				// Big-endian interpretation
				operandBE := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
				fmt.Printf(" | Big-endian: 0x%04x (%d)", operandBE, operandBE)
			}
			i += 2
		} else if op == compiler.OpCall || op == compiler.OpGetLocal || op == compiler.OpSetLocal {
			if i+1 < len(instructions) {
				fmt.Printf(" (%d)", instructions[i+1])
			}
			i += 1
		}
		fmt.Println()
	}
}
