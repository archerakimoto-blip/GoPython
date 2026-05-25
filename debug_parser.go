package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `class Animal:{def speak(self): return "Animal"}`
	
	l := lexer.New(input)
	p := parser.New(l)
	
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Println("\t", err)
		}
	} else {
		fmt.Println("Parse successful!")
		fmt.Println(program.String())
	}
}
