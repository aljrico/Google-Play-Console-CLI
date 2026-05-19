package cmd

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newOneTimeProductOffersCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "one-time-product-offers",
		Short: "Inspect Google Play one-time product offers",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newOneTimeProductOffersListCommand(out, options, &packageName),
		newOneTimeProductOffersGetCommand(out, options, &packageName),
		newOneTimeProductOffersBatchGetCommand(out, options, &packageName),
		newOneTimeProductOffersCreateCommand(out, options, &packageName),
		newOneTimeProductOffersBatchDeleteCommand(out, options, &packageName),
		newOneTimeProductOffersBatchPatchAvailabilityCommand(out, options, &packageName),
		newOneTimeProductOffersBatchPatchRelativeDiscountsCommand(out, options, &packageName),
		newOneTimeProductOffersBatchPatchAbsoluteDiscountsCommand(out, options, &packageName),
		newOneTimeProductOffersBatchPatchNoOverridesCommand(out, options, &packageName),
		newOneTimeProductOffersBatchStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionActivate),
		newOneTimeProductOffersBatchStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionDeactivate),
		newOneTimeProductOffersBatchStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionCancel),
		newOneTimeProductOffersStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionActivate),
		newOneTimeProductOffersStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionDeactivate),
		newOneTimeProductOffersStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionCancel),
	)
	return cmd
}

func newOneTimeProductOffersCreateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offerID          string
		fromJSON         string
		offerTags        []string
		startTime        string
		endTime          string
		redemptionLimit  int64
		relativeDiscount []string
		absoluteDiscount []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a one-time product offer",
		Long: "Create a one-time product offer from a Google Play API OneTimeProductOffer JSON body or gpc one-time product offer JSON output. " +
			"Basic flags build one discounted offer with regional relative or absolute discounts; use JSON for no-override regions or pre-order offers. " +
			"Parent IDs come from flags and override the JSON body; output-only state and regionsVersion are ignored.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewOneTimeProductID(productID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(purchaseOptionID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewOneTimeProductOfferID(offerID)
			if err != nil {
				return err
			}
			offer, err := oneTimeProductOfferCreateBody(oneTimeProductOfferCreateBodyOptions{
				FromJSON:         fromJSON,
				OfferTags:        offerTags,
				StartTime:        startTime,
				EndTime:          endTime,
				RedemptionLimit:  redemptionLimit,
				RelativeDiscount: relativeDiscount,
				AbsoluteDiscount: absoluteDiscount,
				BasicFlagsSet: cmd.Flags().Changed("offer-tag") ||
					cmd.Flags().Changed("start-time") ||
					cmd.Flags().Changed("end-time") ||
					cmd.Flags().Changed("redemption-limit") ||
					cmd.Flags().Changed("relative-discount") ||
					cmd.Flags().Changed("absolute-discount"),
			})
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			createOptions := play.OneTimeProductOfferCreateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				OfferID:          typedOfferID,
				Offer:            offer,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.CreateOneTimeProductOffer(cmd.Context(), nil, createOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if err := createOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.CreateOneTimeProductOffer(cmd.Context(), publisher, createOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Parent one-time product ID")
	cmd.Flags().StringVar(&purchaseOptionID, "purchase-option-id", "", "Parent one-time product purchase option ID")
	cmd.Flags().StringVar(&offerID, "offer-id", "", "One-time product offer ID")
	cmd.Flags().StringVar(&fromJSON, "from-json", "", "Path to a Google Play API or gpc JSON one-time product offer body")
	cmd.Flags().StringArrayVar(&offerTags, "offer-tag", nil, "Basic create offer tag; repeatable")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Basic discounted offer start time as RFC3339")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Basic discounted offer end time as RFC3339")
	cmd.Flags().Int64Var(&redemptionLimit, "redemption-limit", 0, "Basic discounted offer redemption limit from 0 to 50")
	cmd.Flags().StringArrayVar(&relativeDiscount, "relative-discount", nil, "Basic create regional relative discount as REGION:0.5, where 0.5 means the user pays 50% of the purchase option price; repeatable")
	cmd.Flags().StringArrayVar(&absoluteDiscount, "absolute-discount", nil, "Basic create regional absolute discount as REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by oneTimeProductOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Create the one-time product offer")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer creation without calling Google Play")
	return cmd
}

