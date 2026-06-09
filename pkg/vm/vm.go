package vm

import (
	"fmt"
	"math"
	"math/cmplx"
	"strings"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/gc"
	"github.com/go-py/go-python/pkg/jit"
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
	frameIndex      int
	isStar          bool            // except* handler
	starUnhandled   []objects.Object // unhandled exceptions from ExceptionGroup for except*
	exceptOpcodeIP  int             // IP of the OpExceptHandler/OpExceptStarHandler opcode itself
}

type Frame struct {
	fn                  *compiler.CompiledFunction
	ip                  int
	basePointer         int
	generator           *objects.Generator
	freeVars            []objects.Object
	initInstance        objects.Object
	setAttrValue        objects.Object
	metaclassInitClass  *objects.Class
}

type AttrCacheKey struct {
	FrameIndex int
	IP         int
	ObjType    objects.ObjectType
}

type AttrCacheEntry struct {
	Value     objects.Object
	IsMethod  bool
	ClassName string
}

// IndexCacheKey 用于 OpIndex/OpSetIndex 的内联缓存
type IndexCacheKey struct {
	FrameIndex int
	IP         int
}

// IndexCacheEntry 缓存索引操作的类型分派结果
type IndexCacheEntry struct {
	LeftType  objects.ObjectType
	IndexType objects.ObjectType
	Handler   int // 0=list+int, 1=tuple+int, 2=dict, 3=range+int, 4=string+int, 5=bytes+int, 6=dict_keys+int, 7=dict_values+int, 8=dict_items+int, 9=__getitem__
}

const (
	IndexHandlerListInt      = 0
	IndexHandlerTupleInt     = 1
	IndexHandlerDict         = 2
	IndexHandlerRangeInt     = 3
	IndexHandlerStringInt    = 4
	IndexHandlerBytesInt     = 5
	IndexHandlerDictKeysInt  = 6
	IndexHandlerDictValuesInt = 7
	IndexHandlerDictItemsInt = 8
	IndexHandlerGetItem      = 9
)

type GlobalCacheEntry struct {
	Value   objects.Object
	Version uint64
}

type VM struct {
	constants    []objects.Object
	instructions compiler.Instructions

	stack   []objects.Object
	sp      int
	globals []objects.Object
	lastPopped objects.Object // 最后一次从栈弹出的元素

	frames      []*Frame
	framesIndex int

	exceptionStack []ExceptionHandler // 异常处理器栈
	pendingError   objects.Object     // 待处理的异常
	inFinally      bool               // 是否正在执行 finally 块

	unpackExtraArgs int // OpListUnpack 展开的额外参数数量（用于调整 OpCall 的 numArgs）

	gcEnabled      bool              // 是否启用垃圾回收
	gcThreshold    int64             // 垃圾回收阈值（字节）
	allocatedBytes int64             // 当前已分配字节数

	attrCache      map[AttrCacheKey]AttrCacheEntry
	globalCache    []GlobalCacheEntry
	globalVersions []uint64
	indexCache     map[IndexCacheKey]IndexCacheEntry
	jitEngine      *jit.JIT

	useRegisterVM bool
	regVM         *RegisterVM
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &compiler.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainFrame := NewFrame(mainFn, 1)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	vm := &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,

		stack:   make([]objects.Object, StackSize),
		sp:      0,
		globals: make([]objects.Object, GlobalSize),

		frames:      frames,
		framesIndex: 1,

		gcEnabled:      true,
		gcThreshold:    1024 * 1024,
		allocatedBytes: 0,
		attrCache:      make(map[AttrCacheKey]AttrCacheEntry),
		globalCache:    make([]GlobalCacheEntry, GlobalSize),
		globalVersions: make([]uint64, GlobalSize),
		indexCache:     make(map[IndexCacheKey]IndexCacheEntry),
		jitEngine:      jit.New(),
	}

	// Register callback for calling Python functions from Go code (e.g., re.sub with callable repl)
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		return vm.CallCallable(callee, args...)
	})

	return vm
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []objects.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

