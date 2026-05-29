# GoPy 并发架构设计文档

## 概述

本文档描述了 GoPy 的高性能并发架构，该架构借鉴了 Go 语言的 goroutine 和 channel 设计，实现了无 GIL（全局解释器锁）的并发执行模型。

## 设计目标

1. **无 GIL 锁** - 允许多个协程真正并行执行
2. **线程安全** - 提供线程安全的数据结构和通信机制
3. **轻量级** - 协程开销小，支持成千上万个并发协程
4. **简单易用** - 提供类似 Go 语言的简单编程模型
5. **高性能** - 优化的调度器，最小化上下文切换开销

## 核心组件

### 1. 协程 (Goroutine)

协程是 GoPy 中的轻量级执行单元，由调度器管理。

```go
type Goroutine struct {
    ID        uint64
    State     GoroutineState
    Stack     []objects.Object
    StackPtr  int
    Closure   *objects.Closure
    IP        int  // instruction pointer
    Result    objects.Object
    Error     error
    WaitChan  chan struct{}
    Ctx       *GoroutineContext
    Parent    *Goroutine
    Children  []*Goroutine
}
```

#### 协程状态

- `GoroutineIdle` - 空闲，等待调度
- `GoroutineRunning` - 正在运行
- `GoroutineWaiting` - 等待中（如等待 channel、sleep 等）
- `GoroutineSuspended` - 已挂起
- `GoroutineFinished` - 已完成

### 2. 调度器 (Scheduler)

调度器负责协程的调度和执行。采用工作窃取（work-stealing）算法的变种实现。

```go
type Scheduler struct {
    readyQueue  chan *Goroutine
    goroutines  map[uint64]*Goroutine
    workers     []*worker
    workerCount int
    ctx         context.Context
    cancel      context.CancelFunc
}
```

#### 工作原理

1. **多线程调度** - 使用多个工作线程（默认等于 CPU 核心数）
2. **任务队列** - 全局就绪队列保存待执行的协程
3. **负载均衡** - 工作线程从全局队列获取任务
4. **无锁设计** - 尽量减少锁的使用，提高性能

### 3. 通道 (Channel)

通道用于协程之间的通信，实现同步和数据传递。

```go
type Channel struct {
    ID        uint64
    Buffer    []objects.Object
    Capacity  int
    Closed    bool
    mu        sync.Mutex
    sendCond  *sync.Cond
    recvCond  *sync.Cond
}
```

#### 通道操作

- `Send(val)` - 发送值到通道
- `Receive()` - 从通道接收值
- `Close()` - 关闭通道
- `TrySend(val)` - 尝试发送（非阻塞）
- `TryReceive()` - 尝试接收（非阻塞）

### 4. 并发安全数据结构

#### ConcurrentList

并发安全的列表实现，使用读写锁保证线程安全。

#### ConcurrentDict

并发安全的字典实现，使用读写锁保证线程安全。

### 5. 同步原语

- `Mutex` - 互斥锁
- `WaitGroup` - 等待组
- `Once` - 一次性执行
- `AtomicInt32/AtomicInt64` - 原子整数
- `Pool` - 对象池

## 内存模型

GoPy 采用类似 Go 语言的内存模型：

1. **通道同步** - Channel 的发送和接收操作提供 happens-before 保证
2. **锁同步** - Mutex 的 Lock 和 Unlock 提供 happens-before 保证
3. **原子操作** - 原子操作提供 happens-before 保证

## 使用方式

### 在 GoPy 代码中使用

```python
import concurrency

# 创建通道
ch = concurrency.channel(0)  # 无缓冲通道

# 启动协程
def worker():
    concurrency.sleep(0.1)
    concurrency.send(ch, "Hello from goroutine!")

concurrency.go(worker)

# 接收数据
result = concurrency.recv(ch)
print(result)
```

### 使用 WaitGroup

```python
import concurrency

wg = concurrency.waitgroup()
wg.add(3)

def task(id):
    print(f"Task {id} started")
    concurrency.sleep(0.1)
    print(f"Task {id} finished")
    wg.done()

for i in range(3):
    concurrency.go(lambda: task(i))

wg.wait()
print("All tasks finished")
```

## 性能特性

1. **协程开销** - 每个协程初始栈大小小（可配置），内存占用低
2. **上下文切换** - 协程切换开销远小于线程切换
3. **扩展性** - 支持数万到数十万个并发协程
4. **多核利用** - 自动利用多核 CPU 进行并行计算

## 与 CPython 的差异

| 特性 | CPython | GoPy |
|------|---------|------|
| GIL | 有 | 无 |
| 并发模型 | 多线程 + 多进程 | 协程 + M:N 调度 |
| 并行能力 | 受限于 GIL | 真正的多核并行 |
| 通信机制 | 队列、锁 | Channel |
| 内存开销 | 线程开销大 | 协程开销小 |

## 后续工作

1. **完整集成** - 将调度器与 VM 完整集成
2. **异步 IO** - 支持异步 IO 操作（文件、网络等）
3. **性能优化** - 优化调度算法，减少锁竞争
4. **标准库** - 丰富的并发标准库
5. **调优工具** - 并发性能分析和调试工具

## 总结

GoPy 的并发架构借鉴了 Go 语言的成功经验，提供了无 GIL 的轻量级协程和基于 Channel 的通信机制，能够实现真正的并行执行，同时保持简单易用的编程模型。
