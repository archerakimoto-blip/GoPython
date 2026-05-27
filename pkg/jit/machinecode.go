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
			
		case compiler.OpSetGlobal:
			if ip+2 < len(instructions) {
				g.emitSetGlobal()
				ip += 3
			}
			
		case compiler.OpGetGlobal:
			if ip+2 < len(instructions) {
				g.emitGetGlobal()
				ip += 3
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

func (g *MachineCodeGenerator) emitSetGlobal() {
	g.emitPopR64(Rax)
	g.emitPopR64(Rcx)
}

func (g *MachineCodeGenerator) emitGetGlobal() {
	g.emitPopR64(Rax)
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
	
	mem := make([]byte, len(code))
	copy(mem, code)
	
	for i, b := range mem {
		writeExecByte(execMem, i, b)
	}
	
	flushInstructionCache(execMem, len(code))
	
	return execMem, nil
}

func writeExecByte(addr uintptr, offset int, b byte) {
}

func allocateRWXMemory(size int) (uintptr, error) {
	mem := make([]byte, size)
	return uintptr(unsafe.Pointer(&mem[0])), nil
}

func flushInstructionCache(addr uintptr, size int) {
}

type executableMem struct {
	addr uintptr
	size int
}

func (g *MachineCodeGenerator) GetGeneratedCode() []byte {
	return g.buffer
}

func (g *MachineCodeGenerator) GetCodeSize() int {
	return len(g.buffer)
}
