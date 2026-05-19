package play

import (
	"context"
	"fmt"
	"sort"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListSubscriptionOffers(ctx context.Context, options SubscriptionOfferListOptions) (SubscriptionOfferListResult, error) {
	call := p.service.Monetization.Subscriptions.BasePlans.Offers.List(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
	).Context(ctx)
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	response, err := call.Do()
	if err != nil {
		return SubscriptionOfferListResult{}, fmt.Errorf("list subscription offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) GetSubscriptionOffer(ctx context.Context, packageName PackageName, productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, offerID SubscriptionOfferID) (SubscriptionOffer, error) {
	offer, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
		packageName.String(),
		productID.String(),
		basePlanID.String(),
		offerID.String(),
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOffer{}, fmt.Errorf("get subscription offer %s for %s/%s/%s: %w", offerID, packageName, productID, basePlanID, err)
	}
	return subscriptionOfferFromAPI(offer), nil
}

func (p *GooglePublisher) CreateSubscriptionOffer(ctx context.Context, options SubscriptionOfferCreateOptions) (SubscriptionOffer, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOffer{}, err
	}
	offer, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Create(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		subscriptionOfferCreateToAPI(options),
	).
		OfferId(options.OfferID.String()).
		RegionsVersionVersion(options.RegionsVersion).
		Context(ctx).
		Do()
	if err != nil {
		return SubscriptionOffer{}, fmt.Errorf("create subscription offer %s for %s/%s/%s: %w", options.OfferID, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferFromAPI(offer), nil
}

func (p *GooglePublisher) DeleteSubscriptionOffer(ctx context.Context, options SubscriptionOfferDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Monetization.Subscriptions.BasePlans.Offers.Delete(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		options.OfferID.String(),
	).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete subscription offer %s for %s/%s/%s: %w", options.OfferID, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return nil
}

func (p *GooglePublisher) UpdateSubscriptionOfferState(ctx context.Context, options SubscriptionOfferStateUpdateOptions) (SubscriptionOffer, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOffer{}, err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	var (
		offer *androidpublisher.SubscriptionOffer
		err   error
	)
	switch options.Action {
	case SubscriptionOfferStateActionActivate:
		offer, err = p.service.Monetization.Subscriptions.BasePlans.Offers.Activate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			options.OfferID.String(),
			&androidpublisher.ActivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case SubscriptionOfferStateActionDeactivate:
		offer, err = p.service.Monetization.Subscriptions.BasePlans.Offers.Deactivate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			options.OfferID.String(),
			&androidpublisher.DeactivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	default:
		return SubscriptionOffer{}, fmt.Errorf("unsupported subscription offer state action %q", options.Action)
	}
	if err != nil {
		return SubscriptionOffer{}, fmt.Errorf("%s subscription offer %s for %s/%s/%s: %w", options.Action, options.OfferID, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferFromAPI(offer), nil
}

func (p *GooglePublisher) BatchUpdateSubscriptionOfferStates(ctx context.Context, options SubscriptionOfferBatchStateUpdateOptions) (SubscriptionOfferBatchStateUpdateResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchStateUpdateResult{}, err
	}
	request := &androidpublisher.BatchUpdateSubscriptionOfferStatesRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferStateRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, subscriptionOfferStateRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchStateUpdateResult{}, fmt.Errorf("batch %s subscription offers for %s/%s/%s: %w", options.Action, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	offers := subscriptionOffersFromBatchStateUpdateResponse(options, response)
	return SubscriptionOfferBatchStateUpdateResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferBatchMutationRequest(nil), options.Requests...),
		Action:      options.Action,
		Applied:     true,
		Offers:      offers,
	}, nil
}

func (p *GooglePublisher) BatchPatchSubscriptionOfferAvailability(ctx context.Context, options SubscriptionOfferBatchPatchAvailabilityOptions) (SubscriptionOfferBatchPatchAvailabilityResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchAvailabilityResult{}, err
	}
	requestsByOffer := subscriptionOfferAvailabilityPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchAvailabilityResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before availability patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		mergedRegionalConfigs := mergeSubscriptionOfferAvailabilityPatches(subscriptionOfferRegionalConfigsFromAPI(current.RegionalConfigs), offerPatch.Requests)
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName:     options.PackageName.String(),
				ProductId:       offerPatch.ProductID.String(),
				BasePlanId:      offerPatch.BasePlanID.String(),
				OfferId:         offerPatch.OfferID.String(),
				RegionalConfigs: subscriptionOfferRegionalConfigsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       subscriptionOfferAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchAvailabilityResult{}, fmt.Errorf("batch patch subscription offer availability for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchAvailabilityResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferAvailabilityPatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdateResponse(options, response),
	}, nil
}

func (p *GooglePublisher) BatchPatchSubscriptionOfferPhaseRelativeDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) (SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, err
	}
	requestsByOffer := subscriptionOfferPhaseRelativeDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before phase relative discount patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		phases := subscriptionOfferPhasesFromAPI(current.Phases)
		mergedPhases := mergeSubscriptionOfferPhaseRelativeDiscountPatches(phases, offerPatch.Requests)
		if mergedPhases == nil {
			return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, fmt.Errorf("subscription offer phase relative discount patch for %s/%s/%s references a phase or region that is not already configured on the offer", offerPatch.ProductID, offerPatch.BasePlanID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName: options.PackageName.String(),
				ProductId:   offerPatch.ProductID.String(),
				BasePlanId:  offerPatch.BasePlanID.String(),
				OfferId:     offerPatch.OfferID.String(),
				Phases:      subscriptionOfferPhasesToAPI(mergedPhases),
			},
			UpdateMask:       subscriptionOfferPhasesUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, fmt.Errorf("batch patch subscription offer phase relative discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferPhaseRelativeDiscountPatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdatePhaseRelativeDiscountResponse(options, response),
	}, nil
}

func (p *GooglePublisher) BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) (SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, err
	}
	requestsByOffer := subscriptionOfferPhaseAbsoluteDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before phase absolute discount patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		phases := subscriptionOfferPhasesFromAPI(current.Phases)
		mergedPhases := mergeSubscriptionOfferPhaseAbsoluteDiscountPatches(phases, offerPatch.Requests)
		if mergedPhases == nil {
			return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, fmt.Errorf("subscription offer phase absolute discount patch for %s/%s/%s references a phase or region that is not already configured on the offer", offerPatch.ProductID, offerPatch.BasePlanID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName: options.PackageName.String(),
				ProductId:   offerPatch.ProductID.String(),
				BasePlanId:  offerPatch.BasePlanID.String(),
				OfferId:     offerPatch.OfferID.String(),
				Phases:      subscriptionOfferPhasesToAPI(mergedPhases),
			},
			UpdateMask:       subscriptionOfferPhasesUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, fmt.Errorf("batch patch subscription offer phase absolute discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferPhaseAbsoluteDiscountPatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdatePhaseAbsoluteDiscountResponse(options, response),
	}, nil
}

