package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug_parser.go <file.py>")
		return
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	l := lexer.New(string(content))
	p := parser.New(l)

	// Let's manually print all tokens first to see!
	fmt.Println("Tokens:")
	i := 0
	for {
		tok := l.NextToken()
		fmt.Printf("  [%d] Type: %v, Literal: %q\n", i, tok.Type, tok.Literal)
		i++
		if tok.Type == lexer.EOF {
			break
		}
	}

	// Now parse and check
	fmt.Println("\n--- Parsing ---")
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		return
	}

	fmt.Println("Program parsed successfully")
	for _, stmt := range program.Statements {
		fmt.Printf("Stmt: %#v\n", stmt)
		if es, ok := stmt.(*ast.ExpressionStatement); ok {
			fmt.Printf("Expr: %#v\n", es.Expression)
		}
	}
}
