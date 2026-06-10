.PHONY: test fmt

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go')

	