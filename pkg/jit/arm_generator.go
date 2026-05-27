package jit

import (
	"fmt"
	"unsafe"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

const (
	X0 = iota
	X1
	X2
	X3
	X4
	X5
	X6
	X7
	X8
	X9
	X10
	X11
	X12
	X13
	X14
	X15
	X16
	X17
	X18
	X19
	X20
	X21
	X22
	X23
	X24
	X25
	X26
	X27
	X28
	X29
	X30
	SP = 31
)

var ARMRegisterNames = []string{
	"x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7",
	"x8", "x9", "x10", "x11", "x12", "x13", "x14", "x15",
	"x16", "x17", "x18", "x19", "x20", "x21", "x22", "x23",
	"x24", "x25", "x26", "x27", "x28", "x29", "x30", "sp",
}

type ARMMachineCodeGenerator struct {
	buffer        []byte
	functions     map[string]*JITFunction
	constants     []objects.Object
	numLocals     int
	maxStackDepth int
}

func NewARMMachineCodeGenerator() *ARMMachineCodeGenerator {
	return &ARMMachineCodeGenerator{
		buffer:    make([]byte, 0, 1024),
		functions: make(map[string]*JITFunction),
	}
}

func (g *ARMMachineCodeGenerator) Generate(code *compiler.CompiledFunction) (*JITFunction, error) {
	g.buffer = g.buffer[:0]
	g.constants = code.Constants
	g.numLocals = 0
	g.maxStackDepth = 0

	g.emitPrologue(code)

	if err := g.emitBytecode(code.Instructions); err != nil {
		return nil, err
	}

	g.emitEpilogue()

	machineCode := make([]byte, len(g.buffer))
	copy(machineCode, g.buffer)

	jitFunc := &JITFunction{
		MachineCode: machineCode,
		FrameSize:   (g.numLocals + g.maxStackDepth + 1) * 8,
		NumParams:   code.NumParameters,
	}

	execMem, err := generateExecutableMemory(machineCode)
	if err != nil {
		return nil, err
	}
	jitFunc.EntryPoint = execMem

	return jitFunc, nil
}

func (g *ARMMachineCodeGenerator) emitPrologue(code *compiler.CompiledFunction) {
	g.emitStrX(X29, SP, -16)
	g.emitStrX(X30, SP, -8)
	g.emitAddImm(SP, SP, -32)
	g.emitMovX(X29, SP)
}

func (g *ARMMachineCodeGenerator) emitEpilogue() {
	g.emitAddImm(SP, SP, 32)
	g.emitLdrX(X29, SP, -16)
	g.emitLdrX(X30, SP, -8)
	g.emitRet()
}

func (g *ARMMachineCodeGenerator) emitBytes(bytes ...byte) {
	g.buffer = append(g.buffer, bytes...)
}

func (g *ARMMachineCodeGenerator) emitBytecode(instructions []byte) error {
	ip := 0
	for ip < len(instructions) {
		op := compiler.Opcode(instructions[ip])

		switch op {
		case compiler.OpConstant:
			if ip+2 < len(instructions) {
				constIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitLoadConstant(constIndex)
				ip += 3
			} else {
				return fmt.Errorf("invalid constant instruction at %d", ip)
			}

		case compiler.OpAdd:
			g.emitAdd()
			ip++

		case compiler.OpSub:
			g.emitSub()
			ip++

		case compiler.OpMul:
			g.emitMul()
			ip++

		case compiler.OpDiv:
			g.emitDiv()
			ip++

		case compiler.OpEqual:
			g.emitCompareEqual()
			ip++

		case compiler.OpNotEqual:
			g.emitCompareNotEqual()
			ip++

		case compiler.OpGreaterThan:
			g.emitCompareGreaterThan()
			ip++

		case compiler.OpLessThan:
			g.emitCompareLessThan()
			ip++

		case compiler.OpPop:
			g.emitPop(X0)
			ip++

		case compiler.OpDupTop:
			g.emitDup()
			ip++

		case compiler.OpReturn:
			ip++

		case compiler.OpReturnValue:
			g.emitReturnValue()
			ip++

		case compiler.OpJump:
			if ip+2 < len(instructions) {
				target := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitJmp(target - ip - 3)
				ip += 3
			}

		case compiler.OpJumpNotTruthy:
			if ip+2 < len(instructions) {
				target := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitJumpNotTruthy(target - ip - 3)
				ip += 3
			}

		case compiler.OpGetGlobal:
			if ip+2 < len(instructions) {
				globalIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitGetGlobal(globalIndex)
				ip += 3
			}

		case compiler.OpSetGlobal:
			if ip+2 < len(instructions) {
				globalIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitSetGlobal(globalIndex)
				ip += 3
			}

		case compiler.OpGetLocal:
			if ip+1 < len(instructions) {
				localIndex := int(instructions[ip+1])
				g.emitGetLocal(localIndex)
				ip += 2
			}

		case compiler.OpSetLocal:
			if ip+1 < len(instructions) {
				localIndex := int(instructions[ip+1])
				g.emitSetLocal(localIndex)
				ip += 2
			}

		case compiler.OpCall:
			if ip+1 < len(instructions) {
				numArgs := int(instructions[ip+1])
				g.emitCall(numArgs)
				ip += 2
			}

		case compiler.OpTrue:
			g.emitLoadTrue()
			ip++

		case compiler.OpFalse:
			g.emitLoadFalse()
			ip++

		case compiler.OpNull:
			g.emitLoadNull()
			ip++

		case compiler.OpMinus:
			g.emitNeg()
			ip++

		case compiler.OpBang:
			g.emitNot()
			ip++

		case compiler.OpArray:
			if ip+1 < len(instructions) {
				numElements := int(instructions[ip+1])
				g.emitCreateArray(numElements)
				ip += 2
			}

		case compiler.OpHash:
			if ip+1 < len(instructions) {
				numPairs := int(instructions[ip+1])
				g.emitCreateDict(numPairs)
				ip += 2
			}

		case compiler.OpSet:
			if ip+1 < len(instructions) {
				numElements := int(instructions[ip+1])
				g.emitCreateSet(numElements)
				ip += 2
			}

		case compiler.OpIndex:
			g.emitIndex()
			ip++

		case compiler.OpSlice:
			if ip+2 < len(instructions) {
				start := int(instructions[ip+1])
				end := int(instructions[ip+2])
				g.emitSlice(start, end)
				ip += 3
			}

		case compiler.OpGetAttribute:
			if ip+2 < len(instructions) {
				attrIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitGetAttribute(attrIndex)
				ip += 3
			}

		case compiler.OpSetAttribute:
			if ip+2 < len(instructions) {
				attrIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitSetAttribute(attrIndex)
				ip += 3
			}

		case compiler.OpFormatString:
			g.emitFormatString()
			ip++

		case compiler.OpCreateClass:
			g.emitCreateClass()
			ip++

		case compiler.OpCreateClassWithSuper:
			g.emitCreateClassWithSuper()
			ip++

		case compiler.OpYield:
			g.emitYield()
			ip++

		case compiler.OpBeginTry:
			if ip+2 < len(instructions) {
				handlerIP := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitBeginTry(handlerIP)
				ip += 3
			}

		case compiler.OpEndTry:
			g.emitEndTry()
			ip++

		case compiler.OpRaise:
			g.emitRaise()
			ip++

		default:
			ip++
		}

		g.maxStackDepth++
		if g.maxStackDepth > 10 {
			g.maxStackDepth = 10
		}
	}

	return nil
}

func (g *ARMMachineCodeGenerator) emitLoadConstant(index int) {
	g.emitMovImm64(X0, int64(uintptr(unsafe.Pointer(&g.constants[index]))))
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitAdd() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitAddReg(X0, X0, X1)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitSub() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitSubReg(X0, X0, X1)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitMul() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitMulReg(X0, X0, X1)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitDiv() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitDivReg(X0, X0, X1)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCompareEqual() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitCmpReg(X0, X1)
	g.emitCsetEq(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCompareNotEqual() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitCmpReg(X0, X1)
	g.emitCsetNe(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCompareGreaterThan() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitCmpReg(X0, X1)
	g.emitCsetGt(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCompareLessThan() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitCmpReg(X0, X1)
	g.emitCsetLt(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitPop(reg int) {
	g.emitLdrX(reg, SP, 0)
	g.emitAddImm(SP, SP, 8)
}

func (g *ARMMachineCodeGenerator) emitPush(reg int) {
	g.emitSubImm(SP, SP, 8)
	g.emitStrX(reg, SP, 0)
}

func (g *ARMMachineCodeGenerator) emitDup() {
	g.emitLdrX(X0, SP, 0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitJmp(offset int) {
	g.emitB(offset)
}

func (g *ARMMachineCodeGenerator) emitJumpNotTruthy(target int) {
	g.emitPop(X0)
	g.emitCmpImm(X0, 0)
	g.emitBne(target)
}

func (g *ARMMachineCodeGenerator) emitMovImm64(reg int, imm int64) {
	if imm >= -0x80000000 && imm <= 0x7FFFFFFF {
		g.emitMovImm32(reg, int32(imm))
	} else {
		g.emitMovz(uint64(reg), uint64(imm)&0xFFFF, 0)
		if imm >> 16 != 0 {
			g.emitMovk(uint64(reg), uint64(imm>>16)&0xFFFF, 16)
		}
		if imm >> 32 != 0 {
			g.emitMovk(uint64(reg), uint64(imm>>32)&0xFFFF, 32)
		}
		if imm >> 48 != 0 {
			g.emitMovk(uint64(reg), uint64(imm>>48)&0xFFFF, 48)
		}
	}
}

func (g *ARMMachineCodeGenerator) emitMovImm32(reg int, imm int32) {
	if imm >= 0 && imm <= 0xFF {
		g.emitMovz(uint64(reg), uint64(imm), 0)
	} else if imm >= -0x100000 && imm <= 0xFFFFF {
		g.emitMovn(uint64(reg), uint64(-imm), 0)
	} else {
		g.emitMovz(uint64(reg), uint64(imm)&0xFFFF, 0)
		if imm >> 16 != 0 {
			g.emitMovk(uint64(reg), uint64(imm>>16)&0xFFFF, 16)
		}
	}
}

func (g *ARMMachineCodeGenerator) emitMovz(rd, imm uint64, shift int) {
	opcode := uint32(0x52800000)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32((imm>>10)&0x1F) << 5
	opcode |= uint32(imm&0x3FF) << 10
	opcode |= (uint32(shift) / 16) << 21
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitMovk(rd, imm uint64, shift int) {
	opcode := uint32(0x52A00000)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32((imm>>10)&0x1F) << 5
	opcode |= uint32(imm&0x3FF) << 10
	opcode |= (uint32(shift) / 16) << 21
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitMovn(rd uint64, imm uint64, shift int) {
	opcode := uint32(0x53000000)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32((imm>>10)&0x1F) << 5
	opcode |= uint32(imm&0x3FF) << 10
	opcode |= (uint32(shift) / 16) << 21
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitAddReg(rd, rn, rm int) {
	opcode := uint32(0x8B000000)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32(rn&0x1F) << 5
	opcode |= uint32(rm&0x1F) << 16
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitAddImm(rd, rn int, imm int) {
	if imm >= 0 && imm <= 4095 {
		opcode := uint32(0x11000000)
		opcode |= uint32(rd&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32(imm&0xFFF) << 10
		g.emitUint32(opcode)
	} else if imm < 0 && imm >= -4096 {
		opcode := uint32(0x11200000)
		opcode |= uint32(rd&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32((-imm)&0xFFF) << 10
		g.emitUint32(opcode)
	}
}

func (g *ARMMachineCodeGenerator) emitSubReg(rd, rn, rm int) {
	opcode := uint32(0xCB000000)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32(rn&0x1F) << 5
	opcode |= uint32(rm&0x1F) << 16
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitSubImm(rd, rn int, imm int) {
	if imm >= 0 && imm <= 4095 {
		opcode := uint32(0x31000000)
		opcode |= uint32(rd&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32(imm&0xFFF) << 10
		g.emitUint32(opcode)
	} else if imm < 0 && imm >= -4096 {
		opcode := uint32(0x31200000)
		opcode |= uint32(rd&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32((-imm)&0xFFF) << 10
		g.emitUint32(opcode)
	}
}

func (g *ARMMachineCodeGenerator) emitMulReg(rd, rn, rm int) {
	opcode := uint32(0x8B000094)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32(rn&0x1F) << 5
	opcode |= uint32(rm&0x1F) << 16
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitDivReg(rd, rn, rm int) {
	opcode := uint32(0xCB000094)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32(rn&0x1F) << 5
	opcode |= uint32(rm&0x1F) << 16
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitCmpReg(rn, rm int) {
	opcode := uint32(0x6B000000)
	opcode |= uint32(rn&0x1F) << 5
	opcode |= uint32(rm&0x1F) << 16
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitCmpImm(rn int, imm int) {
	if imm >= 0 && imm <= 4095 {
		opcode := uint32(0x71000000)
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32(imm&0xFFF) << 10
		g.emitUint32(opcode)
	}
}

func (g *ARMMachineCodeGenerator) emitCsetEq(rd int) {
	opcode := uint32(0x5B000000 | (1 << 28))
	opcode |= uint32(rd&0x1F) << 0
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitCsetNe(rd int) {
	opcode := uint32(0x5B000010 | (1 << 28))
	opcode |= uint32(rd&0x1F) << 0
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitCsetGt(rd int) {
	opcode := uint32(0x5B000020 | (1 << 28))
	opcode |= uint32(rd&0x1F) << 0
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitCsetLt(rd int) {
	opcode := uint32(0x5B000040 | (1 << 28))
	opcode |= uint32(rd&0x1F) << 0
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitStrX(rt, rn int, imm int) {
	if imm >= 0 && imm <= 510 && imm%8 == 0 {
		opcode := uint32(0xF8000000)
		opcode |= uint32(rt&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32((imm/8)&0x3F) << 10
		g.emitUint32(opcode)
	} else if imm < 0 && imm >= -512 && imm%8 == 0 {
		opcode := uint32(0xF8200000)
		opcode |= uint32(rt&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32((-imm/8)&0x3F) << 10
		g.emitUint32(opcode)
	}
}

func (g *ARMMachineCodeGenerator) emitLdrX(rt, rn int, imm int) {
	if imm >= 0 && imm <= 510 && imm%8 == 0 {
		opcode := uint32(0xF8400000)
		opcode |= uint32(rt&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32((imm/8)&0x3F) << 10
		g.emitUint32(opcode)
	} else if imm < 0 && imm >= -512 && imm%8 == 0 {
		opcode := uint32(0xF8600000)
		opcode |= uint32(rt&0x1F) << 0
		opcode |= uint32(rn&0x1F) << 5
		opcode |= uint32((-imm/8)&0x3F) << 10
		g.emitUint32(opcode)
	}
}

func (g *ARMMachineCodeGenerator) emitB(offset int) {
	if offset >= -1048576 && offset <= 1048574 {
		imm := offset / 4
		opcode := uint32(0x14000000)
		opcode |= uint32(uint32(int32(imm)) & 0xFFFFFF)
		g.emitUint32(opcode)
	}
}

func (g *ARMMachineCodeGenerator) emitBne(offset int) {
	if offset >= -1048576 && offset <= 1048574 {
		imm := offset / 4
		opcode := uint32(0x54000001)
		opcode |= uint32(uint32(int32(imm)) & 0xFFFFFF)
		g.emitUint32(opcode)
	}
}

func (g *ARMMachineCodeGenerator) emitRet() {
	opcode := uint32(0xD65F03C0)
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitMovX(rd, rn int) {
	opcode := uint32(0xAA000000)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32(rn&0x1F) << 5
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitUint32(val uint32) {
	g.buffer = append(g.buffer,
		byte(val),
		byte(val>>8),
		byte(val>>16),
		byte(val>>24),
	)
}

func (g *ARMMachineCodeGenerator) emitReturnValue() {
	g.emitPop(X0)
	g.emitAddImm(SP, SP, 32)
	g.emitLdrX(X29, SP, -16)
	g.emitLdrX(X30, SP, -8)
	g.emitRet()
}

func (g *ARMMachineCodeGenerator) emitGetGlobal(index int) {
	g.emitMovImm64(X0, int64(index))
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitSetGlobal(index int) {
	g.emitPop(X0)
}

func (g *ARMMachineCodeGenerator) emitGetLocal(index int) {
	g.emitMovImm64(X0, int64(index))
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitSetLocal(index int) {
	g.emitPop(X0)
}

func (g *ARMMachineCodeGenerator) emitCall(numArgs int) {
	g.emitPop(X0)
	g.emitPop(X1)
	g.emitAddImm(SP, SP, numArgs*8)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitLoadTrue() {
	g.emitMovImm64(X0, 1)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitLoadFalse() {
	g.emitMovImm64(X0, 0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitLoadNull() {
	g.emitMovImm64(X0, 0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitNeg() {
	g.emitPop(X0)
	g.emitNegReg(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitNot() {
	g.emitPop(X0)
	g.emitCmpImm(X0, 0)
	g.emitCsetEq(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitNegReg(rd int) {
	opcode := uint32(0xCB000080)
	opcode |= uint32(rd&0x1F) << 0
	opcode |= uint32(rd&0x1F) << 5
	g.emitUint32(opcode)
}

func (g *ARMMachineCodeGenerator) emitCreateArray(numElements int) {
	g.emitMovImm64(X0, int64(numElements))
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCreateDict(numPairs int) {
	g.emitMovImm64(X0, int64(numPairs))
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCreateSet(numElements int) {
	g.emitMovImm64(X0, int64(numElements))
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitIndex() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitSlice(start, end int) {
	g.emitPop(X2)
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitGetAttribute(attrIndex int) {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitSetAttribute(attrIndex int) {
	g.emitPop(X2)
	g.emitPop(X1)
	g.emitPop(X0)
}

func (g *ARMMachineCodeGenerator) emitFormatString() {
	g.emitPop(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCreateClass() {
	g.emitPop(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitCreateClassWithSuper() {
	g.emitPop(X1)
	g.emitPop(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitYield() {
	g.emitPop(X0)
	g.emitPush(X0)
}

func (g *ARMMachineCodeGenerator) emitBeginTry(handlerIP int) {
}

func (g *ARMMachineCodeGenerator) emitEndTry() {
}

func (g *ARMMachineCodeGenerator) emitRaise() {
	g.emitPop(X0)
}

func (g *ARMMachineCodeGenerator) GetGeneratedCode() []byte {
	return g.buffer
}

func (g *ARMMachineCodeGenerator) GetCodeSize() int {
	return len(g.buffer)
}
