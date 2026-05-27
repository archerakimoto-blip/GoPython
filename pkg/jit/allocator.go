package jit

import (
	"fmt"
)

type RegisterAllocator struct {
	availableRegs []int
	usedRegs     map[int]int
	varToReg     map[string]int
	regToVar     map[int]string
	constantPool map[int64]int
	stackSlot    int
	maxStack     int
}

type LiveRange struct {
	Start int
	End   int
	Var   string
}

func NewRegisterAllocator(numRegs int) *RegisterAllocator {
	availableRegs := make([]int, numRegs)
	for i := 0; i < numRegs; i++ {
		availableRegs[i] = i
	}
	
	return &RegisterAllocator{
		availableRegs: availableRegs,
		usedRegs:      make(map[int]int),
		varToReg:      make(map[string]int),
		regToVar:      make(map[int]string),
		constantPool:  make(map[int64]int),
		stackSlot:     0,
		maxStack:      0,
	}
}

func (ra *RegisterAllocator) Allocate(varName string) (int, error) {
	if reg, ok := ra.varToReg[varName]; ok {
		return reg, nil
	}
	
	if len(ra.availableRegs) == 0 {
		return -1, fmt.Errorf("no available registers")
	}
	
	reg := ra.availableRegs[0]
	ra.availableRegs = ra.availableRegs[1:]
	ra.usedRegs[reg] = 1
	ra.varToReg[varName] = reg
	ra.regToVar[reg] = varName
	
	return reg, nil
}

func (ra *RegisterAllocator) Free(reg int) {
	if varName, ok := ra.regToVar[reg]; ok {
		delete(ra.varToReg, varName)
		delete(ra.regToVar, reg)
	}
	
	delete(ra.usedRegs, reg)
	ra.availableRegs = append([]int{reg}, ra.availableRegs...)
}

func (ra *RegisterAllocator) AllocateSpecific(reg int, varName string) error {
	if _, used := ra.usedRegs[reg]; used {
		return fmt.Errorf("register %d is already in use", reg)
	}
	
	ra.varToReg[varName] = reg
	ra.regToVar[reg] = varName
	ra.usedRegs[reg] = 1
	
	for i, r := range ra.availableRegs {
		if r == reg {
			ra.availableRegs = append(ra.availableRegs[:i], ra.availableRegs[i+1:]...)
			break
		}
	}
	
	return nil
}

func (ra *RegisterAllocator) GetRegister(varName string) (int, bool) {
	reg, ok := ra.varToReg[varName]
	return reg, ok
}

func (ra *RegisterAllocator) SpillToStack(reg int) (int, error) {
	slot := ra.stackSlot
	ra.stackSlot++
	ra.stackSlot++ 
	
	ra.Free(reg)
	
	return slot, nil
}

func (ra *RegisterAllocator) RestoreFromStack(slot int, reg int) error {
	return ra.AllocateSpecific(reg, fmt.Sprintf("spilled_%d", slot))
}

func (ra *RegisterAllocator) AddConstant(value int64) int {
	if idx, ok := ra.constantPool[value]; ok {
		return idx
	}
	
	idx := len(ra.constantPool)
	ra.constantPool[value] = idx
	return idx
}

func (ra *RegisterAllocator) GetConstantPool() map[int64]int {
	return ra.constantPool
}

func (ra *RegisterAllocator) Reset() {
	ra.availableRegs = []int{Rax, Rcx, Rdx, Rbx, Rsi, Rdi, R8, R9, R10, R11, R12, R13, R14, R15}
	ra.usedRegs = make(map[int]int)
	ra.varToReg = make(map[string]int)
	ra.regToVar = make(map[int]string)
	ra.constantPool = make(map[int64]int)
	ra.stackSlot = 0
}

func (ra *RegisterAllocator) GetMaxStack() int {
	return ra.maxStack
}

func (ra *RegisterAllocator) UpdateMaxStack(depth int) {
	if depth > ra.maxStack {
		ra.maxStack = depth
	}
}

type LinearScanAllocator struct {
	allocator  *RegisterAllocator
	liveRanges []LiveRange
	numRegs    int
}

func NewLinearScanAllocator(numRegs int) *LinearScanAllocator {
	return &LinearScanAllocator{
		allocator:  NewRegisterAllocator(numRegs),
		liveRanges: make([]LiveRange, 0),
		numRegs:    numRegs,
	}
}

func (lsa *LinearScanAllocator) AddLiveRange(varName string, start, end int) {
	lsa.liveRanges = append(lsa.liveRanges, LiveRange{
		Start: start,
		End:   end,
		Var:   varName,
	})
}

