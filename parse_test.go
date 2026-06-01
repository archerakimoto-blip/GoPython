package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	input := `def __getattr__(self, name):
    if name == 'x':
        return self._x
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	fmt.Println(program.String())
}
