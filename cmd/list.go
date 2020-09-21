package cmd

import (
	"osc/pkg/osc"

	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List config profiles, along with active sensuctl settings.",
		Run: func(cmd *cobra.Command, args []string) {
			osc.List()
		},
	}
)
