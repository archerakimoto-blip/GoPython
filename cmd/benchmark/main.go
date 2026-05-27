package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Implementation string

const (
	GoPyJIT    Implementation = "gopy-jit"
	GoPyNoJIT  Implementation = "gopy-nojit"
	CPython    Implementation = "cpython"
)

type BenchmarkResult struct {
	Name   string
	Status string
	Time   time.Duration
	Error  error
}

type BenchmarkComparison struct {
	Name     string
	Results  map[Implementation]BenchmarkResult
}

func runBenchmark(filePath string, impl Implementation) BenchmarkResult {
	var cmd *exec.Cmd

	switch impl {
	case GoPyJIT:
		cmd = exec.Command("go", "run", "cmd/gopy/main.go", "--jit", filePath)
	case GoPyNoJIT:
		cmd = exec.Command("go", "run", "cmd/gopy/main.go", filePath)
	case CPython:
		cmd = exec.Command("python3", filePath)
	}

	// 不打印输出，只执行
	cmd.Stdout = nil
	cmd.Stderr = nil

	start := time.Now()
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

func runComparison(filePath string, warmups int, runs int) BenchmarkComparison {
	comparison := BenchmarkComparison{
		Name:    filepath.Base(filePath),
		Results: make(map[Implementation]BenchmarkResult),
	}

	implementations := []Implementation{GoPyJIT, GoPyNoJIT, CPython}

	for _, impl := range implementations {
		var totalTime time.Duration
		var passCount int
		var lastErr error

		// Warmup runs
		for i := 0; i < warmups; i++ {
			_ = runBenchmark(filePath, impl)
		}

		// Measurement runs
		for i := 0; i < runs; i++ {
			result := runBenchmark(filePath, impl)
			if result.Status == "PASSED" {
				totalTime += result.Time
				passCount++
			} else {
				lastErr = result.Error
			}
		}

		if passCount > 0 {
			avgTime := totalTime / time.Duration(passCount)
			comparison.Results[impl] = BenchmarkResult{
				Name:   filepath.Base(filePath),
				Status: "PASSED",
				Time:   avgTime,
			}
		} else {
			comparison.Results[impl] = BenchmarkResult{
				Name:   filepath.Base(filePath),
				Status: "FAILED",
				Time:   0,
				Error:  lastErr,
			}
		}
	}

	return comparison
}

func printComparisonResults(comparisons []BenchmarkComparison) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("基准测试详细对比结果")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("%-25s %-20s %-20s %-20s %-15s %-15s\n", 
		"测试名称", "GoPy (JIT)", "GoPy (NoJIT)", "CPython", "GoPyJIT/CPython", "GoPyJIT/GoPyNoJIT")
	fmt.Println(strings.Repeat("-", 100))

	var totalGoPyJIT time.Duration
	var totalGoPyNoJIT time.Duration
	var totalCPython time.Duration
	var count int

	for _, comp := range comparisons {
		jitRes := comp.Results[GoPyJIT]
		nojitRes := comp.Results[GoPyNoJIT]
		cpyRes := comp.Results[CPython]

		var jitVsCPython float64
		var jitVsNoJIT float64

		if jitRes.Status == "PASSED" && cpyRes.Status == "PASSED" {
			jitVsCPython = float64(jitRes.Time) / float64(cpyRes.Time)
			totalGoPyJIT += jitRes.Time
			totalCPython += cpyRes.Time
			count++
		}
		if jitRes.Status == "PASSED" && nojitRes.Status == "PASSED" {
			jitVsNoJIT = float64(jitRes.Time) / float64(nojitRes.Time)
			totalGoPyNoJIT += nojitRes.Time
		}

		fmt.Printf("%-25s %-20s %-20s %-20s ",
			comp.Name,
			formatTime(jitRes),
			formatTime(nojitRes),
			formatTime(cpyRes),
		)

		if jitVsCPython > 0 {
			fmt.Printf("%13.2fx ", jitVsCPython)
		} else {
			fmt.Printf("%13s ", "N/A")
		}

		if jitVsNoJIT > 0 {
			fmt.Printf("%15.2fx\n", jitVsNoJIT)
		} else {
			fmt.Printf("%15s\n", "N/A")
		}
	}

	fmt.Println(strings.Repeat("-", 100))

	if count > 0 {
		avgJitVsCPython := float64(totalGoPyJIT) / float64(totalCPython)
		avgJitVsNoJIT := float64(totalGoPyJIT) / float64(totalGoPyNoJIT)
		fmt.Printf("%-25s %-20v %-20v %-20v %13.2fx %15.2fx\n",
			"平均值",
			totalGoPyJIT/time.Duration(count),
			totalGoPyNoJIT/time.Duration(count),
			totalCPython/time.Duration(count),
			avgJitVsCPython,
			avgJitVsNoJIT,
		)
	}
}

