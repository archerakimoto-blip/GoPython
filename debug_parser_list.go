
package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `[1,2,3,4,5]`
	l := lexer.New(input)
	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		fmt.Printf("%+v\n", tok)
		if tok.Type == lexer.EOF {
			break
		}
	}

	l2 := lexer.New(input)
	p := parser.New(l2)
	prog := p.ParseProgram()
	fmt.Println("\n=== Program ===")
	fmt.Println(prog.String())

	fmt.Println("\n=== Parser Errors ===")
	for _, err := range p.Errors() {
		fmt.Println("-", err)
	}
}
