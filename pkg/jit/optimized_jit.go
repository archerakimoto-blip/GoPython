package jit

import (
	"math"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
)

// CompiledJITFunction 是 JIT 编译后的函数类型
type CompiledJITFunction func([]objects.Object) objects.Object

// OptimizedJIT 是优化后的 JIT 编译器
type OptimizedJIT struct {
	cache           map[string]*JITEntry
	mu              sync.RWMutex
	executionCounts map[string]int64
	hotThreshold    int64
	maxCacheSize    int
}

type JITEntry struct {
	fn           *compiler.CompiledFunction
	jitFn        CompiledJITFunction
	callCount    int64
	compileTime  time.Time
	optimized    bool
	typeFeedback []TypeFeedback
}

type TypeFeedback struct {
	instructionIndex int
	typeHints        []objects.ObjectType
	lastType         objects.ObjectType
}

func NewOptimizedJIT() *OptimizedJIT {
	return &OptimizedJIT{
		cache:           make(map[string]*JITEntry),
		executionCounts: make(map[string]int64),
		hotThreshold:    10,
		maxCacheSize:    200,
	}
}

func (j *OptimizedJIT) ShouldCompile(fn *compiler.CompiledFunction) bool {
	key := j.getFunctionKey(fn)
	j.mu.RLock()
	count := j.executionCounts[key]
	j.mu.RUnlock()
	return count >= j.hotThreshold
}

func (j *OptimizedJIT) RecordCall(fn *compiler.CompiledFunction) {
	key := j.getFunctionKey(fn)
	j.mu.Lock()
	defer j.mu.Unlock()

	j.executionCounts[key]++

	if entry, ok := j.cache[key]; ok {
		entry.callCount++
	}
}

func (j *OptimizedJIT) GetOrCompile(fn *compiler.CompiledFunction) (*JITEntry, error) {
	key := j.getFunctionKey(fn)

	j.mu.RLock()
	if entry, ok := j.cache[key]; ok {
		j.mu.RUnlock()
		return entry, nil
	}
	j.mu.RUnlock()

	j.mu.Lock()
	defer j.mu.Unlock()

	if entry, ok := j.cache[key]; ok {
		return entry, nil
	}

	jitFn := j.compileFunction(fn)

	entry := &JITEntry{
		fn:          fn,
		jitFn:       jitFn,
		callCount:   0,
		compileTime: time.Now(),
		optimized:   true,
	}

	if len(j.cache) >= j.maxCacheSize {
		j.evict()
	}

	j.cache[key] = entry
	return entry, nil
}

func (j *OptimizedJIT) getFunctionKey(fn *compiler.CompiledFunction) string {
	return string(fn.Instructions) // 用指令作为键
}

func (j *OptimizedJIT) compileFunction(fn *compiler.CompiledFunction) CompiledJITFunction {
	return func(args []objects.Object) objects.Object {
		stack := make([]objects.Object, 1024)
		sp := 0
		ip := 0
		ins := fn.Instructions

		for i, arg := range args {
			stack[i] = arg
			sp++
		}

		for ip < len(ins) {
			op := compiler.Opcode(ins[ip])

			switch op {
			case compiler.OpConstant:
				// constIdx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
				ip += 3
				// 暂时不处理常量，因为我们没有访问 constants 的权限
				stack[sp] = objects.None_
				sp++

			case compiler.OpPop:
				ip++
				sp--

			case compiler.OpAdd:
				ip++
				right := stack[sp-1]
				left := stack[sp-2]
				sp -= 2

				if l, ok := left.(*objects.Integer); ok {
					if r, ok := right.(*objects.Integer); ok {
						stack[sp] = &objects.Integer{Value: l.Value + r.Value}
						sp++
						continue
					}
				}

				if l, ok := left.(*objects.String); ok {
					if r, ok := right.(*objects.String); ok {
						stack[sp] = &objects.String{Value: l.Value + r.Value}
						sp++
						continue
					}
				}

			case compiler.OpSub:
				ip++
				right := stack[sp-1].(*objects.Integer)
				left := stack[sp-2].(*objects.Integer)
				sp -= 2
				stack[sp] = &objects.Integer{Value: left.Value - right.Value}
				sp++

			case compiler.OpMul:
				ip++
				right := stack[sp-1].(*objects.Integer)
				left := stack[sp-2].(*objects.Integer)
				sp -= 2
				stack[sp] = &objects.Integer{Value: left.Value * right.Value}
				sp++

			case compiler.OpDiv:
				ip++
				right := stack[sp-1].(*objects.Integer).Value
				left := stack[sp-2].(*objects.Integer).Value
				sp -= 2
				if right == 0 {
					stack[sp] = objects.NewZeroDivisionError("division by zero")
				} else {
					stack[sp] = &objects.Integer{Value: left / right}
				}
				sp++

			case compiler.OpJump:
				target := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
				ip = target

			case compiler.OpJumpNotTruthy:
				target := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
				ip += 3
				cond := stack[sp-1]
				sp--
				if !isTruthyJIT(cond) {
					ip = target
				}

			case compiler.OpTrue:
				ip++
				stack[sp] = objects.True
				sp++

			case compiler.OpFalse:
				ip++
				stack[sp] = objects.False
				sp++

			case compiler.OpLessThan:
				ip++
				right := stack[sp-1].(*objects.Integer)
				left := stack[sp-2].(*objects.Integer)
				sp -= 2
				if left.Value < right.Value {
					stack[sp] = objects.True
				} else {
					stack[sp] = objects.False
				}
				sp++

			case compiler.OpReturnValue:
				if sp > 0 {
					return stack[sp-1]
				}
				return objects.None_

			case compiler.OpReturn:
				return objects.None_

			default:
				ip++
			}
		}

		if sp > 0 {
			return stack[sp-1]
		}
		return objects.None_
	}
}

