package functools

import (
	"github.com/go-py/go-python/pkg/objects"
)

// Partial represents a partial function application
type Partial struct {
	Func     objects.Object
	Args     []objects.Object
	Keywords *objects.Dict
}

func (p *Partial) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (p *Partial) Inspect() string {
	return "functools.partial"
}

func (p *Partial) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__call__":
		return &objects.Builtin{
			Name: "partial.__call__",
			Fn: func(args ...objects.Object) objects.Object {
				// Combine stored args and new args
				fullArgs := make([]objects.Object, 0, len(p.Args)+len(args))
				fullArgs = append(fullArgs, p.Args...)
				fullArgs = append(fullArgs, args...)
				
				// For now, we'll pass all args (keywords handling will be basic)
				return objects.CallFunction(p.Func, fullArgs...)
			},
		}, true
	}
	return nil, false
}

// CreateFunctoolsModule creates the functools module
func CreateFunctoolsModule() *objects.Module {
	module := &objects.Module{
		Name:   "functools",
		Fields: make(map[string]objects.Object),
	}

	// functools.reduce(function, iterable[, initializer])
	module.Fields["reduce"] = &objects.Builtin{
		Name: "functools.reduce",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("reduce() takes at least 2 arguments")
			}
			
			function := args[0]
			iterable := args[1]
			
			// Get elements from iterable
			var elements []objects.Object
			switch v := iterable.(type) {
			case *objects.List:
				elements = v.Elements
			case *objects.Tuple:
				elements = v.Elements
			case *objects.Range:
				elements = v.ToList()
			case *objects.Set:
				elements = v.ToSlice()
			default:
				return objects.NewTypeError("'%s' object is not iterable", iterable.Type())
			}
			
			var accumulator objects.Object
			startIndex := 0
			
			if len(args) >= 3 {
				accumulator = args[2]
			} else {
				if len(elements) == 0 {
					return objects.NewTypeError("reduce() of empty sequence with no initial value")
				}
				accumulator = elements[0]
				startIndex = 1
			}
			
			for i := startIndex; i < len(elements); i++ {
				accumulator = objects.CallFunction(function, accumulator, elements[i])
				if accumulator.Type() == objects.ERROR_OBJ {
					return accumulator
				}
			}
			
			return accumulator
		},
	}

	// functools.partial(func, *args, **keywords)
	module.Fields["partial"] = &objects.Builtin{
		Name: "functools.partial",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("partial() takes at least 1 argument")
			}
			
			funcObj := args[0]
			storedArgs := args[1:]
			
			// Check if last argument is a dict for keywords (basic implementation)
			var keywords *objects.Dict = objects.NewDict()
			if len(storedArgs) > 0 {
				if kw, ok := storedArgs[len(storedArgs)-1].(*objects.Dict); ok {
					keywords = kw
					storedArgs = storedArgs[:len(storedArgs)-1]
				}
			}
			
			return &Partial{
				Func:     funcObj,
				Args:     storedArgs,
				Keywords: keywords,
			}
		},
	}

	return module
}
