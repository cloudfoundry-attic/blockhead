package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type BlockheadConfig struct {
	Password string `yaml:"password,omitempty"`
	Username string `yaml:"username,omitempty"`
}

func NewConfig(filepath string) (*BlockheadConfig, error) {
	config := &BlockheadConfig{}
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
