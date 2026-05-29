# Go Python 解释器 (GoPy)

一个用 Go 语言编写的高效 Python 解释器，包含完整的 JIT 编译架构、调试工具、性能分析器和高性能并发系统。

详细更新历史请查看 [CHANGELOG.md](CHANGELOG.md)

## 快速开始

### 构建项目

```bash
go build -o gopy ./cmd/gopy
```

### 运行 REPL

```bash
./gopy
```

### 运行脚本文件

```bash
./gopy test.py
```

## 已支持的特性

### 基础特性
- 整数、浮点数、布尔值、字符串、数组、字典、集合
- 基本算术运算 (+, -, *, /, %, //, **)
- 比较运算 (==, !=, >, <, >=, <=, 链式比较)
- 布尔运算 (and, or, not)
- 变量绑定和作用域 (let)
- 多重赋值/元组解包
- 增强赋值 (+=, -=, *=, /=, %=, **=)
- 函数定义和调用
- 关键字参数和 *args/**kwargs
- 条件语句 (if/elif/else)
- 循环语句 (for/while)
- break/continue 语句
- 列表推导式、集合推导式、字典推导式
- 生成器表达式
- 切片操作和索引赋值
- pass 语句

### 并发特性
- **Goroutine 协程** - 轻量级执行单元，支持成千上万并发协程
- **Channel 通道** - 协程间通信机制，支持有缓冲和无缓冲通道
- **协程调度器** - 多线程调度器，无 GIL 锁，支持真正的并行执行
- **async/await 语法** - Python 风格的异步编程语法，支持 `async def` 和 `await` 表达式
- **异步对象** - Async 和 Future 对象，用于异步任务管理
- **并发安全数据结构** - ConcurrentList、ConcurrentDict
- **同步原语** - Mutex、WaitGroup、Once、原子整数、对象池
- **并发模块** - concurrency 模块，提供完整的并发编程 API（go、channel、send、recv、sleep、mutex 等）

### 高级特性
- 异常处理 (try/except/finally)
- 上下文管理器 (with 语句，支持多个上下文管理器)
- 生成器 (yield 语句)
- Lambda 表达式和闭包
- 类、对象、继承和多态
- 装饰器 (Decorators)
- 标准 Python 缩进语法
- f-string 格式化字符串
- 模块导入系统 (import/from...import)
- JIT 即时编译器（x86-64 和 ARM64）
- 调试器和性能分析器
- 垃圾回收器 (GC)
- CPython 互操作

### 标准库
- math, sys, os, json, gc, random, string, time, datetime, concurrency

## 未支持的特性

以下特性暂不支持，欢迎贡献！
- 类型注解 (Type Hints)
- 更多 Python 标准库

## 项目架构

```
.
├── cmd/
│   └── gopy/           # 主程序入口
├── pkg/
│   ├── ast/            # 抽象语法树
│   ├── lexer/          # 词法分析器
│   ├── parser/         # 语法分析器
│   ├── desugar/        # 语法脱糖
│   ├── compiler/       # 字节码编译器
│   ├── vm/             # 虚拟机
│   ├── objects/        # 核心对象系统
│   ├── concurrency/    # 并发系统（协程、调度器、Channel）
│   ├── jit/            # JIT 编译器
│   ├── gc/             # 垃圾回收器
│   ├── debugger/       # 调试器
│   └── profiler/       # 性能分析器
├── tests/              # 测试用例
│   ├── features/       # 功能测试
│   └── concurrency_examples/  # 并发示例
├── docs/               # 文档
└── go.mod
```
