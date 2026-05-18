package play

import (
	"context"
	"fmt"
)

type PurchaseToken string

func NewPurchaseToken(value string) (PurchaseToken, error) {
	if value == "" {
		return "", fmt.Errorf("purchase token is required")
	}
	return PurchaseToken(value), nil
}

func (t PurchaseToken) String() string {
	return string(t)
}

type ProductPurchaseOptions struct {
	PackageName PackageName     `json:"packageName"`
	ProductID   InAppProductSKU `json:"productId"`
	Token       PurchaseToken   `json:"token"`
}

func (o ProductPurchaseOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewInAppProductSKU(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewPurchaseToken(o.Token.String()); err != nil {
		return err
	}
	return nil
}

type ProductPurchase struct {
	PackageName                 PackageName     `json:"packageName"`
	ProductID                   InAppProductSKU `json:"productId"`
	Token                       PurchaseToken   `json:"token,omitempty"`
	OrderID                     string          `json:"orderId,omitempty"`
	PurchaseState               int64           `json:"purchaseState,omitempty"`
	PurchaseTimeMillis          int64           `json:"purchaseTimeMillis,omitempty"`
	AcknowledgementState        int64           `json:"acknowledgementState,omitempty"`
	ConsumptionState            int64           `json:"consumptionState,omitempty"`
	Quantity                    int64           `json:"quantity,omitempty"`
	RefundableQuantity          int64           `json:"refundableQuantity,omitempty"`
	RegionCode                  string          `json:"regionCode,omitempty"`
	DeveloperPayload            string          `json:"developerPayload,omitempty"`
	ObfuscatedExternalAccountID string          `json:"obfuscatedExternalAccountId,omitempty"`
	ObfuscatedExternalProfileID string          `json:"obfuscatedExternalProfileId,omitempty"`
}

type ProductPurchaseGetter interface {
	GetProductPurchase(ctx context.Context, options ProductPurchaseOptions) (ProductPurchase, error)
}

func GetProductPurchase(ctx context.Context, getter ProductPurchaseGetter, options ProductPurchaseOptions) (ProductPurchase, error) {
	if err := options.Validate(); err != nil {
		return ProductPurchase{}, err
	}
	if getter == nil {
		return ProductPurchase{}, fmt.Errorf("product purchase getter is required")
	}
	return getter.GetProductPurchase(ctx, options)
}

type SubscriptionPurchaseOptions struct {
	PackageName PackageName   `json:"packageName"`
	Token       PurchaseToken `json:"token"`
}

func (o SubscriptionPurchaseOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewPurchaseToken(o.Token.String()); err != nil {
		return err
	}
	return nil
}

type SubscriptionPurchase struct {
	PackageName                 PackageName                    `json:"packageName"`
	Token                       PurchaseToken                  `json:"token,omitempty"`
	SubscriptionState           string                         `json:"subscriptionState,omitempty"`
	AcknowledgementState        string                         `json:"acknowledgementState,omitempty"`
	LatestOrderID               string                         `json:"latestOrderId,omitempty"`
	LinkedPurchaseToken         string                         `json:"linkedPurchaseToken,omitempty"`
	RegionCode                  string                         `json:"regionCode,omitempty"`
	StartTime                   string                         `json:"startTime,omitempty"`
	LineItems                   []SubscriptionPurchaseLineItem `json:"lineItems"`
	ExternalAccountID           string                         `json:"externalAccountId,omitempty"`
	ObfuscatedExternalAccountID string                         `json:"obfuscatedExternalAccountId,omitempty"`
	ObfuscatedExternalProfileID string                         `json:"obfuscatedExternalProfileId,omitempty"`
	TestPurchase                bool                           `json:"testPurchase,omitempty"`
}

type SubscriptionPurchaseLineItem struct {
	ProductID               string   `json:"productId,omitempty"`
	ExpiryTime              string   `json:"expiryTime,omitempty"`
	LatestSuccessfulOrderID string   `json:"latestSuccessfulOrderId,omitempty"`
	BasePlanID              string   `json:"basePlanId,omitempty"`
	OfferID                 string   `json:"offerId,omitempty"`
	OfferTags               []string `json:"offerTags,omitempty"`
	AutoRenewEnabled        *bool    `json:"autoRenewEnabled,omitempty"`
	RecurringPrice          *Money   `json:"recurringPrice,omitempty"`
	Prepaid                 bool     `json:"prepaid,omitempty"`
	AllowExtendAfterTime    string   `json:"allowExtendAfterTime,omitempty"`
}

type SubscriptionPurchaseGetter interface {
	GetSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseOptions) (SubscriptionPurchase, error)
}

func GetSubscriptionPurchase(ctx context.Context, getter SubscriptionPurchaseGetter, options SubscriptionPurchaseOptions) (SubscriptionPurchase, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionPurchase{}, err
	}
	if getter == nil {
		return SubscriptionPurchase{}, fmt.Errorf("subscription purchase getter is required")
	}
	return getter.GetSubscriptionPurchase(ctx, options)
}
