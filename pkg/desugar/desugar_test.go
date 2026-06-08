package desugar

import (
	"testing"

	"github.com/go-py/go-python/pkg/ast"
)

// ===== Helper functions =====

func ident(name string) *ast.Identifier {
	return &ast.Identifier{Token: name, Value: name}
}

func intLit(val int64) *ast.IntegerLiteral {
	return &ast.IntegerLiteral{Token: "1", Value: val}
}

func strLit(val string) *ast.StringLiteral {
	return &ast.StringLiteral{Token: val, Value: val}
}

func boolLit(val bool) *ast.Boolean {
	return &ast.Boolean{Token: "True", Value: val}
}

func simpleProgram(stmts ...ast.Statement) *ast.Program {
	return &ast.Program{Statements: stmts}
}

// ===== 1. For→While loop transformation =====

func TestDesugarForToWhile(t *testing.T) {
	forStmt := &ast.ForStatement{
		Token:    "for",
		Value:    ident("x"),
		Iterable: ident("items"),
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "print", Expression: ident("x")}},
		},
	}

	result := desugarStatement(forStmt)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements in block, got %d", len(block.Statements))
	}

	// First statement should be LetStatement for index variable
	letStmt, ok := block.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", block.Statements[0])
	}
	if len(letStmt.Names) != 1 {
		t.Fatalf("expected 1 name in let, got %d", len(letStmt.Names))
	}
	if letStmt.Names[0].Value == "" {
		t.Fatal("index variable name should not be empty")
	}

	// Second statement should be WhileStatement
	whileStmt, ok := block.Statements[1].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", block.Statements[1])
	}

	// While body should contain: assign x = iterable[index], body, index += 1
	whileBlock := whileStmt.Body
	if len(whileBlock.Statements) < 3 {
		t.Fatalf("expected at least 3 statements in while body, got %d", len(whileBlock.Statements))
	}

	// First in while body: assign x = iterable[index]
	assignStmt, ok := whileBlock.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", whileBlock.Statements[0])
	}
	if assignStmt.Names[0].Value != "x" {
		t.Fatalf("expected variable 'x', got '%s'", assignStmt.Names[0].Value)
	}

	// Last in while body: index += 1 (stored as AssignStatement with "+=" token)
	lastStmt := whileBlock.Statements[len(whileBlock.Statements)-1]
	assignIncr, ok := lastStmt.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", lastStmt)
	}
	if assignIncr.Token != "+=" {
		t.Fatalf("expected '+=' token, got '%s'", assignIncr.Token)
	}
}

// ===== 2. Async for desugaring =====

func TestDesugarAsyncForStatement(t *testing.T) {
	stmt := &ast.AsyncForStatement{
		Token:    "async for",
		Value:    ident("x"),
		Iterable: ident("aiterable"),
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "print", Expression: ident("x")}},
		},
	}

	result := desugarStatement(stmt)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}

	// First: _iter = await iterable.__aiter__()
	iterAssign, ok := block.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", block.Statements[0])
	}
	if len(iterAssign.Names) != 1 {
		t.Fatalf("expected 1 name, got %d", len(iterAssign.Names))
	}

	// Second: while True: ...
	whileStmt, ok := block.Statements[1].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", block.Statements[1])
	}
	// Condition should be True
	boolCond, ok := whileStmt.Condition.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True condition, got %T", whileStmt.Condition)
	}
}

// ===== 3. Async with desugaring =====

func TestDesugarAsyncWithStatement_Single(t *testing.T) {
	stmt := &ast.AsyncWithStatement{
		Token: "async with",
		Items: []*ast.ContextManagerItem{
			{Expr: ident("cm"), Name: ident("var")},
		},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "print", Expression: ident("var")}},
		},
	}

	result := desugarStatement(stmt)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}

	// First: cm var assignment
	_, ok = block.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", block.Statements[0])
	}

	// Second: try/finally
	tryStmt, ok := block.Statements[1].(*ast.TryStatement)
	if !ok {
		t.Fatalf("expected TryStatement, got %T", block.Statements[1])
	}
	if tryStmt.Finally == nil {
		t.Fatal("expected Finally block")
	}
}

func TestDesugarAsyncWithStatement_Multiple(t *testing.T) {
	stmt := &ast.AsyncWithStatement{
		Token: "async with",
		Items: []*ast.ContextManagerItem{
			{Expr: ident("cm1"), Name: ident("v1")},
			{Expr: ident("cm2"), Name: ident("v2")},
		},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "pass", Expression: ident("v1")}},
		},
	}

	result := desugarStatement(stmt)
	// Multiple async with items should be nested
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// Should have 2 statements: cm1 assign + try/finally for cm1
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}
}

func TestDesugarAsyncWithStatement_NoAs(t *testing.T) {
	stmt := &ast.AsyncWithStatement{
		Token: "async with",
		Items: []*ast.ContextManagerItem{
			{Expr: ident("cm"), Name: nil},
		},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{},
		},
	}

	result := desugarStatement(stmt)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// The enterStmt should be an ExpressionStatement (no as clause)
	tryStmt, ok := block.Statements[1].(*ast.TryStatement)
	if !ok {
		t.Fatalf("expected TryStatement, got %T", block.Statements[1])
	}
	// Body should contain ExpressionStatement for __aenter__
	if len(tryStmt.Body.Statements) < 1 {
		t.Fatal("try body should have at least 1 statement")
	}
	_, ok = tryStmt.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement for __aenter__, got %T", tryStmt.Body.Statements[0])
	}
}

// ===== 4. Decorator desugaring =====

func TestDesugarDecorator_Single(t *testing.T) {
	decorator := ident("my_decorator")
	fnLit := &ast.FunctionLiteral{
		Token:       "def",
		Name:        "foo",
		Parameters:  []*ast.Identifier{ident("x")},
		Body:        &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Decorators:  []ast.Expression{decorator},
	}

	stmt := &ast.ExpressionStatement{
		Token:      "def",
		Expression: fnLit,
	}

	result := desugarStatement(stmt)
	// Single decorator produces a BlockStatement with a LetStatement
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	if len(block.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(block.Statements))
	}
	letStmt, ok := block.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", block.Statements[0])
	}
	if letStmt.Names[0].Value != "foo" {
		t.Fatalf("expected name 'foo', got '%s'", letStmt.Names[0].Value)
	}

	// Value should be a CallExpression: my_decorator(temp_fn)
	callExpr, ok := letStmt.Value.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", letStmt.Value)
	}
	funcIdent, ok := callExpr.Function.(*ast.Identifier)
	if !ok || funcIdent.Value != "my_decorator" {
		t.Fatalf("expected 'my_decorator', got %v", callExpr.Function)
	}
}

func TestDesugarDecorator_Multiple(t *testing.T) {
	dec1 := ident("dec1")
	dec2 := ident("dec2")
	fnLit := &ast.FunctionLiteral{
		Token:       "def",
		Name:        "foo",
		Parameters:  []*ast.Identifier{},
		Body:        &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Decorators:  []ast.Expression{dec1, dec2},
	}

	stmt := &ast.ExpressionStatement{
		Token:      "def",
		Expression: fnLit,
	}

	result := desugarStatement(stmt)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// Multiple decorators: temp_fn, _temp_foo = dec2(temp_fn), foo = dec1(_temp_foo)
	if len(block.Statements) < 3 {
		t.Fatalf("expected at least 3 statements, got %d", len(block.Statements))
	}
}

// ===== 5. Match/case desugaring =====

func TestDesugarMatchStatement_IntegerPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: intLit(1),
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: strLit("one")}},
				},
			},
			{
				Token:   "case",
				Pattern: intLit(2),
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{&ast.ExpressionStatement{Token: "2", Expression: strLit("two")}},
				},
			},
		},
	}

	result := desugarStatement(ms)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// Should have: let _match_val = x; if expression
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}

	// First: let _match_val = x
	letStmt, ok := block.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", block.Statements[0])
	}
	if letStmt.Names[0].Value != "_match_val" {
		t.Fatalf("expected '_match_val', got '%s'", letStmt.Names[0].Value)
	}

	// Second: if expression chain
	exprStmt, ok := block.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", block.Statements[1])
	}
	ifExpr, ok := exprStmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", exprStmt.Expression)
	}
	// Condition should be _match_val == 1
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "==" {
		t.Fatalf("expected == comparison, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_WithGuard(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: ident("n"),
				Guard:   &ast.InfixExpression{Token: ">", Left: ident("n"), Operator: ">", Right: intLit(0)},
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: strLit("positive")}},
				},
			},
		},
	}

	result := desugarStatement(ms)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	exprStmt, ok := block.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", block.Statements[1])
	}
	ifExpr, ok := exprStmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", exprStmt.Expression)
	}
	// Guard should produce an 'and' condition
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "and" {
		t.Fatalf("expected 'and' operator for guard, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_WildcardPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: ident("_"),
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: strLit("default")}},
				},
			},
		},
	}

	result := desugarStatement(ms)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	// Wildcard should produce True condition
	boolCond, ok := ifExpr.Condition.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True condition for wildcard, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_NonePattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: ident("None"),
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{},
				},
			},
		},
	}

	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "==" {
		t.Fatalf("expected == for None pattern, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_TrueFalsePattern(t *testing.T) {
	for _, pattern := range []struct {
		name    string
		pattern *ast.Identifier
	}{
		{"True", ident("True")},
		{"False", ident("False")},
	} {
		t.Run(pattern.name, func(t *testing.T) {
			ms := &ast.MatchStatement{
				Token:   "match",
				Subject: ident("x"),
				Cases: []*ast.CaseClause{
					{
						Token:   "case",
						Pattern: pattern.pattern,
						Body:    &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
					},
				},
			}
			result := desugarStatement(ms)
			block := result.(*ast.BlockStatement)
			exprStmt := block.Statements[1].(*ast.ExpressionStatement)
			ifExpr := exprStmt.Expression.(*ast.IfExpression)
			infix, ok := ifExpr.Condition.(*ast.InfixExpression)
			if !ok || infix.Operator != "==" {
				t.Fatalf("expected == for %s pattern, got %T", pattern.name, ifExpr.Condition)
			}
		})
	}
}

func TestDesugarMatchStatement_OrPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token: "case",
				Pattern: &ast.InfixExpression{
					Token:    "|",
					Left:     intLit(1),
					Operator: "|",
					Right:    intLit(2),
				},
				Body: &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}

	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "or" {
		t.Fatalf("expected 'or' for | pattern, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_ListPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token: "case",
				Pattern: &ast.ListLiteral{
					Token:    "[",
					Elements: []ast.Expression{intLit(1), intLit(2)},
				},
				Body: &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}

	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	// Should have len check and element checks combined with 'and'
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "and" {
		t.Fatalf("expected 'and' for list pattern, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_CallPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token: "case",
				Pattern: &ast.CallExpression{
					Token:    "Point",
					Function: ident("Point"),
					Arguments: []ast.Expression{ident("x_val"), ident("y_val")},
				},
				Body: &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}

	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	// Should have isinstance check
	_, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_NoCases(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases:   []*ast.CaseClause{},
	}

	result := desugarStatement(ms)
	// No cases: should return ExpressionStatement with True
	exprStmt, ok := result.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", result)
	}
	boolExpr, ok := exprStmt.Expression.(*ast.Boolean)
	if !ok || !boolExpr.Value {
		t.Fatalf("expected True, got %T", exprStmt.Expression)
	}
}

