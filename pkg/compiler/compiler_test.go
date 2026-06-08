package compiler

import (
	"testing"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/objects"
)

// =============================================================================
// Helper functions
// =============================================================================

func newTestCompiler() *Compiler {
	return New()
}

func newTestRegCompiler() *RegisterCompiler {
	return NewRegisterCompiler()
}

func compileProgram(stmts ...ast.Statement) (*Bytecode, error) {
	c := newTestCompiler()
	prog := &ast.Program{Statements: stmts}
	err := c.Compile(prog)
	if err != nil {
		return nil, err
	}
	return c.Bytecode(), nil
}

// hasOpcode checks if the given opcode appears in the instructions.
func hasOpcode(ins []byte, op Opcode) bool {
	for _, b := range ins {
		if Opcode(b) == op {
			return true
		}
	}
	return false
}

// hasRegOpcode checks if the given register opcode appears in the instructions.
func hasRegOpcode(instrs []RegInstruction, op RegOpcode) bool {
	for _, instr := range instrs {
		if instr.Opcode == op {
			return true
		}
	}
	return false
}

// =============================================================================
// 1. Compiler construction tests
// =============================================================================

func TestNew(t *testing.T) {
	c := newTestCompiler()
	if c == nil {
		t.Fatal("expected non-nil compiler")
	}
	if c.symbolTable == nil {
		t.Fatal("expected non-nil symbol table")
	}
	if len(c.constants) == 0 {
		t.Fatal("expected builtins to register constants")
	}
}

func TestNewWithState(t *testing.T) {
	st := NewSymbolTable()
	constants := []objects.Object{&objects.Integer{Value: 42}}
	c := NewWithState(st, constants)
	if c == nil {
		t.Fatal("expected non-nil compiler")
	}
	if c.symbolTable != st {
		t.Error("expected symbol table to be the provided one")
	}
	if len(c.constants) != 1 {
		t.Errorf("expected 1 constant, got %d", len(c.constants))
	}
}

func TestNewRegisterCompiler(t *testing.T) {
	rc := newTestRegCompiler()
	if rc == nil {
		t.Fatal("expected non-nil register compiler")
	}
	if rc.symbolTable == nil {
		t.Fatal("expected non-nil symbol table")
	}
	if len(rc.constants) == 0 {
		t.Fatal("expected builtins to register constants")
	}
}

func TestNewRegisterCompilerWithState(t *testing.T) {
	st := NewSymbolTable()
	constants := []objects.Object{&objects.Integer{Value: 42}}
	rc := NewRegisterCompilerWithState(st, constants)
	if rc == nil {
		t.Fatal("expected non-nil register compiler")
	}

	rc2 := NewRegisterCompilerWithState(nil, nil)
	if rc2 == nil {
		t.Fatal("expected non-nil register compiler with nil state")
	}
	if rc2.symbolTable == nil {
		t.Fatal("expected symbol table to be created")
	}
	if rc2.constants == nil {
		t.Fatal("expected constants to be created")
	}
}

// =============================================================================
// 2. Symbol Table tests
// =============================================================================

func TestSymbolTableDefine(t *testing.T) {
	st := NewSymbolTable()
	sym := st.Define("x")
	if sym.Name != "x" {
		t.Errorf("expected name 'x', got %q", sym.Name)
	}
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope, got %v", sym.Scope)
	}
	if sym.Index != 0 {
		t.Errorf("expected index 0, got %d", sym.Index)
	}
}

func TestSymbolTableDefineLocal(t *testing.T) {
	outer := NewSymbolTable()
	inner := NewEnclosedSymbolTable(outer)
	sym := inner.Define("y")
	if sym.Scope != LocalScope {
		t.Errorf("expected LocalScope, got %v", sym.Scope)
	}
}

func TestSymbolTableResolve(t *testing.T) {
	st := NewSymbolTable()
	st.Define("x")
	sym, ok := st.Resolve("x")
	if !ok {
		t.Fatal("expected to resolve 'x'")
	}
	if sym.Name != "x" {
		t.Errorf("expected name 'x', got %q", sym.Name)
	}
}

func TestSymbolTableResolveUndefined(t *testing.T) {
	st := NewSymbolTable()
	_, ok := st.Resolve("undefined")
	if ok {
		t.Error("expected undefined variable to not resolve")
	}
}

func TestSymbolTableDefineBuiltin(t *testing.T) {
	st := NewSymbolTable()
	st.DefineBuiltin("len", 0)
	sym, ok := st.Resolve("len")
	if !ok {
		t.Fatal("expected to resolve builtin 'len'")
	}
	if sym.Scope != BuiltinScope {
		t.Errorf("expected BuiltinScope, got %v", sym.Scope)
	}
}

func TestSymbolTableDefineFunctionName(t *testing.T) {
	st := NewSymbolTable()
	sym := st.DefineFunctionName("myFunc")
	if sym.Scope != FunctionScope {
		t.Errorf("expected FunctionScope, got %v", sym.Scope)
	}
	if sym.Name != "myFunc" {
		t.Errorf("expected name 'myFunc', got %q", sym.Name)
	}
}

func TestSymbolTableDefineGlobal(t *testing.T) {
	outer := NewSymbolTable()
	outer.Define("g") // Define g in global scope first
	inner := NewEnclosedSymbolTable(outer)
	inner.DefineGlobal("g")
	// DefineGlobal just marks the name; Define actually creates the symbol
	sym := inner.Define("g")
	// In an inner scope with global declaration, the symbol should resolve to GlobalScope
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope for global declaration, got %q", sym.Scope)
	}
}

func TestSymbolTableDefineNonlocal(t *testing.T) {
	// Define x as a local in outer (by using enclosed table)
	outerInner := NewEnclosedSymbolTable(NewSymbolTable())
	outerInner.Define("x") // This creates x as LocalScope in outerInner
	inner := NewEnclosedSymbolTable(outerInner)
	inner.DefineNonlocal("x")
	sym := inner.Define("x")
	if sym.Scope != FreeScope {
		t.Errorf("expected FreeScope for nonlocal declaration, got %v", sym.Scope)
	}
}

func TestSymbolTableIsGlobal(t *testing.T) {
	st := NewSymbolTable()
	st.DefineGlobal("g")
	if !st.IsGlobal("g") {
		t.Error("expected 'g' to be global")
	}
	if st.IsGlobal("x") {
		t.Error("expected 'x' to not be global")
	}
}

func TestSymbolTableIsNonlocal(t *testing.T) {
	st := NewSymbolTable()
	st.DefineNonlocal("n")
	if !st.IsNonlocal("n") {
		t.Error("expected 'n' to be nonlocal")
	}
	if st.IsNonlocal("x") {
		t.Error("expected 'x' to not be nonlocal")
	}
}

func TestSymbolTableFreeVariables(t *testing.T) {
	// Create a chain: global -> outer (local x) -> inner
	global := NewSymbolTable()
	outer := NewEnclosedSymbolTable(global)
	outer.Define("x") // x is Local in outer
	inner := NewEnclosedSymbolTable(outer)
	sym, ok := inner.Resolve("x")
	if !ok {
		t.Fatal("expected to resolve 'x' in inner scope")
	}
	if sym.Scope != FreeScope {
		t.Errorf("expected FreeScope for outer local variable, got %v", sym.Scope)
	}
}

func TestSymbolTableNumDefinitionsInScope(t *testing.T) {
	st := NewSymbolTable()
	st.Define("a")
	st.Define("b")
	count := st.numDefinitionsInScope()
	if count != 2 {
		t.Errorf("expected 2 definitions, got %d", count)
	}
}

func TestSymbolTableNestedFreeSymbols(t *testing.T) {
	global := NewSymbolTable()
	outer := NewEnclosedSymbolTable(global)
	outer.Define("x") // x is Local in outer
	inner := NewEnclosedSymbolTable(outer)
	inner.Resolve("x")
	if len(outer.NestedFreeSymbols) != 1 {
		t.Errorf("expected 1 nested free symbol, got %d", len(outer.NestedFreeSymbols))
	}
}

func TestSymbolTableResolveGlobalFromInner(t *testing.T) {
	global := NewSymbolTable()
	global.Define("g") // g is Global
	inner := NewEnclosedSymbolTable(global)
	sym, ok := inner.Resolve("g")
	if !ok {
		t.Fatal("expected to resolve 'g'")
	}
	// Global variables are just returned as-is (not FreeScope)
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope, got %v", sym.Scope)
	}
}

// =============================================================================
// 3. Stack-based Compiler: Expression tests
// =============================================================================

func TestCompileIntegerLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.IntegerLiteral{Value: 42},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpConstant) {
		t.Error("expected OpConstant in instructions")
	}
}

func TestCompileFloatLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.FloatLiteral{Value: 3.14},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpConstant) {
		t.Error("expected OpConstant in instructions")
	}
}

func TestCompileComplexLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.ComplexLiteral{Value: "2.5j"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpConstant) {
		t.Error("expected OpConstant in instructions")
	}
}

func TestCompileComplexLiteralInvalid(t *testing.T) {
	_, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.ComplexLiteral{Value: "notanumberj"},
	})
	if err == nil {
		t.Fatal("expected error for invalid complex literal")
	}
}

func TestCompileStringLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.StringLiteral{Value: "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpConstant) {
		t.Error("expected OpConstant in instructions")
	}
}

func TestCompileByteStringLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.ByteStringLiteral{Value: "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpConstant) {
		t.Error("expected OpConstant in instructions")
	}
}

func TestCompileBooleanTrue(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.Boolean{Value: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpTrue) {
		t.Error("expected OpTrue in instructions")
	}
}

func TestCompileBooleanFalse(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.Boolean{Value: false},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpFalse) {
		t.Error("expected OpFalse in instructions")
	}
}

func TestCompileEllipsisLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.EllipsisLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpEllipsis) {
		t.Error("expected OpEllipsis in instructions")
	}
}

func TestCompilePrefixExpression(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected Opcode
	}{
		{"negate", "-", OpMinus},
		{"not", "!", OpBang},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc, err := compileProgram(&ast.ExpressionStatement{
				Expression: &ast.PrefixExpression{
					Operator: tt.operator, Right: &ast.IntegerLiteral{Value: 1},
				},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !hasOpcode(bc.Instructions, tt.expected) {
				t.Errorf("expected opcode %d in instructions", tt.expected)
			}
		})
	}
}

func TestCompileInfixExpression(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected Opcode
	}{
		{"add", "+", OpAdd}, {"sub", "-", OpSub}, {"mul", "*", OpMul},
		{"div", "/", OpDiv}, {"mod", "%", OpMod}, {"floordiv", "//", OpFloorDiv},
		{"power", "**", OpPower}, {"gt", ">", OpGreaterThan}, {"eq", "==", OpEqual},
		{"ne", "!=", OpNotEqual}, {"bitor", "|", OpBitOr},
		{"bitand", "&", OpBitAnd}, {"bitxor", "^", OpBitXor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc, err := compileProgram(&ast.ExpressionStatement{
				Expression: &ast.InfixExpression{
					Left: &ast.IntegerLiteral{Value: 1}, Operator: tt.operator, Right: &ast.IntegerLiteral{Value: 2},
				},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !hasOpcode(bc.Instructions, tt.expected) {
				t.Errorf("expected opcode %d in instructions", tt.expected)
			}
		})
	}
}

func TestCompileLessThanInfix(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.InfixExpression{
			Left: &ast.IntegerLiteral{Value: 1}, Operator: "<", Right: &ast.IntegerLiteral{Value: 2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpGreaterThan) {
		t.Error("expected OpGreaterThan for < operator (swapped operands)")
	}
}

func TestCompileAndOrError(t *testing.T) {
	for _, op := range []string{"and", "or"} {
		_, err := compileProgram(&ast.ExpressionStatement{
			Expression: &ast.InfixExpression{
				Left: &ast.IntegerLiteral{Value: 1}, Operator: op, Right: &ast.IntegerLiteral{Value: 2},
			},
		})
		if err == nil {
			t.Fatalf("expected error for %s operator", op)
		}
	}
}

func TestCompileUnknownOperatorError(t *testing.T) {
	_, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.InfixExpression{
			Left: &ast.IntegerLiteral{Value: 1}, Operator: "???", Right: &ast.IntegerLiteral{Value: 2},
		},
	})
	if err == nil {
		t.Fatal("expected error for unknown operator")
	}
}

func TestCompileIdentifierBuiltin(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.Identifier{Value: "len"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpConstant) {
		t.Error("expected OpConstant for builtin")
	}
}

func TestCompileIdentifierUndefined(t *testing.T) {
	_, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.Identifier{Value: "undefined_var"},
	})
	if err == nil {
		t.Fatal("expected error for undefined variable")
	}
}

func TestCompileIdentifierTrueFalseEllipsis(t *testing.T) {
	// True, False, Ellipsis are NOT builtins, so they use the fallback path
	tests := []struct {
		name     string
		value    string
		expected Opcode
	}{
		{"True", "True", OpTrue},
		{"False", "False", OpFalse},
		{"Ellipsis", "Ellipsis", OpEllipsis},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCompiler()
			err := c.Compile(&ast.ExpressionStatement{
				Expression: &ast.Identifier{Value: tt.value},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			bc := c.Bytecode()
			if !hasOpcode(bc.Instructions, tt.expected) {
				t.Errorf("expected %d for %s in instructions", tt.expected, tt.value)
			}
		})
	}
}

func TestCompileIdentifierNone(t *testing.T) {
	// None IS a builtin, so it resolves via symbol table as OpConstant
	c := newTestCompiler()
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.Identifier{Value: "None"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bc := c.Bytecode()
	if !hasOpcode(bc.Instructions, OpConstant) {
		t.Error("expected OpConstant for None (it's a builtin)")
	}
}

func TestCompileListLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.ListLiteral{
			Elements: []ast.Expression{&ast.IntegerLiteral{Value: 1}, &ast.IntegerLiteral{Value: 2}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpArray) {
		t.Error("expected OpArray in instructions")
	}
}

func TestCompileSetLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.SetLiteral{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 1}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpSet) {
		t.Error("expected OpSet in instructions")
	}
}

func TestCompileHashLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.HashLiteral{
			Pairs: map[ast.Expression]ast.Expression{
				&ast.StringLiteral{Value: "key"}: &ast.IntegerLiteral{Value: 1},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpHash) {
		t.Error("expected OpHash in instructions")
	}
}

func TestCompileIndexExpression(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{Left: &ast.Identifier{Value: "lst"}, Index: &ast.IntegerLiteral{Value: 0}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpIndex) {
		t.Error("expected OpIndex in instructions")
	}
}

