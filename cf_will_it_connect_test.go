package main_test

import (
	"errors"

	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/gambtho/cf_will_it_connect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v0"
)

const apiURL string = "http://api.cfapps.pivotal.io"
const wicPath string = "/v2/willitconnect"
const wicURL string = "http://willitconnect.cfapps.pivotal.io"

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
			fakeCliConnection.ApiEndpointReturns("", errors.New("API unavailable"))
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Unable to determine Api endpoint, use cf login first"}))
		})

		It("displays an error when the CF api is not a valid url", func() {
			fakeCliConnection.ApiEndpointReturns("blah", nil)
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Error parsing Api endpoint"}))
		})

		It("displays the host, port, api, and a connect confirmation", func() {
			defer gock.Off()

			gock.New(wicURL).
				Post(wicPath).
				JSON(`{"target":"` + "foo.com" + `:` + "80" + `}`).
				Reply(200).
				JSON(`{"lastChecked": 0, "entry": "foo.com", "canConnect": true, "httpStatus": 200, "validHostname": false, "validUrl": true}`)

			fakeCliConnection.ApiEndpointReturns(apiURL, nil)
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Host:", "foo.com", "-",
				"Port:", "80", "-",
				"WillItConnect:", wicURL + wicPath}))
			Expect(output).To(ContainSubstrings([]string{"I am able to connect"}))
		})

		It("displays the host, port, api, and an unable to connect failure", func() {
			defer gock.Off()

			gock.New(wicURL).
				Post(wicPath).
				JSON(`{"target":"` + "bar.com" + `:` + "80" + `}`).
				Reply(200).
				JSON(`{"lastChecked": 0,"entry": "bar.com","canConnect":false,"validHostname": false,"validUrl": true}`)

			fakeCliConnection.ApiEndpointReturns(apiURL, nil)
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
