package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	brokerBinPath, sourcePath, absPath string
	serverAddress                      string
	configPath                         string
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

	return []byte(bytes)
}, func(data []byte) {
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
	cfg := config.Config{
		Username: "test",
		Password: "test",
		Port:     port,
	}
	cfgBytes, err := json.Marshal(cfg)
	_, err = f.Write(cfgBytes)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	os.RemoveAll(configPath)
})
