.PHONY: build clean deploy

build:
	dep ensure -v
	

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose
