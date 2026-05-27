package vm

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

type VMOptimized struct {
	*VM
	opcodeCache    []opcodeHandler
	constantCache  []objects.Object
	instructionPtr int
}

type opcodeHandler func() error

func NewOptimized(code *compiler.Bytecode) *VMOptimized {
	vm := New(code)
	optimized := &VMOptimized{
		VM:             vm,
		opcodeCache:    make([]opcodeHandler, 256),
		constantCache:  make([]objects.Object, 0),
		instructionPtr: 0,
	}

	optimized.initializeOpcodeCache()
	optimized.cacheConstants(code)

	return optimized
}

func (vm *VMOptimized) initializeOpcodeCache() {
	for i := 0; i < 256; i++ {
		vm.opcodeCache[i] = vm.defaultHandler
	}

	vm.opcodeCache[compiler.OpConstant] = vm.handleConstant
	vm.opcodeCache[compiler.OpPop] = vm.handlePop
	vm.opcodeCache[compiler.OpDupTop] = vm.handleDupTop
	vm.opcodeCache[compiler.OpAdd] = vm.handleAdd
	vm.opcodeCache[compiler.OpSub] = vm.handleSub
	vm.opcodeCache[compiler.OpMul] = vm.handleMul
	vm.opcodeCache[compiler.OpDiv] = vm.handleDiv
	vm.opcodeCache[compiler.OpTrue] = vm.handleTrue
	vm.opcodeCache[compiler.OpFalse] = vm.handleFalse
	vm.opcodeCache[compiler.OpNull] = vm.handleNull
	vm.opcodeCache[compiler.OpEqual] = vm.handleEqual
	vm.opcodeCache[compiler.OpNotEqual] = vm.handleNotEqual
	vm.opcodeCache[compiler.OpGreaterThan] = vm.handleGreaterThan
	vm.opcodeCache[compiler.OpLessThan] = vm.handleLessThan
	vm.opcodeCache[compiler.OpMinus] = vm.handleMinus
	vm.opcodeCache[compiler.OpBang] = vm.handleBang
	vm.opcodeCache[compiler.OpJump] = vm.handleJump
	vm.opcodeCache[compiler.OpJumpNotTruthy] = vm.handleJumpNotTruthy
	vm.opcodeCache[compiler.OpSetGlobal] = vm.handleSetGlobal
	vm.opcodeCache[compiler.OpGetGlobal] = vm.handleGetGlobal
	vm.opcodeCache[compiler.OpGetLocal] = vm.handleGetLocal
	vm.opcodeCache[compiler.OpSetLocal] = vm.handleSetLocal
	vm.opcodeCache[compiler.OpCall] = vm.handleCall
	vm.opcodeCache[compiler.OpReturn] = vm.handleReturn
	vm.opcodeCache[compiler.OpReturnValue] = vm.handleReturnValue
	vm.opcodeCache[compiler.OpArray] = vm.handleArray
	vm.opcodeCache[compiler.OpHash] = vm.handleHash
	vm.opcodeCache[compiler.OpSet] = vm.handleSet
	vm.opcodeCache[compiler.OpIndex] = vm.handleIndex
	vm.opcodeCache[compiler.OpSlice] = vm.handleSlice
	vm.opcodeCache[compiler.OpGetAttribute] = vm.handleGetAttribute
	vm.opcodeCache[compiler.OpSetAttribute] = vm.handleSetAttribute
	vm.opcodeCache[compiler.OpClosure] = vm.handleClosure
}

func (vm *VMOptimized) cacheConstants(code *compiler.Bytecode) {
	vm.constantCache = make([]objects.Object, len(code.Constants))
	copy(vm.constantCache, code.Constants)
}

func (vm *VMOptimized) defaultHandler() error {
	return fmt.Errorf("unknown opcode")
}

func (vm *VMOptimized) handleConstant() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2

	if constIndex < len(vm.constantCache) {
		return vm.push(vm.constantCache[constIndex])
	}
	return vm.push(vm.constants[constIndex])
}

func (vm *VMOptimized) handlePop() error {
	vm.pop()
	return nil
}

func (vm *VMOptimized) handleDupTop() error {
	if vm.sp > 0 {
		return vm.push(vm.stack[vm.sp-1])
	}
	return nil
}

func (vm *VMOptimized) handleAdd() error {
	return vm.executeBinaryOperation(compiler.OpAdd)
}

func (vm *VMOptimized) handleSub() error {
	return vm.executeBinaryOperation(compiler.OpSub)
}

func (vm *VMOptimized) handleMul() error {
	return vm.executeBinaryOperation(compiler.OpMul)
}

func (vm *VMOptimized) handleDiv() error {
	return vm.executeBinaryOperation(compiler.OpDiv)
}

func (vm *VMOptimized) handleTrue() error {
	return vm.push(objects.True)
}

func (vm *VMOptimized) handleFalse() error {
	return vm.push(objects.False)
}

func (vm *VMOptimized) handleNull() error {
	return vm.push(objects.None_)
}

func (vm *VMOptimized) handleEqual() error {
	return vm.executeComparison(compiler.OpEqual)
}

func (vm *VMOptimized) handleNotEqual() error {
	return vm.executeComparison(compiler.OpNotEqual)
}

func (vm *VMOptimized) handleGreaterThan() error {
	return vm.executeComparison(compiler.OpGreaterThan)
}

func (vm *VMOptimized) handleLessThan() error {
	return vm.executeComparison(compiler.OpLessThan)
}

func (vm *VMOptimized) handleMinus() error {
	return vm.executeMinusOperator()
}

func (vm *VMOptimized) handleBang() error {
	return vm.executeBangOperator()
}

func (vm *VMOptimized) handleJump() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	pos := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip = pos - 1
	return nil
}

func (vm *VMOptimized) handleJumpNotTruthy() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	pos := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2

	condition := vm.pop()
	if !isTruthy(condition) {
		vm.currentFrame().ip = pos - 1
	}
	return nil
}

func (vm *VMOptimized) handleSetGlobal() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2
	vm.globals[globalIndex] = vm.pop()
	return nil
}

func (vm *VMOptimized) handleGetGlobal() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2
	return vm.push(vm.globals[globalIndex])
}

func (vm *VMOptimized) handleGetLocal() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	localIndex := int(ins[ip+1])
	vm.currentFrame().ip += 1
	return vm.push(vm.stack[vm.currentFrame().basePointer+localIndex])
}

func (vm *VMOptimized) handleSetLocal() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	localIndex := int(ins[ip+1])
	vm.currentFrame().ip += 1
	vm.stack[vm.currentFrame().basePointer+localIndex] = vm.pop()
	return nil
}

func (vm *VMOptimized) handleCall() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	numArgs := int(ins[ip+1])
	vm.currentFrame().ip += 1
	return vm.executeCall(numArgs)
}

func (vm *VMOptimized) handleReturn() error {
	return nil
}

func (vm *VMOptimized) handleReturnValue() error {
	return nil
}

func (vm *VMOptimized) handleArray() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2

	array := vm.buildArray(vm.sp-numElements, vm.sp)
	vm.sp = vm.sp - numElements
	return vm.push(array)
}

func (vm *VMOptimized) handleHash() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2

	hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
	if err != nil {
		return err
	}
	vm.sp = vm.sp - numElements
	return vm.push(hash)
}

func (vm *VMOptimized) handleSet() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2

	set := vm.buildSet(vm.sp-numElements, vm.sp)
	vm.sp = vm.sp - numElements
	return vm.push(set)
}

func (vm *VMOptimized) handleIndex() error {
	index := vm.pop()
	left := vm.pop()
	return vm.executeIndexExpression(left, index)
}

func (vm *VMOptimized) handleSlice() error {
	end := vm.pop()
	start := vm.pop()
	left := vm.pop()
	return vm.executeSliceExpression(left, start, end)
}

func (vm *VMOptimized) handleGetAttribute() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	_ = int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2

	_ = vm.pop()
	left := vm.pop()
	
	if left == nil {
		return vm.push(objects.None_)
	}
	return vm.push(left)
}

func (vm *VMOptimized) handleSetAttribute() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	_ = int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	vm.currentFrame().ip += 2

	_ = vm.pop()
	_ = vm.pop()
	return nil
}

func (vm *VMOptimized) handleClosure() error {
	ip := vm.currentFrame().ip
	ins := vm.currentFrame().fn.Instructions
	constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
	numFree := int(ins[ip+3])
	vm.currentFrame().ip += 3

	fn, ok := vm.constants[constIndex].(*compiler.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %T", vm.constants[constIndex])
	}

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

	return vm.push(closure)
}

func (vm *VMOptimized) Run() error {
	for vm.currentFrame().ip < len(vm.currentFrame().fn.Instructions)-1 {
		vm.currentFrame().ip++
		op := compiler.Opcode(vm.currentFrame().fn.Instructions[vm.currentFrame().ip])

		if handler := vm.opcodeCache[op]; handler != nil {
			if err := handler(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown opcode: %d", op)
		}

		if vm.sp > 0 && vm.stack[vm.sp-1].Type() == objects.ERROR_OBJ {
			errObj := vm.stack[vm.sp-1]
			caught := vm.raiseException(errObj)
			if !caught {
				return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
			}
		}
	}

	return nil
}