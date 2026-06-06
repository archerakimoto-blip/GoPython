# GoPy 开发历史

> 本文档归档了 v0.3 ~ v0.17.x 的完整开发路径。当前开发计划见 [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md)。

---

## v0.3 — 装饰器脱糖重构 ✅

将 `@property`、`@classmethod`、`@staticmethod` 和 `__slots__` 的处理从内核迁移至脱糖层。

**`@property`** → `__getattr__` 中添加 getter 调用
**`@property` + `@x.setter`** → `__getattr__` + `__setattr__` 中添加 getter/setter
**`@classmethod`** → `__getattr__` 中返回 `__bind_method__(cls._desugar_cm_foo, cls)`
**`@staticmethod`** → `__getattr__` 中返回 `cls._desugar_sm_bar`
**`__slots__`** → `__setattr__` 白名单检查

## v0.4 — 高影响力低难度特性 ✅

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

## v0.5 — 高影响力中难度特性 ✅

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

## v0.6 — 高影响力高难度特性 ✅

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

## v0.7 — 中影响力特性（性能+语言） ✅

目标：实现中影响力特性，提升运行时性能和语言便利性。

- [x] P3-9：全局变量缓存 — VM `globalCache`/`globalVersions` 版本号缓存
- [x] P1-10：`@lru_cache` 脱糖 — `isLruCacheDecorator` 检测 + `desugarLruCache` 字典记忆化包装
- [x] P1-18：NamedTuple 脱糖 — `desugarNamedTuple` 生成 `__init__` + `__repr__` 类
- [x] P3-8：`zip()` 惰性迭代器 — `Zip` 对象 + `ToList()` 按需物化
- [x] P3-12：对象池 — `GetCachedInteger`（-256~255）+ `GetCachedString`（≤16字符）
- [x] P3-17：Dict 优化 — `KeyOrder` 有序键列表 + `NewDictWithCapacity` 预分配

## v0.8 — 低影响力低难度特性 ✅

- [x] P0-10：0x/0b/0o 字面量 — Lexer `readNumber` 支持 `0x`/`0b`/`0o` 前缀
- [x] P0-11：数字下划线 — Lexer `readNumber` 跳过 `_` 并在返回前剥离

## v0.9 — 低影响力中难度特性 ✅

- [x] P1-20/P0-17：仅关键字参数 + 默认参数值
- [x] P1-19：Enum 脱糖
- [x] P0-2：Byte strings `b"..."`
- [x] P3-14：死代码消除 — `EliminateDeadCode` BFS 可达性分析

**v0.9 Bug 修复：**

- [x] OpSetLocal 编码错误：`c.emit` → `c.emit1`
- [x] None 常量注册为 Builtin 函数 → 改为存储 `objects.None_` 常量
- [x] OpSlice step 值栈泄漏 → 添加 `step := vm.pop()`
- [x] IfExpression 栈不一致 → 添加 None 表达式确保栈一致
- [x] OpGetLocal 读取超出 sp → 根因是 IfExpression 栈不一致（已修复）
- [x] Closure 不支持默认参数 → Closure 结构体新增默认参数字段
- [x] VarArgs 空参数覆盖 → 改用 `vm.push` 替代 `vm.stack` 赋值

## v0.10 — 描述符协议与内置描述符 ✅

- [x] P0-33：描述符协议 — `__get__`/`__set__`/`__delete__`
- [x] 内置 `property`/`classmethod`/`staticmethod`
- [x] `__slots__` — 实例属性白名单检查
- [x] 属性赋值语法 — `obj.attr = value`（`AttributeAssignStatement`）

**v0.10 Bug 修复：**

- [x] 编译器 `lastInstruction` 状态泄漏 → 保存/恢复
- [x] `return vm.push(val)` 导致 VM 提前退出 → 改为 `vm.push(val); continue`
- [x] `__init__` 返回值覆盖实例 → `Frame.initInstance` 标记

## v0.11 — try/except 编译器修复 ✅

- [x] `OpBeginTry` 新增 `handlerIP` 操作数
- [x] try 块无异常时不穿透到 except 块
- [x] 跨帧异常处理
- [x] `matchesException` catch-all
- [x] DCE 保留异常处理器

## v0.11.1 — 关键 Bug 修复 + super() 实现 ✅

**Bug 修复：**

- [x] if 语句解析 bug
- [x] OpCreateClassWithMultiSuper numParents 读取位置错误
- [x] attrCache key 缺少 FrameIndex
- [x] 字符串比较 bug — 使用 `objects.Equal()`
- [x] IfExpression 栈不平衡
- [x] OpEndTry 异常对象栈泄漏

**新功能：**

- [x] `super()` 内建函数
- [x] `parseDotExpression` infix handler
- [x] `Super` 对象类型

## v0.12 — Bug 修复 + 运行时增强 ✅

> **架构决策**：v0.12 不再将 property/classmethod/staticmethod/__slots__ 迁移到脱糖层。
> 原因：脱糖迁移会导致性能崩塌、`__slots__` 失去优化内存本意、与用户自定义方法冲突。
> **结论**：性能关键的内核特性应保留 VM 原生实现，脱糖优先原则应有合理边界。

