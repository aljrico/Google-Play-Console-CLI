package play

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func (p *GooglePublisher) ListOneTimeProducts(ctx context.Context, options OneTimeProductListOptions) (OneTimeProductListResult, error) {
	requestURL := googleapi.ResolveRelative(p.basePath, "androidpublisher/v3/applications/{packageName}/oneTimeProducts")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return OneTimeProductListResult{}, fmt.Errorf("create one-time products list request: %w", err)
	}
	googleapi.Expand(req.URL, map[string]string{"packageName": options.PackageName.String()})
	query := req.URL.Query()
	query.Set("alt", "json")
	query.Set("prettyPrint", "false")
	if options.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(options.PageSize, 10))
	}
	if options.PageToken != "" {
		query.Set("pageToken", options.PageToken)
	}
	req.URL.RawQuery = query.Encode()
	var response rawOneTimeProductListResponse
	err = p.doJSON(req, &response)
	if err != nil {
		return OneTimeProductListResult{}, fmt.Errorf("list one-time products for %s: %w", options.PackageName, err)
	}
	return oneTimeProductListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) GetOneTimeProduct(ctx context.Context, packageName PackageName, productID OneTimeProductID) (OneTimeProduct, error) {
	requestURL := googleapi.ResolveRelative(p.basePath, "androidpublisher/v3/applications/{packageName}/oneTimeProducts/{productId}")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("create one-time product get request: %w", err)
	}
	googleapi.Expand(req.URL, map[string]string{
		"packageName": packageName.String(),
		"productId":   productID.String(),
	})
	query := req.URL.Query()
	query.Set("alt", "json")
	query.Set("prettyPrint", "false")
	req.URL.RawQuery = query.Encode()
	var product rawOneTimeProduct
	err = p.doJSON(req, &product)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("get one-time product %s for %s: %w", productID, packageName, err)
	}
	return oneTimeProductFromAPI(product), nil
}

func (p *GooglePublisher) BatchGetOneTimeProducts(ctx context.Context, options OneTimeProductBatchGetOptions) (OneTimeProductBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductBatchGetResult{}, err
	}
	productIDs := make([]string, 0, len(options.ProductIDs))
	for _, productID := range options.ProductIDs {
		productIDs = append(productIDs, productID.String())
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchGet(options.PackageName.String()).
		ProductIds(productIDs...).
		Context(ctx).
		Do()
	if err != nil {
		return OneTimeProductBatchGetResult{}, fmt.Errorf("batch get one-time products for %s: %w", options.PackageName, err)
	}
	return oneTimeProductBatchGetResultFromAPI(options, response)
}

func (p *GooglePublisher) CreateOneTimeProduct(ctx context.Context, options OneTimeProductCreateOptions) (OneTimeProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProduct{}, err
	}
	existing, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do()
	if err == nil {
		if existing == nil {
			return OneTimeProduct{}, fmt.Errorf("one-time product %s for %s existence check returned an empty product", options.ProductID, options.PackageName)
		}
		return OneTimeProduct{}, fmt.Errorf("one-time product %s for %s already exists; use patch for updates", options.ProductID, options.PackageName)
	}
	if !isGoogleNotFound(err) {
		return OneTimeProduct{}, fmt.Errorf("check one-time product %s for %s before create: %w", options.ProductID, options.PackageName, err)
	}
	desired := oneTimeProductCreateDesiredProduct(options)
	product, err := p.service.Monetization.Onetimeproducts.Patch(options.PackageName.String(), options.ProductID.String(), oneTimeProductToAPI(desired)).
		AllowMissing(true).
		UpdateMask(oneTimeProductCreateUpdateMask).
		RegionsVersionVersion(options.RegionsVersion).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("create one-time product %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return oneTimeProductFromGeneratedAPI(product)
}

func (p *GooglePublisher) PatchOneTimeProduct(ctx context.Context, options OneTimeProductPatchOptions) (OneTimeProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProduct{}, err
	}
	current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("get one-time product %s for %s before patch: %w", options.ProductID, options.PackageName, err)
	}
	currentProduct, err := oneTimeProductFromGeneratedAPI(current)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("decode one-time product %s for %s before patch: %w", options.ProductID, options.PackageName, err)
	}
	mergedListings := mergeOneTimeProductListings(currentProduct.Listings, options)
	if err := validateMergedOneTimeProductListing(options.Listing.LanguageCode, mergedListings); err != nil {
		return OneTimeProduct{}, err
	}
	request := &androidpublisher.OneTimeProduct{
		PackageName: options.PackageName.String(),
		ProductId:   options.ProductID.String(),
		Listings:    oneTimeProductListingsToAPI(mergedListings),
	}
	product, err := p.service.Monetization.Onetimeproducts.Patch(options.PackageName.String(), options.ProductID.String(), request).
		UpdateMask(oneTimeProductPatchUpdateMask).
		RegionsVersionVersion(options.RegionsVersion).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("patch one-time product %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return oneTimeProductFromGeneratedAPI(product)
}

