package gc

import (
	"log"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/objects"
)

type IncrementalGarbageCollector struct {
	mu                sync.Mutex
	objects           []*GCObject
	marked            map[*GCObject]bool
	allocatedBytes    int64
	freedBytes        int64
	collectionCount   int64
	enabled           bool
	verbose           bool
	threshold         int64
	pauseTime         time.Duration
	incrementalStep   int
	incrementalMarking bool
	markingIndex      int
	markingQueue      []*GCObject
	maxPauseTime      time.Duration
	generations       [3][]int
	promotionAge      int
}

var incrementalSingleton *IncrementalGarbageCollector
var incrementalOnce sync.Once

func GetIncrementalGC() *IncrementalGarbageCollector {
	incrementalOnce.Do(func() {
		incrementalSingleton = &IncrementalGarbageCollector{
			objects:           make([]*GCObject, 0),
			marked:            make(map[*GCObject]bool),
			enabled:           true,
			verbose:           false,
			threshold:         1024 * 1024,
			incrementalStep:   100,
			maxPauseTime:      5 * time.Millisecond,
			generations:       [3][]int{},
			promotionAge:      3,
		}
	})
	return incrementalSingleton
}

func (gc *IncrementalGarbageCollector) Enable(enable bool) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.enabled = enable
}

func (gc *IncrementalGarbageCollector) SetVerbose(verbose bool) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.verbose = verbose
}

func (gc *IncrementalGarbageCollector) SetThreshold(threshold int64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.threshold = threshold
}

func (gc *IncrementalGarbageCollector) SetMaxPauseTime(maxPause time.Duration) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.maxPauseTime = maxPause
}

func (gc *IncrementalGarbageCollector) Allocate(obj objects.Object) *GCObject {
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
		go gc.TriggerIncrementalCollection()
	}

	return gcObj
}

func (gc *IncrementalGarbageCollector) Register(obj objects.Object) *GCObject {
	return gc.Allocate(obj)
}

func (gc *IncrementalGarbageCollector) TriggerIncrementalCollection() {
	gc.mu.Lock()
	if gc.incrementalMarking {
		gc.mu.Unlock()
		return
	}
	gc.incrementalMarking = true
	gc.markingIndex = 0
	gc.markingQueue = make([]*GCObject, 0)
	gc.marked = make(map[*GCObject]bool)
	gc.mu.Unlock()

	go gc.incrementalMarkPhase()
}

func (gc *IncrementalGarbageCollector) incrementalMarkPhase() {
	for {
		gc.mu.Lock()
		if !gc.incrementalMarking {
			gc.mu.Unlock()
			return
		}

		start := time.Now()
		steps := 0

		if gc.markingIndex == 0 {
			gc.markRootsIncremental()
		}

		for gc.markingIndex < len(gc.objects) && steps < gc.incrementalStep {
			gcObj := gc.objects[gc.markingIndex]
			if gcObj != nil && !gc.marked[gcObj] {
				gc.marked[gcObj] = true
				gc.markingQueue = append(gc.markingQueue, gcObj)
			}
			gc.markingIndex++
			steps++
		}

		for len(gc.markingQueue) > 0 && steps < gc.incrementalStep {
			gcObj := gc.markingQueue[0]
			gc.markingQueue = gc.markingQueue[1:]
			gc.markReferencesIncremental(gcObj.Object)
			steps++
		}

		if gc.markingIndex >= len(gc.objects) && len(gc.markingQueue) == 0 {
			gc.incrementalMarking = false
			gc.mu.Unlock()
			gc.incrementalSweepPhase()
			return
		}

		elapsed := time.Since(start)
		if elapsed > gc.maxPauseTime {
			gc.mu.Unlock()
			time.Sleep(1 * time.Millisecond)
			continue
		}

		gc.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
}

func (gc *IncrementalGarbageCollector) markRootsIncremental() {
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
		for _, gcObj := range gc.objects {
			if gcObj != nil && gcObj.Object == root && !gc.marked[gcObj] {
				gc.marked[gcObj] = true
				gc.markingQueue = append(gc.markingQueue, gcObj)
				break
			}
		}
	}
}

