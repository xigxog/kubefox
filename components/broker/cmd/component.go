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

	compCmd.Flags().StringVarP(&flags.System, "system", "s", "", "system the component belongs to (required)")
	compCmd.Flags().StringVarP(&flags.CompName, "component", "c", "", "component's short but unique and descriptive name (required)")
	compCmd.Flags().StringVarP(&flags.CompGitHash, "component-hash", "a", "", "hash of the component (required if \"dev\" flag not set)")
	compCmd.Flags().StringVarP(&flags.GRPCSrvAddr, "grpc-addr", "g", "127.0.0.1:6060", "address and port of component gRPC server")
	compCmd.Flags().StringVarP(&flags.GRPCCertsDir, "grpc-certs-dir", "d", "", "path of dir containing certificate and key of component gRPC server")
	compCmd.Flags().StringVar(&flags.DevEnv, "dev-env", "", "environment that events without context will utilize (\"dev\" flag must be set)")
	compCmd.Flags().StringVar(&flags.DevApp, "dev-app", "", "app that events without context will utilize (\"dev\" flag must be set)")
	compCmd.Flags().StringVar(&flags.DevHTTPSrvAddr, "dev-http-ingress-addr", "", "address and port of HTTP ingress adapter for local development (\"dev\" flag must be set)")
	initCommonFlags(compCmd)

	compCmd.MarkFlagRequired("system")
	compCmd.MarkFlagRequired("component")
}

func compValidate(cmd *cobra.Command, args []string) error {
	if flags.IsDevMode {
		if flags.CompGitHash == "" {
			flags.CompGitHash = "0000000"
		}
	} else {
		if flags.DevHTTPSrvAddr != "" {
			return fmt.Errorf("flag \"dev-http-ingress-addr\" set without flag \"dev\" set")
		}
	}

	if flags.CompGitHash == "" {
		return fmt.Errorf("required flag(s) \"component-hash\" not set")
	}

	return nil
}

func runKubeFox(cmd *cobra.Command, args []string) {
	engine.New(flags).StartComponent()
}
