package objects

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
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
	FROZENSET_OBJ    ObjectType = "FROZENSET"
	DICT_OBJ         ObjectType = "DICT"
	FUNCTION_OBJ     ObjectType = "FUNCTION"
	BUILTIN_OBJ      ObjectType = "BUILTIN"
	ERROR_OBJ        ObjectType = "ERROR"
	GENERATOR_OBJ    ObjectType = "GENERATOR"
	CONTEXT_OBJ      ObjectType = "CONTEXT"
	CLASS_OBJ        ObjectType = "CLASS"
	INSTANCE_OBJ     ObjectType = "INSTANCE"
	MODULE_OBJ      ObjectType = "MODULE"
	RANGE_OBJ       ObjectType = "RANGE"
	ZIP_OBJ         ObjectType = "ZIP"
	ASYNC_OBJ         ObjectType = "ASYNC"
	FUTURE_OBJ         ObjectType = "FUTURE"
	ENUM_OBJ           ObjectType = "ENUM"
	ENUM_MEMBER_OBJ    ObjectType = "ENUM_MEMBER"
	BYTES_OBJ          ObjectType = "BYTES"
	ELLIPSIS_OBJ       ObjectType = "ELLIPSIS"
	PROPERTY_OBJ       ObjectType = "PROPERTY"
	CLASSMETHOD_OBJ    ObjectType = "CLASSMETHOD"
	STATICMETHOD_OBJ   ObjectType = "STATICMETHOD"
	SUPER_OBJ          ObjectType = "SUPER"
	EXCEPTION_GROUP_OBJ ObjectType = "EXCEPTION_GROUP"
	COMPLEX_OBJ         ObjectType = "COMPLEX"
	DICT_KEYS_OBJ       ObjectType = "DICT_KEYS"
	DICT_VALUES_OBJ     ObjectType = "DICT_VALUES"
	DICT_ITEMS_OBJ      ObjectType = "DICT_ITEMS"
	REGEX_PATTERN_OBJ   ObjectType = "REGEX_PATTERN"
	REGEX_MATCH_OBJ     ObjectType = "REGEX_MATCH"
	STRING_BUILDER_OBJ  ObjectType = "STRING_BUILDER"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

// AttributeGetter is an interface for types that have attributes
type AttributeGetter interface {
	Object
	GetAttr(name string) (Object, bool)
}

