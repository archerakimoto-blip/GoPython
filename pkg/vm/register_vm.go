package vm

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

// RegisterVM executes bytecode using a register-based model.
// It translates stack-based bytecode to register-based on the fly and executes it.
type RegisterVM struct {
	vm *VM // reference back to the shared VM state

	registers []objects.Object // virtual registers
	numRegs   int              // number of allocated registers

	frames     []*RegFrame
	frameIndex int

	// Translation cache: maps stack bytecode hash to register bytecode
	translationCache map[uint32][]RegInstruction
}

// RegFrame represents a call frame in the register VM.
type RegFrame struct {
	instructions []RegInstruction
	ip           int
	basePointer  int

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
)

// RegInstruction represents a single register-based instruction.
type RegInstruction struct {
	Opcode   RegOpcode
	Operands []int
}

// newRegisterVM creates a new RegisterVM that shares state with the given VM.
func newRegisterVM(vm *VM) *RegisterVM {
	return &RegisterVM{
		vm:               vm,
		registers:        make([]objects.Object, 512),
		numRegs:          0,
		frames:           make([]*RegFrame, MaxFrames),
		frameIndex:       0,
		translationCache: make(map[uint32][]RegInstruction),
	}
}

// bytecodeHash computes a simple hash of the bytecode for caching.
func bytecodeHash(instructions []byte) uint32 {
	h := uint32(0)
	for i, b := range instructions {
		h = h*31 + uint32(b) + uint32(i)*7
	}
	return h
}

// registerAllocator tracks register allocation with a free list for reuse.
type registerAllocator struct {
	nextReg  int   // next register number if free list is empty
	freeList []int // reusable register numbers
}

func (ra *registerAllocator) alloc() int {
	if len(ra.freeList) > 0 {
		reg := ra.freeList[len(ra.freeList)-1]
		ra.freeList = ra.freeList[:len(ra.freeList)-1]
		return reg
	}
	reg := ra.nextReg
	ra.nextReg++
	return reg
}

func (ra *registerAllocator) free(reg int) {
	ra.freeList = append(ra.freeList, reg)
}

