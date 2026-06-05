package gc

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/objects"
)

type GarbageCollector struct {
	mu              sync.Mutex
	objects         []*GCObject      // all objects
	youngObjects    []*GCObject      // generation 0
	oldObjects      []*GCObject      // generation 1
	marked          map[*GCObject]bool
	rememberedSet   []*GCObject      // old objects that reference young objects
	allocatedBytes  int64
	freedBytes      int64
	collectionCount int64
	minorCount      int64
	majorCount      int64
	enabled         bool
	verbose         bool
	generational    bool             // enable generational mode
	threshold       int64
	minorThreshold  int64            // threshold for minor collection (lower than major)
	promotionAge    int              // survive this many minor GCs to be promoted
	pauseTime       time.Duration
}

type GCObject struct {
	Object      objects.Object
	Marked      bool
	Size        int64
	Generation  int              // 0 = young, 1 = old
	Survived    int              // number of minor collections survived
	Next        *GCObject
	Prev        *GCObject
	Finalizer   func(objects.Object)
}

var singleton *GarbageCollector
var once sync.Once

func GetGC() *GarbageCollector {
	once.Do(func() {
		singleton = &GarbageCollector{
			objects:        make([]*GCObject, 0),
			youngObjects:   make([]*GCObject, 0),
			oldObjects:     make([]*GCObject, 0),
			marked:         make(map[*GCObject]bool),
			rememberedSet:  make([]*GCObject, 0),
			enabled:        true,
			verbose:        false,
			generational:   true,
			threshold:      1024 * 1024,     // 1MB threshold
			minorThreshold: 256 * 1024,       // 256KB minor threshold
			promotionAge:   1,                // promote after surviving 1 minor GC
		}
	})
	return singleton
}

func (gc *GarbageCollector) Enable(enable bool) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.enabled = enable
}

func (gc *GarbageCollector) SetVerbose(verbose bool) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.verbose = verbose
}

func (gc *GarbageCollector) SetThreshold(threshold int64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.threshold = threshold
}

