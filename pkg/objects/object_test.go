package objects

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"testing"
)

// ==================== Integer Tests ====================

func TestIntegerType(t *testing.T) {
	i := &Integer{Value: 42}
	if i.Type() != INTEGER_OBJ {
		t.Errorf("expected %s, got %s", INTEGER_OBJ, i.Type())
	}
}

func TestIntegerInspect(t *testing.T) {
	tests := []struct {
		value    int64
		expected string
	}{
		{0, "0"}, {42, "42"}, {-1, "-1"}, {999999, "999999"},
	}
	for _, tt := range tests {
		i := &Integer{Value: tt.value}
		if i.Inspect() != tt.expected {
			t.Errorf("Integer{%d}.Inspect() = %s, want %s", tt.value, i.Inspect(), tt.expected)
		}
	}
}

func TestGetCachedInteger(t *testing.T) {
	for _, v := range []int64{-256, 0, 1, 255, 100} {
		got := GetCachedInteger(v)
		if got.Value != v {
			t.Errorf("GetCachedInteger(%d).Value = %d, want %d", v, got.Value, v)
		}
	}
	if got := GetCachedInteger(1000); got.Value != 1000 {
		t.Errorf("GetCachedInteger(1000).Value = %d, want 1000", got.Value)
	}
}

// ==================== Float Tests ====================

func TestFloatType(t *testing.T) {
	f := &Float{Value: 3.14}
	if f.Type() != FLOAT_OBJ {
		t.Errorf("expected %s, got %s", FLOAT_OBJ, f.Type())
	}
}

func TestFloatInspect(t *testing.T) {
	tests := []struct {
		value    float64
		expected string
	}{
		{3.14, "3.14"}, {0.0, "0"}, {-1.5, "-1.5"}, {1e10, "1e+10"},
	}
	for _, tt := range tests {
		f := &Float{Value: tt.value}
		if f.Inspect() != tt.expected {
			t.Errorf("Float{%g}.Inspect() = %s, want %s", tt.value, f.Inspect(), tt.expected)
		}
	}
}

// ==================== Complex Tests ====================

func TestComplexType(t *testing.T) {
	c := NewComplex(1, 2)
	if c.Type() != COMPLEX_OBJ {
		t.Errorf("expected %s, got %s", COMPLEX_OBJ, c.Type())
	}
}

func TestComplexInspect(t *testing.T) {
	tests := []struct {
		real, imag float64
		expected   string
	}{
		{0, 1, "1j"}, {0, -1, "-1j"}, {1, 2, "(1+2j)"},
		{1, -2, "(1-2j)"}, {3.5, 4.5, "(3.5+4.5j)"},
	}
	for _, tt := range tests {
		c := NewComplex(tt.real, tt.imag)
		if c.Inspect() != tt.expected {
			t.Errorf("Complex{%g,%g}.Inspect() = %s, want %s", tt.real, tt.imag, c.Inspect(), tt.expected)
		}
	}
}

// ==================== Boolean Tests ====================

func TestBooleanType(t *testing.T) {
	if (&Boolean{Value: true}).Type() != BOOLEAN_OBJ {
		t.Errorf("expected BOOLEAN_OBJ")
	}
}

func TestBooleanInspect(t *testing.T) {
	if True.Inspect() != "true" {
		t.Errorf("True.Inspect() = %s, want true", True.Inspect())
	}
	if False.Inspect() != "false" {
		t.Errorf("False.Inspect() = %s, want false", False.Inspect())
	}
}

// ==================== None Tests ====================

func TestNoneType(t *testing.T) {
	if (&None{}).Type() != NONE_OBJ {
		t.Errorf("expected NONE_OBJ")
	}
}

func TestNoneInspect(t *testing.T) {
	if None_.Inspect() != "None" {
		t.Errorf("got %s, want None", None_.Inspect())
	}
}

// ==================== Ellipsis Tests ====================

func TestEllipsisType(t *testing.T) {
	if (&Ellipsis{}).Type() != ELLIPSIS_OBJ {
		t.Errorf("expected ELLIPSIS_OBJ")
	}
}

func TestEllipsisInspect(t *testing.T) {
	if EllipsisSingleton.Inspect() != "Ellipsis" {
		t.Errorf("got %s, want Ellipsis", EllipsisSingleton.Inspect())
	}
}

// ==================== String Tests ====================

func TestStringType(t *testing.T) {
	if (&String{Value: "hello"}).Type() != STRING_OBJ {
		t.Errorf("expected STRING_OBJ")
	}
}

func TestStringInspect(t *testing.T) {
	s := &String{Value: "hello"}
	if s.Inspect() != "hello" {
		t.Errorf("got %s, want hello", s.Inspect())
	}
}

func TestGetCachedString(t *testing.T) {
	s1 := GetCachedString("abc")
	s2 := GetCachedString("abc")
	if s1.Value != "abc" || s2.Value != "abc" {
		t.Error("GetCachedString should return correct value")
	}
	longStr := "this_is_a_very_long_string_that_exceeds_cache"
	if s3 := GetCachedString(longStr); s3.Value != longStr {
		t.Error("GetCachedString should return correct value for long strings")
	}
}

func TestStringGetAttr_IsDigit(t *testing.T) {
	s := &String{Value: "123"}
	attr, ok := s.GetAttr("isdigit")
	if !ok {
		t.Fatal("expected isdigit attribute")
	}
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isdigit('123') should be True")
	}
	s2 := &String{Value: "12a"}
	attr2, _ := s2.GetAttr("isdigit")
	if attr2.(*Builtin).Fn() != False {
		t.Errorf("isdigit('12a') should be False")
	}
	s3 := &String{Value: ""}
	attr3, _ := s3.GetAttr("isdigit")
	if attr3.(*Builtin).Fn() != False {
		t.Errorf("isdigit('') should be False")
	}
}

func TestStringGetAttr_IsAlpha(t *testing.T) {
	s := &String{Value: "abc"}
	attr, _ := s.GetAttr("isalpha")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isalpha('abc') should be True")
	}
	s2 := &String{Value: "ab1"}
	attr2, _ := s2.GetAttr("isalpha")
	if attr2.(*Builtin).Fn() != False {
		t.Errorf("isalpha('ab1') should be False")
	}
}

func TestStringGetAttr_IsAlnum(t *testing.T) {
	s := &String{Value: "abc123"}
	attr, _ := s.GetAttr("isalnum")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isalnum('abc123') should be True")
	}
}

func TestStringGetAttr_IsSpace(t *testing.T) {
	s := &String{Value: " \t\n"}
	attr, _ := s.GetAttr("isspace")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isspace should be True")
	}
}

func TestStringGetAttr_IsUpper(t *testing.T) {
	s := &String{Value: "ABC"}
	attr, _ := s.GetAttr("isupper")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isupper('ABC') should be True")
	}
}

func TestStringGetAttr_IsLower(t *testing.T) {
	s := &String{Value: "abc"}
	attr, _ := s.GetAttr("islower")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("islower('abc') should be True")
	}
}

func TestStringGetAttr_IsTitle(t *testing.T) {
	s := &String{Value: "Hello World"}
	attr, _ := s.GetAttr("istitle")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("istitle('Hello World') should be True")
	}
}

func TestStringGetAttr_Capitalize(t *testing.T) {
	s := &String{Value: "hello"}
	attr, _ := s.GetAttr("capitalize")
	if attr.(*Builtin).Fn().Inspect() != "Hello" {
		t.Errorf("capitalize('hello') should be 'Hello'")
	}
}

func TestStringGetAttr_Title(t *testing.T) {
	s := &String{Value: "hello world"}
	attr, _ := s.GetAttr("title")
	if attr.(*Builtin).Fn().Inspect() != "Hello World" {
		t.Errorf("title('hello world') should be 'Hello World'")
	}
}

func TestStringGetAttr_Swapcase(t *testing.T) {
	s := &String{Value: "Hello"}
	attr, _ := s.GetAttr("swapcase")
	if attr.(*Builtin).Fn().Inspect() != "hELLO" {
		t.Errorf("swapcase('Hello') should be 'hELLO'")
	}
}

func TestStringGetAttr_Center(t *testing.T) {
	s := &String{Value: "hi"}
	attr, _ := s.GetAttr("center")
	if attr.(*Builtin).Fn(&Integer{Value: 10}).Inspect() != "    hi    " {
		t.Errorf("center('hi', 10) wrong")
	}
	if attr.(*Builtin).Fn(&Integer{Value: 1}).Inspect() != "hi" {
		t.Errorf("center('hi', 1) should return 'hi'")
	}
}

func TestStringGetAttr_Ljust(t *testing.T) {
	s := &String{Value: "hi"}
	attr, _ := s.GetAttr("ljust")
	if attr.(*Builtin).Fn(&Integer{Value: 5}).Inspect() != "hi   " {
		t.Errorf("ljust('hi', 5) wrong")
	}
}

func TestStringGetAttr_Rjust(t *testing.T) {
	s := &String{Value: "hi"}
	attr, _ := s.GetAttr("rjust")
	if attr.(*Builtin).Fn(&Integer{Value: 5}).Inspect() != "   hi" {
		t.Errorf("rjust('hi', 5) wrong")
	}
}

func TestStringGetAttr_Zfill(t *testing.T) {
	s := &String{Value: "42"}
	attr, _ := s.GetAttr("zfill")
	if attr.(*Builtin).Fn(&Integer{Value: 5}).Inspect() != "00042" {
		t.Errorf("zfill('42', 5) wrong")
	}
	s2 := &String{Value: "-42"}
	attr2, _ := s2.GetAttr("zfill")
	if attr2.(*Builtin).Fn(&Integer{Value: 5}).Inspect() != "-0042" {
		t.Errorf("zfill('-42', 5) wrong")
	}
}

func TestStringGetAttr_Partition(t *testing.T) {
	s := &String{Value: "hello world"}
	attr, _ := s.GetAttr("partition")
	result := attr.(*Builtin).Fn(&String{Value: " "})
	tup := result.(*Tuple)
	if len(tup.Elements) != 3 || tup.Elements[0].Inspect() != "hello" {
		t.Errorf("partition failed")
	}
	s2 := &String{Value: "hello"}
	attr2, _ := s2.GetAttr("partition")
	result2 := attr2.(*Builtin).Fn(&String{Value: "x"})
	if result2.(*Tuple).Elements[0].Inspect() != "hello" {
		t.Errorf("partition not found: first element should be original string")
	}
}

func TestStringGetAttr_RPartition(t *testing.T) {
	s := &String{Value: "hello world hello"}
	attr, _ := s.GetAttr("rpartition")
	result := attr.(*Builtin).Fn(&String{Value: " "})
	if result.(*Tuple).Elements[0].Inspect() != "hello world" {
		t.Errorf("rpartition failed")
	}
}

func TestStringGetAttr_Encode(t *testing.T) {
	s := &String{Value: "hello"}
	attr, _ := s.GetAttr("encode")
	result := attr.(*Builtin).Fn()
	if string(result.(*Bytes).Value) != "hello" {
		t.Errorf("encode('hello') wrong")
	}
}

func TestStringGetAttr_Count(t *testing.T) {
	s := &String{Value: "hello hello"}
	attr, _ := s.GetAttr("count")
	if attr.(*Builtin).Fn(&String{Value: "hello"}).(*Integer).Value != 2 {
		t.Errorf("count('hello') wrong")
	}
}

func TestStringGetAttr_Rfind(t *testing.T) {
	s := &String{Value: "hello hello"}
	attr, _ := s.GetAttr("rfind")
	if attr.(*Builtin).Fn(&String{Value: "hello"}).(*Integer).Value != 6 {
		t.Errorf("rfind('hello') wrong")
	}
	if attr.(*Builtin).Fn(&String{Value: "xyz"}).(*Integer).Value != -1 {
		t.Errorf("rfind('xyz') should be -1")
	}
}

func TestStringGetAttr_Rindex(t *testing.T) {
	s := &String{Value: "hello hello"}
	attr, _ := s.GetAttr("rindex")
	if attr.(*Builtin).Fn(&String{Value: "hello"}).(*Integer).Value != 6 {
		t.Errorf("rindex('hello') wrong")
	}
	if attr.(*Builtin).Fn(&String{Value: "xyz"}).Type() != ERROR_OBJ {
		t.Errorf("rindex('xyz') should return error")
	}
}

func TestStringGetAttr_IsDecimal(t *testing.T) {
	s := &String{Value: "123"}
	attr, _ := s.GetAttr("isdecimal")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isdecimal('123') should be True")
	}
}

func TestStringGetAttr_IsNumeric(t *testing.T) {
	s := &String{Value: "123"}
	attr, _ := s.GetAttr("isnumeric")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isnumeric('123') should be True")
	}
}

func TestStringGetAttr_IsIdentifier(t *testing.T) {
	tests := []struct {
		input string
		exp   bool
	}{
		{"hello", true}, {"_foo", true}, {"123abc", false}, {"", false}, {"abc123", true},
	}
	for _, tt := range tests {
		s := &String{Value: tt.input}
		attr, _ := s.GetAttr("isidentifier")
		if (attr.(*Builtin).Fn() == True) != tt.exp {
			t.Errorf("isidentifier(%q) wrong", tt.input)
		}
	}
}

func TestStringGetAttr_IsPrintable(t *testing.T) {
	s := &String{Value: "hello"}
	attr, _ := s.GetAttr("isprintable")
	if attr.(*Builtin).Fn() != True {
		t.Errorf("isprintable('hello') should be True")
	}
}

func TestStringGetAttr_ExpandTabs(t *testing.T) {
	s := &String{Value: "a\tb"}
	attr, _ := s.GetAttr("expandtabs")
	if attr.(*Builtin).Fn(&Integer{Value: 4}).Inspect() != "a    b" {
		t.Errorf("expandtabs wrong")
	}
}

func TestStringGetAttr_Translate(t *testing.T) {
	table := NewDict()
	table.Set(&Integer{Value: 97}, &String{Value: "x"})
	s := &String{Value: "abc"}
	attr, _ := s.GetAttr("translate")
	if attr.(*Builtin).Fn(table).Inspect() != "xbc" {
		t.Errorf("translate('abc') wrong")
	}
}

func TestStringGetAttr_FormatMap(t *testing.T) {
	m := NewDict()
	m.Set(&String{Value: "name"}, &String{Value: "world"})
	s := &String{Value: "hello {name}"}
	attr, _ := s.GetAttr("format_map")
	if attr.(*Builtin).Fn(m).Inspect() != "hello world" {
		t.Errorf("format_map wrong")
	}
}

func TestStringGetAttr_Unknown(t *testing.T) {
	s := &String{Value: "hello"}
	_, ok := s.GetAttr("nonexistent")
	if ok {
		t.Error("GetAttr for unknown should return false")
	}
}

// ==================== Bytes Tests ====================

func TestBytesType(t *testing.T) {
	if NewBytes([]byte("hello")).Type() != BYTES_OBJ {
		t.Errorf("expected BYTES_OBJ")
	}
}

func TestBytesInspect(t *testing.T) {
	if NewBytes([]byte("hello")).Inspect() != "b'hello'" {
		t.Errorf("Bytes inspect wrong")
	}
}

// ==================== List Tests ====================

func TestListType(t *testing.T) {
	if NewList(nil).Type() != LIST_OBJ {
		t.Errorf("expected LIST_OBJ")
	}
}

func TestListInspect(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	if l.Inspect() != "[1, 2]" {
		t.Errorf("got %s, want [1, 2]", l.Inspect())
	}
}

func TestListAppend(t *testing.T) {
	l := NewList(nil)
	l.Append(&Integer{Value: 1})
	if len(l.Elements) != 1 {
		t.Errorf("expected length 1")
	}
}

func TestListExtend(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}})
	l.Extend(NewList([]Object{&Integer{Value: 2}, &Integer{Value: 3}}))
	if len(l.Elements) != 3 {
		t.Errorf("expected length 3")
	}
}

func TestListPop(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}, &Integer{Value: 3}})
	obj, err := l.Pop()
	if err != nil || obj.(*Integer).Value != 3 {
		t.Errorf("Pop() wrong")
	}
	obj2, err2 := l.Pop(0)
	if err2 != nil || obj2.(*Integer).Value != 1 {
		t.Errorf("Pop(0) wrong")
	}
	if _, err3 := l.Pop(10); err3 == nil {
		t.Error("Pop(10) should return error")
	}
	l2 := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	if obj4, _ := l2.Pop(-1); obj4.(*Integer).Value != 2 {
		t.Errorf("Pop(-1) wrong")
	}
}

func TestListIndex(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	if l.Index(&Integer{Value: 2}) != 1 {
		t.Errorf("Index(2) wrong")
	}
	if l.Index(&Integer{Value: 99}) != -1 {
		t.Errorf("Index(99) should be -1")
	}
}

