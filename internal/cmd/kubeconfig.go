package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
)

func newKubeconfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Manage kubeconfig files",
	}
	cmd.AddCommand(newKubeconfigGetCmd())
	return cmd
}

func newKubeconfigGetCmd() *cobra.Command {
	var clusters []string
	var connection string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Download a kubeconfig from Komodor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			params := &client.GetApiV2RbacKubeconfigParams{
				ClusterName: strSlicePtr(clusters),
			}

			if connection != "" {
				conn := client.GetApiV2RbacKubeconfigParamsKubeconfigConnection(connection)
				params.KubeconfigConnection = &conn
			}

			resp, err := c.GetApiV2RbacKubeconfigWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), string(resp.Body))
			return err
		},
	}

	cmd.Flags().StringSliceVar(&clusters, "cluster", nil, "Filter by cluster name (may be repeated)")
	cmd.Flags().StringVar(&connection, "connection", "", "Connection type: direct, proxy, or both")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}
