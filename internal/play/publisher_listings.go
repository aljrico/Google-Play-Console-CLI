package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListListings(ctx context.Context, packageName PackageName, editID string) ([]Listing, error) {
	response, err := p.service.Edits.Listings.List(packageName.String(), editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list listings for %s: %w", packageName, err)
	}
	listings := make([]Listing, 0, len(response.Listings))
	for _, apiListing := range response.Listings {
		listings = append(listings, listingFromAPI(apiListing))
	}
	return listings, nil
}

func (p *GooglePublisher) GetListing(ctx context.Context, packageName PackageName, editID string, language ListingLanguage) (Listing, error) {
	listing, err := p.service.Edits.Listings.Get(packageName.String(), editID, language.String()).Context(ctx).Do()
	if err != nil {
		return Listing{}, fmt.Errorf("get %s listing for %s: %w", language, packageName, err)
	}
	return listingFromAPI(listing), nil
}

// UpsertListing creates or updates a localized store listing. It calls
// Edits.Listings.Update (HTTP PUT), which the Play API documents as
// "Creates or updates a localized store listing". Unlike Patch (HTTP PATCH),
// Update does not require the locale to already exist, so it can add new
// languages as well as replace existing ones.
//
// Update is a full replace (PUT): any field absent from the request body is
// cleared on the server. The caller only supplies the fields it wants to set,
// so to preserve the partial-update ergonomics of the old Patch path we first
// read the existing listing (when the locale already exists) and overlay the
// caller's fields on top of it before sending the complete resource. When the
// locale does not exist yet the read returns 404 and we create it from the
// caller's fields alone.
func (p *GooglePublisher) UpsertListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error) {
	existing, err := p.service.Edits.Listings.Get(packageName.String(), editID, listing.Language.String()).Context(ctx).Do()
	if err != nil {
		if !isGoogleNotFound(err) {
			return Listing{}, fmt.Errorf("get %s listing for %s before update: %w", listing.Language, packageName, err)
		}
		existing = nil
	}

	apiListing := listingToAPI(mergeListing(listingFromAPI(existing), listing))
	updatedListing, err := p.service.Edits.Listings.Update(packageName.String(), editID, listing.Language.String(), apiListing).Context(ctx).Do()
	if err != nil {
		return Listing{}, fmt.Errorf("update %s listing for %s: %w", listing.Language, packageName, err)
	}
	return listingFromAPI(updatedListing), nil
}

// mergeListing overlays the explicitly-set fields of override onto base,
// keeping base's values for any field the caller did not provide. The override
// language wins so the merged listing always targets the requested locale.
func mergeListing(base Listing, override Listing) Listing {
	merged := base
	merged.Language = override.Language
	if override.Title != nil {
		merged.Title = override.Title
	}
	if override.ShortDescription != nil {
		merged.ShortDescription = override.ShortDescription
	}
	if override.FullDescription != nil {
		merged.FullDescription = override.FullDescription
	}
	if override.Video != nil {
		merged.Video = override.Video
	}
	return merged
}

func (p *GooglePublisher) DeleteListing(ctx context.Context, packageName PackageName, editID string, language ListingLanguage) error {
	if err := p.service.Edits.Listings.Delete(packageName.String(), editID, language.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete %s listing for %s: %w", language, packageName, err)
	}
	return nil
}

func (p *GooglePublisher) DeleteAllListings(ctx context.Context, packageName PackageName, editID string) error {
	if err := p.service.Edits.Listings.Deleteall(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete all listings for %s: %w", packageName, err)
	}
	return nil
}

func listingFromAPI(apiListing *androidpublisher.Listing) Listing {
	if apiListing == nil {
		return Listing{}
	}
	return Listing{
		Language:         ListingLanguage(apiListing.Language),
		Title:            stringPointer(apiListing.Title),
		ShortDescription: stringPointer(apiListing.ShortDescription),
		FullDescription:  stringPointer(apiListing.FullDescription),
		Video:            stringPointer(apiListing.Video),
	}
}

func listingToAPI(listing Listing) *androidpublisher.Listing {
	apiListing := &androidpublisher.Listing{
		Language: listing.Language.String(),
	}
	apiListing.ForceSendFields = append(apiListing.ForceSendFields, "Language")
	if listing.Title != nil {
		apiListing.Title = *listing.Title
		apiListing.ForceSendFields = append(apiListing.ForceSendFields, "Title")
	}
	if listing.ShortDescription != nil {
		apiListing.ShortDescription = *listing.ShortDescription
		apiListing.ForceSendFields = append(apiListing.ForceSendFields, "ShortDescription")
	}
	if listing.FullDescription != nil {
		apiListing.FullDescription = *listing.FullDescription
		apiListing.ForceSendFields = append(apiListing.ForceSendFields, "FullDescription")
	}
	if listing.Video != nil {
		apiListing.Video = *listing.Video
		apiListing.ForceSendFields = append(apiListing.ForceSendFields, "Video")
	}
	return apiListing
}
