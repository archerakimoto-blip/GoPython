# 类系统实现代码

这个文件包含类系统实现的核心代码。

## Parser 修改

在 parseStatement() 中添加：

```go
case lexer.CLASS:
    return p.parseClassStatement()
```

添加新函数：

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

在 parseExpressionStatement() 或 parseExpression() 中处理 `obj.attr` 语法。
