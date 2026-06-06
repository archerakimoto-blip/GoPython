package struct_

import (
	"bytes"
	"encoding/binary"

	"github.com/go-py/go-python/pkg/objects"
)

// pack packs values according to the format string and returns a bytes object
func pack(args ...objects.Object) objects.Object {
	if len(args) < 1 {
		return objects.NewTypeError("pack() takes at least 1 argument")
	}

	formatObj, ok := args[0].(*objects.String)
	if !ok {
		return objects.NewTypeError("pack() first argument must be a string")
	}
	format := formatObj.Value

	buf := new(bytes.Buffer)
	argIndex := 1
	var byteOrder binary.ByteOrder = binary.BigEndian // Default

	for i := 0; i < len(format); i++ {
		c := format[i]
		switch c {
		case '<':
			byteOrder = binary.LittleEndian
		case '>':
			byteOrder = binary.BigEndian
		case '!':
			byteOrder = binary.BigEndian
		case '@':
			byteOrder = binary.BigEndian
		case 'c':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			var b byte
			switch v := args[argIndex].(type) {
			case *objects.String:
				if len(v.Value) > 0 {
					b = v.Value[0]
				}
			case *objects.Integer:
				b = byte(v.Value)
			default:
				return objects.NewTypeError("pack() argument for 'c' format must be string or int")
			}
			if err := binary.Write(buf, byteOrder, b); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'b':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'b' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, int8(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'B':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'B' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, uint8(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'h':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'h' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, int16(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'H':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'H' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, uint16(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'i', 'l':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'i'/'l' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, int32(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'I', 'L':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'I'/'L' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, uint32(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'q':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'q' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, int64(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'Q':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			val, ok := args[argIndex].(*objects.Integer)
			if !ok {
				return objects.NewTypeError("pack() argument for 'Q' format must be integer")
			}
			if err := binary.Write(buf, byteOrder, uint64(val.Value)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'f':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			var val float64
			switch v := args[argIndex].(type) {
			case *objects.Float:
				val = v.Value
			case *objects.Integer:
				val = float64(v.Value)
			default:
				return objects.NewTypeError("pack() argument for 'f' format must be float or int")
			}
			if err := binary.Write(buf, byteOrder, float32(val)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'd':
			if argIndex >= len(args) {
				return objects.NewValueError("not enough arguments for format string")
			}
			var val float64
			switch v := args[argIndex].(type) {
			case *objects.Float:
				val = v.Value
			case *objects.Integer:
				val = float64(v.Value)
			default:
				return objects.NewTypeError("pack() argument for 'd' format must be float or int")
			}
			if err := binary.Write(buf, byteOrder, val); err != nil {
				return objects.NewError("pack error: %v", err)
			}
			argIndex++
		case 'x':
			if err := binary.Write(buf, byteOrder, byte(0)); err != nil {
				return objects.NewError("pack error: %v", err)
			}
		default:
			return objects.NewValueError("invalid format character '%c'", c)
		}
	}

	return &objects.Bytes{Value: buf.Bytes()}
}

// unpack unpacks bytes according to the format string and returns a tuple of values
func unpack(args ...objects.Object) objects.Object {
	if len(args) < 2 {
		return objects.NewTypeError("unpack() takes exactly 2 arguments")
	}

	formatObj, ok := args[0].(*objects.String)
	if !ok {
		return objects.NewTypeError("unpack() first argument must be a string")
	}
	format := formatObj.Value

	dataObj, ok := args[1].(*objects.Bytes)
	if !ok {
		return objects.NewTypeError("unpack() second argument must be bytes")
	}
	data := dataObj.Value
	buf := bytes.NewReader(data)

	var result []objects.Object
	var byteOrder binary.ByteOrder = binary.BigEndian // Default

	for i := 0; i < len(format); i++ {
		c := format[i]
		switch c {
		case '<':
			byteOrder = binary.LittleEndian
		case '>':
			byteOrder = binary.BigEndian
		case '!':
			byteOrder = binary.BigEndian
		case '@':
			byteOrder = binary.BigEndian
		case 'c':
			var b byte
			if err := binary.Read(buf, byteOrder, &b); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.String{Value: string([]byte{b})})
		case 'b':
			var v int8
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: int64(v)})
		case 'B':
			var v uint8
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: int64(v)})
		case 'h':
			var v int16
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: int64(v)})
		case 'H':
			var v uint16
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: int64(v)})
		case 'i', 'l':
			var v int32
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: int64(v)})
		case 'I', 'L':
			var v uint32
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: int64(v)})
		case 'q':
			var v int64
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: v})
		case 'Q':
			var v uint64
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Integer{Value: int64(v)})
		case 'f':
			var v float32
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Float{Value: float64(v)})
		case 'd':
			var v float64
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
			result = append(result, &objects.Float{Value: v})
		case 'x':
			var v byte
			if err := binary.Read(buf, byteOrder, &v); err != nil {
				return objects.NewValueError("unpack requires a buffer of enough bytes")
			}
		default:
			return objects.NewValueError("invalid format character '%c'", c)
		}
	}

	if buf.Len() > 0 {
		return objects.NewValueError("unpack requires a buffer of exact size")
	}

	return &objects.Tuple{Elements: result}
}

// CreateStructModule creates and returns the struct module
func CreateStructModule() *objects.Module {
	module := &objects.Module{
		Name:   "struct",
		Fields: make(map[string]objects.Object),
	}

	module.Fields["pack"] = &objects.Builtin{
		Name: "struct.pack",
		Fn:   pack,
	}

	module.Fields["unpack"] = &objects.Builtin{
		Name: "struct.unpack",
		Fn:   unpack,
	}

	return module
}
