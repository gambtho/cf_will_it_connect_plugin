package main_test

import (
	"errors"

	"github.com/cloudfoundry/cli/plugin/models"
	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/gambtho/cf_will_it_connect_plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v0"
)

const apiURL string = "http://api.cfapps.pivotal.io"
const wicPath string = "/v2/willitconnect"
const wicURL string = "https://willitconnect.cfapps.io"
const goodRequest string = `{"target":"foo.com:80"}`
const goodResponse string = `{"lastChecked": 0, "entry": "foo.com", "canConnect": true, "httpStatus": 200, "validHostname": false, "validUrl": true}`
const badRequest string = `{"target":"bar.com:80"}`
const badResponse string = `{"lastChecked": 0,"entry": "bar.com","canConnect":false,"validHostname": false,"validUrl": true}`

var _ = Describe("CfWillItConnect", func() {

	Describe("is run", func() {
		var fakeCliConnection *pluginfakes.FakeCliConnection
		var willItConnectPlugin *WillItConnect
		var goodOrg plugin_models.GetOrg_Model
		var badOrg plugin_models.GetOrg_Model

		BeforeEach(func() {
			fakeCliConnection = &pluginfakes.FakeCliConnection{}
			willItConnectPlugin = &WillItConnect{}
			badOrg = plugin_models.GetOrg_Model{Domains: []plugin_models.GetOrg_Domains{plugin_models.GetOrg_Domains{}}}
			goodOrg = plugin_models.GetOrg_Model{Domains: []plugin_models.GetOrg_Domains{plugin_models.GetOrg_Domains{Name: "cfapps.io"}}}
		})

		Context("only one argument is provided", func() {

			It("displays a usage message", func() {

				output := CaptureOutput(func() {
					willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "blah"})
				})
				Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "willitconnect", "<host>", "<port>"}))
			})
		})

		Context("the CF org is unavailable", func() {

			It("displays an error message", func() {

				fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{}, errors.New("No org!"))
				output := CaptureOutput(func() {
					willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
				})
				Expect(output).To(ContainSubstrings([]string{"Unable to connect to CF, use cf login first"}))
			})
		})

		Context("A domain is not returned", func() {

			It("displays an error when an error is thrown getting the org ", func() {
				fakeCliConnection.GetOrgReturns(badOrg, errors.New("Org is a lie!"))
				fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)
				output := CaptureOutput(func() {
					willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
				})
				Expect(output).To(ContainSubstrings([]string{"Unable to find valid org, please view cf target"}))
			})

			It("displays an error when no domains are available", func() {
				fakeCliConnection.GetOrgReturns(badOrg, nil)
				fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)
				output := CaptureOutput(func() {
					willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
				})
				Expect(output).To(ContainSubstrings([]string{"Unable to find valid domain, please view cf domains"}))
			})

		})

		Context("all required CF information is available", func() {

			Context("willitconnect application is unavailable", func() {

				It("displays an error indicating it can't reach willitconnect", func() {
					fakeCliConnection.GetOrgReturns(goodOrg, nil)
					fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)
					defer gock.Off()
					gock.New(wicURL).
						Post("/blah").
						JSON(goodRequest).
						Reply(404).
						JSON(`{}`)
					output := CaptureOutput(func() {
						willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
					})
					Expect(output).To(ContainSubstrings([]string{"Host:", "foo.com", "-",
						"Port:", "80", "-",
						"WillItConnect:", wicURL + wicPath}))
					Expect(output).To(ContainSubstrings([]string{"Unable to access"}))
				})
			})

			Context("willitconnect returns a bad response", func() {

				It("displays an error indicating that the response was bad", func() {
					fakeCliConnection.GetOrgReturns(goodOrg, nil)
					fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)
					defer gock.Off()
					gock.New(wicURL).
						Post(wicPath).
						JSON(goodRequest).
						Reply(200).
						BodyString("totes")
					output := CaptureOutput(func() {
						willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
					})
					Expect(output).To(ContainSubstrings([]string{"Host:", "foo.com", "-",
						"Port:", "80", "-",
						"WillItConnect:", wicURL + wicPath}))
					Expect(output).To(ContainSubstrings([]string{"Invalid response from willitconnect:"}))
				})
			})

			Context("a reachable host is provided", func() {

				It("displays the host, port, api, and a connect confirmation", func() {
					fakeCliConnection.GetOrgReturns(goodOrg, nil)
					fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)
					defer gock.Off()
					gock.New(wicURL).
						Post(wicPath).
						JSON(goodRequest).
						Reply(200).
						JSON(goodResponse)
					output := CaptureOutput(func() {
						willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
					})
					Expect(output).To(ContainSubstrings([]string{"Host:", "foo.com", "-",
						"Port:", "80", "-",
						"WillItConnect:", wicURL + wicPath}))
					Expect(output).To(ContainSubstrings([]string{"I am able to connect"}))
				})
			})
			Context("an unreachable host is provided", func() {

				It("displays the host, port, api, and an unable to connect failure", func() {
					fakeCliConnection.GetOrgReturns(goodOrg, nil)
					fakeCliConnection.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "org"}}, nil)

					defer gock.Off()

					gock.New(wicURL).
						Post(wicPath).
						JSON(badRequest).
						Reply(200).
						JSON(badResponse)

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
	})
})
