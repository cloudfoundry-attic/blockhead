package docker

import (
	"context"
	"io/ioutil"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerClient interface {
	client.ImageAPIClient
	client.ContainerAPIClient
}

type dockerContainerManager struct {
	client DockerClient
	logger lager.Logger
}

func NewDockerContainerManager(logger lager.Logger, client DockerClient) containermanager.ContainerManager {
	return dockerContainerManager{
		client: client,
		logger: logger.Session("docker-container-manager"),
	}
}

func (dc dockerContainerManager) Provision(ctx context.Context, containerConfig *containermanager.ContainerConfig) error {
	// do an image pull before provisioning to ensure we have a fresh
	// version of the image as it may have changed in the background
	reader, err := dc.client.ImagePull(ctx, containerConfig.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	// Waiting on reader to exit, which signifies image has been pulled
	if reader != nil {
		_, err = ioutil.ReadAll(reader)
	}
	if err != nil {
		return err
	}

	config, err := createContainerConfig(containerConfig)
	if err != nil {
		return err
	}

	hostConfig, err := createHostConfig()
	if err != nil {
		return err
	}

	netConfig, err := createNetworkConfig()
	if err != nil {
		return err
	}

	name := containerConfig.Name
	createdContainer, err := dc.client.ContainerCreate(ctx, config, hostConfig, netConfig, name)
	if err != nil {
		dc.logger.Error("container-creation-failed", err)
		return err

	}
	dc.logger.Info("container-created", lager.Data{"id": createdContainer.ID, "name": name})

	err = dc.client.ContainerStart(ctx, name, types.ContainerStartOptions{})
	if err != nil {
		dc.logger.Error("container-start-failed", err)
		return err

	}
	dc.logger.Info("container-started", lager.Data{"name": name})

	return nil
}

func createContainerConfig(containerConfig *containermanager.ContainerConfig) (*container.Config, error) {
	ports, _, err := nat.ParsePortSpecs(containerConfig.ExposedPorts)
	if err != nil {
		return nil, err
	}

	config := &container.Config{
		Hostname:     "",
		Domainname:   "",
		User:         "",
		ExposedPorts: ports,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		OpenStdin:    false,
		StdinOnce:    false,
		// Env
		// Cmd
		// Healthcheck
		// ArgsEscaped
		Image:   containerConfig.Image,
		Volumes: map[string]struct{}{},
		// WorkingDir
		// Entrypoint
		NetworkDisabled: false,
	}

	return config, nil
}

func createHostConfig() (*container.HostConfig, error) {
	hostConfig := &container.HostConfig{
		// Binds:           volume_bindings(guid),
		// Memory:          convert_memory(memory),
		// MemorySwap:      convert_memory(memory_swap),
		// CpuShares:       cpu_shares,
		PublishAllPorts: false,
		Privileged:      false,
	}
	return hostConfig, nil
}

func createNetworkConfig() (*network.NetworkingConfig, error) {
	netConfig := &network.NetworkingConfig{}
	return netConfig, nil
}
