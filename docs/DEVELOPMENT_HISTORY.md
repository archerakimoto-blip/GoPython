# GoPy 开发历史

> 本文档归档了 v0.3 ~ v0.22 的完整开发路径。当前开发计划见 [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md)。

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

## v0.18 — Code Review Bug 修复 + 稳定性 ✅

> 目标：修复全量 Code Review 发现的所有严重和中等问题，提升运行时稳定性。

### 🔴 严重 Bug 修复

- [x] **B1**: `executeSetIndex` List 越界错误返回方式修复 — `vm.push(NewIndexError(...))` → `fmt.Errorf(...)`
- [x] **B2**: `OpListUnpack` 中 `unpackExtraArgs` 使用后重置 — 已存在重置逻辑（误报）
- [x] **B3**: `sweepYoung` 晋升年龄检查逻辑修正 — 逻辑正确（误报）
- [x] **B4**: `binaryFloatOp` 浮点除零检查 — 添加 `rightValue == 0` 检查
- [x] **B5**: `StringIO.read`/`BytesIO.read` size=0 处理 — 添加 `size == 0` 快速返回
- [x] **B6**: `BytesIO.seek` 负位置处理 — 已正确处理（误报）

### 🟡 中等问题修复

- [x] **M2**: `deque` 添加 `__getitem__`/`__setitem__` 支持 — 支持负索引和越界检查
- [x] **M3**: `deque.insert` 负索引行为与 CPython 对齐 — `insert(-1, x)` 插入到末尾元素前
- [x] **M4**: `deque.rotate` 负旋转逻辑修正 — 使用高效的 slice 旋转替代循环
- [x] **M5**: `OrderedDict.update` 顺序保持修复 — 已有键不重复添加到 KeyOrder
- [x] **M6**: `scheduler.Go` nil closure 检查 — 添加 `fn == nil` 返回 nil
- [x] **M7**: `sleep` 负数参数验证 — 返回 `ValueError`
- [x] **M8**: DCE `OpFinally` 跳转目标重写 — 添加 OpFinally 可达性分析和目标重写
- [x] **M10**: `getAttrOp` slots/字段访问行为统一 — slot 存在但值为 nil 时返回 None_
- [x] **M12**: `re.sub`/`subn` callable 替换参数验证 — 已有 `IsCallable` 检查（误报）
- [x] **M13/M14**: `StringIO.write`/`BytesIO.write` 参数类型验证 — 返回 TypeError + 返回写入字节数
- [x] **M15**: 整数除零使用 `NewZeroDivisionError` — 已使用（误报）
- [x] **M17**: Profiler 并发保护 — 添加 `sync.Mutex` 到 RecordInstruction/EnterFunction/ExitFunction

### 🟢 代码质量改进

- [x] **L3**: 提取 `augAssignOperator(tokenType) (string, bool)` 消除重复映射
- [x] **L16**: 统一注释语言为英文 — 范围过大，延后到后续版本

## v0.19 — Python 语义完善 ✅

> 目标：补齐剩余的 Python 语义偏差，提升 CPython 兼容性。

