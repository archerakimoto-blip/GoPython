
# 异步推导式测试

# 测试 async list comprehension
async def test_async_list_comp():
    result = [x async for x in [1, 2, 3]]
    print("async list comp:", result)
    return result

# 测试 async set comprehension
async def test_async_set_comp():
    result = {x async for x in [1, 2, 3, 2, 1]}
    print("async set comp:", result)
    return result

# 测试 async dict comprehension
async def test_async_dict_comp():
    result = {str(x): x async for x in [1, 2, 3]}
    print("async dict comp:", result)
    return result

# 测试带 filter 的 async list comprehension
async def test_async_list_comp_filter():
    result = [x async for x in [1, 2, 3, 4, 5] if x > 2]
    print("async list comp with filter:", result)
    return result

# 运行测试
test_async_list_comp()
test_async_set_comp()
test_async_dict_comp()
test_async_list_comp_filter()

print("\nAll async comprehension tests completed!")
