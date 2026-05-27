
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	test := `
r = range(1, 11)
l = len(r)
print("l is:", l)
`
	fmt.Println("=== Testing ===")
	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) >0 {
		fmt.Println(p.Errors())
		return
	}
	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Constants:")
	for i, constVal := range c.Bytecode().Constants {
		fmt.Printf("  %v: %T=%v\n", i, constVal, constVal)
	}

	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		fmt.Println(err)
	}
}
