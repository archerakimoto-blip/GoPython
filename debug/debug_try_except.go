package main

import (
	"fmt"
	"strings"

	"github.com/go-py/go-python/pkg/compiler"
	"github.com/go-py/go-python/pkg/desugar"
	"github.com/go-py/go-python/pkg/lexer"
	"github.com/go-py/go-python/pkg/objects"
	"github.com/go-py/go-python/pkg/parser"
)

var opcodeNames = map[compiler.Opcode]string{
	compiler.OpConstant:        "OpConstant",
	compiler.OpPop:             "OpPop",
	compiler.OpDupTop:          "OpDupTop",
	compiler.OpAdd:             "OpAdd",
	compiler.OpSub:             "OpSub",
	compiler.OpMul:             "OpMul",
	compiler.OpDiv:             "OpDiv",
	compiler.OpTrue:            "OpTrue",
	compiler.OpFalse:           "OpFalse",
	compiler.OpEqual:           "OpEqual",
	compiler.OpNotEqual:        "OpNotEqual",
	compiler.OpGreaterThan:     "OpGreaterThan",
	compiler.OpLessThan:        "OpLessThan",
	compiler.OpMinus:           "OpMinus",
	compiler.OpBang:            "OpBang",
	compiler.OpJump:            "OpJump",
	compiler.OpJumpNotTruthy:   "OpJumpNotTruthy",
	compiler.OpNull:            "OpNull",
	compiler.OpGetGlobal:       "OpGetGlobal",
	compiler.OpSetGlobal:       "OpSetGlobal",
	compiler.OpArray:           "OpArray",
	compiler.OpHash:            "OpHash",
	compiler.OpSet:             "OpSet",
	compiler.OpIndex:           "OpIndex",
	compiler.OpSlice:           "OpSlice",
	compiler.OpCall:            "OpCall",
	compiler.OpReturnValue:     "OpReturnValue",
	compiler.OpReturn:          "OpReturn",
	compiler.OpGetLocal:        "OpGetLocal",
	compiler.OpSetLocal:        "OpSetLocal",
	compiler.OpBeginTry:        "OpBeginTry",
	compiler.OpEndTry:          "OpEndTry",
	compiler.OpRaise:           "OpRaise",
	compiler.OpExceptHandler:   "OpExceptHandler",
	compiler.OpFinally:         "OpFinally",
	compiler.OpYield:           "OpYield",
	compiler.OpEnterContext:    "OpEnterContext",
	compiler.OpExitContext:     "OpExitContext",
	compiler.OpMakeGenerator:   "OpMakeGenerator",
	compiler.OpYieldValue:      "OpYieldValue",
}

var operandWidths = map[compiler.Opcode][]int{
	compiler.OpConstant:        {2},
	compiler.OpPop:             {},
	compiler.OpDupTop:          {},
	compiler.OpAdd:             {},
	compiler.OpSub:             {},
	compiler.OpMul:             {},
	compiler.OpDiv:             {},
	compiler.OpTrue:            {},
	compiler.OpFalse:           {},
	compiler.OpEqual:           {},
	compiler.OpNotEqual:        {},
	compiler.OpGreaterThan:     {},
	compiler.OpLessThan:        {},
	compiler.OpMinus:           {},
	compiler.OpBang:            {},
	compiler.OpJump:            {2},
	compiler.OpJumpNotTruthy:   {2},
	compiler.OpNull:            {},
	compiler.OpGetGlobal:       {2},
	compiler.OpSetGlobal:       {2},
	compiler.OpArray:           {2},
	compiler.OpHash:            {2},
	compiler.OpSet:             {2},
	compiler.OpIndex:           {},
	compiler.OpSlice:           {},
	compiler.OpCall:            {1},
	compiler.OpReturnValue:     {},
	compiler.OpReturn:          {},
	compiler.OpGetLocal:        {1},
	compiler.OpSetLocal:        {1},
	compiler.OpBeginTry:        {2, 2},
	compiler.OpEndTry:          {},
	compiler.OpRaise:           {},
	compiler.OpExceptHandler:   {2, 2},
	compiler.OpFinally:         {},
	compiler.OpYield:           {},
	compiler.OpEnterContext:    {},
	compiler.OpExitContext:     {},
	compiler.OpMakeGenerator:   {},
	compiler.OpYieldValue:      {},
}

