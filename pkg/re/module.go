package re

import (
	"regexp"
	"strings"

	"github.com/go-py/go-python/pkg/objects"
)

func compilePattern(pattern string, flags int64) (*objects.RegexPattern, error) {
	goFlags := flagsToGoFlags(flags)
	re, err := regexp.Compile(goFlags + pattern)
	if err != nil {
		return nil, err
	}
	return &objects.RegexPattern{
		Regexp:  re,
		Pattern: pattern,
		Flags:   flags,
	}, nil
}

const (
	ReIGNORECASE = 2
	ReMULTILINE  = 8
	ReDOTALL     = 16
	ReASCII      = 256
	ReUNICODE    = 32
)

func flagsToGoFlags(flags int64) string {
	prefix := ""
	if flags&ReIGNORECASE != 0 {
		prefix += "(?i)"
	}
	if flags&ReMULTILINE != 0 {
		prefix += "(?m)"
	}
	if flags&ReDOTALL != 0 {
		prefix += "(?s)"
	}
	return prefix
}

func getPatternAndFlags(args []objects.Object) (*objects.RegexPattern, objects.Object) {
	if len(args) < 1 {
		return nil, objects.NewTypeError("missing required argument")
	}

	switch arg := args[0].(type) {
	case *objects.String:
		var flags int64
		if len(args) >= 2 {
			if f, ok := args[1].(*objects.Integer); ok {
				flags = f.Value
			}
		}
		p, err := compilePattern(arg.Value, flags)
		if err != nil {
			return nil, objects.NewError("invalid regular expression: %s", err.Error())
		}
		return p, nil
	case *objects.RegexPattern:
		return arg, nil
	default:
		return nil, objects.NewTypeError("first argument must be a string or compiled pattern")
	}
}

func getStringArg(args []objects.Object, index int) (string, objects.Object) {
	if index >= len(args) {
		return "", objects.NewTypeError("missing required argument")
	}
	s, ok := args[index].(*objects.String)
	if !ok {
		return "", objects.NewTypeError("argument must be a string")
	}
	return s.Value, nil
}

// callableSub performs regex substitution where repl is a callable.
// For each match, the callable is invoked with a RegexMatch object and
// its return value is used as the replacement string.
func callableSub(p *objects.RegexPattern, repl objects.Object, s string, count int) objects.Object {
	var locs [][]int
	if count > 0 {
		locs = p.Regexp.FindAllStringSubmatchIndex(s, count)
	} else {
		locs = p.Regexp.FindAllStringSubmatchIndex(s, -1)
	}

	if locs == nil {
		return &objects.String{Value: s}
	}

	var buf strings.Builder
	prev := 0
	for _, loc := range locs {
		buf.WriteString(s[prev:loc[0]])

		// Create a RegexMatch object for this match
		match := newMatchFromLoc(p, s, loc)

		// Call the callable with the match object
		result := objects.CallFunction(repl, match)
		if result.Type() == objects.ERROR_OBJ {
			return result
		}

		// Convert the result to a string
		var replacement string
		if strObj, ok := result.(*objects.String); ok {
			replacement = strObj.Value
		} else {
			replacement = result.Inspect()
		}

		buf.WriteString(replacement)
		prev = loc[1]
	}
	buf.WriteString(s[prev:])

	return &objects.String{Value: buf.String()}
}

// newMatchFromLoc creates a RegexMatch from submatch indices.
// This is a package-level version of the same function in objects,
// needed here to create match objects for callable repl support.
func newMatchFromLoc(p *objects.RegexPattern, s string, loc []int) *objects.RegexMatch {
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
	return &objects.RegexMatch{
		Groups:         groups,
		GroupStarts:    groupStarts,
		GroupEnds:      groupEnds,
		OriginalString: s,
		Pattern:        p,
	}
}

