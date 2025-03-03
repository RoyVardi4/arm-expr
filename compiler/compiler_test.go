package compiler_test

import (
	"math"
	"testing"

	"expr/internal/testify/assert"
	"expr/internal/testify/require"

	"expr"
	"expr/vm"
	"expr/vm/runtime"
)

type B struct {
	_ byte
	_ byte
	C struct {
		_ byte
		_ byte
		_ byte
		D int
	}
}

func (B) FuncInB() int {
	return 0
}

type Env struct {
	A struct {
		_   byte
		B   B
		Map map[string]B
		Ptr *int
	}
}

// AFunc is a method what goes before Func in the alphabet.
func (e Env) AFunc() int {
	return 0
}

func (e Env) Func() B {
	return B{}
}

func TestCompile(t *testing.T) {
	var tests = []struct {
		code string
		want vm.Program
	}{
		{
			`65535`,
			vm.Program{
				Constants: []any{
					math.MaxUint16,
				},
				Bytecode: []vm.Opcode{
					vm.OpPush,
				},
				Arguments: []int{0},
			},
		},
		{
			`.5`,
			vm.Program{
				Constants: []any{
					.5,
				},
				Bytecode: []vm.Opcode{
					vm.OpPush,
				},
				Arguments: []int{0},
			},
		},
		{
			`true`,
			vm.Program{
				Bytecode: []vm.Opcode{
					vm.OpTrue,
				},
				Arguments: []int{0},
			},
		},
		{
			`"string"`,
			vm.Program{
				Constants: []any{
					"string",
				},
				Bytecode: []vm.Opcode{
					vm.OpPush,
				},
				Arguments: []int{0},
			},
		},
		{
			`A.B.C.D`,
			vm.Program{
				Constants: []any{
					&runtime.Field{
						Index: []int{0, 1, 2, 3},
						Path:  []string{"A", "B", "C", "D"},
					},
				},
				Bytecode: []vm.Opcode{
					vm.OpLoadField,
				},
				Arguments: []int{0},
			},
		},
		{
			`A.Map["B"].C.D`,
			vm.Program{
				Constants: []any{
					&runtime.Field{
						Index: []int{0, 2},
						Path:  []string{"A", "Map"},
					},
					"B",
					&runtime.Field{
						Index: []int{2, 3},
						Path:  []string{"C", "D"},
					},
				},
				Bytecode: []vm.Opcode{
					vm.OpLoadField,
					vm.OpPush,
					vm.OpFetch,
					vm.OpFetchField,
				},
				Arguments: []int{0, 1, 0, 2},
			},
		},
		{
			`Func()`,
			vm.Program{
				Constants: []any{
					&runtime.Method{
						Index: 1,
						Name:  "Func",
					},
				},
				Bytecode: []vm.Opcode{
					vm.OpLoadMethod,
					vm.OpCall,
				},
				Arguments: []int{0, 0},
			},
		},
		{
			`Func().FuncInB()`,
			vm.Program{
				Constants: []any{
					&runtime.Method{
						Index: 1,
						Name:  "Func",
					},
					&runtime.Method{
						Index: 0,
						Name:  "FuncInB",
					},
				},
				Bytecode: []vm.Opcode{
					vm.OpLoadMethod,
					vm.OpCall,
					vm.OpMethod,
					vm.OpCallTyped,
				},
				Arguments: []int{0, 0, 1, 12},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			program, err := expr.Compile(test.code, expr.Env(Env{}))
			require.NoError(t, err)

			assert.Equal(t, test.want.Disassemble(), program.Disassemble())
		})
	}
}

func TestCompile_FuncTypes(t *testing.T) {
	env := map[string]any{
		"fn": func(string, string) string {
			return "foo"
		},
	}
	program, err := expr.Compile("fn('roy', 'bar')", expr.Env(env))
	require.NoError(t, err)
	require.Equal(t, vm.OpCallTyped, program.Bytecode[3])
	require.Equal(t, 82, program.Arguments[3])
}

func TestCompile_call_on_nil(t *testing.T) {
	env := map[string]any{
		"foo": nil,
	}
	_, err := expr.Compile(`foo()`, expr.Env(env))
	require.Error(t, err)
	require.Contains(t, err.Error(), "foo is nil; cannot call nil as function")
}
