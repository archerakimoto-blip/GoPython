package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	code := `lst = []
lst.append(1)
lst.append(2)
print(len(lst))`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}
	
	program = desugar.Desugar(program)
	
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}
	
	bytecode := comp.Bytecode()
	fmt.Println("Constants:", bytecode.Constants)
	fmt.Println("Instructions length:", len(bytecode.Instructions))
	
	// 检查是否包含 OpListAppend (值 66)
	hasListAppend := false
	for i := 0; i < len(bytecode.Instructions); i++ {
		if bytecode.Instructions[i] == 66 {
			hasListAppend = true
			fmt.Printf("Found OpListAppend at position %d\n", i)
		}
	}
	
	if !hasListAppend {
		fmt.Println("OpListAppend not found in bytecode!")
	}
	
	fmt.Println("\nRunning:")
	machine := vm.New(bytecode)
	err := machine.Run()
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
	}
}
