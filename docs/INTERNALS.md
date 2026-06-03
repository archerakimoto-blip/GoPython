# GoPy 内核架构技术文档

本文档描述 GoPy 解释器内核的关键架构决策、数据流和调试要点，供开发和 debug 参考。

---

## 1. 执行管线

```
Source → Lexer → Parser → AST → Desugar → AST' → Compiler → Bytecode → VM
```

- **Lexer**: 源码 → Token 流（缩进敏感，INDENT/DEDENT）
- **Parser**: Token 流 → AST（递归下降解析器）
- **Desugar**: AST → AST'（高级语法转换为内核支持的节点）
- **Compiler**: AST' → Bytecode（常量池 + 指令序列）
- **VM**: 执行 Bytecode（栈式虚拟机）

---

## 2. 栈式 VM 关键机制

### 2.1 栈布局

VM 使用单一操作数栈（`vm.stack[StackSize]`），栈指针 `vm.sp` 指向下一个空位。

**函数调用栈布局**：
```
[... caller frame ...] [callee] [arg0] [arg1] ... [argN] [local0] [local1] ...
                         ^                                ^
                    calleeIndex                      basePointer = calleeIndex + 1
```

**方法调用（自动 Instance 检测）**：
当 `executeCall` 检测到 callee 是 `CompiledFunction` 且 `calleeIndex-1` 位置是 `Instance` 时，自动将 Instance 作为 `self` 插入，`numArgs+1`。

```
调用前: [... instance method arg0 arg1 ...]
调用后: [... method instance arg0 arg1 ...]  (numArgs += 1)
```

**⚠️ 调试要点**：所有从 VM 内部发起的方法调用（property setter/getter、描述符 `__get__`/`__set__`/`__delete__`）必须将 `descInst`/`instance` 放在 callee 下面，让自动检测正确工作。否则在 `__init__` 内调用时，`calleeIndex-1` 可能是 `__init__` 的 `self`，导致参数数量错误。

### 2.2 帧管理

```go
type Frame struct {
    fn          *compiler.CompiledFunction
    ip          int           // 指令指针
    basePointer int           // 栈基址（第一个参数的位置）
    freeVars    []objects.Object  // 闭包的自由变量
    // 异常处理
    initInstance  *objects.Instance  // __init__ 返回值标记
    setAttrValue  objects.Object     // OpSetAttribute 返回值
}
```

- `vm.frames[]`: 帧数组，`vm.framesIndex` 指向下一个空位
- `vm.pushFrame()`: 保存 `vm.sp` 到 `frame.basePointer`，设置 `vm.sp = basePointer + fn.NumLocals`
- `vm.popFrame()`: 恢复 `vm.sp = frame.basePointer - 1`（保留返回值）

### 2.3 闭包与自由变量

**编译期**：
1. `SymbolTable.Resolve(name)` 解析变量时，若变量在外层作用域：
   - `LocalScope/FunctionScope` → 添加到当前 `Free` + `FreeSymbols`，存入 `store`
   - `FreeScope` → 传播到当前 `Free` + `FreeSymbols`，存入 `store`（嵌套闭包关键）
2. 编译函数结束时，读取 `Free`/`FreeSymbols`/`NestedFreeSymbols`，生成 `OpClosure` 指令
3. `OpClosure` 前的指令序列：`OpGetLocal/OpGetFree/OpGetGlobal` 加载自由变量到栈上

**运行期**：
1. `OpClosure` 从栈上弹出自由变量，创建 `Closure{Free: free}` 对象
2. 调用闭包时，`executeCall` 创建 `Frame{freeVars: closure.Free}`
3. `OpGetFree` 从 `frame.freeVars[index]` 读取自由变量

**⚠️ 调试要点**：
- `FreeSymbols` 保存变量在原始作用域的 scope 信息（`LOCAL`/`FREE`/`GLOBAL`），用于生成正确的加载指令
- `Free` 保存变量在当前作用域的 scope 信息（都是 `FREE`），用于 `OpGetFree` 索引
- 编译器必须在退出作用域前保存 `FreeSymbols`，否则 `c.symbolTable` 恢复后读到的是外层作用域的数据
- `Resolve` 必须将结果缓存到 `s.store`，否则同一变量被多次解析时会重复添加到 `Free` 列表

---

## 3. 异常处理

### 3.1 异常栈

```go
type ExceptionHandler struct {
    handlerIP       int           // OpExceptHandler 扫描起始 IP
    stackPtr        int           // 异常发生时的 sp
    exceptCount     int           // except 子句数量
    hasFinally      bool          // 是否有 finally 块
    finallyStartIP  int           // finally 块起始 IP
    finallyEndIP    int           // finally 块结束 IP
    pendingError    objects.Object // 待传播的异常对象
    frameIndex      int           // 异常处理器所在帧的索引
}
```

- `vm.exceptionStack[]`: 异常处理器栈，`OpBeginTry` push，`OpEndTry` pop
- `vm.pendingError`: 跨帧传播的异常对象

### 3.2 异常传播流程

```
OpRaise / raiseException:
  1. 遍历 exceptionStack（从后向前）
  2. 跳过 exceptCount==0 && !hasFinally 的处理器
  3. 回退帧到 handler.frameIndex
  4. 扫描 OpExceptHandler 寻找匹配的 except 子句
  5. 若无匹配但有 finally → 跳转到 finallyStartIP，保存 pendingError
  6. 若无匹配且无 finally → 继续遍历

OpEndTry:
  1. 弹出当前异常处理器
  2. 若有 pendingError → 调用 raiseException() 传播
  3. raiseException 会跨帧查找匹配的处理器

OpFinally:
  1. 设置 finallyStartIP/finallyEndIP
  2. 将 vm.pendingError 转存到 exceptionStack 的 pendingError
```

