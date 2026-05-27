def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

def factorial(n):
    result = 1
    i = 1
    while i <= n:
        result = result * i
        i = i + 1
    return result

print("=== 算术运算基准测试 ===")

print()
print("1. Fibonacci(20):")
fib_result = fibonacci(20)
print("   Result:", fib_result)

print()
print("2. Factorial(15):")
fact_result = factorial(15)
print("   Result:", fact_result)

print()
print("3. Loop calculation (10000 iterations):")
total = 0
i = 0
while i < 10000:
    total = total + i * 2
    i = i + 1
print("   Total:", total)

print()
print("=== 算术运算测试完成 ===")
