package conf

import (
	"fmt"
	"reflect"

	"expr/ast"
	"expr/builtin"
	"expr/checker/nature"
	"expr/vm/runtime"
)

type FunctionsTable map[string]*builtin.Function

type Config struct {
	EnvObject any
	Env       nature.Nature
	Expect    reflect.Kind
	ExpectAny bool
	Optimize  bool
	Strict    bool
	Profile   bool
	ConstFns  map[string]reflect.Value
	Visitors  []ast.Visitor
}

// CreateNew creates new config with default values.
func CreateNew() *Config {
	c := &Config{
		Optimize: true,
		ConstFns: make(map[string]reflect.Value),
	}
	return c
}

// New creates new config with environment.
func New(env any) *Config {
	c := CreateNew()
	c.WithEnv(env)
	return c
}

func (c *Config) WithEnv(env any) {
	c.Strict = true
	c.EnvObject = env
	c.Env = Env(env)
}

func (c *Config) ConstExpr(name string) {
	if c.EnvObject == nil {
		panic("no environment is specified for ConstExpr()")
	}
	fn := reflect.ValueOf(runtime.Fetch(c.EnvObject, name))
	if fn.Kind() != reflect.Func {
		panic(fmt.Errorf("const expression %q must be a function", name))
	}
	c.ConstFns[name] = fn
}

type Checker interface {
	Check()
}

func (c *Config) Check() {
	for _, v := range c.Visitors {
		if c, ok := v.(Checker); ok {
			c.Check()
		}
	}
}
