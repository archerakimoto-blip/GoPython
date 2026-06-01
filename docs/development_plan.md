# GoPy 开发计划

## 设计理念

GoPy 采用**脱糖优先**（Desugar-First）的架构设计。核心原则是：**尽可能在脱糖层处理高级语法特性，将编译器和虚拟机保持为最小内核**。

这样做的好处：
1. **内核简洁** — 编译器和虚拟机只需处理基础语义，无需感知装饰器、slots 等高级特性
2. **可维护性** — 高级特性的变更集中在脱糖层，不污染内核代码
3. **可测试性** — 脱糖输出是纯 AST，可以独立验证转换正确性
4. **可扩展性** — 新增语法特性只需在脱糖层添加转换规则

---

## 代码审查报告（2026-06）

### 审查范围

对 Lexer、Parser、AST、Desugar、Compiler、VM、JIT 全链路进行审查，按三个优先级分类：

- **P0** — Python 还未实现的语法（影响语言完整性）
- **P1** — 可以通过 desugar 减轻执行负担的功能（影响架构简洁性）
- **P3** — 可以优化执行速度的其他优化（影响运行性能）

---

### P0：Python 还未实现的语法

#### 词法层缺失

| # | 特性 | 当前状态 | 影响 |
|---|------|----------|------|
| P0-1 | Raw strings `r"..."` | 无 RSTRING token，`r` 被解析为标识符 | 正则表达式、Windows 路径无法使用 |
| P0-2 | Byte strings `b"..."` | 无 BSTRING token | 字节操作不可用 |
| P0-3 | 单引号字符串 `'...'` | `readString()` 只处理 `"` | Python 最常用写法不可用 |
| P0-4 | 三引号字符串 `"""..."""` | 不支持多行字符串 | 文档字符串、多行文本不可用 |
| P0-5 | 位运算符 `& \| ^ ~ << >>` | 无对应 token 类型 | 位运算完全不可用 |
| P0-6 | 位运算增强赋值 `&= \|= ^= <<= >>=` | 无对应 token | 位运算增强赋值不可用 |
| P0-7 | `is` / `is not` 运算符 | 无 IS token | 身份比较不可用 |
| P0-8 | `in` / `not in` 成员运算符 | `in` 仅作为 for 关键字 | 成员测试不可用 |
| P0-9 | `assert` 语句 | 无 ASSERT 关键字 | 断言不可用 |
| P0-10 | 二进制/十六进制/八进制字面量 | `readNumber()` 不处理 `0b`/`0x`/`0o` | 非十进制整数不可用 |
| P0-11 | 数字下划线 `1_000_000` | 不处理 | 大数可读性差 |
| P0-12 | 复数字面量 `3+4j` | 无 COMPLEX token | 复数不可用 |
| P0-13 | Ellipsis `...` | 无 ELLIPSIS token | 切片和类型存根不可用 |
| P0-14 | `match` / `case` 关键字 | 无 MATCH/CASE token | 结构化模式匹配不可用 |

#### 解析层缺失

| # | 特性 | 当前状态 | 影响 |
|---|------|----------|------|
| P0-15 | `match/case` 模式匹配 | 无 `parseMatchStatement` | Python 3.10 核心特性不可用 |
| P0-16 | 多继承 | `ClassStatement.SuperClass` 为单个 `*Identifier` | `class C(A, B)` 不可用 |
| P0-17 | 仅关键字参数 `def f(*, key)` | `parseFunctionParameters` 不处理 `*` 分隔符 | 函数签名灵活性差 |
| P0-18 | 仅位置参数 `def f(a, /, b)` | 不处理 `/` 分隔符 | Python 3.8 特性不可用 |
| P0-19 | 默认参数值 | 解析器跳过但不存储 | `def f(x=10)` 不可用 |
| P0-20 | for 循环元组解包 `for x, y in pairs` | `ForStatement.Value` 为单个 `*Identifier` | 迭代解包不可用 |
| P0-21 | 多 for 子句推导式 | 推导式只支持单个 for | `[x+y for x in a for y in b]` 不可用 |
| P0-22 | 异常链 `raise E from e` | `RaiseStatement` 无 `from` 子句 | 异常追踪信息不完整 |
| P0-23 | 多异常类型 `except (A, B)` | `ExceptClause.Type` 为单个 Expression | 多类型捕获不可用 |
| P0-24 | `super()` 内置函数 | 未实现 | 继承中调用父类方法不可用 |
| P0-25 | 负数索引 `list[-1]` | `executeArrayIndex` 拒绝负索引 | Python 最常用操作不可用 |
| P0-26 | 字符串方法 `str.upper()` 等 | 未作为原生方法实现 | 字符串处理能力极弱 |
| P0-27 | `@x.setter` / `@x.deleter` 链式装饰器 | 部分处理，不完整 | property 完整协议不可用 |
| P0-28 | 异步推导式 `[x async for x in a]` | 无解析支持 | Python 3.6 特性不可用 |

