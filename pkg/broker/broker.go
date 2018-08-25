package broker

import (
	"context"

	"github.com/pivotal-cf/brokerapi"

	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
)

type BlockheadBroker struct {
	state *config.State
}

func NewBlockheadBroker(state *config.State) BlockheadBroker {
	return BlockheadBroker{
		state: state,
	}
}

func (b BlockheadBroker) Services(ctx context.Context) ([]brokerapi.Service, error) {
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
	return brokerapi.ProvisionedServiceSpec{}, nil
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