func TestListContains(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}})
	if !l.Contains(&Integer{Value: 1}) {
		t.Error("should contain 1")
	}
	if l.Contains(&Integer{Value: 99}) {
		t.Error("should not contain 99")
	}
}

func TestListInsert(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 3}})
	l.Insert(1, &Integer{Value: 2})
	if l.Elements[1].(*Integer).Value != 2 {
		t.Errorf("Insert failed")
	}
	l2 := NewList([]Object{&Integer{Value: 1}})
	l2.Insert(-1, &Integer{Value: 0})
	if l2.Elements[0].(*Integer).Value != 0 {
		t.Errorf("Insert(-1) failed")
	}
	l3 := NewList([]Object{&Integer{Value: 1}})
	l3.Insert(100, &Integer{Value: 2})
	if len(l3.Elements) != 2 {
		t.Errorf("Insert beyond end should append")
	}
}

func TestListRemove(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	if err := l.Remove(&Integer{Value: 2}); err != nil || len(l.Elements) != 1 {
		t.Errorf("Remove failed")
	}
	if err := l.Remove(&Integer{Value: 99}); err == nil {
		t.Error("Remove(99) should return error")
	}
}

func TestListReverse(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}, &Integer{Value: 3}})
	l.Reverse()
	if l.Elements[0].(*Integer).Value != 3 {
		t.Errorf("Reverse failed")
	}
}

func TestListSizeAndClear(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	if l.Size() != 2 {
		t.Errorf("Size() wrong")
	}
	l.Clear()
	if len(l.Elements) != 0 {
		t.Errorf("Clear() failed")
	}
}

func TestListGetAttr_Sort(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 3}, &Integer{Value: 1}, &Integer{Value: 2}})
	attr, ok := l.GetAttr("sort")
	if !ok {
		t.Fatal("expected sort attribute")
	}
	attr.(*Builtin).Fn()
	if l.Elements[0].(*Integer).Value != 1 {
		t.Errorf("Sort failed")
	}
	l2 := NewList([]Object{&Integer{Value: 3}, &Integer{Value: 1}, &Integer{Value: 2}})
	attr2, _ := l2.GetAttr("sort")
	attr2.(*Builtin).Fn(True)
	if l2.Elements[0].(*Integer).Value != 3 {
		t.Errorf("Sort reverse failed")
	}
}

func TestListGetAttr_IMul(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}})
	attr, _ := l.GetAttr("__imul__")
	attr.(*Builtin).Fn(&Integer{Value: 3})
	if len(l.Elements) != 3 {
		t.Errorf("__imul__ failed")
	}
}

func TestListGetAttr_IAdd(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}})
	attr, _ := l.GetAttr("__iadd__")
	attr.(*Builtin).Fn(NewList([]Object{&Integer{Value: 2}}))
	if len(l.Elements) != 2 {
		t.Errorf("__iadd__ failed")
	}
}

func TestListGetAttr_Unknown(t *testing.T) {
	l := NewList(nil)
	_, ok := l.GetAttr("nonexistent")
	if ok {
		t.Error("GetAttr for unknown should return false")
	}
}

// ==================== Tuple Tests ====================

func TestTupleType(t *testing.T) {
	if NewTuple(nil).Type() != TUPLE_OBJ {
		t.Errorf("expected TUPLE_OBJ")
	}
}

func TestTupleInspect(t *testing.T) {
	if NewTuple([]Object{&Integer{Value: 1}, &Integer{Value: 2}}).Inspect() != "(1, 2)" {
		t.Errorf("Tuple inspect wrong")
	}
	if NewTuple(nil).Inspect() != "()" {
		t.Errorf("Empty tuple inspect wrong")
	}
}

// ==================== Set Tests ====================

func TestSetType(t *testing.T) {
	if NewSet().Type() != SET_OBJ {
		t.Errorf("expected SET_OBJ")
	}
}

func TestSetAddContainsRemove(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	if !s.Contains(&Integer{Value: 1}) {
		t.Error("should contain 1")
	}
	s.Remove(&Integer{Value: 1})
	if s.Contains(&Integer{Value: 1}) {
		t.Error("should not contain 1 after Remove")
	}
}

func TestSetSize(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	s.Add(&Integer{Value: 2})
	if s.Size() != 2 {
		t.Errorf("Size() wrong")
	}
}

func TestSetOperations(t *testing.T) {
	s1 := NewSet()
	s1.Add(&Integer{Value: 1})
	s1.Add(&Integer{Value: 2})
	s2 := NewSet()
	s2.Add(&Integer{Value: 2})
	s2.Add(&Integer{Value: 3})
	if s1.Union(s2).Size() != 3 {
		t.Errorf("Union size wrong")
	}
	if s1.Intersection(s2).Size() != 1 {
		t.Errorf("Intersection size wrong")
	}
	if s1.Difference(s2).Size() != 1 {
		t.Errorf("Difference size wrong")
	}
	if s1.SymmetricDifference(s2).Size() != 2 {
		t.Errorf("SymmetricDifference size wrong")
	}
}

func TestSetToSlice(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	if len(s.ToSlice()) != 1 {
		t.Errorf("ToSlice wrong")
	}
}

func TestSetHashKey(t *testing.T) {
	s := NewSet()
	tests := []struct {
		obj Object
		exp string
	}{
		{&Integer{Value: 42}, "int:42"},
		{&Float{Value: 3.14}, "float:3.14"},
		{&Boolean{Value: true}, "bool:true"},
		{&String{Value: "hi"}, "str:hi"},
	}
	for _, tt := range tests {
		if got := s.HashKey(tt.obj); got != tt.exp {
			t.Errorf("HashKey(%v) = %s, want %s", tt.obj, got, tt.exp)
		}
	}
}

func TestSetGetAttr_Methods(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	s.Add(&Integer{Value: 2})
	s2 := NewSet()
	s2.Add(&Integer{Value: 2})
	s2.Add(&Integer{Value: 3})

	// union
	attr, _ := s.GetAttr("union")
	if attr.(*Builtin).Fn(s2).(*Set).Size() != 3 {
		t.Errorf("union wrong")
	}
	// intersection
	attr2, _ := s.GetAttr("intersection")
	if attr2.(*Builtin).Fn(s2).(*Set).Size() != 1 {
		t.Errorf("intersection wrong")
	}
	// difference
	attr3, _ := s.GetAttr("difference")
	if attr3.(*Builtin).Fn(s2).(*Set).Size() != 1 {
		t.Errorf("difference wrong")
	}
	// symmetric_difference
	attr4, _ := s.GetAttr("symmetric_difference")
	if attr4.(*Builtin).Fn(s2).(*Set).Size() != 2 {
		t.Errorf("symmetric_difference wrong")
	}
	// issubset
	s3 := NewSet()
	s3.Add(&Integer{Value: 1})
	attr5, _ := s3.GetAttr("issubset")
	if attr5.(*Builtin).Fn(s) != True {
		t.Error("issubset wrong")
	}
	// issuperset
	attr6, _ := s.GetAttr("issuperset")
	if attr6.(*Builtin).Fn(s3) != True {
		t.Error("issuperset wrong")
	}
	// copy
	attr7, _ := s.GetAttr("copy")
	if attr7.(*Builtin).Fn().(*Set).Size() != 2 {
		t.Errorf("copy wrong")
	}
	// __isub__
	s4 := NewSet()
	s4.Add(&Integer{Value: 1})
	s4.Add(&Integer{Value: 2})
	attr8, _ := s4.GetAttr("__isub__")
	attr8.(*Builtin).Fn(s3)
	if s4.Contains(&Integer{Value: 1}) {
		t.Error("__isub__ should remove elements")
	}
	// __iand__
	s5 := NewSet()
	s5.Add(&Integer{Value: 1})
	s5.Add(&Integer{Value: 2})
	attr9, _ := s5.GetAttr("__iand__")
	attr9.(*Builtin).Fn(s2)
	if s5.Contains(&Integer{Value: 1}) {
		t.Error("__iand__ should remove non-common")
	}
	// __ixor__
	s6 := NewSet()
	s6.Add(&Integer{Value: 1})
	s6.Add(&Integer{Value: 2})
	attr10, _ := s6.GetAttr("__ixor__")
	attr10.(*Builtin).Fn(s2)
	if s6.Contains(&Integer{Value: 2}) {
		t.Error("__ixor__ should remove common")
	}
	// update
	s7 := NewSet()
	s7.Add(&Integer{Value: 1})
	attr11, _ := s7.GetAttr("update")
	attr11.(*Builtin).Fn(s2)
	if s7.Size() != 3 {
		t.Errorf("update wrong")
	}
	// difference_update
	s8 := NewSet()
	s8.Add(&Integer{Value: 1})
	s8.Add(&Integer{Value: 2})
	attr12, _ := s8.GetAttr("difference_update")
	attr12.(*Builtin).Fn(s3)
	if s8.Contains(&Integer{Value: 1}) {
		t.Error("difference_update should remove")
	}
	// __ior__
	s9 := NewSet()
	s9.Add(&Integer{Value: 1})
	attr13, _ := s9.GetAttr("__ior__")
	attr13.(*Builtin).Fn(s2)
	if s9.Size() != 3 {
		t.Errorf("__ior__ wrong")
	}
}

func TestSetGetAttr_Errors(t *testing.T) {
	s := NewSet()
	for name, attr := range map[string]string{
		"intersection": "intersection", "difference": "difference",
		"symmetric_difference": "symmetric_difference", "issubset": "issubset",
		"issuperset": "issuperset", "__isub__": "__isub__", "__iand__": "__iand__",
		"__ixor__": "__ixor__", "__ior__": "__ior__", "difference_update": "difference_update",
	} {
		a, _ := s.GetAttr(attr)
		if a.(*Builtin).Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
			t.Errorf("%s(int) should return error", name)
		}
	}
	_, ok := s.GetAttr("nonexistent")
	if ok {
		t.Error("GetAttr for unknown should return false")
	}
}

// ==================== Dict Tests ====================

func TestDictType(t *testing.T) {
	if NewDict().Type() != DICT_OBJ {
		t.Errorf("expected DICT_OBJ")
	}
}

func TestDictGetSet(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "key"}, &Integer{Value: 42})
	val, ok := d.Get(&String{Value: "key"})
	if !ok || val.(*Integer).Value != 42 {
		t.Errorf("Get wrong")
	}
	_, ok2 := d.Get(&String{Value: "missing"})
	if ok2 {
		t.Error("Get for missing should return false")
	}
}

func TestDictHasDelete(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "key"}, &Integer{Value: 1})
	if !d.Has(&String{Value: "key"}) {
		t.Error("should have key")
	}
	d.Delete(&String{Value: "key"})
	if d.Has(&String{Value: "key"}) {
		t.Error("should not have key after Delete")
	}
}

func TestDictSize(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	d.Set(&String{Value: "b"}, &Integer{Value: 2})
	if d.Size() != 2 {
		t.Errorf("Size() wrong")
	}
}

func TestDictKeysValuesSlices(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	if len(d.KeysSlice()) != 1 {
		t.Errorf("KeysSlice wrong")
	}
	if len(d.ValuesSlice()) != 1 {
		t.Errorf("ValuesSlice wrong")
	}
}

func TestNewDictWithCapacity(t *testing.T) {
	if NewDictWithCapacity(10) == nil {
		t.Error("NewDictWithCapacity should return non-nil")
	}
}

func TestDictGetAttr_Methods(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "key"}, &Integer{Value: 42})

	// get
	attr, _ := d.GetAttr("get")
	if attr.(*Builtin).Fn(&String{Value: "key"}).(*Integer).Value != 42 {
		t.Errorf("get wrong")
	}
	if attr.(*Builtin).Fn(&String{Value: "missing"}, &Integer{Value: 0}).(*Integer).Value != 0 {
		t.Errorf("get with default wrong")
	}
	// pop
	attr2, _ := d.GetAttr("pop")
	if attr2.(*Builtin).Fn(&String{Value: "key"}).(*Integer).Value != 42 {
		t.Errorf("pop wrong")
	}
	// popitem
	d2 := NewDict()
	d2.Set(&String{Value: "k"}, &Integer{Value: 1})
	attr3, _ := d2.GetAttr("popitem")
	if len(attr3.(*Builtin).Fn().(*Tuple).Elements) != 2 {
		t.Errorf("popitem wrong")
	}
	// setdefault
	d3 := NewDict()
	d3.Set(&String{Value: "k"}, &Integer{Value: 1})
	attr4, _ := d3.GetAttr("setdefault")
	if attr4.(*Builtin).Fn(&String{Value: "k"}).(*Integer).Value != 1 {
		t.Errorf("setdefault existing wrong")
	}
	if attr4.(*Builtin).Fn(&String{Value: "new"}, &Integer{Value: 99}).(*Integer).Value != 99 {
		t.Errorf("setdefault new wrong")
	}
	// clear
	d4 := NewDict()
	d4.Set(&String{Value: "k"}, &Integer{Value: 1})
	attr5, _ := d4.GetAttr("clear")
	attr5.(*Builtin).Fn()
	if d4.Size() != 0 {
		t.Errorf("clear wrong")
	}
	// copy
	d5 := NewDict()
	d5.Set(&String{Value: "k"}, &Integer{Value: 1})
	attr6, _ := d5.GetAttr("copy")
	if attr6.(*Builtin).Fn().(*Dict).Size() != 1 {
		t.Errorf("copy wrong")
	}
	// keys/values/items
	d6 := NewDict()
	d6.Set(&String{Value: "a"}, &Integer{Value: 1})
	attr7, _ := d6.GetAttr("keys")
	if attr7.(*Builtin).Fn().Type() != DICT_KEYS_OBJ {
		t.Errorf("keys wrong type")
	}
	attr8, _ := d6.GetAttr("values")
	if attr8.(*Builtin).Fn().Type() != DICT_VALUES_OBJ {
		t.Errorf("values wrong type")
	}
	attr9, _ := d6.GetAttr("items")
	if attr9.(*Builtin).Fn().Type() != DICT_ITEMS_OBJ {
		t.Errorf("items wrong type")
	}
	// fromkeys
	attr10, _ := d.GetAttr("fromkeys")
	if attr10.(*Builtin).Fn(NewList([]Object{&String{Value: "a"}})).(*Dict).Size() != 1 {
		t.Errorf("fromkeys wrong")
	}
	// update
	d7 := NewDict()
	d7.Set(&String{Value: "a"}, &Integer{Value: 1})
	attr11, _ := d7.GetAttr("update")
	other := NewDict()
	other.Set(&String{Value: "b"}, &Integer{Value: 2})
	attr11.(*Builtin).Fn(other)
	if d7.Size() != 2 {
		t.Errorf("update wrong")
	}
}

func TestDictGetAttr_Errors(t *testing.T) {
	d := NewDict()
	for name, fn := range map[string]func() Object{
		"get":        func() Object { a, _ := d.GetAttr("get"); return a.(*Builtin).Fn() },
		"pop":        func() Object { a, _ := d.GetAttr("pop"); return a.(*Builtin).Fn() },
		"setdefault": func() Object { a, _ := d.GetAttr("setdefault"); return a.(*Builtin).Fn() },
		"fromkeys":   func() Object { a, _ := d.GetAttr("fromkeys"); return a.(*Builtin).Fn() },
	} {
		if fn().Type() != ERROR_OBJ {
			t.Errorf("%s() with no args should return error", name)
		}
	}
	attr, _ := d.GetAttr("popitem")
	if attr.(*Builtin).Fn().Type() != ERROR_OBJ {
		t.Error("popitem on empty dict should return error")
	}
	_, ok := d.GetAttr("nonexistent")
	if ok {
		t.Error("GetAttr for unknown should return false")
	}
}

// ==================== DictKeys/Values/Items Tests ====================

func TestDictViews(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})

	dk := NewDictKeys(d)
	if dk.Type() != DICT_KEYS_OBJ || dk.Inspect() != "dict_keys([a])" || dk.Len() != 1 {
		t.Errorf("DictKeys wrong")
	}
	if item, ok := dk.GetItem(0); !ok || item.Inspect() != "a" {
		t.Errorf("DictKeys GetItem wrong")
	}
	if _, ok := dk.GetItem(10); ok {
		t.Error("DictKeys GetItem out of range should fail")
	}
	if item, ok := dk.GetItem(-1); !ok || item.Inspect() != "a" {
		t.Errorf("DictKeys GetItem(-1) wrong")
	}

	dv := NewDictValues(d)
	if dv.Type() != DICT_VALUES_OBJ || dv.Len() != 1 {
		t.Errorf("DictValues wrong")
	}
	if item, ok := dv.GetItem(0); !ok || item.(*Integer).Value != 1 {
		t.Errorf("DictValues GetItem wrong")
	}

	di := NewDictItems(d)
	if di.Type() != DICT_ITEMS_OBJ || di.Len() != 1 {
		t.Errorf("DictItems wrong")
	}
	if item, ok := di.GetItem(0); !ok || len(item.(*Tuple).Elements) != 2 {
		t.Errorf("DictItems GetItem wrong")
	}
}

