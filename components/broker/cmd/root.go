package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xigxog/kubefox/components/broker/config"
	"github.com/xigxog/kubefox/libs/core/logger"
	"github.com/xigxog/kubefox/libs/core/platform"
)

var (
	log *logger.Log = logger.CLILogger().Named("broker")

	flags config.Flags

	rootCmd = &cobra.Command{
		Use:              "broker",
		Short:            "",
		Long:             ``,
		PersistentPreRun: initViper,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initViper(cmd *cobra.Command, args []string) {
	viper.SetEnvPrefix("kubefox")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Value.String() == f.DefValue && viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}

// initCommonFlags should called by each subcommand. This is used instead of
// persistent flags on root so that completion, docs, and help can be called
// without required flags being set.
func initCommonFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVarP(&flags.Platform, "platform", "p", "", "platform instance component runs on (required)")
	f.StringVarP(&flags.Namespace, "platform-namespace", "l", "", "namespace containing platform instance (required)")
	f.StringVarP(&flags.CACertPath, "ca-cert-path", "f", platform.CACertPath, "path of file containing KubeFox root CA certificate")
	f.StringVarP(&flags.OperatorAddr, "operator-addr", "k", "127.0.0.1:7070", "address and port of operator gRPC server")
	f.StringVarP(&flags.NATSAddr, "nats-addr", "n", "127.0.0.1:4222", "address and port of NATS JetStream server")
	f.StringVarP(&flags.HealthSrvAddr, "health-addr", "m", "0.0.0.0:1111", "address and port of brokers HTTP health server, set to \"false\" to disable")
	f.StringVarP(&flags.TelemetryAgentAddr, "telemetry-agent-addr", "e", "127.0.0.1:4318", "address and port of telemetry agent, set to \"false\" to disable")
	f.Uint8VarP(&flags.EventTimeout, "timeout", "t", 30, "timeout in seconds for processing an event")
	f.Uint8VarP(&flags.MetricsInterval, "metrics-interval", "i", 60, "interval in seconds to report metrics")
	f.BoolVar(&flags.IsDevMode, "dev", false, "enable dev mode")

	cmd.MarkFlagRequired("platform")
	cmd.MarkFlagRequired("namespace")
}

func validate(cmd *cobra.Command, args []string) error {
	if !flags.IsDevMode {
		if flags.DevEnv != "" {
			return fmt.Errorf("flag \"dev-env\" set without flag \"dev\" set")
		}
		if flags.DevApp != "" {
			return fmt.Errorf("flag \"dev-app\" set without flag \"dev\" set")
		}
	}

	return nil
}
