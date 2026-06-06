package desugar

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
)

// Desugar 对整个程序进行脱糖转换
func Desugar(program *ast.Program) *ast.Program {
	desugared := &ast.Program{
		Statements: make([]ast.Statement, 0, len(program.Statements)),
	}

	for _, stmt := range program.Statements {
		desugaredStmt := desugarStatement(stmt)
		if desugaredStmt != nil {
			desugared.Statements = append(desugared.Statements, desugaredStmt)
		}
	}

	return desugared
}

// desugarStatement 脱糖单个语句
func desugarStatement(stmt ast.Statement) ast.Statement {
	if stmt == nil {
		return nil
	}
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		if s == nil || s.Expression == nil {
			return nil
		}
		// 检查是否是带装饰器的函数
		if fnLit, ok := s.Expression.(*ast.FunctionLiteral); ok && len(fnLit.Decorators) > 0 {
			desugaredFn := desugarExpression(fnLit).(*ast.FunctionLiteral)
			desugaredDecorators := make([]ast.Expression, len(desugaredFn.Decorators))
			for i, dec := range desugaredFn.Decorators {
				desugaredDecorators[i] = desugarExpression(dec)
			}

			for _, dec := range desugaredDecorators {
				if isLruCacheDecorator(dec) {
					return desugarLruCache(desugaredFn)
				}
			}

			stmts := []ast.Statement{}

			funcIdent := &ast.Identifier{Token: desugaredFn.Token, Value: desugaredFn.Name}
			tempIdent := &ast.Identifier{Token: desugaredFn.Token, Value: "_temp_" + desugaredFn.Name}

			tempFn := &ast.FunctionLiteral{
				Token:       desugaredFn.Token,
				Name:        "",
				Parameters:  desugaredFn.Parameters,
				Defaults:    desugaredFn.Defaults,
				KeywordOnly: desugaredFn.KeywordOnly,
				Body:        desugaredFn.Body,
				VarArgs:     desugaredFn.VarArgs,
				KwArgs:      desugaredFn.KwArgs,
			}

			letStmt := &ast.LetStatement{
				Token: desugaredFn.Token,
				Names: []*ast.Identifier{tempIdent},
				Value: tempFn,
			}
			stmts = append(stmts, letStmt)

			currentValue := tempIdent
			for i := len(desugaredDecorators) - 1; i >= 0; i-- {
				decorator := desugaredDecorators[i]
				callExpr := &ast.CallExpression{
					Token:     decorator.TokenLiteral(),
					Function:  decorator,
					Arguments: []ast.Expression{currentValue},
				}
				if i == 0 {
					assignStmt := &ast.AssignStatement{
						Token: desugaredFn.Token,
						Names: []*ast.Identifier{funcIdent},
						Value: callExpr,
					}
					stmts = append(stmts, assignStmt)
				} else {
					letStmt = &ast.LetStatement{
						Token: desugaredFn.Token,
						Names: []*ast.Identifier{tempIdent},
						Value: callExpr,
					}
					stmts = append(stmts, letStmt)
				}
			}

			if len(desugaredDecorators) == 1 {
				decorator := desugaredDecorators[0]
				callExpr := &ast.CallExpression{
					Token:     decorator.TokenLiteral(),
					Function:  decorator,
					Arguments: []ast.Expression{tempFn},
				}
				stmts = []ast.Statement{
					&ast.LetStatement{
						Token: desugaredFn.Token,
						Names: []*ast.Identifier{funcIdent},
						Value: callExpr,
					},
				}
			}

			return &ast.BlockStatement{
				Token:      s.Token,
				Statements: stmts,
			}
		}

		// 普通表达式语句处理
		desugaredExpr := desugarExpression(s.Expression)
		if desugaredExpr == nil {
			return nil
		}
		return &ast.ExpressionStatement{
			Token:      s.Token,
			Expression: desugaredExpr,
		}
	case *ast.LetStatement:
		if len(s.Names) > 1 {
			// 多重赋值：let a, b = x, y
			// 脱糖为 let _temp = x; let a = _temp[0]; let b = _temp[1];
			tempIdent := &ast.Identifier{Token: "_temp", Value: "_temp"}
			stmts := []ast.Statement{
				&ast.LetStatement{
					Token: s.Token,
					Names: []*ast.Identifier{tempIdent},
					Value: desugarExpression(s.Value),
				},
			}

			for i, name := range s.Names {
				indexExpr := &ast.IndexExpression{
					Token: "[",
					Left:  tempIdent,
					Index: &ast.IntegerLiteral{Token: string(rune('0' + i)), Value: int64(i)},
				}
				stmts = append(stmts, &ast.LetStatement{
					Token: s.Token,
					Names: []*ast.Identifier{name},
					Value: indexExpr,
				})
			}

			return &ast.BlockStatement{
				Token:      s.Token,
				Statements: stmts,
			}
		}

		return &ast.LetStatement{
			Token: s.Token,
			Names: s.Names,
			Value: desugarExpression(s.Value),
		}
	case *ast.AssignStatement:
		if len(s.Names) == 1 {
			if call, ok := s.Value.(*ast.CallExpression); ok {
				if ident, ok := call.Function.(*ast.Identifier); ok {
					if ident.Value == "NamedTuple" && len(call.Arguments) >= 2 {
						return desugarNamedTuple(s.Names[0].Value, call)
					}
				}
			}
		}
		if len(s.Names) > 1 {
			// 多重赋值：a, b = x, y
			// 脱糖为 let _temp = x; a = _temp[0]; b = _temp[1];
			tempIdent := &ast.Identifier{Token: "_temp", Value: "_temp"}
			stmts := []ast.Statement{
				&ast.LetStatement{
					Token: s.Token,
					Names: []*ast.Identifier{tempIdent},
					Value: desugarExpression(s.Value),
				},
			}

			for i, name := range s.Names {
				indexExpr := &ast.IndexExpression{
					Token: "[",
					Left:  tempIdent,
					Index: &ast.IntegerLiteral{Token: string(rune('0' + i)), Value: int64(i)},
				}
				stmts = append(stmts, &ast.AssignStatement{
					Token: s.Token,
					Names: []*ast.Identifier{name},
					Value: indexExpr,
				})
			}

			return &ast.BlockStatement{
				Token:      s.Token,
				Statements: stmts,
			}
		}

		// 单变量赋值，保持不变
		return &ast.AssignStatement{
			Token: s.Token,
			Names: s.Names,
			Value: desugarExpression(s.Value),
		}
	case *ast.AugAssignStatement:
		if s.IndexLeft != nil {
			// 索引增强赋值：d['a'] += 1 -> d.__setitem__('a', d.__getitem__('a') + 1)
			// 脱糖为 IndexAssignStatement: d['a'] = d['a'] + 1
			indexExpr := &ast.IndexExpression{
				Token: s.Token,
				Left:  s.IndexLeft,
				Index: s.IndexIndex,
			}
			infixExpr := &ast.InfixExpression{
				Token:    s.Token,
				Left:     indexExpr,
				Operator: s.Operator,
				Right:    desugarExpression(s.Value),
			}
			return &ast.IndexAssignStatement{
				Token: s.Token,
				Left:  s.IndexLeft,
				Index: s.IndexIndex,
				Value: infixExpr,
			}
		}
		// 将增强赋值转换为: name = name op value
		leftIdent := &ast.Identifier{Token: s.Name.Token, Value: s.Name.Value}
		infixExpr := &ast.InfixExpression{
			Token:    s.Token,
			Left:     leftIdent,
			Operator: s.Operator,
			Right:    desugarExpression(s.Value),
		}
		return &ast.AssignStatement{
			Token: s.Token,
			Names: []*ast.Identifier{s.Name},
			Value: infixExpr,
		}
	case *ast.IndexAssignStatement:
		return &ast.IndexAssignStatement{
			Token: s.Token,
			Left:  desugarExpression(s.Left),
			Index: desugarExpression(s.Index),
			Value: desugarExpression(s.Value),
		}
	case *ast.ReturnStatement:
		return &ast.ReturnStatement{
			Token:       s.Token,
			ReturnValue: desugarExpression(s.ReturnValue),
		}
	case *ast.BlockStatement:
		return desugarBlockStatement(s)
	case *ast.WhileStatement:
		return &ast.WhileStatement{
			Token:     s.Token,
			Condition: desugarExpression(s.Condition),
			Body:      desugarBlockStatement(s.Body),
		}
	case *ast.ForStatement:
		return desugarForToWhile(s)
	case *ast.BreakStatement:
		return nil
	case *ast.ContinueStatement:
		return nil
	case *ast.TryStatement:
		// 对 try 语句进行脱糖处理：脱糖 body、excepts 和 finally
		desugaredTry := &ast.TryStatement{
			Token: s.Token,
			Body:  desugarBlockStatement(s.Body),
		}
		desugaredTry.Excepts = make([]*ast.ExceptClause, 0, len(s.Excepts))
		for _, ex := range s.Excepts {
			desugaredExcept := &ast.ExceptClause{
				Token:  ex.Token,
				Type:   desugarExpression(ex.Type),
				Name:   ex.Name,
				Body:   desugarBlockStatement(ex.Body),
				IsStar: ex.IsStar,
			}
			desugaredTry.Excepts = append(desugaredTry.Excepts, desugaredExcept)
		}
		if s.Finally != nil {
			desugaredTry.Finally = desugarBlockStatement(s.Finally)
		}
		return desugaredTry
	case *ast.RaiseStatement:
		// 对 raise 语句进行脱糖处理：脱糖表达式
		return &ast.RaiseStatement{
			Token:      s.Token,
			Expression: desugarExpression(s.Expression),
		}
	case *ast.WithStatement:
		// 对 with 语句进行脱糖处理：多重上下文管理器脱糖为嵌套with语句
		desugaredItems := make([]*ast.ContextManagerItem, 0, len(s.Items))
		for _, item := range s.Items {
			desugaredItems = append(desugaredItems, &ast.ContextManagerItem{
				Expr: desugarExpression(item.Expr),
				Name: item.Name,
			})
		}

		// 如果只有1个上下文管理器，直接处理
		if len(desugaredItems) == 1 {
			return &ast.WithStatement{
				Token: s.Token,
				Items: desugaredItems,
				Body:  desugarBlockStatement(s.Body),
			}
		}

		// 多重上下文管理器：从最后一个开始，嵌套到前一个的Body里
		var nestedStatement ast.Statement = &ast.WithStatement{
			Token: s.Token,
			Items: []*ast.ContextManagerItem{desugaredItems[len(desugaredItems)-1]},
			Body:  desugarBlockStatement(s.Body),
		}

		for i := len(desugaredItems) - 2; i >= 0; i-- {
			nestedStatement = &ast.WithStatement{
				Token: s.Token,
				Items: []*ast.ContextManagerItem{desugaredItems[i]},
				Body: &ast.BlockStatement{
					Token:      s.Token,
					Statements: []ast.Statement{nestedStatement},
				},
			}
		}

		return nestedStatement
	case *ast.YieldStatement:
		// 对 yield 语句进行脱糖处理：脱糖表达式
		return &ast.YieldStatement{
			Token:      s.Token,
			Expression: desugarExpression(s.Expression),
		}
	case *ast.GlobalStatement:
		// global 语句本身不需要脱糖，只是声明
		// 实际的符号表处理会在 compiler 阶段通过 AST 节点信息处理
		return s
	case *ast.NonlocalStatement:
		// nonlocal 语句本身不需要脱糖，只是声明
		// 实际的符号表处理会在 compiler 阶段通过 AST 节点信息处理
		return s
	case *ast.DeleteStatement:
		return desugarDeleteStatement(s)
	case *ast.YieldFromStatement:
		return desugarYieldFromStatement(s)
	case *ast.AsyncForStatement:
		return desugarAsyncForStatement(s)
	case *ast.AsyncWithStatement:
		return desugarAsyncWithStatement(s)
	case *ast.ClassStatement:
		if isEnumClass(s) {
			return desugarEnum(s)
		}
		desugaredClass := &ast.ClassStatement{
			Token:        s.Token,
			Name:         s.Name,
			SuperClass:   s.SuperClass,
			SuperClasses: s.SuperClasses,
			Metaclass:    s.Metaclass,
			Body:         desugarBlockStatement(s.Body),
			Methods:      s.Methods,
		}
		return desugaredClass
	case *ast.MatchStatement:
		return desugarMatchStatement(s)
	default:
		return stmt
	}
}