// translate converts stack-based bytecode to register-based bytecode.
// It maintains a virtual stack that maps stack positions to registers.
// Registers are reused via a free list when they are no longer needed.
func (rvm *RegisterVM) translate(instructions []byte, constants []objects.Object) []RegInstruction {
	var regInstructions []RegInstruction
	ra := registerAllocator{}               // register allocator with free list
	stackToReg := make([]int, 1024)         // maps stack position to register
	sp := 0

	// Map from stack bytecode IP to register instruction index.
	// Since register instructions have different lengths and counts than
	// stack bytecode instructions, we must translate jump targets through
	// this mapping to get correct register instruction indices.
	stackIPToRegIP := make(map[int]int)

	ip := 0
	for ip < len(instructions) {
		// Record the mapping before processing this instruction.
		stackIPToRegIP[ip] = len(regInstructions)

		op := compiler.Opcode(instructions[ip])

		switch op {
		case compiler.OpConstant:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpLoadConst, []int{reg, idx}})
			ip += 3

		case compiler.OpPop:
			sp--
			ra.free(stackToReg[sp])
			// RegOpPop is a no-op in register VM; we just decrement the virtual stack
			regInstructions = append(regInstructions, RegInstruction{RegOpPop, []int{}})
			ip += 1

		case compiler.OpDupTop:
			src := stackToReg[sp-1]
			dst := ra.alloc()
			stackToReg[sp] = dst
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpDupTop, []int{dst, src}})
			ip += 1

		case compiler.OpAdd:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpAdd, []int{dst, src1, src2}})
			ip += 1

		case compiler.OpSub:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpSub, []int{dst, src1, src2}})
			ip += 1

		case compiler.OpMul:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpMul, []int{dst, src1, src2}})
			ip += 1

		case compiler.OpDiv:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpDiv, []int{dst, src1, src2}})
			ip += 1

		case compiler.OpMod:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpMod, []int{dst, src1, src2}})
			ip += 1

		case compiler.OpFloorDiv:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpFloorDiv, []int{dst, src1, src2}})
			ip += 1

		case compiler.OpPower:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpPower, []int{dst, src1, src2}})
			ip += 1

		case compiler.OpTrue:
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpTrue, []int{reg}})
			ip += 1

		case compiler.OpFalse:
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpFalse, []int{reg}})
			ip += 1

		case compiler.OpNull:
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpNull, []int{reg}})
			ip += 1

		case compiler.OpEllipsis:
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpEllipsis, []int{reg}})
			ip += 1

		case compiler.OpEqual:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpCompare, []int{dst, src1, src2, int(compiler.OpEqual)}})
			ip += 1

		case compiler.OpNotEqual:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpCompare, []int{dst, src1, src2, int(compiler.OpNotEqual)}})
			ip += 1

		case compiler.OpGreaterThan:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpCompare, []int{dst, src1, src2, int(compiler.OpGreaterThan)}})
			ip += 1

		case compiler.OpLessThan:
			src1 := stackToReg[sp-2]
			src2 := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src1)
			ra.free(src2)
			stackToReg[sp-2] = dst
			sp--
			regInstructions = append(regInstructions, RegInstruction{RegOpCompare, []int{dst, src1, src2, int(compiler.OpLessThan)}})
			ip += 1

		case compiler.OpJump:
			pos := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			regInstructions = append(regInstructions, RegInstruction{RegOpJump, []int{pos}})
			ip += 3

		case compiler.OpJumpNotTruthy:
			pos := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			condReg := stackToReg[sp-1]
			sp--
			ra.free(condReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpJumpIfFalse, []int{condReg, pos}})
			ip += 3

		case compiler.OpMinus:
			src := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpNegate, []int{dst, src}})
			ip += 1

		case compiler.OpBang:
			src := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(src)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpNot, []int{dst, src}})
			ip += 1

		case compiler.OpSetGlobal:
			globalIndex := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			valReg := stackToReg[sp-1]
			sp--
			ra.free(valReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpSetGlobal, []int{globalIndex, valReg}})
			ip += 3

		case compiler.OpGetGlobal:
			globalIndex := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpGetGlobal, []int{reg, globalIndex}})
			ip += 3

		case compiler.OpSetLocal:
			localIndex := int(instructions[ip+1])
			valReg := stackToReg[sp-1]
			sp--
			ra.free(valReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpSetLocal, []int{localIndex, valReg}})
			ip += 2

		case compiler.OpGetLocal:
			localIndex := int(instructions[ip+1])
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpGetLocal, []int{reg, localIndex}})
			ip += 2

		case compiler.OpGetFree:
			freeIndex := int(instructions[ip+1])
			reg := ra.alloc()
			stackToReg[sp] = reg
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpGetFree, []int{reg, freeIndex}})
			ip += 2

		case compiler.OpCall:
			numArgs := int(instructions[ip+1])
			// The function is at sp-numArgs-1, args are at sp-numArgs..sp-1
			funcReg := stackToReg[sp-numArgs-1]
			argRegs := make([]int, numArgs)
			for i := 0; i < numArgs; i++ {
				argRegs[i] = stackToReg[sp-numArgs+i]
			}
			dst := ra.alloc()
			// After call, the func and args are consumed, result is pushed
			ra.free(funcReg)
			for i := 0; i < numArgs; i++ {
				ra.free(argRegs[i])
			}
			sp = sp - numArgs - 1
			stackToReg[sp] = dst
			sp++
			operands := []int{dst, funcReg, numArgs}
			operands = append(operands, argRegs...)
			regInstructions = append(regInstructions, RegInstruction{RegOpCall, operands})
			ip += 2

		case compiler.OpReturnValue:
			valReg := stackToReg[sp-1]
			ra.free(valReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpReturn, []int{valReg}})
			ip += 1

		case compiler.OpReturn:
			regInstructions = append(regInstructions, RegInstruction{RegOpReturnNone, []int{}})
			ip += 1

		case compiler.OpClosure:
			constIndex := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			numFree := int(instructions[ip+3])
			freeRegs := make([]int, numFree)
			for i := 0; i < numFree; i++ {
				freeRegs[i] = stackToReg[sp-numFree+i]
			}
			sp -= numFree
			dst := ra.alloc()
			for i := 0; i < numFree; i++ {
				ra.free(freeRegs[i])
			}
			stackToReg[sp] = dst
			sp++
			operands := []int{dst, constIndex, numFree}
			operands = append(operands, freeRegs...)
			regInstructions = append(regInstructions, RegInstruction{RegOpClosure, operands})
			ip += 4

		case compiler.OpArray:
			numElements := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			elementRegs := make([]int, numElements)
			for i := 0; i < numElements; i++ {
				elementRegs[i] = stackToReg[sp-numElements+i]
			}
			sp -= numElements
			dst := ra.alloc()
			for i := 0; i < numElements; i++ {
				ra.free(elementRegs[i])
			}
			stackToReg[sp] = dst
			sp++
			operands := []int{dst, numElements}
			operands = append(operands, elementRegs...)
			regInstructions = append(regInstructions, RegInstruction{RegOpArray, operands})
			ip += 3

		case compiler.OpHash:
			numElements := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			elementRegs := make([]int, numElements)
			for i := 0; i < numElements; i++ {
				elementRegs[i] = stackToReg[sp-numElements+i]
			}
			sp -= numElements
			dst := ra.alloc()
			for i := 0; i < numElements; i++ {
				ra.free(elementRegs[i])
			}
			stackToReg[sp] = dst
			sp++
			operands := []int{dst, numElements}
			operands = append(operands, elementRegs...)
			regInstructions = append(regInstructions, RegInstruction{RegOpHash, operands})
			ip += 3

		case compiler.OpSet:
			numElements := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			elementRegs := make([]int, numElements)
			for i := 0; i < numElements; i++ {
				elementRegs[i] = stackToReg[sp-numElements+i]
			}
			sp -= numElements
			dst := ra.alloc()
			for i := 0; i < numElements; i++ {
				ra.free(elementRegs[i])
			}
			stackToReg[sp] = dst
			sp++
			operands := []int{dst, numElements}
			operands = append(operands, elementRegs...)
			regInstructions = append(regInstructions, RegInstruction{RegOpSet, operands})
			ip += 3

		case compiler.OpIndex:
			indexReg := stackToReg[sp-1]
			leftReg := stackToReg[sp-2]
			dst := ra.alloc()
			ra.free(leftReg)
			ra.free(indexReg)
			sp -= 2
			stackToReg[sp] = dst
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpIndex, []int{dst, leftReg, indexReg}})
			ip += 1

		case compiler.OpSlice:
			stepReg := stackToReg[sp-1]
			endReg := stackToReg[sp-2]
			startReg := stackToReg[sp-3]
			leftReg := stackToReg[sp-4]
			dst := ra.alloc()
			ra.free(leftReg)
			ra.free(startReg)
			ra.free(endReg)
			ra.free(stepReg)
			sp -= 4
			stackToReg[sp] = dst
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpSlice, []int{dst, leftReg, startReg, endReg, stepReg}})
			ip += 1

		case compiler.OpListUnpack:
			listReg := stackToReg[sp-1]
			ra.free(listReg)
			// We don't know how many elements will be unpacked at translation time,
			// so we emit a special instruction
			regInstructions = append(regInstructions, RegInstruction{RegOpListUnpack, []int{listReg}})
			ip += 1

		case compiler.OpDictUnpack:
			regInstructions = append(regInstructions, RegInstruction{RegOpDictUnpack, []int{}})
			ip += 1

		case compiler.OpGetAttribute:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			objReg := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(objReg)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpGetAttr, []int{dst, objReg, idx}})
			ip += 3

		case compiler.OpSetAttribute:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			valReg := stackToReg[sp-1]
			objReg := stackToReg[sp-2]
			sp -= 2
			ra.free(valReg)
			ra.free(objReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpSetAttr, []int{objReg, valReg, idx}})
			ip += 3

		case compiler.OpDelAttribute:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			objReg := stackToReg[sp-1]
			sp--
			ra.free(objReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpDelAttribute, []int{objReg, idx}})
			ip += 3

		case compiler.OpSetClassField:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			valReg := stackToReg[sp-1]
			classReg := stackToReg[sp-2]
			sp -= 2
			ra.free(valReg)
			ra.free(classReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpSetClassField, []int{classReg, valReg, idx}})
			ip += 3

		case compiler.OpSetMetaclass:
			metaclassReg := stackToReg[sp-1]
			classReg := stackToReg[sp-2]
			sp -= 2
			dst := ra.alloc()
			ra.free(metaclassReg)
			ra.free(classReg)
			stackToReg[sp] = dst
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpSetMetaclass, []int{dst, classReg, metaclassReg}})
			ip += 1

		case compiler.OpCallMetaclassInit:
			classReg := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(classReg)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpCallMetaclassInit, []int{dst, classReg}})
			ip += 1

		case compiler.OpCreateClass:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			dst := ra.alloc()
			stackToReg[sp] = dst
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpCreateClass, []int{dst, idx}})
			ip += 3

		case compiler.OpCreateClassWithSuper:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			superReg := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(superReg)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpCreateClassWithSuper, []int{dst, idx, superReg}})
			ip += 3

		case compiler.OpCreateClassWithMultiSuper:
			idx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			numParents := int(instructions[ip+3])
			parentRegs := make([]int, numParents)
			for i := 0; i < numParents; i++ {
				parentRegs[i] = stackToReg[sp-numParents+i]
			}
			sp -= numParents
			dst := ra.alloc()
			for i := 0; i < numParents; i++ {
				ra.free(parentRegs[i])
			}
			stackToReg[sp] = dst
			sp++
			operands := []int{dst, idx, numParents}
			operands = append(operands, parentRegs...)
			regInstructions = append(regInstructions, RegInstruction{RegOpCreateClassWithMultiSuper, operands})
			ip += 4

		case compiler.OpFormatString:
			partsCount := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			partRegs := make([]int, partsCount)
			for i := 0; i < partsCount; i++ {
				partRegs[i] = stackToReg[sp-partsCount+i]
			}
			sp -= partsCount
			dst := ra.alloc()
			for i := 0; i < partsCount; i++ {
				ra.free(partRegs[i])
			}
			stackToReg[sp] = dst
			sp++
			operands := []int{dst, partsCount}
			operands = append(operands, partRegs...)
			regInstructions = append(regInstructions, RegInstruction{RegOpFormatString, operands})
			ip += 3

		case compiler.OpMakeGenerator:
			fnReg := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(fnReg)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpMakeGenerator, []int{dst, fnReg}})
			ip += 1

		case compiler.OpMakeAsync:
			fnReg := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(fnReg)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpMakeAsync, []int{dst, fnReg}})
			ip += 1

		case compiler.OpAwait:
			valReg := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(valReg)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpAwait, []int{dst, valReg}})
			ip += 1

		case compiler.OpYieldValue:
			valReg := stackToReg[sp-1]
			sp--
			ra.free(valReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpYieldValue, []int{valReg}})
			ip += 1

		case compiler.OpEnterContext:
			ctxReg := stackToReg[sp-1]
			dst := ra.alloc()
			ra.free(ctxReg)
			stackToReg[sp-1] = dst
			regInstructions = append(regInstructions, RegInstruction{RegOpEnterContext, []int{dst, ctxReg}})
			ip += 1

		case compiler.OpExitContext:
			excReg := stackToReg[sp-1]
			ctxReg := stackToReg[sp-2]
			sp -= 2
			dst := ra.alloc()
			ra.free(ctxReg)
			ra.free(excReg)
			stackToReg[sp] = dst
			sp++
			regInstructions = append(regInstructions, RegInstruction{RegOpExitContext, []int{dst, ctxReg, excReg}})
			ip += 1

		case compiler.OpBeginTry:
			exceptCount := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			hasFinally := int(uint16(instructions[ip+3])<<8 | uint16(instructions[ip+4]))
			handlerIP := int(uint16(instructions[ip+5])<<8 | uint16(instructions[ip+6]))
			finallyStartIP := int(uint16(instructions[ip+7])<<8 | uint16(instructions[ip+8]))
			regInstructions = append(regInstructions, RegInstruction{RegOpBeginTry, []int{exceptCount, hasFinally, handlerIP, finallyStartIP}})
			ip += 9

		case compiler.OpEndTry:
			regInstructions = append(regInstructions, RegInstruction{RegOpEndTry, []int{}})
			ip += 1

		case compiler.OpRaise:
			errReg := stackToReg[sp-1]
			sp--
			ra.free(errReg)
			regInstructions = append(regInstructions, RegInstruction{RegOpRaise, []int{errReg}})
			ip += 1

		case compiler.OpExceptHandler:
			typeIdx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			varIdx := int(uint16(instructions[ip+3])<<8 | uint16(instructions[ip+4]))
			regInstructions = append(regInstructions, RegInstruction{RegOpExceptHandler, []int{typeIdx, varIdx}})
			ip += 5

		case compiler.OpExceptStarHandler:
			typeIdx := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			varIdx := int(uint16(instructions[ip+3])<<8 | uint16(instructions[ip+4]))
			regInstructions = append(regInstructions, RegInstruction{RegOpExceptStarHandler, []int{typeIdx, varIdx}})
			ip += 5

		case compiler.OpFinally:
			finallyEndIP := int(uint16(instructions[ip+1])<<8 | uint16(instructions[ip+2]))
			regInstructions = append(regInstructions, RegInstruction{RegOpFinally, []int{finallyEndIP}})
			ip += 3

		default:
			// For unhandled opcodes, skip based on instruction size
			size := compiler.InstructionSize(op)
			if size <= 0 {
				size = 1
			}
			ip += size
		}
	}

	// Second pass: fix up jump targets from stack bytecode IPs to register instruction indices.
	// During the first pass, jump/branch instructions stored the original stack bytecode IP
	// as their target operand. We now translate those through stackIPToRegIP.
	for i, instr := range regInstructions {
		switch instr.Opcode {
		case RegOpJump:
			// Operand[0] is the stack IP target
			if targetRegIP, ok := stackIPToRegIP[instr.Operands[0]]; ok {
				instr.Operands[0] = targetRegIP
				regInstructions[i] = instr
			}
		case RegOpJumpIfFalse:
			// Operand[1] is the stack IP target (Operand[0] is the condition register)
			if targetRegIP, ok := stackIPToRegIP[instr.Operands[1]]; ok {
				instr.Operands[1] = targetRegIP
				regInstructions[i] = instr
			}
		case RegOpBeginTry:
			// Operands[2] = handlerIP (stack IP), Operands[3] = finallyStartIP (stack IP)
			if targetRegIP, ok := stackIPToRegIP[instr.Operands[2]]; ok {
				instr.Operands[2] = targetRegIP
			}
			if targetRegIP, ok := stackIPToRegIP[instr.Operands[3]]; ok {
				instr.Operands[3] = targetRegIP
			}
			regInstructions[i] = instr
		case RegOpFinally:
			// Operand[0] = finallyEndIP (stack IP)
			if targetRegIP, ok := stackIPToRegIP[instr.Operands[0]]; ok {
				instr.Operands[0] = targetRegIP
				regInstructions[i] = instr
			}
		}
	}

	return regInstructions
}

