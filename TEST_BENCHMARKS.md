# GoPy Benchmark Test Suite

This is a comprehensive benchmark test suite for comparing GoPy performance against CPython.

## What This Includes

### Benchmark Categories

1. **Arithmetic Operations** (`001_arithmetic.py`) - Basic arithmetic, floating point, fibonacci, prime calculations
2. **Loop & Control Flow** (`002_loop.py`) - While loops, conditionals, nested loops, list iteration
3. **Data Structures** (`003_list.py`) - List, set, and list comprehension operations
4. **Dictionary Operations** (`004_dict.py`) - Dict creation, access, iteration, nested dicts
5. **Function Calls** (`005_function.py`) - Simple calls, multi-parameter, recursive functions
6. **String Operations** (`006_string.py`) - Concatenation, formatting, slicing, comparison
7. **Factorial Calculations** (`007_factorial.py`) - Iterative and recursive factorial implementations
8. **Matrix Operations** (`008_matrix.py`) - 2D arrays, matrix addition, scalar multiplication
9. **Recursive Functions** (`009_recursive.py`) - Fibonacci (various implementations), Ackermann, etc.
10. **Simple Tests** (`000_test.py`, `000_function_test.py`, etc.) - Basic functionality verification

## Running the Benchmarks

### Quick Start

```bash
# Run the full benchmark suite (compares GoPy JIT, GoPy NoJIT, and CPython)
go run cmd/benchmark/main.go
```

### Configuration

Edit `cmd/benchmark/main.go` to adjust:
- `warmupRuns`: Number of warmup runs (default: 2)
- `measurementRuns`: Number of measurement runs (default: 5)
- `benchmarkDir`: Directory containing test files (default: `tests/benchmarks`)

## Output

The benchmark will produce:
1. **Console Output** - Detailed comparison table showing times for each implementation
2. **CSV File** - Results saved to `benchmark_results_[timestamp].csv` for further analysis

## Running Individual Tests

You can also run individual test files with each implementation:

```bash
# GoPy with JIT enabled
go run cmd/gopy/main.go --jit tests/benchmarks/001_arithmetic.py

# GoPy without JIT
go run cmd/gopy/main.go tests/benchmarks/001_arithmetic.py

# CPython
python3 tests/benchmarks/001_arithmetic.py
```

## Interpreting Results

The comparison table shows:
- `GoPy (JIT)`: Time taken with JIT compilation enabled
- `GoPy (NoJIT)`: Time taken with JIT disabled
- `CPython`: Time taken with standard Python
- `GoPyJIT/CPython`: Speed ratio (lower = faster)
- `GoPyJIT/GoPyNoJIT`: JIT speedup factor (lower = more speedup from JIT)

## Example Results

(Your actual results will vary depending on hardware)

```
====================================================================================================
基准测试详细对比结果
====================================================================================================
测试名称                    GoPy (JIT)           GoPy (NoJIT)         CPython              GoPyJIT/CPython  GoPyJIT/GoPyNoJIT
----------------------------------------------------------------------------------------------------
000_function_test.py       5.234ms              12.456ms             3.123ms                   1.68x            0.42x
001_arithmetic.py          89.123ms             245.678ms            45.234ms                  1.97x            0.36x
...
```

## Notes

- Benchmarks should be run multiple times for consistent results
- First run warmup is important, especially for JIT compilation
- Some operations may be faster in CPython due to optimized C implementations
