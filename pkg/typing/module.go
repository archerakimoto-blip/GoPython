package typing

import (
	"fmt"
	"strings"

	"github.com/go-py/go-python/pkg/objects"
)

// TypingAlias represents a type hint from the typing module, e.g. List, Dict[str, int], Optional[int]
type TypingAlias struct {
	Name   string
	Args   []objects.Object
	Origin objects.Object
}

func (t *TypingAlias) Type() objects.ObjectType { return objects.INSTANCE_OBJ }

func (t *TypingAlias) Inspect() string {
	if len(t.Args) == 0 {
		return "typing." + t.Name
	}
	parts := make([]string, len(t.Args))
	for i, arg := range t.Args {
		parts[i] = arg.Inspect()
	}
	return "typing." + t.Name + "[" + strings.Join(parts, ", ") + "]"
}

func (t *TypingAlias) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__origin__":
		if t.Origin != nil {
			return t.Origin, true
		}
		return objects.None_, true
	case "__args__":
		if len(t.Args) == 0 {
			return &objects.Tuple{Elements: []objects.Object{}}, true
		}
		return &objects.Tuple{Elements: t.Args}, true
	case "__getitem__":
		return &objects.Builtin{
			Name: fmt.Sprintf("typing.%s.__getitem__", t.Name),
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return objects.NewTypeError("__getitem__ takes exactly 1 argument")
				}
				// Support both a single type arg and a tuple of type args
				var newArgs []objects.Object
				if tuple, ok := args[0].(*objects.Tuple); ok {
					newArgs = make([]objects.Object, len(tuple.Elements))
					copy(newArgs, tuple.Elements)
				} else {
					newArgs = []objects.Object{args[0]}
				}
				return &TypingAlias{
					Name:   t.Name,
					Args:   newArgs,
					Origin: t.Origin,
				}
			},
		}, true
	case "__class__":
		return &objects.Class{Name: "typing." + t.Name}, true
	case "__name__":
		return &objects.String{Value: t.Name}, true
	}
	return nil, false
}

// newTypingAlias creates an unparameterized typing alias (e.g. List without type args)
func newTypingAlias(name string, origin objects.Object) *TypingAlias {
	return &TypingAlias{
		Name:   name,
		Args:   nil,
		Origin: origin,
	}
}

// TypeVarInstance represents a typing.TypeVar object
type TypeVarInstance struct {
	Name          string
	Constraints   []objects.Object
	Bound         objects.Object
	Covariant     bool
	Contravariant bool
}

func (tv *TypeVarInstance) Type() objects.ObjectType { return objects.INSTANCE_OBJ }

func (tv *TypeVarInstance) Inspect() string {
	return "~" + tv.Name
}

func (tv *TypeVarInstance) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__name__":
		return &objects.String{Value: tv.Name}, true
	case "__constraints__":
		elements := make([]objects.Object, len(tv.Constraints))
		copy(elements, tv.Constraints)
		return &objects.Tuple{Elements: elements}, true
	case "__bound__":
		if tv.Bound != nil {
			return tv.Bound, true
		}
		return objects.None_, true
	case "__covariant__":
		if tv.Covariant {
			return objects.True, true
		}
		return objects.False, true
	case "__contravariant__":
		if tv.Contravariant {
			return objects.True, true
		}
		return objects.False, true
	}
	return nil, false
}

// ParamSpecInstance represents a typing.ParamSpec object
type ParamSpecInstance struct {
	Name string
}

func (ps *ParamSpecInstance) Type() objects.ObjectType { return objects.INSTANCE_OBJ }

func (ps *ParamSpecInstance) Inspect() string {
	return "typing.ParamSpec(" + ps.Name + ")"
}

func (ps *ParamSpecInstance) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__name__":
		return &objects.String{Value: ps.Name}, true
	}
	return nil, false
}

// ConcatenateInstance represents a typing.Concatenate object
type ConcatenateInstance struct {
	Args []objects.Object
}

