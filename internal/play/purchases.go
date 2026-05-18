package play

import (
	"context"
	"fmt"
	"strings"
	"time"
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
	ProductID   InAppProductSKU `json:"productId,omitempty"`
	Token       PurchaseToken   `json:"token"`
}

func (o ProductPurchaseOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.ProductID != "" {
		if _, err := NewInAppProductSKU(o.ProductID.String()); err != nil {
			return err
		}
	}
	if _, err := NewPurchaseToken(o.Token.String()); err != nil {
		return err
	}
	return nil
}

type ProductPurchase struct {
	PackageName                 PackageName               `json:"packageName"`
	ProductID                   InAppProductSKU           `json:"productId,omitempty"`
	Token                       PurchaseToken             `json:"token,omitempty"`
	OrderID                     string                    `json:"orderId,omitempty"`
	PurchaseState               string                    `json:"purchaseState,omitempty"`
	PurchaseCompletionTime      string                    `json:"purchaseCompletionTime,omitempty"`
	AcknowledgementState        string                    `json:"acknowledgementState,omitempty"`
	RegionCode                  string                    `json:"regionCode,omitempty"`
	ObfuscatedExternalAccountID string                    `json:"obfuscatedExternalAccountId,omitempty"`
	ObfuscatedExternalProfileID string                    `json:"obfuscatedExternalProfileId,omitempty"`
	TestPurchase                bool                      `json:"testPurchase,omitempty"`
	LineItems                   []ProductPurchaseLineItem `json:"lineItems"`
}

type ProductPurchaseLineItem struct {
	ProductID          string   `json:"productId,omitempty"`
	ConsumptionState   string   `json:"consumptionState,omitempty"`
	PurchaseOptionID   string   `json:"purchaseOptionId,omitempty"`
	OfferID            string   `json:"offerId,omitempty"`
	OfferToken         string   `json:"offerToken,omitempty"`
	OfferTags          []string `json:"offerTags,omitempty"`
	Quantity           int64    `json:"quantity"`
	RefundableQuantity int64    `json:"refundableQuantity"`
}

type ProductPurchaseGetter interface {
	GetProductPurchase(ctx context.Context, options ProductPurchaseOptions) (ProductPurchase, error)
}

type ProductPurchaseMutator interface {
	AcknowledgeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error
	ConsumeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error
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

type ProductPurchaseMutationOptions struct {
	PackageName      PackageName     `json:"packageName"`
	ProductID        InAppProductSKU `json:"productId"`
	Token            PurchaseToken   `json:"token"`
	DeveloperPayload string          `json:"developerPayload,omitempty"`
	Confirm          bool            `json:"confirm"`
	DryRun           bool            `json:"dryRun"`
}

func (o ProductPurchaseMutationOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewInAppProductSKU(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewPurchaseToken(o.Token.String()); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("product purchase mutation requires --confirm or --dry-run")
	}
	return nil
}

func (o ProductPurchaseMutationOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live product purchase mutation cannot run with --dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live product purchase mutation requires --confirm")
	}
	return nil
}

type ProductPurchaseMutationPlan struct {
	Action           string          `json:"action"`
	PackageName      PackageName     `json:"packageName"`
	ProductID        InAppProductSKU `json:"productId"`
	Token            PurchaseToken   `json:"token"`
	DeveloperPayload string          `json:"developerPayload,omitempty"`
	Confirm          bool            `json:"confirm"`
	Steps            []string        `json:"steps"`
}

type ProductPurchaseMutationResult struct {
	Action      string                      `json:"action"`
	PackageName PackageName                 `json:"packageName"`
	ProductID   InAppProductSKU             `json:"productId"`
	Token       PurchaseToken               `json:"token"`
	DryRun      bool                        `json:"dryRun"`
	Applied     bool                        `json:"applied"`
	Plan        ProductPurchaseMutationPlan `json:"plan"`
}

func AcknowledgeProductPurchase(ctx context.Context, mutator ProductPurchaseMutator, options ProductPurchaseMutationOptions) (ProductPurchaseMutationResult, error) {
	return mutateProductPurchase(ctx, mutator, options, "acknowledge")
}

func ConsumeProductPurchase(ctx context.Context, mutator ProductPurchaseMutator, options ProductPurchaseMutationOptions) (ProductPurchaseMutationResult, error) {
	return mutateProductPurchase(ctx, mutator, options, "consume")
}

