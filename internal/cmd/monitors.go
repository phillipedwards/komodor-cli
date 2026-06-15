package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newMonitorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "monitors",
		Aliases: []string{"monitor"},
		Short:   "Manage realtime monitor configurations",
	}

	cmd.AddCommand(
		newMonitorsListCmd(),
		newMonitorsGetCmd(),
		newMonitorsCreateCmd(),
		newMonitorsUpdateCmd(),
		newMonitorsDeleteCmd(),
	)

	return cmd
}

func newMonitorsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all monitor configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2RealtimeMonitorsConfigWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("list monitors: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&monitorTable{})
			}
			return f.Print(&monitorTable{monitors: resp.JSON200.Data.Monitors})
		},
	}
}

func newMonitorsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a monitor configuration by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2RealtimeMonitorsConfigIdWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get monitor: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&monitorTable{})
			}
			return f.Print(&monitorTable{monitors: []client.MonitorConfiguration{*resp.JSON200}})
		},
	}
}

func newMonitorsCreateCmd() *cobra.Command {
	var (
		name       string
		monType    string
		active     bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a monitor configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.MonitorConfigurationParams{
				Name:    strPtr(name),
				Type:    monType,
				Sensors: []client.ConfigurationSensor{},
			}
			if cmd.Flags().Changed("active") {
				body.Active = boolPtr(active)
			}

			resp, err := c.PostApiV2RealtimeMonitorsConfigWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("create monitor: %w", err)
			}
			if resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON201 == nil {
				return f.Print(&monitorTable{})
			}
			return f.Print(&monitorTable{monitors: []client.MonitorConfiguration{*resp.JSON201}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Monitor name")
	cmd.Flags().StringVar(&monType, "type", "", "Monitor type (availability, cronJob, deploy, job, node, PVC, service, workflow)")
	cmd.Flags().BoolVar(&active, "active", true, "Whether the monitor is active")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func newMonitorsUpdateCmd() *cobra.Command {
	var (
		name    string
		active  bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a monitor configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			// Fetch current config to preserve existing fields.
			getResp, err := c.GetApiV2RealtimeMonitorsConfigIdWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get monitor: %w", err)
			}
			if getResp.StatusCode() != 200 {
				return apiError(getResp.StatusCode(), getResp.Body)
			}

			current := getResp.JSON200
			body := client.MonitorConfigurationParams{
				Type:         current.Type,
				Sensors:      current.Sensors,
				Name:         current.Name,
				Active:       current.Active,
				Sinks:        current.Sinks,
				SinksOptions: current.SinksOptions,
				Variables:    current.Variables,
			}

			if cmd.Flags().Changed("name") {
				body.Name = strPtr(name)
			}
			if cmd.Flags().Changed("active") {
				body.Active = boolPtr(active)
			}

			resp, err := c.PutApiV2RealtimeMonitorsConfigIdWithResponse(cmd.Context(), args[0], body)
			if err != nil {
				return fmt.Errorf("update monitor: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&monitorTable{})
			}
			return f.Print(&monitorTable{monitors: []client.MonitorConfiguration{*resp.JSON200}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New monitor name")
	cmd.Flags().BoolVar(&active, "active", true, "Set active state")

	return cmd
}

func newMonitorsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a monitor configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			resp, err := c.DeleteApiV2RealtimeMonitorsConfigIdWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("delete monitor: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Printf("monitor %s deleted\n", args[0])
			return nil
		},
	}
}

type monitorTable struct {
	monitors []client.MonitorConfiguration
}

func (t *monitorTable) Headers() []string {
	return []string{"ID", "NAME", "TYPE", "ACTIVE", "CREATED_AT"}
}

func (t *monitorTable) Rows() [][]string {
	rows := make([][]string, len(t.monitors))
	for i, m := range t.monitors {
		rows[i] = []string{
			m.Id,
			output.Str(m.Name),
			m.Type,
			output.Bool(m.Active),
			m.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return rows
}

var _ output.TableData = (*monitorTable)(nil)
