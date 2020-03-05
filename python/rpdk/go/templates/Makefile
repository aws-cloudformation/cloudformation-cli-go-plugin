.PHONY: build test clean
include makevars

build:
	cfn generate
	env GOOS=linux go build -ldflags="-s -w" -tags="$(TAGS)" -o bin/handler cmd/main.go

test:
	cfn generate
	env GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go

clean:
	rm -rf bin
