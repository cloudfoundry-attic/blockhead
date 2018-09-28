# blockhead
[![Build Status](https://travis-ci.org/cloudfoundry-incubator/blockhead.svg?branch=master)](https://travis-ci.org/cloudfoundry-incubator/blockhead)

An OSBAPI-compatible broker written in golang.

Configure the broker, by setting `username` and `password` in `config.json`.

Run the broker by doing:

    go run ./cmd/broker/main.go config.json services

Build the docker images:

    make images

Build the `broker` binary:

    make build

Run the tests:

    make test

Run the tests in a docker container:

    make dockertest

To do all the make steps:

    make



If both the BOSH release and the k8s release are going to be hooked up
to the same service-broker platform, `services/eth.json` needs to be
configured so that the class and plan identities do not collide.
