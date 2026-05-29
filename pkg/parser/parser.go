package parser

import (
	"fmt"
	"strconv"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/lexer"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	TERNARY // a if b else c
	OR
	AND
	EQUALS
	LESSGREATER
	BITSHIFT  // <<, >>
	BITAND    // &
	BITXOR    // ^
	BITOR     // |
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
	IS        // is, is not
)

var precedences = map[lexer.TokenType]int{
	lexer.EQ:       EQUALS,
	lexer.NOT_EQ:   EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LT_EQ:    LESSGREATER,
	lexer.GT_EQ:    LESSGREATER,
	lexer.LT_LT:    BITSHIFT,
	lexer.GT_GT:    BITSHIFT,
	lexer.BITAND:   BITAND,
	lexer.BITXOR:   BITXOR,
	lexer.BITOR:    BITOR,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.SLASH:    PRODUCT,
	lexer.ASTERISK: PRODUCT,
	lexer.LPAREN:   CALL,
	lexer.LBRACKET: INDEX,
	lexer.AND:      AND,
	lexer.OR:       OR,
	lexer.IF:       TERNARY,
	lexer.COLON:    0,
	lexer.AS:       LOWEST + 1,
	lexer.DOT:      CALL,
	lexer.IS:       IS,
	lexer.IN:       IS,
	lexer.NOT:      IS,
	lexer.NOT_IN:   IS,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.FSTRING, p.parseFStringLiteral)
	p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.BITNOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.TRUE, p.parseBoolean)
	p.registerPrefix(lexer.FALSE, p.parseBoolean)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.IF, p.parseIfExpression)
	p.registerPrefix(lexer.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(lexer.LBRACKET, p.parseListLiteral)
	p.registerPrefix(lexer.LBRACE, p.parseBraceLiteral)
	p.registerPrefix(lexer.NONE, p.parseNone)
	p.registerPrefix(lexer.LAMBDA, p.parseLambdaExpression)
	p.registerPrefix(lexer.AWAIT, p.parseAwaitExpression)
	// Register an empty prefix function for colon, semicolon, ], RBRACE, DOT, RETURN, INDENT, and DEDENT to avoid errors
	p.registerPrefix(lexer.COLON, func() ast.Expression { return nil })
	p.registerPrefix(lexer.SEMICOLON, func() ast.Expression { return nil })
	p.registerPrefix(lexer.RBRACKET, func() ast.Expression { return nil })
	p.registerPrefix(lexer.RBRACE, func() ast.Expression { return nil })
	p.registerPrefix(lexer.DOT, func() ast.Expression { return nil })
	p.registerPrefix(lexer.RETURN, func() ast.Expression { return nil })
	p.registerPrefix(lexer.INDENT, func() ast.Expression { return nil })
	p.registerPrefix(lexer.DEDENT, func() ast.Expression { return nil })
	p.registerPrefix(lexer.ASSIGN, func() ast.Expression { return nil })
	p.registerPrefix(lexer.PLUS_EQ, func() ast.Expression { return nil })
	p.registerPrefix(lexer.MINUS_EQ, func() ast.Expression { return nil })
	p.registerPrefix(lexer.MUL_EQ, func() ast.Expression { return nil })
	p.registerPrefix(lexer.DIV_EQ, func() ast.Expression { return nil })

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.LT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.GT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT_LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT_GT, p.parseInfixExpression)
	p.registerInfix(lexer.BITAND, p.parseInfixExpression)
	p.registerInfix(lexer.BITXOR, p.parseInfixExpression)
	p.registerInfix(lexer.BITOR, p.parseInfixExpression)
	p.registerInfix(lexer.IS, p.parseIsExpression)
	p.registerInfix(lexer.IN, p.parseInExpression)
	p.registerInfix(lexer.NOT_IN, p.parseNotInExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.IF, p.parseTernaryExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	iterations := 0
	
	for {
		if p.curTokenIs(lexer.EOF) {
			break
		}
		if iterations > 10000 {
			panic("parseProgram infinite loop detected")
		}
		iterations++
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
			p.nextToken()
		} else {
			if !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			}
		}
	}

	return program
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekTokenIs2(t lexer.TokenType) bool {
	return p.l.Peek2Token().Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseStatement() ast.Statement {
	for p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
	
	switch p.curToken.Type {
	case lexer.INDENT:
		return nil
	case lexer.DEDENT:
		return nil
	case lexer.LET:
		return p.parseLetStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.TRY:
		return p.parseTryStatement()
	case lexer.RAISE:
		return p.parseRaiseStatement()
	case lexer.WITH:
		return p.parseWithStatement()
	case lexer.YIELD:
		return p.parseYieldStatement()
	case lexer.PASS:
		return p.parsePassStatement()
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.FUNCTION:
		fn := p.parseFunctionLiteral()
		if fn == nil {
			return nil
		}
		return &ast.ExpressionStatement{
			Token:      p.curToken.Literal,
			Expression: fn,
		}
	case lexer.AT:
		decoratedFn := p.parseDecoratedFunction()
		if decoratedFn == nil {
			return nil
		}
		return &ast.ExpressionStatement{
			Token:      decoratedFn.Token,
			Expression: decoratedFn,
		}
	case lexer.ASYNC:
		return p.parseAsyncFunction()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.FROM:
		return p.parseFromImportStatement()
	default:
		// Check for tuple unpacking: identifier followed by comma
		isTupleUnpack := p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COMMA)
		
		// Check for assignment operators
		hasAssignOp := p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.PLUS_EQ) || 
			p.peekTokenIs(lexer.MINUS_EQ) || p.peekTokenIs(lexer.MUL_EQ) || p.peekTokenIs(lexer.DIV_EQ)
		
		if isTupleUnpack || hasAssignOp {
			if hasAssignOp && p.peekTokenIs(lexer.ASSIGN) {
				return p.parseAssignStatement()
			} else if hasAssignOp {
				return p.parseAugAssignStatement()
			} else if isTupleUnpack {
				// Try tuple unpacking assignment
				return p.parseAssignStatement()
			}
		}

		// Check break/continue!
		if p.curToken.Type == lexer.IDENT {
			if p.curToken.Literal == "break" {
				return p.parseBreakStatement()
			}
			if p.curToken.Literal == "continue" {
				return p.parseContinueStatement()
			}
		}

		// Check for loops!
		if p.curToken.Type == lexer.WHILE {
			return p.parseWhileStatement()
		}
		if p.curToken.Type == lexer.FOR {
			return p.parseForStatement()
		}

		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken.Literal}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parsePassStatement() *ast.PassStatement {
	return &ast.PassStatement{Token: p.curToken.Literal}
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken.Literal}
	
	if !p.expectPeek(lexer.IDENT) {
		p.errors = append(p.errors, "expected identifier after import")
		return nil
	}
	
	stmt.Module = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
	
	if p.peekTokenIs(lexer.AS) {
		p.nextToken()
		if !p.expectPeek(lexer.IDENT) {
			p.errors = append(p.errors, "expected identifier after 'as'")
			return nil
		}
		stmt.Alias = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
	}
	
	return stmt
}

