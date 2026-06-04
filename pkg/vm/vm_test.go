package vm

import (
	"strings"
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
			if got != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, got)
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
			if got != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, got)
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

func runTestCode(t *testing.T, input string) objects.Object {
	t.Helper()
	l := lexer.New(input)
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

	for _, obj := range machine.globals {
		if obj != nil {
			return obj
		}
	}
	return nil
}

func runTestCodeGetGlobal(t *testing.T, input string, name string) objects.Object {
	t.Helper()
	l := lexer.New(input)
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

	// Find the symbol index for the given name
	symbol, ok := comp.SymbolTable().Resolve(name)
	if !ok {
		return nil
	}
	if symbol.Scope != compiler.GlobalScope {
		return nil
	}
	return machine.globals[symbol.Index]
}

func TestExceptionGroup(t *testing.T) {
	t.Run("create exception group", func(t *testing.T) {
		input := `
_result = ExceptionGroup("eg", [TypeError("a"), ValueError("b")])
`
		result := runTestCode(t, input)
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.EXCEPTION_GROUP_OBJ {
			t.Fatalf("Expected EXCEPTION_GROUP, got %s", result.Type())
		}
		eg := result.(*objects.ExceptionGroup)
		if eg.Message != "eg" {
			t.Errorf("Expected message 'eg', got %q", eg.Message)
		}
		if len(eg.Exceptions) != 2 {
			t.Fatalf("Expected 2 exceptions, got %d", len(eg.Exceptions))
		}
	})

	t.Run("except star catches matching exceptions", func(t *testing.T) {
		input := `
_caught = None
try:
    try:
        raise ExceptionGroup("eg", [TypeError("err1"), ValueError("err2")])
    except* TypeError as e:
        _caught = e
except:
    pass
`
		result := runTestCodeGetGlobal(t, input, "_caught")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.EXCEPTION_GROUP_OBJ {
			t.Fatalf("Expected EXCEPTION_GROUP, got %s", result.Type())
		}
		eg := result.(*objects.ExceptionGroup)
		if len(eg.Exceptions) != 1 {
			t.Fatalf("Expected 1 exception in caught group, got %d", len(eg.Exceptions))
		}
		err := eg.Exceptions[0].(*objects.Error)
		if err.ErrorType != "TypeError" {
			t.Errorf("Expected TypeError, got %s", err.ErrorType)
		}
	})

	t.Run("except star with multiple handlers", func(t *testing.T) {
		input := `
_caught_type = None
_caught_val = None
try:
    raise ExceptionGroup("eg", [TypeError("err1"), ValueError("err2")])
except* TypeError as e:
    _caught_type = e
except* ValueError as e:
    _caught_val = e
_result = 1
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %s", result.Type())
		}
	})

	t.Run("plain except does not catch ExceptionGroup", func(t *testing.T) {
		input := `
_caught = 0
try:
    try:
        raise ExceptionGroup("eg", [TypeError("err1")])
    except TypeError:
        _caught = 1
except:
    _caught = 2
`
		result := runTestCodeGetGlobal(t, input, "_caught")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		got := result.(*objects.Integer).Value
		if got != 2 {
			t.Errorf("Expected 2 (plain except catches EG), got %d", got)
		}
	})
}

func TestEllipsis(t *testing.T) {
	t.Run("ellipsis literal", func(t *testing.T) {
		input := `_result = ...`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.ELLIPSIS_OBJ {
			t.Fatalf("Expected ELLIPSIS, got %s", result.Type())
		}
		if result.Inspect() != "Ellipsis" {
			t.Errorf("Expected 'Ellipsis', got %q", result.Inspect())
		}
	})

	t.Run("Ellipsis identifier", func(t *testing.T) {
		input := `_result = Ellipsis`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.ELLIPSIS_OBJ {
			t.Fatalf("Expected ELLIPSIS, got %s", result.Type())
		}
	})

	t.Run("ellipsis is truthy", func(t *testing.T) {
		input := `
_result = bool(...)
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.BOOLEAN_OBJ {
			t.Fatalf("Expected BOOLEAN, got %s", result.Type())
		}
		got := result.(*objects.Boolean).Value
		if got != true {
			t.Errorf("Expected true, got %v", got)
		}
	})

	t.Run("ellipsis equality", func(t *testing.T) {
		input := `
_result = ... == Ellipsis
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.BOOLEAN_OBJ {
			t.Fatalf("Expected BOOLEAN, got %s", result.Type())
		}
		got := result.(*objects.Boolean).Value
		if got != true {
			t.Errorf("Expected true, got %v", got)
		}
	})
}

func TestComplex(t *testing.T) {
	t.Run("complex literal", func(t *testing.T) {
		input := `_result = 2j`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.COMPLEX_OBJ {
			t.Fatalf("Expected COMPLEX, got %s", result.Type())
		}
		c := result.(*objects.Complex)
		if c.Real != 0 || c.Imag != 2 {
			t.Errorf("Expected (0+2j), got (%g+%gj)", c.Real, c.Imag)
		}
	})

	t.Run("complex addition", func(t *testing.T) {
		input := `_result = 3 + 4j`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.COMPLEX_OBJ {
			t.Fatalf("Expected COMPLEX, got %s", result.Type())
		}
		c := result.(*objects.Complex)
		if c.Real != 3 || c.Imag != 4 {
			t.Errorf("Expected (3+4j), got (%g+%gj)", c.Real, c.Imag)
		}
	})

	t.Run("complex arithmetic", func(t *testing.T) {
		input := `_result = (1+2j) + (3+4j)`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		c := result.(*objects.Complex)
		if c.Real != 4 || c.Imag != 6 {
			t.Errorf("Expected (4+6j), got (%g+%gj)", c.Real, c.Imag)
		}
	})

	t.Run("complex multiplication", func(t *testing.T) {
		input := `_result = 2j * 3j`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		c := result.(*objects.Complex)
		if c.Real != -6 {
			t.Errorf("Expected real=-6, got %g", c.Real)
		}
	})

	t.Run("complex builtin", func(t *testing.T) {
		input := `_result = complex(3, 4)`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		c := result.(*objects.Complex)
		if c.Real != 3 || c.Imag != 4 {
			t.Errorf("Expected (3+4j), got (%g+%gj)", c.Real, c.Imag)
		}
	})

	t.Run("abs of complex", func(t *testing.T) {
		input := `_result = abs(3 + 4j)`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.FLOAT_OBJ {
			t.Fatalf("Expected FLOAT, got %s", result.Type())
		}
		got := result.(*objects.Float).Value
		if got != 5.0 {
			t.Errorf("Expected 5.0, got %g", got)
		}
	})

	t.Run("complex equality", func(t *testing.T) {
		input := `_result = (1+2j) == (1+2j)`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		got := result.(*objects.Boolean).Value
		if got != true {
			t.Errorf("Expected true, got %v", got)
		}
	})

	t.Run("complex negation", func(t *testing.T) {
		input := `_result = -(3+4j)`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		c := result.(*objects.Complex)
		if c.Real != -3 || c.Imag != -4 {
			t.Errorf("Expected (-3-4j), got (%g+%gj)", c.Real, c.Imag)
		}
	})

	t.Run("bool of complex", func(t *testing.T) {
		input := `
_a = bool(0j)
_b = bool(1j)
_result = _a == False and _b == True
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		got := result.(*objects.Boolean).Value
		if got != true {
			t.Errorf("Expected true, got %v", got)
		}
	})
}