#### 编译/虚拟机层缺失

| # | 特性 | 当前状态 | 影响 |
|---|------|----------|------|
| P0-29 | 位运算操作码 | 无 `OpBitAnd` 等 | 位运算无法编译 |
| P0-30 | 成员测试操作码 `in/not in` | 无 `OpIn` / `OpNotIn` | `x in list` 无法编译 |
| P0-31 | 身份测试操作码 `is/is not` | 无 `OpIs` / `OpIsNot` | `x is None` 无法编译 |
| P0-32 | Tuple 构造操作码 | 无 `OpTuple` | 元组字面量不可用 |
| P0-33 | 描述符协议 `__get__/__set__` | 未实现 | 属性描述符不可用 |
| P0-34 | 元类 `metaclass=` | 未实现 | 高级元编程不可用 |

---

### P1：可通过 desugar 减轻执行负担的功能

#### 已完成的脱糖

| 特性 | 脱糖方式 | 状态 |
|------|----------|------|
| `for` → `while` | 索引变量 + while 循环 | ✅ |
| `break`/`continue` | 由 for→while 脱糖隐含处理 | ✅ |
| 增强赋值 `+=` | `x = x + y` | ✅ |
| 链式比较 `a < b < c` | `(a < b) and (b < c)` | ✅ |
| `and`/`or` | `if/else` 表达式 | ✅ |
| 三元表达式 `a if b else c` | `if/else` 表达式 | ✅ |
| 多重赋值 `a, b = x, y` | 临时变量 + 索引访问 | ✅ |
| 多重上下文管理器 `with a, b` | 嵌套 with | ✅ |
| 装饰器 `@dec` | 临时变量 + 装饰器调用 | ✅ |
| `del x[y]` | `x.__delitem__(y)` | ✅ |
| `del x.y` | `x.__delattr__(y)` | ✅ |
| `yield from iter` | `for item in iter: yield item` | ✅ |
| `@property`/`@classmethod`/`@staticmethod` | `__getattr__`/`__setattr__` 生成 | ✅ |
| `__slots__` | `__setattr__` 白名单生成 | ✅ |

#### 新增脱糖机会（按优先级排序）

