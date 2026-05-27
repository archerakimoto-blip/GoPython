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
		if s.Expression == nil {
			return nil
		}
		desugaredExpr := desugarExpression(s.Expression)
		if desugaredExpr == nil {
			return nil
		}
		return &ast.ExpressionStatement{
			Token:      s.Token,
			Expression: desugaredExpr,
		}
	case *ast.LetStatement:
		return &ast.LetStatement{
			Token: s.Token,
			Name:  s.Name,
			Value: desugarExpression(s.Value),
		}
	case *ast.AssignStatement:
		// 保持赋值语句不变，编译器会专门处理它
		return &ast.AssignStatement{
			Token: s.Token,
			Left:  desugarExpression(s.Left),
			Value: desugarExpression(s.Value),
		}
	case *ast.AugAssignStatement:
		// 将增强赋值转换为: left = left op value
		desugaredLeft := desugarExpression(s.Left)
		infixExpr := &ast.InfixExpression{
			Token:    s.Token,
			Left:     desugaredLeft,
			Operator: s.Operator,
			Right:    desugarExpression(s.Value),
		}
		return &ast.AssignStatement{
			Token: s.Token,
			Left:  desugaredLeft,
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
		desugared := &ast.IfExpression{
			Token:       e.Token,
			Condition:   desugarExpression(e.Condition),
			Consequence: desugarBlockStatement(e.Consequence),
		}
		if e.Alternative != nil {
			desugared.Alternative = desugarBlockStatement(e.Alternative)
		}
		return desugared
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
		return &ast.FunctionLiteral{
			Token:      e.Token,
			Name:       e.Name,
			Parameters: e.Parameters,
			Body:       desugarBlockStatement(e.Body),
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
		return e
	case *ast.FStringLiteral:
		// Desugar f-string into a chain of + expressions (concatenation)
		var result ast.Expression
		for _, part := range e.Parts {
			desugaredPart := desugarExpression(part)
			if result == nil {
				result = desugaredPart
			} else {
				result = &ast.InfixExpression{
					Token:    "+",
					Left:     result,
					Operator: "+",
					Right:    desugaredPart,
				}
			}
		}
		// If the f-string is empty, return empty string literal
		if result == nil {
			return &ast.StringLiteral{Value: ""}
		}
		return result
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

func desugarForToWhile(forStmt *ast.ForStatement) *ast.WhileStatement {
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

	// 创建包含初始化和 while 循环的包装块
	wrapperStmts := []ast.Statement{
		// 首先在外部初始化 _i = 0
		&ast.AssignStatement{
			Token: "=",
			Left:  indexVar,
			Value: &ast.IntegerLiteral{Token: "0", Value: 0},
		},
		// 然后是 while 循环
		&ast.WhileStatement{
			Token:     forStmt.Token,
			Condition: condition,
			Body: &ast.BlockStatement{
				Statements: []ast.Statement{
					// 在循环体内赋值给循环变量
					&ast.AssignStatement{
						Token: "=",
						Left:  forStmt.Value,
						Value: &ast.IndexExpression{
							Token: "[",
							Left:  iterable,
							Index: indexVar,
						},
					},
				},
			},
		},
	}

	// 获取 while 循环节点
	whileStmt := wrapperStmts[1].(*ast.WhileStatement)

	// 添加原始 for 循环体的语句到 while 循环体中
	whileStmt.Body.Statements = append(whileStmt.Body.Statements, desugarBlockStatement(forStmt.Body).Statements...)

	// 添加 _i += 1 到 while 循环体末尾
	whileStmt.Body.Statements = append(whileStmt.Body.Statements, &ast.AssignStatement{
		Token: "=",
		Left:  indexVar,
		Value: &ast.InfixExpression{
			Token:    "+",
			Left:     indexVar,
			Operator: "+",
			Right:    &ast.IntegerLiteral{Token: "1", Value: 1},
		},
	})

	// 返回一个包含包装块的无限循环（会在内部 while 循环结束后退出）
	return &ast.WhileStatement{
		Token:     forStmt.Token,
		Condition: &ast.Boolean{Token: "true", Value: true},
		Body: &ast.BlockStatement{
			Statements: wrapperStmts,
		},
	}
}
