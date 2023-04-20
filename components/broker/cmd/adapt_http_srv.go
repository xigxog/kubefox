package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xigxog/kubefox/components/broker/engine"
)

var httpSrvCmd = &cobra.Command{
	Use:     "http-server",
	Short:   "",
	Long:    ``,
	PreRunE: validate,
	Run:     runHTTPSrv,
}

func init() {
	adaptCmd.AddCommand(httpSrvCmd)

	httpSrvCmd.Flags().StringVarP(&flags.HTTPSrvAddr, "http-addr", "r", "127.0.0.1:8080", "address and port of HTTP server")
	httpSrvCmd.Flags().StringVar(&flags.DevEnv, "dev-env", "", "environment that events without context will utilize (\"dev\" flag must be set)")
	httpSrvCmd.Flags().StringVar(&flags.DevApp, "dev-app", "", "app that events without context will utilize (\"dev\" flag must be set)")

	initCommonFlags(httpSrvCmd)
}

func runHTTPSrv(cmd *cobra.Command, args []string) {
	engine.New(flags).StartHTTPSrv()
}
