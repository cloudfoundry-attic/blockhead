package kubernetes

import (
	"context"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type kubernetesContainerManager struct {
	client kubernetes.Interface
	host   string
	logger lager.Logger
}

func NewKubernetesContainerManager(logger lager.Logger, client kubernetes.Interface, host string) containermanager.ContainerManager {
	return kubernetesContainerManager{
		client: client,
		host:   host,
		logger: logger.Session("kubernetes-container-manager"),
	}
}

func (kc kubernetesContainerManager) Provision(ctx context.Context, cc containermanager.ContainerConfig) error {
	var err error
	selector := make(map[string]string)
	selector["app"] = cc.Name
	selector["provisionedBy"] = "blockhead-broker"

	blockheadNamespace := v1.NamespaceDefault // put everything in default? our own ns?

	// create a pod with a label
	pod := &v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      cc.Name,
			Namespace: blockheadNamespace,
			Labels:    selector,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:  cc.Name,
					Image: cc.Image,
				},
			},
		},
	}

	_, err = kc.client.CoreV1().Pods(v1.NamespaceDefault).Create(pod)
	if err != nil {
		kc.logger.Error("error creating pod", err)
	}

	servicePorts := []v1.ServicePort{}
	for _, p := range cc.ExposedPorts {
		kc.logger.Debug("adding exposed port " + p)
		port, err := strconv.Atoi(p)
		if err != nil {
			return err
		}
		sp := v1.ServicePort{
			Port: int32(port),
			// we're not choosing a target port here. It will thus
			// automatically be chosen when the service is created.
		}
		servicePorts = append(servicePorts, sp)
	}

	// create a service with the same label and a label selector of the label
	// nodeport for now?
	svc := &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "B" + cc.Name,
			Namespace: blockheadNamespace,
			Labels:    selector,
		},
		Spec: v1.ServiceSpec{
			// ports, get from passed in
			Ports:    servicePorts,
			Selector: selector, // selector use same as above
			// service type is either nodeport or loadbalancer if we can
			Type: v1.ServiceTypeNodePort,
		},
	}
	kc.logger.Debug(fmt.Sprintf("%++v\n", svc))
	_, err = kc.client.CoreV1().Services(v1.NamespaceDefault).Create(svc)
	if err != nil {
		kc.logger.Error("error creating service", err)
	}

	return nil
}

func (kc kubernetesContainerManager) Deprovision(ctx context.Context, instanceID string) error {
	var err error
	// TODO: investigate delete everything with a label selector
	// everything uses the instance name so we don't have to track it in storage yet.
	// ignore all the errors, because it doesn't hurt to fail to delete things. Log them instead.
	err = kc.client.CoreV1().Pods(v1.NamespaceDefault).Delete(instanceID, &meta_v1.DeleteOptions{})
	if err != nil {
		kc.logger.Info(err.Error())
	}
	err = kc.client.CoreV1().Services(v1.NamespaceDefault).Delete("B" + instanceID, &meta_v1.DeleteOptions{})
	if err != nil {
		kc.logger.Info(err.Error())
	}
	return nil
}

func (kc kubernetesContainerManager) Bind(cts context.Context, bc containermanager.BindConfig) (*containermanager.ContainerInfo, error) {

	// pod and service both with instance name
	instanceName := bc.InstanceId

	// take each port binding and find the nodeport that it is bound too.
	instanceservice, err := kc.client.CoreV1().Services(v1.NamespaceDefault).Get("B" + instanceName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	ports := instanceservice.Spec.Ports

	bindings := make(map[string][]containermanager.Binding)
	for _, port := range ports {
		p := strconv.Itoa(int(port.NodePort))

		containerBindings := []containermanager.Binding{}
		containerBindings = append(containerBindings, containermanager.Binding{
			Port: p,
		})
		bindings[strconv.Itoa(int(port.Port))] = containerBindings
	}

	response := containermanager.ContainerInfo{
		ExternalAddress: kc.host,
		InternalAddress: instanceName,
		Bindings:        bindings,
	}

	return &response, nil
}
