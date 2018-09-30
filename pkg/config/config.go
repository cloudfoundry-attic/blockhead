package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pborman/uuid"
)

type State struct {
	Config   Config
	Services map[string]*Service
}

type Config struct {
	Password         string `json:"password,omitempty"`
	Username         string `json:"username,omitempty"`
	Port             uint16 `json:"port"`
	ContainerManager string `json:"container-manager,omitempty"`
	DeployerPath     string `json:"deployer_path"`
}

type Service struct {
	Name        string
	Description string
	DisplayName string
	Tags        []string
	Plans       map[string]*Plan
}

type Plan struct {
	Name        string   `json:"name"`
	Image       string   `json:"image"`
	Ports       []string `json:"ports"`
	Description string   `json:"description"`
}

type ServiceDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DisplayName string   `json:"display_name"`
	Tags        []string `json:"tags"`
	Plans       []Plan   `json:"plans"`
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
		Services: map[string]*Service{},
	}

	for _, file := range files {
		serviceFilePath := filepath.Join(servicePath, file.Name())
		bytes, err = ioutil.ReadFile(serviceFilePath)
		if err != nil {
			return nil, fmt.Errorf("Error opening config file: %v", err.Error())
		}

		serviceDef := ServiceDefinition{}
		err = json.Unmarshal(bytes, &serviceDef)
		if err != nil {
			return nil, fmt.Errorf("Error parsing service file: %s - %v", serviceFilePath, err.Error())
		}

		service := Service{
			Name:        serviceDef.Name,
			Description: serviceDef.Description,
			DisplayName: serviceDef.DisplayName,
			Tags:        serviceDef.Tags,
			Plans:       make(map[string]*Plan),
		}

		serviceID := uuid.New()

		for _, plan := range serviceDef.Plans {
			id := uuid.New()
			service.Plans[id] = &plan
		}

		state.Services[serviceID] = &service
	}

	return state, nil
}
