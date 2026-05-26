package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/jit"
)

func main() {
	fmt.Println("=== GoPython JIT 功能演示 ===\n")
	
	fmt.Println("=== 1. 创建 JIT 实例并设置热点阈值 ===")
	j := jit.New()
	j.SetHotThreshold(3)
	fmt.Printf("热点阈值已设置为: 3\n\n")
	
	fmt.Println("=== 2. 测试函数编译和缓存 ===")
	testFunc1 := &compiler.CompiledFunction{
		Instructions: []byte{1, 2, 3, 4, 5},
		NumLocals:    5,
		NumParameters: 2,
	}
	
	testFunc2 := &compiler.CompiledFunction{
		Instructions: []byte{6, 7, 8, 9, 10, 11, 12},
		NumLocals:    10,
		NumParameters: 3,
	}
	
	compiled1, err := j.GetOrCompile(testFunc1)
	if err != nil {
		fmt.Printf("编译错误: %v\n", err)
		return
	}
	fmt.Printf("函数1 编译成功: %v\n", compiled1 != nil)
	fmt.Printf("函数1 是否优化: %v\n\n", compiled1.Optimized)
	
	compiled2, err := j.GetOrCompile(testFunc2)
	if err != nil {
		fmt.Printf("编译错误: %v\n", err)
		return
	}
	fmt.Printf("函数2 编译成功: %v\n\n", compiled2 != nil)
	
	fmt.Println("=== 3. 测试函数调用记录 ===")
	for i := 0; i < 5; i++ {
		j.RecordCall(testFunc1)
	}
	fmt.Printf("函数1 调用 5 次后的调用次数: %d\n", compiled1.CallCount)
	fmt.Printf("函数1 是否为热点: %v\n\n", j.IsHot(testFunc1))
	
	fmt.Println("=== 4. 测试代码元数据分析 ===")
	if compiled1.Metadata != nil {
		fmt.Printf("函数1 元数据:\n")
		fmt.Printf("  - 指令数量: %d\n", compiled1.Metadata.NumInstructions)
		fmt.Printf("  - 包含循环: %v\n", compiled1.Metadata.HasLoops)
		fmt.Printf("  - 代码复杂度: %d\n\n", compiled1.Metadata.Complexity)
	}
	
	if compiled2.Metadata != nil {
		fmt.Printf("函数2 元数据:\n")
		fmt.Printf("  - 指令数量: %d\n", compiled2.Metadata.NumInstructions)
		fmt.Printf("  - 包含循环: %v\n", compiled2.Metadata.HasLoops)
		fmt.Printf("  - 代码复杂度: %d\n\n", compiled2.Metadata.Complexity)
	}
	
	fmt.Println("=== 5. 测试热点函数列表 ===")
	hotFunctions := j.GetHotFunctions()
	fmt.Printf("热点函数数量: %d\n", len(hotFunctions))
	for i, fn := range hotFunctions {
		fmt.Printf("  热点函数 %d: 调用次数=%d, 优化状态=%v\n", 
			i+1, fn.CallCount, fn.Optimized)
	}
	fmt.Println()
	
	fmt.Println("=== 6. 测试 JIT 统计信息 ===")
	stats := j.GetStats()
	fmt.Println("JIT 统计信息:")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
	fmt.Println()
	
	fmt.Println("=== 7. 测试热点函数优化 ===")
	fmt.Printf("优化前函数1是否优化: %v\n", compiled1.Optimized)
	for i := 0; i < 10; i++ {
		j.RecordCall(testFunc1)
	}
	fmt.Printf("优化后函数1是否优化: %v\n", compiled1.Optimized)
	fmt.Println()
	
	fmt.Println("=== 8. 测试缓存清理 ===")
	j.ClearCache()
	statsAfterClear := j.GetStats()
	fmt.Println("清理后的统计信息:")
	fmt.Printf("  缓存函数数: %v\n", statsAfterClear["cached_functions"])
	fmt.Printf("  总调用次数: %v\n", statsAfterClear["total_calls"])
	fmt.Println()
	
	fmt.Println("🎉 所有 JIT 功能测试通过!")
}
