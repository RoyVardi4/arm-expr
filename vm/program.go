package vm

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"text/tabwriter"

	"expr/ast"
	"expr/file"
	"expr/vm/runtime"
)

// Program represents a compiled expression.
type Program struct {
	Bytecode  []Opcode
	Arguments []int
	Constants []any

	source    file.Source
	node      ast.Node
	locations []file.Location
	functions []Function
	debugInfo map[string]string
	span      *Span
}

// NewProgram returns a new Program. It's used by the compiler.
func NewProgram(
	source file.Source,
	node ast.Node,
	locations []file.Location,
	constants []any,
	bytecode []Opcode,
	arguments []int,
	functions []Function,
	debugInfo map[string]string,
	span *Span,
) *Program {
	return &Program{
		source:    source,
		node:      node,
		locations: locations,
		Constants: constants,
		Bytecode:  bytecode,
		Arguments: arguments,
		functions: functions,
		debugInfo: debugInfo,
		span:      span,
	}
}

// Source returns origin file.Source.
func (program *Program) Source() file.Source {
	return program.source
}

// Node returns origin ast.Node.
func (program *Program) Node() ast.Node {
	return program.node
}

// Locations returns a slice of bytecode's locations.
func (program *Program) Locations() []file.Location {
	return program.locations
}

// Disassemble returns opcodes as a string.
func (program *Program) Disassemble() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	program.DisassembleWriter(w)
	_ = w.Flush()
	return buf.String()
}

// DisassembleWriter takes a writer and writes opcodes to it.
func (program *Program) DisassembleWriter(w io.Writer) {
	ip := 0
	for ip < len(program.Bytecode) {
		pp := ip
		op := program.Bytecode[ip]
		arg := program.Arguments[ip]
		ip += 1

		code := func(label string) {
			_, _ = fmt.Fprintf(w, "%v\t%v\n", pp, label)
		}
		argument := func(label string) {
			_, _ = fmt.Fprintf(w, "%v\t%v\t<%v>\n", pp, label, arg)
		}
		constant := func(label string) {
			var c any
			if arg < len(program.Constants) {
				c = program.Constants[arg]
			} else {
				c = "out of range"
			}
			if r, ok := c.(*regexp.Regexp); ok {
				c = r.String()
			}
			if field, ok := c.(*runtime.Field); ok {
				c = fmt.Sprintf("{%v %v}", strings.Join(field.Path, "."), field.Index)
			}
			if method, ok := c.(*runtime.Method); ok {
				c = fmt.Sprintf("{%v %v}", method.Name, method.Index)
			}
			_, _ = fmt.Fprintf(w, "%v\t%v\t<%v>\t%v\n", pp, label, arg, c)
		}

		switch op {
		case OpInvalid:
			code("OpInvalid")

		case OpPush:
			constant("OpPush")

		case OpLoadConst:
			constant("OpLoadConst")

		case OpLoadField:
			constant("OpLoadField")

		case OpLoadFast:
			constant("OpLoadFast")

		case OpLoadMethod:
			constant("OpLoadMethod")

		case OpFetch:
			code("OpFetch")

		case OpFetchField:
			constant("OpFetchField")

		case OpMethod:
			constant("OpMethod")

		case OpTrue:
			code("OpTrue")

		case OpFalse:
			code("OpFalse")

		case OpNil:
			code("OpNil")

		case OpCall:
			argument("OpCall")

		case OpCallFast:
			argument("OpCallFast")

		case OpCallTyped:
			signature := reflect.TypeOf(FuncTypes[arg]).Elem().String()
			_, _ = fmt.Fprintf(w, "%v\t%v\t<%v>\t%v\n", pp, "OpCallTyped", arg, signature)

		case OpCast:
			argument("OpCast")

		case OpDeref:
			code("OpDeref")

		case OpProfileStart:
			code("OpProfileStart")

		case OpProfileEnd:
			code("OpProfileEnd")

		case OpBegin:
			code("OpBegin")

		case OpEnd:
			code("OpEnd")

		default:
			_, _ = fmt.Fprintf(w, "%v\t%#x (unknown)\n", ip, op)
		}
	}
}
