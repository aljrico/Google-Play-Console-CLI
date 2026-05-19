package cmd

import (
	"io"

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
	)
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
