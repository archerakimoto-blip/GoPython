package compiler

import "github.com/go-py/go-python/pkg/objects"

func instructionSize(op Opcode) int {
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

	case OpExceptHandler:
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

	case OpExceptHandler:
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
		size := instructionSize(op)
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

		case op == OpReturnValue || op == OpReturn || op == OpRaise:

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
		size := instructionSize(op)

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
		size := instructionSize(op)

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