func (gc *GarbageCollector) Allocate(obj objects.Object) *GCObject {
	if !gc.enabled {
		return &GCObject{Object: obj, Size: estimateSize(obj)}
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	gcObj := &GCObject{
		Object:     obj,
		Marked:     false,
		Size:       estimateSize(obj),
		Generation: 0, // new objects go into young generation
		Survived:   0,
	}

	gc.objects = append(gc.objects, gcObj)
	gc.youngObjects = append(gc.youngObjects, gcObj)
	gc.allocatedBytes += gcObj.Size

	if gc.verbose {
		log.Printf("GC: Allocated %d bytes (young), total %d bytes", gcObj.Size, gc.allocatedBytes)
	}

	if gc.generational {
		if gc.allocatedBytes >= gc.threshold {
			go gc.MajorCollect()
		} else if gc.allocatedBytes >= gc.minorThreshold {
			go gc.MinorCollect()
		}
	} else {
		if gc.allocatedBytes >= gc.threshold {
			go gc.Collect()
		}
	}

	return gcObj
}

func (gc *GarbageCollector) Register(obj objects.Object) *GCObject {
	return gc.Allocate(obj)
}

func (gc *GarbageCollector) Collect() {
	if !gc.enabled {
		return
	}

	if gc.generational {
		// Default: do a minor collection most of the time
		// Every 5 minor collections, do a major collection instead
		if gc.minorCount > 0 && (gc.minorCount+1)%5 == 0 {
			gc.MajorCollect()
		} else {
			gc.MinorCollect()
		}
		return
	}

	// Non-generational: full mark-and-sweep
	gc.mu.Lock()
	start := time.Now()

	if gc.verbose {
		log.Println("GC: Starting collection")
	}

	gc.marked = make(map[*GCObject]bool)

	gc.markRoots()

	freed := gc.sweep()

	gc.collectionCount++
	gc.pauseTime += time.Since(start)

	if gc.verbose {
		log.Printf("GC: Collection completed, freed %d objects (%d bytes) in %v",
			freed, gc.freedBytes, time.Since(start))
	}

	gc.mu.Unlock()
}

func (gc *GarbageCollector) MinorCollect() {
	if !gc.enabled {
		return
	}

	gc.mu.Lock()
	start := time.Now()

	if gc.verbose {
		log.Printf("GC: Starting minor collection, young objects: %d", len(gc.youngObjects))
	}

	gc.marked = make(map[*GCObject]bool)

	// Mark from roots, but only mark young objects
	gc.markRootsForMinor()

	// Also mark from remembered set (old objects that reference young objects)
	for _, oldObj := range gc.rememberedSet {
		if oldObj.Object != nil {
			gc.markReferencesForMinor(oldObj.Object)
		}
	}

	// Sweep only young objects
	freed := gc.sweepYoung()

	gc.minorCount++
	gc.collectionCount++
	gc.pauseTime += time.Since(start)

	// Promote survivors
	gc.promoteSurvivors()

	// Clear remembered set (will be rebuilt as needed)
	gc.rememberedSet = gc.rememberedSet[:0]

	if gc.verbose {
		log.Printf("GC: Minor collection completed, freed %d objects in %v, young: %d, old: %d",
			freed, time.Since(start), len(gc.youngObjects), len(gc.oldObjects))
	}

	gc.mu.Unlock()
}

func (gc *GarbageCollector) MajorCollect() {
	if !gc.enabled {
		return
	}

	gc.mu.Lock()
	start := time.Now()

	if gc.verbose {
		log.Printf("GC: Starting major collection, total objects: %d", len(gc.objects))
	}

	gc.marked = make(map[*GCObject]bool)

	gc.markRoots()

	freed := gc.sweep()

	gc.majorCount++
	gc.collectionCount++
	gc.pauseTime += time.Since(start)

	// Rebuild generation lists after major collection
	gc.rebuildGenerationLists()

	// Clear remembered set
	gc.rememberedSet = gc.rememberedSet[:0]

	if gc.verbose {
		log.Printf("GC: Major collection completed, freed %d objects (%d bytes) in %v",
			freed, gc.freedBytes, time.Since(start))
	}

	gc.mu.Unlock()
}

func (gc *GarbageCollector) markRoots() {
	var roots []objects.Object

	for _, reg := range getRegisterContents() {
		if reg != nil {
			roots = append(roots, reg)
		}
	}

	roots = append(roots, getStackContents()...)
	roots = append(roots, getGlobalContents()...)
	roots = append(roots, getFrameContents()...)

	for _, root := range roots {
		gc.markObject(root)
	}
}

func (gc *GarbageCollector) markObject(obj objects.Object) {
	if obj == nil {
		return
	}

	for _, gcObj := range gc.objects {
		if gcObj.Object == obj && !gc.marked[gcObj] {
			gc.marked[gcObj] = true
			gc.markReferences(gcObj.Object)
			break
		}
	}
}

func (gc *GarbageCollector) markReferences(obj objects.Object) {
	switch o := obj.(type) {
	case *objects.List:
		for _, elem := range o.Elements {
			gc.markObject(elem)
		}
	case *objects.Tuple:
		for _, elem := range o.Elements {
			gc.markObject(elem)
		}
	case *objects.Dict:
		for keyStr, key := range o.Keys {
			gc.markObject(key)
			if val, ok := o.Pairs[keyStr]; ok {
				gc.markObject(val)
			}
		}
	case *objects.Set:
		for _, elem := range o.Elements {
			gc.markObject(elem)
		}
	case *objects.Instance:
		for _, field := range o.Fields {
			gc.markObject(field)
		}
	case *objects.Class:
		if o.SuperClass != nil {
			gc.markObject(o.SuperClass)
		}
	}
}

// markRootsForMinor marks young objects reachable from roots.
// It scans all roots but only marks objects in the young generation.
func (gc *GarbageCollector) markRootsForMinor() {
	var roots []objects.Object

	for _, reg := range getRegisterContents() {
		if reg != nil {
			roots = append(roots, reg)
		}
	}

	roots = append(roots, getStackContents()...)
	roots = append(roots, getGlobalContents()...)
	roots = append(roots, getFrameContents()...)

	for _, root := range roots {
		gc.markObjectForMinor(root)
	}
}

// markObjectForMinor marks a young-generation object as reachable.
// It traverses references but only marks objects in the young generation.
func (gc *GarbageCollector) markObjectForMinor(obj objects.Object) {
	if obj == nil {
		return
	}

	for _, gcObj := range gc.youngObjects {
		if gcObj.Object == obj && !gc.marked[gcObj] {
			gc.marked[gcObj] = true
			gc.markReferencesForMinor(gcObj.Object)
			break
		}
	}
}

// markReferencesForMinor follows references from an object, marking young objects.
// For old objects, we still traverse their references to find young objects,
// but we don't mark the old objects themselves.
func (gc *GarbageCollector) markReferencesForMinor(obj objects.Object) {
	switch o := obj.(type) {
	case *objects.List:
		for _, elem := range o.Elements {
			gc.markObjectForMinor(elem)
		}
	case *objects.Tuple:
		for _, elem := range o.Elements {
			gc.markObjectForMinor(elem)
		}
	case *objects.Dict:
		for keyStr, key := range o.Keys {
			gc.markObjectForMinor(key)
			if val, ok := o.Pairs[keyStr]; ok {
				gc.markObjectForMinor(val)
			}
		}
	case *objects.Set:
		for _, elem := range o.Elements {
			gc.markObjectForMinor(elem)
		}
	case *objects.Instance:
		for _, field := range o.Fields {
			gc.markObjectForMinor(field)
		}
	case *objects.Class:
		if o.SuperClass != nil {
			gc.markObjectForMinor(o.SuperClass)
		}
	}
}

// sweepYoung only sweeps young generation objects that are not marked.
func (gc *GarbageCollector) sweepYoung() int {
	freed := 0
	newYoung := make([]*GCObject, 0, len(gc.youngObjects))
	newAll := make([]*GCObject, 0, len(gc.objects))

	for _, gcObj := range gc.objects {
		if gcObj.Generation == 1 {
			// Old objects are always kept during minor collection
			newAll = append(newAll, gcObj)
			continue
		}
		// Young object: check if marked
		if gc.marked[gcObj] {
			gcObj.Marked = false
			newYoung = append(newYoung, gcObj)
			newAll = append(newAll, gcObj)
		} else {
			if gcObj.Finalizer != nil {
				gcObj.Finalizer(gcObj.Object)
			}
			gc.freedBytes += gcObj.Size
			freed++
			if gc.verbose {
				log.Printf("GC: Freed young %T (%d bytes)", gcObj.Object, gcObj.Size)
			}
		}
	}

	gc.youngObjects = newYoung
	gc.objects = newAll

	return freed
}

// promoteSurvivors moves young objects that have survived enough minor GCs to old generation.
func (gc *GarbageCollector) promoteSurvivors() {
	newYoung := make([]*GCObject, 0, len(gc.youngObjects))

	for _, gcObj := range gc.youngObjects {
		gcObj.Survived++
		if gcObj.Survived >= gc.promotionAge {
			gcObj.Generation = 1
			gc.oldObjects = append(gc.oldObjects, gcObj)
			if gc.verbose {
				log.Printf("GC: Promoted %T to old generation (survived %d minor GCs)", gcObj.Object, gcObj.Survived)
			}
		} else {
			newYoung = append(newYoung, gcObj)
		}
	}

	gc.youngObjects = newYoung
}

// rebuildGenerationLists reconstructs youngObjects and oldObjects from the main objects list.
func (gc *GarbageCollector) rebuildGenerationLists() {
	gc.youngObjects = make([]*GCObject, 0)
	gc.oldObjects = make([]*GCObject, 0)

	for _, gcObj := range gc.objects {
		if gcObj.Generation == 0 {
			gc.youngObjects = append(gc.youngObjects, gcObj)
		} else {
			gc.oldObjects = append(gc.oldObjects, gcObj)
		}
	}
}

// WriteBarrier records that an old-generation object may reference a young-generation object.
// This should be called when modifying old objects to point to young objects.
func (gc *GarbageCollector) WriteBarrier(oldObj, youngObj objects.Object) {
	if !gc.generational || !gc.enabled {
		return
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Find the GCObject for the old object
	var oldGCObj *GCObject
	var youngGCObj *GCObject

	for _, gcObj := range gc.objects {
		if gcObj.Object == oldObj && gcObj.Generation == 1 {
			oldGCObj = gcObj
		}
		if gcObj.Object == youngObj && gcObj.Generation == 0 {
			youngGCObj = gcObj
		}
		if oldGCObj != nil && youngGCObj != nil {
			break
		}
	}

	if oldGCObj != nil && youngGCObj != nil {
		// Check if already in remembered set
		for _, obj := range gc.rememberedSet {
			if obj == oldGCObj {
				return
			}
		}
		gc.rememberedSet = append(gc.rememberedSet, oldGCObj)
	}
}

// SetGenerational enables or disables generational GC mode.
func (gc *GarbageCollector) SetGenerational(enabled bool) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.generational = enabled
}

// SetMinorThreshold sets the threshold for minor collections.
func (gc *GarbageCollector) SetMinorThreshold(threshold int64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.minorThreshold = threshold
}

// SetPromotionAge sets the number of minor GCs an object must survive to be promoted.
func (gc *GarbageCollector) SetPromotionAge(age int) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.promotionAge = age
}

