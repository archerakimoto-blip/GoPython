package collections

import (
	"sort"

	"github.com/go-py/go-python/pkg/objects"
)

// deque is a double-ended queue
type deque struct {
	items []objects.Object
}

func newDeque() *deque {
	return &deque{
		items: make([]objects.Object, 0),
	}
}

func (d *deque) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (d *deque) Inspect() string {
	result := "deque(["
	for i, item := range d.items {
		if i > 0 {
			result += ", "
		}
		result += item.Inspect()
	}
	result += "])"
	return result
}

func (d *deque) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__getitem__":
		return &objects.Builtin{
			Name: "deque.__getitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__getitem__() takes at least 1 argument")
				}
				idx, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("deque indices must be integers, not %s", args[0].Type())
				}
				i := int(idx.Value)
				length := len(d.items)
				if i < 0 {
					i = length + i
				}
				if i < 0 || i >= length {
					return objects.NewIndexError("deque index out of range")
				}
				return d.items[i]
			},
		}, true
	case "__setitem__":
		return &objects.Builtin{
			Name: "deque.__setitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return objects.NewTypeError("__setitem__() takes at least 2 arguments")
				}
				idx, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("deque indices must be integers, not %s", args[0].Type())
				}
				i := int(idx.Value)
				length := len(d.items)
				if i < 0 {
					i = length + i
				}
				if i < 0 || i >= length {
					return objects.NewIndexError("deque index out of range")
				}
				d.items[i] = args[1]
				return objects.None_
			},
		}, true
	case "append":
		return &objects.Builtin{
			Name: "deque.append",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("append() takes at least 1 argument")
				}
				d.items = append(d.items, args[0])
				return objects.None_
			},
		}, true
	case "appendleft":
		return &objects.Builtin{
			Name: "deque.appendleft",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("appendleft() takes at least 1 argument")
				}
				d.items = append([]objects.Object{args[0]}, d.items...)
				return objects.None_
			},
		}, true
	case "pop":
		return &objects.Builtin{
			Name: "deque.pop",
			Fn: func(args ...objects.Object) objects.Object {
				if len(d.items) == 0 {
					return objects.NewIndexError("pop from empty deque")
				}
				item := d.items[len(d.items)-1]
				d.items = d.items[:len(d.items)-1]
				return item
			},
		}, true
	case "popleft":
		return &objects.Builtin{
			Name: "deque.popleft",
			Fn: func(args ...objects.Object) objects.Object {
				if len(d.items) == 0 {
					return objects.NewIndexError("pop from empty deque")
				}
				item := d.items[0]
				d.items = d.items[1:]
				return item
			},
		}, true
	case "clear":
		return &objects.Builtin{
			Name: "deque.clear",
			Fn: func(args ...objects.Object) objects.Object {
				d.items = make([]objects.Object, 0)
				return objects.None_
			},
		}, true
	case "count":
		return &objects.Builtin{
			Name: "deque.count",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("count() takes at least 1 argument")
				}
				count := 0
				for _, item := range d.items {
					if objects.Equal(item, args[0]) {
						count++
					}
				}
				return &objects.Integer{Value: int64(count)}
			},
		}, true
	case "extend":
		return &objects.Builtin{
			Name: "deque.extend",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("extend() takes at least 1 argument")
				}
				var items []objects.Object
				switch iter := args[0].(type) {
				case *objects.List:
					items = iter.Elements
				case *objects.Tuple:
					items = iter.Elements
				case *deque:
					items = iter.items
				default:
					return objects.NewTypeError("extend() argument must be iterable")
				}
				d.items = append(d.items, items...)
				return objects.None_
			},
		}, true
	case "extendleft":
		return &objects.Builtin{
			Name: "deque.extendleft",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("extendleft() takes at least 1 argument")
				}
				var items []objects.Object
				switch iter := args[0].(type) {
				case *objects.List:
					items = iter.Elements
				case *objects.Tuple:
					items = iter.Elements
				case *deque:
					items = iter.items
				default:
					return objects.NewTypeError("extendleft() argument must be iterable")
				}
				// Reverse and prepend
				for i := len(items) - 1; i >= 0; i-- {
					d.items = append([]objects.Object{items[i]}, d.items...)
				}
				return objects.None_
			},
		}, true
	case "index":
		return &objects.Builtin{
			Name: "deque.index",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("index() takes at least 1 argument")
				}
				start := 0
				end := len(d.items)
				if len(args) >= 2 {
					if i, ok := args[1].(*objects.Integer); ok {
						start = int(i.Value)
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*objects.Integer); ok {
						end = int(i.Value)
					}
				}
				for i := start; i < end && i < len(d.items); i++ {
					if objects.Equal(d.items[i], args[0]) {
						return &objects.Integer{Value: int64(i)}
					}
				}
				return objects.NewValueError("deque.index(x): x not in deque")
			},
		}, true
	case "insert":
		return &objects.Builtin{
			Name: "deque.insert",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return objects.NewTypeError("insert() takes at least 2 arguments")
				}
				idx, ok := args[0].(*objects.Integer)
				if !ok {
					return objects.NewTypeError("insert() argument 1 must be int")
				}
				i := int(idx.Value)
			if i < 0 {
				i = len(d.items) + i
				if i < 0 {
					i = 0
				}
			}
			if i > len(d.items) {
				i = len(d.items)
			}
				d.items = append(d.items, objects.None_)
				copy(d.items[i+1:], d.items[i:])
				d.items[i] = args[1]
				return objects.None_
			},
		}, true
	case "remove":
		return &objects.Builtin{
			Name: "deque.remove",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("remove() takes at least 1 argument")
				}
				for i, item := range d.items {
					if objects.Equal(item, args[0]) {
						d.items = append(d.items[:i], d.items[i+1:]...)
						return objects.None_
					}
				}
				return objects.NewValueError("deque.remove(x): x not in deque")
			},
		}, true
	case "reverse":
		return &objects.Builtin{
			Name: "deque.reverse",
			Fn: func(args ...objects.Object) objects.Object {
				for i, j := 0, len(d.items)-1; i < j; i, j = i+1, j-1 {
					d.items[i], d.items[j] = d.items[j], d.items[i]
				}
				return objects.None_
			},
		}, true
	case "rotate":
		return &objects.Builtin{
			Name: "deque.rotate",
			Fn: func(args ...objects.Object) objects.Object {
				n := 1
				if len(args) >= 1 {
					if i, ok := args[0].(*objects.Integer); ok {
						n = int(i.Value)
					}
				}
				length := len(d.items)
				if length == 0 {
					return objects.None_
				}
				n = n % length
				if n < 0 {
					n += length
				}
				if n == 0 {
					return objects.None_
				}
				// Rotate right by n: last n elements move to front
				mid := length - n
				rotated := make([]objects.Object, length)
				copy(rotated[:n], d.items[mid:])
				copy(rotated[n:], d.items[:mid])
				d.items = rotated
				return objects.None_
			},
		}, true
	case "__contains__":
		return &objects.Builtin{
			Name: "deque.__contains__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__contains__() takes at least 1 argument")
				}
				for _, item := range d.items {
					if objects.Equal(item, args[0]) {
						return &objects.Boolean{Value: true}
					}
				}
				return &objects.Boolean{Value: false}
			},
		}, true
	case "copy":
		return &objects.Builtin{
			Name: "deque.copy",
			Fn: func(args ...objects.Object) objects.Object {
				newD := newDeque()
				newD.items = append(newD.items, d.items...)
				return newD
			},
		}, true
	case "maxlen":
		return &objects.Builtin{
			Name: "deque.maxlen",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.None_
			},
		}, true
	case "__len__":
		return &objects.Builtin{
			Name: "deque.__len__",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.Integer{Value: int64(len(d.items))}
			},
		}, true
	}
	return nil, false
}

