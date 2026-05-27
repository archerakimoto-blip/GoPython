package benchmarks

import (
	"fmt"
	"time"
)

type DictBenchmark struct{}

func (b *DictBenchmark) Name() string {
	return "Dict Operations"
}

func (b *DictBenchmark) Run() BenchmarkResult {
	const iterations = 50000

	start := time.Now()

	for i := 0; i < iterations; i++ {
		dict := make(map[string]int)
		for j := 0; j < 50; j++ {
			dict[fmt.Sprintf("key%d", j)] = j
		}
		_ = dict["key25"]
		_ = len(dict)
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