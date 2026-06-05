package gc

import (
	"sync"
	"testing"

	"github.com/go-py/go-python/pkg/objects"
)

// newTestGC creates a fresh GarbageCollector for testing.
// We cannot use GetGC() because it's a singleton with sync.Once.
func newTestGC() *GarbageCollector {
	return &GarbageCollector{
		objects:        make([]*GCObject, 0),
		youngObjects:   make([]*GCObject, 0),
		oldObjects:     make([]*GCObject, 0),
		marked:         make(map[*GCObject]bool),
		rememberedSet:  make([]*GCObject, 0),
		enabled:        true,
		verbose:        false,
		generational:   true,
		threshold:      1024 * 1024,
		minorThreshold: 256 * 1024,
		promotionAge:   1,
	}
}

func TestAllocateGoesToYoungGeneration(t *testing.T) {
	gc := newTestGC()
	obj := &objects.Integer{Value: 42}
	gcObj := gc.Allocate(obj)

	if gcObj.Generation != 0 {
		t.Errorf("expected Generation=0 (young), got %d", gcObj.Generation)
	}
	if gcObj.Survived != 0 {
		t.Errorf("expected Survived=0, got %d", gcObj.Survived)
	}
	if len(gc.youngObjects) != 1 {
		t.Errorf("expected 1 young object, got %d", len(gc.youngObjects))
	}
	if len(gc.oldObjects) != 0 {
		t.Errorf("expected 0 old objects, got %d", len(gc.oldObjects))
	}
	if len(gc.objects) != 1 {
		t.Errorf("expected 1 total object, got %d", len(gc.objects))
	}
}

func TestMinorCollectionSweepsYoungUnreachable(t *testing.T) {
	gc := newTestGC()

	// Create a root that keeps one object alive
	keepAlive := &objects.Integer{Value: 1}
	orphan := &objects.Integer{Value: 2}

	gc.Allocate(keepAlive)
	gc.Allocate(orphan)

	// Set roots so keepAlive is reachable
	SetRoots(nil, []objects.Object{keepAlive}, nil, nil)

	if len(gc.youngObjects) != 2 {
		t.Fatalf("expected 2 young objects before minor GC, got %d", len(gc.youngObjects))
	}

	gc.MinorCollect()

	// After minor collection, only keepAlive should survive
	if len(gc.youngObjects) != 0 {
		t.Errorf("expected 0 young objects after minor GC (survivors promoted), got %d", len(gc.youngObjects))
	}
	// keepAlive should have been promoted to old generation
	if len(gc.oldObjects) != 1 {
		t.Errorf("expected 1 old object (promoted), got %d", len(gc.oldObjects))
	}
	if gc.oldObjects[0].Object != keepAlive {
		t.Error("expected keepAlive to be promoted to old generation")
	}
}

func TestPromotionAfterMinorGC(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 2 // require 2 minor GCs to promote

	obj := &objects.Integer{Value: 42}
	gc.Allocate(obj)

	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// First minor GC: survived=1, not yet promoted
	gc.MinorCollect()
	if len(gc.youngObjects) != 1 {
		t.Errorf("expected 1 young object after 1st minor GC, got %d", len(gc.youngObjects))
	}
	if len(gc.oldObjects) != 0 {
		t.Errorf("expected 0 old objects after 1st minor GC, got %d", len(gc.oldObjects))
	}
	if gc.youngObjects[0].Survived != 1 {
		t.Errorf("expected Survived=1, got %d", gc.youngObjects[0].Survived)
	}

	// Second minor GC: survived=2, should be promoted
	gc.MinorCollect()
	if len(gc.youngObjects) != 0 {
		t.Errorf("expected 0 young objects after 2nd minor GC, got %d", len(gc.youngObjects))
	}
	if len(gc.oldObjects) != 1 {
		t.Errorf("expected 1 old object after 2nd minor GC, got %d", len(gc.oldObjects))
	}
	if gc.oldObjects[0].Generation != 1 {
		t.Errorf("expected Generation=1 (old), got %d", gc.oldObjects[0].Generation)
	}
}

