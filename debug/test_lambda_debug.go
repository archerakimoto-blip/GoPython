package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := "lambda x: x + 1"
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
			
			if lambda, ok := exprStmt.Expression.(*ast.LambdaExpression); ok {
				fmt.Println("Parameters:", lambda.Parameters)
				fmt.Println("Body:", lambda.Body.String())
			}
		}
	}
}
