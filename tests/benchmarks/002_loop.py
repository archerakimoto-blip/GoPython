"""
Loop and Control Flow Benchmark
Tests various loop structures and control flow operations
"""

def test_while_loop(n):
    result = 0
    i = 0
    while i < n:
        result = result + i
        i = i + 1
    return result

def test_conditionals(n):
    result = 0
    i = 0
    while i < n:
        if i % 2 == 0:
            result = result + i
        elif i % 3 == 0:
            result = result + i * 2
        else:
            result = result - i
        i = i + 1
    return result

def test_nested_loops(size):
    result = 0
    i = 0
    while i < size:
        j = 0
        while j < size:
            result = result + i * j
            j = j + 1
        i = i + 1
    return result

def test_list_iteration(n):
    lst = []
    i = 0
    while i < n:
        lst.append(i)
        i = i + 1
    total = 0
    for item in lst:
        total = total + item
    return total

# Run benchmarks
print("=== 循环和控制流基准测试 ===")
print()

print("1. 简单while循环 (1,000,000次):")
_ = test_while_loop(1000000)
print("  完成")

print()
print("2. 条件判断 (100,000次):")
_ = test_conditionals(100000)
print("  完成")

print()
print("3. 嵌套循环 (200x200):")
_ = test_nested_loops(200)
print("  完成")

print()
print("4. 列表迭代 (10,000个元素):")
_ = test_list_iteration(10000)
print("  完成")

print()
print("=== 循环测试完成 ===")
