
package main

import (
	"fmt"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	test := `def f(n):
    return n
print(f(10))`

	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) !=0 {
		fmt.Printf("Errors: %v\n", p.Errors())
		return
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("Compile err: %v\n", err)
		return
	}

	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		fmt.Println(err)
	}
}
