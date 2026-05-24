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
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return &ast.ExpressionStatement{
			Token:      s.Token,
			Expression: desugarExpression(s.Expression),
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
			Name:  s.Name,
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
			Name:  s.Name,
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
	default:
		return stmt
	}
}

// desugarBlockStatement 脱糖块语句
func desugarBlockStatement(block *ast.BlockStatement) *ast.BlockStatement {
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
		// 先检查是否是链式比较（不先脱糖，保持结构）
		if nextInfix, ok := e.Right.(*ast.InfixExpression); ok && isComparisonOp(e.Operator) && isComparisonOp(nextInfix.Operator) {
			// 构建 a < b
			firstComp := &ast.InfixExpression{
				Token:    e.Token,
				Left:     e.Left,
				Operator: e.Operator,
				Right:    nextInfix.Left,
			}

			// 构建 AND 表达式 (a < b) AND (b < c)
			andExpr := &ast.InfixExpression{
				Token:    "and",
				Left:     firstComp,
				Operator: "and",
				Right:    nextInfix,
			}

			// 递归地脱糖这个 AND 表达式
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
		return &ast.FunctionLiteral{
			Token:      e.Token,
			Name:       e.Name,
			Parameters: e.Parameters,
			Body:       desugarBlockStatement(e.Body),
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
	case *ast.DictLiteral:
		desugaredPairs := make(map[ast.Expression]ast.Expression)
		for key, value := range e.Pairs {
			desugaredPairs[desugarExpression(key)] = desugarExpression(value)
		}
		return &ast.DictLiteral{
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
	return op == "==" || op == "!=" || op == "<" || op == ">"
}

func desugarListComprehension(lc *ast.ListComprehension) ast.Expression {
	// 直接返回，让编译器来处理列表推导式
	// 我们需要在这里脱糖子表达式
	lc.Element = desugarExpression(lc.Element)
	lc.Iterable = desugarExpression(lc.Iterable)
	if lc.Condition != nil {
		lc.Condition = desugarExpression(lc.Condition)
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

	bodyStmts := []ast.Statement{
		&ast.AssignStatement{
			Token: "=",
			Name:  forStmt.Variable,
			Value: &ast.IndexExpression{
				Token: "[",
				Left:  iterable,
				Index: indexVar,
			},
		},
	}

	bodyStmts = append(bodyStmts, desugarBlockStatement(forStmt.Body).Statements...)

	bodyStmts = append(bodyStmts, &ast.AugAssignStatement{
		Token:    "+=",
		Name:     indexVar,
		Operator: "+",
		Value:    &ast.IntegerLiteral{Token: "1", Value: 1},
	})

	body := &ast.BlockStatement{
		Token:      forStmt.Token,
		Statements: bodyStmts,
	}

	return &ast.WhileStatement{
		Token:     forStmt.Token,
		Condition: condition,
		Body:      body,
	}
}
