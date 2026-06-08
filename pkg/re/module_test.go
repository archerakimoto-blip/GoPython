package re

import (
	"testing"

	"github.com/go-py/go-python/pkg/objects"
)

func TestCreateReModule(t *testing.T) {
	module := CreateReModule()
	if module.Name != "re" {
		t.Errorf("Expected module name 're', got '%s'", module.Name)
	}

	// Check required functions exist
	requiredFuncs := []string{
		"compile", "search", "match", "fullmatch",
		"findall", "finditer", "sub", "subn", "split", "escape",
	}
	for _, fn := range requiredFuncs {
		if _, ok := module.Fields[fn]; !ok {
			t.Errorf("Missing function: %s", fn)
		}
	}

	// Check constants
	constants := map[string]int64{
		"IGNORECASE": 2,
		"MULTILINE":  8,
		"DOTALL":     16,
		"ASCII":      256,
		"UNICODE":    32,
	}
	for name, expected := range constants {
		field, ok := module.Fields[name]
		if !ok {
			t.Errorf("Missing constant: %s", name)
			continue
		}
		intVal, ok := field.(*objects.Integer)
		if !ok {
			t.Errorf("Constant %s is not an integer", name)
			continue
		}
		if intVal.Value != expected {
			t.Errorf("Constant %s: expected %d, got %d", name, expected, intVal.Value)
		}
	}
}

func TestReSearch(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123def"})
	if result == objects.None_ {
		t.Fatal("Expected match, got None")
	}

	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	group0, _ := match.GetAttr("group")
	if group0 == nil {
		t.Fatal("Match should have 'group' attribute")
	}
}

func TestReMatch(t *testing.T) {
	module := CreateReModule()
	matchFn := module.Fields["match"].(*objects.Builtin)

	// Should match at start
	result := matchFn.Fn(&objects.String{Value: "hello"}, &objects.String{Value: "hello world"})
	if result == objects.None_ {
		t.Fatal("Expected match at start, got None")
	}

	// Should not match in middle
	result = matchFn.Fn(&objects.String{Value: "world"}, &objects.String{Value: "hello world"})
	if result != objects.None_ {
		t.Fatal("Expected no match for non-start pattern")
	}
}

func TestReFindall(t *testing.T) {
	module := CreateReModule()
	findallFn := module.Fields["findall"].(*objects.Builtin)

	result := findallFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "a1b22c333"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(list.Elements))
	}

	// Check individual matches
	expected := []string{"1", "22", "333"}
	for i, exp := range expected {
		s, ok := list.Elements[i].(*objects.String)
		if !ok {
			t.Errorf("Element %d: expected String, got %T", i, list.Elements[i])
			continue
		}
		if s.Value != exp {
			t.Errorf("Element %d: expected %q, got %q", i, exp, s.Value)
		}
	}
}

func TestReSub(t *testing.T) {
	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	result := subFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "a1b22c333"})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if s.Value != "aXbXcX" {
		t.Errorf("Expected 'aXbXcX', got '%s'", s.Value)
	}
}

func TestReSplit(t *testing.T) {
	module := CreateReModule()
	splitFn := module.Fields["split"].(*objects.Builtin)

	result := splitFn.Fn(&objects.String{Value: "[,;]"}, &objects.String{Value: "a,b;c,d"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 4 {
		t.Fatalf("Expected 4 parts, got %d", len(list.Elements))
	}
}

func TestReEscape(t *testing.T) {
	module := CreateReModule()
	escapeFn := module.Fields["escape"].(*objects.Builtin)

	result := escapeFn.Fn(&objects.String{Value: `\d+.txt`})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if s.Value == "" {
		t.Error("Escape result should not be empty")
	}
}

func TestReCompile(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)

	result := compileFn.Fn(&objects.String{Value: `\d+`})
	pattern, ok := result.(*objects.RegexPattern)
	if !ok {
		t.Fatalf("Expected RegexPattern, got %T", result)
	}
	if pattern.Pattern != `\d+` {
		t.Errorf("Expected pattern `\\d+`, got '%s'", pattern.Pattern)
	}

	// Test pattern.search
	searchAttr, ok := pattern.GetAttr("search")
	if !ok {
		t.Fatal("Pattern should have 'search' attribute")
	}
	searchBuiltin, ok := searchAttr.(*objects.Builtin)
	if !ok {
		t.Fatalf("Expected Builtin for search, got %T", searchAttr)
	}

	matchResult := searchBuiltin.Fn(&objects.String{Value: "abc456def"})
	if matchResult == objects.None_ {
		t.Fatal("Pattern.search should find match")
	}
}

func TestRegexPatternGetAttr(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	pattern := compileFn.Fn(&objects.String{Value: `\d+`}).(*objects.RegexPattern)

	// Test pattern attribute
	patAttr, ok := pattern.GetAttr("pattern")
	if !ok {
		t.Fatal("Pattern should have 'pattern' attribute")
	}
	patStr, ok := patAttr.(*objects.String)
	if !ok || patStr.Value != `\d+` {
		t.Errorf("Expected pattern `\\d+`, got %v", patAttr)
	}

	// Test flags attribute
	flagsAttr, ok := pattern.GetAttr("flags")
	if !ok {
		t.Fatal("Pattern should have 'flags' attribute")
	}
	_, ok = flagsAttr.(*objects.Integer)
	if !ok {
		t.Errorf("Expected Integer for flags, got %T", flagsAttr)
	}
}

func TestRegexMatchGetAttr(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `(\d+)-(\d+)`}, &objects.String{Value: "abc12-34def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Test string attribute
	strAttr, ok := match.GetAttr("string")
	if !ok {
		t.Fatal("Match should have 'string' attribute")
	}
	_ = strAttr

	// Test re attribute
	reAttr, ok := match.GetAttr("re")
	if !ok {
		t.Fatal("Match should have 're' attribute")
	}
	_, ok = reAttr.(*objects.RegexPattern)
	if !ok {
		t.Errorf("Expected RegexPattern for 're', got %T", reAttr)
	}

	// Test group function
	groupAttr, ok := match.GetAttr("group")
	if !ok {
		t.Fatal("Match should have 'group' attribute")
	}
	groupFn, ok := groupAttr.(*objects.Builtin)
	if !ok {
		t.Fatalf("Expected Builtin for group, got %T", groupAttr)
	}
	groupResult := groupFn.Fn(&objects.Integer{Value: 0})
	groupStr, ok := groupResult.(*objects.String)
	if !ok {
		t.Fatalf("Expected String for group(0), got %T", groupResult)
	}
	if groupStr.Value != "12-34" {
		t.Errorf("Expected group(0)='12-34', got '%s'", groupStr.Value)
	}

	// Test group(1)
	group1Result := groupFn.Fn(&objects.Integer{Value: 1})
	group1Str, ok := group1Result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String for group(1), got %T", group1Result)
	}
	if group1Str.Value != "12" {
		t.Errorf("Expected group(1)='12', got '%s'", group1Str.Value)
	}
}

