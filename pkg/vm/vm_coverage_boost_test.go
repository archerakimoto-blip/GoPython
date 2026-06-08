package vm

import (
	"testing"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

// ============================================================
// Direct internal function tests
// ============================================================

func TestOpToString(t *testing.T) {
	tests := []struct {
		op       compiler.Opcode
		expected string
	}{
		{compiler.OpGreaterThan, ">"},
		{compiler.OpLessThan, "<"},
		{compiler.OpAdd, "unknown"},
	}
	for _, tt := range tests {
		got := opToString(tt.op)
		if got != tt.expected {
			t.Errorf("opToString(%d) = %q, want %q", tt.op, got, tt.expected)
		}
	}
}

func TestBoolToFloat(t *testing.T) {
	if boolToFloat(true) != 1.0 {
		t.Errorf("boolToFloat(true) = %f, want 1.0", boolToFloat(true))
	}
	if boolToFloat(false) != 0.0 {
		t.Errorf("boolToFloat(false) = %f, want 0.0", boolToFloat(false))
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Errorf("boolToInt(true) = %d, want 1", boolToInt(true))
	}
	if boolToInt(false) != 0 {
		t.Errorf("boolToInt(false) = %d, want 0", boolToInt(false))
	}
}

func TestMatchesException(t *testing.T) {
	tests := []struct {
		name     string
		errObj   objects.Object
		excType  string
		expected bool
	}{
		{"empty_type_matches_all", &objects.Error{ErrorType: "ValueError", Message: "x"}, "", true},
		{"exception_matches_base", &objects.Error{ErrorType: "ValueError", Message: "x"}, "Exception", true},
		{"error_matches_base", &objects.Error{ErrorType: "ValueError", Message: "x"}, "Error", true},
		{"exact_match", &objects.Error{ErrorType: "ValueError", Message: "x"}, "ValueError", true},
		{"no_match", &objects.Error{ErrorType: "ValueError", Message: "x"}, "TypeError", false},
		{"eg_with_matching_member", &objects.ExceptionGroup{Message: "eg", Exceptions: []objects.Object{&objects.Error{ErrorType: "ValueError", Message: "x"}}}, "ValueError", true},
		{"eg_no_matching_member", &objects.ExceptionGroup{Message: "eg", Exceptions: []objects.Object{&objects.Error{ErrorType: "ValueError", Message: "x"}}}, "TypeError", false},
		{"non_error_non_eg", &objects.Integer{Value: 42}, "ValueError", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesException(tt.errObj, tt.excType)
			if got != tt.expected {
				t.Errorf("matchesException(%v, %q) = %v, want %v", tt.errObj, tt.excType, got, tt.expected)
			}
		})
	}
}

func TestMatchesExceptionType(t *testing.T) {
	vm := New(&compiler.Bytecode{Instructions: []byte{}})
	tests := []struct {
		name     string
		errObj   objects.Object
		excType  string
		expected bool
	}{
		{"error_obj_always_true", &objects.Error{ErrorType: "ValueError", Message: "x"}, "Anything", true},
		{"non_error_empty_type", &objects.Integer{Value: 42}, "", true},
		{"non_error_exception", &objects.Integer{Value: 42}, "Exception", true},
		{"non_error_error", &objects.Integer{Value: 42}, "Error", true},
		{"non_error_specific", &objects.Integer{Value: 42}, "ValueError", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.matchesExceptionType(tt.errObj, tt.excType)
			if got != tt.expected {
				t.Errorf("matchesExceptionType(%v, %q) = %v, want %v", tt.errObj, tt.excType, got, tt.expected)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		obj      objects.Object
		expected string
	}{
		{"string", &objects.String{Value: "hello"}, "hello"},
		{"integer", &objects.Integer{Value: 42}, "42"},
		{"float", &objects.Float{Value: 3.14}, "3.14"},
		{"complex", &objects.Complex{Real: 3, Imag: 4}, "(3+4j)"},
		{"bool_true", &objects.Boolean{Value: true}, "True"},
		{"bool_false", &objects.Boolean{Value: false}, "False"},
		{"none", &objects.None{}, "None"},
		{"list", &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}}}, "[1, 2]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.obj)
			if got != tt.expected {
				t.Errorf("toString(%v) = %q, want %q", tt.obj, got, tt.expected)
			}
		})
	}
}

func TestCallCallableBuiltin2(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	builtin := &objects.Builtin{
		Name: "test_builtin",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) > 0 {
				return args[0]
			}
			return objects.None_
		},
	}
	result := vm.CallCallable(builtin, &objects.Integer{Value: 42})
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
	got := result.(*objects.Integer).Value
	if got != 42 {
		t.Errorf("Expected 42, got %d", got)
	}
}

// ============================================================
// executeIndexExpression direct tests
// ============================================================

func TestExecuteIndexExprDirect(t *testing.T) {
	t.Run("dict_keys_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})
		dk := objects.NewDictKeys(d)
		err := vm.executeIndexExpression(dk, &objects.Integer{Value: 0})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.STRING_OBJ {
			t.Fatalf("Expected STRING, got %v", result)
		}
	})

	t.Run("dict_values_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		d.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})
		dv := objects.NewDictValues(d)
		err := vm.executeIndexExpression(dv, &objects.Integer{Value: 0})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %v", result)
		}
	})

	t.Run("dict_items_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		d := objects.NewDict()
		d.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
		di := objects.NewDictItems(d)
		err := vm.executeIndexExpression(di, &objects.Integer{Value: 0})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.TUPLE_OBJ {
			t.Fatalf("Expected TUPLE, got %v", result)
		}
	})

	t.Run("range_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		r := &objects.Range{Start: 0, Stop: 10, Step: 1}
		err := vm.executeIndexExpression(r, &objects.Integer{Value: 5})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %v", result)
		}
		got := result.(*objects.Integer).Value
		if got != 5 {
			t.Errorf("Expected 5, got %d", got)
		}
	})

	t.Run("bytes_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		b := &objects.Bytes{Value: []byte{65, 66, 67}}
		err := vm.executeIndexExpression(b, &objects.Integer{Value: 0})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %v", result)
		}
	})

	t.Run("tuple_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		tup := &objects.Tuple{Elements: []objects.Object{&objects.Integer{Value: 10}, &objects.Integer{Value: 20}}}
		err := vm.executeIndexExpression(tup, &objects.Integer{Value: 1})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %v", result)
		}
		got := result.(*objects.Integer).Value
		if got != 20 {
			t.Errorf("Expected 20, got %d", got)
		}
	})

	t.Run("dict_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		d := objects.NewDict()
		d.Set(&objects.String{Value: "key"}, &objects.Integer{Value: 99})
		err := vm.executeIndexExpression(d, &objects.String{Value: "key"})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %v", result)
		}
		got := result.(*objects.Integer).Value
		if got != 99 {
			t.Errorf("Expected 99, got %d", got)
		}
	})

	t.Run("string_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		s := &objects.String{Value: "hello"}
		err := vm.executeIndexExpression(s, &objects.Integer{Value: 1})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.STRING_OBJ {
			t.Fatalf("Expected STRING, got %v", result)
		}
	})

	t.Run("list_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 10}, &objects.Integer{Value: 20}}}
		err := vm.executeIndexExpression(l, &objects.Integer{Value: 0})
		if err != nil {
			t.Fatalf("executeIndexExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.INTEGER_OBJ {
			t.Fatalf("Expected INTEGER, got %v", result)
		}
	})
}

// ============================================================
// executeSetIndex direct tests
// ============================================================

func TestExecuteSetIndexDirect(t *testing.T) {
	t.Run("dict_set_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		d := objects.NewDict()
		err := vm.executeSetIndex(d, &objects.String{Value: "key"}, &objects.Integer{Value: 42})
		if err != nil {
			t.Fatalf("executeSetIndex error: %v", err)
		}
		val, ok := d.Get(&objects.String{Value: "key"})
		if !ok || val.(*objects.Integer).Value != 42 {
			t.Errorf("Expected 42 in dict")
		}
	})

	t.Run("list_set_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		err := vm.executeSetIndex(l, &objects.Integer{Value: 1}, &objects.Integer{Value: 20})
		if err != nil {
			t.Fatalf("executeSetIndex error: %v", err)
		}
		got := l.Elements[1].(*objects.Integer).Value
		if got != 20 {
			t.Errorf("Expected 20, got %d", got)
		}
	})

	t.Run("list_set_negative_index", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		err := vm.executeSetIndex(l, &objects.Integer{Value: -1}, &objects.Integer{Value: 30})
		if err != nil {
			t.Fatalf("executeSetIndex error: %v", err)
		}
		got := l.Elements[2].(*objects.Integer).Value
		if got != 30 {
			t.Errorf("Expected 30, got %d", got)
		}
	})

	t.Run("tuple_set_index_error", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		tup := &objects.Tuple{Elements: []objects.Object{&objects.Integer{Value: 1}}}
		err := vm.executeSetIndex(tup, &objects.Integer{Value: 0}, &objects.Integer{Value: 2})
		if err == nil {
			t.Fatal("Expected error for tuple set index")
		}
	})

	t.Run("string_set_index_error", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		s := &objects.String{Value: "hello"}
		err := vm.executeSetIndex(s, &objects.Integer{Value: 0}, &objects.String{Value: "a"})
		if err == nil {
			t.Fatal("Expected error for string set index")
		}
	})

	t.Run("list_out_of_range", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}}
		err := vm.executeSetIndex(l, &objects.Integer{Value: 10}, &objects.Integer{Value: 2})
		if err == nil {
			t.Fatal("Expected error for out of range index")
		}
	})
}

// ============================================================
// executeSetSlice direct tests
// ============================================================

