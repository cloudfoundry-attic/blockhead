package deployer_test

import (
	"path/filepath"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/deployer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployer", func() {
	var (
		contractDeployer deployer.Deployer
		logger           *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		contractDeployer = deployer.NewEthereumDeployer(logger, filepath.Join("assets", "deployer_test.js"), "127.0.0.1")
	})

	It("runs the specified contract path", func() {
		contractInfo := &deployer.ContractInfo{
			ContractPath: "path-to-contract",
			ContractArgs: []string{"sample-arg-1", "sample-arg-2"},
		}

		portBindings := make(map[string][]containermanager.Binding)
		portBindings["8545"] = []containermanager.Binding{
			containermanager.Binding{
				HostIP: "12.34.56.78",
				Port:   "1234",
			},
		}
		containerInfo := &containermanager.ContainerInfo{
			IP:       "12.34.56.78",
			Bindings: portBindings,
		}

		nodeInfo, err := contractDeployer.DeployContract(contractInfo, containerInfo)
		Expect(err).ToNot(HaveOccurred())

		expectedNodeInfo := &deployer.NodeInfo{
			Account:         "sample-account",
			Interface:       "sample-abi",
			ContractAddress: "sample-address",
			GasPrice:        "0",
			TransactionHash: "sample-tx-hash",
		}

		Expect(nodeInfo).To(Equal(expectedNodeInfo))
	})
})