// ==================== Error Tests ====================

func TestErrorType(t *testing.T) {
	e := NewError("test")
	if e.Type() != ERROR_OBJ || e.Inspect() != "Error: test" || e.Error() != "Error: test" {
		t.Errorf("Error type/inspect/error wrong")
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		fn      func(string, ...interface{}) *Error
		errType string
	}{
		{NewValueError, "ValueError"}, {NewTypeError, "TypeError"},
		{NewIndexError, "IndexError"}, {NewKeyError, "KeyError"},
		{NewAttributeError, "AttributeError"}, {NewNameError, "NameError"},
		{NewZeroDivisionError, "ZeroDivisionError"}, {NewAssertionError, "AssertionError"},
		{NewRuntimeError, "RuntimeError"}, {NewNotImplementedError, "NotImplementedError"},
		{NewStopIteration, "StopIteration"}, {NewOverflowError, "OverflowError"},
		{NewFileNotFoundError, "FileNotFoundError"}, {NewImportError, "ImportError"},
		{NewSyntaxError, "SyntaxError"}, {NewIndentationError, "IndentationError"},
		{NewUnboundLocalError, "UnboundLocalError"}, {NewRecursionError, "RecursionError"},
		{NewMemoryError, "MemoryError"}, {NewException, "Exception"},
	}
	for _, tt := range tests {
		e := tt.fn("test %s", "msg")
		if e.ErrorType != tt.errType {
			t.Errorf("%s: ErrorType = %s, want %s", tt.errType, e.ErrorType, tt.errType)
		}
	}
}

func TestNewErrorWithType(t *testing.T) {
	e := NewErrorWithType("CustomError", "custom %d", 42)
	if e.ErrorType != "CustomError" || e.Message != "custom 42" {
		t.Errorf("NewErrorWithType wrong")
	}
}

// ==================== ExceptionGroup Tests ====================

func TestExceptionGroup(t *testing.T) {
	eg := &ExceptionGroup{Message: "test", Exceptions: []Object{NewError("e1")}}
	if eg.Type() != EXCEPTION_GROUP_OBJ || eg.Inspect() == "" {
		t.Errorf("ExceptionGroup wrong")
	}
}

// ==================== Builtin Tests ====================

func TestBuiltin(t *testing.T) {
	b := &Builtin{Name: "test", Fn: func(args ...Object) Object { return None_ }}
	if b.Type() != BUILTIN_OBJ || b.Inspect() != "builtin function: test" {
		t.Errorf("Builtin wrong")
	}
	b2 := &Builtin{Fn: func(args ...Object) Object { return None_ }}
	if b2.Inspect() != "builtin function" {
		t.Errorf("Builtin no name wrong")
	}
}

// ==================== ContextManager Tests ====================

func TestContextManager(t *testing.T) {
	cm := &ContextManager{}
	if cm.Type() != CONTEXT_OBJ || cm.Inspect() != "context manager" {
		t.Errorf("ContextManager wrong")
	}
}

// ==================== Generator Tests ====================

func TestGenerator(t *testing.T) {
	g := &Generator{}
	if g.Type() != GENERATOR_OBJ || g.Inspect() == "" {
		t.Errorf("Generator wrong")
	}
}

// ==================== Async Tests ====================

func TestAsync(t *testing.T) {
	a := &Async{}
	if a.Type() != ASYNC_OBJ || a.Inspect() == "" {
		t.Errorf("Async wrong")
	}
}

// ==================== Future Tests ====================

func TestFuture(t *testing.T) {
	f := &Future{Done: false}
	if f.Type() != FUTURE_OBJ || f.Inspect() != "future[pending]" {
		t.Errorf("Future pending wrong")
	}
	f2 := &Future{Done: true, Result: &Integer{Value: 42}}
	if f2.Inspect() != "future[result: 42]" {
		t.Errorf("Future done wrong")
	}
	f3 := &Future{Done: true, Error: fmt.Errorf("err")}
	if f3.Inspect() != "future[error: err]" {
		t.Errorf("Future error wrong")
	}
}

// ==================== Closure Tests ====================

func TestClosure(t *testing.T) {
	c := &Closure{}
	if c.Type() != FUNCTION_OBJ || c.Inspect() != "closure" {
		t.Errorf("Closure wrong")
	}
}

// ==================== Class Tests ====================

func TestClass(t *testing.T) {
	cls := &Class{Name: "MyClass"}
	if cls.Type() != CLASS_OBJ || cls.Inspect() != "<class MyClass>" {
		t.Errorf("Class wrong")
	}
}

func TestClassHasSlots(t *testing.T) {
	cls := &Class{Name: "A"}
	if cls.HasSlots() {
		t.Error("should not have slots")
	}
	cls.Slots = []string{"x"}
	if !cls.HasSlots() {
		t.Error("should have slots")
	}
}

func TestClassIsSlotAllowed(t *testing.T) {
	cls := &Class{Name: "A", Slots: []string{"x"}}
	if !cls.IsSlotAllowed("x") || cls.IsSlotAllowed("z") {
		t.Errorf("IsSlotAllowed wrong")
	}
	cls2 := &Class{Name: "B"}
	if !cls2.IsSlotAllowed("anything") {
		t.Error("no slots means all allowed")
	}
}

func TestClassFindClassAttr(t *testing.T) {
	cls := &Class{Name: "A", Methods: map[string]Object{"foo": &Builtin{Name: "foo"}}, Fields: map[string]Object{"bar": &Integer{Value: 42}}}
	if _, ok := cls.FindClassAttr("foo"); !ok {
		t.Error("should find method")
	}
	if _, ok := cls.FindClassAttr("bar"); !ok {
		t.Error("should find field")
	}
	if _, ok := cls.FindClassAttr("missing"); ok {
		t.Error("should not find missing")
	}
}

func TestClassComputeMRO(t *testing.T) {
	base := &Class{Name: "Base"}
	child := &Class{Name: "Child", SuperClass: base}
	if mro := child.ComputeMRO(); len(mro) != 1 || mro[0] != base {
		t.Errorf("MRO wrong")
	}
	if mro := (&Class{Name: "Solo"}).ComputeMRO(); len(mro) != 0 {
		t.Errorf("Solo MRO should be empty")
	}
}

func TestClassAllSlotNames(t *testing.T) {
	if (&Class{Name: "A"}).AllSlotNames() != nil {
		t.Error("no slots should return nil")
	}
	names := (&Class{Name: "B", Slots: []string{"x", "y"}}).AllSlotNames()
	if len(names) != 2 || names[0] != "x" || names[1] != "y" {
		t.Errorf("AllSlotNames wrong")
	}
}

func TestClassFindClassAttrWithSuper(t *testing.T) {
	base := &Class{Name: "Base", Methods: map[string]Object{"bm": &Builtin{Name: "bm"}}, Fields: map[string]Object{"bf": &Integer{Value: 1}}}
	child := &Class{Name: "Child", SuperClass: base, Methods: map[string]Object{}}
	if _, ok := child.FindClassAttr("bm"); !ok {
		t.Error("should find from SuperClass")
	}
	if _, ok := child.FindClassAttr("bf"); !ok {
		t.Error("should find field from SuperClass")
	}
}

func TestClassFindClassAttrWithSuperClasses(t *testing.T) {
	b1 := &Class{Name: "B1", Methods: map[string]Object{"m1": &Builtin{Name: "m1"}}}
	b2 := &Class{Name: "B2", Fields: map[string]Object{"f2": &Integer{Value: 2}}}
	child := &Class{Name: "Child", SuperClasses: []*Class{b1, b2}, Methods: map[string]Object{}}
	if _, ok := child.FindClassAttr("m1"); !ok {
		t.Error("should find from SuperClasses")
	}
	if _, ok := child.FindClassAttr("f2"); !ok {
		t.Error("should find field from SuperClasses")
	}
}

func TestClassIsSlotAllowedWithSuper(t *testing.T) {
	base := &Class{Name: "Base", Slots: []string{"x"}}
	child := &Class{Name: "Child", SuperClass: base}
	if !child.IsSlotAllowed("x") {
		t.Error("x should be allowed via SuperClass")
	}
}

// ==================== Instance Tests ====================

func TestInstance(t *testing.T) {
	inst := &Instance{Class: &Class{Name: "A"}}
	if inst.Type() != INSTANCE_OBJ || inst.Inspect() != "<A instance>" {
		t.Errorf("Instance wrong")
	}
}

func TestInstanceGetAttr(t *testing.T) {
	cls := &Class{Name: "A", Methods: map[string]Object{"foo": &Builtin{Name: "foo"}}}
	inst := &Instance{Class: cls, Fields: map[string]Object{"bar": &Integer{Value: 42}}}
	if val, ok := inst.GetAttr("bar"); !ok || val.(*Integer).Value != 42 {
		t.Errorf("GetAttr(bar) wrong")
	}
	if _, ok := inst.GetAttr("foo"); !ok {
		t.Error("GetAttr(foo) should find method")
	}
	if _, ok := inst.GetAttr("missing"); ok {
		t.Error("GetAttr(missing) should return false")
	}
}

func TestInstanceSetAttr(t *testing.T) {
	inst := &Instance{Class: &Class{Name: "A"}, Fields: map[string]Object{}}
	inst.SetAttr("x", &Integer{Value: 42})
	if val, _ := inst.GetAttr("x"); val.(*Integer).Value != 42 {
		t.Errorf("SetAttr/GetAttr failed")
	}
	inst2 := &Instance{Class: &Class{Name: "B"}}
	inst2.SetAttr("y", &Integer{Value: 1})
	if inst2.Fields["y"].(*Integer).Value != 1 {
		t.Error("SetAttr should init Fields map")
	}
}

func TestInstanceSlotValues(t *testing.T) {
	cls := &Class{Name: "A", Slots: []string{"x", "y"}}
	inst := &Instance{Class: cls, SlotValues: make([]Object, 2)}
	inst.SetAttr("x", &Integer{Value: 10})
	if val, ok := inst.GetAttr("x"); !ok || val.(*Integer).Value != 10 {
		t.Errorf("Slot SetAttr/GetAttr failed")
	}
	if !inst.IsSlottedInstance() {
		t.Error("should be slotted")
	}
}

func TestInstanceGetSlotIndex(t *testing.T) {
	cls := &Class{Name: "A", Slots: []string{"x", "y"}}
	inst := &Instance{Class: cls}
	if inst.GetSlotIndex("x") != 0 || inst.GetSlotIndex("y") != 1 || inst.GetSlotIndex("z") != -1 {
		t.Errorf("GetSlotIndex wrong")
	}
	if (&Instance{Class: nil}).GetSlotIndex("x") != -1 {
		t.Error("nil class should return -1")
	}
}

func TestInstanceGetAttrSlotNilValue(t *testing.T) {
	cls := &Class{Name: "A", Slots: []string{"x"}}
	inst := &Instance{Class: cls, SlotValues: make([]Object, 1)}
	if _, ok := inst.GetAttr("x"); ok {
		t.Error("nil slot value should return false")
	}
}

func TestInstanceGetAttrWithSuperClass(t *testing.T) {
	base := &Class{Name: "Base", Methods: map[string]Object{"bm": &Builtin{Name: "bm"}}}
	child := &Class{Name: "Child", SuperClass: base, Methods: map[string]Object{}}
	inst := &Instance{Class: child, Fields: map[string]Object{}}
	if _, ok := inst.GetAttr("bm"); !ok {
		t.Error("should find from SuperClass")
	}
}

// ==================== Property Tests ====================

func TestProperty(t *testing.T) {
	p := &Property{}
	if p.Type() != PROPERTY_OBJ || p.Inspect() != "<property object>" {
		t.Errorf("Property wrong")
	}
	if p.IsDataDesc() {
		t.Error("Property without fset should not be data desc")
	}
	if p2 := (&Property{Fset: &Builtin{Name: "s"}}); !p2.IsDataDesc() {
		t.Error("Property with fset should be data desc")
	}
	if p3 := (&Property{Fset: None_}); p3.IsDataDesc() {
		t.Error("Property with fset=None should not be data desc")
	}
	if _, err := p.DescGet(nil, nil); err != nil {
		t.Error("DescGet should return nil error")
	}
	if err := p.DescSet(nil, nil); err != nil {
		t.Error("DescSet should return nil error")
	}
}

// ==================== ClassMethod/StaticMethod/Super/BoundMethod Tests ====================

func TestClassMethod(t *testing.T) {
	cm := &ClassMethod{}
	if cm.Type() != CLASSMETHOD_OBJ || cm.IsDataDesc() {
		t.Errorf("ClassMethod wrong")
	}
}

func TestStaticMethod(t *testing.T) {
	sm := &StaticMethod{}
	if sm.Type() != STATICMETHOD_OBJ || sm.IsDataDesc() {
		t.Errorf("StaticMethod wrong")
	}
}

func TestSuper(t *testing.T) {
	s := &Super{}
	if s.Type() != SUPER_OBJ || s.Inspect() != "<super object>" {
		t.Errorf("Super wrong")
	}
}

func TestBoundMethod(t *testing.T) {
	bm := &BoundMethod{}
	if bm.Type() != FUNCTION_OBJ || bm.Inspect() != "<bound method>" {
		t.Errorf("BoundMethod wrong")
	}
}

// ==================== Module Tests ====================

func TestModule(t *testing.T) {
	m := &Module{Name: "test", Fields: map[string]Object{"foo": &Integer{Value: 42}}}
	if m.Type() != MODULE_OBJ || m.Inspect() != "<module 'test'>" {
		t.Errorf("Module wrong")
	}
	if val, ok := m.GetAttr("foo"); !ok || val.(*Integer).Value != 42 {
		t.Errorf("GetAttr wrong")
	}
	if _, ok := m.GetAttr("missing"); ok {
		t.Error("GetAttr missing should return false")
	}
}

// ==================== Range Tests ====================

func TestRange(t *testing.T) {
	r := NewRange(0, 10, 1)
	if r.Type() != RANGE_OBJ {
		t.Errorf("expected RANGE_OBJ")
	}
}

func TestRangeInspect(t *testing.T) {
	tests := []struct {
		s, e, d int64
		exp     string
	}{
		{0, 10, 1, "range(10)"}, {2, 10, 1, "range(2, 10)"}, {2, 10, 3, "range(2, 10, 3)"},
	}
	for _, tt := range tests {
		if r := NewRange(tt.s, tt.e, tt.d); r.Inspect() != tt.exp {
			t.Errorf("Range inspect = %s, want %s", r.Inspect(), tt.exp)
		}
	}
}

func TestRangeLen(t *testing.T) {
	tests := []struct {
		s, e, d, exp int64
	}{
		{0, 10, 1, 10}, {0, 10, 3, 4}, {5, 0, -1, 5}, {10, 5, -2, 3},
		{0, 0, 1, 0}, {5, 5, 1, 0}, {0, 0, 0, 0},
	}
	for _, tt := range tests {
		if r := NewRange(tt.s, tt.e, tt.d); r.Len() != tt.exp {
			t.Errorf("Range Len = %d, want %d", r.Len(), tt.exp)
		}
	}
}

func TestRangeToList(t *testing.T) {
	if len(NewRange(0, 5, 1).ToList()) != 5 {
		t.Errorf("ToList wrong")
	}
	if len(NewRange(5, 0, -1).ToList()) != 5 {
		t.Errorf("ToList negative step wrong")
	}
}

func TestRangeGetItem(t *testing.T) {
	r := NewRange(0, 10, 2)
	if item, ok := r.GetItem(0); !ok || item.(*Integer).Value != 0 {
		t.Errorf("GetItem(0) wrong")
	}
	if item, ok := r.GetItem(-1); !ok || item.(*Integer).Value != 8 {
		t.Errorf("GetItem(-1) wrong")
	}
	if _, ok := r.GetItem(10); ok {
		t.Error("GetItem(10) should fail")
	}
}

// ==================== Zip Tests ====================

func TestZip(t *testing.T) {
	z := NewZip(nil)
	if z.Type() != ZIP_OBJ || z.Inspect() != "<zip object>" {
		t.Errorf("Zip wrong")
	}
}

func TestZipLen(t *testing.T) {
	l1 := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	l2 := NewList([]Object{&Integer{Value: 3}, &Integer{Value: 4}, &Integer{Value: 5}})
	if NewZip([]Object{l1, l2}).Len() != 2 {
		t.Errorf("Zip Len wrong")
	}
	if NewZip(nil).Len() != 0 {
		t.Errorf("Empty Zip Len wrong")
	}
}

