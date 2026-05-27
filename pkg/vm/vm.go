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

type ExceptionHandler struct {
	handlerIP       int
	stackPtr        int
	exceptionType   string
	varName         string
	handlerStartIP  int
	exceptCount     int
	baseIP          int
	tryBlockStartIP int
	hasFinally      bool
	finallyStartIP  int
	finallyEndIP    int
	pendingError    objects.Object
}

type Frame struct {
	fn          *compiler.CompiledFunction
	ip          int
	basePointer int
	generator   *objects.Generator // 指向创建此帧的生成器（如果是生成器帧）
	freeVars    []objects.Object   // 闭包的自由变量值
}

type VM struct {
	constants    []objects.Object
	instructions compiler.Instructions

	stack   []objects.Object
	sp      int
	globals []objects.Object

	frames      []*Frame
	framesIndex int

	exceptionStack []ExceptionHandler // 异常处理器栈
	pendingError   objects.Object     // 待处理的异常
	inFinally      bool               // 是否正在执行 finally 块
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &compiler.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainFrame := NewFrame(mainFn, 1)

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
		generator:   nil,
		freeVars:    nil,
	}
}

func NewFrameWithFreeVars(fn *compiler.CompiledFunction, basePointer int, freeVars []objects.Object) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
		generator:   nil,
		freeVars:    freeVars,
	}
}

func NewFrameFromGenerator(gen *objects.Generator) *Frame {
	fn := &compiler.CompiledFunction{
		Instructions: gen.Instructions,
		NumLocals:    len(gen.Locals),
	}
	for i, val := range gen.Locals {
		gen.Stack[gen.BasePointer+i] = val
	}
	return &Frame{
		fn:          fn,
		ip:          gen.IP,
		basePointer: gen.BasePointer,
		generator:   gen,
	}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) CurrentFrame() *Frame {
	return vm.currentFrame()
}

func (vm *VM) GetFramesIndex() int {
	return vm.framesIndex
}

func (vm *VM) GetFrame(index int) *Frame {
	if index >= 0 && index < vm.framesIndex {
		return vm.frames[index]
	}
	return nil
}

func (vm *VM) GetSP() int {
	return vm.sp
}

func (vm *VM) GetStack(index int) objects.Object {
	if index >= 0 && index < vm.sp {
		return vm.stack[index]
	}
	return nil
}

