// This file was generated by counterfeiter
package fakes

import (
	"policy-server/models"
	"policy-server/uaa_client"
	"sync"
)

type PolicyFilter struct {
	FilterPoliciesStub        func(policies []models.Policy, userToken uaa_client.CheckTokenResponse) ([]models.Policy, error)
	filterPoliciesMutex       sync.RWMutex
	filterPoliciesArgsForCall []struct {
		policies  []models.Policy
		userToken uaa_client.CheckTokenResponse
	}
	filterPoliciesReturns struct {
		result1 []models.Policy
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *PolicyFilter) FilterPolicies(policies []models.Policy, userToken uaa_client.CheckTokenResponse) ([]models.Policy, error) {
	var policiesCopy []models.Policy
	if policies != nil {
		policiesCopy = make([]models.Policy, len(policies))
		copy(policiesCopy, policies)
	}
	fake.filterPoliciesMutex.Lock()
	fake.filterPoliciesArgsForCall = append(fake.filterPoliciesArgsForCall, struct {
		policies  []models.Policy
		userToken uaa_client.CheckTokenResponse
	}{policiesCopy, userToken})
	fake.recordInvocation("FilterPolicies", []interface{}{policiesCopy, userToken})
	fake.filterPoliciesMutex.Unlock()
	if fake.FilterPoliciesStub != nil {
		return fake.FilterPoliciesStub(policies, userToken)
	}
	return fake.filterPoliciesReturns.result1, fake.filterPoliciesReturns.result2
}

func (fake *PolicyFilter) FilterPoliciesCallCount() int {
	fake.filterPoliciesMutex.RLock()
	defer fake.filterPoliciesMutex.RUnlock()
	return len(fake.filterPoliciesArgsForCall)
}

func (fake *PolicyFilter) FilterPoliciesArgsForCall(i int) ([]models.Policy, uaa_client.CheckTokenResponse) {
	fake.filterPoliciesMutex.RLock()
	defer fake.filterPoliciesMutex.RUnlock()
	return fake.filterPoliciesArgsForCall[i].policies, fake.filterPoliciesArgsForCall[i].userToken
}

func (fake *PolicyFilter) FilterPoliciesReturns(result1 []models.Policy, result2 error) {
	fake.FilterPoliciesStub = nil
	fake.filterPoliciesReturns = struct {
		result1 []models.Policy
		result2 error
	}{result1, result2}
}

func (fake *PolicyFilter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.filterPoliciesMutex.RLock()
	defer fake.filterPoliciesMutex.RUnlock()
	return fake.invocations
}

func (fake *PolicyFilter) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
