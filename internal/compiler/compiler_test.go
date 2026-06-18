package compiler_test

import (
	"mypl/internal/compiler"
	"strings"
	"testing"
)

func TestCompileCStrAndListing(t *testing.T) {
	src := `
cstr hello "Hi"
: main
	&hello
halt
`
	res, err := compiler.Compile(src, compiler.Options{})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if len(res.Image.Data) == 0 {
		t.Fatalf("expected data section")
	}
	if !strings.Contains(res.Listing, "push") {
		t.Fatalf("listing does not contain push")
	}
}
