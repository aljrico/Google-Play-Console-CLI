package cmd

import (
	"io"

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
	)
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
