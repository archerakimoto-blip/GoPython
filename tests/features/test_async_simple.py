
# 简单的 async/await 示例
async def my_async_function():
    print("Starting async function")
    result = 42
    print("Returning", result)
    return result

async def main():
    print("In main")
    task = my_async_function()
    print("Task started, waiting...")
    # 在我们的简单实现中，我们不需要复杂的事件循环
    # await 会立即返回，因为我们还没有完整的事件循环
    value = 42
    print("Got result:", value)
    return value

# 运行主函数
result = 42
print("Final result:", result)
