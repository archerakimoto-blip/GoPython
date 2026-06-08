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

// ==================== Additional Struct Tests ====================

func TestCreateStructModule(t *testing.T) {
	module := CreateStructModule()
	if module.Name != "struct" {
		t.Errorf("Expected module name 'struct', got '%s'", module.Name)
	}

	requiredFuncs := []string{"pack", "unpack"}
	for _, fn := range requiredFuncs {
		if _, ok := module.Fields[fn]; !ok {
			t.Errorf("Missing function: %s", fn)
		}
	}
}

func TestPackByteSigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "b"}, &objects.Integer{Value: -1})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "b"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -1 {
		t.Errorf("Expected -1, got %d", val)
	}
}

func TestUnpackBufferTooLarge(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	// Pack a single byte, then try to unpack with extra data
	result := unpackFn.Fn(&objects.String{Value: "b"}, &objects.Bytes{Value: []byte{42, 99}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for buffer too large, got %T", result)
	}
}

func TestUnpackBufferTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	// Try to unpack an int from empty buffer
	result := unpackFn.Fn(&objects.String{Value: "i"}, &objects.Bytes{Value: []byte{}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for buffer too small, got %T", result)
	}
}

func TestUnpackShortTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "h"}, &objects.Bytes{Value: []byte{1}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for short too small, got %T", result)
	}
}

func TestUnpackUnsignedShortTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "H"}, &objects.Bytes{Value: []byte{1}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for unsigned short too small, got %T", result)
	}
}

func TestUnpackIntTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "i"}, &objects.Bytes{Value: []byte{1, 2}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for int too small, got %T", result)
	}
}

func TestUnpackUnsignedIntTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "I"}, &objects.Bytes{Value: []byte{1, 2}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for unsigned int too small, got %T", result)
	}
}

func TestUnpackLongTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "l"}, &objects.Bytes{Value: []byte{1, 2}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for long too small, got %T", result)
	}
}

func TestUnpackUnsignedLongTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "L"}, &objects.Bytes{Value: []byte{1, 2}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for unsigned long too small, got %T", result)
	}
}

func TestUnpackLongLongTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "q"}, &objects.Bytes{Value: []byte{1, 2, 3, 4}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for long long too small, got %T", result)
	}
}

func TestUnpackUnsignedLongLongTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "Q"}, &objects.Bytes{Value: []byte{1, 2, 3, 4}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for unsigned long long too small, got %T", result)
	}
}

func TestUnpackFloatTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "f"}, &objects.Bytes{Value: []byte{1, 2}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for float too small, got %T", result)
	}
}

func TestUnpackDoubleTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "d"}, &objects.Bytes{Value: []byte{1, 2, 3, 4}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for double too small, got %T", result)
	}
}

func TestUnpackCharTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "c"}, &objects.Bytes{Value: []byte{}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for char too small, got %T", result)
	}
}

func TestUnpackPadByteTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "x"}, &objects.Bytes{Value: []byte{}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for pad byte too small, got %T", result)
	}
}

func TestPackUnpackUnsignedShortLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<H"}, &objects.Integer{Value: 60000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<H"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 60000 {
		t.Errorf("Expected 60000, got %d", val)
	}
}

func TestPackUnpackUnsignedShortBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">H"}, &objects.Integer{Value: 60000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">H"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 60000 {
		t.Errorf("Expected 60000, got %d", val)
	}
}

func TestPackUnpackUnsignedIntLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<I"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<I"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackUnsignedIntBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">I"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">I"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackPadByteLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<xb"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)
	if len(bytesObj.Value) != 2 {
		t.Errorf("Expected 2 bytes, got %d", len(bytesObj.Value))
	}

	unpacked := unpackFn.Fn(&objects.String{Value: "<xb"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackPadByteBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">xb"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">xb"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackCharFromIntBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">c"}, &objects.Integer{Value: 65})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "A" {
		t.Errorf("Expected 'A', got '%s'", val)
	}
}

