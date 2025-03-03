package compiler

import (
	"expr/ast"
	"expr/checker"
	. "expr/checker/nature"
	"expr/conf"
	"expr/file"
	"expr/parser"
	. "expr/vm"
	"expr/vm/runtime"
	"fmt"
	"math"
	"reflect"
)

const (
	placeholder = 12345
)

func Compile(tree *parser.Tree, config *conf.Config) (program *Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	c := &compiler{
		config:         config,
		locations:      make([]file.Location, 0),
		constantsIndex: make(map[any]int),
		functionsIndex: make(map[string]int),
		debugInfo:      make(map[string]string),
	}

	c.compile(tree.Node)

	if c.config != nil {
		switch c.config.Expect {
		case reflect.Int:
			c.emit(OpCast, 0)
		case reflect.Int64:
			c.emit(OpCast, 1)
		case reflect.Float64:
			c.emit(OpCast, 2)
		}
	}

	var span *Span
	if len(c.spans) > 0 {
		span = c.spans[0]
	}

	program = NewProgram(
		tree.Source,
		tree.Node,
		c.locations,
		c.variables,
		c.constants,
		c.bytecode,
		c.arguments,
		c.functions,
		c.debugInfo,
		span,
	)
	return
}

type compiler struct {
	config         *conf.Config
	locations      []file.Location
	bytecode       []Opcode
	variables      int
	scopes         []scope
	constants      []any
	constantsIndex map[any]int
	functions      []Function
	functionsIndex map[string]int
	debugInfo      map[string]string
	nodes          []ast.Node
	spans          []*Span
	chains         [][]int
	arguments      []int
}

type scope struct {
	variableName string
	index        int
}

func (c *compiler) nodeParent() ast.Node {
	if len(c.nodes) > 1 {
		return c.nodes[len(c.nodes)-2]
	}
	return nil
}

func (c *compiler) emitLocation(loc file.Location, op Opcode, arg int) int {
	c.bytecode = append(c.bytecode, op)
	current := len(c.bytecode)
	c.arguments = append(c.arguments, arg)
	c.locations = append(c.locations, loc)
	return current
}

func (c *compiler) emit(op Opcode, args ...int) int {
	arg := 0
	if len(args) > 1 {
		panic("too many arguments")
	}
	if len(args) == 1 {
		arg = args[0]
	}
	var loc file.Location
	if len(c.nodes) > 0 {
		loc = c.nodes[len(c.nodes)-1].Location()
	}
	return c.emitLocation(loc, op, arg)
}

func (c *compiler) emitPush(value any) int {
	return c.emit(OpPush, c.addConstant(value))
}

func (c *compiler) addConstant(constant any) int {
	indexable := true
	hash := constant
	switch reflect.TypeOf(constant).Kind() {
	case reflect.Slice, reflect.Map, reflect.Struct, reflect.Func:
		indexable = false
	}
	if field, ok := constant.(*runtime.Field); ok {
		indexable = true
		hash = fmt.Sprintf("%v", field)
	}
	if method, ok := constant.(*runtime.Method); ok {
		indexable = true
		hash = fmt.Sprintf("%v", method)
	}
	if indexable {
		if p, ok := c.constantsIndex[hash]; ok {
			return p
		}
	}
	c.constants = append(c.constants, constant)
	p := len(c.constants) - 1
	if indexable {
		c.constantsIndex[hash] = p
	}
	return p
}

func (c *compiler) patchJump(placeholder int) {
	offset := len(c.bytecode) - placeholder
	c.arguments[placeholder-1] = offset
}

func (c *compiler) calcBackwardJump(to int) int {
	return len(c.bytecode) + 1 - to
}