func mutateProductPurchase(ctx context.Context, mutator ProductPurchaseMutator, options ProductPurchaseMutationOptions, action string) (ProductPurchaseMutationResult, error) {
	if err := options.Validate(); err != nil {
		return ProductPurchaseMutationResult{}, err
	}
	steps := []string{action + " product purchase"}
	result := ProductPurchaseMutationResult{
		Action:      action,
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		Token:       options.Token,
		DryRun:      options.DryRun,
		Plan: ProductPurchaseMutationPlan{
			Action:           action,
			PackageName:      options.PackageName,
			ProductID:        options.ProductID,
			Token:            options.Token,
			DeveloperPayload: options.DeveloperPayload,
			Confirm:          options.Confirm,
			Steps:            steps,
		},
	}
	if options.DryRun {
		return result, nil
	}
	if mutator == nil {
		return ProductPurchaseMutationResult{}, fmt.Errorf("product purchase mutator is required")
	}
	switch action {
	case "acknowledge":
		if err := mutator.AcknowledgeProductPurchase(ctx, options); err != nil {
			return ProductPurchaseMutationResult{}, err
		}
	case "consume":
		if err := mutator.ConsumeProductPurchase(ctx, options); err != nil {
			return ProductPurchaseMutationResult{}, err
		}
	default:
		return ProductPurchaseMutationResult{}, fmt.Errorf("unsupported product purchase action %q", action)
	}
	result.Applied = true
	return result, nil
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

type SubscriptionPurchaseRevoker interface {
	RevokeSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseRevokeOptions) error
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

type SubscriptionRefundType string

const (
	SubscriptionRefundTypeFull     SubscriptionRefundType = "full"
	SubscriptionRefundTypeProrated SubscriptionRefundType = "prorated"
)

func NewSubscriptionRefundType(value string) (SubscriptionRefundType, error) {
	refundType := SubscriptionRefundType(strings.TrimSpace(value))
	switch refundType {
	case SubscriptionRefundTypeFull, SubscriptionRefundTypeProrated:
		return refundType, nil
	default:
		return "", fmt.Errorf("unsupported subscription refund type %q", value)
	}
}

func (t SubscriptionRefundType) String() string {
	return string(t)
}

func (t SubscriptionRefundType) Validate() error {
	_, err := NewSubscriptionRefundType(t.String())
	return err
}

type SubscriptionPurchaseRevokeOptions struct {
	PackageName PackageName            `json:"packageName"`
	Token       PurchaseToken          `json:"token"`
	RefundType  SubscriptionRefundType `json:"refundType"`
	Confirm     bool                   `json:"confirm"`
	DryRun      bool                   `json:"dryRun"`
}

func (o SubscriptionPurchaseRevokeOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewPurchaseToken(o.Token.String()); err != nil {
		return err
	}
	if err := o.RefundType.Validate(); err != nil {
		return err
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("subscription purchase revoke requires --confirm or --dry-run")
	}
	return nil
}

func (o SubscriptionPurchaseRevokeOptions) ValidateLive() error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.DryRun {
		return fmt.Errorf("live subscription purchase revoke cannot run with --dry-run")
	}
	if !o.Confirm {
		return fmt.Errorf("live subscription purchase revoke requires --confirm")
	}
	return nil
}

type SubscriptionPurchaseRevokePlan struct {
	PackageName PackageName            `json:"packageName"`
	Token       PurchaseToken          `json:"token"`
	RefundType  SubscriptionRefundType `json:"refundType"`
	Confirm     bool                   `json:"confirm"`
	Steps       []string               `json:"steps"`
}

type SubscriptionPurchaseRevokeResult struct {
	PackageName PackageName                    `json:"packageName"`
	Token       PurchaseToken                  `json:"token"`
	RefundType  SubscriptionRefundType         `json:"refundType"`
	DryRun      bool                           `json:"dryRun"`
	Applied     bool                           `json:"applied"`
	Plan        SubscriptionPurchaseRevokePlan `json:"plan"`
}

func RevokeSubscriptionPurchase(ctx context.Context, revoker SubscriptionPurchaseRevoker, options SubscriptionPurchaseRevokeOptions) (SubscriptionPurchaseRevokeResult, error) {
	if err := options.Validate(); err != nil {
		return SubscriptionPurchaseRevokeResult{}, err
	}
	result := SubscriptionPurchaseRevokeResult{
		PackageName: options.PackageName,
		Token:       options.Token,
		RefundType:  options.RefundType,
		DryRun:      options.DryRun,
		Plan: SubscriptionPurchaseRevokePlan{
			PackageName: options.PackageName,
			Token:       options.Token,
			RefundType:  options.RefundType,
			Confirm:     options.Confirm,
			Steps:       []string{"revoke subscription purchase", string(options.RefundType) + " refund"},
		},
	}
	if options.DryRun {
		return result, nil
	}
	if revoker == nil {
		return SubscriptionPurchaseRevokeResult{}, fmt.Errorf("subscription purchase revoker is required")
	}
	if err := revoker.RevokeSubscriptionPurchase(ctx, options); err != nil {
		return SubscriptionPurchaseRevokeResult{}, err
	}
	result.Applied = true
	return result, nil
}

