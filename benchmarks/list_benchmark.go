package benchmarks

import (
	"time"
)

type ListBenchmark struct{}

func (b *ListBenchmark) Name() string {
	return "List Operations"
}

func (b *ListBenchmark) Run() BenchmarkResult {
	const iterations = 100000

	start := time.Now()

	for i := 0; i < iterations; i++ {
		list := make([]int, 0, 100)
		for j := 0; j < 100; j++ {
			list = append(list, j)
		}
		_ = list[50]
		_ = len(list)
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