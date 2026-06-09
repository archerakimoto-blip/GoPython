package vm

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/jit"
	"github.com/go-py/go-python/pkg/objects"
)

// RegisterVM executes bytecode using a register-based model.
type RegisterVM struct {
	vm *VM // reference back to the shared VM state

	registers []objects.Object // virtual registers
	numRegs   int              // number of allocated registers

	frames     []*RegFrame
	frameIndex int

	// Inline caches
	attrCache     map[AttrCacheKey]AttrCacheEntry
	indexCache    map[IndexCacheKey]IndexCacheEntry
	globalCache   []GlobalCacheEntry
	globalVersions []uint64
}

// RegFrame represents a call frame in the register VM.
type RegFrame struct {
	instructions []RegInstruction
	ip           int
	basePointer  int // stack base pointer (for locals on vm.stack)
	regBase      int // register base offset for frame isolation

	// Reference back to the original stack-based frame data
	fn          *compiler.CompiledFunction
	generator   *objects.Generator
	freeVars    []objects.Object
	initInstance   objects.Object
	setAttrValue   objects.Object
	metaclassInitClass *objects.Class
}

// RegOpcode defines the opcodes for the register-based VM.
type RegOpcode byte

const (
	RegOpLoadConst RegOpcode = iota // reg = constants[idx]
	RegOpMove                       // dst = src
	RegOpAdd                        // dst = src1 + src2
	RegOpSub                        // dst = src1 - src2
	RegOpMul                        // dst = src1 * src2
	RegOpDiv                        // dst = src1 / src2
	RegOpMod                        // dst = src1 % src2
	RegOpCompare                    // dst = src1 <op> src2  (op encoded in 3rd operand)
	RegOpJump                       // ip = target
	RegOpJumpIfFalse                // if !reg: ip = target
	RegOpCall                       // dst = func(args...)  (numArgs in operand)
	RegOpReturn                     // return reg
	RegOpReturnNone                 // return None
	RegOpGetAttr                    // dst = obj.attr  (attr as constant index)
	RegOpSetAttr                    // obj.attr = val  (attr as constant index)
	RegOpGetGlobal                  // reg = globals[idx]
	RegOpSetGlobal                  // globals[idx] = reg
	RegOpGetLocal                   // reg = locals[basePointer+idx]
	RegOpSetLocal                   // locals[basePointer+idx] = reg
	RegOpBuildList                  // dst = list(regs[start:start+n])
	RegOpBuildDict                  // dst = dict from n key-value pairs
	RegOpBuildSet                   // dst = set from n elements
	RegOpIndex                      // dst = obj[idx]
	RegOpSlice                      // dst = obj[start:stop:step]
	RegOpNegate                     // dst = -src
	RegOpNot                        // dst = not src
	RegOpNull                       // reg = None
	RegOpTrue                       // reg = True
	RegOpFalse                      // reg = False
	RegOpPop                        // no-op in register VM (for compatibility)
	RegOpDupTop                     // dst = src (duplicate top)
	RegOpClosure                    // dst = closure(constIdx, numFree, freeRegs...)
	RegOpGetFree                    // reg = freeVars[idx]
	RegOpFloorDiv                   // dst = src1 // src2
	RegOpPower                      // dst = src1 ** src2
	RegOpEllipsis                   // reg = Ellipsis
	RegOpArray                      // dst = list from n regs starting at startReg
	RegOpHash                       // dst = dict from n key-value pairs
	RegOpSet                        // dst = set from n elements
	RegOpListUnpack                 // unpack list into multiple registers
	RegOpDictUnpack                 // no-op placeholder
	RegOpFormatString               // dst = formatted string from n parts
	RegOpCreateClass                // dst = class from constant idx
	RegOpCreateClassWithSuper       // dst = class with super
	RegOpCreateClassWithMultiSuper  // dst = class with multiple supers
	RegOpDelAttribute               // del obj.attr
	RegOpSetClassField              // class.field = val
	RegOpSetMetaclass               // set metaclass on class
	RegOpCallMetaclassInit          // call metaclass __init__
	RegOpMakeGenerator              // dst = generator from fn
	RegOpMakeAsync                  // dst = async from fn
	RegOpAwait                      // await value
	RegOpYieldValue                 // yield value
	RegOpEnterContext               // enter context manager
	RegOpExitContext                // exit context manager
	RegOpBeginTry                   // begin try block
	RegOpEndTry                     // end try block
	RegOpRaise                      // raise exception
	RegOpExceptHandler              // except handler
	RegOpExceptStarHandler          // except* handler
	RegOpFinally                    // finally block

	// Bit operations
	RegOpBitOr   // dst = src1 | src2
	RegOpBitAnd  // dst = src1 & src2
	RegOpBitXor  // dst = src1 ^ src2

	// Set operations
	RegOpSetUnion              // dst = src1 | src2 (set)
	RegOpSetIntersection       // dst = src1 & src2 (set)
	RegOpSetDifference         // dst = src1 - src2 (set)
	RegOpSetSymmetricDifference // dst = src1 ^ src2 (set)

	// In-place operations
	RegOpInPlaceAdd      // dst = src1 += src2
	RegOpInPlaceSub      // dst = src1 -= src2
	RegOpInPlaceMul      // dst = src1 *= src2
	RegOpInPlaceDiv      // dst = src1 /= src2
	RegOpInPlaceMod      // dst = src1 %= src2
	RegOpInPlaceFloorDiv // dst = src1 //= src2
	RegOpInPlacePower    // dst = src1 **= src2
	RegOpInPlaceBitOr    // dst = src1 |= src2
	RegOpInPlaceBitAnd   // dst = src1 &= src2
	RegOpInPlaceBitXor   // dst = src1 ^= src2
	RegOpInPlaceLShift   // dst = src1 <<= src2
	RegOpInPlaceRShift   // dst = src1 >>= src2

	// Index/Slice assignment
	RegOpSetIndex  // obj[index] = value
	RegOpSetSlice  // obj[start:end:step] = value

	// StringBuilder
	RegOpStringBuilderCreate // dst = new StringBuilder
	RegOpStringBuilderAppend // builder.append(value)
	RegOpStringBuilderBuild  // dst = builder.build()

	// Array prealloc
	RegOpArrayPrealloc // dst = list with pre-allocated capacity

	// Shift operations
	RegOpLShift // dst = src1 << src2
	RegOpRShift // dst = src1 >> src2

	// Boolean short-circuit
	RegOpBoolAnd // dst = src1 and src2
	RegOpBoolOr  // dst = src1 or src2

	// Membership
	RegOpContains    // dst = src1 in src2
	RegOpNotContains // dst = src1 not in src2
)

// RegInstruction represents a single register-based instruction.
type RegInstruction struct {
	Opcode   RegOpcode
	Operands []int
}

// newRegisterVM creates a new RegisterVM that shares state with the given VM.
func newRegisterVM(vm *VM) *RegisterVM {
	return &RegisterVM{
		vm:             vm,
		registers:      make([]objects.Object, 512),
		numRegs:        0,
		frames:         make([]*RegFrame, MaxFrames),
		frameIndex:     0,
		attrCache:      make(map[AttrCacheKey]AttrCacheEntry),
		indexCache:     make(map[IndexCacheKey]IndexCacheEntry),
		globalCache:    make([]GlobalCacheEntry, GlobalSize),
		globalVersions: make([]uint64, GlobalSize),
	}
}

// regCallClosure calls a closure with parameter matching.
// It handles default arguments, *args, **kwargs, and keyword-only parameters.
func (rvm *RegisterVM) regCallClosure(closure *objects.Closure, args []objects.Object) (objects.Object, error) {
	vm := rvm.vm
	numParams := closure.NumParameters
	numDefaults := closure.NumDefaults
	numKwOnlyParams := closure.NumKeywordOnly
	hasVarArgs := closure.VarArgs
	hasKwArgs := closure.KwArgs

	// Check minimum required arguments
	minArgs := numParams - numDefaults
	if len(args) < minArgs {
		return nil, fmt.Errorf("TypeError: closure takes at least %d arguments (%d given)", minArgs, len(args))
	}

	// Check maximum arguments (if no *args)
	if !hasVarArgs && len(args) > numParams {
		return nil, fmt.Errorf("TypeError: closure takes at most %d arguments (%d given)", numParams, len(args))
	}

	// Build local variables from parameters
	locals := make([]objects.Object, closure.NumLocals)
	for i := range locals {
		locals[i] = objects.None_
	}

	// Fill positional arguments
	for i := 0; i < numParams && i < len(args); i++ {
		locals[i] = args[i]
	}

	// Fill defaults for missing positional arguments
	for i := len(args); i < numParams; i++ {
		defaultIdx := numParams - numDefaults + (i - (numParams - numDefaults))
		if defaultIdx >= 0 && defaultIdx < closure.NumPositionalDefaults {
			locals[i] = objects.None_ // simplified: no defaults stored on Closure
		}
	}

	// Handle *args
	if hasVarArgs {
		varArgsStart := numParams
		remaining := args[varArgsStart:]
		listElements := make([]objects.Object, len(remaining))
		copy(listElements, remaining)
		locals[varArgsStart] = &objects.List{Elements: listElements}
	}

	// Handle keyword-only parameters
	kwOnlyStart := numParams
	if hasVarArgs {
		kwOnlyStart++
	}
	for i := 0; i < numKwOnlyParams; i++ {
		idx := kwOnlyStart + i
		if idx < len(locals) {
			locals[idx] = objects.None_ // simplified: no kwdefaults stored on Closure
		}
	}

	// Handle **kwargs
	if hasKwArgs {
		kwargsIdx := kwOnlyStart + numKwOnlyParams
		if kwargsIdx < len(locals) {
			locals[kwargsIdx] = &objects.Dict{}
		}
	}

	// Execute the closure's compiled function via stack VM
	vm.push(closure)
	for _, a := range args {
		vm.push(a)
	}
	err := vm.executeCall(len(args))
	if err != nil {
		return nil, err
	}
	return vm.pop(), nil
}

// regSet sets a register value, growing the registers slice if needed.
func (rvm *RegisterVM) currentRegBase() int {
	if rvm.frameIndex > 0 {
		return rvm.frames[rvm.frameIndex-1].regBase
	}
	return 0
}

func (rvm *RegisterVM) regSet(reg int, val objects.Object) {
	actualReg := rvm.currentRegBase() + reg
	if actualReg >= len(rvm.registers) {
		newRegs := make([]objects.Object, actualReg*2+1)
		copy(newRegs, rvm.registers)
		rvm.registers = newRegs
	}
	rvm.registers[actualReg] = val
	if actualReg >= rvm.numRegs {
		rvm.numRegs = actualReg + 1
	}
}

// regGet gets a register value.
func (rvm *RegisterVM) regGet(reg int) objects.Object {
	actualReg := rvm.currentRegBase() + reg
	if actualReg < 0 || actualReg >= len(rvm.registers) {
		return nil
	}
	return rvm.registers[actualReg]
}

