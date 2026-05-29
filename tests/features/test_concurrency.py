
# concurrency 模块示例
import concurrency

print("Testing concurrency module")

# 测试简单的函数
print("Testing sleep function...")
print("Sleeping for 0.1 seconds...")
concurrency.sleep(0.1)
print("Done sleeping!")

print("\nTesting channel operations...")

# 创建一个通道
ch = concurrency.channel(2)

# 发送和接收（在我们简化的实现中，这只是示例）
# 实际使用时，应该结合 goroutine 使用
print("Channel ID:", ch)
