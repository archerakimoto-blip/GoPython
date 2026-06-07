package vm

import (
	"testing"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
)

func compileToStackBytecode(src string) *compiler.Bytecode {
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	program = desugar.Desugar(program)
	comp := compiler.New()
	comp.Compile(program)
	return comp.Bytecode()
}

func compileToRegBytecode(src string) *compiler.RegBytecode {
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	program = desugar.Desugar(program)
	regComp := compiler.NewRegisterCompiler()
	regComp.Compile(program)
	return regComp.Bytecode()
}

func TestRegisterVM_BasicFunctionCall(t *testing.T) {
	rbc := compileToRegBytecode(`def add(a, b): return a + b; print(add(3, 4))`)
	globals := make([]objects.Object, GlobalSize)
	rvm := NewRegisterVMWithBytecode(rbc, globals)
	err := rvm.RunRegDirect(rbc)
	if err != nil {
		t.Fatalf("Register VM execution error: %s", err)
	}
}

func BenchmarkStackVM_Arithmetic(b *testing.B) {
	bc := compileToStackBytecode(`print(2 + 3 * 4 - 1)`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewWithGlobalsStore(bc, globals)
		vm.Run()
	}
}

func BenchmarkRegisterVM_Arithmetic(b *testing.B) {
	rbc := compileToRegBytecode(`print(2 + 3 * 4 - 1)`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rvm := NewRegisterVMWithBytecode(rbc, globals)
		rvm.RunRegDirect(rbc)
	}
}

func BenchmarkStackVM_FunctionCall(b *testing.B) {
	bc := compileToStackBytecode(`def add(a, b): return a + b; print(add(3, 4))`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewWithGlobalsStore(bc, globals)
		vm.Run()
	}
}

func BenchmarkRegisterVM_FunctionCall(b *testing.B) {
	rbc := compileToRegBytecode(`def add(a, b): return a + b; print(add(3, 4))`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rvm := NewRegisterVMWithBytecode(rbc, globals)
		rvm.RunRegDirect(rbc)
	}
}

func BenchmarkStackVM_Loop(b *testing.B) {
	bc := compileToStackBytecode(`x = 0
while x < 100:
	x = x + 1
print(x)`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewWithGlobalsStore(bc, globals)
		vm.Run()
	}
}

func BenchmarkRegisterVM_Loop(b *testing.B) {
	rbc := compileToRegBytecode(`x = 0
while x < 100:
	x = x + 1
print(x)`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rvm := NewRegisterVMWithBytecode(rbc, globals)
		rvm.RunRegDirect(rbc)
	}
}

func BenchmarkStackVM_StringConcat(b *testing.B) {
	bc := compileToStackBytecode(`s = "hello"; print(s + " world")`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewWithGlobalsStore(bc, globals)
		vm.Run()
	}
}

func BenchmarkRegisterVM_StringConcat(b *testing.B) {
	rbc := compileToRegBytecode(`s = "hello"; print(s + " world")`)
	globals := make([]objects.Object, GlobalSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rvm := NewRegisterVMWithBytecode(rbc, globals)
		rvm.RunRegDirect(rbc)
	}
}
