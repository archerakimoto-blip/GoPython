package desugar

import (
	"github.com/go-py/go-python/pkg/ast"
)

var walrusStmts []ast.Statement

func collectWalrusStmts() []ast.Statement {
	stmts := walrusStmts
	walrusStmts = nil
	return stmts
}

func Desugar(program *ast.Program) *ast.Program {
	walrusStmts = nil
	desugared := &ast.Program{
		Statements: make([]ast.Statement, 0, len(program.Statements)),
	}

	for _, stmt := range program.Statements {
		desugaredStmt := desugarStatement(stmt)
		collected := collectWalrusStmts()
		if len(collected) > 0 {
			desugared.Statements = append(desugared.Statements, collected...)
		}
		if desugaredStmt != nil {
			desugared.Statements = append(desugared.Statements, desugaredStmt)
		}
	}

	return desugared
}

// desugarStatement 脱糖单个语句
type propDetail struct {
	getter *ast.FunctionLiteral
	setter *ast.FunctionLiteral
	deleter *ast.FunctionLiteral
}

func desugarClassStatement(s *ast.ClassStatement) ast.Statement {
	properties := make(map[string]*propDetail)
	classmethods := make(map[string]*ast.FunctionLiteral)
	staticmethods := make(map[string]*ast.FunctionLiteral)
	var slots []string
	var slotsAssignStmt ast.Statement
	var existingGetattr *ast.FunctionLiteral
	var existingSetattr *ast.FunctionLiteral
	var existingDelattr *ast.FunctionLiteral

	for _, method := range s.Methods {
		for _, dec := range method.Decorators {
			if ident, ok := dec.(*ast.Identifier); ok {
				if ident.Value == "property" {
					if properties[method.Name] == nil {
						properties[method.Name] = &propDetail{}
					}
					properties[method.Name].getter = method
					break
				}
				if ident.Value == "classmethod" {
					classmethods[method.Name] = method
					break
				}
				if ident.Value == "staticmethod" {
					staticmethods[method.Name] = method
					break
				}
			}
			if ma, ok := dec.(*ast.MemberAccess); ok {
				propName := ""
				if objIdent, ok := ma.Object.(*ast.Identifier); ok {
					propName = objIdent.Value
				}
				if propName != "" && ma.Member.Value == "setter" {
					if properties[propName] == nil {
						properties[propName] = &propDetail{}
					}
					properties[propName].setter = method
					break
				}
				if propName != "" && ma.Member.Value == "deleter" {
					if properties[propName] == nil {
						properties[propName] = &propDetail{}
					}
					properties[propName].deleter = method
					break
				}
			}
		}

		switch method.Name {
		case "__getattr__":
			existingGetattr = method
		case "__setattr__":
			existingSetattr = method
		case "__delattr__":
			existingDelattr = method
		}
	}

	if s.Body != nil {
		for _, stmt := range s.Body.Statements {
			if assign, ok := stmt.(*ast.AssignStatement); ok {
				if len(assign.Names) == 1 && assign.Names[0].Value == "__slots__" {
					if listLit, ok := assign.Value.(*ast.ListLiteral); ok {
						for _, elem := range listLit.Elements {
							if str, ok := elem.(*ast.StringLiteral); ok {
								slots = append(slots, str.Value)
							}
						}
						slotsAssignStmt = stmt
					}
				}
			}
		}
	}

	if len(properties) == 0 && len(classmethods) == 0 && len(staticmethods) == 0 && len(slots) == 0 {
		return &ast.ClassStatement{
			Token:        s.Token,
			Name:         s.Name,
			SuperClass:   s.SuperClass,
			Body:         desugarBlockStatement(s.Body),
			Methods:      s.Methods,
		}
	}

	decoratedMethods := make(map[*ast.FunctionLiteral]string)

	for propName, detail := range properties {
		if detail.getter != nil {
			decoratedMethods[detail.getter] = "_desugar_prop_get_" + propName
		}
		if detail.setter != nil {
			decoratedMethods[detail.setter] = "_desugar_prop_set_" + propName
		}
		if detail.deleter != nil {
			decoratedMethods[detail.deleter] = "_desugar_prop_del_" + propName
		}
	}
	for cmName, method := range classmethods {
		decoratedMethods[method] = "_desugar_cm_" + cmName
	}
	for smName, method := range staticmethods {
		decoratedMethods[method] = "_desugar_sm_" + smName
	}

	newMethods := []*ast.FunctionLiteral{}
	for _, method := range s.Methods {
		if method == existingGetattr || method == existingSetattr || method == existingDelattr {
			continue
		}
		if mangledName, ok := decoratedMethods[method]; ok {
			newBody := desugarBlockStatement(method.Body)
			newMethods = append(newMethods, &ast.FunctionLiteral{
				Token:      method.Token,
				Name:       mangledName,
				Parameters: method.Parameters,
				Body:       newBody,
				VarArgs:    method.VarArgs,
				KwArgs:     method.KwArgs,
				IsAsync:    method.IsAsync,
			})
		} else {
			newMethods = append(newMethods, method)
		}
	}

	needGetattr := len(properties) > 0 || len(classmethods) > 0 || len(staticmethods) > 0
	hasPropertySetters := false
	for _, detail := range properties {
		if detail.setter != nil {
			hasPropertySetters = true
			break
		}
	}
	hasPropertyDeleters := false
	for _, detail := range properties {
		if detail.deleter != nil {
			hasPropertyDeleters = true
			break
		}
	}
	needSetattr := hasPropertySetters || len(slots) > 0
	needDelattr := hasPropertyDeleters

	if needGetattr {
		getattrStmts := []ast.Statement{}

		propNames := make([]string, 0, len(properties))
		for name := range properties {
			propNames = append(propNames, name)
		}
		for _, propName := range propNames {
			detail := properties[propName]
			if detail.getter != nil {
				condition := &ast.InfixExpression{
					Token:    "==",
					Left:     &ast.Identifier{Token: "name", Value: "name"},
					Operator: "==",
					Right:    &ast.StringLiteral{Token: propName, Value: propName},
				}
				methodCall := &ast.MethodCall{
					Token:     ".",
					Object:    &ast.Identifier{Token: "self", Value: "self"},
					Method:    &ast.Identifier{Token: "_desugar_prop_get_" + propName, Value: "_desugar_prop_get_" + propName},
					Arguments: []ast.Expression{},
				}
				returnStmt := &ast.ReturnStatement{
					Token:       "return",
					ReturnValue: methodCall,
				}
				getattrStmts = append(getattrStmts, &ast.ExpressionStatement{
					Token: "if",
					Expression: &ast.IfExpression{
						Token:       "if",
						Condition:   condition,
						Consequence: &ast.BlockStatement{Token: "if", Statements: []ast.Statement{returnStmt}},
						Alternative: nil,
					},
				})
			}
		}

		cmNames := make([]string, 0, len(classmethods))
		for name := range classmethods {
			cmNames = append(cmNames, name)
		}
		for _, cmName := range cmNames {
			condition := &ast.InfixExpression{
				Token:    "==",
				Left:     &ast.Identifier{Token: "name", Value: "name"},
				Operator: "==",
				Right:    &ast.StringLiteral{Token: cmName, Value: cmName},
			}
			getClassCall := &ast.CallExpression{
				Token:     "(",
				Function:  &ast.Identifier{Token: "__get_class__", Value: "__get_class__"},
				Arguments: []ast.Expression{&ast.Identifier{Token: "self", Value: "self"}},
			}
			methodAccess := &ast.MemberAccess{
				Token:  ".",
				Object: getClassCall,
				Member: &ast.Identifier{Token: "_desugar_cm_" + cmName, Value: "_desugar_cm_" + cmName},
			}
			getClassCall2 := &ast.CallExpression{
				Token:     "(",
				Function:  &ast.Identifier{Token: "__get_class__", Value: "__get_class__"},
				Arguments: []ast.Expression{&ast.Identifier{Token: "self", Value: "self"}},
			}
			bindMethodCall := &ast.CallExpression{
				Token:     "(",
				Function:  &ast.Identifier{Token: "__bind_method__", Value: "__bind_method__"},
				Arguments: []ast.Expression{methodAccess, getClassCall2},
			}
			returnStmt := &ast.ReturnStatement{
				Token:       "return",
				ReturnValue: bindMethodCall,
			}
			getattrStmts = append(getattrStmts, &ast.ExpressionStatement{
				Token: "if",
				Expression: &ast.IfExpression{
					Token:       "if",
					Condition:   condition,
					Consequence: &ast.BlockStatement{Token: "if", Statements: []ast.Statement{returnStmt}},
					Alternative: nil,
				},
			})
		}

		smNames := make([]string, 0, len(staticmethods))
		for name := range staticmethods {
			smNames = append(smNames, name)
		}
		for _, smName := range smNames {
			condition := &ast.InfixExpression{
				Token:    "==",
				Left:     &ast.Identifier{Token: "name", Value: "name"},
				Operator: "==",
				Right:    &ast.StringLiteral{Token: smName, Value: smName},
			}
			getClassCall := &ast.CallExpression{
				Token:     "(",
				Function:  &ast.Identifier{Token: "__get_class__", Value: "__get_class__"},
				Arguments: []ast.Expression{&ast.Identifier{Token: "self", Value: "self"}},
			}
			methodAccess := &ast.MemberAccess{
				Token:  ".",
				Object: getClassCall,
				Member: &ast.Identifier{Token: "_desugar_sm_" + smName, Value: "_desugar_sm_" + smName},
			}
			returnStmt := &ast.ReturnStatement{
				Token:       "return",
				ReturnValue: methodAccess,
			}
			getattrStmts = append(getattrStmts, &ast.ExpressionStatement{
				Token: "if",
				Expression: &ast.IfExpression{
					Token:       "if",
					Condition:   condition,
					Consequence: &ast.BlockStatement{Token: "if", Statements: []ast.Statement{returnStmt}},
					Alternative: nil,
				},
			})
		}

		if existingGetattr != nil {
			desugaredBody := desugarBlockStatement(existingGetattr.Body)
			getattrStmts = append(getattrStmts, desugaredBody.Statements...)
			newMethods = append(newMethods, &ast.FunctionLiteral{
				Token:      existingGetattr.Token,
				Name:       "__getattr__",
				Parameters: existingGetattr.Parameters,
				Body:       &ast.BlockStatement{Token: "if", Statements: getattrStmts},
				VarArgs:    existingGetattr.VarArgs,
				KwArgs:     existingGetattr.KwArgs,
			})
		} else {
			newMethods = append(newMethods, &ast.FunctionLiteral{
				Token:      "def",
				Name:       "__getattr__",
				Parameters: []*ast.Identifier{{Token: "self", Value: "self"}, {Token: "name", Value: "name"}},
				Body:       &ast.BlockStatement{Token: ":", Statements: getattrStmts},
			})
		}
	}

	if needSetattr {
		setattrStmts := []ast.Statement{}

		for propName, detail := range properties {
		if detail.setter != nil {
			condition := &ast.InfixExpression{
				Token:    "==",
				Left:     &ast.Identifier{Token: "name", Value: "name"},
				Operator: "==",
				Right:    &ast.StringLiteral{Token: propName, Value: propName},
			}
			methodCall := &ast.MethodCall{
				Token:     ".",
				Object:    &ast.Identifier{Token: "self", Value: "self"},
				Method:    &ast.Identifier{Token: "_desugar_prop_set_" + propName, Value: "_desugar_prop_set_" + propName},
				Arguments: []ast.Expression{&ast.Identifier{Token: "value", Value: "value"}},
			}
			callStmt := &ast.ExpressionStatement{
				Token:      "(",
				Expression: methodCall,
			}
			returnStmt := &ast.ReturnStatement{
				Token:       "return",
				ReturnValue: &ast.IntegerLiteral{Token: "0", Value: 0},
			}
			setattrStmts = append(setattrStmts, &ast.ExpressionStatement{
				Token: "if",
				Expression: &ast.IfExpression{
					Token:       "if",
					Condition:   condition,
					Consequence: &ast.BlockStatement{Token: "if", Statements: []ast.Statement{callStmt, returnStmt}},
					Alternative: nil,
				},
			})
		}
	}

	if len(slots) > 0 {
		for _, slotName := range slots {
			condition := &ast.InfixExpression{
				Token:    "==",
				Left:     &ast.Identifier{Token: "name", Value: "name"},
				Operator: "==",
				Right:    &ast.StringLiteral{Token: slotName, Value: slotName},
			}
			setFieldCall := &ast.CallExpression{
				Token:     "(",
				Function:  &ast.Identifier{Token: "__set_field__", Value: "__set_field__"},
				Arguments: []ast.Expression{
					&ast.Identifier{Token: "self", Value: "self"},
					&ast.Identifier{Token: "name", Value: "name"},
					&ast.Identifier{Token: "value", Value: "value"},
				},
			}
			callStmt := &ast.ExpressionStatement{
				Token:      "(",
				Expression: setFieldCall,
			}
			returnStmt := &ast.ReturnStatement{
				Token:       "return",
				ReturnValue: &ast.IntegerLiteral{Token: "0", Value: 0},
			}
			setattrStmts = append(setattrStmts, &ast.ExpressionStatement{
				Token: "if",
				Expression: &ast.IfExpression{
					Token:       "if",
					Condition:   condition,
					Consequence: &ast.BlockStatement{Token: "if", Statements: []ast.Statement{callStmt, returnStmt}},
					Alternative: nil,
				},
			})
		}

		errorMsg := &ast.InfixExpression{
			Token: "+",
			Left: &ast.InfixExpression{
				Token:    "+",
				Left:     &ast.StringLiteral{Token: "", Value: "AttributeError: '" + s.Name.Value + "' object has no attribute '"},
				Operator: "+",
				Right:    &ast.Identifier{Token: "name", Value: "name"},
			},
			Operator: "+",
			Right:    &ast.StringLiteral{Token: "", Value: "'"},
		}
		setattrStmts = append(setattrStmts, &ast.RaiseStatement{
			Token:      "raise",
			Expression: errorMsg,
		})
	} else {
		setFieldCall := &ast.CallExpression{
			Token:     "(",
			Function:  &ast.Identifier{Token: "__set_field__", Value: "__set_field__"},
			Arguments: []ast.Expression{
				&ast.Identifier{Token: "self", Value: "self"},
				&ast.Identifier{Token: "name", Value: "name"},
				&ast.Identifier{Token: "value", Value: "value"},
			},
		}
		setattrStmts = append(setattrStmts, &ast.ExpressionStatement{
			Token:      "(",
			Expression: setFieldCall,
		})
	}

	if existingSetattr != nil {
		desugaredBody := desugarBlockStatement(existingSetattr.Body)
		setattrStmts = append(setattrStmts, desugaredBody.Statements...)
		newMethods = append(newMethods, &ast.FunctionLiteral{
			Token:      existingSetattr.Token,
			Name:       "__setattr__",
			Parameters: existingSetattr.Parameters,
			Body:       &ast.BlockStatement{Token: "if", Statements: setattrStmts},
			VarArgs:    existingSetattr.VarArgs,
			KwArgs:     existingSetattr.KwArgs,
		})
	} else {
		newMethods = append(newMethods, &ast.FunctionLiteral{
			Token:      "def",
			Name:       "__setattr__",
			Parameters: []*ast.Identifier{
				{Token: "self", Value: "self"},
				{Token: "name", Value: "name"},
				{Token: "value", Value: "value"},
			},
			Body: &ast.BlockStatement{Token: ":", Statements: setattrStmts},
		})
	}
	}

	if needDelattr {
		delattrStmts := []ast.Statement{}

		for propName, detail := range properties {
		if detail.deleter != nil {
			condition := &ast.InfixExpression{
				Token:    "==",
				Left:     &ast.Identifier{Token: "name", Value: "name"},
				Operator: "==",
				Right:    &ast.StringLiteral{Token: propName, Value: propName},
			}
			methodCall := &ast.MethodCall{
				Token:     ".",
				Object:    &ast.Identifier{Token: "self", Value: "self"},
				Method:    &ast.Identifier{Token: "_desugar_prop_del_" + propName, Value: "_desugar_prop_del_" + propName},
				Arguments: []ast.Expression{},
			}
			callStmt := &ast.ExpressionStatement{
				Token:      "(",
				Expression: methodCall,
			}
			returnStmt := &ast.ReturnStatement{
				Token:       "return",
				ReturnValue: &ast.IntegerLiteral{Token: "0", Value: 0},
			}
			delattrStmts = append(delattrStmts, &ast.ExpressionStatement{
				Token: "if",
				Expression: &ast.IfExpression{
					Token:       "if",
					Condition:   condition,
					Consequence: &ast.BlockStatement{Token: "if", Statements: []ast.Statement{callStmt, returnStmt}},
					Alternative: nil,
				},
			})
		}
	}

	delFieldCall := &ast.CallExpression{
		Token:     "(",
		Function:  &ast.Identifier{Token: "__del_field__", Value: "__del_field__"},
		Arguments: []ast.Expression{
			&ast.Identifier{Token: "self", Value: "self"},
			&ast.Identifier{Token: "name", Value: "name"},
		},
	}
	delattrStmts = append(delattrStmts, &ast.ExpressionStatement{
		Token:      "(",
		Expression: delFieldCall,
	})

	if existingDelattr != nil {
		desugaredBody := desugarBlockStatement(existingDelattr.Body)
		delattrStmts = append(delattrStmts, desugaredBody.Statements...)
		newMethods = append(newMethods, &ast.FunctionLiteral{
			Token:      existingDelattr.Token,
			Name:       "__delattr__",
			Parameters: existingDelattr.Parameters,
			Body:       &ast.BlockStatement{Token: "if", Statements: delattrStmts},
			VarArgs:    existingDelattr.VarArgs,
			KwArgs:     existingDelattr.KwArgs,
		})
	} else {
		newMethods = append(newMethods, &ast.FunctionLiteral{
			Token:      "def",
			Name:       "__delattr__",
			Parameters: []*ast.Identifier{
				{Token: "self", Value: "self"},
				{Token: "name", Value: "name"},
			},
			Body: &ast.BlockStatement{Token: ":", Statements: delattrStmts},
		})
	}
	}

	var newBodyStmts []ast.Statement
	if s.Body != nil {
		for _, stmt := range s.Body.Statements {
			if stmt == slotsAssignStmt {
				continue
			}
			desugaredStmt := desugarStatement(stmt)
			newBodyStmts = append(newBodyStmts, desugaredStmt)
		}
	}

	return &ast.ClassStatement{
		Token:        s.Token,
		Name:         s.Name,
		SuperClass:   s.SuperClass,
		Body:         &ast.BlockStatement{Token: s.Token, Statements: newBodyStmts},
		Methods:      newMethods,
	}
}

