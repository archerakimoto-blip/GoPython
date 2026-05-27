package benchmarks

import (
	"time"
)

type ArithmeticBenchmark struct{}

func (b *ArithmeticBenchmark) Name() string {
	return "Arithmetic Operations"
}

func (b *ArithmeticBenchmark) Run() BenchmarkResult {
	const iterations = 10000000

	start := time.Now()
	var total int64 = 0

	for i := 0; i < iterations; i++ {
		x := int64(i)
		y := int64(i * 2)
		total += (x + y) * (x - y) / 2
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