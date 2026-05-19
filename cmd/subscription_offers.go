package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newSubscriptionOffersCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "subscription-offers",
		Short: "Inspect Google Play subscription offers",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newSubscriptionOffersListCommand(out, options, &packageName),
		newSubscriptionOffersGetCommand(out, options, &packageName),
		newSubscriptionOffersCreateCommand(out, options, &packageName),
		newSubscriptionOffersBatchGetCommand(out, options, &packageName),
		newSubscriptionOffersDeleteCommand(out, options, &packageName),
		newSubscriptionOffersBatchPatchAvailabilityCommand(out, options, &packageName),
		newSubscriptionOffersBatchPatchPhaseRelativeDiscountsCommand(out, options, &packageName),
		newSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsCommand(out, options, &packageName),
		newSubscriptionOffersBatchPatchPhasePricesCommand(out, options, &packageName),
		newSubscriptionOffersBatchPatchPhaseFreeCommand(out, options, &packageName),
		newSubscriptionOffersBatchStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionActivate),
		newSubscriptionOffersBatchStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionDeactivate),
		newSubscriptionOffersStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionActivate),
		newSubscriptionOffersStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionDeactivate),
	)
	return cmd
}

func newSubscriptionOffersBatchPatchAvailabilityCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		availability     []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-availability",
		Short: "Batch patch subscription offer regional availability",
		Long: "Batch patch subscription offer regional availability. Omit parent IDs to infer the narrowest valid parent path from --availability values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferAvailabilityPatches(availability)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchPatchAvailabilityOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := subscriptionOfferAvailabilityPatchesToMutationRequests(requests)
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, mutationRequests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionOfferBatchPatchAvailabilityOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchSubscriptionOfferAvailability(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferBatchPatchAvailabilityPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchSubscriptionOfferAvailability(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&availability, "availability", nil, "Availability patch as productId/basePlanId/offerId/REGION:true|false; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptionOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer availability batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer availability batch patch without calling Google Play")
	return cmd
}

func newSubscriptionOffersCreateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID         string
		basePlanID        string
		offerID           string
		fromJSON          string
		offerTags         []string
		freeRegions       []string
		prices            []string
		relativeDiscounts []string
		absoluteDiscounts []string
		phase2FreeRegions []string
		phase2Prices      []string
		phase2RelDiscount []string
		phase2AbsDiscount []string
		otherRegionsFree  bool
		otherRegionsUSD   string
		otherRegionsEUR   string
		phase2OtherUSD    string
		phase2OtherEUR    string
		acquisitionScope  string
		upgradeScope      string
		upgradeProductID  string
		upgradePeriod     string
		upgradeOnce       bool
		phaseDuration     string
		phaseRecurrence   int64
		phase2Duration    string
		phase2Recurrence  int64
		regionsVersion    string
		confirm           bool
		dryRun            bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a draft subscription offer",
		Long: "Create a draft subscription offer from a Google Play API SubscriptionOffer JSON body or gpc subscription offer JSON output. " +
			"Basic flags build one or two free, paid-price, relative-discount, or absolute-discount phases across explicit regions, with optional free or paid-price other-regions config and acquisition or upgrade targeting; use JSON for other-regions relative or absolute discounts. " +
			"Immutable parent IDs come from flags and override the JSON body; output-only state is ignored because Google creates draft offers.",
		Args: cobra.NoArgs,
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
			typedOfferID, err := play.NewSubscriptionOfferID(offerID)
			if err != nil {
				return err
			}
			offer, err := subscriptionOfferCreateBody(subscriptionOfferCreateBodyOptions{
				FromJSON:          fromJSON,
				OfferTags:         offerTags,
				FreeRegions:       freeRegions,
				Prices:            prices,
				RelativeDiscounts: relativeDiscounts,
				AbsoluteDiscounts: absoluteDiscounts,
				Phase2FreeRegions: phase2FreeRegions,
				Phase2Prices:      phase2Prices,
				Phase2RelDiscount: phase2RelDiscount,
				Phase2AbsDiscount: phase2AbsDiscount,
				OtherRegionsFree:  otherRegionsFree,
				OtherRegionsUSD:   otherRegionsUSD,
				OtherRegionsEUR:   otherRegionsEUR,
				Phase2OtherUSD:    phase2OtherUSD,
				Phase2OtherEUR:    phase2OtherEUR,
				AcquisitionScope:  acquisitionScope,
				UpgradeScope:      upgradeScope,
				UpgradeProductID:  upgradeProductID,
				UpgradePeriod:     upgradePeriod,
				UpgradeOnce:       upgradeOnce,
				PhaseDuration:     phaseDuration,
				PhaseRecurrence:   phaseRecurrence,
				Phase2Duration:    phase2Duration,
				Phase2Recurrence:  phase2Recurrence,
				SecondPhaseFlagsSet: cmd.Flags().Changed("phase-2-free-region") ||
					cmd.Flags().Changed("phase-2-price") ||
					cmd.Flags().Changed("phase-2-relative-discount") ||
					cmd.Flags().Changed("phase-2-absolute-discount") ||
					cmd.Flags().Changed("phase-2-duration") ||
					cmd.Flags().Changed("phase-2-recurrence"),
				BasicFlagsSet: cmd.Flags().Changed("offer-tag") ||
					cmd.Flags().Changed("free-region") ||
					cmd.Flags().Changed("price") ||
					cmd.Flags().Changed("relative-discount") ||
					cmd.Flags().Changed("absolute-discount") ||
					cmd.Flags().Changed("phase-2-free-region") ||
					cmd.Flags().Changed("phase-2-price") ||
					cmd.Flags().Changed("phase-2-relative-discount") ||
					cmd.Flags().Changed("phase-2-absolute-discount") ||
					cmd.Flags().Changed("other-regions-free") ||
					cmd.Flags().Changed("other-regions-usd-price") ||
					cmd.Flags().Changed("other-regions-eur-price") ||
					cmd.Flags().Changed("phase-2-other-regions-usd-price") ||
					cmd.Flags().Changed("phase-2-other-regions-eur-price") ||
					cmd.Flags().Changed("targeting-acquisition-scope") ||
					cmd.Flags().Changed("targeting-upgrade-scope") ||
					cmd.Flags().Changed("targeting-upgrade-product-id") ||
					cmd.Flags().Changed("targeting-upgrade-billing-period") ||
					cmd.Flags().Changed("targeting-upgrade-once-per-user") ||
					cmd.Flags().Changed("phase-duration") ||
					cmd.Flags().Changed("phase-recurrence") ||
					cmd.Flags().Changed("phase-2-duration") ||
					cmd.Flags().Changed("phase-2-recurrence"),
			})
			if err != nil {
				return err
			}
			createOptions := play.SubscriptionOfferCreateOptions{
				PackageName:    typedPackageName,
				ProductID:      typedProductID,
				BasePlanID:     typedBasePlanID,
				OfferID:        typedOfferID,
				Offer:          offer,
				RegionsVersion: regionsVersion,
				Confirm:        confirm,
				DryRun:         dryRun,
			}
			if dryRun {
				result, err := play.CreateSubscriptionOffer(cmd.Context(), nil, createOptions)
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
			result, err := play.CreateSubscriptionOffer(cmd.Context(), publisher, createOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID",
		"Parent subscription base plan ID",
	)
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Subscription offer ID")
	cmd.Flags().StringVar(&fromJSON, "from-json", "", "Path to a Google Play API or gpc JSON subscription offer body")
	cmd.Flags().StringArrayVar(&offerTags, "offer-tag", nil, "Basic create offer tag; repeatable")
	cmd.Flags().StringArrayVar(&freeRegions, "free-region", nil, "Basic create region with new-subscriber availability and a free phase price mode; repeatable")
	cmd.Flags().StringArrayVar(&prices, "price", nil, "Basic create regional phase price as REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().StringArrayVar(&relativeDiscounts, "relative-discount", nil, "Basic create regional phase relative discount as REGION:0.5, where 0.5 means the user pays 50% of the base plan price; repeatable")
	cmd.Flags().StringArrayVar(&absoluteDiscounts, "absolute-discount", nil, "Basic create regional phase absolute discount as REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().StringArrayVar(&phase2FreeRegions, "phase-2-free-region", nil, "Basic create second phase region with a free price mode; repeatable")
	cmd.Flags().StringArrayVar(&phase2Prices, "phase-2-price", nil, "Basic create second phase regional price as REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().StringArrayVar(&phase2RelDiscount, "phase-2-relative-discount", nil, "Basic create second phase regional relative discount as REGION:0.5; repeatable")
	cmd.Flags().StringArrayVar(&phase2AbsDiscount, "phase-2-absolute-discount", nil, "Basic create second phase regional absolute discount as REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().BoolVar(&otherRegionsFree, "other-regions-free", false, "Basic create free phase mode for other regions")
	cmd.Flags().StringVar(&otherRegionsUSD, "other-regions-usd-price", "", "Basic create first phase other-regions USD price as USD:UNITS[:NANOS]")
	cmd.Flags().StringVar(&otherRegionsEUR, "other-regions-eur-price", "", "Basic create first phase other-regions EUR price as EUR:UNITS[:NANOS]")
	cmd.Flags().StringVar(&phase2OtherUSD, "phase-2-other-regions-usd-price", "", "Basic create second phase other-regions USD price as USD:UNITS[:NANOS]")
	cmd.Flags().StringVar(&phase2OtherEUR, "phase-2-other-regions-eur-price", "", "Basic create second phase other-regions EUR price as EUR:UNITS[:NANOS]")
	cmd.Flags().StringVar(&acquisitionScope, "targeting-acquisition-scope", "", "Basic create acquisition targeting scope: any-subscription-in-app or this-subscription")
	cmd.Flags().StringVar(&upgradeScope, "targeting-upgrade-scope", "", "Basic create upgrade targeting scope: this-subscription or specific-subscription-in-app")
	cmd.Flags().StringVar(&upgradeProductID, "targeting-upgrade-product-id", "", "Basic create upgrade targeting subscription product ID when scope is specific-subscription-in-app")
	cmd.Flags().StringVar(&upgradePeriod, "targeting-upgrade-billing-period", "", "Basic create upgrade targeting billing period duration as an ISO 8601 period")
	cmd.Flags().BoolVar(&upgradeOnce, "targeting-upgrade-once-per-user", false, "Basic create upgrade targeting once-per-user rule")
	cmd.Flags().StringVar(&phaseDuration, "phase-duration", "", "Basic create phase duration as an ISO 8601 period, for example P7D or P1M")
	cmd.Flags().Int64Var(&phaseRecurrence, "phase-recurrence", 1, "Basic create phase recurrence count")
	cmd.Flags().StringVar(&phase2Duration, "phase-2-duration", "", "Basic create second phase duration as an ISO 8601 period, for example P1M")
	cmd.Flags().Int64Var(&phase2Recurrence, "phase-2-recurrence", 1, "Basic create second phase recurrence count")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptionOffers.create")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Create the draft subscription offer")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer creation without calling Google Play")
	return cmd
}

type subscriptionOfferCreateBodyOptions struct {
	FromJSON            string
	OfferTags           []string
	FreeRegions         []string
	Prices              []string
	RelativeDiscounts   []string
	AbsoluteDiscounts   []string
	Phase2FreeRegions   []string
	Phase2Prices        []string
	Phase2RelDiscount   []string
	Phase2AbsDiscount   []string
	OtherRegionsFree    bool
	OtherRegionsUSD     string
	OtherRegionsEUR     string
	Phase2OtherUSD      string
	Phase2OtherEUR      string
	AcquisitionScope    string
	UpgradeScope        string
	UpgradeProductID    string
	UpgradePeriod       string
	UpgradeOnce         bool
	PhaseDuration       string
	PhaseRecurrence     int64
	Phase2Duration      string
	Phase2Recurrence    int64
	SecondPhaseFlagsSet bool
	BasicFlagsSet       bool
}

func subscriptionOfferCreateBody(options subscriptionOfferCreateBodyOptions) (play.SubscriptionOffer, error) {
	if strings.TrimSpace(options.FromJSON) != "" {
		if options.UsesBasicFlags() {
			return play.SubscriptionOffer{}, fmt.Errorf("--from-json cannot be combined with basic create flags")
		}
		return readSubscriptionOfferJSON(options.FromJSON)
	}
	if !options.UsesBasicFlags() {
		return play.SubscriptionOffer{}, fmt.Errorf("subscription offer create requires --from-json or basic create flags")
	}
	regionalConfigs, phases, err := subscriptionOfferCreateBasicPhases(options)
	if err != nil {
		return play.SubscriptionOffer{}, err
	}
	var offerOtherRegionsConfig *play.SubscriptionOfferOtherRegionsConfig
	if options.UsesPaidOtherRegionsFlags() && options.OtherRegionsFree {
		return play.SubscriptionOffer{}, fmt.Errorf("--other-regions-free cannot be combined with paid other-regions price flags")
	}
	if options.OtherRegionsFree {
		offerOtherRegionsConfig = &play.SubscriptionOfferOtherRegionsConfig{NewSubscriberAvailability: true}
		for index := range phases {
			phases[index].OtherRegionsConfig = &play.SubscriptionOfferPhaseOtherRegionsConfig{Free: true}
		}
	}
	if options.UsesPaidOtherRegionsFlags() {
		offerOtherRegionsConfig = &play.SubscriptionOfferOtherRegionsConfig{NewSubscriberAvailability: true}
		if err := setSubscriptionOfferCreatePaidOtherRegions(&phases, options); err != nil {
			return play.SubscriptionOffer{}, err
		}
	}
	targeting, err := subscriptionOfferCreateTargeting(subscriptionOfferCreateTargetingOptions{
		AcquisitionScope: options.AcquisitionScope,
		UpgradeScope:     options.UpgradeScope,
		UpgradeProductID: options.UpgradeProductID,
		UpgradePeriod:    options.UpgradePeriod,
		UpgradeOnce:      options.UpgradeOnce,
	})
	if err != nil {
		return play.SubscriptionOffer{}, err
	}
	return play.SubscriptionOffer{
		OfferTags:          append([]string(nil), options.OfferTags...),
		RegionalConfigs:    regionalConfigs,
		OtherRegionsConfig: offerOtherRegionsConfig,
		Phases:             phases,
		Targeting:          targeting,
	}, nil
}

func (o subscriptionOfferCreateBodyOptions) UsesBasicFlags() bool {
	return o.BasicFlagsSet
}

func subscriptionOfferCreateBasicPhases(options subscriptionOfferCreateBodyOptions) ([]play.SubscriptionOfferRegionalConfig, []play.SubscriptionOfferPhase, error) {
	regionalConfigs, phaseConfigs, err := parseSubscriptionOfferCreatePhaseRegions(options.FreeRegions, options.Prices, options.RelativeDiscounts, options.AbsoluteDiscounts)
	if err != nil {
		return nil, nil, err
	}
	phases := []play.SubscriptionOfferPhase{{
		Duration:        options.PhaseDuration,
		RecurrenceCount: options.PhaseRecurrence,
		RegionalConfigs: phaseConfigs,
	}}
	if !options.UsesSecondPhaseFlags() {
		return regionalConfigs, phases, nil
	}
	_, secondPhaseConfigs, err := parseSubscriptionOfferCreatePhaseRegions(options.Phase2FreeRegions, options.Phase2Prices, options.Phase2RelDiscount, options.Phase2AbsDiscount)
	if err != nil {
		return nil, nil, fmt.Errorf("subscription offer create second phase: %w", err)
	}
	if err := validateSubscriptionOfferCreatePhaseRegionSet(phaseConfigs, secondPhaseConfigs); err != nil {
		return nil, nil, err
	}
	phases = append(phases, play.SubscriptionOfferPhase{
		Duration:        options.Phase2Duration,
		RecurrenceCount: options.Phase2Recurrence,
		RegionalConfigs: secondPhaseConfigs,
	})
	return regionalConfigs, phases, nil
}

func (o subscriptionOfferCreateBodyOptions) UsesSecondPhaseFlags() bool {
	return o.SecondPhaseFlagsSet
}

func (o subscriptionOfferCreateBodyOptions) UsesPaidOtherRegionsFlags() bool {
	return strings.TrimSpace(o.OtherRegionsUSD) != "" ||
		strings.TrimSpace(o.OtherRegionsEUR) != "" ||
		strings.TrimSpace(o.Phase2OtherUSD) != "" ||
		strings.TrimSpace(o.Phase2OtherEUR) != ""
}

func setSubscriptionOfferCreatePaidOtherRegions(phases *[]play.SubscriptionOfferPhase, options subscriptionOfferCreateBodyOptions) error {
	firstConfig, err := subscriptionOfferCreateOtherRegionsPrices(options.OtherRegionsUSD, options.OtherRegionsEUR, "first phase")
	if err != nil {
		return err
	}
	(*phases)[0].OtherRegionsConfig = &play.SubscriptionOfferPhaseOtherRegionsConfig{OtherRegionsPrices: firstConfig}
	if strings.TrimSpace(options.Phase2OtherUSD) == "" && strings.TrimSpace(options.Phase2OtherEUR) == "" {
		if len(*phases) > 1 {
			return fmt.Errorf("second phase other-regions prices are required when paid other-regions prices are used with two phases")
		}
		return nil
	}
	if len(*phases) < 2 {
		return fmt.Errorf("second phase other-regions prices require second phase flags")
	}
	secondConfig, err := subscriptionOfferCreateOtherRegionsPrices(options.Phase2OtherUSD, options.Phase2OtherEUR, "second phase")
	if err != nil {
		return err
	}
	(*phases)[1].OtherRegionsConfig = &play.SubscriptionOfferPhaseOtherRegionsConfig{OtherRegionsPrices: secondConfig}
	return nil
}

func subscriptionOfferCreateOtherRegionsPrices(rawUSDPrice string, rawEURPrice string, phaseLabel string) (*play.SubscriptionOfferOtherRegionsPrices, error) {
	if strings.TrimSpace(rawUSDPrice) == "" || strings.TrimSpace(rawEURPrice) == "" {
		return nil, fmt.Errorf("subscription offer create %s other-regions prices require USD and EUR prices", phaseLabel)
	}
	usdPrice, err := parsePurchaseOptionPatchMoney(rawUSDPrice)
	if err != nil {
		return nil, fmt.Errorf("subscription offer create %s other-regions USD price must use USD:UNITS[:NANOS]", phaseLabel)
	}
	if usdPrice.CurrencyCode != "USD" {
		return nil, fmt.Errorf("subscription offer create %s other-regions USD price currencyCode must be USD", phaseLabel)
	}
	eurPrice, err := parsePurchaseOptionPatchMoney(rawEURPrice)
	if err != nil {
		return nil, fmt.Errorf("subscription offer create %s other-regions EUR price must use EUR:UNITS[:NANOS]", phaseLabel)
	}
	if eurPrice.CurrencyCode != "EUR" {
		return nil, fmt.Errorf("subscription offer create %s other-regions EUR price currencyCode must be EUR", phaseLabel)
	}
	return &play.SubscriptionOfferOtherRegionsPrices{
		USDPrice: &usdPrice,
		EURPrice: &eurPrice,
	}, nil
}

func parseSubscriptionOfferCreatePhaseRegions(freeRegions []string, priceValues []string, relativeDiscountValues []string, absoluteDiscountValues []string) ([]play.SubscriptionOfferRegionalConfig, []play.SubscriptionOfferPhaseRegionalConfig, error) {
	if len(freeRegions) == 0 && len(priceValues) == 0 && len(relativeDiscountValues) == 0 && len(absoluteDiscountValues) == 0 {
		return nil, nil, fmt.Errorf("basic subscription offer create requires at least one --free-region, --price, --relative-discount, or --absolute-discount")
	}
	regionalConfigs := make([]play.SubscriptionOfferRegionalConfig, 0, len(freeRegions)+len(priceValues)+len(relativeDiscountValues)+len(absoluteDiscountValues))
	phaseConfigs := make([]play.SubscriptionOfferPhaseRegionalConfig, 0, len(freeRegions)+len(priceValues)+len(relativeDiscountValues)+len(absoluteDiscountValues))
	for _, value := range freeRegions {
		region := strings.ToUpper(strings.TrimSpace(value))
		regionalConfigs = append(regionalConfigs, play.SubscriptionOfferRegionalConfig{
			RegionCode:                region,
			NewSubscriberAvailability: true,
		})
		phaseConfigs = append(phaseConfigs, play.SubscriptionOfferPhaseRegionalConfig{
			RegionCode: region,
			Free:       true,
		})
	}
	for _, value := range priceValues {
		region, rawPrice, ok := strings.Cut(strings.TrimSpace(value), ":")
		if !ok {
			return nil, nil, errSubscriptionOfferCreatePriceFormat()
		}
		price, err := parsePurchaseOptionPatchMoney(rawPrice)
		if err != nil {
			return nil, nil, errSubscriptionOfferCreatePriceFormat()
		}
		normalizedRegion := strings.ToUpper(strings.TrimSpace(region))
		regionalConfigs = append(regionalConfigs, play.SubscriptionOfferRegionalConfig{
			RegionCode:                normalizedRegion,
			NewSubscriberAvailability: true,
		})
		phaseConfigs = append(phaseConfigs, play.SubscriptionOfferPhaseRegionalConfig{
			RegionCode: normalizedRegion,
			Price:      &price,
		})
	}
	for _, value := range relativeDiscountValues {
		region, rawDiscount, ok := strings.Cut(strings.TrimSpace(value), ":")
		if !ok {
			return nil, nil, errSubscriptionOfferCreateRelativeDiscountFormat()
		}
		relativeDiscount, err := strconv.ParseFloat(strings.TrimSpace(rawDiscount), 64)
		if err != nil {
			return nil, nil, errSubscriptionOfferCreateRelativeDiscountFormat()
		}
		normalizedRegion := strings.ToUpper(strings.TrimSpace(region))
		regionalConfigs = append(regionalConfigs, play.SubscriptionOfferRegionalConfig{
			RegionCode:                normalizedRegion,
			NewSubscriberAvailability: true,
		})
		phaseConfigs = append(phaseConfigs, play.SubscriptionOfferPhaseRegionalConfig{
			RegionCode:       normalizedRegion,
			RelativeDiscount: relativeDiscount,
		})
	}
	for _, value := range absoluteDiscountValues {
		region, rawDiscount, ok := strings.Cut(strings.TrimSpace(value), ":")
		if !ok {
			return nil, nil, errSubscriptionOfferCreateAbsoluteDiscountFormat()
		}
		absoluteDiscount, err := parsePurchaseOptionPatchMoney(rawDiscount)
		if err != nil {
			return nil, nil, errSubscriptionOfferCreateAbsoluteDiscountFormat()
		}
		normalizedRegion := strings.ToUpper(strings.TrimSpace(region))
		regionalConfigs = append(regionalConfigs, play.SubscriptionOfferRegionalConfig{
			RegionCode:                normalizedRegion,
			NewSubscriberAvailability: true,
		})
		phaseConfigs = append(phaseConfigs, play.SubscriptionOfferPhaseRegionalConfig{
			RegionCode:       normalizedRegion,
			AbsoluteDiscount: &absoluteDiscount,
		})
	}
	return regionalConfigs, phaseConfigs, nil
}

func validateSubscriptionOfferCreatePhaseRegionSet(first []play.SubscriptionOfferPhaseRegionalConfig, second []play.SubscriptionOfferPhaseRegionalConfig) error {
	firstRegions := map[string]struct{}{}
	for _, config := range first {
		firstRegions[config.RegionCode] = struct{}{}
	}
	secondRegions := map[string]struct{}{}
	for _, config := range second {
		secondRegions[config.RegionCode] = struct{}{}
		if _, ok := firstRegions[config.RegionCode]; !ok {
			return fmt.Errorf("subscription offer create second phase region %s is not configured in the first phase", config.RegionCode)
		}
	}
	for region := range firstRegions {
		if _, ok := secondRegions[region]; !ok {
			return fmt.Errorf("subscription offer create second phase must configure region %s", region)
		}
	}
	return nil
}

type subscriptionOfferCreateTargetingOptions struct {
	AcquisitionScope string
	UpgradeScope     string
	UpgradeProductID string
	UpgradePeriod    string
	UpgradeOnce      bool
}

func subscriptionOfferCreateTargeting(options subscriptionOfferCreateTargetingOptions) (*play.SubscriptionOfferTargeting, error) {
	acquisitionScope := strings.TrimSpace(options.AcquisitionScope)
	upgradeScope := strings.TrimSpace(options.UpgradeScope)
	upgradeProductID := strings.TrimSpace(options.UpgradeProductID)
	upgradePeriod := strings.TrimSpace(options.UpgradePeriod)
	if acquisitionScope != "" && (upgradeScope != "" || upgradeProductID != "" || upgradePeriod != "" || options.UpgradeOnce) {
		return nil, fmt.Errorf("subscription offer create targeting cannot combine acquisition and upgrade targeting")
	}
	if acquisitionScope != "" {
		return subscriptionOfferCreateAcquisitionTargeting(acquisitionScope)
	}
	if upgradeScope != "" || upgradeProductID != "" || upgradePeriod != "" || options.UpgradeOnce {
		return subscriptionOfferCreateUpgradeTargeting(upgradeScope, upgradeProductID, upgradePeriod, options.UpgradeOnce)
	}
	return nil, nil
}

func subscriptionOfferCreateAcquisitionTargeting(scope string) (*play.SubscriptionOfferTargeting, error) {
	switch scope {
	case "":
		return nil, nil
	case "any-subscription-in-app":
		return &play.SubscriptionOfferTargeting{
			Acquisition: &play.SubscriptionOfferAcquisitionTargeting{
				Scope: &play.SubscriptionOfferTargetingScope{AnySubscriptionInApp: true},
			},
		}, nil
	case "this-subscription":
		return &play.SubscriptionOfferTargeting{
			Acquisition: &play.SubscriptionOfferAcquisitionTargeting{
				Scope: &play.SubscriptionOfferTargetingScope{ThisSubscription: true},
			},
		}, nil
	default:
		return nil, fmt.Errorf("subscription offer create acquisition targeting scope must be any-subscription-in-app or this-subscription")
	}
}

func subscriptionOfferCreateUpgradeTargeting(scope string, productID string, billingPeriod string, oncePerUser bool) (*play.SubscriptionOfferTargeting, error) {
	var targetingScope *play.SubscriptionOfferTargetingScope
	switch scope {
	case "this-subscription":
		if productID != "" {
			return nil, fmt.Errorf("subscription offer create upgrade product ID requires specific-subscription-in-app scope")
		}
		targetingScope = &play.SubscriptionOfferTargetingScope{ThisSubscription: true}
	case "specific-subscription-in-app":
		if productID == "" {
			return nil, fmt.Errorf("subscription offer create upgrade product ID is required for specific-subscription-in-app scope")
		}
		if _, err := play.NewSubscriptionProductID(productID); err != nil {
			return nil, err
		}
		targetingScope = &play.SubscriptionOfferTargetingScope{SpecificSubscriptionInApp: productID}
	case "":
		return nil, fmt.Errorf("subscription offer create upgrade targeting requires --targeting-upgrade-scope")
	default:
		return nil, fmt.Errorf("subscription offer create upgrade targeting scope must be this-subscription or specific-subscription-in-app")
	}
	return &play.SubscriptionOfferTargeting{
		Upgrade: &play.SubscriptionOfferUpgradeTargeting{
			Scope:                 targetingScope,
			BillingPeriodDuration: billingPeriod,
			OncePerUser:           oncePerUser,
		},
	}, nil
}

func errSubscriptionOfferCreatePriceFormat() error {
	return fmt.Errorf("subscription offer create price must use REGION:CURRENCY:UNITS[:NANOS]")
}

func errSubscriptionOfferCreateRelativeDiscountFormat() error {
	return fmt.Errorf("subscription offer create relative discount must use REGION:0.5")
}

func errSubscriptionOfferCreateAbsoluteDiscountFormat() error {
	return fmt.Errorf("subscription offer create absolute discount must use REGION:CURRENCY:UNITS[:NANOS]")
}

func newSubscriptionOffersBatchPatchPhaseRelativeDiscountsCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		relativeDiscount []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-phase-relative-discounts",
		Short: "Batch patch subscription offer phase relative discounts",
		Long: "Batch patch subscription offer phase relative discounts. Omit parent IDs to infer the narrowest valid parent path from --relative-discount values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferPhaseRelativeDiscountPatches(relativeDiscount)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := subscriptionOfferPhaseRelativeDiscountPatchesToMutationRequests(requests)
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, mutationRequests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchSubscriptionOfferPhaseRelativeDiscounts(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferBatchPatchPhaseRelativeDiscountsPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchSubscriptionOfferPhaseRelativeDiscounts(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&relativeDiscount, "relative-discount", nil, "Phase relative discount patch as productId/basePlanId/offerId/phaseIndex/REGION:0.75; phaseIndex is zero-based, so 0 is the first phase; 0.75 means the user pays 75% of the base plan price prorated over the phase duration; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptionOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer phase relative discount batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer phase relative discount batch patch without calling Google Play")
	return cmd
}

func newSubscriptionOffersBatchPatchPhaseAbsoluteDiscountsCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		absoluteDiscount []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-phase-absolute-discounts",
		Short: "Batch patch subscription offer phase absolute discounts",
		Long: "Batch patch subscription offer phase absolute discounts. Omit parent IDs to infer the narrowest valid parent path from --absolute-discount values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferPhaseAbsoluteDiscountPatches(absoluteDiscount)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := subscriptionOfferPhaseAbsoluteDiscountPatchesToMutationRequests(requests)
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, mutationRequests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferBatchPatchPhaseAbsoluteDiscountsPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&absoluteDiscount, "absolute-discount", nil, "Phase absolute discount patch as productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]; phaseIndex is zero-based, so 0 is the first phase; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptionOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer phase absolute discount batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer phase absolute discount batch patch without calling Google Play")
	return cmd
}

func newSubscriptionOffersBatchPatchPhasePricesCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		price            []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-phase-prices",
		Short: "Batch patch subscription offer phase prices",
		Long: "Batch patch subscription offer phase prices. Omit parent IDs to infer the narrowest valid parent path from --price values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferPhasePricePatches(price)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchPatchPhasePricesOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := subscriptionOfferPhasePricePatchesToMutationRequests(requests)
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, mutationRequests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionOfferBatchPatchPhasePricesOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchSubscriptionOfferPhasePrices(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferBatchPatchPhasePricesPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchSubscriptionOfferPhasePrices(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&price, "price", nil, "Phase price patch as productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]; phaseIndex is zero-based, so 0 is the first phase; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptionOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer phase price batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer phase price batch patch without calling Google Play")
	return cmd
}

func newSubscriptionOffersBatchPatchPhaseFreeCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		free             []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-phase-free",
		Short: "Batch patch subscription offer phases to free",
		Long: "Batch patch subscription offer phases to free. Omit parent IDs to infer the narrowest valid parent path from --free values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferPhaseFreePatches(free)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchPatchPhaseFreeOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := subscriptionOfferPhaseFreePatchesToMutationRequests(requests)
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, mutationRequests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionOfferBatchPatchPhaseFreeOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchSubscriptionOfferPhaseFree(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferBatchPatchPhaseFreePlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchSubscriptionOfferPhaseFree(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&free, "free", nil, "Phase free patch as productId/basePlanId/offerId/phaseIndex/REGION; phaseIndex is zero-based, so 0 is the first phase; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptionOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer phase free batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer phase free batch patch without calling Google Play")
	return cmd
}

func parseSubscriptionOfferAvailabilityPatches(values []string) ([]play.SubscriptionOfferAvailabilityPatchRequest, error) {
	requests := make([]play.SubscriptionOfferAvailabilityPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseSubscriptionOfferAvailabilityPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferAvailabilityPatch(value string) (play.SubscriptionOfferAvailabilityPatchRequest, error) {
	path, rawAvailability, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, errSubscriptionOfferAvailabilityFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, errSubscriptionOfferAvailabilityFormat()
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	offerID, err := play.NewSubscriptionOfferID(parts[2])
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	availability, err := parseSubscriptionOfferAvailabilityValue(rawAvailability)
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	return play.SubscriptionOfferAvailabilityPatchRequest{
		ProductID:    productID,
		BasePlanID:   basePlanID,
		OfferID:      offerID,
		RegionCode:   strings.ToUpper(strings.TrimSpace(parts[3])),
		Availability: availability,
	}, nil
}

func errSubscriptionOfferAvailabilityFormat() error {
	return fmt.Errorf("subscription offer availability must use productId/basePlanId/offerId/REGION:true|false")
}

func parseSubscriptionOfferAvailabilityValue(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errSubscriptionOfferAvailabilityFormat()
	}
}

func subscriptionOfferAvailabilityPatchesToMutationRequests(requests []play.SubscriptionOfferAvailabilityPatchRequest) []play.SubscriptionOfferBatchMutationRequest {
	mutations := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return mutations
}

func parseSubscriptionOfferPhaseRelativeDiscountPatches(values []string) ([]play.SubscriptionOfferPhaseRelativeDiscountPatchRequest, error) {
	requests := make([]play.SubscriptionOfferPhaseRelativeDiscountPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseSubscriptionOfferPhaseRelativeDiscountPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferPhaseRelativeDiscountPatch(value string) (play.SubscriptionOfferPhaseRelativeDiscountPatchRequest, error) {
	path, rawRelativeDiscount, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{}, errSubscriptionOfferPhaseRelativeDiscountFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 5 {
		return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{}, errSubscriptionOfferPhaseRelativeDiscountFormat()
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{}, err
	}
	offerID, err := play.NewSubscriptionOfferID(parts[2])
	if err != nil {
		return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{}, err
	}
	phaseIndex, err := strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil {
		return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{}, errSubscriptionOfferPhaseRelativeDiscountFormat()
	}
	relativeDiscount, err := strconv.ParseFloat(strings.TrimSpace(rawRelativeDiscount), 64)
	if err != nil {
		return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{}, errSubscriptionOfferPhaseRelativeDiscountFormat()
	}
	return play.SubscriptionOfferPhaseRelativeDiscountPatchRequest{
		ProductID:        productID,
		BasePlanID:       basePlanID,
		OfferID:          offerID,
		PhaseIndex:       phaseIndex,
		RegionCode:       strings.ToUpper(strings.TrimSpace(parts[4])),
		RelativeDiscount: relativeDiscount,
	}, nil
}

func errSubscriptionOfferPhaseRelativeDiscountFormat() error {
	return fmt.Errorf("subscription offer phase relative discount must use productId/basePlanId/offerId/phaseIndex/REGION:0.75")
}

func subscriptionOfferPhaseRelativeDiscountPatchesToMutationRequests(requests []play.SubscriptionOfferPhaseRelativeDiscountPatchRequest) []play.SubscriptionOfferBatchMutationRequest {
	mutations := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return mutations
}

func parseSubscriptionOfferPhaseAbsoluteDiscountPatches(values []string) ([]play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest, error) {
	requests := make([]play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseSubscriptionOfferPhaseAbsoluteDiscountPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferPhaseAbsoluteDiscountPatch(value string) (play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest, error) {
	path, rawAbsoluteDiscount, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{}, errSubscriptionOfferPhaseAbsoluteDiscountFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 5 {
		return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{}, errSubscriptionOfferPhaseAbsoluteDiscountFormat()
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{}, err
	}
	offerID, err := play.NewSubscriptionOfferID(parts[2])
	if err != nil {
		return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{}, err
	}
	phaseIndex, err := strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil {
		return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{}, errSubscriptionOfferPhaseAbsoluteDiscountFormat()
	}
	absoluteDiscount, err := parsePurchaseOptionPatchMoney(rawAbsoluteDiscount)
	if err != nil {
		return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{}, errSubscriptionOfferPhaseAbsoluteDiscountFormat()
	}
	return play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest{
		ProductID:        productID,
		BasePlanID:       basePlanID,
		OfferID:          offerID,
		PhaseIndex:       phaseIndex,
		RegionCode:       strings.ToUpper(strings.TrimSpace(parts[4])),
		AbsoluteDiscount: absoluteDiscount,
	}, nil
}

func errSubscriptionOfferPhaseAbsoluteDiscountFormat() error {
	return fmt.Errorf("subscription offer phase absolute discount must use productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]")
}

func subscriptionOfferPhaseAbsoluteDiscountPatchesToMutationRequests(requests []play.SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) []play.SubscriptionOfferBatchMutationRequest {
	mutations := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return mutations
}

func parseSubscriptionOfferPhasePricePatches(values []string) ([]play.SubscriptionOfferPhasePricePatchRequest, error) {
	requests := make([]play.SubscriptionOfferPhasePricePatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseSubscriptionOfferPhasePricePatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferPhasePricePatch(value string) (play.SubscriptionOfferPhasePricePatchRequest, error) {
	path, rawPrice, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.SubscriptionOfferPhasePricePatchRequest{}, errSubscriptionOfferPhasePriceFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 5 {
		return play.SubscriptionOfferPhasePricePatchRequest{}, errSubscriptionOfferPhasePriceFormat()
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.SubscriptionOfferPhasePricePatchRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.SubscriptionOfferPhasePricePatchRequest{}, err
	}
	offerID, err := play.NewSubscriptionOfferID(parts[2])
	if err != nil {
		return play.SubscriptionOfferPhasePricePatchRequest{}, err
	}
	phaseIndex, err := strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil {
		return play.SubscriptionOfferPhasePricePatchRequest{}, errSubscriptionOfferPhasePriceFormat()
	}
	price, err := parsePurchaseOptionPatchMoney(rawPrice)
	if err != nil {
		return play.SubscriptionOfferPhasePricePatchRequest{}, errSubscriptionOfferPhasePriceFormat()
	}
	return play.SubscriptionOfferPhasePricePatchRequest{
		ProductID:  productID,
		BasePlanID: basePlanID,
		OfferID:    offerID,
		PhaseIndex: phaseIndex,
		RegionCode: strings.ToUpper(strings.TrimSpace(parts[4])),
		Price:      price,
	}, nil
}

func errSubscriptionOfferPhasePriceFormat() error {
	return fmt.Errorf("subscription offer phase price must use productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]")
}

func subscriptionOfferPhasePricePatchesToMutationRequests(requests []play.SubscriptionOfferPhasePricePatchRequest) []play.SubscriptionOfferBatchMutationRequest {
	mutations := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return mutations
}

func parseSubscriptionOfferPhaseFreePatches(values []string) ([]play.SubscriptionOfferPhaseFreePatchRequest, error) {
	requests := make([]play.SubscriptionOfferPhaseFreePatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseSubscriptionOfferPhaseFreePatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferPhaseFreePatch(value string) (play.SubscriptionOfferPhaseFreePatchRequest, error) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 5 {
		return play.SubscriptionOfferPhaseFreePatchRequest{}, errSubscriptionOfferPhaseFreeFormat()
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.SubscriptionOfferPhaseFreePatchRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.SubscriptionOfferPhaseFreePatchRequest{}, err
	}
	offerID, err := play.NewSubscriptionOfferID(parts[2])
	if err != nil {
		return play.SubscriptionOfferPhaseFreePatchRequest{}, err
	}
	phaseIndex, err := strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil {
		return play.SubscriptionOfferPhaseFreePatchRequest{}, errSubscriptionOfferPhaseFreeFormat()
	}
	return play.SubscriptionOfferPhaseFreePatchRequest{
		ProductID:  productID,
		BasePlanID: basePlanID,
		OfferID:    offerID,
		PhaseIndex: phaseIndex,
		RegionCode: strings.ToUpper(strings.TrimSpace(parts[4])),
	}, nil
}

func errSubscriptionOfferPhaseFreeFormat() error {
	return fmt.Errorf("subscription offer phase free patch must use productId/basePlanId/offerId/phaseIndex/REGION")
}

func subscriptionOfferPhaseFreePatchesToMutationRequests(requests []play.SubscriptionOfferPhaseFreePatchRequest) []play.SubscriptionOfferBatchMutationRequest {
	mutations := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return mutations
}

func readSubscriptionOfferJSON(path string) (play.SubscriptionOffer, error) {
	if strings.TrimSpace(path) == "" {
		return play.SubscriptionOffer{}, fmt.Errorf("subscription offer create requires --from-json")
	}
	data, err := osReadFile(path)
	if err != nil {
		return play.SubscriptionOffer{}, fmt.Errorf("read subscription offer JSON %s: %w", path, err)
	}
	offer, err := play.DecodeSubscriptionOfferCreateJSON(data)
	if err != nil {
		return play.SubscriptionOffer{}, fmt.Errorf("parse subscription offer JSON %s: %w", path, err)
	}
	return offer, nil
}

func newSubscriptionOffersBatchStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.SubscriptionOfferStateAction) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		offers           []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-" + action.String(),
		Short: action.String() + " multiple subscription offers",
		Long: string(action) + " multiple subscription offers. Omit parent IDs to infer the narrowest valid parent path from --offer values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferBatchMutationRequests(offers)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchStateUpdateOptions{PackageName: typedPackageName}.Validate()
			}
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, requests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.SubscriptionOfferBatchStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
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
				result, err := play.BatchUpdateSubscriptionOfferStates(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchUpdateSubscriptionOfferStates(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to update as productId/basePlanId/offerId; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer batch state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer batch state update without calling Google Play")
	return cmd
}

func newSubscriptionOffersStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.SubscriptionOfferStateAction) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		offerID          string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action.String(),
		Short: string(action) + " a subscription offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferGetParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewSubscriptionOfferID(offerID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.SubscriptionOfferStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				OfferID:          typedOfferID,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.UpdateSubscriptionOfferState(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferStateUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateSubscriptionOfferState(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(cmd, &productID, &basePlanID, "Parent subscription product ID", "Parent subscription base plan ID")
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Subscription offer ID")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer state update without calling Google Play")
	return cmd
}

func newSubscriptionOffersBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		offers     []string
	)

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple subscription offers",
		Long:  "Get multiple subscription offers. Use - for both --product-id and --base-plan-id when the batch spans products or base plans. Concrete parent IDs must match every --offer value.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferListParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferBatchRequests(offers)
			if err != nil {
				return err
			}
			batchOptions := play.SubscriptionOfferBatchGetOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				Requests:    requests,
			}
			if err := batchOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchGetSubscriptionOffers(cmd.Context(), publisher, batchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products",
		"Parent subscription base plan ID, or - for offers across base plans",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to fetch as productId/basePlanId/offerId; repeatable, up to 100")
	return cmd
}

func newSubscriptionOffersListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		pageSize   int64
		pageToken  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subscription offers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferListParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			listOptions := play.SubscriptionOfferListOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				PageSize:    pageSize,
				PageToken:   pageToken,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListSubscriptionOffers(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for all products",
		"Parent subscription base plan ID, or - for all base plans",
	)
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum offers to return, capped at 1000")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	return cmd
}

func newSubscriptionOffersGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		offerID    string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one subscription offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferGetParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewSubscriptionOfferID(offerID)
			if err != nil {
				return err
			}
			getOptions := play.SubscriptionOfferGetOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				OfferID:     typedOfferID,
			}
			if err := getOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			offer, err := play.GetSubscriptionOffer(cmd.Context(), publisher, getOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, offer)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID",
		"Parent subscription base plan ID",
	)
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Subscription offer ID")
	return cmd
}

func newSubscriptionOffersDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		offerID    string
		confirm    bool
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a draft subscription offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferGetParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewSubscriptionOfferID(offerID)
			if err != nil {
				return err
			}
			deleteOptions := play.SubscriptionOfferDeleteOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				OfferID:     typedOfferID,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DeleteSubscriptionOffer(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteSubscriptionOffer(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID",
		"Parent subscription base plan ID",
	)
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Subscription offer ID")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer deletion without calling Google Play")
	return cmd
}

func addSubscriptionOfferParentFlags(cmd *cobra.Command, productID *string, basePlanID *string, productIDDescription string, basePlanIDDescription string) {
	cmd.Flags().StringVar(productID, "product-id", "", productIDDescription)
	cmd.Flags().StringVar(basePlanID, "base-plan-id", "", basePlanIDDescription)
}

func parseSubscriptionOfferListParent(packageName string, productID string, basePlanID string) (play.PackageName, play.SubscriptionProductID, play.SubscriptionBasePlanID, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return "", "", "", err
	}
	typedProductID, err := play.NewSubscriptionOfferListProductID(productID)
	if err != nil {
		return "", "", "", err
	}
	typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(basePlanID)
	if err != nil {
		return "", "", "", err
	}
	return typedPackageName, typedProductID, typedBasePlanID, nil
}