// ==================== Additional RE Tests ====================

func TestReFullmatch(t *testing.T) {
	module := CreateReModule()
	fullmatchFn := module.Fields["fullmatch"].(*objects.Builtin)

	// Full match should succeed when the whole string matches
	result := fullmatchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "123"})
	if result == objects.None_ {
		t.Fatal("Expected fullmatch to match '123'")
	}

	// Full match should fail when only part matches
	result = fullmatchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "123abc"})
	if result != objects.None_ {
		t.Fatal("Expected fullmatch to not match '123abc'")
	}
}

func TestReFinditer(t *testing.T) {
	module := CreateReModule()
	finditerFn := module.Fields["finditer"].(*objects.Builtin)

	result := finditerFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "a1b22c333"})
	lst, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}

	if len(lst.Elements) != 3 {
		t.Errorf("Expected 3 matches from finditer, got %d", len(lst.Elements))
	}
}

func TestReSubn(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "a1b22c333"})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple, got %T", result)
	}
	if len(tup.Elements) != 2 {
		t.Fatalf("Expected 2-element tuple, got %d", len(tup.Elements))
	}

	// First element is the result string
	resultStr, ok := tup.Elements[0].(*objects.String)
	if !ok {
		t.Fatalf("Expected String for first element, got %T", tup.Elements[0])
	}
	if resultStr.Value != "aXbXcX" {
		t.Errorf("Expected 'aXbXcX', got '%s'", resultStr.Value)
	}

	// Second element is the count
	countVal, ok := tup.Elements[1].(*objects.Integer)
	if !ok {
		t.Fatalf("Expected Integer for second element, got %T", tup.Elements[1])
	}
	if countVal.Value != 3 {
		t.Errorf("Expected 3 substitutions, got %d", countVal.Value)
	}
}

func TestReSubWithCount(t *testing.T) {
	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	// Replace only 1 occurrence
	result := subFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "a1b22c333"}, &objects.Integer{Value: 1})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if s.Value != "aXb22c333" {
		t.Errorf("Expected 'aXb22c333', got '%s'", s.Value)
	}
}

func TestReSubNoMatch(t *testing.T) {
	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	// No match should return original string
	result := subFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "hello"})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("Expected 'hello', got '%s'", s.Value)
	}
}

func TestReSearchNoMatch(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "hello"})
	if result != objects.None_ {
		t.Errorf("Expected None for no match, got %T", result)
	}
}

func TestReMatchNoMatch(t *testing.T) {
	module := CreateReModule()
	matchFn := module.Fields["match"].(*objects.Builtin)

	result := matchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123"})
	if result != objects.None_ {
		t.Errorf("Expected None for no match at start, got %T", result)
	}
}

func TestReCompileWithFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)

	// Compile with IGNORECASE flag
	result := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	pattern, ok := result.(*objects.RegexPattern)
	if !ok {
		t.Fatalf("Expected RegexPattern, got %T", result)
	}
	if pattern.Flags != ReIGNORECASE {
		t.Errorf("Expected flags=%d, got %d", ReIGNORECASE, pattern.Flags)
	}

	// Test case-insensitive search
	searchAttr, _ := pattern.GetAttr("search")
	searchBuiltin := searchAttr.(*objects.Builtin)
	matchResult := searchBuiltin.Fn(&objects.String{Value: "HELLO world"})
	if matchResult == objects.None_ {
		t.Fatal("Expected case-insensitive match")
	}
}

func TestReSearchWithFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	searchFn := module.Fields["search"].(*objects.Builtin)

	// Compile with IGNORECASE flag first, then search
	pattern := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	result := searchFn.Fn(pattern, &objects.String{Value: "HELLO world"})
	if result == objects.None_ {
		t.Fatal("Expected case-insensitive match")
	}
}

