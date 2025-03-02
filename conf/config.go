package conf

import (
	"reflect"

	"expr/ast"
	"expr/checker/nature"
)

type Config struct {
	EnvObject any
	Env       nature.Nature
	Expect    reflect.Kind
	ExpectAny bool
	Optimize  bool
	Strict    bool
	Profile   bool
	Visitors  []ast.Visitor
}

// CreateNew creates new config with default values.
func CreateNew() *Config {
	c := &Config{
		Optimize: true,
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
