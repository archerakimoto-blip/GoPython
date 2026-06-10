package compiler

import (
	"fmt"
	"os"

	"github.com/go-py/go-python/pkg/objects"
)

// BuiltinEntry represents a builtin function/value to be registered.
type BuiltinEntry struct {
	Name    string
	Builtin *objects.Builtin
	Value   objects.Object // Used for non-callable builtins like None
}

// GetCommonBuiltins returns the list of common builtin functions and values
// that should be registered in both the stack and register compilers.
func GetCommonBuiltins() []BuiltinEntry {
	var entries []BuiltinEntry

	// print
	entries = append(entries, BuiltinEntry{
		Name: "print",
		Builtin: &objects.Builtin{
			Name: "print",
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
		},
	})

	// len
	entries = append(entries, BuiltinEntry{
		Name: "len",
		Builtin: &objects.Builtin{
			Name: "len",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewError("len() takes at least 1 argument")
				}
				switch arg := args[0].(type) {
				case *objects.String:
					return &objects.Integer{Value: int64(len(arg.Value))}
				case *objects.List:
					return &objects.Integer{Value: int64(len(arg.Elements))}
				case *objects.Dict:
					return &objects.Integer{Value: int64(arg.Size())}
				case *objects.Tuple:
					return &objects.Integer{Value: int64(len(arg.Elements))}
				case *objects.Set:
					return &objects.Integer{Value: int64(arg.Size())}
				case *objects.Range:
					return &objects.Integer{Value: arg.Len()}
				default:
					return objects.NewError("argument to 'len' not supported: %s", arg.Type())
				}
			},
		},
	})

	// None
	entries = append(entries, BuiltinEntry{
		Name:  "None",
		Value: objects.None_,
	})

	// range
	entries = append(entries, BuiltinEntry{
		Name: "range",
		Builtin: &objects.Builtin{
			Name: "range",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 3 {
					return objects.NewError("range() takes 1-3 arguments")
				}
				var start, stop, step int64
				switch len(args) {
				case 1:
					start = 0
					if arg, ok := args[0].(*objects.Integer); ok {
						stop = arg.Value
					} else {
						return objects.NewError("range() argument must be an integer")
					}
					step = 1
				case 2:
					if arg, ok := args[0].(*objects.Integer); ok {
						start = arg.Value
					} else {
						return objects.NewError("range() argument must be an integer")
					}
					if arg, ok := args[1].(*objects.Integer); ok {
						stop = arg.Value
					} else {
						return objects.NewError("range() argument must be an integer")
					}
					step = 1
				case 3:
					if arg, ok := args[0].(*objects.Integer); ok {
						start = arg.Value
					} else {
						return objects.NewError("range() argument must be an integer")
					}
					if arg, ok := args[1].(*objects.Integer); ok {
						stop = arg.Value
					} else {
						return objects.NewError("range() argument must be an integer")
					}
					if arg, ok := args[2].(*objects.Integer); ok {
						step = arg.Value
					} else {
						return objects.NewError("range() argument must be an integer")
					}
				}
				return &objects.Range{Start: start, Stop: stop, Step: step}
			},
		},
	})

	// abs
	entries = append(entries, BuiltinEntry{
		Name: "abs",
		Builtin: &objects.Builtin{
			Name: "abs",
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
		},
	})

	// min
	entries = append(entries, BuiltinEntry{
		Name: "min",
		Builtin: &objects.Builtin{
			Name: "min",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewError("min() takes at least 1 argument")
				}
				if len(args) == 1 {
					if lst, ok := args[0].(*objects.List); ok {
						if len(lst.Elements) == 0 {
							return objects.NewError("min() arg is an empty sequence")
						}
						min := lst.Elements[0]
						for _, e := range lst.Elements[1:] {
							if compareObjects(e, min) < 0 {
								min = e
							}
						}
						return min
					}
				}
				min := args[0]
				for _, arg := range args[1:] {
					if compareObjects(arg, min) < 0 {
						min = arg
					}
				}
				return min
			},
		},
	})

	// max
	entries = append(entries, BuiltinEntry{
		Name: "max",
		Builtin: &objects.Builtin{
			Name: "max",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewError("max() takes at least 1 argument")
				}
				if len(args) == 1 {
					if lst, ok := args[0].(*objects.List); ok {
						if len(lst.Elements) == 0 {
							return objects.NewError("max() arg is an empty sequence")
						}
						max := lst.Elements[0]
						for _, e := range lst.Elements[1:] {
							if compareObjects(e, max) > 0 {
								max = e
							}
						}
						return max
					}
				}
				max := args[0]
				for _, arg := range args[1:] {
					if compareObjects(arg, max) > 0 {
						max = arg
					}
				}
				return max
			},
		},
	})

	// sum
	entries = append(entries, BuiltinEntry{
		Name: "sum",
		Builtin: &objects.Builtin{
			Name: "sum",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewError("sum() takes at least 1 argument")
				}
				var elements []objects.Object
				switch arg := args[0].(type) {
				case *objects.List:
					elements = arg.Elements
				case *objects.Tuple:
					elements = arg.Elements
				case *objects.Range:
					elements = arg.ToList()
				default:
					return objects.NewError("sum() argument must be iterable")
				}
				start := int64(0)
				if len(args) > 1 {
					if s, ok := args[1].(*objects.Integer); ok {
						start = s.Value
					}
				}
				result := start
				for _, e := range elements {
					if i, ok := e.(*objects.Integer); ok {
						result += i.Value
					} else {
						return objects.NewError("sum() can only sum integers")
					}
				}
				return &objects.Integer{Value: result}
			},
		},
	})

	// str
	entries = append(entries, BuiltinEntry{
		Name: "str",
		Builtin: &objects.Builtin{
			Name: "str",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("str() takes exactly 1 argument")
				}
				return &objects.String{Value: args[0].Inspect()}
			},
		},
	})

	// int
	entries = append(entries, BuiltinEntry{
		Name: "int",
		Builtin: &objects.Builtin{
			Name: "int",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) == 0 {
					return &objects.Integer{Value: 0}
				}
				if len(args) > 2 {
					return objects.NewError("int() takes at most 2 arguments")
				}
				switch arg := args[0].(type) {
				case *objects.Integer:
					return arg
				case *objects.Float:
					return &objects.Integer{Value: int64(arg.Value)}
				case *objects.String:
					base := 10
					if len(args) > 1 {
						if b, ok := args[1].(*objects.Integer); ok {
							base = int(b.Value)
						}
					}
					var val int64
					if base == 10 {
						_, err := fmt.Sscanf(arg.Value, "%d", &val)
						if err != nil {
							return objects.NewError("invalid literal for int(): %s", arg.Value)
						}
					} else {
						_, err := fmt.Sscanf(arg.Value, "%d", &val)
						if err != nil {
							return objects.NewError("invalid literal for int() with base %d: %s", base, arg.Value)
						}
					}
					return &objects.Integer{Value: val}
				case *objects.Boolean:
					if arg.Value {
						return &objects.Integer{Value: 1}
					}
					return &objects.Integer{Value: 0}
				default:
					return objects.NewError("int() argument must be a number or string")
				}
			},
		},
	})

	// float
	entries = append(entries, BuiltinEntry{
		Name: "float",
		Builtin: &objects.Builtin{
			Name: "float",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) == 0 {
					return &objects.Float{Value: 0.0}
				}
				if len(args) != 1 {
					return objects.NewError("float() takes at most 1 argument")
				}
				switch arg := args[0].(type) {
				case *objects.Integer:
					return &objects.Float{Value: float64(arg.Value)}
				case *objects.Float:
					return arg
				case *objects.String:
					var val float64
					_, err := fmt.Sscanf(arg.Value, "%f", &val)
					if err != nil {
						return objects.NewError("could not convert string to float: %s", arg.Value)
					}
					return &objects.Float{Value: val}
				case *objects.Boolean:
					if arg.Value {
						return &objects.Float{Value: 1.0}
					}
					return &objects.Float{Value: 0.0}
				default:
					return objects.NewError("float() argument must be a number or string")
				}
			},
		},
	})

	// bool
	entries = append(entries, BuiltinEntry{
		Name: "bool",
		Builtin: &objects.Builtin{
			Name: "bool",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) == 0 {
					return objects.False
				}
				if len(args) != 1 {
					return objects.NewError("bool() takes at most 1 argument")
				}
				return nativeBoolToBooleanObject(isTruthy(args[0]))
			},
		},
	})

	// list
	entries = append(entries, BuiltinEntry{
		Name: "list",
		Builtin: &objects.Builtin{
			Name: "list",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return objects.NewError("list() takes at most 1 argument")
				}
				if len(args) == 0 {
					return &objects.List{Elements: []objects.Object{}}
				}
				switch arg := args[0].(type) {
				case *objects.List:
					return arg
				case *objects.Tuple:
					return &objects.List{Elements: append([]objects.Object{}, arg.Elements...)}
				case *objects.String:
					elements := make([]objects.Object, len(arg.Value))
					for i, ch := range arg.Value {
						elements[i] = &objects.String{Value: string(ch)}
					}
					return &objects.List{Elements: elements}
				case *objects.Range:
					return &objects.List{Elements: arg.ToList()}
				default:
					return objects.NewError("list() argument must be iterable")
				}
			},
		},
	})

	// set
	entries = append(entries, BuiltinEntry{
		Name: "set",
		Builtin: &objects.Builtin{
			Name: "set",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return objects.NewError("set() takes at most 1 argument")
				}
				if len(args) == 0 {
					return objects.NewSet()
				}
				switch arg := args[0].(type) {
				case *objects.List:
					s := objects.NewSet()
					for _, e := range arg.Elements {
						s.Add(e)
					}
					return s
				case *objects.Set:
					return arg
				default:
					return objects.NewError("set() argument must be iterable")
				}
			},
		},
	})

	// dict
	entries = append(entries, BuiltinEntry{
		Name: "dict",
		Builtin: &objects.Builtin{
			Name: "dict",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDict()
			},
		},
	})

	// tuple
	entries = append(entries, BuiltinEntry{
		Name: "tuple",
		Builtin: &objects.Builtin{
			Name: "tuple",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return objects.NewError("tuple() takes at most 1 argument")
				}
				if len(args) == 0 {
					return &objects.Tuple{Elements: []objects.Object{}}
				}
				switch arg := args[0].(type) {
				case *objects.List:
					return &objects.Tuple{Elements: append([]objects.Object{}, arg.Elements...)}
				case *objects.Tuple:
					return arg
				default:
					return objects.NewError("tuple() argument must be iterable")
				}
			},
		},
	})

	// type
	entries = append(entries, BuiltinEntry{
		Name: "type",
		Builtin: &objects.Builtin{
			Name: "type",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("type() takes exactly 1 argument")
				}
				return &objects.String{Value: string(args[0].Type())}
			},
		},
	})

	// isinstance
	entries = append(entries, BuiltinEntry{
		Name: "isinstance",
		Builtin: &objects.Builtin{
			Name: "isinstance",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return objects.NewError("isinstance() takes exactly 2 arguments")
				}
				typeName := string(args[0].Type())
				if typeStr, ok := args[1].(*objects.String); ok {
					return nativeBoolToBooleanObject(typeName == typeStr.Value)
				}
				return objects.False
			},
		},
	})

	// append
	entries = append(entries, BuiltinEntry{
		Name: "append",
		Builtin: &objects.Builtin{
			Name: "append",
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
		},
	})

	// enumerate
	entries = append(entries, BuiltinEntry{
		Name: "enumerate",
		Builtin: &objects.Builtin{
			Name: "enumerate",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return objects.NewError("enumerate() takes 1-2 arguments")
				}
				start := int64(0)
				if len(args) > 1 {
					if s, ok := args[1].(*objects.Integer); ok {
						start = s.Value
					}
				}
				var elements []objects.Object
				switch arg := args[0].(type) {
				case *objects.List:
					elements = arg.Elements
				case *objects.Tuple:
					elements = arg.Elements
				case *objects.Range:
					elements = arg.ToList()
				case *objects.String:
					elements = make([]objects.Object, len(arg.Value))
					for i, ch := range arg.Value {
						elements[i] = &objects.String{Value: string(ch)}
					}
				default:
					return objects.NewError("enumerate() argument must be iterable")
				}
				pairs := make([]objects.Object, len(elements))
				for i, e := range elements {
					pairs[i] = &objects.Tuple{Elements: []objects.Object{
						&objects.Integer{Value: start + int64(i)},
						e,
					}}
				}
				return &objects.List{Elements: pairs}
			},
		},
	})

	// map
	entries = append(entries, BuiltinEntry{
		Name: "map",
		Builtin: &objects.Builtin{
			Name: "map",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return objects.NewError("map() takes exactly 2 arguments")
				}
				fn := args[0]
				var elements []objects.Object
				switch arg := args[1].(type) {
				case *objects.List:
					elements = arg.Elements
				case *objects.Tuple:
					elements = arg.Elements
				case *objects.Range:
					elements = arg.ToList()
				default:
					return objects.NewError("map() second argument must be iterable")
				}
				results := make([]objects.Object, len(elements))
				for i, e := range elements {
					if callable, ok := fn.(objects.Callable); ok {
						results[i] = callable.Call(e)
					} else {
						return objects.NewError("map() first argument must be callable")
					}
				}
				return &objects.List{Elements: results}
			},
		},
	})

	// filter
	entries = append(entries, BuiltinEntry{
		Name: "filter",
		Builtin: &objects.Builtin{
			Name: "filter",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return objects.NewError("filter() takes exactly 2 arguments")
				}
				fn := args[0]
				var elements []objects.Object
				switch arg := args[1].(type) {
				case *objects.List:
					elements = arg.Elements
				case *objects.Tuple:
					elements = arg.Elements
				case *objects.Range:
					elements = arg.ToList()
				default:
					return objects.NewError("filter() second argument must be iterable")
				}
				var results []objects.Object
				for _, e := range elements {
					if callable, ok := fn.(objects.Callable); ok {
						result := callable.Call(e)
						if isTruthy(result) {
							results = append(results, e)
						}
					} else {
						return objects.NewError("filter() first argument must be callable")
					}
				}
				return &objects.List{Elements: results}
			},
		},
	})

	// sorted
	entries = append(entries, BuiltinEntry{
		Name: "sorted",
		Builtin: &objects.Builtin{
			Name: "sorted",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewError("sorted() takes at least 1 argument")
				}
				var elements []objects.Object
				switch arg := args[0].(type) {
				case *objects.List:
					elements = append([]objects.Object{}, arg.Elements...)
				case *objects.Tuple:
					elements = append([]objects.Object{}, arg.Elements...)
				case *objects.Range:
					elements = arg.ToList()
				default:
					return objects.NewError("sorted() argument must be iterable")
				}
				// Simple bubble sort
				for i := 0; i < len(elements); i++ {
					for j := i + 1; j < len(elements); j++ {
						if compareObjects(elements[i], elements[j]) > 0 {
							elements[i], elements[j] = elements[j], elements[i]
						}
					}
				}
				return &objects.List{Elements: elements}
			},
		},
	})

	// reversed
	entries = append(entries, BuiltinEntry{
		Name: "reversed",
		Builtin: &objects.Builtin{
			Name: "reversed",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("reversed() takes exactly 1 argument")
				}
				var elements []objects.Object
				switch arg := args[0].(type) {
				case *objects.List:
					elements = append([]objects.Object{}, arg.Elements...)
				case *objects.Tuple:
					elements = append([]objects.Object{}, arg.Elements...)
				case *objects.Range:
					elements = arg.ToList()
				default:
					return objects.NewError("reversed() argument must be iterable")
				}
				for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
					elements[i], elements[j] = elements[j], elements[i]
				}
				return &objects.List{Elements: elements}
			},
		},
	})

	// repr
	entries = append(entries, BuiltinEntry{
		Name: "repr",
		Builtin: &objects.Builtin{
			Name: "repr",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("repr() takes exactly 1 argument")
				}
				return &objects.String{Value: args[0].Inspect()}
			},
		},
	})

	// iter
	entries = append(entries, BuiltinEntry{
		Name: "iter",
		Builtin: &objects.Builtin{
			Name: "iter",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("iter() takes exactly 1 argument")
				}
				return args[0]
			},
		},
	})

	// next
	entries = append(entries, BuiltinEntry{
		Name: "next",
		Builtin: &objects.Builtin{
			Name: "next",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewError("next() takes at least 1 argument")
				}
				return objects.NewError("next() not fully implemented in register VM")
			},
		},
	})

	// any
	entries = append(entries, BuiltinEntry{
		Name: "any",
		Builtin: &objects.Builtin{
			Name: "any",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("any() takes exactly 1 argument")
				}
				var elements []objects.Object
				switch arg := args[0].(type) {
				case *objects.List:
					elements = arg.Elements
				case *objects.Tuple:
					elements = arg.Elements
				case *objects.Range:
					elements = arg.ToList()
				default:
					return objects.NewError("any() argument must be iterable")
				}
				for _, e := range elements {
					if isTruthy(e) {
						return objects.True
					}
				}
				return objects.False
			},
		},
	})

	// all
	entries = append(entries, BuiltinEntry{
		Name: "all",
		Builtin: &objects.Builtin{
			Name: "all",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("all() takes exactly 1 argument")
				}
				var elements []objects.Object
				switch arg := args[0].(type) {
				case *objects.List:
					elements = arg.Elements
				case *objects.Tuple:
					elements = arg.Elements
				case *objects.Range:
					elements = arg.ToList()
				default:
					return objects.NewError("all() argument must be iterable")
				}
				for _, e := range elements {
					if !isTruthy(e) {
						return objects.False
					}
				}
				return objects.True
			},
		},
	})

	// chr
	entries = append(entries, BuiltinEntry{
		Name: "chr",
		Builtin: &objects.Builtin{
			Name: "chr",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("chr() takes exactly 1 argument")
				}
				if arg, ok := args[0].(*objects.Integer); ok {
					return &objects.String{Value: string(rune(arg.Value))}
				}
				return objects.NewError("chr() argument must be an integer")
			},
		},
	})

	// ord
	entries = append(entries, BuiltinEntry{
		Name: "ord",
		Builtin: &objects.Builtin{
			Name: "ord",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("ord() takes exactly 1 argument")
				}
				if arg, ok := args[0].(*objects.String); ok && len(arg.Value) > 0 {
					return &objects.Integer{Value: int64(arg.Value[0])}
				}
				return objects.NewError("ord() argument must be a string of length 1")
			},
		},
	})

	// hex
	entries = append(entries, BuiltinEntry{
		Name: "hex",
		Builtin: &objects.Builtin{
			Name: "hex",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("hex() takes exactly 1 argument")
				}
				if arg, ok := args[0].(*objects.Integer); ok {
					return &objects.String{Value: fmt.Sprintf("0x%x", arg.Value)}
				}
				return objects.NewError("hex() argument must be an integer")
			},
		},
	})

	// oct
	entries = append(entries, BuiltinEntry{
		Name: "oct",
		Builtin: &objects.Builtin{
			Name: "oct",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("oct() takes exactly 1 argument")
				}
				if arg, ok := args[0].(*objects.Integer); ok {
					return &objects.String{Value: fmt.Sprintf("0o%o", arg.Value)}
				}
				return objects.NewError("oct() argument must be an integer")
			},
		},
	})

	// bin
	entries = append(entries, BuiltinEntry{
		Name: "bin",
		Builtin: &objects.Builtin{
			Name: "bin",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("bin() takes exactly 1 argument")
				}
				if arg, ok := args[0].(*objects.Integer); ok {
					return &objects.String{Value: fmt.Sprintf("0b%b", arg.Value)}
				}
				return objects.NewError("bin() argument must be an integer")
			},
		},
	})

	// format
	entries = append(entries, BuiltinEntry{
		Name: "format",
		Builtin: &objects.Builtin{
			Name: "format",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return objects.NewError("format() takes 1-2 arguments")
				}
				if len(args) == 1 {
					return &objects.String{Value: args[0].Inspect()}
				}
				return &objects.String{Value: args[0].Inspect()}
			},
		},
	})

	// divmod
	entries = append(entries, BuiltinEntry{
		Name: "divmod",
		Builtin: &objects.Builtin{
			Name: "divmod",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return objects.NewError("divmod() takes exactly 2 arguments")
				}
				a, ok1 := args[0].(*objects.Integer)
				b, ok2 := args[1].(*objects.Integer)
				if !ok1 || !ok2 {
					return objects.NewError("divmod() arguments must be integers")
				}
				return &objects.Tuple{Elements: []objects.Object{
					&objects.Integer{Value: a.Value / b.Value},
					&objects.Integer{Value: a.Value % b.Value},
				}}
			},
		},
	})

	// input
	entries = append(entries, BuiltinEntry{
		Name: "input",
		Builtin: &objects.Builtin{
			Name: "input",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return objects.NewError("input() takes at most 1 argument")
				}
				if len(args) == 1 {
					fmt.Print(args[0].Inspect())
				}
				var line string
				fmt.Scanln(&line)
				return &objects.String{Value: line}
			},
		},
	})

	// round
	entries = append(entries, BuiltinEntry{
		Name: "round",
		Builtin: &objects.Builtin{
			Name: "round",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 || len(args) > 2 {
					return objects.NewError("round() takes 1-2 arguments")
				}
				switch arg := args[0].(type) {
				case *objects.Integer:
					return arg
				case *objects.Float:
					ndigits := 0
					if len(args) > 1 {
						if d, ok := args[1].(*objects.Integer); ok {
							ndigits = int(d.Value)
						}
					}
					pow := 1.0
					for i := 0; i < ndigits; i++ {
						pow *= 10
					}
					return &objects.Float{Value: float64(int64(arg.Value*pow+0.5)) / pow}
				default:
					return objects.NewError("round() argument must be a number")
				}
			},
		},
	})

	// zip
	entries = append(entries, BuiltinEntry{
		Name: "zip",
		Builtin: &objects.Builtin{
			Name: "zip",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return objects.NewError("zip() takes at least 2 arguments")
				}
				lists := make([][]objects.Object, len(args))
				minLen := -1
				for i, arg := range args {
					switch a := arg.(type) {
					case *objects.List:
						lists[i] = a.Elements
					case *objects.Tuple:
						lists[i] = a.Elements
					case *objects.Range:
						lists[i] = a.ToList()
					default:
						return objects.NewError("zip() arguments must be iterable")
					}
					if minLen == -1 || len(lists[i]) < minLen {
						minLen = len(lists[i])
					}
				}
				result := make([]objects.Object, minLen)
				for i := 0; i < minLen; i++ {
					pair := make([]objects.Object, len(lists))
					for j, lst := range lists {
						pair[j] = lst[i]
					}
					result[i] = &objects.Tuple{Elements: pair}
				}
				return &objects.List{Elements: result}
			},
		},
	})

	// hasattr
	entries = append(entries, BuiltinEntry{
		Name: "hasattr",
		Builtin: &objects.Builtin{
			Name: "hasattr",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return objects.NewError("hasattr() takes exactly 2 arguments")
				}
				return objects.False
			},
		},
	})

	// getattr
	entries = append(entries, BuiltinEntry{
		Name: "getattr",
		Builtin: &objects.Builtin{
			Name: "getattr",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return objects.NewError("getattr() takes at least 2 arguments")
				}
				return objects.NewError("getattr() not fully supported")
			},
		},
	})

	// setattr
	entries = append(entries, BuiltinEntry{
		Name: "setattr",
		Builtin: &objects.Builtin{
			Name: "setattr",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return objects.NewError("setattr() takes exactly 3 arguments")
				}
				return objects.None_
			},
		},
	})

	// dir
	entries = append(entries, BuiltinEntry{
		Name: "dir",
		Builtin: &objects.Builtin{
			Name: "dir",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.List{Elements: []objects.Object{}}
			},
		},
	})

	// id
	entries = append(entries, BuiltinEntry{
		Name: "id",
		Builtin: &objects.Builtin{
			Name: "id",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("id() takes exactly 1 argument")
				}
				return &objects.Integer{Value: int64(uintptr(fmt.Sprintf("%p", args[0])[0]))}
			},
		},
	})

	// hash
	entries = append(entries, BuiltinEntry{
		Name: "hash",
		Builtin: &objects.Builtin{
			Name: "hash",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("hash() takes exactly 1 argument")
				}
				return &objects.Integer{Value: int64(len(args[0].Inspect()))}
			},
		},
	})

	// callable
	entries = append(entries, BuiltinEntry{
		Name: "callable",
		Builtin: &objects.Builtin{
			Name: "callable",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewError("callable() takes exactly 1 argument")
				}
				_, ok := args[0].(objects.Callable)
				return nativeBoolToBooleanObject(ok)
			},
		},
	})

	// setitem
	entries = append(entries, BuiltinEntry{
		Name: "setitem",
		Builtin: &objects.Builtin{
			Name: "setitem",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 3 {
					return objects.NewError("setitem() takes exactly 3 arguments")
				}
				dict, ok := args[0].(*objects.Dict)
				if !ok {
					return objects.NewError("first argument to setitem() must be a dict")
				}
				dict.Set(args[1], args[2])
				return objects.None_
			},
		},
	})

	// setadd
	entries = append(entries, BuiltinEntry{
		Name: "setadd",
		Builtin: &objects.Builtin{
			Name: "setadd",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 2 {
					return objects.NewError("setadd() takes exactly 2 arguments")
				}
				s, ok := args[0].(*objects.Set)
				if !ok {
					return objects.NewError("first argument to setadd() must be a set")
				}
				s.Add(args[1])
				return objects.None_
			},
		},
	})

	// frozenset
	entries = append(entries, BuiltinEntry{
		Name: "frozenset",
		Builtin: &objects.Builtin{
			Name: "frozenset",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) > 1 {
					return objects.NewError("frozenset() takes at most 1 argument")
				}
				return &objects.Set{Elements: make(map[string]objects.Object)}
			},
		},
	})

	// open
	entries = append(entries, BuiltinEntry{
		Name: "open",
		Builtin: &objects.Builtin{
			Name: "open",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewError("open() not fully implemented in register VM")
			},
		},
	})

	// Register exception types
	exceptionTypes := []struct {
		name string
	}{
		{"Exception"}, {"TypeError"}, {"ValueError"}, {"RuntimeError"},
		{"NameError"}, {"IndexError"}, {"KeyError"}, {"AttributeError"},
		{"ZeroDivisionError"}, {"OverflowError"}, {"StopIteration"},
		{"NotImplementedError"}, {"IOError"}, {"OSError"}, {"ImportError"},
		{"FileNotFoundError"}, {"PermissionError"}, {"IsADirectoryError"},
		{"NotADirectoryError"}, {"FileExistsError"}, {"ProcessLookupError"},
		{"ChildProcessError"}, {"InterruptedError"}, {"ConnectionError"},
		{"ConnectionAbortedError"}, {"ConnectionRefusedError"},
		{"ConnectionResetError"}, {"BrokenPipeError"}, {"TimeoutError"},
		{"EOFError"}, {"MemoryError"}, {"RecursionError"},
		{"UnboundLocalError"}, {"UnicodeError"}, {"UnicodeDecodeError"},
		{"UnicodeEncodeError"}, {"UnicodeTranslateError"},
		{"SystemError"}, {"SystemExit"}, {"KeyboardInterrupt"},
		{"GeneratorExit"}, {"ArithmeticError"}, {"BufferError"},
		{"LookupError"}, {"EnvironmentError"},
	}
	for _, et := range exceptionTypes {
		name := et.name
		entries = append(entries, BuiltinEntry{
			Name: name,
			Builtin: &objects.Builtin{
				Name: name,
				Fn: func(args ...objects.Object) objects.Object {
					msg := ""
					if len(args) > 0 {
						msg = args[0].Inspect()
					}
					return &objects.Error{Message: msg}
				},
			},
		})
	}

	return entries
}

