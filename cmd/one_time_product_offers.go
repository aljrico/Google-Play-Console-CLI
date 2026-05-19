package cmd

import (
	"fmt"
	"io"
	"strings"

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
		newOneTimeProductOffersBatchGetCommand(out, options, &packageName),
		newOneTimeProductOffersBatchDeleteCommand(out, options, &packageName),
		newOneTimeProductOffersBatchPatchAvailabilityCommand(out, options, &packageName),
		newOneTimeProductOffersBatchStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionActivate),
		newOneTimeProductOffersBatchStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionDeactivate),
		newOneTimeProductOffersBatchStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionCancel),
		newOneTimeProductOffersStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionActivate),
		newOneTimeProductOffersStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionDeactivate),
		newOneTimeProductOffersStateCommand(out, options, &packageName, play.OneTimeProductOfferStateActionCancel),
	)
	return cmd
}

func newOneTimeProductOffersBatchPatchAvailabilityCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		availability     []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-availability",
		Short: "Batch patch one-time product offer regional availability",
		Long: "Batch patch one-time product offer regional availability. Omit parent IDs to infer the narrowest valid parent path from --availability values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferAvailabilityPatches(availability)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.OneTimeProductOfferBatchPatchAvailabilityOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := oneTimeProductOfferAvailabilityPatchesToMutationRequests(requests)
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, mutationRequests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.OneTimeProductOfferBatchPatchAvailabilityOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchOneTimeProductOfferAvailability(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewOneTimeProductOfferBatchPatchAvailabilityPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchOneTimeProductOfferAvailability(cmd.Context(), publisher, patchOptions)
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
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&availability, "availability", nil, "Availability patch as productId/purchaseOptionId/offerId/REGION:available|noLongerAvailable; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by oneTimeProductOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer availability batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer availability batch patch without calling Google Play")
	return cmd
}

func newOneTimeProductOffersBatchStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.OneTimeProductOfferStateAction) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offers           []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-" + action.String(),
		Short: oneTimeProductOfferBatchStateShort(action),
		Long: oneTimeProductOfferBatchStateLong(action) + " Omit parent IDs to infer the narrowest valid parent path from --offer values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferBatchMutationRequests(offers)
			if err != nil {
				return err
			}
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, requests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.OneTimeProductOfferBatchStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := updateOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.BatchUpdateOneTimeProductOfferStates(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchUpdateOneTimeProductOfferStates(cmd.Context(), publisher, updateOptions)
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
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, oneTimeProductOfferBatchStateConfirmHelp(action))
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, oneTimeProductOfferBatchStateDryRunHelp(action))
	return cmd
}

func oneTimeProductOfferBatchStateShort(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Cancel multiple pre-order one-time product offers and pending orders"
	}
	return action.String() + " multiple one-time product offers"
}

func oneTimeProductOfferBatchStateLong(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Cancel multiple pre-order one-time product offers. Google Play cancels pending orders for the cancelled offers."
	}
	return string(action) + " multiple one-time product offers."
}

func oneTimeProductOfferBatchStateConfirmHelp(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Cancel the pre-order offers and their pending orders"
	}
	return "Apply the one-time product offer batch state update"
}

func oneTimeProductOfferBatchStateDryRunHelp(action play.OneTimeProductOfferStateAction) string {
	if action == play.OneTimeProductOfferStateActionCancel {
		return "Print the planned pre-order offer cancellation without calling Google Play"
	}
	return "Print the planned one-time product offer batch state update without calling Google Play"
}

func newOneTimeProductOffersBatchDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offers           []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Delete multiple one-time product offers",
		Long: "Delete multiple one-time product offers. Omit parent IDs to infer the narrowest valid parent path from --offer values. " +
			"Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferBatchMutationRequests(offers)
			if err != nil {
				return err
			}
			resolvedProductID, resolvedPurchaseOptionID := inferOneTimeProductOfferBatchParent(productID, purchaseOptionID, requests)
			typedProductID, err := play.NewOneTimeProductOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedPurchaseOptionID, err := play.NewOneTimeProductOfferListPurchaseOptionID(resolvedPurchaseOptionID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			deleteOptions := play.OneTimeProductOfferBatchDeleteOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.BatchDeleteOneTimeProductOffers(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchDeleteOneTimeProductOffers(cmd.Context(), publisher, deleteOptions)
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
		"Parent one-time product ID, or - for offers across products; inferred when omitted",
		"Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to delete as productId/purchaseOptionId/offerId; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer batch deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer batch deletion without calling Google Play")
	return cmd
}

func newOneTimeProductOffersBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offers           []string
	)

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple one-time product offers",
		Long:  "Get multiple one-time product offers. Use --product-id - when the batch spans products, and --purchase-option-id - when it spans purchase options. Concrete parent IDs must match every --offer value.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedPurchaseOptionID, err := parseOneTimeProductOfferListParent(*packageName, productID, purchaseOptionID)
			if err != nil {
				return err
			}
			requests, err := parseOneTimeProductOfferBatchRequests(offers)
			if err != nil {
				return err
			}
			batchOptions := play.OneTimeProductOfferBatchGetOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				Requests:         requests,
			}
			if err := batchOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchGetOneTimeProductOffers(cmd.Context(), publisher, batchOptions)
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
		"Parent one-time product ID, or - for offers across products",
		"Parent one-time product purchase option ID, or - for offers across purchase options",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to fetch as productId/purchaseOptionId/offerId; repeatable, up to 100")
	return cmd
}

func newOneTimeProductOffersStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.OneTimeProductOfferStateAction) *cobra.Command {
	var (
		productID        string
		purchaseOptionID string
		offerID          string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action.String(),
		Short: string(action) + " a one-time product offer",
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
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.OneTimeProductOfferStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				PurchaseOptionID: typedPurchaseOptionID,
				OfferID:          typedOfferID,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.UpdateOneTimeProductOfferState(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewOneTimeProductOfferStateUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateOneTimeProductOfferState(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addOneTimeProductOfferParentFlags(cmd, &productID, &purchaseOptionID, "Parent one-time product ID", "Parent one-time product purchase option ID")
	cmd.Flags().StringVar(&offerID, "offer-id", "", "One-time product offer ID")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the one-time product offer state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned one-time product offer state update without calling Google Play")
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

func parseOneTimeProductOfferBatchRequests(values []string) ([]play.OneTimeProductOfferBatchGetRequest, error) {
	requests := make([]play.OneTimeProductOfferBatchGetRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewOneTimeProductOfferBatchGetRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferBatchMutationRequests(values []string) ([]play.OneTimeProductOfferBatchMutationRequest, error) {
	requests := make([]play.OneTimeProductOfferBatchMutationRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewOneTimeProductOfferBatchMutationRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferAvailabilityPatches(values []string) ([]play.OneTimeProductOfferAvailabilityPatchRequest, error) {
	requests := make([]play.OneTimeProductOfferAvailabilityPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseOneTimeProductOfferAvailabilityPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseOneTimeProductOfferAvailabilityPatch(value string) (play.OneTimeProductOfferAvailabilityPatchRequest, error) {
	path, rawAvailability, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, errOneTimeProductOfferAvailabilityFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, errOneTimeProductOfferAvailabilityFormat()
	}
	productID, err := play.NewOneTimeProductID(parts[0])
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	purchaseOptionID, err := play.NewOneTimeProductPurchaseOptionID(parts[1])
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	offerID, err := play.NewOneTimeProductOfferID(parts[2])
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	availability, err := parseOneTimeProductOfferAvailabilityValue(rawAvailability)
	if err != nil {
		return play.OneTimeProductOfferAvailabilityPatchRequest{}, err
	}
	return play.OneTimeProductOfferAvailabilityPatchRequest{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
		RegionCode:       strings.ToUpper(strings.TrimSpace(parts[3])),
		Availability:     availability,
	}, nil
}

func parseOneTimeProductOfferAvailabilityValue(value string) (play.OneTimeProductOfferAvailability, error) {
	switch strings.TrimSpace(value) {
	case play.OneTimeProductOfferAvailabilityAvailable.String():
		return play.OneTimeProductOfferAvailabilityAvailable, nil
	case play.OneTimeProductOfferAvailabilityNoLongerAvailable.String():
		return play.OneTimeProductOfferAvailabilityNoLongerAvailable, nil
	default:
		return "", errOneTimeProductOfferAvailabilityFormat()
	}
}

func errOneTimeProductOfferAvailabilityFormat() error {
	return fmt.Errorf("one-time product offer availability must use productId/purchaseOptionId/offerId/REGION:available|noLongerAvailable")
}

func oneTimeProductOfferAvailabilityPatchesToMutationRequests(requests []play.OneTimeProductOfferAvailabilityPatchRequest) []play.OneTimeProductOfferBatchMutationRequest {
	mutations := make([]play.OneTimeProductOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.OneTimeProductOfferBatchMutationRequest{
			ProductID:        request.ProductID,
			PurchaseOptionID: request.PurchaseOptionID,
			OfferID:          request.OfferID,
		})
	}
	return mutations
}

func inferOneTimeProductOfferBatchParent(productID string, purchaseOptionID string, requests []play.OneTimeProductOfferBatchMutationRequest) (string, string) {
	if len(requests) == 0 {
		return productID, purchaseOptionID
	}
	firstProductID := requests[0].ProductID.String()
	firstPurchaseOptionID := requests[0].PurchaseOptionID.String()
	inferredProductID := firstProductID
	inferredPurchaseOptionID := firstPurchaseOptionID
	for _, request := range requests[1:] {
		if request.ProductID.String() != firstProductID {
			inferredProductID = play.OneTimeProductOfferWildcardID
			inferredPurchaseOptionID = play.OneTimeProductOfferWildcardID
			break
		}
		if request.PurchaseOptionID.String() != firstPurchaseOptionID {
			inferredPurchaseOptionID = play.OneTimeProductOfferWildcardID
		}
	}
	if productID != "" {
		inferredProductID = productID
	}
	if purchaseOptionID != "" {
		inferredPurchaseOptionID = purchaseOptionID
	}
	return inferredProductID, inferredPurchaseOptionID
}
