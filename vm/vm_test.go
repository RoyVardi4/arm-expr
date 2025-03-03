package vm_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"expr/internal/testify/require"

	"expr"
	"expr/checker"
	"expr/compiler"
	"expr/conf"
	"expr/parser"
	"expr/vm"
)

func TestRun_NilProgram(t *testing.T) {
	_, err := vm.Run(nil, nil)
	require.Error(t, err)
}

func TestRun_ReuseVM(t *testing.T) {
	node, err := parser.Parse(`1`)
	require.NoError(t, err)

	program, err := compiler.Compile(node, nil)
	require.NoError(t, err)

	reuse := vm.VM{}
	_, err = reuse.Run(program, nil)
	require.NoError(t, err)
	_, err = reuse.Run(program, nil)
	require.NoError(t, err)
}

func TestRun_ReuseVM_for_different_variables(t *testing.T) {
	v := vm.VM{}

	program, err := expr.Compile(`1`)
	require.NoError(t, err)
	out, err := v.Run(program, nil)
	require.NoError(t, err)
	require.Equal(t, 1, out)

	program, err = expr.Compile(`2`)
	require.NoError(t, err)
	out, err = v.Run(program, nil)
	require.NoError(t, err)
	require.Equal(t, 2, out)

	program, err = expr.Compile(`3`)
	require.NoError(t, err)
	out, err = v.Run(program, nil)
	require.NoError(t, err)
	require.Equal(t, 3, out)
}

func TestRun_Cast(t *testing.T) {
	input := `1`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	program, err := compiler.Compile(tree, &conf.Config{Expect: reflect.Float64})
	require.NoError(t, err)

	out, err := vm.Run(program, nil)
	require.NoError(t, err)

	require.Equal(t, float64(1), out)
}

type ErrorEnv struct {
	InnerEnv InnerEnv
}
type InnerEnv struct{}

func (ErrorEnv) WillError(param string) (bool, error) {
	if param == "yes" {
		return false, errors.New("error")
	}
	return true, nil
}

func (InnerEnv) WillError(param string) (bool, error) {
	if param == "yes" {
		return false, errors.New("inner error")
	}
	return true, nil
}

func TestRun_MethodWithError(t *testing.T) {
	input := `WillError("yes")`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	env := ErrorEnv{}
	funcConf := conf.New(env)
	_, err = checker.Check(tree, funcConf)
	require.NoError(t, err)

	program, err := compiler.Compile(tree, funcConf)
	require.NoError(t, err)

	out, err := vm.Run(program, env)
	require.EqualError(t, err, "error (1:1)\n | WillError(\"yes\")\n | ^")
	require.Equal(t, nil, out)

	selfErr := errors.Unwrap(err)
	require.NotNil(t, err)
	require.Equal(t, "error", selfErr.Error())
}

func TestRun_FastMethods(t *testing.T) {
	input := `hello()`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	env := map[string]any{
		"hello": func(...any) any { return "hello" },
	}
	funcConf := conf.New(env)
	_, err = checker.Check(tree, funcConf)
	require.NoError(t, err)

	program, err := compiler.Compile(tree, funcConf)
	require.NoError(t, err)

	out, err := vm.Run(program, env)
	require.NoError(t, err)

	require.Equal(t, "hello", out)
}

func TestRun_InnerMethodWithError(t *testing.T) {
	input := `InnerEnv.WillError("yes")`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	env := ErrorEnv{}
	funcConf := conf.New(env)
	program, err := compiler.Compile(tree, funcConf)
	require.NoError(t, err)

	out, err := vm.Run(program, env)
	require.EqualError(t, err, "inner error (1:10)\n | InnerEnv.WillError(\"yes\")\n | .........^")
	require.Equal(t, nil, out)
}

func TestRun_InnerMethodWithError_NilSafe(t *testing.T) {
	input := `InnerEnv.WillError("yes")`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	env := ErrorEnv{}
	funcConf := conf.New(env)
	program, err := compiler.Compile(tree, funcConf)
	require.NoError(t, err)

	out, err := vm.Run(program, env)
	require.EqualError(t, err, "inner error (1:10)\n | InnerEnv.WillError(\"yes\")\n | .........^")
	require.Equal(t, nil, out)
}

func TestRun_TaggedFieldName(t *testing.T) {
	input := `value`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	env := struct {
		V string `expr:"value"`
	}{
		V: "hello world",
	}

	funcConf := conf.New(env)
	_, err = checker.Check(tree, funcConf)
	require.NoError(t, err)

	program, err := compiler.Compile(tree, funcConf)
	require.NoError(t, err)

	out, err := vm.Run(program, env)
	require.NoError(t, err)

	require.Equal(t, "hello world", out)
}

func TestRun_OpInvalid(t *testing.T) {
	program := &vm.Program{
		Bytecode:  []vm.Opcode{vm.OpInvalid},
		Arguments: []int{0},
	}

	_, err := vm.Run(program, nil)
	require.EqualError(t, err, "invalid opcode")
}

// TestVM_ProfileOperations tests the profiling opcodes
func TestVM_ProfileOperations(t *testing.T) {
	program := &vm.Program{
		Bytecode: []vm.Opcode{
			vm.OpProfileStart,
			vm.OpPush,
			vm.OpProfileEnd,
		},
		Arguments: []int{0, 0, 0},
		Constants: []any{
			&vm.Span{},
		},
	}

	testVM := &vm.VM{}
	_, err := testVM.Run(program, nil)
	require.NoError(t, err)

	span := program.Constants[0].(*vm.Span)
	require.True(t, span.Duration > 0, "Profile duration should be greater than 0")
}

