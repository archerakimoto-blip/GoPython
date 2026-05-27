package benchmarks

import (
	"fmt"
	"strings"
	"time"
)

type StringBenchmark struct{}

func (b *StringBenchmark) Name() string {
	return "String Operations"
}

func (b *StringBenchmark) Run() BenchmarkResult {
	const iterations = 50000

	start := time.Now()

	for i := 0; i < iterations; i++ {
		s := "hello" + "world" + fmt.Sprintf("%d", i)
		_ = strings.Contains(s, "world")
		_ = len(s)
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