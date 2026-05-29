# 生产者-消费者模式示例

import concurrency

# 数据通道
data_channel = concurrency.channel(10)
# 完成信号通道
done_channel = concurrency.channel(2)

def producer(id, count):
    print(f"Producer {id} starting")
    for i in range(count):
        item = f"Item {id}-{i}"
        concurrency.send(data_channel, item)
        print(f"Producer {id} produced: {item}")
        concurrency.sleep(0.05)
    
    print(f"Producer {id} finished")
    concurrency.send(done_channel, id)

def consumer(id):
    print(f"Consumer {id} starting")
    while True:
        try:
            item = concurrency.recv(data_channel)
            if item is None:
                break
            print(f"Consumer {id} consumed: {item}")
            concurrency.sleep(0.1)
        except:
            break
    
    print(f"Consumer {id} finished")
    concurrency.send(done_channel, f"consumer-{id}")

# 启动生产者
concurrency.go(lambda: producer(1, 5))
concurrency.go(lambda: producer(2, 5))

# 启动消费者
concurrency.go(lambda: consumer(1))
concurrency.go(lambda: consumer(2))

# 等待生产者完成
print("Waiting for producers...")
for _ in range(2):
    concurrency.recv(done_channel)

# 关闭数据通道，让消费者退出
concurrency.close(data_channel)

# 等待消费者完成
print("Waiting for consumers...")
for _ in range(2):
    concurrency.recv(done_channel)

print("All done!")
