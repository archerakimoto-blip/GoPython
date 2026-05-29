
# 简单异步列表推导式测试
items = [1,2,3,4,5]
result = [x * 2 async for x in items]
print("Async list comp result: ", result)
