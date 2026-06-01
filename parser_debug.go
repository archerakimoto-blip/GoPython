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
		fmt.Println("Usage: go run parser_debug.go <file>")
		return
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	l := lexer.New(string(content))
	// First print all tokens
	fmt.Println("Tokens:")
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %-10s Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}

	// Recreate lexer and parser
	l = lexer.New(string(content))
	p := parser.New(l)

	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("\nParser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		return
	}

	fmt.Println("\nProgram parsed successfully!")
	fmt.Printf("Number of statements: %d\n", len(program.Statements))
	for _, stmt := range program.Statements {
		fmt.Printf("Stmt type: %T\n", stmt)
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			fmt.Printf("Expression type: %T\n", exprStmt.Expression)
			if fnLit, ok := exprStmt.Expression.(*ast.FunctionLiteral); ok {
				fmt.Printf("  FunctionLiteral: Name: %s\n", fnLit.Name)
				fmt.Printf("  PositionalOnlyCount: %v\n", fnLit.PositionalOnlyCount)
				fmt.Printf("  KeywordOnlyStart: %v\n", fnLit.KeywordOnlyStart)
			}
		}
	}
}
