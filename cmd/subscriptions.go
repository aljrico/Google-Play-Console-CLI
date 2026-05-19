package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

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
		newSubscriptionsBatchPatchListingsCommand(out, options, &packageName),
		newSubscriptionsDeleteCommand(out, options, &packageName),
		newSubscriptionsBasePlanCommand(out, options, &packageName),
	)
	return cmd
}

func newSubscriptionsBatchPatchListingsCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		listings         []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-listings",
		Short: "Batch patch localized subscription listings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionBatchListingPatches(listings)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionBatchPatchListingsOptions{
				PackageName:      typedPackageName,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchSubscriptionListings(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionBatchPatchListingsPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchSubscriptionListings(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringArrayVar(&listings, "listing", nil, "CSV listing patch productId,language,title,description; repeat for multiple localized listings")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptions.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription listing batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription listing batch patch without calling Google Play")
	return cmd
}

func parseSubscriptionBatchListingPatches(values []string) ([]play.SubscriptionBatchPatchListingRequest, error) {
	requests := make([]play.SubscriptionBatchPatchListingRequest, 0, len(values))
	for _, value := range values {
		request, err := parseSubscriptionBatchListingPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionBatchListingPatch(value string) (play.SubscriptionBatchPatchListingRequest, error) {
	reader := csv.NewReader(strings.NewReader(value))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return play.SubscriptionBatchPatchListingRequest{}, fmt.Errorf("parse subscription listing CSV: %w", err)
	}
	if len(records) != 1 {
		return play.SubscriptionBatchPatchListingRequest{}, fmt.Errorf("subscription listing must contain exactly one CSV record")
	}
	fields := records[0]
	if len(fields) != 4 {
		return play.SubscriptionBatchPatchListingRequest{}, fmt.Errorf("subscription listing must be CSV productId,language,title,description")
	}
	productID, err := play.NewSubscriptionProductID(fields[0])
	if err != nil {
		return play.SubscriptionBatchPatchListingRequest{}, err
	}
	language, err := play.NewListingLanguage(fields[1])
	if err != nil {
		return play.SubscriptionBatchPatchListingRequest{}, err
	}
	return play.SubscriptionBatchPatchListingRequest{
		ProductID: productID,
		Listing: play.SubscriptionListing{
			LanguageCode: language.String(),
			Title:        fields[2],
			Description:  fields[3],
		},
	}, nil
}

func newSubscriptionsBasePlanCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base-plan",
		Short: "Manage subscription base plans",
	}
	cmd.AddCommand(
		newSubscriptionsBasePlanDeleteCommand(out, options, packageName),
		newSubscriptionsBasePlanStateCommand(out, options, packageName, play.BasePlanStateActionActivate),
		newSubscriptionsBasePlanStateCommand(out, options, packageName, play.BasePlanStateActionDeactivate),
		newSubscriptionsBasePlanBatchStateCommand(out, options, packageName, play.BasePlanStateActionActivate),
		newSubscriptionsBasePlanBatchStateCommand(out, options, packageName, play.BasePlanStateActionDeactivate),
		newSubscriptionsBasePlanBatchMigratePricesCommand(out, options, packageName),
	)
	return cmd
}

func newSubscriptionsBasePlanDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		confirm    bool
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a draft-only subscription base plan",
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
			deleteOptions := play.BasePlanDeleteOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if dryRun {
				result, err := play.DeleteBasePlan(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteBasePlan(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID")
	cmd.Flags().StringVar(&basePlanID, "base-plan-id", "", "Subscription base plan ID")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the base plan deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned base plan deletion without calling Google Play")
	return cmd
}

func newSubscriptionsBasePlanBatchMigratePricesCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID         string
		regionsVersion    string
		migrations        []string
		priceIncreaseType string
		latencyTolerance  string
		confirm           bool
		dryRun            bool
	)

	cmd := &cobra.Command{
		Use:   "batch-migrate-prices",
		Short: "Batch migrate subscription base plan prices",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			typedPriceIncreaseType := play.BasePlanPriceIncreaseType(priceIncreaseType)
			if err := typedPriceIncreaseType.Validate(); err != nil {
				return err
			}
			resolvedProductID, requests, err := parseBasePlanPriceMigrationRequests(productID, migrations, typedPriceIncreaseType)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewSubscriptionBasePlanBatchProductID(resolvedProductID)
			if err != nil {
				return err
			}
			migrationOptions := play.BasePlanBatchPriceMigrationOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				RegionsVersion:   regionsVersion,
				Requests:         requests,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchMigrateBasePlanPrices(cmd.Context(), nil, migrationOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewBasePlanBatchPriceMigrationPlan(migrationOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchMigrateBasePlanPrices(cmd.Context(), publisher, migrationOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID, or - for migrations across subscriptions; inferred from --migration values")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by batchMigratePrices")
	cmd.Flags().StringArrayVar(&migrations, "migration", nil, "Price migration as productId/basePlanId/REGION/RFC3339_TIME; repeat for multiple regions or base plans")
	cmd.Flags().StringVar(&priceIncreaseType, "price-increase-type", "", "Price increase type: optIn or optOut")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the base plan price migration")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned base plan price migration without calling Google Play")
	return cmd
}

func parseBasePlanPriceMigrationRequests(productID string, values []string, priceIncreaseType play.BasePlanPriceIncreaseType) (string, []play.BasePlanPriceMigrationRequest, error) {
	grouped := map[string]int{}
	requests := make([]play.BasePlanPriceMigrationRequest, 0, len(values))
	for _, value := range values {
		request, region, err := parseBasePlanPriceMigration(value, priceIncreaseType)
		if err != nil {
			return "", nil, err
		}
		key := request.ProductID.String() + "/" + request.BasePlanID.String()
		index, ok := grouped[key]
		if !ok {
			grouped[key] = len(requests)
			request.Regions = []play.BasePlanPriceMigrationConfig{region}
			requests = append(requests, request)
			continue
		}
		requests[index].Regions = append(requests[index].Regions, region)
	}
	return inferBasePlanPriceMigrationProductID(productID, requests), requests, nil
}

func parseBasePlanPriceMigration(value string, priceIncreaseType play.BasePlanPriceIncreaseType) (play.BasePlanPriceMigrationRequest, play.BasePlanPriceMigrationConfig, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 4 {
		return play.BasePlanPriceMigrationRequest{}, play.BasePlanPriceMigrationConfig{}, fmt.Errorf("price migration must use productId/basePlanId/REGION/RFC3339_TIME")
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.BasePlanPriceMigrationRequest{}, play.BasePlanPriceMigrationConfig{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.BasePlanPriceMigrationRequest{}, play.BasePlanPriceMigrationConfig{}, err
	}
	region := play.BasePlanPriceMigrationConfig{
		RegionCode:                    strings.ToUpper(parts[2]),
		OldestAllowedPriceVersionTime: parts[3],
		PriceIncreaseType:             priceIncreaseType,
	}
	return play.BasePlanPriceMigrationRequest{ProductID: productID, BasePlanID: basePlanID}, region, nil
}

func inferBasePlanPriceMigrationProductID(productID string, requests []play.BasePlanPriceMigrationRequest) string {
	if productID != "" {
		return productID
	}
	if len(requests) == 0 {
		return productID
	}
	firstProductID := requests[0].ProductID.String()
	for _, request := range requests[1:] {
		if request.ProductID.String() != firstProductID {
			return play.SubscriptionOfferWildcardID
		}
	}
	return firstProductID
}

func newSubscriptionsBasePlanBatchStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.BasePlanStateAction) *cobra.Command {
	var (
		productID        string
		basePlanIDs      []string
		basePlans        []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-" + action.String(),
		Short: "Batch " + string(action) + " subscription base plans",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			resolvedProductID, requests, err := parseBasePlanBatchStateUpdateRequests(productID, basePlanIDs, basePlans)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewSubscriptionBasePlanBatchProductID(resolvedProductID)
			if err != nil {
				return err
			}
			updateOptions := play.BasePlanBatchStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				Requests:         requests,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchUpdateBasePlanStates(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewBasePlanBatchStateUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchUpdateBasePlanStates(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID, or - for base plans across subscriptions; inferred when --base-plan is used")
	cmd.Flags().StringArrayVar(&basePlanIDs, "base-plan-id", nil, "Subscription base plan ID; repeat for multiple base plans")
	cmd.Flags().StringArrayVar(&basePlans, "base-plan", nil, "Subscription base plan as productId/basePlanId; repeat for cross-subscription batches")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the base plan batch state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned base plan batch state update without calling Google Play")
	return cmd
}

func parseBasePlanBatchStateUpdateRequests(productID string, basePlanIDs []string, basePlans []string) (string, []play.BasePlanBatchStateUpdateRequest, error) {
	requests := make([]play.BasePlanBatchStateUpdateRequest, 0, len(basePlanIDs)+len(basePlans))
	if len(basePlanIDs) > 0 {
		if productID == "" || productID == play.SubscriptionOfferWildcardID {
			return "", nil, errBasePlanIDRequiresConcreteProduct()
		}
		typedProductID, err := play.NewSubscriptionProductID(productID)
		if err != nil {
			return "", nil, err
		}
		for _, basePlanID := range basePlanIDs {
			typedBasePlanID, err := play.NewSubscriptionBasePlanID(basePlanID)
			if err != nil {
				return "", nil, err
			}
			requests = append(requests, play.BasePlanBatchStateUpdateRequest{ProductID: typedProductID, BasePlanID: typedBasePlanID})
		}
	}
	for _, basePlan := range basePlans {
		request, err := parseBasePlanBatchStateUpdateRequest(basePlan)
		if err != nil {
			return "", nil, err
		}
		requests = append(requests, request)
	}
	return inferBasePlanBatchStateUpdateProductID(productID, requests), requests, nil
}

func parseBasePlanBatchStateUpdateRequest(value string) (play.BasePlanBatchStateUpdateRequest, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return play.BasePlanBatchStateUpdateRequest{}, fmt.Errorf("subscription base plan must use productId/basePlanId")
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.BasePlanBatchStateUpdateRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.BasePlanBatchStateUpdateRequest{}, err
	}
	return play.BasePlanBatchStateUpdateRequest{ProductID: productID, BasePlanID: basePlanID}, nil
}

func inferBasePlanBatchStateUpdateProductID(productID string, requests []play.BasePlanBatchStateUpdateRequest) string {
	if productID != "" {
		return productID
	}
	if len(requests) == 0 {
		return productID
	}
	firstProductID := requests[0].ProductID.String()
	for _, request := range requests[1:] {
		if request.ProductID.String() != firstProductID {
			return play.SubscriptionOfferWildcardID
		}
	}
	return firstProductID
}

func errBasePlanIDRequiresConcreteProduct() error {
	return fmt.Errorf("--base-plan-id requires a concrete --product-id; use --base-plan productId/basePlanId for cross-subscription batches")
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
		regionsVersion   string
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
				DescriptionSet:   cmd.Flags().Changed("description"),
				BenefitsSet:      cmd.Flags().Changed("benefit"),
				RegionsVersion:   regionsVersion,
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
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptions.patch")
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