| # | 特性 | 脱糖方式 | 影响力 | 难度 | 理由 |
|---|------|----------|--------|------|------|
| P1-1 | `assert expr, msg` | `if not expr: raise AssertionError(msg)` | 高 | 低 | 避免新增 AST 节点和操作码 |
| P1-2 | `in` / `not in` 运算符 | `x.__contains__(y)` 或 `__in__(x, y)` | 高 | 低 | 避免新增操作码，复用方法调用 |
| P1-3 | `is` / `is not` 运算符 | `id(x) == id(y)` | 高 | 低 | 避免新增操作码 |
| P1-4 | 位运算符 | `x.__and__(y)` 等方法调用 | 中 | 低 | 避免新增 6 个操作码 |
| P1-5 | `match/case` | `if/elif` 链 + 变量绑定 | 高 | 高 | 避免新增整套模式匹配内核 |
| P1-6 | 多继承 MRO | 脱糖层展平方法解析，生成单继承 + 方法转发 | 高 | 高 | 避免内核实现 C3 线性化 |
| P1-7 | `super()` | 展开为直接父类方法调用 `ParentClass.method(self)` | 高 | 中 | 避免内核实现 super 机制 |
| P1-8 | `@dataclass` | 自动生成 `__init__`、`__repr__`、`__eq__` | 高 | 中 | 纯脱糖层实现，无需内核改动 |
| P1-9 | `@abstractmethod` | 在 `__getattr__` 中添加抽象检查 | 中 | 低 | 纯脱糖层实现 |
| P1-10 | `@functools.lru_cache` | 生成缓存包装方法 | 中 | 中 | 纯脱糖层实现 |
| P1-11 | 默认参数值 `def f(x=10)` | 脱糖为 `if x is None: x = 10` | 高 | 中 | 当前解析器跳过默认值 |
| P1-12 | for 循环元组解包 | 脱糖为索引访问 `x = t[0]; y = t[1]` | 中 | 低 | 扩展现有多重赋值脱糖 |
| P1-13 | 负数索引 `list[-1]` | 脱糖为 `list[len(list) + (-1)]` | 高 | 低 | 修复 VM 限制 |
| P1-14 | 异常链 `raise E from e` | 脱糖为 `exc = E; exc.__cause__ = e; raise exc` | 中 | 低 | 避免修改 RaiseStatement |
| P1-15 | 多异常类型 `except (A, B)` | 脱糖为多个 except 子句 | 中 | 低 | 避免修改 ExceptClause |
| P1-16 | f-string 格式化规格 | 脱糖为 `format(expr, spec)` 调用 | 中 | 中 | 当前 f-string 不支持 `:d`、`:.2f` |
| P1-17 | 多 for 子句推导式 | 脱糖为嵌套循环 | 中 | 中 | 避免修改推导式编译 |
| P1-18 | `NamedTuple` | 生成带字段名的类 | 中 | 中 | 纯脱糖层实现 |
| P1-19 | `Enum` | 生成枚举类 | 中 | 中 | 纯脱糖层实现 |
| P1-20 | 仅关键字参数 `def f(*, key)` | 脱糖为参数验证代码 | 中 | 中 | 避免修改编译器参数处理 |
| P1-21 | 仅位置参数 `def f(a, /, b)` | 脱糖为参数验证代码 | 低 | 中 | 避免修改编译器参数处理 |

---

### P3：可优化执行速度的其他优化

#### JIT 编译器（当前为空壳）

| # | 问题 | 当前状态 | 优化方案 |
|---|------|----------|----------|
| P3-1 | `JIT.Compile()` 返回 nil | 完全未实现 | 实现热点函数的字节码优化 |
| P3-2 | `optimizeFunction()` 仅做 Pop-Pop 合并 | 几乎无效 | 实现常量折叠、死代码消除、内联 |
| P3-3 | ARM/x86 代码生成器未连接 | 存在但未使用 | 连接到 JIT 管线 |
| P3-4 | 无真正的机器码执行 | MachineCode 字段仅复制字节码 | 实现 JIT 编译到本地代码 |

#### 虚拟机优化

| # | 优化 | 影响力 | 难度 | 说明 |
|---|------|--------|------|------|
| P3-5 | 特化操作码 `OpAddInt`/`OpAddFloat` | 高 | 中 | 消除每次运算的类型分派开销 |
| P3-6 | 内联缓存（Inline Cache） | 高 | 高 | 缓存属性查找结果，避免每次 `OpGetAttribute` 的 map 查找 |
| P3-7 | `range()` 惰性迭代器 | 高 | 低 | 当前 `range()` 物化整个列表，改为惰性生成 |
| P3-8 | `zip()` 惰性迭代器 | 高 | 低 | 同上 |
| P3-9 | 全局变量缓存 | 中 | 低 | 在 Frame 中缓存全局变量查找结果 |
| P3-10 | BoundMethod 缓存 | 中 | 低 | 避免每次属性访问创建新 BoundMethod |
| P3-11 | 字符串驻留（String Interning） | 中 | 低 | 去重字符串对象，减少内存和比较开销 |
| P3-12 | 对象池 | 中 | 中 | 复用 Integer/Boolean 对象，减少 GC 压力 |
| P3-13 | 常量折叠 | 中 | 中 | 编译时计算常量表达式如 `2 + 3` → `5` |
| P3-14 | 死代码消除 | 低 | 中 | 移除不可达代码 |
| P3-15 | 直接线程（Computed Goto） | 高 | 高 | 用跳转表替代 switch-case 分派 |
| P3-16 | 寄存器式 VM | 高 | 极高 | 从栈式 VM 改为寄存器式，减少指令数 |

#### 内存 / GC 优化

