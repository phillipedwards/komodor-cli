package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Query audit logs and available filters",
	}

	cmd.AddCommand(
		newAuditLogsCmd(),
		newAuditFiltersCmd(),
	)

	return cmd
}

func newAuditLogsCmd() *cobra.Command {
	var (
		from     string
		to       string
		userID   string
		action   string
		pageSize int
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "List audit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetApiV2AuditLogParams{}

			if from != "" {
				t, err := time.Parse(time.RFC3339, from)
				if err != nil {
					return fmt.Errorf("invalid --from time (use RFC3339 e.g. 2006-01-02T15:04:05Z): %w", err)
				}
				params.StartTime = &t
			}
			if to != "" {
				t, err := time.Parse(time.RFC3339, to)
				if err != nil {
					return fmt.Errorf("invalid --to time (use RFC3339 e.g. 2006-01-02T15:04:05Z): %w", err)
				}
				params.EndTime = &t
			}
			if userID != "" {
				params.UserIds = &[]string{userID}
			}
			if action != "" {
				params.Actions = &[]string{action}
			}
			if cmd.Flags().Changed("page-size") {
				params.PageSize = &pageSize
			}

			// When the user requests CSV output, pass the Accept header so the
			// server returns raw CSV. Write the body directly to stdout instead
			// of going through the formatter.
			format, _ := cmd.Root().PersistentFlags().GetString("output")
			if format == "csv" {
				accept := "text/csv"
				params.Accept = &accept

				resp, err := c.GetApiV2AuditLogWithResponse(cmd.Context(), params)
				if err != nil {
					return fmt.Errorf("list audit logs: %w", err)
				}
				if resp.StatusCode() != 200 {
					return apiError(resp.StatusCode(), resp.Body)
				}
				_, err = os.Stdout.Write(resp.Body)
				return err
			}

			resp, err := c.GetApiV2AuditLogWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("list audit logs: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&auditLogTable{})
			}
			return f.Print(&auditLogTable{logs: resp.JSON200.Logs})
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Start time in RFC3339 format")
	cmd.Flags().StringVar(&to, "to", "", "End time in RFC3339 format")
	cmd.Flags().StringVar(&userID, "user-id", "", "Filter by user ID")
	cmd.Flags().StringVar(&action, "action", "", "Filter by action name")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Number of results per page")

	return cmd
}

func newAuditFiltersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "filters",
		Short: "List available audit log filter values",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2AuditLogFiltersWithResponse(cmd.Context(), &client.GetApiV2AuditLogFiltersParams{})
			if err != nil {
				return fmt.Errorf("get audit filters: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(map[string]interface{}{})
			}
			return f.Print(resp.JSON200)
		},
	}
}

type auditLogTable struct {
	logs []client.AuditLog
}

func (t *auditLogTable) Headers() []string {
	return []string{"ACTION", "USER", "ENTITY", "STATUS", "TIME"}
}

func (t *auditLogTable) Rows() [][]string {
	rows := make([][]string, len(t.logs))
	for i, l := range t.logs {
		entity := l.EntityType
		if l.EntityName != nil && *l.EntityName != "" {
			entity = fmt.Sprintf("%s/%s", l.EntityType, *l.EntityName)
		}

		ts := ""
		if l.EventTime != nil {
			ts = l.EventTime.Format("2006-01-02 15:04:05")
		} else if l.CreatedAt != nil {
			ts = l.CreatedAt.Format("2006-01-02 15:04:05")
		}

		rows[i] = []string{
			l.Action,
			l.UserEmail,
			entity,
			string(l.Status),
			ts,
		}
	}
	return rows
}

var _ output.TableData = (*auditLogTable)(nil)