// OrderedDict is a dict that maintains insertion order
type OrderedDict struct {
	*objects.Instance
	Dict     *objects.Dict
	KeyOrder []string
}

func newOrderedDict() *OrderedDict {
	return &OrderedDict{
		Instance: &objects.Instance{Fields: make(map[string]objects.Object)},
		Dict:     objects.NewDict(),
		KeyOrder: make([]string, 0),
	}
}

func (od *OrderedDict) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (od *OrderedDict) Inspect() string {
	return od.Dict.Inspect()
}

func (od *OrderedDict) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "keys":
		return &objects.Builtin{
			Name: "OrderedDict.keys",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictKeys(od.Dict)
			},
		}, true
	case "values":
		return &objects.Builtin{
			Name: "OrderedDict.values",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictValues(od.Dict)
			},
		}, true
	case "items":
		return &objects.Builtin{
			Name: "OrderedDict.items",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictItems(od.Dict)
			},
		}, true
	case "get":
		return &objects.Builtin{
			Name: "OrderedDict.get",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("get() takes at least 1 argument")
				}
				var defaultVal objects.Object = objects.None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				if val, ok := od.Dict.Get(args[0]); ok {
					return val
				}
				return defaultVal
			},
		}, true
	case "pop":
		return &objects.Builtin{
			Name: "OrderedDict.pop",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("pop() takes at least 1 argument")
				}
				var defaultVal objects.Object = objects.None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				if val, ok := od.Dict.Get(args[0]); ok {
					od.Dict.Delete(args[0])
					// Remove from KeyOrder
					for i, k := range od.KeyOrder {
						if k == od.Dict.HashKey(args[0]) {
							od.KeyOrder = append(od.KeyOrder[:i], od.KeyOrder[i+1:]...)
							break
						}
					}
					return val
				}
				return defaultVal
			},
		}, true
	case "popitem":
		return &objects.Builtin{
			Name: "OrderedDict.popitem",
			Fn: func(args ...objects.Object) objects.Object {
				if len(od.KeyOrder) == 0 {
					return objects.NewKeyError("dictionary is empty")
				}
				last := true
				if len(args) >= 1 {
					if b, ok := args[0].(*objects.Boolean); ok {
						last = b.Value
					}
				}
				var idx int
				if last {
					idx = len(od.KeyOrder) - 1
				} else {
					idx = 0
				}
				keyStr := od.KeyOrder[idx]
				key := od.Dict.Keys[keyStr]
				val, _ := od.Dict.Get(key)
				od.Dict.Delete(key)
				od.KeyOrder = append(od.KeyOrder[:idx], od.KeyOrder[idx+1:]...)
				return &objects.Tuple{Elements: []objects.Object{key, val}}
			},
		}, true
	case "move_to_end":
		return &objects.Builtin{
			Name: "OrderedDict.move_to_end",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("move_to_end() takes at least 1 argument")
				}
				last := true
				if len(args) >= 2 {
					if b, ok := args[1].(*objects.Boolean); ok {
						last = b.Value
					}
				}
				keyStr := od.Dict.HashKey(args[0])
				found := false
				for i, k := range od.KeyOrder {
					if k == keyStr {
						found = true
						// Remove from current position
						od.KeyOrder = append(od.KeyOrder[:i], od.KeyOrder[i+1:]...)
						if last {
							od.KeyOrder = append(od.KeyOrder, keyStr)
						} else {
							od.KeyOrder = append([]string{keyStr}, od.KeyOrder...)
						}
						break
					}
				}
				if !found {
					return objects.NewKeyError("key not found")
				}
				return objects.None_
			},
		}, true
	case "clear":
		return &objects.Builtin{
			Name: "OrderedDict.clear",
			Fn: func(args ...objects.Object) objects.Object {
				od.Dict = objects.NewDict()
				od.KeyOrder = make([]string, 0)
				return objects.None_
			},
		}, true
	case "update":
		return &objects.Builtin{
			Name: "OrderedDict.update",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.None_
				}
				switch other := args[0].(type) {
			case *objects.Dict:
				for keyStr, val := range other.Pairs {
					key := other.Keys[keyStr]
					if _, exists := od.Dict.Pairs[keyStr]; !exists {
						od.KeyOrder = append(od.KeyOrder, keyStr)
					}
					od.Dict.Set(key, val)
				}
			case *OrderedDict:
				for _, keyStr := range other.KeyOrder {
					key := other.Dict.Keys[keyStr]
					val, _ := other.Dict.Get(key)
					if _, exists := od.Dict.Pairs[keyStr]; !exists {
						od.KeyOrder = append(od.KeyOrder, keyStr)
					}
					od.Dict.Set(key, val)
				}
			}
				return objects.None_
			},
		}, true
	case "setdefault":
		return &objects.Builtin{
			Name: "OrderedDict.setdefault",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("setdefault() takes at least 1 argument")
				}
				var defaultVal objects.Object = objects.None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				if val, ok := od.Dict.Get(args[0]); ok {
					return val
				}
				od.Dict.Set(args[0], defaultVal)
				od.KeyOrder = append(od.KeyOrder, od.Dict.HashKey(args[0]))
				return defaultVal
			},
		}, true
	case "__len__":
		return &objects.Builtin{
			Name: "OrderedDict.__len__",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.Integer{Value: int64(len(od.KeyOrder))}
			},
		}, true
	case "__getitem__":
		return &objects.Builtin{
			Name: "OrderedDict.__getitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__getitem__() takes at least 1 argument")
				}
				if val, ok := od.Dict.Get(args[0]); ok {
					return val
				}
				return objects.NewKeyError("%s", args[0].Inspect())
			},
		}, true
	case "__setitem__":
		return &objects.Builtin{
			Name: "OrderedDict.__setitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return objects.NewTypeError("__setitem__() takes at least 2 arguments")
				}
				keyStr := od.Dict.HashKey(args[0])
				if _, exists := od.Dict.Pairs[keyStr]; !exists {
					od.KeyOrder = append(od.KeyOrder, keyStr)
				}
				od.Dict.Set(args[0], args[1])
				return objects.None_
			},
		}, true
	}
	return nil, false
}

