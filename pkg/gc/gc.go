package gc

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/objects"
)

type Generation int

const (
	YoungGen Generation = iota
	OldGen
)

type GCObject struct {
	Object        objects.Object
	Marked        bool
	Size          int64
	Next          *GCObject
	Prev          *GCObject
	Finalizer     func(objects.Object)
	Generation    Generation
	SurvivedCount int // number of minor GCs survived
}

type GarbageCollector struct {
	mu sync.Mutex

	// Young generation
	youngObjects   []*GCObject
	youngAllocated int64
	youngThreshold int64 // default: 256KB

	// Old generation
	oldObjects   []*GCObject
	oldAllocated int64
	oldThreshold int64 // default: 4MB

	// Index for fast lookup by object pointer
	objectMap map[objects.Object]*GCObject

	// Shared
	marked          map[*GCObject]bool
	allocatedBytes  int64
	freedBytes      int64
	collectionCount int64
	minorCount      int64
	majorCount      int64
	enabled         bool
	verbose         bool
	threshold       int64 // kept for compatibility
	pauseTime       time.Duration

	promotionAge int // number of minor GCs before promotion, default: 3

	// Write barrier remembered set: old gen objects that reference young gen objects
	rememberedSet []*GCObject

	// Counter for triggering major GC after N minor GCs
	minorSinceMajor int64
	majorAfterMinor int64 // trigger major GC after this many minor GCs, default: 10
}

var singleton *GarbageCollector
var once sync.Once

