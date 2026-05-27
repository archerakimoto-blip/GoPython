package vm

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

// FastVM 是优化后的虚拟机，移除了调试代码，优化了核心循环
type FastVM struct {
	constants  []objects.Object
	stack      []objects.Object
	sp         int
	globals    []objects.Object
	frames     []*Frame
	frameIndex int
}

func NewFast(bytecode *compiler.Bytecode) *FastVM {
	mainFn := &compiler.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainFrame := NewFrame(mainFn, 1)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &FastVM{
		constants:  bytecode.Constants,
		stack:      make([]objects.Object, StackSize),
		sp:         0,
		globals:    make([]objects.Object, GlobalSize),
		frames:     frames,
		frameIndex: 1,
	}
}

func (vm *FastVM) Run() error {
	for {
		frame := vm.frames[vm.frameIndex-1]

		if frame.ip >= len(frame.fn.Instructions)-1 {
			break
		}

		frame.ip++
		ip := frame.ip
		ins := frame.fn.Instructions
		op := compiler.Opcode(ins[ip])

		switch op {
		case compiler.OpConstant:
			constIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.stack[vm.sp] = vm.constants[constIdx]
			vm.sp++

		case compiler.OpPop:
			vm.sp--

		case compiler.OpAdd:
			right := vm.stack[vm.sp-1]
			left := vm.stack[vm.sp-2]
			vm.sp -= 2

			if left.Type() == objects.INTEGER_OBJ && right.Type() == objects.INTEGER_OBJ {
				result := left.(*objects.Integer).Value + right.(*objects.Integer).Value
				vm.stack[vm.sp] = &objects.Integer{Value: result}
				vm.sp++
			} else if left.Type() == objects.STRING_OBJ && right.Type() == objects.STRING_OBJ {
				result := left.(*objects.String).Value + right.(*objects.String).Value
				vm.stack[vm.sp] = &objects.String{Value: result}
				vm.sp++
			} else {
				return fmt.Errorf("unsupported add")
			}

		case compiler.OpSub:
			right := vm.stack[vm.sp-1].(*objects.Integer).Value
			left := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			vm.stack[vm.sp] = &objects.Integer{Value: left - right}
			vm.sp++

		case compiler.OpMul:
			right := vm.stack[vm.sp-1].(*objects.Integer).Value
			left := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			vm.stack[vm.sp] = &objects.Integer{Value: left * right}
			vm.sp++

		case compiler.OpDiv:
			right := vm.stack[vm.sp-1].(*objects.Integer).Value
			left := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			if right == 0 {
				vm.stack[vm.sp] = objects.NewZeroDivisionError("division by zero")
				vm.sp++
			} else {
				vm.stack[vm.sp] = &objects.Integer{Value: left / right}
				vm.sp++
			}

		case compiler.OpLessThan:
			right := vm.stack[vm.sp-1].(*objects.Integer).Value
			left := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			if left < right {
				vm.stack[vm.sp] = objects.True
			} else {
				vm.stack[vm.sp] = objects.False
			}
			vm.sp++

		case compiler.OpJump:
			target := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip = target - 1

		case compiler.OpJumpNotTruthy:
			target := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2

			cond := vm.stack[vm.sp-1]
			vm.sp--
			if !isFastTruthy(cond) {
				frame.ip = target - 1
			}

		case compiler.OpNull:
			vm.stack[vm.sp] = objects.None_
			vm.sp++

		case compiler.OpTrue:
			vm.stack[vm.sp] = objects.True
			vm.sp++

		case compiler.OpFalse:
			vm.stack[vm.sp] = objects.False
			vm.sp++

		case compiler.OpSetGlobal:
			globalIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.sp--
			vm.globals[globalIdx] = vm.stack[vm.sp]

		case compiler.OpGetGlobal:
			globalIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.stack[vm.sp] = vm.globals[globalIdx]
			vm.sp++

		case compiler.OpSetLocal:
			localIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.sp--
			vm.stack[frame.basePointer+localIdx] = vm.stack[vm.sp]

		case compiler.OpGetLocal:
			localIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.stack[vm.sp] = vm.stack[frame.basePointer+localIdx]
			vm.sp++

		case compiler.OpArray:
			numElem := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2

			elements := make([]objects.Object, numElem)
			copy(elements, vm.stack[vm.sp-numElem:vm.sp])
			vm.sp -= numElem

			vm.stack[vm.sp] = objects.NewList(elements)
			vm.sp++

		case compiler.OpCall:
			numArgs := int(ins[ip+1])
			frame.ip++

			calleeIdx := vm.sp - numArgs - 1
			callee := vm.stack[calleeIdx]

			if builtin, ok := callee.(*objects.Builtin); ok {
				args := vm.stack[vm.sp-numArgs : vm.sp]
				result := builtin.Fn(args...)
				vm.sp = calleeIdx
				vm.stack[vm.sp] = result
				vm.sp++
			} else if fn, ok := callee.(*compiler.CompiledFunction); ok {
				if numArgs != fn.NumParameters {
					return fmt.Errorf("wrong number of arguments")
				}
				newFrame := NewFrame(fn, vm.sp-numArgs)
				vm.frames[vm.frameIndex] = newFrame
				vm.frameIndex++
				vm.sp = newFrame.basePointer + fn.NumLocals
			} else if closure, ok := callee.(*objects.Closure); ok {
				if numArgs != closure.NumParameters {
					return fmt.Errorf("wrong number of arguments")
				}
				fnObj := &compiler.CompiledFunction{
					Instructions:  closure.Instructions,
					NumLocals:     closure.NumLocals,
					NumParameters: closure.NumParameters,
					IsGenerator:   closure.IsGenerator,
				}
				freeVarsCopy := make([]objects.Object, len(closure.Free))
				copy(freeVarsCopy, closure.Free)
				newFrame := NewFrameWithFreeVars(fnObj, vm.sp-numArgs, freeVarsCopy)
				vm.frames[vm.frameIndex] = newFrame
				vm.frameIndex++
				vm.sp = newFrame.basePointer + fnObj.NumLocals
			}

		case compiler.OpReturnValue:
			returnValue := vm.stack[vm.sp-1]
			vm.sp--

			vm.frameIndex--
			frame = vm.frames[vm.frameIndex]

			vm.sp = frame.basePointer - 1
			vm.stack[vm.sp] = returnValue
			vm.sp++

		case compiler.OpReturn:
			vm.frameIndex--
			frame = vm.frames[vm.frameIndex]

			vm.sp = frame.basePointer - 1
			vm.stack[vm.sp] = objects.None_
			vm.sp++

		default:
			return fmt.Errorf("unhandled opcode: %d", op)
		}
	}

	return nil
}

func isFastTruthy(obj objects.Object) bool {
	switch obj := obj.(type) {
	case *objects.Boolean:
		return obj.Value
	case *objects.None:
		return false
	default:
		return true
	}
}
