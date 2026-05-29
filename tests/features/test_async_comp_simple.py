
# 异步列表推导式测试 - 使用简单迭代器
result = [x * 2 async for x in [1, 2, 3, 4, 5]]
print(result)

# 带过滤的异步列表推导式
filtered = [x async for x in [1, 2, 3, 4, 5, 6] if x % 2 == 0]
print(filtered)

print("Async comprehensions test done!")