func TestDesugarMatchStatement_FloatPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: &ast.FloatLiteral{Token: "3.14", Value: 3.14},
				Body:    &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}
	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "==" {
		t.Fatalf("expected == for float pattern, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_StringPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: strLit("hello"),
				Body:    &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}
	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "==" {
		t.Fatalf("expected == for string pattern, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_ComplexPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: &ast.ComplexLiteral{Token: "1j", Value: "1j"},
				Body:    &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}
	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "==" {
		t.Fatalf("expected == for complex pattern, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_BooleanPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: boolLit(true),
				Body:    &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}
	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	infix, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || infix.Operator != "==" {
		t.Fatalf("expected == for boolean pattern, got %T", ifExpr.Condition)
	}
}

func TestDesugarMatchStatement_NilPattern(t *testing.T) {
	ms := &ast.MatchStatement{
		Token:   "match",
		Subject: ident("x"),
		Cases: []*ast.CaseClause{
			{
				Token:   "case",
				Pattern: nil,
				Body:    &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}
	result := desugarStatement(ms)
	block := result.(*ast.BlockStatement)
	exprStmt := block.Statements[1].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	boolCond, ok := ifExpr.Condition.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True for nil pattern, got %T", ifExpr.Condition)
	}
}

// ===== 6. Comprehension desugaring =====

func TestDesugarListComprehension(t *testing.T) {
	lc := &ast.ListComprehension{
		Token:    "[",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("items"),
		Filter:   nil,
	}

	result := desugarExpression(lc)
	resultLC, ok := result.(*ast.ListComprehension)
	if !ok {
		t.Fatalf("expected ListComprehension, got %T", result)
	}
	if resultLC.Element == nil || resultLC.Iterable == nil {
		t.Fatal("element and iterable should be desugared")
	}
}

func TestDesugarListComprehension_WithFilter(t *testing.T) {
	lc := &ast.ListComprehension{
		Token:    "[",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("items"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("x"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(lc)
	resultLC, ok := result.(*ast.ListComprehension)
	if !ok {
		t.Fatalf("expected ListComprehension, got %T", result)
	}
	if resultLC.Filter == nil {
		t.Fatal("filter should be desugared, not nil")
	}
}

func TestDesugarDictComprehension(t *testing.T) {
	dc := &ast.DictComprehension{
		Token:    "{",
		Key:      ident("k"),
		Value:    ident("v"),
		Variable: ident("item"),
		Iterable: ident("items"),
		Filter:   nil,
	}

	result := desugarExpression(dc)
	resultDC, ok := result.(*ast.DictComprehension)
	if !ok {
		t.Fatalf("expected DictComprehension, got %T", result)
	}
	if resultDC.Key == nil || resultDC.Value == nil || resultDC.Iterable == nil {
		t.Fatal("key, value, and iterable should be desugared")
	}
}

func TestDesugarDictComprehension_WithFilter(t *testing.T) {
	dc := &ast.DictComprehension{
		Token:    "{",
		Key:      ident("k"),
		Value:    ident("v"),
		Variable: ident("item"),
		Iterable: ident("items"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("k"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(dc)
	resultDC := result.(*ast.DictComprehension)
	if resultDC.Filter == nil {
		t.Fatal("filter should be desugared")
	}
}

func TestDesugarSetComprehension(t *testing.T) {
	sc := &ast.SetComprehension{
		Token:    "{",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("items"),
		Filter:   nil,
	}

	result := desugarExpression(sc)
	resultSC, ok := result.(*ast.SetComprehension)
	if !ok {
		t.Fatalf("expected SetComprehension, got %T", result)
	}
	if resultSC.Element == nil || resultSC.Iterable == nil {
		t.Fatal("element and iterable should be desugared")
	}
}

func TestDesugarSetComprehension_WithFilter(t *testing.T) {
	sc := &ast.SetComprehension{
		Token:    "{",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("items"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("x"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(sc)
	resultSC := result.(*ast.SetComprehension)
	if resultSC.Filter == nil {
		t.Fatal("filter should be desugared")
	}
}

func TestDesugarGeneratorExpression(t *testing.T) {
	ge := &ast.GeneratorExpression{
		Token:    "(",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("items"),
		Filter:   nil,
	}

	result := desugarExpression(ge)
	resultGE, ok := result.(*ast.GeneratorExpression)
	if !ok {
		t.Fatalf("expected GeneratorExpression, got %T", result)
	}
	if resultGE.Element == nil || resultGE.Iterable == nil {
		t.Fatal("element and iterable should be desugared")
	}
}

func TestDesugarGeneratorExpression_WithFilter(t *testing.T) {
	ge := &ast.GeneratorExpression{
		Token:    "(",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("items"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("x"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(ge)
	resultGE := result.(*ast.GeneratorExpression)
	if resultGE.Filter == nil {
		t.Fatal("filter should be desugared")
	}
}

// ===== 7. Yield from desugaring =====

func TestDesugarYieldFromStatement(t *testing.T) {
	stmt := &ast.YieldFromStatement{
		Token:      "yield from",
		Expression: ident("generator"),
	}

	result := desugarStatement(stmt)
	// yield from should be desugared to for→while
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// Should have: let _i_N = 0; while _i_N < len(generator): ...
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}
	letStmt, ok := block.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", block.Statements[0])
	}
	// Index variable should start with _i_
	if letStmt.Names[0].Value[:3] != "_i_" {
		t.Fatalf("expected index var starting with _i_, got '%s'", letStmt.Names[0].Value)
	}
}

// ===== 8. Named expression (walrus :=) desugaring =====

func TestDesugarNamedExpression(t *testing.T) {
	ne := &ast.NamedExpression{
		Token: ":=",
		Name:  ident("x"),
		Value: intLit(5),
	}

	result := desugarExpression(ne)
	resultNE, ok := result.(*ast.NamedExpression)
	if !ok {
		t.Fatalf("expected NamedExpression, got %T", result)
	}
	if resultNE.Name.Value != "x" {
		t.Fatalf("expected name 'x', got '%s'", resultNE.Name.Value)
	}
}

// ===== 9. Ternary expression desugaring =====

func TestDesugarTernaryExpression(t *testing.T) {
	te := &ast.TernaryExpression{
		Token:       "if",
		Condition:   ident("flag"),
		Consequence: intLit(1),
		Alternative: intLit(2),
	}

	result := desugarExpression(te)
	ifExpr, ok := result.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", result)
	}
	// Condition should be 'flag'
	if _, ok := ifExpr.Condition.(*ast.Identifier); !ok {
		t.Fatalf("expected Identifier condition, got %T", ifExpr.Condition)
	}
	// Consequence should contain 1
	if len(ifExpr.Consequence.Statements) != 1 {
		t.Fatalf("expected 1 consequence statement, got %d", len(ifExpr.Consequence.Statements))
	}
	// Alternative should contain 2
	if ifExpr.Alternative == nil {
		t.Fatal("alternative should not be nil")
	}
	if len(ifExpr.Alternative.Statements) != 1 {
		t.Fatalf("expected 1 alternative statement, got %d", len(ifExpr.Alternative.Statements))
	}
}

func TestDesugarTernaryExpression_NoAlternative(t *testing.T) {
	te := &ast.TernaryExpression{
		Token:       "if",
		Condition:   ident("flag"),
		Consequence: intLit(1),
		Alternative: nil,
	}

	result := desugarExpression(te)
	ifExpr := result.(*ast.IfExpression)
	if ifExpr.Alternative != nil {
		t.Fatal("alternative should be nil when not provided")
	}
}

// ===== 10. Try/except/finally desugaring =====

func TestDesugarTryStatement(t *testing.T) {
	ts := &ast.TryStatement{
		Token: "try",
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("x")}},
		},
		Excepts: []*ast.ExceptClause{
			{
				Token: "except",
				Type:  ident("ValueError"),
				Name:  ident("e"),
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("e")}},
				},
			},
		},
		Finally: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("cleanup")}},
		},
	}

	result := desugarStatement(ts)
	tryResult, ok := result.(*ast.TryStatement)
	if !ok {
		t.Fatalf("expected TryStatement, got %T", result)
	}
	if tryResult.Body == nil {
		t.Fatal("body should not be nil")
	}
	if len(tryResult.Excepts) != 1 {
		t.Fatalf("expected 1 except clause, got %d", len(tryResult.Excepts))
	}
	if tryResult.Finally == nil {
		t.Fatal("finally should not be nil")
	}
}

func TestDesugarTryStatement_NoFinally(t *testing.T) {
	ts := &ast.TryStatement{
		Token: "try",
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{},
		},
		Excepts: []*ast.ExceptClause{
			{
				Token: "except",
				Type:  nil,
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{},
				},
			},
		},
		Finally: nil,
	}

	result := desugarStatement(ts)
	tryResult := result.(*ast.TryStatement)
	if tryResult.Finally != nil {
		t.Fatal("finally should be nil when not provided")
	}
}

func TestDesugarTryStatement_StarExcept(t *testing.T) {
	ts := &ast.TryStatement{
		Token: "try",
		Body:  &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Excepts: []*ast.ExceptClause{
			{
				Token:  "except*",
				Type:   ident("ExceptionGroup"),
				IsStar: true,
				Body:   &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
			},
		},
	}

	result := desugarStatement(ts)
	tryResult := result.(*ast.TryStatement)
	if !tryResult.Excepts[0].IsStar {
		t.Fatal("IsStar should be preserved")
	}
}

// ===== 11. With statement desugaring =====

func TestDesugarWithStatement_Single(t *testing.T) {
	ws := &ast.WithStatement{
		Token: "with",
		Items: []*ast.ContextManagerItem{
			{Expr: ident("cm"), Name: ident("var")},
		},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("var")}},
		},
	}

	result := desugarStatement(ws)
	withResult, ok := result.(*ast.WithStatement)
	if !ok {
		t.Fatalf("expected WithStatement, got %T", result)
	}
	if len(withResult.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(withResult.Items))
	}
}

