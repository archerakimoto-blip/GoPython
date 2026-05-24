package compiler

import (
	"fmt"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/objects"
)

type Opcode byte

const (
	OpConstant Opcode = iota
	OpPop
	OpDupTop
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpLessThan
	OpMinus
	OpBang
	OpJump
	OpJumpNotTruthy
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpSet
	OpIndex
	OpSlice
	OpCall
	OpReturnValue
	OpReturn
	OpGetLocal
	OpSetLocal
	OpBeginTry
	OpEndTry
	OpRaise
	OpExceptHandler
	OpFinally
	OpYield
	OpEnterContext
	OpExitContext
	OpMakeGenerator
	OpYieldValue
)

type EmittedInstruction struct {
	Opcode   Opcode
	Position int
}

type Compiler struct {
	constants   []objects.Object
	symbolTable *SymbolTable

	instructions        Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Bytecode struct {
	Instructions Instructions
	Constants    []objects.Object
}

type Instructions []byte

func New() *Compiler {
	c := &Compiler{
		constants:   []objects.Object{},
		symbolTable: NewSymbolTable(),
	}
	c.registerBuiltins()
	return c
}

func (c *Compiler) registerBuiltins() {
	lenBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("len() takes exactly one argument")
			}
			switch arg := args[0].(type) {
			case *objects.List:
				return &objects.Integer{Value: int64(len(arg.Elements))}
			case *objects.String:
				return &objects.Integer{Value: int64(len(arg.Value))}
			case *objects.Dict:
				return &objects.Integer{Value: int64(len(arg.Pairs))}
			default:
				return objects.NewError("argument to 'len' not supported: %s", arg.Type())
			}
		},
	}
	lenIndex := len(c.constants)
	c.constants = append(c.constants, lenBuiltin)
	c.symbolTable.DefineBuiltin("len", lenIndex)

	appendBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewError("append() takes exactly 2 arguments")
			}
			list, ok := args[0].(*objects.List)
			if !ok {
				return objects.NewError("first argument to append() must be a list")
			}
			list.Elements = append(list.Elements, args[1])
			return objects.None_
		},
	}
	appendIndex := len(c.constants)
	c.constants = append(c.constants, appendBuiltin)
	c.symbolTable.DefineBuiltin("append", appendIndex)

	// setitem: set dict[key] = value
	setitemBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 3 {
				return objects.NewError("setitem() takes exactly 3 arguments: dict, key, value")
			}
			dict, ok := args[0].(*objects.Dict)
			if !ok {
				return objects.NewError("first argument to setitem() must be a dict")
			}
			dict.Pairs[args[1]] = args[2]
			return objects.None_
		},
	}
	setitemIndex := len(c.constants)
	c.constants = append(c.constants, setitemBuiltin)
	c.symbolTable.DefineBuiltin("setitem", setitemIndex)

	setaddBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewError("setadd() takes exactly 2 arguments")
			}
			set, ok := args[0].(*objects.Set)
			if !ok {
				return objects.NewError("first argument to setadd() must be a set")
			}
			for _, el := range set.Elements {
				if objects.Equal(el, args[1]) {
					return objects.None_
				}
			}
			set.Elements = append(set.Elements, args[1])
			return objects.None_
		},
	}
	setaddIndex := len(c.constants)
	c.constants = append(c.constants, setaddBuiltin)
	c.symbolTable.DefineBuiltin("setadd", setaddIndex)

	printBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return objects.None_
		},
	}
	printIndex := len(c.constants)
	c.constants = append(c.constants, printBuiltin)
	c.symbolTable.DefineBuiltin("print", printIndex)

	openBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("open() takes at least 1 argument")
			}

			filename := ""
			if str, ok := args[0].(*objects.String); ok {
				filename = str.Value
			} else {
				return objects.NewError("open(): first argument must be a string (filename)")
			}

			mode := "r"
			if len(args) > 1 {
				if str, ok := args[1].(*objects.String); ok {
					mode = str.Value
				}
			}

			return &objects.ContextManager{
				EnterFunc: func() objects.Object {
					return &objects.String{Value: "file_handle:" + filename + ":" + mode}
				},
				ExitFunc: func(exc objects.Object) objects.Object {
					return objects.None_
				},
			}
		},
	}
	openIndex := len(c.constants)
	c.constants = append(c.constants, openBuiltin)
	c.symbolTable.DefineBuiltin("open", openIndex)

	nextBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("next() takes exactly 1 argument")
			}
			gen, ok := args[0].(*objects.Generator)
			if !ok {
				return objects.NewError("next() argument must be a generator")
			}
			if gen.Done {
				return objects.None_
			}
			return gen
		},
	}
	nextIndex := len(c.constants)
	c.constants = append(c.constants, nextBuiltin)
	c.symbolTable.DefineBuiltin("next", nextIndex)

	typeBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("type() takes exactly 1 argument")
			}
			return &objects.String{Value: string(args[0].Type())}
		},
	}
	typeIndex := len(c.constants)
	c.constants = append(c.constants, typeBuiltin)
	c.symbolTable.DefineBuiltin("type", typeIndex)

	strBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("str() takes exactly 1 argument")
			}
			return &objects.String{Value: args[0].Inspect()}
		},
	}
	strIndex := len(c.constants)
	c.constants = append(c.constants, strBuiltin)
	c.symbolTable.DefineBuiltin("str", strIndex)

	intBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("int() takes exactly 1 argument")
			}
			switch arg := args[0].(type) {
			case *objects.Integer:
				return arg
			case *objects.Float:
				return &objects.Integer{Value: int64(arg.Value)}
			case *objects.String:
				var val int64
				_, err := fmt.Sscanf(arg.Value, "%d", &val)
				if err != nil {
					return objects.NewError("cannot convert string '%s' to int", arg.Value)
				}
				return &objects.Integer{Value: val}
			case *objects.Boolean:
				if arg.Value {
					return &objects.Integer{Value: 1}
				}
				return &objects.Integer{Value: 0}
			default:
				return objects.NewError("cannot convert %s to int", arg.Type())
			}
		},
	}
	intIndex := len(c.constants)
	c.constants = append(c.constants, intBuiltin)
	c.symbolTable.DefineBuiltin("int", intIndex)

	floatBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("float() takes exactly 1 argument")
			}
			switch arg := args[0].(type) {
			case *objects.Float:
				return arg
			case *objects.Integer:
				return &objects.Float{Value: float64(arg.Value)}
			case *objects.String:
				var val float64
				_, err := fmt.Sscanf(arg.Value, "%f", &val)
				if err != nil {
					return objects.NewError("cannot convert string '%s' to float", arg.Value)
				}
				return &objects.Float{Value: val}
			case *objects.Boolean:
				if arg.Value {
					return &objects.Float{Value: 1.0}
				}
				return &objects.Float{Value: 0.0}
			default:
				return objects.NewError("cannot convert %s to float", arg.Type())
			}
		},
	}
	floatIndex := len(c.constants)
	c.constants = append(c.constants, floatBuiltin)
	c.symbolTable.DefineBuiltin("float", floatIndex)

	boolBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("bool() takes exactly 1 argument")
			}
			switch arg := args[0].(type) {
			case *objects.Boolean:
				return arg
			case *objects.Integer:
				if arg.Value != 0 {
					return objects.True
				}
				return objects.False
			case *objects.Float:
				if arg.Value != 0 {
					return objects.True
				}
				return objects.False
			case *objects.String:
				if arg.Value != "" {
					return objects.True
				}
				return objects.False
			case *objects.List:
				if len(arg.Elements) > 0 {
					return objects.True
				}
				return objects.False
			case *objects.Dict:
				if len(arg.Pairs) > 0 {
					return objects.True
				}
				return objects.False
			case *objects.None:
				return objects.False
			default:
				return objects.True
			}
		},
	}
	boolIndex := len(c.constants)
	c.constants = append(c.constants, boolBuiltin)
	c.symbolTable.DefineBuiltin("bool", boolIndex)

	absBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("abs() takes exactly 1 argument")
			}
			switch arg := args[0].(type) {
			case *objects.Integer:
				if arg.Value < 0 {
					return &objects.Integer{Value: -arg.Value}
				}
				return arg
			case *objects.Float:
				if arg.Value < 0 {
					return &objects.Float{Value: -arg.Value}
				}
				return arg
			default:
				return objects.NewError("abs() argument must be a number")
			}
		},
	}
	absIndex := len(c.constants)
	c.constants = append(c.constants, absBuiltin)
	c.symbolTable.DefineBuiltin("abs", absIndex)

	rangeBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 || len(args) > 3 {
				return objects.NewError("range() takes 1 to 3 arguments")
			}
			var start, stop, step int64 = 0, 0, 1
			if len(args) == 1 {
				stopArg, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewError("range() argument must be an integer")
				}
				stop = stopArg.Value
			} else if len(args) == 2 {
				startArg, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewError("range() start argument must be an integer")
				}
				stopArg, ok := args[1].(*objects.Integer)
				if !ok {
					return objects.NewError("range() stop argument must be an integer")
				}
				start = startArg.Value
				stop = stopArg.Value
			} else {
				startArg, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewError("range() start argument must be an integer")
				}
				stopArg, ok := args[1].(*objects.Integer)
				if !ok {
					return objects.NewError("range() stop argument must be an integer")
				}
				stepArg, ok := args[2].(*objects.Integer)
				if !ok {
					return objects.NewError("range() step argument must be an integer")
				}
				start = startArg.Value
				stop = stopArg.Value
				step = stepArg.Value
				if step == 0 {
					return objects.NewError("range() step cannot be zero")
				}
			}
			elements := []objects.Object{}
			if step > 0 {
				for i := start; i < stop; i += step {
					elements = append(elements, &objects.Integer{Value: i})
				}
			} else {
				for i := start; i > stop; i += step {
					elements = append(elements, &objects.Integer{Value: i})
				}
			}
			return &objects.List{Elements: elements}
		},
	}
	rangeIndex := len(c.constants)
	c.constants = append(c.constants, rangeBuiltin)
	c.symbolTable.DefineBuiltin("range", rangeIndex)

	minBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("min() takes at least 1 argument")
			}
			var minVal *objects.Integer
			for _, arg := range args {
				intVal, ok := arg.(*objects.Integer)
				if !ok {
					return objects.NewError("min() arguments must be integers")
				}
				if minVal == nil || intVal.Value < minVal.Value {
					minVal = intVal
				}
			}
			return minVal
		},
	}
	minIndex := len(c.constants)
	c.constants = append(c.constants, minBuiltin)
	c.symbolTable.DefineBuiltin("min", minIndex)

	maxBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("max() takes at least 1 argument")
			}
			var maxVal *objects.Integer
			for _, arg := range args {
				intVal, ok := arg.(*objects.Integer)
				if !ok {
					return objects.NewError("max() arguments must be integers")
				}
				if maxVal == nil || intVal.Value > maxVal.Value {
					maxVal = intVal
				}
			}
			return maxVal
		},
	}
	maxIndex := len(c.constants)
	c.constants = append(c.constants, maxBuiltin)
	c.symbolTable.DefineBuiltin("max", maxIndex)

	sumBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("sum() takes exactly 1 argument")
			}
			list, ok := args[0].(*objects.List)
			if !ok {
				return objects.NewError("sum() argument must be a list")
			}
			var total int64 = 0
			for _, elem := range list.Elements {
				intVal, ok := elem.(*objects.Integer)
				if !ok {
					return objects.NewError("sum() list elements must be integers")
				}
				total += intVal.Value
			}
			return &objects.Integer{Value: total}
		},
	}
	sumIndex := len(c.constants)
	c.constants = append(c.constants, sumBuiltin)
	c.symbolTable.DefineBuiltin("sum", sumIndex)
}

