package sim

import (
	"encoding/binary"
	"fmt"
	"mypl/internal/binaryfmt"
	"mypl/internal/isa"
)

const (
	IOOutDataAddr uint32 = 0xFF04
)

type RunResult struct {
	Output       string
	Trace        Trace
	Ticks        uint64
	Instructions uint64
}

type cpu struct {
	mem []byte

	pc uint32
	ds []int32
	rs []uint32

	outputRunes []rune

	tick         uint64
	instructions uint64
	trace        Trace
}

func Run(img binaryfmt.Image, cfg Config) (RunResult, error) {
	if img.MemorySize == 0 {
		img.MemorySize = binaryfmt.DefaultMemSz
	}
	m := make([]byte, img.MemorySize)
	if int(img.CodeBase)+len(img.Code) > len(m) {
		return RunResult{}, fmt.Errorf("code does not fit memory")
	}
	if int(img.DataBase)+len(img.Data) > len(m) {
		return RunResult{}, fmt.Errorf("data does not fit memory")
	}
	copy(m[img.CodeBase:], img.Code)
	copy(m[img.DataBase:], img.Data)

	c := &cpu{
		mem: m,
		pc:  img.EntryPoint,
	}

	for {
		if cfg.MaxTicks > 0 && c.tick >= cfg.MaxTicks {
			return RunResult{}, fmt.Errorf("tick limit reached (%d)", cfg.MaxTicks)
		}
		op, instrLen, instrText, err := c.fetchDecode()
		if err != nil {
			return RunResult{}, err
		}
		startTick := c.tick
		pcBefore := c.pc
		halt, err := c.execute(op, instrLen)
		if err != nil {
			return RunResult{}, fmt.Errorf("pc=%d instr=%s: %w", pcBefore, instrText, err)
		}
		c.tick += isa.Ticks(op)
		c.instructions++

		c.trace.Add(TraceEntry{
			TickStart: startTick,
			TickEnd:   c.tick,
			PC:        pcBefore,
			Instr:     instrText,
			DS:        cloneInt32(c.ds),
			RS:        cloneU32(c.rs),
		})

		if halt {
			break
		}
	}
	return RunResult{
		Output:       string(c.outputRunes),
		Trace:        c.trace,
		Ticks:        c.tick,
		Instructions: c.instructions,
	}, nil
}

func (c *cpu) fetchDecode() (isa.Opcode, int, string, error) {
	opByte, err := c.readByte(c.pc)
	if err != nil {
		return 0, 0, "", err
	}
	op := isa.Opcode(opByte)
	ln := 1
	switch op {
	case isa.OpPush, isa.OpJmp, isa.OpJz, isa.OpJnz, isa.OpJltz, isa.OpJgez, isa.OpJlez, isa.OpCall:
		ln = 5
		//case TODO: потом допилить остальные операции
	}
	if int(c.pc)+ln > len(c.mem) {
		return 0, 0, "", fmt.Errorf("instruction outside memory")
	}
	text := op.String()
	switch op {
	case isa.OpPush, isa.OpJmp, isa.OpJz, isa.OpJnz, isa.OpJltz, isa.OpJgez, isa.OpJlez, isa.OpCall:
		v := int32(binary.LittleEndian.Uint32(c.mem[c.pc+1 : c.pc+5]))
		text = fmt.Sprintf("%s %d", op.String(), v)
		//case TODO: потом допилить остальные операции
	}
	return op, ln, text, nil
}

func (c *cpu) execute(op isa.Opcode, instrLen int) (bool, error) {
	nextPC := c.pc + uint32(instrLen)
	readImm := func(off uint32) int32 {
		return int32(binary.LittleEndian.Uint32(c.mem[off : off+4]))
	}

	switch op {
	case isa.OpNop:
		c.pc = nextPC
	case isa.OpPush:
		c.ds = append(c.ds, readImm(c.pc+1))
		c.pc = nextPC
	case isa.OpDup:
		v, err := c.peekDS()
		if err != nil {
			return false, err
		}
		c.ds = append(c.ds, v)
		c.pc = nextPC
	case isa.OpDrop:
		if _, err := c.popDS(); err != nil {
			return false, err
		}
		c.pc = nextPC
	case isa.OpSwap:
		if len(c.ds) < 2 {
			return false, fmt.Errorf("stack underflow for swap")
		}
		n := len(c.ds)
		c.ds[n-1], c.ds[n-2] = c.ds[n-2], c.ds[n-1]
		c.pc = nextPC
	case isa.OpOver:
		if len(c.ds) < 2 {
			return false, fmt.Errorf("stack underflow for over")
		}
		c.ds = append(c.ds, c.ds[len(c.ds)-2])
		c.pc = nextPC
	case isa.OpAdd:
		return false, c.binOp(func(a, b int32) int32 { return a + b }, nextPC)
	case isa.OpSub:
		return false, c.binOp(func(a, b int32) int32 { return a - b }, nextPC)
	case isa.OpMul:
		return false, c.binOp(func(a, b int32) int32 { return a * b }, nextPC)
	case isa.OpDiv:
		return false, c.binOp(func(a, b int32) int32 {
			if b == 0 {
				return 0
			}
			return a / b
		}, nextPC)
	case isa.OpMod:
		return false, c.binOp(func(a, b int32) int32 {
			if b == 0 {
				return 0
			}
			return a % b
		}, nextPC)
	case isa.OpAnd:
		return false, c.binOp(func(a, b int32) int32 { return a & b }, nextPC)
	case isa.OpOr:
		return false, c.binOp(func(a, b int32) int32 { return a | b }, nextPC)
	case isa.OpXor:
		return false, c.binOp(func(a, b int32) int32 { return a ^ b }, nextPC)
	case isa.OpEq:
		return false, c.binOp(func(a, b int32) int32 {
			if a == b {
				return 1
			}
			return 0
		}, nextPC)
	case isa.OpLt:
		return false, c.binOp(func(a, b int32) int32 {
			if a < b {
				return 1
			}
			return 0
		}, nextPC)
	case isa.OpGt:
		return false, c.binOp(func(a, b int32) int32 {
			if a > b {
				return 1
			}
			return 0
		}, nextPC)
	case isa.OpShl:
		return false, c.binOp(func(a, b int32) int32 { return a << (uint32(b) & 31) }, nextPC)
	case isa.OpShr:
		return false, c.binOp(func(a, b int32) int32 { return a >> (uint32(b) & 31) }, nextPC)
	case isa.OpInv:
		v, err := c.popDS()
		if err != nil {
			return false, err
		}
		c.ds = append(c.ds, ^v)
		c.pc = nextPC
	case isa.OpLoad:
		addr, err := c.popDS()
		if err != nil {
			return false, err
		}
		v, err := c.readWord(uint32(addr))
		if err != nil {
			return false, err
		}
		c.ds = append(c.ds, v)
		c.pc = nextPC
	case isa.OpStore:
		addr, err := c.popDS()
		if err != nil {
			return false, err
		}
		value, err := c.popDS()
		if err != nil {
			return false, err
		}
		if err := c.writeWord(uint32(addr), value); err != nil {
			return false, err
		}
		c.pc = nextPC
	case isa.OpJmp:
		c.pc = uint32(readImm(c.pc + 1))
	case isa.OpJz:
		cond, err := c.popDS()
		if err != nil {
			return false, err
		}
		if cond == 0 {
			c.pc = uint32(readImm(c.pc + 1))
		} else {
			c.pc = nextPC
		}
	case isa.OpJnz:
		cond, err := c.popDS()
		if err != nil {
			return false, err
		}
		if cond != 0 {
			c.pc = uint32(readImm(c.pc + 1))
		} else {
			c.pc = nextPC
		}
	case isa.OpJltz:
		cond, err := c.popDS()
		if err != nil {
			return false, err
		}
		if cond < 0 {
			c.pc = uint32(readImm(c.pc + 1))
		} else {
			c.pc = nextPC
		}
	case isa.OpJgtz:
		cond, err := c.popDS()
		if err != nil {
			return false, err
		}
		if cond > 0 {
			c.pc = uint32(readImm(c.pc + 1))
		} else {
			c.pc = nextPC
		}
	case isa.OpJgez:
		cond, err := c.popDS()
		if err != nil {
			return false, err
		}
		if cond >= 0 {
			c.pc = uint32(readImm(c.pc + 1))
		} else {
			c.pc = nextPC
		}
	case isa.OpJlez:
		cond, err := c.popDS()
		if err != nil {
			return false, err
		}
		if cond <= 0 {
			c.pc = uint32(readImm(c.pc + 1))
		} else {
			c.pc = nextPC
		}
	case isa.OpCall:
		target := uint32(readImm(c.pc + 1))
		c.rs = append(c.rs, nextPC)
		c.pc = target
	case isa.OpRet:
		ret, err := c.popRS()
		if err != nil {
			return false, err
		}
		c.pc = ret
	case isa.OpExecute:
		addr, err := c.popDS()
		if err != nil {
			return false, err
		}
		c.rs = append(c.rs, nextPC)
		c.pc = uint32(addr)
	case isa.OpHalt:
		c.pc = nextPC
		return true, nil
	default:
		return false, fmt.Errorf("unknown opcode %d", op)
	}
	return false, nil
}