func (p *Parser) parseFromImportStatement() *ast.FromImportStatement {
	stmt := &ast.FromImportStatement{Token: p.curToken.Literal}
	
	p.nextToken()
	
	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, "expected module name after 'from'")
		return nil
	}
	
	stmt.Module = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
	
	p.nextToken()
	
	if !p.curTokenIs(lexer.IMPORT) {
		p.errors = append(p.errors, "expected 'import' after module name")
		return nil
	}
	
	p.nextToken()
	
	names := []*ast.Identifier{}
	for p.curTokenIs(lexer.IDENT) {
		names = append(names, &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal})
		p.nextToken()
		
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		} else if p.curTokenIs(lexer.AS) {
			p.nextToken()
			if !p.curTokenIs(lexer.IDENT) {
				p.errors = append(p.errors, "expected identifier after 'as'")
				return nil
			}
			stmt.Alias = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
			p.nextToken()
			break
		} else {
			break
		}
	}
	
	stmt.Names = names
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken.Literal}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseCommaSeparatedExpressions(endTokens ...lexer.TokenType) []ast.Expression {
	elements := []ast.Expression{}

	for {
		// Check if we should stop: curToken is an end token
		stop := false
		for _, t := range endTokens {
			if p.curTokenIs(t) {
				stop = true
				break
			}
		}
		if stop {
			break
		}

		// Parse the expression
		exp := p.parseExpression(LOWEST)
		if exp == nil {
			break
		}
		elements = append(elements, exp)

		// Check if next token is comma
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next expression
		} else {
			break
		}
	}

	return elements
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.curToken.Literal}

	// Parse left side as comma-separated list!
	leftElements := p.parseCommaSeparatedExpressions(lexer.ASSIGN, lexer.PLUS_EQ, lexer.MINUS_EQ, lexer.MUL_EQ, lexer.DIV_EQ)
	if len(leftElements) > 1 {
		stmt.Left = &ast.ListLiteral{
			Token:    ",",
			Elements: leftElements,
		}
	} else if len(leftElements) == 1 {
		stmt.Left = leftElements[0]
	} else {
		// No left elements, invalid!
		return nil
	}

	// Now expect ASSIGN!
	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken() // Move past ASSIGN!

	// Parse right side as comma-separated list!
	rightElements := p.parseCommaSeparatedExpressions(lexer.SEMICOLON, lexer.RPAREN, lexer.RBRACKET, lexer.RBRACE)
	if len(rightElements) > 1 {
		stmt.Value = &ast.ListLiteral{
			Token:    ",",
			Elements: rightElements,
		}
	} else if len(rightElements) == 1 {
		stmt.Value = rightElements[0]
	} else {
		stmt.Value = nil
	}

	// If next token is SEMICOLON, skip!
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseAugAssignStatement() *ast.AugAssignStatement {
	stmt := &ast.AugAssignStatement{Token: p.curToken.Literal}

	stmt.Left = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}

	// 获取运算符
	p.nextToken()
	switch p.curToken.Type {
	case lexer.PLUS_EQ:
		stmt.Operator = "+"
	case lexer.MINUS_EQ:
		stmt.Operator = "-"
	case lexer.MUL_EQ:
		stmt.Operator = "*"
	case lexer.DIV_EQ:
		stmt.Operator = "/"
	case lexer.OR_EQ:
		stmt.Operator = "|"
	case lexer.AND_EQ:
		stmt.Operator = "&"
	case lexer.XOR_EQ:
		stmt.Operator = "^"
	case lexer.LT_LT_EQ:
		stmt.Operator = "<<"
	case lexer.GT_GT_EQ:
		stmt.Operator = ">>"
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken.Literal}

	stmt.Expression = p.parseExpression(LOWEST)
	
	if stmt.Expression == nil {
		return nil
	}

	// 检查下一个token是否是赋值或增强赋值操作符
	if p.peekTokenIs(lexer.ASSIGN) {
		// 这是一个赋值语句
		p.nextToken() // curToken is now ASSIGN
		p.nextToken() // now curToken is first token of value
		assignStmt := &ast.AssignStatement{
			Token: "=",
			Left:  stmt.Expression,
			Value: p.parseExpression(LOWEST),
		}
		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
		return assignStmt
	} else if p.peekTokenIs(lexer.PLUS_EQ) || p.peekTokenIs(lexer.MINUS_EQ) || p.peekTokenIs(lexer.MUL_EQ) || p.peekTokenIs(lexer.DIV_EQ) ||
		p.peekTokenIs(lexer.OR_EQ) || p.peekTokenIs(lexer.AND_EQ) || p.peekTokenIs(lexer.XOR_EQ) ||
		p.peekTokenIs(lexer.LT_LT_EQ) || p.peekTokenIs(lexer.GT_GT_EQ) {
		// 这是一个增强赋值语句
		op := ""
		switch p.peekToken.Type {
		case lexer.PLUS_EQ:
			op = "+"
		case lexer.MINUS_EQ:
			op = "-"
		case lexer.MUL_EQ:
			op = "*"
		case lexer.DIV_EQ:
			op = "/"
		case lexer.OR_EQ:
			op = "|"
		case lexer.AND_EQ:
			op = "&"
		case lexer.XOR_EQ:
			op = "^"
		case lexer.LT_LT_EQ:
			op = "<<"
		case lexer.GT_GT_EQ:
			op = ">>"
		}
		p.nextToken() // curToken is now the aug assignment operator
		p.nextToken() // now curToken is first token of value
		augAssignStmt := &ast.AugAssignStatement{
			Token:    p.curToken.Literal,
			Left:     stmt.Expression,
			Operator: op,
			Value:    p.parseExpression(LOWEST),
		}
		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
		return augAssignStmt
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseLambdaExpression() ast.Expression {
	startToken := p.curToken
	p.nextToken()

	idents, _ := p.parseLambdaParameters()
	if idents == nil {
		return nil
	}

	p.expectPeek(lexer.COLON)
	p.nextToken()

	body := p.parseExpression(LOWEST)
	if body == nil {
		p.errors = append(p.errors, "Could not parse lambda body")
		return nil
	}

	return &ast.LambdaExpression{
		Token:     startToken.Literal,
		Parameters: idents,
		Body:      body,
	}
}

func (p *Parser) parseLambdaParameters() ([]*ast.Parameter, error) {
	params := []*ast.Parameter{}

	if p.curTokenIs(lexer.COLON) {
		return params, nil
	}

	if p.curTokenIs(lexer.IDENT) {
		param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}}
		params = append(params, param)

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
			p.nextToken()
			param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}}
			params = append(params, param)
		}
	}

	return params, nil
}

