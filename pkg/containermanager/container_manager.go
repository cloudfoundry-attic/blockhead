package containermanager

import (
	"context"
)

type ContainerConfig struct {
	Name         string
	Image        string
	ExposedPorts []string
}

type BindConfig struct {
	InstanceId string
	BindingId  string
}

type Binding struct {
	Port string
}

// internal port to ip:externalport
type ContainerInfo struct {
	ExternalAddress string
	// used for internal communication of broker to container
	InternalAddress string
	// Bindings is a map of container ports to exposed host & port
	Bindings map[string][]Binding
}

//go:generate counterfeiter -o ../fakes/fake_container_manager.go . ContainerManager
type ContainerManager interface {
	Provision(ctx context.Context, cc ContainerConfig) error
	Deprovision(ctx context.Context, instanceID string) error
	Bind(cts context.Context, bc BindConfig) (*ContainerInfo, error)
}
