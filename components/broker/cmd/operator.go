package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xigxog/kubefox/components/broker/engine"
	"github.com/xigxog/kubefox/libs/core/platform"
)

var oprCmd = &cobra.Command{
	Use:     "operator",
	Short:   "",
	Long:    ``,
	PreRunE: validate,
	Run:     runOpr,
}

func init() {
	rootCmd.AddCommand(oprCmd)

	oprCmd.Flags().StringVarP(&flags.GRPCSrvAddr, "grpc-addr", "g", "127.0.0.1:6060", "address and port of broker gRPC server")
	oprCmd.Flags().StringVarP(&flags.HTTPSrvAddr, "http-addr", "r", "0.0.0.0:8080", "address and port of broker HTTP server")
	oprCmd.Flags().StringVarP(&flags.GRPCCertsDir, "grpc-certs-dir", "d", platform.BrokerCertsDir, "path of dir containing certificate and key of component gRPC server")
	oprCmd.Flags().StringVarP(&flags.OperatorCertsDir, "operator-certs-dir", "o", platform.OperatorCertsDir, "path of dir containing certificate and key of operator gRPC and HTTP server")

	initCommonFlags(oprCmd)
	oprCmd.MarkFlagRequired("namespace")
}

func runOpr(cmd *cobra.Command, args []string) {
	flags.System = platform.System
	flags.CompName = platform.OperatorComp.GetName()
	flags.CompGitHash = platform.OperatorComp.GetGitHash()

	engine.New(flags).StartOperator()
}
