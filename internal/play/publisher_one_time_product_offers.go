package play

import (
	"context"
	"fmt"
	"sort"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error) {
	call := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.List(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
	).Context(ctx)
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	response, err := call.Do()
	if err != nil {
		return OneTimeProductOfferListResult{}, fmt.Errorf("list one-time product offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return oneTimeProductOfferListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error) {
	request := batchGetOneTimeProductOffersRequestToAPI(options.PackageName, []OneTimeProductOfferBatchGetRequest{{
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		OfferID:          options.OfferID,
	}})
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchGet(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOffer{}, fmt.Errorf("get one-time product offer %s for %s/%s/%s: %w", options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	if response == nil || len(response.OneTimeProductOffers) == 0 {
		return OneTimeProductOffer{}, fmt.Errorf("get one-time product offer %s for %s/%s/%s: empty response", options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID)
	}
	return oneTimeProductOfferFromAPI(response.OneTimeProductOffers[0]), nil
}

func (p *GooglePublisher) BatchGetOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchGetOptions) (OneTimeProductOfferBatchGetResult, error) {
	request := batchGetOneTimeProductOffersRequestToAPI(options.PackageName, options.Requests)
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchGet(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchGetResult{}, fmt.Errorf("batch get one-time product offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return oneTimeProductOfferBatchGetResultFromAPI(options, response), nil
}

func (p *GooglePublisher) CreateOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferCreateOptions) (OneTimeProductOffer, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOffer{}, err
	}
	exists, err := p.oneTimeProductOfferExists(ctx, options.PackageName, options.ProductID, options.PurchaseOptionID, options.OfferID)
	if err != nil {
		return OneTimeProductOffer{}, err
	}
	if exists {
		return OneTimeProductOffer{}, fmt.Errorf("one-time product offer %s for %s/%s/%s already exists; use batch-patch-* or state commands for updates", options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID)
	}
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: []*androidpublisher.UpdateOneTimeProductOfferRequest{{
			OneTimeProductOffer: oneTimeProductOfferToAPI(oneTimeProductOfferCreateDesiredOffer(options)),
			UpdateMask:          oneTimeProductOfferCreateUpdateMask,
			RegionsVersion:      &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			AllowMissing:        true,
			LatencyTolerance:    productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
			ForceSendFields:     []string{"AllowMissing"},
		}},
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOffer{}, fmt.Errorf("create one-time product offer %s for %s/%s/%s: %w", options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	offers := oneTimeProductOffersFromBatchUpdateResponse(response)
	if len(offers) == 0 {
		return OneTimeProductOffer{}, fmt.Errorf("create one-time product offer %s for %s/%s/%s returned no offers", options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID)
	}
	return offers[0], nil
}

func (p *GooglePublisher) BatchDeleteOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.BatchDeleteOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.DeleteOneTimeProductOfferRequest, 0, len(options.Requests)),
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, &androidpublisher.DeleteOneTimeProductOfferRequest{
			PackageName:      options.PackageName.String(),
			ProductId:        item.ProductID.String(),
			PurchaseOptionId: item.PurchaseOptionID.String(),
			OfferId:          item.OfferID.String(),
			LatencyTolerance: latencyTolerance,
		})
	}
	err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchDelete(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("batch delete one-time product offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return nil
}

func (p *GooglePublisher) BatchUpdateOneTimeProductOfferStates(ctx context.Context, options OneTimeProductOfferBatchStateUpdateOptions) (OneTimeProductOfferBatchStateUpdateResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchStateUpdateResult{}, err
	}
	request := &androidpublisher.BatchUpdateOneTimeProductOfferStatesRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferStateRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, oneTimeProductOfferStateRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchStateUpdateResult{}, fmt.Errorf("batch %s one-time product offers for %s/%s/%s: %w", options.Action, options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	offers := make([]OneTimeProductOffer, 0)
	if response != nil {
		offers = make([]OneTimeProductOffer, 0, len(response.OneTimeProductOffers))
		for _, offer := range response.OneTimeProductOffers {
			offers = append(offers, oneTimeProductOfferFromAPI(offer))
		}
	}
	return OneTimeProductOfferBatchStateUpdateResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferBatchMutationRequest(nil), options.Requests...),
		Action:           options.Action,
		Applied:          true,
		Offers:           offers,
	}, nil
}

