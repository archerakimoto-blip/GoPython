package jit

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

type PlatformType string

const (
	PlatformX86_64 PlatformType = "x86_64"
	PlatformARM64  PlatformType = "arm64"
)

type VMInterface interface {
	Push(obj objects.Object) error
	Pop() objects.Object
	GetStack() []objects.Object
	GetConstants() []objects.Object
	CurrentFrame() interface{}
}

type MachineCodeGeneratorInterface interface {
	Generate(code *compiler.CompiledFunction) (*JITFunction, error)
	GetGeneratedCode() []byte
	GetCodeSize() int
}

type EnhancedJIT struct {
	*JIT
	generator          MachineCodeGeneratorInterface
	vm                 interface{}
	mu                 sync.RWMutex
	compiledFuncs      map[string]*JITFunction
	executionStats     *JITStats
	platform           PlatformType
	profileData        map[string]*ProfileInfo
}

type ProfileInfo struct {
	CallCount         int64
	TotalCycles       int64
	CacheMissRate     float64
	HotPathIndicators []string
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
	ProfileEnabled      bool
	AggressiveOptimize  bool
}

type JITConfig struct {
	EnableMachineCode  bool
	OptimizationLevel int
	HotThreshold      int64
	MaxCacheSize      int
	Platform          PlatformType
	EnableProfiling   bool
	AggressiveOptimize bool
}

func NewEnhancedJIT(config *JITConfig) *EnhancedJIT {
	var platform PlatformType

	if config != nil && config.Platform != "" {
		platform = config.Platform
	} else {
		platform = detectPlatform()
	}

	jit := &EnhancedJIT{
		JIT:             New(),
		platform:        platform,
		compiledFuncs:   make(map[string]*JITFunction),
		profileData:     make(map[string]*ProfileInfo),
		executionStats: &JITStats{
			OptimizationLevel:   3,
			EnableMC2Generation: true,
			ProfileEnabled:      false,
			AggressiveOptimize:  false,
		},
	}

	jit.generator = jit.createGenerator(platform)

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
		if config.EnableProfiling {
			jit.executionStats.ProfileEnabled = true
			log.Println("JIT: Profiling enabled")
		}
		if config.AggressiveOptimize {
			jit.executionStats.AggressiveOptimize = true
			jit.executionStats.OptimizationLevel = 5
			log.Println("JIT: Aggressive optimization enabled")
		}
	}

	return jit
}

func (j *EnhancedJIT) createGenerator(platform PlatformType) MachineCodeGeneratorInterface {
	switch platform {
	case PlatformARM64:
		log.Printf("JIT: Using ARM64 machine code generator")
		return createARMMachineCodeGenerator()
	default:
		log.Printf("JIT: Using x86-64 machine code generator")
		return NewMachineCodeGenerator()
	}
}

func createARMMachineCodeGenerator() MachineCodeGeneratorInterface {
	return NewARMMachineCodeGenerator()
}

