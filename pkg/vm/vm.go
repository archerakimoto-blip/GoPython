package vm

import (
	"fmt"
	"strings"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

type VM struct {
	constants    []objects.Object
	instructions compiler.Instructions

	stack   []objects.Object
	sp      int
	globals []objects.Object

	frames      []*Frame
	framesIndex int
}

type Frame struct {
	fn          *compiler.CompiledFunction
	ip          int
	basePointer int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &compiler.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,

		stack:   make([]objects.Object, StackSize),
		sp:      0,
		globals: make([]objects.Object, GlobalSize),

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []objects.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func NewFrame(fn *compiler.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
	}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) Run() error {
	var ip int
	var ins compiler.Instructions
	var op compiler.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().fn.Instructions)-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().fn.Instructions
		op = compiler.Opcode(ins[ip])
		fmt.Println("VM Step: ip=", ip, "op=", op)
		vm.printStack()

		switch op {
		case compiler.OpConstant:
			constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case compiler.OpPop:
			vm.pop()

		case compiler.OpAdd, compiler.OpSub, compiler.OpMul, compiler.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case compiler.OpTrue:
			err := vm.push(objects.True)
			if err != nil {
				return err
			}

		case compiler.OpFalse:
			err := vm.push(objects.False)
			if err != nil {
				return err
			}

		case compiler.OpEqual, compiler.OpNotEqual, compiler.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case compiler.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case compiler.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case compiler.OpJump:
			pos := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip = pos - 1

		case compiler.OpJumpNotTruthy:
			pos := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case compiler.OpNull:
			err := vm.push(objects.None_)
			if err != nil {
				return err
			}

		case compiler.OpSetGlobal:
			globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()

		case compiler.OpGetGlobal:
			globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case compiler.OpArray:
			numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}

		case compiler.OpHash:
			numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}

		case compiler.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case compiler.OpSlice:
			end := vm.pop()
			start := vm.pop()
			left := vm.pop()

			err := vm.executeSliceExpression(left, start, end)
			if err != nil {
				return err
			}

		case compiler.OpCall:
			numArgs := int(ins[ip+1])
			vm.currentFrame().ip += 1

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}

		case compiler.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		case compiler.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(objects.None_)
			if err != nil {
				return err
			}

		case compiler.OpSetLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+localIndex] = vm.pop()

		case compiler.OpGetLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+localIndex])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) printStack() {
	fmt.Print("VM Stack (sp=", vm.sp, "): ")
	for i := 0; i < vm.sp; i++ {
		if vm.stack[i] != nil {
			fmt.Printf("%v ", vm.stack[i].Inspect())
		}
	}
	fmt.Println()
}

func (vm *VM) executeCall(numArgs int) error {
	vm.printStack()
	calleeIndex := vm.sp - 1 - numArgs
	calleeObj := vm.stack[calleeIndex]
	fmt.Println("VM: executeCall called, callee type:", fmt.Sprintf("%T", calleeObj), "callee:", calleeObj.Inspect())
	callee, ok := calleeObj.(*compiler.CompiledFunction)
	if !ok {
		if builtin, ok := calleeObj.(*objects.Builtin); ok {
			args := vm.stack[vm.sp-numArgs : vm.sp]
			fmt.Println("VM: builtin call with args:", args)
			result := builtin.Fn(args...)
			vm.sp = vm.sp - numArgs - 1
			return vm.push(result)
		}
		return fmt.Errorf("calling non-function: type %T", calleeObj)
	}

	if numArgs != callee.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			callee.NumParameters, numArgs)
	}

	frame := NewFrame(callee, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + callee.NumLocals

	return nil
}

func isTruthy(obj objects.Object) bool {
	switch obj := obj.(type) {
	case *objects.Boolean:
		return obj.Value
	case *objects.None:
		return false
	default:
		return true
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != objects.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*objects.Integer).Value
	return vm.push(&objects.Integer{Value: -value})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case objects.True:
		return vm.push(objects.False)
	case objects.False:
		return vm.push(objects.True)
	case objects.None_:
		return vm.push(objects.True)
	default:
		return vm.push(objects.False)
	}
}