func (gc *IncrementalGarbageCollector) markReferencesIncremental(obj objects.Object) {
	switch o := obj.(type) {
	case *objects.List:
		for _, elem := range o.Elements {
			if elem != nil {
				for _, gcObj := range gc.objects {
					if gcObj != nil && gcObj.Object == elem && !gc.marked[gcObj] {
						gc.marked[gcObj] = true
						gc.markingQueue = append(gc.markingQueue, gcObj)
						break
					}
				}
			}
		}
	case *objects.Tuple:
		for _, elem := range o.Elements {
			if elem != nil {
				for _, gcObj := range gc.objects {
					if gcObj != nil && gcObj.Object == elem && !gc.marked[gcObj] {
						gc.marked[gcObj] = true
						gc.markingQueue = append(gc.markingQueue, gcObj)
						break
					}
				}
			}
		}
	case *objects.Dict:
		for keyStr, key := range o.Keys {
			if key != nil {
				for _, gcObj := range gc.objects {
					if gcObj != nil && gcObj.Object == key && !gc.marked[gcObj] {
						gc.marked[gcObj] = true
						gc.markingQueue = append(gc.markingQueue, gcObj)
						break
					}
				}
			}
			if val, ok := o.Pairs[keyStr]; ok && val != nil {
				for _, gcObj := range gc.objects {
					if gcObj != nil && gcObj.Object == val && !gc.marked[gcObj] {
						gc.marked[gcObj] = true
						gc.markingQueue = append(gc.markingQueue, gcObj)
						break
					}
				}
			}
		}
	case *objects.Set:
		for _, elem := range o.Elements {
			if elem != nil {
				for _, gcObj := range gc.objects {
					if gcObj != nil && gcObj.Object == elem && !gc.marked[gcObj] {
						gc.marked[gcObj] = true
						gc.markingQueue = append(gc.markingQueue, gcObj)
						break
					}
				}
			}
		}
	case *objects.Instance:
		for _, field := range o.Fields {
			if field != nil {
				for _, gcObj := range gc.objects {
					if gcObj != nil && gcObj.Object == field && !gc.marked[gcObj] {
						gc.marked[gcObj] = true
						gc.markingQueue = append(gc.markingQueue, gcObj)
						break
					}
				}
			}
		}
	case *objects.Class:
		if o.SuperClass != nil {
			for _, gcObj := range gc.objects {
				if gcObj != nil && gcObj.Object == o.SuperClass && !gc.marked[gcObj] {
					gc.marked[gcObj] = true
					gc.markingQueue = append(gc.markingQueue, gcObj)
					break
				}
			}
		}
	}
}

func (gc *IncrementalGarbageCollector) incrementalSweepPhase() {
	gc.mu.Lock()
	start := time.Now()

	if gc.verbose {
		log.Println("GC: Starting sweep phase")
	}

	freed := 0
	newObjects := make([]*GCObject, 0, len(gc.objects))

	for i, gcObj := range gc.objects {
		if gcObj == nil {
			continue
		}

		if gc.marked[gcObj] {
			gcObj.Marked = false
			newObjects = append(newObjects, gcObj)
		} else {
			if gcObj.Finalizer != nil {
				gcObj.Finalizer(gcObj.Object)
			}
			freed++
			gc.freedBytes += gcObj.Size
			gc.allocatedBytes -= gcObj.Size
			gc.objects[i] = nil
		}
	}

	gc.objects = newObjects

	gc.collectionCount++
	gc.pauseTime += time.Since(start)

	if gc.verbose {
		log.Printf("GC: Sweep completed, freed %d objects (%d bytes) in %v",
			freed, gc.freedBytes, time.Since(start))
	}

	gc.mu.Unlock()
}

func (gc *IncrementalGarbageCollector) Collect() {
	if !gc.enabled {
		return
	}

	gc.mu.Lock()
	start := time.Now()

	if gc.verbose {
		log.Println("GC: Starting full collection")
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

func (gc *IncrementalGarbageCollector) markRoots() {
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

func (gc *IncrementalGarbageCollector) markObject(obj objects.Object) {
	if obj == nil {
		return
	}

	for _, gcObj := range gc.objects {
		if gcObj != nil && gcObj.Object == obj && !gc.marked[gcObj] {
			gc.marked[gcObj] = true
			gc.markReferences(gcObj.Object)
			break
		}
	}
}

func (gc *IncrementalGarbageCollector) markReferences(obj objects.Object) {
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

func (gc *IncrementalGarbageCollector) sweep() int {
	freed := 0
	newObjects := make([]*GCObject, 0, len(gc.objects))

	for _, gcObj := range gc.objects {
		if gcObj == nil {
			continue
		}

		if gc.marked[gcObj] {
			gcObj.Marked = false
			newObjects = append(newObjects, gcObj)
		} else {
			if gcObj.Finalizer != nil {
				gcObj.Finalizer(gcObj.Object)
			}
			freed++
			gc.freedBytes += gcObj.Size
			gc.allocatedBytes -= gcObj.Size
		}
	}

	gc.objects = newObjects
	return freed
}

func (gc *IncrementalGarbageCollector) GetStats() GCStats {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	return GCStats{
		AllocatedBytes:  gc.allocatedBytes,
		FreedBytes:      gc.freedBytes,
		CollectionCount: gc.collectionCount,
		PauseTime:       gc.pauseTime,
		ObjectCount:     len(gc.objects),
	}
}

type GCStats struct {
	AllocatedBytes  int64
	FreedBytes      int64
	CollectionCount int64
	PauseTime       time.Duration
	ObjectCount     int
}