func desugarStatement(stmt ast.Statement) ast.Statement {
	if stmt == nil {
		return nil
	}
	switch s := stmt.(type) {
	case *ast.ClassStatement:
		return desugarClassStatement(s)
	case *ast.ExpressionStatement:
		if s == nil || s.Expression == nil {
			return nil
		}
		if walrus, ok := s.Expression.(*ast.WalrusExpression); ok {
			return &ast.LetStatement{
				Token: ":=",
				Names: []*ast.Identifier{walrus.Name},
				Value: desugarExpression(walrus.Value),
			}
		}
		if fnLit, ok := s.Expression.(*ast.FunctionLiteral); ok && len(fnLit.Decorators) > 0 {
			desugaredFn := desugarExpression(fnLit).(*ast.FunctionLiteral)
			desugaredDecorators := make([]ast.Expression, len(desugaredFn.Decorators))
			for i, dec := range desugaredFn.Decorators {
				desugaredDecorators[i] = desugarExpression(dec)
			}

			stmts := []ast.Statement{}

			funcIdent := &ast.Identifier{Token: desugaredFn.Token, Value: desugaredFn.Name}
			tempIdent := &ast.Identifier{Token: desugaredFn.Token, Value: "_temp_" + desugaredFn.Name}

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
		condition := desugarExpression(s.Condition)
		collected := collectWalrusStmts()
		body := desugarBlockStatement(s.Body)
		if len(collected) > 0 && body != nil {
			body.Statements = append(collected, body.Statements...)
		}
		whileStmt := &ast.WhileStatement{
			Token:     s.Token,
			Condition: condition,
			Body:      body,
		}
		if len(collected) == 0 {
			return whileStmt
		}
		return &ast.BlockStatement{
			Token:      s.Token,
			Statements: append(collected, whileStmt),
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
	case *ast.MatchStatement:
		return desugarMatchStatement(s)

	case *ast.DelStatement:
		return &ast.DelStatement{
			Token:  s.Token,
			Target: desugarExpression(s.Target),
		}
	case *ast.AssertStatement:
		return desugarAssertStatement(s)
	case *ast.GlobalStatement:
		return &ast.GlobalStatement{
			Token: s.Token,
			Names: s.Names,
		}
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
		saved := walrusStmts
		walrusStmts = nil
		desugaredStmt := desugarStatement(stmt)
		collected := collectWalrusStmts()
		walrusStmts = saved
		if len(collected) > 0 {
			desugared.Statements = append(desugared.Statements, collected...)
		}
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
			IsAsync:    e.IsAsync,
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
	case *ast.FStringLiteral:
		desugaredParts := make([]ast.Expression, 0, len(e.Parts))
		for _, part := range e.Parts {
			desugaredParts = append(desugaredParts, desugarExpression(part))
		}
		return &ast.FStringLiteral{
			Token: e.Token,
			Parts: desugaredParts,
		}
	case *ast.WalrusExpression:
		desugaredValue := desugarExpression(e.Value)
		letStmt := &ast.LetStatement{
			Token: ":=",
			Names: []*ast.Identifier{e.Name},
			Value: desugaredValue,
		}
		walrusStmts = append(walrusStmts, letStmt)
		return &ast.Identifier{Token: e.Name.Token, Value: e.Name.Value}
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

func desugarMatchStatement(matchStmt *ast.MatchStatement) ast.Statement {
	// 首先，创建一个临时变量保存 subject
	tempIdent := &ast.Identifier{Token: "_match_subject", Value: "_match_subject"}
	letStmt := &ast.LetStatement{
		Token: matchStmt.Token,
		Names: []*ast.Identifier{tempIdent},
		Value: desugarExpression(matchStmt.Subject),
	}

	stmts := []ast.Statement{letStmt}

	// 现在，构建 if-elif 链
	var currentIf ast.Statement
	for i := len(matchStmt.Cases) - 1; i >= 0; i-- {
		caseClause := matchStmt.Cases[i]
		// 构建条件表达式
		condition, bindingStmts := buildPatternCondition(tempIdent, caseClause.Pattern)
		if caseClause.Guard != nil {
			condition = &ast.InfixExpression{
				Token:    "and",
				Left:     condition,
				Operator: "and",
				Right:    desugarExpression(caseClause.Guard),
			}
		}
		// 构建 body，包括绑定语句
		bodyStmts := make([]ast.Statement, 0, len(bindingStmts)+len(caseClause.Body.Statements))
		bodyStmts = append(bodyStmts, bindingStmts...)
		bodyStmts = append(bodyStmts, desugarBlockStatement(caseClause.Body).Statements...)
		bodyBlock := &ast.BlockStatement{
			Token:      caseClause.Token,
			Statements: bodyStmts,
		}

		// 构建 if 语句
		ifStmt := &ast.IfExpression{
			Token:       caseClause.Token,
			Condition:   condition,
			Consequence: bodyBlock,
		}
		if currentIf != nil {
			ifStmt.Alternative = &ast.BlockStatement{
				Token:      caseClause.Token,
				Statements: []ast.Statement{currentIf},
			}
		}
		currentIf = &ast.ExpressionStatement{
			Token:      caseClause.Token,
			Expression: ifStmt,
		}
	}

	if currentIf != nil {
		stmts = append(stmts, currentIf)
	}

	return &ast.BlockStatement{
		Token:      matchStmt.Token,
		Statements: stmts,
	}
}

func buildPatternCondition(subject *ast.Identifier, pattern ast.Pattern) (ast.Expression, []ast.Statement) {
	bindings := []ast.Statement{}

	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return &ast.Boolean{Token: "true", Value: true}, bindings
	case *ast.IdentifierPattern:
		// 绑定变量
		bindStmt := &ast.LetStatement{
			Token: p.Token,
			Names: []*ast.Identifier{p.Name},
			Value: subject,
		}
		bindings = append(bindings, bindStmt)
		return &ast.Boolean{Token: "true", Value: true}, bindings
	case *ast.LiteralPattern:
		// 相等比较
		return &ast.InfixExpression{
			Token:    "==",
			Left:     subject,
			Operator: "==",
			Right:    desugarExpression(p.Value),
		}, bindings
	case *ast.TuplePattern:
		// 检查长度，然后逐个匹配元素
		checkLen := &ast.InfixExpression{
			Token:    "==",
			Left: &ast.CallExpression{
				Token:    "len",
				Function: &ast.Identifier{Token: "len", Value: "len"},
				Arguments: []ast.Expression{subject},
			},
			Operator: "==",
			Right:    &ast.IntegerLiteral{Token: string(rune(len(p.Elements) + '0')), Value: int64(len(p.Elements))},
		}
		condition := checkLen
		for i, elemPattern := range p.Elements {
			elemSubject := &ast.IndexExpression{
				Token: "[",
				Left:  subject,
				Index: &ast.IntegerLiteral{Token: string(rune(i + '0')), Value: int64(i)},
			}
			// 使用临时变量保存元素，避免多次计算
			elemTemp := &ast.Identifier{Token: "_match_elem", Value: "_match_elem"}
			bindings = append(bindings, &ast.LetStatement{
				Token: "_match_elem",
				Names: []*ast.Identifier{elemTemp},
				Value: elemSubject,
			})
			elemCond, elemBindings := buildPatternCondition(elemTemp, elemPattern)
			bindings = append(bindings, elemBindings...)
			condition = &ast.InfixExpression{
				Token:    "and",
				Left:     condition,
				Operator: "and",
				Right:    elemCond,
			}
		}
		return condition, bindings
	case *ast.ListPattern:
		// 与 tuple 类似
		checkLen := &ast.InfixExpression{
			Token:    "==",
			Left: &ast.CallExpression{
				Token:    "len",
				Function: &ast.Identifier{Token: "len", Value: "len"},
				Arguments: []ast.Expression{subject},
			},
			Operator: "==",
			Right:    &ast.IntegerLiteral{Token: string(rune(len(p.Elements) + '0')), Value: int64(len(p.Elements))},
		}
		condition := checkLen
		for i, elemPattern := range p.Elements {
			elemSubject := &ast.IndexExpression{
				Token: "[",
				Left:  subject,
				Index: &ast.IntegerLiteral{Token: string(rune(i + '0')), Value: int64(i)},
			}
			elemTemp := &ast.Identifier{Token: "_match_elem", Value: "_match_elem"}
			bindings = append(bindings, &ast.LetStatement{
				Token: "_match_elem",
				Names: []*ast.Identifier{elemTemp},
				Value: elemSubject,
			})
			elemCond, elemBindings := buildPatternCondition(elemTemp, elemPattern)
			bindings = append(bindings, elemBindings...)
			condition = &ast.InfixExpression{
				Token:    "and",
				Left:     condition,
				Operator: "and",
				Right:    elemCond,
			}
		}
		return condition, bindings
	}
	return &ast.Boolean{Token: "true", Value: true}, bindings
}

func desugarAssertStatement(s *ast.AssertStatement) ast.Statement {
	notTest := &ast.PrefixExpression{
		Token:    "not",
		Operator: "not",
		Right:    desugarExpression(s.Test),
	}

	var raiseExpr ast.Expression
	if s.Message != nil {
		raiseExpr = &ast.InfixExpression{
			Token:    "+",
			Left:     &ast.StringLiteral{Token: "AssertionError", Value: "AssertionError: "},
			Operator: "+",
			Right:    desugarExpression(s.Message),
		}
	} else {
		raiseExpr = &ast.StringLiteral{Token: "AssertionError", Value: "AssertionError"}
	}

	raiseStmt := &ast.RaiseStatement{
		Token:      "raise",
		Expression: raiseExpr,
	}

	consequence := &ast.BlockStatement{
		Token: s.Token,
		Statements: []ast.Statement{
			raiseStmt,
		},
	}

	return &ast.ExpressionStatement{
		Token: s.Token,
		Expression: &ast.IfExpression{
			Token:       s.Token,
			Condition:   notTest,
			Consequence: consequence,
		},
	}
}
