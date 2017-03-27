package legacynet

import (
	"fmt"
	"lib/rules"

	multierror "github.com/hashicorp/go-multierror"
)

const prefixDNS = "dns"

type DNS struct {
	ChainNamer chainNamer
	IPTables   rules.IPTablesAdapter
}

func (m *DNS) Initialize(dnsIP string) error {
	inputChain := m.ChainNamer.Prefix(prefixDNS, dnsIP)

	err := m.IPTables.NewChain("filter", inputChain)
	if err != nil {
		return fmt.Errorf("creating chain: %s", err)
	}

	err = m.IPTables.BulkInsert("filter", "INPUT", 1, rules.IPTablesRule{"--jump", inputChain})
	if err != nil {
		return fmt.Errorf("inserting rule: %s", err)
	}

	err = m.IPTables.BulkAppend("filter", inputChain, []rules.IPTablesRule{
		rules.NewInputRelatedEstablishedRule(dnsIP),
		rules.NewInputDefaultRejectRule(dnsIP),
	}...)
	if err != nil {
		return fmt.Errorf("appending rule: %s", err)
	}

	return nil
}

func (m *DNS) Cleanup(dnsIP string) error {
	inputChain := m.ChainNamer.Prefix(prefixDNS, dnsIP)

	var result error
	if err := cleanupChain("filter", "INPUT", inputChain, m.IPTables); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}
