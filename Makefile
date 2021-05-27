.PHONY: build
build: test
	go build ./

.PHONY: test
test:
	go test ./...

.PHONY: bench
bench:
	go test ./... -bench=.

.PHONY: install
install: 
	go install ./

.PHONY: run
run:
	go run ./ run -mode pebble
