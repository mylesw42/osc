package osc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
	server        string
	api           string
	username      string
	password      string
	insecure      bool
	trustedcafile string
)

const (
	sep   = "\u2500"
	arrow = "\u2192"
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

// Connect establishes a new sensuctl config for the chosen backend profile
func Connect(args []string) {

	server = args[0]
	api = viper.GetString(server + ".api")
	username = viper.GetString(server + ".username")
	password = viper.GetString(server + ".password")
	insecure = viper.GetBool(server + ".insecure-skip-tls-verify")
	trustedcafile = viper.GetString(server + ".trusted-ca-file")

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
	if trustedcafile != "" {
		file, err := os.Open(trustedcafile)
		if err != nil {
			fmt.Println(err)
			return
		}
		certData, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Printf("Unable to load trusted-ca-file: %s", err)
			return
		}
		block, _ := pem.Decode(certData)
		if block == nil {
			fmt.Println("Unable to decode trusted-ca-file.  Is it in PEM format?")
			return
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			fmt.Printf("Invalid trusted-ca-file: %v", err)
			return
		}
		if !cert.IsCA {
			fmt.Println("Certificate in trusted-ca-file is not a CA")
			return
		}
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			fmt.Printf("Failed to get System Cert Pool: %v", err)
			rootCAs = x509.NewCertPool()
		}
		rootCAs.AddCert(cert)
		if client.Transport == nil {
			client.Transport = new(http.Transport)
		}
		if transport, ok := client.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = new(tls.Config)
			}
			transport.TLSClientConfig.RootCAs = rootCAs
		}

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
