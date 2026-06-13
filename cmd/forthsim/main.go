package main

import (
	"flag"
	"fmt"
	"mypl/internal/binaryfmt"
	"mypl/internal/sim"
	"os"
)

func main() {
	var (
		binPath    = flag.String("bin", "", "path to program .bin")
		cfgPath    = flag.String("config", "", "path to simulation config .json")
		tracePath  = flag.String("trace", "", "optional path to trace output")
		traceLimit = flag.Int("trace-limit", 120, "how many last trace records to print to stdout")
	)
	flag.Parse()

	if *binPath == "" {
		fmt.Fprintln(os.Stderr, "usage: forthsim -bin program.bin [-config run.json] [-trace run.log]")
		os.Exit(2)
	}

	img, err := binaryfmt.ReadFile(*binPath)
	if err != nil {
		die(err)
	}

	cfg, err := sim.LoadConfig(*cfgPath)
	if err != nil {
		die(err)
	}

	res, err := sim.Run(img, cfg)
	if err != nil {
		die(err)
	}

	fmt.Printf("output: %q\n", res.Output)
	fmt.Printf("instructions=%d ticks=%d\n", res.Instructions, res.Ticks)
	fmt.Printf("---- trace ----")
	fmt.Printf(res.Trace.String(*traceLimit))

	if *tracePath != "" {
		if err := os.WriteFile(*tracePath, []byte(res.Trace.String(0)), 0o644); err != nil {
			die(err)
		}
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
