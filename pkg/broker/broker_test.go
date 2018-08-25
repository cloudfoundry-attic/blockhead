package broker_test

import (
	"context"

	"github.com/pivotal-cf/brokerapi"

	"github.com/cloudfoundry-incubator/blockhead/pkg/broker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/cloudfoundry-incubator/blockhead/pkg/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		blockhead       broker.BlockheadBroker
		ctx             context.Context
		state           *config.State
		expectedService brokerapi.Service
		free            bool
	)

	BeforeEach(func() {
		servicesMap := make(map[string]config.Service)

		service := config.Service{
			Name:        "eth",
			Description: "desc",
			DisplayName: "display-name",
			Tags:        []string{"eth", "geth"},
			Plans:       make(map[string]config.Plan),
		}

		plan := config.Plan{
			Name:        "free",
			Image:       "some-image",
			Description: "free-trial",
		}

		service.Plans["plan-id"] = plan

		free = true
		expectedService = brokerapi.Service{
			ID:          "service-id",
			Name:        "eth",
			Description: "desc",
			Bindable:    true,
			Tags:        []string{"eth", "geth"},
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName: "display-name",
			},
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					ID:          "plan-id",
					Name:        "free",
					Description: "free-trial",
					Free:        &free,
				},
			},
		}

		servicesMap[expectedService.ID] = service
		state = &config.State{Services: servicesMap}
	})

	JustBeforeEach(func() {
		blockhead = broker.NewBlockheadBroker(state)
		ctx = context.Background()
	})

	Context("brokerapi", func() {
		It("implements the 7 brokerapi interface methods", func() {
			provisionDetails := brokerapi.ProvisionDetails{}
			instanceID := "instanceID"
			asyncAllowed := false
			deprovisionDetails := brokerapi.DeprovisionDetails{}
			bindingID := "bindingID"
			bindDetails := brokerapi.BindDetails{}
			unbindDetails := brokerapi.UnbindDetails{}
			updateDetails := brokerapi.UpdateDetails{}
			operationData := "operationData"
			_, err := blockhead.Provision(ctx, instanceID, provisionDetails, asyncAllowed)
			Expect(err).NotTo(HaveOccurred())
			_, err = blockhead.Deprovision(ctx, instanceID, deprovisionDetails, asyncAllowed)
			Expect(err).NotTo(HaveOccurred())
			_, err = blockhead.Bind(ctx, instanceID, bindingID, bindDetails)
			Expect(err).NotTo(HaveOccurred())
			err = blockhead.Unbind(ctx, instanceID, bindingID, unbindDetails)
			Expect(err).NotTo(HaveOccurred())
			_, err = blockhead.Update(ctx, instanceID, updateDetails, asyncAllowed)
			Expect(err).NotTo(HaveOccurred())
			_, err = blockhead.LastOperation(ctx, instanceID, operationData)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Services", func() {
		It("should return service definition", func() {
			services, err := blockhead.Services(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(ConsistOf(expectedService))
		})

		Context("when more than one service or plan exists", func() {
			var expectedService2 brokerapi.Service
			BeforeEach(func() {
				service2 := state.Services["service-id"]
				plan2 := service2.Plans["plan-id"]

				service2.Name = "service-2"
				plan2.Name = "plan-2"

				service2.Plans = make(map[string]config.Plan)
				service2.Plans["plan-id"] = state.Services["service-id"].Plans["plan-id"]
				service2.Plans["plan-id-2"] = plan2

				state.Services["service-id-2"] = service2

				expectedService2 = expectedService
				expectedService2.ID = "service-id-2"
				expectedService2.Name = "service-2"

				expectedPlan2 := brokerapi.ServicePlan{
					ID:          "plan-id-2",
					Name:        "plan-2",
					Description: "free-trial",
					Free:        &free,
				}

				expectedService2.Plans = append(expectedService2.Plans, expectedPlan2)
			})

			It("should return all services and plans", func() {
				services, err := blockhead.Services(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(ConsistOf(
					utils.EquivalentBrokerAPIService(expectedService),
					utils.EquivalentBrokerAPIService(expectedService2),
				))
			})
		})
	})
})
