package config_test

import (
	"github.com/SpringerPE/graphite-du-report/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Describe("Default configuration", func() {
		It("default config contains the expected parameters", func() {
			c := config.DefaultConfig()
			Expect(c.Servers).To(Equal([]string{"127.0.0.1:8080"}))
		})
	})

	Describe("Given a string consisting of comma separated server strings", func() {
		It("returns a list of the servers", func() {
			s := "example.host:8080, 127.0.0.1:7777 , my.host:80"
			sList := config.ParseServerList(s)
			Expect(sList).To(Equal([]string{"example.host:8080","127.0.0.1:7777","my.host:80"}))
		})
	})
})
