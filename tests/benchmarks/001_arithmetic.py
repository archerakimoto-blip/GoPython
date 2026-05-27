print("=== 算术运算基准测试 ===")

print()
print("1. Fibonacci(15):")

def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)

fib_result = fib(15)
print(fib_result)

print()
print("2. Factorial(10):")

def fact(n):
    result = 1
    i = 1
    while i <= n:
        result = result * i
        i = i + 1
    return result

fact_result = fact(10)
print(fact_result)

print()
print("3. Loop calculation (10000 iterations):")

total = 0
i = 0
while i < 10000:
    total = total + i * 2
    i = i + 1
print(total)

print()
print("=== 算术运算测试完成 ===")
