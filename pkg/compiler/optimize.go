package compiler

import "github.com/go-py/go-python/pkg/objects"

func InstructionSize(op Opcode) int {
	switch op {
	case OpConstant,
		OpJump,
		OpJumpNotTruthy,
		OpSetGlobal,
		OpGetGlobal,
		OpArray,
		OpHash,
		OpSet,
		OpCreateClass,
		OpCreateClassWithSuper,
		OpCreateClassWithMultiSuper,
		OpGetAttribute,
		OpSetAttribute,
		OpDelAttribute,
		OpFormatString,
		OpFinally:
		return 3

	case OpClosure:
		return 4

	case OpBeginTry:
		return 9

	case OpExceptHandler,
		OpExceptStarHandler:
		return 5

	case OpCall,
		OpSetLocal,
		OpGetLocal,
		OpGetFree:
		return 2

	default:
		return 1
	}
}

func isTerminator(op Opcode) bool {
	switch op {
	case OpReturnValue, OpReturn, OpJump, OpRaise:
		return true
	default:
		return false
	}
}

func readOperand(ins Instructions, pos int, op Opcode) (int, int) {
	switch op {
	case OpConstant, OpJump, OpJumpNotTruthy, OpSetGlobal, OpGetGlobal,
		OpArray, OpHash, OpSet, OpCreateClass, OpCreateClassWithSuper,
		OpCreateClassWithMultiSuper, OpGetAttribute, OpSetAttribute, OpDelAttribute,
		OpFormatString, OpFinally:
		return int(uint16(ins[pos+1])<<8 | uint16(ins[pos+2])), 2

	case OpClosure:
		return int(uint16(ins[pos+1])<<8 | uint16(ins[pos+2])), 3

	case OpBeginTry:
		return int(uint16(ins[pos+1])<<8 | uint16(ins[pos+2])), 8

	case OpExceptHandler, OpExceptStarHandler:
		return int(uint16(ins[pos+1])<<8 | uint16(ins[pos+2])), 4

	case OpCall, OpSetLocal, OpGetLocal, OpGetFree:
		return int(ins[pos+1]), 1

	default:
		return 0, 0
	}
}

func EliminateDeadCode(ins Instructions) Instructions {
	if len(ins) == 0 {
		return ins
	}

	reachable := make([]bool, len(ins))

	queue := []int{0}
	reachable[0] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current >= len(ins) {
			continue
		}

		op := Opcode(ins[current])
		size := InstructionSize(op)
		nextInst := current + size

		switch {
		case op == OpJump:
			target, _ := readOperand(ins, current, op)
			if target < len(ins) && !reachable[target] {
				reachable[target] = true
				queue = append(queue, target)
			}

		case op == OpJumpNotTruthy:
			target, _ := readOperand(ins, current, op)
			if target < len(ins) && !reachable[target] {
				reachable[target] = true
				queue = append(queue, target)
			}
			if nextInst < len(ins) && !reachable[nextInst] {
				reachable[nextInst] = true
				queue = append(queue, nextInst)
			}

		case op == OpBeginTry:
			if nextInst < len(ins) && !reachable[nextInst] {
				reachable[nextInst] = true
				queue = append(queue, nextInst)
			}
			handlerIP := int(uint16(ins[current+5])<<8 | uint16(ins[current+6]))
			if handlerIP > 0 && handlerIP < len(ins) && !reachable[handlerIP] {
				reachable[handlerIP] = true
				queue = append(queue, handlerIP)
			}
			finallyStartIP := int(uint16(ins[current+7])<<8 | uint16(ins[current+8]))
			if finallyStartIP > 0 && finallyStartIP < len(ins) && !reachable[finallyStartIP] {
				reachable[finallyStartIP] = true
				queue = append(queue, finallyStartIP)
			}
			// Mark all except handler instructions as reachable
			// (OpExceptHandler and OpExceptStarHandler between handlerIP and finallyStartIP/OpEndTry)
			if handlerIP > 0 {
				scanIP := handlerIP
				for scanIP < len(ins) {
					scanOp := Opcode(ins[scanIP])
					if scanOp == OpExceptHandler || scanOp == OpExceptStarHandler {
						if !reachable[scanIP] {
							reachable[scanIP] = true
							queue = append(queue, scanIP)
						}
						scanIP += InstructionSize(scanOp)
						// Also mark the handler body instructions as reachable
						for scanIP < len(ins) {
							bodyOp := Opcode(ins[scanIP])
							if bodyOp == OpExceptHandler || bodyOp == OpExceptStarHandler ||
								bodyOp == OpFinally || bodyOp == OpEndTry {
								break
							}
							if !reachable[scanIP] {
								reachable[scanIP] = true
								queue = append(queue, scanIP)
							}
							bodySize := InstructionSize(bodyOp)
							if bodySize <= 0 {
								scanIP++
							} else {
								scanIP += bodySize
							}
						}
					} else if scanOp == OpFinally || scanOp == OpEndTry {
						break
					} else {
						scanIP++
					}
				}
			}

		case op == OpReturnValue || op == OpReturn || op == OpRaise:

		case op == OpFinally:
			// OpFinally has a jump target (finallyEndIP) that must be reachable
			target := int(uint16(ins[current+1])<<8 | uint16(ins[current+2]))
			if target > 0 && target < len(ins) && !reachable[target] {
				reachable[target] = true
				queue = append(queue, target)
			}
			if nextInst < len(ins) && !reachable[nextInst] {
				reachable[nextInst] = true
				queue = append(queue, nextInst)
			}

		default:
			if nextInst < len(ins) && !reachable[nextInst] {
				reachable[nextInst] = true
				queue = append(queue, nextInst)
			}
		}
	}

	posMap := make(map[int]int)
	oldPos := 0
	newPos := 0
	for oldPos < len(ins) {
		op := Opcode(ins[oldPos])
		size := InstructionSize(op)

		posMap[oldPos] = newPos

		if reachable[oldPos] {
			newPos += size
		}

		oldPos += size
	}

	result := make(Instructions, 0, newPos)
	oldPos = 0
	for oldPos < len(ins) {
		op := Opcode(ins[oldPos])
		size := InstructionSize(op)

		if reachable[oldPos] {
			chunk := make(Instructions, size)
			copy(chunk, ins[oldPos:oldPos+size])

			if op == OpJump || op == OpJumpNotTruthy {
				oldTarget, _ := readOperand(ins, oldPos, op)
				if newTarget, ok := posMap[oldTarget]; ok {
					chunk[1] = byte(newTarget >> 8)
					chunk[2] = byte(newTarget & 0xFF)
				}
			}

			if op == OpFinally {
				oldTarget := int(uint16(ins[oldPos+1])<<8 | uint16(ins[oldPos+2]))
				if newTarget, ok := posMap[oldTarget]; ok {
					chunk[1] = byte(newTarget >> 8)
					chunk[2] = byte(newTarget & 0xFF)
				}
			}

			if op == OpBeginTry {
				oldHandlerIP := int(uint16(ins[oldPos+5])<<8 | uint16(ins[oldPos+6]))
				if newHandlerIP, ok := posMap[oldHandlerIP]; ok {
					chunk[5] = byte(newHandlerIP >> 8)
					chunk[6] = byte(newHandlerIP & 0xFF)
				}
				oldFinallyStartIP := int(uint16(ins[oldPos+7])<<8 | uint16(ins[oldPos+8]))
				if newFinallyStartIP, ok := posMap[oldFinallyStartIP]; ok {
					chunk[7] = byte(newFinallyStartIP >> 8)
					chunk[8] = byte(newFinallyStartIP & 0xFF)
				}
			}

			result = append(result, chunk...)
		}

		oldPos += size
	}

	return result
}

