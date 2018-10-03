package config_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/cloudfoundry-incubator/blockhead/pkg/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("NewState", func() {
		var (
			servicePath string
		)

		BeforeEach(func() {
			servicePath = "assets/services"
		})

		Context("broker config", func() {
			It("Opens the file and pours it into a Config struct", func() {
				state, err := config.NewState("assets/configs/test_config.json", servicePath)

				expectedConfig := config.Config{
					Username:         "username",
					Password:         "password",
					Port:             3335,
					ContainerManager: "docker",
					DeployerPath:     "/path/to/pusher.js",
					ExternalIP:       "1.1.1.1",
				}

				Expect(err).NotTo(HaveOccurred())
				Expect(state.Config).To(Equal(expectedConfig))
			})

			It("errors when provided a nonexistent config file", func() {
				state, err := config.NewState("not/a/real/file.json", servicePath)
				Expect(state).To(BeNil())
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

				defaultConfig := config.Config{
					Username:         "a-username-is-required",
					Password:         "a-password-is-required",
					Port:             3333,
					ContainerManager: "docker",
					DeployerPath:     "/path/to/pusher.js/is/required",
					ExternalIP:       "127.0.0.1",
				}
				Expect(state.Config).To(Equal(defaultConfig))
			})
		})

		Context("service definitions", func() {
			var (
				configPath = "assets/configs/test_config.json"
				err        error
			)

			Context("when service directory is empty", func() {
				It("errors when services dir is missing", func() {
					_, err := config.NewState(configPath, "")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Service Directory Missing"))
				})
			})

			Context("when service directory exists", func() {
				BeforeEach(func() {
					servicePath, err = ioutil.TempDir("", "service-dir")
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.RemoveAll(servicePath)
				})

				Context("when service directory is empty", func() {
					It("should complain about missing services", func() {
						_, err := config.NewState(configPath, servicePath)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("service directory is empty"))
					})
				})

				Context("when files in the directory are bad", func() {
					var serviceFilePath string
					BeforeEach(func() {
						serviceFilePath = fmt.Sprintf("%s/bad_config.json", servicePath)
						copy("assets/configs/bad_config.json", serviceFilePath)
					})

					It("should complain about bad services", func() {
						_, err := config.NewState(configPath, servicePath)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Error parsing service file"))
						Expect(err.Error()).To(ContainSubstring(serviceFilePath))
					})
				})

				Context("when there are service files in the service directory", func() {
					var (
						serviceFilePath  string
						expectedServices map[string]config.Service
						expectedService  config.Service
					)

					BeforeEach(func() {
						serviceFilePath = fmt.Sprintf("%s/service_config.json", servicePath)
						copy("assets/services/service_config.json", serviceFilePath)
						planMap := make(map[string]*config.Plan)
						expectedServices = make(map[string]config.Service)
						expectedService = config.Service{
							Name:        "name",
							Description: "desc",
							DisplayName: "display-name",
							Tags:        []string{"eth", "geth"},
						}

						expectedPlan := config.Plan{
							Name:        "plan-name",
							Image:       "image",
							Ports:       []string{"1234"},
							Description: "plan-desc",
						}

						planMap["uuid-2"] = &expectedPlan
						expectedService.Plans = planMap
						expectedServices["uuid-1"] = expectedService
					})

					It("should load the service", func() {
						parsedState, err := config.NewState(configPath, servicePath)
						Expect(err).NotTo(HaveOccurred())

						Expect(parsedState.Services).To(ConsistOf(utils.EquivalentService(&expectedService)))
					})

					Context("when there is another service", func() {
						var anotherExpectedService config.Service
						BeforeEach(func() {
							serviceFilePath = fmt.Sprintf("%s/service_config2.json", servicePath)
							copy("assets/services/service_config2.json", serviceFilePath)

							anotherExpectedService = config.Service{
								Name:        "name-2",
								Description: "desc",
								DisplayName: "display-name",
								Tags:        []string{"eth", "geth"},
							}

							expectedPlan := config.Plan{
								Name:        "plan-name-2",
								Image:       "image",
								Ports:       []string{"1234"},
								Description: "plan-desc",
							}

							planMap := make(map[string]*config.Plan)
							planMap["uuid-4"] = &expectedPlan
							anotherExpectedService.Plans = planMap
							expectedServices["uuid-3"] = anotherExpectedService

						})

						It("should have both services", func() {
							parsedState, err := config.NewState(configPath, servicePath)
							Expect(err).NotTo(HaveOccurred())

							Expect(parsedState.Services).To(ConsistOf(
								utils.EquivalentService(&expectedService),
								utils.EquivalentService(&anotherExpectedService),
							))
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
