package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newPurchasesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "purchases",
		Short: "Inspect Google Play purchase tokens",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newPurchasesProductCommand(out, options, &packageName),
		newPurchasesSubscriptionCommand(out, options, &packageName),
		newPurchasesVoidedCommand(out, options, &packageName),
	)
	return cmd
}

func newPurchasesProductCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID string
		token     string
	)

	cmd := &cobra.Command{
		Use:   "product",
		Short: "Get one in-app product purchase",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedToken, err := parsePurchaseParent(*packageName, token)
			if err != nil {
				return err
			}
			purchaseOptions := play.ProductPurchaseOptions{
				PackageName: typedPackageName,
				Token:       typedToken,
			}
			if productID != "" {
				typedProductID, err := play.NewInAppProductSKU(productID)
				if err != nil {
					return err
				}
				purchaseOptions.ProductID = typedProductID
			}
			if err := purchaseOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			purchase, err := play.GetProductPurchase(cmd.Context(), publisher, purchaseOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, purchase)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "Optional in-app product ID hint for stable output when Google omits line items")
	cmd.Flags().StringVar(&token, "token", "", "Purchase token")
	return cmd
}

func newPurchasesSubscriptionCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "subscription",
		Short: "Get one subscription purchase",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedToken, err := parsePurchaseParent(*packageName, token)
			if err != nil {
				return err
			}
			purchaseOptions := play.SubscriptionPurchaseOptions{
				PackageName: typedPackageName,
				Token:       typedToken,
			}
			if err := purchaseOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			purchase, err := play.GetSubscriptionPurchase(cmd.Context(), publisher, purchaseOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, purchase)
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Purchase token")
	return cmd
}

func newPurchasesVoidedCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voided",
		Short: "Inspect voided Google Play purchases",
	}
	cmd.AddCommand(newPurchasesVoidedListCommand(out, options, packageName))
	return cmd
}

func newPurchasesVoidedListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		maxResults                        int64
		startIndex                        int64
		token                             string
		startTimeMillis                   int64
		endTimeMillis                     int64
		purchaseType                      int64
		includeQuantityBasedPartialRefund bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List voided purchases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.VoidedPurchaseListOptions{
				PackageName:                       typedPackageName,
				MaxResults:                        maxResults,
				StartIndex:                        startIndex,
				Token:                             token,
				StartTimeMillis:                   startTimeMillis,
				EndTimeMillis:                     endTimeMillis,
				Type:                              play.VoidedPurchaseType(purchaseType),
				IncludeQuantityBasedPartialRefund: includeQuantityBasedPartialRefund,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListVoidedPurchases(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&maxResults, "max-results", 0, "Maximum voided purchases to return")
	cmd.Flags().Int64Var(&startIndex, "start-index", 0, "Zero-based voided purchase offset")
	cmd.Flags().StringVar(&token, "token", "", "Pagination token from a previous response")
	cmd.Flags().Int64Var(&startTimeMillis, "start-time", 0, "Oldest seen-as-voided time in epoch milliseconds")
	cmd.Flags().Int64Var(&endTimeMillis, "end-time", 0, "Newest seen-as-voided time in epoch milliseconds")
	cmd.Flags().Int64Var(&purchaseType, "type", 0, "Voided purchase type: 0 for products, 1 for products and subscriptions")
	cmd.Flags().BoolVar(&includeQuantityBasedPartialRefund, "include-quantity-based-partial-refund", false, "Include quantity-based partial refunds")
	return cmd
}

func parsePurchaseParent(packageName string, token string) (play.PackageName, play.PurchaseToken, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return "", "", err
	}
	typedToken, err := play.NewPurchaseToken(token)
	if err != nil {
		return "", "", err
	}
	return typedPackageName, typedToken, nil
}