func (vm *VM) GetGlobals() []objects.Object {
	return vm.globals
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

		switch op {
		case compiler.OpConstant:
			constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case compiler.OpClosure:
			constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			numFree := int(ins[ip+3])
			vm.currentFrame().ip += 3

			fn, ok := vm.constants[constIndex].(*compiler.CompiledFunction)
			if !ok {
				return fmt.Errorf("not a function: %T", vm.constants[constIndex])
			}

			// Pop free variables from stack
			free := make([]objects.Object, numFree)
			for i := numFree - 1; i >= 0; i-- {
				free[i] = vm.pop()
			}

			closure := &objects.Closure{
				Instructions:  fn.Instructions,
				NumLocals:     fn.NumLocals,
				NumParameters: fn.NumParameters,
				IsGenerator:   fn.IsGenerator,
				Free:          free,
			}

			err := vm.push(closure)
			if err != nil {
				return err
			}

		case compiler.OpPop:
			vm.pop()

		case compiler.OpDupTop:
			if vm.sp > 0 {
				top := vm.stack[vm.sp-1]
				err := vm.push(top)
				if err != nil {
					return err
				}
			}

		case compiler.OpAdd, compiler.OpSub, compiler.OpMul, compiler.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
			if vm.sp > 0 && vm.stack[vm.sp-1].Type() == objects.ERROR_OBJ {
				errObj := vm.stack[vm.sp-1]
				caught := vm.raiseException(errObj)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
				}
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

		case compiler.OpEqual, compiler.OpNotEqual, compiler.OpGreaterThan, compiler.OpLessThan:
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

		case compiler.OpSet:
			numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2

			set := vm.buildSet(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(set)
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
			
			if vm.framesIndex == 0 {
				return nil
			}
			
			// 检查是否是从生成器返回
			if len(vm.frames) > 0 && vm.sp > 0 {
				calleeIndex := vm.sp - 1
				if calleeIndex >= 0 {
					if gen, ok := vm.stack[calleeIndex].(*objects.Generator); ok {
						gen.Done = true
					}
				}
			}
			
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		case compiler.OpReturn:
			frame := vm.popFrame()
			
			if vm.framesIndex == 0 {
				return nil
			}
			
			// 检查是否是从生成器返回
			if len(vm.frames) > 0 && vm.sp > 0 {
				calleeIndex := vm.sp - 1
				if calleeIndex >= 0 {
					if gen, ok := vm.stack[calleeIndex].(*objects.Generator); ok {
						gen.Done = true
					}
				}
			}
			
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

		case compiler.OpGetFree:
			freeIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			// Use the stored free variables from the closure
			if frame.freeVars != nil && freeIndex < len(frame.freeVars) {
				err := vm.push(frame.freeVars[freeIndex])
				if err != nil {
					return err
				}
			} else {
				// Fallback: try to get from stack (for non-closure functions)
				err := vm.push(vm.stack[frame.basePointer-len(frame.fn.Free)+freeIndex])
				if err != nil {
					return err
				}
			}

		case compiler.OpBeginTry:
		exceptCount := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
		hasFinally := int(uint16(ins[ip+3])<<8 | uint16(ins[ip+4]))
		vm.currentFrame().ip += 4
		tryBlockStartIP := ip + 5
		handler := ExceptionHandler{
			handlerIP:       -1,
			stackPtr:        vm.sp,
			exceptionType:   "",
			varName:         "",
			handlerStartIP:  -1,
			tryBlockStartIP: tryBlockStartIP,
			hasFinally:      hasFinally == 1,
			finallyStartIP:  -1,
			finallyEndIP:    -1,
			pendingError:    nil,
		}
		handler.exceptCount = exceptCount
		handler.baseIP = vm.currentFrame().ip + 1
		vm.exceptionStack = append(vm.exceptionStack, handler)
		case compiler.OpEndTry:
			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				pendingError := vm.exceptionStack[lastIdx].pendingError
				vm.exceptionStack = vm.exceptionStack[:lastIdx]
				vm.inFinally = false

				if pendingError != nil {
					err := vm.push(pendingError)
					if err != nil {
						return err
					}
					caught := false

					for i := len(vm.exceptionStack) - 1; i >= 0; i-- {
						handler := vm.exceptionStack[i]
						if handler.handlerIP == -1 {
							continue
						}

						if handler.hasFinally {
							vm.exceptionStack[i].pendingError = pendingError
							vm.currentFrame().ip = handler.finallyStartIP - 1
							caught = true
							break
						}

						ip := handler.tryBlockStartIP
						for ip < len(vm.currentFrame().fn.Instructions) {
							op := compiler.Opcode(vm.currentFrame().fn.Instructions[ip])
							if op == compiler.OpExceptHandler {
								typeIdx := int(uint16(vm.currentFrame().fn.Instructions[ip+2])<<8 | uint16(vm.currentFrame().fn.Instructions[ip+1]))
								var exceptionType string
								if typeIdx > 0 && typeIdx < len(vm.constants) {
									if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
										exceptionType = typeObj.Value
									}
								}
								if exceptionType == "" || matchesException(pendingError, exceptionType) {
									vm.sp = handler.stackPtr
									if err := vm.push(pendingError); err != nil {
										return err
									}
									vm.currentFrame().ip = ip + 5 - 1
									caught = true
								}
								break
							}
							if op == compiler.OpFinally {
								break
							}
							ip++
						}
						if caught {
							break
						}
					}

					if !caught {
						vm.pendingError = pendingError
						return fmt.Errorf("unhandled exception: %s", pendingError.Inspect())
					}
				}
			}
		case compiler.OpRaise:
			errObj := vm.pop()
			caught := false

			for i := len(vm.exceptionStack) - 1; i >= 0; i-- {
				handler := vm.exceptionStack[i]
				if handler.handlerIP == -1 {
					continue
				}

				if handler.hasFinally {
					vm.exceptionStack[i].pendingError = errObj
					vm.currentFrame().ip = handler.finallyStartIP - 1
					caught = true
					break
				}

				ip := handler.tryBlockStartIP
				for ip < len(vm.currentFrame().fn.Instructions) {
					op := compiler.Opcode(vm.currentFrame().fn.Instructions[ip])
					if op == compiler.OpExceptHandler {
						typeIdx := int(uint16(vm.currentFrame().fn.Instructions[ip+2])<<8 | uint16(vm.currentFrame().fn.Instructions[ip+1]))
						var exceptionType string
						if typeIdx > 0 && typeIdx < len(vm.constants) {
							if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
								exceptionType = typeObj.Value
							}
						}
						if exceptionType == "" || matchesException(errObj, exceptionType) {
							vm.sp = handler.stackPtr
							if err := vm.push(errObj); err != nil {
								return err
							}
							vm.currentFrame().ip = ip + 5 - 1
							caught = true
						}
						break
					}
					if op == compiler.OpFinally {
						break
					}
					ip++
				}
				if caught {
					break
				}
			}

			if !caught {
				vm.pendingError = errObj
				return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
			}
		case compiler.OpExceptHandler:
			typeIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			varIdx := int(uint16(ins[ip+3])<<8 | uint16(ins[ip+4]))
			vm.currentFrame().ip += 4

			var exceptionType, varName string
			if typeIdx > 0 {
				if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
					exceptionType = typeObj.Value
				}
			}
			if varIdx > 0 {
				if varObj, ok := vm.constants[varIdx].(*objects.String); ok {
					varName = varObj.Value
				}
			}

			handlerStartIP := vm.currentFrame().ip + 1

			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				for lastIdx >= 0 && vm.exceptionStack[lastIdx].handlerIP != -1 {
					lastIdx--
				}
				if lastIdx >= 0 {
					existingHandler := vm.exceptionStack[lastIdx]
					vm.exceptionStack[lastIdx] = ExceptionHandler{
						handlerIP:       vm.currentFrame().ip,
						stackPtr:        vm.sp - 1,
						exceptionType:   exceptionType,
						varName:         varName,
						handlerStartIP:  handlerStartIP,
						tryBlockStartIP: existingHandler.tryBlockStartIP,
						hasFinally:      existingHandler.hasFinally,
						finallyStartIP:  existingHandler.finallyStartIP,
						finallyEndIP:    existingHandler.finallyEndIP,
						pendingError:    existingHandler.pendingError,
					}
				}
			}
		case compiler.OpFinally:
			finallyEndIP := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2

			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				for lastIdx >= 0 && vm.exceptionStack[lastIdx].handlerIP == -1 {
					lastIdx--
				}
				if lastIdx >= 0 {
					vm.exceptionStack[lastIdx].finallyStartIP = ip + 3
					vm.exceptionStack[lastIdx].finallyEndIP = finallyEndIP
				}
			}

			vm.inFinally = true

			pendingError := vm.pendingError
			vm.pendingError = nil

			if pendingError != nil {
				if len(vm.exceptionStack) > 0 {
					lastIdx := len(vm.exceptionStack) - 1
					for lastIdx >= 0 && vm.exceptionStack[lastIdx].handlerIP == -1 {
						lastIdx--
					}
					if lastIdx >= 0 {
						vm.exceptionStack[lastIdx].pendingError = pendingError
					}
				}
			}
		case compiler.OpEnterContext:
			ctxManager := vm.pop()
			if cm, ok := ctxManager.(*objects.ContextManager); ok {
				if cm.EnterFunc != nil {
					result := cm.EnterFunc()
					err := vm.push(result)
					if err != nil {
						return err
					}
				} else {
					err := vm.push(objects.None_)
					if err != nil {
						return err
					}
				}
			} else {
				err := vm.push(objects.None_)
				if err != nil {
					return err
				}
			}
		case compiler.OpExitContext:
			exc := vm.pop()
			ctxManager := vm.pop()
			if cm, ok := ctxManager.(*objects.ContextManager); ok {
				if cm.ExitFunc != nil {
					result := cm.ExitFunc(exc)
					err := vm.push(result)
					if err != nil {
						return err
					}
				} else {
					err := vm.push(objects.None_)
					if err != nil {
						return err
					}
				}
			} else {
				err := vm.push(objects.None_)
				if err != nil {
					return err
				}
			}
		case compiler.OpMakeGenerator:
			fnObj := vm.pop()
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
				err := vm.push(gen)
				if err != nil {
					return err
				}
			}
		case compiler.OpCreateClass:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			class := vm.constants[idx].(*objects.Class)
			vm.currentFrame().ip += 2
			err := vm.push(class)
			if err != nil {
				return err
			}
		case compiler.OpCreateClassWithSuper:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			class := vm.constants[idx].(*objects.Class)
			vm.currentFrame().ip += 2
			
			// Pop super class from stack
			superClass := vm.pop()
			if superClass != nil {
				if superCls, ok := superClass.(*objects.Class); ok {
					class.SuperClass = superCls
					// Inherit methods from super class
					for name, method := range superCls.Methods {
						if _, exists := class.Methods[name]; !exists {
							class.Methods[name] = method
						}
					}
				}
			}
			
			err := vm.push(class)
			if err != nil {
				return err
			}
		case compiler.OpGetAttribute:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			attrName := vm.constants[idx].(*objects.String).Value

			obj := vm.pop()
			
			if instance, ok := obj.(*objects.Instance); ok {
				if val, ok := instance.GetAttr(attrName); ok {
					if method, ok := val.(*compiler.CompiledFunction); ok {
						vm.push(instance)
						vm.push(method)
						continue
					}
					return vm.push(val)
				}
				if classMethod, ok := instance.Class.Methods[attrName]; ok {
					vm.push(instance)
					vm.push(classMethod)
					continue
				}
				return vm.push(objects.None_)
			}
			
			if module, ok := obj.(*objects.Module); ok {
			if val, ok := module.GetAttr(attrName); ok {
				err := vm.push(val)
				if err != nil {
					return err
				}
				continue
			}
			err := vm.push(objects.None_)
			if err != nil {
				return err
			}
			continue
		}
			
			return fmt.Errorf("cannot get attribute on non-instance: %s", obj.Type())
		case compiler.OpSetAttribute:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			attrName := vm.constants[idx].(*objects.String).Value
			vm.currentFrame().ip += 2
			
			value := vm.pop()
			obj := vm.pop()
			
			if instance, ok := obj.(*objects.Instance); ok {
				instance.SetAttr(attrName, value)
				return vm.push(value)
			}
			
			return fmt.Errorf("cannot set attribute on non-instance: %s", obj.Type())
		case compiler.OpFormatString:
			partsCount := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			
			// 从栈上获取所有的部分，按顺序拼接
			var result string
			for i := partsCount - 1; i >= 0; i-- {
				part := vm.pop()
				var partStr string
				
				if strObj, ok := part.(*objects.String); ok {
					partStr = strObj.Value
				} else {
					partStr = part.Inspect()
				}
				
				result = partStr + result
			}
			
			err := vm.push(&objects.String{Value: result})
			if err != nil {
				return err
			}
		case compiler.OpYieldValue:
			frame := vm.currentFrame()
			// 获取要产出的值
			yieldValue := vm.pop()
			
			// 找到生成器对象（它应该在 basePointer-1 位置）
			genIndex := frame.basePointer - 1
			if gen, ok := vm.stack[genIndex].(*objects.Generator); ok {
				// 保存当前状态到生成器对象
				gen.IP = frame.ip + 1
				gen.StackPtr = vm.sp
				gen.BasePointer = frame.basePointer
				
				// 保存局部变量
				copy(gen.Locals, vm.stack[frame.basePointer:])
				// 保存当前栈的完整状态
				copy(gen.Stack[:vm.sp], vm.stack[:vm.sp])
				
				// 恢复调用者栈
				vm.sp = genIndex
				
				// 弹出当前帧
				vm.popFrame()
				
				// 把产出值压到调用者栈上
				return vm.push(yieldValue)
			}
			
			// 如果找不到生成器对象，回退到旧行为（创建新生成器）
			vm.currentFrame().ip--
			gen := &objects.Generator{
				Instructions: vm.currentFrame().fn.Instructions,
				Constants:    vm.constants,
				Locals:       make([]objects.Object, len(vm.stack)-vm.currentFrame().basePointer),
				IP:           vm.currentFrame().ip + 1,
				Stack:        make([]objects.Object, vm.sp),
				StackPtr:     vm.sp,
				BasePointer:  vm.currentFrame().basePointer,
				Done:         false,
			}
			copy(gen.Locals, vm.stack[vm.currentFrame().basePointer:])
			copy(gen.Stack, vm.stack[:vm.sp])
			vm.sp = vm.currentFrame().basePointer - 1
			vm.popFrame()
			vm.push(gen)
	}
	}

	return nil
}

