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
	Parts []Expression
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
	return "(" + pe.Operator + pe.Right.String() + ")"
}

type InfixExpression struct {
	Token    string
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token }
func (oe *InfixExpression) String() string {
	return "(" + oe.Left.String() + " " + oe.Operator + " " + oe.Right.String() + ")"
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
	if ie.Alternative != nil {
		return "if " + ie.Condition.String() + " {\n" + ie.Consequence.String() + "\n} else {\n" + ie.Alternative.String() + "\n}"
	}
	return "if " + ie.Condition.String() + " {\n" + ie.Consequence.String() + "\n}"
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
	}
	return out.String()
}

type FunctionLiteral struct {
	Token      string
	Name       string
	Parameters []*Identifier
	Body       *BlockStatement
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
		out.WriteString("<" + fl.Name + ">")
	}
	out.WriteString("(" + strings.Join(params, ", ") + ") ")
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
	out.WriteString("(" + strings.Join(args, ", ") + ")")
	return out.String()
}

type ListLiteral struct {
	Token    string
	Elements []Expression
}

func (al *ListLiteral) expressionNode()      {}
func (al *ListLiteral) TokenLiteral() string { return al.Token }
func (al *ListLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
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
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}

type HashLiteral struct {
	Token string
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
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

type SliceExpression struct {
	Token string
	Left  Expression
	Start Expression
	End   Expression
}

func (se *SliceExpression) expressionNode()      {}
func (se *SliceExpression) TokenLiteral() string { return se.Token }
func (se *SliceExpression) String() string {
	var startStr, endStr string
	if se.Start != nil {
		startStr = se.Start.String()
	}
	if se.End != nil {
		endStr = se.End.String()
	}
	return "(" + se.Left.String() + "[" + startStr + ":" + endStr + "])"
}

type TernaryExpression struct {
	Token      string
	Condition  Expression
	Consequence Expression
	Alternative Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token }
func (te *TernaryExpression) String() string {
	return te.Condition.String() + " ? " + te.Consequence.String() + " : " + te.Alternative.String()
}

type LambdaExpression struct {
	Token      string
	Parameters []*Identifier
	Body       Expression
}

func (le *LambdaExpression) expressionNode()      {}
func (le *LambdaExpression) TokenLiteral() string { return le.Token }
func (le *LambdaExpression) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range le.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("lambda ")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(": ")
	out.WriteString(le.Body.String())
	return out.String()
}

type ListComprehension struct {
	Token    string
	Element  Expression
	Variable *Identifier
	Iterable Expression
	Filter   Expression
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
	if lc.Filter != nil {
		out.WriteString(" if ")
		out.WriteString(lc.Filter.String())
	}
	out.WriteString("]")
	return out.String()
}

type SetComprehension struct {
	Token    string
	Element  Expression
	Variable *Identifier
	Iterable Expression
	Filter   Expression
}

func (sc *SetComprehension) expressionNode()      {}
func (sc *SetComprehension) TokenLiteral() string { return sc.Token }
func (sc *SetComprehension) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	out.WriteString(sc.Element.String())
	out.WriteString(" for ")
	out.WriteString(sc.Variable.String())
	out.WriteString(" in ")
	out.WriteString(sc.Iterable.String())
	if sc.Filter != nil {
		out.WriteString(" if ")
		out.WriteString(sc.Filter.String())
	}
	out.WriteString("}")
	return out.String()
}

type DictComprehension struct {
	Token    string
	Key      Expression
	Value    Expression
	Variable *Identifier
	Iterable Expression
	Filter   Expression
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
	if dc.Filter != nil {
		out.WriteString(" if ")
		out.WriteString(dc.Filter.String())
	}
	out.WriteString("}")
	return out.String()
}

type YieldStatement struct {
	Token    string
	Expression Expression
}

func (ys *YieldStatement) statementNode()       {}
func (ys *YieldStatement) TokenLiteral() string { return ys.Token }
func (ys *YieldStatement) String() string {
	if ys.Expression != nil {
		return "yield " + ys.Expression.String()
	}
	return "yield"
}

type ClassStatement struct {
	Token       string
	Name        *Identifier
	SuperClass  *Identifier
	Body        *BlockStatement
	Methods     []*FunctionLiteral
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) expressionNode()      {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token }
func (cs *ClassStatement) String() string {
	var out bytes.Buffer
	out.WriteString("class ")
	out.WriteString(cs.Name.String())
	out.WriteString(" {\n")
	for _, method := range cs.Methods {
		out.WriteString(method.String())
		out.WriteString("\n")
	}
	out.WriteString(cs.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type ClassInstantiation struct {
	Token       string
	ClassName   *Identifier
	Arguments   []Expression
}

func (ci *ClassInstantiation) expressionNode()      {}
func (ci *ClassInstantiation) TokenLiteral() string { return ci.Token }
func (ci *ClassInstantiation) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ci.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ci.ClassName.String())
	out.WriteString("(" + strings.Join(args, ", ") + ")")
	return out.String()
}

type MemberAccess struct {
	Token       string
	Object      Expression
	Member      *Identifier
}

func (ma *MemberAccess) expressionNode()      {}
func (ma *MemberAccess) TokenLiteral() string { return ma.Token }
func (ma *MemberAccess) String() string {
	return ma.Object.String() + "." + ma.Member.String()
}

