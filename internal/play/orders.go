package play

import (
	"context"
	"fmt"
	"strings"
)

type OrderID string

func NewOrderID(value string) (OrderID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("order ID is required")
	}
	return OrderID(value), nil
}

func (o OrderID) String() string {
	return string(o)
}

func (o OrderID) Validate() error {
	_, err := NewOrderID(o.String())
	return err
}

type BuyerAddress struct {
	Country  string `json:"buyerCountry,omitempty"`
	State    string `json:"buyerState,omitempty"`
	Postcode string `json:"buyerPostcode,omitempty"`
}

type OrderLineItem struct {
	ProductID              string                    `json:"productId,omitempty"`
	ProductTitle           string                    `json:"productTitle,omitempty"`
	ListingPrice           *Money                    `json:"listingPrice,omitempty"`
	Tax                    *Money                    `json:"tax,omitempty"`
	Total                  *Money                    `json:"total,omitempty"`
	OneTimePurchaseDetails *OneTimePurchaseDetails   `json:"oneTimePurchaseDetails,omitempty"`
	PaidAppDetails         *PaidAppDetails           `json:"paidAppDetails,omitempty"`
	SubscriptionDetails    *OrderSubscriptionDetails `json:"subscriptionDetails,omitempty"`
}

type OneTimePurchaseDetails struct {
	OfferID  string `json:"offerId,omitempty"`
	Quantity int64  `json:"quantity,omitempty"`
}

type PaidAppDetails struct{}

type OrderSubscriptionDetails struct {
	BasePlanID             string `json:"basePlanId,omitempty"`
	OfferID                string `json:"offerId,omitempty"`
	OfferPhase             string `json:"offerPhase,omitempty"`
	ServicePeriodStartTime string `json:"servicePeriodStartTime,omitempty"`
	ServicePeriodEndTime   string `json:"servicePeriodEndTime,omitempty"`
}

type OrderDetails struct {
	TaxInclusive bool `json:"taxInclusive"`
}

type OrderHistory struct {
	CancellationEvent   *OrderEvent          `json:"cancellationEvent,omitempty"`
	ProcessedEvent      *OrderEvent          `json:"processedEvent,omitempty"`
	RefundEvent         *OrderRefundEvent    `json:"refundEvent,omitempty"`
	PartialRefundEvents []PartialRefundEvent `json:"partialRefundEvents,omitempty"`
}

type OrderEvent struct {
	EventTime string `json:"eventTime,omitempty"`
}

type OrderRefundEvent struct {
	EventTime     string         `json:"eventTime,omitempty"`
	RefundReason  string         `json:"refundReason,omitempty"`
	RefundDetails *RefundDetails `json:"refundDetails,omitempty"`
}

type PartialRefundEvent struct {
	CreateTime    string         `json:"createTime,omitempty"`
	ProcessTime   string         `json:"processTime,omitempty"`
	State         string         `json:"state,omitempty"`
	RefundDetails *RefundDetails `json:"refundDetails,omitempty"`
}

type RefundDetails struct {
	Tax   *Money `json:"tax,omitempty"`
	Total *Money `json:"total,omitempty"`
}

type PointsDetails struct {
	PointsOfferID            string `json:"pointsOfferId,omitempty"`
	PointsSpent              int64  `json:"pointsSpent,omitempty"`
	PointsDiscountRateMicros int64  `json:"pointsDiscountRateMicros,omitempty"`
	PointsCouponValue        *Money `json:"pointsCouponValue,omitempty"`
}

type Order struct {
	OrderID                         string          `json:"orderId,omitempty"`
	PurchaseToken                   string          `json:"purchaseToken,omitempty"`
	State                           string          `json:"state,omitempty"`
	CreateTime                      string          `json:"createTime,omitempty"`
	LastEventTime                   string          `json:"lastEventTime,omitempty"`
	BuyerAddress                    *BuyerAddress   `json:"buyerAddress,omitempty"`
	Total                           *Money          `json:"total,omitempty"`
	Tax                             *Money          `json:"tax,omitempty"`
	DeveloperRevenueInBuyerCurrency *Money          `json:"developerRevenueInBuyerCurrency,omitempty"`
	OrderDetails                    *OrderDetails   `json:"orderDetails,omitempty"`
	OrderHistory                    *OrderHistory   `json:"orderHistory,omitempty"`
	PointsDetails                   *PointsDetails  `json:"pointsDetails,omitempty"`
	LineItems                       []OrderLineItem `json:"lineItems"`
}

type OrderGetOptions struct {
	PackageName PackageName `json:"packageName"`
	OrderID     OrderID     `json:"orderId"`
}