func (c *compiler) compile(node ast.Node) {
	c.nodes = append(c.nodes, node)
	defer func() {
		c.nodes = c.nodes[:len(c.nodes)-1]
	}()

	if c.config != nil && c.config.Profile {
		span := &Span{
			Name:       reflect.TypeOf(node).String(),
			Expression: node.String(),
		}
		if len(c.spans) > 0 {
			prev := c.spans[len(c.spans)-1]
			prev.Children = append(prev.Children, span)
		}
		c.spans = append(c.spans, span)
		defer func() {
			if len(c.spans) > 1 {
				c.spans = c.spans[:len(c.spans)-1]
			}
		}()

		c.emit(OpProfileStart, c.addConstant(span))
		defer func() {
			c.emit(OpProfileEnd, c.addConstant(span))
		}()
	}

	switch n := node.(type) {
	case *ast.NilNode:
		c.NilNode(n)
	case *ast.IdentifierNode:
		c.IdentifierNode(n)
	case *ast.IntegerNode:
		c.IntegerNode(n)
	case *ast.FloatNode:
		c.FloatNode(n)
	case *ast.BoolNode:
		c.BoolNode(n)
	case *ast.StringNode:
		c.StringNode(n)
	case *ast.MemberNode:
		c.MemberNode(n)
	case *ast.CallNode:
		c.CallNode(n)
	default:
		panic(fmt.Sprintf("undefined node type (%T)", node))
	}
}

func (c *compiler) NilNode(_ *ast.NilNode) {
	c.emit(OpNil)
}

func (c *compiler) IdentifierNode(node *ast.IdentifierNode) {
	if index, ok := c.lookupVariable(node.Value); ok {
		c.emit(OpLoadVar, index)
		return
	}
	if node.Value == "$env" {
		c.emit(OpLoadEnv)
		return
	}

	var env Nature
	if c.config != nil {
		env = c.config.Env
	}

	if env.IsFastMap() {
		c.emit(OpLoadFast, c.addConstant(node.Value))
	} else if ok, index, name := checker.FieldIndex(env, node); ok {
		c.emit(OpLoadField, c.addConstant(&runtime.Field{
			Index: index,
			Path:  []string{name},
		}))
	} else if ok, index, name := checker.MethodIndex(env, node); ok {
		c.emit(OpLoadMethod, c.addConstant(&runtime.Method{
			Name:  name,
			Index: index,
		}))
	} else {
		c.emit(OpLoadConst, c.addConstant(node.Value))
	}
}

func (c *compiler) IntegerNode(node *ast.IntegerNode) {
	t := node.Type()
	if t == nil {
		c.emitPush(node.Value)
		return
	}
	switch t.Kind() {
	case reflect.Float32:
		c.emitPush(float32(node.Value))
	case reflect.Float64:
		c.emitPush(float64(node.Value))
	case reflect.Int:
		c.emitPush(node.Value)
	case reflect.Int8:
		if node.Value > math.MaxInt8 || node.Value < math.MinInt8 {
			panic(fmt.Sprintf("constant %d overflows int8", node.Value))
		}
		c.emitPush(int8(node.Value))
	case reflect.Int16:
		if node.Value > math.MaxInt16 || node.Value < math.MinInt16 {
			panic(fmt.Sprintf("constant %d overflows int16", node.Value))
		}
		c.emitPush(int16(node.Value))
	case reflect.Int32:
		if node.Value > math.MaxInt32 || node.Value < math.MinInt32 {
			panic(fmt.Sprintf("constant %d overflows int32", node.Value))
		}
		c.emitPush(int32(node.Value))
	case reflect.Int64:
		c.emitPush(int64(node.Value))
	case reflect.Uint:
		if node.Value < 0 {
			panic(fmt.Sprintf("constant %d overflows uint", node.Value))
		}
		c.emitPush(uint(node.Value))
	case reflect.Uint8:
		if node.Value > math.MaxUint8 || node.Value < 0 {
			panic(fmt.Sprintf("constant %d overflows uint8", node.Value))
		}
		c.emitPush(uint8(node.Value))
	case reflect.Uint16:
		if node.Value > math.MaxUint16 || node.Value < 0 {
			panic(fmt.Sprintf("constant %d overflows uint16", node.Value))
		}
		c.emitPush(uint16(node.Value))
	case reflect.Uint32:
		if node.Value < 0 {
			panic(fmt.Sprintf("constant %d overflows uint32", node.Value))
		}
		c.emitPush(uint32(node.Value))
	case reflect.Uint64:
		if node.Value < 0 {
			panic(fmt.Sprintf("constant %d overflows uint64", node.Value))
		}
		c.emitPush(uint64(node.Value))
	default:
		c.emitPush(node.Value)
	}
}

