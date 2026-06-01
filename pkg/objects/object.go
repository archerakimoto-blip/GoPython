package objects

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode"
)

type ObjectType string

const (
	INTEGER_OBJ      ObjectType = "INTEGER"
	FLOAT_OBJ        ObjectType = "FLOAT"
	BOOLEAN_OBJ      ObjectType = "BOOLEAN"
	STRING_OBJ       ObjectType = "STRING"
	NONE_OBJ         ObjectType = "NONE"
	LIST_OBJ         ObjectType = "LIST"
	TUPLE_OBJ        ObjectType = "TUPLE"
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
	ASYNC_OBJ         ObjectType = "ASYNC"
	FUTURE_OBJ         ObjectType = "FUTURE"
	BOUNDMETHOD_OBJ   ObjectType = "BOUNDMETHOD"
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

func GetStringMethod(s *String, name string) (*Builtin, bool) {
	switch name {
	case "upper":
		return &Builtin{
			Name: "upper",
			Fn: func(args ...Object) Object {
				return &String{Value: strings.ToUpper(s.Value)}
			},
		}, true
	case "lower":
		return &Builtin{
			Name: "lower",
			Fn: func(args ...Object) Object {
				return &String{Value: strings.ToLower(s.Value)}
			},
		}, true
	case "capitalize":
		return &Builtin{
			Name: "capitalize",
			Fn: func(args ...Object) Object {
				if len(s.Value) == 0 {
					return s
				}
				capitalized := strings.ToUpper(string(s.Value[0])) + strings.ToLower(s.Value[1:])
				return &String{Value: capitalized}
			},
		}, true
	case "title":
		return &Builtin{
			Name: "title",
			Fn: func(args ...Object) Object {
				return &String{Value: strings.Title(s.Value)}
			},
		}, true
	case "swapcase":
		return &Builtin{
			Name: "swapcase",
			Fn: func(args ...Object) Object {
				var result strings.Builder
				for _, c := range s.Value {
					if 'a' <= c && c <= 'z' {
						result.WriteRune(c - 32)
					} else if 'A' <= c && c <= 'Z' {
						result.WriteRune(c + 32)
					} else {
						result.WriteRune(c)
					}
				}
				return &String{Value: result.String()}
			},
		}, true
	case "split":
		return &Builtin{
			Name: "split",
			Fn: func(args ...Object) Object {
				var sep string
				if len(args) > 0 {
					if sArg, ok := args[0].(*String); ok {
						sep = sArg.Value
					}
				}
				var parts []string
				if sep == "" {
					parts = strings.Fields(s.Value)
				} else {
					parts = strings.Split(s.Value, sep)
				}
				elements := make([]Object, len(parts))
				for i, part := range parts {
					elements[i] = &String{Value: part}
				}
				return NewList(elements)
			},
		}, true
	case "join":
		return &Builtin{
			Name: "join",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("join() takes exactly 1 argument")
				}
				list, ok := args[0].(*List)
				if !ok {
					return NewTypeError("join() argument must be an iterable of strings")
				}
				var parts []string
				for _, el := range list.Elements {
					if sEl, ok := el.(*String); ok {
						parts = append(parts, sEl.Value)
					} else {
						return NewTypeError("join() argument must be an iterable of strings")
					}
				}
				return &String{Value: strings.Join(parts, s.Value)}
			},
		}, true
	case "strip":
		return &Builtin{
			Name: "strip",
			Fn: func(args ...Object) Object {
				var cutset string
				if len(args) > 0 {
					if sArg, ok := args[0].(*String); ok {
						cutset = sArg.Value
					}
				}
				if cutset == "" {
					return &String{Value: strings.TrimSpace(s.Value)}
				}
				return &String{Value: strings.Trim(s.Value, cutset)}
			},
		}, true
	case "lstrip":
		return &Builtin{
			Name: "lstrip",
			Fn: func(args ...Object) Object {
				var cutset string
				if len(args) > 0 {
					if sArg, ok := args[0].(*String); ok {
						cutset = sArg.Value
					}
				}
				if cutset == "" {
					return &String{Value: strings.TrimLeftFunc(s.Value, unicode.IsSpace)}
				}
				return &String{Value: strings.TrimLeft(s.Value, cutset)}
			},
		}, true
	case "rstrip":
		return &Builtin{
			Name: "rstrip",
			Fn: func(args ...Object) Object {
				var cutset string
				if len(args) > 0 {
					if sArg, ok := args[0].(*String); ok {
						cutset = sArg.Value
					}
				}
				if cutset == "" {
					return &String{Value: strings.TrimRightFunc(s.Value, unicode.IsSpace)}
				}
				return &String{Value: strings.TrimRight(s.Value, cutset)}
			},
		}, true
	case "startswith":
		return &Builtin{
			Name: "startswith",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("startswith() takes exactly 1 argument")
				}
				sArg, ok := args[0].(*String)
				if !ok {
					return NewTypeError("startswith() argument must be a string")
				}
				if strings.HasPrefix(s.Value, sArg.Value) {
					return True
				}
				return False
			},
		}, true
	case "endswith":
		return &Builtin{
			Name: "endswith",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("endswith() takes exactly 1 argument")
				}
				sArg, ok := args[0].(*String)
				if !ok {
					return NewTypeError("endswith() argument must be a string")
				}
				if strings.HasSuffix(s.Value, sArg.Value) {
					return True
				}
				return False
			},
		}, true
	case "replace":
		return &Builtin{
			Name: "replace",
			Fn: func(args ...Object) Object {
				if len(args) < 2 {
					return NewTypeError("replace() takes at least 2 arguments")
				}
				old, ok1 := args[0].(*String)
				newStr, ok2 := args[1].(*String)
				if !ok1 || !ok2 {
					return NewTypeError("replace() arguments must be strings")
				}
				if len(args) >= 3 {
					if n, ok := args[2].(*Integer); ok {
						return &String{Value: strings.Replace(s.Value, old.Value, newStr.Value, int(n.Value))}
					}
				}
				return &String{Value: strings.ReplaceAll(s.Value, old.Value, newStr.Value)}
			},
		}, true
	case "find":
		return &Builtin{
			Name: "find",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("find() takes at least 1 argument")
				}
				sArg, ok := args[0].(*String)
				if !ok {
					return NewTypeError("find() argument must be a string")
				}
				start := 0
				if len(args) >= 2 {
					if n, ok := args[1].(*Integer); ok {
						start = int(n.Value)
					}
				}
				end := len(s.Value)
				if len(args) >= 3 {
					if n, ok := args[2].(*Integer); ok {
						end = int(n.Value)
					}
				}
				idx := strings.Index(s.Value[start:end], sArg.Value)
				if idx == -1 {
					return &Integer{Value: -1}
				}
				return &Integer{Value: int64(idx + start)}
			},
		}, true
	case "count":
		return &Builtin{
			Name: "count",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("count() takes at least 1 argument")
				}
				sArg, ok := args[0].(*String)
				if !ok {
					return NewTypeError("count() argument must be a string")
				}
				start := 0
				if len(args) >= 2 {
					if n, ok := args[1].(*Integer); ok {
						start = int(n.Value)
					}
				}
				end := len(s.Value)
				if len(args) >= 3 {
					if n, ok := args[2].(*Integer); ok {
						end = int(n.Value)
					}
				}
				count := strings.Count(s.Value[start:end], sArg.Value)
				return &Integer{Value: int64(count)}
			},
		}, true
	case "isalpha":
		return &Builtin{
			Name: "isalpha",
			Fn: func(args ...Object) Object {
				for _, c := range s.Value {
					if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z') {
						return False
					}
				}
				return True
			},
		}, true
	case "isdigit":
		return &Builtin{
			Name: "isdigit",
			Fn: func(args ...Object) Object {
				for _, c := range s.Value {
					if !('0' <= c && c <= '9') {
						return False
					}
				}
				return True
			},
		}, true
	case "isalnum":
		return &Builtin{
			Name: "isalnum",
			Fn: func(args ...Object) Object {
				for _, c := range s.Value {
					if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9') {
						return False
					}
				}
				return True
			},
		}, true
	case "isspace":
		return &Builtin{
			Name: "isspace",
			Fn: func(args ...Object) Object {
				for _, c := range s.Value {
					if !(' ' == c || '\t' == c || '\n' == c || '\r' == c) {
						return False
					}
				}
				return True
			},
		}, true
	case "isupper":
		return &Builtin{
			Name: "isupper",
			Fn: func(args ...Object) Object {
				hasAlpha := false
				for _, c := range s.Value {
					if 'a' <= c && c <= 'z' {
						return False
					}
					if 'A' <= c && c <= 'Z' {
						hasAlpha = true
					}
				}
				if hasAlpha {
					return True
				}
				return False
			},
		}, true
	case "islower":
		return &Builtin{
			Name: "islower",
			Fn: func(args ...Object) Object {
				hasAlpha := false
				for _, c := range s.Value {
					if 'A' <= c && c <= 'Z' {
						return False
					}
					if 'a' <= c && c <= 'z' {
						hasAlpha = true
					}
				}
				if hasAlpha {
					return True
				}
				return False
			},
		}, true
	case "format":
		return &Builtin{
			Name: "format",
			Fn: func(args ...Object) Object {
				kwargs := make(map[string]Object)
				if len(args) > 0 {
					if last, ok := args[len(args)-1].(*Dict); ok {
						kwargs = make(map[string]Object, len(last.Keys))
						for keyStr, key := range last.Keys {
							kwargs[keyStr] = last.Pairs[keyStr]
							if sKey, ok := key.(*String); ok {
								kwargs[sKey.Value] = last.Pairs[keyStr]
							}
						}
						args = args[:len(args)-1]
					}
				}
				var result strings.Builder
				i := 0
				argIdx := 0
				for i < len(s.Value) {
					if s.Value[i] == '{' {
						if i+1 < len(s.Value) && s.Value[i+1] == '{' {
							result.WriteByte('{')
							i += 2
							continue
						}
						j := i + 1
						for j < len(s.Value) && s.Value[j] != '}' {
							j++
						}
						if j >= len(s.Value) {
							result.WriteByte(s.Value[i])
							i++
							continue
						}
						field := s.Value[i+1 : j]
						var replacement string
						if field == "" {
							if argIdx < len(args) {
								replacement = args[argIdx].Inspect()
								argIdx++
							} else {
								replacement = ""
							}
						} else if field[0] >= '0' && field[0] <= '9' {
							idx := 0
							for _, c := range field {
								if c >= '0' && c <= '9' {
									idx = idx*10 + int(c-'0')
								} else {
									break
								}
							}
							if idx < len(args) {
								replacement = args[idx].Inspect()
							} else {
								replacement = ""
							}
						} else {
							if val, ok := kwargs[field]; ok {
								replacement = val.Inspect()
							} else {
								replacement = ""
							}
						}
						result.WriteString(replacement)
						i = j + 1
					} else if s.Value[i] == '}' {
						if i+1 < len(s.Value) && s.Value[i+1] == '}' {
							result.WriteByte('}')
							i += 2
							continue
						}
						result.WriteByte(s.Value[i])
						i++
					} else {
						result.WriteByte(s.Value[i])
						i++
					}
				}
				return &String{Value: result.String()}
			},
		}, true
	}
	return nil, false
}

