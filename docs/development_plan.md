# GoPy 开发计划

## 设计理念

GoPy 采用**脱糖优先**（Desugar-First）的架构设计。核心原则是：**尽可能在脱糖层处理高级语法特性，将编译器和虚拟机保持为最小内核**。

这样做的好处：
1. **内核简洁** — 编译器和虚拟机只需处理基础语义，无需感知装饰器、slots 等高级特性
2. **可维护性** — 高级特性的变更集中在脱糖层，不污染内核代码
3. **可测试性** — 脱糖输出是纯 AST，可以独立验证转换正确性
4. **可扩展性** — 新增语法特性只需在脱糖层添加转换规则

---

## 代码审查报告

> 基于 2026-06 对 lexer / parser / ast / desugar / compiler / vm / objects 全模块的深度审查

### P0 — Python 还未实现的语法

| # | 特性 | 现状 | 影响模块 |
|---|------|------|----------|
| 1 | **`nonlocal` 声明** | 词法分析器无 `NONLOCAL` token，解析器无处理，VM 无闭包写回 | lexer / parser / compiler / vm |
| 2 | **`yield from` 表达式** | 仅支持 `yield`，不支持委托生成器 | lexer / parser / ast / compiler / vm |
| 3 | **`async with` / `async for`** | 仅有 `async def` / `await`，无异步上下文和异步迭代 | parser / ast / compiler / vm |
| 4 | **`@=` 矩阵乘法赋值** | 词法分析器无 `AT_EQ` token | lexer / compiler |
| 5 | **位运算符** `& \| ^ ~ << >>` 及其赋值 `&= \|= ^= <<= >>=` | 词法分析器无对应 token，VM 无运算处理 | lexer / compiler / vm |
| 6 | **`<=` `>=` 比较运算符** | 词法分析器无 `LTE` `GTE` token，VM 仅用 `OpGreaterThan` / `OpLessThan` 反转 | lexer / compiler / vm |
| 7 | **三引号字符串 / 多行字符串** | `readString()` 遇到第一个 `"` 即停止 | lexer |
| 8 | **单引号字符串** | 词法分析器仅识别 `"`，不识别 `'` | lexer |
| 9 | **字节串 `b"..."` / `rb"..."`** | 无 `BSTRING` token | lexer / ast / objects |
| 10 | **复数字面量 `1+2j`** | 无 `complex` 类型 | lexer / ast / objects |
| 11 | **`is` / `is not` 身份运算符** | 无 token 和编译支持 | lexer / parser / compiler / vm |
| 12 | **`in` / `not in` 成员运算符** | `in` 仅用于 `for...in`，不作为比较运算符 | parser / compiler / vm |
| 13 | **切片步长 `a[1:10:2]`** | AST `SliceExpression` 无 `Step` 字段 | ast / compiler / vm |
| 14 | **元组字面量 `(1, 2)`** | 解析器不识别无 `[]` `{}` 的容器构造 | parser / compiler |
| 15 | **星号解包赋值 `a, *b, c = [1,2,3,4]`** | 解析器无星号解包支持 | parser / desugar |
| 16 | **字典/列表/集合合并 `**d1, **d2` / `*l1, *l2`** | 解析器不识别展开语法 | parser / compiler |
| 17 | **`match` 模式匹配 — class pattern / as pattern / OR pattern** | 仅支持 identifier / literal / wildcard / tuple / list 模式 | parser / ast / desugar |
| 18 | **`except ... as e` 绑定异常变量** | AST `TryStatement` 无异常变量字段 | ast / compiler / vm |
| 19 | **`raise` 无参数重新抛出** | `raise` 必须带表达式 | parser / vm |
| 20 | **`global` 声明实际生效** | 解析器识别 token 但编译器/VM 不处理 | compiler / vm |
| 21 | **函数默认参数值** | AST `FunctionLiteral` 无默认值字段 | ast / parser / compiler |
| 22 | **函数参数注解 / 返回值注解** | 无类型注解 AST 节点 | ast / parser |
| 23 | **`del` 删除属性 `del obj.attr`** | 编译器仅处理标识符和索引删除 | compiler |
| 24 | **列表 `sort()` 方法** | 对象系统未实现 | objects |
| 25 | **`range()` 返回惰性迭代器** | 当前 `range()` 返回列表 | compiler / objects |

### P1 — 可通过 desugar 减轻执行负担的功能

