package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
)

func newRBACCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rbac",
		Short: "Manage RBAC assignments",
	}

	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Manage user-role assignments",
	}

	usersCmd.AddCommand(
		newRBACUsersAttachCmd(),
		newRBACUsersUpdateCmd(),
		newRBACUsersDetachCmd(),
	)

	cmd.AddCommand(usersCmd)

	return cmd
}

func newRBACUsersAttachCmd() *cobra.Command {
	var userID, roleID string

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a role to a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			body := client.UserRoleCreateRequest{
				UserId: userID,
				RoleId: roleID,
			}

			resp, err := c.PostApiV2RbacUsersRolesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("attach role: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	cmd.Flags().StringVar(&roleID, "role", "", "Role ID (required)")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.RegisterFlagCompletionFunc("role", completeRoleNames)

	return cmd
}

func newRBACUsersUpdateCmd() *cobra.Command {
	var userID, roleID, expiration string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a user-role assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			body := client.UserRoleUpdateRequest{
				UserId:     userID,
				RoleId:     roleID,
				Expiration: expiration,
			}

			resp, err := c.PutApiV2RbacUsersRolesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("update user role: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	cmd.Flags().StringVar(&roleID, "role", "", "Role ID (required)")
	cmd.Flags().StringVar(&expiration, "expiration", "", "Expiration date (RFC3339)")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.RegisterFlagCompletionFunc("role", completeRoleNames)

	return cmd
}

func newRBACUsersDetachCmd() *cobra.Command {
	var userID, roleID string

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach a role from a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			body := client.UserRoleDeleteRequest{
				UserId: userID,
				RoleId: roleID,
			}

			resp, err := c.DeleteApiV2RbacUsersRolesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("detach role: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	cmd.Flags().StringVar(&roleID, "role", "", "Role ID (required)")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.RegisterFlagCompletionFunc("role", completeRoleNames)

	return cmd
}
