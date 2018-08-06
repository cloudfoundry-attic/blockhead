package main

import (
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"

	"github.com/jberkhahn/blockhead/pkg/broker"
	"github.com/jberkhahn/blockhead/pkg/config"
)

func main() {
	logger := lager.NewLogger("blockhead-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	if len(os.Args) < 2 {
		panic("config file missing")
	}

	configFilepath := os.Args[1]
	cfg, err := config.NewConfig(configFilepath)
	if err != nil {
		panic(err)
	}

	broker := broker.BlockheadBroker{}
	creds := brokerapi.BrokerCredentials{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	brokerAPI := brokerapi.New(broker, logger, creds)

	http.Handle("/", brokerAPI)
	logger.Fatal("http-listen", http.ListenAndServe("localhost:3333", nil))
}
