.PHONY: build clean deploy

build:
	cfn-cli generate
	env GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go

test:
	cfn-cli generate
	env GOOS=linux go build -tags logging -tags callback -ldflags="-s -w" -o bin/handler cmd/main.go

clean:
	rm -rf bin
