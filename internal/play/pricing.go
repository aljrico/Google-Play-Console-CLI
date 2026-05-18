package play

import (
	"context"
	"fmt"
	"strings"
)

type CurrencyCode string

func NewCurrencyCode(value string) (CurrencyCode, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 3 {
		return "", fmt.Errorf("currency code must be a 3-letter ISO 4217 code")
	}
	for i := 0; i < len(value); i++ {
		if !isASCIIAlpha(value[i]) {
			return "", fmt.Errorf("currency code must contain only ASCII letters")
		}
	}
	return CurrencyCode(value), nil
}

func (c CurrencyCode) String() string {
	return string(c)
}

func (c CurrencyCode) Validate() error {
	_, err := NewCurrencyCode(c.String())
	return err
}

type RegionPriceConversionOptions struct {
	PackageName PackageName  `json:"packageName"`
	Currency    CurrencyCode `json:"currency"`
	Units       int64        `json:"units"`
	Nanos       int64        `json:"nanos,omitempty"`
}

func (o RegionPriceConversionOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if err := o.Currency.Validate(); err != nil {
		return err
	}
	if o.Units < 0 {
		return fmt.Errorf("price units must be 0 or greater")
	}
	if o.Nanos < 0 || o.Nanos > 999999999 {
		return fmt.Errorf("price nanos must be between 0 and 999999999")
	}
	if o.Units == 0 && o.Nanos == 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	return nil
}

type ConvertedOtherRegionsPrice struct {
	USDPrice *Money `json:"usdPrice,omitempty"`
	EURPrice *Money `json:"eurPrice,omitempty"`
}

type ConvertedRegionPrice struct {
	RegionCode string `json:"regionCode,omitempty"`
	Price      *Money `json:"price,omitempty"`
	TaxAmount  *Money `json:"taxAmount,omitempty"`
}

type RegionPriceConversionResult struct {
	PackageName                PackageName                     `json:"packageName"`
	SourcePrice                Money                           `json:"sourcePrice"`
	RegionVersion              string                          `json:"regionVersion,omitempty"`
	ConvertedOtherRegionsPrice *ConvertedOtherRegionsPrice     `json:"convertedOtherRegionsPrice,omitempty"`
	ConvertedRegionPrices      map[string]ConvertedRegionPrice `json:"convertedRegionPrices"`
}

type RegionPriceConverter interface {
	ConvertRegionPrices(ctx context.Context, options RegionPriceConversionOptions) (RegionPriceConversionResult, error)
}

func ConvertRegionPrices(ctx context.Context, converter RegionPriceConverter, options RegionPriceConversionOptions) (RegionPriceConversionResult, error) {
	if err := options.Validate(); err != nil {
		return RegionPriceConversionResult{}, err
	}
	if converter == nil {
		return RegionPriceConversionResult{}, fmt.Errorf("region price converter is required")
	}
	return converter.ConvertRegionPrices(ctx, options)
}