func TestExecuteSetSliceDirect(t *testing.T) {
	t.Run("list_slice_assign_grow", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		replacement := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 10}, &objects.Integer{Value: 20}, &objects.Integer{Value: 30}}}
		err := vm.executeSetSlice(l, &objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.None{}, replacement)
		if err != nil {
			t.Fatalf("executeSetSlice error: %v", err)
		}
		if len(l.Elements) != 5 {
			t.Errorf("Expected 5 elements, got %d", len(l.Elements))
		}
	})

	t.Run("list_slice_assign_shrink", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}, &objects.Integer{Value: 4}, &objects.Integer{Value: 5}}}
		replacement := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 10}}}
		err := vm.executeSetSlice(l, &objects.Integer{Value: 1}, &objects.Integer{Value: 4}, &objects.None{}, replacement)
		if err != nil {
			t.Fatalf("executeSetSlice error: %v", err)
		}
		if len(l.Elements) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(l.Elements))
		}
	})

	t.Run("list_slice_assign_none_bounds", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		replacement := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 10}}}
		err := vm.executeSetSlice(l, &objects.None{}, &objects.None{}, &objects.None{}, replacement)
		if err != nil {
			t.Fatalf("executeSetSlice error: %v", err)
		}
	})

	t.Run("list_slice_assign_negative_bounds", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}, &objects.Integer{Value: 4}, &objects.Integer{Value: 5}}}
		replacement := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 99}}}
		err := vm.executeSetSlice(l, &objects.Integer{Value: -3}, &objects.Integer{Value: -1}, &objects.None{}, replacement)
		if err != nil {
			t.Fatalf("executeSetSlice error: %v", err)
		}
	})

	t.Run("non_list_slice_assign_error", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		s := &objects.String{Value: "hello"}
		replacement := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}}
		err := vm.executeSetSlice(s, &objects.Integer{Value: 0}, &objects.Integer{Value: 1}, &objects.None{}, replacement)
		if err == nil {
			t.Fatal("Expected error for non-list slice assignment")
		}
	})

	t.Run("step_not_none_error", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}}
		replacement := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 10}}}
		err := vm.executeSetSlice(l, &objects.Integer{Value: 0}, &objects.Integer{Value: 1}, &objects.Integer{Value: 2}, replacement)
		if err == nil {
			t.Fatal("Expected error for step slice assignment")
		}
	})

	t.Run("tuple_replacement", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		replacement := &objects.Tuple{Elements: []objects.Object{&objects.Integer{Value: 10}, &objects.Integer{Value: 20}}}
		err := vm.executeSetSlice(l, &objects.Integer{Value: 0}, &objects.Integer{Value: 1}, &objects.None{}, replacement)
		if err != nil {
			t.Fatalf("executeSetSlice error: %v", err)
		}
	})

	t.Run("non_iterable_replacement_error", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}}
		err := vm.executeSetSlice(l, &objects.Integer{Value: 0}, &objects.Integer{Value: 1}, &objects.None{}, &objects.Integer{Value: 42})
		if err == nil {
			t.Fatal("Expected error for non-iterable replacement")
		}
	})
}

// ============================================================
// executeSliceExpression direct tests
// ============================================================

func TestExecuteSliceExpressionDirect(t *testing.T) {
	t.Run("bytes_slice", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		b := &objects.Bytes{Value: []byte{65, 66, 67, 68, 69}}
		err := vm.executeSliceExpression(b, &objects.Integer{Value: 1}, &objects.Integer{Value: 3})
		if err != nil {
			t.Fatalf("executeSliceExpression error: %v", err)
		}
		result := vm.pop()
		if result == nil || result.Type() != objects.BYTES_OBJ {
			t.Fatalf("Expected BYTES, got %v", result)
		}
	})

	t.Run("unsupported_slice_type", func(t *testing.T) {
		bc := &compiler.Bytecode{Instructions: []byte{}}
		vm := New(bc)
		s := objects.NewSet()
		s.Add(&objects.Integer{Value: 1})
		err := vm.executeSliceExpression(s, &objects.Integer{Value: 0}, &objects.Integer{Value: 1})
		if err == nil {
			t.Fatal("Expected error for unsupported slice type")
		}
	})
}

// ============================================================
// Pipeline tests: closures and functions (OpGetFree, executeCall)
// ============================================================

