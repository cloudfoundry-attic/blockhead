package broker

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/pivotal-cf/brokerapi"
)

type BlockheadBroker struct {
	state   *config.State
	manager containermanager.ContainerManager
	logger  lager.Logger
}

func NewBlockheadBroker(logger lager.Logger, state *config.State, manager containermanager.ContainerManager) BlockheadBroker {
	return BlockheadBroker{
		state:   state,
		manager: manager,
		logger:  logger,
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

	containerConfig := &containermanager.ContainerConfig{
		Name:         instanceID,
		Image:        plan.Image,
		ExposedPorts: plan.Ports,
	}

	return brokerapi.ProvisionedServiceSpec{}, b.manager.Provision(ctx, containerConfig)
}

func (b BlockheadBroker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	return brokerapi.DeprovisionServiceSpec{}, nil
}

func (b BlockheadBroker) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	return brokerapi.Binding{}, nil
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
