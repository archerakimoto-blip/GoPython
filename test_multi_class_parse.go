package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `class Animal:
    def speak(self):
        return "Animal sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"

a = Animal()
print(a.speak())`

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
	for i, stmt := range program.Statements {
		fmt.Printf("\nStatement %d: %T\n", i, stmt)
		if clsStmt, ok := stmt.(*ast.ClassStatement); ok {
			fmt.Printf("  Class: %s\n", clsStmt.Name.Value)
			if clsStmt.SuperClass != nil {
				fmt.Printf("  SuperClass: %s\n", clsStmt.SuperClass.Value)
			}
			fmt.Printf("  Methods: %d\n", len(clsStmt.Methods))
			for _, m := range clsStmt.Methods {
				fmt.Printf("    - %s\n", m.Name)
			}
		} else if assignStmt, ok := stmt.(*ast.AssignStatement); ok {
			fmt.Printf("  Assign: %s\n", assignStmt.Name.Value)
		} else if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			if exprStmt.Expression != nil {
				fmt.Printf("  Expression: %T\n", exprStmt.Expression)
			} else {
				fmt.Printf("  Expression: nil\n")
			}
		}
	}
}