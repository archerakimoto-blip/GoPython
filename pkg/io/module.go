package io

import (
	"bytes"
	"strings"

	"github.com/go-py/go-python/pkg/objects"
)

// StringIO is a wrapper that implements Object interface
type StringIO struct {
	Buffer strings.Builder
	Pos    int
}

func NewStringIO() *StringIO {
	return &StringIO{
		Buffer: strings.Builder{},
		Pos:    0,
	}
}

func (s *StringIO) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (s *StringIO) Inspect() string           { return s.Buffer.String() }

func (s *StringIO) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "read":
		return &objects.Builtin{
			Name: "StringIO.read",
			Fn: func(args ...objects.Object) objects.Object {
				size := -1
				if len(args) >= 1 {
					if i, ok := args[0].(*objects.Integer); ok {
						size = int(i.Value)
					}
				}
				if size == 0 {
					return &objects.String{Value: ""}
				}
				content := s.Buffer.String()
				if size < 0 {
					result := content[s.Pos:]
					s.Pos = len(content)
					return &objects.String{Value: result}
				}
				end := s.Pos + size
				if end > len(content) {
					end = len(content)
				}
				result := content[s.Pos:end]
				s.Pos = end
				return &objects.String{Value: result}
			},
		}, true
	case "write":
		return &objects.Builtin{
			Name: "StringIO.write",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("write() takes at least 1 argument")
				}
				str, ok := args[0].(*objects.String)
				if !ok {
					return objects.NewTypeError("write() argument must be a string, not %s", args[0].Type())
				}
				s.Buffer.WriteString(str.Value)
				s.Pos = s.Buffer.Len()
				return &objects.Integer{Value: int64(len(str.Value))}
			},
		}, true
	case "seek":
		return &objects.Builtin{
			Name: "StringIO.seek",
			Fn: func(args ...objects.Object) objects.Object {
				offset := int64(0)
				whence := 0
				if len(args) >= 1 {
					if i, ok := args[0].(*objects.Integer); ok {
						offset = i.Value
					}
				}
				if len(args) >= 2 {
					if i, ok := args[1].(*objects.Integer); ok {
						whence = int(i.Value)
					}
				}
				switch whence {
				case 0: // SEEK_SET
					if offset < 0 {
						return objects.NewValueError("negative seek position")
					}
					s.Pos = int(offset)
				case 1: // SEEK_CUR
					newPos := s.Pos + int(offset)
					if newPos < 0 {
						return objects.NewValueError("negative seek position")
					}
					s.Pos = newPos
				case 2: // SEEK_END
					newPos := s.Buffer.Len() + int(offset)
					if newPos < 0 {
						return objects.NewValueError("negative seek position")
					}
					s.Pos = newPos
				}
				return &objects.Integer{Value: int64(s.Pos)}
			},
		}, true
	case "tell":
		return &objects.Builtin{
			Name: "StringIO.tell",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.Integer{Value: int64(s.Pos)}
			},
		}, true
	case "getvalue":
		return &objects.Builtin{
			Name: "StringIO.getvalue",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.String{Value: s.Buffer.String()}
			},
		}, true
	case "close":
		return &objects.Builtin{
			Name: "StringIO.close",
			Fn: func(args ...objects.Object) objects.Object {
				s.Buffer.Reset()
				s.Pos = 0
				return objects.None_
			},
		}, true
	case "isatty":
		return &objects.Builtin{
			Name: "StringIO.isatty",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.False
			},
		}, true
	case "readable":
		return &objects.Builtin{
			Name: "StringIO.readable",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.True
			},
		}, true
	case "writable":
		return &objects.Builtin{
			Name: "StringIO.writable",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.True
			},
		}, true
	case "seekable":
		return &objects.Builtin{
			Name: "StringIO.seekable",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.True
			},
		}, true
	}
	return nil, false
}

// BytesIO is a wrapper that implements Object interface
type BytesIO struct {
	Buffer bytes.Buffer
	Pos    int
}

func NewBytesIO() *BytesIO {
	return &BytesIO{
		Buffer: bytes.Buffer{},
		Pos:    0,
	}
}

func (b *BytesIO) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (b *BytesIO) Inspect() string           { return string(b.Buffer.Bytes()) }

