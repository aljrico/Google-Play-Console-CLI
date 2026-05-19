package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newSubscriptionOffersCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "subscription-offers",
		Short: "Inspect Google Play subscription offers",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newSubscriptionOffersListCommand(out, options, &packageName),
		newSubscriptionOffersGetCommand(out, options, &packageName),
		newSubscriptionOffersBatchGetCommand(out, options, &packageName),
		newSubscriptionOffersDeleteCommand(out, options, &packageName),
		newSubscriptionOffersBatchPatchAvailabilityCommand(out, options, &packageName),
		newSubscriptionOffersBatchStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionActivate),
		newSubscriptionOffersBatchStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionDeactivate),
		newSubscriptionOffersStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionActivate),
		newSubscriptionOffersStateCommand(out, options, &packageName, play.SubscriptionOfferStateActionDeactivate),
	)
	return cmd
}

func newSubscriptionOffersBatchPatchAvailabilityCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		availability     []string
		regionsVersion   string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-patch-availability",
		Short: "Batch patch subscription offer regional availability",
		Long: "Batch patch subscription offer regional availability. Omit parent IDs to infer the narrowest valid parent path from --availability values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferAvailabilityPatches(availability)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchPatchAvailabilityOptions{PackageName: typedPackageName}.Validate()
			}
			mutationRequests := subscriptionOfferAvailabilityPatchesToMutationRequests(requests)
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, mutationRequests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			patchOptions := play.SubscriptionOfferBatchPatchAvailabilityOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				Requests:         requests,
				RegionsVersion:   regionsVersion,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.BatchPatchSubscriptionOfferAvailability(cmd.Context(), nil, patchOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferBatchPatchAvailabilityPlan(patchOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchPatchSubscriptionOfferAvailability(cmd.Context(), publisher, patchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&availability, "availability", nil, "Availability patch as productId/basePlanId/offerId/REGION:true|false; repeatable")
	cmd.Flags().StringVar(&regionsVersion, "regions-version", "", "Google Play regions version required by subscriptionOffers.batchUpdate")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer availability batch patch")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer availability batch patch without calling Google Play")
	return cmd
}

func parseSubscriptionOfferAvailabilityPatches(values []string) ([]play.SubscriptionOfferAvailabilityPatchRequest, error) {
	requests := make([]play.SubscriptionOfferAvailabilityPatchRequest, 0, len(values))
	for _, value := range values {
		request, err := parseSubscriptionOfferAvailabilityPatch(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferAvailabilityPatch(value string) (play.SubscriptionOfferAvailabilityPatchRequest, error) {
	path, rawAvailability, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, errSubscriptionOfferAvailabilityFormat()
	}
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, errSubscriptionOfferAvailabilityFormat()
	}
	productID, err := play.NewSubscriptionProductID(parts[0])
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	basePlanID, err := play.NewSubscriptionBasePlanID(parts[1])
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	offerID, err := play.NewSubscriptionOfferID(parts[2])
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	availability, err := parseSubscriptionOfferAvailabilityValue(rawAvailability)
	if err != nil {
		return play.SubscriptionOfferAvailabilityPatchRequest{}, err
	}
	return play.SubscriptionOfferAvailabilityPatchRequest{
		ProductID:    productID,
		BasePlanID:   basePlanID,
		OfferID:      offerID,
		RegionCode:   strings.ToUpper(strings.TrimSpace(parts[3])),
		Availability: availability,
	}, nil
}

func errSubscriptionOfferAvailabilityFormat() error {
	return fmt.Errorf("subscription offer availability must use productId/basePlanId/offerId/REGION:true|false")
}

func parseSubscriptionOfferAvailabilityValue(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errSubscriptionOfferAvailabilityFormat()
	}
}