func TestZipToList(t *testing.T) {
	l1 := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	l2 := NewList([]Object{&String{Value: "a"}, &String{Value: "b"}})
	list := NewZip([]Object{l1, l2}).ToList()
	if len(list) != 2 {
		t.Fatalf("ToList length wrong")
	}
}

// ==================== Enum Tests ====================

func TestEnum(t *testing.T) {
	e := NewEnum("Color", map[string]Object{"RED": &Integer{Value: 1}})
	if e.Type() != ENUM_OBJ || e.Inspect() != "<enum Color>" {
		t.Errorf("Enum wrong")
	}
	if member, ok := e.GetAttr("RED"); !ok || member.(*EnumMember).Name != "RED" {
		t.Errorf("Enum GetAttr wrong")
	}
	if _, ok := e.GetAttr("BLUE"); ok {
		t.Error("GetAttr(BLUE) should return false")
	}
	em := e.Members["RED"]
	if em.Type() != ENUM_MEMBER_OBJ || em.Inspect() != "<Color.RED: 1>" {
		t.Errorf("EnumMember wrong")
	}
}

// ==================== RegexPattern Tests ====================

func TestRegexPattern(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile("a"), Pattern: "a"}
	if p.Type() != REGEX_PATTERN_OBJ || p.Inspect() != `re.compile("a")` {
		t.Errorf("RegexPattern wrong")
	}
}

func TestRegexPatternGetAttr(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`(\d+)`), Pattern: `(\d+)`, Flags: 0}
	if attr, _ := p.GetAttr("pattern"); attr.(*String).Value != `(\d+)` {
		t.Errorf("pattern attr wrong")
	}
	if attr, _ := p.GetAttr("flags"); attr.(*Integer).Value != 0 {
		t.Errorf("flags attr wrong")
	}
	// search
	attr, _ := p.GetAttr("search")
	if attr.(*Builtin).Fn(&String{Value: "abc123"}).Type() != REGEX_MATCH_OBJ {
		t.Errorf("search result wrong type")
	}
	if attr.(*Builtin).Fn(&String{Value: "abc"}) != None_ {
		t.Errorf("search no match should return None")
	}
	// match
	attr2, _ := p.GetAttr("match")
	if attr2.(*Builtin).Fn(&String{Value: "123abc"}).Type() != REGEX_MATCH_OBJ {
		t.Errorf("match result wrong type")
	}
	if attr2.(*Builtin).Fn(&String{Value: "abc123"}) != None_ {
		t.Errorf("match not at start should return None")
	}
	// fullmatch
	attr3, _ := p.GetAttr("fullmatch")
	if attr3.(*Builtin).Fn(&String{Value: "123"}).Type() != REGEX_MATCH_OBJ {
		t.Errorf("fullmatch result wrong type")
	}
	if attr3.(*Builtin).Fn(&String{Value: "123abc"}) != None_ {
		t.Errorf("fullmatch partial should return None")
	}
	// findall
	attr4, _ := p.GetAttr("findall")
	if len(attr4.(*Builtin).Fn(&String{Value: "a1b23c"}).(*List).Elements) != 2 {
		t.Errorf("findall wrong")
	}
	// split (with capturing group, result includes captured text)
	attr5, _ := p.GetAttr("split")
	if len(attr5.(*Builtin).Fn(&String{Value: "a1b23c"}).(*List).Elements) != 5 {
		t.Errorf("split wrong")
	}
	// sub
	attr6, _ := p.GetAttr("sub")
	if attr6.(*Builtin).Fn(&String{Value: "X"}, &String{Value: "a1b23c"}).(*String).Value != "aXbXc" {
		t.Errorf("sub wrong")
	}
	// finditer
	attr7, _ := p.GetAttr("finditer")
	if len(attr7.(*Builtin).Fn(&String{Value: "a1b23c"}).(*List).Elements) != 2 {
		t.Errorf("finditer wrong")
	}
	_, ok := p.GetAttr("nonexistent")
	if ok {
		t.Error("GetAttr for unknown should return false")
	}
}

// ==================== RegexMatch Tests ====================

func TestRegexMatch(t *testing.T) {
	m := &RegexMatch{}
	if m.Type() != REGEX_MATCH_OBJ || m.Inspect() != "<re.Match object>" {
		t.Errorf("RegexMatch wrong")
	}
}

func TestRegexMatchGetAttr(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`(\w+) (\d+)`), Pattern: `(\w+) (\d+)`}
	m := &RegexMatch{
		Groups: []string{"hello 42", "hello", "42"}, GroupStarts: []int{0, 0, 6},
		GroupEnds: []int{8, 5, 8}, OriginalString: "hello 42", Pattern: p,
	}
	if attr, _ := m.GetAttr("string"); attr.(*String).Value != "hello 42" {
		t.Errorf("string attr wrong")
	}
	if attr, _ := m.GetAttr("re"); attr != p {
		t.Errorf("re attr wrong")
	}
	if attr, _ := m.GetAttr("lastindex"); attr.(*Integer).Value != 2 {
		t.Errorf("lastindex wrong")
	}
	attr, _ := m.GetAttr("group")
	if attr.(*Builtin).Fn().(*String).Value != "hello 42" {
		t.Errorf("group(0) wrong")
	}
	if attr.(*Builtin).Fn(&Integer{Value: 1}).(*String).Value != "hello" {
		t.Errorf("group(1) wrong")
	}
	attr2, _ := m.GetAttr("start")
	if attr2.(*Builtin).Fn().(*Integer).Value != 0 {
		t.Errorf("start(0) wrong")
	}
	attr3, _ := m.GetAttr("end")
	if attr3.(*Builtin).Fn().(*Integer).Value != 8 {
		t.Errorf("end(0) wrong")
	}
	attr4, _ := m.GetAttr("span")
	if len(attr4.(*Builtin).Fn().(*Tuple).Elements) != 2 {
		t.Errorf("span wrong")
	}
	attr5, _ := m.GetAttr("groups")
	if len(attr5.(*Builtin).Fn().(*Tuple).Elements) != 2 {
		t.Errorf("groups wrong")
	}
	_, ok := m.GetAttr("nonexistent")
	if ok {
		t.Error("GetAttr for unknown should return false")
	}
}

// ==================== StringBuilder Tests ====================

func TestStringBuilder(t *testing.T) {
	sb := NewStringBuilder()
	if sb.Type() != STRING_BUILDER_OBJ || sb.Inspect() != "<string_builder>" {
		t.Errorf("StringBuilder wrong")
	}
	attr, _ := sb.GetAttr("append")
	attr.(*Builtin).Fn(&String{Value: "hello"})
	attr2, _ := sb.GetAttr("build")
	if attr2.(*Builtin).Fn().(*String).Value != "hello" {
		t.Errorf("build wrong")
	}
	attr3, _ := sb.GetAttr("__len__")
	if attr3.(*Builtin).Fn().(*Integer).Value != 5 {
		t.Errorf("__len__ wrong")
	}
	_, ok := sb.GetAttr("nonexistent")
	if ok {
		t.Error("GetAttr for unknown should return false")
	}
}

// ==================== Equal Tests ====================

func TestEqual(t *testing.T) {
	tests := []struct {
		a, b Object
		exp  bool
	}{
		{&Integer{Value: 42}, &Integer{Value: 42}, true},
		{&Integer{Value: 1}, &Integer{Value: 2}, false},
		{&Float{Value: 3.14}, &Float{Value: 3.14}, true},
		{&String{Value: "hi"}, &String{Value: "hi"}, true},
		{&Boolean{Value: true}, &Boolean{Value: true}, true},
		{&None{}, &None{}, true},
		{&Ellipsis{}, &Ellipsis{}, true},
		{NewBytes([]byte("a")), NewBytes([]byte("a")), true},
		{NewBytes([]byte("a")), NewBytes([]byte("b")), false},
		{&Complex{Real: 1, Imag: 2}, &Complex{Real: 1, Imag: 2}, true},
		{&Integer{Value: 1}, &Float{Value: 1}, false},
		{&List{}, &List{}, false},
	}
	for _, tt := range tests {
		if Equal(tt.a, tt.b) != tt.exp {
			t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, !tt.exp, tt.exp)
		}
	}
}

// ==================== IsInstanceOf Tests ====================

func TestIsInstanceOf(t *testing.T) {
	intClass := &Class{Name: "int"}
	objClass := &Class{Name: "object"}
	tests := []struct {
		obj Object
		cls *Class
		exp bool
	}{
		{&Integer{Value: 1}, intClass, true},
		{&Integer{Value: 1}, objClass, true},
		{&Integer{Value: 1}, &Class{Name: "str"}, false},
		{&String{Value: "hi"}, &Class{Name: "str"}, true},
		{&Float{Value: 1.0}, &Class{Name: "float"}, true},
		{&Boolean{Value: true}, &Class{Name: "bool"}, true},
		{&Boolean{Value: true}, intClass, true},
		{&List{}, &Class{Name: "list"}, true},
		{&Dict{}, &Class{Name: "dict"}, true},
		{&Tuple{}, &Class{Name: "tuple"}, true},
		{&Set{}, &Class{Name: "set"}, true},
		{&None{}, &Class{Name: "NoneType"}, true},
		{&Builtin{}, &Class{Name: "function"}, true},
		{&Closure{}, &Class{Name: "function"}, true},
		{&Class{}, &Class{Name: "type"}, true},
	}
	for _, tt := range tests {
		if IsInstanceOf(tt.obj, tt.cls) != tt.exp {
			t.Errorf("IsInstanceOf(%T, %s) wrong", tt.obj, tt.cls.Name)
		}
	}
}

// ==================== IsSubclassOf Tests ====================

func TestIsSubclassOf(t *testing.T) {
	base := &Class{Name: "Base"}
	child := &Class{Name: "Child", SuperClass: base}
	if !IsSubclassOf(child, base) || IsSubclassOf(base, child) || !IsSubclassOf(base, base) {
		t.Errorf("IsSubclassOf wrong")
	}
	b2 := &Class{Name: "Base2"}
	c2 := &Class{Name: "Child2", SuperClasses: []*Class{base, b2}}
	if !IsSubclassOf(c2, base) || !IsSubclassOf(c2, b2) {
		t.Errorf("IsSubclassOf with SuperClasses wrong")
	}
}

// ==================== IsCallable Tests ====================

func TestIsCallable(t *testing.T) {
	if !IsCallable(&Builtin{}) || !IsCallable(&Closure{}) || !IsCallable(&Class{}) {
		t.Error("Builtin/Closure/Class should be callable")
	}
	if IsCallable(&Integer{Value: 1}) || IsCallable(&String{Value: "hi"}) {
		t.Error("Integer/String should not be callable")
	}
	cls := &Class{Name: "A", Methods: map[string]Object{"__call__": &Builtin{Name: "__call__"}}}
	if !IsCallable(&Instance{Class: cls}) {
		t.Error("Instance with __call__ method should be callable")
	}
	inst2 := &Instance{Class: &Class{Name: "B"}, Fields: map[string]Object{"__call__": &Builtin{Name: "__call__"}}}
	if !IsCallable(inst2) {
		t.Error("Instance with __call__ field should be callable")
	}
	if IsCallable(&Instance{Class: &Class{Name: "C"}}) {
		t.Error("Instance without __call__ should not be callable")
	}
}

// ==================== Descriptor Tests ====================

func TestIsDescriptor(t *testing.T) {
	if !IsDescriptor(&Property{}) || IsDescriptor(&Integer{Value: 1}) {
		t.Errorf("IsDescriptor wrong")
	}
	cls := &Class{Name: "A", Methods: map[string]Object{"__get__": &Builtin{Name: "__get__"}}}
	if !IsDescriptor(&Instance{Class: cls}) {
		t.Error("Instance with __get__ should be descriptor")
	}
}

func TestIsDataDescriptor(t *testing.T) {
	if !IsDataDescriptor(&Property{Fset: &Builtin{Name: "s"}}) {
		t.Error("Property with fset should be data desc")
	}
	cls := &Class{Name: "A", Methods: map[string]Object{"__set__": &Builtin{Name: "__set__"}}}
	if !IsDataDescriptor(&Instance{Class: cls}) {
		t.Error("Instance with __set__ should be data desc")
	}
	if IsDataDescriptor(&Integer{Value: 1}) {
		t.Error("Integer should not be data desc")
	}
}

// ==================== CheckHashable Tests ====================

func TestCheckHashable(t *testing.T) {
	if CheckHashable(&Integer{Value: 1}) != nil {
		t.Error("Integer should be hashable")
	}
	if CheckHashable(&List{}) == nil {
		t.Error("List should not be hashable")
	}
	if CheckHashable(&Tuple{Elements: []Object{&List{}}}) == nil {
		t.Error("Tuple with List should not be hashable")
	}
}

// ==================== FormatString Tests ====================

func TestFormatString(t *testing.T) {
	if r := FormatString("hello %s", &String{Value: "world"}); r != "hello world" {
		t.Errorf("FormatString %%s wrong: %s", r)
	}
	if r := FormatString("num: %d", &Integer{Value: 42}); r != "num: 42" {
		t.Errorf("FormatString %%d wrong: %s", r)
	}
	if r := FormatString("no format"); r != "no format" {
		t.Errorf("FormatString no format wrong")
	}
}

// ==================== Module Registration Tests ====================

func TestRegisterAndGetModule(t *testing.T) {
	RegisterModule("testmod", &Module{Name: "testmod"})
	if GetModule("testmod") == nil || GetModule("testmod").Name != "testmod" {
		t.Error("GetModule wrong")
	}
	if GetModule("nonexistent") != nil {
		t.Error("GetModule for nonexistent should return nil")
	}
}

// ==================== CreateMathModule Tests ====================

func TestCreateMathModule(t *testing.T) {
	m := CreateMathModule()
	if m.Name != "math" {
		t.Errorf("module name wrong")
	}
	if pi, ok := m.Fields["pi"]; !ok || pi.(*Float).Value != math.Pi {
		t.Error("math.pi wrong")
	}
	sinFn := m.Fields["sin"].(*Builtin)
	if sinFn.Fn(&Float{Value: 0}).(*Float).Value != 0 {
		t.Errorf("sin(0) wrong")
	}
	if sinFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("sin(string) should return error")
	}
	cosFn := m.Fields["cos"].(*Builtin)
	if cosFn.Fn(&Float{Value: 0}).(*Float).Value != 1.0 {
		t.Errorf("cos(0) wrong")
	}
	sqrtFn := m.Fields["sqrt"].(*Builtin)
	if sqrtFn.Fn(&Float{Value: 4}).(*Float).Value != 2.0 {
		t.Errorf("sqrt(4) wrong")
	}
	floorFn := m.Fields["floor"].(*Builtin)
	if floorFn.Fn(&Float{Value: 3.7}).(*Integer).Value != 3 {
		t.Errorf("floor(3.7) wrong")
	}
	absFn := m.Fields["abs"].(*Builtin)
	if absFn.Fn(&Integer{Value: -5}).(*Integer).Value != 5 {
		t.Errorf("abs(-5) wrong")
	}
	if math.Abs(absFn.Fn(&Complex{Real: 3, Imag: 4}).(*Float).Value-5.0) > 1e-10 {
		t.Errorf("abs(3+4j) wrong")
	}
}

// ==================== CreateSysModule Tests ====================

func TestCreateSysModule(t *testing.T) {
	m := CreateSysModule()
	if m.Name != "sys" {
		t.Errorf("module name wrong")
	}
	if _, ok := m.Fields["version"]; !ok {
		t.Error("sys.version should exist")
	}
	getsizeofFn := m.Fields["getsizeof"].(*Builtin)
	if getsizeofFn.Fn(&Integer{Value: 1}).(*Integer).Value != 28 {
		t.Errorf("getsizeof(int) wrong")
	}
}

// ==================== CreateOsModule Tests ====================

func TestCreateOsModule(t *testing.T) {
	m := CreateOsModule()
	if m.Name != "os" {
		t.Errorf("module name wrong")
	}
	if _, ok := m.Fields["sep"]; !ok {
		t.Error("os.sep should exist")
	}
	getenvFn := m.Fields["getenv"].(*Builtin)
	if getenvFn.Fn(&String{Value: "NONEXISTENT_XYZ"}, &String{Value: "default"}).(*String).Value != "default" {
		t.Errorf("getenv with default wrong")
	}
}

// ==================== CreateRandomModule Tests ====================

