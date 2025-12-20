.PHONY: test test-sandbox benchmark

test:
	go test ./...

test-sandbox:
	sudo go test -v ./internal/runner/sandbox

benchmark:
	go test -bench=. ./benchmarks