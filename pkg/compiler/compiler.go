package compiler

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/gc"
	"github.com/go-py/go-python/pkg/interop"
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
	OpGetFree
	OpClosure
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
	OpCreateClass
	OpCreateClassWithSuper
	OpGetAttribute
	OpSetAttribute
	OpFormatString
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

func (c *Compiler) emit(op Opcode, operands ...int) int {
	ins := c.make(op, operands...)
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	c.previousInstruction = c.lastInstruction
	c.lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
	return pos
}

func (c *Compiler) emit1(op Opcode, operand int) int {
	ins := []byte{byte(op), byte(operand & 0xFF)}
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	c.previousInstruction = c.lastInstruction
	c.lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
	return pos
}

func (c *Compiler) emitClosure(constIndex int, numFree int) int {
	ins := []byte{byte(OpClosure), byte(constIndex >> 8), byte(constIndex & 0xFF), byte(numFree)}
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	c.previousInstruction = c.lastInstruction
	c.lastInstruction = EmittedInstruction{Opcode: OpClosure, Position: pos}
	return pos
}

func (c *Compiler) make(op Opcode, operands ...int) []byte {
	ins := []byte{byte(op)}
	for _, o := range operands {
		ins = append(ins, c.makeOperand(o)...)
	}
	return ins
}

func (c *Compiler) makeOperand(op int) []byte {
	lo := byte(op & 0xFF)
	hi := byte(op >> 8)
	return []byte{hi, lo}
}

func (c *Compiler) makeOperand1(op int) []byte {
	return []byte{byte(op & 0xFF)}
}