func (p *GooglePublisher) BatchPatchOneTimeProductOfferAvailability(ctx context.Context, options OneTimeProductOfferBatchPatchAvailabilityOptions) (OneTimeProductOfferBatchPatchAvailabilityResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchPatchAvailabilityResult{}, err
	}
	requestsByOffer := oneTimeProductOfferAvailabilityPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.getOneTimeProductOfferForAvailabilityPatch(ctx, options.PackageName, offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		if err != nil {
			return OneTimeProductOfferBatchPatchAvailabilityResult{}, err
		}
		mergedRegionalConfigs := mergeOneTimeProductOfferAvailabilityPatches(oneTimeProductOfferRegionsFromAPI(current.RegionalPricingAndAvailabilityConfigs), offerPatch.Requests)
		if mergedRegionalConfigs == nil {
			return OneTimeProductOfferBatchPatchAvailabilityResult{}, fmt.Errorf("one-time product offer availability patch for %s/%s/%s references a region that is not already configured on the offer; availability-only patches cannot add regional price overrides", offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductOfferRequest{
			OneTimeProductOffer: &androidpublisher.OneTimeProductOffer{
				PackageName:                           options.PackageName.String(),
				ProductId:                             offerPatch.ProductID.String(),
				PurchaseOptionId:                      offerPatch.PurchaseOptionID.String(),
				OfferId:                               offerPatch.OfferID.String(),
				RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       oneTimeProductOfferAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchPatchAvailabilityResult{}, fmt.Errorf("batch patch one-time product offer availability for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return OneTimeProductOfferBatchPatchAvailabilityResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferAvailabilityPatchRequest(nil), options.Requests...),
		Applied:          true,
		Offers:           oneTimeProductOffersFromBatchUpdateResponse(response),
	}, nil
}

func (p *GooglePublisher) BatchPatchOneTimeProductOfferRelativeDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) (OneTimeProductOfferBatchPatchRelativeDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, err
	}
	requestsByOffer := oneTimeProductOfferRelativeDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.getOneTimeProductOfferForRegionalPatch(ctx, options.PackageName, offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID, "relative discount")
		if err != nil {
			return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, err
		}
		mergedRegionalConfigs := mergeOneTimeProductOfferRelativeDiscountPatches(oneTimeProductOfferRegionsFromAPI(current.RegionalPricingAndAvailabilityConfigs), offerPatch.Requests)
		if mergedRegionalConfigs == nil {
			return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, fmt.Errorf("one-time product offer relative discount patch for %s/%s/%s references a region that is not already configured on the offer; relative discount patches cannot add regional price overrides", offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductOfferRequest{
			OneTimeProductOffer: &androidpublisher.OneTimeProductOffer{
				PackageName:                           options.PackageName.String(),
				ProductId:                             offerPatch.ProductID.String(),
				PurchaseOptionId:                      offerPatch.PurchaseOptionID.String(),
				OfferId:                               offerPatch.OfferID.String(),
				RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       oneTimeProductOfferRegionalConfigsUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, fmt.Errorf("batch patch one-time product offer relative discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return OneTimeProductOfferBatchPatchRelativeDiscountsResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferRelativeDiscountPatchRequest(nil), options.Requests...),
		Applied:          true,
		Offers:           oneTimeProductOffersFromBatchUpdateResponse(response),
	}, nil
}

func (p *GooglePublisher) BatchPatchOneTimeProductOfferAbsoluteDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) (OneTimeProductOfferBatchPatchAbsoluteDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, err
	}
	requestsByOffer := oneTimeProductOfferAbsoluteDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.getOneTimeProductOfferForRegionalPatch(ctx, options.PackageName, offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID, "absolute discount")
		if err != nil {
			return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, err
		}
		mergedRegionalConfigs := mergeOneTimeProductOfferAbsoluteDiscountPatches(oneTimeProductOfferRegionsFromAPI(current.RegionalPricingAndAvailabilityConfigs), offerPatch.Requests)
		if mergedRegionalConfigs == nil {
			return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, fmt.Errorf("one-time product offer absolute discount patch for %s/%s/%s references a region that is not already configured on the offer; absolute discount patches cannot add regional price overrides", offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductOfferRequest{
			OneTimeProductOffer: &androidpublisher.OneTimeProductOffer{
				PackageName:                           options.PackageName.String(),
				ProductId:                             offerPatch.ProductID.String(),
				PurchaseOptionId:                      offerPatch.PurchaseOptionID.String(),
				OfferId:                               offerPatch.OfferID.String(),
				RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       oneTimeProductOfferRegionalConfigsUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, fmt.Errorf("batch patch one-time product offer absolute discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferAbsoluteDiscountPatchRequest(nil), options.Requests...),
		Applied:          true,
		Offers:           oneTimeProductOffersFromBatchUpdateResponse(response),
	}, nil
}

func (p *GooglePublisher) BatchPatchOneTimeProductOfferNoOverrides(ctx context.Context, options OneTimeProductOfferBatchPatchNoOverridesOptions) (OneTimeProductOfferBatchPatchNoOverridesResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchPatchNoOverridesResult{}, err
	}
	requestsByOffer := oneTimeProductOfferNoOverridePatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.getOneTimeProductOfferForRegionalPatch(ctx, options.PackageName, offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID, "no-override")
		if err != nil {
			return OneTimeProductOfferBatchPatchNoOverridesResult{}, err
		}
		mergedRegionalConfigs := mergeOneTimeProductOfferNoOverridePatches(oneTimeProductOfferRegionsFromAPI(current.RegionalPricingAndAvailabilityConfigs), offerPatch.Requests)
		if mergedRegionalConfigs == nil {
			return OneTimeProductOfferBatchPatchNoOverridesResult{}, fmt.Errorf("one-time product offer no-override patch for %s/%s/%s references a region that is not already configured on the offer; no-override patches cannot add regional price overrides", offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductOfferRequest{
			OneTimeProductOffer: &androidpublisher.OneTimeProductOffer{
				PackageName:                           options.PackageName.String(),
				ProductId:                             offerPatch.ProductID.String(),
				PurchaseOptionId:                      offerPatch.PurchaseOptionID.String(),
				OfferId:                               offerPatch.OfferID.String(),
				RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       oneTimeProductOfferRegionalConfigsUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchPatchNoOverridesResult{}, fmt.Errorf("batch patch one-time product offer no-overrides for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return OneTimeProductOfferBatchPatchNoOverridesResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferNoOverridePatchRequest(nil), options.Requests...),
		Applied:          true,
		Offers:           oneTimeProductOffersFromBatchUpdateResponse(response),
	}, nil
}

func (p *GooglePublisher) getOneTimeProductOfferForAvailabilityPatch(ctx context.Context, packageName PackageName, productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, offerID OneTimeProductOfferID) (*androidpublisher.OneTimeProductOffer, error) {
	return p.getOneTimeProductOfferForRegionalPatch(ctx, packageName, productID, purchaseOptionID, offerID, "availability")
}

func (p *GooglePublisher) getOneTimeProductOfferForRegionalPatch(ctx context.Context, packageName PackageName, productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, offerID OneTimeProductOfferID, patchKind string) (*androidpublisher.OneTimeProductOffer, error) {
	request := batchGetOneTimeProductOffersRequestToAPI(packageName, []OneTimeProductOfferBatchGetRequest{{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
	}})
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchGet(
		packageName.String(),
		productID.String(),
		purchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get one-time product offer %s for %s/%s/%s before %s patch: %w", offerID, packageName, productID, purchaseOptionID, patchKind, err)
	}
	if response == nil || len(response.OneTimeProductOffers) == 0 {
		return nil, fmt.Errorf("get one-time product offer %s for %s/%s/%s before %s patch: empty response", offerID, packageName, productID, purchaseOptionID, patchKind)
	}
	return response.OneTimeProductOffers[0], nil
}

func (p *GooglePublisher) oneTimeProductOfferExists(ctx context.Context, packageName PackageName, productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, offerID OneTimeProductOfferID) (bool, error) {
	request := batchGetOneTimeProductOffersRequestToAPI(packageName, []OneTimeProductOfferBatchGetRequest{{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
	}})
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchGet(
		packageName.String(),
		productID.String(),
		purchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return false, fmt.Errorf("check one-time product offer %s for %s/%s/%s before create: %w", offerID, packageName, productID, purchaseOptionID, err)
	}
	return response != nil && len(response.OneTimeProductOffers) > 0, nil
}

func (p *GooglePublisher) UpdateOneTimeProductOfferState(ctx context.Context, options OneTimeProductOfferStateUpdateOptions) (OneTimeProductOffer, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOffer{}, err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	var (
		offer *androidpublisher.OneTimeProductOffer
		err   error
	)
	switch options.Action {
	case OneTimeProductOfferStateActionActivate:
		offer, err = p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.Activate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.PurchaseOptionID.String(),
			options.OfferID.String(),
			&androidpublisher.ActivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case OneTimeProductOfferStateActionDeactivate:
		offer, err = p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.Deactivate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.PurchaseOptionID.String(),
			options.OfferID.String(),
			&androidpublisher.DeactivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case OneTimeProductOfferStateActionCancel:
		offer, err = p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.Cancel(
			options.PackageName.String(),
			options.ProductID.String(),
			options.PurchaseOptionID.String(),
			options.OfferID.String(),
			&androidpublisher.CancelOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	default:
		return OneTimeProductOffer{}, fmt.Errorf("unsupported one-time product offer state action %q", options.Action)
	}
	if err != nil {
		return OneTimeProductOffer{}, fmt.Errorf("%s one-time product offer %s for %s/%s/%s: %w", options.Action, options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return oneTimeProductOfferFromAPI(offer), nil
}

func oneTimeProductOfferStateRequestToAPI(options OneTimeProductOfferBatchStateUpdateOptions, item OneTimeProductOfferBatchMutationRequest) *androidpublisher.UpdateOneTimeProductOfferStateRequest {
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	switch options.Action {
	case OneTimeProductOfferStateActionActivate:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{
			ActivateOneTimeProductOfferRequest: &androidpublisher.ActivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				PurchaseOptionId: item.PurchaseOptionID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case OneTimeProductOfferStateActionDeactivate:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{
			DeactivateOneTimeProductOfferRequest: &androidpublisher.DeactivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				PurchaseOptionId: item.PurchaseOptionID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case OneTimeProductOfferStateActionCancel:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{
			CancelOneTimeProductOfferRequest: &androidpublisher.CancelOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				PurchaseOptionId: item.PurchaseOptionID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	default:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{}
	}
}

type oneTimeProductOfferAvailabilityPatchOffer struct {
	ProductID        OneTimeProductID
	PurchaseOptionID OneTimeProductPurchaseOptionID
	OfferID          OneTimeProductOfferID
	Requests         []OneTimeProductOfferAvailabilityPatchRequest
}

type oneTimeProductOfferRelativeDiscountPatchOffer struct {
	ProductID        OneTimeProductID
	PurchaseOptionID OneTimeProductPurchaseOptionID
	OfferID          OneTimeProductOfferID
	Requests         []OneTimeProductOfferRelativeDiscountPatchRequest
}

type oneTimeProductOfferAbsoluteDiscountPatchOffer struct {
	ProductID        OneTimeProductID
	PurchaseOptionID OneTimeProductPurchaseOptionID
	OfferID          OneTimeProductOfferID
	Requests         []OneTimeProductOfferAbsoluteDiscountPatchRequest
}

type oneTimeProductOfferNoOverridePatchOffer struct {
	ProductID        OneTimeProductID
	PurchaseOptionID OneTimeProductPurchaseOptionID
	OfferID          OneTimeProductOfferID
	Requests         []OneTimeProductOfferNoOverridePatchRequest
}

func oneTimeProductOfferAvailabilityPatchRequestsByOffer(requests []OneTimeProductOfferAvailabilityPatchRequest) []oneTimeProductOfferAvailabilityPatchOffer {
	byOffer := map[string]int{}
	offers := make([]oneTimeProductOfferAvailabilityPatchOffer, 0)
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, oneTimeProductOfferAvailabilityPatchOffer{
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func oneTimeProductOfferRelativeDiscountPatchRequestsByOffer(requests []OneTimeProductOfferRelativeDiscountPatchRequest) []oneTimeProductOfferRelativeDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]oneTimeProductOfferRelativeDiscountPatchOffer, 0)
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, oneTimeProductOfferRelativeDiscountPatchOffer{
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func oneTimeProductOfferAbsoluteDiscountPatchRequestsByOffer(requests []OneTimeProductOfferAbsoluteDiscountPatchRequest) []oneTimeProductOfferAbsoluteDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]oneTimeProductOfferAbsoluteDiscountPatchOffer, 0)
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, oneTimeProductOfferAbsoluteDiscountPatchOffer{
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func oneTimeProductOfferNoOverridePatchRequestsByOffer(requests []OneTimeProductOfferNoOverridePatchRequest) []oneTimeProductOfferNoOverridePatchOffer {
	byOffer := map[string]int{}
	offers := make([]oneTimeProductOfferNoOverridePatchOffer, 0)
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, oneTimeProductOfferNoOverridePatchOffer{
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func oneTimeProductOffersFromBatchUpdateResponse(response *androidpublisher.BatchUpdateOneTimeProductOffersResponse) []OneTimeProductOffer {
	if response == nil {
		return []OneTimeProductOffer{}
	}
	offers := make([]OneTimeProductOffer, 0, len(response.OneTimeProductOffers))
	for _, apiOffer := range response.OneTimeProductOffers {
		offers = append(offers, oneTimeProductOfferFromAPI(apiOffer))
	}
	return offers
}

func oneTimeProductOfferListResultFromAPI(options OneTimeProductOfferListOptions, response *androidpublisher.ListOneTimeProductOffersResponse) OneTimeProductOfferListResult {
	result := OneTimeProductOfferListResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Offers:           []OneTimeProductOffer{},
		Options:          options,
	}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiOffer := range response.OneTimeProductOffers {
		result.Offers = append(result.Offers, oneTimeProductOfferFromAPI(apiOffer))
	}
	return result
}

func oneTimeProductOfferBatchGetResultFromAPI(options OneTimeProductOfferBatchGetOptions, response *androidpublisher.BatchGetOneTimeProductOffersResponse) OneTimeProductOfferBatchGetResult {
	result := OneTimeProductOfferBatchGetResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Offers:           []OneTimeProductOffer{},
		Options:          options,
	}
	if response == nil {
		return result
	}
	byKey := make(map[string]OneTimeProductOffer, len(response.OneTimeProductOffers))
	extras := make([]OneTimeProductOffer, 0)
	for _, apiOffer := range response.OneTimeProductOffers {
		offer := oneTimeProductOfferFromAPI(apiOffer)
		key := oneTimeProductOfferKey(offer.ProductID, offer.PurchaseOptionID, offer.OfferID)
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
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			result.Offers = append(result.Offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return oneTimeProductOfferKey(extras[i].ProductID, extras[i].PurchaseOptionID, extras[i].OfferID) < oneTimeProductOfferKey(extras[j].ProductID, extras[j].PurchaseOptionID, extras[j].OfferID)
	})
	result.Offers = append(result.Offers, extras...)
	return result
}

func batchGetOneTimeProductOffersRequestToAPI(packageName PackageName, requests []OneTimeProductOfferBatchGetRequest) *androidpublisher.BatchGetOneTimeProductOffersRequest {
	apiRequest := &androidpublisher.BatchGetOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.GetOneTimeProductOfferRequest, 0, len(requests)),
	}
	for _, request := range requests {
		apiRequest.Requests = append(apiRequest.Requests, &androidpublisher.GetOneTimeProductOfferRequest{
			PackageName:      packageName.String(),
			ProductId:        request.ProductID.String(),
			PurchaseOptionId: request.PurchaseOptionID.String(),
			OfferId:          request.OfferID.String(),
		})
	}
	return apiRequest
}

func oneTimeProductOfferKey(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, offerID OneTimeProductOfferID) string {
	return productID.String() + "/" + purchaseOptionID.String() + "/" + offerID.String()
}

func oneTimeProductOfferFromAPI(apiOffer *androidpublisher.OneTimeProductOffer) OneTimeProductOffer {
	if apiOffer == nil {
		return OneTimeProductOffer{}
	}
	offer := OneTimeProductOffer{
		PackageName:      PackageName(apiOffer.PackageName),
		ProductID:        OneTimeProductID(apiOffer.ProductId),
		PurchaseOptionID: OneTimeProductPurchaseOptionID(apiOffer.PurchaseOptionId),
		OfferID:          OneTimeProductOfferID(apiOffer.OfferId),
		State:            apiOffer.State,
		OfferTags:        offerTagsFromAPI(apiOffer.OfferTags),
		RegionsVersion:   regionsVersionFromGeneratedAPI(apiOffer.RegionsVersion),
		RegionalConfigs:  oneTimeProductOfferRegionsFromAPI(apiOffer.RegionalPricingAndAvailabilityConfigs),
	}
	switch {
	case apiOffer.DiscountedOffer != nil:
		offer.Type = OneTimeProductOfferTypeDiscounted
		offer.DiscountedOffer = oneTimeProductDiscountedOfferFromAPI(apiOffer.DiscountedOffer)
	case apiOffer.PreOrderOffer != nil:
		offer.Type = OneTimeProductOfferTypePreOrder
		offer.PreOrderOffer = oneTimeProductPreOrderOfferFromAPI(apiOffer.PreOrderOffer)
	}
	return offer
}

func oneTimeProductOfferToAPI(offer OneTimeProductOffer) *androidpublisher.OneTimeProductOffer {
	apiOffer := &androidpublisher.OneTimeProductOffer{
		PackageName:                           offer.PackageName.String(),
		ProductId:                             offer.ProductID.String(),
		PurchaseOptionId:                      offer.PurchaseOptionID.String(),
		OfferId:                               offer.OfferID.String(),
		OfferTags:                             offerTagsToAPI(offer.OfferTags),
		RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(offer.RegionalConfigs),
	}
	switch offer.Type {
	case OneTimeProductOfferTypeDiscounted:
		apiOffer.DiscountedOffer = oneTimeProductDiscountedOfferToAPI(offer.DiscountedOffer)
	case OneTimeProductOfferTypePreOrder:
		apiOffer.PreOrderOffer = oneTimeProductPreOrderOfferToAPI(offer.PreOrderOffer)
	}
	return apiOffer
}

func oneTimeProductOfferRegionsFromAPI(apiConfigs []*androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig) []OneTimeProductOfferRegion {
	if len(apiConfigs) == 0 {
		return nil
	}
	regions := make([]OneTimeProductOfferRegion, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		regions = append(regions, OneTimeProductOfferRegion{
			RegionCode:       apiConfig.RegionCode,
			Availability:     apiConfig.Availability,
			AbsoluteDiscount: moneyFromAPI(apiConfig.AbsoluteDiscount),
			RelativeDiscount: apiConfig.RelativeDiscount,
			NoOverride:       apiConfig.NoOverride != nil,
		})
	}
	return regions
}

func oneTimeProductOfferRegionsToAPI(regions []OneTimeProductOfferRegion) []*androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig {
	apiRegions := make([]*androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig, 0, len(regions))
	for _, region := range regions {
		apiRegion := &androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig{
			RegionCode:       region.RegionCode,
			Availability:     oneTimeProductOfferAvailabilityToAPI(region.Availability),
			AbsoluteDiscount: moneyToAPI(region.AbsoluteDiscount),
			RelativeDiscount: region.RelativeDiscount,
		}
		if region.NoOverride {
			apiRegion.NoOverride = &androidpublisher.OneTimeProductOfferNoPriceOverrideOptions{}
		}
		if region.RelativeDiscount != 0 {
			apiRegion.ForceSendFields = append(apiRegion.ForceSendFields, "RelativeDiscount")
		}
		apiRegions = append(apiRegions, apiRegion)
	}
	return apiRegions
}

func mergeOneTimeProductOfferAvailabilityPatches(current []OneTimeProductOfferRegion, patches []OneTimeProductOfferAvailabilityPatchRequest) []OneTimeProductOfferRegion {
	merged := append([]OneTimeProductOfferRegion(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductOfferAvailabilityPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeOneTimeProductOfferAvailabilityPatch(current []OneTimeProductOfferRegion, patch OneTimeProductOfferAvailabilityPatchRequest) []OneTimeProductOfferRegion {
	merged := make([]OneTimeProductOfferRegion, 0, len(current)+1)
	replaced := false
	for _, region := range current {
		if region.RegionCode == patch.RegionCode {
			region.Availability = oneTimeProductOfferAvailabilityToAPI(patch.Availability.String())
			merged = append(merged, region)
			replaced = true
			continue
		}
		merged = append(merged, region)
	}
	if !replaced {
		return nil
	}
	return merged
}

func mergeOneTimeProductOfferRelativeDiscountPatches(current []OneTimeProductOfferRegion, patches []OneTimeProductOfferRelativeDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := append([]OneTimeProductOfferRegion(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductOfferRelativeDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeOneTimeProductOfferRelativeDiscountPatch(current []OneTimeProductOfferRegion, patch OneTimeProductOfferRelativeDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := make([]OneTimeProductOfferRegion, 0, len(current)+1)
	replaced := false
	for _, region := range current {
		if region.RegionCode == patch.RegionCode {
			region.RelativeDiscount = patch.RelativeDiscount
			region.AbsoluteDiscount = nil
			region.NoOverride = false
			merged = append(merged, region)
			replaced = true
			continue
		}
		merged = append(merged, region)
	}
	if !replaced {
		return nil
	}
	return merged
}

func mergeOneTimeProductOfferAbsoluteDiscountPatches(current []OneTimeProductOfferRegion, patches []OneTimeProductOfferAbsoluteDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := append([]OneTimeProductOfferRegion(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductOfferAbsoluteDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeOneTimeProductOfferAbsoluteDiscountPatch(current []OneTimeProductOfferRegion, patch OneTimeProductOfferAbsoluteDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := make([]OneTimeProductOfferRegion, 0, len(current)+1)
	replaced := false
	for _, region := range current {
		if region.RegionCode == patch.RegionCode {
			absoluteDiscount := patch.AbsoluteDiscount
			region.AbsoluteDiscount = &absoluteDiscount
			region.RelativeDiscount = 0
			region.NoOverride = false
			merged = append(merged, region)
			replaced = true
			continue
		}
		merged = append(merged, region)
	}
	if !replaced {
		return nil
	}
	return merged
}

func mergeOneTimeProductOfferNoOverridePatches(current []OneTimeProductOfferRegion, patches []OneTimeProductOfferNoOverridePatchRequest) []OneTimeProductOfferRegion {
	merged := append([]OneTimeProductOfferRegion(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductOfferNoOverridePatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeOneTimeProductOfferNoOverridePatch(current []OneTimeProductOfferRegion, patch OneTimeProductOfferNoOverridePatchRequest) []OneTimeProductOfferRegion {
	merged := make([]OneTimeProductOfferRegion, 0, len(current)+1)
	replaced := false
	for _, region := range current {
		if region.RegionCode == patch.RegionCode {
			region.NoOverride = true
			region.RelativeDiscount = 0
			region.AbsoluteDiscount = nil
			merged = append(merged, region)
			replaced = true
			continue
		}
		merged = append(merged, region)
	}
	if !replaced {
		return nil
	}
	return merged
}

func oneTimeProductOfferAvailabilityToAPI(availability string) string {
	switch OneTimeProductOfferAvailability(availability) {
	case OneTimeProductOfferAvailabilityAvailable:
		return "AVAILABLE"
	case OneTimeProductOfferAvailabilityNoLongerAvailable:
		return "NO_LONGER_AVAILABLE"
	default:
		return availability
	}
}
