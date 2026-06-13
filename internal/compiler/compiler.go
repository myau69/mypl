package compiler

import (
	"encoding/binary"
	"fmt"
	"mypl/internal/binaryfmt"
	"mypl/internal/isa"
	"strconv"
	"strings"
)

const (
	DefaultCodeBase = uint32(0x0200)
	DefaultDataBase = uint32(0x4000)
)

type Options struct {
	CodeBase   uint32
	DataBase   uint32
	MemorySize uint32
}

type Result struct {
	Image   binaryfmt.Image
	Listing string
	Symbols map[string]uint32
}

type fixup struct {
	offset int
	symbol string
	line   int
}

type emitter struct {
	code   []byte
	fixups []fixup
}

func (e *emitter) emit(op isa.Opcode) {
	e.code = append(e.code, byte(op))
}

func (e *emitter) emitImm(op isa.Opcode, v int32) {
	e.code = append(e.code, byte(op))
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	e.code = append(e.code, b...)
}

func (e *emitter) emitImmFix(op isa.Opcode, symbol string, line int) {
	e.code = append(e.code, byte(op))
	off := len(e.code)
	e.code = append(e.code, 0, 0, 0, 0)
	e.fixups = append(e.fixups, fixup{
		offset: off,
		symbol: symbol,
		line:   line,
	})
}

func Compile(src string, opts Options) (Result, error) {
	if opts.CodeBase == 0 {
		opts.CodeBase = DefaultCodeBase
	}
	if opts.DataBase == 0 {
		opts.DataBase = DefaultDataBase
	}
	if opts.MemorySize == 0 {
		opts.MemorySize = binaryfmt.DefaultMemSz
	}

	tokens, err := Tokenize(src)
	if err != nil {
		return Result{}, err
	}
	symbols := map[string]uint32{}
	var emit emitter
	inProc := false
	firstProc := ""

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		t := tok.Text

		if t != ":" && strings.HasSuffix(t, ":") {
			name := strings.TrimSuffix(t, ":")
			if name == "" {
				return Result{}, fmt.Errorf("line: %d empty label", tok.Line)
			}
			if _, exists := symbols[name]; exists {
				return Result{}, fmt.Errorf("line %d: duplicate label %q", tok.Line, name)
			}
			symbols[name] = opts.CodeBase + uint32(len(emit.code))
			continue
		}

		switch t {
		case ":":
			if inProc {
				return Result{}, fmt.Errorf("line %d: nested procedure definition is not supported", tok.Line)
			}
			if i+1 >= len(tokens) {
				return Result{}, fmt.Errorf("line %d: expected procedure name")
			}
			i++
			name := tokens[i].Text
			if _, exists := symbols[name]; exists {
				return Result{}, fmt.Errorf("line %d: duplicate symbol %q", tokens[i].Line, name)
			}
			symbols[name] = opts.CodeBase + uint32(len(emit.code))
			inProc = true
			if firstProc == "" {
				firstProc = name
			}
		case ";":
			if !inProc {
				return Result{}, fmt.Errorf("line %d: '; is allowed only as procedure terminator", tok.Line)
			}
			emit.emit(isa.OpRet)
			inProc = false
		case "lit":
			if i+1 >= len(tokens) {
				return Result{}, fmt.Errorf("line %d: lit requires value", tok.Line)
			}
			i++
			if err := emitPushToken(&emit, tokens[i]); err != nil {
				return Result{}, err
			}
		case "'":
			if i+1 >= len(tokens) {
				return Result{}, fmt.Errorf("line %d: ' requires word name", tok.Line)
			}
			i++
			emit.emitImmFix(isa.OpPush, tokens[i].Text, tokens[i].Line)
		case "jmp", "jz", "jnz", "jltz", "jgtz", "jgez", "jlez", "call":
			if i+1 >= len(tokens) {
				return Result{}, fmt.Errorf("lint %d: %s requires label", tok.Line, t)
			}
			i++
			op := map[string]isa.Opcode{
				"jmp":  isa.OpJmp,
				"jz":   isa.OpJz,
				"jnz":  isa.OpJnz,
				"jltz": isa.OpJltz,
				"jgtz": isa.OpJgtz,
				"jgez": isa.OpJgez,
				"jlez": isa.OpJlez,
				"call": isa.OpCall,
			}[t]
			arg := tokens[i]
			if v, err := parseLiteral(arg.Text); err == nil {
				emit.emitImm(op, v)
			} else {
				emit.emitImmFix(op, arg.Text, arg.Line)
			}
		default:
			if op, ok := builtinNoArg(t); ok {
				emit.emit(op)
				continue
			}
			if err := emitPushToken(&emit, tok); err == nil {
				continue
			}
			emit.emitImmFix(isa.OpCall, t, tok.Line)
		}
	}

	if inProc {
		return Result{}, fmt.Errorf("procedure is not closed with ';'")
	}

	for _, fx := range emit.fixups {
		addr, ok := symbols[fx.symbol]
		if !ok {
			return Result{}, fmt.Errorf("line %d: unresolved symbol %q", fx.line, fx.symbol)
		}
		binary.LittleEndian.PutUint32(emit.code[fx.offset:fx.offset+4], addr)
	}

	entry, ok := symbols["main"]
	if !ok {
		if firstProc == "" {
			return Result{}, fmt.Errorf("no procedures found: define : main ... ;")
		}
		entry = symbols[firstProc]
	}

	img := binaryfmt.Image{
		Version:    binaryfmt.FormatV1,
		MemorySize: opts.MemorySize,
		CodeBase:   opts.CodeBase,
		DataBase:   opts.DataBase,
		EntryPoint: entry,
		Code:       emit.code,
	}

	listing, err := buildListing(img)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Image:   img,
		Listing: listing,
		Symbols: symbols,
	}, nil
}

