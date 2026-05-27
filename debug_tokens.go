package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run debug_tokens.go <filename>")
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}
	fmt.Println("=== Lexing ===")
	l := lexer.New(string(data))
	for i := 0; i < 1000; i++ {
		tok := l.NextToken()
		fmt.Printf("[%3d] Type: %-20v Literal: %q\n", i, tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
