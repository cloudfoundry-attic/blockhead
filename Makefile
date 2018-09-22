
all: images build test

.PHONY: images
images:
	docker build -f images/Dockerfile.broker .
	docker build -f images/Dockerfile.geth .

build:
	go build -v github.com/cloudfoundry-incubator/blockhead/cmd/broker

test:
	go test -v ./...
