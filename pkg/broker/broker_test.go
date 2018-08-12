package broker_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/blockhead/pkg/broker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/pivotal-cf/brokerapi"
)

var _ = Describe("Broker", func() {
	var (
		blockhead broker.BlockheadBroker
		ctx       context.Context
		cfg       *config.Config
		err       error
	)

	JustBeforeEach(func() {
		blockhead = broker.NewBlockheadBroker(*cfg)
		ctx = context.Background()
	})

	Context("brokerapi", func() {
		BeforeEach(func() {
			cfg = &config.Config{}
		})

		It("implements the 7 brokerapi interface methods", func() {
			_, err := blockhead.Services(ctx)
			Expect(err).NotTo(HaveOccurred())
			provisionDetails := brokerapi.ProvisionDetails{}

			instanceID := "instanceID"
			asyncAllowed := false
			deprovisionDetails := brokerapi.DeprovisionDetails{}
			bindingID := "bindingID"
			bindDetails := brokerapi.BindDetails{}
			unbindDetails := brokerapi.UnbindDetails{}
			updateDetails := brokerapi.UpdateDetails{}
			operationData := "operationData"
			_, err = blockhead.Provision(ctx, instanceID, provisionDetails, asyncAllowed)
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
		var expectedService brokerapi.Service

		BeforeEach(func() {
			cfg, err = config.NewConfig(
				"../config/assets/test_config.json",
				config.ServiceFlags{
					"../config/assets/service_config.json",
				},
			)
			Expect(err).NotTo(HaveOccurred())

			True := true
			expectedService = brokerapi.Service{
				ID:          "some-id",
				Name:        "eth",
				Description: "some-desc",
				Bindable:    true,
				Tags:        []string{"eth", "geth"},
				Metadata: &brokerapi.ServiceMetadata{
					DisplayName:         "some-name",
					LongDescription:     "some-long-desc",
					ProviderDisplayName: "some-provider-display-name",
				},
				DashboardClient: &brokerapi.ServiceDashboardClient{
					ID:     "some-client-id",
					Secret: "some-secret",
				},
				Plans: []brokerapi.ServicePlan{
					brokerapi.ServicePlan{
						ID:          "some-plan-id",
						Name:        "free",
						Description: "free-trial",
						Free:        &True,
						Metadata: &brokerapi.ServicePlanMetadata{
							DisplayName: "service-plan-metadata",
							Costs: []brokerapi.ServicePlanCost{
								brokerapi.ServicePlanCost{
									Amount: map[string]float64{"usd": 1.0},
									Unit:   "monthly",
								},
							},
							Bullets: []string{"dedicated-node", "another-node"},
							AdditionalMetadata: map[string]interface{}{
								"container": map[string]interface{}{
									"backend": "docker",
									"image":   "some-image",
								},
							},
						},
					},
				},
			}
		})

		It("should return service definition", func() {
			services, err := blockhead.Services(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(ConsistOf(expectedService))
		})
	})
})
