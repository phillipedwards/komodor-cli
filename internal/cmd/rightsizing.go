package cmd

import (
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newRightSizingPoliciesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "right-sizing-policies",
		Aliases: []string{"rsp"},
		Short:   "Manage right-sizing policies",
	}

	cmd.AddCommand(
		newRSPListCmd(),
		newRSPGetCmd(),
		newRSPDefaultsCmd(),
		newRSPCreateCmd(),
		newRSPUpdateCmd(),
		newRSPDeleteCmd(),
	)

	return cmd
}

func newRSPListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all right-sizing policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.ListRightSizingPoliciesWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("list right-sizing policies: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&rspTable{})
			}

			all, err := resp.JSON200.AsGetAllRightSizingPoliciesResponse()
			if err != nil {
				return f.Print(resp.JSON200)
			}
			return f.Print(&rspTable{rows: all.Policies})
		},
	}
}

func newRSPGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a right-sizing policy by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			id, err := parseUUID(args[0])
			if err != nil {
				return err
			}

			resp, err := c.GetRightSizingPolicyByIdWithResponse(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("get right-sizing policy: %w", err)
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

func newRSPDefaultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "defaults",
		Short: "Get the default right-sizing policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetDefaultRightSizingPolicyWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("get default right-sizing policy: %w", err)
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

func newRSPCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a right-sizing policy",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			policy := client.RightSizingMultiScopePolicy{
				Name:               name,
				OptimizationPreset: client.OptimizationPresetType("moderate"),
				ApplyProtocol:      client.RightSizingMultiScopePolicyApplyProtocol("onCreation"),
				Priority:           0,
				Scopes:             []client.PolicyResourceScope{},
			}

			var body client.RightSizingPolicyRequest
			if err := body.FromRightSizingMultiScopePolicy(policy); err != nil {
				return fmt.Errorf("build policy request: %w", err)
			}

			resp, err := c.CreateRightSizingPolicyWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("create right-sizing policy: %w", err)
			}
			if resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON201 == nil {
				return f.Print(map[string]interface{}{})
			}
			return f.Print(resp.JSON201)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Policy name")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newRSPUpdateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a right-sizing policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			id, err := parseUUID(args[0])
			if err != nil {
				return err
			}

			// Fetch existing to preserve fields.
			getResp, err := c.GetRightSizingPolicyByIdWithResponse(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("get right-sizing policy: %w", err)
			}
			if getResp.StatusCode() != 200 {
				return apiError(getResp.StatusCode(), getResp.Body)
			}

			existing, err := getResp.JSON200.AsGetMultiScopePolicyResponse()
			if err != nil {
				return fmt.Errorf("parse existing policy: %w", err)
			}

			policy := existing.Policy
			if cmd.Flags().Changed("name") {
				policy.Name = name
			}

			var body client.RightSizingPolicyRequest
			if err := body.FromRightSizingMultiScopePolicy(policy); err != nil {
				return fmt.Errorf("build policy request: %w", err)
			}

			resp, err := c.UpdateRightSizingPolicyByIdWithResponse(cmd.Context(), id, body)
			if err != nil {
				return fmt.Errorf("update right-sizing policy: %w", err)
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

	cmd.Flags().StringVar(&name, "name", "", "New policy name")

	return cmd
}

func newRSPDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a right-sizing policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			id, err := parseUUID(args[0])
			if err != nil {
				return err
			}

			params := &client.DeleteRightSizingPolicyByIdParams{}
			if cmd.Flags().Changed("force") {
				params.Force = boolPtr(force)
			}

			resp, err := c.DeleteRightSizingPolicyByIdWithResponse(cmd.Context(), id, params)
			if err != nil {
				return fmt.Errorf("delete right-sizing policy: %w", err)
			}
			if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			fmt.Printf("right-sizing policy %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force delete even if workloads are affected")

	return cmd
}

type rspTable struct {
	rows []client.GetAllRightSizingPoliciesRow
}

func (t *rspTable) Headers() []string {
	return []string{"ID", "NAME", "CREATED_AT"}
}

func (t *rspTable) Rows() [][]string {
	rows := make([][]string, len(t.rows))
	for i, r := range t.rows {
		rows[i] = []string{
			r.Id.String(),
			r.Name,
			r.CreatedAt,
		}
	}
	return rows
}

var _ output.TableData = (*rspTable)(nil)

func parseUUID(s string) (openapi_types.UUID, error) {
	var id openapi_types.UUID
	if err := id.UnmarshalText([]byte(s)); err != nil {
		return id, fmt.Errorf("invalid UUID %q: %w", s, err)
	}
	return id, nil
}
