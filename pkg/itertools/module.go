package itertools

import (
	"fmt"

	"github.com/go-py/go-python/pkg/objects"
)

// toSlice converts an iterable object to a slice of Objects.
func toSlice(iterable objects.Object) ([]objects.Object, *objects.Error) {
	switch v := iterable.(type) {
	case *objects.List:
		return v.Elements, nil
	case *objects.Tuple:
		return v.Elements, nil
	case *objects.Range:
		return v.ToList(), nil
	case *objects.Set:
		return v.ToSlice(), nil
	case *objects.String:
		result := make([]objects.Object, len(v.Value))
		for i, r := range v.Value {
			result[i] = &objects.String{Value: string(r)}
		}
		return result, nil
	case *objects.Dict:
		return v.KeysSlice(), nil
	case *objects.DictKeys:
		return v.ToList(), nil
	case *objects.DictValues:
		return v.ToList(), nil
	case *objects.DictItems:
		return v.ToList(), nil
	default:
		return nil, objects.NewTypeError("'%s' object is not iterable", iterable.Type())
	}
}

// toInt converts an Object to int64, returning an error if not possible.
func toInt(obj objects.Object, name string) (int64, *objects.Error) {
	switch v := obj.(type) {
	case *objects.Integer:
		return v.Value, nil
	case *objects.Float:
		return int64(v.Value), nil
	default:
		return 0, objects.NewTypeError("'%s' argument must be an integer, not '%s'", name, obj.Type())
	}
}

// ---------------------------------------------------------------
// Iterator types for lazy evaluation
// ---------------------------------------------------------------

// chainIterator iterates over multiple iterables sequentially
type chainIterator struct {
	iterables [][]objects.Object
	idx       int // current iterable index
	elemIdx   int // current element index within current iterable
}

func (ci *chainIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ci *chainIterator) Inspect() string          { return "<itertools.chain iterator>" }

func (ci *chainIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "chain.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				for ci.idx < len(ci.iterables) {
					if ci.elemIdx < len(ci.iterables[ci.idx]) {
						val := ci.iterables[ci.idx][ci.elemIdx]
						ci.elemIdx++
						return val
					}
					ci.idx++
					ci.elemIdx = 0
				}
				return objects.NewStopIteration("")
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "chain.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ci
			},
		}, true
	}
	return nil, false
}

// countIterator is an infinite counter
type countIterator struct {
	start int64
	step  int64
	cur   int64
}

func (ci *countIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ci *countIterator) Inspect() string          { return "<itertools.count iterator>" }

func (ci *countIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "count.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				val := ci.cur
				ci.cur += ci.step
				return &objects.Integer{Value: val}
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "count.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ci
			},
		}, true
	}
	return nil, false
}

// cycleIterator cycles through an iterable infinitely
type cycleIterator struct {
	elements []objects.Object
	idx      int
}

func (ci *cycleIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ci *cycleIterator) Inspect() string          { return "<itertools.cycle iterator>" }

func (ci *cycleIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "cycle.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(ci.elements) == 0 {
					return objects.NewStopIteration("")
				}
				val := ci.elements[ci.idx]
				ci.idx = (ci.idx + 1) % len(ci.elements)
				return val
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "cycle.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ci
			},
		}, true
	}
	return nil, false
}

// isliceIterator selects elements from an iterable by indices
type isliceIterator struct {
	elements []objects.Object
	start    int64
	stop     int64
	step     int64
	cur      int64 // logical index into the original iterable
}

func (ii *isliceIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ii *isliceIterator) Inspect() string          { return "<itertools.islice iterator>" }

func (ii *isliceIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "islice.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				for ii.cur < ii.stop && int(ii.cur) < len(ii.elements) {
					if ii.cur >= ii.start && (ii.cur-ii.start)%ii.step == 0 {
						val := ii.elements[ii.cur]
						ii.cur++
						return val
					}
					ii.cur++
				}
				return objects.NewStopIteration("")
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "islice.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ii
			},
		}, true
	}
	return nil, false
}

// repeatIterator repeats an element n times (or infinitely if n < 0)
type repeatIterator struct {
	elem   objects.Object
	n      int64 // -1 means infinite
	count  int64
}

