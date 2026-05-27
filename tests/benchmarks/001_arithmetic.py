"""
Arithmetic and Mathematical Operations Benchmark
Tests various mathematical operations, loops, and calculations
"""

def test_basic_arithmetic(n):
    result = 0
    i = 0
    while i < n:
        result = result + i
        result = result * 2
        result = result - (i % 3)
        i = i + 1
    return result

def test_floating_point(n):
    result = 0.0
    i = 0
    while i < n:
        result = result + 1.5 * i
        result = result / (i + 1)
        i = i + 1
    return result

def test_fibonacci(n):
    a, b = 0, 1
    i = 2
    while i <= n:
        a, b = b, a + b
        i = i + 1
    return b

def test_prime_sum(limit):
    primes = []
    num = 2
    while num < limit:
        is_prime = True
        i = 2
        while i * i <= num:
            if num % i == 0:
                is_prime = False
                break
            i = i + 1
        if is_prime:
            primes.append(num)
        num = num + 1
    total = 0
    for p in primes:
        total = total + p
    return total

# Run benchmarks
print("=== 算术运算基准测试 ===")
print()

print("1. 基础算术运算 (1,000,000次):")
start_basic = 0
_ = test_basic_arithmetic(1000000)
print("  完成")

print()
print("2. 浮点运算 (100,000次):")
_ = test_floating_point(100000)
print("  完成")

print()
print("3. 斐波那契数列 (n=40):")
fib_result = test_fibonacci(40)
print("  结果:", fib_result)

print()
print("4. 素数求和 (limit=2000):")
prime_sum = test_prime_sum(2000)
print("  结果:", prime_sum)

print()
print("=== 算术测试完成 ===")
