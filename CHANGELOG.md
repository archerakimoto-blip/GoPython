# Changelog

所有重要的更改都会记录在这个文件中。

## [Unreleased]

### 重构

- **装饰器脱糖重构**：将 `@property`、`@classmethod`、`@staticmethod` 和 `__slots__` 的处理逻辑从编译器/虚拟机内核迁移至脱糖层，大幅简化内核执行路径
  - `@property` / `@name.setter` / `@name.deleter`：脱糖层自动生成混淆方法（`_desugar_prop_get_xxx` 等）和 `__getattr__` / `__setattr__` / `__delattr__` 拦截逻辑
  - `@classmethod`：脱糖层生成 `_desugar_cm_xxx` 方法和 `__getattr__`，通过 `__get_class__` + `__bind_method__` 绑定类对象
  - `@staticmethod`：脱糖层生成 `_desugar_sm_xxx` 方法和 `__getattr__`，直接返回函数无需绑定
  - `__slots__`：脱糖层生成 `__setattr__` 白名单检查，未在 slots 列表中的属性赋值将抛出 `AttributeError`
- **移除旧类型**：从对象系统中移除 `Property`、`ClassMethod`、`StaticMethod` 类型和 `Class.Slots` 字段，这些语义现在完全由脱糖层实现
- **新增 `BoundMethod` 对象类型**：用于表示绑定到实例/类的方法，在 VM 的 `executeCall` 中自动展开为 `fn(self, args...)`
- **新增内置辅助函数**：
  - `__get_class__(obj)` — 获取实例的类对象
  - `__bind_method__(fn, self)` — 创建 BoundMethod 对象
  - `__set_field__(instance, name, value)` — 直接设置实例字段
  - `__del_field__(instance, name)` — 直接删除实例字段

### 新增特性

- **并发架构**：实现了完整的类似 Go 语言 goroutine 的高性能并发架构，无 GIL 锁，支持真正的并行执行
  - 协程调度器：多线程调度器，支持成千上万并发协程
  - Channel 通信：支持有缓冲和无缓冲通道，提供协程间安全通信
  - **async/await 语法**：Python 风格的异步编程语法，支持 `async def` 函数声明和 `await` 表达式
  - 异步对象：Async 和 Future 对象，用于异步任务管理和结果处理
  - 并发安全数据结构：ConcurrentList、ConcurrentDict
  - 同步原语：Mutex、WaitGroup、Once、原子整数、对象池
  - 并发模块：concurrency 模块，提供完整的并发编程 API（go、channel、send、recv、sleep、mutex 等）
