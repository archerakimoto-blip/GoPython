# GoPy 开发计划

**当前版本**: 0.22.x
**目标版本**: 1.0.0
**最后更新**: 2026-06

---

## 目录

1. [设计理念](#设计理念)
2. [当前实现状态](#当前实现状态)
3. [全量 Code Review (v0.18)](#全量-code-review-v018)
4. [全代码未完善功能审查 (v0.21)](#全代码未完善功能审查-v021)
5. [路线图](#路线图)
6. [架构约束](#架构约束)
7. [测试策略](#测试策略)
8. [风险评估](#风险评估)
9. [开发历史](#开发历史)

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
✅ 增强赋值 (+=, -=, *=, /=, %=, **=, |=, &=, ^=)
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
❌ asyncio 模块 — 完全缺失

### 4. 运行时优化 (完成度: 85%)

✅ 内联缓存 + 全局变量缓存 + 特化操作码 + 常量折叠
✅ 对象池 + Dict 优化 + 死代码消除 + 字符串驻留
✅ BoundMethod + Range/Zip 惰性迭代器
✅ 寄存器 VM（翻译层模式）+ 分代 GC
✅ 内联缓存泛化（OpIndex/OpSetIndex）+ 快速整数算术 + StringBuilder + 列表预分配
✅ JIT 热点检测 + 逃逸分析
✅ OpInPlaceLShift/RShift inPlaceAttrMap 条目
✅ 寄存器 VM slice step 支持
✅ StringBuilder 操作码 — VM 已处理，编译器端 WhileStatement 循环模式检测生成 OpStringBuilderCreate/Append/Build
✅ OpArrayPrealloc 操作码 — VM 已处理，编译器端脱糖 for 循环 range(N) 常量模式检测生成预分配指令
⚠️ JIT 框架 — 热点检测已实现，但 copyPropagation/registerAllocation/loopOptimizations 为空函数体，ExecuteFunction 返回 nil
⚠️ 直接线程 — 未实现
⚠️ 寄存器 VM — 翻译层模式运行，缺失位运算/集合运算/原地操作/SetIndex/SetSlice/StringBuilder/ArrayPrealloc 操作码，无内联缓存，函数调用回退到栈式 VM

### 5. 标准库 (完成度: 90%)

✅ math, sys, os, json, gc, random, string, time, datetime
✅ re (正则表达式), io (StringIO/BytesIO), concurrency
✅ collections (defaultdict, Counter, OrderedDict, deque)
✅ 调试器 + 性能分析器 + JIT + CPython 互操作
✅ functools (reduce, partial, wraps, lru_cache, cached_property, total_ordering, singledispatch) / operator / collections.abc (16个ABC) / pathlib (基础)
✅ typing (List/Dict/Tuple/Set/Optional/Union/Any/Callable/TypeVar/Generic/Protocol/Literal/Final) / hashlib (md5/sha1/sha224/sha256/sha384/sha512/sha3_*)
✅ base64 / struct / itertools (15个函数)
✅ sys.getsizeof 真实实现
❌ asyncio 模块 — 完全缺失

---

## 全量 Code Review (v0.18)

> 基于 2026-06 对 `pkg/` 全部 17 个源文件的逐行审查。

### 🔴 严重 Bug (必须修复)

| # | 模块 | 位置 | 问题 | 影响 |
|---|------|------|------|------|
| B1 | vm | `vm.go:3276` | `executeSetIndex` 对 List 越界索引使用 `vm.push(NewIndexError(...))` 而非返回 error | 索引越界错误被当作正常值压栈，赋值操作静默"成功" |
| B2 | vm | `vm.go:478-489` | `OpListUnpack` 中 `vm.unpackExtraArgs` 未在使用后重置 | 后续 OpCall 调用参数数量错误 |
| B3 | gc | `gc.go:440-475` | `sweepYoung` 晋升年龄检查可能遗漏应晋升的对象 | 对象可能被过早回收 |
| B4 | register_vm | `register_vm.go:1848-1860` | `binaryFloatOp` 除法不检查除零 | 浮点除零导致 Go runtime panic |
| B5 | io | `module.go:38-50` | `StringIO.read` 不验证 size 参数非负 | 负数 size 导致异常行为 |
| B6 | io | `module.go:218-253` | `BytesIO.seek` 负位置处理不一致 | seek 行为与 CPython 不符 |

### 🟡 中等问题 (应修复)

| # | 模块 | 位置 | 问题 | 影响 |
|---|------|------|------|------|
| M1 | vm | `vm.go:3260-3295` | `executeSetIndex` 对 `__setitem__` 返回值忽略，始终 push 原 value | 如果 `__setitem__` 做了类型转换，VM 不会反映 |
| M2 | collections | `module.go` | `deque` 缺少 `__getitem__`/`__setitem__` | `d[0]` 和 `d[0]=x` 不工作，与其他 collections 类型不一致 |
| M3 | collections | `module.go:140-200` | `deque.insert` 负索引重置为 0 而非报错 | 与 CPython 行为不符 |
| M4 | collections | `module.go:233-258` | `deque.rotate` 负旋转逻辑可能不正确 | 旋转结果错误 |
| M5 | collections | `module.go:417-432` | `OrderedDict.update` 从另一个 OrderedDict 更新时可能丢失顺序 | 顺序不一致 |
| M6 | concurrency | `scheduler.go:138-168` | `Go` 方法不检查 closure 为 nil | nil 指针 panic |
| M7 | concurrency | `module.go:150-170` | `sleep` 不验证负数参数 | 负数休眠时间行为未定义 |
| M8 | compiler | `optimize.go:212-215` | DCE 跳转目标重写可能不完整 | 消除代码后跳转目标错误 |
| M9 | gc | `gc.go:503-520` | `WriteBarrier` 可能重复插入 remembered set | 不必要的性能开销 |
| M10 | register_vm | `register_vm.go:2217-2280` | `getAttrOp` 对 slots 和普通字段访问行为不一致 | 属性访问结果错误 |
| M11 | register_vm | `register_vm.go:1940-1950` | `compareOp` 不同类型比较缺乏错误处理 | 不支持的比较静默返回错误结果 |
| M12 | re | `module.go:300-380` | `sub`/`subn` callable 替换不验证参数类型 | 类型断言 panic |
| M13 | io | `module.go:59-65` | `StringIO.write` 不验证参数是否为字符串 | 类型断言 panic |
| M14 | io | `module.go:206-215` | `BytesIO.write` 不验证参数是否为 Bytes | 类型断言 panic |
| M15 | vm | `vm.go:2900-2901` | 整数除零使用 `fmt.Errorf` 而非 `NewZeroDivisionError` | 异常类型不正确 |
| M16 | debugger | `debugger.go:55-75` | `ShouldBreak` 对 stepOut/stepOver 深度判断可能不正确 | 调试器断点行为错误 |
| M17 | profiler | `profiler.go:20-40` | Profiler 的 map 字段无并发保护 | 并发访问数据竞争 |

### 🟢 低优先级 / 代码质量

| # | 模块 | 问题 |
|---|------|------|
| L1 | ast | `AugAssignStatement` 使用 `IndexLeft`/`IndexIndex` 联合字段区分两种赋值类型，设计不清晰 |
| L2 | ast | `ListComprehension`/`SetComprehension`/`DictComprehension` 的 `String()` 方法不包含 filter 表达式 |
| L3 | parser | `parseAugAssignStatement` 和 `parseExpressionOrAttrAssign` 中运算符映射表重复 3 次 |
| L4 | vm | `executeIndexExpression` 使用 type comparison switch，`executeSetIndex` 使用 type switch，风格不一致 |
| L5 | vm | 命名不一致：`varName` vs `variableName` |
| L6 | jit | `detectLoops` 可能遗漏某些后向跳转模式 |
| L7 | jit | `calculateComplexity` 只考虑部分操作码 |
| L8 | jit | `optimizeFunction` 优化逻辑过于简单，大部分字节码不变 |
| L9 | debugger | `printStack` 不验证 frame 有效性，可能 nil 指针 |
| L10 | debugger | `printGlobals` 只打印前 20 个变量，无截断提示 |
| L11 | profiler | `ExitFunction` 不检查 timerStack 为空的情况 |
| L12 | desugar | `desugarWithStatement` 空上下文管理器列表返回 nil |
| L13 | desugar | `desugarAsyncForStatement` 是占位实现，未实际脱糖 |
| L14 | compiler | `Compile` 对 `ForStatement` 直接报错，依赖脱糖层保证不出现，但缺少保护 |
| L15 | compiler | `HashLiteral` 元素处理顺序不确定 |
| L16 | 全局 | 新增代码中中英文注释混用，应统一 |

---

## 全代码未完善功能审查 (v0.21)

> 基于 2026-06 对 `pkg/` 全代码的未完善功能审查，覆盖 VM/编译器/对象系统/标准库/脱糖层/JIT 框架。

### 🔴 严重问题 (状态标记错误)

| # | 模块 | 位置 | 问题 | 影响 |
|---|------|------|------|------|
| I1 | stdlib | itertools | **itertools 模块完全缺失** — v0.20 开发计划标记为 ✅，但代码库中无任何 itertools 相关文件 | `import itertools` 将失败；开发计划状态与实际不符 |

### 🟡 VM/编译器未完善实现

| # | 模块 | 位置 | 问题 | 影响 |
|---|------|------|------|------|
| V1 | vm | `vm.go:1304-1326` | **OpAwait 占位实现** — async 对象未完成时返回 None 而非真正等待 | `await` 语义不正确，异步代码无法正确执行 |
| V2 | vm | `vm.go:3183-3197` | **OpInPlaceLShift/OpInPlaceRShift 缺少 inPlaceAttrMap 条目** — 操作码已定义，编译器有 `augAssignToInPlaceOp` 映射，但 VM 的 `inPlaceAttrMap` 中无对应条目 | `x <<= 1` / `x >>= 1` 原地操作走 fallback 路径，不会调用 `__ilshift__`/`__irshift__` |
| V3 | compiler | `compiler.go:116-118` | **OpStringBuilderCreate/Append/Build 编译器端无生成路径** — 操作码已定义，VM 已处理，但编译器中无任何代码生成这些操作码 | StringBuilder 优化功能实际未启用 |
| V4 | compiler | `compiler.go:119` | **OpArrayPrealloc 编译器端无生成路径** — 操作码已定义，VM 已处理，但编译器中无代码生成此操作码（从 compileListComprehension 中移除后无替代） | 列表预分配功能实际未启用 |
| V5 | compiler/vm | `compiler.go:75` | **OpYield 死操作码** — 操作码已定义，但编译器从不发射（全部使用 OpYieldValue），VM 也不处理 | 死代码，应清理或移除 |
| V6 | register_vm | `register_vm.go:86,1286` | **RegOpDictUnpack 是 no-op placeholder** | 字典解包在寄存器 VM 中不工作 |
| V7 | register_vm | `register_vm.go:1266` | **slice step "not yet fully implemented"** | 切片步长在寄存器 VM 中不完整 |

### 🟡 JIT 框架未完善实现

| # | 模块 | 位置 | 问题 | 影响 |
|---|------|------|------|------|
| J1 | jit | `enhanced.go:353-354` | **copyPropagation()** — 空函数体 | 优化 pass 未实现 |
| J2 | jit | `enhanced.go:356-357` | **registerAllocation()** — 空函数体 | 优化 pass 未实现 |
| J3 | jit | `enhanced.go:359-360` | **loopOptimizations()** — 空函数体 | 优化 pass 未实现 |
| J4 | jit | `enhanced.go:576-578` | **findTargetFunction()** — 返回 nil | 内联优化无法找到目标函数 |
| J5 | jit | `enhanced.go:208` | **ExecuteFunction()** — 返回 `nil, nil` | JIT 编译后的函数无法实际执行 |
| J6 | jit | `jit.go` | **Compile()** — 返回 nil | JIT 编译为占位实现 |

### 🟡 标准库模块不完整

| # | 模块 | 已实现 | 缺失 |
|---|------|--------|------|
| S1 | functools | reduce, partial | wraps, lru_cache, cached_property, total_ordering, singledispatch, update_wrapper |
| S2 | typing | List, Dict, Tuple, Set, Optional, Union, Any, Callable | TypeVar, Generic, Protocol, Literal, Final, TypeAlias, ParamSpec, Concatenate |
| S3 | hashlib | md5, sha256 | sha1, sha224, sha384, sha512, sha3_224/256/384/512, blake2b, blake2s, shake_128/256 |
| S4 | collections.abc | Iterable, Sequence, Mapping, Set, Callable | Container, Iterator, MutableSequence, ByteString, MutableSet, MutableMapping, MappingView, ItemsView, KeysView, ValuesView, Reversible |
| S5 | itertools | ❌ 完全缺失 | chain, count, cycle, islice, repeat, accumulate, product, permutations, combinations, groupby, starmap, filterfalse, zip_longest, tee, pairwise, batched |
| S6 | asyncio | ❌ 完全缺失 | 事件循环, gather, sleep, create_task, run, Future, Task |
| S7 | sys | getsizeof 占位 | 返回固定值 24，未计算实际对象大小 |

### 🟢 Parser/脱糖层未完善

| # | 模块 | 位置 | 问题 |
|---|------|------|------|
| P1 | parser | `parser.go:1161` | 返回类型注解 "简单实现：跳过直到遇到冒号"，不保留类型信息 |
| P2 | desugar | `desugar.go:1303-1316` | `desugarAsyncForStatement` 保留原样，未实际脱糖 |
| P3 | desugar | `desugar.go:1318-1320` | `desugarAsyncWithStatement` 保留原样，未实际脱糖 |

---

## 路线图

### v0.18 — Code Review Bug 修复 + 稳定性 ✅

> 目标：修复全量 Code Review 发现的所有严重和中等问题，提升运行时稳定性。

#### 🔴 严重 Bug 修复

- [x] **B1**: `executeSetIndex` List 越界错误返回方式修复 — `vm.push(NewIndexError(...))` → `fmt.Errorf(...)`
- [x] **B2**: `OpListUnpack` 中 `unpackExtraArgs` 使用后重置 — 已存在重置逻辑（误报）
- [x] **B3**: `sweepYoung` 晋升年龄检查逻辑修正 — 逻辑正确（误报）
- [x] **B4**: `binaryFloatOp` 浮点除零检查 — 添加 `rightValue == 0` 检查
- [x] **B5**: `StringIO.read`/`BytesIO.read` size=0 处理 — 添加 `size == 0` 快速返回
- [x] **B6**: `BytesIO.seek` 负位置处理 — 已正确处理（误报）

#### 🟡 中等问题修复

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

#### 🟢 代码质量改进

- [x] **L3**: 提取 `augAssignOperator(tokenType) (string, bool)` 消除重复映射
- [x] **L16**: 统一注释语言为英文 — 范围过大，延后到后续版本

### v0.19 — Python 语义完善 ✅

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

### v0.20 — 标准库扩展 ✅

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

### v0.21 — 运行时优化 ✅

> 目标：提升运行时性能，优化热点路径。

- [x] **内联缓存泛化** — OpIndex/OpSetIndex 添加 indexCache，缓存类型分派结果（list+int, tuple+int, dict, range+int, string+int, bytes+int 等 10 种处理器）
- [x] **快速整数算术** — 二元操作（OpAdd/OpSub/OpMul/OpDiv/OpMod/OpFloorDiv/OpPower/OpBitOr/OpBitAnd/OpBitXor）添加 int+int 快速路径，避免函数调用开销；OpDiv 返回 Float 保持 Python 语义
- [x] **字符串构建优化** — StringBuilder 对象 + OpStringBuilderCreate/Append/Build 操作码 + 字符串 += 快速路径（使用 strings.Builder 避免 O(n²) 重复分配）
- [x] **列表预分配** — OpArrayPrealloc 操作码 + VM 处理 + 编译器端检测 range(N) 常量模式并生成预分配指令
- [x] **JIT 热点检测** — VM 集成 JIT 引擎，在 executeCall 中记录 CompiledFunction/Closure 调用次数，超过阈值（默认5次）标记为热点函数；添加 GetJITStats/SetJITHotThreshold/GetJITHotFunctions/ClearJITCache API
- [x] **逃逸分析** — CompiledFunction 添加 NonEscapingLocals 位图，编译器 optimize.go 中实现 analyzeEscape 分析 pass，检测闭包捕获（OpGetFree）、返回值（OpGetLocal+OpReturnValue）、全局赋值（OpSetGlobal）等逃逸模式
- [x] **快速整数算术 bug 修复** — 修复 goto 跳过变量声明的编译错误；修复 OpDiv 快速路径返回 Integer 而非 Float 的 Python 语义错误

### v0.22 — 未完善功能补齐 + 标准库扩展 ✅

> 目标：修复全代码审查发现的未完善实现，补齐缺失的标准库模块。

#### VM/编译器修复

- [x] **V1**: OpAwait 完善实现 — 同步执行 async 帧并缓存结果，OpReturnValue 检测 async 对象并标记 Done
- [x] **V2**: OpInPlaceLShift/OpInPlaceRShift 添加 inPlaceAttrMap 条目 + augAssignToInPlaceOp 映射
- [x] **V5**: OpYield 标记为 Deprecated 死操作码（保留定义避免 iota 值偏移）
- [x] **V6**: RegOpDictUnpack 文档化 — 字典已通过寄存器传递给 RegOpCall，executeCall 自动检测
- [x] **V7**: 寄存器 VM slice step 完善 — 新增 sliceOpWithStep 支持 step 参数（正/负步长，List/String/Bytes/Tuple）
- [x] **V3**: StringBuilder 编译器端生成路径 — WhileStatement 循环体扫描 s += expr 模式，生成 OpStringBuilderCreate/Append/Build
- [x] **V4**: OpArrayPrealloc 编译器端生成路径 — 脱糖 for 循环 range(N) 常量模式检测，生成 OpArrayPrealloc

#### 标准库补齐

- [x] **S5**: itertools 模块 — chain, count, cycle, islice, repeat, accumulate, product, permutations, combinations, groupby, starmap, filterfalse, zip_longest, tee, pairwise（15 个函数，全部惰性迭代器实现）
- [x] **S1**: functools 扩展 — wraps, lru_cache, cached_property, total_ordering, singledispatch, update_wrapper（LRU 缓存使用双向链表实现）
- [x] **S3**: hashlib 扩展 — sha1, sha224, sha384, sha512, sha3_224/256/384/512 + algorithms_available/algorithms_guaranteed
- [x] **S4**: collections.abc 扩展 — Container, Iterator, MutableSequence, ByteString, MutableSet, MutableMapping, MappingView, ItemsView, KeysView, ValuesView, Reversible（11 个新 ABC）
- [x] **S2**: typing 扩展 — TypeVar, Generic, Protocol, Literal, Final, TypeAlias, ParamSpec, Concatenate
- [x] **S7**: sys.getsizeof 真实实现 — 根据对象类型返回 CPython 对齐的内存大小

#### Parser/脱糖层

- [x] **P1**: 返回类型注解保留 — parseExpression(LOWEST) 解析类型表达式，存储到 FunctionLiteral.ReturnType
- [x] **P2/P3**: async for/with 脱糖实现 — async for → while+await __anext__+StopAsyncIteration；async with → await __aenter__/__aexit__

### v0.23 — 寄存器 VM 迁移 Phase 1 + 并发完善

> 目标：补齐寄存器 VM 缺失操作码，引入直接编译器后端，完善 asyncio 生态。

#### 寄存器 VM 迁移 Phase 1：操作码补齐 + 直接编译器后端

**缺失操作码补齐（翻译层 + 执行层）**

- [ ] **R1.1**: 位运算操作码 — RegOpBitOr/RegOpBitAnd/RegOpBitXor + 翻译层 OpBitOr/OpBitAnd/OpBitXor 映射
- [ ] **R1.2**: 集合运算操作码 — RegOpSetUnion/RegOpSetIntersection/RegOpSetDifference/RegOpSetSymmetricDifference + 翻译层映射
- [ ] **R1.3**: 原地操作码 — RegOpInPlaceAdd/Sub/Mul/Div/Mod/FloorDiv/Power/BitOr/BitAnd/BitXor/LShift/RShift + inPlaceAttrMap 支持
- [ ] **R1.4**: OpSetIndex 寄存器化 — RegOpSetIndex + 内联缓存快速路径（List+Int, Dict）
- [ ] **R1.5**: OpSetSlice 寄存器化 — RegOpSetSlice
- [ ] **R1.6**: StringBuilder 操作码 — RegOpStringBuilderCreate/RegOpStringBuilderAppend/RegOpStringBuilderBuild
- [ ] **R1.7**: OpArrayPrealloc 寄存器化 — RegOpArrayPrealloc

**直接编译器后端（绕过栈式字节码翻译）**

- [ ] **R1.8**: 编译器寄存器后端框架 — `RegisterCompiler` 结构体，直接生成 `[]RegInstruction`
- [ ] **R1.9**: 寄存器分配器 — 基于活跃变量分析的线性扫描寄存器分配
- [ ] **R1.10**: 基础表达式编译 — 常量加载、二元运算、比较运算、一元运算的寄存器端编译
- [ ] **R1.11**: 变量存取编译 — GetLocal/SetLocal/GetGlobal/SetGlobal/GetFree 的寄存器端编译
- [ ] **R1.12**: 控制流编译 — if/while 的寄存器端跳转指令生成

#### asyncio 模块

- [ ] **asyncio 事件循环** — 基础事件循环实现
- [ ] **asyncio.gather** — 并发执行多个协程
- [ ] **asyncio.sleep** — 异步休眠
- [ ] **asyncio.create_task** — 创建任务
- [ ] **asyncio.run** — 运行协程入口

### v0.24 — 寄存器 VM 迁移 Phase 2 + JIT 框架

> 目标：寄存器 VM 功能对齐栈式 VM，引入内联缓存和性能优化，实现 JIT 框架核心。

#### 寄存器 VM 迁移 Phase 2：功能对齐 + 性能优化

**内联缓存与性能优化**

- [ ] **R2.1**: 属性访问内联缓存 — RegOpGetAttr 添加 attrCache，缓存类型分派结果
- [ ] **R2.2**: 索引访问内联缓存 — RegOpIndex/RegOpSetIndex 添加 indexCache
- [ ] **R2.3**: 全局变量缓存 — RegOpGetGlobal 已有 globalCache，验证正确性
- [ ] **R2.4**: 快速整数算术 — RegOpAdd/Sub/Mul 等添加 int+int 快速路径，避免函数调用开销
- [ ] **R2.5**: 属性访问完整实现 — getAttrOp 对齐栈式 VM 的完整属性查找链（Instance→Class→MRO→__getattr__）

**高级功能对齐**

- [ ] **R2.6**: 函数调用原生实现 — RegOpCall 不再回退到栈式 executeCall，原生实现参数匹配/默认值/*args/**kwargs
- [ ] **R2.7**: 闭包调用原生实现 — 原生处理闭包参数传递、默认值、可变参数
- [ ] **R2.8**: 生成器/异步原生实现 — RegOpMakeGenerator/RegOpMakeAsync/RegOpYieldValue/RegOpAwait 完整实现
- [ ] **R2.9**: 异常处理完善 — try/except/finally/raise 与栈式 VM 行为完全对齐
- [ ] **R2.10**: 上下文管理器完善 — RegOpEnterContext/RegOpExitContext 对齐栈式 VM

**编译器寄存器后端扩展**

- [ ] **R2.11**: 函数/闭包编译 — 函数定义、闭包创建的寄存器端编译
- [ ] **R2.12**: 数据结构编译 — 列表/字典/集合/元组字面量的寄存器端编译
- [ ] **R2.13**: 属性访问编译 — get/set/del attribute 的寄存器端编译
- [ ] **R2.14**: 类定义编译 — 类创建/继承/元类的寄存器端编译
- [ ] **R2.15**: 异常处理编译 — try/except/finally 的寄存器端编译

#### JIT 框架实现

- [ ] **J1**: copyPropagation 实现 — 复写传播优化 pass
- [ ] **J2**: registerAllocation 实现 — 寄存器分配优化 pass
- [ ] **J3**: loopOptimizations 实现 — 循环优化 pass（循环不变量外提、强度削减）
- [ ] **J4**: findTargetFunction 实现 — 内联优化目标函数查找
- [ ] **J5/J6**: ExecuteFunction/Compile 实现 — JIT 编译后函数的实际执行

### v0.25 — 寄存器 VM 迁移 Phase 3 + 生产就绪

> 目标：寄存器 VM 成为默认执行引擎，栈式 VM 降级为兼容后备，跨平台验证。

#### 寄存器 VM 迁移 Phase 3：默认引擎切换

**编译器端切换**

- [ ] **R3.1**: 寄存器后端覆盖全部 AST 节点 — 所有语句/表达式类型均可直接编译为寄存器指令
- [ ] **R3.2**: 编译器模式选择 — `--vm=register`/`--vm=stack` 命令行参数，默认使用寄存器后端
- [ ] **R3.3**: 字节码序列化格式 — 定义寄存器字节码的二进制序列化格式（.pyc 兼容设计）
- [ ] **R3.4**: 翻译层移除 — 删除 translate() 函数，所有代码路径直接使用寄存器后端

**性能验证与优化**

- [ ] **R3.5**: 性能基准测试 — 寄存器 VM vs 栈式 VM 全面对比，确保无性能回退
- [ ] **R3.6**: 寄存器分配优化 — 图着色寄存器分配替代线性扫描，减少寄存器溢出
- [ ] **R3.7**: 指令调度优化 — 基于依赖分析的指令重排序，提升流水线效率
- [ ] **R3.8**: 栈式 VM 兼容模式 — 保留栈式 VM 作为后备，通过 `--vm=stack` 启用

**跨平台验证**

- [ ] **R3.9**: Linux/AMD64 验证 — 全部测试通过
- [ ] **R3.10**: Linux/ARM64 验证 — 全部测试通过
- [ ] **R3.11**: macOS 验证 — 全部测试通过
- [ ] **R3.12**: Windows 验证 — 全部测试通过
- [ ] **R3.13**: WebAssembly 验证 — 确保寄存器 VM 可在 WASM 环境运行

**异步 I/O 扩展**

- [ ] **异步文件 I/O** — 基于协程的文件操作
- [ ] **异步网络** — 基于协程的 TCP/UDP

### v1.0.0 — Production Ready

- [ ] 所有 v0.18-v0.25 里程碑完成
- [ ] 寄存器 VM 为默认执行引擎，栈式 VM 为兼容后备
- [ ] 性能基准测试达标（CPython 80%+）
- [ ] 测试覆盖率 > 80%
- [ ] 关键路径测试覆盖率 > 95%
- [ ] 跨平台验证（Linux/macOS/Windows/WASM）
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

| 模块 | 当前覆盖 | 目标覆盖 |
|------|----------|----------|
| Lexer | 70% | 90% |
| Parser | 75% | 95% |
| AST | 80% | 95% |
| Desugar | 70% | 90% |
| Compiler | 70% | 90% |
| VM | 65% | 85% |
| Objects | 60% | 85% |
| Collections | 50% | 80% |
| Concurrency | 55% | 80% |
| IO | 40% | 80% |
| Re | 60% | 85% |

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
| `executeSetIndex` 错误处理修复影响现有行为 | 高 | 低 | 回归测试 + 语义验证 |
| DCE 跳转重写不完整导致字节码损坏 | 高 | 中 | 端到端测试 + fuzzing |
| 寄存器 VM 属性访问不一致 | 中 | 中 | 统一 getAttrOp/setAttrOp 逻辑 |
| 并发模块线程安全 | 中 | 中 | 竞态检测 + 压力测试 |
| JIT 优化引入正确性回归 | 高 | 低 | 优化前后结果对比测试 |
| asyncio 实现复杂度 | 高 | 中 | 分阶段实现，对标 CPython 子集 |
| itertools 完全缺失但计划标记已完成 | 高 | 低 | 已修正计划状态，v0.22 补齐 |
| StringBuilder/ArrayPrealloc 编译器端生成路径已完成 | 中 | 低 | v0.22 已实现 WhileStatement 循环模式检测 |
| JIT 框架大量空函数体 | 中 | 高 | v0.24 分阶段实现核心优化 pass |
| 寄存器 VM 迁移 — 翻译层到直接编译器后端切换 | 高 | 中 | 三版本渐进迁移，Phase 1 保留翻译层，Phase 2 双后端并行，Phase 3 切换默认 |
| 寄存器 VM 函数调用原生实现复杂度 | 高 | 中 | Phase 2 逐步从回退模式迁移到原生实现，保留回退路径 |
| 跨平台兼容性 — 寄存器分配器平台差异 | 中 | 低 | 纯 Go 实现，无平台相关代码；WASM 环境需验证寄存器数量限制 |

---

## 开发历史

> v0.3 ~ v0.17.x 的完整开发路径已归档至 [DEVELOPMENT_HISTORY.md](DEVELOPMENT_HISTORY.md)。

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
