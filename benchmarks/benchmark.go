package benchmarks

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type BenchmarkResult struct {
	Name         string
	Iterations   int
	TotalTime    time.Duration
	AverageTime  time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	Throughput   float64
}

type Benchmark interface {
	Name() string
	Run() BenchmarkResult
}

func (r BenchmarkResult) String() string {
	return fmt.Sprintf("%s: %d iterations, avg=%v, min=%v, max=%v, throughput=%.2f ops/s",
		r.Name, r.Iterations, r.AverageTime, r.MinTime, r.MaxTime, r.Throughput)
}

func RunAll() []BenchmarkResult {
	benchmarks := []Benchmark{
		&ArithmeticBenchmark{},
		&LoopBenchmark{},
		&ListBenchmark{},
		&DictBenchmark{},
		&FunctionBenchmark{},
		&StringBenchmark{},
	}

	var results []BenchmarkResult
	for _, b := range benchmarks {
		fmt.Printf("Running benchmark: %s...\n", b.Name())
		result := b.Run()
		fmt.Println(result)
		results = append(results, result)
	}

	return results
}

func WriteResults(results []BenchmarkResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("Benchmark Results\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString("=" + strings.Repeat("=", 60) + "\n")
	if err != nil {
		return err
	}

	for _, result := range results {
		_, err = file.WriteString(fmt.Sprintf("%s\n", result.String()))
		if err != nil {
			return err
		}
	}

	return nil
}
