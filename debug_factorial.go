
package main

import (
	"fmt"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	test := `def factorial(n):
    result = 1
    for i in range(1, n+1):
        result *= i
    return result
print(factorial(10))`

	fmt.Println("\n=== Testing factorial(10) ===\n")
	
	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}

	fmt.Println("\n=== After desugar ===")
	program = desugar.Desugar(program)
	for _, stmt := range program.Statements {
		fmt.Printf("%v\n", stmt)
	}
	
	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("Compile error: %v\n", err)
		return
	}
	
	fmt.Println("\n=== Constants ===")
	bc := c.Bytecode()
	for i, constVal := range bc.Constants {
		fmt.Printf("%d: %v\n", i, constVal)
	}

	fmt.Println("\n=== Bytecode Instructions ===")
	opNames := map[compiler.Opcode]string{
		compiler.OpConstant:       "OpConstant",
		compiler.OpPop:            "OpPop",
		compiler.OpDupTop:         "OpDupTop",
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
		compiler.OpGreaterThanEqual: "OpGreaterThanEqual",
		compiler.OpLessThanEqual:  "OpLessThanEqual",
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
		compiler.OpGetFree:        "OpGetFree",
		compiler.OpClosure:        "OpClosure",
		compiler.OpBeginTry:       "OpBeginTry",
		compiler.OpEndTry:         "OpEndTry",
		compiler.OpRaise:          "OpRaise",
		compiler.OpExceptHandler:  "OpExceptHandler",
		compiler.OpFinally:        "OpFinally",
		compiler.OpYield:          "OpYield",
		compiler.OpEnterContext:   "OpEnterContext",
		compiler.OpExitContext:    "OpExitContext",
		compiler.OpMakeGenerator:  "OpMakeGenerator",
		compiler.OpYieldValue:     "OpYieldValue",
		compiler.OpCreateClass:    "OpCreateClass",
		compiler.OpCreateClassWithSuper: "OpCreateClassWithSuper",
		compiler.OpGetAttribute:   "OpGetAttribute",
		compiler.OpSetAttribute:   "OpSetAttribute",
		compiler.OpFormatString:   "OpFormatString",
		compiler.OpIndexAssign:    "OpIndexAssign",
		compiler.OpListAppend:     "OpListAppend",
		compiler.OpDictSet:        "OpDictSet",
	}

	ip := 0
	for ip < len(bc.Instructions) {
		op := compiler.Opcode(bc.Instructions[ip])
		name, ok := opNames[op]
		if !ok {
			name = fmt.Sprintf("Unknown(%d)", op)
		}
		fmt.Printf("%04d: %s", ip, name)
		
		// Handle operand size
		switch op {
		case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal,
			compiler.OpArray, compiler.OpHash, compiler.OpSet:
			if ip+2 < len(bc.Instructions) {
				operand := int(uint16(bc.Instructions[ip+1])<<8 | uint16(bc.Instructions[ip+2]))
				fmt.Printf(" (%d)", operand)
			}
			ip += 3
		case compiler.OpGetLocal, compiler.OpSetLocal, compiler.OpCall:
			if ip+1 < len(bc.Instructions) {
				operand := int(bc.Instructions[ip+1])
				fmt.Printf(" (%d)", operand)
			}
			ip += 2
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			if ip+2 < len(bc.Instructions) {
				operand := int(uint16(bc.Instructions[ip+1])<<8 | uint16(bc.Instructions[ip+2]))
				fmt.Printf(" -> %d", operand)
			}
			ip += 3
		case compiler.OpClosure:
			if ip+3 < len(bc.Instructions) {
				constIdx := int(uint16(bc.Instructions[ip+1])<<8 | uint16(bc.Instructions[ip+2]))
				numFree := int(bc.Instructions[ip+3])
				fmt.Printf(" (const=%d, free=%d)", constIdx, numFree)
			}
			ip += 4
		case compiler.OpExceptHandler:
			if ip+4 < len(bc.Instructions) {
				typeIdx := int(uint16(bc.Instructions[ip+1])<<8 | uint16(bc.Instructions[ip+2]))
				varIdx := int(uint16(bc.Instructions[ip+3])<<8 | uint16(bc.Instructions[ip+4]))
				fmt.Printf(" (type=%d, var=%d)", typeIdx, varIdx)
			}
			ip +=5
		case compiler.OpFinally:
			if ip+2 < len(bc.Instructions) {
				endIp := int(uint16(bc.Instructions[ip+1])<<8 | uint16(bc.Instructions[ip+2]))
				fmt.Printf(" (endIp=%d)", endIp)
			}
			ip +=3
		case compiler.OpFormatString:
			if ip+2 < len(bc.Instructions) {
				count := int(uint16(bc.Instructions[ip+1])<<8 | uint16(bc.Instructions[ip+2]))
				fmt.Printf(" (%d parts)", count)
			}
			ip +=3
		default:
			ip +=1
		}
		fmt.Println()
	}

	start := time.Now()
	machine := vm.New(bc)
	err := machine.Run()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("\nExecution error: %v\n", err)
	} else {
		fmt.Println("\n=== Success ===")
	}

	fmt.Printf("\nTime: %v\n", elapsed)
	fmt.Printf("Instructions executed: %d\n", machine.InstructionCount())
}
