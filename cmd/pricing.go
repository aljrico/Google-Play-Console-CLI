package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newPricingCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "pricing",
		Short: "Build and inspect Google Play price conversions",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newPricingConvertRegionPricesCommand(out, options, &packageName), newPricingBuildPricePatchesCommand(out, options, &packageName))
	return cmd
}

func newPricingConvertRegionPricesCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		currency string
		units    int64
		nanos    int64
	)

	cmd := &cobra.Command{
		Use:   "convert-region-prices",
		Short: "Convert one source price into Play region prices",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedCurrency, err := play.NewCurrencyCode(currency)
			if err != nil {
				return err
			}
			convertOptions := play.RegionPriceConversionOptions{
				PackageName: typedPackageName,
				Currency:    typedCurrency,
				Units:       units,
				Nanos:       nanos,
			}
			if err := convertOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ConvertRegionPrices(cmd.Context(), publisher, convertOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&currency, "currency", "", "Source price currency code, for example USD")
	cmd.Flags().Int64Var(&units, "units", 0, "Whole source price units")
	cmd.Flags().Int64Var(&nanos, "nanos", 0, "Fractional source price nanos, 0 to 999999999")
	return cmd
}

func newPricingBuildPricePatchesCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		fromJSON         string
		target           string
		sku              string
		productID        string
		purchaseOptionID string
		basePlanID       string
		offerID          string
		phaseIndex       int
	)

	cmd := &cobra.Command{
		Use:   "build-price-patches",
		Short: "Build regional price patch arguments from converted Play prices",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			conversion, err := readRegionPriceConversionResult(fromJSON)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("package") || cmd.InheritedFlags().Changed("package") {
				typedPackageName, err := play.NewPackageName(*packageName)
				if err != nil {
					return err
				}
				if conversion.PackageName != typedPackageName {
					return fmt.Errorf("--package %s does not match conversion packageName %s", typedPackageName, conversion.PackageName)
				}
			}
			buildOptions := play.PricePatchBuildOptions{
				Conversion:       conversion,
				Target:           play.PricePatchBuildTarget(target),
				SKU:              play.InAppProductSKU(sku),
				ProductID:        productID,
				PurchaseOptionID: play.OneTimeProductPurchaseOptionID(purchaseOptionID),
				BasePlanID:       play.SubscriptionBasePlanID(basePlanID),
				OfferID:          play.SubscriptionOfferID(offerID),
				PhaseIndex:       phaseIndex,
			}
			result, err := play.BuildPricePatches(buildOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&fromJSON, "from-json", "", "Path to playpub pricing convert-region-prices JSON output")
	cmd.Flags().StringVar(&target, "target", "", "Patch target: in-app-product, one-time-product, subscription-base-plan, or subscription-offer-phase")
	cmd.Flags().StringVar(&sku, "sku", "", "In-app product SKU for --target in-app-product")
	cmd.Flags().StringVar(&productID, "product-id", "", "One-time product or subscription product ID")
	cmd.Flags().StringVar(&purchaseOptionID, "purchase-option-id", "", "Purchase option ID for --target one-time-product")
	cmd.Flags().StringVar(&basePlanID, "base-plan-id", "", "Base plan ID for subscription targets")
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Offer ID for --target subscription-offer-phase")
	cmd.Flags().IntVar(&phaseIndex, "phase-index", -1, "Zero-based offer phase index for --target subscription-offer-phase")
	return cmd
}

func readRegionPriceConversionResult(path string) (play.RegionPriceConversionResult, error) {
	if path == "" {
		return play.RegionPriceConversionResult{}, fmt.Errorf("--from-json is required")
	}
	data, err := osReadFile(path)
	if err != nil {
		return play.RegionPriceConversionResult{}, fmt.Errorf("read price conversion JSON: %w", err)
	}
	var result play.RegionPriceConversionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return play.RegionPriceConversionResult{}, fmt.Errorf("decode price conversion JSON: %w", err)
	}
	if err := assertNoDuplicateConvertedRegionKeys(data); err != nil {
		return play.RegionPriceConversionResult{}, fmt.Errorf("decode price conversion JSON: %w", err)
	}
	return result, nil
}

func assertNoDuplicateConvertedRegionKeys(data []byte) error {
	var raw struct {
		ConvertedRegionPrices json.RawMessage `json:"convertedRegionPrices"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.ConvertedRegionPrices) == 0 {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw.ConvertedRegionPrices))
	tok, err := decoder.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		keyTok, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("expected string key in convertedRegionPrices")
		}
		if _, dup := seen[key]; dup {
			return fmt.Errorf("converted region price %q is duplicated", key)
		}
		seen[key] = struct{}{}
		var skip json.RawMessage
		if err := decoder.Decode(&skip); err != nil {
			return err
		}
	}
	return nil
}
