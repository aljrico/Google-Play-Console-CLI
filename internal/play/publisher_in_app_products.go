package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListInAppProducts(ctx context.Context, options InAppProductListOptions) (InAppProductListResult, error) {
	call := p.service.Inappproducts.List(options.PackageName.String()).Context(ctx)
	if options.Token != "" {
		call.Token(options.Token)
	}
	response, err := call.Do()
	if err != nil {
		return InAppProductListResult{}, fmt.Errorf("list in-app products for %s: %w", options.PackageName, err)
	}
	return inAppProductListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) GetInAppProduct(ctx context.Context, packageName PackageName, sku InAppProductSKU) (InAppProduct, error) {
	product, err := p.service.Inappproducts.Get(packageName.String(), sku.String()).Context(ctx).Do()
	if err != nil {
		return InAppProduct{}, fmt.Errorf("get in-app product %s for %s: %w", sku, packageName, err)
	}
	return inAppProductFromAPI(product), nil
}

func (p *GooglePublisher) BatchGetInAppProducts(ctx context.Context, options InAppProductBatchGetOptions) (InAppProductBatchGetResult, error) {
	skus := make([]string, 0, len(options.SKUs))
	for _, sku := range options.SKUs {
		skus = append(skus, sku.String())
	}
	response, err := p.service.Inappproducts.BatchGet(options.PackageName.String()).
		Sku(skus...).
		Context(ctx).
		Do()
	if err != nil {
		return InAppProductBatchGetResult{}, fmt.Errorf("batch get in-app products for %s: %w", options.PackageName, err)
	}
	return inAppProductBatchGetResultFromAPI(options, response), nil
}

