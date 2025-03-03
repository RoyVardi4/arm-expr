package expr_test

import (
	"expr"
	"expr/file"
	"expr/internal/testify/assert"
	"expr/internal/testify/require"
	"expr/types"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	default:
		strValue := fmt.Sprintf("%v", v)
		return strconv.ParseBool(strValue)
	}
}

func TestExpr_roy(t *testing.T) {
	env := map[string]any{
		"foo": map[string]any{
			"bar": []string{"hello", "world"},
		},
		"and": func(args ...bool) bool {
			for _, arg := range args {
				if !arg {
					return false
				}
			}
			return true
		},
		"or": func(args ...bool) bool {
			for _, arg := range args {
				if arg {
					return true
				}
			}
			return false
		},
		"bool": func(val interface{}) bool {
			boolVal, err := toBool(val)
			if err != nil {
				panic(err)
			}
			return boolVal
		},
		"not": func(val bool) bool {
			return !val
		},
		"true":  func() bool { return true },
		"false": func() bool { return false },
		"if": func(cond bool, trueVal interface{}, falseVal interface{}) interface{} {
			if cond {
				return trueVal
			}
			return falseVal
		},
	}

	tests := []struct{ code string }{
		{`or(false, false, false)`},
		{"if(not(false()), false, 'byeeee')"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			program, err := expr.Compile(tt.code, expr.Env(env))

			require.NoError(t, err)

			output, err := expr.Run(program, env)
			require.NoError(t, err)
			require.Equal(t, false, output)
		})
	}
}

func ExampleEnv_tagged_field_names() {
	env := struct {
		FirstWord  string
		Separator  string `expr:"Space"`
		SecondWord string `expr:"second_word"`
	}{
		FirstWord:  "Hello",
		Separator:  " ",
		SecondWord: "World",
	}

	output, err := expr.Eval(`FirstWord + Space + second_word`, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output)

	// Output : Hello World
}

func ExampleAsInt() {
	program, err := expr.Compile("42", expr.AsInt())
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := expr.Run(program, nil)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%T(%v)", output, output)

	// Output: int(42)
}

func ExampleAsInt64() {
	env := map[string]any{
		"rating": 5.5,
	}

	program, err := expr.Compile("rating", expr.Env(env), expr.AsInt64())
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := expr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output.(int64))

	// Output: 5
}

func ExampleAsFloat64() {
	program, err := expr.Compile("42", expr.AsFloat64())
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := expr.Run(program, nil)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output.(float64))

	// Output: 42
}

func ExampleWarnOnAny() {
	// Arrays always have []any type. The expression return type is any.
	// AsInt() instructs compiler to expect int or any, and cast to int,
	// if possible. WarnOnAny() instructs to return an error on any type.
	_, err := expr.Compile(`[42, true, "yes"][0]`, expr.AsInt(), expr.WarnOnAny())

	fmt.Printf("%v", err)

	// Output: expected int, but got interface {}
}

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func TestExpr_readme_example(t *testing.T) {
	env := map[string]any{
		"greet":   "Hello, %v!",
		"names":   []string{"world", "you"},
		"sprintf": fmt.Sprintf,
	}

	code := `sprintf(greet, names[0])`

	program, err := expr.Compile(code, expr.Env(env))
	require.NoError(t, err)

	output, err := expr.Run(program, env)
	require.NoError(t, err)

	require.Equal(t, "Hello, world!", output)
}

func TestExpr_chaining_nested_chains(t *testing.T) {
	env := map[string]any{
		"foo": map[string]any{
			"id": 1,
			"bar": []map[string]any{
				1: {
					"baz": "baz",
				},
			},
		},
	}
	program, err := expr.Compile("foo.bar[foo.id]['baz']", expr.Env(env))
	require.NoError(t, err)

	got, err := expr.Run(program, env)
	require.NoError(t, err)
	assert.Equal(t, "baz", got)
}

