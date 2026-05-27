
package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	code := `count = 0
str1 = "test"
str2 = "test"
i = 0
while i < 10:
    if str1 == str2:
        count = count + 1
    i = i + 1
print(count)`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, e := range p.Errors() {
			fmt.Printf("- %s\n", e)
		}
		return
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}

	fmt.Println("Running program:")
	machine := vm.New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		fmt.Printf("Execution error: %v\n", err)
	}
}
