# Blockhead Service Broker
[![Build Status](https://travis-ci.org/cloudfoundry-incubator/blockhead.svg?branch=master)](https://travis-ci.org/cloudfoundry-incubator/blockhead)

 BlockHead is a dedicated service broker developed based on the Open Service Broker (OSB) API that allows for the creation and deployment of smart contracts through creation and binding of services.

## Configure and Run
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

# Blockchain Support
Blockhead currently supports two types of Blockchain networks:

- The Ethereum public blockchain network.
- The permissioned Fabric blockchain network. 

The type of the blockchain network that is offered by Blockhead is determined by setting the `Tag` in the service definition provided to Blockhead. 

For an Ethereum blockchain network, the service definition tags should include `ethereum`. For the Fabric network, the tags in the service definition should include `fabric`. A service definition lacking either of the two tags is considered incompatible with the broker.

Examples can be found under the `services/` directory in the repository.

