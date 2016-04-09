package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cloudfoundry/cli/plugin"
)

const wicPath string = "/v2/willitconnect"
const wicRoute string = "willitconnect"

//WillItConnect ...
type WillItConnect struct{}

//GetMetadata ...
func (c *WillItConnect) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "WillItConnect",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 1,
		},
		Commands: []plugin.Command{
			{
				Name:     "willitconnect",
				Alias:    "wic",
				HelpText: "Validates connectivity between CF and a target \n",
				UsageDetails: plugin.Usage{
					Usage: "willitconnect\n   cf willitconnect <host> <port>",
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(WillItConnect))
}

//Run ...
func (c *WillItConnect) Run(cliConnection plugin.CliConnection, args []string) {

	if len(args[1:]) < 2 {
		fmt.Println([]string{"Usage: cf willitconnect <host> <port>"})
		return

	}

	host := args[1]
	port := args[2]
	api, err := cliConnection.ApiEndpoint()

	if err != nil {
		fmt.Println([]string{"Unable to determine Api endpoint, use cf login first"})
		return
	}

	fmt.Println(api)
	apiURL, err := url.Parse(api)

	if err != nil || len(apiURL.Host) < 3 {
		fmt.Println([]string{"Error parsing Api endpoint"})
		return
	}

	baseURL := (apiURL.Host[3:])
	wicURL := apiURL.Scheme + "://" + wicRoute + baseURL + wicPath
	fmt.Println([]string{"Host: ", host, " - Port: ", port, " - WillItConnect: ", wicURL})
	c.connect(host, port, wicURL)

}

// WicResponse ...
type wicResponse struct {
	LastChecked   int    `json:"lastChecked"`
	Entry         string `json:"entry"`
	CanConnect    bool   `json:"canConnect"`
	HTTPStatus    int    `json:"httpStatus"`
	ValidHostname bool   `json:"validHostname"`
	ValidURL      bool   `json:"validUrl"`
}

func (c *WillItConnect) connect(host string, port string, url string) {
	payload := []byte(`{"target":"` + host + `:` + port + `}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println([]string{"Unable to access willitconnect: ", err.Error()})
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var body wicResponse
	decodeErr := decoder.Decode(&body)
	if decodeErr != nil {
		fmt.Println([]string{"Invalid response from willitconnect: ", decodeErr.Error()})
		return
	}
	if body.CanConnect {
		fmt.Println("I am able to connect")
	} else {
		fmt.Println("I am unable to connect")
	}
}