type oneTimeProductOfferCreateBodyOptions struct {
	FromJSON         string
	OfferTags        []string
	StartTime        string
	EndTime          string
	RedemptionLimit  int64
	RelativeDiscount []string
	AbsoluteDiscount []string
	BasicFlagsSet    bool
}

func oneTimeProductOfferCreateBody(options oneTimeProductOfferCreateBodyOptions) (play.OneTimeProductOffer, error) {
	if strings.TrimSpace(options.FromJSON) != "" {
		if options.UsesBasicFlags() {
			return play.OneTimeProductOffer{}, fmt.Errorf("--from-json cannot be combined with basic create flags")
		}
		return readOneTimeProductOfferJSON(options.FromJSON)
	}
	if !options.UsesBasicFlags() {
		return play.OneTimeProductOffer{}, fmt.Errorf("one-time product offer create requires --from-json or basic create flags")
	}
	regionalConfigs, err := oneTimeProductOfferCreateRegionalConfigs(options.RelativeDiscount, options.AbsoluteDiscount)
	if err != nil {
		return play.OneTimeProductOffer{}, err
	}
	if err := validateOneTimeProductOfferCreateRFC3339("start time", options.StartTime); err != nil {
		return play.OneTimeProductOffer{}, err
	}
	if err := validateOneTimeProductOfferCreateRFC3339("end time", options.EndTime); err != nil {
		return play.OneTimeProductOffer{}, err
	}
	return play.OneTimeProductOffer{
		Type:      play.OneTimeProductOfferTypeDiscounted,
		OfferTags: append([]string(nil), options.OfferTags...),
		DiscountedOffer: &play.OneTimeProductDiscountedOffer{
			StartTime:       options.StartTime,
			EndTime:         options.EndTime,
			RedemptionLimit: options.RedemptionLimit,
		},
		RegionalConfigs: regionalConfigs,
	}, nil
}

func (o oneTimeProductOfferCreateBodyOptions) UsesBasicFlags() bool {
	return o.BasicFlagsSet
}

func oneTimeProductOfferCreateRegionalConfigs(relativeDiscounts []string, absoluteDiscounts []string) ([]play.OneTimeProductOfferRegion, error) {
	if len(relativeDiscounts) == 0 && len(absoluteDiscounts) == 0 {
		return nil, fmt.Errorf("basic one-time product offer create requires at least one --relative-discount or --absolute-discount")
	}
	regions := make([]play.OneTimeProductOfferRegion, 0, len(relativeDiscounts)+len(absoluteDiscounts))
	if len(relativeDiscounts) > 0 {
		relativeRegions, err := parseOneTimeProductOfferCreateRelativeDiscounts(relativeDiscounts)
		if err != nil {
			return nil, err
		}
		regions = append(regions, relativeRegions...)
	}
	if len(absoluteDiscounts) > 0 {
		absoluteRegions, err := parseOneTimeProductOfferCreateAbsoluteDiscounts(absoluteDiscounts)
		if err != nil {
			return nil, err
		}
		regions = append(regions, absoluteRegions...)
	}
	return regions, nil
}

func parseOneTimeProductOfferCreateRelativeDiscounts(values []string) ([]play.OneTimeProductOfferRegion, error) {
	regions := make([]play.OneTimeProductOfferRegion, 0, len(values))
	for _, value := range values {
		region, rawDiscount, ok := strings.Cut(strings.TrimSpace(value), ":")
		if !ok {
			return nil, errOneTimeProductOfferCreateRelativeDiscountFormat()
		}
		discount, err := strconv.ParseFloat(strings.TrimSpace(rawDiscount), 64)
		if err != nil {
			return nil, errOneTimeProductOfferCreateRelativeDiscountFormat()
		}
		if math.IsNaN(discount) || math.IsInf(discount, 0) || discount <= 0 || discount >= 1 {
			return nil, fmt.Errorf("one-time product offer create relative discount must be greater than 0 and less than 1")
		}
		regions = append(regions, play.OneTimeProductOfferRegion{
			RegionCode:       strings.ToUpper(strings.TrimSpace(region)),
			Availability:     play.OneTimeProductOfferAvailabilityAvailable.String(),
			RelativeDiscount: discount,
		})
	}
	return regions, nil
}