| # | 优化 | 影响力 | 难度 | 说明 |
|---|------|--------|------|------|
| P3-17 | Dict/Set 哈希表优化 | 中 | 中 | 使用 Swiss Table 替代 Go map |
| P3-18 | 逃逸分析 | 中 | 高 | 短生命周期对象避免堆分配 |
| P3-19 | 分代 GC | 中 | 高 | 区分年轻代/老年代减少全量扫描 |

---

### 影响力-难度矩阵

```
                    低难度                中难度                高难度
            ┌───────────────────┬───────────────────┬───────────────────┐
            │ P1-1 assert       │ P1-8 @dataclass   │ P1-5 match/case   │
            │ P1-2 in/not in    │ P1-11 默认参数值   │ P1-6 多继承MRO    │
  高影响力   │ P1-3 is/is not    │ P1-7 super()      │ P0-15 match/case  │
            │ P1-13 负数索引    │ P3-5 特化操作码    │ P0-16 多继承      │
            │ P3-7 range惰性    │ P3-13 常量折叠    │ P3-6 内联缓存     │
            │ P0-26 字符串方法  │ P0-19 默认参数值  │ P3-15 直接线程    │
            ├───────────────────┼───────────────────┼───────────────────┤
            │ P1-9 @abstractmethod│ P1-10 lru_cache  │ P3-16 寄存器VM    │
            │ P1-12 for元组解包 │ P1-16 f-string规格│ P3-4 JIT机器码    │
  中影响力   │ P1-14 异常链from  │ P1-17 多for推导式 │ P0-33 描述符      │
            │ P1-15 多except类型│ P1-18 NamedTuple  │ P0-34 元类        │
            │ P3-9 全局变量缓存 │ P3-12 对象池      │ P3-19 分代GC      │
            │ P3-10 BoundMethod │ P3-17 Dict优化    │                   │
            ├───────────────────┼───────────────────┼───────────────────┤
            │ P1-4 位运算脱糖   │ P1-20 仅关键字参数│ P1-21 仅位置参数  │
            │ P0-1 Raw strings  │ P0-17 仅关键字参数│ P0-18 仅位置参数  │
  低影响力   │ P0-10 0x/0b/0o   │ P1-19 Enum        │ P0-12 复数        │
            │ P0-11 数字下划线  │ P0-2 Byte strings │ P0-14 Ellipsis    │
            │ P3-11 字符串驻留  │ P3-14 死代码消除  │ P0-34 元类        │
            └───────────────────┴───────────────────┴───────────────────┘
```

#### 第一象限（高影响力 × 低难度）— 立即执行

| 编号 | 特性 | 预估工时 | 理由 |
|------|------|----------|------|
| P1-1 | `assert` 脱糖 | 0.5天 | 仅需脱糖层，无需修改内核 |
| P1-2 | `in/not in` 脱糖 | 1天 | 复用 `__contains__` 方法调用 |
| P1-3 | `is/is not` 脱糖 | 0.5天 | 复用 `id()` 比较 |
| P1-13 | 负数索引脱糖 | 0.5天 | 修复 VM 限制，高频操作 |
| P3-7 | `range()` 惰性迭代器 | 1天 | 消除大列表物化，性能提升显著 |
| P0-26 | 字符串方法 | 3天 | 使用频率最高的缺失特性 |

#### 第二象限（高影响力 × 中难度）— 优先执行

| 编号 | 特性 | 预估工时 | 理由 |
|------|------|----------|------|
| P1-8 | `@dataclass` 脱糖 | 3天 | 纯脱糖层，无需内核改动 |
| P1-11 | 默认参数值 | 3天 | 需修改解析器+脱糖层 |
| P1-7 | `super()` 脱糖 | 2天 | 需脱糖层识别继承关系 |
| P3-5 | 特化操作码 | 5天 | 消除类型分派开销，性能提升显著 |
| P3-13 | 常量折叠 | 3天 | 编译时优化，减少运行时计算 |

#### 第三象限（高影响力 × 高难度）— 规划执行

| 编号 | 特性 | 预估工时 | 理由 |
|------|------|----------|------|
| P1-5 | `match/case` 脱糖 | 10天 | 需完整模式匹配解析+脱糖 |
| P1-6 | 多继承 MRO 脱糖 | 7天 | 需实现 C3 线性化+方法转发 |
| P3-6 | 内联缓存 | 7天 | 需修改 VM 属性查找机制 |
| P3-15 | 直接线程 | 5天 | 需 Go 编译器支持或汇编 |

