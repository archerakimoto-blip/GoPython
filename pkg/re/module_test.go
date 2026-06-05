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
