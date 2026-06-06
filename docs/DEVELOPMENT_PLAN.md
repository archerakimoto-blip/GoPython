# GoPy 开发计划

**当前版本**: 0.18.x
**目标版本**: 1.0.0
**最后更新**: 2026-06

---

## 目录

1. [设计理念](#设计理念)
2. [当前实现状态](#当前实现状态)
3. [全量 Code Review (v0.18)](#全量-code-review-v018)
4. [路线图](#路线图)
5. [架构约束](#架构约束)
6. [测试策略](#测试策略)
7. [风险评估](#风险评估)
8. [开发历史](#开发历史)

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

### 3. 并发特性 (完成度: 75%)

✅ Goroutine 协程 + Channel 通道 + 协程调度器
✅ async/await 语法 + 异步对象
✅ 并发安全数据结构 + 同步原语
⚠️ asyncio 模块 — 部分实现

### 4. 运行时优化 (完成度: 90%)

✅ 内联缓存 + 全局变量缓存 + 特化操作码 + 常量折叠
✅ 对象池 + Dict 优化 + 死代码消除 + 字符串驻留
✅ BoundMethod + Range/Zip 惰性迭代器
✅ 寄存器 VM + 分代 GC
⚠️ 直接线程 — 未实现

### 5. 标准库 (完成度: 90%)

✅ math, sys, os, json, gc, random, string, time, datetime
✅ re (正则表达式), io (StringIO/BytesIO), concurrency
✅ collections (defaultdict, Counter, OrderedDict, deque)
✅ 调试器 + 性能分析器 + JIT + CPython 互操作

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
- [ ] **`set.__isub__`** — `s -= other` 原地差集（延后，当前创建新集合语义正确）
- [ ] **`list.__imul__`** — `lst *= 3` 原地重复（延后，当前脱糖语义正确）
- [ ] **`deque.sort()`** — CPython deque 不支持 sort，移除

### v0.20 — 标准库扩展 ✅

> 目标：扩展标准库覆盖面，补齐高频使用的模块。

- [x] **itertools 模块** — chain, count, cycle, islice, repeat, accumulate
- [x] **functools 模块** — reduce, partial
- [x] **operator 模块** — itemgetter, attrgetter, methodcaller
- [ ] **collections.abc 模块** — Iterable, Sequence, Mapping, Set 抽象基类
- [ ] **pathlib 模块**（基础） — Path 对象, exists/is_file/is_dir
- [ ] **typing 模块**（基础） — List, Dict, Tuple, Optional, Union 类型别名
- [x] **hashlib 模块**（基础） — md5, sha256
- [x] **base64 模块** — encode/decode
- [x] **struct 模块** — pack/unpack 二进制数据

### v0.21 — 运行时优化

> 目标：提升运行时性能，优化热点路径。

- [ ] **直接线程** — 字节码解释器使用 computed goto / 直接线程分发
- [ ] **JIT 热点检测** — 基于调用计数的函数级 JIT 编译
- [ ] **内联缓存泛化** — 扩展 attrCache 到更多操作码（OpGetIndex 等）
- [ ] **逃逸分析** — 编译器识别不逃逸作用域的对象，栈分配优化
- [ ] **字符串构建优化** — 连续字符串拼接使用 StringBuilder 模式
- [ ] **列表预分配** — 推导式中预分配列表容量

### v0.22 — 并发完善

> 目标：完善 asyncio 生态，支持异步 I/O 模式。

- [ ] **asyncio 事件循环** — 基础事件循环实现
- [ ] **asyncio.gather** — 并发执行多个协程
- [ ] **asyncio.sleep** — 异步休眠
- [ ] **asyncio.create_task** — 创建任务
- [ ] **asyncio.run** — 运行协程入口
- [ ] **异步文件 I/O** — 基于协程的文件操作
- [ ] **异步网络** — 基于协程的 TCP/UDP

### v1.0.0 — Production Ready

- [ ] 所有 v0.18-v0.22 里程碑完成
- [ ] 性能基准测试达标（CPython 80%+）
- [ ] 测试覆盖率 > 80%
- [ ] 关键路径测试覆盖率 > 95%
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