func TestDictViews(t *testing.T) {
	t.Run("dict keys view object", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})

		dk := objects.NewDictKeys(d)
		if dk.Type() != objects.DICT_KEYS_OBJ {
			t.Fatalf("Expected DICT_KEYS, got %s", dk.Type())
		}
		if dk.Len() != 2 {
			t.Fatalf("Expected len 2, got %d", dk.Len())
		}
		keys := dk.ToList()
		if len(keys) != 2 {
			t.Fatalf("Expected 2 keys, got %d", len(keys))
		}
		// Test dynamic behavior - add a key to the dict
		d.Set(&objects.String{Value: "c"}, &objects.Integer{Value: 3})
		if dk.Len() != 3 {
			t.Fatalf("Expected dynamic len 3, got %d", dk.Len())
		}
	})

	t.Run("dict values view object", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})

		dv := objects.NewDictValues(d)
		if dv.Type() != objects.DICT_VALUES_OBJ {
			t.Fatalf("Expected DICT_VALUES, got %s", dv.Type())
		}
		if dv.Len() != 2 {
			t.Fatalf("Expected len 2, got %d", dv.Len())
		}
		vals := dv.ToList()
		if len(vals) != 2 {
			t.Fatalf("Expected 2 values, got %d", len(vals))
		}
	})

	t.Run("dict items view object", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})

		di := objects.NewDictItems(d)
		if di.Type() != objects.DICT_ITEMS_OBJ {
			t.Fatalf("Expected DICT_ITEMS, got %s", di.Type())
		}
		if di.Len() != 2 {
			t.Fatalf("Expected len 2, got %d", di.Len())
		}
		items := di.ToList()
		if len(items) != 2 {
			t.Fatalf("Expected 2 items, got %d", len(items))
		}
		// Items should be tuples
		for _, item := range items {
			if item.Type() != objects.TUPLE_OBJ {
				t.Fatalf("Expected TUPLE, got %s", item.Type())
			}
		}
	})

	t.Run("dict keys inspect", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})

		dk := objects.NewDictKeys(d)
		inspect := dk.Inspect()
		if !strings.HasPrefix(inspect, "dict_keys([") {
			t.Errorf("Expected inspect to start with 'dict_keys([', got %q", inspect)
		}
	})

	t.Run("dict values inspect", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})

		dv := objects.NewDictValues(d)
		inspect := dv.Inspect()
		if !strings.HasPrefix(inspect, "dict_values([") {
			t.Errorf("Expected inspect to start with 'dict_values([', got %q", inspect)
		}
	})

	t.Run("dict items inspect", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})

		di := objects.NewDictItems(d)
		inspect := di.Inspect()
		if !strings.HasPrefix(inspect, "dict_items([") {
			t.Errorf("Expected inspect to start with 'dict_items([', got %q", inspect)
		}
	})

	t.Run("dict keys getitem", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})

		dk := objects.NewDictKeys(d)
		val, ok := dk.GetItem(0)
		if !ok {
			t.Fatal("Expected GetItem(0) to succeed")
		}
		if val.Type() != objects.STRING_OBJ {
			t.Fatalf("Expected STRING, got %s", val.Type())
		}

		// Negative index
		val, ok = dk.GetItem(-1)
		if !ok {
			t.Fatal("Expected GetItem(-1) to succeed")
		}
		if val.(*objects.String).Value != "b" {
			t.Errorf("Expected 'b', got %q", val.(*objects.String).Value)
		}

		// Out of bounds
		_, ok = dk.GetItem(10)
		if ok {
			t.Fatal("Expected GetItem(10) to fail")
		}
	})

	t.Run("dict values getitem", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})

		dv := objects.NewDictValues(d)
		val, ok := dv.GetItem(0)
		if !ok {
			t.Fatal("Expected GetItem(0) to succeed")
		}
		if val.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %s", val.Type())
		}
	})

	t.Run("dict items getitem", func(t *testing.T) {
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})

		di := objects.NewDictItems(d)
		val, ok := di.GetItem(0)
		if !ok {
			t.Fatal("Expected GetItem(0) to succeed")
		}
		if val.Type() != objects.TUPLE_OBJ {
			t.Fatalf("Expected TUPLE, got %s", val.Type())
		}
		tuple := val.(*objects.Tuple)
		if len(tuple.Elements) != 2 {
			t.Fatalf("Expected tuple with 2 elements, got %d", len(tuple.Elements))
		}
	})
}