type None struct{}

func (n *None) Type() ObjectType { return NONE_OBJ }
func (n *None) Inspect() string  { return "None" }

func GetListMethod(l *List, name string) (*Builtin, bool) {
	switch name {
	case "append":
		return &Builtin{
			Name: "append",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("append() takes exactly 1 argument")
				}
				l.Append(args[0])
				return None_
			},
		}, true
	case "extend":
		return &Builtin{
			Name: "extend",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("extend() takes exactly 1 argument")
				}
				if other, ok := args[0].(*List); ok {
					l.Extend(other)
				}
				return None_
			},
		}, true
	case "pop":
		return &Builtin{
			Name: "pop",
			Fn: func(args ...Object) Object {
				if len(args) > 1 {
					return NewTypeError("pop() takes at most 1 argument")
				}
				if len(args) == 0 {
					val, err := l.Pop()
					if err != nil {
						return NewIndexError(err.Error())
					}
					return val
				}
				if idx, ok := args[0].(*Integer); ok {
					val, err := l.Pop(int(idx.Value))
					if err != nil {
						return NewIndexError(err.Error())
					}
					return val
				}
				return NewTypeError("pop() argument must be an integer")
			},
		}, true
	case "insert":
		return &Builtin{
			Name: "insert",
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return NewTypeError("insert() takes exactly 2 arguments")
				}
				if idx, ok := args[0].(*Integer); ok {
					l.Insert(int(idx.Value), args[1])
				}
				return None_
			},
		}, true
	case "remove":
		return &Builtin{
			Name: "remove",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("remove() takes exactly 1 argument")
				}
				err := l.Remove(args[0])
				if err != nil {
					return NewValueError(err.Error())
				}
				return None_
			},
		}, true
	case "reverse":
		return &Builtin{
			Name: "reverse",
			Fn: func(args ...Object) Object {
				l.Reverse()
				return None_
			},
		}, true
	case "clear":
		return &Builtin{
			Name: "clear",
			Fn: func(args ...Object) Object {
				l.Clear()
				return None_
			},
		}, true
	case "index":
		return &Builtin{
			Name: "index",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("index() takes at least 1 argument")
				}
				idx := l.Index(args[0])
				if idx == -1 {
					return NewValueError("value not in list")
				}
				return &Integer{Value: int64(idx)}
			},
		}, true
	case "count":
		return &Builtin{
			Name: "count",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("count() takes exactly 1 argument")
				}
				count := 0
				for _, el := range l.Elements {
					if Equal(el, args[0]) {
						count++
					}
				}
				return &Integer{Value: int64(count)}
			},
		}, true
	}
	return nil, false
}

