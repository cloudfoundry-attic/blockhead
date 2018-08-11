package config

import (
	"encoding/json"
	"fmt"
	"github.com/pivotal-cf/brokerapi"
	"io/ioutil"

	"github.com/pivotal-cf/brokerapi"
)

type Config struct {
	Password string            `json:"password,omitempty"`
	Username string            `json:"username,omitempty"`
	Service  brokerapi.Service `json:"service"`
}

func NewConfig(filepath string) (*Config, error) {
	config := &Config{}
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Error opening config file: %v", err.Error())
	}

	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v", err.Error())
	}
	return config, nil
}
