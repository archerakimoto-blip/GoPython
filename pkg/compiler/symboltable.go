package compiler

import (
	"github.com/go-py/go-python/pkg/objects"
)

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
	// GlobalNames tracks variables declared with 'global' statement in this scope
	GlobalNames map[string]bool
	// NonlocalNames tracks variables declared with 'nonlocal' statement in this scope
	NonlocalNames map[string]bool
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store:              make(map[string]Symbol),
		Free:               []Symbol{},
		FreeSymbols:        []Symbol{},
		NestedFreeSymbols:  []Symbol{},
		GlobalNames:        make(map[string]bool),
		NonlocalNames:      make(map[string]bool),
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		outer:              outer,
		store:              make(map[string]Symbol),
		Free:               []Symbol{},
		FreeSymbols:        []Symbol{},
		NestedFreeSymbols:  []Symbol{},
		GlobalNames:        make(map[string]bool),
		NonlocalNames:      make(map[string]bool),
	}
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: len(s.Free) + s.numDefinitions}

	// Check if this is a global declaration in current scope
	if s.GlobalNames[name] {
		// In a function scope with global declaration, define in global scope
		if s.outer != nil {
			// Find global scope
			globalTable := s
			for globalTable.outer != nil {
				globalTable = globalTable.outer
			}
			// Define in global scope
			if existing, ok := globalTable.store[name]; ok {
				symbol = existing
			} else {
				symbol = Symbol{Name: name, Index: len(globalTable.store)}
				globalTable.store[name] = symbol
			}
			// Also store in current scope for resolution
			s.store[name] = symbol
			return symbol
		}
		symbol.Scope = GlobalScope
		s.store[name] = symbol
		return symbol
	}

	// Check if this is a nonlocal declaration
	if s.NonlocalNames[name] {
		// Find the enclosing scope that defines this variable
		current := s.outer
		for current != nil {
			if existing, ok := current.store[name]; ok {
				if existing.Scope == LocalScope || existing.Scope == FunctionScope {
					// Mark as free variable for this scope
					if s.Free == nil {
						s.Free = []Symbol{}
					}
					s.FreeSymbols = append(s.FreeSymbols, existing)
					newSymbol := Symbol{Name: existing.Name, Scope: FreeScope, Index: len(s.Free)}
					s.Free = append(s.Free, newSymbol)
					s.store[name] = newSymbol
					return newSymbol
				}
			}
			current = current.outer
		}
	}

	if s.outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

// DefineGlobal defines a variable as global in the current scope
// Used for 'global' statement
func (s *SymbolTable) DefineGlobal(name string) {
	s.GlobalNames[name] = true
}

// DefineNonlocal defines a variable as nonlocal in the current scope
// Used for 'nonlocal' statement
func (s *SymbolTable) DefineNonlocal(name string) {
	s.NonlocalNames[name] = true
}

// IsGlobal checks if a name is declared as global in the current scope
func (s *SymbolTable) IsGlobal(name string) bool {
	return s.GlobalNames[name]
}

// IsNonlocal checks if a name is declared as nonlocal in the current scope
func (s *SymbolTable) IsNonlocal(name string) bool {
	return s.NonlocalNames[name]
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
	IsAsync       bool
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
