# Changelog

所有重要的更改都会记录在这个文件中。

## [Unreleased]

## [0.16.0] - 2026-06-05

### 新增内置函数 (24 个)

`isinstance()`、`issubclass()`、`hasattr()`、`getattr()`、`setattr()`、`dir()`、`id()`、`hash()`、`callable()`、`enumerate()`、`map()`、`filter()`、`sorted()`、`reversed()`、`repr()`、`iter()`、`any()`、`all()`、`chr()`、`ord()`、`hex()`、`oct()`、`bin()`、`format()`

### 新增字符串方法 (22 个)

`rfind()`、`rindex()`、`count()`、`isdigit()`、`isalpha()`、`isalnum()`、`isspace()`、`isupper()`、`islower()`、`istitle()`、`capitalize()`、`title()`、`swapcase()`、`center()`、`ljust()`、`rjust()`、`zfill()`、`partition()`、`rpartition()`、`encode()`、`isdecimal()`、`isnumeric()`、`isidentifier()`、`isprintable()`、`expandtabs()`

### 新增列表/字典/集合方法

- `list.sort(key=None, reverse=False)` — 支持 key 函数和 reverse 参数
- `dict.fromkeys(iterable[, value])` — 从可迭代对象创建字典
- `set.union()`、`set.intersection()`、`set.difference()`、`set.symmetric_difference()`、`set.issubset()`、`set.issuperset()`、`set.update()`、`set.copy()`

### 新增异常类型 (9 个)

`StopIteration`、`OverflowError`、`FileNotFoundError`、`ImportError`、`SyntaxError`、`IndentationError`、`UnboundLocalError`、`RecursionError`、`MemoryError`

### 其他改进

- 分代 GC `markReferences` 现在追踪 `Instance.SlotValues`
- `RegexMatch` 字段命名规范化（`Groups_` → `Groups`，`Pattern_` → `Pattern`，`OrigString` → `OriginalString`，`GroupIndices` → `GroupStarts`）
- 新增 `IsInstanceOf`/`IsSubclassOf` 辅助函数
- 新增 `BoundMethod` 对象类型

## [0.15.1] - 2026-06-05

### Bug 修复 — Python 语义对齐

- **整数除法返回 Float**：`6/3` 现在正确返回 `2.0`（float），`//` 才是整除
- **取模运算符号修正**：`-7 % 3` 现在正确返回 `2`（与除数同号），而非 `-1`（Go 语义）
- **整除运算修正**：`-7 // 3` 现在正确返回 `-3`（floor 语义），修复了符号判断逻辑
- **整数负数幂返回 Float**：`2 ** -1` 现在正确返回 `0.5`，而非报错
- **List 负索引支持**：`lst[-1]` 现在正确返回最后一个元素
- **越界索引抛出 IndexError**：`lst[100]` 现在抛出 `IndexError`，而非返回 `None`（影响 List/Tuple/String/Range）
- **Boolean 参与算术运算**：`True + 1 = 2`，`False * 3 = 0` 等现在正确工作
- **字符串乘法支持**：`"abc" * 3` 和 `3 * "abc"` 现在正确返回 `"abcabcabc"`
- **列表拼接支持**：`[1] + [2]` 现在正确返回 `[1, 2]`
- **OpAdd 类型检查**：`1 + "a"` 现在正确抛出 `TypeError`，而非隐式拼接为 `"1a"`
- **StopIteration 异常类型**：生成器耗尽时抛出 `StopIteration` 异常，可被 `except StopIteration` 捕获
- **Dict/Set 不可哈希类型拒绝**：`{[]: 1}` 现在正确抛出 `TypeError: unhashable type`
- **异步推导式编译器修复**：生成正确的 for 循环 + 列表构建字节码，filter 跳转正确回填
- **寄存器 VM 跳转映射修复**：使用 `stackIPToRegIP` 映射表正确转换跳转目标
- **寄存器 VM 寄存器泄漏修复**：使用 `registerAllocator` 空闲列表回收不再使用的寄存器

### re 模块修复

- **re.sub/re.subn 支持 count 参数**：`re.sub(pattern, repl, string, count=n)` 限制替换次数
- **re.subn 替换计数准确**：返回实际替换次数而非重新搜索计数
- **re.findall 处理可选组**：可选组未匹配时返回空字符串
- **re.split 保留捕获组分隔符**：`re.split(r'(\W+)', ...)` 现在正确保留分隔符
- **re 模块支持 callable 替换**：`re.sub(pattern, lambda m: m.group(1).upper(), string)` 现在支持 callable 作为 repl
- **re 模块 flags 使用命名常量**：替代魔术数字