func TestDesugarWithStatement_Multiple(t *testing.T) {
	ws := &ast.WithStatement{
		Token: "with",
		Items: []*ast.ContextManagerItem{
			{Expr: ident("cm1"), Name: ident("v1")},
			{Expr: ident("cm2"), Name: ident("v2")},
		},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("v1")}},
		},
	}

	result := desugarStatement(ws)
	// Multiple context managers should be nested
	outerWith, ok := result.(*ast.WithStatement)
	if !ok {
		t.Fatalf("expected WithStatement, got %T", result)
	}
	if len(outerWith.Items) != 1 {
		t.Fatalf("expected 1 item in outer with, got %d", len(outerWith.Items))
	}
	// Inner body should contain another WithStatement
	innerWith, ok := outerWith.Body.Statements[0].(*ast.WithStatement)
	if !ok {
		t.Fatalf("expected nested WithStatement, got %T", outerWith.Body.Statements[0])
	}
	if len(innerWith.Items) != 1 {
		t.Fatalf("expected 1 item in inner with, got %d", len(innerWith.Items))
	}
}

func TestDesugarWithStatement_ThreeItems(t *testing.T) {
	ws := &ast.WithStatement{
		Token: "with",
		Items: []*ast.ContextManagerItem{
			{Expr: ident("cm1"), Name: ident("v1")},
			{Expr: ident("cm2"), Name: ident("v2")},
			{Expr: ident("cm3"), Name: ident("v3")},
		},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{},
		},
	}

	result := desugarStatement(ws)
	outerWith := result.(*ast.WithStatement)
	// cm1 is outermost, cm3 is innermost
	if outerWith.Items[0].Expr.(*ast.Identifier).Value != "cm1" {
		t.Fatal("outermost with should be cm1")
	}
	innerWith := outerWith.Body.Statements[0].(*ast.WithStatement)
	if innerWith.Items[0].Expr.(*ast.Identifier).Value != "cm2" {
		t.Fatal("middle with should be cm2")
	}
	innermostWith := innerWith.Body.Statements[0].(*ast.WithStatement)
	if innermostWith.Items[0].Expr.(*ast.Identifier).Value != "cm3" {
		t.Fatal("innermost with should be cm3")
	}
}

// ===== 12. Enum desugaring =====

func TestDesugarEnum(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:        "class",
		Name:         ident("Color"),
		SuperClass:   ident("Enum"),
		SuperClasses: nil,
		Body: &ast.BlockStatement{
			Token: ":",
			Statements: []ast.Statement{
				&ast.AssignStatement{
					Token: "=",
					Names: []*ast.Identifier{ident("RED")},
					Value: intLit(1),
				},
				&ast.AssignStatement{
					Token: "=",
					Names: []*ast.Identifier{ident("GREEN")},
					Value: intLit(2),
				},
			},
		},
	}

	result := desugarStatement(cs)
	assignResult, ok := result.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", result)
	}
	if assignResult.Names[0].Value != "Color" {
		t.Fatalf("expected name 'Color', got '%s'", assignResult.Names[0].Value)
	}
	// Value should be a call to __enum__
	callExpr, ok := assignResult.Value.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", assignResult.Value)
	}
	funcIdent, ok := callExpr.Function.(*ast.Identifier)
	if !ok || funcIdent.Value != "__enum__" {
		t.Fatalf("expected __enum__ call, got %v", callExpr.Function)
	}
}

func TestDesugarEnum_SuperClasses(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:        "class",
		Name:         ident("MyEnum"),
		SuperClass:   nil,
		SuperClasses: []*ast.Identifier{ident("Enum")},
		Body: &ast.BlockStatement{
			Token: ":",
			Statements: []ast.Statement{
				&ast.AssignStatement{
					Token: "=",
					Names: []*ast.Identifier{ident("A")},
					Value: intLit(1),
				},
			},
		},
	}

	result := desugarStatement(cs)
	assignResult, ok := result.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement for enum, got %T", result)
	}
	callExpr := assignResult.Value.(*ast.CallExpression)
	funcIdent := callExpr.Function.(*ast.Identifier)
	if funcIdent.Value != "__enum__" {
		t.Fatalf("expected __enum__ call, got '%s'", funcIdent.Value)
	}
}

func TestDesugarEnum_AutoValue(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:        "class",
		Name:         ident("AutoEnum"),
		SuperClass:   ident("Enum"),
		SuperClasses: nil,
		Body: &ast.BlockStatement{
			Token: ":",
			Statements: []ast.Statement{
				&ast.AssignStatement{
					Token: "=",
					Names: []*ast.Identifier{ident("A")},
					Value: nil, // auto value
				},
				&ast.AssignStatement{
					Token: "=",
					Names: []*ast.Identifier{ident("B")},
					Value: nil, // auto value
				},
			},
		},
	}

	result := desugarStatement(cs)
	assignResult := result.(*ast.AssignStatement)
	callExpr := assignResult.Value.(*ast.CallExpression)
	// Second argument should be a HashLiteral with auto values
	hashLit, ok := callExpr.Arguments[1].(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expected HashLiteral, got %T", callExpr.Arguments[1])
	}
	if len(hashLit.Pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(hashLit.Pairs))
	}
}

func TestDesugarEnum_LetStatementMembers(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:        "class",
		Name:         ident("LetEnum"),
		SuperClass:   ident("Enum"),
		SuperClasses: nil,
		Body: &ast.BlockStatement{
			Token: ":",
			Statements: []ast.Statement{
				&ast.LetStatement{
					Token: "let",
					Names: []*ast.Identifier{ident("X")},
					Value: intLit(10),
				},
			},
		},
	}

	result := desugarStatement(cs)
	assignResult, ok := result.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", result)
	}
	callExpr := assignResult.Value.(*ast.CallExpression)
	hashLit := callExpr.Arguments[1].(*ast.HashLiteral)
	if len(hashLit.Pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(hashLit.Pairs))
	}
}

func TestDesugarEnum_LetStatementAutoValue(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:        "class",
		Name:         ident("LetAutoEnum"),
		SuperClass:   ident("Enum"),
		SuperClasses: nil,
		Body: &ast.BlockStatement{
			Token: ":",
			Statements: []ast.Statement{
				&ast.LetStatement{
					Token: "let",
					Names: []*ast.Identifier{ident("Y")},
					Value: nil, // auto value
				},
			},
		},
	}

	result := desugarStatement(cs)
	assignResult := result.(*ast.AssignStatement)
	callExpr := assignResult.Value.(*ast.CallExpression)
	hashLit := callExpr.Arguments[1].(*ast.HashLiteral)
	if len(hashLit.Pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(hashLit.Pairs))
	}
}

// ===== 13. Slice assignment desugaring =====

func TestDesugarSliceAssignStatement(t *testing.T) {
	sas := &ast.SliceAssignStatement{
		Token: "=",
		Left:  ident("lst"),
		Lower: intLit(1),
		Upper: intLit(3),
		Step:  nil,
		Value: &ast.ListLiteral{Token: "[", Elements: []ast.Expression{intLit(4), intLit(5)}},
	}

	result := desugarStatement(sas)
	sliceResult, ok := result.(*ast.SliceAssignStatement)
	if !ok {
		t.Fatalf("expected SliceAssignStatement, got %T", result)
	}
	if sliceResult.Left == nil || sliceResult.Lower == nil || sliceResult.Upper == nil {
		t.Fatal("left, lower, upper should be desugared")
	}
}

