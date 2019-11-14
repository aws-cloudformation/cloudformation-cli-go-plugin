.PHONY: build clean deploy

build:
	cfn generate
	env GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go

build-prod:
	cfn generate
	env GOOS=linux go build -ldflags="-s -w" -tags="logging callback" -o bin/handler cmd/main.go

test:
	cfn generate
	env GOOS=linux go build -tags logging -tags callback -ldflags="-s -w" -o bin/handler cmd/main.go

clean:
	rm -rf bin
