all: test build
test:
	go test -v -race
build: fmt
	CGO_ENABLED=0 go build -v .
fmt:
	go fmt .
run:
	go run . 8080
