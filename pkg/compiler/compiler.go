package compiler

import (
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/go-py/go-python/pkg/ast"
	"github.com/go-py/go-python/pkg/concurrency"
	"github.com/go-py/go-python/pkg/gc"
	"github.com/go-py/go-python/pkg/interop"
	"github.com/go-py/go-python/pkg/objects"
	re "github.com/go-py/go-python/pkg/re"
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
	OpMod
	OpFloorDiv
	OpPower
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
	OpExceptStarHandler
	OpFinally
	OpYield
	OpEnterContext
	OpExitContext
	OpMakeGenerator
	OpYieldValue
	OpCreateClass
	OpCreateClassWithSuper
	OpCreateClassWithMultiSuper
	OpGetAttribute
	OpSetAttribute
	OpDelAttribute
	OpSetClassField
	OpSetMetaclass
	OpCallMetaclassInit
	OpFormatString
	OpMakeAsync
	OpAwait
	OpListUnpack
	OpDictUnpack
	OpEllipsis
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
	return EliminateDeadCodeInFunctions(&Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	})
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

	// 注册 concurrency 模块
	concurrencyModule := concurrency.CreateConcurrencyModule()
	objects.RegisterModule("concurrency", concurrencyModule)
	concurrencyIndex := len(c.constants)
	c.constants = append(c.constants, concurrencyModule)
	c.symbolTable.DefineBuiltin("concurrency", concurrencyIndex)

	// 注册 re 模块
	reModule := re.CreateReModule()
	objects.RegisterModule("re", reModule)
	reIndex := len(c.constants)
	c.constants = append(c.constants, reModule)
	c.symbolTable.DefineBuiltin("re", reIndex)

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
			case *objects.Range:
				return &objects.Integer{Value: arg.Len()}
			case *objects.Zip:
				l := arg.Len()
				if l < 0 {
					return objects.NewError("cannot determine length of zip with non-sequence argument")
				}
				return &objects.Integer{Value: l}
			case *objects.Bytes:
				return &objects.Integer{Value: int64(len(arg.Value))}
			case *objects.DictKeys:
				return &objects.Integer{Value: arg.Len()}
			case *objects.DictValues:
				return &objects.Integer{Value: arg.Len()}
			case *objects.DictItems:
				return &objects.Integer{Value: arg.Len()}
			default:
				return objects.NewError("argument to 'len' not supported: %s", arg.Type())
			}
		},
	}
	lenIndex := len(c.constants)
	c.constants = append(c.constants, lenBuiltin)
	c.symbolTable.DefineBuiltin("len", lenIndex)

	noneIndex := len(c.constants)
	c.constants = append(c.constants, objects.None_)
	c.symbolTable.DefineBuiltin("None", noneIndex)

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
			if err := objects.CheckHashable(args[1]); err != nil {
				return err.(*objects.Error)
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
			if err := objects.CheckHashable(args[1]); err != nil {
				return err.(*objects.Error)
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
			case *objects.Complex:
				return objects.NewTypeError("can't convert complex to int")
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
			case *objects.Complex:
				return objects.NewTypeError("can't convert complex to float")
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
			case *objects.Bytes:
				if len(arg.Value) > 0 {
					return objects.True
				}
				return objects.False
			case *objects.None:
				return objects.False
			case *objects.Complex:
				if arg.Real != 0 || arg.Imag != 0 {
					return objects.True
				}
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
			case *objects.Complex:
				magnitude := cmplx.Abs(complex(arg.Real, arg.Imag))
				return &objects.Float{Value: magnitude}
			default:
				return objects.NewError("abs() argument must be a number")
			}
		},
	}
	absIndex := len(c.constants)
	c.constants = append(c.constants, absBuiltin)
	c.symbolTable.DefineBuiltin("abs", absIndex)

	complexBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewTypeError("complex() takes exactly 2 arguments")
			}
			var realPart, imagPart float64
			switch arg := args[0].(type) {
			case *objects.Integer:
				realPart = float64(arg.Value)
			case *objects.Float:
				realPart = arg.Value
			case *objects.Complex:
				return objects.NewTypeError("complex() first argument must be a real number, not complex")
			default:
				return objects.NewTypeError("complex() argument must be a number")
			}
			switch arg := args[1].(type) {
			case *objects.Integer:
				imagPart = float64(arg.Value)
			case *objects.Float:
				imagPart = arg.Value
			case *objects.Complex:
				return objects.NewTypeError("complex() second argument must be a real number, not complex")
			default:
				return objects.NewTypeError("complex() argument must be a number")
			}
			return objects.NewComplex(realPart, imagPart)
		},
	}
	complexIndex := len(c.constants)
	c.constants = append(c.constants, complexBuiltin)
	c.symbolTable.DefineBuiltin("complex", complexIndex)

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
			return objects.NewRange(start, stop, step)
		},
	}
	rangeIndex := len(c.constants)
	c.constants = append(c.constants, rangeBuiltin)
	c.symbolTable.DefineBuiltin("range", rangeIndex)

	listBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("list() takes exactly 1 argument")
			}
			switch arg := args[0].(type) {
			case *objects.List:
				// Return a copy
				elements := make([]objects.Object, len(arg.Elements))
				copy(elements, arg.Elements)
				return &objects.List{Elements: elements}
			case *objects.Tuple:
				return &objects.List{Elements: arg.Elements}
			case *objects.String:
				elements := make([]objects.Object, len(arg.Value))
				for i, ch := range arg.Value {
					elements[i] = &objects.String{Value: string(ch)}
				}
				return &objects.List{Elements: elements}
			case *objects.Range:
				return &objects.List{Elements: arg.ToList()}
			case *objects.Zip:
				elements := arg.ToList()
				if elements == nil {
					return objects.NewTypeError("list() cannot convert zip with non-sequence argument")
				}
				return &objects.List{Elements: elements}
			case *objects.DictKeys:
				return &objects.List{Elements: arg.ToList()}
			case *objects.DictValues:
				return &objects.List{Elements: arg.ToList()}
			case *objects.DictItems:
				return &objects.List{Elements: arg.ToList()}
			case *objects.Dict:
				// list(dict) returns list of keys, same as Python
				return &objects.List{Elements: arg.KeysSlice()}
			case *objects.Set:
				return &objects.List{Elements: arg.ToSlice()}
			default:
				return objects.NewTypeError("'%s' object is not iterable", arg.Type())
			}
		},
	}
	listIndex := len(c.constants)
	c.constants = append(c.constants, listBuiltin)
	c.symbolTable.DefineBuiltin("list", listIndex)

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
			return objects.NewZip(args)
		},
	}
	zipIndex := len(c.constants)
	c.constants = append(c.constants, zipBuiltin)
	c.symbolTable.DefineBuiltin("zip", zipIndex)

	enumBuiltin := &objects.Builtin{
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewError("__enum__ requires 2 arguments: name and members")
			}
			nameObj, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewError("__enum__ first argument must be a string")
			}
			membersDict, ok := args[1].(*objects.Dict)
			if !ok {
				return objects.NewError("__enum__ second argument must be a dict")
			}
			members := make(map[string]objects.Object)
			for k, v := range membersDict.Pairs {
				members[k] = v
			}
			return objects.NewEnum(nameObj.Value, members)
		},
	}
	enumIndex := len(c.constants)
	c.constants = append(c.constants, enumBuiltin)
	c.symbolTable.DefineBuiltin("__enum__", enumIndex)

	propertyBuiltin := &objects.Builtin{
		Name: "property",
		Fn: func(args ...objects.Object) objects.Object {
			prop := &objects.Property{}
			if len(args) >= 1 {
				prop.Fget = args[0]
			}
			if len(args) >= 2 {
				prop.Fset = args[1]
			}
			if len(args) >= 3 {
				prop.Fdel = args[2]
			}
			return prop
		},
	}
	propertyIndex := len(c.constants)
	c.constants = append(c.constants, propertyBuiltin)
	c.symbolTable.DefineBuiltin("property", propertyIndex)

	classmethodBuiltin := &objects.Builtin{
		Name: "classmethod",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("classmethod() takes exactly 1 argument")
			}
			return &objects.ClassMethod{Fn: args[0]}
		},
	}
	classmethodIndex := len(c.constants)
	c.constants = append(c.constants, classmethodBuiltin)
	c.symbolTable.DefineBuiltin("classmethod", classmethodIndex)

	staticmethodBuiltin := &objects.Builtin{
		Name: "staticmethod",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewError("staticmethod() takes exactly 1 argument")
			}
			return &objects.StaticMethod{Fn: args[0]}
		},
	}
	staticmethodIndex := len(c.constants)
	c.constants = append(c.constants, staticmethodBuiltin)
	c.symbolTable.DefineBuiltin("staticmethod", staticmethodIndex)

	superBuiltin := &objects.Builtin{
		Name: "super",
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.Super{
				Instance:   nil,
				SuperClass: nil,
			}
		},
	}
	superIndex := len(c.constants)
	c.constants = append(c.constants, superBuiltin)
	c.symbolTable.DefineBuiltin("super", superIndex)

	exceptionGroupBuiltin := &objects.Builtin{
		Name: "ExceptionGroup",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("ExceptionGroup() takes at least 2 arguments (%d given)", len(args))
			}
			message, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("ExceptionGroup() first argument must be a string")
			}
			excList, ok := args[1].(*objects.List)
			if !ok {
				return objects.NewTypeError("ExceptionGroup() second argument must be a list")
			}
			var exceptions []objects.Object
			for _, exc := range excList.Elements {
				if exc.Type() != objects.ERROR_OBJ && exc.Type() != objects.EXCEPTION_GROUP_OBJ {
					return objects.NewTypeError("ExceptionGroup() second argument must contain exceptions")
				}
				exceptions = append(exceptions, exc)
			}
			return &objects.ExceptionGroup{
				Message:    message.Value,
				Exceptions: exceptions,
			}
		},
	}
	exceptionGroupIndex := len(c.constants)
	c.constants = append(c.constants, exceptionGroupBuiltin)
	c.symbolTable.DefineBuiltin("ExceptionGroup", exceptionGroupIndex)

	// Exception type constructors
	exceptionTypes := []struct {
		name      string
		errorType string
	}{
		{"TypeError", "TypeError"},
		{"ValueError", "ValueError"},
		{"KeyError", "KeyError"},
		{"IndexError", "IndexError"},
		{"AttributeError", "AttributeError"},
		{"ZeroDivisionError", "ZeroDivisionError"},
		{"RuntimeError", "RuntimeError"},
		{"NameError", "NameError"},
		{"StopIteration", "StopIteration"},
		{"NotImplementedError", "NotImplementedError"},
		{"OverflowError", "OverflowError"},
	}

	for _, et := range exceptionTypes {
		errorType := et.errorType
		builtin := &objects.Builtin{
			Name: et.name,
			Fn: func(args ...objects.Object) objects.Object {
				msg := ""
				if len(args) > 0 {
					if s, ok := args[0].(*objects.String); ok {
						msg = s.Value
					} else {
						msg = args[0].Inspect()
					}
				}
				return objects.NewErrorWithType(errorType, "%s", msg)
			},
		}
		idx := len(c.constants)
		c.constants = append(c.constants, builtin)
		c.symbolTable.DefineBuiltin(et.name, idx)
	}

	// --- Additional Python builtin functions ---

	// 1. isinstance(obj, cls)
	isinstanceBuiltin := &objects.Builtin{
		Name: "isinstance",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewTypeError("isinstance() takes exactly 2 arguments")
			}
			obj := args[0]
			cls := args[1]

			// Handle tuple of classes
			if clsTuple, ok := cls.(*objects.Tuple); ok {
				for _, c := range clsTuple.Elements {
					if classObj, ok := c.(*objects.Class); ok {
						if objects.IsInstanceOf(obj, classObj) {
							return objects.True
						}
					}
				}
				return objects.False
			}

			classObj, ok := cls.(*objects.Class)
			if !ok {
				return objects.NewTypeError("isinstance() arg 2 must be a class or tuple of classes")
			}
			if objects.IsInstanceOf(obj, classObj) {
				return objects.True
			}
			return objects.False
		},
	}
	isinstanceIndex := len(c.constants)
	c.constants = append(c.constants, isinstanceBuiltin)
	c.symbolTable.DefineBuiltin("isinstance", isinstanceIndex)

	// 2. issubclass(cls, parent)
	issubclassBuiltin := &objects.Builtin{
		Name: "issubclass",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewTypeError("issubclass() takes exactly 2 arguments")
			}
			cls, ok := args[0].(*objects.Class)
			if !ok {
				return objects.NewTypeError("issubclass() arg 1 must be a class")
			}
			parent := args[1]

			if parentTuple, ok := parent.(*objects.Tuple); ok {
				for _, p := range parentTuple.Elements {
					if parentCls, ok := p.(*objects.Class); ok {
						if objects.IsSubclassOf(cls, parentCls) {
							return objects.True
						}
					}
				}
				return objects.False
			}

			parentCls, ok := parent.(*objects.Class)
			if !ok {
				return objects.NewTypeError("issubclass() arg 2 must be a class or tuple of classes")
			}
			if objects.IsSubclassOf(cls, parentCls) {
				return objects.True
			}
			return objects.False
		},
	}
	issubclassIndex := len(c.constants)
	c.constants = append(c.constants, issubclassBuiltin)
	c.symbolTable.DefineBuiltin("issubclass", issubclassIndex)

	// 3. hasattr(obj, name)
	hasattrBuiltin := &objects.Builtin{
		Name: "hasattr",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewTypeError("hasattr() takes exactly 2 arguments")
			}
			name, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("hasattr(): attribute name must be a string")
			}
			switch obj := args[0].(type) {
			case *objects.Instance:
				if _, ok := obj.Fields[name.Value]; ok {
					return objects.True
				}
				if _, ok := obj.Class.Methods[name.Value]; ok {
					return objects.True
				}
			case *objects.Module:
				if _, ok := obj.Fields[name.Value]; ok {
					return objects.True
				}
			case *objects.Class:
				if _, ok := obj.Methods[name.Value]; ok {
					return objects.True
				}
			}
			return objects.False
		},
	}
	hasattrIndex := len(c.constants)
	c.constants = append(c.constants, hasattrBuiltin)
	c.symbolTable.DefineBuiltin("hasattr", hasattrIndex)

	// 4. getattr(obj, name[, default])
	getattrBuiltin := &objects.Builtin{
		Name: "getattr",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("getattr() takes at least 2 arguments")
			}
			name, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("getattr(): attribute name must be a string")
			}
			switch obj := args[0].(type) {
			case *objects.Instance:
				if val, ok := obj.Fields[name.Value]; ok {
					return val
				}
				if method, ok := obj.Class.Methods[name.Value]; ok {
					return &objects.BoundMethod{Self: obj, Method: method}
				}
			case *objects.Module:
				if val, ok := obj.Fields[name.Value]; ok {
					return val
				}
			case *objects.Class:
				if val, ok := obj.Methods[name.Value]; ok {
					return val
				}
			case *objects.RegexPattern:
				if val, ok := obj.GetAttr(name.Value); ok {
					return val
				}
			case *objects.RegexMatch:
				if val, ok := obj.GetAttr(name.Value); ok {
					return val
				}
			}
			if len(args) >= 3 {
				return args[2] // default value
			}
			return objects.NewAttributeError("'%s' object has no attribute '%s'", args[0].Type(), name.Value)
		},
	}
	getattrIndex := len(c.constants)
	c.constants = append(c.constants, getattrBuiltin)
	c.symbolTable.DefineBuiltin("getattr", getattrIndex)

	// 5. setattr(obj, name, value)
	setattrBuiltin := &objects.Builtin{
		Name: "setattr",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 3 {
				return objects.NewTypeError("setattr() takes exactly 3 arguments")
			}
			name, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("setattr(): attribute name must be a string")
			}
			switch obj := args[0].(type) {
			case *objects.Instance:
				obj.Fields[name.Value] = args[2]
			case *objects.Module:
				obj.Fields[name.Value] = args[2]
			default:
				return objects.NewAttributeError("cannot set attribute '%s' on '%s' object", name.Value, args[0].Type())
			}
			return objects.None_
		},
	}
	setattrIndex := len(c.constants)
	c.constants = append(c.constants, setattrBuiltin)
	c.symbolTable.DefineBuiltin("setattr", setattrIndex)

	// 6. dir(obj)
	dirBuiltin := &objects.Builtin{
		Name: "dir",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("dir() takes exactly 1 argument")
			}
			var names []string
			switch obj := args[0].(type) {
			case *objects.Instance:
				for k := range obj.Fields {
					names = append(names, k)
				}
				for k := range obj.Class.Methods {
					names = append(names, k)
				}
			case *objects.Module:
				for k := range obj.Fields {
					names = append(names, k)
				}
			case *objects.Class:
				for k := range obj.Methods {
					names = append(names, k)
				}
			}
			sort.Strings(names)
			elements := make([]objects.Object, len(names))
			for i, n := range names {
				elements[i] = &objects.String{Value: n}
			}
			return &objects.List{Elements: elements}
		},
	}
	dirIndex := len(c.constants)
	c.constants = append(c.constants, dirBuiltin)
	c.symbolTable.DefineBuiltin("dir", dirIndex)

	// 7. id(obj)
	idBuiltin := &objects.Builtin{
		Name: "id",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("id() takes exactly 1 argument")
			}
			return &objects.Integer{Value: int64(reflect.ValueOf(args[0]).Pointer())}
		},
	}
	idIndex := len(c.constants)
	c.constants = append(c.constants, idBuiltin)
	c.symbolTable.DefineBuiltin("id", idIndex)

	// 8. hash(obj)
	hashBuiltin := &objects.Builtin{
		Name: "hash",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("hash() takes exactly 1 argument")
			}
			switch obj := args[0].(type) {
			case *objects.Integer:
				return &objects.Integer{Value: obj.Value}
			case *objects.String:
				h := int64(0)
				for _, c := range obj.Value {
					h = h*31 + int64(c)
				}
				return &objects.Integer{Value: h}
			case *objects.Boolean:
				if obj.Value {
					return &objects.Integer{Value: 1}
				}
				return &objects.Integer{Value: 0}
			case *objects.Float:
				return &objects.Integer{Value: int64(math.Float64bits(obj.Value))}
			default:
				return &objects.Integer{Value: int64(reflect.ValueOf(args[0]).Pointer())}
			}
		},
	}
	hashIndex := len(c.constants)
	c.constants = append(c.constants, hashBuiltin)
	c.symbolTable.DefineBuiltin("hash", hashIndex)

	// 9. callable(obj)
	callableBuiltin := &objects.Builtin{
		Name: "callable",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("callable() takes exactly 1 argument")
			}
			if objects.IsCallable(args[0]) {
				return objects.True
			}
			return objects.False
		},
	}
	callableIndex := len(c.constants)
	c.constants = append(c.constants, callableBuiltin)
	c.symbolTable.DefineBuiltin("callable", callableIndex)

	// 10. enumerate(iterable[, start])
	enumerateBuiltin := &objects.Builtin{
		Name: "enumerate",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("enumerate() takes at least 1 argument")
			}
			start := int64(0)
			if len(args) >= 2 {
				if s, ok := args[1].(*objects.Integer); ok {
					start = s.Value
				}
			}
			var elements []objects.Object
			switch iter := args[0].(type) {
			case *objects.List:
				for i, elem := range iter.Elements {
					elements = append(elements, &objects.Tuple{Elements: []objects.Object{
						&objects.Integer{Value: start + int64(i)},
						elem,
					}})
				}
			case *objects.Tuple:
				for i, elem := range iter.Elements {
					elements = append(elements, &objects.Tuple{Elements: []objects.Object{
						&objects.Integer{Value: start + int64(i)},
						elem,
					}})
				}
			case *objects.String:
				for i, ch := range iter.Value {
					elements = append(elements, &objects.Tuple{Elements: []objects.Object{
						&objects.Integer{Value: start + int64(i)},
						&objects.String{Value: string(ch)},
					}})
				}
			case *objects.Range:
				items := iter.ToList()
				for i, elem := range items {
					elements = append(elements, &objects.Tuple{Elements: []objects.Object{
						&objects.Integer{Value: start + int64(i)},
						elem,
					}})
				}
			default:
				return objects.NewTypeError("'%s' object is not iterable", args[0].Type())
			}
			return &objects.List{Elements: elements}
		},
	}
	enumerateIndex := len(c.constants)
	c.constants = append(c.constants, enumerateBuiltin)
	c.symbolTable.DefineBuiltin("enumerate", enumerateIndex)

	// 11. map(func, iterable)
	mapBuiltin := &objects.Builtin{
		Name: "map",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewTypeError("map() takes exactly 2 arguments")
			}
			fn := args[0]
			var items []objects.Object
			switch iter := args[1].(type) {
			case *objects.List:
				items = iter.Elements
			case *objects.Tuple:
				items = iter.Elements
			default:
				return objects.NewTypeError("'%s' object is not iterable", args[1].Type())
			}
			result := make([]objects.Object, 0, len(items))
			for _, item := range items {
				mapped := objects.CallFunction(fn, item)
				if mapped.Type() == objects.ERROR_OBJ {
					return mapped
				}
				result = append(result, mapped)
			}
			return &objects.List{Elements: result}
		},
	}
	mapIndex := len(c.constants)
	c.constants = append(c.constants, mapBuiltin)
	c.symbolTable.DefineBuiltin("map", mapIndex)

	// 12. filter(func, iterable)
	filterBuiltin := &objects.Builtin{
		Name: "filter",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 2 {
				return objects.NewTypeError("filter() takes exactly 2 arguments")
			}
			fn := args[0]
			var items []objects.Object
			switch iter := args[1].(type) {
			case *objects.List:
				items = iter.Elements
			case *objects.Tuple:
				items = iter.Elements
			default:
				return objects.NewTypeError("'%s' object is not iterable", args[1].Type())
			}
			result := make([]objects.Object, 0)
			for _, item := range items {
				filtered := objects.CallFunction(fn, item)
				if filtered.Type() == objects.ERROR_OBJ {
					return filtered
				}
				if isTruthy(filtered) {
					result = append(result, item)
				}
			}
			return &objects.List{Elements: result}
		},
	}
	filterIndex := len(c.constants)
	c.constants = append(c.constants, filterBuiltin)
	c.symbolTable.DefineBuiltin("filter", filterIndex)

	// 13. sorted(iterable[, key][, reverse])
	sortedBuiltin := &objects.Builtin{
		Name: "sorted",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("sorted() takes at least 1 argument")
			}
			var items []objects.Object
			switch iter := args[0].(type) {
			case *objects.List:
				items = append([]objects.Object{}, iter.Elements...)
			case *objects.Tuple:
				items = append([]objects.Object{}, iter.Elements...)
			default:
				return objects.NewTypeError("'%s' object is not iterable", args[0].Type())
			}
			var keyFn objects.Object
			reverse := false
			if len(args) >= 2 {
				if _, ok := args[1].(*objects.Boolean); ok {
					reverse = args[1].(*objects.Boolean).Value
				} else if args[1] != objects.None_ {
					keyFn = args[1]
				}
			}
			if len(args) >= 3 {
				if b, ok := args[2].(*objects.Boolean); ok {
					reverse = b.Value
				}
			}
			sort.SliceStable(items, func(i, j int) bool {
				var aVal, bVal objects.Object
				if keyFn != nil {
					aVal = objects.CallFunction(keyFn, items[i])
					bVal = objects.CallFunction(keyFn, items[j])
				} else {
					aVal = items[i]
					bVal = items[j]
				}
				cmp := compareObjects(aVal, bVal)
				if reverse {
					return cmp > 0
				}
				return cmp < 0
			})
			return &objects.List{Elements: items}
		},
	}
	sortedIndex := len(c.constants)
	c.constants = append(c.constants, sortedBuiltin)
	c.symbolTable.DefineBuiltin("sorted", sortedIndex)

	// 14. reversed(iterable)
	reversedBuiltin := &objects.Builtin{
		Name: "reversed",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("reversed() takes exactly 1 argument")
			}
			switch iter := args[0].(type) {
			case *objects.List:
				n := len(iter.Elements)
				result := make([]objects.Object, n)
				for i, elem := range iter.Elements {
					result[n-1-i] = elem
				}
				return &objects.List{Elements: result}
			case *objects.Tuple:
				n := len(iter.Elements)
				result := make([]objects.Object, n)
				for i, elem := range iter.Elements {
					result[n-1-i] = elem
				}
				return &objects.List{Elements: result}
			default:
				return objects.NewTypeError("'%s' object is not reversible", args[0].Type())
			}
		},
	}
	reversedIndex := len(c.constants)
	c.constants = append(c.constants, reversedBuiltin)
	c.symbolTable.DefineBuiltin("reversed", reversedIndex)

	// 15. repr(obj)
	reprBuiltin := &objects.Builtin{
		Name: "repr",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("repr() takes exactly 1 argument")
			}
			return &objects.String{Value: args[0].Inspect()}
		},
	}
	reprIndex := len(c.constants)
	c.constants = append(c.constants, reprBuiltin)
	c.symbolTable.DefineBuiltin("repr", reprIndex)

	// 16. iter(obj)
	iterBuiltin := &objects.Builtin{
		Name: "iter",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("iter() takes exactly 1 argument")
			}
			return args[0]
		},
	}
	iterIndex := len(c.constants)
	c.constants = append(c.constants, iterBuiltin)
	c.symbolTable.DefineBuiltin("iter", iterIndex)

	// 17. any(iterable)
	anyBuiltin := &objects.Builtin{
		Name: "any",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("any() takes exactly 1 argument")
			}
			var items []objects.Object
			switch iter := args[0].(type) {
			case *objects.List:
				items = iter.Elements
			case *objects.Tuple:
				items = iter.Elements
			default:
				return objects.NewTypeError("'%s' object is not iterable", args[0].Type())
			}
			for _, item := range items {
				if isTruthy(item) {
					return objects.True
				}
			}
			return objects.False
		},
	}
	anyIndex := len(c.constants)
	c.constants = append(c.constants, anyBuiltin)
	c.symbolTable.DefineBuiltin("any", anyIndex)

	// 18. all(iterable)
	allBuiltin := &objects.Builtin{
		Name: "all",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("all() takes exactly 1 argument")
			}
			var items []objects.Object
			switch iter := args[0].(type) {
			case *objects.List:
				items = iter.Elements
			case *objects.Tuple:
				items = iter.Elements
			default:
				return objects.NewTypeError("'%s' object is not iterable", args[0].Type())
			}
			for _, item := range items {
				if !isTruthy(item) {
					return objects.False
				}
			}
			return objects.True
		},
	}
	allIndex := len(c.constants)
	c.constants = append(c.constants, allBuiltin)
	c.symbolTable.DefineBuiltin("all", allIndex)

	// 19. chr(i)
	chrBuiltin := &objects.Builtin{
		Name: "chr",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("chr() takes exactly 1 argument")
			}
			i, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("chr() argument must be an integer")
			}
			if i.Value < 0 || i.Value > 0x10FFFF {
				return objects.NewValueError("chr() arg not in range(0x110000)")
			}
			return &objects.String{Value: string(rune(i.Value))}
		},
	}
	chrIndex := len(c.constants)
	c.constants = append(c.constants, chrBuiltin)
	c.symbolTable.DefineBuiltin("chr", chrIndex)

	// 20. ord(c)
	ordBuiltin := &objects.Builtin{
		Name: "ord",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("ord() takes exactly 1 argument")
			}
			s, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("ord() argument must be a string")
			}
			runes := []rune(s.Value)
			if len(runes) != 1 {
				return objects.NewTypeError("ord() expected a character, but string of length %d found", len(runes))
			}
			return &objects.Integer{Value: int64(runes[0])}
		},
	}
	ordIndex := len(c.constants)
	c.constants = append(c.constants, ordBuiltin)
	c.symbolTable.DefineBuiltin("ord", ordIndex)

	// 21. hex(i)
	hexBuiltin := &objects.Builtin{
		Name: "hex",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("hex() takes exactly 1 argument")
			}
			i, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("hex() argument must be an integer")
			}
			return &objects.String{Value: fmt.Sprintf("0x%x", i.Value)}
		},
	}
	hexIndex := len(c.constants)
	c.constants = append(c.constants, hexBuiltin)
	c.symbolTable.DefineBuiltin("hex", hexIndex)

	// 22. oct(i)
	octBuiltin := &objects.Builtin{
		Name: "oct",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("oct() takes exactly 1 argument")
			}
			i, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("oct() argument must be an integer")
			}
			return &objects.String{Value: fmt.Sprintf("0o%o", i.Value)}
		},
	}
	octIndex := len(c.constants)
	c.constants = append(c.constants, octBuiltin)
	c.symbolTable.DefineBuiltin("oct", octIndex)

	// 23. bin(i)
	binBuiltin := &objects.Builtin{
		Name: "bin",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("bin() takes exactly 1 argument")
			}
			i, ok := args[0].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("bin() argument must be an integer")
			}
			return &objects.String{Value: fmt.Sprintf("0b%b", i.Value)}
		},
	}
	binIndex := len(c.constants)
	c.constants = append(c.constants, binBuiltin)
	c.symbolTable.DefineBuiltin("bin", binIndex)

	// 24. format(value[, format_spec])
	formatNewBuiltin := &objects.Builtin{
		Name: "format",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 || len(args) > 2 {
				return objects.NewTypeError("format() takes 1 or 2 arguments")
			}
			formatSpec := ""
			if len(args) >= 2 {
				if s, ok := args[1].(*objects.String); ok {
					formatSpec = s.Value
				}
			}
			switch val := args[0].(type) {
			case *objects.Integer:
				if formatSpec == "" {
					return &objects.String{Value: fmt.Sprintf("%d", val.Value)}
				}
				return &objects.String{Value: fmt.Sprintf("%"+formatSpec, val.Value)}
			case *objects.Float:
				if formatSpec == "" {
					return &objects.String{Value: fmt.Sprintf("%g", val.Value)}
				}
				return &objects.String{Value: fmt.Sprintf("%"+formatSpec, val.Value)}
			case *objects.String:
				return val
			default:
				return &objects.String{Value: val.Inspect()}
			}
		},
	}
	formatNewIndex := len(c.constants)
	c.constants = append(c.constants, formatNewBuiltin)
	c.symbolTable.DefineBuiltin("format", formatNewIndex)
}

