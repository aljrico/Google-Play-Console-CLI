package play

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func (p *GooglePublisher) ListAppRecoveries(ctx context.Context, options AppRecoveryListOptions) (AppRecoveryListResult, error) {
	response, err := p.service.Apprecovery.List(options.PackageName.String()).
		VersionCode(options.VersionCode).
		Context(ctx).
		Do()
	if err != nil {
		return AppRecoveryListResult{}, fmt.Errorf("list app recoveries for %s version code %d: %w", options.PackageName, options.VersionCode, err)
	}
	return appRecoveryListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) CreateAppRecovery(ctx context.Context, options AppRecoveryCreateOptions) (AppRecoveryAction, error) {
	if err := options.ValidateLive(); err != nil {
		return AppRecoveryAction{}, err
	}
	action, err := p.service.Apprecovery.Create(options.PackageName.String(), appRecoveryCreateToAPI(options)).
		Context(ctx).
		Do()
	if err != nil {
		return AppRecoveryAction{}, fmt.Errorf("create app recovery for %s: %w", options.PackageName, err)
	}
	return appRecoveryActionFromAPI(action), nil
}

func (p *GooglePublisher) AddAppRecoveryTargeting(ctx context.Context, options AppRecoveryTargetingUpdateOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if _, err := p.service.Apprecovery.AddTargeting(
		options.PackageName.String(),
		options.AppRecoveryID.Int64(),
		appRecoveryTargetingUpdateToAPI(options),
	).Context(ctx).Do(); err != nil {
		return fmt.Errorf("add targeting to app recovery %s for %s: %w", options.AppRecoveryID, options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) DeployAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if _, err := p.service.Apprecovery.Deploy(options.PackageName.String(), options.AppRecoveryID.Int64(), &androidpublisher.DeployAppRecoveryRequest{}).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("deploy app recovery %s for %s: %w", options.AppRecoveryID, options.PackageName, err)
	}
	return nil
}

func (p *GooglePublisher) CancelAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if _, err := p.service.Apprecovery.Cancel(options.PackageName.String(), options.AppRecoveryID.Int64(), &androidpublisher.CancelAppRecoveryRequest{}).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("cancel app recovery %s for %s: %w", options.AppRecoveryID, options.PackageName, err)
	}
	return nil
}

func appRecoveryListResultFromAPI(options AppRecoveryListOptions, response *androidpublisher.ListAppRecoveriesResponse) AppRecoveryListResult {
	result := AppRecoveryListResult{
		PackageName: options.PackageName,
		VersionCode: options.VersionCode,
		Actions:     []AppRecoveryAction{},
	}
	if response == nil {
		return result
	}
	for _, apiAction := range response.RecoveryActions {
		if apiAction == nil {
			continue
		}
		result.Actions = append(result.Actions, appRecoveryActionFromAPI(apiAction))
	}
	return result
}

func appRecoveryActionFromAPI(apiAction *androidpublisher.AppRecoveryAction) AppRecoveryAction {
	if apiAction == nil {
		return AppRecoveryAction{}
	}
	action := AppRecoveryAction{
		AppRecoveryID:         strconv.FormatInt(apiAction.AppRecoveryId, 10),
		Status:                apiAction.Status,
		CreateTime:            apiAction.CreateTime,
		DeployTime:            apiAction.DeployTime,
		CancelTime:            apiAction.CancelTime,
		LastUpdateTime:        apiAction.LastUpdateTime,
		Targeting:             appRecoveryTargetingFromAPI(apiAction.Targeting),
		RemoteInAppUpdateData: appRecoveryRemoteInAppUpdateDataFromAPI(apiAction.RemoteInAppUpdateData),
	}
	return action
}

func appRecoveryCreateToAPI(options AppRecoveryCreateOptions) *androidpublisher.CreateDraftAppRecoveryRequest {
	targeting := &androidpublisher.Targeting{}
	if len(options.VersionCodes) > 0 {
		targeting.VersionList = &androidpublisher.AppVersionList{VersionCodes: googleapi.Int64s(append([]int64(nil), options.VersionCodes...))}
	}
	if options.VersionCodeStart > 0 {
		targeting.VersionRange = &androidpublisher.AppVersionRange{
			VersionCodeStart: options.VersionCodeStart,
			VersionCodeEnd:   options.VersionCodeEnd,
		}
	}
	if options.AllUsers {
		targeting.AllUsers = &androidpublisher.AllUsers{IsAllUsersRequested: true}
	}
	if len(options.SDKLevels) > 0 {
		targeting.AndroidSdks = &androidpublisher.AndroidSdks{SdkLevels: googleapi.Int64s(append([]int64(nil), options.SDKLevels...))}
	}
	if len(options.RegionCodes) > 0 {
		targeting.Regions = &androidpublisher.Regions{RegionCode: append([]string(nil), options.RegionCodes...)}
	}
	return &androidpublisher.CreateDraftAppRecoveryRequest{
		RemoteInAppUpdate: &androidpublisher.RemoteInAppUpdate{IsRemoteInAppUpdateRequested: true},
		Targeting:         targeting,
	}
}

