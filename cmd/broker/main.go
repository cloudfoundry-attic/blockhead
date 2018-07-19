package main

import (
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/jberkhahn/blockhead/pkg/broker"
	"github.com/jberkhahn/blockhead/pkg/config"
	"github.com/pivotal-cf/brokerapi"
)

func main() {
	configFilepath := os.Args[1]
	cfg, err := config.NewConfig(configFilepath)
	if err != nil {
		panic(err)
	}

	broker := broker.BlockheadBroker{}
	logger := lager.NewLogger("blockhead-broker")
	creds := brokerapi.BrokerCredentials{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	brokerAPI := brokerapi.New(broker, logger, creds)

	http.Handle("/", brokerAPI)
	logger.Fatal("http-listen", http.ListenAndServe("localhost:3333", nil))
}