func isEnumClass(s *ast.ClassStatement) bool {
	if s.SuperClass != nil && s.SuperClass.Value == "Enum" {
		return true
	}
	for _, sc := range s.SuperClasses {
		if sc.Value == "Enum" {
			return true
		}
	}
	return false
}

func desugarEnum(s *ast.ClassStatement) ast.Statement {
	enumName := s.Name.Value
	members := make(map[string]ast.Expression)
	autoCounter := int64(1)

	if s.Body != nil {
		for _, stmt := range s.Body.Statements {
			if assign, ok := stmt.(*ast.AssignStatement); ok && len(assign.Names) == 1 {
				fieldName := assign.Names[0].Value
				if assign.Value != nil {
					members[fieldName] = desugarExpression(assign.Value)
				} else {
					members[fieldName] = &ast.IntegerLiteral{Token: string(rune('0' + autoCounter)), Value: autoCounter}
					autoCounter++
				}
			}
			if let, ok := stmt.(*ast.LetStatement); ok && len(let.Names) == 1 {
				fieldName := let.Names[0].Value
				if let.Value != nil {
					members[fieldName] = desugarExpression(let.Value)
				} else {
					members[fieldName] = &ast.IntegerLiteral{Token: string(rune('0' + autoCounter)), Value: autoCounter}
					autoCounter++
				}
			}
		}
	}

	dictPairs := make(map[ast.Expression]ast.Expression)
	for k, v := range members {
		dictPairs[&ast.StringLiteral{Token: k, Value: k}] = v
	}

	enumCall := &ast.CallExpression{
		Token:    "__enum__",
		Function: &ast.Identifier{Token: "__enum__", Value: "__enum__"},
		Arguments: []ast.Expression{
			&ast.StringLiteral{Token: enumName, Value: enumName},
			&ast.HashLiteral{Token: "{", Pairs: dictPairs},
		},
	}

	return &ast.AssignStatement{
		Token: "=",
		Names: []*ast.Identifier{{Token: enumName, Value: enumName}},
		Value: enumCall,
	}
}

