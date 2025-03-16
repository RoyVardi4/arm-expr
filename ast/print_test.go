package ast_test

import (
	"testing"

	"expr/internal/testify/assert"
	"expr/internal/testify/require"

	"expr/ast"
	"expr/parser"
)

func TestPrint(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`nil`, `nil`},
		{`true`, `true`},
		{`false`, `false`},
		{`1`, `1`},
		{`1.1`, `1.1`},
		{`"a"`, `"a"`},
		{`'a'`, `"a"`},
		{`a`, `a`},
		{`a.b`, `a.b`},
		{`a[0]`, `a[0]`},
		{`a["the b"]`, `a["the b"]`},
		{`a.b[0]`, `a.b[0]`},
		{`x[0][1]`, `x[0][1]`},
		{`func()`, `func()`},
		{`func(a)`, `func(a)`},
		{`func(a, b)`, `func(a, b)`},
		{`len(a)`, `len(a)`},
		{`a.b()`, `a.b()`},
		{`a.b(c)`, `a.b(c)`}}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tree, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, tree.Node.String())
		})
	}
}

func TestPrint_MemberNode(t *testing.T) {
	node := &ast.MemberNode{
		Node: &ast.IdentifierNode{
			Value: "a",
		},
		Property: &ast.StringNode{Value: "b c"},
	}
	require.Equal(t, `a["b c"]`, node.String())
}