// ===== 14. LRU cache desugaring =====

func TestIsLruCacheDecorator_Identifier(t *testing.T) {
	dec := ident("lru_cache")
	if !isLruCacheDecorator(dec) {
		t.Fatal("expected lru_cache identifier to be recognized")
	}
}

func TestIsLruCacheDecorator_CallExpression(t *testing.T) {
	dec := &ast.CallExpression{
		Token:    "lru_cache",
		Function: ident("lru_cache"),
	}
	if !isLruCacheDecorator(dec) {
		t.Fatal("expected lru_cache() call to be recognized")
	}
}

func TestIsLruCacheDecorator_MemberAccess(t *testing.T) {
	dec := &ast.MemberAccess{
		Token:  ".",
		Object: ident("functools"),
		Member: ident("lru_cache"),
	}
	if !isLruCacheDecorator(dec) {
		t.Fatal("expected functools.lru_cache to be recognized")
	}
}

func TestIsLruCacheDecorator_CallMemberAccess(t *testing.T) {
	dec := &ast.CallExpression{
		Token: "lru_cache",
		Function: &ast.MemberAccess{
			Token:  ".",
			Object: ident("functools"),
			Member: ident("lru_cache"),
		},
	}
	if !isLruCacheDecorator(dec) {
		t.Fatal("expected functools.lru_cache() to be recognized")
	}
}

func TestIsLruCacheDecorator_NotLruCache(t *testing.T) {
	dec := ident("other_decorator")
	if isLruCacheDecorator(dec) {
		t.Fatal("expected other_decorator to NOT be recognized as lru_cache")
	}
}

func TestDesugarLruCache(t *testing.T) {
	fnLit := &ast.FunctionLiteral{
		Token:       "def",
		Name:        "fib",
		Parameters:  []*ast.Identifier{ident("n")},
		Body:        &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Decorators:  []ast.Expression{ident("lru_cache")},
	}

	stmt := &ast.ExpressionStatement{
		Token:      "def",
		Expression: fnLit,
	}

	result := desugarStatement(stmt)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// Should have: let _lru_orig_fib = ..., let fib = wrapper, fib._cache = {}
	if len(block.Statements) != 3 {
		t.Fatalf("expected 3 statements for lru_cache, got %d", len(block.Statements))
	}

	// First: let _lru_orig_fib = original function
	let1, ok := block.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", block.Statements[0])
	}
	if let1.Names[0].Value != "_lru_orig_fib" {
		t.Fatalf("expected '_lru_orig_fib', got '%s'", let1.Names[0].Value)
	}

	// Second: let fib = wrapper function
	let2, ok := block.Statements[1].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", block.Statements[1])
	}
	if let2.Names[0].Value != "fib" {
		t.Fatalf("expected 'fib', got '%s'", let2.Names[0].Value)
	}

	// Third: fib._cache = {}
	_, ok = block.Statements[2].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", block.Statements[2])
	}
}

// ===== 15. Async comprehension desugaring =====

func TestDesugarAsyncListComprehension(t *testing.T) {
	alc := &ast.AsyncListComprehension{
		Token:    "[",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("aiterable"),
		Filter:   nil,
	}

	result := desugarExpression(alc)
	resultALC, ok := result.(*ast.AsyncListComprehension)
	if !ok {
		t.Fatalf("expected AsyncListComprehension, got %T", result)
	}
	if resultALC.Element == nil || resultALC.Iterable == nil {
		t.Fatal("element and iterable should be desugared")
	}
}

func TestDesugarAsyncListComprehension_WithFilter(t *testing.T) {
	alc := &ast.AsyncListComprehension{
		Token:    "[",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("aiterable"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("x"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(alc)
	resultALC := result.(*ast.AsyncListComprehension)
	if resultALC.Filter == nil {
		t.Fatal("filter should be desugared")
	}
}

func TestDesugarAsyncSetComprehension(t *testing.T) {
	asc := &ast.AsyncSetComprehension{
		Token:    "{",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("aiterable"),
		Filter:   nil,
	}

	result := desugarExpression(asc)
	resultASC, ok := result.(*ast.AsyncSetComprehension)
	if !ok {
		t.Fatalf("expected AsyncSetComprehension, got %T", result)
	}
	if resultASC.Element == nil || resultASC.Iterable == nil {
		t.Fatal("element and iterable should be desugared")
	}
}

func TestDesugarAsyncSetComprehension_WithFilter(t *testing.T) {
	asc := &ast.AsyncSetComprehension{
		Token:    "{",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("aiterable"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("x"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(asc)
	resultASC := result.(*ast.AsyncSetComprehension)
	if resultASC.Filter == nil {
		t.Fatal("filter should be desugared")
	}
}

func TestDesugarAsyncDictComprehension(t *testing.T) {
	adc := &ast.AsyncDictComprehension{
		Token:    "{",
		Key:      ident("k"),
		Value:    ident("v"),
		Variable: ident("item"),
		Iterable: ident("aiterable"),
		Filter:   nil,
	}

	result := desugarExpression(adc)
	resultADC, ok := result.(*ast.AsyncDictComprehension)
	if !ok {
		t.Fatalf("expected AsyncDictComprehension, got %T", result)
	}
	if resultADC.Key == nil || resultADC.Value == nil || resultADC.Iterable == nil {
		t.Fatal("key, value, and iterable should be desugared")
	}
}

func TestDesugarAsyncDictComprehension_WithFilter(t *testing.T) {
	adc := &ast.AsyncDictComprehension{
		Token:    "{",
		Key:      ident("k"),
		Value:    ident("v"),
		Variable: ident("item"),
		Iterable: ident("aiterable"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("k"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(adc)
	resultADC := result.(*ast.AsyncDictComprehension)
	if resultADC.Filter == nil {
		t.Fatal("filter should be desugared")
	}
}

func TestDesugarAsyncGeneratorExpression(t *testing.T) {
	age := &ast.AsyncGeneratorExpression{
		Token:    "(",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("aiterable"),
		Filter:   nil,
	}

	result := desugarExpression(age)
	resultAGE, ok := result.(*ast.AsyncGeneratorExpression)
	if !ok {
		t.Fatalf("expected AsyncGeneratorExpression, got %T", result)
	}
	if resultAGE.Element == nil || resultAGE.Iterable == nil {
		t.Fatal("element and iterable should be desugared")
	}
}

func TestDesugarAsyncGeneratorExpression_WithFilter(t *testing.T) {
	age := &ast.AsyncGeneratorExpression{
		Token:    "(",
		Element:  ident("x"),
		Variable: ident("x"),
		Iterable: ident("aiterable"),
		Filter:   &ast.InfixExpression{Token: ">", Left: ident("x"), Operator: ">", Right: intLit(0)},
	}

	result := desugarExpression(age)
	resultAGE := result.(*ast.AsyncGeneratorExpression)
	if resultAGE.Filter == nil {
		t.Fatal("filter should be desugared")
	}
}

// ===== Additional: Multi-variable assignment =====

func TestDesugarLetStatement_MultiAssign(t *testing.T) {
	ls := &ast.LetStatement{
		Token: "let",
		Names: []*ast.Identifier{ident("a"), ident("b")},
		Value: &ast.ListLiteral{Token: "[", Elements: []ast.Expression{intLit(1), intLit(2)}},
	}

	result := desugarStatement(ls)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// Should have: let _temp = [1, 2]; let a = _temp[0]; let b = _temp[1]
	if len(block.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(block.Statements))
	}
	// First: let _temp = ...
	letStmt := block.Statements[0].(*ast.LetStatement)
	if letStmt.Names[0].Value != "_temp" {
		t.Fatalf("expected '_temp', got '%s'", letStmt.Names[0].Value)
	}
	// Second: let a = _temp[0]
	letA := block.Statements[1].(*ast.LetStatement)
	if letA.Names[0].Value != "a" {
		t.Fatalf("expected 'a', got '%s'", letA.Names[0].Value)
	}
	// Third: let b = _temp[1]
	letB := block.Statements[2].(*ast.LetStatement)
	if letB.Names[0].Value != "b" {
		t.Fatalf("expected 'b', got '%s'", letB.Names[0].Value)
	}
}

func TestDesugarLetStatement_SingleAssign(t *testing.T) {
	ls := &ast.LetStatement{
		Token: "let",
		Names: []*ast.Identifier{ident("x")},
		Value: intLit(5),
	}

	result := desugarStatement(ls)
	letResult, ok := result.(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", result)
	}
	if letResult.Names[0].Value != "x" {
		t.Fatalf("expected 'x', got '%s'", letResult.Names[0].Value)
	}
}

func TestDesugarAssignStatement_MultiAssign(t *testing.T) {
	as := &ast.AssignStatement{
		Token: "=",
		Names: []*ast.Identifier{ident("a"), ident("b")},
		Value: &ast.ListLiteral{Token: "[", Elements: []ast.Expression{intLit(1), intLit(2)}},
	}

	result := desugarStatement(as)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	// Should have: let _temp = [1, 2]; a = _temp[0]; b = _temp[1]
	if len(block.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(block.Statements))
	}
	// Second: a = _temp[0] (AssignStatement, not LetStatement)
	assignA := block.Statements[1].(*ast.AssignStatement)
	if assignA.Names[0].Value != "a" {
		t.Fatalf("expected 'a', got '%s'", assignA.Names[0].Value)
	}
}

func TestDesugarAssignStatement_SingleAssign(t *testing.T) {
	as := &ast.AssignStatement{
		Token: "=",
		Names: []*ast.Identifier{ident("x")},
		Value: intLit(5),
	}

	result := desugarStatement(as)
	assignResult, ok := result.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", result)
	}
	if assignResult.Names[0].Value != "x" {
		t.Fatalf("expected 'x', got '%s'", assignResult.Names[0].Value)
	}
}

// ===== Additional: Chain comparison =====

func TestDesugarChainComparison(t *testing.T) {
	// a < b < c -> (a < b) AND (b < c)
	expr := &ast.InfixExpression{
		Token:    "<",
		Left:     &ast.InfixExpression{Token: "<", Left: ident("a"), Operator: "<", Right: ident("b")},
		Operator: "<",
		Right:    ident("c"),
	}

	result := desugarExpression(expr)
	// Should be desugared to IfExpression (AND desugaring)
	_, ok := result.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression (from AND desugaring), got %T", result)
	}
}

func TestDesugarChainComparison_RightAssociative(t *testing.T) {
	// a < (b < c) - right-associative chain
	expr := &ast.InfixExpression{
		Token:    "<",
		Left:     ident("a"),
		Operator: "<",
		Right:    &ast.InfixExpression{Token: "<", Left: ident("b"), Operator: "<", Right: ident("c")},
	}

	result := desugarExpression(expr)
	// Should be desugared
	_, ok := result.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", result)
	}
}

// ===== Additional: AND/OR desugaring =====

func TestDesugarAndExpression(t *testing.T) {
	expr := &ast.InfixExpression{
		Token:    "and",
		Left:     ident("a"),
		Operator: "and",
		Right:    ident("b"),
	}

	result := desugarExpression(expr)
	ifExpr, ok := result.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", result)
	}
	// if a then b else a
	if ifExpr.Condition.(*ast.Identifier).Value != "a" {
		t.Fatal("condition should be 'a'")
	}
}

func TestDesugarOrExpression(t *testing.T) {
	expr := &ast.InfixExpression{
		Token:    "or",
		Left:     ident("a"),
		Operator: "or",
		Right:    ident("b"),
	}

	result := desugarExpression(expr)
	ifExpr, ok := result.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", result)
	}
	// if a then a else b
	if ifExpr.Condition.(*ast.Identifier).Value != "a" {
		t.Fatal("condition should be 'a'")
	}
}

func TestDesugarInfixExpression_Normal(t *testing.T) {
	expr := &ast.InfixExpression{
		Token:    "+",
		Left:     ident("a"),
		Operator: "+",
		Right:    ident("b"),
	}

	result := desugarExpression(expr)
	infixResult, ok := result.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression, got %T", result)
	}
	if infixResult.Operator != "+" {
		t.Fatalf("expected '+', got '%s'", infixResult.Operator)
	}
}

// ===== Additional: NamedTuple desugaring =====

func TestDesugarNamedTuple(t *testing.T) {
	as := &ast.AssignStatement{
		Token: "=",
		Names: []*ast.Identifier{ident("Point")},
		Value: &ast.CallExpression{
			Token:    "NamedTuple",
			Function: ident("NamedTuple"),
			Arguments: []ast.Expression{
				strLit("Point"),
				&ast.ListLiteral{
					Token: "[",
					Elements: []ast.Expression{
						&ast.ListLiteral{
							Token:    "[",
							Elements: []ast.Expression{ident("x"), ident("int")},
						},
						&ast.ListLiteral{
							Token:    "[",
							Elements: []ast.Expression{ident("y"), ident("int")},
						},
					},
				},
			},
		},
	}

	result := desugarStatement(as)
	assignResult, ok := result.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", result)
	}
	classStmt, ok := assignResult.Value.(*ast.ClassStatement)
	if !ok {
		t.Fatalf("expected ClassStatement, got %T", assignResult.Value)
	}
	if classStmt.Name.Value != "Point" {
		t.Fatalf("expected 'Point', got '%s'", classStmt.Name.Value)
	}
	if len(classStmt.Methods) < 2 {
		t.Fatalf("expected at least 2 methods (__init__ and __repr__), got %d", len(classStmt.Methods))
	}
}

