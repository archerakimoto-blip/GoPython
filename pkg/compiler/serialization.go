package compiler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-py/go-python/pkg/objects"
)

// Serialization format for register bytecode:
//
//	Magic:    4 bytes "GPYC"
//	Version:  uint16 (currently 1)
//	NumRegs:  uint16
//	NumConst: uint16 (number of constants)
//	Constants: [serializeConstant] * NumConst
//	NumInstr: uint16 (number of instructions)
//	Instructions: [serializeInstruction] * NumInstr

const regBytecodeMagic = "GPYC"
const regBytecodeVersion uint16 = 1

// Constant type tags for serialization
const (
	constTagInteger  byte = 1
	constTagFloat    byte = 2
	constTagString   byte = 3
	constTagBoolean  byte = 4
	constTagNull     byte = 5
	constTagFunction byte = 6
)

// SerializeRegBytecode serializes a RegBytecode into a binary format.
func SerializeRegBytecode(rb *RegBytecode) ([]byte, error) {
	var buf bytes.Buffer

	// Magic number
	if _, err := buf.WriteString(regBytecodeMagic); err != nil {
		return nil, fmt.Errorf("serialize magic: %w", err)
	}

	// Version
	if err := binary.Write(&buf, binary.BigEndian, regBytecodeVersion); err != nil {
		return nil, fmt.Errorf("serialize version: %w", err)
	}

	// NumRegs
	if err := binary.Write(&buf, binary.BigEndian, uint16(rb.NumRegs)); err != nil {
		return nil, fmt.Errorf("serialize numRegs: %w", err)
	}

	// Number of constants
	if err := binary.Write(&buf, binary.BigEndian, uint16(len(rb.Constants))); err != nil {
		return nil, fmt.Errorf("serialize numConstants: %w", err)
	}

	// Serialize each constant
	for i, c := range rb.Constants {
		if err := serializeConstant(&buf, c); err != nil {
			return nil, fmt.Errorf("serialize constant[%d]: %w", i, err)
		}
	}

	// Number of instructions
	if err := binary.Write(&buf, binary.BigEndian, uint16(len(rb.Instructions))); err != nil {
		return nil, fmt.Errorf("serialize numInstructions: %w", err)
	}

	// Serialize each instruction
	for i, instr := range rb.Instructions {
		if err := serializeInstruction(&buf, instr); err != nil {
			return nil, fmt.Errorf("serialize instruction[%d]: %w", i, err)
		}
	}

	return buf.Bytes(), nil
}

// DeserializeRegBytecode deserializes binary data back into a RegBytecode.
func DeserializeRegBytecode(data []byte) (*RegBytecode, error) {
	buf := bytes.NewReader(data)

	// Read and verify magic
	magic := make([]byte, len(regBytecodeMagic))
	if _, err := io.ReadFull(buf, magic); err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}
	if string(magic) != regBytecodeMagic {
		return nil, fmt.Errorf("invalid magic: expected %q, got %q", regBytecodeMagic, string(magic))
	}

	// Read version
	var version uint16
	if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	if version != regBytecodeVersion {
		return nil, fmt.Errorf("unsupported version: %d (expected %d)", version, regBytecodeVersion)
	}

	// Read NumRegs
	var numRegs uint16
	if err := binary.Read(buf, binary.BigEndian, &numRegs); err != nil {
		return nil, fmt.Errorf("read numRegs: %w", err)
	}

	// Read number of constants
	var numConstants uint16
	if err := binary.Read(buf, binary.BigEndian, &numConstants); err != nil {
		return nil, fmt.Errorf("read numConstants: %w", err)
	}

	// Read constants
	constants := make([]objects.Object, numConstants)
	for i := 0; i < int(numConstants); i++ {
		c, err := deserializeConstant(buf)
		if err != nil {
			return nil, fmt.Errorf("deserialize constant[%d]: %w", i, err)
		}
		constants[i] = c
	}

	// Read number of instructions
	var numInstructions uint16
	if err := binary.Read(buf, binary.BigEndian, &numInstructions); err != nil {
		return nil, fmt.Errorf("read numInstructions: %w", err)
	}

	// Read instructions
	instructions := make([]RegInstruction, numInstructions)
	for i := 0; i < int(numInstructions); i++ {
		instr, err := deserializeInstruction(buf)
		if err != nil {
			return nil, fmt.Errorf("deserialize instruction[%d]: %w", i, err)
		}
		instructions[i] = instr
	}

	return &RegBytecode{
		Constants:    constants,
		Instructions: instructions,
		NumRegs:      int(numRegs),
	}, nil
}

