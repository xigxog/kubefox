package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xigxog/kubefox/components/broker/engine"
)

var compCmd = &cobra.Command{
	Use:     "component",
	Short:   "",
	Long:    ``,
	PreRunE: compValidate,
	Run:     runKubeFox,
}

func init() {
	rootCmd.AddCommand(compCmd)

	compCmd.Flags().StringVarP(&flags.GRPCSrvAddr, "grpc-addr", "g", "127.0.0.1:7070", "address and port of gRPC server")
	compCmd.Flags().StringVar(&flags.DevEnv, "dev-env", "", "environment that events without context will utilize (\"dev\" flag must be set)")
	compCmd.Flags().StringVar(&flags.DevApp, "dev-app", "", "app that events without context will utilize (\"dev\" flag must be set)")
	compCmd.Flags().StringVar(&flags.DevHTTPSrvAddr, "dev-http-ingress-addr", "", "address and port of HTTP ingress adapter for local development (\"dev\" flag must be set)")
	compCmd.MarkFlagRequired("system")
	compCmd.MarkFlagRequired("component")

	initCommonFlags(compCmd)
}

func compValidate(cmd *cobra.Command, args []string) error {
	if flags.IsDevMode && flags.CompGitHash == "" {
		flags.CompGitHash = "0"
	}

	if flags.CompGitHash == "" {
		return fmt.Errorf("required flag(s) \"component-hash\" not set")
	}

	return nil
}

func runKubeFox(cmd *cobra.Command, args []string) {
	engine.New(flags).StartComponent()
}