func (p *GooglePublisher) BatchPatchOneTimeProductListings(ctx context.Context, options OneTimeProductBatchPatchListingsOptions) (OneTimeProductBatchPatchListingsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductBatchPatchListingsResult{}, err
	}
	requestsByProduct := oneTimeProductBatchListingPatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductsRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return OneTimeProductBatchPatchListingsResult{}, fmt.Errorf("get one-time product %s for %s before batch patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		currentProduct, err := oneTimeProductFromGeneratedAPI(current)
		if err != nil {
			return OneTimeProductBatchPatchListingsResult{}, fmt.Errorf("decode one-time product %s for %s before batch patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedListings := mergeOneTimeProductListingPatches(currentProduct.Listings, productPatch.Requests)
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductRequest{
			OneTimeProduct: &androidpublisher.OneTimeProduct{
				PackageName: options.PackageName.String(),
				ProductId:   productPatch.ProductID.String(),
				Listings:    oneTimeProductListingsToAPI(mergedListings),
			},
			UpdateMask:       oneTimeProductPatchUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return OneTimeProductBatchPatchListingsResult{}, fmt.Errorf("batch patch one-time product listings for %s: %w", options.PackageName, err)
	}
	products := make([]OneTimeProduct, 0)
	if response != nil {
		products = make([]OneTimeProduct, 0, len(response.OneTimeProducts))
		for _, apiProduct := range response.OneTimeProducts {
			product, err := oneTimeProductFromGeneratedAPI(apiProduct)
			if err != nil {
				return OneTimeProductBatchPatchListingsResult{}, err
			}
			products = append(products, product)
		}
	}
	return OneTimeProductBatchPatchListingsResult{
		PackageName: options.PackageName,
		DryRun:      false,
		Applied:     true,
		Products:    products,
	}, nil
}

type oneTimeProductBatchListingPatchProduct struct {
	ProductID OneTimeProductID
	Requests  []OneTimeProductBatchPatchListingRequest
}

func oneTimeProductBatchListingPatchRequestsByProduct(requests []OneTimeProductBatchPatchListingRequest) []oneTimeProductBatchListingPatchProduct {
	byProduct := map[OneTimeProductID]int{}
	products := make([]oneTimeProductBatchListingPatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, oneTimeProductBatchListingPatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

func (p *GooglePublisher) DeleteOneTimeProduct(ctx context.Context, options OneTimeProductDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	err := p.service.Monetization.Onetimeproducts.Delete(
		options.PackageName.String(),
		options.ProductID.String(),
	).LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("delete one-time product %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) BatchDeleteOneTimeProducts(ctx context.Context, options OneTimeProductBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.BatchDeleteOneTimeProductsRequest{
		Requests: make([]*androidpublisher.DeleteOneTimeProductRequest, 0, len(options.ProductIDs)),
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	for _, productID := range options.ProductIDs {
		request.Requests = append(request.Requests, &androidpublisher.DeleteOneTimeProductRequest{
			PackageName:      options.PackageName.String(),
			ProductId:        productID.String(),
			LatencyTolerance: latencyTolerance,
		})
	}
	err := p.service.Monetization.Onetimeproducts.BatchDelete(options.PackageName.String(), request).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("batch delete one-time products for %s: %w", options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) BatchDeletePurchaseOptions(ctx context.Context, options PurchaseOptionBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.BatchDeletePurchaseOptionsRequest{
		Requests: make([]*androidpublisher.DeletePurchaseOptionRequest, 0, len(options.Requests)),
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	for _, item := range options.Requests {
		apiRequest := &androidpublisher.DeletePurchaseOptionRequest{
			PackageName:      options.PackageName.String(),
			ProductId:        item.ProductID.String(),
			PurchaseOptionId: item.PurchaseOptionID.String(),
			LatencyTolerance: latencyTolerance,
		}
		if options.Force {
			apiRequest.Force = true
			apiRequest.ForceSendFields = []string{"Force"}
		}
		request.Requests = append(request.Requests, apiRequest)
	}
	err := p.service.Monetization.Onetimeproducts.PurchaseOptions.BatchDelete(
		options.PackageName.String(),
		options.ParentProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("batch delete purchase options for %s/%s: %w", options.PackageName, options.ParentProductID, err)
	}
	return nil
}

func (p *GooglePublisher) BatchPatchPurchaseOptionAvailability(ctx context.Context, options PurchaseOptionBatchPatchAvailabilityOptions) (PurchaseOptionBatchPatchAvailabilityResult, error) {
	if err := options.ValidateLive(); err != nil {
		return PurchaseOptionBatchPatchAvailabilityResult{}, err
	}
	requestsByProduct := purchaseOptionAvailabilityPatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductsRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return PurchaseOptionBatchPatchAvailabilityResult{}, fmt.Errorf("get one-time product %s for %s before purchase option availability patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		currentProduct, err := oneTimeProductFromGeneratedAPI(current)
		if err != nil {
			return PurchaseOptionBatchPatchAvailabilityResult{}, fmt.Errorf("decode one-time product %s for %s before purchase option availability patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedOptions, err := mergePurchaseOptionAvailabilityPatches(currentProduct.PurchaseOptions, productPatch.Requests)
		if err != nil {
			return PurchaseOptionBatchPatchAvailabilityResult{}, err
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductRequest{
			OneTimeProduct: &androidpublisher.OneTimeProduct{
				PackageName:     options.PackageName.String(),
				ProductId:       productPatch.ProductID.String(),
				PurchaseOptions: oneTimeProductPurchaseOptionsToAPI(mergedOptions),
			},
			UpdateMask:       purchaseOptionAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return PurchaseOptionBatchPatchAvailabilityResult{}, fmt.Errorf("batch patch purchase option availability for %s: %w", options.PackageName, err)
	}
	products := make([]OneTimeProduct, 0)
	if response != nil {
		products = make([]OneTimeProduct, 0, len(response.OneTimeProducts))
		for _, apiProduct := range response.OneTimeProducts {
			product, err := oneTimeProductFromGeneratedAPI(apiProduct)
			if err != nil {
				return PurchaseOptionBatchPatchAvailabilityResult{}, err
			}
			products = append(products, product)
		}
	}
	return PurchaseOptionBatchPatchAvailabilityResult{
		PackageName: options.PackageName,
		DryRun:      false,
		Applied:     true,
		Products:    products,
	}, nil
}

func (p *GooglePublisher) BatchPatchPurchaseOptionPrices(ctx context.Context, options PurchaseOptionBatchPatchPriceOptions) (PurchaseOptionBatchPatchPriceResult, error) {
	if err := options.ValidateLive(); err != nil {
		return PurchaseOptionBatchPatchPriceResult{}, err
	}
	requestsByProduct := purchaseOptionPricePatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductsRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return PurchaseOptionBatchPatchPriceResult{}, fmt.Errorf("get one-time product %s for %s before purchase option price patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		currentProduct, err := oneTimeProductFromGeneratedAPI(current)
		if err != nil {
			return PurchaseOptionBatchPatchPriceResult{}, fmt.Errorf("decode one-time product %s for %s before purchase option price patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedOptions, err := mergePurchaseOptionPricePatches(currentProduct.PurchaseOptions, productPatch.Requests)
		if err != nil {
			return PurchaseOptionBatchPatchPriceResult{}, err
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductRequest{
			OneTimeProduct: &androidpublisher.OneTimeProduct{
				PackageName:     options.PackageName.String(),
				ProductId:       productPatch.ProductID.String(),
				PurchaseOptions: oneTimeProductPurchaseOptionsToAPI(mergedOptions),
			},
			UpdateMask:       purchaseOptionAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return PurchaseOptionBatchPatchPriceResult{}, fmt.Errorf("batch patch purchase option prices for %s: %w", options.PackageName, err)
	}
	products := make([]OneTimeProduct, 0)
	if response != nil {
		products = make([]OneTimeProduct, 0, len(response.OneTimeProducts))
		for _, apiProduct := range response.OneTimeProducts {
			product, err := oneTimeProductFromGeneratedAPI(apiProduct)
			if err != nil {
				return PurchaseOptionBatchPatchPriceResult{}, err
			}
			products = append(products, product)
		}
	}
	return PurchaseOptionBatchPatchPriceResult{
		PackageName: options.PackageName,
		DryRun:      false,
		Applied:     true,
		Products:    products,
	}, nil
}

type purchaseOptionAvailabilityPatchProduct struct {
	ProductID OneTimeProductID
	Requests  []PurchaseOptionAvailabilityPatchRequest
}

func purchaseOptionAvailabilityPatchRequestsByProduct(requests []PurchaseOptionAvailabilityPatchRequest) []purchaseOptionAvailabilityPatchProduct {
	byProduct := map[OneTimeProductID]int{}
	products := make([]purchaseOptionAvailabilityPatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, purchaseOptionAvailabilityPatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

type purchaseOptionPricePatchProduct struct {
	ProductID OneTimeProductID
	Requests  []PurchaseOptionPricePatchRequest
}

func purchaseOptionPricePatchRequestsByProduct(requests []PurchaseOptionPricePatchRequest) []purchaseOptionPricePatchProduct {
	byProduct := map[OneTimeProductID]int{}
	products := make([]purchaseOptionPricePatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, purchaseOptionPricePatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

func (p *GooglePublisher) UpdatePurchaseOptionState(ctx context.Context, options PurchaseOptionStateUpdateOptions) (OneTimeProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProduct{}, err
	}
	request := &androidpublisher.BatchUpdatePurchaseOptionStatesRequest{
		Requests: []*androidpublisher.UpdatePurchaseOptionStateRequest{
			purchaseOptionStateRequestToAPI(options),
		},
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("%s purchase option %s for %s/%s: %w", options.Action, options.PurchaseOptionID, options.PackageName, options.ProductID, err)
	}
	if response == nil || len(response.OneTimeProducts) == 0 {
		return OneTimeProduct{}, fmt.Errorf("%s purchase option %s for %s/%s: empty response", options.Action, options.PurchaseOptionID, options.PackageName, options.ProductID)
	}
	product, err := oneTimeProductFromGeneratedAPI(response.OneTimeProducts[0])
	if err != nil {
		return OneTimeProduct{}, err
	}
	return product, nil
}

type rawOneTimeProductListResponse struct {
	NextPageToken   string              `json:"nextPageToken,omitempty"`
	OneTimeProducts []rawOneTimeProduct `json:"oneTimeProducts,omitempty"`
}

type rawOneTimeProduct struct {
	PackageName                string                                  `json:"packageName,omitempty"`
	ProductID                  string                                  `json:"productId,omitempty"`
	Listings                   []rawOneTimeProductListing              `json:"listings,omitempty"`
	OfferTags                  []rawOfferTag                           `json:"offerTags,omitempty"`
	PurchaseOptions            []rawOneTimeProductPurchaseOption       `json:"purchaseOptions,omitempty"`
	RegionsVersion             *rawRegionsVersion                      `json:"regionsVersion,omitempty"`
	RestrictedPaymentCountries *rawRestrictedPaymentCountries          `json:"restrictedPaymentCountries,omitempty"`
	TaxAndComplianceSettings   *rawOneTimeProductTaxComplianceSettings `json:"taxAndComplianceSettings,omitempty"`
}

type rawOneTimeProductListing struct {
	LanguageCode string `json:"languageCode,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
}

type rawOneTimeProductPurchaseOption struct {
	PurchaseOptionID                      string                                  `json:"purchaseOptionId,omitempty"`
	State                                 string                                  `json:"state,omitempty"`
	BuyOption                             *rawOneTimeProductBuyOption             `json:"buyOption,omitempty"`
	RentOption                            *rawOneTimeProductRentOption            `json:"rentOption,omitempty"`
	OfferTags                             []rawOfferTag                           `json:"offerTags,omitempty"`
	RegionalPricingAndAvailabilityConfigs []rawOneTimeProductRegionalConfig       `json:"regionalPricingAndAvailabilityConfigs,omitempty"`
	NewRegionsConfig                      *rawOneTimeProductNewRegionsConfig      `json:"newRegionsConfig,omitempty"`
	TaxAndComplianceSettings              *rawPurchaseOptionTaxComplianceSettings `json:"taxAndComplianceSettings,omitempty"`
}

type rawOneTimeProductBuyOption struct {
	LegacyCompatible     bool `json:"legacyCompatible,omitempty"`
	MultiQuantityEnabled bool `json:"multiQuantityEnabled,omitempty"`
}

type rawOneTimeProductRentOption struct {
	RentalPeriod     string `json:"rentalPeriod,omitempty"`
	ExpirationPeriod string `json:"expirationPeriod,omitempty"`
}

type rawOneTimeProductRegionalConfig struct {
	RegionCode   string    `json:"regionCode,omitempty"`
	Availability string    `json:"availability,omitempty"`
	Price        *rawMoney `json:"price,omitempty"`
}

type rawOneTimeProductNewRegionsConfig struct {
	Availability string    `json:"availability,omitempty"`
	USDPrice     *rawMoney `json:"usdPrice,omitempty"`
	EURPrice     *rawMoney `json:"eurPrice,omitempty"`
}

type rawPurchaseOptionTaxComplianceSettings struct {
	WithdrawalRightType string `json:"withdrawalRightType,omitempty"`
}

type rawOneTimeProductTaxComplianceSettings struct {
	IsTokenizedDigitalAsset       bool                          `json:"isTokenizedDigitalAsset,omitempty"`
	ProductTaxCategoryCode        string                        `json:"productTaxCategoryCode,omitempty"`
	RegionalProductAgeRatingInfos []rawRegionalProductAgeRating `json:"regionalProductAgeRatingInfos,omitempty"`
	RegionalTaxConfigs            []rawRegionalTaxConfig        `json:"regionalTaxConfigs,omitempty"`
}

func oneTimeProductListResultFromAPI(options OneTimeProductListOptions, response rawOneTimeProductListResponse) OneTimeProductListResult {
	result := OneTimeProductListResult{
		PackageName: options.PackageName,
		Products:    []OneTimeProduct{},
		Options:     options,
	}
	result.NextPageToken = response.NextPageToken
	for _, apiProduct := range response.OneTimeProducts {
		result.Products = append(result.Products, oneTimeProductFromAPI(apiProduct))
	}
	return result
}

func oneTimeProductBatchGetResultFromAPI(options OneTimeProductBatchGetOptions, response *androidpublisher.BatchGetOneTimeProductsResponse) (OneTimeProductBatchGetResult, error) {
	result := OneTimeProductBatchGetResult{
		PackageName: options.PackageName,
		Products:    []OneTimeProduct{},
		Options:     options,
	}
	if response == nil {
		return result, nil
	}
	for _, apiProduct := range response.OneTimeProducts {
		product, err := oneTimeProductFromGeneratedAPI(apiProduct)
		if err != nil {
			return OneTimeProductBatchGetResult{}, err
		}
		result.Products = append(result.Products, product)
	}
	return result, nil
}

func oneTimeProductFromAPI(apiProduct rawOneTimeProduct) OneTimeProduct {
	if apiProduct.PackageName == "" && apiProduct.ProductID == "" {
		return OneTimeProduct{Listings: []OneTimeProductListing{}, PurchaseOptions: []OneTimeProductPurchaseOption{}}
	}
	return OneTimeProduct{
		PackageName:              PackageName(apiProduct.PackageName),
		ProductID:                OneTimeProductID(apiProduct.ProductID),
		Listings:                 oneTimeProductListingsFromAPI(apiProduct.Listings),
		OfferTags:                rawOfferTagsFromAPI(apiProduct.OfferTags),
		PurchaseOptions:          oneTimeProductPurchaseOptionsFromAPI(apiProduct.PurchaseOptions),
		RegionsVersion:           regionsVersionFromAPI(apiProduct.RegionsVersion),
		RestrictedCountries:      rawRestrictedCountriesFromAPI(apiProduct.RestrictedPaymentCountries),
		TaxAndComplianceSettings: oneTimeProductTaxComplianceSettingsFromAPI(apiProduct.TaxAndComplianceSettings),
	}
}

func oneTimeProductFromGeneratedAPI(apiProduct *androidpublisher.OneTimeProduct) (OneTimeProduct, error) {
	if apiProduct == nil {
		return OneTimeProduct{}, fmt.Errorf("one-time product response is empty")
	}
	content, err := json.Marshal(apiProduct)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("marshal one-time product response: %w", err)
	}
	var rawProduct rawOneTimeProduct
	if err := json.Unmarshal(content, &rawProduct); err != nil {
		return OneTimeProduct{}, fmt.Errorf("decode one-time product response: %w", err)
	}
	return oneTimeProductFromAPI(rawProduct), nil
}

func oneTimeProductToAPI(product OneTimeProduct) *androidpublisher.OneTimeProduct {
	return &androidpublisher.OneTimeProduct{
		PackageName:                product.PackageName.String(),
		ProductId:                  product.ProductID.String(),
		Listings:                   oneTimeProductListingsToAPI(product.Listings),
		OfferTags:                  offerTagsToAPI(product.OfferTags),
		PurchaseOptions:            oneTimeProductPurchaseOptionsToAPI(product.PurchaseOptions),
		RestrictedPaymentCountries: restrictedCountriesToAPI(product.RestrictedCountries),
		TaxAndComplianceSettings:   oneTimeProductTaxComplianceSettingsToAPI(product.TaxAndComplianceSettings),
	}
}

func purchaseOptionStateRequestToAPI(options PurchaseOptionStateUpdateOptions) *androidpublisher.UpdatePurchaseOptionStateRequest {
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	switch options.Action {
	case PurchaseOptionStateActionActivate:
		return &androidpublisher.UpdatePurchaseOptionStateRequest{
			ActivatePurchaseOptionRequest: &androidpublisher.ActivatePurchaseOptionRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case PurchaseOptionStateActionDeactivate:
		return &androidpublisher.UpdatePurchaseOptionStateRequest{
			DeactivatePurchaseOptionRequest: &androidpublisher.DeactivatePurchaseOptionRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	default:
		return &androidpublisher.UpdatePurchaseOptionStateRequest{}
	}
}

func oneTimeProductListingsFromAPI(apiListings []rawOneTimeProductListing) []OneTimeProductListing {
	listings := make([]OneTimeProductListing, 0, len(apiListings))
	for _, apiListing := range apiListings {
		listings = append(listings, OneTimeProductListing{
			LanguageCode: apiListing.LanguageCode,
			Title:        apiListing.Title,
			Description:  apiListing.Description,
		})
	}
	return listings
}

func oneTimeProductListingsToAPI(listings []OneTimeProductListing) []*androidpublisher.OneTimeProductListing {
	apiListings := make([]*androidpublisher.OneTimeProductListing, 0, len(listings))
	for _, listing := range listings {
		apiListings = append(apiListings, oneTimeProductListingToAPI(listing))
	}
	return apiListings
}

func oneTimeProductListingToAPI(listing OneTimeProductListing) *androidpublisher.OneTimeProductListing {
	return &androidpublisher.OneTimeProductListing{
		LanguageCode: listing.LanguageCode,
		Title:        listing.Title,
		Description:  listing.Description,
	}
}

func mergeOneTimeProductListings(current []OneTimeProductListing, options OneTimeProductPatchOptions) []OneTimeProductListing {
	patch := options.Listing
	merged := make([]OneTimeProductListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			merged = append(merged, mergeOneTimeProductListing(listing, options))
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

func mergeOneTimeProductListingPatches(current []OneTimeProductListing, patches []OneTimeProductBatchPatchListingRequest) []OneTimeProductListing {
	merged := append([]OneTimeProductListing(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductListingPatch(merged, patch.Listing)
	}
	return merged
}

func mergeOneTimeProductListingPatch(current []OneTimeProductListing, patch OneTimeProductListing) []OneTimeProductListing {
	merged := make([]OneTimeProductListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			merged = append(merged, patch)
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

func mergeOneTimeProductListing(current OneTimeProductListing, options OneTimeProductPatchOptions) OneTimeProductListing {
	if options.TitleSet {
		current.Title = options.Listing.Title
	}
	if options.DescriptionSet {
		current.Description = options.Listing.Description
	}
	return current
}

func validateMergedOneTimeProductListing(languageCode string, listings []OneTimeProductListing) error {
	for _, listing := range listings {
		if listing.LanguageCode != languageCode {
			continue
		}
		if strings.TrimSpace(listing.Title) == "" {
			return fmt.Errorf("one-time product listing title is required when adding or replacing listing %s", languageCode)
		}
		if strings.TrimSpace(listing.Description) == "" {
			return fmt.Errorf("one-time product listing description is required when adding or replacing listing %s", languageCode)
		}
		return nil
	}
	return fmt.Errorf("one-time product listing %s was not merged", languageCode)
}

func oneTimeProductPurchaseOptionsFromAPI(apiOptions []rawOneTimeProductPurchaseOption) []OneTimeProductPurchaseOption {
	options := make([]OneTimeProductPurchaseOption, 0, len(apiOptions))
	for _, apiOption := range apiOptions {
		options = append(options, oneTimeProductPurchaseOptionFromAPI(apiOption))
	}
	return options
}

func oneTimeProductPurchaseOptionFromAPI(apiOption rawOneTimeProductPurchaseOption) OneTimeProductPurchaseOption {
	option := OneTimeProductPurchaseOption{
		PurchaseOptionID:         apiOption.PurchaseOptionID,
		State:                    apiOption.State,
		OfferTags:                rawOfferTagsFromAPI(apiOption.OfferTags),
		RegionalConfigs:          oneTimeProductRegionalConfigsFromAPI(apiOption.RegionalPricingAndAvailabilityConfigs),
		NewRegionsConfig:         oneTimeProductNewRegionsConfigFromAPI(apiOption.NewRegionsConfig),
		TaxAndComplianceSettings: oneTimeProductPurchaseOptionTaxComplianceSettingsFromAPI(apiOption.TaxAndComplianceSettings),
	}
	switch {
	case apiOption.BuyOption != nil:
		option.Type = OneTimeProductPurchaseOptionTypeBuy
		option.LegacyCompatible = apiOption.BuyOption.LegacyCompatible
		option.MultiQuantityEnabled = apiOption.BuyOption.MultiQuantityEnabled
	case apiOption.RentOption != nil:
		option.Type = OneTimeProductPurchaseOptionTypeRent
		option.RentalPeriod = apiOption.RentOption.RentalPeriod
		option.ExpirationPeriod = apiOption.RentOption.ExpirationPeriod
	}
	return option
}

func oneTimeProductPurchaseOptionsToAPI(options []OneTimeProductPurchaseOption) []*androidpublisher.OneTimeProductPurchaseOption {
	apiOptions := make([]*androidpublisher.OneTimeProductPurchaseOption, 0, len(options))
	for _, option := range options {
		apiOptions = append(apiOptions, oneTimeProductPurchaseOptionToAPI(option))
	}
	return apiOptions
}

func oneTimeProductPurchaseOptionToAPI(option OneTimeProductPurchaseOption) *androidpublisher.OneTimeProductPurchaseOption {
	apiOption := &androidpublisher.OneTimeProductPurchaseOption{
		PurchaseOptionId:                      option.PurchaseOptionID,
		OfferTags:                             offerTagsToAPI(option.OfferTags),
		RegionalPricingAndAvailabilityConfigs: oneTimeProductRegionalConfigsToAPI(option.RegionalConfigs),
		NewRegionsConfig:                      oneTimeProductNewRegionsConfigToAPI(option.NewRegionsConfig),
		TaxAndComplianceSettings:              oneTimeProductPurchaseOptionTaxComplianceSettingsToAPI(option.TaxAndComplianceSettings),
	}
	switch option.Type {
	case OneTimeProductPurchaseOptionTypeBuy:
		apiOption.BuyOption = &androidpublisher.OneTimeProductBuyPurchaseOption{
			LegacyCompatible:     option.LegacyCompatible,
			MultiQuantityEnabled: option.MultiQuantityEnabled,
		}
		if option.LegacyCompatible {
			apiOption.BuyOption.ForceSendFields = append(apiOption.BuyOption.ForceSendFields, "LegacyCompatible")
		}
		if option.MultiQuantityEnabled {
			apiOption.BuyOption.ForceSendFields = append(apiOption.BuyOption.ForceSendFields, "MultiQuantityEnabled")
		}
	case OneTimeProductPurchaseOptionTypeRent:
		apiOption.RentOption = &androidpublisher.OneTimeProductRentPurchaseOption{
			RentalPeriod:     option.RentalPeriod,
			ExpirationPeriod: option.ExpirationPeriod,
		}
	}
	return apiOption
}

func oneTimeProductRegionalConfigsFromAPI(apiConfigs []rawOneTimeProductRegionalConfig) []OneTimeProductRegionalConfig {
	if len(apiConfigs) == 0 {
		return nil
	}
	configs := make([]OneTimeProductRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		configs = append(configs, OneTimeProductRegionalConfig{
			RegionCode:   apiConfig.RegionCode,
			Availability: apiConfig.Availability,
			Price:        rawMoneyFromAPI(apiConfig.Price),
		})
	}
	return configs
}

func oneTimeProductRegionalConfigsToAPI(configs []OneTimeProductRegionalConfig) []*androidpublisher.OneTimeProductPurchaseOptionRegionalPricingAndAvailabilityConfig {
	apiConfigs := make([]*androidpublisher.OneTimeProductPurchaseOptionRegionalPricingAndAvailabilityConfig, 0, len(configs))
	for _, config := range configs {
		apiConfigs = append(apiConfigs, &androidpublisher.OneTimeProductPurchaseOptionRegionalPricingAndAvailabilityConfig{
			RegionCode:   config.RegionCode,
			Availability: purchaseOptionAvailabilityToAPI(config.Availability),
			Price:        moneyToAPI(config.Price),
		})
	}
	return apiConfigs
}

func oneTimeProductNewRegionsConfigFromAPI(apiConfig *rawOneTimeProductNewRegionsConfig) *OneTimeProductNewRegionsPricingAndAvailability {
	if apiConfig == nil {
		return nil
	}
	return &OneTimeProductNewRegionsPricingAndAvailability{
		Availability: apiConfig.Availability,
		USDPrice:     rawMoneyFromAPI(apiConfig.USDPrice),
		EURPrice:     rawMoneyFromAPI(apiConfig.EURPrice),
	}
}

func oneTimeProductNewRegionsConfigToAPI(config *OneTimeProductNewRegionsPricingAndAvailability) *androidpublisher.OneTimeProductPurchaseOptionNewRegionsConfig {
	if config == nil {
		return nil
	}
	return &androidpublisher.OneTimeProductPurchaseOptionNewRegionsConfig{
		Availability: purchaseOptionAvailabilityToAPI(config.Availability),
		UsdPrice:     moneyToAPI(config.USDPrice),
		EurPrice:     moneyToAPI(config.EURPrice),
	}
}

func oneTimeProductTaxComplianceSettingsFromAPI(apiSettings *rawOneTimeProductTaxComplianceSettings) *OneTimeProductTaxComplianceSetting {
	if apiSettings == nil {
		return nil
	}
	return &OneTimeProductTaxComplianceSetting{
		IsTokenizedDigitalAsset: apiSettings.IsTokenizedDigitalAsset,
		ProductTaxCategoryCode:  apiSettings.ProductTaxCategoryCode,
		RegionalAgeRatings:      regionalAgeRatingsFromAPI(apiSettings.RegionalProductAgeRatingInfos),
		RegionalTaxConfigs:      regionalTaxConfigsFromAPI(apiSettings.RegionalTaxConfigs),
	}
}

func oneTimeProductTaxComplianceSettingsToAPI(settings *OneTimeProductTaxComplianceSetting) *androidpublisher.OneTimeProductTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	apiSettings := &androidpublisher.OneTimeProductTaxAndComplianceSettings{
		IsTokenizedDigitalAsset: settings.IsTokenizedDigitalAsset,
		RegionalTaxConfigs:      oneTimeProductRegionalTaxConfigsToAPI(settings.RegionalTaxConfigs),
	}
	if settings.IsTokenizedDigitalAsset {
		apiSettings.ForceSendFields = append(apiSettings.ForceSendFields, "IsTokenizedDigitalAsset")
	}
	return apiSettings
}

func oneTimeProductRegionalTaxConfigsToAPI(configs []RegionalTaxConfig) []*androidpublisher.RegionalTaxConfig {
	if len(configs) == 0 {
		return nil
	}
	apiConfigs := make([]*androidpublisher.RegionalTaxConfig, 0, len(configs))
	for _, config := range configs {
		apiConfigs = append(apiConfigs, &androidpublisher.RegionalTaxConfig{
			RegionCode:                         config.RegionCode,
			EligibleForStreamingServiceTaxRate: config.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   config.StreamingTaxType,
			TaxTier:                            config.TaxTier,
		})
	}
	return apiConfigs
}

func oneTimeProductPurchaseOptionTaxComplianceSettingsFromAPI(apiSettings *rawPurchaseOptionTaxComplianceSettings) *OneTimeProductPurchaseOptionTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	return &OneTimeProductPurchaseOptionTaxComplianceSettings{
		WithdrawalRightType: apiSettings.WithdrawalRightType,
	}
}

func oneTimeProductPurchaseOptionTaxComplianceSettingsToAPI(settings *OneTimeProductPurchaseOptionTaxComplianceSettings) *androidpublisher.PurchaseOptionTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	return &androidpublisher.PurchaseOptionTaxAndComplianceSettings{
		WithdrawalRightType: settings.WithdrawalRightType,
	}
}

func mergePurchaseOptionAvailabilityPatches(current []OneTimeProductPurchaseOption, patches []PurchaseOptionAvailabilityPatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := append([]OneTimeProductPurchaseOption(nil), current...)
	for _, patch := range patches {
		next, err := mergePurchaseOptionAvailabilityPatch(merged, patch)
		if err != nil {
			return nil, err
		}
		merged = next
	}
	return merged, nil
}

func mergePurchaseOptionAvailabilityPatch(current []OneTimeProductPurchaseOption, patch PurchaseOptionAvailabilityPatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := make([]OneTimeProductPurchaseOption, 0, len(current))
	for _, option := range current {
		if option.PurchaseOptionID != patch.PurchaseOptionID.String() {
			merged = append(merged, option)
			continue
		}
		regionalConfigs, err := mergePurchaseOptionRegionalAvailabilityPatch(option.RegionalConfigs, patch)
		if err != nil {
			return nil, err
		}
		option.RegionalConfigs = regionalConfigs
		merged = append(merged, option)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("purchase option availability patch for %s/%s references a purchase option that is not already configured", patch.ProductID, patch.PurchaseOptionID)
}

func mergePurchaseOptionRegionalAvailabilityPatch(current []OneTimeProductRegionalConfig, patch PurchaseOptionAvailabilityPatchRequest) ([]OneTimeProductRegionalConfig, error) {
	merged := make([]OneTimeProductRegionalConfig, 0, len(current))
	for _, region := range current {
		if region.RegionCode != patch.RegionCode {
			merged = append(merged, region)
			continue
		}
		if patch.Availability == PurchaseOptionAvailabilityNoLongerAvailable && region.Availability != "AVAILABLE" {
			return nil, fmt.Errorf("purchase option availability patch for %s/%s/%s cannot set noLongerAvailable from current availability %q", patch.ProductID, patch.PurchaseOptionID, patch.RegionCode, region.Availability)
		}
		region.Availability = purchaseOptionAvailabilityToAPI(patch.Availability.String())
		merged = append(merged, region)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("purchase option availability patch for %s/%s/%s references a region that is not already configured; availability-only patches cannot add regional price data", patch.ProductID, patch.PurchaseOptionID, patch.RegionCode)
}

func mergePurchaseOptionPricePatches(current []OneTimeProductPurchaseOption, patches []PurchaseOptionPricePatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := append([]OneTimeProductPurchaseOption(nil), current...)
	for _, patch := range patches {
		next, err := mergePurchaseOptionPricePatch(merged, patch)
		if err != nil {
			return nil, err
		}
		merged = next
	}
	return merged, nil
}

func mergePurchaseOptionPricePatch(current []OneTimeProductPurchaseOption, patch PurchaseOptionPricePatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := make([]OneTimeProductPurchaseOption, 0, len(current))
	for _, option := range current {
		if option.PurchaseOptionID != patch.PurchaseOptionID.String() {
			merged = append(merged, option)
			continue
		}
		regionalConfigs, err := mergePurchaseOptionRegionalPricePatch(option.RegionalConfigs, patch)
		if err != nil {
			return nil, err
		}
		option.RegionalConfigs = regionalConfigs
		merged = append(merged, option)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("purchase option price patch for %s/%s references a purchase option that is not already configured", patch.ProductID, patch.PurchaseOptionID)
}

func mergePurchaseOptionRegionalPricePatch(current []OneTimeProductRegionalConfig, patch PurchaseOptionPricePatchRequest) ([]OneTimeProductRegionalConfig, error) {
	merged := make([]OneTimeProductRegionalConfig, 0, len(current))
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
	return nil, fmt.Errorf("purchase option price patch for %s/%s/%s references a region that is not already configured; price patches cannot add regional availability", patch.ProductID, patch.PurchaseOptionID, patch.RegionCode)
}

func purchaseOptionAvailabilityToAPI(availability string) string {
	switch PurchaseOptionAvailability(availability) {
	case PurchaseOptionAvailabilityAvailable:
		return "AVAILABLE"
	case PurchaseOptionAvailabilityNoLongerAvailable:
		return "NO_LONGER_AVAILABLE"
	case PurchaseOptionAvailabilityAvailableIfReleased:
		return "AVAILABLE_IF_RELEASED"
	case PurchaseOptionAvailabilityAvailableForOffersOnly:
		return "AVAILABLE_FOR_OFFERS_ONLY"
	default:
		return availability
	}
}

func oneTimeProductDiscountedOfferFromAPI(apiOffer *androidpublisher.OneTimeProductDiscountedOffer) *OneTimeProductDiscountedOffer {
	if apiOffer == nil {
		return nil
	}
	return &OneTimeProductDiscountedOffer{
		StartTime:       apiOffer.StartTime,
		EndTime:         apiOffer.EndTime,
		RedemptionLimit: apiOffer.RedemptionLimit,
	}
}

func oneTimeProductDiscountedOfferToAPI(offer *OneTimeProductDiscountedOffer) *androidpublisher.OneTimeProductDiscountedOffer {
	if offer == nil {
		return nil
	}
	apiOffer := &androidpublisher.OneTimeProductDiscountedOffer{
		StartTime:       offer.StartTime,
		EndTime:         offer.EndTime,
		RedemptionLimit: offer.RedemptionLimit,
	}
	if offer.RedemptionLimit != 0 {
		apiOffer.ForceSendFields = append(apiOffer.ForceSendFields, "RedemptionLimit")
	}
	return apiOffer
}

func oneTimeProductPreOrderOfferFromAPI(apiOffer *androidpublisher.OneTimeProductPreOrderOffer) *OneTimeProductPreOrderOffer {
	if apiOffer == nil {
		return nil
	}
	return &OneTimeProductPreOrderOffer{
		StartTime:           apiOffer.StartTime,
		EndTime:             apiOffer.EndTime,
		ReleaseTime:         apiOffer.ReleaseTime,
		PriceChangeBehavior: apiOffer.PriceChangeBehavior,
	}
}

func oneTimeProductPreOrderOfferToAPI(offer *OneTimeProductPreOrderOffer) *androidpublisher.OneTimeProductPreOrderOffer {
	if offer == nil {
		return nil
	}
	return &androidpublisher.OneTimeProductPreOrderOffer{
		StartTime:           offer.StartTime,
		EndTime:             offer.EndTime,
		ReleaseTime:         offer.ReleaseTime,
		PriceChangeBehavior: offer.PriceChangeBehavior,
	}
}

type rawOfferTag struct {
	Tag string `json:"tag,omitempty"`
}

type rawRegionalProductAgeRating struct {
	RegionCode           string `json:"regionCode,omitempty"`
	ProductAgeRatingTier string `json:"productAgeRatingTier,omitempty"`
}

type rawRegionalTaxConfig struct {
	RegionCode                         string `json:"regionCode,omitempty"`
	EligibleForStreamingServiceTaxRate bool   `json:"eligibleForStreamingServiceTaxRate,omitempty"`
	StreamingTaxType                   string `json:"streamingTaxType,omitempty"`
	TaxTier                            string `json:"taxTier,omitempty"`
}

type rawRestrictedPaymentCountries struct {
	RegionCodes []string `json:"regionCodes,omitempty"`
}

type rawRegionsVersion struct {
	Version string `json:"version,omitempty"`
}

type rawMoney struct {
	CurrencyCode string        `json:"currencyCode,omitempty"`
	Units        flexibleInt64 `json:"units,omitempty"`
	Nanos        int64         `json:"nanos,omitempty"`
}

type flexibleInt64 int64

func (i *flexibleInt64) UnmarshalJSON(data []byte) error {
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		parsed, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return err
		}
		*i = flexibleInt64(parsed)
		return nil
	}
	var numberValue int64
	if err := json.Unmarshal(data, &numberValue); err != nil {
		return err
	}
	*i = flexibleInt64(numberValue)
	return nil
}

func regionalAgeRatingsFromAPI(apiRatings []rawRegionalProductAgeRating) []RegionalAgeRating {
	if len(apiRatings) == 0 {
		return nil
	}
	ratings := make([]RegionalAgeRating, 0, len(apiRatings))
	for _, apiRating := range apiRatings {
		ratings = append(ratings, RegionalAgeRating{
			RegionCode:           apiRating.RegionCode,
			ProductAgeRatingTier: apiRating.ProductAgeRatingTier,
		})
	}
	return ratings
}

func regionalTaxConfigsFromAPI(apiConfigs []rawRegionalTaxConfig) []RegionalTaxConfig {
	if len(apiConfigs) == 0 {
		return nil
	}
	configs := make([]RegionalTaxConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		configs = append(configs, RegionalTaxConfig{
			RegionCode:                         apiConfig.RegionCode,
			EligibleForStreamingServiceTaxRate: apiConfig.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   apiConfig.StreamingTaxType,
			TaxTier:                            apiConfig.TaxTier,
		})
	}
	return configs
}

func regionsVersionFromAPI(apiVersion *rawRegionsVersion) *RegionsVersion {
	if apiVersion == nil {
		return nil
	}
	return &RegionsVersion{Version: apiVersion.Version}
}

func rawMoneyFromAPI(apiMoney *rawMoney) *Money {
	if apiMoney == nil {
		return nil
	}
	return &Money{
		CurrencyCode: apiMoney.CurrencyCode,
		Units:        int64(apiMoney.Units),
		Nanos:        apiMoney.Nanos,
	}
}

func rawOfferTagsFromAPI(apiOfferTags []rawOfferTag) []string {
	if len(apiOfferTags) == 0 {
		return nil
	}
	tags := make([]string, 0, len(apiOfferTags))
	for _, apiOfferTag := range apiOfferTags {
		tags = append(tags, apiOfferTag.Tag)
	}
	return tags
}

func rawRestrictedCountriesFromAPI(apiCountries *rawRestrictedPaymentCountries) []string {
	if apiCountries == nil {
		return nil
	}
	return apiCountries.RegionCodes
}
