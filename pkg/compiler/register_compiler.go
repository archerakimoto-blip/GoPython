package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/objects"
)

// RegOpcode defines opcodes for the register-based compiler backend.
// Defined locally to avoid circular imports with the vm package.
type RegOpcode byte

const (
	ROpLoadConst RegOpcode = iota
	ROpMove
	ROpAdd
	ROpSub
	ROpMul
	ROpDiv
	ROpMod
	ROpFloorDiv
	ROpPower
	ROpCompare
	ROpJump
	ROpJumpIfFalse
	ROpCall
	ROpReturn
	ROpReturnNone
	ROpGetAttr
	ROpSetAttr
	ROpGetGlobal
	ROpSetGlobal
	ROpGetLocal
	ROpSetLocal
	ROpBuildList
	ROpBuildDict
	ROpBuildSet
	ROpIndex
	ROpSlice
	ROpNegate
	ROpNot
	ROpNull
	ROpTrue
	ROpFalse
	ROpBitOr
	ROpBitAnd
	ROpBitXor
	ROpSetUnion
	ROpSetIntersection
	ROpSetDifference
	ROpSetSymmetricDifference
	ROpInPlaceAdd
	ROpInPlaceSub
	ROpInPlaceMul
	ROpInPlaceDiv
	ROpInPlaceMod
	ROpInPlaceFloorDiv
	ROpInPlacePower
	ROpInPlaceBitOr
	ROpInPlaceBitAnd
	ROpInPlaceBitXor
	ROpInPlaceLShift
	ROpInPlaceRShift
	ROpSetIndex
	ROpSetSlice
	ROpStringBuilderCreate
	ROpStringBuilderAppend
	ROpStringBuilderBuild
	ROpArrayPrealloc
	ROpClosure
	ROpGetFree
	ROpEllipsis
	ROpPop
	ROpDupTop
	ROpRaise
	ROpBeginTry
	ROpEndTry
	ROpExceptHandler
	ROpFinally
	ROpEnterContext
	ROpExitContext
	ROpMakeGenerator
	ROpMakeAsync
	ROpAwait
	ROpYieldValue
	ROpCreateClass
	ROpCreateClassWithSuper
	ROpCreateClassWithMultiSuper
	ROpDelAttribute
	ROpSetClassField
	ROpSetMetaclass
	ROpCallMetaclassInit
	ROpFormatString
	ROpListUnpack
	ROpDictUnpack
	ROpExceptStarHandler
	ROpLShift
	ROpRShift
	ROpBoolAnd
	ROpBoolOr
	ROpContains
	ROpNotContains
)

// RegInstruction represents a single register-based instruction.
type RegInstruction struct {
	Opcode   RegOpcode
	Operands []int
}

// RegBytecode holds the complete output of the register compiler:
// constants, instructions, and the number of registers needed.
type RegBytecode struct {
	Constants    []objects.Object
	Instructions []RegInstruction
	NumRegs      int
}

// Bytecode returns the compiled RegBytecode from the register compiler.
func (rc *RegisterCompiler) Bytecode() *RegBytecode {
	return &RegBytecode{
		Constants:    rc.constants,
		Instructions: rc.instructions,
		NumRegs:      rc.nextReg,
	}
}

// RegisterCompiler directly generates register instructions, bypassing
// the stack-based compilation and translation layers.
type RegisterCompiler struct {
	constants   []objects.Object
	symbolTable *SymbolTable
	instructions []RegInstruction
	nextReg     int
	freeRegs    []int
	// loopContexts tracks break/continue jump targets for nested loops
	loopContexts []loopContext
}

type loopContext struct {
	breakTarget    int // instruction index to jump to for break
	continueTarget int // instruction index to jump to for continue
	startIP        int // instruction index of loop start (for back-patching)
}

// NewRegisterCompiler creates a new RegisterCompiler.
func NewRegisterCompiler() *RegisterCompiler {
	c := &RegisterCompiler{
		constants:   []objects.Object{},
		symbolTable: NewSymbolTable(),
		instructions: make([]RegInstruction, 0),
	}
	c.registerBuiltins()
	return c
}

// NewRegisterCompilerWithState creates a RegisterCompiler with existing
// symbol table and constants (for REPL sessions).
func NewRegisterCompilerWithState(st *SymbolTable, constants []objects.Object) *RegisterCompiler {
	c := &RegisterCompiler{
		constants:   constants,
		symbolTable: st,
		instructions: make([]RegInstruction, 0),
	}
	if st == nil {
		c.symbolTable = NewSymbolTable()
	}
	if constants == nil {
		c.constants = []objects.Object{}
	}
	c.registerBuiltins()
	return c
}

// allocReg allocates a new register.
// It ensures the register doesn't overlap with local variable slots.
func (rc *RegisterCompiler) allocReg() int {
	minReg := rc.symbolTable.numDefinitions
	if len(rc.freeRegs) > 0 {
		// Find a free register that doesn't overlap with local slots
		for i := len(rc.freeRegs) - 1; i >= 0; i-- {
			if rc.freeRegs[i] >= minReg {
				reg := rc.freeRegs[i]
				rc.freeRegs = append(rc.freeRegs[:i], rc.freeRegs[i+1:]...)
				return reg
			}
		}
	}
	reg := rc.nextReg
	if reg < minReg {
		reg = minReg
		rc.nextReg = minReg
	}
	rc.nextReg++
	return reg
}

// freeReg returns a register to the free list.
func (rc *RegisterCompiler) freeReg(reg int) {
	rc.freeRegs = append(rc.freeRegs, reg)
}

// emitReg emits a register instruction.
func (rc *RegisterCompiler) emitReg(op RegOpcode, operands ...int) {
	rc.instructions = append(rc.instructions, RegInstruction{Opcode: op, Operands: operands})
}

// addConstant adds a constant and returns its index.
func (rc *RegisterCompiler) addConstant(obj objects.Object) int {
	rc.constants = append(rc.constants, obj)
	return len(rc.constants) - 1
}

// SymbolTable returns the compiler's symbol table.
func (rc *RegisterCompiler) SymbolTable() *SymbolTable {
	return rc.symbolTable
}

// Instructions returns the generated register instructions.
func (rc *RegisterCompiler) Instructions() []RegInstruction {
	return rc.instructions
}

// Constants returns the constants table.
func (rc *RegisterCompiler) Constants() []objects.Object {
	return rc.constants
}

