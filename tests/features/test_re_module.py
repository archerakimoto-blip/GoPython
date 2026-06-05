
# re 正则表达式模块测试
import re

# 测试 re.search
result = re.search(r'\d+', 'abc123def')
if result:
    print("re.search found:", result.group())
else:
    print("re.search: no match")

# 测试 re.match
result = re.match(r'hello', 'hello world')
if result:
    print("re.match found:", result.group())
else:
    print("re.match: no match")

# 测试 re.match 不匹配
result = re.match(r'world', 'hello world')
if result:
    print("re.match should not match")
else:
    print("re.match: correctly no match at start")

# 测试 re.fullmatch
result = re.fullmatch(r'\d+', '123')
if result:
    print("re.fullmatch found:", result.group())
else:
    print("re.fullmatch: no match")

# 测试 re.findall
results = re.findall(r'\d+', 'a1b22c333')
print("re.findall:", results)

# 测试 re.sub
result = re.sub(r'\d+', 'X', 'a1b22c333')
print("re.sub:", result)

# 测试 re.split
result = re.split(r'[,;]', 'a,b;c,d')
print("re.split:", result)

# 测试 re.escape
result = re.escape(r'\d+.txt')
print("re.escape:", result)

# 测试 re.compile
pattern = re.compile(r'\d+')
result = pattern.search('abc456def')
if result:
    print("re.compile.search:", result.group())
else:
    print("re.compile.search: no match")

# 测试 pattern.findall
results = pattern.findall('a1b22c333')
print("re.compile.findall:", results)

# 测试 match 对象属性
result = re.search(r'(\d+)-(\d+)', 'abc12-34def')
if result:
    print("match.group():", result.group())
    print("match.group(1):", result.group(1))
    print("match.group(2):", result.group(2))
    print("match.start():", result.start())
    print("match.end():", result.end())
    print("match.span():", result.span())
else:
    print("match: no match")

# 测试 re.IGNORECASE
result = re.search(r'hello', 'HELLO world')
if result:
    print("re.search IGNORECASE: found (unexpected)")
else:
    print("re.search without IGNORECASE: no match (expected)")

# 测试模块常量
print("re.IGNORECASE:", re.IGNORECASE)
print("re.MULTILINE:", re.MULTILINE)
print("re.DOTALL:", re.DOTALL)

print("\nAll re module tests completed!")