// binaryOp performs a binary operation on two objects.
func (rvm *RegisterVM) binaryOp(op compiler.Opcode, left, right objects.Object) (objects.Object, error) {
	leftType := left.Type()
	rightType := right.Type()

	if leftType == objects.INTEGER_OBJ && rightType == objects.INTEGER_OBJ {
		return rvm.binaryIntegerOp(op, left, right)
	}
	if leftType == objects.FLOAT_OBJ && rightType == objects.FLOAT_OBJ {
		return rvm.binaryFloatOp(op, left, right)
	}
	if leftType == objects.INTEGER_OBJ && rightType == objects.FLOAT_OBJ {
		return rvm.binaryFloatOp(op, &objects.Float{Value: float64(left.(*objects.Integer).Value)}, right)
	}
	if leftType == objects.FLOAT_OBJ && rightType == objects.INTEGER_OBJ {
		return rvm.binaryFloatOp(op, left, &objects.Float{Value: float64(right.(*objects.Integer).Value)})
	}

	// Complex number operations
	if leftType == objects.COMPLEX_OBJ && rightType == objects.COMPLEX_OBJ {
		return rvm.binaryComplexOp(op, left, right)
	}
	if leftType == objects.COMPLEX_OBJ && rightType == objects.FLOAT_OBJ {
		return rvm.binaryComplexOp(op, left, objects.NewComplex(right.(*objects.Float).Value, 0))
	}
	if leftType == objects.COMPLEX_OBJ && rightType == objects.INTEGER_OBJ {
		return rvm.binaryComplexOp(op, left, objects.NewComplex(float64(right.(*objects.Integer).Value), 0))
	}
	if leftType == objects.FLOAT_OBJ && rightType == objects.COMPLEX_OBJ {
		return rvm.binaryComplexOp(op, objects.NewComplex(left.(*objects.Float).Value, 0), right)
	}
	if leftType == objects.INTEGER_OBJ && rightType == objects.COMPLEX_OBJ {
		return rvm.binaryComplexOp(op, objects.NewComplex(float64(left.(*objects.Integer).Value), 0), right)
	}

	if op == compiler.OpAdd {
		leftStr := toString(left)
		rightStr := toString(right)
		return &objects.String{Value: leftStr + rightStr}, nil
	}

	return nil, fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (rvm *RegisterVM) binaryIntegerOp(op compiler.Opcode, left, right objects.Object) (objects.Object, error) {
	leftValue := left.(*objects.Integer).Value
	rightValue := right.(*objects.Integer).Value

	var result int64
	switch op {
	case compiler.OpAdd:
		result = leftValue + rightValue
	case compiler.OpSub:
		result = leftValue - rightValue
	case compiler.OpMul:
		result = leftValue * rightValue
	case compiler.OpDiv:
		if rightValue == 0 {
			return objects.NewZeroDivisionError("division by zero"), nil
		}
		result = leftValue / rightValue
	case compiler.OpMod:
		if rightValue == 0 {
			return objects.NewZeroDivisionError("modulo by zero"), nil
		}
		result = leftValue % rightValue
	case compiler.OpFloorDiv:
		if rightValue == 0 {
			return objects.NewZeroDivisionError("floor division by zero"), nil
		}
		result = leftValue / rightValue
		if leftValue < 0 && leftValue%rightValue != 0 {
			result -= 1
		}
	case compiler.OpPower:
		if rightValue < 0 {
			return nil, fmt.Errorf("negative exponent not supported for integers")
		}
		result = 1
		for i := int64(0); i < rightValue; i++ {
			result *= leftValue
		}
	default:
		return nil, fmt.Errorf("unknown integer operator: %d", op)
	}

	return objects.GetCachedInteger(result), nil
}

func (rvm *RegisterVM) binaryFloatOp(op compiler.Opcode, left, right objects.Object) (objects.Object, error) {
	leftValue := left.(*objects.Float).Value
	rightValue := right.(*objects.Float).Value

	var result float64
	switch op {
	case compiler.OpAdd:
		result = leftValue + rightValue
	case compiler.OpSub:
		result = leftValue - rightValue
	case compiler.OpMul:
		result = leftValue * rightValue
	case compiler.OpDiv:
		if rightValue == 0 {
			return nil, fmt.Errorf("float division by zero")
		}
		result = leftValue / rightValue
	case compiler.OpMod:
		result = math.Mod(leftValue, rightValue)
	case compiler.OpFloorDiv:
		result = math.Floor(leftValue / rightValue)
	case compiler.OpPower:
		result = math.Pow(leftValue, rightValue)
	default:
		return nil, fmt.Errorf("unknown float operator: %d", op)
	}

	return &objects.Float{Value: result}, nil
}

func (rvm *RegisterVM) binaryComplexOp(op compiler.Opcode, left, right objects.Object) (objects.Object, error) {
	leftComplex := left.(*objects.Complex)
	rightComplex := right.(*objects.Complex)

	lc := complex(leftComplex.Real, leftComplex.Imag)
	rc := complex(rightComplex.Real, rightComplex.Imag)

	switch op {
	case compiler.OpAdd:
		result := lc + rc
		return objects.NewComplex(real(result), imag(result)), nil
	case compiler.OpSub:
		result := lc - rc
		return objects.NewComplex(real(result), imag(result)), nil
	case compiler.OpMul:
		result := lc * rc
		return objects.NewComplex(real(result), imag(result)), nil
	case compiler.OpDiv:
		if rc == 0 {
			return objects.NewZeroDivisionError("complex division by zero"), nil
		}
		result := lc / rc
		return objects.NewComplex(real(result), imag(result)), nil
	case compiler.OpPower:
		result := cmplx.Pow(lc, rc)
		return objects.NewComplex(real(result), imag(result)), nil
	default:
		return nil, fmt.Errorf("unknown complex operator: %d", op)
	}
}

// compareOp performs a comparison operation.
func (rvm *RegisterVM) compareOp(op compiler.Opcode, left, right objects.Object) (objects.Object, error) {
	if leftEm, ok := left.(*objects.EnumMember); ok {
		left = leftEm.Value
	}
	if rightEm, ok := right.(*objects.EnumMember); ok {
		right = rightEm.Value
	}

	if left.Type() == objects.INTEGER_OBJ && right.Type() == objects.INTEGER_OBJ {
		leftValue := left.(*objects.Integer).Value
		rightValue := right.(*objects.Integer).Value
		switch op {
		case compiler.OpEqual:
			return nativeBoolToBooleanObject(leftValue == rightValue), nil
		case compiler.OpNotEqual:
			return nativeBoolToBooleanObject(leftValue != rightValue), nil
		case compiler.OpGreaterThan:
			return nativeBoolToBooleanObject(leftValue > rightValue), nil
		case compiler.OpLessThan:
			return nativeBoolToBooleanObject(leftValue < rightValue), nil
		case compiler.OpGreaterEqual:
			return nativeBoolToBooleanObject(leftValue >= rightValue), nil
		case compiler.OpLessEqual:
			return nativeBoolToBooleanObject(leftValue <= rightValue), nil
		}
	}

	if left.Type() == objects.FLOAT_OBJ && right.Type() == objects.FLOAT_OBJ {
		leftValue := left.(*objects.Float).Value
		rightValue := right.(*objects.Float).Value
		switch op {
		case compiler.OpEqual:
			return nativeBoolToBooleanObject(leftValue == rightValue), nil
		case compiler.OpNotEqual:
			return nativeBoolToBooleanObject(leftValue != rightValue), nil
		case compiler.OpGreaterThan:
			return nativeBoolToBooleanObject(leftValue > rightValue), nil
		case compiler.OpLessThan:
			return nativeBoolToBooleanObject(leftValue < rightValue), nil
		case compiler.OpGreaterEqual:
			return nativeBoolToBooleanObject(leftValue >= rightValue), nil
		case compiler.OpLessEqual:
			return nativeBoolToBooleanObject(leftValue <= rightValue), nil
		}
	}

	if left.Type() == objects.COMPLEX_OBJ && right.Type() == objects.COMPLEX_OBJ {
		lc := left.(*objects.Complex)
		rc := right.(*objects.Complex)
		switch op {
		case compiler.OpEqual:
			return nativeBoolToBooleanObject(lc.Real == rc.Real && lc.Imag == rc.Imag), nil
		case compiler.OpNotEqual:
			return nativeBoolToBooleanObject(lc.Real != rc.Real || lc.Imag != rc.Imag), nil
		default:
			return nil, fmt.Errorf("'%s' not supported between instances of 'complex' and 'complex'", opToString(op))
		}
	}

	switch op {
	case compiler.OpEqual:
		return nativeBoolToBooleanObject(objects.Equal(left, right)), nil
	case compiler.OpNotEqual:
		return nativeBoolToBooleanObject(!objects.Equal(left, right)), nil
	case compiler.OpGreaterThan:
		return nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() > right.Inspect()), nil
	case compiler.OpLessThan:
		return nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() < right.Inspect()), nil
	case compiler.OpGreaterEqual:
		return nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() >= right.Inspect()), nil
	case compiler.OpLessEqual:
		return nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() <= right.Inspect()), nil
	}

	return nil, fmt.Errorf("unknown comparison operator: %d", op)
}

// negateOp performs the unary minus operation.
func (rvm *RegisterVM) negateOp(operand objects.Object) (objects.Object, error) {
	switch op := operand.(type) {
	case *objects.Integer:
		return objects.GetCachedInteger(-op.Value), nil
	case *objects.Float:
		return &objects.Float{Value: -op.Value}, nil
	case *objects.Complex:
		return objects.NewComplex(-op.Real, -op.Imag), nil
	default:
		return nil, fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
}

// notOp performs the unary not operation.
func (rvm *RegisterVM) notOp(operand objects.Object) objects.Object {
	switch operand {
	case objects.True:
		return objects.False
	case objects.False:
		return objects.True
	case objects.None_:
		return objects.True
	default:
		return objects.False
	}
}

// regContains performs the "in" membership test.
func (rvm *RegisterVM) regContains(element, container objects.Object) objects.Object {
	switch c := container.(type) {
	case *objects.List:
		for _, item := range c.Elements {
			if objects.Equal(element, item) {
				return objects.True
			}
		}
		return objects.False
	case *objects.Tuple:
		for _, item := range c.Elements {
			if objects.Equal(element, item) {
				return objects.True
			}
		}
		return objects.False
	case *objects.Set:
		for _, item := range c.Elements {
			if objects.Equal(element, item) {
				return objects.True
			}
		}
		return objects.False
	case *objects.Dict:
		keyStr := element.Inspect()
		if _, ok := c.Pairs[keyStr]; ok {
			return objects.True
		}
		return objects.False
	case *objects.String:
		if str, ok := element.(*objects.String); ok {
			for i := 0; i <= len(c.Value)-len(str.Value); i++ {
				if c.Value[i:i+len(str.Value)] == str.Value {
					return objects.True
				}
			}
			return objects.False
		}
		return objects.False
	default:
		return objects.False
	}
}

// indexOp performs an index operation.
func (rvm *RegisterVM) indexOp(left, index objects.Object) (objects.Object, error) {
	switch {
	case left.Type() == objects.LIST_OBJ && index.Type() == objects.INTEGER_OBJ:
		arrayObject := left.(*objects.List)
		idx := index.(*objects.Integer).Value
		max := int64(len(arrayObject.Elements) - 1)
		if idx < 0 || idx > max {
			return objects.None_, nil
		}
		return arrayObject.Elements[idx], nil

	case left.Type() == objects.TUPLE_OBJ && index.Type() == objects.INTEGER_OBJ:
		tupleObject := left.(*objects.Tuple)
		idx := index.(*objects.Integer).Value
		max := int64(len(tupleObject.Elements) - 1)
		if idx < 0 {
			idx = int64(len(tupleObject.Elements)) + idx
		}
		if idx < 0 || idx > max {
			return objects.None_, nil
		}
		return tupleObject.Elements[idx], nil

	case left.Type() == objects.DICT_OBJ:
		hashObject := left.(*objects.Dict)
		value, ok := hashObject.Get(index)
		if !ok {
			return objects.None_, nil
		}
		return value, nil

	case left.Type() == objects.RANGE_OBJ && index.Type() == objects.INTEGER_OBJ:
		rangeObj := left.(*objects.Range)
		val, ok := rangeObj.GetItem(index.(*objects.Integer).Value)
		if !ok {
			return objects.None_, nil
		}
		return val, nil

	case left.Type() == objects.STRING_OBJ && index.Type() == objects.INTEGER_OBJ:
		str := left.(*objects.String)
		idx := index.(*objects.Integer).Value
		length := int64(len(str.Value))
		if idx < 0 {
			idx = length + idx
		}
		if idx < 0 || idx >= length {
			return objects.None_, nil
		}
		return &objects.String{Value: string(str.Value[idx])}, nil

	case left.Type() == objects.BYTES_OBJ && index.Type() == objects.INTEGER_OBJ:
		bts := left.(*objects.Bytes)
		idx := index.(*objects.Integer).Value
		length := int64(len(bts.Value))
		if idx < 0 {
			idx = length + idx
		}
		if idx < 0 || idx >= length {
			return objects.None_, nil
		}
		return objects.GetCachedInteger(int64(bts.Value[idx])), nil

	default:
		return nil, fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

// sliceOp performs a slice operation.
func (rvm *RegisterVM) sliceOpWithStep(left, start, end, step objects.Object) (objects.Object, error) {
	// Parse step value
	var stepVal int64 = 1
	if step != nil && step != objects.None_ {
		if s, ok := step.(*objects.Integer); ok {
			stepVal = s.Value
		}
	}
	if stepVal == 0 {
		return nil, fmt.Errorf("slice step cannot be zero")
	}

	switch {
	case left.Type() == objects.LIST_OBJ:
		return rvm.listSliceWithStep(left, start, end, stepVal)
	case left.Type() == objects.STRING_OBJ:
		return rvm.stringSliceWithStep(left, start, end, stepVal)
	case left.Type() == objects.BYTES_OBJ:
		return rvm.bytesSliceWithStep(left, start, end, stepVal)
	case left.Type() == objects.TUPLE_OBJ:
		return rvm.tupleSliceWithStep(left, start, end, stepVal)
	default:
		return nil, fmt.Errorf("slice operator not supported: %s", left.Type())
	}
}

func normalizeIndex(idx, length int64) int64 {
	if idx < 0 {
		idx += length
		if idx < 0 {
			idx = 0
		}
	}
	if idx > length {
		idx = length
	}
	return idx
}

func sliceBounds(length, startIdx, endIdx, step int64) (int64, int64) {
	if step > 0 {
		startIdx = normalizeIndex(startIdx, length)
		endIdx = normalizeIndex(endIdx, length)
		if startIdx > endIdx {
			endIdx = startIdx
		}
	} else {
		startIdx = normalizeIndex(startIdx, length)
		endIdx = normalizeIndex(endIdx, length)
		if startIdx >= length {
			startIdx = length - 1
		}
		if endIdx < -1 {
			endIdx = -1
		}
	}
	return startIdx, endIdx
}

func (rvm *RegisterVM) listSliceWithStep(left, start, end objects.Object, step int64) (objects.Object, error) {
	list := left.(*objects.List)
	length := int64(len(list.Elements))

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
	default:
		if step > 0 {
			startIdx = 0
		} else {
			startIdx = length - 1
		}
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		endIdx = e.Value
	default:
		if step > 0 {
			endIdx = length
		} else {
			endIdx = -1
		}
	}

	startIdx, endIdx = sliceBounds(length, startIdx, endIdx, step)

	var elements []objects.Object
	if step > 0 {
		for i := startIdx; i < endIdx; i += step {
			elements = append(elements, list.Elements[i])
		}
	} else {
		for i := startIdx; i > endIdx; i += step {
			elements = append(elements, list.Elements[i])
		}
	}
	return &objects.List{Elements: elements}, nil
}

func (rvm *RegisterVM) stringSliceWithStep(left, start, end objects.Object, step int64) (objects.Object, error) {
	str := left.(*objects.String)
	length := int64(len(str.Value))

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
	default:
		if step > 0 {
			startIdx = 0
		} else {
			startIdx = length - 1
		}
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		endIdx = e.Value
	default:
		if step > 0 {
			endIdx = length
		} else {
			endIdx = -1
		}
	}

	startIdx, endIdx = sliceBounds(length, startIdx, endIdx, step)

	var result []rune
	runes := []rune(str.Value)
	if step > 0 {
		for i := startIdx; i < endIdx; i += step {
			result = append(result, runes[i])
		}
	} else {
		for i := startIdx; i > endIdx; i += step {
			result = append(result, runes[i])
		}
	}
	return &objects.String{Value: string(result)}, nil
}

func (rvm *RegisterVM) bytesSliceWithStep(left, start, end objects.Object, step int64) (objects.Object, error) {
	b := left.(*objects.Bytes)
	length := int64(len(b.Value))

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
	default:
		if step > 0 {
			startIdx = 0
		} else {
			startIdx = length - 1
		}
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		endIdx = e.Value
	default:
		if step > 0 {
			endIdx = length
		} else {
			endIdx = -1
		}
	}

	startIdx, endIdx = sliceBounds(length, startIdx, endIdx, step)

	var result []byte
	if step > 0 {
		for i := startIdx; i < endIdx; i += step {
			result = append(result, b.Value[i])
		}
	} else {
		for i := startIdx; i > endIdx; i += step {
			result = append(result, b.Value[i])
		}
	}
	return &objects.Bytes{Value: result}, nil
}