func TestCreateRandomModule(t *testing.T) {
	m := CreateRandomModule()
	if m.Name != "random" {
		t.Errorf("module name wrong")
	}
	seedFn := m.Fields["seed"].(*Builtin)
	if seedFn.Fn(&Integer{Value: 42}) != None_ {
		t.Errorf("seed should return None")
	}
	randomFn := m.Fields["random"].(*Builtin)
	if v := randomFn.Fn().(*Float).Value; v < 0 || v >= 1 {
		t.Errorf("random() out of range")
	}
}

// ==================== CreateStringModule Tests ====================

func TestCreateStringModule(t *testing.T) {
	m := CreateStringModule()
	if m.Name != "string" {
		t.Errorf("module name wrong")
	}
	for _, name := range []string{"ascii_letters", "digits", "whitespace"} {
		if _, ok := m.Fields[name]; !ok {
			t.Errorf("string.%s should exist", name)
		}
	}
}

// ==================== CreateJsonModule Tests ====================

func TestCreateJsonModule(t *testing.T) {
	m := CreateJsonModule()
	if m.Name != "json" {
		t.Errorf("module name wrong")
	}
	dumpsFn := m.Fields["dumps"].(*Builtin)
	if dumpsFn.Fn(&Integer{Value: 42}).(*String).Value != "42" {
		t.Errorf("dumps(42) wrong")
	}
	if dumpsFn.Fn(None_).(*String).Value != "null" {
		t.Errorf("dumps(None) wrong")
	}
	loadsFn := m.Fields["loads"].(*Builtin)
	if loadsFn.Fn(&String{Value: "42"}).(*Float).Value != 42.0 {
		t.Errorf("loads('42') wrong")
	}
	if loadsFn.Fn(&String{Value: "null"}) != None_ {
		t.Errorf("loads('null') wrong")
	}
	if loadsFn.Fn(&String{Value: "true"}) != True {
		t.Errorf("loads('true') wrong")
	}
}

// ==================== CreateTimeModule Tests ====================

func TestCreateTimeModule(t *testing.T) {
	m := CreateTimeModule()
	if m.Name != "time" {
		t.Errorf("module name wrong")
	}
	timeFn := m.Fields["time"].(*Builtin)
	if timeFn.Fn().(*Float).Value <= 0 {
		t.Errorf("time() should be positive")
	}
}

// ==================== CreateDatetimeModule Tests ====================

func TestCreateDatetimeModule(t *testing.T) {
	m := CreateDatetimeModule()
	if m.Name != "datetime" {
		t.Errorf("module name wrong")
	}
	dtMod := m.Fields["datetime"].(*Module)
	if dtMod.Fields["now"].(*Builtin).Fn().(*String).Value == "" {
		t.Error("datetime.now() should return non-empty")
	}
}

// ==================== CompareObjectsForSort Tests ====================

func TestCompareObjectsForSort(t *testing.T) {
	tests := []struct {
		a, b Object
		exp  int
	}{
		{&Integer{Value: 1}, &Integer{Value: 2}, -1},
		{&Integer{Value: 2}, &Integer{Value: 1}, 1},
		{&Integer{Value: 1}, &Integer{Value: 1}, 0},
		{&Float{Value: 1.0}, &Float{Value: 2.0}, -1},
		{&String{Value: "a"}, &String{Value: "b"}, -1},
		{&Boolean{Value: false}, &Boolean{Value: true}, -1},
	}
	for _, tt := range tests {
		if r := compareObjectsForSort(tt.a, tt.b); r != tt.exp {
			t.Errorf("compareObjectsForSort wrong: got %d, want %d", r, tt.exp)
		}
	}
}

// ==================== Pattern Helper Tests ====================

func TestPatternSearch(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	if PatternSearch(p, "abc123").Type() != REGEX_MATCH_OBJ {
		t.Errorf("PatternSearch wrong")
	}
	if PatternSearch(p, "abc") != None_ {
		t.Errorf("PatternSearch no match should return None")
	}
}

func TestPatternMatch(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	if PatternMatch(p, "123abc").Type() != REGEX_MATCH_OBJ {
		t.Errorf("PatternMatch wrong")
	}
	if PatternMatch(p, "abc123") != None_ {
		t.Errorf("PatternMatch not at start should return None")
	}
}

func TestPatternFullmatch(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	if PatternFullmatch(p, "123").Type() != REGEX_MATCH_OBJ {
		t.Errorf("PatternFullmatch wrong")
	}
	if PatternFullmatch(p, "123abc") != None_ {
		t.Errorf("PatternFullmatch partial should return None")
	}
}

func TestPatternFindall(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	if len(PatternFindall(p, "a1b23c").(*List).Elements) != 2 {
		t.Errorf("findall wrong")
	}
}

func TestPatternSplit(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	if len(PatternSplit(p, "a1b23c").(*List).Elements) != 3 {
		t.Errorf("split wrong")
	}
}

func TestCallFunction(t *testing.T) {
	if CallFunction(&Builtin{Name: "test"}).Type() != ERROR_OBJ {
		t.Error("CallFunction without callback should return error")
	}
}

// ==================== Additional Error Case Tests ====================

func TestStringGetAttr_Errors(t *testing.T) {
	s := &String{Value: "hello"}
	for name, args := range map[string][]Object{
		"rfind": {&Integer{Value: 1}}, "rindex": {&Integer{Value: 1}},
		"count": {&Integer{Value: 1}}, "center": {}, "ljust": {},
		"rjust": {}, "zfill": {}, "partition": {}, "rpartition": {},
		"translate": {}, "format_map": {},
	} {
		attr, _ := s.GetAttr(name)
		if attr.(*Builtin).Fn(args...).Type() != ERROR_OBJ {
			t.Errorf("%s with wrong args should return error", name)
		}
	}
}

func TestRegexPatternGetAttr_Errors(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	for name, args := range map[string][]Object{
		"search": {&Integer{Value: 1}}, "match": {}, "fullmatch": {&Integer{Value: 1}},
		"findall": {}, "finditer": {&Integer{Value: 1}}, "split": {&Integer{Value: 1}},
	} {
		attr, _ := p.GetAttr(name)
		if attr.(*Builtin).Fn(args...).Type() != ERROR_OBJ {
			t.Errorf("%s with wrong args should return error", name)
		}
	}
}

func TestRegexMatchGetAttr_Errors(t *testing.T) {
	m := &RegexMatch{Groups: []string{"hello"}, GroupStarts: []int{0}, GroupEnds: []int{5}}
	groupAttr, _ := m.GetAttr("group")
	if groupAttr.(*Builtin).Fn(&Integer{Value: 99}).Type() != ERROR_OBJ {
		t.Error("group(99) should return error")
	}
	if groupAttr.(*Builtin).Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("group('a') should return error")
	}
	m2 := &RegexMatch{Groups: []string{"hello"}}
	if attr, _ := m2.GetAttr("lastindex"); attr != None_ {
		t.Error("lastindex with no groups should be None")
	}
}

func TestListGetAttr_Errors(t *testing.T) {
	l := NewList([]Object{&Integer{Value: 1}})
	attr, _ := l.GetAttr("__imul__")
	if attr.(*Builtin).Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("__imul__(string) should return error")
	}
	attr2, _ := l.GetAttr("__iadd__")
	if attr2.(*Builtin).Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("__iadd__(int) should return error")
	}
}

// ==================== Additional Coverage Tests ====================

// Set.Inspect
func TestSetInspectDetailed(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	if s.Inspect() != "{1}" {
		t.Errorf("single element set inspect = %s, want {1}", s.Inspect())
	}
	s2 := NewSet()
	if s2.Inspect() != "{}" {
		t.Errorf("empty set inspect = %s, want {}", s2.Inspect())
	}
}

// Dict.Inspect
func TestDictInspectDetailed(t *testing.T) {
	d := NewDict()
	if d.Inspect() != "{}" {
		t.Errorf("empty dict inspect = %s, want {}", d.Inspect())
	}
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	d.Set(&String{Value: "b"}, &Integer{Value: 2})
	insp := d.Inspect()
	if len(insp) < 5 {
		t.Errorf("dict inspect too short: %s", insp)
	}
}

// Dict HashKey with Complex - falls through to default
func TestDictHashKeyComplex(t *testing.T) {
	d := NewDict()
	got := d.HashKey(&Complex{Real: 1, Imag: 2})
	if got == "" {
		t.Error("HashKey should return non-empty string")
	}
}

// DictKeys ToList
func TestDictKeysToListDetailed(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	dk := NewDictKeys(d)
	list := dk.ToList()
	if len(list) != 1 {
		t.Errorf("ToList length = %d, want 1", len(list))
	}
	if list[0].Inspect() != "a" {
		t.Errorf("ToList[0] = %s, want a", list[0].Inspect())
	}
}

// DictValues ToList
func TestDictValuesToListDetailed(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	dv := NewDictValues(d)
	list := dv.ToList()
	if len(list) != 1 {
		t.Errorf("ToList length = %d, want 1", len(list))
	}
	if list[0].(*Integer).Value != 1 {
		t.Errorf("ToList[0] wrong")
	}
}

// DictItems ToList
func TestDictItemsToListDetailed(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	di := NewDictItems(d)
	list := di.ToList()
	if len(list) != 1 {
		t.Errorf("ToList length = %d, want 1", len(list))
	}
	tup := list[0].(*Tuple)
	if len(tup.Elements) != 2 {
		t.Errorf("item tuple length = %d, want 2", len(tup.Elements))
	}
}

// DictValues Inspect
func TestDictValuesInspectDetailed(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	dv := NewDictValues(d)
	if dv.Inspect() != "dict_values([1])" {
		t.Errorf("DictValues inspect = %s, want dict_values([1])", dv.Inspect())
	}
}

// DictItems Inspect
func TestDictItemsInspectDetailed(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	di := NewDictItems(d)
	insp := di.Inspect()
	if insp == "" {
		t.Error("DictItems inspect should not be empty")
	}
}

// ClassMethod Inspect/DescGet/DescSet
func TestClassMethodDetailed(t *testing.T) {
	cm := &ClassMethod{Fn: &Builtin{Name: "cm"}}
	if cm.Inspect() != "<classmethod object>" {
		t.Errorf("ClassMethod inspect = %s, want <classmethod object>", cm.Inspect())
	}
	result, err := cm.DescGet(nil, nil)
	if err != nil || result != nil {
		t.Errorf("DescGet should return nil, nil")
	}
	if err := cm.DescSet(nil, nil); err != nil {
		t.Errorf("DescSet should return nil")
	}
}

// StaticMethod Inspect/DescGet/DescSet
func TestStaticMethodDetailed(t *testing.T) {
	sm := &StaticMethod{Fn: &Builtin{Name: "sm"}}
	if sm.Inspect() != "<staticmethod object>" {
		t.Errorf("StaticMethod inspect = %s, want <staticmethod object>", sm.Inspect())
	}
	result, err := sm.DescGet(nil, nil)
	if err != nil || result != nil {
		t.Errorf("DescGet should return nil, nil")
	}
	if err := sm.DescSet(nil, nil); err != nil {
		t.Errorf("DescSet should return nil")
	}
}

// ComputeMRO with SuperClasses
func TestClassComputeMROWithSuperClasses(t *testing.T) {
	base1 := &Class{Name: "Base1"}
	base2 := &Class{Name: "Base2"}
	child := &Class{Name: "Child", SuperClasses: []*Class{base1, base2}}
	mro := child.ComputeMRO()
	if len(mro) != 3 {
		t.Errorf("MRO with SuperClasses length = %d, want 3", len(mro))
	}
}

// Zip Len with Range
func TestZipLenWithRangeDetailed(t *testing.T) {
	r := NewRange(0, 3, 1)
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	z := NewZip([]Object{r, l})
	if z.Len() != 2 {
		t.Errorf("Zip with Range Len() = %d, want 2", z.Len())
	}
}

// Zip Len with unsupported
func TestZipLenWithUnsupported(t *testing.T) {
	z := NewZip([]Object{&Integer{Value: 1}})
	if z.Len() != -1 {
		t.Errorf("Zip with unsupported Len() = %d, want -1", z.Len())
	}
}

// Zip ToList with Range and String
func TestZipToListWithRangeAndString(t *testing.T) {
	r := NewRange(0, 2, 1)
	l := NewList([]Object{&String{Value: "a"}, &String{Value: "b"}})
	z := NewZip([]Object{r, l})
	list := z.ToList()
	if len(list) != 2 {
		t.Fatalf("ToList length = %d, want 2", len(list))
	}
}

// Zip ToList with unsupported
func TestZipToListWithUnsupported(t *testing.T) {
	z := NewZip([]Object{&Integer{Value: 1}})
	if z.ToList() != nil {
		t.Error("ToList with unsupported should return nil")
	}
}