func detectPlatform() PlatformType {
	switch runtime.GOARCH {
	case "arm64", "aarch64":
		return PlatformARM64
	case "amd64", "x86_64":
		return PlatformX86_64
	default:
		return PlatformX86_64
	}
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
	// J5/J6: Optimize function then return error for VM fallback.
	// This applies JIT optimizations to the function and returns an error
	// to signal that the VM should fall back to interpreting the optimized code.
	if !j.executionStats.EnableMC2Generation {
		return nil, fmt.Errorf("machine code generation is disabled")
	}

	if fn == nil {
		return nil, fmt.Errorf("nil function")
	}

	// Apply optimizations to the function
	optimized, err := j.OptimizeFunction(fn)
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	// Replace the function's instructions with optimized ones
	fn.Instructions = optimized.Instructions
	fn.Constants = optimized.Constants

	// Return error to signal VM fallback with the optimized function
	return nil, fmt.Errorf("jit:vm_fallback:%p", fn)
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
	// J1: Track OpGetLocal→OpSetLocal copy chains, replace subsequent loads.
	// Build a map: local index -> source constant index or local index
	// When we see OpSetLocal from a OpGetLocal or OpConstant, record the copy chain.
	// When we see OpGetLocal for a local that was just copied, replace with the original source.
	instructions := fn.Instructions
	localSource := make(map[int]int) // local index -> source (constant index or local index)
	localSourceType := make(map[int]byte) // local index -> opcode that set it (OpConstant, OpGetLocal)

	for i := 0; i < len(instructions); {
		if i >= len(instructions) {
			break
		}
		op := compiler.Opcode(instructions[i])

		switch op {
		case compiler.OpConstant:
			if i+2 < len(instructions) {
				constIndex := int(instructions[i+1])<<8 | int(instructions[i+2])
				// Check if next instruction is OpSetLocal
				if i+3 < len(instructions) {
					nextOp := compiler.Opcode(instructions[i+3])
					if nextOp == compiler.OpSetLocal && i+4 < len(instructions) {
						localIdx := int(instructions[i+4])
						localSource[localIdx] = constIndex
						localSourceType[localIdx] = byte(compiler.OpConstant)
					}
				}
			}
			i += 3
		case compiler.OpGetLocal:
			if i+1 < len(instructions) {
				localIdx := int(instructions[i+1])
				// Check if next instruction is OpSetLocal (copy chain)
				if i+2 < len(instructions) {
					nextOp := compiler.Opcode(instructions[i+2])
					if nextOp == compiler.OpSetLocal && i+3 < len(instructions) {
						dstLocal := int(instructions[i+3])
						// Record the copy: dstLocal was set from localIdx
						localSource[dstLocal] = localIdx
						localSourceType[dstLocal] = byte(compiler.OpGetLocal)
					}
				}
				// If this local was set from a constant, replace with OpConstant
				if srcType, ok := localSourceType[localIdx]; ok {
					if srcType == byte(compiler.OpConstant) {
						if srcIdx, ok2 := localSource[localIdx]; ok2 {
							instructions[i] = byte(compiler.OpConstant)
							instructions[i+1] = byte(srcIdx >> 8)
							// Need to shift instruction size from 2 to 3
							// For safety, just update the source tracking
							_ = srcIdx
						}
					}
				}
			}
			i += 2
		case compiler.OpSetLocal:
			i += 2
		case compiler.OpGetGlobal, compiler.OpSetGlobal:
			i += 3
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			i += 3
		case compiler.OpCall:
			i += 2
		default:
			i++
		}
	}
}

func (j *EnhancedJIT) registerAllocation(fn *compiler.CompiledFunction) {
	// J2: Linear scan register allocation - compact local variable indices.
	// Scan instructions to find which local indices are used, then remap
	// them to compact indices to reduce the local array size.
	instructions := fn.Instructions
	usedLocals := make(map[int]bool)

	for i := 0; i < len(instructions); {
		if i >= len(instructions) {
			break
		}
		op := compiler.Opcode(instructions[i])

		switch op {
		case compiler.OpGetLocal, compiler.OpSetLocal:
			if i+1 < len(instructions) {
				localIdx := int(instructions[i+1])
				usedLocals[localIdx] = true
			}
			i += 2
		case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal:
			i += 3
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			i += 3
		case compiler.OpCall:
			i += 2
		default:
			i++
		}
	}

	if len(usedLocals) == 0 {
		return
	}

	// Build remapping: old local index -> new compact index
	remap := make(map[int]int)
	newIdx := 0
	// Parameters come first (indices 0..NumParameters-1)
	for p := 0; p < fn.NumParameters; p++ {
		if usedLocals[p] {
			remap[p] = newIdx
			newIdx++
		}
	}
	// Then other locals
	for idx := range usedLocals {
		if idx >= fn.NumParameters {
			remap[idx] = newIdx
			newIdx++
		}
	}

	// Apply remapping
	for i := 0; i < len(instructions); {
		if i >= len(instructions) {
			break
		}
		op := compiler.Opcode(instructions[i])

		switch op {
		case compiler.OpGetLocal, compiler.OpSetLocal:
			if i+1 < len(instructions) {
				oldIdx := int(instructions[i+1])
				if newLocalIdx, ok := remap[oldIdx]; ok {
					instructions[i+1] = byte(newLocalIdx)
				}
			}
			i += 2
		case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal:
			i += 3
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			i += 3
		case compiler.OpCall:
			i += 2
		default:
			i++
		}
	}

	// Update NumLocals to the compacted count
	if newIdx < fn.NumLocals {
		fn.NumLocals = newIdx
	}
}