func (p *GooglePublisher) BatchPatchSubscriptionOfferPhasePrices(ctx context.Context, options SubscriptionOfferBatchPatchPhasePricesOptions) (SubscriptionOfferBatchPatchPhasePricesResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchPhasePricesResult{}, err
	}
	requestsByOffer := subscriptionOfferPhasePricePatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchPhasePricesResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before phase price patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		phases := subscriptionOfferPhasesFromAPI(current.Phases)
		mergedPhases := mergeSubscriptionOfferPhasePricePatches(phases, offerPatch.Requests)
		if mergedPhases == nil {
			return SubscriptionOfferBatchPatchPhasePricesResult{}, fmt.Errorf("subscription offer phase price patch for %s/%s/%s references a phase or region that is not already configured on the offer", offerPatch.ProductID, offerPatch.BasePlanID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName: options.PackageName.String(),
				ProductId:   offerPatch.ProductID.String(),
				BasePlanId:  offerPatch.BasePlanID.String(),
				OfferId:     offerPatch.OfferID.String(),
				Phases:      subscriptionOfferPhasesToAPI(mergedPhases),
			},
			UpdateMask:       subscriptionOfferPhasesUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchPhasePricesResult{}, fmt.Errorf("batch patch subscription offer phase prices for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchPhasePricesResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferPhasePricePatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdatePhasePriceResponse(options, response),
	}, nil
}