func TestReMultilineFlag(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	searchFn := module.Fields["search"].(*objects.Builtin)

	// Compile with MULTILINE flag first, then search
	pattern := compileFn.Fn(&objects.String{Value: `^world`}, &objects.Integer{Value: ReMULTILINE})
	result := searchFn.Fn(pattern, &objects.String{Value: "hello\nworld"})
	if result == objects.None_ {
		t.Fatal("Expected multiline match")
	}
}

func TestReDotallFlag(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	searchFn := module.Fields["search"].(*objects.Builtin)

	// Compile with DOTALL flag first, then search
	pattern := compileFn.Fn(&objects.String{Value: `a.b`}, &objects.Integer{Value: ReDOTALL})
	result := searchFn.Fn(pattern, &objects.String{Value: "a\nb"})
	if result == objects.None_ {
		t.Fatal("Expected DOTALL match")
	}
}

func TestReCompileInvalidPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)

	result := compileFn.Fn(&objects.String{Value: `[invalid`})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for invalid pattern, got %T", result)
	}
}

func TestReCompileNoArgs(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)

	result := compileFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReSearchNoArgs(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReMatchNoArgs(t *testing.T) {
	module := CreateReModule()
	matchFn := module.Fields["match"].(*objects.Builtin)

	result := matchFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReFullmatchNoArgs(t *testing.T) {
	module := CreateReModule()
	fullmatchFn := module.Fields["fullmatch"].(*objects.Builtin)

	result := fullmatchFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReFindallNoArgs(t *testing.T) {
	module := CreateReModule()
	findallFn := module.Fields["findall"].(*objects.Builtin)

	result := findallFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReFinditerNoArgs(t *testing.T) {
	module := CreateReModule()
	finditerFn := module.Fields["finditer"].(*objects.Builtin)

	result := finditerFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReSubNoArgs(t *testing.T) {
	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	result := subFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReSubnNoArgs(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReSplitNoArgs(t *testing.T) {
	module := CreateReModule()
	splitFn := module.Fields["split"].(*objects.Builtin)

	result := splitFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestReEscapeSpecialChars(t *testing.T) {
	module := CreateReModule()
	escapeFn := module.Fields["escape"].(*objects.Builtin)

	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `hello`},
		{`.`, `\.`},
		{`*`, `\*`},
		{`+`, `\+`},
		{`?`, `\?`},
		{`[`, `\[`},
		{`]`, `\]`},
		{`(`, `\(`},
		{`)`, `\)`},
	}

	for _, tt := range tests {
		result := escapeFn.Fn(&objects.String{Value: tt.input})
		s, ok := result.(*objects.String)
		if !ok {
			t.Fatalf("Expected String, got %T for input '%s'", result, tt.input)
		}
		if s.Value != tt.expected {
			t.Errorf("escape(%q): expected %q, got %q", tt.input, tt.expected, s.Value)
		}
	}
}

func TestReEscapeWrongType(t *testing.T) {
	module := CreateReModule()
	escapeFn := module.Fields["escape"].(*objects.Builtin)

	result := escapeFn.Fn(&objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string argument, got %T", result)
	}
}

func TestReEscapeWrongArgCount(t *testing.T) {
	module := CreateReModule()
	escapeFn := module.Fields["escape"].(*objects.Builtin)

	result := escapeFn.Fn(&objects.String{Value: "a"}, &objects.String{Value: "b"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for too many args, got %T", result)
	}
}

func TestReCompileWrongType(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)

	result := compileFn.Fn(&objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string argument, got %T", result)
	}
}

func TestReSearchWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	searchFn := module.Fields["search"].(*objects.Builtin)

	// Compile a pattern first
	pattern := compileFn.Fn(&objects.String{Value: `\d+`})

	// Use the compiled pattern with search
	result := searchFn.Fn(pattern, &objects.String{Value: "abc123def"})
	if result == objects.None_ {
		t.Fatal("Expected match with compiled pattern")
	}
}

func TestReMatchWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	matchFn := module.Fields["match"].(*objects.Builtin)

	// Compile a pattern first
	pattern := compileFn.Fn(&objects.String{Value: `\d+`})

	// Use the compiled pattern with match
	result := matchFn.Fn(pattern, &objects.String{Value: "123abc"})
	if result == objects.None_ {
		t.Fatal("Expected match with compiled pattern")
	}
}

func TestReFullmatchWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	fullmatchFn := module.Fields["fullmatch"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: `\d+`})

	result := fullmatchFn.Fn(pattern, &objects.String{Value: "123"})
	if result == objects.None_ {
		t.Fatal("Expected fullmatch with compiled pattern")
	}
}

func TestReFindallWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	findallFn := module.Fields["findall"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: `\d+`})

	result := findallFn.Fn(pattern, &objects.String{Value: "a1b22c333"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(list.Elements))
	}
}

func TestReFinditerWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	finditerFn := module.Fields["finditer"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: `\d+`})

	result := finditerFn.Fn(pattern, &objects.String{Value: "a1b22c333"})
	lst, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}

	if len(lst.Elements) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(lst.Elements))
	}
}

func TestReSubWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	subFn := module.Fields["sub"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: `\d+`})

	result := subFn.Fn(pattern, &objects.String{Value: "X"}, &objects.String{Value: "a1b22c333"})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if s.Value != "aXbXcX" {
		t.Errorf("Expected 'aXbXcX', got '%s'", s.Value)
	}
}

func TestReSubnWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	subnFn := module.Fields["subn"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: `\d+`})

	result := subnFn.Fn(pattern, &objects.String{Value: "X"}, &objects.String{Value: "a1b22c333"})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple, got %T", result)
	}
	if len(tup.Elements) != 2 {
		t.Fatalf("Expected 2-element tuple, got %d", len(tup.Elements))
	}
}

func TestReSplitWithPattern(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	splitFn := module.Fields["split"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: "[,;]"})

	result := splitFn.Fn(pattern, &objects.String{Value: "a,b;c,d"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 4 {
		t.Errorf("Expected 4 parts, got %d", len(list.Elements))
	}
}

func TestReSearchWrongType(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.Integer{Value: 42}, &objects.String{Value: "hello"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string pattern, got %T", result)
	}
}

func TestReSearchWrongStringType(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string input, got %T", result)
	}
}

func TestMatchObjectGroups(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `(\d+)-(\d+)`}, &objects.String{Value: "abc12-34def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Test groups() method
	groupsAttr, ok := match.GetAttr("groups")
	if !ok {
		t.Fatal("Match should have 'groups' attribute")
	}
	groupsFn, ok := groupsAttr.(*objects.Builtin)
	if !ok {
		t.Fatalf("Expected Builtin for groups, got %T", groupsAttr)
	}
	groupsResult := groupsFn.Fn()
	groupsList, ok := groupsResult.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple for groups(), got %T", groupsResult)
	}
	if len(groupsList.Elements) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groupsList.Elements))
	}
}

func TestMatchObjectStart(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Test start() method
	startAttr, ok := match.GetAttr("start")
	if !ok {
		t.Fatal("Match should have 'start' attribute")
	}
	startFn, ok := startAttr.(*objects.Builtin)
	if !ok {
		t.Fatalf("Expected Builtin for start, got %T", startAttr)
	}
	startResult := startFn.Fn()
	startInt, ok := startResult.(*objects.Integer)
	if !ok {
		t.Fatalf("Expected Integer for start(), got %T", startResult)
	}
	if startInt.Value != 3 {
		t.Errorf("Expected start=3, got %d", startInt.Value)
	}
}

func TestMatchObjectEnd(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Test end() method
	endAttr, ok := match.GetAttr("end")
	if !ok {
		t.Fatal("Match should have 'end' attribute")
	}
	endFn, ok := endAttr.(*objects.Builtin)
	if !ok {
		t.Fatalf("Expected Builtin for end, got %T", endAttr)
	}
	endResult := endFn.Fn()
	endInt, ok := endResult.(*objects.Integer)
	if !ok {
		t.Fatalf("Expected Integer for end(), got %T", endResult)
	}
	if endInt.Value != 6 {
		t.Errorf("Expected end=6, got %d", endInt.Value)
	}
}

func TestMatchObjectSpan(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Test span() method
	spanAttr, ok := match.GetAttr("span")
	if !ok {
		t.Fatal("Match should have 'span' attribute")
	}
	spanFn, ok := spanAttr.(*objects.Builtin)
	if !ok {
		t.Fatalf("Expected Builtin for span, got %T", spanAttr)
	}
	spanResult := spanFn.Fn()
	spanTuple, ok := spanResult.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple for span(), got %T", spanResult)
	}
	if len(spanTuple.Elements) != 2 {
		t.Fatalf("Expected 2 elements in span, got %d", len(spanTuple.Elements))
	}
	startVal := spanTuple.Elements[0].(*objects.Integer).Value
	endVal := spanTuple.Elements[1].(*objects.Integer).Value
	if startVal != 3 || endVal != 6 {
		t.Errorf("Expected span=(3,6), got (%d,%d)", startVal, endVal)
	}
}

func TestMatchObjectGroupdict(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `(?P<first>\d+)-(?P<second>\d+)`}, &objects.String{Value: "abc12-34def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Verify the match has the expected groups
	groupAttr, ok := match.GetAttr("group")
	if !ok {
		t.Fatal("Match should have 'group' attribute")
	}
	groupFn := groupAttr.(*objects.Builtin)

	// Test group(1) - named group "first"
	group1Result := groupFn.Fn(&objects.Integer{Value: 1})
	group1Str, ok := group1Result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String for group(1), got %T", group1Result)
	}
	if group1Str.Value != "12" {
		t.Errorf("Expected group(1)='12', got '%s'", group1Str.Value)
	}

	// Test group(2) - named group "second"
	group2Result := groupFn.Fn(&objects.Integer{Value: 2})
	group2Str, ok := group2Result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String for group(2), got %T", group2Result)
	}
	if group2Str.Value != "34" {
		t.Errorf("Expected group(2)='34', got '%s'", group2Str.Value)
	}
}

func TestReFindallWithGroups(t *testing.T) {
	module := CreateReModule()
	findallFn := module.Fields["findall"].(*objects.Builtin)

	// When pattern has groups, findall returns the groups
	result := findallFn.Fn(&objects.String{Value: `(\d+)-(\d+)`}, &objects.String{Value: "12-34 56-78"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(list.Elements))
	}
}

func TestReSplitSimple(t *testing.T) {
	module := CreateReModule()
	splitFn := module.Fields["split"].(*objects.Builtin)

	result := splitFn.Fn(&objects.String{Value: `\s+`}, &objects.String{Value: "hello world  foo"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(list.Elements))
	}
}

