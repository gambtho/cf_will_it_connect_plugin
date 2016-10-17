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
const wicRoute string = "willitconnect"
const usage string = "cf willitconnect -host=<host> -port=<port> [proxyHost=<proxyHost>] proxyPort=<proxyPort>] [-route=<route>] "

//WillItConnect ...
type WillItConnect struct{}

//GetMetadata ...
func (c *WillItConnect) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cf-willitconnect",
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
					Usage: "willitconnect\n   Usage: cf willitconnect -host=<host> -port=<port>\n" +
						"cf willitconnect <url>\n" +
						"cf willitconnect -host=<host -port=<port> -proxyHost=<proxyHost -proxyPort=<proxyPort -route=<route>\n",
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

	baseURL, cfErr := c.getBaseURL(cliConnection)

	if cfErr != nil {
		fmt.Println(cfErr)
		return
	}

	request, argsErr := c.parseArgs(args, baseURL)

	if argsErr != nil {
		fmt.Println(argsErr)
		return
	}

	fmt.Println([]string{"Host: ", request.host, " - Port: ", request.port, " - WillItConnect: ", request.url})
	if request.hasProxy {
		fmt.Println([]string{"Proxy: " + request.proxyHost + ":" + request.proxyPort})
	}

	response, conErr := c.connect(request)

	if conErr != nil {
		fmt.Println(conErr)
		return
	}
	fmt.Println(response)
}

type wicRequest struct {
	host      string
	port      string
	url       string
	hasProxy  bool
	proxyHost string
	proxyPort string
}

type wicResponse struct {
	LastChecked   int    `json:"lastChecked"`
	Entry         string `json:"entry"`
	CanConnect    bool   `json:"canConnect"`
	HTTPStatus    int    `json:"httpStatus"`
	ValidHostname bool   `json:"validHostname"`
	ValidURL      bool   `json:"validUrl"`
}

func (c *WillItConnect) getBaseURL(cliConnection plugin.CliConnection) (*string, []string) {

	currOrg, err := cliConnection.GetCurrentOrg()

	if (err != nil || currOrg == plugin_models.Organization{}) {
		return nil, []string{"Unable to connect to CF, use cf login first"}
	}
	org, err := cliConnection.GetOrg(currOrg.OrganizationFields.Name)
	if err != nil {
		return nil, []string{"Unable to find valid org, please view cf target"}
	}

	if len(org.Domains) < 1 {
		return nil, []string{"Unable to find valid domain, please view cf domains"}
	}

	baseURL := org.Domains[0].Name
	if baseURL == "" {
		return nil, []string{"Unable to find valid domain, please view cf domains"}
	}
	return &baseURL, nil
}

func (c *WillItConnect) parseArgs(args []string, baseURL *string) (*wicRequest, []string) {
	wicFlags := flag.NewFlagSet("wicFlags", flag.ExitOnError)

	hostPtr := wicFlags.String("host", "", "host for connection")
	portPtr := wicFlags.Int("port", -1, "port for connection")
	proxyHostPtr := wicFlags.String("proxyHost", "", "host for proxy")
	proxyPortPtr := wicFlags.Int("proxyPort", -1, "port for proxy")
	routePtr := wicFlags.String("route", "", "route for willitconnect")

	wicFlags.Parse(args[1:])

	if strings.HasPrefix(*hostPtr, "http://") {
		*portPtr = 80
	}

	if strings.HasPrefix(*hostPtr, "https://") {
		*portPtr = 443
	}

	if *portPtr == -1 || *hostPtr == "" {
		if len(wicFlags.Args()) == 1 {
			if strings.HasPrefix(wicFlags.Args()[0], "http://") {
				*hostPtr = wicFlags.Args()[0]
				*portPtr = 80
			} else if strings.HasPrefix(wicFlags.Args()[0], "https://") {
				*hostPtr = wicFlags.Args()[0]
				*portPtr = 443
			} else {
				return nil, []string{"Usage: cf willitconnect -host=<host> -port=<port>"}
			}
		} else {
			return nil, []string{"Usage: cf willitconnect -host=<host> -port=<port>"}
		}
	}

	wicURL := "https://" + wicRoute + "." + *baseURL
	if *routePtr != "" {
		if 2 > strings.Count(*routePtr, ".") {
			return nil, []string{"-route must be a fqdn"}
		}

		if strings.HasPrefix(*routePtr, "http") {
			wicURL = *routePtr
		} else {
			wicURL = "https://" + *routePtr
		}
	}
	wicURL += wicPath

	hasProxy := false
	if *proxyHostPtr != "" && *proxyPortPtr != -1 {
		hasProxy = true
	}
	request := wicRequest{*hostPtr, strconv.Itoa(*portPtr), wicURL, hasProxy, *proxyHostPtr, strconv.Itoa(*proxyPortPtr)}
	return &request, nil
}

func (c *WillItConnect) connect(request *wicRequest) ([]string, []string) {
	var payload []byte
	if request.hasProxy {
		payload = []byte(`{"target":"` + request.host + `:` + request.port + `", "http_proxy":"` + request.proxyHost + `:` + request.proxyPort + `"}`)
	} else {
		payload = []byte(`{"target":"` + request.host + `:` + request.port + `"}`)
	}
	req, err := http.NewRequest("POST", request.url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, []string{"Unable to access willitconnect: ", err.Error()}
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var body wicResponse
	decodeErr := decoder.Decode(&body)
	if decodeErr != nil {
		return nil, []string{"Invalid response from willitconnect: ", decodeErr.Error()}
	}
	var response []string
	if body.CanConnect {
		response = []string{"I am able to connect"}
	} else {
		response = []string{"I am unable to connect"}
	}
	return response, nil
}
