package regex

import (
	"regexp"

	"github.com/go-py/go-python/pkg/objects"
)

const (
	IGNORECASE = 1
	MULTILINE  = 2
	DOTALL     = 4
)

// applyFlags prepends Go regex flag syntax based on Python flags
func applyFlags(pattern string, flags int) string {
	var prefix string
	if flags&IGNORECASE != 0 {
		prefix += "i"
	}
	if flags&MULTILINE != 0 {
		prefix += "m"
	}
	if flags&DOTALL != 0 {
		prefix += "s"
	}
	if prefix != "" {
		return "(?" + prefix + ")" + pattern
	}
	return pattern
}

// compilePattern compiles a regex pattern with optional flags
func compilePattern(pattern string, flags int) (*regexp.Regexp, error) {
	goPattern := applyFlags(pattern, flags)
	return regexp.Compile(goPattern)
}

// makeMatch creates a RegexMatch object from a regexp match result
func makeMatch(matchString string, loc []int, re *regexp.Regexp) *objects.RegexMatch {
	if loc == nil {
		return nil
	}
	groups := make([]string, 0, re.NumSubexp()+1)
	// Group 0 is the full match
	groups = append(groups, matchString[loc[0]:loc[1]])
	// Subgroups
	for i := 1; i < len(loc)/2; i++ {
		start := loc[i*2]
		end := loc[i*2+1]
		if start >= 0 && end >= 0 {
			groups = append(groups, matchString[start:end])
		} else {
			groups = append(groups, "")
		}
	}
	return &objects.RegexMatch{
		MatchString: matchString,
		Groups:      groups,
		StartPos:    loc[0],
		EndPos:      loc[1],
	}
}

// getRegexp extracts the compiled regexp from a RegexPattern
func getRegexp(rp *objects.RegexPattern) *regexp.Regexp {
	return rp.GetRegex().(*regexp.Regexp)
}

// PatternMatch implements match/search for a compiled pattern (exported for VM use)
func PatternMatch(rp *objects.RegexPattern, s string, search bool) objects.Object {
	re := getRegexp(rp)
	var loc []int
	if search {
		loc = re.FindStringIndex(s)
	} else {
		// match = anchored at beginning
		loc = re.FindStringIndex(s)
		if loc != nil && loc[0] != 0 {
			return objects.None_
		}
	}
	if loc == nil {
		return objects.None_
	}
	return makeMatch(s, loc, re)
}

// PatternFullmatch checks if the entire string matches (exported for VM use)
func PatternFullmatch(rp *objects.RegexPattern, s string) objects.Object {
	re := getRegexp(rp)
	loc := re.FindStringIndex(s)
	if loc == nil || loc[0] != 0 || loc[1] != len(s) {
		return objects.None_
	}
	return makeMatch(s, loc, re)
}

// PatternFindall returns all non-overlapping matches (exported for VM use)
func PatternFindall(rp *objects.RegexPattern, s string) objects.Object {
	re := getRegexp(rp)
	matches := re.FindAllStringSubmatch(s, -1)
	elements := make([]objects.Object, 0, len(matches))

	numGroups := re.NumSubexp()

	for _, match := range matches {
		if numGroups == 0 {
			// No groups: return list of matched strings
			elements = append(elements, &objects.String{Value: match[0]})
		} else if numGroups == 1 {
			// One group: return list of group values
			elements = append(elements, &objects.String{Value: match[1]})
		} else {
			// Multiple groups: return list of tuples
			tupleElems := make([]objects.Object, numGroups)
			for i := 1; i <= numGroups; i++ {
				tupleElems[i-1] = &objects.String{Value: match[i]}
			}
			elements = append(elements, &objects.Tuple{Elements: tupleElems})
		}
	}

	return &objects.List{Elements: elements}
}

// patternFinditer returns list of Match objects
func patternFinditer(rp *objects.RegexPattern, s string) objects.Object {
	re := rp.GetRegex().(*regexp.Regexp)
	allLocs := re.FindAllStringSubmatchIndex(s, -1)
	elements := make([]objects.Object, 0, len(allLocs))

	for _, loc := range allLocs {
		match := makeMatchFromSubmatchIndex(s, loc, re)
		elements = append(elements, match)
	}

	return &objects.List{Elements: elements}
}