// Callable is an interface for types that can be called as functions
type Callable interface {
	Object
	Call(args ...Object) Object
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

var integerPool [512]*Integer
var integerPoolReady bool

func initIntegerPool() {
	for i := int64(0); i < 512; i++ {
		integerPool[i] = &Integer{Value: i - 256}
	}
	integerPoolReady = true
}

func GetCachedInteger(v int64) *Integer {
	if !integerPoolReady {
		initIntegerPool()
	}
	if v >= -256 && v < 256 {
		return integerPool[v+256]
	}
	return &Integer{Value: v}
}

func (i *Integer) GetAttr(name string) (Object, bool) {
	switch name {
	case "bit_length":
		return &Builtin{
			Name: "int.bit_length",
			Fn: func(args ...Object) Object {
				bits := 0
				v := i.Value
				if v < 0 {
					v = -v
				}
				for v > 0 {
					bits++
					v >>= 1
				}
				if bits == 0 {
					bits = 1
				}
				return &Integer{Value: int64(bits)}
			},
		}, true
	case "bit_count":
		return &Builtin{
			Name: "int.bit_count",
			Fn: func(args ...Object) Object {
				v := i.Value
				if v < 0 {
					v = -v
				}
				cnt := 0
				for v > 0 {
					if v&1 == 1 {
						cnt++
					}
					v >>= 1
				}
				return &Integer{Value: int64(cnt)}
			},
		}, true
	case "to_bytes":
		return &Builtin{
			Name: "int.to_bytes",
			Fn: func(args ...Object) Object {
				if len(args) < 2 {
					return NewTypeError("to_bytes() takes at least 2 arguments (%d given)", len(args))
				}
				length, ok := args[0].(*Integer)
				if !ok {
					return NewTypeError("to_bytes() first argument must be an integer")
				}
				byteorder, ok := args[1].(*String)
				if !ok {
					return NewTypeError("to_bytes() second argument must be a string")
				}
				signed := false
				if len(args) >= 3 {
					if b, ok := args[2].(*Boolean); ok {
						signed = b.Value
					}
				}
				n := int(length.Value)
				if n < 0 {
					return NewValueError("length must be non-negative")
				}
				result := make([]byte, n)
				v := i.Value
				if byteorder.Value == "big" {
					for j := n - 1; j >= 0; j-- {
						result[j] = byte(v & 0xFF)
						v >>= 8
					}
				} else {
					for j := 0; j < n; j++ {
						result[j] = byte(v & 0xFF)
						v >>= 8
					}
				}
				_ = signed
				return &Bytes{Value: result}
			},
		}, true
	case "__class__":
		return &String{Value: "int"}, true
	}
	return nil, false
}

var stringPool map[string]*String
var stringPoolReady bool

func initStringPool() {
	stringPool = make(map[string]*String, 256)
	stringPoolReady = true
}

func GetCachedString(v string) *String {
	if !stringPoolReady {
		initStringPool()
	}
	if len(v) <= 16 {
		if cached, ok := stringPool[v]; ok {
			return cached
		}
		s := &String{Value: v}
		if len(stringPool) < 4096 {
			stringPool[v] = s
		}
		return s
	}
	return &String{Value: v}
}

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

func (f *Float) GetAttr(name string) (Object, bool) {
	switch name {
	case "is_integer":
		return &Builtin{
			Name: "float.is_integer",
			Fn: func(args ...Object) Object {
				return &Boolean{Value: f.Value == math.Trunc(f.Value)}
			},
		}, true
	case "hex":
		return &Builtin{
			Name: "float.hex",
			Fn: func(args ...Object) Object {
				return &String{Value: fmt.Sprintf("%x", math.Float64bits(f.Value))}
			},
		}, true
	case "as_integer_ratio":
		return &Builtin{
			Name: "float.as_integer_ratio",
			Fn: func(args ...Object) Object {
				if math.IsNaN(f.Value) || math.IsInf(f.Value, 0) {
					return NewValueError("cannot convert float to integer ratio")
				}
				if f.Value == 0 {
					return &Tuple{Elements: []Object{&Integer{Value: 0}, &Integer{Value: 1}}}
				}
				ratio := new(big.Rat).SetFloat64(f.Value)
				num := ratio.Num().Int64()
				den := ratio.Denom().Int64()
				return &Tuple{Elements: []Object{&Integer{Value: num}, &Integer{Value: den}}}
			},
		}, true
	case "__class__":
		return &String{Value: "float"}, true
	}
	return nil, false
}

type Complex struct {
	Real float64
	Imag float64
}

func NewComplex(real, imag float64) *Complex {
	return &Complex{Real: real, Imag: imag}
}

func (c *Complex) Type() ObjectType { return COMPLEX_OBJ }
func (c *Complex) Inspect() string {
	if c.Real == 0 {
		return fmt.Sprintf("%gj", c.Imag)
	}
	if c.Imag < 0 {
		return fmt.Sprintf("(%g%gj)", c.Real, c.Imag)
	}
	return fmt.Sprintf("(%g+%gj)", c.Real, c.Imag)
}

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

func (s *String) GetAttr(name string) (Object, bool) {
	switch name {
	case "rfind":
		return &Builtin{
			Name: "str.rfind",
			Fn: func(args ...Object) Object {
				sub, ok := args[0].(*String)
				if !ok {
					return NewTypeError("rfind() argument must be a string")
				}
				start := 0
				end := len(s.Value)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(s.Value) {
					end = len(s.Value)
				}
				if start >= end {
					return &Integer{Value: -1}
				}
				idx := strings.LastIndex(s.Value[start:end], sub.Value)
				if idx == -1 {
					return &Integer{Value: -1}
				}
				return &Integer{Value: int64(start + idx)}
			},
		}, true
	case "rindex":
		return &Builtin{
			Name: "str.rindex",
			Fn: func(args ...Object) Object {
				sub, ok := args[0].(*String)
				if !ok {
					return NewTypeError("rindex() argument must be a string")
				}
				start := 0
				end := len(s.Value)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(s.Value) {
					end = len(s.Value)
				}
				if start >= end {
					return NewValueError("substring not found")
				}
				idx := strings.LastIndex(s.Value[start:end], sub.Value)
				if idx == -1 {
					return NewValueError("substring not found")
				}
				return &Integer{Value: int64(start + idx)}
			},
		}, true
	case "count":
		return &Builtin{
			Name: "str.count",
			Fn: func(args ...Object) Object {
				sub, ok := args[0].(*String)
				if !ok {
					return NewTypeError("count() argument must be a string")
				}
				return &Integer{Value: int64(strings.Count(s.Value, sub.Value))}
			},
		}, true
	case "isdigit":
		return &Builtin{Name: "str.isdigit", Fn: func(args ...Object) Object {
			for _, r := range s.Value {
				if !unicode.IsDigit(r) {
					return False
				}
			}
			if len(s.Value) > 0 {
				return True
			}
			return False
		}}, true
	case "isalpha":
		return &Builtin{Name: "str.isalpha", Fn: func(args ...Object) Object {
			for _, r := range s.Value {
				if !unicode.IsLetter(r) {
					return False
				}
			}
			if len(s.Value) > 0 {
				return True
			}
			return False
		}}, true
	case "isalnum":
		return &Builtin{Name: "str.isalnum", Fn: func(args ...Object) Object {
			for _, r := range s.Value {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return False
				}
			}
			if len(s.Value) > 0 {
				return True
			}
			return False
		}}, true
	case "isspace":
		return &Builtin{Name: "str.isspace", Fn: func(args ...Object) Object {
			for _, r := range s.Value {
				if !unicode.IsSpace(r) {
					return False
				}
			}
			if len(s.Value) > 0 {
				return True
			}
			return False
		}}, true
	case "isupper":
		return &Builtin{Name: "str.isupper", Fn: func(args ...Object) Object {
			hasLetter := false
			for _, r := range s.Value {
				if unicode.IsLetter(r) {
					hasLetter = true
					if !unicode.IsUpper(r) {
						return False
					}
				}
			}
			if hasLetter {
				return True
			}
			return False
		}}, true
	case "islower":
		return &Builtin{Name: "str.islower", Fn: func(args ...Object) Object {
			hasLetter := false
			for _, r := range s.Value {
				if unicode.IsLetter(r) {
					hasLetter = true
					if !unicode.IsLower(r) {
						return False
					}
				}
			}
			if hasLetter {
				return True
			}
			return False
		}}, true
	case "istitle":
		return &Builtin{Name: "str.istitle", Fn: func(args ...Object) Object {
			words := strings.Fields(s.Value)
			if len(words) == 0 {
				return False
			}
			for _, w := range words {
				runes := []rune(w)
				if len(runes) == 0 {
					continue
				}
				if !unicode.IsUpper(runes[0]) {
					return False
				}
				for _, r := range runes[1:] {
					if unicode.IsLetter(r) && !unicode.IsLower(r) {
						return False
					}
				}
			}
			return True
		}}, true
	case "capitalize":
		return &Builtin{Name: "str.capitalize", Fn: func(args ...Object) Object {
			if len(s.Value) == 0 {
				return s
			}
			runes := []rune(s.Value)
			runes[0] = unicode.ToUpper(runes[0])
			for i := 1; i < len(runes); i++ {
				runes[i] = unicode.ToLower(runes[i])
			}
			return &String{Value: string(runes)}
		}}, true
	case "title":
		return &Builtin{Name: "str.title", Fn: func(args ...Object) Object {
			runes := []rune(s.Value)
			nextUpper := true
			for i, r := range runes {
				if unicode.IsLetter(r) {
					if nextUpper {
						runes[i] = unicode.ToUpper(r)
					} else {
						runes[i] = unicode.ToLower(r)
					}
					nextUpper = false
				} else {
					nextUpper = true
				}
			}
			return &String{Value: string(runes)}
		}}, true
	case "swapcase":
		return &Builtin{Name: "str.swapcase", Fn: func(args ...Object) Object {
			runes := []rune(s.Value)
			for i, r := range runes {
				if unicode.IsUpper(r) {
					runes[i] = unicode.ToLower(r)
				} else if unicode.IsLower(r) {
					runes[i] = unicode.ToUpper(r)
				}
			}
			return &String{Value: string(runes)}
		}}, true
	case "center":
		return &Builtin{Name: "str.center", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("center() takes at least 1 argument")
			}
			width, ok := args[0].(*Integer)
			if !ok {
				return NewTypeError("center() argument must be an integer")
			}
			fillChar := " "
			if len(args) >= 2 {
				if fc, ok := args[1].(*String); ok {
					fillChar = fc.Value
				}
			}
			strLen := len(s.Value)
			w := int(width.Value)
			if w <= strLen {
				return s
			}
			totalPad := w - strLen
			leftPad := totalPad / 2
			rightPad := totalPad - leftPad
			return &String{Value: strings.Repeat(fillChar, leftPad) + s.Value + strings.Repeat(fillChar, rightPad)}
		}}, true
	case "ljust":
		return &Builtin{Name: "str.ljust", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("ljust() takes at least 1 argument")
			}
			width, ok := args[0].(*Integer)
			if !ok {
				return NewTypeError("ljust() argument must be an integer")
			}
			fillChar := " "
			if len(args) >= 2 {
				if fc, ok := args[1].(*String); ok {
					fillChar = fc.Value
				}
			}
			w := int(width.Value)
			if w <= len(s.Value) {
				return s
			}
			return &String{Value: s.Value + strings.Repeat(fillChar, w-len(s.Value))}
		}}, true
	case "rjust":
		return &Builtin{Name: "str.rjust", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("rjust() takes at least 1 argument")
			}
			width, ok := args[0].(*Integer)
			if !ok {
				return NewTypeError("rjust() argument must be an integer")
			}
			fillChar := " "
			if len(args) >= 2 {
				if fc, ok := args[1].(*String); ok {
					fillChar = fc.Value
				}
			}
			w := int(width.Value)
			if w <= len(s.Value) {
				return s
			}
			return &String{Value: strings.Repeat(fillChar, w-len(s.Value)) + s.Value}
		}}, true
	case "zfill":
		return &Builtin{Name: "str.zfill", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("zfill() takes at least 1 argument")
			}
			width, ok := args[0].(*Integer)
			if !ok {
				return NewTypeError("zfill() argument must be an integer")
			}
			w := int(width.Value)
			if w <= len(s.Value) {
				return s
			}
			sign := ""
			val := s.Value
			if len(val) > 0 && (val[0] == '+' || val[0] == '-') {
				sign = string(val[0])
				val = val[1:]
			}
			return &String{Value: sign + strings.Repeat("0", w-len(s.Value)) + val}
		}}, true
	case "partition":
		return &Builtin{Name: "str.partition", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("partition() takes at least 1 argument")
			}
			sep, ok := args[0].(*String)
			if !ok {
				return NewTypeError("partition() argument must be a string")
			}
			idx := strings.Index(s.Value, sep.Value)
			if idx == -1 {
				return &Tuple{Elements: []Object{s, &String{Value: ""}, &String{Value: ""}}}
			}
			return &Tuple{Elements: []Object{
				&String{Value: s.Value[:idx]},
				&String{Value: sep.Value},
				&String{Value: s.Value[idx+len(sep.Value):]},
			}}
		}}, true
	case "rpartition":
		return &Builtin{Name: "str.rpartition", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("rpartition() takes at least 1 argument")
			}
			sep, ok := args[0].(*String)
			if !ok {
				return NewTypeError("rpartition() argument must be a string")
			}
			idx := strings.LastIndex(s.Value, sep.Value)
			if idx == -1 {
				return &Tuple{Elements: []Object{&String{Value: ""}, &String{Value: ""}, s}}
			}
			return &Tuple{Elements: []Object{
				&String{Value: s.Value[:idx]},
				&String{Value: sep.Value},
				&String{Value: s.Value[idx+len(sep.Value):]},
			}}
		}}, true
	case "encode":
		return &Builtin{Name: "str.encode", Fn: func(args ...Object) Object {
			encoding := "utf-8"
			if len(args) >= 1 {
				if e, ok := args[0].(*String); ok {
					encoding = e.Value
				}
			}
			switch encoding {
			case "utf-8", "utf8":
				return &Bytes{Value: []byte(s.Value)}
			case "ascii":
				return &Bytes{Value: []byte(s.Value)}
			default:
				return &Bytes{Value: []byte(s.Value)}
			}
		}}, true
	case "isdecimal":
		return &Builtin{Name: "str.isdecimal", Fn: func(args ...Object) Object {
			for _, r := range s.Value {
				if !unicode.IsDigit(r) {
					return False
				}
			}
			if len(s.Value) > 0 {
				return True
			}
			return False
		}}, true
	case "isnumeric":
		return &Builtin{Name: "str.isnumeric", Fn: func(args ...Object) Object {
			for _, r := range s.Value {
				if !unicode.IsNumber(r) {
					return False
				}
			}
			if len(s.Value) > 0 {
				return True
			}
			return False
		}}, true
	case "isidentifier":
		return &Builtin{Name: "str.isidentifier", Fn: func(args ...Object) Object {
			if len(s.Value) == 0 {
				return False
			}
			runes := []rune(s.Value)
			if !unicode.IsLetter(runes[0]) && runes[0] != '_' {
				return False
			}
			for _, r := range runes[1:] {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
					return False
				}
			}
			return True
		}}, true
	case "isprintable":
		return &Builtin{Name: "str.isprintable", Fn: func(args ...Object) Object {
			for _, r := range s.Value {
				if !unicode.IsPrint(r) {
					return False
				}
			}
			return True
		}}, true
	case "expandtabs":
		return &Builtin{Name: "str.expandtabs", Fn: func(args ...Object) Object {
			tabSize := 8
			if len(args) >= 1 {
				if i, ok := args[0].(*Integer); ok {
					tabSize = int(i.Value)
				}
			}
			return &String{Value: strings.ReplaceAll(s.Value, "\t", strings.Repeat(" ", tabSize))}
		}}, true
	case "translate":
		return &Builtin{Name: "str.translate", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("translate() takes at least 1 argument")
			}
			table, ok := args[0].(*Dict)
			if !ok {
				return NewTypeError("translate() argument must be a dict or None")
			}
			result := make([]byte, 0, len(s.Value))
			for _, r := range s.Value {
				if repl, ok := table.Get(&Integer{Value: int64(r)}); ok {
					// Check for delete flag
					if repl == nil {
						continue
					}
					if replStr, ok := repl.(*String); ok {
						result = append(result, replStr.Value...)
					} else if replInt, ok := repl.(*Integer); ok {
						result = append(result, byte(replInt.Value))
					}
				} else {
					result = append(result, byte(r))
				}
			}
			return &String{Value: string(result)}
		}}, true
	case "format_map":
		return &Builtin{Name: "str.format_map", Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return NewTypeError("format_map() takes at least 1 argument")
			}
			mapping, ok := args[0].(Object)
			if !ok {
				return NewTypeError("format_map() argument must be a mapping")
			}
			// Simple format_map implementation - replace {} with values from mapping
			result := s.Value
			for {
				openIdx := strings.Index(result, "{")
				if openIdx == -1 {
					break
				}
				closeIdx := strings.Index(result[openIdx:], "}")
				if closeIdx == -1 {
					break
				}
				closeIdx += openIdx
				key := result[openIdx+1 : closeIdx]
				var value string
				// Try to get from mapping
				if m, ok := mapping.(*Dict); ok {
					if v, ok := m.Get(&String{Value: key}); ok {
						value = v.Inspect()
					} else {
						value = ""
					}
				} else if m, ok := mapping.(*Instance); ok {
					if v, ok := m.GetAttr(key); ok {
						value = v.Inspect()
					} else {
						value = ""
					}
				}
				result = result[:openIdx] + value + result[closeIdx+1:]
			}
			return &String{Value: result}
		}}, true
	case "find":
		return &Builtin{
			Name: "str.find",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("find() takes at least 1 argument")
				}
				sub, ok := args[0].(*String)
				if !ok {
					return NewTypeError("find() argument must be a string")
				}
				start := 0
				end := len(s.Value)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(s.Value) {
					end = len(s.Value)
				}
				if start >= end {
					return &Integer{Value: -1}
				}
				idx := strings.Index(s.Value[start:end], sub.Value)
				if idx == -1 {
					return &Integer{Value: -1}
				}
				return &Integer{Value: int64(start + idx)}
			},
		}, true
	case "index":
		return &Builtin{
			Name: "str.index",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("index() takes at least 1 argument")
				}
				sub, ok := args[0].(*String)
				if !ok {
					return NewTypeError("index() argument must be a string")
				}
				start := 0
				end := len(s.Value)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(s.Value) {
					end = len(s.Value)
				}
				if start >= end {
					return NewValueError("substring not found")
				}
				idx := strings.Index(s.Value[start:end], sub.Value)
				if idx == -1 {
					return NewValueError("substring not found")
				}
				return &Integer{Value: int64(start + idx)}
			},
		}, true
	case "replace":
		return &Builtin{
			Name: "str.replace",
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
					if count, ok := args[2].(*Integer); ok {
						return &String{Value: strings.Replace(s.Value, old.Value, newStr.Value, int(count.Value))}
					}
				}
				return &String{Value: strings.ReplaceAll(s.Value, old.Value, newStr.Value)}
			},
		}, true
	case "split":
		return &Builtin{
			Name: "str.split",
			Fn: func(args ...Object) Object {
				if len(args) == 0 {
					// No args: split on whitespace
					parts := strings.Fields(s.Value)
					elements := make([]Object, len(parts))
					for i, p := range parts {
						elements[i] = &String{Value: p}
					}
					return &List{Elements: elements}
				}
				sep, ok := args[0].(*String)
				if !ok {
					return NewTypeError("split() argument must be a string or None")
				}
				if sep.Value == "" {
					return NewValueError("empty separator")
				}
				var parts []string
				if len(args) >= 2 {
					if maxsplit, ok := args[1].(*Integer); ok && maxsplit.Value >= 0 {
						parts = strings.SplitN(s.Value, sep.Value, int(maxsplit.Value)+1)
					} else {
						parts = strings.Split(s.Value, sep.Value)
					}
				} else {
					parts = strings.Split(s.Value, sep.Value)
				}
				elements := make([]Object, len(parts))
				for i, p := range parts {
					elements[i] = &String{Value: p}
				}
				return &List{Elements: elements}
			},
		}, true
	case "rsplit":
		return &Builtin{
			Name: "str.rsplit",
			Fn: func(args ...Object) Object {
				if len(args) == 0 {
					// No args: split on whitespace
					parts := strings.Fields(s.Value)
					elements := make([]Object, len(parts))
					for i, p := range parts {
						elements[i] = &String{Value: p}
					}
					return &List{Elements: elements}
				}
				sep, ok := args[0].(*String)
				if !ok {
					return NewTypeError("rsplit() argument must be a string or None")
				}
				if sep.Value == "" {
					return NewValueError("empty separator")
				}
				var parts []string
				if len(args) >= 2 {
					if maxsplit, ok := args[1].(*Integer); ok && maxsplit.Value >= 0 {
						// rsplit: split from right, which Go doesn't natively support
						// Use SplitN and then re-split the last element from the right
						parts = strings.Split(s.Value, sep.Value)
						if int(maxsplit.Value) < len(parts)-1 {
							// Need to rejoin the first len(parts)-1-maxsplit parts
							joinIdx := len(parts) - int(maxsplit.Value)
							result := make([]string, 0, int(maxsplit.Value)+1)
							result = append(result, strings.Join(parts[:joinIdx], sep.Value))
							result = append(result, parts[joinIdx:]...)
							parts = result
						}
					} else {
						parts = strings.Split(s.Value, sep.Value)
					}
				} else {
					parts = strings.Split(s.Value, sep.Value)
				}
				elements := make([]Object, len(parts))
				for i, p := range parts {
					elements[i] = &String{Value: p}
				}
				return &List{Elements: elements}
			},
		}, true
	case "splitlines":
		return &Builtin{
			Name: "str.splitlines",
			Fn: func(args ...Object) Object {
				keepends := false
				if len(args) >= 1 {
					if b, ok := args[0].(*Boolean); ok {
						keepends = b.Value
					}
				}
				var lines []string
				current := ""
				for i := 0; i < len(s.Value); i++ {
					if s.Value[i] == '\r' {
						if i+1 < len(s.Value) && s.Value[i+1] == '\n' {
							if keepends {
								current += "\r\n"
							}
							lines = append(lines, current)
							current = ""
							i++ // skip \n
						} else {
							if keepends {
								current += "\r"
							}
							lines = append(lines, current)
							current = ""
						}
					} else if s.Value[i] == '\n' {
						if keepends {
							current += "\n"
						}
						lines = append(lines, current)
						current = ""
					} else {
						current += string(s.Value[i])
					}
				}
				if current != "" {
					lines = append(lines, current)
				}
				elements := make([]Object, len(lines))
				for i, l := range lines {
					elements[i] = &String{Value: l}
				}
				return &List{Elements: elements}
			},
		}, true
	case "join":
		return &Builtin{
			Name: "str.join",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("join() takes exactly 1 argument")
				}
				var elements []string
				switch iter := args[0].(type) {
				case *List:
					for _, el := range iter.Elements {
						elements = append(elements, el.Inspect())
					}
				case *Tuple:
					for _, el := range iter.Elements {
						elements = append(elements, el.Inspect())
					}
				default:
					return NewTypeError("can only join an iterable")
				}
				return &String{Value: strings.Join(elements, s.Value)}
			},
		}, true
	case "strip":
		return &Builtin{
			Name: "str.strip",
			Fn: func(args ...Object) Object {
				chars := " \t\n\r\f\v"
				if len(args) >= 1 {
					if c, ok := args[0].(*String); ok {
						chars = c.Value
					}
				}
				return &String{Value: strings.Trim(s.Value, chars)}
			},
		}, true
	case "lstrip":
		return &Builtin{
			Name: "str.lstrip",
			Fn: func(args ...Object) Object {
				chars := " \t\n\r\f\v"
				if len(args) >= 1 {
					if c, ok := args[0].(*String); ok {
						chars = c.Value
					}
				}
				return &String{Value: strings.TrimLeft(s.Value, chars)}
			},
		}, true
	case "rstrip":
		return &Builtin{
			Name: "str.rstrip",
			Fn: func(args ...Object) Object {
				chars := " \t\n\r\f\v"
				if len(args) >= 1 {
					if c, ok := args[0].(*String); ok {
						chars = c.Value
					}
				}
				return &String{Value: strings.TrimRight(s.Value, chars)}
			},
		}, true
	case "upper":
		return &Builtin{Name: "str.upper", Fn: func(args ...Object) Object {
			return &String{Value: strings.ToUpper(s.Value)}
		}}, true
	case "lower":
		return &Builtin{Name: "str.lower", Fn: func(args ...Object) Object {
			return &String{Value: strings.ToLower(s.Value)}
		}}, true
	case "startswith":
		return &Builtin{
			Name: "str.startswith",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("startswith() takes at least 1 argument")
				}
				start := 0
				end := len(s.Value)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(s.Value) {
					end = len(s.Value)
				}
				if start > end {
					return False
				}
				substr := s.Value[start:end]
				switch prefix := args[0].(type) {
				case *String:
					return &Boolean{Value: strings.HasPrefix(substr, prefix.Value)}
				case *Tuple:
					for _, el := range prefix.Elements {
						if p, ok := el.(*String); ok {
							if strings.HasPrefix(substr, p.Value) {
								return True
							}
						}
					}
					return False
				default:
					return NewTypeError("startswith() argument must be a string or a tuple of strings")
				}
			},
		}, true
	case "endswith":
		return &Builtin{
			Name: "str.endswith",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("endswith() takes at least 1 argument")
				}
				start := 0
				end := len(s.Value)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(s.Value) {
					end = len(s.Value)
				}
				if start > end {
					return False
				}
				substr := s.Value[start:end]
				switch suffix := args[0].(type) {
				case *String:
					return &Boolean{Value: strings.HasSuffix(substr, suffix.Value)}
				case *Tuple:
					for _, el := range suffix.Elements {
						if suf, ok := el.(*String); ok {
							if strings.HasSuffix(substr, suf.Value) {
								return True
							}
						}
					}
					return False
				default:
					return NewTypeError("endswith() argument must be a string or a tuple of strings")
				}
			},
		}, true
	case "format":
		return &Builtin{
			Name: "str.format",
			Fn: func(args ...Object) Object {
				result := s.Value
				// Collect keyword arguments from the last argument if it's a Dict
				kwargs := make(map[string]string)
				posArgs := args
				if len(args) > 0 {
					if d, ok := args[len(args)-1].(*Dict); ok {
						// Check if this looks like kwargs (string keys)
						isKwargs := true
						for _, keyStr := range d.KeyOrder {
							if _, ok := d.Keys[keyStr].(*String); !ok {
								isKwargs = false
								break
							}
						}
						if isKwargs && len(d.KeyOrder) > 0 {
							for _, keyStr := range d.KeyOrder {
								if sk, ok := d.Keys[keyStr].(*String); ok {
									kwargs[sk.Value] = d.Pairs[keyStr].Inspect()
								}
							}
							posArgs = args[:len(args)-1]
						}
					}
				}
				// Replace positional {0}, {1}, etc.
				for i, arg := range posArgs {
					result = strings.ReplaceAll(result, "{"+fmt.Sprintf("%d", i)+"}", arg.Inspect())
				}
				// Replace keyword {name}
				for key, val := range kwargs {
					result = strings.ReplaceAll(result, "{"+key+"}", val)
				}
				return &String{Value: result}
			},
		}, true
	case "casefold":
		return &Builtin{Name: "str.casefold", Fn: func(args ...Object) Object {
			return &String{Value: strings.ToLower(s.Value)}
		}}, true
	case "maketrans":
		return &Builtin{
			Name: "str.maketrans",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("maketrans() takes at least 1 argument")
				}
				result := NewDict()
				if len(args) == 1 {
					// dict argument
					d, ok := args[0].(*Dict)
					if !ok {
						return NewTypeError("maketrans() argument must be a dict")
					}
					for _, keyStr := range d.KeyOrder {
						k := d.Keys[keyStr]
						v := d.Pairs[keyStr]
						if ki, ok := k.(*Integer); ok {
							result.Set(ki, v)
						} else if ks, ok := k.(*String); ok {
							if len(ks.Value) == 1 {
								result.Set(&Integer{Value: int64(ks.Value[0])}, v)
							}
						}
					}
				} else if len(args) >= 2 {
					from, ok1 := args[0].(*String)
					to, ok2 := args[1].(*String)
					if !ok1 || !ok2 {
						return NewTypeError("maketrans() arguments must be strings")
					}
					if len(from.Value) != len(to.Value) {
						return NewValueError("maketrans() arguments must have same length")
					}
					for i := 0; i < len(from.Value); i++ {
						result.Set(&Integer{Value: int64(from.Value[i])}, &Integer{Value: int64(to.Value[i])})
					}
					if len(args) >= 3 {
						if del, ok := args[2].(*String); ok {
							for i := 0; i < len(del.Value); i++ {
								result.Set(&Integer{Value: int64(del.Value[i])}, &None{})
							}
						}
					}
				}
				return result
			},
		}, true
	}
	return nil, false
}

