package jit

import (
	"fmt"
	"unsafe"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

type MachineCodeGenerator struct {
	buffer     []byte
	functions  map[string]*JITFunction
	constants  []objects.Object
	numLocals  int
	maxStackDepth int
}

type JITFunction struct {
	MachineCode []byte
	EntryPoint  uintptr
	FrameSize   int
	NumParams   int
}

const (
	Rax = iota
	Rcx
	Rdx
	Rbx
	Rsp
	Rbp
	Rsi
	Rdi
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
)

var RegisterNames = []string{
	"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi",
	"r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15",
}

func NewMachineCodeGenerator() *MachineCodeGenerator {
	return &MachineCodeGenerator{
		buffer:    make([]byte, 0, 1024),
		functions: make(map[string]*JITFunction),
	}
}

func (g *MachineCodeGenerator) Generate(code *compiler.CompiledFunction) (*JITFunction, error) {
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

func (g *MachineCodeGenerator) emitPrologue(code *compiler.CompiledFunction) {
	g.emitBytes(0x55)
	g.emitBytes(0x48, 0x89, 0xE5)
	g.emitSubRsp(16)
}

func (g *MachineCodeGenerator) emitEpilogue() {
	g.emitAddRsp(16)
	g.emitBytes(0x5D)
	g.emitBytes(0xC3)
}

func (g *MachineCodeGenerator) emitBytes(bytes ...byte) {
	g.buffer = append(g.buffer, bytes...)
}

func (g *MachineCodeGenerator) emitMovR64Imm64(reg int, imm int64) {
	g.emitBytes(0x48)
	g.emitBytes(0xB8 + byte(reg&0x7))
	g.emitInt64(imm)
}

func (g *MachineCodeGenerator) emitMovR64R64(dst, src int) {
	g.emitBytes(0x48)
	g.emitBytes(0x89)
	g.emitModRM(3, dst&0x7, src&0x7)
}

func (g *MachineCodeGenerator) emitAddR64R64(dst, src int) {
	g.emitBytes(0x48)
	g.emitBytes(0x01)
	g.emitModRM(3, dst&0x7, src&0x7)
}

func (g *MachineCodeGenerator) emitSubR64R64(dst, src int) {
	g.emitBytes(0x48)
	g.emitBytes(0x29)
	g.emitModRM(3, dst&0x7, src&0x7)
}

func (g *MachineCodeGenerator) emitAddRsp(imm byte) {
	g.emitBytes(0x48)
	g.emitBytes(0x83)
	g.emitBytes(0xC4)
	g.emitBytes(imm)
}

func (g *MachineCodeGenerator) emitSubRsp(imm byte) {
	g.emitBytes(0x48)
	g.emitBytes(0x83)
	g.emitBytes(0xEC)
	g.emitBytes(imm)
}

func (g *MachineCodeGenerator) emitMulR64(reg int) {
	g.emitBytes(0x48)
	g.emitBytes(0xF7)
	g.emitBytes(0xE0 + byte(reg&0x7))
}

func (g *MachineCodeGenerator) emitDivR64(reg int) {
	g.emitBytes(0x48)
	g.emitBytes(0xF7)
	g.emitBytes(0xF0 + byte(reg&0x7))
}

func (g *MachineCodeGenerator) emitCmpR64R64(r1, r2 int) {
	g.emitBytes(0x48)
	g.emitBytes(0x39)
	g.emitModRM(3, r1&0x7, r2&0x7)
}

func (g *MachineCodeGenerator) emitSetCond(cond byte, reg int) {
	g.emitBytes(0x0F)
	g.emitBytes(0x90 + cond)
	g.emitBytes(0xC0 + byte(reg&0x7))
}

func (g *MachineCodeGenerator) emitJmpRel32(offset int32) {
	g.emitBytes(0xE9)
	g.emitInt32(offset)
}

func (g *MachineCodeGenerator) emitJccRel32(cond byte, offset int32) {
	g.emitBytes(0x0F)
	g.emitBytes(0x80 + cond)
	g.emitInt32(offset)
}

func (g *MachineCodeGenerator) emitRet() {
	g.emitBytes(0xC3)
}

func (g *MachineCodeGenerator) emitCallRel32(offset int32) {
	g.emitBytes(0xE8)
	g.emitInt32(offset)
}

func (g *MachineCodeGenerator) emitPushR64(reg int) {
	g.emitBytes(0x50 + byte(reg&0x7))
}

func (g *MachineCodeGenerator) emitPopR64(reg int) {
	g.emitBytes(0x58 + byte(reg&0x7))
}

func (g *MachineCodeGenerator) emitMovR64Mem64(reg, base, offset int) {
	g.emitBytes(0x48)
	g.emitBytes(0x8B)
	g.emitModRMDisp8(0, reg&0x7, base&0x7)
	g.emitBytes(byte(offset))
}

func (g *MachineCodeGenerator) emitMovMem64R64(base, offset, reg int) {
	g.emitBytes(0x48)
	g.emitBytes(0x89)
	g.emitModRMDisp8(reg&0x7, 0, base&0x7)
	g.emitBytes(byte(offset))
}

func (g *MachineCodeGenerator) emitModRM(mod, reg, rm int) {
	g.emitBytes(byte((mod << 6) | ((reg & 7) << 3) | (rm & 7)))
}

func (g *MachineCodeGenerator) emitModRMDisp8(mod, reg, rm int) {
	g.emitBytes(byte((mod << 6) | ((reg & 7) << 3) | (rm & 7)))
}

func (g *MachineCodeGenerator) emitInt64(val int64) {
	g.buffer = append(g.buffer,
		byte(val),
		byte(val>>8),
		byte(val>>16),
		byte(val>>24),
		byte(val>>32),
		byte(val>>40),
		byte(val>>48),
		byte(val>>56),
	)
}

func (g *MachineCodeGenerator) emitInt32(val int32) {
	g.buffer = append(g.buffer,
		byte(val),
		byte(val>>8),
		byte(val>>16),
		byte(val>>24),
	)
}

func (g *MachineCodeGenerator) emitBytecode(instructions []byte) error {
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
			g.emitPop()
			ip++
			
		case compiler.OpDupTop:
			g.emitDup()
			ip++
			
		case compiler.OpReturn:
			ip++
			
		case compiler.OpReturnValue:
			g.emitReturnValue()
			ip++
			
		case compiler.OpSetGlobal:
			if ip+2 < len(instructions) {
				globalIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitSetGlobal(globalIndex)
				ip += 3
			}
			
		case compiler.OpGetGlobal:
			if ip+2 < len(instructions) {
				globalIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				g.emitGetGlobal(globalIndex)
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
			g.emitComment("unsupported opcode: %d", op)
			ip++
		}
		
		g.maxStackDepth++
		if g.maxStackDepth > 10 {
			g.maxStackDepth = 10
		}
	}
	
	return nil
}

func (g *MachineCodeGenerator) emitLoadConstant(index int) {
	g.emitMovR64Imm64(Rax, int64(uintptr(unsafe.Pointer(&g.constants[index]))))
}

func (g *MachineCodeGenerator) emitAdd() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitAddR64R64(Rax, Rcx)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitSub() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitSubR64R64(Rax, Rcx)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitMul() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitMulR64(Rcx)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitDiv() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitDivR64(Rcx)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitMod() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitMovR64R64(Rdx, Rax)
	g.emitDivR64(Rcx)
	g.emitMovR64R64(Rax, Rdx)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitNeg() {
	g.emitPopR64(Rax)
	g.emitBytes(0x48, 0xF7, 0xD8)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCompareEqual() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitCmpR64R64(Rcx, Rax)
	g.emitSetCond(0x4, Rax)
	g.emitMovzxR64R8(Rax, Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCompareNotEqual() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitCmpR64R64(Rcx, Rax)
	g.emitSetCond(0x5, Rax)
	g.emitMovzxR64R8(Rax, Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCompareGreaterThan() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitCmpR64R64(Rcx, Rax)
	g.emitSetCond(0x7, Rax)
	g.emitMovzxR64R8(Rax, Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCompareLessThan() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitCmpR64R64(Rcx, Rax)
	g.emitSetCond(0x2, Rax)
	g.emitMovzxR64R8(Rax, Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitMovzxR64R8(dest, src int) {
	g.emitBytes(0x48, 0x0F, 0xB6)
	g.emitModRM(3, dest&0x7, src&0x7)
}

func (g *MachineCodeGenerator) emitPop() {
	g.emitAddRsp(8)
}

func (g *MachineCodeGenerator) emitDup() {
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
	g.emitPushR64(Rax)
}



func (g *MachineCodeGenerator) emitJmp(target int) {
	offset := int32(target - len(g.buffer) - 5)
	g.emitJmpRel32(offset)
}

func (g *MachineCodeGenerator) emitJumpNotTruthy(target int) {
	g.emitPopR64(Rax)
	g.emitCmpR64Imm8(Rax, 0)
	offset := int32(target - len(g.buffer) - 6)
	g.emitJccRel32(0x4, offset)
}

func (g *MachineCodeGenerator) emitCmpR64Imm8(reg int, imm byte) {
	g.emitBytes(0x48, 0x83)
	g.emitBytes(0xF8 + byte(reg&0x7))
	g.emitBytes(imm)
}

func (g *MachineCodeGenerator) emitComment(format string, args ...interface{}) {
}

func generateExecutableMemory(code []byte) (uintptr, error) {
	execMem, err := allocateRWXMemory(len(code))
	if err != nil {
		return 0, err
	}
	
	for i, b := range code {
		writeExecByte(execMem, i, b)
	}
	
	flushInstructionCache(execMem, len(code))
	
	return execMem, nil
}

func writeExecByte(addr uintptr, offset int, b byte) {
	ptr := unsafe.Pointer(uintptr(addr) + uintptr(offset))
	*(*byte)(ptr) = b
}

func flushInstructionCache(addr uintptr, size int) {
	// On x86/x64 architectures, the instruction cache is automatically coherent
	// with data cache, so no explicit flush is needed.
	// However, we can use CPU instructions to ensure cache coherency if needed.
	
	// For cross-platform compatibility, we use the appropriate instruction:
	// - x86/x64: CPUID instruction implicitly flushes instruction cache
	// - ARM: IC IALLUIS (implemented in machine code)
	
	// Since this is a Go package, we'll use a simple approach
	// The x86 architecture guarantees cache coherency between instruction and data caches
	// through snooping, so no explicit flush is typically needed.
	
	// If we need explicit flush, we can emit the following in generated code:
	// x86: MFENCE or CPUID
	// ARM64: ISB
	
	// For now, we assume the platform handles cache coherency automatically
}

type executableMem struct {
	addr uintptr
	size int
}

func (g *MachineCodeGenerator) emitReturnValue() {
	g.emitPopR64(Rax)
	g.emitAddRsp(16)
	g.emitBytes(0x5D)
	g.emitBytes(0xC3)
}

func (g *MachineCodeGenerator) emitGetGlobal(index int) {
	g.emitMovR64Imm64(Rax, int64(index))
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitSetGlobal(index int) {
	g.emitPopR64(Rax)
}

func (g *MachineCodeGenerator) emitGetLocal(index int) {
	g.emitMovR64Imm64(Rax, int64(index))
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitSetLocal(index int) {
	g.emitPopR64(Rax)
}

func (g *MachineCodeGenerator) emitCall(numArgs int) {
	g.emitPopR64(Rax)
	g.emitPopR64(Rcx)
	g.emitAddRsp(byte(numArgs * 8))
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitLoadTrue() {
	g.emitMovR64Imm64(Rax, 1)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitLoadFalse() {
	g.emitMovR64Imm64(Rax, 0)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitLoadNull() {
	g.emitMovR64Imm64(Rax, 0)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitNot() {
	g.emitPopR64(Rax)
	g.emitCmpR64Imm8(Rax, 0)
	g.emitSetCond(0x4, Rax)
	g.emitMovzxR64R8(Rax, Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCreateArray(numElements int) {
	g.emitMovR64Imm64(Rax, int64(numElements))
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCreateDict(numPairs int) {
	g.emitMovR64Imm64(Rax, int64(numPairs))
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCreateSet(numElements int) {
	g.emitMovR64Imm64(Rax, int64(numElements))
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitIndex() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitSlice(start, end int) {
	g.emitPopR64(Rdx)
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitGetAttribute(attrIndex int) {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitSetAttribute(attrIndex int) {
	g.emitPopR64(Rdx)
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
}

func (g *MachineCodeGenerator) emitFormatString() {
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCreateClass() {
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitCreateClassWithSuper() {
	g.emitPopR64(Rcx)
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitYield() {
	g.emitPopR64(Rax)
	g.emitPushR64(Rax)
}

func (g *MachineCodeGenerator) emitBeginTry(handlerIP int) {
}

func (g *MachineCodeGenerator) emitEndTry() {
}

func (g *MachineCodeGenerator) emitRaise() {
	g.emitPopR64(Rax)
}

func (g *MachineCodeGenerator) GetGeneratedCode() []byte {
	return g.buffer
}

func (g *MachineCodeGenerator) GetCodeSize() int {
	return len(g.buffer)
}
