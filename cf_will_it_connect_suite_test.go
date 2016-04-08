package main_test

import (
	"github.com/cloudfoundry/cli/testhelpers/plugin_builder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfWillItConnect(t *testing.T) {

	RegisterFailHandler(Fail)

	plugin_builder.BuildTestBinary("", "cf_will_it_connect")
	RunSpecs(t, "CfWillItConnect Suite")
}
