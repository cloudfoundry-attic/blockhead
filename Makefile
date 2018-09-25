
all: images build dockertest

.PHONY: images
images:
	docker build -f images/Dockerfile.broker -t blockheads/broker .
	docker build -f images/Dockerfile.geth -t blockheads/geth .

build:
	go build -v github.com/cloudfoundry-incubator/blockhead/cmd/broker

test:
	go test -v ./...

dockertest:
	docker build -f images/Dockerfile.test -t blockheads/tests .
	docker run -v /var/run/docker.sock:/var/run/docker.sock -it --rm blockheads/tests
