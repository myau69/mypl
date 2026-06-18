package sim_test

import (
	"mypl/internal/compiler"
	"mypl/internal/sim"
	"testing"
)

func TestHelloOutput(t *testing.T) {
	src := `
const IO_OUT 65284

: main
	72 IO_OUT !
	73 IO_OUT !
	halt
;
`
	res, err := compiler.Compile(src, compiler.Options{})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	run, err := sim.Run(res.Image, sim.Config{MaxTicks: 1000})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if run.Output != "HI" {
		t.Fatalf("unexpected output: %q", run.Output)
	}
}

func TestExecutionToken(t *testing.T) {
	src := `
const IO_OUT 65284

: emit
	65 IO_OUT !
;

: main
	' emitA
	execute
	halt
;
`
	res, err := compiler.Compile(src, compiler.Options{})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	run, err := sim.Run(res.Image, sim.Config{MaxTicks: 1000})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if run.Output != "A" {
		t.Fatalf("unexpected output: %q", run.Output)
	}
}
