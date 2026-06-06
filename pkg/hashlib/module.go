package hashlib

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/go-py/go-python/pkg/objects"
)

// Hash represents a hash object (either md5 or sha256)
type Hash struct {
	algorithm string
	hash      hash.Hash
}

func newMD5Hash() *Hash {
	h := md5.New()
	return &Hash{
		algorithm: "md5",
		hash:      h,
	}
}

func newSHA256Hash() *Hash {
	h := sha256.New()
	return &Hash{
		algorithm: "sha256",
		hash:      h,
	}
}

func (h *Hash) Type() objects.ObjectType {
	return objects.INSTANCE_OBJ
}

func (h *Hash) Inspect() string {
	return fmt.Sprintf("<%s HASH object>", h.algorithm)
}

func (h *Hash) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "update":
		return &objects.Builtin{
			Name: "hash.update",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("update() takes at least 1 argument")
				}
				data, ok := args[0].(*objects.Bytes)
				if !ok {
					return objects.NewTypeError("update() argument must be bytes")
				}
				h.hash.Write(data.Value)
				return objects.None_
			},
		}, true
	case "digest":
		return &objects.Builtin{
			Name: "hash.digest",
			Fn: func(args ...objects.Object) objects.Object {
				sum := h.hash.Sum(nil)
				return &objects.Bytes{Value: sum}
			},
		}, true
	case "hexdigest":
		return &objects.Builtin{
			Name: "hash.hexdigest",
			Fn: func(args ...objects.Object) objects.Object {
				sum := h.hash.Sum(nil)
				return &objects.String{Value: hex.EncodeToString(sum)}
			},
		}, true
	}
	return nil, false
}

// CreateHashlibModule creates the hashlib module
func CreateHashlibModule() *objects.Module {
	module := &objects.Module{
		Name:   "hashlib",
		Fields: make(map[string]objects.Object),
	}

	module.Fields["md5"] = &objects.Builtin{
		Name: "hashlib.md5",
		Fn: func(args ...objects.Object) objects.Object {
			h := newMD5Hash()
			if len(args) >= 1 {
				data, ok := args[0].(*objects.Bytes)
				if ok {
					h.hash.Write(data.Value)
				}
			}
			return h
		},
	}

	module.Fields["sha256"] = &objects.Builtin{
		Name: "hashlib.sha256",
		Fn: func(args ...objects.Object) objects.Object {
			h := newSHA256Hash()
			if len(args) >= 1 {
				data, ok := args[0].(*objects.Bytes)
				if ok {
					h.hash.Write(data.Value)
				}
			}
			return h
		},
	}

	return module
}
