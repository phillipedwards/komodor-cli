package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/config"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication credentials",
	}
	cmd.AddCommand(newAuthSetKeyCmd(), newAuthShowCmd())
	return cmd
}

func newAuthSetKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-key <key>",
		Short: "Save an API key to the config file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			cfg.APIKey = args[0]
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "API key saved.")
			return nil
		},
	}
}

func newAuthShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the active API key and its source",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			key, source := resolveKeyWithSource(cfg)
			if key == "" {
				return fmt.Errorf("no API key configured — set KOMODOR_API_KEY or run 'komodor auth set-key <key>'")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Key:    %s\nSource: %s\n", maskKey(key), source)
			return nil
		},
	}
}

func resolveKeyWithSource(cfg *config.Config) (string, string) {
	if flagAPIKey != "" {
		return flagAPIKey, "flag"
	}
	if v := os.Getenv("KOMODOR_API_KEY"); v != "" {
		return v, "environment variable"
	}
	if cfg != nil && cfg.APIKey != "" {
		return cfg.APIKey, "config file"
	}
	return "", ""
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return strings.Repeat("*", len(key))
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}
