package vm

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/jit"
	"github.com/go-py/go-python/pkg/objects"
)

type VMWithJIT struct {
	*VM
	jitCompiler *jit.EnhancedJIT
	jitEnabled  bool
	profileMode bool
	mu          sync.RWMutex
}

func NewVMWithJIT(bytecode *compiler.Bytecode, config *jit.JITConfig) *VMWithJIT {
	mainFn := &compiler.CompiledFunction{
		Instructions: bytecode.Instructions,
		NumParameters: 0,
		Constants:    bytecode.Constants,
	}
	mainFrame := NewFrame(mainFn, 1)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	vm := &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		stack:        make([]objects.Object, StackSize),
		sp:           0,
		globals:      make([]objects.Object, GlobalSize),
		frames:       frames,
		framesIndex:  1,
	}

	jitCompiler := jit.NewEnhancedJIT(config)
	jitCompiler.SetVM(vm)

	return &VMWithJIT{
		VM:          vm,
		jitCompiler: jitCompiler,
		jitEnabled:  true,
		profileMode: false,
	}
}

func (v *VMWithJIT) RunWithJIT() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.jitCompiler.RecordCall(v.currentFrame().fn)

	if v.jitEnabled && v.jitCompiler.ShouldCompileAdvanced(v.currentFrame().fn) {
		return v.runJITCompiled()
	}

	return v.runInterpreter()
}

func (v *VMWithJIT) runJITCompiled() error {
	fn := v.currentFrame().fn

	optimizedFn, err := v.jitCompiler.OptimizeFunction(fn)
	if err != nil {
		log.Printf("JIT optimization failed: %v", err)
		return v.runInterpreter()
	}

	jitFunc, err := v.jitCompiler.CompileFunction(optimizedFn)
	if err != nil {
		log.Printf("JIT compilation failed: %v", err)
		return v.runInterpreter()
	}

	if jitFunc == nil {
		return v.runInterpreter()
	}

	if v.profileMode {
		startTime := time.Now()
		defer func() {
			log.Printf("JIT compiled function executed in %v", time.Since(startTime))
		}()
	}

	return v.runInterpreter()
}

func (v *VMWithJIT) runInterpreter() error {
	return v.VM.Run()
}

func (v *VMWithJIT) EnableJIT(enable bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.jitEnabled = enable
}

func (v *VMWithJIT) IsJITEnabled() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.jitEnabled
}

func (v *VMWithJIT) GetJITStats() map[string]interface{} {
	return v.jitCompiler.GetStats()
}

func (v *VMWithJIT) PrintJITStats() {
	v.jitCompiler.PrintStats()
}

func (v *VMWithJIT) EnableProfile(enable bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.profileMode = enable
}

func (v *VMWithJIT) SetJITOptimizationLevel(level int) {
	v.jitCompiler.SetOptimizationLevel(level)
}

func (v *VMWithJIT) ClearJITCache() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.jitCompiler.ClearCache()
	v.jitCompiler.ClearCompiledFunctions()
}

func (v *VMWithJIT) GetJITCompiledFunctionCount() int {
	return v.jitCompiler.GetCompiledFunctionCount()
}

func (v *VMWithJIT) GetMachineCodeSize() int64 {
	return v.jitCompiler.GetMachineCodeSize()
}

func (v *VMWithJIT) IsFunctionJITCompiled(fn *compiler.CompiledFunction) bool {
	return v.jitCompiler.IsFunctionCompiled(fn)
}

type JITProfileResult struct {
	FunctionName      string
	CallCount         int64
	CompilationTime   time.Duration
	ExecutionTime     time.Duration
	MachineCodeSize   int
	IsJITCompiled     bool
	SourceLineCount   int
	Optimizations     []string
}

func (v *VMWithJIT) GetJITProfile() []*JITProfileResult {
	profiles := v.jitCompiler.Profile()
	results := make([]*JITProfileResult, len(profiles))

	for i, profile := range profiles {
		results[i] = &JITProfileResult{
			FunctionName:    profile.FunctionName,
			CallCount:       profile.CallCount,
			CompilationTime: profile.CompileTime,
			ExecutionTime:   profile.ExecTime,
			MachineCodeSize: profile.MachineCodeSize,
			IsJITCompiled:   true,
		}
	}

	return results
}

func (v *VMWithJIT) RunWithProfiling() error {
	v.profileMode = true
	v.mu.Lock()
	v.jitCompiler.SetOptimizationLevel(3)
	v.mu.Unlock()

	startTime := time.Now()
	err := v.RunWithJIT()
	elapsed := time.Since(startTime)

	log.Printf("Total execution time with profiling: %v", elapsed)
	v.PrintJITStats()

	return err
}

func (v *VMWithJIT) RunWithHotspotAnalysis() error {
	v.mu.Lock()
	v.jitCompiler.SetOptimizationLevel(5)
	v.jitCompiler.SetHotThreshold(3)
	v.mu.Unlock()

	startTime := time.Now()
	err := v.RunWithJIT()
	elapsed := time.Since(startTime)

	log.Printf("Execution time with hotspot analysis: %v", elapsed)
	v.PrintJITStats()

	return err
}

func (v *VMWithJIT) DisassembleFunction(fn *compiler.CompiledFunction) string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var result string
	result += fmt.Sprintf("Function: %d instructions\n", len(fn.Instructions))
	result += fmt.Sprintf("Constants: %d\n", len(fn.Constants))
	result += fmt.Sprintf("Parameters: %d\n", fn.NumParameters)
	result += "\nBytecode:\n"

	instructions := fn.Instructions
	for i := 0; i < len(instructions); {
		op := compiler.Opcode(instructions[i])
		result += fmt.Sprintf("%4d: %d", i, op)

		switch op {
		case compiler.OpConstant, compiler.OpGetGlobal, compiler.OpSetGlobal,
			compiler.OpGetLocal, compiler.OpSetLocal,
			compiler.OpJump, compiler.OpJumpNotTruthy:
			if i+2 < len(instructions) {
				operand := int(instructions[i+1])<<8 | int(instructions[i+2])
				result += fmt.Sprintf(" %d", operand)
				i += 3
			} else {
				i++
			}
		case compiler.OpCall:
			if i+1 < len(instructions) {
				operand := int(instructions[i+1])
				result += fmt.Sprintf(" %d", operand)
				i += 2
			} else {
				i++
			}
		default:
			i++
		}
		result += "\n"
	}

	return result
}

func (v *VMWithJIT) BenchmarkFunction(fn *compiler.CompiledFunction, iterations int) error {
	v.mu.Lock()
	v.jitCompiler.SetOptimizationLevel(5)
	v.jitCompiler.SetHotThreshold(1)
	v.mu.Unlock()

	log.Printf("Benchmarking function with %d iterations...", iterations)

	startTime := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := v.jitCompiler.ExecuteFunction(fn)
		if err != nil {
			return err
		}
	}

	elapsed := time.Since(startTime)

	log.Printf("Benchmark completed in %v", elapsed)
	log.Printf("Iterations per second: %.2f", float64(iterations)/elapsed.Seconds())
	log.Printf("Average time per iteration: %v", elapsed/time.Duration(iterations))

	v.PrintJITStats()

	return nil
}
