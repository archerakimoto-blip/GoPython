# GoPy 开发计划

**当前版本**: 0.15.x
**目标版本**: 1.0.0
**最后更新**: 2026-06

---

## 目录

1. [设计理念](#设计理念)
2. [当前实现状态](#当前实现状态)
3. [影响力-难度矩阵](#影响力-难度矩阵)
4. [版本历史](#版本历史)
5. [进行中](#进行中)
6. [路线图](#路线图)
7. [架构约束](#架构约束)
8. [测试策略](#测试策略)
9. [风险评估](#风险评估)

---

## 设计理念

GoPy 采用**脱糖优先**（Desugar-First）的架构设计。核心原则是：**尽可能在脱糖层处理高级语法特性，将编译器和虚拟机保持为最小内核**。

这样做的好处：
1. **内核简洁** — 编译器和虚拟机只需处理基础语义，无需感知装饰器、slots 等高级特性
2. **可维护性** — 高级特性的变更集中在脱糖层，不污染内核代码
3. **可测试性** — 脱糖输出是纯 AST，可以独立验证转换正确性
4. **可扩展性** — 新增语法特性只需在脱糖层添加转换规则

---

## 当前实现状态

### 1. 基础语法 (完成度: 100%)

✅ 整数、浮点数、布尔值、字符串
✅ 数组、字典、集合
✅ 基本算术运算 (+, -, *, /, %, //, **)
✅ 比较运算 (==, !=, >, <, >=, <=)
✅ 链式比较 (a < b < c)
✅ 布尔运算 (and, or, not)
✅ 变量绑定和作用域
✅ 多重赋值/元组解包
✅ 增强赋值 (+=, -=, *=, /=, %=, **=)
✅ 函数定义和调用
✅ 关键字参数和 *args/**kwargs
✅ 仅关键字参数 (*分隔符)
✅ 条件语句 (if/elif/else)
✅ 循环语句 (for/while)
✅ break/continue 语句
✅ 列表推导式、集合推导式、字典推导式
✅ 生成器表达式
✅ 切片操作（支持步长）
✅ pass 语句
✅ Walrus 运算符 (:=) - Python 3.8+
✅ global/nonlocal 语句
✅ 类型注解
✅ del 语句
✅ yield from 语句
✅ async for 语句
✅ async with 语句
✅ assert 语句
✅ 位运算符 (& | ^ ~ << >>)
✅ Raw strings (r"...")
✅ 三引号字符串 ("""...""")
✅ Byte strings (b"...")
✅ 0x/0b/0o 字面量 + 数字下划线
✅ match/case 模式匹配 - Python 3.10+
⚠️ 仅位置参数 (/) - **未实现** → ✅ 仅位置参数 (/) - v0.13 实现

### 2. 高级特性 (完成度: 95%)

✅ 异常处理 (try/except/finally) + 跨帧异常捕获 + 异常链 (raise E from e)
✅ 上下文管理器 (with)
✅ 生成器 (yield, yield from)
✅ Lambda 表达式和闭包
✅ 类、对象、单继承/多继承
✅ C3 MRO 算法
✅ 装饰器 (Decorators)
✅ f-string 格式化字符串（含格式化规格）
✅ 模块导入系统
✅ 描述符协议 (__get__/__set__/__delete__)
✅ 内置描述符 (property/classmethod/staticmethod)
✅ __slots__ 实例属性限制
✅ @dataclass 装饰器
✅ @abstractmethod 装饰器
✅ @lru_cache 装饰器
✅ NamedTuple
✅ Enum
✅ JIT 即时编译器
✅ 调试器和性能分析器
✅ 垃圾回收器 (GC)
✅ CPython 互操作
⚠️ Metaclasses - **未实现** → ✅ Metaclasses - v0.13 实现
⚠️ Exception groups - **未实现** → ✅ Exception groups - v0.13 实现

### 3. 并发特性 (完成度: 75%)

✅ Goroutine 协程
✅ Channel 通道
✅ 协程调度器
✅ async/await 语法
✅ 异步对象 (Async, Future)
✅ 并发安全数据结构
✅ 同步原语 (Mutex, WaitGroup, Once)
✅ concurrency 模块
⚠️ asyncio 模块 - **部分实现**
⚠️ async comprehensions - **未实现** → ✅ async comprehensions - v0.15 实现

### 4. 运行时优化 (完成度: 90%)

✅ 内联缓存 (attrCache)
✅ 全局变量缓存 (globalCache/globalVersions)
✅ 特化操作码 (int+int 快速路径)
✅ 常量折叠
✅ 对象池 (GetCachedInteger/GetCachedString)
✅ Dict 优化 (KeyOrder + NewDictWithCapacity)
✅ 死代码消除 (EliminateDeadCode)
✅ 字符串驻留
✅ BoundMethod 对象
✅ Range/Zip 惰性迭代器
⚠️ 寄存器 VM - **未实现** → ✅ 寄存器 VM - v0.15 实现
⚠️ 分代 GC - **未实现** → ✅ 分代 GC - v0.15 实现
⚠️ 直接线程 - **未实现**

### 5. 标准库 (完成度: 85%)

✅ math, sys, os, json, gc
✅ random, string, time, datetime
✅ concurrency
⚠️ re (正则表达式) - **未实现** → ✅ re (正则表达式) - v0.15 实现
⚠️ io (IO 操作) - **部分实现**

### 未实现的 Python 特性

#### 🔴 严重缺失 (5 项)

1. **Metaclasses** - 元类 → ✅ v0.13 实现
2. **Exception groups** - 异常组支持 → ✅ v0.13 实现
3. **Positional-only arguments (/)** - 仅位置参数 → ✅ v0.13 实现
4. **正则表达式 (re 模块)** - 完整支持 → ✅ v0.15 实现
5. **Async comprehensions** - 异步推导式 → ✅ v0.15 实现

#### 🟡 部分实现 (5 项)

6. **Type hints generics** - 泛型类型提示
7. **Ellipsis (...)** - 省略号字面量 → ✅ v0.14 实现
8. **Dictionary Views** - 字典视图 → ✅ v0.14 实现
9. **Complex numbers** - 复数 → ✅ v0.14 实现
10. **asyncio 模块** - 异步生态

#### 🐛 已知问题 (3 项)

1. **try-only-finally 异常穿透** — 已修复：`OpEndTry` 改用 `raiseException()` 传播异常
2. **描述符保留 VM 原生实现** — 这是正确的架构选择（性能关键路径不应迁移到脱糖层）
3. **range 迭代器死循环** — 已修复：`desugarForToWhile` 改用 `AssignStatement` + 嵌套循环唯一索引变量名

#### 🐍 Python 语义偏差 (v0.15 Code Review 新发现)

1. ~~**整数除法语义**：`6/3` 应返回 `2.0`（float），当前返回 `2`（int）~~ → ✅ v0.15.1 修复
2. ~~**取模运算符号**：`-7 % 3` 应返回 `2`（与除数同号），当前返回 `-1`（与被除数同号）~~ → ✅ v0.15.1 修复
3. ~~**List 负索引**：`lst[-1]` 不支持，`executeArrayIndex` 缺少负索引转换~~ → ✅ v0.15.1 修复
4. ~~**越界索引**：`lst[100]` 应抛出 `IndexError`，当前返回 `None`~~ → ✅ v0.15.1 修复
5. ~~**Boolean 算术**：`True + 1` 应等于 `2`，当前不支持~~ → ✅ v0.15.1 修复
6. ~~**字符串乘法**：`"abc" * 3` 不支持~~ → ✅ v0.15.1 修复
7. ~~**列表拼接**：`[1] + [2]` 不支持~~ → ✅ v0.15.1 修复
8. ~~**OpAdd 隐式拼接**：`1 + "a"` 应抛出 `TypeError`，当前返回 `"1a"`~~ → ✅ v0.15.1 修复
9. ~~**不可哈希类型作键**：`{[]: 1}` 应抛出 `TypeError`，当前使用指针地址~~ → ✅ v0.15.1 修复
10. ~~**整数负数幂**：`2 ** -1` 应返回 `0.5`，当前报错~~ → ✅ v0.15.1 修复

---

## 影响力-难度矩阵

```
                    低难度                中难度                高难度
            ┌───────────────────┬───────────────────┬───────────────────┐
            │ ✅ P1-1 assert    │ ✅ P1-8 @dataclass│ ✅ P1-5 match/case │
            │ ✅ P1-2 in/not in │ ✅ P1-11 默认参数  │ ✅ P1-6 多继承MRO  │
  高影响力   │ ✅ P1-3 is/is not │ ✅ P1-7 super()   │ ✅ P0-15 match/case│
            │ ✅ P1-13 负数索引 │ ✅ P3-5 特化操作码 │ ✅ P0-16 多继承    │
            │ ✅ P3-7 range惰性 │ ✅ P3-13 常量折叠 │ ✅ P3-6 内联缓存   │
            │ ✅ P0-26 字符串方法│ ✅ P0-19 默认参数 │ ❌ P3-15 直接线程  │
            ├───────────────────┼───────────────────┼───────────────────┤
            │ ✅ P1-9 @abstrmeth│ ✅ P1-10 lru_cache│ ❌ P3-16 寄存器VM  │
│ ✅ P1-12 for元组解包│ ✅ P1-16 f-string│ ✅ P0-33 描述符    │
  中影响力   │ ✅ P1-14 异常链   │ ✅ P1-17 多for推导│ ✅ P0-34 元类      │
│ ✅ P1-15 多except │ ✅ P1-18 NamedTuple│ ❌ P3-19 分代GC    │
            │ ✅ P3-9 全局变量缓存│ ✅ P3-12 对象池  │                   │
            │ ✅ P3-10 BoundMethod│ ✅ P3-17 Dict优化│                   │
            ├───────────────────┼───────────────────┼───────────────────┤
            │ ✅ P1-4 位运算脱糖 │ ✅ P1-20 仅关键字参数│ ✅ P1-21 仅位置参数│
│ ✅ P0-1 Raw strings│ ✅ P0-17 仅关键字参数│ ✅ P0-18 仅位置参数│
  低影响力   │ ✅ P0-10 0x/0b/0o │ ✅ P1-19 Enum        │ ✅ P0-12 复数    │
│ ✅ P0-11 数字下划线│ ✅ P0-2 Byte strings │ ✅ P0-14 Ellipsis│
            │ ✅ P3-11 字符串驻留│ ✅ P3-14 死代码消除  │ ✅ P0-34 元类     │
            └───────────────────┴───────────────────┴───────────────────┘
```

---

## 版本历史

### v0.3 — 装饰器脱糖重构 ✅

将 `@property`、`@classmethod`、`@staticmethod` 和 `__slots__` 的处理从内核迁移至脱糖层。

**`@property`** → `__getattr__` 中添加 getter 调用
**`@property` + `@x.setter`** → `__getattr__` + `__setattr__` 中添加 getter/setter
**`@classmethod`** → `__getattr__` 中返回 `__bind_method__(cls._desugar_cm_foo, cls)`
**`@staticmethod`** → `__getattr__` 中返回 `cls._desugar_sm_bar`
**`__slots__`** → `__setattr__` 白名单检查

### v0.4 — 高影响力低难度特性 ✅

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

**Bug 修复（v0.4）：**

- [x] BUG-1：`BreakStatement`/`ContinueStatement` 在 desugar 中返回 nil → 改为返回自身
- [x] BUG-2：`readString()` 只处理 `"` 不处理 `'` → 支持单引号
- [x] BUG-3：`readString()` 不处理转义字符 → 支持 `\n`, `\t`, `\r`, `\\`, `\'`, `\"`, `\0`
- [x] BUG-5：`f-string` 解析中 `f` 前缀只匹配 `f"` → 支持 `f'...'`

### v0.5 — 高影响力中难度特性 ✅

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

**v0.5 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/ast/ast.go` | 新增 `AssertStatement`、`FormattedExpression`、`ComprehensionFor`；`ForStatement` 新增 `Values` 字段；`FunctionLiteral` 新增 `Defaults` 字段；`RaiseStatement` 新增 `From` 字段；`ExceptClause` 新增 `Types` 字段；`ListComprehension` 新增 `Clauses`/`Filters` 字段 |
| `pkg/lexer/lexer.go` | 新增 `RSTRING`/`ASSERT`/`IS`/`AMPERSAND`/`PIPE`/`CARET`/`TILDE`/`LSHIFT`/`RSHIFT` 等 token；支持单引号字符串、raw string、三引号字符串、转义字符、位运算符 |
| `pkg/parser/parser.go` | 新增 `parseAssertStatement`/`parseIsExpression`/`parseInExpression`；支持位运算符、默认参数值、多异常类型、f-string 格式说明符、多 for 子句推导式 |
| `pkg/desugar/desugar.go` | 新增 `desugarAssertStatement`、`desugarDefaultParams`、`desugarSuperCall`、`desugarDataclass`、`foldConstants`、`desugarMultiClauseListComprehension`；`is`/`in`/位运算脱糖；`raise E from e` 脱糖；`FormattedExpression` → `format()` 调用 |
| `pkg/compiler/compiler.go` | 新增 `formatValue`/`formatInteger`/`formatFloat`/`formatString` 函数；`format` 内置函数支持 `format(value, spec)` 语义；`FormattedExpression` 编译支持；字符串驻留；`id`/`range`/`len` 内置函数改进 |
| `pkg/vm/vm.go` | 新增 `getIntegerAttribute`（`__and__`/`__or__`/`__xor__`/`__lshift__`/`__rshift__`/`__invert__`）；OpAdd/OpSub/OpMul 内联快速路径；`nextInstruction` 标签支持 goto 跳转 |
| `pkg/objects/object.go` | 新增 `NATIVE_METHOD_OBJ`/`RANGE_OBJ`/`BOUNDMETHOD_OBJ`；`NativeMethod`/`Range`/`BoundMethod` 类型 |

### v0.6 — 高影响力高难度特性 ✅

目标：实现高级语言特性，接近 Python 3.10+ 兼容。

- [x] P1-5：`match/case` 模式匹配脱糖 — `match expr: case pattern: body` → `if/elif` 链
- [x] P1-6：多继承 MRO — C3 线性化算法 + `SuperClasses` 列表 + `ComputeMRO()`
- [x] P0-15：`match/case` 解析器支持 — MATCH/CASE token + `parseMatchStatement` + `parseCaseClause`
- [x] P0-16：多继承解析器支持 — `class C(A, B):` 逗号分隔多父类
- [x] P3-6：内联缓存 — VM `attrCache` 基于 IP+ObjType 的属性查找缓存

**v0.6 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/ast/ast.go` | 新增 `MatchStatement`（`Subject`/`Cases`）和 `CaseClause`（`Pattern`/`Guard`/`Body`）；`ClassStatement` 新增 `SuperClasses []*Identifier` 字段 |
| `pkg/lexer/lexer.go` | 新增 `MATCH`/`CASE` token 常量；注册 `"match": MATCH`、`"case": CASE` 关键字 |
| `pkg/parser/parser.go` | 新增 `parseMatchStatement`/`parseCaseClause`；`parseClassStatement` 支持逗号分隔多父类；`parseBlockStatement` 终止条件添加 `lexer.CASE` |
| `pkg/desugar/desugar.go` | 新增 `desugarMatchStatement`：match → `_match_val = expr; if/elif` 链；`buildMatchCondition`：字面量/变量/通配符/或/类/列表模式 → 比较表达式；`collectPatternBindings`：变量绑定赋值；`ClassStatement` 脱糖保留 `SuperClasses` |
| `pkg/compiler/compiler.go` | 新增 `OpCreateClassWithMultiSuper` 操作码；`compileMatchStatement` fallback（应被脱糖）；`compileClassStatement` 多父类编译路径 |
| `pkg/vm/vm.go` | 新增 `AttrCacheKey`/`AttrCacheEntry` 类型；VM 新增 `attrCache` 字段；`OpGetAttribute` 内联缓存：Instance/Module 属性查找缓存；`OpCreateClassWithMultiSuper` 处理：弹出多父类、设置 `SuperClasses`、调用 `ComputeMRO()`；`OpCreateClassWithSuper` 改进：同时设置 `SuperClasses` |
| `pkg/objects/object.go` | `Class` 新增 `SuperClasses []*Class` 和 `MRO []*Class` 字段；`Instance.GetAttr` 支持 MRO 查找；新增 `ComputeMRO()` 和 `c3Linearize()` C3 线性化算法 |

### v0.7 — 中影响力特性（性能+语言） ✅

目标：实现中影响力特性，提升运行时性能和语言便利性。

- [x] P3-9：全局变量缓存 — VM `globalCache`/`globalVersions` 版本号缓存
- [x] P1-10：`@lru_cache` 脱糖 — `isLruCacheDecorator` 检测 + `desugarLruCache` 字典记忆化包装
- [x] P1-18：NamedTuple 脱糖 — `desugarNamedTuple` 生成 `__init__` + `__repr__` 类
- [x] P3-8：`zip()` 惰性迭代器 — `Zip` 对象 + `ToList()` 按需物化
- [x] P3-12：对象池 — `GetCachedInteger`（-256~255）+ `GetCachedString`（≤16字符）
- [x] P3-17：Dict 优化 — `KeyOrder` 有序键列表 + `NewDictWithCapacity` 预分配

**v0.7 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/objects/object.go` | 新增 `RANGE_OBJ`/`ZIP_OBJ` 类型；`Range` 结构体（`Start`/`Stop`/`Step`）+ `Len()`/`ToList()`/`GetItem()`；`Zip` 结构体（`Iterables`）+ `Len()`/`ToList()`；`GetCachedInteger`（-256~255 整数池）；`GetCachedString`（≤16字符字符串池）；`Dict` 新增 `KeyOrder []string` 有序键列表 + `NewDictWithCapacity` 预分配构造函数；`Dict.Set`/`Delete` 维护 `KeyOrder` |
| `pkg/vm/vm.go` | 新增 `GlobalCacheEntry` 结构体；VM 新增 `globalCache`/`globalVersions` 字段；`OpGetGlobal` 版本号缓存快速路径；`OpSetGlobal` 递增版本号；`executeRangeIndex`/`executeStringIndex` 新增；`executeTupleIndex` 支持负数索引；`executeBinaryIntegerOperation`/`executeBangOperator` 使用 `GetCachedInteger` |
| `pkg/compiler/compiler.go` | `range` 内置函数返回 `NewRange` 惰性对象；`zip` 内置函数返回 `NewZip` 惰性对象；`len` 支持 `Range`/`Zip` 对象 |
| `pkg/desugar/desugar.go` | 新增 `isLruCacheDecorator`：检测 `@lru_cache` 装饰器；`desugarLruCache`：生成 `_cache` 字典 + 键查找 + 结果缓存包装函数；`desugarNamedTuple`：`NamedTuple('Name', [...])` → 生成带 `__init__` + `__repr__` 的类 |

### v0.8 — 低影响力低难度特性 ✅

目标：完善语言兼容性，补齐低影响力低难度象限。

- [x] P0-10：0x/0b/0o 字面量 — Lexer `readNumber` 支持 `0x`/`0b`/`0o` 前缀，`0o` 自动转换为 Go 兼容的 `0` 前缀
- [x] P0-11：数字下划线 — Lexer `readNumber` 跳过 `_` 并在返回前剥离，支持十进制/十六进制/二进制/八进制/浮点数中的下划线

**v0.8 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/lexer/lexer.go` | `readNumber` 新增 `0x`/`0X`/`0b`/`0B`/`0o`/`0O` 前缀检测分支，分别读取十六进制/二进制/八进制数字；`0o` 前缀自动转换为 Go 兼容的 `0` 前缀；所有数字读取循环支持 `_` 字符；新增 `stripUnderscores` 辅助函数在返回前剥离下划线；新增 `isHexDigit`/`isBinaryDigit`/`isOctalDigit` 辅助函数 |

### v0.9 — 低影响力中难度特性 ✅

目标：实现低影响力中难度象限特性，完善语言兼容性和运行时优化。

- [x] P1-20/P0-17：仅关键字参数 + 默认参数值 — 解析器 `*` 分隔符支持 + `KeywordOnly` 标记 + VM kwargs 字典处理 + 默认值填充
- [x] P1-19：Enum 脱糖 — `class Color(Enum): RED=1, GREEN=2` → 类属性 + `_members_` 字典 + 枚举值对象
- [x] P0-2：Byte strings `b"..."` — Lexer BYTESTRING token + `Bytes` 对象 + VM 索引/切片/比较/len/bool
- [x] P3-14：死代码消除 — `EliminateDeadCode` BFS 可达性分析 + 跳转目标重写 + 函数内 DCE

**v0.9 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/lexer/lexer.go` | 新增 `BYTESTRING` token；`case 'b':` 处理 `b"..."`/`b'...'` 前缀；`readStringWithQuote(quote byte)` 支持引号类型参数；`case '\'':` 单引号字符串处理 |
| `pkg/ast/ast.go` | 新增 `ByteStringLiteral` 结构体（`Token`/`Value` 字段） |
| `pkg/parser/parser.go` | 新增 `parseByteStringLiteral`；注册 `BYTESTRING` 前缀解析器 |
| `pkg/desugar/desugar.go` | 默认参数值脱糖：`if param == None: param = default` + `None` 表达式确保 IfExpression 栈一致性；Enum 脱糖：`class E(Enum):` → 类属性 + `_members_` 字典 |
| `pkg/compiler/compiler.go` | `ByteStringLiteral` 编译为 `OpConstant` + `Bytes` 常量；`None` 常量修复（从 Builtin 改为 `objects.None_`）；`OpSetLocal` 编码修复（`emit` → `emit1`）；`len`/`bool` 内置函数支持 `Bytes` 类型；`Bytecode()` 启用 `EliminateDeadCodeInFunctions` |
| `pkg/compiler/optimize.go` | 新增 `EliminateDeadCode`：BFS 从指令 0 开始可达性分析，terminator（OpReturn/OpReturnValue/OpJump/OpRaise）不跟随 fall-through，重写跳转目标；`EliminateDeadCodeInFunctions`：递归优化 CompiledFunction 常量中的指令；`instructionSize`/`isTerminator`/`readOperand` 辅助函数 |
| `pkg/vm/vm.go` | `executeBytesIndex`/`executeBytesSlice`/`executeBytesComparison`：Bytes 索引/切片/比较；`OpSlice` 修复：弹出 step 值；VarArgs 修复：`posArgsCount == minParams` 时 `vm.push` 替代 `vm.stack` 赋值；Closure 默认参数支持：`NumKeywordOnly`/`NumDefaults`/`NumPositionalDefaults`/`ParameterNames` 字段 + kwargs 字典处理；`isTruthy` 支持 `Bytes` |
| `pkg/objects/object.go` | 新增 `BYTES_OBJ` 类型；`Bytes` 结构体（`Value []byte`）+ `NewBytes` 构造函数 + `Inspect` 输出 `b'...'` 格式；`Equal` 支持 `Bytes` 逐字节比较；`Closure` 新增 `NumKeywordOnly`/`NumDefaults`/`NumPositionalDefaults`/`ParameterNames` 字段 |

**v0.9 Bug 修复：**

- [x] OpSetLocal 编码错误：`c.emit(OpSetLocal, ...)` 使用 2 字节操作数编码但 OpSetLocal 期望 1 字节 → 改为 `c.emit1(OpSetLocal, ...)`
- [x] None 常量注册为 Builtin 函数：`b == None` 始终返回 false → 改为存储 `objects.None_` 常量
- [x] OpSlice step 值栈泄漏：编译器推送 4 值但 VM 只弹出 3 个 → 添加 `step := vm.pop()`
- [x] IfExpression 栈不一致：默认参数脱糖生成的 `if b == None: b = 10` 中 AssignStatement 不留值 → 添加 None 表达式确保栈一致
- [x] OpGetLocal 读取超出 sp：默认参数 IfExpression 的 OpPop 弹出了错误栈位 → 根因是 IfExpression 栈不一致（已修复）
- [x] Closure 不支持默认参数：`make_adder(n=10)` 调用报错 → Closure 结构体新增默认参数字段 + VM 闭包调用路径支持
- [x] VarArgs 空参数覆盖：`greet("Alice")` 中 `args` 覆盖了 `name` → 改用 `vm.push` 替代 `vm.stack` 赋值

### v0.10 — 描述符协议与内置描述符 ✅

目标：实现 Python 描述符协议，支持 property/classmethod/staticmethod 和 __slots__。

- [x] P0-33：描述符协议 — `__get__`/`__set__`/`__delete__`，数据描述符优先于实例属性
- [x] 内置 `property` — getter/setter/deleter，VM `OpGetAttribute`/`OpSetAttribute` 中拦截
- [x] 内置 `classmethod` — `__getattr__` 返回绑定类的方法
- [x] 内置 `staticmethod` — `__getattr__` 返回原始函数
- [x] `__slots__` — 实例属性白名单检查，继承场景下父类 slots 合并
- [x] 属性赋值语法 — `obj.attr = value`（`AttributeAssignStatement` AST 节点）
- [x] 装饰器解析/编译 — `@property`/`@classmethod`/`@staticmethod` 在类体中的解析和编译

**v0.10 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/ast/ast.go` | 新增 `AttributeAssignStatement` 结构体（`Object`/`Attribute`/`Value` 字段） |
| `pkg/parser/parser.go` | 新增 `parseExpressionOrAttrAssign()`；类体中装饰器解析修复 |
| `pkg/objects/object.go` | 新增 `PROPERTY_OBJ`/`CLASSMETHOD_OBJ`/`STATICMETHOD_OBJ` 类型；`Property`/`ClassMethod`/`StaticMethod` 结构体；`Descriptor` 接口 + `IsDescriptor()`/`IsDataDescriptor()` 函数；`Class` 新增 `Slots []string` + `HasSlots()`/`IsSlotAllowed()` |
| `pkg/compiler/compiler.go` | 新增 `property`/`classmethod`/`staticmethod` 内置函数；`compileClassStatement` 处理 `@staticmethod`/`@classmethod` 装饰器 |
| `pkg/vm/vm.go` | `OpGetAttribute`：数据描述符 → 实例属性 → 非数据描述符查找优先级；`OpSetAttribute`：property setter / 数据描述符 `__set__` / `__slots__` 检查；`Frame` 新增 `initInstance`/`setAttrValue` 字段 |

**v0.10 Bug 修复：**

- [x] 编译器 `lastInstruction` 状态泄漏：`compileFunction`/`FunctionLiteral` 进入新作用域时未重置 → 保存/恢复 `lastInstruction`/`previousInstruction`
- [x] `return vm.push(val)` 导致 VM 提前退出 → 改为 `vm.push(val); continue`
- [x] `__init__` 返回值覆盖实例 → `Frame.initInstance` 标记

### v0.11 — try/except 编译器修复 ✅

目标：修复 try/except 的编译器和 VM 问题，确保异常处理正确工作。

- [x] `OpBeginTry` 新增 `handlerIP` 操作数 — 编译器回填第一个 `OpExceptHandler` 的位置
- [x] try 块无异常时不穿透到 except 块 — 编译器在 try body 后始终生成 `OpJump`
- [x] 跨帧异常处理 — `OpRaise`/`raiseException` 正确回退帧到 `try/except` 所在帧
- [x] `matchesException` catch-all — 裸 `except:` 捕获任何类型异常
- [x] DCE 保留异常处理器 — `EliminateDeadCode` 理解 `OpBeginTry` 控制流，标记 `handlerIP` 为可达

**v0.11 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/compiler/compiler.go` | `compileTryStatement`：`OpBeginTry` 新增第 3 个操作数 `handlerIP`（7 字节指令）；try body 后始终生成 `OpJump`；回填 `handlerIP` |
| `pkg/compiler/optimize.go` | `instructionSize`：`OpBeginTry` 从 5 字节改为 7 字节；`EliminateDeadCode`：`OpBeginTry` 标记 `handlerIP` 位置为可达；重写时更新 `handlerIP` |
| `pkg/vm/vm.go` | `OpBeginTry`：读取 `handlerIP` 操作数；`OpRaise`/`OpEndTry`：使用 `handlerIP` 扫描 `OpExceptHandler`；`ExceptionHandler.handlerIP` 在 `OpBeginTry` 时设置；`raiseException`：使用 `handlerIP` 替代 `tryBlockStartIP` 扫描；移除 `handlerIP == -1` 检查，改用 `exceptCount == 0` |

### v0.11.1 — 关键 Bug 修复 + super() 实现 ✅

目标：修复影响核心功能的多个关键 bug，实现 `super()` 内建函数。

**Bug 修复：**

- [x] **if 语句解析 bug**：顶层 `if` 语句被 parser 当作表达式解析，后续语句被误解析为 if 的调用（如 `if true: print("a") print("b")` → `if_expr("b")`）。修复：在 `parseStatement` 中添加 `case lexer.IF`，将顶层 if 作为 `ExpressionStatement` 处理，设置 `lastStmtAdvanced`
- [x] **OpCreateClassWithMultiSuper numParents 读取位置错误**：`numParents := int(ins[ip+1])` 读取了 idx 的低字节而非 `ins[ip+3]`，导致多继承时父类数量解析错误。修复：改为 `numParents := int(ins[ip+3])`
- [x] **attrCache key 缺少 FrameIndex**：`AttrCacheKey{IP: ip, ObjType: ...}` 在不同函数帧中 IP 相同时缓存污染（如 `self.name` 和 `self.sound` 在不同函数中 IP 相同，缓存命中返回错误值）。修复：`AttrCacheKey` 新增 `FrameIndex int` 字段
- [x] **字符串比较 bug**：`OpEqual`/`OpNotEqual` 对非数值类型使用 Go 指针比较，导致 `"abc" == "abc"` 返回 `false`。修复：使用 `objects.Equal()` 进行值比较
- [x] **IfExpression 栈不平衡**：`compileIfExpression` 中 consequence/alternative 块没有值时未 emit `OpNull`，导致栈不平衡。修复：对不以 `OpPop` 结尾的块 emit `OpNull`
- [x] **OpEndTry 异常对象栈泄漏**：except 块处理后 pending error 对象仍留在栈上。修复：在 `OpEndTry` 中检查 `vm.sp > handler.stackPtr` 时 pop 错误对象

**新功能：**

- [x] **`super()` 内建函数**：支持 `super().__init__(args)` 调用模式。在 `executeCall` 中检测 `super()` 调用，从调用者帧中搜索 Instance 对象，返回 `Super{Instance, SuperClass}` 对象
- [x] **`parseDotExpression` infix handler**：注册 `lexer.DOT` 的 infix 解析器，支持 `expr.attr` 和 `expr.method(args)` 语法（如 `super().__init__(name)`）
- [x] **`Super` 对象类型**：`objects.Super` 结构体 + `SUPER_OBJ` 类型 + `OpGetAttribute` 中 Super 对象的属性查找（委托给父类）

**v0.11.1 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/parser/parser.go` | 新增 `case lexer.IF` 在 `parseStatement` 中处理顶层 if 语句；注册 `lexer.DOT` infix handler → `parseDotExpression`；新增 `parseDotExpression` 函数处理 `expr.member` 和 `expr.method(args)` |
| `pkg/compiler/compiler.go` | `compileIfExpression` 修复：consequence/alternative 不以 OpPop 结尾时 emit OpNull；注册 `super` 内建函数 |
| `pkg/vm/vm.go` | `AttrCacheKey` 新增 `FrameIndex` 字段；`OpCreateClassWithMultiSuper` 修复 numParents 读取位置；`executeComparison` 使用 `objects.Equal()`；`executeCall` 新增 `super()` 处理：搜索调用者帧中的 Instance，返回 Super 对象；`OpGetAttribute` 新增 Super 对象属性查找；`OpEndTry` 修复异常对象栈泄漏 |
| `pkg/objects/object.go` | 新增 `SUPER_OBJ` 类型；`Super` 结构体（`Instance`/`SuperClass` 字段） |

---

## 进行中

### Bug 修复

_(v0.12 所有计划中的 bug 修复已完成)_

### 已完成

- [x] **`del obj.attr` 完整支持**：新增 `OpDelAttribute` 操作码 + compiler 编译 `DeleteStatement` + VM 处理 property deleter 和实例属性删除 + desugar 修复 MemberAccess 参数错误
- [x] **`@x.deleter` 端到端测试**：test_property_deleter.py 通过
- [x] **varargs/kwargs 装饰器包装**：`@log_decorator` 包装的函数通过 `*args, **kwargs` 调用时参数数量错误 — Closure 结构体新增 VarArgs/KwArgs 字段，VM executeCall Closure 分支添加 VarArgs/KwArgs 处理，OpListUnpack/OpDictUnpack VM 实现
- [x] **自定义描述符 `__set__`/`__get__`/`__delete__` 在 `__init__` 内参数数量错误**：根因是描述符方法调用的栈布局与 `executeCall` 的自动 Instance 检测冲突——`calleeIndex-1` 位置恰好是 `__init__` 的 `self` 实例，导致多插入一个参数。修复：将 `descInst` 放在 callee 下面，利用自动检测正确添加 `self`，与 property setter/deleter 的调用模式一致
- [x] **try-only-finally 异常穿透**：`try: ... finally:` (无 except) 中抛出异常时，finally 块执行后异常不传播到外层。根因：`OpEndTry` 的 `pendingError` 处理只在当前帧的 `exceptionStack` 中查找处理器，无法跨帧传播。修复：改用 `raiseException()` 让异常正确传播到外层帧
- [x] **嵌套闭包自由变量捕获**：三层嵌套闭包（如 `repeat(times)` → `decorator(func)` → `wrapper()`）无法正确捕获外层自由变量。三个根因：(1) `Resolve` 方法不处理 `FreeScope` 变量传播——外层 free 变量不会传递到内层作用域；(2) 编译器在退出作用域后读取 `FreeSymbols`，但此时 `c.symbolTable` 已恢复为外层——需在退出前保存；(3) `Resolve` 不缓存结果到 `s.store`——同一变量被多次解析时重复添加到 `Free` 列表

### 脱糖层增强

- [ ] 用户自定义 `__getattr__` / `__setattr__` 与 VM 内置描述符的共存测试

### 词法分析器 / 解析器

- [ ] 类体中 `@property` / `@classmethod` / `@staticmethod` 装饰器解析的健壮性改进
- [ ] INDENT/DEDENT 边界情况处理

---

## 路线图

### v0.12 — Bug 修复 + 运行时增强 ✅

> **架构决策**：v0.12 不再将 property/classmethod/staticmethod/__slots__ 迁移到脱糖层。
> 原因：脱糖迁移会导致 (1) 性能崩塌——每次属性访问都要经过方法查找+帧创建+函数调用，
> 远慢于 VM OpGetAttribute/OpSetAttribute 的直接拦截；(2) `__slots__` 失去优化内存的本意——
> 脱糖生成的 `__setattr__` 白名单是运行时方法调用，而 VM 的 HasSlots()/IsSlotAllowed() 是 O(1) 直接判断，
> 且 `__slots__` 在 CPython 中的核心价值是节省 `__dict__` 内存，脱糖方案无法实现；
> (3) 脱糖生成的 `__getattr__`/`__setattr__` 会与用户自定义的方法冲突。
> **结论**：性能关键的内核特性应保留 VM 原生实现，脱糖优先原则应有合理边界。

- [x] try-only-finally 异常穿透修复 — `OpEndTry` 改用 `raiseException()` 传播异常
- [x] 自定义描述符 `__set__`/`__get__`/`__delete__` 参数数量修复 — 栈布局与 executeCall 自动 Instance 检测对齐
- [x] varargs/kwargs 装饰器包装修复 — `OpListUnpack`/`OpDictUnpack` VM 实现 + Closure VarArgs/KwArgs 字段
- [x] `del obj.attr` 完整支持 — `OpDelAttribute` + property deleter
- [x] 嵌套闭包自由变量捕获修复 — `Resolve` FreeScope 传播 + `FreeSymbols` 保存 + `store` 缓存
- [x] `__slots__` 内存优化 — Instance 使用 `SlotValues []Object` 固定数组替代 `map[string]Object`，O(1) 索引访问，节省内存
- [x] range 迭代器死循环修复 — `desugarForToWhile` 改用 `AssignStatement` + 嵌套循环唯一索引变量名

### v0.13 — 剩余高难度特性 ✅

- [x] Metaclasses — `class Foo(metaclass=Meta):` 语法 + metaclass `__call__` 控制实例化 + metaclass `__init__` 自动调用
- [x] 仅位置参数 (/) — 解析器 `SLASH` token + `PositionalOnly` 标记 + VM 位置参数处理
- [x] Exception groups — `BaseExceptionGroup`/`ExceptionGroup` 类型 + `except*` 语法 + 异常组分割

**v0.13 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/ast/ast.go` | `ClassStatement` 新增 `Metaclass *Identifier` 字段；`ExceptClause` 新增 `IsStar bool` 字段 |
| `pkg/lexer/lexer.go` | 新增 `SLASH` token（仅位置参数分隔符）；新增 `STAR` 在 `except*` 上下文中的处理 |
| `pkg/parser/parser.go` | `parseClassStatement` 支持 `metaclass=XXX` 关键字参数解析；`parseExceptClause` 支持 `except*` 语法；仅位置参数 `/` 解析 |
| `pkg/desugar/desugar.go` | `ClassStatement` 脱糖传播 `Metaclass` 字段 |
| `pkg/compiler/compiler.go` | 新增 `OpSetMetaclass`/`OpCallMetaclassInit`/`OpExceptStarHandler` 操作码；`compileClassStatement` metaclass 编译；`compileTryStatement` `except*` 编译；`*ast.Identifier` 识别 `True`/`False`/`None` 内置常量；`compileFunction` 错误路径修复 `exitScope()` |
| `pkg/compiler/optimize.go` | `InstructionSize` 支持 `OpSetMetaclass`/`OpCallMetaclassInit`/`OpExceptStarHandler`；DCE 理解 `OpExceptStarHandler` 控制流 |
| `pkg/vm/vm.go` | `Frame` 新增 `metaclassInitClass *Class` 字段；`OpSetMetaclass`：设置 class.Metaclass；`OpCallMetaclassInit`：自动调用 metaclass `__init__`；`executeCall` metaclass `__call__` 拦截；`OpSetAttribute` 支持 Class 对象属性设置；`OpExceptStarHandler`：异常组分割和匹配；`splitExceptionGroup`/`findNextExceptStarHandler` 辅助函数；`OpReturn`/`OpReturnValue` metaclassInitClass 返回值处理 |
| `pkg/objects/object.go` | `Class` 新增 `Metaclass *Class` 字段；新增 `EXCEPTION_GROUP_OBJ` 类型；`ExceptionGroup` 结构体（`Message`/`Exceptions`）；`NewErrorWithType` 构造函数 |

**v0.13 Bug 修复：**

- [x] `True`/`False`/`None` 未被编译器识别为内置常量：`*ast.Identifier` Resolve 失败时检查 `True`→`OpTrue`、`False`→`OpFalse`、`None`→`OpNull`
- [x] `compileFunction` 作用域泄漏：编译函数体出错时未调用 `exitScope()`，导致符号表永久嵌套在子作用域中，后续所有 `Define()` 创建 LOCAL 而非 GLOBAL 符号
- [x] `OpSetAttribute` 不支持 Class 对象：`cls._registered = True` 需要 Class 对象属性设置，之前只支持 Instance
- [x] `NewError` vet 警告：`NewError(err.Error())` 非常量格式字符串 → `NewError("%s", err.Error())`

### v0.14 — 低影响力特性 ✅

- [x] Ellipsis (...) — `...` 字面量 + `Ellipsis` 标识符 + `OpEllipsis` 操作码
- [x] Complex numbers — `2j` 字面量 + `complex(real, imag)` 内置函数 + 复数算术运算
- [x] Dictionary Views — `dict.keys()`/`dict.values()`/`dict.items()` 返回动态视图对象

**v0.14 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/lexer/lexer.go` | 新增 `ELLIPSIS`/`COMPLEX` token；`readNumber` 支持 `j`/`J` 后缀；`case '.'` 检测 `...` |
| `pkg/ast/ast.go` | 新增 `EllipsisLiteral`/`ComplexLiteral` AST 节点 |
| `pkg/parser/parser.go` | 注册 `ELLIPSIS`/`COMPLEX` 前缀解析器 |
| `pkg/desugar/desugar.go` | `ComplexLiteral` 模式匹配支持 |
| `pkg/compiler/compiler.go` | `OpEllipsis` 操作码；`ComplexLiteral` 编译；`complex()`/`abs()`/`bool()` 内置函数支持复数；`Ellipsis` 标识符识别 |
| `pkg/vm/vm.go` | `OpEllipsis` 处理；复数算术运算（`executeBinaryComplexOperation`）；复数比较/取反/真值；Dict 视图对象属性访问和索引 |
| `pkg/objects/object.go` | `Ellipsis` 单例对象；`Complex` 结构体；`DictKeys`/`DictValues`/`DictItems` 视图对象；`Equal` 支持 Ellipsis/Complex |

### v0.15 — 标准库与运行时 ✅

- [x] re 正则表达式模块
- [x] Async comprehensions
- [x] 分代 GC
- [x] 寄存器 VM

**v0.15 变更详情：**

| 模块 | 变更 |
|------|------|
| `pkg/re/module.go` | 新增 `re` 模块，实现 `compile`/`search`/`match`/`fullmatch`/`findall`/`finditer`/`sub`/`subn`/`split`/`escape` 10 个函数和 `IGNORECASE`/`MULTILINE`/`DOTALL`/`ASCII`/`UNICODE` 5 个常量 |
| `pkg/objects/object.go` | 新增 `REGEX_PATTERN_OBJ`/`REGEX_MATCH_OBJ` 类型；`RegexPattern` 结构体（`Regexp`/`Pattern`/`Flags`）+ `GetAttr()` 支持 `pattern`/`flags`/`search`/`match`/`fullmatch`/`findall`/`finditer`/`sub`/`subn`/`split`；`RegexMatch` 结构体（`Groups_`/`GroupIndices`/`GroupEnds`/`OrigString`/`Pattern_`）+ `GetAttr()` 支持 `group`/`start`/`end`/`span`/`groups`/`string`/`re`/`lastindex` |
| `pkg/vm/vm.go` | `OpGetAttribute` 支持 `RegexPattern`/`RegexMatch` 属性访问；新增 `useRegisterVM`/`regVM` 字段；新增 `NewRegisterVM()` 构造器 |
| `pkg/vm/register_vm.go` | 寄存器 VM 实现：`RegisterVM` 结构体 + 40+ 寄存器操作码 + 栈式字节码翻译器 + 寄存器执行器 + 翻译缓存 |
| `pkg/compiler/compiler.go` | 注册 `re` 模块；新增 `compileAsyncListComprehension`/`compileAsyncSetComprehension`/`compileAsyncDictComprehension`/`compileAsyncGeneratorExpression` |
| `pkg/ast/ast.go` | 新增 `AsyncListComprehension`/`AsyncSetComprehension`/`AsyncDictComprehension`/`AsyncGeneratorExpression` AST 节点 |
| `pkg/parser/parser.go` | 解析器支持 `[x async for x in iter]`/`{x async for x in iter}`/`{k:v async for x in iter}`/`(x async for x in iter)` 语法 |
| `pkg/desugar/desugar.go` | 脱糖层支持异步推导式节点，脱糖子表达式并保留 async 语义 |
| `pkg/gc/gc.go` | 分代 GC：`YoungGen`/`OldGen` 双代；`MinorCollect()`/`MajorCollect()` 双收集器；`promotionAge=3` 晋升机制；`WriteBarrier()` + `rememberedSet` 写屏障；`youngThreshold=256KB`/`oldThreshold=4MB` 阈值；`minorAfterMajor=10` 自动触发 Major GC |

**v0.15 Code Review 已知问题：**

#### 🔴 严重问题 (3 项)

1. **异步推导式编译器缺少循环结构**：`compileAsyncListComprehension` 等函数只编译了 iterable、element、filter，但没有生成 `for` 循环和列表构建的字节码。当前实现只是编译子表达式然后 `OpReturnValue`，不会产生正确的列表结果。需要生成类似 `result = []; for x in iter: if filter: result.append(element); return result` 的字节码
2. **异步推导式 filter 跳转未回填**：`jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)` 使用了占位符 `9999` 但从未回填目标地址，导致 filter 为 false 时跳转到错误位置
3. **寄存器 VM 翻译器跳转目标不一致**：翻译器使用栈式字节码的 IP 作为 `RegOpJump`/`RegOpJumpIfFalse` 的目标，但寄存器指令的 IP 与栈式指令的 IP 不对应（寄存器指令数量 ≠ 栈式指令数量），导致跳转目标错误

#### 🟡 中等问题 (5 项)

4. **re.sub 不支持 count 参数**：`re.sub(pattern, repl, string, count=0)` 的 `count` 参数被忽略，始终替换所有匹配
5. **re.subn 替换计数不准确**：`subn` 通过 `FindAllString` 重新搜索来计算替换次数，而非统计实际替换次数。如果 `repl` 包含 `$` 引用，替换结果和计数可能不一致
6. **分代 GC markObject 线性扫描**：`markObject`/`markObjectMinor` 通过遍历 `youngObjects`/`oldObjects` 切片查找对象，时间复杂度 O(n)。应使用 `map[objects.Object]*GCObject` 索引加速
7. **分代 GC 并发安全风险**：`MinorCollect` 中 `gc.mu.Unlock()` 后调用 `MajorCollect()`，但 `MajorCollect` 又获取 `gc.mu.Lock()`。在解锁和重新加锁之间可能有其他 goroutine 修改 GC 状态
8. **寄存器 VM 寄存器泄漏**：翻译器中 `regAlloc` 只增不减，对于循环体内的指令，每次迭代都会分配新寄存器，导致寄存器数量无限增长。需要寄存器分配器回收不再使用的寄存器

#### 🟢 低优先级问题 (4 项)

9. **RegexMatch.Groups_ 命名不规范**：`Groups_`/`GroupIndices`/`GroupEnds`/`OrigString`/`Pattern_` 使用下划线后缀/前缀不一致，应统一为 Go 惯用命名（如 `Groups`/`StartPositions`/`EndPositions`/`OriginalString`/`Pattern`）
10. **re 模块 flags 处理使用魔术数字**：`flagsToGoFlags` 中 `flags&2`/`flags&8`/`flags&16` 应使用常量名 `IGNORECASE`/`MULTILINE`/`DOTALL`
11. **寄存器 VM 翻译缓存哈希冲突**：`bytecodeHash` 使用简单哈希函数，不同字节码可能产生相同哈希导致缓存命中错误结果
12. **分代 GC markReferences 不追踪 SlotValues**：`markReferences`/`markReferencesMinor` 中 `Instance` 只遍历 `Fields`，不遍历 `SlotValues`，导致 slotted instance 的属性引用可能被错误回收

### v0.16 — Bug 修复与 Python 语义对齐 ✅

#### 🔴 严重语义错误 (已全部修复)

- [x] **整数除法返回类型错误** → v0.15.1 修复：OpDiv 返回 Float
- [x] **整数取模符号不符合 Python 语义** → v0.15.1 修复：Python 风格取模
- [x] **List 负索引不支持** → v0.15.1 修复：添加负索引转换
- [x] **越界索引返回 None 而非 IndexError** → v0.15.1 修复：抛出 IndexError
- [x] **异步推导式编译器缺少循环结构** → v0.15.1 修复：生成正确 for 循环字节码
- [x] **异步推导式 filter 跳转未回填** → v0.15.1 修复：正确回填跳转目标
- [x] **寄存器 VM 翻译器跳转目标不一致** → v0.15.1 修复：stackIPToRegIP 映射

#### 🟡 中等语义偏差 (大部分已修复)

- [x] **Boolean 不参与算术运算** → v0.15.1 修复
- [x] **字符串乘法不支持** → v0.15.1 修复
- [x] **列表拼接不支持** → v0.15.1 修复
- [x] **OpAdd 对非字符串类型错误拼接** → v0.15.1 修复：抛出 TypeError
- [x] **re.sub/re.subn 不支持 count 参数** → v0.15.1 修复
- [x] **re.subn 替换计数不准确** → v0.15.1 修复
- [x] **re.findall 不处理可选组** → v0.15.1 修复
- [x] **re.split 不保留分隔符** → v0.15.1 修复
- [ ] **re 模块不支持 callable 替换**：Python 中 `re.sub(pattern, lambda m: m.group(1).upper(), string)` 支持 callable 作为 repl
- [x] **Dict/Set 允许不可哈希类型作为键** → v0.15.1 修复：CheckHashable
- [x] **StopIteration 不是异常类型** → v0.15.1 修复
- [x] **整数负数幂运算错误** → v0.15.1 修复：返回 Float
- [x] **分代 GC markObject 线性扫描 O(n)** → v0.15.1 修复：objectMap O(1)
- [x] **分代 GC MinorCollect → MajorCollect 并发安全风险** → v0.15.1 修复：majorCollectLocked
- [ ] **寄存器 VM 寄存器泄漏**：`regAlloc` 只增不减（已知限制，不影响正确性）

#### 🟢 缺失功能

- [ ] **缺失内置函数**：`isinstance()`、`issubclass()`、`hasattr()`、`getattr()`、`setattr()`、`dir()`、`id()`、`hash()`、`callable()`、`enumerate()`、`map()`、`filter()`、`sorted()`、`reversed()`、`repr()`、`iter()`、`any()`、`all()`、`chr()`、`ord()`、`hex()`、`oct()`、`bin()`、`format()`
- [ ] **缺失异常类型**：`StopIteration`、`OverflowError`、`FileNotFoundError`、`ImportError`、`SyntaxError`、`IndentationError`、`UnboundLocalError`、`RecursionError`、`MemoryError`
- [ ] **缺失字符串方法**：`str.rfind()`、`str.rindex()`、`str.count()`、`str.isdigit()`、`str.isalpha()`、`str.isalnum()`、`str.isspace()`、`str.isupper()`、`str.islower()`、`str.istitle()`、`str.capitalize()`、`str.title()`、`str.swapcase()`、`str.center()`、`str.ljust()`、`str.rjust()`、`str.zfill()`、`str.partition()`、`str.rpartition()`、`str.encode()`、`str.format_map()`、`str.expandtabs()`、`str.isdecimal()`、`str.isnumeric()`、`str.isidentifier()`、`str.isprintable()`、`str.maketrans()`、`str.translate()`
- [ ] **缺失列表方法**：`list.sort()`（带 key 和 reverse 参数）、`list.__iadd__`（`lst += [4]` 原地扩展）
- [ ] **缺失字典方法**：`dict.fromkeys()`、`dict.__ior__`（`dict |= other`）
- [ ] **缺失集合运算符**：`set | set`、`set & set`、`set - set`、`set ^ set`、`set |= set`、`set &= set` 等
- [ ] **分代 GC markReferences 追踪 SlotValues**
- [ ] **RegexMatch 字段命名规范化**（`Groups_` → `Groups`，`Pattern_` → `Pattern`）
- [x] **re 模块 flags 使用常量名替代魔术数字** → v0.15.1 修复

### v0.17 — 标准库补全 (计划中)

- [ ] 补全缺失内置函数（isinstance, getattr, setattr, dir, id, hash, callable, enumerate, map, filter, sorted, reversed, repr, iter, any, all, chr, ord, hex, oct, bin, format）
- [ ] 补全字符串方法（rfind, rindex, count, isdigit, isalpha, isalnum, isspace, capitalize, title, swapcase, center, ljust, rjust, zfill, partition, encode 等）
- [ ] 补全列表方法（sort with key/reverse, __iadd__）
- [ ] 补全集合运算符（|, &, -, ^, |=, &=, -=, ^=）
- [ ] 补全异常类型（StopIteration, FileNotFoundError, ImportError, SyntaxError 等）
- [ ] io 模块完善（StringIO, BytesIO）
- [ ] collections 模块（defaultdict, Counter, OrderedDict, deque）

### v1.0.0 — Production Ready

- [ ] 完整的 asyncio 模块
- [ ] 性能基准测试达标
- [ ] 生产环境验证

---

## 架构约束

1. **脱糖层不修改内核** — 脱糖层只能生成由现有内核支持的 AST 节点和内置函数调用
2. **内置函数最小化** — 新增内置函数需有充分理由，优先使用现有指令组合
3. **向后兼容** — 脱糖转换不应改变用户可见的语义行为
4. **可调试性** — 脱糖生成的混淆方法名（`_desugar_*`）应保持可追溯性
5. **脱糖优先原则** — 新增语法特性优先考虑脱糖实现，仅在性能关键路径上引入内核支持

---

## 测试策略

### 测试覆盖目标

| 模块 | 当前覆盖 | 目标覆盖 |
|------|----------|----------|
| Parser | 75% | 95% |
| AST | 80% | 95% |
| Compiler | 70% | 90% |
| VM | 65% | 85% |
| Objects | 60% | 85% |
| Desugar | 70% | 90% |
| Concurrency | 55% | 80% |

### 成功指标

- [ ] 支持 95% 的核心 Python 语法
- [ ] 执行速度达到 CPython 的 80%+
- [ ] 测试覆盖率 > 80%
- [ ] 关键路径测试覆盖率 > 95%

---

## 风险评估

| 风险 | 影响 | 可能性 | 缓解策略 |
|------|------|--------|----------|
| 脱糖迁移引入回归 | 高 | 中 | 端到端测试覆盖 + 逐步迁移 |
| Metaclasses 实现复杂度 | 高 | 中 | 分阶段实现，对标 CPython |
| 性能回归 | 高 | 低 | 完整的性能测试套件 |
| try-only-finally 修复影响现有异常处理 | 中 | 低 | 回归测试 + 增量修改 |

---

## 术语表

- **AST**: Abstract Syntax Tree，抽象语法树
- **Desugar**: 语法脱糖，将高级语法转换为低级语法
- **DCE**: Dead Code Elimination，死代码消除
- **JIT**: Just-In-Time，运行时编译
- **MRO**: Method Resolution Order，方法解析顺序
- **VM**: Virtual Machine，虚拟机
- **GC**: Garbage Collection，垃圾回收
