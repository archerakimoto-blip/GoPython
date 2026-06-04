# Changelog

所有重要的更改都会记录在这个文件中。

## [Unreleased]

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
