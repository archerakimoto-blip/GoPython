package gc

import (
	"testing"
	"time"

	"github.com/go-py/go-python/pkg/objects"
)

func TestGenerationalGCAllocate(t *testing.T) {
	gc := &GarbageCollector{
		youngObjects:    make([]*GCObject, 0),
		oldObjects:      make([]*GCObject, 0),
		objectMap:       make(map[objects.Object]*GCObject),
		marked:          make(map[*GCObject]bool),
		enabled:         true,
		youngThreshold:  1024 * 256,
		oldThreshold:    1024 * 1024 * 4,
		promotionAge:    3,
		majorAfterMinor: 10,
	}

	// Allocate should put objects in young gen
	obj := &objects.Integer{Value: 42}
	gcObj := gc.Allocate(obj)

	if gcObj.Generation != YoungGen {
		t.Error("New object should be in young generation")
	}
	if len(gc.youngObjects) != 1 {
		t.Errorf("Expected 1 young object, got %d", len(gc.youngObjects))
	}
	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects, got %d", len(gc.oldObjects))
	}
}

func TestGenerationalGCMinorCollect(t *testing.T) {
	gc := &GarbageCollector{
		youngObjects:    make([]*GCObject, 0),
		oldObjects:      make([]*GCObject, 0),
		objectMap:       make(map[objects.Object]*GCObject),
		marked:          make(map[*GCObject]bool),
		enabled:         true,
		youngThreshold:  1024 * 256,
		oldThreshold:    1024 * 1024 * 4,
		promotionAge:    3,
		majorAfterMinor: 10,
	}

	// Allocate objects
	obj1 := &objects.Integer{Value: 1}
	obj2 := &objects.Integer{Value: 2}
	gc.Allocate(obj1)
	gc.Allocate(obj2)

	if len(gc.youngObjects) != 2 {
		t.Fatalf("Expected 2 young objects, got %d", len(gc.youngObjects))
	}

	// Set up roots so objects survive
	SetRoots(nil, []objects.Object{obj1, obj2}, nil, nil)

	// Minor collect should keep alive objects
	gc.MinorCollect()

	// Objects should still be in young gen (survived count < promotionAge)
	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects after minor GC, got %d", len(gc.youngObjects))
	}
}

func TestGenerationalGCPromotion(t *testing.T) {
	gc := &GarbageCollector{
		youngObjects:    make([]*GCObject, 0),
		oldObjects:      make([]*GCObject, 0),
		objectMap:       make(map[objects.Object]*GCObject),
		marked:          make(map[*GCObject]bool),
		enabled:         true,
		youngThreshold:  1024 * 256,
		oldThreshold:    1024 * 1024 * 4,
		promotionAge:    2,
		majorAfterMinor: 10,
	}

	// Allocate objects
	obj1 := &objects.Integer{Value: 1}
	gc.Allocate(obj1)

	// Set up roots
	SetRoots(nil, []objects.Object{obj1}, nil, nil)

	// Run minor GC promotionAge times
	for i := 0; i < gc.promotionAge; i++ {
		gc.MinorCollect()
	}

	// Object should be promoted to old gen
	if len(gc.oldObjects) != 1 {
		t.Errorf("Expected 1 old object after promotion, got %d", len(gc.oldObjects))
	}
	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects after promotion, got %d", len(gc.youngObjects))
	}
}

func TestGenerationalGCMajorCollect(t *testing.T) {
	gc := &GarbageCollector{
		youngObjects:    make([]*GCObject, 0),
		oldObjects:      make([]*GCObject, 0),
		objectMap:       make(map[objects.Object]*GCObject),
		marked:          make(map[*GCObject]bool),
		enabled:         true,
		youngThreshold:  1024 * 256,
		oldThreshold:    1024 * 1024 * 4,
		promotionAge:    2,
		majorAfterMinor: 10,
	}

	// Allocate and promote an object
	obj1 := &objects.Integer{Value: 1}
	gc.Allocate(obj1)
	SetRoots(nil, []objects.Object{obj1}, nil, nil)

	for i := 0; i < gc.promotionAge; i++ {
		gc.MinorCollect()
	}

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Remove root, major collect should free it
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC, got %d", len(gc.oldObjects))
	}
}

func TestGenerationalGCStats(t *testing.T) {
	gc := &GarbageCollector{
		youngObjects:    make([]*GCObject, 0),
		oldObjects:      make([]*GCObject, 0),
		objectMap:       make(map[objects.Object]*GCObject),
		marked:          make(map[*GCObject]bool),
		enabled:         true,
		youngThreshold:  1024 * 256,
		oldThreshold:    1024 * 1024 * 4,
		promotionAge:    3,
		majorAfterMinor: 10,
	}

	gc.Allocate(&objects.Integer{Value: 1})
	gc.Allocate(&objects.Integer{Value: 2})

	stats := gc.GetStats()
	if stats["young_object_count"].(int) != 2 {
		t.Errorf("Expected 2 young objects in stats, got %v", stats["young_object_count"])
	}
	if stats["old_object_count"].(int) != 0 {
		t.Errorf("Expected 0 old objects in stats, got %v", stats["old_object_count"])
	}
	if stats["minor_count"].(int64) != 0 {
		t.Errorf("Expected 0 minor collections, got %v", stats["minor_count"])
	}
}

func TestGenerationalGCModule(t *testing.T) {
	module := CreateGCModule()
	if module.Name != "gc" {
		t.Errorf("Expected module name 'gc', got '%s'", module.Name)
	}

	// Check new functions exist
	newFuncs := []string{"minor_collect", "major_collect", "set_young_threshold", "set_old_threshold", "set_promotion_age"}
	for _, fn := range newFuncs {
		if _, ok := module.Fields[fn]; !ok {
			t.Errorf("Missing function: %s", fn)
		}
	}

	// Check backward compatible functions exist
	oldFuncs := []string{"collect", "enable", "disable", "get_stats", "print_stats", "set_threshold", "set_verbose"}
	for _, fn := range oldFuncs {
		if _, ok := module.Fields[fn]; !ok {
			t.Errorf("Missing backward-compatible function: %s", fn)
		}
	}
}

