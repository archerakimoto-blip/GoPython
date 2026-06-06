package operator

import (
	"fmt"

	"github.com/go-py/go-python/pkg/objects"
)

// itemGetter represents the itemgetter function object
type itemGetter struct {
	items []objects.Object
}

// Ensure itemGetter implements both Object and Callable interfaces
var _ objects.Object = (*itemGetter)(nil)
var _ objects.Callable = (*itemGetter)(nil)

func (ig *itemGetter) Type() objects.ObjectType { return objects.FUNCTION_OBJ }
func (ig *itemGetter) Inspect() string {
	if len(ig.items) == 1 {
		return fmt.Sprintf("operator.itemgetter(%s)", ig.items[0].Inspect())
	}
	parts := make([]string, len(ig.items))
	for i, item := range ig.items {
		parts[i] = item.Inspect()
	}
	return fmt.Sprintf("operator.itemgetter(%s)", join(parts, ", "))
}

// Call implements the callable behavior for itemGetter
func (ig *itemGetter) Call(args ...objects.Object) objects.Object {
	if len(args) < 1 {
		return objects.NewTypeError("itemgetter expected at least 1 argument, got 0")
	}
	obj := args[0]

	if len(ig.items) == 1 {
		return getItem(obj, ig.items[0])
	}

	// Multiple items, return a tuple
	results := make([]objects.Object, len(ig.items))
	for i, item := range ig.items {
		results[i] = getItem(obj, item)
	}
	return &objects.Tuple{Elements: results}
}

// getItem gets an item from an object using __getitem__
func getItem(obj, key objects.Object) objects.Object {
	switch o := obj.(type) {
	case *objects.List:
		if idx, ok := key.(*objects.Integer); ok {
			i := int(idx.Value)
			if i < 0 {
				i += len(o.Elements)
			}
			if i >= 0 && i < len(o.Elements) {
				return o.Elements[i]
			}
			return objects.NewIndexError("list index out of range")
		}
		return objects.NewTypeError("list indices must be integers")
	case *objects.Tuple:
		if idx, ok := key.(*objects.Integer); ok {
			i := int(idx.Value)
			if i < 0 {
				i += len(o.Elements)
			}
			if i >= 0 && i < len(o.Elements) {
				return o.Elements[i]
			}
			return objects.NewIndexError("tuple index out of range")
		}
		return objects.NewTypeError("tuple indices must be integers")
	case *objects.Dict:
		if val, ok := o.Get(key); ok {
			return val
		}
		return objects.NewKeyError("%s", key.Inspect())
	case *objects.String:
		if idx, ok := key.(*objects.Integer); ok {
			i := int(idx.Value)
			if i < 0 {
				i += len(o.Value)
			}
			if i >= 0 && i < len(o.Value) {
				return &objects.String{Value: string(o.Value[i])}
			}
			return objects.NewIndexError("string index out of range")
		}
		return objects.NewTypeError("string indices must be integers")
	default:
		// Try __getitem__ attribute
		if attrGetter, ok := obj.(objects.AttributeGetter); ok {
			if getitem, ok := attrGetter.GetAttr("__getitem__"); ok {
				return objects.CallFunction(getitem, key)
			}
		}
		return objects.NewTypeError("'%s' object is not subscriptable", obj.Type())
	}
}

// getAttrHelper is a helper to get an attribute from any Object
func getAttrHelper(obj objects.Object, name string) (objects.Object, bool) {
	if attrGetter, ok := obj.(objects.AttributeGetter); ok {
		return attrGetter.GetAttr(name)
	}
	return nil, false
}

// attrGetter represents the attrgetter function object
type attrGetter struct {
	attrs []string
}

// Ensure attrGetter implements both Object and Callable interfaces
var _ objects.Object = (*attrGetter)(nil)
var _ objects.Callable = (*attrGetter)(nil)

