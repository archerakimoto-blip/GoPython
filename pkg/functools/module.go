package functools

import (
	"container/list"
	"fmt"

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

// --- LRU Cache types ---

// lruCacheEntry holds a cached result and its list element for LRU ordering
type lruCacheEntry struct {
	Key   string
	Args  []objects.Object
	Value objects.Object
}

// LRUCache represents an LRU cached function wrapper
type LRUCache struct {
	Func      objects.Object
	MaxSize   int // 0 means no caching, -1 means unlimited
	Cache     map[string]*list.Element
	Order     *list.List
	Hits      int64
	Misses    int64
}

func (lc *LRUCache) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (lc *LRUCache) Inspect() string {
	return fmt.Sprintf("functools.lru_cache(maxsize=%d, currsize=%d)", lc.MaxSize, len(lc.Cache))
}

func (lc *LRUCache) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__call__":
		return &objects.Builtin{
			Name: "lru_cache.__call__",
			Fn: func(args ...objects.Object) objects.Object {
				// maxsize=0 means no caching
				if lc.MaxSize == 0 {
					return objects.CallFunction(lc.Func, args...)
				}

				key := lruCacheKey(args)

				// Check cache
				if elem, ok := lc.Cache[key]; ok {
					lc.Order.MoveToFront(elem)
					lc.Hits++
					return elem.Value.(*lruCacheEntry).Value
				}

				// Cache miss
				lc.Misses++
				result := objects.CallFunction(lc.Func, args...)
				if result.Type() == objects.ERROR_OBJ {
					return result
				}

				// Store in cache
				entry := &lruCacheEntry{
					Key:   key,
					Args:  args,
					Value: result,
				}
				elem := lc.Order.PushFront(entry)
				lc.Cache[key] = elem

				// Evict if over capacity (MaxSize == -1 means unlimited)
				if lc.MaxSize > 0 && lc.Order.Len() > lc.MaxSize {
					oldest := lc.Order.Back()
					if oldest != nil {
						oldEntry := lc.Order.Remove(oldest).(*lruCacheEntry)
						delete(lc.Cache, oldEntry.Key)
					}
				}

				return result
			},
		}, true
	case "cache_info":
		return &objects.Builtin{
			Name: "lru_cache.cache_info",
			Fn: func(args ...objects.Object) objects.Object {
				maxsize := lc.MaxSize
				if maxsize == -1 {
					maxsize = 0 // represent None as 0 for display
				}
				return &objects.String{
					Value: fmt.Sprintf("CacheInfo(hits=%d, misses=%d, maxsize=%d, currsize=%d)",
						lc.Hits, lc.Misses, maxsize, len(lc.Cache)),
				}
			},
		}, true
	case "cache_clear":
		return &objects.Builtin{
			Name: "lru_cache.cache_clear",
			Fn: func(args ...objects.Object) objects.Object {
				lc.Cache = make(map[string]*list.Element)
				lc.Order.Init()
				lc.Hits = 0
				lc.Misses = 0
				return objects.None_
			},
		}, true
	}
	return nil, false
}

func lruCacheKey(args []objects.Object) string {
	key := ""
	for _, arg := range args {
		key += arg.Inspect() + "|"
	}
	return key
}

// --- Cached Property ---

// CachedProperty is a descriptor that computes a value once and caches it on the instance
type CachedProperty struct {
	Func objects.Object
	Attr string
}

func (cp *CachedProperty) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (cp *CachedProperty) Inspect() string {
	return "functools.cached_property"
}

// --- Single Dispatch ---

// SingleDispatch represents a generic function with type-based dispatch
type SingleDispatch struct {
	Func     objects.Object
	Registry map[string]objects.Object // type name -> implementation
}

func (sd *SingleDispatch) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (sd *SingleDispatch) Inspect() string {
	return "functools.singledispatch"
}

