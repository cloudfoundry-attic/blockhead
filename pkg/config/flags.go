package config

import (
	"flag"
	"strings"
)

type ServiceFlags []string

func (s *ServiceFlags) String() string {
	return strings.Join(*s, ",")
}

func (s *ServiceFlags) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var (
	ServicePaths ServiceFlags
	ConfigPath   string
)

func init() {
	flag.Var(&ServicePaths, "service", "Service config file.")
	flag.StringVar(&ConfigPath, "config", "config.json", "Broker config file.")
}