func NewWithState(s *SymbolTable, constants []objects.Object) *Compiler {
	return &Compiler{
		constants:   constants,
		symbolTable: s,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(OpPop)

	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(OpAdd)
		case "-":
			c.emit(OpSub)
		case "*":
			c.emit(OpMul)
		case "/":
			c.emit(OpDiv)
		case ">":
			c.emit(OpGreaterThan)
		case "==":
			c.emit(OpEqual)
		case "!=":
			c.emit(OpNotEqual)
		case "and", "or":
			// 为了测试链式比较，我们暂时把 AND/OR 当作普通运算符传递给虚拟机
			// 我们会在虚拟机里处理它们，或者用其他方法
			// 先暂时不实现短路逻辑，直接让它们通过编译器
			// 让我们先不处理它们，看看我们的脱糖阶段能不能处理好
			return fmt.Errorf("unknown operator %s", node.Operator)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(OpBang)
		case "-":
			c.emit(OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &objects.Integer{Value: node.Value}
		c.emit(OpConstant, c.addConstant(integer))

	case *ast.FloatLiteral:
		float := &objects.Float{Value: node.Value}
		c.emit(OpConstant, c.addConstant(float))

	case *ast.StringLiteral:
		str := &objects.String{Value: node.Value}
		c.emit(OpConstant, c.addConstant(str))

	case *ast.Boolean:
		if node.Value {
			c.emit(OpTrue)
		} else {
			c.emit(OpFalse)
		}

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(OpPop) {
			c.removeLastPop()
		}

		jumpPos := c.emit(OpJump, 9999)
		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(OpSetGlobal, symbol.Index)
		} else {
			c.emit(OpSetLocal, symbol.Index)
		}

	case *ast.AssignStatement:
		// 编译右侧表达式
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		// 查找变量，如果不存在就自动定义
		symbol, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			symbol = c.symbolTable.Define(node.Name.Value)
		}

		if symbol.Scope == GlobalScope {
			c.emit(OpSetGlobal, symbol.Index)
		} else {
			c.emit(OpSetLocal, symbol.Index)
		}

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		if symbol.Scope == BuiltinScope {
			c.emit(OpConstant, symbol.Index)
		} else if symbol.Scope == GlobalScope {
			c.emit(OpGetGlobal, symbol.Index)
		} else {
			c.emit(OpGetLocal, symbol.Index)
		}

	case *ast.ListLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(OpArray, len(node.Elements))

	case *ast.SetLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(OpSet, len(node.Elements))

	case *ast.ListComprehension:
		return c.compileListComprehension(node)

	case *ast.SetComprehension:
		return c.compileSetComprehension(node)

	case *ast.DictComprehension:
		return c.compileDictComprehension(node)

	case *ast.DictLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(OpHash, len(node.Pairs)*2)

	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(OpIndex)
	case *ast.SliceExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		if node.Start == nil {
			c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 0}))
		} else {
			err = c.Compile(node.Start)
			if err != nil {
				return err
			}
		}
		if node.End == nil {
			c.emit(OpConstant, c.addConstant(&objects.Integer{Value: -1}))
		} else {
			err = c.Compile(node.End)
			if err != nil {
				return err
			}
		}
		c.emit(OpSlice)

	case *ast.FunctionLiteral:
		// 保存当前指令列表
		outerInstructions := c.instructions
		// 创建新的指令列表用于函数体
		c.instructions = make(Instructions, 0)

		c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(OpReturnValue) {
			c.emit(OpReturn)
		}

		// 获取函数体的指令
		fnInstructions := c.instructions

		// 恢复外部符号表和指令列表
		c.symbolTable = c.symbolTable.Outer
		c.instructions = outerInstructions

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions

		// 处理自由符号（需要从外部作用域获取）
		for _, s := range freeSymbols {
			if s.Scope == GlobalScope {
				c.emit(OpGetGlobal, s.Index)
			} else {
				c.emit(OpGetLocal, s.Index)
			}
		}

		compiledFn := &CompiledFunction{
		Instructions:  fnInstructions,
		NumLocals:     numLocals,
		NumParameters: len(node.Parameters),
		IsGenerator:   c.hasYieldInBody(node.Body),
	}

	c.emit(OpConstant, c.addConstant(compiledFn))
	if compiledFn.IsGenerator {
		c.emit(OpMakeGenerator)
	}

	case *ast.LambdaExpression:
		funcLit := &ast.FunctionLiteral{
			Token:      node.Token,
			Parameters: node.Parameters,
			Body: &ast.BlockStatement{
				Token: node.Token,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      node.Token,
						Expression: node.Body,
					},
				},
			},
		}
		return c.Compile(funcLit)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(OpCall, len(node.Arguments))

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(OpReturnValue)

	case *ast.WhileStatement:
		conditionPos := len(c.instructions)
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)

		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(OpPop) {
			c.removeLastPop()
		}

		c.emit(OpJump, conditionPos)

		afterLoopPos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterLoopPos)

	case *ast.ForStatement:
		return fmt.Errorf("for loops should be desugared before compilation")
	case *ast.BreakStatement:
		return fmt.Errorf("break statements should be desugared before compilation")
	case *ast.ContinueStatement:
		return fmt.Errorf("continue statements should be desugared before compilation")
	case *ast.TryStatement:
		return c.compileTryStatement(node)
	case *ast.RaiseStatement:
		if node.Expression != nil {
			if err := c.Compile(node.Expression); err != nil {
				return err
			}
		} else {
			c.emit(OpNull)
		}
		c.emit(OpRaise)
		return nil
	case *ast.WithStatement:
		if err := c.Compile(node.Expr); err != nil {
			return err
		}

		c.emit(OpEnterContext)

		if node.Name != nil {
			symbol := c.symbolTable.Define(node.Name.Value)
			if symbol.Scope == GlobalScope {
				c.emit(OpSetGlobal, symbol.Index)
			} else {
				c.emit(OpSetLocal, symbol.Index)
			}
		} else {
			c.emit(OpPop)
		}

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		c.emit(OpExitContext)
		return nil
	case *ast.YieldStatement:
		if node.Expression != nil {
			if err := c.Compile(node.Expression); err != nil {
				return err
			}
		} else {
			c.emit(OpNull)
		}
		c.emit(OpYieldValue)
		return nil
	}

	return nil
}