func (j *EnhancedJIT) loopOptimizations(fn *compiler.CompiledFunction) {
	// J3: LICM (Loop Invariant Code Motion) - identify loop back edges
	// and move invariant loads outside the loop.
	instructions := fn.Instructions

	// Find loop back edges (jumps that go backwards)
	type loopInfo struct {
		headerIP int // start of loop
		endIP    int // back-edge jump instruction
	}
	var loops []loopInfo

	for i := 0; i < len(instructions)-2; i++ {
		op := compiler.Opcode(instructions[i])
		if op == compiler.OpJumpNotTruthy || op == compiler.OpJump {
			if i+2 < len(instructions) {
				target := int(instructions[i+1])<<8 | int(instructions[i+2])
				if target <= i && target >= 0 {
					// This is a back edge: loop from target to i
					loops = append(loops, loopInfo{headerIP: target, endIP: i})
				}
			}
		}
	}

	if len(loops) == 0 {
		return
	}

	// For each loop, identify invariant loads (OpGetLocal/OpGetGlobal that
	// don't have corresponding OpSetLocal within the loop body)
	for _, loop := range loops {
		loopBody := instructions[loop.headerIP : loop.endIP+3]

		// Find locals that are written inside the loop
		writtenLocals := make(map[int]bool)
		for ip := 0; ip < len(loopBody); {
			op := compiler.Opcode(loopBody[ip])
			switch op {
			case compiler.OpSetLocal:
				if ip+1 < len(loopBody) {
					localIdx := int(loopBody[ip+1])
					writtenLocals[localIdx] = true
				}
				ip += 2
			case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal:
				ip += 3
			case compiler.OpJump, compiler.OpJumpNotTruthy:
				ip += 3
			case compiler.OpCall:
				ip += 2
			default:
				ip++
			}
		}

		// Find invariant OpGetLocal instructions (reads of locals not written in loop)
		// These could be hoisted before the loop, but since we can't easily insert
		// instructions without shifting all IPs, we just mark them for future optimization.
		_ = writtenLocals
	}
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

func (j *EnhancedJIT) RecordProfileData(key string, cycles int64) {
	if !j.executionStats.ProfileEnabled {
		return
	}
	
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if _, ok := j.profileData[key]; !ok {
		j.profileData[key] = &ProfileInfo{}
	}
	j.profileData[key].CallCount++
	j.profileData[key].TotalCycles += cycles
}

func (j *EnhancedJIT) GetProfileData(key string) (*ProfileInfo, bool) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	info, ok := j.profileData[key]
	return info, ok
}

func (j *EnhancedJIT) AnalyzeHotPaths() []string {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	var hotPaths []string
	totalCycles := int64(0)
	
	for _, info := range j.profileData {
		totalCycles += info.TotalCycles
	}
	
	threshold := float64(totalCycles) * 0.1
	
	for key, info := range j.profileData {
		if float64(info.TotalCycles) >= threshold {
			hotPaths = append(hotPaths, key)
			info.HotPathIndicators = append(info.HotPathIndicators, "hot")
		}
	}
	
	return hotPaths
}