---

## 已完成

### v0.3 — 装饰器脱糖重构（当前版本）

将 `@property`、`@classmethod`、`@staticmethod` 和 `__slots__` 的处理从内核迁移至脱糖层。

#### 变更详情

| 模块 | 变更 |
|------|------|
| `pkg/desugar/desugar.go` | 新增 `desugarClassStatement` 函数，处理类级别的装饰器和 `__slots__` 脱糖 |
| `pkg/objects/object.go` | 新增 `BoundMethod` 类型和 `BOUNDMETHOD_OBJ` 类型常量；移除 `Property`、`ClassMethod`、`StaticMethod` 类型和 `Class.Slots` 字段 |
| `pkg/compiler/compiler.go` | 新增内置函数 `__get_class__`、`__bind_method__`、`__set_field__`、`__del_field__`；移除装饰器编译逻辑 |
| `pkg/vm/vm.go` | `executeCall` 新增 `BoundMethod` 展开（将 `bm.Fn` + `bm.Self` + args 推入栈）；移除 `Property`/`ClassMethod`/`StaticMethod` 运行时处理 |

#### 脱糖转换规则

**`@property`**

```python
class C:
    @property
    def x(self):
        return self._x
```

脱糖为：

```python
class C:
    def _desugar_prop_get_x(self):
        return self._x

    def __getattr__(self, name):
        if name == 'x':
            return self._desugar_prop_get_x()
```

**`@property` + `@x.setter`**

```python
class C:
    @property
    def x(self):
        return self._x

    @x.setter
    def x(self, value):
        self._x = value
```

脱糖为：

```python
class C:
    def _desugar_prop_get_x(self):
        return self._x

    def _desugar_prop_set_x(self, value):
        self._x = value

    def __getattr__(self, name):
        if name == 'x':
            return self._desugar_prop_get_x()

    def __setattr__(self, name, value):
        if name == 'x':
            self._desugar_prop_set_x(value)
            return
        __set_field__(self, name, value)
```

**`@classmethod`**

```python
class C:
    @classmethod
    def foo(cls, arg):
        return arg
```

脱糖为：

```python
class C:
    def _desugar_cm_foo(cls, arg):
        return arg

    def __getattr__(self, name):
        if name == 'foo':
            return __bind_method__(__get_class__(self)._desugar_cm_foo, __get_class__(self))
```

**`@staticmethod`**

```python
class C:
    @staticmethod
    def bar(arg):
        return arg
```

脱糖为：

```python
class C:
    def _desugar_sm_bar(arg):
        return arg

    def __getattr__(self, name):
        if name == 'bar':
            return __get_class__(self)._desugar_sm_bar
```

**`__slots__`**

```python
class C:
    __slots__ = ['x', 'y']
```

脱糖为：

```python
class C:
    def __setattr__(self, name, value):
        if name == 'x':
            __set_field__(self, name, value)
            return
        if name == 'y':
            __set_field__(self, name, value)
            return
        raise AttributeError("'C' object has no attribute '" + name + "'")
```

---

## 进行中

### 脱糖层增强

- [ ] `@property` 的 `@x.deleter` 端到端测试
- [ ] `@classmethod` / `@staticmethod` 端到端测试
- [ ] `__slots__` 端到端测试
- [ ] 装饰器与继承的交互测试
- [ ] 用户自定义 `__getattr__` / `__setattr__` 与脱糖生成代码的合并测试

### 词法分析器 / 解析器

- [ ] 类体中 `@property` / `@classmethod` / `@staticmethod` 装饰器解析的健壮性改进
- [ ] INDENT/DEDENT 边界情况处理

---

## 计划中

### v0.4 — 高影响力低难度特性（脱糖层快赢）✅ 已完成

目标：快速补齐最高频缺失特性，全部通过脱糖层实现，不增加内核复杂度。