func main() {
	code := `try:
x = 1 / 0
except:
print("error")
`

	sep := strings.Repeat("=", 71)
	fmt.Println(sep)
	fmt.Println("Python 代码:")
	fmt.Println(sep)
	fmt.Println(code)

	fmt.Println(sep)
	fmt.Println("1. 词法分析 (Lexer)")
	fmt.Println(sep)
	l := lexer.New(code)
	tokens := []string{}
	for {
		token := l.NextToken()
		tokens = append(tokens, fmt.Sprintf("%s:%s", token.Type, token.Literal))
		if token.Type == lexer.EOF {
			break
		}
	}
	for i, t := range tokens {
		fmt.Printf("%d: %s\n", i, t)
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Println("2. 语法分析 (Parser)")
	fmt.Println(sep)
	l2 := lexer.New(code)
	p := parser.New(l2)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
	}

	fmt.Println("成功解析 AST")
	fmt.Printf("Statements count: %d\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Println("3. 脱糖 (Desugar)")
	fmt.Println(sep)
	program = desugar.Desugar(program)
	fmt.Println("脱糖后的 Statements:")
	for i, stmt := range program.Statements {
		fmt.Printf("  %d: %T\n", i, stmt)
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Println("4. 编译 (Compiler)")
	fmt.Println(sep)
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("Compilation error: %s\n", err)
		return
	}

	bytecode := comp.Bytecode()
	instructions := bytecode.Instructions

	fmt.Printf("编译成功!\n")
	fmt.Printf("指令长度: %d 字节\n", len(instructions))
	fmt.Printf("常量池大小: %d\n", len(bytecode.Constants))

	fmt.Println()
	fmt.Println(sep)
	fmt.Println("5. 字节码详情")
	fmt.Println(sep)
	fmt.Println("格式: 位置  操作码名称 [操作数]")
	fmt.Println()

	printBytecode(instructions)

	fmt.Println()
	fmt.Println(sep)
	fmt.Println("6. 常量池详情")
	fmt.Println(sep)
	for i, c := range bytecode.Constants {
		fmt.Printf("Index %d: Type=%s, Value=%s\n", i, c.Type(), c.Inspect())
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Println("7. 异常处理流程分析")
	fmt.Println(sep)
	analyzeTryExcept(instructions, bytecode.Constants)

	fmt.Println()
	fmt.Println(sep)
	fmt.Println("8. 问题分析")
	fmt.Println(sep)
	analyzeProblem(instructions, bytecode.Constants)
}

func printBytecode(instructions compiler.Instructions) {
	i := 0
	for i < len(instructions) {
		pos := i
		op := compiler.Opcode(instructions[i])
		name := opcodeNames[op]
		if name == "" {
			name = fmt.Sprintf("UNKNOWN(0x%02x)", op)
		}

		fmt.Printf("%04d: %-20s", pos, name)

		widths, ok := operandWidths[op]
		if !ok {
			fmt.Println()
			i++
			continue
		}

		operands := []string{}
		offset := 1
		for j, width := range widths {
			if offset+width > len(instructions)-i {
				break
			}

			var val int
			switch width {
			case 1:
				val = int(instructions[i+offset])
			case 2:
				val = int(uint16(instructions[i+offset])<<8 | uint16(instructions[i+offset+1]))
			case 4:
				if j == 0 {
					val = int(uint16(instructions[i+offset])<<8 | uint16(instructions[i+offset+1]))
				} else if j == 1 {
					val = int(uint16(instructions[i+offset+2])<<8 | uint16(instructions[i+offset+3]))
				}
			default:
				val = int(instructions[i+offset])
			}
			operands = append(operands, fmt.Sprintf("%d", val))
			offset += width
		}

		if len(operands) > 0 {
			fmt.Printf("[%s]", joinStrings(operands, ", "))
		}
		fmt.Println()

		step := 1
		for _, w := range widths {
			step += w
		}
		i += step
	}
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func analyzeTryExcept(instructions compiler.Instructions, constants []objects.Object) {
	i := 0
	inTryBlock := false
	tryStart := -1
	handlerPositions := []int{}

	for i < len(instructions) {
		op := compiler.Opcode(instructions[i])
		widths, ok := operandWidths[op]
		if !ok {
			i++
			continue
		}

		switch op {
		case compiler.OpBeginTry:
			inTryBlock = true
			tryStart = i
			exceptCount := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
			hasFinally := int(uint16(instructions[i+3])<<8 | uint16(instructions[i+4]))
			fmt.Printf("发现 OpBeginTry @ %d: exceptCount=%d, hasFinally=%d\n", i, exceptCount, hasFinally)
			fmt.Printf("  -> try 块开始于 %d\n", i+5)

		case compiler.OpExceptHandler:
			handlerPositions = append(handlerPositions, i)
			typeIdx := int(uint16(instructions[i+1])<<8 | uint16(instructions[i+2]))
			varIdx := int(uint16(instructions[i+3])<<8 | uint16(instructions[i+4]))

			var typeName string
			if typeIdx > 0 && typeIdx < len(constants) {
				if obj, ok := constants[typeIdx].(*objects.String); ok {
					typeName = obj.Value
				}
			}

			var varName string
			if varIdx > 0 && varIdx < len(constants) {
				if obj, ok := constants[varIdx].(*objects.String); ok {
					varName = obj.Value
				}
			}

			fmt.Printf("发现 OpExceptHandler @ %d: typeIdx=%d (type='%s'), varIdx=%d (var='%s')\n", i, typeIdx, typeName, varIdx, varName)
			fmt.Printf("  -> 异常处理开始于 %d\n", i+5)

		case compiler.OpRaise:
			fmt.Printf("发现 OpRaise @ %d: 将触发异常处理查找\n", i)
			fmt.Println("  -> 查找逻辑: 从当前 IP 向后查找 OpExceptHandler")

		case compiler.OpDiv:
			if inTryBlock {
				fmt.Printf("发现 OpDiv @ %d: 可能产生除零异常\n", i)
			}
		}

		step := 1
		for _, w := range widths {
			step += w
		}
		i += step
	}

	fmt.Printf("\n总结:\n")
	fmt.Printf("  - try 块从 %d 开始\n", tryStart)
	fmt.Printf("  - 找到 %d 个异常处理器\n", len(handlerPositions))
	for i, pos := range handlerPositions {
		fmt.Printf("    Handler %d @ %d\n", i+1, pos)
	}
}

func analyzeProblem(instructions compiler.Instructions, constants []objects.Object) {
	fmt.Println("问题分析:")
	fmt.Println()

	fmt.Println("1. OpBeginTry 和 OpExceptHandler 的位置关系:")
	i := 0
	var beginTryPos, exceptHandlerPos int = -1, -1

	for i < len(instructions) {
		op := compiler.Opcode(instructions[i])
		if op == compiler.OpBeginTry {
			beginTryPos = i
		}
		if op == compiler.OpExceptHandler {
			exceptHandlerPos = i
		}
		step := 1
		for _, w := range operandWidths[op] {
			step += w
		}
		i += step
	}

	if beginTryPos >= 0 && exceptHandlerPos >= 0 {
		fmt.Printf("   - OpBeginTry 位置: %d\n", beginTryPos)
		fmt.Printf("   - OpExceptHandler 位置: %d\n", exceptHandlerPos)
		fmt.Printf("   - 相对位置: OpExceptHandler 在 OpBeginTry 之后 %d 字节\n", exceptHandlerPos-beginTryPos)
	}

	fmt.Println()
	fmt.Println("2. OpExceptHandler 的字节布局:")
	if exceptHandlerPos >= 0 {
		fmt.Printf("   - 字节 %d: 操作码 OpExceptHandler (值=%d)\n", exceptHandlerPos, compiler.OpExceptHandler)
		fmt.Printf("   - 字节 %d: exceptionType index 低字节\n", exceptHandlerPos+1)
		fmt.Printf("   - 字节 %d: exceptionType index 高字节\n", exceptHandlerPos+2)
		fmt.Printf("   - 字节 %d: varName index 低字节\n", exceptHandlerPos+3)
		fmt.Printf("   - 字节 %d: varName index 高字节\n", exceptHandlerPos+4)
		fmt.Printf("   - 字节 %d: 异常处理代码开始\n", exceptHandlerPos+5)
	}

	fmt.Println()
	fmt.Println("3. 虚拟机中的异常处理查找逻辑:")
	fmt.Println("   当 OpRaise 执行时:")
	fmt.Println("   a) 异常被抛出")
	fmt.Println("   b) 虚拟机从 OpRaise 的位置向前查找 OpExceptHandler")
	fmt.Println("   c) 如果找到匹配的异常类型，跳转到处理代码")
	fmt.Println()
	fmt.Println("   问题: OpExceptHandler 在 OpRaise 之后！")
	fmt.Println("   当 OpRaise 执行时，查找方向是向后而不是向前，导致找不到处理器。")

	fmt.Println()
	fmt.Println("4. 解决方案建议:")
	fmt.Println("   - 方案1: 修改虚拟机的查找逻辑，让它从 OpBeginTry 的位置开始查找")
	fmt.Println("   - 方案2: 在编译时，在 OpRaise 之前插入一个跳转指令，跳转到异常处理器")
	fmt.Println("   - 方案3: 使用 exceptionStack 来跟踪 try 块的起始位置，而不是依赖字节码顺序")
}