### GC 修复

- **分代 GC markObject O(1) 查找**：使用 `objectMap` 替代线性扫描
- **分代 GC 并发安全修复**：`MinorCollect` 不再解锁后调用 `MajorCollect`

## [0.15.0] - 2026-06-05

### 新增特性

- **re 正则表达式模块**：完整支持 Python `re` 模块 API，包括 `re.compile()`、`re.search()`、`re.match()`、`re.fullmatch()`、`re.findall()`、`re.finditer()`、`re.sub()`、`re.subn()`、`re.split()`、`re.escape()`；Pattern 对象支持方法调用和属性访问（`pattern`、`flags`）；Match 对象支持 `group()`、`start()`、`end()`、`span()`、`groups()` 和属性访问（`string`、`re`、`lastindex`）；模块常量 `IGNORECASE`、`MULTILINE`、`DOTALL`、`ASCII`、`UNICODE`
- **Async comprehensions（异步推导式）**：支持 `[x async for x in iter]`、`{x async for x in iter}`、`{k:v async for x in iter}`、`(x async for x in iter)` 语法，包括带 `if` 过滤条件的异步推导式
- **分代 GC（Generational GC）**：将简单标记-清除 GC 升级为分代垃圾回收器，包含 Young Generation（256KB 阈值）和 Old Generation（4MB 阈值）；Minor GC 频繁回收年轻代，Major GC 回收全部代；对象经过 3 次 Minor GC 后晋升到老年代；写屏障（Write Barrier）+ 记忆集（Remembered Set）处理跨代引用；新增 `gc.minor_collect()`、`gc.major_collect()` 等 API
- **寄存器 VM（Register VM）**：新增基于寄存器的虚拟机执行模式，将栈式字节码翻译为寄存器字节码执行，减少内存操作次数；支持常用操作码的翻译和执行；通过 `NewRegisterVM()` 构造器启用；翻译结果缓存避免重复翻译

### 新增对象类型

- `REGEX_PATTERN_OBJ`：RegexPattern 正则表达式编译对象
- `REGEX_MATCH_OBJ`：RegexMatch 正则匹配结果对象

### 新增 AST 节点

- `AsyncListComprehension`：异步列表推导式
- `AsyncSetComprehension`：异步集合推导式
- `AsyncDictComprehension`：异步字典推导式
- `AsyncGeneratorExpression`：异步生成器表达式

### 变更详情

| 模块 | 变更 |
|------|------|
| `pkg/re/module.go` | 新增 `re` 模块，实现 10 个正则表达式函数和 5 个常量 |
| `pkg/objects/object.go` | 新增 `RegexPattern`/`RegexMatch` 结构体和 `GetAttr()` 方法；新增 `REGEX_PATTERN_OBJ`/`REGEX_MATCH_OBJ` 类型 |
| `pkg/vm/vm.go` | `OpGetAttribute` 支持 `RegexPattern`/`RegexMatch` 属性访问；新增 `useRegisterVM`/`regVM` 字段；新增 `NewRegisterVM()` 构造器 |
| `pkg/vm/register_vm.go` | 新增寄存器 VM 实现，包含 40+ 寄存器操作码、翻译器和执行器 |
| `pkg/compiler/compiler.go` | 注册 `re` 模块；新增异步推导式编译支持 |
| `pkg/ast/ast.go` | 新增 `AsyncListComprehension`/`AsyncSetComprehension`/`AsyncDictComprehension`/`AsyncGeneratorExpression` 节点 |
| `pkg/parser/parser.go` | 解析器支持 `async for` 推导式语法 |
| `pkg/desugar/desugar.go` | 脱糖层支持异步推导式节点 |
| `pkg/gc/gc.go` | 分代 GC 实现：Young/Old 双代、Minor/Major 收集、晋升机制、写屏障、记忆集 |

## [0.14.0] - 2026-06-04

### 新增特性

