package cmd

import (
	"io"
	"strconv"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newInAppProductsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "in-app-products",
		Short: "Inspect legacy Google Play in-app products",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newInAppProductsListCommand(out, options, &packageName),
		newInAppProductsGetCommand(out, options, &packageName),
		newInAppProductsBatchGetCommand(out, options, &packageName),
		newInAppProductsCreateCommand(out, options, &packageName),
		newInAppProductsPatchCommand(out, options, &packageName),
		newInAppProductsDeleteCommand(out, options, &packageName),
		newInAppProductsBatchDeleteCommand(out, options, &packageName),
	)
	return cmd
}

func newInAppProductsBatchDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		skus             []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Delete multiple legacy managed in-app products",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedSKUs, err := parseInAppProductSKUs(skus)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			deleteOptions := play.InAppProductBatchDeleteOptions{
				PackageName:      typedPackageName,
				SKUs:             typedSKUs,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchDeleteInAppProducts(cmd.Context(), nil, deleteOptions)
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
			result, err := play.BatchDeleteInAppProducts(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringArrayVar(&skus, "sku", nil, "In-app product SKU; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Delete the managed in-app products")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned managed in-app product batch deletion without calling Google Play")
	return cmd
}

func newInAppProductsDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		sku              string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a legacy managed in-app product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedSKU, err := play.NewInAppProductSKU(sku)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			deleteOptions := play.InAppProductDeleteOptions{
				PackageName:      typedPackageName,
				SKU:              typedSKU,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.DeleteInAppProduct(cmd.Context(), nil, deleteOptions)
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
			result, err := play.DeleteInAppProduct(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&sku, "sku", "", "In-app product SKU")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Delete the managed in-app product")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned managed in-app product deletion without calling Google Play")
	return cmd
}

func newInAppProductsBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var skus []string

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple legacy in-app products",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedSKUs, err := parseInAppProductSKUs(skus)
			if err != nil {
				return err
			}
			batchOptions := play.InAppProductBatchGetOptions{
				PackageName: typedPackageName,
				SKUs:        typedSKUs,
			}
			if err := batchOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchGetInAppProducts(cmd.Context(), publisher, batchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringArrayVar(&skus, "sku", nil, "In-app product SKU; repeatable")
	return cmd
}

func newInAppProductsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List legacy in-app products",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.InAppProductListOptions{
				PackageName: typedPackageName,
				Token:       token,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListInAppProducts(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Pagination token from a previous response")
	return cmd
}

func parseInAppProductSKUs(values []string) ([]play.InAppProductSKU, error) {
	skus := make([]play.InAppProductSKU, 0, len(values))
	for _, value := range values {
		sku, err := play.NewInAppProductSKU(value)
		if err != nil {
			return nil, err
		}
		skus = append(skus, sku)
	}
	return skus, nil
}

func newInAppProductsGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var sku string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one legacy in-app product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedSKU, err := play.NewInAppProductSKU(sku)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			product, err := play.GetInAppProduct(cmd.Context(), publisher, play.InAppProductGetOptions{
				PackageName: typedPackageName,
				SKU:         typedSKU,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, product)
		},
	}
	cmd.Flags().StringVar(&sku, "sku", "", "In-app product SKU")
	return cmd
}

func newInAppProductsCreateCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		sku             string
		status          string
		defaultLanguage string
		defaultPrice    string
		title           string
		description     string
		confirm         bool
		dryRun          bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a legacy managed in-app product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedSKU, err := play.NewInAppProductSKU(sku)
			if err != nil {
				return err
			}
			typedStatus, err := play.NewProductStatus(status)
			if err != nil {
				return err
			}
			typedDefaultLanguage, err := play.NewListingLanguage(defaultLanguage)
			if err != nil {
				return err
			}
			typedDefaultPrice, err := play.NewProductPrice(defaultPrice)
			if err != nil {
				return err
			}
			createOptions := play.InAppProductCreateOptions{
				PackageName:     typedPackageName,
				SKU:             typedSKU,
				Status:          typedStatus,
				DefaultLanguage: typedDefaultLanguage,
				DefaultPrice:    typedDefaultPrice,
				Listing: play.InAppProductListing{
					Title:       title,
					Description: description,
				},
				Confirm: confirm,
				DryRun:  dryRun,
			}
			return runInAppProductCreate(cmd, out, options, createOptions)
		},
	}
	cmd.Flags().StringVar(&sku, "sku", "", "In-app product SKU")
	cmd.Flags().StringVar(&status, "status", play.ProductStatusInactive.String(), "Initial product status: active or inactive")
	cmd.Flags().StringVar(&defaultLanguage, "default-language", "", "Default BCP-47 listing language, for example en-US")
	cmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default checkout price as CURRENCY:MICROS, for example USD:1990000")
	cmd.Flags().StringVar(&title, "title", "", "Default listing title")
	cmd.Flags().StringVar(&description, "description", "", "Default listing description")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Create the managed in-app product")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned managed in-app product creation without calling Google Play")
	return cmd
}

func runInAppProductCreate(cmd *cobra.Command, out io.Writer, options *globalOptions, createOptions play.InAppProductCreateOptions) error {
	if createOptions.DryRun {
		result, err := play.CreateInAppProduct(cmd.Context(), nil, createOptions)
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
	result, err := play.CreateInAppProduct(cmd.Context(), publisher, createOptions)
	if err != nil {
		return err
	}
	return output.Write(out, options.output, options.pretty, result)
}

func newInAppProductsPatchCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		sku                    string
		status                 string
		defaultLanguage        string
		listingLanguage        string
		defaultPrice           string
		regionalPrices         []string
		regionalTaxTiers       []string
		regionalStreamingTaxes []string
		eeaWithdrawalRightType string
		tokenizedDigitalAsset  string
		title                  string
		description            string
		confirm                bool
		dryRun                 bool
	)

	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Patch a legacy managed in-app product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedSKU, err := play.NewInAppProductSKU(sku)
			if err != nil {
				return err
			}
			patchOptions := play.InAppProductPatchOptions{
				PackageName: typedPackageName,
				SKU:         typedSKU,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if status != "" {
				typedStatus, err := play.NewProductStatus(status)
				if err != nil {
					return err
				}
				patchOptions.Status = typedStatus
			}
			if defaultLanguage != "" {
				typedDefaultLanguage, err := play.NewListingLanguage(defaultLanguage)
				if err != nil {
					return err
				}
				patchOptions.DefaultLanguage = typedDefaultLanguage
			}
			if listingLanguage != "" {
				typedListingLanguage, err := play.NewListingLanguage(listingLanguage)
				if err != nil {
					return err
				}
				patchOptions.ListingLanguage = typedListingLanguage
			}
			if defaultPrice != "" {
				typedDefaultPrice, err := play.NewProductPrice(defaultPrice)
				if err != nil {
					return err
				}
				patchOptions.DefaultPrice = &typedDefaultPrice
			}
			for _, regionalPrice := range regionalPrices {
				typedRegionalPrice, err := play.NewRegionalProductPrice(regionalPrice)
				if err != nil {
					return err
				}
				patchOptions.RegionalPrices = append(patchOptions.RegionalPrices, typedRegionalPrice)
			}
			if eeaWithdrawalRightType != "" || tokenizedDigitalAsset != "" || len(regionalTaxTiers) > 0 || len(regionalStreamingTaxes) > 0 {
				taxSettings := play.ProductTaxComplianceSettings{
					EEAWithdrawalRightType: eeaWithdrawalRightType,
				}
				if tokenizedDigitalAsset != "" {
					parsedTokenizedDigitalAsset, err := strconv.ParseBool(tokenizedDigitalAsset)
					if err != nil {
						return err
					}
					taxSettings.IsTokenizedDigitalAsset = &parsedTokenizedDigitalAsset
				}
				typedRegionalTaxTiers, err := parseRegionalProductTaxTiers(regionalTaxTiers)
				if err != nil {
					return err
				}
				taxRateInfo, err := play.RegionalProductTaxTiersToTaxRateInfo(typedRegionalTaxTiers)
				if err != nil {
					return err
				}
				typedRegionalStreamingTaxes, err := parseRegionalProductStreamingTaxes(regionalStreamingTaxes)
				if err != nil {
					return err
				}
				streamingTaxRateInfo, err := play.RegionalProductStreamingTaxesToTaxRateInfo(typedRegionalStreamingTaxes)
				if err != nil {
					return err
				}
				taxSettings.TaxRateInfoByRegionCode = play.MergeRegionalTaxRateInfo(taxRateInfo, streamingTaxRateInfo)
				patchOptions.TaxComplianceSettings = &taxSettings
			}
			if title != "" || description != "" {
				patchOptions.Listing = &play.InAppProductListing{
					Title:       title,
					Description: description,
				}
			}
			return runInAppProductPatch(cmd, out, options, patchOptions)
		},
	}
	cmd.Flags().StringVar(&sku, "sku", "", "In-app product SKU")
	cmd.Flags().StringVar(&status, "status", "", "Product status: active or inactive")
	cmd.Flags().StringVar(&defaultLanguage, "default-language", "", "Default BCP-47 listing language to set on the product")
	cmd.Flags().StringVar(&listingLanguage, "listing-language", "", "BCP-47 listing language to update when --title and --description are set")
	cmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default checkout price as CURRENCY:MICROS, for example USD:1990000")
	cmd.Flags().StringArrayVar(&regionalPrices, "regional-price", nil, "Regional checkout price as REGION:CURRENCY:MICROS, for example US:USD:2990000; repeatable")
	cmd.Flags().StringArrayVar(&regionalTaxTiers, "regional-tax-tier", nil, "Regional reduced tax tier as REGION:TAX_TIER, for example FR:TAX_TIER_NEWS_1; repeatable")
	cmd.Flags().StringArrayVar(&regionalStreamingTaxes, "regional-streaming-tax", nil, "US streaming tax type as US:STREAMING_TAX_TYPE, for example US:STREAMING_TAX_TYPE_TELCO_VIDEO_SALES; repeatable")
	cmd.Flags().StringVar(&eeaWithdrawalRightType, "eea-withdrawal-right-type", "", "EEA withdrawal right type: WITHDRAWAL_RIGHT_DIGITAL_CONTENT or WITHDRAWAL_RIGHT_SERVICE")
	cmd.Flags().StringVar(&tokenizedDigitalAsset, "tokenized-digital-asset", "", "Whether the managed product represents a tokenized digital asset: true or false")
	cmd.Flags().StringVar(&title, "title", "", "Default listing title")
	cmd.Flags().StringVar(&description, "description", "", "Default listing description")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the managed in-app product patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned managed in-app product patch without calling Google Play")
	return cmd
}

