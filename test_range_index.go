
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	test := `r = range(1, 11)
print(r[0])
print(r[5])`
	fmt.Println("=== Testing range indexing")
	
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
	
	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		fmt.Printf("Run error: %v\n", err)
	}
}
