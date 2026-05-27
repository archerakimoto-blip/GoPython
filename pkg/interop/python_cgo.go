package interop

/*
#cgo LDFLAGS: -lpython3.14 -ldl -lm -L/root/.pyenv/versions/3.14.4/lib -Wl,-rpath,/root/.pyenv/versions/3.14.4/lib
#cgo CFLAGS: -I/root/.pyenv/versions/3.14.4/include/python3.14
#include <Python.h>

static inline PyObject* PyTuple_FromObjects(PyObject** objs, int n) {
    PyObject* tuple = PyTuple_New(n);
    if (!tuple) return NULL;
    for (int i = 0; i < n; i++) {
        PyTuple_SetItem(tuple, i, objs[i]);
    }
    return tuple;
}
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/go-py/go-python/pkg/objects"
)

type PythonRuntime struct {
	initialized bool
}

var runtime = &PythonRuntime{}

func (p *PythonRuntime) Initialize() error {
	if p.initialized {
		return nil
	}
	C.Py_Initialize()
	p.initialized = true
	return nil
}

func (p *PythonRuntime) Finalize() {
	if p.initialized {
		C.Py_Finalize()
		p.initialized = false
	}
}

func (p *PythonRuntime) IsInitialized() bool {
	return p.initialized
}

func Initialize() error {
	return runtime.Initialize()
}

func Finalize() {
	runtime.Finalize()
}

func IsInitialized() bool {
	return runtime.IsInitialized()
}

func goToCPython(obj objects.Object) *C.PyObject {
	if obj == nil || obj == objects.None_ {
		C.Py_IncRef(C.Py_None)
		return C.Py_None
	}

	switch v := obj.(type) {
	case *objects.Integer:
		return C.PyLong_FromLongLong(C.longlong(v.Value))
	case *objects.Float:
		return C.PyFloat_FromDouble(C.double(v.Value))
	case *objects.Boolean:
		if v.Value {
			C.Py_IncRef(C.Py_True)
			return C.Py_True
		} else {
			C.Py_IncRef(C.Py_False)
			return C.Py_False
		}
	case *objects.String:
		cstr := C.CString(v.Value)
		defer C.free(unsafe.Pointer(cstr))
		return C.PyUnicode_FromString(cstr)
	case *objects.List:
		length := len(v.Elements)
		pylist := C.PyList_New(C.Py_ssize_t(length))
		for i, elem := range v.Elements {
			pyelem := goToCPython(elem)
			if pyelem == nil {
				C.Py_DecRef(pylist)
				return nil
			}
			C.PyList_SetItem(pylist, C.Py_ssize_t(i), pyelem)
		}
		return pylist
	case *objects.Dict:
		pydict := C.PyDict_New()
		for key, val := range v.Pairs {
			pykey := goToCPython(&objects.String{Value: key})
			if pykey == nil {
				C.Py_DecRef(pydict)
				return nil
			}
			pyval := goToCPython(val)
			if pyval == nil {
				C.Py_DecRef(pykey)
				C.Py_DecRef(pydict)
				return nil
			}
			C.PyDict_SetItem(pydict, pykey, pyval)
			C.Py_DecRef(pykey)
			C.Py_DecRef(pyval)
		}
		return pydict
	default:
		cstr := C.CString(fmt.Sprintf("%v", obj.Inspect()))
		defer C.free(unsafe.Pointer(cstr))
		return C.PyUnicode_FromString(cstr)
	}
}

func cpythonToGo(pyobj *C.PyObject) objects.Object {
	if pyobj == C.Py_None {
		return objects.None_
	}

	if pyobj == C.Py_True {
		return objects.True
	}

	if pyobj == C.Py_False {
		return objects.False
	}

	if C.PyLong_AsLongLong(pyobj) != -1 || C.PyErr_Occurred() == nil {
		val := C.PyLong_AsLongLong(pyobj)
		C.PyErr_Clear()
		return &objects.Integer{Value: int64(val)}
	}
	C.PyErr_Clear()

	if C.PyFloat_AsDouble(pyobj) != -1.0 || C.PyErr_Occurred() == nil {
		val := C.PyFloat_AsDouble(pyobj)
		C.PyErr_Clear()
		return &objects.Float{Value: float64(val)}
	}
	C.PyErr_Clear()

	cstr := C.PyUnicode_AsUTF8(pyobj)
	if cstr != nil {
		return &objects.String{Value: C.GoString(cstr)}
	}

	pyType := C.PyObject_Type(pyobj)
	listType := C.PyObject_Type(C.PyList_New(0))
	if C.PyObject_RichCompareBool(pyType, listType, C.Py_EQ) != 0 {
		C.Py_DecRef(listType)
		length := C.PyList_Size(pyobj)
		elements := make([]objects.Object, length)
		for i := C.Py_ssize_t(0); i < length; i++ {
			item := C.PyList_GetItem(pyobj, i)
			if item == nil {
				continue
			}
			C.Py_IncRef(item)
			elements[i] = cpythonToGo(item)
			C.Py_DecRef(item)
		}
		return objects.NewList(elements)
	}
	C.Py_DecRef(listType)

	dictType := C.PyObject_Type(C.PyDict_New())
	if C.PyObject_RichCompareBool(pyType, dictType, C.Py_EQ) != 0 {
		C.Py_DecRef(dictType)
		dict := objects.NewDict()
		keys := C.PyDict_Keys(pyobj)
		if keys == nil {
			return dict
		}
		defer C.Py_DecRef(keys)

		length := C.PyList_Size(keys)
		for i := C.Py_ssize_t(0); i < length; i++ {
			pykey := C.PyList_GetItem(keys, i)
			if pykey == nil {
				continue
			}
			C.Py_IncRef(pykey)

			pyval := C.PyDict_GetItem(pyobj, pykey)
			if pyval == nil {
				C.Py_DecRef(pykey)
				continue
			}
			C.Py_IncRef(pyval)

			keyObj := cpythonToGo(pykey)
			valObj := cpythonToGo(pyval)

			if strKey, ok := keyObj.(*objects.String); ok {
				dict.Set(strKey, valObj)
			}

			C.Py_DecRef(pykey)
			C.Py_DecRef(pyval)
		}
		return dict
	}
	C.Py_DecRef(dictType)

	strObj := C.PyObject_Str(pyobj)
	if strObj == nil {
		return objects.None_
	}
	defer C.Py_DecRef(strObj)

	return &objects.String{Value: C.GoString(C.PyUnicode_AsUTF8(strObj))}
}

func ImportModule(name string) (*CPythonModule, error) {
	if !IsInitialized() {
		if err := Initialize(); err != nil {
			return nil, err
		}
	}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	pymod := C.PyImport_ImportModule(cname)
	if pymod == nil {
		C.PyErr_Print()
		return nil, fmt.Errorf("failed to import module: %s", name)
	}

	return &CPythonModule{pyobj: pymod}, nil
}

type CPythonModule struct {
	pyobj *C.PyObject
}

func (m *CPythonModule) GetAttr(name string) (*CPythonObject, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	pyattr := C.PyObject_GetAttrString(m.pyobj, cname)
	if pyattr == nil {
		C.PyErr_Print()
		return nil, fmt.Errorf("failed to get attribute: %s", name)
	}

	return &CPythonObject{pyobj: pyattr}, nil
}

func (m *CPythonModule) CallMethod(name string, args ...objects.Object) (objects.Object, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	pyfunc := C.PyObject_GetAttrString(m.pyobj, cname)
	if pyfunc == nil {
		C.PyErr_Print()
		return nil, fmt.Errorf("method not found: %s", name)
	}
	defer C.Py_DecRef(pyfunc)

	if C.PyCallable_Check(pyfunc) == 0 {
		return nil, fmt.Errorf("%s is not callable", name)
	}

	return callObject(pyfunc, args...)
}

func (m *CPythonModule) GetDict() (objects.Object, error) {
	pydict := C.PyModule_GetDict(m.pyobj)
	if pydict == nil {
		return nil, fmt.Errorf("failed to get module dict")
	}
	C.Py_IncRef(pydict)
	defer C.Py_DecRef(pydict)
	return cpythonToGo(pydict), nil
}

func (m *CPythonModule) DecRef() {
	if m.pyobj != nil {
		C.Py_DecRef(m.pyobj)
		m.pyobj = nil
	}
}

type CPythonObject struct {
	pyobj *C.PyObject
}

func (o *CPythonObject) Call(args ...objects.Object) (objects.Object, error) {
	if C.PyCallable_Check(o.pyobj) == 0 {
		return nil, fmt.Errorf("object is not callable")
	}
	return callObject(o.pyobj, args...)
}

func (o *CPythonObject) GetAttr(name string) (*CPythonObject, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	pyattr := C.PyObject_GetAttrString(o.pyobj, cname)
	if pyattr == nil {
		C.PyErr_Print()
		return nil, fmt.Errorf("failed to get attribute: %s", name)
	}

	return &CPythonObject{pyobj: pyattr}, nil
}

func (o *CPythonObject) ToGo() objects.Object {
	return cpythonToGo(o.pyobj)
}

func (o *CPythonObject) DecRef() {
	if o.pyobj != nil {
		C.Py_DecRef(o.pyobj)
		o.pyobj = nil
	}
}

func callObject(pyfunc *C.PyObject, args ...objects.Object) (objects.Object, error) {
	argc := len(args)
	var pyargs *C.PyObject

	if argc == 0 {
		pyargs = C.PyTuple_New(0)
	} else {
		pyargs = C.PyTuple_New(C.Py_ssize_t(argc))
		if pyargs == nil {
			return nil, fmt.Errorf("failed to create argument tuple")
		}

		for i, arg := range args {
			pyarg := goToCPython(arg)
			if pyarg == nil {
				C.Py_DecRef(pyargs)
				return nil, fmt.Errorf("failed to convert argument %d", i)
			}
			C.PyTuple_SetItem(pyargs, C.Py_ssize_t(i), pyarg)
		}
	}
	defer C.Py_DecRef(pyargs)

	result := C.PyObject_CallObject(pyfunc, pyargs)
	if result == nil {
		C.PyErr_Print()
		return nil, fmt.Errorf("function call failed")
	}
	defer C.Py_DecRef(result)

	return cpythonToGo(result), nil
}

func Evaluate(expression string) (objects.Object, error) {
	if !IsInitialized() {
		if err := Initialize(); err != nil {
			return nil, err
		}
	}

	cexp := C.CString(expression)
	defer C.free(unsafe.Pointer(cexp))

	globals := C.PyDict_New()
	if globals == nil {
		return nil, fmt.Errorf("failed to create globals dict")
	}
	defer C.Py_DecRef(globals)

	locals := C.PyDict_New()
	if locals == nil {
		return nil, fmt.Errorf("failed to create locals dict")
	}
	defer C.Py_DecRef(locals)

	result := C.PyRun_String(cexp, C.Py_eval_input, globals, locals)
	if result == nil {
		C.PyErr_Print()
		return nil, fmt.Errorf("failed to evaluate expression")
	}
	defer C.Py_DecRef(result)

	return cpythonToGo(result), nil
}

func ExecuteFile(filename string) error {
	if !IsInitialized() {
		if err := Initialize(); err != nil {
			return err
		}
	}

	cfile := C.CString(filename)
	defer C.free(unsafe.Pointer(cfile))

	file := C.fopen(cfile, C.CString("r"))
	if file == nil {
		return fmt.Errorf("failed to open file: %s", filename)
	}
	defer C.fclose(file)

	result := C.PyRun_File(file, cfile, C.Py_file_input, C.PyEval_GetGlobals(), C.PyEval_GetLocals())
	if result == nil {
		C.PyErr_Print()
		return fmt.Errorf("failed to execute file: %s", filename)
	}
	C.Py_DecRef(result)
	return nil
}