// Counter is a dict subclass for counting hashable objects
type Counter struct {
	*objects.Instance
	Dict     *objects.Dict
}

func newCounter() *Counter {
	return &Counter{
		Instance: &objects.Instance{Fields: make(map[string]objects.Object)},
		Dict:     objects.NewDict(),
	}
}

func (c *Counter) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (c *Counter) Inspect() string {
	result := "Counter({"
	first := true
	for _, keyStr := range c.Dict.KeyOrder {
		val := c.Dict.Pairs[keyStr]
		if val.(*objects.Integer).Value != 0 {
			if !first {
				result += ", "
			}
			first = false
			key := c.Dict.Keys[keyStr]
			result += key.Inspect() + ": " + val.Inspect()
		}
	}
	result += "})"
	return result
}

func (c *Counter) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "keys":
		return &objects.Builtin{
			Name: "Counter.keys",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictKeys(c.Dict)
			},
		}, true
	case "values":
		return &objects.Builtin{
			Name: "Counter.values",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictValues(c.Dict)
			},
		}, true
	case "items":
		return &objects.Builtin{
			Name: "Counter.items",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictItems(c.Dict)
			},
		}, true
	case "get":
		return &objects.Builtin{
			Name: "Counter.get",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("get() takes at least 1 argument")
				}
				defaultVal := &objects.Integer{Value: 0}
				if len(args) >= 2 {
					if i, ok := args[1].(*objects.Integer); ok {
						defaultVal = i
					}
				}
				if val, ok := c.Dict.Get(args[0]); ok {
					return val
				}
				return defaultVal
			},
		}, true
	case "update":
		return &objects.Builtin{
			Name: "Counter.update",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.None_
				}
				switch other := args[0].(type) {
				case *objects.List:
					for _, elem := range other.Elements {
						c.increment(elem)
					}
				case *Counter:
					for _, keyStr := range other.Dict.KeyOrder {
						key := other.Dict.Keys[keyStr]
						val := other.Dict.Pairs[keyStr]
						c.incrementBy(key, val.(*objects.Integer).Value)
					}
				case *objects.Dict:
					for _, keyStr := range other.KeyOrder {
						key := other.Keys[keyStr]
						val := other.Pairs[keyStr]
						c.incrementBy(key, val.(*objects.Integer).Value)
					}
				}
				return objects.None_
			},
		}, true
	case "most_common":
		return &objects.Builtin{
			Name: "Counter.most_common",
			Fn: func(args ...objects.Object) objects.Object {
				n := -1
				if len(args) >= 1 {
					if i, ok := args[0].(*objects.Integer); ok {
						n = int(i.Value)
					}
				}
				type kv struct {
					key   objects.Object
					value int64
				}
				pairs := make([]kv, 0)
				for _, keyStr := range c.Dict.KeyOrder {
					val := c.Dict.Pairs[keyStr]
					if val.(*objects.Integer).Value > 0 {
						key := c.Dict.Keys[keyStr]
						pairs = append(pairs, kv{key, val.(*objects.Integer).Value})
					}
				}
				sort.Slice(pairs, func(i, j int) bool {
					if pairs[i].value != pairs[j].value {
						return pairs[i].value > pairs[j].value
					}
					return pairs[i].key.Inspect() < pairs[j].key.Inspect()
				})
				if n >= 0 && n < len(pairs) {
					pairs = pairs[:n]
				}
				result := make([]objects.Object, len(pairs))
				for i, p := range pairs {
					result[i] = &objects.Tuple{Elements: []objects.Object{p.key, &objects.Integer{Value: p.value}}}
				}
				return &objects.List{Elements: result}
			},
		}, true
	case "elements":
		return &objects.Builtin{
			Name: "Counter.elements",
			Fn: func(args ...objects.Object) objects.Object {
				type kv struct {
					key   objects.Object
					value int64
				}
				items := make([]kv, 0)
				for _, keyStr := range c.Dict.KeyOrder {
					val := c.Dict.Pairs[keyStr]
					if val.(*objects.Integer).Value > 0 {
						key := c.Dict.Keys[keyStr]
						items = append(items, kv{key, val.(*objects.Integer).Value})
					}
				}
				result := make([]objects.Object, 0)
				for _, item := range items {
					for i := int64(0); i < item.value; i++ {
						result = append(result, item.key)
					}
				}
				return &objects.List{Elements: result}
			},
		}, true
	case "subtract":
		return &objects.Builtin{
			Name: "Counter.subtract",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.None_
				}
				switch other := args[0].(type) {
				case *objects.List:
					for _, elem := range other.Elements {
						c.incrementBy(elem, -1)
					}
				case *Counter:
					for _, keyStr := range other.Dict.KeyOrder {
						key := other.Dict.Keys[keyStr]
						val := other.Dict.Pairs[keyStr]
						c.incrementBy(key, -val.(*objects.Integer).Value)
					}
				}
				return objects.None_
			},
		}, true
	case "clear":
		return &objects.Builtin{
			Name: "Counter.clear",
			Fn: func(args ...objects.Object) objects.Object {
				c.Dict = objects.NewDict()
				return objects.None_
			},
		}, true
	case "copy":
		return &objects.Builtin{
			Name: "Counter.copy",
			Fn: func(args ...objects.Object) objects.Object {
				newCounter := newCounter()
				for _, keyStr := range c.Dict.KeyOrder {
					key := c.Dict.Keys[keyStr]
					val := c.Dict.Pairs[keyStr]
					newCounter.Dict.Set(key, val)
				}
				return newCounter
			},
		}, true
	case "__len__":
		return &objects.Builtin{
			Name: "Counter.__len__",
			Fn: func(args ...objects.Object) objects.Object {
				count := 0
				for _, keyStr := range c.Dict.KeyOrder {
					if c.Dict.Pairs[keyStr].(*objects.Integer).Value > 0 {
						count++
					}
				}
				return &objects.Integer{Value: int64(count)}
			},
		}, true
	case "__getitem__":
		return &objects.Builtin{
			Name: "Counter.__getitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__getitem__() takes at least 1 argument")
				}
				if val, ok := c.Dict.Get(args[0]); ok {
					return val
				}
				return &objects.Integer{Value: 0}
			},
		}, true
	case "__setitem__":
		return &objects.Builtin{
			Name: "Counter.__setitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return objects.NewTypeError("__setitem__() takes at least 2 arguments")
				}
				c.Dict.Set(args[0], args[1])
				return objects.None_
			},
		}, true
	}
	return nil, false
}

