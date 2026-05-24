
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	input := `[x * 2 for x in [1,2,3]]`
	l := lexer.New(input)

	for {
		tok := l.NextToken()
		fmt.Printf("Type: %-10s Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
