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
		nodePort         string
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		contractDeployer = deployer.NewEthereumDeployer(logger, filepath.Join("assets", "deployer_test.js"))
		nodePort = "8545"
	})

	It("runs the specified contract path", func() {
		contractInfo := &deployer.ContractInfo{
			ContractPath: "path-to-contract",
			ContractArgs: []string{"sample-arg-1", "sample-arg-2"},
		}

		portBindings := make(map[string][]containermanager.Binding)
		portBindings["8545"] = []containermanager.Binding{
			containermanager.Binding{
				Port: "1234",
			},
		}
		containerInfo := &containermanager.ContainerInfo{
			InternalAddress: "127.0.0.1",
			ExternalAddress: "12.34.56.78",
			Bindings:        portBindings,
		}

		nodeInfo, err := contractDeployer.DeployContract(contractInfo, containerInfo, nodePort)
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
