
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"os"
)

func main() {
	content, err := os.ReadFile("tests/benchmarks/006_string.py")
	if err != nil {
		fmt.Println(err)
		return
	}
	
	l := lexer.New(string(content))
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %15q Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