func (p *GooglePublisher) BatchPatchSubscriptionOfferPhaseFree(ctx context.Context, options SubscriptionOfferBatchPatchPhaseFreeOptions) (SubscriptionOfferBatchPatchPhaseFreeResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchPhaseFreeResult{}, err
	}
	requestsByOffer := subscriptionOfferPhaseFreePatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchPhaseFreeResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before phase free patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		phases := subscriptionOfferPhasesFromAPI(current.Phases)
		mergedPhases := mergeSubscriptionOfferPhaseFreePatches(phases, offerPatch.Requests)
		if mergedPhases == nil {
			return SubscriptionOfferBatchPatchPhaseFreeResult{}, fmt.Errorf("subscription offer phase free patch for %s/%s/%s references a phase or region that is not already configured on the offer", offerPatch.ProductID, offerPatch.BasePlanID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName: options.PackageName.String(),
				ProductId:   offerPatch.ProductID.String(),
				BasePlanId:  offerPatch.BasePlanID.String(),
				OfferId:     offerPatch.OfferID.String(),
				Phases:      subscriptionOfferPhasesToAPI(mergedPhases),
			},
			UpdateMask:       subscriptionOfferPhasesUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseFreeResult{}, fmt.Errorf("batch patch subscription offer phase free for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchPhaseFreeResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferPhaseFreePatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdatePhaseFreeResponse(options, response),
	}, nil
}

type subscriptionOfferAvailabilityPatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferAvailabilityPatchRequest
}

type subscriptionOfferPhaseRelativeDiscountPatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferPhaseRelativeDiscountPatchRequest
}

type subscriptionOfferPhaseAbsoluteDiscountPatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest
}

type subscriptionOfferPhasePricePatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferPhasePricePatchRequest
}

type subscriptionOfferPhaseFreePatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferPhaseFreePatchRequest
}

