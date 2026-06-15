package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"event"},
		Short:   "Manage Komodor events",
	}

	cmd.AddCommand(
		newEventsSearchCmd(),
		newEventsSearchClusterCmd(),
		newEventsCreateCmd(),
	)

	return cmd
}

// eventsTable wraps a slice of SingleEvent for table output.
type eventsTable struct {
	events []client.SingleEvent
}

func (t eventsTable) Headers() []string {
	return []string{"SUMMARY", "TYPE", "CLUSTER", "NAMESPACE", "KIND", "TIME"}
}

func (t eventsTable) Rows() [][]string {
	rows := make([][]string, len(t.events))
	for i, e := range t.events {
		rows[i] = []string{
			e.Summary,
			string(e.Type),
			e.Resource.Cluster,
			output.Str(e.Resource.Namespace),
			e.Resource.Kind,
			time.Unix(e.StartTime, 0).UTC().Format(time.RFC3339),
		}
	}
	return rows
}

func newEventsSearchCmd() *cobra.Command {
	var (
		serviceID string
		from      string
		to        string
		pageSize  int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search events for a service",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := client.SearchServicesK8sEventsBody{
				Scope: client.ServiceScope{
					Service: serviceID,
				},
				Props: client.K8sEventProps{
					Type: "all",
				},
			}

			if from != "" {
				t, err := time.Parse(time.RFC3339, from)
				if err != nil {
					return fmt.Errorf("invalid --from value: %w", err)
				}
				epoch := client.EpochTime(t.Unix())
				body.Props.FromEpoch = &epoch
			}

			if to != "" {
				t, err := time.Parse(time.RFC3339, to)
				if err != nil {
					return fmt.Errorf("invalid --to value: %w", err)
				}
				epoch := client.EpochTime(t.Unix())
				body.Props.ToEpoch = &epoch
			}

			if pageSize > 0 {
				body.Pagination = &client.PaginationTokenParams{PageSize: &pageSize}
			}

			resp, err := clientFromCtx(cmd).PostApiV2ServicesK8sEventsSearchWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("search service events: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			return formatterFromCtx(cmd).Print(eventsTable{events: resp.JSON200.Data.Events})
		},
	}

	cmd.Flags().StringVar(&serviceID, "service-id", "", "Service ID to search events for (required)")
	cmd.Flags().StringVar(&from, "from", "", "Start time in RFC3339 format")
	cmd.Flags().StringVar(&to, "to", "", "End time in RFC3339 format")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Number of results per page")
	_ = cmd.MarkFlagRequired("service-id")
	_ = cmd.RegisterFlagCompletionFunc("service-id", completeServiceNames)

	return cmd
}

func newEventsSearchClusterCmd() *cobra.Command {
	var (
		cluster  string
		from     string
		to       string
		pageSize int
	)

	cmd := &cobra.Command{
		Use:   "search-cluster",
		Short: "Search events for a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := client.SearchClustersK8sEventsBody{
				Scope: client.ClusterScope{
					Cluster: cluster,
				},
				Props: client.K8sEventProps{
					Type: "all",
				},
			}

			if from != "" {
				t, err := time.Parse(time.RFC3339, from)
				if err != nil {
					return fmt.Errorf("invalid --from value: %w", err)
				}
				epoch := client.EpochTime(t.Unix())
				body.Props.FromEpoch = &epoch
			}

			if to != "" {
				t, err := time.Parse(time.RFC3339, to)
				if err != nil {
					return fmt.Errorf("invalid --to value: %w", err)
				}
				epoch := client.EpochTime(t.Unix())
				body.Props.ToEpoch = &epoch
			}

			if pageSize > 0 {
				body.Pagination = &client.PaginationTokenParams{PageSize: &pageSize}
			}

			resp, err := clientFromCtx(cmd).PostApiV2ClustersK8sEventsSearchWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("search cluster events: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			return formatterFromCtx(cmd).Print(eventsTable{events: resp.JSON200.Data.Events})
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name to search events for (required)")
	cmd.Flags().StringVar(&from, "from", "", "Start time in RFC3339 format")
	cmd.Flags().StringVar(&to, "to", "", "End time in RFC3339 format")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "Number of results per page")
	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

func newEventsCreateCmd() *cobra.Command {
	var (
		serviceID string
		title     string
		message   string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom event",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := client.CustomEventDto{
				EventType: title,
				Summary:   title,
			}

			if message != "" {
				body.Summary = message
			}

			if serviceID != "" {
				body.Scope = &struct {
					Clusters      *[]string `json:"clusters,omitempty"`
					Namespaces    *[]string `json:"namespaces,omitempty"`
					ServicesNames *[]string `json:"servicesNames,omitempty"`
				}{
					ServicesNames: &[]string{serviceID},
				}
			}

			params := &client.EventsControllerCreateCustomEventParams{
				XApiKey: "",
			}

			resp, err := clientFromCtx(cmd).EventsControllerCreateCustomEventWithResponse(cmd.Context(), params, body)
			if err != nil {
				return fmt.Errorf("create custom event: %w", err)
			}
			if resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON201 != nil && resp.JSON201.Message != nil {
				fmt.Println(output.Str(resp.JSON201.Message))
			} else {
				fmt.Println("Event created successfully.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serviceID, "service-id", "", "Service ID to associate the event with")
	cmd.Flags().StringVar(&title, "title", "", "Event title / type (required)")
	cmd.Flags().StringVar(&message, "message", "", "Event message / summary")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.RegisterFlagCompletionFunc("service-id", completeServiceNames)

	return cmd
}
