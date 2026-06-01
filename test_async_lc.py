
# 测试异步列表推导式
result = [i * 2 async for i in [1, 2, 3]]
print("Result of async list comprehension:", result)
print("Type of result:", type(result))