// Math module comprehensive
func TestCreateMathModuleComprehensive(t *testing.T) {
	m := CreateMathModule()

	// tan
	if tanFn := m.Fields["tan"].(*Builtin); tanFn.Fn(&Float{Value: 0}).(*Float).Value != 0 {
		t.Errorf("tan(0) wrong")
	}
	// asin
	if asinFn := m.Fields["asin"].(*Builtin); asinFn.Fn(&Float{Value: 0}).(*Float).Value != 0 {
		t.Errorf("asin(0) wrong")
	}
	// acos
	if acosFn := m.Fields["acos"].(*Builtin); acosFn.Fn(&Float{Value: 1}).(*Float).Value != 0 {
		t.Errorf("acos(1) wrong")
	}
	// atan
	if atanFn := m.Fields["atan"].(*Builtin); atanFn.Fn(&Float{Value: 0}).(*Float).Value != 0 {
		t.Errorf("atan(0) wrong")
	}
	// trunc
	if truncFn := m.Fields["trunc"].(*Builtin); truncFn.Fn(&Float{Value: 3.7}).(*Integer).Value != 3 {
		t.Errorf("trunc(3.7) wrong")
	}
	// pow
	if powFn := m.Fields["pow"].(*Builtin); powFn.Fn(&Float{Value: 2}, &Float{Value: 3}).(*Float).Value != 8 {
		t.Errorf("pow(2,3) wrong")
	}
	// hypot
	if hypotFn := m.Fields["hypot"].(*Builtin); hypotFn.Fn(&Float{Value: 3}, &Float{Value: 4}).(*Float).Value != 5 {
		t.Errorf("hypot(3,4) wrong")
	}
	// log10
	if log10Fn := m.Fields["log10"].(*Builtin); math.Abs(log10Fn.Fn(&Float{Value: 100}).(*Float).Value-2.0) > 1e-10 {
		t.Errorf("log10(100) wrong")
	}
	// log2
	if log2Fn := m.Fields["log2"].(*Builtin); math.Abs(log2Fn.Fn(&Float{Value: 8}).(*Float).Value-3.0) > 1e-10 {
		t.Errorf("log2(8) wrong")
	}
	// exp
	if expFn := m.Fields["exp"].(*Builtin); expFn.Fn(&Float{Value: 0}).(*Float).Value != 1 {
		t.Errorf("exp(0) wrong")
	}
	// degrees
	if degFn := m.Fields["degrees"].(*Builtin); degFn.Fn(&Float{Value: math.Pi}).(*Float).Value != 180 {
		t.Errorf("degrees(pi) wrong")
	}
	// radians
	if radFn := m.Fields["radians"].(*Builtin); math.Abs(radFn.Fn(&Float{Value: 180}).(*Float).Value-math.Pi) > 1e-10 {
		t.Errorf("radians(180) wrong")
	}
	// abs with positive integer
	if absFn := m.Fields["abs"].(*Builtin); absFn.Fn(&Integer{Value: 5}).(*Integer).Value != 5 {
		t.Errorf("abs(5) wrong")
	}
	// abs with float
	if absFn := m.Fields["abs"].(*Builtin); absFn.Fn(&Float{Value: -3.14}).(*Float).Value != 3.14 {
		t.Errorf("abs(-3.14) wrong")
	}
	// floor with integer
	if floorFn := m.Fields["floor"].(*Builtin); floorFn.Fn(&Integer{Value: 5}).(*Integer).Value != 5 {
		t.Errorf("floor(5) should return 5")
	}
	// ceil with integer
	if ceilFn := m.Fields["ceil"].(*Builtin); ceilFn.Fn(&Integer{Value: 5}).(*Integer).Value != 5 {
		t.Errorf("ceil(5) should return 5")
	}
	// trunc with integer
	if truncFn := m.Fields["trunc"].(*Builtin); truncFn.Fn(&Integer{Value: 5}).(*Integer).Value != 5 {
		t.Errorf("trunc(5) should return 5")
	}
	// cos with integer
	if cosFn := m.Fields["cos"].(*Builtin); cosFn.Fn(&Integer{Value: 0}).(*Float).Value != 1.0 {
		t.Errorf("cos(0) with int wrong")
	}
	// sqrt with integer
	if sqrtFn := m.Fields["sqrt"].(*Builtin); sqrtFn.Fn(&Integer{Value: 4}).(*Float).Value != 2.0 {
		t.Errorf("sqrt(4) with int wrong")
	}
	// log with integer
	if logFn := m.Fields["log"].(*Builtin); math.Abs(logFn.Fn(&Integer{Value: 1}).(*Float).Value) > 1e-10 {
		t.Errorf("log(1) with int wrong")
	}
	// log10 with integer
	if log10Fn := m.Fields["log10"].(*Builtin); math.Abs(log10Fn.Fn(&Integer{Value: 100}).(*Float).Value-2.0) > 1e-10 {
		t.Errorf("log10(100) with int wrong")
	}
	// log2 with integer
	if log2Fn := m.Fields["log2"].(*Builtin); math.Abs(log2Fn.Fn(&Integer{Value: 8}).(*Float).Value-3.0) > 1e-10 {
		t.Errorf("log2(8) with int wrong")
	}
	// exp with integer
	if expFn := m.Fields["exp"].(*Builtin); expFn.Fn(&Integer{Value: 0}).(*Float).Value != 1.0 {
		t.Errorf("exp(0) with int wrong")
	}
	// degrees with integer
	if degFn := m.Fields["degrees"].(*Builtin); degFn.Fn(&Integer{Value: 0}).(*Float).Value != 0 {
		t.Errorf("degrees(0) with int wrong")
	}
	// radians with integer
	if radFn := m.Fields["radians"].(*Builtin); radFn.Fn(&Integer{Value: 0}).(*Float).Value != 0 {
		t.Errorf("radians(0) with int wrong")
	}
	// tan with integer
	if tanFn := m.Fields["tan"].(*Builtin); tanFn.Fn(&Integer{Value: 0}).(*Float).Value != 0 {
		t.Errorf("tan(0) with int wrong")
	}
	// asin with integer
	if asinFn := m.Fields["asin"].(*Builtin); asinFn.Fn(&Integer{Value: 0}).(*Float).Value != 0 {
		t.Errorf("asin(0) with int wrong")
	}
	// acos with integer
	if acosFn := m.Fields["acos"].(*Builtin); acosFn.Fn(&Integer{Value: 1}).(*Float).Value != 0 {
		t.Errorf("acos(1) with int wrong")
	}
	// atan with integer
	if atanFn := m.Fields["atan"].(*Builtin); atanFn.Fn(&Integer{Value: 0}).(*Float).Value != 0 {
		t.Errorf("atan(0) with int wrong")
	}
	// Error cases for math
	if powFn := m.Fields["pow"].(*Builtin); powFn.Fn(&Float{Value: 2}).Type() != ERROR_OBJ {
		t.Error("pow(1 arg) should return error")
	}
	if powFn := m.Fields["pow"].(*Builtin); powFn.Fn(&Integer{Value: 2}, &Float{Value: 3}).Type() != ERROR_OBJ {
		t.Error("pow(int, float) should return error")
	}
	if hypotFn := m.Fields["hypot"].(*Builtin); hypotFn.Fn(&Float{Value: 3}).Type() != ERROR_OBJ {
		t.Error("hypot(1 arg) should return error")
	}
	if hypotFn := m.Fields["hypot"].(*Builtin); hypotFn.Fn(&Integer{Value: 3}, &Float{Value: 4}).Type() != ERROR_OBJ {
		t.Error("hypot(int, float) should return error")
	}
	if absFn := m.Fields["abs"].(*Builtin); absFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("abs(string) should return error")
	}
	if floorFn := m.Fields["floor"].(*Builtin); floorFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("floor(string) should return error")
	}
	if ceilFn := m.Fields["ceil"].(*Builtin); ceilFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("ceil(string) should return error")
	}
	if truncFn := m.Fields["trunc"].(*Builtin); truncFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("trunc(string) should return error")
	}
	if cosFn := m.Fields["cos"].(*Builtin); cosFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("cos(string) should return error")
	}
	if tanFn := m.Fields["tan"].(*Builtin); tanFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("tan(string) should return error")
	}
	if asinFn := m.Fields["asin"].(*Builtin); asinFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("asin(string) should return error")
	}
	if acosFn := m.Fields["acos"].(*Builtin); acosFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("acos(string) should return error")
	}
	if atanFn := m.Fields["atan"].(*Builtin); atanFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("atan(string) should return error")
	}
	if sqrtFn := m.Fields["sqrt"].(*Builtin); sqrtFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("sqrt(string) should return error")
	}
	if logFn := m.Fields["log"].(*Builtin); logFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("log(string) should return error")
	}
	if log10Fn := m.Fields["log10"].(*Builtin); log10Fn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("log10(string) should return error")
	}
	if log2Fn := m.Fields["log2"].(*Builtin); log2Fn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("log2(string) should return error")
	}
	if expFn := m.Fields["exp"].(*Builtin); expFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("exp(string) should return error")
	}
	if degFn := m.Fields["degrees"].(*Builtin); degFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("degrees(string) should return error")
	}
	if radFn := m.Fields["radians"].(*Builtin); radFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("radians(string) should return error")
	}
}

// Sys module comprehensive
func TestCreateSysModuleComprehensive(t *testing.T) {
	m := CreateSysModule()
	getsizeofFn := m.Fields["getsizeof"].(*Builtin)
	// Test all type sizes
	if getsizeofFn.Fn(&Float{Value: 1.0}).(*Integer).Value != 24 {
		t.Errorf("getsizeof(float) wrong")
	}
	if getsizeofFn.Fn(&Boolean{Value: true}).(*Integer).Value != 28 {
		t.Errorf("getsizeof(bool) wrong")
	}
	if getsizeofFn.Fn(&String{Value: "hi"}).(*Integer).Value != 51 {
		t.Errorf("getsizeof(str) wrong")
	}
	if getsizeofFn.Fn(NewList([]Object{&Integer{Value: 1}})).(*Integer).Value != 64 {
		t.Errorf("getsizeof(list) wrong")
	}
	if getsizeofFn.Fn(NewDict()).(*Integer).Value != 64 {
		t.Errorf("getsizeof(dict) wrong")
	}
	if getsizeofFn.Fn(NewTuple([]Object{&Integer{Value: 1}})).(*Integer).Value != 48 {
		t.Errorf("getsizeof(tuple) wrong")
	}
	if getsizeofFn.Fn(NewSet()).(*Integer).Value != 216 {
		t.Errorf("getsizeof(set) wrong")
	}
	if getsizeofFn.Fn(NewBytes([]byte{1, 2})).(*Integer).Value != 35 {
		t.Errorf("getsizeof(bytes) wrong")
	}
	if getsizeofFn.Fn(None_).(*Integer).Value != 16 {
		t.Errorf("getsizeof(None) wrong")
	}
	if getsizeofFn.Fn(&Complex{Real: 1, Imag: 2}).(*Integer).Value != 32 {
		t.Errorf("getsizeof(complex) wrong")
	}
	if getsizeofFn.Fn(&Closure{}).(*Integer).Value != 136 {
		t.Errorf("getsizeof(closure) wrong")
	}
	if getsizeofFn.Fn(&Builtin{}).(*Integer).Value != 136 {
		t.Errorf("getsizeof(builtin) wrong")
	}
	if getsizeofFn.Fn(&BoundMethod{}).(*Integer).Value != 136 {
		t.Errorf("getsizeof(boundmethod) wrong")
	}
	if getsizeofFn.Fn(&Class{Name: "A"}).(*Integer).Value != 64 {
		t.Errorf("getsizeof(class) wrong")
	}
	if getsizeofFn.Fn(&Instance{Class: &Class{Name: "A"}, Fields: map[string]Object{"x": &Integer{Value: 1}}}).(*Integer).Value != 64 {
		t.Errorf("getsizeof(instance) wrong")
	}
	if getsizeofFn.Fn(&Instance{Class: &Class{Name: "A"}, SlotValues: make([]Object, 2)}).(*Integer).Value != 72 {
		t.Errorf("getsizeof(slotted instance) wrong")
	}
	// Default case
	if getsizeofFn.Fn(&Generator{}).(*Integer).Value != 24 {
		t.Errorf("getsizeof(generator) wrong")
	}
}

// Os module comprehensive
func TestCreateOsModuleComprehensive(t *testing.T) {
	m := CreateOsModule()
	// getcwd
	getcwdFn := m.Fields["getcwd"].(*Builtin)
	if getcwdFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("getcwd(1) should return error")
	}
	// chdir
	chdirFn := m.Fields["chdir"].(*Builtin)
	if chdirFn.Fn().Type() != ERROR_OBJ {
		t.Error("chdir() should return error")
	}
	if chdirFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("chdir(int) should return error")
	}
	// listdir
	listdirFn := m.Fields["listdir"].(*Builtin)
	if listdirFn.Fn().Type() != ERROR_OBJ {
		t.Error("listdir() should return error")
	}
	if listdirFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("listdir(int) should return error")
	}
	// mkdir
	mkdirFn := m.Fields["mkdir"].(*Builtin)
	if mkdirFn.Fn().Type() != ERROR_OBJ {
		t.Error("mkdir() should return error")
	}
	if mkdirFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("mkdir(int) should return error")
	}
	// remove
	removeFn := m.Fields["remove"].(*Builtin)
	if removeFn.Fn().Type() != ERROR_OBJ {
		t.Error("remove() should return error")
	}
	if removeFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("remove(int) should return error")
	}
	// rename
	renameFn := m.Fields["rename"].(*Builtin)
	if renameFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("rename(1 arg) should return error")
	}
	if renameFn.Fn(&Integer{Value: 1}, &String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("rename(int, str) should return error")
	}
	// getenv
	getenvFn := m.Fields["getenv"].(*Builtin)
	if getenvFn.Fn().Type() != ERROR_OBJ {
		t.Error("getenv() should return error")
	}
	if getenvFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("getenv(int) should return error")
	}
}

// Random module comprehensive
func TestCreateRandomModuleComprehensive(t *testing.T) {
	m := CreateRandomModule()
	// seed with no args
	seedFn := m.Fields["seed"].(*Builtin)
	if seedFn.Fn() != None_ {
		t.Errorf("seed() should return None")
	}
	// random with args (may be ignored)
	randomFn := m.Fields["random"].(*Builtin)
	result2 := randomFn.Fn(&Integer{Value: 1})
	_ = result2 // just ensure no panic
	// randint with wrong args
	randintFn := m.Fields["randint"].(*Builtin)
	if randintFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("randint(1 arg) should return error")
	}
	if randintFn.Fn(&String{Value: "a"}, &Integer{Value: 10}).Type() != ERROR_OBJ {
		t.Error("randint(str, int) should return error")
	}
	// choice
	choiceFn := m.Fields["choice"].(*Builtin)
	if choiceFn.Fn().Type() != ERROR_OBJ {
		t.Error("choice() should return error")
	}
	if choiceFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("choice(int) should return error")
	}
	// shuffle
	shuffleFn := m.Fields["shuffle"].(*Builtin)
	if shuffleFn.Fn().Type() != ERROR_OBJ {
		t.Error("shuffle() should return error")
	}
	if shuffleFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("shuffle(int) should return error")
	}
	// uniform
	uniformFn := m.Fields["uniform"].(*Builtin)
	if uniformFn.Fn(&Float{Value: 0}).Type() != ERROR_OBJ {
		t.Error("uniform(1 arg) should return error")
	}
	if uniformFn.Fn(&String{Value: "a"}, &String{Value: "b"}).Type() != ERROR_OBJ {
		t.Error("uniform(str,str) should return error")
	}
}

// JSON module comprehensive
func TestCreateJsonModuleComprehensive(t *testing.T) {
	m := CreateJsonModule()
	// dumps with no args
	dumpsFn := m.Fields["dumps"].(*Builtin)
	if dumpsFn.Fn().Type() != ERROR_OBJ {
		t.Error("dumps() should return error")
	}
	// dumps with bool
	if dumpsFn.Fn(True).(*String).Value != "true" {
		t.Errorf("dumps(True) wrong")
	}
	// dumps with list
	if dumpsFn.Fn(NewList([]Object{&Integer{Value: 1}, &String{Value: "a"}})).(*String).Value != `[1,"a"]` {
		t.Errorf("dumps(list) wrong")
	}
	// dumps with dict
	d := NewDict()
	d.Set(&String{Value: "k"}, &Integer{Value: 1})
	if dumpsFn.Fn(d).(*String).Value != `{"k":1}` {
		t.Errorf("dumps(dict) wrong")
	}
	// loads with no args
	loadsFn := m.Fields["loads"].(*Builtin)
	if loadsFn.Fn().Type() != ERROR_OBJ {
		t.Error("loads() should return error")
	}
	// loads with non-string
	if loadsFn.Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("loads(int) should return error")
	}
	// loads with invalid JSON
	if loadsFn.Fn(&String{Value: "not json"}).Type() != ERROR_OBJ {
		t.Error("loads('not json') should return error")
	}
	// loads with nested list
	result := loadsFn.Fn(&String{Value: `[1,"a",true]`})
	l := result.(*List)
	if len(l.Elements) != 3 {
		t.Errorf("loads nested list length wrong")
	}
	// loads with nested dict
	jsonStr := "{\"a\":1,\"b\":\"c\"}"
	result2 := loadsFn.Fn(&String{Value: jsonStr})
	dict := result2.(*Dict)
	if dict.Size() != 2 {
		t.Errorf("loads nested dict size wrong")
	}
	// loads with false
	if loadsFn.Fn(&String{Value: "false"}) != False {
		t.Errorf("loads('false') wrong")
	}
}

// String module comprehensive
func TestCreateStringModuleComprehensive(t *testing.T) {
	m := CreateStringModule()
	// maketrans with 3 args
	maketransFn := m.Fields["maketrans"].(*Builtin)
	result := maketransFn.Fn(&String{Value: "ab"}, &String{Value: "xy"}, &String{Value: "c"})
	if result.Type() != DICT_OBJ {
		t.Errorf("maketrans(from, to, del) type wrong")
	}
	// capitalize with no args
	capFn := m.Fields["capitalize"].(*Builtin)
	if capFn.Fn().Type() != ERROR_OBJ {
		t.Error("capitalize() should return error")
	}
}

// Time module comprehensive
func TestCreateTimeModuleComprehensive(t *testing.T) {
	m := CreateTimeModule()
	// localtime with arg - note: current impl ignores args
	localtimeFn := m.Fields["localtime"].(*Builtin)
	result := localtimeFn.Fn()
	if len(result.(*Tuple).Elements) != 8 {
		t.Errorf("localtime() wrong")
	}
	// sleep
	sleepFn := m.Fields["sleep"].(*Builtin)
	result2 := sleepFn.Fn(&Float{Value: 0})
	if result2 != None_ {
		t.Errorf("sleep should return None")
	}
	// sleep with integer
	result3 := sleepFn.Fn(&Integer{Value: 0})
	if result3 != None_ {
		t.Errorf("sleep(int) should return None")
	}
	// sleep with wrong type
	result4 := sleepFn.Fn(&String{Value: "a"})
	if result4.Type() != ERROR_OBJ {
		t.Error("sleep(string) should return error")
	}
	// sleep with no args
	result5 := sleepFn.Fn()
	if result5.Type() != ERROR_OBJ {
		t.Error("sleep() should return error")
	}
	// ctime with integer
	ctimeFn := m.Fields["ctime"].(*Builtin)
	result6 := ctimeFn.Fn(&Integer{Value: 0})
	if result6.(*String).Value == "" {
		t.Error("ctime(0) should return non-empty")
	}
	// ctime with too many args
	result7 := ctimeFn.Fn(&Integer{Value: 0}, &Integer{Value: 1})
	if result7.Type() != ERROR_OBJ {
		t.Error("ctime(0, 1) should return error")
	}
	// ctime with wrong type
	result8 := ctimeFn.Fn(&String{Value: "a"})
	if result8.Type() != ERROR_OBJ {
		t.Error("ctime(string) should return error")
	}
}

// FormatString comprehensive
func TestFormatStringComprehensive(t *testing.T) {
	// %f with integer
	if r := FormatString("val: %f", &Integer{Value: 5}); r != "val: 5.000000" {
		t.Errorf("FormatString %%f with int = %s", r)
	}
	// %d with float
	if r := FormatString("val: %d", &Float{Value: 42.5}); r != "val: 42" {
		t.Errorf("FormatString %%d with float = %s", r)
	}
	// %s with non-string
	if r := FormatString("val: %s", &Integer{Value: 42}); r != "val: 42" {
		t.Errorf("FormatString %%s with int = %s", r)
	}
	// %g
	if r := FormatString("val: %g", &Float{Value: 3.14}); r != "val: 3.14" {
		t.Errorf("FormatString %%g = %s", r)
	}
}

