package asyncio

import (
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/go-py/go-python/pkg/objects"
)

// AsyncTask represents a coroutine task managed by the asyncio event loop.
type AsyncTask struct {
	Coro   objects.Object
	Result objects.Object
	Done   bool
	mu     sync.Mutex
}

func (t *AsyncTask) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (t *AsyncTask) Inspect() string {
	return "<asyncio.Task>"
}

func (t *AsyncTask) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "result":
		return &objects.Builtin{
			Name: "Task.result",
			Fn: func(args ...objects.Object) objects.Object {
				t.mu.Lock()
				defer t.mu.Unlock()
				if !t.Done {
					return objects.NewError("result is not set")
				}
				return t.Result
			},
		}, true
	case "done":
		return &objects.Builtin{
			Name: "Task.done",
			Fn: func(args ...objects.Object) objects.Object {
				t.mu.Lock()
				defer t.mu.Unlock()
				if t.Done {
					return objects.True
				}
				return objects.False
			},
		}, true
	case "__await__":
		return &objects.Builtin{
			Name: "Task.__await__",
			Fn: func(args ...objects.Object) objects.Object {
				return t
			},
		}, true
	}
	return nil, false
}

// AsyncEvent represents an asyncio.Event with set/clear/is_set/wait methods.
type AsyncEvent struct {
	isSet bool
	mu    sync.Mutex
}

func (e *AsyncEvent) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (e *AsyncEvent) Inspect() string {
	return "<asyncio.Event>"
}

func (e *AsyncEvent) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "set":
		return &objects.Builtin{
			Name: "Event.set",
			Fn: func(args ...objects.Object) objects.Object {
				e.mu.Lock()
				defer e.mu.Unlock()
				e.isSet = true
				return objects.None_
			},
		}, true
	case "clear":
		return &objects.Builtin{
			Name: "Event.clear",
			Fn: func(args ...objects.Object) objects.Object {
				e.mu.Lock()
				defer e.mu.Unlock()
				e.isSet = false
				return objects.None_
			},
		}, true
	case "is_set":
		return &objects.Builtin{
			Name: "Event.is_set",
			Fn: func(args ...objects.Object) objects.Object {
				e.mu.Lock()
				defer e.mu.Unlock()
				if e.isSet {
					return objects.True
				}
				return objects.False
			},
		}, true
	case "wait":
		return &objects.Builtin{
			Name: "Event.wait",
			Fn: func(args ...objects.Object) objects.Object {
				// Simple spin-wait with a short sleep
				for {
					e.mu.Lock()
					set := e.isSet
					e.mu.Unlock()
					if set {
						return objects.True
					}
					time.Sleep(1 * time.Millisecond)
				}
			},
		}, true
	}
	return nil, false
}

// AsyncFile represents an asynchronously-opened file with read/write/close methods.
type AsyncFile struct {
	filename string
	mode     string
	data     []byte
	offset   int
	closed   bool
	mu       sync.Mutex
}

func (af *AsyncFile) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (af *AsyncFile) Inspect() string {
	return "<asyncio.File '" + af.filename + "' mode='" + af.mode + "'>"
}

func (af *AsyncFile) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "read":
		return &objects.Builtin{
			Name: "AsyncFile.read",
			Fn: func(args ...objects.Object) objects.Object {
				af.mu.Lock()
				defer af.mu.Unlock()
				if af.closed {
					return objects.NewError("I/O operation on closed file")
				}
				return &objects.String{Value: string(af.data[af.offset:])}
			},
		}, true
	case "readline":
		return &objects.Builtin{
			Name: "AsyncFile.readline",
			Fn: func(args ...objects.Object) objects.Object {
				af.mu.Lock()
				defer af.mu.Unlock()
				if af.closed {
					return objects.NewError("I/O operation on closed file")
				}
				// Find next newline
				remaining := af.data[af.offset:]
				for i, b := range remaining {
					if b == '\n' {
						line := string(remaining[:i+1])
						af.offset += i + 1
						return &objects.String{Value: line}
					}
				}
				// No newline found, return rest
				line := string(remaining)
				af.offset = len(af.data)
				return &objects.String{Value: line}
			},
		}, true
	case "write":
		return &objects.Builtin{
			Name: "AsyncFile.write",
			Fn: func(args ...objects.Object) objects.Object {
				af.mu.Lock()
				defer af.mu.Unlock()
				if af.closed {
					return objects.NewError("I/O operation on closed file")
				}
				if len(args) < 1 {
					return objects.NewTypeError("write() takes at least 1 argument")
				}
				content, ok := args[0].(*objects.String)
				if !ok {
					return objects.NewTypeError("write() argument must be a string")
				}
				af.data = append(af.data, []byte(content.Value)...)
				return &objects.Integer{Value: int64(len(content.Value))}
			},
		}, true
	case "close":
		return &objects.Builtin{
			Name: "AsyncFile.close",
			Fn: func(args ...objects.Object) objects.Object {
				af.mu.Lock()
				defer af.mu.Unlock()
				if !af.closed {
					// Write back if in write/append mode
					if af.mode == "w" || af.mode == "a" {
						_ = os.WriteFile(af.filename, af.data, 0644)
					}
					af.closed = true
				}
				return objects.None_
			},
		}, true
	case "closed":
		return &objects.Builtin{
			Name: "AsyncFile.closed",
			Fn: func(args ...objects.Object) objects.Object {
				af.mu.Lock()
				defer af.mu.Unlock()
				if af.closed {
					return objects.True
				}
				return objects.False
			},
		}, true
	case "name":
		return &objects.String{Value: af.filename}, true
	case "mode":
		return &objects.String{Value: af.mode}, true
	}
	return nil, false
}