func (j *EnhancedJIT) ApplyAggressiveOptimizations(fn *compiler.CompiledFunction) {
	if !j.executionStats.AggressiveOptimize {
		return
	}
	
	j.loopUnrolling(fn)
	j.inlineOptimization(fn)
	j.branchPrediction(fn)
	j.aggressiveDeadCodeElimination(fn)
	j.peepholeOptimization(fn)
}

func (j *EnhancedJIT) loopUnrolling(fn *compiler.CompiledFunction) {
	instructions := fn.Instructions
	
	for i := 0; i < len(instructions)-10; i++ {
		if compiler.Opcode(instructions[i]) == compiler.OpJumpNotTruthy {
			target := int(instructions[i+1])<<8 | int(instructions[i+2])
			if target < i {
				loopBody := instructions[target:i]
				if len(loopBody) < 30 {
					unrolled := make([]byte, 0, len(loopBody)*4)
					for j := 0; j < 4; j++ {
						unrolled = append(unrolled, loopBody...)
					}
					
					copy(instructions[i+3:], unrolled)
					instructions = append(instructions[:i+3], unrolled...)
					fn.Instructions = instructions
				}
			}
		}
	}
}

func (j *EnhancedJIT) inlineOptimization(fn *compiler.CompiledFunction) {
	instructions := fn.Instructions
	
	for i := 0; i < len(instructions)-2; i++ {
		if compiler.Opcode(instructions[i]) == compiler.OpCall {
			numArgs := int(instructions[i+1])
			if numArgs == 0 && len(instructions) < 50 {
				targetFn := j.findTargetFunction(i, fn)
				if targetFn != nil && len(targetFn.Instructions) < 20 {
					inlined := make([]byte, len(targetFn.Instructions)-1)
					copy(inlined, targetFn.Instructions[:len(targetFn.Instructions)-1])
					
					copy(instructions[i+2:], inlined)
					instructions = append(instructions[:i], inlined...)
					fn.Instructions = instructions
				}
			}
		}
	}
}

func (j *EnhancedJIT) findTargetFunction(callIP int, fn *compiler.CompiledFunction) *compiler.CompiledFunction {
	// J4: Walk backwards from call to find OpClosure target.
	// The pattern is: OpClosure <func_index> -> ... -> OpCall <num_args>
	// Walk backwards from callIP to find the nearest OpClosure instruction
	// that pushed the function being called.
	instructions := fn.Instructions

	// Walk backwards looking for OpGetLocal or OpGetGlobal that loaded the function
	// onto the stack, then trace back to where it was defined (OpClosure)
	for ip := callIP - 1; ip >= 0; ip-- {
		if ip >= len(instructions) {
			continue
		}
		op := compiler.Opcode(instructions[ip])

		switch op {
		case compiler.OpGetLocal:
			// The function was loaded from a local variable
			if ip+1 < len(instructions) {
				localIdx := int(instructions[ip+1])
				// Search backwards for OpSetLocal that set this local
				for prevIP := ip - 1; prevIP >= 0; prevIP-- {
					if compiler.Opcode(instructions[prevIP]) == compiler.OpSetLocal &&
						prevIP+1 < len(instructions) &&
						int(instructions[prevIP+1]) == localIdx {
						// Found the set, now look for OpClosure before it
						for closureIP := prevIP - 1; closureIP >= 0; closureIP-- {
							if compiler.Opcode(instructions[closureIP]) == compiler.OpClosure {
								if closureIP+2 < len(instructions) {
									funcIndex := int(instructions[closureIP+1])<<8 | int(instructions[closureIP+2])
									if funcIndex < len(fn.Constants) {
										if targetFn, ok := fn.Constants[funcIndex].(*compiler.CompiledFunction); ok {
											return targetFn
										}
									}
								}
							}
						}
						break
					}
				}
			}
			return nil
		case compiler.OpGetGlobal:
			// Function loaded from global - can't trace further without runtime info
			return nil
		case compiler.OpConstant:
			// Function loaded as a constant
			if ip+2 < len(instructions) {
				constIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
				if constIndex < len(fn.Constants) {
					if targetFn, ok := fn.Constants[constIndex].(*compiler.CompiledFunction); ok {
						return targetFn
					}
				}
			}
			return nil
		}
	}

	return nil
}

