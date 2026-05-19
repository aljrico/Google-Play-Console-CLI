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
		newSubscriptionsPatchCommand(out, options, &packageName),
		newSubscriptionsDeleteCommand(out, options, &packageName),
		newSubscriptionsBasePlanCommand(out, options, &packageName),
	)
	return cmd
}

func newSubscriptionsBasePlanCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base-plan",
		Short: "Manage subscription base plans",
	}
	cmd.AddCommand(
		newSubscriptionsBasePlanStateCommand(out, options, packageName, play.BasePlanStateActionActivate),
		newSubscriptionsBasePlanStateCommand(out, options, packageName, play.BasePlanStateActionDeactivate),
	)
	return cmd
}

func newSubscriptionsBasePlanStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.BasePlanStateAction) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action.String(),
		Short: string(action) + " a subscription base plan",
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
			typedBasePlanID, err := play.NewSubscriptionBasePlanID(basePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.BasePlanStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.UpdateBasePlanState(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewBasePlanStateUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateBasePlanState(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID")
	cmd.Flags().StringVar(&basePlanID, "base-plan-id", "", "Subscription base plan ID")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the base plan state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned base plan state update without calling Google Play")
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

func newSubscriptionsDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID string
		confirm   bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a draft-only monetization subscription",
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
			deleteOptions := play.SubscriptionDeleteOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DeleteSubscription(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteSubscription(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription deletion without calling Google Play")
	return cmd
}

func newSubscriptionsPatchCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		listingLanguage  string
		title            string
		description      string
		benefits         []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Patch a subscription listing",
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
			typedListingLanguage, err := play.NewListingLanguage(listingLanguage)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionPatchOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				Listing: play.SubscriptionListing{
					LanguageCode: typedListingLanguage.String(),
					Title:        title,
					Description:  description,
					Benefits:     benefits,
				},
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.PatchSubscription(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionPatchPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.PatchSubscription(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID")
	cmd.Flags().StringVar(&listingLanguage, "listing-language", "", "BCP-47 language code for the listing to patch, for example en-US")
	cmd.Flags().StringVar(&title, "title", "", "Localized subscription title")
	cmd.Flags().StringVar(&description, "description", "", "Localized subscription description")
	cmd.Flags().StringArrayVar(&benefits, "benefit", nil, "Localized subscription benefit; repeatable, up to 4")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription listing patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription listing patch without calling Google Play")
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
