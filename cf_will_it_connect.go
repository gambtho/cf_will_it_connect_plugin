package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/models"
)

const wicPath string = "/v2/willitconnect"
const wicRoute string = "willitconnect."

//WillItConnect ...
type WillItConnect struct{}

//GetMetadata ...
func (c *WillItConnect) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cf-willitconnect",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 2,
		},
		Commands: []plugin.Command{
			{
				Name:     "willitconnect",
				Alias:    "wic",
				HelpText: "Validates connectivity between CF and a target \n",
				UsageDetails: plugin.Usage{
					Usage: "willitconnect\n   Usage: cf willitconnect -host=<host> -port=<port>",
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
	wicFlags := flag.NewFlagSet("wicFlags", flag.ExitOnError)

	hostPtr := wicFlags.String("host", "", "host for connection")
	portPtr := wicFlags.Int("port", -1, "port for connection")
	proxyHostPtr := wicFlags.String("proxyHost", "", "host for proxy")
	proxyPortPtr := wicFlags.Int("proxyPort", -1, "port for proxy")

	wicFlags.Parse(args[1:]) // first arg is "willitconnect"
	fmt.Println("host:", *hostPtr)
	fmt.Println("port:", *portPtr)
	fmt.Println("proxyHost:", *proxyHostPtr)
	fmt.Println("proxyPort:", *proxyPortPtr)
	fmt.Println("tail:", wicFlags.Args())

	// port is not requied if host is a url

	if strings.HasPrefix(*hostPtr, "http://") {
		*portPtr = 80
	}

	if strings.HasPrefix(*hostPtr, "https://") {
		*portPtr = 443
	}

	if *portPtr == -1 || *hostPtr == "" {
		fmt.Println([]string{"Usage: cf willitconnect -host=<host> -port=<port>"})
	}

	currOrg, err := cliConnection.GetCurrentOrg()
	if (err != nil || currOrg == plugin_models.Organization{}) {
		fmt.Println("Unable to connect to CF, use cf login first")
		return
	}

	org, err := cliConnection.GetOrg(currOrg.OrganizationFields.Name)
	if err != nil {
		fmt.Println("Unable to find valid org, please view cf target")
		return
	}

	if len(org.Domains) < 1 {
		fmt.Println("Unable to find valid domain, please view cf domains")
	}

	baseURL := org.Domains[0].Name
	if baseURL == "" {
		fmt.Println("Unable to find valid domain, please view cf domains")
		return
	}

	wicURL := "https://" + wicRoute + baseURL + wicPath
	fmt.Println([]string{"Host: ", *hostPtr, " - Port: ", strconv.Itoa(*portPtr), " - WillItConnect: ", wicURL})
	if *proxyHostPtr != "" && *proxyPortPtr != -1 {
		fmt.Println([]string{"Proxy: " + *proxyHostPtr + ":" + strconv.Itoa(*proxyPortPtr)})
	}
	c.connect(*hostPtr, strconv.Itoa(*portPtr), wicURL, *proxyHostPtr, strconv.Itoa(*proxyPortPtr))

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

func (c *WillItConnect) connect(host string, port string, url string, proxyHost string, proxyPort string) {

	var payload []byte
	if proxyHost != "" && proxyPort != "-1" {
		fmt.Println(`{"target":"` + host + `:` + port + `", "http_proxy":"` + proxyHost + `:` + proxyPort + `"}`)
		payload = []byte(`{"target":"` + host + `:` + port + `", "http_proxy":"` + proxyHost + `:` + proxyPort + `"}`)
	} else {
		payload = []byte(`{"target":"` + host + `:` + port + `"}`)
	}
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