func TestReSubnWithCount(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "a1b22c333"}, &objects.Integer{Value: 2})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple, got %T", result)
	}

	resultStr := tup.Elements[0].(*objects.String).Value
	countVal := tup.Elements[1].(*objects.Integer).Value

	if resultStr != "aXbXc333" {
		t.Errorf("Expected 'aXbXc333', got '%s'", resultStr)
	}
	if countVal != 2 {
		t.Errorf("Expected 2 substitutions, got %d", countVal)
	}
}

func TestReSubnNoMatch(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "hello"})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple, got %T", result)
	}

	resultStr := tup.Elements[0].(*objects.String).Value
	countVal := tup.Elements[1].(*objects.Integer).Value

	if resultStr != "hello" {
		t.Errorf("Expected 'hello', got '%s'", resultStr)
	}
	if countVal != 0 {
		t.Errorf("Expected 0 substitutions, got %d", countVal)
	}
}

func TestReSubWrongReplType(t *testing.T) {
	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	result := subFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42}, &objects.String{Value: "a1b22c333"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string replacement, got %T", result)
	}
}

func TestReSubWrongStringType(t *testing.T) {
	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	result := subFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string input, got %T", result)
	}
}

func TestCompilePatternHelper(t *testing.T) {
	p, err := compilePattern(`\d+`, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("Expected non-nil pattern")
	}
	if p.Pattern != `\d+` {
		t.Errorf("Expected pattern `\\d+`, got '%s'", p.Pattern)
	}
}

func TestCompilePatternWithFlags(t *testing.T) {
	p, err := compilePattern("hello", ReIGNORECASE)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if p.Flags != ReIGNORECASE {
		t.Errorf("Expected flags=%d, got %d", ReIGNORECASE, p.Flags)
	}
}

func TestCompilePatternInvalid(t *testing.T) {
	_, err := compilePattern(`[invalid`, 0)
	if err == nil {
		t.Error("Expected error for invalid pattern")
	}
}

func TestFlagsToGoFlags(t *testing.T) {
	tests := []struct {
		flags    int64
		expected string
	}{
		{0, ""},
		{ReIGNORECASE, "(?i)"},
		{ReMULTILINE, "(?m)"},
		{ReDOTALL, "(?s)"},
		{ReIGNORECASE | ReMULTILINE, "(?i)(?m)"},
		{ReIGNORECASE | ReDOTALL, "(?i)(?s)"},
	}

	for _, tt := range tests {
		result := flagsToGoFlags(tt.flags)
		if result != tt.expected {
			t.Errorf("flagsToGoFlags(%d): expected %q, got %q", tt.flags, tt.expected, result)
		}
	}
}

func TestGetPatternAndFlagsNoArgs(t *testing.T) {
	p, errObj := getPatternAndFlags([]objects.Object{})
	if p != nil {
		t.Error("Expected nil pattern for no args")
	}
	if errObj == nil {
		t.Error("Expected error for no args")
	}
}

func TestGetPatternAndFlagsWrongType(t *testing.T) {
	p, errObj := getPatternAndFlags([]objects.Object{&objects.Integer{Value: 42}})
	if p != nil {
		t.Error("Expected nil pattern for wrong type")
	}
	if errObj == nil {
		t.Error("Expected error for wrong type")
	}
}

func TestGetStringArg(t *testing.T) {
	// Valid string arg
	s, errObj := getStringArg([]objects.Object{&objects.String{Value: "hello"}}, 0)
	if errObj != nil {
		t.Fatalf("Unexpected error: %v", errObj)
	}
	if s != "hello" {
		t.Errorf("Expected 'hello', got '%s'", s)
	}

	// Out of bounds
	s, errObj = getStringArg([]objects.Object{}, 0)
	if errObj == nil {
		t.Error("Expected error for out of bounds")
	}

	// Wrong type
	s, errObj = getStringArg([]objects.Object{&objects.Integer{Value: 42}}, 0)
	if errObj == nil {
		t.Error("Expected error for wrong type")
	}
}

func TestReConstants(t *testing.T) {
	module := CreateReModule()

	constants := map[string]int64{
		"IGNORECASE": ReIGNORECASE,
		"MULTILINE":  ReMULTILINE,
		"DOTALL":     ReDOTALL,
		"ASCII":      ReASCII,
		"UNICODE":    ReUNICODE,
	}

	for name, expected := range constants {
		field, ok := module.Fields[name]
		if !ok {
			t.Errorf("Missing constant: %s", name)
			continue
		}
		intVal, ok := field.(*objects.Integer)
		if !ok {
			t.Errorf("Constant %s is not an integer", name)
			continue
		}
		if intVal.Value != expected {
			t.Errorf("Constant %s: expected %d, got %d", name, expected, intVal.Value)
		}
	}
}

func TestReFindallNoMatch(t *testing.T) {
	module := CreateReModule()
	findallFn := module.Fields["findall"].(*objects.Builtin)

	result := findallFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "hello world"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(list.Elements))
	}
}

func TestReSplitNoMatch(t *testing.T) {
	module := CreateReModule()
	splitFn := module.Fields["split"].(*objects.Builtin)

	result := splitFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "hello world"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 1 {
		t.Errorf("Expected 1 part (original string), got %d", len(list.Elements))
	}
}

func TestMatchObjectGroupOutOfBounds(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Test group with out-of-bounds index
	groupAttr, _ := match.GetAttr("group")
	groupFn := groupAttr.(*objects.Builtin)
	groupResult := groupFn.Fn(&objects.Integer{Value: 99})
	// Should return an error or None
	if groupResult == nil {
		t.Error("Expected non-nil result for out-of-bounds group")
	}
}

