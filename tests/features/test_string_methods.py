# String Methods Tests
# 测试字符串方法

print("Testing string methods...")

# 测试基本字符串方法
print("\n1. Testing string transformation methods:")

s = "  hello World  "
print(f"Original: '{s}'")
print(f"Upper: '{s.upper()}'")
print(f"Lower: '{s.lower()}'")
print(f"Capitalize: '{s.capitalize()}'")
print(f"Title: '{s.title()}'")
print(f"Swapcase: '{s.swapcase()}'")

# 测试 strip 方法
print("\n2. Testing strip methods:")
print(f"Strip: '{s.strip()}'")
print(f"Lstrip: '{s.lstrip()}'")
print(f"Rstrip: '{s.rstrip()}'")

# 测试 startswith/endswith
print("\n3. Testing startswith and endswith:")
s2 = "test.txt"
print(f"String: '{s2}'")
print(f"startswith('test'): {s2.startswith('test')}")
print(f"startswith('TEST'): {s2.startswith('TEST')}")
print(f"endswith('.txt'): {s2.endswith('.txt')}")
print(f"endswith('.py'): {s2.endswith('.py')}")

# 测试 find
print("\n4. Testing find:")
s3 = "The quick brown fox jumps over the lazy dog"
print(f"String: '{s3}'")
print(f"find('quick'): {s3.find('quick')}")
print(f"find('Quick'): {s3.find('Quick')}")
print(f"find('fox', 10): {s3.find('fox', 10)}")

# 测试 replace
print("\n5. Testing replace:")
s4 = "Hello World, World!"
print(f"Original: '{s4}'")
print(f"Replace 'World' with 'GoPython': '{s4.replace('World', 'GoPython')}'")
print(f"Replace first 'World': '{s4.replace('World', 'GoPython', 1)}'")

# 测试 split
print("\n6. Testing split:")
s5 = "apple,banana,cherry,date"
print(f"Original: '{s5}'")
result = s5.split(',')
print(f"Split by ',': {result}")
print(f"Split by ',' with maxsplit 2: {s5.split(',', 2)}")

# 测试 join
print("\n7. Testing join:")
parts = ['Hello', 'from', 'GoPython']
joined = ' '.join(parts)
print(f"Parts: {parts}")
print(f"Joined: '{joined}'")

# 测试 is* 方法
print("\n8. Testing is* methods:")
s6 = "Hello"
print(f"String: '{s6}'")
print(f"isalpha: {s6.isalpha()}")
print(f"isdigit: {s6.isdigit()}")
print(f"isspace: {s6.isspace()}")
print(f"isupper: {s6.isupper()}")
print(f"islower: {s6.islower()}")

s7 = "12345"
print(f"\nString: '{s7}'")
print(f"isalpha: {s7.isalpha()}")
print(f"isdigit: {s7.isdigit()}")

s8 = "   "
print(f"\nString: '{s8}'")
print(f"isspace: {s8.isspace()}")

s9 = "HELLO"
print(f"\nString: '{s9}'")
print(f"isupper: {s9.isupper()}")
print(f"islower: {s9.islower()}")

s10 = "hello"
print(f"\nString: '{s10}'")
print(f"isupper: {s10.isupper()}")
print(f"islower: {s10.islower()}")

print("\nAll tests completed!")
