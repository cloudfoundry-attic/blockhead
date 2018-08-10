package main

import (
	"errors"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"

	"github.com/cloudfoundry-incubator/blockhead/pkg/broker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
)

func main() {
	logger := lager.NewLogger("blockhead-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	if len(os.Args) < 2 {
		logger.Fatal("main", errors.New("config file is missing"))
	}

	configFilepath := os.Args[1]
	cfg, err := config.NewConfig(configFilepath)
	if err != nil {
		logger.Fatal("main", err)
	}

	broker := broker.BlockheadBroker{}
	creds := brokerapi.BrokerCredentials{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	brokerAPI := brokerapi.New(broker, logger, creds)

	http.Handle("/", brokerAPI)
	logger.Fatal("http-listen", http.ListenAndServe("0.0.0.0:3333", nil))
}
