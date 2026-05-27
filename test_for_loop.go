
package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/parser"
	"github.com/go-py/go-python/pkg/vm"
)

func main() {
	test := `def f(n):
    result = 1
    for i in range(1, n+1):
        result = result * i
    return result
print(f(10))`

	fmt.Println("=== Testing f(10) ===")

	l := lexer.New(test)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors())>0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return
	}
	fmt.Println("=== Parsed ===")
	for i, s := range program.Statements {
		fmt.Printf("stmt #%d: Type=%T, Value=%#v\n", i, s, s)
		if es, ok := s.(*ast.ExpressionStatement); ok {
			if fl, ok := es.Expression.(*ast.FunctionLiteral); ok {
				fmt.Printf("    FuncLit: Name=%v, Body Statements:\n", fl.Name)
				for j, bs := range fl.Body.Statements {
					fmt.Printf("      %d: Type=%T, Value=%#v\n", j, bs, bs)
					if fs, ok := bs.(*ast.ForStatement); ok {
						fmt.Printf("        ForStmt Iterable Type: %T, Value: %#v\n", fs.Iterable, fs.Iterable)
						if ce, ok := fs.Iterable.(*ast.CallExpression); ok {
							fmt.Printf("        CallExpr.Func Type: %T, Value: %#v\n", ce.Function, ce.Function)
							if id, ok := ce.Function.(*ast.Identifier); ok {
								fmt.Printf("        Identifier Value: %#v\n", id.Value)
							}
						}
					}
				}
			}
		}
	}

	program = desugar.Desugar(program)
	fmt.Println("\n=== After desugar ===")
	for _, s := range program.Statements {
		fmt.Println(s)
	}

	c := compiler.New()
	if err := c.Compile(program); err != nil {
		fmt.Printf("Compile err: %v\n", err)
		return
	}

	fmt.Println("\n=== Running ===")
	v := vm.New(c.Bytecode())
	if err := v.Run(); err != nil {
		fmt.Printf("Run err: %v\n", err)
	}
}
