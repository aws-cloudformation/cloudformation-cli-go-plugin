.PHONY: build clean deploy

build:
	cfn-cli generate
	env GOOS=linux go build -ldflags="-s -w" -o .bin/handler cmd/main.go
	zip .bin/handler.zip .bin/handler


clean:
	rm -rf .bin