- [x] **集合差集运算符** `s - other` — OpSub 已支持 Set 类型（已存在）
- [x] **切片赋值** `lst[1:3] = [4,5]` — 新增 `SliceAssignStatement` AST 节点 + `OpSetSlice` 操作码 + VM `executeSetSlice`
- [x] **多目标索引赋值** `d['a'] = d['b'] = 1` — 验证通过，当前实现正确
- [x] **Tuple 索引赋值禁用** — `t[0] = 1` 抛出 TypeError
- [x] **String 索引赋值禁用** — `s[0] = 'a'` 抛出 TypeError
- [x] **`dict.update()` 接受关键字参数** — 支持 Dict/List[Tuple]/Tuple[Tuple] + kwargs
- [x] **`dict | other` 合并运算符** — `d1 | d2` 创建新 dict（右覆盖左）
- [x] **`dict.setdefault()`** — `d.setdefault('a', 0)` 方法
- [x] **`dict.popitem()`** — LIFO 顺序弹出，返回 (key, value) 元组
- [x] **`dict.pop()`** — 按键弹出，支持默认值
- [x] **`dict.get()`** — 按键获取，支持默认值
- [x] **`dict.clear()`** / **`dict.copy()`** — 清空/浅拷贝
- [x] **`dict.keys()`** / **`dict.values()`** / **`dict.items()`** — 视图对象
- [x] **`Counter.most_common()`** — 返回最常见元素（已存在）
- [x] **`Counter.elements()`** — 返回迭代器（已存在）
- [x] **`OrderedDict.popitem(last=True)`** — 支持 FIFO/LIFO 弹出
- [x] **`OrderedDict.move_to_end(last=True)`** — 移动键到末尾/开头
- [x] **`deque.__getitem__`/`__setitem__`** — 索引访问支持（v0.18 已完成）
- [x] **`deque.maxlen`** — 最大长度属性（已存在，返回 None）
- [x] **`deque.remove()`** — 按值删除（已存在）
- [x] **`deque.__contains__`** — `in` 运算符
- [x] **`deque.index()`** — 查找元素位置（已存在）
- [x] **`deque.reverse()`** — 原地反转（已存在）
- [x] **`deque.copy()`** — 浅拷贝
- [x] **`deque.clear()`** — 清空（已存在）
- [x] **`deque.count()`** — 计数（已存在）
- [x] **`deque.extendleft()`** — 左侧扩展（已存在）
- [x] **`deque.rotate(n)`** — 旋转修正（v0.18 已完成）
- [x] **`set.__isub__`** — `s -= other` 原地差集
- [x] **`list.__imul__`** — `lst *= 3` 原地重复
- [x] **`deque.sort()`** — CPython deque 不支持 sort，移除

## v0.20 — 标准库扩展 ✅

> 目标：扩展标准库覆盖面，补齐高频使用的模块。

- [x] **functools 模块** — reduce, partial
- [x] **operator 模块** — itemgetter, attrgetter, methodcaller
- [x] **collections.abc 模块** — Iterable, Sequence, Mapping, Set 抽象基类
- [x] **pathlib 模块**（基础） — Path 对象, exists/is_file/is_dir
- [x] **typing 模块**（基础） — List, Dict, Tuple, Optional, Union 类型别名
- [x] **hashlib 模块**（基础） — md5, sha256
- [x] **base64 模块** — encode/decode
- [x] **struct 模块** — pack/unpack 二进制数据
- ⚠️ **itertools 模块** — v0.20 标记为 ✅ 但实际完全缺失，已在 v0.22 补齐

## v0.21 — 运行时优化 ✅

> 目标：提升运行时性能，优化热点路径。

- [x] **内联缓存泛化** — OpIndex/OpSetIndex 添加 indexCache，缓存类型分派结果（list+int, tuple+int, dict, range+int, string+int, bytes+int 等 10 种处理器）
- [x] **快速整数算术** — 二元操作（OpAdd/OpSub/OpMul/OpDiv/OpMod/OpFloorDiv/OpPower/OpBitOr/OpBitAnd/OpBitXor）添加 int+int 快速路径，避免函数调用开销；OpDiv 返回 Float 保持 Python 语义
- [x] **字符串构建优化** — StringBuilder 对象 + OpStringBuilderCreate/Append/Build 操作码 + 字符串 += 快速路径（使用 strings.Builder 避免 O(n²) 重复分配）
- [x] **列表预分配** — OpArrayPrealloc 操作码 + VM 处理 + 编译器端检测 range(N) 常量模式并生成预分配指令
- [x] **JIT 热点检测** — VM 集成 JIT 引擎，在 executeCall 中记录 CompiledFunction/Closure 调用次数，超过阈值（默认5次）标记为热点函数；添加 GetJITStats/SetJITHotThreshold/GetJITHotFunctions/ClearJITCache API
- [x] **逃逸分析** — CompiledFunction 添加 NonEscapingLocals 位图，编译器 optimize.go 中实现 analyzeEscape 分析 pass，检测闭包捕获（OpGetFree）、返回值（OpGetLocal+OpReturnValue）、全局赋值（OpSetGlobal）等逃逸模式
- [x] **快速整数算术 bug 修复** — 修复 goto 跳过变量声明的编译错误；修复 OpDiv 快速路径返回 Integer 而非 Float 的 Python 语义错误

## v0.22 — 未完善功能补齐 + 标准库扩展 ✅