func subscriptionOfferAvailabilityPatchRequestsByOffer(requests []SubscriptionOfferAvailabilityPatchRequest) []subscriptionOfferAvailabilityPatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferAvailabilityPatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferAvailabilityPatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOfferPhaseRelativeDiscountPatchRequestsByOffer(requests []SubscriptionOfferPhaseRelativeDiscountPatchRequest) []subscriptionOfferPhaseRelativeDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferPhaseRelativeDiscountPatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferPhaseRelativeDiscountPatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOfferPhaseAbsoluteDiscountPatchRequestsByOffer(requests []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) []subscriptionOfferPhaseAbsoluteDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferPhaseAbsoluteDiscountPatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferPhaseAbsoluteDiscountPatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOfferPhasePricePatchRequestsByOffer(requests []SubscriptionOfferPhasePricePatchRequest) []subscriptionOfferPhasePricePatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferPhasePricePatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferPhasePricePatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOfferPhaseFreePatchRequestsByOffer(requests []SubscriptionOfferPhaseFreePatchRequest) []subscriptionOfferPhaseFreePatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferPhaseFreePatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferPhaseFreePatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOffersFromBatchUpdateResponse(options SubscriptionOfferBatchPatchAvailabilityOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchUpdatePhaseRelativeDiscountResponse(options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchUpdatePhaseAbsoluteDiscountResponse(options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchUpdatePhasePriceResponse(options SubscriptionOfferBatchPatchPhasePricesOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchUpdatePhaseFreeResponse(options SubscriptionOfferBatchPatchPhaseFreeOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchStateUpdateResponse(options SubscriptionOfferBatchStateUpdateOptions, response *androidpublisher.BatchUpdateSubscriptionOfferStatesResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	for _, request := range options.Requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func (p *GooglePublisher) BatchGetSubscriptionOffers(ctx context.Context, options SubscriptionOfferBatchGetOptions) (SubscriptionOfferBatchGetResult, error) {
	request := &androidpublisher.BatchGetSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.GetSubscriptionOfferRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, &androidpublisher.GetSubscriptionOfferRequest{
			PackageName: options.PackageName.String(),
			ProductId:   item.ProductID.String(),
			BasePlanId:  item.BasePlanID.String(),
			OfferId:     item.OfferID.String(),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchGet(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchGetResult{}, fmt.Errorf("batch get subscription offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferBatchGetResultFromAPI(options, response), nil
}

func subscriptionOfferStateRequestToAPI(options SubscriptionOfferBatchStateUpdateOptions, item SubscriptionOfferBatchMutationRequest) *androidpublisher.UpdateSubscriptionOfferStateRequest {
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	switch options.Action {
	case SubscriptionOfferStateActionActivate:
		return &androidpublisher.UpdateSubscriptionOfferStateRequest{
			ActivateSubscriptionOfferRequest: &androidpublisher.ActivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				BasePlanId:       item.BasePlanID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case SubscriptionOfferStateActionDeactivate:
		return &androidpublisher.UpdateSubscriptionOfferStateRequest{
			DeactivateSubscriptionOfferRequest: &androidpublisher.DeactivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				BasePlanId:       item.BasePlanID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	default:
		return &androidpublisher.UpdateSubscriptionOfferStateRequest{}
	}
}

func subscriptionOfferBatchGetResultFromAPI(options SubscriptionOfferBatchGetOptions, response *androidpublisher.BatchGetSubscriptionOffersResponse) SubscriptionOfferBatchGetResult {
	result := SubscriptionOfferBatchGetResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Offers:      []SubscriptionOffer{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	for _, request := range options.Requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			result.Offers = append(result.Offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	result.Offers = append(result.Offers, extras...)
	return result
}

func subscriptionOfferKey(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, offerID SubscriptionOfferID) string {
	return productID.String() + "/" + basePlanID.String() + "/" + offerID.String()
}

func subscriptionOfferListResultFromAPI(options SubscriptionOfferListOptions, response *androidpublisher.ListSubscriptionOffersResponse) SubscriptionOfferListResult {
	result := SubscriptionOfferListResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Offers:      []SubscriptionOffer{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiOffer := range response.SubscriptionOffers {
		result.Offers = append(result.Offers, subscriptionOfferFromAPI(apiOffer))
	}
	return result
}

func subscriptionOfferFromAPI(apiOffer *androidpublisher.SubscriptionOffer) SubscriptionOffer {
	if apiOffer == nil {
		return SubscriptionOffer{RegionalConfigs: []SubscriptionOfferRegionalConfig{}, Phases: []SubscriptionOfferPhase{}}
	}
	return SubscriptionOffer{
		PackageName:        PackageName(apiOffer.PackageName),
		ProductID:          SubscriptionProductID(apiOffer.ProductId),
		BasePlanID:         SubscriptionBasePlanID(apiOffer.BasePlanId),
		OfferID:            SubscriptionOfferID(apiOffer.OfferId),
		State:              SubscriptionOfferState(apiOffer.State),
		OfferTags:          offerTagsFromAPI(apiOffer.OfferTags),
		RegionalConfigs:    subscriptionOfferRegionalConfigsFromAPI(apiOffer.RegionalConfigs),
		OtherRegionsConfig: subscriptionOfferOtherRegionsConfigFromAPI(apiOffer.OtherRegionsConfig),
		Phases:             subscriptionOfferPhasesFromAPI(apiOffer.Phases),
		Targeting:          subscriptionOfferTargetingFromAPI(apiOffer.Targeting),
	}
}

func subscriptionOfferCreateToAPI(options SubscriptionOfferCreateOptions) *androidpublisher.SubscriptionOffer {
	offer := subscriptionOfferCreateDesiredOffer(options)
	return &androidpublisher.SubscriptionOffer{
		PackageName:        offer.PackageName.String(),
		ProductId:          offer.ProductID.String(),
		BasePlanId:         offer.BasePlanID.String(),
		OfferId:            offer.OfferID.String(),
		OfferTags:          offerTagsToAPI(offer.OfferTags),
		RegionalConfigs:    subscriptionOfferRegionalConfigsToAPI(offer.RegionalConfigs),
		OtherRegionsConfig: subscriptionOfferOtherRegionsConfigToAPI(offer.OtherRegionsConfig),
		Phases:             subscriptionOfferPhasesToAPI(offer.Phases),
		Targeting:          subscriptionOfferTargetingToAPI(offer.Targeting),
	}
}

func subscriptionOfferRegionalConfigsFromAPI(apiConfigs []*androidpublisher.RegionalSubscriptionOfferConfig) []SubscriptionOfferRegionalConfig {
	configs := make([]SubscriptionOfferRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		configs = append(configs, SubscriptionOfferRegionalConfig{
			RegionCode:                apiConfig.RegionCode,
			NewSubscriberAvailability: apiConfig.NewSubscriberAvailability,
		})
	}
	return configs
}

func subscriptionOfferRegionalConfigsToAPI(configs []SubscriptionOfferRegionalConfig) []*androidpublisher.RegionalSubscriptionOfferConfig {
	apiConfigs := make([]*androidpublisher.RegionalSubscriptionOfferConfig, 0, len(configs))
	for _, config := range configs {
		apiConfig := &androidpublisher.RegionalSubscriptionOfferConfig{
			RegionCode:                config.RegionCode,
			NewSubscriberAvailability: config.NewSubscriberAvailability,
		}
		apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "NewSubscriberAvailability")
		apiConfigs = append(apiConfigs, apiConfig)
	}
	return apiConfigs
}

func mergeSubscriptionOfferAvailabilityPatches(current []SubscriptionOfferRegionalConfig, patches []SubscriptionOfferAvailabilityPatchRequest) []SubscriptionOfferRegionalConfig {
	merged := append([]SubscriptionOfferRegionalConfig(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferAvailabilityPatch(merged, patch)
	}
	return merged
}

func mergeSubscriptionOfferAvailabilityPatch(current []SubscriptionOfferRegionalConfig, patch SubscriptionOfferAvailabilityPatchRequest) []SubscriptionOfferRegionalConfig {
	merged := make([]SubscriptionOfferRegionalConfig, 0, len(current)+1)
	replaced := false
	for _, config := range current {
		if config.RegionCode == patch.RegionCode {
			config.NewSubscriberAvailability = patch.Availability
			merged = append(merged, config)
			replaced = true
			continue
		}
		merged = append(merged, config)
	}
	if !replaced {
		merged = append(merged, SubscriptionOfferRegionalConfig{
			RegionCode:                patch.RegionCode,
			NewSubscriberAvailability: patch.Availability,
		})
	}
	return merged
}

func subscriptionOfferOtherRegionsConfigFromAPI(apiConfig *androidpublisher.OtherRegionsSubscriptionOfferConfig) *SubscriptionOfferOtherRegionsConfig {
	if apiConfig == nil {
		return nil
	}
	return &SubscriptionOfferOtherRegionsConfig{NewSubscriberAvailability: apiConfig.OtherRegionsNewSubscriberAvailability}
}

func subscriptionOfferOtherRegionsConfigToAPI(config *SubscriptionOfferOtherRegionsConfig) *androidpublisher.OtherRegionsSubscriptionOfferConfig {
	if config == nil {
		return nil
	}
	apiConfig := &androidpublisher.OtherRegionsSubscriptionOfferConfig{
		OtherRegionsNewSubscriberAvailability: config.NewSubscriberAvailability,
	}
	apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "OtherRegionsNewSubscriberAvailability")
	return apiConfig
}

func subscriptionOfferPhasesFromAPI(apiPhases []*androidpublisher.SubscriptionOfferPhase) []SubscriptionOfferPhase {
	phases := make([]SubscriptionOfferPhase, 0, len(apiPhases))
	for _, apiPhase := range apiPhases {
		if apiPhase == nil {
			continue
		}
		phases = append(phases, SubscriptionOfferPhase{
			Duration:           apiPhase.Duration,
			RecurrenceCount:    apiPhase.RecurrenceCount,
			RegionalConfigs:    subscriptionOfferPhaseRegionalConfigsFromAPI(apiPhase.RegionalConfigs),
			OtherRegionsConfig: subscriptionOfferPhaseOtherRegionsConfigFromAPI(apiPhase.OtherRegionsConfig),
		})
	}
	return phases
}

func subscriptionOfferPhaseRegionalConfigsFromAPI(apiConfigs []*androidpublisher.RegionalSubscriptionOfferPhaseConfig) []SubscriptionOfferPhaseRegionalConfig {
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		configs = append(configs, SubscriptionOfferPhaseRegionalConfig{
			RegionCode:       apiConfig.RegionCode,
			Price:            moneyFromAPI(apiConfig.Price),
			AbsoluteDiscount: moneyFromAPI(apiConfig.AbsoluteDiscount),
			RelativeDiscount: apiConfig.RelativeDiscount,
			Free:             apiConfig.Free != nil,
		})
	}
	return configs
}

func subscriptionOfferPhasesToAPI(phases []SubscriptionOfferPhase) []*androidpublisher.SubscriptionOfferPhase {
	apiPhases := make([]*androidpublisher.SubscriptionOfferPhase, 0, len(phases))
	for _, phase := range phases {
		apiPhases = append(apiPhases, &androidpublisher.SubscriptionOfferPhase{
			Duration:           phase.Duration,
			RecurrenceCount:    phase.RecurrenceCount,
			RegionalConfigs:    subscriptionOfferPhaseRegionalConfigsToAPI(phase.RegionalConfigs),
			OtherRegionsConfig: subscriptionOfferPhaseOtherRegionsConfigToAPI(phase.OtherRegionsConfig),
		})
	}
	return apiPhases
}

func subscriptionOfferPhaseRegionalConfigsToAPI(configs []SubscriptionOfferPhaseRegionalConfig) []*androidpublisher.RegionalSubscriptionOfferPhaseConfig {
	apiConfigs := make([]*androidpublisher.RegionalSubscriptionOfferPhaseConfig, 0, len(configs))
	for _, config := range configs {
		apiConfig := &androidpublisher.RegionalSubscriptionOfferPhaseConfig{
			RegionCode:       config.RegionCode,
			Price:            moneyToAPI(config.Price),
			AbsoluteDiscount: moneyToAPI(config.AbsoluteDiscount),
			RelativeDiscount: config.RelativeDiscount,
		}
		if config.Free {
			apiConfig.Free = &androidpublisher.RegionalSubscriptionOfferPhaseFreePriceOverride{}
		}
		if config.RelativeDiscount != 0 {
			apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "RelativeDiscount")
		}
		apiConfigs = append(apiConfigs, apiConfig)
	}
	return apiConfigs
}

func subscriptionOfferPhaseOtherRegionsConfigToAPI(config *SubscriptionOfferPhaseOtherRegionsConfig) *androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig {
	if config == nil {
		return nil
	}
	apiConfig := &androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig{
		OtherRegionsPrices: otherRegionsSubscriptionOfferPhasePricesToAPI(config.OtherRegionsPrices),
		AbsoluteDiscounts:  otherRegionsSubscriptionOfferPhasePricesToAPI(config.AbsoluteDiscounts),
		RelativeDiscount:   config.RelativeDiscount,
	}
	if config.Free {
		apiConfig.Free = &androidpublisher.OtherRegionsSubscriptionOfferPhaseFreePriceOverride{}
	}
	if config.RelativeDiscount != 0 {
		apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "RelativeDiscount")
	}
	return apiConfig
}

func otherRegionsSubscriptionOfferPhasePricesToAPI(prices *SubscriptionOfferOtherRegionsPrices) *androidpublisher.OtherRegionsSubscriptionOfferPhasePrices {
	if prices == nil {
		return nil
	}
	return &androidpublisher.OtherRegionsSubscriptionOfferPhasePrices{
		UsdPrice: moneyToAPI(prices.USDPrice),
		EurPrice: moneyToAPI(prices.EURPrice),
	}
}

func mergeSubscriptionOfferPhaseRelativeDiscountPatches(current []SubscriptionOfferPhase, patches []SubscriptionOfferPhaseRelativeDiscountPatchRequest) []SubscriptionOfferPhase {
	merged := append([]SubscriptionOfferPhase(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferPhaseRelativeDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeSubscriptionOfferPhaseRelativeDiscountPatch(current []SubscriptionOfferPhase, patch SubscriptionOfferPhaseRelativeDiscountPatchRequest) []SubscriptionOfferPhase {
	if patch.PhaseIndex < 0 || patch.PhaseIndex >= len(current) {
		return nil
	}
	merged := append([]SubscriptionOfferPhase(nil), current...)
	phase := merged[patch.PhaseIndex]
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(phase.RegionalConfigs))
	replaced := false
	for _, config := range phase.RegionalConfigs {
		if config.RegionCode == patch.RegionCode {
			config.RelativeDiscount = patch.RelativeDiscount
			config.Price = nil
			config.AbsoluteDiscount = nil
			config.Free = false
			configs = append(configs, config)
			replaced = true
			continue
		}
		configs = append(configs, config)
	}
	if !replaced {
		return nil
	}
	phase.RegionalConfigs = configs
	merged[patch.PhaseIndex] = phase
	return merged
}

func mergeSubscriptionOfferPhaseAbsoluteDiscountPatches(current []SubscriptionOfferPhase, patches []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) []SubscriptionOfferPhase {
	merged := append([]SubscriptionOfferPhase(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferPhaseAbsoluteDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeSubscriptionOfferPhaseAbsoluteDiscountPatch(current []SubscriptionOfferPhase, patch SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) []SubscriptionOfferPhase {
	if patch.PhaseIndex < 0 || patch.PhaseIndex >= len(current) {
		return nil
	}
	merged := append([]SubscriptionOfferPhase(nil), current...)
	phase := merged[patch.PhaseIndex]
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(phase.RegionalConfigs))
	replaced := false
	for _, config := range phase.RegionalConfigs {
		if config.RegionCode == patch.RegionCode {
			absoluteDiscount := patch.AbsoluteDiscount
			config.AbsoluteDiscount = &absoluteDiscount
			config.Price = nil
			config.RelativeDiscount = 0
			config.Free = false
			configs = append(configs, config)
			replaced = true
			continue
		}
		configs = append(configs, config)
	}
	if !replaced {
		return nil
	}
	phase.RegionalConfigs = configs
	merged[patch.PhaseIndex] = phase
	return merged
}

func mergeSubscriptionOfferPhasePricePatches(current []SubscriptionOfferPhase, patches []SubscriptionOfferPhasePricePatchRequest) []SubscriptionOfferPhase {
	merged := append([]SubscriptionOfferPhase(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferPhasePricePatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeSubscriptionOfferPhasePricePatch(current []SubscriptionOfferPhase, patch SubscriptionOfferPhasePricePatchRequest) []SubscriptionOfferPhase {
	if patch.PhaseIndex < 0 || patch.PhaseIndex >= len(current) {
		return nil
	}
	merged := append([]SubscriptionOfferPhase(nil), current...)
	phase := merged[patch.PhaseIndex]
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(phase.RegionalConfigs))
	replaced := false
	for _, config := range phase.RegionalConfigs {
		if config.RegionCode == patch.RegionCode {
			price := patch.Price
			config.Price = &price
			config.AbsoluteDiscount = nil
			config.RelativeDiscount = 0
			config.Free = false
			configs = append(configs, config)
			replaced = true
			continue
		}
		configs = append(configs, config)
	}
	if !replaced {
		return nil
	}
	phase.RegionalConfigs = configs
	merged[patch.PhaseIndex] = phase
	return merged
}

func mergeSubscriptionOfferPhaseFreePatches(current []SubscriptionOfferPhase, patches []SubscriptionOfferPhaseFreePatchRequest) []SubscriptionOfferPhase {
	merged := append([]SubscriptionOfferPhase(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferPhaseFreePatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeSubscriptionOfferPhaseFreePatch(current []SubscriptionOfferPhase, patch SubscriptionOfferPhaseFreePatchRequest) []SubscriptionOfferPhase {
	if patch.PhaseIndex < 0 || patch.PhaseIndex >= len(current) {
		return nil
	}
	merged := append([]SubscriptionOfferPhase(nil), current...)
	phase := merged[patch.PhaseIndex]
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(phase.RegionalConfigs))
	replaced := false
	for _, config := range phase.RegionalConfigs {
		if config.RegionCode == patch.RegionCode {
			config.Free = true
			config.Price = nil
			config.AbsoluteDiscount = nil
			config.RelativeDiscount = 0
			configs = append(configs, config)
			replaced = true
			continue
		}
		configs = append(configs, config)
	}
	if !replaced {
		return nil
	}
	phase.RegionalConfigs = configs
	merged[patch.PhaseIndex] = phase
	return merged
}

func subscriptionOfferPhaseOtherRegionsConfigFromAPI(apiConfig *androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig) *SubscriptionOfferPhaseOtherRegionsConfig {
	if apiConfig == nil {
		return nil
	}
	return &SubscriptionOfferPhaseOtherRegionsConfig{
		OtherRegionsPrices: otherRegionsSubscriptionOfferPhasePricesFromAPI(apiConfig.OtherRegionsPrices),
		AbsoluteDiscounts:  otherRegionsSubscriptionOfferPhasePricesFromAPI(apiConfig.AbsoluteDiscounts),
		RelativeDiscount:   apiConfig.RelativeDiscount,
		Free:               apiConfig.Free != nil,
	}
}

func otherRegionsSubscriptionOfferPhasePricesFromAPI(apiPrices *androidpublisher.OtherRegionsSubscriptionOfferPhasePrices) *SubscriptionOfferOtherRegionsPrices {
	if apiPrices == nil {
		return nil
	}
	return &SubscriptionOfferOtherRegionsPrices{
		USDPrice: moneyFromAPI(apiPrices.UsdPrice),
		EURPrice: moneyFromAPI(apiPrices.EurPrice),
	}
}

func subscriptionOfferTargetingFromAPI(apiTargeting *androidpublisher.SubscriptionOfferTargeting) *SubscriptionOfferTargeting {
	if apiTargeting == nil {
		return nil
	}
	return &SubscriptionOfferTargeting{
		Acquisition: subscriptionOfferAcquisitionTargetingFromAPI(apiTargeting.AcquisitionRule),
		Upgrade:     subscriptionOfferUpgradeTargetingFromAPI(apiTargeting.UpgradeRule),
	}
}

func subscriptionOfferTargetingToAPI(targeting *SubscriptionOfferTargeting) *androidpublisher.SubscriptionOfferTargeting {
	if targeting == nil {
		return nil
	}
	return &androidpublisher.SubscriptionOfferTargeting{
		AcquisitionRule: subscriptionOfferAcquisitionTargetingToAPI(targeting.Acquisition),
		UpgradeRule:     subscriptionOfferUpgradeTargetingToAPI(targeting.Upgrade),
	}
}

func subscriptionOfferAcquisitionTargetingFromAPI(apiRule *androidpublisher.AcquisitionTargetingRule) *SubscriptionOfferAcquisitionTargeting {
	if apiRule == nil {
		return nil
	}
	return &SubscriptionOfferAcquisitionTargeting{Scope: subscriptionOfferTargetingScopeFromAPI(apiRule.Scope)}
}

func subscriptionOfferAcquisitionTargetingToAPI(rule *SubscriptionOfferAcquisitionTargeting) *androidpublisher.AcquisitionTargetingRule {
	if rule == nil {
		return nil
	}
	return &androidpublisher.AcquisitionTargetingRule{Scope: subscriptionOfferTargetingScopeToAPI(rule.Scope)}
}

func subscriptionOfferUpgradeTargetingFromAPI(apiRule *androidpublisher.UpgradeTargetingRule) *SubscriptionOfferUpgradeTargeting {
	if apiRule == nil {
		return nil
	}
	return &SubscriptionOfferUpgradeTargeting{
		Scope:                 subscriptionOfferTargetingScopeFromAPI(apiRule.Scope),
		BillingPeriodDuration: apiRule.BillingPeriodDuration,
		OncePerUser:           apiRule.OncePerUser,
	}
}

func subscriptionOfferUpgradeTargetingToAPI(rule *SubscriptionOfferUpgradeTargeting) *androidpublisher.UpgradeTargetingRule {
	if rule == nil {
		return nil
	}
	apiRule := &androidpublisher.UpgradeTargetingRule{
		Scope:                 subscriptionOfferTargetingScopeToAPI(rule.Scope),
		BillingPeriodDuration: rule.BillingPeriodDuration,
		OncePerUser:           rule.OncePerUser,
	}
	if rule.OncePerUser {
		apiRule.ForceSendFields = append(apiRule.ForceSendFields, "OncePerUser")
	}
	return apiRule
}

func subscriptionOfferTargetingScopeFromAPI(apiScope *androidpublisher.TargetingRuleScope) *SubscriptionOfferTargetingScope {
	if apiScope == nil {
		return nil
	}
	return &SubscriptionOfferTargetingScope{
		AnySubscriptionInApp:      apiScope.AnySubscriptionInApp != nil,
		ThisSubscription:          apiScope.ThisSubscription != nil,
		SpecificSubscriptionInApp: apiScope.SpecificSubscriptionInApp,
	}
}

func subscriptionOfferTargetingScopeToAPI(scope *SubscriptionOfferTargetingScope) *androidpublisher.TargetingRuleScope {
	if scope == nil {
		return nil
	}
	apiScope := &androidpublisher.TargetingRuleScope{
		SpecificSubscriptionInApp: scope.SpecificSubscriptionInApp,
	}
	if scope.AnySubscriptionInApp {
		apiScope.AnySubscriptionInApp = &androidpublisher.TargetingRuleScopeAnySubscriptionInApp{}
	}
	if scope.ThisSubscription {
		apiScope.ThisSubscription = &androidpublisher.TargetingRuleScopeThisSubscription{}
	}
	return apiScope
}