| # | 特性 | 当前位置 | 脱糖方案 | 预期收益 |
|---|------|----------|----------|----------|
| 1 | **列表推导式编译** | compiler.go 1839-2162 行内联编译循环 | 脱糖为 `let _result = []; for ...: _result.append(expr)` | 编译器减少 ~300 行，统一推导式处理 |
| 2 | **集合推导式编译** | compiler.go 2165-2175 行 | 脱糖为 `let _result = set(); for ...: _result.add(expr)` | 与列表推导式统一 |
| 3 | **字典推导式编译** | compiler.go 2178-2225 行 | 脱糖为 `let _result = {}; for ...: _result[key] = value` | 与列表推导式统一 |
| 4 | **`assert` 语句** | desugar 已部分实现，但 compiler 仍有独立处理 | 完全在 desugar 层处理 | 消除编译器重复逻辑 |
| 5 | **`match` 语句** | desugar 已有 `desugarMatchStatement`，但模式支持不全 | 扩展脱糖支持 class pattern / as pattern / OR pattern | 无需 VM 支持 |
| 6 | **`@dataclass` 装饰器** | 不存在 | 脱糖生成 `__init__`、`__repr__`、`__eq__` | 常见需求，零内核负担 |
| 7 | **`@abstractmethod` 装饰器** | 不存在 | 脱糖在 `__getattr__` 中插入实例化检查 | 零内核负担 |
| 8 | **`super()` 调用** | VM 无特殊处理 | 脱糖展开为 `ParentClass.method(self, ...)` | 消除 VM 中的 super 逻辑 |
| 9 | **多重继承 MRO** | VM 仅支持单继承 `SuperClass` | 脱糖生成 `__getattr__` 链按 MRO 顺序查找 | VM 无需感知 MRO |
| 10 | **`__str__` / `__repr__` / `__eq__` 等双下划线方法** | VM 硬编码 `print()` 调用 | 脱糖在类中注册特殊方法，VM 统一通过 `__getattr__` 分发 | VM 消除硬编码分支 |
| 11 | **`in` / `not in` 运算符** | 不存在 | 脱糖为 `__contains__` 方法调用 + 回退迭代 | 无需新 opcode |
| 12 | **`is` / `is not` 运算符** | 不存在 | 脱糖为 `id(a) == id(b)` 内置函数调用 | 无需新 opcode |
| 13 | **`@functools.lru_cache`** | 不存在 | 脱糖生成缓存字典包装方法 | 纯脱糖实现 |
| 14 | **`@functools.total_ordering`** | 不存在 | 脱糖根据 `__eq__` + 一个比较方法生成全部 | 纯脱糖实现 |
| 15 | **`try/except` 多异常类型** | 仅支持单类型 | 脱糖为嵌套 except 块 | 无需 VM 变更 |
| 16 | **`finally` 块** | VM 有复杂的状态机 | 脱糖为 try/except + 临时变量保存 | 简化 VM 异常处理 |

### P2 — 可优化执行速度的其他优化

| # | 优化点 | 现状 | 方案 | 预期收益 |
|---|--------|------|------|----------|
| 1 | **属性访问 `OpGetAttribute` 热路径** | VM 每次调用 `GetAttr` 遍历 map | 添加内联缓存（Inline Cache）：记录上次查找的类+偏移 | 2-5x 属性访问加速 |
| 2 | **BoundMethod 重复创建** | 每次访问 `@classmethod` 属性都新建 `BoundMethod` | 在 Instance 上缓存已创建的 BoundMethod | 减少分配 |
| 3 | **`__getattr__` if-else 链** | 脱糖生成线性 if-else | 编译器识别模式，生成跳转表或哈希查找 | O(n) → O(1) |
| 4 | **整数运算 boxing/unboxing** | 每次运算都分配 `objects.Integer` | 栈上缓存小整数（-128~127），延迟分配 | 减少 GC 压力 |
| 5 | **字符串拼接优化** | `+` 每次创建新字符串 | 编译器识别连续 `+`，生成 `"".join(parts)` | 减少中间分配 |
| 6 | **编译器常量折叠** | `1 + 2` 编译为 `OpConstant(3)` | 在编译期计算纯整数/浮点/字符串表达式 | 减少运行时运算 |
| 7 | **死代码消除** | `if False: ...` 仍编译 | 编译器跳过不可达代码 | 减小字节码体积 |
| 8 | **`for` 循环迭代器协议优化** | 每次迭代调用 `next()` + 捕获 StopIteration | 对 list/dict/range 使用快速路径直接索引访问 | 2-3x 循环加速 |
| 9 | **VM 主循环分支预测** | `switch-case` 无优先级排序 | 按频率排序 case（OpCall/OpGetLocal/OpSetLocal 在前） | CPU 分支预测友好 |
| 10 | **字典实现优化** | `map[string]Object` + 单独 `Keys` map | 使用有序字典（slice + map 双索引） | 减少 `Keys` 冗余 |
| 11 | **对象类型判断优化** | `ObjectType` 是字符串比较 | 改为整数枚举 `ObjectType = int` | 类型判断从字符串比较变为整数比较 |

