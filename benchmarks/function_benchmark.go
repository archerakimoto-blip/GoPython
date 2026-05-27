package benchmarks

import (
	"time"
)

type FunctionBenchmark struct{}

func (b *FunctionBenchmark) Name() string {
	return "Function Calls"
}

func (b *FunctionBenchmark) Run() BenchmarkResult {
	const iterations = 1000000

	start := time.Now()

	for i := 0; i < iterations; i++ {
		_ = benchmarkAdd(i, i*2)
	}

	elapsed := time.Since(start)

	return BenchmarkResult{
		Name:        b.Name(),
		Iterations:  iterations,
		TotalTime:   elapsed,
		AverageTime: elapsed / time.Duration(iterations),
		MinTime:     elapsed / time.Duration(iterations),
		MaxTime:     elapsed / time.Duration(iterations),
		Throughput:  float64(iterations) / elapsed.Seconds(),
	}
}

func benchmarkAdd(a, b int) int {
	return a + b
}