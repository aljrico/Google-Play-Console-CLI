package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListSubscriptions(ctx context.Context, options SubscriptionListOptions) (SubscriptionListResult, error) {
	call := p.service.Monetization.Subscriptions.List(options.PackageName.String()).Context(ctx)
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	if options.ShowArchived {
		call.ShowArchived(options.ShowArchived)
	}
	response, err := call.Do()
	if err != nil {
		return SubscriptionListResult{}, fmt.Errorf("list subscriptions for %s: %w", options.PackageName, err)
	}
	return subscriptionListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) GetSubscription(ctx context.Context, packageName PackageName, productID SubscriptionProductID) (Subscription, error) {
	subscription, err := p.service.Monetization.Subscriptions.Get(packageName.String(), productID.String()).Context(ctx).Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("get subscription %s for %s: %w", productID, packageName, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p *GooglePublisher) CreateSubscription(ctx context.Context, options SubscriptionCreateOptions) (Subscription, error) {
	if err := options.ValidateLive(); err != nil {
		return Subscription{}, err
	}
	subscription, err := p.service.Monetization.Subscriptions.Create(
		options.PackageName.String(),
		subscriptionCreateToAPI(options),
	).
		ProductId(options.ProductID.String()).
		RegionsVersionVersion(options.RegionsVersion).
		Context(ctx).
		Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("create subscription %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p *GooglePublisher) BatchGetSubscriptions(ctx context.Context, options SubscriptionBatchGetOptions) (SubscriptionBatchGetResult, error) {
	productIDs := make([]string, 0, len(options.ProductIDs))
	for _, productID := range options.ProductIDs {
		productIDs = append(productIDs, productID.String())
	}
	response, err := p.service.Monetization.Subscriptions.BatchGet(options.PackageName.String()).
		ProductIds(productIDs...).
		Context(ctx).
		Do()
	if err != nil {
		return SubscriptionBatchGetResult{}, fmt.Errorf("batch get subscriptions for %s: %w", options.PackageName, err)
	}
	return subscriptionBatchGetResultFromAPI(options, response), nil
}

func (p *GooglePublisher) DeleteSubscription(ctx context.Context, options SubscriptionDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Monetization.Subscriptions.Delete(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete subscription %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) DeleteBasePlan(ctx context.Context, options BasePlanDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Monetization.Subscriptions.BasePlans.Delete(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
	).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete base plan %s for %s/%s: %w", options.BasePlanID, options.PackageName, options.ProductID, err)
	}
	return nil
}

func (p *GooglePublisher) PatchSubscription(ctx context.Context, options SubscriptionPatchOptions) (Subscription, error) {
	if err := options.ValidateLive(); err != nil {
		return Subscription{}, err
	}
	current, err := p.service.Monetization.Subscriptions.Get(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("get subscription %s for %s before patch: %w", options.ProductID, options.PackageName, err)
	}
	mergedListings := mergeSubscriptionListings(subscriptionListingsFromAPI(current.Listings), options)
	request := &androidpublisher.Subscription{
		PackageName: options.PackageName.String(),
		ProductId:   options.ProductID.String(),
		Listings:    subscriptionListingsToAPI(mergedListings),
	}
	subscription, err := p.service.Monetization.Subscriptions.Patch(options.PackageName.String(), options.ProductID.String(), request).
		UpdateMask(subscriptionPatchUpdateMask).
		RegionsVersionVersion(options.RegionsVersion).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("patch subscription %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p *GooglePublisher) BatchPatchSubscriptionListings(ctx context.Context, options SubscriptionBatchPatchListingsOptions) (SubscriptionBatchPatchListingsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionBatchPatchListingsResult{}, err
	}
	requestsByProduct := subscriptionBatchListingPatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionsRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Subscriptions.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return SubscriptionBatchPatchListingsResult{}, fmt.Errorf("get subscription %s for %s before batch patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedListings := mergeSubscriptionListingPatches(subscriptionListingsFromAPI(current.Listings), productPatch.Requests)
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionRequest{
			Subscription: &androidpublisher.Subscription{
				PackageName: options.PackageName.String(),
				ProductId:   productPatch.ProductID.String(),
				Listings:    subscriptionListingsToAPI(mergedListings),
			},
			UpdateMask:       subscriptionPatchUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return SubscriptionBatchPatchListingsResult{}, fmt.Errorf("batch patch subscription listings for %s: %w", options.PackageName, err)
	}
	subscriptions := make([]Subscription, 0)
	if response != nil {
		subscriptions = make([]Subscription, 0, len(response.Subscriptions))
		for _, apiSubscription := range response.Subscriptions {
			subscriptions = append(subscriptions, subscriptionFromAPI(apiSubscription))
		}
	}
	return SubscriptionBatchPatchListingsResult{
		PackageName:   options.PackageName,
		DryRun:        false,
		Applied:       true,
		Subscriptions: subscriptions,
	}, nil
}

type subscriptionBatchListingPatchProduct struct {
	ProductID SubscriptionProductID
	Requests  []SubscriptionBatchPatchListingRequest
}

func subscriptionBatchListingPatchRequestsByProduct(requests []SubscriptionBatchPatchListingRequest) []subscriptionBatchListingPatchProduct {
	byProduct := map[SubscriptionProductID]int{}
	products := make([]subscriptionBatchListingPatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, subscriptionBatchListingPatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

type basePlanPricePatchProduct struct {
	ProductID SubscriptionProductID
	Requests  []BasePlanPricePatchRequest
}

func basePlanPricePatchRequestsByProduct(requests []BasePlanPricePatchRequest) []basePlanPricePatchProduct {
	byProduct := map[SubscriptionProductID]int{}
	products := make([]basePlanPricePatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, basePlanPricePatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

func (p *GooglePublisher) UpdateBasePlanState(ctx context.Context, options BasePlanStateUpdateOptions) (Subscription, error) {
	if err := options.ValidateLive(); err != nil {
		return Subscription{}, err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	var (
		subscription *androidpublisher.Subscription
		err          error
	)
	switch options.Action {
	case BasePlanStateActionActivate:
		subscription, err = p.service.Monetization.Subscriptions.BasePlans.Activate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			&androidpublisher.ActivateBasePlanRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case BasePlanStateActionDeactivate:
		subscription, err = p.service.Monetization.Subscriptions.BasePlans.Deactivate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			&androidpublisher.DeactivateBasePlanRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	default:
		return Subscription{}, fmt.Errorf("unsupported base plan state action %q", options.Action)
	}
	if err != nil {
		return Subscription{}, fmt.Errorf("%s base plan %s for %s/%s: %w", options.Action, options.BasePlanID, options.PackageName, options.ProductID, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p *GooglePublisher) BatchUpdateBasePlanStates(ctx context.Context, options BasePlanBatchStateUpdateOptions) (BasePlanBatchStateUpdateResult, error) {
	if err := options.ValidateLive(); err != nil {
		return BasePlanBatchStateUpdateResult{}, err
	}
	request := &androidpublisher.BatchUpdateBasePlanStatesRequest{
		Requests: make([]*androidpublisher.UpdateBasePlanStateRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, basePlanStateRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return BasePlanBatchStateUpdateResult{}, fmt.Errorf("batch %s base plans for %s/%s: %w", options.Action, options.PackageName, options.ProductID, err)
	}
	subscriptions := make([]Subscription, 0)
	if response != nil {
		subscriptions = make([]Subscription, 0, len(response.Subscriptions))
		for _, apiSubscription := range response.Subscriptions {
			subscriptions = append(subscriptions, subscriptionFromAPI(apiSubscription))
		}
	}
	return BasePlanBatchStateUpdateResult{
		PackageName:   options.PackageName,
		ProductID:     options.ProductID,
		Requests:      options.Requests,
		Action:        options.Action,
		DryRun:        false,
		Applied:       true,
		Subscriptions: subscriptions,
	}, nil
}

func (p *GooglePublisher) BatchMigrateBasePlanPrices(ctx context.Context, options BasePlanBatchPriceMigrationOptions) (BasePlanBatchPriceMigrationResult, error) {
	if err := options.ValidateLive(); err != nil {
		return BasePlanBatchPriceMigrationResult{}, err
	}
	request := &androidpublisher.BatchMigrateBasePlanPricesRequest{
		Requests: make([]*androidpublisher.MigrateBasePlanPricesRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, basePlanPriceMigrationRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.BatchMigratePrices(
		options.PackageName.String(),
		options.ProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return BasePlanBatchPriceMigrationResult{}, fmt.Errorf("batch migrate base plan prices for %s/%s: %w", options.PackageName, options.ProductID, err)
	}
	responses := basePlanPriceMigrationResponsesFromAPI(options, response)
	return BasePlanBatchPriceMigrationResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      false,
		Applied:     true,
		Responses:   responses,
	}, nil
}

func (p *GooglePublisher) BatchPatchBasePlanPrices(ctx context.Context, options BasePlanBatchPatchPriceOptions) (BasePlanBatchPatchPriceResult, error) {
	if err := options.ValidateLive(); err != nil {
		return BasePlanBatchPatchPriceResult{}, err
	}
	requestsByProduct := basePlanPricePatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionsRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Subscriptions.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return BasePlanBatchPatchPriceResult{}, fmt.Errorf("get subscription %s for %s before base plan price patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		currentSubscription := subscriptionFromAPI(current)
		mergedBasePlans, err := mergeBasePlanPricePatches(currentSubscription.BasePlans, productPatch.Requests)
		if err != nil {
			return BasePlanBatchPatchPriceResult{}, err
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionRequest{
			Subscription: &androidpublisher.Subscription{
				PackageName: options.PackageName.String(),
				ProductId:   productPatch.ProductID.String(),
				BasePlans:   subscriptionBasePlansToAPI(mergedBasePlans),
			},
			UpdateMask:       subscriptionBasePlanPatchUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return BasePlanBatchPatchPriceResult{}, fmt.Errorf("batch patch base plan prices for %s: %w", options.PackageName, err)
	}
	subscriptions := make([]Subscription, 0)
	if response != nil {
		subscriptions = make([]Subscription, 0, len(response.Subscriptions))
		for _, apiSubscription := range response.Subscriptions {
			subscriptions = append(subscriptions, subscriptionFromAPI(apiSubscription))
		}
	}
	return BasePlanBatchPatchPriceResult{
		PackageName:   options.PackageName,
		ProductID:     options.ProductID,
		DryRun:        false,
		Applied:       true,
		Subscriptions: subscriptions,
	}, nil
}

func basePlanPriceMigrationRequestToAPI(options BasePlanBatchPriceMigrationOptions, item BasePlanPriceMigrationRequest) *androidpublisher.MigrateBasePlanPricesRequest {
	return &androidpublisher.MigrateBasePlanPricesRequest{
		PackageName:             options.PackageName.String(),
		ProductId:               item.ProductID.String(),
		BasePlanId:              item.BasePlanID.String(),
		RegionsVersion:          &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
		LatencyTolerance:        productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		RegionalPriceMigrations: regionalPriceMigrationsToAPI(item.Regions),
	}
}

func basePlanPriceIncreaseTypeToAPI(value BasePlanPriceIncreaseType) string {
	switch value {
	case BasePlanPriceIncreaseTypeOptIn:
		return "PRICE_INCREASE_TYPE_OPT_IN"
	case BasePlanPriceIncreaseTypeOptOut:
		return "PRICE_INCREASE_TYPE_OPT_OUT"
	default:
		return ""
	}
}

func basePlanPriceMigrationResponsesFromAPI(options BasePlanBatchPriceMigrationOptions, response *androidpublisher.BatchMigrateBasePlanPricesResponse) []BasePlanPriceMigrationResponse {
	if response == nil {
		return []BasePlanPriceMigrationResponse{}
	}
	count := len(response.Responses)
	if count > len(options.Requests) {
		count = len(options.Requests)
	}
	responses := make([]BasePlanPriceMigrationResponse, 0, count)
	for _, request := range options.Requests[:count] {
		responses = append(responses, BasePlanPriceMigrationResponse{ProductID: request.ProductID, BasePlanID: request.BasePlanID})
	}
	return responses
}

func basePlanStateRequestToAPI(options BasePlanBatchStateUpdateOptions, item BasePlanBatchStateUpdateRequest) *androidpublisher.UpdateBasePlanStateRequest {
	switch options.Action {
	case BasePlanStateActionActivate:
		return &androidpublisher.UpdateBasePlanStateRequest{ActivateBasePlanRequest: activateBasePlanRequestToAPI(options, item)}
	case BasePlanStateActionDeactivate:
		return &androidpublisher.UpdateBasePlanStateRequest{DeactivateBasePlanRequest: deactivateBasePlanRequestToAPI(options, item)}
	default:
		return &androidpublisher.UpdateBasePlanStateRequest{}
	}
}

func activateBasePlanRequestToAPI(options BasePlanBatchStateUpdateOptions, item BasePlanBatchStateUpdateRequest) *androidpublisher.ActivateBasePlanRequest {
	return &androidpublisher.ActivateBasePlanRequest{
		PackageName:      options.PackageName.String(),
		ProductId:        item.ProductID.String(),
		BasePlanId:       item.BasePlanID.String(),
		LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
	}
}

func deactivateBasePlanRequestToAPI(options BasePlanBatchStateUpdateOptions, item BasePlanBatchStateUpdateRequest) *androidpublisher.DeactivateBasePlanRequest {
	return &androidpublisher.DeactivateBasePlanRequest{
		PackageName:      options.PackageName.String(),
		ProductId:        item.ProductID.String(),
		BasePlanId:       item.BasePlanID.String(),
		LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
	}
}

func subscriptionBatchGetResultFromAPI(options SubscriptionBatchGetOptions, response *androidpublisher.BatchGetSubscriptionsResponse) SubscriptionBatchGetResult {
	result := SubscriptionBatchGetResult{
		PackageName:   options.PackageName,
		Subscriptions: []Subscription{},
		Options:       options,
	}
	if response == nil {
		return result
	}
	for _, apiSubscription := range response.Subscriptions {
		result.Subscriptions = append(result.Subscriptions, subscriptionFromAPI(apiSubscription))
	}
	return result
}

func subscriptionTaxComplianceSettingsToAPI(settings *ProductTaxComplianceSettings) *androidpublisher.SubscriptionTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	apiSettings := &androidpublisher.SubscriptionTaxAndComplianceSettings{
		EeaWithdrawalRightType: settings.EEAWithdrawalRightType,
	}
	if settings.IsTokenizedDigitalAsset != nil {
		apiSettings.IsTokenizedDigitalAsset = *settings.IsTokenizedDigitalAsset
		apiSettings.ForceSendFields = append(apiSettings.ForceSendFields, "IsTokenizedDigitalAsset")
	}
	if len(settings.TaxRateInfoByRegionCode) > 0 {
		apiSettings.TaxRateInfoByRegionCode = regionalTaxRateInfoToAPI(settings.TaxRateInfoByRegionCode)
	}
	return apiSettings
}

func subscriptionTaxComplianceSettingsFromAPI(apiSettings *androidpublisher.SubscriptionTaxAndComplianceSettings) *ProductTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	tokenizedDigitalAsset := apiSettings.IsTokenizedDigitalAsset
	return &ProductTaxComplianceSettings{
		EEAWithdrawalRightType:  apiSettings.EeaWithdrawalRightType,
		IsTokenizedDigitalAsset: &tokenizedDigitalAsset,
		TaxRateInfoByRegionCode: regionalTaxRateInfoFromAPI(apiSettings.TaxRateInfoByRegionCode),
	}
}

func subscriptionListResultFromAPI(options SubscriptionListOptions, response *androidpublisher.ListSubscriptionsResponse) SubscriptionListResult {
	result := SubscriptionListResult{
		PackageName:   options.PackageName,
		Subscriptions: []Subscription{},
		Options:       options,
	}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiSubscription := range response.Subscriptions {
		result.Subscriptions = append(result.Subscriptions, subscriptionFromAPI(apiSubscription))
	}
	return result
}

func subscriptionFromAPI(apiSubscription *androidpublisher.Subscription) Subscription {
	if apiSubscription == nil {
		return Subscription{Listings: []SubscriptionListing{}, BasePlans: []SubscriptionBasePlan{}}
	}
	return Subscription{
		PackageName:              PackageName(apiSubscription.PackageName),
		ProductID:                SubscriptionProductID(apiSubscription.ProductId),
		Archived:                 apiSubscription.Archived,
		Listings:                 subscriptionListingsFromAPI(apiSubscription.Listings),
		BasePlans:                subscriptionBasePlansFromAPI(apiSubscription.BasePlans),
		RestrictedCountries:      restrictedCountriesFromAPI(apiSubscription.RestrictedPaymentCountries),
		TaxAndComplianceSettings: subscriptionTaxComplianceSettingsFromAPI(apiSubscription.TaxAndComplianceSettings),
	}
}

func subscriptionCreateToAPI(options SubscriptionCreateOptions) *androidpublisher.Subscription {
	subscription := subscriptionCreateDesiredSubscription(options)
	return &androidpublisher.Subscription{
		PackageName:                subscription.PackageName.String(),
		ProductId:                  subscription.ProductID.String(),
		Listings:                   subscriptionListingsToAPI(subscription.Listings),
		BasePlans:                  subscriptionBasePlansToAPI(subscription.BasePlans),
		RestrictedPaymentCountries: restrictedCountriesToAPI(subscription.RestrictedCountries),
		TaxAndComplianceSettings:   subscriptionTaxComplianceSettingsToAPI(subscription.TaxAndComplianceSettings),
	}
}

func subscriptionListingsFromAPI(apiListings []*androidpublisher.SubscriptionListing) []SubscriptionListing {
	listings := make([]SubscriptionListing, 0, len(apiListings))
	for _, apiListing := range apiListings {
		if apiListing == nil {
			continue
		}
		listings = append(listings, SubscriptionListing{
			LanguageCode: apiListing.LanguageCode,
			Title:        apiListing.Title,
			Description:  apiListing.Description,
			Benefits:     apiListing.Benefits,
		})
	}
	return listings
}

func subscriptionListingsToAPI(listings []SubscriptionListing) []*androidpublisher.SubscriptionListing {
	apiListings := make([]*androidpublisher.SubscriptionListing, 0, len(listings))
	for _, listing := range listings {
		apiListings = append(apiListings, subscriptionListingToAPI(listing))
	}
	return apiListings
}

func subscriptionListingToAPI(listing SubscriptionListing) *androidpublisher.SubscriptionListing {
	return &androidpublisher.SubscriptionListing{
		LanguageCode: listing.LanguageCode,
		Title:        listing.Title,
		Description:  listing.Description,
		Benefits:     listing.Benefits,
	}
}

func mergeSubscriptionListings(current []SubscriptionListing, options SubscriptionPatchOptions) []SubscriptionListing {
	patch := options.Listing
	merged := make([]SubscriptionListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			merged = append(merged, mergeSubscriptionListing(listing, options))
			replaced = true
			continue
		}
		merged = append(merged, listing)
	}
	if !replaced {
		merged = append(merged, patch)
	}
	return merged
}

func mergeSubscriptionListing(current SubscriptionListing, options SubscriptionPatchOptions) SubscriptionListing {
	current.Title = options.Listing.Title
	if options.DescriptionSet {
		current.Description = options.Listing.Description
	}
	if options.BenefitsSet {
		current.Benefits = options.Listing.Benefits
	}
	return current
}

func mergeSubscriptionListingPatches(current []SubscriptionListing, patches []SubscriptionBatchPatchListingRequest) []SubscriptionListing {
	merged := append([]SubscriptionListing(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionListingPatch(merged, patch.Listing)
	}
	return merged
}

func mergeSubscriptionListingPatch(current []SubscriptionListing, patch SubscriptionListing) []SubscriptionListing {
	merged := make([]SubscriptionListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			listing.Title = patch.Title
			listing.Description = patch.Description
			merged = append(merged, listing)
			replaced = true
			continue
		}
		merged = append(merged, listing)
	}
	if !replaced {
		merged = append(merged, patch)
	}
	return merged
}

func subscriptionBasePlansFromAPI(apiBasePlans []*androidpublisher.BasePlan) []SubscriptionBasePlan {
	basePlans := make([]SubscriptionBasePlan, 0, len(apiBasePlans))
	for _, apiBasePlan := range apiBasePlans {
		if apiBasePlan == nil {
			continue
		}
		basePlans = append(basePlans, subscriptionBasePlanFromAPI(apiBasePlan))
	}
	return basePlans
}

func subscriptionBasePlanFromAPI(apiBasePlan *androidpublisher.BasePlan) SubscriptionBasePlan {
	basePlan := SubscriptionBasePlan{
		BasePlanID:         apiBasePlan.BasePlanId,
		State:              SubscriptionState(apiBasePlan.State),
		OfferTags:          offerTagsFromAPI(apiBasePlan.OfferTags),
		RegionalConfigs:    subscriptionRegionalConfigsFromAPI(apiBasePlan.RegionalConfigs),
		OtherRegionsConfig: subscriptionOtherRegionsConfigFromAPI(apiBasePlan.OtherRegionsConfig),
	}
	switch {
	case apiBasePlan.AutoRenewingBasePlanType != nil:
		basePlan.Type = SubscriptionBasePlanTypeAutoRenewing
		basePlan.BillingPeriodDuration = apiBasePlan.AutoRenewingBasePlanType.BillingPeriodDuration
		basePlan.GracePeriodDuration = apiBasePlan.AutoRenewingBasePlanType.GracePeriodDuration
		basePlan.AccountHoldDuration = apiBasePlan.AutoRenewingBasePlanType.AccountHoldDuration
		basePlan.LegacyCompatible = apiBasePlan.AutoRenewingBasePlanType.LegacyCompatible
		basePlan.LegacyCompatibleSubscriptionOfferID = apiBasePlan.AutoRenewingBasePlanType.LegacyCompatibleSubscriptionOfferId
		basePlan.ProrationMode = apiBasePlan.AutoRenewingBasePlanType.ProrationMode
		basePlan.ResubscribeState = apiBasePlan.AutoRenewingBasePlanType.ResubscribeState
	case apiBasePlan.PrepaidBasePlanType != nil:
		basePlan.Type = SubscriptionBasePlanTypePrepaid
		basePlan.BillingPeriodDuration = apiBasePlan.PrepaidBasePlanType.BillingPeriodDuration
		basePlan.TimeExtension = apiBasePlan.PrepaidBasePlanType.TimeExtension
	case apiBasePlan.InstallmentsBasePlanType != nil:
		basePlan.Type = SubscriptionBasePlanTypeInstallments
		basePlan.BillingPeriodDuration = apiBasePlan.InstallmentsBasePlanType.BillingPeriodDuration
		basePlan.GracePeriodDuration = apiBasePlan.InstallmentsBasePlanType.GracePeriodDuration
		basePlan.AccountHoldDuration = apiBasePlan.InstallmentsBasePlanType.AccountHoldDuration
		basePlan.CommittedPaymentsCount = apiBasePlan.InstallmentsBasePlanType.CommittedPaymentsCount
		basePlan.ProrationMode = apiBasePlan.InstallmentsBasePlanType.ProrationMode
		basePlan.RenewalType = apiBasePlan.InstallmentsBasePlanType.RenewalType
		basePlan.ResubscribeState = apiBasePlan.InstallmentsBasePlanType.ResubscribeState
	}
	return basePlan
}

func subscriptionBasePlansToAPI(basePlans []SubscriptionBasePlan) []*androidpublisher.BasePlan {
	apiBasePlans := make([]*androidpublisher.BasePlan, 0, len(basePlans))
	for _, basePlan := range basePlans {
		apiBasePlans = append(apiBasePlans, subscriptionBasePlanToAPI(basePlan))
	}
	return apiBasePlans
}

func subscriptionBasePlanToAPI(basePlan SubscriptionBasePlan) *androidpublisher.BasePlan {
	apiBasePlan := &androidpublisher.BasePlan{
		BasePlanId:         basePlan.BasePlanID,
		OfferTags:          offerTagsToAPI(basePlan.OfferTags),
		RegionalConfigs:    subscriptionRegionalConfigsToAPI(basePlan.RegionalConfigs),
		OtherRegionsConfig: subscriptionOtherRegionsConfigToAPI(basePlan.OtherRegionsConfig),
	}
	switch basePlan.Type {
	case SubscriptionBasePlanTypeAutoRenewing:
		apiBasePlan.AutoRenewingBasePlanType = &androidpublisher.AutoRenewingBasePlanType{
			BillingPeriodDuration:               basePlan.BillingPeriodDuration,
			GracePeriodDuration:                 basePlan.GracePeriodDuration,
			AccountHoldDuration:                 basePlan.AccountHoldDuration,
			LegacyCompatible:                    basePlan.LegacyCompatible,
			LegacyCompatibleSubscriptionOfferId: basePlan.LegacyCompatibleSubscriptionOfferID,
			ProrationMode:                       basePlan.ProrationMode,
			ResubscribeState:                    basePlan.ResubscribeState,
		}
		if basePlan.LegacyCompatible {
			apiBasePlan.AutoRenewingBasePlanType.ForceSendFields = append(apiBasePlan.AutoRenewingBasePlanType.ForceSendFields, "LegacyCompatible")
		}
	case SubscriptionBasePlanTypePrepaid:
		apiBasePlan.PrepaidBasePlanType = &androidpublisher.PrepaidBasePlanType{
			BillingPeriodDuration: basePlan.BillingPeriodDuration,
			TimeExtension:         basePlan.TimeExtension,
		}
	case SubscriptionBasePlanTypeInstallments:
		apiBasePlan.InstallmentsBasePlanType = &androidpublisher.InstallmentsBasePlanType{
			BillingPeriodDuration:  basePlan.BillingPeriodDuration,
			GracePeriodDuration:    basePlan.GracePeriodDuration,
			AccountHoldDuration:    basePlan.AccountHoldDuration,
			CommittedPaymentsCount: basePlan.CommittedPaymentsCount,
			ProrationMode:          basePlan.ProrationMode,
			RenewalType:            basePlan.RenewalType,
			ResubscribeState:       basePlan.ResubscribeState,
		}
	}
	return apiBasePlan
}

func subscriptionRegionalConfigsFromAPI(apiConfigs []*androidpublisher.RegionalBasePlanConfig) []SubscriptionRegionalConfig {
	if len(apiConfigs) == 0 {
		return nil
	}
	configs := make([]SubscriptionRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		configs = append(configs, SubscriptionRegionalConfig{
			RegionCode:                apiConfig.RegionCode,
			NewSubscriberAvailability: apiConfig.NewSubscriberAvailability,
			Price:                     moneyFromAPI(apiConfig.Price),
		})
	}
	return configs
}

func subscriptionRegionalConfigsToAPI(configs []SubscriptionRegionalConfig) []*androidpublisher.RegionalBasePlanConfig {
	apiConfigs := make([]*androidpublisher.RegionalBasePlanConfig, 0, len(configs))
	for _, config := range configs {
		apiConfig := &androidpublisher.RegionalBasePlanConfig{
			RegionCode:                config.RegionCode,
			NewSubscriberAvailability: config.NewSubscriberAvailability,
			Price:                     moneyToAPI(config.Price),
		}
		apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "NewSubscriberAvailability")
		apiConfigs = append(apiConfigs, apiConfig)
	}
	return apiConfigs
}

func mergeBasePlanPricePatches(current []SubscriptionBasePlan, patches []BasePlanPricePatchRequest) ([]SubscriptionBasePlan, error) {
	merged := append([]SubscriptionBasePlan(nil), current...)
	for _, patch := range patches {
		next, err := mergeBasePlanPricePatch(merged, patch)
		if err != nil {
			return nil, err
		}
		merged = next
	}
	return merged, nil
}

func mergeBasePlanPricePatch(current []SubscriptionBasePlan, patch BasePlanPricePatchRequest) ([]SubscriptionBasePlan, error) {
	merged := make([]SubscriptionBasePlan, 0, len(current))
	for _, basePlan := range current {
		if basePlan.BasePlanID != patch.BasePlanID.String() {
			merged = append(merged, basePlan)
			continue
		}
		regionalConfigs, err := mergeBasePlanRegionalPricePatch(basePlan.RegionalConfigs, patch)
		if err != nil {
			return nil, err
		}
		basePlan.RegionalConfigs = regionalConfigs
		merged = append(merged, basePlan)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("base plan price patch for %s/%s references a base plan that is not already configured", patch.ProductID, patch.BasePlanID)
}

func mergeBasePlanRegionalPricePatch(current []SubscriptionRegionalConfig, patch BasePlanPricePatchRequest) ([]SubscriptionRegionalConfig, error) {
	merged := make([]SubscriptionRegionalConfig, 0, len(current))
	for _, region := range current {
		if region.RegionCode != patch.RegionCode {
			merged = append(merged, region)
			continue
		}
		region.Price = &Money{
			CurrencyCode: patch.Price.CurrencyCode,
			Units:        patch.Price.Units,
			Nanos:        patch.Price.Nanos,
		}
		merged = append(merged, region)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("base plan price patch for %s/%s/%s references a region that is not already configured; price patches cannot add regional availability", patch.ProductID, patch.BasePlanID, patch.RegionCode)
}

func subscriptionOtherRegionsConfigFromAPI(apiConfig *androidpublisher.OtherRegionsBasePlanConfig) *SubscriptionOtherRegionsConfig {
	if apiConfig == nil {
		return nil
	}
	return &SubscriptionOtherRegionsConfig{
		NewSubscriberAvailability: apiConfig.NewSubscriberAvailability,
		USDPrice:                  moneyFromAPI(apiConfig.UsdPrice),
		EURPrice:                  moneyFromAPI(apiConfig.EurPrice),
	}
}

func subscriptionOtherRegionsConfigToAPI(config *SubscriptionOtherRegionsConfig) *androidpublisher.OtherRegionsBasePlanConfig {
	if config == nil {
		return nil
	}
	apiConfig := &androidpublisher.OtherRegionsBasePlanConfig{
		NewSubscriberAvailability: config.NewSubscriberAvailability,
		UsdPrice:                  moneyToAPI(config.USDPrice),
		EurPrice:                  moneyToAPI(config.EURPrice),
	}
	apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "NewSubscriberAvailability")
	return apiConfig
}

func orderSubscriptionDetailsFromAPI(apiDetails *androidpublisher.SubscriptionDetails) *OrderSubscriptionDetails {
	if apiDetails == nil {
		return nil
	}
	return &OrderSubscriptionDetails{
		BasePlanID:             apiDetails.BasePlanId,
		OfferID:                apiDetails.OfferId,
		OfferPhase:             apiDetails.OfferPhase,
		ServicePeriodStartTime: apiDetails.ServicePeriodStartTime,
		ServicePeriodEndTime:   apiDetails.ServicePeriodEndTime,
	}
}

func subscriptionRevokeRequestToAPI(options SubscriptionPurchaseRevokeOptions) *androidpublisher.RevokeSubscriptionPurchaseRequest {
	context := &androidpublisher.RevocationContext{}
	switch options.RefundType {
	case SubscriptionRefundTypeFull:
		context.FullRefund = &androidpublisher.RevocationContextFullRefund{}
	case SubscriptionRefundTypeProrated:
		context.ProratedRefund = &androidpublisher.RevocationContextProratedRefund{}
	case SubscriptionRefundTypeItem:
		context.ItemBasedRefund = &androidpublisher.RevocationContextItemBasedRefund{
			ProductId: options.RefundProductID.String(),
		}
	}
	return &androidpublisher.RevokeSubscriptionPurchaseRequest{
		RevocationContext: context,
	}
}