func (ri *repeatIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ri *repeatIterator) Inspect() string {
	if ri.n < 0 {
		return fmt.Sprintf("repeat(%s)", ri.elem.Inspect())
	}
	return fmt.Sprintf("repeat(%s, %d)", ri.elem.Inspect(), ri.n)
}

func (ri *repeatIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "repeat.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if ri.n >= 0 && ri.count >= ri.n {
					return objects.NewStopIteration("")
				}
				ri.count++
				return ri.elem
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "repeat.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ri
			},
		}, true
	case "__length_hint__":
		return &objects.Builtin{
			Name: "repeat.__length_hint__",
			Fn: func(args ...objects.Object) objects.Object {
				if ri.n < 0 {
					return &objects.Integer{Value: 0}
				}
				remaining := ri.n - ri.count
				if remaining < 0 {
					remaining = 0
				}
				return &objects.Integer{Value: remaining}
			},
		}, true
	}
	return nil, false
}

// accumulateIterator tracks the accumulated value across iterations
type accumulateIterator struct {
	elements []objects.Object
	func_    objects.Object
	acc      objects.Object
	idx      int
	started  bool
}

func (ai *accumulateIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ai *accumulateIterator) Inspect() string          { return "<itertools.accumulate iterator>" }

func (ai *accumulateIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "accumulate.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if !ai.started {
					ai.started = true
					if len(ai.elements) == 0 {
						return objects.NewStopIteration("")
					}
					ai.acc = ai.elements[0]
					ai.idx = 1
					return ai.acc
				}
				if ai.idx >= len(ai.elements) {
					return objects.NewStopIteration("")
				}
				if ai.func_ == nil {
					// Default: addition
					accInt, accOk := ai.acc.(*objects.Integer)
					elemInt, elemOk := ai.elements[ai.idx].(*objects.Integer)
					if accOk && elemOk {
						ai.acc = &objects.Integer{Value: accInt.Value + elemInt.Value}
					} else {
						accFloat, accFOk := ai.acc.(*objects.Float)
						elemFloat, elemFOk := ai.elements[ai.idx].(*objects.Float)
						if accFOk && elemFOk {
							ai.acc = &objects.Float{Value: accFloat.Value + elemFloat.Value}
						} else if accOk && elemFOk {
							ai.acc = &objects.Float{Value: float64(accInt.Value) + elemFloat.Value}
						} else if accFOk && elemOk {
							ai.acc = &objects.Float{Value: accFloat.Value + float64(elemInt.Value)}
						} else {
							return objects.NewTypeError("unsupported operand type(s) for +")
						}
					}
				} else {
					result := objects.CallFunction(ai.func_, ai.acc, ai.elements[ai.idx])
					if result.Type() == objects.ERROR_OBJ {
						return result
					}
					ai.acc = result
				}
				ai.idx++
				return ai.acc
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "accumulate.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ai
			},
		}, true
	}
	return nil, false
}

// starmapIterator applies function using argument tuples
type starmapIterator struct {
	elements []objects.Object
	func_    objects.Object
	idx      int
}

func (si *starmapIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (si *starmapIterator) Inspect() string          { return "<itertools.starmap iterator>" }

func (si *starmapIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "starmap.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if si.idx >= len(si.elements) {
					return objects.NewStopIteration("")
				}
				elem := si.elements[si.idx]
				si.idx++
				// elem should be an iterable (tuple/list)
				argsSlice, err := toSlice(elem)
				if err != nil {
					return err
				}
				return objects.CallFunction(si.func_, argsSlice...)
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "starmap.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return si
			},
		}, true
	}
	return nil, false
}

// filterfalseIterator yields items where predicate is false
type filterfalseIterator struct {
	elements  []objects.Object
	predicate objects.Object
	idx       int
}

func (fi *filterfalseIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (fi *filterfalseIterator) Inspect() string          { return "<itertools.filterfalse iterator>" }

func (fi *filterfalseIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "filterfalse.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				for fi.idx < len(fi.elements) {
					elem := fi.elements[fi.idx]
					fi.idx++
					if fi.predicate == nil {
						// None predicate: return falsy items
						if !isTruthy(elem) {
							return elem
						}
					} else {
						result := objects.CallFunction(fi.predicate, elem)
						if result.Type() == objects.ERROR_OBJ {
							return result
						}
						if !isTruthy(result) {
							return elem
						}
					}
				}
				return objects.NewStopIteration("")
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "filterfalse.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return fi
			},
		}, true
	}
	return nil, false
}

