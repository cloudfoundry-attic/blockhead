package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/cloudfoundry-incubator/blockhead/pkg/utils"
	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
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

			BeforeEach(func() {
				True := true
				expectedService = brokerapi.Service{
					ID:          "not-checked-in-service-matcher",
					Name:        "eth",
					Description: "Ethereum Geth Node",
					Bindable:    true,
					Tags:        []string{"eth", "geth", "dev"},
					Metadata: &brokerapi.ServiceMetadata{
						DisplayName: "Geth 1.8",
					},
					Plans: []brokerapi.ServicePlan{
						brokerapi.ServicePlan{
							ID:          "not-checked-in-service-matcher",
							Name:        "free",
							Description: "Free Trial",
							Free:        &True,
						},
					},
				}
			})

			var requestCatalog = func() *http.Response {
				catalogURL := fmt.Sprintf("%s/v2/%s", serverAddress, "catalog")
				req, err = http.NewRequest("GET", catalogURL, nil)
				Expect(err).NotTo(HaveOccurred())
				req.SetBasicAuth("test", "test")
				req.Header.Add("X-Broker-API-Version", "2.0")

				var resp *http.Response
				Eventually(func() error {
					resp, err = client.Do(req)
					return err
				}).Should(Succeed())
				return resp
			}

			var requestProvision = func(serviceId, planId, instanceId string) *http.Response {
				request := struct {
					ServiceId string `json:"service_id"`
					PlanId    string `json:"plan_id"`
				}{
					ServiceId: serviceId,
					PlanId:    planId,
				}

				payload := new(bytes.Buffer)
				json.NewEncoder(payload).Encode(request)

				provisionURL := fmt.Sprintf("%s/v2/%s/%s", serverAddress, "service_instances", instanceId)
				req, err = http.NewRequest("PUT", provisionURL, payload)
				Expect(err).NotTo(HaveOccurred())
				req.SetBasicAuth("test", "test")
				req.Header.Add("X-Broker-API-Version", "2.0")
				req.Header.Add("Content-Type", "application/json")

				var resp *http.Response
				Eventually(func() error {
					resp, err = client.Do(req)
					return err
				}).Should(Succeed())
				return resp
			}

			var requestDeprovision = func(serviceId, planId, instanceId string) *http.Response {
				payload := new(bytes.Buffer)
				json.NewEncoder(payload).Encode("{}")

				deprovisionURL := fmt.Sprintf("%s/v2/%s/%s?service_id=%s&plan_id=%s",
					serverAddress,
					"service_instances",
					instanceId,
					serviceId,
					planId,
				)
				req, err = http.NewRequest("DELETE", deprovisionURL, payload)
				Expect(err).NotTo(HaveOccurred())
				req.SetBasicAuth("test", "test")
				req.Header.Add("X-Broker-API-Version", "2.0")
				req.Header.Add("Content-Type", "application/json")

				var resp *http.Response
				Eventually(func() error {
					resp, err = client.Do(req)
					return err
				}).Should(Succeed())
				return resp
			}

			var parseCatalogResponse = func(resp *http.Response) brokerapi.CatalogResponse {
				defer resp.Body.Close()
				bytes, _ := ioutil.ReadAll(resp.Body)

				catalog := brokerapi.CatalogResponse{}
				err = json.Unmarshal(bytes, &catalog)
				Expect(err).NotTo(HaveOccurred())
				return catalog
			}

			Context("with an existing service", func() {

				It("should successfully return service catalog", func() {
					resp := requestCatalog()
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))

					catalog := parseCatalogResponse(resp)
					Expect(catalog.Services).To(ConsistOf(
						utils.EquivalentBrokerAPIService(expectedService)),
					)
				})

				Context("when provisioning", func() {
					var (
						serviceId, planId string
						cli               *dockerclient.Client
						containers        []string
					)

					var newContainerId = func() string {
						instanceId := uuid.New()
						containers = append(containers, instanceId)
						return instanceId
					}

					BeforeEach(func() {
						containers = []string{}

						cli, err = dockerclient.NewEnvClient()
						Expect(err).NotTo(HaveOccurred())

						resp := requestCatalog()
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.StatusCode).To(Equal(http.StatusOK))

						catalog := parseCatalogResponse(resp)
						service := catalog.Services[0]
						Expect(service.Plans).To(HaveLen(1))
						plan := service.Plans[0]
						serviceId = service.ID
						planId = plan.ID
					})

					AfterEach(func() {
						Expect(err).NotTo(HaveOccurred())
						for _, containerId := range containers {
							err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
						}
						Expect(err).NotTo(HaveOccurred())
					})

					It("should successfully provision the service", func() {
						instanceId := newContainerId()
						resp := requestProvision(serviceId, planId, instanceId)
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.StatusCode).To(Equal(http.StatusCreated))

						info, err := cli.ContainerInspect(context.Background(), instanceId)
						Expect(err).NotTo(HaveOccurred())
						Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("8545/tcp")))
					})

					Context("with an existing node", func() {
						BeforeEach(func() {
							existingInstanceId := newContainerId()
							resp := requestProvision(serviceId, planId, existingInstanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))
						})

						It("successfully launches a second node", func() {
							instanceId := newContainerId()
							resp := requestProvision(serviceId, planId, instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))
						})
					})
				})
				Context("when deprovisioning", func() {
					var (
						instanceId, serviceId, planId string
						cli                           *dockerclient.Client
						containers                    []string
					)

					var newContainerId = func() string {
						instanceId := uuid.New()
						containers = append(containers, instanceId)
						return instanceId
					}

					BeforeEach(func() {
						containers = []string{}

						cli, err = dockerclient.NewEnvClient()
						Expect(err).NotTo(HaveOccurred())

						resp := requestCatalog()
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.StatusCode).To(Equal(http.StatusOK))

						catalog := parseCatalogResponse(resp)
						service := catalog.Services[0]
						Expect(service.Plans).To(HaveLen(1))
						plan := service.Plans[0]
						serviceId = service.ID
						planId = plan.ID

						instanceId = newContainerId()
						resp = requestProvision(serviceId, planId, instanceId)
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.StatusCode).To(Equal(http.StatusCreated))

						info, err := cli.ContainerInspect(context.Background(), instanceId)
						Expect(err).NotTo(HaveOccurred())
						Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("8545/tcp")))
					})

					AfterEach(func() {
						for _, containerId := range containers {
							err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
						}
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("No such container"))
					})

					It("should successfully deprovision the service", func() {
						resp := requestDeprovision(serviceId, planId, instanceId)
						Expect(err).NotTo(HaveOccurred())
						Expect(resp.StatusCode).To(Equal(http.StatusOK))
					})
				})
			})
		})
	})
})