func builtinNoArg(tok string) (isa.Opcode, bool) {
	m := map[string]isa.Opcode{
		"nop":     isa.OpNop,
		"dup":     isa.OpDup,
		"drop":    isa.OpDrop,
		"swap":    isa.OpSwap,
		"over":    isa.OpOver,
		"+":       isa.OpAdd,
		"-":       isa.OpSub,
		"*":       isa.OpMul,
		"/":       isa.OpDiv,
		"eq":      isa.OpEq,
		"lt":      isa.OpLt,
		"gt":      isa.OpGt,
		"mod":     isa.OpMod,
		"and":     isa.OpAnd,
		"or":      isa.OpOr,
		"xor":     isa.OpXor,
		"shl":     isa.OpShl,
		"shr":     isa.OpShr,
		"inv":     isa.OpInv,
		"@":       isa.OpLoad,
		"!":       isa.OpStore,
		"ret":     isa.OpRet,
		"execute": isa.OpExecute,
		"halt":    isa.OpHalt,
	}
	op, ok := m[tok]
	return op, ok
}

func emitPushToken(e *emitter, tok Token) error {
	if v, err := parseLiteral(tok.Text); err == nil {
		e.emitImm(isa.OpPush, v)
		return nil
	}
	return fmt.Errorf("line %d: %q is not pushable literal", tok.Line, tok.Text)
}

func parseLiteral(tok string) (int32, error) {
	if strings.HasPrefix(tok, "0x") || strings.HasPrefix(tok, "-0x") {
		v, err := strconv.ParseInt(tok, 0, 32)
		return int32(v), err
	}
	if len(tok) == 3 && tok[0] == '\'' && tok[2] == '\'' {
		return int32(tok[1]), nil
	}
	if strings.HasPrefix(tok, "'") && strings.HasSuffix(tok, "'") && len(tok) >= 3 {
		r := []rune(tok[1 : len(tok)-1])
		if len(r) == 1 {
			return int32(r[0]), nil
		}
	}
	v, err := strconv.ParseInt(tok, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func buildListing(img binaryfmt.Image) (string, error) {
	var b strings.Builder
	pc := 0
	code := img.Code
	for pc < len(code) {
		addr := img.CodeBase + uint32(pc)
		op := isa.Opcode(code[pc])
		ln, err := isa.LengthAt(code, uint32(pc))
		if err != nil {
			return "", err
		}
		if pc+ln > len(code) {
			return "", fmt.Errorf("truncated instruction at pc=%d", pc)
		}
		hexPart := strings.ToUpper(fmt.Sprintf("%X", code[pc:pc+ln]))
		mn := op.String()
		switch op {
		case isa.OpPush, isa.OpJmp, isa.OpJz, isa.OpJnz, isa.OpJltz, isa.OpJgtz, isa.OpJgez, isa.OpJlez, isa.OpCall:
			v := int32(binary.LittleEndian.Uint32(code[pc+1 : pc+5]))
			mn = fmt.Sprintf("%s %d", op.String(), v)
		}
		fmt.Fprintf(&b, "%05d - %s - %s\n", addr, hexPart, mn)
		pc += ln
	}
	return b.String(), nil
}
