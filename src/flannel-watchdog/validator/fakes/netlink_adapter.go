// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/vishvananda/netlink"
)

type NetlinkAdapter struct {
	LinkByNameStub        func(string) (netlink.Link, error)
	linkByNameMutex       sync.RWMutex
	linkByNameArgsForCall []struct {
		arg1 string
	}
	linkByNameReturns struct {
		result1 netlink.Link
		result2 error
	}
	AddrListStub        func(netlink.Link, int) ([]netlink.Addr, error)
	addrListMutex       sync.RWMutex
	addrListArgsForCall []struct {
		arg1 netlink.Link
		arg2 int
	}
	addrListReturns struct {
		result1 []netlink.Addr
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *NetlinkAdapter) LinkByName(arg1 string) (netlink.Link, error) {
	fake.linkByNameMutex.Lock()
	fake.linkByNameArgsForCall = append(fake.linkByNameArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("LinkByName", []interface{}{arg1})
	fake.linkByNameMutex.Unlock()
	if fake.LinkByNameStub != nil {
		return fake.LinkByNameStub(arg1)
	}
	return fake.linkByNameReturns.result1, fake.linkByNameReturns.result2
}

func (fake *NetlinkAdapter) LinkByNameCallCount() int {
	fake.linkByNameMutex.RLock()
	defer fake.linkByNameMutex.RUnlock()
	return len(fake.linkByNameArgsForCall)
}

func (fake *NetlinkAdapter) LinkByNameArgsForCall(i int) string {
	fake.linkByNameMutex.RLock()
	defer fake.linkByNameMutex.RUnlock()
	return fake.linkByNameArgsForCall[i].arg1
}

func (fake *NetlinkAdapter) LinkByNameReturns(result1 netlink.Link, result2 error) {
	fake.LinkByNameStub = nil
	fake.linkByNameReturns = struct {
		result1 netlink.Link
		result2 error
	}{result1, result2}
}

func (fake *NetlinkAdapter) AddrList(arg1 netlink.Link, arg2 int) ([]netlink.Addr, error) {
	fake.addrListMutex.Lock()
	fake.addrListArgsForCall = append(fake.addrListArgsForCall, struct {
		arg1 netlink.Link
		arg2 int
	}{arg1, arg2})
	fake.recordInvocation("AddrList", []interface{}{arg1, arg2})
	fake.addrListMutex.Unlock()
	if fake.AddrListStub != nil {
		return fake.AddrListStub(arg1, arg2)
	}
	return fake.addrListReturns.result1, fake.addrListReturns.result2
}

func (fake *NetlinkAdapter) AddrListCallCount() int {
	fake.addrListMutex.RLock()
	defer fake.addrListMutex.RUnlock()
	return len(fake.addrListArgsForCall)
}

func (fake *NetlinkAdapter) AddrListArgsForCall(i int) (netlink.Link, int) {
	fake.addrListMutex.RLock()
	defer fake.addrListMutex.RUnlock()
	return fake.addrListArgsForCall[i].arg1, fake.addrListArgsForCall[i].arg2
}

func (fake *NetlinkAdapter) AddrListReturns(result1 []netlink.Addr, result2 error) {
	fake.AddrListStub = nil
	fake.addrListReturns = struct {
		result1 []netlink.Addr
		result2 error
	}{result1, result2}
}

func (fake *NetlinkAdapter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.linkByNameMutex.RLock()
	defer fake.linkByNameMutex.RUnlock()
	fake.addrListMutex.RLock()
	defer fake.addrListMutex.RUnlock()
	return fake.invocations
}

func (fake *NetlinkAdapter) recordInvocation(key string, args []interface{}) {
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
