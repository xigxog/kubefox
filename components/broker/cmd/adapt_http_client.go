package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xigxog/kubefox/components/broker/engine"
)

var httpClientCmd = &cobra.Command{
	Use:     "http-client",
	Short:   "",
	Long:    ``,
	PreRunE: validate,
	Run:     runHTTPClient,
}

func init() {
	adaptCmd.AddCommand(httpClientCmd)
	initCommonFlags(httpClientCmd)
}

func runHTTPClient(cmd *cobra.Command, args []string) {
	engine.New(flags).StartHTTPClient()
}
