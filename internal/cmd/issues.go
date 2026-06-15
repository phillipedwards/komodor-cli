package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
)

func newIssuesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issues",
		Aliases: []string{"issue"},
		Short:   "Search Komodor issues",
	}

	cmd.AddCommand(
		newIssuesSearchCmd(),
		newIssuesSearchClusterCmd(),
	)

	return cmd
}

// issuesTable wraps a slice of SingleIssue for table output.
type issuesTable struct {
	issues []client.SingleIssue
}

func (t issuesTable) Headers() []string {
	return []string{"SUMMARY", "TYPE", "STATUS", "START TIME", "END TIME"}
}

func (t issuesTable) Rows() [][]string {
	rows := make([][]string, len(t.issues))
	for i, issue := range t.issues {
		endTime := ""
		if issue.EndTime != nil {
			endTime = time.Unix(*issue.EndTime, 0).UTC().Format(time.RFC3339)
		}
		rows[i] = []string{
			issue.Summary,
			string(issue.Type),
			string(issue.Status),
			time.Unix(issue.StartTime, 0).UTC().Format(time.RFC3339),
			endTime,
		}
	}
	return rows
}

func newIssuesSearchCmd() *cobra.Command {
	var (
		serviceID string
		from      string
		to        string
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search issues for a service",
		RunE: func(cmd *cobra.Command, args []string) error {
			fromTime, err := time.Parse(time.RFC3339, from)
			if err != nil {
				return fmt.Errorf("invalid --from value: %w", err)
			}
			toTime, err := time.Parse(time.RFC3339, to)
			if err != nil {
				return fmt.Errorf("invalid --to value: %w", err)
			}

			fromEpoch := client.EpochTime(fromTime.Unix())
			toEpoch := client.EpochTime(toTime.Unix())

			body := client.SearchServicesIssuesBody{
				Scope: client.ServiceScope{
					Service: serviceID,
				},
				Props: client.IssuesProps{
					FromEpoch: &fromEpoch,
					ToEpoch:   &toEpoch,
					Statuses:  []client.IssueStatus{},
					Type:      "all",
				},
			}

			resp, err := clientFromCtx(cmd).PostApiV2ServicesIssuesSearchWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("search service issues: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			return formatterFromCtx(cmd).Print(issuesTable{issues: resp.JSON200.Data.Issues})
		},
	}

	cmd.Flags().StringVar(&serviceID, "service-id", "", "Service ID to search issues for (required)")
	cmd.Flags().StringVar(&from, "from", "", "Start time in RFC3339 format (required)")
	cmd.Flags().StringVar(&to, "to", "", "End time in RFC3339 format (required)")
	_ = cmd.MarkFlagRequired("service-id")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.RegisterFlagCompletionFunc("service-id", completeServiceNames)

	return cmd
}

func newIssuesSearchClusterCmd() *cobra.Command {
	var (
		cluster string
		from    string
		to      string
	)

	cmd := &cobra.Command{
		Use:   "search-cluster",
		Short: "Search issues for a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			fromTime, err := time.Parse(time.RFC3339, from)
			if err != nil {
				return fmt.Errorf("invalid --from value: %w", err)
			}
			toTime, err := time.Parse(time.RFC3339, to)
			if err != nil {
				return fmt.Errorf("invalid --to value: %w", err)
			}

			fromEpoch := client.EpochTime(fromTime.Unix())
			toEpoch := client.EpochTime(toTime.Unix())

			body := client.SearchClustersIssuesBody{
				Scope: client.ClusterScope{
					Cluster: cluster,
				},
				Props: client.IssuesProps{
					FromEpoch: &fromEpoch,
					ToEpoch:   &toEpoch,
					Statuses:  []client.IssueStatus{},
					Type:      "all",
				},
			}

			resp, err := clientFromCtx(cmd).PostApiV2ClustersIssuesSearchWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("search cluster issues: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			return formatterFromCtx(cmd).Print(issuesTable{issues: resp.JSON200.Data.Issues})
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name to search issues for (required)")
	cmd.Flags().StringVar(&from, "from", "", "Start time in RFC3339 format (required)")
	cmd.Flags().StringVar(&to, "to", "", "End time in RFC3339 format (required)")
	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}