- **Ellipsis (...) 省略号字面量**：支持 `...` 语法和 `Ellipsis` 标识符，两者等价。新增 `OpEllipsis` 操作码和 `EllipsisSingleton` 对象
- **Complex numbers 复数**：支持 `2j`/`3.14j` 纯虚数字面量，`complex(real, imag)` 内置函数，完整的复数算术运算（+、-、*、/、**），`abs()` 支持复数返回模，复数比较（==、!=），取反运算
- **Dictionary Views 字典视图**：`dict.keys()`/`dict.values()`/`dict.items()` 返回动态视图对象（`DictKeys`/`DictValues`/`DictItems`），视图引用原字典，修改字典后视图自动更新，支持 `len()`、索引访问和迭代

### 新增操作码

- `OpEllipsis`：推送 Ellipsis 单例对象到栈

### 新增对象类型

- `ELLIPSIS_OBJ`：Ellipsis 省略号对象
- `COMPLEX_OBJ`：Complex 复数对象（Real + Imag float64）
- `DICT_KEYS_OBJ`/`DICT_VALUES_OBJ`/`DICT_ITEMS_OBJ`：字典视图对象

### 变更详情

| 模块 | 变更 |
|------|------|
| `pkg/lexer/lexer.go` | 新增 `ELLIPSIS`/`COMPLEX` token；`readNumber` 支持 `j`/`J` 后缀；`case '.'` 检测 `...` 三点序列 |
| `pkg/ast/ast.go` | 新增 `EllipsisLiteral`/`ComplexLiteral` AST 节点 |
| `pkg/parser/parser.go` | 注册 `ELLIPSIS`/`COMPLEX` 前缀解析器 |
| `pkg/desugar/desugar.go` | `ComplexLiteral` 模式匹配支持 |
| `pkg/compiler/compiler.go` | `OpEllipsis` 操作码；`ComplexLiteral` 编译为 `OpConstant`；`complex()`/`abs()`/`bool()` 内置函数支持复数；`Ellipsis` 标识符识别 |
| `pkg/vm/vm.go` | `OpEllipsis` 处理；`executeBinaryComplexOperation` 复数算术；复数比较/取反/真值；Dict 视图属性访问和索引 |
| `pkg/objects/object.go` | `Ellipsis` 单例；`Complex` 结构体；`DictKeys`/`DictValues`/`DictItems` 视图；`Equal` 支持 Ellipsis/Complex |

## [0.13.0] - 2026-06-04

### 新增特性

- **Metaclasses（元类）**：支持 `class Foo(metaclass=Meta):` 语法，metaclass 的 `__call__` 控制实例化过程，metaclass 的 `__init__` 在类创建时自动调用（接收 cls, name, bases 参数）
- **仅位置参数 (/)**：支持 Python 风格的仅位置参数分隔符 `/`，如 `def func(a, b, /, c, d)`
- **Exception groups（异常组）**：支持 Python 3.11+ 的 `BaseExceptionGroup`/`ExceptionGroup` 类型和 `except*` 语法，支持异常组分割和多 `except*` 处理器
- **`True`/`False`/`None` 内置常量**：编译器现在正确识别 `True`、`False`、`None` 为内置常量，不再报 "undefined variable" 错误
- **Class 对象属性设置**：`OpSetAttribute` 支持 Class 对象，允许在 metaclass `__init__` 中设置类属性（如 `cls._registered = True`）

### 新增操作码

- `OpSetMetaclass`：设置类的 metaclass 字段
- `OpCallMetaclassInit`：自动调用 metaclass 的 `__init__` 方法
- `OpExceptStarHandler`：`except*` 异常处理器调度和异常组分割

### 修复的问题

- **`compileFunction` 作用域泄漏**：编译函数体出错时未调用 `exitScope()`，导致符号表永久嵌套在子作用域中，后续所有 `Define()` 创建 LOCAL 而非 GLOBAL 符号
- **`True`/`False` 未被编译器识别**：`*ast.Identifier` 解析失败时检查 `True`→`OpTrue`、`False`→`OpFalse`、`None`→`OpNull`
- **`OpSetAttribute` 不支持 Class 对象**：metaclass `__init__` 中 `cls.attr = value` 需要设置类对象属性
- **`NewError` vet 警告**：`NewError(err.Error())` 非常量格式字符串改为 `NewError("%s", err.Error())`

### 变更详情

