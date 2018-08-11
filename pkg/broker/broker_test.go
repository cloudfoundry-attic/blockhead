package broker_test

import (
	"context"

	"github.com/cloudfoundry-incubator/blockhead/pkg/broker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi"
)

var _ = Describe("Broker", func() {
	var (
		blockhead BlockheadBroker
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
		BeforeEach(func() {
			cfg, err = config.NewConfig("../config/assets/service_config.json")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return service definition", func() {
			services, err := blockhead.Services(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(HaveLen(1))

			service := services[0]
			Expect(service.Name).To(Equal("eth"))
			Expect(service.Description).To(Equal("some-desc"))
			Expect(service.Bindable).To(BeTrue())
			Expect(service.Tags).To(ConsistOf("eth", "geth"))
			Expect(service.Metadata.DisplayName).To(Equal("some-name"))
			Expect(service.Metadata.LongDescription).To(Equal("some-long-desc"))
			Expect(service.Metadata.ProviderDisplayName).To(Equal("some-provider-display-name"))
			Expect(service.DashboardClient.ID).To(Equal("some-client-id"))
			Expect(service.DashboardClient.Secret).To(Equal("some-secret"))
			Expect(service.Plans).To(HaveLen(1))
			Expect(service.Plans[0].ID).To(Equal("some-plan-id"))
			Expect(service.Plans[0].Name).To(Equal("free"))
			Expect(service.Plans[0].Description).To(Equal("free-trial"))
			Expect(service.Plans[0].Metadata.Costs).To(HaveLen(1))
			Expect(service.Plans[0].Metadata.Costs[0].Amount["usd"]).To(Equal(1.0))
			Expect(service.Plans[0].Metadata.Costs[0].Unit).To(Equal("monthly"))
			Expect(service.Plans[0].Metadata.Bullets).To(ConsistOf("dedicated-node", "another-node"))
			Expect((service.Plans[0].Metadata.AdditionalMetadata["container"]).(map[string]interface{})["backend"]).To(Equal("docker"))
			Expect((service.Plans[0].Metadata.AdditionalMetadata["container"]).(map[string]interface{})["image"]).To(Equal("some-image"))
		})
	})
})