// zipLongestIterator zips iterables, filling missing values
type zipLongestIterator struct {
	iterables [][]objects.Object
	fillvalue objects.Object
	idx       int
}

func (zi *zipLongestIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (zi *zipLongestIterator) Inspect() string          { return "<itertools.zip_longest iterator>" }

func (zi *zipLongestIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "zip_longest.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				allDone := true
				for _, iter := range zi.iterables {
					if zi.idx < len(iter) {
						allDone = false
						break
					}
				}
				if allDone {
					return objects.NewStopIteration("")
				}
				tuple := make([]objects.Object, len(zi.iterables))
				for i, iter := range zi.iterables {
					if zi.idx < len(iter) {
						tuple[i] = iter[zi.idx]
					} else {
						tuple[i] = zi.fillvalue
					}
				}
				zi.idx++
				return &objects.Tuple{Elements: tuple}
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "zip_longest.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return zi
			},
		}, true
	}
	return nil, false
}

// pairwiseIterator yields successive overlapping pairs
type pairwiseIterator struct {
	elements []objects.Object
	idx      int
}

func (pi *pairwiseIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (pi *pairwiseIterator) Inspect() string          { return "<itertools.pairwise iterator>" }

func (pi *pairwiseIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "pairwise.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if pi.idx+1 >= len(pi.elements) {
					return objects.NewStopIteration("")
				}
				pair := &objects.Tuple{Elements: []objects.Object{
					pi.elements[pi.idx],
					pi.elements[pi.idx+1],
				}}
				pi.idx++
				return pair
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "pairwise.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return pi
			},
		}, true
	}
	return nil, false
}

// groupbyIterator groups consecutive elements by key
type groupbyIterator struct {
	elements []objects.Object
	keyFunc  objects.Object // nil means identity
	idx      int
}

func (gi *groupbyIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (gi *groupbyIterator) Inspect() string          { return "<itertools.groupby iterator>" }

func (gi *groupbyIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "groupby.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if gi.idx >= len(gi.elements) {
					return objects.NewStopIteration("")
				}
				// Get the key for the current element
				currentElem := gi.elements[gi.idx]
				var currentKey objects.Object
				if gi.keyFunc == nil {
					currentKey = currentElem
				} else {
					currentKey = objects.CallFunction(gi.keyFunc, currentElem)
					if currentKey.Type() == objects.ERROR_OBJ {
						return currentKey
					}
				}
				// Collect all consecutive elements with the same key
				group := make([]objects.Object, 0)
				for gi.idx < len(gi.elements) {
					elem := gi.elements[gi.idx]
					var key objects.Object
					if gi.keyFunc == nil {
						key = elem
					} else {
						key = objects.CallFunction(gi.keyFunc, elem)
						if key.Type() == objects.ERROR_OBJ {
							return key
						}
					}
					if !objects.Equal(key, currentKey) {
						break
					}
					group = append(group, elem)
					gi.idx++
				}
				return &objects.Tuple{Elements: []objects.Object{
					currentKey,
					&objects.List{Elements: group},
				}}
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "groupby.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return gi
			},
		}, true
	}
	return nil, false
}

// teeIterator is one of n independent iterators from a single iterable
type teeIterator struct {
	elements []objects.Object
	idx      int
}

func (ti *teeIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ti *teeIterator) Inspect() string          { return "<itertools.tee iterator>" }

func (ti *teeIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "tee.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if ti.idx >= len(ti.elements) {
					return objects.NewStopIteration("")
				}
				val := ti.elements[ti.idx]
				ti.idx++
				return val
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "tee.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ti
			},
		}, true
	}
	return nil, false
}