func TestMatchObjectGroupDefault(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123def"})
	match, ok := result.(*objects.RegexMatch)
	if !ok {
		t.Fatalf("Expected RegexMatch, got %T", result)
	}

	// Test group() with no args (same as group(0))
	groupAttr, _ := match.GetAttr("group")
	groupFn := groupAttr.(*objects.Builtin)
	groupResult := groupFn.Fn()
	groupStr, ok := groupResult.(*objects.String)
	if !ok {
		t.Fatalf("Expected String for group(), got %T", groupResult)
	}
	if groupStr.Value != "123" {
		t.Errorf("Expected group()='123', got '%s'", groupStr.Value)
	}
}

func TestNewMatchFromLoc(t *testing.T) {
	p, _ := compilePattern(`(\d+)-(\d+)`, 0)
	loc := []int{3, 8, 3, 5, 6, 8}

	match := newMatchFromLoc(p, "abc12-34def", loc)
	if match == nil {
		t.Fatal("Expected non-nil match")
	}
	if len(match.Groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(match.Groups))
	}
	if match.Groups[0] != "12-34" {
		t.Errorf("Expected group 0='12-34', got '%s'", match.Groups[0])
	}
	if match.Groups[1] != "12" {
		t.Errorf("Expected group 1='12', got '%s'", match.Groups[1])
	}
	if match.Groups[2] != "34" {
		t.Errorf("Expected group 2='34', got '%s'", match.Groups[2])
	}
}

// ==================== Additional RE Tests for 90%+ Coverage ====================

func TestReSubWithCallable(t *testing.T) {
	// Set up the CallFunction callback so callableSub can invoke builtins
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		if builtin, ok := callee.(*objects.Builtin); ok {
			return builtin.Fn(args...)
		}
		return objects.NewTypeError("object is not callable")
	})
	defer objects.SetCallFunctionCallback(nil)

	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	// Create a callable replacement using a Builtin
	callableRepl := &objects.Builtin{
		Name: "replacer",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return &objects.String{Value: "X"}
			}
			// The argument is a RegexMatch - return a string replacement
			return &objects.String{Value: "X"}
		},
	}

	result := subFn.Fn(&objects.String{Value: `\d+`}, callableRepl, &objects.String{Value: "a1b22c333"})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String from callable sub, got %T", result)
	}
	if s.Value != "aXbXcX" {
		t.Errorf("Expected 'aXbXcX' from callable sub, got '%s'", s.Value)
	}
}

func TestReSubWithCallableAndCount(t *testing.T) {
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		if builtin, ok := callee.(*objects.Builtin); ok {
			return builtin.Fn(args...)
		}
		return objects.NewTypeError("object is not callable")
	})
	defer objects.SetCallFunctionCallback(nil)

	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	callableRepl := &objects.Builtin{
		Name: "replacer",
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.String{Value: "X"}
		},
	}

	result := subFn.Fn(&objects.String{Value: `\d+`}, callableRepl, &objects.String{Value: "a1b22c333"}, &objects.Integer{Value: 2})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String from callable sub with count, got %T", result)
	}
	if s.Value != "aXbXc333" {
		t.Errorf("Expected 'aXbXc333' from callable sub with count, got '%s'", s.Value)
	}
}

func TestReSubWithCallableNoMatch(t *testing.T) {
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		if builtin, ok := callee.(*objects.Builtin); ok {
			return builtin.Fn(args...)
		}
		return objects.NewTypeError("object is not callable")
	})
	defer objects.SetCallFunctionCallback(nil)

	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	callableRepl := &objects.Builtin{
		Name: "replacer",
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.String{Value: "X"}
		},
	}

	result := subFn.Fn(&objects.String{Value: `\d+`}, callableRepl, &objects.String{Value: "hello"})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String from callable sub no match, got %T", result)
	}
	if s.Value != "hello" {
		t.Errorf("Expected 'hello' from callable sub no match, got '%s'", s.Value)
	}
}

func TestReSubWithCallableNonStringResult(t *testing.T) {
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		if builtin, ok := callee.(*objects.Builtin); ok {
			return builtin.Fn(args...)
		}
		return objects.NewTypeError("object is not callable")
	})
	defer objects.SetCallFunctionCallback(nil)

	module := CreateReModule()
	subFn := module.Fields["sub"].(*objects.Builtin)

	// Callable that returns a non-string object (should use Inspect())
	callableRepl := &objects.Builtin{
		Name: "replacer",
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.Integer{Value: 99}
		},
	}

	result := subFn.Fn(&objects.String{Value: `\d+`}, callableRepl, &objects.String{Value: "a1b"})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String from callable sub with int result, got %T", result)
	}
	// Integer{99}.Inspect() returns "99"
	if s.Value != "a99b" {
		t.Errorf("Expected 'a99b' from callable sub with int result, got '%s'", s.Value)
	}
}

func TestReSubnWithCallable(t *testing.T) {
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		if builtin, ok := callee.(*objects.Builtin); ok {
			return builtin.Fn(args...)
		}
		return objects.NewTypeError("object is not callable")
	})
	defer objects.SetCallFunctionCallback(nil)

	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	callableRepl := &objects.Builtin{
		Name: "replacer",
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.String{Value: "X"}
		},
	}

	result := subnFn.Fn(&objects.String{Value: `\d+`}, callableRepl, &objects.String{Value: "a1b22c333"})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple from callable subn, got %T", result)
	}
	if len(tup.Elements) != 2 {
		t.Fatalf("Expected 2-element tuple, got %d", len(tup.Elements))
	}

	resultStr := tup.Elements[0].(*objects.String).Value
	countVal := tup.Elements[1].(*objects.Integer).Value

	if resultStr != "aXbXcX" {
		t.Errorf("Expected 'aXbXcX' from callable subn, got '%s'", resultStr)
	}
	if countVal != 3 {
		t.Errorf("Expected 3 substitutions from callable subn, got %d", countVal)
	}
}

