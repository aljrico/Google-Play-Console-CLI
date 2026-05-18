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
			typedProductID, err := play.NewInAppProductSKU(productID)
			if err != nil {
				return err
			}
			purchaseOptions := play.ProductPurchaseOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				Token:       typedToken,
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
	cmd.Flags().StringVar(&productID, "product-id", "", "In-app product ID")
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