// makeMatchFromSubmatchIndex creates a RegexMatch from submatch index slice
func makeMatchFromSubmatchIndex(matchString string, loc []int, re *regexp.Regexp) *objects.RegexMatch {
	groups := make([]string, 0, re.NumSubexp()+1)
	// Group 0 is the full match
	groups = append(groups, matchString[loc[0]:loc[1]])
	// Subgroups
	for i := 1; i < len(loc)/2; i++ {
		start := loc[i*2]
		end := loc[i*2+1]
		if start >= 0 && end >= 0 {
			groups = append(groups, matchString[start:end])
		} else {
			groups = append(groups, "")
		}
	}
	return &objects.RegexMatch{
		MatchString: matchString,
		Groups:      groups,
		StartPos:    loc[0],
		EndPos:      loc[1],
	}
}

// PatternSub replaces matches in the string (exported for VM use)
func PatternSub(rp *objects.RegexPattern, replacement string, s string, count int) objects.Object {
	re := rp.GetRegex().(*regexp.Regexp)
	result := replaceN(re, s, replacement, count)
	return &objects.String{Value: result}
}

// replaceN replaces at most n occurrences (n <= 0 means replace all)
func replaceN(re *regexp.Regexp, s, replacement string, n int) string {
	if n <= 0 {
		return re.ReplaceAllString(s, replacement)
	}
	result := ""
	lastEnd := 0
	matches := re.FindAllStringSubmatchIndex(s, -1)
	replaced := 0
	for _, loc := range matches {
		if replaced >= n {
			break
		}
		result += s[lastEnd:loc[0]] + replacement
		lastEnd = loc[1]
		replaced++
	}
	result += s[lastEnd:]
	return result
}

// PatternSplit splits string by pattern (exported for VM use)
func PatternSplit(rp *objects.RegexPattern, s string, maxsplit int) objects.Object {
	re := rp.GetRegex().(*regexp.Regexp)

	if maxsplit == 0 {
		maxsplit = -1
	}

	indices := re.FindAllStringSubmatchIndex(s, maxsplit)
	if len(indices) == 0 {
		return &objects.List{Elements: []objects.Object{&objects.String{Value: s}}}
	}

	elements := make([]objects.Object, 0)
	lastEnd := 0
	numGroups := re.NumSubexp()

	for _, loc := range indices {
		// Add the text before the match
		elements = append(elements, &objects.String{Value: s[lastEnd:loc[0]]})

		// Add captured groups (like Python does)
		if numGroups > 0 {
			for i := 1; i < len(loc)/2; i++ {
				start := loc[i*2]
				end := loc[i*2+1]
				if start >= 0 && end >= 0 {
					elements = append(elements, &objects.String{Value: s[start:end]})
				} else {
					elements = append(elements, objects.None_)
				}
			}
		}

		lastEnd = loc[1]
	}

	// Add remaining text
	elements = append(elements, &objects.String{Value: s[lastEnd:]})

	return &objects.List{Elements: elements}
}