func (gc *GarbageCollector) sweep() int {
	freed := 0
	newObjects := make([]*GCObject, 0, len(gc.objects))

	for _, gcObj := range gc.objects {
		if gc.marked[gcObj] {
			gcObj.Marked = false
			newObjects = append(newObjects, gcObj)
		} else {
			if gcObj.Finalizer != nil {
				gcObj.Finalizer(gcObj.Object)
			}
			gc.freedBytes += gcObj.Size
			freed++
			if gc.verbose {
				log.Printf("GC: Freed %T (%d bytes)", gcObj.Object, gcObj.Size)
			}
		}
	}

	gc.objects = newObjects

	return freed
}

func (gc *GarbageCollector) GetStats() map[string]interface{} {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	return map[string]interface{}{
		"allocated_bytes":  gc.allocatedBytes,
		"freed_bytes":      gc.freedBytes,
		"collection_count": gc.collectionCount,
		"minor_count":      gc.minorCount,
		"major_count":      gc.majorCount,
		"object_count":     len(gc.objects),
		"young_objects":    len(gc.youngObjects),
		"old_objects":      len(gc.oldObjects),
		"enabled":          gc.enabled,
		"generational":     gc.generational,
		"threshold":        gc.threshold,
		"minor_threshold":  gc.minorThreshold,
		"promotion_age":    gc.promotionAge,
		"total_pause_time": gc.pauseTime,
	}
}

