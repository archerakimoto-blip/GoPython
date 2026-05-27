"""
String Operations Benchmark
Tests string concatenation, formatting, and manipulation
"""

def test_string_concatenation(n):
    s = ""
    i = 0
    while i < n:
        s = s + str(i)
        i = i + 1
    return len(s)

def test_string_formatting(n):
    s = ""
    i = 0
    while i < n:
        s = "Value: " + str(i) + ", Next: " + str(i + 1)
        i = i + 1
    return len(s)

def test_string_slicing(n):
    base = "abcdefghijklmnopqrstuvwxyz" * 100
    i = 0
    total = 0
    while i < n:
        start = i % len(base)
        end = min(start + 10, len(base))
        part = base[start:end]
        total = total + len(part)
        i = i + 1
    return total

def test_string_comparison(n):
    count = 0
    i = 0
    while i < n:
        s1 = "string" + str(i)
        s2 = "string" + str(i + 1)
        if s1 < s2:
            count = count + 1
        i = i + 1
    return count

def test_list_join(n):
    lst = []
    i = 0
    while i < n:
        lst.append(str(i))
        i = i + 1
    result = "".join(lst)
    return len(result)

# Run benchmarks
print("=== 字符串操作基准测试 ===")
print()

print("1. 字符串拼接 (10,000次):")
_ = test_string_concatenation(10000)
print("  完成")

print()
print("2. 字符串格式化 (5,000次):")
_ = test_string_formatting(5000)
print("  完成")

print()
print("3. 字符串切片 (10,000次):")
_ = test_string_slicing(10000)
print("  完成")

print()
print("4. 字符串比较 (10,000次):")
_ = test_string_comparison(10000)
print("  完成")

print()
print("5. 列表join (10,000元素):")
_ = test_list_join(10000)
print("  完成")

print()
print("=== 字符串测试完成 ===")