func GetDictMethod(d *Dict, name string) (*Builtin, bool) {
	switch name {
	case "get":
		return &Builtin{
			Name: "get",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("get() takes at least 1 argument")
				}
				if val, ok := d.Get(args[0]); ok {
					return val
				}
				if len(args) >= 2 {
					return args[1]
				}
				return None_
			},
		}, true
	case "set":
		fallthrough
	case "setdefault":
		return &Builtin{
			Name: "setdefault",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("setdefault() takes at least 1 argument")
				}
				if val, ok := d.Get(args[0]); ok {
					return val
				}
				var defaultVal Object = None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				d.Set(args[0], defaultVal)
				return defaultVal
			},
		}, true
	case "clear":
		return &Builtin{
			Name: "clear",
			Fn: func(args ...Object) Object {
				*d = *NewDict()
				return None_
			},
		}, true
	case "keys":
		return &Builtin{
			Name: "keys",
			Fn: func(args ...Object) Object {
				return NewList(d.KeysSlice())
			},
		}, true
	case "values":
		return &Builtin{
			Name: "values",
			Fn: func(args ...Object) Object {
				return NewList(d.ValuesSlice())
			},
		}, true
	case "items":
		return &Builtin{
			Name: "items",
			Fn: func(args ...Object) Object {
				elements := make([]Object, 0, len(d.Keys))
				for _, key := range d.Keys {
					value, _ := d.Get(key)
					elements = append(elements, NewList([]Object{key, value}))
				}
				return NewList(elements)
			},
		}, true
	case "pop":
		return &Builtin{
			Name: "pop",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("pop() takes at least 1 argument")
				}
				if val, ok := d.Get(args[0]); ok {
					d.Delete(args[0])
					return val
				}
				if len(args) >= 2 {
					return args[1]
				}
				return NewKeyError("key not found")
			},
		}, true
	case "popitem":
		return &Builtin{
			Name: "popitem",
			Fn: func(args ...Object) Object {
				if len(d.Keys) == 0 {
					return NewKeyError("popitem(): dictionary is empty")
				}
				for keyStr, key := range d.Keys {
					val := d.Pairs[keyStr]
					d.Delete(key)
					return NewTuple([]Object{key, val})
				}
				return None_
			},
		}, true
	case "update":
		return &Builtin{
			Name: "update",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("update() takes exactly 1 argument")
				}
				if other, ok := args[0].(*Dict); ok {
					for keyStr, key := range other.Keys {
						d.Set(key, other.Pairs[keyStr])
					}
				}
				return None_
			},
		}, true
	}
	return nil, false
}

