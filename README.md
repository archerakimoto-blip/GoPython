# Go Python 解释器 (GoPy)

一个用 Go 语言编写的高效 Python 解释器，包含完整的 JIT 编译架构、调试工具和性能分析器。

## 项目架构

这个解释器采用经典的多层架构：

1. **词法分析器 (Lexer)**: 将源代码转换为标记序列
2. **语法分析器 (Parser)**: 将标记序列构建为抽象语法树 (AST)
3. **编译器 (Compiler)**: 将 AST 编译为字节码
4. **虚拟机 (VM)**: 执行字节码
5. **JIT 编译器**: 即时编译热点代码，支持自动优化，支持 x86-64 和 ARM64 双平台
6. **垃圾回收器 (GC)**: 自动内存管理，标记-清除算法
7. **调试器 (Debugger)**: 交互式调试工具，支持断点、单步执行等
8. **性能分析器 (Profiler)**: 性能分析工具，统计函数调用和执行时间

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
│   ├── jit/            # JIT 编译器完整实现（支持 x86-64 和 ARM64）
│   ├── gc/             # 垃圾回收器（标记-清除算法）
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

### 使用 JIT 编译

```bash
# 启用 JIT 编译
./gopy --jit test.py

# 指定 ARM64 平台
./gopy --jit --jit-platform arm64 test.py

# 指定 x86-64 平台
./gopy --jit --jit-platform x86_64 test.py

# 启用激进优化
./gopy --jit --jit-aggressive test.py

# 启用性能分析
./gopy --jit --jit-profiling test.py

# 组合使用所有选项
./gopy --jit --jit-platform arm64 --jit-aggressive --jit-profiling test.py
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
- [x] **sys 系统模块**
- [x] **os 操作系统模块**
- [x] **json 数据处理模块**
- [x] **gc 垃圾回收模块**
- [x] **JIT 即时编译器（支持 x86-64 和 ARM64）**
- [x] **调试器工具**
- [x] **性能分析器**
- [x] **模块导入系统**（import 和 from...import）
- [x] **激进优化功能**（循环展开、内联优化、分支预测等）
- [x] **垃圾回收器**（标记-清除算法）

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

### sys 模块

GoPy 提供了 sys 模块，包含以下功能：

| 名称 | 描述 | 示例 |
|------|------|------|
| `sys.version` | 版本信息字符串 | `sys.version` |
| `sys.platform` | 系统平台标识 | `sys.platform` |
| `sys.version_info` | 版本信息元组 | `sys.version_info` |
| `sys.argv` | 命令行参数列表 | `sys.argv` |
| `sys.path` | 模块搜索路径 | `sys.path` |
| `sys.exit([code])` | 退出程序 | `sys.exit(0)` |
| `sys.getsizeof(obj)` | 返回对象的大小 | `sys.getsizeof(42)` |

### os 模块

GoPy 提供了 os 模块，用于与操作系统交互：

| 函数/属性 | 描述 | 示例 |
|-----------|------|------|
| `os.sep` | 路径分隔符 | `os.sep` |
| `os.getcwd()` | 获取当前工作目录 | `os.getcwd()` |
| `os.chdir(path)` | 切换工作目录 | `os.chdir('/tmp')` |
| `os.listdir(path)` | 列出目录内容 | `os.listdir('.')` |
| `os.mkdir(path[, mode])` | 创建目录 | `os.mkdir('newdir')` |
| `os.remove(path)` | 删除文件 | `os.remove('old.txt')` |
| `os.rename(src, dst)` | 重命名文件 | `os.rename('a.txt', 'b.txt')` |
| `os.getenv(key[, default])` | 获取环境变量 | `os.getenv('HOME')` |
| `os.environ` | 环境变量字典 | `os.environ` |

### json 模块

GoPy 提供了 json 模块，用于 JSON 数据处理：

| 函数 | 描述 | 示例 |
|------|------|------|
| `json.dumps(obj)` | 将对象序列化为 JSON 字符串 | `json.dumps({'name': 'Bob'})` |
| `json.loads(str)` | 将 JSON 字符串反序列化为对象 | `json.loads('{"name": "Bob"}')` |

### gc 垃圾回收模块

GoPy 提供了 gc 模块，用于垃圾回收控制：

| 函数/属性 | 描述 | 示例 |
|----------|------|------|
| `gc.enable()` | 启用垃圾回收 | `gc.enable()` |
| `gc.disable()` | 禁用垃圾回收 | `gc.disable()` |
| `gc.collect()` | 手动触发垃圾回收 | `gc.collect()` |
| `gc.get_stats()` | 获取垃圾回收统计信息 | `gc.get_stats()` |
| `gc.print_stats()` | 打印详细的垃圾回收统计 | `gc.print_stats()` |
| `gc.set_threshold(n)` | 设置垃圾回收阈值（字节） | `gc.set_threshold(2097152)` |
| `gc.set_verbose(flag)` | 设置垃圾回收详细输出模式 | `gc.set_verbose(True)` |

#### 垃圾回收器特性

GoPy 的垃圾回收器采用标记-清除算法，具有以下特性：

- **自动回收**: 当对象分配达到阈值时自动触发回收（默认 1MB）
- **手动控制**: 支持 `gc.collect()` 手动触发回收
- **统计信息**: 提供详细的内存分配、释放和对象存活情况统计
- **可配置**: 支持设置回收阈值和详细输出模式
- **根对象扫描**: 自动从栈、全局变量、帧等位置开始扫描根对象

### 使用 gc 模块

```python
import gc

