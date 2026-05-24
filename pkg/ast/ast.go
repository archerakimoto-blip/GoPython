package ast

import (
	"bytes"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type Identifier struct {
	Token string
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token string
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token }
func (il *IntegerLiteral) String() string       { return il.Token }

type FloatLiteral struct {
	Token string
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token }
func (fl *FloatLiteral) String() string       { return fl.Token }

type StringLiteral struct {
	Token string
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

type FStringLiteral struct {
	Token string
	Parts []Expression // Mix of StringLiterals and Expressions (for {expr} parts)
}

func (fsl *FStringLiteral) expressionNode()      {}
func (fsl *FStringLiteral) TokenLiteral() string { return fsl.Token }
func (fsl *FStringLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("f\"")
	for _, p := range fsl.Parts {
		if sl, ok := p.(*StringLiteral); ok {
			out.WriteString(sl.Value)
		} else {
			out.WriteString("{" + p.String() + "}")
		}
	}
	out.WriteString("\"")
	return out.String()
}

type Boolean struct {
	Token string
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token }
func (b *Boolean) String() string       { return b.Token }

type PrefixExpression struct {
	Token    string
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

type InfixExpression struct {
	Token    string
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

type IfExpression struct {
	Token       string
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(": ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString(" else: ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

type BlockStatement struct {
	Token      string
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
		out.WriteString("; ")
	}
	return out.String()
}

type FunctionLiteral struct {
	Token      string
	Parameters []*Identifier
	Body       *BlockStatement
	Name       string
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.Token)
	if fl.Name != "" {
		out.WriteString(" " + fl.Name)
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

type CallExpression struct {
	Token     string
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

type ListLiteral struct {
	Token    string
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token }
func (ll *ListLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range ll.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

type IndexExpression struct {
	Token string
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

type SliceExpression struct {
	Token string
	Left  Expression
	Start Expression
	End   Expression
}

func (se *SliceExpression) expressionNode()      {}
func (se *SliceExpression) TokenLiteral() string { return se.Token }
func (se *SliceExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(se.Left.String())
	out.WriteString("[")
	if se.Start != nil {
		out.WriteString(se.Start.String())
	}
	out.WriteString(":")
	if se.End != nil {
		out.WriteString(se.End.String())
	}
	out.WriteString("])")
	return out.String()
}

type DictLiteral struct {
	Token string
	Pairs map[Expression]Expression
}

func (dl *DictLiteral) expressionNode()      {}
func (dl *DictLiteral) TokenLiteral() string { return dl.Token }
func (dl *DictLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range dl.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

type ListComprehension struct {
	Token       string
	Element     Expression
	Variable    *Identifier
	Iterable    Expression
	Condition   Expression
}

func (lc *ListComprehension) expressionNode()      {}
func (lc *ListComprehension) TokenLiteral() string { return lc.Token }
func (lc *ListComprehension) String() string {
	var out bytes.Buffer
	out.WriteString("[")
	out.WriteString(lc.Element.String())
	out.WriteString(" for ")
	out.WriteString(lc.Variable.String())
	out.WriteString(" in ")
	out.WriteString(lc.Iterable.String())
	if lc.Condition != nil {
		out.WriteString(" if ")
		out.WriteString(lc.Condition.String())
	}
	out.WriteString("]")
	return out.String()
}

type DictComprehension struct {
	Token       string
	Key         Expression
	Value       Expression
	Variable    *Identifier
	Iterable    Expression
	Condition   Expression
}

func (dc *DictComprehension) expressionNode()      {}
func (dc *DictComprehension) TokenLiteral() string { return dc.Token }
func (dc *DictComprehension) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(dc.Key.String())
	out.WriteString(": ")
	out.WriteString(dc.Value.String())
	out.WriteString(" for ")
	out.WriteString(dc.Variable.String())
	out.WriteString(" in ")
	out.WriteString(dc.Iterable.String())
	if dc.Condition != nil {
		out.WriteString(" if ")
		out.WriteString(dc.Condition.String())
	}
	out.WriteString("}")
	return out.String()
}

type ExpressionStatement struct {
	Token      string
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type LetStatement struct {
	Token string
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.Token + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	return out.String()
}

type ReturnStatement struct {
	Token       string
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.Token + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	return out.String()
}

type AssignStatement struct {
	Token string
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token }
func (as *AssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	return out.String()
}

type AugAssignStatement struct {
	Token    string
	Name     *Identifier
	Operator string
	Value    Expression
}

func (aas *AugAssignStatement) statementNode()       {}
func (aas *AugAssignStatement) TokenLiteral() string { return aas.Token }
func (aas *AugAssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(aas.Name.String())
	out.WriteString(" " + aas.Operator + " ")
	if aas.Value != nil {
		out.WriteString(aas.Value.String())
	}
	return out.String()
}

type TernaryExpression struct {
	Token       string
	Consequence Expression
	Condition   Expression
	Alternative Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString(te.Consequence.String())
	out.WriteString(" if ")
	out.WriteString(te.Condition.String())
	out.WriteString(" else ")
	out.WriteString(te.Alternative.String())
	return out.String()
}

type WhileStatement struct {
	Token    string
	Condition Expression
	Body     *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(": ")
	out.WriteString(ws.Body.String())
	return out.String()
}

type ForStatement struct {
	Token    string
	Variable *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(fs.Variable.String())
	out.WriteString(" in ")
	out.WriteString(fs.Iterable.String())
	out.WriteString(": ")
	out.WriteString(fs.Body.String())
	return out.String()
}

type BreakStatement struct {
	Token string
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token }
func (bs *BreakStatement) String() string {
	return "break"
}

type ContinueStatement struct {
	Token string
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token }
func (cs *ContinueStatement) String() string {
	return "continue"
}

type SetLiteral struct {
	Token    string
	Elements []Expression
}

func (sl *SetLiteral) expressionNode()      {}
func (sl *SetLiteral) TokenLiteral() string { return sl.Token }
func (sl *SetLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range sl.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("}")
	return out.String()
}

type SetComprehension struct {
	Token     string
	Element   Expression
	Variable  *Identifier
	Iterable  Expression
	Condition Expression
}

func (sc *SetComprehension) expressionNode()      {}
func (sc *SetComprehension) TokenLiteral() string { return "{" }
func (sc *SetComprehension) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(sc.Element.String())
	out.WriteString(" for ")
	out.WriteString(sc.Variable.String())
	out.WriteString(" in ")
	out.WriteString(sc.Iterable.String())
	if sc.Condition != nil {
		out.WriteString(" if ")
		out.WriteString(sc.Condition.String())
	}
	out.WriteString("}")
	return out.String()
}

