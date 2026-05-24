
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	input := `[x * 2 for x in [1, 2, 3, 4, 5]]`
	fmt.Println("Testing input:", input)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println("Parser error:", err)
		}
		return
	}
	fmt.Println("Parsed program:", program.String())

	if len(program.Statements) > 0 {
		if exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement); ok {
			fmt.Println("Expression type:", fmt.Sprintf("%T", exprStmt.Expression))
		}
	}

	desugared := desugar.Desugar(program)
	fmt.Println("Desugared program:", desugared.String())

	c := compiler.New()
	err := c.Compile(desugared)
	if err != nil {
		fmt.Println("Compiler error:", err)
		return
	}

	fmt.Println("Constants:")
	for i, obj := range c.Bytecode().Constants {
		fmt.Printf("  %d: %s (Type: %T)\n", i, obj.Inspect(), obj)
	}
	instructions := c.Bytecode().Instructions
	fmt.Print("Decoded Instructions: ")
	i := 0
	for i < len(instructions) {
		op := compiler.Opcode(instructions[i])
		switch op {
		case compiler.OpConstant, compiler.OpJump, compiler.OpJumpNotTruthy, compiler.OpGetGlobal, compiler.OpSetGlobal, compiler.OpArray, compiler.OpHash:
			operand := int(instructions[i+1])<<8 | int(instructions[i+2])
			fmt.Printf("%s(%d) ", op, operand)
			i += 3
		case compiler.OpCall, compiler.OpGetLocal, compiler.OpSetLocal:
			operand := int(instructions[i+1])
			fmt.Printf("%s(%d) ", op, operand)
			i += 2
		default:
			fmt.Printf("%s ", op)
			i += 1
		}
	}
	fmt.Println()

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		fmt.Println("VM error:", err)
		return
	}

	stackTop := machine.LastPoppedStackElem()
	if stackTop != nil {
		fmt.Println("Result:", stackTop.Inspect())
	}
}