func (vm *VM) executeCall(numArgs int) error {
	calleeIndex := vm.sp - numArgs - 1
	
	calleeObj := vm.stack[calleeIndex]

	if classObj, ok := calleeObj.(*objects.Class); ok {
		instance := &objects.Instance{
			Class:  classObj,
			Fields: make(map[string]objects.Object),
		}

		for i := 0; i < numArgs; i++ {
			instance.Fields[fmt.Sprintf("arg%d", i)] = vm.stack[vm.sp-numArgs+i]
		}

		vm.sp = vm.sp - numArgs
		vm.stack[vm.sp] = instance
		vm.sp++

		if initMethod, ok := classObj.Methods["__init__"]; ok {
			if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
				frame := NewFrame(fn, vm.sp-numArgs)
				vm.pushFrame(frame)
			}
		}

		return nil
	}

	// Check if this is a method call where instance is on the stack
	// Stack is [..., prev, instance, method] and we're calling method with numArgs
	// We need to arrange stack as [..., prev, method, self, arg0, arg1, ...]
	// where calleeIndex points to method and method will be called with self + numArgs arguments
	if calleeIndex > 0 && numArgs == 0 {
		if _, isMethod := calleeObj.(*compiler.CompiledFunction); isMethod {
			if instance, isInstance := vm.stack[calleeIndex-1].(*objects.Instance); isInstance {
				// This is a method call with no additional arguments
				// Current stack: [..., prev, instance, method]
				// Rearrange to: [..., prev, method, self]
				vm.stack[calleeIndex-1] = calleeObj  // method
				vm.stack[calleeIndex] = instance       // self
				// vm.sp stays the same (method is already at calleeIndex)
				numArgs = 1  // method needs self as its only argument
			}
		}
	}

	if gen, ok := calleeObj.(*objects.Generator); ok {
			if gen.Done {
				vm.sp = vm.sp - numArgs - 1
				return vm.push(objects.NewError("StopIteration: generator is exhausted"))
			}
		// 保存生成器对象引用，用于后续在栈上找到它
		// 先移除生成器对象，后面在恢复栈后再放回去
		genObj := vm.stack[calleeIndex]
		vm.sp = calleeIndex
		// 恢复生成器状态
		frame := NewFrameFromGenerator(gen)
		vm.pushFrame(frame)
		// 恢复 VM 的栈到生成器保存的状态
		copy(vm.stack[:gen.StackPtr], gen.Stack[:gen.StackPtr])
		vm.sp = gen.StackPtr
		// 把生成器对象放回栈上，用于后续在 OpYieldValue 或返回时找到
		// 我们把它放在 frame.basePointer - 1 的位置，就像普通函数调用那样
		vm.stack[calleeIndex] = genObj
		vm.sp = calleeIndex + 1
		// 让 Run() 继续执行
		return nil
	}

	// Check if it's a closure
	if closure, ok := calleeObj.(*objects.Closure); ok {
		if numArgs != closure.NumParameters {
			return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
				closure.NumParameters, numArgs)
		}

		// Create a CompiledFunction from the closure
		fn := &compiler.CompiledFunction{
			Instructions:  closure.Instructions,
			NumLocals:     closure.NumLocals,
			NumParameters: closure.NumParameters,
			IsGenerator:   closure.IsGenerator,
		}

		// Store a copy of free variables for OpGetFree to use
		freeVarsCopy := make([]objects.Object, len(closure.Free))
		copy(freeVarsCopy, closure.Free)

		// basePointer points to the first argument
		// Free variables are stored before basePointer
		frame := NewFrameWithFreeVars(fn, vm.sp-numArgs, freeVarsCopy)
		vm.pushFrame(frame)
		vm.sp = frame.basePointer + closure.NumLocals

		return nil
	}

	callee, ok := calleeObj.(*compiler.CompiledFunction)
	if !ok {
		if builtin, ok := calleeObj.(*objects.Builtin); ok {
			args := vm.stack[vm.sp-numArgs : vm.sp]
			result := builtin.Fn(args...)
			vm.sp = vm.sp - numArgs - 1
			return vm.push(result)
		}
		return fmt.Errorf("calling non-function: type %T", calleeObj)
	}

	if callee.IsGenerator {
		gen := &objects.Generator{
			Instructions: callee.Instructions,
			Constants:    vm.constants,
			Locals:       make([]objects.Object, callee.NumLocals),
			IP:           -1,
			Stack:        make([]objects.Object, StackSize),
			StackPtr:     0,
			BasePointer:  vm.sp - numArgs,
			Done:         false,
		}
		for i := 0; i < numArgs; i++ {
			gen.Locals[i] = vm.stack[vm.sp-numArgs+i]
		}
		vm.sp = vm.sp - numArgs - 1
		return vm.push(gen)
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
	case compiler.OpLessThan:
		return vm.push(nativeBoolToBooleanObject(leftValue < rightValue))
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
	case compiler.OpLessThan:
		return vm.push(nativeBoolToBooleanObject(leftValue < rightValue))
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
		if rightValue == 0 {
			return vm.push(objects.NewZeroDivisionError("division by zero"))
		}
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	err := vm.push(&objects.Integer{Value: result})
	if err != nil {
		return err
	}

	return nil
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
	if vm.sp > 0 {
		return vm.stack[vm.sp-1]
	}
	return nil
}

