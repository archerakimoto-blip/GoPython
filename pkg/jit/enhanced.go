package jit

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

type VMInterface interface {
	Push(obj objects.Object) error
	Pop() objects.Object
	GetStack() []objects.Object
	GetConstants() []objects.Object
	CurrentFrame() interface{}
}

type EnhancedJIT struct {
	*JIT
	generator      *MachineCodeGenerator
	vm             interface{}
	mu             sync.RWMutex
	compiledFuncs  map[string]*JITFunction
	executionStats *JITStats
}

type JITStats struct {
	TotalCompilations   int64
	TotalExecutions     int64
	MachineCodeSize     int64
	CompilationTime      time.Duration
	CacheHits           int64
	CacheMisses         int64
	OptimizationLevel   int
	EnableMC2Generation bool
}

type JITConfig struct {
	EnableMachineCode  bool
	OptimizationLevel int
	HotThreshold      int64
	MaxCacheSize      int
}

func NewEnhancedJIT(config *JITConfig) *EnhancedJIT {
	jit := &EnhancedJIT{
		JIT:             New(),
		generator:       NewMachineCodeGenerator(),
		compiledFuncs:   make(map[string]*JITFunction),
		executionStats: &JITStats{
			OptimizationLevel:   3,
			EnableMC2Generation: true,
		},
	}
	
	if config != nil {
		if config.HotThreshold > 0 {
			jit.SetHotThreshold(config.HotThreshold)
		}
		if config.MaxCacheSize > 0 {
			// Update max cache size logic if needed
		}
		if config.EnableMachineCode {
			jit.executionStats.EnableMC2Generation = true
		}
		if config.OptimizationLevel > 0 {
			jit.executionStats.OptimizationLevel = config.OptimizationLevel
		}
	}
	
	return jit
}

func (j *EnhancedJIT) SetVM(vm interface{}) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.vm = vm
}

func (j *EnhancedJIT) CompileFunction(fn *compiler.CompiledFunction) (*JITFunction, error) {
	if fn == nil {
		return nil, fmt.Errorf("nil function")
	}
	
	key := j.getFunctionKey(fn)
	
	j.mu.RLock()
	if jitFunc, ok := j.compiledFuncs[key]; ok {
		j.executionStats.CacheHits++
		j.mu.RUnlock()
		return jitFunc, nil
	}
	j.mu.RUnlock()
	
	j.mu.Lock()
	defer j.mu.Unlock()
	
	startTime := time.Now()
	
	jitFunc, err := j.generator.Generate(fn)
	if err != nil {
		return nil, fmt.Errorf("machine code generation failed: %w", err)
	}
	
	j.compiledFuncs[key] = jitFunc
	j.executionStats.TotalCompilations++
	j.executionStats.MachineCodeSize += int64(len(jitFunc.MachineCode))
	j.executionStats.CompilationTime += time.Since(startTime)
	
	return jitFunc, nil
}

func (j *EnhancedJIT) ExecuteFunction(fn *compiler.CompiledFunction) (objects.Object, error) {
	if !j.executionStats.EnableMC2Generation {
		return nil, fmt.Errorf("machine code generation is disabled")
	}
	
	jitFunc, err := j.CompileFunction(fn)
	if err != nil {
		return nil, err
	}
	
	if jitFunc.EntryPoint == 0 {
		return nil, fmt.Errorf("no entry point generated")
	}
	
	j.executionStats.TotalExecutions++
	
	return nil, nil
}

func (j *EnhancedJIT) ShouldCompileAdvanced(fn *compiler.CompiledFunction) bool {
	if fn == nil {
		return false
	}
	
	if !j.executionStats.EnableMC2Generation {
		return false
	}
	
	if fn.NumParameters == 0 && len(fn.Instructions) < 10 {
		return false
	}
	
	key := j.getFunctionKey(fn)
	j.mu.RLock()
	count := j.executionCounts[key]
	j.mu.RUnlock()
	
	hotThreshold := j.hotThreshold
	if j.executionStats.OptimizationLevel >= 3 {
		hotThreshold = hotThreshold / 2
	}
	
	return count >= hotThreshold
}

func (j *EnhancedJIT) OptimizeFunction(fn *compiler.CompiledFunction) (*compiler.CompiledFunction, error) {
	if fn == nil {
		return nil, fmt.Errorf("nil function")
	}
	
	optimized := &compiler.CompiledFunction{
		Instructions:  make([]byte, len(fn.Instructions)),
		Constants:     fn.Constants,
		NumParameters: fn.NumParameters,
	}
	
	copy(optimized.Instructions, fn.Instructions)
	
	j.applyOptimizations(optimized)
	
	return optimized, nil
}

func (j *EnhancedJIT) applyOptimizations(fn *compiler.CompiledFunction) {
	if j.executionStats.OptimizationLevel < 1 {
		return
	}
	
	j.deadCodeElimination(fn)
	
	if j.executionStats.OptimizationLevel >= 2 {
		j.constantFolding(fn)
		j.copyPropagation(fn)
	}
	
	if j.executionStats.OptimizationLevel >= 3 {
		j.registerAllocation(fn)
		j.loopOptimizations(fn)
	}
}

func (j *EnhancedJIT) deadCodeElimination(fn *compiler.CompiledFunction) {
	instructions := fn.Instructions
	reachable := make([]bool, len(instructions))
	
	for i := 0; i < len(instructions); i++ {
		if i == 0 || reachable[i] {
			reachable[i] = true
			
			op := compiler.Opcode(instructions[i])
			switch op {
			case compiler.OpJump:
				if i+2 < len(instructions) {
					target := int(instructions[i+1])<<8 | int(instructions[i+2])
					if target < len(reachable) {
						reachable[target] = true
					}
				}
				return
			case compiler.OpJumpNotTruthy:
				if i+2 < len(instructions) {
					target := int(instructions[i+1])<<8 | int(instructions[i+2])
					if target < len(reachable) {
						reachable[target] = true
					}
				}
			}
		}
	}
}