// compileExpr compiles an expression node and returns the register holding the result.
func (rc *RegisterCompiler) compileExpr(node ast.Node) (int, error) {
	if node == nil {
		reg := rc.allocReg()
		rc.emitReg(ROpNull, reg)
		return reg, nil
	}

	switch node := node.(type) {
	case *ast.IntegerLiteral:
		reg := rc.allocReg()
		idx := rc.addConstant(&objects.Integer{Value: node.Value})
		rc.emitReg(ROpLoadConst, reg, idx)
		return reg, nil

	case *ast.FloatLiteral:
		reg := rc.allocReg()
		idx := rc.addConstant(&objects.Float{Value: node.Value})
		rc.emitReg(ROpLoadConst, reg, idx)
		return reg, nil

	case *ast.StringLiteral:
		reg := rc.allocReg()
		idx := rc.addConstant(&objects.String{Value: node.Value})
		rc.emitReg(ROpLoadConst, reg, idx)
		return reg, nil

	case *ast.Boolean:
		reg := rc.allocReg()
		if node.Value {
			rc.emitReg(ROpTrue, reg)
		} else {
			rc.emitReg(ROpFalse, reg)
		}
		return reg, nil

	case *ast.EllipsisLiteral:
		reg := rc.allocReg()
		rc.emitReg(ROpEllipsis, reg)
		return reg, nil

	case *ast.Identifier:
		symbol, ok := rc.symbolTable.Resolve(node.Value)
		if !ok {
			if node.Value == "True" {
				reg := rc.allocReg()
				rc.emitReg(ROpTrue, reg)
				return reg, nil
			} else if node.Value == "False" {
				reg := rc.allocReg()
				rc.emitReg(ROpFalse, reg)
				return reg, nil
			} else if node.Value == "None" {
				reg := rc.allocReg()
				rc.emitReg(ROpNull, reg)
				return reg, nil
			}
			return 0, fmt.Errorf("undefined variable %s", node.Value)
		}
		reg := rc.allocReg()
		if symbol.Scope == BuiltinScope {
			rc.emitReg(ROpLoadConst, reg, symbol.Index)
		} else if symbol.Scope == GlobalScope || symbol.Scope == FunctionScope {
			rc.emitReg(ROpGetGlobal, reg, symbol.Index)
		} else if symbol.Scope == FreeScope {
			rc.emitReg(ROpGetFree, reg, symbol.Index)
		} else {
			rc.emitReg(ROpGetLocal, reg, symbol.Index)
		}
		return reg, nil

	case *ast.PrefixExpression:
		right, err := rc.compileExpr(node.Right)
		if err != nil {
			return 0, err
		}
		dst := rc.allocReg()
		switch node.Operator {
		case "-":
			rc.emitReg(ROpNegate, dst, right)
		case "!", "not":
			rc.emitReg(ROpNot, dst, right)
		}
		rc.freeReg(right)
		return dst, nil

	case *ast.InfixExpression:
		left, err := rc.compileExpr(node.Left)
		if err != nil {
			return 0, err
		}
		right, err := rc.compileExpr(node.Right)
		if err != nil {
			return 0, err
		}
		dst := rc.allocReg()
		rc.freeReg(left)
		rc.freeReg(right)

		switch node.Operator {
		case "+":
			rc.emitReg(ROpAdd, dst, left, right)
		case "-":
			rc.emitReg(ROpSub, dst, left, right)
		case "*":
			rc.emitReg(ROpMul, dst, left, right)
		case "/":
			rc.emitReg(ROpDiv, dst, left, right)
		case "%":
			rc.emitReg(ROpMod, dst, left, right)
		case "//":
			rc.emitReg(ROpFloorDiv, dst, left, right)
		case "**":
			rc.emitReg(ROpPower, dst, left, right)
		case "|":
			rc.emitReg(ROpBitOr, dst, left, right)
		case "&":
			rc.emitReg(ROpBitAnd, dst, left, right)
		case "^":
			rc.emitReg(ROpBitXor, dst, left, right)
		case "<<":
			rc.emitReg(ROpLShift, dst, left, right)
		case ">>":
			rc.emitReg(ROpRShift, dst, left, right)
		case "==":
			rc.emitReg(ROpCompare, dst, left, right, int(OpEqual))
		case "!=":
			rc.emitReg(ROpCompare, dst, left, right, int(OpNotEqual))
		case ">":
			rc.emitReg(ROpCompare, dst, left, right, int(OpGreaterThan))
		case "<":
			rc.emitReg(ROpCompare, dst, left, right, int(OpLessThan))
		case ">=":
			rc.emitReg(ROpCompare, dst, left, right, int(OpGreaterEqual))
		case "<=":
			rc.emitReg(ROpCompare, dst, left, right, int(OpLessEqual))
		case "in":
			rc.emitReg(ROpContains, dst, left, right)
		case "not in":
			rc.emitReg(ROpNotContains, dst, left, right)
		case "and":
			// Short-circuit: if left is falsy, result = left; else result = right
			// Jump if left is falsy over right evaluation
			rc.freeReg(right) // we won't use the pre-compiled right
			rc.freeReg(dst)
			rc.freeReg(left)
			// Re-compile: evaluate left, jump if false, evaluate right, move to result
			leftReg, err := rc.compileExpr(node.Left)
			if err != nil {
				return 0, err
			}
			jumpIdx := len(rc.instructions)
			rc.emitReg(ROpJumpIfFalse, leftReg, -1) // placeholder
			rc.freeReg(leftReg)
			rightReg, err := rc.compileExpr(node.Right)
			if err != nil {
				return 0, err
			}
			resultReg := rc.allocReg()
			rc.emitReg(ROpMove, resultReg, rightReg)
			rc.freeReg(rightReg)
			endIdx := len(rc.instructions)
			rc.emitReg(ROpJump, -1) // placeholder
			rc.instructions[jumpIdx].Operands[1] = len(rc.instructions)
			// Falsy path: result = left
			leftReg2, err := rc.compileExpr(node.Left)
			if err != nil {
				return 0, err
			}
			rc.emitReg(ROpMove, resultReg, leftReg2)
			rc.freeReg(leftReg2)
			rc.instructions[endIdx].Operands[0] = len(rc.instructions)
			return resultReg, nil
		case "or":
			// Short-circuit: if left is truthy, result = left; else result = right
			rc.freeReg(right)
			rc.freeReg(dst)
			rc.freeReg(left)
			leftReg, err := rc.compileExpr(node.Left)
			if err != nil {
				return 0, err
			}
			jumpIdx := len(rc.instructions)
			rc.emitReg(ROpJumpIfFalse, leftReg, -1) // placeholder: if false, go to right
			rc.freeReg(leftReg)
			// Truthy path: result = left
			leftReg2, err := rc.compileExpr(node.Left)
			if err != nil {
				return 0, err
			}
			resultReg := rc.allocReg()
			rc.emitReg(ROpMove, resultReg, leftReg2)
			rc.freeReg(leftReg2)
			endIdx := len(rc.instructions)
			rc.emitReg(ROpJump, -1) // placeholder
			rc.instructions[jumpIdx].Operands[1] = len(rc.instructions)
			// Falsy path: result = right
			rightReg, err := rc.compileExpr(node.Right)
			if err != nil {
				return 0, err
			}
			rc.emitReg(ROpMove, resultReg, rightReg)
			rc.freeReg(rightReg)
			rc.instructions[endIdx].Operands[0] = len(rc.instructions)
			return resultReg, nil
		default:
			return 0, fmt.Errorf("unknown operator %s", node.Operator)
		}
		return dst, nil

	case *ast.IfExpression:
		cond, err := rc.compileExpr(node.Condition)
		if err != nil {
			return 0, err
		}
		// Jump if false placeholder
		jumpIdx := len(rc.instructions)
		rc.emitReg(ROpJumpIfFalse, cond, -1) // placeholder
		rc.freeReg(cond)

		// Compile consequence - try to extract result if it's a single expression
		var conseqReg int
		conseqCompiled := false
		if node.Consequence != nil && len(node.Consequence.Statements) == 1 {
			if exprStmt, ok := node.Consequence.Statements[0].(*ast.ExpressionStatement); ok {
				conseqReg, err = rc.compileExpr(exprStmt.Expression)
				if err != nil {
					return 0, err
				}
				conseqCompiled = true
			}
		}
		if !conseqCompiled {
			if node.Consequence != nil {
				err = rc.compileStmt(node.Consequence)
				if err != nil {
					return 0, err
				}
			}
			conseqReg = rc.allocReg()
			rc.emitReg(ROpNull, conseqReg)
		}

		// Jump over alternative
		jumpOverIdx := len(rc.instructions)
		rc.emitReg(ROpJump, -1) // placeholder

		// Patch the jump-if-false target
		rc.instructions[jumpIdx].Operands[1] = len(rc.instructions)

		// Compile alternative - try to extract result if it's a single expression
		var altReg int
		altCompiled := false
		if node.Alternative != nil && len(node.Alternative.Statements) == 1 {
			if exprStmt, ok := node.Alternative.Statements[0].(*ast.ExpressionStatement); ok {
				altReg, err = rc.compileExpr(exprStmt.Expression)
				if err != nil {
					return 0, err
				}
				altCompiled = true
			}
		}
		if !altCompiled {
			if node.Alternative != nil {
				err = rc.compileStmt(node.Alternative)
				if err != nil {
					return 0, err
				}
			}
			altReg = rc.allocReg()
			rc.emitReg(ROpNull, altReg)
		}

		// Move alternative result to same register as consequence
		rc.emitReg(ROpMove, conseqReg, altReg)
		rc.freeReg(altReg)

		// Patch the jump-over target
		rc.instructions[jumpOverIdx].Operands[0] = len(rc.instructions)

		return conseqReg, nil

	case *ast.CallExpression:
		funcReg, err := rc.compileExpr(node.Function)
		if err != nil {
			return 0, err
		}
		argRegs := make([]int, 0, len(node.Arguments))
		for _, arg := range node.Arguments {
			argReg, err := rc.compileExpr(arg)
			if err != nil {
				return 0, err
			}
			argRegs = append(argRegs, argReg)
		}
		dst := rc.allocReg()
		operands := []int{dst, funcReg, len(argRegs)}
		operands = append(operands, argRegs...)
		rc.emitReg(ROpCall, operands...)
		rc.freeReg(funcReg)
		for _, r := range argRegs {
			rc.freeReg(r)
		}
		return dst, nil

	case *ast.IndexExpression:
		left, err := rc.compileExpr(node.Left)
		if err != nil {
			return 0, err
		}
		if slice, ok := node.Index.(*ast.SliceExpression); ok {
			var startReg, endReg, stepReg int
			if slice.Lower != nil {
				startReg, err = rc.compileExpr(slice.Lower)
				if err != nil {
					return 0, err
				}
			} else {
				startReg = rc.allocReg()
				rc.emitReg(ROpLoadConst, startReg, rc.addConstant(&objects.Integer{Value: 0}))
			}
			if slice.Upper != nil {
				endReg, err = rc.compileExpr(slice.Upper)
				if err != nil {
					return 0, err
				}
			} else {
				endReg = rc.allocReg()
				rc.emitReg(ROpLoadConst, endReg, rc.addConstant(&objects.Integer{Value: -1}))
			}
			if slice.Step != nil {
				stepReg, err = rc.compileExpr(slice.Step)
				if err != nil {
					return 0, err
				}
			} else {
				stepReg = rc.allocReg()
				rc.emitReg(ROpNull, stepReg)
			}
			dst := rc.allocReg()
			rc.emitReg(ROpSlice, dst, left, startReg, endReg, stepReg)
			rc.freeReg(left)
			rc.freeReg(startReg)
			rc.freeReg(endReg)
			rc.freeReg(stepReg)
			return dst, nil
		}

		index, err := rc.compileExpr(node.Index)
		if err != nil {
			return 0, err
		}
		dst := rc.allocReg()
		rc.emitReg(ROpIndex, dst, left, index)
		rc.freeReg(left)
		rc.freeReg(index)
		return dst, nil

	case *ast.MemberAccess:
		obj, err := rc.compileExpr(node.Object)
		if err != nil {
			return 0, err
		}
		dst := rc.allocReg()
		idx := rc.addConstant(&objects.String{Value: node.Member.Value})
		rc.emitReg(ROpGetAttr, dst, obj, idx)
		rc.freeReg(obj)
		return dst, nil

	case *ast.ListLiteral:
		elemRegs := make([]int, len(node.Elements))
		for i, el := range node.Elements {
			reg, err := rc.compileExpr(el)
			if err != nil {
				return 0, err
			}
			elemRegs[i] = reg
		}
		dst := rc.allocReg()
		operands := []int{dst, len(elemRegs)}
		operands = append(operands, elemRegs...)
		rc.emitReg(ROpBuildList, operands...)
		for _, r := range elemRegs {
			rc.freeReg(r)
		}
		return dst, nil

	case *ast.HashLiteral:
		keys := make([]ast.Expression, 0, len(node.Pairs))
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		pairRegs := make([]int, 0, len(keys)*2)
		for _, k := range keys {
			keyReg, err := rc.compileExpr(k)
			if err != nil {
				return 0, err
			}
			valReg, err := rc.compileExpr(node.Pairs[k])
			if err != nil {
				return 0, err
			}
			pairRegs = append(pairRegs, keyReg, valReg)
		}
		dst := rc.allocReg()
		operands := []int{dst, len(pairRegs)}
		operands = append(operands, pairRegs...)
		rc.emitReg(ROpBuildDict, operands...)
		for _, r := range pairRegs {
			rc.freeReg(r)
		}
		return dst, nil

	case *ast.SetLiteral:
		elemRegs := make([]int, len(node.Elements))
		for i, el := range node.Elements {
			reg, err := rc.compileExpr(el)
			if err != nil {
				return 0, err
			}
			elemRegs[i] = reg
		}
		dst := rc.allocReg()
		operands := []int{dst, len(elemRegs)}
		operands = append(operands, elemRegs...)
		rc.emitReg(ROpBuildSet, operands...)
		for _, r := range elemRegs {
			rc.freeReg(r)
		}
		return dst, nil

	case *ast.FunctionLiteral:
		// If the function has a name, define it in the outer symbol table
		// before compiling the body, so recursive calls can find it
		var fnSymbol Symbol
		var fnSymbolOk bool
		if node.Name != "" {
			fnSymbol, fnSymbolOk = rc.symbolTable.Resolve(node.Name)
			if !fnSymbolOk {
				fnSymbol = rc.symbolTable.DefineFunctionName(node.Name)
			}
		}

		// Save outer state
		outerInstructions := rc.instructions
		outerNextReg := rc.nextReg
		outerFreeRegs := rc.freeRegs
		rc.instructions = make([]RegInstruction, 0)
		rc.nextReg = 0
		rc.freeRegs = nil

		rc.symbolTable = NewEnclosedSymbolTable(rc.symbolTable)
		for _, p := range node.Parameters {
			rc.symbolTable.Define(p.Value)
		}
		// Reserve registers for parameters so temporaries don't overwrite them
		rc.nextReg = len(node.Parameters)

		// Handle global/nonlocal
		for _, stmt := range node.Body.Statements {
			switch s := stmt.(type) {
			case *ast.GlobalStatement:
				for _, name := range s.Names {
					rc.symbolTable.DefineGlobal(name.Value)
				}
			case *ast.NonlocalStatement:
				for _, name := range s.Names {
					rc.symbolTable.DefineNonlocal(name.Value)
				}
			}
		}

		err := rc.compileStmt(node.Body)
		if err != nil {
			return 0, err
		}

		// Ensure return
		if len(rc.instructions) == 0 || rc.instructions[len(rc.instructions)-1].Opcode != ROpReturn {
			nullReg := rc.allocReg()
			rc.emitReg(ROpNull, nullReg)
			rc.emitReg(ROpReturn, nullReg)
		}

		fnInstructions := rc.instructions
		numLocals := rc.symbolTable.numDefinitions
		freeVars := rc.symbolTable.Free
		freeSymbols := rc.symbolTable.FreeSymbols

		// In the register compiler, we only use freeVars (not nestedFreeSymbols).
		// nestedFreeSymbols is a stack-VM concept for propagating captures through
		// intermediate functions. In the register compiler, when a nested function
		// resolves a variable from an outer scope, the Resolve() function already
		// creates free variable entries in each intermediate scope, so the chain
		// is handled correctly through freeSymbols alone.
		allFreeVars := make([]Symbol, 0, len(freeVars))
		allFreeVars = append(allFreeVars, freeVars...)

		numFree := len(freeVars)

		rc.symbolTable = rc.symbolTable.outer

		// Restore outer state (instructions, nextReg, freeRegs) so that
		// free variable loading instructions are emitted into the outer
		// function's instruction list with correct register allocation.
		rc.instructions = outerInstructions
		rc.nextReg = outerNextReg
		rc.freeRegs = outerFreeRegs

		// Load free variable values using the freeSymbols' original scope info.
		// freeSymbols contains the ORIGINAL symbols from outer scopes, so we
		// emit load instructions based on their Scope/Index directly, rather
		// than re-resolving through the symbol table.
		var freeRegs []int
		if numFree > 0 {
			freeRegs = make([]int, 0, numFree)
			for _, freeSym := range freeSymbols {
				fr := rc.allocReg()
				switch freeSym.Scope {
				case LocalScope:
					rc.emitReg(ROpGetLocal, fr, freeSym.Index)
				case FreeScope:
					rc.emitReg(ROpGetFree, fr, freeSym.Index)
				case GlobalScope:
					rc.emitReg(ROpGetGlobal, fr, freeSym.Index)
				case FunctionScope:
					rc.emitReg(ROpGetGlobal, fr, freeSym.Index)
				case BuiltinScope:
					rc.emitReg(ROpLoadConst, fr, freeSym.Index)
				}
				freeRegs = append(freeRegs, fr)
			}
		}

		paramNames := make([]string, len(node.Parameters))
		for i, p := range node.Parameters {
			paramNames[i] = p.Value
		}

		compiledFn := &CompiledFunction{
			Instructions:    regInstructionsToBytes(fnInstructions),
			NumLocals:       numLocals,
			NumParameters:   len(node.Parameters),
			ParameterNames:  paramNames,
			IsGenerator:     hasYieldInBody(node.Body),
			IsAsync:         node.IsAsync,
			Free:            allFreeVars,
			VarArgs:         node.VarArgs != nil,
			KwArgs:          node.KwArgs != nil,
			RegInstructions: fnInstructions,
		}

		dst := rc.allocReg()
		if numFree > 0 {
			idx := rc.addConstant(compiledFn)
			operands := []int{dst, idx, numFree}
			operands = append(operands, freeRegs...)
			rc.emitReg(ROpClosure, operands...)
			for _, r := range freeRegs {
				rc.freeReg(r)
			}
		} else {
			idx := rc.addConstant(compiledFn)
			rc.emitReg(ROpLoadConst, dst, idx)
		}

		if compiledFn.IsGenerator {
			rc.emitReg(ROpMakeGenerator, dst, dst)
		}
		if compiledFn.IsAsync {
			rc.emitReg(ROpMakeAsync, dst, dst)
		}

		// If the function has a name, store it using the pre-defined symbol
		if node.Name != "" {
			if fnSymbol.Scope == GlobalScope || fnSymbol.Scope == FunctionScope {
				rc.emitReg(ROpSetGlobal, fnSymbol.Index, dst)
			} else {
				rc.emitReg(ROpSetLocal, fnSymbol.Index, dst)
			}
		}

		return dst, nil

	case *ast.LambdaExpression:
		funcLit := &ast.FunctionLiteral{
			Token:      node.Token,
			Parameters: node.Parameters,
			Body: &ast.BlockStatement{
				Token: node.Token,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      node.Token,
						Expression: node.Body,
					},
				},
			},
		}
		return rc.compileExpr(funcLit)

	case *ast.ComplexLiteral:
		s := node.Value
		s = strings.TrimRight(s, "jJ")
		imagPart, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("could not parse complex literal %q: %s", node.Value, err)
		}
		complexObj := objects.NewComplex(0, imagPart)
		reg := rc.allocReg()
		idx := rc.addConstant(complexObj)
		rc.emitReg(ROpLoadConst, reg, idx)
		return reg, nil

	case *ast.ByteStringLiteral:
		reg := rc.allocReg()
		idx := rc.addConstant(&objects.Bytes{Value: []byte(node.Value)})
		rc.emitReg(ROpLoadConst, reg, idx)
		return reg, nil

	case *ast.FStringLiteral:
		partRegs := make([]int, 0, len(node.Parts))
		for _, part := range node.Parts {
			pr, err := rc.compileExpr(part)
			if err != nil {
				return 0, err
			}
			partRegs = append(partRegs, pr)
		}
		dst := rc.allocReg()
		operands := []int{dst, len(partRegs)}
		operands = append(operands, partRegs...)
		rc.emitReg(ROpFormatString, operands...)
		for _, r := range partRegs {
			rc.freeReg(r)
		}
		return dst, nil

	case *ast.TernaryExpression:
		// Ternary: value_if_true if condition else value_if_false
		// Compile as: evaluate condition, jump to else branch if false
		condReg, err := rc.compileExpr(node.Condition)
		if err != nil {
			return 0, err
		}
		// Placeholder jump (will patch later)
		jumpIdx := len(rc.instructions)
		rc.emitReg(ROpJumpIfFalse, condReg, 0) // placeholder target
		rc.freeReg(condReg)

		// Consequence branch
		conseqReg, err := rc.compileExpr(node.Consequence)
		if err != nil {
			return 0, err
		}
		// Jump over alternative
		jumpEndIdx := len(rc.instructions)
		rc.emitReg(ROpJump, 0) // placeholder target

		// Patch the false-jump to land here (alternative branch)
		rc.instructions[jumpIdx].Operands[1] = len(rc.instructions)

		altReg, err := rc.compileExpr(node.Alternative)
		if err != nil {
			return 0, err
		}

		// Move alternative to same register as consequence
		rc.emitReg(ROpMove, conseqReg, altReg)
		rc.freeReg(altReg)

		// Patch the end-jump
		rc.instructions[jumpEndIdx].Operands[0] = len(rc.instructions)

		return conseqReg, nil

	case *ast.NamedExpression:
		// Walrus operator: x := expr
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return 0, err
		}
		symbol := rc.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			rc.emitReg(ROpSetGlobal, symbol.Index, val)
		} else {
			rc.emitReg(ROpSetLocal, symbol.Index, val)
		}
		// The expression result is the assigned value
		return val, nil

	case *ast.AwaitExpression:
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return 0, err
		}
		dst := rc.allocReg()
		rc.emitReg(ROpAwait, dst, val)
		rc.freeReg(val)
		return dst, nil

	case *ast.DictionaryUnpack:
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return 0, err
		}
		rc.emitReg(ROpDictUnpack, val)
		return val, nil

	case *ast.ListUnpack:
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return 0, err
		}
		rc.emitReg(ROpListUnpack, val)
		return val, nil

	case *ast.SliceExpression:
		// Standalone slice expression (rare; usually wrapped in IndexExpression)
		var startReg, endReg, stepReg int
		var sliceErr error
		if node.Lower != nil {
			startReg, sliceErr = rc.compileExpr(node.Lower)
			if sliceErr != nil {
				return 0, sliceErr
			}
		} else {
			startReg = rc.allocReg()
			rc.emitReg(ROpLoadConst, startReg, rc.addConstant(&objects.Integer{Value: 0}))
		}
		if node.Upper != nil {
			endReg, sliceErr = rc.compileExpr(node.Upper)
			if sliceErr != nil {
				return 0, sliceErr
			}
		} else {
			endReg = rc.allocReg()
			rc.emitReg(ROpLoadConst, endReg, rc.addConstant(&objects.Integer{Value: -1}))
		}
		if node.Step != nil {
			stepReg, sliceErr = rc.compileExpr(node.Step)
			if sliceErr != nil {
				return 0, sliceErr
			}
		} else {
			stepReg = rc.allocReg()
			rc.emitReg(ROpNull, stepReg)
		}
		// Return a tuple (start, end, step) as a simplified representation
		dst := rc.allocReg()
		rc.emitReg(ROpBuildList, dst, 3, startReg, endReg, stepReg)
		rc.freeReg(startReg)
		rc.freeReg(endReg)
		rc.freeReg(stepReg)
		return dst, nil

	case *ast.MethodCall:
		obj, err := rc.compileExpr(node.Object)
		if err != nil {
			return 0, err
		}
		// Get method attribute
		methodReg := rc.allocReg()
		idx := rc.addConstant(&objects.String{Value: node.Method.Value})
		rc.emitReg(ROpGetAttr, methodReg, obj, idx)
		rc.freeReg(obj)

		argRegs := make([]int, 0, len(node.Arguments))
		for _, arg := range node.Arguments {
			argReg, err := rc.compileExpr(arg)
			if err != nil {
				return 0, err
			}
			argRegs = append(argRegs, argReg)
		}
		dst := rc.allocReg()
		operands := []int{dst, methodReg, len(argRegs)}
		operands = append(operands, argRegs...)
		rc.emitReg(ROpCall, operands...)
		rc.freeReg(methodReg)
		for _, r := range argRegs {
			rc.freeReg(r)
		}
		return dst, nil

	case *ast.ClassInstantiation:
		// Class instantiation: ClassName(args...)
		// Load the class name as a global
		className := node.ClassName.Value
		symbol, ok := rc.symbolTable.Resolve(className)
		if !ok {
			symbol = rc.symbolTable.Define(className)
		}
		funcReg := rc.allocReg()
		if symbol.Scope == GlobalScope {
			rc.emitReg(ROpGetGlobal, funcReg, symbol.Index)
		} else {
			rc.emitReg(ROpGetLocal, funcReg, symbol.Index)
		}

		argRegs := make([]int, 0, len(node.Arguments))
		for _, arg := range node.Arguments {
			argReg, err := rc.compileExpr(arg)
			if err != nil {
				return 0, err
			}
			argRegs = append(argRegs, argReg)
		}
		dst := rc.allocReg()
		operands := []int{dst, funcReg, len(argRegs)}
		operands = append(operands, argRegs...)
		rc.emitReg(ROpCall, operands...)
		rc.freeReg(funcReg)
		for _, r := range argRegs {
			rc.freeReg(r)
		}
		return dst, nil

	case *ast.ListComprehension, *ast.SetComprehension, *ast.DictComprehension,
		*ast.GeneratorExpression,
		*ast.AsyncListComprehension, *ast.AsyncSetComprehension,
		*ast.AsyncDictComprehension, *ast.AsyncGeneratorExpression:
		// Comprehensions and generator expressions should be desugared
		// before reaching the register compiler. Emit a simplified version:
		// return None as a placeholder.
		reg := rc.allocReg()
		rc.emitReg(ROpNull, reg)
		return reg, nil

	case *ast.KeywordArgument:
		// Just compile the value
		return rc.compileExpr(node.Value)

	case *ast.SpreadItem:
		// Spread item in list/set literal: *expr
		return rc.compileExpr(node.Value)

	case *ast.YieldStatement:
		if node.Expression != nil {
			val, err := rc.compileExpr(node.Expression)
			if err != nil {
				return 0, err
			}
			rc.emitReg(ROpYieldValue, val)
			rc.freeReg(val)
		} else {
			nullReg := rc.allocReg()
			rc.emitReg(ROpNull, nullReg)
			rc.emitReg(ROpYieldValue, nullReg)
			rc.freeReg(nullReg)
		}
		reg := rc.allocReg()
		rc.emitReg(ROpNull, reg)
		return reg, nil

	default:
		// Fallback: return None in a register
		reg := rc.allocReg()
		rc.emitReg(ROpNull, reg)
		return reg, nil
	}
}