func desugarNamedTuple(name string, call *ast.CallExpression) ast.Statement {
	var fields []struct {
		Name string
		Type string
	}

	if list, ok := call.Arguments[1].(*ast.ListLiteral); ok {
		for _, elem := range list.Elements {
			if fieldList, ok := elem.(*ast.ListLiteral); ok && len(fieldList.Elements) >= 1 {
				fieldName := ""
				fieldType := ""
				if ident, ok := fieldList.Elements[0].(*ast.Identifier); ok {
					fieldName = ident.Value
				}
				if len(fieldList.Elements) >= 2 {
					if ident, ok := fieldList.Elements[1].(*ast.Identifier); ok {
						fieldType = ident.Value
					}
				}
				if fieldName != "" {
					fields = append(fields, struct {
						Name string
						Type string
					}{Name: fieldName, Type: fieldType})
				}
			}
			if str, ok := elem.(*ast.StringLiteral); ok {
				fields = append(fields, struct {
					Name string
					Type string
				}{Name: str.Value, Type: ""})
			}
		}
	}

	params := make([]*ast.Identifier, len(fields))
	initBody := make([]ast.Statement, len(fields))
	reprParts := make([]ast.Expression, 0, len(fields)*2+1)

	reprParts = append(reprParts, &ast.StringLiteral{Token: name + "(", Value: name + "("})

	for i, f := range fields {
		params[i] = &ast.Identifier{Token: f.Name, Value: f.Name}

		initBody[i] = &ast.ExpressionStatement{
			Token: "=",
			Expression: &ast.InfixExpression{
				Token:    "=",
				Left:     &ast.MemberAccess{Token: ".", Object: &ast.Identifier{Token: "self", Value: "self"}, Member: &ast.Identifier{Token: f.Name, Value: f.Name}},
				Operator: "=",
				Right:    &ast.Identifier{Token: f.Name, Value: f.Name},
			},
		}

		if i > 0 {
			reprParts = append(reprParts, &ast.StringLiteral{Token: ", ", Value: ", "})
		}
		reprParts = append(reprParts, &ast.StringLiteral{Token: f.Name + "=", Value: f.Name + "="})
		reprParts = append(reprParts, &ast.CallExpression{
			Token:     "str",
			Function:  &ast.Identifier{Token: "str", Value: "str"},
			Arguments: []ast.Expression{&ast.MemberAccess{Token: ".", Object: &ast.Identifier{Token: "self", Value: "self"}, Member: &ast.Identifier{Token: f.Name, Value: f.Name}}},
		})
	}

	reprParts = append(reprParts, &ast.StringLiteral{Token: ")", Value: ")"})

	initFn := &ast.FunctionLiteral{
		Token:      "def",
		Name:       "__init__",
		Parameters: append([]*ast.Identifier{{Token: "self", Value: "self"}}, params...),
		Body:       &ast.BlockStatement{Token: ":", Statements: initBody},
	}

	reprConcat := reprParts[0]
	for i := 1; i < len(reprParts); i++ {
		reprConcat = &ast.InfixExpression{
			Token:    "+",
			Left:     reprConcat,
			Operator: "+",
			Right:    reprParts[i],
		}
	}

	reprFn := &ast.FunctionLiteral{
		Token:      "def",
		Name:       "__repr__",
		Parameters: []*ast.Identifier{{Token: "self", Value: "self"}},
		Body: &ast.BlockStatement{Token: ":", Statements: []ast.Statement{
			&ast.ReturnStatement{Token: "return", ReturnValue: reprConcat},
		}},
	}

	classStmt := &ast.ClassStatement{
		Token:   "class",
		Name:    &ast.Identifier{Token: name, Value: name},
		Body:    &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Methods: []*ast.FunctionLiteral{initFn, reprFn},
	}

	return &ast.AssignStatement{
		Token: "=",
		Names: []*ast.Identifier{{Token: name, Value: name}},
		Value: classStmt,
	}
}

// desugarBlockStatement 脱糖块语句
func desugarBlockStatement(block *ast.BlockStatement) *ast.BlockStatement {
	if block == nil {
		return nil
	}
	desugared := &ast.BlockStatement{
		Token:      block.Token,
		Statements: make([]ast.Statement, 0, len(block.Statements)),
	}

	for _, stmt := range block.Statements {
		desugaredStmt := desugarStatement(stmt)
		if desugaredStmt != nil {
			desugared.Statements = append(desugared.Statements, desugaredStmt)
		}
	}

	return desugared
}