// compareObjects compares two objects and returns -1, 0, or 1.
func compareObjects(a, b objects.Object) int {
	switch a := a.(type) {
	case *objects.Integer:
		if b, ok := b.(*objects.Integer); ok {
			if a.Value < b.Value {
				return -1
			} else if a.Value > b.Value {
				return 1
			}
			return 0
		}
	case *objects.Float:
		if b, ok := b.(*objects.Float); ok {
			if a.Value < b.Value {
				return -1
			} else if a.Value > b.Value {
				return 1
			}
			return 0
		}
	case *objects.String:
		if b, ok := b.(*objects.String); ok {
			if a.Value < b.Value {
				return -1
			} else if a.Value > b.Value {
				return 1
			}
			return 0
		}
	}
	return 0
}

// isTruthy checks if an object is truthy.
func isTruthy(obj objects.Object) bool {
	switch obj := obj.(type) {
	case *objects.Boolean:
		return obj.Value
	case *objects.Integer:
		return obj.Value != 0
	case *objects.Float:
		return obj.Value != 0.0
	case *objects.String:
		return obj.Value != ""
	case *objects.List:
		return len(obj.Elements) > 0
	case *objects.Tuple:
		return len(obj.Elements) > 0
	case *objects.Dict:
		return obj.Size() > 0
	case *objects.Set:
		return len(obj.Elements) > 0
	case *objects.None:
		return false
	case nil:
		return false
	default:
		return true
	}
}

// nativeBoolToBooleanObject converts a Go bool to a Boolean object.
func nativeBoolToBooleanObject(b bool) *objects.Boolean {
	if b {
		return objects.True
	}
	return objects.False
}
