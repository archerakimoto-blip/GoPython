package interop

import (
	"github.com/go-py/go-python/pkg/objects"
)

func CreateCPythonModule() *objects.Module {
	module := &objects.Module{
		Name:    "cpython",
		Fields: make(map[string]objects.Object),
	}

	module.Fields["import"] = &objects.Builtin{
		Name: "cpython.import",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("import() takes exactly 1 argument")
			}
			name, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("import() argument must be a string")
			}

			pymod, err := ImportModule(name.Value)
			if err != nil {
				return objects.NewError(err.Error())
			}

			return wrapModule(pymod)
		},
	}

	module.Fields["eval"] = &objects.Builtin{
		Name: "cpython.eval",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("eval() takes exactly 1 argument")
			}
			expr, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("eval() argument must be a string")
			}

			result, err := Evaluate(expr.Value)
			if err != nil {
				return objects.NewError(err.Error())
			}
			return result
		},
	}

	module.Fields["exec"] = &objects.Builtin{
		Name: "cpython.exec",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("exec() takes exactly 1 argument")
			}
			code, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("exec() argument must be a string")
			}

			_, err := Evaluate(code.Value)
			if err != nil {
				return objects.NewError(err.Error())
			}
			return objects.None_
		},
	}

	module.Fields["initialize"] = &objects.Builtin{
		Name: "cpython.initialize",
		Fn: func(args ...objects.Object) objects.Object {
			if err := Initialize(); err != nil {
				return objects.NewError(err.Error())
			}
			return objects.None_
		},
	}

	module.Fields["finalize"] = &objects.Builtin{
		Name: "cpython.finalize",
		Fn: func(args ...objects.Object) objects.Object {
			Finalize()
			return objects.None_
		},
	}

	return module
}

type wrappedModule struct {
	*objects.Module
	pyModule *CPythonModule
}

func wrapModule(pyModule *CPythonModule) *objects.Module {
	module := &objects.Module{
		Name:    "cpython_module",
		Fields: make(map[string]objects.Object),
	}

	module.Fields["call"] = &objects.Builtin{
		Name: "cpython_module.call",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("call() takes at least 1 argument")
			}
			funcName, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("call() first argument must be a string")
			}

			callArgs := args[1:]
			result, err := pyModule.CallMethod(funcName.Value, callArgs...)
			if err != nil {
				return objects.NewError(err.Error())
			}
			return result
		},
	}

	module.Fields["get"] = &objects.Builtin{
		Name: "cpython_module.get",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) != 1 {
				return objects.NewTypeError("get() takes exactly 1 argument")
			}
			attrName, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("get() argument must be a string")
			}

			pyObj, err := pyModule.GetAttr(attrName.Value)
			if err != nil {
				return objects.NewError(err.Error())
			}
			defer pyObj.DecRef()

			return pyObj.ToGo()
		},
	}

	return module
}