func TestDesugarNamedTuple_StringFields(t *testing.T) {
	as := &ast.AssignStatement{
		Token: "=",
		Names: []*ast.Identifier{ident("Simple")},
		Value: &ast.CallExpression{
			Token:    "NamedTuple",
			Function: ident("NamedTuple"),
			Arguments: []ast.Expression{
				strLit("Simple"),
				&ast.ListLiteral{
					Token:    "[",
					Elements: []ast.Expression{strLit("name"), strLit("value")},
				},
			},
		},
	}

	result := desugarStatement(as)
	assignResult := result.(*ast.AssignStatement)
	classStmt := assignResult.Value.(*ast.ClassStatement)
	if len(classStmt.Methods) < 2 {
		t.Fatalf("expected at least 2 methods, got %d", len(classStmt.Methods))
	}
}

// ===== Additional: Delete statement desugaring =====

func TestDesugarDeleteStatement_Identifier(t *testing.T) {
	ds := &ast.DeleteStatement{
		Token:   "del",
		Targets: []ast.Expression{ident("x")},
	}

	result := desugarStatement(ds)
	deleteResult, ok := result.(*ast.DeleteStatement)
	if !ok {
		t.Fatalf("expected DeleteStatement, got %T", result)
	}
	if len(deleteResult.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(deleteResult.Targets))
	}
}

func TestDesugarDeleteStatement_IndexExpression(t *testing.T) {
	ds := &ast.DeleteStatement{
		Token: "del",
		Targets: []ast.Expression{
			&ast.IndexExpression{
				Token: "[",
				Left:  ident("d"),
				Index: strLit("key"),
			},
		},
	}

	result := desugarStatement(ds)
	// del d["key"] should be desugared to d.__delitem__("key")
	exprStmt, ok := result.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", result)
	}
	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", exprStmt.Expression)
	}
	memberAccess, ok := callExpr.Function.(*ast.MemberAccess)
	if !ok || memberAccess.Member.Value != "__delitem__" {
		t.Fatalf("expected __delitem__ call, got %v", callExpr.Function)
	}
}

func TestDesugarDeleteStatement_MemberAccess(t *testing.T) {
	ds := &ast.DeleteStatement{
		Token: "del",
		Targets: []ast.Expression{
			&ast.MemberAccess{
				Token:  ".",
				Object: ident("obj"),
				Member: ident("attr"),
			},
		},
	}

	result := desugarStatement(ds)
	_, ok := result.(*ast.DeleteStatement)
	if !ok {
		t.Fatalf("expected DeleteStatement for member access, got %T", result)
	}
}

func TestDesugarDeleteStatement_MultipleTargets(t *testing.T) {
	ds := &ast.DeleteStatement{
		Token:   "del",
		Targets: []ast.Expression{ident("x"), ident("y")},
	}

	result := desugarStatement(ds)
	block, ok := result.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected BlockStatement, got %T", result)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}
}

// ===== Additional: Mixed list/dict literal with unpacking =====

func TestDesugarMixedListLiteral(t *testing.T) {
	ll := &ast.ListLiteral{
		Token: "[",
		Elements: []ast.Expression{
			intLit(1),
			&ast.ListUnpack{Token: "*", Value: ident("other")},
			intLit(3),
		},
	}

	result := desugarExpression(ll)
	// Should be desugared to IIFE
	callExpr, ok := result.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression (IIFE), got %T", result)
	}
	fnLit, ok := callExpr.Function.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected FunctionLiteral, got %T", callExpr.Function)
	}
	if fnLit.Body == nil {
		t.Fatal("function body should not be nil")
	}
}

func TestDesugarMixedDictLiteral(t *testing.T) {
	hl := &ast.HashLiteral{
		Token:    "{",
		Elements: []ast.Expression{
			&ast.KeyValuePair{Token: ":", Key: strLit("a"), Value: intLit(1)},
			&ast.DictionaryUnpack{Token: "**", Value: ident("other")},
		},
	}

	result := desugarExpression(hl)
	// Should be desugared to IIFE
	callExpr, ok := result.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression (IIFE), got %T", result)
	}
	_, ok = callExpr.Function.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected FunctionLiteral, got %T", callExpr.Function)
	}
}

func TestDesugarNormalListLiteral(t *testing.T) {
	ll := &ast.ListLiteral{
		Token:    "[",
		Elements: []ast.Expression{intLit(1), intLit(2), intLit(3)},
	}

	result := desugarExpression(ll)
	listResult, ok := result.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected ListLiteral, got %T", result)
	}
	if len(listResult.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(listResult.Elements))
	}
}

func TestDesugarHashLiteral_OldFormat(t *testing.T) {
	hl := &ast.HashLiteral{
		Token: "{",
		Pairs: map[ast.Expression]ast.Expression{
			strLit("a"): intLit(1),
			strLit("b"): intLit(2),
		},
	}

	result := desugarExpression(hl)
	hashResult, ok := result.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("expected HashLiteral, got %T", result)
	}
	if len(hashResult.Pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(hashResult.Pairs))
	}
}

// ===== Additional: Break/Continue statement =====

func TestDesugarBreakStatement(t *testing.T) {
	bs := &ast.BreakStatement{Token: "break"}
	result := desugarStatement(bs)
	if result != nil {
		t.Fatalf("expected nil for break statement, got %T", result)
	}
}

