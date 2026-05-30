package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run token_test.go <file>")
		return
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	l := lexer.New(string(content))

	fmt.Println("Tokens:")
	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF {
			break
		}
		fmt.Printf("Type: %-10q Literal: %q\n", tok.Type, tok.Literal)
	}
}