// NewRegisterVM creates a VM that uses the register-based execution mode.
// It translates stack-based bytecode to register-based bytecode at runtime
// and executes it, providing an optimization over the standard stack-based VM.
func NewRegisterVM(bytecode *compiler.Bytecode) *VM {
	vm := New(bytecode)
	vm.useRegisterVM = true
	vm.regVM = newRegisterVM(vm)
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

		case compiler.OpAdd, compiler.OpSub, compiler.OpMul, compiler.OpDiv, compiler.OpMod, compiler.OpFloorDiv, compiler.OpPower, compiler.OpBitOr, compiler.OpBitAnd, compiler.OpBitXor, compiler.OpLShift, compiler.OpRShift:
			var binaryErr error
			// 快速整数算术路径：避免函数调用开销
			right := vm.stack[vm.sp-1]
			left := vm.stack[vm.sp-2]
			if leftInt, ok := left.(*objects.Integer); ok {
				if rightInt, ok := right.(*objects.Integer); ok {
					vm.sp--
					var result int64
					switch op {
					case compiler.OpAdd:
						result = leftInt.Value + rightInt.Value
					case compiler.OpSub:
						result = leftInt.Value - rightInt.Value
					case compiler.OpMul:
						result = leftInt.Value * rightInt.Value
					case compiler.OpDiv:
						if rightInt.Value == 0 {
							vm.push(objects.NewError("division by zero"))
							goto checkBinaryError
						}
						vm.stack[vm.sp-1] = &objects.Float{Value: float64(leftInt.Value) / float64(rightInt.Value)}
						continue
					case compiler.OpMod:
						if rightInt.Value == 0 {
							vm.push(objects.NewError("modulo by zero"))
							goto checkBinaryError
						}
						result = leftInt.Value % rightInt.Value
					case compiler.OpFloorDiv:
						if rightInt.Value == 0 {
							vm.push(objects.NewError("integer division or modulo by zero"))
							goto checkBinaryError
						}
						result = leftInt.Value / rightInt.Value
					case compiler.OpPower:
						result = int64(math.Pow(float64(leftInt.Value), float64(rightInt.Value)))
					case compiler.OpBitOr:
						result = leftInt.Value | rightInt.Value
					case compiler.OpBitAnd:
						result = leftInt.Value & rightInt.Value
					case compiler.OpBitXor:
						result = leftInt.Value ^ rightInt.Value
					case compiler.OpLShift:
						result = leftInt.Value << uint(rightInt.Value)
					case compiler.OpRShift:
						result = leftInt.Value >> uint(rightInt.Value)
					default:
						vm.sp++
						goto slowBinaryPath
					}
					vm.stack[vm.sp-1] = &objects.Integer{Value: result}
					continue
				}
			}
		slowBinaryPath:
			binaryErr = vm.executeBinaryOperation(op)
			if binaryErr != nil {
				return binaryErr
			}
		checkBinaryError:
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

		case compiler.OpEqual, compiler.OpNotEqual, compiler.OpGreaterThan, compiler.OpLessThan, compiler.OpGreaterEqual, compiler.OpLessEqual:
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

		case compiler.OpContains:
			err := vm.executeContainsOp(false)
			if err != nil {
				return err
			}

		case compiler.OpNotContains:
			err := vm.executeContainsOp(true)
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

		case compiler.OpEllipsis:
			err := vm.push(objects.EllipsisSingleton)
			if err != nil {
				return err
			}

		case compiler.OpSetGlobal:
			globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			val := vm.pop()
			vm.globals[globalIndex] = val
			vm.globalVersions[globalIndex]++
			vm.globalCache[globalIndex] = GlobalCacheEntry{Value: val, Version: vm.globalVersions[globalIndex]}

		case compiler.OpGetGlobal:
			globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			cached := vm.globalCache[globalIndex]
			var val objects.Object
			if cached.Version == vm.globalVersions[globalIndex] && cached.Value != nil {
				val = cached.Value
			} else {
				val = vm.globals[globalIndex]
				vm.globalCache[globalIndex] = GlobalCacheEntry{Value: val, Version: vm.globalVersions[globalIndex]}
			}
			err := vm.push(val)
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
				if errObj, ok := err.(*objects.Error); ok {
					vm.sp = vm.sp - numElements
					caught := vm.raiseException(errObj)
					if !caught {
						return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
					}
					continue
				}
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

			if set.Type() == objects.ERROR_OBJ {
				caught := vm.raiseException(set)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", set.Inspect())
				}
				continue
			}

			err := vm.push(set)
			if err != nil {
				return err
			}

		case compiler.OpIndex:
			index := vm.pop()
			left := vm.pop()

			// 内联缓存快速路径
			cacheKey := IndexCacheKey{FrameIndex: vm.framesIndex - 1, IP: ip}
			leftType := left.Type()
			indexType := index.Type()
			if entry, hit := vm.indexCache[cacheKey]; hit {
				if entry.LeftType == leftType && entry.IndexType == indexType {
					switch entry.Handler {
					case IndexHandlerListInt:
						vm.executeArrayIndex(left, index)
						continue
					case IndexHandlerTupleInt:
						vm.executeTupleIndex(left, index)
						continue
					case IndexHandlerDict:
						vm.executeHashIndex(left, index)
						continue
					case IndexHandlerStringInt:
						vm.executeStringIndex(left, index)
						continue
					case IndexHandlerRangeInt:
						vm.executeRangeIndex(left, index)
						continue
					case IndexHandlerBytesInt:
						vm.executeBytesIndex(left, index)
						continue
					}
				}
			}

			// 缓存未命中，走正常路径并填充缓存
			handler := -1
			switch {
			case leftType == objects.LIST_OBJ && indexType == objects.INTEGER_OBJ:
				handler = IndexHandlerListInt
				vm.executeArrayIndex(left, index)
			case leftType == objects.TUPLE_OBJ && indexType == objects.INTEGER_OBJ:
				handler = IndexHandlerTupleInt
				vm.executeTupleIndex(left, index)
			case leftType == objects.DICT_OBJ:
				handler = IndexHandlerDict
				vm.executeHashIndex(left, index)
			case leftType == objects.RANGE_OBJ && indexType == objects.INTEGER_OBJ:
				handler = IndexHandlerRangeInt
				vm.executeRangeIndex(left, index)
			case leftType == objects.STRING_OBJ && indexType == objects.INTEGER_OBJ:
				handler = IndexHandlerStringInt
				vm.executeStringIndex(left, index)
			case leftType == objects.BYTES_OBJ && indexType == objects.INTEGER_OBJ:
				handler = IndexHandlerBytesInt
				vm.executeBytesIndex(left, index)
			default:
				err := vm.executeIndexExpression(left, index)
				if err != nil {
					return err
				}
				continue
			}
			if handler >= 0 {
				vm.indexCache[cacheKey] = IndexCacheEntry{
					LeftType:  leftType,
					IndexType: indexType,
					Handler:   handler,
				}
			}
		case compiler.OpSetIndex:
			value := vm.pop()
			index := vm.pop()
			left := vm.pop()

			// 内联缓存快速路径
			cacheKey := IndexCacheKey{FrameIndex: vm.framesIndex - 1, IP: ip}
			leftType := left.Type()
			indexType := index.Type()
			if entry, hit := vm.indexCache[cacheKey]; hit {
				if entry.LeftType == leftType && entry.IndexType == indexType {
					switch entry.Handler {
					case IndexHandlerListInt:
						if lst, ok := left.(*objects.List); ok {
							if idx, ok := index.(*objects.Integer); ok {
								i := idx.Value
								length := int64(len(lst.Elements))
								if i < 0 {
									i = length + i
								}
								if i >= 0 && i < length {
									lst.Elements[i] = value
									vm.push(value)
									continue
								}
								break
							}
						}
					case IndexHandlerDict:
						if d, ok := left.(*objects.Dict); ok {
							if err := objects.CheckHashable(index); err == nil {
								d.Set(index, value)
								vm.push(value)
								continue
							}
							break
						}
					}
				}
			}

			// 缓存未命中，走正常路径并填充缓存
			handler := -1
			switch {
			case leftType == objects.LIST_OBJ && indexType == objects.INTEGER_OBJ:
				handler = IndexHandlerListInt
			case leftType == objects.DICT_OBJ:
				handler = IndexHandlerDict
			}

			err := vm.executeSetIndex(left, index, value)
			if err != nil {
				return err
			}

			if handler >= 0 {
				vm.indexCache[cacheKey] = IndexCacheEntry{
					LeftType:  leftType,
					IndexType: indexType,
					Handler:   handler,
				}
			}
		case compiler.OpSetSlice:
			value := vm.pop()
			step := vm.pop()
			upper := vm.pop()
			lower := vm.pop()
			left := vm.pop()

			err := vm.executeSetSlice(left, lower, upper, step, value)
			if err != nil {
				return err
			}

		case compiler.OpInPlaceAdd,
			compiler.OpInPlaceSub,
			compiler.OpInPlaceMul,
			compiler.OpInPlaceDiv,
			compiler.OpInPlaceMod,
			compiler.OpInPlaceFloorDiv,
			compiler.OpInPlacePower,
			compiler.OpInPlaceBitOr,
			compiler.OpInPlaceBitAnd,
			compiler.OpInPlaceBitXor:
			right := vm.pop()
			left := vm.pop()

			err := vm.executeInPlaceOperation(op, left, right)
			if err != nil {
				return err
			}

		case compiler.OpStringBuilderCreate:
			sb := objects.NewStringBuilder()
			err := vm.push(sb)
			if err != nil {
				return err
			}

		case compiler.OpStringBuilderAppend:
			value := vm.pop()
			sb := vm.stack[vm.sp-1] // StringBuilder 在栈顶
			if builder, ok := sb.(*objects.StringBuilder); ok {
				var s string
				if str, ok := value.(*objects.String); ok {
					s = str.Value
				} else {
					s = value.Inspect()
				}
				builder.Builder.WriteString(s)
			}
			// value 已消费，StringBuilder 仍在栈顶

		case compiler.OpStringBuilderBuild:
			sb := vm.pop()
			if builder, ok := sb.(*objects.StringBuilder); ok {
				err := vm.push(&objects.String{Value: builder.Builder.String()})
				if err != nil {
					return err
				}
			} else {
				// 类型不匹配，将原对象放回栈上避免栈不平衡
				err := vm.push(sb)
				if err != nil {
					return err
				}
			}

		case compiler.OpArrayPrealloc:
			// 预分配列表容量，参数为预估元素数量
			capacity := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			list := &objects.List{Elements: make([]objects.Object, 0, capacity)}
			err := vm.push(list)
			if err != nil {
				return err
			}

		case compiler.OpSlice:
			step := vm.pop()
			end := vm.pop()
			start := vm.pop()
			left := vm.pop()

			_ = step

			err := vm.executeSliceExpression(left, start, end)
			if err != nil {
				return err
			}

		case compiler.OpListUnpack:
			listObj := vm.pop()
			if list, ok := listObj.(*objects.List); ok {
				for _, elem := range list.Elements {
					err := vm.push(elem)
					if err != nil {
						return err
					}
				}
				vm.unpackExtraArgs += len(list.Elements) - 1 // -1 因为 list 本身占1个位置
			}

		case compiler.OpDictUnpack:
			// Dict 已经在栈上，executeCall 会检测最后一个参数是否是 Dict
			// 不需要额外操作

		case compiler.OpCall:
			numArgs := int(ins[ip+1])
			vm.currentFrame().ip += 1

			numArgs += vm.unpackExtraArgs
			vm.unpackExtraArgs = 0

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
			
			// Check if the caller is an async object (from OpAwait)
			// The async object was placed at basePointer-1 by OpAwait
			if frame.basePointer > 0 && frame.basePointer-1 < len(vm.stack) {
				if asyncObj, ok := vm.stack[frame.basePointer-1].(*objects.Async); ok {
					asyncObj.Done = true
					asyncObj.Result = returnValue
				}
			}
			
			if len(vm.frames) > 0 && vm.sp > 0 {
				calleeIndex := vm.sp - 1
				if calleeIndex >= 0 {
					if gen, ok := vm.stack[calleeIndex].(*objects.Generator); ok {
						gen.Done = true
					}
				}
			}
			
			vm.sp = frame.basePointer - 1

			if frame.setAttrValue != nil {
				err := vm.push(frame.setAttrValue)
				if err != nil {
					return err
				}
			} else if frame.initInstance != nil {
				err := vm.push(frame.initInstance)
				if err != nil {
					return err
				}
			} else if frame.metaclassInitClass != nil {
				err := vm.push(frame.metaclassInitClass)
				if err != nil {
					return err
				}
			} else {
				err := vm.push(returnValue)
				if err != nil {
					return err
				}
			}

		case compiler.OpReturn:
			frame := vm.popFrame()
			
			if vm.framesIndex == 0 {
				return nil
			}
			
			if len(vm.frames) > 0 && vm.sp > 0 {
				calleeIndex := vm.sp - 1
				if calleeIndex >= 0 {
					if gen, ok := vm.stack[calleeIndex].(*objects.Generator); ok {
						gen.Done = true
					}
				}
			}
			
			vm.sp = frame.basePointer - 1

			if frame.setAttrValue != nil {
				err := vm.push(frame.setAttrValue)
				if err != nil {
					return err
				}
			} else if frame.initInstance != nil {
				err := vm.push(frame.initInstance)
				if err != nil {
					return err
				}
			} else if frame.metaclassInitClass != nil {
				err := vm.push(frame.metaclassInitClass)
				if err != nil {
					return err
				}
			} else {
				err := vm.push(objects.None_)
				if err != nil {
					return err
				}
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
		val := vm.stack[frame.basePointer+localIndex]
		err := vm.push(val)
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
		handlerIP := int(uint16(ins[ip+5])<<8 | uint16(ins[ip+6]))
		finallyStartIP := int(uint16(ins[ip+7])<<8 | uint16(ins[ip+8]))
		vm.currentFrame().ip += 8
		tryBlockStartIP := ip + 9

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
		handler.baseIP = vm.currentFrame().ip + 1
		vm.exceptionStack = append(vm.exceptionStack, handler)
		case compiler.OpEndTry:
			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
				handler := vm.exceptionStack[lastIdx]

				// 如果 except 块处理了异常，我们需要把之前 push 到栈上的异常对象 pop 掉
				// 判断依据是我们当前栈是否比 handler.stackPtr 多一个值（也就是那个 errObj）
				if vm.sp > handler.stackPtr {
					_ = vm.pop()
				}

				// For except* handlers with unmatched exceptions, try the next except* handler
				if handler.isStar && len(handler.starUnhandled) > 0 {
					nextStarIP := vm.findNextExceptStarHandler(handler.exceptOpcodeIP)
					if nextStarIP >= 0 {
						ins := vm.currentFrame().fn.Instructions
						typeIdx := int(uint16(ins[nextStarIP+1])<<8 | uint16(ins[nextStarIP+2]))
						var exceptionType string
						if typeIdx > 0 && typeIdx < len(vm.constants) {
							if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
								exceptionType = typeObj.Value
							}
						}

						unhandledEG := &objects.ExceptionGroup{
							Message:    "unhandled",
							Exceptions: handler.starUnhandled,
						}
						matched, remaining := splitExceptionGroup(unhandledEG, exceptionType)

						if len(matched) > 0 {
							matchedEG := &objects.ExceptionGroup{
								Message:    unhandledEG.Message,
								Exceptions: matched,
							}
							// Update handler's starUnhandled with remaining
							vm.exceptionStack[lastIdx].starUnhandled = remaining
							// Reset handlerIP so OpExceptStarHandler dispatch can update this handler
							vm.exceptionStack[lastIdx].handlerIP = -1
							// Push matched EG onto stack
							vm.sp = handler.stackPtr
							if err := vm.push(matchedEG); err != nil {
								return err
							}
							// Jump to next OpExceptStarHandler instruction
							// -1 because main loop will increment ip
							vm.currentFrame().ip = nextStarIP - 1
							break
						}
					}

					// No more matching except* handlers, pop and re-raise
					vm.exceptionStack = vm.exceptionStack[:lastIdx]
					vm.inFinally = false
					unhandledEG := &objects.ExceptionGroup{
						Message:    "unhandled",
						Exceptions: handler.starUnhandled,
					}
					caught := vm.raiseException(unhandledEG)
					if !caught {
						vm.pendingError = unhandledEG
						return fmt.Errorf("unhandled exception: %s", unhandledEG.Inspect())
					}
					break
				}

				// Normal handling: pop handler
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
		case compiler.OpRaise:
			errObj := vm.pop()
			caught := false

			for i := len(vm.exceptionStack) - 1; i >= 0; i-- {
				handler := vm.exceptionStack[i]
				if handler.exceptCount == 0 && !handler.hasFinally {
					continue
				}

				for vm.framesIndex > handler.frameIndex {
					vm.popFrame()
				}

				if handler.hasFinally && handler.finallyStartIP > 0 && handler.exceptCount == 0 {
					vm.exceptionStack[i].pendingError = errObj
					vm.currentFrame().ip = handler.finallyStartIP - 1
					caught = true
					break
				}

				if handler.handlerIP >= 0 {
					scanIP := handler.handlerIP
					ins := vm.currentFrame().fn.Instructions
					for scanIP < len(ins) {
						op := compiler.Opcode(ins[scanIP])
						if op == compiler.OpExceptStarHandler {
							typeIdx := int(uint16(ins[scanIP+1])<<8 | uint16(ins[scanIP+2]))
							var exceptionType string
							if typeIdx > 0 && typeIdx < len(vm.constants) {
								if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
									exceptionType = typeObj.Value
								}
							}
							// except* handles ExceptionGroup
							if eg, ok := errObj.(*objects.ExceptionGroup); ok {
								matched, unmatched := splitExceptionGroup(eg, exceptionType)
								if len(matched) > 0 {
									matchedEG := &objects.ExceptionGroup{
										Message:    eg.Message,
										Exceptions: matched,
									}
									vm.sp = handler.stackPtr
									if err := vm.push(matchedEG); err != nil {
										return err
									}
									vm.currentFrame().ip = scanIP + 5 - 1
									vm.exceptionStack[i].isStar = true
									vm.exceptionStack[i].exceptOpcodeIP = scanIP
									vm.exceptionStack[i].starUnhandled = append(vm.exceptionStack[i].starUnhandled, unmatched...)
									caught = true
								}
								break
							} else if _, ok := errObj.(*objects.Error); ok {
								if exceptionType == "" || matchesException(errObj, exceptionType) {
									wrappedEG := &objects.ExceptionGroup{
										Message:    "",
										Exceptions: []objects.Object{errObj},
									}
									vm.sp = handler.stackPtr
									if err := vm.push(wrappedEG); err != nil {
										return err
									}
									vm.currentFrame().ip = scanIP + 5 - 1
									vm.exceptionStack[i].isStar = true
									vm.exceptionStack[i].exceptOpcodeIP = scanIP
									caught = true
								}
								break
							}
							break
						}
						if op == compiler.OpExceptHandler {
							typeIdx := int(uint16(ins[scanIP+1])<<8 | uint16(ins[scanIP+2]))
							var exceptionType string
							if typeIdx > 0 && typeIdx < len(vm.constants) {
								if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
									exceptionType = typeObj.Value
								}
							}
							// Typed except does NOT catch ExceptionGroup (but bare except: does)
							if _, isEG := errObj.(*objects.ExceptionGroup); isEG && exceptionType != "" {
								scanIP += 5
								continue
							}
							if exceptionType == "" || matchesException(errObj, exceptionType) {
								vm.sp = handler.stackPtr
								if err := vm.push(errObj); err != nil {
									return err
								}
								vm.currentFrame().ip = scanIP + 5 - 1
								caught = true
							}
							break
						}
						if op == compiler.OpFinally || op == compiler.OpEndTry {
							break
						}
						scanIP++
					}
				}

				if !caught && handler.hasFinally && handler.finallyStartIP > 0 {
					vm.exceptionStack[i].pendingError = errObj
					vm.currentFrame().ip = handler.finallyStartIP - 1
					caught = true
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
						exceptOpcodeIP:  ip,
					}
				}
			}
		case compiler.OpExceptStarHandler:
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
						isStar:          true,
						starUnhandled:   existingHandler.starUnhandled,
						exceptOpcodeIP:  ip,
					}
				}
			}
		case compiler.OpFinally:
			finallyEndIP := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2

			if len(vm.exceptionStack) > 0 {
				lastIdx := len(vm.exceptionStack) - 1
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
		case compiler.OpMakeAsync:
			fnObj := vm.pop()
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
				err := vm.push(async)
				if err != nil {
					return err
				}
			}
		case compiler.OpAwait:
			val := vm.pop()
			if asyncObj, ok := val.(*objects.Async); ok {
				if asyncObj.Done {
					// Already completed, return cached result
					err := vm.push(asyncObj.Result)
					if err != nil {
						return err
					}
				} else {
					// Execute the async coroutine synchronously by pushing a new frame.
					// The main Run() loop will execute this frame naturally.
					// When the async function returns, OpReturnValue will pop the frame
					// and push the result. We rely on the async function's return value
					// being the awaited result.
					asyncFn := &compiler.CompiledFunction{
						Instructions: asyncObj.Instructions,
						NumLocals:    len(asyncObj.Locals),
					}
					basePointer := vm.sp
					frame := NewFrame(asyncFn, basePointer)
					vm.pushFrame(frame)
					// Copy saved locals into the new frame's stack space
					for i, localVal := range asyncObj.Locals {
						if i < asyncFn.NumLocals {
							vm.stack[basePointer+i] = localVal
						}
					}
					vm.sp = basePointer + asyncFn.NumLocals
					// Set the frame's IP to the async object's saved IP
					frame.ip = asyncObj.IP
					// Mark the async object on the stack so OpReturnValue can detect it
					// and cache the result. We push the asyncObj reference just below the frame.
					vm.stack[basePointer-1] = asyncObj
				}
			} else {
				// Not an async object, return the value directly
				err := vm.push(val)
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

			superClass := vm.pop()
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

			err := vm.push(class)
			if err != nil {
				return err
			}
		case compiler.OpCreateClassWithMultiSuper:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			class := vm.constants[idx].(*objects.Class)
			vm.currentFrame().ip += 2

			numParents := int(ins[ip+3])
			vm.currentFrame().ip += 1

			parents := make([]*objects.Class, 0, numParents)
			for i := 0; i < numParents; i++ {
				superObj := vm.pop()
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
				if attrName == "__class__" {
					err := vm.push(instance.Class)
					if err != nil {
						return err
					}
					continue
				}
				cacheKey := AttrCacheKey{FrameIndex: vm.framesIndex - 1, IP: ip, ObjType: objects.INSTANCE_OBJ}
				if entry, hit := vm.attrCache[cacheKey]; hit {
					if entry.ClassName == instance.Class.Name {
						if entry.IsMethod {
							vm.push(instance)
							vm.push(entry.Value)
							continue
						}
						err := vm.push(entry.Value)
						if err != nil {
							return err
						}
						continue
					}
				}

				if instance.Class != nil {
					if classAttr, ok := instance.Class.FindClassAttr(attrName); ok {
						if prop, ok := classAttr.(*objects.Property); ok {
							if prop.Fget != nil && prop.Fget != objects.None_ {
								vm.push(instance)
								vm.push(prop.Fget)
								err := vm.executeCall(0)
								if err != nil {
									return err
								}
								continue
							}
						}
						if cm, ok := classAttr.(*objects.ClassMethod); ok {
							vm.push(instance.Class)
							vm.push(cm.Fn)
							continue
						}
						if sm, ok := classAttr.(*objects.StaticMethod); ok {
							err := vm.push(sm.Fn)
							if err != nil {
								return err
							}
							continue
						}
						if objects.IsDataDescriptor(classAttr) {
							if descInst, ok := classAttr.(*objects.Instance); ok {
								getMethod, getFound := descInst.GetAttr("__get__")
								if getFound {
									vm.push(descInst)
									vm.push(getMethod)
									vm.push(instance)
									classObj := &objects.String{Value: instance.Class.Name}
									vm.push(classObj)
									err := vm.executeCall(2)
									if err != nil {
										return err
									}
									continue
								}
							}
						}
					}
				}

				// 查找实例属性：优先 SlotValues，然后 Fields
				if instance.IsSlottedInstance() {
					if idx := instance.GetSlotIndex(attrName); idx >= 0 && idx < len(instance.SlotValues) {
						if instance.SlotValues[idx] != nil {
							val := instance.SlotValues[idx]
							vm.attrCache[cacheKey] = AttrCacheEntry{
								Value:     val,
								IsMethod:  false,
								ClassName: instance.Class.Name,
							}
							vm.push(val)
							continue
						} else {
							// slot 存在但值为 nil（已被 del），抛出 AttributeError
							errObj := objects.NewAttributeError("'%s' object has no attribute '%s'", instance.Class.Name, attrName)
							caught := vm.raiseException(errObj)
							if !caught {
								return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
							}
							continue
						}
					}
				}
				if val, ok := instance.Fields[attrName]; ok {
					vm.attrCache[cacheKey] = AttrCacheEntry{
						Value:     val,
						IsMethod:  false,
						ClassName: instance.Class.Name,
					}
					err := vm.push(val)
					if err != nil {
						return err
					}
					continue
				}

				if instance.Class != nil {
					if classAttr, ok := instance.Class.FindClassAttr(attrName); ok {
						if cm, ok := classAttr.(*objects.ClassMethod); ok {
							vm.push(instance.Class)
							vm.push(cm.Fn)
							continue
						}
						if sm, ok := classAttr.(*objects.StaticMethod); ok {
							err := vm.push(sm.Fn)
							if err != nil {
								return err
							}
							continue
						}
						if objects.IsDescriptor(classAttr) && !objects.IsDataDescriptor(classAttr) {
							if descInst, ok := classAttr.(*objects.Instance); ok {
								if getMethod, ok := descInst.GetAttr("__get__"); ok {
									vm.push(descInst)
									vm.push(getMethod)
									vm.push(instance)
									classObj := &objects.String{Value: instance.Class.Name}
									vm.push(classObj)
									err := vm.executeCall(2)
									if err != nil {
										return err
									}
									continue
								}
							}
						}
						if method, ok := classAttr.(*compiler.CompiledFunction); ok {
							vm.attrCache[cacheKey] = AttrCacheEntry{
								Value:     method,
								IsMethod:  true,
								ClassName: instance.Class.Name,
							}
							vm.push(instance)
							vm.push(method)
							continue
						}
						vm.attrCache[cacheKey] = AttrCacheEntry{
							Value:     classAttr,
							IsMethod:  false,
							ClassName: instance.Class.Name,
						}
						err := vm.push(classAttr)
						if err != nil {
							return err
						}
						continue
					}
				}

				err := vm.push(objects.None_)
				if err != nil {
					return err
				}
				continue
			}

			if classObj, ok := obj.(*objects.Class); ok {
				if attrName == "__name__" {
					err := vm.push(&objects.String{Value: classObj.Name})
					if err != nil {
						return err
					}
					continue
				}
				if classAttr, ok := classObj.FindClassAttr(attrName); ok {
					if cm, ok := classAttr.(*objects.ClassMethod); ok {
						vm.push(classObj)
						vm.push(cm.Fn)
						continue
					}
					if sm, ok := classAttr.(*objects.StaticMethod); ok {
						err := vm.push(sm.Fn)
						if err != nil {
							return err
						}
						continue
					}
					if prop, ok := classAttr.(*objects.Property); ok {
						if prop.Fget != nil && prop.Fget != objects.None_ {
							vm.push(prop.Fget)
							vm.push(objects.None_)
							err := vm.executeCall(1)
							if err != nil {
								return err
							}
							continue
						}
					}
					if method, ok := classAttr.(*compiler.CompiledFunction); ok {
						vm.push(classObj)
						vm.push(method)
						continue
					}
					err := vm.push(classAttr)
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

			if module, ok := obj.(*objects.Module); ok {
				cacheKey := AttrCacheKey{FrameIndex: vm.framesIndex - 1, IP: ip, ObjType: objects.MODULE_OBJ}
				if entry, hit := vm.attrCache[cacheKey]; hit {
					if entry.ClassName == module.Name {
						err := vm.push(entry.Value)
						if err != nil {
							return err
						}
						continue
					}
				}
				if val, ok := module.GetAttr(attrName); ok {
					vm.attrCache[cacheKey] = AttrCacheEntry{
						Value:     val,
						IsMethod:  false,
						ClassName: module.Name,
					}
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

			if enumObj, ok := obj.(*objects.Enum); ok {
				cacheKey := AttrCacheKey{FrameIndex: vm.framesIndex - 1, IP: ip, ObjType: objects.ENUM_OBJ}
				if entry, hit := vm.attrCache[cacheKey]; hit {
					if entry.ClassName == enumObj.Name {
						err := vm.push(entry.Value)
						if err != nil {
							return err
						}
						continue
					}
				}
				if val, ok := enumObj.GetAttr(attrName); ok {
					vm.attrCache[cacheKey] = AttrCacheEntry{
						Value:     val,
						IsMethod:  false,
						ClassName: enumObj.Name,
					}
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

			// Handle super() attribute access
			if superObj, ok := obj.(*objects.Super); ok {
				if superObj.SuperClass != nil {
					if classAttr, ok := superObj.SuperClass.FindClassAttr(attrName); ok {
						if method, ok := classAttr.(*compiler.CompiledFunction); ok {
							vm.push(superObj.Instance)
							vm.push(method)
							continue
						}
						if prop, ok := classAttr.(*objects.Property); ok {
							if prop.Fget != nil && prop.Fget != objects.None_ {
								vm.push(superObj.Instance)
								vm.push(prop.Fget)
								err := vm.executeCall(0)
								if err != nil {
									return err
								}
								continue
							}
						}
						err := vm.push(classAttr)
						if err != nil {
							return err
						}
						continue
					}
				}
				err := vm.push(objects.None_)
				if err != nil {
					return err
				}
				continue
			}

			// Handle String attribute access
			if strObj, ok := obj.(*objects.String); ok {
				if val, found := strObj.GetAttr(attrName); found {
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

			// Handle List attribute access
			if listObj, ok := obj.(*objects.List); ok {
				if val, found := listObj.GetAttr(attrName); found {
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

			// Handle Set attribute access
			if setObj, ok := obj.(*objects.Set); ok {
				if val, found := setObj.GetAttr(attrName); found {
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

			// Handle Dict attribute access (keys, values, items, etc.)
			if dictObj, ok := obj.(*objects.Dict); ok {
				switch attrName {
				case "keys":
					err := vm.push(objects.NewDictKeys(dictObj))
					if err != nil {
						return err
					}
					continue
				case "values":
					err := vm.push(objects.NewDictValues(dictObj))
					if err != nil {
						return err
					}
					continue
				case "items":
					err := vm.push(objects.NewDictItems(dictObj))
					if err != nil {
						return err
					}
					continue
				case "get":
					// Return a builtin that captures the dict
					getFn := &objects.Builtin{
						Name: "dict.get",
						Fn: func(args ...objects.Object) objects.Object {
							if len(args) < 1 || len(args) > 2 {
								return objects.NewTypeError("dict.get() takes at most 2 arguments")
							}
							val, ok := dictObj.Get(args[0])
							if ok {
								return val
							}
							if len(args) == 2 {
								return args[1]
							}
							return objects.None_
						},
					}
					err := vm.push(getFn)
					if err != nil {
						return err
					}
					continue
				case "pop":
					popFn := &objects.Builtin{
						Name: "dict.pop",
						Fn: func(args ...objects.Object) objects.Object {
							if len(args) < 1 || len(args) > 2 {
								return objects.NewTypeError("dict.pop() takes at most 2 arguments")
							}
							val, ok := dictObj.Get(args[0])
							if !ok {
								if len(args) == 2 {
									return args[1]
								}
								return objects.NewKeyError("'%s'", args[0].Inspect())
							}
							dictObj.Delete(args[0])
							return val
						},
					}
					err := vm.push(popFn)
					if err != nil {
						return err
					}
					continue
				case "setdefault":
					setdefaultFn := &objects.Builtin{
						Name: "dict.setdefault",
						Fn: func(args ...objects.Object) objects.Object {
							if len(args) < 1 || len(args) > 2 {
								return objects.NewTypeError("dict.setdefault() takes at most 2 arguments")
							}
							if err := objects.CheckHashable(args[0]); err != nil {
								return err.(*objects.Error)
							}
							val, ok := dictObj.Get(args[0])
							if ok {
								return val
							}
							var defaultVal objects.Object = objects.None_
							if len(args) == 2 {
								defaultVal = args[1]
							}
							dictObj.Set(args[0], defaultVal)
							return defaultVal
						},
					}
					err := vm.push(setdefaultFn)
					if err != nil {
						return err
					}
					continue
				case "update":
					updateFn := &objects.Builtin{
						Name: "dict.update",
						Fn: func(args ...objects.Object) objects.Object {
							if len(args) < 1 {
								return objects.NewTypeError("dict.update() takes at least 1 argument")
							}
							if other, ok := args[0].(*objects.Dict); ok {
								for _, keyStr := range other.KeyOrder {
									dictObj.Set(other.Keys[keyStr], other.Pairs[keyStr])
								}
							}
							return objects.None_
						},
					}
					err := vm.push(updateFn)
					if err != nil {
						return err
					}
					continue
				case "clear":
					clearFn := &objects.Builtin{
						Name: "dict.clear",
						Fn: func(args ...objects.Object) objects.Object {
							dictObj.Pairs = make(map[string]objects.Object)
							dictObj.Keys = make(map[string]objects.Object)
							dictObj.KeyOrder = make([]string, 0)
							return objects.None_
						},
					}
					err := vm.push(clearFn)
					if err != nil {
						return err
					}
					continue
				case "copy":
					copyFn := &objects.Builtin{
						Name: "dict.copy",
						Fn: func(args ...objects.Object) objects.Object {
							newDict := objects.NewDict()
							for _, keyStr := range dictObj.KeyOrder {
								newDict.Set(dictObj.Keys[keyStr], dictObj.Pairs[keyStr])
							}
							return newDict
						},
					}
					err := vm.push(copyFn)
					if err != nil {
						return err
					}
					continue
				}
				// Fallback to Dict's GetAttr method (e.g., fromkeys)
				if val, found := dictObj.GetAttr(attrName); found {
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

			// Handle RegexPattern attribute access
			if regexPattern, ok := obj.(*objects.RegexPattern); ok {
				if val, found := regexPattern.GetAttr(attrName); found {
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

			// Handle RegexMatch attribute access
			if regexMatch, ok := obj.(*objects.RegexMatch); ok {
				if val, found := regexMatch.GetAttr(attrName); found {
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

			// Fallback: check if object implements GetAttr method (for custom objects like io.StringIO)
			if getter, ok := obj.(interface{ GetAttr(string) (objects.Object, bool) }); ok {
				if val, found := getter.GetAttr(attrName); found {
					err := vm.push(val)
					if err != nil {
						return err
					}
					continue
				}
			}

			return fmt.Errorf("cannot get attribute on non-instance: %s", obj.Type())

		case compiler.OpSetAttribute:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			attrName := vm.constants[idx].(*objects.String).Value
			vm.currentFrame().ip += 2
			
			value := vm.pop()
			obj := vm.pop()
			
			if instance, ok := obj.(*objects.Instance); ok {
				if instance.Class != nil {
					classAttr, found := instance.Class.FindClassAttr(attrName)
					if found {
						if prop, ok := classAttr.(*objects.Property); ok {
							if prop.Fset != nil && prop.Fset != objects.None_ {
								vm.push(instance)
								vm.push(prop.Fset)
								vm.push(value)
								err := vm.executeCall(1)
								if err != nil {
									return err
								}
								vm.currentFrame().setAttrValue = value
								continue
							} else {
								errObj := objects.NewAttributeError("can't set attribute '%s'", attrName)
								caught := vm.raiseException(errObj)
								if !caught {
									return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
								}
								continue
							}
						}
						if objects.IsDataDescriptor(classAttr) {
							if descInst, ok := classAttr.(*objects.Instance); ok {
								setMethod, setFound := descInst.GetAttr("__set__")
								if setFound {
									vm.push(descInst)
									vm.push(setMethod)
									vm.push(instance)
									vm.push(value)
									err := vm.executeCall(2)
									if err != nil {
										return err
									}
									vm.currentFrame().setAttrValue = value
									continue
								}
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
					continue
				}
				instance.SetAttr(attrName, value)
				err := vm.push(value)
				if err != nil {
					return err
				}
				continue
			}
			
			if classObj, ok := obj.(*objects.Class); ok {
				classObj.Fields[attrName] = value
				if attrName == "__slots__" {
					if listObj, ok := value.(*objects.List); ok {
						slots := make([]string, 0, len(listObj.Elements))
						for _, elem := range listObj.Elements {
							if strElem, ok := elem.(*objects.String); ok {
								slots = append(slots, strElem.Value)
							}
						}
						classObj.Slots = slots
					}
				}
				err := vm.push(value)
				if err != nil {
					return err
				}
				continue
			}

			return fmt.Errorf("cannot set attribute on non-instance: %s", obj.Type())

		case compiler.OpDelAttribute:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			attrName := vm.constants[idx].(*objects.String).Value
			vm.currentFrame().ip += 2

			obj := vm.pop()

			if instance, ok := obj.(*objects.Instance); ok {
				if instance.Class != nil {
					classAttr, found := instance.Class.FindClassAttr(attrName)
					if found {
						if prop, ok := classAttr.(*objects.Property); ok {
							if prop.Fdel != nil && prop.Fdel != objects.None_ {
								vm.push(instance)
								vm.push(prop.Fdel)
								err := vm.executeCall(0)
								if err != nil {
									return err
								}
								vm.pop() // pop the return value of deleter
								continue
							} else {
								errObj := objects.NewAttributeError("can't delete attribute '%s'", attrName)
								caught := vm.raiseException(errObj)
								if !caught {
									return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
								}
								continue
							}
						}
						if objects.IsDataDescriptor(classAttr) {
							if descInst, ok := classAttr.(*objects.Instance); ok {
								delMethod, delFound := descInst.GetAttr("__delete__")
								if delFound {
									vm.push(descInst)
									vm.push(delMethod)
									vm.push(instance)
									err := vm.executeCall(1)
									if err != nil {
										return err
									}
									vm.pop() // pop return value
									continue
								}
							}
						}
					}
				}
				vm.attrCache = make(map[AttrCacheKey]AttrCacheEntry)
				if instance.IsSlottedInstance() {
					idx := instance.GetSlotIndex(attrName)
					if idx >= 0 && idx < len(instance.SlotValues) {
						instance.SlotValues[idx] = nil
					} else {
						errObj := objects.NewAttributeError("'%s' object has no attribute '%s'", instance.Class.Name, attrName)
						caught := vm.raiseException(errObj)
						if !caught {
							return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
						}
					}
				} else {
					delete(instance.Fields, attrName)
				}
				continue
			}

			return fmt.Errorf("cannot delete attribute on non-instance: %s", obj.Type())
		case compiler.OpSetClassField:
			idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			vm.currentFrame().ip += 2
			fieldName := vm.constants[idx].(*objects.String).Value

			value := vm.pop()
			classObj := vm.pop()

			if classInst, ok := classObj.(*objects.Class); ok {
				delete(classInst.Methods, fieldName)
				classInst.Fields[fieldName] = value
				if fieldName == "__slots__" {
					if listObj, ok := value.(*objects.List); ok {
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
		case compiler.OpSetMetaclass:
			metaclassObj := vm.pop()
			classObj := vm.pop()
			if classInst, ok := classObj.(*objects.Class); ok {
				if metaclass, ok := metaclassObj.(*objects.Class); ok {
					classInst.Metaclass = metaclass
				}
			}
			vm.push(classObj)
		case compiler.OpCallMetaclassInit:
			// Stack top: class object
			// If class has a metaclass with __init__, call it
			classObj := vm.pop()
			if classInst, ok := classObj.(*objects.Class); ok {
				if classInst.Metaclass != nil {
					if initMethod, ok := classInst.Metaclass.Methods["__init__"]; ok {
						if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
							// metaclass.__init__(self, cls, name, bases)
							vm.push(nil) // placeholder for callee slot
							vm.push(classInst.Metaclass) // self
							vm.push(classInst)           // cls
							vm.push(&objects.String{Value: classInst.Name}) // name
							// bases as list
							bases := &objects.List{Elements: make([]objects.Object, len(classInst.SuperClasses))}
							for i, sc := range classInst.SuperClasses {
								bases.Elements[i] = sc
							}
							vm.push(bases)
							bp := vm.sp - 4
							frame := NewFrame(fn, bp)
							frame.metaclassInitClass = classInst
							vm.pushFrame(frame)
							vm.sp = frame.basePointer + fn.NumLocals
							continue
						}
					}
				}
				vm.push(classObj)
			} else {
				vm.push(classObj)
			}
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

	// First, check if it's a Callable type
	if callable, ok := calleeObj.(objects.Callable); ok {
		args := vm.stack[vm.sp-numArgs : vm.sp]
		result := callable.Call(args...)
		vm.sp = vm.sp - numArgs - 1
		return vm.push(result)
	}

	// Next, check if it's an Instance with __call__ method
	if inst, ok := calleeObj.(*objects.Instance); ok {
		if callMethod, ok := inst.GetAttr("__call__"); ok {
			// Push the instance as self, then the args
			args := make([]objects.Object, numArgs+1)
			args[0] = inst
			copy(args[1:], vm.stack[vm.sp-numArgs:vm.sp])
			// Now set up the call
			vm.stack[calleeIndex] = callMethod
			copy(vm.stack[calleeIndex+1:], args)
			vm.sp = calleeIndex + 1 + numArgs
			return vm.executeCall(numArgs + 1)
		}
	}

	if classObj, ok := calleeObj.(*objects.Class); ok {
		// Check if the class has a metaclass with __call__
		if classObj.Metaclass != nil {
			if callMethod, ok := classObj.Metaclass.Methods["__call__"]; ok {
				if fn, ok := callMethod.(*compiler.CompiledFunction); ok {
					// metaclass.__call__(self, cls, *args)
					// self = metaclass instance, cls = class being instantiated
					args := make([]objects.Object, numArgs)
					for i := 0; i < numArgs; i++ {
						args[i] = vm.stack[vm.sp-numArgs+i]
					}
					vm.stack[calleeIndex] = nil                 // placeholder
					vm.stack[calleeIndex+1] = classObj.Metaclass // self
					vm.stack[calleeIndex+2] = classObj           // cls
					for i := 0; i < numArgs; i++ {
						vm.stack[calleeIndex+3+i] = args[i]
					}
					vm.sp = calleeIndex + 3 + numArgs
					basePointer := calleeIndex + 1
					frame := NewFrame(fn, basePointer)
					vm.pushFrame(frame)
					vm.sp = frame.basePointer + fn.NumLocals
					return nil
				}
			}
		}

		instance := &objects.Instance{
			Class: classObj,
		}
		// 当类有 __slots__ 时，使用固定数组替代 map，节省内存
		if classObj.HasSlots() {
			allSlots := classObj.AllSlotNames()
			instance.SlotValues = make([]objects.Object, len(allSlots))
		} else {
			instance.Fields = make(map[string]objects.Object)
		}

		if initMethod, ok := classObj.Methods["__init__"]; ok {
			if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = vm.stack[vm.sp-numArgs+i]
				}
				vm.stack[calleeIndex] = instance
				vm.stack[calleeIndex+1] = instance
				for i := 0; i < numArgs; i++ {
					vm.stack[calleeIndex+2+i] = args[i]
				}
				vm.sp = calleeIndex + 2 + numArgs
				basePointer := calleeIndex + 1
				frame := NewFrame(fn, basePointer)
				frame.initInstance = instance
				vm.pushFrame(frame)
				vm.sp = frame.basePointer + fn.NumLocals
				return nil
			}
		}

		for i := 0; i < numArgs; i++ {
			instance.Fields[fmt.Sprintf("arg%d", i)] = vm.stack[vm.sp-numArgs+i]
		}

		vm.stack[calleeIndex] = instance
		vm.sp = calleeIndex + 1

		return nil
	}

	// Handle super() call
	if builtin, ok := calleeObj.(*objects.Builtin); ok && builtin.Name == "super" {
		// Find self (the current instance) from the caller's frame
		var currentInstance *objects.Instance
		var currentClass *objects.Class
		// Search all frames for the nearest Instance
		for fi := vm.framesIndex - 2; fi >= 0; fi-- {
			frame := vm.frames[fi]
			for i := frame.basePointer; i < frame.basePointer+frame.fn.NumLocals+frame.fn.NumParameters+5 && i < vm.sp; i++ {
				if inst, ok := vm.stack[i].(*objects.Instance); ok {
					currentInstance = inst
					currentClass = inst.Class
					break
				}
			}
			if currentInstance != nil {
				break
			}
		}
		if currentInstance == nil {
			vm.sp = calleeIndex + 1
			vm.stack[calleeIndex] = objects.None_
			return nil
		}
		superObj := &objects.Super{
			Instance:   currentInstance,
			SuperClass: currentClass.SuperClass,
		}
		vm.sp = calleeIndex + 1
		vm.stack[calleeIndex] = superObj
		return nil
	}

	if calleeIndex > 0 && numArgs == 0 {
		if _, isMethod := calleeObj.(*compiler.CompiledFunction); isMethod {
			if instance, isInstance := vm.stack[calleeIndex-1].(*objects.Instance); isInstance {
				vm.stack[calleeIndex-1] = calleeObj
				vm.stack[calleeIndex] = instance
				numArgs = 1
			} else if cls, isClass := vm.stack[calleeIndex-1].(*objects.Class); isClass {
				vm.stack[calleeIndex-1] = calleeObj
				vm.stack[calleeIndex] = cls
				numArgs = 1
			}
		}
	}

	if calleeIndex > 0 && numArgs > 0 {
		if _, isMethod := calleeObj.(*compiler.CompiledFunction); isMethod {
			if instance, isInstance := vm.stack[calleeIndex-1].(*objects.Instance); isInstance {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = vm.stack[vm.sp-numArgs+i]
				}
				vm.stack[calleeIndex-1] = calleeObj
				vm.stack[calleeIndex] = instance
				for i := 0; i < numArgs; i++ {
					vm.stack[calleeIndex+1+i] = args[i]
				}
				vm.sp = calleeIndex + 1 + numArgs
				numArgs = numArgs + 1
			} else if cls, isClass := vm.stack[calleeIndex-1].(*objects.Class); isClass {
				args := make([]objects.Object, numArgs)
				for i := 0; i < numArgs; i++ {
					args[i] = vm.stack[vm.sp-numArgs+i]
				}
				vm.stack[calleeIndex-1] = calleeObj
				vm.stack[calleeIndex] = cls
				for i := 0; i < numArgs; i++ {
					vm.stack[calleeIndex+1+i] = args[i]
				}
				vm.sp = calleeIndex + 1 + numArgs
				numArgs = numArgs + 1
			}
		}
	}

	if gen, ok := calleeObj.(*objects.Generator); ok {
			if gen.Done {
				vm.sp = vm.sp - numArgs - 1
				return vm.push(objects.NewErrorWithType("StopIteration", "generator is exhausted"))
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
		posArgsCount := numArgs
		var kwargsDict *objects.Dict = nil

		if numArgs > 0 {
			lastArgIdx := vm.sp - 1
			if dict, ok := vm.stack[lastArgIdx].(*objects.Dict); ok {
				kwargsDict = dict
				posArgsCount = numArgs - 1
			}
		}

		if closure.VarArgs || closure.KwArgs {
			minParams := closure.NumParameters
			if closure.VarArgs {
				minParams -= 1
			}
			if closure.KwArgs {
				minParams -= 1
			}

			if posArgsCount < minParams {
				return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
					minParams, posArgsCount)
			}

			if closure.VarArgs {
				if posArgsCount > minParams {
					extraArgs := posArgsCount - minParams
					args := make([]objects.Object, extraArgs)
					for i := 0; i < extraArgs; i++ {
						args[i] = vm.stack[vm.sp-numArgs+minParams+i]
					}
					tuple := &objects.List{Elements: args}
					vm.stack[vm.sp-numArgs+minParams] = tuple
					vm.sp = vm.sp - extraArgs + 1 // +1 因为 tuple 占1个位置替代了 extraArgs 个元素
					numArgs = minParams + 1
					posArgsCount = numArgs
				} else if posArgsCount == minParams {
					vm.push(&objects.List{Elements: []objects.Object{}})
					numArgs = minParams + 1
					posArgsCount = numArgs
				}
			}

			if closure.KwArgs {
				if kwargsDict == nil {
					kwargsDict = objects.NewDict()
				}
				vm.stack[vm.sp-numArgs+posArgsCount] = kwargsDict
				vm.sp++
				numArgs = posArgsCount + 1
			}
		} else if closure.NumKeywordOnly > 0 || closure.NumDefaults > 0 || closure.NumPositionalOnly > 0 {
			maxPosArgs := closure.NumParameters - closure.NumKeywordOnly
			minPosArgs := maxPosArgs - closure.NumPositionalDefaults

			if kwargsDict != nil {
				// 先检查 positional-only 参数不能作为关键字传递
				for _, hashKey := range kwargsDict.KeyOrder {
					keyObj, ok := kwargsDict.Keys[hashKey]
					if !ok {
						continue
					}
					keyStr, ok := keyObj.(*objects.String)
					if !ok {
						continue
					}
					for i, paramName := range closure.ParameterNames {
						if paramName == keyStr.Value && i < len(closure.PositionalOnly) && closure.PositionalOnly[i] {
							errObj := objects.NewTypeError("function() got some positional-only arguments passed as keyword arguments: '%s'",
								keyStr.Value)
							caught := vm.raiseException(errObj)
							if !caught {
								return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
							}
							return nil
						}
					}
				}
				// 计算 kwargs 中有多少非 keyword-only 参数（可作为位置参数的补充）
				kwargsPosCount := 0
				kwOnlyStart := closure.NumParameters - closure.NumKeywordOnly
				for _, hashKey := range kwargsDict.KeyOrder {
					keyObj, ok := kwargsDict.Keys[hashKey]
					if !ok {
						continue
					}
					keyStr, ok := keyObj.(*objects.String)
					if !ok {
						continue
					}
					for i, paramName := range closure.ParameterNames {
						if paramName == keyStr.Value && i < kwOnlyStart && (i >= len(closure.PositionalOnly) || !closure.PositionalOnly[i]) {
							kwargsPosCount++
						}
					}
				}
				effectivePosArgs := posArgsCount + kwargsPosCount
				if effectivePosArgs < minPosArgs {
					return fmt.Errorf("wrong number of arguments: want=%d to %d, got=%d",
						minPosArgs, maxPosArgs, posArgsCount)
				}
				if posArgsCount > maxPosArgs {
					return fmt.Errorf("takes %d positional arguments but %d were given",
						maxPosArgs, posArgsCount)
				}

				vm.sp--
				for i := posArgsCount; i < closure.NumParameters; i++ {
					if i < maxPosArgs {
						paramName := closure.ParameterNames[i]
						key := &objects.String{Value: paramName}
						if val, ok := kwargsDict.Get(key); ok {
							vm.push(val)
						} else {
							vm.push(objects.None_)
						}
					} else {
						paramName := closure.ParameterNames[i]
						key := &objects.String{Value: paramName}
						if val, ok := kwargsDict.Get(key); ok {
							vm.push(val)
						} else {
							vm.push(objects.None_)
						}
					}
				}
				numArgs = closure.NumParameters
			} else {
				for i := posArgsCount; i < closure.NumParameters; i++ {
					vm.push(objects.None_)
				}
				numArgs = closure.NumParameters
			}
		} else if numArgs != closure.NumParameters {
			// 如果有 kwargs dict 但函数不接受 kwargs，检查 posArgsCount
			if kwargsDict != nil && posArgsCount == closure.NumParameters {
				// kwargs dict 是空的或包含不需要的关键字参数，移除它
				vm.sp--
				numArgs = posArgsCount
			} else if kwargsDict != nil && posArgsCount != closure.NumParameters {
				return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
					closure.NumParameters, posArgsCount)
			} else {
				return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
					closure.NumParameters, numArgs)
			}
		}

		fn := &compiler.CompiledFunction{
			Instructions:          closure.Instructions,
			NumLocals:             closure.NumLocals,
			NumParameters:         closure.NumParameters,
			NumKeywordOnly:        closure.NumKeywordOnly,
			NumPositionalOnly:     closure.NumPositionalOnly,
			NumDefaults:           closure.NumDefaults,
			NumPositionalDefaults: closure.NumPositionalDefaults,
			ParameterNames:        closure.ParameterNames,
			PositionalOnly:        closure.PositionalOnly,
			IsGenerator:           closure.IsGenerator,
			VarArgs:               closure.VarArgs,
			KwArgs:                closure.KwArgs,
		}

		// JIT 热点检测：记录闭包调用
		vm.jitEngine.RecordCall(fn)

		freeVarsCopy := make([]objects.Object, len(closure.Free))
		copy(freeVarsCopy, closure.Free)

		basePointer := calleeIndex + 1
		frame := NewFrameWithFreeVars(fn, basePointer, freeVarsCopy)
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

	// JIT 热点检测：记录函数调用
	vm.jitEngine.RecordCall(callee)

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

	var basePointer int
	var kwargsDict *objects.Dict = nil
	posArgsCount := numArgs

	if numArgs > 0 {
		lastArgIdx := vm.sp - 1
		if dict, ok := vm.stack[lastArgIdx].(*objects.Dict); ok {
			kwargsDict = dict
			posArgsCount = numArgs - 1
		}
	}

	if callee.VarArgs || callee.KwArgs {
		minParams := callee.NumParameters
		if callee.VarArgs {
			minParams -= 1
		}
		if callee.KwArgs {
			minParams -= 1
		}

		if posArgsCount < minParams {
			return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
				minParams, posArgsCount)
		}

		if callee.VarArgs {
			if posArgsCount > minParams {
				extraArgs := posArgsCount - minParams
				args := make([]objects.Object, extraArgs)
				for i := 0; i < extraArgs; i++ {
					args[i] = vm.stack[vm.sp-numArgs+minParams+i]
				}
				tuple := &objects.List{Elements: args}
				vm.stack[vm.sp-numArgs+minParams] = tuple
				vm.sp = vm.sp - extraArgs + 1 // +1 因为 tuple 占1个位置替代了 extraArgs 个元素
				numArgs = minParams + 1
				posArgsCount = numArgs
			} else if posArgsCount == minParams {
				vm.push(&objects.List{Elements: []objects.Object{}})
				numArgs = minParams + 1
				posArgsCount = numArgs
			}
		}

		if callee.KwArgs {
			if kwargsDict == nil {
				kwargsDict = objects.NewDict()
			}
			vm.stack[vm.sp-numArgs+posArgsCount] = kwargsDict
			numArgs = posArgsCount + 1
		}

		basePointer = calleeIndex + 1
	} else if callee.NumKeywordOnly > 0 || callee.NumDefaults > 0 || callee.NumPositionalOnly > 0 {
		maxPosArgs := callee.NumParameters - callee.NumKeywordOnly
		minPosArgs := maxPosArgs - callee.NumPositionalDefaults

		if kwargsDict != nil {
			// 先检查 positional-only 参数不能作为关键字传递
			for _, hashKey := range kwargsDict.KeyOrder {
				// 从 Keys map 中获取原始键对象
				keyObj, ok := kwargsDict.Keys[hashKey]
				if !ok {
					continue
				}
				keyStr, ok := keyObj.(*objects.String)
				if !ok {
					continue
				}
				for i, paramName := range callee.ParameterNames {
					if paramName == keyStr.Value && i < len(callee.PositionalOnly) && callee.PositionalOnly[i] {
						funcName := callee.Name
						if funcName == "" {
							funcName = "function"
						}
						errObj := objects.NewTypeError("%s() got some positional-only arguments passed as keyword arguments: '%s'",
							funcName, keyStr.Value)
						caught := vm.raiseException(errObj)
						if !caught {
							return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
						}
						return nil
					}
				}
			}
			// 计算 kwargs 中有多少非 keyword-only 参数（可作为位置参数的补充）
			kwargsPosCount := 0
			kwOnlyStart := callee.NumParameters - callee.NumKeywordOnly
			for _, hashKey := range kwargsDict.KeyOrder {
				keyObj, ok := kwargsDict.Keys[hashKey]
				if !ok {
					continue
				}
				keyStr, ok := keyObj.(*objects.String)
				if !ok {
					continue
				}
				for i, paramName := range callee.ParameterNames {
					if paramName == keyStr.Value && i < kwOnlyStart && (i >= len(callee.PositionalOnly) || !callee.PositionalOnly[i]) {
						kwargsPosCount++
					}
				}
			}
			effectivePosArgs := posArgsCount + kwargsPosCount
			if effectivePosArgs < minPosArgs {
				return fmt.Errorf("wrong number of arguments: want=%d to %d, got=%d",
					minPosArgs, maxPosArgs, posArgsCount)
			}
			if posArgsCount > maxPosArgs {
				return fmt.Errorf("takes %d positional arguments but %d were given",
					maxPosArgs, posArgsCount)
			}

			vm.sp--
			for i := posArgsCount; i < callee.NumParameters; i++ {
				if i < maxPosArgs {
					paramName := callee.ParameterNames[i]
					key := &objects.String{Value: paramName}
					if val, ok := kwargsDict.Get(key); ok {
						vm.push(val)
					} else {
						vm.push(objects.None_)
					}
				} else {
					paramName := callee.ParameterNames[i]
					key := &objects.String{Value: paramName}
					if val, ok := kwargsDict.Get(key); ok {
						vm.push(val)
					} else {
						vm.push(objects.None_)
					}
				}
			}
			numArgs = callee.NumParameters
		} else {
			for i := posArgsCount; i < callee.NumParameters; i++ {
				vm.push(objects.None_)
			}
			numArgs = callee.NumParameters
		}

		basePointer = vm.sp - numArgs
	} else if callee.NumDefaults > 0 {
		minArgs := callee.NumParameters - callee.NumDefaults
		if posArgsCount < minArgs || posArgsCount > callee.NumParameters {
			return fmt.Errorf("wrong number of arguments: want=%d to %d, got=%d",
				minArgs, callee.NumParameters, posArgsCount)
		}
		for i := posArgsCount; i < callee.NumParameters; i++ {
			vm.push(objects.None_)
		}
		numArgs = callee.NumParameters
		basePointer = vm.sp - numArgs
	} else {
		if kwargsDict != nil && posArgsCount == callee.NumParameters {
			// 函数不接受 kwargs，但传入了空的 kwargs dict，移除它
			vm.sp--
			numArgs = posArgsCount
		} else if kwargsDict != nil && posArgsCount != callee.NumParameters {
			return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
				callee.NumParameters, posArgsCount)
		} else if numArgs != callee.NumParameters {
			return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
				callee.NumParameters, numArgs)
		}
		basePointer = vm.sp - numArgs
	}

	frame := NewFrame(callee, basePointer)
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
	case *objects.Bytes:
		return len(obj.Value) > 0
	case *objects.Complex:
		return obj.Real != 0 || obj.Imag != 0
	default:
		return true
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	switch op := operand.(type) {
	case *objects.Integer:
		return vm.push(objects.GetCachedInteger(-op.Value))
	case *objects.Float:
		return vm.push(&objects.Float{Value: -op.Value})
	case *objects.Complex:
		return vm.push(objects.NewComplex(-op.Real, -op.Imag))
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
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

	if leftEm, ok := left.(*objects.EnumMember); ok {
		left = leftEm.Value
	}
	if rightEm, ok := right.(*objects.EnumMember); ok {
		right = rightEm.Value
	}

	if left.Type() == objects.INTEGER_OBJ && right.Type() == objects.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == objects.FLOAT_OBJ && right.Type() == objects.FLOAT_OBJ {
		return vm.executeFloatComparison(op, left, right)
	}

	if left.Type() == objects.COMPLEX_OBJ && right.Type() == objects.COMPLEX_OBJ {
		return vm.executeComplexComparison(op, left, right)
	}

	if left.Type() == objects.BYTES_OBJ && right.Type() == objects.BYTES_OBJ {
		return vm.executeBytesComparison(op, left, right)
	}

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(objects.Equal(left, right)))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(!objects.Equal(left, right)))
	case compiler.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() > right.Inspect()))
	case compiler.OpLessThan:
		return vm.push(nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() < right.Inspect()))
	case compiler.OpGreaterEqual:
		return vm.push(nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() >= right.Inspect()))
	case compiler.OpLessEqual:
		return vm.push(nativeBoolToBooleanObject(left.Type() == right.Type() && left.Inspect() <= right.Inspect()))
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
	case compiler.OpGreaterEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue >= rightValue))
	case compiler.OpLessEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue <= rightValue))
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
	case compiler.OpGreaterEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue >= rightValue))
	case compiler.OpLessEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue <= rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeContainsOp(negate bool) error {
	container := vm.pop()
	element := vm.pop()

	result := vm.containsCheck(element, container)
	if negate {
		result = !result
	}
	return vm.push(nativeBoolToBooleanObject(result))
}

func (vm *VM) containsCheck(element, container objects.Object) bool {
	switch c := container.(type) {
	case *objects.List:
		for _, item := range c.Elements {
			if objects.Equal(element, item) {
				return true
			}
		}
		return false
	case *objects.Tuple:
		for _, item := range c.Elements {
			if objects.Equal(element, item) {
				return true
			}
		}
		return false
	case *objects.Set:
		for _, item := range c.Elements {
			if objects.Equal(element, item) {
				return true
			}
		}
		return false
	case *objects.Dict:
		keyStr := element.Inspect()
		if _, ok := c.Pairs[keyStr]; ok {
			return true
		}
		return false
	case *objects.String:
		if str, ok := element.(*objects.String); ok {
			for i := 0; i <= len(c.Value)-len(str.Value); i++ {
				if c.Value[i:i+len(str.Value)] == str.Value {
					return true
				}
			}
			return false
		}
		return false
	default:
		return false
	}
}

func (vm *VM) executeComplexComparison(op compiler.Opcode, left, right objects.Object) error {
	leftValue := left.(*objects.Complex)
	rightValue := right.(*objects.Complex)

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue.Real == rightValue.Real && leftValue.Imag == rightValue.Imag))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue.Real != rightValue.Real || leftValue.Imag != rightValue.Imag))
	default:
		return fmt.Errorf("'%s' not supported between instances of 'complex' and 'complex'", opToString(op))
	}
}

func opToString(op compiler.Opcode) string {
	switch op {
	case compiler.OpGreaterThan:
		return ">"
	case compiler.OpLessThan:
		return "<"
	case compiler.OpGreaterEqual:
		return ">="
	case compiler.OpLessEqual:
		return "<="
	default:
		return "unknown"
	}
}

func (vm *VM) executeBytesComparison(op compiler.Opcode, left, right objects.Object) error {
	leftBytes := left.(*objects.Bytes).Value
	rightBytes := right.(*objects.Bytes).Value

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(objects.Equal(left, right)))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(!objects.Equal(left, right)))
	case compiler.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(string(leftBytes) > string(rightBytes)))
	case compiler.OpLessThan:
		return vm.push(nativeBoolToBooleanObject(string(leftBytes) < string(rightBytes)))
	default:
		return fmt.Errorf("unknown operator for bytes: %d", op)
	}
}