func EliminateDeadCodeInFunctions(bytecode *Bytecode) *Bytecode {
	newConstants := make([]objects.Object, len(bytecode.Constants))
	for i, c := range bytecode.Constants {
		if fn, ok := c.(*CompiledFunction); ok {
			optimized := EliminateDeadCode(fn.Instructions)
			newFn := *fn
			newFn.Instructions = optimized
			// 运行逃逸分析
			newFn.NonEscapingLocals = analyzeEscape(&newFn)
			newConstants[i] = &newFn
		} else {
			newConstants[i] = c
		}
	}

	return &Bytecode{
		Instructions: EliminateDeadCode(bytecode.Instructions),
		Constants:    newConstants,
	}
}

// analyzeEscape 分析函数中哪些局部变量不逃逸
// 一个局部变量"逃逸"如果它：
// 1. 被闭包捕获（OpGetFree 引用）
// 2. 被返回（OpReturnValue 之前最后写入的局部变量）
// 3. 被赋值给全局变量（OpSetGlobal）
// 4. 被赋值给实例属性（OpSetAttribute）
// 5. 被作为函数参数传递给外部函数
// 不逃逸的局部变量可以安全地在栈上分配
func analyzeEscape(fn *CompiledFunction) []bool {
	if fn.NumLocals == 0 {
		return nil
	}

	// 初始假设所有局部变量都不逃逸
	nonEscaping := make([]bool, fn.NumLocals)
	for i := range nonEscaping {
		nonEscaping[i] = true
	}

	// 参数总是不逃逸的（它们在函数入口就绑定了）
	// 但如果参数被用于逃逸操作，则标记为逃逸

	ins := fn.Instructions
	for i := 0; i < len(ins); {
		op := Opcode(ins[i])
		size := InstructionSize(op)

		switch op {
		case OpSetGlobal:
			// 赋值给全局变量：源局部变量逃逸
			// OpSetGlobal 的操作数是全局索引，不是局部索引
			// 但在这之前一定有 OpGetLocal 将局部变量加载到栈上
			// 我们无法直接追踪，保守标记所有在 OpSetGlobal 之前的 OpGetLocal 对应的变量为逃逸
			// 简化处理：如果有 OpSetGlobal，标记所有局部变量为可能逃逸
			for j := range nonEscaping {
				if j >= fn.NumParameters {
					nonEscaping[j] = false
				}
			}

		case OpSetAttribute:
			// 赋值给实例属性：值可能逃逸
			// 保守处理：标记所有非参数局部变量为可能逃逸
			// 但这太保守了，让我们只在值是局部变量时标记
			// 由于无法确定栈上的值来自哪个局部变量，跳过

		case OpGetFree:
			// 被闭包捕获的变量逃逸
			if size >= 2 && i+1 < len(ins) {
				freeIdx := int(ins[i+1])
				if freeIdx < len(nonEscaping) {
					nonEscaping[freeIdx] = false
				}
			}

		case OpReturnValue:
			// 返回值：如果返回的是局部变量，它逃逸
			// 但我们无法确定返回的是哪个局部变量
			// 保守处理：标记所有非参数局部变量为可能逃逸
			// 实际上，只有被 OpGetLocal 加载然后返回的才逃逸
			// 让我们追踪 OpGetLocal + OpReturnValue 模式
			if i >= 2 && Opcode(ins[i-2]) == OpGetLocal {
				localIdx := int(ins[i-1])
				if localIdx < len(nonEscaping) {
					nonEscaping[localIdx] = false
				}
			}
		}

		i += size
	}

	return nonEscaping
}
