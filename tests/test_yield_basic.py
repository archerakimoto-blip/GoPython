# yield 语句测试

def counter(n):
    for i in range(n):
        yield i

gen = counter(3)
print("Created generator")

# 获取第一个值
try:
    val1 = next(gen)
    print(val1)
except Exception as e:
    print("Error: ")
    print(e)

# 获取第二个值
try:
    val2 = next(gen)
    print(val2)
except Exception as e:
    print("Error: ")
    print(e)
