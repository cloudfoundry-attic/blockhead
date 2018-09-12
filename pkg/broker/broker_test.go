package broker_test

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/pivotal-cf/brokerapi"

	"github.com/cloudfoundry-incubator/blockhead/fakes"
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
		manager         *fakes.FakeContainerManager
		servicesMap     map[string]*config.Service
		fakeLogger      *lagertest.TestLogger
	)

	BeforeEach(func() {
		servicesMap = make(map[string]*config.Service)
		manager = &fakes.FakeContainerManager{}
		fakeLogger = lagertest.NewTestLogger("test")

		service := config.Service{
			Name:        "eth",
			Description: "desc",
			DisplayName: "display-name",
			Tags:        []string{"eth", "geth"},
			Plans:       make(map[string]*config.Plan),
		}

		plan := config.Plan{
			Name:        "free",
			Image:       "some-image",
			Ports:       []string{"1234"},
			Description: "free-trial",
		}

		service.Plans["plan-id"] = &plan

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

		servicesMap[expectedService.ID] = &service
	})

	JustBeforeEach(func() {
		state = &config.State{Services: servicesMap}
		blockhead = broker.NewBlockheadBroker(fakeLogger, state, manager)
		ctx = context.Background()
	})

	Context("brokerapi", func() {
		It("implements the 7 brokerapi interface methods", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			instanceID := "instanceID"
			asyncAllowed := false
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
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
				service2 := config.Service{
					Name:        "eth",
					Description: "desc-2",
					DisplayName: "display-name-2",
					Tags:        []string{"eth", "geth"},
					Plans:       make(map[string]*config.Plan),
				}

				plan := config.Plan{
					Name:        "plan-2",
					Image:       "some-image-2",
					Description: "free-trial",
				}

				service2.Plans["plan-id-2"] = &plan

				free = true
				expectedService2 = brokerapi.Service{
					ID:          "service-id-2",
					Name:        "eth",
					Description: "desc-2",
					Bindable:    true,
					Tags:        []string{"eth", "geth"},
					Metadata: &brokerapi.ServiceMetadata{
						DisplayName: "display-name-2",
					},
					Plans: []brokerapi.ServicePlan{
						brokerapi.ServicePlan{
							ID:          "plan-id-2",
							Name:        "plan-2",
							Description: "free-trial",
							Free:        &free,
						},
					},
				}

				servicesMap[expectedService2.ID] = &service2
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

	Context("Provision", func() {
		It("returns an error if the service is missing", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "non-existing",
			}
			_, err := blockhead.Provision(ctx, "some-instance", provisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("service not found"))
		})

		It("returns an error if the plan is missing", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "service-id",
				PlanID:    "non-existing",
			}
			_, err := blockhead.Provision(ctx, "some-instance", provisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("plan not found"))
		})

		It("calls the manager's provisioner", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			_, err := blockhead.Provision(ctx, "some-instance", provisionDetails, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(manager.ProvisionCallCount()).To(Equal(1))

			_, config := manager.ProvisionArgsForCall(0)
			Expect(config.Name).To(Equal("some-instance"))
			Expect(config.ExposedPorts).To(ConsistOf("1234"))
			Expect(config.Image).To(Equal("some-image"))
		})
	})
	Context("Deprovision", func() {
		It("returns an error if the service is missing", func() {
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "non-existing",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("service not found"))
		})

		It("returns an error if the plan is missing", func() {
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "non-existing",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("plan not found"))
		})

		It("Calls the manager's deprovisioner", func() {
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(manager.DeprovisionCallCount()).To(Equal(1))

			_, instanceIDForCall := manager.DeprovisionArgsForCall(0)
			Expect(instanceIDForCall).To(Equal("some-instance"))
		})
		It("Bubbles up errors from the container manager", func() {
			errorMessage := "docker exploded"
			manager.DeprovisionReturns(errors.New(errorMessage))
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
	})
})
