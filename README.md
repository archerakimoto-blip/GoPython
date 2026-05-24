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

- [x] 支持整数、浮点数、布尔值、字符串、数组、字典
- [x] 基本算术运算 (+, -, *, /)
- [x] 比较运算 (==, !=, >, <)
- [x] 变量绑定和作用域
- [x] 函数定义和调用
- [x] 条件语句 (if/else)
- [x] 循环语句 (for/while)
- [x] **异常处理 (try/except/finally)**
- [x] **上下文管理器 (with 语句)**
- [x] **生成器 (yield 语句)**
- [x] 基础数据结构
- [ ] JIT 编译 (正在开发)
- [ ] 完整 Python 语法支持
- [ ] Python 标准库支持

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

## 测试用例

项目包含完整的测试套件：

- `tests/test_try_multiline.py` - try/except 基本测试
- `tests/test_try_finally_multiline.py` - try/except/finally 测试
- `tests/test_with_simple.py` - with 语句测试
- `tests/test_yield_simple.py` - 生成器基本测试
- `debug/debug_try_except.go` - 字节码调试工具

运行测试：

```bash
./gopy tests/test_try_multiline.py
```

## 未来计划

- 实现完整的 Python 语法支持
- 开发完整的 JIT 编译器
- 实现与现有 Python 库的兼容性
- 添加性能优化和缓存机制
- 实现垃圾回收
- 开发调试工具

## 贡献

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT 许可证
