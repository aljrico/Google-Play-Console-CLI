package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) GetOrder(ctx context.Context, options OrderGetOptions) (OrderGetResult, error) {
	apiOrder, err := p.service.Orders.Get(options.PackageName.String(), options.OrderID.String()).Context(ctx).Do()
	if err != nil {
		return OrderGetResult{}, fmt.Errorf("get order %s for %s: %w", options.OrderID, options.PackageName, err)
	}
	return OrderGetResult{
		PackageName: options.PackageName,
		OrderID:     options.OrderID,
		Order:       orderFromAPI(apiOrder),
	}, nil
}

func (p *GooglePublisher) BatchGetOrders(ctx context.Context, options OrderBatchGetOptions) (OrderBatchGetResult, error) {
	response, err := p.service.Orders.Batchget(options.PackageName.String()).
		OrderIds(orderIDStrings(options.OrderIDs)...).
		Context(ctx).
		Do()
	if err != nil {
		return OrderBatchGetResult{}, fmt.Errorf("batch get orders for %s: %w", options.PackageName, err)
	}
	return orderBatchGetResultFromAPI(options, response), nil
}

func (p *GooglePublisher) RefundOrder(ctx context.Context, options OrderRefundOptions) error {
	call := p.service.Orders.Refund(options.PackageName.String(), options.OrderID.String()).Context(ctx)
	if options.Revoke {
		call.Revoke(true)
	}
	if err := call.Do(); err != nil {
		return fmt.Errorf("refund order %s for %s: %w", options.OrderID, options.PackageName, err)
	}
	return nil
}

func orderBatchGetResultFromAPI(options OrderBatchGetOptions, response *androidpublisher.BatchGetOrdersResponse) OrderBatchGetResult {
	result := OrderBatchGetResult{
		PackageName: options.PackageName,
		OrderIDs:    append([]OrderID(nil), options.OrderIDs...),
		Orders:      []Order{},
	}
	if response == nil {
		return result
	}
	for _, apiOrder := range response.Orders {
		if apiOrder == nil {
			continue
		}
		result.Orders = append(result.Orders, orderFromAPI(apiOrder))
	}
	return result
}

func orderFromAPI(apiOrder *androidpublisher.Order) Order {
	if apiOrder == nil {
		return Order{LineItems: []OrderLineItem{}}
	}
	lineItems := make([]OrderLineItem, 0, len(apiOrder.LineItems))
	for _, apiLineItem := range apiOrder.LineItems {
		if apiLineItem == nil {
			continue
		}
		lineItems = append(lineItems, orderLineItemFromAPI(apiLineItem))
	}
	return Order{
		OrderID:                         apiOrder.OrderId,
		PurchaseToken:                   apiOrder.PurchaseToken,
		State:                           apiOrder.State,
		CreateTime:                      apiOrder.CreateTime,
		LastEventTime:                   apiOrder.LastEventTime,
		BuyerAddress:                    buyerAddressFromAPI(apiOrder.BuyerAddress),
		Total:                           moneyFromAPI(apiOrder.Total),
		Tax:                             moneyFromAPI(apiOrder.Tax),
		DeveloperRevenueInBuyerCurrency: moneyFromAPI(apiOrder.DeveloperRevenueInBuyerCurrency),
		OrderDetails:                    orderDetailsFromAPI(apiOrder.OrderDetails),
		OrderHistory:                    orderHistoryFromAPI(apiOrder.OrderHistory),
		PointsDetails:                   pointsDetailsFromAPI(apiOrder.PointsDetails),
		LineItems:                       lineItems,
	}
}

func orderLineItemFromAPI(apiLineItem *androidpublisher.LineItem) OrderLineItem {
	return OrderLineItem{
		ProductID:              apiLineItem.ProductId,
		ProductTitle:           apiLineItem.ProductTitle,
		ListingPrice:           moneyFromAPI(apiLineItem.ListingPrice),
		Tax:                    moneyFromAPI(apiLineItem.Tax),
		Total:                  moneyFromAPI(apiLineItem.Total),
		OneTimePurchaseDetails: oneTimePurchaseDetailsFromAPI(apiLineItem.OneTimePurchaseDetails),
		PaidAppDetails:         paidAppDetailsFromAPI(apiLineItem.PaidAppDetails),
		SubscriptionDetails:    orderSubscriptionDetailsFromAPI(apiLineItem.SubscriptionDetails),
	}
}

func oneTimePurchaseDetailsFromAPI(apiDetails *androidpublisher.OneTimePurchaseDetails) *OneTimePurchaseDetails {
	if apiDetails == nil {
		return nil
	}
	return &OneTimePurchaseDetails{
		OfferID:  apiDetails.OfferId,
		Quantity: apiDetails.Quantity,
	}
}

func paidAppDetailsFromAPI(apiDetails *androidpublisher.PaidAppDetails) *PaidAppDetails {
	if apiDetails == nil {
		return nil
	}
	return &PaidAppDetails{}
}

func orderDetailsFromAPI(apiDetails *androidpublisher.OrderDetails) *OrderDetails {
	if apiDetails == nil {
		return nil
	}
	return &OrderDetails{TaxInclusive: apiDetails.TaxInclusive}
}

func orderHistoryFromAPI(apiHistory *androidpublisher.OrderHistory) *OrderHistory {
	if apiHistory == nil {
		return nil
	}
	partialRefundEvents := make([]PartialRefundEvent, 0, len(apiHistory.PartialRefundEvents))
	for _, apiEvent := range apiHistory.PartialRefundEvents {
		if apiEvent == nil {
			continue
		}
		partialRefundEvents = append(partialRefundEvents, partialRefundEventFromAPI(apiEvent))
	}
	return &OrderHistory{
		CancellationEvent:   cancellationEventFromAPI(apiHistory.CancellationEvent),
		ProcessedEvent:      processedEventFromAPI(apiHistory.ProcessedEvent),
		RefundEvent:         orderRefundEventFromAPI(apiHistory.RefundEvent),
		PartialRefundEvents: partialRefundEvents,
	}
}

func refundDetailsFromAPI(apiDetails *androidpublisher.RefundDetails) *RefundDetails {
	if apiDetails == nil {
		return nil
	}
	return &RefundDetails{
		Tax:   moneyFromAPI(apiDetails.Tax),
		Total: moneyFromAPI(apiDetails.Total),
	}
}

func pointsDetailsFromAPI(apiDetails *androidpublisher.PointsDetails) *PointsDetails {
	if apiDetails == nil {
		return nil
	}
	return &PointsDetails{
		PointsOfferID:            apiDetails.PointsOfferId,
		PointsSpent:              apiDetails.PointsSpent,
		PointsDiscountRateMicros: apiDetails.PointsDiscountRateMicros,
		PointsCouponValue:        moneyFromAPI(apiDetails.PointsCouponValue),
	}
}

func buyerAddressFromAPI(apiAddress *androidpublisher.BuyerAddress) *BuyerAddress {
	if apiAddress == nil {
		return nil
	}
	return &BuyerAddress{
		Country:  apiAddress.BuyerCountry,
		State:    apiAddress.BuyerState,
		Postcode: apiAddress.BuyerPostcode,
	}
}

func orderIDStrings(orderIDs []OrderID) []string {
	values := make([]string, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		values = append(values, orderID.String())
	}
	return values
}
