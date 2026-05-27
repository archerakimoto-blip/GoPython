"""
Factorial Calculation Benchmark
Tests iterative and recursive factorial implementations
"""

def factorial_iterative(n):
    result = 1
    i = 1
    while i <= n:
        result = result * i
        i = i + 1
    return result

def factorial_recursive(n):
    if n <= 1:
        return 1
    return n * factorial_recursive(n - 1)

def test_iterative_factorial():
    results = []
    i = 1
    while i <= 10:
        results.append(factorial_iterative(i * 50))
        i = i + 1
    return results

def test_recursive_factorial():
    results = []
    i = 1
    while i <= 10:
        results.append(factorial_recursive(i * 10))
        i = i + 1
    return results

def test_mixed_operations():
    total = 0
    i = 1
    while i <= 20:
        fact = factorial_iterative(i)
        total = total + fact
        i = i + 1
    return total

# Run benchmarks
print("=== 阶乘计算基准测试 ===")
print()

print("1. 迭代阶乘 (10个计算，最大n=500):")
iter_results = test_iterative_factorial()
print("  最后一个结果长度:", len(str(iter_results[-1])))

print()
print("2. 递归阶乘 (10个计算，最大n=100):")
recursive_results = test_recursive_factorial()
print("  最后一个结果长度:", len(str(recursive_results[-1])))

print()
print("3. 混合操作 (20个计算求和):")
mixed_total = test_mixed_operations()
print("  结果长度:", len(str(mixed_total)))

print()
print("=== 阶乘测试完成 ===")