func (p *GooglePublisher) DeleteInAppProduct(ctx context.Context, options InAppProductDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Inappproducts.Delete(options.PackageName.String(), options.SKU.String()).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("delete in-app product %s for %s: %w", options.SKU, options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) BatchDeleteInAppProducts(ctx context.Context, options InAppProductBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	request := &androidpublisher.InappproductsBatchDeleteRequest{
		Requests: make([]*androidpublisher.InappproductsDeleteRequest, 0, len(options.SKUs)),
	}
	for _, sku := range options.SKUs {
		request.Requests = append(request.Requests, &androidpublisher.InappproductsDeleteRequest{
			PackageName:      options.PackageName.String(),
			Sku:              sku.String(),
			LatencyTolerance: latencyTolerance,
		})
	}
	if err := p.service.Inappproducts.BatchDelete(options.PackageName.String(), request).Context(ctx).Do(); err != nil {
		return fmt.Errorf("batch delete in-app products for %s: %w", options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) CreateInAppProduct(ctx context.Context, options InAppProductCreateOptions) (InAppProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return InAppProduct{}, err
	}
	product, err := p.service.Inappproducts.Insert(options.PackageName.String(), inAppProductCreateToAPI(options)).
		AutoConvertMissingPrices(true).
		Context(ctx).
		Do()
	if err != nil {
		return InAppProduct{}, fmt.Errorf("create in-app product %s for %s: %w", options.SKU, options.PackageName, err)
	}
	return inAppProductFromAPI(product), nil
}

func (p *GooglePublisher) PatchInAppProduct(ctx context.Context, options InAppProductPatchOptions) (InAppProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return InAppProduct{}, err
	}
	call := p.service.Inappproducts.Patch(options.PackageName.String(), options.SKU.String(), inAppProductPatchToAPI(options))
	if shouldAutoConvertInAppProductPatchPrices(options) {
		call.AutoConvertMissingPrices(true)
	}
	product, err := call.Context(ctx).Do()
	if err != nil {
		return InAppProduct{}, fmt.Errorf("patch in-app product %s for %s: %w", options.SKU, options.PackageName, err)
	}
	return inAppProductFromAPI(product), nil
}

func inAppProductListResultFromAPI(options InAppProductListOptions, response *androidpublisher.InappproductsListResponse) InAppProductListResult {
	result := InAppProductListResult{
		PackageName: options.PackageName,
		Products:    []InAppProduct{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	for _, apiProduct := range response.Inappproduct {
		result.Products = append(result.Products, inAppProductFromAPI(apiProduct))
	}
	if response.TokenPagination != nil {
		result.Pagination = &InAppProductPagination{
			NextPageToken:     response.TokenPagination.NextPageToken,
			PreviousPageToken: response.TokenPagination.PreviousPageToken,
		}
	}
	return result
}

func inAppProductBatchGetResultFromAPI(options InAppProductBatchGetOptions, response *androidpublisher.InappproductsBatchGetResponse) InAppProductBatchGetResult {
	result := InAppProductBatchGetResult{
		PackageName: options.PackageName,
		Products:    []InAppProduct{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	for _, apiProduct := range response.Inappproduct {
		result.Products = append(result.Products, inAppProductFromAPI(apiProduct))
	}
	return result
}

func inAppProductFromAPI(apiProduct *androidpublisher.InAppProduct) InAppProduct {
	if apiProduct == nil {
		return InAppProduct{}
	}
	return InAppProduct{
		PackageName:                            PackageName(apiProduct.PackageName),
		SKU:                                    InAppProductSKU(apiProduct.Sku),
		Status:                                 ProductStatus(apiProduct.Status),
		PurchaseType:                           ProductPurchaseType(apiProduct.PurchaseType),
		DefaultLanguage:                        apiProduct.DefaultLanguage,
		DefaultPrice:                           productPriceFromAPI(apiProduct.DefaultPrice),
		Prices:                                 productPricesFromAPI(apiProduct.Prices),
		Listings:                               inAppProductListingsFromAPI(apiProduct.Listings),
		SubscriptionPeriod:                     apiProduct.SubscriptionPeriod,
		TrialPeriod:                            apiProduct.TrialPeriod,
		GracePeriod:                            apiProduct.GracePeriod,
		ManagedProductTaxAndComplianceSettings: managedProductTaxComplianceSettingsFromAPI(apiProduct.ManagedProductTaxesAndComplianceSettings),
		SubscriptionTaxAndComplianceSettings:   subscriptionTaxComplianceSettingsFromAPI(apiProduct.SubscriptionTaxesAndComplianceSettings),
	}
}

func inAppProductCreateToAPI(options InAppProductCreateOptions) *androidpublisher.InAppProduct {
	return &androidpublisher.InAppProduct{
		PackageName:     options.PackageName.String(),
		Sku:             options.SKU.String(),
		Status:          options.Status.String(),
		PurchaseType:    ProductPurchaseTypeManagedUser.String(),
		DefaultLanguage: options.DefaultLanguage.String(),
		DefaultPrice:    productPriceToAPI(options.DefaultPrice),
		Listings: map[string]androidpublisher.InAppProductListing{
			options.DefaultLanguage.String(): {
				Title:       options.Listing.Title,
				Description: options.Listing.Description,
				Benefits:    options.Listing.Benefits,
			},
		},
	}
}

func inAppProductPatchToAPI(options InAppProductPatchOptions) *androidpublisher.InAppProduct {
	product := &androidpublisher.InAppProduct{
		PackageName: options.PackageName.String(),
		Sku:         options.SKU.String(),
	}
	if options.Status != "" {
		product.Status = options.Status.String()
	}
	if options.DefaultLanguage != "" {
		product.DefaultLanguage = options.DefaultLanguage.String()
	}
	if options.DefaultPrice != nil {
		product.DefaultPrice = productPriceToAPI(*options.DefaultPrice)
	}
	if len(options.RegionalPrices) > 0 {
		product.Prices = productPricesToAPI(regionalProductPricesToMap(options.RegionalPrices))
	}
	if options.TaxComplianceSettings != nil {
		product.ManagedProductTaxesAndComplianceSettings = managedProductTaxComplianceSettingsToAPI(options.TaxComplianceSettings)
	}
	if options.Listing != nil {
		product.Listings = map[string]androidpublisher.InAppProductListing{
			options.ListingLanguage.String(): {
				Title:       options.Listing.Title,
				Description: options.Listing.Description,
				Benefits:    options.Listing.Benefits,
			},
		}
	}
	return product
}

func inAppProductListingsFromAPI(apiListings map[string]androidpublisher.InAppProductListing) map[string]InAppProductListing {
	if len(apiListings) == 0 {
		return nil
	}
	listings := make(map[string]InAppProductListing, len(apiListings))
	for language, apiListing := range apiListings {
		listings[language] = InAppProductListing{
			Title:       apiListing.Title,
			Description: apiListing.Description,
			Benefits:    apiListing.Benefits,
		}
	}
	return listings
}