func (c *Compiler) hasYieldInBody(node ast.Node) bool {
	switch node := node.(type) {
	case *ast.YieldStatement:
		return true
	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			if c.hasYieldInBody(stmt) {
				return true
			}
		}
	case *ast.IfExpression:
		if c.hasYieldInBody(node.Consequence) {
			return true
		}
		if node.Alternative != nil && c.hasYieldInBody(node.Alternative) {
			return true
		}
	case *ast.WhileStatement:
		if c.hasYieldInBody(node.Body) {
			return true
		}
	case *ast.FunctionLiteral:
		return false
	}
	return false
}

func (c *Compiler) compileTryStatement(ts *ast.TryStatement) error {
	hasExcept := len(ts.Excepts) > 0
	hasFinally := ts.Finally != nil

	c.emit(OpBeginTry, boolToInt(hasExcept), boolToInt(hasFinally))

	if err := c.Compile(ts.Body); err != nil {
		return err
	}

	jumpPositions := []int{}
	if hasFinally {
		jumpPositions = append(jumpPositions, c.emit(OpJump, 0))
	}

	if hasExcept {
		for _, ex := range ts.Excepts {
			var typeIdx int
			if ex.Type != nil {
				if typeStr, ok := ex.Type.(*ast.Identifier); ok {
					typeIdx = c.addConstant(&objects.String{Value: typeStr.Value})
				} else {
					typeIdx = c.addConstant(&objects.String{Value: "Exception"})
				}
			} else {
				typeIdx = c.addConstant(&objects.String{Value: ""})
			}

			var varIdx int
			var varSymbol Symbol
			if ex.Name != nil {
				varIdx = c.addConstant(&objects.String{Value: ex.Name.Value})
				varSymbol = c.symbolTable.Define(ex.Name.Value)
			} else {
				varIdx = c.addConstant(&objects.String{Value: ""})
			}

			c.emit(OpExceptHandler, typeIdx, varIdx)

			if ex.Name != nil {
				c.emit(OpDupTop)
				if varSymbol.Scope == GlobalScope {
					c.emit(OpSetGlobal, varSymbol.Index)
				} else {
					c.emit(OpSetLocal, varSymbol.Index)
				}
			}

			if err := c.Compile(ex.Body); err != nil {
				return err
			}

			if hasFinally {
				jumpPositions = append(jumpPositions, c.emit(OpJump, 0))
			}
		}
	}

	if hasFinally {
		finallyStartPos := len(c.instructions)
		c.emit(OpFinally, 0)
		if err := c.Compile(ts.Finally); err != nil {
			return err
		}
		for _, pos := range jumpPositions {
			c.changeOperand(pos, finallyStartPos)
		}
	}

	c.emit(OpEndTry)

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj objects.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op Opcode, operands ...int) int {
	ins := Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op Opcode, pos int) {
	previous := c.lastInstruction
	c.previousInstruction = previous

	last := EmittedInstruction{Opcode: op, Position: pos}
	c.lastInstruction = last
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) lastInstructionIs(op Opcode) bool {
	if len(c.instructions) == 0 {
		return false
	}
	return c.lastInstruction.Opcode == op
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.lastInstruction.Position
	c.replaceInstruction(lastPos, Make(OpReturnValue))
	c.lastInstruction.Opcode = OpReturnValue
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := Opcode(c.instructions[opPos])
	newInstruction := Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			instruction[offset] = byte(o & 0xFF)
			instruction[offset+1] = byte((o >> 8) & 0xFF)
		case 1:
			instruction[offset] = byte(o & 0xFF)
		}
		offset += width
	}

	return instruction
}

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpPop:           {"OpPop", []int{}},
	OpAdd:           {"OpAdd", []int{}},
	OpSub:           {"OpSub", []int{}},
	OpMul:           {"OpMul", []int{}},
	OpDiv:           {"OpDiv", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpLessThan:      {"OpLessThan", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpNull:          {"OpNull", []int{}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpSet:           {"OpSet", []int{2}},
	OpIndex:         {"OpIndex", []int{}},
	OpSlice:         {"OpSlice", []int{}},
	OpCall:          {"OpCall", []int{1}},
	OpReturnValue:   {"OpReturnValue", []int{}},
	OpReturn:        {"OpReturn", []int{}},
	OpGetLocal:      {"OpGetLocal", []int{1}},
	OpSetLocal:      {"OpSetLocal", []int{1}},
	OpBeginTry:      {"OpBeginTry", []int{2, 2}},
	OpEndTry:        {"OpEndTry", []int{}},
	OpRaise:         {"OpRaise", []int{}},
	OpDupTop:        {"OpDupTop", []int{}},
	OpExceptHandler: {"OpExceptHandler", []int{2, 2}},
	OpFinally:       {"OpFinally", []int{2}},
	OpYield:         {"OpYield", []int{}},
	OpEnterContext:  {"OpEnterContext", []int{}},
	OpExitContext:   {"OpExitContext", []int{}},
	OpMakeGenerator: {"OpMakeGenerator", []int{}},
	OpYieldValue:    {"OpYieldValue", []int{}},
}

type CompiledFunction struct {
	Instructions  Instructions
	NumLocals     int
	NumParameters int
	IsGenerator   bool
}

func (cf *CompiledFunction) Type() objects.ObjectType { return "COMPILED_FUNCTION" }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	FreeScope    SymbolScope = "FREE"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
	FreeSymbols    []Symbol
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, Scope: LocalScope}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(name string, index int) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := s.store[name]
	if !ok && s.Outer != nil {
		symbol, ok = s.Outer.Resolve(name)
		if !ok {
			return symbol, ok
		}

		if symbol.Scope == GlobalScope {
			return symbol, ok
		}

		free := s.defineFree(symbol)
		return free, true
	}
	return symbol, ok
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)

	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1, Scope: FreeScope}
	s.store[original.Name] = symbol
	return symbol
}

