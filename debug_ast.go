package main

import (
	"fmt"
	"os"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	data, err := os.ReadFile("test_lambda_call.py")
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}
	
	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()
	
	fmt.Println("=== Number of statements:", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d type: %T\n", i, stmt)
		fmt.Printf("Statement %d string: %s\n", i, stmt.String())
	}
	
	if len(p.Errors()) > 0 {
		fmt.Println("\n=== Parser Errors ===")
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
	}
}
