# GoPy 开发计划

## 设计理念

GoPy 采用**脱糖优先**（Desugar-First）的架构设计。核心原则是：**尽可能在脱糖层处理高级语法特性，将编译器和虚拟机保持为最小内核**。

这样做的好处：
1. **内核简洁** — 编译器和虚拟机只需处理基础语义，无需感知装饰器、slots 等高级特性
2. **可维护性** — 高级特性的变更集中在脱糖层，不污染内核代码
3. **可测试性** — 脱糖输出是纯 AST，可以独立验证转换正确性
4. **可扩展性** — 新增语法特性只需在脱糖层添加转换规则

---

## 已完成

### v0.3 — 装饰器脱糖重构

将 `@property`、`@classmethod`、`@staticmethod` 和 `__slots__` 的处理从内核迁移至脱糖层。

**`@property`** → `__getattr__` 中添加 getter 调用
**`@property` + `@x.setter`** → `__getattr__` + `__setattr__` 中添加 getter/setter
**`@classmethod`** → `__getattr__` 中返回 `__bind_method__(cls._desugar_cm_foo, cls)`
**`@staticmethod`** → `__getattr__` 中返回 `cls._desugar_sm_bar`
**`__slots__`** → `__setattr__` 白名单检查

### v0.4 — 高影响力低难度特性 ✅ 已完成

目标：快速补齐最高频缺失特性，全部通过脱糖层实现，不增加内核复杂度。

- [x] P1-1：`assert` 语句脱糖 — `assert expr, msg` → `if not expr: raise AssertionError(msg)`
- [x] P1-2：`in`/`not in` 运算符脱糖 — `x in y` → `y.__contains__(x)`
- [x] P1-3：`is`/`is not` 运算符脱糖 — `x is y` → `id(x) == id(y)`
- [x] P1-13：负数索引 — VM 直接支持负数索引
- [x] P1-12：for 循环元组解包脱糖 — `for x, y in pairs` → 索引访问
- [x] P1-9：`@abstractmethod` 装饰器脱糖
- [x] P1-14：异常链 `raise E from e` 脱糖 — `exc.__cause__ = e; raise exc`
- [x] P1-15：多异常类型 `except (A, B)` 脱糖 — 展开为多个 except 子句
- [x] P0-1：Raw strings `r"..."` / `r'...'` — Lexer 新增 RSTRING token
- [x] P0-3：单引号字符串 `'...'` — Lexer readString 支持
- [x] P0-25：负数索引 VM 修复
- [x] P0-26：字符串方法 — `str.upper()`、`str.split()` 等（含 List/Dict/Integer 原生方法）
- [x] P3-7：`range()` 惰性迭代器 — Range 对象替代列表物化
- [x] P3-10：BoundMethod 对象类型
- [x] P3-11：字符串驻留 — 编译器常量池去重

### Bug 修复（v0.4 中一并完成）

- [x] BUG-1：`BreakStatement`/`ContinueStatement` 在 desugar 中返回 nil → 改为返回自身
- [x] BUG-2：`readString()` 只处理 `"` 不处理 `'` → 支持单引号
- [x] BUG-3：`readString()` 不处理转义字符 → 支持 `\n`, `\t`, `\r`, `\\`, `\'`, `\"`, `\0`
- [x] BUG-5：`f-string` 解析中 `f` 前缀只匹配 `f"` → 支持 `f'...'`

### v0.5 — 高影响力中难度特性 ✅ 已完成

目标：实现核心语言特性，提升语言完整性。

- [x] P1-8：`@dataclass` 装饰器脱糖 — 自动生成 `__init__` 和 `__repr__`
- [x] P1-11：默认参数值 — 修改解析器存储默认值 + 脱糖层 `if x == None: x = default`
- [x] P1-7：`super()` 脱糖 — `super()` → `__super__()`
- [x] P1-16：f-string 格式化规格脱糖 — `{expr:fmt}` → `format(expr, "fmt")`
- [x] P1-17：多 for 子句推导式脱糖 — 嵌套循环 + IIFE 包装
- [x] P0-4：三引号字符串 `"""..."""` / `'''...'''` — 含 `r"""..."""` 和 `f"""..."""`
- [x] P0-5：位运算符 token `& | ^ ~ << >>` + P1-4 位运算脱糖 → `__and__` 等方法调用
- [x] P0-6：位运算增强赋值 `&= |= ^= <<= >>=` token
- [x] P0-7/P0-8：`is`/`in` 运算符 token（配合 P1-2/P1-3 脱糖）
- [x] P0-9：`assert` 关键字（配合 P1-1 脱糖）
- [x] P0-19：默认参数值解析器支持
- [x] P0-20：for 循环元组解包（配合 P1-12 脱糖）
- [x] P0-22：异常链 `from` 子句（配合 P1-14 脱糖）
- [x] P0-23：多异常类型（配合 P1-15 脱糖）
- [x] P3-5：特化操作码 — VM 内联快速路径（int+int 的 OpAdd/OpSub/OpMul 直接计算）
- [x] P3-13：常量折叠 — 脱糖层编译期计算常量表达式（int/float/string/bool）

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

### 影响力-难度矩阵

