package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

var opcodeNames = map[compiler.Opcode]string{
	compiler.OpConstant:        "OpConstant",
	compiler.OpPop:             "OpPop",
	compiler.OpDupTop:          "OpDupTop",
	compiler.OpAdd:             "OpAdd",
	compiler.OpSub:             "OpSub",
	compiler.OpMul:             "OpMul",
	compiler.OpDiv:             "OpDiv",
	compiler.OpTrue:            "OpTrue",
	compiler.OpFalse:           "OpFalse",
	compiler.OpEqual:           "OpEqual",
	compiler.OpNotEqual:        "OpNotEqual",
	compiler.OpGreaterThan:     "OpGreaterThan",
	compiler.OpLessThan:        "OpLessThan",
	compiler.OpMinus:           "OpMinus",
	compiler.OpBang:            "OpBang",
	compiler.OpJump:            "OpJump",
	compiler.OpJumpNotTruthy:   "OpJumpNotTruthy",
	compiler.OpNull:            "OpNull",
	compiler.OpGetGlobal:       "OpGetGlobal",
	compiler.OpSetGlobal:       "OpSetGlobal",
	compiler.OpArray:           "OpArray",
	compiler.OpHash:            "OpHash",
	compiler.OpSet:             "OpSet",
	compiler.OpIndex:           "OpIndex",
	compiler.OpSlice:           "OpSlice",
	compiler.OpCall:            "OpCall",
	compiler.OpReturnValue:     "OpReturnValue",
	compiler.OpReturn:          "OpReturn",
	compiler.OpGetLocal:        "OpGetLocal",
	compiler.OpSetLocal:        "OpSetLocal",
	compiler.OpBeginTry:        "OpBeginTry",
	compiler.OpEndTry:          "OpEndTry",
	compiler.OpRaise:           "OpRaise",
	compiler.OpExceptHandler:   "OpExceptHandler",
	compiler.OpFinally:         "OpFinally",
	compiler.OpYield:           "OpYield",
	compiler.OpEnterContext:    "OpEnterContext",
	compiler.OpExitContext:     "OpExitContext",
	compiler.OpMakeGenerator:   "OpMakeGenerator",
	compiler.OpYieldValue:      "OpYieldValue",
}

func main() {
	code := `try:
    print("Try")
    x = 1 / 0
except:
    print("Except")
finally:
    print("Finally")
print("Done")`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
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

	fmt.Println("Bytecode:")
	i := 0
	for i < len(instructions) {
		pos := i
		op := compiler.Opcode(instructions[i])
		name := opcodeNames[op]
		if name == "" {
			name = fmt.Sprintf("UNKNOWN(0x%02x)", op)
		}

		fmt.Printf("%04d: %-20s", pos, name)

		switch op {
		case compiler.OpBeginTry:
			exceptCount := int(uint16(instructions[i+2])<<8 | uint16(instructions[i+1]))
			hasFinally := int(uint16(instructions[i+4])<<8 | uint16(instructions[i+3]))
			fmt.Printf("[exceptCount=%d, hasFinally=%d]", exceptCount, hasFinally)
			i += 5
		case compiler.OpExceptHandler:
			typeIdx := int(uint16(instructions[i+2])<<8 | uint16(instructions[i+1]))
			varIdx := int(uint16(instructions[i+4])<<8 | uint16(instructions[i+3]))
			fmt.Printf("[typeIdx=%d, varIdx=%d]", typeIdx, varIdx)
			i += 5
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			target := int(uint16(instructions[i+2])<<8 | uint16(instructions[i+1]))
			fmt.Printf("[target=%d]", target)
			i += 3
		case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal, compiler.OpArray, compiler.OpHash, compiler.OpSet:
			idx := int(uint16(instructions[i+2])<<8 | uint16(instructions[i+1]))
			fmt.Printf("[idx=%d]", idx)
			i += 3
		case compiler.OpCall:
			numArgs := int(instructions[i+1])
			fmt.Printf("[numArgs=%d]", numArgs)
			i += 2
		case compiler.OpGetLocal, compiler.OpSetLocal:
			idx := int(instructions[i+1])
			fmt.Printf("[idx=%d]", idx)
			i += 2
		case compiler.OpFinally:
			fmt.Printf("[]")
			i += 3
		default:
			i++
		}
		fmt.Println()
	}

	fmt.Println("\nConstants:")
	for i, c := range bytecode.Constants {
		fmt.Printf("  %d: %s - %s\n", i, c.Type(), c.Inspect())
	}

	fmt.Println("\nRunning...")
	machine := vm.New(bytecode)
	err = machine.Run()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