func (c *Compiler) lastInstructionIs(op Opcode) bool {
	if len(c.instructions) == 0 {
		return false
	}
	return c.lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.lastInstruction
	prev := c.previousInstruction

	c.instructions = c.instructions[:last.Position]
	c.lastInstruction = prev
	
	if len(c.instructions) > 0 {
		prevPrev := EmittedInstruction{}
		for i := len(c.instructions) - 1; i >= 0; i-- {
			if c.instructions[i] != byte(OpPop) {
				break
			}
			prevPrev.Opcode = OpPop
			prevPrev.Position = i
		}
		if prevPrev.Position > 0 {
			c.previousInstruction = prevPrev
		}
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.lastInstruction.Position
	c.instructions[lastPos] = byte(OpReturnValue)
	c.lastInstruction.Opcode = OpReturnValue
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	oldInstruction := c.instructions[opPos]
	c.instructions[opPos] = byte(oldInstruction)
	c.instructions[opPos+1] = byte(operand >> 8)
	c.instructions[opPos+2] = byte(operand & 0xFF)
}

func (c *Compiler) adjustLocalIndices(instructions []byte, numFree int) []byte {
	if numFree == 0 {
		return append([]byte{}, instructions...)
	}

	result := make([]byte, len(instructions))
	copy(result, instructions)

	for i := 0; i < len(result); i++ {
		op := Opcode(result[i])
		switch op {
		case OpGetLocal, OpSetLocal:
			// Next byte is the local index
			if i+1 < len(result) {
				oldIndex := int(result[i+1])
				newIndex := oldIndex + numFree
				result[i+1] = byte(newIndex)
				i++ // Skip the index byte
			}
		}
	}

	return result
}

func (c *Compiler) enterScope() {
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) exitScope() {
	c.symbolTable = c.symbolTable.outer
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) SymbolTable() *SymbolTable {
	return c.symbolTable
}

func New() *Compiler {
	c := &Compiler{
		constants:   []objects.Object{},
		symbolTable: NewSymbolTable(),
	}
	c.registerBuiltins()
	return c
}

func (c *Compiler) registerBuiltins() {
	mathModule := objects.CreateMathModule()
	objects.RegisterModule("math", mathModule)
	mathIndex := len(c.constants)
	c.constants = append(c.constants, mathModule)
	c.symbolTable.DefineBuiltin("math", mathIndex)

	// 注册 sys 模块
	sysModule := objects.CreateSysModule()
	objects.RegisterModule("sys", sysModule)
	sysIndex := len(c.constants)
	c.constants = append(c.constants, sysModule)
	c.symbolTable.DefineBuiltin("sys", sysIndex)

	// 注册 os 模块
	osModule := objects.CreateOsModule()
	objects.RegisterModule("os", osModule)
	osIndex := len(c.constants)
	c.constants = append(c.constants, osModule)
	c.symbolTable.DefineBuiltin("os", osIndex)

	// 注册 json 模块
	jsonModule := objects.CreateJsonModule()
	objects.RegisterModule("json", jsonModule)
	jsonIndex := len(c.constants)
	c.constants = append(c.constants, jsonModule)
	c.symbolTable.DefineBuiltin("json", jsonIndex)

	// 注册 gc 模块
	gcModule := gc.CreateGCModule()
	objects.RegisterModule("gc", gcModule)
	gcIndex := len(c.constants)
	c.constants = append(c.constants, gcModule)
	c.symbolTable.DefineBuiltin("gc", gcIndex)

	// 注册 random 模块
	randomModule := objects.CreateRandomModule()
	objects.RegisterModule("random", randomModule)
	randomIndex := len(c.constants)
	c.constants = append(c.constants, randomModule)
	c.symbolTable.DefineBuiltin("random", randomIndex)

	// 注册 string 模块
	stringModule := objects.CreateStringModule()
	objects.RegisterModule("string", stringModule)
	stringIndex := len(c.constants)
	c.constants = append(c.constants, stringModule)
	c.symbolTable.DefineBuiltin("string", stringIndex)

	// 注册 time 模块
	timeModule := objects.CreateTimeModule()
	objects.RegisterModule("time", timeModule)
	timeIndex := len(c.constants)
	c.constants = append(c.constants, timeModule)
	c.symbolTable.DefineBuiltin("time", timeIndex)

	// 注册 datetime 模块
	datetimeModule := objects.CreateDatetimeModule()
	objects.RegisterModule("datetime", datetimeModule)
	datetimeIndex := len(c.constants)
	c.constants = append(c.constants, datetimeModule)
	c.symbolTable.DefineBuiltin("datetime", datetimeIndex)

	// 注册 cpython 互操作模块
	cpythonModule := interop.CreateCPythonModule()
	objects.RegisterModule("cpython", cpythonModule)
	cpythonIndex := len(c.constants)
	c.constants = append(c.constants, cpythonModule)
	c.symbolTable.DefineBuiltin("cpython", cpythonIndex)

	lenBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("len() takes exactly one argument")
			}
			switch arg := args[0].(type) {
			case *objects.List:
				return &objects.Integer{Value: int64(len(arg.Elements))}
			case *objects.Tuple:
				return &objects.Integer{Value: int64(len(arg.Elements))}
			case *objects.String:
				return &objects.Integer{Value: int64(len(arg.Value))}
			case *objects.Dict:
				return &objects.Integer{Value: int64(len(arg.Pairs))}
			case *objects.Set:
				return &objects.Integer{Value: int64(len(arg.Elements))}
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
			dict.Set(args[1], args[2])
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
			set.Add(args[1])
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
				if arg != nil {
					fmt.Print(arg.Inspect())
				}
			}
			fmt.Println()
			os.Stdout.Sync()
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
				return objects.NewError("StopIteration")
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
				return objects.NewTypeError("int() takes exactly 1 argument")
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
					return objects.NewValueError("cannot convert string '%s' to int", arg.Value)
				}
				return &objects.Integer{Value: val}
			case *objects.Boolean:
				if arg.Value {
					return &objects.Integer{Value: 1}
				}
				return &objects.Integer{Value: 0}
			default:
				return objects.NewTypeError("cannot convert %s to int", arg.Type())
			}
		},
	}
	intIndex := len(c.constants)
	c.constants = append(c.constants, intBuiltin)
	c.symbolTable.DefineBuiltin("int", intIndex)

	floatBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("float() takes exactly 1 argument")
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
					return objects.NewValueError("cannot convert string '%s' to float", arg.Value)
				}
				return &objects.Float{Value: val}
			case *objects.Boolean:
				if arg.Value {
					return &objects.Float{Value: 1.0}
				}
				return &objects.Float{Value: 0.0}
			default:
				return objects.NewTypeError("cannot convert %s to float", arg.Type())
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
				return objects.NewTypeError("range() takes 1 to 3 arguments")
			}
			var start, stop, step int64 = 0, 0, 1
			if len(args) == 1 {
				stopArg, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("range() argument must be an integer")
				}
				stop = stopArg.Value
			} else if len(args) == 2 {
				startArg, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("range() start argument must be an integer")
				}
				stopArg, ok := args[1].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("range() stop argument must be an integer")
				}
				start = startArg.Value
				stop = stopArg.Value
			} else {
				startArg, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("range() start argument must be an integer")
				}
				stopArg, ok := args[1].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("range() stop argument must be an integer")
				}
				stepArg, ok := args[2].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("range() step argument must be an integer")
				}
				start = startArg.Value
				stop = stopArg.Value
				step = stepArg.Value
				if step == 0 {
					return objects.NewValueError("range() step cannot be zero")
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
			var minInt *objects.Integer
			var minFloat *objects.Float
			hasFloat := false

			for _, arg := range args {
				switch val := arg.(type) {
				case *objects.Integer:
					if minInt == nil {
						minInt = val
					} else if val.Value < minInt.Value {
						minInt = val
					}
				case *objects.Float:
					hasFloat = true
					if minFloat == nil {
						minFloat = val
					} else if val.Value < minFloat.Value {
						minFloat = val
					}
				default:
					return objects.NewError("min() arguments must be numbers")
				}
			}

			if hasFloat {
				if minInt != nil {
					intAsFloat := float64(minInt.Value)
					if minFloat == nil || intAsFloat < minFloat.Value {
						return &objects.Float{Value: intAsFloat}
					}
				}
				return minFloat
			}
			return minInt
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
			var maxInt *objects.Integer
			var maxFloat *objects.Float
			hasFloat := false

			for _, arg := range args {
				switch val := arg.(type) {
				case *objects.Integer:
					if maxInt == nil {
						maxInt = val
					} else if val.Value > maxInt.Value {
						maxInt = val
					}
				case *objects.Float:
					hasFloat = true
					if maxFloat == nil {
						maxFloat = val
					} else if val.Value > maxFloat.Value {
						maxFloat = val
					}
				default:
					return objects.NewError("max() arguments must be numbers")
				}
			}

			if hasFloat {
				if maxInt != nil {
					intAsFloat := float64(maxInt.Value)
					if maxFloat == nil || intAsFloat > maxFloat.Value {
						return &objects.Float{Value: intAsFloat}
					}
				}
				return maxFloat
			}
			return maxInt
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
			var totalInt int64 = 0
			var totalFloat float64 = 0.0
			hasFloat := false

			for _, elem := range list.Elements {
				switch val := elem.(type) {
				case *objects.Integer:
					if hasFloat {
						totalFloat += float64(val.Value)
					} else {
						totalInt += val.Value
					}
				case *objects.Float:
					if !hasFloat {
						hasFloat = true
						totalFloat = float64(totalInt)
					}
					totalFloat += val.Value
				default:
					return objects.NewError("sum() list elements must be numbers")
				}
			}

			if hasFloat {
				return &objects.Float{Value: totalFloat}
			}
			return &objects.Integer{Value: totalInt}
		},
	}
	sumIndex := len(c.constants)
	c.constants = append(c.constants, sumBuiltin)
	c.symbolTable.DefineBuiltin("sum", sumIndex)

	formatBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewError("format() takes at least 1 argument")
			}
			template, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewError("format() first argument must be a string")
			}
			result := objects.FormatString(template.Value, args[1:]...)
			return &objects.String{Value: result}
		},
	}
	formatIndex := len(c.constants)
	c.constants = append(c.constants, formatBuiltin)
	c.symbolTable.DefineBuiltin("format", formatIndex)

	inputBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.String{Value: ""}
		},
	}
	inputIndex := len(c.constants)
	c.constants = append(c.constants, inputBuiltin)
	c.symbolTable.DefineBuiltin("input", inputIndex)

	roundBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 || len(args) > 2 {
				return objects.NewError("round() takes 1 or 2 arguments")
			}
			arg := args[0]
			var ndigits int64 = 0
			if len(args) == 2 {
				nd, ok := args[1].(*objects.Integer)
				if !ok {
					return objects.NewError("round() second argument must be an integer")
				}
				ndigits = nd.Value
			}
			
			switch v := arg.(type) {
			case *objects.Float:
				multiplier := 1.0
				for i := int64(0); i < ndigits; i++ {
					multiplier *= 10.0
				}
				if ndigits > 0 {
					rounded := float64(int64(v.Value*multiplier+0.5)) / multiplier
					return &objects.Float{Value: rounded}
				}
				return &objects.Integer{Value: int64(v.Value + 0.5)}
			case *objects.Integer:
				return arg
			default:
				return objects.NewError("round() argument must be a number")
			}
		},
	}
	roundIndex := len(c.constants)
	c.constants = append(c.constants, roundBuiltin)
	c.symbolTable.DefineBuiltin("round", roundIndex)

	zipBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) == 0 {
				return &objects.List{Elements: []objects.Object{}}
			}

			// 检查所有参数是否都是列表
			lists := make([]*objects.List, len(args))
			minLen := -1
			for i, arg := range args {
				list, ok := arg.(*objects.List)
				if !ok {
					return objects.NewError("zip() arguments must be lists")
				}
				lists[i] = list
				if minLen == -1 || len(list.Elements) < minLen {
					minLen = len(list.Elements)
				}
			}

			// 构建结果
			result := &objects.List{Elements: []objects.Object{}}
			for i := 0; i < minLen; i++ {
				tuple := &objects.List{Elements: []objects.Object{}}
				for _, list := range lists {
					tuple.Elements = append(tuple.Elements, list.Elements[i])
				}
				result.Elements = append(result.Elements, tuple)
			}

			return result
		},
	}
	zipIndex := len(c.constants)
	c.constants = append(c.constants, zipBuiltin)
	c.symbolTable.DefineBuiltin("zip", zipIndex)
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
		if fl, ok := node.Expression.(*ast.FunctionLiteral); ok {
			if fl.Name != "" {
				symbol := c.symbolTable.DefineFunctionName(fl.Name)
				c.emit(OpSetGlobal, symbol.Index)
			} else {
				c.emit(OpPop)
			}
		} else {
			c.emit(OpPop)
		}

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

	case *ast.FStringLiteral:
		// 编译 f-string 的所有部分
		partsCount := 0
		for _, part := range node.Parts {
			err := c.Compile(part)
			if err != nil {
				return err
			}
			partsCount++
		}
		// 执行格式化操作
		c.emit(OpFormatString, partsCount)

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
			c.emit1(OpSetLocal, symbol.Index)
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
		} else if symbol.Scope == GlobalScope || symbol.Scope == FunctionScope {
			c.emit(OpGetGlobal, symbol.Index)
		} else if symbol.Scope == FreeScope {
			c.emit1(OpGetFree, symbol.Index)
		} else {
			c.emit1(OpGetLocal, symbol.Index)
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

	case *ast.HashLiteral:
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
		outerInstructions := c.instructions
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

		fnInstructions := c.instructions
		numLocals := c.symbolTable.numDefinitions
		freeVars := c.symbolTable.Free
		numFree := len(freeVars)

		// Get the nested free symbols before exiting scope
		// These are the locals from this scope that nested functions reference
		nestedFreeSymbols := c.symbolTable.NestedFreeSymbols

		c.symbolTable = c.symbolTable.outer

		// Note: We don't need to adjust local variable indices anymore
		// because free variables are stored in the closure object, not in the local variable table
		// Local variables always start at index 0

		// Combine freeVars and nestedFreeSymbols for the CompiledFunction
		// nestedFreeSymbols are locals that nested functions need as free variables
		allFreeVars := make([]Symbol, 0, len(freeVars)+len(nestedFreeSymbols))
		allFreeVars = append(allFreeVars, freeVars...)
		for _, nfs := range nestedFreeSymbols {
			// Convert nested free symbol to FreeScope
			allFreeVars = append(allFreeVars, Symbol{
				Name:  nfs.Name,
				Scope: FreeScope,
				Index: len(allFreeVars),
			})
		}

		compiledFn := &CompiledFunction{
			Instructions:  fnInstructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
			IsGenerator:   c.hasYieldInBody(node.Body),
			Free:          allFreeVars,
		}

		c.instructions = make(Instructions, 0, len(outerInstructions))
		c.instructions = append(c.instructions, outerInstructions...)

		// Determine if this function needs to be a closure
		// It needs to be a closure if:
		// 1. It has its own free variables (references outer scope), OR
		// 2. Nested functions reference this function's locals
		needsClosure := numFree > 0 || len(nestedFreeSymbols) > 0

		if needsClosure {
			// Emit instructions to load free variables onto stack
			// First, load this function's own free variables
			for _, free := range freeVars {
				if free.Scope == GlobalScope {
					c.emit(OpGetGlobal, free.Index)
				} else {
					c.emit1(OpGetLocal, free.Index)
				}
			}
			// Then, load the nested free variables (locals that nested functions need)
			for _, nestedFree := range nestedFreeSymbols {
				c.emit1(OpGetLocal, nestedFree.Index)
			}
			totalFree := numFree + len(nestedFreeSymbols)
			c.emitClosure(c.addConstant(compiledFn), totalFree)
		} else {
			c.emit(OpConstant, c.addConstant(compiledFn))
		}
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
		if err := c.Compile(node.Function); err != nil {
			return err
		}

		for _, a := range node.Arguments {
			if err := c.Compile(a); err != nil {
				return err
			}
		}

		c.emit1(OpCall, len(node.Arguments))

	case *ast.PassStatement:
		// pass is a no-op, do nothing

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
				c.emit1(OpSetLocal, symbol.Index)
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
	case *ast.ClassStatement:
		return c.compileClassStatement(node)
	case *ast.MemberAccess:
		return c.compileMemberAccess(node)
	case *ast.MethodCall:
		return c.compileMethodCall(node)
	case *ast.ImportStatement:
		return c.compileImportStatement(node)
	case *ast.FromImportStatement:
		return c.compileFromImportStatement(node)
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
					c.emit1(OpSetLocal, varSymbol.Index)
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

func (c *Compiler) addConstant(obj objects.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) compileImportStatement(node *ast.ImportStatement) error {
	moduleName := node.Module.Value
	alias := moduleName
	if node.Alias != nil {
		alias = node.Alias.Value
	}
	
	module := objects.GetModule(moduleName)
	if module == nil {
		return fmt.Errorf("module '%s' not found", moduleName)
	}
	
	var symbol Symbol
	if existingSymbol, ok := c.symbolTable.Resolve(alias); ok {
		if existingSymbol.Scope == BuiltinScope {
			symbol = c.symbolTable.Define(alias)
		} else {
			symbol = existingSymbol
		}
	} else {
		symbol = c.symbolTable.Define(alias)
	}
	
	c.addConstant(module)
	c.emit(OpConstant, len(c.constants)-1)
	c.emit(OpSetGlobal, symbol.Index)
	
	return nil
}

func (c *Compiler) compileFromImportStatement(node *ast.FromImportStatement) error {
	moduleName := node.Module.Value
	
	module := objects.GetModule(moduleName)
	if module == nil {
		return fmt.Errorf("module '%s' not found", moduleName)
	}
	
	if node.Alias != nil {
		symbol := c.symbolTable.Define(node.Alias.Value)
		c.addConstant(module)
		c.emit(OpConstant, len(c.constants)-1)
		c.emit(OpSetGlobal, symbol.Index)
	} else {
		for _, name := range node.Names {
			value, ok := module.Fields[name.Value]
			if !ok {
				return fmt.Errorf("name '%s' not found in module '%s'", name.Value, moduleName)
			}
			
			symbol := c.symbolTable.Define(name.Value)
			c.addConstant(value)
			c.emit(OpConstant, len(c.constants)-1)
			c.emit(OpSetGlobal, symbol.Index)
		}
	}
	
	return nil
}

func (c *Compiler) compileClassStatement(node *ast.ClassStatement) error {
	class := &objects.Class{
		Name:    node.Name.Value,
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}

	for _, method := range node.Methods {
		compiledFn := c.compileFunction(method)
		if compiledFn != nil {
			class.Methods[method.Name] = compiledFn
		}
	}

	// Handle inheritance
	if node.SuperClass != nil {
		// First, emit instruction to get the super class
		superClassIdx, ok := c.symbolTable.Resolve(node.SuperClass.Value)
		if ok && superClassIdx.Scope == GlobalScope {
			c.emit(OpGetGlobal, superClassIdx.Index)
		}
		// Emit OpCreateClass with inheritance
		c.emit(OpCreateClassWithSuper, c.addConstant(class))
	} else {
		c.emit(OpCreateClass, c.addConstant(class))
	}

	// Define class name in symbol table (after OpCreateClass so stack has the value)
	symbol := c.symbolTable.Define(node.Name.Value)
	c.emit(OpSetGlobal, symbol.Index)
	
	return nil
}

func (c *Compiler) compileFunction(fn *ast.FunctionLiteral) *CompiledFunction {
	savedInstructions := c.instructions
	c.instructions = []byte{}

	c.enterScope()

	for _, param := range fn.Parameters {
		c.symbolTable.Define(param.Value)
	}

	for _, stmt := range fn.Body.Statements {
		if err := c.Compile(stmt); err != nil {
			return nil
		}
	}

	if c.lastInstruction.Opcode != OpReturnValue && c.lastInstruction.Opcode != OpReturn {
		c.emit(OpNull)
		c.emit(OpReturnValue)
	}

	numLocals := c.symbolTable.numDefinitions
	free := c.symbolTable.Free
	c.exitScope()

	// Get function instructions before restoring outer scope instructions
	fnInstructions := c.instructions
	
	// Restore outer scope instructions
	c.instructions = savedInstructions

	return &CompiledFunction{
		Instructions:   fnInstructions,
		NumLocals:      numLocals,
		NumParameters:  len(fn.Parameters),
		Free:           free,
	}
}

func (c *Compiler) compileMemberAccess(node *ast.MemberAccess) error {
	if err := c.Compile(node.Object); err != nil {
		return err
	}
	c.emit(OpGetAttribute, c.addConstant(&objects.String{Value: node.Member.Value}))
	return nil
}

func (c *Compiler) compileMethodCall(node *ast.MethodCall) error {
	if err := c.Compile(node.Object); err != nil {
		return err
	}
	
	c.emit(OpGetAttribute, c.addConstant(&objects.String{Value: node.Method.Value}))
	
	for _, arg := range node.Arguments {
		if err := c.Compile(arg); err != nil {
			return err
		}
	}
	
	c.emit1(OpCall, len(node.Arguments))
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (c *Compiler) compileListComprehension(node *ast.ListComprehension) error {
	compilationScope := c.instructions
	c.instructions = []byte{}
	c.enterScope()

	iterSymbol := c.symbolTable.Define("__iter__")
	c.emit(OpGetGlobal, iterSymbol.Index)
	c.emit1(OpCall, 0)
	
	loopStart := len(c.instructions)

	c.emit(OpGetGlobal, iterSymbol.Index)
	c.emit1(OpCall, 0)
	c.emit(OpMakeGenerator)

	endPos := c.emit(OpJump, 0)

	c.changeOperand(endPos, len(c.instructions))

	c.exitScope()
	compilationScope = append(compilationScope, c.instructions...)

	c.enterScope()

	if err := c.Compile(node.Element); err != nil {
		return err
	}

	if node.Filter != nil {
		if err := c.Compile(node.Filter); err != nil {
			return err
		}
		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 0)
		c.changeOperand(jumpNotTruthyPos, loopStart)
	}

	c.emit(OpYieldValue)
	c.emit(OpPop)

	c.exitScope()

	c.instructions = append(c.instructions, compilationScope...)

	return nil
}

func (c *Compiler) compileSetComprehension(node *ast.SetComprehension) error {
	if err := c.compileListComprehension(&ast.ListComprehension{
		Token:    node.Token,
		Element:  node.Element,
		Variable: node.Variable,
		Iterable: node.Iterable,
		Filter:  node.Filter,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) compileDictComprehension(node *ast.DictComprehension) error {
	compilationScope := c.instructions
	c.instructions = []byte{}
	c.enterScope()

	iterSymbol := c.symbolTable.Define("__iter__")
	c.emit(OpGetGlobal, iterSymbol.Index)
	c.emit1(OpCall, 0)

	loopStart := len(c.instructions)

	c.emit(OpGetGlobal, iterSymbol.Index)
	c.emit1(OpCall, 0)
	c.emit(OpMakeGenerator)

	endPos := c.emit(OpJump, 0)
	c.changeOperand(endPos, len(c.instructions))
	c.exitScope()

	compilationScope = append(compilationScope, c.instructions...)

	c.enterScope()

	if err := c.Compile(node.Key); err != nil {
		return err
	}
	if err := c.Compile(node.Value); err != nil {
		return err
	}
	c.emit(OpHash, 2)

	if node.Filter != nil {
		if err := c.Compile(node.Filter); err != nil {
			return err
		}
		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 0)
		c.changeOperand(jumpNotTruthyPos, loopStart)
	}

	c.emit(OpYieldValue)
	c.emit(OpPop)

	c.exitScope()

	c.instructions = append(c.instructions, compilationScope...)

	return nil
}