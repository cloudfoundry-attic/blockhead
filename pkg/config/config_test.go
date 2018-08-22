package config_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/blockhead/pkg/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi"
)

var _ = Describe("Config", func() {
	var servicePath = "assets/services"

	Context("NewConfig", func() {
		It("Opens the file and pours it into a Config struct", func() {
			state, err := config.NewState("assets/configs/test_config.json", servicePath)

			Expect(err).NotTo(HaveOccurred())
			Expect(state.Config.Username).To(Equal("username"))
			Expect(state.Config.Password).To(Equal("password"))
			Expect(state.Config.Port).To(Equal(uint16(3335)))
		})

		It("errors when provided a nonexistent config file", func() {
			cfg, err := config.NewState("not/a/real/file.json", servicePath)
			Expect(cfg).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error opening config file"))
		})

		It("errors when provided a bad config file", func() {
			state, err := config.NewState("assets/configs/bad_config.json", servicePath)
			Expect(state).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error parsing config file"))
		})

		It("fills in defaults when provided an empty config file", func() {
			state, err := config.NewState("assets/configs/required_config.json", servicePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(state.Config.Username).To(Equal(""))
			Expect(state.Config.Password).To(Equal(""))
			Expect(state.Config.Port).To(Equal(uint16(3333)))
		})

		Context("service config", func() {
			var (
				tempFileName string
				cfg          config.Config
				err          error
			)

			JustBeforeEach(func() {
				tempfile, err := ioutil.TempFile("", "temp-config")
				Expect(err).NotTo(HaveOccurred())
				defer tempfile.Close()
				tempFileName = tempfile.Name()
				err = json.NewEncoder(tempfile).Encode(cfg)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.RemoveAll(tempFileName)
			})

			Context("when service directory is empty", func() {
				It("errors when services dir is missing", func() {
					_, err := config.NewState(tempFileName, "")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Service Directory Missing"))
				})
			})

			Context("when service directory exists", func() {
				var (
					tempDirName string
				)

				BeforeEach(func() {
					tempDirName, err = ioutil.TempDir("", "some-dir")
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.RemoveAll(tempDirName)
				})

				Context("when service directory is empty", func() {
					It("should complain about missing services", func() {
						_, err := config.NewState(tempFileName, tempDirName)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("service directory is empty"))
					})
				})

				Context("when files in the directory are bad", func() {
					var serviceFilePath string
					BeforeEach(func() {
						serviceFilePath = fmt.Sprintf("%s/bad_config.json", tempDirName)
						copy("assets/configs/bad_config.json", serviceFilePath)
					})

					It("should complain about bad services", func() {
						_, err := config.NewState(tempFileName, tempDirName)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Error parsing service file"))
						Expect(err.Error()).To(ContainSubstring(serviceFilePath))
					})
				})

				Context("when there are service files in the service directory", func() {
					var (
						serviceFilePath string
						True            bool
						expectedService brokerapi.Service
					)

					BeforeEach(func() {
						serviceFilePath = fmt.Sprintf("%s/service_config.json", tempDirName)
						copy("assets/services/service_config.json", serviceFilePath)

						True = true
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

					It("should load the service", func() {
						parsedState, err := config.NewState(tempFileName, tempDirName)
						Expect(err).NotTo(HaveOccurred())
						Expect(parsedState.Services).To(ConsistOf(expectedService))
					})

					Context("when there is another service", func() {
						var anotherExpectedService brokerapi.Service

						BeforeEach(func() {
							serviceFilePath = fmt.Sprintf("%s/another_service_config.json", tempDirName)
							copy("assets/services/another_service_config.json", serviceFilePath)

							anotherExpectedService = brokerapi.Service{
								ID:          "another-some-id",
								Name:        "another-eth",
								Description: "another-some-desc",
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

						It("should have both services", func() {
							parsedState, err := config.NewState(tempFileName, tempDirName)
							Expect(err).NotTo(HaveOccurred())
							Expect(parsedState.Services).To(ConsistOf(expectedService, anotherExpectedService))
						})
					})
				})
			})
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
