.PHONY: build clean deploy

build:
	cfn-cli generate
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go

sam:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/handler cmd/main.go
	sam local invoke  --event cmd/test/data/create.request.json
	sam local invoke  --event cmd/test/data/update.request.json

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose