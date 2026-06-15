package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newRolesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "roles",
		Aliases: []string{"role"},
		Short:   "Manage RBAC roles",
	}

	cmd.AddCommand(
		newRolesListCmd(),
		newRolesGetCmd(),
		newRolesCreateCmd(),
		newRolesUpdateCmd(),
		newRolesDeleteCmd(),
		newRolesAttachPolicyCmd(),
		newRolesDetachPolicyCmd(),
	)

	return cmd
}

func newRolesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2RbacRolesWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("list roles: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&roleTable{})
			}
			return f.Print(&roleTable{roles: *resp.JSON200})
		},
	}
}

func newRolesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-name>",
		Short: "Get a role by ID or name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2RbacRolesIdOrNameWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get role: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&roleTable{})
			}
			return f.Print(&roleTable{roles: []client.Role{*resp.JSON200}})
		},
	}
}

func newRolesCreateCmd() *cobra.Command {
	var name string
	var policyIDs []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new role",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.RoleCreateRequest{
				Name:      name,
				PolicyIds: strSlicePtr(policyIDs),
			}

			resp, err := c.PostApiV2RbacRolesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("create role: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON201 == nil {
				return f.Print(&roleTable{})
			}
			return f.Print(&roleTable{roles: []client.Role{*resp.JSON201}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Role name (required)")
	cmd.Flags().StringSliceVar(&policyIDs, "policy", nil, "Policy IDs to assign (may be repeated)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.RegisterFlagCompletionFunc("policy", completePolicyNames)

	return cmd
}

func newRolesUpdateCmd() *cobra.Command {
	var name string
	var policyIDs []string

	cmd := &cobra.Command{
		Use:   "update <id-or-name>",
		Short: "Update a role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.UpdateRoleRequest{
				Name:      strPtr(name),
				PolicyIds: strSlicePtr(policyIDs),
			}

			resp, err := c.PutApiV2RbacRolesIdOrNameWithResponse(cmd.Context(), args[0], body)
			if err != nil {
				return fmt.Errorf("update role: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&roleTable{})
			}
			return f.Print(&roleTable{roles: []client.Role{*resp.JSON200}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New role name")
	cmd.Flags().StringSliceVar(&policyIDs, "policy", nil, "Policy IDs to assign (may be repeated)")
	_ = cmd.RegisterFlagCompletionFunc("policy", completePolicyNames)

	return cmd
}

func newRolesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id-or-name>",
		Short: "Delete a role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			resp, err := c.DeleteApiV2RbacRolesIdOrNameWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("delete role: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}
}

func newRolesAttachPolicyCmd() *cobra.Command {
	var roleID, policyID string

	cmd := &cobra.Command{
		Use:   "attach-policy",
		Short: "Attach a policy to a role",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			body := client.RbacRolePolicyCreateRequest{
				RoleId:   roleID,
				PolicyId: policyID,
			}

			resp, err := c.PostApiV2RbacRolesPoliciesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("attach policy: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}

	cmd.Flags().StringVar(&roleID, "role", "", "Role ID (required)")
	cmd.Flags().StringVar(&policyID, "policy", "", "Policy ID (required)")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("policy")
	_ = cmd.RegisterFlagCompletionFunc("role", completeRoleNames)
	_ = cmd.RegisterFlagCompletionFunc("policy", completePolicyNames)

	return cmd
}

func newRolesDetachPolicyCmd() *cobra.Command {
	var roleID, policyID string

	cmd := &cobra.Command{
		Use:   "detach-policy",
		Short: "Detach a policy from a role",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			body := client.RolePolicyDeleteRequest{
				RoleId:   roleID,
				PolicyId: policyID,
			}

			resp, err := c.DeleteApiV2RbacRolesPoliciesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("detach policy: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}

	cmd.Flags().StringVar(&roleID, "role", "", "Role ID (required)")
	cmd.Flags().StringVar(&policyID, "policy", "", "Policy ID (required)")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("policy")
	_ = cmd.RegisterFlagCompletionFunc("role", completeRoleNames)
	_ = cmd.RegisterFlagCompletionFunc("policy", completePolicyNames)

	return cmd
}

type roleTable struct {
	roles []client.Role
}

func (t *roleTable) Headers() []string {
	return []string{"ID", "NAME", "DEFAULT", "POLICIES"}
}

func (t *roleTable) Rows() [][]string {
	rows := make([][]string, len(t.roles))
	for i, r := range t.roles {
		policyNames := make([]string, len(r.Policies))
		for j, p := range r.Policies {
			policyNames[j] = p.Name
		}
		isDefault := "false"
		if r.IsDefault {
			isDefault = "true"
		}
		rows[i] = []string{r.Id, r.Name, isDefault, joinStrings(policyNames)}
	}
	return rows
}

var _ output.TableData = (*roleTable)(nil)