> 目标：修复全代码审查发现的未完善实现，补齐缺失的标准库模块。

### VM/编译器修复

- [x] **V1**: OpAwait 完善实现 — 同步执行 async 帧并缓存结果，OpReturnValue 检测 async 对象并标记 Done
- [x] **V2**: OpInPlaceLShift/OpInPlaceRShift 添加 inPlaceAttrMap 条目 + augAssignToInPlaceOp 映射
- [x] **V3**: StringBuilder 编译器端生成路径 — WhileStatement 循环体扫描 s += expr 模式，生成 OpStringBuilderCreate/Append/Build
- [x] **V4**: OpArrayPrealloc 编译器端生成路径 — 脱糖 for 循环 range(N) 常量模式检测，生成 OpArrayPrealloc
- [x] **V5**: OpYield 标记为 Deprecated 死操作码（保留定义避免 iota 值偏移）
- [x] **V6**: RegOpDictUnpack 文档化 — 字典已通过寄存器传递给 RegOpCall，executeCall 自动检测
- [x] **V7**: 寄存器 VM slice step 完善 — 新增 sliceOpWithStep 支持 step 参数（正/负步长，List/String/Bytes/Tuple）

### 标准库补齐

- [x] **S1**: functools 扩展 — wraps, lru_cache, cached_property, total_ordering, singledispatch, update_wrapper（LRU 缓存使用双向链表实现）
- [x] **S2**: typing 扩展 — TypeVar, Generic, Protocol, Literal, Final, TypeAlias, ParamSpec, Concatenate
- [x] **S3**: hashlib 扩展 — sha1, sha224, sha384, sha512, sha3_224/256/384/512 + algorithms_available/algorithms_guaranteed
- [x] **S4**: collections.abc 扩展 — Container, Iterator, MutableSequence, ByteString, MutableSet, MutableMapping, MappingView, ItemsView, KeysView, ValuesView, Reversible（11 个新 ABC）
- [x] **S5**: itertools 模块 — chain, count, cycle, islice, repeat, accumulate, product, permutations, combinations, groupby, starmap, filterfalse, zip_longest, tee, pairwise（15 个函数，全部惰性迭代器实现）
- [x] **S7**: sys.getsizeof 真实实现 — 根据对象类型返回 CPython 对齐的内存大小

### Parser/脱糖层

- [x] **P1**: 返回类型注解保留 — parseExpression(LOWEST) 解析类型表达式，存储到 FunctionLiteral.ReturnType
- [x] **P2/P3**: async for/with 脱糖实现 — async for → while+await __anext__+StopAsyncIteration；async with → await __aenter__/__aexit__

## v0.23 — 寄存器 VM 迁移 Phase 1 + asyncio ✅

> 目标：补齐寄存器 VM 缺失操作码，引入直接编译器后端，完善 asyncio 生态。

### 寄存器 VM 操作码补齐

- [x] **R1.1**: 位运算操作码 — RegOpBitOr/RegOpBitAnd/RegOpBitXor + 翻译层映射
- [x] **R1.2**: 集合运算操作码 — RegOpSetUnion/RegOpSetIntersection/RegOpSetDifference/RegOpSetSymmetricDifference
- [x] **R1.3**: 原地操作码 — RegOpInPlaceAdd/Sub/Mul/Div/Mod/FloorDiv/Power/BitOr/BitAnd/BitXor/LShift/RShift
- [x] **R1.4**: RegOpSetIndex + 内联缓存快速路径
- [x] **R1.5**: RegOpSetSlice
- [x] **R1.6**: RegOpStringBuilderCreate/RegOpStringBuilderAppend/RegOpStringBuilderBuild
- [x] **R1.7**: RegOpArrayPrealloc

### 直接编译器后端

- [x] **R1.8**: RegisterCompiler 框架 — 直接生成 []RegInstruction，allocReg/freeReg 寄存器分配
- [x] **R1.9**: 基础表达式编译 — 常量加载、二元运算、比较运算、一元运算
- [x] **R1.10**: 变量存取编译 — GetGlobal/SetGlobal/GetLocal/SetLocal/GetFree
- [x] **R1.11**: 控制流编译 — if/while 跳转指令生成
- [x] **R1.12**: 函数/闭包编译 — FunctionLiteral 编译后定义函数名到符号表

