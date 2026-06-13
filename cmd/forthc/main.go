package main

import (
	"flag"
	"fmt"
	"mypl/internal/binaryfmt"
	"mypl/internal/compiler"
	"os"
)

func main() {
	var (
		srcPath  = flag.String("src", "", "path to source .fth")
		outPath  = flag.String("out", "", "path to output .bin")
		listPath = flag.String("list", "", "optional path to debug listing")
		codeBase = flag.Uint("code-base", uint(compiler.DefaultCodeBase), "code base address")
		dataBase = flag.Uint("data-base", uint(compiler.DefaultDataBase), "data base address")
		memorySz = flag.Uint("mem", uint(binaryfmt.DefaultMemSz), "memory size in bytes")
	)

	flag.Parse()
	if *srcPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: forthc -src program.fth -out program.bin [-list program.lst]")
		os.Exit(2)
	}

	src, err := os.ReadFile(*srcPath)
	if err != nil {
		die(err)
	}

	res, err := compiler.Compile(string(src), compiler.Options{
		CodeBase:   uint32(*codeBase),
		DataBase:   uint32(*dataBase),
		MemorySize: uint32(*memorySz),
	})
	if err != nil {
		die(err)
	}

	if err := binaryfmt.WriteFile(*outPath, res.Image); err != nil {
		die(err)
	}

	if *listPath != "" {
		if err := os.WriteFile(*listPath, []byte(res.Listing), 0o644); err != nil {
			die(err)
		}
	}

	fmt.Printf("compiled: %s\n", *outPath)
	fmt.Printf("entry=%d input_handler=%d code=%dB data=%dB\n",
		res.Image.EntryPoint, res.Image.InputHandlerAddr, len(res.Image.Code), len(res.Image.Data))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