// compileStmt compiles a statement node.
func (rc *RegisterCompiler) compileStmt(node ast.Node) error {
	if node == nil {
		return nil
	}

	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := rc.compileStmt(s); err != nil {
				return err
			}
		}

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			if err := rc.compileStmt(s); err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		reg, err := rc.compileExpr(node.Expression)
		if err != nil {
			return err
		}
		rc.freeReg(reg)

	case *ast.LetStatement:
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return err
		}
		symbol := rc.symbolTable.Define(node.Names[0].Value)
		if symbol.Scope == GlobalScope {
			rc.emitReg(ROpSetGlobal, symbol.Index, val)
		} else {
			rc.emitReg(ROpSetLocal, symbol.Index, val)
		}
		rc.freeReg(val)

	case *ast.AssignStatement:
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return err
		}
		symbol, ok := rc.symbolTable.Resolve(node.Names[0].Value)
		if !ok {
			symbol = rc.symbolTable.Define(node.Names[0].Value)
		}
		if symbol.Scope == GlobalScope {
			rc.emitReg(ROpSetGlobal, symbol.Index, val)
		} else {
			rc.emitReg(ROpSetLocal, symbol.Index, val)
		}
		rc.freeReg(val)

	case *ast.ReturnStatement:
		val, err := rc.compileExpr(node.ReturnValue)
		if err != nil {
			return err
		}
		rc.emitReg(ROpReturn, val)
		rc.freeReg(val)

	case *ast.WhileStatement:
		conditionPos := len(rc.instructions)
		cond, err := rc.compileExpr(node.Condition)
		if err != nil {
			return err
		}
		jumpIdx := len(rc.instructions)
		rc.emitReg(ROpJumpIfFalse, cond, -1) // placeholder
		rc.freeReg(cond)

		// Push loop context for break/continue
		rc.loopContexts = append(rc.loopContexts, loopContext{
			breakTarget:    -1, // will be back-patched
			continueTarget: conditionPos,
			startIP:        conditionPos,
		})

		if err := rc.compileStmt(node.Body); err != nil {
			rc.loopContexts = rc.loopContexts[:len(rc.loopContexts)-1]
			return err
		}
		rc.emitReg(ROpJump, conditionPos)

		afterLoop := len(rc.instructions)
		rc.instructions[jumpIdx].Operands[1] = afterLoop

		// Back-patch any break jumps in this loop
		lc := &rc.loopContexts[len(rc.loopContexts)-1]
		if lc.breakTarget == -1 {
			lc.breakTarget = afterLoop
		}
		for i := lc.startIP; i < afterLoop; i++ {
			if rc.instructions[i].Opcode == ROpJump && rc.instructions[i].Operands[0] == -2 {
				// -2 is our sentinel for break jump
				rc.instructions[i].Operands[0] = afterLoop
			}
			if rc.instructions[i].Opcode == ROpJumpIfFalse && len(rc.instructions[i].Operands) >= 2 && rc.instructions[i].Operands[1] == -3 {
				// -3 is our sentinel for continue jump
				rc.instructions[i].Operands[1] = conditionPos
			}
		}
		rc.loopContexts = rc.loopContexts[:len(rc.loopContexts)-1]

	case *ast.AugAssignStatement:
		if node.Name != nil {
			left, err := rc.compileExpr(node.Name)
			if err != nil {
				return err
			}
			right, err := rc.compileExpr(node.Value)
			if err != nil {
				return err
			}
			dst := rc.allocReg()
			rc.freeReg(left)
			rc.freeReg(right)

			regOp := augAssignToRegInPlaceOp(node.Operator)
			rc.emitReg(regOp, dst, left, right)

			symbol, ok := rc.symbolTable.Resolve(node.Name.Value)
			if !ok {
				symbol = rc.symbolTable.Define(node.Name.Value)
			}
			rc.emitSetSymbol(symbol, dst)
			rc.freeReg(dst)
		}

	case *ast.DeleteStatement:
		for _, target := range node.Targets {
			switch t := target.(type) {
			case *ast.Identifier:
				symbol, ok := rc.symbolTable.Resolve(t.Value)
				if !ok {
					return fmt.Errorf("undefined variable %s", t.Value)
				}
				nullReg := rc.allocReg()
				rc.emitReg(ROpNull, nullReg)
				if symbol.Scope == GlobalScope {
					rc.emitReg(ROpSetGlobal, symbol.Index, nullReg)
				} else {
					rc.emitReg(ROpSetLocal, symbol.Index, nullReg)
				}
				rc.freeReg(nullReg)
			case *ast.MemberAccess:
				obj, err := rc.compileExpr(t.Object)
				if err != nil {
					return err
				}
				idx := rc.addConstant(&objects.String{Value: t.Member.Value})
				rc.emitReg(ROpDelAttribute, obj, idx)
				rc.freeReg(obj)
			}
		}

	case *ast.RaiseStatement:
		if node.Expression != nil {
			val, err := rc.compileExpr(node.Expression)
			if err != nil {
				return err
			}
			rc.emitReg(ROpRaise, val)
			rc.freeReg(val)
		} else {
			nullReg := rc.allocReg()
			rc.emitReg(ROpNull, nullReg)
			rc.emitReg(ROpRaise, nullReg)
			rc.freeReg(nullReg)
		}

	case *ast.ClassStatement:
		class := &objects.Class{
			Name:    node.Name.Value,
			Methods: make(map[string]objects.Object),
			Fields:  make(map[string]objects.Object),
		}
		for _, method := range node.Methods {
			compiledFn := rc.compileFunction(method)
			if compiledFn != nil {
				class.Methods[method.Name] = compiledFn
			}
		}

		dst := rc.allocReg()
		if node.SuperClass != nil {
			superReg, err := rc.compileExpr(&ast.Identifier{Value: node.SuperClass.Value})
			if err != nil {
				return err
			}
			idx := rc.addConstant(class)
			rc.emitReg(ROpCreateClassWithSuper, dst, idx, superReg)
			rc.freeReg(superReg)
		} else {
			idx := rc.addConstant(class)
			rc.emitReg(ROpCreateClass, dst, idx)
		}

		// Handle metaclass
		if node.Metaclass != nil {
			metaReg, err := rc.compileExpr(node.Metaclass)
			if err != nil {
				return err
			}
			rc.emitReg(ROpSetMetaclass, dst, metaReg)
			rc.freeReg(metaReg)
		}

		// Handle multiple inheritance
		if len(node.SuperClasses) > 1 {
			for i := 1; i < len(node.SuperClasses); i++ {
				superReg, err := rc.compileExpr(&ast.Identifier{Value: node.SuperClasses[i].Value})
				if err != nil {
					return err
				}
				idx := rc.addConstant(class)
				rc.emitReg(ROpCreateClassWithMultiSuper, dst, idx, superReg)
				rc.freeReg(superReg)
			}
		}

		symbol := rc.symbolTable.Define(node.Name.Value)
		rc.emitSetSymbol(symbol, dst)
		rc.freeReg(dst)

	case *ast.TryStatement:
		return rc.compileTryStatement(node)

	case *ast.WithStatement:
		if len(node.Items) > 0 {
			item := node.Items[0]
			ctx, err := rc.compileExpr(item.Expr)
			if err != nil {
				return err
			}
			dst := rc.allocReg()
			rc.emitReg(ROpEnterContext, dst, ctx)
			rc.freeReg(ctx)

			if item.Name != nil {
				symbol := rc.symbolTable.Define(item.Name.Value)
				if symbol.Scope == GlobalScope {
					rc.emitReg(ROpSetGlobal, symbol.Index, dst)
				} else {
					rc.emitReg(ROpSetLocal, symbol.Index, dst)
				}
			}
			rc.freeReg(dst)

			if err := rc.compileStmt(node.Body); err != nil {
				return err
			}

			// Exit context (simplified)
			exitDst := rc.allocReg()
			nullReg := rc.allocReg()
			rc.emitReg(ROpNull, nullReg)
			rc.emitReg(ROpExitContext, exitDst, dst, nullReg)
			rc.freeReg(exitDst)
			rc.freeReg(nullReg)
		}

	case *ast.PassStatement:
		// no-op

	case *ast.GlobalStatement, *ast.NonlocalStatement:
		// declarations only, no code

	case *ast.ForStatement:
		// For loops should be desugared before reaching the register compiler.
		// If we get here, emit a simplified version using while-loop pattern.
		return fmt.Errorf("for loops should be desugared before compilation")

	case *ast.ImportStatement:
		moduleName := node.Module.Value
		alias := moduleName
		if node.Alias != nil {
			alias = node.Alias.Value
		}
		module := objects.GetModule(moduleName)
		if module == nil {
			return fmt.Errorf("module '%s' not found", moduleName)
		}
		var symbol Symbol
		if existingSymbol, ok := rc.symbolTable.Resolve(alias); ok {
			if existingSymbol.Scope == BuiltinScope {
				symbol = rc.symbolTable.Define(alias)
			} else {
				symbol = existingSymbol
			}
		} else {
			symbol = rc.symbolTable.Define(alias)
		}
		dst := rc.allocReg()
		idx := rc.addConstant(module)
		rc.emitReg(ROpLoadConst, dst, idx)
		if symbol.Scope == GlobalScope {
			rc.emitReg(ROpSetGlobal, symbol.Index, dst)
		} else {
			rc.emitReg(ROpSetLocal, symbol.Index, dst)
		}
		rc.freeReg(dst)

	case *ast.FromImportStatement:
		moduleName := node.Module.Value
		module := objects.GetModule(moduleName)
		if module == nil {
			return fmt.Errorf("module '%s' not found", moduleName)
		}
		if node.Alias != nil {
			symbol := rc.symbolTable.Define(node.Alias.Value)
			dst := rc.allocReg()
			idx := rc.addConstant(module)
			rc.emitReg(ROpLoadConst, dst, idx)
			if symbol.Scope == GlobalScope {
				rc.emitReg(ROpSetGlobal, symbol.Index, dst)
			} else {
				rc.emitReg(ROpSetLocal, symbol.Index, dst)
			}
			rc.freeReg(dst)
		} else {
			for _, name := range node.Names {
				value, ok := module.Fields[name.Value]
				if !ok {
					return fmt.Errorf("name '%s' not found in module '%s'", name.Value, moduleName)
				}
				symbol := rc.symbolTable.Define(name.Value)
				dst := rc.allocReg()
				idx := rc.addConstant(value)
				rc.emitReg(ROpLoadConst, dst, idx)
				if symbol.Scope == GlobalScope {
					rc.emitReg(ROpSetGlobal, symbol.Index, dst)
				} else {
					rc.emitReg(ROpSetLocal, symbol.Index, dst)
				}
				rc.freeReg(dst)
			}
		}

	case *ast.MatchStatement:
		// Match statements should be desugared before reaching the register compiler.
		return fmt.Errorf("match statement should be desugared before compilation")

	case *ast.YieldFromStatement:
		if node.Expression != nil {
			val, err := rc.compileExpr(node.Expression)
			if err != nil {
				return err
			}
			rc.emitReg(ROpYieldValue, val)
			rc.freeReg(val)
		}

	case *ast.AsyncForStatement:
		// Async for should be desugared. Fall back to simplified handling.
		return fmt.Errorf("async for should be desugared before compilation")

	case *ast.AsyncWithStatement:
		// Handle similar to WithStatement
		if len(node.Items) > 0 {
			item := node.Items[0]
			ctx, err := rc.compileExpr(item.Expr)
			if err != nil {
				return err
			}
			dst := rc.allocReg()
			rc.emitReg(ROpEnterContext, dst, ctx)
			rc.freeReg(ctx)

			if item.Name != nil {
				symbol := rc.symbolTable.Define(item.Name.Value)
				if symbol.Scope == GlobalScope {
					rc.emitReg(ROpSetGlobal, symbol.Index, dst)
				} else {
					rc.emitReg(ROpSetLocal, symbol.Index, dst)
				}
			}
			rc.freeReg(dst)

			if err := rc.compileStmt(node.Body); err != nil {
				return err
			}

			exitDst := rc.allocReg()
			nullReg := rc.allocReg()
			rc.emitReg(ROpNull, nullReg)
			rc.emitReg(ROpExitContext, exitDst, dst, nullReg)
			rc.freeReg(exitDst)
			rc.freeReg(nullReg)
		}

	case *ast.AttributeAssignStatement:
		obj, err := rc.compileExpr(node.Object)
		if err != nil {
			return err
		}
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return err
		}
		attrIdx := rc.addConstant(&objects.String{Value: node.Attr.Value})
		rc.emitReg(ROpSetAttr, obj, val, attrIdx)
		rc.freeReg(obj)
		rc.freeReg(val)

	case *ast.IndexAssignStatement:
		left, err := rc.compileExpr(node.Left)
		if err != nil {
			return err
		}
		index, err := rc.compileExpr(node.Index)
		if err != nil {
			return err
		}
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return err
		}
		rc.emitReg(ROpSetIndex, left, index, val)
		rc.freeReg(left)
		rc.freeReg(index)
		rc.freeReg(val)

	case *ast.SliceAssignStatement:
		left, err := rc.compileExpr(node.Left)
		if err != nil {
			return err
		}
		var startReg, endReg, stepReg int
		if node.Lower != nil {
			startReg, err = rc.compileExpr(node.Lower)
			if err != nil {
				return err
			}
		} else {
			startReg = rc.allocReg()
			rc.emitReg(ROpNull, startReg)
		}
		if node.Upper != nil {
			endReg, err = rc.compileExpr(node.Upper)
			if err != nil {
				return err
			}
		} else {
			endReg = rc.allocReg()
			rc.emitReg(ROpNull, endReg)
		}
		if node.Step != nil {
			stepReg, err = rc.compileExpr(node.Step)
			if err != nil {
				return err
			}
		} else {
			stepReg = rc.allocReg()
			rc.emitReg(ROpNull, stepReg)
		}
		val, err := rc.compileExpr(node.Value)
		if err != nil {
			return err
		}
		rc.emitReg(ROpSetSlice, left, startReg, endReg, stepReg, val)
		rc.freeReg(left)
		rc.freeReg(startReg)
		rc.freeReg(endReg)
		rc.freeReg(stepReg)
		rc.freeReg(val)

	case *ast.BreakStatement:
		if len(rc.loopContexts) == 0 {
			return fmt.Errorf("break statement outside of loop")
		}
		// Emit a jump with sentinel value -2, will be back-patched to after-loop
		rc.emitReg(ROpJump, -2)

	case *ast.ContinueStatement:
		if len(rc.loopContexts) == 0 {
			return fmt.Errorf("continue statement outside of loop")
		}
		lc := rc.loopContexts[len(rc.loopContexts)-1]
		rc.emitReg(ROpJump, lc.continueTarget)

	case *ast.YieldStatement:
		if node.Expression != nil {
			val, err := rc.compileExpr(node.Expression)
			if err != nil {
				return err
			}
			rc.emitReg(ROpYieldValue, val)
			rc.freeReg(val)
		} else {
			nullReg := rc.allocReg()
			rc.emitReg(ROpNull, nullReg)
			rc.emitReg(ROpYieldValue, nullReg)
			rc.freeReg(nullReg)
		}

	default:
		// Try as expression
		reg, err := rc.compileExpr(node)
		if err != nil {
			return err
		}
		rc.freeReg(reg)
	}

	return nil
}

