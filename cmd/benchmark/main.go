package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type BenchmarkResult struct {
	Name     string
	Status   string
	Time     time.Duration
	Error    error
}

func runBenchmark(filePath string) BenchmarkResult {
	fmt.Printf("Running benchmark (JIT enabled): %s\n", filePath)
	
	start := time.Now()
	
	cmd := exec.Command("./gopy", "--jit", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	
	elapsed := time.Since(start)
	
	if err != nil {
		return BenchmarkResult{
			Name:   filepath.Base(filePath),
			Status: "FAILED",
			Time:   elapsed,
			Error:  err,
		}
	}
	
	return BenchmarkResult{
		Name:   filepath.Base(filePath),
		Status: "PASSED",
		Time:   elapsed,
	}
}

func main() {
	fmt.Println("=== GoPy 基准测试套件 ===")
	fmt.Printf("开始时间: %s\n\n", time.Now().Format(time.RFC3339))
	
	benchmarkDir := "tests/benchmarks"
	
	files, err := filepath.Glob(filepath.Join(benchmarkDir, "*.py"))
	if err != nil {
		fmt.Printf("Error finding benchmark files: %v\n", err)
		os.Exit(1)
	}
	
	var results []BenchmarkResult
	totalTime := time.Duration(0)
	
	for _, file := range files {
		fmt.Println("----------------------------------------")
		result := runBenchmark(file)
		results = append(results, result)
		totalTime += result.Time
		fmt.Printf("\n%s 完成，耗时: %v\n\n", result.Name, result.Time)
	}
	
	// 打印总结
	fmt.Println("========================================")
	fmt.Println("基准测试总结")
	fmt.Println("========================================")
	
	for _, result := range results {
		fmt.Printf("%-25s %-8s %10v\n", result.Name, result.Status, result.Time)
		if result.Error != nil {
			fmt.Printf("  错误: %v\n", result.Error)
		}
	}
	
	fmt.Println("----------------------------------------")
	fmt.Printf("总耗时: %v\n", totalTime)
	fmt.Printf("完成时间: %s\n", time.Now().Format(time.RFC3339))
	fmt.Println("========================================")
}
