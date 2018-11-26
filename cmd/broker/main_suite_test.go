package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const (
	CLI_FLAGS_VERSION = "2.0.7"
	WEB3_VERSION      = "1.0.0-beta.36"
	SOLC_VERSION      = "0.4.25"
)

var (
	brokerBinPath, sourcePath, absPath string
	serverAddress                      string
	configPath                         string
	servicePath                        string
	cli                                *dockerclient.Client
)

func TestBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	sourcePath = "github.com/cloudfoundry-incubator/blockhead"
	absPath = filepath.Join(os.Getenv("GOPATH"), "src", sourcePath)
	binPath, err := gexec.Build(filepath.Join(sourcePath, "cmd/broker"))
	Expect(err).NotTo(HaveOccurred())

	bytes, err := json.Marshal([]string{binPath})
	Expect(err).NotTo(HaveOccurred())

	cli, err = dockerclient.NewEnvClient()
	Expect(err).NotTo(HaveOccurred())

	// Deleting this image so that integration tests will fail if broker never pulls the image down.
	// The image being used is shown in eth.json as `nimak/geth`. We only delete latest so other tagged
	// images are unaffected
	jsonString := `{
			"reference":{
				"nimak/geth:latest": true
		}
	}
	`

	filters, err := filters.FromParam(jsonString)
	Expect(err).ToNot(HaveOccurred())

	options := types.ImageListOptions{
		Filters: filters,
	}
	images, err := cli.ImageList(context.Background(), options)
	Expect(err).ToNot(HaveOccurred())

	if len(images) > 0 {
		Expect(images).To(HaveLen(1))
		imageId := images[0].ID

		removedImages, err := cli.ImageRemove(context.Background(), imageId, types.ImageRemoveOptions{Force: true})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(removedImages)).Should(BeNumerically(">", 0))
	}

	nodeLibVersionMap := make(map[string]string)
	nodeLibVersionMap["cli-flags"] = CLI_FLAGS_VERSION
	nodeLibVersionMap["web3"] = WEB3_VERSION
	nodeLibVersionMap["solc"] = SOLC_VERSION

	for lib, version := range nodeLibVersionMap {
		cmd := exec.Command("npm", "list", lib)
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Errored when verifying npm modeule: %s", lib))
		By(fmt.Sprintf("version for node module %s", lib), func() {
			Expect(output).To(ContainSubstring(version))
		})
	}

	return []byte(bytes)
}, func(data []byte) {
	sourcePath = "github.com/cloudfoundry-incubator/blockhead"
	absPath = filepath.Join(os.Getenv("GOPATH"), "src", sourcePath)
	paths := []string{}
	err := json.Unmarshal(data, &paths)
	Expect(err).NotTo(HaveOccurred())
	brokerBinPath = paths[0]

	port := 3333 + GinkgoParallelNode()
	serverAddress = fmt.Sprintf("http://0.0.0.0:%d", port)

	f, err := ioutil.TempFile("", "config.json")
	Expect(err).NotTo(HaveOccurred())
	defer f.Close()

	configPath = f.Name()
	By("using a temporary configuration file at: " + configPath)
	cfg := config.Config{
		Username:         "test",
		Password:         "test",
		Port:             uint16(port),
		DeployerPath:     filepath.Join(absPath, "pusher.js"),
		ContainerManager: "docker",
		ExternalAddress:  "127.0.0.1",
	}
	cfgBytes, err := json.Marshal(cfg)
	_, err = f.Write(cfgBytes)
	Expect(err).NotTo(HaveOccurred())
	servicePath = filepath.Join(absPath, "services")

	// Have to re-initialize the client so all nodes get a cli object.
	cli, err = dockerclient.NewEnvClient()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	os.RemoveAll(configPath)
})