type Bytes struct {
	Value []byte
}

func NewBytes(data []byte) *Bytes {
	return &Bytes{Value: data}
}

func (b *Bytes) Type() ObjectType { return BYTES_OBJ }
func (b *Bytes) Inspect() string  { return fmt.Sprintf("b'%s'", string(b.Value)) }

func (b *Bytes) GetAttr(name string) (Object, bool) {
	switch name {
	case "decode":
		return &Builtin{
			Name: "bytes.decode",
			Fn: func(args ...Object) Object {
				return &String{Value: string(b.Value)}
			},
		}, true
	case "hex":
		return &Builtin{
			Name: "bytes.hex",
			Fn: func(args ...Object) Object {
				return &String{Value: fmt.Sprintf("%x", b.Value)}
			},
		}, true
	case "len":
		return &Builtin{
			Name: "bytes.len",
			Fn: func(args ...Object) Object {
				return &Integer{Value: int64(len(b.Value))}
			},
		}, true
	case "__len__":
		return &Builtin{
			Name: "bytes.__len__",
			Fn: func(args ...Object) Object {
				return &Integer{Value: int64(len(b.Value))}
			},
		}, true
	case "__class__":
		return &String{Value: "bytes"}, true
	}
	return nil, false
}

type None struct{}

func (n *None) Type() ObjectType { return NONE_OBJ }
func (n *None) Inspect() string  { return "None" }

type Ellipsis struct{}

func (e *Ellipsis) Type() ObjectType { return ELLIPSIS_OBJ }
func (e *Ellipsis) Inspect() string  { return "Ellipsis" }

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

func (t *Tuple) GetAttr(name string) (Object, bool) {
	switch name {
	case "count":
		return &Builtin{
			Name: "tuple.count",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("count() takes exactly one argument (%d given)", len(args))
				}
				cnt := 0
				for _, el := range t.Elements {
					if Equal(el, args[0]) {
						cnt++
					}
				}
				return &Integer{Value: int64(cnt)}
			},
		}, true
	case "index":
		return &Builtin{
			Name: "tuple.index",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("index() takes at least 1 argument (%d given)", len(args))
				}
				start := 0
				end := len(t.Elements)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
						if start < 0 {
							start = len(t.Elements) + start
						}
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
						if end < 0 {
							end = len(t.Elements) + end
						}
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(t.Elements) {
					end = len(t.Elements)
				}
				for i := start; i < end; i++ {
					if Equal(t.Elements[i], args[0]) {
						return &Integer{Value: int64(i)}
					}
				}
				return NewValueError("tuple.index(x): x not in tuple")
			},
		}, true
	}
	return nil, false
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

func (l *List) GetAttr(name string) (Object, bool) {
	switch name {
	case "sort":
		return &Builtin{
			Name: "list.sort",
			Fn: func(args ...Object) Object {
				var keyFn Object
				reverse := false
				if len(args) >= 1 {
					if b, ok := args[0].(*Boolean); ok {
						reverse = b.Value
					} else if _, ok := args[0].(*None); !ok {
						keyFn = args[0]
					}
				}
				if len(args) >= 2 {
					if b, ok := args[1].(*Boolean); ok {
						reverse = b.Value
					}
				}
				sort.SliceStable(l.Elements, func(i, j int) bool {
					var aVal, bVal Object
					if keyFn != nil {
						aVal = CallFunction(keyFn, l.Elements[i])
						bVal = CallFunction(keyFn, l.Elements[j])
					} else {
						aVal = l.Elements[i]
						bVal = l.Elements[j]
					}
					cmp := compareObjectsForSort(aVal, bVal)
					if reverse {
						return cmp > 0
					}
					return cmp < 0
				})
				return None_
			},
		}, true
	case "__imul__":
		return &Builtin{
			Name: "list.__imul__",
			Fn: func(args ...Object) Object {
				n, ok := args[0].(*Integer)
				if !ok {
					return NewTypeError("can't multiply sequence by non-int of type '%s'", args[0].Type())
				}
				if n.Value <= 0 {
					l.Elements = []Object{}
					return l
				}
				original := make([]Object, len(l.Elements))
				copy(original, l.Elements)
				for i := int64(1); i < n.Value; i++ {
					l.Elements = append(l.Elements, original...)
				}
				return l
			},
		}, true
	case "__iadd__":
		return &Builtin{
			Name: "list.__iadd__",
			Fn: func(args ...Object) Object {
				switch other := args[0].(type) {
				case *List:
					l.Elements = append(l.Elements, other.Elements...)
				default:
					return NewTypeError("'%s' object is not iterable", args[0].Type())
				}
				return l
			},
		}, true
	case "append":
		return &Builtin{
			Name: "list.append",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("append() takes exactly one argument (%d given)", len(args))
				}
				l.Elements = append(l.Elements, args[0])
				return None_
			},
		}, true
	case "extend":
		return &Builtin{
			Name: "list.extend",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("extend() takes exactly one argument (%d given)", len(args))
				}
				switch other := args[0].(type) {
				case *List:
					l.Elements = append(l.Elements, other.Elements...)
				case *Tuple:
					l.Elements = append(l.Elements, other.Elements...)
				default:
					return NewTypeError("'%s' object is not iterable", args[0].Type())
				}
				return None_
			},
		}, true
	case "insert":
		return &Builtin{
			Name: "list.insert",
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return NewTypeError("insert() takes exactly 2 arguments (%d given)", len(args))
				}
				idx, ok := args[0].(*Integer)
				if !ok {
					return NewTypeError("'%s' object cannot be interpreted as an integer", args[0].Type())
				}
				l.Insert(int(idx.Value), args[1])
				return None_
			},
		}, true
	case "remove":
		return &Builtin{
			Name: "list.remove",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("remove() takes exactly one argument (%d given)", len(args))
				}
				err := l.Remove(args[0])
				if err != nil {
					return NewValueError("list.remove(x): x not in list")
				}
				return None_
			},
		}, true
	case "pop":
		return &Builtin{
			Name: "list.pop",
			Fn: func(args ...Object) Object {
				if len(args) > 1 {
					return NewTypeError("pop() takes at most 1 argument (%d given)", len(args))
				}
				idx := -1
				if len(args) == 1 {
					i, ok := args[0].(*Integer)
					if !ok {
						return NewTypeError("'%s' object cannot be interpreted as an integer", args[0].Type())
					}
					idx = int(i.Value)
				}
				obj, err := l.Pop(idx)
				if err != nil {
					return NewIndexError("pop index out of range")
				}
				return obj
			},
		}, true
	case "clear":
		return &Builtin{
			Name: "list.clear",
			Fn: func(args ...Object) Object {
				l.Elements = []Object{}
				return None_
			},
		}, true
	case "index":
		return &Builtin{
			Name: "list.index",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("index() takes at least 1 argument (%d given)", len(args))
				}
				start := 0
				end := len(l.Elements)
				if len(args) >= 2 {
					if i, ok := args[1].(*Integer); ok {
						start = int(i.Value)
						if start < 0 {
							start = len(l.Elements) + start
						}
					}
				}
				if len(args) >= 3 {
					if i, ok := args[2].(*Integer); ok {
						end = int(i.Value)
						if end < 0 {
							end = len(l.Elements) + end
						}
					}
				}
				if start < 0 {
					start = 0
				}
				if end > len(l.Elements) {
					end = len(l.Elements)
				}
				for i := start; i < end; i++ {
					if Equal(l.Elements[i], args[0]) {
						return &Integer{Value: int64(i)}
					}
				}
				return NewValueError("%s is not in list", args[0].Inspect())
			},
		}, true
	case "count":
		return &Builtin{
			Name: "list.count",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("count() takes exactly one argument (%d given)", len(args))
				}
				cnt := 0
				for _, el := range l.Elements {
					if Equal(el, args[0]) {
						cnt++
					}
				}
				return &Integer{Value: int64(cnt)}
			},
		}, true
	case "reverse":
		return &Builtin{
			Name: "list.reverse",
			Fn: func(args ...Object) Object {
				l.Reverse()
				return None_
			},
		}, true
	case "copy":
		return &Builtin{
			Name: "list.copy",
			Fn: func(args ...Object) Object {
				elements := make([]Object, len(l.Elements))
				copy(elements, l.Elements)
				return &List{Elements: elements}
			},
		}, true
	}
	return nil, false
}

func compareObjectsForSort(a, b Object) int {
	switch a := a.(type) {
	case *Integer:
		if bInt, ok := b.(*Integer); ok {
			if a.Value < bInt.Value {
				return -1
			}
			if a.Value > bInt.Value {
				return 1
			}
			return 0
		}
	case *Float:
		if bFloat, ok := b.(*Float); ok {
			if a.Value < bFloat.Value {
				return -1
			}
			if a.Value > bFloat.Value {
				return 1
			}
			return 0
		}
	case *String:
		if bStr, ok := b.(*String); ok {
			if a.Value < bStr.Value {
				return -1
			}
			if a.Value > bStr.Value {
				return 1
			}
			return 0
		}
	case *Boolean:
		aInt := int64(0)
		if a.Value {
			aInt = 1
		}
		if bBool, ok := b.(*Boolean); ok {
			bInt := int64(0)
			if bBool.Value {
				bInt = 1
			}
			if aInt < bInt {
				return -1
			}
			if aInt > bInt {
				return 1
			}
			return 0
		}
	}
	return 0
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

// Union returns a new set with elements from both s and other
func (s *Set) Union(other *Set) *Set {
	result := NewSet()
	for _, elem := range s.Elements {
		result.Add(elem)
	}
	for _, elem := range other.Elements {
		result.Add(elem)
	}
	return result
}

// Intersection returns a new set with elements common to s and other
func (s *Set) Intersection(other *Set) *Set {
	result := NewSet()
	for _, elem := range s.Elements {
		if other.Contains(elem) {
			result.Add(elem)
		}
	}
	return result
}

// Difference returns a new set with elements in s but not in other
func (s *Set) Difference(other *Set) *Set {
	result := NewSet()
	for _, elem := range s.Elements {
		if !other.Contains(elem) {
			result.Add(elem)
		}
	}
	return result
}

// SymmetricDifference returns a new set with elements in either s or other but not both
func (s *Set) SymmetricDifference(other *Set) *Set {
	result := NewSet()
	for _, elem := range s.Elements {
		if !other.Contains(elem) {
			result.Add(elem)
		}
	}
	for _, elem := range other.Elements {
		if !s.Contains(elem) {
			result.Add(elem)
		}
	}
	return result
}

func (s *Set) ToSlice() []Object {
	slice := make([]Object, 0, len(s.Elements))
	for _, v := range s.Elements {
		slice = append(slice, v)
	}
	return slice
}

func (s *Set) GetAttr(name string) (Object, bool) {
	switch name {
	case "union":
		return &Builtin{
			Name: "set.union",
			Fn: func(args ...Object) Object {
				result := NewSet()
				for k, v := range s.Elements {
					result.Elements[k] = v
					result.Keys[k] = v
				}
				for _, arg := range args {
					switch other := arg.(type) {
					case *Set:
						for k, v := range other.Elements {
							result.Elements[k] = v
							result.Keys[k] = v
						}
					case *List:
						for _, elem := range other.Elements {
							result.Add(elem)
						}
					case *Tuple:
						for _, elem := range other.Elements {
							result.Add(elem)
						}
					}
				}
				return result
			},
		}, true
	case "intersection":
		return &Builtin{
			Name: "set.intersection",
			Fn: func(args ...Object) Object {
				result := NewSet()
				if len(args) == 0 {
					return result
				}
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("intersection() argument must be a set")
				}
				for k, v := range s.Elements {
					if _, ok := other.Elements[k]; ok {
						result.Elements[k] = v
						result.Keys[k] = v
					}
				}
				return result
			},
		}, true
	case "difference":
		return &Builtin{
			Name: "set.difference",
			Fn: func(args ...Object) Object {
				result := NewSet()
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("difference() argument must be a set")
				}
				for k, v := range s.Elements {
					if _, ok := other.Elements[k]; !ok {
						result.Elements[k] = v
						result.Keys[k] = v
					}
				}
				return result
			},
		}, true
	case "symmetric_difference":
		return &Builtin{
			Name: "set.symmetric_difference",
			Fn: func(args ...Object) Object {
				result := NewSet()
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("symmetric_difference() argument must be a set")
				}
				for k, v := range s.Elements {
					if _, ok := other.Elements[k]; !ok {
						result.Elements[k] = v
						result.Keys[k] = v
					}
				}
				for k, v := range other.Elements {
					if _, ok := s.Elements[k]; !ok {
						result.Elements[k] = v
						result.Keys[k] = v
					}
				}
				return result
			},
		}, true
	case "issubset":
		return &Builtin{
			Name: "set.issubset",
			Fn: func(args ...Object) Object {
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("issubset() argument must be a set")
				}
				for k := range s.Elements {
					if _, ok := other.Elements[k]; !ok {
						return False
					}
				}
				return True
			},
		}, true
	case "issuperset":
		return &Builtin{
			Name: "set.issuperset",
			Fn: func(args ...Object) Object {
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("issuperset() argument must be a set")
				}
				for k := range other.Elements {
					if _, ok := s.Elements[k]; !ok {
						return False
					}
				}
				return True
			},
		}, true
	case "update":
		return &Builtin{
			Name: "set.update",
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					switch other := arg.(type) {
					case *Set:
						for k, v := range other.Elements {
							s.Elements[k] = v
							s.Keys[k] = v
						}
					case *List:
						for _, elem := range other.Elements {
							s.Add(elem)
						}
					}
				}
				return None_
			},
		}, true
	case "copy":
		return &Builtin{
			Name: "set.copy",
			Fn: func(args ...Object) Object {
				result := NewSet()
				for k, v := range s.Elements {
					result.Elements[k] = v
					result.Keys[k] = v
				}
				return result
			},
		}, true
	case "__isub__":
		return &Builtin{
			Name: "set.__isub__",
			Fn: func(args ...Object) Object {
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("unsupported operand type(s) for -=: 'set' and '%s'", args[0].Type())
				}
				for k := range other.Elements {
					delete(s.Elements, k)
					delete(s.Keys, k)
				}
				return s
			},
		}, true
	case "difference_update":
		return &Builtin{
			Name: "set.difference_update",
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					switch other := arg.(type) {
					case *Set:
						for k := range other.Elements {
							delete(s.Elements, k)
							delete(s.Keys, k)
						}
					case *List:
						for _, elem := range other.Elements {
							key := s.HashKey(elem)
							delete(s.Elements, key)
							delete(s.Keys, key)
						}
					case *Tuple:
						for _, elem := range other.Elements {
							key := s.HashKey(elem)
							delete(s.Elements, key)
							delete(s.Keys, key)
						}
					default:
						return NewTypeError("'%s' object is not iterable", arg.Type())
					}
				}
				return None_
			},
		}, true
	case "__ior__":
		return &Builtin{
			Name: "set.__ior__",
			Fn: func(args ...Object) Object {
				switch other := args[0].(type) {
				case *Set:
					for k, v := range other.Elements {
						s.Elements[k] = v
						s.Keys[k] = v
					}
				case *List:
					for _, elem := range other.Elements {
						s.Add(elem)
					}
				case *Tuple:
					for _, elem := range other.Elements {
						s.Add(elem)
					}
				default:
					return NewTypeError("unsupported operand type(s) for |=: 'set' and '%s'", args[0].Type())
				}
				return s
			},
		}, true
	case "__iand__":
		return &Builtin{
			Name: "set.__iand__",
			Fn: func(args ...Object) Object {
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("unsupported operand type(s) for &=: 'set' and '%s'", args[0].Type())
				}
				for k := range s.Elements {
					if _, ok := other.Elements[k]; !ok {
						delete(s.Elements, k)
						delete(s.Keys, k)
					}
				}
				return s
			},
		}, true
	case "__ixor__":
		return &Builtin{
			Name: "set.__ixor__",
			Fn: func(args ...Object) Object {
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("unsupported operand type(s) for ^=: 'set' and '%s'", args[0].Type())
				}
				// Elements in either but not both: remove common, add exclusive from other
				for k, v := range other.Elements {
					if _, exists := s.Elements[k]; exists {
						delete(s.Elements, k)
						delete(s.Keys, k)
					} else {
						s.Elements[k] = v
						s.Keys[k] = v
					}
				}
				return s
			},
		}, true
	case "add":
		return &Builtin{
			Name: "set.add",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("add() takes exactly one argument (%d given)", len(args))
				}
				s.Add(args[0])
				return None_
			},
		}, true
	case "remove":
		return &Builtin{
			Name: "set.remove",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("remove() takes exactly one argument (%d given)", len(args))
				}
				key := s.HashKey(args[0])
				if _, ok := s.Elements[key]; !ok {
					return NewKeyError("%s", args[0].Inspect())
				}
				delete(s.Elements, key)
				delete(s.Keys, key)
				return None_
			},
		}, true
	case "discard":
		return &Builtin{
			Name: "set.discard",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewTypeError("discard() takes exactly one argument (%d given)", len(args))
				}
				key := s.HashKey(args[0])
				delete(s.Elements, key)
				delete(s.Keys, key)
				return None_
			},
		}, true
	case "pop":
		return &Builtin{
			Name: "set.pop",
			Fn: func(args ...Object) Object {
				if len(s.Elements) == 0 {
					return NewKeyError("pop from an empty set")
				}
				var elem Object
				var key string
				for k, v := range s.Elements {
					elem = v
					key = k
					break
				}
				delete(s.Elements, key)
				delete(s.Keys, key)
				return elem
			},
		}, true
	case "clear":
		return &Builtin{
			Name: "set.clear",
			Fn: func(args ...Object) Object {
				s.Elements = make(map[string]Object)
				s.Keys = make(map[string]Object)
				return None_
			},
		}, true
	case "intersection_update":
		return &Builtin{
			Name: "set.intersection_update",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("intersection_update() takes at least one argument (%d given)", len(args))
				}
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("'%s' object is not a set", args[0].Type())
				}
				for k := range s.Elements {
					if _, ok := other.Elements[k]; !ok {
						delete(s.Elements, k)
						delete(s.Keys, k)
					}
				}
				return None_
			},
		}, true
	case "symmetric_difference_update":
		return &Builtin{
			Name: "set.symmetric_difference_update",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("symmetric_difference_update() takes at least one argument (%d given)", len(args))
				}
				other, ok := args[0].(*Set)
				if !ok {
					return NewTypeError("'%s' object is not a set", args[0].Type())
				}
				for k, v := range other.Elements {
					if _, exists := s.Elements[k]; exists {
						delete(s.Elements, k)
						delete(s.Keys, k)
					} else {
						s.Elements[k] = v
						s.Keys[k] = v
					}
				}
				return None_
			},
		}, true
	}
	return nil, false
}

