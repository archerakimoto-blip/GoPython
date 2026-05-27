
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"os"
)

func main() {
	content, err := os.ReadFile("tests/benchmarks/006_string.py")
	if err != nil {
		fmt.Println(err)
		return
	}
	
	l := lexer.New(string(content))
	p := parser.New(l)
	_ = p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Println("Errors:")
		for _, e := range p.Errors() {
			fmt.Println(e)
		}
	}
}
