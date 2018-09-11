package docker_test

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry-incubator/blockhead/fakes"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager/docker"
	"github.com/docker/go-connections/nat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DockerManager", func() {
	var (
		manager         containermanager.ContainerManager
		client          *fakes.FakeDockerClient
		logger          *lagertest.TestLogger
		containerConfig *containermanager.ContainerConfig
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		client = &fakes.FakeDockerClient{}
		manager = docker.NewDockerContainerManager(logger, client)

		containerConfig = &containermanager.ContainerConfig{
			Name:         "some-name",
			Image:        "some-image",
			ExposedPorts: []string{"1234", "2345/udp"},
		}
	})

	Context("Provision", func() {
		It("pulls the image", func() {
			manager.Provision(context.TODO(), containerConfig)
			Expect(client.ImagePullCallCount()).To(Equal(1))
			_, img, _ := client.ImagePullArgsForCall(0)
			Expect(img).To(Equal("some-image"))
		})

		It("bubbles up errors from the docker client", func() {
			errorMessage := "potato not found"
			client.ImagePullReturns(nil, errors.New(errorMessage))

			err := manager.Provision(context.TODO(), containerConfig)
			Expect(err).To(HaveOccurred())
		})

		It("creates the container", func() {
			manager.Provision(context.TODO(), containerConfig)
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
			manager.Provision(context.TODO(), containerConfig)
			Expect(client.ContainerCreateCallCount()).To(Equal(1))
			_, _, _, _, name := client.ContainerCreateArgsForCall(0)
			Expect(client.ContainerStartCallCount()).To(Equal(1))
			_, startedContainerName, _ := client.ContainerStartArgsForCall(0)
			Expect(startedContainerName).To(Equal(name))
		})
	})
})