// desugarExpression 脱糖单个表达式
func desugarExpression(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.PrefixExpression:
		return &ast.PrefixExpression{
			Token:    e.Token,
			Operator: e.Operator,
			Right:    desugarExpression(e.Right),
		}
	case *ast.AwaitExpression:
		return &ast.AwaitExpression{
			Token: e.Token,
			Value: desugarExpression(e.Value),
		}
	case *ast.NamedExpression:
		// Walrus 运算符脱糖：x := expr 脱糖为 (x = expr; x)
		// 创建一个临时变量来存储赋值和返回值
		value := desugarExpression(e.Value)

		// 简化处理：在赋值的同时返回右值
		// 实际上，这需要编译器特殊处理，我们先把赋值语句添加到外层
		// 这里我们返回一个特殊的结构，编译器会识别并处理
		return &ast.NamedExpression{
			Token: e.Token,
			Name:  e.Name,
			Value: value,
		}
	case *ast.InfixExpression:
		// 检查是否是链式比较：left 是比较表达式，operator 是比较运算符
		if leftInfix, ok := e.Left.(*ast.InfixExpression); ok && isComparisonOp(leftInfix.Operator) && isComparisonOp(e.Operator) {
			// a < b < c -> (a < b) AND (b < c)
			// 提取第一个比较 a < b
			firstComp := &ast.InfixExpression{
				Token:    leftInfix.Token,
				Left:     leftInfix.Left,
				Operator: leftInfix.Operator,
				Right:    leftInfix.Right,
			}
			// 第二个比较使用 leftInfix.Right 和 e.Right
			secondComp := &ast.InfixExpression{
				Token:    e.Token,
				Left:     leftInfix.Right,
				Operator: e.Operator,
				Right:    e.Right,
			}
			// 构建 AND 表达式 (a < b) AND (b < c)
			andExpr := &ast.InfixExpression{
				Token:    "and",
				Left:     firstComp,
				Operator: "and",
				Right:    secondComp,
			}
			// 递归地脱糖这个 AND 表达式
			return desugarExpression(andExpr)
		}

		// 继续检查单个比较表达式后是否还有链式比较
		if nextInfix, ok := e.Right.(*ast.InfixExpression); ok && isComparisonOp(e.Operator) && isComparisonOp(nextInfix.Operator) {
			// a < b < c -> (a < b) AND (b < c) (right-associative)
			// 第一个比较
			firstComp := &ast.InfixExpression{
				Token:    e.Token,
				Left:     e.Left,
				Operator: e.Operator,
				Right:    nextInfix.Left,
			}
			// 第二个比较
			secondComp := &ast.InfixExpression{
				Token:    nextInfix.Token,
				Left:     nextInfix.Left,
				Operator: nextInfix.Operator,
				Right:    nextInfix.Right,
			}
			// 构建 AND 表达式
			andExpr := &ast.InfixExpression{
				Token:    "and",
				Left:     firstComp,
				Operator: "and",
				Right:    secondComp,
			}
			// 递归地脱糖
			return desugarExpression(andExpr)
		}

		// 检查是否是 AND 或 OR，特殊处理
		if e.Operator == "and" {
			// a AND b -> if a then b else a
			left := desugarExpression(e.Left)
			right := desugarExpression(e.Right)
			consequence := &ast.BlockStatement{
				Token: e.Token,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      e.Token,
						Expression: right,
					},
				},
			}
			alternative := &ast.BlockStatement{
				Token: e.Token,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      e.Token,
						Expression: left,
					},
				},
			}
			return &ast.IfExpression{
				Token:       e.Token,
				Condition:   left,
				Consequence: consequence,
				Alternative: alternative,
			}
		} else if e.Operator == "or" {
			// a OR b -> if a then a else b
			left := desugarExpression(e.Left)
			right := desugarExpression(e.Right)
			consequence := &ast.BlockStatement{
				Token: e.Token,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      e.Token,
						Expression: left,
					},
				},
			}
			alternative := &ast.BlockStatement{
				Token: e.Token,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      e.Token,
						Expression: right,
					},
				},
			}
			return &ast.IfExpression{
				Token:       e.Token,
				Condition:   left,
				Consequence: consequence,
				Alternative: alternative,
			}
		}

		// 对于其他运算符，正常脱糖
		return &ast.InfixExpression{
			Token:    e.Token,
			Left:     desugarExpression(e.Left),
			Operator: e.Operator,
			Right:    desugarExpression(e.Right),
		}
	case *ast.IfExpression:
		return &ast.IfExpression{
			Token:       e.Token,
			Condition:   desugarExpression(e.Condition),
			Consequence: desugarBlockStatement(e.Consequence),
			Alternative: desugarBlockStatement(e.Alternative),
		}
	case *ast.TernaryExpression:
		// 将三元表达式转换为 IfExpression
		// a if b else c -> if b { a } else { c }
		consequenceBlock := &ast.BlockStatement{
			Token: e.Token,
			Statements: []ast.Statement{
				&ast.ExpressionStatement{
					Token:      e.Token,
					Expression: desugarExpression(e.Consequence),
				},
			},
		}

		var alternativeBlock *ast.BlockStatement
		if e.Alternative != nil {
			alternativeBlock = &ast.BlockStatement{
				Token: e.Token,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      e.Token,
						Expression: desugarExpression(e.Alternative),
					},
				},
			}
		}

		return &ast.IfExpression{
			Token:       e.Token,
			Condition:   desugarExpression(e.Condition),
			Consequence: consequenceBlock,
			Alternative: alternativeBlock,
		}
	case *ast.FunctionLiteral:
		desugaredDecorators := make([]ast.Expression, len(e.Decorators))
		for i, dec := range e.Decorators {
			desugaredDecorators[i] = desugarExpression(dec)
		}

		desugaredBody := desugarBlockStatement(e.Body)

		defaultStmts := []ast.Statement{}
		for i, def := range e.Defaults {
			if def != nil {
				paramName := e.Parameters[i].Value
				defaultStmts = append(defaultStmts, &ast.ExpressionStatement{
					Token: "if",
					Expression: &ast.IfExpression{
						Token: "if",
						Condition: &ast.InfixExpression{
							Token:    "==",
							Left:     &ast.Identifier{Token: paramName, Value: paramName},
							Operator: "==",
							Right:    &ast.Identifier{Token: "None", Value: "None"},
						},
						Consequence: &ast.BlockStatement{
							Statements: []ast.Statement{
								&ast.AssignStatement{
									Token: paramName,
									Names: []*ast.Identifier{{Token: paramName, Value: paramName}},
									Value: desugarExpression(def),
								},
								&ast.ExpressionStatement{
									Token:      "None",
									Expression: &ast.Identifier{Token: "None", Value: "None"},
								},
							},
						},
					},
				})
			}
		}

		if len(defaultStmts) > 0 {
			newStmts := make([]ast.Statement, 0, len(defaultStmts)+len(desugaredBody.Statements))
			newStmts = append(newStmts, defaultStmts...)
			newStmts = append(newStmts, desugaredBody.Statements...)
			desugaredBody = &ast.BlockStatement{
				Statements: newStmts,
			}
		}

		return &ast.FunctionLiteral{
			Token:         e.Token,
			Name:          e.Name,
			Parameters:    e.Parameters,
			Defaults:      e.Defaults,
			KeywordOnly:   e.KeywordOnly,
			PositionalOnly: e.PositionalOnly,
			Body:          desugaredBody,
			VarArgs:       e.VarArgs,
			KwArgs:        e.KwArgs,
			Decorators:    desugaredDecorators,
			IsAsync:       e.IsAsync,
		}
	case *ast.LambdaExpression:
		return &ast.LambdaExpression{
			Token:     e.Token,
			Parameters: e.Parameters,
			Body:      desugarExpression(e.Body),
		}
	case *ast.CallExpression:
		desugaredArgs := make([]ast.Expression, 0, len(e.Arguments))
		for _, arg := range e.Arguments {
			desugaredArgs = append(desugaredArgs, desugarExpression(arg))
		}
		return &ast.CallExpression{
			Token:     e.Token,
			Function:  desugarExpression(e.Function),
			Arguments: desugaredArgs,
		}
	case *ast.IndexExpression:
		return &ast.IndexExpression{
			Token: e.Token,
			Left:  desugarExpression(e.Left),
			Index: desugarExpression(e.Index),
		}
	case *ast.SliceExpression:
		// SliceExpression 现在是 IndexExpression.Index，不会直接出现
		// 这个 case 是为了完整性
		return &ast.SliceExpression{
			Token: e.Token,
			Lower: desugarExpression(e.Lower),
			Upper: desugarExpression(e.Upper),
			Step:  desugarExpression(e.Step),
		}
	case *ast.DictionaryUnpack:
		return &ast.DictionaryUnpack{
			Token: e.Token,
			Value: desugarExpression(e.Value),
		}
	case *ast.ListUnpack:
		return &ast.ListUnpack{
			Token: e.Token,
			Value: desugarExpression(e.Value),
		}
	case *ast.KeyValuePair:
		return &ast.KeyValuePair{
			Token: e.Token,
			Key:   desugarExpression(e.Key),
			Value: desugarExpression(e.Value),
		}
	case *ast.HashLiteral:
		if e.Elements != nil {
			// 混合字典字面量，需要脱糖
			desugaredElements := []ast.Expression{}
			for _, el := range e.Elements {
				desugaredElements = append(desugaredElements, desugarExpression(el))
			}
			return desugarMixedDictLiteral(desugaredElements)
		}
		// 旧格式兼容
		desugaredPairs := make(map[ast.Expression]ast.Expression)
		for key, value := range e.Pairs {
			desugaredPairs[desugarExpression(key)] = desugarExpression(value)
		}
		return &ast.HashLiteral{
			Token: e.Token,
			Pairs: desugaredPairs,
		}
	case *ast.ListLiteral:
		// 检查是否包含解包元素
		hasUnpack := false
		for _, el := range e.Elements {
			if _, ok := el.(*ast.ListUnpack); ok {
				hasUnpack = true
				break
			}
		}
		if hasUnpack {
			// 混合列表字面量，脱糖
			desugaredElements := []ast.Expression{}
			for _, el := range e.Elements {
				desugaredElements = append(desugaredElements, desugarExpression(el))
			}
			return desugarMixedListLiteral(desugaredElements)
		}
		// 普通列表字面量
		desugaredElements := []ast.Expression{}
		for _, el := range e.Elements {
			desugaredElements = append(desugaredElements, desugarExpression(el))
		}
		return &ast.ListLiteral{
			Token:    e.Token,
			Elements: desugaredElements,
		}
	case *ast.ListComprehension:
		return desugarListComprehension(e)
	case *ast.DictComprehension:
		// 对字典推导式中的子表达式进行脱糖
		dc := &ast.DictComprehension{
			Token:    e.Token,
			Key:      desugarExpression(e.Key),
			Value:    desugarExpression(e.Value),
			Variable: e.Variable,
			Iterable: desugarExpression(e.Iterable),
			Filter:   e.Filter,
		}
		if e.Filter != nil {
			dc.Filter = desugarExpression(e.Filter)
		}
		return dc
	case *ast.SetComprehension:
		// 对集合推导式中的子表达式进行脱糖
		sc := &ast.SetComprehension{
			Token:    e.Token,
			Element:  desugarExpression(e.Element),
			Variable: e.Variable,
			Iterable: desugarExpression(e.Iterable),
			Filter:   e.Filter,
		}
		if e.Filter != nil {
			sc.Filter = desugarExpression(e.Filter)
		}
		return sc
	case *ast.GeneratorExpression:
		// 对生成器表达式中的子表达式进行脱糖
		ge := &ast.GeneratorExpression{
			Token:    e.Token,
			Element:  desugarExpression(e.Element),
			Variable: e.Variable,
			Iterable: desugarExpression(e.Iterable),
			Filter:   e.Filter,
		}
		if e.Filter != nil {
			ge.Filter = desugarExpression(e.Filter)
		}
		return ge
	case *ast.AsyncListComprehension:
		return desugarAsyncListComprehension(e)
	case *ast.AsyncSetComprehension:
		asc := &ast.AsyncSetComprehension{
			Token:    e.Token,
			Element:  desugarExpression(e.Element),
			Variable: e.Variable,
			Iterable: desugarExpression(e.Iterable),
			Filter:   e.Filter,
		}
		if e.Filter != nil {
			asc.Filter = desugarExpression(e.Filter)
		}
		return asc
	case *ast.AsyncDictComprehension:
		adc := &ast.AsyncDictComprehension{
			Token:    e.Token,
			Key:      desugarExpression(e.Key),
			Value:    desugarExpression(e.Value),
			Variable: e.Variable,
			Iterable: desugarExpression(e.Iterable),
			Filter:   e.Filter,
		}
		if e.Filter != nil {
			adc.Filter = desugarExpression(e.Filter)
		}
		return adc
	case *ast.AsyncGeneratorExpression:
		age := &ast.AsyncGeneratorExpression{
			Token:    e.Token,
			Element:  desugarExpression(e.Element),
			Variable: e.Variable,
			Iterable: desugarExpression(e.Iterable),
			Filter:   e.Filter,
		}
		if e.Filter != nil {
			age.Filter = desugarExpression(e.Filter)
		}
		return age
	case *ast.FStringLiteral:
		// Keep f-string as-is, the compiler will handle it
		desugaredParts := make([]ast.Expression, 0, len(e.Parts))
		for _, part := range e.Parts {
			desugaredParts = append(desugaredParts, desugarExpression(part))
		}
		return &ast.FStringLiteral{
			Token: e.Token,
			Parts: desugaredParts,
		}
	default:
		return expr
	}
}

