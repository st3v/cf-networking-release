package legacynet

import (
	"lib/rules"

	"code.cloudfoundry.org/garden"
)

type NetOutRuleConverter struct {
}

func (c *NetOutRuleConverter) BulkConvert(netOutRules []garden.NetOutRule, containerIP, logChainName string, globalLogging bool) []rules.IPTablesRule {
	ruleSpec := []rules.IPTablesRule{}
	for _, rule := range netOutRules {
		for _, t := range c.Convert(rule, containerIP, logChainName, globalLogging) {
			ruleSpec = append(ruleSpec, t)
		}
	}
	return ruleSpec
}

func (c *NetOutRuleConverter) Convert(rule garden.NetOutRule, containerIP, logChainName string, globalLogging bool) []rules.IPTablesRule {
	ruleSpec := []rules.IPTablesRule{}
	for _, network := range rule.Networks {
		if len(rule.Ports) > 0 && udpOrTcp(rule.Protocol) {
			for _, portRange := range rule.Ports {
				if rule.Log || globalLogging {
					ruleSpec = append(ruleSpec, rules.NewNetOutWithPortsLogRule(
						containerIP, network.Start.String(), network.End.String(),
						int(portRange.Start), int(portRange.End), lookupProtocol(rule.Protocol), logChainName),
					)
				} else {
					ruleSpec = append(ruleSpec, rules.NewNetOutWithPortsRule(
						containerIP, network.Start.String(), network.End.String(),
						int(portRange.Start), int(portRange.End), lookupProtocol(rule.Protocol)),
					)
				}
			}
		} else {
			if rule.Log || globalLogging {
				ruleSpec = append(ruleSpec, rules.NewNetOutLogRule(
					containerIP, network.Start.String(), network.End.String(), logChainName),
				)
			} else {
				ruleSpec = append(ruleSpec, rules.NewNetOutRule(
					containerIP, network.Start.String(), network.End.String()),
				)
			}
		}
	}
	return ruleSpec
}

func udpOrTcp(protocol garden.Protocol) bool {
	return protocol == garden.ProtocolTCP || protocol == garden.ProtocolUDP
}

func lookupProtocol(protocol garden.Protocol) string {
	switch protocol {
	case garden.ProtocolTCP:
		return "tcp"
	case garden.ProtocolUDP:
		return "udp"
	default:
		return "all"
	}
}