func (c *Counter) increment(elem objects.Object) {
	c.incrementBy(elem, 1)
}

func (c *Counter) incrementBy(elem objects.Object, n int64) {
	if val, ok := c.Dict.Get(elem); ok {
		c.Dict.Set(elem, &objects.Integer{Value: val.(*objects.Integer).Value + n})
	} else {
		c.Dict.Set(elem, &objects.Integer{Value: n})
	}
}

// defaultdict is a dict subclass that calls a factory function to supply missing values
type defaultdict struct {
	*objects.Instance
	Dict    *objects.Dict
	Factory objects.Object
}

func newDefaultDict(factory objects.Object) *defaultdict {
	return &defaultdict{
		Instance: &objects.Instance{Fields: make(map[string]objects.Object)},
		Dict:     objects.NewDict(),
		Factory:  factory,
	}
}

func (dd *defaultdict) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (dd *defaultdict) Inspect() string {
	return dd.Dict.Inspect()
}

func (dd *defaultdict) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "keys":
		return &objects.Builtin{
			Name: "defaultdict.keys",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictKeys(dd.Dict)
			},
		}, true
	case "values":
		return &objects.Builtin{
			Name: "defaultdict.values",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictValues(dd.Dict)
			},
		}, true
	case "items":
		return &objects.Builtin{
			Name: "defaultdict.items",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.NewDictItems(dd.Dict)
			},
		}, true
	case "get":
		return &objects.Builtin{
			Name: "defaultdict.get",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("get() takes at least 1 argument")
				}
				var defaultVal objects.Object = objects.None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				if val, ok := dd.Dict.Get(args[0]); ok {
					return val
				}
				return defaultVal
			},
		}, true
	case "__getitem__":
		return &objects.Builtin{
			Name: "defaultdict.__getitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__getitem__() takes at least 1 argument")
				}
				if val, ok := dd.Dict.Get(args[0]); ok {
					return val
				}
				// Call factory to get default value
				defaultVal := objects.CallFunction(dd.Factory)
				dd.Dict.Set(args[0], defaultVal)
				return defaultVal
			},
		}, true
	case "__setitem__":
		return &objects.Builtin{
			Name: "defaultdict.__setitem__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 2 {
					return objects.NewTypeError("__setitem__() takes at least 2 arguments")
				}
				dd.Dict.Set(args[0], args[1])
				return objects.None_
			},
		}, true
	case "pop":
		return &objects.Builtin{
			Name: "defaultdict.pop",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("pop() takes at least 1 argument")
				}
				var defaultVal objects.Object = objects.None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				if val, ok := dd.Dict.Get(args[0]); ok {
					dd.Dict.Delete(args[0])
					return val
				}
				return defaultVal
			},
		}, true
	case "clear":
		return &objects.Builtin{
			Name: "defaultdict.clear",
			Fn: func(args ...objects.Object) objects.Object {
				dd.Dict = objects.NewDict()
				return objects.None_
			},
		}, true
	case "update":
		return &objects.Builtin{
			Name: "defaultdict.update",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.None_
				}
				switch other := args[0].(type) {
				case *objects.Dict:
					for keyStr, val := range other.Pairs {
						key := other.Keys[keyStr]
						dd.Dict.Set(key, val)
					}
				case *defaultdict:
					for keyStr, val := range other.Dict.Pairs {
						key := other.Dict.Keys[keyStr]
						dd.Dict.Set(key, val)
					}
				}
				return objects.None_
			},
		}, true
	case "copy":
		return &objects.Builtin{
			Name: "defaultdict.copy",
			Fn: func(args ...objects.Object) objects.Object {
				newDD := newDefaultDict(dd.Factory)
				for _, keyStr := range dd.Dict.KeyOrder {
					key := dd.Dict.Keys[keyStr]
					val := dd.Dict.Pairs[keyStr]
					newDD.Dict.Set(key, val)
				}
				return newDD
			},
		}, true
	case "setdefault":
		return &objects.Builtin{
			Name: "defaultdict.setdefault",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("setdefault() takes at least 1 argument")
				}
				var defaultVal objects.Object = objects.None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				if val, ok := dd.Dict.Get(args[0]); ok {
					return val
				}
				dd.Dict.Set(args[0], defaultVal)
				return defaultVal
			},
		}, true
	case "__len__":
		return &objects.Builtin{
			Name: "defaultdict.__len__",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.Integer{Value: int64(dd.Dict.Size())}
			},
		}, true
	}
	return nil, false
}

