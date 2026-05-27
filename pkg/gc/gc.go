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
	objects         []*GCObject
	marked          map[*GCObject]bool
	allocatedBytes  int64
	freedBytes      int64
	collectionCount int64
	enabled         bool
	verbose         bool
	threshold       int64
	pauseTime       time.Duration
}

type GCObject struct {
	Object      objects.Object
	Marked      bool
	Size        int64
	Next        *GCObject
	Prev        *GCObject
	Finalizer   func(objects.Object)
}

var singleton *GarbageCollector
var once sync.Once

func GetGC() *GarbageCollector {
	once.Do(func() {
		singleton = &GarbageCollector{
			objects:   make([]*GCObject, 0),
			marked:    make(map[*GCObject]bool),
			enabled:   true,
			verbose:   false,
			threshold: 1024 * 1024, // 1MB threshold
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
		Object: obj,
		Marked: false,
		Size:   estimateSize(obj),
	}

	gc.objects = append(gc.objects, gcObj)
	gc.allocatedBytes += gcObj.Size

	if gc.verbose {
		log.Printf("GC: Allocated %d bytes, total %d bytes", gcObj.Size, gc.allocatedBytes)
	}

	if gc.allocatedBytes >= gc.threshold {
		go gc.Collect()
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
		"object_count":     len(gc.objects),
		"enabled":          gc.enabled,
		"threshold":        gc.threshold,
		"total_pause_time": gc.pauseTime,
	}
}

func (gc *GarbageCollector) PrintStats() {
	stats := gc.GetStats()
	log.Println("=== GC Statistics ===")
	log.Printf("Allocated: %d bytes", stats["allocated_bytes"])
	log.Printf("Freed: %d bytes", stats["freed_bytes"])
	log.Printf("Collections: %d", stats["collection_count"])
	log.Printf("Live objects: %d", stats["object_count"])
	log.Printf("Enabled: %v", stats["enabled"])
	log.Printf("Threshold: %d bytes", stats["threshold"])
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
	result.Set(&objects.String{Value: "object_count"}, &objects.Integer{Value: int64(len(gc.objects))})
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
	return module
}
