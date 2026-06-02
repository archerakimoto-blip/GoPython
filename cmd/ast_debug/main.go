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
		fmt.Fprintf(os.Stderr, "Usage: %s <file.py>\n", os.Args[0])
		os.Exit(1)
	}

	codeBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	code := string(codeBytes)
	l := lexer.New(code)
	
	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Type=%q Literal=%q\n", tok.Type, tok.Literal)
		if tok.Type == "EOF" {
			break
		}
	}
	
	l2 := lexer.New(code)
	p := parser.New(l2)
	program := p.ParseProgram()
	
	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:", p.Errors())
		return
	}
	
	fmt.Println("\n=== AST ===")
	for _, stmt := range program.Statements {
		fmt.Printf("Statement type=%T: %s\n", stmt, stmt.String())
		if cs, ok := stmt.(*ast.ClassStatement); ok {
			for _, m := range cs.Methods {
				fmt.Printf("  Method %q decorators=%d\n", m.Name, len(m.Decorators))
				for _, d := range m.Decorators {
					fmt.Printf("    Decorator type=%T: %s\n", d, d.String())
				}
			}
		}
	}
}
