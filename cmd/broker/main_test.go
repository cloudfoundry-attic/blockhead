package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	"github.com/cloudfoundry-incubator/blockhead/pkg/utils"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockhead", func() {
	var (
		session *gexec.Session
		args    []string
		err     error
	)

	Context("when no args are passed in ", func() {
		BeforeEach(func() {
			args = []string{}
		})

		It("errors", func() {
			cmd := exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode).ToNot(Equal(0))
		})
	})

	Context("when only one arg is passed in ", func() {
		BeforeEach(func() {
			args = []string{
				configPath,
			}
		})

		It("errors", func() {
			cmd := exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode).ToNot(Equal(0))
		})
	})

	Context("when both args are passed in", func() {
		var (
			req    *http.Request
			client *http.Client
			resp   *http.Response
		)
		BeforeEach(func() {
			args = []string{
				configPath,
				servicePath,
			}
			client = &http.Client{}

			cmd := exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("Service", func() {
			var expectedService brokerapi.Service

			poller := func() error {
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
					Tags:        []string{"eth", "geth", "dev"},
					Metadata: &brokerapi.ServiceMetadata{
						DisplayName: "Geth 1.8",
					},
					Plans: []brokerapi.ServicePlan{
						brokerapi.ServicePlan{
							ID:          "d42fc3cc-1341-4aa3-866e-01bc5243dc2e",
							Name:        "free",
							Description: "Free Trial",
							Free:        &True,
						},
					},
				}
			})

			It("should successfully return services", func() {
				catalogURL := fmt.Sprintf("%s%s", serverAddress, "/v2/catalog")
				req, err = http.NewRequest("GET", catalogURL, nil)
				Expect(err).NotTo(HaveOccurred())
				req.SetBasicAuth("test", "test")
				req.Header.Add("X-Broker-API-Version", "2.0")

				Eventually(poller, 5*time.Second).Should(Succeed())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close()
				bytes, _ := ioutil.ReadAll(resp.Body)

				catalog := brokerapi.CatalogResponse{}
				err = json.Unmarshal(bytes, &catalog)
				Expect(err).NotTo(HaveOccurred())
				Expect(catalog.Services).To(ConsistOf(
					utils.EquivalentBrokerAPIService(expectedService)),
				)
			})
		})
	})
})