---

## 影响力-开发难度矩阵

```
                        高影响力
                           │
          ┌────────────────┼────────────────┐
          │   🔥 快赢区     │   🎯 核心攻坚   │
          │  (高影响低难度)  │  (高影响高难度)  │
  高难度  │                │                │
          │  P1-1 推导式脱糖 │  P0-1 nonlocal  │
          │  P1-8 super脱糖 │  P0-2 yield from│
          │  P1-11 in脱糖   │  P0-3 async with│
          │  P1-12 is脱糖   │  P0-5 位运算符   │
          │  P1-6 @dataclass│  P0-15 星号解包  │
          │  P2-6 常量折叠   │  P0-17 完整match │
          │  P2-7 死代码消除 │  P1-16 finally脱糖│
          │  P2-11 类型枚举  │  P1-10 双下划线  │
          │                │                │
  ────────┼────────────────┼────────────────┼─── 低难度
          │                │                │
          │  📋 低优先补丁   │   💎 锦上添花    │
          │  (低影响高难度)  │  (低影响低难度)  │
          │                │                │
          │  P0-9 复数字面量 │  P0-4 @= 运算符 │
          │  P0-10 字节串   │  P0-7 三引号字符串│
          │  P2-4 整数boxing│  P0-8 单引号字符串│
          │  P2-1 内联缓存  │  P0-6 <= >= 运算符│
          │  P2-5 字符串拼接│  P0-13 切片步长  │
          │                │  P0-19 raise无参 │
          │                │  P0-20 global生效 │
          │                │  P2-9 VM分支排序  │
          │                │  P2-10 字典优化   │
          │                │                │
          └────────────────┼────────────────┘
                           │
                        低影响力
```

### 详细分类表

#### 🔥 快赢区（高影响 · 低难度）— 优先实施

| 编号 | 特性 | 优先级 | 影响力 | 难度 | 预估工时 |
|------|------|--------|--------|------|----------|
| P1-1 | 列表/集合/字典推导式脱糖 | P1 | ★★★★★ | ★★ | 2天 |
| P1-6 | `@dataclass` 装饰器脱糖 | P1 | ★★★★ | ★★ | 1天 |
| P1-8 | `super()` 脱糖展开 | P1 | ★★★★ | ★★ | 1天 |
| P1-11 | `in` / `not in` 运算符脱糖 | P1 | ★★★★ | ★ | 0.5天 |
| P1-12 | `is` / `is not` 运算符脱糖 | P1 | ★★★ | ★ | 0.5天 |
| P2-6 | 编译器常量折叠 | P2 | ★★★★ | ★ | 1天 |
| P2-7 | 死代码消除 | P2 | ★★★ | ★★ | 1天 |
| P2-11 | ObjectType 字符串→整数枚举 | P2 | ★★★ | ★ | 0.5天 |

#### 🎯 核心攻坚（高影响 · 高难度）— 规划实施

| 编号 | 特性 | 优先级 | 影响力 | 难度 | 预估工时 |
|------|------|--------|--------|------|----------|
| P0-1 | `nonlocal` 声明 | P0 | ★★★★★ | ★★★★ | 3天 |
| P0-2 | `yield from` 委托生成器 | P0 | ★★★★ | ★★★★ | 3天 |
| P0-3 | `async with` / `async for` | P0 | ★★★ | ★★★★ | 4天 |
| P0-5 | 位运算符全套 | P0 | ★★★★ | ★★★ | 2天 |
| P0-15 | 星号解包赋值 | P0 | ★★★★ | ★★★ | 2天 |
| P0-17 | 完整 match 模式匹配 | P0 | ★★★ | ★★★★ | 5天 |
| P1-10 | 双下划线方法统一分发 | P1 | ★★★★★ | ★★★★ | 5天 |
| P1-16 | `finally` 块脱糖 | P1 | ★★★ | ★★★★ | 3天 |

#### 💎 锦上添花（低影响 · 低难度）— 穿插实施

