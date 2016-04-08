package main_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/gambtho/cf_will_it_connect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const wicURLSuffix string = "/v2/willitconnect"

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
			Expect(output).To(ContainSubstrings([]string{"Unable to determine CF ApiEndpoint"}))
		})

		It("displays the host, port, api, and a connect confirmation", func() {
			server := fakeServer(200, `{"lastChecked": 0,"entry": "foo.com","canConnect": true,"httpStatus": 200,"validHostname": false,"validUrl": true}`)
			defer server.Close()
			fakeCliConnection.ApiEndpointReturns(server.URL, nil)
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "foo.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Host:", "foo.com", "-",
				"Port:", "80", "-",
				"WillItConnect:", server.URL + wicURLSuffix}))
			Expect(output).To(ContainSubstrings([]string{"I am able to connect"}))
		})

		It("displays the host, port, api, and an unable to connect failure", func() {
			server := fakeServer(200, `{"lastChecked": 0,"entry": "bar.com","canConnect":
				false,"httpStatus": 200,"validHostname": false,"validUrl": true}`)
			defer server.Close()
			fakeCliConnection.ApiEndpointReturns(server.URL, nil)
			output := CaptureOutput(func() {
				willItConnectPlugin.Run(fakeCliConnection, []string{"willitconnect", "bar.com", "80"})
			})
			Expect(output).To(ContainSubstrings([]string{"Host:", "bar.com", "-",
				"Port:", "80", "-",
				"WillItConnect:", server.URL + wicURLSuffix}))
			Expect(output).To(ContainSubstrings([]string{"I am unable to connect"}))
		})
	})
})

func fakeServer(code int, body string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}))
	return server
}
