package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
)

func main() {
	code := `lst = []
lst.append(1)
lst.append(2)
print(len(lst))`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}
	
	program = desugar.Desugar(program)
	
	// 检查 AST
	inspectNode(program, 0)
	
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return
	}
	
	bytecode := comp.Bytecode()
	fmt.Println("\nConstants:", bytecode.Constants)
	fmt.Println("Instructions length:", len(bytecode.Instructions))
}

func inspectNode(node ast.Node, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}
	
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			inspectNode(stmt, indent)
		}
	case *ast.ExpressionStatement:
		fmt.Printf("%sExpressionStatement: %T\n", prefix, n.Expression)
		if call, ok := n.Expression.(*ast.CallExpression); ok {
			fmt.Printf("%s  CallExpression.Function: %T\n", prefix, call.Function)
			if member, ok := call.Function.(*ast.MemberAccess); ok {
				fmt.Printf("%s    MemberAccess.Object: %T, Member: %s\n", prefix, member.Object, member.Member.Value)
			}
		}
		inspectNode(n.Expression, indent+1)
	case *ast.CallExpression:
		fmt.Printf("%s  CallExpression.Function: %T\n", prefix, n.Function)
		if member, ok := n.Function.(*ast.MemberAccess); ok {
			fmt.Printf("%s    MemberAccess.Member: %s\n", prefix, member.Member.Value)
		}
		for _, arg := range n.Arguments {
			inspectNode(arg, indent+1)
		}
	case *ast.AssignStatement:
		fmt.Printf("%sAssignStatement: Left=%T, Value=%T\n", prefix, n.Left, n.Value)
		inspectNode(n.Left, indent+1)
		inspectNode(n.Value, indent+1)
	default:
		if expr, ok := node.(ast.Expression); ok {
			fmt.Printf("%s%s (Expression)\n", prefix, expr)
		} else {
			fmt.Printf("%s%s\n", prefix, node)
		}
	}
}
