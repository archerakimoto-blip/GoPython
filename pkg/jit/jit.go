package jit

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/compiler"
)

type JIT struct {
	cache           map[string]*CompiledCode
	mu              sync.RWMutex
	executionCounts map[string]int64
	hotThreshold    int64
	maxCacheSize    int
}

type CompiledCode struct {
	Function      *compiler.CompiledFunction
	MachineCode   []byte
	CallCount     int64
	CompileTime   time.Time
	Optimized     bool
	Metadata      *CodeMetadata
}

type CodeMetadata struct {
	NumInstructions int
	HasLoops       bool
	HasRecursion   bool
	Complexity     int
	HotPaths       []string
}

func New() *JIT {
	j := &JIT{
		cache:           make(map[string]*CompiledCode),
		executionCounts: make(map[string]int64),
		hotThreshold:    5,
		maxCacheSize:    100,
	}
	return j
}

func (j *JIT) Compile(node ast.Node) interface{} {
	return nil
}

func (j *JIT) ShouldCompile(fn *compiler.CompiledFunction) bool {
	if fn == nil {
		return false
	}
	
	key := j.getFunctionKey(fn)
	j.mu.RLock()
	count := j.executionCounts[key]
	j.mu.RUnlock()
	
	return count >= j.hotThreshold
}

func (j *JIT) GetOrCompile(fn *compiler.CompiledFunction) (*CompiledCode, error) {
	if fn == nil {
		return nil, fmt.Errorf("nil function")
	}

	key := j.getFunctionKey(fn)
	
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if compiled, ok := j.cache[key]; ok {
		compiled.CallCount++
		return compiled, nil
	}

	metadata := j.analyzeFunction(fn)
	compiled := &CompiledCode{
		Function:    fn,
		CallCount:   1,
		CompileTime: time.Now(),
		Metadata:    metadata,
	}

	j.executionCounts[key] = 1
	
	if len(j.cache) >= j.maxCacheSize {
		j.evictOldEntries()
	}
	
	j.cache[key] = compiled
	return compiled, nil
}

func (j *JIT) RecordCall(fn *compiler.CompiledFunction) {
	if fn == nil {
		return
	}
	
	key := j.getFunctionKey(fn)
	j.mu.Lock()
	defer j.mu.Unlock()
	
	j.executionCounts[key]++
	
	if compiled, ok := j.cache[key]; ok {
		compiled.CallCount++
		
		if compiled.CallCount >= j.hotThreshold && !compiled.Optimized {
			j.optimizeFunction(compiled)
		}
	}
}

func (j *JIT) getFunctionKey(fn *compiler.CompiledFunction) string {
	return fmt.Sprintf("%p:%d:%d", fn, len(fn.Instructions), fn.NumParameters)
}

func (j *JIT) analyzeFunction(fn *compiler.CompiledFunction) *CodeMetadata {
	metadata := &CodeMetadata{
		NumInstructions: len(fn.Instructions),
		HasLoops:       j.detectLoops(fn.Instructions),
		Complexity:     j.calculateComplexity(fn.Instructions),
	}
	
	return metadata
}

func (j *JIT) detectLoops(instructions []byte) bool {
	for i := 0; i < len(instructions); i++ {
		if instructions[i] == byte(compiler.OpJump) || 
		   instructions[i] == byte(compiler.OpJumpNotTruthy) {
			if i+2 < len(instructions) {
				target := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
				if target <= i {
					return true
				}
			}
		}
	}
	return false
}

func (j *JIT) calculateComplexity(instructions []byte) int {
	complexity := 0
	for i := 0; i < len(instructions); i++ {
		switch compiler.Opcode(instructions[i]) {
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			complexity += 2
		case compiler.OpCall:
			complexity += 3
		default:
			complexity++
		}
	}
	return complexity
}

func (j *JIT) optimizeFunction(code *CompiledCode) {
	if code == nil || code.Optimized {
		return
	}
	
	code.Optimized = true
	
	code.MachineCode = make([]byte, len(code.Function.Instructions))
	copy(code.MachineCode, code.Function.Instructions)
	
	for i := 0; i < len(code.MachineCode); i++ {
		if code.MachineCode[i] == byte(compiler.OpPop) && i+1 < len(code.MachineCode) {
			if code.MachineCode[i+1] == byte(compiler.OpPop) {
				code.MachineCode[i] = byte(compiler.OpDupTop)
				code.MachineCode[i+1] = 0x90
			}
		}
	}
}

func (j *JIT) evictOldEntries() {
	oldestKey := ""
	oldestTime := time.Now()
	
	for key, code := range j.cache {
		if code.CompileTime.Before(oldestTime) {
			oldestTime = code.CompileTime
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(j.cache, oldestKey)
		delete(j.executionCounts, oldestKey)
	}
}

func (j *JIT) GetStats() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	totalCalls := int64(0)
	for _, count := range j.executionCounts {
		totalCalls += count
	}
	
	hotFunctions := 0
	for _, code := range j.cache {
		if code.CallCount >= j.hotThreshold {
			hotFunctions++
		}
	}
	
	return map[string]interface{}{
		"cached_functions": len(j.cache),
		"total_calls":      totalCalls,
		"hot_functions":    hotFunctions,
		"hot_threshold":    j.hotThreshold,
	}
}

func (j *JIT) SetHotThreshold(threshold int64) {
	if threshold > 0 {
		j.hotThreshold = threshold
	}
}

func (j *JIT) ClearCache() {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	j.cache = make(map[string]*CompiledCode)
	j.executionCounts = make(map[string]int64)
}

func (j *JIT) IsHot(fn *compiler.CompiledFunction) bool {
	if fn == nil {
		return false
	}
	
	key := j.getFunctionKey(fn)
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	if count, ok := j.executionCounts[key]; ok {
		return count >= j.hotThreshold
	}
	return false
}

func (j *JIT) GetHotFunctions() []*CompiledCode {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	hotFunctions := make([]*CompiledCode, 0)
	for _, code := range j.cache {
		if code.CallCount >= j.hotThreshold {
			hotFunctions = append(hotFunctions, code)
		}
	}
	
	return hotFunctions
}
