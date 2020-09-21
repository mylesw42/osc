package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	sensuctlClusterFile = "~/.config/sensu/sensuctl/cluster"
	sensuctlProfileFile = "~/.config/sensu/sensuctl/profile"
)

// Cluster sensuctl format
type Cluster struct {
	APIUrl                string `json:"api-url"`
	TrustedCAFile         string `json:"trusted-ca-file"`
	InsecureSkipTLSVerify bool   `json:"insecure-skip-tls-verify"`
	AccessToken           string `json:"access_token"`
	ExpiresAt             int    `json:"expires_at"`
	RefreshToken          string `json:"refresh_token"`
}

// Profile sensuctl format
type Profile struct {
	Format    string `json:"format"`
	Namespace string `json:"namespace"`
}

// Backend format
type Backend struct {
	Cluster
	Profile
}

// Create generates new sensuctl cluster/profile configs
func Create(c Cluster, server string) error {

	c.APIUrl = viper.GetString(server + ".api")
	newCluster, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	err = WriteSensuClusterConfig(newCluster)
	if err != nil {
		return err
	}

	tmpProfile := Profile{
		Format:    viper.GetString(server + ".format"),
		Namespace: viper.GetString(server + ".namespace"),
	}
	newProfile, err := json.MarshalIndent(tmpProfile, "", "  ")
	if err != nil {
		return err
	}
	err = WriteSensuProfileConfig(newProfile)
	if err != nil {
		return err
	}

	return nil
}

// ReadSensuConfig loads the current sensuctl config and returns a config.Backend{}
func ReadSensuConfig() Backend {

	home, _ := homedir.Dir()
	// read in config, ignore errors for now
	currentConfig, _ := os.Open(home + "/.config/sensu/sensuctl/cluster")
	defer currentConfig.Close()
	data, _ := ioutil.ReadAll(currentConfig)
	var showConfig Cluster

	json.Unmarshal(data, &showConfig)

	// read in profile
	currentProfile, _ := os.Open(home + "/.config/sensu/sensuctl/profile")
	defer currentProfile.Close()

	data, _ = ioutil.ReadAll(currentProfile)
	var showProfile Profile

	json.Unmarshal(data, &showProfile)

	return Backend{
		showConfig,
		showProfile,
	}

}

// WriteSensuClusterConfig creates a new cluster.json for sensuctl
func WriteSensuClusterConfig(newCluster []byte) error {
	fileloc, err := homedir.Expand(sensuctlClusterFile)
	if err != nil {
		return err
	}
	ioutil.WriteFile(fileloc, newCluster, 0644)
	return nil
}

// WriteSensuProfileConfig creates a new profile.json for sensuctl
func WriteSensuProfileConfig(newProfile []byte) error {
	fileloc, err := homedir.Expand(sensuctlProfileFile)
	if err != nil {
		return err
	}
	ioutil.WriteFile(fileloc, newProfile, 0644)
	return nil
}
