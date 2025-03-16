package issues584_test

type Env struct{}

type Program struct {
}

func (p *Program) Foo() Value {
	return func(e *Env) float64 {
		return 5
	}
}

func (p *Program) Bar() Value {
	return func(e *Env) float64 {
		return 100
	}
}

func (p *Program) AndCondition(a, b Condition) Conditions {
	return Conditions{a, b}
}

func (p *Program) AndConditions(a Conditions, b Condition) Conditions {
	return append(a, b)
}

func (p *Program) ValueGreaterThan_float(v Value, i float64) Condition {
	return func(e *Env) bool {
		realized := v(e)
		return realized > i
	}
}

func (p *Program) ValueLessThan_float(v Value, i float64) Condition {
	return func(e *Env) bool {
		realized := v(e)
		return realized < i
	}
}

type Condition func(e *Env) bool
type Conditions []Condition

type Value func(e *Env) float64
