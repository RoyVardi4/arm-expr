package vm

type Opcode byte

const (
	OpInvalid Opcode = iota
	OpPush
	OpInt
	OpLoadVar
	OpLoadConst
	OpLoadField
	OpLoadFast
	OpLoadMethod
	OpLoadFunc
	OpFetch
	OpFetchField
	OpMethod
	OpTrue
	OpFalse
	OpNil
	OpCall
	OpCall0
	OpCall1
	OpCall2
	OpCall3
	OpCallN
	OpCallFast
	OpCallTyped
	OpCast
	OpDeref
	OpProfileStart
	OpProfileEnd
	OpBegin
	OpEnd // This opcode must be at the end of this list.
)