type Frozenset struct {
	Elements map[string]Object
	Keys     map[string]Object
}

func NewFrozenset(elements map[string]Object, keys map[string]Object) *Frozenset {
	return &Frozenset{Elements: elements, Keys: keys}
}

func (fs *Frozenset) Type() ObjectType { return FROZENSET_OBJ }
func (fs *Frozenset) Inspect() string {
	result := "frozenset({"
	first := true
	for _, key := range fs.Keys {
		if !first {
			result += ", "
		}
		result += key.Inspect()
		first = false
	}
	result += "})"
	return result
}

func (fs *Frozenset) GetAttr(name string) (Object, bool) {
	switch name {
	case "union":
		return &Builtin{Name: "frozenset.union", Fn: func(args ...Object) Object {
			result := NewSet()
			for k, v := range fs.Elements { result.Elements[k] = v; result.Keys[k] = v }
			for _, arg := range args {
				switch other := arg.(type) {
				case *Set:
					for k, v := range other.Elements { result.Elements[k] = v; result.Keys[k] = v }
				case *Frozenset:
					for k, v := range other.Elements { result.Elements[k] = v; result.Keys[k] = v }
				}
			}
			return NewFrozenset(result.Elements, result.Keys)
		}}, true
	case "intersection":
		return &Builtin{Name: "frozenset.intersection", Fn: func(args ...Object) Object {
			result := NewSet()
			if len(args) == 0 { return NewFrozenset(result.Elements, result.Keys) }
			other, ok := args[0].(*Set)
			if !ok { if fs2, ok2 := args[0].(*Frozenset); ok2 { other = &Set{Elements: fs2.Elements, Keys: fs2.Keys} } }
			if other == nil { return NewTypeError("intersection() argument must be a set") }
			for k, v := range fs.Elements {
				if _, ok := other.Elements[k]; ok { result.Elements[k] = v; result.Keys[k] = v }
			}
			return NewFrozenset(result.Elements, result.Keys)
		}}, true
	case "issubset":
		return &Builtin{Name: "frozenset.issubset", Fn: func(args ...Object) Object {
			if len(args) < 1 { return NewTypeError("issubset() takes at least 1 argument") }
			var otherElements map[string]Object
			switch o := args[0].(type) {
			case *Set: otherElements = o.Elements
			case *Frozenset: otherElements = o.Elements
			default: return NewTypeError("issubset() argument must be a set")
			}
			for k := range fs.Elements {
				if _, ok := otherElements[k]; !ok { return False }
			}
			return True
		}}, true
	case "issuperset":
		return &Builtin{Name: "frozenset.issuperset", Fn: func(args ...Object) Object {
			if len(args) < 1 { return NewTypeError("issuperset() takes at least 1 argument") }
			var otherElements map[string]Object
			switch o := args[0].(type) {
			case *Set: otherElements = o.Elements
			case *Frozenset: otherElements = o.Elements
			default: return NewTypeError("issuperset() argument must be a set")
			}
			for k := range otherElements {
				if _, ok := fs.Elements[k]; !ok { return False }
			}
			return True
		}}, true
	case "copy":
		return &Builtin{Name: "frozenset.copy", Fn: func(args ...Object) Object {
			newElements := make(map[string]Object)
			newKeys := make(map[string]Object)
			for k, v := range fs.Elements { newElements[k] = v; newKeys[k] = v }
			return NewFrozenset(newElements, newKeys)
		}}, true
	case "__class__":
		return &String{Value: "frozenset"}, true
	}
	return nil, false
}

type Dict struct {
	Pairs    map[string]Object
	Keys     map[string]Object
	KeyOrder []string
}

func NewDict() *Dict {
	return &Dict{
		Pairs:    make(map[string]Object),
		Keys:     make(map[string]Object),
		KeyOrder: make([]string, 0),
	}
}

func NewDictWithCapacity(n int) *Dict {
	return &Dict{
		Pairs:    make(map[string]Object, n),
		Keys:     make(map[string]Object, n),
		KeyOrder: make([]string, 0, n),
	}
}

func (d *Dict) Type() ObjectType { return DICT_OBJ }
func (d *Dict) Inspect() string {
	result := "{"
	first := true
	for _, keyStr := range d.KeyOrder {
		if !first {
			result += ", "
		}
		first = false
		key := d.Keys[keyStr]
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
	if _, exists := d.Pairs[keyStr]; !exists {
		d.KeyOrder = append(d.KeyOrder, keyStr)
	}
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
	if _, exists := d.Pairs[keyStr]; exists {
		delete(d.Pairs, keyStr)
		delete(d.Keys, keyStr)
		for i, k := range d.KeyOrder {
			if k == keyStr {
				d.KeyOrder = append(d.KeyOrder[:i], d.KeyOrder[i+1:]...)
				break
			}
		}
	}
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
	for _, key := range d.KeyOrder {
		values = append(values, d.Pairs[key])
	}
	return values
}

func (d *Dict) GetAttr(name string) (Object, bool) {
	switch name {
	case "fromkeys":
		return &Builtin{
			Name: "dict.fromkeys",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("fromkeys() takes at least 1 argument")
				}
				var keys []Object
				switch iter := args[0].(type) {
				case *List:
					keys = iter.Elements
				case *Tuple:
					keys = iter.Elements
				case *Set:
					keys = iter.ToSlice()
				default:
					return NewTypeError("'%s' object is not iterable", args[0].Type())
				}
				defaultValue := Object(None_)
				if len(args) >= 2 {
					defaultValue = args[1]
				}
				result := NewDict()
				for _, key := range keys {
					result.Set(key, defaultValue)
				}
				return result
			},
		}, true
	case "update":
		return &Builtin{
			Name: "dict.update",
			Fn: func(args ...Object) Object {
				// Handle dict/iterable argument
				if len(args) >= 1 {
					switch other := args[0].(type) {
					case *Dict:
						for _, keyStr := range other.KeyOrder {
							key := other.Keys[keyStr]
							val := other.Pairs[keyStr]
							d.Set(key, val)
						}
					case *List:
						for _, elem := range other.Elements {
							if pair, ok := elem.(*Tuple); ok && len(pair.Elements) == 2 {
								d.Set(pair.Elements[0], pair.Elements[1])
							}
						}
					case *Tuple:
						for _, elem := range other.Elements {
							if pair, ok := elem.(*Tuple); ok && len(pair.Elements) == 2 {
								d.Set(pair.Elements[0], pair.Elements[1])
							}
						}
					}
				}
				// Handle keyword arguments (passed as a Dict named "kwargs" after positional args)
				if len(args) >= 2 {
					if kwargs, ok := args[len(args)-1].(*Dict); ok {
						for _, keyStr := range kwargs.KeyOrder {
							key := kwargs.Keys[keyStr]
							val := kwargs.Pairs[keyStr]
							d.Set(key, val)
						}
					}
				}
				return None_
			},
		}, true
	case "setdefault":
		return &Builtin{
			Name: "dict.setdefault",
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
	case "popitem":
		return &Builtin{
			Name: "dict.popitem",
			Fn: func(args ...Object) Object {
				if len(d.KeyOrder) == 0 {
					return NewKeyError("dictionary is empty")
				}
				keyStr := d.KeyOrder[len(d.KeyOrder)-1]
				key := d.Keys[keyStr]
				val := d.Pairs[keyStr]
				d.Delete(key)
				return &Tuple{Elements: []Object{key, val}}
			},
		}, true
	case "pop":
		return &Builtin{
			Name: "dict.pop",
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
				return NewKeyError("%s", args[0].Inspect())
			},
		}, true
	case "get":
		return &Builtin{
			Name: "dict.get",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("get() takes at least 1 argument")
				}
				var defaultVal Object = None_
				if len(args) >= 2 {
					defaultVal = args[1]
				}
				if val, ok := d.Get(args[0]); ok {
					return val
				}
				return defaultVal
			},
		}, true
	case "clear":
		return &Builtin{
			Name: "dict.clear",
			Fn: func(args ...Object) Object {
				d.Pairs = make(map[string]Object)
				d.Keys = make(map[string]Object)
				d.KeyOrder = make([]string, 0)
				return None_
			},
		}, true
	case "copy":
		return &Builtin{
			Name: "dict.copy",
			Fn: func(args ...Object) Object {
				newDict := NewDict()
				for _, keyStr := range d.KeyOrder {
					key := d.Keys[keyStr]
					val := d.Pairs[keyStr]
					newDict.Set(key, val)
				}
				return newDict
			},
		}, true
	case "keys":
		return &Builtin{
			Name: "dict.keys",
			Fn: func(args ...Object) Object {
				return NewDictKeys(d)
			},
		}, true
	case "values":
		return &Builtin{
			Name: "dict.values",
			Fn: func(args ...Object) Object {
				return NewDictValues(d)
			},
		}, true
	case "items":
		return &Builtin{
			Name: "dict.items",
			Fn: func(args ...Object) Object {
				return NewDictItems(d)
			},
		}, true
	}
	return nil, false
}

// DictKeys is a view object for dict.keys()
type DictKeys struct {
	Dict *Dict
}

func NewDictKeys(d *Dict) *DictKeys {
	return &DictKeys{Dict: d}
}

func (dk *DictKeys) Type() ObjectType { return DICT_KEYS_OBJ }
func (dk *DictKeys) Inspect() string {
	elements := make([]string, len(dk.Dict.KeyOrder))
	for i, keyStr := range dk.Dict.KeyOrder {
		key := dk.Dict.Keys[keyStr]
		elements[i] = key.Inspect()
	}
	return fmt.Sprintf("dict_keys([%s])", strings.Join(elements, ", "))
}

func (dk *DictKeys) Len() int64 {
	return int64(len(dk.Dict.KeyOrder))
}

func (dk *DictKeys) ToList() []Object {
	keys := make([]Object, len(dk.Dict.KeyOrder))
	for i, keyStr := range dk.Dict.KeyOrder {
		keys[i] = dk.Dict.Keys[keyStr]
	}
	return keys
}

func (dk *DictKeys) GetItem(index int64) (Object, bool) {
	length := dk.Len()
	if index < 0 {
		index = length + index
	}
	if index < 0 || index >= length {
		return nil, false
	}
	keyStr := dk.Dict.KeyOrder[index]
	return dk.Dict.Keys[keyStr], true
}

// DictValues is a view object for dict.values()
type DictValues struct {
	Dict *Dict
}

func NewDictValues(d *Dict) *DictValues {
	return &DictValues{Dict: d}
}

func (dv *DictValues) Type() ObjectType { return DICT_VALUES_OBJ }
func (dv *DictValues) Inspect() string {
	elements := make([]string, len(dv.Dict.KeyOrder))
	for i, keyStr := range dv.Dict.KeyOrder {
		val := dv.Dict.Pairs[keyStr]
		elements[i] = val.Inspect()
	}
	return fmt.Sprintf("dict_values([%s])", strings.Join(elements, ", "))
}

func (dv *DictValues) Len() int64 {
	return int64(len(dv.Dict.KeyOrder))
}

func (dv *DictValues) ToList() []Object {
	values := make([]Object, len(dv.Dict.KeyOrder))
	for i, keyStr := range dv.Dict.KeyOrder {
		values[i] = dv.Dict.Pairs[keyStr]
	}
	return values
}

func (dv *DictValues) GetItem(index int64) (Object, bool) {
	length := dv.Len()
	if index < 0 {
		index = length + index
	}
	if index < 0 || index >= length {
		return nil, false
	}
	keyStr := dv.Dict.KeyOrder[index]
	return dv.Dict.Pairs[keyStr], true
}

// DictItems is a view object for dict.items()
type DictItems struct {
	Dict *Dict
}