// compileTryStatement compiles a try statement.
func (rc *RegisterCompiler) compileTryStatement(ts *ast.TryStatement) error {
	hasExcept := len(ts.Excepts) > 0
	hasFinally := ts.Finally != nil

	beginTryIdx := len(rc.instructions)
	rc.emitReg(ROpBeginTry, len(ts.Excepts), boolToInt(hasFinally), -1, -1)

	if err := rc.compileStmt(ts.Body); err != nil {
		return err
	}

	// Jump past except handlers (to finally or after try)
	jumpIdx := len(rc.instructions)
	rc.emitReg(ROpJump, -1)

	// Record the IP of the first except handler
	firstHandlerIP := -1
	var exceptJumpIdxs []int

	if hasExcept {
		firstHandlerIP = len(rc.instructions)
		for _, ex := range ts.Excepts {
			var typeIdx int
			if ex.Type != nil {
				if typeStr, ok := ex.Type.(*ast.Identifier); ok {
					typeIdx = rc.addConstant(&objects.String{Value: typeStr.Value})
				} else {
					typeIdx = rc.addConstant(&objects.String{Value: "Exception"})
				}
			} else {
				typeIdx = rc.addConstant(&objects.String{Value: ""})
			}

			var varIdx int
			if ex.Name != nil {
				varIdx = rc.addConstant(&objects.String{Value: ex.Name.Value})
			} else {
				varIdx = rc.addConstant(&objects.String{Value: ""})
			}

			// Support except* handlers
			if ex.IsStar {
				rc.emitReg(ROpExceptStarHandler, typeIdx, varIdx)
			} else {
				rc.emitReg(ROpExceptHandler, typeIdx, varIdx)
			}

			if err := rc.compileStmt(ex.Body); err != nil {
				return err
			}

			// Jump to after try (or to finally) after except body
			ejIdx := len(rc.instructions)
			rc.emitReg(ROpJump, -1)
			exceptJumpIdxs = append(exceptJumpIdxs, ejIdx)
		}
	}

	var finallyStartIP int
	if hasFinally {
		finallyStartIP = len(rc.instructions)
		rc.emitReg(ROpFinally, -1)
		if err := rc.compileStmt(ts.Finally); err != nil {
			return err
		}
	}

	// Back-patch ROpBeginTry with handler IP and finally IP
	rc.instructions[beginTryIdx].Operands[2] = firstHandlerIP
	if hasFinally {
		rc.instructions[beginTryIdx].Operands[3] = finallyStartIP
	}

	afterTryIP := len(rc.instructions)
	rc.emitReg(ROpEndTry)

	// Back-patch the jump after try body
	if hasFinally {
		rc.instructions[jumpIdx].Operands[0] = finallyStartIP
	} else {
		rc.instructions[jumpIdx].Operands[0] = afterTryIP
	}

	// Back-patch the jumps after each except handler body
	for _, ejIdx := range exceptJumpIdxs {
		if hasFinally {
			rc.instructions[ejIdx].Operands[0] = finallyStartIP
		} else {
			rc.instructions[ejIdx].Operands[0] = afterTryIP
		}
	}

	return nil
}

