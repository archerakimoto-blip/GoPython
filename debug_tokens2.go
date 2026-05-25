package main

import (
	"fmt"
	"os"
	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	data, err := os.ReadFile("test_simple_assign.py")
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}
	
	l := lexer.New(string(data))
	
	fmt.Println("=== Token Sequence for test_simple_assign.py ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Token: Type=%s, Literal=%q\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
