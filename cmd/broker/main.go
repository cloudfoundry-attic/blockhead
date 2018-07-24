package main

import (
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"

	"github.com/jberkhahn/blockhead/pkg/broker"
	"github.com/jberkhahn/blockhead/pkg/config"
)

func main() {
	logger := lager.NewLogger("blockhead-broker")
	configFilepath := os.Args[1]
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.Debug(fmt.Sprintf("reading config file %q", configFilepath))
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
