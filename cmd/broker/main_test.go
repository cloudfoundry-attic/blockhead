package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Blockhead", func() {
	var (
		session *gexec.Session
		args    []string
		err     error
	)
	Context("when no args are passed in ", func() {
		BeforeEach(func() {
			args = []string{}
		})

		It("errors", func() {
			cmd := exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode).ToNot(Equal(0))
		})
	})

	Context("when only one arg is passed in ", func() {
		BeforeEach(func() {
			args = []string{
				configPath,
			}
		})

		It("errors", func() {
			cmd := exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode).ToNot(Equal(0))
		})
	})

	Context("when both args are passed in", func() {
		BeforeEach(func() {
			args = []string{
				configPath,
				servicePath,
			}
		})

		It("does not error", func() {
			cmd := exec.Command(brokerBinPath, args...)
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
