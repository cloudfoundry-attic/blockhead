package docker

import (
	"context"
	"io/ioutil"
	"os"

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
	client          DockerClient
	logger          lager.Logger
	externalAddress string
}

func NewDockerContainerManager(logger lager.Logger, client DockerClient, externalAddress string) containermanager.ContainerManager {
	return dockerContainerManager{
		client:          client,
		logger:          logger.Session("docker-container-manager"),
		externalAddress: externalAddress,
	}
}

func (dc dockerContainerManager) Provision(ctx context.Context, containerConfig containermanager.ContainerConfig) error {
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

func (dc dockerContainerManager) Deprovision(ctx context.Context, instanceID string) error {
	err := dc.client.ContainerStop(ctx, instanceID, nil)
	if err != nil {
		dc.logger.Error("container-stopping-failed", err)
		return err
	}
	err = dc.client.ContainerRemove(ctx, instanceID, types.ContainerRemoveOptions{})
	if err != nil {
		dc.logger.Error("container-removal", err)
		return err
	}
	return nil
}

func (dc dockerContainerManager) Bind(ctx context.Context, bindingConfig containermanager.BindConfig) (*containermanager.ContainerInfo, error) {
	containerInfo, err := dc.client.ContainerInspect(ctx, bindingConfig.InstanceId)
	if err != nil {
		return nil, err
	}

	bindings := make(map[string][]containermanager.Binding)
	for port, dockerBindings := range containerInfo.NetworkSettings.NetworkSettingsBase.Ports {
		containerBindings := []containermanager.Binding{}
		for _, dockerBinding := range dockerBindings {
			containerBindings = append(containerBindings, containermanager.Binding{
				Port: dockerBinding.HostPort,
			})
		}
		bindings[port.Port()] = containerBindings
	}

	dockerServer := os.Getenv("DOCKER_SERVER")
	if dockerServer == "" {
		dockerServer = "127.0.0.1"
	}

	response := containermanager.ContainerInfo{
		InternalAddress: dockerServer,
		ExternalAddress: dc.externalAddress,
		Bindings:        bindings,
	}

	return &response, nil
}

func createContainerConfig(containerConfig containermanager.ContainerConfig) (*container.Config, error) {
	ports, _, err := nat.ParsePortSpecs(containerConfig.ExposedPorts)
	if err != nil {
		return nil, err
	}

	config := &container.Config{
		Hostname:        "",
		Domainname:      "",
		User:            "",
		ExposedPorts:    ports,
		AttachStdin:     false,
		AttachStdout:    true,
		AttachStderr:    true,
		Tty:             false,
		OpenStdin:       false,
		StdinOnce:       false,
		Image:           containerConfig.Image,
		Volumes:         map[string]struct{}{},
		NetworkDisabled: false,
	}

	return config, nil
}

func createHostConfig() (*container.HostConfig, error) {
	hostConfig := &container.HostConfig{
		PublishAllPorts: true,
		Privileged:      false,
	}
	return hostConfig, nil
}

func createNetworkConfig() (*network.NetworkingConfig, error) {
	netConfig := &network.NetworkingConfig{}
	return netConfig, nil
}