func GetGC() *GarbageCollector {
	once.Do(func() {
		singleton = &GarbageCollector{
			youngObjects:    make([]*GCObject, 0),
			oldObjects:      make([]*GCObject, 0),
			objectMap:       make(map[objects.Object]*GCObject),
			marked:          make(map[*GCObject]bool),
			enabled:         true,
			verbose:         false,
			threshold:       1024 * 1024, // 1MB threshold (kept for compatibility)
			youngThreshold:  256 * 1024,  // 256KB
			oldThreshold:    4 * 1024 * 1024, // 4MB
			promotionAge:    3,
			majorAfterMinor: 10,
			rememberedSet:   make([]*GCObject, 0),
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

func (gc *GarbageCollector) SetYoungThreshold(threshold int64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.youngThreshold = threshold
}

func (gc *GarbageCollector) SetOldThreshold(threshold int64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.oldThreshold = threshold
}

func (gc *GarbageCollector) SetPromotionAge(age int) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.promotionAge = age
}

func (gc *GarbageCollector) Allocate(obj objects.Object) *GCObject {
	if !gc.enabled {
		return &GCObject{Object: obj, Size: estimateSize(obj), Generation: YoungGen}
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	gcObj := &GCObject{
		Object:     obj,
		Marked:     false,
		Size:       estimateSize(obj),
		Generation: YoungGen,
	}

	gc.youngObjects = append(gc.youngObjects, gcObj)
	gc.objectMap[obj] = gcObj
	gc.youngAllocated += gcObj.Size
	gc.allocatedBytes += gcObj.Size

	if gc.verbose {
		log.Printf("GC: Allocated %d bytes in young gen, total %d bytes", gcObj.Size, gc.allocatedBytes)
	}

	if gc.youngAllocated >= gc.youngThreshold {
		go gc.MinorCollect()
	} else if gc.oldAllocated >= gc.oldThreshold {
		go gc.MajorCollect()
	}

	return gcObj
}

func (gc *GarbageCollector) Register(obj objects.Object) *GCObject {
	return gc.Allocate(obj)
}

// Collect performs a minor GC for backward compatibility.
func (gc *GarbageCollector) Collect() {
	gc.MinorCollect()
}

// MinorCollect performs a young generation collection.
func (gc *GarbageCollector) MinorCollect() {
	if !gc.enabled {
		return
	}

	gc.mu.Lock()
	start := time.Now()

	if gc.verbose {
		log.Println("GC: Starting minor collection (young gen)")
	}

	gc.marked = make(map[*GCObject]bool)

	// Mark from roots (only need to scan young gen objects)
	gc.markRootsMinor()

	// Also mark from remembered set (old gen -> young gen references)
	gc.markFromRememberedSet()

	// Sweep young generation
	freed := gc.sweepYoung()

	gc.collectionCount++
	gc.minorCount++
	gc.minorSinceMajor++
	gc.pauseTime += time.Since(start)

	if gc.verbose {
		log.Printf("GC: Minor collection completed, freed %d objects in %v", freed, time.Since(start))
	}

	// Check if we should trigger a major GC
	if gc.oldAllocated >= gc.oldThreshold || gc.minorSinceMajor >= gc.majorAfterMinor {
		gc.majorCollectLocked()
	}

	gc.mu.Unlock()
}

// majorCollectLocked performs a full collection while holding the lock.
// This is called from MinorCollect to avoid unlocking between minor and major GC.
func (gc *GarbageCollector) majorCollectLocked() {
	start := time.Now()

	if gc.verbose {
		log.Println("GC: Starting major collection (full, from minor)")
	}

	gc.marked = make(map[*GCObject]bool)

	// Mark from roots (scan ALL objects)
	gc.markRoots()

	// Sweep both generations
	freedYoung := gc.sweepYoung()
	freedOld := gc.sweepOld()

	gc.collectionCount++
	gc.majorCount++
	gc.minorSinceMajor = 0
	gc.pauseTime += time.Since(start)

	// Clear remembered set on major GC since all objects were scanned
	gc.rememberedSet = gc.rememberedSet[:0]

	if gc.verbose {
		log.Printf("GC: Major collection completed, freed %d young + %d old objects in %v",
			freedYoung, freedOld, time.Since(start))
	}
}

// MajorCollect performs a full collection of both generations.
func (gc *GarbageCollector) MajorCollect() {
	if !gc.enabled {
		return
	}

	gc.mu.Lock()
	start := time.Now()

	if gc.verbose {
		log.Println("GC: Starting major collection (full)")
	}

	gc.marked = make(map[*GCObject]bool)

	// Mark from roots (scan ALL objects)
	gc.markRoots()

	// Sweep both generations
	freedYoung := gc.sweepYoung()
	freedOld := gc.sweepOld()

	gc.collectionCount++
	gc.majorCount++
	gc.minorSinceMajor = 0
	gc.pauseTime += time.Since(start)

	// Clear remembered set on major GC since all objects were scanned
	gc.rememberedSet = gc.rememberedSet[:0]

	if gc.verbose {
		log.Printf("GC: Major collection completed, freed %d young + %d old objects in %v",
			freedYoung, freedOld, time.Since(start))
	}

	gc.mu.Unlock()
}

// markRootsMinor marks reachable objects from roots, only considering young gen.
func (gc *GarbageCollector) markRootsMinor() {
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
		gc.markObjectMinor(root)
	}
}

// markObjectMinor marks an object only if it's in the young generation.
func (gc *GarbageCollector) markObjectMinor(obj objects.Object) {
	if obj == nil {
		return
	}

	gcObj, ok := gc.objectMap[obj]
	if !ok || gc.marked[gcObj] || gcObj.Generation != YoungGen {
		return
	}
	gc.marked[gcObj] = true
	gc.markReferencesMinor(gcObj.Object)
}

// markReferencesMinor follows references from a container, only marking young gen objects.
func (gc *GarbageCollector) markReferencesMinor(obj objects.Object) {
	switch o := obj.(type) {
	case *objects.List:
		for _, elem := range o.Elements {
			gc.markObjectMinor(elem)
		}
	case *objects.Tuple:
		for _, elem := range o.Elements {
			gc.markObjectMinor(elem)
		}
	case *objects.Dict:
		for keyStr, key := range o.Keys {
			gc.markObjectMinor(key)
			if val, ok := o.Pairs[keyStr]; ok {
				gc.markObjectMinor(val)
			}
		}
	case *objects.Set:
		for _, elem := range o.Elements {
			gc.markObjectMinor(elem)
		}
	case *objects.Instance:
		for _, field := range o.Fields {
			gc.markObjectMinor(field)
		}
	case *objects.Class:
		if o.SuperClass != nil {
			gc.markObjectMinor(o.SuperClass)
		}
	}
}

// markFromRememberedSet marks young gen objects referenced by old gen objects
// in the remembered set.
func (gc *GarbageCollector) markFromRememberedSet() {
	for _, oldObj := range gc.rememberedSet {
		if oldObj.Object == nil {
			continue
		}
		gc.markReferencesMinor(oldObj.Object)
	}
}

// markRoots marks reachable objects from roots, scanning ALL objects (both generations).
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

// markObject marks an object in either generation.
func (gc *GarbageCollector) markObject(obj objects.Object) {
	if obj == nil {
		return
	}

	gcObj, ok := gc.objectMap[obj]
	if !ok || gc.marked[gcObj] {
		return
	}
	gc.marked[gcObj] = true
	gc.markReferences(gcObj.Object)
}

// markReferences follows references from a container, marking objects in both generations.
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

// sweepYoung sweeps the young generation:
// - Unmarked young objects are freed
// - Marked young objects with SurvivedCount >= promotionAge are promoted to old gen
// - Marked young objects with SurvivedCount < promotionAge have SurvivedCount++
func (gc *GarbageCollector) sweepYoung() int {
	freed := 0
	newYoungObjects := make([]*GCObject, 0, len(gc.youngObjects))

	for _, gcObj := range gc.youngObjects {
		if gc.marked[gcObj] {
			gcObj.Marked = false
			gcObj.SurvivedCount++
			if gcObj.SurvivedCount >= gc.promotionAge {
				// Promote to old generation
				gcObj.Generation = OldGen
				gc.oldObjects = append(gc.oldObjects, gcObj)
				gc.youngAllocated -= gcObj.Size
				gc.oldAllocated += gcObj.Size
				if gc.verbose {
					log.Printf("GC: Promoted %T to old gen (survived %d minor GCs)", gcObj.Object, gcObj.SurvivedCount)
				}
			} else {
				newYoungObjects = append(newYoungObjects, gcObj)
			}
		} else {
			if gcObj.Finalizer != nil {
				gcObj.Finalizer(gcObj.Object)
			}
			gc.freedBytes += gcObj.Size
			gc.youngAllocated -= gcObj.Size
			gc.allocatedBytes -= gcObj.Size
			delete(gc.objectMap, gcObj.Object)
			freed++
			if gc.verbose {
				log.Printf("GC: Freed young %T (%d bytes)", gcObj.Object, gcObj.Size)
			}
		}
	}

	gc.youngObjects = newYoungObjects
	return freed
}

// sweepOld sweeps the old generation.
func (gc *GarbageCollector) sweepOld() int {
	freed := 0
	newOldObjects := make([]*GCObject, 0, len(gc.oldObjects))

	for _, gcObj := range gc.oldObjects {
		if gc.marked[gcObj] {
			gcObj.Marked = false
			newOldObjects = append(newOldObjects, gcObj)
		} else {
			if gcObj.Finalizer != nil {
				gcObj.Finalizer(gcObj.Object)
			}
			gc.freedBytes += gcObj.Size
			gc.oldAllocated -= gcObj.Size
			gc.allocatedBytes -= gcObj.Size
			delete(gc.objectMap, gcObj.Object)
			freed++
			if gc.verbose {
				log.Printf("GC: Freed old %T (%d bytes)", gcObj.Object, gcObj.Size)
			}
		}
	}

	gc.oldObjects = newOldObjects
	return freed
}

// WriteBarrier should be called when an old gen object gains a reference to a young gen object.
// It adds the old gen object to the remembered set so the young gen object is not prematurely collected.
func (gc *GarbageCollector) WriteBarrier(container *GCObject, contained *GCObject) {
	if container == nil || contained == nil {
		return
	}
	if container.Generation == OldGen && contained.Generation == YoungGen {
		gc.mu.Lock()
		defer gc.mu.Unlock()
		// Check if already in remembered set
		for _, obj := range gc.rememberedSet {
			if obj == container {
				return
			}
		}
		gc.rememberedSet = append(gc.rememberedSet, container)
	}
}

func (gc *GarbageCollector) GetStats() map[string]interface{} {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	return map[string]interface{}{
		"allocated_bytes":    gc.allocatedBytes,
		"freed_bytes":        gc.freedBytes,
		"collection_count":   gc.collectionCount,
		"minor_count":        gc.minorCount,
		"major_count":        gc.majorCount,
		"object_count":       len(gc.youngObjects) + len(gc.oldObjects),
		"young_object_count": len(gc.youngObjects),
		"old_object_count":   len(gc.oldObjects),
		"young_allocated":    gc.youngAllocated,
		"old_allocated":      gc.oldAllocated,
		"young_threshold":    gc.youngThreshold,
		"old_threshold":      gc.oldThreshold,
		"promotion_age":      gc.promotionAge,
		"remembered_set_size": len(gc.rememberedSet),
		"enabled":            gc.enabled,
		"threshold":          gc.threshold,
		"total_pause_time":   gc.pauseTime,
	}
}

func (gc *GarbageCollector) PrintStats() {
	stats := gc.GetStats()
	log.Println("=== GC Statistics ===")
	log.Printf("Allocated: %d bytes", stats["allocated_bytes"])
	log.Printf("  Young gen: %d bytes (%d objects)", stats["young_allocated"], stats["young_object_count"])
	log.Printf("  Old gen: %d bytes (%d objects)", stats["old_allocated"], stats["old_object_count"])
	log.Printf("Freed: %d bytes", stats["freed_bytes"])
	log.Printf("Collections: %d (minor: %d, major: %d)", stats["collection_count"], stats["minor_count"], stats["major_count"])
	log.Printf("Live objects: %d", stats["object_count"])
	log.Printf("Promotion age: %d", stats["promotion_age"])
	log.Printf("Remembered set: %d", stats["remembered_set_size"])
	log.Printf("Enabled: %v", stats["enabled"])
	log.Printf("Threshold: %d bytes (young: %d, old: %d)", stats["threshold"], stats["young_threshold"], stats["old_threshold"])
	log.Printf("Total pause time: %v", stats["total_pause_time"])
}

func (gc *GarbageCollector) AddFinalizer(obj objects.Object, finalizer func(objects.Object)) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if gcObj, ok := gc.objectMap[obj]; ok {
		gcObj.Finalizer = finalizer
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

// Package-level functions for backward compatibility

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

func (gc *GarbageCollector) GetStatsAsDict() *objects.Dict {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	result := objects.NewDict()
	result.Set(&objects.String{Value: "allocated_bytes"}, &objects.Integer{Value: gc.allocatedBytes})
	result.Set(&objects.String{Value: "freed_bytes"}, &objects.Integer{Value: gc.freedBytes})
	result.Set(&objects.String{Value: "collection_count"}, &objects.Integer{Value: gc.collectionCount})
	result.Set(&objects.String{Value: "minor_count"}, &objects.Integer{Value: gc.minorCount})
	result.Set(&objects.String{Value: "major_count"}, &objects.Integer{Value: gc.majorCount})
	result.Set(&objects.String{Value: "object_count"}, &objects.Integer{Value: int64(len(gc.youngObjects) + len(gc.oldObjects))})
	result.Set(&objects.String{Value: "young_object_count"}, &objects.Integer{Value: int64(len(gc.youngObjects))})
	result.Set(&objects.String{Value: "old_object_count"}, &objects.Integer{Value: int64(len(gc.oldObjects))})
	result.Set(&objects.String{Value: "young_allocated"}, &objects.Integer{Value: gc.youngAllocated})
	result.Set(&objects.String{Value: "old_allocated"}, &objects.Integer{Value: gc.oldAllocated})
	result.Set(&objects.String{Value: "young_threshold"}, &objects.Integer{Value: gc.youngThreshold})
	result.Set(&objects.String{Value: "old_threshold"}, &objects.Integer{Value: gc.oldThreshold})
	result.Set(&objects.String{Value: "promotion_age"}, &objects.Integer{Value: int64(gc.promotionAge)})
	result.Set(&objects.String{Value: "remembered_set_size"}, &objects.Integer{Value: int64(len(gc.rememberedSet))})
	result.Set(&objects.String{Value: "enabled"}, &objects.Boolean{Value: gc.enabled})
	result.Set(&objects.String{Value: "threshold"}, &objects.Integer{Value: gc.threshold})
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
		MinorCollect()
		return objects.None_
	}}
	module.Fields["major_collect"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		MajorCollect()
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
	module.Fields["set_young_threshold"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) > 0 {
			if t, ok := args[0].(*objects.Integer); ok {
				GetGC().SetYoungThreshold(t.Value)
			}
		}
		return objects.None_
	}}
	module.Fields["set_old_threshold"] = &objects.Builtin{Fn: func(args ...objects.Object) objects.Object {
		if len(args) > 0 {
			if t, ok := args[0].(*objects.Integer); ok {
				GetGC().SetOldThreshold(t.Value)
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
	return module
}
