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

	adaptCmd = &cobra.Command{
		Use:   "adapter",
		Short: "",
		Long:  ``,
	}

	bootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "",
		Long:  ``,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(adaptCmd)
	rootCmd.AddCommand(bootstrapCmd)
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

// initCommonFlags is called by each subcommand. This is used instead of
// persistent flags on root so that completion, docs, and help can be called
// with required flags being set.
func initCommonFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVarP(&flags.Platform, "platform", "p", "", "platform instance the component runs on (required)")
	f.StringVarP(&flags.System, "system", "s", "", "system the component belongs to (required)")
	f.StringVarP(&flags.CompName, "component", "c", "", "component's short but unique and descriptive name (required)")
	f.StringVarP(&flags.CompGitHash, "component-hash", "a", "", "hash of the component (required if \"dev\" flag not set)")
	f.StringVarP(&flags.RuntimeSrvAddr, "runtime-srv-addr", "k", "127.0.0.1:6060", "address and port of gRPC server for KubeFox Runtime Server")
	f.StringVarP(&flags.NatsAddr, "nats-addr", "n", "127.0.0.1:4222", "address and port of NATS server, JetStream must be enabled on NATS server")
	f.StringVarP(&flags.TelemetrySrvAddr, "telemetry-srv-addr", "m", "127.0.0.1:8888", "address and port of the brokers's HTTP telemetry server (disabled if \"dev\" flag is set)")
	f.StringVarP(&flags.TraceAgentAddr, "trace-agent-addr", "e", "127.0.0.1:6831", "address and port of the trace agent")
	f.StringVarP(&flags.Namespace, "namespace", "l", "", "Kubernetes namespace containing the KubeFox Platform")
	f.Int64VarP(&flags.EventTimeout, "timeout", "t", 30, "timeout in seconds for processing an event")
	f.BoolVar(&flags.IsDevMode, "dev", false, "enable dev mode")

	cmd.MarkFlagRequired("platform")
	cmd.MarkFlagRequired("system")
	cmd.MarkFlagRequired("component")
}

func validate(cmd *cobra.Command, args []string) error {
	fmt.Printf("VALIDATE CALLED\n\n")
	if flags.IsDevMode {
		if flags.CompGitHash == "" {
			flags.CompGitHash = "0000000"
		}

	} else {
		if flags.CompGitHash == "" {
			return fmt.Errorf("required flag(s) \"component-hash\" not set")
		}
		if flags.DevHTTPSrvAddr != "" {
			return fmt.Errorf("flag \"dev-http-ingress-addr\" set without flag \"dev\" set")
		}
		if flags.DevEnv != "" {
			return fmt.Errorf("flag \"dev-env\" set without flag \"dev\" set")
		}
		if flags.DevApp != "" {
			return fmt.Errorf("flag \"dev-app\" set without flag \"dev\" set")
		}
	}

	return nil
}