func (rvm *RegisterVM) tupleSliceWithStep(left, start, end objects.Object, step int64) (objects.Object, error) {
	tuple := left.(*objects.Tuple)
	length := int64(len(tuple.Elements))

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
	default:
		if step > 0 {
			startIdx = 0
		} else {
			startIdx = length - 1
		}
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		endIdx = e.Value
	default:
		if step > 0 {
			endIdx = length
		} else {
			endIdx = -1
		}
	}

	startIdx, endIdx = sliceBounds(length, startIdx, endIdx, step)

	var elements []objects.Object
	if step > 0 {
		for i := startIdx; i < endIdx; i += step {
			elements = append(elements, tuple.Elements[i])
		}
	} else {
		for i := startIdx; i > endIdx; i += step {
			elements = append(elements, tuple.Elements[i])
		}
	}
	return &objects.Tuple{Elements: elements}, nil
}

func (rvm *RegisterVM) sliceOp(left, start, end objects.Object) (objects.Object, error) {
	switch {
	case left.Type() == objects.LIST_OBJ:
		return rvm.listSlice(left, start, end)
	case left.Type() == objects.STRING_OBJ:
		return rvm.stringSlice(left, start, end)
	case left.Type() == objects.BYTES_OBJ:
		return rvm.bytesSlice(left, start, end)
	default:
		return nil, fmt.Errorf("slice operator not supported: %s", left.Type())
	}
}

func (rvm *RegisterVM) listSlice(left, start, end objects.Object) (objects.Object, error) {
	list := left.(*objects.List)
	length := len(list.Elements)

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
		if startIdx < 0 {
			startIdx = int64(length) + startIdx
		}
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > int64(length) {
			startIdx = int64(length)
		}
	default:
		startIdx = 0
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		if e.Value == -1 {
			endIdx = int64(length)
		} else {
			endIdx = e.Value
			if endIdx < 0 {
				endIdx = int64(length) + endIdx
			}
			if endIdx < 0 {
				endIdx = 0
			}
			if endIdx > int64(length) {
				endIdx = int64(length)
			}
		}
	default:
		endIdx = int64(length)
	}

	if startIdx > endIdx {
		return &objects.List{Elements: []objects.Object{}}, nil
	}

	elements := make([]objects.Object, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		elements[i-startIdx] = list.Elements[i]
	}
	return &objects.List{Elements: elements}, nil
}

func (rvm *RegisterVM) stringSlice(left, start, end objects.Object) (objects.Object, error) {
	str := left.(*objects.String)
	length := len(str.Value)

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
		if startIdx < 0 {
			startIdx = int64(length) + startIdx
		}
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > int64(length) {
			startIdx = int64(length)
		}
	default:
		startIdx = 0
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		if e.Value == -1 {
			endIdx = int64(length)
		} else {
			endIdx = e.Value
			if endIdx < 0 {
				endIdx = int64(length) + endIdx
			}
			if endIdx < 0 {
				endIdx = 0
			}
			if endIdx > int64(length) {
				endIdx = int64(length)
			}
		}
	default:
		endIdx = int64(length)
	}

	if startIdx > endIdx {
		return &objects.String{Value: ""}, nil
	}
	return &objects.String{Value: str.Value[startIdx:endIdx]}, nil
}

func (rvm *RegisterVM) bytesSlice(left, start, end objects.Object) (objects.Object, error) {
	bts := left.(*objects.Bytes)
	length := len(bts.Value)

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
		if startIdx < 0 {
			startIdx = int64(length) + startIdx
		}
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > int64(length) {
			startIdx = int64(length)
		}
	default:
		startIdx = 0
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		if e.Value == -1 {
			endIdx = int64(length)
		} else {
			endIdx = e.Value
			if endIdx < 0 {
				endIdx = int64(length) + endIdx
			}
			if endIdx < 0 {
				endIdx = 0
			}
			if endIdx > int64(length) {
				endIdx = int64(length)
			}
		}
	default:
		endIdx = int64(length)
	}

	if startIdx > endIdx {
		return &objects.Bytes{Value: []byte{}}, nil
	}
	sliced := make([]byte, endIdx-startIdx)
	copy(sliced, bts.Value[startIdx:endIdx])
	return &objects.Bytes{Value: sliced}, nil
}

// getAttrOp performs attribute access. Falls back to stack VM for complex cases.
func (rvm *RegisterVM) getAttrOp(obj objects.Object, attrName string) (objects.Object, error) {
	vm := rvm.vm

	if instance, ok := obj.(*objects.Instance); ok {
		if attrName == "__class__" {
			return instance.Class, nil
		}
		if instance.Class != nil {
			if classAttr, ok := instance.Class.FindClassAttr(attrName); ok {
				if prop, ok := classAttr.(*objects.Property); ok {
					if prop.Fget != nil && prop.Fget != objects.None_ {
						// Fall back to stack for property access
						vm.push(instance)
						vm.push(prop.Fget)
						vm.executeCall(0)
						return vm.pop(), nil
					}
				}
				if cm, ok := classAttr.(*objects.ClassMethod); ok {
					vm.push(instance.Class)
					vm.push(cm.Fn)
					return vm.pop(), nil // return the method, caller handles it
				}
				if sm, ok := classAttr.(*objects.StaticMethod); ok {
					return sm.Fn, nil
				}
			}
		}
		if instance.IsSlottedInstance() {
			if idx := instance.GetSlotIndex(attrName); idx >= 0 && idx < len(instance.SlotValues) {
				if instance.SlotValues[idx] != nil {
					return instance.SlotValues[idx], nil
				}
				// Slot exists but value is nil - return None for uninitialized slots
				return objects.None_, nil
			}
		}
		if val, ok := instance.Fields[attrName]; ok {
			return val, nil
		}
		if instance.Class != nil {
			if classAttr, ok := instance.Class.FindClassAttr(attrName); ok {
				if method, ok := classAttr.(*compiler.CompiledFunction); ok {
					// For methods, we need to return both instance and method
					// This is complex - fall back to stack
					vm.push(instance)
					vm.push(method)
					return vm.pop(), nil
				}
				return classAttr, nil
			}
		}
		return objects.None_, nil
	}

	if classObj, ok := obj.(*objects.Class); ok {
		if attrName == "__name__" {
			return &objects.String{Value: classObj.Name}, nil
		}
		if classAttr, ok := classObj.FindClassAttr(attrName); ok {
			if sm, ok := classAttr.(*objects.StaticMethod); ok {
				return sm.Fn, nil
			}
			return classAttr, nil
		}
		return objects.None_, nil
	}

	if module, ok := obj.(*objects.Module); ok {
		if val, ok := module.GetAttr(attrName); ok {
			return val, nil
		}
		return objects.None_, nil
	}

	if enumObj, ok := obj.(*objects.Enum); ok {
		if val, ok := enumObj.GetAttr(attrName); ok {
			return val, nil
		}
		return objects.None_, nil
	}

	if superObj, ok := obj.(*objects.Super); ok {
		if superObj.SuperClass != nil {
			if classAttr, ok := superObj.SuperClass.FindClassAttr(attrName); ok {
				if method, ok := classAttr.(*compiler.CompiledFunction); ok {
					vm.push(superObj.Instance)
					vm.push(method)
					return vm.pop(), nil
				}
				return classAttr, nil
			}
		}
		return objects.None_, nil
	}

	if dictObj, ok := obj.(*objects.Dict); ok {
		switch attrName {
		case "keys":
			return objects.NewDictKeys(dictObj), nil
		case "values":
			return objects.NewDictValues(dictObj), nil
		case "items":
			return objects.NewDictItems(dictObj), nil
		}
		return objects.None_, nil
	}

	return objects.None_, nil
}

// setAttrOp performs attribute setting.
func (rvm *RegisterVM) setAttrOp(obj, val objects.Object, attrName string) error {
	vm := rvm.vm

	if instance, ok := obj.(*objects.Instance); ok {
		if instance.Class != nil {
			classAttr, found := instance.Class.FindClassAttr(attrName)
			if found {
				if prop, ok := classAttr.(*objects.Property); ok {
					if prop.Fset != nil && prop.Fset != objects.None_ {
						vm.push(instance)
						vm.push(prop.Fset)
						vm.push(val)
						vm.executeCall(1)
						vm.currentFrame().setAttrValue = val
						return nil
					}
				}
			}
		}
		vm.attrCache = make(map[AttrCacheKey]AttrCacheEntry)
		if instance.Class != nil && instance.Class.HasSlots() && !instance.Class.IsSlotAllowed(attrName) {
			errObj := objects.NewAttributeError("'%s' object has no attribute '%s'", instance.Class.Name, attrName)
			caught := vm.raiseException(errObj)
			if !caught {
				return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
			}
			return nil
		}
		instance.SetAttr(attrName, val)
		vm.push(val)
		return nil
	}

	if classObj, ok := obj.(*objects.Class); ok {
		classObj.Fields[attrName] = val
		if attrName == "__slots__" {
			if listObj, ok := val.(*objects.List); ok {
				slots := make([]string, 0, len(listObj.Elements))
				for _, elem := range listObj.Elements {
					if strElem, ok := elem.(*objects.String); ok {
						slots = append(slots, strElem.Value)
					}
				}
				classObj.Slots = slots
			}
		}
		vm.push(val)
		return nil
	}

	return fmt.Errorf("cannot set attribute on non-instance: %s", obj.Type())
}

// regBinaryOp performs a binary operation for the register VM, handling
// bit and set operations that are not covered by the existing binaryOp.
func (rvm *RegisterVM) regBinaryOp(op compiler.Opcode, left, right objects.Object) (objects.Object, error) {
	// Integer bit operations
	if left.Type() == objects.INTEGER_OBJ && right.Type() == objects.INTEGER_OBJ {
		leftVal := left.(*objects.Integer).Value
		rightVal := right.(*objects.Integer).Value
		switch op {
		case compiler.OpBitOr:
			return objects.GetCachedInteger(leftVal | rightVal), nil
		case compiler.OpBitAnd:
			return objects.GetCachedInteger(leftVal & rightVal), nil
		case compiler.OpBitXor:
			return objects.GetCachedInteger(leftVal ^ rightVal), nil
		}
	}

	// Boolean bit operations (treated as integers 0/1)
	if left.Type() == objects.BOOLEAN_OBJ && right.Type() == objects.BOOLEAN_OBJ {
		leftVal := int64(0)
		rightVal := int64(0)
		if left.(*objects.Boolean).Value {
			leftVal = 1
		}
		if right.(*objects.Boolean).Value {
			rightVal = 1
		}
		switch op {
		case compiler.OpBitOr:
			return objects.GetCachedInteger(leftVal | rightVal), nil
		case compiler.OpBitAnd:
			return objects.GetCachedInteger(leftVal & rightVal), nil
		case compiler.OpBitXor:
			return objects.GetCachedInteger(leftVal ^ rightVal), nil
		}
	}

	// Set operations
	if left.Type() == objects.SET_OBJ && right.Type() == objects.SET_OBJ {
		leftSet := left.(*objects.Set)
		rightSet := right.(*objects.Set)
		switch op {
		case compiler.OpSetUnion:
			result := objects.NewSet()
			for _, elem := range leftSet.ToSlice() {
				result.Add(elem)
			}
			for _, elem := range rightSet.ToSlice() {
				result.Add(elem)
			}
			return result, nil
		case compiler.OpSetIntersection:
			result := objects.NewSet()
			for _, elem := range leftSet.ToSlice() {
				if rightSet.Contains(elem) {
					result.Add(elem)
				}
			}
			return result, nil
		case compiler.OpSetDifference:
			result := objects.NewSet()
			for _, elem := range leftSet.ToSlice() {
				if !rightSet.Contains(elem) {
					result.Add(elem)
				}
			}
			return result, nil
		case compiler.OpSetSymmetricDifference:
			result := objects.NewSet()
			for _, elem := range leftSet.ToSlice() {
				if !rightSet.Contains(elem) {
					result.Add(elem)
				}
			}
			for _, elem := range rightSet.ToSlice() {
				if !leftSet.Contains(elem) {
					result.Add(elem)
				}
			}
			return result, nil
		}
	}

	// Fall back to existing binaryOp for other types
	return rvm.binaryOp(op, left, right)
}

// regExecuteSetIndex implements obj[index] = value for the register VM.
func (rvm *RegisterVM) regExecuteSetIndex(obj, index, value objects.Object) error {
	switch {
	case obj.Type() == objects.LIST_OBJ && index.Type() == objects.INTEGER_OBJ:
		list := obj.(*objects.List)
		idx := index.(*objects.Integer).Value
		if idx < 0 {
			idx = int64(len(list.Elements)) + idx
		}
		if idx < 0 || idx >= int64(len(list.Elements)) {
			return fmt.Errorf("list index out of range")
		}
		list.Elements[idx] = value

	case obj.Type() == objects.DICT_OBJ:
		dict := obj.(*objects.Dict)
		if err := objects.CheckHashable(index); err != nil {
			return err
		}
		dict.Set(index, value)

	default:
		// Try __setitem__ method
		if getter, ok := obj.(objects.AttributeGetter); ok {
			if method, found := getter.GetAttr("__setitem__"); found {
				if builtin, ok := method.(*objects.Builtin); ok {
					result := builtin.Fn(index, value)
					if result.Type() == objects.ERROR_OBJ {
						return fmt.Errorf("%s", result.Inspect())
					}
					return nil
				}
			}
		}
		return fmt.Errorf("'%s' object does not support item assignment", obj.Type())
	}
	return nil
}