// getOrCreateTranslation translates bytecode and caches the result.
func (rvm *RegisterVM) getOrCreateTranslation(instructions []byte, constants []objects.Object) []RegInstruction {
	hash := bytecodeHash(instructions)
	if cached, ok := rvm.translationCache[hash]; ok {
		return cached
	}
	result := rvm.translate(instructions, constants)
	rvm.translationCache[hash] = result
	return result
}

// regSet sets a register value, growing the registers slice if needed.
func (rvm *RegisterVM) regSet(reg int, val objects.Object) {
	if reg >= len(rvm.registers) {
		newRegs := make([]objects.Object, reg*2+1)
		copy(newRegs, rvm.registers)
		rvm.registers = newRegs
	}
	rvm.registers[reg] = val
	if reg >= rvm.numRegs {
		rvm.numRegs = reg + 1
	}
}

// regGet gets a register value.
func (rvm *RegisterVM) regGet(reg int) objects.Object {
	if reg < 0 || reg >= len(rvm.registers) {
		return nil
	}
	return rvm.registers[reg]
}

// RunReg executes the register-based VM for the current frame.
// It translates the stack-based bytecode to register-based and executes it.
// For opcodes that are too complex to handle in the register VM, it falls
// back to the stack-based VM by syncing state.
func (rvm *RegisterVM) RunReg() error {
	vm := rvm.vm
	frame := vm.currentFrame()

	// Translate the bytecode for this frame
	regInstructions := rvm.getOrCreateTranslation(frame.fn.Instructions, vm.constants)

	// Initialize registers from the stack state
	// Copy existing stack values into registers
	rvm.numRegs = 0
	for i := 0; i < vm.sp; i++ {
		rvm.regSet(i, vm.stack[i])
	}

	// Execute register instructions
	ip := frame.ip + 1 // start from current IP + 1 (since Run() already increments)

	for ip < len(regInstructions) {
		inst := regInstructions[ip]
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
			if result != nil && result.Type() == objects.ERROR_OBJ {
				caught := vm.raiseException(result)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", result.Inspect())
				}
				return nil // exception was caught, exit register VM for this frame
			}

		case RegOpSub:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpSub, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)
			if result != nil && result.Type() == objects.ERROR_OBJ {
				caught := vm.raiseException(result)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", result.Inspect())
				}
				return nil
			}

		case RegOpMul:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpMul, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)
			if result != nil && result.Type() == objects.ERROR_OBJ {
				caught := vm.raiseException(result)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", result.Inspect())
				}
				return nil
			}

		case RegOpDiv:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpDiv, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)
			if result != nil && result.Type() == objects.ERROR_OBJ {
				caught := vm.raiseException(result)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", result.Inspect())
				}
				return nil
			}

		case RegOpMod:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpMod, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)
			if result != nil && result.Type() == objects.ERROR_OBJ {
				caught := vm.raiseException(result)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", result.Inspect())
				}
				return nil
			}

		case RegOpFloorDiv:
			dst := inst.Operands[0]
			src1 := inst.Operands[1]
			src2 := inst.Operands[2]
			result, err := rvm.binaryOp(compiler.OpFloorDiv, rvm.regGet(src1), rvm.regGet(src2))
			if err != nil {
				return err
			}
			rvm.regSet(dst, result)
			if result != nil && result.Type() == objects.ERROR_OBJ {
				caught := vm.raiseException(result)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", result.Inspect())
				}
				return nil
			}

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
				rvm.regSet(dst, vm.stack[frame.basePointer-len(frame.fn.Free)+freeIndex])
			}

		case RegOpCall:
			// Fall back to stack-based execution for calls
			// because executeCall is too complex to duplicate
			dst := inst.Operands[0]
			funcReg := inst.Operands[1]
			numArgs := inst.Operands[2]

			// Push function and args onto the stack
			vm.push(rvm.regGet(funcReg))
			for i := 0; i < numArgs; i++ {
				vm.push(rvm.regGet(inst.Operands[3+i]))
			}

			numArgs += vm.unpackExtraArgs
			vm.unpackExtraArgs = 0

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}

			// The result is on top of the stack, pop it into the destination register
			result := vm.pop()
			rvm.regSet(dst, result)

		case RegOpReturn:
			valReg := inst.Operands[0]
			returnValue := rvm.regGet(valReg)

			poppedFrame := vm.popFrame()

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

			vm.sp = poppedFrame.basePointer - 1

			if poppedFrame.setAttrValue != nil {
				vm.push(poppedFrame.setAttrValue)
			} else if poppedFrame.initInstance != nil {
				vm.push(poppedFrame.initInstance)
			} else if poppedFrame.metaclassInitClass != nil {
				vm.push(poppedFrame.metaclassInitClass)
			} else {
				vm.push(returnValue)
			}
			return nil

		case RegOpReturnNone:
			poppedFrame := vm.popFrame()

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

			vm.sp = poppedFrame.basePointer - 1

			if poppedFrame.setAttrValue != nil {
				vm.push(poppedFrame.setAttrValue)
			} else if poppedFrame.initInstance != nil {
				vm.push(poppedFrame.initInstance)
			} else if poppedFrame.metaclassInitClass != nil {
				vm.push(poppedFrame.metaclassInitClass)
			} else {
				vm.push(objects.None_)
			}
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

		case RegOpArray:
			dst := inst.Operands[0]
			numElements := inst.Operands[1]
			elements := make([]objects.Object, numElements)
			for i := 0; i < numElements; i++ {
				elements[i] = rvm.regGet(inst.Operands[2+i])
			}
			rvm.regSet(dst, objects.NewList(elements))

		case RegOpHash:
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

		case RegOpSet:
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

		case RegOpListUnpack:
			// Fall back to stack for list unpack
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
			// Dict is already in a register; when RegOpCall pushes args onto the stack,
			// the dict will be the last argument. executeCall detects a Dict as the last
			// argument and treats it as **kwargs. We just need to ensure the dict register
			// is included in the call's argument list.
			// No additional action needed here - the dict stays in its register and
			// will be pushed by RegOpCall as part of the arguments.

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

		case RegOpSetClassField:
			classReg := inst.Operands[0]
			valReg := inst.Operands[1]
			fieldIdx := inst.Operands[2]
			classObj := rvm.regGet(classReg)
			val := rvm.regGet(valReg)
			fieldName := vm.constants[fieldIdx].(*objects.String).Value
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
			}

		case RegOpSetMetaclass:
			dst := inst.Operands[0]
			classReg := inst.Operands[1]
			metaclassReg := inst.Operands[2]
			classObj := rvm.regGet(classReg)
			metaclassObj := rvm.regGet(metaclassReg)
			if classInst, ok := classObj.(*objects.Class); ok {
				if metaclass, ok := metaclassObj.(*objects.Class); ok {
					classInst.Metaclass = metaclass
				}
			}
			rvm.regSet(dst, classObj)

		case RegOpCallMetaclassInit:
			dst := inst.Operands[0]
			classReg := inst.Operands[1]
			classObj := rvm.regGet(classReg)
			if classInst, ok := classObj.(*objects.Class); ok {
				if classInst.Metaclass != nil {
					if initMethod, ok := classInst.Metaclass.Methods["__init__"]; ok {
						if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
							vm.push(nil)
							vm.push(classInst.Metaclass)
							vm.push(classInst)
							vm.push(&objects.String{Value: classInst.Name})
							bases := &objects.List{Elements: make([]objects.Object, len(classInst.SuperClasses))}
							for i, sc := range classInst.SuperClasses {
								bases.Elements[i] = sc
							}
							vm.push(bases)
							bp := vm.sp - 4
							newFrame := NewFrame(fn, bp)
							newFrame.metaclassInitClass = classInst
							vm.pushFrame(newFrame)
							vm.sp = newFrame.basePointer + fn.NumLocals
							// Need to continue execution in stack mode for this frame
							return nil
						}
					}
				}
				rvm.regSet(dst, classObj)
			} else {
				rvm.regSet(dst, classObj)
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
					rvm.regSet(dst, objects.None_)
				}
			} else {
				rvm.regSet(dst, val)
			}

		case RegOpYieldValue:
			valReg := inst.Operands[0]
			yieldValue := rvm.regGet(valReg)

			genIndex := frame.basePointer - 1
			if gen, ok := vm.stack[genIndex].(*objects.Generator); ok {
				gen.IP = frame.ip
				gen.StackPtr = vm.sp
				gen.BasePointer = frame.basePointer
				copy(gen.Locals, vm.stack[frame.basePointer:])
				copy(gen.Stack[:vm.sp], vm.stack[:vm.sp])
				vm.sp = genIndex
				vm.popFrame()
				vm.push(yieldValue)
			}
			return nil

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

		case RegOpRaise:
			errReg := inst.Operands[0]
			errObj := rvm.regGet(errReg)
			caught := vm.raiseException(errObj)
			if !caught {
				vm.pendingError = errObj
				return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
			}
			return nil

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
			// Unknown register opcode - this shouldn't happen
			return fmt.Errorf("unknown register opcode: %d", inst.Opcode)
		}

		ip++
	}

	// Sync registers back to stack before returning
	for i := 0; i < rvm.numRegs && i < StackSize; i++ {
		vm.stack[i] = rvm.registers[i]
	}
	if rvm.numRegs > vm.sp {
		vm.sp = rvm.numRegs
	}

	return nil
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