func TestCompileSliceExpression(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{
			Left: &ast.Identifier{Value: "lst"},
			Index: &ast.SliceExpression{Lower: &ast.IntegerLiteral{Value: 1}, Upper: &ast.IntegerLiteral{Value: 3}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpSlice) {
		t.Error("expected OpSlice in instructions")
	}
}

func TestCompileSliceExpressionWithStep(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{
			Left: &ast.Identifier{Value: "lst"},
			Index: &ast.SliceExpression{
				Lower: &ast.IntegerLiteral{Value: 0}, Upper: &ast.IntegerLiteral{Value: 10}, Step: &ast.IntegerLiteral{Value: 2},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpSlice) {
		t.Error("expected OpSlice in instructions")
	}
}

func TestCompileCallExpression(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.CallExpression{
			Function: &ast.Identifier{Value: "len"},
			Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCall) {
		t.Error("expected OpCall in instructions")
	}
}

func TestCompileCallWithKeywordArgs(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.CallExpression{
			Function: &ast.Identifier{Value: "print"},
			Arguments: []ast.Expression{
				&ast.KeywordArgument{Name: &ast.Identifier{Value: "end"}, Value: &ast.StringLiteral{Value: "\n"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCall) {
		t.Error("expected OpCall in instructions")
	}
}

func TestCompileMemberAccess(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.MemberAccess{Object: &ast.Identifier{Value: "math"}, Member: &ast.Identifier{Value: "pi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpGetAttribute) {
		t.Error("expected OpGetAttribute in instructions")
	}
}

func TestCompileMethodCall(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.MethodCall{
			Object: &ast.Identifier{Value: "lst"}, Method: &ast.Identifier{Value: "append"},
			Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 1}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpCall) {
		t.Error("expected OpCall in instructions")
	}
}

func TestCompileAwaitExpression(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("coro")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AwaitExpression{Value: &ast.Identifier{Value: "coro"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpAwait) {
		t.Error("expected OpAwait in instructions")
	}
}

func TestCompileNamedExpression(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.NamedExpression{Name: &ast.Identifier{Value: "x"}, Value: &ast.IntegerLiteral{Value: 42}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpDupTop) {
		t.Error("expected OpDupTop in instructions")
	}
}

func TestCompileFStringLiteral(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("name")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.FStringLiteral{
			Parts: []ast.Expression{&ast.StringLiteral{Value: "hello "}, &ast.Identifier{Value: "name"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpFormatString) {
		t.Error("expected OpFormatString in instructions")
	}
}

func TestCompileDictionaryUnpack(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("d")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.DictionaryUnpack{Value: &ast.Identifier{Value: "d"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpDictUnpack) {
		t.Error("expected OpDictUnpack in instructions")
	}
}

func TestCompileListUnpack(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.ListUnpack{Value: &ast.Identifier{Value: "lst"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpListUnpack) {
		t.Error("expected OpListUnpack in instructions")
	}
}

func TestCompileLambdaExpression(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.LambdaExpression{Parameters: []*ast.Identifier{{Value: "x"}}, Body: &ast.Identifier{Value: "x"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bc.Instructions) == 0 {
		t.Fatal("expected instructions to be generated")
	}
}

// =============================================================================
// 4. Stack-based Compiler: Statement tests
// =============================================================================

func TestCompileLetStatement(t *testing.T) {
	bc, err := compileProgram(&ast.LetStatement{
		Names: []*ast.Identifier{{Value: "x"}}, Value: &ast.IntegerLiteral{Value: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}
}

func TestCompileAssignStatement(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	err := c.Compile(&ast.AssignStatement{
		Names: []*ast.Identifier{{Value: "x"}}, Value: &ast.IntegerLiteral{Value: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in instructions")
	}
}

func TestCompileAttributeAssignStatement(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("obj")
	err := c.Compile(&ast.AttributeAssignStatement{
		Object: &ast.Identifier{Value: "obj"}, Attr: &ast.Identifier{Value: "field"}, Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpSetAttribute) {
		t.Error("expected OpSetAttribute in instructions")
	}
}

func TestCompileIndexAssignStatement(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("d")
	err := c.Compile(&ast.IndexAssignStatement{
		Left: &ast.Identifier{Value: "d"}, Index: &ast.StringLiteral{Value: "key"}, Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpSetIndex) {
		t.Error("expected OpSetIndex in instructions")
	}
}

func TestCompileSliceAssignStatement(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.SliceAssignStatement{
		Left: &ast.Identifier{Value: "lst"}, Lower: &ast.IntegerLiteral{Value: 1},
		Upper: &ast.IntegerLiteral{Value: 3}, Value: &ast.ListLiteral{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 4}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpSetSlice) {
		t.Error("expected OpSetSlice in instructions")
	}
}

func TestCompileAugAssignStatement(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	err := c.Compile(&ast.AugAssignStatement{
		Name: &ast.Identifier{Value: "x"}, Operator: "+", Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpInPlaceAdd) {
		t.Error("expected OpInPlaceAdd in instructions")
	}
}

func TestCompileAugAssignIndexStatement(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("d")
	err := c.Compile(&ast.AugAssignStatement{
		IndexLeft: &ast.Identifier{Value: "d"}, IndexIndex: &ast.StringLiteral{Value: "key"},
		Operator: "+", Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpInPlaceAdd) {
		t.Error("expected OpInPlaceAdd in instructions")
	}
}

func TestCompileReturnStatement(t *testing.T) {
	bc, err := compileProgram(&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 42}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpReturnValue) {
		t.Error("expected OpReturnValue in instructions")
	}
}

func TestCompilePassStatement(t *testing.T) {
	_, err := compileProgram(&ast.PassStatement{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileRaiseStatement(t *testing.T) {
	bc, err := compileProgram(&ast.RaiseStatement{Expression: &ast.Identifier{Value: "ValueError"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpRaise) {
		t.Error("expected OpRaise in instructions")
	}
}

func TestCompileRaiseStatementNil(t *testing.T) {
	bc, err := compileProgram(&ast.RaiseStatement{Expression: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpRaise) {
		t.Error("expected OpRaise in instructions")
	}
}

func TestCompileDeleteStatementIdentifier(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	err := c.Compile(&ast.DeleteStatement{Targets: []ast.Expression{&ast.Identifier{Value: "x"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileDeleteStatementUndefined(t *testing.T) {
	err := newTestCompiler().Compile(&ast.DeleteStatement{
		Targets: []ast.Expression{&ast.Identifier{Value: "undefined"}},
	})
	if err == nil {
		t.Fatal("expected error for deleting undefined variable")
	}
}

func TestCompileDeleteStatementMemberAccess(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("obj")
	err := c.Compile(&ast.DeleteStatement{
		Targets: []ast.Expression{
			&ast.MemberAccess{Object: &ast.Identifier{Value: "obj"}, Member: &ast.Identifier{Value: "field"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpDelAttribute) {
		t.Error("expected OpDelAttribute in instructions")
	}
}

func TestCompileDeleteStatementIndex(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.DeleteStatement{
		Targets: []ast.Expression{
			&ast.IndexExpression{Left: &ast.Identifier{Value: "lst"}, Index: &ast.IntegerLiteral{Value: 0}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileIfExpression(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.IfExpression{
			Condition:   &ast.Boolean{Value: true},
			Consequence: &ast.BlockStatement{Statements: []ast.Statement{&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}}}},
			Alternative: &ast.BlockStatement{Statements: []ast.Statement{&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 2}}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundJump := false
	for _, b := range bc.Instructions {
		if Opcode(b) == OpJumpNotTruthy || Opcode(b) == OpJump {
			foundJump = true
			break
		}
	}
	if !foundJump {
		t.Error("expected jump instruction in if expression")
	}
}

func TestCompileIfExpressionNoAlternative(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.IfExpression{
			Condition:   &ast.Boolean{Value: true},
			Consequence: &ast.BlockStatement{Statements: []ast.Statement{&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpNull) {
		t.Error("expected OpNull for missing alternative")
	}
}

func TestCompileWhileStatement(t *testing.T) {
	bc, err := compileProgram(&ast.WhileStatement{
		Condition: &ast.Boolean{Value: true},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpJumpNotTruthy) {
		t.Error("expected OpJumpNotTruthy in while loop")
	}
}

func TestCompileForStatementError(t *testing.T) {
	err := newTestCompiler().Compile(&ast.ForStatement{
		Value: &ast.Identifier{Value: "i"}, Iterable: &ast.Identifier{Value: "range"}, Body: &ast.BlockStatement{},
	})
	if err == nil {
		t.Fatal("expected error for for statement (should be desugared)")
	}
}

func TestCompileBreakStatementError(t *testing.T) {
	err := newTestCompiler().Compile(&ast.BreakStatement{})
	if err == nil {
		t.Fatal("expected error for break statement (should be desugared)")
	}
}

func TestCompileContinueStatementError(t *testing.T) {
	err := newTestCompiler().Compile(&ast.ContinueStatement{})
	if err == nil {
		t.Fatal("expected error for continue statement (should be desugared)")
	}
}

func TestCompileMatchStatementError(t *testing.T) {
	err := newTestCompiler().Compile(&ast.MatchStatement{
		Subject: &ast.IntegerLiteral{Value: 1}, Cases: []*ast.CaseClause{},
	})
	if err == nil {
		t.Fatal("expected error for match statement (should be desugared)")
	}
}

func TestCompileGlobalStatement(t *testing.T) {
	err := newTestCompiler().Compile(&ast.GlobalStatement{Names: []*ast.Identifier{{Value: "x"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileNonlocalStatement(t *testing.T) {
	err := newTestCompiler().Compile(&ast.NonlocalStatement{Names: []*ast.Identifier{{Value: "x"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileYieldStatement(t *testing.T) {
	bc, err := compileProgram(&ast.YieldStatement{Expression: &ast.IntegerLiteral{Value: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpYieldValue) {
		t.Error("expected OpYieldValue in instructions")
	}
}

func TestCompileYieldStatementNil(t *testing.T) {
	bc, err := compileProgram(&ast.YieldStatement{Expression: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpYieldValue) {
		t.Error("expected OpYieldValue in instructions")
	}
}

func TestCompileFunctionLiteral(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "myFunc", Parameters: []*ast.Identifier{{Value: "x"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.Identifier{Value: "x"}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bc.Constants) == 0 {
		t.Fatal("expected constants to be generated")
	}
}

func TestCompileFunctionLiteralNoName(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Parameters: []*ast.Identifier{{Value: "x"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.Identifier{Value: "x"}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpPop) {
		t.Error("expected OpPop for anonymous function expression")
	}
}

func TestCompileFunctionLiteralWithGlobalNonlocal(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Parameters: []*ast.Identifier{{Value: "x"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.GlobalStatement{Names: []*ast.Identifier{{Value: "g"}}},
				&ast.NonlocalStatement{Names: []*ast.Identifier{{Value: "n"}}},
				&ast.ReturnStatement{ReturnValue: &ast.Identifier{Value: "x"}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bc.Constants) == 0 {
		t.Fatal("expected constants to be generated")
	}
}

func TestCompileFunctionLiteralAsync(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "asyncFunc", IsAsync: true, Parameters: []*ast.Identifier{},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpMakeAsync) {
		t.Error("expected OpMakeAsync in instructions")
	}
}

func TestCompileFunctionLiteralGenerator(t *testing.T) {
	bc, err := compileProgram(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "gen", Parameters: []*ast.Identifier{},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.YieldStatement{Expression: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpMakeGenerator) {
		t.Error("expected OpMakeGenerator in instructions")
	}
}

func TestCompileTryStatement(t *testing.T) {
	bc, err := compileProgram(&ast.TryStatement{
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		Excepts: []*ast.ExceptClause{{
			Type: &ast.Identifier{Value: "ValueError"}, Name: &ast.Identifier{Value: "e"},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpBeginTry) {
		t.Error("expected OpBeginTry in instructions")
	}
}

func TestCompileTryStatementWithFinally(t *testing.T) {
	bc, err := compileProgram(&ast.TryStatement{
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		Excepts: []*ast.ExceptClause{},
		Finally: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpBeginTry) {
		t.Error("expected OpBeginTry in instructions")
	}
	if !hasOpcode(bc.Instructions, OpFinally) {
		t.Error("expected OpFinally in instructions")
	}
}

func TestCompileTryStatementWithStarExcept(t *testing.T) {
	bc, err := compileProgram(&ast.TryStatement{
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		Excepts: []*ast.ExceptClause{{
			Type: &ast.Identifier{Value: "ExceptionGroup"}, IsStar: true,
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpExceptStarHandler) {
		t.Error("expected OpExceptStarHandler in instructions")
	}
}

func TestCompileWithStatement(t *testing.T) {
	bc, err := compileProgram(&ast.WithStatement{
		Items: []*ast.ContextManagerItem{{Expr: &ast.Identifier{Value: "open"}, Name: &ast.Identifier{Value: "f"}}},
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpEnterContext) {
		t.Error("expected OpEnterContext in instructions")
	}
	if !hasOpcode(bc.Instructions, OpExitContext) {
		t.Error("expected OpExitContext in instructions")
	}
}

func TestCompileWithStatementNoName(t *testing.T) {
	bc, err := compileProgram(&ast.WithStatement{
		Items: []*ast.ContextManagerItem{{Expr: &ast.Identifier{Value: "open"}, Name: nil}},
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpPop) {
		t.Error("expected OpPop when with item has no name")
	}
}

func TestCompileClassStatement(t *testing.T) {
	bc, err := compileProgram(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"}, Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "__init__", Parameters: []*ast.Identifier{{Value: "self"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCreateClass) {
		t.Error("expected OpCreateClass in instructions")
	}
}

func TestCompileClassStatementWithSuper(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("Base")
	err := c.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "Child"}, SuperClass: &ast.Identifier{Value: "Base"},
		Body: &ast.BlockStatement{}, Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpCreateClassWithSuper) {
		t.Error("expected OpCreateClassWithSuper in instructions")
	}
}

func TestCompileClassStatementWithMultiSuper(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("Base1")
	c.symbolTable.Define("Base2")
	err := c.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "Child"}, SuperClasses: []*ast.Identifier{{Value: "Base1"}, {Value: "Base2"}},
		Body: &ast.BlockStatement{}, Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpCreateClassWithMultiSuper) {
		t.Error("expected OpCreateClassWithMultiSuper in instructions")
	}
}

func TestCompileClassStatementWithMetaclass(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("Meta")
	err := c.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"}, Metaclass: &ast.Identifier{Value: "Meta"},
		Body: &ast.BlockStatement{}, Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(c.Bytecode().Instructions, OpSetMetaclass) {
		t.Error("expected OpSetMetaclass in instructions")
	}
}

func TestCompileClassStatementWithBodyAssign(t *testing.T) {
	bc, err := compileProgram(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AssignStatement{Names: []*ast.Identifier{{Value: "class_var"}}, Value: &ast.IntegerLiteral{Value: 42}},
		}},
		Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpSetClassField) {
		t.Error("expected OpSetClassField in instructions")
	}
}

func TestCompileImportStatement(t *testing.T) {
	bc, err := compileProgram(&ast.ImportStatement{Module: &ast.Identifier{Value: "math"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in import instructions")
	}
}

func TestCompileImportStatementNotFound(t *testing.T) {
	err := newTestCompiler().Compile(&ast.ImportStatement{Module: &ast.Identifier{Value: "nonexistent_module"}})
	if err == nil {
		t.Fatal("expected error for non-existent module")
	}
}

func TestCompileImportStatementWithAlias(t *testing.T) {
	bc, err := compileProgram(&ast.ImportStatement{
		Module: &ast.Identifier{Value: "math"}, Alias: &ast.Identifier{Value: "m"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in import with alias")
	}
}

func TestCompileFromImportStatement(t *testing.T) {
	bc, err := compileProgram(&ast.FromImportStatement{
		Module: &ast.Identifier{Value: "math"}, Names: []*ast.Identifier{{Value: "pi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in from-import instructions")
	}
}

func TestCompileFromImportStatementNotFound(t *testing.T) {
	err := newTestCompiler().Compile(&ast.FromImportStatement{
		Module: &ast.Identifier{Value: "nonexistent"}, Names: []*ast.Identifier{{Value: "x"}},
	})
	if err == nil {
		t.Fatal("expected error for non-existent module")
	}
}

func TestCompileFromImportStatementNameNotFound(t *testing.T) {
	err := newTestCompiler().Compile(&ast.FromImportStatement{
		Module: &ast.Identifier{Value: "math"}, Names: []*ast.Identifier{{Value: "nonexistent_name"}},
	})
	if err == nil {
		t.Fatal("expected error for non-existent name in module")
	}
}

func TestCompileFromImportStatementWithAlias(t *testing.T) {
	bc, err := compileProgram(&ast.FromImportStatement{
		Module: &ast.Identifier{Value: "math"}, Alias: &ast.Identifier{Value: "m"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpSetGlobal) {
		t.Error("expected OpSetGlobal in from-import with alias")
	}
}

func TestCompileNilNode(t *testing.T) {
	err := newTestCompiler().Compile(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileBlockStatement(t *testing.T) {
	bc, err := compileProgram(&ast.BlockStatement{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}},
			&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 2}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bc.Instructions) == 0 {
		t.Fatal("expected instructions to be generated")
	}
}

func TestCompileCaseClause(t *testing.T) {
	err := newTestCompiler().Compile(&ast.CaseClause{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileSliceExpressionStandalone(t *testing.T) {
	err := newTestCompiler().Compile(&ast.SliceExpression{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 5. Compiler internal method tests
// =============================================================================

func TestCompilerAddConstant(t *testing.T) {
	c := newTestCompiler()
	idx := c.addConstant(&objects.Integer{Value: 42})
	if idx != len(c.constants)-1 {
		t.Errorf("expected index %d, got %d", len(c.constants)-1, idx)
	}
}

func TestCompilerEmit(t *testing.T) {
	c := newTestCompiler()
	pos := c.emit(OpConstant, 0)
	if pos != 0 {
		t.Errorf("expected position 0, got %d", pos)
	}
}

func TestCompilerEmit1(t *testing.T) {
	c := newTestCompiler()
	pos := c.emit1(OpCall, 3)
	if pos != 0 {
		t.Errorf("expected position 0, got %d", pos)
	}
}

func TestCompilerEmitClosure(t *testing.T) {
	c := newTestCompiler()
	fn := &CompiledFunction{Instructions: []byte{byte(OpReturn)}, NumLocals: 0}
	idx := c.addConstant(fn)
	pos := c.emitClosure(idx, 2)
	if pos != 0 {
		t.Errorf("expected position 0, got %d", pos)
	}
}

func TestCompilerLastInstructionIs(t *testing.T) {
	c := newTestCompiler()
	if c.lastInstructionIs(OpPop) {
		t.Error("expected false when no instructions")
	}
	c.emit(OpPop)
	if !c.lastInstructionIs(OpPop) {
		t.Error("expected true for last instruction OpPop")
	}
}

func TestCompilerRemoveLastPop(t *testing.T) {
	c := newTestCompiler()
	c.emit(OpConstant, 0)
	c.emit(OpPop)
	initialLen := len(c.instructions)
	c.removeLastPop()
	if len(c.instructions) >= initialLen {
		t.Error("expected instructions to be shorter after removing last pop")
	}
}

func TestCompilerReplaceLastPopWithReturn(t *testing.T) {
	c := newTestCompiler()
	c.emit(OpPop)
	c.replaceLastPopWithReturn()
	if !c.lastInstructionIs(OpReturnValue) {
		t.Error("expected last instruction to be OpReturnValue")
	}
}

func TestCompilerChangeOperand(t *testing.T) {
	c := newTestCompiler()
	pos := c.emit(OpJump, 9999)
	c.changeOperand(pos, 42)
	hi := c.instructions[pos+1]
	lo := c.instructions[pos+2]
	newOperand := int(uint16(hi)<<8 | uint16(lo))
	if newOperand != 42 {
		t.Errorf("expected operand 42, got %d", newOperand)
	}
}

func TestCompilerAdjustLocalIndices(t *testing.T) {
	c := newTestCompiler()
	ins := []byte{byte(OpGetLocal), 0, byte(OpSetLocal), 1}
	result := c.adjustLocalIndices(ins, 2)
	if result[1] != 2 {
		t.Errorf("expected adjusted local index 2, got %d", result[1])
	}
	if result[3] != 3 {
		t.Errorf("expected adjusted local index 3, got %d", result[3])
	}
}

func TestCompilerAdjustLocalIndicesNoFree(t *testing.T) {
	c := newTestCompiler()
	ins := []byte{byte(OpGetLocal), 0}
	result := c.adjustLocalIndices(ins, 0)
	if len(result) != len(ins) {
		t.Error("expected same length when numFree is 0")
	}
}

func TestCompilerEnterExitScope(t *testing.T) {
	c := newTestCompiler()
	originalST := c.symbolTable
	c.enterScope()
	if c.symbolTable == originalST {
		t.Error("expected new symbol table after entering scope")
	}
	c.exitScope()
	if c.symbolTable != originalST {
		t.Error("expected original symbol table after exiting scope")
	}
}

func TestCompilerBytecode(t *testing.T) {
	c := newTestCompiler()
	c.emit(OpConstant, 0)
	bc := c.Bytecode()
	if bc == nil {
		t.Fatal("expected non-nil bytecode")
	}
}

func TestCompilerSymbolTable(t *testing.T) {
	c := newTestCompiler()
	if c.SymbolTable() == nil {
		t.Fatal("expected non-nil symbol table")
	}
}

func TestCompilerMake(t *testing.T) {
	c := newTestCompiler()
	ins := c.make(OpConstant, 42)
	if len(ins) != 3 {
		t.Errorf("expected 3 bytes, got %d", len(ins))
	}
	if ins[0] != byte(OpConstant) {
		t.Errorf("expected OpConstant, got %d", ins[0])
	}
}

func TestCompilerMakeOperand(t *testing.T) {
	c := newTestCompiler()
	op := c.makeOperand(256)
	if len(op) != 2 {
		t.Errorf("expected 2 bytes, got %d", len(op))
	}
	if op[0] != 1 || op[1] != 0 {
		t.Errorf("expected [1, 0] for 256, got [%d, %d]", op[0], op[1])
	}
}

func TestCompilerMakeOperand1(t *testing.T) {
	c := newTestCompiler()
	op := c.makeOperand1(42)
	if len(op) != 1 {
		t.Errorf("expected 1 byte, got %d", len(op))
	}
	if op[0] != 42 {
		t.Errorf("expected 42, got %d", op[0])
	}
}

// =============================================================================
// 6. CompiledFunction tests
// =============================================================================

func TestCompiledFunctionType(t *testing.T) {
	fn := &CompiledFunction{}
	if fn.Type() != objects.FUNCTION_OBJ {
		t.Errorf("expected FUNCTION_OBJ, got %v", fn.Type())
	}
}

func TestCompiledFunctionInspect(t *testing.T) {
	fn := &CompiledFunction{}
	if fn.Inspect() != "compiled function" {
		t.Errorf("expected 'compiled function', got %q", fn.Inspect())
	}
}

func TestCompiledFunctionHasNonEscapingLocals(t *testing.T) {
	fn := &CompiledFunction{NonEscapingLocals: []bool{true, false}}
	if !fn.HasNonEscapingLocals() {
		t.Error("expected HasNonEscapingLocals to be true")
	}
	fn2 := &CompiledFunction{NonEscapingLocals: []bool{false, false}}
	if fn2.HasNonEscapingLocals() {
		t.Error("expected HasNonEscapingLocals to be false")
	}
}

func TestCompiledFunctionIsLocalNonEscaping(t *testing.T) {
	fn := &CompiledFunction{NonEscapingLocals: []bool{true, false}}
	if !fn.IsLocalNonEscaping(0) {
		t.Error("expected index 0 to be non-escaping")
	}
	if fn.IsLocalNonEscaping(1) {
		t.Error("expected index 1 to be escaping")
	}
	if fn.IsLocalNonEscaping(-1) {
		t.Error("expected negative index to be escaping")
	}
	if fn.IsLocalNonEscaping(10) {
		t.Error("expected out-of-range index to be escaping")
	}
}

// =============================================================================
// 7. Optimization tests
// =============================================================================

func TestInstructionSize(t *testing.T) {
	tests := []struct {
		opcode   Opcode
		expected int
	}{
		{OpConstant, 3}, {OpJump, 3}, {OpJumpNotTruthy, 3},
		{OpSetGlobal, 3}, {OpGetGlobal, 3}, {OpArray, 3},
		{OpHash, 3}, {OpSet, 3}, {OpClosure, 4},
		{OpBeginTry, 9}, {OpExceptHandler, 5}, {OpExceptStarHandler, 5},
		{OpCall, 2}, {OpSetLocal, 2}, {OpGetLocal, 2}, {OpGetFree, 2},
		{OpAdd, 1}, {OpSub, 1}, {OpPop, 1}, {OpReturn, 1}, {OpReturnValue, 1},
	}
	for _, tt := range tests {
		size := InstructionSize(tt.opcode)
		if size != tt.expected {
			t.Errorf("InstructionSize(%d): expected %d, got %d", tt.opcode, tt.expected, size)
		}
	}
}

func TestIsTerminator(t *testing.T) {
	tests := []struct {
		opcode   Opcode
		expected bool
	}{
		{OpReturnValue, true}, {OpReturn, true}, {OpJump, true}, {OpRaise, true},
		{OpAdd, false}, {OpPop, false},
	}
	for _, tt := range tests {
		result := isTerminator(tt.opcode)
		if result != tt.expected {
			t.Errorf("isTerminator(%d): expected %v, got %v", tt.opcode, tt.expected, result)
		}
	}
}

func TestEliminateDeadCodeEmpty(t *testing.T) {
	result := EliminateDeadCode([]byte{})
	if len(result) != 0 {
		t.Error("expected empty result for empty input")
	}
}

func TestEliminateDeadCodeSimple(t *testing.T) {
	c := newTestCompiler()
	ins := c.make(OpConstant, 0)
	result := EliminateDeadCode(ins)
	if len(result) != len(ins) {
		t.Errorf("expected same length %d, got %d", len(ins), len(result))
	}
}

func TestEliminateDeadCodeUnreachable(t *testing.T) {
	c := newTestCompiler()
	ins := append(c.make(OpReturn), c.make(OpConstant, 0)...)
	result := EliminateDeadCode(ins)
	if len(result) >= len(ins) {
		t.Error("expected dead code to be eliminated")
	}
}

func TestEliminateDeadCodeInFunctions(t *testing.T) {
	fn := &CompiledFunction{Instructions: []byte{byte(OpReturn)}, NumLocals: 0}
	bc := &Bytecode{Instructions: []byte{byte(OpReturn)}, Constants: []objects.Object{fn}}
	result := EliminateDeadCodeInFunctions(bc)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// =============================================================================
// 8. Helper function tests
// =============================================================================

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("expected 1 for true")
	}
	if boolToInt(false) != 0 {
		t.Error("expected 0 for false")
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		obj      objects.Object
		expected bool
	}{
		{"true bool", &objects.Boolean{Value: true}, true},
		{"false bool", &objects.Boolean{Value: false}, false},
		{"nonzero int", &objects.Integer{Value: 1}, true},
		{"zero int", &objects.Integer{Value: 0}, false},
		{"nonzero float", &objects.Float{Value: 1.0}, true},
		{"zero float", &objects.Float{Value: 0.0}, false},
		{"nonempty string", &objects.String{Value: "hello"}, true},
		{"empty string", &objects.String{Value: ""}, false},
		{"none", objects.None_, false},
		{"nonempty list", &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}}, true},
		{"empty list", &objects.List{Elements: []objects.Object{}}, false},
		{"nonempty tuple", &objects.Tuple{Elements: []objects.Object{&objects.Integer{Value: 1}}}, true},
		{"empty tuple", &objects.Tuple{Elements: []objects.Object{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTruthy(tt.obj)
			if result != tt.expected {
				t.Errorf("isTruthy(%s): expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func TestCompareObjects(t *testing.T) {
	tests := []struct {
		name     string
		a, b     objects.Object
		expected int
	}{
		{"int less", &objects.Integer{Value: 1}, &objects.Integer{Value: 2}, -1},
		{"int equal", &objects.Integer{Value: 1}, &objects.Integer{Value: 1}, 0},
		{"int greater", &objects.Integer{Value: 2}, &objects.Integer{Value: 1}, 1},
		{"float less", &objects.Float{Value: 1.0}, &objects.Float{Value: 2.0}, -1},
		{"string less", &objects.String{Value: "a"}, &objects.String{Value: "b"}, -1},
		{"string equal", &objects.String{Value: "a"}, &objects.String{Value: "a"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareObjects(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareObjects(%s): expected %d, got %d", tt.name, tt.expected, result)
			}
		})
	}
}

func TestAugAssignToInPlaceOp(t *testing.T) {
	tests := []struct {
		operator string
		expected Opcode
	}{
		{"+", OpInPlaceAdd}, {"-", OpInPlaceSub}, {"*", OpInPlaceMul},
		{"/", OpInPlaceDiv}, {"%", OpInPlaceMod}, {"//", OpInPlaceFloorDiv},
		{"**", OpInPlacePower}, {"|", OpInPlaceBitOr}, {"&", OpInPlaceBitAnd},
		{"^", OpInPlaceBitXor}, {"<<", OpInPlaceLShift}, {">>", OpInPlaceRShift},
		{"unknown", OpInPlaceAdd},
	}
	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			result := augAssignToInPlaceOp(tt.operator)
			if result != tt.expected {
				t.Errorf("augAssignToInPlaceOp(%q): expected %d, got %d", tt.operator, tt.expected, result)
			}
		})
	}
}

// =============================================================================
// 9. Register Compiler: Expression tests
// =============================================================================

func TestRegCompilerIntegerLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 42}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpLoadConst) {
		t.Error("expected ROpLoadConst")
	}
}

func TestRegCompilerFloatLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.FloatLiteral{Value: 3.14}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpLoadConst) {
		t.Error("expected ROpLoadConst")
	}
}

func TestRegCompilerStringLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.StringLiteral{Value: "hello"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpLoadConst) {
		t.Error("expected ROpLoadConst")
	}
}

func TestRegCompilerBoolean(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.Boolean{Value: true}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpTrue) {
		t.Error("expected ROpTrue")
	}
}

func TestRegCompilerEllipsis(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.EllipsisLiteral{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpEllipsis) {
		t.Error("expected ROpEllipsis")
	}
}

func TestRegCompilerIdentifierBuiltin(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.Identifier{Value: "len"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpLoadConst) {
		t.Error("expected ROpLoadConst for builtin")
	}
}

func TestRegCompilerIdentifierUndefined(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.Identifier{Value: "undefined_var"}})
	if err == nil {
		t.Fatal("expected error for undefined variable")
	}
}

func TestRegCompilerIdentifierTrueFalseNone(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.Identifier{Value: "True"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpTrue) {
		t.Error("expected ROpTrue for True identifier")
	}
}

func TestRegCompilerPrefixExpression(t *testing.T) {
	tests := []struct {
		operator string
		expected RegOpcode
	}{
		{"-", ROpNegate}, {"!", ROpNot},
	}
	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			rc := newTestRegCompiler()
			err := rc.Compile(&ast.ExpressionStatement{
				Expression: &ast.PrefixExpression{Operator: tt.operator, Right: &ast.IntegerLiteral{Value: 1}},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !hasRegOpcode(rc.Instructions(), tt.expected) {
				t.Errorf("expected %d in instructions", tt.expected)
			}
		})
	}
}

func TestRegCompilerInfixExpression(t *testing.T) {
	tests := []struct {
		operator string
		expected RegOpcode
	}{
		{"+", ROpAdd}, {"-", ROpSub}, {"*", ROpMul}, {"/", ROpDiv},
		{"%", ROpMod}, {"//", ROpFloorDiv}, {"**", ROpPower},
		{"|", ROpBitOr}, {"&", ROpBitAnd}, {"^", ROpBitXor},
	}
	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			rc := newTestRegCompiler()
			err := rc.Compile(&ast.ExpressionStatement{
				Expression: &ast.InfixExpression{
					Left: &ast.IntegerLiteral{Value: 1}, Operator: tt.operator, Right: &ast.IntegerLiteral{Value: 2},
				},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !hasRegOpcode(rc.Instructions(), tt.expected) {
				t.Errorf("expected %d in instructions", tt.expected)
			}
		})
	}
}

func TestRegCompilerCompareOperators(t *testing.T) {
	for _, op := range []string{"==", "!=", ">", "<"} {
		t.Run(op, func(t *testing.T) {
			rc := newTestRegCompiler()
			err := rc.Compile(&ast.ExpressionStatement{
				Expression: &ast.InfixExpression{
					Left: &ast.IntegerLiteral{Value: 1}, Operator: op, Right: &ast.IntegerLiteral{Value: 2},
				},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !hasRegOpcode(rc.Instructions(), ROpCompare) {
				t.Errorf("expected ROpCompare for %s", op)
			}
		})
	}
}

func TestRegCompilerUnknownOperatorError(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.InfixExpression{
			Left: &ast.IntegerLiteral{Value: 1}, Operator: "???", Right: &ast.IntegerLiteral{Value: 2},
		},
	})
	if err == nil {
		t.Fatal("expected error for unknown operator")
	}
}

func TestRegCompilerCallExpression(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.CallExpression{
			Function: &ast.Identifier{Value: "len"}, Arguments: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpCall) {
		t.Error("expected ROpCall in instructions")
	}
}

func TestRegCompilerListLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.ListLiteral{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 1}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpBuildList) {
		t.Error("expected ROpBuildList")
	}
}

func TestRegCompilerHashLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.HashLiteral{
			Pairs: map[ast.Expression]ast.Expression{
				&ast.StringLiteral{Value: "key"}: &ast.IntegerLiteral{Value: 1},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpBuildDict) {
		t.Error("expected ROpBuildDict")
	}
}

func TestRegCompilerSetLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.SetLiteral{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 1}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpBuildSet) {
		t.Error("expected ROpBuildSet")
	}
}

func TestRegCompilerIndexExpression(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{Left: &ast.Identifier{Value: "lst"}, Index: &ast.IntegerLiteral{Value: 0}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpIndex) {
		t.Error("expected ROpIndex")
	}
}

func TestRegCompilerSliceExpression(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{
			Left: &ast.Identifier{Value: "lst"},
			Index: &ast.SliceExpression{Lower: &ast.IntegerLiteral{Value: 1}, Upper: &ast.IntegerLiteral{Value: 3}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSlice) {
		t.Error("expected ROpSlice")
	}
}

func TestRegCompilerMemberAccess(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.MemberAccess{Object: &ast.Identifier{Value: "math"}, Member: &ast.Identifier{Value: "pi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpGetAttr) {
		t.Error("expected ROpGetAttr")
	}
}

func TestRegCompilerMethodCall(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.MethodCall{
			Object: &ast.Identifier{Value: "lst"}, Method: &ast.Identifier{Value: "append"},
			Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 1}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpCall) {
		t.Error("expected ROpCall")
	}
}

func TestRegCompilerComplexLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.ComplexLiteral{Value: "2.5j"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerByteStringLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{Expression: &ast.ByteStringLiteral{Value: "hello"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerFStringLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FStringLiteral{Parts: []ast.Expression{&ast.StringLiteral{Value: "hello "}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpFormatString) {
		t.Error("expected ROpFormatString")
	}
}

func TestRegCompilerTernaryExpression(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.TernaryExpression{
			Condition: &ast.Boolean{Value: true}, Consequence: &ast.IntegerLiteral{Value: 1},
			Alternative: &ast.IntegerLiteral{Value: 2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpJumpIfFalse) {
		t.Error("expected ROpJumpIfFalse")
	}
}

func TestRegCompilerNamedExpression(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.NamedExpression{Name: &ast.Identifier{Value: "x"}, Value: &ast.IntegerLiteral{Value: 42}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetGlobal) {
		t.Error("expected ROpSetGlobal")
	}
}

func TestRegCompilerAwaitExpression(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("coro")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.AwaitExpression{Value: &ast.Identifier{Value: "coro"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpAwait) {
		t.Error("expected ROpAwait")
	}
}

func TestRegCompilerDictionaryUnpack(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("d")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.DictionaryUnpack{Value: &ast.Identifier{Value: "d"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpDictUnpack) {
		t.Error("expected ROpDictUnpack")
	}
}

func TestRegCompilerListUnpack(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.ListUnpack{Value: &ast.Identifier{Value: "lst"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpListUnpack) {
		t.Error("expected ROpListUnpack")
	}
}

func TestRegCompilerClassInstantiation(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("MyClass")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.ClassInstantiation{
			ClassName: &ast.Identifier{Value: "MyClass"}, Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 1}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpCall) {
		t.Error("expected ROpCall")
	}
}

func TestRegCompilerComprehensionsReturnNull(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.ListComprehension{
			Element: &ast.Identifier{Value: "x"}, Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "range"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpNull) {
		t.Error("expected ROpNull for comprehension fallback")
	}
}

func TestRegCompilerKeywordArgument(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.KeywordArgument{Name: &ast.Identifier{Value: "key"}, Value: &ast.IntegerLiteral{Value: 1}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerSpreadItem(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.SpreadItem{Token: "*", Value: &ast.Identifier{Value: "lst"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerYieldStatementExpr(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.YieldStatement{Expression: &ast.IntegerLiteral{Value: 1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpYieldValue) {
		t.Error("expected ROpYieldValue")
	}
}

func TestRegCompilerDefaultExpression(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.GeneratorExpression{
			Element: &ast.Identifier{Value: "x"}, Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "range"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 10. Register Compiler: Statement tests
// =============================================================================

func TestRegCompilerLetStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.LetStatement{Names: []*ast.Identifier{{Value: "x"}}, Value: &ast.IntegerLiteral{Value: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetGlobal) {
		t.Error("expected ROpSetGlobal")
	}
}

func TestRegCompilerAssignStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("x")
	err := rc.Compile(&ast.AssignStatement{Names: []*ast.Identifier{{Value: "x"}}, Value: &ast.IntegerLiteral{Value: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerReturnStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 42}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpReturn) {
		t.Error("expected ROpReturn")
	}
}

func TestRegCompilerWhileStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.WhileStatement{
		Condition: &ast.Boolean{Value: true},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpJumpIfFalse) {
		t.Error("expected ROpJumpIfFalse")
	}
}

func TestRegCompilerAugAssignStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("x")
	err := rc.Compile(&ast.AugAssignStatement{
		Name: &ast.Identifier{Value: "x"}, Operator: "+", Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpInPlaceAdd) {
		t.Error("expected ROpInPlaceAdd")
	}
}

func TestRegCompilerDeleteStatementIdentifier(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("x")
	err := rc.Compile(&ast.DeleteStatement{Targets: []ast.Expression{&ast.Identifier{Value: "x"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerDeleteStatementUndefined(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.DeleteStatement{Targets: []ast.Expression{&ast.Identifier{Value: "undefined"}}})
	if err == nil {
		t.Fatal("expected error for deleting undefined variable")
	}
}

func TestRegCompilerDeleteStatementMemberAccess(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("obj")
	err := rc.Compile(&ast.DeleteStatement{
		Targets: []ast.Expression{
			&ast.MemberAccess{Object: &ast.Identifier{Value: "obj"}, Member: &ast.Identifier{Value: "field"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpDelAttribute) {
		t.Error("expected ROpDelAttribute")
	}
}

func TestRegCompilerRaiseStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("ValueError")
	err := rc.Compile(&ast.RaiseStatement{Expression: &ast.Identifier{Value: "ValueError"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpRaise) {
		t.Error("expected ROpRaise")
	}
}

func TestRegCompilerRaiseStatementNil(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.RaiseStatement{Expression: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerClassStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"}, Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "__init__", Parameters: []*ast.Identifier{{Value: "self"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpCreateClass) {
		t.Error("expected ROpCreateClass")
	}
}

func TestRegCompilerClassStatementWithSuper(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("Base")
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "Child"}, SuperClass: &ast.Identifier{Value: "Base"},
		Body: &ast.BlockStatement{}, Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpCreateClassWithSuper) {
		t.Error("expected ROpCreateClassWithSuper")
	}
}

func TestRegCompilerTryStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.TryStatement{
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		Excepts: []*ast.ExceptClause{{
			Type: &ast.Identifier{Value: "ValueError"},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpBeginTry) {
		t.Error("expected ROpBeginTry")
	}
}

func TestRegCompilerWithStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("open")
	err := rc.Compile(&ast.WithStatement{
		Items: []*ast.ContextManagerItem{{Expr: &ast.Identifier{Value: "open"}, Name: &ast.Identifier{Value: "f"}}},
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpEnterContext) {
		t.Error("expected ROpEnterContext")
	}
}

func TestRegCompilerPassStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.PassStatement{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerGlobalNonlocalStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.GlobalStatement{Names: []*ast.Identifier{{Value: "x"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = rc.Compile(&ast.NonlocalStatement{Names: []*ast.Identifier{{Value: "y"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerForStatementError(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ForStatement{
		Value: &ast.Identifier{Value: "i"}, Iterable: &ast.Identifier{Value: "range"}, Body: &ast.BlockStatement{},
	})
	if err == nil {
		t.Fatal("expected error for for statement")
	}
}

func TestRegCompilerBreakContinueError(t *testing.T) {
	rc := newTestRegCompiler()
	if err := rc.Compile(&ast.BreakStatement{}); err == nil {
		t.Fatal("expected error for break statement")
	}
	if err := rc.Compile(&ast.ContinueStatement{}); err == nil {
		t.Fatal("expected error for continue statement")
	}
}

func TestRegCompilerMatchStatementError(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.MatchStatement{Subject: &ast.IntegerLiteral{Value: 1}})
	if err == nil {
		t.Fatal("expected error for match statement")
	}
}

func TestRegCompilerImportStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ImportStatement{Module: &ast.Identifier{Value: "math"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerImportStatementNotFound(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ImportStatement{Module: &ast.Identifier{Value: "nonexistent"}})
	if err == nil {
		t.Fatal("expected error for non-existent module")
	}
}

func TestRegCompilerFromImportStatement(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.FromImportStatement{Module: &ast.Identifier{Value: "math"}, Names: []*ast.Identifier{{Value: "pi"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerFromImportStatementNotFound(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.FromImportStatement{Module: &ast.Identifier{Value: "nonexistent"}, Names: []*ast.Identifier{{Value: "x"}}})
	if err == nil {
		t.Fatal("expected error for non-existent module")
	}
}

func TestRegCompilerFromImportStatementNameNotFound(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.FromImportStatement{Module: &ast.Identifier{Value: "math"}, Names: []*ast.Identifier{{Value: "nonexistent_name"}}})
	if err == nil {
		t.Fatal("expected error for non-existent name in module")
	}
}

func TestRegCompilerYieldFromStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("gen")
	err := rc.Compile(&ast.YieldFromStatement{Expression: &ast.Identifier{Value: "gen"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerAsyncForStatementError(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.AsyncForStatement{
		Value: &ast.Identifier{Value: "i"}, Iterable: &ast.Identifier{Value: "aiter"}, Body: &ast.BlockStatement{},
	})
	if err == nil {
		t.Fatal("expected error for async for statement")
	}
}

func TestRegCompilerAsyncWithStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("cm")
	err := rc.Compile(&ast.AsyncWithStatement{
		Items: []*ast.ContextManagerItem{{Expr: &ast.Identifier{Value: "cm"}, Name: &ast.Identifier{Value: "x"}}},
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerAttributeAssignStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("obj")
	err := rc.Compile(&ast.AttributeAssignStatement{
		Object: &ast.Identifier{Value: "obj"}, Attr: &ast.Identifier{Value: "field"}, Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetAttr) {
		t.Error("expected ROpSetAttr")
	}
}

func TestRegCompilerIndexAssignStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("d")
	err := rc.Compile(&ast.IndexAssignStatement{
		Left: &ast.Identifier{Value: "d"}, Index: &ast.StringLiteral{Value: "key"}, Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetIndex) {
		t.Error("expected ROpSetIndex")
	}
}

func TestRegCompilerSliceAssignStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.SliceAssignStatement{
		Left: &ast.Identifier{Value: "lst"}, Lower: &ast.IntegerLiteral{Value: 1},
		Upper: &ast.IntegerLiteral{Value: 3}, Value: &ast.ListLiteral{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 4}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetSlice) {
		t.Error("expected ROpSetSlice")
	}
}

func TestRegCompilerYieldStatementNil(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.YieldStatement{Expression: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerNilNode(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.compileStmt(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerFunctionLiteral(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "myFunc", Parameters: []*ast.Identifier{{Value: "x"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.Identifier{Value: "x"}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerFunctionLiteralAsync(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "asyncFunc", IsAsync: true, Parameters: []*ast.Identifier{},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpMakeAsync) {
		t.Error("expected ROpMakeAsync")
	}
}

func TestRegCompilerFunctionLiteralGenerator(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "gen", Parameters: []*ast.Identifier{},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.YieldStatement{Expression: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpMakeGenerator) {
		t.Error("expected ROpMakeGenerator")
	}
}

func TestRegCompilerLambdaExpression(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.LambdaExpression{Parameters: []*ast.Identifier{{Value: "x"}}, Body: &ast.Identifier{Value: "x"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerIfExpression(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.IfExpression{
			Condition: &ast.Boolean{Value: true},
			Consequence: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
			Alternative: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpJumpIfFalse) {
		t.Error("expected ROpJumpIfFalse")
	}
}

func TestRegCompilerIfExpressionNoConsequence(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.IfExpression{Condition: &ast.Boolean{Value: true}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerIfExpressionNoAlternative(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.IfExpression{
			Condition: &ast.Boolean{Value: true},
			Consequence: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 11. Register Compiler internal method tests
// =============================================================================

func TestRegCompilerAllocFreeReg(t *testing.T) {
	rc := newTestRegCompiler()
	r1 := rc.allocReg()
	r2 := rc.allocReg()
	if r1 != 0 {
		t.Errorf("expected first register to be 0, got %d", r1)
	}
	if r2 != 1 {
		t.Errorf("expected second register to be 1, got %d", r2)
	}
	rc.freeReg(r1)
	r3 := rc.allocReg()
	if r3 != 0 {
		t.Errorf("expected reused register to be 0, got %d", r3)
	}
}

func TestRegCompilerAddConstant(t *testing.T) {
	rc := newTestRegCompiler()
	idx := rc.addConstant(&objects.Integer{Value: 42})
	if idx < 0 {
		t.Errorf("expected non-negative index, got %d", idx)
	}
}

func TestRegCompilerSymbolTable(t *testing.T) {
	rc := newTestRegCompiler()
	if rc.SymbolTable() == nil {
		t.Fatal("expected non-nil symbol table")
	}
}

func TestRegCompilerConstants(t *testing.T) {
	rc := newTestRegCompiler()
	if rc.Constants() == nil {
		t.Fatal("expected non-nil constants")
	}
}

func TestRegCompilerBytecode(t *testing.T) {
	rc := newTestRegCompiler()
	rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 42}})
	bc := rc.Bytecode()
	if bc == nil {
		t.Fatal("expected non-nil bytecode")
	}
	if bc.NumRegs < 1 {
		t.Error("expected at least 1 register")
	}
}

func TestRegCompilerPushPopScope(t *testing.T) {
	rc := newTestRegCompiler()
	originalST := rc.symbolTable
	rc.pushScope()
	if rc.symbolTable == originalST {
		t.Error("expected new symbol table after pushScope")
	}
	rc.popScope()
	if rc.symbolTable != originalST {
		t.Error("expected original symbol table after popScope")
	}
}

func TestRegCompilerPopScopeNoOuter(t *testing.T) {
	rc := newTestRegCompiler()
	st := rc.symbolTable
	rc.popScope()
	if rc.symbolTable != st {
		t.Error("expected same symbol table when no outer")
	}
}

func TestRegCompilerEmitSetSymbol(t *testing.T) {
	rc := newTestRegCompiler()
	rc.emitSetSymbol(Symbol{Name: "x", Scope: GlobalScope, Index: 0}, 0)
	if len(rc.Instructions()) != 1 {
		t.Fatal("expected 1 instruction")
	}
	if rc.Instructions()[0].Opcode != ROpSetGlobal {
		t.Errorf("expected ROpSetGlobal, got %d", rc.Instructions()[0].Opcode)
	}
}

func TestRegCompilerEmitSetSymbolLocal(t *testing.T) {
	rc := newTestRegCompiler()
	rc.emitSetSymbol(Symbol{Name: "x", Scope: LocalScope, Index: 0}, 0)
	if rc.Instructions()[0].Opcode != ROpSetLocal {
		t.Errorf("expected ROpSetLocal, got %d", rc.Instructions()[0].Opcode)
	}
}

func TestRegCompilerEmitSetSymbolFree(t *testing.T) {
	rc := newTestRegCompiler()
	rc.emitSetSymbol(Symbol{Name: "x", Scope: FreeScope, Index: 0}, 0)
	if rc.Instructions()[0].Opcode != ROpSetLocal {
		t.Errorf("expected ROpSetLocal for FreeScope, got %d", rc.Instructions()[0].Opcode)
	}
}

func TestAugAssignToRegInPlaceOp(t *testing.T) {
	tests := []struct {
		operator string
		expected RegOpcode
	}{
		{"+", ROpInPlaceAdd}, {"-", ROpInPlaceSub}, {"*", ROpInPlaceMul},
		{"/", ROpInPlaceDiv}, {"%", ROpInPlaceMod}, {"//", ROpInPlaceFloorDiv},
		{"**", ROpInPlacePower}, {"|", ROpInPlaceBitOr}, {"&", ROpInPlaceBitAnd},
		{"^", ROpInPlaceBitXor}, {"<<", ROpInPlaceLShift}, {">>", ROpInPlaceRShift},
		{"unknown", ROpInPlaceAdd},
	}
	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			result := augAssignToRegInPlaceOp(tt.operator)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestRegInstructionsToBytes(t *testing.T) {
	instrs := []RegInstruction{
		{Opcode: ROpLoadConst, Operands: []int{0, 1}},
		{Opcode: ROpReturn, Operands: []int{0}},
	}
	bytes := regInstructionsToBytes(instrs)
	if len(bytes) == 0 {
		t.Fatal("expected non-empty bytes")
	}
	if RegOpcode(bytes[0]) != ROpLoadConst {
		t.Errorf("expected ROpLoadConst, got %d", bytes[0])
	}
}

func TestHasYieldInBody(t *testing.T) {
	tests := []struct {
		name     string
		node     ast.Node
		expected bool
	}{
		{"yield", &ast.YieldStatement{Expression: &ast.IntegerLiteral{Value: 1}}, true},
		{"block with yield", &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}, true},
		{"block no yield", &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}}, false},
		{"if consequence yield", &ast.IfExpression{Condition: &ast.Boolean{Value: true}, Consequence: &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}}, true},
		{"if alternative yield", &ast.IfExpression{Condition: &ast.Boolean{Value: true}, Consequence: &ast.BlockStatement{}, Alternative: &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}}, true},
		{"while yield", &ast.WhileStatement{Condition: &ast.Boolean{Value: true}, Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}}, true},
		{"function no yield", &ast.FunctionLiteral{Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}}}, false},
		{"pass", &ast.PassStatement{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasYieldInBody(tt.node)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// =============================================================================
// 12. Serialization tests
// =============================================================================

func TestSerializeDeserializeRegBytecode(t *testing.T) {
	rc := newTestRegCompiler()
	rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 42}})
	bc := rc.Bytecode()
	data, err := SerializeRegBytecode(bc)
	if err != nil {
		t.Fatalf("unexpected serialization error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty serialized data")
	}
	decoded, err := DeserializeRegBytecode(data)
	if err != nil {
		t.Fatalf("unexpected deserialization error: %v", err)
	}
	if decoded.NumRegs != bc.NumRegs {
		t.Errorf("expected NumRegs %d, got %d", bc.NumRegs, decoded.NumRegs)
	}
}

func TestSerializeDeserializeConstants(t *testing.T) {
	tests := []struct {
		name string
		obj  objects.Object
	}{
		{"integer", &objects.Integer{Value: 42}},
		{"float", &objects.Float{Value: 3.14}},
		{"string", &objects.String{Value: "hello"}},
		{"boolean true", &objects.Boolean{Value: true}},
		{"boolean false", &objects.Boolean{Value: false}},
		{"null", objects.None_},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := newTestRegCompiler()
			rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}})
			bc := rc.Bytecode()
			bc.Constants = append(bc.Constants, tt.obj)
			data, err := SerializeRegBytecode(bc)
			if err != nil {
				t.Fatalf("serialization error: %v", err)
			}
			decoded, err := DeserializeRegBytecode(data)
			if err != nil {
				t.Fatalf("deserialization error: %v", err)
			}
			lastConst := decoded.Constants[len(decoded.Constants)-1]
			switch expected := tt.obj.(type) {
			case *objects.Integer:
				actual, ok := lastConst.(*objects.Integer)
				if !ok || actual.Value != expected.Value {
					t.Errorf("expected %v, got %v", expected, actual)
				}
			case *objects.Float:
				actual, ok := lastConst.(*objects.Float)
				if !ok || actual.Value != expected.Value {
					t.Errorf("expected %v, got %v", expected, actual)
				}
			case *objects.String:
				actual, ok := lastConst.(*objects.String)
				if !ok || actual.Value != expected.Value {
					t.Errorf("expected %v, got %v", expected, actual)
				}
			case *objects.Boolean:
				actual, ok := lastConst.(*objects.Boolean)
				if !ok || actual.Value != expected.Value {
					t.Errorf("expected %v, got %v", expected, actual)
				}
			case *objects.None:
				if lastConst != objects.None_ {
					t.Errorf("expected None, got %v", lastConst)
				}
			}
		})
	}
}

func TestDeserializeInvalidMagic(t *testing.T) {
	_, err := DeserializeRegBytecode([]byte("XXXX" + string([]byte{0, 1})))
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestDeserializeInvalidVersion(t *testing.T) {
	rc := newTestRegCompiler()
	rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}})
	bc := rc.Bytecode()
	data, _ := SerializeRegBytecode(bc)
	data[4] = 0xFF
	data[5] = 0xFF
	_, err := DeserializeRegBytecode(data)
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestDeserializeTruncatedData(t *testing.T) {
	_, err := DeserializeRegBytecode([]byte("GPYC"))
	if err == nil {
		t.Fatal("expected error for truncated data")
	}
}

func TestSerializeDeserializeFunctionConstant(t *testing.T) {
	rc := newTestRegCompiler()
	rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "testFunc", Parameters: []*ast.Identifier{{Value: "x"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.Identifier{Value: "x"}},
			}},
		},
	})
	bc := rc.Bytecode()
	data, err := SerializeRegBytecode(bc)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	decoded, err := DeserializeRegBytecode(data)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	if decoded.NumRegs != bc.NumRegs {
		t.Errorf("expected NumRegs %d, got %d", bc.NumRegs, decoded.NumRegs)
	}
}

// =============================================================================
// 13. StringBuilder optimization tests
// =============================================================================

func TestFindStringBuilderCandidates(t *testing.T) {
	// Need a local scope variable for StringBuilder detection
	global := NewSymbolTable()
	outer := NewEnclosedSymbolTable(global)
	outer.Define("s") // s is Local in outer
	c := NewWithState(outer, []objects.Object{})
	node := &ast.WhileStatement{
		Condition: &ast.Boolean{Value: true},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "x"}},
		}},
	}
	candidates := c.findStringBuilderCandidates(node)
	if len(candidates) == 0 {
		t.Error("expected to find string builder candidates")
	}
}

func TestFindStringBuilderCandidatesNoMatch(t *testing.T) {
	c := newTestCompiler()
	node := &ast.WhileStatement{
		Condition: &ast.Boolean{Value: true},
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	}
	candidates := c.findStringBuilderCandidates(node)
	if len(candidates) != 0 {
		t.Error("expected no string builder candidates")
	}
}

func TestScanBlockForStringBuildersNil(t *testing.T) {
	st := NewSymbolTable()
	seen := make(map[string]bool)
	scanBlockForStringBuilders(nil, st, seen)
	if len(seen) != 0 {
		t.Error("expected empty seen map for nil block")
	}
}

func TestFindArrayPreallocCandidate(t *testing.T) {
	c := newTestCompiler()
	node := &ast.WhileStatement{
		Condition: &ast.InfixExpression{
			Left: &ast.Identifier{Value: "_i_0"}, Operator: "<", Right: &ast.IntegerLiteral{Value: 10},
		},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.ExpressionStatement{Expression: &ast.CallExpression{
				Function: &ast.MemberAccess{Object: &ast.Identifier{Value: "lst"}, Member: &ast.Identifier{Value: "append"}},
				Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 1}},
			}},
		}},
	}
	info := c.findArrayPreallocCandidate(node)
	if info == nil {
		t.Fatal("expected array prealloc info")
	}
	if info.capacity != 10 {
		t.Errorf("expected capacity 10, got %d", info.capacity)
	}
}

func TestFindArrayPreallocCandidateNoMatch(t *testing.T) {
	c := newTestCompiler()
	node := &ast.WhileStatement{
		Condition: &ast.Boolean{Value: true},
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	}
	info := c.findArrayPreallocCandidate(node)
	if info != nil {
		t.Error("expected nil for non-matching pattern")
	}
}

func TestFindArrayPreallocCandidateNotDesugaredLoop(t *testing.T) {
	c := newTestCompiler()
	node := &ast.WhileStatement{
		Condition: &ast.InfixExpression{
			Left: &ast.Identifier{Value: "x"}, Operator: "<", Right: &ast.IntegerLiteral{Value: 10},
		},
		Body: &ast.BlockStatement{Statements: []ast.Statement{}},
	}
	info := c.findArrayPreallocCandidate(node)
	if info != nil {
		t.Error("expected nil for non-desugared loop")
	}
}

func TestFindArrayPreallocCandidateZeroCapacity(t *testing.T) {
	c := newTestCompiler()
	node := &ast.WhileStatement{
		Condition: &ast.InfixExpression{
			Left: &ast.Identifier{Value: "_i_0"}, Operator: "<", Right: &ast.IntegerLiteral{Value: 0},
		},
		Body: &ast.BlockStatement{Statements: []ast.Statement{}},
	}
	info := c.findArrayPreallocCandidate(node)
	if info != nil {
		t.Error("expected nil for zero capacity")
	}
}

func TestFindAppendVars(t *testing.T) {
	block := &ast.BlockStatement{Statements: []ast.Statement{
		&ast.ExpressionStatement{Expression: &ast.CallExpression{
			Function: &ast.MemberAccess{Object: &ast.Identifier{Value: "lst"}, Member: &ast.Identifier{Value: "append"}},
			Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 1}},
		}},
	}}
	vars := findAppendVars(block)
	if len(vars) != 1 || vars[0] != "lst" {
		t.Errorf("expected ['lst'], got %v", vars)
	}
}

func TestFindAppendVarsNil(t *testing.T) {
	vars := findAppendVars(nil)
	if len(vars) != 0 {
		t.Errorf("expected empty, got %v", vars)
	}
}

func TestCompileWhileWithStringBuilderOptimization(t *testing.T) {
	// Need local scope for StringBuilder optimization
	global := NewSymbolTable()
	outer := NewEnclosedSymbolTable(global)
	outer.Define("s") // s is Local in outer
	c := NewWithState(outer, []objects.Object{})
	err := c.Compile(&ast.WhileStatement{
		Condition: &ast.Boolean{Value: true},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "x"}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bc := c.Bytecode()
	if !hasOpcode(bc.Instructions, OpStringBuilderCreate) {
		t.Error("expected OpStringBuilderCreate in instructions")
	}
}

// =============================================================================
// 14. Read operand test
// =============================================================================

func TestReadOperand(t *testing.T) {
	c := newTestCompiler()
	ins := c.make(OpConstant, 42)
	val, size := readOperand(ins, 0, OpConstant)
	if val != 42 {
		t.Errorf("expected value 42, got %d", val)
	}
	if size != 2 {
		t.Errorf("expected size 2, got %d", size)
	}
}

func TestReadOperandDefault(t *testing.T) {
	val, size := readOperand([]byte{byte(OpAdd)}, 0, OpAdd)
	if val != 0 {
		t.Errorf("expected value 0, got %d", val)
	}
	if size != 0 {
		t.Errorf("expected size 0, got %d", size)
	}
}

// =============================================================================
// 15. RegCompiler compileFunction test
// =============================================================================

func TestRegCompilerCompileFunction(t *testing.T) {
	rc := newTestRegCompiler()
	fn := &ast.FunctionLiteral{
		Name: "test", Parameters: []*ast.Identifier{{Value: "x"}},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.ReturnStatement{ReturnValue: &ast.Identifier{Value: "x"}},
		}},
	}
	compiled := rc.compileFunction(fn)
	if compiled == nil {
		t.Fatal("expected non-nil compiled function")
	}
	if compiled.NumParameters != 1 {
		t.Errorf("expected 1 parameter, got %d", compiled.NumParameters)
	}
}

// =============================================================================
// 16. RegCompiler with Global/Nonlocal in function
// =============================================================================

func TestRegCompilerFunctionWithGlobalNonlocal(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Parameters: []*ast.Identifier{{Value: "x"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.GlobalStatement{Names: []*ast.Identifier{{Value: "g"}}},
				&ast.NonlocalStatement{Names: []*ast.Identifier{{Value: "n"}}},
				&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 17. RegCompiler ClassStatement with metaclass and multi-super
// =============================================================================

func TestRegCompilerClassStatementWithMetaclass(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("Meta")
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"}, Metaclass: &ast.Identifier{Value: "Meta"},
		Body: &ast.BlockStatement{}, Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerClassStatementWithMultiSuper(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("Base1")
	rc.symbolTable.Define("Base2")
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "Child"}, SuperClasses: []*ast.Identifier{{Value: "Base1"}, {Value: "Base2"}},
		Body: &ast.BlockStatement{}, Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 18. RegCompiler TryStatement with finally and star
// =============================================================================

func TestRegCompilerTryStatementWithFinally(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.TryStatement{
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		Excepts: []*ast.ExceptClause{},
		Finally: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerTryStatementWithStarExcept(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.TryStatement{
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		Excepts: []*ast.ExceptClause{{
			Type: &ast.Identifier{Value: "ExceptionGroup"}, IsStar: true,
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 19. RegCompiler FromImport with alias
// =============================================================================

func TestRegCompilerFromImportWithAlias(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.FromImportStatement{Module: &ast.Identifier{Value: "math"}, Alias: &ast.Identifier{Value: "m"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 20. RegCompiler compileExpr nil node
// =============================================================================

func TestRegCompilerCompileExprNil(t *testing.T) {
	rc := newTestRegCompiler()
	reg, err := rc.compileExpr(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg < 0 {
		t.Errorf("expected non-negative register, got %d", reg)
	}
}

// =============================================================================
// 21. RegCompiler complex literal invalid
// =============================================================================

func TestRegCompilerComplexLiteralInvalid(t *testing.T) {
	rc := newTestRegCompiler()
	_, err := rc.compileExpr(&ast.ComplexLiteral{Value: "notanumberj"})
	if err == nil {
		t.Fatal("expected error for invalid complex literal")
	}
}

// =============================================================================
// 22. RegCompiler with statement no name
// =============================================================================

func TestRegCompilerWithStatementNoName(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("open")
	err := rc.Compile(&ast.WithStatement{
		Items: []*ast.ContextManagerItem{{Expr: &ast.Identifier{Value: "open"}, Name: nil}},
		Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 23. Slice assign with nil bounds
// =============================================================================

func TestCompileSliceAssignNilBounds(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.SliceAssignStatement{
		Left: &ast.Identifier{Value: "lst"},
		Value: &ast.ListLiteral{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 1}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerSliceAssignNilBounds(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.SliceAssignStatement{
		Left: &ast.Identifier{Value: "lst"},
		Value: &ast.ListLiteral{Elements: []ast.Expression{&ast.IntegerLiteral{Value: 1}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 24. Slice expression with nil bounds
// =============================================================================

func TestCompileSliceExpressionNilBounds(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("lst")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{
			Left:  &ast.Identifier{Value: "lst"},
			Index: &ast.SliceExpression{},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerSliceExpressionNilBounds(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{
			Left:  &ast.Identifier{Value: "lst"},
			Index: &ast.SliceExpression{},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 25. Compiler hasYieldInBody tests (stack-based)
// =============================================================================

func TestCompilerHasYieldInBody(t *testing.T) {
	c := newTestCompiler()
	tests := []struct {
		name     string
		node     ast.Node
		expected bool
	}{
		{"yield", &ast.YieldStatement{Expression: &ast.IntegerLiteral{Value: 1}}, true},
		{"block with yield", &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}, true},
		{"block no yield", &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}}, false},
		{"if consequence yield", &ast.IfExpression{Condition: &ast.Boolean{Value: true}, Consequence: &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}}, true},
		{"if alternative yield", &ast.IfExpression{Condition: &ast.Boolean{Value: true}, Consequence: &ast.BlockStatement{}, Alternative: &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}}, true},
		{"while yield", &ast.WhileStatement{Condition: &ast.Boolean{Value: true}, Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.YieldStatement{}}}}, true},
		{"function no yield", &ast.FunctionLiteral{Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}}}, false},
		{"pass", &ast.PassStatement{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.hasYieldInBody(tt.node)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// =============================================================================
// Additional Coverage: compileListComprehension
// =============================================================================

func TestCompileListComprehension(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.ListComprehension{
			Token:    "[",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileListComprehensionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.ListComprehension{
			Token:    "[",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Additional Coverage: compileSetComprehension
// =============================================================================

func TestCompileSetComprehension(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.SetComprehension{
			Token:    "{",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Additional Coverage: compileDictComprehension
// =============================================================================

func TestCompileDictComprehension(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("k")
	c.symbolTable.Define("v")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.DictComprehension{
			Token:    "{",
			Key:      &ast.Identifier{Value: "k"},
			Value:    &ast.Identifier{Value: "v"},
			Variable: &ast.Identifier{Value: "k"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileDictComprehensionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("k")
	c.symbolTable.Define("v")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.DictComprehension{
			Token:    "{",
			Key:      &ast.Identifier{Value: "k"},
			Value:    &ast.Identifier{Value: "v"},
			Variable: &ast.Identifier{Value: "k"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Additional Coverage: compileAsyncListComprehension
// =============================================================================

func TestCompileAsyncListComprehension(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncListComprehension{
			Token:    "[",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileAsyncListComprehensionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncListComprehension{
			Token:    "[",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Additional Coverage: compileAsyncSetComprehension
// =============================================================================

func TestCompileAsyncSetComprehension(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncSetComprehension{
			Token:    "{",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Additional Coverage: compileAsyncDictComprehension
// =============================================================================

func TestCompileAsyncDictComprehension(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("k")
	c.symbolTable.Define("v")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncDictComprehension{
			Token:    "{",
			Key:      &ast.Identifier{Value: "k"},
			Value:    &ast.Identifier{Value: "v"},
			Variable: &ast.Identifier{Value: "k"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Additional Coverage: compileAsyncGeneratorExpression
// =============================================================================

func TestCompileAsyncGeneratorExpression(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncGeneratorExpression{
			Token:    "(",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileAsyncGeneratorExpressionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncGeneratorExpression{
			Token:    "(",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSymbolTableGlobalInFunction(t *testing.T) {
	global := NewSymbolTable()
	global.Define("x")
	fn := NewEnclosedSymbolTable(global)
	fn.DefineGlobal("x")
	sym := fn.Define("x")
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope, got %v", sym.Scope)
	}
}

func TestSymbolTableGlobalExistingInOuter(t *testing.T) {
	global := NewSymbolTable()
	global.Define("x")
	fn := NewEnclosedSymbolTable(global)
	fn.DefineGlobal("x")
	sym := fn.Define("x")
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope, got %v", sym.Scope)
	}
}

// =============================================================================
// Additional Coverage: CompiledFunction methods
// =============================================================================

func TestCompiledFunctionTypeAndInspect(t *testing.T) {
	cf := &CompiledFunction{}
	if cf.Type() != objects.FUNCTION_OBJ {
		t.Errorf("expected FUNCTION_OBJ, got %v", cf.Type())
	}
	if cf.Inspect() != "compiled function" {
		t.Errorf("expected 'compiled function', got %s", cf.Inspect())
	}
}

// =============================================================================
// Additional Coverage: scanStmtForStringBuilders
// =============================================================================

func TestScanStmtForStringBuilders(t *testing.T) {
	outer := NewSymbolTable()
	st := NewEnclosedSymbolTable(outer)
	st.Define("s") // LocalScope in enclosed table
	seen := make(map[string]bool)

	// Test AugAssignStatement with += operator
	augAssign := &ast.AugAssignStatement{
		Name:     &ast.Identifier{Value: "s"},
		Operator: "+",
		Value:    &ast.StringLiteral{Value: "hello"},
	}
	scanStmtForStringBuilders(augAssign, st, seen)
	if !seen["s"] {
		t.Error("expected 's' to be a string builder candidate")
	}
}

func TestScanStmtForStringBuildersNonString(t *testing.T) {
	outer := NewSymbolTable()
	st := NewEnclosedSymbolTable(outer)
	st.Define("n") // LocalScope in enclosed table
	seen := make(map[string]bool)

	// Test AugAssignStatement with += but non-local variable
	augAssign := &ast.AugAssignStatement{
		Name:     &ast.Identifier{Value: "n"},
		Operator: "+",
		Value:    &ast.IntegerLiteral{Value: 1},
	}
	scanStmtForStringBuilders(augAssign, st, seen)
	// n is local, so it should be a candidate
	if !seen["n"] {
		t.Error("expected 'n' to be a candidate")
	}
}

// =============================================================================
// Additional Coverage: Serialization
// =============================================================================

func TestSerializeRegBytecode(t *testing.T) {
	rc := newTestRegCompiler()
	prog := &ast.Program{Statements: []ast.Statement{
		&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 42}},
	}}
	err := rc.Compile(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := SerializeRegBytecode(rc.Bytecode())
	if err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty serialized data")
	}
}

// =============================================================================
// 26. Additional coverage: isTruthy with Dict and Set
// =============================================================================

func TestIsTruthyDict(t *testing.T) {
	emptyDict := objects.NewDict()
	if isTruthy(emptyDict) {
		t.Error("expected empty dict to be falsy")
	}
	nonEmptyDict := objects.NewDict()
	nonEmptyDict.Pairs["key"] = &objects.Integer{Value: 1}
	nonEmptyDict.Keys["key"] = &objects.String{Value: "key"}
	nonEmptyDict.KeyOrder = append(nonEmptyDict.KeyOrder, "key")
	if !isTruthy(nonEmptyDict) {
		t.Error("expected non-empty dict to be truthy")
	}
}

func TestIsTruthySet(t *testing.T) {
	emptySet := objects.NewSet()
	if isTruthy(emptySet) {
		t.Error("expected empty set to be falsy")
	}
	nonEmptySet := objects.NewSet()
	nonEmptySet.Elements["1"] = &objects.Integer{Value: 1}
	nonEmptySet.Keys["1"] = &objects.Integer{Value: 1}
	if !isTruthy(nonEmptySet) {
		t.Error("expected non-empty set to be truthy")
	}
}

func TestIsTruthyDefault(t *testing.T) {
	// Unknown type should default to true
	if !isTruthy(&CompiledFunction{}) {
		t.Error("expected unknown object type to be truthy")
	}
}

// =============================================================================
// 27. Additional coverage: compareObjects more cases
// =============================================================================

func TestCompareObjectsFloatGreater(t *testing.T) {
	result := compareObjects(&objects.Float{Value: 2.0}, &objects.Float{Value: 1.0})
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestCompareObjectsTypeMismatch(t *testing.T) {
	result := compareObjects(&objects.Integer{Value: 1}, &objects.Float{Value: 1.0})
	if result != 0 {
		t.Errorf("expected 0 for type mismatch, got %d", result)
	}
}

func TestCompareObjectsStringGreater(t *testing.T) {
	result := compareObjects(&objects.String{Value: "b"}, &objects.String{Value: "a"})
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestCompareObjectsUnsupportedType(t *testing.T) {
	result := compareObjects(&objects.Boolean{Value: true}, &objects.Boolean{Value: false})
	if result != 0 {
		t.Errorf("expected 0 for unsupported type, got %d", result)
	}
}

// =============================================================================
// 28. Additional coverage: scanStmtForStringBuilders more cases
// =============================================================================

func TestScanStmtForStringBuildersWhileStatement(t *testing.T) {
	outer := NewSymbolTable()
	st := NewEnclosedSymbolTable(outer)
	st.Define("s")
	seen := make(map[string]bool)

	// WhileStatement should not recurse into nested loops
	scanStmtForStringBuilders(&ast.WhileStatement{
		Condition: &ast.Boolean{Value: true},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "x"}},
		}},
	}, st, seen)
	if seen["s"] {
		t.Error("expected WhileStatement to not recurse into nested loops")
	}
}

func TestScanStmtForStringBuildersExpressionStatementIf(t *testing.T) {
	outer := NewSymbolTable()
	st := NewEnclosedSymbolTable(outer)
	st.Define("s")
	seen := make(map[string]bool)

	// ExpressionStatement with IfExpression
	scanStmtForStringBuilders(&ast.ExpressionStatement{
		Expression: &ast.IfExpression{
			Condition: &ast.Boolean{Value: true},
			Consequence: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "x"}},
			}},
			Alternative: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "y"}},
			}},
		},
	}, st, seen)
	if !seen["s"] {
		t.Error("expected 's' to be found in if expression branches")
	}
}

func TestScanStmtForStringBuildersTryStatement(t *testing.T) {
	outer := NewSymbolTable()
	st := NewEnclosedSymbolTable(outer)
	st.Define("s")
	seen := make(map[string]bool)

	scanStmtForStringBuilders(&ast.TryStatement{
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "x"}},
		}},
		Excepts: []*ast.ExceptClause{{
			Type: &ast.Identifier{Value: "ValueError"},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "e"}},
			}},
		}},
		Finally: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "f"}},
		}},
	}, st, seen)
	if !seen["s"] {
		t.Error("expected 's' to be found in try statement")
	}
}

func TestScanStmtForStringBuildersBlockStatement(t *testing.T) {
	outer := NewSymbolTable()
	st := NewEnclosedSymbolTable(outer)
	st.Define("s")
	seen := make(map[string]bool)

	scanStmtForStringBuilders(&ast.BlockStatement{Statements: []ast.Statement{
		&ast.AugAssignStatement{Name: &ast.Identifier{Value: "s"}, Operator: "+", Value: &ast.StringLiteral{Value: "x"}},
	}}, st, seen)
	if !seen["s"] {
		t.Error("expected 's' to be found in block statement")
	}
}

// =============================================================================
// 29. Additional coverage: readOperand more opcode cases
// =============================================================================

func TestReadOperandClosure(t *testing.T) {
	c := newTestCompiler()
	ins := c.make(OpClosure, 42)
	// OpClosure: 1 byte opcode + 2 bytes operand + 1 byte numFree
	val, size := readOperand(ins, 0, OpClosure)
	if val != 42 {
		t.Errorf("expected value 42, got %d", val)
	}
	if size != 3 {
		t.Errorf("expected size 3, got %d", size)
	}
}

func TestReadOperandBeginTry(t *testing.T) {
	ins := make([]byte, 9)
	ins[0] = byte(OpBeginTry)
	ins[1] = 0
	ins[2] = 5 // main operand
	ins[3] = 0
	ins[4] = 0
	ins[5] = 0
	ins[6] = 10 // handler IP
	ins[7] = 0
	ins[8] = 20 // finally IP
	val, size := readOperand(ins, 0, OpBeginTry)
	if val != 5 {
		t.Errorf("expected value 5, got %d", val)
	}
	if size != 8 {
		t.Errorf("expected size 8, got %d", size)
	}
}

func TestReadOperandExceptHandler(t *testing.T) {
	ins := make([]byte, 5)
	ins[0] = byte(OpExceptHandler)
	ins[1] = 0
	ins[2] = 42
	ins[3] = 0
	ins[4] = 0
	val, size := readOperand(ins, 0, OpExceptHandler)
	if val != 42 {
		t.Errorf("expected value 42, got %d", val)
	}
	if size != 4 {
		t.Errorf("expected size 4, got %d", size)
	}
}

func TestReadOperandCall(t *testing.T) {
	ins := make([]byte, 2)
	ins[0] = byte(OpCall)
	ins[1] = 3
	val, size := readOperand(ins, 0, OpCall)
	if val != 3 {
		t.Errorf("expected value 3, got %d", val)
	}
	if size != 1 {
		t.Errorf("expected size 1, got %d", size)
	}
}

func TestReadOperandSetLocal(t *testing.T) {
	ins := make([]byte, 2)
	ins[0] = byte(OpSetLocal)
	ins[1] = 5
	val, size := readOperand(ins, 0, OpSetLocal)
	if val != 5 {
		t.Errorf("expected value 5, got %d", val)
	}
	if size != 1 {
		t.Errorf("expected size 1, got %d", size)
	}
}

func TestReadOperandGetLocal(t *testing.T) {
	ins := make([]byte, 2)
	ins[0] = byte(OpGetLocal)
	ins[1] = 7
	val, size := readOperand(ins, 0, OpGetLocal)
	if val != 7 {
		t.Errorf("expected value 7, got %d", val)
	}
	if size != 1 {
		t.Errorf("expected size 1, got %d", size)
	}
}

func TestReadOperandGetFree(t *testing.T) {
	ins := make([]byte, 2)
	ins[0] = byte(OpGetFree)
	ins[1] = 2
	val, size := readOperand(ins, 0, OpGetFree)
	if val != 2 {
		t.Errorf("expected value 2, got %d", val)
	}
	if size != 1 {
		t.Errorf("expected size 1, got %d", size)
	}
}

// =============================================================================
// 30. Additional coverage: analyzeEscape
// =============================================================================

func TestAnalyzeEscapeNoLocals(t *testing.T) {
	fn := &CompiledFunction{NumLocals: 0}
	result := analyzeEscape(fn)
	if result != nil {
		t.Error("expected nil for function with no locals")
	}
}

func TestAnalyzeEscapeSetGlobal(t *testing.T) {
	// OpGetLocal followed by OpSetGlobal should mark local as escaping
	ins := []byte{
		byte(OpGetLocal), 0,  // load local 0
		byte(OpSetGlobal), 0, 0, // set global 0
		byte(OpReturn),
	}
	fn := &CompiledFunction{NumLocals: 1, Instructions: ins}
	result := analyzeEscape(fn)
	if result[0] {
		t.Error("expected local 0 to be escaping (assigned to global)")
	}
}

func TestAnalyzeEscapeSetAttribute(t *testing.T) {
	ins := []byte{
		byte(OpGetLocal), 0, // load local 0
		byte(OpSetAttribute), 0, 0, // set attribute
		byte(OpReturn),
	}
	fn := &CompiledFunction{NumLocals: 1, Instructions: ins}
	result := analyzeEscape(fn)
	if result[0] {
		t.Error("expected local 0 to be escaping (assigned to attribute)")
	}
}

func TestAnalyzeEscapeReturnValue(t *testing.T) {
	ins := []byte{
		byte(OpGetLocal), 0, // load local 0
		byte(OpReturnValue),
	}
	fn := &CompiledFunction{NumLocals: 1, Instructions: ins}
	result := analyzeEscape(fn)
	if result[0] {
		t.Error("expected local 0 to be escaping (returned)")
	}
}

func TestAnalyzeEscapeCall(t *testing.T) {
	ins := []byte{
		byte(OpGetLocal), 0, // load local 0
		byte(OpCall), 1, // call with 1 arg
		byte(OpReturn),
	}
	fn := &CompiledFunction{NumLocals: 1, Instructions: ins}
	result := analyzeEscape(fn)
	if result[0] {
		t.Error("expected local 0 to be escaping (passed as arg)")
	}
}

func TestAnalyzeEscapeClosure(t *testing.T) {
	ins := []byte{
		byte(OpGetLocal), 0, // load local 0
		byte(OpClosure), 0, 0, 0, // create closure
		byte(OpReturn),
	}
	fn := &CompiledFunction{NumLocals: 1, Instructions: ins}
	result := analyzeEscape(fn)
	if result[0] {
		t.Error("expected local 0 to be escaping (captured in closure)")
	}
}

func TestAnalyzeEscapeGetFree(t *testing.T) {
	ins := []byte{
		byte(OpGetFree), 0, // load free var 0
		byte(OpReturn),
	}
	fn := &CompiledFunction{NumLocals: 1, Instructions: ins}
	result := analyzeEscape(fn)
	if result[0] {
		t.Error("expected local 0 to be escaping (free var reference)")
	}
}

func TestAnalyzeEscapeNonEscaping(t *testing.T) {
	// Local that is only used locally (OpSetLocal + OpGetLocal)
	ins := []byte{
		byte(OpConstant), 0, 0, // load constant
		byte(OpSetLocal), 0, // set local 0
		byte(OpGetLocal), 0, // get local 0
		byte(OpPop),
		byte(OpReturn),
	}
	fn := &CompiledFunction{NumLocals: 1, Instructions: ins}
	result := analyzeEscape(fn)
	if !result[0] {
		t.Error("expected local 0 to be non-escaping")
	}
}

// =============================================================================
// 31. Additional coverage: Serialization - CompiledFunction and unsupported types
// =============================================================================

func TestSerializeDeserializeCompiledFunctionConstant(t *testing.T) {
	fn := &CompiledFunction{
		Instructions:  []byte{byte(OpReturn)},
		NumLocals:     2,
		NumParameters: 1,
		IsGenerator:   true,
		IsAsync:       true,
	}
	rc := newTestRegCompiler()
	rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}})
	bc := rc.Bytecode()
	bc.Constants = append(bc.Constants, fn)
	data, err := SerializeRegBytecode(bc)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	decoded, err := DeserializeRegBytecode(data)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	lastConst := decoded.Constants[len(decoded.Constants)-1]
	decodedFn, ok := lastConst.(*CompiledFunction)
	if !ok {
		t.Fatal("expected CompiledFunction constant")
	}
	if decodedFn.NumLocals != 2 {
		t.Errorf("expected NumLocals 2, got %d", decodedFn.NumLocals)
	}
	if decodedFn.NumParameters != 1 {
		t.Errorf("expected NumParameters 1, got %d", decodedFn.NumParameters)
	}
	if !decodedFn.IsGenerator {
		t.Error("expected IsGenerator to be true")
	}
	if !decodedFn.IsAsync {
		t.Error("expected IsAsync to be true")
	}
}

func TestSerializeUnsupportedConstantType(t *testing.T) {
	rc := newTestRegCompiler()
	rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}})
	bc := rc.Bytecode()
	// Add an unsupported type (e.g., a Builtin) - should serialize as Null
	bc.Constants = append(bc.Constants, &objects.Builtin{Fn: func(...objects.Object) objects.Object { return nil }})
	data, err := SerializeRegBytecode(bc)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	decoded, err := DeserializeRegBytecode(data)
	if err != nil {
		t.Fatalf("deserialization error: %v", err)
	}
	// The unsupported type should have been serialized as Null
	lastConst := decoded.Constants[len(decoded.Constants)-1]
	if lastConst != objects.None_ {
		t.Errorf("expected None for unsupported type, got %T", lastConst)
	}
}

func TestDeserializeUnknownConstantTag(t *testing.T) {
	rc := newTestRegCompiler()
	rc.Compile(&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}})
	bc := rc.Bytecode()
	data, err := SerializeRegBytecode(bc)
	if err != nil {
		t.Fatalf("serialization error: %v", err)
	}
	// Corrupt a constant tag to an unknown value
	// Find the first constant tag after the header (4 magic + 2 version + 2 numRegs + 2 numConstants = 10 bytes)
	if len(data) > 10 {
		data[10] = 0xFF // invalid constant tag
		_, err := DeserializeRegBytecode(data)
		if err == nil {
			t.Error("expected error for unknown constant tag")
		}
	}
}

// =============================================================================
// 32. Additional coverage: SymbolTable - global in function scope without outer
// =============================================================================

func TestSymbolTableGlobalInGlobalScope(t *testing.T) {
	// DefineGlobal in global scope (no outer)
	st := NewSymbolTable()
	st.DefineGlobal("g")
	sym := st.Define("g")
	if sym.Scope != GlobalScope {
		t.Errorf("expected GlobalScope, got %v", sym.Scope)
	}
}

func TestSymbolTableNonlocalNotFound(t *testing.T) {
	// Nonlocal where the variable is not found in any outer scope
	global := NewSymbolTable()
	inner := NewEnclosedSymbolTable(global)
	inner.DefineNonlocal("x")
	sym := inner.Define("x")
	// When nonlocal can't find the variable in outer scopes, it falls through
	// to the normal definition path
	if sym.Scope != LocalScope {
		t.Errorf("expected LocalScope when nonlocal not found, got %v", sym.Scope)
	}
}

func TestSymbolTableResolveFreeScopePropagation(t *testing.T) {
	// Test that FreeScope variables propagate through multiple levels
	global := NewSymbolTable()
	outer := NewEnclosedSymbolTable(global)
	outer.Define("x") // x is Local in outer
	middle := NewEnclosedSymbolTable(outer)
	inner := NewEnclosedSymbolTable(middle)

	// Resolve in middle - should be Free
	sym, ok := middle.Resolve("x")
	if !ok {
		t.Fatal("expected to resolve 'x' in middle scope")
	}
	if sym.Scope != FreeScope {
		t.Errorf("expected FreeScope in middle, got %v", sym.Scope)
	}

	// Resolve in inner - should also be Free (propagated)
	sym2, ok := inner.Resolve("x")
	if !ok {
		t.Fatal("expected to resolve 'x' in inner scope")
	}
	if sym2.Scope != FreeScope {
		t.Errorf("expected FreeScope in inner, got %v", sym2.Scope)
	}
}

// =============================================================================
// 33. Additional coverage: compileClassStatement with decorators
// =============================================================================

func TestCompileClassStatementWithStaticMethod(t *testing.T) {
	bc, err := compileProgram(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "static_method", Parameters: []*ast.Identifier{},
			Decorators: []ast.Expression{&ast.Identifier{Value: "staticmethod"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCreateClass) {
		t.Error("expected OpCreateClass in instructions")
	}
}

func TestCompileClassStatementWithClassMethod(t *testing.T) {
	bc, err := compileProgram(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "class_method", Parameters: []*ast.Identifier{{Value: "cls"}},
			Decorators: []ast.Expression{&ast.Identifier{Value: "classmethod"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCreateClass) {
		t.Error("expected OpCreateClass in instructions")
	}
}

func TestCompileClassStatementWithProperty(t *testing.T) {
	bc, err := compileProgram(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "prop", Parameters: []*ast.Identifier{{Value: "self"}},
			Decorators: []ast.Expression{&ast.Identifier{Value: "property"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 42}}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCreateClass) {
		t.Error("expected OpCreateClass in instructions")
	}
}

func TestCompileClassStatementWithPropertySetter(t *testing.T) {
	bc, err := compileProgram(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{
			{
				Name: "prop", Parameters: []*ast.Identifier{{Value: "self"}},
				Decorators: []ast.Expression{&ast.Identifier{Value: "property"}},
				Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
			},
			{
				Name: "prop", Parameters: []*ast.Identifier{{Value: "self"}, {Value: "value"}},
				Decorators: []ast.Expression{&ast.MemberAccess{
					Object: &ast.Identifier{Value: "prop"}, Member: &ast.Identifier{Value: "setter"},
				}},
				Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCreateClass) {
		t.Error("expected OpCreateClass in instructions")
	}
}

func TestCompileClassStatementWithPropertyDeleter(t *testing.T) {
	bc, err := compileProgram(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{
			{
				Name: "prop", Parameters: []*ast.Identifier{{Value: "self"}},
				Decorators: []ast.Expression{&ast.Identifier{Value: "property"}},
				Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
			},
			{
				Name: "prop", Parameters: []*ast.Identifier{{Value: "self"}},
				Decorators: []ast.Expression{&ast.MemberAccess{
					Object: &ast.Identifier{Value: "prop"}, Member: &ast.Identifier{Value: "deleter"},
				}},
				Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasOpcode(bc.Instructions, OpCreateClass) {
		t.Error("expected OpCreateClass in instructions")
	}
}

// =============================================================================
// 34. Additional coverage: compileSetComprehension with filter
// =============================================================================

func TestCompileSetComprehensionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.SetComprehension{
			Token:    "{",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 35. Additional coverage: compileAsyncSetComprehension with filter
// =============================================================================

func TestCompileAsyncSetComprehensionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncSetComprehension{
			Token:    "{",
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 36. Additional coverage: compileAsyncDictComprehension with filter
// =============================================================================

func TestCompileAsyncDictComprehensionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("k")
	c.symbolTable.Define("v")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.AsyncDictComprehension{
			Token:    "{",
			Key:      &ast.Identifier{Value: "k"},
			Value:    &ast.Identifier{Value: "v"},
			Variable: &ast.Identifier{Value: "k"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 37. Additional coverage: EliminateDeadCode with try/finally
// =============================================================================

func TestEliminateDeadCodeWithJumpNotTruthy(t *testing.T) {
	c := newTestCompiler()
	// Create instructions: OpTrue, OpJumpNotTruthy(5), OpConstant(0), OpJump(0)
	ins := make([]byte, 0)
	ins = append(ins, byte(OpTrue))
	jumpPos := len(ins)
	ins = append(ins, c.make(OpJumpNotTruthy, 0)...)
	ins = append(ins, c.make(OpConstant, 0)...)
	// Patch the jump target to skip over the constant
	target := len(ins)
	ins[jumpPos+1] = byte(target >> 8)
	ins[jumpPos+2] = byte(target & 0xFF)

	result := EliminateDeadCode(ins)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestEliminateDeadCodeWithBeginTry(t *testing.T) {
	// Build a minimal try block
	ins := make([]byte, 0)
	ins = append(ins, byte(OpBeginTry))
	tryPos := len(ins) - 1
	// Pad to 9 bytes total for OpBeginTry
	for len(ins) < tryPos+9 {
		ins = append(ins, 0)
	}
	// Set handler IP and finally IP
	ins[tryPos+5] = 0
	ins[tryPos+6] = 0
	ins[tryPos+7] = 0
	ins[tryPos+8] = 0
	ins = append(ins, byte(OpReturn))

	result := EliminateDeadCode(ins)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestEliminateDeadCodeWithFinally(t *testing.T) {
	c := newTestCompiler()
	ins := make([]byte, 0)
	ins = append(ins, c.make(OpFinally, 0)...)
	ins = append(ins, byte(OpReturn))

	result := EliminateDeadCode(ins)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// =============================================================================
// 38. Additional coverage: Register compiler - Identifier with FreeScope
// =============================================================================

func TestRegCompilerIdentifierFreeScope(t *testing.T) {
	rc := newTestRegCompiler()
	// Create a scope chain: global -> outer (with x) -> inner
	// x defined in outer becomes Free when resolved from inner
	outer := NewEnclosedSymbolTable(rc.symbolTable)
	outer.Define("x") // x is Local in outer
	inner := NewEnclosedSymbolTable(outer)
	rc.symbolTable = inner
	// Resolve x from inner scope - it should become Free
	sym, ok := rc.symbolTable.Resolve("x")
	if !ok {
		t.Fatal("expected to resolve 'x'")
	}
	if sym.Scope != FreeScope {
		t.Errorf("expected FreeScope, got %v", sym.Scope)
	}
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.Identifier{Value: "x"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpGetFree) {
		t.Error("expected ROpGetFree for free variable")
	}
}

func TestRegCompilerIdentifierLocalScope(t *testing.T) {
	rc := newTestRegCompiler()
	// Create a symbol with LocalScope by using enclosed table
	outer := NewEnclosedSymbolTable(rc.symbolTable)
	outer.Define("local_var")
	rc.symbolTable = outer
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.Identifier{Value: "local_var"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpGetLocal) {
		t.Error("expected ROpGetLocal for local variable")
	}
}

// =============================================================================
// 39. Additional coverage: Register compiler - FunctionLiteral with free vars
// =============================================================================

func TestRegCompilerFunctionWithFreeVars(t *testing.T) {
	rc := newTestRegCompiler()
	// Define outer_var in an enclosed scope so it becomes LocalScope
	// Then the inner function will see it as FreeScope
	outer := NewEnclosedSymbolTable(rc.symbolTable)
	outer.Define("outer_var") // outer_var is Local in this scope
	rc.symbolTable = outer
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "inner", Parameters: []*ast.Identifier{},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ExpressionStatement{Expression: &ast.Identifier{Value: "outer_var"}},
				&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpClosure) {
		t.Error("expected ROpClosure for function with free variables")
	}
}

// =============================================================================
// 40. Additional coverage: Register compiler - Identifier None
// =============================================================================

func TestRegCompilerIdentifierNone(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.Identifier{Value: "None"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpNull) {
		t.Error("expected ROpNull for None identifier")
	}
}

func TestRegCompilerIdentifierFalse(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.Identifier{Value: "False"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpFalse) {
		t.Error("expected ROpFalse for False identifier")
	}
}

// =============================================================================
// 41. Additional coverage: Register compiler - compileExpr default case
// =============================================================================

func TestRegCompilerCompileExprDefault(t *testing.T) {
	rc := newTestRegCompiler()
	// Use an AST node that doesn't have a specific handler in compileExpr
	// CaseClause is a statement that shouldn't be handled in compileExpr
	reg, err := rc.compileExpr(&ast.CaseClause{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg < 0 {
		t.Errorf("expected non-negative register, got %d", reg)
	}
}

// =============================================================================
// 42. Additional coverage: Register compiler - FunctionLiteral with VarArgs/KwArgs
// =============================================================================

func TestRegCompilerFunctionWithVarArgs(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name:       "varfunc",
			Parameters: []*ast.Identifier{{Value: "args"}},
			VarArgs:    &ast.Identifier{Value: "args"},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerFunctionWithKwArgs(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name:       "kwfunc",
			Parameters: []*ast.Identifier{{Value: "kwargs"}},
			KwArgs:     &ast.Identifier{Value: "kwargs"},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 43. Additional coverage: Register compiler - compileStmt more cases
// =============================================================================

func TestRegCompilerDeleteStatementIndex(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.DeleteStatement{
		Targets: []ast.Expression{
			&ast.IndexExpression{Left: &ast.Identifier{Value: "lst"}, Index: &ast.IntegerLiteral{Value: 0}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerAugAssignIndexStatement(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("d")
	err := rc.Compile(&ast.AugAssignStatement{
		IndexLeft: &ast.Identifier{Value: "d"}, IndexIndex: &ast.StringLiteral{Value: "key"},
		Operator: "+", Value: &ast.IntegerLiteral{Value: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 44. Additional coverage: Register compiler - AssignStatement with local scope
// =============================================================================

func TestRegCompilerAssignStatementLocal(t *testing.T) {
	rc := newTestRegCompiler()
	// Create a local scope by entering an enclosed symbol table
	outer := rc.symbolTable
	rc.symbolTable = NewEnclosedSymbolTable(outer)
	rc.symbolTable.Define("x")
	err := rc.Compile(&ast.AssignStatement{
		Names: []*ast.Identifier{{Value: "x"}}, Value: &ast.IntegerLiteral{Value: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetLocal) {
		t.Error("expected ROpSetLocal for local variable assignment")
	}
}

// =============================================================================
// 45. Additional coverage: Register compiler - LetStatement with local scope
// =============================================================================

func TestRegCompilerLetStatementLocal(t *testing.T) {
	rc := newTestRegCompiler()
	outer := rc.symbolTable
	rc.symbolTable = NewEnclosedSymbolTable(outer)
	err := rc.Compile(&ast.LetStatement{
		Names: []*ast.Identifier{{Value: "y"}}, Value: &ast.IntegerLiteral{Value: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetLocal) {
		t.Error("expected ROpSetLocal for local variable in enclosed scope")
	}
}

// =============================================================================
// 46. Additional coverage: Register compiler - ClassStatement with body assign
// =============================================================================

func TestRegCompilerClassStatementWithBodyAssign(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{Statements: []ast.Statement{
			&ast.AssignStatement{Names: []*ast.Identifier{{Value: "class_var"}}, Value: &ast.IntegerLiteral{Value: 42}},
		}},
		Methods: []*ast.FunctionLiteral{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 47. Additional coverage: Register compiler - ClassStatement with decorators
// =============================================================================

func TestRegCompilerClassStatementWithStaticMethod(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "static_method", Parameters: []*ast.Identifier{},
			Decorators: []ast.Expression{&ast.Identifier{Value: "staticmethod"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerClassStatementWithClassMethod(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "class_method", Parameters: []*ast.Identifier{{Value: "cls"}},
			Decorators: []ast.Expression{&ast.Identifier{Value: "classmethod"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegCompilerClassStatementWithProperty(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ClassStatement{
		Name: &ast.Identifier{Value: "MyClass"},
		Body: &ast.BlockStatement{},
		Methods: []*ast.FunctionLiteral{{
			Name: "prop", Parameters: []*ast.Identifier{{Value: "self"}},
			Decorators: []ast.Expression{&ast.Identifier{Value: "property"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.PassStatement{}}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 48. Additional coverage: Register compiler - FunctionLiteral named with local scope
// =============================================================================

func TestRegCompilerFunctionNamedLocalScope(t *testing.T) {
	rc := newTestRegCompiler()
	outer := rc.symbolTable
	rc.symbolTable = NewEnclosedSymbolTable(outer)
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "localFunc", Parameters: []*ast.Identifier{{Value: "x"}},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.Identifier{Value: "x"}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 49. Additional coverage: Register compiler - slice with step
// =============================================================================

func TestRegCompilerSliceWithStep(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("lst")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.IndexExpression{
			Left: &ast.Identifier{Value: "lst"},
			Index: &ast.SliceExpression{
				Lower: &ast.IntegerLiteral{Value: 0}, Upper: &ast.IntegerLiteral{Value: 10}, Step: &ast.IntegerLiteral{Value: 2},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSlice) {
		t.Error("expected ROpSlice")
	}
}

// =============================================================================
// 50. Additional coverage: Register compiler - ImportStatement with alias
// =============================================================================

func TestRegCompilerImportStatementWithAlias(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.ImportStatement{
		Module: &ast.Identifier{Value: "math"}, Alias: &ast.Identifier{Value: "m"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 52. Additional coverage: GeneratorExpression
// =============================================================================

func TestCompileGeneratorExpression(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.GeneratorExpression{
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileGeneratorExpressionWithFilter(t *testing.T) {
	c := newTestCompiler()
	c.symbolTable.Define("x")
	c.symbolTable.Define("items")
	err := c.Compile(&ast.ExpressionStatement{
		Expression: &ast.GeneratorExpression{
			Element:  &ast.Identifier{Value: "x"},
			Variable: &ast.Identifier{Value: "x"},
			Iterable: &ast.Identifier{Value: "items"},
			Filter:   &ast.Boolean{Value: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 53. Additional coverage: Register compiler - compileStmt Program
// =============================================================================

func TestRegCompilerProgram(t *testing.T) {
	rc := newTestRegCompiler()
	err := rc.Compile(&ast.Program{Statements: []ast.Statement{
		&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 1}},
		&ast.ExpressionStatement{Expression: &ast.IntegerLiteral{Value: 2}},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 54. Additional coverage: Register compiler - slice expression standalone
// =============================================================================

func TestRegCompilerSliceExpressionStandalone(t *testing.T) {
	rc := newTestRegCompiler()
	_, err := rc.compileExpr(&ast.SliceExpression{
		Lower: &ast.IntegerLiteral{Value: 1},
		Upper: &ast.IntegerLiteral{Value: 3},
		Step:  &ast.IntegerLiteral{Value: 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// 55. Additional coverage: Register compiler - ClassInstantiation with local scope
// =============================================================================

func TestRegCompilerClassInstantiationLocalScope(t *testing.T) {
	rc := newTestRegCompiler()
	outer := rc.symbolTable
	rc.symbolTable = NewEnclosedSymbolTable(outer)
	rc.symbolTable.Define("MyClass")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.ClassInstantiation{
			ClassName: &ast.Identifier{Value: "MyClass"},
			Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 1}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpGetLocal) {
		t.Error("expected ROpGetLocal for class in local scope")
	}
}

// =============================================================================
// 56. Additional coverage: Register compiler - NamedExpression local scope
// =============================================================================

func TestRegCompilerNamedExpressionLocalScope(t *testing.T) {
	rc := newTestRegCompiler()
	outer := rc.symbolTable
	rc.symbolTable = NewEnclosedSymbolTable(outer)
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.NamedExpression{Name: &ast.Identifier{Value: "x"}, Value: &ast.IntegerLiteral{Value: 42}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasRegOpcode(rc.Instructions(), ROpSetLocal) {
		t.Error("expected ROpSetLocal for named expression in local scope")
	}
}

// =============================================================================
// 57. Additional coverage: Register compiler - FunctionLiteral with existing name
// =============================================================================

func TestRegCompilerFunctionWithExistingName(t *testing.T) {
	rc := newTestRegCompiler()
	rc.symbolTable.Define("existingFunc")
	err := rc.Compile(&ast.ExpressionStatement{
		Expression: &ast.FunctionLiteral{
			Name: "existingFunc", Parameters: []*ast.Identifier{},
			Body: &ast.BlockStatement{Statements: []ast.Statement{
				&ast.ReturnStatement{ReturnValue: &ast.IntegerLiteral{Value: 1}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
