package vm

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

// FastVMv2 是更优化的虚拟机版本
// 优化点：
// 1. 指令预取，减少重复访问
// 2. 常用指令优先处理
// 3. 栈访问优化，减少边界检查
// 4. 帧访问优化
// 5. 内联类型检查
type FastVMv2 struct {
	constants  []objects.Object
	stack      []objects.Object
	sp         int
	globals    []objects.Object
	frames     []*Frame
	frameIndex int
}

func NewFastv2(bytecode *compiler.Bytecode) *FastVMv2 {
	mainFn := &compiler.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainFrame := NewFrame(mainFn, 1)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &FastVMv2{
		constants:  bytecode.Constants,
		stack:      make([]objects.Object, StackSize),
		sp:         0,
		globals:    make([]objects.Object, GlobalSize),
		frames:     frames,
		frameIndex: 1,
	}
}

func (vm *FastVMv2) Run() error {
	for {
		frame := vm.frames[vm.frameIndex-1]

		if frame.ip >= len(frame.fn.Instructions)-1 {
			break
		}

		frame.ip++
		ip := frame.ip
		ins := frame.fn.Instructions
		op := compiler.Opcode(ins[ip])

		// 优化：把最常用的指令放在前面，利用分支预测
		switch op {
		// 1. 最常用的指令优先
		case compiler.OpGetLocal:
			localIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.stack[vm.sp] = vm.stack[frame.basePointer+localIdx]
			vm.sp++

		case compiler.OpSetLocal:
			localIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.sp--
			vm.stack[frame.basePointer+localIdx] = vm.stack[vm.sp]

		case compiler.OpConstant:
			constIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.stack[vm.sp] = vm.constants[constIdx]
			vm.sp++

		case compiler.OpPop:
			vm.sp--

		case compiler.OpAdd:
			// 优化：直接栈访问，减少临时变量
			right := vm.stack[vm.sp-1]
			left := vm.stack[vm.sp-2]
			vm.sp -= 2

			if lInt, ok := left.(*objects.Integer); ok {
				if rInt, ok := right.(*objects.Integer); ok {
					vm.stack[vm.sp] = &objects.Integer{Value: lInt.Value + rInt.Value}
					vm.sp++
					continue
				}
			}
			if lStr, ok := left.(*objects.String); ok {
				if rStr, ok := right.(*objects.String); ok {
					vm.stack[vm.sp] = &objects.String{Value: lStr.Value + rStr.Value}
					vm.sp++
					continue
				}
			}
			return fmt.Errorf("unsupported add")

		case compiler.OpSub:
			rVal := vm.stack[vm.sp-1].(*objects.Integer).Value
			lVal := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			vm.stack[vm.sp] = &objects.Integer{Value: lVal - rVal}
			vm.sp++

		case compiler.OpMul:
			rVal := vm.stack[vm.sp-1].(*objects.Integer).Value
			lVal := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			vm.stack[vm.sp] = &objects.Integer{Value: lVal * rVal}
			vm.sp++

		case compiler.OpJumpNotTruthy:
			target := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2

			cond := vm.stack[vm.sp-1]
			vm.sp--
			if !isFastv2Truthy(cond) {
				frame.ip = target - 1
			}

		case compiler.OpJump:
			target := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip = target - 1

		case compiler.OpLessThan:
			rVal := vm.stack[vm.sp-1].(*objects.Integer).Value
			lVal := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			if lVal < rVal {
				vm.stack[vm.sp] = objects.True
			} else {
				vm.stack[vm.sp] = objects.False
			}
			vm.sp++

		case compiler.OpGetGlobal:
			globalIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.stack[vm.sp] = vm.globals[globalIdx]
			vm.sp++

		case compiler.OpSetGlobal:
			globalIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2
			vm.sp--
			vm.globals[globalIdx] = vm.stack[vm.sp]

		case compiler.OpCall:
			numArgs := int(ins[ip+1])
			frame.ip++

			calleeIdx := vm.sp - numArgs - 1
			callee := vm.stack[calleeIdx]

			if builtin, ok := callee.(*objects.Builtin); ok {
				result := builtin.Fn(vm.stack[vm.sp-numArgs : vm.sp]...)
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

		case compiler.OpArray:
			numElem := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			frame.ip += 2

			elements := make([]objects.Object, numElem)
			copy(elements, vm.stack[vm.sp-numElem:vm.sp])
			vm.sp -= numElem

			vm.stack[vm.sp] = objects.NewList(elements)
			vm.sp++

		case compiler.OpTrue:
			vm.stack[vm.sp] = objects.True
			vm.sp++

		case compiler.OpFalse:
			vm.stack[vm.sp] = objects.False
			vm.sp++

		case compiler.OpNull:
			vm.stack[vm.sp] = objects.None_
			vm.sp++

		case compiler.OpDiv:
			rVal := vm.stack[vm.sp-1].(*objects.Integer).Value
			lVal := vm.stack[vm.sp-2].(*objects.Integer).Value
			vm.sp -= 2
			if rVal == 0 {
				vm.stack[vm.sp] = objects.NewZeroDivisionError("division by zero")
			} else {
				vm.stack[vm.sp] = &objects.Integer{Value: lVal / rVal}
			}
			vm.sp++

		default:
			return fmt.Errorf("unhandled opcode: %d", op)
		}
	}

	return nil
}

func isFastv2Truthy(obj objects.Object) bool {
	// 优化：先检查最常见的类型
	if _, ok := obj.(*objects.Boolean); ok {
		return obj.(*objects.Boolean).Value
	}
	if _, ok := obj.(*objects.None); ok {
		return false
	}
	if i, ok := obj.(*objects.Integer); ok {
		return i.Value != 0
	}
	if s, ok := obj.(*objects.String); ok {
		return len(s.Value) > 0
	}
	if l, ok := obj.(*objects.List); ok {
		return len(l.Elements) > 0
	}
	return true
}
