package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
		session    *gexec.Session
		args       []string
		err        error
		containers []string
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
			client *http.Client
			cmd    *exec.Cmd
		)

		var newContainerId = func() string {
			instanceId := uuid.New()
			containers = append(containers, instanceId)
			return instanceId
		}

		BeforeEach(func() {
			containers = []string{}

			args = []string{
				configPath,
				servicePath,
			}
			client = &http.Client{}

			cmd = exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			cmd.Process.Signal(os.Kill)
		})

		Context("Service", func() {
			var expectedETHService, expectedFabricService brokerapi.Service

			BeforeEach(func() {
				True := true
				expectedETHService = brokerapi.Service{
					ID:          "not-checked-in-service-matcher",
					Name:        "eth",
					Description: "Ethereum Geth Node",
					Bindable:    true,
					Tags:        []string{"ethereum", "geth", "dev"},
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

				expectedFabricService = brokerapi.Service{
					ID:          "not-checked",
					Name:        "fab3",
					Description: "Hyperledger Fabric Proxy",
					Bindable:    true,
					Tags:        []string{"fabric", "proxy", "evm"},
					Metadata: &brokerapi.ServiceMetadata{
						DisplayName: "Fabric Proxy 0.1",
					},
					Plans: []brokerapi.ServicePlan{
						brokerapi.ServicePlan{
							ID:          "not-checked",
							Name:        "free",
							Description: "Free Trial",
							Free:        &True,
						},
					},
				}
			})

			Context("with an existing service", func() {
				var (
					serviceId, planId string
					cli               *dockerclient.Client
				)

				It("should successfully return service catalog", func() {
					resp := requestCatalog(client)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))

					catalog := parseCatalogResponse(resp)
					Expect(catalog.Services).To(ConsistOf(
						utils.EquivalentBrokerAPIService(expectedETHService),
						utils.EquivalentBrokerAPIService(expectedFabricService),
					))
				})

				Context("for ethereum service", func() {
					BeforeEach(func() {
						cli, err = dockerclient.NewEnvClient()
						Expect(err).NotTo(HaveOccurred())

						resp := requestCatalog(client)
						Expect(resp.StatusCode).To(Equal(http.StatusOK))

						catalog := parseCatalogResponse(resp)

						var service *brokerapi.Service
						for _, s := range catalog.Services {
							if contains(s.Tags, "ethereum") {
								service = &s
								break
							}
						}
						Expect(service).NotTo(BeNil())
						Expect(service.Plans).To(HaveLen(1))
						plan := service.Plans[0]
						serviceId = service.ID
						planId = plan.ID
					})

					Context("when provisioning", func() {
						AfterEach(func() {
							for _, containerId := range containers {
								err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
								Expect(err).NotTo(HaveOccurred())
							}
						})

						It("should successfully provision the service", func() {
							instanceId := newContainerId()
							resp := requestProvision(client, serviceId, planId, instanceId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							info, err := cli.ContainerInspect(context.Background(), instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("8545/tcp")))
						})

						Context("with an existing node", func() {
							BeforeEach(func() {
								existingInstanceId := newContainerId()
								resp := requestProvision(client, serviceId, planId, existingInstanceId)
								Expect(resp.StatusCode).To(Equal(http.StatusCreated))
							})

							It("successfully launches a second node", func() {
								instanceId := newContainerId()
								resp := requestProvision(client, serviceId, planId, instanceId)
								Expect(resp.StatusCode).To(Equal(http.StatusCreated))
							})
						})
					})

					Context("when deprovisioning", func() {
						var (
							instanceId string
						)

						BeforeEach(func() {
							instanceId = newContainerId()
							resp := requestProvision(client, serviceId, planId, instanceId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							info, err := cli.ContainerInspect(context.Background(), instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("8545/tcp")))
						})

						AfterEach(func() {
							for _, containerId := range containers {
								err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("No such container"))
							}
						})

						It("should successfully deprovision the service", func() {
							resp := requestDeprovision(client, serviceId, planId, instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(resp.StatusCode).To(Equal(http.StatusOK))
						})
					})

					Context("when binding", func() {
						var (
							instanceId string
						)

						BeforeEach(func() {
							instanceId = newContainerId()
							resp := requestProvision(client, serviceId, planId, instanceId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							info, err := cli.ContainerInspect(context.Background(), instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("8545/tcp")))
						})

						AfterEach(func() {
							for _, containerId := range containers {
								err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
								Expect(err).NotTo(HaveOccurred())
							}
						})

						It("should successfully return node information", func() {
							bindingId := uuid.New()
							resp := requestBind(client, serviceId, planId, instanceId, bindingId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							bindingResults := brokerapi.Binding{}
							body, err := ioutil.ReadAll(resp.Body)
							Expect(err).NotTo(HaveOccurred())
							json.Unmarshal(body, &bindingResults)
							creds := bindingResults.Credentials.(map[string]interface{})
							containerInfo := creds["ContainerInfo"].(map[string]interface{})
							Expect(containerInfo["Bindings"]).To(HaveKey("8545"))
							nodeInfo := creds["NodeInfo"].(map[string]interface{})
							Expect(nodeInfo["Account"]).NotTo(Equal(""))
							Expect(nodeInfo["ContractAddress"]).NotTo(Equal(""))
						})
					})
				})

				Context("for fabric service", func() {
					BeforeEach(func() {
						cli, err = dockerclient.NewEnvClient()
						Expect(err).NotTo(HaveOccurred())

						resp := requestCatalog(client)
						Expect(resp.StatusCode).To(Equal(http.StatusOK))

						catalog := parseCatalogResponse(resp)

						var service *brokerapi.Service
						for _, s := range catalog.Services {
							if contains(s.Tags, "fabric") {
								service = &s
								break
							}
						}
						Expect(service).NotTo(BeNil())
						Expect(service.Plans).To(HaveLen(1))
						plan := service.Plans[0]
						serviceId = service.ID
						planId = plan.ID
					})

					Context("when provisioning", func() {
						AfterEach(func() {
							for _, containerId := range containers {
								err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
								Expect(err).NotTo(HaveOccurred())
							}
						})

						It("should successfully provision the service", func() {
							instanceId := newContainerId()
							resp := requestProvision(client, serviceId, planId, instanceId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							info, err := cli.ContainerInspect(context.Background(), instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("5000/tcp")))
						})

						Context("with an existing node", func() {
							BeforeEach(func() {
								existingInstanceId := newContainerId()
								resp := requestProvision(client, serviceId, planId, existingInstanceId)
								Expect(resp.StatusCode).To(Equal(http.StatusCreated))
							})

							It("successfully launches a second node", func() {
								instanceId := newContainerId()
								resp := requestProvision(client, serviceId, planId, instanceId)
								Expect(resp.StatusCode).To(Equal(http.StatusCreated))
							})
						})
					})

					Context("when deprovisioning", func() {
						var (
							instanceId string
						)

						BeforeEach(func() {
							instanceId = newContainerId()
							resp := requestProvision(client, serviceId, planId, instanceId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							info, err := cli.ContainerInspect(context.Background(), instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("5000/tcp")))
						})

						AfterEach(func() {
							for _, containerId := range containers {
								err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("No such container"))
							}
						})

						It("should successfully deprovision the service", func() {
							resp := requestDeprovision(client, serviceId, planId, instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(resp.StatusCode).To(Equal(http.StatusOK))
						})
					})

					Context("when binding", func() {
						var (
							instanceId string
						)

						BeforeEach(func() {
							instanceId = newContainerId()
							resp := requestProvision(client, serviceId, planId, instanceId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							info, err := cli.ContainerInspect(context.Background(), instanceId)
							Expect(err).NotTo(HaveOccurred())
							Expect(info.Config.ExposedPorts).To(HaveKey(nat.Port("5000/tcp")))
						})

						AfterEach(func() {
							for _, containerId := range containers {
								err = cli.ContainerRemove(context.Background(), containerId, types.ContainerRemoveOptions{Force: true})
								Expect(err).NotTo(HaveOccurred())
							}
						})

						It("should successfully return node information", func() {
							bindingId := uuid.New()
							resp := requestBind(client, serviceId, planId, instanceId, bindingId)
							Expect(resp.StatusCode).To(Equal(http.StatusCreated))

							bindingResults := brokerapi.Binding{}
							body, err := ioutil.ReadAll(resp.Body)
							Expect(err).NotTo(HaveOccurred())
							json.Unmarshal(body, &bindingResults)
							creds := bindingResults.Credentials.(map[string]interface{})
							containerInfo := creds["ContainerInfo"].(map[string]interface{})
							Expect(containerInfo["Bindings"]).To(HaveKey("5000"))
							nodeInfo := creds["NodeInfo"].(map[string]interface{})
							Expect(nodeInfo["Account"]).NotTo(Equal(""))
							Expect(nodeInfo["ContractAddress"]).NotTo(Equal(""))
						})
					})
				})
			})
		})
	})
})