// regExecuteSetSlice implements obj[start:end:step] = value for the register VM.
func (rvm *RegisterVM) regExecuteSetSlice(obj, start, end, step, value objects.Object) error {
	if obj.Type() != objects.LIST_OBJ {
		return fmt.Errorf("'%s' object does not support slice assignment", obj.Type())
	}

	list := obj.(*objects.List)
	replacement, ok := value.(*objects.List)
	if !ok {
		return fmt.Errorf("can only assign list to slice")
	}

	var startIdx int64
	switch s := start.(type) {
	case *objects.Integer:
		startIdx = s.Value
		if startIdx < 0 {
			startIdx = int64(len(list.Elements)) + startIdx
		}
		if startIdx < 0 {
			startIdx = 0
		}
	default:
		startIdx = 0
	}

	var endIdx int64
	switch e := end.(type) {
	case *objects.Integer:
		endIdx = e.Value
		if endIdx < 0 {
			endIdx = int64(len(list.Elements)) + endIdx
		}
		if endIdx < 0 {
			endIdx = 0
		}
	default:
		endIdx = int64(len(list.Elements))
	}

	if startIdx > int64(len(list.Elements)) {
		startIdx = int64(len(list.Elements))
	}
	if endIdx > int64(len(list.Elements)) {
		endIdx = int64(len(list.Elements))
	}

	// Simple slice replacement (step=1 or no step)
	newElements := make([]objects.Object, 0, startIdx+int64(len(replacement.Elements))+(int64(len(list.Elements))-endIdx))
	newElements = append(newElements, list.Elements[:startIdx]...)
	newElements = append(newElements, replacement.Elements...)
	newElements = append(newElements, list.Elements[endIdx:]...)
	list.Elements = newElements

	return nil
}

// NewRegisterVMWithBytecode creates a RegisterVM that directly executes
// register bytecode produced by the RegisterCompiler, without the
// translate() step. The globals slice allows sharing state (e.g., for REPL).
func NewRegisterVMWithBytecode(regBytecode *compiler.RegBytecode, globals []objects.Object) *RegisterVM {
	// Create a minimal VM to hold shared state
	vm := &VM{
		constants:      regBytecode.Constants,
		stack:          make([]objects.Object, StackSize),
		sp:             0,
		globals:        globals,
		frames:         make([]*Frame, MaxFrames),
		framesIndex:    0,
		attrCache:      make(map[AttrCacheKey]AttrCacheEntry),
		globalCache:    make([]GlobalCacheEntry, GlobalSize),
		globalVersions: make([]uint64, GlobalSize),
		indexCache:     make(map[IndexCacheKey]IndexCacheEntry),
		jitEngine:      jit.New(),
	}

	rvm := &RegisterVM{
		vm:         vm,
		registers:  make([]objects.Object, regBytecode.NumRegs+64),
		numRegs:    0,
		frames:     make([]*RegFrame, MaxFrames),
		frameIndex: 0,
	}

	// Store the register bytecode instructions as the main frame
	mainFrame := &RegFrame{
		instructions: convertCompilerRegInstructions(regBytecode.Instructions),
		ip:           -1,
		basePointer:  0,
		regBase:      0,
	}
	rvm.frames[0] = mainFrame
	rvm.frameIndex = 1

	// Set constants on the VM
	vm.constants = regBytecode.Constants

	return rvm
}

// convertCompilerRegInstructions converts compiler.RegInstruction to vm.RegInstruction.
// Since both types have the same structure but different types (compiler.RegOpcode vs vm.RegOpcode),
// we need to map between them.
// compilerToVMOpcode maps compiler register opcode values to VM register opcode values.
// The compiler and VM define opcodes in different orders, so we need an explicit mapping.
// Compiler opcodes (from register_compiler.go):
//   0=LoadConst 1=Move 2=Add 3=Sub 4=Mul 5=Div 6=Mod 7=FloorDiv 8=Power
//   9=Compare 10=Jump 11=JumpIfFalse 12=Call 13=Return 14=ReturnNone
//   15=GetAttr 16=SetAttr 17=GetGlobal 18=SetGlobal 19=GetLocal 20=SetLocal
//   21=BuildList 22=BuildDict 23=BuildSet 24=Index 25=Slice
//   26=Negate 27=Not 28=Null 29=True 30=False
//   31=BitOr 32=BitAnd 33=BitXor 34=SetUnion 35=SetIntersection 36=SetDifference 37=SetSymmetricDifference
//   38-49=InPlace ops 50=SetIndex 51=SetSlice 52-55=StringBuilder/ArrayPrealloc
//   56=Closure 57=GetFree 58=Ellipsis 59=Pop 60=DupTop 61=Raise
//   62=BeginTry 63=EndTry 64=ExceptHandler 65=Finally 66=EnterContext 67=ExitContext
//   68=MakeGenerator 69=MakeAsync 70=Await 71=YieldValue
//   72=CreateClass 73=CreateClassWithSuper 74=CreateClassWithMultiSuper
//   75=DelAttribute 76=SetClassField 77=SetMetaclass 78=CallMetaclassInit
//   79=FormatString 80=ListUnpack 81=DictUnpack 82=ExceptStarHandler
var compilerToVMOpcode = map[byte]RegOpcode{
	0:  RegOpLoadConst,
	1:  RegOpMove,
	2:  RegOpAdd,
	3:  RegOpSub,
	4:  RegOpMul,
	5:  RegOpDiv,
	6:  RegOpMod,
	7:  RegOpFloorDiv,
	8:  RegOpPower,
	9:  RegOpCompare,
	10: RegOpJump,
	11: RegOpJumpIfFalse,
	12: RegOpCall,
	13: RegOpReturn,
	14: RegOpReturnNone,
	15: RegOpGetAttr,
	16: RegOpSetAttr,
	17: RegOpGetGlobal,
	18: RegOpSetGlobal,
	19: RegOpGetLocal,
	20: RegOpSetLocal,
	21: RegOpBuildList,
	22: RegOpBuildDict,
	23: RegOpBuildSet,
	24: RegOpIndex,
	25: RegOpSlice,
	26: RegOpNegate,
	27: RegOpNot,
	28: RegOpNull,
	29: RegOpTrue,
	30: RegOpFalse,
	31: RegOpBitOr,
	32: RegOpBitAnd,
	33: RegOpBitXor,
	34: RegOpSetUnion,
	35: RegOpSetIntersection,
	36: RegOpSetDifference,
	37: RegOpSetSymmetricDifference,
	38: RegOpInPlaceAdd,
	39: RegOpInPlaceSub,
	40: RegOpInPlaceMul,
	41: RegOpInPlaceDiv,
	42: RegOpInPlaceMod,
	43: RegOpInPlaceFloorDiv,
	44: RegOpInPlacePower,
	45: RegOpInPlaceBitOr,
	46: RegOpInPlaceBitAnd,
	47: RegOpInPlaceBitXor,
	48: RegOpInPlaceLShift,
	49: RegOpInPlaceRShift,
	50: RegOpSetIndex,
	51: RegOpSetSlice,
	52: RegOpStringBuilderCreate,
	53: RegOpStringBuilderAppend,
	54: RegOpStringBuilderBuild,
	55: RegOpArrayPrealloc,
	56: RegOpClosure,
	57: RegOpGetFree,
	58: RegOpEllipsis,
	59: RegOpPop,
	60: RegOpDupTop,
	61: RegOpRaise,
	62: RegOpBeginTry,
	63: RegOpEndTry,
	64: RegOpExceptHandler,
	65: RegOpFinally,
	66: RegOpEnterContext,
	67: RegOpExitContext,
	68: RegOpMakeGenerator,
	69: RegOpMakeAsync,
	70: RegOpAwait,
	71: RegOpYieldValue,
	72: RegOpCreateClass,
	73: RegOpCreateClassWithSuper,
	74: RegOpCreateClassWithMultiSuper,
	75: RegOpDelAttribute,
	76: RegOpSetClassField,
	77: RegOpSetMetaclass,
	78: RegOpCallMetaclassInit,
	79: RegOpFormatString,
	80: RegOpListUnpack,
	81: RegOpDictUnpack,
	82: RegOpExceptStarHandler,
	83: RegOpLShift,
	84: RegOpRShift,
	85: RegOpBoolAnd,
	86: RegOpBoolOr,
	87: RegOpContains,
	88: RegOpNotContains,
}

func convertCompilerRegInstructions(instrs []compiler.RegInstruction) []RegInstruction {
	result := make([]RegInstruction, len(instrs))
	for i, instr := range instrs {
		vmOp, ok := compilerToVMOpcode[byte(instr.Opcode)]
		if !ok {
			// Fallback: try direct mapping for unknown opcodes
			vmOp = RegOpcode(instr.Opcode)
		}
		result[i] = RegInstruction{
			Opcode:   vmOp,
			Operands: instr.Operands,
		}
	}
	return result
}

