package expr

import (
	"expr/checker"
	"expr/compiler"
	"expr/conf"
	"expr/vm"
	"fmt"
	"reflect"
)

// Option for configuring config.
type Option func(c *conf.Config)

// Env specifies expected input of env for type checks.
// If struct is passed, all fields will be treated as variables,
// as well as all fields of embedded structs and struct itself.
// If map is passed, all items will be treated as variables.
// Methods defined on this type will be available as functions.
func Env(env any) Option {
	return func(c *conf.Config) {
		c.WithEnv(env)
	}
}

// AllowUndefinedVariables allows to use undefined variables inside expressions.
// This can be used with expr.Env option to partially define a few variables.
func AllowUndefinedVariables() Option {
	return func(c *conf.Config) {
		c.Strict = false
	}
}

// AsAny tells the compiler to expect any result.
func AsAny() Option {
	return func(c *conf.Config) {
		c.ExpectAny = true
	}
}

// AsKind tells the compiler to expect kind of the result.
func AsKind(kind reflect.Kind) Option {
	return func(c *conf.Config) {
		c.Expect = kind
		c.ExpectAny = true
	}
}

// AsBool tells the compiler to expect a boolean result.
func AsBool() Option {
	return func(c *conf.Config) {
		c.Expect = reflect.Bool
		c.ExpectAny = true
	}
}

// AsInt tells the compiler to expect an int result.
func AsInt() Option {
	return func(c *conf.Config) {
		c.Expect = reflect.Int
		c.ExpectAny = true
	}
}

// AsInt64 tells the compiler to expect an int64 result.
func AsInt64() Option {
	return func(c *conf.Config) {
		c.Expect = reflect.Int64
		c.ExpectAny = true
	}
}

// AsFloat64 tells the compiler to expect a float64 result.
func AsFloat64() Option {
	return func(c *conf.Config) {
		c.Expect = reflect.Float64
		c.ExpectAny = true
	}
}

// WarnOnAny tells the compiler to warn if expression return any type.
func WarnOnAny() Option {
	return func(c *conf.Config) {
		if c.Expect == reflect.Invalid {
			panic("WarnOnAny() works only with combination with AsInt(), AsBool(), etc. options")
		}
		c.ExpectAny = false
	}
}

// Compile parses and compiles given input expression to bytecode program.
func Compile(input string, ops ...Option) (*vm.Program, error) {
	config := conf.CreateNew()
	for _, op := range ops {
		op(config)
	}

	tree, err := checker.ParseCheck(input, config)
	if err != nil {
		return nil, err
	}

	program, err := compiler.Compile(tree, config)
	if err != nil {
		return nil, err
	}

	return program, nil
}

// Run evaluates given bytecode program.
func Run(program *vm.Program, env any) (any, error) {
	return vm.Run(program, env)
}

// Eval parses, compiles and runs given input.
func Eval(input string, env any) (any, error) {
	if _, ok := env.(Option); ok {
		return nil, fmt.Errorf("misused expr.Eval: second argument (env) should be passed without expr.Env")
	}

	program, err := Compile(input)
	if err != nil {
		return nil, err
	}

	output, err := Run(program, env)
	if err != nil {
		return nil, err
	}

	return output, nil
}