func GetSetMethod(s *Set, name string) (*Builtin, bool) {
	switch name {
	case "add":
		return &Builtin{
			Name: "add",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("add() takes exactly 1 argument")
				}
				s.Add(args[0])
				return None_
			},
		}, true
	case "remove":
		return &Builtin{
			Name: "remove",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("remove() takes exactly 1 argument")
				}
				if !s.Contains(args[0]) {
					return NewKeyError("element not in set")
				}
				s.Remove(args[0])
				return None_
			},
		}, true
	case "discard":
		return &Builtin{
			Name: "discard",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("discard() takes exactly 1 argument")
				}
				s.Remove(args[0])
				return None_
			},
		}, true
	case "clear":
		return &Builtin{
			Name: "clear",
			Fn: func(args ...Object) Object {
				*s = *NewSet()
				return None_
			},
		}, true
	case "pop":
		return &Builtin{
			Name: "pop",
			Fn: func(args ...Object) Object {
				for _, val := range s.Elements {
					s.Remove(val)
					return val
				}
				return NewKeyError("pop from an empty set")
			},
		}, true
	case "issubset":
		return &Builtin{
			Name: "issubset",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("issubset() takes exactly 1 argument")
				}
				if other, ok := args[0].(*Set); ok {
					for _, val := range s.Elements {
						if !other.Contains(val) {
							return False
						}
					}
					return True
				}
				return False
			},
		}, true
	case "issuperset":
		return &Builtin{
			Name: "issuperset",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("issuperset() takes exactly 1 argument")
				}
				if other, ok := args[0].(*Set); ok {
					for _, val := range other.Elements {
						if !s.Contains(val) {
							return False
						}
					}
					return True
				}
				return False
			},
		}, true
	case "union":
		return &Builtin{
			Name: "union",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("union() takes exactly 1 argument")
				}
				result := NewSet()
				for _, val := range s.Elements {
					result.Add(val)
				}
				if other, ok := args[0].(*Set); ok {
					for _, val := range other.Elements {
						result.Add(val)
					}
				}
				return result
			},
		}, true
	case "intersection":
		return &Builtin{
			Name: "intersection",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("intersection() takes exactly 1 argument")
				}
				result := NewSet()
				if other, ok := args[0].(*Set); ok {
					for _, val := range s.Elements {
						if other.Contains(val) {
							result.Add(val)
						}
					}
				}
				return result
			},
		}, true
	case "difference":
		return &Builtin{
			Name: "difference",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("difference() takes exactly 1 argument")
				}
				result := NewSet()
				if other, ok := args[0].(*Set); ok {
					for _, val := range s.Elements {
						if !other.Contains(val) {
							result.Add(val)
						}
					}
				}
				return result
			},
		}, true
	case "symmetric_difference":
		return &Builtin{
			Name: "symmetric_difference",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("symmetric_difference() takes exactly 1 argument")
				}
				result := NewSet()
				if other, ok := args[0].(*Set); ok {
					for _, val := range s.Elements {
						if !other.Contains(val) {
							result.Add(val)
						}
					}
					for _, val := range other.Elements {
						if !s.Contains(val) {
							result.Add(val)
						}
					}
				}
				return result
			},
		}, true
	}
	return nil, false
}

