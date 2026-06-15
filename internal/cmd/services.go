package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "services",
		Aliases: []string{"svc"},
		Short:   "Manage and inspect services",
	}

	cmd.AddCommand(
		newServicesSearchCmd(),
		newServicesYAMLCmd(),
	)

	return cmd
}

func newServicesSearchCmd() *cobra.Command {
	var (
		cluster   string
		namespace string
		name      string
		page      int
		pageSize  int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for services",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.SearchServicesBody{}

			if cluster != "" || namespace != "" {
				scope := &client.OptionalClusterScope{}
				if cluster != "" {
					scope.Cluster = strPtr(cluster)
				}
				if namespace != "" {
					scope.Namespaces = &[]string{namespace}
				}
				body.Scope = scope
			}

			if cmd.Flags().Changed("page") || cmd.Flags().Changed("page-size") {
				pagination := &client.PaginationTokenParams{}
				if cmd.Flags().Changed("page-size") {
					pagination.PageSize = &pageSize
				}
				// The services search API uses cursor-based pagination; --page maps to the token concept
				// but is not directly supported. Warn the user if --page is set.
				if cmd.Flags().Changed("page") {
					fmt.Fprintf(os.Stderr, "warning: --page is not supported for service search (cursor-based pagination); use the token from previous response\n")
				}
				body.Pagination = pagination
			}

			// Filter by name via kind: use the name as a service name filter is not directly
			// available in SearchServicesBody, so we leave it unset and note this limitation.
			_ = name

			resp, err := c.PostApiV2ServicesSearchWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("search services: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil {
				return f.Print(&serviceSearchTable{})
			}
			return f.Print(&serviceSearchTable{services: resp.JSON200.Data.Services})
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Filter by cluster name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Filter by namespace")
	cmd.Flags().StringVar(&name, "name", "", "Filter by service name (note: not supported server-side; provided for future use)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (not supported for service search; cursor-based pagination only)")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Number of results per page")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

func newServicesYAMLCmd() *cobra.Command {
	var (
		cluster   string
		namespace string
		kind      string
		name      string
	)

	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Get the YAML for a service",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			params := &client.GetApiV2ServiceYamlParams{
				Cluster:   cluster,
				Namespace: namespace,
				Kind:      client.ServiceKind(kind),
				Name:      name,
			}

			resp, err := c.GetApiV2ServiceYamlWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("get service yaml: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			_, err = fmt.Fprint(os.Stdout, string(resp.Body))
			return err
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name (required)")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace (required)")
	cmd.Flags().StringVar(&kind, "kind", "", "Service kind, e.g. deployment, statefulset (required)")
	cmd.Flags().StringVar(&name, "name", "", "Service name (required)")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("namespace")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// serviceSearchTable renders SingleService slices as a table.
type serviceSearchTable struct {
	services []client.SingleService
}

func (t *serviceSearchTable) Headers() []string {
	return []string{"NAME", "NAMESPACE", "CLUSTER", "KIND"}
}

func (t *serviceSearchTable) Rows() [][]string {
	rows := make([][]string, len(t.services))
	for i, s := range t.services {
		rows[i] = []string{s.Service, s.Namespace, s.Cluster, s.Kind}
	}
	return rows
}

var _ output.TableData = (*serviceSearchTable)(nil)