- [x] try-only-finally 异常穿透修复
- [x] 自定义描述符 `__set__`/`__get__`/`__delete__` 参数数量修复
- [x] varargs/kwargs 装饰器包装修复
- [x] `del obj.attr` 完整支持
- [x] 嵌套闭包自由变量捕获修复
- [x] `__slots__` 内存优化 — `SlotValues []Object` 固定数组
- [x] range 迭代器死循环修复

## v0.13 — 剩余高难度特性 ✅

- [x] Metaclasses — `class Foo(metaclass=Meta):` + metaclass `__call__`/`__init__`
- [x] 仅位置参数 (/) — SLASH token + PositionalOnly 标记
- [x] Exception groups — `BaseExceptionGroup`/`ExceptionGroup` + `except*` 语法

**v0.13 Bug 修复：**

- [x] `True`/`False`/`None` 未被编译器识别为内置常量
- [x] `compileFunction` 作用域泄漏
- [x] `OpSetAttribute` 不支持 Class 对象
- [x] `NewError` vet 警告

## v0.14 — 低影响力特性 ✅

- [x] Ellipsis (...) — `...` 字面量 + `Ellipsis` 标识符 + `OpEllipsis`
- [x] Complex numbers — `2j` 字面量 + `complex()` + 复数算术
- [x] Dictionary Views — `dict.keys()`/`dict.values()`/`dict.items()` 动态视图

## v0.15 — 标准库与运行时 ✅

- [x] re 正则表达式模块
- [x] Async comprehensions
- [x] 分代 GC
- [x] 寄存器 VM

**v0.15 Code Review 已知问题（已在 v0.15.1/v0.16 修复）：**

- 异步推导式编译器缺少循环结构 → ✅ 修复
- 异步推导式 filter 跳转未回填 → ✅ 修复
- 寄存器 VM 翻译器跳转目标不一致 → ✅ 修复
- re.sub 不支持 count 参数 → ✅ 修复
- 分代 GC markObject 线性扫描 O(n) → ✅ 修复
- 分代 GC 并发安全风险 → ✅ 修复
- 寄存器 VM 寄存器泄漏 → ✅ 修复

## v0.16 — Bug 修复与 Python 语义对齐 ✅

#### 🔴 严重语义错误 (已全部修复)

- [x] 整数除法返回类型错误 → OpDiv 返回 Float
- [x] 整数取模符号不符合 Python 语义 → Python 风格取模
- [x] List 负索引不支持 → 添加负索引转换
- [x] 越界索引返回 None 而非 IndexError → 抛出 IndexError
- [x] Boolean 不参与算术运算 → 修复
- [x] 字符串乘法不支持 → 修复
- [x] 列表拼接不支持 → 修复
- [x] OpAdd 对非字符串类型错误拼接 → 抛出 TypeError
- [x] Dict/Set 允许不可哈希类型作为键 → CheckHashable
- [x] 整数负数幂运算错误 → 返回 Float

#### 🟢 缺失功能 (已全部补齐)

- [x] 缺失内置函数 → isinstance, issubclass, hasattr, getattr, setattr, dir, id, hash, callable, enumerate, map, filter, sorted, reversed, repr, iter, any, all, chr, ord, hex, oct, bin, format
- [x] 缺失异常类型 → StopIteration, OverflowError, FileNotFoundError, ImportError, SyntaxError, IndentationError, UnboundLocalError, RecursionError, MemoryError
- [x] 缺失字符串方法 → rfind, rindex, count, isdigit, isalpha, isalnum, isspace, isupper, islower, istitle, capitalize, title, swapcase, center, ljust, rjust, zfill, partition, rpartition, encode, isdecimal, isnumeric, isidentifier, isprintable, expandtabs
- [x] 缺失列表方法 → list.sort(key, reverse)
- [x] 缺失字典方法 → dict.fromkeys()
- [x] 缺失集合运算符 → set.union, intersection, difference, symmetric_difference, issubset, issuperset, update, copy
- [x] 分代 GC markReferences 追踪 SlotValues
- [x] RegexMatch 字段命名规范化
- [x] re 模块 flags 使用常量名替代魔术数字

## v0.17 — 标准库补全 ✅

- [x] 补全剩余字符串方法（maketrans, translate, format_map）
- [x] 补全列表方法（__iadd__ 原地扩展）
- [x] 补全集合方法（union, intersection, difference, symmetric_difference 等）
- [x] 补全字典方法（__ior__ 运算符重载）
- [x] io 模块完善（StringIO, BytesIO）
- [x] collections 模块（defaultdict, Counter, OrderedDict, deque）

## v0.17.x — 标准库补全（续）✅

- [x] 集合字面量解析修复 — `parseBraceLiteral` 正确识别 `{1, 2, 3}`
- [x] `set()` 内置函数添加
- [x] `int()` 内置函数支持无参数调用
- [x] 集合运算符 `|` `&` `-` `^` 作为运算符可用
- [x] VM 索引运算符支持自定义对象（`__getitem__`/`__setitem__` via GetAttr）
- [x] 索引赋值 `d['a'] = 1` — `IndexAssignStatement` + `OpSetIndex`
- [x] 索引增强赋值 `d['a'] += 1` — `AugAssignStatement` 扩展 + 脱糖
- [x] 集合增强赋值 `|=` `&=` `^=` — Parser + Desugar
- [x] collections 类型 `__setitem__`/`__getitem__` — defaultdict, Counter, OrderedDict