func NewDictItems(d *Dict) *DictItems {
	return &DictItems{Dict: d}
}

func (di *DictItems) Type() ObjectType { return DICT_ITEMS_OBJ }
func (di *DictItems) Inspect() string {
	elements := make([]string, len(di.Dict.KeyOrder))
	for i, keyStr := range di.Dict.KeyOrder {
		key := di.Dict.Keys[keyStr]
		val := di.Dict.Pairs[keyStr]
		elements[i] = fmt.Sprintf("(%s, %s)", key.Inspect(), val.Inspect())
	}
	return fmt.Sprintf("dict_items([%s])", strings.Join(elements, ", "))
}

func (di *DictItems) Len() int64 {
	return int64(len(di.Dict.KeyOrder))
}

func (di *DictItems) ToList() []Object {
	items := make([]Object, len(di.Dict.KeyOrder))
	for i, keyStr := range di.Dict.KeyOrder {
		key := di.Dict.Keys[keyStr]
		val := di.Dict.Pairs[keyStr]
		items[i] = &Tuple{Elements: []Object{key, val}}
	}
	return items
}

func (di *DictItems) GetItem(index int64) (Object, bool) {
	length := di.Len()
	if index < 0 {
		index = length + index
	}
	if index < 0 || index >= length {
		return nil, false
	}
	keyStr := di.Dict.KeyOrder[index]
	key := di.Dict.Keys[keyStr]
	val := di.Dict.Pairs[keyStr]
	return &Tuple{Elements: []Object{key, val}}, true
}

type Error struct {
	ErrorType string
	Message   string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string { return e.ErrorType + ": " + e.Message }
func (e *Error) Error() string   { return e.ErrorType + ": " + e.Message }

type ExceptionGroup struct {
	Message    string
	Exceptions []Object
}

func (eg *ExceptionGroup) Type() ObjectType { return EXCEPTION_GROUP_OBJ }
func (eg *ExceptionGroup) Inspect() string {
	var parts []string
	for _, exc := range eg.Exceptions {
		parts = append(parts, exc.Inspect())
	}
	return fmt.Sprintf("ExceptionGroup(%q, [%s])", eg.Message, strings.Join(parts, ", "))
}

type BuiltinFunction func(args ...Object) Object

// StringBuilder 用于高效字符串拼接优化
type StringBuilder struct {
	Builder *strings.Builder
}

func NewStringBuilder() *StringBuilder {
	return &StringBuilder{Builder: &strings.Builder{}}
}

func (sb *StringBuilder) Type() ObjectType { return STRING_BUILDER_OBJ }
func (sb *StringBuilder) Inspect() string  { return "<string_builder>" }

func (sb *StringBuilder) GetAttr(name string) (Object, bool) {
	switch name {
	case "append":
		return &Builtin{
			Name: "string_builder.append",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("append() takes at least 1 argument")
				}
				for _, arg := range args {
					var s string
					if str, ok := arg.(*String); ok {
						s = str.Value
					} else {
						s = arg.Inspect()
					}
					sb.Builder.WriteString(s)
				}
				return None_
			},
		}, true
	case "build":
		return &Builtin{
			Name: "string_builder.build",
			Fn: func(args ...Object) Object {
				return &String{Value: sb.Builder.String()}
			},
		}, true
	case "__len__":
		return &Builtin{
			Name: "string_builder.__len__",
			Fn: func(args ...Object) Object {
				return &Integer{Value: int64(sb.Builder.Len())}
			},
		}, true
	}
	return nil, false
}

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
	Instructions          []byte
	RegInstructions       interface{} // []compiler.RegInstruction, stored as interface{} to avoid import cycle
	NumLocals             int
	NumParameters         int
	NumKeywordOnly        int
	NumPositionalOnly     int
	NumDefaults           int
	NumPositionalDefaults int
	ParameterNames        []string
	PositionalOnly        []bool // parallel to ParameterNames, true if positional-only
	IsGenerator           bool
	Free                  []Object
	VarArgs               bool
	KwArgs                bool
}

func (c *Closure) Type() ObjectType { return FUNCTION_OBJ }
func (c *Closure) Inspect() string  { return "closure" }

type Class struct {
	Name         string
	Methods      map[string]Object
	Fields       map[string]Object
	SuperClass   *Class
	SuperClasses []*Class
	MRO          []*Class
	Slots        []string
	Metaclass    *Class
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string  { return fmt.Sprintf("<class %s>", c.Name) }

func (c *Class) HasSlots() bool {
	return len(c.Slots) > 0
}

func (c *Class) IsSlotAllowed(name string) bool {
	if !c.HasSlots() {
		return true
	}
	for _, s := range c.Slots {
		if s == name {
			return true
		}
	}
	if c.SuperClass != nil {
		return c.SuperClass.IsSlotAllowed(name)
	}
	for _, sc := range c.SuperClasses {
		if sc.IsSlotAllowed(name) {
			return true
		}
	}
	return false
}

func (c *Class) FindClassAttr(name string) (Object, bool) {
	if val, ok := c.Methods[name]; ok {
		return val, true
	}
	if val, ok := c.Fields[name]; ok {
		return val, true
	}
	if len(c.MRO) > 0 {
		for _, cls := range c.MRO {
			if val, ok := cls.Methods[name]; ok {
				return val, true
			}
			if val, ok := cls.Fields[name]; ok {
				return val, true
			}
		}
	} else if c.SuperClass != nil {
		if val, ok := c.SuperClass.Methods[name]; ok {
			return val, true
		}
		if val, ok := c.SuperClass.Fields[name]; ok {
			return val, true
		}
	}
	if len(c.SuperClasses) > 0 {
		for _, sc := range c.SuperClasses {
			if val, ok := sc.Methods[name]; ok {
				return val, true
			}
			if val, ok := sc.Fields[name]; ok {
				return val, true
			}
		}
	}
	return nil, false
}

type Instance struct {
	Class      *Class
	Fields     map[string]Object // 非 __slots__ 实例使用
	SlotValues []Object          // __slots__ 实例使用，按 AllSlotIndices 索引
}

// GetSlotIndex 返回属性名在类的所有 slots 中的索引，-1 表示不在 slots 中
func (i *Instance) GetSlotIndex(name string) int {
	if i.Class == nil {
		return -1
	}
	for idx, slotName := range i.Class.AllSlotNames() {
		if slotName == name {
			return idx
		}
	}
	return -1
}

// AllSlotNames 返回类及其所有父类的 slot 名称（按 MRO 顺序）
func (c *Class) AllSlotNames() []string {
	if len(c.Slots) == 0 && c.SuperClass == nil && len(c.SuperClasses) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var names []string
	// 按 MRO 顺序收集
	classes := c.MRO
	if len(classes) == 0 {
		classes = []*Class{c}
		if c.SuperClass != nil {
			classes = append([]*Class{c.SuperClass}, classes...)
		}
	}
	for _, cls := range classes {
		for _, s := range cls.Slots {
			if !seen[s] {
				seen[s] = true
				names = append(names, s)
			}
		}
	}
	return names
}

// IsSlottedInstance 返回实例是否使用 __slots__ 优化
func (i *Instance) IsSlottedInstance() bool {
	return len(i.SlotValues) > 0
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string  { return fmt.Sprintf("<%s instance>", i.Class.Name) }

func (i *Instance) GetAttr(name string) (Object, bool) {
	// 优先从 SlotValues 中查找
	if len(i.SlotValues) > 0 {
		if idx := i.GetSlotIndex(name); idx >= 0 && idx < len(i.SlotValues) {
			if i.SlotValues[idx] != nil {
				return i.SlotValues[idx], true
			}
			return nil, false
		}
	}
	if val, ok := i.Fields[name]; ok {
		return val, true
	}
	if i.Class != nil {
		if method, ok := i.Class.Methods[name]; ok {
			return method, true
		}
		if len(i.Class.MRO) > 0 {
			for _, cls := range i.Class.MRO {
				if method, ok := cls.Methods[name]; ok {
					return method, true
				}
			}
		} else if i.Class.SuperClass != nil {
			if method, ok := i.Class.SuperClass.Methods[name]; ok {
				return method, true
			}
		}
		if len(i.Class.SuperClasses) > 0 {
			for _, sc := range i.Class.SuperClasses {
				if method, ok := sc.Methods[name]; ok {
					return method, true
				}
			}
		}
	}
	return nil, false
}

func (i *Instance) SetAttr(name string, value Object) {
	if len(i.SlotValues) > 0 {
		if idx := i.GetSlotIndex(name); idx >= 0 {
			i.SlotValues[idx] = value
			return
		}
	}
	if i.Fields == nil {
		i.Fields = make(map[string]Object)
	}
	i.Fields[name] = value
}

type Descriptor interface {
	Object
	DescGet(obj Object, classObj *Class) (Object, error)
	DescSet(obj Object, value Object) error
	IsDataDesc() bool
}

func IsDescriptor(obj Object) bool {
	if _, ok := obj.(Descriptor); ok {
		return true
	}
	inst, ok := obj.(*Instance)
	if !ok {
		return false
	}
	_, hasGet := inst.GetAttr("__get__")
	return hasGet
}

func IsDataDescriptor(obj Object) bool {
	if desc, ok := obj.(Descriptor); ok {
		return desc.IsDataDesc()
	}
	inst, ok := obj.(*Instance)
	if !ok {
		return false
	}
	_, hasSet := inst.GetAttr("__set__")
	return hasSet
}

type Property struct {
	Fget Object
	Fset Object
	Fdel Object
}

func (p *Property) Type() ObjectType { return PROPERTY_OBJ }
func (p *Property) Inspect() string  { return "<property object>" }
func (p *Property) DescGet(obj Object, classObj *Class) (Object, error) {
	return nil, nil
}
func (p *Property) DescSet(obj Object, value Object) error {
	return nil
}
func (p *Property) IsDataDesc() bool {
	return p.Fset != nil && p.Fset != None_
}

type ClassMethod struct {
	Fn Object
}

func (cm *ClassMethod) Type() ObjectType { return CLASSMETHOD_OBJ }
func (cm *ClassMethod) Inspect() string  { return "<classmethod object>" }
func (cm *ClassMethod) DescGet(obj Object, classObj *Class) (Object, error) {
	return nil, nil
}
func (cm *ClassMethod) DescSet(obj Object, value Object) error {
	return nil
}

type Super struct {
	Instance  *Instance
	SuperClass *Class
}

func (s *Super) Type() ObjectType { return SUPER_OBJ }
func (s *Super) Inspect() string  { return "<super object>" }
func (cm *ClassMethod) IsDataDesc() bool { return false }

type StaticMethod struct {
	Fn Object
}

func (sm *StaticMethod) Type() ObjectType { return STATICMETHOD_OBJ }
func (sm *StaticMethod) Inspect() string  { return "<staticmethod object>" }
func (sm *StaticMethod) DescGet(obj Object, classObj *Class) (Object, error) {
	return nil, nil
}
func (sm *StaticMethod) DescSet(obj Object, value Object) error {
	return nil
}
func (sm *StaticMethod) IsDataDesc() bool { return false }

type BoundMethod struct {
	Self   *Instance
	Method Object
}

func (bm *BoundMethod) Type() ObjectType { return FUNCTION_OBJ }
func (bm *BoundMethod) Inspect() string  { return "<bound method>" }

// IsInstanceOf checks if obj is an instance of the given class
func IsInstanceOf(obj Object, class *Class) bool {
	switch o := obj.(type) {
	case *Instance:
		if o.Class == class {
			return true
		}
		// Check class hierarchy
		return isSubclassOf(o.Class, class)
	case *Integer:
		return class.Name == "int" || class.Name == "object"
	case *Float:
		return class.Name == "float" || class.Name == "object"
	case *String:
		return class.Name == "str" || class.Name == "object"
	case *Boolean:
		return class.Name == "bool" || class.Name == "int" || class.Name == "object"
	case *List:
		return class.Name == "list" || class.Name == "object"
	case *Dict:
		return class.Name == "dict" || class.Name == "object"
	case *Tuple:
		return class.Name == "tuple" || class.Name == "object"
	case *Set:
		return class.Name == "set" || class.Name == "object"
	case *None:
		return class.Name == "NoneType" || class.Name == "object"
	case *Builtin:
		return class.Name == "function" || class.Name == "object"
	case *Closure:
		return class.Name == "function" || class.Name == "object"
	case *Class:
		return class.Name == "type" || class.Name == "object"
	}
	return class.Name == "object"
}

// IsSubclassOf checks if cls is a subclass of parent
func IsSubclassOf(cls, parent *Class) bool {
	return isSubclassOf(cls, parent)
}

func isSubclassOf(cls, parent *Class) bool {
	if cls == parent {
		return true
	}
	for _, base := range cls.SuperClasses {
		if isSubclassOf(base, parent) {
			return true
		}
	}
	if cls.SuperClass != nil {
		return isSubclassOf(cls.SuperClass, parent)
	}
	return false
}

func (c *Class) ComputeMRO() []*Class {
	if len(c.SuperClasses) == 0 {
		if c.SuperClass != nil {
			c.MRO = []*Class{c.SuperClass}
		}
		return c.MRO
	}

	result := c3Linearize(c)
	c.MRO = result
	return result
}

func c3Linearize(cls *Class) []*Class {
	result := []*Class{cls}

	if len(cls.SuperClasses) == 0 {
		if cls.SuperClass != nil {
			result = append(result, cls.SuperClass)
		}
		return result
	}

	var linearizations [][]*Class
	for _, parent := range cls.SuperClasses {
		parentMRO := parent.ComputeMRO()
		parentL := make([]*Class, 0, len(parentMRO)+1)
		parentL = append(parentL, parent)
		parentL = append(parentL, parentMRO...)
		linearizations = append(linearizations, parentL)
	}

	parentsList := make([]*Class, len(cls.SuperClasses))
	copy(parentsList, cls.SuperClasses)
	linearizations = append(linearizations, parentsList)

	for {
		allEmpty := true
		for _, l := range linearizations {
			if len(l) > 0 {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			break
		}

		var candidate *Class
		found := false
		for _, l := range linearizations {
			if len(l) == 0 {
				continue
			}
			c := l[0]
			inTail := false
			for _, l2 := range linearizations {
				for j := 1; j < len(l2); j++ {
					if l2[j] == c {
						inTail = true
						break
					}
				}
				if inTail {
					break
				}
			}
			if !inTail {
				candidate = c
				found = true
				break
			}
		}

		if !found {
			break
		}

		result = append(result, candidate)

		for i, l := range linearizations {
			if len(l) > 0 && l[0] == candidate {
				linearizations[i] = l[1:]
			}
		}
	}

	return result
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

type Range struct {
	Start int64
	Stop  int64
	Step  int64
}

func NewRange(start, stop, step int64) *Range {
	return &Range{Start: start, Stop: stop, Step: step}
}

func (r *Range) Type() ObjectType { return RANGE_OBJ }
func (r *Range) Inspect() string {
	if r.Step == 1 {
		if r.Start == 0 {
			return fmt.Sprintf("range(%d)", r.Stop)
		}
		return fmt.Sprintf("range(%d, %d)", r.Start, r.Stop)
	}
	return fmt.Sprintf("range(%d, %d, %d)", r.Start, r.Stop, r.Step)
}

func (r *Range) Len() int64 {
	if r.Step > 0 {
		if r.Stop <= r.Start {
			return 0
		}
		return (r.Stop - r.Start + r.Step - 1) / r.Step
	}
	if r.Step < 0 {
		if r.Stop >= r.Start {
			return 0
		}
		return (r.Start - r.Stop - r.Step - 1) / (-r.Step)
	}
	return 0
}

func (r *Range) ToList() []Object {
	result := make([]Object, 0, r.Len())
	if r.Step > 0 {
		for i := r.Start; i < r.Stop; i += r.Step {
			result = append(result, &Integer{Value: i})
		}
	} else {
		for i := r.Start; i > r.Stop; i += r.Step {
			result = append(result, &Integer{Value: i})
		}
	}
	return result
}

func (r *Range) GetItem(index int64) (Object, bool) {
	length := r.Len()
	if index < 0 {
		index = length + index
	}
	if index < 0 || index >= length {
		return nil, false
	}
	return &Integer{Value: r.Start + index*r.Step}, true
}

type Zip struct {
	Iterables []Object
}

func NewZip(iterables []Object) *Zip {
	return &Zip{Iterables: iterables}
}

func (z *Zip) Type() ObjectType { return ZIP_OBJ }
func (z *Zip) Inspect() string  { return "<zip object>" }

func (z *Zip) Len() int64 {
	minLen := int64(-1)
	for _, iter := range z.Iterables {
		switch v := iter.(type) {
		case *List:
			if minLen == -1 || int64(len(v.Elements)) < minLen {
				minLen = int64(len(v.Elements))
			}
		case *Range:
			l := v.Len()
			if minLen == -1 || l < minLen {
				minLen = l
			}
		case *String:
			if minLen == -1 || int64(len(v.Value)) < minLen {
				minLen = int64(len(v.Value))
			}
		case *Tuple:
			if minLen == -1 || int64(len(v.Elements)) < minLen {
				minLen = int64(len(v.Elements))
			}
		default:
			return -1
		}
	}
	if minLen == -1 {
		return 0
	}
	return minLen
}

func (z *Zip) ToList() []Object {
	length := z.Len()
	if length < 0 {
		return nil
	}
	result := make([]Object, 0, length)
	for i := int64(0); i < length; i++ {
		tuple := make([]Object, len(z.Iterables))
		for j, iter := range z.Iterables {
			switch v := iter.(type) {
			case *List:
				tuple[j] = v.Elements[i]
			case *Range:
				tuple[j], _ = v.GetItem(i)
			case *String:
				tuple[j] = &String{Value: string(v.Value[i])}
			case *Tuple:
				tuple[j] = v.Elements[i]
			}
		}
		result = append(result, &Tuple{Elements: tuple})
	}
	return result
}

// RegexPattern represents a compiled regular expression pattern
type RegexPattern struct {
	Regexp  *regexp.Regexp
	Pattern string
	Flags   int64
}

func (p *RegexPattern) Type() ObjectType { return REGEX_PATTERN_OBJ }
func (p *RegexPattern) Inspect() string  { return fmt.Sprintf("re.compile(%q)", p.Pattern) }

func (p *RegexPattern) GetAttr(name string) (Object, bool) {
	switch name {
	case "pattern":
		return &String{Value: p.Pattern}, true
	case "flags":
		return &Integer{Value: p.Flags}, true
	case "search":
		return &Builtin{
			Name: "Pattern.search",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("search() takes at least 1 argument")
				}
				s, ok := args[0].(*String)
				if !ok {
					return NewTypeError("search() argument must be a string")
				}
				return PatternSearch(p, s.Value)
			},
		}, true
	case "match":
		return &Builtin{
			Name: "Pattern.match",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("match() takes at least 1 argument")
				}
				s, ok := args[0].(*String)
				if !ok {
					return NewTypeError("match() argument must be a string")
				}
				return PatternMatch(p, s.Value)
			},
		}, true
	case "fullmatch":
		return &Builtin{
			Name: "Pattern.fullmatch",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("fullmatch() takes at least 1 argument")
				}
				s, ok := args[0].(*String)
				if !ok {
					return NewTypeError("fullmatch() argument must be a string")
				}
				return PatternFullmatch(p, s.Value)
			},
		}, true
	case "findall":
		return &Builtin{
			Name: "Pattern.findall",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("findall() takes at least 1 argument")
				}
				s, ok := args[0].(*String)
				if !ok {
					return NewTypeError("findall() argument must be a string")
				}
				return PatternFindall(p, s.Value)
			},
		}, true
	case "finditer":
		return &Builtin{
			Name: "Pattern.finditer",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("finditer() takes at least 1 argument")
				}
				s, ok := args[0].(*String)
				if !ok {
					return NewTypeError("finditer() argument must be a string")
				}
				return PatternFinditer(p, s.Value)
			},
		}, true
	case "sub":
		return &Builtin{
			Name: "Pattern.sub",
			Fn: func(args ...Object) Object {
				if len(args) < 2 {
					return NewTypeError("sub() takes at least 2 arguments")
				}
				s, ok2 := args[1].(*String)
				if !ok2 {
					return NewTypeError("sub() second argument must be a string")
				}
				var count int
				if len(args) >= 3 {
					if c, ok := args[2].(*Integer); ok {
						count = int(c.Value)
					}
				}

				repl := args[0]

				// If repl is callable, use callable replacement
				if IsCallable(repl) {
					return PatternCallableSub(p, repl, s.Value, count)
				}

				replStr, ok1 := repl.(*String)
				if !ok1 {
					return NewTypeError("sub() first argument must be a string or callable")
				}
				var result string
				if count > 0 {
					locs := p.Regexp.FindAllStringIndex(s.Value, count)
					if locs == nil {
						result = s.Value
					} else {
						var buf strings.Builder
						prev := 0
						for _, loc := range locs {
							buf.WriteString(s.Value[prev:loc[0]])
							buf.WriteString(p.Regexp.ReplaceAllString(s.Value[loc[0]:loc[1]], replStr.Value))
							prev = loc[1]
						}
						buf.WriteString(s.Value[prev:])
						result = buf.String()
					}
				} else {
					result = p.Regexp.ReplaceAllString(s.Value, replStr.Value)
				}
				return &String{Value: result}
			},
		}, true
	case "subn":
		return &Builtin{
			Name: "Pattern.subn",
			Fn: func(args ...Object) Object {
				if len(args) < 2 {
					return NewTypeError("subn() takes at least 2 arguments")
				}
				s, ok2 := args[1].(*String)
				if !ok2 {
					return NewTypeError("subn() second argument must be a string")
				}
				var count int
				if len(args) >= 3 {
					if c, ok := args[2].(*Integer); ok {
						count = int(c.Value)
					}
				}

				repl := args[0]

				// If repl is callable, use callable replacement
				if IsCallable(repl) {
					result := PatternCallableSub(p, repl, s.Value, count)
					if result.Type() == ERROR_OBJ {
						return result
					}
					resultStr := result.(*String).Value
					locs := p.Regexp.FindAllStringIndex(s.Value, count)
					n := 0
					if locs != nil {
						n = len(locs)
					}
					return &Tuple{Elements: []Object{&String{Value: resultStr}, &Integer{Value: int64(n)}}}
				}

				replStr, ok1 := repl.(*String)
				if !ok1 {
					return NewTypeError("subn() first argument must be a string or callable")
				}
				var result string
				var n int
				if count > 0 {
					locs := p.Regexp.FindAllStringIndex(s.Value, count)
					if locs == nil {
						result = s.Value
						n = 0
					} else {
						var buf strings.Builder
						prev := 0
						for _, loc := range locs {
							buf.WriteString(s.Value[prev:loc[0]])
							buf.WriteString(p.Regexp.ReplaceAllString(s.Value[loc[0]:loc[1]], replStr.Value))
							prev = loc[1]
						}
						buf.WriteString(s.Value[prev:])
						result = buf.String()
						n = len(locs)
					}
				} else {
					result = p.Regexp.ReplaceAllString(s.Value, replStr.Value)
					locs := p.Regexp.FindAllStringIndex(s.Value, -1)
					if locs != nil {
						n = len(locs)
					}
				}
				return &Tuple{Elements: []Object{&String{Value: result}, &Integer{Value: int64(n)}}}
			},
		}, true
	case "split":
		return &Builtin{
			Name: "Pattern.split",
			Fn: func(args ...Object) Object {
				if len(args) < 1 {
					return NewTypeError("split() takes at least 1 argument")
				}
				s, ok := args[0].(*String)
				if !ok {
					return NewTypeError("split() argument must be a string")
				}
				return PatternSplit(p, s.Value)
			},
		}, true
	}
	return nil, false
}