func (ag *attrGetter) Type() objects.ObjectType { return objects.FUNCTION_OBJ }
func (ag *attrGetter) Inspect() string {
	if len(ag.attrs) == 1 {
		return fmt.Sprintf("operator.attrgetter(%q)", ag.attrs[0])
	}
	parts := make([]string, len(ag.attrs))
	for i, attr := range ag.attrs {
		parts[i] = fmt.Sprintf("%q", attr)
	}
	return fmt.Sprintf("operator.attrgetter(%s)", join(parts, ", "))
}

// Call implements the callable behavior for attrGetter
func (ag *attrGetter) Call(args ...objects.Object) objects.Object {
	if len(args) < 1 {
		return objects.NewTypeError("attrgetter expected at least 1 argument, got 0")
	}
	obj := args[0]

	if len(ag.attrs) == 1 {
		return getAttrPath(obj, ag.attrs[0])
	}

	// Multiple attrs, return a tuple
	results := make([]objects.Object, len(ag.attrs))
	for i, attr := range ag.attrs {
		results[i] = getAttrPath(obj, attr)
	}
	return &objects.Tuple{Elements: results}
}

// getAttrPath gets an attribute from an object, supporting nested attributes like "a.b.c"
func getAttrPath(obj objects.Object, path string) objects.Object {
	current := obj
	parts := splitPath(path)
	for _, part := range parts {
		if attr, ok := getAttrHelper(current, part); ok {
			current = attr
		} else {
			return objects.NewAttributeError("'%s' object has no attribute '%s'", current.Type(), part)
		}
	}
	return current
}

// splitPath splits "a.b.c" into ["a", "b", "c"]
func splitPath(path string) []string {
	var parts []string
	start := 0
	for i, c := range path {
		if c == '.' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	parts = append(parts, path[start:])
	return parts
}

// methodCaller represents the methodcaller function object
type methodCaller struct {
	name string
	args []objects.Object
}

// Ensure methodCaller implements both Object and Callable interfaces
var _ objects.Object = (*methodCaller)(nil)
var _ objects.Callable = (*methodCaller)(nil)

func (mc *methodCaller) Type() objects.ObjectType { return objects.FUNCTION_OBJ }
func (mc *methodCaller) Inspect() string {
	argParts := make([]string, len(mc.args))
	for i, arg := range mc.args {
		argParts[i] = arg.Inspect()
	}
	return fmt.Sprintf("operator.methodcaller(%q, %s)", mc.name, join(argParts, ", "))
}

// Call implements the callable behavior for methodCaller
func (mc *methodCaller) Call(args ...objects.Object) objects.Object {
	if len(args) < 1 {
		return objects.NewTypeError("methodcaller expected at least 1 argument, got 0")
	}
	obj := args[0]

	if method, ok := getAttrHelper(obj, mc.name); ok {
		// Call the method with the stored args
		return objects.CallFunction(method, mc.args...)
	}
	return objects.NewAttributeError("'%s' object has no attribute '%s'", obj.Type(), mc.name)
}

// join is a helper to join string slices
func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// CreateOperatorModule creates the operator module
func CreateOperatorModule() *objects.Module {
	module := &objects.Module{
		Name:   "operator",
		Fields: make(map[string]objects.Object),
	}

	// operator.itemgetter
	module.Fields["itemgetter"] = &objects.Builtin{
		Name: "operator.itemgetter",
		Fn: func(args ...objects.Object) objects.Object {
			return &itemGetter{items: args}
		},
	}

	// operator.attrgetter
	module.Fields["attrgetter"] = &objects.Builtin{
		Name: "operator.attrgetter",
		Fn: func(args ...objects.Object) objects.Object {
			attrs := make([]string, len(args))
			for i, arg := range args {
				if s, ok := arg.(*objects.String); ok {
					attrs[i] = s.Value
				} else {
					return objects.NewTypeError("attrgetter requires string arguments")
				}
			}
			return &attrGetter{attrs: attrs}
		},
	}

	// operator.methodcaller
	module.Fields["methodcaller"] = &objects.Builtin{
		Name: "operator.methodcaller",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("methodcaller requires at least a method name")
			}
			name, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("methodcaller first argument must be a string")
			}
			return &methodCaller{name: name.Value, args: args[1:]}
		},
	}

	return module
}
