# 测试 type() 函数
print("=== Testing type() ===")
print(type(42))
print(type(3.14))
print(type("hello"))
print(type(True))
print(type([1, 2, 3]))
print(type({"a": 1}))

# 测试 str() 函数
print("\n=== Testing str() ===")
print(str(42))
print(str(3.14))
print(str(True))
print(str([1, 2, 3]))

# 测试 int() 函数
print("\n=== Testing int() ===")
print(int(42))
print(int(3.14))
print(int("10"))
print(int(True))
print(int(False))

# 测试 float() 函数
print("\n=== Testing float() ===")
print(float(42))
print(float("3.14"))
print(float(True))

# 测试 bool() 函数
print("\n=== Testing bool() ===")
print(bool(0))
print(bool(1))
print(bool(""))
print(bool("hello"))
print(bool([]))
print(bool([1]))

# 测试 abs() 函数
print("\n=== Testing abs() ===")
print(abs(-5))
print(abs(5))
print(abs(-3.14))

# 测试 range() 函数
print("\n=== Testing range() ===")
print(range(5))
print(range(1, 5))
print(range(0, 10, 2))

# 测试 min() 函数
print("\n=== Testing min() ===")
print(min(1, 2, 3))
print(min(10, 5, 8))

# 测试 max() 函数
print("\n=== Testing max() ===")
print(max(1, 2, 3))
print(max(10, 5, 8))

# 测试 sum() 函数
print("\n=== Testing sum() ===")
print(sum([1, 2, 3, 4, 5]))
print(sum(range(1, 101)))