func subscriptionOfferAvailabilityPatchesToMutationRequests(requests []play.SubscriptionOfferAvailabilityPatchRequest) []play.SubscriptionOfferBatchMutationRequest {
	mutations := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(requests))
	for _, request := range requests {
		mutations = append(mutations, play.SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	return mutations
}

func newSubscriptionOffersBatchStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.SubscriptionOfferStateAction) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		offers           []string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   "batch-" + action.String(),
		Short: action.String() + " multiple subscription offers",
		Long: string(action) + " multiple subscription offers. Omit parent IDs to infer the narrowest valid parent path from --offer values. " +
			"Use --product-id - when the batch spans products, and --base-plan-id - when it spans base plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferBatchMutationRequests(offers)
			if err != nil {
				return err
			}
			if len(requests) == 0 {
				return play.SubscriptionOfferBatchStateUpdateOptions{PackageName: typedPackageName}.Validate()
			}
			resolvedProductID, resolvedBasePlanID := inferSubscriptionOfferBatchParent(productID, basePlanID, requests)
			typedProductID, err := play.NewSubscriptionOfferListProductID(resolvedProductID)
			if err != nil {
				return err
			}
			typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(resolvedBasePlanID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.SubscriptionOfferBatchStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
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
				result, err := play.BatchUpdateSubscriptionOfferStates(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchUpdateSubscriptionOfferStates(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products; inferred when omitted",
		"Parent subscription base plan ID, or - for offers across base plans; inferred when omitted",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to update as productId/basePlanId/offerId; repeatable, up to 100")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer batch state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer batch state update without calling Google Play")
	return cmd
}

func newSubscriptionOffersStateCommand(out io.Writer, options *globalOptions, packageName *string, action play.SubscriptionOfferStateAction) *cobra.Command {
	var (
		productID        string
		basePlanID       string
		offerID          string
		latencyTolerance string
		confirm          bool
		dryRun           bool
	)

	cmd := &cobra.Command{
		Use:   action.String(),
		Short: string(action) + " a subscription offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferGetParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewSubscriptionOfferID(offerID)
			if err != nil {
				return err
			}
			typedLatencyTolerance, err := play.NewProductUpdateLatencyTolerance(latencyTolerance)
			if err != nil {
				return err
			}
			updateOptions := play.SubscriptionOfferStateUpdateOptions{
				PackageName:      typedPackageName,
				ProductID:        typedProductID,
				BasePlanID:       typedBasePlanID,
				OfferID:          typedOfferID,
				Action:           action,
				LatencyTolerance: typedLatencyTolerance,
				Confirm:          confirm,
				DryRun:           dryRun,
			}
			if dryRun {
				result, err := play.UpdateSubscriptionOfferState(cmd.Context(), nil, updateOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewSubscriptionOfferStateUpdatePlan(updateOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.UpdateSubscriptionOfferState(cmd.Context(), publisher, updateOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(cmd, &productID, &basePlanID, "Parent subscription product ID", "Parent subscription base plan ID")
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Subscription offer ID")
	cmd.Flags().StringVar(&latencyTolerance, "latency-tolerance", play.ProductUpdateLatencyToleranceSensitive.String(), "Propagation latency: latencySensitive or latencyTolerant")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer state update")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer state update without calling Google Play")
	return cmd
}

func newSubscriptionOffersBatchGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		offers     []string
	)

	cmd := &cobra.Command{
		Use:   "batch-get",
		Short: "Get multiple subscription offers",
		Long:  "Get multiple subscription offers. Use - for both --product-id and --base-plan-id when the batch spans products or base plans. Concrete parent IDs must match every --offer value.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferListParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			requests, err := parseSubscriptionOfferBatchRequests(offers)
			if err != nil {
				return err
			}
			batchOptions := play.SubscriptionOfferBatchGetOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				Requests:    requests,
			}
			if err := batchOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.BatchGetSubscriptionOffers(cmd.Context(), publisher, batchOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for offers across products",
		"Parent subscription base plan ID, or - for offers across base plans",
	)
	cmd.Flags().StringArrayVar(&offers, "offer", nil, "Offer to fetch as productId/basePlanId/offerId; repeatable, up to 100")
	return cmd
}

func newSubscriptionOffersListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		pageSize   int64
		pageToken  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subscription offers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferListParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			listOptions := play.SubscriptionOfferListOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
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
			result, err := play.ListSubscriptionOffers(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID, or - for all products",
		"Parent subscription base plan ID, or - for all base plans",
	)
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum offers to return, capped at 1000")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	return cmd
}

func newSubscriptionOffersGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		offerID    string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one subscription offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferGetParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewSubscriptionOfferID(offerID)
			if err != nil {
				return err
			}
			getOptions := play.SubscriptionOfferGetOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				OfferID:     typedOfferID,
			}
			if err := getOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			offer, err := play.GetSubscriptionOffer(cmd.Context(), publisher, getOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, offer)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID",
		"Parent subscription base plan ID",
	)
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Subscription offer ID")
	return cmd
}

func newSubscriptionOffersDeleteCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		productID  string
		basePlanID string
		offerID    string
		confirm    bool
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a draft subscription offer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, typedProductID, typedBasePlanID, err := parseSubscriptionOfferGetParent(*packageName, productID, basePlanID)
			if err != nil {
				return err
			}
			typedOfferID, err := play.NewSubscriptionOfferID(offerID)
			if err != nil {
				return err
			}
			deleteOptions := play.SubscriptionOfferDeleteOptions{
				PackageName: typedPackageName,
				ProductID:   typedProductID,
				BasePlanID:  typedBasePlanID,
				OfferID:     typedOfferID,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DeleteSubscriptionOffer(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteSubscriptionOffer(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	addSubscriptionOfferParentFlags(
		cmd,
		&productID,
		&basePlanID,
		"Parent subscription product ID",
		"Parent subscription base plan ID",
	)
	cmd.Flags().StringVar(&offerID, "offer-id", "", "Subscription offer ID")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the subscription offer deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned subscription offer deletion without calling Google Play")
	return cmd
}

func addSubscriptionOfferParentFlags(cmd *cobra.Command, productID *string, basePlanID *string, productIDDescription string, basePlanIDDescription string) {
	cmd.Flags().StringVar(productID, "product-id", "", productIDDescription)
	cmd.Flags().StringVar(basePlanID, "base-plan-id", "", basePlanIDDescription)
}

func parseSubscriptionOfferListParent(packageName string, productID string, basePlanID string) (play.PackageName, play.SubscriptionProductID, play.SubscriptionBasePlanID, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return "", "", "", err
	}
	typedProductID, err := play.NewSubscriptionOfferListProductID(productID)
	if err != nil {
		return "", "", "", err
	}
	typedBasePlanID, err := play.NewSubscriptionOfferListBasePlanID(basePlanID)
	if err != nil {
		return "", "", "", err
	}
	return typedPackageName, typedProductID, typedBasePlanID, nil
}

func parseSubscriptionOfferGetParent(packageName string, productID string, basePlanID string) (play.PackageName, play.SubscriptionProductID, play.SubscriptionBasePlanID, error) {
	typedPackageName, err := play.NewPackageName(packageName)
	if err != nil {
		return "", "", "", err
	}
	typedProductID, err := play.NewSubscriptionProductID(productID)
	if err != nil {
		return "", "", "", err
	}
	typedBasePlanID, err := play.NewSubscriptionBasePlanID(basePlanID)
	if err != nil {
		return "", "", "", err
	}
	return typedPackageName, typedProductID, typedBasePlanID, nil
}

func parseSubscriptionOfferBatchRequests(values []string) ([]play.SubscriptionOfferBatchGetRequest, error) {
	requests := make([]play.SubscriptionOfferBatchGetRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewSubscriptionOfferBatchGetRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func parseSubscriptionOfferBatchMutationRequests(values []string) ([]play.SubscriptionOfferBatchMutationRequest, error) {
	requests := make([]play.SubscriptionOfferBatchMutationRequest, 0, len(values))
	for _, value := range values {
		request, err := play.NewSubscriptionOfferBatchMutationRequest(value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func inferSubscriptionOfferBatchParent(productID string, basePlanID string, requests []play.SubscriptionOfferBatchMutationRequest) (string, string) {
	resolvedProductID := productID
	resolvedBasePlanID := basePlanID
	if len(requests) == 0 {
		return resolvedProductID, resolvedBasePlanID
	}
	firstProductID := requests[0].ProductID.String()
	firstBasePlanID := requests[0].BasePlanID.String()
	oneProduct := true
	oneBasePlan := true
	for _, request := range requests[1:] {
		if request.ProductID.String() != firstProductID {
			oneProduct = false
		}
		if request.BasePlanID.String() != firstBasePlanID {
			oneBasePlan = false
		}
	}
	if resolvedProductID == "" {
		if oneProduct {
			resolvedProductID = firstProductID
		} else {
			resolvedProductID = play.SubscriptionOfferWildcardID
		}
	}
	if resolvedBasePlanID == "" {
		if oneBasePlan {
			resolvedBasePlanID = firstBasePlanID
		} else {
			resolvedBasePlanID = play.SubscriptionOfferWildcardID
		}
	}
	return resolvedProductID, resolvedBasePlanID
}