// RunRegDirect executes register bytecode directly without the translate() step.
// It uses the RegBytecode produced by the RegisterCompiler.
func (rvm *RegisterVM) RunRegDirect(regBytecode *compiler.RegBytecode) error {
	vm := rvm.vm
	vm.constants = regBytecode.Constants

	// Ensure registers are large enough
	needed := regBytecode.NumRegs + 64
	if needed > len(rvm.registers) {
		newRegs := make([]objects.Object, needed*2)
		copy(newRegs, rvm.registers)
		rvm.registers = newRegs
	}

	// Convert instructions
	instructions := convertCompilerRegInstructions(regBytecode.Instructions)

	// Set up main frame
	frame := &RegFrame{
		instructions: instructions,
		ip:           -1,
		basePointer:  0,
		regBase:      0,
	}
	rvm.frames[0] = frame
	rvm.frameIndex = 1

	// Execute the instructions
	ip := 0
	for ip < len(instructions) {
		inst := instructions[ip]
		frame.ip = ip

		switch inst.Opcode {
		case RegOpLoadConst:
			dst := inst.Operands[0]
			constIdx := inst.Operands[1]
			rvm.regSet(dst, vm.constants[constIdx])

		case RegOpMove, RegOpDupTop:
			dst := inst.Operands[0]
			src := inst.Operands[1]
			rvm.regSet(dst, rvm.regGet(src))

		case RegOpPop:
			// No-op in register VM

		case RegOpAdd:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int+int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value+rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.binaryOp(compiler.OpAdd, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSub:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int-int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value-rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.binaryOp(compiler.OpSub, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpMul:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int*int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value*rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.binaryOp(compiler.OpMul, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpDiv:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int/int path (returns Float)
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					if rightInt.Value == 0 {
						return fmt.Errorf("ZeroDivisionError: division by zero")
					}
					rvm.regSet(dst, &objects.Float{Value: float64(leftInt.Value) / float64(rightInt.Value)})
					ip++
					continue
				}
			}
			result, err := rvm.binaryOp(compiler.OpDiv, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpMod:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int%int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					if rightInt.Value == 0 {
						return fmt.Errorf("ZeroDivisionError: modulo by zero")
					}
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value%rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.binaryOp(compiler.OpMod, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpFloorDiv:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int//int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					if rightInt.Value == 0 {
						return fmt.Errorf("ZeroDivisionError: floor division by zero")
					}
					result := leftInt.Value / rightInt.Value
					if (leftInt.Value < 0) != (rightInt.Value < 0) && leftInt.Value%rightInt.Value != 0 {
						result -= 1
					}
					rvm.regSet(dst, objects.GetCachedInteger(result))
					ip++
					continue
				}
			}
			result, err := rvm.binaryOp(compiler.OpFloorDiv, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpPower:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int**int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					if rightInt.Value < 0 {
						rvm.regSet(dst, &objects.Float{Value: math.Pow(float64(leftInt.Value), float64(rightInt.Value))})
						ip++
						continue
					}
					result := int64(1)
					for i := int64(0); i < rightInt.Value; i++ {
						result *= leftInt.Value
					}
					rvm.regSet(dst, objects.GetCachedInteger(result))
					ip++
					continue
				}
			}
			result, err := rvm.binaryOp(compiler.OpPower, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpCompare:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			cmpOp := compiler.Opcode(inst.Operands[3])
			result, err := rvm.compareOp(cmpOp, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpJump:
			target := inst.Operands[0]
			ip = target
			continue

		case RegOpJumpIfFalse:
			condReg := inst.Operands[0]
			target := inst.Operands[1]
			condition := rvm.regGet(condReg)
			if !isTruthy(condition) {
				ip = target
				continue
			}

		case RegOpNegate:
			dst := inst.Operands[0]
			src := inst.Operands[1]
			result, err := rvm.negateOp(rvm.regGet(src))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpNot:
			dst := inst.Operands[0]
			src := inst.Operands[1]
			result := rvm.notOp(rvm.regGet(src))
			rvm.regSet(dst, result)

		case RegOpTrue:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.True)

		case RegOpFalse:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.False)

		case RegOpNull:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.None_)

		case RegOpEllipsis:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.EllipsisSingleton)

		case RegOpSetGlobal:
			globalIndex := inst.Operands[0]
			valReg := inst.Operands[1]
			val := rvm.regGet(valReg)
			vm.globals[globalIndex] = val
			vm.globalVersions[globalIndex]++
			vm.globalCache[globalIndex] = GlobalCacheEntry{Value: val, Version: vm.globalVersions[globalIndex]}

		case RegOpGetGlobal:
			dst := inst.Operands[0]
			globalIndex := inst.Operands[1]
			cached := vm.globalCache[globalIndex]
			var val objects.Object
			if cached.Version == vm.globalVersions[globalIndex] && cached.Value != nil {
				val = cached.Value
			} else {
				val = vm.globals[globalIndex]
				vm.globalCache[globalIndex] = GlobalCacheEntry{Value: val, Version: vm.globalVersions[globalIndex]}
			}
			rvm.regSet(dst, val)

		case RegOpSetLocal:
			localIndex := inst.Operands[0]
			valReg := inst.Operands[1]
			vm.stack[frame.basePointer+localIndex] = rvm.regGet(valReg)

		case RegOpGetLocal:
			dst := inst.Operands[0]
			localIndex := inst.Operands[1]
			rvm.regSet(dst, vm.stack[frame.basePointer+localIndex])

		case RegOpGetFree:
			dst := inst.Operands[0]
			freeIndex := inst.Operands[1]
			if frame.freeVars != nil && freeIndex < len(frame.freeVars) {
				rvm.regSet(dst, frame.freeVars[freeIndex])
			} else {
				rvm.regSet(dst, objects.None_)
			}

		case RegOpCall:
			dst := inst.Operands[0]
			funcReg := inst.Operands[1]
			numArgs := inst.Operands[2]

			callee := rvm.regGet(funcReg)

			// Handle Builtin functions directly
			if builtin, ok := callee.(*objects.Builtin); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}
				result := builtin.Fn(args...)
				rvm.regSet(dst, result)
				break
			}

			// Handle Callable interface
			if callable, ok := callee.(objects.Callable); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}
				result := callable.Call(args...)
				rvm.regSet(dst, result)
				break
			}

			// Handle Closure: execute the function body directly
			if closure, ok := callee.(*objects.Closure); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}

				// Build the function's register bytecode from the closure instructions
				fnInstructions := closure.Instructions
				if len(fnInstructions) == 0 {
					rvm.regSet(dst, objects.None_)
					break
				}

				// Convert bytes back to RegInstructions
				regInstrs := bytesToRegInstructions(fnInstructions)

				// Save current state
				oldFrameIndex := rvm.frameIndex

				// Calculate new regBase for the callee
				newRegBase := rvm.numRegs
				neededRegs := newRegBase + closure.NumLocals + 64
				if neededRegs > len(rvm.registers) {
					newRegs := make([]objects.Object, neededRegs*2)
					copy(newRegs, rvm.registers)
					rvm.registers = newRegs
				}

				convertedInstrs := convertCompilerRegInstructions(regInstrs)
				newFrame := &RegFrame{
					instructions: convertedInstrs,
					ip:           -1,
					basePointer:  0,
					regBase:      newRegBase,
					freeVars:     closure.Free,
				}
				rvm.frames[rvm.frameIndex] = newFrame
				rvm.frameIndex++

				// Set up argument registers in the callee's frame
				for i := 0; i < numArgs && i < closure.NumParameters; i++ {
					rvm.regSet(i, args[i])
				}

				err := rvm.executeRegFrame(newFrame)
				rvm.frameIndex = oldFrameIndex

				if err != nil {
					return err
				}

				result := vm.lastPopped
				rvm.regSet(dst, result)
				break
			}

			// Handle CompiledFunction: execute the function body directly
			if compiledFn, ok := callee.(*compiler.CompiledFunction); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}

				// Use RegInstructions if available, otherwise decode from bytes
				var regInstrs []compiler.RegInstruction
				if len(compiledFn.RegInstructions) > 0 {
					regInstrs = compiledFn.RegInstructions
				} else {
					regInstrs = bytesToRegInstructions(compiledFn.Instructions)
				}

				if len(regInstrs) == 0 {
					rvm.regSet(dst, objects.None_)
					break
				}

				// Save current state
				oldSP := vm.sp
				oldFrameIndex := rvm.frameIndex

				// Set up locals on the stack (for RegOpGetLocal/RegOpSetLocal)
				for i := 0; i < compiledFn.NumLocals; i++ {
					if i < numArgs && i < compiledFn.NumParameters {
						vm.push(args[i])
					} else {
						vm.push(objects.None_)
					}
				}

				// Calculate new regBase for the callee
				newRegBase := rvm.numRegs
				neededRegs := newRegBase + compiledFn.NumLocals + 64
				if neededRegs > len(rvm.registers) {
					newRegs := make([]objects.Object, neededRegs*2)
					copy(newRegs, rvm.registers)
					rvm.registers = newRegs
				}

				// Set up a new frame with the correct basePointer and regBase
				convertedInstrs := convertCompilerRegInstructions(regInstrs)
				newFrame := &RegFrame{
					instructions: convertedInstrs,
					ip:           -1,
					basePointer:  oldSP,
					regBase:      newRegBase,
				}
				rvm.frames[rvm.frameIndex] = newFrame
				rvm.frameIndex++

				// Set up argument registers in the callee's frame
				for i := 0; i < numArgs && i < compiledFn.NumParameters; i++ {
					rvm.regSet(i, args[i])
				}

				// Execute the function body directly using the instruction loop
				err := rvm.executeRegFrame(newFrame)

				// Restore state
				rvm.frameIndex = oldFrameIndex
				vm.sp = oldSP

				if err != nil {
					return err
				}

				result := vm.lastPopped
				rvm.regSet(dst, result)
				break
			}

			// Handle Class instantiation
			if classObj, ok := callee.(*objects.Class); ok {
				instance := &objects.Instance{
					Class: classObj,
				}
				if classObj.HasSlots() {
					allSlots := classObj.AllSlotNames()
					instance.SlotValues = make([]objects.Object, len(allSlots))
				} else {
					instance.Fields = make(map[string]objects.Object)
				}

				if initMethod, ok := classObj.Methods["__init__"]; ok {
					if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
						// Execute __init__
						_ = fn // simplified: skip init for now
					}
				}
				rvm.regSet(dst, instance)
				break
			}

			// Fallback: use the stack VM's executeCall
			vm.push(callee)
			for i := 0; i < numArgs; i++ {
				vm.push(rvm.regGet(inst.Operands[3+i]))
			}

			numArgs += vm.unpackExtraArgs
			vm.unpackExtraArgs = 0

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}

			result := vm.pop()
			rvm.regSet(dst, result)

		case RegOpReturn:
			valReg := inst.Operands[0]
			returnValue := rvm.regGet(valReg)
			vm.lastPopped = returnValue
			return nil

		case RegOpReturnNone:
			vm.lastPopped = objects.None_
			return nil

		case RegOpClosure:
			dst := inst.Operands[0]
			constIndex := inst.Operands[1]
			numFree := inst.Operands[2]

			fn, ok := vm.constants[constIndex].(*compiler.CompiledFunction)
			if !ok {
				return fmt.Errorf("not a function: %T", vm.constants[constIndex])
			}

			free := make([]objects.Object, numFree)
			for i := 0; i < numFree; i++ {
				free[i] = rvm.regGet(inst.Operands[3+i])
			}

			closure := &objects.Closure{
				Instructions:          fn.Instructions,
				NumLocals:             fn.NumLocals,
				NumParameters:         fn.NumParameters,
				NumKeywordOnly:        fn.NumKeywordOnly,
				NumPositionalOnly:     fn.NumPositionalOnly,
				NumDefaults:           fn.NumDefaults,
				NumPositionalDefaults: fn.NumPositionalDefaults,
				ParameterNames:        fn.ParameterNames,
				PositionalOnly:        fn.PositionalOnly,
				IsGenerator:           fn.IsGenerator,
				Free:                  free,
				VarArgs:               fn.VarArgs,
				KwArgs:                fn.KwArgs,
			}
			rvm.regSet(dst, closure)

		case RegOpBuildList:
			dst := inst.Operands[0]
			numElements := inst.Operands[1]
			elements := make([]objects.Object, numElements)
			for i := 0; i < numElements; i++ {
				elements[i] = rvm.regGet(inst.Operands[2+i])
			}
			rvm.regSet(dst, objects.NewList(elements))

		case RegOpBuildDict:
			dst := inst.Operands[0]
			numElements := inst.Operands[1]
			dict := objects.NewDict()
			for i := 0; i < numElements; i += 2 {
				key := rvm.regGet(inst.Operands[2+i])
				val := rvm.regGet(inst.Operands[2+i+1])
				if err := objects.CheckHashable(key); err != nil {
					return err
				}
				dict.Set(key, val)
			}
			rvm.regSet(dst, dict)

		case RegOpBuildSet:
			dst := inst.Operands[0]
			numElements := inst.Operands[1]
			set := objects.NewSet()
			for i := 0; i < numElements; i++ {
				elem := rvm.regGet(inst.Operands[2+i])
				if err := objects.CheckHashable(elem); err != nil {
					return err
				}
				set.Add(elem)
			}
			rvm.regSet(dst, set)

		case RegOpIndex:
			dst := inst.Operands[0]
			leftReg := inst.Operands[1]
			indexReg := inst.Operands[2]
			left := rvm.regGet(leftReg)
			index := rvm.regGet(indexReg)
			result, err := rvm.indexOp(left, index)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSlice:
			dst := inst.Operands[0]
			leftReg := inst.Operands[1]
			startReg := inst.Operands[2]
			endReg := inst.Operands[3]
			stepReg := inst.Operands[4]
			left := rvm.regGet(leftReg)
			start := rvm.regGet(startReg)
			end := rvm.regGet(endReg)
			step := rvm.regGet(stepReg)
			result, err := rvm.sliceOpWithStep(left, start, end, step)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpGetAttr:
			dst := inst.Operands[0]
			objReg := inst.Operands[1]
			attrIdx := inst.Operands[2]
			obj := rvm.regGet(objReg)
			attrName := vm.constants[attrIdx].(*objects.String).Value
			result, err := rvm.getAttrOp(obj, attrName)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSetAttr:
			objReg := inst.Operands[0]
			valReg := inst.Operands[1]
			attrIdx := inst.Operands[2]
			obj := rvm.regGet(objReg)
			val := rvm.regGet(valReg)
			attrName := vm.constants[attrIdx].(*objects.String).Value
			err := rvm.setAttrOp(obj, val, attrName)
			if err != nil {
				return err
			}

		case RegOpDelAttribute:
			objReg := inst.Operands[0]
			attrIdx := inst.Operands[1]
			obj := rvm.regGet(objReg)
			attrName := vm.constants[attrIdx].(*objects.String).Value
			if instance, ok := obj.(*objects.Instance); ok {
				if instance.IsSlottedInstance() {
					idx := instance.GetSlotIndex(attrName)
					if idx >= 0 && idx < len(instance.SlotValues) {
						instance.SlotValues[idx] = nil
					}
				} else {
					delete(instance.Fields, attrName)
				}
			}

		case RegOpCreateClass:
			dst := inst.Operands[0]
			idx := inst.Operands[1]
			class := vm.constants[idx].(*objects.Class)
			rvm.regSet(dst, class)

		case RegOpCreateClassWithSuper:
			dst := inst.Operands[0]
			idx := inst.Operands[1]
			superReg := inst.Operands[2]
			class := vm.constants[idx].(*objects.Class)
			superClass := rvm.regGet(superReg)
			if superClass != nil {
				if superCls, ok := superClass.(*objects.Class); ok {
					class.SuperClass = superCls
					class.SuperClasses = []*objects.Class{superCls}
					for name, method := range superCls.Methods {
						if _, exists := class.Methods[name]; !exists {
							class.Methods[name] = method
						}
					}
				}
			}
			rvm.regSet(dst, class)

		case RegOpFormatString:
			dst := inst.Operands[0]
			partsCount := inst.Operands[1]
			var result string
			for i := partsCount - 1; i >= 0; i-- {
				part := rvm.regGet(inst.Operands[2+i])
				var partStr string
				if strObj, ok := part.(*objects.String); ok {
					partStr = strObj.Value
				} else {
					partStr = part.Inspect()
				}
				result = partStr + result
			}
			rvm.regSet(dst, &objects.String{Value: result})

		case RegOpMakeGenerator:
			dst := inst.Operands[0]
			fnReg := inst.Operands[1]
			fnObj := rvm.regGet(fnReg)
			if cf, ok := fnObj.(*compiler.CompiledFunction); ok {
				gen := &objects.Generator{
					Instructions: cf.Instructions,
					Constants:    vm.constants,
					Locals:       make([]objects.Object, cf.NumLocals),
					IP:           -1,
					Stack:        make([]objects.Object, StackSize),
					StackPtr:     0,
					BasePointer:  0,
					Done:         false,
				}
				rvm.regSet(dst, gen)
			}

		case RegOpMakeAsync:
			dst := inst.Operands[0]
			fnReg := inst.Operands[1]
			fnObj := rvm.regGet(fnReg)
			if cf, ok := fnObj.(*compiler.CompiledFunction); ok {
				async := &objects.Async{
					Instructions: cf.Instructions,
					Constants:    vm.constants,
					Locals:       make([]objects.Object, cf.NumLocals),
					IP:           -1,
					Stack:        make([]objects.Object, StackSize),
					StackPtr:     0,
					BasePointer:  0,
					Done:         false,
				}
				rvm.regSet(dst, async)
			}

		case RegOpAwait:
			dst := inst.Operands[0]
			valReg := inst.Operands[1]
			val := rvm.regGet(valReg)
			if asyncObj, ok := val.(*objects.Async); ok {
				if asyncObj.Done {
					rvm.regSet(dst, asyncObj.Result)
				} else {
					// R2.8: Not done - execute the async coroutine synchronously
					result := objects.CallFunction(asyncObj)
					asyncObj.Result = result
					asyncObj.Done = true
					rvm.regSet(dst, result)
				}
			} else {
				rvm.regSet(dst, val)
			}

		case RegOpYieldValue:
			valReg := inst.Operands[0]
			yieldValue := rvm.regGet(valReg)
			vm.lastPopped = yieldValue
			return nil

		case RegOpRaise:
			errReg := inst.Operands[0]
			errObj := rvm.regGet(errReg)
			caught := vm.raiseException(errObj)
			if !caught {
				vm.pendingError = errObj
				return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
			}
			return nil

		case RegOpListUnpack:
			listReg := inst.Operands[0]
			listObj := rvm.regGet(listReg)
			if list, ok := listObj.(*objects.List); ok {
				for _, elem := range list.Elements {
					reg := rvm.numRegs
					rvm.regSet(reg, elem)
				}
				vm.unpackExtraArgs += len(list.Elements) - 1
			}

		case RegOpDictUnpack:
			// No additional action needed

		// Bit operations
		case RegOpBitOr:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int|int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value|rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitOr, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpBitAnd:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int&int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value&rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitAnd, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpBitXor:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int^int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value^rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitXor, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpLShift:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value<<rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitOr, left, right) // fallback
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpRShift:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value>>rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitOr, left, right) // fallback
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpContains:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			result := rvm.regContains(left, right)
			rvm.regSet(dst, result)

		case RegOpNotContains:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			result := rvm.regContains(left, right)
			if b, ok := result.(*objects.Boolean); ok {
				if b.Value {
					rvm.regSet(dst, &objects.Boolean{Value: false})
				} else {
					rvm.regSet(dst, &objects.Boolean{Value: true})
				}
			} else {
				rvm.regSet(dst, result)
			}

		// In-place operations
		case RegOpInPlaceAdd, RegOpInPlaceSub, RegOpInPlaceMul,
			RegOpInPlaceDiv, RegOpInPlaceMod, RegOpInPlaceFloorDiv,
			RegOpInPlacePower, RegOpInPlaceBitOr, RegOpInPlaceBitAnd,
			RegOpInPlaceBitXor, RegOpInPlaceLShift, RegOpInPlaceRShift:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)

			var compOp compiler.Opcode
			switch inst.Opcode {
			case RegOpInPlaceAdd:
				compOp = compiler.OpInPlaceAdd
			case RegOpInPlaceSub:
				compOp = compiler.OpInPlaceSub
			case RegOpInPlaceMul:
				compOp = compiler.OpInPlaceMul
			case RegOpInPlaceDiv:
				compOp = compiler.OpInPlaceDiv
			case RegOpInPlaceMod:
				compOp = compiler.OpInPlaceMod
			case RegOpInPlaceFloorDiv:
				compOp = compiler.OpInPlaceFloorDiv
			case RegOpInPlacePower:
				compOp = compiler.OpInPlacePower
			case RegOpInPlaceBitOr:
				compOp = compiler.OpInPlaceBitOr
			case RegOpInPlaceBitAnd:
				compOp = compiler.OpInPlaceBitAnd
			case RegOpInPlaceBitXor:
				compOp = compiler.OpInPlaceBitXor
			case RegOpInPlaceLShift:
				compOp = compiler.OpInPlaceLShift
			case RegOpInPlaceRShift:
				compOp = compiler.OpInPlaceRShift
			}

			// Try __ixxx__ method
			info, ok := inPlaceAttrMap[compOp]
			if ok {
				if getter, ok := left.(objects.AttributeGetter); ok {
					if method, found := getter.GetAttr(info.attrName); found {
						if builtin, ok := method.(*objects.Builtin); ok {
							result := builtin.Fn(right)
							if result.Type() != objects.ERROR_OBJ {
								rvm.regSet(dst, result)
								break
							}
						}
					}
				}
			}
			// Fallback to normal binary op
			result, err := rvm.binaryOp(info.fallbackOp, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		// Index/Slice assignment
		case RegOpSetIndex:
			objReg := inst.Operands[0]
			indexReg := inst.Operands[1]
			valueReg := inst.Operands[2]
			err := rvm.regExecuteSetIndex(rvm.regGet(objReg), rvm.regGet(indexReg), rvm.regGet(valueReg))
			if err != nil {
				return err
			}

		case RegOpSetSlice:
			objReg := inst.Operands[0]
			startReg := inst.Operands[1]
			endReg := inst.Operands[2]
			stepReg := inst.Operands[3]
			valueReg := inst.Operands[4]
			err := rvm.regExecuteSetSlice(rvm.regGet(objReg), rvm.regGet(startReg), rvm.regGet(endReg), rvm.regGet(stepReg), rvm.regGet(valueReg))
			if err != nil {
				return err
			}

		case RegOpEnterContext:
			dst := inst.Operands[0]
			ctxReg := inst.Operands[1]
			ctxManager := rvm.regGet(ctxReg)
			if cm, ok := ctxManager.(*objects.ContextManager); ok {
				if cm.EnterFunc != nil {
					rvm.regSet(dst, cm.EnterFunc())
				} else {
					rvm.regSet(dst, objects.None_)
				}
			} else {
				rvm.regSet(dst, objects.None_)
			}

		case RegOpExitContext:
			dst := inst.Operands[0]
			ctxReg := inst.Operands[1]
			excReg := inst.Operands[2]
			ctxManager := rvm.regGet(ctxReg)
			exc := rvm.regGet(excReg)
			if cm, ok := ctxManager.(*objects.ContextManager); ok {
				if cm.ExitFunc != nil {
					rvm.regSet(dst, cm.ExitFunc(exc))
				} else {
					rvm.regSet(dst, objects.None_)
				}
			} else {
				rvm.regSet(dst, objects.None_)
			}

		case RegOpBeginTry:
			exceptCount := inst.Operands[0]
			hasFinally := inst.Operands[1]
			handlerIP := inst.Operands[2]
			finallyStartIP := inst.Operands[3]
			tryBlockStartIP := ip + 1

			handler := ExceptionHandler{
				handlerIP:       handlerIP,
				stackPtr:        vm.sp,
				exceptionType:   "",
				varName:         "",
				handlerStartIP:  -1,
				tryBlockStartIP: tryBlockStartIP,
				hasFinally:      hasFinally == 1,
				finallyStartIP:  finallyStartIP,
				finallyEndIP:    -1,
				pendingError:    nil,
				frameIndex:      vm.framesIndex,
			}
			handler.exceptCount = exceptCount
			handler.baseIP = frame.ip + 1
			vm.exceptionStack = append(vm.exceptionStack, handler)

		case RegOpEndTry:
			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				handler := vm.exceptionStack[lastIdx]
				if vm.sp > handler.stackPtr {
					_ = vm.pop()
				}
				pendingError := handler.pendingError
				vm.exceptionStack = vm.exceptionStack[:lastIdx]
				vm.inFinally = false
				if pendingError != nil {
					caught := vm.raiseException(pendingError)
					if !caught {
						vm.pendingError = pendingError
						return fmt.Errorf("unhandled exception: %s", pendingError.Inspect())
					}
				}
			}

		case RegOpExceptHandler:
			typeIdx := inst.Operands[0]
			varIdx := inst.Operands[1]

			var exceptionType, varName string
			if typeIdx > 0 && typeIdx < len(vm.constants) {
				if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
					exceptionType = typeObj.Value
				}
			}
			if varIdx > 0 && varIdx < len(vm.constants) {
				if varObj, ok := vm.constants[varIdx].(*objects.String); ok {
					varName = varObj.Value
				}
			}

			handlerStartIP := ip + 1

			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				for lastIdx >= 0 && vm.exceptionStack[lastIdx].handlerIP != -1 {
					lastIdx--
				}
				if lastIdx >= 0 {
					existingHandler := vm.exceptionStack[lastIdx]
					vm.exceptionStack[lastIdx] = ExceptionHandler{
						handlerIP:       frame.ip,
						stackPtr:        vm.sp - 1,
						exceptionType:   exceptionType,
						varName:         varName,
						handlerStartIP:  handlerStartIP,
						tryBlockStartIP: existingHandler.tryBlockStartIP,
						hasFinally:      existingHandler.hasFinally,
						finallyStartIP:  existingHandler.finallyStartIP,
						finallyEndIP:    existingHandler.finallyEndIP,
						pendingError:    existingHandler.pendingError,
						exceptOpcodeIP:  frame.ip,
					}
				}
			}

		case RegOpFinally:
			finallyEndIP := inst.Operands[0]
			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				if lastIdx >= 0 {
					vm.exceptionStack[lastIdx].finallyStartIP = ip + 1
					vm.exceptionStack[lastIdx].finallyEndIP = finallyEndIP
				}
			}
			vm.inFinally = true
			pendingError := vm.pendingError
			vm.pendingError = nil
			if pendingError != nil {
				if len(vm.exceptionStack) > 0 {
					lastIdx := len(vm.exceptionStack) - 1
					if lastIdx >= 0 {
						vm.exceptionStack[lastIdx].pendingError = pendingError
					}
				}
			}

		default:
			// Unknown register opcode
			return fmt.Errorf("unknown register opcode in RunRegDirect: %d", inst.Opcode)
		}

		ip++
	}

	return nil
}