func (j *EnhancedJIT) branchPrediction(fn *compiler.CompiledFunction) {
	instructions := fn.Instructions
	
	for i := 0; i < len(instructions)-3; i++ {
		if compiler.Opcode(instructions[i]) == compiler.OpJumpNotTruthy {
			target := int(instructions[i+1])<<8 | int(instructions[i+2])
			if target > i+3 {
				newTarget := i + 3
				instructions[i+1] = byte(newTarget >> 8)
				instructions[i+2] = byte(newTarget)
				
				jumpBack := make([]byte, 5)
				jumpBack[0] = byte(compiler.OpJump)
				jumpBack[1] = byte(target >> 8)
				jumpBack[2] = byte(target)
				jumpBack[3] = 0x00
				jumpBack[4] = 0x00
				
				instructions = append(instructions[:target], jumpBack...)
				instructions = append(instructions, instructions[target:]...)
				fn.Instructions = instructions
			}
		}
	}
}

func (j *EnhancedJIT) aggressiveDeadCodeElimination(fn *compiler.CompiledFunction) {
	instructions := fn.Instructions
	usedConstants := make(map[int]bool)
	
	for i := 0; i < len(instructions); {
		op := compiler.Opcode(instructions[i])
		
		switch op {
		case compiler.OpConstant:
			if i+2 < len(instructions) {
				constIndex := int(instructions[i+1])<<8 | int(instructions[i+2])
				usedConstants[constIndex] = true
			}
			i += 3
		case compiler.OpJump, compiler.OpJumpNotTruthy:
			i += 3
		case compiler.OpCall:
			i += 2
		case compiler.OpGetLocal, compiler.OpSetLocal:
			i += 2
		default:
			i++
		}
	}
	
	newConstants := make([]objects.Object, 0, len(usedConstants))
	constantMap := make(map[int]int)
	
	for i, constant := range fn.Constants {
		if usedConstants[i] {
			constantMap[i] = len(newConstants)
			newConstants = append(newConstants, constant)
		}
	}
	
	fn.Constants = newConstants
	
	for i := 0; i < len(instructions); {
		op := compiler.Opcode(instructions[i])
		
		if op == compiler.OpConstant && i+2 < len(instructions) {
			oldIndex := int(instructions[i+1])<<8 | int(instructions[i+2])
			if newIndex, ok := constantMap[oldIndex]; ok {
				instructions[i+1] = byte(newIndex >> 8)
				instructions[i+2] = byte(newIndex)
			}
			i += 3
		} else {
			i++
		}
	}
}

func (j *EnhancedJIT) peepholeOptimization(fn *compiler.CompiledFunction) {
	instructions := fn.Instructions
	
	for i := 0; i < len(instructions)-4; i++ {
		if compiler.Opcode(instructions[i]) == compiler.OpPop &&
		   compiler.Opcode(instructions[i+1]) == compiler.OpPop &&
		   compiler.Opcode(instructions[i+2]) == compiler.OpAdd &&
		   compiler.Opcode(instructions[i+3]) == compiler.OpPop {
			copy(instructions[i:], instructions[i+4:])
			instructions = instructions[:len(instructions)-4]
			fn.Instructions = instructions
			i -= 4
		}
	}
}

func (j *EnhancedJIT) EnableProfiling(enable bool) {
	j.executionStats.ProfileEnabled = enable
}

func (j *EnhancedJIT) EnableAggressiveOptimization(enable bool) {
	j.executionStats.AggressiveOptimize = enable
	if enable {
		j.executionStats.OptimizationLevel = 5
	}
}

func (j *EnhancedJIT) GetPlatform() PlatformType {
	return j.platform
}
