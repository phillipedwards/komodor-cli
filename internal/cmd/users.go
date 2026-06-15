package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "Manage users",
	}

	cmd.AddCommand(
		newUsersListCmd(),
		newUsersGetCmd(),
		newUsersCreateCmd(),
		newUsersUpdateCmd(),
		newUsersDeleteCmd(),
		newUsersEffectivePermissionsCmd(),
	)

	return cmd
}

func newUsersListCmd() *cobra.Command {
	var email string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetApiV2UsersParams{
				Email: strPtr(email),
			}

			resp, err := c.GetApiV2UsersWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("list users: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&userTable{})
			}
			return f.Print(&userTable{users: *resp.JSON200})
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Filter by email address")

	return cmd
}

func newUsersGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-email>",
		Short: "Get a user by ID or email",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2UsersIdOrEmailWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get user: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&userTable{})
			}
			return f.Print(&userTable{users: []client.User{*resp.JSON200}})
		},
	}
}

func newUsersCreateCmd() *cobra.Command {
	var email, name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.CreateUserRequest{
				Email:       email,
				DisplayName: name,
			}

			resp, err := c.PostApiV2UsersWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("create user: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&userTable{})
			}
			return f.Print(&userTable{users: []client.User{*resp.JSON200}})
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "User email address (required)")
	cmd.Flags().StringVar(&name, "name", "", "User display name")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}

func newUsersUpdateCmd() *cobra.Command {
	var name string
	var roleIDs []string

	cmd := &cobra.Command{
		Use:   "update <id-or-email>",
		Short: "Update a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.UpdateUserRequest{
				DisplayName: strPtr(name),
				RoleIds:     strSlicePtr(roleIDs),
			}

			resp, err := c.PutApiV2UsersIdOrEmailWithResponse(cmd.Context(), args[0], body)
			if err != nil {
				return fmt.Errorf("update user: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&userTable{})
			}
			return f.Print(&userTable{users: []client.User{*resp.JSON200}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New display name")
	cmd.Flags().StringSliceVar(&roleIDs, "role", nil, "Role IDs to assign (may be repeated)")
	_ = cmd.RegisterFlagCompletionFunc("role", completeRoleNames)

	return cmd
}

func newUsersDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id-or-email>",
		Short: "Delete a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			resp, err := c.DeleteApiV2UsersIdOrEmailWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("delete user: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}
}

func newUsersEffectivePermissionsCmd() *cobra.Command {
	var id, email string

	cmd := &cobra.Command{
		Use:   "effective-permissions",
		Short: "Get effective permissions for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetApiV2UsersEffectivePermissionsParams{
				Id:    strPtr(id),
				Email: strPtr(email),
			}

			resp, err := c.GetApiV2UsersEffectivePermissionsWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("get effective permissions: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&effectivePermissionsTable{})
			}
			return f.Print(&effectivePermissionsTable{perms: *resp.JSON200})
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "User ID")
	cmd.Flags().StringVar(&email, "email", "", "User email")

	return cmd
}

type userTable struct {
	users []client.User
}

func (t *userTable) Headers() []string {
	return []string{"ID", "EMAIL", "NAME", "ROLES"}
}

func (t *userTable) Rows() [][]string {
	rows := make([][]string, len(t.users))
	for i, u := range t.users {
		roleNames := make([]string, len(u.Roles))
		for j, r := range u.Roles {
			roleNames[j] = r.Name
		}
		rows[i] = []string{u.Id, u.Email, u.DisplayName, joinStrings(roleNames)}
	}
	return rows
}

type effectivePermissionsTable struct {
	perms []client.EffectivePermission
}

func (t *effectivePermissionsTable) Headers() []string {
	return []string{"ACTION", "CLUSTER", "NAMESPACE", "POLICY", "ROLE"}
}

func (t *effectivePermissionsTable) Rows() [][]string {
	rows := make([][]string, len(t.perms))
	for i, p := range t.perms {
		rows[i] = []string{p.Action, p.Cluster, output.Str(p.Namespace), p.PolicyName, p.RoleName}
	}
	return rows
}

var _ output.TableData = (*userTable)(nil)
var _ output.TableData = (*effectivePermissionsTable)(nil)
