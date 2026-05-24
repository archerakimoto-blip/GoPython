# GoPy 类系统完整实现

## 📋 实现清单

本指南提供了完整实现类系统所需的代码修改。

## ✅ 前提条件（已完成）

- [x] AST 定义（ClassStatement, ClassInstantiation, MemberAccess, MethodCall）
- [x] 对象系统（Class, Instance 类型）
- [x] 关键字支持（CLASS）

## 🔧 需要修改的文件

### 1. pkg/parser/parser.go

#### 修改 1.1: 添加 DOT 优先级
在 `precedences` map 中添加（第 47 行后）：
```go
lexer.DOT:    CALL,
```

#### 修改 1.2: 修改 parseStatement
在 switch 语句中添加（第 165 行附近）：
```go
case lexer.CLASS:
    return p.parseClassStatement()
```

#### 修改 1.3: 添加 parseClassStatement 函数
在文件末尾添加：
```go
func (p *Parser) parseClassStatement() ast.Statement {
    token := p.curToken
    
    if !p.expectPeek(lexer.IDENT) {
        return nil
    }
    name := &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
    
    if !p.expectPeek(lexer.COLON) {
        return nil
    }
    p.nextToken()
    
    if !p.expectPeek(lexer.LBRACE) {
        return nil
    }
    
    body := p.parseBlockStatement()
    
    if !p.expectPeek(lexer.RBRACE) {
        return nil
    }
    
    methods := []*ast.FunctionLiteral{}
    for _, stmt := range body.Statements {
        if es, ok := stmt.(*ast.ExpressionStatement); ok {
            if fl, ok := es.Expression.(*ast.FunctionLiteral); ok {
                methods = append(methods, fl)
            }
        }
    }
    
    return &ast.ClassStatement{
        Token:   token.Literal,
        Name:    name,
        Body:    body,
        Methods: methods,
    }
}
```

#### 修改 1.4: 修改 parseExpression 处理 DOT
找到处理 IDENT 的部分，添加对 DOT 的处理：
```go
// 在 IDENT case 中，default 分支添加：
if p.peekTokenIs(lexer.DOT) {
    p.nextToken()
    p.nextToken()
    member := &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
    left := expr
    for p.peekTokenIs(lexer.DOT) {
        p.nextToken()
        p.nextToken()
        nextMember := &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
        left = &ast.MemberAccess{
            Object: left,
            Member: nextMember,
        }
    }
    return left
}
```

### 2. pkg/compiler/compiler.go

#### 修改 2.1: 添加 OpCreateClass 操作码
找到 OpCodes 定义，添加：
```go
OpCreateClass
```

#### 修改 2.2: 修改 Compile 函数
在 switch 中添加：
```go
case *ast.ClassStatement:
    return c.compileClassStatement(node)
case *ast.MemberAccess:
    return c.compileMemberAccess(node)
```

#### 修改 2.3: 添加 compileClassStatement
```go
func (c *Compiler) compileClassStatement(node *ast.ClassStatement) error {
    class := &objects.Class{
        Name:    node.Name.Value,
        Methods: make(map[string]objects.Object),
        Fields:  make(map[string]objects.Object),
    }
    
    for _, method := range node.Methods {
        compiledFn := c.compileFunction(method)
        class.Methods[method.Name] = compiledFn
    }
    
    c.emit(OpCreateClass, c.addConstant(class))
    return nil
}

func (c *Compiler) compileFunction(fn *ast.FunctionLiteral) *CompiledFunction {
    instructions := c.instructions
    c.instructions = []byte{}
    
    c.enterScope()
    
    for _, param := range fn.Parameters {
        c.symbolTable.Define(param.Value)
    }
    
    for _, stmt := range fn.Body.Statements {
        if err := c.Compile(stmt); err != nil {
            return nil
        }
    }
    
    if c.lastInstruction.opcode != OpReturnValue && c.lastInstruction.opcode != OpReturn {
        c.emit(OpReturnValue)
        c.emit(OpNull)
    }
    
    numLocals := c.symbolTable.numDefinitions
    free := c.symbolTable.Free
    c.exitScope()
    
    c.instructions = append(instructions, c.instructions...)
    
    return &CompiledFunction{
        Instructions: c.instructions,
        NumLocals:    numLocals,
        NumParameters: len(fn.Parameters),
        Free:         free,
    }
}
```

#### 修改 2.4: 添加 compileMemberAccess
```go
func (c *Compiler) compileMemberAccess(node *ast.MemberAccess) error {
    if err := c.Compile(node.Object); err != nil {
        return err
    }
    c.emit(OpGetAttribute, c.addConstant(&objects.String{Value: node.Member.Value}))
    return nil
}
```

### 3. pkg/vm/vm.go

#### 修改 3.1: 添加操作码定义
```go
OpCreateClass
OpGetAttribute
```

#### 修改 3.2: 添加 OpCreateClass 处理
```go
case compiler.OpCreateClass:
    idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
    class := vm.constants[idx].(*objects.Class)
    return vm.push(class)
```

#### 修改 3.3: 修改 OpCall 处理类实例化
找到处理 OpCall 的部分，添加：
```go
if classObj, ok := calleeObj.(*objects.Class); ok {
    instance := &objects.Instance{
        Class:  classObj,
        Fields: make(map[string]objects.Object),
    }
    vm.push(instance)
    
    if initMethod, ok := classObj.Methods["__init__"]; ok {
        if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
            frame := &Frame{
                fn:          fn,
                ip:          -1,
                basePointer:  vm.sp,
            }
            vm.pushFrame(frame)
        }
    }
    return nil
}
```

#### 修改 3.4: 添加 OpGetAttribute 处理
```go
case compiler.OpGetAttribute:
    idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
    attrName := vm.constants[idx].(*objects.String).Value
    
    obj := vm.pop()
    if instance, ok := obj.(*objects.Instance); ok {
        if val, ok := instance.GetAttr(attrName); ok {
            return vm.push(val)
        }
        if method, ok := instance.Class.Methods[attrName]; ok {
            vm.push(instance)
            return vm.push(method)
        }
        return vm.push(objects.None_)
    }
    return fmt.Errorf("cannot get attribute on %s", obj.Type())
```

## 📝 注意事项

1. **__init__ 参数**：__init__ 方法的第一个参数应该是 self
2. **self 参数传递**：调用方法时需要将 instance 作为第一个参数传递
3. **属性查找顺序**：先查找实例属性，再查找类方法
4. **编译顺序**：先编译类定义，再实例化

## 🧪 测试命令

```bash
# 编译
go build -o gopy ./cmd/gopy

# 测试
./gopy tests/test_class_basic.py
./gopy tests/test_class_complete.py
```

## 📚 参考资料

- Python 类系统文档
- GoPy 现有代码结构
- AST 和对象定义

## ⏱️ 预计实现时间

- Parser 修改：1-2 小时
- Compiler 修改：2-3 小时
- VM 修改：2-3 小时
- 测试和调试：2-3 小时

总计：约 7-11 小时