// ==================== Additional GC Tests ====================

func newTestGC() *GarbageCollector {
	return &GarbageCollector{
		youngObjects:    make([]*GCObject, 0),
		oldObjects:      make([]*GCObject, 0),
		objectMap:       make(map[objects.Object]*GCObject),
		marked:          make(map[*GCObject]bool),
		enabled:         true,
		youngThreshold:  1024 * 256,
		oldThreshold:    1024 * 1024 * 4,
		promotionAge:    3,
		majorAfterMinor: 10,
		rememberedSet:   make([]*GCObject, 0),
	}
}

func TestGCEnableDisable(t *testing.T) {
	gc := newTestGC()

	if !gc.enabled {
		t.Error("Expected GC to be enabled by default")
	}

	gc.Enable(false)
	if gc.enabled {
		t.Error("Expected GC to be disabled after Enable(false)")
	}

	// When disabled, Allocate should not add to young objects
	obj := &objects.Integer{Value: 42}
	gcObj := gc.Allocate(obj)
	if gcObj == nil {
		t.Fatal("Expected non-nil GCObject even when disabled")
	}
	if gcObj.Generation != YoungGen {
		t.Error("Expected YoungGen even when disabled")
	}
	// The object should NOT be tracked in the internal lists
	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects when disabled, got %d", len(gc.youngObjects))
	}

	gc.Enable(true)
	if !gc.enabled {
		t.Error("Expected GC to be enabled after Enable(true)")
	}
}

func TestGCDisabledCollect(t *testing.T) {
	gc := newTestGC()
	gc.Enable(false)

	// Collect should be a no-op when disabled
	gc.MinorCollect()
	gc.MajorCollect()
	gc.Collect()
}

func TestGCWriteBarrier(t *testing.T) {
	gc := newTestGC()

	// Create an old gen container and a young gen contained object
	oldObj := &objects.Integer{Value: 1}
	youngObj := &objects.Integer{Value: 2}

	oldGCObj := gc.Allocate(oldObj)
	youngGCObj := gc.Allocate(youngObj)

	// Manually promote the old object
	oldGCObj.Generation = OldGen
	gc.youngObjects = []*GCObject{youngGCObj}
	gc.oldObjects = []*GCObject{oldGCObj}

	// WriteBarrier should add old object to remembered set
	gc.WriteBarrier(oldGCObj, youngGCObj)

	if len(gc.rememberedSet) != 1 {
		t.Errorf("Expected remembered set size 1, got %d", len(gc.rememberedSet))
	}
	if gc.rememberedSet[0] != oldGCObj {
		t.Error("Expected oldGCObj in remembered set")
	}
}

func TestGCWriteBarrierNilArgs(t *testing.T) {
	gc := newTestGC()

	// Should not panic with nil args
	gc.WriteBarrier(nil, nil)
	gc.WriteBarrier(nil, &GCObject{})
	gc.WriteBarrier(&GCObject{}, nil)
}

func TestGCWriteBarrierDuplicate(t *testing.T) {
	gc := newTestGC()

	oldObj := &objects.Integer{Value: 1}
	youngObj := &objects.Integer{Value: 2}

	oldGCObj := gc.Allocate(oldObj)
	youngGCObj := gc.Allocate(youngObj)

	oldGCObj.Generation = OldGen
	gc.youngObjects = []*GCObject{youngGCObj}
	gc.oldObjects = []*GCObject{oldGCObj}

	// Call WriteBarrier twice - should only add once
	gc.WriteBarrier(oldGCObj, youngGCObj)
	gc.WriteBarrier(oldGCObj, youngGCObj)

	if len(gc.rememberedSet) != 1 {
		t.Errorf("Expected remembered set size 1 (no duplicates), got %d", len(gc.rememberedSet))
	}
}

func TestGCWriteBarrierYoungToYoung(t *testing.T) {
	gc := newTestGC()

	obj1 := &objects.Integer{Value: 1}
	obj2 := &objects.Integer{Value: 2}

	gcObj1 := gc.Allocate(obj1)
	gcObj2 := gc.Allocate(obj2)

	// Both are young gen - WriteBarrier should NOT add to remembered set
	gc.WriteBarrier(gcObj1, gcObj2)

	if len(gc.rememberedSet) != 0 {
		t.Errorf("Expected remembered set size 0 (young->young), got %d", len(gc.rememberedSet))
	}
}

func TestGCMinorCollectUnreachable(t *testing.T) {
	gc := newTestGC()

	// Allocate objects without setting roots
	obj1 := &objects.Integer{Value: 1}
	obj2 := &objects.Integer{Value: 2}
	gc.Allocate(obj1)
	gc.Allocate(obj2)

	SetRoots(nil, nil, nil, nil)

	gc.MinorCollect()

	// Unreachable objects should be freed
	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects after collecting unreachable, got %d", len(gc.youngObjects))
	}
}

func TestGCMajorCollectUnreachable(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 2

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Promote to old gen
	for i := 0; i < gc.promotionAge; i++ {
		gc.MinorCollect()
	}

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Remove root and major collect
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC, got %d", len(gc.oldObjects))
	}
	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects after major GC, got %d", len(gc.youngObjects))
	}
}

func TestGCRegister(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 42}
	gcObj := gc.Register(obj)

	if gcObj == nil {
		t.Fatal("Expected non-nil GCObject from Register")
	}
	if gcObj.Generation != YoungGen {
		t.Error("Expected YoungGen from Register")
	}
	if len(gc.youngObjects) != 1 {
		t.Errorf("Expected 1 young object, got %d", len(gc.youngObjects))
	}
}

