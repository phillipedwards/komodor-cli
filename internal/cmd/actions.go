package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newActionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "actions",
		Aliases: []string{"action"},
		Short:   "Manage RBAC custom actions",
	}

	cmd.AddCommand(
		newActionsListCmd(),
		newActionsGetCmd(),
		newActionsCreateCmd(),
		newActionsUpdateCmd(),
		newActionsDeleteCmd(),
	)

	return cmd
}

func newActionsListCmd() *cobra.Command {
	var actions []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom actions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetApiV2RbacActionsParams{
				Actions: strSlicePtr(actions),
			}

			resp, err := c.GetApiV2RbacActionsWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("list actions: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&actionTable{})
			}
			return f.Print(&actionTable{actions: *resp.JSON200})
		},
	}

	cmd.Flags().StringSliceVar(&actions, "action", nil, "Filter by action name (may be repeated)")

	return cmd
}

func newActionsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get a custom action by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2RbacActionsActionWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get action: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&actionTable{})
			}
			return f.Print(&actionTable{actions: []client.CustomK8sAction{*resp.JSON200}})
		},
	}
}

func newActionsCreateCmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom action",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.CustomK8sActionCreateRequest{
				Action:      name,
				Description: strPtr(description),
				K8sRuleset:  client.KubernetesRbacPolicyRule{},
			}

			resp, err := c.PostApiV2RbacActionsWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("create action: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON201 == nil {
				return f.Print(&actionTable{})
			}
			return f.Print(&actionTable{actions: []client.CustomK8sAction{*resp.JSON201}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Action name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Action description")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newActionsUpdateCmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a custom action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.CustomK8sActionUpdateRequest{
				Action:      strPtr(name),
				Description: strPtr(description),
			}

			resp, err := c.PutApiV2RbacActionsIdWithResponse(cmd.Context(), args[0], body)
			if err != nil {
				return fmt.Errorf("update action: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&actionTable{})
			}
			return f.Print(&actionTable{actions: []client.CustomK8sAction{*resp.JSON200}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New action name")
	cmd.Flags().StringVar(&description, "description", "", "New action description")

	return cmd
}

func newActionsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a custom action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			resp, err := c.DeleteApiV2RbacActionsIdWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("delete action: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}
}

type actionTable struct {
	actions []client.CustomK8sAction
}

func (t *actionTable) Headers() []string {
	return []string{"ID", "ACTION", "TYPE", "DESCRIPTION"}
}

func (t *actionTable) Rows() [][]string {
	rows := make([][]string, len(t.actions))
	for i, a := range t.actions {
		rows[i] = []string{a.Id, a.Action, a.Type, a.Description}
	}
	return rows
}

var _ output.TableData = (*actionTable)(nil)
