package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
)

var opNames = map[byte]string{
	0: "OpConstant", 1: "OpPop", 2: "OpDupTop", 3: "OpAdd", 4: "OpSub",
	5: "OpMul", 6: "OpDiv", 7: "OpMod", 8: "OpFloorDiv", 9: "OpPower",
	10: "OpTrue", 11: "OpFalse", 12: "OpEqual", 13: "OpNotEqual",
	14: "OpGreaterThan", 15: "OpLessThan", 16: "OpMinus", 17: "OpBang",
	18: "OpJump", 19: "OpJumpNotTruthy", 20: "OpNull", 21: "OpGetGlobal",
	22: "OpSetGlobal", 23: "OpArray", 24: "OpHash", 25: "OpSet",
	26: "OpIndex", 27: "OpSlice", 28: "OpCall", 29: "OpReturnValue",
	30: "OpReturn", 31: "OpGetLocal", 32: "OpSetLocal", 33: "OpGetFree",
	34: "OpClosure", 35: "OpBeginTry", 36: "OpEndTry", 37: "OpRaise",
	38: "OpExceptHandler", 39: "OpExceptStarHandler", 40: "OpFinally", 41: "OpYield",
	42: "OpEnterContext", 43: "OpExitContext", 44: "OpMakeGenerator",
	45: "OpYieldValue", 46: "OpCreateClass", 47: "OpCreateClassWithSuper",
	48: "OpCreateClassWithMultiSuper", 49: "OpGetAttribute", 50: "OpSetAttribute",
	51: "OpFormatString", 52: "OpMakeAsync", 53: "OpAwait",
	54: "OpListUnpack", 55: "OpDictUnpack",
	56: "OpEllipsis",
}

func disassemble(ins []byte, constants []objects.Object) {
	for i := 0; i < len(ins); {
		op := ins[i]
		name := opNames[op]
		if name == "" {
			name = fmt.Sprintf("Unknown(%d)", op)
		}
		fmt.Printf("  %04d: %s", i, name)

		switch op {
		case 0, 18, 19, 21, 22, 23, 24, 25, 45, 46, 47, 48, 49, 50, 40:
			if i+2 < len(ins) {
				operand := int(uint16(ins[i+1])<<8 | uint16(ins[i+2]))
				fmt.Printf(" %d", operand)
				if op == 0 && operand < len(constants) {
					fmt.Printf(" (%T)", constants[operand])
				}
			}
			i += 3
		case 34:
			if i+3 < len(ins) {
				constIdx := int(uint16(ins[i+1])<<8 | uint16(ins[i+2]))
				numFree := int(ins[i+3])
				fmt.Printf(" const=%d free=%d", constIdx, numFree)
			}
			i += 4
		case 35, 38, 39:
			if i+4 < len(ins) {
				fmt.Printf(" %d %d %d %d", ins[i+1], ins[i+2], ins[i+3], ins[i+4])
			}
			i += 5
		case 28, 31, 32, 33:
			if i+1 < len(ins) {
				fmt.Printf(" %d", ins[i+1])
			}
			i += 2
		default:
			i += 1
		}
		fmt.Println()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.py>\n", os.Args[0])
		os.Exit(1)
	}

	codeBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	code := string(codeBytes)
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:", p.Errors())
		return
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Println("Compile error:", err)
		return
	}

	bytecode := comp.Bytecode()

	fmt.Println("=== Constants ===")
	for i, c := range bytecode.Constants {
		switch v := c.(type) {
		case *compiler.CompiledFunction:
			fmt.Printf("[%d] CompiledFunction (params=%d, locals=%d, kwOnly=%d, posOnly=%d, defaults=%d, posDefaults=%d, names=%v, positionalOnly=%v)\n",
				i, v.NumParameters, v.NumLocals, v.NumKeywordOnly, v.NumPositionalOnly, v.NumDefaults, v.NumPositionalDefaults, v.ParameterNames, v.PositionalOnly)
			fmt.Println("  Instructions:")
			disassemble(v.Instructions, bytecode.Constants)
		case *objects.None:
			fmt.Printf("[%d] None\n", i)
		case *objects.Integer:
			fmt.Printf("[%d] Integer(%d)\n", i, v.Value)
		case *objects.Builtin:
			fmt.Printf("[%d] Builtin\n", i)
		default:
			fmt.Printf("[%d] %T\n", i, c)
		}
	}

	fmt.Println("\n=== Main Instructions ===")
	disassemble(bytecode.Instructions, bytecode.Constants)
}