func TestGCCollectBackwardCompat(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, nil, nil, nil)

	// Collect() should behave like MinorCollect()
	gc.Collect()

	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects after Collect(), got %d", len(gc.youngObjects))
	}
}

func TestGCGetStatsDetailed(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	gc.MinorCollect()

	stats := gc.GetStats()

	// Check all expected keys exist
	expectedKeys := []string{
		"allocated_bytes", "freed_bytes", "collection_count",
		"minor_count", "major_count", "object_count",
		"young_object_count", "old_object_count",
		"young_allocated", "old_allocated",
		"young_threshold", "old_threshold",
		"promotion_age", "remembered_set_size",
		"enabled", "threshold", "total_pause_time",
	}

	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("Missing stats key: %s", key)
		}
	}

	if stats["minor_count"].(int64) != 1 {
		t.Errorf("Expected 1 minor collection, got %v", stats["minor_count"])
	}
	if stats["enabled"].(bool) != true {
		t.Error("Expected enabled=true in stats")
	}
}

func TestGCGetStatsAsDict(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	dict := gc.GetStatsAsDict()
	if dict == nil {
		t.Fatal("Expected non-nil dict from GetStatsAsDict")
	}

	// Check that the dict has the expected number of keys
	// (17 keys as defined in GetStatsAsDict)
	if len(dict.Pairs) != 17 {
		t.Errorf("Expected 17 keys in stats dict, got %d", len(dict.Pairs))
	}

	// Verify a few specific keys exist using the proper lookup
	allocatedKey := &objects.String{Value: "allocated_bytes"}
	if val, ok := dict.Get(allocatedKey); !ok {
		t.Error("Missing 'allocated_bytes' key in stats dict")
	} else {
		if val.Type() != objects.INTEGER_OBJ {
			t.Errorf("Expected integer for 'allocated_bytes', got %s", val.Type())
		}
	}

	enabledKey := &objects.String{Value: "enabled"}
	if val, ok := dict.Get(enabledKey); !ok {
		t.Error("Missing 'enabled' key in stats dict")
	} else {
		if val.Type() != objects.BOOLEAN_OBJ {
			t.Errorf("Expected boolean for 'enabled', got %s", val.Type())
		}
	}
}

func TestGCSetThreshold(t *testing.T) {
	gc := newTestGC()

	gc.SetThreshold(2048)
	if gc.threshold != 2048 {
		t.Errorf("Expected threshold 2048, got %d", gc.threshold)
	}
}

func TestGCSetYoungThreshold(t *testing.T) {
	gc := newTestGC()

	gc.SetYoungThreshold(512)
	if gc.youngThreshold != 512 {
		t.Errorf("Expected young threshold 512, got %d", gc.youngThreshold)
	}
}

func TestGCSetOldThreshold(t *testing.T) {
	gc := newTestGC()

	gc.SetOldThreshold(8192)
	if gc.oldThreshold != 8192 {
		t.Errorf("Expected old threshold 8192, got %d", gc.oldThreshold)
	}
}

func TestGCSetPromotionAge(t *testing.T) {
	gc := newTestGC()

	gc.SetPromotionAge(5)
	if gc.promotionAge != 5 {
		t.Errorf("Expected promotion age 5, got %d", gc.promotionAge)
	}
}

func TestGCSetVerbose(t *testing.T) {
	gc := newTestGC()

	gc.SetVerbose(true)
	if !gc.verbose {
		t.Error("Expected verbose=true")
	}

	gc.SetVerbose(false)
	if gc.verbose {
		t.Error("Expected verbose=false")
	}
}

func TestGCAddFinalizer(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	finalized := false
	gc.AddFinalizer(obj, func(o objects.Object) {
		finalized = true
	})

	// Remove root and collect - finalizer should be called
	SetRoots(nil, nil, nil, nil)
	gc.MinorCollect()

	if !finalized {
		t.Error("Expected finalizer to be called")
	}
}

func TestGCAddFinalizerNonExistent(t *testing.T) {
	gc := newTestGC()

	// Should not panic when adding finalizer to non-tracked object
	obj := &objects.Integer{Value: 1}
	gc.AddFinalizer(obj, func(o objects.Object) {})
}

func TestGCAllocateDifferentTypes(t *testing.T) {
	gc := newTestGC()

	tests := []struct {
		name string
		obj  objects.Object
	}{
		{"integer", &objects.Integer{Value: 42}},
		{"float", &objects.Float{Value: 3.14}},
		{"boolean", &objects.Boolean{Value: true}},
		{"string", &objects.String{Value: "hello"}},
		{"list", &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}}},
		{"dict", objects.NewDict()},
		{"set", objects.NewSet()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcObj := gc.Allocate(tt.obj)
			if gcObj == nil {
				t.Fatal("Expected non-nil GCObject")
			}
			if gcObj.Size <= 0 {
				t.Errorf("Expected positive size, got %d", gcObj.Size)
			}
		})
	}
}

func TestGCPackageLevelFunctions(t *testing.T) {
	// Test package-level convenience functions
	// These use the singleton, so we just verify they don't panic
	Enable()
	Disable()
	Collect()
	MinorCollect()
	MajorCollect()
	stats := GetStats()
	if stats == nil {
		t.Error("Expected non-nil stats from package-level GetStats")
	}
}