func isComparisonOp(op string) bool {
	return op == "==" || op == "!=" || op == "<" || op == ">" || op == "<=" || op == ">="
}

func desugarListComprehension(lc *ast.ListComprehension) ast.Expression {
	// 直接返回，让编译器来处理列表推导式
	// 我们需要在这里脱糖子表达式
	lc.Element = desugarExpression(lc.Element)
	lc.Iterable = desugarExpression(lc.Iterable)
	if lc.Filter != nil {
		lc.Filter = desugarExpression(lc.Filter)
	}
	return lc
}

func desugarAsyncListComprehension(alc *ast.AsyncListComprehension) ast.Expression {
	// 脱糖子表达式，保持AsyncListComprehension节点不变，让编译器处理
	alc.Element = desugarExpression(alc.Element)
	alc.Iterable = desugarExpression(alc.Iterable)
	if alc.Filter != nil {
		alc.Filter = desugarExpression(alc.Filter)
	}
	return alc
}

var forLoopCounter int

func desugarForToWhile(forStmt *ast.ForStatement) *ast.BlockStatement {
	forLoopCounter++
	indexVar := &ast.Identifier{Token: fmt.Sprintf("_i_%d", forLoopCounter), Value: fmt.Sprintf("_i_%d", forLoopCounter)}
	iterable := desugarExpression(forStmt.Iterable)

	condition := &ast.InfixExpression{
		Token:    "<",
		Left:     indexVar,
		Operator: "<",
		Right: &ast.CallExpression{
			Token:    "len",
			Function: &ast.Identifier{Token: "len", Value: "len"},
			Arguments: []ast.Expression{iterable},
		},
	}

	loopBodyStmts := []ast.Statement{
		&ast.AssignStatement{
			Token: "=",
			Names: []*ast.Identifier{forStmt.Value},
			Value: &ast.IndexExpression{
				Token: "[",
				Left:  iterable,
				Index: indexVar,
			},
		},
	}

	loopBodyStmts = append(loopBodyStmts, desugarBlockStatement(forStmt.Body).Statements...)

	loopBodyStmts = append(loopBodyStmts, &ast.AssignStatement{
		Token: "+=",
		Names: []*ast.Identifier{indexVar},
		Value: &ast.InfixExpression{
			Token:    "+",
			Left:     indexVar,
			Operator: "+",
			Right:    &ast.IntegerLiteral{Token: "1", Value: 1},
		},
	})

	loopBody := &ast.BlockStatement{
		Token:      forStmt.Token,
		Statements: loopBodyStmts,
	}

	whileStmt := &ast.WhileStatement{
		Token:     forStmt.Token,
		Condition: condition,
		Body:      loopBody,
	}

	block := &ast.BlockStatement{
		Token: forStmt.Token,
		Statements: []ast.Statement{
			&ast.LetStatement{
				Token: "let",
				Names: []*ast.Identifier{indexVar},
				Value: &ast.IntegerLiteral{Token: "0", Value: 0},
			},
			whileStmt,
		},
	}
	return block
}