func parseOneTimeProductOfferCreateAbsoluteDiscounts(values []string) ([]play.OneTimeProductOfferRegion, error) {
	regions := make([]play.OneTimeProductOfferRegion, 0, len(values))
	for _, value := range values {
		region, priceValue, ok := strings.Cut(strings.TrimSpace(value), ":")
		if !ok {
			return nil, errOneTimeProductOfferCreateAbsoluteDiscountFormat()
		}
		discount, err := parseRegionalPricePatchMoney(priceValue, errOneTimeProductOfferCreateAbsoluteDiscountFormat)
		if err != nil {
			return nil, errOneTimeProductOfferCreateAbsoluteDiscountFormat()
		}
		regions = append(regions, play.OneTimeProductOfferRegion{
			RegionCode:       strings.ToUpper(strings.TrimSpace(region)),
			Availability:     play.OneTimeProductOfferAvailabilityAvailable.String(),
			AbsoluteDiscount: &discount,
		})
	}
	return regions, nil
}

func validateOneTimeProductOfferCreateRFC3339(fieldName, value string) error {
	if value == "" {
		return nil
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("one-time product offer create %s cannot have leading or trailing whitespace", fieldName)
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return fmt.Errorf("one-time product offer create %s must be RFC3339: %w", fieldName, err)
	}
	return nil
}

func errOneTimeProductOfferCreateRelativeDiscountFormat() error {
	return fmt.Errorf("one-time product offer create relative discount must use REGION:0.5")
}

func errOneTimeProductOfferCreateAbsoluteDiscountFormat() error {
	return fmt.Errorf("one-time product offer create absolute discount must use REGION:CURRENCY:UNITS[:NANOS]")
}

func readOneTimeProductOfferJSON(path string) (play.OneTimeProductOffer, error) {
	if strings.TrimSpace(path) == "" {
		return play.OneTimeProductOffer{}, fmt.Errorf("one-time product offer create requires --from-json")
	}
	data, err := osReadFile(path)
	if err != nil {
		return play.OneTimeProductOffer{}, fmt.Errorf("read one-time product offer JSON %s: %w", path, err)
	}
	offer, err := play.DecodeOneTimeProductOfferCreateJSON(data)
	if err != nil {
		return play.OneTimeProductOffer{}, fmt.Errorf("parse one-time product offer JSON %s: %w", path, err)
	}
	return offer, nil
}

func newOneTimeProductOffersBatchPatchAvailabilityCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		availability     []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-availability",
		Short: "Batch patch one-time product offer regional availability",
		Long: "Batch patch one-time product offer regional availability. Omit parent IDs to infer the narrowest valid parent path from --availability values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferAvailabilityPatches(availability)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.OneTimeProductOfferBatchPatchAvailabilityOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := oneTimeProductOfferAvailabilityPatchesToMutationRequests(requests)
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, mutationRequests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.OneTimeProductOfferBatchPatchAvailabilityOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchOneTimeProductOfferAvailability(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewOneTimeProductOfferBatchPatchAvailabilityPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchOneTimeProductOfferAvailability(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&availability, "availability", nil, "Availability patch as productId/purchaseOptionId/offerId/REGION:available|noLongerAvailable; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by oneTimeProductOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer availability batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer availability batch patch without calling Google Play")
	return cmd
}

func newOneTimeProductOffersBatchPatchRelativeDiscountsCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		relativeDiscount []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-relative-discounts",
		Short: "Batch patch one-time product offer relative discounts",
		Long: "Batch patch one-time product offer relative discounts. Omit parent IDs to infer the narrowest valid parent path from --relative-discount values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferRelativeDiscountPatches(relativeDiscount)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.OneTimeProductOfferBatchPatchRelativeDiscountsOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := oneTimeProductOfferRelativeDiscountPatchesToMutationRequests(requests)
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, mutationRequests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.OneTimeProductOfferBatchPatchRelativeDiscountsOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchOneTimeProductOfferRelativeDiscounts(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewOneTimeProductOfferBatchPatchRelativeDiscountsPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchOneTimeProductOfferRelativeDiscounts(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&relativeDiscount, "relative-discount", nil, "Relative discount patch as productId/purchaseOptionId/offerId/REGION:0.75, where 0.75 means the user pays 75% of the purchase option price; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by oneTimeProductOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer relative discount batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer relative discount batch patch without calling Google Play")
	return cmd
}

