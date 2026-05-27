"""
Dictionary Operations Benchmark
Tests various dictionary operations including nested structures
"""

def test_dict_creation(size):
    d = {}
    i = 0
    while i < size:
        d["key_" + str(i)] = i
        i = i + 1
    return d

def test_dict_get_put(d, size):
    total = 0
    i = 0
    while i < size:
        key = "key_" + str(i)
        total = total + d[key]
        d[key] = i * 2
        i = i + 1
    return total

def test_dict_iteration(d):
    count = 0
    for key in d:
        count = count + 1
    return count

def test_nested_dicts(depth, width):
    def create_nested(level):
        if level == 0:
            return 42
        d = {}
        i = 0
        while i < width:
            d["level" + str(level) + "_key" + str(i)] = create_nested(level - 1)
            i = i + 1
        return d
    
    nested = create_nested(depth)
    
    def count_elements(obj):
        if isinstance(obj, dict):
            c = 0
            for k in obj:
                c = c + count_elements(obj[k])
            return c
        return 1
    
    return count_elements(nested)

def test_dict_comprehension(n):
    d = {i: i * i for i in range(n)}
    total = 0
    for key in d:
        total = total + d[key]
    return total

# Run benchmarks
print("=== 字典操作基准测试 ===")
print()

print("1. 字典创建 (10,000键):")
d = test_dict_creation(10000)
print("  完成")

print()
print("2. 字典读写 (10,000次):")
_ = test_dict_get_put(d, 10000)
print("  完成")

print()
print("3. 字典迭代 (10,000键):")
_ = test_dict_iteration(d)
print("  完成")

print()
print("4. 嵌套字典 (4层, 每层5个键):")
_ = test_nested_dicts(4, 5)
print("  完成")

print()
print("5. 字典推导式 (5,000元素):")
_ = test_dict_comprehension(5000)
print("  完成")

print()
print("=== 字典测试完成 ===")
