package profiler

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-py/go-python/pkg/compiler"
)

type Profiler struct {
	enabled          bool
	startTime        time.Time
	instructionCount int
	functionCalls    map[string]int
	functionTime     map[string]time.Duration
	currentFunction  string
	timerStack       []time.Time
}

func New() *Profiler {
	return &Profiler{
		enabled:          false,
		instructionCount: 0,
		functionCalls:    make(map[string]int),
		functionTime:     make(map[string]time.Duration),
		timerStack:       []time.Time{},
	}
}

func (p *Profiler) Enable(enable bool) {
	p.enabled = enable
	if enable {
		p.startTime = time.Now()
		p.instructionCount = 0
		p.functionCalls = make(map[string]int)
		p.functionTime = make(map[string]time.Duration)
		p.timerStack = []time.Time{}
	}
}

func (p *Profiler) RecordInstruction(op compiler.Opcode) {
	if !p.enabled {
		return
	}
	p.instructionCount++
}

func (p *Profiler) EnterFunction(name string) {
	if !p.enabled {
		return
	}
	p.timerStack = append(p.timerStack, time.Now())
	p.currentFunction = name
	p.functionCalls[name]++
}

func (p *Profiler) ExitFunction() {
	if !p.enabled || len(p.timerStack) == 0 {
		return
	}
	start := p.timerStack[len(p.timerStack)-1]
	p.timerStack = p.timerStack[:len(p.timerStack)-1]
	elapsed := time.Since(start)

	if p.currentFunction != "" {
		p.functionTime[p.currentFunction] += elapsed
	}

	if len(p.timerStack) > 0 {
		p.currentFunction = ""
	}
}

func (p *Profiler) Report() string {
	if !p.enabled {
		return "Profiler is not enabled"
	}

	totalTime := time.Since(p.startTime)

	result := "\n=== Performance Profile Report ===\n"
	result += fmt.Sprintf("Total execution time: %v\n", totalTime)
	result += fmt.Sprintf("Total instructions executed: %d\n", p.instructionCount)
	result += fmt.Sprintf("Instructions per second: %.2f\n", float64(p.instructionCount)/totalTime.Seconds())
	result += "\nFunction Call Statistics:\n"
	result += "-------------------------\n"

	type funcStat struct {
		name      string
		calls     int
		totalTime time.Duration
		avgTime   time.Duration
	}

	var stats []funcStat
	for name, calls := range p.functionCalls {
		total := p.functionTime[name]
		stats = append(stats, funcStat{
			name:      name,
			calls:     calls,
			totalTime: total,
			avgTime:   total / time.Duration(calls),
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].totalTime > stats[j].totalTime
	})

	for _, stat := range stats {
		result += fmt.Sprintf("%-20s calls: %4d  total: %10v  avg: %10v\n",
			stat.name, stat.calls, stat.totalTime, stat.avgTime)
	}

	return result
}

type ProfileVM struct {
	profiler *Profiler
	innerVM  VMInterface
}

type VMInterface interface {
	Run() error
}

func NewProfileVM(vm VMInterface) *ProfileVM {
	return &ProfileVM{
		profiler: New(),
		innerVM:  vm,
	}
}

func (pvm *ProfileVM) EnableProfiling(enable bool) {
	pvm.profiler.Enable(enable)
}

func (pvm *ProfileVM) Run() error {
	if pvm.profiler.enabled {
		pvm.profiler.startTime = time.Now()
	}

	err := pvm.innerVM.Run()

	if pvm.profiler.enabled {
		fmt.Println(pvm.profiler.Report())
	}

	return err
}

type InstructionStats struct {
	opcode    compiler.Opcode
	count     int
	totalTime time.Duration
}

type DetailedProfiler struct {
	enabled           bool
	instructionStats  map[compiler.Opcode]*InstructionStats
	functionStats     map[string]*FunctionStats
	currentFunction   string
	startTime         time.Time
}

type FunctionStats struct {
	calls        int
	totalTime    time.Duration
	instructions int
}

func NewDetailedProfiler() *DetailedProfiler {
	return &DetailedProfiler{
		instructionStats: make(map[compiler.Opcode]*InstructionStats),
		functionStats:    make(map[string]*FunctionStats),
	}
}

func (dp *DetailedProfiler) Enable(enable bool) {
	dp.enabled = enable
	if enable {
		dp.startTime = time.Now()
		dp.instructionStats = make(map[compiler.Opcode]*InstructionStats)
		dp.functionStats = make(map[string]*FunctionStats)
	}
}

func (dp *DetailedProfiler) RecordInstruction(op compiler.Opcode, duration time.Duration) {
	if !dp.enabled {
		return
	}

	if _, ok := dp.instructionStats[op]; !ok {
		dp.instructionStats[op] = &InstructionStats{opcode: op}
	}
	dp.instructionStats[op].count++
	dp.instructionStats[op].totalTime += duration

	if dp.currentFunction != "" {
		if _, ok := dp.functionStats[dp.currentFunction]; !ok {
			dp.functionStats[dp.currentFunction] = &FunctionStats{}
		}
		dp.functionStats[dp.currentFunction].instructions++
	}
}

func (dp *DetailedProfiler) EnterFunction(name string) {
	if !dp.enabled {
		return
	}
	dp.currentFunction = name
	if _, ok := dp.functionStats[name]; !ok {
		dp.functionStats[name] = &FunctionStats{}
	}
	dp.functionStats[name].calls++
}

