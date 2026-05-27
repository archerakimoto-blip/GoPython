package main

import (
	"fmt"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	// 测试不同的代码片段
	tests := []struct {
		name string
		code string
	}{
		{"Simple loop", "sum = 0\nfor i in range(1000):\n    sum += i\nprint(sum)"},
		{"Fibonacci", "def fib(n):\n    if n <= 1:\n        return n\n    return fib(n-1) + fib(n-2)\nprint(fib(20))"},
		{"List append", "lst = []\nfor i in range(1000):\n    lst.append(i)\nprint(len(lst))"},
		{"Dict operations", "d = {}\nfor i in range(1000):\n    d[i] = i * 2\nprint(len(d))"},
	}

	for _, test := range tests {
		fmt.Printf("\n=== Testing: %s ===\n", test.name)
		
		l := lexer.New(test.code)
		p := parser.New(l)
		program := p.ParseProgram()
		
		if len(p.Errors()) > 0 {
			fmt.Printf("Parser errors: %v\n", p.Errors())
			continue
		}
		
		program = desugar.Desugar(program)
		
		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			fmt.Printf("Compilation error: %v\n", err)
			continue
		}
		
		start := time.Now()
		machine := vm.New(comp.Bytecode())
		err := machine.Run()
		elapsed := time.Since(start)
		
		if err != nil {
			fmt.Printf("Execution error: %v\n", err)
		}
		
		fmt.Printf("Time: %v\n", elapsed)
		fmt.Printf("Instructions: %d\n", machine.InstructionCount())
		fmt.Printf("Speed: %.2f M instructions/sec\n", 
			float64(machine.InstructionCount())/elapsed.Seconds()/1000000)
	}
}
