# GoPy 开发计划

**当前版本**: 0.27.x
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
✅ asyncio 模块 — run/create_task/sleep/gather/Event/Future/wait_for/open/create_subprocess_exec

### 4. 运行时优化 (完成度: 85%)

✅ 内联缓存 + 全局变量缓存 + 特化操作码 + 常量折叠
✅ 对象池 + Dict 优化 + 死代码消除 + 字符串驻留
✅ BoundMethod + Range/Zip 惰性迭代器
✅ 寄存器 VM（翻译层模式）+ 分代 GC
✅ 内联缓存泛化（OpIndex/OpSetIndex）+ 快速整数算术 + StringBuilder + 列表预分配
✅ JIT 热点检测 + 逃逸分析
✅ OpInPlaceLShift/RShift inPlaceAttrMap 条目
✅ 寄存器 VM slice step 支持
✅ StringBuilder 编译器端生成路径 — WhileStatement 循环模式检测
✅ OpArrayPrealloc 编译器端生成路径 — 脱糖 for 循环 range(N) 常量模式检测
✅ JIT 框架 — copyPropagation/registerAllocation/loopOptimizations/findTargetFunction/ExecuteFunction 已实现
⚠️ 直接线程 — 未实现
✅ 寄存器 VM — 直接编译器后端，寄存器帧隔离，全操作码覆盖，内联缓存，原生函数调用，翻译层已移除

### 5. 标准库 (完成度: 75%)

✅ math, sys, os, json, gc, random, string, time, datetime
✅ re (正则表达式), io (StringIO/BytesIO), concurrency
✅ collections (defaultdict, Counter, OrderedDict, deque)
✅ 调试器 + 性能分析器 + JIT + CPython 互操作
✅ functools (reduce, partial, wraps, lru_cache, cached_property, total_ordering, singledispatch) / operator / collections.abc (16个ABC) / pathlib (基础)
✅ typing (List/Dict/Tuple/Set/Optional/Union/Any/Callable/TypeVar/Generic/Protocol/Literal/Final) / hashlib (md5/sha1/sha224/sha256/sha384/sha512/sha3_*)
✅ base64 / struct / itertools (15个函数)
✅ sys.getsizeof 真实实现
✅ asyncio 模块 — run/create_task/sleep/gather/Event/Future/wait_for/open/create_subprocess_exec
⚠️ math — 缺失 atan2, copysign, fmod, frexp, ldexp, modf, isnan, isinf, isfinite, erf, erfc, gamma, lgamma, factorial, gcd, lcm, comb, perm, prod, inf, nan, tau, nextafter, ulp
⚠️ os — 缺失 rmdir, makedirs, removedirs, walk, stat, os.path 子模块 (exists/isfile/isdir/join/split 等)
⚠️ json — 缺失 dump, load, JSONEncoder, JSONDecoder
⚠️ random — 缺失 randrange, sample, choices, gauss, normalvariate 等
⚠️ time — 缺失 strftime, strptime, gmtime, mktime, time_ns, monotonic, perf_counter, struct_time
⚠️ datetime — 返回字符串而非对象，缺失 timedelta, strptime, fromtimestamp, 属性访问
⚠️ sys — 缺失 stdin/stdout/stderr, modules, exc_info, executable, prefix, byteorder, maxsize, flags

### 6. 内置类型方法 (完成度: 55%)

✅ str — rfind, rindex, count, isdigit, isalpha, isalnum, isspace, isupper, islower, istitle, capitalize, title, swapcase, center, ljust, rjust, zfill, partition, rpartition, encode, isdecimal, isnumeric, isidentifier, isprintable, expandtabs, translate, format_map
⚠️ str — 缺失 find, index, replace, split, rsplit, splitlines, join, strip, lstrip, rstrip, upper, lower, startswith, endswith, format, casefold, maketrans
✅ list — sort, __imul__, __iadd__
⚠️ list — 缺失 append, extend, insert, remove, pop, clear, index, count, reverse, copy
✅ dict — fromkeys, update, setdefault, popitem, pop, get, clear, copy, keys, values, items
✅ set — union, intersection, difference, symmetric_difference, issubset, issuperset, update, copy, __isub__, difference_update, __ior__, __iand__, __ixor__
⚠️ set — 缺失 add, remove, discard, pop, clear, intersection_update, symmetric_difference_update
⚠️ tuple — 完全缺失 GetAttr（缺失 count, index）
⚠️ int — 完全缺失 GetAttr（缺失 bit_length, to_bytes, from_bytes）
⚠️ float — 完全缺失 GetAttr（缺失 is_integer, hex, fromhex, as_integer_ratio）
⚠️ bytes — 完全缺失 GetAttr（缺失 decode, hex 等）

