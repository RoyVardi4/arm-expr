package value_test

type myInt struct {
	Int int
}

func (v *myInt) AsInt() int {
	return v.Int
}

func (v *myInt) AsAny() any {
	return v.Int
}