func (vm *VM) executeComparison(op compiler.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == objects.INTEGER_OBJ && right.Type() == objects.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == objects.FLOAT_OBJ && right.Type() == objects.FLOAT_OBJ {
		return vm.executeFloatComparison(op, left, right)
	}

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(left == right))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op compiler.Opcode, left, right objects.Object) error {
	leftValue := left.(*objects.Integer).Value
	rightValue := right.(*objects.Integer).Value

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case compiler.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeFloatComparison(op compiler.Opcode, left, right objects.Object) error {
	leftValue := left.(*objects.Float).Value
	rightValue := right.(*objects.Float).Value

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case compiler.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(input bool) *objects.Boolean {
	if input {
		return objects.True
	}
	return objects.False
}

func (vm *VM) executeBinaryOperation(op compiler.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == objects.INTEGER_OBJ && rightType == objects.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	if leftType == objects.FLOAT_OBJ && rightType == objects.FLOAT_OBJ {
		return vm.executeBinaryFloatOperation(op, left, right)
	}

	if op == compiler.OpAdd {
		// If either operand is a string, convert both to string and concatenate
		leftStr := toString(left)
		rightStr := toString(right)
		return vm.push(&objects.String{Value: leftStr + rightStr})
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func toString(obj objects.Object) string {
	switch o := obj.(type) {
	case *objects.String:
		return o.Value
	case *objects.Integer:
		return fmt.Sprintf("%d", o.Value)
	case *objects.Float:
		return fmt.Sprintf("%g", o.Value)
	case *objects.Boolean:
		if o.Value {
			return "True"
		}
		return "False"
	case *objects.None:
		return "None"
	case *objects.List:
		// For simplicity, return a basic representation
		elements := []string{}
		for _, elem := range o.Elements {
			elements = append(elements, toString(elem))
		}
		return "[" + strings.Join(elements, ", ") + "]"
	default:
		return o.Inspect()
	}
}

func (vm *VM) executeBinaryIntegerOperation(op compiler.Opcode, left, right objects.Object) error {
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
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&objects.Integer{Value: result})
}

func (vm *VM) executeBinaryFloatOperation(op compiler.Opcode, left, right objects.Object) error {
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
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	return vm.push(&objects.Float{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op compiler.Opcode, left, right objects.Object) error {
	leftValue := left.(*objects.String).Value
	rightValue := right.(*objects.String).Value

	return vm.push(&objects.String{Value: leftValue + rightValue})
}

func (vm *VM) push(o objects.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() objects.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() objects.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) buildArray(startIndex, endIndex int) objects.Object {
	elements := make([]objects.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &objects.List{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (objects.Object, error) {
	hashedPairs := make(map[objects.Object]objects.Object)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		hashedPairs[key] = value
	}

	return &objects.Dict{Pairs: hashedPairs}, nil
}

func (vm *VM) executeIndexExpression(left, index objects.Object) error {
	switch {
	case left.Type() == objects.LIST_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == objects.DICT_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index objects.Object) error {
	arrayObject := array.(*objects.List)
	idx := index.(*objects.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return vm.push(objects.None_)
	}

	return vm.push(arrayObject.Elements[idx])
}

func (vm *VM) executeHashIndex(hash, index objects.Object) error {
	hashObject := hash.(*objects.Dict)

	pair, ok := hashObject.Pairs[index]
	if !ok {
		return vm.push(objects.None_)
	}

	return vm.push(pair)
}

func (vm *VM) executeSliceExpression(left, start, end objects.Object) error {
	switch {
	case left.Type() == objects.LIST_OBJ:
		return vm.executeListSlice(left, start, end)
	case left.Type() == objects.STRING_OBJ:
		return vm.executeStringSlice(left, start, end)
	default:
		return fmt.Errorf("slice operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeListSlice(left, start, end objects.Object) error {
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
		if e.Value == -1 { // 我们用 -1 表示未指定
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
		return vm.push(&objects.List{Elements: []objects.Object{}})
	}

	elements := make([]objects.Object, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		elements[i-startIdx] = list.Elements[i]
	}

	return vm.push(&objects.List{Elements: elements})
}

func (vm *VM) executeStringSlice(left, start, end objects.Object) error {
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
		return vm.push(&objects.String{Value: ""})
	}

	return vm.push(&objects.String{Value: str.Value[startIdx:endIdx]})
}

