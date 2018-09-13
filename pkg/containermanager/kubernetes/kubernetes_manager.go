package kubernetes

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/blockhead/pkg/containermanager"
	"k8s.io/client-go/kubernetes"
)

type kubernetesContainerManager struct {
	client kubernetes.Interface
	logger lager.Logger
}

func NewKubernetesContainerManager(logger lager.Logger, client kubernetes.Interface) containermanager.ContainerManager {
	return kubernetesContainerManager{
		client: client,
		logger: logger.Session("kubernetes-container-manager"),
	}
}

func (kc kubernetesContainerManager) Provision(ctx context.Context, cc containermanager.ContainerConfig) error {
	return fmt.Errorf("kube provision unimplemented")
}

func (kc kubernetesContainerManager) Deprovision(ctx context.Context, instanceID string) error {
	return fmt.Errorf("kube deprovision unimplemented")
}
