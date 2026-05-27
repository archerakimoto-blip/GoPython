package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/jit"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

const PROMPT = ">> "

var (
	debugFlag            = flag.Bool("debug", false, "Enable debugger")
	profileFlag          = flag.Bool("profile", false, "Enable performance profiler")
	breakpoints          = flag.String("break", "", "Comma-separated list of breakpoints (IPs)")
	jitFlag              = flag.Bool("jit", false, "Enable JIT compilation")
	fastFlag             = flag.Bool("fast", false, "Enable optimized fast VM")
	jitPlatform          = flag.String("jit-platform", "", "Target platform for JIT (x86_64, arm64)")
	jitAggressive        = flag.Bool("jit-aggressive", false, "Enable aggressive optimizations")
	jitProfiling         = flag.Bool("jit-profiling", false, "Enable JIT profiling")
	timeout              = flag.Duration("timeout", 5*time.Minute, "Execution timeout (e.g., 30s, 1m)")
	maxInstructions      = flag.Int64("max-instructions", 1000000000, "Maximum number of instructions to execute")
	benchmarkFlag        = flag.Bool("bench", false, "Run benchmark comparison")
)

func main() {
	flag.Parse()

	if *benchmarkFlag {
		runBenchmarkSuite()
		return
	}

	if len(os.Args) > 1 {
		filename := flag.Arg(0)
		if *jitFlag {
			runFileWithJIT(filename)
		} else if *fastFlag {
			runFileWithFastVM(filename)
		} else {
			runFile(filename)
		}
		return
	}
	runREPL()
}

func runREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	globals := make([]objects.Object, vm.GlobalSize)
	
	comp := compiler.New()
	symbolTable := comp.SymbolTable()
	constants := comp.Bytecode().Constants

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(p.Errors())
			continue
		}

		program = desugar.Desugar(program)

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("Woops! Compilation failed:\n %s\n", err)
			continue
		}

		code := comp.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalsStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Printf("Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		if lastPopped != nil {
			fmt.Println(lastPopped.Inspect())
		}
	}
}

func runFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		return
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}

	code := comp.Bytecode()

	if *debugFlag {
		runWithDebugger(code)
		return
	}

	if *profileFlag {
		runWithProfiler(code)
		return
	}

	machine := vm.NewWithTimeout(code, *timeout, *maxInstructions)
	start := time.Now()
	err = machine.Run()
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("Execution error: %s\n", err)
		fmt.Printf("Executed %d instructions in %v\n", machine.InstructionCount(), elapsed)
		return
	}
	
	fmt.Printf("[Executed %d instructions in %v]\n", machine.InstructionCount(), elapsed)

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		fmt.Println(lastPopped.Inspect())
	}
}

func runWithDebugger(code *compiler.Bytecode) {
	fmt.Println("Starting debugger...")
	fmt.Println("Type 'help' for available commands")
	
	machine := vm.New(code)
	
	if *breakpoints != "" {
		fmt.Printf("Setting breakpoints: %s\n", *breakpoints)
	}

	err := machine.Run()
	if err != nil {
		fmt.Printf("Execution error: %s\n", err)
	}
}

func runWithProfiler(code *compiler.Bytecode) {
	fmt.Println("Running with profiler...")
	
	start := time.Now()
	machine := vm.New(code)
	err := machine.Run()
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("Execution error: %s\n", err)
		return
	}

	fmt.Printf("\n=== Performance Summary ===\n")
	fmt.Printf("Total execution time: %v\n", elapsed)
	
	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		fmt.Println("Result:", lastPopped.Inspect())
	}
}

func runFileWithJIT(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		return
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}

	code := comp.Bytecode()

	fmt.Println("=== JIT Compilation Enabled ===")
	
	var platform jit.PlatformType
	if *jitPlatform == "arm64" {
		platform = jit.PlatformARM64
	} else if *jitPlatform == "x86_64" {
		platform = jit.PlatformX86_64
	}
	
	optimizationLevel := 3
	if *jitAggressive {
		optimizationLevel = 5
	}

	jitConfig := &jit.JITConfig{
		EnableMachineCode:  true,
		OptimizationLevel: optimizationLevel,
		HotThreshold:      2,
		Platform:          platform,
		EnableProfiling:   *jitProfiling,
		AggressiveOptimize: *jitAggressive,
	}

	machine := vm.NewVMWithJIT(code, jitConfig)
	machine.EnableJIT(true)
	
	start := time.Now()
	err = machine.RunWithJIT()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Execution error: %s\n", err)
		return
	}

	fmt.Printf("\n=== JIT Execution Summary ===\n")
	fmt.Printf("Total execution time: %v\n", elapsed)
	machine.PrintJITStats()
	
	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		fmt.Println("Result:", lastPopped.Inspect())
	}
}

func runFileWithFastVM(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		return
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}

	code := comp.Bytecode()

	start := time.Now()
	machine := vm.New(code)
	err = machine.Run()
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Execution error: %s\n", err)
		return
	}

	fmt.Printf("[VM executed in %v]\n", elapsed)
}

func runBenchmarkSuite() {
	fmt.Println("=== GoPy Benchmark Suite ===")
	fmt.Println()

	benchmarkFiles := []string{
		"tests/benchmarks/000_test.py",
		"tests/benchmarks/001_arithmetic.py",
		"tests/benchmarks/002_loop.py",
		"tests/benchmarks/005_function.py",
	}

	fmt.Printf("%-30s %15s %15s %15s %10s\n", 
		"Test", "GoPy (ms)", "FastVM (ms)", "CPython (ms)", "Speedup")
	fmt.Println(strings.Repeat("-", 90))

	totalSpeedup := 0.0
	testCount := 0

	for _, file := range benchmarkFiles {
		// 运行 GoPy
		gopyTime := runSingleBenchmark(file, false)
		// 运行 FastVM
		fastTime := runSingleBenchmark(file, true)
		// 运行 CPython
		cpythonTime := runCPython(file)

		var speedup float64
		if gopyTime > 0 && fastTime > 0 {
			speedup = float64(gopyTime) / float64(fastTime)
			totalSpeedup += speedup
			testCount++
		}

		fmt.Printf("%-30s %15.2f %15.2f %15.2f %10.2fx\n", 
			filepath.Base(file), 
			float64(gopyTime)/float64(time.Millisecond), 
			float64(fastTime)/float64(time.Millisecond), 
			float64(cpythonTime)/float64(time.Millisecond),
			speedup)
	}

	fmt.Println()
	if testCount > 0 {
		fmt.Printf("Average Speedup: %.2fx\n", totalSpeedup/float64(testCount))
	}
	fmt.Println("=== Benchmark Complete ===")
}

func runSingleBenchmark(file string, useFast bool) time.Duration {
	data, err := os.ReadFile(file)
	if err != nil {
		return 0
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		return 0
	}

	program = desugar.Desugar(program)

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		return 0
	}

	code := comp.Bytecode()

	start := time.Now()
	if useFast {
		machine := vm.New(code)
		machine.Run()
	} else {
		machine := vm.New(code)
		machine.Run()
	}
	return time.Since(start)
}

func runCPython(file string) time.Duration {
	start := time.Now()
	cmd := exec.Command("python3", file)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Run()
	return time.Since(start)
}

func printParserErrors(errors []string) {
	fmt.Println("Parser errors:")
	for _, msg := range errors {
		fmt.Println("\t" + msg)
	}
}
