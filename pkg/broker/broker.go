package broker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/deployer"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi"
)

type BlockheadBroker struct {
	state    *config.State
	manager  containermanager.ContainerManager
	logger   lager.Logger
	deployer deployer.Deployer
}

type BindResponse struct {
	ContainerInfo *containermanager.ContainerInfo
	NodeInfo      *deployer.NodeInfo
}

func NewBlockheadBroker(logger lager.Logger, state *config.State, manager containermanager.ContainerManager, deployer deployer.Deployer) BlockheadBroker {
	return BlockheadBroker{
		state:    state,
		manager:  manager,
		logger:   logger,
		deployer: deployer,
	}
}

func (b BlockheadBroker) Services(ctx context.Context) ([]brokerapi.Service, error) {
	logger := b.logger.Session("services")
	logger.Info("started")
	defer logger.Info("finished")

	services := []brokerapi.Service{}
	free := true
	for serviceID, service := range b.state.Services {
		s := brokerapi.Service{
			ID:          serviceID,
			Name:        service.Name,
			Description: service.Description,
			Bindable:    true,
			Tags:        service.Tags,
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName: service.DisplayName,
			},
			Plans: []brokerapi.ServicePlan{},
		}

		for planID, plan := range service.Plans {
			p := brokerapi.ServicePlan{
				ID:          planID,
				Name:        plan.Name,
				Description: plan.Description,
				Free:        &free,
			}
			s.Plans = append(s.Plans, p)
		}
		services = append(services, s)
	}

	return services, nil
}

func (b BlockheadBroker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	logger := b.logger.Session("provision")
	logger.Info("started")
	defer logger.Info("finished")

	service := b.state.Services[details.ServiceID]
	if service == nil {
		return brokerapi.ProvisionedServiceSpec{}, errors.New("service not found")
	}

	plan := service.Plans[details.PlanID]
	if plan == nil {
		return brokerapi.ProvisionedServiceSpec{}, errors.New("plan not found")
	}

	containerConfig := containermanager.ContainerConfig{
		Name:         instanceID,
		Image:        plan.Image,
		ExposedPorts: plan.Ports,
	}

	return brokerapi.ProvisionedServiceSpec{}, b.manager.Provision(ctx, containerConfig)
}

func (b BlockheadBroker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	logger := b.logger.Session("deprovision")
	logger.Info("started")
	defer logger.Info("finished")

	service := b.state.Services[details.ServiceID]
	if service == nil {
		return brokerapi.DeprovisionServiceSpec{}, errors.New("service not found")
	}

	plan := service.Plans[details.PlanID]
	if plan == nil {
		return brokerapi.DeprovisionServiceSpec{}, errors.New("plan not found")
	}

	return brokerapi.DeprovisionServiceSpec{}, b.manager.Deprovision(ctx, instanceID)
}

func (b BlockheadBroker) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	logger := b.logger.Session("bind")
	logger.Info("started")
	defer logger.Info("finished")

	service := b.state.Services[details.ServiceID]
	if service == nil {
		return brokerapi.Binding{}, errors.New("service not found")
	}

	plan := service.Plans[details.PlanID]
	if plan == nil {
		return brokerapi.Binding{}, errors.New("plan not found")
	}

	contractInfo := &deployer.ContractInfo{}
	err := json.Unmarshal(details.RawParameters, contractInfo)
	if err != nil {
		return brokerapi.Binding{}, err
	}

	if contractInfo.ContractUrl == "" {
		return brokerapi.Binding{}, errors.New("contract_url not found")
	}

	file, err := ioutil.TempFile("", uuid.New())
	if err != nil {
		return brokerapi.Binding{}, err
	}
	contractInfo.ContractPath = file.Name()
	downloadFile(contractInfo.ContractPath, contractInfo.ContractUrl)
	defer os.RemoveAll(contractInfo.ContractPath)

	bindConfig := containermanager.BindConfig{
		InstanceId: instanceID,
		BindingId:  bindingID,
	}

	containerInfo, err := b.manager.Bind(ctx, bindConfig)
	if err != nil {
		return brokerapi.Binding{}, err
	}

	nodeInfo, err := b.deployer.DeployContract(contractInfo, containerInfo, plan.Ports[0])
	if err != nil {
		return brokerapi.Binding{}, err
	}

	return brokerapi.Binding{
		Credentials: BindResponse{
			ContainerInfo: containerInfo,
			NodeInfo:      nodeInfo,
		},
	}, nil
}

func (b BlockheadBroker) Unbind(ctx context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	return nil
}

func (b BlockheadBroker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	return brokerapi.UpdateServiceSpec{}, nil
}

func (b BlockheadBroker) LastOperation(ctx context.Context, instanceID, operationData string) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