func TestGCModuleFunctions(t *testing.T) {
	module := CreateGCModule()

	// Test collect function
	collectFn := module.Fields["collect"].(*objects.Builtin)
	result := collectFn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from collect, got %v", result)
	}

	// Test enable function
	enableFn := module.Fields["enable"].(*objects.Builtin)
	result = enableFn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from enable, got %v", result)
	}

	// Test disable function
	disableFn := module.Fields["disable"].(*objects.Builtin)
	result = disableFn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from disable, got %v", result)
	}

	// Test get_stats function
	getStatsFn := module.Fields["get_stats"].(*objects.Builtin)
	result = getStatsFn.Fn()
	if result == nil {
		t.Error("Expected non-nil result from get_stats")
	}

	// Test set_threshold function
	setThresholdFn := module.Fields["set_threshold"].(*objects.Builtin)
	result = setThresholdFn.Fn(&objects.Integer{Value: 2048})
	if result != objects.None_ {
		t.Errorf("Expected None from set_threshold, got %v", result)
	}

	// Test set_verbose function
	setVerboseFn := module.Fields["set_verbose"].(*objects.Builtin)
	result = setVerboseFn.Fn(&objects.Boolean{Value: true})
	if result != objects.None_ {
		t.Errorf("Expected None from set_verbose, got %v", result)
	}

	// Re-enable for other tests
	Enable()
}

func TestGCMinorCollectWithList(t *testing.T) {
	gc := newTestGC()

	inner := &objects.Integer{Value: 1}
	lst := &objects.List{Elements: []objects.Object{inner}}
	gc.Allocate(inner)
	gc.Allocate(lst)

	// Only the list is a root
	SetRoots(nil, []objects.Object{lst}, nil, nil)

	gc.MinorCollect()

	// Both should survive since list references inner
	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects (list + inner), got %d", len(gc.youngObjects))
	}
}

func TestGCMajorCollectWithDict(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 2

	val := &objects.Integer{Value: 42}
	d := objects.NewDict()
	d.Set(&objects.String{Value: "key"}, val)
	gc.Allocate(val)
	gc.Allocate(d)

	SetRoots(nil, []objects.Object{d}, nil, nil)

	// Promote to old gen
	for i := 0; i < gc.promotionAge; i++ {
		gc.MinorCollect()
	}

	// Dict should be in old gen
	if len(gc.oldObjects) == 0 {
		t.Error("Expected at least 1 old object after promotion")
	}

	// Remove root and major collect
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC, got %d", len(gc.oldObjects))
	}
}

func TestGCSetRoots(t *testing.T) {
	obj1 := &objects.Integer{Value: 1}
	obj2 := &objects.Integer{Value: 2}
	obj3 := &objects.Integer{Value: 3}
	obj4 := &objects.Integer{Value: 4}

	SetRoots(
		[]objects.Object{obj1},
		[]objects.Object{obj2},
		[]objects.Object{obj3},
		[]objects.Object{obj4},
	)

	regs := getRegisterContents()
	if len(regs) != 1 || regs[0] != obj1 {
		t.Error("Register contents not set correctly")
	}

	stack := getStackContents()
	if len(stack) != 1 || stack[0] != obj2 {
		t.Error("Stack contents not set correctly")
	}

	globals := getGlobalContents()
	if len(globals) != 1 || globals[0] != obj3 {
		t.Error("Global contents not set correctly")
	}

	frames := getFrameContents()
	if len(frames) != 1 || frames[0] != obj4 {
		t.Error("Frame contents not set correctly")
	}

	// Clean up
	SetRoots(nil, nil, nil, nil)
}

func TestEstimateSize(t *testing.T) {
	tests := []struct {
		name     string
		obj      objects.Object
		minSize  int64
	}{
		{"integer", &objects.Integer{Value: 42}, 8},
		{"float", &objects.Float{Value: 3.14}, 8},
		{"boolean true", &objects.Boolean{Value: true}, 1},
		{"boolean false", &objects.Boolean{Value: false}, 1},
		{"string empty", &objects.String{Value: ""}, 16},
		{"string hello", &objects.String{Value: "hello"}, 21},
		{"list empty", &objects.List{Elements: []objects.Object{}}, 24},
		{"list with items", &objects.List{Elements: []objects.Object{&objects.Integer{Value: 1}}}, 32},
		{"dict empty", objects.NewDict(), 48},
		{"set empty", objects.NewSet(), 48},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := estimateSize(tt.obj)
			if size < tt.minSize {
				t.Errorf("Expected size >= %d, got %d", tt.minSize, size)
			}
		})
	}
}

func TestGCPromotionAgeOne(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// After just 1 minor GC, object should be promoted
	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Errorf("Expected 1 old object with promotionAge=1, got %d", len(gc.oldObjects))
	}
	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects with promotionAge=1, got %d", len(gc.youngObjects))
	}
}

func TestGCGetGC(t *testing.T) {
	gc1 := GetGC()
	gc2 := GetGC()
	if gc1 != gc2 {
		t.Error("GetGC should return the same singleton instance")
	}
}

func TestGCModuleSetYoungThreshold(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["set_young_threshold"].(*objects.Builtin)
	result := fn.Fn(&objects.Integer{Value: 512})
	if result != objects.None_ {
		t.Errorf("Expected None from set_young_threshold, got %v", result)
	}
}

func TestGCModuleSetOldThreshold(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["set_old_threshold"].(*objects.Builtin)
	result := fn.Fn(&objects.Integer{Value: 8192})
	if result != objects.None_ {
		t.Errorf("Expected None from set_old_threshold, got %v", result)
	}
}

func TestGCModuleSetPromotionAge(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["set_promotion_age"].(*objects.Builtin)
	result := fn.Fn(&objects.Integer{Value: 5})
	if result != objects.None_ {
		t.Errorf("Expected None from set_promotion_age, got %v", result)
	}
}

func TestGCModuleMinorCollect(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["minor_collect"].(*objects.Builtin)
	result := fn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from minor_collect, got %v", result)
	}
}

func TestGCModuleMajorCollect(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["major_collect"].(*objects.Builtin)
	result := fn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from major_collect, got %v", result)
	}
}

