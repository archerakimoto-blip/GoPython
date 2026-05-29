# 乒乓球游戏示例 - 使用 Channel 进行协程通信

import concurrency

# 创建两个通道
ping = concurrency.channel(1)
pong = concurrency.channel(1)

def player(name, receive_ch, send_ch):
    count = 0
    while count < 5:
        # 等待接收
        msg = concurrency.recv(receive_ch)
        print(f"{name} received: {msg}")
        count += 1
        concurrency.sleep(0.1)
        # 发送回去
        concurrency.send(send_ch, f"{name} #{count}")
    
    concurrency.close(send_ch)
    print(f"{name} done!")

# 启动两个玩家协程
concurrency.go(lambda: player("Player 1", ping, pong))
concurrency.go(lambda: player("Player 2", pong, ping))

# 开始游戏
print("Starting game!")
concurrency.send(ping, "Serve!")

# 等待游戏结束
concurrency.sleep(1.5)
print("Game finished!")
