package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newPricingCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "pricing",
		Short: "Inspect Google Play price conversions",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(newPricingConvertRegionPricesCommand(out, options, &packageName))
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
