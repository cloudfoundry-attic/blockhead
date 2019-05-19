
all: images build test dockertest

IMAGE_REGISTRY ?= blockheads

.PHONY: images
images:
	docker build -f images/Dockerfile.broker -t $(IMAGE_REGISTRY)/broker .
	docker build -f images/Dockerfile.geth -t $(IMAGE_REGISTRY)/geth .

push:
	docker push $(IMAGE_REGISTRY)/broker
	docker push $(IMAGE_REGISTRY)/geth

build:
	go build -v github.com/cloudfoundry-incubator/blockhead/cmd/broker

test:
	ginkgo -r -race -randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --progress

dockertest:
	docker build -f images/Dockerfile.test -t blockheads/tests .
	docker run -v /var/run/docker.sock:/var/run/docker.sock --network host -it --rm blockheads/tests