func TestMetaclass(t *testing.T) {
	t.Run("class with metaclass __call__", func(t *testing.T) {
		input := `
class Meta:
    def __call__(self, cls):
        return 42

class Foo(metaclass=Meta):
    pass

_result = Foo()
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %s: %s", result.Type(), result.Inspect())
		}
		got := result.(*objects.Integer).Value
		if got != 42 {
			t.Errorf("Expected 42, got %d", got)
		}
	})

	t.Run("metaclass __call__ with args", func(t *testing.T) {
		input := `
class Meta:
    def __call__(self, cls, x):
        return x * 2

class Foo(metaclass=Meta):
    pass

_result = Foo(21)
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %s: %s", result.Type(), result.Inspect())
		}
		got := result.(*objects.Integer).Value
		if got != 42 {
			t.Errorf("Expected 42, got %d", got)
		}
	})

	t.Run("class without metaclass uses default instantiation", func(t *testing.T) {
		input := `
class Foo:
    def __init__(self, x):
        self.x = x

_result = Foo(10)
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.INSTANCE_OBJ {
			t.Fatalf("Expected INSTANCE, got %s", result.Type())
		}
	})

	t.Run("metaclass __init__ called on class creation", func(t *testing.T) {
		input := `
class Meta:
    def __init__(self, cls, name, bases):
        cls._registered = True

class Foo(metaclass=Meta):
    pass

_result = Foo._registered
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.BOOLEAN_OBJ {
			t.Fatalf("Expected BOOLEAN, got %s: %s", result.Type(), result.Inspect())
		}
		got := result.(*objects.Boolean).Value
		if !got {
			t.Errorf("Expected True, got %v", got)
		}
	})

	t.Run("metaclass __init__ receives class name", func(t *testing.T) {
		input := `
class Meta:
    def __init__(self, cls, name, bases):
        cls._name = name

class Foo(metaclass=Meta):
    pass

_result = Foo._name
`
		result := runTestCodeGetGlobal(t, input, "_result")
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Type() != objects.STRING_OBJ {
			t.Fatalf("Expected STRING, got %s: %s", result.Type(), result.Inspect())
		}
		got := result.(*objects.String).Value
		if got != "Foo" {
			t.Errorf("Expected 'Foo', got %q", got)
		}
	})
}