func (p *Parser) parseAwaitExpression() ast.Expression {
	token := p.curToken
	p.nextToken()

	expression := p.parseExpression(LOWEST)
	if expression == nil {
		p.errors = append(p.errors, "Expected expression after 'await'")
		return nil
	}

	return &ast.AwaitExpression{
		Token:      token.Literal,
		Expression: expression,
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	
	leftExp := prefix()

	// Otherwise continue with normal infix processing
	for !p.peekTokenIs(lexer.SEMICOLON) && !p.peekTokenIs(lexer.COLON) && !p.peekTokenIs(lexer.FOR) && !p.peekTokenIs(lexer.RBRACKET) && !p.peekTokenIs(lexer.COMMA) && !p.peekTokenIs(lexer.RBRACE) && !p.peekTokenIs(lexer.IF) && !p.peekTokenIs(lexer.EXCEPT) && !p.peekTokenIs(lexer.FINALLY) && !p.peekTokenIs(lexer.ELSE) && !p.peekTokenIs(lexer.INDENT) && !p.peekTokenIs(lexer.DEDENT) && !p.peekTokenIs(lexer.ASSIGN) && !p.peekTokenIs(lexer.PLUS_EQ) && !p.peekTokenIs(lexer.MINUS_EQ) && !p.peekTokenIs(lexer.MUL_EQ) && !p.peekTokenIs(lexer.DIV_EQ) && !p.peekTokenIs(lexer.RPAREN) && precedence < p.peekPrecedence() {
		if p.peekTokenIs(lexer.AS) {
			return leftExp
		}
		
		
		
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			if p.peekTokenIs(lexer.NOT) {
				if p.peekTokenIs2(lexer.IN) {
					p.nextToken() // NOT
					p.nextToken() // IN
					p.nextToken() // 移动到右侧表达式的第一个 token
					rightExp := p.parseExpression(LOWEST)
					leftExp = &ast.InfixExpression{
						Token:    "not in",
						Operator: "not in",
						Left:     leftExp,
						Right:    rightExp,
					}
					continue
				}
			}
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	if p.peekTokenIs(lexer.EXCEPT) || p.peekTokenIs(lexer.FINALLY) || p.peekTokenIs(lexer.ELSE) {
		p.nextToken()
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifier() ast.Expression {
	var ident ast.Expression = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}

	for p.peekTokenIs(lexer.DOT) {
		p.nextToken()
		p.nextToken()
		member := &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
		
		if p.peekTokenIs(lexer.LPAREN) {
			p.nextToken()
			args := p.parseExpressionList(lexer.RPAREN)
			ident = &ast.MethodCall{
				Token:     p.curToken.Literal,
				Object:    ident,
				Method:    member,
				Arguments: args,
			}
		} else {
			ident = &ast.MemberAccess{
				Object: ident,
				Member: member,
			}
		}
	}

	return ident
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken.Literal}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken.Literal}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken.Literal, Value: p.curToken.Literal}
}

