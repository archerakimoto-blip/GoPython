
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
	input := `f"Hello {1+2}, World!"`
	fmt.Println("Testing input:", input)
	
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println("parser error:", err)
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
		fmt.Println("compiler error:", err)
		return
	}
	fmt.Println("Constants:")
	for i, obj := range c.Bytecode().Constants {
		fmt.Printf("  %d: %s\n", i, obj.Inspect())
	}
	fmt.Println("Instructions:", c.Bytecode().Instructions)
	
	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		fmt.Println("vm error:", err)
		return
	}
	
	stackTop := machine.LastPoppedStackElem()
	if stackTop != nil {
		fmt.Println("Result:", stackTop.Inspect())
	}
}