func (c *Compiler) compileListComprehension(node *ast.ListComprehension) error {
	// 使用全局临时变量
	resultSym := c.symbolTable.Define("_lc_result")
	iterSym := c.symbolTable.Define("_lc_i")
	elemSym := c.symbolTable.Define(node.Variable.Value)

	// 1. 创建空结果数组
	c.emit(OpArray, 0)
	c.emit(OpSetGlobal, resultSym.Index)

	// 2. 初始化迭代索引为0
	c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 0}))
	c.emit(OpSetGlobal, iterSym.Index)

	// 3. 循环开始
	loopStart := len(c.instructions)

	// 5. 先 push i
	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	// 4. 编译可迭代对象并获取其长度: 先 push len builtin，再 push iterable
	c.emit(OpConstant, 0) // len builtin
	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	c.emit(OpCall, 1)
	// 现在栈上是 [i, len_val]，比较 i < len_val
	c.emit(OpLessThan)

	// 6. 如果条件不满足则跳出
	jumpEndPos := c.emit(OpJumpNotTruthy, 9999)

	// 7. elemVar = iterable[i]: compile node.Iterable, push i, index, assign
	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	c.emit(OpIndex)
	if elemSym.Scope == GlobalScope {
		c.emit(OpSetGlobal, elemSym.Index)
	} else {
		c.emit(OpSetLocal, elemSym.Index)
	}

	// 8. 如果有条件，检查一下
	if node.Condition != nil {
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		jumpCondPos := c.emit(OpJumpNotTruthy, 9999)

		// 9. append 元素到结果: push append builtin, push list, push elem, call, pop
		c.emit(OpConstant, 1)
		if resultSym.Scope == GlobalScope {
			c.emit(OpGetGlobal, resultSym.Index)
		} else {
			c.emit(OpGetLocal, resultSym.Index)
		}
		if err := c.Compile(node.Element); err != nil {
			return err
		}
		c.emit(OpCall, 2)
		c.emit(OpPop)

		jumpAfterCondPos := len(c.instructions)
		c.changeOperand(jumpCondPos, jumpAfterCondPos)
	} else {
		// 9. append 元素到结果
		c.emit(OpConstant, 1)
		if resultSym.Scope == GlobalScope {
			c.emit(OpGetGlobal, resultSym.Index)
		} else {
			c.emit(OpGetLocal, resultSym.Index)
		}
		if err := c.Compile(node.Element); err != nil {
			return err
		}
		c.emit(OpCall, 2)
		c.emit(OpPop)
	}

	// 10. i +=1
	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 1}))
	c.emit(OpAdd)
	if iterSym.Scope == GlobalScope {
		c.emit(OpSetGlobal, iterSym.Index)
	} else {
		c.emit(OpSetLocal, iterSym.Index)
	}

	// 11. 跳转回循环开始
	c.emit(OpJump, loopStart)

	// 12. 设置跳出循环的跳转位置
	loopEnd := len(c.instructions)
	c.changeOperand(jumpEndPos, loopEnd)

	// 13. 返回结果数组
	if resultSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, resultSym.Index)
	} else {
		c.emit(OpGetLocal, resultSym.Index)
	}

	return nil
}