func appRecoveryTargetingUpdateToAPI(options AppRecoveryTargetingUpdateOptions) *androidpublisher.AddTargetingRequest {
	update := &androidpublisher.TargetingUpdate{}
	if options.AllUsers {
		update.AllUsers = &androidpublisher.AllUsers{IsAllUsersRequested: true}
	}
	if len(options.SDKLevels) > 0 {
		update.AndroidSdks = &androidpublisher.AndroidSdks{SdkLevels: googleapi.Int64s(append([]int64(nil), options.SDKLevels...))}
	}
	if len(options.RegionCodes) > 0 {
		update.Regions = &androidpublisher.Regions{RegionCode: append([]string(nil), options.RegionCodes...)}
	}
	return &androidpublisher.AddTargetingRequest{TargetingUpdate: update}
}

func appRecoveryTargetingFromAPI(apiTargeting *androidpublisher.Targeting) *AppRecoveryTargeting {
	if apiTargeting == nil {
		return nil
	}
	return &AppRecoveryTargeting{
		AllUsers:     appRecoveryAllUsersFromAPI(apiTargeting.AllUsers),
		AndroidSDKs:  appRecoveryAndroidSDKsFromAPI(apiTargeting.AndroidSdks),
		Regions:      appRecoveryRegionsFromAPI(apiTargeting.Regions),
		VersionList:  appRecoveryVersionListFromAPI(apiTargeting.VersionList),
		VersionRange: appRecoveryVersionRangeFromAPI(apiTargeting.VersionRange),
	}
}

func appRecoveryAllUsersFromAPI(apiAllUsers *androidpublisher.AllUsers) *AppRecoveryAllUsers {
	if apiAllUsers == nil {
		return nil
	}
	return &AppRecoveryAllUsers{IsAllUsersRequested: apiAllUsers.IsAllUsersRequested}
}

func appRecoveryAndroidSDKsFromAPI(apiAndroidSDKs *androidpublisher.AndroidSdks) *AppRecoveryAndroidSDKs {
	if apiAndroidSDKs == nil {
		return nil
	}
	return &AppRecoveryAndroidSDKs{SDKLevels: append([]int64(nil), apiAndroidSDKs.SdkLevels...)}
}

func appRecoveryRegionsFromAPI(apiRegions *androidpublisher.Regions) *AppRecoveryRegions {
	if apiRegions == nil {
		return nil
	}
	return &AppRecoveryRegions{RegionCodes: append([]string(nil), apiRegions.RegionCode...)}
}

func appRecoveryVersionListFromAPI(apiVersionList *androidpublisher.AppVersionList) *AppRecoveryVersionList {
	if apiVersionList == nil {
		return nil
	}
	return &AppRecoveryVersionList{VersionCodes: append([]int64(nil), apiVersionList.VersionCodes...)}
}

func appRecoveryVersionRangeFromAPI(apiVersionRange *androidpublisher.AppVersionRange) *AppRecoveryVersionRange {
	if apiVersionRange == nil {
		return nil
	}
	return &AppRecoveryVersionRange{
		VersionCodeStart: strconv.FormatInt(apiVersionRange.VersionCodeStart, 10),
		VersionCodeEnd:   strconv.FormatInt(apiVersionRange.VersionCodeEnd, 10),
	}
}

func appRecoveryRemoteInAppUpdateDataFromAPI(apiData *androidpublisher.RemoteInAppUpdateData) *AppRecoveryRemoteInAppUpdateData {
	if apiData == nil {
		return nil
	}
	perBundle := make([]AppRecoveryRemoteInAppUpdateDataPerBundle, 0, len(apiData.RemoteAppUpdateDataPerBundle))
	for _, apiBundle := range apiData.RemoteAppUpdateDataPerBundle {
		if apiBundle == nil {
			continue
		}
		perBundle = append(perBundle, AppRecoveryRemoteInAppUpdateDataPerBundle{
			VersionCode:          strconv.FormatInt(apiBundle.VersionCode, 10),
			RecoveredDeviceCount: strconv.FormatInt(apiBundle.RecoveredDeviceCount, 10),
			TotalDeviceCount:     strconv.FormatInt(apiBundle.TotalDeviceCount, 10),
		})
	}
	return &AppRecoveryRemoteInAppUpdateData{PerBundle: perBundle}
}
