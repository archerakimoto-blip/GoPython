package desugar

import (
	"fmt"
	"github.com/go-py/go-python/pkg/ast"
)

// Desugar 对整个程序进行脱糖转换
func Desugar(program *ast.Program) *ast.Program {
	tempCounter := 0
	desugared := &ast.Program{
		Statements: make([]ast.Statement, 0, len(program.Statements)),
	}

	for _, stmt := range program.Statements {
		desugaredStmt, newCounter := desugarStatement(stmt, tempCounter)
		tempCounter = newCounter
		if desugaredStmt != nil {
			desugared.Statements = append(desugared.Statements, desugaredStmt)
		}
	}

	return desugared
}

// desugarStatement 脱糖单个语句
func desugarStatement(stmt ast.Statement, tempCounter int) (ast.Statement, int) {
	if stmt == nil {
		return nil, tempCounter
	}
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		if s.Expression == nil {
			return nil, tempCounter
		}
		desugaredExpr := desugarExpression(s.Expression)
		if desugaredExpr == nil {
			return nil, tempCounter
		}
		return &ast.ExpressionStatement{
			Token:      s.Token,
			Expression: desugaredExpr,
		}, tempCounter
	case *ast.LetStatement:
		return &ast.LetStatement{
			Token: s.Token,
			Name:  s.Name,
			Value: desugarExpression(s.Value),
		}, tempCounter
	case *ast.AssignStatement:
		// Check for tuple unpacking (left side is a list literal)
		if listLit, ok := s.Left.(*ast.ListLiteral); ok && len(listLit.Elements) > 1 {
			// Generate unique temp variable name
			tempName := fmt.Sprintf("_unpack_temp_%d", tempCounter)
			tempCounter++
			tempVar := &ast.Identifier{Token: tempName, Value: tempName}
			
			// Create statements: _unpack_temp = value
			tempAssign := &ast.AssignStatement{
				Token: "=",
				Left:  tempVar,
				Value: desugarExpression(s.Value),
			}
			
			// Create statements for each element: a = _unpack_temp[0], b = _unpack_temp[1], etc.
			stmts := []ast.Statement{tempAssign}
			
			for i, elem := range listLit.Elements {
				stmts = append(stmts, &ast.AssignStatement{
					Token: "=",
					Left:  desugarExpression(elem),
					Value: &ast.IndexExpression{
						Token: "[",
						Left:  tempVar,
						Index: &ast.IntegerLiteral{Token: fmt.Sprintf("%d", i), Value: int64(i)},
					},
				})
			}
			
			// Now desugar the statements inside this block!
			desugaredStmts := make([]ast.Statement, 0, len(stmts))
			for _, subStmt := range stmts {
				desugaredSubStmt, newCounter := desugarStatement(subStmt, tempCounter)
				tempCounter = newCounter
				if desugaredSubStmt != nil {
					desugaredStmts = append(desugaredStmts, desugaredSubStmt)
				}
			}
			
			return &ast.BlockStatement{
				Token:      s.Token,
				Statements: desugaredStmts,
			}, tempCounter
		}
		
		// Regular assignment
		return &ast.AssignStatement{
			Token: s.Token,
			Left:  desugarExpression(s.Left),
			Value: desugarExpression(s.Value),
		}, tempCounter
	case *ast.AugAssignStatement:
		// 将增强赋值转换为: temp = left; left = temp op value
		// 这样可以避免在编译右侧表达式时找不到左侧变量的问题
		origLeft := s.Left
		tempName := fmt.Sprintf("_aug_%d", tempCounter)
		tempCounter++
		tempVar := &ast.Identifier{Token: tempName, Value: tempName}
		infixExpr := &ast.InfixExpression{
			Token:    s.Token,
			Left:     tempVar,
			Operator: s.Operator,
			Right:    desugarExpression(s.Value),
		}
		stmts := []ast.Statement{
			&ast.AssignStatement{
				Token: "=",
				Left:  tempVar,
				Value: desugarExpression(origLeft),
			},
			&ast.AssignStatement{
				Token: s.Token,
				Left:  desugarExpression(origLeft),
				Value: infixExpr,
			},
		}
		// Desugar these statements
		desugaredStmts := make([]ast.Statement, 0, len(stmts))
		for _, subStmt := range stmts {
			desugaredSubStmt, newCounter := desugarStatement(subStmt, tempCounter)
			tempCounter = newCounter
			if desugaredSubStmt != nil {
				desugaredStmts = append(desugaredStmts, desugaredSubStmt)
			}
		}
		return &ast.BlockStatement{
			Token: s.Token,
			Statements: desugaredStmts,
		}, tempCounter
	case *ast.ReturnStatement:
		return &ast.ReturnStatement{
			Token:       s.Token,
			ReturnValue: desugarExpression(s.ReturnValue),
		}, tempCounter
	case *ast.BlockStatement:
		desugaredBlock, newCounter := desugarBlockStatement(s, tempCounter)
		return desugaredBlock, newCounter
	case *ast.WhileStatement:
		desugaredBody, newCounter := desugarBlockStatement(s.Body, tempCounter)
		return &ast.WhileStatement{
			Token:     s.Token,
			Condition: desugarExpression(s.Condition),
			Body:      desugaredBody,
		}, newCounter
	case *ast.ForStatement:
		desugaredFor, newCounter := desugarForToWhile(s, tempCounter)
		return desugaredFor, newCounter
	case *ast.BreakStatement:
		return nil, tempCounter
	case *ast.ContinueStatement:
		return nil, tempCounter
	case *ast.TryStatement:
		// 对 try 语句进行脱糖处理：脱糖 body、excepts 和 finally
		desugaredBody, newCounter := desugarBlockStatement(s.Body, tempCounter)
		tempCounter = newCounter
		desugaredTry := &ast.TryStatement{
			Token: s.Token,
			Body:  desugaredBody,
		}
		desugaredTry.Excepts = make([]*ast.ExceptClause, 0, len(s.Excepts))
		for _, ex := range s.Excepts {
			desugaredExceptBody, newCounter2 := desugarBlockStatement(ex.Body, tempCounter)
			tempCounter = newCounter2
			desugaredExcept := &ast.ExceptClause{
				Token: ex.Token,
				Type:  desugarExpression(ex.Type),
				Name:  ex.Name,
				Body:  desugaredExceptBody,
			}
			desugaredTry.Excepts = append(desugaredTry.Excepts, desugaredExcept)
		}
		if s.Finally != nil {
			desugaredFinally, newCounter3 := desugarBlockStatement(s.Finally, tempCounter)
			tempCounter = newCounter3
			desugaredTry.Finally = desugaredFinally
		}
		return desugaredTry, tempCounter
	case *ast.RaiseStatement:
		// 对 raise 语句进行脱糖处理：脱糖表达式
		return &ast.RaiseStatement{
			Token:      s.Token,
			Expression: desugarExpression(s.Expression),
		}, tempCounter
	case *ast.WithStatement:
		// 对 with 语句进行脱糖处理：脱糖表达式和 body
		desugaredBody, newCounter := desugarBlockStatement(s.Body, tempCounter)
		return &ast.WithStatement{
			Token: s.Token,
			Expr:  desugarExpression(s.Expr),
			Name:  s.Name,
			Body:  desugaredBody,
		}, newCounter
	case *ast.YieldStatement:
		// 对 yield 语句进行脱糖处理：脱糖表达式
		return &ast.YieldStatement{
			Token:      s.Token,
			Expression: desugarExpression(s.Expression),
		}, tempCounter
	case *ast.ClassStatement:
		desugaredBody, newCounter := desugarBlockStatement(s.Body, tempCounter)
		return &ast.ClassStatement{
			Token:       s.Token,
			Name:        s.Name,
			SuperClass:  s.SuperClass,
			Body:        desugaredBody,
			Methods:     s.Methods,
		}, newCounter
	default:
		return stmt, tempCounter
	}
}

