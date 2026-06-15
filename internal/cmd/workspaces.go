package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newWorkspacesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspaces",
		Aliases: []string{"ws"},
		Short:   "Manage workspaces",
	}
	cmd.AddCommand(
		newWorkspacesListCmd(),
		newWorkspacesGetCmd(),
		newWorkspacesCreateCmd(),
		newWorkspacesUpdateCmd(),
		newWorkspacesDeleteCmd(),
	)
	return cmd
}

func newWorkspacesListCmd() *cobra.Command {
	var page, pageSize int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)
			params := &client.GetApiV2WorkspacesParams{}
			if pageSize > 0 {
				ps := pageSize
				params.PageSize = &ps
			}
			if page > 0 {
				params.Page = &page
			}
			resp, err := c.GetApiV2WorkspacesWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("list workspaces: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			var ws []client.Workspace
			if resp.JSON200 != nil && resp.JSON200.Workspaces != nil {
				ws = *resp.JSON200.Workspaces
			}
			return f.Print(&workspacesTable{rows: ws})
		},
	}
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Results per page")
	return cmd
}

func newWorkspacesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get workspace by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)
			resp, err := c.GetApiV2WorkspacesIdWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get workspace: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			return f.Print(resp.JSON200)
		},
	}
}

func newWorkspacesCreateCmd() *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)
			body := client.WorkspaceRequest{
				Name:   name,
				Scopes: []client.SchemasResourcesScope{},
			}
			if description != "" {
				body.Description = &description
			}
			resp, err := c.PostApiV2WorkspacesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("create workspace: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			return f.Print(resp.JSON201)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Workspace name")
	cmd.Flags().StringVar(&description, "description", "", "Workspace description")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newWorkspacesUpdateCmd() *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)
			body := client.WorkspaceRequest{
				Name:   name,
				Scopes: []client.SchemasResourcesScope{},
			}
			if description != "" {
				body.Description = &description
			}
			resp, err := c.PutApiV2WorkspacesIdWithResponse(cmd.Context(), args[0], body)
			if err != nil {
				return fmt.Errorf("update workspace: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			return f.Print(resp.JSON200)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New workspace name")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newWorkspacesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			resp, err := c.DeleteApiV2WorkspacesIdWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("delete workspace: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			cmd.Println("Deleted")
			return nil
		},
	}
}

type workspacesTable struct{ rows []client.Workspace }

func (t *workspacesTable) Headers() []string {
	return []string{"ID", "NAME", "CREATED_AT", "LAST_UPDATED"}
}

func (t *workspacesTable) Rows() [][]string {
	rows := make([][]string, len(t.rows))
	for i, w := range t.rows {
		rows[i] = []string{
			w.Id,
			w.Name,
			w.CreatedAt.Format("2006-01-02"),
			w.LastUpdated.Format("2006-01-02"),
		}
	}
	return rows
}

var _ output.TableData = (*workspacesTable)(nil)