func formatTime(res BenchmarkResult) string {
	if res.Status == "FAILED" {
		return "FAILED"
	}
	return res.Time.String()
}

func writeCSV(comparisons []BenchmarkComparison, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{
		"Test Name",
		"GoPy (JIT) (ms)",
		"GoPy (NoJIT) (ms)",
		"CPython (ms)",
		"GoPyJIT vs CPython",
		"GoPyJIT vs GoPyNoJIT",
	}
	writer.Write(header)

	for _, comp := range comparisons {
		jitRes := comp.Results[GoPyJIT]
		nojitRes := comp.Results[GoPyNoJIT]
		cpyRes := comp.Results[CPython]

		row := []string{comp.Name}

		if jitRes.Status == "PASSED" {
			row = append(row, fmt.Sprintf("%.2f", float64(jitRes.Time.Microseconds())/1000.0))
		} else {
			row = append(row, "FAILED")
		}

		if nojitRes.Status == "PASSED" {
			row = append(row, fmt.Sprintf("%.2f", float64(nojitRes.Time.Microseconds())/1000.0))
		} else {
			row = append(row, "FAILED")
		}

		if cpyRes.Status == "PASSED" {
			row = append(row, fmt.Sprintf("%.2f", float64(cpyRes.Time.Microseconds())/1000.0))
		} else {
			row = append(row, "FAILED")
		}

		if jitRes.Status == "PASSED" && cpyRes.Status == "PASSED" {
			ratio := float64(jitRes.Time) / float64(cpyRes.Time)
			row = append(row, fmt.Sprintf("%.2fx", ratio))
		} else {
			row = append(row, "N/A")
		}

		if jitRes.Status == "PASSED" && nojitRes.Status == "PASSED" {
			ratio := float64(jitRes.Time) / float64(nojitRes.Time)
			row = append(row, fmt.Sprintf("%.2fx", ratio))
		} else {
			row = append(row, "N/A")
		}

		writer.Write(row)
	}

	return nil
}

func main() {
	fmt.Println("=")
	fmt.Println("      GoPy 全面基准测试套件")
	fmt.Println("=")
	fmt.Printf("开始时间: %s\n\n", time.Now().Format(time.RFC3339))

	// Configuration
	warmupRuns := 2
	measurementRuns := 5
	benchmarkDir := "tests/benchmarks"
	csvOutput := fmt.Sprintf("benchmark_results_%d.csv", time.Now().Unix())

	fmt.Println("配置:")
	fmt.Printf("  预热运行次数: %d\n", warmupRuns)
	fmt.Printf("  测量运行次数: %d\n", measurementRuns)
	fmt.Printf("  测试目录: %s\n", benchmarkDir)
	fmt.Println()

	files, err := filepath.Glob(filepath.Join(benchmarkDir, "*.py"))
	if err != nil {
		fmt.Printf("错误: 找不到测试文件: %v\n", err)
		os.Exit(1)
	}

	var comparisons []BenchmarkComparison

	for i, file := range files {
		fmt.Printf("[%d/%d] 正在测试: %s\n", i+1, len(files), filepath.Base(file))
		comp := runComparison(file, warmupRuns, measurementRuns)
		comparisons = append(comparisons, comp)
	}

	printComparisonResults(comparisons)

	fmt.Println("\n正在写入CSV结果文件:", csvOutput)
	if err := writeCSV(comparisons, csvOutput); err != nil {
		fmt.Printf("警告: 写入CSV失败: %v\n", err)
	}

	fmt.Println("\n测试完成!")
	fmt.Printf("完成时间: %s\n", time.Now().Format(time.RFC3339))
}
