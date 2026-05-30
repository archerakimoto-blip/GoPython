package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run parser_debug.go <file>")
		return
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	l := lexer.New(string(content))
	p := parser.New(l)

	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		return
	}

	fmt.Println("Program parsed successfully!")
	fmt.Printf("Number of statements: %d\n", len(program.Statements))
}
