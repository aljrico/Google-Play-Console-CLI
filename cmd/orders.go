package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newOrdersCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "orders",
		Short: "Inspect and refund Google Play orders",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newOrdersGetCommand(out, options, &packageName),
		newOrdersBatchGetCommand(out, options, &packageName),
		newOrdersRefundCommand(out, options, &packageName),
	)
	return cmd
}

func newOrdersGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var orderID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one Google Play order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedOrderID, err := play.NewOrderID(orderID)
			if err != nil {
				return err
			}
			getOptions := play.OrderGetOptions{
				PackageName: typedPackageName,
				OrderID:     typedOrderID,
			}
			if err := getOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.GetOrder(cmd.Context(), publisher, getOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&orderID, "order-id", "", "Google Play order ID")
	return cmd
}

func newOrdersBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var orderIDs []string

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple Google Play orders",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedOrderIDs, err := parseOrderIDs(orderIDs)
			if err != nil {
				return err
			}
			getOptions := play.OrderBatchGetOptions{
				PackageName: typedPackageName,
				OrderIDs:    typedOrderIDs,
			}
			if err := getOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchGetOrders(cmd.Context(), publisher, getOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringArrayVar(&orderIDs, "order-id", nil, "Google Play order ID, repeatable")
	return cmd
}

func newOrdersRefundCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		orderID string
		revoke  bool
		confirm bool
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "refund",
		Short: "Refund one Google Play order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedOrderID, err := play.NewOrderID(orderID)
			if err != nil {
				return err
			}
			refundOptions := play.OrderRefundOptions{
				PackageName: typedPackageName,
				OrderID:     typedOrderID,
				Revoke:      revoke,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if err := refundOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.RefundOrder(cmd.Context(), nil, refundOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.RefundOrder(cmd.Context(), publisher, refundOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&orderID, "order-id", "", "Google Play order ID")
	cmd.Flags().BoolVar(&revoke, "revoke", false, "Revoke the purchased item after refunding")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the refund")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned refund without calling Google Play")
	return cmd
}

func parseOrderIDs(values []string) ([]play.OrderID, error) {
	orderIDs := make([]play.OrderID, 0, len(values))
	for _, value := range values {
		orderID, err := play.NewOrderID(value)
		if err != nil {
			return nil, err
		}
		orderIDs = append(orderIDs, orderID)
	}
	return orderIDs, nil
}
