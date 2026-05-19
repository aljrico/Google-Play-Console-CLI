package cmd

import (
	"fmt"
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newPurchasesCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "purchases",
		Short: "Inspect and manage Google Play purchase tokens",
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
		Short: "Get or mutate one in-app product purchase",
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
	cmd.AddCommand(
		newPurchasesProductAcknowledgeCommand(out, options, packageName),
		newPurchasesProductConsumeCommand(out, options, packageName),
	)
	return cmd
}

func newPurchasesProductAcknowledgeCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	return newPurchasesProductMutationCommand(out, options, packageName, "acknowledge", "Acknowledge an in-app product purchase")
}

func newPurchasesProductConsumeCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	return newPurchasesProductMutationCommand(out, options, packageName, "consume", "Consume an in-app product purchase")
}

func newPurchasesProductMutationCommand(out io.Writer, options *globalOptions, packageName *string, action string, short string) *cobra.Command {
	var (
		productID        string
		token            string
		developerPayload string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action,
		Short: short,
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
			mutationOptions := play.ProductPurchaseMutationOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				Token:            typedToken,
				DeveloperPayload: developerPayload,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := mutationOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				return runProductPurchaseMutation(cmd, out, options, nil, mutationOptions, action)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			return runProductPurchaseMutation(cmd, out, options, publisher, mutationOptions, action)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "In-app product ID")
	cmd.Flags().StringVar(&token, "token", "", "Purchase token")
	if action == "acknowledge" {
		cmd.Flags().StringVar(&developerPayload, "developer-payload", "", "Optional developer payload to attach to the acknowledgement")
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the product purchase mutation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned product purchase mutation without calling Google Play")
	return cmd
}

func runProductPurchaseMutation(cmd *cobra.Command, out io.Writer, options *globalOptions, mutator play.ProductPurchaseMutator, mutationOptions play.ProductPurchaseMutationOptions, action string) error {
	var (
		result play.ProductPurchaseMutationResult
		err    error
	)
	switch action {
	case "acknowledge":
		result, err = play.AcknowledgeProductPurchase(cmd.Context(), mutator, mutationOptions)
	case "consume":
		result, err = play.ConsumeProductPurchase(cmd.Context(), mutator, mutationOptions)
	default:
		err = fmt.Errorf("unsupported product purchase action %q", action)
	}
	if err != nil {
		return err
	}
	return output.Write(out, options.output, options.pretty, result)
}

func newPurchasesSubscriptionCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "subscription",
		Short: "Get or revoke one subscription purchase",
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
	cmd.AddCommand(
		newPurchasesSubscriptionMutationCommand(out, options, packageName, play.SubscriptionPurchaseMutationActionAcknowledge, "Acknowledge a subscription purchase through the legacy subscriptions API"),
		newPurchasesSubscriptionMutationCommand(out, options, packageName, play.SubscriptionPurchaseMutationActionCancel, "Cancel a subscription purchase through the legacy subscriptions API"),
		newPurchasesSubscriptionRevokeCommand(out, options, packageName),
	)
	return cmd
}

func newPurchasesSubscriptionMutationCommand(out io.Writer, options *globalOptions, packageName *string, action play.SubscriptionPurchaseMutationAction, short string) *cobra.Command {
	var (
		subscriptionID   string
		token            string
		developerPayload string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action.String(),
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedToken, err := parsePurchaseParent(*packageName, token)
			if err != nil {
				return err
			}
			typedSubscriptionID, err := play.NewSubscriptionProductID(subscriptionID)
			if err != nil {
				return err
			}
			mutationOptions := play.SubscriptionPurchaseMutationOptions{
				PackageName:      typedPackageName,
				SubscriptionID:   typedSubscriptionID,
				Token:            typedToken,
				Action:           action,
				DeveloperPayload: developerPayload,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := mutationOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.MutateSubscriptionPurchase(cmd.Context(), nil, mutationOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.MutateSubscriptionPurchase(cmd.Context(), publisher, mutationOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Legacy subscription product ID")
	cmd.Flags().StringVar(&token, "token", "", "Purchase token")
	if action == play.SubscriptionPurchaseMutationActionAcknowledge {
		cmd.Flags().StringVar(&developerPayload, "developer-payload", "", "Optional developer payload to attach to the acknowledgement")
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription purchase mutation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription purchase mutation without calling Google Play")
	return cmd
}

func newPurchasesSubscriptionRevokeCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		token      string
		refundType string
		confirm    bool
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a subscription purchase",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedToken, err := parsePurchaseParent(*packageName, token)
			if err != nil {
				return err
			}
			typedRefundType, err := play.NewSubscriptionRefundType(refundType)
			if err != nil {
				return err
			}
			revokeOptions := play.SubscriptionPurchaseRevokeOptions{
				PackageName: typedPackageName,
				Token:       typedToken,
				RefundType:  typedRefundType,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if err := revokeOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.RevokeSubscriptionPurchase(cmd.Context(), nil, revokeOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.RevokeSubscriptionPurchase(cmd.Context(), publisher, revokeOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Purchase token")
	cmd.Flags().StringVar(&refundType, "refund", "", "Refund type: full or prorated")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription revocation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription revocation without calling Google Play")
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
