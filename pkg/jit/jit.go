package jit

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/compiler"
)

// JIT Compiler - 简单的即时编译器
type JIT struct {
	cache map[string]compiledFunction
}

type compiledFunction struct {
	code     []byte
	executed int
}

func New() *JIT {
	return &JIT{
		cache: make(map[string]compiledFunction),
	}
}

func (j *JIT) Compile(node ast.Node) interface{} {
	// 这里应该编译 AST 到本地代码（简单实现返回nil
	return nil
}

func (j *JIT) ShouldCompile(fn *compiler.CompiledFunction) bool {
	// 简单的启发式：函数被调用超过10次就编译
	// 实际实现需要更复杂的分析
	return false
}

func (j *JIT) GetOrCompile(fn *compiler.CompiledFunction) (interface{}, error) {
	key := fmt.Sprintf("%p", fn)
	
	if compiled, ok := j.cache[key]; ok {
		compiled.executed++
		if compiled.executed >= 10 && compiled.code != nil {
			return compiled.code, nil
		}
		j.cache[key] = compiled
		return nil, nil
	}

	j.cache[key] = compiledFunction{executed: 1}
	return nil, nil
}