// LastPopped returns the last popped value from the underlying VM.
func (rvm *RegisterVM) LastPopped() objects.Object {
	return rvm.vm.lastPopped
}

// bytesToRegInstructions decodes register bytecode bytes back into RegInstruction slice.
// This is the inverse of regInstructionsToBytes in the compiler package.
// Format: [opcode, operand1_hi, operand1_lo, operand2_hi, operand2_lo, ...]
// Each operand is encoded as big-endian uint16.
// We use the compilerToVMOpcode map to translate compiler opcodes to VM opcodes.
func bytesToRegInstructions(data []byte) []compiler.RegInstruction {
	var instrs []compiler.RegInstruction
	ip := 0
	for ip < len(data) {
		opcode := compiler.RegOpcode(data[ip])
		ip++

		// Determine number of operands based on opcode
		numOperands := regOpcodeNumOperands(opcode)
		operands := make([]int, numOperands)
		for i := 0; i < numOperands && ip+1 < len(data); i++ {
			operands[i] = int(data[ip])<<8 | int(data[ip+1])
			ip += 2
		}

		instrs = append(instrs, compiler.RegInstruction{
			Opcode:   opcode,
			Operands: operands,
		})
	}
	return instrs
}

// regOpcodeNumOperands returns the expected number of operands for each register opcode.
func regOpcodeNumOperands(op compiler.RegOpcode) int {
	switch op {
	case compiler.RegOpcode(0): // ROpLoadConst
		return 2
	case compiler.RegOpcode(1): // ROpMove
		return 2
	case compiler.RegOpcode(2), // ROpAdd
		compiler.RegOpcode(3), // ROpSub
		compiler.RegOpcode(4), // ROpMul
		compiler.RegOpcode(5), // ROpDiv
		compiler.RegOpcode(6), // ROpMod
		compiler.RegOpcode(7), // ROpFloorDiv
		compiler.RegOpcode(8): // ROpPower
		return 3
	case compiler.RegOpcode(9): // ROpCompare
		return 4
	case compiler.RegOpcode(10): // ROpJump
		return 1
	case compiler.RegOpcode(11): // ROpJumpIfFalse
		return 2
	case compiler.RegOpcode(13): // ROpReturn
		return 1
	case compiler.RegOpcode(14), // ROpReturnNone
		compiler.RegOpcode(28), // ROpNull
		compiler.RegOpcode(29), // ROpTrue
		compiler.RegOpcode(30): // ROpFalse
		return 1
	case compiler.RegOpcode(15): // ROpGetAttr
		return 3
	case compiler.RegOpcode(16): // ROpSetAttr
		return 3
	case compiler.RegOpcode(17): // ROpGetGlobal
		return 2
	case compiler.RegOpcode(18): // ROpSetGlobal
		return 2
	case compiler.RegOpcode(19): // ROpGetLocal
		return 2
	case compiler.RegOpcode(20): // ROpSetLocal
		return 2
	case compiler.RegOpcode(26): // ROpNegate
		return 2
	case compiler.RegOpcode(27): // ROpNot
		return 2
	case compiler.RegOpcode(58): // ROpEllipsis
		return 1
	case compiler.RegOpcode(59): // ROpPop
		return 0
	case compiler.RegOpcode(60): // ROpDupTop
		return 2
	case compiler.RegOpcode(61): // ROpRaise
		return 1
	default:
		// For unknown opcodes, try to read operands until we can't
		return 0
	}
}

