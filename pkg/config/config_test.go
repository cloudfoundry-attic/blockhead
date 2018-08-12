package config_test

import (
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("Opens the file and pours it into a Config struct", func() {
			config, err := config.NewConfig("assets/test_config.json", config.ServiceFlags{})

			Expect(err).NotTo(HaveOccurred())
			Expect(config.Username).To(Equal("username"))
			Expect(config.Password).To(Equal("password"))
		})

		It("errors when provided a nonexistent config file", func() {
			cfg, err := config.NewConfig("not/a/real/file.json", config.ServiceFlags{})
			Expect(cfg).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error opening config file"))
		})

		It("errors when provided a bad config file", func() {
			config, err := config.NewConfig("assets/bad_config.json", config.ServiceFlags{})
			Expect(config).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error parsing config file"))
		})

		Context("service config", func() {
			It("errors when service file is missing", func() {
				cfg, err := config.NewConfig("assets/test_config.json", config.ServiceFlags{"some-path"})
				Expect(cfg).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error opening service file"))
			})

			It("errors when provided a bad service file", func() {
				config, err := config.NewConfig("assets/test_config.json", config.ServiceFlags{"assets/bad_config.json"})
				Expect(config).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error parsing service file"))
			})
		})

		It("fills in the service configuration", func() {
			True := true
			expectedService := brokerapi.Service{
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

			cfg, err := config.NewConfig(
				"assets/test_config.json",
				config.ServiceFlags{"assets/service_config.json"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Services).To(ConsistOf(expectedService))
		})
	})
})