// RegexMatch represents a match result from a regular expression search
type RegexMatch struct {
	Groups         []string
	GroupStarts    []int // start positions for each group
	GroupEnds      []int // end positions for each group
	OriginalString string
	Pattern        *RegexPattern
}

func (m *RegexMatch) Type() ObjectType { return REGEX_MATCH_OBJ }
func (m *RegexMatch) Inspect() string  { return "<re.Match object>" }

func (m *RegexMatch) GetAttr(name string) (Object, bool) {
	switch name {
	case "string":
		return &String{Value: m.OriginalString}, true
	case "re":
		return m.Pattern, true
	case "lastindex":
		if len(m.Groups) > 1 {
			return &Integer{Value: int64(len(m.Groups) - 1)}, true
		}
		return None_, true
	case "group":
		return &Builtin{
			Name: "Match.group",
			Fn: func(args ...Object) Object {
				idx := int64(0)
				if len(args) > 0 {
					if i, ok := args[0].(*Integer); ok {
						idx = i.Value
					} else {
						return NewTypeError("group() argument must be an integer")
					}
				}
				if idx < 0 || int(idx) >= len(m.Groups) {
					return NewIndexError("no such group")
				}
				return &String{Value: m.Groups[idx]}
			},
		}, true
	case "start":
		return &Builtin{
			Name: "Match.start",
			Fn: func(args ...Object) Object {
				idx := int64(0)
				if len(args) > 0 {
					if i, ok := args[0].(*Integer); ok {
						idx = i.Value
					}
				}
				if idx < 0 || int(idx) >= len(m.GroupStarts) {
					return NewIndexError("no such group")
				}
				return &Integer{Value: int64(m.GroupStarts[idx])}
			},
		}, true
	case "end":
		return &Builtin{
			Name: "Match.end",
			Fn: func(args ...Object) Object {
				idx := int64(0)
				if len(args) > 0 {
					if i, ok := args[0].(*Integer); ok {
						idx = i.Value
					}
				}
				if idx < 0 || int(idx) >= len(m.GroupEnds) {
					return NewIndexError("no such group")
				}
				return &Integer{Value: int64(m.GroupEnds[idx])}
			},
		}, true
	case "span":
		return &Builtin{
			Name: "Match.span",
			Fn: func(args ...Object) Object {
				idx := int64(0)
				if len(args) > 0 {
					if i, ok := args[0].(*Integer); ok {
						idx = i.Value
					}
				}
				if idx < 0 || int(idx) >= len(m.GroupStarts) {
					return NewIndexError("no such group")
				}
				return &Tuple{Elements: []Object{
					&Integer{Value: int64(m.GroupStarts[idx])},
					&Integer{Value: int64(m.GroupEnds[idx])},
				}}
			},
		}, true
	case "groups":
		return &Builtin{
			Name: "Match.groups",
			Fn: func(args ...Object) Object {
				elements := make([]Object, 0, len(m.Groups)-1)
				for i := 1; i < len(m.Groups); i++ {
					elements = append(elements, &String{Value: m.Groups[i]})
				}
				return &Tuple{Elements: elements}
			},
		}, true
	}
	return nil, false
}

// Helper functions for Pattern methods that are also used by the re module

func PatternSearch(p *RegexPattern, s string) Object {
	loc := p.Regexp.FindStringSubmatchIndex(s)
	if loc == nil {
		return None_
	}
	return newMatchFromLoc(p, s, loc)
}

func PatternMatch(p *RegexPattern, s string) Object {
	loc := p.Regexp.FindStringSubmatchIndex(s)
	if loc == nil || loc[0] != 0 {
		return None_
	}
	return newMatchFromLoc(p, s, loc)
}

func PatternFullmatch(p *RegexPattern, s string) Object {
	loc := p.Regexp.FindStringSubmatchIndex(s)
	if loc == nil || loc[0] != 0 || loc[1] != len(s) {
		return None_
	}
	return newMatchFromLoc(p, s, loc)
}

func PatternFindall(p *RegexPattern, s string) Object {
	matches := p.Regexp.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return &List{Elements: []Object{}}
	}

	numGroups := p.Regexp.NumSubexp()
	result := make([]Object, 0, len(matches))

	if numGroups == 0 {
		// No groups: return list of matched strings
		for _, m := range matches {
			result = append(result, &String{Value: m[0]})
		}
	} else if numGroups == 1 {
		// One group: return list of group values
		for _, m := range matches {
			if len(m) > 1 && m[1] != "" {
				result = append(result, &String{Value: m[1]})
			} else {
				result = append(result, &String{Value: ""})
			}
		}
	} else {
		// Multiple groups: return list of tuples
		for _, m := range matches {
			elements := make([]Object, numGroups)
			for i := 0; i < numGroups; i++ {
				if i+1 < len(m) {
					elements[i] = &String{Value: m[i+1]}
				} else {
					elements[i] = &String{Value: ""}
				}
			}
			result = append(result, &Tuple{Elements: elements})
		}
	}
	return &List{Elements: result}
}

func PatternFinditer(p *RegexPattern, s string) Object {
	locs := p.Regexp.FindAllStringSubmatchIndex(s, -1)
	if locs == nil {
		return &List{Elements: []Object{}}
	}
	result := make([]Object, 0, len(locs))
	for _, loc := range locs {
		result = append(result, newMatchFromLoc(p, s, loc))
	}
	return &List{Elements: result}
}

func PatternSplit(p *RegexPattern, s string) Object {
	numGroups := p.Regexp.NumSubexp()
	if numGroups == 0 {
		// No capturing groups: simple split
		result := p.Regexp.Split(s, -1)
		elements := make([]Object, len(result))
		for i, part := range result {
			elements[i] = &String{Value: part}
		}
		return &List{Elements: elements}
	}

	// With capturing groups: include captured text in result
	locs := p.Regexp.FindAllStringSubmatchIndex(s, -1)
	if locs == nil {
		return &List{Elements: []Object{&String{Value: s}}}
	}

	result := make([]Object, 0)
	prev := 0
	for _, loc := range locs {
		// Add text before match
		result = append(result, &String{Value: s[prev:loc[0]]})
		// Add captured groups
		for i := 0; i < numGroups; i++ {
			start := loc[2*(i+1)]
			end := loc[2*(i+1)+1]
			if start >= 0 && end >= 0 {
				result = append(result, &String{Value: s[start:end]})
			} else {
				result = append(result, None_)
			}
		}
		prev = loc[1]
	}
	// Add remaining text
	result = append(result, &String{Value: s[prev:]})

	return &List{Elements: result}
}

// PatternCallableSub performs regex substitution where repl is a callable.
// For each match, the callable is invoked with a RegexMatch object and
// its return value is used as the replacement string.
func PatternCallableSub(p *RegexPattern, repl Object, s string, count int) Object {
	var locs [][]int
	if count > 0 {
		locs = p.Regexp.FindAllStringSubmatchIndex(s, count)
	} else {
		locs = p.Regexp.FindAllStringSubmatchIndex(s, -1)
	}

	if locs == nil {
		return &String{Value: s}
	}

	var buf strings.Builder
	prev := 0
	for _, loc := range locs {
		buf.WriteString(s[prev:loc[0]])

		// Create a RegexMatch object for this match
		match := newMatchFromLoc(p, s, loc)

		// Call the callable with the match object
		result := CallFunction(repl, match)
		if result.Type() == ERROR_OBJ {
			return result
		}

		// Convert the result to a string
		var replacement string
		if strObj, ok := result.(*String); ok {
			replacement = strObj.Value
		} else {
			replacement = result.Inspect()
		}

		buf.WriteString(replacement)
		prev = loc[1]
	}
	buf.WriteString(s[prev:])

	return &String{Value: buf.String()}
}