func TestDesugarContinueStatement(t *testing.T) {
	cs := &ast.ContinueStatement{Token: "continue"}
	result := desugarStatement(cs)
	if result != nil {
		t.Fatalf("expected nil for continue statement, got %T", result)
	}
}

// ===== Additional: Function with default parameters =====

func TestDesugarFunctionLiteral_WithDefaults(t *testing.T) {
	fnLit := &ast.FunctionLiteral{
		Token:      "def",
		Name:       "foo",
		Parameters: []*ast.Identifier{ident("x"), ident("y")},
		Defaults:   []ast.Expression{nil, intLit(10)},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ReturnStatement{Token: "return", ReturnValue: ident("x")}},
		},
	}

	result := desugarExpression(fnLit)
	fnResult, ok := result.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected FunctionLiteral, got %T", result)
	}
	// Body should have default parameter handling prepended
	if len(fnResult.Body.Statements) < 2 {
		t.Fatalf("expected at least 2 statements (default + original), got %d", len(fnResult.Body.Statements))
	}
}

func TestDesugarFunctionLiteral_NoDefaults(t *testing.T) {
	fnLit := &ast.FunctionLiteral{
		Token:      "def",
		Name:       "foo",
		Parameters: []*ast.Identifier{ident("x")},
		Defaults:   []ast.Expression{nil},
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ReturnStatement{Token: "return", ReturnValue: ident("x")}},
		},
	}

	result := desugarExpression(fnLit)
	fnResult := result.(*ast.FunctionLiteral)
	// No defaults should mean body is unchanged (just desugared)
	if len(fnResult.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(fnResult.Body.Statements))
	}
}

func TestDesugarFunctionLiteral_WithDecorators(t *testing.T) {
	fnLit := &ast.FunctionLiteral{
		Token:       "def",
		Name:        "foo",
		Parameters:  []*ast.Identifier{},
		Body:        &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Decorators:  []ast.Expression{ident("deco")},
	}

	result := desugarExpression(fnLit)
	fnResult := result.(*ast.FunctionLiteral)
	if len(fnResult.Decorators) != 1 {
		t.Fatalf("expected 1 decorator, got %d", len(fnResult.Decorators))
	}
}

func TestDesugarFunctionLiteral_WithReturnType(t *testing.T) {
	fnLit := &ast.FunctionLiteral{
		Token:       "def",
		Name:        "foo",
		Parameters:  []*ast.Identifier{ident("x")},
		Body:        &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		ReturnType:  ident("int"),
	}

	result := desugarExpression(fnLit)
	fnResult := result.(*ast.FunctionLiteral)
	if fnResult.ReturnType == nil {
		t.Fatal("return type should be preserved")
	}
}

// ===== Additional: Various expression types =====

func TestDesugarPrefixExpression(t *testing.T) {
	pe := &ast.PrefixExpression{
		Token:    "-",
		Operator: "-",
		Right:    ident("x"),
	}

	result := desugarExpression(pe)
	prefixResult, ok := result.(*ast.PrefixExpression)
	if !ok {
		t.Fatalf("expected PrefixExpression, got %T", result)
	}
	if prefixResult.Operator != "-" {
		t.Fatalf("expected '-', got '%s'", prefixResult.Operator)
	}
}

func TestDesugarAwaitExpression(t *testing.T) {
	ae := &ast.AwaitExpression{
		Token: "await",
		Value: ident("coroutine"),
	}

	result := desugarExpression(ae)
	awaitResult, ok := result.(*ast.AwaitExpression)
	if !ok {
		t.Fatalf("expected AwaitExpression, got %T", result)
	}
	if awaitResult.Value == nil {
		t.Fatal("value should be desugared")
	}
}

func TestDesugarCallExpression(t *testing.T) {
	ce := &ast.CallExpression{
		Token:    "foo",
		Function: ident("foo"),
		Arguments: []ast.Expression{intLit(1), intLit(2)},
	}

	result := desugarExpression(ce)
	callResult, ok := result.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", result)
	}
	if len(callResult.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(callResult.Arguments))
	}
}

func TestDesugarIndexExpression(t *testing.T) {
	ie := &ast.IndexExpression{
		Token: "[",
		Left:  ident("arr"),
		Index: intLit(0),
	}

	result := desugarExpression(ie)
	indexResult, ok := result.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected IndexExpression, got %T", result)
	}
	if indexResult.Left == nil || indexResult.Index == nil {
		t.Fatal("left and index should be desugared")
	}
}

func TestDesugarSliceExpression(t *testing.T) {
	se := &ast.SliceExpression{
		Token: ":",
		Lower: intLit(1),
		Upper: intLit(3),
		Step:  nil,
	}

	result := desugarExpression(se)
	sliceResult, ok := result.(*ast.SliceExpression)
	if !ok {
		t.Fatalf("expected SliceExpression, got %T", result)
	}
	if sliceResult.Lower == nil || sliceResult.Upper == nil {
		t.Fatal("lower and upper should be desugared")
	}
}

func TestDesugarLambdaExpression(t *testing.T) {
	le := &ast.LambdaExpression{
		Token:      "lambda",
		Parameters: []*ast.Identifier{ident("x")},
		Body:       ident("x"),
	}

	result := desugarExpression(le)
	lambdaResult, ok := result.(*ast.LambdaExpression)
	if !ok {
		t.Fatalf("expected LambdaExpression, got %T", result)
	}
	if lambdaResult.Body == nil {
		t.Fatal("body should be desugared")
	}
}

func TestDesugarIfExpression(t *testing.T) {
	ie := &ast.IfExpression{
		Token:       "if",
		Condition:   ident("flag"),
		Consequence: &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Alternative: &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
	}

	result := desugarExpression(ie)
	ifResult, ok := result.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", result)
	}
	if ifResult.Condition == nil {
		t.Fatal("condition should be desugared")
	}
}

func TestDesugarFStringLiteral(t *testing.T) {
	fsl := &ast.FStringLiteral{
		Token: "f\"",
		Parts: []ast.Expression{strLit("hello "), ident("name")},
	}

	result := desugarExpression(fsl)
	fslResult, ok := result.(*ast.FStringLiteral)
	if !ok {
		t.Fatalf("expected FStringLiteral, got %T", result)
	}
	if len(fslResult.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(fslResult.Parts))
	}
}

func TestDesugarKeyValuePair(t *testing.T) {
	kv := &ast.KeyValuePair{
		Token: ":",
		Key:   strLit("a"),
		Value: intLit(1),
	}

	result := desugarExpression(kv)
	kvResult, ok := result.(*ast.KeyValuePair)
	if !ok {
		t.Fatalf("expected KeyValuePair, got %T", result)
	}
	if kvResult.Key == nil || kvResult.Value == nil {
		t.Fatal("key and value should be desugared")
	}
}

func TestDesugarDictionaryUnpack(t *testing.T) {
	du := &ast.DictionaryUnpack{
		Token: "**",
		Value: ident("d"),
	}

	result := desugarExpression(du)
	duResult, ok := result.(*ast.DictionaryUnpack)
	if !ok {
		t.Fatalf("expected DictionaryUnpack, got %T", result)
	}
	if duResult.Value == nil {
		t.Fatal("value should be desugared")
	}
}

func TestDesugarListUnpack(t *testing.T) {
	lu := &ast.ListUnpack{
		Token: "*",
		Value: ident("lst"),
	}

	result := desugarExpression(lu)
	luResult, ok := result.(*ast.ListUnpack)
	if !ok {
		t.Fatalf("expected ListUnpack, got %T", result)
	}
	if luResult.Value == nil {
		t.Fatal("value should be desugared")
	}
}

// ===== Additional: Statement types =====

func TestDesugarAugAssignStatement(t *testing.T) {
	aas := &ast.AugAssignStatement{
		Token:    "+=",
		Name:     ident("x"),
		Operator: "+=",
		Value:    intLit(1),
	}

	result := desugarStatement(aas)
	augResult, ok := result.(*ast.AugAssignStatement)
	if !ok {
		t.Fatalf("expected AugAssignStatement, got %T", result)
	}
	if augResult.Operator != "+=" {
		t.Fatalf("expected '+=', got '%s'", augResult.Operator)
	}
}

func TestDesugarAugAssignStatement_WithIndex(t *testing.T) {
	aas := &ast.AugAssignStatement{
		Token:      "+=",
		Name:       ident("d"),
		Operator:   "+",
		Value:      intLit(1),
		IndexLeft:  ident("d"),
		IndexIndex: strLit("key"),
	}

	result := desugarStatement(aas)
	augResult := result.(*ast.AugAssignStatement)
	if augResult.IndexLeft == nil || augResult.IndexIndex == nil {
		t.Fatal("IndexLeft and IndexIndex should be desugared")
	}
}

func TestDesugarIndexAssignStatement(t *testing.T) {
	ias := &ast.IndexAssignStatement{
		Token: "=",
		Left:  ident("d"),
		Index: strLit("key"),
		Value: intLit(1),
	}

	result := desugarStatement(ias)
	indexResult, ok := result.(*ast.IndexAssignStatement)
	if !ok {
		t.Fatalf("expected IndexAssignStatement, got %T", result)
	}
	if indexResult.Left == nil || indexResult.Index == nil || indexResult.Value == nil {
		t.Fatal("left, index, and value should be desugared")
	}
}

