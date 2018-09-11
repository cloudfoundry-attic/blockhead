package utils

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry-incubator/blockhead/pkg/config"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

type ServiceMatcher struct {
	service *config.Service
}

func EquivalentService(service *config.Service) types.GomegaMatcher {
	return &ServiceMatcher{
		service: service,
	}
}

func (m *ServiceMatcher) Match(actual interface{}) (success bool, err error) {
	otherService, ok := actual.(*config.Service)
	if !ok {
		return false, errors.New("Not a config.Service object")
	}

	if otherService.Name != m.service.Name {
		return false, fmt.Errorf("Service names do not match, actual: %s, expected: %s", otherService.Name, m.service.Name)
	}

	if otherService.Description != m.service.Description {
		return false, fmt.Errorf("Service descriptions do not match, actual: %s, expected: %s", otherService.Description, m.service.Description)
	}

	if otherService.DisplayName != m.service.DisplayName {
		return false, fmt.Errorf("Service display names do not match, actual: %s, expected: %s", otherService.Description, m.service.Description)
	}

	tagsMatcher := gomega.ConsistOf(m.service.Tags)
	successful, err := tagsMatcher.Match(otherService.Tags)
	if !successful {
		return false, fmt.Errorf("Services tags do not match, %s", err.Error())
	}

	plans := []*config.Plan{}
	for _, plan := range m.service.Plans {
		plans = append(plans, plan)
	}

	plansMatcher := gomega.ConsistOf(plans)
	return plansMatcher.Match(otherService.Plans)
}

func (m *ServiceMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be ", m.service)
}

func (m *ServiceMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("to not be"), m.service)
}
