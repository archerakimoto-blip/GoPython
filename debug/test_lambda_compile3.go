package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := "f = lambda x: x + 1"
	fmt.Println("Testing input:", input)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("  ", err)
		}
		return
	}
	
	desugared := desugar.Desugar(program)
	fmt.Println("Desugared program:", desugared.String())
	
	c := compiler.New()
	err := c.Compile(desugared)
	if err != nil {
		fmt.Println("Compiler error:", err)
		return
	}
	
	fmt.Println("\nMain instructions:")
	for i := 0; i < len(c.Bytecode().Instructions); {
		op := compiler.Opcode(c.Bytecode().Instructions[i])
		fmt.Printf("[%d] %v", i, op)
		switch op {
		case compiler.OpConstant:
			operand := int(c.Bytecode().Instructions[i+1])<<8 | int(c.Bytecode().Instructions[i+2])
			fmt.Printf("(%d)", operand)
			i += 3
		case compiler.OpGetLocal, compiler.OpSetLocal:
			operand := int(c.Bytecode().Instructions[i+1])
			fmt.Printf("(%d)", operand)
			i += 2
		case compiler.OpGetGlobal, compiler.OpSetGlobal:
			operand := int(c.Bytecode().Instructions[i+1])<<8 | int(c.Bytecode().Instructions[i+2])
			fmt.Printf("(%d)", operand)
			i += 3
		case compiler.OpReturn:
			i++
		default:
			i++
		}
		fmt.Println()
	}
	
	fmt.Println("\nConstants:")
	for i, obj := range c.Bytecode().Constants {
		fmt.Printf("  %d: %s (Type: %T)\n", i, obj.Inspect(), obj)
	}
	
	if cf, ok := c.Bytecode().Constants[5].(*compiler.CompiledFunction); ok {
		fmt.Println("\nCompiled function details:")
		fmt.Printf("  NumParameters: %d\n", cf.NumParameters)
		fmt.Printf("  NumLocals: %d\n", cf.NumLocals)
		fmt.Println("\nCompiled function instructions:")
		for i := 0; i < len(cf.Instructions); {
			op := compiler.Opcode(cf.Instructions[i])
			fmt.Printf("[%d] %v", i, op)
			switch op {
			case compiler.OpConstant:
				operand := int(cf.Instructions[i+1])<<8 | int(cf.Instructions[i+2])
				fmt.Printf("(%d)", operand)
				i += 3
			case compiler.OpGetLocal, compiler.OpSetLocal:
				operand := int(cf.Instructions[i+1])
				fmt.Printf("(%d)", operand)
				i += 2
			case compiler.OpGetGlobal, compiler.OpSetGlobal:
				operand := int(cf.Instructions[i+1])<<8 | int(cf.Instructions[i+2])
				fmt.Printf("(%d)", operand)
				i += 3
			case compiler.OpReturn:
				i++
			default:
				i++
			}
			fmt.Println()
		}
	}
}
