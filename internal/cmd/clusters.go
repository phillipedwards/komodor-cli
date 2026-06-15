package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newClustersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "Manage and inspect clusters",
	}

	cmd.AddCommand(
		newClustersListCmd(),
		newClustersUserClustersCmd(),
	)

	return cmd
}

func newClustersListCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetApiV2ClustersParams{
				ClusterName: strSlicePtr(splitAndFilter(name)),
			}

			resp, err := c.GetApiV2ClustersWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("list clusters: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&clusterListTable{})
			}
			return f.Print(&clusterListTable{clusters: resp.JSON200.Data.Clusters})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Filter by cluster name")

	return cmd
}

func newClustersUserClustersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "user-clusters",
		Short: "List clusters accessible to the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2UserClustersWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("get user clusters: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&userClusterTable{})
			}
			return f.Print(&userClusterTable{clusters: resp.JSON200.Data.Clusters})
		},
	}
}

// splitAndFilter splits a comma-separated string and drops empty entries.
// It is used to turn a single --name flag value into a slice for API params.
func splitAndFilter(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

// clusterListTable renders SingleCluster slices as a table.
type clusterListTable struct {
	clusters []client.SingleCluster
}

func (t *clusterListTable) Headers() []string {
	return []string{"NAME", "API_SERVER", "CLUSTER_ID"}
}

func (t *clusterListTable) Rows() [][]string {
	rows := make([][]string, len(t.clusters))
	for i, c := range t.clusters {
		clusterID := "-"
		if c.ClusterId != nil {
			clusterID = *c.ClusterId
		}
		rows[i] = []string{c.Name, c.ApiServerUrl, clusterID}
	}
	return rows
}

// userClusterTable renders UserCluster slices as a table.
type userClusterTable struct {
	clusters []client.UserCluster
}

func (t *userClusterTable) Headers() []string {
	return []string{"NAME", "CLUSTER_ID"}
}

func (t *userClusterTable) Rows() [][]string {
	rows := make([][]string, len(t.clusters))
	for i, c := range t.clusters {
		clusterID := "-"
		if c.ClusterId != nil {
			clusterID = *c.ClusterId
		}
		rows[i] = []string{c.Name, clusterID}
	}
	return rows
}

var _ output.TableData = (*clusterListTable)(nil)
var _ output.TableData = (*userClusterTable)(nil)
