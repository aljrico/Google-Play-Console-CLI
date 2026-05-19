package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
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
		newSubscriptionsCreateCommand(out, options, &packageName),
		newSubscriptionsBatchGetCommand(out, options, &packageName),
		newSubscriptionsPatchCommand(out, options, &packageName),
		newSubscriptionsBatchPatchListingsCommand(out, options, &packageName),
		newSubscriptionsDeleteCommand(out, options, &packageName),
		newSubscriptionsBasePlanCommand(out, options, &packageName),
	)
	return cmd
}

func newSubscriptionsCreateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID              string
		fromJSON               string
		prepaid                bool
		installments           bool
		listings               []string
		basePlanID             string
		billingPeriod          string
		timeExtension          string
		committedPayments      int64
		renewalType            string
		restrictedCountries    []string
		regionalTaxTiers       []string
		regionalStreamingTaxes []string
		eeaWithdrawalRightType string
		tokenizedDigitalAsset  string
		prices                 []string
		offerTags              []string
		legacyCompatible       bool
		regionsVersion         string
		confirm                bool
		dryRun                 bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a draft subscription",
		Long: "Create a draft subscription from a Google Play API Subscription JSON body or playpub subscription JSON output. " +
			"Basic flags build one auto-renewing, prepaid, or installments base plan; use JSON for compliance or other advanced fields. " +
			"Immutable package and product IDs come from flags and override JSON bodies; output-only subscription and base-plan state is ignored.",
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
			subscription, err := subscriptionCreateBody(subscriptionCreateBodyOptions{
				FromJSON:               fromJSON,
				Prepaid:                prepaid,
				Installments:           installments,
				Listings:               listings,
				BasePlanID:             basePlanID,
				BillingPeriod:          billingPeriod,
				TimeExtension:          timeExtension,
				TimeExtensionSet:       cmd.Flags().Changed("time-extension"),
				CommittedPayments:      committedPayments,
				CommittedPaymentsSet:   cmd.Flags().Changed("committed-payments"),
				RenewalType:            renewalType,
				RenewalTypeSet:         cmd.Flags().Changed("renewal-type"),
				RestrictedCountries:    restrictedCountries,
				RegionalTaxTiers:       regionalTaxTiers,
				RegionalStreamingTaxes: regionalStreamingTaxes,
				EEAWithdrawalRightType: eeaWithdrawalRightType,
				TokenizedDigitalAsset:  tokenizedDigitalAsset,
				Prices:                 prices,
				OfferTags:              offerTags,
				LegacyCompatible:       legacyCompatible,
				LegacySet:              cmd.Flags().Changed("legacy-compatible"),
				BasicFlagsSet: cmd.Flags().Changed("prepaid") ||
					cmd.Flags().Changed("installments") ||
					cmd.Flags().Changed("listing") ||
					cmd.Flags().Changed("base-plan-id") ||
					cmd.Flags().Changed("billing-period") ||
					cmd.Flags().Changed("time-extension") ||
					cmd.Flags().Changed("committed-payments") ||
					cmd.Flags().Changed("renewal-type") ||
					cmd.Flags().Changed("restricted-country") ||
					cmd.Flags().Changed("regional-tax-tier") ||
					cmd.Flags().Changed("regional-streaming-tax") ||
					cmd.Flags().Changed("eea-withdrawal-right-type") ||
					cmd.Flags().Changed("tokenized-digital-asset") ||
					cmd.Flags().Changed("price") ||
					cmd.Flags().Changed("offer-tag") ||
					cmd.Flags().Changed("legacy-compatible"),
			})
			if err != nil {
				return err
			}
			createOptions := play.SubscriptionCreateOptions{
				PackageName:    typedPackageName,
				ProductID:      typedProductID,
				Subscription:   subscription,
				RegionsVersion: regionsVersion,
				Confirm:        confirm,
				DryRun:         dryRun,
			}
			if dryRun {
				result, err := play.CreateSubscription(cmd.Context(), nil, createOptions)
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
			result, err := play.CreateSubscription(cmd.Context(), publisher, createOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID")
	cmd.Flags().StringVar(&fromJSON, "from-json", "", "Path to a Google Play API or playpub JSON subscription body")
	cmd.Flags().BoolVar(&prepaid, "prepaid", false, "Build a basic prepaid base plan instead of an auto-renewing base plan")
	cmd.Flags().BoolVar(&installments, "installments", false, "Build a basic installments base plan instead of an auto-renewing base plan")
	cmd.Flags().StringArrayVar(&listings, "listing", nil, "Basic create listing as CSV language,title,description; repeatable")
	cmd.Flags().StringVar(&basePlanID, "base-plan-id", "", "Basic create base plan ID")
	cmd.Flags().StringVar(&billingPeriod, "billing-period", "", "Basic create billing period: P1W, P4W, P1M, P3M, P6M, or P1Y")
	cmd.Flags().StringVar(&timeExtension, "time-extension", "", "Basic prepaid time extension: TIME_EXTENSION_ACTIVE or TIME_EXTENSION_INACTIVE")
	cmd.Flags().Int64Var(&committedPayments, "committed-payments", 0, "Basic installments committed payments count")
	cmd.Flags().StringVar(&renewalType, "renewal-type", "", "Basic installments renewal type: RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT or RENEWAL_TYPE_RENEWS_WITH_COMMITMENT")
	cmd.Flags().StringArrayVar(&restrictedCountries, "restricted-country", nil, "Basic create restricted payment country as REGION; repeatable")
	cmd.Flags().StringArrayVar(&regionalTaxTiers, "regional-tax-tier", nil, "Basic create regional reduced tax tier as REGION:TAX_TIER; repeatable")
	cmd.Flags().StringArrayVar(&regionalStreamingTaxes, "regional-streaming-tax", nil, "Basic create US streaming tax type as US:STREAMING_TAX_TYPE; repeatable")
	cmd.Flags().StringVar(&eeaWithdrawalRightType, "eea-withdrawal-right-type", "", "Basic create EEA withdrawal right type: WITHDRAWAL_RIGHT_DIGITAL_CONTENT or WITHDRAWAL_RIGHT_SERVICE")
	cmd.Flags().StringVar(&tokenizedDigitalAsset, "tokenized-digital-asset", "", "Basic create tokenized digital asset declaration: true or false")
	cmd.Flags().StringArrayVar(&prices, "price", nil, "Basic create regional price as REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().StringArrayVar(&offerTags, "offer-tag", nil, "Basic create base plan offer tag; repeatable")
	cmd.Flags().BoolVar(&legacyCompatible, "legacy-compatible", true, "Mark the basic auto-renewing base plan as legacy compatible")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptions.create")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Create the draft subscription")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription creation without calling Google Play")
	return cmd
}

type subscriptionCreateBodyOptions struct {
	FromJSON               string
	Prepaid                bool
	Installments           bool
	Listings               []string
	BasePlanID             string
	BillingPeriod          string
	TimeExtension          string
	TimeExtensionSet       bool
	CommittedPayments      int64
	CommittedPaymentsSet   bool
	RenewalType            string
	RenewalTypeSet         bool
	RestrictedCountries    []string
	RegionalTaxTiers       []string
	RegionalStreamingTaxes []string
	EEAWithdrawalRightType string
	TokenizedDigitalAsset  string
	Prices                 []string
	OfferTags              []string
	LegacyCompatible       bool
	LegacySet              bool
	BasicFlagsSet          bool
}

func subscriptionCreateBody(options subscriptionCreateBodyOptions) (play.Subscription, error) {
	if strings.TrimSpace(options.FromJSON) != "" {
		if options.UsesBasicFlags() {
			return play.Subscription{}, fmt.Errorf("--from-json cannot be combined with basic create flags")
		}
		return readSubscriptionJSON(options.FromJSON)
	}
	if !options.UsesBasicFlags() {
		return play.Subscription{}, fmt.Errorf("subscription create requires --from-json or basic create flags")
	}
	listings, err := parseSubscriptionCreateListings(options.Listings)
	if err != nil {
		return play.Subscription{}, err
	}
	regionalConfigs, err := parseSubscriptionCreateRegionalPrices(options.Prices)
	if err != nil {
		return play.Subscription{}, err
	}
	taxComplianceSettings, err := subscriptionCreateTaxComplianceSettings(options)
	if err != nil {
		return play.Subscription{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(options.BasePlanID)
	if err != nil {
		return play.Subscription{}, err
	}
	basePlanType := play.SubscriptionBasePlanTypeAutoRenewing
	legacyCompatible := options.LegacyCompatible
	if options.Prepaid && options.Installments {
		return play.Subscription{}, fmt.Errorf("--prepaid and --installments cannot be used together")
	}
	if options.Prepaid {
		if options.LegacySet {
			return play.Subscription{}, fmt.Errorf("--legacy-compatible cannot be used with --prepaid")
		}
		if options.CommittedPaymentsSet {
			return play.Subscription{}, fmt.Errorf("--committed-payments requires --installments")
		}
		if options.RenewalTypeSet {
			return play.Subscription{}, fmt.Errorf("--renewal-type requires --installments")
		}
		basePlanType = play.SubscriptionBasePlanTypePrepaid
		legacyCompatible = false
	} else if options.Installments {
		if options.LegacySet {
			return play.Subscription{}, fmt.Errorf("--legacy-compatible cannot be used with --installments")
		}
		if options.TimeExtensionSet {
			return play.Subscription{}, fmt.Errorf("--time-extension requires --prepaid")
		}
		basePlanType = play.SubscriptionBasePlanTypeInstallments
		legacyCompatible = false
	} else {
		if options.TimeExtensionSet {
			return play.Subscription{}, fmt.Errorf("--time-extension requires --prepaid")
		}
		if options.CommittedPaymentsSet {
			return play.Subscription{}, fmt.Errorf("--committed-payments requires --installments")
		}
		if options.RenewalTypeSet {
			return play.Subscription{}, fmt.Errorf("--renewal-type requires --installments")
		}
	}
	return play.Subscription{
		Listings:                 listings,
		RestrictedCountries:      parseSubscriptionCreateRestrictedCountries(options.RestrictedCountries),
		TaxAndComplianceSettings: taxComplianceSettings,
		BasePlans: []play.SubscriptionBasePlan{{
			BasePlanID:             basePlanID.String(),
			Type:                   basePlanType,
			BillingPeriodDuration:  options.BillingPeriod,
			TimeExtension:          options.TimeExtension,
			CommittedPaymentsCount: options.CommittedPayments,
			RenewalType:            options.RenewalType,
			LegacyCompatible:       legacyCompatible,
			OfferTags:              append([]string(nil), options.OfferTags...),
			RegionalConfigs:        regionalConfigs,
		}},
	}, nil
}

func subscriptionCreateTaxComplianceSettings(options subscriptionCreateBodyOptions) (*play.ProductTaxComplianceSettings, error) {
	if options.EEAWithdrawalRightType == "" && options.TokenizedDigitalAsset == "" && len(options.RegionalTaxTiers) == 0 && len(options.RegionalStreamingTaxes) == 0 {
		return nil, nil
	}
	settings := play.ProductTaxComplianceSettings{
		EEAWithdrawalRightType: options.EEAWithdrawalRightType,
	}
	if options.TokenizedDigitalAsset != "" {
		tokenizedDigitalAsset, err := strconv.ParseBool(options.TokenizedDigitalAsset)
		if err != nil {
			return nil, err
		}
		settings.IsTokenizedDigitalAsset = &tokenizedDigitalAsset
	}
	regionalTaxTiers, err := parseRegionalProductTaxTiers(options.RegionalTaxTiers)
	if err != nil {
		return nil, err
	}
	taxRateInfo, err := play.RegionalProductTaxTiersToTaxRateInfo(regionalTaxTiers)
	if err != nil {
		return nil, err
	}
	regionalStreamingTaxes, err := parseRegionalProductStreamingTaxes(options.RegionalStreamingTaxes)
	if err != nil {
		return nil, err
	}
	streamingTaxRateInfo, err := play.RegionalProductStreamingTaxesToTaxRateInfo(regionalStreamingTaxes)
	if err != nil {
		return nil, err
	}
	settings.TaxRateInfoByRegionCode = play.MergeRegionalTaxRateInfo(taxRateInfo, streamingTaxRateInfo)
	return &settings, nil
}

func parseSubscriptionCreateRestrictedCountries(values []string) []string {
	countries := make([]string, 0, len(values))
	for _, value := range values {
		countries = append(countries, strings.ToUpper(strings.TrimSpace(value)))
	}
	return countries
}

func (o subscriptionCreateBodyOptions) UsesBasicFlags() bool {
	return o.BasicFlagsSet
}

func readSubscriptionJSON(path string) (play.Subscription, error) {
	if strings.TrimSpace(path) == "" {
		return play.Subscription{}, fmt.Errorf("subscription create requires --from-json")
	}
	data, err := osReadFile(path)
	if err != nil {
		return play.Subscription{}, fmt.Errorf("read subscription JSON %s: %w", path, err)
	}
	subscription, err := play.DecodeSubscriptionCreateJSON(data)
	if err != nil {
		return play.Subscription{}, fmt.Errorf("parse subscription JSON %s: %w", path, err)
	}
	return subscription, nil
}

func parseSubscriptionCreateListings(values []string) ([]play.SubscriptionListing, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("basic subscription create requires at least one --listing")
	}
	listings := make([]play.SubscriptionListing, 0, len(values))
	for _, value := range values {
		listing, err := parseSubscriptionCreateListing(value)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}
	return listings, nil
}

func parseSubscriptionCreateListing(value string) (play.SubscriptionListing, error) {
	reader := csv.NewReader(strings.NewReader(value))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return play.SubscriptionListing{}, fmt.Errorf("parse subscription create listing CSV: %w", err)
	}
	if len(records) != 1 {
		return play.SubscriptionListing{}, fmt.Errorf("subscription create listing must contain exactly one CSV record")
	}
	fields := records[0]
	if len(fields) != 3 {
		return play.SubscriptionListing{}, fmt.Errorf("subscription create listing must be CSV language,title,description")
	}
	language, err := play.NewListingLanguage(fields[0])
	if err != nil {
		return play.SubscriptionListing{}, err
	}
	return play.SubscriptionListing{
		LanguageCode: language.String(),
		Title:        fields[1],
		Description:  fields[2],
	}, nil
}

func parseSubscriptionCreateRegionalPrices(values []string) ([]play.SubscriptionRegionalConfig, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("basic subscription create requires at least one --price")
	}
	configs := make([]play.SubscriptionRegionalConfig, 0, len(values))
	for _, value := range values {
		config, err := parseSubscriptionCreateRegionalPrice(value)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func parseSubscriptionCreateRegionalPrice(value string) (play.SubscriptionRegionalConfig, error) {
	regionCode, priceValue, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.SubscriptionRegionalConfig{}, errSubscriptionCreateRegionalPriceFormat()
	}
	price, err := parseRegionalPricePatchMoney(priceValue, errSubscriptionCreateRegionalPriceFormat)
	if err != nil {
		return play.SubscriptionRegionalConfig{}, err
	}
	return play.SubscriptionRegionalConfig{
		RegionCode:                strings.ToUpper(strings.TrimSpace(regionCode)),
		NewSubscriberAvailability: true,
		Price:                     &price,
	}, nil
}

func errSubscriptionCreateRegionalPriceFormat() error {
	return fmt.Errorf("subscription create price must use REGION:CURRENCY:UNITS[:NANOS]")
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
		newSubscriptionsBasePlanBatchPatchPricesCommand(out, options, packageName),
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

func newSubscriptionsBasePlanBatchPatchPricesCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		prices           []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-prices",
		Short: "Batch patch subscription base plan regional prices",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			resolvedProductID, requests, err := parseBasePlanPricePatches(productID, prices)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewSubscriptionBasePlanBatchProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.BasePlanBatchPatchPriceOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				RegionsVersion:   regionsVersion,
				Requests:         requests,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchBasePlanPrices(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewBasePlanBatchPatchPricePlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchBasePlanPrices(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Subscription product ID, or - for price patches across subscriptions; inferred from --price values")
	cmd.Flags().StringArrayVar(&prices, "price", nil, "Regional price patch as productId/basePlanId/REGION:CURRENCY:UNITS[:NANOS]; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptions.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the base plan price batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned base plan price batch patch without calling Google Play")
	return cmd
}

func parseBasePlanPricePatches(productID string, values []string) (string, []play.BasePlanPricePatchRequest, error) {
	requests := make([]play.BasePlanPricePatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseBasePlanPricePatch(value)
		if err != nil {
			return "", nil, err
		}
		requests = append(requests, request)
	}
	return inferBasePlanPricePatchProductID(productID, requests), requests, nil
}

func parseBasePlanPricePatch(value string) (play.BasePlanPricePatchRequest, error) {
	path, priceValue, ok := strings.Cut(value, ":")
	if !ok {
		return play.BasePlanPricePatchRequest{}, errBasePlanPricePatchFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		return play.BasePlanPricePatchRequest{}, errBasePlanPricePatchFormat()
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.BasePlanPricePatchRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.BasePlanPricePatchRequest{}, err
	}
	price, err := parseBasePlanPatchMoney(priceValue)
	if err != nil {
		return play.BasePlanPricePatchRequest{}, err
	}
	return play.BasePlanPricePatchRequest{
		ProductID:  productID,
		BasePlanID: basePlanID,
		RegionCode: strings.ToUpper(parts[2]),
		Price:      price,
	}, nil
}

func parseBasePlanPatchMoney(value string) (play.Money, error) {
	return parseRegionalPricePatchMoney(value, errBasePlanPricePatchFormat)
}

func inferBasePlanPricePatchProductID(productID string, requests []play.BasePlanPricePatchRequest) string {
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

func errBasePlanPricePatchFormat() error {
	return fmt.Errorf("base plan price must use productId/basePlanId/REGION:CURRENCY:UNITS[:NANOS]")
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