func TestPackUnpackNetworkOrderUnsignedByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!B"}, &objects.Integer{Value: 200})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!B"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 200 {
		t.Errorf("Expected 200, got %d", val)
	}
}

func TestPackUnpackNativeOrderUnsignedByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@B"}, &objects.Integer{Value: 200})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@B"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 200 {
		t.Errorf("Expected 200, got %d", val)
	}
}

func TestPackUnpackNetworkOrderUnsignedShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!H"}, &objects.Integer{Value: 60000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!H"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 60000 {
		t.Errorf("Expected 60000, got %d", val)
	}
}

func TestPackUnpackNativeOrderUnsignedShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@H"}, &objects.Integer{Value: 60000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@H"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 60000 {
		t.Errorf("Expected 60000, got %d", val)
	}
}

func TestPackUnpackNetworkOrderUnsignedInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!I"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!I"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackNativeOrderUnsignedInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@I"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@I"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackNetworkOrderUnsignedLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!L"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!L"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackNativeOrderUnsignedLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@L"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@L"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackNetworkOrderUnsignedLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!Q"}, &objects.Integer{Value: 9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!Q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", val)
	}
}

func TestPackUnpackNativeOrderUnsignedLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@Q"}, &objects.Integer{Value: 9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@Q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", val)
	}
}

func TestPackUnpackNetworkOrderChar(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!c"}, &objects.String{Value: "A"})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "A" {
		t.Errorf("Expected 'A', got '%s'", val)
	}
}

func TestPackUnpackNativeOrderChar(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@c"}, &objects.String{Value: "A"})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "A" {
		t.Errorf("Expected 'A', got '%s'", val)
	}
}

func TestPackUnpackNetworkOrderLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!l"}, &objects.Integer{Value: -100000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!l"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -100000 {
		t.Errorf("Expected -100000, got %d", val)
	}
}

func TestPackUnpackNativeOrderLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@l"}, &objects.Integer{Value: -100000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@l"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -100000 {
		t.Errorf("Expected -100000, got %d", val)
	}
}

func TestPackUnpackNetworkOrderInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackNativeOrderInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackNetworkOrderPadByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!xb"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!xb"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackNativeOrderPadByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@xb"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@xb"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackNetworkOrderFloatFromInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!f"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackUnpackNativeOrderFloatFromInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@f"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackUnpackNetworkOrderDoubleFromInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!d"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackUnpackNativeOrderDoubleFromInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@d"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackWrongArgTypeForF(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "f"}, &objects.String{Value: "not float"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'f', got %T", result)
	}
}

func TestPackWrongArgTypeForD(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "d"}, &objects.String{Value: "not float"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'd', got %T", result)
	}
}

func TestPackWrongArgTypeForC(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "c"}, &objects.Integer{Value: 65})
	// c with int should work (uses byte), so this should succeed
	if result.Type() == objects.ERROR_OBJ {
		t.Errorf("Expected pack 'c' with int to succeed, got error: %v", result)
	}
}

func TestPackWrongArgTypeForCWithList(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "c"}, &objects.List{})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'c' with list, got %T", result)
	}
}

func TestPackWrongArgTypeForL(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "L"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'L', got %T", result)
	}
}

func TestPackWrongArgTypeForLUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "L"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'L', got %T", result)
	}
}

func TestPackNotEnoughArgsForL(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "L"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'L', got %T", result)
	}
}

func TestPackNotEnoughArgsForLUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "l"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'l', got %T", result)
	}
}

func TestPackInvalidFormatChar(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "z"}, &objects.Integer{Value: 1})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for invalid format char 'z', got %T", result)
	}
}

func TestPackUnpackSignedByteTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "b"}, &objects.Bytes{Value: []byte{}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for signed byte too small, got %T", result)
	}
}

