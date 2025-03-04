package checker

import (
	. "expr/checker/nature"
	"reflect"
)

var (
	unknown       = Nature{}
	nilNature     = Nature{Nil: true}
	boolNature    = Nature{Type: reflect.TypeOf(true)}
	integerNature = Nature{Type: reflect.TypeOf(0)}
	floatNature   = Nature{Type: reflect.TypeOf(float64(0))}
	stringNature  = Nature{Type: reflect.TypeOf("")}
)

var (
	anyType = reflect.TypeOf(new(any)).Elem()
)

func isNil(nt Nature) bool {
	return nt.Nil
}

func isUnknown(nt Nature) bool {
	switch {
	case nt.Type == nil && !nt.Nil:
		return true
	case nt.Kind() == reflect.Interface:
		return true
	}
	return false
}

func isInteger(nt Nature) bool {
	switch nt.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isFloat(nt Nature) bool {
	switch nt.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func isNumber(nt Nature) bool {
	return isInteger(nt) || isFloat(nt)
}

func kind(t reflect.Type) reflect.Kind {
	if t == nil {
		return reflect.Invalid
	}
	return t.Kind()
}