type VoidedPurchaseType int64

const (
	VoidedPurchaseTypeProductsOnly          VoidedPurchaseType = 0
	VoidedPurchaseTypeProductsSubscriptions VoidedPurchaseType = 1
	voidedPurchaseWindow                    time.Duration      = 30 * 24 * time.Hour
)

type VoidedPurchaseListOptions struct {
	PackageName                       PackageName        `json:"packageName"`
	MaxResults                        int64              `json:"maxResults,omitempty"`
	StartIndex                        int64              `json:"startIndex,omitempty"`
	Token                             string             `json:"token,omitempty"`
	StartTimeMillis                   int64              `json:"startTimeMillis,omitempty"`
	EndTimeMillis                     int64              `json:"endTimeMillis,omitempty"`
	Type                              VoidedPurchaseType `json:"type,omitempty"`
	IncludeQuantityBasedPartialRefund bool               `json:"includeQuantityBasedPartialRefund,omitempty"`
}

func (o VoidedPurchaseListOptions) Validate() error {
	return o.ValidateAt(time.Now())
}

func (o VoidedPurchaseListOptions) ValidateAt(now time.Time) error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.MaxResults < 0 {
		return fmt.Errorf("max results cannot be negative")
	}
	if o.StartIndex < 0 {
		return fmt.Errorf("start index cannot be negative")
	}
	if o.StartTimeMillis < 0 {
		return fmt.Errorf("start time cannot be negative")
	}
	if o.EndTimeMillis < 0 {
		return fmt.Errorf("end time cannot be negative")
	}
	if o.Token != "" && (o.StartTimeMillis > 0 || o.EndTimeMillis > 0) {
		return fmt.Errorf("start time and end time cannot be used with a pagination token")
	}
	if o.StartTimeMillis > 0 && o.EndTimeMillis > 0 && o.StartTimeMillis > o.EndTimeMillis {
		return fmt.Errorf("start time cannot be after end time")
	}
	nowMillis := now.UnixMilli()
	if o.StartTimeMillis > 0 && o.StartTimeMillis < now.Add(-voidedPurchaseWindow).UnixMilli() {
		return fmt.Errorf("start time cannot be older than 30 days")
	}
	if o.EndTimeMillis > nowMillis {
		return fmt.Errorf("end time cannot be in the future")
	}
	if o.Type != VoidedPurchaseTypeProductsOnly && o.Type != VoidedPurchaseTypeProductsSubscriptions {
		return fmt.Errorf("voided purchase type must be 0 or 1")
	}
	return nil
}

type VoidedPurchase struct {
	OrderID            string `json:"orderId,omitempty"`
	PurchaseToken      string `json:"purchaseToken,omitempty"`
	PurchaseTimeMillis int64  `json:"purchaseTimeMillis,omitempty"`
	VoidedTimeMillis   int64  `json:"voidedTimeMillis,omitempty"`
	VoidedReason       int64  `json:"voidedReason"`
	VoidedSource       int64  `json:"voidedSource"`
	VoidedQuantity     int64  `json:"voidedQuantity"`
}

type VoidedPurchasePageInfo struct {
	ResultPerPage int64 `json:"resultPerPage,omitempty"`
	StartIndex    int64 `json:"startIndex,omitempty"`
	TotalResults  int64 `json:"totalResults,omitempty"`
}

type VoidedPurchasePagination struct {
	NextPageToken     string `json:"nextPageToken,omitempty"`
	PreviousPageToken string `json:"previousPageToken,omitempty"`
}

type VoidedPurchaseListResult struct {
	PackageName PackageName               `json:"packageName"`
	Options     VoidedPurchaseListOptions `json:"options"`
	PageInfo    *VoidedPurchasePageInfo   `json:"pageInfo,omitempty"`
	Pagination  *VoidedPurchasePagination `json:"pagination,omitempty"`
	Purchases   []VoidedPurchase          `json:"purchases"`
}

type VoidedPurchaseLister interface {
	ListVoidedPurchases(ctx context.Context, options VoidedPurchaseListOptions) (VoidedPurchaseListResult, error)
}

func ListVoidedPurchases(ctx context.Context, lister VoidedPurchaseLister, options VoidedPurchaseListOptions) (VoidedPurchaseListResult, error) {
	if err := options.Validate(); err != nil {
		return VoidedPurchaseListResult{}, err
	}
	if lister == nil {
		return VoidedPurchaseListResult{}, fmt.Errorf("voided purchase lister is required")
	}
	return lister.ListVoidedPurchases(ctx, options)
}
