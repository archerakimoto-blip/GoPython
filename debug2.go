package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run debug2.go <filename>")
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

	fmt.Println("\n=== Parsing with debug ===")
	l = lexer.New(string(data))
	p := parser.New(l)

	// Now let's manually step through parseBlockStatement
	fmt.Println("\n--- Stepping through tokens ---")

	count := 0
	for {
		cur := p.CurToken() // Wait does parser expose these? Oh wait we need to check the struct of Parser!
		peek := p.PeekToken()
		fmt.Printf("Step %d: cur=%-10v (%q), peek=%-10v (%q)\n", count, cur.Type, cur.Literal, peek.Type, peek.Literal)
		if cur.Type == lexer.EOF {
			break
		}
		p.NextToken()
		count++
		if count > 100 {
			break
		}
	}
}
