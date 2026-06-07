package hashlib

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"

	"golang.org/x/crypto/sha3"

	"github.com/go-py/go-python/pkg/objects"
)

// Hash represents a hash object (e.g. md5, sha1, sha256, etc.)
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

func newSHA1Hash() *Hash {
	h := sha1.New()
	return &Hash{
		algorithm: "sha1",
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

func newSHA224Hash() *Hash {
	h := sha256.New224()
	return &Hash{
		algorithm: "sha224",
		hash:      h,
	}
}

func newSHA384Hash() *Hash {
	h := sha512.New384()
	return &Hash{
		algorithm: "sha384",
		hash:      h,
	}
}

func newSHA512Hash() *Hash {
	h := sha512.New()
	return &Hash{
		algorithm: "sha512",
		hash:      h,
	}
}

func newSHA3_224Hash() *Hash {
	h := sha3.New224()
	return &Hash{
		algorithm: "sha3_224",
		hash:      h,
	}
}

func newSHA3_256Hash() *Hash {
	h := sha3.New256()
	return &Hash{
		algorithm: "sha3_256",
		hash:      h,
	}
}

func newSHA3_384Hash() *Hash {
	h := sha3.New384()
	return &Hash{
		algorithm: "sha3_384",
		hash:      h,
	}
}

func newSHA3_512Hash() *Hash {
	h := sha3.New512()
	return &Hash{
		algorithm: "sha3_512",
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
	case "name":
		return &objects.String{Value: h.algorithm}, true
	}
	return nil, false
}

// makeHashBuiltin creates a Builtin that constructs a Hash and optionally feeds initial data.
func makeHashBuiltin(name string, newHash func() *Hash) *objects.Builtin {
	return &objects.Builtin{
		Name: "hashlib." + name,
		Fn: func(args ...objects.Object) objects.Object {
			h := newHash()
			if len(args) >= 1 {
				data, ok := args[0].(*objects.Bytes)
				if ok {
					h.hash.Write(data.Value)
				}
			}
			return h
		},
	}
}

// CreateHashlibModule creates the hashlib module
func CreateHashlibModule() *objects.Module {
	module := &objects.Module{
		Name:   "hashlib",
		Fields: make(map[string]objects.Object),
	}

	// Hash algorithm constructors
	module.Fields["md5"] = makeHashBuiltin("md5", newMD5Hash)
	module.Fields["sha1"] = makeHashBuiltin("sha1", newSHA1Hash)
	module.Fields["sha256"] = makeHashBuiltin("sha256", newSHA256Hash)
	module.Fields["sha224"] = makeHashBuiltin("sha224", newSHA224Hash)
	module.Fields["sha384"] = makeHashBuiltin("sha384", newSHA384Hash)
	module.Fields["sha512"] = makeHashBuiltin("sha512", newSHA512Hash)
	module.Fields["sha3_224"] = makeHashBuiltin("sha3_224", newSHA3_224Hash)
	module.Fields["sha3_256"] = makeHashBuiltin("sha3_256", newSHA3_256Hash)
	module.Fields["sha3_384"] = makeHashBuiltin("sha3_384", newSHA3_384Hash)
	module.Fields["sha3_512"] = makeHashBuiltin("sha3_512", newSHA3_512Hash)

	// algorithms_guaranteed - a set containing the names of hash algorithms
	// guaranteed to be available on this platform
	guaranteedAlgorithms := []string{
		"md5", "sha1", "sha224", "sha256", "sha384", "sha512",
		"sha3_224", "sha3_256", "sha3_384", "sha3_512",
	}
	guaranteedSet := objects.NewSet()
	for _, name := range guaranteedAlgorithms {
		guaranteedSet.Add(&objects.String{Value: name})
	}
	module.Fields["algorithms_guaranteed"] = guaranteedSet

	// algorithms_available - a set containing the names of hash algorithms
	// available in the running Python interpreter
	availableSet := objects.NewSet()
	for _, name := range guaranteedAlgorithms {
		availableSet.Add(&objects.String{Value: name})
	}
	module.Fields["algorithms_available"] = availableSet

	return module
}
