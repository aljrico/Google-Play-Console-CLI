package play

import (
	"context"
	"encoding/json"
	"fmt"
)

type DeviceTierConfig struct {
	ID              string                 `json:"deviceTierConfigId,omitempty"`
	DeviceGroups    []DeviceGroup          `json:"deviceGroups"`
	DeviceTierSet   *DeviceTierSet         `json:"deviceTierSet,omitempty"`
	UserCountrySets []DeviceUserCountrySet `json:"userCountrySets"`
}

type DeviceGroup struct {
	Name            string            `json:"name,omitempty"`
	DeviceSelectors []json.RawMessage `json:"deviceSelectors"`
}

type DeviceTierSet struct {
	DeviceTiers []DeviceTier `json:"deviceTiers"`
}

type DeviceTier struct {
	Level            int64    `json:"level"`
	DeviceGroupNames []string `json:"deviceGroupNames,omitempty"`
}

type DeviceUserCountrySet struct {
	Name         string   `json:"name,omitempty"`
	CountryCodes []string `json:"countryCodes,omitempty"`
}

type DeviceTierConfigListOptions struct {
	PackageName PackageName `json:"packageName"`
	PageSize    int64       `json:"pageSize,omitempty"`
	PageToken   string      `json:"pageToken,omitempty"`
}

func (o DeviceTierConfigListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.PageSize < 0 || o.PageSize > 100 {
		return fmt.Errorf("page size must be between 0 and 100")
	}
	return nil
}

type DeviceTierConfigGetOptions struct {
	PackageName        PackageName `json:"packageName"`
	DeviceTierConfigID int64       `json:"deviceTierConfigId"`
}

func (o DeviceTierConfigGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.DeviceTierConfigID <= 0 {
		return fmt.Errorf("device tier config ID must be greater than 0")
	}
	return nil
}

type DeviceTierConfigListResult struct {
	PackageName   PackageName        `json:"packageName"`
	NextPageToken string             `json:"nextPageToken,omitempty"`
	Configs       []DeviceTierConfig `json:"deviceTierConfigs"`
}

type DeviceTierConfigGetResult struct {
	PackageName PackageName      `json:"packageName"`
	Config      DeviceTierConfig `json:"deviceTierConfig"`
}

type DeviceTierConfigLister interface {
	ListDeviceTierConfigs(ctx context.Context, options DeviceTierConfigListOptions) (DeviceTierConfigListResult, error)
}

type DeviceTierConfigGetter interface {
	GetDeviceTierConfig(ctx context.Context, options DeviceTierConfigGetOptions) (DeviceTierConfigGetResult, error)
}

func ListDeviceTierConfigs(ctx context.Context, lister DeviceTierConfigLister, options DeviceTierConfigListOptions) (DeviceTierConfigListResult, error) {
	if err := options.Validate(); err != nil {
		return DeviceTierConfigListResult{}, err
	}
	if lister == nil {
		return DeviceTierConfigListResult{}, fmt.Errorf("device tier config lister is required")
	}
	return lister.ListDeviceTierConfigs(ctx, options)
}

func GetDeviceTierConfig(ctx context.Context, getter DeviceTierConfigGetter, options DeviceTierConfigGetOptions) (DeviceTierConfigGetResult, error) {
	if err := options.Validate(); err != nil {
		return DeviceTierConfigGetResult{}, err
	}
	if getter == nil {
		return DeviceTierConfigGetResult{}, fmt.Errorf("device tier config getter is required")
	}
	return getter.GetDeviceTierConfig(ctx, options)
}
