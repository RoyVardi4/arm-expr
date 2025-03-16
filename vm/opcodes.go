package vm

type Opcode byte

const (
	OpInvalid Opcode = iota
	OpPush
	OpLoadConst
	OpLoadField
	OpLoadFast
	OpLoadMethod
	OpFetch
	OpFetchField
	OpMethod
	OpTrue
	OpFalse
	OpNil
	OpCall
	OpCallFast
	OpCallTyped
	OpCast
	OpDeref
	OpProfileStart
	OpProfileEnd
	OpBegin
	OpEnd // This opcode must be at the end of this list.
)