func NewWithState(s *SymbolTable, constants []objects.Object) *Compiler {
	return &Compiler{
		constants:   constants,
		symbolTable: s,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	if node == nil {
		return nil
	}
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
		case "%":
			c.emit(OpMod)
		case "//":
			c.emit(OpFloorDiv)
		case "**":
			c.emit(OpPower)
		case ">":
			c.emit(OpGreaterThan)
		case "==":
			c.emit(OpEqual)
		case "!=":
			c.emit(OpNotEqual)
		case "and", "or":
			return fmt.Errorf("and/or operators should be desugared before compilation")
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
		}

	case *ast.AwaitExpression:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(OpAwait)
	case *ast.NamedExpression:
		// Walrus 运算符: x := expr
		// 编译表达式并将结果赋值给变量
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		// Walrus 运算符需要：1. 赋值给变量，2. 返回赋值结果
		// 策略：复制栈顶值 -> 一份用于赋值，一份用于返回
		c.emit(OpDupTop) // 复制栈顶值，此时栈上有两个相同的值

		// 获取栈顶的值用于赋值
		symbol := c.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(OpSetGlobal, symbol.Index)
		} else {
			c.emit1(OpSetLocal, symbol.Index)
		}
		// 此时栈顶还有一个值（原始值的副本），作为表达式结果返回

	case *ast.IntegerLiteral:
		integer := &objects.Integer{Value: node.Value}
		c.emit(OpConstant, c.addConstant(integer))

	case *ast.FloatLiteral:
		float := &objects.Float{Value: node.Value}
		c.emit(OpConstant, c.addConstant(float))

	case *ast.ComplexLiteral:
		// Parse the complex literal string (e.g., "3.14j", "2j")
		s := node.Value
		s = strings.TrimRight(s, "jJ")
		imagPart, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("could not parse complex literal %q: %s", node.Value, err)
		}
		complexObj := objects.NewComplex(0, imagPart)
		c.emit(OpConstant, c.addConstant(complexObj))

	case *ast.StringLiteral:
		str := &objects.String{Value: node.Value}
		c.emit(OpConstant, c.addConstant(str))

	case *ast.ByteStringLiteral:
		bts := &objects.Bytes{Value: []byte(node.Value)}
		c.emit(OpConstant, c.addConstant(bts))

	case *ast.KeywordArgument:
		// For simplicity, we'll just compile the value - we'll handle keyword arguments
		// by creating a hash/dictionary to pass them
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

	case *ast.DictionaryUnpack:
		// 字典解包: **dict
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(OpDictUnpack)

	case *ast.ListUnpack:
		// 列表解包: *list
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(OpListUnpack)

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

	case *ast.EllipsisLiteral:
		c.emit(OpEllipsis)

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
		} else {
			c.emit(OpNull)
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
			} else {
				c.emit(OpNull)
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
		symbol := c.symbolTable.Define(node.Names[0].Value)
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
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		symbol, ok := c.symbolTable.Resolve(node.Names[0].Value)
		if !ok {
			symbol = c.symbolTable.Define(node.Names[0].Value)
		}

		if symbol.Scope == GlobalScope {
			c.emit(OpSetGlobal, symbol.Index)
		} else {
			c.emit1(OpSetLocal, symbol.Index)
		}

	case *ast.AttributeAssignStatement:
		err := c.Compile(node.Object)
		if err != nil {
			return err
		}

		err = c.Compile(node.Value)
		if err != nil {
			return err
		}

		c.emit(OpSetAttribute, c.addConstant(&objects.String{Value: node.Attr.Value}))
		c.emit(OpPop)

	case *ast.DeleteStatement:
		for _, target := range node.Targets {
			switch t := target.(type) {
			case *ast.MemberAccess:
				err := c.Compile(t.Object)
				if err != nil {
					return err
				}
				c.emit(OpDelAttribute, c.addConstant(&objects.String{Value: t.Member.Value}))
			case *ast.IndexExpression:
				err := c.Compile(t.Left)
				if err != nil {
					return err
				}
				err = c.Compile(t.Index)
				if err != nil {
					return err
				}
				c.emit(OpPop) // pop index
				c.emit(OpPop) // pop object (simplified: no OpDelIndex yet)
			case *ast.Identifier:
				symbol, ok := c.symbolTable.Resolve(t.Value)
				if !ok {
					return fmt.Errorf("undefined variable %s", t.Value)
				}
				if symbol.Scope == GlobalScope {
					c.emit(OpNull)
					c.emit(OpSetGlobal, symbol.Index)
				} else {
					c.emit(OpNull)
					c.emit1(OpSetLocal, symbol.Index)
				}
			}
		}

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			if node.Value == "True" {
			c.emit(OpTrue)
		} else if node.Value == "False" {
			c.emit(OpFalse)
		} else if node.Value == "None" {
			c.emit(OpNull)
		} else if node.Value == "Ellipsis" {
			c.emit(OpEllipsis)
		} else {
				return fmt.Errorf("undefined variable %s", node.Value)
			}
		} else {
			if symbol.Scope == BuiltinScope {
				c.emit(OpConstant, symbol.Index)
			} else if symbol.Scope == GlobalScope || symbol.Scope == FunctionScope {
				c.emit(OpGetGlobal, symbol.Index)
			} else if symbol.Scope == FreeScope {
				c.emit1(OpGetFree, symbol.Index)
			} else {
				c.emit1(OpGetLocal, symbol.Index)
			}
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

	case *ast.AsyncListComprehension:
		return c.compileAsyncListComprehension(node)

	case *ast.AsyncSetComprehension:
		return c.compileAsyncSetComprehension(node)

	case *ast.AsyncDictComprehension:
		return c.compileAsyncDictComprehension(node)

	case *ast.AsyncGeneratorExpression:
		return c.compileAsyncGeneratorExpression(node)

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
		if slice, ok := node.Index.(*ast.SliceExpression); ok {
			// 处理切片表达式
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}

			// 编译 lower
			if slice.Lower == nil {
				c.emit(OpConstant, c.addConstant(&objects.Integer{Value: 0}))
			} else {
				err = c.Compile(slice.Lower)
				if err != nil {
					return err
				}
			}

			// 编译 upper
			if slice.Upper == nil {
				c.emit(OpConstant, c.addConstant(&objects.Integer{Value: -1}))
			} else {
				err = c.Compile(slice.Upper)
				if err != nil {
					return err
				}
			}

			// 编译 step
			if slice.Step == nil {
				// 默认步长为 None，VM 会根据上下文决定是 1 还是 -1
				c.emit(OpNull)
			} else {
				err = c.Compile(slice.Step)
				if err != nil {
					return err
				}
			}

			c.emit(OpSlice)
		} else {
			// 正常索引
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}

			err = c.Compile(node.Index)
			if err != nil {
				return err
			}
			c.emit(OpIndex)
		}
	case *ast.SliceExpression:
		// SliceExpression 现在是 IndexExpression.Index，需要单独处理
		// 但实际上，解析器现在返回的是 IndexExpression{Index: SliceExpression}
		// 所以这个 case 不会被直接触发
		return nil

	case *ast.FunctionLiteral:
		outerInstructions := c.instructions
		outerLastInstruction := c.lastInstruction
		outerPreviousInstruction := c.previousInstruction
		c.instructions = make(Instructions, 0)
		c.lastInstruction = EmittedInstruction{}
		c.previousInstruction = EmittedInstruction{}

		c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		// 处理 global 和 nonlocal 声明
		// 需要在编译 body 之前处理这些声明，因为它们会影响变量的作用域解析
		blockStmt := node.Body
		for _, stmt := range blockStmt.Statements {
			switch s := stmt.(type) {
			case *ast.GlobalStatement:
				for _, name := range s.Names {
					c.symbolTable.DefineGlobal(name.Value)
				}
			case *ast.NonlocalStatement:
				for _, name := range s.Names {
					c.symbolTable.DefineNonlocal(name.Value)
				}
			}
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
		freeSymbols := c.symbolTable.FreeSymbols
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

		numKeywordOnly := 0
		for _, kw := range node.KeywordOnly {
			if kw {
				numKeywordOnly++
			}
		}

		numDefaults := 0
		numPositionalDefaults := 0
		for i, d := range node.Defaults {
			if d != nil {
				numDefaults++
				if !node.KeywordOnly[i] {
					numPositionalDefaults++
				}
			}
		}

		paramNames := make([]string, len(node.Parameters))
		for i, p := range node.Parameters {
			paramNames[i] = p.Value
		}

		numPositionalOnly := 0
		for _, po := range node.PositionalOnly {
			if po {
				numPositionalOnly++
			}
		}

		compiledFn := &CompiledFunction{
			Instructions:          fnInstructions,
			NumLocals:             numLocals,
			NumParameters:         len(node.Parameters),
			NumKeywordOnly:        numKeywordOnly,
			NumPositionalOnly:     numPositionalOnly,
			NumDefaults:           numDefaults,
			NumPositionalDefaults: numPositionalDefaults,
			ParameterNames:        paramNames,
			PositionalOnly:        node.PositionalOnly,
			IsGenerator:           c.hasYieldInBody(node.Body),
			IsAsync:               node.IsAsync,
			Free:                  allFreeVars,
			VarArgs:               node.VarArgs != nil,
			KwArgs:                node.KwArgs != nil,
		}

		c.instructions = make(Instructions, 0, len(outerInstructions))
		c.instructions = append(c.instructions, outerInstructions...)
		c.lastInstruction = outerLastInstruction
		c.previousInstruction = outerPreviousInstruction

		// Determine if this function needs to be a closure
		// It needs to be a closure if:
		// 1. It has its own free variables (references outer scope), OR
		// 2. Nested functions reference this function's locals
		needsClosure := numFree > 0 || len(nestedFreeSymbols) > 0

		if needsClosure {
			// Emit instructions to load free variables onto stack
			// First, load this function's own free variables
			// Use FreeSymbols (original scope info) to determine the correct opcode
			for _, freeSym := range freeSymbols {
				if freeSym.Scope == GlobalScope {
					c.emit(OpGetGlobal, freeSym.Index)
				} else if freeSym.Scope == FreeScope {
					c.emit1(OpGetFree, freeSym.Index)
				} else {
					c.emit1(OpGetLocal, freeSym.Index)
				}
			}
			// Then, load the nested free variables (locals that nested functions need)
			for _, nestedFree := range nestedFreeSymbols {
				if nestedFree.Scope == FreeScope {
					c.emit1(OpGetFree, nestedFree.Index)
				} else {
					c.emit1(OpGetLocal, nestedFree.Index)
				}
			}
			totalFree := numFree + len(nestedFreeSymbols)
			c.emitClosure(c.addConstant(compiledFn), totalFree)
		} else {
			c.emit(OpConstant, c.addConstant(compiledFn))
		}
		if compiledFn.IsGenerator {
			c.emit(OpMakeGenerator)
		}
		if compiledFn.IsAsync {
			c.emit(OpMakeAsync)
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

		// 分离位置参数和关键字参数
		posArgs := []ast.Expression{}
		kwargsPairs := []ast.Expression{} // key1, value1, key2, value2...

		for _, a := range node.Arguments {
			if keywordArg, ok := a.(*ast.KeywordArgument); ok {
				// 处理关键字参数: 编译字符串作为key，然后是value
				keyStr := &ast.StringLiteral{Token: keywordArg.Token, Value: keywordArg.Name.Value}
				kwargsPairs = append(kwargsPairs, keyStr, keywordArg.Value)
			} else {
				posArgs = append(posArgs, a)
			}
		}

		// 编译位置参数
		for _, arg := range posArgs {
			if err := c.Compile(arg); err != nil {
				return err
			}
		}

		// 如果有关键字参数，编译成字典
		numKwargs := len(kwargsPairs)
		if numKwargs > 0 {
			for _, pair := range kwargsPairs {
				if err := c.Compile(pair); err != nil {
					return err
				}
			}
			c.emit(OpHash, numKwargs)
		}

		// 总参数数量 = 位置参数 + (如果有kwargs则+1)
		totalArgs := len(posArgs)
		if numKwargs > 0 {
			totalArgs += 1
		}
		c.emit1(OpCall, totalArgs)

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
		// 处理单个上下文管理器（因为多重的已经被desugar成嵌套with了）
		if len(node.Items) > 0 {
			item := node.Items[0]
			if err := c.Compile(item.Expr); err != nil {
				return err
			}

			c.emit(OpEnterContext)

			if item.Name != nil {
				symbol := c.symbolTable.Define(item.Name.Value)
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
		}
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
	case *ast.GlobalStatement:
		// global 语句只是声明，不需要编译任何代码
		return nil
	case *ast.NonlocalStatement:
		// nonlocal 语句只是声明，不需要编译任何代码
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
	case *ast.MatchStatement:
		return c.compileMatchStatement(node)
	case *ast.CaseClause:
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

	beginTryPos := c.emit(OpBeginTry, len(ts.Excepts), boolToInt(hasFinally), 0, 0)

	if err := c.Compile(ts.Body); err != nil {
		return err
	}

	jumpPositions := []int{}
	jumpPositions = append(jumpPositions, c.emit(OpJump, 0))

	firstHandlerIP := len(c.instructions)

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

			if ex.IsStar {
				c.emit(OpExceptStarHandler, typeIdx, varIdx)
			} else {
				c.emit(OpExceptHandler, typeIdx, varIdx)
			}

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

			jumpPositions = append(jumpPositions, c.emit(OpJump, 0))
		}
	}

	if hasExcept {
		c.changeOperand(beginTryPos+4, firstHandlerIP)
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
		c.changeOperand(beginTryPos+6, finallyStartPos)
	} else {
		afterTryPos := len(c.instructions)
		for _, pos := range jumpPositions {
			c.changeOperand(pos, afterTryPos)
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

func (c *Compiler) compileMatchStatement(node *ast.MatchStatement) error {
	return fmt.Errorf("match statement should be desugared before compilation")
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
			var methodObj objects.Object = compiledFn
			for _, dec := range method.Decorators {
				if ident, ok := dec.(*ast.Identifier); ok {
					if ident.Value == "staticmethod" {
						methodObj = &objects.StaticMethod{Fn: compiledFn}
					} else if ident.Value == "classmethod" {
						methodObj = &objects.ClassMethod{Fn: compiledFn}
					} else if ident.Value == "property" {
						methodObj = &objects.Property{Fget: compiledFn}
					}
				}
				if ma, ok := dec.(*ast.MemberAccess); ok {
					if ident, ok := ma.Object.(*ast.Identifier); ok {
						if ident.Value == method.Name {
							prop := &objects.Property{Fget: objects.None_}
							if existing, ok := class.Methods[method.Name]; ok {
								if existingProp, ok := existing.(*objects.Property); ok {
									prop = existingProp
								}
							}
							if ma.Member.Value == "setter" {
								prop.Fset = compiledFn
							} else if ma.Member.Value == "deleter" {
								prop.Fdel = compiledFn
							}
							methodObj = prop
						}
					}
				}
			}
			class.Methods[method.Name] = methodObj
		}
	}

	if len(node.SuperClasses) > 1 {
		for _, sc := range node.SuperClasses {
			superClassIdx, ok := c.symbolTable.Resolve(sc.Value)
			if ok && superClassIdx.Scope == GlobalScope {
				c.emit(OpGetGlobal, superClassIdx.Index)
			}
		}
		numParents := len(node.SuperClasses)
		c.emit(OpCreateClassWithMultiSuper, c.addConstant(class))
		c.emit(Opcode(numParents))
	} else if node.SuperClass != nil {
		superClassIdx, ok := c.symbolTable.Resolve(node.SuperClass.Value)
		if ok && superClassIdx.Scope == GlobalScope {
			c.emit(OpGetGlobal, superClassIdx.Index)
		}
		c.emit(OpCreateClassWithSuper, c.addConstant(class))
	} else {
		c.emit(OpCreateClass, c.addConstant(class))
	}

	// Set metaclass if specified
	if node.Metaclass != nil {
		metaclassIdx, ok := c.symbolTable.Resolve(node.Metaclass.Value)
		if ok && metaclassIdx.Scope == GlobalScope {
			c.emit(OpGetGlobal, metaclassIdx.Index)
			c.emit(OpSetMetaclass)
			c.emit(OpCallMetaclassInit)
		}
	}

	symbol := c.symbolTable.Define(node.Name.Value)
	c.emit(OpSetGlobal, symbol.Index)

	for _, method := range node.Methods {
		if compiledFn, ok := class.Methods[method.Name]; ok {
			idx := c.addConstant(compiledFn)
			c.symbolTable.DefineBuiltin(method.Name, idx)
		}
	}

	for _, stmt := range node.Body.Statements {
		if stmt == nil {
			continue
		}
		if assign, ok := stmt.(*ast.AssignStatement); ok && len(assign.Names) == 1 {
			c.emit(OpGetGlobal, symbol.Index)
			err := c.Compile(assign.Value)
			if err != nil {
				return err
			}
			c.emit(OpSetClassField, c.addConstant(&objects.String{Value: assign.Names[0].Value}))
		}
	}

	for _, method := range node.Methods {
		delete(c.symbolTable.store, method.Name)
	}
	
	return nil
}

func (c *Compiler) compileFunction(fn *ast.FunctionLiteral) *CompiledFunction {
	savedInstructions := c.instructions
	savedLastInstruction := c.lastInstruction
	savedPreviousInstruction := c.previousInstruction
	c.instructions = []byte{}
	c.lastInstruction = EmittedInstruction{}
	c.previousInstruction = EmittedInstruction{}

	c.enterScope()

	for _, param := range fn.Parameters {
		c.symbolTable.Define(param.Value)
	}

	for _, stmt := range fn.Body.Statements {
		if err := c.Compile(stmt); err != nil {
			c.exitScope()
			c.instructions = savedInstructions
			c.lastInstruction = savedLastInstruction
			c.previousInstruction = savedPreviousInstruction
			return nil
		}
	}

	if !c.lastInstructionIs(OpReturnValue) && !c.lastInstructionIs(OpReturn) {
		c.emit(OpNull)
		c.emit(OpReturnValue)
	}

	numLocals := c.symbolTable.numDefinitions
	free := c.symbolTable.Free
	c.exitScope()

	fnInstructions := c.instructions

	c.instructions = savedInstructions
	c.lastInstruction = savedLastInstruction
	c.previousInstruction = savedPreviousInstruction

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

func isTruthy(obj objects.Object) bool {
	switch o := obj.(type) {
	case *objects.Boolean:
		return o.Value
	case *objects.Integer:
		return o.Value != 0
	case *objects.Float:
		return o.Value != 0.0
	case *objects.String:
		return o.Value != ""
	case *objects.None:
		return false
	case *objects.List:
		return len(o.Elements) > 0
	case *objects.Tuple:
		return len(o.Elements) > 0
	case *objects.Dict:
		return len(o.Pairs) > 0
	case *objects.Set:
		return o.Size() > 0
	default:
		return true
	}
}

func compareObjects(a, b objects.Object) int {
	switch a := a.(type) {
	case *objects.Integer:
		if bInt, ok := b.(*objects.Integer); ok {
			if a.Value < bInt.Value {
				return -1
			}
			if a.Value > bInt.Value {
				return 1
			}
			return 0
		}
	case *objects.Float:
		if bFloat, ok := b.(*objects.Float); ok {
			if a.Value < bFloat.Value {
				return -1
			}
			if a.Value > bFloat.Value {
				return 1
			}
			return 0
		}
	case *objects.String:
		if bStr, ok := b.(*objects.String); ok {
			if a.Value < bStr.Value {
				return -1
			}
			if a.Value > bStr.Value {
				return 1
			}
			return 0
		}
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

func (c *Compiler) compileAsyncListComprehension(node *ast.AsyncListComprehension) error {
	// Same pattern as compileListComprehension but with IsAsync: true
	outerInstructions := c.instructions
	c.instructions = make(Instructions, 0)
	c.lastInstruction = EmittedInstruction{}
	c.previousInstruction = EmittedInstruction{}

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

	fnInstructions := c.instructions
	numLocals := c.symbolTable.numDefinitions
	freeVars := c.symbolTable.Free
	freeSymbols := c.symbolTable.FreeSymbols
	nestedFreeSymbols := c.symbolTable.NestedFreeSymbols

	c.instructions = outerInstructions

	allFreeVars := make([]Symbol, 0, len(freeVars)+len(nestedFreeSymbols))
	allFreeVars = append(allFreeVars, freeVars...)
	for _, nfs := range nestedFreeSymbols {
		allFreeVars = append(allFreeVars, Symbol{
			Name:  nfs.Name,
			Scope: FreeScope,
			Index: len(allFreeVars),
		})
	}

	compiledFn := &CompiledFunction{
		Instructions:  fnInstructions,
		NumLocals:     numLocals,
		NumParameters: 0,
		IsAsync:       true,
		Free:          allFreeVars,
	}

	needsClosure := len(freeVars) > 0 || len(nestedFreeSymbols) > 0
	if needsClosure {
		for _, freeSym := range freeSymbols {
			if freeSym.Scope == GlobalScope {
				c.emit(OpGetGlobal, freeSym.Index)
			} else if freeSym.Scope == FreeScope {
				c.emit1(OpGetFree, freeSym.Index)
			} else {
				c.emit1(OpGetLocal, freeSym.Index)
			}
		}
		for _, nestedFree := range nestedFreeSymbols {
			if nestedFree.Scope == FreeScope {
				c.emit1(OpGetFree, nestedFree.Index)
			} else {
				c.emit1(OpGetLocal, nestedFree.Index)
			}
		}
		totalFree := len(freeVars) + len(nestedFreeSymbols)
		c.emitClosure(c.addConstant(compiledFn), totalFree)
	} else {
		c.emit(OpConstant, c.addConstant(compiledFn))
	}

	c.emit(OpMakeAsync)

	return nil
}

func (c *Compiler) compileAsyncSetComprehension(node *ast.AsyncSetComprehension) error {
	// Same pattern as compileSetComprehension (which delegates to compileListComprehension) but with IsAsync: true
	return c.compileAsyncListComprehension(&ast.AsyncListComprehension{
		Token:    node.Token,
		Element:  node.Element,
		Variable: node.Variable,
		Iterable: node.Iterable,
		Filter:   node.Filter,
	})
}

func (c *Compiler) compileAsyncDictComprehension(node *ast.AsyncDictComprehension) error {
	// Same pattern as compileDictComprehension but with IsAsync: true
	outerInstructions := c.instructions
	c.instructions = make(Instructions, 0)
	c.lastInstruction = EmittedInstruction{}
	c.previousInstruction = EmittedInstruction{}

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

	fnInstructions := c.instructions
	numLocals := c.symbolTable.numDefinitions
	freeVars := c.symbolTable.Free
	freeSymbols := c.symbolTable.FreeSymbols
	nestedFreeSymbols := c.symbolTable.NestedFreeSymbols

	c.instructions = outerInstructions

	allFreeVars := make([]Symbol, 0, len(freeVars)+len(nestedFreeSymbols))
	allFreeVars = append(allFreeVars, freeVars...)
	for _, nfs := range nestedFreeSymbols {
		allFreeVars = append(allFreeVars, Symbol{
			Name:  nfs.Name,
			Scope: FreeScope,
			Index: len(allFreeVars),
		})
	}

	compiledFn := &CompiledFunction{
		Instructions:  fnInstructions,
		NumLocals:     numLocals,
		NumParameters: 0,
		IsAsync:       true,
		Free:          allFreeVars,
	}

	needsClosure := len(freeVars) > 0 || len(nestedFreeSymbols) > 0
	if needsClosure {
		for _, freeSym := range freeSymbols {
			if freeSym.Scope == GlobalScope {
				c.emit(OpGetGlobal, freeSym.Index)
			} else if freeSym.Scope == FreeScope {
				c.emit1(OpGetFree, freeSym.Index)
			} else {
				c.emit1(OpGetLocal, freeSym.Index)
			}
		}
		for _, nestedFree := range nestedFreeSymbols {
			if nestedFree.Scope == FreeScope {
				c.emit1(OpGetFree, nestedFree.Index)
			} else {
				c.emit1(OpGetLocal, nestedFree.Index)
			}
		}
		totalFree := len(freeVars) + len(nestedFreeSymbols)
		c.emitClosure(c.addConstant(compiledFn), totalFree)
	} else {
		c.emit(OpConstant, c.addConstant(compiledFn))
	}

	c.emit(OpMakeAsync)

	return nil
}

func (c *Compiler) compileAsyncGeneratorExpression(node *ast.AsyncGeneratorExpression) error {
	// Same pattern as compileListComprehension but with IsAsync: true
	outerInstructions := c.instructions
	c.instructions = make(Instructions, 0)
	c.lastInstruction = EmittedInstruction{}
	c.previousInstruction = EmittedInstruction{}

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

	fnInstructions := c.instructions
	numLocals := c.symbolTable.numDefinitions
	freeVars := c.symbolTable.Free
	freeSymbols := c.symbolTable.FreeSymbols
	nestedFreeSymbols := c.symbolTable.NestedFreeSymbols

	c.instructions = outerInstructions

	allFreeVars := make([]Symbol, 0, len(freeVars)+len(nestedFreeSymbols))
	allFreeVars = append(allFreeVars, freeVars...)
	for _, nfs := range nestedFreeSymbols {
		allFreeVars = append(allFreeVars, Symbol{
			Name:  nfs.Name,
			Scope: FreeScope,
			Index: len(allFreeVars),
		})
	}

	compiledFn := &CompiledFunction{
		Instructions:  fnInstructions,
		NumLocals:     numLocals,
		NumParameters: 0,
		IsAsync:       true,
		Free:          allFreeVars,
	}

	needsClosure := len(freeVars) > 0 || len(nestedFreeSymbols) > 0
	if needsClosure {
		for _, freeSym := range freeSymbols {
			if freeSym.Scope == GlobalScope {
				c.emit(OpGetGlobal, freeSym.Index)
			} else if freeSym.Scope == FreeScope {
				c.emit1(OpGetFree, freeSym.Index)
			} else {
				c.emit1(OpGetLocal, freeSym.Index)
			}
		}
		for _, nestedFree := range nestedFreeSymbols {
			if nestedFree.Scope == FreeScope {
				c.emit1(OpGetFree, nestedFree.Index)
			} else {
				c.emit1(OpGetLocal, nestedFree.Index)
			}
		}
		totalFree := len(freeVars) + len(nestedFreeSymbols)
		c.emitClosure(c.addConstant(compiledFn), totalFree)
	} else {
		c.emit(OpConstant, c.addConstant(compiledFn))
	}

	c.emit(OpMakeAsync)

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