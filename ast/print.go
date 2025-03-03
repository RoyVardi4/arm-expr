package ast

import (
	"fmt"
	"strings"

	"expr/parser/utils"
)

func (n *NilNode) String() string {
	return "nil"
}

func (n *IdentifierNode) String() string {
	return n.Value
}

func (n *IntegerNode) String() string {
	return fmt.Sprintf("%d", n.Value)
}

func (n *FloatNode) String() string {
	return fmt.Sprintf("%v", n.Value)
}

func (n *BoolNode) String() string {
	return fmt.Sprintf("%t", n.Value)
}

func (n *StringNode) String() string {
	return fmt.Sprintf("%q", n.Value)
}

func (n *ChainNode) String() string {
	return n.Node.String()
}

func (n *MemberNode) String() string {
	node := n.Node.String()

	if str, ok := n.Property.(*StringNode); ok && utils.IsValidIdentifier(str.Value) {
		return fmt.Sprintf("%s.%s", node, str.Value)
	}
	return fmt.Sprintf("%s[%s]", node, n.Property.String())
}

func (n *CallNode) String() string {
	arguments := make([]string, len(n.Arguments))
	for i, arg := range n.Arguments {
		arguments[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", n.Callee.String(), strings.Join(arguments, ", "))
}

func (n *ArrayNode) String() string {
	nodes := make([]string, len(n.Nodes))
	for i, node := range n.Nodes {
		nodes[i] = node.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(nodes, ", "))
}

func (n *MapNode) String() string {
	pairs := make([]string, len(n.Pairs))
	for i, pair := range n.Pairs {
		pairs[i] = pair.String()
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func (n *PairNode) String() string {
	if str, ok := n.Key.(*StringNode); ok {
		if utils.IsValidIdentifier(str.Value) {
			return fmt.Sprintf("%s: %s", str.Value, n.Value.String())
		}
		return fmt.Sprintf("%s: %s", str.String(), n.Value.String())
	}
	return fmt.Sprintf("(%s): %s", n.Key.String(), n.Value.String())
}
