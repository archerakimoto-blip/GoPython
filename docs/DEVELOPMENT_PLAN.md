# Go Python 解释器 - 开发计划文档

## 项目概述

GoPython 是一个用 Go 语言编写的高效 Python 解释器，目标是为 Python 提供高性能的 JIT 编译实现。本文档详细规划了未来版本的特性开发优先级和实现路线图。

**当前版本**: 0.2.x
**目标版本**: 1.0.0
**最后更新**: 2026年

---

## 目录

1. [当前实现状态](#当前实现状态)
2. [特性优先级矩阵](#特性优先级矩阵)
3. [高优先级特性详细规划](#高优先级特性详细规划)
4. [中优先级特性详细规划](#中优先级特性详细规划)
5. [低优先级特性详细规划](#低优先级特性详细规划)
6. [技术架构改进](#技术架构改进)
7. [测试策略](#测试策略)
8. [文档计划](#文档计划)
9. [里程碑规划](#里程碑规划)

---

## 当前实现状态

### 已实现的核心特性

#### 1. 基础语法 (完成度: 93.75%)
✅ 整数、浮点数、布尔值、字符串
✅ 数组、字典、集合
✅ 基本算术运算 (+, -, *, /, %, //, **)
✅ 比较运算 (==, !=, >, <, >=, <=)
✅ 链式比较 (a < b < c)
✅ 布尔运算 (and, or, not)
✅ 变量绑定和作用域 (let)
✅ 多重赋值/元组解包
✅ 增强赋值 (+=, -=, *=, /=, %=, **=)
✅ 函数定义和调用
✅ 关键字参数和 *args/**kwargs
✅ 条件语句 (if/elif/else)
✅ 循环语句 (for/while)
✅ break/continue 语句
✅ 列表推导式、集合推导式、字典推导式
✅ 生成器表达式
✅ 切片操作（支持步长）
✅ pass 语句
✅ **Walrus 运算符 (:=)** - Python 3.8+
✅ **global/nonlocal 语句**
✅ **类型注解**
✅ **del 语句** ✨ 新增
✅ **yield from 语句** ✨ 新增
✅ **async for 语句** ✨ 新增
✅ **async with 语句** ✨ 新增

#### 2. 并发特性 (完成度: 60%)
✅ Goroutine 协程
✅ Channel 通道
✅ 协程调度器
✅ async/await 语法
✅ 异步对象 (Async, Future)
✅ 并发安全数据结构
✅ 同步原语 (Mutex, WaitGroup, Once)
✅ concurrency 模块
⚠️ asyncio 模块 - **部分实现**
⚠️ async comprehensions - **未实现**

#### 3. 高级特性 (完成度: 75%)
✅ 异常处理 (try/except/finally)
✅ 上下文管理器 (with)
✅ 生成器 (yield, yield from)
✅ Lambda 表达式和闭包
✅ 类、对象、继承
✅ 装饰器 (Decorators)
✅ f-string 格式化字符串
✅ 模块导入系统
✅ JIT 即时编译器
✅ 调试器和性能分析器
✅ 垃圾回收器 (GC)
✅ CPython 互操作

#### 4. 标准库 (完成度: 70%)
✅ math, sys, os, json, gc
✅ random, string, time, datetime
✅ concurrency
⚠️ re (正则表达式) - **未实现**
⚠️ io (IO 操作) - **部分实现**

### 未实现的 Python 特性

#### 🔴 严重缺失 (15 项)

1. **Pattern matching (match/case)** - 结构化模式匹配
2. **Exception chaining (from)** - 异常原因链
3. **Exception groups** - 异常组支持
4. **完整 f-string 支持** - 所有格式化选项
5. **字符串方法** - str.upper(), str.split() 等
6. **Raw strings (r"...")** - 原始字符串
7. **多继承** - 多父类继承
8. **MRO (Method Resolution Order)** - 方法解析顺序
9. **Descriptors** - 属性描述符
10. **Metaclasses** - 元类
11. **__slots__** - 实例属性限制
12. **Async comprehensions** - 异步推导式
13. **Keyword-only arguments** - 仅关键字参数
14. **Positional-only arguments (/)** - 仅位置参数
15. **正则表达式 (re 模块)** - 完整支持

#### 🟡 部分实现 (10 项)

16. **Abstract Base Classes** - 抽象基类
17. **Data Classes** - 数据类
18. **Named Tuples** - 命名元组
19. **Type hints generics** - 泛型类型提示
20. **Ellipsis (...)** - 省略号字面量
21. **Ordered Dict** - 有序字典
22. **Dictionary Views** - 字典视图
23. **Byte strings (b"...")** - 字节串
24. **Complex numbers** - 复数
25. **asyncio 模块** - 异步生态

---

## 特性优先级矩阵

### 影响力-难度矩阵

```
        低难度        中难度        高难度
高影响力  ┌──────────┬──────────┬──────────┐
         │字符串方法 │多继承    │match/case│
         │async     │异常链    │Descriptors│
         │comprehens│__slots__ │Metaclasses│
         └──────────┼──────────┼──────────┤
中影响力  │Raw strings│Keyword- │正则表达式│
         │Ellipsis  │only args│Exception  │
         │Byte str  │         │groups     │
         └──────────┼──────────┼──────────┤
低影响力  │Ordered   │Complex  │          │
         │Dict      │numbers  │          │
         │Dict views│         │          │
         └──────────┴──────────┴──────────┘
```

### 优先级评分标准

- **业务价值** (1-5分): 实际使用频率
- **实现难度** (1-5分): 代码复杂度
- **依赖关系** (1-5分): 其他特性的依赖程度
- **优先级分数**: 业务价值 × 3 + (6 - 实现难度) × 2 + (6 - 依赖关系)

---

## 高优先级特性详细规划

### 1. Pattern Matching (match/case) 🔴

**优先级**: ⭐⭐⭐⭐⭐ (最高)
**估计开发时间**: 4-6 周
**技术挑战**: 高

#### 1.1 功能描述

Python 3.10+ 引入的结构化模式匹配语法：

```python
match command.split():
    case ["quit"]:
        print("Goodbye!")
    case ["look"]:
        print("Looking...")
    case ["go", direction]:
        print(f"Going {direction}")
    case ["drop", *items]:
        print(f"Dropping {items}")
    case _:
        print("Unknown command")
```

#### 1.2 支持的模式类型

1. **通配符模式** - `case _:` 
2. **字面量模式** - `case 42:`, `case "text":`
3. **捕获模式** - `case name:`
4. **序列模式** - `case [x, y, *rest]:`
5. **映射模式** - `case {"key": value}:`
6. **类模式** - `case Point(x, y):`
7. **OR 模式** - `case 1 | 2 | 3:`
8. **AS 模式** - `case [x, y] as pair:`
9. **Guard 子句** - `case x if x > 0:`

#### 1.3 实现计划

**Phase 1: AST 节点定义 (1 周)**
- 创建 MatchStatement 节点
- 创建 CaseClause 节点
- 创建 Pattern 接口和实现
  - WildcardPattern
  - LiteralPattern
  - CapturePattern
  - SequencePattern
  - MappingPattern
  - ClassPattern
  - OrPattern
  - AsPattern

**Phase 2: Parser 实现 (2 周)**
- 实现 match 语句解析
- 实现 case 子句解析
- 实现所有模式类型的解析
- 处理 guard 条件

**Phase 3: Compiler 实现 (1 周)**
- 设计模式匹配字节码指令
- 实现模式编译策略
- 优化模式匹配顺序

**Phase 4: VM 执行 (2 周)**
- 实现模式匹配解释器
- 性能优化
- 错误处理

#### 1.4 技术细节

```go
// AST 节点定义
type MatchStatement struct {
    Token   token.Token
    Subject Expression
    Cases   []*CaseClause
}

type CaseClause struct {
    Token    token.Token
    Pattern  Pattern
    Guard    Expression  // 可选的 guard 条件
    Body     *BlockStatement
}

// Pattern 接口
type Pattern interface {
    Match(subject Object, env *Environment) (bool, Object, error)
    TypeCheck(subject Object) error
}
```

#### 1.5 测试用例

```python
# test_pattern_matching.py

# 测试通配符模式
match 42:
    case _:
        assert True

# 测试字面量模式
match "quit":
    case "quit":
        assert True
    case _:
        assert False

# 测试序列模式
match [1, 2, 3]:
    case [x, y, z]:
        assert x == 1 and y == 2 and z == 3

# 测试 OR 模式
match 2:
    case 1 | 2 | 3:
        assert True

# 测试带 guard
match 10:
    case x if x > 5:
        assert x == 10
```

---

### 2. String Methods 🟡

**优先级**: ⭐⭐⭐⭐⭐ (最高)
**估计开发时间**: 2-3 周
**技术挑战**: 中

#### 2.1 功能描述

完整的字符串方法支持：

**大小写转换**
- `str.upper()` - 转换为大写
- `str.lower()` - 转换为小写
- `str.capitalize()` - 首字母大写
- `str.title()` - 每个单词首字母大写
- `str.swapcase()` - 大小写交换

**查找和替换**
- `str.find(sub)` - 查找子串位置
- `str.replace(old, new)` - 替换
- `str.count(sub)` - 计数
- `str.startswith(prefix)` - 是否以某串开始
- `str.endswith(suffix)` - 是否以某串结束

**分割和连接**
- `str.split(sep)` - 分割
- `str.rsplit(sep)` - 从右分割
- `str.splitlines()` - 按行分割
- `str.join(iterable)` - 连接

**去除空白**
- `str.strip()` - 去除两端空白
- `str.lstrip()` - 去除左端
- `str.rstrip()` - 去除右端

**格式化和对齐**
- `str.center(width)` - 居中对齐
- `str.ljust(width)` - 左对齐
- `str.rjust(width)` - 右对齐
- `str.zfill(width)` - 零填充

**布尔判断**
- `str.isalpha()` - 是否全字母
- `str.isdigit()` - 是否全数字
- `str.isalnum()` - 是否字母数字
- `str.isspace()` - 是否空白字符
- `str.isupper()` / `str.islower()`

#### 2.2 实现计划

**Phase 1: 核心方法 (1 周)**
- 大小写转换方法
- 查找替换方法
- 分割连接方法

**Phase 2: 高级方法 (1 周)**
- 格式化对齐方法
- 布尔判断方法
- 编码转换方法

**Phase 3: 优化和测试 (1 周)**
- Unicode 支持
- 性能优化
- 完整测试覆盖

#### 2.3 代码示例

```go
// pkg/objects/string.go
func (s *String) Upper() *String {
    return &String{Value: strings.ToUpper(s.Value)}
}

func (s *String) Split(sep *String) *List {
    parts := strings.Split(s.Value, sep.Value)
    result := make([]Object, len(parts))
    for i, p := range parts {
        result[i] = &String{Value: p}
    }
    return &List{Elements: result}
}
```

---

### 3. 多继承和 MRO 🟡

**优先级**: ⭐⭐⭐⭐
**估计开发时间**: 3-4 周
**技术挑战**: 高

#### 3.1 功能描述

```python
class Base1:
    def method(self):
        return "Base1"

class Base2:
    def method(self):
        return "Base2"

class MyClass(Base1, Base2):
    pass

obj = MyClass()
print(obj.method())  # 应该遵循 C3 线性化算法
```

#### 3.2 实现计划

**Phase 1: 解析器支持 (1 周)**
- 修改 ClassStatement 支持多父类
- 验证逗号分隔的父类列表

**Phase 2: MRO 算法 (2 周)**
- 实现 C3 线性化算法
- 创建继承层次结构
- 方法解析顺序计算

**Phase 3: 运行时支持 (1 周)**
- 修改属性查找逻辑
- 实现 super() 内置函数
- 处理方法冲突

#### 3.3 C3 线性化算法

```go
func ComputeMRO(classes []*Class) ([]*Class, error) {
    // 实现 C3 线性化算法
    // Python 使用 C3 线性化算法确定方法解析顺序
}
```

---

### 4. Exception Chaining 🟡

**优先级**: ⭐⭐⭐⭐
**估计开发时间**: 2 周
**技术挑战**: 中

#### 4.1 功能描述

```python
try:
    raise ValueError("original")
except ValueError as e:
    raise TypeError("new") from e  # 异常链
```

#### 4.2 实现细节

```go
type Exception struct {
    Type      *String
    Value     *String
    Traceback *Traceback
    Context   *Exception  // 来自的异常
    __cause__ *Exception   // 显式指定的异常
}
```

---

### 5. Descriptors 🟡

**优先级**: ⭐⭐⭐⭐
**估计开发时间**: 3 周
**技术挑战**: 高

#### 5.1 功能描述

属性描述符协议：

```python
class Property:
    def __init__(self, fget):
        self.fget = fget
    
    def __get__(self, obj, objtype=None):
        return self.fget(obj)
    
    def __set__(self, obj, value):
        raise AttributeError("Read-only")

class MyClass:
    @property
    def value(self):
        return self._value
    
    @value.setter
    def value(self, val):
        self._value = val
```

#### 5.2 实现计划

**Phase 1: 描述符协议 (2 周)**
- `__get__(self, obj, objtype)`
- `__set__(self, obj, value)`
- `__delete__(self, obj)`

**Phase 2: 内置描述符 (1 周)**
- property
- classmethod
- staticmethod

---

### 6. Async Comprehensions 🟡

**优先级**: ⭐⭐⭐⭐
**估计开发时间**: 2 周
**技术挑战**: 中

#### 6.1 功能描述

```python
# 异步列表推导式
result = [x async for x in agenator if x > 0]

# 异步生成器表达式
gen = (x async for x in agenator)

# 异步字典/集合推导式
dict_result = {k: v async for k, v in agen.items()}
```

---

### 7. __slots__ 🟡

**优先级**: ⭐⭐⭐
**估计开发时间**: 2 周
**技术挑战**: 中

#### 7.1 功能描述

```python
class Point:
    __slots__ = ['x', 'y']
    
    def __init__(self, x, y):
        self.x = x
        self.y = y
```

---

### 8. Raw Strings 🟢

**优先级**: ⭐⭐⭐
**估计开发时间**: 1 周
**技术挑战**: 低

#### 8.1 功能描述

```python
path = r"C:\Users\Name\Documents"
regex = r"\d+\.\d+"  # 不需要转义
```

---

## 中优先级特性详细规划

### 9. Keyword-only Arguments ⚠️

**优先级**: ⭐⭐⭐
**估计开发时间**: 2 周

```python
def func(a, b, *, key1, key2):
    pass

func(1, 2, key1=3, key2=4)  # 正确
func(1, 2, 3, 4)            # 错误
```

### 10. Positional-only Arguments ⚠️

**优先级**: ⭐⭐⭐
**估计开发时间**: 2 周

```python
def func(a, b, /, c, d):
    pass

func(1, 2, 3, 4)  # 正确
func(1, 2, c=3, d=4)  # 正确
```

### 11. Ellipsis (...) ⚠️

**优先级**: ⭐⭐⭐
**估计开发时间**: 1 周

```python
# 类型标注
def func(x: List[int]) -> Dict[str, int]: ...

# 切片
arr[..., 1]  # NumPy 风格
```

### 12. Abstract Base Classes ⚠️

**优先级**: ⭐⭐
**估计开发时间**: 3 周

```python
from abc import ABC, abstractmethod

class Shape(ABC):
    @abstractmethod
    def area(self):
        pass
```

### 13. Data Classes ⚠️

**优先级**: ⭐⭐
**估计开发时间**: 4 周

```python
from dataclasses import dataclass

@dataclass
class Point:
    x: float
    y: float
    color: str = "red"
```

### 14. Metaclasses ⚠️

**优先级**: ⭐⭐
**估计开发时间**: 4 周

```python
class Meta(type):
    def __new__(cls, name, bases, dct):
        return super().__new__(cls, name, bases, dct)

class MyClass(metaclass=Meta):
    pass
```

### 15. Asyncio 模块 ⚠️

**优先级**: ⭐⭐⭐
**估计开发时间**: 4-6 周

---

## 低优先级特性详细规划

### 16. Ordered Dict ⚠️

**优先级**: ⭐
**估计开发时间**: 1 周

Python 3.7+ 的普通 dict 已经保持插入顺序，但 OrderedDict 类仍需实现以保持兼容性。

### 17. Dictionary Views ⚠️

**优先级**: ⭐
**估计开发时间**: 2 周

```python
d = {'a': 1, 'b': 2}
keys = d.keys()   # dict_keys(['a', 'b'])
values = d.values()  # dict_values([1, 2])
items = d.items()  # dict_items([('a', 1), ('b', 2)])
```

### 18. Complex Numbers ⚠️

**优先级**: ⭐
**估计开发时间**: 2 周

```python
c = 3 + 4j
print(c.real)  # 3.0
print(c.imag)  # 4.0
```

### 19. Byte Strings ⚠️

**优先级**: ⭐
**估计开发时间**: 2 周

```python
b = b"hello"
print(b[0])  # 104
```

---

## 技术架构改进

### 1. 性能优化

#### 1.1 JIT 编译器改进
- **热点检测**: 实现基于采样的热点检测
- **内联优化**: 函数内联和特化
- **代码生成**: 更高效的 x86-64 和 ARM64 代码生成
- **缓存机制**: 编译结果缓存

#### 1.2 内存管理
- **对象池**: 减少 GC 压力
- **内存分配器**: 定制化内存分配策略
- **延迟回收**: 减少短生命周期对象的 GC 开销

#### 1.3 并发优化
- **工作窃取**: 改进协程调度器
- **锁优化**: 减少锁竞争
- **批量操作**: 批量系统调用

### 2. 错误处理改进

#### 2.1 更好的错误信息
```python
# 当前
Error: unsupported operand type(s) for +: 'int' and 'str'

# 改进后
TypeError: unsupported operand type(s) for +: 'int' and 'str'
  File "test.py", line 3, in <module>
    result = 5 + "hello"
            ~ ^ ~~~~~~~~~~
```

#### 2.2 堆栈跟踪改进
- 显示更多上下文信息
- 源代码片段
- 变量值检查

### 3. 调试能力增强

#### 3.1 REPL 改进
- 自动补全
- 多行编辑
- 历史记录
- 帮助系统

#### 3.2 调试器增强
- 条件断点
- 监视点
- 异步调试

---

## 测试策略

### 1. 测试金字塔

```
        ┌─────────────┐
        │   E2E Tests │    端到端测试
        │    (50+)    │    真实 Python 脚本
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │  API Tests   │    集成测试
        │   (200+)     │    包/模块级别
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │  Unit Tests   │    单元测试
        │   (500+)     │    函数/类级别
        └─────────────┘
```

### 2. 测试覆盖目标

| 模块 | 当前覆盖 | 目标覆盖 |
|------|----------|----------|
| Parser | 75% | 95% |
| AST | 80% | 95% |
| Compiler | 70% | 90% |
| VM | 65% | 85% |
| Objects | 60% | 85% |
| Desugar | 70% | 90% |
| Concurrency | 55% | 80% |

### 3. 兼容性测试

#### 3.1 Python 标准测试套件
- 使用 Python 官方测试套件
- 重点关注已实现的特性
- 跟踪 CPython 行为

#### 3.2 跨版本测试
```bash
# 测试 Python 3.8-3.12 特性
python3.8 -m pytest tests/
python3.9 -m pytest tests/
python3.10 -m pytest tests/
```

### 4. 性能基准测试

```python
# benchmarks/test_performance.py
import time

def benchmark_list_comprehension():
    start = time.time()
    result = [x ** 2 for x in range(10000)]
    return time.time() - start

def benchmark_function_calls():
    start = time.time()
    for i in range(10000):
        fibonacci(i)
    return time.time() - start
```

---

## 文档计划

### 1. 用户文档

#### 1.1 快速开始指南
- 安装和配置
- 基本用法
- 常见示例

#### 1.2 语言参考
- 完整的语法规范
- 内置函数文档
- 标准库文档

#### 1.3 教程
- Python 到 GoPython
- 并发编程指南
- 性能优化技巧

### 2. 开发者文档

#### 2.1 架构设计
- 整体架构图
- 组件交互
- 设计决策

#### 2.2 代码规范
- Go 代码风格
- 提交规范
- PR 流程

#### 2.3 贡献指南
- 开发环境设置
- 测试指南
- 提交流程

### 3. API 文档

#### 3.1 Go API
```go
// 包文档
package gopy // import "github.com/go-py/go-python"

import (
    "github.com/go-py/go-python/pkg/vm"
)

// New creates a new Python VM instance
func New() *vm.VM

// Run executes Python code
func (vm *VM) Run(code string) error
```

#### 3.2 命令行工具
```bash
gopy run script.py
gopy compile script.py -o output.gyc
gopy disasm script.py
```

---

## 里程碑规划

### 版本 0.3.0 - 增强的字符串和函数特性
**目标日期**: 2026-Q2

- ✅ 完整的字符串方法支持
- ✅ Raw strings (r"...")
- ✅ Keyword-only arguments
- ✅ Positional-only arguments
- ✅ Ellipsis support
- ✅ 字符串测试覆盖率 > 90%

### 版本 0.4.0 - 高级面向对象
**目标日期**: 2026-Q3

- ✅ 多继承支持
- ✅ C3 MRO 算法
- ✅ Descriptors
- ✅ __slots__
- ✅ property/classmethod/staticmethod
- ✅ Metaclasses (基础)

### 版本 0.5.0 - 模式匹配
**目标日期**: 2026-Q4

- ✅ match/case 语法
- ✅ 所有模式类型
- ✅ Guard 子句
- ✅ 优化模式匹配性能

### 版本 1.0.0 - Production Ready
**目标日期**: 2027-Q1

- ✅ 异常链和异常组
- ✅ Abstract Base Classes
- ✅ Data Classes
- ✅ Async comprehensions
- ✅ 完整的 asyncio 模块
- ✅ 性能基准测试达标
- ✅ 生产环境验证

---

## 资源估算

### 开发时间总览

| 特性类别 | 特性数量 | 总时间 | 占比 |
|----------|----------|--------|------|
| 高优先级 | 15 | 30 周 | 50% |
| 中优先级 | 10 | 25 周 | 42% |
| 低优先级 | 5 | 5 周 | 8% |
| **总计** | **30** | **60 周** | **100%** |

### 团队规模建议

| 阶段 | 推荐团队规模 | 角色分配 |
|------|--------------|----------|
| 核心开发 | 2-3 人 | 1 核心架构 + 1-2 功能开发 |
| 全面开发 | 4-6 人 | 1 架构 + 2 功能 + 1 测试 + 1-2 文档 |
| 维护阶段 | 2-3 人 | 1 维护 + 1-2 功能 |

### 基础设施需求

1. **CI/CD**
   - GitHub Actions
   - 自动化测试
   - 性能基准测试

2. **代码质量**
   - Code Review
   - 代码覆盖率工具
   - 静态分析

3. **文档**
   - GitBook/Sphinx
   - API 文档生成
   - 示例代码

---

## 风险评估

### 1. 技术风险

| 风险 | 影响 | 可能性 | 缓解策略 |
|------|------|--------|----------|
| Pattern matching 实现复杂度 | 高 | 中 | 分阶段实现，逐步测试 |
| 性能回归 | 高 | 低 | 完整的性能测试套件 |
| 多继承 MRO 算法错误 | 高 | 中 | 对标 CPython 测试用例 |
| JIT 编译器 bug | 高 | 中 | 强化集成测试 |

### 2. 项目风险

| 风险 | 影响 | 可能性 | 缓解策略 |
|------|------|--------|----------|
| 开发时间超期 | 中 | 高 | 敏捷迭代，定期评估 |
| 资源不足 | 高 | 中 | 优先级排序，聚焦核心 |
| 需求变更 | 低 | 中 | 灵活架构，易于扩展 |

---

## 成功指标

### 1. 功能指标

- [ ] 支持 95% 的核心 Python 语法
- [ ] 字符串方法覆盖率 > 90%
- [ ] 类特性覆盖率 > 85%
- [ ] 异常处理完整度 > 90%

### 2. 性能指标

- [ ] 执行速度达到 CPython 的 80%+
- [ ] 启动时间 < 100ms
- [ ] 内存占用 < CPython 的 150%
- [ ] 并发性能线性扩展

### 3. 质量指标

- [ ] 测试覆盖率 > 80%
- [ ] 关键路径测试覆盖率 > 95%
- [ ] 文档覆盖率 > 90%
- [ ] Bug 修复响应时间 < 48h

### 4. 社区指标

- [ ] GitHub Stars > 1000
- [ ] 活跃贡献者 > 20
- [ ] Stack Overflow 问题 > 100
- [ ] 生产用户 > 10

---

## 附录

### A. Python 语法检查清单

```python
# 基础语法
✅ 变量和类型
✅ 运算符
✅ 控制流
✅ 函数定义
✅ 类定义
✅ 异常处理
✅ 上下文管理器
✅ 导入系统

# 高级语法
⚠️ Pattern matching (待实现)
✅ 生成器
✅ 装饰器
✅ 上下文管理器
✅ 元编程特性 (部分实现)

# 标准库
✅ 内置函数
⚠️ 字符串方法 (部分实现)
✅ 列表/字典/集合
⚠️ 正则表达式 (未实现)
⚠️ IO 操作 (部分实现)
```

### B. 参考资源

- [Python 官方文档](https://docs.python.org/3/)
- [CPython 源码](https://github.com/python/cpython)
- [Python 语言参考](https://docs.python.org/3/reference/)
- [Go 语言文档](https://go.dev/doc/)

### C. 术语表

- **AST**: Abstract Syntax Tree，抽象语法树
- **Desugar**: 语法脱糖，将高级语法转换为低级语法
- **JIT**: Just-In-Time，运行时编译
- **MRO**: Method Resolution Order，方法解析顺序
- **VM**: Virtual Machine，虚拟机
- **GC**: Garbage Collection，垃圾回收

---

**文档版本**: 1.0
**维护者**: GoPython 开发团队
**最后更新**: 2026年
