package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pivotal-cf/brokerapi"

	"github.com/cloudfoundry-incubator/blockhead/pkg/broker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		blockhead broker.BlockheadBroker
		ctx       context.Context
		state     *config.State
	)

	JustBeforeEach(func() {
		blockhead = broker.NewBlockheadBroker(state)
		ctx = context.Background()
	})

	Context("brokerapi", func() {
		BeforeEach(func() {
			state = &config.State{}
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
			tempDirName, err := ioutil.TempDir("", "some-dir")
			Expect(err).NotTo(HaveOccurred())
			tmpCfg := config.Config{}

			serviceFilePath := fmt.Sprintf("%s/service_config.json", tempDirName)
			copy("../../pkg/config/assets/services/service_config.json", serviceFilePath)

			tempfile, err := ioutil.TempFile("", "temp-config")
			Expect(err).NotTo(HaveOccurred())
			defer tempfile.Close()
			tempFileName := tempfile.Name()
			err = json.NewEncoder(tempfile).Encode(tmpCfg)
			Expect(err).NotTo(HaveOccurred())

			state, err = config.NewState(tempFileName, tempDirName)
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

func copy(src string, dst string) {
	// Read all content of src to data
	data, err := ioutil.ReadFile(src)
	Expect(err).NotTo(HaveOccurred())
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0644)
	Expect(err).NotTo(HaveOccurred())
}