- [x] P1-1：`assert` 语句脱糖 — `assert expr, msg` → `if not expr: raise AssertionError(msg)`
- [x] P1-2：`in`/`not in` 运算符脱糖 — `x in y` → `y.__contains__(x)`
- [x] P1-3：`is`/`is not` 运算符脱糖 — `x is y` → `id(x) == id(y)`
- [x] P1-13：负数索引 — VM 直接支持负数索引
- [x] P1-12：for 循环元组解包脱糖 — `for x, y in pairs` → 索引访问
- [x] P1-9：`@abstractmethod` 装饰器脱糖
- [ ] P1-14：异常链 `raise E from e` 脱糖
- [ ] P1-15：多异常类型 `except (A, B)` 脱糖
- [x] P0-1：Raw strings `r"..."` — Lexer 新增 RSTRING token
- [x] P0-3：单引号字符串 `'...'` — Lexer readString 支持
- [x] P0-25：负数索引 VM 修复
- [x] P0-26：字符串方法 — `str.upper()`、`str.split()` 等（含 List/Dict 原生方法）
- [x] P3-7：`range()` 惰性迭代器 — Range 对象替代列表物化
- [ ] P3-8：`zip()` 惰性迭代器
- [ ] P3-9：全局变量缓存
- [x] P3-10：BoundMethod 对象类型
- [x] P3-11：字符串驻留 — 编译器常量池去重

### Bug 修复（v0.4 中一并完成）

- [x] BUG-1：`BreakStatement`/`ContinueStatement` 在 desugar 中返回 nil → 改为返回自身
- [x] BUG-2：`readString()` 只处理 `"` 不处理 `'` → 支持单引号
- [x] BUG-3：`readString()` 不处理转义字符 → 支持 `\n`, `\t`, `\r`, `\\`, `\'`, `\"`, `\0`
- [x] BUG-5：`f-string` 解析中 `f` 前缀只匹配 `f"` → 支持 `f'...'`

### v0.5 — 高影响力中难度特性

目标：实现核心语言特性，提升语言完整性。

- [ ] P1-8：`@dataclass` 装饰器脱糖
- [ ] P1-11：默认参数值 — 修改解析器存储默认值 + 脱糖层处理
- [ ] P1-7：`super()` 脱糖
- [ ] P1-10：`@functools.lru_cache` 脱糖
- [ ] P1-16：f-string 格式化规格脱糖
- [ ] P1-17：多 for 子句推导式脱糖
- [ ] P0-4：三引号字符串
- [ ] P0-5：位运算符 token + P1-4 位运算脱糖
- [x] P0-7/P0-8：`is`/`in` 运算符 token（配合 P1-2/P1-3 脱糖）
- [x] P0-9：`assert` 关键字（配合 P1-1 脱糖）
- [ ] P0-17：仅关键字参数（配合 P1-20 脱糖）
- [ ] P0-19：默认参数值解析器支持
- [x] P0-20：for 循环元组解包（配合 P1-12 脱糖）
- [ ] P0-22：异常链 `from` 子句（配合 P1-14 脱糖）
- [ ] P0-23：多异常类型（配合 P1-15 脱糖）
- [ ] P3-5：特化操作码 `OpAddInt`/`OpAddFloat`
- [ ] P3-13：常量折叠
- [ ] P3-12：对象池

### v0.6 — 高影响力高难度特性

目标：实现高级语言特性，接近 Python 3.10+ 兼容。

- [ ] P1-5：`match/case` 模式匹配脱糖
- [ ] P1-6：多继承 MRO 脱糖
- [ ] P0-15：`match/case` 解析器支持（配合 P1-5）
- [ ] P0-16：多继承解析器支持（配合 P1-6）
- [ ] P0-2：Byte strings `b"..."`
- [ ] P0-10：二进制/十六进制/八进制字面量
- [ ] P0-27：`@x.setter`/`@x.deleter` 链式装饰器完整支持
- [ ] P0-28：异步推导式
- [ ] P0-29-P0-31：位运算/成员测试/身份测试操作码（配合脱糖）
- [ ] P0-32：Tuple 构造操作码
- [ ] P3-6：内联缓存
- [ ] P3-15：直接线程分派

### v0.7 — 性能优化

目标：JIT 编译器从空壳变为可用。

- [ ] P3-4：JIT 真正的机器码生成
- [ ] P3-2：JIT 优化器（常量折叠、死代码消除、内联）
- [ ] P3-3：ARM/x86 代码生成器连接
- [ ] P3-17：Dict/Set 哈希表优化
- [ ] P3-19：分代 GC