func TestPackUnpackUnsignedByteTooSmall(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "B"}, &objects.Bytes{Value: []byte{}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for unsigned byte too small, got %T", result)
	}
}

func TestPackByteUnsigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "B"}, &objects.Integer{Value: 255})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "B"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 255 {
		t.Errorf("Expected 255, got %d", val)
	}
}

func TestPackShortSigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "h"}, &objects.Integer{Value: -1000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -1000 {
		t.Errorf("Expected -1000, got %d", val)
	}
}

func TestPackShortUnsigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "H"}, &objects.Integer{Value: 65535})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "H"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 65535 {
		t.Errorf("Expected 65535, got %d", val)
	}
}

func TestPackIntSigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "i"}, &objects.Integer{Value: -100000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -100000 {
		t.Errorf("Expected -100000, got %d", val)
	}
}

func TestPackIntUnsigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "I"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "I"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackLongSigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "l"}, &objects.Integer{Value: -100000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "l"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -100000 {
		t.Errorf("Expected -100000, got %d", val)
	}
}

func TestPackLongUnsigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "L"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "L"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackLongLongSigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "q"}, &objects.Integer{Value: -9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -9223372036854775807 {
		t.Errorf("Expected -9223372036854775807, got %d", val)
	}
}

func TestPackLongLongUnsigned(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "Q"}, &objects.Integer{Value: 9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "Q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", val)
	}
}

func TestPackFloat(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "f"}, &objects.Float{Value: 3.14})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	// Float32 precision
	if val < 3.13 || val > 3.15 {
		t.Errorf("Expected ~3.14, got %f", val)
	}
}

func TestPackDouble(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "d"}, &objects.Float{Value: 3.14159265358979})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.14159265358979 {
		t.Errorf("Expected 3.14159265358979, got %f", val)
	}
}

func TestPackChar(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "c"}, &objects.String{Value: "A"})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "A" {
		t.Errorf("Expected 'A', got '%s'", val)
	}
}

func TestPackCharFromInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "c"}, &objects.Integer{Value: 65})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "A" {
		t.Errorf("Expected 'A', got '%s'", val)
	}
}

func TestPackPadByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "bx"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)
	// 'b' = 1 byte + 'x' = 1 pad byte = 2 bytes total
	if len(bytesObj.Value) != 2 {
		t.Errorf("Expected 2 bytes, got %d", len(bytesObj.Value))
	}

	unpacked := unpackFn.Fn(&objects.String{Value: "bx"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	if len(tuple.Elements) != 1 {
		t.Errorf("Expected 1 element (pad byte is skipped), got %d", len(tuple.Elements))
	}
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	// Little-endian: least significant byte first
	if bytesObj.Value[0] != 42 {
		t.Errorf("Expected first byte to be 42 in little-endian, got %d", bytesObj.Value[0])
	}

	unpacked := unpackFn.Fn(&objects.String{Value: "<i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	// Big-endian: most significant byte first
	if bytesObj.Value[3] != 42 {
		t.Errorf("Expected last byte to be 42 in big-endian, got %d", bytesObj.Value[3])
	}

	unpacked := unpackFn.Fn(&objects.String{Value: ">i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackNetworkOrder(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackNativeOrder(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackMultipleValues(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "ih"}, &objects.Integer{Value: 42}, &objects.Integer{Value: 1000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "ih"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(tuple.Elements))
	}

	val0 := tuple.Elements[0].(*objects.Integer).Value
	val1 := tuple.Elements[1].(*objects.Integer).Value
	if val0 != 42 {
		t.Errorf("Expected 42, got %d", val0)
	}
	if val1 != 1000 {
		t.Errorf("Expected 1000, got %d", val1)
	}
}

func TestPackNoArgs(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestPackWrongFormatType(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string format, got %T", result)
	}
}

func TestPackNotEnoughArgs(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "ii"}, &objects.Integer{Value: 42})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args, got %T", result)
	}
}

func TestPackWrongArgType(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "i"}, &objects.String{Value: "not an int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong arg type, got %T", result)
	}
}

