package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) GetAppDetails(ctx context.Context, packageName PackageName, editID string) (AppDetails, error) {
	details, err := p.service.Edits.Details.Get(packageName.String(), editID).Context(ctx).Do()
	if err != nil {
		return AppDetails{}, fmt.Errorf("get app details for %s: %w", packageName, err)
	}
	return appDetailsFromAPI(details), nil
}

func (p *GooglePublisher) PatchAppDetails(ctx context.Context, packageName PackageName, editID string, details AppDetails) (AppDetails, error) {
	apiDetails := appDetailsToAPI(details)
	updatedDetails, err := p.service.Edits.Details.Patch(packageName.String(), editID, apiDetails).Context(ctx).Do()
	if err != nil {
		return AppDetails{}, fmt.Errorf("patch app details for %s: %w", packageName, err)
	}
	return appDetailsFromAPI(updatedDetails), nil
}

func appDetailsFromAPI(apiDetails *androidpublisher.AppDetails) AppDetails {
	if apiDetails == nil {
		return AppDetails{}
	}
	return AppDetails{
		DefaultLanguage: stringPointer(apiDetails.DefaultLanguage),
		ContactWebsite:  stringPointer(apiDetails.ContactWebsite),
		ContactEmail:    stringPointer(apiDetails.ContactEmail),
		ContactPhone:    stringPointer(apiDetails.ContactPhone),
	}
}

func appDetailsToAPI(details AppDetails) *androidpublisher.AppDetails {
	apiDetails := &androidpublisher.AppDetails{}
	if details.DefaultLanguage != nil {
		apiDetails.DefaultLanguage = *details.DefaultLanguage
		apiDetails.ForceSendFields = append(apiDetails.ForceSendFields, "DefaultLanguage")
	}
	if details.ContactWebsite != nil {
		apiDetails.ContactWebsite = *details.ContactWebsite
		apiDetails.ForceSendFields = append(apiDetails.ForceSendFields, "ContactWebsite")
	}
	if details.ContactEmail != nil {
		apiDetails.ContactEmail = *details.ContactEmail
		apiDetails.ForceSendFields = append(apiDetails.ForceSendFields, "ContactEmail")
	}
	if details.ContactPhone != nil {
		apiDetails.ContactPhone = *details.ContactPhone
		apiDetails.ForceSendFields = append(apiDetails.ForceSendFields, "ContactPhone")
	}
	return apiDetails
}
