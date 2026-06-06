package base64

import (
	"encoding/base64"

	"github.com/go-py/go-python/pkg/objects"
)

// CreateBase64Module creates the base64 module
func CreateBase64Module() *objects.Module {
	module := &objects.Module{
		Name:   "base64",
		Fields: make(map[string]objects.Object),
	}

	// b64encode(s) - Encode bytes to base64
	module.Fields["b64encode"] = &objects.Builtin{
		Name: "base64.b64encode",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("b64encode() takes at least 1 argument")
			}

			var data []byte
			switch arg := args[0].(type) {
			case *objects.Bytes:
				data = arg.Value
			case *objects.String:
				data = []byte(arg.Value)
			default:
				return objects.NewTypeError("b64encode() argument must be bytes or string")
			}

			encoded := base64.StdEncoding.EncodeToString(data)
			return &objects.Bytes{Value: []byte(encoded)}
		},
	}

	// b64decode(s) - Decode base64 to bytes
	module.Fields["b64decode"] = &objects.Builtin{
		Name: "base64.b64decode",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("b64decode() takes at least 1 argument")
			}

			var data string
			switch arg := args[0].(type) {
			case *objects.Bytes:
				data = string(arg.Value)
			case *objects.String:
				data = arg.Value
			default:
				return objects.NewTypeError("b64decode() argument must be bytes or string")
			}

			decoded, err := base64.StdEncoding.DecodeString(data)
			if err != nil {
				return objects.NewError("b64decode() error: %s", err.Error())
			}

			return &objects.Bytes{Value: decoded}
		},
	}

	return module
}
