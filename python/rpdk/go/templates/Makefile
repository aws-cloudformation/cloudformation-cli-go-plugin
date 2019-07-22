.PHONY: build clean deploy

build:
	cfn-cli generate
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go

clean:
	rm -rf ./bin 

deploy: clean build
	sls deploy --verbose