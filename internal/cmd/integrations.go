package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newIntegrationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "integrations",
		Aliases: []string{"integration"},
		Short:   "Manage Komodor integrations",
	}

	k8s := &cobra.Command{
		Use:   "k8s",
		Short: "Manage Kubernetes cluster integrations",
	}

	k8s.AddCommand(
		newIntegrationsK8sCreateCmd(),
		newIntegrationsK8sGetCmd(),
		newIntegrationsK8sDeleteCmd(),
	)

	cmd.AddCommand(k8s)
	return cmd
}

func newIntegrationsK8sCreateCmd() *cobra.Command {
	var clusterName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new Kubernetes cluster integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.CreateClusterDto{ClusterName: clusterName}
			resp, err := c.ClusterControllerPostWithResponse(cmd.Context(), &client.ClusterControllerPostParams{}, body)
			if err != nil {
				return fmt.Errorf("create k8s integration: %w", err)
			}
			if resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON201 == nil {
				return f.Print(&clusterIntegrationTable{})
			}
			return f.Print(&clusterIntegrationTable{entries: []clusterEntry{{name: clusterName, apiKey: resp.JSON201.ApiKey}}})
		},
	}

	cmd.Flags().StringVar(&clusterName, "cluster", "", "Cluster name to register")
	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

func newIntegrationsK8sGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <cluster-name>",
		Short: "Get a Kubernetes cluster integration by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.ClusterControllerGetByClusterNameWithResponse(cmd.Context(), args[0], &client.ClusterControllerGetByClusterNameParams{})
			if err != nil {
				return fmt.Errorf("get k8s integration: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&clusterIntegrationTable{})
			}
			return f.Print(&clusterIntegrationTable{entries: []clusterEntry{{name: args[0], apiKey: resp.JSON200.ApiKey}}})
		},
	}
}

func newIntegrationsK8sDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Kubernetes cluster integration by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			resp, err := c.ClusterControllerDeleteWithResponse(cmd.Context(), args[0], &client.ClusterControllerDeleteParams{})
			if err != nil {
				return fmt.Errorf("delete k8s integration: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Printf("cluster integration %s deleted\n", args[0])
			return nil
		},
	}
}

type clusterEntry struct {
	name   string
	apiKey string
}

type clusterIntegrationTable struct {
	entries []clusterEntry
}

func (t *clusterIntegrationTable) Headers() []string {
	return []string{"CLUSTER_NAME", "API_KEY"}
}

func (t *clusterIntegrationTable) Rows() [][]string {
	rows := make([][]string, len(t.entries))
	for i, e := range t.entries {
		rows[i] = []string{e.name, e.apiKey}
	}
	return rows
}

var _ output.TableData = (*clusterIntegrationTable)(nil)
