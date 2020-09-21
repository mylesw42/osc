package cmd

import (
	"osc/pkg/osc"

	"github.com/spf13/cobra"
)

var (
	connectCmd = &cobra.Command{
		Use:   "connect {profile}",
		Short: "Connect sensuctl to a configured Sensu cluster.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osc.Connect(args)
		},
	}
)
