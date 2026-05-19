package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newSubscriptionsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Inspect Google Play monetization subscriptions",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newSubscriptionsListCommand(out, options, &packageName),
		newSubscriptionsGetCommand(out, options, &packageName),
		newSubscriptionsBatchGetCommand(out, options, &packageName),
	)
	return cmd
}

func newSubscriptionsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		pageSize     int64
		pageToken    string
		showArchived bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List monetization subscriptions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.SubscriptionListOptions{
				PackageName:  typedPackageName,
				PageSize:     pageSize,
				PageToken:    pageToken,
				ShowArchived: showArchived,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListSubscriptions(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum subscriptions to return, capped at 1000")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	cmd.Flags().BoolVar(&showArchived, "show-archived", false, "Deprecated by Google; subscription archiving is no longer supported")
	return cmd
}

func newSubscriptionsGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var productID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one monetization subscription",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewSubscriptionProductID(productID)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			subscription, err := play.GetSubscription(cmd.Context(), publisher, play.SubscriptionGetOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, subscription)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID")
	return cmd
}

func newSubscriptionsBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var productIDs []string

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple monetization subscriptions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductIDs, err := parseSubscriptionProductIDs(productIDs)
			if err != nil {
				return err
			}
			batchOptions := play.SubscriptionBatchGetOptions{
				PackageName: typedPackageName,
				ProductIDs:  typedProductIDs,
			}
			if err := batchOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchGetSubscriptions(cmd.Context(), publisher, batchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringArrayVar(&productIDs, "product-id", nil, "Subscription product ID; repeatable, up to 100")
	return cmd
}

func parseSubscriptionProductIDs(values []string) ([]play.SubscriptionProductID, error) {
	productIDs := make([]play.SubscriptionProductID, 0, len(values))
	for _, value := range values {
		productID, err := play.NewSubscriptionProductID(value)
		if err != nil {
			return nil, err
		}
		productIDs = append(productIDs, productID)
	}
	return productIDs, nil
}
