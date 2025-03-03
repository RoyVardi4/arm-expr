package ast

import (
	"reflect"

	"expr/checker/nature"
	"expr/file"
)

var (
	anyType = reflect.TypeOf(new(any)).Elem()
)

// Node represents items of abstract syntax tree.
type Node interface {
	Location() file.Location
	SetLocation(file.Location)
	Nature() nature.Nature
	SetNature(nature.Nature)
	Type() reflect.Type
	SetType(reflect.Type)
	String() string
}

// Patch replaces the node with a new one.
// Location information is preserved.
// Type information is lost.
func Patch(node *Node, newNode Node) {
	newNode.SetLocation((*node).Location())
	*node = newNode
}

// base is a base struct for all nodes.
type base struct {
	loc    file.Location
	nature nature.Nature
}

// Location returns the location of the node in the source code.
func (n *base) Location() file.Location {
	return n.loc
}

// SetLocation sets the location of the node in the source code.
func (n *base) SetLocation(loc file.Location) {
	n.loc = loc
}

// Nature returns the nature of the node.
func (n *base) Nature() nature.Nature {
	return n.nature
}

// SetNature sets the nature of the node.
func (n *base) SetNature(nature nature.Nature) {
	n.nature = nature
}

// Type returns the type of the node.
func (n *base) Type() reflect.Type {
	if n.nature.Type == nil {
		return anyType
	}
	return n.nature.Type
}

// SetType sets the type of the node.
func (n *base) SetType(t reflect.Type) {
	n.nature.Type = t
}

// NilNode represents nil.
type NilNode struct {
	base
}

// IdentifierNode represents an identifier.
type IdentifierNode struct {
	base
	Value string // Name of the identifier. Like "foo" in "foo.bar".
}

// IntegerNode represents an integer.
type IntegerNode struct {
	base
	Value int // Value of the integer.
}

// FloatNode represents a float.
type FloatNode struct {
	base
	Value float64 // Value of the float.
}

// BoolNode represents a boolean.
type BoolNode struct {
	base
	Value bool // Value of the boolean.
}

// StringNode represents a string.
type StringNode struct {
	base
	Value string // Value of the string.
}

// MemberNode represents a member access.
// It can be a field access, a method call,
// or an array element access.
// Example:
//
//	foo.bar or foo["bar"]
//	foo.bar()
//	array[0]
type MemberNode struct {
	base
	Node     Node // Node of the member access. Like "foo" in "foo.bar".
	Property Node // Property of the member access. For property access it is a StringNode.
	Method   bool
}

// CallNode represents a function or a method call.
type CallNode struct {
	base
	Callee    Node   // Node of the call. Like "foo" in "foo()".
	Arguments []Node // Arguments of the call.
}

// ArrayNode represents an array.
type ArrayNode struct {
	base
	Nodes []Node // Nodes of the array.
}

// MapNode represents a map.
type MapNode struct {
	base
	Pairs []Node // PairNode nodes.
}

// PairNode represents a key-value pair of a map.
type PairNode struct {
	base
	Key   Node // Key of the pair.
	Value Node // Value of the pair.
}
