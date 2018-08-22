package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/brokerapi"
)

type State struct {
	Config   Config
	Services []brokerapi.Service
}

type Config struct {
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
	Port     uint16 `json:"port"`
}

func NewState(configPath string, servicePath string) (*State, error) {
	config := &Config{}
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening config file: %v", err.Error())
	}

	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v", err.Error())
	}

	if config.Port == 0 {
		config.Port = 3333
	}

	if servicePath == "" {
		return nil, fmt.Errorf("Service Directory Missing")
	}

	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%v: %s", err, servicePath)
	}

	files, err := ioutil.ReadDir(servicePath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("service directory is empty: %s", servicePath)
	}

	state := &State{
		Config:   *config,
		Services: []brokerapi.Service{},
	}

	for _, file := range files {
		serviceFilePath := filepath.Join(servicePath, file.Name())
		bytes, err = ioutil.ReadFile(serviceFilePath)
		if err != nil {
			return nil, fmt.Errorf("Error opening config file: %v", err.Error())
		}

		service := brokerapi.Service{}
		err = json.Unmarshal(bytes, &service)
		if err != nil {
			return nil, fmt.Errorf("Error parsing service file: %s - %v", serviceFilePath, err.Error())
		}

		state.Services = append(state.Services, service)
	}

	return state, nil
}