func TestExpr_eval_with_env(t *testing.T) {
	_, err := expr.Eval("true", expr.Env(map[string]any{}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "misused")
}

func TestExpr_fetch_from_func(t *testing.T) {
	_, err := expr.Eval("foo.Value", map[string]any{
		"foo": func() {},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot fetch Value from func()")
}

func TestExpr_map_default_values(t *testing.T) {
	env := map[string]any{
		"foo": map[string]string{},
	}

	input := `foo['missing']`

	program, err := expr.Compile(input, expr.Env(env))
	require.NoError(t, err)

	output, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, "", output)
}

type divideError struct{ Message string }

func (e divideError) Error() string {
	return e.Message
}

func TestAsBool_exposed_error(t *testing.T) {
	_, err := expr.Compile(`42`, expr.AsBool())
	require.Error(t, err)

	_, ok := err.(*file.Error)
	require.False(t, ok, "error must not be of type *file.Error")
	require.Equal(t, "expected bool, but got int", err.Error())
}

func TestIssue271(t *testing.T) {
	type BarArray []float64

	type Foo struct {
		Bar BarArray
		Baz int
	}

	type Env struct {
		Foo Foo
	}

	code := `Foo.Bar[0]`

	program, err := expr.Compile(code, expr.Env(Env{}))
	require.NoError(t, err)

	output, err := expr.Run(program, Env{
		Foo: Foo{
			Bar: BarArray{1.0, 2.0, 3.0},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1.0, output)
}

type Issue346Array []Issue346Type

type Issue346Type struct {
	Bar string
}

func (i Issue346Array) Len() int {
	return len(i)
}

func TestIssue346(t *testing.T) {
	code := `Foo[0].Bar`

	env := map[string]any{
		"Foo": Issue346Array{
			{Bar: "bar"},
		},
	}
	program, err := expr.Compile(code, expr.Env(env))
	require.NoError(t, err)

	output, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, "bar", output)
}

func TestFastCall(t *testing.T) {
	env := map[string]any{
		"func": func(in any) float64 {
			return 8
		},
	}
	code := `func("8")`

	program, err := expr.Compile(code, expr.Env(env))
	assert.NoError(t, err)

	out, err := expr.Run(program, env)
	assert.NoError(t, err)
	assert.Equal(t, float64(8), out)
}

func TestFastCall_OpCallFastErr(t *testing.T) {
	env := map[string]any{
		"func": func(...any) (any, error) {
			return 8, nil
		},
	}
	code := `func("8")`

	program, err := expr.Compile(code, expr.Env(env))
	assert.NoError(t, err)

	out, err := expr.Run(program, env)
	assert.NoError(t, err)
	assert.Equal(t, 8, out)
}

func TestRun_custom_func_returns_an_error_as_second_arg(t *testing.T) {
	env := map[string]any{
		"semver": func(value string, cmp string) (bool, error) { return true, nil },
	}

	p, err := expr.Compile(`semver("1.2.3", "= 1.2.3")`, expr.Env(env))
	assert.NoError(t, err)

	out, err := expr.Run(p, env)
	assert.NoError(t, err)
	assert.Equal(t, true, out)
}

func TestEval_nil_in_maps(t *testing.T) {
	env := map[string]any{
		"m":     map[any]any{nil: "bar"},
		"empty": map[any]any{},
	}
	t.Run("nil key exists", func(t *testing.T) {
		p, err := expr.Compile(`m[nil]`, expr.Env(env))
		assert.NoError(t, err)

		out, err := expr.Run(p, env)
		assert.NoError(t, err)
		assert.Equal(t, "bar", out)
	})
	t.Run("no nil key", func(t *testing.T) {
		p, err := expr.Compile(`empty[nil]`, expr.Env(env))
		assert.NoError(t, err)

		out, err := expr.Run(p, env)
		assert.NoError(t, err)
		assert.Equal(t, nil, out)
	})
}

// Test the use of env keyword.  Forms env[] and env[”] are valid.
// The enclosed identifier must be in the expression env.
func TestEnv_keyword(t *testing.T) {
	env := map[string]any{
		"space test":                       "ok",
		"space_test":                       "not ok", // Seems to be some underscore substituting happening, check that.
		"Section 1-2a":                     "ok",
		`c:\ndrive\2015 Information Table`: "ok",
		"%*worst function name ever!!": func() string {
			return "ok"
		}(),
		"1":      "o",
		"2":      "k",
		"num":    10,
		"mylist": []int{1, 2, 3, 4, 5},
		"MIN": func(a, b int) int {
			if a < b {
				return a
			} else {
				return b
			}
		},
		"red":   "n",
		"irect": "um",
		"String Map": map[string]string{
			"one":   "two",
			"three": "four",
		},
		"OtherMap": map[string]string{
			"a": "b",
			"c": "d",
		},
	}

	// No error cases
	var tests = []struct {
		code string
		want any
	}{
		{"$env['space test']", "ok"},
		{"$env['Section 1-2a']", "ok"},
		{`$env["c:\\ndrive\\2015 Information Table"]`, "ok"},
		{"$env['%*worst function name ever!!']", "ok"},
		{"$env['String Map'].one", "two"},
		{"MIN($env['num'],0)", 0},
		{"$env.red", "n"},
		{"$env.mylist[1]", 2},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {

			program, err := expr.Compile(tt.code, expr.Env(env))
			require.NoError(t, err, "compile error")

			got, err := expr.Run(program, env)
			require.NoError(t, err, "execution error")

			assert.Equal(t, tt.want, got, tt.code)
		})
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got, err := expr.Eval(tt.code, env)
			require.NoError(t, err, "eval error: "+tt.code)

			assert.Equal(t, tt.want, got, "eval: "+tt.code)
		})
	}

	// error cases
	tests = []struct {
		code string
		want any
	}{
		{"env()", "bad"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			_, err := expr.Eval(tt.code, expr.Env(env))
			require.Error(t, err, "compile error")

		})
	}
}