func TestDesugarReturnStatement(t *testing.T) {
	rs := &ast.ReturnStatement{
		Token:       "return",
		ReturnValue: ident("x"),
	}

	result := desugarStatement(rs)
	returnResult, ok := result.(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", result)
	}
	if returnResult.ReturnValue == nil {
		t.Fatal("return value should be desugared")
	}
}

func TestDesugarWhileStatement(t *testing.T) {
	ws := &ast.WhileStatement{
		Token:     "while",
		Condition: ident("flag"),
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("x")}},
		},
	}

	result := desugarStatement(ws)
	whileResult, ok := result.(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", result)
	}
	if whileResult.Condition == nil {
		t.Fatal("condition should be desugared")
	}
}

func TestDesugarRaiseStatement(t *testing.T) {
	rs := &ast.RaiseStatement{
		Token:      "raise",
		Expression: ident("ValueError"),
	}

	result := desugarStatement(rs)
	raiseResult, ok := result.(*ast.RaiseStatement)
	if !ok {
		t.Fatalf("expected RaiseStatement, got %T", result)
	}
	if raiseResult.Expression == nil {
		t.Fatal("expression should be desugared")
	}
}

func TestDesugarYieldStatement(t *testing.T) {
	ys := &ast.YieldStatement{
		Token:      "yield",
		Expression: ident("x"),
	}

	result := desugarStatement(ys)
	yieldResult, ok := result.(*ast.YieldStatement)
	if !ok {
		t.Fatalf("expected YieldStatement, got %T", result)
	}
	if yieldResult.Expression == nil {
		t.Fatal("expression should be desugared")
	}
}

func TestDesugarGlobalStatement(t *testing.T) {
	gs := &ast.GlobalStatement{
		Token: "global",
		Names: []*ast.Identifier{ident("x"), ident("y")},
	}

	result := desugarStatement(gs)
	globalResult, ok := result.(*ast.GlobalStatement)
	if !ok {
		t.Fatalf("expected GlobalStatement, got %T", result)
	}
	if len(globalResult.Names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(globalResult.Names))
	}
}

func TestDesugarNonlocalStatement(t *testing.T) {
	ns := &ast.NonlocalStatement{
		Token: "nonlocal",
		Names: []*ast.Identifier{ident("x")},
	}

	result := desugarStatement(ns)
	nonlocalResult, ok := result.(*ast.NonlocalStatement)
	if !ok {
		t.Fatalf("expected NonlocalStatement, got %T", result)
	}
	if len(nonlocalResult.Names) != 1 {
		t.Fatalf("expected 1 name, got %d", len(nonlocalResult.Names))
	}
}

// ===== Additional: Class statement (non-enum) =====

func TestDesugarClassStatement_NonEnum(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:      "class",
		Name:       ident("Foo"),
		SuperClass: ident("Bar"),
		Body: &ast.BlockStatement{
			Token:      ":",
			Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("x")}},
		},
	}

	result := desugarStatement(cs)
	classResult, ok := result.(*ast.ClassStatement)
	if !ok {
		t.Fatalf("expected ClassStatement, got %T", result)
	}
	if classResult.Name.Value != "Foo" {
		t.Fatalf("expected 'Foo', got '%s'", classResult.Name.Value)
	}
}

// ===== Additional: Block statement =====

func TestDesugarBlockStatement(t *testing.T) {
	bs := &ast.BlockStatement{
		Token: ":",
		Statements: []ast.Statement{
			&ast.ExpressionStatement{Token: "1", Expression: ident("x")},
			&ast.BreakStatement{Token: "break"}, // should be filtered out
		},
	}

	result := desugarBlockStatement(bs)
	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement (break filtered), got %d", len(result.Statements))
	}
}

func TestDesugarBlockStatement_Nil(t *testing.T) {
	result := desugarBlockStatement(nil)
	if result != nil {
		t.Fatalf("expected nil, got %T", result)
	}
}

// ===== Additional: Nil/edge cases =====

func TestDesugarStatement_Nil(t *testing.T) {
	result := desugarStatement(nil)
	if result != nil {
		t.Fatalf("expected nil, got %T", result)
	}
}

func TestDesugarExpression_Nil(t *testing.T) {
	result := desugarExpression(nil)
	if result != nil {
		t.Fatalf("expected nil, got %T", result)
	}
}

func TestDesugarExpressionStatement_NilExpression(t *testing.T) {
	es := &ast.ExpressionStatement{Token: "nil", Expression: nil}
	result := desugarStatement(es)
	if result != nil {
		t.Fatalf("expected nil for nil expression, got %T", result)
	}
}

func TestDesugarExpressionStatement_NilSelf(t *testing.T) {
	var es *ast.ExpressionStatement = nil
	result := desugarStatement(es)
	if result != nil {
		t.Fatalf("expected nil, got %T", result)
	}
}

func TestDesugarExpression_DefaultCase(t *testing.T) {
	// Test with an expression type that hits the default case
	el := &ast.EllipsisLiteral{Token: "..."}
	result := desugarExpression(el)
	if result == nil {
		t.Fatal("expected non-nil result for EllipsisLiteral")
	}
	ellResult, ok := result.(*ast.EllipsisLiteral)
	if !ok {
		t.Fatalf("expected EllipsisLiteral, got %T", result)
	}
	if ellResult.Token != "..." {
		t.Fatalf("expected '...', got '%s'", ellResult.Token)
	}
}

func TestDesugarStatement_DefaultCase(t *testing.T) {
	// ImportStatement should hit the default case (returned as-is)
	is := &ast.ImportStatement{
		Token:  "import",
		Module: ident("os"),
	}
	result := desugarStatement(is)
	importResult, ok := result.(*ast.ImportStatement)
	if !ok {
		t.Fatalf("expected ImportStatement, got %T", result)
	}
	if importResult.Module.Value != "os" {
		t.Fatalf("expected 'os', got '%s'", importResult.Module.Value)
	}
}

// ===== Additional: Desugar (top-level) =====

func TestDesugar(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				Token: "let",
				Names: []*ast.Identifier{ident("x")},
				Value: intLit(5),
			},
			&ast.ForStatement{
				Token:    "for",
				Value:    ident("i"),
				Iterable: ident("range"),
				Body: &ast.BlockStatement{
					Token:      ":",
					Statements: []ast.Statement{&ast.ExpressionStatement{Token: "1", Expression: ident("i")}},
				},
			},
		},
	}

	result := Desugar(program)
	if len(result.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(result.Statements))
	}
	// First should be LetStatement
	if _, ok := result.Statements[0].(*ast.LetStatement); !ok {
		t.Fatalf("expected LetStatement, got %T", result.Statements[0])
	}
	// Second should be BlockStatement (for→while)
	if _, ok := result.Statements[1].(*ast.BlockStatement); !ok {
		t.Fatalf("expected BlockStatement, got %T", result.Statements[1])
	}
}

func TestDesugar_EmptyProgram(t *testing.T) {
	program := &ast.Program{Statements: []ast.Statement{}}
	result := Desugar(program)
	if len(result.Statements) != 0 {
		t.Fatalf("expected 0 statements, got %d", len(result.Statements))
	}
}

func TestDesugar_NilStatementsFiltered(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.BreakStatement{Token: "break"}, // desugars to nil
			&ast.LetStatement{Token: "let", Names: []*ast.Identifier{ident("x")}, Value: intLit(1)},
		},
	}

	result := Desugar(program)
	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement (break filtered), got %d", len(result.Statements))
	}
}

// ===== Additional: isComparisonOp =====

func TestIsComparisonOp(t *testing.T) {
	tests := []struct {
		op      string
		isComp  bool
	}{
		{"==", true},
		{"!=", true},
		{"<", true},
		{">", true},
		{"<=", true},
		{">=", true},
		{"+", false},
		{"-", false},
		{"and", false},
		{"or", false},
	}

	for _, tt := range tests {
		result := isComparisonOp(tt.op)
		if result != tt.isComp {
			t.Errorf("isComparisonOp(%q) = %v, want %v", tt.op, result, tt.isComp)
		}
	}
}

// ===== Additional: isEnumClass =====

func TestIsEnumClass_SuperClass(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:      "class",
		Name:       ident("MyEnum"),
		SuperClass: ident("Enum"),
	}
	if !isEnumClass(cs) {
		t.Fatal("expected class with Enum superclass to be recognized")
	}
}

func TestIsEnumClass_SuperClasses(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:        "class",
		Name:         ident("MyEnum"),
		SuperClass:   nil,
		SuperClasses: []*ast.Identifier{ident("Enum")},
	}
	if !isEnumClass(cs) {
		t.Fatal("expected class with Enum in SuperClasses to be recognized")
	}
}

func TestIsEnumClass_NotEnum(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:      "class",
		Name:       ident("MyClass"),
		SuperClass: ident("Base"),
	}
	if isEnumClass(cs) {
		t.Fatal("expected non-Enum class to NOT be recognized")
	}
}

// ===== Additional: collectPatternBindings =====

func TestCollectPatternBindings_Identifier(t *testing.T) {
	// collectPatternBindings(matchVar, pattern): pattern is identifier "x", so binding is x = matchVar
	bindings := collectPatternBindings(ident("_match_val"), ident("x"))
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
	assign := bindings[0].(*ast.AssignStatement)
	if assign.Names[0].Value != "x" {
		t.Fatalf("expected 'x', got '%s'", assign.Names[0].Value)
	}
}

