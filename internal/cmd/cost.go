package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newCostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost",
		Short: "View cost allocation and right-sizing recommendations",
	}

	rightSizing := &cobra.Command{
		Use:   "right-sizing",
		Short: "View right-sizing recommendations",
	}

	rightSizing.AddCommand(
		newCostRightSizingServicesCmd(),
		newCostRightSizingContainersCmd(),
	)

	cmd.AddCommand(
		newCostAllocationCmd(),
		rightSizing,
	)

	return cmd
}

func newCostAllocationCmd() *cobra.Command {
	var (
		timeFrame string
		groupBy   string
		cluster   string
		pageSize  int32
	)

	cmd := &cobra.Command{
		Use:   "allocation",
		Short: "View cost allocation data",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetCostAllocationParams{
				TimeFrame: client.GetCostAllocationParamsTimeFrame(timeFrame),
				GroupBy:   client.GetCostAllocationParamsGroupBy(groupBy),
				PageSize:  pageSize,
			}

			if cluster != "" {
				cs := client.ClusterScopeRequestParam{cluster}
				params.ClusterScope = &cs
			}

			resp, err := c.GetCostAllocationWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("get cost allocation: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}

			if resp.JSON200 == nil || resp.JSON200.GroupedCosts == nil {
				return f.Print(&costAllocationTable{})
			}
			return f.Print(&costAllocationTable{rows: *resp.JSON200.GroupedCosts})
		},
	}

	cmd.Flags().StringVar(&timeFrame, "time-frame", "", "Time frame: Past_7_days, Past_14_days, Past_30_days, Yesterday")
	cmd.Flags().StringVar(&groupBy, "group-by", "", "Group by: clusterName, namespace, komodorServiceName")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Filter by cluster name")
	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "Number of results per page")

	_ = cmd.MarkFlagRequired("time-frame")
	_ = cmd.MarkFlagRequired("group-by")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

func newCostRightSizingServicesCmd() *cobra.Command {
	var (
		cluster              string
		namespace            string
		optimizationStrategy string
		pageSize             int32
	)

	cmd := &cobra.Command{
		Use:   "services",
		Short: "View right-sizing recommendations per service",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetCostRightSizingPerServiceParams{
				OptimizationStrategy: client.GetCostRightSizingPerServiceParamsOptimizationStrategy(optimizationStrategy),
				PageSize:             pageSize,
			}

			if cluster != "" {
				cs := client.ClusterScopeRequestParam{cluster}
				params.ClusterScope = &cs
			}

			if namespace != "" {
				fb := client.GetCostRightSizingPerServiceParamsFilterBy("namespace")
				fv := namespace
				params.FilterBy = &fb
				params.FilterValueEquals = &fv
			}

			resp, err := c.GetCostRightSizingPerServiceWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("get right-sizing per service: %w", err)
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

	cmd.Flags().StringVar(&cluster, "cluster", "", "Filter by cluster name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Filter by namespace")
	cmd.Flags().StringVar(&optimizationStrategy, "strategy", "moderate", "Optimization strategy: conservative, moderate, aggressive")
	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "Number of results per page")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

func newCostRightSizingContainersCmd() *cobra.Command {
	var (
		cluster     string
		namespace   string
		serviceKind string
		serviceName string
	)

	cmd := &cobra.Command{
		Use:   "containers",
		Short: "View right-sizing recommendations per container",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			params := &client.GetCostRightSizingPerContainerParams{
				ClusterName: cluster,
				Namespace:   namespace,
				ServiceKind: serviceKind,
				ServiceName: serviceName,
			}

			resp, err := c.GetCostRightSizingPerContainerWithResponse(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("get right-sizing per container: %w", err)
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

	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name (required)")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace (required)")
	cmd.Flags().StringVar(&serviceKind, "kind", "", "Service kind e.g. Deployment (required)")
	cmd.Flags().StringVar(&serviceName, "name", "", "Service name (required)")

	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("namespace")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

type costAllocationTable struct {
	rows []client.CostAllocationSummaryRow
}

func (t *costAllocationTable) Headers() []string {
	return []string{"SERVICE", "CLUSTER", "NAMESPACE", "TOTAL_COST", "POTENTIAL_SAVING", "OPTIMIZATION_SCORE"}
}

func (t *costAllocationTable) Rows() [][]string {
	rows := make([][]string, len(t.rows))
	for i, r := range t.rows {
		cluster := ""
		if r.ClusterName != nil {
			cluster = *r.ClusterName
		}
		rows[i] = []string{
			r.KomodorServiceName,
			cluster,
			r.Namespace,
			fmt.Sprintf("%.4f", r.TotalCost),
			fmt.Sprintf("%.4f", r.PotentialSaving),
			fmt.Sprintf("%.2f", r.OptimizationScore),
		}
	}
	return rows
}

var _ output.TableData = (*costAllocationTable)(nil)
