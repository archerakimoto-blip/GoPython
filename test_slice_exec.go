
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
	input := `[1,2,3,4,5][1:3]`
	fmt.Println("Testing input:", input)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:", p.Errors())
		return
	}
	fmt.Println("Parsed program:", program.String())
	if len(program.Statements) > 0 {
		if stmt, ok := program.Statements[0].(*ast.ExpressionStatement); ok {
			fmt.Printf("Expression type: %T\n", stmt.Expression)
		}
	}

	program = desugar.Desugar(program)
	fmt.Println("Desugared program:", program.String())
	if len(program.Statements) > 0 {
		if stmt, ok := program.Statements[0].(*ast.ExpressionStatement); ok {
			fmt.Printf("Desugared expression type: %T\n", stmt.Expression)
		}
	}

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		fmt.Println("Compiler error:", err)
		return
	}

	fmt.Println("Constants:")
	for i, obj := range c.Bytecode().Constants {
		fmt.Printf("  %d: %s\n", i, obj.Inspect())
	}

	fmt.Print("Instructions:")
	for _, b := range c.Bytecode().Instructions {
		fmt.Printf(" %d", b)
	}
	fmt.Println()

	opNames := map[compiler.Opcode]string{
		compiler.OpConstant:    "OpConstant",
		compiler.OpPop:         "OpPop",
		compiler.OpAdd:         "OpAdd",
		compiler.OpSub:         "OpSub",
		compiler.OpMul:         "OpMul",
		compiler.OpDiv:         "OpDiv",
		compiler.OpTrue:        "OpTrue",
		compiler.OpFalse:       "OpFalse",
		compiler.OpEqual:       "OpEqual",
		compiler.OpNotEqual:    "OpNotEqual",
		compiler.OpGreaterThan: "OpGreaterThan",
		compiler.OpLessThan:    "OpLessThan",
		compiler.OpMinus:       "OpMinus",
		compiler.OpBang:        "OpBang",
		compiler.OpJump:        "OpJump",
		compiler.OpJumpNotTruthy: "OpJumpNotTruthy",
		compiler.OpNull:        "OpNull",
		compiler.OpGetGlobal:   "OpGetGlobal",
		compiler.OpSetGlobal:   "OpSetGlobal",
		compiler.OpArray:       "OpArray",
		compiler.OpHash:        "OpHash",
		compiler.OpIndex:       "OpIndex",
		compiler.OpSlice:       "OpSlice",
		compiler.OpCall:        "OpCall",
		compiler.OpReturnValue: "OpReturnValue",
		compiler.OpReturn:      "OpReturn",
		compiler.OpGetLocal:    "OpGetLocal",
		compiler.OpSetLocal:    "OpSetLocal",
	}
	fmt.Print("Decoded instructions:")
	i := 0
	instructions := c.Bytecode().Instructions
	for i < len(instructions) {
		op := compiler.Opcode(instructions[i])
		name, ok := opNames[op]
		if !ok {
			name = fmt.Sprintf("OpUnknown(%d)", op)
		}
		fmt.Printf(" %s", name)
		switch op {
		case compiler.OpConstant, compiler.OpJump, compiler.OpJumpNotTruthy,
			compiler.OpGetGlobal, compiler.OpSetGlobal, compiler.OpArray, compiler.OpHash:
			i += 3
		case compiler.OpCall, compiler.OpGetLocal, compiler.OpSetLocal:
			i += 2
		default:
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

	result := machine.LastPoppedStackElem()
	fmt.Println("Result:", result.Inspect())
}
