package parser_test

import (
	"fmt"
	"strings"
	"testing"

	"expr/internal/testify/assert"
	"expr/internal/testify/require"

	. "expr/ast"
	"expr/parser"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  Node
	}{
		{
			"a",
			&IdentifierNode{Value: "a"},
		},
		{
			`"str"`,
			&StringNode{Value: "str"},
		},
		{
			"`hello\nworld`",
			&StringNode{Value: `hello
world`},
		},
		{
			"3",
			&IntegerNode{Value: 3},
		},
		{
			"0xFF",
			&IntegerNode{Value: 255},
		},
		{
			"0x6E",
			&IntegerNode{Value: 110},
		},
		{
			"0X63",
			&IntegerNode{Value: 99},
		},
		{
			"0o600",
			&IntegerNode{Value: 384},
		},
		{
			"0O45",
			&IntegerNode{Value: 37},
		},
		{
			"0b10",
			&IntegerNode{Value: 2},
		},
		{
			"0B101011",
			&IntegerNode{Value: 43},
		},
		{
			"10_000_000",
			&IntegerNode{Value: 10_000_000},
		},
		{
			"2.5",
			&FloatNode{Value: 2.5},
		},
		{
			"1e9",
			&FloatNode{Value: 1e9},
		},
		{
			"true",
			&BoolNode{Value: true},
		},
		{
			"false",
			&BoolNode{Value: false},
		},
		{
			"nil",
			&NilNode{},
		},
		{
			"foo(bar())",
			&CallNode{Callee: &IdentifierNode{Value: "foo"},
				Arguments: []Node{&CallNode{Callee: &IdentifierNode{Value: "bar"},
					Arguments: []Node{}}}},
		},
		{
			`foo("arg1", 2, true)`,
			&CallNode{Callee: &IdentifierNode{Value: "foo"},
				Arguments: []Node{&StringNode{Value: "arg1"},
					&IntegerNode{Value: 2},
					&BoolNode{Value: true}}},
		},
		{
			"foo.bar",
			&MemberNode{Node: &IdentifierNode{Value: "foo"},
				Property: &StringNode{Value: "bar"}},
		},
		{
			"foo['all']",
			&MemberNode{Node: &IdentifierNode{Value: "foo"},
				Property: &StringNode{Value: "all"}},
		},
		{
			"foo.bar()",
			&CallNode{Callee: &MemberNode{Node: &IdentifierNode{Value: "foo"},
				Property: &StringNode{Value: "bar"}, Method: true},
				Arguments: []Node{}},
		},
		{
			`foo.bar("arg1", 2, true)`,
			&CallNode{Callee: &MemberNode{Node: &IdentifierNode{Value: "foo"},
				Property: &StringNode{Value: "bar"}, Method: true},
				Arguments: []Node{&StringNode{Value: "arg1"},
					&IntegerNode{Value: 2},
					&BoolNode{Value: true}}},
		},
		{
			"foo[3]",
			&MemberNode{Node: &IdentifierNode{Value: "foo"},
				Property: &IntegerNode{Value: 3}},
		},
		{
			"a.b().c().d[33]",
			&MemberNode{
				Node: &MemberNode{
					Node: &CallNode{
						Callee: &MemberNode{
							Node: &CallNode{
								Callee: &MemberNode{
									Node: &IdentifierNode{
										Value: "a",
									},
									Property: &StringNode{
										Value: "b",
									},
									Method: true,
								},
								Arguments: []Node{},
							},
							Property: &StringNode{
								Value: "c",
							},
							Method: true,
						},
						Arguments: []Node{},
					},
					Property: &StringNode{
						Value: "d",
					},
				},
				Property: &IntegerNode{Value: 33}},
		},
		{
			"[a, b, c]",
			&ArrayNode{Nodes: []Node{&IdentifierNode{Value: "a"},
				&IdentifierNode{Value: "b"},
				&IdentifierNode{Value: "c"}}},
		},
		{
			"[1].foo",
			&MemberNode{Node: &ArrayNode{Nodes: []Node{&IntegerNode{Value: 1}}},
				Property: &StringNode{Value: "foo"}},
		},
		{
			"[]",
			&ArrayNode{},
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual, err := parser.Parse(test.input)
			require.NoError(t, err)
			assert.Equal(t, Dump(test.want), Dump(actual.Node))
		})
	}
}

const errorTests = `
foo.
unexpected end of expression (1:4)
 | foo.
 | ...^

foo.bar(a b)
unexpected token Identifier("b") (1:11)
 | foo.bar(a b)
 | ..........^
`

func TestParse_error(t *testing.T) {
	tests := strings.Split(strings.Trim(errorTests, "\n"), "\n\n")
	for _, test := range tests {
		input := strings.SplitN(test, "\n", 2)
		if len(input) != 2 {
			t.Errorf("syntax error in test: %q", test)
			break
		}
		_, err := parser.Parse(input[0])
		if err == nil {
			err = fmt.Errorf("<nil>")
		}
		assert.Equal(t, input[1], err.Error(), input[0])
	}
}

func TestParse_optional_chaining(t *testing.T) {
	parseTests := []struct {
		input    string
		expected Node
	}{
		{
			"foo.bar.baz",
			&ChainNode{
				Node: &MemberNode{
					Node: &MemberNode{
						Node:     &IdentifierNode{Value: "foo"},
						Property: &StringNode{Value: "bar"},
					},
					Property: &StringNode{Value: "baz"},
				},
			},
		},
		{
			"foo.bar[a.b].baz",
			&ChainNode{
				Node: &MemberNode{
					Node: &MemberNode{
						Node: &MemberNode{
							Node:     &IdentifierNode{Value: "foo"},
							Property: &StringNode{Value: "bar"},
						},
						Property: &ChainNode{
							Node: &MemberNode{
								Node:     &IdentifierNode{Value: "a"},
								Property: &StringNode{Value: "b"},
							},
						},
					},
					Property: &StringNode{Value: "baz"},
				},
			},
		},
		{
			"foo.bar[0]",
			&ChainNode{
				Node: &MemberNode{
					Node: &MemberNode{
						Node:     &IdentifierNode{Value: "foo"},
						Property: &StringNode{Value: "bar"},
					},
					Property: &IntegerNode{Value: 0},
				},
			},
		},
	}
	for _, test := range parseTests {
		actual, err := parser.Parse(test.input)
		if err != nil {
			t.Errorf("%s:\n%v", test.input, err)
			continue
		}
		assert.Equal(t, Dump(test.expected), Dump(actual.Node), test.input)
	}
}
