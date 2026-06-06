package collections_abc

import (
	"github.com/go-py/go-python/pkg/objects"
)

// AbstractBaseClass represents a Python ABC from collections.abc
type AbstractBaseClass struct {
	Name              string
	AbstractMethods   []string
	ParentABCs        []*AbstractBaseClass
}

func (abc *AbstractBaseClass) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (abc *AbstractBaseClass) Inspect() string {
	return "<class 'collections.abc." + abc.Name + "'>"
}

func (abc *AbstractBaseClass) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__abstractmethods__":
		// Collect all abstract method names from this ABC and its parents
		allMethods := abc.allAbstractMethods()
		elements := make([]objects.Object, len(allMethods))
		for i, m := range allMethods {
			elements[i] = &objects.String{Value: m}
		}
		return &objects.Tuple{Elements: elements}, true
	case "__subclasshook__":
		return &objects.Builtin{
			Name: "collections.abc." + abc.Name + ".__subclasshook__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__subclasshook__() takes at least 1 argument")
				}
				if isinstance_of_abc(args[0], abc) {
					return objects.True
				}
				return objects.False
			},
		}, true
	case "__name__":
		return &objects.String{Value: abc.Name}, true
	}

	// Return abstract method names as Builtin stubs
	for _, method := range abc.AbstractMethods {
		if name == method {
			return &objects.Builtin{
				Name: "collections.abc." + abc.Name + "." + method,
				Fn: func(args ...objects.Object) objects.Object {
					return objects.NewTypeError("'%s' is an abstract method of %s", method, abc.Name)
				},
			}, true
		}
	}

	return nil, false
}

// allAbstractMethods returns the deduplicated set of abstract method names
// from this ABC and all parent ABCs.
func (abc *AbstractBaseClass) allAbstractMethods() []string {
	seen := make(map[string]bool)
	var result []string

	// Add methods from parent ABCs first
	for _, parent := range abc.ParentABCs {
		for _, m := range parent.allAbstractMethods() {
			if !seen[m] {
				seen[m] = true
				result = append(result, m)
			}
		}
	}

	// Add this ABC's own methods
	for _, m := range abc.AbstractMethods {
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}

	return result
}

// isinstance_of_abc checks if an object has all the required methods
// for a given ABC by checking its GetAttr interface.
func isinstance_of_abc(obj objects.Object, abc *AbstractBaseClass) bool {
	methods := abc.allAbstractMethods()
	getter, ok := obj.(objects.AttributeGetter)
	if !ok {
		return false
	}
	for _, method := range methods {
		if _, found := getter.GetAttr(method); !found {
			return false
		}
	}
	return true
}

// IsInstanceOfABC is the exported helper function that checks if an object
// has the required methods for a given ABC.
func IsInstanceOfABC(obj objects.Object, abc *AbstractBaseClass) bool {
	return isinstance_of_abc(obj, abc)
}

// Pre-defined ABC instances

var (
	// Iterable ABC with __iter__ abstract method
	IterableABC = &AbstractBaseClass{
		Name:            "Iterable",
		AbstractMethods: []string{"__iter__"},
		ParentABCs:      nil,
	}

	// Sequence ABC inheriting from Iterable, with __getitem__ and __len__
	SequenceABC = &AbstractBaseClass{
		Name:            "Sequence",
		AbstractMethods: []string{"__getitem__", "__len__"},
		ParentABCs:      []*AbstractBaseClass{IterableABC},
	}

	// Mapping ABC with __getitem__, __len__, __iter__
	MappingABC = &AbstractBaseClass{
		Name:            "Mapping",
		AbstractMethods: []string{"__getitem__", "__len__", "__iter__"},
		ParentABCs:      nil,
	}

	// Set ABC with __contains__, __len__, __iter__
	SetABC = &AbstractBaseClass{
		Name:            "Set",
		AbstractMethods: []string{"__contains__", "__len__", "__iter__"},
		ParentABCs:      nil,
	}

	// Callable ABC with __call__
	CallableABC = &AbstractBaseClass{
		Name:            "Callable",
		AbstractMethods: []string{"__call__"},
		ParentABCs:      nil,
	}
)

// CreateCollectionsABCModule creates the collections.abc module
func CreateCollectionsABCModule() *objects.Module {
	module := &objects.Module{
		Name:   "collections.abc",
		Fields: make(map[string]objects.Object),
	}

	module.Fields["Iterable"] = IterableABC
	module.Fields["Sequence"] = SequenceABC
	module.Fields["Mapping"] = MappingABC
	module.Fields["Set"] = SetABC
	module.Fields["Callable"] = CallableABC

	// Helper function: isinstance_of_abc(obj, abc_name)
	module.Fields["isinstance_of_abc"] = &objects.Builtin{
		Name: "collections.abc.isinstance_of_abc",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("isinstance_of_abc() takes at least 2 arguments")
			}
			abcObj, ok := args[1].(*AbstractBaseClass)
			if !ok {
				// Try by name
				if name, ok := args[1].(*objects.String); ok {
					switch name.Value {
					case "Iterable":
						abcObj = IterableABC
					case "Sequence":
						abcObj = SequenceABC
					case "Mapping":
						abcObj = MappingABC
					case "Set":
						abcObj = SetABC
					case "Callable":
						abcObj = CallableABC
					default:
						return objects.NewTypeError("unknown ABC: %s", name.Value)
					}
				} else {
					return objects.NewTypeError("isinstance_of_abc() second argument must be an ABC or string")
				}
			}
			if isinstance_of_abc(args[0], abcObj) {
				return objects.True
			}
			return objects.False
		},
	}

	return module
}
