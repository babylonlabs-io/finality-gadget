package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "vrf",
		Short:             "vrf - Babylon OP Finality Gadget Verifier",
		Long:              `vrf is a daemon to track consecutive quorum and query the Babylon BTC block finalization status of OP stack chains.`,
		SilenceErrors:     false,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.PersistentFlags())
		},
	}

	return rootCmd
}

func main() {
	cmd := NewRootCmd()

	cmd.PersistentFlags().String("cfg", "config.toml", "config file")
	viper.BindPFlag("cfg", cmd.PersistentFlags().Lookup("cfg"))

	cmd.AddCommand(CommandStart())

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your vrf CLI '%s'", err)
		os.Exit(1)
	}
}

// Runs cmd with client context and returns an error.
func runEWithClientCtx(
	fRunWithCtx func(ctx client.Context, cmd *cobra.Command, args []string) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		return fRunWithCtx(clientCtx, cmd, args)
	}
}