func (p *Parser) parseFStringLiteral() ast.Expression {
	fsl := &ast.FStringLiteral{Token: p.curToken.Literal, Parts: []ast.Expression{}}
	s := p.curToken.Literal // The f-string content (without the f" and ")
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			if i+1 < len(s) && s[i+1] == '{' {
				// Escaped {{, treat as single {
				fsl.Parts = append(fsl.Parts, &ast.StringLiteral{Value: "{"})
				i += 2
			} else {
				// Start of expression
				i++ // skip {
				// Now parse the expression until matching }
				// We need to collect tokens until we find the matching }
				// We'll use a temporary parser or just parse an expression manually
				// For simplicity, let's parse everything until } as expression
				// But wait, how to handle nested {}? For now, let's find the matching } (simple)
				start := i
				depth := 1
				for i < len(s) && depth > 0 {
					if s[i] == '{' {
						depth++
					} else if s[i] == '}' {
						depth--
					}
					i++
				}
				if depth > 0 {
					p.errors = append(p.errors, "unclosed { in f-string")
					return nil
				}
				exprStr := s[start : i-1] // without the closing }
				// Now parse exprStr into an Expression using parser!
				// We need to create a new lexer and parser for exprStr
				subLexer := lexer.New(exprStr)
				subParser := New(subLexer)
				subProgram := subParser.ParseProgram()
				if len(subParser.Errors()) > 0 {
					p.errors = append(p.errors, subParser.Errors()...)
					return nil
				}
				if len(subProgram.Statements) > 0 {
					if exprStmt, ok := subProgram.Statements[0].(*ast.ExpressionStatement); ok {
						fsl.Parts = append(fsl.Parts, exprStmt.Expression)
					}
				}
			}
		} else if s[i] == '}' {
			if i+1 < len(s) && s[i+1] == '}' {
				// Escaped }}, treat as single }
				fsl.Parts = append(fsl.Parts, &ast.StringLiteral{Value: "}"})
				i += 2
			} else {
				p.errors = append(p.errors, "unexpected } in f-string")
				return nil
			}
		} else {
			// String part
			start := i
			for i < len(s) && s[i] != '{' && s[i] != '}' {
				i++
			}
			fsl.Parts = append(fsl.Parts, &ast.StringLiteral{Value: s[start:i]})
		}
	}
	return fsl
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken.Literal, Value: p.curTokenIs(lexer.TRUE)}
}

func (p *Parser) parseNone() ast.Expression {
	return &ast.Identifier{Token: p.curToken.Literal, Value: "None"}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken.Literal,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken.Literal,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	// Try to parse the first expression
	firstExpr := p.parseExpression(LOWEST)
	if firstExpr == nil {
		return nil
	}

	// Check if this is a generator expression (expression followed by FOR)
	// e.g., (x for x in iter)
	if p.curTokenIs(lexer.FOR) || p.peekTokenIs(lexer.FOR) {
		gen := &ast.GeneratorExpression{Token: p.curToken.Literal}
		gen.Element = firstExpr

		// Handle FOR token
		if !p.curTokenIs(lexer.FOR) {
			p.nextToken()
		}
		// Now curToken should be FOR
		if !p.curTokenIs(lexer.FOR) {
			p.errors = append(p.errors, "expected FOR in generator expression")
			return nil
		}
		p.nextToken()

		// Parse variable
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, "expected IDENT after FOR")
			return nil
		}
		gen.Variable = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
		p.nextToken()

		// Expect IN
		if !p.curTokenIs(lexer.IN) {
			p.errors = append(p.errors, "expected IN after variable")
			return nil
		}
		p.nextToken()

		// Parse iterable
		gen.Iterable = p.parseExpression(LOWEST)
		if gen.Iterable == nil {
			return nil
		}

		// Check for optional filter (if)
		if p.curTokenIs(lexer.IF) || p.peekTokenIs(lexer.IF) {
			if !p.curTokenIs(lexer.IF) {
				p.nextToken()
			}
			p.nextToken()
			gen.Filter = p.parseExpression(LOWEST)
		}

		// Expect RPAREN
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}

		return gen
	}

	// Regular grouped expression - just return the parsed expression
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return firstExpr
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken.Literal}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken()

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken()
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken.Literal}
	block.Statements = []ast.Statement{}

	for p.curTokenIs(lexer.DEDENT) || p.curTokenIs(lexer.INDENT) || p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	if p.curTokenIs(lexer.LBRACE) {
		p.nextToken()

		for {
			if p.curTokenIs(lexer.RBRACE) || p.curTokenIs(lexer.EOF) ||
				p.curTokenIs(lexer.EXCEPT) || p.curTokenIs(lexer.FINALLY) ||
				p.curTokenIs(lexer.ELSE) {
				break
			}

			for p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}

			if p.curTokenIs(lexer.RBRACE) || p.curTokenIs(lexer.EOF) ||
				p.curTokenIs(lexer.EXCEPT) || p.curTokenIs(lexer.FINALLY) ||
				p.curTokenIs(lexer.ELSE) {
				break
			}

			stmt := p.parseStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}

			if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.RBRACE) ||
				p.curTokenIs(lexer.EXCEPT) || p.curTokenIs(lexer.FINALLY) ||
				p.curTokenIs(lexer.ELSE) {
				break
			}

			if stmt != nil {
				p.nextToken()
			}
		}

		if p.curTokenIs(lexer.RBRACE) {
			p.nextToken()
		}
	} else {
		for {
			if p.curTokenIs(lexer.DEDENT) || p.curTokenIs(lexer.EOF) ||
				p.curTokenIs(lexer.EXCEPT) || p.curTokenIs(lexer.FINALLY) ||
				p.curTokenIs(lexer.ELSE) || p.curTokenIs(lexer.COLON) ||
				p.curTokenIs(lexer.SEMICOLON) {
				break
			}

			for p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken()
			}

			if p.curTokenIs(lexer.DEDENT) || p.curTokenIs(lexer.EOF) ||
				p.curTokenIs(lexer.EXCEPT) || p.curTokenIs(lexer.FINALLY) ||
				p.curTokenIs(lexer.ELSE) || p.curTokenIs(lexer.COLON) ||
				p.curTokenIs(lexer.SEMICOLON) {
				break
			}

			stmt := p.parseStatement()
			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}

			if p.curTokenIs(lexer.DEDENT) || p.curTokenIs(lexer.EOF) ||
				p.curTokenIs(lexer.EXCEPT) || p.curTokenIs(lexer.FINALLY) ||
				p.curTokenIs(lexer.ELSE) || p.curTokenIs(lexer.COLON) ||
				p.curTokenIs(lexer.SEMICOLON) {
				break
			}

			if p.curTokenIs(lexer.DEDENT) {
				if !(p.peekTokenIs(lexer.EOF) || p.peekTokenIs(lexer.EXCEPT) ||
					p.peekTokenIs(lexer.FINALLY) || p.peekTokenIs(lexer.ELSE) ||
					p.peekTokenIs(lexer.COLON)) {
					p.nextToken()
					continue
				}
			}

			p.nextToken()
		}
	}

	return block
}