func newMatchFromLoc(p *RegexPattern, s string, loc []int) *RegexMatch {
	groups := make([]string, len(loc)/2)
	groupStarts := make([]int, len(loc)/2)
	groupEnds := make([]int, len(loc)/2)
	for i := 0; i < len(loc)/2; i++ {
		start := loc[i*2]
		end := loc[i*2+1]
		groupStarts[i] = start
		groupEnds[i] = end
		if start >= 0 && end >= 0 {
			groups[i] = s[start:end]
		}
	}
	return &RegexMatch{
		Groups:         groups,
		GroupStarts:    groupStarts,
		GroupEnds:      groupEnds,
		OriginalString: s,
		Pattern:        p,
	}
}

var modules = make(map[string]*Module)

// callFunctionFn is a callback that allows calling Python callable objects from Go code.
// The VM registers this at startup. This is needed for features like re.sub(pattern, callable, string).
var callFunctionFn func(callee Object, args ...Object) Object

// SetCallFunctionCallback registers the callback for calling Python callable objects.
func SetCallFunctionCallback(fn func(callee Object, args ...Object) Object) {
	callFunctionFn = fn
}

// CallFunction calls a Python callable object with the given arguments.
// Returns the result or an Error if the callable cannot be invoked.
func CallFunction(callee Object, args ...Object) Object {
	if callFunctionFn == nil {
		return NewTypeError("cannot call Python function in this context")
	}
	return callFunctionFn(callee, args...)
}

