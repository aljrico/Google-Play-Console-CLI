package play

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func (p *GooglePublisher) GetProductPurchase(ctx context.Context, options ProductPurchaseOptions) (ProductPurchase, error) {
	purchase, err := p.service.Purchases.Productsv2.Getproductpurchasev2(options.PackageName.String(), options.Token.String()).Context(ctx).Do()
	if err != nil {
		return ProductPurchase{}, fmt.Errorf("get product purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return productPurchaseFromAPI(options, purchase), nil
}

func (p *GooglePublisher) AcknowledgeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.ProductPurchasesAcknowledgeRequest{}
	if options.DeveloperPayload != "" {
		request.DeveloperPayload = options.DeveloperPayload
	}
	if err := p.service.Purchases.Products.Acknowledge(options.PackageName.String(), options.ProductID.String(), options.Token.String(), request).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("acknowledge product purchase %s for %s/%s: %w", options.Token, options.PackageName, options.ProductID, err)
	}
	return nil
}

func (p *GooglePublisher) ConsumeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Purchases.Products.Consume(options.PackageName.String(), options.ProductID.String(), options.Token.String()).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("consume product purchase %s for %s/%s: %w", options.Token, options.PackageName, options.ProductID, err)
	}
	return nil
}

func (p *GooglePublisher) GetSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseOptions) (SubscriptionPurchase, error) {
	purchase, err := p.service.Purchases.Subscriptionsv2.Get(options.PackageName.String(), options.Token.String()).Context(ctx).Do()
	if err != nil {
		return SubscriptionPurchase{}, fmt.Errorf("get subscription purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return subscriptionPurchaseFromAPI(options.PackageName, options.Token, purchase), nil
}

func (p *GooglePublisher) AcknowledgeSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if options.Action != SubscriptionPurchaseMutationActionAcknowledge {
		return fmt.Errorf("acknowledge subscription purchase requires action %q", SubscriptionPurchaseMutationActionAcknowledge)
	}
	request := &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{}
	if options.DeveloperPayload != "" {
		request.DeveloperPayload = options.DeveloperPayload
	}
	if err := p.service.Purchases.Subscriptions.Acknowledge(options.PackageName.String(), options.SubscriptionID.String(), options.Token.String(), request).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("acknowledge subscription purchase %s for %s/%s: %w", options.Token, options.PackageName, options.SubscriptionID, err)
	}
	return nil
}

