package legacynet_test

import (
	"cni-wrapper-plugin/fakes"
	"cni-wrapper-plugin/legacynet"
	"errors"

	lib_fakes "lib/fakes"
	"lib/rules"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS", func() {
	var (
		dns        *legacynet.DNS
		chainNamer *fakes.ChainNamer
		ipTables   *lib_fakes.IPTablesAdapter
	)
	BeforeEach(func() {
		chainNamer = &fakes.ChainNamer{}
		ipTables = &lib_fakes.IPTablesAdapter{}
		dns = &legacynet.DNS{
			ChainNamer: chainNamer,
			IPTables:   ipTables,
		}
		chainNamer.PrefixStub = func(prefix, handle string) string {
			return prefix + "-" + handle
		}
		chainNamer.PostfixReturns("some-other-chain-name", nil)
	})

	Describe("Initialize", func() {
		It("creates the input chain", func() {
			err := dns.Initialize("5.6.7.8")
			Expect(err).NotTo(HaveOccurred())

			Expect(chainNamer.PrefixCallCount()).To(Equal(1))
			prefix, handle := chainNamer.PrefixArgsForCall(0)
			Expect(prefix).To(Equal("dns"))
			Expect(handle).To(Equal("5.6.7.8"))

			Expect(ipTables.NewChainCallCount()).To(Equal(1))
			table, chain := ipTables.NewChainArgsForCall(0)
			Expect(table).To(Equal("filter"))
			Expect(chain).To(Equal("dns-5.6.7.8"))
		})

		It("inserts a jump rule for the new chains", func() {
			err := dns.Initialize("5.6.7.8")
			Expect(err).NotTo(HaveOccurred())

			Expect(ipTables.BulkInsertCallCount()).To(Equal(1))
			table, chain, position, rulespec := ipTables.BulkInsertArgsForCall(0)
			Expect(table).To(Equal("filter"))
			Expect(chain).To(Equal("INPUT"))
			Expect(position).To(Equal(1))
			Expect(rulespec).To(Equal([]rules.IPTablesRule{{"--jump", "dns-5.6.7.8"}}))

		})

		It("writes the default netout and logging rules", func() {
			err := dns.Initialize("5.6.7.8")
			Expect(err).NotTo(HaveOccurred())

			Expect(ipTables.BulkAppendCallCount()).To(Equal(1))

			table, chain, rulespec := ipTables.BulkAppendArgsForCall(0)
			Expect(table).To(Equal("filter"))
			Expect(chain).To(Equal("dns-5.6.7.8"))
			Expect(rulespec).To(Equal([]rules.IPTablesRule{
				{"-s", "5.6.7.8",
					"-m", "state", "--state", "RELATED,ESTABLISHED",
					"--jump", "RETURN"},
				{"-s", "5.6.7.8",
					"--jump", "REJECT",
					"--reject-with", "icmp-port-unreachable"},
			}))
		})

		Context("when creating a new chain fails", func() {
			BeforeEach(func() {
				ipTables.NewChainReturns(errors.New("potata"))
			})
			It("returns the error", func() {
				err := dns.Initialize("5.6.7.8")
				Expect(err).To(MatchError("creating chain: potata"))
			})
		})

		Context("when inserting the rule that jumps to the new DNS chain fails", func() {
			BeforeEach(func() {
				ipTables.BulkInsertReturns(errors.New("potato"))
			})
			It("returns the error", func() {
				err := dns.Initialize("5.6.7.8")
				Expect(err).To(MatchError("inserting rule: potato"))
			})
		})

		Context("when writing the rules into the new DNS chain fails", func() {
			BeforeEach(func() {
				ipTables.BulkAppendReturns(errors.New("potato"))
			})
			It("returns the error", func() {
				err := dns.Initialize("5.6.7.8")
				Expect(err).To(MatchError("appending rule: potato"))
			})
		})
	})

	Describe("Cleanup", func() {
		It("deletes the jump rule from the input chain", func() {
			err := dns.Cleanup("5.6.7.8")
			Expect(err).NotTo(HaveOccurred())

			Expect(chainNamer.PrefixCallCount()).To(Equal(1))
			prefix, handle := chainNamer.PrefixArgsForCall(0)
			Expect(prefix).To(Equal("dns"))
			Expect(handle).To(Equal("5.6.7.8"))

			Expect(ipTables.DeleteCallCount()).To(Equal(1))
			table, chain, extraArgs := ipTables.DeleteArgsForCall(0)
			Expect(table).To(Equal("filter"))
			Expect(chain).To(Equal("INPUT"))
			Expect(extraArgs).To(Equal(rules.IPTablesRule{"--jump", "dns-5.6.7.8"}))
		})

		It("clears the dns chain", func() {
			err := dns.Cleanup("5.6.7.8")
			Expect(err).NotTo(HaveOccurred())

			Expect(ipTables.ClearChainCallCount()).To(Equal(1))
			table, chain := ipTables.ClearChainArgsForCall(0)
			Expect(table).To(Equal("filter"))
			Expect(chain).To(Equal("dns-5.6.7.8"))
		})

		It("deletes the dns chain", func() {
			err := dns.Cleanup("5.6.7.8")
			Expect(err).NotTo(HaveOccurred())

			Expect(ipTables.DeleteChainCallCount()).To(Equal(1))
			table, chain := ipTables.DeleteChainArgsForCall(0)
			Expect(table).To(Equal("filter"))
			Expect(chain).To(Equal("dns-5.6.7.8"))
		})

		Context("when deleting the jump rule fails", func() {
			BeforeEach(func() {
				ipTables.DeleteReturns(errors.New("yukon potato"))
			})
			It("returns an error", func() {
				err := dns.Cleanup("5.6.7.8")
				Expect(err).To(MatchError(ContainSubstring("delete rule: yukon potato")))
			})
		})

		Context("when clearing the dns chain fails", func() {
			BeforeEach(func() {
				ipTables.ClearChainReturns(errors.New("idaho potato"))
			})
			It("returns an error", func() {
				err := dns.Cleanup("5.6.7.8")
				Expect(err).To(MatchError(ContainSubstring("clear chain: idaho potato")))
			})
		})

		Context("when deleting the dns chain fails", func() {
			BeforeEach(func() {
				ipTables.DeleteChainReturns(errors.New("purple potato"))
			})
			It("returns an error", func() {
				err := dns.Cleanup("some-container-handle")
				Expect(err).To(MatchError(ContainSubstring("delete chain: purple potato")))
			})
		})

		Context("when all the steps fail", func() {
			BeforeEach(func() {
				ipTables.DeleteReturns(errors.New("yukon potato"))
				ipTables.ClearChainReturns(errors.New("idaho potato"))
				ipTables.DeleteChainReturns(errors.New("purple potato"))
			})
			It("returns all the errors", func() {
				err := dns.Cleanup("some-container-handle")
				Expect(err).To(MatchError(ContainSubstring("delete rule: yukon potato")))
				Expect(err).To(MatchError(ContainSubstring("clear chain: idaho potato")))
				Expect(err).To(MatchError(ContainSubstring("delete chain: purple potato")))
			})
		})
	})
})
