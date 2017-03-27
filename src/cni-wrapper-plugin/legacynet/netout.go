package legacynet

import (
	"errors"
	"fmt"
	"lib/rules"
	"net"

	multierror "github.com/hashicorp/go-multierror"

	"code.cloudfoundry.org/garden"
)

const prefixInput = "input"
const prefixNetOut = "netout"
const suffixNetOutLog = "log"

//go:generate counterfeiter -o ../fakes/net_out_rule_converter.go --fake-name NetOutRuleConverter . netOutRuleConverter
type netOutRuleConverter interface {
	Convert(rule garden.NetOutRule, containerIP, logChainName string, logging bool) []rules.IPTablesRule
	BulkConvert(rules []garden.NetOutRule, containerIP, logChainName string, logging bool) []rules.IPTablesRule
}

type NetOut struct {
	ChainNamer    chainNamer
	IPTables      rules.IPTablesAdapter
	Converter     netOutRuleConverter
	GlobalLogging bool
}

func (m *NetOut) Initialize(containerHandle string, containerIP net.IP, overlayNetwork string, dnsServers []string) error {
	if containerHandle == "" {
		return errors.New("invalid handle")
	}

	inputChain := m.ChainNamer.Prefix(prefixInput, containerHandle)
	forwardChain := m.ChainNamer.Prefix(prefixNetOut, containerHandle)
	logChain, err := m.ChainNamer.Postfix(forwardChain, suffixNetOutLog)
	if err != nil {
		return fmt.Errorf("getting chain name: %s", err)
	}

	type rulesAndChain struct {
		ParentChain string
		Chain       string
		Rules       []rules.IPTablesRule
	}

	args := []rulesAndChain{
		{
			ParentChain: "INPUT",
			Chain:       inputChain,
			Rules: []rules.IPTablesRule{
				rules.NewInputRelatedEstablishedRule(containerIP.String()),
				rules.NewInputDefaultRejectRule(containerIP.String()),
			},
		},
		{
			ParentChain: "FORWARD",
			Chain:       forwardChain,
			Rules: []rules.IPTablesRule{
				rules.NewNetOutRelatedEstablishedRule(containerIP.String(), overlayNetwork),
				rules.NewNetOutDefaultRejectRule(containerIP.String(), overlayNetwork),
			},
		},
		{
			Chain: logChain,
			Rules: []rules.IPTablesRule{
				rules.NewNetOutDefaultLogRule(containerHandle),
				rules.NewReturnRule(),
			},
		},
	}

	if m.GlobalLogging {
		args[1].Rules = []rules.IPTablesRule{
			rules.NewNetOutRelatedEstablishedRule(containerIP.String(), overlayNetwork),
			rules.NewNetOutDefaultRejectLogRule(containerHandle, containerIP.String(), overlayNetwork),
			rules.NewNetOutDefaultRejectRule(containerIP.String(), overlayNetwork),
		}
	}

	if len(dnsServers) > 0 {
		args[0].Rules = []rules.IPTablesRule{
			rules.NewInputRelatedEstablishedRule(containerIP.String()),
		}
		for _, dnsServer := range dnsServers {
			args[0].Rules = append(args[0].Rules, rules.NewInputDNSRule(containerIP.String(), dnsServer))
		}
		args[0].Rules = append(args[0].Rules, rules.NewInputDefaultRejectRule(containerIP.String()))
	}

	for _, arg := range args {
		err = m.IPTables.NewChain("filter", arg.Chain)
		if err != nil {
			return fmt.Errorf("creating chain: %s", err)
		}

		if arg.ParentChain != "" {
			err = m.IPTables.BulkInsert("filter", arg.ParentChain, 1, rules.IPTablesRule{"--jump", arg.Chain})
			if err != nil {
				return fmt.Errorf("inserting rule: %s", err)
			}
		}

		err = m.IPTables.BulkAppend("filter", arg.Chain, arg.Rules...)
		if err != nil {
			return fmt.Errorf("appending rule: %s", err)
		}
	}

	return nil
}

func (m *NetOut) Cleanup(containerHandle string) error {
	forwardChain := m.ChainNamer.Prefix(prefixNetOut, containerHandle)
	inputChain := m.ChainNamer.Prefix(prefixInput, containerHandle)
	logChain, err := m.ChainNamer.Postfix(forwardChain, suffixNetOutLog)
	if err != nil {
		return fmt.Errorf("getting chain name: %s", err)
	}

	var result error
	if err := cleanupChain("filter", "FORWARD", forwardChain, m.IPTables); err != nil {
		result = multierror.Append(result, err)
	}
	if err := cleanupChain("filter", "INPUT", inputChain, m.IPTables); err != nil {
		result = multierror.Append(result, err)
	}
	if err := cleanupChain("filter", "", logChain, m.IPTables); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (m *NetOut) InsertRule(containerHandle string, rule garden.NetOutRule, containerIP string) error {
	chain := m.ChainNamer.Prefix(prefixNetOut, containerHandle)
	logChain, err := m.ChainNamer.Postfix(chain, suffixNetOutLog)
	if err != nil {
		return fmt.Errorf("getting chain name: %s", err)
	}

	ruleSpec := m.Converter.Convert(rule, containerIP, logChain, m.GlobalLogging)
	err = m.IPTables.BulkInsert("filter", chain, 1, ruleSpec...)
	if err != nil {
		return fmt.Errorf("inserting net-out rule: %s", err)
	}

	return nil
}

func (m *NetOut) BulkInsertRules(containerHandle string, netOutRules []garden.NetOutRule, containerIP string) error {
	chain := m.ChainNamer.Prefix(prefixNetOut, containerHandle)
	logChain, err := m.ChainNamer.Postfix(chain, suffixNetOutLog)
	if err != nil {
		return fmt.Errorf("getting chain name: %s", err)
	}

	ruleSpec := m.Converter.BulkConvert(netOutRules, containerIP, logChain, m.GlobalLogging)
	err = m.IPTables.BulkInsert("filter", chain, 1, ruleSpec...)
	if err != nil {
		return fmt.Errorf("bulk inserting net-out rules: %s", err)
	}

	return nil
}
