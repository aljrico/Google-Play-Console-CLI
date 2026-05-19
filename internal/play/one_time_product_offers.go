package play

import (
	"context"
	"fmt"
)

const OneTimeProductOfferWildcardID = "-"

type OneTimeProductPurchaseOptionID string

func NewOneTimeProductPurchaseOptionID(value string) (OneTimeProductPurchaseOptionID, error) {
	if value == "" {
		return "", fmt.Errorf("one-time product purchase option ID is required")
	}
	if !isValidOneTimeProductDashID(value) {
		return "", fmt.Errorf("invalid one-time product purchase option ID %q", value)
	}
	return OneTimeProductPurchaseOptionID(value), nil
}

func NewOneTimeProductOfferListProductID(value string) (OneTimeProductID, error) {
	if value == OneTimeProductOfferWildcardID {
		return OneTimeProductID(value), nil
	}
	return NewOneTimeProductID(value)
}

func NewOneTimeProductOfferListPurchaseOptionID(value string) (OneTimeProductPurchaseOptionID, error) {
	if value == OneTimeProductOfferWildcardID {
		return OneTimeProductPurchaseOptionID(value), nil
	}
	return NewOneTimeProductPurchaseOptionID(value)
}

func (p OneTimeProductPurchaseOptionID) String() string {
	return string(p)
}

type OneTimeProductOfferID string

func NewOneTimeProductOfferID(value string) (OneTimeProductOfferID, error) {
	if value == "" {
		return "", fmt.Errorf("one-time product offer ID is required")
	}
	if !isValidOneTimeProductDashID(value) {
		return "", fmt.Errorf("invalid one-time product offer ID %q", value)
	}
	return OneTimeProductOfferID(value), nil
}

func (o OneTimeProductOfferID) String() string {
	return string(o)
}

func isValidOneTimeProductDashID(value string) bool {
	if len(value) > 63 {
		return false
	}
	first := value[0]
	if !isASCIILower(first) && !isASCIIDigit(first) {
		return false
	}
	for i := 1; i < len(value); i++ {
		character := value[i]
		if !isASCIILower(character) && !isASCIIDigit(character) && character != '-' {
			return false
		}
	}
	return true
}

type OneTimeProductOfferType string

const (
	OneTimeProductOfferTypeDiscounted OneTimeProductOfferType = "discounted"
	OneTimeProductOfferTypePreOrder   OneTimeProductOfferType = "preOrder"
)

type OneTimeProductOffer struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
	State            string                         `json:"state,omitempty"`
	Type             OneTimeProductOfferType        `json:"type,omitempty"`
	OfferTags        []string                       `json:"offerTags,omitempty"`
	RegionsVersion   *RegionsVersion                `json:"regionsVersion,omitempty"`
	DiscountedOffer  *OneTimeProductDiscountedOffer `json:"discountedOffer,omitempty"`
	PreOrderOffer    *OneTimeProductPreOrderOffer   `json:"preOrderOffer,omitempty"`
	RegionalConfigs  []OneTimeProductOfferRegion    `json:"regionalConfigs,omitempty"`
}

type OneTimeProductDiscountedOffer struct {
	StartTime       string `json:"startTime,omitempty"`
	EndTime         string `json:"endTime,omitempty"`
	RedemptionLimit int64  `json:"redemptionLimit,omitempty"`
}

type OneTimeProductPreOrderOffer struct {
	StartTime           string `json:"startTime,omitempty"`
	EndTime             string `json:"endTime,omitempty"`
	ReleaseTime         string `json:"releaseTime,omitempty"`
	PriceChangeBehavior string `json:"priceChangeBehavior,omitempty"`
}

type OneTimeProductOfferRegion struct {
	RegionCode       string  `json:"regionCode"`
	Availability     string  `json:"availability,omitempty"`
	AbsoluteDiscount *Money  `json:"absoluteDiscount,omitempty"`
	RelativeDiscount float64 `json:"relativeDiscount,omitempty"`
	NoOverride       bool    `json:"noOverride,omitempty"`
}

type OneTimeProductOfferListOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	PageSize         int64                          `json:"pageSize,omitempty"`
	PageToken        string                         `json:"pageToken,omitempty"`
}

func (o OneTimeProductOfferListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferListProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferListPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if o.ProductID.String() == OneTimeProductOfferWildcardID && o.PurchaseOptionID.String() != OneTimeProductOfferWildcardID {
		return fmt.Errorf("one-time product purchase option ID must be %q when product ID is %q", OneTimeProductOfferWildcardID, OneTimeProductOfferWildcardID)
	}
	if o.PageSize < 0 {
		return fmt.Errorf("page size cannot be negative")
	}
	if o.PageSize > 1000 {
		return fmt.Errorf("page size cannot exceed 1000")
	}
	return nil
}

type OneTimeProductOfferListResult struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	Offers           []OneTimeProductOffer          `json:"offers"`
	NextPageToken    string                         `json:"nextPageToken,omitempty"`
	Options          OneTimeProductOfferListOptions `json:"options"`
}

type OneTimeProductOfferLister interface {
	ListOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error)
}

func ListOneTimeProductOffers(ctx context.Context, lister OneTimeProductOfferLister, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOfferListResult{}, err
	}
	if lister == nil {
		return OneTimeProductOfferListResult{}, fmt.Errorf("one-time product offer lister is required")
	}
	return lister.ListOneTimeProductOffers(ctx, options)
}

type OneTimeProductOfferGetOptions struct {
	PackageName      PackageName                    `json:"packageName"`
	ProductID        OneTimeProductID               `json:"productId"`
	PurchaseOptionID OneTimeProductPurchaseOptionID `json:"purchaseOptionId"`
	OfferID          OneTimeProductOfferID          `json:"offerId"`
}

func (o OneTimeProductOfferGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewOneTimeProductID(o.ProductID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductPurchaseOptionID(o.PurchaseOptionID.String()); err != nil {
		return err
	}
	if _, err := NewOneTimeProductOfferID(o.OfferID.String()); err != nil {
		return err
	}
	return nil
}

type OneTimeProductOfferGetter interface {
	GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error)
}

func GetOneTimeProductOffer(ctx context.Context, getter OneTimeProductOfferGetter, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductOffer{}, err
	}
	if getter == nil {
		return OneTimeProductOffer{}, fmt.Errorf("one-time product offer getter is required")
	}
	return getter.GetOneTimeProductOffer(ctx, options)
}
