package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/auth"
	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/config"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

type contextKey string

const (
	keyClient    contextKey = "client"
	keyFormatter contextKey = "formatter"
)

var (
	flagAPIKey  string
	flagOutput  string
	flagBaseURL string
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "komodor",
		Short: "CLI for the Komodor API",
		Long:  "komodor is a command-line interface for the Komodor platform API.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip auth for auth sub-commands and completion
			if skipAuth(cmd) {
				return nil
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// Base URL: flag > config
			baseURL := cfg.BaseURL
			if flagBaseURL != "" {
				baseURL = flagBaseURL
			}

			apiKey, err := auth.Resolve(flagAPIKey, cfg)
			if err != nil {
				return err
			}

			c, err := client.NewClientWithResponses(baseURL, client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
				req.Header.Set("x-api-key", apiKey)
				return nil
			}))
			if err != nil {
				return fmt.Errorf("create API client: %w", err)
			}

			formatter, err := output.New(flagOutput, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			cmd.SetContext(context.WithValue(
				context.WithValue(cmd.Context(), keyClient, c),
				keyFormatter, formatter,
			))
			return nil
		},
	}

	root.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "Komodor API key (overrides KOMODOR_API_KEY and config file)")
	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "Output format: table|json|yaml|csv")
	root.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "Komodor API base URL override")

	root.AddCommand(
		newAuthCmd(),
		newAPIKeyCmd(),
		newServicesCmd(),
		newClustersCmd(),
		newJobsCmd(),
		newEventsCmd(),
		newIssuesCmd(),
		newHealthCmd(),
		newKubeconfigCmd(),
		newAuditCmd(),
		newUsersCmd(),
		newRolesCmd(),
		newPoliciesCmd(),
		newActionsCmd(),
		newRBACCmd(),
		newMonitorsCmd(),
		newIntegrationsCmd(),
		newCostCmd(),
		newRightSizingPoliciesCmd(),
		newWorkspacesCmd(),
		newKlaudiaCmd(),
		newCompletionCmd(),
	)

	return root
}

// Execute runs the root command.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func clientFromCtx(cmd *cobra.Command) client.ClientWithResponsesInterface {
	return cmd.Context().Value(keyClient).(client.ClientWithResponsesInterface)
}

func formatterFromCtx(cmd *cobra.Command) output.Formatter {
	return cmd.Context().Value(keyFormatter).(output.Formatter)
}

// skipAuth returns true for commands that don't need an API key yet.
func skipAuth(cmd *cobra.Command) bool {
	exempt := map[string]bool{
		"set-key":          true,
		"show":             true,
		"completion":       true,
		"bash":             true,
		"zsh":              true,
		"fish":             true,
		"help":             true,
		"__complete":       true,
		"__completeNoDesc": true,
	}
	return exempt[cmd.Name()]
}