// compileFunction compiles a function literal for class methods.
func (rc *RegisterCompiler) compileFunction(fn *ast.FunctionLiteral) *CompiledFunction {
	outerInstructions := rc.instructions
	outerNextReg := rc.nextReg
	outerFreeRegs := rc.freeRegs
	rc.instructions = make([]RegInstruction, 0)
	rc.nextReg = 0
	rc.freeRegs = nil

	rc.symbolTable = NewEnclosedSymbolTable(rc.symbolTable)
	for _, param := range fn.Parameters {
		rc.symbolTable.Define(param.Value)
	}

	for _, stmt := range fn.Body.Statements {
		if err := rc.compileStmt(stmt); err != nil {
			rc.symbolTable = rc.symbolTable.outer
			rc.instructions = outerInstructions
			rc.nextReg = outerNextReg
			rc.freeRegs = outerFreeRegs
			return nil
		}
	}

	if len(rc.instructions) == 0 || rc.instructions[len(rc.instructions)-1].Opcode != ROpReturn {
		nullReg := rc.allocReg()
		rc.emitReg(ROpNull, nullReg)
		rc.emitReg(ROpReturn, nullReg)
	}

	fnInstructions := rc.instructions
	numLocals := rc.symbolTable.numDefinitions
	free := rc.symbolTable.Free
	rc.symbolTable = rc.symbolTable.outer

	rc.instructions = outerInstructions
	rc.nextReg = outerNextReg
	rc.freeRegs = outerFreeRegs

	return &CompiledFunction{
		Instructions:   regInstructionsToBytes(fnInstructions),
		NumLocals:      numLocals,
		NumParameters:  len(fn.Parameters),
		Free:           free,
	}
}

