package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func printAST(node ast.Node, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	switch n := node.(type) {
	case *ast.Program:
		fmt.Printf("%sProgram:\n", prefix)
		for _, stmt := range n.Statements {
			printAST(stmt, indent+1)
		}
	case *ast.ClassStatement:
		fmt.Printf("%sClassStatement: %s\n", prefix, n.Name.Value)
		if n.SuperClass != nil {
			fmt.Printf("%s  SuperClass: %s\n", prefix, n.SuperClass.Value)
		}
		fmt.Printf("%s  Body:\n", prefix)
		for _, stmt := range n.Body.Statements {
			printAST(stmt, indent+2)
		}
	case *ast.ExpressionStatement:
		fmt.Printf("%sExpressionStatement:\n", prefix)
		printAST(n.Expression, indent+1)
	case *ast.FunctionLiteral:
		fmt.Printf("%sFunctionLiteral: %s\n", prefix, n.Name)
		fmt.Printf("%s  Parameters: ", prefix)
		for i, param := range n.Parameters {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", param.Value)
		}
		fmt.Println()
		fmt.Printf("%s  Body:\n", prefix)
		for _, stmt := range n.Body.Statements {
			printAST(stmt, indent+2)
		}
	case *ast.ReturnStatement:
		fmt.Printf("%sReturnStatement:\n", prefix)
		printAST(n.ReturnValue, indent+1)
	case *ast.StringLiteral:
		fmt.Printf("%sStringLiteral: %s\n", prefix, n.Value)
	case *ast.AssignStatement:
		fmt.Printf("%sAssignStatement: %s\n", prefix, n.Name.Value)
		printAST(n.Value, indent+1)
	case *ast.Identifier:
		fmt.Printf("%sIdentifier: %s\n", prefix, n.Value)
	case *ast.IntegerLiteral:
		fmt.Printf("%sIntegerLiteral: %d\n", prefix, n.Value)
	default:
		fmt.Printf("%sUnknown: %T\n", prefix, node)
	}
}

func main() {
	input := `class Animal:
    def speak(self):
        return "Animal sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"

a = Animal()
d = Dog()
print(a.speak())
print(d.speak())`

	fmt.Println("Input code:")
	fmt.Println(input)

	fmt.Println("\n--- Parser AST ---")
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
		return
	}

	fmt.Printf("\nProgram has %d statements\n", len(program.Statements))
	printAST(program, 0)
}