// AsyncSubprocess represents an async subprocess with communicate/wait methods.
type AsyncSubprocess struct {
	cmd    *exec.Cmd
	stdout []byte
	stderr []byte
	done   bool
	mu     sync.Mutex
}

func (asp *AsyncSubprocess) Type() objects.ObjectType { return objects.INSTANCE_OBJ }
func (asp *AsyncSubprocess) Inspect() string {
	return "<asyncio.subprocess>"
}

func (asp *AsyncSubprocess) GetAttr(name string) (objects.Object, bool) {
	switch name {
	case "wait":
		return &objects.Builtin{
			Name: "Subprocess.wait",
			Fn: func(args ...objects.Object) objects.Object {
				asp.mu.Lock()
				if asp.done {
					asp.mu.Unlock()
					return objects.None_
				}
				asp.mu.Unlock()
				err := asp.cmd.Wait()
				asp.mu.Lock()
				asp.done = true
				asp.mu.Unlock()
				if err != nil {
					return objects.NewError("subprocess error: %s", err.Error())
				}
				return objects.None_
			},
		}, true
	case "communicate":
		return &objects.Builtin{
			Name: "Subprocess.communicate",
			Fn: func(args ...objects.Object) objects.Object {
				asp.mu.Lock()
				if asp.done {
					result := objects.NewList([]objects.Object{
						&objects.String{Value: string(asp.stdout)},
						&objects.String{Value: string(asp.stderr)},
					})
					asp.mu.Unlock()
					return result
				}
				asp.mu.Unlock()

				stdout, stderr, err := func() ([]byte, []byte, error) {
					out, err := asp.cmd.Output()
					if exitErr, ok := err.(*exec.ExitError); ok {
						return out, exitErr.Stderr, nil
					}
					return out, nil, err
				}()

				asp.mu.Lock()
				asp.stdout = stdout
				asp.stderr = stderr
				asp.done = true
				asp.mu.Unlock()

				if err != nil {
					return objects.NewError("subprocess error: %s", err.Error())
				}

				return objects.NewList([]objects.Object{
					&objects.String{Value: string(stdout)},
					&objects.String{Value: string(stderr)},
				})
			},
		}, true
	case "returncode":
		return &objects.Builtin{
			Name: "Subprocess.returncode",
			Fn: func(args ...objects.Object) objects.Object {
				asp.mu.Lock()
				defer asp.mu.Unlock()
				if !asp.done {
					return objects.None_
				}
				if asp.cmd.ProcessState != nil {
					return &objects.Integer{Value: int64(asp.cmd.ProcessState.ExitCode())}
				}
				return &objects.Integer{Value: 0}
			},
		}, true
	}
	return nil, false
}

