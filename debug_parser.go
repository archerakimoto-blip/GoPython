package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `f = lambda x: x + 1
print(f(5))`
	
	l := lexer.New(input)
	p := parser.New(l)
	
	fmt.Println("=== Before Parsing ===")
	fmt.Printf("curToken: Type=%s, Literal=%q\n", p.CurToken().Type, p.CurToken().Literal)
	fmt.Printf("peekToken: Type=%s, Literal=%q\n", p.PeekToken().Type, p.PeekToken().Literal)
	
	program := p.ParseProgram()
	
	fmt.Println("\n=== After Parsing ===")
	fmt.Printf("Number of statements: %d\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
		fmt.Printf("  String: %s\n", stmt.String())
	}
	
	if len(p.Errors()) > 0 {
		fmt.Println("\n=== Parser Errors ===")
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
	}
}