func TestGCModulePrintStats(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["print_stats"].(*objects.Builtin)
	result := fn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from print_stats, got %v", result)
	}
}

func TestGCModuleSetThresholdNoArgs(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["set_threshold"].(*objects.Builtin)
	// Should not panic with no args
	result := fn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from set_threshold with no args, got %v", result)
	}
}

func TestGCModuleSetVerboseNoArgs(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["set_verbose"].(*objects.Builtin)
	// Should not panic with no args
	result := fn.Fn()
	if result != objects.None_ {
		t.Errorf("Expected None from set_verbose with no args, got %v", result)
	}
}

func TestGCModuleSetThresholdWrongType(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["set_threshold"].(*objects.Builtin)
	// Should not panic with wrong type
	result := fn.Fn(&objects.String{Value: "not a number"})
	if result != objects.None_ {
		t.Errorf("Expected None from set_threshold with wrong type, got %v", result)
	}
}

func TestGCModuleSetVerboseWrongType(t *testing.T) {
	module := CreateGCModule()
	fn := module.Fields["set_verbose"].(*objects.Builtin)
	// Should not panic with wrong type
	result := fn.Fn(&objects.String{Value: "not a bool"})
	if result != objects.None_ {
		t.Errorf("Expected None from set_verbose with wrong type, got %v", result)
	}
}

func TestGCMultipleMinorCollections(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 3

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Run 3 minor collections - should promote
	for i := 0; i < 3; i++ {
		gc.MinorCollect()
	}

	stats := gc.GetStats()
	if stats["minor_count"].(int64) != 3 {
		t.Errorf("Expected 3 minor collections, got %v", stats["minor_count"])
	}
	if stats["old_object_count"].(int) != 1 {
		t.Errorf("Expected 1 old object, got %v", stats["old_object_count"])
	}
}

func TestGCAllocateString(t *testing.T) {
	gc := newTestGC()

	obj := &objects.String{Value: "hello world"}
	gcObj := gc.Allocate(obj)

	if gcObj.Size <= 0 {
		t.Errorf("Expected positive size for string, got %d", gcObj.Size)
	}
}

func TestGCAllocateList(t *testing.T) {
	gc := newTestGC()

	lst := &objects.List{Elements: []objects.Object{
		&objects.Integer{Value: 1},
		&objects.Integer{Value: 2},
		&objects.Integer{Value: 3},
	}}
	gcObj := gc.Allocate(lst)

	if gcObj.Size <= 0 {
		t.Errorf("Expected positive size for list, got %d", gcObj.Size)
	}
}

func TestGCAllocateTuple(t *testing.T) {
	gc := newTestGC()

	tup := &objects.Tuple{Elements: []objects.Object{
		&objects.Integer{Value: 1},
		&objects.Integer{Value: 2},
	}}
	gcObj := gc.Allocate(tup)

	if gcObj.Size <= 0 {
		t.Errorf("Expected positive size for tuple, got %d", gcObj.Size)
	}
}

func TestGCAllocateSet(t *testing.T) {
	gc := newTestGC()

	s := objects.NewSet()
	s.Add(&objects.Integer{Value: 1})
	gcObj := gc.Allocate(s)

	if gcObj.Size <= 0 {
		t.Errorf("Expected positive size for set, got %d", gcObj.Size)
	}
}

func TestGCAllocateDict(t *testing.T) {
	gc := newTestGC()

	d := objects.NewDict()
	d.Set(&objects.String{Value: "key"}, &objects.Integer{Value: 42})
	gcObj := gc.Allocate(d)

	if gcObj.Size <= 0 {
		t.Errorf("Expected positive size for dict, got %d", gcObj.Size)
	}
}

func TestGCFinalizerOnOldGen(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 2

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	finalized := false
	gc.AddFinalizer(obj, func(o objects.Object) {
		finalized = true
	})

	// Promote to old gen
	for i := 0; i < gc.promotionAge; i++ {
		gc.MinorCollect()
	}

	// Remove root and major collect
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if !finalized {
		t.Error("Expected finalizer to be called on old gen object")
	}
}

func TestGCRememberedSetClearedOnMajor(t *testing.T) {
	gc := newTestGC()

	oldObj := &objects.Integer{Value: 1}
	youngObj := &objects.Integer{Value: 2}

	oldGCObj := gc.Allocate(oldObj)
	youngGCObj := gc.Allocate(youngObj)

	oldGCObj.Generation = OldGen
	gc.youngObjects = []*GCObject{youngGCObj}
	gc.oldObjects = []*GCObject{oldGCObj}

	// Add to remembered set
	gc.WriteBarrier(oldGCObj, youngGCObj)
	if len(gc.rememberedSet) != 1 {
		t.Fatalf("Expected remembered set size 1, got %d", len(gc.rememberedSet))
	}

	// Major collect should clear remembered set
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.rememberedSet) != 0 {
		t.Errorf("Expected remembered set to be cleared after major GC, got %d", len(gc.rememberedSet))
	}
}

// ==================== Additional GC Tests for 90%+ Coverage ====================

func TestGCMajorCollectLocked(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 2
	gc.majorAfterMinor = 3

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Promote to old gen
	for i := 0; i < gc.promotionAge; i++ {
		gc.MinorCollect()
	}

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Run enough minor GCs to trigger majorCollectLocked via minorSinceMajor >= majorAfterMinor
	for i := 0; i < int(gc.majorAfterMinor); i++ {
		gc.MinorCollect()
	}

	// Object should still be alive (rooted)
	if len(gc.oldObjects) != 1 {
		t.Errorf("Expected 1 old object after majorCollectLocked (rooted), got %d", len(gc.oldObjects))
	}

	// Remove root, then trigger another majorCollectLocked
	SetRoots(nil, nil, nil, nil)
	// Need to allocate something to trigger minor, which will trigger major
	obj2 := &objects.Integer{Value: 2}
	gc.Allocate(obj2)
	// Run enough minors to trigger major
	for i := 0; i < int(gc.majorAfterMinor); i++ {
		gc.MinorCollect()
	}

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after majorCollectLocked (unrooted), got %d", len(gc.oldObjects))
	}
}