func desugarMixedListLiteral(elements []ast.Expression) ast.Expression {
	// 创建临时变量名称
	tempName := &ast.Identifier{Token: "_list", Value: "_list"}
	statements := []ast.Statement{}
	// 创建空列表
	statements = append(statements, &ast.LetStatement{
		Token: "let",
		Names: []*ast.Identifier{tempName},
		Value: &ast.ListLiteral{
			Token:    "[",
			Elements: []ast.Expression{},
		},
	})

	for _, el := range elements {
		if unpack, ok := el.(*ast.ListUnpack); ok {
			// 解包：调用 extend
			statements = append(statements, &ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Token: "extend",
					Function: &ast.MemberAccess{
						Token: ".",
						Object: tempName,
						Member: &ast.Identifier{Token: "extend", Value: "extend"},
					},
					Arguments: []ast.Expression{unpack.Value},
				},
			})
		} else {
			// 普通元素：调用 append
			statements = append(statements, &ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Token: "append",
					Function: &ast.MemberAccess{
						Token: ".",
						Object: tempName,
						Member: &ast.Identifier{Token: "append", Value: "append"},
					},
					Arguments: []ast.Expression{el},
				},
			})
		}
	}

	// 返回临时变量作为结果
	statements = append(statements, &ast.ReturnStatement{
		Token:       "return",
		ReturnValue: tempName,
	})

	// 将整个序列包装在一个立即执行的函数中
	resultExpr := &ast.CallExpression{
		Token: "()",
		Function: &ast.FunctionLiteral{
			Token:       "def",
			Name:        "",
			Parameters:  []*ast.Identifier{},
			Body:        &ast.BlockStatement{Token: "{", Statements: statements},
			Decorators:  []ast.Expression{},
			IsAsync:     false,
		},
		Arguments: []ast.Expression{},
	}

	return resultExpr
}

func desugarMixedDictLiteral(elements []ast.Expression) ast.Expression {
	// 创建临时变量名称
	tempName := &ast.Identifier{Token: "_dict", Value: "_dict"}
	statements := []ast.Statement{}
	// 创建空字典
	statements = append(statements, &ast.LetStatement{
		Token: "let",
		Names: []*ast.Identifier{tempName},
		Value: &ast.HashLiteral{
			Token: "{",
			Pairs: map[ast.Expression]ast.Expression{},
		},
	})

	for _, el := range elements {
		if unpack, ok := el.(*ast.DictionaryUnpack); ok {
			// 字典解包：调用 update
			statements = append(statements, &ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Token: "update",
					Function: &ast.MemberAccess{
						Token: ".",
						Object: tempName,
						Member: &ast.Identifier{Token: "update", Value: "update"},
					},
					Arguments: []ast.Expression{unpack.Value},
				},
			})
		} else if kv, ok := el.(*ast.KeyValuePair); ok {
			// 键值对：调用 __setitem__
			statements = append(statements, &ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Token: "__setitem__",
					Function: &ast.MemberAccess{
						Token: ".",
						Object: tempName,
						Member: &ast.Identifier{Token: "__setitem__", Value: "__setitem__"},
					},
					Arguments: []ast.Expression{kv.Key, kv.Value},
				},
			})
		}
	}

	// 返回临时变量作为结果
	statements = append(statements, &ast.ReturnStatement{
		Token:       "return",
		ReturnValue: tempName,
	})

	// 将整个序列包装在一个立即执行的函数中
	resultExpr := &ast.CallExpression{
		Token: "()",
		Function: &ast.FunctionLiteral{
			Token:       "def",
			Name:        "",
			Parameters:  []*ast.Identifier{},
			Body:        &ast.BlockStatement{Token: "{", Statements: statements},
			Decorators:  []ast.Expression{},
			IsAsync:     false,
		},
		Arguments: []ast.Expression{},
	}

	return resultExpr
}

// desugarDeleteStatement 脱糖 del 语句
// 对于简单变量和成员访问，保留 DeleteStatement 由编译器直接处理
// 对于下标访问，转换为 __delitem__ 调用
func desugarDeleteStatement(stmt *ast.DeleteStatement) ast.Statement {
	desugaredStmts := make([]ast.Statement, 0, len(stmt.Targets))

	for _, target := range stmt.Targets {
		switch t := desugarExpression(target).(type) {
		case *ast.Identifier:
			// 简单标识符，保留原样
			desugaredStmts = append(desugaredStmts, &ast.DeleteStatement{
				Token:   stmt.Token,
				Targets: []ast.Expression{t},
			})
		case *ast.IndexExpression:
			// 下标访问：del x[y] -> x.__delitem__(y)
			desugaredStmts = append(desugaredStmts, &ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Token: "__delitem__",
					Function: &ast.MemberAccess{
						Token:  ".",
						Object: t.Left,
						Member: &ast.Identifier{Token: "__delitem__", Value: "__delitem__"},
					},
					Arguments: []ast.Expression{t.Index},
				},
			})
		case *ast.MemberAccess:
			// 成员访问：保留 DeleteStatement，由编译器 OpDelAttribute 处理
			desugaredStmts = append(desugaredStmts, &ast.DeleteStatement{
				Token:   stmt.Token,
				Targets: []ast.Expression{t},
			})
		default:
			// 其他情况，保留原样
			desugaredStmts = append(desugaredStmts, &ast.DeleteStatement{
				Token:   stmt.Token,
				Targets: []ast.Expression{t},
			})
		}
	}

	if len(desugaredStmts) == 1 {
		return desugaredStmts[0]
	}

	return &ast.BlockStatement{
		Token:      stmt.Token,
		Statements: desugaredStmts,
	}
}

// desugarYieldFromStatement 脱糖 yield from 语句
// yield from iter 脱糖为：for item in iter: yield item
func desugarYieldFromStatement(stmt *ast.YieldFromStatement) ast.Statement {
	// 首先脱糖迭代器表达式
	iterExpr := desugarExpression(stmt.Expression)

	// 创建临时变量来迭代
	itemIdent := &ast.Identifier{Token: "_item", Value: "_item"}

	// 创建 for 循环
	forStmt := &ast.ForStatement{
		Token:    "for",
		Value:    itemIdent,
		Iterable: iterExpr,
		Body: &ast.BlockStatement{
			Token: "{",
			Statements: []ast.Statement{
				&ast.YieldStatement{
					Token:      "yield",
					Expression: itemIdent,
				},
			},
		},
	}

	// 返回脱糖后的 for 循环
	return desugarForToWhile(forStmt)
}

// desugarAsyncForStatement 脱糖 async for 语句
// 我们保留原样，因为异步操作需要特殊处理
func desugarAsyncForStatement(stmt *ast.AsyncForStatement) ast.Statement {
	// 脱糖迭代器和循环体
	desugaredIterable := desugarExpression(stmt.Iterable)
	desugaredBody := desugarBlockStatement(stmt.Body)

	return &ast.AsyncForStatement{
		Token:    stmt.Token,
		Value:    stmt.Value,
		Iterable: desugaredIterable,
		Body:     desugaredBody,
	}
}

