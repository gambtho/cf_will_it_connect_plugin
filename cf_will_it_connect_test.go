package main_test

import (
	"errors"

	"gopkg.in/h2non/gock.v0"

	"github.com/cloudfoundry/cli/plugin/models"
	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/gambtho/cf_will_it_connect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const apiURL string = "http://api.cfapps.pivotal.io"
const wicPath string = "/v2/willitconnect"
const wicURL string = "https://willitconnect.cfapps.io"

var _ = Describe("CfWillItConnect", func() {

	Describe(".Run", func() {
		var fakeCliConnection *pluginfakes.FakeCliConnection
		var willItConnectPlugin *WillItConnect

		BeforeEach(func() {
			fakeCliConnection = &pluginfakes.FakeCliConnection{}
			willItConnectPlugin = &WillItConnect{}
		})

		It("displays a usage message when called with too few arguments", func() {
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "blah"})
			})
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "willitconnect", "<host>", "<port>"}))
		})

		It("displays an error when the CF api is unavailable", func() {

			fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{}, errors.New("No org!"))
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Unable to connect to CF, use cf login first"}))
		})

		It("displays an error when no domains are available", func() {
			fakeDomain := []plugin_models.GetOrg_Domains{plugin_models.GetOrg_Domains{}}
			fakeOrg := plugin_models.GetOrg_Model{Domains: fakeDomain}
			fakeCliConnection.GetOrgReturns(fakeOrg, errors.New("Org is a lie!"))
			fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Unable to find valid org, please view cf target"}))
		})

		It("displays an error when no domains are available", func() {
			fakeDomain := []plugin_models.GetOrg_Domains{plugin_models.GetOrg_Domains{}}
			fakeOrg := plugin_models.GetOrg_Model{Domains: fakeDomain}
			fakeCliConnection.GetOrgReturns(fakeOrg, nil)
			fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Unable to find valid domain, please view cf domains"}))
		})

		It("displays the host, port, api, and a connect confirmation", func() {

			fakeDomain := []plugin_models.GetOrg_Domains{plugin_models.GetOrg_Domains{Name: "cfapps.io"}}
			fakeOrg := plugin_models.GetOrg_Model{Domains: fakeDomain}
			fakeCliConnection.GetOrgReturns(fakeOrg, nil)
			fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)

			defer gock.Off()

			gock.New(wicURL).
				Post(wicPath).
				JSON(`{"target":"foo.com:80"}`).
				Reply(200).
				JSON(`{"lastChecked": 0, "entry": "foo.com", "canConnect": true, "httpStatus": 200, "validHostname": false, "validUrl": true}`)
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Host:", "foo.com", "-",
				"Port:", "80", "-",
				"WillItConnect:", wicURL + wicPath}))
			Expect(output).To(ContainSubstrings([]string{"I am able to connect"}))
		})

		It("displays the host, port, api, and an unable to connect failure", func() {

			fakeDomain := []plugin_models.GetOrg_Domains{plugin_models.GetOrg_Domains{Name: "cfapps.io"}}
			fakeOrg := plugin_models.GetOrg_Model{Domains: fakeDomain}
			fakeCliConnection.GetOrgReturns(fakeOrg, nil)
			fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)

			defer gock.Off()

			gock.New(wicURL).
				Post(wicPath).
				JSON(`{"target":"bar.com:80"}`).
				Reply(200).
				JSON(`{"lastChecked": 0,"entry": "bar.com","canConnect":false,"validHostname": false,"validUrl": true}`)

			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "bar.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Host:", "bar.com", "-",
				"Port:", "80", "-",
				"WillItConnect:", wicURL + wicPath}))
			Expect(output).To(ContainSubstrings([]string{"I am unable to connect"}))
		})
	})
})