// TestVM_DirectCallOpcodes tests the specialized call opcodes directly
func TestVM_DirectCallOpcodes(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []vm.Opcode
		args     []int
		consts   []any
		funcs    []vm.Function
		want     any
		wantErr  bool
	}{
		{
			name:     "OpCall0",
			bytecode: []vm.Opcode{vm.OpCall0},
			args:     []int{0},
			funcs: []vm.Function{
				func(args ...any) (any, error) {
					return 42, nil
				},
			},
			want: 42,
		},
		{
			name: "OpCall1",
			bytecode: []vm.Opcode{
				vm.OpPush,
				vm.OpCall1,
			},
			args:   []int{0, 0},
			consts: []any{10},
			funcs: []vm.Function{
				func(args ...any) (any, error) {
					return args[0].(int) * 2, nil
				},
			},
			want: 20,
		},
		{
			name: "OpCall2",
			bytecode: []vm.Opcode{
				vm.OpPush,
				vm.OpPush,
				vm.OpCall2,
			},
			args:   []int{0, 1, 0},
			consts: []any{10, 5},
			funcs: []vm.Function{
				func(args ...any) (any, error) {
					return args[0].(int) + args[1].(int), nil
				},
			},
			want: 15,
		},
		{
			name: "OpCall3",
			bytecode: []vm.Opcode{
				vm.OpPush,
				vm.OpPush,
				vm.OpPush,
				vm.OpCall3,
			},
			args:   []int{0, 1, 2, 0},
			consts: []any{10, 5, 2},
			funcs: []vm.Function{
				func(args ...any) (any, error) {
					return args[0].(int) + args[1].(int) + args[2].(int), nil
				},
			},
			want: 17,
		},
		{
			name: "OpCallN with error",
			bytecode: []vm.Opcode{
				vm.OpLoadFunc,
				vm.OpCallN,
			},
			args: []int{0, 0}, // Function index, number of args (0)
			funcs: []vm.Function{
				func(args ...any) (any, error) {
					return nil, fmt.Errorf("test error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := vm.NewProgram(
				nil, // source
				nil, // node
				nil, // locations
				0,   // variables
				tt.consts,
				tt.bytecode,
				tt.args,
				tt.funcs,
				nil, // debugInfo
				nil, // span
			)
			vm := &vm.VM{}
			got, err := vm.Run(program, nil)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestVM_CallN(t *testing.T) {
	input := `fn(1, 2, 3)`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	env := map[string]any{
		"fn": func(args ...any) (any, error) {
			sum := 0
			for _, arg := range args {
				sum += arg.(int)
			}
			return sum, nil
		},
	}

	config := conf.New(env)
	program, err := compiler.Compile(tree, config)
	require.NoError(t, err)

	out, err := vm.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, 6, out)
}

// TestVM_DirectBasicOpcodes tests basic opcodes directly
func TestVM_DirectBasicOpcodes(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []vm.Opcode
		args     []int
		consts   []any
		env      any
		want     any
		wantErr  bool
	}{
		{
			name: "OpLoadEnv",
			bytecode: []vm.Opcode{
				vm.OpLoadEnv, // Load entire environment
			},
			args: []int{0},
			env:  map[string]any{"key": "value"},
			want: map[string]any{"key": "value"},
		},
		{
			name: "OpTrue",
			bytecode: []vm.Opcode{
				vm.OpTrue,
			},
			args: []int{0},
			want: true,
		},
		{
			name: "OpFalse",
			bytecode: []vm.Opcode{
				vm.OpFalse,
			},
			args: []int{0},
			want: false,
		},
		{
			name: "OpNil",
			bytecode: []vm.Opcode{
				vm.OpNil,
			},
			args: []int{0},
			want: nil,
		},
		{
			name: "OpInt",
			bytecode: []vm.Opcode{
				vm.OpInt, // Push int directly from args
			},
			args:   []int{42}, // The value 42 is passed directly in args
			consts: []any{},   // No constants needed
			want:   42,
		},
		{
			name: "OpInt negative",
			bytecode: []vm.Opcode{
				vm.OpInt, // Push negative int directly from args
			},
			args:   []int{-42}, // The value -42 is passed directly in args
			consts: []any{},    // No constants needed
			want:   -42,
		},
		{
			name: "OpInt zero",
			bytecode: []vm.Opcode{
				vm.OpInt, // Push zero directly from args
			},
			args:   []int{0}, // The value 0 is passed directly in args
			consts: []any{},  // No constants needed
			want:   0,
		},
		{
			name: "OpCast int to float64",
			bytecode: []vm.Opcode{
				vm.OpPush, // Push int
				vm.OpCast, // Cast to float64
			},
			args:   []int{0, 2},
			consts: []any{42},
			want:   float64(42),
		},
		{
			name: "OpCast int32 to int64",
			bytecode: []vm.Opcode{
				vm.OpPush, // Push int32
				vm.OpCast, // Cast to int64
			},
			args:   []int{0, 1},
			consts: []any{int32(42)},
			want:   int64(42),
		},
		{
			name: "OpCast invalid type",
			bytecode: []vm.Opcode{
				vm.OpPush, // Push string
				vm.OpCast, // Try to cast to float64
			},
			args:    []int{0, 0},
			consts:  []any{"not a number"},
			wantErr: true,
		},
		{
			name: "OpDefault",
			bytecode: []vm.Opcode{
				vm.OpEnd + 1, // OpEnd is always last, this is an unknown opcode
			},
			args:    []int{0, 0},
			consts:  []any{fmt.Errorf("test error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := vm.NewProgram(
				nil, // source
				nil, // node
				nil, // locations
				0,   // variables
				tt.consts,
				tt.bytecode,
				tt.args,
				nil, // functions
				nil, // debugInfo
				nil, // span
			)
			vm := &vm.VM{}
			got, err := vm.Run(program, tt.env)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