type List struct {
	Elements []Object
}

func NewList(elements []Object) *List {
	return &List{Elements: elements}
}

type Tuple struct {
	Elements []Object
}

func NewTuple(elements []Object) *Tuple {
	return &Tuple{Elements: elements}
}

func (t *Tuple) Type() ObjectType { return TUPLE_OBJ }
func (t *Tuple) Inspect() string {
	result := "("
	for i, el := range t.Elements {
		if i > 0 {
			result += ", "
		}
		result += el.Inspect()
	}
	result += ")"
	return result
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
func (e *Error) Error() string { return e.ErrorType + ": " + e.Message }

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

type Async struct {
	Instructions  []byte
	Constants     []Object
	Locals        []Object
	IP            int
	Stack         []Object
	StackPtr      int
	BasePointer   int
	Done          bool
	Result        Object
}

func (a *Async) Type() ObjectType { return ASYNC_OBJ }
func (a *Async) Inspect() string  { return fmt.Sprintf("async[%p]", a) }

type Future struct {
	Done         bool
	Result       Object
	Error        error
	WaitChan     chan struct{}
}

func (f *Future) Type() ObjectType { return FUTURE_OBJ }
func (f *Future) Inspect() string {
	if f.Done {
		if f.Error != nil {
			return fmt.Sprintf("future[error: %v]", f.Error)
		}
		return fmt.Sprintf("future[result: %v]", f.Result.Inspect())
	}
	return "future[pending]"
}

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

type BoundMethod struct {
	Fn   Object
	Self Object
}

func (bm *BoundMethod) Type() ObjectType { return BOUNDMETHOD_OBJ }
func (bm *BoundMethod) Inspect() string  { return "<bound method>" }

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

// CreateSysModule 创建 sys 模块
func CreateSysModule() *Module {
	sysModule := &Module{
		Name:    "sys",
		Fields: make(map[string]Object),
	}

	// sys.version
	sysModule.Fields["version"] = &String{
		Value: "GoPython 0.1.0 (Go implementation)",
	}

	// sys.platform
	sysModule.Fields["platform"] = &String{
		Value: runtime.GOOS,
	}

	// sys.version_info
	sysModule.Fields["version_info"] = &Tuple{
		Elements: []Object{
			&Integer{Value: 0},
			&Integer{Value: 1},
			&Integer{Value: 0},
			&String{Value: "final"},
			&Integer{Value: 0},
		},
	}

	// sys.argv
	sysModule.Fields["argv"] = &List{
		Elements: []Object{&String{Value: "gopy"}},
	}

	// sys.path
	sysModule.Fields["path"] = &List{
		Elements: []Object{&String{Value: "."}},
	}

	// sys.exit
	sysModule.Fields["exit"] = &Builtin{
		Name: "sys.exit",
		Fn: func(args ...Object) Object {
			code := 0
			if len(args) >= 1 {
				if i, ok := args[0].(*Integer); ok {
					code = int(i.Value)
				}
			}
			os.Exit(code)
			return None_
		},
	}

	// sys.getsizeof
	sysModule.Fields["getsizeof"] = &Builtin{
		Name: "sys.getsizeof",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("getsizeof() takes exactly 1 argument")
			}
			// 简单实现，返回固定大小
			return &Integer{Value: 24}
		},
	}

	return sysModule
}