func (gc *GarbageCollector) PrintStats() {
	stats := gc.GetStats()
	log.Println("=== GC Statistics ===")
	log.Printf("Allocated: %d bytes", stats["allocated_bytes"])
	log.Printf("Freed: %d bytes", stats["freed_bytes"])
	log.Printf("Collections: %d (minor: %d, major: %d)", stats["collection_count"], stats["minor_count"], stats["major_count"])
	log.Printf("Live objects: %d (young: %d, old: %d)", stats["object_count"], stats["young_objects"], stats["old_objects"])
	log.Printf("Enabled: %v", stats["enabled"])
	log.Printf("Generational: %v", stats["generational"])
	log.Printf("Threshold: %d bytes", stats["threshold"])
	log.Printf("Minor threshold: %d bytes", stats["minor_threshold"])
	log.Printf("Promotion age: %d", stats["promotion_age"])
	log.Printf("Total pause time: %v", stats["total_pause_time"])
}

func (gc *GarbageCollector) AddFinalizer(obj objects.Object, finalizer func(objects.Object)) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	for _, gcObj := range gc.objects {
		if gcObj.Object == obj {
			gcObj.Finalizer = finalizer
			return
		}
	}
}

func estimateSize(obj objects.Object) int64 {
	switch o := obj.(type) {
	case *objects.Integer:
		return 8
	case *objects.Float:
		return 8
	case *objects.Boolean:
		return 1
	case *objects.String:
		return int64(len(o.Value)) + 16
	case *objects.List:
		size := int64(24)
		for _, elem := range o.Elements {
			size += estimateSize(elem)
		}
		return size
	case *objects.Tuple:
		size := int64(24)
		for _, elem := range o.Elements {
			size += estimateSize(elem)
		}
		return size
	case *objects.Dict:
		return int64(48 + len(o.Keys)*32)
	case *objects.Set:
		return int64(48 + len(o.Elements)*16)
	default:
		return 32
	}
}