func (vm *VM) TopStackElem() objects.Object {
	if vm.sp > 0 {
		return vm.stack[vm.sp-1]
	}
	return nil
}

func (vm *VM) buildArray(startIndex, endIndex int) objects.Object {
	elements := make([]objects.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return objects.NewList(elements)
}

func (vm *VM) buildHash(startIndex, endIndex int) (objects.Object, error) {
	dict := objects.NewDict()

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		dict.Set(key, value)
	}

	return dict, nil
}

func (vm *VM) buildSet(startIndex, endIndex int) objects.Object {
	set := objects.NewSet()

	for i := startIndex; i < endIndex; i++ {
		element := vm.stack[i]
		set.Add(element)
	}

	return set
}

func (vm *VM) executeIndexExpression(left, index objects.Object) error {
	switch {
	case left.Type() == objects.LIST_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == objects.TUPLE_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeTupleIndex(left, index)
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

func (vm *VM) executeTupleIndex(tuple, index objects.Object) error {
	tupleObject := tuple.(*objects.Tuple)
	idx := index.(*objects.Integer).Value
	max := int64(len(tupleObject.Elements) - 1)

	if idx < 0 || idx > max {
		return vm.push(objects.None_)
	}

	return vm.push(tupleObject.Elements[idx])
}

func (vm *VM) executeHashIndex(hash, index objects.Object) error {
	hashObject := hash.(*objects.Dict)

	value, ok := hashObject.Get(index)
	if !ok {
		return vm.push(objects.None_)
	}

	return vm.push(value)
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

func matchesException(errObj objects.Object, exceptionType string) bool {
	if errObj.Type() == objects.ERROR_OBJ {
		err := errObj.(*objects.Error)
		if exceptionType == "" {
			return true
		}
		if exceptionType == "Exception" || exceptionType == "Error" {
			return true
		}
		return err.ErrorType == exceptionType
	}
	return false
}

func (vm *VM) raiseException(errObj objects.Object) bool {
	ins := vm.currentFrame().fn.Instructions

	for i := len(vm.exceptionStack) - 1; i >= 0; i-- {
		handler := vm.exceptionStack[i]
		
		if handler.tryBlockStartIP <= 0 {
			continue
		}

		foundHandler := false
		ip := handler.tryBlockStartIP
		
		for ip < len(ins) {
			op := compiler.Opcode(ins[ip])
			if op == compiler.OpExceptHandler {
				typeIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
				var exceptionType string
				if typeIdx > 0 && typeIdx < len(vm.constants) {
					if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
						exceptionType = typeObj.Value
					}
				}
				if exceptionType == "" || matchesException(errObj, exceptionType) {
					vm.sp = handler.stackPtr
					if err := vm.push(errObj); err != nil {
						return false
					}
					vm.currentFrame().ip = ip + 5 - 1
					foundHandler = true
					break
				}
				ip += 5
			} else if op == compiler.OpFinally {
				break
			} else if op == compiler.OpJump || op == compiler.OpJumpNotTruthy {
				ip += 3
			} else if op == compiler.OpConstant || op == compiler.OpGetGlobal || op == compiler.OpSetGlobal || op == compiler.OpArray || op == compiler.OpHash || op == compiler.OpSet {
				ip += 3
			} else if op == compiler.OpCall {
				ip += 2
			} else if op == compiler.OpGetLocal || op == compiler.OpSetLocal {
				ip += 2
			} else {
				ip++
			}
		}

		if foundHandler {
			return true
		}

		if handler.hasFinally && handler.finallyStartIP > 0 {
			vm.exceptionStack[i].pendingError = errObj
			vm.currentFrame().ip = handler.finallyStartIP - 1
			return true
		}
	}
	return false
}

func (vm *VM) matchesExceptionType(errObj objects.Object, exceptionType string) bool {
	if errObj.Type() == objects.ERROR_OBJ {
		return true
	}
	return exceptionType == "" || exceptionType == "Exception" || exceptionType == "Error"
}

func (vm *VM) findMatchingExceptHandlerFrom(startIP int, errObj objects.Object, exceptCount int) int {
	ip := startIP
	for i := 0; i < exceptCount; i++ {
		if ip >= len(vm.currentFrame().fn.Instructions) {
			break
		}
		op := compiler.Opcode(vm.currentFrame().fn.Instructions[ip])
		if op == compiler.OpExceptHandler {
			typeIdx := int(uint16(vm.currentFrame().fn.Instructions[ip+2])<<8 | uint16(vm.currentFrame().fn.Instructions[ip+1]))

			var exceptionType string
			if typeIdx > 0 && typeIdx < len(vm.constants) {
				if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
					exceptionType = typeObj.Value
				}
			}

			if exceptionType == "" || matchesException(errObj, exceptionType) {
				handlerStartIP := ip + 5
				vm.currentFrame().ip = handlerStartIP
				return handlerStartIP
			}

			if ip+7 < len(vm.currentFrame().fn.Instructions) {
				jumpIP := int(uint16(vm.currentFrame().fn.Instructions[ip+6])<<8 | uint16(vm.currentFrame().fn.Instructions[ip+7]))
				ip = jumpIP
			} else {
				break
			}
		} else {
			break
		}
	}
	return -1
}

