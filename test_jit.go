package main

import (
	"fmt"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/jit"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	// Test case: fib(20)
	code := `def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)
print(fib(20))`

	fmt.Println("=== Testing JIT with fib(20) ===")
	
	// Lexer
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}
	
	// Desugar
	program = desugar.Desugar(program)
	
	// Compile
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}
	
	// Test without JIT
	fmt.Println("\n--- Running without JIT ---")
	start := time.Now()
	machine := vm.New(comp.Bytecode())
	err := machine.Run()
	elapsed := time.Now().Sub(start)
	
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
	}
	
	fmt.Printf("Time: %v\n", elapsed)
	fmt.Printf("Instructions: %d\n", machine.InstructionCount())
	
	// Test with JIT
	fmt.Println("\n--- Running with JIT ---")
	jitConfig := &jit.JITConfig{
		EnableMachineCode:  true,
		OptimizationLevel: 3,
		HotThreshold:      1,
		Platform:          jit.PlatformX86_64,
	}
	
	start = time.Now()
	jitVM := vm.NewVMWithJIT(comp.Bytecode(), jitConfig)
	err = jitVM.RunWithJIT()
	elapsed = time.Now().Sub(start)
	
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
	}
	
	fmt.Printf("Time: %v\n", elapsed)
	fmt.Printf("JIT compiled functions: %d\n", jitVM.GetJITCompiledFunctionCount())
	jitVM.PrintJITStats()
}
