package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
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
	
	// 检查所有字节码值
	fmt.Println("\nAll opcodes:")
	for i := 0; i < len(bytecode.Instructions); {
		op := bytecode.Instructions[i]
		fmt.Printf("%04d: OpCode=%d\n", i, op)
		if op == byte(compiler.OpConstant) || op == byte(compiler.OpCall) || 
		   op == byte(compiler.OpGetGlobal) || op == byte(compiler.OpSetGlobal) ||
		   op == byte(compiler.OpGetAttribute) {
			i += 3
		} else {
			i++
		}
	}
	
	// 检查 OpListAppend
	fmt.Println("\nChecking for OpListAppend:")
	for i := 0; i < len(bytecode.Instructions); i++ {
		if bytecode.Instructions[i] == byte(compiler.OpListAppend) {
			fmt.Printf("Found OpListAppend at position %d\n", i)
		}
	}
}
