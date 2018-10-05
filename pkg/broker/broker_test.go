package broker_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/pivotal-cf/brokerapi"

	"github.com/cloudfoundry-incubator/blockhead/pkg/broker"
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/deployer"
	"github.com/cloudfoundry-incubator/blockhead/pkg/fakes"
	"github.com/cloudfoundry-incubator/blockhead/pkg/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Broker", func() {
	var (
		blockhead       broker.BlockheadBroker
		ctx             context.Context
		expectedService brokerapi.Service
		servicesMap     map[string]*config.Service
		fakeLogger      *lagertest.TestLogger

		fakeManager  *fakes.FakeContainerManager
		fakeDeployer *fakes.FakeDeployer

		state *config.State
		free  bool
	)

	BeforeEach(func() {
		servicesMap = make(map[string]*config.Service)
		fakeManager = &fakes.FakeContainerManager{}
		fakeDeployer = &fakes.FakeDeployer{}
		fakeLogger = lagertest.NewTestLogger("test")

		service := config.Service{
			Name:        "eth",
			Description: "desc",
			DisplayName: "display-name",
			Tags:        []string{"eth", "geth"},
			Plans:       make(map[string]*config.Plan),
		}

		plan := config.Plan{
			Name:        "free",
			Image:       "some-image",
			Ports:       []string{"1234"},
			Description: "free-trial",
		}

		service.Plans["plan-id"] = &plan

		free = true
		expectedService = brokerapi.Service{
			ID:          "service-id",
			Name:        "eth",
			Description: "desc",
			Bindable:    true,
			Tags:        []string{"eth", "geth"},
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName: "display-name",
			},
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					ID:          "plan-id",
					Name:        "free",
					Description: "free-trial",
					Free:        &free,
				},
			},
		}

		servicesMap[expectedService.ID] = &service
	})

	JustBeforeEach(func() {
		state = &config.State{Services: servicesMap}
		blockhead = broker.NewBlockheadBroker(fakeLogger, state, fakeManager, fakeDeployer)
		ctx = context.Background()
	})

	Context("brokerapi", func() {
		It("implements the 7 brokerapi interface methods", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			instanceID := "instanceID"
			asyncAllowed := false
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			bindingID := "bindingID"
			bindDetails := brokerapi.BindDetails{
				ServiceID:     "service-id",
				PlanID:        "plan-id",
				RawParameters: []byte(`{"contract_url": "some-url"}`),
			}
			unbindDetails := brokerapi.UnbindDetails{}
			updateDetails := brokerapi.UpdateDetails{}
			operationData := "operationData"
			_, err := blockhead.Provision(ctx, instanceID, provisionDetails, asyncAllowed)
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

	Context("Services", func() {
		It("should return service definition", func() {
			services, err := blockhead.Services(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(services).To(ConsistOf(expectedService))
		})

		Context("when more than one service or plan exists", func() {
			var expectedService2 brokerapi.Service
			BeforeEach(func() {
				service2 := config.Service{
					Name:        "eth",
					Description: "desc-2",
					DisplayName: "display-name-2",
					Tags:        []string{"eth", "geth"},
					Plans:       make(map[string]*config.Plan),
				}

				plan := config.Plan{
					Name:        "plan-2",
					Image:       "some-image-2",
					Description: "free-trial",
				}

				service2.Plans["plan-id-2"] = &plan

				free = true
				expectedService2 = brokerapi.Service{
					ID:          "service-id-2",
					Name:        "eth",
					Description: "desc-2",
					Bindable:    true,
					Tags:        []string{"eth", "geth"},
					Metadata: &brokerapi.ServiceMetadata{
						DisplayName: "display-name-2",
					},
					Plans: []brokerapi.ServicePlan{
						brokerapi.ServicePlan{
							ID:          "plan-id-2",
							Name:        "plan-2",
							Description: "free-trial",
							Free:        &free,
						},
					},
				}

				servicesMap[expectedService2.ID] = &service2
			})

			It("should return all services and plans", func() {
				services, err := blockhead.Services(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(services).To(ConsistOf(
					utils.EquivalentBrokerAPIService(expectedService),
					utils.EquivalentBrokerAPIService(expectedService2),
				))
			})
		})
	})

	Context("Provision", func() {
		It("returns an error if the service is missing", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "non-existing",
			}
			_, err := blockhead.Provision(ctx, "some-instance", provisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("service not found"))
		})

		It("returns an error if the plan is missing", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "service-id",
				PlanID:    "non-existing",
			}
			_, err := blockhead.Provision(ctx, "some-instance", provisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("plan not found"))
		})

		It("calls the fakeManager's provisioner", func() {
			provisionDetails := brokerapi.ProvisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			_, err := blockhead.Provision(ctx, "some-instance", provisionDetails, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeManager.ProvisionCallCount()).To(Equal(1))

			_, config := fakeManager.ProvisionArgsForCall(0)
			Expect(config.Name).To(Equal("some-instance"))
			Expect(config.ExposedPorts).To(ConsistOf("1234"))
			Expect(config.Image).To(Equal("some-image"))
		})
	})

	Context("Deprovision", func() {
		It("returns an error if the service is missing", func() {
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "non-existing",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("service not found"))
		})

		It("returns an error if the plan is missing", func() {
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "non-existing",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("plan not found"))
		})

		It("Calls the fakeManager's deprovisioner", func() {
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeManager.DeprovisionCallCount()).To(Equal(1))

			_, instanceIDForCall := fakeManager.DeprovisionArgsForCall(0)
			Expect(instanceIDForCall).To(Equal("some-instance"))
		})

		It("Bubbles up errors from the container fakeManager", func() {
			errorMessage := "docker exploded"
			fakeManager.DeprovisionReturns(errors.New(errorMessage))
			deprovisionDetails := brokerapi.DeprovisionDetails{
				ServiceID: "service-id",
				PlanID:    "plan-id",
			}
			_, err := blockhead.Deprovision(ctx, "some-instance", deprovisionDetails, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
	})

	Context("Binding", func() {
		It("returns an error if the service is missing", func() {
			bindDetails := brokerapi.BindDetails{
				ServiceID: "non-existing",
			}
			_, err := blockhead.Bind(ctx, "some-instance", "some-binding-id", bindDetails)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("service not found"))
		})

		It("returns an error if the plan is missing", func() {
			bindDetails := brokerapi.BindDetails{
				ServiceID: "service-id",
				PlanID:    "non-existing",
			}
			_, err := blockhead.Bind(ctx, "some-instance", "some-binding-id", bindDetails)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("plan not found"))
		})

		Context("with bind params", func() {
			var (
				jsonParams  string
				bindDetails brokerapi.BindDetails
			)

			JustBeforeEach(func() {
				params := []byte(jsonParams)
				bindDetails = brokerapi.BindDetails{
					ServiceID:     "service-id",
					PlanID:        "plan-id",
					RawParameters: params,
				}
			})

			Context("when invalid", func() {
				Context("when contract_url is missing", func() {
					BeforeEach(func() {
						jsonParams = ` {
							"contract_args": ["arg1", "arg2"]
						} `
					})

					It("returns a contract_url not found error", func() {
						_, err := blockhead.Bind(ctx, "some-instance", "some-binding-id", bindDetails)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("contract_url not found"))
					})
				})

			})

			Context("when valid", func() {
				var (
					server                *ghttp.Server
					downloadContent       []byte
					expectedContainerInfo *containermanager.ContainerInfo
					expectedBindResponse  broker.BindResponse
				)

				Context("with the contract url", func() {
					BeforeEach(func() {
						server = ghttp.NewServer()
						urlStr := server.URL() + "/contract_path"

						jsonParams = fmt.Sprintf(` {
							"contract_url": "%s",
							"contract_args": ["arg1", "arg2"]
						} `, urlStr)

						downloadContent = []byte("some-content")

						header := http.Header{}
						server.AppendHandlers(ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/contract_path"),
							ghttp.RespondWith(http.StatusOK, downloadContent, header),
						))

						bindings := make(map[string][]containermanager.Binding)
						bindings["1234"] = []containermanager.Binding{
							containermanager.Binding{
								Port: "6789",
							},
						}

						expectedContainerInfo = &containermanager.ContainerInfo{
							InternalAddress: "127.0.0.1",
							ExternalAddress: "some-ip",
							Bindings:        bindings,
						}

						nodeInfo := &deployer.NodeInfo{
							Account: "some-account",
						}

						expectedBindResponse = broker.BindResponse{
							ContainerInfo: expectedContainerInfo,
							NodeInfo:      nodeInfo,
						}

						fakeManager.BindReturns(expectedContainerInfo, nil)
						fakeDeployer.DeployContractReturns(nodeInfo, nil)
					})

					It("should call the container fakeManager for bind", func() {
						_, err := blockhead.Bind(context.TODO(), "some-instance", "some-binding-id", bindDetails)
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeManager.BindCallCount()).To(Equal(1))
						_, bindConfig := fakeManager.BindArgsForCall(0)
						Expect(bindConfig).To(Equal(containermanager.BindConfig{
							InstanceId: "some-instance",
							BindingId:  "some-binding-id",
						}))
					})

					It("should download the contract and call the deployer with contract info", func() {
						_, err := blockhead.Bind(context.TODO(), "some-instance", "some-binding-id", bindDetails)
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeDeployer.DeployContractCallCount()).To(Equal(1))
						fakeDeployer.DeployContractStub = func(contractInfo *deployer.ContractInfo, containerInfo *containermanager.ContainerInfo) (*deployer.NodeInfo, error) {
							Expect(contractInfo.ContractPath).To(BeAnExistingFile())
							Expect(containerInfo).To(Equal(expectedContainerInfo))
							return nil, nil
						}
					})

					It("fills in the credentials in the api response", func() {
						resp, _ := blockhead.Bind(context.TODO(), "some-instance", "some-binding-id", bindDetails)
						Expect(resp.Credentials).To(Equal(expectedBindResponse))
					})
				})
			})

			Context("when contract_args is missing", func() {
				BeforeEach(func() {
					jsonParams = ` {
							"contract_url": "some-url"
						} `
				})

				It("succeeds calling the fakeManager's bind", func() {
					_, err := blockhead.Bind(ctx, "some-instance", "some-binding-id", bindDetails)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeManager.BindCallCount()).To(Equal(1))
					_, bindConfig := fakeManager.BindArgsForCall(0)
					Expect(bindConfig.InstanceId).To(Equal("some-instance"))
					Expect(bindConfig.BindingId).To(Equal("some-binding-id"))
				})
			})
		})

		It("Bubbles up errors from the container fakeManager", func() {
			errorMessage := "docker exploded"
			fakeManager.BindReturns(nil, errors.New(errorMessage))
			bindDetails := brokerapi.BindDetails{
				ServiceID:     "service-id",
				PlanID:        "plan-id",
				RawParameters: []byte(`{"contract_url": "some-url"}`),
			}
			_, err := blockhead.Bind(ctx, "some-instance", "some-binding-id", bindDetails)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
	})
})
