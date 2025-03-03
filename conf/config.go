package conf

import (
	"reflect"

	"expr/checker/nature"
)

type Config struct {
	EnvObject any
	Env       nature.Nature
	Expect    reflect.Kind
	ExpectAny bool
	Strict    bool
	Profile   bool
}

// CreateNew creates new config with default values.
func CreateNew() *Config {
	c := &Config{}
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
