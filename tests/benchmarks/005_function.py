print("=== 函数调用基准测试 ===")

print()
print("1. Simple function calls (10000):")

def simple_add(a, b):
    return a + b

result = 0
i = 0
while i < 10000:
    result = simple_add(i, i + 1)
    i = i + 1
print(result)

print()
print("2. Multiple parameters function:")

def multiply(x, y, z):
    return x * y * z

result = multiply(2, 3, 4)
print(result)

print()
print("3. Recursive function (fibonacci-like, n=10):")

def recursive_func(n):
    if n <= 0:
        return 1
    return recursive_func(n-1) + n

result = recursive_func(10)
print(result)

print()
print("=== 函数调用测试完成 ===")