// desugarAsyncWithStatement 脱糖 async with 语句
// 我们保留原样，因为异步操作需要特殊处理
func desugarAsyncWithStatement(stmt *ast.AsyncWithStatement) ast.Statement {
	// 脱糖上下文管理器表达式和循环体
	desugaredItems := make([]*ast.ContextManagerItem, 0, len(stmt.Items))
	for _, item := range stmt.Items {
		desugaredItems = append(desugaredItems, &ast.ContextManagerItem{
			Expr: desugarExpression(item.Expr),
			Name: item.Name,
		})
	}
	desugaredBody := desugarBlockStatement(stmt.Body)

	// 如果只有一个上下文管理器，直接返回
	if len(desugaredItems) == 1 {
		return &ast.AsyncWithStatement{
			Token: stmt.Token,
			Items: desugaredItems,
			Body:  desugaredBody,
		}
	}

	// 多个上下文管理器，嵌套处理
	var nestedStatement ast.Statement = &ast.AsyncWithStatement{
		Token: stmt.Token,
		Items: []*ast.ContextManagerItem{desugaredItems[len(desugaredItems)-1]},
		Body:  desugaredBody,
	}

	for i := len(desugaredItems) - 2; i >= 0; i-- {
		nestedStatement = &ast.AsyncWithStatement{
			Token: stmt.Token,
			Items: []*ast.ContextManagerItem{desugaredItems[i]},
			Body: &ast.BlockStatement{
				Token:      stmt.Token,
				Statements: []ast.Statement{nestedStatement},
			},
		}
	}

	return nestedStatement
}

func desugarMatchStatement(ms *ast.MatchStatement) ast.Statement {
	matchVar := &ast.Identifier{Token: "_match_val", Value: "_match_val"}

	stmts := []ast.Statement{
		&ast.LetStatement{
			Token: ms.Token,
			Names: []*ast.Identifier{matchVar},
			Value: desugarExpression(ms.Subject),
		},
	}

	var chain ast.Expression
	for i := len(ms.Cases) - 1; i >= 0; i-- {
		cc := ms.Cases[i]
		condition := buildMatchCondition(matchVar, cc.Pattern)
		if cc.Guard != nil {
			condition = &ast.InfixExpression{
				Token:    "and",
				Left:     condition,
				Operator: "and",
				Right:    desugarExpression(cc.Guard),
			}
		}

		bindings := collectPatternBindings(matchVar, cc.Pattern)
		desugaredBody := desugarBlockStatement(cc.Body)

		bodyStmts := make([]ast.Statement, 0, len(bindings)+len(desugaredBody.Statements))
		bodyStmts = append(bodyStmts, bindings...)
		bodyStmts = append(bodyStmts, desugaredBody.Statements...)

		body := &ast.BlockStatement{
			Token:      cc.Token,
			Statements: bodyStmts,
		}

		if chain == nil {
			chain = &ast.IfExpression{
				Token:       "if",
				Condition:   condition,
				Consequence: body,
			}
		} else {
			chain = &ast.IfExpression{
				Token:       "if",
				Condition:   condition,
				Consequence: body,
				Alternative: &ast.BlockStatement{
					Token:      "else",
					Statements: []ast.Statement{&ast.ExpressionStatement{Token: "else", Expression: chain}},
				},
			}
		}
	}

	if chain != nil {
		stmts = append(stmts, &ast.ExpressionStatement{
			Token:      ms.Token,
			Expression: chain,
		})
	}

	if len(stmts) == 1 {
		return &ast.ExpressionStatement{Token: ms.Token, Expression: &ast.Boolean{Token: "True", Value: true}}
	}

	return &ast.BlockStatement{
		Token:      ms.Token,
		Statements: stmts,
	}
}

func buildMatchCondition(matchVar ast.Expression, pattern ast.Expression) ast.Expression {
	if pattern == nil {
		return &ast.Boolean{Token: "True", Value: true}
	}

	switch p := pattern.(type) {
	case *ast.IntegerLiteral:
		return &ast.InfixExpression{
			Token:    "==",
			Left:     matchVar,
			Operator: "==",
			Right:    p,
		}
	case *ast.FloatLiteral:
		return &ast.InfixExpression{
			Token:    "==",
			Left:     matchVar,
			Operator: "==",
			Right:    p,
		}
	case *ast.ComplexLiteral:
		return &ast.InfixExpression{
			Token:    "==",
			Left:     matchVar,
			Operator: "==",
			Right:    p,
		}
	case *ast.StringLiteral:
		return &ast.InfixExpression{
			Token:    "==",
			Left:     matchVar,
			Operator: "==",
			Right:    p,
		}
	case *ast.Boolean:
		return &ast.InfixExpression{
			Token:    "==",
			Left:     matchVar,
			Operator: "==",
			Right:    p,
		}
	case *ast.Identifier:
		if p.Value == "_" {
			return &ast.Boolean{Token: "True", Value: true}
		}
		if p.Value == "None" {
			return &ast.InfixExpression{
				Token:    "==",
				Left:     matchVar,
				Operator: "==",
				Right:    &ast.Identifier{Token: "None", Value: "None"},
			}
		}
		if p.Value == "True" {
			return &ast.InfixExpression{
				Token:    "==",
				Left:     matchVar,
				Operator: "==",
				Right:    &ast.Boolean{Token: "True", Value: true},
			}
		}
		if p.Value == "False" {
			return &ast.InfixExpression{
				Token:    "==",
				Left:     matchVar,
				Operator: "==",
				Right:    &ast.Boolean{Token: "False", Value: false},
			}
		}
		return &ast.Boolean{Token: "True", Value: true}
	case *ast.InfixExpression:
		if p.Operator == "|" {
			left := buildMatchCondition(matchVar, p.Left)
			right := buildMatchCondition(matchVar, p.Right)
			return &ast.InfixExpression{
				Token:    "or",
				Left:     left,
				Operator: "or",
				Right:    right,
			}
		}
	case *ast.CallExpression:
		if ident, ok := p.Function.(*ast.Identifier); ok {
			isInstanceCall := &ast.CallExpression{
				Token: "isinstance",
				Function: &ast.Identifier{Token: "isinstance", Value: "isinstance"},
				Arguments: []ast.Expression{matchVar, ident},
			}
			if len(p.Arguments) > 0 {
				argConditions := []ast.Expression{isInstanceCall}
				for i, arg := range p.Arguments {
					idx := &ast.IntegerLiteral{Token: fmt.Sprintf("%d", i), Value: int64(i)}
					getItem := &ast.IndexExpression{
						Token: "[",
						Left:  matchVar,
						Index: idx,
					}
					argCond := &ast.InfixExpression{
						Token: "and",
						Left: &ast.InfixExpression{
							Token:    ">=",
							Left: &ast.CallExpression{
								Token:    "len",
								Function: &ast.Identifier{Token: "len", Value: "len"},
								Arguments: []ast.Expression{matchVar},
							},
							Operator: ">=",
							Right:    idx,
						},
						Operator: "and",
						Right:    buildMatchCondition(getItem, arg),
					}
					argConditions = append(argConditions, argCond)
				}
				result := argConditions[0]
				for _, cond := range argConditions[1:] {
					result = &ast.InfixExpression{
						Token:    "and",
						Left:     result,
						Operator: "and",
						Right:    cond,
					}
				}
				return result
			}
			return isInstanceCall
		}
	case *ast.ListLiteral:
		conditions := []ast.Expression{
			&ast.InfixExpression{
				Token: "==",
				Left: &ast.CallExpression{
					Token:    "len",
					Function: &ast.Identifier{Token: "len", Value: "len"},
					Arguments: []ast.Expression{matchVar},
				},
				Operator: "==",
				Right:    &ast.IntegerLiteral{Token: fmt.Sprintf("%d", len(p.Elements)), Value: int64(len(p.Elements))},
			},
		}
		for i, elem := range p.Elements {
			idx := &ast.IntegerLiteral{Token: fmt.Sprintf("%d", i), Value: int64(i)}
			getItem := &ast.IndexExpression{
				Token: "[",
				Left:  matchVar,
				Index: idx,
			}
			elemCond := buildMatchCondition(getItem, elem)
			conditions = append(conditions, elemCond)
		}
		result := conditions[0]
		for _, cond := range conditions[1:] {
			result = &ast.InfixExpression{
				Token:    "and",
				Left:     result,
				Operator: "and",
				Right:    cond,
			}
		}
		return result
	}

	return &ast.Boolean{Token: "True", Value: true}
}

