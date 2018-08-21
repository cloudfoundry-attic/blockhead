package config_test

import (
	"github.com/cloudfoundry-incubator/blockhead/pkg/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("Opens the file and pours it into a Config struct", func() {
			config, err := config.NewConfig("assets/test_config.json")

			Expect(err).NotTo(HaveOccurred())
			Expect(config.Username).To(Equal("username"))
			Expect(config.Password).To(Equal("password"))
		})

		It("errors when provided a nonexistent config file", func() {
			cfg, err := config.NewConfig("not/a/real/file.json")
			Expect(cfg).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error opening config file"))
		})

		It("errors when provided a bad config file", func() {
			config, err := config.NewConfig("assets/bad_config.json")
			Expect(config).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error parsing config file"))
		})
	})
})
