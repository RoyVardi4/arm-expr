// Package value provides a Patcher that uses interfaces to allow custom types that can be represented as standard go values to be used more easily in expressions.
package value

import (
	"reflect"
	"time"

	"expr/ast"
)

// A AnyValuer provides a generic function for a custom type to return standard go values.
// It allows for returning a `nil` value but does not provide any type checking at expression compile.
//
// A custom type may implement both AnyValuer and a type specific interface to enable both
// compile time checking and the ability to return a `nil` value.
type AnyValuer interface {
	AsAny() any
}

type IntValuer interface {
	AsInt() int
}

type BoolValuer interface {
	AsBool() bool
}

type Int8Valuer interface {
	AsInt8() int8
}

type Int16Valuer interface {
	AsInt16() int16
}

type Int32Valuer interface {
	AsInt32() int32
}

type Int64Valuer interface {
	AsInt64() int64
}

type UintValuer interface {
	AsUint() uint
}

type Uint8Valuer interface {
	AsUint8() uint8
}

type Uint16Valuer interface {
	AsUint16() uint16
}

type Uint32Valuer interface {
	AsUint32() uint32
}

type Uint64Valuer interface {
	AsUint64() uint64
}

type Float32Valuer interface {
	AsFloat32() float32
}

type Float64Valuer interface {
	AsFloat64() float64
}

type StringValuer interface {
	AsString() string
}

type TimeValuer interface {
	AsTime() time.Time
}

type DurationValuer interface {
	AsDuration() time.Duration
}

type ArrayValuer interface {
	AsArray() []any
}

type MapValuer interface {
	AsMap() map[string]any
}

var supportedInterfaces = []reflect.Type{
	reflect.TypeOf((*AnyValuer)(nil)).Elem(),
	reflect.TypeOf((*BoolValuer)(nil)).Elem(),
	reflect.TypeOf((*IntValuer)(nil)).Elem(),
	reflect.TypeOf((*Int8Valuer)(nil)).Elem(),
	reflect.TypeOf((*Int16Valuer)(nil)).Elem(),
	reflect.TypeOf((*Int32Valuer)(nil)).Elem(),
	reflect.TypeOf((*Int64Valuer)(nil)).Elem(),
	reflect.TypeOf((*UintValuer)(nil)).Elem(),
	reflect.TypeOf((*Uint8Valuer)(nil)).Elem(),
	reflect.TypeOf((*Uint16Valuer)(nil)).Elem(),
	reflect.TypeOf((*Uint32Valuer)(nil)).Elem(),
	reflect.TypeOf((*Uint64Valuer)(nil)).Elem(),
	reflect.TypeOf((*Float32Valuer)(nil)).Elem(),
	reflect.TypeOf((*Float64Valuer)(nil)).Elem(),
	reflect.TypeOf((*StringValuer)(nil)).Elem(),
	reflect.TypeOf((*TimeValuer)(nil)).Elem(),
	reflect.TypeOf((*DurationValuer)(nil)).Elem(),
	reflect.TypeOf((*ArrayValuer)(nil)).Elem(),
	reflect.TypeOf((*MapValuer)(nil)).Elem(),
}

type patcher struct{}

func (patcher) Visit(node *ast.Node) {
	switch id := (*node).(type) {
	case *ast.IdentifierNode, *ast.MemberNode:
		nodeType := id.Type()
		for _, t := range supportedInterfaces {
			if nodeType.Implements(t) {
				ast.Patch(node, &ast.CallNode{
					Callee:    &ast.IdentifierNode{Value: "$patcher_value_getter"},
					Arguments: []ast.Node{id},
				})
				return
			}
		}
	}
}