func TestCollectPatternBindings_Wildcard(t *testing.T) {
	// collectPatternBindings(matchVar, pattern): pattern is wildcard "_", should produce 0 bindings
	bindings := collectPatternBindings(ident("_match_val"), ident("_"))
	if len(bindings) != 0 {
		t.Fatalf("expected 0 bindings for wildcard, got %d", len(bindings))
	}
}

func TestCollectPatternBindings_OrPattern(t *testing.T) {
	pattern := &ast.InfixExpression{
		Token:    "|",
		Left:     ident("x"),
		Operator: "|",
		Right:    intLit(2),
	}
	bindings := collectPatternBindings(ident("val"), pattern)
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding from or pattern, got %d", len(bindings))
	}
}

func TestCollectPatternBindings_CallPattern(t *testing.T) {
	pattern := &ast.CallExpression{
		Token:    "Point",
		Function: ident("Point"),
		Arguments: []ast.Expression{ident("x"), ident("y")},
	}
	bindings := collectPatternBindings(ident("val"), pattern)
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings from call pattern, got %d", len(bindings))
	}
}

func TestCollectPatternBindings_ListPattern(t *testing.T) {
	pattern := &ast.ListLiteral{
		Token:    "[",
		Elements: []ast.Expression{ident("a"), ident("b")},
	}
	bindings := collectPatternBindings(ident("val"), pattern)
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings from list pattern, got %d", len(bindings))
	}
}

func TestCollectPatternBindings_Nil(t *testing.T) {
	bindings := collectPatternBindings(ident("val"), nil)
	if len(bindings) != 0 {
		t.Fatalf("expected 0 bindings for nil pattern, got %d", len(bindings))
	}
}

// ===== Additional: buildMatchCondition =====

func TestBuildMatchCondition_Nil(t *testing.T) {
	cond := buildMatchCondition(ident("x"), nil)
	boolCond, ok := cond.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True for nil pattern, got %T", cond)
	}
}

func TestBuildMatchCondition_UnknownIdentifier(t *testing.T) {
	// An identifier that's not _, None, True, or False should match anything
	cond := buildMatchCondition(ident("x"), ident("myvar"))
	boolCond, ok := cond.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True for unknown identifier pattern, got %T", cond)
	}
}

func TestBuildMatchCondition_CallPatternNoArgs(t *testing.T) {
	pattern := &ast.CallExpression{
		Token:     "MyClass",
		Function:  ident("MyClass"),
		Arguments: []ast.Expression{},
	}
	cond := buildMatchCondition(ident("x"), pattern)
	callExpr, ok := cond.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression (isinstance), got %T", cond)
	}
	funcIdent := callExpr.Function.(*ast.Identifier)
	if funcIdent.Value != "isinstance" {
		t.Fatalf("expected isinstance call, got '%s'", funcIdent.Value)
	}
}

func TestBuildMatchCondition_CallPatternNotIdentifier(t *testing.T) {
	pattern := &ast.CallExpression{
		Token:    "call",
		Function: &ast.MemberAccess{Token: ".", Object: ident("obj"), Member: ident("method")},
	}
	cond := buildMatchCondition(ident("x"), pattern)
	boolCond, ok := cond.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True for non-identifier call pattern, got %T", cond)
	}
}

func TestBuildMatchCondition_InfixNotOr(t *testing.T) {
	pattern := &ast.InfixExpression{
		Token:    "+",
		Left:     ident("a"),
		Operator: "+",
		Right:    ident("b"),
	}
	cond := buildMatchCondition(ident("x"), pattern)
	boolCond, ok := cond.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True for non-or infix pattern, got %T", cond)
	}
}

func TestBuildMatchCondition_UnknownType(t *testing.T) {
	// A type not handled by buildMatchCondition should return True
	pattern := &ast.SetLiteral{Token: "{", Elements: []ast.Expression{intLit(1)}}
	cond := buildMatchCondition(ident("x"), pattern)
	boolCond, ok := cond.(*ast.Boolean)
	if !ok || !boolCond.Value {
		t.Fatalf("expected True for unknown pattern type, got %T", cond)
	}
}

// ===== Additional: Delete statement with unknown target type =====

func TestDesugarDeleteStatement_UnknownTarget(t *testing.T) {
	ds := &ast.DeleteStatement{
		Token:   "del",
		Targets: []ast.Expression{intLit(1)}, // unusual but should not crash
	}

	result := desugarStatement(ds)
	_, ok := result.(*ast.DeleteStatement)
	if !ok {
		t.Fatalf("expected DeleteStatement for unknown target, got %T", result)
	}
}

// ===== Additional: ExpressionStatement without decorator =====

func TestDesugarExpressionStatement_Plain(t *testing.T) {
	es := &ast.ExpressionStatement{
		Token:      "1",
		Expression: ident("x"),
	}

	result := desugarStatement(es)
	exprResult, ok := result.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", result)
	}
	if exprResult.Expression.(*ast.Identifier).Value != "x" {
		t.Fatal("expression should be desugared")
	}
}

func TestDesugarExpressionStatement_NilDesugared(t *testing.T) {
	es := &ast.ExpressionStatement{
		Token:      "1",
		Expression: &ast.NamedExpression{Token: ":=", Name: ident("x"), Value: nil},
	}

	// This shouldn't panic
	result := desugarStatement(es)
	// NamedExpression with nil value still returns NamedExpression
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ===== Additional: Enum with empty body =====

func TestDesugarEnum_EmptyBody(t *testing.T) {
	cs := &ast.ClassStatement{
		Token:      "class",
		Name:       ident("EmptyEnum"),
		SuperClass: ident("Enum"),
		Body:       nil,
	}

	result := desugarStatement(cs)
	assignResult, ok := result.(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", result)
	}
	callExpr := assignResult.Value.(*ast.CallExpression)
	hashLit := callExpr.Arguments[1].(*ast.HashLiteral)
	if len(hashLit.Pairs) != 0 {
		t.Fatalf("expected 0 pairs for empty enum, got %d", len(hashLit.Pairs))
	}
}

// ===== Additional: SliceAssignStatement with Step =====

func TestDesugarSliceAssignStatement_WithStep(t *testing.T) {
	sas := &ast.SliceAssignStatement{
		Token: "=",
		Left:  ident("lst"),
		Lower: intLit(0),
		Upper: intLit(10),
		Step:  intLit(2),
		Value: &ast.ListLiteral{Token: "[", Elements: []ast.Expression{intLit(1)}},
	}

	result := desugarStatement(sas)
	sliceResult := result.(*ast.SliceAssignStatement)
	if sliceResult.Step == nil {
		t.Fatal("step should be desugared")
	}
}

// ===== Additional: isLruCacheDecorator edge cases =====

func TestIsLruCacheDecorator_CallNotIdentifier(t *testing.T) {
	dec := &ast.CallExpression{
		Token:    "something",
		Function: &ast.MemberAccess{Token: ".", Object: ident("mod"), Member: ident("lru_cache")},
	}
	if !isLruCacheDecorator(dec) {
		t.Fatal("expected mod.lru_cache() to be recognized")
	}
}

func TestIsLruCacheDecorator_CallNotLruCache(t *testing.T) {
	dec := &ast.CallExpression{
		Token:    "cache",
		Function: ident("cache"),
	}
	if isLruCacheDecorator(dec) {
		t.Fatal("expected 'cache' to NOT be recognized as lru_cache")
	}
}

func TestIsLruCacheDecorator_CallMemberNotLruCache(t *testing.T) {
	dec := &ast.CallExpression{
		Token: "other",
		Function: &ast.MemberAccess{
			Token:  ".",
			Object: ident("mod"),
			Member: ident("other_cache"),
		},
	}
	if isLruCacheDecorator(dec) {
		t.Fatal("expected mod.other_cache() to NOT be recognized as lru_cache")
	}
}

func TestIsLruCacheDecorator_MemberNotLruCache(t *testing.T) {
	dec := &ast.MemberAccess{
		Token:  ".",
		Object: ident("mod"),
		Member: ident("other"),
	}
	if isLruCacheDecorator(dec) {
		t.Fatal("expected mod.other to NOT be recognized as lru_cache")
	}
}

func TestIsLruCacheDecorator_NotMatchingType(t *testing.T) {
	dec := &ast.IntegerLiteral{Token: "1", Value: 1}
	if isLruCacheDecorator(dec) {
		t.Fatal("expected integer to NOT be recognized as lru_cache")
	}
}

// ===== Additional: collectPatternBindings with wildcard in call =====

func TestCollectPatternBindings_CallWithWildcard(t *testing.T) {
	pattern := &ast.CallExpression{
		Token:    "Point",
		Function: ident("Point"),
		Arguments: []ast.Expression{ident("_"), ident("y")},
	}
	bindings := collectPatternBindings(ident("val"), pattern)
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding (wildcard excluded), got %d", len(bindings))
	}
}

// ===== Additional: Decorator with function that has no decorators (should not enter decorator path) =====

func TestDesugarExpressionStatement_FunctionNoDecorators(t *testing.T) {
	fnLit := &ast.FunctionLiteral{
		Token:      "def",
		Name:       "foo",
		Parameters: []*ast.Identifier{},
		Body:       &ast.BlockStatement{Token: ":", Statements: []ast.Statement{}},
		Decorators: []ast.Expression{}, // empty decorators
	}

	es := &ast.ExpressionStatement{
		Token:      "def",
		Expression: fnLit,
	}

	result := desugarStatement(es)
	exprResult, ok := result.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", result)
	}
	// Should be treated as a normal expression, not decorator path
	_, ok = exprResult.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected FunctionLiteral, got %T", exprResult.Expression)
	}
}