func (lsa *LinearScanAllocator) Allocate() error {
	sortedRanges := make([]LiveRange, len(lsa.liveRanges))
	copy(sortedRanges, lsa.liveRanges)
	sortByStart(sortedRanges)
	
	active := make([]LiveRange, 0)
	
	for _, range_ := range sortedRanges {
		lsa.expireOldRanges(range_, &active)
		
		if len(active) >= lsa.numRegs {
			spillRange := lsa.findSpillRange(active)
			if spillRange.End > range_.End {
				if reg, ok := lsa.allocator.varToReg[spillRange.Var]; ok {
					lsa.allocator.Free(reg)
					lsa.allocator.AllocateSpecific(reg, range_.Var)
					active = lsa.updateActive(spillRange, range_, active)
				} else {
					lsa.allocator.Allocate(range_.Var)
				}
			} else {
				lsa.allocator.Allocate(range_.Var)
			}
		} else {
			_, err := lsa.allocator.Allocate(range_.Var)
			if err != nil {
				return err
			}
		}
		
		active = append(active, range_)
	}
	
	return nil
}

func (lsa *LinearScanAllocator) expireOldRanges(current LiveRange, active *[]LiveRange) {
	newActive := make([]LiveRange, 0)
	
	for _, range_ := range *active {
		if range_.End > current.Start {
			newActive = append(newActive, range_)
		} else {
			if reg, ok := lsa.allocator.GetRegister(range_.Var); ok {
				lsa.allocator.Free(reg)
			}
		}
	}
	
	*active = newActive
}

func (lsa *LinearScanAllocator) findSpillRange(active []LiveRange) LiveRange {
	spill := active[0]
	for i := 1; i < len(active); i++ {
		if active[i].End > spill.End {
			spill = active[i]
		}
	}
	return spill
}

func (lsa *LinearScanAllocator) updateActive(oldRange, newRange LiveRange, active []LiveRange) []LiveRange {
	result := make([]LiveRange, 0, len(active))
	for _, range_ := range active {
		if range_.Var != oldRange.Var {
			result = append(result, range_)
		}
	}
	result = append(result, newRange)
	return result
}

func sortByStart(ranges []LiveRange) {
	n := len(ranges)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if ranges[i].Start > ranges[j].Start {
				ranges[i], ranges[j] = ranges[j], ranges[i]
			}
		}
	}
}

type GraphColorAllocator struct {
	allocator   *RegisterAllocator
	interference map[string]map[string]bool
	colors      map[string]int
	numRegs     int
}

func NewGraphColorAllocator(numRegs int) *GraphColorAllocator {
	return &GraphColorAllocator{
		allocator:   NewRegisterAllocator(numRegs),
		interference: make(map[string]map[string]bool),
		colors:      make(map[string]int),
		numRegs:     numRegs,
	}
}

func (gca *GraphColorAllocator) AddInterference(var1, var2 string) {
	if _, ok := gca.interference[var1]; !ok {
		gca.interference[var1] = make(map[string]bool)
	}
	if _, ok := gca.interference[var2]; !ok {
		gca.interference[var2] = make(map[string]bool)
	}
	
	gca.interference[var1][var2] = true
	gca.interference[var2][var1] = true
}

func (gca *GraphColorAllocator) Color() error {
	simplifyStack := make([]string, 0)
	simplified := make(map[string]bool)
	
	for len(gca.colors) < len(gca.interference) {
		spillable := gca.getSpillableNodes(simplified)
		if len(spillable) == 0 {
			break
		}
		
		node := spillable[0]
		simplifyStack = append(simplifyStack, node)
		simplified[node] = true
	}
	
	for len(simplifyStack) > 0 {
		node := simplifyStack[len(simplifyStack)-1]
		simplifyStack = simplifyStack[:len(simplifyStack)-1]
		
		usedColors := make(map[int]bool)
		if neighbors, ok := gca.interference[node]; ok {
			for neighbor := range neighbors {
				if color, hasColor := gca.colors[neighbor]; hasColor {
					usedColors[color] = true
				}
			}
		}
		
		for color := 0; color < gca.numRegs; color++ {
			if !usedColors[color] {
				gca.colors[node] = color
				break
			}
		}
	}
	
	for varName, color := range gca.colors {
		gca.allocator.AllocateSpecific(color, varName)
	}
	
	return nil
}

func (gca *GraphColorAllocator) getSpillableNodes(simplified map[string]bool) []string {
	spillable := make([]string, 0)
	
	for node := range gca.interference {
		if _, hasColor := gca.colors[node]; hasColor {
			continue
		}
		if simplified[node] {
			continue
		}
		
		activeNeighbors := 0
		if neighbors, ok := gca.interference[node]; ok {
			for neighbor := range neighbors {
				if _, hasColor := gca.colors[neighbor]; !hasColor && !simplified[neighbor] {
					activeNeighbors++
				}
			}
		}
		
		if activeNeighbors < gca.numRegs {
			spillable = append(spillable, node)
		}
	}
	
	return spillable
}

func (gca *GraphColorAllocator) GetAllocator() *RegisterAllocator {
	return gca.allocator
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
