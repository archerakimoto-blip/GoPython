
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	test := `print(len(range(1, 11)))`
	fmt.Println("=== Testing len(range(1, 11))")
	
	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Errors: %v\n", p.Errors())
		return
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("Compile error: %v\n", err)
		return
	}
	
	bc := c.Bytecode()
	fmt.Printf("Constants: %v\n", bc.Constants)

	v := vm.New(bc)
	if err := v.Run(); err != nil {
		fmt.Printf("Run error: %v\n", err)
	}
}