// executeRegFrame executes register instructions from a given frame.
// This is used for function calls within RunRegDirect.
func (rvm *RegisterVM) executeRegFrame(frame *RegFrame) error {
	vm := rvm.vm
	instructions := frame.instructions

	ip := 0
	for ip < len(instructions) {
		inst := instructions[ip]
		frame.ip = ip

		switch inst.Opcode {
		case RegOpLoadConst:
			dst := inst.Operands[0]
			constIdx := inst.Operands[1]
			rvm.regSet(dst, vm.constants[constIdx])

		case RegOpMove, RegOpDupTop:
			dst := inst.Operands[0]
			src := inst.Operands[1]
			rvm.regSet(dst, rvm.regGet(src))

		case RegOpPop:
			// No-op in register VM

		case RegOpAdd:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpAdd, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSub:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpSub, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpMul:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpMul, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpDiv:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpDiv, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpMod:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpMod, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpFloorDiv:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpFloorDiv, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpPower:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpPower, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpCompare:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			cmpOp := compiler.Opcode(inst.Operands[3])
			result, err := rvm.compareOp(cmpOp, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpJump:
			target := inst.Operands[0]
			ip = target
			continue

		case RegOpJumpIfFalse:
			condReg := inst.Operands[0]
			target := inst.Operands[1]
			condition := rvm.regGet(condReg)
			if !isTruthy(condition) {
				ip = target
				continue
			}

		case RegOpNegate:
			dst := inst.Operands[0]
			src := inst.Operands[1]
			result, err := rvm.negateOp(rvm.regGet(src))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpNot:
			dst := inst.Operands[0]
			src := inst.Operands[1]
			result := rvm.notOp(rvm.regGet(src))
			rvm.regSet(dst, result)

		case RegOpNull:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.None_)

		case RegOpTrue:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.True)

		case RegOpFalse:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.False)

		case RegOpEllipsis:
			dst := inst.Operands[0]
			rvm.regSet(dst, objects.EllipsisSingleton)

		case RegOpSetGlobal:
			globalIndex := inst.Operands[0]
			valReg := inst.Operands[1]
			val := rvm.regGet(valReg)
			vm.globals[globalIndex] = val
			vm.globalVersions[globalIndex]++
			vm.globalCache[globalIndex] = GlobalCacheEntry{Value: val, Version: vm.globalVersions[globalIndex]}

		case RegOpGetGlobal:
			dst := inst.Operands[0]
			globalIndex := inst.Operands[1]
			cached := vm.globalCache[globalIndex]
			var val objects.Object
			if cached.Version == vm.globalVersions[globalIndex] && cached.Value != nil {
				val = cached.Value
			} else {
				val = vm.globals[globalIndex]
				vm.globalCache[globalIndex] = GlobalCacheEntry{Value: val, Version: vm.globalVersions[globalIndex]}
			}
			rvm.regSet(dst, val)

		case RegOpSetLocal:
			localIndex := inst.Operands[0]
			valReg := inst.Operands[1]
			vm.stack[frame.basePointer+localIndex] = rvm.regGet(valReg)

		case RegOpGetLocal:
			dst := inst.Operands[0]
			localIndex := inst.Operands[1]
			rvm.regSet(dst, vm.stack[frame.basePointer+localIndex])

		case RegOpGetFree:
			dst := inst.Operands[0]
			freeIndex := inst.Operands[1]
			if frame.freeVars != nil && freeIndex < len(frame.freeVars) {
				rvm.regSet(dst, frame.freeVars[freeIndex])
			} else {
				rvm.regSet(dst, objects.None_)
			}

		case RegOpCall:
			dst := inst.Operands[0]
			funcReg := inst.Operands[1]
			numArgs := inst.Operands[2]

			callee := rvm.regGet(funcReg)

			// Handle Builtin functions directly
			if builtin, ok := callee.(*objects.Builtin); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}
				result := builtin.Fn(args...)
				rvm.regSet(dst, result)
				break
			}

			// Handle Callable interface
			if callable, ok := callee.(objects.Callable); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}
				result := callable.Call(args...)
				rvm.regSet(dst, result)
				break
			}

			// Handle Closure: execute the function body directly
			if closure, ok := callee.(*objects.Closure); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}

				// Build the function's register bytecode from the closure instructions
				fnInstructions := closure.Instructions
				if len(fnInstructions) == 0 {
					rvm.regSet(dst, objects.None_)
					break
				}

				// Convert bytes back to RegInstructions
				regInstrs := bytesToRegInstructions(fnInstructions)

				// Save current state
				oldFrameIndex := rvm.frameIndex

				// Calculate new regBase for the callee
				newRegBase := rvm.numRegs
				neededRegs := newRegBase + closure.NumLocals + 64
				if neededRegs > len(rvm.registers) {
					newRegs := make([]objects.Object, neededRegs*2)
					copy(newRegs, rvm.registers)
					rvm.registers = newRegs
				}

				convertedInstrs := convertCompilerRegInstructions(regInstrs)
				newFrame := &RegFrame{
					instructions: convertedInstrs,
					ip:           -1,
					basePointer:  0,
					regBase:      newRegBase,
					freeVars:     closure.Free,
				}
				rvm.frames[rvm.frameIndex] = newFrame
				rvm.frameIndex++

				// Set up argument registers in the callee's frame
				for i := 0; i < numArgs && i < closure.NumParameters; i++ {
					rvm.regSet(i, args[i])
				}

				err := rvm.executeRegFrame(newFrame)
				rvm.frameIndex = oldFrameIndex

				if err != nil {
					return err
				}

				result := vm.lastPopped
				rvm.regSet(dst, result)
				break
			}

			// Handle CompiledFunction: execute the function body directly
			if compiledFn, ok := callee.(*compiler.CompiledFunction); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = rvm.regGet(inst.Operands[3+i])
				}

				// Use RegInstructions if available, otherwise decode from bytes
				var regInstrs []compiler.RegInstruction
				if len(compiledFn.RegInstructions) > 0 {
					regInstrs = compiledFn.RegInstructions
				} else {
					regInstrs = bytesToRegInstructions(compiledFn.Instructions)
				}

				if len(regInstrs) == 0 {
					rvm.regSet(dst, objects.None_)
					break
				}

				// Save current state
				oldSP := vm.sp
				oldFrameIndex := rvm.frameIndex

				// Set up locals on the stack (for RegOpGetLocal/RegOpSetLocal)
				for i := 0; i < compiledFn.NumLocals; i++ {
					if i < numArgs && i < compiledFn.NumParameters {
						vm.push(args[i])
					} else {
						vm.push(objects.None_)
					}
				}

				// Calculate new regBase for the callee
				newRegBase := rvm.numRegs
				neededRegs := newRegBase + compiledFn.NumLocals + 64
				if neededRegs > len(rvm.registers) {
					newRegs := make([]objects.Object, neededRegs*2)
					copy(newRegs, rvm.registers)
					rvm.registers = newRegs
				}

				// Set up a new frame with the correct basePointer and regBase
				convertedInstrs := convertCompilerRegInstructions(regInstrs)
				newFrame := &RegFrame{
					instructions: convertedInstrs,
					ip:           -1,
					basePointer:  oldSP,
					regBase:      newRegBase,
				}
				rvm.frames[rvm.frameIndex] = newFrame
				rvm.frameIndex++

				// Set up argument registers in the callee's frame
				for i := 0; i < numArgs && i < compiledFn.NumParameters; i++ {
					rvm.regSet(i, args[i])
				}

				// Execute the function body directly using the instruction loop
				err := rvm.executeRegFrame(newFrame)

				// Restore state
				rvm.frameIndex = oldFrameIndex
				vm.sp = oldSP

				if err != nil {
					return err
				}

				result := vm.lastPopped
				rvm.regSet(dst, result)
				break
			}

			// Handle Class instantiation
			if classObj, ok := callee.(*objects.Class); ok {
				instance := &objects.Instance{
					Class: classObj,
				}
				if classObj.HasSlots() {
					allSlots := classObj.AllSlotNames()
					instance.SlotValues = make([]objects.Object, len(allSlots))
				} else {
					instance.Fields = make(map[string]objects.Object)
				}

				if initMethod, ok := classObj.Methods["__init__"]; ok {
					if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
						// Execute __init__
						_ = fn // simplified: skip init for now
					}
				}
				rvm.regSet(dst, instance)
				break
			}

			// Fallback: use the stack VM's executeCall
			vm.push(callee)
			for i := 0; i < numArgs; i++ {
				vm.push(rvm.regGet(inst.Operands[3+i]))
			}

			numArgs += vm.unpackExtraArgs
			vm.unpackExtraArgs = 0

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}

			result := vm.pop()
			rvm.regSet(dst, result)

		case RegOpReturn:
			valReg := inst.Operands[0]
			returnValue := rvm.regGet(valReg)
			vm.lastPopped = returnValue
			return nil

		case RegOpReturnNone:
			vm.lastPopped = objects.None_
			return nil

		case RegOpClosure:
			dst := inst.Operands[0]
			constIndex := inst.Operands[1]
			numFree := inst.Operands[2]

			fn, ok := vm.constants[constIndex].(*compiler.CompiledFunction)
			if !ok {
				return fmt.Errorf("not a function: %T", vm.constants[constIndex])
			}

			free := make([]objects.Object, numFree)
			for i := 0; i < numFree; i++ {
				free[i] = rvm.regGet(inst.Operands[3+i])
			}

			closure := &objects.Closure{
				Instructions:          fn.Instructions,
				NumLocals:             fn.NumLocals,
				NumParameters:         fn.NumParameters,
				NumKeywordOnly:        fn.NumKeywordOnly,
				NumPositionalOnly:     fn.NumPositionalOnly,
				NumDefaults:           fn.NumDefaults,
				NumPositionalDefaults: fn.NumPositionalDefaults,
				ParameterNames:        fn.ParameterNames,
				PositionalOnly:        fn.PositionalOnly,
				IsGenerator:           fn.IsGenerator,
				Free:                  free,
				VarArgs:               fn.VarArgs,
				KwArgs:                fn.KwArgs,
			}
			rvm.regSet(dst, closure)

		case RegOpBuildList:
			dst := inst.Operands[0]
			numElements := inst.Operands[1]
			elements := make([]objects.Object, numElements)
			for i := 0; i < numElements; i++ {
				elements[i] = rvm.regGet(inst.Operands[2+i])
			}
			rvm.regSet(dst, objects.NewList(elements))

		case RegOpBuildDict:
			dst := inst.Operands[0]
			numElements := inst.Operands[1]
			dict := objects.NewDict()
			for i := 0; i < numElements; i += 2 {
				key := rvm.regGet(inst.Operands[2+i])
				val := rvm.regGet(inst.Operands[2+i+1])
				if err := objects.CheckHashable(key); err != nil {
					return err
				}
				dict.Set(key, val)
			}
			rvm.regSet(dst, dict)

		case RegOpBuildSet:
			dst := inst.Operands[0]
			numElements := inst.Operands[1]
			set := objects.NewSet()
			for i := 0; i < numElements; i++ {
				elem := rvm.regGet(inst.Operands[2+i])
				if err := objects.CheckHashable(elem); err != nil {
					return err
				}
				set.Add(elem)
			}
			rvm.regSet(dst, set)

		case RegOpIndex:
			dst := inst.Operands[0]
			leftReg := inst.Operands[1]
			indexReg := inst.Operands[2]
			left := rvm.regGet(leftReg)
			index := rvm.regGet(indexReg)
			result, err := rvm.indexOp(left, index)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSlice:
			dst := inst.Operands[0]
			leftReg := inst.Operands[1]
			startReg := inst.Operands[2]
			endReg := inst.Operands[3]
			stepReg := inst.Operands[4]
			left := rvm.regGet(leftReg)
			start := rvm.regGet(startReg)
			end := rvm.regGet(endReg)
			step := rvm.regGet(stepReg)
			result, err := rvm.sliceOpWithStep(left, start, end, step)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpGetAttr:
			dst := inst.Operands[0]
			objReg := inst.Operands[1]
			attrIdx := inst.Operands[2]
			obj := rvm.regGet(objReg)
			attrName := vm.constants[attrIdx].(*objects.String).Value
			result, err := rvm.getAttrOp(obj, attrName)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSetAttr:
			objReg := inst.Operands[0]
			valReg := inst.Operands[1]
			attrIdx := inst.Operands[2]
			obj := rvm.regGet(objReg)
			val := rvm.regGet(valReg)
			attrName := vm.constants[attrIdx].(*objects.String).Value
			err := rvm.setAttrOp(obj, val, attrName)
			if err != nil {
				return err
			}

		case RegOpDelAttribute:
			objReg := inst.Operands[0]
			attrIdx := inst.Operands[1]
			obj := rvm.regGet(objReg)
			attrName := vm.constants[attrIdx].(*objects.String).Value
			if instance, ok := obj.(*objects.Instance); ok {
				if instance.IsSlottedInstance() {
					idx := instance.GetSlotIndex(attrName)
					if idx >= 0 && idx < len(instance.SlotValues) {
						instance.SlotValues[idx] = nil
					}
				} else {
					delete(instance.Fields, attrName)
				}
			}

		case RegOpCreateClass:
			dst := inst.Operands[0]
			idx := inst.Operands[1]
			class := vm.constants[idx].(*objects.Class)
			rvm.regSet(dst, class)

		case RegOpCreateClassWithSuper:
			dst := inst.Operands[0]
			idx := inst.Operands[1]
			superReg := inst.Operands[2]
			class := vm.constants[idx].(*objects.Class)
			superClass := rvm.regGet(superReg)
			if superClass != nil {
				if superCls, ok := superClass.(*objects.Class); ok {
					class.SuperClass = superCls
					class.SuperClasses = []*objects.Class{superCls}
					for name, method := range superCls.Methods {
						if _, exists := class.Methods[name]; !exists {
							class.Methods[name] = method
						}
					}
				}
			}
			rvm.regSet(dst, class)

		case RegOpCreateClassWithMultiSuper:
			dst := inst.Operands[0]
			idx := inst.Operands[1]
			numParents := inst.Operands[2]
			class := vm.constants[idx].(*objects.Class)

			parents := make([]*objects.Class, 0, numParents)
			for i := 0; i < numParents; i++ {
				superObj := rvm.regGet(inst.Operands[3+i])
				if superObj != nil {
					if superCls, ok := superObj.(*objects.Class); ok {
						parents = append(parents, superCls)
					}
				}
			}

			for i, j := 0, len(parents)-1; i < j; i, j = i+1, j-1 {
				parents[i], parents[j] = parents[j], parents[i]
			}

			class.SuperClasses = parents
			if len(parents) > 0 {
				class.SuperClass = parents[0]
				for _, parent := range parents {
					for name, method := range parent.Methods {
						if _, exists := class.Methods[name]; !exists {
							class.Methods[name] = method
						}
					}
				}
				class.ComputeMRO()
			}
			rvm.regSet(dst, class)

		case RegOpSetClassField:
			classReg := inst.Operands[0]
			valReg := inst.Operands[1]
			attrIdx := inst.Operands[2]
			classObj := rvm.regGet(classReg)
			val := rvm.regGet(valReg)
			fieldName := vm.constants[attrIdx].(*objects.String).Value

			if classInst, ok := classObj.(*objects.Class); ok {
				delete(classInst.Methods, fieldName)
				classInst.Fields[fieldName] = val
				if fieldName == "__slots__" {
					if listObj, ok := val.(*objects.List); ok {
						slots := make([]string, 0, len(listObj.Elements))
						for _, elem := range listObj.Elements {
							if strElem, ok := elem.(*objects.String); ok {
								slots = append(slots, strElem.Value)
							}
						}
						classInst.Slots = slots
					}
				}
			} else {
				return fmt.Errorf("cannot set class field on non-class: %T", classObj)
			}

		case RegOpSetMetaclass:
			classReg := inst.Operands[0]
			metaclassReg := inst.Operands[1]
			classObj := rvm.regGet(classReg)
			metaclassObj := rvm.regGet(metaclassReg)
			if classInst, ok := classObj.(*objects.Class); ok {
				if metaclass, ok := metaclassObj.(*objects.Class); ok {
					classInst.Metaclass = metaclass
				}
			}

		case RegOpCallMetaclassInit:
			classReg := inst.Operands[0]
			classObj := rvm.regGet(classReg)
			if classInst, ok := classObj.(*objects.Class); ok {
				if classInst.Metaclass != nil {
					if initMethod, ok := classInst.Metaclass.Methods["__init__"]; ok {
						if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
							// metaclass.__init__(self, cls, name, bases)
							_ = fn // simplified: skip metaclass init for now
						}
					}
				}
			}

		case RegOpFormatString:
			dst := inst.Operands[0]
			partsCount := inst.Operands[1]
			var result string
			for i := partsCount - 1; i >= 0; i-- {
				part := rvm.regGet(inst.Operands[2+i])
				var partStr string
				if strObj, ok := part.(*objects.String); ok {
					partStr = strObj.Value
				} else {
					partStr = part.Inspect()
				}
				result = partStr + result
			}
			rvm.regSet(dst, &objects.String{Value: result})

		case RegOpMakeGenerator:
			dst := inst.Operands[0]
			fnReg := inst.Operands[1]
			fnObj := rvm.regGet(fnReg)
			if cf, ok := fnObj.(*compiler.CompiledFunction); ok {
				gen := &objects.Generator{
					Instructions: cf.Instructions,
					Constants:    vm.constants,
					Locals:       make([]objects.Object, cf.NumLocals),
					IP:           -1,
					Stack:        make([]objects.Object, StackSize),
					StackPtr:     0,
					BasePointer:  0,
					Done:         false,
				}
				rvm.regSet(dst, gen)
			}

		case RegOpMakeAsync:
			dst := inst.Operands[0]
			fnReg := inst.Operands[1]
			fnObj := rvm.regGet(fnReg)
			if cf, ok := fnObj.(*compiler.CompiledFunction); ok {
				async := &objects.Async{
					Instructions: cf.Instructions,
					Constants:    vm.constants,
					Locals:       make([]objects.Object, cf.NumLocals),
					IP:           -1,
					Stack:        make([]objects.Object, StackSize),
					StackPtr:     0,
					BasePointer:  0,
					Done:         false,
				}
				rvm.regSet(dst, async)
			}

		case RegOpAwait:
			dst := inst.Operands[0]
			valReg := inst.Operands[1]
			val := rvm.regGet(valReg)
			if asyncObj, ok := val.(*objects.Async); ok {
				if asyncObj.Done {
					rvm.regSet(dst, asyncObj.Result)
				} else {
					// Not done - execute the async coroutine synchronously
					result := objects.CallFunction(asyncObj)
					asyncObj.Result = result
					asyncObj.Done = true
					rvm.regSet(dst, result)
				}
			} else {
				rvm.regSet(dst, val)
			}

		case RegOpYieldValue:
			valReg := inst.Operands[0]
			yieldValue := rvm.regGet(valReg)
			vm.lastPopped = yieldValue
			return nil

		case RegOpRaise:
			errReg := inst.Operands[0]
			errObj := rvm.regGet(errReg)
			caught := vm.raiseException(errObj)
			if !caught {
				vm.pendingError = errObj
				return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
			}
			return nil

		case RegOpListUnpack:
			listReg := inst.Operands[0]
			listObj := rvm.regGet(listReg)
			if list, ok := listObj.(*objects.List); ok {
				for _, elem := range list.Elements {
					reg := rvm.numRegs
					rvm.regSet(reg, elem)
				}
				vm.unpackExtraArgs += len(list.Elements) - 1
			}

		case RegOpDictUnpack:
			// No additional action needed

		// Bit operations
		case RegOpBitOr:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int|int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value|rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitOr, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpBitAnd:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int&int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value&rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitAnd, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpBitXor:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			// Fast int^int path
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value^rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitXor, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpLShift:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value<<rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitOr, left, right) // fallback
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpRShift:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					rvm.regSet(dst, objects.GetCachedInteger(leftInt.Value>>rightInt.Value))
					ip++
					continue
				}
			}
			result, err := rvm.regBinaryOp(compiler.OpBitOr, left, right) // fallback
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpContains:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			result := rvm.regContains(left, right)
			rvm.regSet(dst, result)

		case RegOpNotContains:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)
			result := rvm.regContains(left, right)
			if b, ok := result.(*objects.Boolean); ok {
				if b.Value {
					rvm.regSet(dst, &objects.Boolean{Value: false})
				} else {
					rvm.regSet(dst, &objects.Boolean{Value: true})
				}
			} else {
				rvm.regSet(dst, result)
			}

		// Set operations
		case RegOpSetUnion:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.regBinaryOp(compiler.OpSetUnion, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSetIntersection:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.regBinaryOp(compiler.OpSetIntersection, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSetDifference:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.regBinaryOp(compiler.OpSetDifference, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		case RegOpSetSymmetricDifference:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.regBinaryOp(compiler.OpSetSymmetricDifference, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		// In-place operations
		case RegOpInPlaceAdd, RegOpInPlaceSub, RegOpInPlaceMul,
			RegOpInPlaceDiv, RegOpInPlaceMod, RegOpInPlaceFloorDiv,
			RegOpInPlacePower, RegOpInPlaceBitOr, RegOpInPlaceBitAnd,
			RegOpInPlaceBitXor, RegOpInPlaceLShift, RegOpInPlaceRShift:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			left := rvm.regGet(src1)
			right := rvm.regGet(src2)

			var compOp compiler.Opcode
			switch inst.Opcode {
			case RegOpInPlaceAdd:
				compOp = compiler.OpInPlaceAdd
			case RegOpInPlaceSub:
				compOp = compiler.OpInPlaceSub
			case RegOpInPlaceMul:
				compOp = compiler.OpInPlaceMul
			case RegOpInPlaceDiv:
				compOp = compiler.OpInPlaceDiv
			case RegOpInPlaceMod:
				compOp = compiler.OpInPlaceMod
			case RegOpInPlaceFloorDiv:
				compOp = compiler.OpInPlaceFloorDiv
			case RegOpInPlacePower:
				compOp = compiler.OpInPlacePower
			case RegOpInPlaceBitOr:
				compOp = compiler.OpInPlaceBitOr
			case RegOpInPlaceBitAnd:
				compOp = compiler.OpInPlaceBitAnd
			case RegOpInPlaceBitXor:
				compOp = compiler.OpInPlaceBitXor
			case RegOpInPlaceLShift:
				compOp = compiler.OpInPlaceLShift
			case RegOpInPlaceRShift:
				compOp = compiler.OpInPlaceRShift
			}

			// Try __ixxx__ method
			info, ok := inPlaceAttrMap[compOp]
			if ok {
				if getter, ok := left.(objects.AttributeGetter); ok {
					if method, found := getter.GetAttr(info.attrName); found {
						if builtin, ok := method.(*objects.Builtin); ok {
							result := builtin.Fn(right)
							if result.Type() != objects.ERROR_OBJ {
								rvm.regSet(dst, result)
								break
							}
						}
					}
				}
			}
			// Fallback to normal binary op
			result, err := rvm.binaryOp(info.fallbackOp, left, right)
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)

		// Index/Slice assignment
		case RegOpSetIndex:
			objReg := inst.Operands[0]
			indexReg := inst.Operands[1]
			valueReg := inst.Operands[2]
			err := rvm.regExecuteSetIndex(rvm.regGet(objReg), rvm.regGet(indexReg), rvm.regGet(valueReg))
			if err != nil {
				return err
			}

		case RegOpSetSlice:
			objReg := inst.Operands[0]
			startReg := inst.Operands[1]
			endReg := inst.Operands[2]
			stepReg := inst.Operands[3]
			valueReg := inst.Operands[4]
			err := rvm.regExecuteSetSlice(rvm.regGet(objReg), rvm.regGet(startReg), rvm.regGet(endReg), rvm.regGet(stepReg), rvm.regGet(valueReg))
			if err != nil {
				return err
			}

		case RegOpStringBuilderCreate:
			dst := inst.Operands[0]
			sb := objects.NewStringBuilder()
			rvm.regSet(dst, sb)

		case RegOpStringBuilderAppend:
			builderReg := inst.Operands[0]
			valueReg := inst.Operands[1]
			value := rvm.regGet(valueReg)
			sb := rvm.regGet(builderReg)
			if builder, ok := sb.(*objects.StringBuilder); ok {
				var s string
				if str, ok := value.(*objects.String); ok {
					s = str.Value
				} else {
					s = value.Inspect()
				}
				builder.Builder.WriteString(s)
			}

		case RegOpStringBuilderBuild:
			dst := inst.Operands[0]
			builderReg := inst.Operands[1]
			sb := rvm.regGet(builderReg)
			if builder, ok := sb.(*objects.StringBuilder); ok {
				rvm.regSet(dst, &objects.String{Value: builder.Builder.String()})
			} else {
				rvm.regSet(dst, sb)
			}

		case RegOpArrayPrealloc:
			dst := inst.Operands[0]
			capacity := inst.Operands[1]
			list := &objects.List{Elements: make([]objects.Object, 0, capacity)}
			rvm.regSet(dst, list)

		case RegOpEnterContext:
			dst := inst.Operands[0]
			ctxReg := inst.Operands[1]
			ctxManager := rvm.regGet(ctxReg)
			if cm, ok := ctxManager.(*objects.ContextManager); ok {
				if cm.EnterFunc != nil {
					rvm.regSet(dst, cm.EnterFunc())
				} else {
					rvm.regSet(dst, objects.None_)
				}
			} else {
				rvm.regSet(dst, objects.None_)
			}

		case RegOpExitContext:
			dst := inst.Operands[0]
			ctxReg := inst.Operands[1]
			excReg := inst.Operands[2]
			ctxManager := rvm.regGet(ctxReg)
			exc := rvm.regGet(excReg)
			if cm, ok := ctxManager.(*objects.ContextManager); ok {
				if cm.ExitFunc != nil {
					rvm.regSet(dst, cm.ExitFunc(exc))
				} else {
					rvm.regSet(dst, objects.None_)
				}
			} else {
				rvm.regSet(dst, objects.None_)
			}

		case RegOpBeginTry:
			exceptCount := inst.Operands[0]
			hasFinally := inst.Operands[1]
			handlerIP := inst.Operands[2]
			finallyStartIP := inst.Operands[3]
			tryBlockStartIP := ip + 1

			handler := ExceptionHandler{
				handlerIP:       handlerIP,
				stackPtr:        vm.sp,
				exceptionType:   "",
				varName:         "",
				handlerStartIP:  -1,
				tryBlockStartIP: tryBlockStartIP,
				hasFinally:      hasFinally == 1,
				finallyStartIP:  finallyStartIP,
				finallyEndIP:    -1,
				pendingError:    nil,
				frameIndex:      vm.framesIndex,
			}
			handler.exceptCount = exceptCount
			handler.baseIP = frame.ip + 1
			vm.exceptionStack = append(vm.exceptionStack, handler)

		case RegOpEndTry:
			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				handler := vm.exceptionStack[lastIdx]
				if vm.sp > handler.stackPtr {
					_ = vm.pop()
				}
				pendingError := handler.pendingError
				vm.exceptionStack = vm.exceptionStack[:lastIdx]
				vm.inFinally = false
				if pendingError != nil {
					caught := vm.raiseException(pendingError)
					if !caught {
						vm.pendingError = pendingError
						return fmt.Errorf("unhandled exception: %s", pendingError.Inspect())
					}
				}
			}

		case RegOpExceptHandler:
			typeIdx := inst.Operands[0]
			varIdx := inst.Operands[1]

			var exceptionType, varName string
			if typeIdx > 0 && typeIdx < len(vm.constants) {
				if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
					exceptionType = typeObj.Value
				}
			}
			if varIdx > 0 && varIdx < len(vm.constants) {
				if varObj, ok := vm.constants[varIdx].(*objects.String); ok {
					varName = varObj.Value
				}
			}

			handlerStartIP := ip + 1

			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				for lastIdx >= 0 && vm.exceptionStack[lastIdx].handlerIP != -1 {
					lastIdx--
				}
				if lastIdx >= 0 {
					existingHandler := vm.exceptionStack[lastIdx]
					vm.exceptionStack[lastIdx] = ExceptionHandler{
						handlerIP:       frame.ip,
						stackPtr:        vm.sp - 1,
						exceptionType:   exceptionType,
						varName:         varName,
						handlerStartIP:  handlerStartIP,
						tryBlockStartIP: existingHandler.tryBlockStartIP,
						hasFinally:      existingHandler.hasFinally,
						finallyStartIP:  existingHandler.finallyStartIP,
						finallyEndIP:    existingHandler.finallyEndIP,
						pendingError:    existingHandler.pendingError,
						exceptOpcodeIP:  frame.ip,
					}
				}
			}

		case RegOpExceptStarHandler:
			typeIdx := inst.Operands[0]
			varIdx := inst.Operands[1]

			var exceptionType, varName string
			if typeIdx > 0 && typeIdx < len(vm.constants) {
				if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
					exceptionType = typeObj.Value
				}
			}
			if varIdx > 0 && varIdx < len(vm.constants) {
				if varObj, ok := vm.constants[varIdx].(*objects.String); ok {
					varName = varObj.Value
				}
			}

			handlerStartIP := ip + 1

			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				for lastIdx >= 0 && vm.exceptionStack[lastIdx].handlerIP != -1 {
					lastIdx--
				}
				if lastIdx >= 0 {
					existingHandler := vm.exceptionStack[lastIdx]
					vm.exceptionStack[lastIdx] = ExceptionHandler{
						handlerIP:       frame.ip,
						stackPtr:        vm.sp - 1,
						exceptionType:   exceptionType,
						varName:         varName,
						handlerStartIP:  handlerStartIP,
						tryBlockStartIP: existingHandler.tryBlockStartIP,
						hasFinally:      existingHandler.hasFinally,
						finallyStartIP:  existingHandler.finallyStartIP,
						finallyEndIP:    existingHandler.finallyEndIP,
						pendingError:    existingHandler.pendingError,
						isStar:          true,
						starUnhandled:   existingHandler.starUnhandled,
						exceptOpcodeIP:  frame.ip,
					}
				}
			}

		case RegOpFinally:
			finallyEndIP := inst.Operands[0]
			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				if lastIdx >= 0 {
					vm.exceptionStack[lastIdx].finallyStartIP = ip + 1
					vm.exceptionStack[lastIdx].finallyEndIP = finallyEndIP
				}
			}
			vm.inFinally = true
			pendingError := vm.pendingError
			vm.pendingError = nil
			if pendingError != nil {
				if len(vm.exceptionStack) > 0 {
					lastIdx := len(vm.exceptionStack) - 1
					if lastIdx >= 0 {
						vm.exceptionStack[lastIdx].pendingError = pendingError
					}
				}
			}

		default:
			return fmt.Errorf("unknown register opcode in executeRegFrame: %d", inst.Opcode)
		}

		ip++
	}

	return nil
}
