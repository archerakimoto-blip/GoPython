package benchmarks

import (
	"time"
)

type LoopBenchmark struct{}

func (b *LoopBenchmark) Name() string {
	return "Loop Operations"
}

func (b *LoopBenchmark) Run() BenchmarkResult {
	const iterations = 1000000

	start := time.Now()
	sum := 0

	for i := 0; i < iterations; i++ {
		for j := 0; j < 100; j++ {
			sum += i + j
		}
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