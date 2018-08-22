package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

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
				configPath,
				servicePath,
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
			var expectedService brokerapi.Service

			puller := func() error {
				resp, err = client.Do(req)
				return err
			}

			BeforeEach(func() {
				True := true
				expectedService = brokerapi.Service{
					ID:          "24736f4a-72b8-4298-96f7-b48c4045ddfd",
					Name:        "eth",
					Description: "Ethereum Geth Node",
					Bindable:    true,
					Tags:        []string{"eth", "geth"},
					Metadata: &brokerapi.ServiceMetadata{
						DisplayName:         "Geth 1.8",
						LongDescription:     "Ethereum Geth Node",
						ProviderDisplayName: "Cloud Foundry Community",
					},
					DashboardClient: &brokerapi.ServiceDashboardClient{
						ID:     "blockhead-broker-eth",
						Secret: "",
					},
					Plans: []brokerapi.ServicePlan{
						brokerapi.ServicePlan{
							ID:          "d42fc3cc-1341-4aa3-866e-01bc5243dc2e",
							Name:        "free",
							Description: "Free Trial",
							Free:        &True,
							Metadata: &brokerapi.ServicePlanMetadata{
								DisplayName: "service-plan-metadata",
								Costs: []brokerapi.ServicePlanCost{
									brokerapi.ServicePlanCost{
										Amount: map[string]float64{"usd": 1.0},
										Unit:   "monthly",
									},
								},
								Bullets: []string{"dedicated-node"},
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
				Expect(catalog.Services).To(ConsistOf(expectedService))
			})
		})
	})
})
