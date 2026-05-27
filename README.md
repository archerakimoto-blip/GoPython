# Go Python 解释器 (GoPy)

一个用 Go 语言编写的高效 Python 解释器，包含完整的 JIT 编译架构、调试工具和性能分析器。

## 项目架构

这个解释器采用经典的多层架构：

1. **词法分析器 (Lexer)**: 将源代码转换为标记序列
2. **语法分析器 (Parser)**: 将标记序列构建为抽象语法树 (AST)
3. **编译器 (Compiler)**: 将 AST 编译为字节码
4. **虚拟机 (VM)**: 执行字节码
5. **JIT 编译器**: 即时编译热点代码，支持自动优化
6. **调试器 (Debugger)**: 交互式调试工具，支持断点、单步执行等
7. **性能分析器 (Profiler)**: 性能分析工具，统计函数调用和执行时间

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
│   ├── vm/             # 虚拟机（包含单元测试）
│   ├── objects/        # 核心对象系统
│   ├── jit/            # JIT 编译器完整实现
│   ├── debugger/       # 调试器
│   └── profiler/       # 性能分析器
├── tests/
│   ├── python/         # Python 测试脚本
│   ├── syntax_indent/  # 标准 Python 缩进语法测试
│   ├── syntax_braces/  # 大括号语法测试
│   └── features/       # 功能测试
├── gopy                # 编译后的可执行文件
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

### 使用调试器

```bash
./gopy -debug test.py
```

### 使用性能分析器

```bash
./gopy -profile test.py
```

## 特性

### 基础特性

- [x] 支持整数、浮点数、布尔值、字符串、数组、字典、集合
- [x] 基本算术运算 (+, -, *, /)
- [x] 比较运算 (==, !=, >, <)
- [x] 布尔运算 (and, or, not)
- [x] 变量绑定和作用域
- [x] 函数定义和调用
- [x] 条件语句 (if/else)
- [x] 循环语句 (for/while)
- [x] break/continue 语句
- [x] 列表推导式和集合推导式
- [x] 字典推导式
- [x] 切片操作
- [x] 索引和切片赋值
- [x] `pass` 语句支持

### 高级特性

- [x] **异常处理 (try/except/finally)**
- [x] **上下文管理器 (with 语句)**
- [x] **生成器 (yield 语句)**
- [x] **Lambda 表达式**
- [x] **闭包和自由变量**
- [x] **类和对象系统**
- [x] **类继承和多态**
- [x] **成员访问和方法调用**
- [x] **标准 Python 缩进语法**（与向后兼容的大括号语法共存）
- [x] **f-string 格式化字符串支持**
- [x] **丰富的内置函数库**
- [x] **math 数学模块（增强版）**
- [x] **JIT 即时编译器**
- [x] **调试器工具**
- [x] **性能分析器**

### 异常类型系统

GoPy 支持多种标准 Python 异常类型：

| 异常类型 | 描述 | 示例 |
|---------|------|------|
| `ValueError` | 值错误 | `int("abc")` |
| `TypeError` | 类型错误 | `abs("hello")` |
| `ZeroDivisionError` | 除零错误 | `1 / 0` |
| `IndexError` | 索引错误 | `lst[100]` |
| `KeyError` | 键错误 | `d["nonexistent"]` |
| `AttributeError` | 属性错误 | `obj.nonexistent` |
| `NameError` | 名称错误 | `undefined_var` |
| `AssertionError` | 断言错误 | `assert False` |
| `RuntimeError` | 运行时错误 | 各种运行时错误 |
| `NotImplementedError` | 未实现错误 | 未实现的功能 |

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
| `round(x, ndigits)` | 四舍五入 | `round(3.1415, 2)` 返回 3.14 |
| `range(start, stop, step)` | 数字序列 | `range(5)` 返回 [0,1,2,3,4] |
| `min(...)` | 最小值（支持整数和浮点数） | `min(1, 2.5, 3)` 返回 1 |
| `max(...)` | 最大值（支持整数和浮点数） | `max(1, 2.5, 3)` 返回 3 |
| `sum(iterable)` | 求和（支持整数和浮点数） | `sum([1, 2.5, 3])` 返回 6.5 |
| `zip(...)` | 将多个列表打包 | `zip([1,2], ["a","b"])` 返回 [[1,"a"], [2,"b"]] |
| `open(file, mode)` | 打开文件（上下文管理器） | `open("test.txt", "r")` |
| `next(generator)` | 获取生成器下一个值 | `next(gen)` |
| `append(list, item)` | 添加元素到列表 | `append(lst, 4)` |
| `setitem(dict, key, value)` | 设置字典项 | `setitem(d, "key", value)` |
| `setadd(set, item)` | 添加元素到集合 | `setadd(s, item)` |

