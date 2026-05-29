package desugar

import (
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
			// 脱糖子表达式
			desugaredFn := desugarExpression(fnLit).(*ast.FunctionLiteral)
			desugaredDecorators := make([]ast.Expression, len(desugaredFn.Decorators))
			for i, dec := range desugaredFn.Decorators {
				desugaredDecorators[i] = desugarExpression(dec)
			}

			// 创建一个块语句，包含：
			// 1. 定义原始函数（临时名称）
			// 2. 用装饰器包装它
			// 3. 将结果赋值回原函数名
			stmts := []ast.Statement{}

			// 原始函数名
			funcIdent := &ast.Identifier{Token: desugaredFn.Token, Value: desugaredFn.Name}
			// 临时函数名
			tempIdent := &ast.Identifier{Token: desugaredFn.Token, Value: "_temp_" + desugaredFn.Name}

			// 1. 把原始函数定义赋值给临时变量
			tempFn := &ast.FunctionLiteral{
				Token:      desugaredFn.Token,
				Name:       "",
				Parameters: desugaredFn.Parameters,
				Body:       desugaredFn.Body,
				VarArgs:    desugaredFn.VarArgs,
				KwArgs:     desugaredFn.KwArgs,
			}

			letStmt := &ast.LetStatement{
				Token: desugaredFn.Token,
				Names: []*ast.Identifier{tempIdent},
				Value: tempFn,
			}
			stmts = append(stmts, letStmt)

			// 2. 应用装饰器，从最后一个装饰器开始（因为装饰器是从下往上应用的）
			currentValue := tempIdent
			for i := len(desugaredDecorators) - 1; i >= 0; i-- {
				decorator := desugaredDecorators[i]
				// 调用装饰器
				callExpr := &ast.CallExpression{
					Token:     decorator.TokenLiteral(),
					Function:  decorator,
					Arguments: []ast.Expression{currentValue},
				}
				// 赋值给临时变量或最终变量
				if i == 0 {
					// 最后一个装饰器，赋值回原函数名
					assignStmt := &ast.AssignStatement{
						Token: desugaredFn.Token,
						Names: []*ast.Identifier{funcIdent},
						Value: callExpr,
					}
					stmts = append(stmts, assignStmt)
				} else {
					// 中间步骤，赋值给临时变量
					letStmt = &ast.LetStatement{
						Token: desugaredFn.Token,
						Names: []*ast.Identifier{tempIdent},
						Value: callExpr,
					}
					stmts = append(stmts, letStmt)
				}
			}

			// 如果只有一个装饰器，直接赋值
			if len(desugaredDecorators) == 1 {
				decorator := desugaredDecorators[0]
				callExpr := &ast.CallExpression{
					Token:     decorator.TokenLiteral(),
					Function:  decorator,
					Arguments: []ast.Expression{tempFn},
				}
				// 清空 stmts，用更简单的方式
				stmts = []ast.Statement{
					&ast.LetStatement{
						Token: desugaredFn.Token,
						Names: []*ast.Identifier{funcIdent},
						Value: callExpr,
					},
				}
			}

			// 返回块语句
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
				Token: ex.Token,
				Type:  desugarExpression(ex.Type),
				Name:  ex.Name,
				Body:  desugarBlockStatement(ex.Body),
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
		// 对 with 语句进行脱糖处理：脱糖表达式和 body
		desugaredWith := &ast.WithStatement{
			Token: s.Token,
			Expr:  desugarExpression(s.Expr),
			Name:  s.Name,
			Body:  desugarBlockStatement(s.Body),
		}
		return desugaredWith
	case *ast.YieldStatement:
		// 对 yield 语句进行脱糖处理：脱糖表达式
		return &ast.YieldStatement{
			Token:      s.Token,
			Expression: desugarExpression(s.Expression),
		}
	case *ast.ClassStatement:
		desugaredClass := &ast.ClassStatement{
			Token:       s.Token,
			Name:        s.Name,
			SuperClass:  s.SuperClass,
			Body:        desugarBlockStatement(s.Body),
			Methods:     s.Methods,
		}
		return desugaredClass
	default:
		return stmt
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
		// 脱糖装饰器
		desugaredDecorators := make([]ast.Expression, len(e.Decorators))
		for i, dec := range e.Decorators {
			desugaredDecorators[i] = desugarExpression(dec)
		}
		return &ast.FunctionLiteral{
			Token:      e.Token,
			Name:       e.Name,
			Parameters: e.Parameters,
			Body:       desugarBlockStatement(e.Body),
			VarArgs:    e.VarArgs,
			KwArgs:     e.KwArgs,
			Decorators: desugaredDecorators,
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
	case *ast.ListLiteral:
		desugaredElements := make([]ast.Expression, 0, len(e.Elements))
		for _, el := range e.Elements {
			desugaredElements = append(desugaredElements, desugarExpression(el))
		}
		return &ast.ListLiteral{
			Token:    e.Token,
			Elements: desugaredElements,
		}
	case *ast.IndexExpression:
		return &ast.IndexExpression{
			Token: e.Token,
			Left:  desugarExpression(e.Left),
			Index: desugarExpression(e.Index),
		}
	case *ast.SliceExpression:
		return &ast.SliceExpression{
			Token: e.Token,
			Left:  desugarExpression(e.Left),
			Start: desugarExpression(e.Start),
			End:   desugarExpression(e.End),
		}
	case *ast.HashLiteral:
		desugaredPairs := make(map[ast.Expression]ast.Expression)
		for key, value := range e.Pairs {
			desugaredPairs[desugarExpression(key)] = desugarExpression(value)
		}
		return &ast.HashLiteral{
			Token: e.Token,
			Pairs: desugaredPairs,
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

func desugarForToWhile(forStmt *ast.ForStatement) *ast.BlockStatement {
	indexVar := &ast.Identifier{Token: "_i", Value: "_i"}
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

	loopBodyStmts = append(loopBodyStmts, &ast.AugAssignStatement{
		Token:    "+=",
		Name:     indexVar,
		Operator: "+",
		Value:    &ast.IntegerLiteral{Token: "1", Value: 1},
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
