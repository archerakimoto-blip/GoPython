package debugger

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/vm"
)

type Debugger struct {
	vm          *vm.VM
	breakpoints map[int]bool
	stepMode    bool
	stepOver    bool
	stepOut     bool
	depth       int
	scanner     *bufio.Scanner
}

func New(vm *vm.VM) *Debugger {
	return &Debugger{
		vm:          vm,
		breakpoints: make(map[int]bool),
		stepMode:    false,
		stepOver:    false,
		stepOut:     false,
		depth:       0,
		scanner:     bufio.NewScanner(os.Stdin),
	}
}

func (d *Debugger) AddBreakpoint(ip int) {
	d.breakpoints[ip] = true
}

func (d *Debugger) RemoveBreakpoint(ip int) {
	delete(d.breakpoints, ip)
}

func (d *Debugger) ClearBreakpoints() {
	d.breakpoints = make(map[int]bool)
}

func (d *Debugger) SetStepMode(enabled bool) {
	d.stepMode = enabled
}

func (d *Debugger) ShouldBreak(ip int, depth int) bool {
	if d.stepMode {
		if d.stepOut {
			if depth < d.depth {
				d.stepOut = false
				d.stepMode = false
				return true
			}
			return false
		}
		if d.stepOver {
			if depth <= d.depth {
				d.stepOver = false
				d.stepMode = false
				return true
			}
			return false
		}
		return true
	}
	return d.breakpoints[ip]
}

func (d *Debugger) HandleBreak(ip int, depth int) {
	d.depth = depth
	fmt.Printf("\n[Breakpoint at IP %d, depth %d]\n", ip, depth)
	d.printStack()
	d.commandLoop()
}

func (d *Debugger) printStack() {
	fmt.Println("Stack trace:")
	for i := d.vm.GetFramesIndex() - 1; i >= 0; i-- {
		frame := d.vm.GetFrame(i)
		if frame != nil {
			fmt.Printf("  [%d] Frame\n", i)
		}
	}
}

func (d *Debugger) commandLoop() {
	for {
		fmt.Print("(gopy-debug) ")
		if !d.scanner.Scan() {
			return
		}
		line := strings.TrimSpace(d.scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		cmd := parts[0]

		switch cmd {
		case "continue", "c":
			d.stepMode = false
			d.stepOver = false
			d.stepOut = false
			return
		case "step", "s":
			d.stepMode = true
			d.stepOver = false
			d.stepOut = false
			return
		case "next", "n":
			d.stepMode = true
			d.stepOver = true
			d.stepOut = false
			return
		case "finish", "f":
			d.stepMode = true
			d.stepOver = false
			d.stepOut = true
			return
		case "break", "b":
			if len(parts) > 1 {
				ip, err := strconv.Atoi(parts[1])
				if err == nil {
					d.AddBreakpoint(ip)
					fmt.Printf("Breakpoint set at IP %d\n", ip)
				} else {
					fmt.Println("Invalid IP")
				}
			} else {
				fmt.Println("Current breakpoints:", d.breakpoints)
			}
		case "clear", "cl":
			if len(parts) > 1 {
				ip, err := strconv.Atoi(parts[1])
				if err == nil {
					d.RemoveBreakpoint(ip)
					fmt.Printf("Breakpoint cleared at IP %d\n", ip)
				} else {
					fmt.Println("Invalid IP")
				}
			} else {
				d.ClearBreakpoints()
				fmt.Println("All breakpoints cleared")
			}
		case "stack", "bt":
			d.printStack()
		case "locals", "l":
			d.printLocals()
		case "globals", "g":
			d.printGlobals()
		case "help", "h":
			d.printHelp()
		case "quit", "q":
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}

func (d *Debugger) printLocals() {
	fmt.Println("Local variables:")
	frame := d.vm.CurrentFrame()
	if frame != nil {
		for i := 0; i < d.vm.GetSP(); i++ {
			obj := d.vm.GetStack(i)
			if obj != nil {
				fmt.Printf("  stack[%d] = %s\n", i, obj.Inspect())
			}
		}
	}
}

func (d *Debugger) printGlobals() {
	fmt.Println("Global variables (first 20):")
	globals := d.vm.GetGlobals()
	count := 0
	for i, obj := range globals {
		if obj != nil && obj != objects.None_ {
			fmt.Printf("  globals[%d] = %s\n", i, obj.Inspect())
			count++
			if count >= 20 {
				break
			}
		}
	}
}

func (d *Debugger) printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  continue (c)    - Continue execution")
	fmt.Println("  step (s)        - Step into next instruction")
	fmt.Println("  next (n)        - Step over (don't enter functions)")
	fmt.Println("  finish (f)      - Step out of current function")
	fmt.Println("  break (b) [ip]  - List breakpoints or set breakpoint at IP")
	fmt.Println("  clear (cl) [ip] - Clear all breakpoints or specific one")
	fmt.Println("  stack (bt)      - Print stack trace")
	fmt.Println("  locals (l)      - Print local variables")
	fmt.Println("  globals (g)     - Print global variables")
	fmt.Println("  help (h)        - Show this help")
	fmt.Println("  quit (q)        - Exit debugger")
}

type DebugVM struct {
	*vm.VM
	debugger *Debugger
	enabled  bool
}

func NewDebugVM(bytecode *compiler.Bytecode) *DebugVM {
	baseVM := vm.New(bytecode)
	return &DebugVM{
		VM:       baseVM,
		debugger: New(baseVM),
		enabled:  false,
	}
}

func (dvm *DebugVM) EnableDebugging(enable bool) {
	dvm.enabled = enable
	dvm.debugger.vm = dvm.VM
}

func (dvm *DebugVM) AddBreakpoint(ip int) {
	dvm.debugger.AddBreakpoint(ip)
}

func (dvm *DebugVM) Run() error {
	if !dvm.enabled {
		return dvm.VM.Run()
	}
	fmt.Println("Debugger not fully implemented, running without debugging")
	return dvm.VM.Run()
}