// isTruthy checks if an object is truthy in Python semantics
func isTruthy(obj objects.Object) bool {
	switch v := obj.(type) {
	case *objects.Boolean:
		return v.Value
	case *objects.Integer:
		return v.Value != 0
	case *objects.Float:
		return v.Value != 0.0
	case *objects.String:
		return len(v.Value) != 0
	case *objects.List:
		return len(v.Elements) != 0
	case *objects.Tuple:
		return len(v.Elements) != 0
	case *objects.Dict:
		return v.Size() != 0
	case *objects.Set:
		return v.Size() != 0
	case *objects.None:
		return false
	default:
		return true
	}
}

// ---------------------------------------------------------------
// Combinatoric iterator types
// ---------------------------------------------------------------

// permutationsIterator yields r-length permutations
type permutationsIterator struct {
	elements []objects.Object
	r        int
	indices  []int
	done     bool
	started  bool
}

func newPermutationsIterator(elements []objects.Object, r int) *permutationsIterator {
	n := len(elements)
	pi := &permutationsIterator{
		elements: elements,
		r:        r,
	}
	if n == 0 || r > n {
		pi.done = true
		return pi
	}
	pi.indices = make([]int, n)
	for i := range pi.indices {
		pi.indices[i] = i
	}
	return pi
}

func (pi *permutationsIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (pi *permutationsIterator) Inspect() string          { return "<itertools.permutations iterator>" }

func (pi *permutationsIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "permutations.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if pi.done {
					return objects.NewStopIteration("")
				}
				if !pi.started {
					pi.started = true
					return pi.current()
				}
				// Advance to next permutation using next lexicographic permutation of indices
				// We only permute the first r elements (swap-based approach)
				// Use the standard algorithm for generating permutations of indices[0:r]
				if !pi.nextPermutation() {
					pi.done = true
					return objects.NewStopIteration("")
				}
				return pi.current()
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "permutations.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return pi
			},
		}, true
	}
	return nil, false
}

func (pi *permutationsIterator) current() objects.Object {
	result := make([]objects.Object, pi.r)
	for i := 0; i < pi.r; i++ {
		result[i] = pi.elements[pi.indices[i]]
	}
	return &objects.Tuple{Elements: result}
}

// nextPermutation generates the next permutation using Heap's algorithm variant
// We use the standard lexicographic next_permutation on indices
func (pi *permutationsIterator) nextPermutation() bool {
	n := len(pi.indices)
	// Find the largest index k such that indices[k] < indices[k+1]
	k := -1
	for i := n - 2; i >= 0; i-- {
		if pi.indices[i] < pi.indices[i+1] {
			k = i
			break
		}
	}
	if k == -1 {
		return false
	}
	// Find the largest index l > k such that indices[k] < indices[l]
	l := -1
	for i := n - 1; i > k; i-- {
		if pi.indices[k] < pi.indices[i] {
			l = i
			break
		}
	}
	// Swap indices[k] and indices[l]
	pi.indices[k], pi.indices[l] = pi.indices[l], pi.indices[k]
	// Reverse indices[k+1:]
	for i, j := k+1, n-1; i < j; i, j = i+1, j-1 {
		pi.indices[i], pi.indices[j] = pi.indices[j], pi.indices[i]
	}
	return true
}

// combinationsIterator yields r-length combinations
type combinationsIterator struct {
	elements []objects.Object
	r        int
	indices  []int
	done     bool
	started  bool
}

func newCombinationsIterator(elements []objects.Object, r int) *combinationsIterator {
	n := len(elements)
	ci := &combinationsIterator{
		elements: elements,
		r:        r,
	}
	if n == 0 || r > n {
		ci.done = true
		return ci
	}
	ci.indices = make([]int, r)
	for i := range ci.indices {
		ci.indices[i] = i
	}
	return ci
}

func (ci *combinationsIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (ci *combinationsIterator) Inspect() string          { return "<itertools.combinations iterator>" }

func (ci *combinationsIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "combinations.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if ci.done {
					return objects.NewStopIteration("")
				}
				if !ci.started {
					ci.started = true
					return ci.current()
				}
				if !ci.nextCombination() {
					ci.done = true
					return objects.NewStopIteration("")
				}
				return ci.current()
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "combinations.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return ci
			},
		}, true
	}
	return nil, false
}

