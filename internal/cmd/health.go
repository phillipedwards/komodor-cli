package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Manage Komodor health risks",
	}

	risksCmd := &cobra.Command{
		Use:   "risks",
		Short: "Manage health risk violations",
	}

	risksCmd.AddCommand(
		newHealthRisksListCmd(),
		newHealthRisksGetCmd(),
		newHealthRisksUpdateCmd(),
	)

	cmd.AddCommand(risksCmd)
	return cmd
}

// violationsTable wraps a slice of BasicViolation for table output.
type violationsTable struct {
	violations []client.BasicViolation
}

func (t violationsTable) Headers() []string {
	return []string{"ID", "CHECK TYPE", "SEVERITY", "STATUS", "LINK"}
}

func (t violationsTable) Rows() [][]string {
	rows := make([][]string, len(t.violations))
	for i, v := range t.violations {
		checkType := ""
		if v.CheckType != nil {
			checkType = string(*v.CheckType)
		}
		severity := ""
		if v.Severity != nil {
			severity = string(*v.Severity)
		}
		rows[i] = []string{
			output.Str(v.Id),
			checkType,
			severity,
			string(v.Status),
			v.Link,
		}
	}
	return rows
}

func newHealthRisksListCmd() *cobra.Command {
	var (
		cluster  string
		severity string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List health risk violations",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &client.GetHealthRisksParams{
				PageSize:        100,
				Offset:          0,
				ImpactGroupType: []client.ImpactGroupType1{},
			}

			if cluster != "" {
				clusterNames := client.ClusterName{cluster}
				params.ClusterName = &clusterNames
			}

			if severity != "" {
				sev := client.Severity(severity)
				severities := client.SeverityRequestParam{sev}
				params.Severity = &severities
			}

			resp, err := clientFromCtx(cmd).GetHealthRisksWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("list health risks: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			return formatterFromCtx(cmd).Print(violationsTable{violations: resp.JSON200.Violations})
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Filter by cluster name")
	cmd.Flags().StringVar(&severity, "severity", "", "Filter by severity (e.g. low, medium, high, critical)")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

func newHealthRisksGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a health risk violation by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			resp, err := clientFromCtx(cmd).GetHealthRiskDataWithResponse(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("get health risk: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			return formatterFromCtx(cmd).Print(resp.JSON200.Violation)
		},
	}
}

func newHealthRisksUpdateCmd() *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update the status of a health risk violation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			statusVal := client.ViolationStatusRequestParam(status)

			body := client.UpdateHealthRiskStatusJSONRequestBody{
				Status: &statusVal,
			}

			resp, err := clientFromCtx(cmd).UpdateHealthRiskStatusWithResponse(cmd.Context(), id, body)
			if err != nil {
				return fmt.Errorf("update health risk status: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			return formatterFromCtx(cmd).Print(violationsTable{violations: []client.BasicViolation{resp.JSON200.Violation}})
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "New status (open, confirmed, resolved, dismissed, ignored, manually_resolved) (required)")
	_ = cmd.MarkFlagRequired("status")

	return cmd
}