func (dp *DetailedProfiler) ExitFunction(duration time.Duration) {
	if !dp.enabled || dp.currentFunction == "" {
		return
	}
	if stats, ok := dp.functionStats[dp.currentFunction]; ok {
		stats.totalTime += duration
	}
	dp.currentFunction = ""
}

func (dp *DetailedProfiler) Report() string {
	if !dp.enabled {
		return "Profiler is not enabled"
	}

	totalTime := time.Since(dp.startTime)

	result := "\n=== Detailed Performance Profile ===\n"
	result += fmt.Sprintf("Total execution time: %v\n", totalTime)

	result += "\nInstruction Statistics:\n"
	result += "----------------------\n"

	var instStats []*InstructionStats
	for _, stats := range dp.instructionStats {
		instStats = append(instStats, stats)
	}

	sort.Slice(instStats, func(i, j int) bool {
		return instStats[i].totalTime > instStats[j].totalTime
	})

	for _, stats := range instStats {
		opName := opcodeName(stats.opcode)
		avgTime := stats.totalTime / time.Duration(stats.count)
		result += fmt.Sprintf("%-15s count: %6d  total: %10v  avg: %6v\n",
			opName, stats.count, stats.totalTime, avgTime)
	}

	result += "\nFunction Statistics:\n"
	result += "--------------------\n"

	var funcStats []*FunctionStats
	var funcNames []string
	for name, stats := range dp.functionStats {
		funcStats = append(funcStats, stats)
		funcNames = append(funcNames, name)
	}

	sort.Slice(funcStats, func(i, j int) bool {
		return funcStats[i].totalTime > funcStats[j].totalTime
	})

	for i, stats := range funcStats {
		name := funcNames[i]
		if name == "" {
			name = "(anonymous)"
		}
		avgTime := stats.totalTime / time.Duration(stats.calls)
		result += fmt.Sprintf("%-20s calls: %4d  instr: %6d  total: %10v  avg: %10v\n",
			name, stats.calls, stats.instructions, stats.totalTime, avgTime)
	}

	return result
}

func opcodeName(op compiler.Opcode) string {
	switch op {
	case compiler.OpConstant:
		return "CONSTANT"
	case compiler.OpClosure:
		return "CLOSURE"
	case compiler.OpPop:
		return "POP"
	case compiler.OpDupTop:
		return "DUP_TOP"
	case compiler.OpAdd:
		return "ADD"
	case compiler.OpSub:
		return "SUB"
	case compiler.OpMul:
		return "MUL"
	case compiler.OpDiv:
		return "DIV"
	case compiler.OpTrue:
		return "TRUE"
	case compiler.OpFalse:
		return "FALSE"
	case compiler.OpEqual:
		return "EQUAL"
	case compiler.OpNotEqual:
		return "NOT_EQUAL"
	case compiler.OpGreaterThan:
		return "GREATER"
	case compiler.OpLessThan:
		return "LESS"
	case compiler.OpMinus:
		return "MINUS"
	case compiler.OpBang:
		return "BANG"
	case compiler.OpJump:
		return "JUMP"
	case compiler.OpJumpNotTruthy:
		return "JUMP_NOT_TRUTHY"
	case compiler.OpNull:
		return "NULL"
	case compiler.OpSetGlobal:
		return "SET_GLOBAL"
	case compiler.OpGetGlobal:
		return "GET_GLOBAL"
	case compiler.OpArray:
		return "ARRAY"
	case compiler.OpHash:
		return "HASH"
	case compiler.OpSet:
		return "SET"
	case compiler.OpCall:
		return "CALL"
	case compiler.OpReturn:
		return "RETURN"
	case compiler.OpReturnValue:
		return "RETURN_VAL"
	case compiler.OpGetLocal:
		return "GET_LOCAL"
	case compiler.OpSetLocal:
		return "SET_LOCAL"
	case compiler.OpGetFree:
		return "GET_FREE"
	case compiler.OpBeginTry:
		return "BEGIN_TRY"
	case compiler.OpEndTry:
		return "END_TRY"
	case compiler.OpRaise:
		return "RAISE"
	case compiler.OpExceptHandler:
		return "EXCEPT_HANDLER"
	case compiler.OpFinally:
		return "FINALLY"
	case compiler.OpYield:
		return "YIELD"
	case compiler.OpYieldValue:
		return "YIELD_VALUE"
	case compiler.OpEnterContext:
		return "ENTER_CONTEXT"
	case compiler.OpExitContext:
		return "EXIT_CONTEXT"
	case compiler.OpMakeGenerator:
		return "MAKE_GENERATOR"
	case compiler.OpCreateClass:
		return "CREATE_CLASS"
	case compiler.OpCreateClassWithSuper:
		return "CREATE_CLASS_SUPER"
	case compiler.OpGetAttribute:
		return "GET_ATTR"
	case compiler.OpSetAttribute:
		return "SET_ATTR"
	case compiler.OpFormatString:
		return "FORMAT_STRING"
	case compiler.OpIndex:
		return "INDEX"
	case compiler.OpSlice:
		return "SLICE"
	default:
		return fmt.Sprintf("OP_%d", op)
	}
}