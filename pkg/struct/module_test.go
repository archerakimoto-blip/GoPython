package struct_

import (
	"testing"

	"github.com/go-py/go-python/pkg/objects"
)

func TestPackUnpack(t *testing.T) {
	// Test pack and unpack
	module := CreateStructModule()
	packFn, _ := module.Fields["pack"].(*objects.Builtin)
	unpackFn, _ := module.Fields["unpack"].(*objects.Builtin)

	// Test: pack('i', 42)
	result := packFn.Fn(&objects.String{Value: "i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}

	bytesObj, ok := result.(*objects.Bytes)
	if !ok {
		t.Fatalf("pack returned %T, not Bytes", result)
	}

	// Now unpack it
	unpacked := unpackFn.Fn(&objects.String{Value: "i"}, bytesObj)
	if unpacked.Type() == objects.ERROR_OBJ {
		t.Fatalf("unpack failed: %v", unpacked)
	}

	tuple, ok := unpacked.(*objects.Tuple)
	if !ok {
		t.Fatalf("unpack returned %T, not Tuple", unpacked)
	}

	if len(tuple.Elements) != 1 {
		t.Fatalf("unpack returned %d elements, expected 1", len(tuple.Elements))
	}

	ival, ok := tuple.Elements[0].(*objects.Integer)
	if !ok {
		t.Fatalf("unpacked element is %T, not Integer", tuple.Elements[0])
	}

	if ival.Value != 42 {
		t.Fatalf("unpacked %d, expected 42", ival.Value)
	}

	t.Logf("Test passed: pack('i', 42) and unpack('i', ...) works correctly")
}