func TestMajorCollectionSweepsAllUnreachable(t *testing.T) {
	gc := newTestGC()

	keepAlive := &objects.Integer{Value: 1}
	orphan := &objects.Integer{Value: 2}

	gc.Allocate(keepAlive)
	gc.Allocate(orphan)

	SetRoots(nil, []objects.Object{keepAlive}, nil, nil)

	// First promote keepAlive to old gen
	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Fatalf("expected 1 old object after minor GC, got %d", len(gc.oldObjects))
	}

	// Now allocate another young object that won't be reachable
	anotherOrphan := &objects.Integer{Value: 3}
	gc.Allocate(anotherOrphan)

	gc.MajorCollect()

	// After major GC, only keepAlive should remain
	if len(gc.objects) != 1 {
		t.Errorf("expected 1 object after major GC, got %d", len(gc.objects))
	}
	if gc.objects[0].Object != keepAlive {
		t.Error("expected keepAlive to survive major GC")
	}
}

func TestMajorCollectionResetsGenerations(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 42}
	gc.Allocate(obj)

	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Promote to old gen
	gc.MinorCollect()
	if len(gc.oldObjects) != 1 {
		t.Fatalf("expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Major collection rebuilds generation lists
	gc.MajorCollect()

	if len(gc.objects) != 1 {
		t.Errorf("expected 1 object after major GC, got %d", len(gc.objects))
	}
	// The object should still be in old generation
	if gc.objects[0].Generation != 1 {
		t.Errorf("expected Generation=1, got %d", gc.objects[0].Generation)
	}
}

func TestMinorCollectionDoesNotSweepOldObjects(t *testing.T) {
	gc := newTestGC()

	oldObj := &objects.Integer{Value: 1}
	gc.Allocate(oldObj)

	SetRoots(nil, nil, nil, nil) // no roots

	// Promote oldObj to old gen via minor GC
	// But wait, with no roots it will be swept. Let's add it as root first.
	SetRoots(nil, []objects.Object{oldObj}, nil, nil)
	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Fatalf("expected 1 old object after promotion, got %d", len(gc.oldObjects))
	}

	// Now remove from roots and do a minor collection
	SetRoots(nil, nil, nil, nil)
	gc.MinorCollect()

	// Old objects should NOT be collected during minor GC
	if len(gc.oldObjects) != 1 {
		t.Errorf("expected old object to survive minor GC, got %d old objects", len(gc.oldObjects))
	}
	if len(gc.objects) != 1 {
		t.Errorf("expected 1 total object after minor GC, got %d", len(gc.objects))
	}

	// But major GC should collect it
	gc.MajorCollect()
	if len(gc.objects) != 0 {
		t.Errorf("expected 0 objects after major GC, got %d", len(gc.objects))
	}
}

func TestCollectDefaultsToMinorInGenerationalMode(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 42}
	gc.Allocate(obj)

	SetRoots(nil, []objects.Object{obj}, nil, nil)

	gc.Collect()

	if gc.minorCount != 1 {
		t.Errorf("expected minorCount=1, got %d", gc.minorCount)
	}
	if gc.majorCount != 0 {
		t.Errorf("expected majorCount=0, got %d", gc.majorCount)
	}
}

func TestCollectFallsBackToMajorPeriodically(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 42}
	gc.Allocate(obj)

	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Do 4 minor collections
	for i := 0; i < 4; i++ {
		gc.Collect()
	}

	if gc.minorCount != 4 {
		t.Errorf("expected minorCount=4, got %d", gc.minorCount)
	}
	if gc.majorCount != 0 {
		t.Errorf("expected majorCount=0, got %d", gc.majorCount)
	}

	// 5th collection should trigger major
	gc.Collect()
	if gc.majorCount != 1 {
		t.Errorf("expected majorCount=1 after 5th Collect(), got %d", gc.majorCount)
	}
}