### math 模块函数

GoPy 的 math 模块提供了丰富的数学函数：

| 函数 | 描述 | 示例 |
|------|------|------|
| `math.pi` | 圆周率 π | `math.pi` 返回 3.14159... |
| `math.e` | 自然常数 e | `math.e` 返回 2.71828... |
| `math.sin(x)` | 正弦函数 | `math.sin(math.pi/2)` 返回 1.0 |
| `math.cos(x)` | 余弦函数 | `math.cos(0)` 返回 1.0 |
| `math.tan(x)` | 正切函数 | `math.tan(math.pi/4)` 返回 1.0 |
| `math.asin(x)` | 反正弦函数 | `math.asin(1)` 返回 π/2 |
| `math.acos(x)` | 反余弦函数 | `math.acos(0)` 返回 π/2 |
| `math.atan(x)` | 反正切函数 | `math.atan(1)` 返回 π/4 |
| `math.sqrt(x)` | 平方根 | `math.sqrt(16)` 返回 4.0 |
| `math.pow(x, y)` | 幂运算 | `math.pow(2, 3)` 返回 8.0 |
| `math.hypot(x, y)` | 计算斜边 | `math.hypot(3, 4)` 返回 5.0 |
| `math.exp(x)` | 指数函数 | `math.exp(1)` 返回 e |
| `math.log(x)` | 自然对数 | `math.log(math.e)` 返回 1.0 |
| `math.log10(x)` | 常用对数 | `math.log10(100)` 返回 2.0 |
| `math.log2(x)` | 以2为底的对数 | `math.log2(8)` 返回 3.0 |
| `math.floor(x)` | 向下取整 | `math.floor(3.9)` 返回 3 |
| `math.ceil(x)` | 向上取整 | `math.ceil(3.1)` 返回 4 |
| `math.trunc(x)` | 截断取整 | `math.trunc(3.9)` 返回 3 |
| `math.degrees(x)` | 弧度转角度 | `math.degrees(math.pi)` 返回 180 |
| `math.radians(x)` | 角度转弧度 | `math.radians(180)` 返回 π |

### JIT 编译器特性

GoPy 的 JIT 编译器提供以下功能：

- **热点检测**: 自动识别频繁调用的函数
- **代码分析**: 分析指令数量、循环检测、复杂度计算
- **自动优化**: 对热点函数进行优化编译
- **缓存管理**: 智能缓存和驱逐策略
- **线程安全**: 支持并发访问

### 调试器特性

GoPy 提供了交互式调试器：

- **断点管理**: 设置、清除、查看断点
- **执行控制**:
  - `continue` (c) - 继续执行
  - `step` (s) - 单步进入
  - `next` (n) - 单步跳过（不进入函数）
  - `finish` (f) - 跳出当前函数
- **状态检查**:
  - `stack` (bt) - 打印堆栈跟踪
  - `locals` (l) - 打印局部变量
  - `globals` (g) - 打印全局变量
  - `break` (b) - 设置断点

### 性能分析器特性

GoPy 提供了性能分析工具：

- **函数调用统计**: 统计函数调用次数和总执行时间
- **指令统计**: 统计每条字节码指令的执行次数和耗时
- **详细报告**: 按执行时间排序的性能报告

## 示例

### 简单算术

```python
a = 10
b = 20
a + b
```

### f-string 格式化

```python
name = "Alice"
age = 30
pi = 3.14159

greeting = f"Hello, {name}!"
print(greeting)

info = f"Name: {name}, Age: {age}, Pi: {pi}"
print(info)

math_result = f"{age + 5}"
print(math_result)
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
print(float("3.14"))      # 3.14
print(str(123))            # "123"

# 数值操作（支持浮点数）
print(abs(-10.5))          # 10.5
print(min(1, 2.5, 3))     # 1
print(max(1, 2.5, 3))      # 3
print(sum([1, 2.5, 3.5])) # 7.0

# 生成序列
nums = range(1, 10, 2)    # [1, 3, 5, 7, 9]

# zip 函数
a = [1, 2, 3]
b = ["one", "two", "three"]
print(zip(a, b))  # [[1,"one"], [2,"two"], [3,"three"]]
```