| 模块 | 变更 |
|------|------|
| `pkg/ast/ast.go` | `ClassStatement` 新增 `Metaclass *Identifier` 字段；`ExceptClause` 新增 `IsStar bool` 字段 |
| `pkg/parser/parser.go` | `parseClassStatement` 支持 `metaclass=XXX` 关键字参数解析；`parseExceptClause` 支持 `except*` 语法；仅位置参数 `/` 解析 |
| `pkg/desugar/desugar.go` | `ClassStatement` 脱糖传播 `Metaclass` 字段 |
| `pkg/compiler/compiler.go` | 新增 `OpSetMetaclass`/`OpCallMetaclassInit`/`OpExceptStarHandler` 操作码；`*ast.Identifier` 识别 `True`/`False`/`None`；`compileFunction` 错误路径修复 `exitScope()` |
| `pkg/compiler/optimize.go` | `InstructionSize` 支持新操作码；DCE 理解 `OpExceptStarHandler` 控制流 |
| `pkg/vm/vm.go` | `Frame` 新增 `metaclassInitClass` 字段；metaclass `__call__` 拦截；`OpSetAttribute` 支持 Class 对象；`OpExceptStarHandler` 异常组分割 |
| `pkg/objects/object.go` | `Class` 新增 `Metaclass *Class` 字段；新增 `ExceptionGroup` 结构体；`NewErrorWithType` 构造函数 |

## [0.12.0] - 2026-06

### 新增特性

- **描述符协议**：支持 `__get__`/`__set__`/`__delete__` 描述符协议，包括数据描述符和非数据描述符的属性查找优先级
- **内置描述符**：`property`（getter/setter/deleter）、`classmethod`、`staticmethod` 内置函数
- **`__slots__`**：支持 `__slots__` 限制实例属性，包括继承场景下的白名单检查
- **属性赋值语法**：支持 `obj.attr = value` 语法（`AttributeAssignStatement` AST 节点）
- **跨帧异常处理**：`raise` 在被调用函数中抛出异常时，能正确回退到调用者的 `try/except` 块捕获
- **try/except 编译器修复**：`OpBeginTry` 新增 `handlerIP` 操作数，直接编码异常处理器位置；DCE 正确保留 `OpExceptHandler` 指令
- **`super()` 内建函数**：支持 `super().__init__(args)` 调用模式
- **`del obj.attr` 完整支持**：`OpDelAttribute` 操作码 + property deleter
- **varargs/kwargs 装饰器包装**：`OpListUnpack`/`OpDictUnpack` + Closure VarArgs/KwArgs 字段
- **嵌套闭包自由变量捕获**：`Resolve` FreeScope 传播 + `FreeSymbols` 保存 + `store` 缓存
- **`__slots__` 内存优化**：Instance 使用 `SlotValues []Object` 固定数组替代 `map[string]Object`
- **range 迭代器死循环修复**：`desugarForToWhile` 改用 `AssignStatement` + 嵌套循环唯一索引变量名

### 修复的问题

- **编译器 `lastInstruction` 状态泄漏**：`compileFunction` 和 `FunctionLiteral` 编译时未重置 `lastInstruction`
- **try/except 穿透问题**：try 块无异常时不再错误地落入 except 块
- **`matchesException` catch-all**：裸 `except:` 现在能捕获非 ERROR_OBJ 类型的异常
- **DCE 删除异常处理器**：死代码消除器现在理解 `OpBeginTry` 的控制流
- **try-only-finally 异常穿透**：`OpEndTry` 改用 `raiseException()` 传播异常
- **自定义描述符 `__set__`/`__get__`/`__delete__` 参数数量错误**：栈布局与 `executeCall` 自动 Instance 检测对齐
- **if 语句解析 bug**：顶层 `if` 语句被 parser 当作表达式解析
- **OpCreateClassWithMultiSuper numParents 读取位置错误**
- **attrCache key 缺少 FrameIndex**：不同函数帧中 IP 相同导致缓存污染
- **字符串比较 bug**：`OpEqual`/`OpNotEqual` 对非数值类型使用 Go 指针比较
- **IfExpression 栈不平衡**：consequence/alternative 块没有值时未 emit `OpNull`
- **OpEndTry 异常对象栈泄漏**

## [0.1.0] - 2026-01-01

### 新增特性

- 基本的算术运算
- 变量绑定
- 函数定义和调用
- if/else 条件语句
- for/while 循环
- 列表、字典、集合
- 基本的字符串和数字处理
- f-string 基本支持
- Lambda 表达式
- 类和对象系统基本支持
