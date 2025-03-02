package patch_test

import (
	"expr/ast"
)

type Env struct {
	Foo int `expr:"foo"`
	Bar int `expr:"bar"`
}

type changeIdent struct{}

func (changeIdent) Visit(node *ast.Node) {
	id, ok := (*node).(*ast.IdentifierNode)
	if !ok {
		return
	}
	if id.Value == "foo" {
		// A correct way to patch the node:
		//
		//	newNode := &ast.IdentifierNode{Value: "bar"}
		//	ast.Patch(node, newNode)
		//
		// But we can do it in a wrong way:
		id.Value = "bar"
	}
}
