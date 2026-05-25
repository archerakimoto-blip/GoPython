package objects

import (
	"fmt"
	"math"
)

type ObjectType string

const (
	INTEGER_OBJ      ObjectType = "INTEGER"
	FLOAT_OBJ        ObjectType = "FLOAT"
	BOOLEAN_OBJ      ObjectType = "BOOLEAN"
	STRING_OBJ       ObjectType = "STRING"
	NONE_OBJ         ObjectType = "NONE"
	LIST_OBJ         ObjectType = "LIST"
	SET_OBJ          ObjectType = "SET"
	DICT_OBJ         ObjectType = "DICT"
	FUNCTION_OBJ     ObjectType = "FUNCTION"
	BUILTIN_OBJ      ObjectType = "BUILTIN"
	ERROR_OBJ        ObjectType = "ERROR"
	GENERATOR_OBJ    ObjectType = "GENERATOR"
	CONTEXT_OBJ      ObjectType = "CONTEXT"
	CLASS_OBJ        ObjectType = "CLASS"
	INSTANCE_OBJ     ObjectType = "INSTANCE"
	MODULE_OBJ       ObjectType = "MODULE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type None struct{}

func (n *None) Type() ObjectType { return NONE_OBJ }
func (n *None) Inspect() string  { return "None" }

type List struct {
	Elements []Object
}

func (l *List) Type() ObjectType { return LIST_OBJ }
func (l *List) Inspect() string {
	result := "["
	for i, el := range l.Elements {
		if i > 0 {
			result += ", "
		}
		result += el.Inspect()
	}
	result += "]"
	return result
}

type Set struct {
	Elements []Object
}

func (s *Set) Type() ObjectType { return SET_OBJ }
func (s *Set) Inspect() string {
	result := "{"
	for i, el := range s.Elements {
		if i > 0 {
			result += ", "
		}
		result += el.Inspect()
	}
	result += "}"
	return result
}

type Dict struct {
	Pairs map[Object]Object
}

func (d *Dict) Type() ObjectType { return DICT_OBJ }
func (d *Dict) Inspect() string {
	result := "{"
	first := true
	for k, v := range d.Pairs {
		if !first {
			result += ", "
		}
		first = false
		result += fmt.Sprintf("%s: %s", k.Inspect(), v.Inspect())
	}
	result += "}"
	return result
}

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { 
	if b.Name != "" {
		return "builtin function: " + b.Name
	}
	return "builtin function" 
}

type ContextManager struct {
	EnterFunc func() Object
	ExitFunc  func(exc Object) Object
}

func (cm *ContextManager) Type() ObjectType { return CONTEXT_OBJ }
func (cm *ContextManager) Inspect() string  { return "context manager" }

type Generator struct {
	Instructions  []byte
	Constants     []Object
	Locals        []Object
	IP            int
	Stack         []Object
	StackPtr      int
	BasePointer   int
	Done          bool
}

func (g *Generator) Type() ObjectType { return GENERATOR_OBJ }
func (g *Generator) Inspect() string  { return fmt.Sprintf("generator[%p]", g) }

type Class struct {
	Name       string
	Methods    map[string]Object
	Fields     map[string]Object
	SuperClass *Class
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string  { return fmt.Sprintf("<class %s>", c.Name) }

type Instance struct {
	Class  *Class
	Fields map[string]Object
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string  { return fmt.Sprintf("<%s instance>", i.Class.Name) }

func (i *Instance) GetAttr(name string) (Object, bool) {
	if val, ok := i.Fields[name]; ok {
		return val, true
	}
	if i.Class != nil {
		if method, ok := i.Class.Methods[name]; ok {
			return method, true
		}
		if i.Class.SuperClass != nil {
			if method, ok := i.Class.SuperClass.Methods[name]; ok {
				return method, true
			}
		}
	}
	return nil, false
}

func (i *Instance) SetAttr(name string, value Object) {
	i.Fields[name] = value
}

type Module struct {
	Name   string
	Fields map[string]Object
}

func (m *Module) Type() ObjectType { return MODULE_OBJ }
func (m *Module) Inspect() string  { return fmt.Sprintf("<module '%s'>", m.Name) }

func (m *Module) GetAttr(name string) (Object, bool) {
	if val, ok := m.Fields[name]; ok {
		return val, true
	}
	return nil, false
}

var (
	True  = &Boolean{Value: true}
	False = &Boolean{Value: false}
	None_ = &None{}
)

func NewError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func Equal(a, b Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch a := a.(type) {
	case *Integer:
		b := b.(*Integer)
		return a.Value == b.Value
	case *Float:
		b := b.(*Float)
		return a.Value == b.Value
	case *String:
		b := b.(*String)
		return a.Value == b.Value
	case *Boolean:
		b := b.(*Boolean)
		return a.Value == b.Value
	default:
		return false
	}
}

func FormatString(template string, args ...Object) string {
	result := template
	argIndex := 0
	
	for {
		idx := -1
		for i := 0; i < len(result); i++ {
			if result[i] == '%' {
				if i+1 < len(result) {
					next := result[i+1]
					if next == 's' || next == 'd' || next == 'f' || next == 'g' {
						idx = i
						break
					}
				}
			}
		}
		
		if idx == -1 || argIndex >= len(args) {
			break
		}
		
		formatChar := result[idx+1]
		var replacement string
		
		if argIndex < len(args) {
			arg := args[argIndex]
			switch formatChar {
			case 's':
				replacement = arg.Inspect()
			case 'd':
				if intObj, ok := arg.(*Integer); ok {
					replacement = fmt.Sprintf("%d", intObj.Value)
				} else if floatObj, ok := arg.(*Float); ok {
					replacement = fmt.Sprintf("%d", int64(floatObj.Value))
				} else {
					replacement = arg.Inspect()
				}
			case 'f', 'g':
				if floatObj, ok := arg.(*Float); ok {
					if formatChar == 'f' {
						replacement = fmt.Sprintf("%f", floatObj.Value)
					} else {
						replacement = fmt.Sprintf("%g", floatObj.Value)
					}
				} else if intObj, ok := arg.(*Integer); ok {
					if formatChar == 'f' {
						replacement = fmt.Sprintf("%f", float64(intObj.Value))
					} else {
						replacement = fmt.Sprintf("%g", float64(intObj.Value))
					}
				} else {
					replacement = arg.Inspect()
				}
			}
			argIndex++
		}
		
		result = result[:idx] + replacement + result[idx+2:]
	}
	
	return result
}

func CreateMathModule() *Module {
	mathModule := &Module{
		Name:   "math",
		Fields: make(map[string]Object),
	}
	
	mathModule.Fields["pi"] = &Float{Value: math.Pi}
	mathModule.Fields["e"] = &Float{Value: math.E}
	
	mathModule.Fields["sin"] = &Builtin{
		Name: "math.sin",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("sin() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Sin(v.Value)}
			case *Integer:
				return &Float{Value: math.Sin(float64(v.Value))}
			default:
				return NewError("sin() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["cos"] = &Builtin{
		Name: "math.cos",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("cos() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Cos(v.Value)}
			case *Integer:
				return &Float{Value: math.Cos(float64(v.Value))}
			default:
				return NewError("cos() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["sqrt"] = &Builtin{
		Name: "math.sqrt",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("sqrt() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Sqrt(v.Value)}
			case *Integer:
				return &Float{Value: math.Sqrt(float64(v.Value))}
			default:
				return NewError("sqrt() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["floor"] = &Builtin{
		Name: "math.floor",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("floor() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Integer{Value: int64(math.Floor(v.Value))}
			case *Integer:
				return v
			default:
				return NewError("floor() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["ceil"] = &Builtin{
		Name: "math.ceil",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("ceil() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Integer{Value: int64(math.Ceil(v.Value))}
			case *Integer:
				return v
			default:
				return NewError("ceil() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["abs"] = &Builtin{
		Name: "math.abs",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("abs() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Abs(v.Value)}
			case *Integer:
				if v.Value < 0 {
					return &Integer{Value: -v.Value}
				}
				return v
			default:
				return NewError("abs() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["pow"] = &Builtin{
		Name: "math.pow",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return NewError("pow() takes exactly 2 arguments")
			}
			base, ok1 := args[0].(*Float)
			exp, ok2 := args[1].(*Float)
			if !ok1 || !ok2 {
				return NewError("pow() arguments must be numbers")
			}
			return &Float{Value: math.Pow(base.Value, exp.Value)}
		},
	}
	
	mathModule.Fields["log"] = &Builtin{
		Name: "math.log",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("log() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Log(v.Value)}
			case *Integer:
				return &Float{Value: math.Log(float64(v.Value))}
			default:
				return NewError("log() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["exp"] = &Builtin{
		Name: "math.exp",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("exp() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Exp(v.Value)}
			case *Integer:
				return &Float{Value: math.Exp(float64(v.Value))}
			default:
				return NewError("exp() argument must be a number")
			}
		},
	}
	
	return mathModule
}