// CreateAsyncioModule creates the asyncio module.
func CreateAsyncioModule() *objects.Module {
	module := &objects.Module{
		Name:   "asyncio",
		Fields: make(map[string]objects.Object),
	}

	// asyncio.run(coro) — execute coroutine synchronously
	module.Fields["run"] = &objects.Builtin{
		Name: "asyncio.run",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("asyncio.run() takes at least 1 argument")
			}
			coro := args[0]
			// If it's an Async object, try to get its result synchronously
			if asyncObj, ok := coro.(*objects.Async); ok {
				if asyncObj.Done {
					return asyncObj.Result
				}
				// Execute the coroutine synchronously by calling it
				result := objects.CallFunction(coro)
				return result
			}
			// If it's a Task, wait for it
			if task, ok := coro.(*AsyncTask); ok {
				for {
					task.mu.Lock()
					done := task.Done
					task.mu.Unlock()
					if done {
						return task.Result
					}
					time.Sleep(1 * time.Millisecond)
				}
			}
			// Otherwise, try calling it
			return objects.CallFunction(coro)
		},
	}

	// asyncio.create_task(coro) — create Task object
	module.Fields["create_task"] = &objects.Builtin{
		Name: "asyncio.create_task",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("asyncio.create_task() takes at least 1 argument")
			}
			coro := args[0]
			task := &AsyncTask{
				Coro:   coro,
				Result: objects.None_,
				Done:   false,
			}
			return task
		},
	}

	// asyncio.sleep(delay) — time.Sleep
	module.Fields["sleep"] = &objects.Builtin{
		Name: "asyncio.sleep",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("asyncio.sleep() takes at least 1 argument")
			}
			var delay float64
			switch d := args[0].(type) {
			case *objects.Integer:
				delay = float64(d.Value)
			case *objects.Float:
				delay = d.Value
			default:
				return objects.NewTypeError("asyncio.sleep() argument must be a number")
			}
			time.Sleep(time.Duration(delay * float64(time.Second)))
			return objects.None_
		},
	}

	// asyncio.gather(*coros) — concurrent execution with goroutines + sync.WaitGroup
	module.Fields["gather"] = &objects.Builtin{
		Name: "asyncio.gather",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) == 0 {
				return objects.NewList([]objects.Object{})
			}

			var wg sync.WaitGroup
			results := make([]objects.Object, len(args))

			for i, coro := range args {
				wg.Add(1)
				idx := i
				go func(c objects.Object) {
					defer wg.Done()
					// Execute the coroutine
					var result objects.Object
					if asyncObj, ok := c.(*objects.Async); ok {
						if asyncObj.Done {
							result = asyncObj.Result
						} else {
							result = objects.CallFunction(c)
						}
					} else {
						result = objects.CallFunction(c)
					}
					if result == nil {
						result = objects.None_
					}
					results[idx] = result
				}(coro)
			}

			wg.Wait()

			return objects.NewList(results)
		},
	}

	// asyncio.open(filename, mode='r') — async file I/O
	module.Fields["open"] = &objects.Builtin{
		Name: "asyncio.open",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("asyncio.open() takes at least 1 argument")
			}
			filename, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("filename must be a string")
			}
			mode := "r"
			if len(args) >= 2 {
				if modeObj, ok := args[1].(*objects.String); ok {
					mode = modeObj.Value
				}
			}

			af := &AsyncFile{
				filename: filename.Value,
				mode:     mode,
				closed:   false,
			}

			switch mode {
			case "r":
				data, err := os.ReadFile(filename.Value)
				if err != nil {
					return objects.NewError("could not open file '%s': %s", filename.Value, err.Error())
				}
				af.data = data
			case "w":
				af.data = []byte{}
			case "a":
				data, err := os.ReadFile(filename.Value)
				if err != nil {
					data = []byte{}
				}
				af.data = data
				af.offset = len(data)
			default:
				return objects.NewError("unsupported mode '%s' for asyncio.open()", mode)
			}

			return af
		},
	}

	// asyncio.create_subprocess_exec(*args) — async subprocess execution
	module.Fields["create_subprocess_exec"] = &objects.Builtin{
		Name: "asyncio.create_subprocess_exec",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 1 {
				return objects.NewTypeError("create_subprocess_exec() takes at least 1 argument")
			}
			cmdStr, ok := args[0].(*objects.String)
			if !ok {
				return objects.NewTypeError("command must be a string")
			}
			cmdArgs := make([]string, 0, len(args)-1)
			for _, arg := range args[1:] {
				if s, ok := arg.(*objects.String); ok {
					cmdArgs = append(cmdArgs, s.Value)
				} else {
					cmdArgs = append(cmdArgs, arg.Inspect())
				}
			}
			cmd := exec.Command(cmdStr.Value, cmdArgs...)
			_ = cmd.Start() // Start the process asynchronously
			return &AsyncSubprocess{
				cmd:  cmd,
				done: false,
			}
		},
	}

	// asyncio.Event — event class with set/clear/is_set/wait methods
	module.Fields["Event"] = &objects.Builtin{
		Name: "asyncio.Event",
		Fn: func(args ...objects.Object) objects.Object {
			return &AsyncEvent{
				isSet: false,
			}
		},
	}

	// asyncio.Future — simple future placeholder
	module.Fields["Future"] = &objects.Builtin{
		Name: "asyncio.Future",
		Fn: func(args ...objects.Object) objects.Object {
			return &AsyncTask{
				Coro:   objects.None_,
				Result: objects.None_,
				Done:   false,
			}
		},
	}

	// asyncio.wait_for(coro, timeout) — wait with timeout
	module.Fields["wait_for"] = &objects.Builtin{
		Name: "asyncio.wait_for",
		Fn: func(args ...objects.Object) objects.Object {
			if len(args) < 2 {
				return objects.NewTypeError("asyncio.wait_for() takes at least 2 arguments")
			}
			coro := args[0]
			var timeout float64
			switch t := args[1].(type) {
			case *objects.Integer:
				timeout = float64(t.Value)
			case *objects.Float:
				timeout = t.Value
			default:
				return objects.NewTypeError("timeout must be a number")
			}

			done := make(chan objects.Object, 1)
			go func() {
				result := objects.CallFunction(coro)
				done <- result
			}()

			select {
			case result := <-done:
				return result
			case <-time.After(time.Duration(timeout * float64(time.Second))):
				return objects.NewError("asyncio.wait_for() timed out after %.1f seconds", timeout)
			}
		},
	}

	return module
}