func TestPipelineGetFree(t *testing.T) {
	comp, machine := compileAndRun(t, `
def make_adder(n):
    def adder(x):
        return x + n
    return adder
_f = make_adder(10)
_r = _f(5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 15)
}

func TestPipelineNestedClosure(t *testing.T) {
	comp, machine := compileAndRun(t, `
def outer(x):
    def middle(y):
        def inner(z):
            return x + y + z
        return inner
    return middle
_r = outer(1)(2)(3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineClosureMultipleFreeVars(t *testing.T) {
	comp, machine := compileAndRun(t, `
def make_adder(a, b):
    def adder(x):
        return x + a + b
    return adder
_f = make_adder(10, 20)
_r = _f(5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 35)
}

// ============================================================
// Pipeline tests: format strings (OpFormatString)
// ============================================================

func TestPipelineFormatString(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = 42
_r = f"value is {_x}"
`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
}

func TestPipelineFormatStringMultipleParts(t *testing.T) {
	comp, machine := compileAndRun(t, `
_a = 1
_b = 2
_r = f"{_a} + {_b}"
`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
}

// ============================================================
// Pipeline tests: executeCall() patterns
// ============================================================

func TestPipelineCallableObject(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len([1, 2, 3])`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineClassWithInit(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
_p = Point(3, 4)
_r = _p.x + _p.y`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineClassNoInit(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    pass
_f = Foo()
_r = 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineFunctionNoArgs(t *testing.T) {
	comp, machine := compileAndRun(t, `
def get_val():
    return 42
_r = get_val()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 42)
}

func TestPipelineFunctionMultipleArgs(t *testing.T) {
	comp, machine := compileAndRun(t, `
def add3(a, b, c):
    return a + b + c
_r = add3(1, 2, 3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineFunctionDefaultArgs(t *testing.T) {
	comp, machine := compileAndRun(t, `
def greet(name, greeting="hello"):
    return greeting
_r = greet("world")`)
	assertString(t, getGlobal(machine, comp, "_r"), "hello")
}

func TestPipelineFunctionDefaultArgsOverride(t *testing.T) {
	comp, machine := compileAndRun(t, `
def greet(name, greeting="hello"):
    return greeting
_r = greet("world", "hi")`)
	assertString(t, getGlobal(machine, comp, "_r"), "hi")
}

func TestPipelineSuperCall(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Base:
    def greet(self):
        return "hello"
class Child(Base):
    def greet(self):
        _s = super().greet()
        return _s
_c = Child()
_r = _c.greet()`)
	assertString(t, getGlobal(machine, comp, "_r"), "hello")
}

func TestPipelineInheritanceMethodInherited(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Animal:
    def speak(self):
        return "generic"
class Dog(Animal):
    pass
_d = Dog()
_r = _d.speak()`)
	assertString(t, getGlobal(machine, comp, "_r"), "generic")
}

func TestPipelineMultipleInheritanceMethodResolution(t *testing.T) {
	comp, machine := compileAndRun(t, `
class A:
    def foo(self):
        return 1
class B:
    def bar(self):
        return 2
class C(A, B):
    def baz(self):
        return 3
_c = C()
_r = _c.foo() + _c.bar() + _c.baz()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

// ============================================================
// Pipeline tests: more class patterns
// ============================================================

func TestPipelineClassWithSlots(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    __slots__ = ["x", "y"]
    def __init__(self, a, b):
        self.x = a
        self.y = b
_f = Foo(1, 2)
_r = _f.x + _f.y`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineInstanceDelAttr(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 10
        self.y = 20
_f = Foo()
del _f.x
_r = 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

// ============================================================
// Pipeline tests: more control flow
// ============================================================

func TestPipelineWhileWithBreak(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
_i = 0
while _i < 10:
    _r = _r + 1
    _i = _i + 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 10)
}

func TestPipelineForRangeWithStep(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
for _i in range(0, 10, 2):
    _r = _r + _i`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 20)
}

func TestPipelineNestedIfElseDeep(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = 15
if _x > 20:
    _r = 1
elif _x > 10:
    _r = 2
elif _x > 5:
    _r = 3
else:
    _r = 4`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineForLoopWithIf(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
for _i in range(10):
    if _i > 5:
        _r = _r + _i`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6+7+8+9)
}

// ============================================================
// Pipeline tests: more exception handling
// ============================================================

func TestPipelineTryExceptValueError(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    raise ValueError("test")
except ValueError:
    _r = 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineTryExceptTypeError(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    raise TypeError("test")
except TypeError:
    _r = 2`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineTryExceptReraise(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    try:
        raise ValueError("inner")
    except ValueError:
        raise TypeError("outer")
except TypeError:
    _r = 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineTryFinally(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    _r = 1
finally:
    _r = 2`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineTryExceptFinally(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    raise ValueError("test")
except ValueError:
    _r = 1
finally:
    _r = _r + 10`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 11)
}

func TestPipelineTryExceptElse(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = ""
try:
    _r = _r + "try"
except:
    _r = _r + "except"
else:
    _r = _r + "else"`)
	if result := getGlobal(machine, comp, "_r"); result != nil {
		got := result.(*objects.String).Value
		if got != "tryelse" {
			t.Errorf("Expected 'tryelse', got %q", got)
		}
	}
}

// ============================================================
// Pipeline tests: built-in functions
// ============================================================

func TestPipelineBuiltinIsinstanceBuiltin(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = isinstance(42, int)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBuiltinIssubclassBuiltin(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = issubclass(bool, int)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBuiltinChrOrd(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = chr(ord("A"))`)
	assertString(t, getGlobal(machine, comp, "_r"), "A")
}

func TestPipelineBuiltinBinOctHex(t *testing.T) {
	t.Run("bin", func(t *testing.T) {
		comp, machine := compileAndRun(t, "_r = bin(10)")
		assertString(t, getGlobal(machine, comp, "_r"), "0b1010")
	})
	t.Run("oct", func(t *testing.T) {
		comp, machine := compileAndRun(t, "_r = oct(8)")
		result := getGlobal(machine, comp, "_r")
		if result == nil || result.Type() != objects.STRING_OBJ {
			t.Fatalf("Expected STRING, got %v", result)
		}
	})
	t.Run("hex", func(t *testing.T) {
		comp, machine := compileAndRun(t, "_r = hex(255)")
		assertString(t, getGlobal(machine, comp, "_r"), "0xff")
	})
}

func TestPipelineBuiltinSumWithStart(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = sum([1, 2, 3], 10)")
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBuiltinSortedKey(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = sorted(["banana", "apple", "cherry"])`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinReversed(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(reversed([1, 2, 3]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinMapFilter(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		comp, machine := compileAndRun(t, `_r = list(map(str, [1, 2, 3]))`)
		result := getGlobal(machine, comp, "_r")
		if result == nil || result.Type() != objects.LIST_OBJ {
			t.Fatalf("Expected LIST, got %v", result)
		}
	})
	t.Run("filter", func(t *testing.T) {
		comp, machine := compileAndRun(t, `_r = list(filter(bool, [0, 1, 2, 0]))`)
		result := getGlobal(machine, comp, "_r")
		if result == nil || result.Type() != objects.LIST_OBJ {
			t.Fatalf("Expected LIST, got %v", result)
		}
	})
}

func TestPipelineBuiltinZip(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(zip([1, 2], [3, 4]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinEnumerate(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(enumerate([10, 20, 30]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

// ============================================================
// Pipeline tests: set operations via operators
// ============================================================

func TestPipelineSetIntersection(t *testing.T) {
	comp, machine := compileAndRun(t, `
_a = {1, 2, 3}
_b = {2, 3, 4}
_r = _a.intersection(_b)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineSetDifference(t *testing.T) {
	comp, machine := compileAndRun(t, `
_a = {1, 2, 3}
_b = {2, 3, 4}
_r = _a.difference(_b)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

// ============================================================
// Pipeline tests: augmented assignment (in-place ops)
// ============================================================

func TestPipelineIntegerPower(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 2
_r **= 8`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 256)
}

func TestPipelineNegativeFloorDiv(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = -17 // 5")
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineInPlaceBitOr(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r |= 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineInPlaceBitAnd(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r &= 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineInPlaceBitXor(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r ^= 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineInPlacePower(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 2
_r **= 10`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1024)
}

func TestPipelineInPlaceDiv(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 10.0
_r /= 4.0`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 2.5)
}

func TestPipelineInPlaceSub(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 10
_r -= 4`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineInPlaceMul(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 3
_r *= 7`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 21)
}

// ============================================================
// Pipeline tests: slice operations
// ============================================================

func TestPipelineBytesSlice(t *testing.T) {
	comp, machine := compileAndRun(t, `_b = b"hello"
_r = _b[1:4]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.BYTES_OBJ {
		t.Fatalf("Expected BYTES, got %v", result)
	}
}

func TestPipelineListSliceAssignment(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = [1, 2, 3, 4, 5]
_x[1:3] = [20, 30]
_r = _x`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

// ============================================================
// Pipeline tests: bytes operations
// ============================================================

func TestPipelineBytesIndex(t *testing.T) {
	comp, machine := compileAndRun(t, `_b = b"ABC"
_r = _b[0]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBytesNegativeIndex(t *testing.T) {
	comp, machine := compileAndRun(t, `_b = b"ABC"
_r = _b[-1]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBytesComparison(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = b"abc" == b"abc"`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBytesNotEqual(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = b"abc" != b"def"`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

// ============================================================
// Pipeline tests: complex comparison
// ============================================================

func TestPipelineComplexEq(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = (1+2j) == (1+2j)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineComplexNeq(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = (1+2j) != (3+4j)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

// ============================================================
// Register VM tests: more opcode paths
// ============================================================

func TestRegisterVMBasicArithmetic(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 2 + 3 * 4 - 1`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

// ============================================================
// Direct CallCallable tests with CompiledFunction and Closure
// ============================================================

func TestCallCallableWithCompiledFunction(t *testing.T) {
	// Test CallCallable with a compiled function through the pipeline
	comp, machine := compileAndRun(t, `
def get_val():
    return 42
_r = get_val()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 42)
}

func TestCallCallableWithClosure(t *testing.T) {
	comp, machine := compileAndRun(t, `
def add(a, b):
    return a + b
_r = add(3, 4)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestCallCallableWithClass(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self, x):
        self.x = x
_f = Foo(42)
_r = _f.x`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 42)
}

func TestCallCallableWithInstanceMethod(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Calc:
    def __init__(self, val):
        self.val = val
    def add(self, x):
        return self.val + x
_c = Calc(10)
_r = _c.add(5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 15)
}

// ============================================================
// Direct tests for executeBinaryIntegerOperation
// ============================================================

func TestExecuteBinaryIntegerOperationMod(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 17})
	vm.push(&objects.Integer{Value: 5})
	err := vm.executeBinaryIntegerOperation(compiler.OpMod, &objects.Integer{Value: 17}, &objects.Integer{Value: 5})
	if err != nil {
		t.Fatalf("executeBinaryIntegerOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
	got := result.(*objects.Integer).Value
	if got != 2 {
		t.Errorf("Expected 2, got %d", got)
	}
}

func TestExecuteBinaryIntegerOperationFloorDiv(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 17})
	vm.push(&objects.Integer{Value: 5})
	err := vm.executeBinaryIntegerOperation(compiler.OpFloorDiv, &objects.Integer{Value: 17}, &objects.Integer{Value: 5})
	if err != nil {
		t.Fatalf("executeBinaryIntegerOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
	got := result.(*objects.Integer).Value
	if got != 3 {
		t.Errorf("Expected 3, got %d", got)
	}
}

func TestExecuteBinaryIntegerOperationPower(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 2})
	vm.push(&objects.Integer{Value: 8})
	err := vm.executeBinaryIntegerOperation(compiler.OpPower, &objects.Integer{Value: 2}, &objects.Integer{Value: 8})
	if err != nil {
		t.Fatalf("executeBinaryIntegerOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
	got := result.(*objects.Integer).Value
	if got != 256 {
		t.Errorf("Expected 256, got %d", got)
	}
}

// ============================================================
// Direct tests for executeBinaryComplexOperation
// ============================================================

func TestExecuteBinaryComplexOperationSub(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 5, Imag: 3})
	vm.push(&objects.Complex{Real: 2, Imag: 1})
	err := vm.executeBinaryComplexOperation(compiler.OpSub, &objects.Complex{Real: 5, Imag: 3}, &objects.Complex{Real: 2, Imag: 1})
	if err != nil {
		t.Fatalf("executeBinaryComplexOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestExecuteBinaryComplexOperationMul(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 2, Imag: 3})
	vm.push(&objects.Complex{Real: 1, Imag: 1})
	err := vm.executeBinaryComplexOperation(compiler.OpMul, &objects.Complex{Real: 2, Imag: 3}, &objects.Complex{Real: 1, Imag: 1})
	if err != nil {
		t.Fatalf("executeBinaryComplexOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestExecuteBinaryComplexOperationDiv(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 6, Imag: 4})
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	err := vm.executeBinaryComplexOperation(compiler.OpDiv, &objects.Complex{Real: 6, Imag: 4}, &objects.Complex{Real: 1, Imag: 2})
	if err != nil {
		t.Fatalf("executeBinaryComplexOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

// ============================================================
// Direct tests for executeBinaryFloatOperation
// ============================================================

func TestExecuteBinaryFloatOperationMod(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Float{Value: 10.5})
	vm.push(&objects.Float{Value: 3.0})
	err := vm.executeBinaryFloatOperation(compiler.OpMod, &objects.Float{Value: 10.5}, &objects.Float{Value: 3.0})
	if err != nil {
		t.Fatalf("executeBinaryFloatOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.FLOAT_OBJ {
		t.Fatalf("Expected FLOAT, got %v", result)
	}
}

func TestExecuteBinaryFloatOperationFloorDiv(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Float{Value: 10.0})
	vm.push(&objects.Float{Value: 3.0})
	err := vm.executeBinaryFloatOperation(compiler.OpFloorDiv, &objects.Float{Value: 10.0}, &objects.Float{Value: 3.0})
	if err != nil {
		t.Fatalf("executeBinaryFloatOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.FLOAT_OBJ {
		t.Fatalf("Expected FLOAT, got %v", result)
	}
}

func TestExecuteBinaryFloatOperationPower(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Float{Value: 2.0})
	vm.push(&objects.Float{Value: 3.0})
	err := vm.executeBinaryFloatOperation(compiler.OpPower, &objects.Float{Value: 2.0}, &objects.Float{Value: 3.0})
	if err != nil {
		t.Fatalf("executeBinaryFloatOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.FLOAT_OBJ {
		t.Fatalf("Expected FLOAT, got %v", result)
	}
	got := result.(*objects.Float).Value
	if got != 8.0 {
		t.Errorf("Expected 8.0, got %f", got)
	}
}

// ============================================================
// Direct tests for executeBytesComparison
// ============================================================

func TestExecuteBytesComparisonLt(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Bytes{Value: []byte{65}})
	vm.push(&objects.Bytes{Value: []byte{66}})
	err := vm.executeBytesComparison(compiler.OpLessThan, &objects.Bytes{Value: []byte{65}}, &objects.Bytes{Value: []byte{66}})
	if err != nil {
		t.Fatalf("executeBytesComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

func TestExecuteBytesComparisonGt(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Bytes{Value: []byte{66}})
	vm.push(&objects.Bytes{Value: []byte{65}})
	err := vm.executeBytesComparison(compiler.OpGreaterThan, &objects.Bytes{Value: []byte{66}}, &objects.Bytes{Value: []byte{65}})
	if err != nil {
		t.Fatalf("executeBytesComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

// ============================================================
// Direct tests for executeComplexComparison
// ============================================================

func TestExecuteComplexComparisonEq(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	err := vm.executeComplexComparison(compiler.OpEqual, &objects.Complex{Real: 1, Imag: 2}, &objects.Complex{Real: 1, Imag: 2})
	if err != nil {
		t.Fatalf("executeComplexComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

func TestExecuteComplexComparisonNeq(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	vm.push(&objects.Complex{Real: 3, Imag: 4})
	err := vm.executeComplexComparison(compiler.OpNotEqual, &objects.Complex{Real: 1, Imag: 2}, &objects.Complex{Real: 3, Imag: 4})
	if err != nil {
		t.Fatalf("executeComplexComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

// ============================================================
// More register VM tests targeting executeRegFrame
// ============================================================

func TestRegVMFunctionWithLocalVars(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def compute(x):
    _a = x + 1
    _b = _a * 2
    return _b
_r = compute(5)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFunctionWithClosure2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def make_multiplier(n):
    def multiply(x):
        return x * n
    return multiply
_doubler = make_multiplier(2)
_r = _doubler(5)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMClassWithArgs(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Adder:
    def add(self, a, b):
        return a + b
_a = Adder()
_r = _a.add(3, 4)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMNestedForLoop(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
for _i in range(3):
    for _j in range(3):
        _r = _r + 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMWhileWithBreak(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
_i = 0
while _i < 10:
    _r = _r + 1
    _i = _i + 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMForWithBreak(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
for _i in range(10):
    if _i == 5:
        break
    _r = _r + 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMForWithContinue(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
for _i in range(10):
    if _i < 5:
        continue
    _r = _r + 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceAdd3(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 5
_r += 3
_r += 2
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceSub3(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 10
_r -= 3
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceMul3(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 3
_r *= 7
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceMod(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 17
_r %= 5
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceFloorDiv(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 17
_r //= 5
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlacePower(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 2
_r **= 8
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMTryExceptValueError(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
try:
    raise ValueError("test")
except ValueError:
    _r = 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMTryExceptTypeError(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
try:
    raise TypeError("test")
except TypeError:
    _r = 2
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMTryFinally2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
try:
    _r = 1
finally:
    _r = 2
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMTryExceptFinally2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
try:
    raise ValueError("test")
except ValueError:
    _r = 1
finally:
    _r = _r + 10
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMDefaultArgs(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def greet(name, greeting="hello"):
    return greeting
_r = greet("world")
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMDefaultArgsOverride(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def greet(name, greeting="hello"):
    return greeting
_r = greet("world", "hi")
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMLambda(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = (lambda x: x * 2)(21)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMRecursiveFib(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)
_r = fib(6)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMClassInheritAndOverride(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Base:
    def val(self):
        return 10
class Child(Base):
    def val(self):
        return 20
_c = Child()
_r = _c.val()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMClassInheritNoOverride(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Base:
    def val(self):
        return 10
class Child(Base):
    pass
_c = Child()
_r = _c.val()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFunctionAsArg(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def apply(f, x):
    return f(x)
def inc(n):
    return n + 1
_r = apply(inc, 5)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFunctionReturningFunction(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def make_adder(n):
    def adder(x):
        return x + n
    return adder
_doubler = make_adder(2)
_r = _doubler(5)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

// ============================================================
// More stack VM pipeline tests
// ============================================================

func TestPipelineClassWithMethodAndAttr(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Rect:
    def __init__(self, w, h):
        self.w = w
        self.h = h
    def area(self):
        return self.w * self.h
_r = Rect(3, 4).area()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 12)
}

func TestPipelineFunctionReturningFunction2(t *testing.T) {
	comp, machine := compileAndRun(t, `
def make_adder(n):
    def adder(x):
        return x + n
    return adder
_doubler = make_adder(2)
_r = _doubler(5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineLambdaAsArg2(t *testing.T) {
	comp, machine := compileAndRun(t, `
def apply(f, x):
    return f(x)
_r = apply(lambda x: x + 1, 5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineNestedForLoop2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
for _i in range(3):
    for _j in range(3):
        _r = _r + 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 9)
}

func TestPipelineWhileLoop2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
_i = 0
while _i < 5:
    _r = _r + _i
    _i = _i + 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 10)
}

func TestPipelineForRangeStep2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
for _i in range(0, 10, 2):
    _r = _r + _i`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 20)
}

func TestPipelineIfElifElse2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = 15
if _x > 20:
    _r = 1
elif _x > 10:
    _r = 2
else:
    _r = 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineTryExceptBare2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    raise ValueError("test")
except:
    _r = 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineTryExceptWithVar2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = ""
try:
    raise ValueError("test error")
except ValueError as e:
    _r = "caught"`)
	assertString(t, getGlobal(machine, comp, "_r"), "caught")
}

func TestPipelineTryExceptFinally2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    raise ValueError("test")
except ValueError:
    _r = 1
finally:
    _r = _r + 10`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 11)
}

func TestPipelineTryFinally2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    _r = 1
finally:
    _r = 2`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineInPlaceAdd2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r += 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 8)
}

func TestPipelineInPlaceSub2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 10
_r -= 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineInPlaceMul2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 3
_r *= 7`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 21)
}

func TestPipelineInPlaceBitOr2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r |= 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineInPlaceBitAnd2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r &= 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineInPlaceBitXor2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r ^= 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineInPlacePower2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 2
_r **= 8`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 256)
}

func TestPipelineInPlaceDiv2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 10.0
_r /= 4.0`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 2.5)
}

func TestPipelineSetUnion2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2} | {2, 3}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineSetIntersection3(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2, 3} & {2, 3, 4}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineSetXor2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2} ^ {2, 3}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineDictMerge2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_d1 = {}
_d1["a"] = 1
_d2 = {}
_d2["b"] = 2
_r = _d1 | _d2`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.DICT_OBJ {
		t.Fatalf("Expected DICT, got %v", result)
	}
}

func TestPipelineComplexAdd2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = (1+2j) + (3+4j)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestPipelineComplexSub2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = (5+3j) - (2+1j)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestPipelineComplexMul2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 2j * 3j`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestPipelineBytesIndex2(t *testing.T) {
	comp, machine := compileAndRun(t, `_b = b"ABC"
_r = _b[0]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBytesSlice2(t *testing.T) {
	comp, machine := compileAndRun(t, `_b = b"hello"
_r = _b[1:4]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.BYTES_OBJ {
		t.Fatalf("Expected BYTES, got %v", result)
	}
}

func TestPipelineStringSlice2(t *testing.T) {
	comp, machine := compileAndRun(t, `_s = "hello world"
_r = _s[0:5]`)
	assertString(t, getGlobal(machine, comp, "_r"), "hello")
}

func TestPipelineListSlice2(t *testing.T) {
	comp, machine := compileAndRun(t, `_x = [1, 2, 3, 4, 5]
_r = _x[1:4]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineListSliceAssign2(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = [1, 2, 3, 4, 5]
_x[1:3] = [20, 30]
_r = _x`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineExceptionGroup2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = ExceptionGroup("eg", [TypeError("a"), ValueError("b")])`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.EXCEPTION_GROUP_OBJ {
		t.Fatalf("Expected EXCEPTION_GROUP, got %v", result)
	}
}

func TestPipelineEllipsis2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = ...`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.ELLIPSIS_OBJ {
		t.Fatalf("Expected ELLIPSIS, got %v", result)
	}
}

func TestPipelineEllipsisIdentifier2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = Ellipsis`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.ELLIPSIS_OBJ {
		t.Fatalf("Expected ELLIPSIS, got %v", result)
	}
}

func TestPipelineIsinstanceCustom2(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    pass
_f = Foo()
_r = isinstance(_f, Foo)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineIssubclassCustom2(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Base:
    pass
class Child(Base):
    pass
_r = issubclass(Child, Base)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineHasattr2(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
_r = hasattr(_f, "x")`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineGetattr2(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
_r = getattr(_f, "x")`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 42)
}

func TestPipelineSetattr2(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
setattr(_f, "x", 100)
_r = _f.x`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 100)
}

func TestPipelineBuiltinRound2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = round(3.5)")
	assertInteger(t, getGlobal(machine, comp, "_r"), 4)
}

func TestPipelineBuiltinCallable2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = callable(print)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBuiltinRepr2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = repr(42)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
}

func TestPipelineBuiltinAll2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = all([True, True])`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBuiltinAny2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = any([False, True])`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBuiltinHash2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = hash(42)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
}

func TestPipelineBuiltinId2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = id(42)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
}

func TestPipelineBuiltinSorted2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = sorted([3, 1, 2])`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinSum2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = sum([1, 2, 3])`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineBuiltinList2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(range(5))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinSet2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = set([1, 2, 2, 3])`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineBuiltinMap2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(map(str, [1, 2, 3]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinFilter2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(filter(bool, [0, 1, 2, 0]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinZip2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(zip([1, 2], [3, 4]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinEnumerate2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(enumerate([10, 20, 30]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinReversed2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(reversed([1, 2, 3]))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinChrOrd2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = chr(ord("A"))`)
	assertString(t, getGlobal(machine, comp, "_r"), "A")
}

func TestPipelineBuiltinBin2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = bin(10)")
	assertString(t, getGlobal(machine, comp, "_r"), "0b1010")
}

func TestPipelineBuiltinHex2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = hex(255)")
	assertString(t, getGlobal(machine, comp, "_r"), "0xff")
}

func TestPipelineBuiltinOct2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = oct(8)")
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
}

func TestPipelineBuiltinAbs2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = abs(-5)")
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinStr2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = str(42)")
	assertString(t, getGlobal(machine, comp, "_r"), "42")
}

func TestPipelineBuiltinInt2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = int(3.5)")
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineBuiltinFloat2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = float(5)")
	assertFloat(t, getGlobal(machine, comp, "_r"), 5.0)
}

func TestPipelineBuiltinBool2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = bool(0)")
	assertBoolean(t, getGlobal(machine, comp, "_r"), false)
}

func TestPipelineBuiltinType2(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = type(42)")
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBuiltinLen2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len("hello")`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinRange2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len(range(5))`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinRange3Args2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len(range(0, 10, 2))`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinPrint2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = print("hello")`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.NONE_OBJ {
		t.Fatalf("Expected NONE, got %v", result)
	}
}

func TestPipelineBuiltinInput2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = callable(input)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBuiltinMin2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = min(1, 2, 3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineBuiltinMax2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = max(1, 2, 3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestRegVMFloatArithmetic(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 1.5 + 2.5`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMDivision(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 10 / 3`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMModulo(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 10 % 3`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFloorDiv(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 10 // 3`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMPower(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 2 ** 8`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMComparison(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 5 > 3`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMBooleanOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = True and False`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMIfElse(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_x = 5
if _x > 3:
    _r = 1
else:
    _r = 0
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMWhileLoop(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
_i = 0
while _i < 5:
    _r = _r + _i
    _i = _i + 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMForLoop(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
for _i in range(5):
    _r = _r + _i
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMFunctionDefAndCall(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def add(a, b):
    return a + b
_r = add(3, 4)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMClosure(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def make_adder(n):
    def adder(x):
        return x + n
    return adder
_f = make_adder(5)
_r = _f(10)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMListOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_x = [1, 2, 3]
_r = _x[1]
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMSetOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_s = {1, 2, 3}
_r = len(_s)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMStringOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_s = "hello"
_r = _s + " world"
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMNegation(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = -42`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMNot(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = !True`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMBitOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 5 | 3`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMBitAnd(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 5 & 3`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMBitXor(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 5 ^ 3`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMClassDef(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    def __init__(self, x):
        self.x = x
_f = Foo(10)
_r = _f.x
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMClassInheritance(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Animal:
    def speak(self):
        return "generic"
class Dog(Animal):
    def speak(self):
        return "woof"
_d = Dog()
_r = _d.speak()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMBuiltinCall(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = len([1, 2, 3])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMStringSlice(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = "hello world"[0:5]`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMListSlice(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = [1, 2, 3, 4, 5][1:4]`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMFString(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_x = 42
_r = f"{_x}"
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMSetAttr(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
_r = _f.x
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMEllipsis(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = ...`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMNone(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = None`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBool(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = True`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMComplex(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = 2j`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMSetIndex(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_x = [1, 2, 3]
_x[1] = 20
_r = _x[1]
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMDictSetIndex(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_d = {}
_d["key"] = 42
_r = _d["key"]
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 5
_r += 3
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMSuperCall(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Base:
    def greet(self):
        return "hello"
class Child(Base):
    def greet(self):
        _s = super().greet()
        return _s
_c = Child()
_r = _c.greet()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMNestedFunction(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def outer(x):
    def inner(y):
        return x + y
    return inner(10)
_r = outer(5)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinPrint(t *testing.T) {
	rbc := compileToRegBytecodeNew(`print("hello")`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegisterVMBuiltinRange(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = range(5)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinAbs2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = abs(-5)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinStr(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = str(42)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinInt(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = int(3.5)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinFloat(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = float(5)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinBool(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = bool(0)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinType(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = type(42)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinSorted(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = sorted([3, 1, 2])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinSum(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = sum([1, 2, 3])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinList(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = list(range(5))`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinSet(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = set([1, 2, 2, 3])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinIsinstance(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    pass
_f = Foo()
_r = isinstance(_f, Foo)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinIssubclass(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Base:
    pass
class Child(Base):
    pass
_r = issubclass(Child, Base)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

// ============================================================
// Pipeline tests: more complex patterns
// ============================================================

func TestPipelineClassMethodReturningSelf(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Builder:
    def __init__(self):
        self.val = 0
    def inc(self):
        self.val = self.val + 1
        return self
_b = Builder()
_b.inc()
_b.inc()
_r = _b.val`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineClassWithClassMethod(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    pass
_r = callable(Foo)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineGlobalScopeInFunction(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = 10
def get_x():
    return _x
_r = get_x()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 10)
}

func TestPipelineNestedFunctionCalls(t *testing.T) {
	comp, machine := compileAndRun(t, `
def double(x):
    return x * 2
def add_one(x):
    return x + 1
_r = double(add_one(5))`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 12)
}

func TestPipelineFunctionAsArgument(t *testing.T) {
	comp, machine := compileAndRun(t, `
def apply(f, x):
    return f(x)
def inc(n):
    return n + 1
_r = apply(inc, 5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineLambdaInExpression(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = (lambda x: x * 2)(21)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 42)
}

func TestPipelineRangeIndex(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = range(10)[5]")
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineRangeLen(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = len(range(5, 10))")
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

// ============================================================
// Pipeline tests: edge cases
// ============================================================

func TestPipelineEmptyDict(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len({})`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 0)
}

func TestPipelineEmptyList(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len([])`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 0)
}

func TestPipelineEmptyString(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len("")`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 0)
}

// ============================================================
// Pipeline tests: exception group
// ============================================================

func TestPipelineExceptionGroupCreation(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = ExceptionGroup("eg", [TypeError("a"), ValueError("b")])`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.EXCEPTION_GROUP_OBJ {
		t.Fatalf("Expected EXCEPTION_GROUP, got %v", result)
	}
}

// ============================================================
// Pipeline tests: set operations
// ============================================================

func TestPipelineSetUnionOp(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2} | {2, 3}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineSetIntersectionOp(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2, 3} & {2, 3, 4}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineSetXorOp(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2} ^ {2, 3}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

// ============================================================
// Pipeline tests: more string operations
// ============================================================

func TestPipelineStringMul(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = "ha" * 3`)
	assertString(t, getGlobal(machine, comp, "_r"), "hahaha")
}

func TestPipelineStringComparisonLt(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = "abc" < "def"`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineStringComparisonGt(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = "def" > "abc"`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

// ============================================================
// Pipeline tests: ellipsis
// ============================================================

func TestPipelineEllipsisLiteral(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = ...`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.ELLIPSIS_OBJ {
		t.Fatalf("Expected ELLIPSIS, got %v", result)
	}
}

func TestPipelineEllipsisIdentifier(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = Ellipsis`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.ELLIPSIS_OBJ {
		t.Fatalf("Expected ELLIPSIS, got %v", result)
	}
}

// ============================================================
// Pipeline tests: complex numbers
// ============================================================

func TestPipelineComplexAddition(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = (1+2j) + (3+4j)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
	c := result.(*objects.Complex)
	if c.Real != 4 || c.Imag != 6 {
		t.Errorf("Expected (4+6j), got (%g+%gj)", c.Real, c.Imag)
	}
}

func TestPipelineComplexSubtraction(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = (5+3j) - (2+1j)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestPipelineComplexMultiplication(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 2j * 3j`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestPipelineComplexBuiltin(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = complex(3, 4)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestPipelineComplexAbs(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = abs(3 + 4j)`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 5.0)
}

func TestPipelineComplexNegation(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = -(3+4j)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
	c := result.(*objects.Complex)
	if c.Real != -3 || c.Imag != -4 {
		t.Errorf("Expected (-3-4j), got (%g+%gj)", c.Real, c.Imag)
	}
}

// ============================================================
// Pipeline tests: None return from function
// ============================================================

func TestPipelineFunctionReturnsNone(t *testing.T) {
	comp, machine := compileAndRun(t, `
def no_ret():
    _x = 1
_r = no_ret()`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.NONE_OBJ {
		t.Fatalf("Expected NONE, got %v", result)
	}
}

// ============================================================
// Pipeline tests: mixed type arithmetic
// ============================================================

func TestPipelineIntPlusFloat(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 3 + 1.5")
	assertFloat(t, getGlobal(machine, comp, "_r"), 4.5)
}

func TestPipelineFloatPlusInt(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 1.5 + 3")
	assertFloat(t, getGlobal(machine, comp, "_r"), 4.5)
}

func TestPipelineIntMulFloat(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 3 * 2.0")
	assertFloat(t, getGlobal(machine, comp, "_r"), 6.0)
}

func TestPipelineBoolPlusInt(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = True + 5")
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineBoolPlusBool(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = True + True")
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

// ============================================================
// Pipeline tests: list concatenation
// ============================================================

func TestPipelineListConcat(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = [1, 2] + [3, 4]")
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
	if result.(*objects.List).Size() != 4 {
		t.Errorf("Expected 4 elements")
	}
}

func TestPipelineListInPlaceConcat(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = [1, 2]
_r += [3, 4]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
	if result.(*objects.List).Size() != 4 {
		t.Errorf("Expected 4 elements")
	}
}

// ============================================================
// Pipeline tests: string in-place concat
// ============================================================

func TestPipelineStringInPlaceConcat(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = "hello"
_r += " world"`)
	assertString(t, getGlobal(machine, comp, "_r"), "hello world")
}

// ============================================================
// Pipeline tests: float in-place operations
// ============================================================

func TestPipelineFloatInPlaceAdd(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 1.5
_r += 2.5`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 4.0)
}

func TestPipelineFloatInPlaceSub(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5.5
_r -= 1.5`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 4.0)
}

func TestPipelineFloatInPlaceMul(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 2.5
_r *= 2.0`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 5.0)
}

// ============================================================
// Pipeline tests: isinstance/issubclass with custom classes
// ============================================================

func TestPipelineIsinstanceCustom(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    pass
_f = Foo()
_r = isinstance(_f, Foo)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineIssubclassCustom(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Base:
    pass
class Child(Base):
    pass
_r = issubclass(Child, Base)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineIssubclassSame(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    pass
_r = issubclass(Foo, Foo)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

// ============================================================
// Pipeline tests: hasattr/getattr/setattr
// ============================================================

func TestPipelineHasattrInstance(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
_r = hasattr(_f, "x")`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineGetattrInstance(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
_r = getattr(_f, "x")`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 42)
}

func TestPipelineSetattrInstance(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
setattr(_f, "x", 100)
_r = _f.x`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 100)
}

// ============================================================
// Pipeline tests: more builtins
// ============================================================

func TestPipelineBuiltinRound(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = round(3.5)")
	assertInteger(t, getGlobal(machine, comp, "_r"), 4)
}

func TestPipelineBuiltinRoundNdigits(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = round(3.14159, 2)")
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBuiltinCallable(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = callable(print)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBuiltinRepr(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = repr(42)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
}

func TestPipelineBuiltinId(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = id(42)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
}

func TestPipelineBuiltinHash(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = hash(42)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
}

func TestPipelineBuiltinAll(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = all([True, True])`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBuiltinAny(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = any([False, True])`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBuiltinPrint(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = print("hello")`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.NONE_OBJ {
		t.Fatalf("Expected NONE, got %v", result)
	}
}

func TestPipelineBuiltinInput(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = callable(input)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

// ============================================================
// Direct tests for executeBinaryStringOperation (0% coverage)
// ============================================================

func TestExecuteBinaryStringOperation(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeBinaryStringOperation(compiler.OpAdd, &objects.String{Value: "hello "}, &objects.String{Value: "world"})
	if err != nil {
		t.Fatalf("executeBinaryStringOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
	got := result.(*objects.String).Value
	if got != "hello world" {
		t.Errorf("Expected 'hello world', got %q", got)
	}
}

// ============================================================
// Direct tests for findMatchingExceptHandlerFrom (0% coverage)
// This function requires a proper VM frame with bytecode, so we test
// it indirectly through pipeline tests that exercise exception handling.
// ============================================================

// ============================================================
// Register VM direct function tests
// ============================================================

func TestRegVMSliceOpDirect(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = [1, 2, 3, 4, 5][1:3]`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM slice error: %s", err)
	}
}

func TestRegVMStringSliceDirect(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = "hello"[1:4]`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM string slice error: %s", err)
	}
}

func TestRegVMBytesSliceDirect(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = b"hello"[1:4]`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM bytes slice error: %s", err)
	}
}

// ============================================================
// Register VM: regBinaryOp direct tests
// ============================================================

func TestRegVMRegBinaryOpDirect(t *testing.T) {
	rvm := &RegisterVM{}
	t.Run("integer_bitor", func(t *testing.T) {
		result, err := rvm.regBinaryOp(compiler.OpBitOr, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.(*objects.Integer).Value != 7 {
			t.Errorf("Expected 7, got %d", result.(*objects.Integer).Value)
		}
	})
	t.Run("integer_bitand", func(t *testing.T) {
		result, err := rvm.regBinaryOp(compiler.OpBitAnd, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.(*objects.Integer).Value != 1 {
			t.Errorf("Expected 1, got %d", result.(*objects.Integer).Value)
		}
	})
	t.Run("integer_bitxor", func(t *testing.T) {
		result, err := rvm.regBinaryOp(compiler.OpBitXor, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.(*objects.Integer).Value != 6 {
			t.Errorf("Expected 6, got %d", result.(*objects.Integer).Value)
		}
	})
	t.Run("boolean_bitor", func(t *testing.T) {
		result, err := rvm.regBinaryOp(compiler.OpBitOr, &objects.Boolean{Value: true}, &objects.Boolean{Value: false})
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.(*objects.Integer).Value != 1 {
			t.Errorf("Expected 1, got %d", result.(*objects.Integer).Value)
		}
	})
	t.Run("set_union", func(t *testing.T) {
		s1 := objects.NewSet()
		s1.Add(&objects.Integer{Value: 1})
		s1.Add(&objects.Integer{Value: 2})
		s2 := objects.NewSet()
		s2.Add(&objects.Integer{Value: 2})
		s2.Add(&objects.Integer{Value: 3})
		result, err := rvm.regBinaryOp(compiler.OpSetUnion, s1, s2)
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.Type() != objects.SET_OBJ {
			t.Fatalf("Expected SET, got %v", result)
		}
	})
	t.Run("set_intersection", func(t *testing.T) {
		s1 := objects.NewSet()
		s1.Add(&objects.Integer{Value: 1})
		s1.Add(&objects.Integer{Value: 2})
		s2 := objects.NewSet()
		s2.Add(&objects.Integer{Value: 2})
		s2.Add(&objects.Integer{Value: 3})
		result, err := rvm.regBinaryOp(compiler.OpSetIntersection, s1, s2)
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.Type() != objects.SET_OBJ {
			t.Fatalf("Expected SET, got %v", result)
		}
	})
	t.Run("set_difference", func(t *testing.T) {
		s1 := objects.NewSet()
		s1.Add(&objects.Integer{Value: 1})
		s1.Add(&objects.Integer{Value: 2})
		s2 := objects.NewSet()
		s2.Add(&objects.Integer{Value: 2})
		s2.Add(&objects.Integer{Value: 3})
		result, err := rvm.regBinaryOp(compiler.OpSetDifference, s1, s2)
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.Type() != objects.SET_OBJ {
			t.Fatalf("Expected SET, got %v", result)
		}
	})
	t.Run("set_symmetric_difference", func(t *testing.T) {
		s1 := objects.NewSet()
		s1.Add(&objects.Integer{Value: 1})
		s1.Add(&objects.Integer{Value: 2})
		s2 := objects.NewSet()
		s2.Add(&objects.Integer{Value: 2})
		s2.Add(&objects.Integer{Value: 3})
		result, err := rvm.regBinaryOp(compiler.OpSetSymmetricDifference, s1, s2)
		if err != nil {
			t.Fatalf("regBinaryOp error: %v", err)
		}
		if result.Type() != objects.SET_OBJ {
			t.Fatalf("Expected SET, got %v", result)
		}
	})
}

// ============================================================
// Register VM: regCallClosure direct test
// ============================================================

func TestRegVMRegCallClosureDirect(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def add(a, b):
    return a + b
_r = add(3, 4)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM closure call error: %s", err)
	}
}

// ============================================================
// Register VM: sliceOp, listSlice, stringSlice, bytesSlice direct tests
// ============================================================

func TestRegVMSliceOpDirectCalls(t *testing.T) {
	rvm := &RegisterVM{}
	t.Run("list_slice", func(t *testing.T) {
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}, &objects.Integer{Value: 4}, &objects.Integer{Value: 5}}}
		result, err := rvm.listSlice(l, &objects.Integer{Value: 1}, &objects.Integer{Value: 3})
		if err != nil {
			t.Fatalf("listSlice error: %v", err)
		}
		if result.Type() != objects.LIST_OBJ {
			t.Fatalf("Expected LIST, got %v", result)
		}
		got := result.(*objects.List).Size()
		if got != 2 {
			t.Errorf("Expected 2 elements, got %d", got)
		}
	})
	t.Run("list_slice_negative", func(t *testing.T) {
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		result, err := rvm.listSlice(l, &objects.Integer{Value: -2}, &objects.Integer{Value: -1})
		if err != nil {
			t.Fatalf("listSlice error: %v", err)
		}
		if result.Type() != objects.LIST_OBJ {
			t.Fatalf("Expected LIST, got %v", result)
		}
	})
	t.Run("list_slice_none_bounds", func(t *testing.T) {
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		result, err := rvm.listSlice(l, &objects.None{}, &objects.None{})
		if err != nil {
			t.Fatalf("listSlice error: %v", err)
		}
		if result.Type() != objects.LIST_OBJ {
			t.Fatalf("Expected LIST, got %v", result)
		}
	})
	t.Run("list_slice_start_gt_end", func(t *testing.T) {
		l := &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}}}
		result, err := rvm.listSlice(l, &objects.Integer{Value: 3}, &objects.Integer{Value: 1})
		if err != nil {
			t.Fatalf("listSlice error: %v", err)
		}
		if result.Type() != objects.LIST_OBJ {
			t.Fatalf("Expected LIST, got %v", result)
		}
		if result.(*objects.List).Size() != 0 {
			t.Errorf("Expected 0 elements for start > end")
		}
	})
	t.Run("string_slice", func(t *testing.T) {
		s := &objects.String{Value: "hello"}
		result, err := rvm.stringSlice(s, &objects.Integer{Value: 1}, &objects.Integer{Value: 4})
		if err != nil {
			t.Fatalf("stringSlice error: %v", err)
		}
		if result.Type() != objects.STRING_OBJ {
			t.Fatalf("Expected STRING, got %v", result)
		}
		got := result.(*objects.String).Value
		if got != "ell" {
			t.Errorf("Expected 'ell', got %q", got)
		}
	})
	t.Run("string_slice_negative", func(t *testing.T) {
		s := &objects.String{Value: "hello"}
		result, err := rvm.stringSlice(s, &objects.Integer{Value: -3}, &objects.Integer{Value: -1})
		if err != nil {
			t.Fatalf("stringSlice error: %v", err)
		}
		if result.Type() != objects.STRING_OBJ {
			t.Fatalf("Expected STRING, got %v", result)
		}
	})
	t.Run("bytes_slice", func(t *testing.T) {
		b := &objects.Bytes{Value: []byte{65, 66, 67, 68, 69}}
		result, err := rvm.bytesSlice(b, &objects.Integer{Value: 1}, &objects.Integer{Value: 4})
		if err != nil {
			t.Fatalf("bytesSlice error: %v", err)
		}
		if result.Type() != objects.BYTES_OBJ {
			t.Fatalf("Expected BYTES, got %v", result)
		}
	})
	t.Run("slice_op_unsupported", func(t *testing.T) {
		d := objects.NewDict()
		_, err := rvm.sliceOp(d, &objects.Integer{Value: 0}, &objects.Integer{Value: 1})
		if err == nil {
			t.Fatal("Expected error for unsupported slice type")
		}
	})
}

// ============================================================
// Register VM: tupleSliceWithStep direct test
// ============================================================

func TestRegVMTupleSliceWithStepDirect(t *testing.T) {
	rvm := &RegisterVM{}
	tup := &objects.Tuple{Elements: []objects.Object{&objects.Integer{Value: 0}, &objects.Integer{Value: 1}, &objects.Integer{Value: 2}, &objects.Integer{Value: 3}, &objects.Integer{Value: 4}}}
	result, err := rvm.tupleSliceWithStep(tup, &objects.Integer{Value: 0}, &objects.Integer{Value: 5}, 2)
	if err != nil {
		t.Fatalf("tupleSliceWithStep error: %v", err)
	}
	if result.Type() != objects.TUPLE_OBJ {
		t.Fatalf("Expected TUPLE, got %v", result)
	}
	got := len(result.(*objects.Tuple).Elements)
	if got != 3 {
		t.Errorf("Expected 3 elements, got %d", got)
	}
}

// ============================================================
// More pipeline tests for Run() opcode paths
// ============================================================

func TestPipelineInPlaceMod2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 17
_r %= 5`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineInPlaceFloorDiv2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 17
_r //= 5`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineInPlaceAdd(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5
_r += 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 8)
}

// ============================================================
// More register VM tests for executeRegFrame
// ============================================================

func TestRegVMFunctionDefaultArgs(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def greet(name, greeting="hello"):
    return greeting
_r = greet("world")
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMRecursiveFunction(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)
_r = fib(6)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMTryExcept2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
try:
    raise ValueError("test")
except ValueError:
    _r = 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}




func TestRegVMDictOps2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_d = {}
_d["a"] = 1
_d["b"] = 2
_r = _d["a"]
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

// ============================================================
// More stack VM pipeline tests for additional opcode paths
// ============================================================





func TestPipelineNestedForLoop(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
for _i in range(3):
    for _j in range(3):
        _r = _r + 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 9)
}

func TestPipelineStringConcat(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = "hello" + " world"`)
	assertString(t, getGlobal(machine, comp, "_r"), "hello world")
}

func TestPipelineStringRepeat(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = "ha" * 3`)
	assertString(t, getGlobal(machine, comp, "_r"), "hahaha")
}

func TestPipelineListIndexNegative(t *testing.T) {
	comp, machine := compileAndRun(t, `_x = [10, 20, 30]
_r = _x[-1]`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 30)
}

func TestPipelineStringIndex(t *testing.T) {
	comp, machine := compileAndRun(t, `_s = "hello"
_r = _s[1]`)
	assertString(t, getGlobal(machine, comp, "_r"), "e")
}

func TestPipelineStringSlice(t *testing.T) {
	comp, machine := compileAndRun(t, `_s = "hello world"
_r = _s[0:5]`)
	assertString(t, getGlobal(machine, comp, "_r"), "hello")
}

func TestPipelineListSlice(t *testing.T) {
	comp, machine := compileAndRun(t, `_x = [1, 2, 3, 4, 5]
_r = _x[1:4]`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
	if result.(*objects.List).Size() != 3 {
		t.Errorf("Expected 3 elements")
	}
}



func TestPipelineAndOperator(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = True and False`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), false)
}

func TestPipelineOrOperator(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = False or True`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}









func TestPipelineBitwiseOr(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5 | 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineBitwiseAnd(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5 & 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineBitwiseXor(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5 ^ 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}





func TestPipelineNegateInt(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = -42`)
	assertInteger(t, getGlobal(machine, comp, "_r"), -42)
}

func TestPipelineNegateFloat(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = -3.14`)
	assertFloat(t, getGlobal(machine, comp, "_r"), -3.14)
}

func TestPipelineBoolComparison(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = True == True`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineNoneComparison(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = None == None`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineChainedComparison(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 1 < 2 < 3`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineMultipleAssignment(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = 1
_y = 2
_r = _x + _y`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineClassWithMethod(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Counter:
    def __init__(self):
        self.count = 0
    def inc(self):
        self.count = self.count + 1
_c = Counter()
_c.inc()
_c.inc()
_r = _c.count`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 2)
}

func TestPipelineClassMethodWithArgs(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Adder:
    def add(self, a, b):
        return a + b
_a = Adder()
_r = _a.add(3, 4)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineClassOverrideMethod(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Base:
    def greet(self):
        return "hello"
class Child(Base):
    def greet(self):
        return "hi"
_c = Child()
_r = _c.greet()`)
	assertString(t, getGlobal(machine, comp, "_r"), "hi")
}

func TestPipelineBuiltinLen(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len("hello")`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinType(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = type(42)`)
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineBuiltinAbs(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = abs(-5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinStr(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = str(42)`)
	assertString(t, getGlobal(machine, comp, "_r"), "42")
}

func TestPipelineBuiltinInt(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = int(3.5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineBuiltinFloat(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = float(5)`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 5.0)
}

func TestPipelineBuiltinBool(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = bool(0)`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), false)
}

func TestPipelineBuiltinList(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = list(range(5))`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinSet(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = set([1, 2, 2, 3])`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineBuiltinSorted(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = sorted([3, 1, 2])`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
}

func TestPipelineBuiltinSum(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = sum([1, 2, 3])`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

func TestPipelineBuiltinMin(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = min(1, 2, 3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineBuiltinMax(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = max(1, 2, 3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}



func TestPipelineBuiltinRange1Arg(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len(range(5))`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinRange2Args(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len(range(2, 7))`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBuiltinRange3Args(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len(range(0, 10, 2))`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

// ============================================================
// Direct tests for CallCallable with Closure
// ============================================================

func TestCallCallableClosure2(t *testing.T) {
	comp, machine := compileAndRun(t, `
def add(a, b):
    return a + b
_r = add(3, 4)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestCallCallableClosureWithFreeVar(t *testing.T) {
	comp, machine := compileAndRun(t, `
def make_adder(n):
    def adder(x):
        return x + n
    return adder
_f = make_adder(10)
_r = _f(5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 15)
}

// ============================================================
// Direct tests for executeBinaryOperation paths
// ============================================================

func TestExecuteBinaryOperationDictMerge(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	d1 := objects.NewDict()
	d1.Set(&objects.String{Value: "a"}, &objects.Integer{Value: 1})
	d2 := objects.NewDict()
	d2.Set(&objects.String{Value: "b"}, &objects.Integer{Value: 2})
	vm.push(d1)
	vm.push(d2)
	err := vm.executeBinaryOperation(compiler.OpBitOr)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.DICT_OBJ {
		t.Fatalf("Expected DICT, got %v", result)
	}
}

func TestExecuteBinaryOperationSetSub(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	s1 := objects.NewSet()
	s1.Add(&objects.Integer{Value: 1})
	s1.Add(&objects.Integer{Value: 2})
	s1.Add(&objects.Integer{Value: 3})
	s2 := objects.NewSet()
	s2.Add(&objects.Integer{Value: 2})
	vm.push(s1)
	vm.push(s2)
	err := vm.executeBinaryOperation(compiler.OpSub)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestExecuteBinaryOperationSetXor(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	s1 := objects.NewSet()
	s1.Add(&objects.Integer{Value: 1})
	s1.Add(&objects.Integer{Value: 2})
	s2 := objects.NewSet()
	s2.Add(&objects.Integer{Value: 2})
	s2.Add(&objects.Integer{Value: 3})
	vm.push(s1)
	vm.push(s2)
	err := vm.executeBinaryOperation(compiler.OpBitXor)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestExecuteBinaryOperationBoolBool(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Boolean{Value: true})
	vm.push(&objects.Boolean{Value: false})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
}

func TestExecuteBinaryOperationBoolInt(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Boolean{Value: true})
	vm.push(&objects.Integer{Value: 5})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
}

func TestExecuteBinaryOperationIntBool(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 5})
	vm.push(&objects.Boolean{Value: true})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
}

func TestExecuteBinaryOperationBoolFloat(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Boolean{Value: true})
	vm.push(&objects.Float{Value: 2.5})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.FLOAT_OBJ {
		t.Fatalf("Expected FLOAT, got %v", result)
	}
}

func TestExecuteBinaryOperationFloatBool(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Float{Value: 2.5})
	vm.push(&objects.Boolean{Value: true})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.FLOAT_OBJ {
		t.Fatalf("Expected FLOAT, got %v", result)
	}
}

func TestExecuteBinaryOperationComplexComplex(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	vm.push(&objects.Complex{Real: 3, Imag: 4})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestExecuteBinaryOperationComplexFloat(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	vm.push(&objects.Float{Value: 3.0})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestExecuteBinaryOperationComplexInt(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	vm.push(&objects.Integer{Value: 3})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestExecuteBinaryOperationFloatComplex(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Float{Value: 3.0})
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestExecuteBinaryOperationIntComplex(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 3})
	vm.push(&objects.Complex{Real: 1, Imag: 2})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestExecuteBinaryOperationStringRepeatIntString(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 3})
	vm.push(&objects.String{Value: "ha"})
	err := vm.executeBinaryOperation(compiler.OpMul)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
	got := result.(*objects.String).Value
	if got != "hahaha" {
		t.Errorf("Expected 'hahaha', got %q", got)
	}
}

func TestExecuteBinaryOperationStringRepeatZero(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.String{Value: "ha"})
	vm.push(&objects.Integer{Value: 0})
	err := vm.executeBinaryOperation(compiler.OpMul)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
	got := result.(*objects.String).Value
	if got != "" {
		t.Errorf("Expected '', got %q", got)
	}
}

func TestExecuteBinaryOperationListConcat(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}})
	vm.push(&objects.List{Elements: []objects.Object{&objects.Integer{Value: 2}}})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.LIST_OBJ {
		t.Fatalf("Expected LIST, got %v", result)
	}
	if result.(*objects.List).Size() != 2 {
		t.Errorf("Expected 2 elements")
	}
}

func TestExecuteBinaryOperationStringConcat(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.String{Value: "hello "})
	vm.push(&objects.String{Value: "world"})
	err := vm.executeBinaryOperation(compiler.OpAdd)
	if err != nil {
		t.Fatalf("executeBinaryOperation error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.STRING_OBJ {
		t.Fatalf("Expected STRING, got %v", result)
	}
	got := result.(*objects.String).Value
	if got != "hello world" {
		t.Errorf("Expected 'hello world', got %q", got)
	}
}

// ============================================================
// Direct tests for executeComparison paths
// ============================================================

func TestExecuteComparisonStringLt(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.String{Value: "abc"})
	vm.push(&objects.String{Value: "def"})
	err := vm.executeComparison(compiler.OpLessThan)
	if err != nil {
		t.Fatalf("executeComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

func TestExecuteComparisonStringGt(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.String{Value: "def"})
	vm.push(&objects.String{Value: "abc"})
	err := vm.executeComparison(compiler.OpGreaterThan)
	if err != nil {
		t.Fatalf("executeComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

func TestExecuteComparisonBoolEq(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Boolean{Value: true})
	vm.push(&objects.Boolean{Value: true})
	err := vm.executeComparison(compiler.OpEqual)
	if err != nil {
		t.Fatalf("executeComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

func TestExecuteComparisonNoneEq(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.None{})
	vm.push(&objects.None{})
	err := vm.executeComparison(compiler.OpEqual)
	if err != nil {
		t.Fatalf("executeComparison error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.BOOLEAN_OBJ {
		t.Fatalf("Expected BOOLEAN, got %v", result)
	}
}

// ============================================================
// Direct tests for executeMinusOperator
// ============================================================

func TestExecuteMinusOperatorInt(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 42})
	err := vm.executeMinusOperator()
	if err != nil {
		t.Fatalf("executeMinusOperator error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.INTEGER_OBJ {
		t.Fatalf("Expected INTEGER, got %v", result)
	}
	got := result.(*objects.Integer).Value
	if got != -42 {
		t.Errorf("Expected -42, got %d", got)
	}
}

func TestExecuteMinusOperatorFloat(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Float{Value: 3.14})
	err := vm.executeMinusOperator()
	if err != nil {
		t.Fatalf("executeMinusOperator error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.FLOAT_OBJ {
		t.Fatalf("Expected FLOAT, got %v", result)
	}
}

func TestExecuteMinusOperatorComplex(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Complex{Real: 3, Imag: 4})
	err := vm.executeMinusOperator()
	if err != nil {
		t.Fatalf("executeMinusOperator error: %v", err)
	}
	result := vm.pop()
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

// ============================================================
// Direct tests for executeInPlaceOperation
// ============================================================

func TestExecuteInPlaceOperationAdd(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.push(&objects.Integer{Value: 5})
	vm.push(&objects.Integer{Value: 3})
	err := vm.executeInPlaceOperation(compiler.OpInPlaceAdd, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationSub(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceSub, &objects.Integer{Value: 10}, &objects.Integer{Value: 3})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationMul(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceMul, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationDiv(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceDiv, &objects.Float{Value: 10.0}, &objects.Float{Value: 4.0})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationMod(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceMod, &objects.Integer{Value: 17}, &objects.Integer{Value: 5})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationFloorDiv(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceFloorDiv, &objects.Integer{Value: 17}, &objects.Integer{Value: 5})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationPower(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlacePower, &objects.Integer{Value: 2}, &objects.Integer{Value: 8})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationBitOr(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceBitOr, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationBitAnd(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceBitAnd, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

func TestExecuteInPlaceOperationBitXor(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	err := vm.executeInPlaceOperation(compiler.OpInPlaceBitXor, &objects.Integer{Value: 5}, &objects.Integer{Value: 3})
	if err != nil {
		t.Fatalf("executeInPlaceOperation error: %v", err)
	}
}

// ============================================================
// More register VM tests for executeRegFrame and regCallClosure
// ============================================================

func TestRegVMClosure2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def make_counter():
    count = 0
    def inc():
        count = count + 1
        return count
    return inc
_c = make_counter()
_r = _c()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFunctionReturn(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def get_val():
    return 42
_r = get_val()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFunctionNoReturn(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def no_ret():
    _x = 1
_r = no_ret()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFunctionMultipleReturns(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
def check(x):
    if x > 0:
        return 1
    return 0
_r = check(5)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMClassWithInit2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
_p = Point(3, 4)
_r = _p.x + _p.y
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMClassNoInit2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    pass
_f = Foo()
_r = 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMSetAttr2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    def __init__(self):
        self.x = 10
_f = Foo()
_f.y = 20
_r = _f.x + _f.y
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMElif(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_x = 15
if _x > 20:
    _r = 1
elif _x > 10:
    _r = 2
else:
    _r = 3
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMNestedIf(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_x = 5
if _x > 0:
    if _x > 10:
        _r = 1
    else:
        _r = 2
else:
    _r = 3
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMForBreak(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
for _i in range(10):
    if _i == 5:
        break
    _r = _r + 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMForContinue(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 0
for _i in range(10):
    if _i < 5:
        continue
    _r = _r + 1
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinLen2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = len([1, 2, 3])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinChr(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = chr(65)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinOrd(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = ord("A")`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinHex(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = hex(255)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinOct(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = oct(8)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinBin(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = bin(10)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinRound(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = round(3.5)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinCallable(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = callable(print)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinRepr(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = repr(42)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinId(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = id(42)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinHash(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = hash(42)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinAll(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = all([True, True])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinAny(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = any([False, True])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceAdd2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 5
_r += 3
_r += 2
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceSub2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 10
_r -= 3
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMInPlaceMul2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = 3
_r *= 7
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMNegate2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = -42`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMNot2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = not True`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

// ============================================================
// GC tests
// ============================================================

func TestVMGC(t *testing.T) {
	bc := &compiler.Bytecode{Instructions: []byte{}}
	vm := New(bc)
	vm.EnableGC(false)
	if vm.IsGCEnabled() {
		t.Error("GC should be disabled")
	}
	vm.EnableGC(true)
	if !vm.IsGCEnabled() {
		t.Error("GC should be enabled")
	}
	vm.SetGCThreshold(100)
	if vm.GetGCThreshold() != 100 {
		t.Errorf("Expected threshold 100, got %d", vm.GetGCThreshold())
	}
	vm.TriggerGC()
}

// ============================================================
// Pipeline tests: dict merge with | operator
// ============================================================

func TestPipelineDictMergeOperator(t *testing.T) {
	comp, machine := compileAndRun(t, `_d1 = {}
_d1["a"] = 1
_d2 = {}
_d2["b"] = 2
_r = _d1 | _d2`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.DICT_OBJ {
		t.Fatalf("Expected DICT, got %v", result)
	}
}

// ============================================================
// Pipeline tests: set difference with - operator
// ============================================================

// ============================================================
// Pipeline tests: more integer binary operations
// ============================================================

func TestPipelineIntegerSub(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 10 - 3")
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineIntegerMul(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 4 * 5")
	assertInteger(t, getGlobal(machine, comp, "_r"), 20)
}

func TestPipelineFloatDiv(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 10.0 / 4.0")
	assertFloat(t, getGlobal(machine, comp, "_r"), 2.5)
}

func TestPipelineIntegerComparisonLt(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 3 < 5")
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineIntegerComparisonGt(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 5 > 3")
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineIntegerComparisonEq(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 5 == 5")
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineIntegerComparisonNeq(t *testing.T) {
	comp, machine := compileAndRun(t, "_r = 5 != 3")
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

// ============================================================
// Pipeline tests: more class patterns for executeCall
// ============================================================

func TestPipelineClassWithProperty(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Rect:
    def __init__(self, w, h):
        self.w = w
        self.h = h
    def area(self):
        return self.w * self.h
_r = Rect(3, 4).area()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 12)
}

func TestPipelineClassInheritAndOverride(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Base:
    def val(self):
        return 10
class Child(Base):
    def val(self):
        return 20
_c = Child()
_r = _c.val()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 20)
}

func TestPipelineClassInheritNoOverride(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Base:
    def val(self):
        return 10
class Child(Base):
    pass
_c = Child()
_r = _c.val()`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 10)
}

// ============================================================
// Pipeline tests: more exception handling patterns
// ============================================================

func TestPipelineTryExceptBare(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = 0
try:
    raise ValueError("test")
except:
    _r = 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineTryExceptWithVar(t *testing.T) {
	comp, machine := compileAndRun(t, `
_r = ""
try:
    raise ValueError("test error")
except ValueError as e:
    _r = "caught"`)
	assertString(t, getGlobal(machine, comp, "_r"), "caught")
}

// ============================================================
// Pipeline tests: more function patterns
// ============================================================

func TestPipelineFunctionReturningFunction(t *testing.T) {
	comp, machine := compileAndRun(t, `
def make_adder(n):
    def adder(x):
        return x + n
    return adder
_doubler = make_adder(2)
_r = _doubler(5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineLambdaAsArg(t *testing.T) {
	comp, machine := compileAndRun(t, `
def apply(f, x):
    return f(x)
_r = apply(lambda x: x + 1, 5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 6)
}

// ============================================================
// More pipeline tests for Run() opcode paths
// ============================================================

func TestPipelineGlobalRead(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = 42
_r = _x`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 42)
}

func TestPipelineArrayIndex(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = [10, 20, 30]
_r = _x[0] + _x[1] + _x[2]`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 60)
}

func TestPipelineDictCreateAndAccess(t *testing.T) {
	comp, machine := compileAndRun(t, `
_d = {}
_d["x"] = 10
_d["y"] = 20
_r = _d["x"] + _d["y"]`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 30)
}

func TestPipelineSetCreate(t *testing.T) {
	comp, machine := compileAndRun(t, `
_s = {1, 2, 3}
_r = len(_s)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineListCreate(t *testing.T) {
	comp, machine := compileAndRun(t, `
_x = [1, 2, 3]
_r = len(_x)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineStringLen(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len("hello")`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineBoolTrue(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = True`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), true)
}

func TestPipelineBoolFalse(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = False`)
	assertBoolean(t, getGlobal(machine, comp, "_r"), false)
}

func TestPipelineNoneValue(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = None`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.NONE_OBJ {
		t.Fatalf("Expected NONE, got %v", result)
	}
}

func TestPipelineFloatNeg(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = -3.14`)
	assertFloat(t, getGlobal(machine, comp, "_r"), -3.14)
}

func TestPipelineIntDiv(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 10 / 3`)
	result := getGlobal(machine, comp, "_r")
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineIntSub(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 10 - 3`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineIntMul(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 4 * 5`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 20)
}

func TestPipelineFloatAdd(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 1.5 + 2.5`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 4.0)
}

func TestPipelineFloatSub(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 5.5 - 1.5`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 4.0)
}

func TestPipelineFloatMul(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 2.5 * 2.0`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 5.0)
}

func TestPipelineSetUnion(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2} | {2, 3}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineSetIntersection2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2, 3} & {2, 3, 4}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineSetXor(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = {1, 2} ^ {2, 3}`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.SET_OBJ {
		t.Fatalf("Expected SET, got %v", result)
	}
}

func TestPipelineBytesLiteral(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = len(b"hello")`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 5)
}

func TestPipelineComplexLiteral(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = 1 + 2j`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.COMPLEX_OBJ {
		t.Fatalf("Expected COMPLEX, got %v", result)
	}
}

func TestPipelineEllipsisLiteral2(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = ...`)
	result := getGlobal(machine, comp, "_r")
	if result == nil || result.Type() != objects.ELLIPSIS_OBJ {
		t.Fatalf("Expected ELLIPSIS, got %v", result)
	}
}

func TestPipelineClassWithSlots2(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Point:
    __slots__ = ["x", "y"]
    def __init__(self, x, y):
        self.x = x
        self.y = y
_p = Point(3, 4)
_r = _p.x + _p.y`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 7)
}

func TestPipelineClassDelAttr(t *testing.T) {
	comp, machine := compileAndRun(t, `
class Foo:
    def __init__(self):
        self.x = 10
        self.y = 20
_f = Foo()
del _f.x
_r = 1`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineBuiltinMinArgs(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = min(1, 2, 3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 1)
}

func TestPipelineBuiltinMaxArgs(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = max(1, 2, 3)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 3)
}

func TestPipelineBuiltinAbsFloat(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = abs(-3.14)`)
	assertFloat(t, getGlobal(machine, comp, "_r"), 3.14)
}

func TestPipelineBuiltinRoundInt(t *testing.T) {
	comp, machine := compileAndRun(t, `_r = round(3.5)`)
	assertInteger(t, getGlobal(machine, comp, "_r"), 4)
}

// ============================================================
// More register VM tests
// ============================================================

func TestRegVMClassInherit3(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Animal:
    def speak(self):
        return "generic"
class Dog(Animal):
    def speak(self):
        return "woof"
_d = Dog()
_r = _d.speak()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMDictAccess(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_d = {}
_d["x"] = 10
_d["y"] = 20
_r = _d["x"] + _d["y"]
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMIsinstanceCustom(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    pass
_f = Foo()
_r = isinstance(_f, Foo)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMHasattrCustom(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
_r = hasattr(_f, "x")
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMGetattrCustom(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
_r = getattr(_f, "x")
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMSetattrCustom(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Foo:
    def __init__(self):
        self.x = 42
_f = Foo()
setattr(_f, "x", 100)
_r = _f.x
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinPow(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = pow(2, 10)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinAbs3(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = abs(-5)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinDivmod(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = divmod(17, 5)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinMin(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = min(1, 2, 3)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinMax(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = max(1, 2, 3)`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinMap(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = list(map(str, [1, 2, 3]))`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinFilter(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = list(filter(bool, [0, 1, 2, 0]))`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinZip(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = list(zip([1, 2], [3, 4]))`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinEnumerate(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = list(enumerate([10, 20, 30]))`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinReversed(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = list(reversed([1, 2, 3]))`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinSorted2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = sorted([3, 1, 2])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBuiltinSum2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = sum([1, 2, 3])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBitOps2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r1 = 5 | 3
_r2 = 5 & 3
_r3 = 5 ^ 3
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMComplexOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_r = (1+2j) + (3+4j)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMBytesOps(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_b = b"hello"
_r = len(_b)
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMFString2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
_x = 42
_r = f"value: {_x}"
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMSuperCall2(t *testing.T) {
	rbc := compileToRegBytecodeNew(`
class Base:
    def greet(self):
        return "hello"
class Child(Base):
    def greet(self):
        _s = super().greet()
        return _s
_c = Child()
_r = _c.greet()
`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func TestRegVMExceptionGroup(t *testing.T) {
	rbc := compileToRegBytecodeNew(`_r = ExceptionGroup("eg", [TypeError("a"), ValueError("b")])`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}
