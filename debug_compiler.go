
package main

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	code := `print("=== 字符串操作基准测试 ===\n\n1. String concatenation (1000 iterations):\n\nresult = \"\"\ni = 0\nwhile i < 1000:\n    result = result + str(i)\n    i = i + 1\nprint(len(result))\n\nprint()\nprint("2. String comparison (1000 iterations):\n\ncount = 0\nstr1 = \"test\"\nstr2 = \"test\"\ni = 0\nwhile i < 1000:\n    if str1 == str2:\n        count = count + 1\n    i = i + 1\nprint(count)`
	`

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

	bytecode := comp.Bytecode()
	fmt.Println("Instructions:")
	for i, ins := range bytecode.Instructions {
		fmt.Printf("%04d: %s\n", i, compiler.InstructionToString(ins))
	}

	fmt.Println("\nConstants:")
	for i, c := range bytecode.Constants {
		fmt.Printf("%04d: %v\n", i, c)
	}
}