func newOneTimeProductOffersBatchPatchAbsoluteDiscountsCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		absoluteDiscount []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-absolute-discounts",
		Short: "Batch patch one-time product offer absolute discounts",
		Long: "Batch patch one-time product offer absolute discounts. Omit parent IDs to infer the narrowest valid parent path from --absolute-discount values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferAbsoluteDiscountPatches(absoluteDiscount)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := oneTimeProductOfferAbsoluteDiscountPatchesToMutationRequests(requests)
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, mutationRequests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchOneTimeProductOfferAbsoluteDiscounts(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewOneTimeProductOfferBatchPatchAbsoluteDiscountsPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchOneTimeProductOfferAbsoluteDiscounts(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&absoluteDiscount, "absolute-discount", nil, "Absolute discount patch as productId/purchaseOptionId/offerId/REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by oneTimeProductOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer absolute discount batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer absolute discount batch patch without calling Google Play")
	return cmd
}

func newOneTimeProductOffersBatchPatchNoOverridesCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		noOverride       []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-no-overrides",
		Short: "Batch reset one-time product offer regional discounts to no override",
		Long: "Batch reset one-time product offer regional discounts to no override. Omit parent IDs to infer the narrowest valid parent path from --no-override values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferNoOverridePatches(noOverride)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.OneTimeProductOfferBatchPatchNoOverridesOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := oneTimeProductOfferNoOverridePatchesToMutationRequests(requests)
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, mutationRequests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.OneTimeProductOfferBatchPatchNoOverridesOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchOneTimeProductOfferNoOverrides(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewOneTimeProductOfferBatchPatchNoOverridesPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchOneTimeProductOfferNoOverrides(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&noOverride, "no-override", nil, "No-override patch as productId/purchaseOptionId/offerId/REGION; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by oneTimeProductOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer no-override batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer no-override batch patch without calling Google Play")
	return cmd
}

func newOneTimeProductOffersBatchStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.OneTimeProductOfferStateAction) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offers           []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-" + action.String(),
		Short: oneTimeProductOfferBatchStateShort(action),
		Long: oneTimeProductOfferBatchStateLong(action) + " Omit parent IDs to infer the narrowest valid parent path from --offer values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferBatchMutationRequests(offers)
			if err != nil {
				return err
			}
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, requests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.OneTimeProductOfferBatchStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := updateOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.BatchUpdateOneTimeProductOfferStates(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchUpdateOneTimeProductOfferStates(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, oneTimeProductOfferBatchStateConfirmHelp(action))
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, oneTimeProductOfferBatchStateDryRunHelp(action))
	return cmd
}

func oneTimeProductOfferBatchStateShort(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Cancel multiple pre-order one-time product offers and pending orders"
	}
	return action.String() + " multiple one-time product offers"
}

func oneTimeProductOfferBatchStateLong(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Cancel multiple pre-order one-time product offers. Google Play cancels pending orders for the cancelled offers."
	}
	return string(action) + " multiple one-time product offers."
}

func oneTimeProductOfferBatchStateConfirmHelp(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Cancel the pre-order offers and their pending orders"
	}
	return "Apply the one-time product offer batch state update"
}

func oneTimeProductOfferBatchStateDryRunHelp(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Print the planned pre-order offer cancellation without calling Google Play"
	}
	return "Print the planned one-time product offer batch state update without calling Google Play"
}

func newOneTimeProductOffersBatchDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offers           []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Delete multiple one-time product offers",
		Long: "Delete multiple one-time product offers. Omit parent IDs to infer the narrowest valid parent path from --offer values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferBatchMutationRequests(offers)
			if err != nil {
				return err
			}
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, requests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			deleteOptions := play.OneTimeProductOfferBatchDeleteOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.BatchDeleteOneTimeProductOffers(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchDeleteOneTimeProductOffers(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to delete as productId/purchaseOptionId/offerId; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer batch deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer batch deletion without calling Google Play")
	return cmd
}

func newOneTimeProductOffersBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offers           []string
	)

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple one-time product offers",
		Long:  "Get multiple one-time product offers. Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options. Concrete parent IDs must match every --offer value.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedPurchaseOptionID, err := parseOneTimeProductOfferListParent(*packageName, productID, purchaseOptionID)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferBatchRequests(offers)
			if err != nil {
				return err
			}
			batchOptions := play.OneTimeProductOfferBatchGetOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
			}
			if err := batchOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchGetOneTimeProductOffers(cmd.Context(), publisher, batchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for offers across products",
		"Parent one-time product purchase option ID, or - for offers across purchase options",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to fetch as productId/purchaseOptionId/offerId; repeatable, up to 100")
	return cmd
}

func newOneTimeProductOffersStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.OneTimeProductOfferStateAction) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offerID          string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action.String(),
		Short: string(action) + " a one-time product offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedPurchaseOptionID, err := parseOneTimeProductOfferGetParent(*packageName, productID, purchaseOptionID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewOneTimeProductOfferID(offerID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.OneTimeProductOfferStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				OfferID:          typedOfferID,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.UpdateOneTimeProductOfferState(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewOneTimeProductOfferStateUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateOneTimeProductOfferState(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(cmd, &productID, &purchaseOptionID, "Parent one-time product ID", "Parent one-time product purchase option ID")
	cmd.Flags().StringVar(&offerID, "offer-id", "", "One-time product offer ID")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer state update without calling Google Play")
	return cmd
}

func newOneTimeProductOffersListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		pageSize         int64
		pageToken        string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List one-time product offers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedPurchaseOptionID, err := parseOneTimeProductOfferListParent(*packageName, productID, purchaseOptionID)
			if err != nil {
				return err
			}
			listOptions := play.OneTimeProductOfferListOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				PageSize:         pageSize,
				PageToken:        pageToken,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListOneTimeProductOffers(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID, or - for all products",
		"Parent one-time product purchase option ID, or - for all purchase options",
	)
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum offers to return, capped at 1000")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	return cmd
}

func newOneTimeProductOffersGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offerID          string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a one-time product offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedPurchaseOptionID, err := parseOneTimeProductOfferGetParent(*packageName, productID, purchaseOptionID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewOneTimeProductOfferID(offerID)
			if err != nil {
				return err
			}
			getOptions := play.OneTimeProductOfferGetOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				OfferID:          typedOfferID,
			}
			if err := getOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			offer, err := play.GetOneTimeProductOffer(cmd.Context(), publisher, getOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, offer)
		},
	}
	addOneTimeProductOfferParentFlags(
		cmd,
		&productID,
		&purchaseOptionID,
		"Parent one-time product ID",
		"Parent one-time product purchase option ID",
	)
	cmd.Flags().StringVar(&offerID, "offer-id", "", "One-time product offer ID")
	return cmd
}

func addOneTimeProductOfferParentFlags(cmd *cobra.Command, productID *string, purchaseOptionID *string, productIDDescription string, purchaseOptionIDDescription string) {
	cmd.Flags().StringVar(productID, "product-id", "", productIDDescription)
	cmd.Flags().StringVar(purchaseOptionID, "purchase-option-id", "", purchaseOptionIDDescription)
}

func parseOneTimeProductOfferListParent(packageName string, productID string, purchaseOptionID string) (play.PackageName, play.OneTimeProductID, play.OneTimeProductPurchaseOptionID, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return "", "", "", err
	}
	typedProductID, err := play.NewOneTimeProductOfferListProductID(productID)
	if err != nil {
		return "", "", "", err
	}
	typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(purchaseOptionID)
	if err != nil {
		return "", "", "", err
	}
	return typedPackageName, typedProductID, typedPurchaseOptionID, nil
}

func parseOneTimeProductOfferGetParent(packageName string, productID string, purchaseOptionID string) (play.PackageName, play.OneTimeProductID, play.OneTimeProductPurchaseOptionID, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return "", "", "", err
	}
	typedProductID, err := play.NewOneTimeProductID(productID)
	if err != nil {
		return "", "", "", err
	}
	typedPurchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(purchaseOptionID)
	if err != nil {
		return "", "", "", err
	}
	return typedPackageName, typedProductID, typedPurchaseOptionID, nil
}

