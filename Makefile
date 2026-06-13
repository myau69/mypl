.PHONY: test fmt

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go')

compile:
	go run ./cmd/forthc -src examples/execution_token.fth -out build/xt.bin -list build/xt.lst

simulate:
	go run ./cmd/forthsim -bin build/xt.bin -trace build/xt.trace.log