# 类系统简化实现指南

## 最小可行实现

为了在合理时间内完成类系统，我们采用简化版本：

### 支持的语法

```python
# 类定义（只有 __init__ 方法）
class Point:
    __init__(self, x, y):
        self.x = x
        self.y = y

# 实例化
p = Point(1, 2)

# 属性访问
print(p.x)

# 属性赋值
p.x = 10
```

### 不支持的功能（暂时）

- 类继承
- 类变量
- 静态方法
- 类方法
- property 装饰器
- 私有属性

## 实现步骤

### Step 1: 添加 DOT 优先级

在 parser.go 的 precedences 中添加：
```go
lexer.DOT: CALL,
```

### Step 2: 修改 parseStatement

添加 CLASS case：
```go
case lexer.CLASS:
    return p.parseClassStatement()
```

### Step 3: 实现 parseClassStatement

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

### Step 4: 添加 MemberAccess 处理

在 parseExpression 中处理 `.`：
```go
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

### Step 5: Compiler 实现

添加 ClassStatement 编译：
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
```

### Step 6: VM 实现

添加 OpCreateClass：
```go
case compiler.OpCreateClass:
    idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
    class := vm.constants[idx].(*objects.Class)
    return vm.push(class)
```

修改 OpCall 处理 Class：
```go
if classObj, ok := calleeObj.(*objects.Class); ok {
    instance := &objects.Instance{
        Class:  classObj,
        Fields: make(map[string]objects.Object),
    }
    vm.push(instance)
    
    if initMethod, ok := classObj.Methods["__init__"]; ok {
        if fn, ok := initMethod.(*compiler.CompiledFunction); ok {
            frame := NewFrame(fn, instance)
            vm.pushFrame(frame)
        }
    }
    return nil
}
```

添加 OpGetAttribute：
```go
case compiler.OpGetAttribute:
    idx := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
    attrName := vm.constants[idx].(*objects.String).Value
    
    obj := vm.pop()
    if instance, ok := obj.(*objects.Instance); ok {
        if val, ok := instance.GetAttr(attrName); ok {
            return vm.push(val)
        }
        return vm.push(objects.None_)
    }
    return fmt.Errorf("cannot get attribute on non-instance")
```

## 测试用例

```python
class Point:
    __init__(self, x, y):
        self.x = x
        self.y = y

p = Point(1, 2)
print(p.x)
print(p.y)
p.x = 10
print(p.x)
```
