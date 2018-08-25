package utils

import (
	"errors"
	"fmt"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/pivotal-cf/brokerapi"
)

type BrokerAPIServiceMatcher struct {
	service brokerapi.Service
}

func EquivalentBrokerAPIService(service brokerapi.Service) types.GomegaMatcher {
	return &BrokerAPIServiceMatcher{
		service: service,
	}
}

func (m *BrokerAPIServiceMatcher) Match(actual interface{}) (success bool, err error) {
	service2, ok := actual.(brokerapi.Service)
	if !ok {
		return false, errors.New("Not a brokerapi.Service object")
	}

	if service2.Name != m.service.Name {
		return false, fmt.Errorf("Service names do not match, actual: %s, expected: %s", service2.Name, m.service.Name)
	}

	if service2.Description != m.service.Description {
		return false, fmt.Errorf("Service descriptions do not match, actual: %s, expected: %s", service2.Description, m.service.Description)
	}

	if service2.Bindable != m.service.Bindable {
		return false, fmt.Errorf("Service field Bindable do not match, actual: %t, expected: %t", service2.Bindable, m.service.Bindable)
	}

	metadataMatcher := gomega.Equal(m.service.Metadata)
	successful, err := metadataMatcher.Match(service2.Metadata)
	if !successful {
		return false, fmt.Errorf("Service Metadata do not match %s", err.Error())
	}

	tagsMatcher := gomega.ConsistOf(m.service.Tags)
	successful, err = tagsMatcher.Match(service2.Tags)
	if !successful {
		return false, fmt.Errorf("Services tags do not match, %s", err.Error())
	}

	if service2.PlanUpdatable != m.service.PlanUpdatable {
		return false, fmt.Errorf("Service field Plan Updatable do not match, actual: %t, expected: %t", service2.PlanUpdatable, m.service.PlanUpdatable)
	}

	if service2.DashboardClient != m.service.DashboardClient {
		return false, fmt.Errorf("Service DashboardClient do not match, actual: %+v, expected: %+v", service2.Metadata, m.service.Metadata)
	}

	plans := []types.GomegaMatcher{}
	for _, plan := range m.service.Plans {
		p := EquivalentBrokerAPIPlan(plan)
		plans = append(plans, p)
	}

	plansMatcher := gomega.ConsistOf(plans)
	return plansMatcher.Match(service2.Plans)
}

func (m *BrokerAPIServiceMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be ", m.service)
}

func (m *BrokerAPIServiceMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("to not be"), m.service)
}

type BrokerAPIPlanMatcher struct {
	plan brokerapi.ServicePlan
}

func EquivalentBrokerAPIPlan(plan brokerapi.ServicePlan) types.GomegaMatcher {
	return &BrokerAPIPlanMatcher{
		plan: plan,
	}
}

func (m *BrokerAPIPlanMatcher) Match(actual interface{}) (success bool, err error) {
	plan2, ok := actual.(brokerapi.ServicePlan)
	if !ok {
		return false, errors.New("Not a brokerapi.Service object")
	}

	if plan2.Name != m.plan.Name {
		return false, fmt.Errorf("Plan names do not match, actual: %s, expected: %s", plan2.Name, m.plan.Name)
	}

	if plan2.Description != m.plan.Description {
		return false, fmt.Errorf("Plan descriptions do not match, actual: %s, expected: %s", plan2.Description, m.plan.Description)
	}

	freeMatcher := gomega.BeEquivalentTo(m.plan.Free)
	successful, err := freeMatcher.Match(plan2.Free)
	if !successful {
		return false, fmt.Errorf("Plan field Free do not match %s", err.Error())
	}

	metadataMatcher := gomega.Equal(m.plan.Metadata)
	successful, err = metadataMatcher.Match(plan2.Metadata)
	if !successful {
		return false, fmt.Errorf("Plan Metadata do not match %s", err.Error())
	}

	bindableMatcher := gomega.BeEquivalentTo(m.plan.Bindable)
	successful, err = bindableMatcher.Match(plan2.Bindable)
	if !successful {
		return false, fmt.Errorf("Plan field Bindable do not match %s", err.Error())
	}

	schemaMatcher := gomega.Equal(m.plan.Schemas)
	successful, err = schemaMatcher.Match(plan2.Schemas)
	if !successful {
		return false, fmt.Errorf("Plan Schemas do not match, %s", err.Error())
	}

	return true, nil
}

func (m *BrokerAPIPlanMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be ", m.plan)
}

func (m *BrokerAPIPlanMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("to not be"), m.plan)
}
