package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newPoliciesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policies",
		Aliases: []string{"policy"},
		Short:   "Manage RBAC policies",
	}

	cmd.AddCommand(
		newPoliciesListCmd(),
		newPoliciesGetCmd(),
		newPoliciesCreateCmd(),
		newPoliciesUpdateCmd(),
		newPoliciesDeleteCmd(),
	)

	return cmd
}

func newPoliciesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2RbacPoliciesWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("list policies: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&policyTable{})
			}
			return f.Print(&policyTable{policies: *resp.JSON200})
		},
	}
}

func newPoliciesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-name>",
		Short: "Get a policy by ID or name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetApiV2RbacPoliciesIdOrNameWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get policy: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&policyTable{})
			}
			return f.Print(&policyTable{policies: []client.Policy{*resp.JSON200}})
		},
	}
}

func newPoliciesCreateCmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.PolicyCreateRequest{
				Name:        name,
				Description: strPtr(description),
				Statements:  []client.Statement{},
			}

			resp, err := c.PostApiV2RbacPoliciesWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("create policy: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&policyTable{})
			}
			return f.Print(&policyTable{policies: []client.Policy{*resp.JSON200}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Policy name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Policy description")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newPoliciesUpdateCmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "update <id-or-name>",
		Short: "Update a policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.UpdatePolicyRequest{
				Name:        strPtr(name),
				Description: strPtr(description),
			}

			resp, err := c.PutApiV2RbacPoliciesIdOrNameWithResponse(cmd.Context(), args[0], body)
			if err != nil {
				return fmt.Errorf("update policy: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&policyTable{})
			}
			return f.Print(&policyTable{policies: []client.Policy{*resp.JSON200}})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New policy name")
	cmd.Flags().StringVar(&description, "description", "", "New policy description")

	return cmd
}

func newPoliciesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id-or-name>",
		Short: "Delete a policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			resp, err := c.DeleteApiV2RbacPoliciesIdOrNameWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("delete policy: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Println("Success")
			return nil
		},
	}
}

type policyTable struct {
	policies []client.Policy
}

func (t *policyTable) Headers() []string {
	return []string{"ID", "NAME", "DESCRIPTION", "CREATED_AT"}
}

func (t *policyTable) Rows() [][]string {
	rows := make([][]string, len(t.policies))
	for i, p := range t.policies {
		rows[i] = []string{p.Id, p.Name, output.Str(p.Description), p.CreatedAt.Format("2006-01-02")}
	}
	return rows
}

var _ output.TableData = (*policyTable)(nil)