| 编号 | 特性 | 优先级 | 影响力 | 难度 | 预估工时 |
|------|------|--------|--------|------|----------|
| P0-4 | `@=` 矩阵乘法赋值 | P0 | ★ | ★ | 0.5天 |
| P0-6 | `<=` `>=` 比较运算符 | P0 | ★★★ | ★ | 0.5天 |
| P0-7 | 三引号多行字符串 | P0 | ★★★ | ★★ | 1天 |
| P0-8 | 单引号字符串 | P0 | ★★★★ | ★ | 0.5天 |
| P0-13 | 切片步长 `a[1:10:2]` | P0 | ★★★ | ★★ | 1天 |
| P0-19 | `raise` 无参重新抛出 | P0 | ★★ | ★ | 0.5天 |
| P0-20 | `global` 声明生效 | P0 | ★★★ | ★★ | 1天 |
| P2-9 | VM 主循环分支预测排序 | P2 | ★★ | ★ | 0.5天 |
| P2-10 | 字典双索引优化 | P2 | ★★ | ★★ | 1天 |

#### 📋 低优先补丁（低影响 · 高难度）— 长期规划

| 编号 | 特性 | 优先级 | 影响力 | 难度 | 预估工时 |
|------|------|--------|--------|------|----------|
| P0-9 | 复数字面量 `1+2j` | P0 | ★★ | ★★★★ | 3天 |
| P0-10 | 字节串 `b"..."` | P0 | ★★ | ★★★ | 2天 |
| P2-1 | 属性访问内联缓存 | P2 | ★★★★ | ★★★★ | 5天 |
| P2-4 | 整数 boxing/unboxing 优化 | P2 | ★★★ | ★★★★ | 5天 |
| P2-5 | 字符串拼接优化 | P2 | ★★ | ★★★ | 2天 |

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

## 里程碑路线图

### v0.4 — 快赢 + 基础语法补全

**目标**：消除最常见的 Python 兼容性缺口，脱糖推导式

- [ ] P0-8 单引号字符串支持（lexer）
- [ ] P0-6 `<=` `>=` 比较运算符（lexer + compiler + vm）
- [ ] P1-11 `in` / `not in` 运算符脱糖
- [ ] P1-12 `is` / `is not` 运算符脱糖
- [ ] P1-1 列表/集合/字典推导式脱糖（从 compiler 迁移到 desugar）
- [ ] P1-6 `@dataclass` 装饰器脱糖
- [ ] P1-8 `super()` 脱糖展开
- [ ] P2-6 编译器常量折叠
- [ ] P2-11 ObjectType 字符串→整数枚举
- [ ] P0-20 `global` 声明生效

### v0.5 — 核心语法补全

**目标**：支持位运算、完整切片、元组、异常增强

- [ ] P0-5 位运算符全套（`& | ^ ~ << >>` 及赋值）
- [ ] P0-7 三引号多行字符串
- [ ] P0-13 切片步长 `a[1:10:2]`
- [ ] P0-14 元组字面量
- [ ] P0-18 `except ... as e` 绑定异常变量
- [ ] P0-19 `raise` 无参重新抛出
- [ ] P0-15 星号解包赋值
- [ ] P0-21 函数默认参数值
- [ ] P0-25 `range()` 返回惰性迭代器

### v0.6 — 高级特性

**目标**：闭包写回、委托生成器、双下划线方法体系

- [ ] P0-1 `nonlocal` 声明
- [ ] P0-2 `yield from` 委托生成器
- [ ] P0-3 `async with` / `async for`
- [ ] P1-10 双下划线方法统一分发（`__str__` / `__repr__` / `__eq__` / `__len__` / `__contains__` 等）
- [ ] P1-9 多重继承 MRO 脱糖
- [ ] P1-16 `finally` 块脱糖

### v0.7 — 性能优化

**目标**：运行时性能大幅提升

- [ ] P2-1 属性访问内联缓存
- [ ] P2-2 BoundMethod 缓存
- [ ] P2-3 `__getattr__` 跳转表优化
- [ ] P2-4 整数 boxing/unboxing 优化
- [ ] P2-8 `for` 循环迭代器快速路径
- [ ] P2-9 VM 主循环分支预测排序

### v0.8 — 完整性

**目标**：完整 Python 3 兼容

- [ ] P0-9 复数字面量
- [ ] P0-10 字节串
- [ ] P0-17 完整 match 模式匹配
- [ ] P0-22 函数参数注解 / 返回值注解
- [ ] P0-23 `del` 删除属性
- [ ] P0-24 列表 `sort()` 方法
- [ ] P0-16 字典/列表合并展开语法

---

## 架构约束

1. **脱糖层不修改内核** — 脱糖层只能生成由现有内核支持的 AST 节点和内置函数调用
2. **内置函数最小化** — 新增内置函数需有充分理由，优先使用现有指令组合
3. **向后兼容** — 脱糖转换不应改变用户可见的语义行为
4. **可调试性** — 脱糖生成的混淆方法名（`_desugar_*`）应保持可追溯性
5. **矩阵优先级** — 快赢区项目优先实施，核心攻坚项目需提前设计