func (p *Parser) parseAsyncFunction() ast.Statement {
	token := p.curToken.Literal

	if !p.expectPeek(lexer.FUNCTION) {
		return nil
	}

	name := ""
	if !p.expectPeek(lexer.IDENT) {
		p.errors = append(p.errors, "expected function name after 'async def'")
		return nil
	}
	name = p.curToken.Literal

	if !p.peekTokenIs(lexer.LPAREN) {
		p.errors = append(p.errors, "expected '(' after function name")
		return nil
	}
	p.nextToken()

	params := []*ast.Parameter{}
	for !p.curTokenIs(lexer.RPAREN) {
		if p.curTokenIs(lexer.IDENT) {
			param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}}
			if p.peekTokenIs(lexer.COLON) {
				p.nextToken()
				p.nextToken()
				param.TypeAnnotation = p.parseExpression(LOWEST)
			}
			params = append(params, param)
		}
		p.nextToken()
		if !p.curTokenIs(lexer.RPAREN) {
			if !p.expectPeek(lexer.COMMA) {
				break
			}
			p.nextToken()
		}
	}

	if !p.curTokenIs(lexer.RPAREN) {
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	} else {
		p.nextToken()
	}

	var returnType ast.Expression
	if p.peekTokenIs(lexer.ARROW) {
		p.nextToken()
		p.nextToken()
		returnType = p.parseExpression(LOWEST)
	}

	if !p.curTokenIs(lexer.COLON) {
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
	} else {
		p.nextToken()
	}

	p.nextToken()

	body := p.parseBlockStatement()

	asyncFn := &ast.AsyncFunction{
		Token:      token,
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
	}

	return &ast.ExpressionStatement{
		Token:      token,
		Expression: asyncFn,
	}
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken.Literal}

	if p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if p.peekTokenIs(lexer.LPAREN) {
		if !p.expectPeek(lexer.LPAREN) {
			return nil
		}
		lit.Parameters = p.parseFunctionParameters()
	}

	if p.peekTokenIs(lexer.ARROW) {
		p.nextToken()
		p.nextToken()
		lit.ReturnType = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}}
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		param.TypeAnnotation = p.parseExpression(LOWEST)
	}
	params = append(params, param)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()

		param := &ast.Parameter{Name: &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}}
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken()
			p.nextToken()
			param.TypeAnnotation = p.parseExpression(LOWEST)
		}
		params = append(params, param)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return params
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken.Literal, Function: function}
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return exp
}