func CreateReModule() *objects.Module {
	module := &objects.Module{
		Name:   "re",
		Fields: make(map[string]objects.Object),
	}

	// Constants
	module.Fields["IGNORECASE"] = &objects.Integer{Value: ReIGNORECASE}
	module.Fields["MULTILINE"] = &objects.Integer{Value: ReMULTILINE}
	module.Fields["DOTALL"] = &objects.Integer{Value: ReDOTALL}
	module.Fields["ASCII"] = &objects.Integer{Value: ReASCII}
	module.Fields["UNICODE"] = &objects.Integer{Value: ReUNICODE}

	// re.compile(pattern, flags=0)
	module.Fields["compile"] = &objects.Builtin{
		Name: "re.compile",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("compile() takes at least 1 argument")
			}
			patternStr, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("compile() first argument must be a string")
			}
			var flags int64
			if len(args) >= 2 {
				if f, ok := args[1].(*objects.Integer); ok {
					flags = f.Value
				}
			}
			p, err := compilePattern(patternStr.Value, flags)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			return p
		},
	}

	// re.search(pattern, string, flags=0)
	module.Fields["search"] = &objects.Builtin{
		Name: "re.search",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("search() takes at least 2 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, errObj := getStringArg(args, 1)
			if errObj != nil {
				return errObj
			}
			return objects.PatternSearch(p, s)
		},
	}

	// re.match(pattern, string, flags=0)
	module.Fields["match"] = &objects.Builtin{
		Name: "re.match",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("match() takes at least 2 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, errObj := getStringArg(args, 1)
			if errObj != nil {
				return errObj
			}
			return objects.PatternMatch(p, s)
		},
	}

	// re.fullmatch(pattern, string, flags=0)
	module.Fields["fullmatch"] = &objects.Builtin{
		Name: "re.fullmatch",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("fullmatch() takes at least 2 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, errObj := getStringArg(args, 1)
			if errObj != nil {
				return errObj
			}
			return objects.PatternFullmatch(p, s)
		},
	}

	// re.findall(pattern, string, flags=0)
	module.Fields["findall"] = &objects.Builtin{
		Name: "re.findall",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("findall() takes at least 2 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, errObj := getStringArg(args, 1)
			if errObj != nil {
				return errObj
			}
			return objects.PatternFindall(p, s)
		},
	}

	// re.finditer(pattern, string, flags=0)
	module.Fields["finditer"] = &objects.Builtin{
		Name: "re.finditer",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("finditer() takes at least 2 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, errObj := getStringArg(args, 1)
			if errObj != nil {
				return errObj
			}
			return objects.PatternFinditer(p, s)
		},
	}

	// re.sub(pattern, repl, string, count=0)
	module.Fields["sub"] = &objects.Builtin{
		Name: "re.sub",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 3 {
				return objects.NewTypeError("sub() takes at least 3 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, ok := args[2].(*objects.String)
			if !ok {
				return objects.NewTypeError("sub() third argument must be a string")
			}
			var count int
			if len(args) >= 4 {
				if c, ok := args[3].(*objects.Integer); ok {
					count = int(c.Value)
				}
			}

			repl := args[1]

			// If repl is callable, use callable replacement
			if objects.IsCallable(repl) {
				return callableSub(p, repl, s.Value, count)
			}

			// String replacement
			replStr, ok := repl.(*objects.String)
			if !ok {
				return objects.NewTypeError("sub() replacement must be a string or callable")
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
			return &objects.String{Value: result}
		},
	}

	// re.subn(pattern, repl, string, count=0)
	module.Fields["subn"] = &objects.Builtin{
		Name: "re.subn",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 3 {
				return objects.NewTypeError("subn() takes at least 3 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, ok := args[2].(*objects.String)
			if !ok {
				return objects.NewTypeError("subn() third argument must be a string")
			}
			var count int
			if len(args) >= 4 {
				if c, ok := args[3].(*objects.Integer); ok {
					count = int(c.Value)
				}
			}

			repl := args[1]

			// If repl is callable, use callable replacement
			if objects.IsCallable(repl) {
				result := callableSub(p, repl, s.Value, count)
				if result.Type() == objects.ERROR_OBJ {
					return result
				}
				resultStr := result.(*objects.String).Value
				// Count the number of substitutions
				findN := count
				if findN <= 0 {
					findN = -1
				}
				locs := p.Regexp.FindAllStringIndex(s.Value, findN)
				n := 0
				if locs != nil {
					n = len(locs)
				}
				return &objects.Tuple{Elements: []objects.Object{&objects.String{Value: resultStr}, &objects.Integer{Value: int64(n)}}}
			}

			// String replacement
			replStr, ok := repl.(*objects.String)
			if !ok {
				return objects.NewTypeError("subn() replacement must be a string or callable")
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
			return &objects.Tuple{Elements: []objects.Object{&objects.String{Value: result}, &objects.Integer{Value: int64(n)}}}
		},
	}

	// re.split(pattern, string, maxsplit=0)
	module.Fields["split"] = &objects.Builtin{
		Name: "re.split",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("split() takes at least 2 arguments")
			}
			p, errObj := getPatternAndFlags(args)
			if errObj != nil {
				return errObj
			}
			s, errObj := getStringArg(args, 1)
			if errObj != nil {
				return errObj
			}
			return objects.PatternSplit(p, s)
		},
	}

	// re.escape(string)
	module.Fields["escape"] = &objects.Builtin{
		Name: "re.escape",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("escape() takes exactly 1 argument")
			}
			s, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("escape() argument must be a string")
			}
			return &objects.String{Value: regexp.QuoteMeta(s.Value)}
		},
	}

	return module
}