var registerContents []objects.Object
var stackContents []objects.Object
var globalContents []objects.Object
var frameContents []objects.Object

func getRegisterContents() []objects.Object {
	return registerContents
}

func getStackContents() []objects.Object {
	return stackContents
}

func getGlobalContents() []objects.Object {
	return globalContents
}

func getFrameContents() []objects.Object {
	return frameContents
}

func SetRoots(reg, stack, global, frame []objects.Object) {
	registerContents = reg
	stackContents = stack
	globalContents = global
	frameContents = frame
}



func Collect() {
	GetGC().Collect()
}

func MinorCollect() {
	GetGC().MinorCollect()
}

func MajorCollect() {
	GetGC().MajorCollect()
}

func Enable() {
	GetGC().Enable(true)
}

func Disable() {
	GetGC().Enable(false)
}

func GetStats() map[string]interface{} {
	return GetGC().GetStats()
}

func PrintStats() {
	GetGC().PrintStats()
}

func WriteBarrier(oldObj, youngObj objects.Object) {
	GetGC().WriteBarrier(oldObj, youngObj)
}

func (gc *GarbageCollector) GetStatsAsDict() *objects.Dict {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	result := objects.NewDict()
	result.Set(&objects.String{Value: "allocated_bytes"}, &objects.Integer{Value: gc.allocatedBytes})
	result.Set(&objects.String{Value: "freed_bytes"}, &objects.Integer{Value: gc.freedBytes})
	result.Set(&objects.String{Value: "collection_count"}, &objects.Integer{Value: gc.collectionCount})
	result.Set(&objects.String{Value: "minor_count"}, &objects.Integer{Value: gc.minorCount})
	result.Set(&objects.String{Value: "major_count"}, &objects.Integer{Value: gc.majorCount})
	result.Set(&objects.String{Value: "object_count"}, &objects.Integer{Value: int64(len(gc.objects))})
	result.Set(&objects.String{Value: "young_objects"}, &objects.Integer{Value: int64(len(gc.youngObjects))})
	result.Set(&objects.String{Value: "old_objects"}, &objects.Integer{Value: int64(len(gc.oldObjects))})
	result.Set(&objects.String{Value: "enabled"}, &objects.Boolean{Value: gc.enabled})
	result.Set(&objects.String{Value: "generational"}, &objects.Boolean{Value: gc.generational})
	result.Set(&objects.String{Value: "threshold"}, &objects.Integer{Value: gc.threshold})
	result.Set(&objects.String{Value: "minor_threshold"}, &objects.Integer{Value: gc.minorThreshold})
	result.Set(&objects.String{Value: "promotion_age"}, &objects.Integer{Value: int64(gc.promotionAge)})
	result.Set(&objects.String{Value: "total_pause_time"}, &objects.Float{Value: float64(gc.pauseTime.Seconds())})

	return result
}