func collectPatternBindings(matchVar ast.Expression, pattern ast.Expression) []ast.Statement {
	if pattern == nil {
		return nil
	}

	var bindings []ast.Statement

	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Value != "_" {
			bindings = append(bindings, &ast.AssignStatement{
				Token: p.Token,
				Names: []*ast.Identifier{p},
				Value: matchVar,
			})
		}
	case *ast.InfixExpression:
		if p.Operator == "|" {
			bindings = append(bindings, collectPatternBindings(matchVar, p.Left)...)
		}
	case *ast.CallExpression:
		if _, ok := p.Function.(*ast.Identifier); ok {
			for i, arg := range p.Arguments {
				if ident, ok := arg.(*ast.Identifier); ok && ident.Value != "_" {
					idx := &ast.IntegerLiteral{Token: fmt.Sprintf("%d", i), Value: int64(i)}
					getItem := &ast.IndexExpression{
						Token: "[",
						Left:  matchVar,
						Index: idx,
					}
					bindings = append(bindings, &ast.AssignStatement{
						Token: ident.Token,
						Names: []*ast.Identifier{ident},
						Value: getItem,
					})
				}
			}
		}
	case *ast.ListLiteral:
		for i, elem := range p.Elements {
			idx := &ast.IntegerLiteral{Token: fmt.Sprintf("%d", i), Value: int64(i)}
			getItem := &ast.IndexExpression{
				Token: "[",
				Left:  matchVar,
				Index: idx,
			}
			bindings = append(bindings, collectPatternBindings(getItem, elem)...)
		}
	}

	return bindings
}

func isLruCacheDecorator(dec ast.Expression) bool {
	if ident, ok := dec.(*ast.Identifier); ok {
		return ident.Value == "lru_cache"
	}
	if call, ok := dec.(*ast.CallExpression); ok {
		if ident, ok := call.Function.(*ast.Identifier); ok {
			return ident.Value == "lru_cache"
		}
		if ma, ok := call.Function.(*ast.MemberAccess); ok {
			return ma.Member.Value == "lru_cache"
		}
	}
	if ma, ok := dec.(*ast.MemberAccess); ok {
		return ma.Member.Value == "lru_cache"
	}
	return false
}

func desugarLruCache(fn *ast.FunctionLiteral) ast.Statement {
	funcIdent := &ast.Identifier{Token: fn.Token, Value: fn.Name}
	origName := "_lru_orig_" + fn.Name
	origIdent := &ast.Identifier{Token: fn.Token, Value: origName}
	cacheAttr := "_cache"

	cacheIdent := &ast.MemberAccess{
		Token:  ".",
		Object: funcIdent,
		Member: &ast.Identifier{Token: cacheAttr, Value: cacheAttr},
	}

	keyVar := &ast.Identifier{Token: "_lru_key", Value: "_lru_key"}
	resultVar := &ast.Identifier{Token: "_lru_result", Value: "_lru_result"}

	keyElements := make([]ast.Expression, len(fn.Parameters))
	for i, p := range fn.Parameters {
		keyElements[i] = p
	}
	keyTuple := &ast.ListLiteral{Token: "(", Elements: keyElements}

	origArgs := make([]ast.Expression, len(fn.Parameters))
	for i, p := range fn.Parameters {
		origArgs[i] = p
	}
	origCall := &ast.CallExpression{
		Token:     "(",
		Function:  origIdent,
		Arguments: origArgs,
	}

	cacheLookup := &ast.InfixExpression{
		Token:    "in",
		Left:     keyTuple,
		Operator: "in",
		Right:    cacheIdent,
	}

	cacheGet := &ast.IndexExpression{
		Token: "[",
		Left:  cacheIdent,
		Index: keyTuple,
	}

	wrapperBody := &ast.BlockStatement{
		Token: ":",
		Statements: []ast.Statement{
			&ast.LetStatement{
				Token: "let",
				Names: []*ast.Identifier{keyVar},
				Value: keyTuple,
			},
			&ast.ExpressionStatement{
				Token: "if",
				Expression: &ast.IfExpression{
					Token:     "if",
					Condition: cacheLookup,
					Consequence: &ast.BlockStatement{
						Token: ":",
						Statements: []ast.Statement{
							&ast.ReturnStatement{
								Token:       "return",
								ReturnValue: cacheGet,
							},
						},
					},
				},
			},
			&ast.LetStatement{
				Token: "let",
				Names: []*ast.Identifier{resultVar},
				Value: origCall,
			},
			&ast.ExpressionStatement{
				Token:      "=",
				Expression: &ast.InfixExpression{
					Token:    "=",
					Left:     &ast.IndexExpression{Token: "[", Left: cacheIdent, Index: keyTuple},
					Operator: "=",
					Right:    resultVar,
				},
			},
			&ast.ReturnStatement{
				Token:       "return",
				ReturnValue: resultVar,
			},
		},
	}

	wrapperFn := &ast.FunctionLiteral{
		Token:       fn.Token,
		Name:        fn.Name,
		Parameters:  fn.Parameters,
		Defaults:    fn.Defaults,
		KeywordOnly: fn.KeywordOnly,
		Body:        wrapperBody,
		VarArgs:     fn.VarArgs,
		KwArgs:      fn.KwArgs,
	}

	origFn := &ast.FunctionLiteral{
		Token:       fn.Token,
		Name:        "",
		Parameters:  fn.Parameters,
		Defaults:    fn.Defaults,
		KeywordOnly: fn.KeywordOnly,
		Body:        desugarBlockStatement(fn.Body),
		VarArgs:    fn.VarArgs,
		KwArgs:     fn.KwArgs,
	}

	return &ast.BlockStatement{
		Token: fn.Token,
		Statements: []ast.Statement{
			&ast.LetStatement{
				Token: fn.Token,
				Names: []*ast.Identifier{origIdent},
				Value: origFn,
			},
			&ast.LetStatement{
				Token: fn.Token,
				Names: []*ast.Identifier{funcIdent},
				Value: wrapperFn,
			},
			&ast.ExpressionStatement{
				Token: fn.Token,
				Expression: &ast.InfixExpression{
					Token:    "=",
					Left:     cacheIdent,
					Operator: "=",
					Right:    &ast.HashLiteral{Token: "{", Pairs: map[ast.Expression]ast.Expression{}},
				},
			},
		},
	}
}
