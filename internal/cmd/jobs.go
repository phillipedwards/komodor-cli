package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage and inspect jobs",
	}

	cmd.AddCommand(
		newJobsSearchCmd(),
	)

	return cmd
}

func newJobsSearchCmd() *cobra.Command {
	var (
		cluster   string
		namespace string
		name      string
		page      int
		pageSize  int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for jobs and cronjobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.SearchJobsBody{}

			if cluster != "" || namespace != "" {
				scope := &client.OptionalClusterScope{}
				if cluster != "" {
					scope.Cluster = strPtr(cluster)
				}
				if namespace != "" {
					scope.Namespaces = &[]string{namespace}
				}
				body.Scope = scope
			}

			if cmd.Flags().Changed("page") || cmd.Flags().Changed("page-size") {
				pagination := &client.PaginationParams{}
				if cmd.Flags().Changed("page") {
					pagination.Page = &page
				}
				if cmd.Flags().Changed("page-size") {
					pagination.PageSize = &pageSize
				}
				body.Pagination = pagination
			}

			// The jobs search API does not expose a name filter in SearchJobsBody.
			_ = name

			resp, err := c.PostApiV2JobsSearchWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("search jobs: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&jobSearchTable{})
			}
			return f.Print(&jobSearchTable{jobs: resp.JSON200.Data.Jobs})
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Filter by cluster name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Filter by namespace")
	cmd.Flags().StringVar(&name, "name", "", "Filter by job name (not supported server-side; provided for future use)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Number of results per page")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

// jobSearchTable renders SingleJob slices as a table.
type jobSearchTable struct {
	jobs []client.SingleJob
}

func (t *jobSearchTable) Headers() []string {
	return []string{"NAME", "NAMESPACE", "CLUSTER", "KIND"}
}

func (t *jobSearchTable) Rows() [][]string {
	rows := make([][]string, len(t.jobs))
	for i, j := range t.jobs {
		rows[i] = []string{j.Name, j.Namespace, j.Cluster, j.Kind}
	}
	return rows
}

var _ output.TableData = (*jobSearchTable)(nil)
