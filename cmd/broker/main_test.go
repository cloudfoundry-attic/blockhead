package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi"
)

var _ = Describe("Blockhead", func() {
	var (
		client  *http.Client
		req     *http.Request
		session *gexec.Session

		args []string

		err error
	)

	Context("when broker is running", func() {
		BeforeEach(func() {
			args = []string{
				"-config",
				configPath,
				"-service",
				filepath.Join(absPath, "pkg/config/assets/service_config.json"),
				"-service",
				filepath.Join(absPath, "pkg/config/assets/another_service_config.json"),
			}
			client = &http.Client{}

		})

		JustBeforeEach(func() {
			cmd := exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			session.Interrupt()
		})

		Context("Service", func() {
			var resp *http.Response
			var expectedService1, expectedService2 brokerapi.Service

			puller := func() error {
				resp, err = client.Do(req)
				return err
			}

			BeforeEach(func() {
				True := true
				expectedService1 = brokerapi.Service{
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

				expectedService2 = brokerapi.Service{
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

			It("should successfully register multiple services", func() {
				catalogURL := fmt.Sprintf("%s%s", serverAddress, "/v2/catalog")
				req, err = http.NewRequest("GET", catalogURL, nil)
				Expect(err).NotTo(HaveOccurred())
				req.SetBasicAuth("test", "test")
				req.Header.Add("X-Broker-API-Version", "2.0")

				Eventually(puller).Should(Succeed())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close()
				bytes, _ := ioutil.ReadAll(resp.Body)

				catalog := brokerapi.CatalogResponse{}
				err = json.Unmarshal(bytes, &catalog)
				Expect(err).NotTo(HaveOccurred())
				Expect(catalog.Services).To(ConsistOf(expectedService1, expectedService2))
			})
		})
	})
})
