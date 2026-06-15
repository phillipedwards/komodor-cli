package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAPIKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apikey",
		Short: "Manage API keys",
	}
	cmd.AddCommand(newAPIKeyValidateCmd())
	return cmd
}

func newAPIKeyValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the active API key against the Komodor API",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			resp, err := c.GetApiV2ApikeyValidateWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "API key is valid.")
			return nil
		},
	}
}
