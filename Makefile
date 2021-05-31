SOCKET=-socket ./.test.socket
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
	go run ./ run -mode pebble -log-level trace ${SOCKET} -r .

.PHONY: read
read:
	go run ./ read -debug ${SOCKET}

.PHONY: generate
generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/rpc.proto
