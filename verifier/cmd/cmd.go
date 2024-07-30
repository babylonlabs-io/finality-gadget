package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

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
