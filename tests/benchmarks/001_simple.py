print("=== GoPy 简单基准测试 ===")

print()
print("1. Simple loop (100000 iterations):")

total = 0
i = 0
while i < 100000:
    total = total + 1
    i = i + 1
print(total)

print()
print("2. Fibonacci(20):")

def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)

fib_result = fib(20)
print(fib_result)

print()
print("3. List operations:")

lst = []
i = 0
while i < 1000:
    lst.append(i)
    i = i + 1
print(len(lst))

print()
print("=== 测试完成 ===")
