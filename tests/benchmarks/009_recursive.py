"""
Recursive Function Calls Benchmark
Tests various recursive function implementations
"""

def fibonacci_recursive(n):
    if n <= 1:
        return n
    return fibonacci_recursive(n - 1) + fibonacci_recursive(n - 2)

def fibonacci_memoized(n):
    memo = {}
    
    def helper(k):
        if k in memo:
            return memo[k]
        if k <= 1:
            result = k
        else:
            result = helper(k - 1) + helper(k - 2)
        memo[k] = result
        return result
    
    return helper(n)

def fibonacci_iterative(n):
    a, b = 0, 1
    i = 0
    while i < n:
        a, b = b, a + b
        i = i + 1
    return a

def recursive_sum(n):
    if n <= 0:
        return 0
    return n + recursive_sum(n - 1)

def recursive_countdown(n):
    if n <= 0:
        return 0
    return 1 + recursive_countdown(n - 1)

def ackermann(m, n):
    if m == 0:
        return n + 1
    elif n == 0:
        return ackermann(m - 1, 1)
    else:
        return ackermann(m - 1, ackermann(m, n - 1))

# Run benchmarks
print("=== 递归函数调用基准测试 ===")
print()

print("1. 递归斐波那契 (n=25):")
fib_rec = fibonacci_recursive(25)
print("  结果:", fib_rec)

print()
print("2. 记忆化斐波那契 (n=100):")
fib_memo = fibonacci_memoized(100)
print("  结果长度:", len(str(fib_memo)))

print()
print("3. 迭代斐波那契 (n=100):")
fib_iter = fibonacci_iterative(100)
print("  结果长度:", len(str(fib_iter)))

print()
print("4. 递归求和 (n=1000):")
sum_result = recursive_sum(1000)
print("  结果:", sum_result)

print()
print("5. 递归倒计时 (n=500):")
count_result = recursive_countdown(500)
print("  结果:", count_result)

print()
print("6. Ackermann函数 (m=3, n=5):")
ack_result = ackermann(3, 5)
print("  结果:", ack_result)

print()
print("=== 递归测试完成 ===")
