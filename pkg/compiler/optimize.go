package compiler

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
		OpFormatString,
		OpFinally:
		return 3

	case OpClosure:
		return 4

	case OpBeginTry,
		OpExceptHandler:
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

func isJump(op Opcode) bool {
	return op == OpJump || op == OpJumpNotTruthy
}

func readOperand(ins Instructions, pos int, op Opcode) (int, int) {
	switch op {
	case OpConstant, OpJump, OpJumpNotTruthy, OpSetGlobal, OpGetGlobal,
		OpArray, OpHash, OpSet, OpCreateClass, OpCreateClassWithSuper,
		OpCreateClassWithMultiSuper, OpGetAttribute, OpSetAttribute,
		OpFormatString, OpFinally:
		return int(uint16(ins[pos+1])<<8 | uint16(ins[pos+2])), 2

	case OpClosure:
		return int(uint16(ins[pos+1])<<8 | uint16(ins[pos+2])), 3

	case OpBeginTry, OpExceptHandler:
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
	jumpTargets := make(map[int]bool)

	pos := 0
	for pos < len(ins) {
		op := Opcode(ins[pos])
		size := instructionSize(op)

		if op == OpJump || op == OpJumpNotTruthy {
			target, _ := readOperand(ins, pos, op)
			jumpTargets[target] = true
		}

		pos += size
	}

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

		case op == OpReturnValue || op == OpReturn || op == OpRaise:
			for t := range jumpTargets {
				if t > current && t < len(ins) && !reachable[t] {
					reachable[t] = true
					queue = append(queue, t)
				}
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
		size := instructionSize(op)

		if reachable[oldPos] {
			posMap[oldPos] = newPos
			newPos += size
		} else {
			posMap[oldPos] = newPos
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
				oldTarget, operandSize := readOperand(ins, oldPos, op)
				if newTarget, ok := posMap[oldTarget]; ok {
					chunk[1] = byte(newTarget >> 8)
					chunk[2] = byte(newTarget & 0xFF)
				} else {
					chunk[1] = byte(operandSize)
					_ = operandSize
				}
			}

			result = append(result, chunk...)
		}

		oldPos += size
	}

	return result
}