### asyncio 模块

- [x] **asyncio.run** — 运行协程入口
- [x] **asyncio.create_task** — 创建任务
- [x] **asyncio.sleep** — 异步休眠
- [x] **asyncio.gather** — 并发执行多个协程
- [x] **asyncio.Event** — 事件对象

## v0.24 — 寄存器 VM 迁移 Phase 2 + JIT 框架 ✅

> 目标：寄存器 VM 功能对齐栈式 VM，引入内联缓存和性能优化，实现 JIT 框架核心。

### 内联缓存与性能优化

- [x] **R2.1**: 属性访问内联缓存 — attrCache 迁移到寄存器 VM
- [x] **R2.2**: 索引访问内联缓存 — indexCache 迁移到寄存器 VM
- [x] **R2.3**: 全局变量缓存 — globalCache 验证正确性
- [x] **R2.4**: 快速整数算术 — RegOpAdd/Sub/Mul 等添加 int+int 快速路径
- [x] **R2.5**: 属性访问完整实现 — getAttrOp 对齐栈式 VM

### 高级功能对齐

- [x] **R2.6**: 函数调用原生实现 — RegOpCall 原生实现 Builtin/Callable/CompiledFunction/Closure/Class
- [x] **R2.7**: 闭包调用原生实现 — 原生处理闭包参数传递
- [x] **R2.8**: 生成器/异步原生实现 — RegOpMakeGenerator/RegOpMakeAsync/RegOpYieldValue/RegOpAwait
- [x] **R2.9**: 异常处理完善 — try/except/finally/raise IP 映射
- [x] **R2.10**: 上下文管理器完善 — RegOpEnterContext/RegOpExitContext

### 编译器寄存器后端扩展

- [x] **R2.11**: 函数/闭包编译 — FunctionLiteral/Closure 完整编译
- [x] **R2.12**: 数据结构编译 — 列表/字典/集合字面量
- [x] **R2.13**: 属性访问编译 — get/set/del attribute
- [x] **R2.14**: 类定义编译 — 类创建/继承
- [x] **R2.15**: 异常处理编译 — try/except/finally

### JIT 框架实现

- [x] **J1**: copyPropagation — 复写传播优化 pass
- [x] **J2**: registerAllocation — 寄存器分配优化 pass
- [x] **J3**: loopOptimizations — 循环优化 pass（不变量外提+强度削减）
- [x] **J4**: findTargetFunction — 内联优化目标函数查找
- [x] **J5/J6**: ExecuteFunction/Compile — JIT 编译后函数执行

## v0.25 — 寄存器 VM 迁移 Phase 3 + 生产就绪 ✅

> 目标：寄存器 VM 成为可用执行引擎，栈式 VM 保持为默认，跨平台验证。

### 编译器端切换

- [x] **R3.1**: 寄存器后端覆盖全部 AST 节点 — ~35 种 AST 节点可直接编译为寄存器指令
- [x] **R3.2**: 编译器模式选择 — `--vm=register`/`--vm=stack` 命令行参数
- [x] **R3.3**: 字节码序列化格式 — GPYC 魔数 + 版本号 + 常量 + 指令 + NumRegs
- [x] **R3.4**: 翻译层移除 — 删除 translate()/RunReg()/registerAllocator 等翻译层代码（~2400 行）

### 性能验证与优化

- [x] **R3.5**: 性能基准测试 — Stack VM vs Register VM 基准测试，函数调用场景快 ~44%
- [x] **R3.8**: 栈式 VM 兼容模式 — 保留栈式 VM 作为默认，通过 `--vm=stack` 启用

### 关键 Bug 修复

- [x] **寄存器帧隔离**：RegFrame 新增 regBase 字段，regSet/regGet 自动加上当前帧偏移，修复函数调用覆盖调用者寄存器的 bug
- [x] **executeRegFrame 操作码补全**：补全 BuildList/BuildDict/BuildSet/Index/Slice/GetAttr/SetAttr/Closure/BitOr/BitAnd/BitXor/InPlace*/SetIndex/SetSlice/StringBuilder/ArrayPrealloc/Class/Exception/Context 等所有缺失操作码
- [x] **compilerToVMOpcode 映射**：编译器和 VM 操作码排序不同，通过显式映射表转换
- [x] **FunctionLiteral 未绑定函数名**：编译后添加符号定义和 ROpSetGlobal/ROpSetLocal
- [x] **registerBuiltins stub 修复**：print/len 使用实际实现替代返回 None 的 stub