// CreateOsModule 创建 os 模块
func CreateOsModule() *Module {
	osModule := &Module{
		Name:    "os",
		Fields: make(map[string]Object),
	}

	// os.path.sep
	osModule.Fields["sep"] = &String{
		Value: string(os.PathSeparator),
	}

	// os.getcwd
	osModule.Fields["getcwd"] = &Builtin{
		Name: "os.getcwd",
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return NewTypeError("getcwd() takes no arguments")
			}
			cwd, err := os.Getwd()
			if err != nil {
				return NewError(err.Error())
			}
			return &String{Value: cwd}
		},
	}

	// os.chdir
	osModule.Fields["chdir"] = &Builtin{
		Name: "os.chdir",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("chdir() takes exactly 1 argument")
			}
			path, ok := args[0].(*String)
			if !ok {
				return NewTypeError("chdir() argument must be a string")
			}
			err := os.Chdir(path.Value)
			if err != nil {
				return NewError(err.Error())
			}
			return None_
		},
	}

	// os.listdir
	osModule.Fields["listdir"] = &Builtin{
		Name: "os.listdir",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("listdir() takes exactly 1 argument")
			}
			path, ok := args[0].(*String)
			if !ok {
				return NewTypeError("listdir() argument must be a string")
			}
			entries, err := os.ReadDir(path.Value)
			if err != nil {
				return NewError(err.Error())
			}
			files := make([]Object, 0, len(entries))
			for _, entry := range entries {
				files = append(files, &String{Value: entry.Name()})
			}
			return &List{Elements: files}
		},
	}

	// os.mkdir
	osModule.Fields["mkdir"] = &Builtin{
		Name: "os.mkdir",
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("mkdir() takes at least 1 argument")
			}
			path, ok := args[0].(*String)
			if !ok {
				return NewTypeError("mkdir() first argument must be a string")
			}
			mode := 0755
			if len(args) >= 2 {
				if i, ok := args[1].(*Integer); ok {
					mode = int(i.Value)
				}
			}
			err := os.Mkdir(path.Value, os.FileMode(mode))
			if err != nil {
				return NewError(err.Error())
			}
			return None_
		},
	}

	// os.remove
	osModule.Fields["remove"] = &Builtin{
		Name: "os.remove",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("remove() takes exactly 1 argument")
			}
			path, ok := args[0].(*String)
			if !ok {
				return NewTypeError("remove() argument must be a string")
			}
			err := os.Remove(path.Value)
			if err != nil {
				return NewError(err.Error())
			}
			return None_
		},
	}

	// os.rename
	osModule.Fields["rename"] = &Builtin{
		Name: "os.rename",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return NewTypeError("rename() takes exactly 2 arguments")
			}
			src, ok1 := args[0].(*String)
			dst, ok2 := args[1].(*String)
			if !ok1 || !ok2 {
				return NewTypeError("rename() arguments must be strings")
			}
			err := os.Rename(src.Value, dst.Value)
			if err != nil {
				return NewError(err.Error())
			}
			return None_
		},
	}

	// os.getenv
	osModule.Fields["getenv"] = &Builtin{
		Name: "os.getenv",
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("getenv() takes at least 1 argument")
			}
			key, ok := args[0].(*String)
			if !ok {
				return NewTypeError("getenv() first argument must be a string")
			}
			value := os.Getenv(key.Value)
			if value == "" && len(args) >= 2 {
				return args[1]
			}
			if value == "" {
				return None_
			}
			return &String{Value: value}
		},
	}

	// os.environ
	osModule.Fields["environ"] = &Dict{
		Pairs: make(map[string]Object),
		Keys:  make(map[string]Object),
	}

	return osModule
}

// CreateJsonModule 创建 json 模块
func CreateJsonModule() *Module {
	jsonModule := &Module{
		Name:    "json",
		Fields: make(map[string]Object),
	}

	// json.dumps
	jsonModule.Fields["dumps"] = &Builtin{
		Name: "json.dumps",
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("dumps() takes at least 1 argument")
			}

			// 将对象转换为 Go 值
			goValue := convertToGoValue(args[0])

			// 序列化到 JSON
			data, err := json.Marshal(goValue)
			if err != nil {
				return NewError(err.Error())
			}

			return &String{Value: string(data)}
		},
	}

	// json.loads
	jsonModule.Fields["loads"] = &Builtin{
		Name: "json.loads",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("loads() takes exactly 1 argument")
			}
			s, ok := args[0].(*String)
			if !ok {
				return NewTypeError("loads() argument must be a string")
			}

			var data interface{}
			err := json.Unmarshal([]byte(s.Value), &data)
			if err != nil {
				return NewError(err.Error())
			}

			return convertToObject(data)
		},
	}

	return jsonModule
}

