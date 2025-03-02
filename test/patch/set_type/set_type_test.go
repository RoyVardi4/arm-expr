package set_type_test

import (
	"expr/ast"
	"reflect"
)

type Value struct {
	Int int
}

type Env struct {
	Value Value
}

var valueType = reflect.TypeOf((*Value)(nil)).Elem()

type getValuePatcher struct{}

func (getValuePatcher) Visit(node *ast.Node) {
	id, ok := (*node).(*ast.IdentifierNode)
	if !ok {
		return
	}
	if id.Type() == valueType {
		newNode := &ast.CallNode{
			Callee:    &ast.IdentifierNode{Value: "getValue"},
			Arguments: []ast.Node{id},
		}
		newNode.SetType(reflect.TypeOf(0))
		ast.Patch(node, newNode)
	}
}
