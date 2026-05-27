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
	tests := []struct {
		name string
		code string
	}{
		{"fib(20)", `def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)
print(fib(20))`},
		{"factorial(10)", `def factorial(n):
    result = 1
    for i in range(1, n+1):
        result *= i
    return result
print(factorial(10))`},
		{"loop 10000", `sum_result = 0
for i in range(10000):
    sum_result += i
print(sum_result)`},
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
	}
}