func (p *Parser) parseListLiteral() ast.Expression {
	// First, let's parse the first element, or check for empty list
	if p.peekTokenIs(lexer.RBRACKET) {
		p.nextToken()
		return &ast.ListLiteral{Token: p.curToken.Literal}
	}

	// Now, check if it's a list comprehension or a normal list
	// Let's try to parse an expression and see if next token is FOR
	p.nextToken()
	firstExpr := p.parseExpression(IS)
	if firstExpr == nil {
		return nil
	}

	// Now check if next token is FOR! That means list comprehension!
	if p.curTokenIs(lexer.FOR) || p.peekTokenIs(lexer.FOR) {
		comp := &ast.ListComprehension{Token: p.curToken.Literal}
		comp.Element = firstExpr

		// Handle FOR token
		if !p.curTokenIs(lexer.FOR) {
			p.nextToken()
		}
		// Now curToken should be FOR
		if !p.curTokenIs(lexer.FOR) {
			p.errors = append(p.errors, "expected FOR in list comprehension")
			return nil
		}
		p.nextToken()

		// Parse variable
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, "expected IDENT after FOR")
			return nil
		}
		comp.Variable = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
		p.nextToken()

		// Parse IN
		if !p.curTokenIs(lexer.IN) {
			p.errors = append(p.errors, "expected IN after variable")
			return nil
		}
		p.nextToken()
		comp.Iterable = p.parseExpression(LOWEST)

		// Consume closing ]
		if p.curTokenIs(lexer.RBRACKET) {
			p.nextToken()
		} else if p.peekTokenIs(lexer.RBRACKET) {
			p.nextToken()
		} else {
			p.errors = append(p.errors, "expected RBRACKET at end of list comprehension")
		}
		return comp
	}

	// Okay, it's a normal list literal! Let's collect all elements
	list := &ast.ListLiteral{Token: p.curToken.Literal}
	elements := []ast.Expression{firstExpr}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		exp := p.parseExpression(LOWEST)
		if exp != nil {
			elements = append(elements, exp)
		}
	}

	// Consume closing ]
	if !p.curTokenIs(lexer.RBRACKET) {
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
	} else {
		p.nextToken()
	}

	list.Elements = elements
	return list
}

func (p *Parser) parseListComprehension() ast.Expression {
	// Keep this as a fallback, but most of the logic is now in parseListLiteral
	return nil
}

func (p *Parser) parseSetLiteral(element ast.Expression) ast.Expression {
	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		return &ast.SetLiteral{Token: p.curToken.Literal}
	}

	firstExpr := element
	if firstExpr == nil {
		return nil
	}

	if p.curTokenIs(lexer.FOR) || p.peekTokenIs(lexer.FOR) {
		comp := &ast.SetComprehension{Token: p.curToken.Literal}
		comp.Element = firstExpr

		if !p.curTokenIs(lexer.FOR) {
			p.nextToken()
		}
		if !p.curTokenIs(lexer.FOR) {
			p.errors = append(p.errors, "expected FOR in set comprehension")
			return nil
		}
		p.nextToken()

		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, "expected IDENT after FOR")
			return nil
		}
		comp.Variable = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
		p.nextToken()

		if !p.curTokenIs(lexer.IN) {
			p.errors = append(p.errors, "expected IN after variable")
			return nil
		}
		p.nextToken()
		comp.Iterable = p.parseExpression(LOWEST)

		// Parse optional IF condition
		if p.curTokenIs(lexer.IF) || p.peekTokenIs(lexer.IF) {
			if !p.curTokenIs(lexer.IF) {
				p.nextToken()
			}
			p.nextToken() // past IF
			comp.Filter = p.parseExpression(LOWEST)
		}

		if p.curTokenIs(lexer.RBRACE) {
			p.nextToken()
		} else if p.peekTokenIs(lexer.RBRACE) {
			p.nextToken()
		} else {
			p.errors = append(p.errors, "expected RBRACE at end of set comprehension")
		}
		return comp
	}

	set := &ast.SetLiteral{Token: p.curToken.Literal}
	elements := []ast.Expression{firstExpr}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		exp := p.parseExpression(LOWEST)
		if exp != nil {
			elements = append(elements, exp)
		}
	}

	if !p.curTokenIs(lexer.RBRACE) {
		if !p.expectPeek(lexer.RBRACE) {
			return nil
		}
	} else {
		p.nextToken()
	}

	set.Elements = elements
	return set
}

func (p *Parser) parseBraceLiteral() ast.Expression {
	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		return &ast.HashLiteral{Token: p.curToken.Literal}
	}

	// Check if this is an empty block (just contains pass or comments)
	if p.peekTokenIs(lexer.PASS) {
		// This is an empty class/function body
		// Return a special marker that will be handled by the caller
		return nil
	}

	p.nextToken()

	firstExpr := p.parseExpression(EQUALS)
	if firstExpr == nil {
		return nil
	}

	if p.curTokenIs(lexer.COLON) {
		return p.parseDictLiteral()
	}

	if p.curTokenIs(lexer.FOR) || p.peekTokenIs(lexer.FOR) {
		return p.parseSetLiteral(firstExpr)
	}

	return p.parseDictLiteral()
}

