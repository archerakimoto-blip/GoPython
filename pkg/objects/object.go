package objects

import (
	"encoding/json"
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
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

type Bytes struct {
	Value []byte
}

func NewBytes(data []byte) *Bytes {
	return &Bytes{Value: data}
}

func (b *Bytes) Type() ObjectType { return BYTES_OBJ }
func (b *Bytes) Inspect() string  { return fmt.Sprintf("b'%s'", string(b.Value)) }

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

type RegexPattern struct {
	Pattern string
	Flags   int
	regex   interface{} // compiled Go regex, stored as interface{} to avoid import
}

func NewRegexPattern(pattern string, flags int, compiledRegex interface{}) *RegexPattern {
	return &RegexPattern{Pattern: pattern, Flags: flags, regex: compiledRegex}
}

func (rp *RegexPattern) GetRegex() interface{} { return rp.regex }

func (rp *RegexPattern) Type() ObjectType { return REGEX_PATTERN_OBJ }
func (rp *RegexPattern) Inspect() string  { return fmt.Sprintf("<re.Pattern '%s'>", rp.Pattern) }

type RegexMatch struct {
	MatchString string
	Groups      []string
	StartPos    int
	EndPos      int
}

func (rm *RegexMatch) Type() ObjectType { return REGEX_MATCH_OBJ }
func (rm *RegexMatch) Inspect() string {
	if len(rm.Groups) > 0 {
		return fmt.Sprintf("<re.Match object; span=(%d, %d), match='%s'>", rm.StartPos, rm.EndPos, rm.Groups[0])
	}
	return "<re.Match object>"
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

var modules = make(map[string]*Module)

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