---

## v0.26 — 测试覆盖率基础建设

> 目标：关键路径测试率全覆盖，从 0% 建立测试基础设施。

### 新增测试文件

- [x] **lexer_test.go** — 10 个测试，覆盖主要 token 类型
- [x] **parser_test.go** — 21 个测试，覆盖主要语句和表达式类型
- [x] **desugar_test.go** — 64 个测试，覆盖所有脱糖变换
- [x] **object_test.go** — 56 个测试，覆盖主要对象类型
- [x] **compiler_test.go** — 20 个测试，覆盖基础编译
- [x] **register_compiler_test.go** — 20 个测试，覆盖寄存器编译器
- [x] **vm_test.go** — 15 个测试，覆盖栈式 VM
- [x] **register_vm_test.go** — 15 个测试，覆盖寄存器 VM

### 覆盖率结果

| 模块 | 覆盖率 |
|------|--------|
| Lexer | 67.4% |
| Desugar | 74.9% |
| Parser | 26.3% |
| Objects | 20.7% |
| Compiler | 27.2% |
| VM (Stack) | 19.2% |
| VM (Register) | 15.5% |

---

## v0.27 — 测试覆盖率大幅提升

> 目标：测试覆盖率达到 90%+，修复发现的 Parser bug。

### 测试覆盖率提升

| 模块 | v0.26 | v0.27 | 提升 |
|------|-------|-------|------|
| Lexer | 67.4% | **97.9%** | +30.5% |
| Desugar | 74.9% | **99.6%** | +24.7% |
| Parser | 26.3% | **81.7%** | +55.4% |
| Objects | 20.7% | **90.3%** | +69.6% |
| Compiler | 27.2% | **78.3%** | +51.1% |
| VM | 15.5% | **52.4%** | +36.9% |
| GC | - | **93.3%** | 新增 |
| RE | - | **95.0%** | 新增 |
| Struct | - | **93.7%** | 新增 |

### Parser Bug 修复

- [x] **`yield from` 语句解析**：`from` 被词法分析为 FROM 关键字而非 IDENT，导致 `yield from` 语句无法识别。修复：在 `parseYieldStatement` 中添加 `p.curTokenIs(lexer.FROM)` 检查
- [x] **`match/case` 语句解析**：三个 bug 修复：
  1. `parseMatchStatement` 未 advance past MATCH 关键字，导致 `parseExpression(LOWEST)` 尝试解析 MATCH 作为表达式
  2. `parseMatchStatement` 中 `parseExpression` 返回后 curToken 不在 COLON 上，需要检查 peekToken
  3. `parseCaseClause` 中同样需要处理 curToken/peekToken 与 COLON 的关系
- [x] **`break`/`continue` 语句**：在 `parseStatement` 的 IDENT case 中添加 break/continue 检查，确保在 while/for 循环体内正确解析

### 新增测试文件

- [x] **lexer_test.go** — 50 个测试，覆盖所有 token 类型、字面量、运算符、关键字
- [x] **desugar_test.go** — 75 个测试，覆盖所有 15 种脱糖变换
- [x] **parser_test.go** — 150+ 个测试，覆盖所有语句和表达式类型
- [x] **object_test.go** — 100+ 个测试，覆盖所有对象类型和模块创建
- [x] **compiler_test.go** — 90+ 个测试，覆盖编译器、符号表、寄存器编译器、序列化
- [x] **gc_test.go** — 30+ 个测试，覆盖 GC 收集、写屏障、finalizer
- [x] **re/module_test.go** — 30+ 个测试，覆盖正则表达式所有操作
- [x] **struct/module_test.go** — 30+ 个测试，覆盖所有格式码和字节序
- [x] **vm/vm_test_new_test.go** — 200+ 个测试，覆盖栈式 VM 操作码
- [x] **vm/vm_coverage_boost_test.go** — 200+ 个测试，覆盖 VM 内部函数和寄存器 VM
