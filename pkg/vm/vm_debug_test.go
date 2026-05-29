package vm

import (
	"fmt"
	"testing"

	"github.com/go-py/go-python/pkg/lexer"
)

func TestDebugTokens(t *testing.T) {
	input := `def add(x: int): return x + 1; _result = add(5)`

	fmt.Println("Input:", input)
	fmt.Println("\n=== Tokens ===")
	l := lexer.New(input)
	for {
		tok := l.NextToken()
		fmt.Printf("Token: %-20s Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