func parseOneTimeProductOfferBatchRequests(values []string) ([]play.OneTimeProductOfferBatchGetRequest, error) {
	requests := make([]play.OneTimeProductOfferBatchGetRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewOneTimeProductOfferBatchGetRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferBatchMutationRequests(values []string) ([]play.OneTimeProductOfferBatchMutationRequest, error) {
	requests := make([]play.OneTimeProductOfferBatchMutationRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewOneTimeProductOfferBatchMutationRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferAvailabilityPatches(values []string) ([]play.OneTimeProductOfferAvailabilityPatchRequest, error) {
	requests := make([]play.OneTimeProductOfferAvailabilityPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseOneTimeProductOfferAvailabilityPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferAvailabilityPatch(value string) (play.OneTimeProductOfferAvailabilityPatchRequest, error) {
	path, rawAvailability, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, errOneTimeProductOfferAvailabilityFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, errOneTimeProductOfferAvailabilityFormat()
	}
	productID, err := play.NewOneTimeProductID(parts[0])
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	purchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(parts[1])
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	offerID, err := play.NewOneTimeProductOfferID(parts[2])
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	availability, err := parseOneTimeProductOfferAvailabilityValue(rawAvailability)
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	return play.OneTimeProductOfferAvailabilityPatchRequest{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
		RegionCode:       strings.ToUpper(strings.TrimSpace(parts[3])),
		Availability:     availability,
	}, nil
}

func parseOneTimeProductOfferAvailabilityValue(value string) (play.OneTimeProductOfferAvailability, error) {
	switch strings.TrimSpace(value) {
	case play.OneTimeProductOfferAvailabilityAvailable.String():
		return play.OneTimeProductOfferAvailabilityAvailable, nil
	case play.OneTimeProductOfferAvailabilityNoLongerAvailable.String():
		return play.OneTimeProductOfferAvailabilityNoLongerAvailable, nil
	default:
		return "", errOneTimeProductOfferAvailabilityFormat()
	}
}

func parseOneTimeProductOfferRelativeDiscountPatches(values []string) ([]play.OneTimeProductOfferRelativeDiscountPatchRequest, error) {
	requests := make([]play.OneTimeProductOfferRelativeDiscountPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseOneTimeProductOfferRelativeDiscountPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferRelativeDiscountPatch(value string) (play.OneTimeProductOfferRelativeDiscountPatchRequest, error) {
	path, rawRelativeDiscount, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.OneTimeProductOfferRelativeDiscountPatchRequest{}, errOneTimeProductOfferRelativeDiscountFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return play.OneTimeProductOfferRelativeDiscountPatchRequest{}, errOneTimeProductOfferRelativeDiscountFormat()
	}
	productID, err := play.NewOneTimeProductID(parts[0])
	if err != nil {
		return play.OneTimeProductOfferRelativeDiscountPatchRequest{}, err
	}
	purchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(parts[1])
	if err != nil {
		return play.OneTimeProductOfferRelativeDiscountPatchRequest{}, err
	}
	offerID, err := play.NewOneTimeProductOfferID(parts[2])
	if err != nil {
		return play.OneTimeProductOfferRelativeDiscountPatchRequest{}, err
	}
	relativeDiscount, err := strconv.ParseFloat(strings.TrimSpace(rawRelativeDiscount), 64)
	if err != nil {
		return play.OneTimeProductOfferRelativeDiscountPatchRequest{}, errOneTimeProductOfferRelativeDiscountFormat()
	}
	return play.OneTimeProductOfferRelativeDiscountPatchRequest{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
		RegionCode:       strings.ToUpper(strings.TrimSpace(parts[3])),
		RelativeDiscount: relativeDiscount,
	}, nil
}

func parseOneTimeProductOfferAbsoluteDiscountPatches(values []string) ([]play.OneTimeProductOfferAbsoluteDiscountPatchRequest, error) {
	requests := make([]play.OneTimeProductOfferAbsoluteDiscountPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseOneTimeProductOfferAbsoluteDiscountPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferAbsoluteDiscountPatch(value string) (play.OneTimeProductOfferAbsoluteDiscountPatchRequest, error) {
	path, rawAbsoluteDiscount, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.OneTimeProductOfferAbsoluteDiscountPatchRequest{}, errOneTimeProductOfferAbsoluteDiscountFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return play.OneTimeProductOfferAbsoluteDiscountPatchRequest{}, errOneTimeProductOfferAbsoluteDiscountFormat()
	}
	productID, err := play.NewOneTimeProductID(parts[0])
	if err != nil {
		return play.OneTimeProductOfferAbsoluteDiscountPatchRequest{}, err
	}
	purchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(parts[1])
	if err != nil {
		return play.OneTimeProductOfferAbsoluteDiscountPatchRequest{}, err
	}
	offerID, err := play.NewOneTimeProductOfferID(parts[2])
	if err != nil {
		return play.OneTimeProductOfferAbsoluteDiscountPatchRequest{}, err
	}
	absoluteDiscount, err := parsePurchaseOptionPatchMoney(rawAbsoluteDiscount)
	if err != nil {
		return play.OneTimeProductOfferAbsoluteDiscountPatchRequest{}, errOneTimeProductOfferAbsoluteDiscountFormat()
	}
	return play.OneTimeProductOfferAbsoluteDiscountPatchRequest{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
		RegionCode:       strings.ToUpper(strings.TrimSpace(parts[3])),
		AbsoluteDiscount: absoluteDiscount,
	}, nil
}

func parseOneTimeProductOfferNoOverridePatches(values []string) ([]play.OneTimeProductOfferNoOverridePatchRequest, error) {
	requests := make([]play.OneTimeProductOfferNoOverridePatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseOneTimeProductOfferNoOverridePatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferNoOverridePatch(value string) (play.OneTimeProductOfferNoOverridePatchRequest, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 4 {
		return play.OneTimeProductOfferNoOverridePatchRequest{}, errOneTimeProductOfferNoOverrideFormat()
	}
	productID, err := play.NewOneTimeProductID(parts[0])
	if err != nil {
		return play.OneTimeProductOfferNoOverridePatchRequest{}, err
	}
	purchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(parts[1])
	if err != nil {
		return play.OneTimeProductOfferNoOverridePatchRequest{}, err
	}
	offerID, err := play.NewOneTimeProductOfferID(parts[2])
	if err != nil {
		return play.OneTimeProductOfferNoOverridePatchRequest{}, err
	}
	return play.OneTimeProductOfferNoOverridePatchRequest{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
		RegionCode:       strings.ToUpper(strings.TrimSpace(parts[3])),
		NoOverride:       true,
	}, nil
}

func errOneTimeProductOfferAvailabilityFormat() error {
	return fmt.Errorf("one-time product offer availability must use productId/purchaseOptionId/offerId/REGION:available|noLongerAvailable")
}

func errOneTimeProductOfferRelativeDiscountFormat() error {
	return fmt.Errorf("one-time product offer relative discount must use productId/purchaseOptionId/offerId/REGION:0.5")
}

func errOneTimeProductOfferAbsoluteDiscountFormat() error {
	return fmt.Errorf("one-time product offer absolute discount must use productId/purchaseOptionId/offerId/REGION:CURRENCY:UNITS[:NANOS]")
}

func errOneTimeProductOfferNoOverrideFormat() error {
	return fmt.Errorf("one-time product offer no-override must use productId/purchaseOptionId/offerId/REGION")
}

func oneTimeProductOfferAvailabilityPatchesToMutationRequests(requests []play.OneTimeProductOfferAvailabilityPatchRequest) []play.OneTimeProductOfferBatchMutationRequest {
	mutations := make([]play.OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return mutations
}

func oneTimeProductOfferRelativeDiscountPatchesToMutationRequests(requests []play.OneTimeProductOfferRelativeDiscountPatchRequest) []play.OneTimeProductOfferBatchMutationRequest {
	mutations := make([]play.OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return mutations
}

func oneTimeProductOfferAbsoluteDiscountPatchesToMutationRequests(requests []play.OneTimeProductOfferAbsoluteDiscountPatchRequest) []play.OneTimeProductOfferBatchMutationRequest {
	mutations := make([]play.OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return mutations
}

func oneTimeProductOfferNoOverridePatchesToMutationRequests(requests []play.OneTimeProductOfferNoOverridePatchRequest) []play.OneTimeProductOfferBatchMutationRequest {
	mutations := make([]play.OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return mutations
}

func inferOneTimeProductOfferBatchParent(productID string, purchaseOptionID string, requests []play.OneTimeProductOfferBatchMutationRequest) (string, string) {
	if len(requests) == 0 {
		return productID, purchaseOptionID
	}
	firstProductID := requests[0].ProductID.String()
	firstPurchaseOptionID := requests[0].PurchaseOptionID.String()
	inferredProductID := firstProductID
	inferredPurchaseOptionID := firstPurchaseOptionID
	for _, request := range requests[1:] {
		if request.ProductID.String() != firstProductID {
			inferredProductID = play.OneTimeProductOfferWildcardID
			inferredPurchaseOptionID = play.OneTimeProductOfferWildcardID
			break
		}
		if request.PurchaseOptionID.String() != firstPurchaseOptionID {
			inferredPurchaseOptionID = play.OneTimeProductOfferWildcardID
		}
	}
	if productID != "" {
		inferredProductID = productID
	}
	if purchaseOptionID != "" {
		inferredPurchaseOptionID = purchaseOptionID
	}
	return inferredProductID, inferredPurchaseOptionID
}
