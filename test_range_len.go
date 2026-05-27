
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
def test_range():
    r = range(1, 11)
    print("type:", type(r))
    print("r:", r)
    l = len(r)
    print("len:", l)
    return l
print("result:", test_range())
`
	fmt.Println("=== Testing ===")
	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Printf("parser errors: %v\n", p.Errors())
		return
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("compile error: %v\n", err)
		return
	}

	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		fmt.Printf("run err: %v\n", err)
	}
}