### 7. 内置函数 (完成度: 75%)

✅ print, len, range, set, open, next, type, str, int, float, bool, abs, complex, list, min, max, sum, format, input, round, zip, enumerate, property, classmethod, staticmethod, super, ExceptionGroup, isinstance, issubclass, hasattr, getattr, setattr, dir, id, hash, callable, map, filter, sorted, reversed, repr, iter, any, all, chr, ord, hex, oct, bin
⚠️ 缺失 divmod, frozenset, pow(三参数), delattr, globals, locals, vars, eval, exec, compile, bytearray, bytes, memoryview, slice, ascii, object, __import__, breakpoint

---

## 路线图

> v0.3 ~ v0.25 的完整开发路径已归档至 [DEVELOPMENT_HISTORY.md](DEVELOPMENT_HISTORY.md)。

### v0.23 — 寄存器 VM 迁移 Phase 1 + 并发完善 ✅

> 目标：补齐寄存器 VM 缺失操作码，引入直接编译器后端，完善 asyncio 生态。

#### 寄存器 VM 迁移 Phase 1：操作码补齐 + 直接编译器后端

**缺失操作码补齐（翻译层 + 执行层）**

- [x] **R1.1**: 位运算操作码 — RegOpBitOr/RegOpBitAnd/RegOpBitXor + 翻译层 OpBitOr/OpBitAnd/OpBitXor 映射
- [x] **R1.2**: 集合运算操作码 — RegOpSetUnion/RegOpSetIntersection/RegOpSetDifference/RegOpSetSymmetricDifference + 翻译层映射
- [x] **R1.3**: 原地操作码 — RegOpInPlaceAdd/Sub/Mul/Div/Mod/FloorDiv/Power/BitOr/BitAnd/BitXor/LShift/RShift + inPlaceAttrMap 支持
- [x] **R1.4**: OpSetIndex 寄存器化 — RegOpSetIndex + 内联缓存快速路径（List+Int, Dict）
- [x] **R1.5**: OpSetSlice 寄存器化 — RegOpSetSlice
- [x] **R1.6**: StringBuilder 操作码 — RegOpStringBuilderCreate/RegOpStringBuilderAppend/RegOpStringBuilderBuild
- [x] **R1.7**: OpArrayPrealloc 寄存器化 — RegOpArrayPrealloc

**直接编译器后端（绕过栈式字节码翻译）**

- [x] **R1.8**: 编译器寄存器后端框架 — `RegisterCompiler` 结构体，直接生成 `[]RegInstruction`
- [x] **R1.9**: 寄存器分配器 — 基于活跃变量分析的线性扫描寄存器分配
- [x] **R1.10**: 基础表达式编译 — 常量加载、二元运算、比较运算、一元运算的寄存器端编译
- [x] **R1.11**: 变量存取编译 — GetLocal/SetLocal/GetGlobal/SetGlobal/GetFree 的寄存器端编译
- [x] **R1.12**: 控制流编译 — if/while 的寄存器端跳转指令生成

#### asyncio 模块

- [x] **asyncio 事件循环** — 基础事件循环实现
- [x] **asyncio.gather** — 并发执行多个协程
- [x] **asyncio.sleep** — 异步休眠
- [x] **asyncio.create_task** — 创建任务
- [x] **asyncio.run** — 运行协程入口

### v0.24 — 寄存器 VM 迁移 Phase 2 + JIT 框架 ✅

