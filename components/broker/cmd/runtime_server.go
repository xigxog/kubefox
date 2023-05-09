package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xigxog/kubefox/components/broker/engine"
	"github.com/xigxog/kubefox/libs/core/platform"
)

var fabSrvCmd = &cobra.Command{
	Use:     "runtime-server",
	Short:   "",
	Long:    ``,
	PreRunE: validate,
	Run:     runFabSrv,
}

func init() {
	rootCmd.AddCommand(fabSrvCmd)

	fabSrvCmd.Flags().StringVarP(&flags.GRPCSrvAddr, "grpc-addr", "g", "127.0.0.1:7070", "address and port of gRPC server")

	initCommonFlags(fabSrvCmd)
	fabSrvCmd.MarkFlagRequired("namespace")
}

func runFabSrv(cmd *cobra.Command, args []string) {
	flags.System = platform.System
	flags.CompName = platform.RuntimeSrvComp.GetName()
	flags.CompGitHash = platform.RuntimeSrvComp.GetGitHash()

	engine.New(flags).StartRuntimeSrv()
}
