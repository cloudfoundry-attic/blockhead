package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pivotal-cf/brokerapi"
)

type Config struct {
	Password string              `json:"password,omitempty"`
	Username string              `json:"username,omitempty"`
	Port     int                 `json:"port"`
	Services []brokerapi.Service `json:"service"`
}

func NewConfig(configPath string, servicePaths ServiceFlags) (*Config, error) {
	config := &Config{}
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening config file: %v", err.Error())
	}

	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v", err.Error())
	}

	for _, servicePath := range servicePaths {
		var service brokerapi.Service

		bytes, err := ioutil.ReadFile(servicePath)
		if err != nil {
			return nil, fmt.Errorf("Error opening service file: %v", err.Error())
		}

		err = json.Unmarshal(bytes, &service)
		if err != nil {
			return nil, fmt.Errorf("Error parsing service file: %v", err.Error())
		}
		config.Services = append(config.Services, service)
	}

	return config, nil
}