// desugarBlockStatement 脱糖块语句
func desugarBlockStatement(block *ast.BlockStatement, tempCounter int) (*ast.BlockStatement, int) {
	if block == nil {
		return nil, tempCounter
	}
	desugared := &ast.BlockStatement{
		Token:      block.Token,
		Statements: make([]ast.Statement, 0, len(block.Statements)),
	}

	for _, stmt := range block.Statements {
		desugaredStmt, newCounter := desugarStatement(stmt, tempCounter)
		tempCounter = newCounter
		if desugaredStmt != nil {
			desugared.Statements = append(desugared.Statements, desugaredStmt)
		}
	}

	return desugared, tempCounter
}

// desugarExpression 脱糖单个表达式
func desugarExpression(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.PrefixExpression:
		// 检查是否是位取反，转换为内置函数调用
		if isBitNotOp(e.Operator) {
			right := desugarExpression(e.Right)
			return &ast.CallExpression{
				Token: "__bitnot__",
				Function: &ast.Identifier{Token: "__bitnot__", Value: "__bitnot__"},
				Arguments: []ast.Expression{right},
			}
		}
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

		// 检查是否是位运算，转换为内置函数调用
		if isBitOp(e.Operator) {
			left := desugarExpression(e.Left)
			right := desugarExpression(e.Right)
			var builtinName string
			switch e.Operator {
			case "|":
				builtinName = "__bitor__"
			case "&":
				builtinName = "__bitand__"
			case "^":
				builtinName = "__bitxor__"
			case "<<":
				builtinName = "__lshift__"
			case ">>":
				builtinName = "__rshift__"
			}
			return &ast.CallExpression{
				Token: builtinName,
				Function: &ast.Identifier{Token: builtinName, Value: builtinName},
				Arguments: []ast.Expression{left, right},
			}
		}

		// 检查是否是身份运算符，转换为 id() 比较
		if isIdentityOp(e.Operator) {
			left := desugarExpression(e.Left)
			right := desugarExpression(e.Right)
			
			leftId := &ast.CallExpression{
				Token: "id",
				Function: &ast.Identifier{Token: "id", Value: "id"},
				Arguments: []ast.Expression{left},
			}
			rightId := &ast.CallExpression{
				Token: "id",
				Function: &ast.Identifier{Token: "id", Value: "id"},
				Arguments: []ast.Expression{right},
			}
			
			op := "=="
			if e.Operator == "is not" {
				op = "!="
			}
			
			return &ast.InfixExpression{
				Token:    op,
				Left:     leftId,
				Operator: op,
				Right:    rightId,
			}
		}

		// 检查是否是成员运算符，转换为 __contains__ 调用
		if isMembershipOp(e.Operator) {
			left := desugarExpression(e.Left)
			right := desugarExpression(e.Right)
			
			// 构建 __contains__(right, left) 调用
			containsCall := &ast.CallExpression{
				Token: "__contains__",
				Function: &ast.Identifier{Token: "__contains__", Value: "__contains__"},
				Arguments: []ast.Expression{right, left},
			}
			
			// 如果是 "not in"，添加 not 前缀
			if e.Operator == "not in" {
				return &ast.PrefixExpression{
					Token:    "not",
					Operator: "not",
					Right:    containsCall,
				}
			}
			
			return containsCall
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
		desugaredConsequence, _ := desugarBlockStatement(e.Consequence, 0)
		desugared := &ast.IfExpression{
			Token:       e.Token,
			Condition:   desugarExpression(e.Condition),
			Consequence: desugaredConsequence,
		}
		if e.Alternative != nil {
			desugaredAlternative, _ := desugarBlockStatement(e.Alternative, 0)
			desugared.Alternative = desugaredAlternative
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
		desugaredBody, _ := desugarBlockStatement(e.Body, 0)
		return &ast.FunctionLiteral{
			Token:      e.Token,
			Name:       e.Name,
			Parameters: e.Parameters,
			Body:       desugaredBody,
		}
	case *ast.LambdaExpression:
		return &ast.LambdaExpression{
			Token:     e.Token,
			Parameters: e.Parameters,
			Body:      desugarExpression(e.Body),
		}
	case *ast.GeneratorExpression:
		return desugarGeneratorExpression(e)
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

func isBitOp(op string) bool {
	return op == "|" || op == "&" || op == "^" || op == "<<" || op == ">>"
}

func isBitNotOp(op string) bool {
	return op == "~"
}

func isIdentityOp(op string) bool {
	return op == "is" || op == "is not"
}

func isMembershipOp(op string) bool {
	return op == "in" || op == "not in"
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

func desugarForToWhile(forStmt *ast.ForStatement, tempCounter int) (ast.Statement, int) {
	// First check if iterable is a call to range
	if callExpr, ok := forStmt.Iterable.(*ast.CallExpression); ok {
		if ident, ok := callExpr.Function.(*ast.Identifier); ok && ident.Value == "range" {
			// We found a for loop over range, optimize it!
			var start, end, step ast.Expression

			switch len(callExpr.Arguments) {
			case 1: // range(end)
				start = &ast.IntegerLiteral{Token: "0", Value: 0}
				end = desugarExpression(callExpr.Arguments[0])
				step = &ast.IntegerLiteral{Token: "1", Value: 1}
			case 2: // range(start, end)
				start = desugarExpression(callExpr.Arguments[0])
				end = desugarExpression(callExpr.Arguments[1])
				step = &ast.IntegerLiteral{Token: "1", Value: 1}
			case 3: // range(start, end, step)
				start = desugarExpression(callExpr.Arguments[0])
				end = desugarExpression(callExpr.Arguments[1])
				step = desugarExpression(callExpr.Arguments[2])
			default:
				// fallback to original behavior
			}

			if start != nil {
				indexVar := &ast.Identifier{Token: "_i", Value: "_i"}

				var condition ast.Expression
				// check if step is positive (1) or negative
				condition = &ast.InfixExpression{
					Token:    "<",
					Left:     indexVar,
					Operator: "<",
					Right:    end,
				}

				bodyStmts := []ast.Statement{
					&ast.AssignStatement{
						Token: "=",
						Left:  forStmt.Value,
						Value: indexVar,
					},
				}

				desugaredBody, newCounter1 := desugarBlockStatement(forStmt.Body, tempCounter)
				tempCounter = newCounter1
				bodyStmts = append(bodyStmts, desugaredBody.Statements...)

				bodyStmts = append(bodyStmts, &ast.AssignStatement{
					Token: "=",
					Left:  indexVar,
					Value: &ast.InfixExpression{
						Token:    "+",
						Left:     indexVar,
						Operator: "+",
						Right:    step,
					},
				})

				body := &ast.BlockStatement{
					Token:      forStmt.Token,
					Statements: bodyStmts,
				}

				whileStmt := &ast.WhileStatement{
					Token:     forStmt.Token,
					Condition: condition,
					Body:      body,
				}

				return &ast.BlockStatement{
					Token: forStmt.Token,
					Statements: []ast.Statement{
						&ast.AssignStatement{
							Token: "=",
							Left:  indexVar,
							Value: start,
						},
						whileStmt,
					},
				}, tempCounter
			}
		}
	}

	// Fallback to original behavior for non-range loops
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
			Left:  forStmt.Value,
			Value: &ast.IndexExpression{
				Token: "[",
				Left:  iterable,
				Index: indexVar,
			},
		},
	}

	desugaredBody, newCounter2 := desugarBlockStatement(forStmt.Body, tempCounter)
	tempCounter = newCounter2
	bodyStmts = append(bodyStmts, desugaredBody.Statements...)

	bodyStmts = append(bodyStmts, &ast.AssignStatement{
		Token: "=",
		Left:  indexVar,
		Value: &ast.InfixExpression{
			Token:    "+",
			Left:     indexVar,
			Operator: "+",
			Right:    &ast.IntegerLiteral{Token: "1", Value: 1},
		},
	})

	body := &ast.BlockStatement{
		Token:      forStmt.Token,
		Statements: bodyStmts,
	}

	whileStmt := &ast.WhileStatement{
		Token:     forStmt.Token,
		Condition: condition,
		Body:      body,
	}

	return &ast.BlockStatement{
		Token: forStmt.Token,
		Statements: []ast.Statement{
			&ast.AssignStatement{
				Token: "=",
				Left:  indexVar,
				Value: &ast.IntegerLiteral{Token: "0", Value: 0},
			},
			whileStmt,
		},
	}, tempCounter
}

func desugarGeneratorExpression(gen *ast.GeneratorExpression) ast.Expression {
	iterable := desugarExpression(gen.Iterable)
	element := desugarExpression(gen.Element)

	paramName := fmt.Sprintf("_gen_iter_%d", tempGenCounter)
	tempGenCounter++

	yieldStmt := &ast.YieldStatement{
		Token:      "yield",
		Expression: element,
	}

	forBody := &ast.BlockStatement{
		Token: "for",
		Statements: []ast.Statement{
			yieldStmt,
		},
	}

	forStmt := &ast.ForStatement{
		Token:    "for",
		Value:    gen.Variable,
		Iterable: &ast.Identifier{Token: paramName, Value: paramName},
		Body:     forBody,
	}

	desugaredForStmt, _ := desugarForToWhile(forStmt, tempGenCounter)

	funcLit := &ast.FunctionLiteral{
		Token: "def",
		Name:  "__gen__",
		Parameters: []*ast.Identifier{
			{Token: paramName, Value: paramName},
		},
		Body: &ast.BlockStatement{
			Token:      "def",
			Statements: []ast.Statement{desugaredForStmt},
		},
	}

	return &ast.CallExpression{
		Token:    "call",
		Function: funcLit,
		Arguments: []ast.Expression{iterable},
	}
}

var tempGenCounter int = 0
