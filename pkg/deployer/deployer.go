package deployer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/pborman/uuid"
)

type ContractInfo struct {
	ContractUrl  string   `json:"contract_url"`
	ContractArgs []string `json:"contract_args"`
	ContractPath string
}

type NodeInfo struct {
	Account         string `json:"address"`
	Interface       string `json:"abi"`
	ContractAddress string `json:"contract_address"`
	GasPrice        string `json:"gas_price"`
	TransactionHash string `json:"transaction_hash"`
}

//go:generate counterfeiter -o ../fakes/fake_deployer.go . Deployer
type Deployer interface {
	DeployContract(contractInfo *ContractInfo, containerInfo *containermanager.ContainerInfo, nodePort string) (*NodeInfo, error)
}

type ethereumDeployer struct {
	logger       lager.Logger
	deployerPath string
}

func NewEthereumDeployer(logger lager.Logger, deployerPath string) Deployer {
	return &ethereumDeployer{
		logger:       logger,
		deployerPath: deployerPath,
	}
}

func (e ethereumDeployer) DeployContract(contractInfo *ContractInfo, containerInfo *containermanager.ContainerInfo, nodePort string) (*NodeInfo, error) {
	e.logger.Info("deploy-started")
	defer e.logger.Info("deploy-finished")

	// nodePort is the port we want from the blockchain node
	portBindings := containerInfo.Bindings[nodePort]
	if len(portBindings) <= 0 {
		return nil, errors.New(fmt.Sprintf("Port Bindings do not have %s port mapping", nodePort))
	}
	config := struct {
		Provider string   `json:"provider"`
		Password string   `json:"password"`
		Args     []string `json:"args"`
	}{
		Provider: fmt.Sprintf("http://%s:%s", containerInfo.InternalAddress, portBindings[0].Port),
		Password: "",
		Args:     contractInfo.ContractArgs,
	}

	configFile, err := ioutil.TempFile("", uuid.New())
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(configFile.Name())

	configJson, _ := json.Marshal(config)
	err = ioutil.WriteFile(configFile.Name(), configJson, 0644)
	if err != nil {
		return nil, err
	}

	outputFile, err := ioutil.TempFile("", uuid.New())
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(outputFile.Name())

	cmd := exec.Command("node", e.deployerPath, "-c", configFile.Name(), "-o", outputFile.Name(), contractInfo.ContractPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		e.logger.Error("run-failed", err, lager.Data{"output": string(output)})
		return nil, err
	}

	content, err := ioutil.ReadFile(outputFile.Name())
	if err != nil {
		e.logger.Error("reading-output-failed", err, lager.Data{"output": string(output), "content": string(content)})
		return nil, err
	}

	nodeInfo := &NodeInfo{}
	err = json.Unmarshal(content, nodeInfo)
	if err != nil {
		e.logger.Error("parsing-content-failed", err, lager.Data{"output": string(output), "content": string(content)})
		return nil, err
	}

	e.logger.Debug("deploy-data", lager.Data{"output": string(output), "content": string(content)})
	e.logger.Info("deploy-succeeded")
	return nodeInfo, nil
}
