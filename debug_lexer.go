package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: lexer <file>")
		return
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	l := lexer.New(string(data))
	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %-10s Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}