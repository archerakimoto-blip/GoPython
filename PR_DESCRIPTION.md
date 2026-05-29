# feat: 添加完整并发架构支持

## 概述

此次 PR 为 GoPy 解释器添加了完整的并发架构支持，借鉴 Go 语言的 goroutine 和 channel 设计，实现了无 GIL 锁的高性能并发系统，并引入了 Python 风格的 async/await 异步编程语法。

## 主要变更

### 新增功能

- **并发架构**：实现了完整的类似 Go 语言 goroutine 的高性能并发架构，无 GIL 锁，支持真正的并行执行
  - 协程调度器：多线程调度器，支持成千上万并发协程
  - Channel 通信：支持有缓冲和无缓冲通道，提供协程间安全通信
  - async/await 语法：Python 风格的异步编程语法，支持 `async def` 函数声明和 `await` 表达式
  - 异步对象：Async 和 Future 对象，用于异步任务管理和结果处理
  - 并发安全数据结构：ConcurrentList、ConcurrentDict
  - 同步原语：Mutex、WaitGroup、Once、原子整数、对象池
  - 并发模块：concurrency 模块，提供完整的并发编程 API

### 新增文件

| 文件 | 说明 |
|------|------|
| `pkg/concurrency/types.go` | 定义协程、Channel、同步原语等核心类型 |
| `pkg/concurrency/scheduler.go` | 协程调度器实现，支持多线程工作池 |
| `pkg/concurrency/channel.go` | Channel 通道实现，支持有缓冲和无缓冲通道 |
| `pkg/concurrency/module.go` | concurrency 模块 API 实现 |
| `pkg/concurrency/objects.go` | 并发安全数据结构（ConcurrentList、ConcurrentDict） |
| `docs/concurrency_architecture.md` | 并发架构设计文档 |
| `tests/concurrency_examples/ping_pong.py` | 乒乓球游戏示例 |
| `tests/concurrency_examples/producer_consumer.py` | 生产者-消费者示例 |
| `tests/features/test_async_simple.py` | async/await 简单测试 |
| `tests/features/test_concurrency.py` | concurrency 模块测试 |

### 修改文件

| 文件 | 变更内容 |
|------|----------|
| `pkg/ast/ast.go` | 添加 AwaitExpression 节点类型，FunctionLiteral 添加 IsAsync 字段 |
| `pkg/compiler/compiler.go` | 添加 OpMakeAsync/OpAwait 操作码，注册 concurrency 模块 |
| `pkg/compiler/symboltable.go` | CompiledFunction 添加 IsAsync 字段 |
| `pkg/lexer/lexer.go` | 添加 async/await 关键字 token |
| `pkg/parser/parser.go` | 添加 async def 和 await 表达式解析 |
| `pkg/objects/object.go` | 添加 Async/Future 对象类型 |
| `pkg/vm/vm.go` | 添加 OpMakeAsync/OpAwait 操作码实现 |
| `pkg/desugar/desugar.go` | 添加 await 表达式和 async 函数脱糖处理 |
| `CHANGELOG.md` | 更新变更日志 |
| `README.md` | 更新项目文档和功能列表 |

## 使用示例

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

## 性能特性

- 协程开销小，内存占用低
- 上下文切换开销远小于线程切换
- 支持数万到数十万个并发协程
- 自动利用多核 CPU 进行并行计算

## 相关 issue

无

## 检查清单

- [x] 代码遵循项目编码规范
- [x] 添加了相应的测试用例
- [x] 更新了相关文档
- [x] 通过所有现有测试
