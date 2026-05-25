package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	input := `f = lambda x: x + 1
print(f(5))`
	
	l := lexer.New(input)
	
	fmt.Println("=== Token Sequence ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Token: Type=%s, Literal=%q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