// CreateReModule creates the re module
func CreateReModule() *objects.Module {
	module := &objects.Module{
		Name:   "re",
		Fields: make(map[string]objects.Object),
	}

	// Flags
	module.Fields["IGNORECASE"] = &objects.Integer{Value: IGNORECASE}
	module.Fields["I"] = &objects.Integer{Value: IGNORECASE}
	module.Fields["MULTILINE"] = &objects.Integer{Value: MULTILINE}
	module.Fields["M"] = &objects.Integer{Value: MULTILINE}
	module.Fields["DOTALL"] = &objects.Integer{Value: DOTALL}
	module.Fields["S"] = &objects.Integer{Value: DOTALL}

	// re.compile(pattern, flags=0)
	module.Fields["compile"] = &objects.Builtin{
		Name: "re.compile",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("compile() takes at least 1 argument")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("compile() first argument must be a string")
			}
			flags := 0
			if len(args) >= 2 {
				if f, ok := args[1].(*objects.Integer); ok {
					flags = int(f.Value)
				}
			}
			re, err := compilePattern(pattern.Value, flags)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			return objects.NewRegexPattern(pattern.Value, flags, re)
		},
	}

	// re.match(pattern, string, flags=0)
	module.Fields["match"] = &objects.Builtin{
		Name: "re.match",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("match() takes at least 2 arguments")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("match() first argument must be a string")
			}
			s, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("match() second argument must be a string")
			}
			flags := 0
			if len(args) >= 3 {
				if f, ok := args[2].(*objects.Integer); ok {
					flags = int(f.Value)
				}
			}
			re, err := compilePattern(pattern.Value, flags)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			rp := objects.NewRegexPattern(pattern.Value, flags, re)
			return PatternMatch(rp, s.Value, false)
		},
	}

	// re.search(pattern, string, flags=0)
	module.Fields["search"] = &objects.Builtin{
		Name: "re.search",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("search() takes at least 2 arguments")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("search() first argument must be a string")
			}
			s, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("search() second argument must be a string")
			}
			flags := 0
			if len(args) >= 3 {
				if f, ok := args[2].(*objects.Integer); ok {
					flags = int(f.Value)
				}
			}
			re, err := compilePattern(pattern.Value, flags)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			rp := objects.NewRegexPattern(pattern.Value, flags, re)
			return PatternMatch(rp, s.Value, true)
		},
	}

	// re.findall(pattern, string, flags=0)
	module.Fields["findall"] = &objects.Builtin{
		Name: "re.findall",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("findall() takes at least 2 arguments")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("findall() first argument must be a string")
			}
			s, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("findall() second argument must be a string")
			}
			flags := 0
			if len(args) >= 3 {
				if f, ok := args[2].(*objects.Integer); ok {
					flags = int(f.Value)
				}
			}
			re, err := compilePattern(pattern.Value, flags)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			rp := objects.NewRegexPattern(pattern.Value, flags, re)
			return PatternFindall(rp, s.Value)
		},
	}

	// re.finditer(pattern, string, flags=0)
	module.Fields["finditer"] = &objects.Builtin{
		Name: "re.finditer",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("finditer() takes at least 2 arguments")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("finditer() first argument must be a string")
			}
			s, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("finditer() second argument must be a string")
			}
			flags := 0
			if len(args) >= 3 {
				if f, ok := args[2].(*objects.Integer); ok {
					flags = int(f.Value)
				}
			}
			re, err := compilePattern(pattern.Value, flags)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			rp := objects.NewRegexPattern(pattern.Value, flags, re)
			return patternFinditer(rp, s.Value)
		},
	}

	// re.sub(pattern, replacement, string, count=0)
	module.Fields["sub"] = &objects.Builtin{
		Name: "re.sub",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 3 {
				return objects.NewTypeError("sub() takes at least 3 arguments")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("sub() first argument must be a string")
			}
			replacement, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("sub() second argument must be a string")
			}
			s, ok := args[2].(*objects.String)
			if !ok {
				return objects.NewTypeError("sub() third argument must be a string")
			}
			count := 0
			if len(args) >= 4 {
				if c, ok := args[3].(*objects.Integer); ok {
					count = int(c.Value)
				}
			}
			re, err := compilePattern(pattern.Value, 0)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			rp := objects.NewRegexPattern(pattern.Value, 0, re)
			return PatternSub(rp, replacement.Value, s.Value, count)
		},
	}

	// re.split(pattern, string, maxsplit=0)
	module.Fields["split"] = &objects.Builtin{
		Name: "re.split",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("split() takes at least 2 arguments")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("split() first argument must be a string")
			}
			s, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("split() second argument must be a string")
			}
			maxsplit := 0
			if len(args) >= 3 {
				if m, ok := args[2].(*objects.Integer); ok {
					maxsplit = int(m.Value)
				}
			}
			re, err := compilePattern(pattern.Value, 0)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			rp := objects.NewRegexPattern(pattern.Value, 0, re)
			return PatternSplit(rp, s.Value, maxsplit)
		},
	}

	// re.fullmatch(pattern, string, flags=0)
	module.Fields["fullmatch"] = &objects.Builtin{
		Name: "re.fullmatch",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("fullmatch() takes at least 2 arguments")
			}
			pattern, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("fullmatch() first argument must be a string")
			}
			s, ok := args[1].(*objects.String)
			if !ok {
				return objects.NewTypeError("fullmatch() second argument must be a string")
			}
			flags := 0
			if len(args) >= 3 {
				if f, ok := args[2].(*objects.Integer); ok {
					flags = int(f.Value)
				}
			}
			re, err := compilePattern(pattern.Value, flags)
			if err != nil {
				return objects.NewError("invalid regular expression: %s", err.Error())
			}
			rp := objects.NewRegexPattern(pattern.Value, flags, re)
			return PatternFullmatch(rp, s.Value)
		},
	}

	return module
}