// Compile compiles a program using the register-based backend.
func (rc *RegisterCompiler) Compile(node ast.Node) error {
	return rc.compileStmt(node)
}

// pushScope creates a new enclosed symbol table scope for nested blocks.
func (rc *RegisterCompiler) pushScope() {
	rc.symbolTable = NewEnclosedSymbolTable(rc.symbolTable)
}

// popScope restores the outer symbol table scope.
func (rc *RegisterCompiler) popScope() {
	if rc.symbolTable.outer != nil {
		rc.symbolTable = rc.symbolTable.outer
	}
}

// emitSetSymbol emits the appropriate set instruction (SetGlobal/SetLocal/SetFree)
// for a given symbol, storing the value from the given register.
func (rc *RegisterCompiler) emitSetSymbol(symbol Symbol, valReg int) {
	switch symbol.Scope {
	case GlobalScope:
		rc.emitReg(ROpSetGlobal, symbol.Index, valReg)
	case FreeScope:
		rc.emitReg(ROpSetLocal, symbol.Index+rc.symbolTable.numDefinitions, valReg)
	default:
		rc.emitReg(ROpSetLocal, symbol.Index, valReg)
	}
}

// augAssignToRegInPlaceOp maps an augmented assignment operator string
// to the corresponding register in-place opcode.
func augAssignToRegInPlaceOp(op string) RegOpcode {
	switch op {
	case "+":
		return ROpInPlaceAdd
	case "-":
		return ROpInPlaceSub
	case "*":
		return ROpInPlaceMul
	case "/":
		return ROpInPlaceDiv
	case "%":
		return ROpInPlaceMod
	case "//":
		return ROpInPlaceFloorDiv
	case "**":
		return ROpInPlacePower
	case "|":
		return ROpInPlaceBitOr
	case "&":
		return ROpInPlaceBitAnd
	case "^":
		return ROpInPlaceBitXor
	case "<<":
		return ROpInPlaceLShift
	case ">>":
		return ROpInPlaceRShift
	default:
		return ROpInPlaceAdd
	}
}

