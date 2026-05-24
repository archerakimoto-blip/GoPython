package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	code := `try:
    x = 1 / 0
except Exception as e:
    print("Caught")
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

	for i := 0; i < len(instructions); {
		op := instructions[i]
		def, ok := compiler.Definitions[compiler.Opcode(op)]
		if ok {
			fmt.Printf("%04d  %s", i, def.Name)
			for j := 0; j < len(def.OperandWidths); j++ {
				offset := 1
				for k := 0; k < j; k++ {
					offset += def.OperandWidths[k]
				}
				if offset+j < len(def.OperandWidths) {
					val := int(uint16(instructions[i+offset])<<8 | uint16(instructions[i+offset+1]))
					fmt.Printf(" %d", val)
				}
			}
			fmt.Println()
			// 计算下一条指令的位置
			step := 1
			for _, w := range def.OperandWidths {
				step += w
			}
			i += step
		} else {
			fmt.Printf("%04d  UNKNOWN: 0x%02x\n", i, op)
			i++
		}
	}
}