// serializeConstant writes a single constant to the buffer.
func serializeConstant(buf *bytes.Buffer, c objects.Object) error {
	switch v := c.(type) {
	case *objects.Integer:
		buf.WriteByte(constTagInteger)
		return binary.Write(buf, binary.BigEndian, v.Value)

	case *objects.Float:
		buf.WriteByte(constTagFloat)
		return binary.Write(buf, binary.BigEndian, v.Value)

	case *objects.String:
		buf.WriteByte(constTagString)
		// Write string length then bytes
		if err := binary.Write(buf, binary.BigEndian, uint32(len(v.Value))); err != nil {
			return err
		}
		_, err := buf.WriteString(v.Value)
		return err

	case *objects.Boolean:
		buf.WriteByte(constTagBoolean)
		if v.Value {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		return nil

	case *objects.None:
		buf.WriteByte(constTagNull)
		return nil

	case *CompiledFunction:
		buf.WriteByte(constTagFunction)
		// Serialize the function's instructions as bytes
		if err := binary.Write(buf, binary.BigEndian, uint32(len(v.Instructions))); err != nil {
			return err
		}
		if _, err := buf.Write(v.Instructions); err != nil {
			return err
		}
		if err := binary.Write(buf, binary.BigEndian, uint16(v.NumLocals)); err != nil {
			return err
		}
		if err := binary.Write(buf, binary.BigEndian, uint16(v.NumParameters)); err != nil {
			return err
		}
		if v.IsGenerator {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		if v.IsAsync {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		return nil

	default:
		// For unsupported types (Builtin, Module, Class, etc.),
		// serialize as Null since they are typically registered at startup
		buf.WriteByte(constTagNull)
		return nil
	}
}

// deserializeConstant reads a single constant from the buffer.
func deserializeConstant(buf *bytes.Reader) (objects.Object, error) {
	tag, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read constant tag: %w", err)
	}

	switch tag {
	case constTagInteger:
		var val int64
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, fmt.Errorf("read integer: %w", err)
		}
		return &objects.Integer{Value: val}, nil

	case constTagFloat:
		var val float64
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return nil, fmt.Errorf("read float: %w", err)
		}
		return &objects.Float{Value: val}, nil

	case constTagString:
		var length uint32
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return nil, fmt.Errorf("read string length: %w", err)
		}
		data := make([]byte, length)
		if _, err := io.ReadFull(buf, data); err != nil {
			return nil, fmt.Errorf("read string data: %w", err)
		}
		return &objects.String{Value: string(data)}, nil

	case constTagBoolean:
		b, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read boolean: %w", err)
		}
		return &objects.Boolean{Value: b == 1}, nil

	case constTagNull:
		return objects.None_, nil

	case constTagFunction:
		var instrLen uint32
		if err := binary.Read(buf, binary.BigEndian, &instrLen); err != nil {
			return nil, fmt.Errorf("read function instruction length: %w", err)
		}
		instrData := make([]byte, instrLen)
		if _, err := io.ReadFull(buf, instrData); err != nil {
			return nil, fmt.Errorf("read function instructions: %w", err)
		}
		var numLocals uint16
		if err := binary.Read(buf, binary.BigEndian, &numLocals); err != nil {
			return nil, fmt.Errorf("read function numLocals: %w", err)
		}
		var numParams uint16
		if err := binary.Read(buf, binary.BigEndian, &numParams); err != nil {
			return nil, fmt.Errorf("read function numParams: %w", err)
		}
		isGen, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read function isGenerator: %w", err)
		}
		isAsync, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read function isAsync: %w", err)
		}
		return &CompiledFunction{
			Instructions:  instrData,
			NumLocals:     int(numLocals),
			NumParameters: int(numParams),
			IsGenerator:   isGen == 1,
			IsAsync:       isAsync == 1,
		}, nil

	default:
		return nil, fmt.Errorf("unknown constant tag: %d", tag)
	}
}

// serializeInstruction writes a single register instruction to the buffer.
func serializeInstruction(buf *bytes.Buffer, instr RegInstruction) error {
	buf.WriteByte(byte(instr.Opcode))
	if err := binary.Write(buf, binary.BigEndian, uint8(len(instr.Operands))); err != nil {
		return err
	}
	for _, op := range instr.Operands {
		if err := binary.Write(buf, binary.BigEndian, int32(op)); err != nil {
			return err
		}
	}
	return nil
}

// deserializeInstruction reads a single register instruction from the buffer.
func deserializeInstruction(buf *bytes.Reader) (RegInstruction, error) {
	opcode, err := buf.ReadByte()
	if err != nil {
		return RegInstruction{}, fmt.Errorf("read opcode: %w", err)
	}

	var numOperands uint8
	if err := binary.Read(buf, binary.BigEndian, &numOperands); err != nil {
		return RegInstruction{}, fmt.Errorf("read numOperands: %w", err)
	}

	operands := make([]int, numOperands)
	for i := 0; i < int(numOperands); i++ {
		var val int32
		if err := binary.Read(buf, binary.BigEndian, &val); err != nil {
			return RegInstruction{}, fmt.Errorf("read operand[%d]: %w", i, err)
		}
		operands[i] = int(val)
	}

	return RegInstruction{
		Opcode:   RegOpcode(opcode),
		Operands: operands,
	}, nil
}