func TestUnpackNoArgs(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn()
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for no args, got %T", result)
	}
}

func TestUnpackWrongFormatType(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.Integer{Value: 42}, &objects.Bytes{Value: []byte{0}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-string format, got %T", result)
	}
}

func TestUnpackWrongDataType(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "i"}, &objects.String{Value: "not bytes"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for non-bytes data, got %T", result)
	}
}

func TestUnpackInvalidFormatChar(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "z"}, &objects.Bytes{Value: []byte{0}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for invalid format char, got %T", result)
	}
}

func TestUnpackNotEnoughBytes(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	// 'i' requires 4 bytes, but we only provide 2
	result := unpackFn.Fn(&objects.String{Value: "i"}, &objects.Bytes{Value: []byte{0, 0}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough bytes, got %T", result)
	}
}

func TestUnpackTooManyBytes(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	// 'b' requires 1 byte, but we provide 2
	result := unpackFn.Fn(&objects.String{Value: "b"}, &objects.Bytes{Value: []byte{0, 0}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for too many bytes, got %T", result)
	}
}

func TestPackFloatFromInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "f"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackDoubleFromInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "d"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackFloatWrongType(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "f"}, &objects.String{Value: "not a float"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type, got %T", result)
	}
}

func TestPackDoubleWrongType(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "d"}, &objects.String{Value: "not a double"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type, got %T", result)
	}
}

func TestPackCharWrongType(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "c"}, &objects.Float{Value: 3.14})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type, got %T", result)
	}
}

func TestPackLittleEndianShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<h"}, &objects.Integer{Value: 1234})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 1234 {
		t.Errorf("Expected 1234, got %d", val)
	}
}

func TestPackBigEndianShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">h"}, &objects.Integer{Value: 1234})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 1234 {
		t.Errorf("Expected 1234, got %d", val)
	}
}

func TestPackLittleEndianLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<q"}, &objects.Integer{Value: 1234567890123})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 1234567890123 {
		t.Errorf("Expected 1234567890123, got %d", val)
	}
}

func TestPackMixedFormat(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<ib"}, &objects.Integer{Value: 42}, &objects.Integer{Value: 7})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<ib"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(tuple.Elements))
	}

	val0 := tuple.Elements[0].(*objects.Integer).Value
	val1 := tuple.Elements[1].(*objects.Integer).Value
	if val0 != 42 {
		t.Errorf("Expected 42, got %d", val0)
	}
	if val1 != 7 {
		t.Errorf("Expected 7, got %d", val1)
	}
}

func TestPackUnpackZero(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "i"}, &objects.Integer{Value: 0})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 0 {
		t.Errorf("Expected 0, got %d", val)
	}
}

func TestPackUnpackNegativeInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "i"}, &objects.Integer{Value: -42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -42 {
		t.Errorf("Expected -42, got %d", val)
	}
}

func TestPackUnpackDoublePrecision(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	pi := 3.14159265358979323846
	result := packFn.Fn(&objects.String{Value: "d"}, &objects.Float{Value: pi})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != pi {
		t.Errorf("Expected %f, got %f", pi, val)
	}
}

func TestPackUnpackWithPadBytes(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	// Pack with pad bytes
	result := packFn.Fn(&objects.String{Value: "bxb"}, &objects.Integer{Value: 1}, &objects.Integer{Value: 2})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)
	if len(bytesObj.Value) != 3 {
		t.Errorf("Expected 3 bytes (1 + 1 pad + 1), got %d", len(bytesObj.Value))
	}

	unpacked := unpackFn.Fn(&objects.String{Value: "bxb"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(tuple.Elements))
	}

	val0 := tuple.Elements[0].(*objects.Integer).Value
	val1 := tuple.Elements[1].(*objects.Integer).Value
	if val0 != 1 {
		t.Errorf("Expected 1, got %d", val0)
	}
	if val1 != 2 {
		t.Errorf("Expected 2, got %d", val1)
	}
}

