.PHONY: test fmt compile simulate run-hello run-cat run-prob1 run-vector run-hello-username

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

run-prob1:
	mkdir -p build
	go run ./cmd/forthc -src examples/prob1_scalar.fth -out build/prob1.bin -list build/prob1.lst
	go run ./cmd/forthsim -bin build/prob1.bin -config configs/prob1_scalar.json -trace build/prob1.trace.log

run-vector:
	mkdir -p build
	go run ./cmd/forthc -src examples/vector_demo.fth -out build/vector.bin -list build/vector.lst
	go run ./cmd/forthsim -bin build/vector.bin -config configs/vector_demo.json -trace build/vector.trace.log

run-hello-username:
	mkdir -p build
	go run ./cmd/forthc -src examples/hello_username.fth -out build/hello_username.bin -list build/hello_username.lst
	go run ./cmd/forthsim -bin build/hello_username.bin -config configs/hello_username.json -trace build/hello_username.trace.log
