.PHONY: build
build: test
	go build ./

test:
	go test ./...

.PHONY: install
install: 
	go install ./

.PHONY: run
run:
	go run ./ run -mode pebble