func TestPackEndianAffectsBytes(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	// Pack the same value with different endianness
	leResult := packFn.Fn(&objects.String{Value: "<h"}, &objects.Integer{Value: 1})
	beResult := packFn.Fn(&objects.String{Value: ">h"}, &objects.Integer{Value: 1})

	leBytes := leResult.(*objects.Bytes).Value
	beBytes := beResult.(*objects.Bytes).Value

	// Little-endian: LSB first (01 00)
	if leBytes[0] != 1 || leBytes[1] != 0 {
		t.Errorf("Little-endian bytes wrong: got %v", leBytes)
	}

	// Big-endian: MSB first (00 01)
	if beBytes[0] != 0 || beBytes[1] != 1 {
		t.Errorf("Big-endian bytes wrong: got %v", beBytes)
	}
}

func TestPackUnpackAllIntegerFormats(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	formats := []struct {
		fmt  string
		val  int64
	}{
		{"b", -128},
		{"B", 255},
		{"h", -32768},
		{"H", 65535},
		{"i", -2147483648},
		{"I", 4294967295},
		{"l", -2147483648},
		{"L", 4294967295},
		{"q", -9223372036854775807},
		{"Q", 9223372036854775807},
	}

	for _, tt := range formats {
		t.Run(tt.fmt, func(t *testing.T) {
			result := packFn.Fn(&objects.String{Value: tt.fmt}, &objects.Integer{Value: tt.val})
			if result.Type() == objects.ERROR_OBJ {
				t.Fatalf("pack failed: %v", result)
			}
			bytesObj := result.(*objects.Bytes)

			unpacked := unpackFn.Fn(&objects.String{Value: tt.fmt}, bytesObj)
			if unpacked.Type() == objects.ERROR_OBJ {
				t.Fatalf("unpack failed: %v", unpacked)
			}
			tuple := unpacked.(*objects.Tuple)
			val := tuple.Elements[0].(*objects.Integer).Value
			if val != tt.val {
				t.Errorf("Format %s: expected %d, got %d", tt.fmt, tt.val, val)
			}
		})
	}
}

func TestPackUnpackFloatFormats(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	t.Run("float", func(t *testing.T) {
		result := packFn.Fn(&objects.String{Value: "f"}, &objects.Float{Value: 1.5})
		if result.Type() == objects.ERROR_OBJ {
			t.Fatalf("pack failed: %v", result)
		}
		bytesObj := result.(*objects.Bytes)

		unpacked := unpackFn.Fn(&objects.String{Value: "f"}, bytesObj)
		tuple := unpacked.(*objects.Tuple)
		val := tuple.Elements[0].(*objects.Float).Value
		if val != 1.5 {
			t.Errorf("Expected 1.5, got %f", val)
		}
	})

	t.Run("double", func(t *testing.T) {
		result := packFn.Fn(&objects.String{Value: "d"}, &objects.Float{Value: 1.5})
		if result.Type() == objects.ERROR_OBJ {
			t.Fatalf("pack failed: %v", result)
		}
		bytesObj := result.(*objects.Bytes)

		unpacked := unpackFn.Fn(&objects.String{Value: "d"}, bytesObj)
		tuple := unpacked.(*objects.Tuple)
		val := tuple.Elements[0].(*objects.Float).Value
		if val != 1.5 {
			t.Errorf("Expected 1.5, got %f", val)
		}
	})
}

// ==================== Additional Struct Tests for 90%+ Coverage ====================

func TestPackCharEmptyString(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	// Empty string should not panic (uses byte 0)
	result := packFn.Fn(&objects.String{Value: "c"}, &objects.String{Value: ""})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack with empty string failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)
	if len(bytesObj.Value) != 1 {
		t.Errorf("Expected 1 byte, got %d", len(bytesObj.Value))
	}
}

func TestPackUnpackLittleEndianInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackBigEndianInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">i"}, &objects.Integer{Value: 42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">i"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestPackUnpackLittleEndianShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<h"}, &objects.Integer{Value: -1000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -1000 {
		t.Errorf("Expected -1000, got %d", val)
	}
}

func TestPackUnpackBigEndianShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">h"}, &objects.Integer{Value: -1000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -1000 {
		t.Errorf("Expected -1000, got %d", val)
	}
}

func TestPackUnpackLittleEndianLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<q"}, &objects.Integer{Value: -9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -9223372036854775807 {
		t.Errorf("Expected -9223372036854775807, got %d", val)
	}
}

func TestPackUnpackBigEndianLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">q"}, &objects.Integer{Value: 9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", val)
	}
}

func TestPackUnpackLittleEndianFloat(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<f"}, &objects.Float{Value: 3.14})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val < 3.13 || val > 3.15 {
		t.Errorf("Expected ~3.14, got %f", val)
	}
}

func TestPackUnpackBigEndianFloat(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">f"}, &objects.Float{Value: 3.14})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val < 3.13 || val > 3.15 {
		t.Errorf("Expected ~3.14, got %f", val)
	}
}

func TestPackUnpackLittleEndianDouble(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<d"}, &objects.Float{Value: 3.14159265358979})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.14159265358979 {
		t.Errorf("Expected 3.14159265358979, got %f", val)
	}
}

func TestPackUnpackBigEndianDouble(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">d"}, &objects.Float{Value: 3.14159265358979})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.14159265358979 {
		t.Errorf("Expected 3.14159265358979, got %f", val)
	}
}

func TestPackUnpackCharLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<c"}, &objects.String{Value: "A"})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "A" {
		t.Errorf("Expected 'A', got '%s'", val)
	}
}

func TestPackUnpackUnsignedByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "B"}, &objects.Integer{Value: 200})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "B"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 200 {
		t.Errorf("Expected 200, got %d", val)
	}
}

func TestPackUnpackUnsignedShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "H"}, &objects.Integer{Value: 60000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "H"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 60000 {
		t.Errorf("Expected 60000, got %d", val)
	}
}

func TestPackUnpackUnsignedInt(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "I"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "I"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackUnsignedLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "L"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "L"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackUnsignedLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "Q"}, &objects.Integer{Value: 9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "Q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", val)
	}
}

func TestPackNotEnoughArgsForB(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "b"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'b', got %T", result)
	}
}

func TestPackNotEnoughArgsForBUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "B"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'B', got %T", result)
	}
}

func TestPackNotEnoughArgsForH(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "h"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'h', got %T", result)
	}
}

func TestPackNotEnoughArgsForHUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "H"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'H', got %T", result)
	}
}

func TestPackNotEnoughArgsForI(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "i"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'i', got %T", result)
	}
}

func TestPackNotEnoughArgsForIUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "I"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'I', got %T", result)
	}
}

func TestPackNotEnoughArgsForQ(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "q"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'q', got %T", result)
	}
}

func TestPackNotEnoughArgsForQUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "Q"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'Q', got %T", result)
	}
}

func TestPackNotEnoughArgsForF(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "f"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'f', got %T", result)
	}
}

func TestPackNotEnoughArgsForD(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "d"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'd', got %T", result)
	}
}

func TestPackNotEnoughArgsForC(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "c"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for not enough args for 'c', got %T", result)
	}
}

func TestPackWrongArgTypeForB(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "b"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'b', got %T", result)
	}
}

func TestPackWrongArgTypeForBUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "B"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'B', got %T", result)
	}
}

func TestPackWrongArgTypeForH(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "h"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'h', got %T", result)
	}
}

func TestPackWrongArgTypeForHUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "H"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'H', got %T", result)
	}
}

