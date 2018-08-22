package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/pivotal-cf/brokerapi"
	"go.uber.org/zap"

	"github.com/cloudfoundry-incubator/blockhead/pkg/broker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/cloudfoundry-incubator/blockhead/pkg/util/logging"
)

func main() {
	zlogger, _ := zap.NewProduction()
	defer zlogger.Sync() // flushes buffer on exit
	zlogger.Info("starting blockhead broker")

	lag := logging.NewLagerAdapter(zlogger)

	if len(os.Args) < 2 {
		lag.Fatal("main", errors.New("config file is missing"))
	}

	configFilepath := os.Args[1]
	cfg, err := config.NewConfig(configFilepath)
	if err != nil {
		lag.Fatal("main", err)
	}

	broker := broker.NewBlockheadBroker(*cfg)
	creds := brokerapi.BrokerCredentials{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	brokerAPI := brokerapi.New(broker, lag, creds)

	http.Handle("/", brokerAPI)
	lag.Fatal("http-listen", http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}
