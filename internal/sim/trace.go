package sim

import (
	"fmt"
	"strings"
)

type TraceEntry struct {
	TickStart uint64
	TickEnd   uint64
	PC        uint32
	Instr     string
	DS        []int32
	RS        []uint32
	InIRQ     bool
	Event     string
}

type Trace struct {
	Entries []TraceEntry
}

func (t *Trace) Add(e TraceEntry) {
	t.Entries = append(t.Entries, e)
}

func (t Trace) String(limit int) string {
	var b strings.Builder
	start := 0
	if limit > 0 && len(t.Entries) > limit {
		start = len(t.Entries) - limit
	}
	for _, e := range t.Entries[start:] {
		fmt.Fprintf(&b, "[%06d..%06d] pc=%05d irq=%t instr=%-24s ds=%v rs=%v",
			e.TickStart, e.TickEnd, e.PC, e.InIRQ, e.Instr, e.DS, e.RS)
		if e.Event != "" {
			fmt.Fprintf(&b, " events=%s", e.Event)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