func (c *cpu) binOp(fn func(a, b int32) int32, nextPC uint32) error {
	b, err := c.popDS()
	if err != nil {
		return err
	}
	a, err := c.popDS()
	if err != nil {
		return err
	}
	c.ds = append(c.ds, fn(a, b))
	c.pc = nextPC
	return nil
}

func (c *cpu) popDS() (int32, error) {
	if len(c.ds) == 0 {
		return 0, fmt.Errorf("data stack underflow")
	}
	v := c.ds[len(c.ds)-1]
	c.ds = c.ds[:len(c.ds)-1]
	return v, nil
}

func (c *cpu) peekDS() (int32, error) {
	if len(c.ds) == 0 {
		return 0, fmt.Errorf("data stack is empty")
	}
	return c.ds[len(c.ds)-1], nil
}

func (c *cpu) popRS() (uint32, error) {
	if len(c.rs) == 0 {
		return 0, fmt.Errorf("return stack underflow")
	}
	v := c.rs[len(c.rs)-1]
	c.rs = c.rs[:len(c.rs)-1]
	return v, nil
}

func (c *cpu) readByte(addr uint32) (byte, error) {
	if int(addr) >= len(c.mem) {
		return 0, fmt.Errorf("read byte out of memory at %d", addr)
	}
	return c.mem[addr], nil
}

func (c *cpu) readWord(addr uint32) (int32, error) {
	if int(addr)+4 > len(c.mem) {
		return 0, fmt.Errorf("read word out of memory at %d", addr)
	}
	return int32(binary.LittleEndian.Uint32(c.mem[addr : addr+4])), nil
}

func (c *cpu) writeWord(addr uint32, v int32) error {
	if addr == IOOutDataAddr {
		c.outputRunes = append(c.outputRunes, rune(v&0xFF))
		return nil
	}
	if int(addr)+4 > len(c.mem) {
		return fmt.Errorf("write word out of memory at %d", addr)
	}
	binary.LittleEndian.PutUint32(c.mem[addr:addr+4], uint32(v))
	return nil
}

func cloneInt32(s []int32) []int32 {
	return append([]int32(nil), s...)
}

func cloneU32(s []uint32) []uint32 {
	return append([]uint32(nil), s...)
}
