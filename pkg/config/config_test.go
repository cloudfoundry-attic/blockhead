package config_test

import (
	. "github.com/cloudfoundry-incubator/blockhead/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("Opens the file and pours it into a Config struct", func() {
			config, err := NewConfig("assets/test_config.json")

			Expect(err).NotTo(HaveOccurred())
			Expect(config.Username).To(Equal("username"))
			Expect(config.Password).To(Equal("password"))
		})

		It("fills in the service configuration", func() {
			cfg, err := NewConfig("assets/service_config.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Service.Name).To(Equal("eth"))
			Expect(cfg.Service.ID).To(Equal("some-id"))
			Expect(cfg.Service.Description).To(Equal("some-desc"))
			Expect(cfg.Service.Bindable).To(BeTrue())
			Expect(cfg.Service.Tags).To(ConsistOf("eth", "geth"))
			Expect(cfg.Service.Metadata.DisplayName).To(Equal("some-name"))
			Expect(cfg.Service.Metadata.LongDescription).To(Equal("some-long-desc"))
			Expect(cfg.Service.Metadata.ProviderDisplayName).To(Equal("some-provider-display-name"))
			Expect(cfg.Service.DashboardClient.ID).To(Equal("some-client-id"))
			Expect(cfg.Service.DashboardClient.Secret).To(Equal("some-secret"))
			Expect(cfg.Service.Plans).To(HaveLen(1))
			Expect(cfg.Service.Plans[0].ID).To(Equal("some-plan-id"))
			Expect(cfg.Service.Plans[0].Name).To(Equal("free"))
			Expect(cfg.Service.Plans[0].Description).To(Equal("free-trial"))
			Expect(cfg.Service.Plans[0].Metadata.Costs).To(HaveLen(1))
			Expect(cfg.Service.Plans[0].Metadata.Costs[0].Amount["usd"]).To(Equal(1.0))
			Expect(cfg.Service.Plans[0].Metadata.Costs[0].Unit).To(Equal("monthly"))
			Expect(cfg.Service.Plans[0].Metadata.Bullets).To(ConsistOf("dedicated-node", "another-node"))
			Expect((cfg.Service.Plans[0].Metadata.AdditionalMetadata["container"]).(map[string]interface{})["backend"]).To(Equal("docker"))
			Expect((cfg.Service.Plans[0].Metadata.AdditionalMetadata["container"]).(map[string]interface{})["image"]).To(Equal("some-image"))
		})

		It("errors when provided a nonexistent file", func() {
			cfg, err := NewConfig("not/a/real/file.json")
			Expect(cfg).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error opening config file"))
		})

		It("errors when provided a bad json file", func() {
			config, err := NewConfig("assets/bad_config.json")
			Expect(config).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error parsing config file"))
		})
	})
})