// 辅助函数：将 GoPython 对象转换为 Go 值
func convertToGoValue(obj Object) interface{} {
	switch v := obj.(type) {
	case *Integer:
		return v.Value
	case *Float:
		return v.Value
	case *Boolean:
		return v.Value
	case *String:
		return v.Value
	case *List:
		result := make([]interface{}, len(v.Elements))
		for i, elem := range v.Elements {
			result[i] = convertToGoValue(elem)
		}
		return result
	case *Dict:
		result := make(map[string]interface{})
		for keyStr, key := range v.Keys {
			value := v.Pairs[keyStr]
			// 简单处理：只支持字符串键
			if keyObj, ok := key.(*String); ok {
				result[keyObj.Value] = convertToGoValue(value)
			}
		}
		return result
	case *None:
		return nil
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 辅助函数：将 Go 值转换为 GoPython 对象
func convertToObject(value interface{}) Object {
	switch v := value.(type) {
	case int:
		return &Integer{Value: int64(v)}
	case int64:
		return &Integer{Value: v}
	case float64:
		return &Float{Value: v}
	case bool:
		if v {
			return True
		}
		return False
	case string:
		return &String{Value: v}
	case []interface{}:
		elements := make([]Object, len(v))
		for i, elem := range v {
			elements[i] = convertToObject(elem)
		}
		return &List{Elements: elements}
	case map[string]interface{}:
		dict := NewDict()
		for key, val := range v {
			dict.Set(&String{Value: key}, convertToObject(val))
		}
		return dict
	case nil:
		return None_
	default:
		// 对于未知类型，返回字符串表示
		return &String{Value: fmt.Sprintf("%v", v)}
	}
}

// CreateRandomModule 创建 random 模块
func CreateRandomModule() *Module {
	randomModule := &Module{
		Name:    "random",
		Fields: make(map[string]Object),
	}
	
	// random.seed
	randomModule.Fields["seed"] = &Builtin{
		Name: "random.seed",
		Fn: func(args ...Object) Object {
			if len(args) == 0 {
				rand.Seed(time.Now().UnixNano())
				return None_
			}
			switch v := args[0].(type) {
			case *Integer:
				rand.Seed(v.Value)
			case *Float:
				rand.Seed(int64(v.Value))
			default:
				return NewTypeError("seed() argument must be a number")
			}
			return None_
		},
	}
	
	// random.random
	randomModule.Fields["random"] = &Builtin{
		Name: "random.random",
		Fn: func(args ...Object) Object {
			return &Float{Value: rand.Float64()}
		},
	}
	
	// random.uniform
	randomModule.Fields["uniform"] = &Builtin{
		Name: "random.uniform",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return NewTypeError("uniform() takes exactly 2 arguments")
			}
			a, ok1 := args[0].(*Float)
			b, ok2 := args[1].(*Float)
			if !ok1 || !ok2 {
				if i1, ok := args[0].(*Integer); ok {
					a = &Float{Value: float64(i1.Value)}
					ok1 = true
				}
				if i2, ok := args[1].(*Integer); ok {
					b = &Float{Value: float64(i2.Value)}
					ok2 = true
				}
				if !ok1 || !ok2 {
					return NewTypeError("uniform() arguments must be numbers")
				}
			}
			result := a.Value + rand.Float64()*(b.Value - a.Value)
			return &Float{Value: result}
		},
	}
	
	// random.randint
	randomModule.Fields["randint"] = &Builtin{
		Name: "random.randint",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return NewTypeError("randint() takes exactly 2 arguments")
			}
			a, ok1 := args[0].(*Integer)
			b, ok2 := args[1].(*Integer)
			if !ok1 || !ok2 {
				return NewTypeError("randint() arguments must be integers")
			}
			result := a.Value + rand.Int63n(b.Value - a.Value + 1)
			return &Integer{Value: result}
		},
	}
	
	// random.choice
	randomModule.Fields["choice"] = &Builtin{
		Name: "random.choice",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("choice() takes exactly 1 argument")
			}
			list, ok := args[0].(*List)
			if !ok {
				return NewTypeError("choice() argument must be a list")
			}
			if len(list.Elements) == 0 {
				return NewIndexError("cannot choose from empty list")
			}
			idx := rand.Intn(len(list.Elements))
			return list.Elements[idx]
		},
	}
	
	// random.shuffle
	randomModule.Fields["shuffle"] = &Builtin{
		Name: "random.shuffle",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("shuffle() takes exactly 1 argument")
			}
			list, ok := args[0].(*List)
			if !ok {
				return NewTypeError("shuffle() argument must be a list")
			}
			rand.Shuffle(len(list.Elements), func(i, j int) {
				list.Elements[i], list.Elements[j] = list.Elements[j], list.Elements[i]
			})
			return None_
		},
	}
	
	return randomModule
}