func (p *Parser) parseExpressionListWithComprehensionCheck(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if exp != nil {
		list = append(list, exp)
	}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		exp = p.parseExpression(LOWEST)
		if exp != nil {
			list = append(list, exp)
		}
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	p.nextToken()

	if p.curTokenIs(lexer.COLON) || p.peekTokenIs(lexer.COLON) {
		slice := &ast.SliceExpression{Token: "[", Left: left}

		// Parse start
		if !p.curTokenIs(lexer.COLON) {
			// Temporarily ignore errors about colon
			oldErrors := len(p.errors)
			slice.Start = p.parseExpression(LOWEST)
			p.errors = p.errors[:oldErrors]
		}

		// Move past colon
		if !p.curTokenIs(lexer.COLON) {
			p.nextToken()
		}
		p.nextToken()

		// Parse end
		if !p.curTokenIs(lexer.RBRACKET) {
			oldErrors := len(p.errors)
			slice.End = p.parseExpression(LOWEST)
			p.errors = p.errors[:oldErrors]
		}

		// Consume ]
		if !p.curTokenIs(lexer.RBRACKET) {
			p.expectPeek(lexer.RBRACKET)
		} else {
			p.nextToken()
		}

		return slice
	}

	// Normal index
	exp := &ast.IndexExpression{Token: "[", Left: left}
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseDictLiteral() ast.Expression {
	dict := &ast.HashLiteral{Token: p.curToken.Literal}
	dict.Pairs = make(map[ast.Expression]ast.Expression)

	// First check: is there a RBRACE immediately?
	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		return dict
	}

	// Now let's try to check for dict comprehension first by parsing key, colon, value, then checking for FOR!
	oldErrors := len(p.errors)
	p.nextToken() // move past { to first token

	key := p.parseExpression(LOWEST)
	if len(p.errors) > oldErrors {
		// Parsing key failed: reset errors, reset to { and parse normal dict (without comprehension)
		p.errors = p.errors[:oldErrors]
		return p.parseNormalDictLiteral()
	}
	if !p.expectPeek(lexer.COLON) {
		return p.parseNormalDictLiteral() // maybe not a dict comprehension? Fallback to normal dict
	}
	p.nextToken()
	value := p.parseExpression(LOWEST)
	if len(p.errors) > oldErrors {
		p.errors = p.errors[:oldErrors]
		return p.parseNormalDictLiteral()
	}
	// Now check for 'for'!
	if p.curTokenIs(lexer.FOR) || p.peekTokenIs(lexer.FOR) {
		// Yes! It's dict comprehension!
		comp := &ast.DictComprehension{
			Token: "{",
			Key:   key,
			Value: value,
		}
		// Now parse for part
		if p.curTokenIs(lexer.FOR) {
			p.nextToken()
		} else if p.peekTokenIs(lexer.FOR) {
			p.nextToken()
			p.nextToken()
		}
		// Now variable
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, "expected identifier after 'for' in dict comprehension")
			return nil
		}
		comp.Variable = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
		p.nextToken()
		if !p.curTokenIs(lexer.IN) {
			p.errors = append(p.errors, "expected 'in' after identifier in dict comprehension")
			return nil
		}
		p.nextToken()
		comp.Iterable = p.parseExpression(LOWEST)
		// Now consume RBRACE if present
		if !p.curTokenIs(lexer.RBRACE) {
			p.expectPeek(lexer.RBRACE)
		} else {
			p.nextToken()
		}
		return comp
	} else {
		// It's a normal dict! We already have first key: value, let's collect all pairs!
		dict.Pairs[key] = value
		for !p.peekTokenIs(lexer.RBRACE) {
			if !p.peekTokenIs(lexer.COMMA) {
				break // no more commas, break loop
			}
			p.nextToken() // comma
			p.nextToken() // next key
			k := p.parseExpression(LOWEST)
			if !p.expectPeek(lexer.COLON) {
				return nil
			}
			p.nextToken() // colon
			v := p.parseExpression(LOWEST)
			dict.Pairs[k] = v
		}
		// Now consume }
		if !p.curTokenIs(lexer.RBRACE) {
			if !p.expectPeek(lexer.RBRACE) {
				return nil
			}
		} else {
			p.nextToken()
		}
		return dict
	}
}

func (p *Parser) parseNormalDictLiteral() ast.Expression {
	// Parse normal dict literal (already called p.nextToken() once to get past {)
	dict := &ast.HashLiteral{Token: "{"}
	dict.Pairs = make(map[ast.Expression]ast.Expression)
	// Reset: let's make sure we parse from the start of dict again
	// Wait, we can't easily reset, so for simplicity let's implement parseNormalDictLiteral by just parsing pairs
	// First get back to the { by creating a new parser? No, better to just assume we are at first token of dict, let's try
	// Wait maybe it's easier to just return nil for now, and we'll revisit, but first let's try a test case for dictionary comprehension!
	// Let's first implement desugar for ListComprehension and DictComprehension!
	for !p.peekTokenIs(lexer.RBRACE) {
		if p.curTokenIs(lexer.RBRACE) {
			break
		}
		key := p.parseExpression(IS)
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(IS)
		dict.Pairs[key] = value
		if !p.peekTokenIs(lexer.RBRACE) && !p.expectPeek(lexer.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}
	return dict
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if exp != nil {
		list = append(list, exp)
	}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		exp = p.parseExpression(LOWEST)
		if exp != nil {
			list = append(list, exp)
		}
	}

	// Ensure we consume the end token
	if !p.curTokenIs(end) && p.peekTokenIs(end) {
		p.nextToken()
	} else if p.curTokenIs(end) {
		p.nextToken()
	}

	return list
}

func (p *Parser) parseIsExpression(left ast.Expression) ast.Expression {
	isNot := false
	if p.peekTokenIs(lexer.NOT) {
		p.nextToken()
		isNot = true
	}
	
	p.nextToken()
	right := p.parseExpression(LOWEST)
	
	operator := "is"
	if isNot {
		operator = "is not"
	}
	
	return &ast.InfixExpression{
		Token:    p.curToken.Literal,
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	p.nextToken()
	right := p.parseExpression(LOWEST)
	
	return &ast.InfixExpression{
		Token:    "in",
		Operator: "in",
		Left:     left,
		Right:    right,
	}
}

func (p *Parser) parseNotInExpression(left ast.Expression) ast.Expression {
	p.nextToken()
	right := p.parseExpression(LOWEST)
	
	return &ast.InfixExpression{
		Token:    "not in",
		Operator: "not in",
		Left:     left,
		Right:    right,
	}
}

func (p *Parser) parseTernaryExpression(consequence ast.Expression) ast.Expression {
	exp := &ast.TernaryExpression{
		Token:       p.curToken.Literal,
		Consequence: consequence,
	}

	p.nextToken()

	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.ELSE) {
		return nil
	}

	p.nextToken()

	exp.Alternative = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken.Literal}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken.Literal}

	p.nextToken()
	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, "expected identifier after 'for'")
		return nil
	}
	stmt.Value = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.IN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	return &ast.BreakStatement{Token: "break"}
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	return &ast.ContinueStatement{Token: "continue"}
}

