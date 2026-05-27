"""
Data Structure and Collection Benchmark
Tests list, set, and dictionary operations
"""

def test_list_operations(size):
    lst = []
    i = 0
    while i < size:
        lst.append(i)
        i = i + 1
    
    # Access
    total = 0
    i = 0
    while i < size:
        total = total + lst[i]
        i = i + 1
    
    # Update
    i = 0
    while i < size:
        lst[i] = lst[i] * 2
        i = i + 1
    
    return total

def test_dict_operations(size):
    d = {}
    i = 0
    while i < size:
        d[str(i)] = i
        i = i + 1
    
    total = 0
    i = 0
    while i < size:
        total = total + d[str(i)]
        i = i + 1
    
    return total

def test_set_operations(size):
    s = set()
    i = 0
    while i < size:
        s.add(i)
        i = i + 1
    
    count = 0
    i = 0
    while i < size * 2:
        if i in s:
            count = count + 1
        i = i + 1
    
    return count

def test_list_comprehensions(n):
    lst = [i * 2 for i in range(n)]
    total = 0
    for item in lst:
        total = total + item
    return total

# Run benchmarks
print("=== 数据结构基准测试 ===")
print()

print("1. 列表操作 (10,000元素):")
_ = test_list_operations(10000)
print("  完成")

print()
print("2. 字典操作 (5,000键值对):")
_ = test_dict_operations(5000)
print("  完成")

print()
print("3. 集合操作 (2,000元素):")
_ = test_set_operations(2000)
print("  完成")

print()
print("4. 列表推导式 (10,000元素):")
_ = test_list_comprehensions(10000)
print("  完成")

print()
print("=== 数据结构测试完成 ===")