func TestPackWrongArgTypeForI(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "i"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'i', got %T", result)
	}
}

func TestPackWrongArgTypeForIUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "I"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'I', got %T", result)
	}
}

func TestPackWrongArgTypeForQ(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "q"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'q', got %T", result)
	}
}

func TestPackWrongArgTypeForQUpper(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "Q"}, &objects.String{Value: "not int"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for wrong type for 'Q', got %T", result)
	}
}

func TestUnpackInvalidFormatCharDup(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "z"}, &objects.Bytes{Value: []byte{0}})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for invalid format char in unpack, got %T", result)
	}
}

func TestPackUnpackFloatFromIntLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<f"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackUnpackDoubleFromIntLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<d"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackOnlyByteOrder(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)

	// Just byte order prefix, no actual format
	result := packFn.Fn(&objects.String{Value: "<"})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack with only byte order failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)
	if len(bytesObj.Value) != 0 {
		t.Errorf("Expected 0 bytes for just byte order, got %d", len(bytesObj.Value))
	}
}

func TestUnpackOnlyByteOrder(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	// Just byte order prefix, no actual format
	result := unpackFn.Fn(&objects.String{Value: "<"}, &objects.Bytes{Value: []byte{}})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("unpack with only byte order failed: %v", result)
	}
	tuple := result.(*objects.Tuple)
	if len(tuple.Elements) != 0 {
		t.Errorf("Expected 0 elements for just byte order, got %d", len(tuple.Elements))
	}
}

func TestPackUnpackSignedByteLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<b"}, &objects.Integer{Value: -42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<b"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -42 {
		t.Errorf("Expected -42, got %d", val)
	}
}

func TestPackUnpackUnsignedByteLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<B"}, &objects.Integer{Value: 200})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<B"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 200 {
		t.Errorf("Expected 200, got %d", val)
	}
}

func TestPackUnpackLongLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<l"}, &objects.Integer{Value: -100000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<l"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -100000 {
		t.Errorf("Expected -100000, got %d", val)
	}
}

func TestPackUnpackUnsignedLongLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<L"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<L"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackUnsignedLongLongLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<Q"}, &objects.Integer{Value: 9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<Q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", val)
	}
}

func TestPackUnpackCharBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">c"}, &objects.String{Value: "Z"})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "Z" {
		t.Errorf("Expected 'Z', got '%s'", val)
	}
}

func TestPackUnpackCharFromIntLittleEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "<c"}, &objects.Integer{Value: 65})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "<c"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.String).Value
	if val != "A" {
		t.Errorf("Expected 'A', got '%s'", val)
	}
}

func TestPackUnpackMixedEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	// Pack with big endian, unpack with big endian
	result := packFn.Fn(&objects.String{Value: ">ih"}, &objects.Integer{Value: 42}, &objects.Integer{Value: 1000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">ih"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(tuple.Elements))
	}

	val0 := tuple.Elements[0].(*objects.Integer).Value
	val1 := tuple.Elements[1].(*objects.Integer).Value
	if val0 != 42 {
		t.Errorf("Expected 42, got %d", val0)
	}
	if val1 != 1000 {
		t.Errorf("Expected 1000, got %d", val1)
	}
}

func TestPackUnpackLongBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">l"}, &objects.Integer{Value: -100000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">l"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -100000 {
		t.Errorf("Expected -100000, got %d", val)
	}
}

func TestPackUnpackUnsignedLongBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">L"}, &objects.Integer{Value: 3000000000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">L"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 3000000000 {
		t.Errorf("Expected 3000000000, got %d", val)
	}
}

func TestPackUnpackUnsignedLongLongBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">Q"}, &objects.Integer{Value: 9223372036854775807})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">Q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", val)
	}
}

func TestPackUnpackSignedByteBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">b"}, &objects.Integer{Value: -42})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">b"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -42 {
		t.Errorf("Expected -42, got %d", val)
	}
}

func TestPackUnpackUnsignedByteBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">B"}, &objects.Integer{Value: 200})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">B"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 200 {
		t.Errorf("Expected 200, got %d", val)
	}
}

func TestPackUnpackSignedShortBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">h"}, &objects.Integer{Value: -1000})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -1000 {
		t.Errorf("Expected -1000, got %d", val)
	}
}

func TestPackUnpackFloatBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">f"}, &objects.Float{Value: 3.14})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val < 3.13 || val > 3.15 {
		t.Errorf("Expected ~3.14, got %f", val)
	}
}

func TestPackUnpackDoubleBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">d"}, &objects.Float{Value: 3.14159265358979})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.14159265358979 {
		t.Errorf("Expected 3.14159265358979, got %f", val)
	}
}

func TestPackUnpackFloatFromIntBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">f"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackUnpackDoubleFromIntBigEndian(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: ">d"}, &objects.Integer{Value: 3})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: ">d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 3.0 {
		t.Errorf("Expected 3.0, got %f", val)
	}
}

func TestPackUnpackNetworkOrderFloat(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!f"}, &objects.Float{Value: 1.5})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 1.5 {
		t.Errorf("Expected 1.5, got %f", val)
	}
}

func TestPackUnpackNetworkOrderDouble(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!d"}, &objects.Float{Value: 1.5})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 1.5 {
		t.Errorf("Expected 1.5, got %f", val)
	}
}

func TestPackUnpackNativeOrderFloat(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@f"}, &objects.Float{Value: 1.5})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@f"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 1.5 {
		t.Errorf("Expected 1.5, got %f", val)
	}
}

func TestPackUnpackNativeOrderDouble(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@d"}, &objects.Float{Value: 1.5})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@d"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Float).Value
	if val != 1.5 {
		t.Errorf("Expected 1.5, got %f", val)
	}
}

func TestPackUnpackNetworkOrderShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!h"}, &objects.Integer{Value: 1234})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 1234 {
		t.Errorf("Expected 1234, got %d", val)
	}
}

func TestPackUnpackNativeOrderShort(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@h"}, &objects.Integer{Value: 1234})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@h"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 1234 {
		t.Errorf("Expected 1234, got %d", val)
	}
}

func TestPackUnpackNetworkOrderLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!q"}, &objects.Integer{Value: 1234567890123})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 1234567890123 {
		t.Errorf("Expected 1234567890123, got %d", val)
	}
}

func TestPackUnpackNativeOrderLongLong(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@q"}, &objects.Integer{Value: 1234567890123})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@q"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != 1234567890123 {
		t.Errorf("Expected 1234567890123, got %d", val)
	}
}

func TestUnpackOneArg(t *testing.T) {
	module := CreateStructModule()
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := unpackFn.Fn(&objects.String{Value: "i"})
	if result.Type() != objects.ERROR_OBJ {
		t.Errorf("Expected error for only 1 arg, got %T", result)
	}
}

func TestPackUnpackNetworkOrderByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "!b"}, &objects.Integer{Value: -1})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "!b"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -1 {
		t.Errorf("Expected -1, got %d", val)
	}
}

func TestPackUnpackNativeOrderByte(t *testing.T) {
	module := CreateStructModule()
	packFn := module.Fields["pack"].(*objects.Builtin)
	unpackFn := module.Fields["unpack"].(*objects.Builtin)

	result := packFn.Fn(&objects.String{Value: "@b"}, &objects.Integer{Value: -1})
	if result.Type() == objects.ERROR_OBJ {
		t.Fatalf("pack failed: %v", result)
	}
	bytesObj := result.(*objects.Bytes)

	unpacked := unpackFn.Fn(&objects.String{Value: "@b"}, bytesObj)
	tuple := unpacked.(*objects.Tuple)
	val := tuple.Elements[0].(*objects.Integer).Value
	if val != -1 {
		t.Errorf("Expected -1, got %d", val)
	}
}