# 启用/禁用垃圾回收
gc.enable()
gc.disable()
gc.enable()

# 获取统计信息
stats = gc.get_stats()
print("集合总数:", stats['collection_count'])
print("已分配字节数:", stats['allocated_bytes'])
print("已释放字节数:", stats['freed_bytes'])
print("存活对象数:", stats['live_objects'])

# 打印详细统计
gc.print_stats()

# 手动触发垃圾回收
gc.collect()

# 设置阈值（2MB）
gc.set_threshold(2097152)

# 设置详细模式
gc.set_verbose(True)
gc.set_verbose(False)
```

### 模块导入系统

GoPy 支持完整的模块导入功能：

```python
import math

print(math.pi)       # 3.14159...
print(math.e)        # 2.71828...
print(math.sqrt(16)) # 4.0
```

### 使用 sys 模块

```python
import sys

print("Version:", sys.version)
print("Platform:", sys.platform)
print("Version info:", sys.version_info)
print("Argv:", sys.argv)
print("Path:", sys.path)
print("Size of 42:", sys.getsizeof(42))
```

### 使用 os 模块

```python
import os

print("Path separator:", os.sep)
print("Current directory:", os.getcwd())
print("HOME env:", os.getenv("HOME", "/tmp"))
```

### 使用 json 模块

```python
import json

# Serialize to JSON
data = "Hello, GoPy!"
json_str = json.dumps(data)
print(json_str)  # "Hello, GoPy!"

# Parse from JSON
parsed = json.loads(json_str)
print(parsed)  # Hello, GoPy!
```

### JIT 编译器特性

GoPy 的 JIT 编译器提供以下功能：

- **跨平台支持**: 支持 x86-64 和 ARM64 双平台
- **自动平台检测**: 基于 runtime.GOARCH 自动选择目标平台
- **热点检测**: 自动识别频繁调用的函数
- **代码分析**: 分析指令数量、循环检测、复杂度计算
- **多级别优化**: 5级优化级别控制
- **激进优化**: 循环展开、内联优化、分支预测、死代码消除
- **缓存管理**: 智能缓存和驱逐策略
- **线程安全**: 支持并发访问
- **性能分析**: 收集函数调用次数、执行周期、热点路径分析

#### JIT 命令行选项

| 选项 | 描述 | 示例 |
|------|------|------|
| `--jit` | 启用 JIT 编译 | `./gopy --jit test.py` |
| `--jit-platform` | 指定目标平台 (x86_64/arm64) | `./gopy --jit --jit-platform arm64 test.py` |
| `--jit-aggressive` | 启用激进优化 | `./gopy --jit --jit-aggressive test.py` |
| `--jit-profiling` | 启用性能分析 | `./gopy --jit --jit-profiling test.py` |

#### 激进优化功能

GoPy 的激进优化包括：

1. **循环展开**: 将小循环展开4次减少分支开销
2. **内联优化**: 将小函数内联到调用点
3. **分支预测**: 优化分支指令顺序，提高 CPU 分支预测命中率
4. **激进死代码消除**: 移除未使用的常量和指令
5. **Peephole 优化**: 优化指令序列，合并冗余操作

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
- **热点路径分析**: 识别占用 10% 以上执行时间的函数

### 对象系统特性

GoPy 的对象系统采用了现代化的实现：

- **List**: 支持 append, extend, pop, insert, remove, reverse 等方法
- **Tuple**: 不可变序列类型
- **Dict**: 使用哈希映射，支持高效的键值对操作
- **Set**: 使用哈希集合，支持去重和高效的成员检查

## 示例

### 简单算术

```python
a = 10
b = 20
a + b
```

### 模块导入

```python
import math

print(math.pi)
print(math.e)
print(math.sqrt(16))
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
print(max(1, 2.5, 3))     # 3
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

### 垃圾回收测试
- `tests/python/test_gc.py` - gc 模块功能测试

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

# 使用 JIT 编译
./gopy --jit tests/python/test_all_comprehensive.py

# 使用 JIT 激进优化
./gopy --jit --jit-aggressive tests/python/test_all_comprehensive.py
```

## 未来计划

- [x] 支持更多异常类型（ValueError, TypeError 等）
- [x] 开发调试工具（调试器、性能分析器）
- [x] 添加字符串格式化（f-string 支持）
- [x] 支持模块导入系统（import 和 from...import）
- [x] 完善对象系统（Set, Dict, List 的现代化实现）
- [x] 实现更多内置模块（os, sys, json 等）
- [x] 增强 JIT 编译优化（真正编译到机器码，支持 x86-64 和 ARM64）
- [x] 添加激进优化功能（循环展开、内联优化等）
- [x] 实现垃圾回收（标记-清除算法）
- [ ] 实现与现有 Python 库的兼容性
- [ ] 添加更多 Python 标准库功能

## 贡献

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT 许可证
