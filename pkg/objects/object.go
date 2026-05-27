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

func NewList(elements []Object) *List {
	return &List{Elements: elements}
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

func (l *List) Append(obj Object) {
	l.Elements = append(l.Elements, obj)
}

func (l *List) Extend(other *List) {
	l.Elements = append(l.Elements, other.Elements...)
}

func (l *List) Pop(index ...int) (Object, error) {
	idx := len(l.Elements) - 1
	if len(index) > 0 {
		idx = index[0]
		if idx < 0 {
			idx = len(l.Elements) + idx
		}
	}
	
	if idx < 0 || idx >= len(l.Elements) {
		return nil, fmt.Errorf("pop index out of range")
	}
	
	obj := l.Elements[idx]
	l.Elements = append(l.Elements[:idx], l.Elements[idx+1:]...)
	return obj, nil
}

func (l *List) Index(obj Object) int {
	for i, el := range l.Elements {
		if Equal(el, obj) {
			return i
		}
	}
	return -1
}

func (l *List) Contains(obj Object) bool {
	return l.Index(obj) != -1
}

func (l *List) Insert(index int, obj Object) {
	if index < 0 {
		index = len(l.Elements) + index
	}
	if index < 0 {
		index = 0
	}
	if index > len(l.Elements) {
		index = len(l.Elements)
	}
	
	l.Elements = append(l.Elements[:index], append([]Object{obj}, l.Elements[index:]...)...)
}

func (l *List) Remove(obj Object) error {
	idx := l.Index(obj)
	if idx == -1 {
		return fmt.Errorf("value not in list")
	}
	_, _ = l.Pop(idx)
	return nil
}

func (l *List) Reverse() {
	for i, j := 0, len(l.Elements)-1; i < j; i, j = i+1, j-1 {
		l.Elements[i], l.Elements[j] = l.Elements[j], l.Elements[i]
	}
}

func (l *List) Size() int {
	return len(l.Elements)
}

func (l *List) Clear() {
	l.Elements = []Object{}
}

type Set struct {
	Elements map[string]Object
	Keys     map[string]Object
}

func NewSet() *Set {
	return &Set{
		Elements: make(map[string]Object),
		Keys:     make(map[string]Object),
	}
}

func (s *Set) Type() ObjectType { return SET_OBJ }
func (s *Set) Inspect() string {
	result := "{"
	first := true
	for _, el := range s.Elements {
		if !first {
			result += ", "
		}
		first = false
		result += el.Inspect()
	}
	result += "}"
	return result
}

func (s *Set) HashKey(obj Object) string {
	switch o := obj.(type) {
	case *Integer:
		return fmt.Sprintf("int:%d", o.Value)
	case *Float:
		return fmt.Sprintf("float:%g", o.Value)
	case *Boolean:
		return fmt.Sprintf("bool:%t", o.Value)
	case *String:
		return fmt.Sprintf("str:%s", o.Value)
	default:
		return fmt.Sprintf("obj:%p", obj)
	}
}

func (s *Set) Add(obj Object) {
	key := s.HashKey(obj)
	s.Elements[key] = obj
	s.Keys[key] = obj
}

func (s *Set) Contains(obj Object) bool {
	key := s.HashKey(obj)
	_, ok := s.Elements[key]
	return ok
}

func (s *Set) Remove(obj Object) {
	key := s.HashKey(obj)
	delete(s.Elements, key)
	delete(s.Keys, key)
}

func (s *Set) Size() int {
	return len(s.Elements)
}

func (s *Set) ToSlice() []Object {
	slice := make([]Object, 0, len(s.Elements))
	for _, v := range s.Elements {
		slice = append(slice, v)
	}
	return slice
}

type Dict struct {
	Pairs map[string]Object
	Keys  map[string]Object
}

func NewDict() *Dict {
	return &Dict{
		Pairs: make(map[string]Object),
		Keys:  make(map[string]Object),
	}
}

func (d *Dict) Type() ObjectType { return DICT_OBJ }
func (d *Dict) Inspect() string {
	result := "{"
	first := true
	for keyStr, key := range d.Keys {
		if !first {
			result += ", "
		}
		first = false
		value := d.Pairs[keyStr]
		result += fmt.Sprintf("%s: %s", key.Inspect(), value.Inspect())
	}
	result += "}"
	return result
}

func (d *Dict) HashKey(obj Object) string {
	switch o := obj.(type) {
	case *Integer:
		return fmt.Sprintf("int:%d", o.Value)
	case *Float:
		return fmt.Sprintf("float:%g", o.Value)
	case *Boolean:
		return fmt.Sprintf("bool:%t", o.Value)
	case *String:
		return fmt.Sprintf("str:%s", o.Value)
	default:
		return fmt.Sprintf("obj:%p", obj)
	}
}

func (d *Dict) Get(key Object) (Object, bool) {
	keyStr := d.HashKey(key)
	value, ok := d.Pairs[keyStr]
	return value, ok
}

func (d *Dict) Set(key, value Object) {
	keyStr := d.HashKey(key)
	d.Pairs[keyStr] = value
	d.Keys[keyStr] = key
}

func (d *Dict) Has(key Object) bool {
	keyStr := d.HashKey(key)
	_, ok := d.Pairs[keyStr]
	return ok
}

func (d *Dict) Delete(key Object) {
	keyStr := d.HashKey(key)
	delete(d.Pairs, keyStr)
	delete(d.Keys, keyStr)
}

func (d *Dict) Size() int {
	return len(d.Pairs)
}

func (d *Dict) KeysSlice() []Object {
	keys := make([]Object, 0, len(d.Keys))
	for _, key := range d.Keys {
		keys = append(keys, key)
	}
	return keys
}

func (d *Dict) ValuesSlice() []Object {
	values := make([]Object, 0, len(d.Pairs))
	for _, value := range d.Pairs {
		values = append(values, value)
	}
	return values
}

type Error struct {
	ErrorType string
	Message   string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string { return e.ErrorType + ": " + e.Message }

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

type Closure struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
	IsGenerator   bool
	Free          []Object
}

func (c *Closure) Type() ObjectType { return FUNCTION_OBJ }
func (c *Closure) Inspect() string  { return "closure" }

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

var modules = make(map[string]*Module)

func RegisterModule(name string, module *Module) {
	modules[name] = module
}

func GetModule(name string) *Module {
	return modules[name]
}

var (
	True  = &Boolean{Value: true}
	False = &Boolean{Value: false}
	None_ = &None{}
)

func newErrorWithType(errorType, format string, a ...interface{}) *Error {
	return &Error{
		ErrorType: errorType,
		Message:   fmt.Sprintf(format, a...),
	}
}

func NewError(format string, a ...interface{}) *Error {
	return &Error{
		ErrorType: "Error",
		Message:   fmt.Sprintf(format, a...),
	}
}

func NewException(format string, a ...interface{}) *Error {
	return newErrorWithType("Exception", format, a...)
}

func NewValueError(format string, a ...interface{}) *Error {
	return newErrorWithType("ValueError", format, a...)
}

func NewTypeError(format string, a ...interface{}) *Error {
	return newErrorWithType("TypeError", format, a...)
}

func NewZeroDivisionError(format string, a ...interface{}) *Error {
	return newErrorWithType("ZeroDivisionError", format, a...)
}

func NewIndexError(format string, a ...interface{}) *Error {
	return newErrorWithType("IndexError", format, a...)
}

func NewKeyError(format string, a ...interface{}) *Error {
	return newErrorWithType("KeyError", format, a...)
}

func NewAttributeError(format string, a ...interface{}) *Error {
	return newErrorWithType("AttributeError", format, a...)
}

func NewNameError(format string, a ...interface{}) *Error {
	return newErrorWithType("NameError", format, a...)
}

func NewAssertionError(format string, a ...interface{}) *Error {
	return newErrorWithType("AssertionError", format, a...)
}

func NewRuntimeError(format string, a ...interface{}) *Error {
	return newErrorWithType("RuntimeError", format, a...)
}

func NewNotImplementedError(format string, a ...interface{}) *Error {
	return newErrorWithType("NotImplementedError", format, a...)
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
		Name:    "math",
		Fields: make(map[string]Object),
	}
	
	mathModule.Fields["pi"] = &Float{Value: math.Pi}
	mathModule.Fields["e"] = &Float{Value: math.E}
	
	mathModule.Fields["sin"] = &Builtin{
		Name: "math.sin",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("sin() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Sin(v.Value)}
			case *Integer:
				return &Float{Value: math.Sin(float64(v.Value))}
			default:
				return NewTypeError("sin() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["cos"] = &Builtin{
		Name: "math.cos",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("cos() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Cos(v.Value)}
			case *Integer:
				return &Float{Value: math.Cos(float64(v.Value))}
			default:
				return NewTypeError("cos() argument must be a number")
			}
		},
	}

	mathModule.Fields["tan"] = &Builtin{
		Name: "math.tan",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("tan() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Tan(v.Value)}
			case *Integer:
				return &Float{Value: math.Tan(float64(v.Value))}
			default:
				return NewTypeError("tan() argument must be a number")
			}
		},
	}

	mathModule.Fields["asin"] = &Builtin{
		Name: "math.asin",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("asin() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Asin(v.Value)}
			case *Integer:
				return &Float{Value: math.Asin(float64(v.Value))}
			default:
				return NewTypeError("asin() argument must be a number")
			}
		},
	}

	mathModule.Fields["acos"] = &Builtin{
		Name: "math.acos",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("acos() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Acos(v.Value)}
			case *Integer:
				return &Float{Value: math.Acos(float64(v.Value))}
			default:
				return NewTypeError("acos() argument must be a number")
			}
		},
	}

	mathModule.Fields["atan"] = &Builtin{
		Name: "math.atan",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("atan() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Atan(v.Value)}
			case *Integer:
				return &Float{Value: math.Atan(float64(v.Value))}
			default:
				return NewTypeError("atan() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["sqrt"] = &Builtin{
		Name: "math.sqrt",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("sqrt() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Sqrt(v.Value)}
			case *Integer:
				return &Float{Value: math.Sqrt(float64(v.Value))}
			default:
				return NewTypeError("sqrt() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["floor"] = &Builtin{
		Name: "math.floor",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("floor() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Integer{Value: int64(math.Floor(v.Value))}
			case *Integer:
				return v
			default:
				return NewTypeError("floor() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["ceil"] = &Builtin{
		Name: "math.ceil",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("ceil() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Integer{Value: int64(math.Ceil(v.Value))}
			case *Integer:
				return v
			default:
				return NewTypeError("ceil() argument must be a number")
			}
		},
	}

	mathModule.Fields["trunc"] = &Builtin{
		Name: "math.trunc",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("trunc() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Integer{Value: int64(math.Trunc(v.Value))}
			case *Integer:
				return v
			default:
				return NewTypeError("trunc() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["abs"] = &Builtin{
		Name: "math.abs",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("abs() takes exactly 1 argument")
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
				return NewTypeError("abs() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["pow"] = &Builtin{
		Name: "math.pow",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return NewTypeError("pow() takes exactly 2 arguments")
			}
			base, ok1 := args[0].(*Float)
			exp, ok2 := args[1].(*Float)
			if !ok1 || !ok2 {
				return NewTypeError("pow() arguments must be numbers")
			}
			return &Float{Value: math.Pow(base.Value, exp.Value)}
		},
	}

	mathModule.Fields["hypot"] = &Builtin{
		Name: "math.hypot",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return NewTypeError("hypot() takes exactly 2 arguments")
			}
			x, ok1 := args[0].(*Float)
			y, ok2 := args[1].(*Float)
			if !ok1 || !ok2 {
				return NewTypeError("hypot() arguments must be numbers")
			}
			return &Float{Value: math.Hypot(x.Value, y.Value)}
		},
	}
	
	mathModule.Fields["log"] = &Builtin{
		Name: "math.log",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("log() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Log(v.Value)}
			case *Integer:
				return &Float{Value: math.Log(float64(v.Value))}
			default:
				return NewTypeError("log() argument must be a number")
			}
		},
	}

	mathModule.Fields["log10"] = &Builtin{
		Name: "math.log10",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("log10() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Log10(v.Value)}
			case *Integer:
				return &Float{Value: math.Log10(float64(v.Value))}
			default:
				return NewTypeError("log10() argument must be a number")
			}
		},
	}

	mathModule.Fields["log2"] = &Builtin{
		Name: "math.log2",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("log2() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Log2(v.Value)}
			case *Integer:
				return &Float{Value: math.Log2(float64(v.Value))}
			default:
				return NewTypeError("log2() argument must be a number")
			}
		},
	}
	
	mathModule.Fields["exp"] = &Builtin{
		Name: "math.exp",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("exp() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: math.Exp(v.Value)}
			case *Integer:
				return &Float{Value: math.Exp(float64(v.Value))}
			default:
				return NewTypeError("exp() argument must be a number")
			}
		},
	}

	mathModule.Fields["degrees"] = &Builtin{
		Name: "math.degrees",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("degrees() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: v.Value * 180 / math.Pi}
			case *Integer:
				return &Float{Value: float64(v.Value) * 180 / math.Pi}
			default:
				return NewTypeError("degrees() argument must be a number")
			}
		},
	}

	mathModule.Fields["radians"] = &Builtin{
		Name: "math.radians",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("radians() takes exactly 1 argument")
			}
			arg := args[0]
			switch v := arg.(type) {
			case *Float:
				return &Float{Value: v.Value * math.Pi / 180}
			case *Integer:
				return &Float{Value: float64(v.Value) * math.Pi / 180}
			default:
				return NewTypeError("radians() argument must be a number")
			}
		},
	}
	
	return mathModule
}
