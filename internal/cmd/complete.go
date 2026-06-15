package cmd

import (
	"context"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/config"
)

// newCompletionClient builds an API client for use inside shell completion functions.
// It reads auth from the environment and config file only — flags are not reliably
// parsed during completion. Returns nil on any failure; callers must handle nil.
func newCompletionClient() *client.ClientWithResponses {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	apiKey := os.Getenv("KOMODOR_API_KEY")
	if apiKey == "" {
		apiKey = cfg.APIKey
	}
	if apiKey == "" {
		return nil
	}

	baseURL := cfg.BaseURL
	c, err := client.NewClientWithResponses(baseURL, client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-api-key", apiKey)
		return nil
	}))
	if err != nil {
		return nil
	}
	return c
}

// completeClusterNames returns cluster names for --cluster flag completion.
func completeClusterNames(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	c := newCompletionClient()
	if c == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	resp, err := c.GetApiV2ClustersWithResponse(cmd.Context(), &client.GetApiV2ClustersParams{})
	if err != nil || resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, len(resp.JSON200.Data.Clusters))
	for i, cl := range resp.JSON200.Data.Clusters {
		names[i] = cl.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeRoleNames returns role names for --role flag completion.
func completeRoleNames(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	c := newCompletionClient()
	if c == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	resp, err := c.GetApiV2RbacRolesWithResponse(cmd.Context())
	if err != nil || resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, len(*resp.JSON200))
	for i, r := range *resp.JSON200 {
		names[i] = r.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completePolicyNames returns policy names for --policy flag completion.
func completePolicyNames(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	c := newCompletionClient()
	if c == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	resp, err := c.GetApiV2RbacPoliciesWithResponse(cmd.Context())
	if err != nil || resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, len(*resp.JSON200))
	for i, p := range *resp.JSON200 {
		names[i] = p.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeServiceNames returns service names for --service-id flag completion.
func completeServiceNames(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	c := newCompletionClient()
	if c == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	resp, err := c.PostApiV2ServicesSearchWithResponse(cmd.Context(), client.SearchServicesBody{})
	if err != nil || resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	svcs := resp.JSON200.Data.Services
	names := make([]string, len(svcs))
	for i, s := range svcs {
		names[i] = s.Service
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
