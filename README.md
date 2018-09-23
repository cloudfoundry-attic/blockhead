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