func requestCatalog(client *http.Client) *http.Response {
	catalogURL := fmt.Sprintf("%s/v2/catalog", serverAddress)
	req, err := http.NewRequest("GET", catalogURL, nil)
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

func requestProvision(client *http.Client, serviceId, planId, instanceId string) *http.Response {
	request := struct {
		ServiceId string `json:"service_id"`
		PlanId    string `json:"plan_id"`
	}{
		ServiceId: serviceId,
		PlanId:    planId,
	}

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(request)

	provisionURL := fmt.Sprintf("%s/v2/service_instances/%s", serverAddress, instanceId)
	req, err := http.NewRequest("PUT", provisionURL, payload)
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

func requestDeprovision(client *http.Client, serviceId, planId, instanceId string) *http.Response {
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode("{}")

	deprovisionURL := fmt.Sprintf("%s/v2/service_instances/%s?service_id=%s&plan_id=%s",
		serverAddress,
		instanceId,
		serviceId,
		planId,
	)
	req, err := http.NewRequest("DELETE", deprovisionURL, payload)
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

func requestBind(client *http.Client, serviceId, planId, instanceId, bindingId string) *http.Response {
	request := struct {
		ServiceId string      `json:"service_id"`
		PlanId    string      `json:"plan_id"`
		Params    interface{} `json:"parameters"`
	}{
		ServiceId: serviceId,
		PlanId:    planId,
		Params: struct {
			ContractUrl  string   `json:"contract_url"`
			ContractArgs []string `json:"contract_args"`
		}{
			ContractUrl:  "https://raw.githubusercontent.com/swetharepakula/hyperledger-fabric-evm-demo/blockhead_demo/poll.sol",
			ContractArgs: []string{"[1]"},
		},
	}

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(request)

	bindURL := fmt.Sprintf("%s/v2/service_instances/%s/service_bindings/%s",
		serverAddress,
		instanceId,
		bindingId,
	)
	req, err := http.NewRequest("PUT", bindURL, payload)
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

func parseCatalogResponse(resp *http.Response) brokerapi.CatalogResponse {
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)

	catalog := brokerapi.CatalogResponse{}
	err := json.Unmarshal(bytes, &catalog)
	Expect(err).NotTo(HaveOccurred())
	return catalog
}

func contains(array []string, elem string) bool {
	for _, v := range array {
		if v == elem {
			return true
		}
	}
	return false
}