func parseSubscriptionOfferGetParent(packageName string, productID string, basePlanID string) (play.PackageName, play.SubscriptionProductID, play.SubscriptionBasePlanID, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return "", "", "", err
	}
	typedProductID, err := play.NewSubscriptionProductID(productID)
	if err != nil {
		return "", "", "", err
	}
	typedBasePlanID, err := play.NewSubscriptionBasePlanID(basePlanID)
	if err != nil {
		return "", "", "", err
	}
	return typedPackageName, typedProductID, typedBasePlanID, nil
}

func parseSubscriptionOfferBatchRequests(values []string) ([]play.SubscriptionOfferBatchGetRequest, error) {
	requests := make([]play.SubscriptionOfferBatchGetRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewSubscriptionOfferBatchGetRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferBatchMutationRequests(values []string) ([]play.SubscriptionOfferBatchMutationRequest, error) {
	requests := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewSubscriptionOfferBatchMutationRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func inferSubscriptionOfferBatchParent(productID string, basePlanID string, requests []play.SubscriptionOfferBatchMutationRequest) (string, string) {
	resolvedProductID := productID
	resolvedBasePlanID := basePlanID
	if len(requests) == 0 {
		return resolvedProductID, resolvedBasePlanID
	}
	firstProductID := requests[0].ProductID.String()
	firstBasePlanID := requests[0].BasePlanID.String()
	oneProduct := true
	oneBasePlan := true
	for _, request := range requests[1:] {
		if request.ProductID.String() != firstProductID {
			oneProduct = false
		}
		if request.BasePlanID.String() != firstBasePlanID {
			oneBasePlan = false
		}
	}
	if resolvedProductID == "" {
		if oneProduct {
			resolvedProductID = firstProductID
		} else {
			resolvedProductID = play.SubscriptionOfferWildcardID
		}
	}
	if resolvedBasePlanID == "" {
		if oneBasePlan {
			resolvedBasePlanID = firstBasePlanID
		} else {
			resolvedBasePlanID = play.SubscriptionOfferWildcardID
		}
	}
	return resolvedProductID, resolvedBasePlanID
}
