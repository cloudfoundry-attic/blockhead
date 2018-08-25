package broker_test

import (
	"context"

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
})
