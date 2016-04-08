package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cli/plugin"
)

//WillItConnect ...
type WillItConnect struct{}

//GetMetadata ...
func (c *WillItConnect) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "WillItConnect",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 1,
			Build: 0,
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

	wicURLSuffix := "/v2/willitconnect"
	//wicPrefix := "willitconnect"

	if len(args[1:]) < 2 {
		fmt.Println([]string{"Usage: cf willitconnect <host> <port>"})

	} else {
		host := args[1]
		port := args[2]
		url, err := cliConnection.ApiEndpoint()
		//wicHost, wicPort, _ := net.SplitHostPort(url)
		//if net.HasPort(url) -- need to check for port and behave accordingly

		if err != nil {
			fmt.Println([]string{"Unable to determine CF ApiEndpoint"})

		} else {
			wicURL := url + wicURLSuffix
			fmt.Println([]string{"Host: ", host, " - Port: ", port, " - WillItConnect: ", wicURL})
			c.connect(host, port, wicURL)
		}

	}
}

// WicResponse ...
type WicResponse struct {
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
	}
	//fmt.Println(resp)
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var t WicResponse
	decodeErr := decoder.Decode(&t)
	if decodeErr != nil {
		fmt.Println([]string{"Invalid response from willitconnect: ", decodeErr.Error()})
	}
	if t.CanConnect {
		fmt.Println("I am able to connect")
	} else {
		fmt.Println("I am unable to connect")
	}
}