type GCModule struct{}

func (m *GCModule) Collect() objects.Object {
	Collect()
	return objects.None_
}

func (m *GCModule) Enable() objects.Object {
	Enable()
	return objects.None_
}

func (m *GCModule) Disable() objects.Object {
	Disable()
	return objects.None_
}

func (m *GCModule) GetStats() objects.Object {
	stats := GetGC().GetStats()
	result := objects.NewDict()
	for k, v := range stats {
		var valObj objects.Object
		switch val := v.(type) {
		case int64:
			valObj = &objects.Integer{Value: val}
		case int:
			valObj = &objects.Integer{Value: int64(val)}
		case bool:
			valObj = &objects.Boolean{Value: val}
		case time.Duration:
			valObj = &objects.Float{Value: float64(val.Seconds())}
		default:
			valObj = &objects.String{Value: fmt.Sprintf("%v", val)}
		}
		result.Set(&objects.String{Value: k}, valObj)
	}
	return result
}

func (m *GCModule) PrintStats() objects.Object {
	PrintStats()
	return objects.None_
}

func (m *GCModule) SetThreshold(threshold objects.Object) objects.Object {
	if t, ok := threshold.(*objects.Integer); ok {
		GetGC().SetThreshold(t.Value)
	}
	return objects.None_
}

func (m *GCModule) SetVerbose(verbose objects.Object) objects.Object {
	if v, ok := verbose.(*objects.Boolean); ok {
		GetGC().SetVerbose(v.Value)
	}
	return objects.None_
}

func CreateGCModule() *objects.Module {
	module := &objects.Module{
		Name:   "gc",
		Fields: make(map[string]objects.Object),
	}
	module.Fields["collect"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		Collect()
		return objects.None_
	}}
	module.Fields["minor_collect"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		GetGC().MinorCollect()
		return objects.None_
	}}
	module.Fields["major_collect"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		GetGC().MajorCollect()
		return objects.None_
	}}
	module.Fields["enable"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		Enable()
		return objects.None_
	}}
	module.Fields["disable"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		Disable()
		return objects.None_
	}}
	module.Fields["get_stats"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		return GetGC().GetStatsAsDict()
	}}
	module.Fields["print_stats"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		PrintStats()
		return objects.None_
	}}
	module.Fields["set_threshold"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) > 0 {
			if t, ok := args[0].(*objects.Integer); ok {
				GetGC().SetThreshold(t.Value)
			}
		}
		return objects.None_
	}}
	module.Fields["set_verbose"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) > 0 {
			if v, ok := args[0].(*objects.Boolean); ok {
				GetGC().SetVerbose(v.Value)
			}
		}
		return objects.None_
	}}
	module.Fields["set_promotion_age"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) > 0 {
			if t, ok := args[0].(*objects.Integer); ok {
				GetGC().SetPromotionAge(int(t.Value))
			}
		}
		return objects.None_
	}}
	module.Fields["set_generational"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) > 0 {
			if v, ok := args[0].(*objects.Boolean); ok {
				GetGC().SetGenerational(v.Value)
			}
		}
		return objects.None_
	}}
	module.Fields["set_minor_threshold"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) > 0 {
			if t, ok := args[0].(*objects.Integer); ok {
				GetGC().SetMinorThreshold(t.Value)
			}
		}
		return objects.None_
	}}
	module.Fields["write_barrier"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) >= 2 {
			GetGC().WriteBarrier(args[0], args[1])
		}
		return objects.None_
	}}
	return module
}