func (b *BytesIO) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "read":
		return &objects.Builtin{
			Name: "BytesIO.read",
			Fn: func(args ...objects.Object) objects.Object {
				size := -1
				if len(args) >= 1 {
					if i, ok := args[0].(*objects.Integer); ok {
						size = int(i.Value)
					}
				}
				if size == 0 {
					return &objects.Bytes{Value: []byte{}}
				}
				content := b.Buffer.Bytes()
				if size < 0 {
					result := content[b.Pos:]
					b.Pos = len(content)
					return &objects.Bytes{Value: result}
				}
				end := b.Pos + size
				if end > len(content) {
					end = len(content)
				}
				result := content[b.Pos:end]
				b.Pos = end
				return &objects.Bytes{Value: result}
			},
		}, true
	case "write":
		return &objects.Builtin{
			Name: "BytesIO.write",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("write() takes at least 1 argument")
				}
				bs, ok := args[0].(*objects.Bytes)
				if !ok {
					return objects.NewTypeError("write() argument must be bytes, not %s", args[0].Type())
				}
				b.Buffer.Write(bs.Value)
				b.Pos = b.Buffer.Len()
				return &objects.Integer{Value: int64(len(bs.Value))}
			},
		}, true
	case "seek":
		return &objects.Builtin{
			Name: "BytesIO.seek",
			Fn: func(args ...objects.Object) objects.Object {
				offset := int64(0)
				whence := 0
				if len(args) >= 1 {
					if i, ok := args[0].(*objects.Integer); ok {
						offset = i.Value
					}
				}
				if len(args) >= 2 {
					if i, ok := args[1].(*objects.Integer); ok {
						whence = int(i.Value)
					}
				}
				switch whence {
				case 0: // SEEK_SET
					if offset < 0 {
						return objects.NewValueError("negative seek position")
					}
					b.Pos = int(offset)
				case 1: // SEEK_CUR
					newPos := b.Pos + int(offset)
					if newPos < 0 {
						return objects.NewValueError("negative seek position")
					}
					b.Pos = newPos
				case 2: // SEEK_END
					newPos := b.Buffer.Len() + int(offset)
					if newPos < 0 {
						return objects.NewValueError("negative seek position")
					}
					b.Pos = newPos
				}
				return &objects.Integer{Value: int64(b.Pos)}
			},
		}, true
	case "tell":
		return &objects.Builtin{
			Name: "BytesIO.tell",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.Integer{Value: int64(b.Pos)}
			},
		}, true
	case "getvalue":
		return &objects.Builtin{
			Name: "BytesIO.getvalue",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.Bytes{Value: b.Buffer.Bytes()}
			},
		}, true
	case "close":
		return &objects.Builtin{
			Name: "BytesIO.close",
			Fn: func(args ...objects.Object) objects.Object {
				b.Buffer.Reset()
				b.Pos = 0
				return objects.None_
			},
		}, true
	case "isatty":
		return &objects.Builtin{
			Name: "BytesIO.isatty",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.False
			},
		}, true
	case "readable":
		return &objects.Builtin{
			Name: "BytesIO.readable",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.True
			},
		}, true
	case "writable":
		return &objects.Builtin{
			Name: "BytesIO.writable",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.True
			},
		}, true
	case "seekable":
		return &objects.Builtin{
			Name: "BytesIO.seekable",
			Fn: func(args ...objects.Object) objects.Object {
				return objects.True
			},
		}, true
	}
	return nil, false
}

// CreateIOModule creates the io module
func CreateIOModule() *objects.Module {
	module := &objects.Module{
		Name:    "io",
		Fields: make(map[string]objects.Object),
	}

	// io.StringIO
	module.Fields["StringIO"] = &objects.Builtin{
		Name: "io.StringIO",
		Fn: func(args ...objects.Object) objects.Object {
			return NewStringIO()
		},
	}

	// io.BytesIO
	module.Fields["BytesIO"] = &objects.Builtin{
		Name: "io.BytesIO",
		Fn: func(args ...objects.Object) objects.Object {
			return NewBytesIO()
		},
	}

	// io.DEFAULT_BUFFER_SIZE
	module.Fields["DEFAULT_BUFFER_SIZE"] = &objects.Integer{Value: 8192}

	// io.SEEK_SET, io.SEEK_CUR, io.SEEK_END
	module.Fields["SEEK_SET"] = &objects.Integer{Value: 0}
	module.Fields["SEEK_CUR"] = &objects.Integer{Value: 1}
	module.Fields["SEEK_END"] = &objects.Integer{Value: 2}

	return module
}