func (c *Compiler) compileDictComprehension(node *ast.DictComprehension) error {
	// 使用全局临时变量
	resultSym := c.symbolTable.Define("_dc_result")
	iterSym := c.symbolTable.Define("_dc_i")
	elemSym := c.symbolTable.Define(node.Variable.Value)

	// 1. 创建空结果字典
	c.emit(OpHash, 0)
	c.emit(OpSetGlobal, resultSym.Index)

	// 2. 初始化迭代索引为0
	c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 0}))
	c.emit(OpSetGlobal, iterSym.Index)

	// 3. 循环开始
	loopStart := len(c.instructions)

	// 5. 先 push i
	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	// 4. 编译可迭代对象并获取其长度: 先 push len builtin，再 push iterable
	c.emit(OpConstant, 0) // len builtin
	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	c.emit(OpCall, 1)
	// 现在栈上是 [i, len_val]，比较 i < len_val
	c.emit(OpLessThan)

	// 6. 如果条件不满足则跳出
	jumpEndPos := c.emit(OpJumpNotTruthy, 9999)

	// 7. elemVar = iterable[i]: compile node.Iterable, push i, index, assign
	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	c.emit(OpIndex)
	if elemSym.Scope == GlobalScope {
		c.emit(OpSetGlobal, elemSym.Index)
	} else {
		c.emit(OpSetLocal, elemSym.Index)
	}

	// 8. 如果有条件，检查一下
	if node.Condition != nil {
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		jumpCondPos := c.emit(OpJumpNotTruthy, 9999)

		// 9. setitem(dict, key, value)
		c.emit(OpConstant, 2) // setitem builtin index is 2 (len is 0, append is 1)
		if resultSym.Scope == GlobalScope {
			c.emit(OpGetGlobal, resultSym.Index)
		} else {
			c.emit(OpGetLocal, resultSym.Index)
		}
		if err := c.Compile(node.Key); err != nil {
			return err
		}
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		c.emit(OpCall, 3)
		c.emit(OpPop)

		jumpAfterCondPos := len(c.instructions)
		c.changeOperand(jumpCondPos, jumpAfterCondPos)
	} else {
		// 9. setitem(dict, key, value)
		c.emit(OpConstant, 2) // setitem builtin index is 2
		if resultSym.Scope == GlobalScope {
			c.emit(OpGetGlobal, resultSym.Index)
		} else {
			c.emit(OpGetLocal, resultSym.Index)
		}
		if err := c.Compile(node.Key); err != nil {
			return err
		}
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		c.emit(OpCall, 3)
		c.emit(OpPop)
	}

	// 10. i +=1
	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 1}))
	c.emit(OpAdd)
	if iterSym.Scope == GlobalScope {
		c.emit(OpSetGlobal, iterSym.Index)
	} else {
		c.emit(OpSetLocal, iterSym.Index)
	}

	// 11. 跳转回循环开始
	c.emit(OpJump, loopStart)

	// 12. 设置跳出循环的跳转位置
	loopEnd := len(c.instructions)
	c.changeOperand(jumpEndPos, loopEnd)

	// 13. 返回结果字典
	if resultSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, resultSym.Index)
	} else {
		c.emit(OpGetLocal, resultSym.Index)
	}

	return nil
}

