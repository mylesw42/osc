package osc

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"osc/pkg/config"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/viper"
)

var (
	server   string
	api      string
	username string
	password string
	insecure bool
)

const (
	sep   = "\u2500"
	arrow = "\u2BC8"
)

// List shows the current sensuctl config and all OSC profiles
func List() {
	current := config.ReadSensuConfig()

	if current.APIUrl == "" {
		fmt.Println("Unable to read sensuctl (~/.config/sensu/sensuctl/) cluster & profile.\n")
	} else {
		fmt.Printf("Active Config\n%s\n%s API: %s\n%s Namespace: %s\n%s Format: %s\n\n",
			frmt("active config"),
			arrow, current.APIUrl,
			arrow, current.Namespace,
			arrow, current.Format)
	}

	w := tabwriter.NewWriter(os.Stdout, 16, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Profile\tEnvironment\tUsername\tNamespace\tFormat\tAPI\t")
	fmt.Fprintln(w,
		frmt("profile")+"\t"+
			frmt("environment")+"\t"+
			frmt("username")+"\t"+
			frmt("namespace")+"\t"+
			frmt("format")+"\t"+
			frmt("api")+"\t")

	for profile := range viper.AllSettings() {
		cfg := viper.GetStringMapString(profile)
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t\n",
			profile,
			cfg["env"],
			cfg["username"],
			cfg["namespace"],
			cfg["format"],
			cfg["api"])
		if err != nil {
			fmt.Println(err)
		}
		w.Flush()
	}
}

// Connect establishes a new sensuctl config for the chosen backend
func Connect(args []string) {

	server = args[0]
	api = viper.GetString(server + ".api")
	username = viper.GetString(server + ".username")
	password = viper.GetString(server + ".password")
	insecure = viper.GetBool(server + ".insecure")

	if viper.GetString(server+".api") == "" {
		fmt.Printf("Config profile (%s) does not exist.\n", server)
		os.Exit(1)
	}

	// create the http client
	var client = http.Client{
		Timeout: 10 * time.Second,
	}
	if insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		client.Transport = tr
	}

	isValid := backendAuth(client)
	if isValid {
		newConfig, err := backendToken(client)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = config.Create(newConfig, server)
		if err != nil {
			fmt.Printf("Error connecting: %s\n", err)
			return
		}
		fmt.Printf("Connected to Sensu backend: %s (%s)\n", server, api)
	} else {
		fmt.Println("Auth failed! Check profile credentials.")
		os.Exit(1)
	}

}

// backendAuth tests the credentials to ensure a token can be created for the config
func backendAuth(c http.Client) bool {

	req, err := http.NewRequest("GET", api+"/auth/test", nil)
	req.SetBasicAuth(username, password)
	if err != nil {
		fmt.Println(err)
		return false
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if resp.StatusCode == http.StatusOK {
		return true
	}
	defer resp.Body.Close()

	return false
}

// backendToken executes if the auth test succeeded. It generates an access_token and refresh_token
// from the sensu backend API
func backendToken(c http.Client) (config.Cluster, error) {
	req, err := http.NewRequest("GET", api+"/auth", nil)
	req.SetBasicAuth(username, password)
	if err != nil {
		return config.Cluster{}, err
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	respData, err := ioutil.ReadAll(resp.Body)

	var respConfig config.Cluster
	if resp.StatusCode == http.StatusOK {
		json.Unmarshal(respData, &respConfig)
	}
	defer resp.Body.Close()

	return respConfig, nil
}

// frmt returns a unicode separator the length of the provided string
// to make the CLI output formatted a bit better.
func frmt(s string) string {
	return strings.Repeat(sep, len(s))
}
