CHAMPI_SEED=ABCDEFG42

all: test server

server: $(wildcard *.go)
	CGO_ENABLED=0 go build -o server -v .

.PHONY: test
test: server.go server_test.go
	go test -v -race

.PHONY: fmt
fmt:
	go fmt .

.PHONY: run
run:
	@CHAMPI_SEED=$(CHAMPI_SEED) go run . :8080
