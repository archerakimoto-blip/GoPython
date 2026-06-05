package re

import (
	"regexp"

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

func flagsToGoFlags(flags int64) string {
	prefix := ""
	if flags&2 != 0 { // IGNORECASE
		prefix += "(?i)"
	}
	if flags&8 != 0 { // MULTILINE
		prefix += "(?m)"
	}
	if flags&16 != 0 { // DOTALL
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

func CreateReModule() *objects.Module {
	module := &objects.Module{
		Name:   "re",
		Fields: make(map[string]objects.Object),
	}

	// Constants
	module.Fields["IGNORECASE"] = &objects.Integer{Value: 2}
	module.Fields["MULTILINE"] = &objects.Integer{Value: 8}
	module.Fields["DOTALL"] = &objects.Integer{Value: 16}
	module.Fields["ASCII"] = &objects.Integer{Value: 256}
	module.Fields["UNICODE"] = &objects.Integer{Value: 32}

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
			repl, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("sub() replacement must be a string")
			}
			s, ok := args[2].(*objects.String)
			if !ok {
				return objects.NewTypeError("sub() third argument must be a string")
			}
			result := p.Regexp.ReplaceAllString(s.Value, repl.Value)
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
			repl, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("subn() replacement must be a string")
			}
			s, ok := args[2].(*objects.String)
			if !ok {
				return objects.NewTypeError("subn() third argument must be a string")
			}
			result := p.Regexp.ReplaceAllString(s.Value, repl.Value)
			matches := p.Regexp.FindAllString(s.Value, -1)
			n := int64(len(matches))
			return &objects.Tuple{Elements: []objects.Object{&objects.String{Value: result}, &objects.Integer{Value: n}}}
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
