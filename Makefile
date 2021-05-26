TAGS=""

.PHONY: build
build: test
	go build ${TAGS} ./

test:
	go test ${TAGS} ./...

.PHONY: install
install: 
	go install ${TAGS} ./

.PHONY: run
run:
	go run ${TAGS} ./ run
