package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "osc",
	Short: "Manage configured sensuctl clusters",
	Long: `OSC (Operate Sensu Cluster) is a utility to enable Sensu administrators the
ability to quickly switch between clusters using profile configurations.`,
}

func init() {
	viper.SetConfigName("osc.config")
	viper.SetConfigType("yaml")

	home, err := homedir.Dir()
	if err != nil {
		fmt.Printf("Error finding home directory: %s\n", err)
		os.Exit(1)
	}
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.AddConfigPath(home + "/.config")

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error loading config file: %s\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(connectCmd, listCmd)
}

// Execute now!
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
