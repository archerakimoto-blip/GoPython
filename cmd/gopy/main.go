package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
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
	debugFlag   = flag.Bool("debug", false, "Enable debugger")
	profileFlag = flag.Bool("profile", false, "Enable performance profiler")
	breakpoints = flag.String("break", "", "Comma-separated list of breakpoints (IPs)")
	jitFlag    = flag.Bool("jit", false, "Enable JIT compilation")
)

func main() {
	flag.Parse()

	if len(os.Args) > 1 {
		filename := flag.Arg(0)
		if *jitFlag {
			runFileWithJIT(filename)
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

	machine := vm.New(code)
	err = machine.Run()
	if err != nil {
		fmt.Printf("Execution error: %s\n", err)
		return
	}

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
	jitConfig := &jit.JITConfig{
		EnableMachineCode: true,
		OptimizationLevel: 3,
		HotThreshold:      2,
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

func printParserErrors(errors []string) {
	fmt.Println("Parser errors:")
	for _, msg := range errors {
		fmt.Println("\t" + msg)
	}
}