func (ci *combinationsIterator) current() objects.Object {
	result := make([]objects.Object, ci.r)
	for i := 0; i < ci.r; i++ {
		result[i] = ci.elements[ci.indices[i]]
	}
	return &objects.Tuple{Elements: result}
}

func (ci *combinationsIterator) nextCombination() bool {
	n := len(ci.elements)
	r := ci.r
	// Find the rightmost index that can be incremented
	i := r - 1
	for i >= 0 {
		maxVal := n - r + i
		if ci.indices[i] < maxVal {
			ci.indices[i]++
			// Reset all indices to the right
			for j := i + 1; j < r; j++ {
				ci.indices[j] = ci.indices[j-1] + 1
			}
			return true
		}
		i--
	}
	return false
}

// productIterator yields the Cartesian product
type productIterator struct {
	pools   [][]objects.Object
	repeat  int
	indices []int
	done    bool
	started bool
}

func newProductIterator(pools [][]objects.Object, repeat int) *productIterator {
	// Repeat the pools
	actualPools := make([][]objects.Object, 0, len(pools)*repeat)
	for i := 0; i < repeat; i++ {
		actualPools = append(actualPools, pools...)
	}
	pi := &productIterator{
		pools:  actualPools,
		repeat: repeat,
	}
	// Check for empty pools
	for _, pool := range actualPools {
		if len(pool) == 0 {
			pi.done = true
			return pi
		}
	}
	pi.indices = make([]int, len(actualPools))
	return pi
}

func (pi *productIterator) Type() objects.ObjectType { return objects.GENERATOR_OBJ }
func (pi *productIterator) Inspect() string          { return "<itertools.product iterator>" }

func (pi *productIterator) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__next__":
		return &objects.Builtin{
			Name: "product.__next__",
			Fn: func(args ...objects.Object) objects.Object {
				if pi.done {
					return objects.NewStopIteration("")
				}
				if !pi.started {
					pi.started = true
					return pi.current()
				}
				if !pi.nextProduct() {
					pi.done = true
					return objects.NewStopIteration("")
				}
				return pi.current()
			},
		}, true
	case "__iter__":
		return &objects.Builtin{
			Name: "product.__iter__",
			Fn: func(args ...objects.Object) objects.Object {
				return pi
			},
		}, true
	}
	return nil, false
}

func (pi *productIterator) current() objects.Object {
	result := make([]objects.Object, len(pi.pools))
	for i, pool := range pi.pools {
		result[i] = pool[pi.indices[i]]
	}
	return &objects.Tuple{Elements: result}
}

func (pi *productIterator) nextProduct() bool {
	for i := len(pi.pools) - 1; i >= 0; i-- {
		pi.indices[i]++
		if pi.indices[i] < len(pi.pools[i]) {
			return true
		}
		pi.indices[i] = 0
	}
	return false
}

// ---------------------------------------------------------------
// CreateItertoolsModule creates the itertools module
// ---------------------------------------------------------------