// compareObjectsForSort comprehensive
func TestCompareObjectsForSortComprehensive(t *testing.T) {
	// Float comparisons
	if r := compareObjectsForSort(&Float{Value: 2.0}, &Float{Value: 1.0}); r != 1 {
		t.Errorf("float 2.0 vs 1.0 = %d, want 1", r)
	}
	if r := compareObjectsForSort(&Float{Value: 1.0}, &Float{Value: 1.0}); r != 0 {
		t.Errorf("float 1.0 vs 1.0 = %d, want 0", r)
	}
	// String comparisons
	if r := compareObjectsForSort(&String{Value: "b"}, &String{Value: "a"}); r != 1 {
		t.Errorf("string b vs a = %d, want 1", r)
	}
	if r := compareObjectsForSort(&String{Value: "a"}, &String{Value: "a"}); r != 0 {
		t.Errorf("string a vs a = %d, want 0", r)
	}
	// Mixed types
	if r := compareObjectsForSort(&Integer{Value: 1}, &String{Value: "a"}); r != 0 {
		t.Errorf("mixed types should return 0")
	}
}

// Range Len with zero step
func TestRangeLenZeroStep(t *testing.T) {
	r := NewRange(0, 0, 0)
	if r.Len() != 0 {
		t.Errorf("Range with zero step Len should be 0")
	}
}

// Range GetItem with zero step
func TestRangeGetItemZeroStep(t *testing.T) {
	r := NewRange(0, 0, 0)
	_, ok := r.GetItem(0)
	if ok {
		t.Error("GetItem on zero-step range should fail")
	}
}

// Range ToList with negative step
func TestRangeToListNegativeStep(t *testing.T) {
	r := NewRange(5, 0, -1)
	list := r.ToList()
	if len(list) != 5 {
		t.Errorf("ToList negative step length = %d, want 5", len(list))
	}
	if list[0].(*Integer).Value != 5 {
		t.Errorf("ToList[0] = %d, want 5", list[0].(*Integer).Value)
	}
}

// Range GetItem with step
func TestRangeGetItemWithStep(t *testing.T) {
	r := NewRange(0, 10, 3)
	item, ok := r.GetItem(1)
	if !ok || item.(*Integer).Value != 3 {
		t.Errorf("GetItem(1) with step 3 = %v, want 3", item)
	}
}

// PatternFindall with no matches
func TestPatternFindallNoMatches(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	result := PatternFindall(p, "abc")
	if len(result.(*List).Elements) != 0 {
		t.Errorf("findall no matches should return empty list")
	}
}

// PatternFinditer with no matches
func TestPatternFinditerNoMatches(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	result := PatternFinditer(p, "abc")
	if len(result.(*List).Elements) != 0 {
		t.Errorf("finditer no matches should return empty list")
	}
}

// PatternSplit with no matches
func TestPatternSplitNoMatches(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	result := PatternSplit(p, "abc")
	if len(result.(*List).Elements) != 1 {
		t.Errorf("split no matches should return 1-element list")
	}
}

// RegexMatch group with multiple args - returns first group only
func TestRegexMatchGroupMultipleArgs(t *testing.T) {
	m := &RegexMatch{
		Groups:      []string{"hello 42", "hello", "42"},
		GroupStarts: []int{0, 0, 6},
		GroupEnds:   []int{8, 5, 8},
	}
	attr, _ := m.GetAttr("group")
	result := attr.(*Builtin).Fn(&Integer{Value: 1}, &Integer{Value: 2})
	if result.Type() != STRING_OBJ {
		t.Errorf("group(1, 2) should return string, got %s", result.Type())
	}
}

// RegexMatch start/end/span with group index
func TestRegexMatchStartEndSpanWithGroup(t *testing.T) {
	m := &RegexMatch{
		Groups:      []string{"hello 42", "hello", "42"},
		GroupStarts: []int{0, 0, 6},
		GroupEnds:   []int{8, 5, 8},
	}
	attr, _ := m.GetAttr("start")
	result := attr.(*Builtin).Fn(&Integer{Value: 1})
	if result.(*Integer).Value != 0 {
		t.Errorf("start(1) = %d, want 0", result.(*Integer).Value)
	}
	attr2, _ := m.GetAttr("end")
	result2 := attr2.(*Builtin).Fn(&Integer{Value: 1})
	if result2.(*Integer).Value != 5 {
		t.Errorf("end(1) = %d, want 5", result2.(*Integer).Value)
	}
	attr3, _ := m.GetAttr("span")
	result3 := attr3.(*Builtin).Fn(&Integer{Value: 1})
	tup := result3.(*Tuple)
	if tup.Elements[0].(*Integer).Value != 0 || tup.Elements[1].(*Integer).Value != 5 {
		t.Errorf("span(1) wrong")
	}
}

// Set GetAttr update with list/tuple
func TestSetGetAttr_UpdateWithListAndTuple(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	attr, _ := s.GetAttr("update")
	// With list
	attr.(*Builtin).Fn(NewList([]Object{&Integer{Value: 2}}))
	if s.Size() != 2 {
		t.Errorf("update with list wrong")
	}
	// With set
	s2 := NewSet()
	s2.Add(&Integer{Value: 3})
	attr.(*Builtin).Fn(s2)
	if s.Size() != 3 {
		t.Errorf("update with set wrong")
	}
}

// Set GetAttr difference_update with list/tuple
func TestSetGetAttr_DifferenceUpdateWithListAndTuple(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	s.Add(&Integer{Value: 2})
	attr, _ := s.GetAttr("difference_update")
	attr.(*Builtin).Fn(NewList([]Object{&Integer{Value: 1}}))
	if s.Contains(&Integer{Value: 1}) {
		t.Error("difference_update with list should remove")
	}
}

// Set GetAttr __ior__ with list/tuple
func TestSetGetAttr_IORWithListAndTuple(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	attr, _ := s.GetAttr("__ior__")
	attr.(*Builtin).Fn(NewList([]Object{&Integer{Value: 2}}))
	if s.Size() != 2 {
		t.Errorf("__ior__ with list wrong")
	}
	attr.(*Builtin).Fn(NewTuple([]Object{&Integer{Value: 3}}))
	if s.Size() != 3 {
		t.Errorf("__ior__ with tuple wrong")
	}
}

// Dict GetAttr update with list of tuples
func TestDictGetAttr_UpdateWithListOfTuples(t *testing.T) {
	d := NewDict()
	attr, _ := d.GetAttr("update")
	pairs := NewList([]Object{
		NewTuple([]Object{&String{Value: "a"}, &Integer{Value: 1}}),
	})
	attr.(*Builtin).Fn(pairs)
	if d.Size() != 1 {
		t.Errorf("update with list of tuples wrong")
	}
}

// Dict GetAttr fromkeys with default
func TestDictGetAttr_FromKeysWithDefault(t *testing.T) {
	d := NewDict()
	attr, _ := d.GetAttr("fromkeys")
	keys := NewList([]Object{&String{Value: "a"}, &String{Value: "b"}})
	result := attr.(*Builtin).Fn(keys, &Integer{Value: 42})
	dict := result.(*Dict)
	val, _ := dict.Get(&String{Value: "a"})
	if val.(*Integer).Value != 42 {
		t.Errorf("fromkeys with default wrong")
	}
}

// Dict GetAttr pop with default and without
func TestDictGetAttr_PopDefault(t *testing.T) {
	d := NewDict()
	attr, _ := d.GetAttr("pop")
	// Missing key with default
	result := attr.(*Builtin).Fn(&String{Value: "missing"}, &Integer{Value: 0})
	if result.(*Integer).Value != 0 {
		t.Errorf("pop with default wrong")
	}
	// Missing key without default
	result2 := attr.(*Builtin).Fn(&String{Value: "missing"})
	if result2.Type() != ERROR_OBJ {
		t.Error("pop missing key without default should return error")
	}
}

// Instance GetAttr with SuperClasses
func TestInstanceGetAttrWithSuperClasses(t *testing.T) {
	b1 := &Class{Name: "B1", Methods: map[string]Object{"m1": &Builtin{Name: "m1"}}}
	b2 := &Class{Name: "B2", Methods: map[string]Object{"m2": &Builtin{Name: "m2"}}}
	child := &Class{Name: "Child", SuperClasses: []*Class{b1, b2}, Methods: map[string]Object{}}
	inst := &Instance{Class: child, Fields: map[string]Object{}}
	if _, ok := inst.GetAttr("m1"); !ok {
		t.Error("should find m1 from SuperClasses")
	}
	if _, ok := inst.GetAttr("m2"); !ok {
		t.Error("should find m2 from SuperClasses")
	}
}

// Instance GetAttr with MRO
func TestInstanceGetAttrWithMRO(t *testing.T) {
	base := &Class{Name: "Base", Methods: map[string]Object{"mro_m": &Builtin{Name: "mro_m"}}}
	child := &Class{Name: "Child", SuperClass: base, Methods: map[string]Object{}}
	child.ComputeMRO()
	inst := &Instance{Class: child, Fields: map[string]Object{}}
	if _, ok := inst.GetAttr("mro_m"); !ok {
		t.Error("should find mro_m via MRO")
	}
}

// Class IsSlotAllowed with SuperClasses
func TestClassIsSlotAllowedWithSuperClasses(t *testing.T) {
	base := &Class{Name: "Base", Slots: []string{"y"}}
	child := &Class{Name: "Child", SuperClasses: []*Class{base}}
	if !child.IsSlotAllowed("y") {
		t.Error("y should be allowed via SuperClasses slots")
	}
}

// Class AllSlotNames with Super
func TestClassAllSlotNamesWithSuper(t *testing.T) {
	base := &Class{Name: "Base", Slots: []string{"x"}}
	child := &Class{Name: "Child", SuperClass: base, Slots: []string{"y"}}
	names := child.AllSlotNames()
	if len(names) != 2 {
		t.Errorf("AllSlotNames = %v, want 2 elements", names)
	}
}

// String GetAttr rfind with start/end
func TestStringGetAttr_RfindWithStartEnd(t *testing.T) {
	s := &String{Value: "hello hello"}
	attr, _ := s.GetAttr("rfind")
	result := attr.(*Builtin).Fn(&String{Value: "hello"}, &Integer{Value: 0}, &Integer{Value: 5})
	if result.(*Integer).Value != 0 {
		t.Errorf("rfind with start/end wrong")
	}
}

// String GetAttr translate with int replacement
func TestStringGetAttr_TranslateIntRepl(t *testing.T) {
	table := NewDict()
	table.Set(&Integer{Value: 97}, &Integer{Value: 120}) // 'a' -> 'x'
	s := &String{Value: "abc"}
	attr, _ := s.GetAttr("translate")
	result := attr.(*Builtin).Fn(table)
	if result.Inspect() != "xbc" {
		t.Errorf("translate with int replacement wrong")
	}
}

// String GetAttr format_map with Instance
func TestStringGetAttr_FormatMapInstance(t *testing.T) {
	cls := &Class{Name: "A", Methods: map[string]Object{}}
	inst := &Instance{Class: cls, Fields: map[string]Object{"name": &String{Value: "world"}}}
	s := &String{Value: "hello {name}"}
	attr, _ := s.GetAttr("format_map")
	result := attr.(*Builtin).Fn(inst)
	if result.Inspect() != "hello world" {
		t.Errorf("format_map with instance wrong")
	}
}

// Set GetAttr intersection with no args
func TestSetGetAttr_IntersectionNoArgs(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	attr, _ := s.GetAttr("intersection")
	result := attr.(*Builtin).Fn()
	if result.(*Set).Size() != 0 {
		t.Errorf("intersection with no args should return empty set")
	}
}

// Set GetAttr union with list
func TestSetGetAttr_UnionWithListTuple(t *testing.T) {
	s := NewSet()
	s.Add(&Integer{Value: 1})
	attr, _ := s.GetAttr("union")
	result := attr.(*Builtin).Fn(NewList([]Object{&Integer{Value: 2}}))
	if result.(*Set).Size() != 2 {
		t.Errorf("union with list wrong, got %d", result.(*Set).Size())
	}
}

// RegexPattern subn
func TestRegexPatternGetAttr_Subn(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	attr, ok := p.GetAttr("subn")
	if !ok {
		t.Fatal("expected subn attribute")
	}
	result := attr.(*Builtin).Fn(&String{Value: "X"}, &String{Value: "a1b23c"})
	tup := result.(*Tuple)
	if len(tup.Elements) != 2 {
		t.Errorf("subn should return 2-element tuple")
	}
	if tup.Elements[0].(*String).Value != "aXbXc" {
		t.Errorf("subn result wrong")
	}
	if tup.Elements[1].(*Integer).Value != 2 {
		t.Errorf("subn count wrong")
	}
}

// RegexPattern sub with count
func TestRegexPatternSubWithCount(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	attr, _ := p.GetAttr("sub")
	result := attr.(*Builtin).Fn(&String{Value: "X"}, &String{Value: "a1b23c"}, &Integer{Value: 1})
	if result.(*String).Value != "aXb23c" {
		t.Errorf("sub with count wrong")
	}
}

// RegexPattern sub error cases
func TestRegexPatternSubErrors(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	attr, _ := p.GetAttr("sub")
	// No args
	if attr.(*Builtin).Fn().Type() != ERROR_OBJ {
		t.Error("sub() should return error")
	}
	// Wrong first arg type
	if attr.(*Builtin).Fn(&Integer{Value: 1}, &String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("sub(int, str) should return error")
	}
	// Wrong second arg type
	if attr.(*Builtin).Fn(&String{Value: "X"}, &Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("sub(str, int) should return error")
	}
}

// RegexPattern subn error cases
func TestRegexPatternSubnErrors(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	attr, _ := p.GetAttr("subn")
	if attr.(*Builtin).Fn().Type() != ERROR_OBJ {
		t.Error("subn() should return error")
	}
}

// RegexPattern split error
func TestRegexPatternSplitError(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	attr, _ := p.GetAttr("split")
	// No args
	if attr.(*Builtin).Fn().Type() != ERROR_OBJ {
		t.Error("split() should return error")
	}
	// Wrong type
	if attr.(*Builtin).Fn(&Integer{Value: 1}).Type() != ERROR_OBJ {
		t.Error("split(int) should return error")
	}
}

// RegexMatch groups with default
func TestRegexMatchGroupsWithDefault(t *testing.T) {
	m := &RegexMatch{
		Groups:      []string{"hello 42", "hello", "42"},
		GroupStarts: []int{0, 0, 6},
		GroupEnds:   []int{8, 5, 8},
	}
	attr, _ := m.GetAttr("groups")
	result := attr.(*Builtin).Fn(&String{Value: "N/A"})
	tup := result.(*Tuple)
	if len(tup.Elements) != 2 {
		t.Errorf("groups with default wrong")
	}
}

// RegexMatch start/end/span with no group
func TestRegexMatchStartEndSpanNoGroup(t *testing.T) {
	m := &RegexMatch{
		Groups:      []string{"hello"},
		GroupStarts: []int{0},
		GroupEnds:   []int{5},
	}
	attr, _ := m.GetAttr("start")
	result := attr.(*Builtin).Fn(&Integer{Value: 0})
	if result.(*Integer).Value != 0 {
		t.Errorf("start(0) wrong")
	}
}

// RegexMatch start with no args
func TestRegexMatchStartNoArgs(t *testing.T) {
	m := &RegexMatch{
		Groups:      []string{"hello"},
		GroupStarts: []int{0},
		GroupEnds:   []int{5},
	}
	attr, _ := m.GetAttr("start")
	result := attr.(*Builtin).Fn()
	if result.(*Integer).Value != 0 {
		t.Errorf("start() wrong")
	}
}

// RegexMatch end with no args
func TestRegexMatchEndNoArgs(t *testing.T) {
	m := &RegexMatch{
		Groups:      []string{"hello"},
		GroupStarts: []int{0},
		GroupEnds:   []int{5},
	}
	attr, _ := m.GetAttr("end")
	result := attr.(*Builtin).Fn()
	if result.(*Integer).Value != 5 {
		t.Errorf("end() wrong")
	}
}

// RegexMatch span with no args
func TestRegexMatchSpanNoArgs(t *testing.T) {
	m := &RegexMatch{
		Groups:      []string{"hello"},
		GroupStarts: []int{0},
		GroupEnds:   []int{5},
	}
	attr, _ := m.GetAttr("span")
	result := attr.(*Builtin).Fn()
	tup := result.(*Tuple)
	if len(tup.Elements) != 2 {
		t.Errorf("span() wrong")
	}
}