func TestNonGenerationalMode(t *testing.T) {
	gc := newTestGC()
	gc.generational = false

	obj := &objects.Integer{Value: 42}
	gc.Allocate(obj)

	SetRoots(nil, []objects.Object{obj}, nil, nil)

	gc.Collect()

	// In non-generational mode, Collect() should do full mark-and-sweep
	if gc.collectionCount != 1 {
		t.Errorf("expected collectionCount=1, got %d", gc.collectionCount)
	}
	if gc.minorCount != 0 {
		t.Errorf("expected minorCount=0 in non-generational mode, got %d", gc.minorCount)
	}
	if gc.majorCount != 0 {
		t.Errorf("expected majorCount=0 in non-generational mode, got %d", gc.majorCount)
	}
}

func TestGetStatsIncludesGenerationalFields(t *testing.T) {
	gc := newTestGC()

	stats := gc.GetStats()

	if _, ok := stats["minor_count"]; !ok {
		t.Error("expected 'minor_count' in stats")
	}
	if _, ok := stats["major_count"]; !ok {
		t.Error("expected 'major_count' in stats")
	}
	if _, ok := stats["young_objects"]; !ok {
		t.Error("expected 'young_objects' in stats")
	}
	if _, ok := stats["old_objects"]; !ok {
		t.Error("expected 'old_objects' in stats")
	}
	if _, ok := stats["generational"]; !ok {
		t.Error("expected 'generational' in stats")
	}
	if _, ok := stats["minor_threshold"]; !ok {
		t.Error("expected 'minor_threshold' in stats")
	}
	if _, ok := stats["promotion_age"]; !ok {
		t.Error("expected 'promotion_age' in stats")
	}

	if stats["generational"] != true {
		t.Errorf("expected generational=true, got %v", stats["generational"])
	}
	if stats["promotion_age"] != 1 {
		t.Errorf("expected promotion_age=1, got %v", stats["promotion_age"])
	}
}

func TestSetPromotionAge(t *testing.T) {
	gc := newTestGC()
	gc.SetPromotionAge(3)

	if gc.promotionAge != 3 {
		t.Errorf("expected promotionAge=3, got %d", gc.promotionAge)
	}
}

func TestSetMinorThreshold(t *testing.T) {
	gc := newTestGC()
	gc.SetMinorThreshold(512 * 1024)

	if gc.minorThreshold != 512*1024 {
		t.Errorf("expected minorThreshold=524288, got %d", gc.minorThreshold)
	}
}

func TestSetGenerational(t *testing.T) {
	gc := newTestGC()
	gc.SetGenerational(false)

	if gc.generational != false {
		t.Error("expected generational=false")
	}

	gc.SetGenerational(true)
	if gc.generational != true {
		t.Error("expected generational=true")
	}
}

