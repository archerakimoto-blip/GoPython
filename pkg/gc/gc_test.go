package gc

import (
	"testing"

	"github.com/go-py/go-python/pkg/objects"
)

func TestGenerationalGCAllocate(t *testing.T) {
	gc := &GarbageCollector{
		youngObjects:    make([]*GCObject, 0),
		oldObjects:      make([]*GCObject, 0),
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
