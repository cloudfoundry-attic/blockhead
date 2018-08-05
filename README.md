# blockhead
[![Build Status](https://travis-ci.org/cloudfoundry-incubator/blockhead.svg?branch=master)](https://travis-ci.org/cloudfoundry-incubator/blockhead)

An OSBAPI-compatible broker written in golang.

Configure the broker, by setting `username` and `password` in `creds.json`.

Run the broker by doing:

    go build cmd/broker && ./broker creds.json

OR

    docker build -t blockhead . && docker run blockhead
