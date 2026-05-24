# Go Python 解释器 (GoPy)

一个用 Go 语言编写的高效 Python 解释器，包含 JIT 编译基础架构。

## 项目架构

这个解释器采用经典的三层架构：

1. **词法分析器 (Lexer)**: 将源代码转换为标记序列
2. **语法分析器 (Parser)**: 将标记序列构建为抽象语法树 (AST)
3. **编译器 (Compiler)**: 将 AST 编译为字节码
4. **虚拟机 (VM)**: 执行字节码
5. **JIT 编译器 (正在开发中)**: 即时编译热点代码到本地机器码

## 目录结构

```
.
├── cmd/
│   └── gopy/           # 主程序入口点
├── pkg/
│   ├── ast/            # 抽象语法树定义
│   ├── lexer/          # 词法分析器
│   ├── parser/         # 语法分析器
│   ├── desugar/        # 语法脱糖（for/while循环转换）
│   ├── compiler/       # 字节码编译器
│   ├── vm/             # 虚拟机
│   ├── objects/        # 核心对象系统
│   └── jit/            # JIT 编译器基础架构
├── tests/              # 测试用例
├── debug/              # 调试工具
├── test.py             # 示例测试脚本
└── go.mod              # Go 模块定义
```

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

## 特性

### 基础特性

- [x] 支持整数、浮点数、布尔值、字符串、数组、字典、集合
- [x] 基本算术运算 (+, -, *, /)
- [x] 比较运算 (==, !=, >, <)
- [x] 变量绑定和作用域
- [x] 函数定义和调用
- [x] 条件语句 (if/else)
- [x] 循环语句 (for/while)
- [x] 列表推导式和集合推导式
- [x] 切片操作
- [x] 索引和切片赋值

### 高级特性

- [x] **异常处理 (try/except/finally)**
- [x] **上下文管理器 (with 语句)**
- [x] **生成器 (yield 语句)**
- [x] **丰富的内置函数库**

### 内置函数

GoPy 提供了丰富的内置函数：

| 函数 | 描述 | 示例 |
|------|------|------|
| `print(...)` | 打印输出 | `print("Hello")` |
| `len(obj)` | 返回长度 | `len([1,2,3])` 返回 3 |
| `type(obj)` | 返回类型名 | `type(42)` 返回 "INTEGER" |
| `str(obj)` | 转换为字符串 | `str(42)` 返回 "42" |
| `int(obj)` | 转换为整数 | `int(3.7)` 返回 3 |
| `float(obj)` | 转换为浮点数 | `float(42)` 返回 42.0 |
| `bool(obj)` | 转换为布尔值 | `bool(0)` 返回 False |
| `abs(x)` | 绝对值 | `abs(-5)` 返回 5 |
| `range(start, stop, step)` | 数字序列 | `range(5)` 返回 [0,1,2,3,4] |
| `min(...)` | 最小值 | `min(1, 2, 3)` 返回 1 |
| `max(...)` | 最大值 | `max(1, 2, 3)` 返回 3 |
| `sum(iterable)` | 求和 | `sum([1,2,3])` 返回 6 |
| `open(file, mode)` | 打开文件（上下文管理器） | `open("test.txt", "r")` |
| `next(generator)` | 获取生成器下一个值 | `next(gen)` |
| `append(list, item)` | 添加元素到列表 | `append(lst, 4)` |
| `setitem(dict, key, value)` | 设置字典项 | `setitem(d, "key", value)` |
| `setadd(set, item)` | 添加元素到集合 | `setadd(s, item)` |

## 示例

### 简单算术

```python
a = 10
b = 20
a + b
```

### 函数

```python
def add(x, y):
    return x + y
add(5, 3)
```

### 条件

```python
a = 10
if a > 5:
    print("greater")
else:
    print("less")
```

### 异常处理

```python
try:
    x = 1 / 0
except Exception as e:
    print("Caught exception:", e)
finally:
    print("Always runs")
```

### 上下文管理器 (with 语句)

```python
with open("test.txt", "r") as f:
    print("File opened:", f)
```

### 生成器 (yield 语句)

```python
def counter(n):
    for i in range(n):
        yield i

gen = counter(3)
print(next(gen))  # 0
print(next(gen))  # 1
```

### 列表推导式

```python
squares = [x * x for x in range(5)]
print(squares)  # [0, 1, 4, 9, 16]
```

### 使用内置函数

```python
# 类型检查
print(type(42))           # INTEGER
print(type("hello"))       # STRING

# 类型转换
print(int("42"))           # 42
print(float("3.14"))       # 3.14
print(str(123))            # "123"

# 数值操作
print(abs(-10))            # 10
print(min(1, 2, 3))       # 1
print(max(1, 2, 3))        # 3
print(sum([1, 2, 3, 4]))  # 10

# 生成序列
nums = range(1, 10, 2)     # [1, 3, 5, 7, 9]
```

## 测试用例

项目包含完整的测试套件：

### 异常处理测试
- `tests/test_try_multiline.py` - try/except 基本测试
- `tests/test_try_finally_multiline.py` - try/except/finally 测试

### 上下文管理器测试
- `tests/test_with_simple.py` - with 语句测试
- `tests/test_with_multiline.py` - 包含函数的 with 语句测试

### 生成器测试
- `tests/test_yield_simple.py` - 生成器基本测试
- `tests/test_yield_multiline.py` - 更复杂的生成器测试

### 内置函数测试
- `tests/test_builtins_new.py` - 新增内置函数测试
- `tests/test_all_comprehensive.py` - 综合功能测试

### 调试工具
- `debug/debug_try_except.go` - 字节码调试工具
- `debug/debug_finally.go` - finally 块字节码分析

运行测试：

```bash
./gopy tests/test_try_multiline.py
./gopy tests/test_all_comprehensive.py
```

## 未来计划

- [ ] 实现完整的 Python 语法支持
- [ ] 开发完整的 JIT 编译器
- [ ] 实现与现有 Python 库的兼容性
- [ ] 添加性能优化和缓存机制
- [ ] 实现垃圾回收
- [ ] 开发调试工具
- [ ] 实现类和对象系统
- [ ] 支持 lambda 表达式
- [ ] 添加字符串格式化
- [ ] 实现更多内置模块（os, sys, json 等）

## 贡献

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT 许可证
