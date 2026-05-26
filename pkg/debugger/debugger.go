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
		fmt.Printf("  [%d] IP: %d\n", i, frame.ip)
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
		base := frame.basePointer
		for i := base; i < d.vm.GetSP(); i++ {
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
	return &DebugVM{
		VM:       vm.New(bytecode),
		debugger: New(vm.New(bytecode)),
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

	var ip int
	var ins compiler.Instructions
	var op compiler.Opcode

	for dvm.currentFrame().ip < len(dvm.currentFrame().fn.Instructions)-1 {
		dvm.currentFrame().ip++
		ip = dvm.currentFrame().ip
		ins = dvm.currentFrame().fn.Instructions
		op = compiler.Opcode(ins[ip])

		depth := dvm.framesIndex

		if dvm.debugger.ShouldBreak(ip, depth) {
			dvm.debugger.HandleBreak(ip, depth)
		}

		switch op {
		case compiler.OpConstant:
			constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			err := dvm.push(dvm.constants[constIndex])
			if err != nil {
				return err
			}

		case compiler.OpClosure:
			constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			numFree := int(ins[ip+3])
			dvm.currentFrame().ip += 3

			fn, ok := dvm.constants[constIndex].(*compiler.CompiledFunction)
			if !ok {
				return fmt.Errorf("not a function: %T", dvm.constants[constIndex])
			}

			free := make([]objects.Object, numFree)
			for i := numFree - 1; i >= 0; i-- {
				free[i] = dvm.pop()
			}

			closure := &objects.Closure{
				Instructions:  fn.Instructions,
				NumLocals:     fn.NumLocals,
				NumParameters: fn.NumParameters,
				IsGenerator:   fn.IsGenerator,
				Free:          free,
			}

			err := dvm.push(closure)
			if err != nil {
				return err
			}

		case compiler.OpPop:
			dvm.pop()

		case compiler.OpDupTop:
			if dvm.sp > 0 {
				top := dvm.stack[dvm.sp-1]
				err := dvm.push(top)
				if err != nil {
					return err
				}
			}

		case compiler.OpAdd, compiler.OpSub, compiler.OpMul, compiler.OpDiv:
			err := dvm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
			if dvm.sp > 0 && dvm.stack[dvm.sp-1].Type() == objects.ERROR_OBJ {
				errObj := dvm.stack[dvm.sp-1]
				caught := dvm.raiseException(errObj)
				if !caught {
					return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
				}
			}

		case compiler.OpTrue:
			err := dvm.push(objects.True)
			if err != nil {
				return err
			}

		case compiler.OpFalse:
			err := dvm.push(objects.False)
			if err != nil {
				return err
			}

		case compiler.OpEqual, compiler.OpNotEqual, compiler.OpGreaterThan, compiler.OpLessThan:
			err := dvm.executeComparison(op)
			if err != nil {
				return err
			}

		case compiler.OpMinus:
			err := dvm.executeMinusOperator()
			if err != nil {
				return err
			}

		case compiler.OpBang:
			err := dvm.executeBangOperator()
			if err != nil {
				return err
			}

		case compiler.OpJump:
			pos := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip = pos - 1

		case compiler.OpJumpNotTruthy:
			pos := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2

			condition := dvm.pop()
			if !isTruthy(condition) {
				dvm.currentFrame().ip = pos - 1
			}

		case compiler.OpNull:
			err := dvm.push(objects.None_)
			if err != nil {
				return err
			}

		case compiler.OpSetGlobal:
			globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			dvm.globals[globalIndex] = dvm.pop()

		case compiler.OpGetGlobal:
			globalIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			err := dvm.push(dvm.globals[globalIndex])
			if err != nil {
				return err
			}

		case compiler.OpArray:
			numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2

			array := dvm.buildArray(dvm.sp-numElements, dvm.sp)
			dvm.sp = dvm.sp - numElements

			err := dvm.push(array)
			if err != nil {
				return err
			}

		case compiler.OpHash:
			numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2

			hash, err := dvm.buildHash(dvm.sp-numElements, dvm.sp)
			if err != nil {
				return err
			}
			dvm.sp = dvm.sp - numElements

			err = dvm.push(hash)
			if err != nil {
				return err
			}

		case compiler.OpSet:
			numElements := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2

			set := dvm.buildSet(dvm.sp-numElements, dvm.sp)
			dvm.sp = dvm.sp - numElements

			err := dvm.push(set)
			if err != nil {
				return err
			}

		case compiler.OpCall:
			numArgs := int(ins[ip+1])
			dvm.currentFrame().ip += 1
			err := dvm.executeCall(numArgs)
			if err != nil {
				return err
			}

		case compiler.OpReturn:
			returnValue := dvm.pop()
			if dvm.framesIndex == 1 {
				dvm.stack[0] = returnValue
				dvm.sp = 1
				return nil
			}

			dvm.popFrame()
			dvm.sp = dvm.currentFrame().basePointer - 1
			err := dvm.push(returnValue)
			if err != nil {
				return err
			}

		case compiler.OpGetLocal:
			localIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			err := dvm.push(dvm.stack[dvm.currentFrame().basePointer+localIndex])
			if err != nil {
				return err
			}

		case compiler.OpSetLocal:
			localIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			dvm.stack[dvm.currentFrame().basePointer+localIndex] = dvm.pop()

		case compiler.OpGetAttr:
			attrNameIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			attrName, ok := dvm.constants[attrNameIndex].(*objects.String)
			if !ok {
				return fmt.Errorf("expected string for attribute name")
			}
			err := dvm.executeGetAttr(attrName.Value)
			if err != nil {
				return err
			}

		case compiler.OpSetAttr:
			attrNameIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			attrName, ok := dvm.constants[attrNameIndex].(*objects.String)
			if !ok {
				return fmt.Errorf("expected string for attribute name")
			}
			err := dvm.executeSetAttr(attrName.Value)
			if err != nil {
				return err
			}

		case compiler.OpIndex:
			err := dvm.executeIndex()
			if err != nil {
				return err
			}

		case compiler.OpSlice:
			err := dvm.executeSlice()
			if err != nil {
				return err
			}

		case compiler.OpPrint:
			err := dvm.executePrint()
			if err != nil {
				return err
			}

		case compiler.OpGetBuiltin:
			builtinIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2
			err := dvm.push(dvm.constants[builtinIndex])
			if err != nil {
				return err
			}

		case compiler.OpMakeFunction:
			constIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			numFree := int(ins[ip+3])
			dvm.currentFrame().ip += 3

			fn, ok := dvm.constants[constIndex].(*compiler.CompiledFunction)
			if !ok {
				return fmt.Errorf("not a function: %T", dvm.constants[constIndex])
			}

			free := make([]objects.Object, numFree)
			for i := numFree - 1; i >= 0; i-- {
				free[i] = dvm.pop()
			}

			closure := &objects.Closure{
				Instructions:  fn.Instructions,
				NumLocals:     fn.NumLocals,
				NumParameters: fn.NumParameters,
				IsGenerator:   fn.IsGenerator,
				Free:          free,
			}

			err := dvm.push(closure)
			if err != nil {
				return err
			}

		case compiler.OpTry:
			handlerIP := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			finallyIP := int(uint16(ins[ip+3])<<8 | uint16(ins[ip+4]))
			dvm.currentFrame().ip += 4

			dvm.exceptionStack = append(dvm.exceptionStack, vm.ExceptionHandler{
				handlerIP: handlerIP,
				stackPtr:  dvm.sp,
			})

		case compiler.OpExcept:
			exceptionTypeIndex := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			varNameIndex := int(uint16(ins[ip+3])<<8 | uint16(ins[ip+4]))
			finallyIP := int(uint16(ins[ip+5])<<8 | uint16(ins[ip+6]))
			dvm.currentFrame().ip += 6

			exceptionType := ""
			if exceptionTypeIndex != 0 {
				exceptionTypeObj, ok := dvm.constants[exceptionTypeIndex].(*objects.String)
				if ok {
					exceptionType = exceptionTypeObj.Value
				}
			}

			var varName string
			if varNameIndex != 0 {
				varNameObj, ok := dvm.constants[varNameIndex].(*objects.String)
				if ok {
					varName = varNameObj.Value
				}
			}

			handler := dvm.exceptionStack[len(dvm.exceptionStack)-1]
			handler.exceptionType = exceptionType
			handler.varName = varName
			handler.finallyStartIP = finallyIP
			dvm.exceptionStack[len(dvm.exceptionStack)-1] = handler

		case compiler.OpRaise:
			errObj := dvm.pop()
			caught := dvm.raiseException(errObj)
			if !caught {
				return fmt.Errorf("unhandled exception: %s", errObj.Inspect())
			}

		case compiler.OpGetIter:
			err := dvm.executeGetIter()
			if err != nil {
				return err
			}

		case compiler.OpForIter:
			jumpPos := int(uint16(ins[ip+1])<<8 | uint16(ins[ip+2]))
			dvm.currentFrame().ip += 2

			err := dvm.executeForIter(jumpPos)
			if err != nil {
				return err
			}

		case compiler.OpNot:
			err := dvm.executeNotOperator()
			if err != nil {
				return err
			}

		case compiler.OpAnd:
			err := dvm.executeAndOperator()
			if err != nil {
				return err
			}

		case compiler.OpOr:
			err := dvm.executeOrOperator()
			if err != nil {
				return err
			}

		case compiler.OpYield:
			err := dvm.executeYield()
			if err != nil {
				return err
			}

		case compiler.OpSend:
			err := dvm.executeSend()
			if err != nil {
				return err
			}

		case compiler.OpListAppend:
			err := dvm.executeListAppend()
			if err != nil {
				return err
			}

		case compiler.OpDictSetItem:
			err := dvm.executeDictSetItem()
			if err != nil {
				return err
			}

		case compiler.OpGetItem:
			err := dvm.executeGetItem()
			if err != nil {
				return err
			}

		case compiler.OpSetItem:
			err := dvm.executeSetItem()
			if err != nil {
				return err
			}

		case compiler.OpDeleteItem:
			err := dvm.executeDeleteItem()
			if err != nil {
				return err
			}

		case compiler.OpCompare:
			compareOp := compiler.Opcode(ins[ip+1])
			dvm.currentFrame().ip += 1
			err := dvm.executeCompare(compareOp)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown opcode: %d", op)
		}
	}

	return nil
}

func isTruthy(obj objects.Object) bool {
	if obj == objects.None_ {
		return false
	}
	if obj == objects.False {
		return false
	}
	if obj == objects.True {
		return true
	}
	switch o := obj.(type) {
	case *objects.Integer:
		return o.Value != 0
	case *objects.Float:
		return o.Value != 0
	case *objects.String:
		return len(o.Value) > 0
	case *objects.List:
		return len(o.Elements) > 0
	case *objects.Dict:
		return len(o.Pairs) > 0
	default:
		return true
	}
}
