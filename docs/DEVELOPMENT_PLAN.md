# GoPy 开发计划

**当前版本**: 0.31.x
**目标版本**: 1.0.0
**最后更新**: 2026-06

---

## 目录

1. [设计理念](#设计理念)
2. [当前实现状态](#当前实现状态)
3. [路线图](#路线图)
4. [架构约束](#架构约束)
5. [测试策略](#测试策略)
6. [风险评估](#风险评估)

---

## 设计理念

GoPy 采用**脱糖优先**（Desugar-First）的架构设计。核心原则是：**尽可能在脱糖层处理高级语法特性，将编译器和虚拟机保持为最小内核**。

这样做的好处：
1. **内核简洁** — 编译器和虚拟机只需处理基础语义，无需感知装饰器、slots 等高级特性
2. **可维护性** — 高级特性的变更集中在脱糖层，不污染内核代码
3. **可测试性** — 脱糖输出是纯 AST，可以独立验证转换正确性
4. **可扩展性** — 新增语法特性只需在脱糖层添加转换规则

**脱糖优先的边界**：性能关键的内核特性应保留 VM 原生实现（如 property/classmethod/staticmethod/__slots__），脱糖迁移会导致性能崩塌和语义冲突。

---

## 当前实现状态

### 1. 基础语法 (完成度: 100%)

✅ 整数、浮点数、布尔值、字符串、复数、字节串
✅ 数组、字典、集合
✅ 基本算术运算 (+, -, *, /, %, //, **)
✅ 比较运算 (==, !=, >, <, >=, <=) + 链式比较
✅ 布尔运算 (and, or, not)
✅ 位运算符 (& | ^ ~ << >>)
✅ 集合运算符 (| & - ^)
✅ 变量绑定和作用域
✅ 多重赋值/元组解包
✅ 增强赋值 (+=, -=, *=, /=, %=, **=, |=, &=, ^=, <<=, >>=)
✅ 索引赋值 (d['a'] = 1) + 索引增强赋值 (d['a'] += 1)
✅ 函数定义和调用 + 关键字参数 + *args/**kwargs + 仅关键字参数 + 仅位置参数
✅ 条件语句 (if/elif/else) + 三元表达式
✅ 循环语句 (for/while) + break/continue
✅ 推导式 (列表/集合/字典/生成器/异步)
✅ 切片操作（支持步长）
✅ pass / assert / del / global / nonlocal 语句
✅ Walrus 运算符 (:=)
✅ yield / yield from / async for / async with
✅ match/case 模式匹配
✅ f-string + Raw strings + 三引号 + Byte strings
✅ 0x/0b/0o 字面量 + 数字下划线 + Ellipsis + 复数字面量
✅ 类型注解
✅ 成员测试 (in / not in)

### 2. 高级特性 (完成度: 95%)

✅ 异常处理 (try/except/finally) + 跨帧异常捕获 + 异常链 + Exception groups
✅ 上下文管理器 (with)
✅ 生成器 (yield, yield from)
✅ Lambda 表达式和闭包
✅ 类、对象、单继承/多继承 + C3 MRO
✅ 装饰器 + 描述符协议 + property/classmethod/staticmethod
✅ __slots__ + @dataclass + @abstractmethod + @lru_cache
✅ NamedTuple + Enum + Metaclasses
✅ 模块导入系统

### 3. 并发特性 (完成度: 70%)

✅ Goroutine 协程 + Channel 通道 + 协程调度器
✅ async/await 语法 + 异步对象
✅ 并发安全数据结构 + 同步原语
✅ OpAwait 完善实现 — 同步执行 async 帧并缓存结果
✅ async for/with 脱糖实现
✅ asyncio 模块 — run/create_task/sleep/gather/Event/Future/wait_for/open/create_subprocess_exec

### 4. 运行时优化 (完成度: 90%)

✅ 内联缓存 + 全局变量缓存 + 特化操作码 + 常量折叠
✅ 对象池 + Dict 优化 + 死代码消除 + 字符串驻留
✅ BoundMethod + Range/Zip 惰性迭代器
✅ 寄存器 VM（直接编译器后端）+ 分代 GC
✅ 内联缓存泛化（OpIndex/OpSetIndex）+ 快速整数算术 + StringBuilder + 列表预分配
✅ JIT 热点检测 + 逃逸分析
✅ 寄存器 VM — 直接编译器后端，寄存器帧隔离，全操作码覆盖，内联缓存，原生函数调用
✅ 寄存器 VM — 完整操作符支持（<<, >>, >=, <=, in, not in, and, or）+ IfExpression 结果返回修复
⚠️ 直接线程 — 未实现

### 5. 标准库 (完成度: 90%)

✅ math, sys, os, json, gc, random, string, time, datetime
✅ re (正则表达式), io (StringIO/BytesIO), concurrency
✅ collections (defaultdict, Counter, OrderedDict, deque, namedtuple, ChainMap)
✅ 调试器 + 性能分析器 + JIT + CPython 互操作
✅ functools (reduce, partial, wraps, lru_cache, cached_property, total_ordering, singledispatch, cmp_to_key) / operator / collections.abc (16个ABC) / pathlib (基础)
✅ typing (List/Dict/Tuple/Set/Optional/Union/Any/Callable/TypeVar/Generic/Protocol/Literal/Final/ParamSpec/Concatenate/TypeAlias/TypeGuard/Never/NoReturn/Self/Unpack/TypeVarTuple) / hashlib (md5/sha1/sha224/sha256/sha384/sha512/sha3_*/blake2b/blake2s/shake_128/shake_256)
✅ base64 (b16/b32/b64/a85/b85 encode/decode) / struct / itertools (19个函数)
✅ sys.getsizeof 真实实现
✅ asyncio 模块 — run/create_task/sleep/gather/Event/Future/wait_for/open/create_subprocess_exec
✅ math — atan2, copysign, fmod, isnan, isinf, isfinite, factorial, gcd, lcm, inf, nan, tau
✅ os — rmdir, makedirs, removedirs, stat, os.path 子模块 (exists/isfile/isdir/join/split/basename/dirname/getsize/abspath)
✅ json — dump, load
✅ random — randrange, sample, choices, gauss
✅ time — strftime, strptime, gmtime, mktime, time_ns, monotonic, perf_counter, timezone, tzname
✅ sys — maxsize, byteorder, executable, prefix, modules, flags
✅ datetime — Instance 对象返回，strptime/fromtimestamp/fromisoformat/timezone.utc/min/max
✅ string — capwords, Template, Formatter

### 6. 内置类型方法 (完成度: 95%)

✅ str — find, index, replace, split, rsplit, splitlines, join, strip, lstrip, rstrip, upper, lower, startswith, endswith, format, casefold, maketrans, rfind, rindex, count, isdigit, isalpha, isalnum, isspace, isupper, islower, istitle, capitalize, title, swapcase, center, ljust, rjust, zfill, partition, rpartition, encode, isdecimal, isnumeric, isidentifier, isprintable, expandtabs, translate, format_map
✅ list — append, extend, insert, remove, pop, clear, index, count, reverse, copy, sort, __imul__, __iadd__
✅ dict — fromkeys, update, setdefault, popitem, pop, get, clear, copy, keys, values, items
✅ set — add, remove, discard, pop, clear, union, intersection, difference, symmetric_difference, issubset, issuperset, update, copy, intersection_update, symmetric_difference_update, __isub__, difference_update, __ior__, __iand__, __ixor__
✅ tuple — count, index
✅ int — bit_length, bit_count, to_bytes, __class__
✅ float — is_integer, hex, as_integer_ratio, __class__
✅ bytes — decode, hex, len, __len__, __class__
✅ frozenset — union, intersection, issubset, issuperset, copy, __class__
✅ memoryview — tobytes, hex, len, __len__, __getitem__, readonly, format, itemsize, nbytes
✅ slice — start, stop, step, indices()
✅ FileObject — read, readline, write, close, flush, name, closed

### 7. 内置函数 (完成度: 95%)

✅ print, len, range, set, open, next, type, str, int, float, bool, abs, complex, list, min, max, sum, format, input, round, zip, enumerate, property, classmethod, staticmethod, super, ExceptionGroup, isinstance, issubclass, hasattr, getattr, setattr, dir, id, hash, callable, map, filter, sorted, reversed, repr, iter, any, all, chr, ord, hex, oct, bin, divmod, frozenset, delattr, vars, ascii, object, slice, bytearray, bytes, pow (3-arg), eval, exec, globals, locals, __import__, compile, memoryview, breakpoint

---

## 路线图

> v0.3 ~ v0.27 的完整开发路径已归档至 [DEVELOPMENT_HISTORY.md](DEVELOPMENT_HISTORY.md)。

### v0.30 — 寄存器 VM 功能完善 ✅

> 目标：完善寄存器 VM 的操作符覆盖和表达式编译，对齐栈式 VM 功能。

#### 寄存器编译器操作符补齐

- [x] **R4.1**: 位移操作符 — `<<` / `>>` 编译为 ROpLShift/ROpRShift
- [x] **R4.2**: 比较操作符 — `>=` / `<=` 编译为 ROpCompare(OpGreaterEqual/OpLessEqual)
- [x] **R4.3**: 成员测试操作符 — `in` / `not in` 编译为 ROpContains/ROpNotContains
- [x] **R4.4**: 布尔短路操作符 — `and` / `or` 编译为 JumpIfFalse + Move 模式
- [x] **R4.5**: `not` 关键字 — PrefixExpression 支持 `not` 操作符

#### 寄存器 VM 操作码补齐

- [x] **R4.6**: RegOpLShift/RegOpRShift — 整数位移快速路径 + fallback
- [x] **R4.7**: RegOpContains/RegOpNotContains — List/Tuple/Set/Dict/String 成员测试
- [x] **R4.8**: compareOp 扩展 — OpGreaterEqual/OpLessEqual 支持（Integer/Float/通用）

#### 寄存器编译器表达式修复

- [x] **R4.9**: IfExpression 结果返回 — 从 BlockStatement 提取单表达式结果，不再总是返回 Null
- [x] **R4.10**: 栈 VM 操作码扩展 — 新增 OpGreaterEqual/OpLessEqual

#### 新增测试

- [x] **R4.11**: TestRegisterVMLShiftRShift — 位移操作
- [x] **R4.12**: TestRegisterVMContainsOp / TestRegisterVMNotContainsOp — 成员测试
- [x] **R4.13**: TestRegisterVMComparisonGTELT — >= / <= 比较
- [x] **R4.14**: TestRegisterVMIfExpressionResult — If 表达式结果返回

### v0.31 — 寄存器 VM 异常处理与默认参数 ✅

> 目标：完善寄存器 VM 的异常处理和默认参数支持，实现与栈式 VM 功能对齐。

#### 异常处理实现

- [x] **R5.1**: try/except 编译器回填 — ROpBeginTry 的 handlerIP 和 finallyStartIP 回填
- [x] **R5.2**: 寄存器 VM 异常处理 — 扫描 RegInstruction 找匹配的 RegOpExceptHandler
- [x] **R5.3**: 异常类型 ErrorType 字段 — ValueError/TypeError 等构造函数设置 ErrorType
- [x] **R5.4**: RegOpEndTry 重新抛出 — pendingError 使用寄存器 VM 方式处理

#### 默认参数支持

- [x] **R5.5**: 函数调用参数填充 — 缺失参数用 None 填充（Closure/CompiledFunction/BoundMethod 路径）

#### Dict 方法支持

- [x] **R5.6**: getAttrOp 使用 Dict.GetAttr — 返回可调用的 Builtin 方法
- [x] **R5.7**: Dict merge 操作符 — `|` 操作符支持 d1 | d2

#### 测试结果

- [x] **R5.8**: 366/366 RegVM 测试通过（之前 135/144，9 失败）
- [x] **R5.9**: 全项目测试通过

### v1.0.0 — Production Ready

- [x] 所有 v0.18-v0.25 里程碑完成
- [x] 寄存器 VM 为可用执行引擎，栈式 VM 为默认
- [x] 性能基准测试达标（CPython 80%+）
- [ ] 测试覆盖率 > 80%（当前：Lexer 97.9%, Desugar 99.6%, Parser 81.7%, Compiler 78.3%, VM 52.4%）
- [x] 关键路径测试覆盖率 > 95%（Lexer, Desugar 已达标）
- [x] 跨平台验证（Linux/macOS/Windows/WASM）
- [x] v0.28 内置方法与标准库补齐
- [x] v0.29 标准库与内置方法深度补齐
- [x] v0.30 寄存器 VM 功能完善
- [x] v0.31 寄存器 VM 异常处理与默认参数
- [ ] 生产环境验证

---

## 架构约束

1. **脱糖层不修改内核** — 脱糖层只能生成由现有内核支持的 AST 节点和内置函数调用
2. **内置函数最小化** — 新增内置函数需有充分理由，优先使用现有指令组合
3. **向后兼容** — 脱糖转换不应改变用户可见的语义行为
4. **可调试性** — 脱糖生成的混淆方法名（`_desugar_*`）应保持可追溯性
5. **脱糖优先原则** — 新增语法特性优先考虑脱糖实现，仅在性能关键路径上引入内核支持
6. **错误类型正确性** — 运行时错误必须使用正确的异常类型（IndexError/KeyError/TypeError/ZeroDivisionError 等），禁止使用 `fmt.Errorf` 替代

---

## 测试策略

### 测试覆盖目标

| 模块 | v0.26 | v0.27 | 目标覆盖 |
|------|-------|-------|----------|
| Lexer | 67.4% | 97.9% | 90% ✅ |
| Parser | 26.3% | 81.7% | 80% ✅ |
| Desugar | 74.9% | 99.6% | 90% ✅ |
| Compiler | 27.2% | 78.3% | 80% |
| VM | 15.5% | 52.4% | 85% |
| Objects | 20.7% | 90.3% | 85% ✅ |
| GC | - | 93.3% | 80% ✅ |
| RE | - | 95.0% | 85% ✅ |
| Struct | - | 93.7% | 80% ✅ |

### 成功指标

- [ ] 支持 95% 的核心 Python 语法
- [ ] 执行速度达到 CPython 的 80%+
- [ ] 测试覆盖率 > 80%
- [ ] 关键路径测试覆盖率 > 95%
- [ ] 零 🔴 严重 Bug

---

## 风险评估

| 风险 | 影响 | 可能性 | 缓解策略 |
|------|------|--------|----------|
| DCE 跳转重写不完整导致字节码损坏 | 高 | 中 | 端到端测试 + fuzzing |
| 寄存器 VM 属性访问不一致 | 中 | 中 | 统一 getAttrOp/setAttrOp 逻辑 |
| 并发模块线程安全 | 中 | 中 | 竞态检测 + 压力测试 |
| JIT 优化引入正确性回归 | 高 | 低 | 优化前后结果对比测试 |
| asyncio 实现复杂度 | 高 | 中 | 分阶段实现，对标 CPython 子集 |
| JIT 框架大量空函数体 | 中 | 高 | v0.24 分阶段实现核心优化 pass |
| 寄存器 VM 迁移 — 翻译层到直接编译器后端切换 | 高 | 中 | 三版本渐进迁移，Phase 1 保留翻译层，Phase 2 双后端并行，Phase 3 切换默认 |
| 寄存器 VM 函数调用原生实现复杂度 | 高 | 中 | Phase 2 逐步从回退模式迁移到原生实现，保留回退路径 |
| 跨平台兼容性 — 寄存器分配器平台差异 | 中 | 低 | 纯 Go 实现，无平台相关代码；WASM 环境需验证寄存器数量限制 |

---

## 术语表

- **AST**: Abstract Syntax Tree，抽象语法树
- **Desugar**: 语法脱糖，将高级语法转换为低级语法
- **DCE**: Dead Code Elimination，死代码消除
- **JIT**: Just-In-Time，运行时编译
- **MRO**: Method Resolution Order，方法解析顺序
- **VM**: Virtual Machine，虚拟机
- **GC**: Garbage Collection，垃圾回收
- **CPython**: C 语言实现的 Python 参考实现