- **装饰器支持**（Decorators）：支持 `@decorator` 语法，包括简单装饰器、多个装饰器、带参数的装饰器
- **多重赋值/元组解包**：支持 `let a, b = 1, 2` 和 `x, y = [3,4]` 语法
- **链式比较**：支持 `a < b < c` 语法，自动转换为 `(a < b) and (b < c)`
- **关键字参数和 ****kwargs**：支持 `func(a=1, b=2)` 关键字参数调用和 `def func(**kwargs)` 可变关键字参数
- ****args 可变参数**：支持 `def func(*args)` 可变位置参数
- **增强赋值**（Augmented Assignment）：支持 `a += 1`、`a -= 1`、`a *= 2`、`a /= 2`、`a %= 2`、`a **= 2` 语法
- **字典推导式**：完整支持 `{key: value for key, value in iterable}` 和 `{key: value for key, value in iterable if condition}`
- **集合推导式**（SetComprehension）：支持 `{x for x in iterable}` 语法
- **生成器表达式**（GeneratorExpression）：支持 `(x for x in iterable)` 语法
- **多重上下文管理器**：支持 `with a, b:` 语法，自动脱糖成嵌套with语句
- **属性装饰器**：支持 @property、@name.setter、@name.deleter 装饰器语法，通过脱糖层自动生成 `__getattr__` / `__setattr__` / `__delattr__` 拦截逻辑
- **elif 语句**：完整支持条件分支 `if-elif-else` 结构
- **运算符增强**：支持 `%`、`//`、`**` 运算符，包括整数和浮点数
- **f-string 增强**：支持转义花括号、复杂表达式、多语句 f-string
- **词法分析器改进**：支持处理包含数字的标识符，支持 Python 风格的 `#` 注释，添加 `async` 和 `await` 关键字支持
- **AST 改进**：添加 `AwaitExpression` 节点类型，`FunctionLiteral` 添加 `IsAsync` 字段
- **Parser 改进**：添加 `parseAsyncFunction` 和 `parseAwaitExpression` 解析函数，修复 DEDENT token 处理，添加 ELIF 和 ELSE token 支持，修改 parseExpressionList 支持关键字参数解析
- **VM 改进**：修复可变参数 basePointer 计算错误，支持 OpGreaterThan 和 OpLessThan，添加 lastPopped 字段用于修复 Lambda 测试问题，添加 `OpMakeAsync` 和 `OpAwait` 操作码支持，添加 `Async` 和 `Future` 对象类型
- **Desugar 模块**：完善 For 循环脱糖为 While 循环，增强赋值脱糖，链式比较脱糖，装饰器脱糖，多重赋值脱糖，集合推导式脱糖，生成器表达式脱糖，多重上下文管理器脱糖，保留 `AwaitExpression` 和 `FunctionLiteral.IsAsync`
- **Compiler 改进**：添加 `OpMakeAsync` 和 `OpAwait` 操作码，`CompiledFunction` 添加 `IsAsync` 字段，修改 CallExpression 编译支持关键字参数打包成字典，修改 Let 语句和 Assign 语句处理 Names 数组（原先是单个 Name）
- **新增测试文件**：
  - tests/features/test_decorators.py
  - tests/features/test_varargs.py
  - tests/features/test_kwargs.py
  - tests/features/test_keyword_args.py
  - tests/features/test_async_simple.py
  - tests/features/test_concurrency.py
  - tests/concurrency_examples/ping_pong.py
  - tests/concurrency_examples/producer_consumer.py
- **新增文档**：
  - docs/concurrency_architecture.md：并发架构设计文档

### 修复的问题

- 修复了 *args 可变参数 basePointer 计算错误，导致访问 stack[-1] panic
- 修复了 return 语句在 for 循环中丢失的问题
- 修复了整数比较运算符缺失（OpGreaterThan 和 OpLessThan）
- 修复了变量赋值需要 `let` 关键字的问题（在测试文件中）
- 修复了 ELIF/ELSE 解析错误的问题
- 修复了词法分析器不能处理包含数字的标识符的问题
- 修复了词法分析器不能处理 Python 注释的问题
- 修复了 f-string 脱糖处理错误的问题
- 修复了 Lambda 函数测试失败的问题

## [0.2.0] - 2026-05-28

### 新增特性

- **JIT 编译器**：支持 x86-64 和 ARM64 双平台
- **调试器工具**：支持断点、单步执行、查看变量、堆栈跟踪
- **性能分析器**：统计函数调用次数和执行时间
- **垃圾回收器**：使用标记-清除算法，支持自动和手动回收
- **CPython 互操作**：通过 CGO 调用 CPython 库
- **新增模块**：random、string、time、datetime
- **增强的 math 模块**：支持三角函数、指数、对数、弧度角度转换
- **类继承支持**：支持 `class Child(Parent)` 继承语法
- **上下文管理器**：支持 `with` 语句
- **异常处理增强**：完整的 try/except/finally 处理，支持多种异常类型
- **丰富的内置函数**：zip、min、max、sum 等
- **字典和集合推导式**
- **生成器支持**：yield 语句和 next() 函数
- **模块导入系统**：import 和 from...import 语句
- **标准 Python 缩进语法**：与大括号语法共存

### 修复的问题

- 多个解析器问题
- VM 堆栈管理优化
- 对象系统改进

## [0.1.0] - 2026-01-01

### 新增特性

- 基本的算术运算
- 变量绑定
- 函数定义和调用
- if/else 条件语句
- for/while 循环
- 列表、字典、集合
- 基本的字符串和数字处理
- f-string 基本支持
- Lambda 表达式
- 类和对象系统基本支持
