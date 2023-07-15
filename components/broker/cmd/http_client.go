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
	rootCmd.AddCommand(httpClientCmd)

	httpClientCmd.Flags().StringVarP(&flags.System, "system", "s", "", "system the component belongs to (required)")
	httpClientCmd.Flags().StringVarP(&flags.CompName, "component", "c", "", "component's short but unique and descriptive name (required)")
	initCommonFlags(httpClientCmd)

	httpClientCmd.MarkFlagRequired("system")
	httpClientCmd.MarkFlagRequired("component")
}

func runHTTPClient(cmd *cobra.Command, args []string) {
	engine.New(flags).StartHTTPClient()
}