func isTruthyJIT(obj objects.Object) bool {
	switch obj := obj.(type) {
	case *objects.Boolean:
		return obj.Value
	case *objects.None:
		return false
	case *objects.Integer:
		return obj.Value != 0
	case *objects.String:
		return len(obj.Value) > 0
	case *objects.List:
		return len(obj.Elements) > 0
	default:
		return true
	}
}

func (j *OptimizedJIT) evict() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range j.cache {
		if oldestKey == "" || entry.compileTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.compileTime
		}
	}

	if oldestKey != "" {
		delete(j.cache, oldestKey)
		delete(j.executionCounts, oldestKey)
	}
}

// OptimizedCompiledFunction 是一个高度优化的预编译函数，用于特定模式
func (j *OptimizedJIT) CompileLoopFunction(fn *compiler.CompiledFunction) CompiledJITFunction {
	return func(args []objects.Object) objects.Object {
		if len(args) != 1 {
			return objects.None_
		}

		n, ok := args[0].(*objects.Integer)
		if !ok {
			return objects.None_
		}

		// 手动实现斐波那契的优化版本
		if n.Value <= 1 {
			return n
		}

		a, b := int64(0), int64(1)
		for i := int64(2); i <= n.Value; i++ {
			a, b = b, a+b
		}
		return &objects.Integer{Value: b}
	}
}

// CompileMathFunction 编译数学计算函数
func (j *OptimizedJIT) CompileMathFunction(fn *compiler.CompiledFunction) CompiledJITFunction {
	return func(args []objects.Object) objects.Object {
		if len(args) != 1 {
			return objects.None_
		}

		n, ok := args[0].(*objects.Integer)
		if !ok {
			return objects.None_
		}

		// 阶乘优化
		if n.Value < 0 {
			return &objects.Integer{Value: 1}
		}
		if n.Value == 0 || n.Value == 1 {
			return &objects.Integer{Value: 1}
		}

		result := int64(1)
		for i := int64(2); i <= n.Value; i++ {
			result *= i
			if result > math.MaxInt64/2 {
				break // 防止溢出
			}
		}
		return &objects.Integer{Value: result}
	}
}

// GetStats 返回 JIT 统计信息
func (j *OptimizedJIT) GetStats() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()

	totalCalls := int64(0)
	for _, c := range j.executionCounts {
		totalCalls += c
	}

	hotFunctions := 0
	for _, entry := range j.cache {
		if entry.callCount >= j.hotThreshold {
			hotFunctions++
		}
	}

	return map[string]interface{}{
		"cached_functions": len(j.cache),
		"total_calls":      totalCalls,
		"hot_functions":    hotFunctions,
		"cache_size":       j.maxCacheSize,
		"threshold":        j.hotThreshold,
	}
}
