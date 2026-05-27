
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	test := `print(12345)`

	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("Compile error: %v\n", err)
		return
	}

	fmt.Printf("Constants len: %d\n", len(c.Bytecode().Constants))
	for i, cst := range c.Bytecode().Constants {
		fmt.Printf("  %d: %T=%v\n", i, cst, cst)
	}

	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		fmt.Println(err)
	}
}
