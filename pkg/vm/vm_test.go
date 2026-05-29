package vm

import (
	"math/big"
	"testing"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
)

func TestSimpleArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "addition",
			input:    "_result = 1 + 2",
			expected: 3,
		},
		{
			name:     "multiplication",
			input:    "_result = 3 * 4",
			expected: 12,
		},
		{
			name:     "subtraction",
			input:    "_result = 10 - 5",
			expected: 5,
		},
		{
			name:     "division",
			input:    "_result = 20 / 4",
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			program = desugar.Desugar(program)

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

			// Find _result in globals
			var result objects.Object
			for _, obj := range machine.globals {
				if obj != nil {
					result = obj
					break
				}
			}
			
			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.Type() != objects.INTEGER_OBJ {
				t.Fatalf("Expected integer, got %s", result.Type())
			}

			got := result.(*objects.Integer).Value
			if got.Cmp(big.NewInt(tt.expected)) != 0 {
				t.Errorf("Expected %d, got %s", tt.expected, got.String())
			}
		})
	}
}

func TestLambdaFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "simple lambda",
			input:    "(lambda x: x + 1)(5)",
			expected: 6,
		},
		{
			name:     "identity function",
			input:    "(lambda x: x)(42)",
			expected: 42,
		},
		{
			name:     "nested closure",
			input:    "((lambda x: lambda y: x + y)(5))(10)",
			expected: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			program = desugar.Desugar(program)

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

			result := machine.LastPoppedStackElem()
			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.Type() != objects.INTEGER_OBJ {
				t.Fatalf("Expected integer, got %s", result.Type())
			}

			got := result.(*objects.Integer).Value
			if got.Cmp(big.NewInt(tt.expected)) != 0 {
				t.Errorf("Expected %d, got %s", tt.expected, got.String())
			}
		})
	}
}

func TestBooleanOperations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "true and true",
			input:    "_result = true and true",
			expected: true,
		},
		{
			name:     "true and false",
			input:    "_result = true and false",
			expected: false,
		},
		{
			name:     "true or false",
			input:    "_result = true or false",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			program = desugar.Desugar(program)

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

			var result objects.Object
			for _, obj := range machine.globals {
				if obj != nil {
					result = obj
					break
				}
			}
			
			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.Type() != objects.BOOLEAN_OBJ {
				t.Fatalf("Expected boolean, got %s", result.Type())
			}

			got := result.(*objects.Boolean).Value
			if got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestTypeAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "function with type annotation on parameter",
			input:    "def add(x: int):\n    return x + 1\nadd2 = add(5)",
			expected: 6,
		},
		{
			name:     "function with type annotation on multiple parameters",
			input:    "def add(x: int, y: int):\n    return x + y\nadd2 = add(3, 4)",
			expected: 7,
		},
		{
			name:     "function with return type annotation",
			input:    "def double(x: int) -> int:\n    return x * 2\nadd2 = double(5)",
			expected: 10,
		},
		{
			name:     "function with both parameter and return type annotation",
			input:    "def add(x: int, y: int) -> int:\n    return x + y\nadd2 = add(3, 4)",
			expected: 7,
		},
		{
			name:     "function without type annotation still works",
			input:    "def add(x, y):\n    return x + y\nadd2 = add(3, 4)",
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			program = desugar.Desugar(program)

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

			var result objects.Object
			for _, obj := range machine.globals {
				if obj != nil {
					result = obj
				}
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.Type() != objects.INTEGER_OBJ {
				t.Fatalf("Expected integer, got %s", result.Type())
			}

			got := result.(*objects.Integer).Value
			if got.Cmp(big.NewInt(tt.expected)) != 0 {
				t.Errorf("Expected %d, got %s", tt.expected, got.String())
			}
		})
	}
}