func TestWriteBarrier(t *testing.T) {
	gc := newTestGC()

	// Create an old object and a young object
	oldObj := &objects.Integer{Value: 1}
	youngObj := &objects.Integer{Value: 2}

	oldGCObj := gc.Allocate(oldObj)
	youngGCObj := gc.Allocate(youngObj)

	// Promote oldObj to old generation
	oldGCObj.Generation = 1
	gc.youngObjects = gc.youngObjects[:0] // remove from young
	gc.oldObjects = append(gc.oldObjects, oldGCObj)

	// Call write barrier
	gc.WriteBarrier(oldObj, youngObj)

	// Check that oldGCObj is in the remembered set
	found := false
	for _, obj := range gc.rememberedSet {
		if obj == oldGCObj {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected oldGCObj to be in remembered set after write barrier")
	}

	// Young object should still be young
	if youngGCObj.Generation != 0 {
		t.Errorf("expected youngGCObj.Generation=0, got %d", youngGCObj.Generation)
	}
}

func TestWriteBarrierNoDuplicate(t *testing.T) {
	gc := newTestGC()

	oldObj := &objects.Integer{Value: 1}
	youngObj := &objects.Integer{Value: 2}

	oldGCObj := gc.Allocate(oldObj)
	gc.Allocate(youngObj)

	oldGCObj.Generation = 1
	gc.youngObjects = gc.youngObjects[:0]
	gc.oldObjects = append(gc.oldObjects, oldGCObj)

	gc.WriteBarrier(oldObj, youngObj)
	gc.WriteBarrier(oldObj, youngObj)

	if len(gc.rememberedSet) != 1 {
		t.Errorf("expected 1 entry in remembered set (no duplicates), got %d", len(gc.rememberedSet))
	}
}

func TestWriteBarrierDisabledWhenNotGenerational(t *testing.T) {
	gc := newTestGC()
	gc.generational = false

	oldObj := &objects.Integer{Value: 1}
	youngObj := &objects.Integer{Value: 2}

	gc.Allocate(oldObj)
	gc.Allocate(youngObj)

	gc.WriteBarrier(oldObj, youngObj)

	if len(gc.rememberedSet) != 0 {
		t.Errorf("expected empty remembered set when generational=false, got %d", len(gc.rememberedSet))
	}
}

func TestMinorCollectionWithRememberedSet(t *testing.T) {
	gc := newTestGC()

	// Create an old list that references a young integer
	oldList := &objects.List{Elements: []objects.Object{}}
	youngInt := &objects.Integer{Value: 99}

	gc.Allocate(oldList)
	gc.Allocate(youngInt)

	// Promote oldList to old gen
	SetRoots(nil, []objects.Object{oldList}, nil, nil)
	gc.MinorCollect() // oldList gets promoted

	if len(gc.oldObjects) != 1 {
		t.Fatalf("expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Now add a young object and make old list reference it
	anotherYoung := &objects.Integer{Value: 77}
	gc.Allocate(anotherYoung)
	oldList.Elements = append(oldList.Elements, anotherYoung)

	// Add old list to remembered set
	gc.WriteBarrier(oldList, anotherYoung)

	// Remove from roots so young object is only reachable via old list
	SetRoots(nil, []objects.Object{oldList}, nil, nil)

	gc.MinorCollect()

	// anotherYoung should survive because it's reachable via the remembered set
	// It should be promoted to old gen
	found := false
	for _, obj := range gc.oldObjects {
		if obj.Object == anotherYoung {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected anotherYoung to be promoted to old generation via remembered set")
	}
}

func TestFinalizerCalledOnMinorSweep(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 42}
	gcObj := gc.Allocate(obj)

	finalized := false
	gcObj.Finalizer = func(o objects.Object) {
		finalized = true
	}

	// No roots, so obj should be collected
	SetRoots(nil, nil, nil, nil)

	gc.MinorCollect()

	if !finalized {
		t.Error("expected finalizer to be called during minor collection")
	}
}

func TestFinalizerCalledOnMajorSweep(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 42}
	gcObj := gc.Allocate(obj)

	finalized := false
	gcObj.Finalizer = func(o objects.Object) {
		finalized = true
	}

	// Promote to old gen first
	SetRoots(nil, []objects.Object{obj}, nil, nil)
	gc.MinorCollect()

	// Now remove from roots
	SetRoots(nil, nil, nil, nil)

	gc.MajorCollect()

	if !finalized {
		t.Error("expected finalizer to be called during major collection")
	}
}

func TestGetStatsAsDictIncludesGenerationalFields(t *testing.T) {
	gc := newTestGC()

	dict := gc.GetStatsAsDict()

	expectedKeys := []string{
		"minor_count", "major_count", "young_objects", "old_objects",
		"generational", "minor_threshold", "promotion_age",
	}

	for _, key := range expectedKeys {
		keyObj := &objects.String{Value: key}
		if _, ok := dict.Get(keyObj); !ok {
			t.Errorf("expected key '%s' in GetStatsAsDict result", key)
		}
	}
}

func TestCreateGCModuleHasNewFunctions(t *testing.T) {
	module := CreateGCModule()

	expectedFunctions := []string{
		"minor_collect", "major_collect", "set_promotion_age",
		"set_generational", "set_minor_threshold", "write_barrier",
	}

	for _, name := range expectedFunctions {
		if _, ok := module.Fields[name]; !ok {
			t.Errorf("expected function '%s' in GC module", name)
		}
	}
}

func TestConcurrentMinorAndMajor(t *testing.T) {
	gc := newTestGC()

	var wg sync.WaitGroup

	// Allocate objects from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				gc.Allocate(&objects.Integer{Value: int64(j)})
			}
		}()
	}

	wg.Wait()

	initialCount := len(gc.objects)

	SetRoots(nil, nil, nil, nil)

	gc.MinorCollect()
	gc.MajorCollect()

	if len(gc.objects) >= initialCount {
		t.Errorf("expected objects to be freed after GC, had %d, now %d", initialCount, len(gc.objects))
	}
}