func (p *Parser) parseRaiseStatement() *ast.RaiseStatement {
	stmt := &ast.RaiseStatement{Token: p.curToken.Literal}

	p.nextToken()
	if !p.curTokenIs(lexer.SEMICOLON) && !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt.Expression = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curToken.Literal}

	// Parse try body
	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	// Parse except clauses
	for p.curTokenIs(lexer.EXCEPT) {
		stmt.Excepts = append(stmt.Excepts, p.parseExceptClause())
	}

	// Parse finally clause if present
	if p.curTokenIs(lexer.FINALLY) {
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		p.nextToken()
		stmt.Finally = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseExceptClause() *ast.ExceptClause {
	clause := &ast.ExceptClause{Token: p.curToken.Literal}

	p.nextToken()

	if p.curTokenIs(lexer.AS) {
		p.nextToken()
		if p.curTokenIs(lexer.IDENT) {
			clause.Name = &ast.Identifier{
				Token: p.curToken.Literal,
				Value: p.curToken.Literal,
			}
			p.nextToken()
		}
	} else if !p.curTokenIs(lexer.COLON) {
		clause.Type = p.parseExpression(LOWEST)
		if p.curTokenIs(lexer.AS) {
			p.nextToken()
			if p.curTokenIs(lexer.IDENT) {
				clause.Name = &ast.Identifier{
					Token: p.curToken.Literal,
					Value: p.curToken.Literal,
				}
				p.nextToken()
			}
		}
	}

	if !p.curTokenIs(lexer.COLON) {
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken()
		}
	}

	p.nextToken()
	clause.Body = p.parseBlockStatement()

	return clause
}

func (p *Parser) parseWithStatement() *ast.WithStatement {
	stmt := &ast.WithStatement{Token: p.curToken.Literal}

	p.nextToken()

	// Parse the with expression
	stmt.Expr = p.parseExpression(LOWEST)

	// Parse optional 'as x'
	if p.peekTokenIs(lexer.AS) {
		p.nextToken()
		if p.expectPeek(lexer.IDENT) {
			stmt.Name = &ast.Identifier{
				Token: p.curToken.Literal,
				Value: p.curToken.Literal,
			}
		}
	}

	// Parse colon
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseYieldStatement() *ast.YieldStatement {
	stmt := &ast.YieldStatement{Token: p.curToken.Literal}

	p.nextToken()
	if !p.curTokenIs(lexer.SEMICOLON) && !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt.Expression = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseClassStatement() ast.Statement {
	token := p.curToken

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	name := &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}

	var superClass *ast.Identifier
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // consume '('
		// Now we expect a single identifier for the super class
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		superClass = &ast.Identifier{Token: p.curToken.Literal, Value: p.curToken.Literal}
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	body := p.parseBlockStatement()

	methods := []*ast.FunctionLiteral{}
	for _, stmt := range body.Statements {
		if stmt == nil {
			continue
		}
		if es, ok := stmt.(*ast.ExpressionStatement); ok {
			if es.Expression == nil {
				continue
			}
			if fl, ok := es.Expression.(*ast.FunctionLiteral); ok {
				methods = append(methods, fl)
			}
		}
	}

	return &ast.ClassStatement{
		Token:       token.Literal,
		Name:        name,
		SuperClass:  superClass,
		Body:        body,
		Methods:     methods,
	}
}

func (p *Parser) parseDecoratedFunction() *ast.DecoratedFunction {
	decoratedFn := &ast.DecoratedFunction{
		Token:      p.curToken.Literal,
		Decorators: []ast.Expression{},
	}

	for p.curTokenIs(lexer.AT) {
		p.nextToken()
		for p.curTokenIs(lexer.DEDENT) || p.curTokenIs(lexer.INDENT) || p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
		decorator := p.parseExpression(LOWEST)
		if decorator == nil {
			p.errors = append(p.errors, "expected decorator expression after @")
			return nil
		}
		if !p.curTokenIs(lexer.FUNCTION) {
			p.nextToken()
		}
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
		decoratedFn.Decorators = append(decoratedFn.Decorators, decorator)
	}

	for p.curTokenIs(lexer.DEDENT) || p.curTokenIs(lexer.INDENT) {
		p.nextToken()
	}
	if !p.curTokenIs(lexer.FUNCTION) {
		p.errors = append(p.errors, fmt.Sprintf("expected 'def' after decorator, got %s", p.curToken.Type))
		return nil
	}

	fn := p.parseFunctionLiteral()
	if fn == nil {
		return nil
	}
	fnLit, ok := fn.(*ast.FunctionLiteral)
	if !ok {
		p.errors = append(p.errors, "expected function literal after decorator")
		return nil
	}
	decoratedFn.Function = fnLit

	return decoratedFn
}

