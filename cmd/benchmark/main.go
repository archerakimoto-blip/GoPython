package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-py/go-python/benchmarks"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	fmt.Println("=== GoPy Benchmark Suite ===")
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339))

	results := benchmarks.RunAll()

	fmt.Println("\n=== Summary ===")
	for _, result := range results {
		fmt.Println(result)
	}

	err := benchmarks.WriteResults(results, fmt.Sprintf("benchmark_results_%d.txt", time.Now().Unix()))
	if err != nil {
		fmt.Printf("Error writing results: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nResults written to file.")

	testVM()
}

func testVM() {
	fmt.Println("\n=== VM Performance Test ===")
	
	testCode := `
import math

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

result = fibonacci(25)
print("Fibonacci(25) =", result)
`
	
	fmt.Println("Testing VM execution...")
	start := time.Now()
	
	for i := 0; i < 10; i++ {
		l := lexer.New(testCode)
		p := parser.New(l)
		program := p.ParseProgram()
		
		if len(p.Errors()) != 0 {
			fmt.Println("Parser errors:", p.Errors())
			return
		}
		
		program = desugar.Desugar(program)
		
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("Compilation error: %v\n", err)
			return
		}
		
		code := comp.Bytecode()
		machine := vm.New(code)
		err = machine.Run()
		if err != nil {
			fmt.Printf("Execution error: %v\n", err)
			return
		}
	}
	
	elapsed := time.Since(start)
	fmt.Printf("VM execution time for 10 runs: %v (avg: %v)\n", elapsed, elapsed/10)
}