func (j *EnhancedJIT) constantFolding(fn *compiler.CompiledFunction) {
	instructions := fn.Instructions
	
	for i := 0; i < len(instructions)-2; {
		op := compiler.Opcode(instructions[i])
		
		if op == compiler.OpConstant {
			constIndex := int(instructions[i+1])<<8 | int(instructions[i+2])
			if constIndex < len(fn.Constants) {
				if constObj, ok := fn.Constants[constIndex].(*objects.Integer); ok {
					if i+3 < len(instructions) {
						nextOp := compiler.Opcode(instructions[i+3])
						if nextOp == compiler.OpConstant {
							nextConstIndex := int(instructions[i+4])<<8 | int(instructions[i+5])
							if nextConstIndex < len(fn.Constants) {
								if nextConstObj, ok := fn.Constants[nextConstIndex].(*objects.Integer); ok {
									if i+6 < len(instructions) {
										arithOp := compiler.Opcode(instructions[i+6])
										var result int64
										switch arithOp {
										case compiler.OpAdd:
											result = constObj.Value + nextConstObj.Value
										case compiler.OpSub:
											result = constObj.Value - nextConstObj.Value
										case compiler.OpMul:
											result = constObj.Value * nextConstObj.Value
										}
										
										if result != 0 || arithOp == compiler.OpMul {
											foldedConst := &objects.Integer{Value: result}
											newConstIndex := len(fn.Constants)
											fn.Constants = append(fn.Constants, foldedConst)
											
											instructions[i] = byte(compiler.OpConstant)
											instructions[i+1] = byte(newConstIndex >> 8)
											instructions[i+2] = byte(newConstIndex)
										}
									}
								}
							}
						}
					}
				}
			}
		}
		
		i++
	}
}

func (j *EnhancedJIT) copyPropagation(fn *compiler.CompiledFunction) {
}

func (j *EnhancedJIT) registerAllocation(fn *compiler.CompiledFunction) {
}

func (j *EnhancedJIT) loopOptimizations(fn *compiler.CompiledFunction) {
}

func (j *EnhancedJIT) GetStats() map[string]interface{} {
	stats := j.JIT.GetStats()
	
	stats["total_compilations"] = j.executionStats.TotalCompilations
	stats["total_executions"] = j.executionStats.TotalExecutions
	stats["machine_code_size"] = j.executionStats.MachineCodeSize
	stats["compilation_time_ms"] = j.executionStats.CompilationTime.Milliseconds()
	stats["cache_hits"] = j.executionStats.CacheHits
	stats["cache_misses"] = j.executionStats.CacheMisses
	stats["optimization_level"] = j.executionStats.OptimizationLevel
	stats["machine_code_enabled"] = j.executionStats.EnableMC2Generation
	
	if j.executionStats.TotalExecutions > 0 {
		stats["avg_compilation_time_ms"] = 
			float64(j.executionStats.CompilationTime.Milliseconds()) / float64(j.executionStats.TotalCompilations)
	}
	
	return stats
}

func (j *EnhancedJIT) PrintStats() {
	stats := j.GetStats()
	
	log.Println("=== JIT Statistics ===")
	log.Printf("Total compilations: %d", stats["total_compilations"])
	log.Printf("Total executions: %d", stats["total_executions"])
	log.Printf("Machine code size: %d bytes", stats["machine_code_size"])
	log.Printf("Compilation time: %d ms", stats["compilation_time_ms"])
	log.Printf("Cache hits: %d", stats["cache_hits"])
	log.Printf("Cache misses: %d", stats["cache_misses"])
	log.Printf("Optimization level: %d", stats["optimization_level"])
	log.Printf("Machine code enabled: %v", stats["machine_code_enabled"])
}

func (j *EnhancedJIT) EnableMachineCode(enable bool) {
	j.executionStats.EnableMC2Generation = enable
}

func (j *EnhancedJIT) SetOptimizationLevel(level int) {
	if level < 0 {
		level = 0
	}
	if level > 5 {
		level = 5
	}
	j.executionStats.OptimizationLevel = level
}

func (j *EnhancedJIT) ClearCompiledFunctions() {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	j.compiledFuncs = make(map[string]*JITFunction)
}

func (j *EnhancedJIT) GetCompiledFunctionCount() int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	return len(j.compiledFuncs)
}

func (j *EnhancedJIT) GetMachineCodeSize() int64 {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	total := int64(0)
	for _, fn := range j.compiledFuncs {
		total += int64(len(fn.MachineCode))
	}
	
	return total
}

func (j *EnhancedJIT) IsFunctionCompiled(fn *compiler.CompiledFunction) bool {
	if fn == nil {
		return false
	}
	
	key := j.getFunctionKey(fn)
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	_, ok := j.compiledFuncs[key]
	return ok
}

type JITProfile struct {
	FunctionName string
	CallCount    int64
	CompileTime  time.Duration
	ExecTime     time.Duration
	MachineCodeSize int
}

func (j *EnhancedJIT) Profile() []*JITProfile {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	profiles := make([]*JITProfile, 0, len(j.compiledFuncs))
	
	for key, jitFunc := range j.compiledFuncs {
		profile := &JITProfile{
			FunctionName:   key,
			MachineCodeSize: len(jitFunc.MachineCode),
		}
		profiles = append(profiles, profile)
	}
	
	return profiles
}