func parseRegionalProductTaxTiers(values []string) ([]play.RegionalProductTaxTier, error) {
	tiers := make([]play.RegionalProductTaxTier, 0, len(values))
	for _, value := range values {
		tier, err := play.NewRegionalProductTaxTier(value)
		if err != nil {
			return nil, err
		}
		tiers = append(tiers, tier)
	}
	return tiers, nil
}

func parseRegionalProductStreamingTaxes(values []string) ([]play.RegionalProductStreamingTax, error) {
	taxes := make([]play.RegionalProductStreamingTax, 0, len(values))
	for _, value := range values {
		tax, err := play.NewRegionalProductStreamingTax(value)
		if err != nil {
			return nil, err
		}
		taxes = append(taxes, tax)
	}
	return taxes, nil
}

func runInAppProductPatch(cmd *cobra.Command, out io.Writer, options *globalOptions, patchOptions play.InAppProductPatchOptions) error {
	if patchOptions.DryRun {
		result, err := play.PatchInAppProduct(cmd.Context(), nil, patchOptions)
		if err != nil {
			return err
		}
		return output.Write(out, options.output, options.pretty, result)
	}
	if err := patchOptions.Validate(); err != nil {
		return err
	}
	publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
	if err != nil {
		return err
	}
	result, err := play.PatchInAppProduct(cmd.Context(), publisher, patchOptions)
	if err != nil {
		return err
	}
	return output.Write(out, options.output, options.pretty, result)
}
