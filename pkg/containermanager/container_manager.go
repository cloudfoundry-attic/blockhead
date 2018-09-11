package containermanager

import (
	"context"
)

type ContainerConfig struct {
	Name         string
	Image        string
	ExposedPorts []string
}

//go:generate counterfeiter -o fakes/fake_container_manager.go . ContainerManager
type ContainerManager interface {
	Provision(ctx context.Context, cc *ContainerConfig) error
}