// CreateCollectionsModule creates the collections module
func CreateCollectionsModule() *objects.Module {
	module := &objects.Module{
		Name:    "collections",
		Fields: make(map[string]objects.Object),
	}

	// collections.deque
	module.Fields["deque"] = &objects.Builtin{
		Name: "collections.deque",
		Fn: func(args ...objects.Object) objects.Object {
			d := newDeque()
			if len(args) >= 1 {
				switch iter := args[0].(type) {
				case *objects.List:
					d.items = append(d.items, iter.Elements...)
				case *objects.Tuple:
					d.items = append(d.items, iter.Elements...)
				}
			}
			return d
		},
	}

	// collections.OrderedDict
	module.Fields["OrderedDict"] = &objects.Builtin{
		Name: "collections.OrderedDict",
		Fn: func(args ...objects.Object) objects.Object {
			od := newOrderedDict()
			if len(args) >= 1 {
				switch other := args[0].(type) {
				case *objects.Dict:
					for _, keyStr := range other.KeyOrder {
						key := other.Keys[keyStr]
						val := other.Pairs[keyStr]
						od.Dict.Set(key, val)
						od.KeyOrder = append(od.KeyOrder, keyStr)
					}
				}
			}
			return od
		},
	}

	// collections.Counter
	module.Fields["Counter"] = &objects.Builtin{
		Name: "collections.Counter",
		Fn: func(args ...objects.Object) objects.Object {
			c := newCounter()
			if len(args) >= 1 {
				switch iter := args[0].(type) {
				case *objects.List:
					for _, elem := range iter.Elements {
						c.increment(elem)
					}
				case *objects.Dict:
					for _, keyStr := range iter.KeyOrder {
						key := iter.Keys[keyStr]
						if v, ok := iter.Pairs[keyStr].(*objects.Integer); ok {
							c.Dict.Set(key, v)
						}
					}
				}
			}
			return c
		},
	}

	// collections.defaultdict
	module.Fields["defaultdict"] = &objects.Builtin{
		Name: "collections.defaultdict",
		Fn: func(args ...objects.Object) objects.Object {
			var factory objects.Object = objects.None_
			if len(args) >= 1 {
				factory = args[0]
			}
			return newDefaultDict(factory)
		},
	}

	return module
}
