package pathlib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-py/go-python/pkg/objects"
)

// Path represents a filesystem path, analogous to Python's pathlib.Path
type Path struct {
	path string
}

// NewPath creates a new Path object
func NewPath(path string) *Path {
	return &Path{path: path}
}

func (p *Path) Type() objects.ObjectType { return objects.INSTANCE_OBJ }

func (p *Path) Inspect() string {
	return fmt.Sprintf("Path('%s')", p.path)
}

func (p *Path) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "__str__":
		return &objects.Builtin{
			Name: "Path.__str__",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.String{Value: p.path}
			},
		}, true
	case "__repr__":
		return &objects.Builtin{
			Name: "Path.__repr__",
			Fn: func(args ...objects.Object) objects.Object {
				return &objects.String{Value: p.Inspect()}
			},
		}, true
	case "exists":
		return &objects.Builtin{
			Name: "Path.exists",
			Fn: func(args ...objects.Object) objects.Object {
				_, err := os.Stat(p.path)
				if err != nil {
					return objects.False
				}
				return objects.True
			},
		}, true
	case "is_file":
		return &objects.Builtin{
			Name: "Path.is_file",
			Fn: func(args ...objects.Object) objects.Object {
				info, err := os.Stat(p.path)
				if err != nil {
					return objects.False
				}
				return &objects.Boolean{Value: !info.IsDir()}
			},
		}, true
	case "is_dir":
		return &objects.Builtin{
			Name: "Path.is_dir",
			Fn: func(args ...objects.Object) objects.Object {
				info, err := os.Stat(p.path)
				if err != nil {
					return objects.False
				}
				return &objects.Boolean{Value: info.IsDir()}
			},
		}, true
	case "name":
		return &objects.String{Value: filepath.Base(p.path)}, true
	case "suffix":
		ext := filepath.Ext(p.path)
		return &objects.String{Value: ext}, true
	case "stem":
		base := filepath.Base(p.path)
		ext := filepath.Ext(base)
		return &objects.String{Value: strings.TrimSuffix(base, ext)}, true
	case "parent":
		dir := filepath.Dir(p.path)
		return NewPath(dir), true
	case "joinpath":
		return &objects.Builtin{
			Name: "Path.joinpath",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("joinpath() takes at least 1 argument")
				}
				joined := p.path
				for _, arg := range args {
					var other string
					switch v := arg.(type) {
					case *objects.String:
						other = v.Value
					case *Path:
						other = v.path
					default:
						return objects.NewTypeError("joinpath() arguments must be strings or Path objects")
					}
					joined = filepath.Join(joined, other)
				}
				return NewPath(joined)
			},
		}, true
	case "resolve":
		return &objects.Builtin{
			Name: "Path.resolve",
			Fn: func(args ...objects.Object) objects.Object {
				abs, err := filepath.Abs(p.path)
				if err != nil {
					return objects.NewError("resolve() error: %s", err.Error())
				}
				return NewPath(abs)
			},
		}, true
	case "mkdir":
		return &objects.Builtin{
			Name: "Path.mkdir",
			Fn: func(args ...objects.Object) objects.Object {
				parents := false
				existOK := false

				if len(args) >= 1 {
					if kwargs, ok := args[0].(*objects.Dict); ok {
						if v, ok := kwargs.Get(&objects.String{Value: "parents"}); ok {
							if b, ok := v.(*objects.Boolean); ok {
								parents = b.Value
							}
						}
						if v, ok := kwargs.Get(&objects.String{Value: "exist_ok"}); ok {
							if b, ok := v.(*objects.Boolean); ok {
								existOK = b.Value
							}
						}
					}
				}

				// Also check positional boolean args for simpler calling
				if len(args) >= 1 {
					if b, ok := args[0].(*objects.Boolean); ok {
						parents = b.Value
					}
				}
				if len(args) >= 2 {
					if b, ok := args[1].(*objects.Boolean); ok {
						existOK = b.Value
					}
				}

				var err error
				if parents {
					err = os.MkdirAll(p.path, 0755)
				} else {
					err = os.Mkdir(p.path, 0755)
				}

				if err != nil {
					if os.IsExist(err) && existOK {
						return objects.None_
					}
					return objects.NewError("mkdir() error: %s", err.Error())
				}
				return objects.None_
			},
		}, true
	case "read_text":
		return &objects.Builtin{
			Name: "Path.read_text",
			Fn: func(args ...objects.Object) objects.Object {
				data, err := os.ReadFile(p.path)
				if err != nil {
					return objects.NewError("read_text() error: %s", err.Error())
				}
				return &objects.String{Value: string(data)}
			},
		}, true
	case "write_text":
		return &objects.Builtin{
			Name: "Path.write_text",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("write_text() takes at least 1 argument")
				}
				var data string
				switch v := args[0].(type) {
				case *objects.String:
					data = v.Value
				default:
					return objects.NewTypeError("write_text() argument must be a string")
				}
				err := os.WriteFile(p.path, []byte(data), 0644)
				if err != nil {
					return objects.NewError("write_text() error: %s", err.Error())
				}
				return objects.None_
			},
		}, true
	case "__truediv__":
		return &objects.Builtin{
			Name: "Path.__truediv__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__truediv__ takes at least 1 argument")
				}
				var other string
				switch v := args[0].(type) {
				case *objects.String:
					other = v.Value
				case *Path:
					other = v.path
				default:
					return objects.NewTypeError("__truediv__ argument must be a string or Path")
				}
				return NewPath(filepath.Join(p.path, other))
			},
		}, true
	case "__eq__":
		return &objects.Builtin{
			Name: "Path.__eq__",
			Fn: func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return objects.NewTypeError("__eq__ takes at least 1 argument")
				}
				switch v := args[0].(type) {
				case *Path:
					return &objects.Boolean{Value: p.path == v.path}
				case *objects.String:
					return &objects.Boolean{Value: p.path == v.Value}
				default:
					return objects.False
				}
			},
		}, true
	}
	return nil, false
}

// CreatePathlibModule creates the pathlib module
func CreatePathlibModule() *objects.Module {
	module := &objects.Module{
		Name:   "pathlib",
		Fields: make(map[string]objects.Object),
	}

	// pathlib.Path - callable that creates a new Path instance
	module.Fields["Path"] = &objects.Builtin{
		Name: "pathlib.Path",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return NewPath(".")
			}
			switch v := args[0].(type) {
			case *objects.String:
				return NewPath(v.Value)
			case *Path:
				return NewPath(v.path)
			default:
				return objects.NewTypeError("Path() argument must be a string or Path")
			}
		},
	}

	// pathlib.PurePosixPath
	module.Fields["PurePosixPath"] = &objects.Builtin{
		Name: "pathlib.PurePosixPath",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return NewPath(".")
			}
			if s, ok := args[0].(*objects.String); ok {
				return NewPath(s.Value)
			}
			return NewPath(".")
		},
	}

	// pathlib.PureWindowsPath
	module.Fields["PureWindowsPath"] = &objects.Builtin{
		Name: "pathlib.PureWindowsPath",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return NewPath(".")
			}
			if s, ok := args[0].(*objects.String); ok {
				return NewPath(s.Value)
			}
			return NewPath(".")
		},
	}

	return module
}