func TestGCMajorCollectLockedViaOldThreshold(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1
	gc.oldThreshold = 1 // Set very low so oldAllocated >= oldThreshold triggers major

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Promote to old gen
	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Remove root and allocate a new object to trigger minor collect
	// which will see oldAllocated >= oldThreshold and call majorCollectLocked
	SetRoots(nil, nil, nil, nil)
	obj2 := &objects.Integer{Value: 2}
	gc.Allocate(obj2)
	gc.MinorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after majorCollectLocked via oldThreshold, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkReferencesMinorDict(t *testing.T) {
	gc := newTestGC()

	inner := &objects.Integer{Value: 42}
	d := objects.NewDict()
	d.Set(&objects.String{Value: "key"}, inner)
	gc.Allocate(inner)
	gc.Allocate(d)

	// Only dict is a root
	SetRoots(nil, []objects.Object{d}, nil, nil)

	gc.MinorCollect()

	// Both should survive since dict references inner
	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects (dict + inner), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkReferencesMinorSet(t *testing.T) {
	gc := newTestGC()

	inner := &objects.Integer{Value: 42}
	s := objects.NewSet()
	s.Add(inner)
	gc.Allocate(inner)
	gc.Allocate(s)

	SetRoots(nil, []objects.Object{s}, nil, nil)

	gc.MinorCollect()

	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects (set + inner), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkReferencesMinorInstance(t *testing.T) {
	gc := newTestGC()

	inner := &objects.Integer{Value: 42}
	cls := &objects.Class{
		Name:    "MyClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	inst := &objects.Instance{
		Class:  cls,
		Fields: map[string]objects.Object{"attr": inner},
	}
	gc.Allocate(inner)
	gc.Allocate(inst)

	SetRoots(nil, []objects.Object{inst}, nil, nil)

	gc.MinorCollect()

	// Instance and inner should survive
	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects (instance + inner), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkReferencesMinorInstanceWithSlots(t *testing.T) {
	gc := newTestGC()

	inner := &objects.Integer{Value: 42}
	cls := &objects.Class{
		Name:    "MyClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	inst := &objects.Instance{
		Class:      cls,
		Fields:     map[string]objects.Object{},
		SlotValues: []objects.Object{inner},
	}
	gc.Allocate(inner)
	gc.Allocate(inst)

	SetRoots(nil, []objects.Object{inst}, nil, nil)

	gc.MinorCollect()

	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects (instance + slot value), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkReferencesMinorInstanceSlotNil(t *testing.T) {
	gc := newTestGC()

	cls := &objects.Class{
		Name:    "MyClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	inst := &objects.Instance{
		Class:      cls,
		Fields:     map[string]objects.Object{},
		SlotValues: []objects.Object{nil},
	}
	gc.Allocate(inst)

	SetRoots(nil, []objects.Object{inst}, nil, nil)

	gc.MinorCollect()

	if len(gc.youngObjects) != 1 {
		t.Errorf("Expected 1 young object (instance with nil slot), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkReferencesMinorClassWithSuper(t *testing.T) {
	gc := newTestGC()

	superCls := &objects.Class{
		Name:    "SuperClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	cls := &objects.Class{
		Name:       "MyClass",
		Methods:    make(map[string]objects.Object),
		Fields:     make(map[string]objects.Object),
		SuperClass: superCls,
	}
	gc.Allocate(superCls)
	gc.Allocate(cls)

	SetRoots(nil, []objects.Object{cls}, nil, nil)

	gc.MinorCollect()

	// Both class and superclass should survive
	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects (class + superclass), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkReferencesMinorTuple(t *testing.T) {
	gc := newTestGC()

	inner := &objects.Integer{Value: 42}
	tup := &objects.Tuple{Elements: []objects.Object{inner}}
	gc.Allocate(inner)
	gc.Allocate(tup)

	SetRoots(nil, []objects.Object{tup}, nil, nil)

	gc.MinorCollect()

	if len(gc.youngObjects) != 2 {
		t.Errorf("Expected 2 young objects (tuple + inner), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkFromRememberedSet(t *testing.T) {
	gc := newTestGC()

	// Create old gen object that references young gen object via a list
	inner := &objects.Integer{Value: 42}
	lst := &objects.List{Elements: []objects.Object{inner}}

	oldGCObj := &GCObject{
		Object:     lst,
		Size:       estimateSize(lst),
		Generation: OldGen,
	}
	youngGCObj := &GCObject{
		Object:     inner,
		Size:       estimateSize(inner),
		Generation: YoungGen,
	}

	gc.oldObjects = []*GCObject{oldGCObj}
	gc.youngObjects = []*GCObject{youngGCObj}
	gc.objectMap[lst] = oldGCObj
	gc.objectMap[inner] = youngGCObj
	gc.rememberedSet = []*GCObject{oldGCObj}

	// No roots set - the young object should still survive because
	// markFromRememberedSet marks it via the old gen list
	SetRoots(nil, nil, nil, nil)
	gc.MinorCollect()

	// The inner object should survive because it's referenced from the remembered set
	if len(gc.youngObjects) != 1 {
		t.Errorf("Expected 1 young object (survived via remembered set), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkFromRememberedSetNilObject(t *testing.T) {
	gc := newTestGC()

	// Add an entry with nil Object to remembered set
	nilObj := &GCObject{Object: nil, Generation: OldGen}
	gc.rememberedSet = []*GCObject{nilObj}

	SetRoots(nil, nil, nil, nil)
	gc.MinorCollect()
	// Should not panic
}

func TestGCMarkRootsMinorWithRegister(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	// Set register root
	SetRoots([]objects.Object{obj}, nil, nil, nil)

	gc.MinorCollect()

	if len(gc.youngObjects) != 1 {
		t.Errorf("Expected 1 young object (rooted via register), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkRootsMinorWithGlobal(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	SetRoots(nil, nil, []objects.Object{obj}, nil)

	gc.MinorCollect()

	if len(gc.youngObjects) != 1 {
		t.Errorf("Expected 1 young object (rooted via global), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkRootsMinorWithFrame(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	SetRoots(nil, nil, nil, []objects.Object{obj})

	gc.MinorCollect()

	if len(gc.youngObjects) != 1 {
		t.Errorf("Expected 1 young object (rooted via frame), got %d", len(gc.youngObjects))
	}
}

func TestGCMarkRootsWithRegister(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	// Set register root
	SetRoots([]objects.Object{obj}, nil, nil, nil)

	// Promote to old gen
	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Remove root and major collect
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkRootsWithGlobal(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	SetRoots(nil, nil, []objects.Object{obj}, nil)

	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkRootsWithFrame(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	SetRoots(nil, nil, nil, []objects.Object{obj})

	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkReferencesDict(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	inner := &objects.Integer{Value: 42}
	d := objects.NewDict()
	d.Set(&objects.String{Value: "key"}, inner)
	gc.Allocate(inner)
	gc.Allocate(d)

	SetRoots(nil, []objects.Object{d}, nil, nil)

	// Promote to old gen
	gc.MinorCollect()

	if len(gc.oldObjects) < 1 {
		t.Fatal("Expected at least 1 old object")
	}

	// Remove root and major collect
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC with dict, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkReferencesSet(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	inner := &objects.Integer{Value: 42}
	s := objects.NewSet()
	s.Add(inner)
	gc.Allocate(inner)
	gc.Allocate(s)

	SetRoots(nil, []objects.Object{s}, nil, nil)

	gc.MinorCollect()

	if len(gc.oldObjects) < 1 {
		t.Fatal("Expected at least 1 old object")
	}

	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC with set, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkReferencesInstance(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	inner := &objects.Integer{Value: 42}
	cls := &objects.Class{
		Name:    "MyClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	inst := &objects.Instance{
		Class:  cls,
		Fields: map[string]objects.Object{"attr": inner},
	}
	gc.Allocate(inner)
	gc.Allocate(inst)

	SetRoots(nil, []objects.Object{inst}, nil, nil)

	gc.MinorCollect()

	if len(gc.oldObjects) < 1 {
		t.Fatal("Expected at least 1 old object")
	}

	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC with instance, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkReferencesInstanceSlots(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	inner := &objects.Integer{Value: 42}
	cls := &objects.Class{
		Name:    "MyClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	inst := &objects.Instance{
		Class:      cls,
		Fields:     map[string]objects.Object{},
		SlotValues: []objects.Object{inner},
	}
	gc.Allocate(inner)
	gc.Allocate(inst)

	SetRoots(nil, []objects.Object{inst}, nil, nil)

	gc.MinorCollect()

	if len(gc.oldObjects) < 1 {
		t.Fatal("Expected at least 1 old object")
	}

	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC with instance slots, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkReferencesInstanceSlotNil(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	cls := &objects.Class{
		Name:    "MyClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	inst := &objects.Instance{
		Class:      cls,
		Fields:     map[string]objects.Object{},
		SlotValues: []objects.Object{nil},
	}
	gc.Allocate(inst)

	SetRoots(nil, []objects.Object{inst}, nil, nil)

	gc.MinorCollect()

	// Should not panic with nil slot values
	if len(gc.oldObjects) < 1 {
		t.Fatal("Expected at least 1 old object")
	}
}

func TestGCMarkReferencesClassWithSuper(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1

	superCls := &objects.Class{
		Name:    "SuperClass",
		Methods: make(map[string]objects.Object),
		Fields:  make(map[string]objects.Object),
	}
	cls := &objects.Class{
		Name:       "MyClass",
		Methods:    make(map[string]objects.Object),
		Fields:     make(map[string]objects.Object),
		SuperClass: superCls,
	}
	gc.Allocate(superCls)
	gc.Allocate(cls)

	SetRoots(nil, []objects.Object{cls}, nil, nil)

	gc.MinorCollect()

	if len(gc.oldObjects) < 1 {
		t.Fatal("Expected at least 1 old object")
	}

	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after major GC with class+super, got %d", len(gc.oldObjects))
	}
}

func TestGCMarkObjectAlreadyMarked(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gcObj := gc.Allocate(obj)

	// Manually mark it
	gc.marked[gcObj] = true

	// markObject should skip already-marked objects
	gc.markObject(obj)
	// No panic and no double-marking
}

func TestGCMarkObjectNotInMap(t *testing.T) {
	gc := newTestGC()

	// Create an object not in the objectMap
	obj := &objects.Integer{Value: 1}

	// markObject should skip objects not in the map
	gc.markObject(obj)
	// No panic
}

func TestGCMarkObjectNil(t *testing.T) {
	gc := newTestGC()

	// markObject should handle nil
	gc.markObject(nil)
	// No panic
}

func TestGCMarkObjectMinorNil(t *testing.T) {
	gc := newTestGC()

	// markObjectMinor should handle nil
	gc.markObjectMinor(nil)
	// No panic
}

func TestGCMarkObjectMinorOldGen(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gcObj := gc.Allocate(obj)
	gcObj.Generation = OldGen

	// markObjectMinor should skip old gen objects
	gc.markObjectMinor(obj)
	if gc.marked[gcObj] {
		t.Error("Old gen object should not be marked by markObjectMinor")
	}
}

func TestGCSweepOldVerbose(t *testing.T) {
	gc := newTestGC()
	gc.verbose = true
	gc.promotionAge = 1

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Promote
	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Fatalf("Expected 1 old object, got %d", len(gc.oldObjects))
	}

	// Remove root and major collect (verbose mode)
	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if len(gc.oldObjects) != 0 {
		t.Errorf("Expected 0 old objects after verbose major GC, got %d", len(gc.oldObjects))
	}
}

func TestGCAllocateVerbose(t *testing.T) {
	gc := newTestGC()
	gc.verbose = true

	obj := &objects.Integer{Value: 1}
	gcObj := gc.Allocate(obj)

	if gcObj == nil {
		t.Fatal("Expected non-nil GCObject in verbose mode")
	}
}

func TestGCAllocateTriggerMinorCollect(t *testing.T) {
	gc := newTestGC()
	gc.youngThreshold = 1 // Very low threshold to trigger minor collect

	obj1 := &objects.Integer{Value: 1}
	gc.Allocate(obj1)

	// The next allocation should trigger a minor collect via goroutine
	// We just verify it doesn't panic
	obj2 := &objects.Integer{Value: 2}
	gc.Allocate(obj2)

	// Give time for the goroutine to finish
	time.Sleep(50 * time.Millisecond)
}

func TestGCAllocateTriggerMajorCollect(t *testing.T) {
	gc := newTestGC()
	gc.promotionAge = 1
	gc.oldThreshold = 1 // Very low to trigger major
	gc.youngThreshold = 1024 * 1024 // High to not trigger minor

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	// Promote to old gen
	gc.MinorCollect()

	// Now oldAllocated >= oldThreshold, next allocation should trigger major via goroutine
	obj2 := &objects.Integer{Value: 2}
	gc.Allocate(obj2)

	// Give time for the goroutine to finish
	time.Sleep(50 * time.Millisecond)
}

func TestGCSweepYoungVerbose(t *testing.T) {
	gc := newTestGC()
	gc.verbose = true

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, nil, nil, nil)

	// Minor collect in verbose mode should log freed objects
	gc.MinorCollect()

	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects, got %d", len(gc.youngObjects))
	}
}

func TestGCSweepYoungVerbosePromotion(t *testing.T) {
	gc := newTestGC()
	gc.verbose = true
	gc.promotionAge = 1

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	gc.MinorCollect()

	if len(gc.oldObjects) != 1 {
		t.Errorf("Expected 1 old object (promoted in verbose mode), got %d", len(gc.oldObjects))
	}
}

func TestGCMinorCollectVerbose(t *testing.T) {
	gc := newTestGC()
	gc.verbose = true

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	gc.MinorCollect()

	if gc.minorCount != 1 {
		t.Errorf("Expected 1 minor collection, got %d", gc.minorCount)
	}
}

func TestGCMajorCollectVerbose(t *testing.T) {
	gc := newTestGC()
	gc.verbose = true
	gc.promotionAge = 1

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)
	SetRoots(nil, []objects.Object{obj}, nil, nil)

	gc.MinorCollect()

	SetRoots(nil, nil, nil, nil)
	gc.MajorCollect()

	if gc.majorCount != 1 {
		t.Errorf("Expected 1 major collection, got %d", gc.majorCount)
	}
}

func TestGCModuleCollect(t *testing.T) {
	m := &GCModule{}
	result := m.Collect()
	if result != objects.None_ {
		t.Errorf("Expected None from GCModule.Collect, got %v", result)
	}
}

func TestGCModuleEnable(t *testing.T) {
	m := &GCModule{}
	result := m.Enable()
	if result != objects.None_ {
		t.Errorf("Expected None from GCModule.Enable, got %v", result)
	}
}

func TestGCModuleDisable(t *testing.T) {
	m := &GCModule{}
	result := m.Disable()
	if result != objects.None_ {
		t.Errorf("Expected None from GCModule.Disable, got %v", result)
	}
}

func TestGCModuleGetStats(t *testing.T) {
	m := &GCModule{}
	result := m.GetStats()
	if result == nil {
		t.Error("Expected non-nil result from GCModule.GetStats")
	}
	// Result should be a Dict
	if result.Type() != objects.DICT_OBJ {
		t.Errorf("Expected Dict from GCModule.GetStats, got %s", result.Type())
	}
}

func TestGCModuleSetThreshold(t *testing.T) {
	m := &GCModule{}
	result := m.SetThreshold(&objects.Integer{Value: 2048})
	if result != objects.None_ {
		t.Errorf("Expected None from GCModule.SetThreshold, got %v", result)
	}
}

func TestGCModuleSetVerbose(t *testing.T) {
	m := &GCModule{}
	result := m.SetVerbose(&objects.Boolean{Value: true})
	if result != objects.None_ {
		t.Errorf("Expected None from GCModule.SetVerbose, got %v", result)
	}
}

func TestGCEstimateSizeDefault(t *testing.T) {
	// Test the default case in estimateSize (unknown type)
	// We can use a type that's not explicitly handled
	obj := &objects.None{}
	size := estimateSize(obj)
	if size != 32 {
		t.Errorf("Expected default size 32, got %d", size)
	}
}

func TestGCRegisterRootsWithNil(t *testing.T) {
	gc := newTestGC()

	obj := &objects.Integer{Value: 1}
	gc.Allocate(obj)

	// Set register root with nil entry (should be skipped)
	SetRoots([]objects.Object{nil}, nil, nil, nil)

	gc.MinorCollect()

	// Object should be freed since nil root doesn't reference it
	if len(gc.youngObjects) != 0 {
		t.Errorf("Expected 0 young objects with nil register root, got %d", len(gc.youngObjects))
	}
}

func TestGCPackageLevelPrintStats(t *testing.T) {
	// Just verify it doesn't panic
	PrintStats()
}