func (c *Compiler) compileSetComprehension(node *ast.SetComprehension) error {
	resultSym := c.symbolTable.Define("_sc_result")
	iterSym := c.symbolTable.Define("_sc_i")
	elemSym := c.symbolTable.Define(node.Variable.Value)

	c.emit(OpSet, 0)
	c.emit(OpSetGlobal, resultSym.Index)

	c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 0}))
	c.emit(OpSetGlobal, iterSym.Index)

	loopStart := len(c.instructions)

	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	c.emit(OpConstant, 0)
	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	c.emit(OpCall, 1)
	c.emit(OpLessThan)

	jumpEndPos := c.emit(OpJumpNotTruthy, 9999)

	if err := c.Compile(node.Iterable); err != nil {
		return err
	}
	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	c.emit(OpIndex)
	if elemSym.Scope == GlobalScope {
		c.emit(OpSetGlobal, elemSym.Index)
	} else {
		c.emit(OpSetLocal, elemSym.Index)
	}

	if node.Condition != nil {
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		jumpCondPos := c.emit(OpJumpNotTruthy, 9999)

		c.emit(OpConstant, 3)
		if resultSym.Scope == GlobalScope {
			c.emit(OpGetGlobal, resultSym.Index)
		} else {
			c.emit(OpGetLocal, resultSym.Index)
		}
		if err := c.Compile(node.Element); err != nil {
			return err
		}
		c.emit(OpCall, 2)
		c.emit(OpPop)

		jumpAfterCondPos := len(c.instructions)
		c.changeOperand(jumpCondPos, jumpAfterCondPos)
	} else {
		c.emit(OpConstant, 3)
		if resultSym.Scope == GlobalScope {
			c.emit(OpGetGlobal, resultSym.Index)
		} else {
			c.emit(OpGetLocal, resultSym.Index)
		}
		if err := c.Compile(node.Element); err != nil {
			return err
		}
		c.emit(OpCall, 2)
		c.emit(OpPop)
	}

	if iterSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, iterSym.Index)
	} else {
		c.emit(OpGetLocal, iterSym.Index)
	}
	c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 1}))
	c.emit(OpAdd)
	if iterSym.Scope == GlobalScope {
		c.emit(OpSetGlobal, iterSym.Index)
	} else {
		c.emit(OpSetLocal, iterSym.Index)
	}

	c.emit(OpJump, loopStart)

	loopEnd := len(c.instructions)
	c.changeOperand(jumpEndPos, loopEnd)

	if resultSym.Scope == GlobalScope {
		c.emit(OpGetGlobal, resultSym.Index)
	} else {
		c.emit(OpGetLocal, resultSym.Index)
	}

	return nil
}



