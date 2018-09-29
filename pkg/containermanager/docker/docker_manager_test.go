package docker_test

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager/docker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/fakes"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DockerManager", func() {
	var (
		manager         containermanager.ContainerManager
		client          *fakes.FakeDockerClient
		logger          *lagertest.TestLogger
		containerConfig containermanager.ContainerConfig
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		client = &fakes.FakeDockerClient{}
		manager = docker.NewDockerContainerManager(logger, client, "../../utils/pusher.js")

		containerConfig = containermanager.ContainerConfig{
			Name:         "some-name",
			Image:        "some-image",
			ExposedPorts: []string{"1234", "2345/udp"},
		}
	})

	Context("Provision", func() {
		It("pulls the image", func() {
			manager.Provision(context.Background(), containerConfig)
			Expect(client.ImagePullCallCount()).To(Equal(1))
			_, img, _ := client.ImagePullArgsForCall(0)
			Expect(img).To(Equal("some-image"))
		})

		It("bubbles up errors from the docker client", func() {
			errorMessage := "potato not found"
			client.ImagePullReturns(nil, errors.New(errorMessage))

			err := manager.Provision(context.Background(), containerConfig)
			Expect(err).To(HaveOccurred())
		})

		It("creates the container", func() {
			manager.Provision(context.Background(), containerConfig)
			Expect(client.ContainerCreateCallCount()).To(Equal(1))
			_, config, hostConfig, networkConfig, name := client.ContainerCreateArgsForCall(0)

			Expect(config).NotTo(BeNil())
			Expect(config.ExposedPorts).To(HaveKey(nat.Port("1234/tcp")))
			Expect(config.ExposedPorts).To(HaveKey(nat.Port("2345/udp")))
			Expect(hostConfig).NotTo(BeNil())
			Expect(networkConfig).NotTo(BeNil())
			Expect(name).To(Equal("some-name"))
		})

		It("starts the container", func() {
			manager.Provision(context.Background(), containerConfig)
			Expect(client.ContainerCreateCallCount()).To(Equal(1))
			_, _, _, _, name := client.ContainerCreateArgsForCall(0)
			Expect(client.ContainerStartCallCount()).To(Equal(1))
			_, startedContainerName, _ := client.ContainerStartArgsForCall(0)
			Expect(startedContainerName).To(Equal(name))
		})
	})

	Context("Deprovision", func() {
		It("calls the docker client to stop and remove the specificed container", func() {
			manager.Deprovision(context.Background(), containerConfig.Name)
			Expect(client.ContainerStopCallCount()).To(Equal(1))
			_, stoppedContainerName, _ := client.ContainerStopArgsForCall(0)
			Expect(stoppedContainerName).To(Equal(containerConfig.Name))
			Expect(client.ContainerRemoveCallCount()).To(Equal(1))
			_, removedContainerName, _ := client.ContainerRemoveArgsForCall(0)
			Expect(removedContainerName).To(Equal(containerConfig.Name))
		})
	})

	Context("Bind", func() {
		var bindConfig containermanager.BindConfig

		BeforeEach(func() {
			bindConfig = containermanager.BindConfig{
				InstanceId: "some-instance",
				BindingId:  "some-bind",
			}
		})

		Context("when service container is missing", func() {
			BeforeEach(func() {
				client.ContainerInspectReturns(types.ContainerJSON{}, errors.New("boom"))
			})

			It("returns an error", func() {
				bindResponse, err := manager.Bind(context.TODO(), bindConfig)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("boom"))
				Expect(bindResponse).To(BeNil())
			})
		})

		Context("when service container exists", func() {
			var expectedBindResponse *containermanager.ContainerInfo
			BeforeEach(func() {
				port, err := nat.NewPort("tcp", "1234")
				Expect(err).NotTo(HaveOccurred())

				portBinding := nat.PortBinding{
					HostIP:   "some-host-ip",
					HostPort: "6789",
				}

				portMap := nat.PortMap{}
				portMap[port] = []nat.PortBinding{portBinding}

				containerResponse := types.ContainerJSON{
					NetworkSettings: &types.NetworkSettings{
						NetworkSettingsBase: types.NetworkSettingsBase{
							Ports: portMap,
						},
						DefaultNetworkSettings: types.DefaultNetworkSettings{
							IPAddress: "some-ip-address",
						},
					},
				}
				client.ContainerInspectReturns(containerResponse, nil)

				bindings := make(map[string][]containermanager.Binding)
				bindings["1234"] = []containermanager.Binding{
					containermanager.Binding{
						HostIP: "some-host-ip",
						Port:   "6789",
					},
				}
				expectedBindResponse = &containermanager.ContainerInfo{
					IP:       "some-ip-address",
					Bindings: bindings,
				}
			})

			It("inspects the service container", func() {
				manager.Bind(context.TODO(), bindConfig)
				Expect(client.ContainerInspectCallCount()).To(Equal(1))
				_, instanceId := client.ContainerInspectArgsForCall(0)
				Expect(instanceId).To(Equal(bindConfig.InstanceId))
			})

			It("should return container info in the bind response", func() {
				bindResponse, _ := manager.Bind(context.TODO(), bindConfig)
				Expect(bindResponse).To(Equal(expectedBindResponse))
			})
		})
	})
})