func CreateItertoolsModule() *objects.Module {
	module := &objects.Module{
		Name:   "itertools",
		Fields: make(map[string]objects.Object),
	}

	// itertools.chain(*iterables)
	module.Fields["chain"] = &objects.Builtin{
		Name: "itertools.chain",
		Fn: func(args ...objects.Object) objects.Object {
			iterables := make([][]objects.Object, 0, len(args))
			for _, arg := range args {
				elems, err := toSlice(arg)
				if err != nil {
					return err
				}
				iterables = append(iterables, elems)
			}
			return &chainIterator{iterables: iterables}
		},
	}

	// itertools.count(start=0, step=1)
	module.Fields["count"] = &objects.Builtin{
		Name: "itertools.count",
		Fn: func(args ...objects.Object) objects.Object {
			start := int64(0)
			step := int64(1)
			if len(args) >= 1 {
				s, err := toInt(args[0], "count")
				if err != nil {
					return err
				}
				start = s
			}
			if len(args) >= 2 {
				s, err := toInt(args[1], "count")
				if err != nil {
					return err
				}
				step = s
			}
			return &countIterator{start: start, step: step, cur: start}
		},
	}

	// itertools.cycle(iterable)
	module.Fields["cycle"] = &objects.Builtin{
		Name: "itertools.cycle",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("cycle() takes at least 1 argument")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}
			if len(elems) == 0 {
				return objects.NewValueError("cycle() argument must not be empty")
			}
			return &cycleIterator{elements: elems}
		},
	}

	// itertools.islice(iterable, stop) / islice(iterable, start, stop[, step])
	module.Fields["islice"] = &objects.Builtin{
		Name: "itertools.islice",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("islice() takes at least 2 arguments")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}

			var start, stop, step int64
			if len(args) == 2 {
				// islice(iterable, stop)
				start = 0
				s, err := toInt(args[1], "islice")
				if err != nil {
					return err
				}
				stop = s
				step = 1
			} else if len(args) == 3 {
				// islice(iterable, start, stop)
				s, err := toInt(args[1], "islice")
				if err != nil {
					return err
				}
				start = s
				s, err = toInt(args[2], "islice")
				if err != nil {
					return err
				}
				stop = s
				step = 1
			} else {
				// islice(iterable, start, stop, step)
				s, err := toInt(args[1], "islice")
				if err != nil {
					return err
				}
				start = s
				s, err = toInt(args[2], "islice")
				if err != nil {
					return err
				}
				stop = s
				s, err = toInt(args[3], "islice")
				if err != nil {
					return err
				}
				step = s
			}

			if step == 0 {
				return objects.NewValueError("islice() step argument must not be zero")
			}

			// Handle negative values and None-like behavior
			n := int64(len(elems))
			if start < 0 {
				start = 0
			}
			if stop < 0 {
				stop = 0
			}
			if start > n {
				start = n
			}
			if stop > n {
				stop = n
			}

			// Pre-compute the selected elements for simplicity
			result := make([]objects.Object, 0)
			if step > 0 {
				for i := start; i < stop; i += step {
					if int(i) < len(elems) {
						result = append(result, elems[i])
					}
				}
			} else {
				for i := start; i > stop; i += step {
					if int(i) < len(elems) && i >= 0 {
						result = append(result, elems[i])
					}
				}
			}

			return &isliceIterator{elements: result, start: 0, stop: int64(len(result)), step: 1, cur: 0}
		},
	}

	// itertools.repeat(elem[, n])
	module.Fields["repeat"] = &objects.Builtin{
		Name: "itertools.repeat",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("repeat() takes at least 1 argument")
			}
			elem := args[0]
			n := int64(-1) // -1 means infinite
			if len(args) >= 2 {
				times, err := toInt(args[1], "repeat")
				if err != nil {
					return err
				}
				n = times
			}
			return &repeatIterator{elem: elem, n: n}
		},
	}

	// itertools.accumulate(iterable[, func])
	module.Fields["accumulate"] = &objects.Builtin{
		Name: "itertools.accumulate",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("accumulate() takes at least 1 argument")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}
			var func_ objects.Object
			if len(args) >= 2 {
				func_ = args[1]
			}
			return &accumulateIterator{
				elements: elems,
				func_:    func_,
			}
		},
	}

	// itertools.product(*iterables[, repeat=1])
	module.Fields["product"] = &objects.Builtin{
		Name: "itertools.product",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("product() takes at least 1 argument")
			}

			// Check if last argument is a keyword-like dict for repeat
			repeat := 1
			iterableArgs := args

			// Simple heuristic: if last arg is an integer and there are multiple args,
			// it might be repeat. But in Python, repeat is keyword-only.
			// Since we don't have proper keyword argument support here,
			// we'll check if the last arg is a Dict with a "repeat" key.
			if len(args) >= 2 {
				if last, ok := args[len(args)-1].(*objects.Dict); ok {
					if val, ok := last.Get(&objects.String{Value: "repeat"}); ok {
						if r, ok := val.(*objects.Integer); ok {
							repeat = int(r.Value)
						}
					}
					iterableArgs = args[:len(args)-1]
				}
			}

			pools := make([][]objects.Object, 0, len(iterableArgs))
			for _, arg := range iterableArgs {
				elems, err := toSlice(arg)
				if err != nil {
					return err
				}
				pools = append(pools, elems)
			}

			return newProductIterator(pools, repeat)
		},
	}

	// itertools.permutations(iterable[, r])
	module.Fields["permutations"] = &objects.Builtin{
		Name: "itertools.permutations",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("permutations() takes at least 1 argument")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}
			r := len(elems)
			if len(args) >= 2 {
				rVal, err := toInt(args[1], "permutations")
				if err != nil {
					return err
				}
				r = int(rVal)
			}
			if r < 0 {
				r = 0
			}
			return newPermutationsIterator(elems, r)
		},
	}

	// itertools.combinations(iterable, r)
	module.Fields["combinations"] = &objects.Builtin{
		Name: "itertools.combinations",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("combinations() takes at least 2 arguments")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}
			rVal, err := toInt(args[1], "combinations")
			if err != nil {
				return err
			}
			r := int(rVal)
			if r < 0 {
				return objects.NewValueError("r must be non-negative")
			}
			return newCombinationsIterator(elems, r)
		},
	}

	// itertools.groupby(iterable[, key])
	module.Fields["groupby"] = &objects.Builtin{
		Name: "itertools.groupby",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("groupby() takes at least 1 argument")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}
			var keyFunc objects.Object
			if len(args) >= 2 {
				keyFunc = args[1]
			}
			return &groupbyIterator{elements: elems, keyFunc: keyFunc}
		},
	}

	// itertools.starmap(function, iterable)
	module.Fields["starmap"] = &objects.Builtin{
		Name: "itertools.starmap",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("starmap() takes at least 2 arguments")
			}
			func_ := args[0]
			elems, err := toSlice(args[1])
			if err != nil {
				return err
			}
			return &starmapIterator{elements: elems, func_: func_}
		},
	}

	// itertools.filterfalse(predicate, iterable)
	module.Fields["filterfalse"] = &objects.Builtin{
		Name: "itertools.filterfalse",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("filterfalse() takes at least 2 arguments")
			}
			predicate := args[0]
			if _, ok := predicate.(*objects.None); ok {
				predicate = nil
			}
			elems, err := toSlice(args[1])
			if err != nil {
				return err
			}
			return &filterfalseIterator{elements: elems, predicate: predicate}
		},
	}

	// itertools.zip_longest(*iterables[, fillvalue=None])
	module.Fields["zip_longest"] = &objects.Builtin{
		Name: "itertools.zip_longest",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("zip_longest() takes at least 1 argument")
			}

			fillvalue := objects.Object(objects.None_)
			iterableArgs := args

			// Check if last arg is a Dict with fillvalue key
			if len(args) >= 2 {
				if last, ok := args[len(args)-1].(*objects.Dict); ok {
					if val, ok := last.Get(&objects.String{Value: "fillvalue"}); ok {
						fillvalue = val
					}
					iterableArgs = args[:len(args)-1]
				}
			}

			iterables := make([][]objects.Object, 0, len(iterableArgs))
			for _, arg := range iterableArgs {
				elems, err := toSlice(arg)
				if err != nil {
					return err
				}
				iterables = append(iterables, elems)
			}

			return &zipLongestIterator{iterables: iterables, fillvalue: fillvalue}
		},
	}

	// itertools.tee(iterable, n=2)
	module.Fields["tee"] = &objects.Builtin{
		Name: "itertools.tee",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("tee() takes at least 1 argument")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}
			n := int64(2)
			if len(args) >= 2 {
				nVal, err := toInt(args[1], "tee")
				if err != nil {
					return err
				}
				n = nVal
			}
			if n < 1 {
				return objects.NewValueError("n must be at least 1")
			}

			iterators := make([]objects.Object, n)
			for i := int64(0); i < n; i++ {
				// Each iterator gets its own copy of the elements
				elemsCopy := make([]objects.Object, len(elems))
				copy(elemsCopy, elems)
				iterators[i] = &teeIterator{elements: elemsCopy}
			}
			return &objects.Tuple{Elements: iterators}
		},
	}

	// itertools.pairwise(iterable)
	module.Fields["pairwise"] = &objects.Builtin{
		Name: "itertools.pairwise",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("pairwise() takes at least 1 argument")
			}
			elems, err := toSlice(args[0])
			if err != nil {
				return err
			}
			return &pairwiseIterator{elements: elems}
		},
	}

	return module
}