### Lambda 表达式和闭包

```python
# 简单 lambda
add = lambda x, y: x + y
print(add(5, 3))  # 8

# 嵌套闭包
make_adder = lambda x: lambda y: x + y
add5 = make_adder(5)
print(add5(10))   # 15
```

### 类和对象（支持继承）

GoPy 支持两种语法风格：标准 Python 缩进语法和向后兼容的大括号语法。

#### 标准 Python 缩进语法（推荐）

```python
class Animal:
    def __init__(self, name):
        self.name = name
    
    def speak(self):
        print("Animal speaks")

class Dog(Animal):
    def speak(self):
        print(self.name + " says Woof!")

d = Dog("Buddy")
d.speak()  # Buddy says Woof!
print(d.name)  # Buddy
```

#### 向后兼容的大括号语法（可选）

```python
class Animal {
    def __init__(self, name) {
        self.name = name
    }
    
    def speak(self) {
        print("Animal speaks")
    }
}

class Dog(Animal) {
    def speak(self) {
        print(self.name + " says Woof!")
    }
}

d = Dog("Buddy")
d.speak()  # Buddy says Woof!
print(d.name)  # Buddy
```

### 数学模块

```python
import math

print(math.pi)       # 3.14159...
print(math.e)        # 2.71828...
print(math.sqrt(16)) # 4.0
print(math.sin(0))   # 0.0
print(math.cos(math.pi)) # -1.0
print(math.tan(math.pi/4)) # 1.0
print(math.log10(100)) # 2.0
print(math.hypot(3, 4)) # 5.0
```

## 测试用例

项目包含完整的测试套件：

### 异常处理测试
- `tests/python/test_try_multiline.py` - try/except 基本测试
- `tests/python/test_try_finally_multiline.py` - try/except/finally 测试

### 上下文管理器测试
- `tests/python/test_with_simple.py` - with 语句测试
- `tests/python/test_with_multiline.py` - 包含函数的 with 语句测试

### 生成器测试
- `tests/python/test_yield_simple.py` - 生成器基本测试
- `tests/python/test_yield_multiline.py` - 更复杂的生成器测试

### Lambda 和闭包测试
- `tests/python/test_lambda_simple.py` - Lambda 表达式基本测试
- `tests/python/test_lambda_call.py` - Lambda 调用测试

### 类和继承测试
- `tests/syntax_braces/test_class_inheritance.py` - 类继承测试
- `tests/syntax_braces/test_class_method2.py` - 类方法测试
- `tests/syntax_indent/test_simple_inheritance.py` - 标准语法继承测试

### f-string 测试
- `tests/features/test_fstring_braces.py` - f-string 基本测试
- `tests/syntax_indent/test_fstring_indent.py` - 标准语法 f-string 测试

### 内置函数测试
- `tests/python/test_builtins_new.py` - 新增内置函数测试
- `tests/python/test_all_comprehensive.py` - 综合功能测试

### Go 单元测试
- `pkg/vm/vm_test.go` - 虚拟机单元测试（包含算术、lambda、布尔运算测试）

运行测试：

```bash
# 运行 Python 测试脚本
./gopy tests/python/test_try_multiline.py
./gopy tests/python/test_all_comprehensive.py
./gopy tests/syntax_indent/test_fstring_indent.py

# 运行 Go 单元测试
go test ./pkg/vm -v
go test ./...

# 使用性能分析器
./gopy -profile tests/features/test_fstring_braces.py
```

## 未来计划

- [x] 支持更多异常类型（ValueError, TypeError 等）
- [x] 开发调试工具（调试器、性能分析器）
- [x] 添加字符串格式化（f-string 支持）
- [ ] 实现与现有 Python 库的兼容性
- [ ] 实现垃圾回收
- [ ] 实现更多内置模块（os, sys, json 等）
- [ ] 支持模块导入系统
- [ ] 添加更多 Python 标准库功能
- [ ] 增强 JIT 编译优化（真正编译到机器码）

## 贡献

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT 许可证
