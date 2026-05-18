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
	ProductID    string `json:"productId,omitempty"`
	ProductTitle string `json:"productTitle,omitempty"`
	ListingPrice *Money `json:"listingPrice,omitempty"`
	Tax          *Money `json:"tax,omitempty"`
	Total        *Money `json:"total,omitempty"`
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
