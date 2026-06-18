package isa

import "fmt"

type Opcode byte

const (
	OpNop Opcode = iota

	OpPush
	OpDup
	OpDrop
	OpSwap
	OpOver

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpAnd
	OpOr
	OpXor
	OpShl
	OpShr
	OpInv
	OpEq
	OpLt
	OpGt

	OpLoad
	OpStore

	OpJmp
	OpJz
	OpJnz
	OpJltz
	OpJgtz
	OpJgez
	OpJlez
	OpCall
	OpRet
	OpExecute

	OpEI
	OpDI
	OpIret

	OpHalt

	OpVLoad
	OpVStore

	OpVAdd
	OpVSub
	OpVMul
	OpVDiv
	OpVCmp
)

const (
	VectorRegisters = 4
	VectorLanes     = 4
)

func (op Opcode) String() string {
	switch op {
	case OpNop:
		return "nop"
	case OpPush:
		return "push"
	case OpDup:
		return "dup"
	case OpDrop:
		return "drop"
	case OpSwap:
		return "swap"
	case OpOver:
		return "over"
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpMod:
		return "mod"
	case OpAnd:
		return "and"
	case OpOr:
		return "or"
	case OpXor:
		return "xor"
	case OpShl:
		return "shl"
	case OpShr:
		return "shr"
	case OpInv:
		return "inv"
	case OpEq:
		return "eq"
	case OpLt:
		return "lt"
	case OpGt:
		return "gt"
	case OpLoad:
		return "@"
	case OpStore:
		return "!"
	case OpJmp:
		return "jmp"
	case OpJz:
		return "jz"
	case OpJnz:
		return "jnz"
	case OpJltz:
		return "jltz"
	case OpJgtz:
		return "jgtz"
	case OpJgez:
		return "jgez"
	case OpJlez:
		return "jlez"
	case OpCall:
		return "call"
	case OpRet:
		return "ret"
	case OpExecute:
		return "execute"
	case OpEI:
		return "ei"
	case OpDI:
		return "di"
	case OpIret:
		return "iret"
	case OpHalt:
		return "halt"
	case OpVLoad:
		return "vload"
	case OpVStore:
		return "vstore"
	case OpVAdd:
		return "vadd"
	case OpVSub:
		return "vsub"
	case OpVMul:
		return "vmul"
	case OpVDiv:
		return "vdiv"
	case OpVCmp:
		return "vcmp"
	default:
		return fmt.Sprintf("unknown(%d)", byte(op))
	}
}

func LengthAt(code []byte, pc uint32) (int, error) {
	if int(pc) >= len(code) {
		return 0, fmt.Errorf("pc out of range: %d", pc)
	}
	op := Opcode(code[pc])
	switch op {
	case OpPush, OpJmp, OpJz, OpJnz, OpJltz, OpJgtz, OpJgez, OpJlez, OpCall:
		return 5, nil
	case OpVLoad, OpVStore:
		return 6, nil
	case OpVAdd, OpVSub, OpVMul, OpVDiv, OpVCmp:
		return 4, nil
	default:
		return 1, nil
	}
}

func Ticks(op Opcode) uint64 {
	switch op {
	case OpVLoad, OpVStore:
		return 6
	case OpVAdd, OpVSub, OpVMul, OpVDiv, OpVCmp:
		return 4
	case OpLoad, OpStore:
		return 3
	case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpAnd, OpOr, OpXor, OpShl, OpShr, OpEq, OpLt, OpGt:
		return 2
	case OpJmp, OpJz, OpJnz, OpJltz, OpJgtz, OpJgez, OpJlez, OpCall, OpRet, OpExecute:
		return 2
	case OpPush:
		return 2
	case OpDup, OpDrop, OpSwap, OpOver, OpInv:
		return 1
	case OpEI, OpDI, OpIret:
		return 1
	case OpHalt:
		return 1
	default:
		return 1
	}
}