func (o OrderGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	return o.OrderID.Validate()
}

type OrderGetResult struct {
	PackageName PackageName `json:"packageName"`
	OrderID     OrderID     `json:"orderId"`
	Order       Order       `json:"order"`
}

type OrderBatchGetOptions struct {
	PackageName PackageName `json:"packageName"`
	OrderIDs    []OrderID   `json:"orderIds"`
}

func (o OrderBatchGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if len(o.OrderIDs) == 0 {
		return fmt.Errorf("at least one order ID is required")
	}
	if len(o.OrderIDs) > 1000 {
		return fmt.Errorf("at most 1000 order IDs are allowed")
	}
	seen := make(map[OrderID]struct{}, len(o.OrderIDs))
	for _, orderID := range o.OrderIDs {
		if err := orderID.Validate(); err != nil {
			return err
		}
		if _, ok := seen[orderID]; ok {
			return fmt.Errorf("order ID %q is duplicated", orderID)
		}
		seen[orderID] = struct{}{}
	}
	return nil
}

type OrderBatchGetResult struct {
	PackageName PackageName `json:"packageName"`
	OrderIDs    []OrderID   `json:"orderIds"`
	Orders      []Order     `json:"orders"`
}

type OrderGetter interface {
	GetOrder(ctx context.Context, options OrderGetOptions) (OrderGetResult, error)
}

type OrderBatchGetter interface {
	BatchGetOrders(ctx context.Context, options OrderBatchGetOptions) (OrderBatchGetResult, error)
}

type OrderRefunder interface {
	RefundOrder(ctx context.Context, options OrderRefundOptions) error
}

func GetOrder(ctx context.Context, getter OrderGetter, options OrderGetOptions) (OrderGetResult, error) {
	if err := options.Validate(); err != nil {
		return OrderGetResult{}, err
	}
	if getter == nil {
		return OrderGetResult{}, fmt.Errorf("order getter is required")
	}
	return getter.GetOrder(ctx, options)
}

func BatchGetOrders(ctx context.Context, getter OrderBatchGetter, options OrderBatchGetOptions) (OrderBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return OrderBatchGetResult{}, err
	}
	if getter == nil {
		return OrderBatchGetResult{}, fmt.Errorf("order batch getter is required")
	}
	return getter.BatchGetOrders(ctx, options)
}

type OrderRefundOptions struct {
	PackageName PackageName `json:"packageName"`
	OrderID     OrderID     `json:"orderId"`
	Revoke      bool        `json:"revoke"`
	Confirm     bool        `json:"confirm"`
	DryRun      bool        `json:"dryRun"`
}

func (o OrderRefundOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := o.OrderID.Validate(); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("order refund requires --confirm or --dry-run")
	}
	return nil
}

type OrderRefundPlan struct {
	PackageName PackageName `json:"packageName"`
	OrderID     OrderID     `json:"orderId"`
	Revoke      bool        `json:"revoke"`
	Confirm     bool        `json:"confirm"`
	Steps       []string    `json:"steps"`
}

type OrderRefundResult struct {
	PackageName PackageName     `json:"packageName"`
	OrderID     OrderID         `json:"orderId"`
	Revoke      bool            `json:"revoke"`
	DryRun      bool            `json:"dryRun"`
	Applied     bool            `json:"applied"`
	Plan        OrderRefundPlan `json:"plan"`
}

func NewOrderRefundPlan(options OrderRefundOptions) (OrderRefundPlan, error) {
	if err := options.Validate(); err != nil {
		return OrderRefundPlan{}, err
	}
	steps := []string{"refund order"}
	if options.Revoke {
		steps = append(steps, "revoke purchased item")
	}
	return OrderRefundPlan{
		PackageName: options.PackageName,
		OrderID:     options.OrderID,
		Revoke:      options.Revoke,
		Confirm:     options.Confirm,
		Steps:       steps,
	}, nil
}

func RefundOrder(ctx context.Context, refunder OrderRefunder, options OrderRefundOptions) (OrderRefundResult, error) {
	plan, err := NewOrderRefundPlan(options)
	if err != nil {
		return OrderRefundResult{}, err
	}
	result := OrderRefundResult{
		PackageName: options.PackageName,
		OrderID:     options.OrderID,
		Revoke:      options.Revoke,
		DryRun:      options.DryRun,
		Applied:     false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if refunder == nil {
		return OrderRefundResult{}, fmt.Errorf("order refunder is required")
	}
	if err := refunder.RefundOrder(ctx, options); err != nil {
		return OrderRefundResult{}, err
	}
	result.Applied = true
	return result, nil
}
