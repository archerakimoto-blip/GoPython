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

	return module
}