// regInstructionsToBytes serializes register instructions to a byte representation.
// This is a simplified encoding for storage in CompiledFunction.Instructions.
func regInstructionsToBytes(instrs []RegInstruction) []byte {
	var bytes []byte
	for _, instr := range instrs {
		bytes = append(bytes, byte(instr.Opcode))
		for _, op := range instr.Operands {
			bytes = append(bytes, byte(op>>8), byte(op&0xFF))
		}
	}
	return bytes
}

// hasYieldInBody checks if a body contains yield statements.
func hasYieldInBody(node ast.Node) bool {
	switch node := node.(type) {
	case *ast.YieldStatement:
		return true
	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			if hasYieldInBody(stmt) {
				return true
			}
		}
	case *ast.IfExpression:
		if hasYieldInBody(node.Consequence) {
			return true
		}
		if node.Alternative != nil && hasYieldInBody(node.Alternative) {
			return true
		}
	case *ast.WhileStatement:
		if hasYieldInBody(node.Body) {
			return true
		}
	case *ast.FunctionLiteral:
		return false
	}
	return false
}

// registerBuiltins registers the same builtins as the stack-based compiler.
func (rc *RegisterCompiler) registerBuiltins() {
	// Register modules (same as stack-based compiler)
	mathModule := objects.CreateMathModule()
	objects.RegisterModule("math", mathModule)
	mathIndex := len(rc.constants)
	rc.constants = append(rc.constants, mathModule)
	rc.symbolTable.DefineBuiltin("math", mathIndex)

	sysModule := objects.CreateSysModule()
	objects.RegisterModule("sys", sysModule)
	sysIndex := len(rc.constants)
	rc.constants = append(rc.constants, sysModule)
	rc.symbolTable.DefineBuiltin("sys", sysIndex)

	osModule := objects.CreateOsModule()
	objects.RegisterModule("os", osModule)
	osIndex := len(rc.constants)
	rc.constants = append(rc.constants, osModule)
	rc.symbolTable.DefineBuiltin("os", osIndex)

	jsonModule := objects.CreateJsonModule()
	objects.RegisterModule("json", jsonModule)
	jsonIndex := len(rc.constants)
	rc.constants = append(rc.constants, jsonModule)
	rc.symbolTable.DefineBuiltin("json", jsonIndex)

	timeModule := objects.CreateTimeModule()
	objects.RegisterModule("time", timeModule)
	timeIndex := len(rc.constants)
	rc.constants = append(rc.constants, timeModule)
	rc.symbolTable.DefineBuiltin("time", timeIndex)

	// Register common builtins (print, len, range, abs, etc.)
	for _, entry := range GetCommonBuiltins() {
		idx := len(rc.constants)
		if entry.Builtin != nil {
			rc.constants = append(rc.constants, entry.Builtin)
		} else {
			rc.constants = append(rc.constants, entry.Value)
		}
		rc.symbolTable.DefineBuiltin(entry.Name, idx)
	}
}
