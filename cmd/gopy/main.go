package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

const PROMPT = ">> "

func main() {
	if len(os.Args) > 1 {
		runFile(os.Args[1])
		return
	}
	runREPL()
}

func runREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	globals := make([]objects.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()
	constants := []objects.Object{}

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

		// 进行脱糖转换
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

	// 进行脱糖转换
	program = desugar.Desugar(program)

	fmt.Println("=== Parsed AST ===")
	fmt.Println(program.String())

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}

	code := comp.Bytecode()
	fmt.Println("\n=== Bytecode ===")
	fmt.Printf("Constants (%d):\n", len(code.Constants))
	for i, c := range code.Constants {
		fmt.Printf("  %d: %s\n", i, c.Inspect())
	}
	fmt.Printf("\nInstructions:\n")
	for i := 0; i < len(code.Instructions); {
		op := code.Instructions[i]
		fmt.Printf("%04d: Opcode %d", i, op)
		if op == byte(compiler.OpConstant) || op == byte(compiler.OpGetGlobal) || op == byte(compiler.OpSetGlobal) ||
			op == byte(compiler.OpArray) || op == byte(compiler.OpHash) || op == byte(compiler.OpSet) ||
			op == byte(compiler.OpJump) || op == byte(compiler.OpJumpNotTruthy) {
			if i+2 < len(code.Instructions) {
				operand := int(uint16(code.Instructions[i+2])<<8 | uint16(code.Instructions[i+1]))
				fmt.Printf(" (0x%04x)", operand)
			}
			i += 3
		} else if op == byte(compiler.OpCall) || op == byte(compiler.OpGetLocal) || op == byte(compiler.OpSetLocal) {
			if i+1 < len(code.Instructions) {
				fmt.Printf(" (%d)", int(code.Instructions[i+1]))
			}
			i += 2
		} else {
			i += 1
		}
		fmt.Println()
	}
	fmt.Println()

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

func printParserErrors(errors []string) {
	fmt.Println("Parser errors:")
	for _, msg := range errors {
		fmt.Println("\t" + msg)
	}
}
