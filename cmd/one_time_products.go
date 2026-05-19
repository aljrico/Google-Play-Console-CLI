package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newOneTimeProductsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "one-time-products",
		Short: "Inspect Google Play one-time products",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newOneTimeProductsListCommand(out, options, &packageName),
		newOneTimeProductsGetCommand(out, options, &packageName),
		newOneTimeProductsBatchGetCommand(out, options, &packageName),
		newOneTimeProductsDeleteCommand(out, options, &packageName),
		newOneTimeProductsBatchDeleteCommand(out, options, &packageName),
		newOneTimeProductsPurchaseOptionCommand(out, options, &packageName),
	)
	return cmd
}

func newOneTimeProductsPurchaseOptionCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purchase-option",
		Short: "Manage one-time product purchase options",
	}
	cmd.AddCommand(
		newOneTimeProductsPurchaseOptionBatchDeleteCommand(out, options, packageName),
		newOneTimeProductsPurchaseOptionStateCommand(out, options, packageName, play.PurchaseOptionStateActionActivate),
		newOneTimeProductsPurchaseOptionStateCommand(out, options, packageName, play.PurchaseOptionStateActionDeactivate),
	)
	return cmd
}

func newOneTimeProductsPurchaseOptionBatchDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		parentProductID  string
		purchaseOptions  []string
		latencyTolerance string
		force            bool
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Delete one-time product purchase options",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedParentProductID, err := play.NewOneTimeProductBatchParentProductID(parentProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptions, err := parsePurchaseOptionBatchDeleteRequests(purchaseOptions)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			deleteOptions := play.PurchaseOptionBatchDeleteOptions{
				PackageName:      typedPackageName,
				ParentProductID:  typedParentProductID,
				Requests:         typedPurchaseOptions,
				LatencyTolerance: typedLatencyTolerance,
				Force:            force,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.BatchDeletePurchaseOptions(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchDeletePurchaseOptions(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&parentProductID, "product-id", play.OneTimeProductWildcardID, "Parent one-time product ID, or - when deleting across products")
	cmd.Flags().StringArrayVar(&purchaseOptions, "purchase-option", nil, "Purchase option to delete as productId/purchaseOptionId; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&force, "force", false, "Also delete associated offers under each purchase option")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the purchase option batch deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned purchase option batch deletion without calling Google Play")
	return cmd
}

func parsePurchaseOptionBatchDeleteRequests(values []string) ([]play.PurchaseOptionBatchDeleteRequest, error) {
	requests := make([]play.PurchaseOptionBatchDeleteRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewPurchaseOptionBatchDeleteRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func newOneTimeProductsPurchaseOptionStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.PurchaseOptionStateAction) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action.String(),
		Short: string(action) + " a one-time product purchase option",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewOneTimeProductID(productID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(purchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.PurchaseOptionStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.UpdatePurchaseOptionState(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewPurchaseOptionStateUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdatePurchaseOptionState(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "One-time product ID")
	cmd.Flags().StringVar(&purchaseOptionID, "purchase-option-id", "", "One-time product purchase option ID")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the purchase option state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned purchase option state update without calling Google Play")
	return cmd
}

func newOneTimeProductsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		pageSize  int64
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List one-time products",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			listOptions := play.OneTimeProductListOptions{
				PackageName: typedPackageName,
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
			result, err := play.ListOneTimeProducts(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum one-time products to return, capped at 1000")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	return cmd
}

func newOneTimeProductsGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var productID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a one-time product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewOneTimeProductID(productID)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			product, err := play.GetOneTimeProduct(cmd.Context(), publisher, play.OneTimeProductGetOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, product)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "One-time product ID")
	return cmd
}

func newOneTimeProductsBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var productIDs []string

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple one-time products",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductIDs, err := parseOneTimeProductIDs(productIDs)
			if err != nil {
				return err
			}
			batchOptions := play.OneTimeProductBatchGetOptions{
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
			result, err := play.BatchGetOneTimeProducts(cmd.Context(), publisher, batchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringArrayVar(&productIDs, "product-id", nil, "One-time product ID; repeatable, up to 100")
	return cmd
}

func parseOneTimeProductIDs(values []string) ([]play.OneTimeProductID, error) {
	productIDs := make([]play.OneTimeProductID, 0, len(values))
	for _, value := range values {
		productID, err := play.NewOneTimeProductID(value)
		if err != nil {
			return nil, err
		}
		productIDs = append(productIDs, productID)
	}
	return productIDs, nil
}

func newOneTimeProductsDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a one-time product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductID, err := play.NewOneTimeProductID(productID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			deleteOptions := play.OneTimeProductDeleteOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DeleteOneTimeProduct(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteOneTimeProduct(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&productID, "product-id", "", "One-time product ID")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product deletion without calling Google Play")
	return cmd
}

func newOneTimeProductsBatchDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productIDs       []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Delete multiple one-time products",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedProductIDs, err := parseOneTimeProductIDs(productIDs)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			deleteOptions := play.OneTimeProductBatchDeleteOptions{
				PackageName:      typedPackageName,
				ProductIDs:       typedProductIDs,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.BatchDeleteOneTimeProducts(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchDeleteOneTimeProducts(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringArrayVar(&productIDs, "product-id", nil, "One-time product ID; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product batch deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product batch deletion without calling Google Play")
	return cmd
}