func TestRebuildGenerationLists(t *testing.T) {
	gc := newTestGC()

	// Manually create objects in different generations
	youngObj := &GCObject{Object: &objects.Integer{Value: 1}, Generation: 0}
	oldObj := &GCObject{Object: &objects.Integer{Value: 2}, Generation: 1}

	gc.objects = []*GCObject{youngObj, oldObj}
	gc.youngObjects = []*GCObject{youngObj}
	gc.oldObjects = []*GCObject{oldObj}

	// Corrupt the generation lists
	gc.youngObjects = nil
	gc.oldObjects = nil

	gc.rebuildGenerationLists()

	if len(gc.youngObjects) != 1 || gc.youngObjects[0] != youngObj {
		t.Error("rebuildGenerationLists failed to rebuild youngObjects")
	}
	if len(gc.oldObjects) != 1 || gc.oldObjects[0] != oldObj {
		t.Error("rebuildGenerationLists failed to rebuild oldObjects")
	}
}

func TestAllocateDisabledGC(t *testing.T) {
	gc := newTestGC()
	gc.enabled = false

	obj := &objects.Integer{Value: 42}
	gcObj := gc.Allocate(obj)

	if gcObj == nil {
		t.Fatal("expected non-nil GCObject even when GC is disabled")
	}
	if gcObj.Generation != 0 {
		t.Errorf("expected Generation=0 even when disabled, got %d", gcObj.Generation)
	}
	// When disabled, object should not be tracked
	if len(gc.objects) != 0 {
		t.Errorf("expected 0 tracked objects when GC disabled, got %d", len(gc.objects))
	}
}

func TestMinorCollectDisabledGC(t *testing.T) {
	gc := newTestGC()
	gc.enabled = false

	obj := &objects.Integer{Value: 42}
	gc.Allocate(obj)

	// This should be a no-op
	gc.MinorCollect()

	if gc.minorCount != 0 {
		t.Errorf("expected minorCount=0 when GC disabled, got %d", gc.minorCount)
	}
}

func TestListReferencesTraversedInMinorGC(t *testing.T) {
	gc := newTestGC()

	// Create a list that references an integer
	innerObj := &objects.Integer{Value: 99}
	list := &objects.List{Elements: []objects.Object{innerObj}}

	gc.Allocate(list)
	gc.Allocate(innerObj)

	SetRoots(nil, []objects.Object{list}, nil, nil)

	if len(gc.youngObjects) != 2 {
		t.Fatalf("expected 2 young objects, got %d", len(gc.youngObjects))
	}

	gc.MinorCollect()

	// Both should survive and be promoted
	if len(gc.oldObjects) != 2 {
		t.Errorf("expected 2 old objects (both promoted), got %d", len(gc.oldObjects))
	}
}

func TestInstanceFieldsTraversedInMinorGC(t *testing.T) {
	gc := newTestGC()

	fieldValue := &objects.Integer{Value: 42}
	instance := &objects.Instance{
		Class:  &objects.Class{Name: "Test", Methods: make(map[string]objects.Object), Fields: make(map[string]objects.Object)},
		Fields: map[string]objects.Object{"x": fieldValue},
	}

	gc.Allocate(instance)
	gc.Allocate(fieldValue)

	SetRoots(nil, []objects.Object{instance}, nil, nil)

	gc.MinorCollect()

	// Both should survive and be promoted
	if len(gc.oldObjects) != 2 {
		t.Errorf("expected 2 old objects (instance + field value), got %d", len(gc.oldObjects))
	}
}