func (sd *SingleDispatch) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__call__":
		return &objects.Builtin{
			Name: "singledispatch.__call__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) == 0 {
					return objects.NewTypeError("singledispatch function requires at least 1 argument")
				}
				// Determine the type of the first argument
				typeName := getTypeName(args[0])
				if impl, ok := sd.Registry[typeName]; ok {
					return objects.CallFunction(impl, args...)
				}
				// Fall back to the default function
				return objects.CallFunction(sd.Func, args...)
			},
		}, true
	case "register":
		return &objects.Builtin{
			Name: "singledispatch.register",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("register() takes at least 1 argument")
				}
				// register(type, func) or register(func) as decorator
				if len(args) >= 2 {
					typeName := getTypeName(args[0])
					sd.Registry[typeName] = args[1]
					return args[1]
				}
				// Decorator form: register(func) - uses the function's first arg annotation
				// For simplicity, we treat single-arg register as returning a decorator
				// that registers the wrapped function for its type
				return &objects.Builtin{
					Name: "singledispatch.register_decorator",
					Fn: func(innerArgs ...objects.Object) objects.Object {
						if len(innerArgs) < 1 {
							return objects.NewTypeError("register decorator requires a function argument")
						}
						// Try to get the type from the function's __annotations__ or __module__
						// For simplicity, we register with the function name as key
						// In practice, this would need type annotation support
						if cls, ok := args[0].(*objects.Class); ok {
							sd.Registry[cls.Name] = innerArgs[0]
						} else if str, ok := args[0].(*objects.String); ok {
							sd.Registry[str.Value] = innerArgs[0]
						}
						return innerArgs[0]
					},
				}
			},
		}, true
	case "dispatch":
		return &objects.Builtin{
			Name: "singledispatch.dispatch",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("dispatch() takes at least 1 argument")
				}
				typeName := getTypeName(args[0])
				if impl, ok := sd.Registry[typeName]; ok {
					return impl
				}
				return sd.Func
			},
		}, true
	}
	return nil, false
}

func getTypeName(obj objects.Object) string {
	switch obj.(type) {
	case *objects.Integer:
		return "int"
	case *objects.Float:
		return "float"
	case *objects.String:
		return "str"
	case *objects.Boolean:
		return "bool"
	case *objects.List:
		return "list"
	case *objects.Dict:
		return "dict"
	case *objects.Tuple:
		return "tuple"
	case *objects.Set:
		return "set"
	case *objects.None:
		return "NoneType"
	case *objects.Instance:
		inst := obj.(*objects.Instance)
		if inst.Class != nil {
			return inst.Class.Name
		}
		return "object"
	case *objects.Class:
		return obj.(*objects.Class).Name
	default:
		return string(obj.Type())
	}
}

// --- Wrapper metadata helpers ---

// wrapperMetadata stores metadata copied from wrapped function
type wrapperMetadata struct {
	Wrapped    objects.Object
	WrappedDict *objects.Dict
	Name       string
	Doc        string
	Module     string
	Qualname   string
}

func getFuncAttr(fn objects.Object, attr string) objects.Object {
	if ag, ok := fn.(objects.AttributeGetter); ok {
		if val, found := ag.GetAttr(attr); found {
			return val
		}
	}
	switch f := fn.(type) {
	case *objects.Builtin:
		switch attr {
		case "__name__":
			return &objects.String{Value: f.Name}
		case "__module__":
			return &objects.String{Value: "builtins"}
		}
	case *objects.Closure:
		switch attr {
		case "__name__":
			return &objects.String{Value: "closure"}
		case "__module__":
			return &objects.String{Value: "__main__"}
		}
	}
	return objects.None_
}

// updateWrapper copies metadata from wrapped to wrapper
func updateWrapper(wrapper, wrapped objects.Object) objects.Object {
	// Copy standard attributes from wrapped to wrapper
	attrs := []string{"__name__", "__doc__", "__module__", "__qualname__", "__annotations__"}
	for _, attr := range attrs {
		val := getFuncAttr(wrapped, attr)
		if val != objects.None_ {
			// Try to set the attribute on the wrapper
			if inst, ok := wrapper.(*objects.Instance); ok {
				if inst.Fields == nil {
					inst.Fields = make(map[string]objects.Object)
				}
				inst.Fields[attr] = val
			}
		}
	}

	// Copy __dict__ from wrapped to wrapper
	if wrappedAg, ok := wrapped.(objects.AttributeGetter); ok {
		if wrappedDict, found := wrappedAg.GetAttr("__dict__"); found {
			if dict, ok := wrappedDict.(*objects.Dict); ok {
				if inst, ok := wrapper.(*objects.Instance); ok {
					if inst.Fields == nil {
						inst.Fields = make(map[string]objects.Object)
					}
					for _, keyStr := range dict.KeyOrder {
						inst.Fields[keyStr] = dict.Pairs[keyStr]
					}
				}
			}
		}
	}

	// Set __wrapped__ on wrapper
	if inst, ok := wrapper.(*objects.Instance); ok {
		if inst.Fields == nil {
			inst.Fields = make(map[string]objects.Object)
		}
		inst.Fields["__wrapped__"] = wrapped
	}

	return wrapper
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

	// functools.update_wrapper(wrapper, wrapped)
	module.Fields["update_wrapper"] = &objects.Builtin{
		Name: "functools.update_wrapper",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("update_wrapper() takes at least 2 arguments")
			}
			wrapper := args[0]
			wrapped := args[1]
			return updateWrapper(wrapper, wrapped)
		},
	}

	// functools.wraps(wrapped)
	module.Fields["wraps"] = &objects.Builtin{
		Name: "functools.wraps",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("wraps() takes at least 1 argument")
			}
			wrapped := args[0]

			// Return a decorator function that takes a wrapper and updates it
			return &objects.Builtin{
				Name: "functools.wraps_decorator",
				Fn: func(innerArgs ...objects.Object) objects.Object {
					if len(innerArgs) < 1 {
						return objects.NewTypeError("wraps decorator requires a function argument")
					}
					wrapper := innerArgs[0]
					return updateWrapper(wrapper, wrapped)
				},
			}
		},
	}

	// functools.lru_cache(maxsize=128, typed=False)
	module.Fields["lru_cache"] = &objects.Builtin{
		Name: "functools.lru_cache",
		Fn: func(args ...objects.Object) objects.Object {
			maxSize := 128 // default

			// lru_cache() called without args - use defaults
			// lru_cache(maxsize) or lru_cache(maxsize=N, typed=False)
			if len(args) == 0 {
				// Return a decorator with defaults
				return &objects.Builtin{
					Name: "functools.lru_cache_decorator",
					Fn: func(decoratorArgs ...objects.Object) objects.Object {
						if len(decoratorArgs) < 1 {
							return objects.NewTypeError("lru_cache decorator requires a function argument")
						}
						return newLRUCache(decoratorArgs[0], maxSize)
					},
				}
			}

			// Check if first arg is a function (direct decoration: @lru_cache without parens)
			if _, ok := args[0].(*objects.Closure); ok {
				return newLRUCache(args[0], maxSize)
			}
			if _, ok := args[0].(*objects.Builtin); ok {
				return newLRUCache(args[0], maxSize)
			}

			// First arg is maxsize
			switch v := args[0].(type) {
			case *objects.Integer:
				maxSize = int(v.Value)
			case *objects.None:
				maxSize = -1 // unlimited
			default:
				return objects.NewTypeError("lru_cache() first argument must be an integer or None")
			}

			return &objects.Builtin{
				Name: "functools.lru_cache_decorator",
				Fn: func(decoratorArgs ...objects.Object) objects.Object {
					if len(decoratorArgs) < 1 {
						return objects.NewTypeError("lru_cache decorator requires a function argument")
					}
					return newLRUCache(decoratorArgs[0], maxSize)
				},
			}
		},
	}

	// functools.cached_property(func)
	module.Fields["cached_property"] = &objects.Builtin{
		Name: "functools.cached_property",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("cached_property() takes at least 1 argument")
			}
			return &CachedProperty{
				Func: args[0],
			}
		},
	}

	// functools.total_ordering(cls)
	module.Fields["total_ordering"] = &objects.Builtin{
		Name: "functools.total_ordering",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("total_ordering() takes at least 1 argument")
			}
			cls, ok := args[0].(*objects.Class)
			if !ok {
				return objects.NewTypeError("total_ordering() argument must be a class")
			}

			// Check which comparison method is defined and derive the others
			hasLt := hasClassMethod(cls, "__lt__")
			hasLe := hasClassMethod(cls, "__le__")
			hasGt := hasClassMethod(cls, "__gt__")
			hasGe := hasClassMethod(cls, "__ge__")

			// If __lt__ is defined, derive __le__, __gt__, __ge__
			if hasLt {
				if !hasLe {
					cls.Methods["__le__"] = &objects.Builtin{
						Name: cls.Name + ".__le__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__le__ requires 2 arguments")
							}
							ltResult := objects.CallFunction(cls.Methods["__lt__"], callArgs[0], callArgs[1])
							if ltResult.Type() == objects.ERROR_OBJ {
								return ltResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return orBooleans(ltResult, eqResult)
						},
					}
				}
				if !hasGt {
					cls.Methods["__gt__"] = &objects.Builtin{
						Name: cls.Name + ".__gt__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__gt__ requires 2 arguments")
							}
							ltResult := objects.CallFunction(cls.Methods["__lt__"], callArgs[1], callArgs[0])
							if ltResult.Type() == objects.ERROR_OBJ {
								return ltResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return andBooleans(ltResult, notBoolean(eqResult))
						},
					}
				}
				if !hasGe {
					cls.Methods["__ge__"] = &objects.Builtin{
						Name: cls.Name + ".__ge__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__ge__ requires 2 arguments")
							}
							ltResult := objects.CallFunction(cls.Methods["__lt__"], callArgs[0], callArgs[1])
							if ltResult.Type() == objects.ERROR_OBJ {
								return ltResult
							}
							return notBoolean(ltResult)
						},
					}
				}
			} else if hasLe {
				// If __le__ is defined, derive others
				if !hasLt {
					cls.Methods["__lt__"] = &objects.Builtin{
						Name: cls.Name + ".__lt__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__lt__ requires 2 arguments")
							}
							leResult := objects.CallFunction(cls.Methods["__le__"], callArgs[0], callArgs[1])
							if leResult.Type() == objects.ERROR_OBJ {
								return leResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return andBooleans(leResult, notBoolean(eqResult))
						},
					}
				}
				if !hasGt {
					cls.Methods["__gt__"] = &objects.Builtin{
						Name: cls.Name + ".__gt__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__gt__ requires 2 arguments")
							}
							return notBoolean(objects.CallFunction(cls.Methods["__le__"], callArgs[0], callArgs[1]))
						},
					}
				}
				if !hasGe {
					cls.Methods["__ge__"] = &objects.Builtin{
						Name: cls.Name + ".__ge__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__ge__ requires 2 arguments")
							}
							leResult := objects.CallFunction(cls.Methods["__le__"], callArgs[1], callArgs[0])
							if leResult.Type() == objects.ERROR_OBJ {
								return leResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return orBooleans(leResult, eqResult)
						},
					}
				}
			} else if hasGt {
				// If __gt__ is defined, derive others
				if !hasGe {
					cls.Methods["__ge__"] = &objects.Builtin{
						Name: cls.Name + ".__ge__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__ge__ requires 2 arguments")
							}
							gtResult := objects.CallFunction(cls.Methods["__gt__"], callArgs[0], callArgs[1])
							if gtResult.Type() == objects.ERROR_OBJ {
								return gtResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return orBooleans(gtResult, eqResult)
						},
					}
				}
				if !hasLt {
					cls.Methods["__lt__"] = &objects.Builtin{
						Name: cls.Name + ".__lt__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__lt__ requires 2 arguments")
							}
							gtResult := objects.CallFunction(cls.Methods["__gt__"], callArgs[1], callArgs[0])
							if gtResult.Type() == objects.ERROR_OBJ {
								return gtResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return andBooleans(gtResult, notBoolean(eqResult))
						},
					}
				}
				if !hasLe {
					cls.Methods["__le__"] = &objects.Builtin{
						Name: cls.Name + ".__le__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__le__ requires 2 arguments")
							}
							return notBoolean(objects.CallFunction(cls.Methods["__gt__"], callArgs[0], callArgs[1]))
						},
					}
				}
			} else if hasGe {
				// If __ge__ is defined, derive others
				if !hasGt {
					cls.Methods["__gt__"] = &objects.Builtin{
						Name: cls.Name + ".__gt__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__gt__ requires 2 arguments")
							}
							geResult := objects.CallFunction(cls.Methods["__ge__"], callArgs[0], callArgs[1])
							if geResult.Type() == objects.ERROR_OBJ {
								return geResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return andBooleans(geResult, notBoolean(eqResult))
						},
					}
				}
				if !hasLt {
					cls.Methods["__lt__"] = &objects.Builtin{
						Name: cls.Name + ".__lt__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__lt__ requires 2 arguments")
							}
							return notBoolean(objects.CallFunction(cls.Methods["__ge__"], callArgs[0], callArgs[1]))
						},
					}
				}
				if !hasLe {
					cls.Methods["__le__"] = &objects.Builtin{
						Name: cls.Name + ".__le__",
						Fn: func(callArgs ...objects.Object) objects.Object {
							if len(callArgs) < 2 {
								return objects.NewTypeError("__le__ requires 2 arguments")
							}
							geResult := objects.CallFunction(cls.Methods["__ge__"], callArgs[1], callArgs[0])
							if geResult.Type() == objects.ERROR_OBJ {
								return geResult
							}
							eqResult := objects.CallFunction(cls.Methods["__eq__"], callArgs[0], callArgs[1])
							if eqResult.Type() == objects.ERROR_OBJ {
								return eqResult
							}
							return orBooleans(geResult, eqResult)
						},
					}
				}
			}

			return cls
		},
	}

	// functools.singledispatch(func)
	module.Fields["singledispatch"] = &objects.Builtin{
		Name: "functools.singledispatch",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("singledispatch() takes at least 1 argument")
			}
			return &SingleDispatch{
				Func:     args[0],
				Registry: make(map[string]objects.Object),
			}
		},
	}

	return module
}

// --- Helper functions ---

func newLRUCache(fn objects.Object, maxSize int) *LRUCache {
	return &LRUCache{
		Func:    fn,
		MaxSize: maxSize,
		Cache:   make(map[string]*list.Element),
		Order:   list.New(),
		Hits:    0,
		Misses:  0,
	}
}

func hasClassMethod(cls *objects.Class, name string) bool {
	_, ok := cls.Methods[name]
	if !ok && cls.SuperClass != nil {
		return hasClassMethod(cls.SuperClass, name)
	}
	return ok
}

func toBool(obj objects.Object) bool {
	switch v := obj.(type) {
	case *objects.Boolean:
		return v.Value
	case *objects.Integer:
		return v.Value != 0
	case *objects.None:
		return false
	default:
		return true
	}
}

func notBoolean(obj objects.Object) objects.Object {
	b := toBool(obj)
	if b {
		return objects.False
	}
	return objects.True
}

func andBooleans(a, b objects.Object) objects.Object {
	if toBool(a) && toBool(b) {
		return objects.True
	}
	return objects.False
}

func orBooleans(a, b objects.Object) objects.Object {
	if toBool(a) || toBool(b) {
		return objects.True
	}
	return objects.False
}