func TestIssue432(t *testing.T) {
	env := map[string]any{
		"func": func(
			paramUint32 uint32,
			paramUint16 uint16,
			paramUint8 uint8,
			paramUint uint,
			paramInt32 int32,
			paramInt16 int16,
			paramInt8 int8,
			paramInt int,
			paramFloat64 float64,
			paramFloat32 float32,
		) float64 {
			return float64(paramUint32) + float64(paramUint16) + float64(paramUint8) + float64(paramUint) +
				float64(paramInt32) + float64(paramInt16) + float64(paramInt8) + float64(paramInt) +
				float64(paramFloat64) + float64(paramFloat32)
		},
	}
	code := `func(1,1,1,1,1,1,1,1,1,1)`

	program, err := expr.Compile(code, expr.Env(env))
	assert.NoError(t, err)

	out, err := expr.Run(program, env)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), out)
}

func TestIssue_integer_truncated_by_compiler(t *testing.T) {
	env := map[string]any{
		"fn": func(x byte) byte {
			return x
		},
	}

	_, err := expr.Compile("fn(255)", expr.Env(env))
	require.NoError(t, err)

	_, err = expr.Compile("fn(256)", expr.Env(env))
	require.Error(t, err)
}

func TestExpr_crash(t *testing.T) {
	content, err := os.ReadFile("testdata/crash.txt")
	require.NoError(t, err)

	_, err = expr.Compile(string(content))
	require.Error(t, err)
}

func TestExpr_env_types_map(t *testing.T) {
	envTypes := types.Map{
		"foo": types.Map{
			"bar": types.String,
		},
	}

	program, err := expr.Compile(`foo.bar`, expr.Env(envTypes))
	require.NoError(t, err)

	env := map[string]any{
		"foo": map[string]any{
			"bar": "value",
		},
	}

	output, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, "value", output)
}

func TestExpr_env_types_map_error(t *testing.T) {
	envTypes := types.Map{
		"foo": types.Map{
			"bar": types.String,
		},
	}

	program, err := expr.Compile(`foo.bar`, expr.Env(envTypes))
	require.NoError(t, err)

	_, err = expr.Run(program, envTypes)
	require.Error(t, err)
}