type MethodCall struct {
	Token       string
	Object      Expression
	Method      *Identifier
	Arguments   []Expression
}

func (mc *MethodCall) expressionNode()      {}
func (mc *MethodCall) TokenLiteral() string { return mc.Token }
func (mc *MethodCall) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range mc.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(mc.Object.String())
	out.WriteString(".")
	out.WriteString(mc.Method.String())
	out.WriteString("(" + strings.Join(args, ", ") + ")")
	return out.String()
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
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type AssignStatement struct {
	Token       string
	Left        Expression
	Value       Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token }
func (as *AssignStatement) String() string {
	leftStr := "nil"
	if as.Left != nil {
		leftStr = as.Left.String()
	}
	valueStr := "nil"
	if as.Value != nil {
		valueStr = as.Value.String()
	}
	return leftStr + " = " + valueStr
}

type AugAssignStatement struct {
	Token    string
	Left     Expression
	Operator string
	Value    Expression
}

func (as *AugAssignStatement) statementNode()       {}
func (as *AugAssignStatement) TokenLiteral() string { return as.Token }
func (as *AugAssignStatement) String() string {
	leftStr := "nil"
	if as.Left != nil {
		leftStr = as.Left.String()
	}
	valueStr := "nil"
	if as.Value != nil {
		valueStr = as.Value.String()
	}
	return leftStr + " " + as.Operator + "= " + valueStr
}

type ImportStatement struct {
	Token    string
	Module   *Identifier
	Alias    *Identifier
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token }
func (is *ImportStatement) String() string {
	if is.Alias != nil {
		return "import " + is.Module.String() + " as " + is.Alias.String()
	}
	return "import " + is.Module.String()
}

type FromImportStatement struct {
	Token      string
	Module     *Identifier
	Names      []*Identifier
	Alias      *Identifier
}

func (fis *FromImportStatement) statementNode()       {}
func (fis *FromImportStatement) TokenLiteral() string { return fis.Token }
func (fis *FromImportStatement) String() string {
	var out bytes.Buffer
	out.WriteString("from " + fis.Module.String() + " import ")
	names := []string{}
	for _, name := range fis.Names {
		names = append(names, name.String())
	}
	out.WriteString(strings.Join(names, ", "))
	if fis.Alias != nil {
		out.WriteString(" as " + fis.Alias.String())
	}
	return out.String()
}

type PassStatement struct {
	Token string
}

func (ps *PassStatement) statementNode()       {}
func (ps *PassStatement) TokenLiteral() string { return ps.Token }
func (ps *PassStatement) String() string       { return ps.Token }

type ReturnStatement struct {
	Token       string
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
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

type WhileStatement struct {
	Token    string
	Condition Expression
	Body     *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token }
func (ws *WhileStatement) String() string {
	return "while " + ws.Condition.String() + " {\n" + ws.Body.String() + "\n}"
}

type ForStatement struct {
	Token    string
	Value    *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token }
func (fs *ForStatement) String() string {
	return "for " + fs.Value.String() + " in " + fs.Iterable.String() + " {\n" + fs.Body.String() + "\n}"
}

type BreakStatement struct {
	Token string
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token }
func (bs *BreakStatement) String() string     { return "break" }

type ContinueStatement struct {
	Token string
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token }
func (cs *ContinueStatement) String() string     { return "continue" }

type TryStatement struct {
	Token    string
	Body     *BlockStatement
	Excepts  []*ExceptClause
	Finally  *BlockStatement
}

type ExceptClause struct {
	Token string
	Type  Expression
	Name  *Identifier
	Body  *BlockStatement
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token }
func (ts *TryStatement) String() string {
	var out bytes.Buffer
	out.WriteString("try {\n")
	out.WriteString(ts.Body.String())
	out.WriteString("\n}")
	for _, ex := range ts.Excepts {
		if ex.Type != nil {
			if ex.Name != nil {
				out.WriteString(" except " + ex.Type.String() + " as " + ex.Name.String() + " {\n")
			} else {
				out.WriteString(" except " + ex.Type.String() + " {\n")
			}
		} else {
			out.WriteString(" except {\n")
		}
		out.WriteString(ex.Body.String())
		out.WriteString("\n}")
	}
	if ts.Finally != nil {
		out.WriteString(" finally {\n")
		out.WriteString(ts.Finally.String())
		out.WriteString("\n}")
	}
	return out.String()
}

type RaiseStatement struct {
	Token       string
	Expression  Expression
}

func (rs *RaiseStatement) statementNode()       {}
func (rs *RaiseStatement) TokenLiteral() string { return rs.Token }
func (rs *RaiseStatement) String() string {
	if rs.Expression != nil {
		return "raise " + rs.Expression.String()
	}
	return "raise"
}

type WithStatement struct {
	Token string
	Expr  Expression
	Name  *Identifier
	Body  *BlockStatement
}

func (ws *WithStatement) statementNode()       {}
func (ws *WithStatement) TokenLiteral() string { return ws.Token }
func (ws *WithStatement) String() string {
	var out bytes.Buffer
	out.WriteString("with " + ws.Expr.String())
	if ws.Name != nil {
		out.WriteString(" as " + ws.Name.String())
	}
	out.WriteString(" {\n")
	out.WriteString(ws.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type Map map[string]Expression
