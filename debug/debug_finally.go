package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

var opcodeNames = map[compiler.Opcode]string{
	compiler.OpConstant:       "OpConstant",
	compiler.OpPop:            "OpPop",
	compiler.OpAdd:            "OpAdd",
	compiler.OpSub:            "OpSub",
	compiler.OpMul:            "OpMul",
	compiler.OpDiv:            "OpDiv",
	compiler.OpTrue:           "OpTrue",
	compiler.OpFalse:          "OpFalse",
	compiler.OpEqual:          "OpEqual",
	compiler.OpNotEqual:       "OpNotEqual",
	compiler.OpGreaterThan:    "OpGreaterThan",
	compiler.OpLessThan:       "OpLessThan",
	compiler.OpMinus:          "OpMinus",
	compiler.OpBang:           "OpBang",
	compiler.OpJump:           "OpJump",
	compiler.OpJumpNotTruthy:  "OpJumpNotTruthy",
	compiler.OpNull:           "OpNull",
	compiler.OpGetGlobal:      "OpGetGlobal",
	compiler.OpSetGlobal:      "OpSetGlobal",
	compiler.OpArray:          "OpArray",
	compiler.OpHash:           "OpHash",
	compiler.OpSet:            "OpSet",
	compiler.OpIndex:          "OpIndex",
	compiler.OpSlice:          "OpSlice",
	compiler.OpCall:           "OpCall",
	compiler.OpReturnValue:    "OpReturnValue",
	compiler.OpReturn:         "OpReturn",
	compiler.OpGetLocal:       "OpGetLocal",
	compiler.OpSetLocal:       "OpSetLocal",
	compiler.OpBeginTry:       "OpBeginTry",
	compiler.OpEndTry:         "OpEndTry",
	compiler.OpRaise:          "OpRaise",
	compiler.OpExceptHandler:  "OpExceptHandler",
	compiler.OpFinally:        "OpFinally",
	compiler.OpYield:          "OpYield",
	compiler.OpDupTop:         "OpDupTop",
}

func main() {
	code := `try:
    print("Try block")
    x = 10 / 0
except Exception as e:
    print("Except block")
finally:
    print("Finally block")
print("After everything")
`
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("\t", err)
		}
		return
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}

	bytecode := comp.Bytecode()
	instructions := bytecode.Instructions

	fmt.Printf("Instructions length: %d\n", len(instructions))
	fmt.Printf("Constants: %d\n", len(bytecode.Constants))
	fmt.Println("\nConstants:")
	for i, c := range bytecode.Constants {
		fmt.Printf("  [%d] %s\n", i, c.Inspect())
	}
	fmt.Println("\nInstructions:")

	for i := 0; i < len(instructions); {
		op := compiler.Opcode(instructions[i])
		name, ok := opcodeNames[op]
		if !ok {
			name = fmt.Sprintf("UNKNOWN(0x%02x)", op)
		}
		fmt.Printf("%04d  %s", i, name)

		switch op {
		case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal, compiler.OpArray, compiler.OpHash, compiler.OpSet:
			if i+2 < len(instructions) {
				val := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
				fmt.Printf(" %d", val)
			}
		case compiler.OpCall, compiler.OpGetLocal, compiler.OpSetLocal:
			if i+1 < len(instructions) {
				fmt.Printf(" %d", instructions[i+1])
			}
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			if i+2 < len(instructions) {
				val := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
				fmt.Printf(" -> %d", val)
			}
		case compiler.OpBeginTry:
			if i+4 < len(instructions) {
				val1 := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
				val2 := int(uint16(instructions[i+3])<<8 | uint16(instructions[i+4]))
				fmt.Printf(" exceptCount=%d, hasFinally=%d", val1, val2)
			}
		case compiler.OpExceptHandler:
			if i+4 < len(instructions) {
				val1 := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
				val2 := int(uint16(instructions[i+3])<<8 | uint16(instructions[i+4]))
				fmt.Printf(" typeIdx=%d, varIdx=%d", val1, val2)
				if val1 > 0 && val1 < len(bytecode.Constants) {
					fmt.Printf(" (type=%s)", bytecode.Constants[val1].Inspect())
				}
			}
		case compiler.OpFinally:
			if i+2 < len(instructions) {
				val := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
				fmt.Printf(" finallyEndIP=%d", val)
			}
		}
		fmt.Println()

		step := 1
		switch op {
		case compiler.OpConstant, compiler.OpJump, compiler.OpJumpNotTruthy, compiler.OpGetGlobal, compiler.OpSetGlobal, compiler.OpArray, compiler.OpHash, compiler.OpSet, compiler.OpBeginTry, compiler.OpExceptHandler, compiler.OpFinally:
			step = 3
		case compiler.OpCall, compiler.OpGetLocal, compiler.OpSetLocal:
			step = 2
		}
		i += step
	}
}
