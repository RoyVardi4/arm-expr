package checker_test

import (
	"fmt"
	"reflect"
	"testing"

	"expr/internal/testify/assert"
	"expr/internal/testify/require"
	"expr/types"

	"expr"
	"expr/checker"
	"expr/conf"
	"expr/parser"
)

func TestCheck_AsBool(t *testing.T) {
	tree, err := parser.Parse(`1`)
	require.NoError(t, err)

	config := &conf.Config{}
	expr.AsBool()(config)

	_, err = checker.Check(tree, config)
	assert.Error(t, err)
	assert.Equal(t, "expected bool, but got int", err.Error())
}

func TestCheck_AsInt64(t *testing.T) {
	tree, err := parser.Parse(`true`)
	require.NoError(t, err)

	config := &conf.Config{}
	expr.AsInt64()(config)

	_, err = checker.Check(tree, config)
	assert.Error(t, err)
	assert.Equal(t, "expected int64, but got bool", err.Error())
}

func TestCheck_TaggedFieldName(t *testing.T) {
	tree, err := parser.Parse(`foo.bar`)
	require.NoError(t, err)

	config := conf.CreateNew()
	expr.Env(struct {
		x struct {
			y bool `expr:"bar"`
		} `expr:"foo"`
	}{})(config)
	expr.AsBool()(config)

	_, err = checker.Check(tree, config)
	assert.NoError(t, err)
}

func TestCheck_NoConfig(t *testing.T) {
	tree, err := parser.Parse(`any`)
	require.NoError(t, err)

	_, err = checker.Check(tree, conf.CreateNew())
	assert.NoError(t, err)
}

func TestCheck_AllowUndefinedVariables_DefaultType(t *testing.T) {
	env := map[string]bool{}

	tree, err := parser.Parse(`Any`)
	require.NoError(t, err)

	config := conf.New(env)
	expr.AllowUndefinedVariables()(config)
	expr.AsBool()(config)

	_, err = checker.Check(tree, config)
	assert.NoError(t, err)
}

func TestCheck_works_with_nil_types(t *testing.T) {
	env := map[string]any{
		"null": nil,
	}

	tree, err := parser.Parse("null")
	require.NoError(t, err)

	_, err = checker.Check(tree, conf.New(env))
	require.NoError(t, err)
}

func TestCheck_cast_to_expected_works_with_interface(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		type Env struct {
			Any any
		}

		tree, err := parser.Parse("Any")
		require.NoError(t, err)

		config := conf.New(Env{})
		expr.AsFloat64()(config)
		expr.AsAny()(config)

		_, err = checker.Check(tree, config)
		require.NoError(t, err)
	})

	t.Run("kind", func(t *testing.T) {
		env := map[string]any{
			"Any": any("foo"),
		}

		tree, err := parser.Parse("Any")
		require.NoError(t, err)

		config := conf.New(env)
		expr.AsKind(reflect.String)(config)

		_, err = checker.Check(tree, config)
		require.NoError(t, err)
	})
}

func TestCheck_dont_panic_on_nil_arguments_for_builtins(t *testing.T) {
	tests := []string{
		"len(nil)",
		"abs(nil)",
		"int(nil)",
		"float(nil)",
	}
	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			tree, err := parser.Parse(test)
			require.NoError(t, err)

			_, err = checker.Check(tree, conf.New(nil))
			require.Error(t, err)
		})
	}
}

func TestCheck_env_keyword(t *testing.T) {
	env := map[string]any{
		"num":  42,
		"str":  "foo",
		"name": "str",
	}

	tests := []struct {
		input string
		want  reflect.Kind
	}{
		{`$env['str']`, reflect.String},
		{`$env['num']`, reflect.Int},
		{`$env[name]`, reflect.Interface},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			tree, err := parser.Parse(test.input)
			require.NoError(t, err)

			rtype, err := checker.Check(tree, conf.New(env))
			require.NoError(t, err)
			require.True(t, rtype.Kind() == test.want, fmt.Sprintf("expected %s, got %s", test.want, rtype.Kind()))
		})
	}
}

func TestCheck_types(t *testing.T) {
	env := types.Map{
		"foo": types.Map{
			"bar": types.Map{
				"baz":       types.String,
				types.Extra: types.String,
			},
		},
		"arr": types.Array(types.Map{
			"value": types.String,
		}),
		types.Extra: types.Any,
	}

	noerr := "no error"
	tests := []struct {
		code string
		err  string
	}{
		{`unknown`, noerr},
		{`foo.unknown.baz`, `unknown field unknown (1:5)`},
		{`foo.bar.unknown`, noerr},
	}

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			tree, err := parser.Parse(test.code)
			require.NoError(t, err)

			config := conf.New(env)
			_, err = checker.Check(tree, config)
			if test.err == noerr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.err)
			}
		})
	}
}