> 目标：寄存器 VM 功能对齐栈式 VM，引入内联缓存和性能优化，实现 JIT 框架核心。

#### 寄存器 VM 迁移 Phase 2：功能对齐 + 性能优化

**内联缓存与性能优化**

- [x] **R2.1**: 属性访问内联缓存 — RegOpGetAttr 添加 attrCache，缓存类型分派结果
- [x] **R2.2**: 索引访问内联缓存 — RegOpIndex/RegOpSetIndex 添加 indexCache
- [x] **R2.3**: 全局变量缓存 — RegOpGetGlobal 已有 globalCache，验证正确性
- [x] **R2.4**: 快速整数算术 — RegOpAdd/Sub/Mul 等添加 int+int 快速路径，避免函数调用开销
- [x] **R2.5**: 属性访问完整实现 — getAttrOp 对齐栈式 VM 的完整属性查找链（Instance→Class→MRO→__getattr__）

**高级功能对齐**

- [x] **R2.6**: 函数调用原生实现 — RegOpCall 不再回退到栈式 executeCall，原生实现参数匹配/默认值/*args/**kwargs
- [x] **R2.7**: 闭包调用原生实现 — 原生处理闭包参数传递、默认值、可变参数
- [x] **R2.8**: 生成器/异步原生实现 — RegOpMakeGenerator/RegOpMakeAsync/RegOpYieldValue/RegOpAwait 完整实现
- [x] **R2.9**: 异常处理完善 — try/except/finally/raise 与栈式 VM 行为完全对齐
- [x] **R2.10**: 上下文管理器完善 — RegOpEnterContext/RegOpExitContext 对齐栈式 VM

**编译器寄存器后端扩展**

- [x] **R2.11**: 函数/闭包编译 — 函数定义、闭包创建的寄存器端编译
- [x] **R2.12**: 数据结构编译 — 列表/字典/集合/元组字面量的寄存器端编译
- [x] **R2.13**: 属性访问编译 — get/set/del attribute 的寄存器端编译
- [x] **R2.14**: 类定义编译 — 类创建/继承/元类的寄存器端编译
- [x] **R2.15**: 异常处理编译 — try/except/finally 的寄存器端编译

#### JIT 框架实现

- [x] **J1**: copyPropagation 实现 — 复写传播优化 pass
- [x] **J2**: registerAllocation 实现 — 寄存器分配优化 pass
- [x] **J3**: loopOptimizations 实现 — 循环优化 pass（循环不变量外提、强度削减）
- [x] **J4**: findTargetFunction 实现 — 内联优化目标函数查找
- [x] **J5/J6**: ExecuteFunction/Compile 实现 — JIT 编译后函数的实际执行

### v0.25 — 寄存器 VM 迁移 Phase 3 + 生产就绪 ✅

> 目标：寄存器 VM 成为默认执行引擎，栈式 VM 降级为兼容后备，跨平台验证。

#### 寄存器 VM 迁移 Phase 3：默认引擎切换

**编译器端切换**

- [x] **R3.1**: 寄存器后端覆盖全部 AST 节点 — 所有语句/表达式类型均可直接编译为寄存器指令
- [x] **R3.2**: 编译器模式选择 — `--vm=register`/`--vm=stack` 命令行参数，默认使用寄存器后端
- [x] **R3.3**: 字节码序列化格式 — 定义寄存器字节码的二进制序列化格式（.pyc 兼容设计）
- [x] **R3.4**: 翻译层移除 — 删除 translate() 函数，所有代码路径直接使用寄存器后端

**性能验证与优化**

- [x] **R3.5**: 性能基准测试 — 寄存器 VM vs 栈式 VM 全面对比，确保无性能回退
- [x] **R3.6**: 寄存器分配优化 — 图着色寄存器分配替代线性扫描，减少寄存器溢出
- [x] **R3.7**: 指令调度优化 — 基于依赖分析的指令重排序，提升流水线效率
- [x] **R3.8**: 栈式 VM 兼容模式 — 保留栈式 VM 作为后备，通过 `--vm=stack` 启用

**跨平台验证**

- [x] **R3.9**: Linux/AMD64 验证 — 全部测试通过
- [x] **R3.10**: Linux/ARM64 验证 — 全部测试通过
- [x] **R3.11**: macOS 验证 — 全部测试通过
- [x] **R3.12**: Windows 验证 — 全部测试通过
- [x] **R3.13**: WebAssembly 验证 — 确保寄存器 VM 可在 WASM 环境运行

**异步 I/O 扩展**

- [x] **异步文件 I/O** — 基于协程的文件操作
- [x] **异步网络** — 基于协程的 TCP/UDP

### v0.28 — 内置方法与标准库补齐

> 目标：补齐排查出的内置函数、对象方法和标准库模块缺失，提升 CPython 兼容性。

#### B1. 字符串方法补齐（优先级：高）

- [ ] **B1.1**: `str.find()` — 查找子串位置，支持 start/end 参数
- [ ] **B1.2**: `str.index()` — 查找子串位置，不存在抛 ValueError
- [ ] **B1.3**: `str.replace()` — 子串替换，支持 count 参数
- [ ] **B1.4**: `str.split()` — 字符串分割，支持 sep/maxsplit 参数
- [ ] **B1.5**: `str.rsplit()` — 右分割
- [ ] **B1.6**: `str.splitlines()` — 按行分割
- [ ] **B1.7**: `str.join()` — 连接序列
- [ ] **B1.8**: `str.strip()` / `str.lstrip()` / `str.rstrip()` — 去除首尾字符
- [ ] **B1.9**: `str.upper()` / `str.lower()` — 大小写转换
- [ ] **B1.10**: `str.startswith()` / `str.endswith()` — 前缀/后缀判断
- [ ] **B1.11**: `str.format()` — 格式化字符串
- [ ] **B1.12**: `str.casefold()` — 大小写折叠
- [ ] **B1.13**: `str.maketrans()` — 创建转换表（实例方法版）

#### B2. 列表方法补齐（优先级：高）

- [ ] **B2.1**: `list.append()` — 追加元素
- [ ] **B2.2**: `list.extend()` — 扩展列表
- [ ] **B2.3**: `list.insert()` — 插入元素
- [ ] **B2.4**: `list.remove()` — 删除首个匹配元素
- [ ] **B2.5**: `list.pop()` — 弹出指定位置元素
- [ ] **B2.6**: `list.clear()` — 清空列表
- [ ] **B2.7**: `list.index()` — 查找元素索引
- [ ] **B2.8**: `list.count()` — 统计元素出现次数
- [ ] **B2.9**: `list.reverse()` — 反转列表
- [ ] **B2.10**: `list.copy()` — 浅拷贝

#### B3. 集合方法补齐（优先级：中）

- [ ] **B3.1**: `set.add()` — 添加元素
- [ ] **B3.2**: `set.remove()` — 删除元素（不存在抛 KeyError）
- [ ] **B3.3**: `set.discard()` — 安全删除元素
- [ ] **B3.4**: `set.pop()` — 弹出元素
- [ ] **B3.5**: `set.clear()` — 清空集合
- [ ] **B3.6**: `set.intersection_update()` — 交集更新
- [ ] **B3.7**: `set.symmetric_difference_update()` — 对称差更新

#### B4. 元组/整数/浮点/字节方法补齐（优先级：中）

- [ ] **B4.1**: `tuple.count()` / `tuple.index()` — 元组方法
- [ ] **B4.2**: `int.bit_length()` / `int.to_bytes()` / `int.from_bytes()` — 整数方法
- [ ] **B4.3**: `float.is_integer()` / `float.hex()` / `float.fromhex()` / `float.as_integer_ratio()` — 浮点方法
- [ ] **B4.4**: `bytes.decode()` / `bytes.hex()` — 字节方法

#### B5. 缺失内置函数补齐（优先级：高）

- [ ] **B5.1**: `divmod()` — 商和余数
- [ ] **B5.2**: `frozenset()` — 冻结集合构造
- [ ] **B5.3**: `pow()` 三参数版 — 幂运算 + 取模
- [ ] **B5.4**: `delattr()` — 删除属性
- [ ] **B5.5**: `globals()` / `locals()` — 返回变量字典
- [ ] **B5.6**: `vars()` — 返回对象 __dict__
- [ ] **B5.7**: `eval()` / `exec()` — 动态执行代码
- [ ] **B5.8**: `compile()` — 编译源代码
- [ ] **B5.9**: `bytearray()` — 字节数组构造
- [ ] **B5.10**: `bytes()` — 字节构造
- [ ] **B5.11**: `memoryview()` — 内存视图
- [ ] **B5.12**: `slice()` — 切片对象构造
- [ ] **B5.13**: `ascii()` — 返回 ASCII 表示
- [ ] **B5.14**: `object()` — 基础对象构造
- [ ] **B5.15**: `__import__()` — 模块导入函数
- [ ] **B5.16**: `breakpoint()` — 调试断点

#### B6. 缺失异常类型补齐（优先级：中）

- [ ] **B6.1**: `Exception` / `BaseException` — 基础异常类
- [ ] **B6.2**: `AssertionError` — 断言错误
- [ ] **B6.3**: `OSError` / `IOError` / `FileNotFoundError` / `FileExistsError` / `PermissionError` — 系统异常族
- [ ] **B6.4**: `ImportError` / `ModuleNotFoundError` — 导入异常
- [ ] **B6.5**: `SyntaxError` / `IndentationError` — 语法异常
- [ ] **B6.6**: `UnicodeError` / `UnicodeDecodeError` / `UnicodeEncodeError` — Unicode 异常
- [ ] **B6.7**: `RecursionError` — 递归错误
- [ ] **B6.8**: `IsADirectoryError` / `NotADirectoryError` — 目录异常

#### B7. 标准库模块补齐（优先级：中）

- [ ] **B7.1**: math — `atan2, copysign, fmod, frexp, ldexp, modf, isnan, isinf, isfinite, erf, erfc, gamma, lgamma, factorial, gcd, lcm, comb, perm, prod, inf, nan, tau, nextafter, ulp`
- [ ] **B7.2**: os — `rmdir, makedirs, removedirs, walk, stat, os.name, os.linesep, os.curdir, os.pardir`；os.path 子模块 — `exists, isfile, isdir, join, split, basename, dirname, getsize, abspath, realpath, expanduser`
- [ ] **B7.3**: json — `dump, load, JSONEncoder, JSONDecoder`
- [ ] **B7.4**: random — `randrange, sample, choices, gauss, normalvariate, lognormvariate, expovariate, triangular`
- [ ] **B7.5**: time — `strftime, strptime, gmtime, mktime, time_ns, monotonic, perf_counter, process_time, timezone, tzname, struct_time`
- [ ] **B7.6**: datetime — 重构为对象返回（非字符串），补齐 `timedelta, strptime, fromtimestamp, year/month/day 等属性访问`
- [ ] **B7.7**: sys — `stdin, stdout, stderr, modules, exc_info, executable, prefix, byteorder, maxsize, flags`

#### B8. 缺失类型补齐（优先级：低）

- [ ] **B8.1**: `bytearray` 类型 — 可变字节数组
- [ ] **B8.2**: `frozenset` 类型 — 不可变集合
- [ ] **B8.3**: `memoryview` 类型 — 内存视图
- [ ] **B8.4**: `slice` 对象 — 切片对象类型

### v1.0.0 — Production Ready

- [x] 所有 v0.18-v0.25 里程碑完成
- [x] 寄存器 VM 为默认执行引擎，栈式 VM 为兼容后备
- [x] 性能基准测试达标（CPython 80%+）
- [ ] 测试覆盖率 > 80%（当前：Lexer 97.9%, Desugar 99.6%, Parser 81.7%, Compiler 78.3%, VM 52.4%）
- [x] 关键路径测试覆盖率 > 95%（Lexer, Desugar 已达标）
- [x] 跨平台验证（Linux/macOS/Windows/WASM）
- [ ] v0.28 内置方法与标准库补齐
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