func (c *compiler) FloatNode(node *ast.FloatNode) {
	switch node.Type().Kind() {
	case reflect.Float32:
		c.emitPush(float32(node.Value))
	case reflect.Float64:
		c.emitPush(node.Value)
	default:
		c.emitPush(node.Value)
	}
}

func (c *compiler) BoolNode(node *ast.BoolNode) {
	if node.Value {
		c.emit(OpTrue)
	} else {
		c.emit(OpFalse)
	}
}

func (c *compiler) StringNode(node *ast.StringNode) {
	c.emitPush(node.Value)
}

func (c *compiler) MemberNode(node *ast.MemberNode) {
	var env Nature
	if c.config != nil {
		env = c.config.Env
	}

	if ok, index, name := checker.MethodIndex(env, node); ok {
		c.compile(node.Node)
		c.emit(OpMethod, c.addConstant(&runtime.Method{
			Name:  name,
			Index: index,
		}))
		return
	}
	op := OpFetch
	base := node.Node

	ok, index, nodeName := checker.FieldIndex(env, node)
	path := []string{nodeName}

	if ok {
		op = OpFetchField
		for {
			if ident, isIdent := base.(*ast.IdentifierNode); isIdent {
				if ok, identIndex, name := checker.FieldIndex(env, ident); ok {
					index = append(identIndex, index...)
					path = append([]string{name}, path...)
					c.emitLocation(ident.Location(), OpLoadField, c.addConstant(
						&runtime.Field{Index: index, Path: path},
					))
					return
				}
			}

			if member, isMember := base.(*ast.MemberNode); isMember {
				if ok, memberIndex, name := checker.FieldIndex(env, member); ok {
					index = append(memberIndex, index...)
					path = append([]string{name}, path...)
					node = member
					base = member.Node
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	c.compile(base)

	if op == OpFetch {
		c.compile(node.Property)
		c.emit(OpFetch)
	} else {
		c.emitLocation(node.Location(), op, c.addConstant(
			&runtime.Field{Index: index, Path: path},
		))
	}
}

func (c *compiler) CallNode(node *ast.CallNode) {
	fn := node.Callee.Type()
	if fn.Kind() == reflect.Func {
		fnInOffset := 0
		fnNumIn := fn.NumIn()
		switch callee := node.Callee.(type) {
		case *ast.MemberNode:
			if prop, ok := callee.Property.(*ast.StringNode); ok {
				if _, ok = callee.Node.Type().MethodByName(prop.Value); ok && callee.Node.Type().Kind() != reflect.Interface {
					fnInOffset = 1
					fnNumIn--
				}
			}
		case *ast.IdentifierNode:
			if t, ok := c.config.Env.MethodByName(callee.Value); ok && t.Method {
				fnInOffset = 1
				fnNumIn--
			}
		}
		for i, arg := range node.Arguments {
			c.compile(arg)
			if k := kind(arg.Type()); k == reflect.Ptr || k == reflect.Interface {
				var in reflect.Type
				if fn.IsVariadic() && i >= fnNumIn-1 {
					in = fn.In(fn.NumIn() - 1).Elem()
				} else {
					in = fn.In(i + fnInOffset)
				}
				if k = kind(in); k != reflect.Ptr && k != reflect.Interface {
					c.emit(OpDeref)
				}
			}
		}
	} else {
		for _, arg := range node.Arguments {
			c.compile(arg)
		}
	}

	c.compile(node.Callee)

	isMethod, _, _ := checker.MethodIndex(c.config.Env, node.Callee)
	if index, ok := checker.TypedFuncIndex(node.Callee.Type(), isMethod); ok {
		c.emit(OpCallTyped, index)
		return
	} else if checker.IsFastFunc(node.Callee.Type(), isMethod) {
		c.emit(OpCallFast, len(node.Arguments))
	} else {
		c.emit(OpCall, len(node.Arguments))
	}
}

func (c *compiler) lookupVariable(name string) (int, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if c.scopes[i].variableName == name {
			return c.scopes[i].index, true
		}
	}
	return 0, false
}

func kind(t reflect.Type) reflect.Kind {
	if t == nil {
		return reflect.Invalid
	}
	return t.Kind()
}