**⚠️ 调试要点**：
- `try-only-finally`（无 except）的异常传播依赖 `OpEndTry` 中的 `raiseException()` 调用
- `finallyStartIP` 在 `OpBeginTry` 中由编译器预编码，在 `OpFinally` 执行时更新为精确值
- 跨帧异常传播时，`raiseException` 会 `popFrame()` 回退到处理器所在帧

---

## 4. 描述符协议

### 4.1 属性查找优先级

```
OpGetAttribute:
  1. 检查 attrCache（内联缓存）
  2. 检查类属性中的 Property → 调用 fget
  3. 检查类属性中的 ClassMethod → 返回 cm.Fn（绑定类）
  4. 检查类属性中的 StaticMethod → 返回 sm.Fn
  5. 检查类属性中的数据描述符 → 调用 __get__(descInst, instance, classObj)
  6. 检查实例属性 instance.Fields
  7. 检查类属性中的非数据描述符 → 调用 __get__
  8. 检查类属性中的方法 → 返回 CompiledFunction（绑定实例）
```

### 4.2 描述符方法调用栈布局

所有描述符方法调用必须遵循统一的栈布局模式：

```
push(descInst)     // 描述符实例（放在 callee 下面）
push(method)       // __get__/__set__/__delete__ 方法
push(instance)     // 被访问属性的对象
push(args...)      // 额外参数
executeCall(N)     // N = 显式参数数量（不含 self）
```

`executeCall` 的自动 Instance 检测会将 `descInst` 识别为 `self` 并插入，最终调用 `method(descInst, instance, args...)`。

**⚠️ 调试要点**：如果描述符方法调用参数数量错误，检查：
1. `calleeIndex-1` 位置是否是预期的 `descInst`
2. `executeCall` 的 `numArgs` 是否正确（不含自动插入的 self）
3. 是否在 `__init__` 内调用——此时 `calleeIndex-1` 可能是 `__init__` 的 `self`

---

## 5. 编译器符号表

### 5.1 变量作用域

| Scope | 含义 | 指令 |
|-------|------|------|
| `GLOBAL` | 全局变量 | `OpGetGlobal`/`OpSetGlobal` |
| `LOCAL` | 局部变量 | `OpGetLocal`/`OpSetLocal` |
| `FREE` | 自由变量（闭包捕获） | `OpGetFree` |
| `BUILTIN` | 内置函数 | `OpConstant` |
| `FUNCTION` | 函数名（递归引用） | `OpGetGlobal` |

### 5.2 变量索引

```go
symbol.Index = len(s.Free) + s.numDefinitions
```

局部变量索引从 `len(Free)` 开始，确保 free 变量和 local 变量不冲突。

### 5.3 Resolve 流程

```
Resolve(name):
  1. 检查 s.store[name] → 命中则直接返回
  2. 调用 s.outer.Resolve(name)
  3. 根据返回符号的 Scope:
     - LOCAL/FUNCTION: 添加到 s.Free + s.FreeSymbols, 缓存到 s.store
     - FREE: 传播到 s.Free + s.FreeSymbols, 缓存到 s.store
     - GLOBAL/BUILTIN: 直接返回
```

**⚠️ 调试要点**：
- `FreeSymbols` 记录变量在**原始作用域**的 scope（LOCAL/GLOBAL/FREE），用于生成加载指令
- `Free` 记录变量在**当前作用域**的 scope（都是 FREE），用于 `OpGetFree` 索引
- 必须缓存到 `s.store`，否则重复解析导致 `Free` 列表重复
- 嵌套闭包（3层+）需要 `FreeScope` 传播分支

---

## 6. 常见调试模式

### 6.1 参数数量错误 (wrong number of arguments)

**症状**：`wrong number of arguments: want=N, got=M`

**排查步骤**：
1. 确认 `executeCall(numArgs)` 的 `numArgs` 是否正确
2. 检查栈上 `[calleeIndex-1 .. vm.sp-1]` 的实际内容
3. 检查自动 Instance 检测是否意外触发（`calleeIndex-1` 位置是 Instance）
4. 检查 VarArgs/KwArgs 处理是否正确调整了 `numArgs`

### 6.2 闭包自由变量为 nil

**症状**：闭包内访问外层变量得到 nil

**排查步骤**：
1. 检查 `OpClosure` 创建时 `free` 数组是否正确
2. 检查 `OpGetFree` 读取 `frame.freeVars[index]` 是否有值
3. 检查编译器生成的加载指令是否正确（`OpGetLocal` vs `OpGetFree` vs `OpGetGlobal`）
4. 检查 `FreeSymbols` 中的 scope 是否正确（LOCAL → `OpGetLocal`，FREE → `OpGetFree`）

### 6.3 异常不传播

**症状**：try/except 不捕获异常，或 finally 后异常消失

**排查步骤**：
1. 检查 `exceptionStack` 是否有对应的处理器
2. 检查 `handlerIP` 是否正确指向 `OpExceptHandler`
3. 检查 `finallyStartIP` 是否 > 0（-1 表示未设置）
4. 检查 `OpEndTry` 是否正确传播 `pendingError`

### 6.4 属性访问错误

**症状**：属性找不到、类型错误、描述符调用失败

**排查步骤**：
1. 检查 `attrCache` 是否过期（修改属性后应清空缓存）
2. 检查 MRO 查找顺序（`Class.FindClassAttr` 遍历 MRO）
3. 检查描述符优先级（数据描述符 > 实例属性 > 非数据描述符）
4. 检查 `__slots__` 限制（`HasSlots()` + `IsSlotAllowed()`）