```
                    低难度                中难度                高难度
            ┌───────────────────┬───────────────────┬───────────────────┐
            │ ✅ P1-1 assert    │ ✅ P1-8 @dataclass│ P1-5 match/case   │
            │ ✅ P1-2 in/not in │ ✅ P1-11 默认参数  │ P1-6 多继承MRO    │
  高影响力   │ ✅ P1-3 is/is not │ ✅ P1-7 super()   │ P0-15 match/case  │
            │ ✅ P1-13 负数索引 │ ✅ P3-5 特化操作码 │ P0-16 多继承      │
            │ ✅ P3-7 range惰性 │ ✅ P3-13 常量折叠 │ P3-6 内联缓存     │
            │ ✅ P0-26 字符串方法│ ✅ P0-19 默认参数 │ P3-15 直接线程    │
            ├───────────────────┼───────────────────┼───────────────────┤
            │ ✅ P1-9 @abstrmeth│ P1-10 lru_cache   │ P3-16 寄存器VM    │
            │ ✅ P1-12 for元组解包│ ✅ P1-16 f-string│ P0-33 描述符      │
  中影响力   │ ✅ P1-14 异常链   │ ✅ P1-17 多for推导│ P0-34 元类        │
            │ ✅ P1-15 多except │ P1-18 NamedTuple  │ P3-19 分代GC      │
            │ P3-9 全局变量缓存 │ P3-12 对象池      │                   │
            │ P3-10 BoundMethod │ P3-17 Dict优化    │                   │
            ├───────────────────┼───────────────────┼───────────────────┤
            │ ✅ P1-4 位运算脱糖 │ P1-20 仅关键字参数│ P1-21 仅位置参数  │
            │ ✅ P0-1 Raw strings│ P0-17 仅关键字参数│ P0-18 仅位置参数  │
  低影响力   │ P0-10 0x/0b/0o   │ P1-19 Enum        │ P0-12 复数        │
            │ P0-11 数字下划线  │ P0-2 Byte strings │ P0-14 Ellipsis    │
            │ ✅ P3-11 字符串驻留│ P3-14 死代码消除  │ P0-34 元类        │
            └───────────────────┴───────────────────┴───────────────────┘
```

### v0.6 — 高影响力高难度特性

目标：实现高级语言特性，接近 Python 3.10+ 兼容。

- [ ] P1-5：`match/case` 模式匹配脱糖
- [ ] P1-6：多继承 MRO 脱糖
- [ ] P1-10：`@functools.lru_cache` 脱糖
- [ ] P0-15：`match/case` 解析器支持（配合 P1-5）
- [ ] P0-16：多继承解析器支持（配合 P1-6）
- [ ] P0-2：Byte strings `b"..."`
- [ ] P0-10：二进制/十六进制/八进制字面量
- [ ] P0-17：仅关键字参数（配合 P1-20 脱糖）
- [ ] P0-27：`@x.setter`/`@x.deleter` 链式装饰器完整支持
- [ ] P0-28：异步推导式
- [ ] P0-32：Tuple 构造操作码
- [ ] P3-6：内联缓存
- [ ] P3-15：直接线程分派

### v0.7 — 性能优化

目标：JIT 编译器从空壳变为可用。

- [ ] P3-4：JIT 真正的机器码生成
- [ ] P3-2：JIT 优化器（常量折叠、死代码消除、内联）
- [ ] P3-3：ARM/x86 代码生成器连接
- [ ] P3-8：`zip()` 惰性迭代器
- [ ] P3-9：全局变量缓存
- [ ] P3-12：对象池
- [ ] P3-17：Dict/Set 哈希表优化
- [ ] P3-19：分代 GC

### v0.8 — 低影响力特性

目标：完善语言兼容性。

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

## 关键代码变更记录

### v0.5 变更详情

| 模块 | 变更 |
|------|------|
| `pkg/ast/ast.go` | 新增 `AssertStatement`、`FormattedExpression`、`ComprehensionFor`；`ForStatement` 新增 `Values` 字段；`FunctionLiteral` 新增 `Defaults` 字段；`RaiseStatement` 新增 `From` 字段；`ExceptClause` 新增 `Types` 字段；`ListComprehension` 新增 `Clauses`/`Filters` 字段 |
| `pkg/lexer/lexer.go` | 新增 `RSTRING`/`ASSERT`/`IS`/`AMPERSAND`/`PIPE`/`CARET`/`TILDE`/`LSHIFT`/`RSHIFT` 等 token；支持单引号字符串、raw string、三引号字符串、转义字符、位运算符 |
| `pkg/parser/parser.go` | 新增 `parseAssertStatement`/`parseIsExpression`/`parseInExpression`；支持位运算符、默认参数值、多异常类型、f-string 格式说明符、多 for 子句推导式 |
| `pkg/desugar/desugar.go` | 新增 `desugarAssertStatement`、`desugarDefaultParams`、`desugarSuperCall`、`desugarDataclass`、`foldConstants`、`desugarMultiClauseListComprehension`；`is`/`in`/位运算脱糖；`raise E from e` 脱糖；`FormattedExpression` → `format()` 调用 |
| `pkg/compiler/compiler.go` | 新增 `formatValue`/`formatInteger`/`formatFloat`/`formatString` 函数；`format` 内置函数支持 `format(value, spec)` 语义；`FormattedExpression` 编译支持；字符串驻留；`id`/`range`/`len` 内置函数改进 |
| `pkg/vm/vm.go` | 新增 `getIntegerAttribute`（`__and__`/`__or__`/`__xor__`/`__lshift__`/`__rshift__`/`__invert__`）；OpAdd/OpSub/OpMul 内联快速路径；`nextInstruction` 标签支持 goto 跳转 |
| `pkg/objects/object.go` | 新增 `NATIVE_METHOD_OBJ`/`RANGE_OBJ`/`BOUNDMETHOD_OBJ`；`NativeMethod`/`Range`/`BoundMethod` 类型 |