func nativeBoolToBooleanObject(input bool) *objects.Boolean {
	if input {
		return objects.True
	}
	return objects.False
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func (vm *VM) executeBinaryOperation(op compiler.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	// Dict merge: d | other creates a new dict; d |= other merges in-place
	if leftType == objects.DICT_OBJ && rightType == objects.DICT_OBJ {
		leftDict := left.(*objects.Dict)
		rightDict := right.(*objects.Dict)
		switch op {
		case compiler.OpBitOr:
			result := objects.NewDict()
			for _, keyStr := range leftDict.KeyOrder {
				key := leftDict.Keys[keyStr]
				val := leftDict.Pairs[keyStr]
				result.Set(key, val)
			}
			for _, keyStr := range rightDict.KeyOrder {
				key := rightDict.Keys[keyStr]
				val := rightDict.Pairs[keyStr]
				result.Set(key, val)
			}
			return vm.push(result)
		default:
			return vm.push(objects.NewTypeError("unsupported operand type(s) for binary operation: '%s' and '%s'", leftType, rightType))
		}
	}

	// Set operations: | (union), & (intersection), - (difference), ^ (symmetric difference)
	if leftType == objects.SET_OBJ && rightType == objects.SET_OBJ {
		leftSet := left.(*objects.Set)
		rightSet := right.(*objects.Set)
		var result *objects.Set
		switch op {
		case compiler.OpBitOr:
			result = leftSet.Union(rightSet)
		case compiler.OpBitAnd:
			result = leftSet.Intersection(rightSet)
		case compiler.OpSub:
			result = leftSet.Difference(rightSet)
		case compiler.OpBitXor:
			result = leftSet.SymmetricDifference(rightSet)
		default:
			return vm.push(objects.NewTypeError("unsupported operand type(s) for binary operation: '%s' and '%s'", leftType, rightType))
		}
		return vm.push(result)
	}

	if leftType == objects.INTEGER_OBJ && rightType == objects.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	// Boolean arithmetic: treat True as 1, False as 0
	if leftType == objects.BOOLEAN_OBJ && rightType == objects.BOOLEAN_OBJ {
		leftInt := &objects.Integer{Value: boolToInt(left.(*objects.Boolean).Value)}
		rightInt := &objects.Integer{Value: boolToInt(right.(*objects.Boolean).Value)}
		return vm.executeBinaryIntegerOperation(op, leftInt, rightInt)
	}
	if leftType == objects.BOOLEAN_OBJ && rightType == objects.INTEGER_OBJ {
		leftInt := &objects.Integer{Value: boolToInt(left.(*objects.Boolean).Value)}
		return vm.executeBinaryIntegerOperation(op, leftInt, right)
	}
	if leftType == objects.INTEGER_OBJ && rightType == objects.BOOLEAN_OBJ {
		rightInt := &objects.Integer{Value: boolToInt(right.(*objects.Boolean).Value)}
		return vm.executeBinaryIntegerOperation(op, left, rightInt)
	}
	if leftType == objects.BOOLEAN_OBJ && rightType == objects.FLOAT_OBJ {
		leftFloat := &objects.Float{Value: boolToFloat(left.(*objects.Boolean).Value)}
		return vm.executeBinaryFloatOperation(op, leftFloat, right)
	}
	if leftType == objects.FLOAT_OBJ && rightType == objects.BOOLEAN_OBJ {
		rightFloat := &objects.Float{Value: boolToFloat(right.(*objects.Boolean).Value)}
		return vm.executeBinaryFloatOperation(op, left, rightFloat)
	}

	if leftType == objects.FLOAT_OBJ && rightType == objects.FLOAT_OBJ {
		return vm.executeBinaryFloatOperation(op, left, right)
	}

	if leftType == objects.INTEGER_OBJ && rightType == objects.FLOAT_OBJ {
		return vm.executeBinaryFloatOperation(op, &objects.Float{Value: float64(left.(*objects.Integer).Value)}, right)
	}

	if leftType == objects.FLOAT_OBJ && rightType == objects.INTEGER_OBJ {
		return vm.executeBinaryFloatOperation(op, left, &objects.Float{Value: float64(right.(*objects.Integer).Value)})
	}

	// Complex number operations
	if leftType == objects.COMPLEX_OBJ && rightType == objects.COMPLEX_OBJ {
		return vm.executeBinaryComplexOperation(op, left, right)
	}
	if leftType == objects.COMPLEX_OBJ && rightType == objects.FLOAT_OBJ {
		return vm.executeBinaryComplexOperation(op, left, objects.NewComplex(right.(*objects.Float).Value, 0))
	}
	if leftType == objects.COMPLEX_OBJ && rightType == objects.INTEGER_OBJ {
		return vm.executeBinaryComplexOperation(op, left, objects.NewComplex(float64(right.(*objects.Integer).Value), 0))
	}
	if leftType == objects.FLOAT_OBJ && rightType == objects.COMPLEX_OBJ {
		return vm.executeBinaryComplexOperation(op, objects.NewComplex(left.(*objects.Float).Value, 0), right)
	}
	if leftType == objects.INTEGER_OBJ && rightType == objects.COMPLEX_OBJ {
		return vm.executeBinaryComplexOperation(op, objects.NewComplex(float64(left.(*objects.Integer).Value), 0), right)
	}

	// String repetition: "abc" * 3 or 3 * "abc"
	if op == compiler.OpMul {
		if leftType == objects.STRING_OBJ && rightType == objects.INTEGER_OBJ {
			s := left.(*objects.String).Value
			n := right.(*objects.Integer).Value
			if n <= 0 {
				return vm.push(&objects.String{Value: ""})
			}
			result := strings.Repeat(s, int(n))
			return vm.push(&objects.String{Value: result})
		}
		if leftType == objects.INTEGER_OBJ && rightType == objects.STRING_OBJ {
			n := left.(*objects.Integer).Value
			s := right.(*objects.String).Value
			if n <= 0 {
				return vm.push(&objects.String{Value: ""})
			}
			result := strings.Repeat(s, int(n))
			return vm.push(&objects.String{Value: result})
		}
	}

	// List concatenation: [1] + [2]
	if op == compiler.OpAdd {
		if leftType == objects.LIST_OBJ && rightType == objects.LIST_OBJ {
			leftList := left.(*objects.List)
			rightList := right.(*objects.List)
			result := make([]objects.Object, 0, len(leftList.Elements)+len(rightList.Elements))
			result = append(result, leftList.Elements...)
			result = append(result, rightList.Elements...)
			return vm.push(&objects.List{Elements: result})
		}
		// String concatenation: "a" + "b"
		if leftType == objects.STRING_OBJ && rightType == objects.STRING_OBJ {
			return vm.push(&objects.String{Value: left.(*objects.String).Value + right.(*objects.String).Value})
		}
		return vm.push(objects.NewTypeError("unsupported operand type(s) for +: '%s' and '%s'", leftType, rightType))
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

// inPlaceAttrMap maps in-place opcodes to their corresponding __ixxx__ attribute names
// and the fallback regular opcodes.
var inPlaceAttrMap = map[compiler.Opcode]struct {
	attrName   string
	fallbackOp compiler.Opcode
}{
	compiler.OpInPlaceAdd:       {"__iadd__", compiler.OpAdd},
	compiler.OpInPlaceSub:       {"__isub__", compiler.OpSub},
	compiler.OpInPlaceMul:       {"__imul__", compiler.OpMul},
	compiler.OpInPlaceDiv:       {"__itruediv__", compiler.OpDiv},
	compiler.OpInPlaceMod:       {"__imod__", compiler.OpMod},
	compiler.OpInPlaceFloorDiv:  {"__ifloordiv__", compiler.OpFloorDiv},
	compiler.OpInPlacePower:     {"__ipow__", compiler.OpPower},
	compiler.OpInPlaceBitOr:     {"__ior__", compiler.OpBitOr},
	compiler.OpInPlaceBitAnd:    {"__iand__", compiler.OpBitAnd},
	compiler.OpInPlaceBitXor:    {"__ixor__", compiler.OpBitXor},
	compiler.OpInPlaceLShift:    {"__ilshift__", compiler.OpBitOr}, // fallback to OpBitOr; true OpLShift not yet defined
	compiler.OpInPlaceRShift:    {"__irshift__", compiler.OpBitOr}, // fallback to OpBitOr; true OpRShift not yet defined
}

func (vm *VM) executeInPlaceOperation(op compiler.Opcode, left, right objects.Object) error {
	info, ok := inPlaceAttrMap[op]
	if !ok {
		// Unknown in-place opcode, fall back to regular binary operation
		vm.push(left)
		vm.push(right)
		return vm.executeBinaryOperation(op)
	}

	// 字符串 += 快速路径：使用 strings.Builder 避免重复分配
	if op == compiler.OpInPlaceAdd {
		if leftStr, ok := left.(*objects.String); ok {
			var rightStr string
			if rs, ok := right.(*objects.String); ok {
				rightStr = rs.Value
			} else {
				rightStr = right.Inspect()
			}
			var builder strings.Builder
			builder.Grow(len(leftStr.Value) + len(rightStr))
			builder.WriteString(leftStr.Value)
			builder.WriteString(rightStr)
			return vm.push(&objects.String{Value: builder.String()})
		}
	}

	// Try __ixxx__ method on left operand
	if getter, ok := left.(objects.AttributeGetter); ok {
		if method, found := getter.GetAttr(info.attrName); found {
			if builtin, ok := method.(*objects.Builtin); ok {
				result := builtin.Fn(right)
				if result.Type() != objects.ERROR_OBJ {
					return vm.push(result)
				}
				// __ixxx__ returned error, fall through to regular operation
			}
		}
	}

	// Fallback: use regular binary operation
	vm.push(left)
	vm.push(right)
	return vm.executeBinaryOperation(info.fallbackOp)
}

func toString(obj objects.Object) string {
	switch o := obj.(type) {
	case *objects.String:
		return o.Value
	case *objects.Integer:
		return fmt.Sprintf("%d", o.Value)
	case *objects.Float:
		return fmt.Sprintf("%g", o.Value)
	case *objects.Complex:
		return o.Inspect()
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
		return vm.push(&objects.Float{Value: float64(leftValue) / float64(rightValue)})
	case compiler.OpMod:
		if rightValue == 0 {
			return vm.push(objects.NewZeroDivisionError("modulo by zero"))
		}
		result = leftValue % rightValue
		if result != 0 && ((result < 0) != (rightValue < 0)) {
			result += rightValue
		}
	case compiler.OpFloorDiv:
		if rightValue == 0 {
			return vm.push(objects.NewZeroDivisionError("floor division by zero"))
		}
		result = leftValue / rightValue
		// Python floor division: floor(a/b), adjust if signs differ and there's a remainder
		if (leftValue < 0) != (rightValue < 0) && leftValue%rightValue != 0 {
			result -= 1
		}
	case compiler.OpPower:
		if rightValue < 0 {
			return vm.push(&objects.Float{Value: math.Pow(float64(leftValue), float64(rightValue))})
		}
		result = int64(1)
		for i := int64(0); i < rightValue; i++ {
			result *= leftValue
		}
	case compiler.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	case compiler.OpLessThan:
		return vm.push(nativeBoolToBooleanObject(leftValue < rightValue))
	case compiler.OpBitOr:
		result = leftValue | rightValue
	case compiler.OpBitAnd:
		result = leftValue & rightValue
	case compiler.OpBitXor:
		result = leftValue ^ rightValue
	case compiler.OpLShift:
		result = leftValue << uint(rightValue)
	case compiler.OpRShift:
		result = leftValue >> uint(rightValue)
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	err := vm.push(objects.GetCachedInteger(result))
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
	case compiler.OpMod:
		result = math.Mod(leftValue, rightValue)
	case compiler.OpFloorDiv:
		result = math.Floor(leftValue / rightValue)
	case compiler.OpPower:
		result = math.Pow(leftValue, rightValue)
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	return vm.push(&objects.Float{Value: result})
}

func (vm *VM) executeBinaryComplexOperation(op compiler.Opcode, left, right objects.Object) error {
	leftComplex := left.(*objects.Complex)
	rightComplex := right.(*objects.Complex)

	lc := complex(leftComplex.Real, leftComplex.Imag)
	rc := complex(rightComplex.Real, rightComplex.Imag)

	switch op {
	case compiler.OpAdd:
		result := lc + rc
		return vm.push(objects.NewComplex(real(result), imag(result)))
	case compiler.OpSub:
		result := lc - rc
		return vm.push(objects.NewComplex(real(result), imag(result)))
	case compiler.OpMul:
		result := lc * rc
		return vm.push(objects.NewComplex(real(result), imag(result)))
	case compiler.OpDiv:
		if rc == 0 {
			return vm.push(objects.NewZeroDivisionError("complex division by zero"))
		}
		result := lc / rc
		return vm.push(objects.NewComplex(real(result), imag(result)))
	case compiler.OpPower:
		result := cmplx.Pow(lc, rc)
		return vm.push(objects.NewComplex(real(result), imag(result)))
	default:
		return fmt.Errorf("unknown complex operator: %d", op)
	}
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
	vm.lastPopped = o
	return o
}

func (vm *VM) LastPoppedStackElem() objects.Object {
	return vm.lastPopped
}

// CallCallable calls a Python callable object and returns the result.
// This is used by Go code that needs to invoke Python functions (e.g., re.sub with callable repl).
// It supports Builtin, Closure, and CompiledFunction callees.
func (vm *VM) CallCallable(callee objects.Object, args ...objects.Object) objects.Object {
	// For builtins, just call directly
	if builtin, ok := callee.(*objects.Builtin); ok {
		return builtin.Fn(args...)
	}

	// For closures and compiled functions, set up a call frame and run the VM
	// until that frame returns.
	oldFramesIndex := vm.framesIndex

	// Push callee and args onto the stack
	vm.push(callee)
	for _, arg := range args {
		vm.push(arg)
	}

	// Execute the call setup
	err := vm.executeCall(len(args))
	if err != nil {
		return objects.NewError("callable invocation failed: %s", err.Error())
	}

	// If the call was to a builtin (which executeCall handles immediately),
	// the result is already on the stack.
	if vm.framesIndex == oldFramesIndex {
		result := vm.pop()
		return result
	}

	// For compiled functions/closures, a new frame was pushed.
	// Run the VM until we return to the original frame depth.
	for vm.framesIndex > oldFramesIndex {
		frame := vm.currentFrame()
		if frame.ip >= len(frame.fn.Instructions)-1 {
			// Frame is done without explicit return
			vm.popFrame()
			vm.sp = frame.basePointer - 1
			vm.push(objects.None_)
			break
		}
		runErr := vm.Run()
		if runErr != nil {
			// Clean up any extra frames
			for vm.framesIndex > oldFramesIndex {
				vm.popFrame()
			}
			return objects.NewError("callable execution failed: %s", runErr.Error())
		}
		break
	}

	// The result should be on the stack
	if vm.sp > 0 {
		result := vm.pop()
		return result
	}

	return objects.None_
}

func (vm *VM) EnableGC(enable bool) {
	vm.gcEnabled = enable
	if enable {
		gc.Enable()
	} else {
		gc.Disable()
	}
}

func (vm *VM) IsGCEnabled() bool {
	return vm.gcEnabled
}

func (vm *VM) SetGCThreshold(threshold int64) {
	vm.gcThreshold = threshold
	gc.GetGC().SetThreshold(threshold)
}

func (vm *VM) GetGCThreshold() int64 {
	return vm.gcThreshold
}

func (vm *VM) TriggerGC() {
	if vm.gcEnabled {
		vm.updateGCRoots()
		gc.Collect()
	}
}

func (vm *VM) updateGCRoots() {
	var roots []objects.Object

	for i := 0; i < vm.sp; i++ {
		if vm.stack[i] != nil {
			roots = append(roots, vm.stack[i])
		}
	}

	for i := 0; i < vm.framesIndex; i++ {
		frame := vm.frames[i]
		if frame != nil {
			if frame.freeVars != nil {
				roots = append(roots, frame.freeVars...)
			}
		}
	}

	for _, global := range vm.globals {
		if global != nil {
			roots = append(roots, global)
		}
	}

	for _, constant := range vm.constants {
		if constant != nil {
			roots = append(roots, constant)
		}
	}

	gc.SetRoots(nil, roots, nil, nil)
}

func (vm *VM) TrackAllocation(obj objects.Object) {
	if vm.gcEnabled {
		gc.GetGC().Register(obj)
	}
}

func (vm *VM) GetGCStats() map[string]interface{} {
	return gc.GetStats()
}

func (vm *VM) PrintGCStats() {
	gc.PrintStats()
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
		if err := objects.CheckHashable(key); err != nil {
			return nil, err
		}
		dict.Set(key, value)
	}

	return dict, nil
}

func (vm *VM) buildSet(startIndex, endIndex int) objects.Object {
	set := objects.NewSet()

	for i := startIndex; i < endIndex; i++ {
		element := vm.stack[i]
		if err := objects.CheckHashable(element); err != nil {
			return err.(*objects.Error)
		}
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
	case left.Type() == objects.RANGE_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeRangeIndex(left, index)
	case left.Type() == objects.STRING_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeStringIndex(left, index)
	case left.Type() == objects.BYTES_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeBytesIndex(left, index)
	case left.Type() == objects.DICT_KEYS_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeDictKeysIndex(left, index)
	case left.Type() == objects.DICT_VALUES_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeDictValuesIndex(left, index)
	case left.Type() == objects.DICT_ITEMS_OBJ && index.Type() == objects.INTEGER_OBJ:
		return vm.executeDictItemsIndex(left, index)
	default:
		// Check if object has __getitem__ method via GetAttr
		if getter, ok := left.(interface{ GetAttr(string) (objects.Object, bool) }); ok {
			if getitem, found := getter.GetAttr("__getitem__"); found {
				// Call __getitem__ method
				result := objects.CallFunction(getitem, index)
				if err, isErr := result.(*objects.Error); isErr {
					return err
				}
				return vm.push(result)
			}
		}
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeSetIndex(left, index, value objects.Object) error {
	switch left := left.(type) {
	case *objects.Tuple:
		return fmt.Errorf("'tuple' object does not support item assignment")
	case *objects.String:
		return fmt.Errorf("'str' object does not support item assignment")
	case *objects.Dict:
		if err := objects.CheckHashable(index); err != nil {
			return err
		}
		left.Set(index, value)
		return vm.push(value)
	case *objects.List:
		if idx, ok := index.(*objects.Integer); ok {
			i := idx.Value
			length := int64(len(left.Elements))
			if i < 0 {
				i = length + i
			}
			if i < 0 || i >= length {
				return fmt.Errorf("list assignment index out of range")
			}
			left.Elements[i] = value
			return vm.push(value)
		}
		return fmt.Errorf("list indices must be integers, not %s", index.Type())
	default:
		// Check if object has __setitem__ method via GetAttr
		if setter, ok := left.(interface{ GetAttr(string) (objects.Object, bool) }); ok {
			if setitem, found := setter.GetAttr("__setitem__"); found {
				result := objects.CallFunction(setitem, index, value)
				if err, isErr := result.(*objects.Error); isErr {
					return err
				}
				return vm.push(value)
			}
		}
		return fmt.Errorf("item assignment not supported: %s", left.Type())
	}
}

func (vm *VM) executeSetSlice(left, lower, upper, step, value objects.Object) error {
	// Only lists support slice assignment
	list, ok := left.(*objects.List)
	if !ok {
		return fmt.Errorf("'%s' object does not support slice assignment", left.Type())
	}

	// step must be None for now (we don't support extended slice assignment)
	if _, isNone := step.(*objects.None); !isNone {
		return fmt.Errorf("slice assignment with step is not supported")
	}

	length := int64(len(list.Elements))

	// Parse lower bound
	var lo int64
	if lower == nil {
		lo = 0
	} else if _, isNone := lower.(*objects.None); isNone {
		lo = 0
	} else if i, ok := lower.(*objects.Integer); ok {
		lo = i.Value
		if lo < 0 {
			lo = length + lo
			if lo < 0 {
				lo = 0
			}
		}
		if lo > length {
			lo = length
		}
	} else {
		return fmt.Errorf("slice indices must be integers or None")
	}

	// Parse upper bound
	var hi int64
	if upper == nil {
		hi = length
	} else if _, isNone := upper.(*objects.None); isNone {
		hi = length
	} else if i, ok := upper.(*objects.Integer); ok {
		hi = i.Value
		if hi < 0 {
			hi = length + hi
			if hi < 0 {
				hi = 0
			}
		}
		if hi > length {
			hi = length
		}
	} else {
		return fmt.Errorf("slice indices must be integers or None")
	}

	// Get replacement values
	var replacements []objects.Object
	switch v := value.(type) {
	case *objects.List:
		replacements = v.Elements
	case *objects.Tuple:
		replacements = v.Elements
	default:
		return fmt.Errorf("can only assign an iterable to a slice")
	}

	// Perform slice replacement
	if lo > hi {
		lo = hi
	}
	oldLen := int64(len(list.Elements))
	newLen := oldLen - (hi - lo) + int64(len(replacements))

	if newLen > oldLen {
		// Grow the list
		list.Elements = append(list.Elements, make([]objects.Object, newLen-oldLen)...)
		// Shift elements right
		copy(list.Elements[lo+int64(len(replacements)):], list.Elements[hi:])
	} else if newLen < oldLen {
		// Shift elements left first
		copy(list.Elements[lo+int64(len(replacements)):], list.Elements[hi:])
		// Truncate
		list.Elements = list.Elements[:newLen]
	}

	copy(list.Elements[lo:], replacements)
	return vm.push(value)
}

func (vm *VM) executeArrayIndex(array, index objects.Object) error {
	arrayObject := array.(*objects.List)
	idx := index.(*objects.Integer).Value
	length := int64(len(arrayObject.Elements))

	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return vm.push(objects.NewIndexError("list index out of range"))
	}

	return vm.push(arrayObject.Elements[idx])
}

func (vm *VM) executeTupleIndex(tuple, index objects.Object) error {
	tupleObject := tuple.(*objects.Tuple)
	idx := index.(*objects.Integer).Value
	length := int64(len(tupleObject.Elements))

	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return vm.push(objects.NewIndexError("tuple index out of range"))
	}

	return vm.push(tupleObject.Elements[idx])
}

func (vm *VM) executeRangeIndex(left, index objects.Object) error {
	rangeObj := left.(*objects.Range)
	idx := index.(*objects.Integer).Value
	val, ok := rangeObj.GetItem(idx)
	if !ok {
		return vm.push(objects.NewIndexError("range object index out of range"))
	}
	return vm.push(val)
}

func (vm *VM) executeStringIndex(left, index objects.Object) error {
	str := left.(*objects.String)
	idx := index.(*objects.Integer).Value
	length := int64(len(str.Value))
	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return vm.push(objects.NewIndexError("string index out of range"))
	}
	return vm.push(&objects.String{Value: string(str.Value[idx])})
}

func (vm *VM) executeBytesIndex(left, index objects.Object) error {
	bts := left.(*objects.Bytes)
	idx := index.(*objects.Integer).Value
	length := int64(len(bts.Value))
	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return vm.push(objects.None_)
	}
	return vm.push(objects.GetCachedInteger(int64(bts.Value[idx])))
}

func (vm *VM) executeHashIndex(hash, index objects.Object) error {
	hashObject := hash.(*objects.Dict)

	value, ok := hashObject.Get(index)
	if !ok {
		return vm.push(objects.None_)
	}

	return vm.push(value)
}

func (vm *VM) executeDictKeysIndex(left, index objects.Object) error {
	dk := left.(*objects.DictKeys)
	idx := index.(*objects.Integer).Value
	val, ok := dk.GetItem(idx)
	if !ok {
		return vm.push(objects.None_)
	}
	return vm.push(val)
}

func (vm *VM) executeDictValuesIndex(left, index objects.Object) error {
	dv := left.(*objects.DictValues)
	idx := index.(*objects.Integer).Value
	val, ok := dv.GetItem(idx)
	if !ok {
		return vm.push(objects.None_)
	}
	return vm.push(val)
}

func (vm *VM) executeDictItemsIndex(left, index objects.Object) error {
	di := left.(*objects.DictItems)
	idx := index.(*objects.Integer).Value
	val, ok := di.GetItem(idx)
	if !ok {
		return vm.push(objects.None_)
	}
	return vm.push(val)
}

func (vm *VM) executeSliceExpression(left, start, end objects.Object) error {
	switch {
	case left.Type() == objects.LIST_OBJ:
		return vm.executeListSlice(left, start, end)
	case left.Type() == objects.STRING_OBJ:
		return vm.executeStringSlice(left, start, end)
	case left.Type() == objects.BYTES_OBJ:
		return vm.executeBytesSlice(left, start, end)
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

func (vm *VM) executeBytesSlice(left, start, end objects.Object) error {
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
		return vm.push(&objects.Bytes{Value: []byte{}})
	}

	sliced := make([]byte, endIdx-startIdx)
	copy(sliced, bts.Value[startIdx:endIdx])
	return vm.push(&objects.Bytes{Value: sliced})
}

func matchesException(errObj objects.Object, exceptionType string) bool {
	if exceptionType == "" {
		return true
	}
	switch err := errObj.(type) {
	case *objects.Error:
		if exceptionType == "Exception" || exceptionType == "Error" {
			return true
		}
		return err.ErrorType == exceptionType
	case *objects.ExceptionGroup:
		// For ExceptionGroup, check if any contained exception matches
		for _, exc := range err.Exceptions {
			if matchesException(exc, exceptionType) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// splitExceptionGroup splits an ExceptionGroup into matching and non-matching exceptions
// based on the exception type. Returns (matched, unmatched) slices.
// findNextExceptStarHandler scans the bytecode after the given opcode IP
// to find the next OpExceptStarHandler instruction. Returns its IP or -1 if not found.
func (vm *VM) findNextExceptStarHandler(afterOpcodeIP int) int {
	ins := vm.currentFrame().fn.Instructions
	scanIP := afterOpcodeIP + 5 // skip past the current 5-byte handler instruction
	for scanIP < len(ins) {
		op := compiler.Opcode(ins[scanIP])
		if op == compiler.OpExceptStarHandler {
			return scanIP
		}
		if op == compiler.OpEndTry || op == compiler.OpFinally {
			return -1
		}
		// Skip past this instruction
		size := compiler.InstructionSize(op)
		if size <= 0 {
			scanIP++
		} else {
			scanIP += size
		}
	}
	return -1
}

// GetJITStats 返回JIT编译器的统计信息
func (vm *VM) GetJITStats() map[string]interface{} {
	if vm.jitEngine == nil {
		return map[string]interface{}{"enabled": false}
	}
	return vm.jitEngine.GetStats()
}

// SetJITHotThreshold 设置JIT热点检测阈值
func (vm *VM) SetJITHotThreshold(threshold int64) {
	if vm.jitEngine != nil {
		vm.jitEngine.SetHotThreshold(threshold)
	}
}

// GetJITHotFunctions 返回热点函数列表
func (vm *VM) GetJITHotFunctions() []*jit.CompiledCode {
	if vm.jitEngine == nil {
		return nil
	}
	return vm.jitEngine.GetHotFunctions()
}

// ClearJITCache 清除JIT缓存
func (vm *VM) ClearJITCache() {
	if vm.jitEngine != nil {
		vm.jitEngine.ClearCache()
	}
}

func splitExceptionGroup(eg *objects.ExceptionGroup, exceptionType string) ([]objects.Object, []objects.Object) {
	var matched, unmatched []objects.Object
	for _, exc := range eg.Exceptions {
		if matchesException(exc, exceptionType) {
			matched = append(matched, exc)
		} else {
			unmatched = append(unmatched, exc)
		}
	}
	return matched, unmatched
}

func (vm *VM) raiseException(errObj objects.Object) bool {
	ins := vm.currentFrame().fn.Instructions

	for i := len(vm.exceptionStack) - 1; i >= 0; i-- {
		handler := vm.exceptionStack[i]

		if handler.exceptCount == 0 && !handler.hasFinally {
			continue
		}

		for vm.framesIndex > handler.frameIndex {
			vm.popFrame()
		}
		ins = vm.currentFrame().fn.Instructions

		foundHandler := false

		if handler.handlerIP >= 0 {
			scanIP := handler.handlerIP
			for scanIP < len(ins) {
				op := compiler.Opcode(ins[scanIP])
				if op == compiler.OpExceptStarHandler {
					typeIdx := int(uint16(ins[scanIP+1])<<8 | uint16(ins[scanIP+2]))
					var exceptionType string
					if typeIdx > 0 && typeIdx < len(vm.constants) {
						if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
							exceptionType = typeObj.Value
						}
					}
					// except* handles ExceptionGroup
					if eg, ok := errObj.(*objects.ExceptionGroup); ok {
						matched, unmatched := splitExceptionGroup(eg, exceptionType)
						if len(matched) > 0 {
							matchedEG := &objects.ExceptionGroup{
								Message:    eg.Message,
								Exceptions: matched,
							}
							vm.sp = handler.stackPtr
							if err := vm.push(matchedEG); err != nil {
								return false
							}
							vm.currentFrame().ip = scanIP + 5 - 1
							// Mark this handler as except* and store unhandled exceptions
							vm.exceptionStack[i].isStar = true
							vm.exceptionStack[i].exceptOpcodeIP = scanIP
							vm.exceptionStack[i].starUnhandled = append(vm.exceptionStack[i].starUnhandled, unmatched...)
							foundHandler = true
							break
						}
						// No matching exceptions in this group, skip this handler
						scanIP += 5
					} else if err, ok := errObj.(*objects.Error); ok {
						// Plain error with except*: wrap in ExceptionGroup if matches
						if exceptionType == "" || matchesException(errObj, exceptionType) {
							wrappedEG := &objects.ExceptionGroup{
								Message:    "",
								Exceptions: []objects.Object{err},
							}
							vm.sp = handler.stackPtr
							if err := vm.push(wrappedEG); err != nil {
								return false
							}
							vm.currentFrame().ip = scanIP + 5 - 1
							// Mark this handler as except*
							vm.exceptionStack[i].isStar = true
							vm.exceptionStack[i].exceptOpcodeIP = scanIP
							foundHandler = true
							break
						}
						scanIP += 5
					} else {
						scanIP += 5
					}
				} else if op == compiler.OpExceptHandler {
					typeIdx := int(uint16(ins[scanIP+1])<<8 | uint16(ins[scanIP+2]))
					var exceptionType string
					if typeIdx > 0 && typeIdx < len(vm.constants) {
						if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
							exceptionType = typeObj.Value
						}
					}
					// Typed except does NOT catch ExceptionGroup (but bare except: does)
					if _, isEG := errObj.(*objects.ExceptionGroup); isEG && exceptionType != "" {
						scanIP += 5
						continue
					}
					if exceptionType == "" || matchesException(errObj, exceptionType) {
						vm.sp = handler.stackPtr
						if err := vm.push(errObj); err != nil {
							return false
						}
						vm.currentFrame().ip = scanIP + 5 - 1
						foundHandler = true
						break
					}
					scanIP += 5
				} else if op == compiler.OpFinally || op == compiler.OpEndTry {
					break
				} else {
					scanIP++
				}
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

			// Typed except does NOT catch ExceptionGroup (but bare except: does)
			if _, isEG := errObj.(*objects.ExceptionGroup); isEG && exceptionType != "" {
				ip += 5
				continue
			}

			if exceptionType == "" || matchesException(errObj, exceptionType) {
				handlerStartIP := ip + 5
				vm.currentFrame().ip = handlerStartIP
				return handlerStartIP
			}

			ip += 5
		} else if op == compiler.OpExceptStarHandler {
			typeIdx := int(uint16(vm.currentFrame().fn.Instructions[ip+2])<<8 | uint16(vm.currentFrame().fn.Instructions[ip+1]))

			var exceptionType string
			if typeIdx > 0 && typeIdx < len(vm.constants) {
				if typeObj, ok := vm.constants[typeIdx].(*objects.String); ok {
					exceptionType = typeObj.Value
				}
			}

			if eg, ok := errObj.(*objects.ExceptionGroup); ok {
				matched, _ := splitExceptionGroup(eg, exceptionType)
				if len(matched) > 0 {
					handlerStartIP := ip + 5
					vm.currentFrame().ip = handlerStartIP
					return handlerStartIP
				}
			} else if _, ok := errObj.(*objects.Error); ok {
				if exceptionType == "" || matchesException(errObj, exceptionType) {
					handlerStartIP := ip + 5
					vm.currentFrame().ip = handlerStartIP
					return handlerStartIP
				}
			}

			ip += 5
		} else {
			break
		}
	}
	return -1
}

