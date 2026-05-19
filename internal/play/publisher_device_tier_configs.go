package play

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListDeviceTierConfigs(ctx context.Context, options DeviceTierConfigListOptions) (DeviceTierConfigListResult, error) {
	call := p.service.Applications.DeviceTierConfigs.List(options.PackageName.String()).Context(ctx)
	if options.PageSize != 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	response, err := call.Do()
	if err != nil {
		return DeviceTierConfigListResult{}, fmt.Errorf("list device tier configs for %s: %w", options.PackageName, err)
	}
	return deviceTierConfigListResultFromAPI(options.PackageName, response), nil
}

func (p *GooglePublisher) GetDeviceTierConfig(ctx context.Context, options DeviceTierConfigGetOptions) (DeviceTierConfigGetResult, error) {
	apiConfig, err := p.service.Applications.DeviceTierConfigs.Get(options.PackageName.String(), options.DeviceTierConfigID).Context(ctx).Do()
	if err != nil {
		return DeviceTierConfigGetResult{}, fmt.Errorf("get device tier config %d for %s: %w", options.DeviceTierConfigID, options.PackageName, err)
	}
	return DeviceTierConfigGetResult{
		PackageName: options.PackageName,
		Config:      deviceTierConfigFromAPI(apiConfig),
	}, nil
}

func deviceTierConfigListResultFromAPI(packageName PackageName, response *androidpublisher.ListDeviceTierConfigsResponse) DeviceTierConfigListResult {
	result := DeviceTierConfigListResult{
		PackageName: packageName,
		Configs:     []DeviceTierConfig{},
	}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiConfig := range response.DeviceTierConfigs {
		if apiConfig == nil {
			continue
		}
		result.Configs = append(result.Configs, deviceTierConfigFromAPI(apiConfig))
	}
	return result
}

func deviceTierConfigFromAPI(apiConfig *androidpublisher.DeviceTierConfig) DeviceTierConfig {
	if apiConfig == nil {
		return DeviceTierConfig{}
	}
	return DeviceTierConfig{
		ID:              strconv.FormatInt(apiConfig.DeviceTierConfigId, 10),
		DeviceGroups:    deviceGroupsFromAPI(apiConfig.DeviceGroups),
		DeviceTierSet:   deviceTierSetFromAPI(apiConfig.DeviceTierSet),
		UserCountrySets: deviceUserCountrySetsFromAPI(apiConfig.UserCountrySets),
	}
}

func deviceGroupsFromAPI(apiGroups []*androidpublisher.DeviceGroup) []DeviceGroup {
	groups := make([]DeviceGroup, 0, len(apiGroups))
	for _, apiGroup := range apiGroups {
		if apiGroup == nil {
			continue
		}
		groups = append(groups, DeviceGroup{
			Name:            apiGroup.Name,
			DeviceSelectors: deviceSelectorsFromAPI(apiGroup.DeviceSelectors),
		})
	}
	return groups
}

func deviceSelectorsFromAPI(apiSelectors []*androidpublisher.DeviceSelector) []json.RawMessage {
	selectors := make([]json.RawMessage, 0, len(apiSelectors))
	for _, apiSelector := range apiSelectors {
		if apiSelector == nil {
			continue
		}
		payload, err := json.Marshal(apiSelector)
		if err != nil || string(payload) == "{}" {
			continue
		}
		selectors = append(selectors, payload)
	}
	return selectors
}

func deviceTierSetFromAPI(apiSet *androidpublisher.DeviceTierSet) *DeviceTierSet {
	if apiSet == nil {
		return nil
	}
	tiers := make([]DeviceTier, 0, len(apiSet.DeviceTiers))
	for _, apiTier := range apiSet.DeviceTiers {
		if apiTier == nil {
			continue
		}
		tiers = append(tiers, DeviceTier{
			Level:            apiTier.Level,
			DeviceGroupNames: append([]string(nil), apiTier.DeviceGroupNames...),
		})
	}
	return &DeviceTierSet{DeviceTiers: tiers}
}