func (p *GooglePublisher) CancelSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if options.Action != SubscriptionPurchaseMutationActionCancel {
		return fmt.Errorf("cancel subscription purchase requires action %q", SubscriptionPurchaseMutationActionCancel)
	}
	requestBody := subscriptionCancelRequestToAPI(options)
	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("encode subscription cancel request: %w", err)
	}
	requestURL := googleapi.ResolveRelative(p.basePath, "androidpublisher/v3/applications/{packageName}/purchases/subscriptionsv2/tokens/{token}:cancel")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create subscription cancel request: %w", err)
	}
	googleapi.Expand(req.URL, map[string]string{
		"packageName": options.PackageName.String(),
		"token":       options.Token.String(),
	})
	query := req.URL.Query()
	query.Set("alt", "json")
	query.Set("prettyPrint", "false")
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", "application/json")
	if err := p.doNoContent(req); err != nil {
		return fmt.Errorf("cancel subscription purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) RevokeSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseRevokeOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if _, err := p.service.Purchases.Subscriptionsv2.Revoke(options.PackageName.String(), options.Token.String(), subscriptionRevokeRequestToAPI(options)).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("revoke subscription purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) ListVoidedPurchases(ctx context.Context, options VoidedPurchaseListOptions) (VoidedPurchaseListResult, error) {
	call := p.service.Purchases.Voidedpurchases.List(options.PackageName.String()).Context(ctx)
	if options.MaxResults > 0 {
		call.MaxResults(options.MaxResults)
	}
	if options.StartIndex > 0 {
		call.StartIndex(options.StartIndex)
	}
	if options.Token != "" {
		call.Token(options.Token)
	}
	if options.StartTimeMillis > 0 {
		call.StartTime(options.StartTimeMillis)
	}
	if options.EndTimeMillis > 0 {
		call.EndTime(options.EndTimeMillis)
	}
	if options.Type != VoidedPurchaseTypeProductsOnly {
		call.Type(int64(options.Type))
	}
	if options.IncludeQuantityBasedPartialRefund {
		call.IncludeQuantityBasedPartialRefund(true)
	}
	response, err := call.Do()
	if err != nil {
		return VoidedPurchaseListResult{}, fmt.Errorf("list voided purchases for %s: %w", options.PackageName, err)
	}
	return voidedPurchaseListResultFromAPI(options, response), nil
}

func cancellationEventFromAPI(apiEvent *androidpublisher.CancellationEvent) *OrderEvent {
	if apiEvent == nil {
		return nil
	}
	return &OrderEvent{EventTime: apiEvent.EventTime}
}

func processedEventFromAPI(apiEvent *androidpublisher.ProcessedEvent) *OrderEvent {
	if apiEvent == nil {
		return nil
	}
	return &OrderEvent{EventTime: apiEvent.EventTime}
}

func orderRefundEventFromAPI(apiEvent *androidpublisher.RefundEvent) *OrderRefundEvent {
	if apiEvent == nil {
		return nil
	}
	return &OrderRefundEvent{
		EventTime:     apiEvent.EventTime,
		RefundReason:  apiEvent.RefundReason,
		RefundDetails: refundDetailsFromAPI(apiEvent.RefundDetails),
	}
}

func partialRefundEventFromAPI(apiEvent *androidpublisher.PartialRefundEvent) PartialRefundEvent {
	return PartialRefundEvent{
		CreateTime:    apiEvent.CreateTime,
		ProcessTime:   apiEvent.ProcessTime,
		State:         apiEvent.State,
		RefundDetails: refundDetailsFromAPI(apiEvent.RefundDetails),
	}
}

func productPurchaseFromAPI(options ProductPurchaseOptions, apiPurchase *androidpublisher.ProductPurchaseV2) ProductPurchase {
	if apiPurchase == nil {
		return ProductPurchase{PackageName: options.PackageName, ProductID: options.ProductID, Token: options.Token, LineItems: []ProductPurchaseLineItem{}}
	}
	return ProductPurchase{
		PackageName:                 options.PackageName,
		ProductID:                   firstProductID(options.ProductID, apiPurchase.ProductLineItem),
		Token:                       options.Token,
		OrderID:                     apiPurchase.OrderId,
		PurchaseState:               purchaseStateFromAPI(apiPurchase.PurchaseStateContext),
		PurchaseCompletionTime:      apiPurchase.PurchaseCompletionTime,
		AcknowledgementState:        apiPurchase.AcknowledgementState,
		RegionCode:                  apiPurchase.RegionCode,
		ObfuscatedExternalAccountID: apiPurchase.ObfuscatedExternalAccountId,
		ObfuscatedExternalProfileID: apiPurchase.ObfuscatedExternalProfileId,
		TestPurchase:                apiPurchase.TestPurchaseContext != nil,
		LineItems:                   productPurchaseLineItemsFromAPI(apiPurchase.ProductLineItem),
	}
}

func purchaseStateFromAPI(apiState *androidpublisher.PurchaseStateContext) string {
	if apiState == nil {
		return ""
	}
	return apiState.PurchaseState
}

func firstProductID(fallback InAppProductSKU, apiItems []*androidpublisher.ProductLineItem) InAppProductSKU {
	for _, apiItem := range apiItems {
		if apiItem != nil && apiItem.ProductId != "" {
			return InAppProductSKU(apiItem.ProductId)
		}
	}
	return fallback
}

func productPurchaseLineItemsFromAPI(apiItems []*androidpublisher.ProductLineItem) []ProductPurchaseLineItem {
	items := make([]ProductPurchaseLineItem, 0, len(apiItems))
	for _, apiItem := range apiItems {
		if apiItem == nil {
			continue
		}
		item := ProductPurchaseLineItem{ProductID: apiItem.ProductId}
		if apiItem.ProductOfferDetails != nil {
			item.ConsumptionState = apiItem.ProductOfferDetails.ConsumptionState
			item.PurchaseOptionID = apiItem.ProductOfferDetails.PurchaseOptionId
			item.OfferID = apiItem.ProductOfferDetails.OfferId
			item.OfferToken = apiItem.ProductOfferDetails.OfferToken
			item.OfferTags = apiItem.ProductOfferDetails.OfferTags
			item.Quantity = apiItem.ProductOfferDetails.Quantity
			item.RefundableQuantity = apiItem.ProductOfferDetails.RefundableQuantity
		}
		items = append(items, item)
	}
	return items
}

type rawSubscriptionCancelRequest struct {
	CancellationContext rawSubscriptionCancellationContext `json:"cancellationContext"`
}

type rawSubscriptionCancellationContext struct {
	CancellationType string `json:"cancellationType"`
}

func subscriptionCancelRequestToAPI(options SubscriptionPurchaseMutationOptions) rawSubscriptionCancelRequest {
	return rawSubscriptionCancelRequest{
		CancellationContext: rawSubscriptionCancellationContext{
			CancellationType: subscriptionCancellationTypeToAPI(options.CancellationType),
		},
	}
}

func subscriptionCancellationTypeToAPI(cancellationType SubscriptionCancellationType) string {
	switch cancellationType {
	case SubscriptionCancellationTypeUserRequestedStopRenewals:
		return "USER_REQUESTED_STOP_RENEWALS"
	case SubscriptionCancellationTypeDeveloperRequestedStopPayments:
		return "DEVELOPER_REQUESTED_STOP_PAYMENTS"
	default:
		return "CANCELLATION_TYPE_UNSPECIFIED"
	}
}

func subscriptionPurchaseFromAPI(packageName PackageName, token PurchaseToken, apiPurchase *androidpublisher.SubscriptionPurchaseV2) SubscriptionPurchase {
	if apiPurchase == nil {
		return SubscriptionPurchase{PackageName: packageName, Token: token, LineItems: []SubscriptionPurchaseLineItem{}}
	}
	return SubscriptionPurchase{
		PackageName:                 packageName,
		Token:                       token,
		SubscriptionState:           apiPurchase.SubscriptionState,
		AcknowledgementState:        apiPurchase.AcknowledgementState,
		LatestOrderID:               apiPurchase.LatestOrderId,
		LinkedPurchaseToken:         apiPurchase.LinkedPurchaseToken,
		RegionCode:                  apiPurchase.RegionCode,
		StartTime:                   apiPurchase.StartTime,
		LineItems:                   subscriptionPurchaseLineItemsFromAPI(apiPurchase.LineItems),
		ExternalAccountID:           externalAccountIDFromAPI(apiPurchase.ExternalAccountIdentifiers),
		ObfuscatedExternalAccountID: obfuscatedExternalAccountIDFromAPI(apiPurchase.ExternalAccountIdentifiers),
		ObfuscatedExternalProfileID: obfuscatedExternalProfileIDFromAPI(apiPurchase.ExternalAccountIdentifiers),
		TestPurchase:                apiPurchase.TestPurchase != nil,
	}
}

func subscriptionPurchaseLineItemsFromAPI(apiItems []*androidpublisher.SubscriptionPurchaseLineItem) []SubscriptionPurchaseLineItem {
	items := make([]SubscriptionPurchaseLineItem, 0, len(apiItems))
	for _, apiItem := range apiItems {
		if apiItem == nil {
			continue
		}
		item := SubscriptionPurchaseLineItem{
			ProductID:               apiItem.ProductId,
			ExpiryTime:              apiItem.ExpiryTime,
			LatestSuccessfulOrderID: apiItem.LatestSuccessfulOrderId,
		}
		if apiItem.OfferDetails != nil {
			item.BasePlanID = apiItem.OfferDetails.BasePlanId
			item.OfferID = apiItem.OfferDetails.OfferId
			item.OfferTags = apiItem.OfferDetails.OfferTags
		}
		if apiItem.AutoRenewingPlan != nil {
			autoRenewEnabled := apiItem.AutoRenewingPlan.AutoRenewEnabled
			item.AutoRenewEnabled = &autoRenewEnabled
			item.RecurringPrice = moneyFromAPI(apiItem.AutoRenewingPlan.RecurringPrice)
		}
		if apiItem.PrepaidPlan != nil {
			item.Prepaid = true
			item.AllowExtendAfterTime = apiItem.PrepaidPlan.AllowExtendAfterTime
		}
		items = append(items, item)
	}
	return items
}

func externalAccountIDFromAPI(apiIdentifiers *androidpublisher.ExternalAccountIdentifiers) string {
	if apiIdentifiers == nil {
		return ""
	}
	return apiIdentifiers.ExternalAccountId
}

func obfuscatedExternalAccountIDFromAPI(apiIdentifiers *androidpublisher.ExternalAccountIdentifiers) string {
	if apiIdentifiers == nil {
		return ""
	}
	return apiIdentifiers.ObfuscatedExternalAccountId
}

func obfuscatedExternalProfileIDFromAPI(apiIdentifiers *androidpublisher.ExternalAccountIdentifiers) string {
	if apiIdentifiers == nil {
		return ""
	}
	return apiIdentifiers.ObfuscatedExternalProfileId
}

func voidedPurchaseListResultFromAPI(options VoidedPurchaseListOptions, response *androidpublisher.VoidedPurchasesListResponse) VoidedPurchaseListResult {
	result := VoidedPurchaseListResult{
		PackageName: options.PackageName,
		Options:     options,
		Purchases:   []VoidedPurchase{},
	}
	if response == nil {
		return result
	}
	for _, apiPurchase := range response.VoidedPurchases {
		if apiPurchase == nil {
			continue
		}
		result.Purchases = append(result.Purchases, voidedPurchaseFromAPI(apiPurchase))
	}
	if response.PageInfo != nil {
		result.PageInfo = &VoidedPurchasePageInfo{
			ResultPerPage: response.PageInfo.ResultPerPage,
			StartIndex:    response.PageInfo.StartIndex,
			TotalResults:  response.PageInfo.TotalResults,
		}
	}
	if response.TokenPagination != nil {
		result.Pagination = &VoidedPurchasePagination{
			NextPageToken:     response.TokenPagination.NextPageToken,
			PreviousPageToken: response.TokenPagination.PreviousPageToken,
		}
	}
	return result
}

func voidedPurchaseFromAPI(apiPurchase *androidpublisher.VoidedPurchase) VoidedPurchase {
	return VoidedPurchase{
		OrderID:            apiPurchase.OrderId,
		PurchaseToken:      apiPurchase.PurchaseToken,
		PurchaseTimeMillis: apiPurchase.PurchaseTimeMillis,
		VoidedTimeMillis:   apiPurchase.VoidedTimeMillis,
		VoidedReason:       apiPurchase.VoidedReason,
		VoidedSource:       apiPurchase.VoidedSource,
		VoidedQuantity:     apiPurchase.VoidedQuantity,
	}
}
