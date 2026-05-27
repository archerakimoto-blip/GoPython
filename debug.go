package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func printNode(node ast.Node, indent string) {
	switch n := node.(type) {
	case *ast.Program:
		fmt.Println(indent + "Program:")
		for _, stmt := range n.Statements {
			printNode(stmt, indent+"  ")
		}
	case *ast.ExpressionStatement:
		fmt.Println(indent + "ExpressionStatement:")
		printNode(n.Expression, indent+"  ")
	case *ast.FunctionLiteral:
		fmt.Println(indent + "FunctionLiteral: " + n.Name)
		fmt.Println(indent+"  Params:")
		for _, p := range n.Parameters {
			fmt.Println(indent+"    " + p.Value)
		}
		fmt.Println(indent+"  Body:")
		printNode(n.Body, indent+"  ")
	case *ast.BlockStatement:
		fmt.Println(indent + "BlockStatement:")
		for _, stmt := range n.Statements {
			printNode(stmt, indent+"  ")
		}
	case *ast.ReturnStatement:
		fmt.Println(indent + "ReturnStatement:")
		printNode(n.ReturnValue, indent+"  ")
	case *ast.IfExpression:
		fmt.Println(indent + "IfExpression:")
		fmt.Println(indent+"  Condition:")
		printNode(n.Condition, indent+"    ")
		fmt.Println(indent+"  Consequence:")
		printNode(n.Consequence, indent+"    ")
		if n.Alternative != nil {
			fmt.Println(indent+"  Alternative:")
			printNode(n.Alternative, indent+"    ")
		}
	case *ast.InfixExpression:
		fmt.Println(indent + "InfixExpression: " + n.Operator)
		printNode(n.Left, indent+"  L: ")
		printNode(n.Right, indent+"  R: ")
	case *ast.Identifier:
		fmt.Println(indent + "Identifier: " + n.Value)
	case *ast.IntegerLiteral:
		fmt.Printf(indent+"IntegerLiteral: %d\n", n.Value)
	case *ast.AssignStatement:
		fmt.Println(indent + "AssignStatement:")
		printNode(n.Name, indent+"  Name: ")
		printNode(n.Value, indent+"  Value: ")
	case *ast.CallExpression:
		fmt.Println(indent + "CallExpression:")
		printNode(n.Function, indent+"  Func: ")
		for _, arg := range n.Arguments {
			printNode(arg, indent+"  Arg: ")
		}
	default:
		fmt.Printf(indent+"Unknown type: %T %+v\n", node, node)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run debug.go <filename>")
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}
	fmt.Println("=== Lexing ===")
	l := lexer.New(string(data))
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %-10v Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}

	fmt.Println("\n=== Parsing ===")
	l = lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, e := range p.Errors() {
			fmt.Println("  " + e)
		}
		return
	}

	printNode(program, "")

	fmt.Println("\n=== Desugaring ===")
	desugared := desugar.Desugar(program)
	printNode(desugared, "")
}