### v0.8 — 低影响力特性

目标：完善语言兼容性。

- [ ] P0-6：位运算增强赋值
- [ ] P0-11：数字下划线
- [ ] P0-12：复数字面量
- [ ] P0-13：Ellipsis
- [ ] P0-14：`match`/`case` 关键字 token
- [ ] P0-18：仅位置参数
- [ ] P0-33：描述符协议
- [ ] P0-34：元类
- [ ] P1-18：NamedTuple 脱糖
- [ ] P1-19：Enum 脱糖
- [ ] P1-20：仅关键字参数脱糖
- [ ] P1-21：仅位置参数脱糖
- [ ] P3-14：死代码消除
- [ ] P3-16：寄存器式 VM（长期目标）
- [ ] P3-18：逃逸分析

---

## 架构约束

1. **脱糖层不修改内核** — 脱糖层只能生成由现有内核支持的 AST 节点和内置函数调用
2. **内置函数最小化** — 新增内置函数需有充分理由，优先使用现有指令组合
3. **向后兼容** — 脱糖转换不应改变用户可见的语义行为
4. **可调试性** — 脱糖生成的混淆方法名（`_desugar_*`）应保持可追溯性
5. **脱糖优先原则** — 新增语法特性优先考虑脱糖实现，仅在性能关键路径上引入内核支持

---

## 关键代码问题清单

### 需要立即修复的 Bug

| # | 问题 | 文件 | 说明 |
|---|------|------|------|
| BUG-1 | `BreakStatement`/`ContinueStatement` 在 desugar 中返回 nil | [desugar.go:239-241](file:///workspace/pkg/desugar/desugar.go#L239-L241) | for→while 脱糖后 break/continue 丢失，循环控制流断裂 |
| BUG-2 | `readString()` 只处理 `"` 不处理 `'` | [lexer.go:388-397](file:///workspace/pkg/lexer/lexer.go#L388-L397) | Python 最常用的单引号字符串不可用 |
| BUG-3 | `readString()` 不处理转义字符 `\n`, `\t`, `\\` | [lexer.go:388-397](file:///workspace/pkg/lexer/lexer.go#L388-L397) | 含转义字符的字符串行为不正确 |
| BUG-4 | 负数索引 `list[-1]` 返回 None | [vm.go:1537-1545](file:///workspace/pkg/vm/vm.go#L1537-L1545) | `executeArrayIndex` 拒绝负索引 |
| BUG-5 | `f-string` 解析中 `f` 前缀只匹配 `f"` | [lexer.go:313-323](file:///workspace/pkg/lexer/lexer.go#L313-L323) | `f'...'` 不可用 |
| BUG-6 | `and`/`or` 脱糖为 IfExpression 但编译器报错 | [compiler.go:989](file:///workspace/pkg/compiler/compiler.go#L989) | 编译器拒绝 and/or 但脱糖层已转换，若脱糖遗漏则崩溃 |
| BUG-7 | `range()` 物化整个列表 | [compiler.go:615-673](file:///workspace/pkg/compiler/compiler.go#L615-L673) | `range(1000000)` 分配百万元素列表 |
| BUG-8 | `OpSlice` 弹出 3 个值但步长切片只压入 3 个 | [vm.go:357-364](file:///workspace/pkg/vm/vm.go#L357-L364) | 切片操作栈不平衡 |

### 架构改进建议

| # | 建议 | 说明 |
|---|------|------|
| ARCH-1 | `ClassStatement.SuperClass` 应改为 `[]*Identifier` | 支持多继承 |
| ARCH-2 | `ForStatement.Value` 应改为 `[]*Identifier` | 支持 for 循环元组解包 |
| ARCH-3 | `FunctionLiteral` 应增加 `Defaults` 字段 | 存储默认参数值 |
| ARCH-4 | `RaiseStatement` 应增加 `From` 字段 | 支持异常链 |
| ARCH-5 | `ListComprehension` 应支持多个 for/if 子句 | 支持嵌套推导式 |
| ARCH-6 | JIT `Compile()` 方法应返回实际编译结果 | 当前返回 nil |
| ARCH-7 | VM 应使用接口分派替代大 switch-case | 提高可维护性和性能 |