// CreateStringModule 创建 string 模块
func CreateStringModule() *Module {
	stringModule := &Module{
		Name:    "string",
		Fields: make(map[string]Object),
	}
	
	// string.ascii_letters
	stringModule.Fields["ascii_letters"] = &String{Value: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"}
	
	// string.ascii_lowercase
	stringModule.Fields["ascii_lowercase"] = &String{Value: "abcdefghijklmnopqrstuvwxyz"}
	
	// string.ascii_uppercase
	stringModule.Fields["ascii_uppercase"] = &String{Value: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}
	
	// string.digits
	stringModule.Fields["digits"] = &String{Value: "0123456789"}
	
	// string.hexdigits
	stringModule.Fields["hexdigits"] = &String{Value: "0123456789abcdefABCDEF"}
	
	// string.octdigits
	stringModule.Fields["octdigits"] = &String{Value: "01234567"}
	
	// string.punctuation
	stringModule.Fields["punctuation"] = &String{Value: "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"}
	
	// string.printable
	stringModule.Fields["printable"] = &String{Value: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~ \t\n\r\x0b\x0c"}
	
	// string.whitespace
	stringModule.Fields["whitespace"] = &String{Value: " \t\n\r\x0b\x0c"}
	
	// string.capitalize (作为演示，添加一些字符串处理函数)
	stringModule.Fields["capitalize"] = &Builtin{
		Name: "string.capitalize",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("capitalize() takes exactly 1 argument")
			}
			s, ok := args[0].(*String)
			if !ok {
				return NewTypeError("capitalize() argument must be a string")
			}
			if len(s.Value) == 0 {
				return s
			}
			capitalized := strings.ToUpper(string(s.Value[0])) + strings.ToLower(s.Value[1:])
			return &String{Value: capitalized}
		},
	}
	
	return stringModule
}

// CreateTimeModule 创建 time 模块
func CreateTimeModule() *Module {
	timeModule := &Module{
		Name:    "time",
		Fields: make(map[string]Object),
	}
	
	// time.time
	timeModule.Fields["time"] = &Builtin{
		Name: "time.time",
		Fn: func(args ...Object) Object {
			return &Float{Value: float64(time.Now().UnixNano()) / 1e9}
		},
	}
	
	// time.sleep
	timeModule.Fields["sleep"] = &Builtin{
		Name: "time.sleep",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return NewTypeError("sleep() takes exactly 1 argument")
			}
			var secs float64
			switch v := args[0].(type) {
			case *Float:
				secs = v.Value
			case *Integer:
				secs = float64(v.Value)
			default:
				return NewTypeError("sleep() argument must be a number")
			}
			time.Sleep(time.Duration(secs * float64(time.Second)))
			return None_
		},
	}
	
	// time.ctime
	timeModule.Fields["ctime"] = &Builtin{
		Name: "time.ctime",
		Fn: func(args ...Object) Object {
			var t time.Time
			if len(args) == 0 {
				t = time.Now()
			} else if len(args) == 1 {
				switch v := args[0].(type) {
				case *Float:
					t = time.Unix(0, int64(v.Value*1e9))
				case *Integer:
					t = time.Unix(v.Value, 0)
				default:
					return NewTypeError("ctime() argument must be a number")
				}
			} else {
				return NewTypeError("ctime() takes at most 1 argument")
			}
			return &String{Value: t.Format(time.UnixDate)}
		},
	}
	
	// time.localtime (返回简单表示)
	timeModule.Fields["localtime"] = &Builtin{
		Name: "time.localtime",
		Fn: func(args ...Object) Object {
			t := time.Now()
			// 返回一个简单的 tuple：(year, month, day, hour, minute, second, weekday, yearday)
			return &Tuple{
				Elements: []Object{
					&Integer{Value: int64(t.Year())},
					&Integer{Value: int64(t.Month())},
					&Integer{Value: int64(t.Day())},
					&Integer{Value: int64(t.Hour())},
					&Integer{Value: int64(t.Minute())},
					&Integer{Value: int64(t.Second())},
					&Integer{Value: int64(t.Weekday())},
					&Integer{Value: int64(t.YearDay())},
				},
			}
		},
	}
	
	return timeModule
}

// CreateDatetimeModule 创建 datetime 模块
func CreateDatetimeModule() *Module {
	datetimeModule := &Module{
		Name:    "datetime",
		Fields: make(map[string]Object),
	}
	
	// datetime.datetime.now
	datetimeModule.Fields["datetime"] = &Module{
		Name: "datetime",
		Fields: map[string]Object{
			"now": &Builtin{
				Name: "datetime.datetime.now",
				Fn: func(args ...Object) Object {
					t := time.Now()
					return &String{Value: t.Format("2006-01-02 15:04:05")}
				},
			},
		},
	}
	
	// datetime.date.today
	datetimeModule.Fields["date"] = &Module{
		Name: "date",
		Fields: map[string]Object{
			"today": &Builtin{
				Name: "datetime.date.today",
				Fn: func(args ...Object) Object {
					t := time.Now()
					return &String{Value: t.Format("2006-01-02")}
				},
			},
		},
	}
	
	return datetimeModule
}
