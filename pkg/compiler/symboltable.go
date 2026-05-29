package compiler

import "github.com/go-py/go-python/pkg/objects"

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	outer           *SymbolTable
	store           map[string]Symbol
	numDefinitions  int
	FreeSymbols     []Symbol
	Free            []Symbol
	numFree         int
	// NestedFreeSymbols tracks which of this scope's locals are referenced by nested functions
	// This allows outer functions to know what to capture for their closures
	NestedFreeSymbols []Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store:             make(map[string]Symbol),
		Free:              []Symbol{},
		FreeSymbols:       []Symbol{},
		NestedFreeSymbols: []Symbol{},
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		outer:             outer,
		store:             make(map[string]Symbol),
		Free:              []Symbol{},
		FreeSymbols:       []Symbol{},
		NestedFreeSymbols: []Symbol{},
	}
}

func (s *SymbolTable) Define(name string) Symbol {
	// Local variables start after free variables
	symbol := Symbol{Name: name, Index: len(s.Free) + s.numDefinitions}
	if s.outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: len(s.Free) + s.numDefinitions, Scope: FunctionScope}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(name string, index int) {
	symbol := Symbol{Name: name, Scope: BuiltinScope, Index: index}
	s.store[name] = symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if ok {
		return obj, true
	}
	if s.outer != nil {
		obj, ok = s.outer.Resolve(name)
		if ok {
			if obj.Scope == LocalScope || obj.Scope == FunctionScope {
				if s.Free == nil {
					s.Free = []Symbol{}
				}

				s.FreeSymbols = append(s.FreeSymbols, obj)
				newSymbol := Symbol{Name: obj.Name, Scope: FreeScope, Index: len(s.Free)}
				s.Free = append(s.Free, newSymbol)

				// Notify the outer scope that one of its locals is being referenced by a nested function
				// This is crucial for closure support - outer function needs to know what to capture
				// We need to find the immediate outer scope that defines this variable
				// and add it to that scope's NestedFreeSymbols
				current := s.outer
				for current != nil {
					if outerSymbol, exists := current.store[obj.Name]; exists {
						current.NestedFreeSymbols = append(current.NestedFreeSymbols, outerSymbol)
						break
					}
					current = current.outer
				}

				return newSymbol, true
			}
			return obj, true
		}
	}
	return obj, ok
}

func (s *SymbolTable) numDefinitionsInScope() int {
	count := 0
	current := s
	for current != nil {
		count += current.numDefinitions
		current = current.outer
	}
	return count
}

type CompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
	IsGenerator   bool
	Free          []Symbol
	Name          string
	Constants     []objects.Object
	VarArgs       bool
	KwArgs        bool
}

func (cf *CompiledFunction) Type() objects.ObjectType {
	return objects.FUNCTION_OBJ
}

func (cf *CompiledFunction) Inspect() string {
	return "compiled function"
}
