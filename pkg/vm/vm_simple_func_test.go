package vm

import (
	"fmt"
	"testing"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
)

func TestSimpleFunction(t *testing.T) {
	input := `def add(x):
    return x + 1
_result = add(5)`

	fmt.Println("Input:", input)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	fmt.Printf("AST statements count: %d\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("  [%d] %T: %s\n", i, stmt, stmt)
	}

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("Compilation error: %v", err)
	}

	bc := comp.Bytecode()

	machine := New(bc)
	err = machine.Run()
	if err != nil {
		t.Fatalf("Execution error: %v", err)
	}

	fmt.Println("\nGlobals:")
	for i, obj := range machine.globals {
		if obj != nil {
			fmt.Printf("  [%d]: %s (%v)\n", i, obj.Type(), obj.Inspect())
		}
	}

	var result objects.Object
	for _, obj := range machine.globals {
		if obj != nil {
			result = obj
		}
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	fmt.Println("\nResult Type:", result.Type())
	fmt.Println("Result Value:", result.Inspect())

	if result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected integer, got %s", result.Type())
	}

	got := result.(*objects.Integer).Value
	expected := int64(6)
	if got.Int64() != expected {
		t.Errorf("Expected %d, got %s", expected, got.String())
	}
}