func (c *ConcatenateInstance) Type() objects.ObjectType { return objects.INSTANCE_OBJ }

func (c *ConcatenateInstance) Inspect() string {
	parts := make([]string, len(c.Args))
	for i, arg := range c.Args {
		parts[i] = arg.Inspect()
	}
	return "typing.Concatenate[" + strings.Join(parts, ", ") + "]"
}

func (c *ConcatenateInstance) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__args__":
		return &objects.Tuple{Elements: c.Args}, true
	}
	return nil, false
}

// CreateTypingModule creates the typing module with basic type hints
func CreateTypingModule() *objects.Module {
	module := &objects.Module{
		Name:   "typing",
		Fields: make(map[string]objects.Object),
	}

	// Origin classes representing the base Python types
	listOrigin := &objects.Class{Name: "list"}
	dictOrigin := &objects.Class{Name: "dict"}
	tupleOrigin := &objects.Class{Name: "tuple"}
	setOrigin := &objects.Class{Name: "set"}

	// List - generic type alias for list
	module.Fields["List"] = newTypingAlias("List", listOrigin)

	// Dict - generic type alias for dict
	module.Fields["Dict"] = newTypingAlias("Dict", dictOrigin)

	// Tuple - generic type alias for tuple
	module.Fields["Tuple"] = newTypingAlias("Tuple", tupleOrigin)

	// Set - generic type alias for set
	module.Fields["Set"] = newTypingAlias("Set", setOrigin)

	// Optional - Optional[X] is equivalent to Union[X, None]
	module.Fields["Optional"] = newTypingAlias("Optional", nil)

	// Union - Union type, e.g. Union[int, str]
	module.Fields["Union"] = newTypingAlias("Union", nil)

	// Any - special type indicating any type
	module.Fields["Any"] = &TypingAlias{
		Name:   "Any",
		Args:   nil,
		Origin: nil,
	}

	// Callable - callable type hint
	module.Fields["Callable"] = newTypingAlias("Callable", nil)

	// TypeVar - creates a type variable for generic types
	module.Fields["TypeVar"] = &objects.Builtin{
		Name: "typing.TypeVar",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("TypeVar() takes at least 1 argument")
			}
			nameObj, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("TypeVar() first argument must be a string")
			}

			tv := &TypeVarInstance{
				Name:        nameObj.Value,
				Constraints: []objects.Object{},
			}

			// Remaining positional args are constraints
			for i := 1; i < len(args); i++ {
				tv.Constraints = append(tv.Constraints, args[i])
			}

			return tv
		},
	}

	// Generic - base class for generic types
	module.Fields["Generic"] = &TypingAlias{
		Name:   "Generic",
		Args:   nil,
		Origin: nil,
	}

	// Protocol - base class for structural subtyping
	module.Fields["Protocol"] = &TypingAlias{
		Name:   "Protocol",
		Args:   nil,
		Origin: nil,
	}

	// Literal - type representing a specific set of literal values
	module.Fields["Literal"] = newTypingAlias("Literal", nil)

	// Final - type indicating a value that cannot be reassigned
	module.Fields["Final"] = newTypingAlias("Final", nil)

	// TypeAlias - marker for type aliases
	module.Fields["TypeAlias"] = &TypingAlias{
		Name:   "TypeAlias",
		Args:   nil,
		Origin: nil,
	}

	// ParamSpec - parameter specification type variable
	module.Fields["ParamSpec"] = &objects.Builtin{
		Name: "typing.ParamSpec",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("ParamSpec() takes at least 1 argument")
			}
			nameObj, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("ParamSpec() first argument must be a string")
			}
			return &ParamSpecInstance{Name: nameObj.Value}
		},
	}

	// Concatenate - concatenate parameter specifications
	module.Fields["Concatenate"] = &objects.Builtin{
		Name: "typing.Concatenate",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("Concatenate() takes at least 2 arguments")
			}
			return &ConcatenateInstance{Args: args}
		},
	}

	return module
}