// IsCallable checks if an object can be called as a Python function.
func IsCallable(obj Object) bool {
	// First check if it implements Callable interface
	if _, ok := obj.(Callable); ok {
		return true
	}
	switch obj.(type) {
	case *Builtin:
		return true
	case *Closure:
		return true
	case *Class:
		return true
	case *Instance:
		// Check if instance has __call__
		if inst, ok := obj.(*Instance); ok {
			if _, hasCall := inst.Class.Methods["__call__"]; hasCall {
				return true
			}
			if _, hasCall := inst.Fields["__call__"]; hasCall {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func RegisterModule(name string, module *Module) {
	modules[name] = module
}

func GetModule(name string) *Module {
	return modules[name]
}

var (
	True            = &Boolean{Value: true}
	False           = &Boolean{Value: false}
	None_           = &None{}
	EllipsisSingleton = &Ellipsis{}
)

// CheckHashable returns an error if the object type cannot be used as a dict/set key
func CheckHashable(obj Object) error {
	switch o := obj.(type) {
	case *Integer, *Float, *Boolean, *String:
		return nil
	case *Tuple:
		for _, elem := range o.Elements {
			if err := CheckHashable(elem); err != nil {
				return err
			}
		}
		return nil
	default:
		return NewTypeError("unhashable type: '%s'", obj.Type())
	}
}

func NewErrorWithType(errorType, format string, a ...interface{}) *Error {
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
	return NewErrorWithType("Exception", format, a...)
}

func NewValueError(format string, a ...interface{}) *Error {
	return NewErrorWithType("ValueError", format, a...)
}

func NewTypeError(format string, a ...interface{}) *Error {
	return NewErrorWithType("TypeError", format, a...)
}

func NewZeroDivisionError(format string, a ...interface{}) *Error {
	return NewErrorWithType("ZeroDivisionError", format, a...)
}

func NewIndexError(format string, a ...interface{}) *Error {
	return NewErrorWithType("IndexError", format, a...)
}

func NewKeyError(format string, a ...interface{}) *Error {
	return NewErrorWithType("KeyError", format, a...)
}

func NewAttributeError(format string, a ...interface{}) *Error {
	return NewErrorWithType("AttributeError", format, a...)
}

type EnumMember struct {
	Name  string
	Value Object
	Enum  *Enum
}

func (em *EnumMember) Type() ObjectType { return ENUM_MEMBER_OBJ }
func (em *EnumMember) Inspect() string  { return fmt.Sprintf("<%s.%s: %s>", em.Enum.Name, em.Name, em.Value.Inspect()) }

type Enum struct {
	Name    string
	Members map[string]*EnumMember
}

func NewEnum(name string, members map[string]Object) *Enum {
	e := &Enum{
		Name:    name,
		Members: make(map[string]*EnumMember),
	}
	for k, v := range members {
		e.Members[k] = &EnumMember{Name: k, Value: v, Enum: e}
	}
	return e
}

func (e *Enum) Type() ObjectType { return ENUM_OBJ }
func (e *Enum) Inspect() string  { return fmt.Sprintf("<enum %s>", e.Name) }

func (e *Enum) GetAttr(name string) (Object, bool) {
	if member, ok := e.Members[name]; ok {
		return member, true
	}
	return nil, false
}

func NewNameError(format string, a ...interface{}) *Error {
	return NewErrorWithType("NameError", format, a...)
}

func NewAssertionError(format string, a ...interface{}) *Error {
	return NewErrorWithType("AssertionError", format, a...)
}

func NewRuntimeError(format string, a ...interface{}) *Error {
	return NewErrorWithType("RuntimeError", format, a...)
}

func NewNotImplementedError(format string, a ...interface{}) *Error {
	return NewErrorWithType("NotImplementedError", format, a...)
}

func NewStopIteration(format string, a ...interface{}) *Error {
	return NewErrorWithType("StopIteration", format, a...)
}

func NewOverflowError(format string, a ...interface{}) *Error {
	return NewErrorWithType("OverflowError", format, a...)
}

func NewFileNotFoundError(format string, a ...interface{}) *Error {
	return NewErrorWithType("FileNotFoundError", format, a...)
}

func NewImportError(format string, a ...interface{}) *Error {
	return NewErrorWithType("ImportError", format, a...)
}

func NewSyntaxError(format string, a ...interface{}) *Error {
	return NewErrorWithType("SyntaxError", format, a...)
}

func NewIndentationError(format string, a ...interface{}) *Error {
	return NewErrorWithType("IndentationError", format, a...)
}

func NewUnboundLocalError(format string, a ...interface{}) *Error {
	return NewErrorWithType("UnboundLocalError", format, a...)
}

func NewRecursionError(format string, a ...interface{}) *Error {
	return NewErrorWithType("RecursionError", format, a...)
}

func NewMemoryError(format string, a ...interface{}) *Error {
	return NewErrorWithType("MemoryError", format, a...)
}

func NewOSError(format string, a ...interface{}) *Error {
	return NewErrorWithType("OSError", format, a...)
}

func NewIOError(format string, a ...interface{}) *Error {
	return NewErrorWithType("IOError", format, a...)
}

func NewFileExistsError(format string, a ...interface{}) *Error {
	return NewErrorWithType("FileExistsError", format, a...)
}

func NewPermissionError(format string, a ...interface{}) *Error {
	return NewErrorWithType("PermissionError", format, a...)
}

func NewModuleNotFoundError(format string, a ...interface{}) *Error {
	return NewErrorWithType("ModuleNotFoundError", format, a...)
}

func NewUnicodeError(format string, a ...interface{}) *Error {
	return NewErrorWithType("UnicodeError", format, a...)
}

func NewBufferError(format string, a ...interface{}) *Error {
	return NewErrorWithType("BufferError", format, a...)
}

func NewArithmeticError(format string, a ...interface{}) *Error {
	return NewErrorWithType("ArithmeticError", format, a...)
}

func NewLookupError(format string, a ...interface{}) *Error {
	return NewErrorWithType("LookupError", format, a...)
}

func NewReferenceError(format string, a ...interface{}) *Error {
	return NewErrorWithType("ReferenceError", format, a...)
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
	case *Complex:
		b := b.(*Complex)
		return a.Real == b.Real && a.Imag == b.Imag
	case *String:
		b := b.(*String)
		return a.Value == b.Value
	case *Boolean:
		b := b.(*Boolean)
		return a.Value == b.Value
	case *Bytes:
		b := b.(*Bytes)
		if len(a.Value) != len(b.Value) {
			return false
		}
		for i := range a.Value {
			if a.Value[i] != b.Value[i] {
				return false
			}
		}
		return true
	case *None:
		return true
	case *Ellipsis:
		return true
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

func toFloat(obj Object) (float64, bool) {
	switch v := obj.(type) {
	case *Float:
		return v.Value, true
	case *Integer:
		return float64(v.Value), true
	default:
		return 0, false
	}
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
			case *Complex:
				magnitude := cmplx.Abs(complex(v.Real, v.Imag))
				return &Float{Value: magnitude}
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

	// math.atan2
	mathModule.Fields["atan2"] = &Builtin{Name: "math.atan2", Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return NewTypeError("atan2() takes exactly 2 arguments")
		}
		y, ok1 := toFloat(args[0]); x, ok2 := toFloat(args[1])
		if !ok1 || !ok2 {
			return NewTypeError("atan2() arguments must be numbers")
		}
		return &Float{Value: math.Atan2(y, x)}
	}}

	// math.copysign
	mathModule.Fields["copysign"] = &Builtin{Name: "math.copysign", Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return NewTypeError("copysign() takes exactly 2 arguments")
		}
		x, ok1 := toFloat(args[0]); y, ok2 := toFloat(args[1])
		if !ok1 || !ok2 {
			return NewTypeError("copysign() arguments must be numbers")
		}
		return &Float{Value: math.Copysign(x, y)}
	}}

	// math.fmod
	mathModule.Fields["fmod"] = &Builtin{Name: "math.fmod", Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return NewTypeError("fmod() takes exactly 2 arguments")
		}
		x, ok1 := toFloat(args[0]); y, ok2 := toFloat(args[1])
		if !ok1 || !ok2 {
			return NewTypeError("fmod() arguments must be numbers")
		}
		return &Float{Value: math.Mod(x, y)}
	}}

	// math.isnan
	mathModule.Fields["isnan"] = &Builtin{Name: "math.isnan", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("isnan() takes exactly 1 argument")
		}
		x, ok := toFloat(args[0]); if !ok {
			return NewTypeError("isnan() argument must be a number")
		}
		return &Boolean{Value: math.IsNaN(x)}
	}}

	// math.isinf
	mathModule.Fields["isinf"] = &Builtin{Name: "math.isinf", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("isinf() takes exactly 1 argument")
		}
		x, ok := toFloat(args[0]); if !ok {
			return NewTypeError("isinf() argument must be a number")
		}
		return &Boolean{Value: math.IsInf(x, 0)}
	}}

	// math.isfinite
	mathModule.Fields["isfinite"] = &Builtin{Name: "math.isfinite", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("isfinite() takes exactly 1 argument")
		}
		x, ok := toFloat(args[0]); if !ok {
			return NewTypeError("isfinite() argument must be a number")
		}
		return &Boolean{Value: !math.IsInf(x, 0) && !math.IsNaN(x)}
	}}

	// math.factorial
	mathModule.Fields["factorial"] = &Builtin{Name: "math.factorial", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("factorial() takes exactly 1 argument")
		}
		n, ok := args[0].(*Integer); if !ok {
			return NewTypeError("factorial() argument must be an integer")
		}
		if n.Value < 0 {
			return NewValueError("factorial() not defined for negative values")
		}
		result := int64(1)
		for i := int64(2); i <= n.Value; i++ {
			result *= i
		}
		return &Integer{Value: result}
	}}

	// math.gcd
	mathModule.Fields["gcd"] = &Builtin{Name: "math.gcd", Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return NewTypeError("gcd() takes 1 or 2 arguments")
		}
		a, ok1 := args[0].(*Integer)
		if !ok1 {
			return NewTypeError("gcd() arguments must be integers")
		}
		b := &Integer{Value: 0}
		if len(args) == 2 {
			b, ok1 = args[1].(*Integer); if !ok1 {
				return NewTypeError("gcd() arguments must be integers")
			}
		}
		av, bv := a.Value, b.Value
		if av < 0 {
			av = -av
		}
		if bv < 0 {
			bv = -bv
		}
		for bv != 0 {
			av, bv = bv, av%bv
		}
		return &Integer{Value: av}
	}}

	// math.lcm
	mathModule.Fields["lcm"] = &Builtin{Name: "math.lcm", Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return NewTypeError("lcm() takes 1 or 2 arguments")
		}
		a, ok1 := args[0].(*Integer)
		if !ok1 {
			return NewTypeError("lcm() arguments must be integers")
		}
		if a.Value == 0 {
			return &Integer{Value: 0}
		}
		b := &Integer{Value: 1}
		if len(args) == 2 {
			b, ok1 = args[1].(*Integer); if !ok1 {
				return NewTypeError("lcm() arguments must be integers")
			}
		}
		if b.Value == 0 {
			return &Integer{Value: 0}
		}
		av, bv := a.Value, b.Value
		if av < 0 {
			av = -av
		}
		if bv < 0 {
			bv = -bv
		}
		g := av; tmp := bv; for tmp != 0 {
			g, tmp = tmp, g%tmp
		}
		return &Integer{Value: av / g * bv}
	}}

	// Constants
	mathModule.Fields["inf"] = &Float{Value: math.Inf(1)}
	mathModule.Fields["nan"] = &Float{Value: math.NaN()}
	mathModule.Fields["tau"] = &Float{Value: math.Pi * 2}

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
		Value: "GoPython 0.11.0 (Go implementation)",
	}

	// sys.platform
	sysModule.Fields["platform"] = &String{
		Value: runtime.GOOS,
	}

	// sys.version_info
	sysModule.Fields["version_info"] = &Tuple{
		Elements: []Object{
			&Integer{Value: 0},
			&Integer{Value: 11},
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
			obj := args[0]
			switch o := obj.(type) {
			case *Integer:
				return &Integer{Value: 28}
			case *Float:
				return &Integer{Value: 24}
			case *Boolean:
				return &Integer{Value: 28}
			case *String:
				return &Integer{Value: int64(49 + len(o.Value))}
			case *List:
				return &Integer{Value: int64(56 + 8*len(o.Elements))}
			case *Dict:
				n := len(o.Pairs)
				return &Integer{Value: int64(64 + 8*n + 8*n)}
			case *Tuple:
				return &Integer{Value: int64(40 + 8*len(o.Elements))}
			case *Set:
				n := len(o.Elements)
				return &Integer{Value: int64(216 + 8*n)}
			case *Bytes:
				return &Integer{Value: int64(33 + len(o.Value))}
			case *None:
				return &Integer{Value: 16}
			case *Complex:
				return &Integer{Value: 32}
			case *Closure:
				return &Integer{Value: 136}
			case *Builtin:
				return &Integer{Value: 136}
			case *BoundMethod:
				return &Integer{Value: 136}
			case *Class:
				return &Integer{Value: 64}
			case *Instance:
				size := int64(56)
				if o.Fields != nil {
					size += int64(8 * len(o.Fields))
				}
				if o.SlotValues != nil {
					size += int64(8 * len(o.SlotValues))
				}
				return &Integer{Value: size}
			default:
				return &Integer{Value: 24}
			}
		},
	}

	// sys.maxsize
	sysModule.Fields["maxsize"] = &Integer{Value: int64(1<<63 - 1)}

	// sys.byteorder
	sysModule.Fields["byteorder"] = &String{Value: "little"}

	// sys.executable
	sysModule.Fields["executable"] = &String{Value: "gopy"}

	// sys.prefix
	sysModule.Fields["prefix"] = &String{Value: "/usr/local"}

	// sys.modules
	sysModule.Fields["modules"] = &Dict{Pairs: make(map[string]Object), Keys: make(map[string]Object)}

	// sys.flags
	flagsModule := &Module{Name: "sys.flags", Fields: make(map[string]Object)}
	flagsModule.Fields["debug"] = &Integer{Value: 0}
	flagsModule.Fields["optimize"] = &Integer{Value: 0}
	sysModule.Fields["flags"] = flagsModule

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
				return NewError("%s", err.Error())
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
				return NewError("%s", err.Error())
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
				return NewError("%s", err.Error())
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
				return NewError("%s", err.Error())
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
				return NewError("%s", err.Error())
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
				return NewError("%s", err.Error())
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

	// os.rmdir
	osModule.Fields["rmdir"] = &Builtin{Name: "os.rmdir", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("rmdir() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("rmdir() argument must be a string")
		}
		if err := os.Remove(path.Value); err != nil {
			return NewError("%s", err.Error())
		}
		return None_
	}}

	// os.makedirs
	osModule.Fields["makedirs"] = &Builtin{Name: "os.makedirs", Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return NewTypeError("makedirs() takes at least 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("makedirs() argument must be a string")
		}
		mode := os.FileMode(0755)
		if len(args) >= 2 {
			if i, ok := args[1].(*Integer); ok {
				mode = os.FileMode(i.Value)
			}
		}
		if err := os.MkdirAll(path.Value, mode); err != nil {
			return NewError("%s", err.Error())
		}
		return None_
	}}

	// os.removedirs
	osModule.Fields["removedirs"] = &Builtin{Name: "os.removedirs", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("removedirs() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("removedirs() argument must be a string")
		}
		if err := os.RemoveAll(path.Value); err != nil {
			return NewError("%s", err.Error())
		}
		return None_
	}}

	// os.stat
	osModule.Fields["stat"] = &Builtin{Name: "os.stat", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("stat() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("stat() argument must be a string")
		}
		info, err := os.Stat(path.Value)
		if err != nil {
			return NewError("%s", err.Error())
		}
		d := NewDict()
		d.Set(&String{Value: "st_size"}, &Integer{Value: info.Size()})
		d.Set(&String{Value: "st_mode"}, &Integer{Value: int64(info.Mode())})
		d.Set(&String{Value: "st_isdir"}, &Boolean{Value: info.IsDir()})
		d.Set(&String{Value: "st_mtime"}, &Float{Value: float64(info.ModTime().UnixNano()) / 1e9})
		return d
	}}

	// os.path sub-module
	pathModule := &Module{Name: "os.path", Fields: make(map[string]Object)}
	pathModule.Fields["exists"] = &Builtin{Name: "os.path.exists", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("exists() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("exists() argument must be a string")
		}
		_, err := os.Stat(path.Value)
		return &Boolean{Value: err == nil}
	}}
	pathModule.Fields["isfile"] = &Builtin{Name: "os.path.isfile", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("isfile() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("isfile() argument must be a string")
		}
		info, err := os.Stat(path.Value)
		if err != nil {
			return False
		}
		return &Boolean{Value: !info.IsDir()}
	}}
	pathModule.Fields["isdir"] = &Builtin{Name: "os.path.isdir", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("isdir() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("isdir() argument must be a string")
		}
		info, err := os.Stat(path.Value)
		if err != nil {
			return False
		}
		return &Boolean{Value: info.IsDir()}
	}}
	pathModule.Fields["join"] = &Builtin{Name: "os.path.join", Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return NewTypeError("join() takes at least 1 argument")
		}
		parts := make([]string, len(args))
		for i, a := range args {
			s, ok := a.(*String); if !ok {
				return NewTypeError("join() arguments must be strings")
			}
			parts[i] = s.Value
		}
		return &String{Value: filepath.Join(parts...)}
	}}
	pathModule.Fields["basename"] = &Builtin{Name: "os.path.basename", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("basename() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("basename() argument must be a string")
		}
		return &String{Value: filepath.Base(path.Value)}
	}}
	pathModule.Fields["dirname"] = &Builtin{Name: "os.path.dirname", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("dirname() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("dirname() argument must be a string")
		}
		return &String{Value: filepath.Dir(path.Value)}
	}}
	pathModule.Fields["split"] = &Builtin{Name: "os.path.split", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("split() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("split() argument must be a string")
		}
		dir, file := filepath.Split(path.Value)
		return &Tuple{Elements: []Object{&String{Value: dir}, &String{Value: file}}}
	}}
	pathModule.Fields["getsize"] = &Builtin{Name: "os.path.getsize", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("getsize() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("getsize() argument must be a string")
		}
		info, err := os.Stat(path.Value)
		if err != nil {
			return NewError("%s", err.Error())
		}
		return &Integer{Value: info.Size()}
	}}
	pathModule.Fields["abspath"] = &Builtin{Name: "os.path.abspath", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("abspath() takes exactly 1 argument")
		}
		path, ok := args[0].(*String); if !ok {
			return NewTypeError("abspath() argument must be a string")
		}
		abs, err := filepath.Abs(path.Value)
		if err != nil {
			return NewError("%s", err.Error())
		}
		return &String{Value: abs}
	}}
	osModule.Fields["path"] = pathModule

	// os.name
	osModule.Fields["name"] = &String{Value: runtime.GOOS}
	// os.linesep
	osModule.Fields["linesep"] = &String{Value: "\n"}
	// os.curdir
	osModule.Fields["curdir"] = &String{Value: "."}
	// os.pardir
	osModule.Fields["pardir"] = &String{Value: ".."}

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
				return NewError("%s", err.Error())
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
				return NewError("%s", err.Error())
			}

			return convertToObject(data)
		},
	}

	// json.dump
	jsonModule.Fields["dump"] = &Builtin{Name: "json.dump", Fn: func(args ...Object) Object {
		if len(args) < 2 {
			return NewTypeError("dump() takes at least 2 arguments")
		}
		goValue := convertToGoValue(args[0])
		filename, ok := args[1].(*String)
		if !ok {
			return NewTypeError("dump() second argument must be a string (filename)")
		}
		data, err := json.MarshalIndent(goValue, "", "  ")
		if err != nil {
			return NewError("%s", err.Error())
		}
		if err := os.WriteFile(filename.Value, data, 0644); err != nil {
			return NewError("%s", err.Error())
		}
		return None_
	}}

	// json.load
	jsonModule.Fields["load"] = &Builtin{Name: "json.load", Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return NewTypeError("load() takes at least 1 argument")
		}
		filename, ok := args[0].(*String)
		if !ok {
			return NewTypeError("load() argument must be a string (filename)")
		}
		data, err := os.ReadFile(filename.Value)
		if err != nil {
			return NewError("%s", err.Error())
		}
		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			return NewError("%s", err.Error())
		}
		return convertToObject(value)
	}}

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

	// random.randrange
	randomModule.Fields["randrange"] = &Builtin{Name: "random.randrange", Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 3 {
			return NewTypeError("randrange() takes 1-3 arguments")
		}
		start := int64(0)
		var stop int64
		if len(args) == 1 {
			s, ok := args[0].(*Integer); if !ok {
				return NewTypeError("randrange() arguments must be integers")
			}
			stop = s.Value
		} else {
			s, ok := args[0].(*Integer); if !ok {
				return NewTypeError("randrange() arguments must be integers")
			}
			start = s.Value
			e, ok := args[1].(*Integer); if !ok {
				return NewTypeError("randrange() arguments must be integers")
			}
			stop = e.Value
		}
		if stop <= start {
			return NewValueError("empty range for randrange()")
		}
		return &Integer{Value: start + rand.Int63n(stop-start)}
	}}

	// random.sample
	randomModule.Fields["sample"] = &Builtin{Name: "random.sample", Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return NewTypeError("sample() takes exactly 2 arguments")
		}
		list, ok := args[0].(*List); if !ok {
			return NewTypeError("sample() first argument must be a list")
		}
		k, ok := args[1].(*Integer); if !ok {
			return NewTypeError("sample() second argument must be an integer")
		}
		if k.Value > int64(len(list.Elements)) {
			return NewValueError("sample larger than population")
		}
		shuffled := make([]Object, len(list.Elements))
		copy(shuffled, list.Elements)
		rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		return &List{Elements: shuffled[:k.Value]}
	}}

	// random.choices
	randomModule.Fields["choices"] = &Builtin{Name: "random.choices", Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return NewTypeError("choices() takes at least 1 argument")
		}
		list, ok := args[0].(*List); if !ok {
			return NewTypeError("choices() first argument must be a list")
		}
		k := 1
		if len(args) >= 2 {
			if i, ok := args[1].(*Integer); ok {
				k = int(i.Value)
			}
		}
		result := make([]Object, k)
		for i := 0; i < k; i++ {
			result[i] = list.Elements[rand.Intn(len(list.Elements))]
		}
		return &List{Elements: result}
	}}

	// random.gauss
	randomModule.Fields["gauss"] = &Builtin{Name: "random.gauss", Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return NewTypeError("gauss() takes exactly 2 arguments")
		}
		mu, ok1 := toFloat(args[0]); sigma, ok2 := toFloat(args[1])
		if !ok1 || !ok2 {
			return NewTypeError("gauss() arguments must be numbers")
		}
		// Box-Muller transform
		u1 := rand.Float64(); u2 := rand.Float64()
		z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		return &Float{Value: mu + sigma*z}
	}}

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

	// string.maketrans - 创建字符映射表用于 str.translate()
	// maketrans(from, to) 或 maketrans(dict)
	stringModule.Fields["maketrans"] = &Builtin{
		Name: "string.maketrans",
		Fn: func(args ...Object) Object {
			if len(args) == 0 {
				return NewTypeError("maketrans() takes at least 1 argument")
			}

			// maketrans(dict) 形式
			if len(args) == 1 {
				dictArg, ok := args[0].(*Dict)
				if !ok {
					return NewTypeError("maketrans() argument must be a dict")
				}
				return dictArg
			}

			// maketrans(from, to) 形式
			fromStr, ok1 := args[0].(*String)
			toStr, ok2 := args[1].(*String)
			if !ok1 || !ok2 {
				return NewTypeError("maketrans() both arguments must be strings")
			}

			if len(fromStr.Value) != len(toStr.Value) {
				return NewTypeError("maketrans() arguments must have same length")
			}

			result := NewDict()
			for i, r := range fromStr.Value {
				result.Set(&Integer{Value: int64(r)}, &String{Value: string(toStr.Value[i])})
			}
			return result
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

	// time.strftime
	timeModule.Fields["strftime"] = &Builtin{Name: "time.strftime", Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return NewTypeError("strftime() takes at least 1 argument")
		}
		format, ok := args[0].(*String); if !ok {
			return NewTypeError("strftime() argument must be a string")
		}
		t := time.Now()
		if len(args) >= 2 {
			if ts, ok := args[1].(*Float); ok {
				t = time.Unix(0, int64(ts.Value*1e9))
			}
			if ts, ok := args[1].(*Integer); ok {
				t = time.Unix(ts.Value, 0)
			}
		}
		// Simple Python format -> Go format mapping
		pyFormat := format.Value
		pyFormat = strings.ReplaceAll(pyFormat, "%Y", "2006")
		pyFormat = strings.ReplaceAll(pyFormat, "%m", "01")
		pyFormat = strings.ReplaceAll(pyFormat, "%d", "02")
		pyFormat = strings.ReplaceAll(pyFormat, "%H", "15")
		pyFormat = strings.ReplaceAll(pyFormat, "%M", "04")
		pyFormat = strings.ReplaceAll(pyFormat, "%S", "05")
		return &String{Value: t.Format(pyFormat)}
	}}

	// time.strptime
	timeModule.Fields["strptime"] = &Builtin{Name: "time.strptime", Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return NewTypeError("strptime() takes at least 1 argument")
		}
		s, ok := args[0].(*String); if !ok {
			return NewTypeError("strptime() argument must be a string")
		}
		format := "%Y-%m-%d %H:%M:%S"
		if len(args) >= 2 {
			if f, ok := args[1].(*String); ok {
				format = f.Value
			}
		}
		// Convert Python format to Go format (simplified)
		goFormat := format
		goFormat = strings.ReplaceAll(goFormat, "%Y", "2006")
		goFormat = strings.ReplaceAll(goFormat, "%m", "01")
		goFormat = strings.ReplaceAll(goFormat, "%d", "02")
		goFormat = strings.ReplaceAll(goFormat, "%H", "15")
		goFormat = strings.ReplaceAll(goFormat, "%M", "04")
		goFormat = strings.ReplaceAll(goFormat, "%S", "05")
		t, err := time.Parse(goFormat, s.Value)
		if err != nil {
			return NewValueError("time data does not match format")
		}
		return &Tuple{Elements: []Object{
			&Integer{Value: int64(t.Year())}, &Integer{Value: int64(t.Month())},
			&Integer{Value: int64(t.Day())}, &Integer{Value: int64(t.Hour())},
			&Integer{Value: int64(t.Minute())}, &Integer{Value: int64(t.Second())},
			&Integer{Value: int64(t.Weekday())}, &Integer{Value: int64(t.YearDay())},
			&Integer{Value: -1},
		}}
	}}

	// time.mktime
	timeModule.Fields["mktime"] = &Builtin{Name: "time.mktime", Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return NewTypeError("mktime() takes exactly 1 argument")
		}
		t := time.Now()
		return &Float{Value: float64(t.Unix())}
	}}

	// time.time_ns
	timeModule.Fields["time_ns"] = &Builtin{Name: "time.time_ns", Fn: func(args ...Object) Object {
		return &Integer{Value: time.Now().UnixNano()}
	}}

	// time.monotonic
	timeModule.Fields["monotonic"] = &Builtin{Name: "time.monotonic", Fn: func(args ...Object) Object {
		return &Float{Value: float64(time.Now().UnixNano()) / 1e9}
	}}

	// time.perf_counter
	timeModule.Fields["perf_counter"] = &Builtin{Name: "time.perf_counter", Fn: func(args ...Object) Object {
		return &Float{Value: float64(time.Now().UnixNano()) / 1e9}
	}}

	// time.gmtime
	timeModule.Fields["gmtime"] = &Builtin{Name: "time.gmtime", Fn: func(args ...Object) Object {
		t := time.Now().UTC()
		if len(args) >= 1 {
			if ts, ok := args[0].(*Float); ok {
				t = time.Unix(0, int64(ts.Value*1e9)).UTC()
			}
			if ts, ok := args[0].(*Integer); ok {
				t = time.Unix(ts.Value, 0).UTC()
			}
		}
		return &Tuple{Elements: []Object{
			&Integer{Value: int64(t.Year())}, &Integer{Value: int64(t.Month())},
			&Integer{Value: int64(t.Day())}, &Integer{Value: int64(t.Hour())},
			&Integer{Value: int64(t.Minute())}, &Integer{Value: int64(t.Second())},
			&Integer{Value: int64(t.Weekday())}, &Integer{Value: int64(t.YearDay())},
		}}
	}}

	// time.timezone
	timeModule.Fields["timezone"] = &Integer{Value: 0}
	// time.tzname
	timeModule.Fields["tzname"] = &Tuple{Elements: []Object{&String{Value: "UTC"}, &String{Value: "UTC"}}}

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