func TestReSubnWithCallableAndCount(t *testing.T) {
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		if builtin, ok := callee.(*objects.Builtin); ok {
			return builtin.Fn(args...)
		}
		return objects.NewTypeError("object is not callable")
	})
	defer objects.SetCallFunctionCallback(nil)

	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	callableRepl := &objects.Builtin{
		Name: "replacer",
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.String{Value: "X"}
		},
	}

	result := subnFn.Fn(&objects.String{Value: `\d+`}, callableRepl, &objects.String{Value: "a1b22c333"}, &objects.Integer{Value: 2})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple from callable subn with count, got %T", result)
	}

	resultStr := tup.Elements[0].(*objects.String).Value
	countVal := tup.Elements[1].(*objects.Integer).Value

	if resultStr != "aXbXc333" {
		t.Errorf("Expected 'aXbXc333' from callable subn with count, got '%s'", resultStr)
	}
	if countVal != 2 {
		t.Errorf("Expected 2 substitutions from callable subn with count, got %d", countVal)
	}
}

func TestReSubnWithCallableNoMatch(t *testing.T) {
	objects.SetCallFunctionCallback(func(callee objects.Object, args ...objects.Object) objects.Object {
		if builtin, ok := callee.(*objects.Builtin); ok {
			return builtin.Fn(args...)
		}
		return objects.NewTypeError("object is not callable")
	})
	defer objects.SetCallFunctionCallback(nil)

	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	callableRepl := &objects.Builtin{
		Name: "replacer",
		Fn: func(args ...objects.Object) objects.Object {
			return &objects.String{Value: "X"}
		},
	}

	result := subnFn.Fn(&objects.String{Value: `\d+`}, callableRepl, &objects.String{Value: "hello"})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple from callable subn no match, got %T", result)
	}

	resultStr := tup.Elements[0].(*objects.String).Value
	countVal := tup.Elements[1].(*objects.Integer).Value

	if resultStr != "hello" {
		t.Errorf("Expected 'hello' from callable subn no match, got '%s'", resultStr)
	}
	if countVal != 0 {
		t.Errorf("Expected 0 substitutions from callable subn no match, got %d", countVal)
	}
}

func TestReSubnWrongReplType(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42}, &objects.String{Value: "a1b"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string/non-callable replacement in subn, got %T", result)
	}
}

func TestReSubnWrongStringType(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string third arg in subn, got %T", result)
	}
}

func TestReSubWithCountStringRepl(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "a1b22c333"}, &objects.Integer{Value: 1})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple, got %T", result)
	}

	resultStr := tup.Elements[0].(*objects.String).Value
	countVal := tup.Elements[1].(*objects.Integer).Value

	if resultStr != "aXb22c333" {
		t.Errorf("Expected 'aXb22c333', got '%s'", resultStr)
	}
	if countVal != 1 {
		t.Errorf("Expected 1 substitution, got %d", countVal)
	}
}

func TestReSubnNoMatchStringRepl(t *testing.T) {
	module := CreateReModule()
	subnFn := module.Fields["subn"].(*objects.Builtin)

	result := subnFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "X"}, &objects.String{Value: "hello"})
	tup, ok := result.(*objects.Tuple)
	if !ok {
		t.Fatalf("Expected Tuple, got %T", result)
	}

	resultStr := tup.Elements[0].(*objects.String).Value
	countVal := tup.Elements[1].(*objects.Integer).Value

	if resultStr != "hello" {
		t.Errorf("Expected 'hello', got '%s'", resultStr)
	}
	if countVal != 0 {
		t.Errorf("Expected 0 substitutions, got %d", countVal)
	}
}

func TestReSearchWithCompiledPatternAndFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	searchFn := module.Fields["search"].(*objects.Builtin)

	// Compile with IGNORECASE flag first, then search
	pattern := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	result := searchFn.Fn(pattern, &objects.String{Value: "HELLO world"})
	if result == objects.None_ {
		t.Fatal("Expected case-insensitive match via search with compiled pattern")
	}
}

func TestReMatchWithCompiledPatternAndFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	matchFn := module.Fields["match"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	result := matchFn.Fn(pattern, &objects.String{Value: "HELLO world"})
	if result == objects.None_ {
		t.Fatal("Expected case-insensitive match via match with compiled pattern")
	}
}

func TestReFullmatchWithCompiledPatternAndFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	fullmatchFn := module.Fields["fullmatch"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	result := fullmatchFn.Fn(pattern, &objects.String{Value: "HELLO"})
	if result == objects.None_ {
		t.Fatal("Expected case-insensitive fullmatch with compiled pattern")
	}
}

func TestReFindallWithCompiledPatternAndFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	findallFn := module.Fields["findall"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	result := findallFn.Fn(pattern, &objects.String{Value: "HELLO hello"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 2 {
		t.Errorf("Expected 2 matches with IGNORECASE, got %d", len(list.Elements))
	}
}

func TestReFinditerWithCompiledPatternAndFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	finditerFn := module.Fields["finditer"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	result := finditerFn.Fn(pattern, &objects.String{Value: "HELLO hello"})
	lst, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(lst.Elements) != 2 {
		t.Errorf("Expected 2 matches with IGNORECASE, got %d", len(lst.Elements))
	}
}

func TestReSubWithCompiledPatternAndFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	subFn := module.Fields["sub"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: "hello"}, &objects.Integer{Value: ReIGNORECASE})
	result := subFn.Fn(pattern, &objects.String{Value: "X"}, &objects.String{Value: "HELLO hello"})
	s, ok := result.(*objects.String)
	if !ok {
		t.Fatalf("Expected String, got %T", result)
	}
	if s.Value != "X X" {
		t.Errorf("Expected 'X X' with IGNORECASE, got '%s'", s.Value)
	}
}

func TestReSplitWithCompiledPatternAndFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)
	splitFn := module.Fields["split"].(*objects.Builtin)

	pattern := compileFn.Fn(&objects.String{Value: `[;,]`}, &objects.Integer{Value: 0})
	result := splitFn.Fn(pattern, &objects.String{Value: "a,b;c,d"})
	list, ok := result.(*objects.List)
	if !ok {
		t.Fatalf("Expected List, got %T", result)
	}
	if len(list.Elements) != 4 {
		t.Errorf("Expected 4 parts, got %d", len(list.Elements))
	}
}

func TestReSearchWrongStringArgType(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	// Pattern is valid string but second arg is not a string
	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string second arg in search, got %T", result)
	}
}

func TestReMatchWrongStringArgType(t *testing.T) {
	module := CreateReModule()
	matchFn := module.Fields["match"].(*objects.Builtin)

	result := matchFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string second arg in match, got %T", result)
	}
}

func TestReFullmatchWrongStringArgType(t *testing.T) {
	module := CreateReModule()
	fullmatchFn := module.Fields["fullmatch"].(*objects.Builtin)

	result := fullmatchFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string second arg in fullmatch, got %T", result)
	}
}

func TestReFindallWrongStringArgType(t *testing.T) {
	module := CreateReModule()
	findallFn := module.Fields["findall"].(*objects.Builtin)

	result := findallFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string second arg in findall, got %T", result)
	}
}

func TestReFinditerWrongStringArgType(t *testing.T) {
	module := CreateReModule()
	finditerFn := module.Fields["finditer"].(*objects.Builtin)

	result := finditerFn.Fn(&objects.String{Value: `\d+`}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string second arg in finditer, got %T", result)
	}
}

func TestReSplitWrongStringArgType(t *testing.T) {
	module := CreateReModule()
	splitFn := module.Fields["split"].(*objects.Builtin)

	result := splitFn.Fn(&objects.String{Value: `[;,]`}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string second arg in split, got %T", result)
	}
}

func TestReCompileWithNonIntegerFlags(t *testing.T) {
	module := CreateReModule()
	compileFn := module.Fields["compile"].(*objects.Builtin)

	// Pass a string as flags (should be ignored, not crash)
	result := compileFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "not a flag"})
	pattern, ok := result.(*objects.RegexPattern)
	if !ok {
		t.Fatalf("Expected RegexPattern even with non-integer flags, got %T", result)
	}
	if pattern.Flags != 0 {
		t.Errorf("Expected flags=0 when non-integer flags passed, got %d", pattern.Flags)
	}
}

func TestReSearchWithNonIntegerFlags(t *testing.T) {
	module := CreateReModule()
	searchFn := module.Fields["search"].(*objects.Builtin)

	// Pass a string as flags (should be ignored)
	result := searchFn.Fn(&objects.String{Value: `\d+`}, &objects.String{Value: "abc123"}, &objects.String{Value: "not a flag"})
	if result == objects.None_ {
		t.Fatal("Expected match even with non-integer flags")
	}
}

func TestNewMatchFromLocWithNegativeIndices(t *testing.T) {
	p, _ := compilePattern(`(\d+)-(\d+)`, 0)
	// Simulate a match where some groups didn't participate (negative indices)
	loc := []int{3, 8, -1, -1, 6, 8}

	match := newMatchFromLoc(p, "abc12-34def", loc)
	if match == nil {
		t.Fatal("Expected non-nil match")
	}
	// Group 1 has negative indices - should be empty string
	if match.Groups[1] != "" {
		t.Errorf("Expected empty group 1 for negative indices, got '%s'", match.Groups[1])
	}
	// Group 2 should have value
	if match.Groups[2] != "34" {
		t.Errorf("Expected group 2='34', got '%s'", match.Groups[2])
	}
}

func TestGetPatternAndFlagsWithRegexPattern(t *testing.T) {
	// Test passing a RegexPattern directly
	p, _ := compilePattern(`\d+`, 0)
	result, errObj := getPatternAndFlags([]objects.Object{p})
	if errObj != nil {
		t.Fatalf("Unexpected error: %v", errObj)
	}
	if result != p {
		t.Error("Expected same pattern object back")
	}
}

func TestGetPatternAndFlagsInvalidPattern(t *testing.T) {
	// Test passing an invalid pattern string
	result, errObj := getPatternAndFlags([]objects.Object{&objects.String{Value: `[invalid`}})
	if result != nil {
		t.Error("Expected nil pattern for invalid regex")
	}
	if errObj == nil {
		t.Error("Expected error for invalid regex")
	}
}
