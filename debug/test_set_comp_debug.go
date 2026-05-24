package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := "{x for x in [1, 2, 3, 4, 5] if x > 2}"
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
	
	fmt.Println("Parsed program:", program.String())
	
	if len(program.Statements) > 0 {
		if exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement); ok {
			fmt.Println("Expression type:", fmt.Sprintf("%T", exprStmt.Expression))
			
			if setComp, ok := exprStmt.Expression.(*ast.SetComprehension); ok {
				fmt.Println("Element:", setComp.Element.String())
				fmt.Println("Variable:", setComp.Variable.Value)
				fmt.Println("Iterable:", setComp.Iterable.String())
				if setComp.Condition != nil {
					fmt.Println("Condition:", setComp.Condition.String())
				} else {
					fmt.Println("Condition: nil")
				}
			}
		}
	}
}