// ExceptionGroup type
func TestExceptionGroupInspectDetailed(t *testing.T) {
	eg := &ExceptionGroup{Message: "group", Exceptions: []Object{NewError("e1"), NewError("e2")}}
	insp := eg.Inspect()
	if insp == "" {
		t.Error("ExceptionGroup inspect should not be empty")
	}
}

// Set inspect with empty
func TestSetInspectEmpty(t *testing.T) {
	s := NewSet()
	if s.Inspect() != "{}" {
		t.Errorf("empty set inspect = %s, want {}", s.Inspect())
	}
}

// Dict inspect with empty
func TestDictInspectEmpty(t *testing.T) {
	d := NewDict()
	if d.Inspect() != "{}" {
		t.Errorf("empty dict inspect = %s, want {}", d.Inspect())
	}
}

// Dict inspect with multiple entries
func TestDictInspectMultiple(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	d.Set(&String{Value: "b"}, &Integer{Value: 2})
	insp := d.Inspect()
	if insp == "{}" {
		t.Errorf("dict with entries should not be empty")
	}
}

// ==================== Additional Coverage: PatternCallableSub ====================

func TestPatternCallableSub(t *testing.T) {
	// Save original callback
	origFn := callFunctionFn
	defer func() { callFunctionFn = origFn }()

	// Set up callback for PatternCallableSub to work
	SetCallFunctionCallback(func(callee Object, args ...Object) Object {
		return callee.(*Builtin).Fn(args...)
	})

	p := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	// Test with a callable repl that returns a string
	repl := &Builtin{
		Name: "replacer",
		Fn: func(args ...Object) Object {
			return &String{Value: "NUM"}
		},
	}
	result := PatternCallableSub(p, repl, "a1b23c", 0)
	if result.Type() == ERROR_OBJ {
		t.Fatalf("PatternCallableSub returned error: %s", result.Inspect())
	}
	if result.(*String).Value != "aNUMbNUMc" {
		t.Errorf("PatternCallableSub wrong: got %s", result.(*String).Value)
	}

	// Test with count > 0
	result2 := PatternCallableSub(p, repl, "a1b23c", 1)
	if result2.(*String).Value != "aNUMb23c" {
		t.Errorf("PatternCallableSub with count wrong: got %s", result2.(*String).Value)
	}

	// Test with no matches
	p2 := &RegexPattern{Regexp: regexp.MustCompile(`\d+`), Pattern: `\d+`}
	result3 := PatternCallableSub(p2, repl, "abc", 0)
	if result3.(*String).Value != "abc" {
		t.Errorf("PatternCallableSub no match wrong: got %s", result3.(*String).Value)
	}

	// Test with callable that returns non-string (uses Inspect)
	repl2 := &Builtin{
		Name: "int_replacer",
		Fn: func(args ...Object) Object {
			return &Integer{Value: 99}
		},
	}
	result4 := PatternCallableSub(p, repl2, "a1b", 0)
	if result4.(*String).Value != "a99b" {
		t.Errorf("PatternCallableSub non-string result wrong: got %s", result4.(*String).Value)
	}
}

// ==================== Additional Coverage: SetCallFunctionCallback ====================

func TestSetCallFunctionCallback(t *testing.T) {
	// Save original
	origFn := callFunctionFn
	defer func() { callFunctionFn = origFn }()

	// Test setting callback
	called := false
	SetCallFunctionCallback(func(callee Object, args ...Object) Object {
		called = true
		return &String{Value: "called"}
	})
	result := CallFunction(&Builtin{Name: "test"})
	if !called {
		t.Error("callback should have been called")
	}
	if result.(*String).Value != "called" {
		t.Errorf("CallFunction with callback wrong: got %s", result.(*String).Value)
	}
}

// ==================== Additional Coverage: PatternFindall with multiple groups ====================

func TestPatternFindallMultipleGroups(t *testing.T) {
	p := &RegexPattern{Regexp: regexp.MustCompile(`(\w+)-(\d+)`), Pattern: `(\w+)-(\d+)`}
	result := PatternFindall(p, "abc-123 def-456")
	list := result.(*List)
	if len(list.Elements) != 2 {
		t.Fatalf("findall with multiple groups: expected 2, got %d", len(list.Elements))
	}
	tup := list.Elements[0].(*Tuple)
	if len(tup.Elements) != 2 {
		t.Errorf("tuple should have 2 elements, got %d", len(tup.Elements))
	}
	if tup.Elements[0].(*String).Value != "abc" {
		t.Errorf("first group wrong: got %s", tup.Elements[0].(*String).Value)
	}
}

// ==================== Additional Coverage: Dict HashKey default case ====================

func TestDictHashKeyDefaultCase(t *testing.T) {
	d := NewDict()
	// Test with an object type that falls through to default case
	got := d.HashKey(&Generator{})
	if got == "" {
		t.Error("HashKey for Generator should return non-empty string")
	}
	if len(got) < 5 {
		t.Errorf("HashKey for Generator should return obj: prefix, got %s", got)
	}
}

// ==================== Additional Coverage: Os module success paths ====================

func TestCreateOsModuleSuccessPaths(t *testing.T) {
	m := CreateOsModule()
	// getcwd with no args (success path)
	getcwdFn := m.Fields["getcwd"].(*Builtin)
	result := getcwdFn.Fn()
	if result.Type() != STRING_OBJ {
		t.Errorf("getcwd() should return string, got %s", result.Type())
	}

	// listdir with valid path
	listdirFn := m.Fields["listdir"].(*Builtin)
	result2 := listdirFn.Fn(&String{Value: "/tmp"})
	if result2.Type() == ERROR_OBJ {
		t.Errorf("listdir('/tmp') should succeed, got error: %s", result2.Inspect())
	}

	// getenv with set env var
	os.Setenv("TEST_GO_PY_OBJ", "hello")
	getenvFn := m.Fields["getenv"].(*Builtin)
	result3 := getenvFn.Fn(&String{Value: "TEST_GO_PY_OBJ"})
	if result3.(*String).Value != "hello" {
		t.Errorf("getenv for set var wrong: got %s", result3.(*String).Value)
	}

	// getenv with empty result and no default
	os.Unsetenv("NONEXISTENT_VAR_XYZ_123")
	result4 := getenvFn.Fn(&String{Value: "NONEXISTENT_VAR_XYZ_123"})
	if result4 != None_ {
		t.Errorf("getenv for unset var without default should return None, got %s", result4.Type())
	}

	// chdir with valid path
	chdirFn := m.Fields["chdir"].(*Builtin)
	result5 := chdirFn.Fn(&String{Value: "/tmp"})
	if result5 != None_ {
		t.Errorf("chdir('/tmp') should return None, got %s", result5.Type())
	}
	// Restore cwd
	os.Chdir("/workspace")

	// mkdir with valid path (temp)
	tmpDir := "/tmp/test_go_py_mkdir_12345"
	mkdirFn := m.Fields["mkdir"].(*Builtin)
	result6 := mkdirFn.Fn(&String{Value: tmpDir})
	if result6 != None_ {
		t.Errorf("mkdir should return None, got %s", result6.Type())
	}
	os.RemoveAll(tmpDir)

	// mkdir with mode
	tmpDir2 := "/tmp/test_go_py_mkdir_mode_12345"
	result7 := mkdirFn.Fn(&String{Value: tmpDir2}, &Integer{Value: 755})
	if result7 != None_ {
		t.Errorf("mkdir with mode should return None, got %s", result7.Type())
	}
	os.RemoveAll(tmpDir2)

	// remove with valid path
	tmpFile := "/tmp/test_go_py_remove_12345.txt"
	os.WriteFile(tmpFile, []byte("test"), 0644)
	removeFn := m.Fields["remove"].(*Builtin)
	result8 := removeFn.Fn(&String{Value: tmpFile})
	if result8 != None_ {
		t.Errorf("remove should return None, got %s", result8.Type())
	}

	// rename with valid paths
	tmpFile1 := "/tmp/test_go_py_rename1_12345.txt"
	tmpFile2 := "/tmp/test_go_py_rename2_12345.txt"
	os.WriteFile(tmpFile1, []byte("test"), 0644)
	renameFn := m.Fields["rename"].(*Builtin)
	result9 := renameFn.Fn(&String{Value: tmpFile1}, &String{Value: tmpFile2})
	if result9 != None_ {
		t.Errorf("rename should return None, got %s", result9.Type())
	}
	os.Remove(tmpFile2)

	// os.environ should exist
	if _, ok := m.Fields["environ"]; !ok {
		t.Error("os.environ should exist")
	}
}

// ==================== Additional Coverage: String module maketrans ====================

func TestCreateStringModuleMaketransDetailed(t *testing.T) {
	m := CreateStringModule()
	// maketrans with dict
	maketransFn := m.Fields["maketrans"].(*Builtin)
	dictArg := NewDict()
	dictArg.Set(&Integer{Value: 97}, &String{Value: "x"})
	result := maketransFn.Fn(dictArg)
	if result.Type() != DICT_OBJ {
		t.Errorf("maketrans(dict) should return dict, got %s", result.Type())
	}

	// maketrans with no args
	result2 := maketransFn.Fn()
	if result2.Type() != ERROR_OBJ {
		t.Error("maketrans() should return error")
	}

	// maketrans with non-string first arg
	result3 := maketransFn.Fn(&Integer{Value: 1}, &String{Value: "a"})
	if result3.Type() != ERROR_OBJ {
		t.Error("maketrans(int, str) should return error")
	}

	// maketrans with different length strings
	result4 := maketransFn.Fn(&String{Value: "ab"}, &String{Value: "x"})
	if result4.Type() != ERROR_OBJ {
		t.Error("maketrans('ab', 'x') should return error")
	}

	// maketrans with 2 args (from, to)
	result5 := maketransFn.Fn(&String{Value: "ab"}, &String{Value: "xy"})
	if result5.Type() != DICT_OBJ {
		t.Errorf("maketrans('ab', 'xy') should return dict, got %s", result5.Type())
	}

	// capitalize with string
	capFn := m.Fields["capitalize"].(*Builtin)
	result6 := capFn.Fn(&String{Value: "hello"})
	if result6.(*String).Value != "Hello" {
		t.Errorf("capitalize('hello') wrong: got %s", result6.(*String).Value)
	}

	// capitalize with empty string
	result7 := capFn.Fn(&String{Value: ""})
	if result7.(*String).Value != "" {
		t.Errorf("capitalize('') wrong: got %s", result7.(*String).Value)
	}

	// capitalize with non-string
	result8 := capFn.Fn(&Integer{Value: 1})
	if result8.Type() != ERROR_OBJ {
		t.Error("capitalize(int) should return error")
	}

	// Test all string module constants
	for _, name := range []string{"ascii_lowercase", "ascii_uppercase", "hexdigits", "octdigits", "punctuation", "printable"} {
		if _, ok := m.Fields[name]; !ok {
			t.Errorf("string.%s should exist", name)
		}
	}
}

// ==================== Additional Coverage: DictValues/DictItems GetItem negative index ====================

func TestDictValuesGetItemNegative(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	d.Set(&String{Value: "b"}, &Integer{Value: 2})
	dv := NewDictValues(d)
	// Negative index
	item, ok := dv.GetItem(-1)
	if !ok {
		t.Error("GetItem(-1) should succeed")
	}
	if item == nil {
		t.Error("GetItem(-1) should return non-nil")
	}
	// Out of range negative
	_, ok2 := dv.GetItem(-10)
	if ok2 {
		t.Error("GetItem(-10) should fail")
	}
}

func TestDictItemsGetItemNegative(t *testing.T) {
	d := NewDict()
	d.Set(&String{Value: "a"}, &Integer{Value: 1})
	d.Set(&String{Value: "b"}, &Integer{Value: 2})
	di := NewDictItems(d)
	// Negative index
	item, ok := di.GetItem(-1)
	if !ok {
		t.Error("GetItem(-1) should succeed")
	}
	tup := item.(*Tuple)
	if len(tup.Elements) != 2 {
		t.Errorf("GetItem(-1) tuple should have 2 elements")
	}
	// Out of range negative
	_, ok2 := di.GetItem(-10)
	if ok2 {
		t.Error("GetItem(-10) should fail")
	}
}

// ==================== Additional Coverage: Class FindClassAttr with MRO ====================

func TestClassFindClassAttrWithMRO(t *testing.T) {
	base := &Class{Name: "Base", Methods: map[string]Object{"mro_m": &Builtin{Name: "mro_m"}}}
	child := &Class{Name: "Child", SuperClass: base, Methods: map[string]Object{}}
	child.ComputeMRO()
	if _, ok := child.FindClassAttr("mro_m"); !ok {
		t.Error("should find mro_m via MRO")
	}
	if _, ok := child.FindClassAttr("nonexistent"); ok {
		t.Error("should not find nonexistent")
	}
}

// ==================== Additional Coverage: IsInstanceOf with Instance ====================

func TestIsInstanceOfWithInstance(t *testing.T) {
	base := &Class{Name: "Base"}
	child := &Class{Name: "Child", SuperClass: base}
	inst := &Instance{Class: child}
	if !IsInstanceOf(inst, base) {
		t.Error("instance of Child should be instance of Base")
	}
	if !IsInstanceOf(inst, child) {
		t.Error("instance should be instance of its own class")
	}
	unrelated := &Class{Name: "Unrelated"}
	if IsInstanceOf(inst, unrelated) {
		t.Error("instance should not be instance of unrelated class")
	}
}

// ==================== Additional Coverage: IsInstanceOf default case ====================

func TestIsInstanceOfDefaultCase(t *testing.T) {
	objClass := &Class{Name: "object"}
	// Generator falls through to default case
	if !IsInstanceOf(&Generator{}, objClass) {
		t.Error("Generator should be instance of object")
	}
	// Non-object class
	if IsInstanceOf(&Generator{}, &Class{Name: "int"}) {
		t.Error("Generator should not be instance of int")
	}
}

// ==================== Additional Coverage: Zip Len with String and Tuple ====================

func TestZipLenWithStringAndTuple(t *testing.T) {
	s := &String{Value: "abc"}
	tup := NewTuple([]Object{&Integer{Value: 1}, &Integer{Value: 2}})
	z := NewZip([]Object{s, tup})
	if z.Len() != 2 {
		t.Errorf("Zip with String and Tuple Len() = %d, want 2", z.Len())
	}
}

// ==================== Additional Coverage: Random module success paths ====================

func TestCreateRandomModuleSuccessPaths(t *testing.T) {
	m := CreateRandomModule()
	// seed with float
	seedFn := m.Fields["seed"].(*Builtin)
	if seedFn.Fn(&Float{Value: 1.5}) != None_ {
		t.Error("seed(float) should return None")
	}
	// seed with wrong type
	if seedFn.Fn(&String{Value: "a"}).Type() != ERROR_OBJ {
		t.Error("seed(string) should return error")
	}
	// randint success
	randintFn := m.Fields["randint"].(*Builtin)
	result := randintFn.Fn(&Integer{Value: 1}, &Integer{Value: 10})
	if result.Type() != INTEGER_OBJ {
		t.Errorf("randint(1, 10) should return integer, got %s", result.Type())
	}
	// choice success
	choiceFn := m.Fields["choice"].(*Builtin)
	result2 := choiceFn.Fn(NewList([]Object{&Integer{Value: 42}}))
	if result2.(*Integer).Value != 42 {
		t.Errorf("choice([42]) should return 42")
	}
	// choice with empty list
	if choiceFn.Fn(NewList([]Object{})).Type() != ERROR_OBJ {
		t.Error("choice([]) should return error")
	}
	// shuffle success
	shuffleFn := m.Fields["shuffle"].(*Builtin)
	l := NewList([]Object{&Integer{Value: 1}, &Integer{Value: 2}, &Integer{Value: 3}})
	result3 := shuffleFn.Fn(l)
	if result3 != None_ {
		t.Errorf("shuffle should return None")
	}
	// uniform success with integers
	uniformFn := m.Fields["uniform"].(*Builtin)
	result4 := uniformFn.Fn(&Integer{Value: 0}, &Integer{Value: 10})
	if result4.Type() != FLOAT_OBJ {
		t.Errorf("uniform(int, int) should return float, got %s", result4.Type())
	}
	// uniform success with floats
	result5 := uniformFn.Fn(&Float{Value: 0.0}, &Float{Value: 1.0})
	if result5.Type() != FLOAT_OBJ {
		t.Errorf("uniform(float, float) should return float")
	}
}

// ==================== Additional Coverage: Datetime module ====================

func TestCreateDatetimeModuleComprehensive(t *testing.T) {
	m := CreateDatetimeModule()
	// date.today
	dateMod := m.Fields["date"].(*Module)
	todayFn := dateMod.Fields["today"].(*Builtin)
	result := todayFn.Fn()
	if result.(*String).Value == "" {
		t.Error("date.today() should return non-empty")
	}
}
