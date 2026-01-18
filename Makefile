.PHONY: test test-sandbox benchmark

test:
	go test ./...

test-sandbox:
	sudo go test -v ./internal/runner/sandbox

benchmark:
	sudo go test -bench=. ./benchmarks