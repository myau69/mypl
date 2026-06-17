.PHONY: test fmt compile simulate run-hello run-cat

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go')

compile:
	go run ./cmd/forthc -src examples/execution_token.fth -out build/xt.bin -list build/xt.lst

simulate:
	go run ./cmd/forthsim -bin build/xt.bin -trace build/xt.trace.log

run-hello:
	mkdir -p build
	go run ./cmd/forthc -src examples/hello.fth -out build/hello.bin -list build/hello.lst
	go run ./cmd/forthsim -bin build/hello.bin -config configs/hello.json -trace build/hello.trace.log

run-cat:
	mkdir -p build
	go run ./cmd/forthc -src examples/cat_trap.fth -out build/cat.bin -list build/cat.lst
	go run ./cmd/forthsim -bin build/cat.bin -config configs/cat_trap.json -trace build/cat.trace.log
