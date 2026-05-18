package play

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/aljrico/Google-Play-Console-CLI/internal/config"
)

func NewPublisherFromActiveProfile(ctx context.Context) (*GooglePublisher, error) {
	store, err := config.Load()
	if err != nil {
		return nil, err
	}
	profile, ok := store.Profiles[store.ActiveProfile]
	if !ok || store.ActiveProfile == "" {
		return nil, fmt.Errorf("no active auth profile; run gpc auth login")
	}

	service, err := androidpublisher.NewService(
		ctx,
		option.WithCredentialsFile(profile.ServiceAccountFile),
		option.WithScopes(androidpublisher.AndroidpublisherScope),
	)
	if err != nil {
		return nil, fmt.Errorf("create Google Play API service: %w", err)
	}
	return &GooglePublisher{service: service}, nil
}

type GooglePublisher struct {
	service *androidpublisher.Service
}

func (p GooglePublisher) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	edit, err := p.service.Edits.Insert(packageName.String(), &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return Edit{}, fmt.Errorf("insert edit for %s: %w", packageName, err)
	}
	return Edit{ID: edit.Id, ExpiryTimeSeconds: edit.ExpiryTimeSeconds}, nil
}

func (p GooglePublisher) UploadBundle(ctx context.Context, packageName PackageName, editID string, bundlePath string) (BundleArtifact, error) {
	file, err := os.Open(bundlePath)
	if err != nil {
		return BundleArtifact{}, fmt.Errorf("open bundle %s: %w", bundlePath, err)
	}
	defer file.Close()

	bundle, err := p.service.Edits.Bundles.Upload(packageName.String(), editID).
		Media(file, googleapi.ContentType("application/octet-stream")).
		Context(ctx).
		Do()
	if err != nil {
		return BundleArtifact{}, fmt.Errorf("upload bundle %s for %s: %w", bundlePath, packageName, err)
	}
	return BundleArtifact{VersionCode: bundle.VersionCode, SHA1: bundle.Sha1, SHA256: bundle.Sha256}, nil
}

func (p GooglePublisher) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	if _, err := p.service.Edits.Validate(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("validate edit %s for %s: %w", editID, packageName, err)
	}
	return nil
}

func (p GooglePublisher) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	edit, err := p.service.Edits.Commit(packageName.String(), editID).Context(ctx).Do()
	if err != nil {
		return Edit{}, fmt.Errorf("commit edit %s for %s: %w", editID, packageName, err)
	}
	return Edit{ID: edit.Id, ExpiryTimeSeconds: edit.ExpiryTimeSeconds}, nil
}

func (p GooglePublisher) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	if err := p.service.Edits.Delete(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete edit %s for %s: %w", editID, packageName, err)
	}
	return nil
}

func (p GooglePublisher) ListTracks(ctx context.Context, packageName PackageName, editID string) ([]Track, error) {
	response, err := p.service.Edits.Tracks.List(packageName.String(), editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list tracks for %s: %w", packageName, err)
	}
	tracks := make([]Track, 0, len(response.Tracks))
	for _, apiTrack := range response.Tracks {
		tracks = append(tracks, trackFromAPI(apiTrack))
	}
	return tracks, nil
}

func (p GooglePublisher) ListListings(ctx context.Context, packageName PackageName, editID string) ([]Listing, error) {
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

func (p GooglePublisher) GetListing(ctx context.Context, packageName PackageName, editID string, language ListingLanguage) (Listing, error) {
	listing, err := p.service.Edits.Listings.Get(packageName.String(), editID, language.String()).Context(ctx).Do()
	if err != nil {
		return Listing{}, fmt.Errorf("get %s listing for %s: %w", language, packageName, err)
	}
	return listingFromAPI(listing), nil
}

func (p GooglePublisher) PatchListing(ctx context.Context, packageName PackageName, editID string, listing Listing) (Listing, error) {
	apiListing := listingToAPI(listing)
	updatedListing, err := p.service.Edits.Listings.Patch(packageName.String(), editID, listing.Language.String(), apiListing).Context(ctx).Do()
	if err != nil {
		return Listing{}, fmt.Errorf("patch %s listing for %s: %w", listing.Language, packageName, err)
	}
	return listingFromAPI(updatedListing), nil
}

func (p GooglePublisher) GetAppDetails(ctx context.Context, packageName PackageName, editID string) (AppDetails, error) {
	details, err := p.service.Edits.Details.Get(packageName.String(), editID).Context(ctx).Do()
	if err != nil {
		return AppDetails{}, fmt.Errorf("get app details for %s: %w", packageName, err)
	}
	return appDetailsFromAPI(details), nil
}

func (p GooglePublisher) PatchAppDetails(ctx context.Context, packageName PackageName, editID string, details AppDetails) (AppDetails, error) {
	apiDetails := appDetailsToAPI(details)
	updatedDetails, err := p.service.Edits.Details.Patch(packageName.String(), editID, apiDetails).Context(ctx).Do()
	if err != nil {
		return AppDetails{}, fmt.Errorf("patch app details for %s: %w", packageName, err)
	}
	return appDetailsFromAPI(updatedDetails), nil
}

func (p GooglePublisher) DeleteListing(ctx context.Context, packageName PackageName, editID string, language ListingLanguage) error {
	if err := p.service.Edits.Listings.Delete(packageName.String(), editID, language.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete %s listing for %s: %w", language, packageName, err)
	}
	return nil
}

func (p GooglePublisher) DeleteAllListings(ctx context.Context, packageName PackageName, editID string) error {
	if err := p.service.Edits.Listings.Deleteall(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete all listings for %s: %w", packageName, err)
	}
	return nil
}

func (p GooglePublisher) ListReviews(ctx context.Context, options ReviewListOptions) (ReviewListResult, error) {
	call := p.service.Reviews.List(options.PackageName.String()).Context(ctx)
	if options.MaxResults > 0 {
		call.MaxResults(options.MaxResults)
	}
	if options.StartIndex > 0 {
		call.StartIndex(options.StartIndex)
	}
	if options.Token != "" {
		call.Token(options.Token)
	}
	if options.TranslationLanguage != "" {
		call.TranslationLanguage(options.TranslationLanguage)
	}
	response, err := call.Do()
	if err != nil {
		return ReviewListResult{}, fmt.Errorf("list reviews for %s: %w", options.PackageName, err)
	}
	return reviewListResultFromAPI(options, response), nil
}

func (p GooglePublisher) GetReview(ctx context.Context, packageName PackageName, reviewID ReviewID, translationLanguage string) (Review, error) {
	call := p.service.Reviews.Get(packageName.String(), reviewID.String()).Context(ctx)
	if translationLanguage != "" {
		call.TranslationLanguage(translationLanguage)
	}
	review, err := call.Do()
	if err != nil {
		return Review{}, fmt.Errorf("get review %s for %s: %w", reviewID, packageName, err)
	}
	return reviewFromAPI(review), nil
}

func (p GooglePublisher) ReplyToReview(ctx context.Context, packageName PackageName, reviewID ReviewID, text string) (DeveloperReply, error) {
	request := &androidpublisher.ReviewsReplyRequest{
		ReplyText:       text,
		ForceSendFields: []string{"ReplyText"},
	}
	response, err := p.service.Reviews.Reply(packageName.String(), reviewID.String(), request).Context(ctx).Do()
	if err != nil {
		return DeveloperReply{}, fmt.Errorf("reply to review %s for %s: %w", reviewID, packageName, err)
	}
	if response == nil || response.Result == nil {
		return DeveloperReply{}, nil
	}
	return DeveloperReply{
		Text:       response.Result.ReplyText,
		LastEdited: timestampFromAPI(response.Result.LastEdited),
	}, nil
}

func (p GooglePublisher) ListInAppProducts(ctx context.Context, options InAppProductListOptions) (InAppProductListResult, error) {
	call := p.service.Inappproducts.List(options.PackageName.String()).Context(ctx)
	if options.Token != "" {
		call.Token(options.Token)
	}
	response, err := call.Do()
	if err != nil {
		return InAppProductListResult{}, fmt.Errorf("list in-app products for %s: %w", options.PackageName, err)
	}
	return inAppProductListResultFromAPI(options, response), nil
}

func (p GooglePublisher) GetInAppProduct(ctx context.Context, packageName PackageName, sku InAppProductSKU) (InAppProduct, error) {
	product, err := p.service.Inappproducts.Get(packageName.String(), sku.String()).Context(ctx).Do()
	if err != nil {
		return InAppProduct{}, fmt.Errorf("get in-app product %s for %s: %w", sku, packageName, err)
	}
	return inAppProductFromAPI(product), nil
}

func (p GooglePublisher) ListSubscriptions(ctx context.Context, options SubscriptionListOptions) (SubscriptionListResult, error) {
	call := p.service.Monetization.Subscriptions.List(options.PackageName.String()).Context(ctx)
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	if options.ShowArchived {
		call.ShowArchived(options.ShowArchived)
	}
	response, err := call.Do()
	if err != nil {
		return SubscriptionListResult{}, fmt.Errorf("list subscriptions for %s: %w", options.PackageName, err)
	}
	return subscriptionListResultFromAPI(options, response), nil
}

func (p GooglePublisher) GetSubscription(ctx context.Context, packageName PackageName, productID SubscriptionProductID) (Subscription, error) {
	subscription, err := p.service.Monetization.Subscriptions.Get(packageName.String(), productID.String()).Context(ctx).Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("get subscription %s for %s: %w", productID, packageName, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p GooglePublisher) ListSubscriptionOffers(ctx context.Context, options SubscriptionOfferListOptions) (SubscriptionOfferListResult, error) {
	call := p.service.Monetization.Subscriptions.BasePlans.Offers.List(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
	).Context(ctx)
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	response, err := call.Do()
	if err != nil {
		return SubscriptionOfferListResult{}, fmt.Errorf("list subscription offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferListResultFromAPI(options, response), nil
}

func (p GooglePublisher) GetSubscriptionOffer(ctx context.Context, packageName PackageName, productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, offerID SubscriptionOfferID) (SubscriptionOffer, error) {
	offer, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
		packageName.String(),
		productID.String(),
		basePlanID.String(),
		offerID.String(),
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOffer{}, fmt.Errorf("get subscription offer %s for %s/%s/%s: %w", offerID, packageName, productID, basePlanID, err)
	}
	return subscriptionOfferFromAPI(offer), nil
}

func (p GooglePublisher) GetProductPurchase(ctx context.Context, options ProductPurchaseOptions) (ProductPurchase, error) {
	purchase, err := p.service.Purchases.Productsv2.Getproductpurchasev2(options.PackageName.String(), options.Token.String()).Context(ctx).Do()
	if err != nil {
		return ProductPurchase{}, fmt.Errorf("get product purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return productPurchaseFromAPI(options, purchase), nil
}

func (p GooglePublisher) GetSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseOptions) (SubscriptionPurchase, error) {
	purchase, err := p.service.Purchases.Subscriptionsv2.Get(options.PackageName.String(), options.Token.String()).Context(ctx).Do()
	if err != nil {
		return SubscriptionPurchase{}, fmt.Errorf("get subscription purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return subscriptionPurchaseFromAPI(options.PackageName, options.Token, purchase), nil
}

func (p GooglePublisher) AppendTrackRelease(ctx context.Context, packageName PackageName, editID string, trackName TrackName, release TrackRelease) (Track, error) {
	apiTrack, err := p.service.Edits.Tracks.Get(packageName.String(), editID, trackName.String()).Context(ctx).Do()
	if err != nil {
		return Track{}, fmt.Errorf("get %s track for %s: %w", trackName, packageName, err)
	}
	if apiTrack == nil {
		apiTrack = &androidpublisher.Track{Track: trackName.String()}
	}
	apiTrack.Track = trackName.String()
	upsertTrackRelease(apiTrack, releaseToAPI(release))

	updatedTrack, err := p.service.Edits.Tracks.Update(packageName.String(), editID, trackName.String(), apiTrack).Context(ctx).Do()
	if err != nil {
		return Track{}, fmt.Errorf("append release to %s track for %s: %w", trackName, packageName, err)
	}
	return trackFromAPI(updatedTrack), nil
}

func (p GooglePublisher) PromoteTrackRelease(ctx context.Context, packageName PackageName, editID string, sourceTrack TrackName, targetTrack TrackName, versionCode int64, status ReleaseStatus, userFraction *float64) (TrackRelease, error) {
	source, err := p.service.Edits.Tracks.Get(packageName.String(), editID, sourceTrack.String()).Context(ctx).Do()
	if err != nil {
		return TrackRelease{}, fmt.Errorf("get %s track for %s: %w", sourceTrack, packageName, err)
	}
	apiRelease, err := selectReleaseByVersionCode(source, versionCode)
	if err != nil {
		return TrackRelease{}, fmt.Errorf("find version code %d on %s track for %s: %w", versionCode, sourceTrack, packageName, err)
	}
	setReleaseStatus(apiRelease, status, userFraction)

	target, err := p.service.Edits.Tracks.Get(packageName.String(), editID, targetTrack.String()).Context(ctx).Do()
	if err != nil {
		return TrackRelease{}, fmt.Errorf("get %s track for %s: %w", targetTrack, packageName, err)
	}
	if target == nil {
		target = &androidpublisher.Track{Track: targetTrack.String()}
	}
	target.Track = targetTrack.String()
	upsertTrackRelease(target, apiRelease)

	if _, err := p.service.Edits.Tracks.Update(packageName.String(), editID, targetTrack.String(), target).Context(ctx).Do(); err != nil {
		return TrackRelease{}, fmt.Errorf("promote release from %s to %s for %s: %w", sourceTrack, targetTrack, packageName, err)
	}
	return releaseFromAPI(apiRelease), nil
}

func (p GooglePublisher) UpdateTrackReleaseStatus(ctx context.Context, packageName PackageName, editID string, trackName TrackName, versionCode int64, status ReleaseStatus, userFraction *float64) (TrackRelease, error) {
	apiTrack, err := p.service.Edits.Tracks.Get(packageName.String(), editID, trackName.String()).Context(ctx).Do()
	if err != nil {
		return TrackRelease{}, fmt.Errorf("get %s track for %s: %w", trackName, packageName, err)
	}
	apiRelease, err := selectReleaseByVersionCode(apiTrack, versionCode)
	if err != nil {
		return TrackRelease{}, fmt.Errorf("find version code %d on %s track for %s: %w", versionCode, trackName, packageName, err)
	}
	setReleaseStatus(apiRelease, status, userFraction)

	if _, err := p.service.Edits.Tracks.Update(packageName.String(), editID, trackName.String(), apiTrack).Context(ctx).Do(); err != nil {
		return TrackRelease{}, fmt.Errorf("update rollout status for version code %d on %s track for %s: %w", versionCode, trackName, packageName, err)
	}
	return releaseFromAPI(apiRelease), nil
}

func trackFromAPI(apiTrack *androidpublisher.Track) Track {
	if apiTrack == nil {
		return Track{}
	}
	releases := make([]TrackRelease, 0, len(apiTrack.Releases))
	for _, apiRelease := range apiTrack.Releases {
		releases = append(releases, releaseFromAPI(apiRelease))
	}
	return Track{Name: TrackName(apiTrack.Track), Releases: releases}
}

func releaseToAPI(release TrackRelease) *androidpublisher.TrackRelease {
	apiRelease := &androidpublisher.TrackRelease{
		Name:            release.Name,
		Status:          release.Status.String(),
		VersionCodes:    googleapi.Int64s(release.VersionCodes),
		ForceSendFields: []string{"Status"},
	}
	if release.UserFraction != nil {
		apiRelease.UserFraction = *release.UserFraction
		apiRelease.ForceSendFields = append(apiRelease.ForceSendFields, "UserFraction")
	}
	return apiRelease
}

func selectReleaseByVersionCode(track *androidpublisher.Track, versionCode int64) (*androidpublisher.TrackRelease, error) {
	if track == nil || len(track.Releases) == 0 {
		return nil, fmt.Errorf("track has no releases")
	}
	for _, release := range track.Releases {
		if hasVersionCode(release, versionCode) {
			return release, nil
		}
	}
	return nil, fmt.Errorf("version code not found")
}

func hasVersionCode(release *androidpublisher.TrackRelease, versionCode int64) bool {
	if release == nil {
		return false
	}
	for _, candidate := range release.VersionCodes {
		if candidate == versionCode {
			return true
		}
	}
	return false
}

func upsertTrackRelease(track *androidpublisher.Track, release *androidpublisher.TrackRelease) {
	for index, existingRelease := range track.Releases {
		if releasesShareVersionCode(existingRelease, release) {
			track.Releases[index] = release
			return
		}
	}
	track.Releases = append(track.Releases, release)
}

func releasesShareVersionCode(left *androidpublisher.TrackRelease, right *androidpublisher.TrackRelease) bool {
	if left == nil || right == nil {
		return false
	}
	for _, versionCode := range right.VersionCodes {
		if hasVersionCode(left, versionCode) {
			return true
		}
	}
	return false
}

func setReleaseStatus(release *androidpublisher.TrackRelease, status ReleaseStatus, userFraction *float64) {
	release.Status = status.String()
	release.UserFraction = 0
	release.ForceSendFields = appendUniqueField(release.ForceSendFields, "Status")
	release.NullFields = removeField(release.NullFields, "UserFraction")
	if userFraction == nil {
		release.ForceSendFields = removeField(release.ForceSendFields, "UserFraction")
		return
	}
	release.UserFraction = *userFraction
	release.ForceSendFields = appendUniqueField(release.ForceSendFields, "UserFraction")
}

func appendUniqueField(fields []string, field string) []string {
	for _, candidate := range fields {
		if candidate == field {
			return fields
		}
	}
	return append(fields, field)
}

func removeField(fields []string, field string) []string {
	result := fields[:0]
	for _, candidate := range fields {
		if candidate != field {
			result = append(result, candidate)
		}
	}
	return result
}

func releaseFromAPI(apiRelease *androidpublisher.TrackRelease) TrackRelease {
	if apiRelease == nil {
		return TrackRelease{}
	}
	return TrackRelease{
		Name:         apiRelease.Name,
		Status:       ReleaseStatus(apiRelease.Status),
		UserFraction: userFractionFromAPI(apiRelease),
		VersionCodes: []int64(apiRelease.VersionCodes),
	}
}

func userFractionFromAPI(apiRelease *androidpublisher.TrackRelease) *float64 {
	if apiRelease.UserFraction == 0 {
		return nil
	}
	userFraction := apiRelease.UserFraction
	return &userFraction
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

func reviewListResultFromAPI(options ReviewListOptions, response *androidpublisher.ReviewsListResponse) ReviewListResult {
	result := ReviewListResult{
		PackageName: options.PackageName,
		Reviews:     []Review{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	for _, apiReview := range response.Reviews {
		result.Reviews = append(result.Reviews, reviewFromAPI(apiReview))
	}
	if response.PageInfo != nil {
		result.PageInfo = &ReviewPageInfo{
			ResultPerPage: response.PageInfo.ResultPerPage,
			StartIndex:    response.PageInfo.StartIndex,
			TotalResults:  response.PageInfo.TotalResults,
		}
	}
	if response.TokenPagination != nil {
		result.Pagination = &ReviewPagination{
			NextPageToken:     response.TokenPagination.NextPageToken,
			PreviousPageToken: response.TokenPagination.PreviousPageToken,
		}
	}
	return result
}

func reviewFromAPI(apiReview *androidpublisher.Review) Review {
	if apiReview == nil {
		return Review{Comments: []ReviewComment{}}
	}
	comments := make([]ReviewComment, 0, len(apiReview.Comments))
	for _, apiComment := range apiReview.Comments {
		comments = append(comments, reviewCommentFromAPI(apiComment))
	}
	return Review{
		ReviewID:   ReviewID(apiReview.ReviewId),
		AuthorName: apiReview.AuthorName,
		Comments:   comments,
	}
}

func reviewCommentFromAPI(apiComment *androidpublisher.Comment) ReviewComment {
	if apiComment == nil {
		return ReviewComment{}
	}
	if apiComment.UserComment != nil {
		userComment := apiComment.UserComment
		return ReviewComment{
			Kind:      ReviewCommentKindUser,
			Text:      userComment.Text,
			UpdatedAt: timestampFromAPI(userComment.LastModified),
			User: &UserReviewComment{
				ReviewerLanguage: userComment.ReviewerLanguage,
				StarRating:       userComment.StarRating,
				AppVersionCode:   userComment.AppVersionCode,
				AppVersionName:   userComment.AppVersionName,
				AndroidOSVersion: userComment.AndroidOsVersion,
				Device:           userComment.Device,
				OriginalText:     userComment.OriginalText,
				ThumbsUpCount:    userComment.ThumbsUpCount,
				ThumbsDownCount:  userComment.ThumbsDownCount,
				DeviceMetadata:   deviceMetadataFromAPI(userComment.DeviceMetadata),
			},
		}
	}
	if apiComment.DeveloperComment != nil {
		developerComment := apiComment.DeveloperComment
		return ReviewComment{
			Kind:      ReviewCommentKindDeveloper,
			Text:      developerComment.Text,
			UpdatedAt: timestampFromAPI(developerComment.LastModified),
			Developer: &DeveloperReviewComment{
				LastEdited: timestampFromAPI(developerComment.LastModified),
			},
		}
	}
	return ReviewComment{}
}

func deviceMetadataFromAPI(apiMetadata *androidpublisher.DeviceMetadata) *DeviceMetadata {
	if apiMetadata == nil {
		return nil
	}
	return &DeviceMetadata{
		CPUMake:            apiMetadata.CpuMake,
		CPUModel:           apiMetadata.CpuModel,
		DeviceClass:        apiMetadata.DeviceClass,
		GLESVersion:        apiMetadata.GlEsVersion,
		Manufacturer:       apiMetadata.Manufacturer,
		NativePlatform:     apiMetadata.NativePlatform,
		ProductName:        apiMetadata.ProductName,
		RAMMegabytes:       apiMetadata.RamMb,
		ScreenDensityDPI:   apiMetadata.ScreenDensityDpi,
		ScreenHeightPixels: apiMetadata.ScreenHeightPx,
		ScreenWidthPixels:  apiMetadata.ScreenWidthPx,
	}
}

func timestampFromAPI(apiTimestamp *androidpublisher.Timestamp) *Timestamp {
	if apiTimestamp == nil {
		return nil
	}
	return &Timestamp{Seconds: apiTimestamp.Seconds, Nanos: apiTimestamp.Nanos}
}

func inAppProductListResultFromAPI(options InAppProductListOptions, response *androidpublisher.InappproductsListResponse) InAppProductListResult {
	result := InAppProductListResult{
		PackageName: options.PackageName,
		Products:    []InAppProduct{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	for _, apiProduct := range response.Inappproduct {
		result.Products = append(result.Products, inAppProductFromAPI(apiProduct))
	}
	if response.TokenPagination != nil {
		result.Pagination = &InAppProductPagination{
			NextPageToken:     response.TokenPagination.NextPageToken,
			PreviousPageToken: response.TokenPagination.PreviousPageToken,
		}
	}
	return result
}

func inAppProductFromAPI(apiProduct *androidpublisher.InAppProduct) InAppProduct {
	if apiProduct == nil {
		return InAppProduct{}
	}
	return InAppProduct{
		PackageName:                            PackageName(apiProduct.PackageName),
		SKU:                                    InAppProductSKU(apiProduct.Sku),
		Status:                                 ProductStatus(apiProduct.Status),
		PurchaseType:                           ProductPurchaseType(apiProduct.PurchaseType),
		DefaultLanguage:                        apiProduct.DefaultLanguage,
		DefaultPrice:                           productPriceFromAPI(apiProduct.DefaultPrice),
		Prices:                                 productPricesFromAPI(apiProduct.Prices),
		Listings:                               inAppProductListingsFromAPI(apiProduct.Listings),
		SubscriptionPeriod:                     apiProduct.SubscriptionPeriod,
		TrialPeriod:                            apiProduct.TrialPeriod,
		GracePeriod:                            apiProduct.GracePeriod,
		ManagedProductTaxAndComplianceSettings: managedProductTaxComplianceSettingsFromAPI(apiProduct.ManagedProductTaxesAndComplianceSettings),
		SubscriptionTaxAndComplianceSettings:   subscriptionTaxComplianceSettingsFromAPI(apiProduct.SubscriptionTaxesAndComplianceSettings),
	}
}

func productPriceFromAPI(apiPrice *androidpublisher.Price) *ProductPrice {
	if apiPrice == nil {
		return nil
	}
	return &ProductPrice{Currency: apiPrice.Currency, PriceMicros: apiPrice.PriceMicros}
}

func productPricesFromAPI(apiPrices map[string]androidpublisher.Price) map[string]ProductPrice {
	if len(apiPrices) == 0 {
		return nil
	}
	prices := make(map[string]ProductPrice, len(apiPrices))
	for region, apiPrice := range apiPrices {
		prices[region] = ProductPrice{Currency: apiPrice.Currency, PriceMicros: apiPrice.PriceMicros}
	}
	return prices
}

func inAppProductListingsFromAPI(apiListings map[string]androidpublisher.InAppProductListing) map[string]InAppProductListing {
	if len(apiListings) == 0 {
		return nil
	}
	listings := make(map[string]InAppProductListing, len(apiListings))
	for language, apiListing := range apiListings {
		listings[language] = InAppProductListing{
			Title:       apiListing.Title,
			Description: apiListing.Description,
			Benefits:    apiListing.Benefits,
		}
	}
	return listings
}

func managedProductTaxComplianceSettingsFromAPI(apiSettings *androidpublisher.ManagedProductTaxAndComplianceSettings) *ProductTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	return &ProductTaxComplianceSettings{
		EEAWithdrawalRightType:  apiSettings.EeaWithdrawalRightType,
		IsTokenizedDigitalAsset: apiSettings.IsTokenizedDigitalAsset,
		TaxRateInfoByRegionCode: regionalTaxRateInfoFromAPI(apiSettings.TaxRateInfoByRegionCode),
	}
}

func subscriptionTaxComplianceSettingsFromAPI(apiSettings *androidpublisher.SubscriptionTaxAndComplianceSettings) *ProductTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	return &ProductTaxComplianceSettings{
		EEAWithdrawalRightType:  apiSettings.EeaWithdrawalRightType,
		IsTokenizedDigitalAsset: apiSettings.IsTokenizedDigitalAsset,
		TaxRateInfoByRegionCode: regionalTaxRateInfoFromAPI(apiSettings.TaxRateInfoByRegionCode),
	}
}

func regionalTaxRateInfoFromAPI(apiTaxRateInfo map[string]androidpublisher.RegionalTaxRateInfo) map[string]RegionalTaxRateInfo {
	if len(apiTaxRateInfo) == 0 {
		return nil
	}
	taxRateInfo := make(map[string]RegionalTaxRateInfo, len(apiTaxRateInfo))
	for region, apiInfo := range apiTaxRateInfo {
		taxRateInfo[region] = RegionalTaxRateInfo{
			EligibleForStreamingServiceTaxRate: apiInfo.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   apiInfo.StreamingTaxType,
			TaxTier:                            apiInfo.TaxTier,
		}
	}
	return taxRateInfo
}

func subscriptionListResultFromAPI(options SubscriptionListOptions, response *androidpublisher.ListSubscriptionsResponse) SubscriptionListResult {
	result := SubscriptionListResult{
		PackageName:   options.PackageName,
		Subscriptions: []Subscription{},
		Options:       options,
	}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiSubscription := range response.Subscriptions {
		result.Subscriptions = append(result.Subscriptions, subscriptionFromAPI(apiSubscription))
	}
	return result
}

func subscriptionFromAPI(apiSubscription *androidpublisher.Subscription) Subscription {
	if apiSubscription == nil {
		return Subscription{Listings: []SubscriptionListing{}, BasePlans: []SubscriptionBasePlan{}}
	}
	return Subscription{
		PackageName:              PackageName(apiSubscription.PackageName),
		ProductID:                SubscriptionProductID(apiSubscription.ProductId),
		Archived:                 apiSubscription.Archived,
		Listings:                 subscriptionListingsFromAPI(apiSubscription.Listings),
		BasePlans:                subscriptionBasePlansFromAPI(apiSubscription.BasePlans),
		RestrictedCountries:      restrictedCountriesFromAPI(apiSubscription.RestrictedPaymentCountries),
		TaxAndComplianceSettings: subscriptionTaxComplianceSettingsFromAPI(apiSubscription.TaxAndComplianceSettings),
	}
}

func subscriptionListingsFromAPI(apiListings []*androidpublisher.SubscriptionListing) []SubscriptionListing {
	listings := make([]SubscriptionListing, 0, len(apiListings))
	for _, apiListing := range apiListings {
		if apiListing == nil {
			continue
		}
		listings = append(listings, SubscriptionListing{
			LanguageCode: apiListing.LanguageCode,
			Title:        apiListing.Title,
			Description:  apiListing.Description,
			Benefits:     apiListing.Benefits,
		})
	}
	return listings
}

func subscriptionBasePlansFromAPI(apiBasePlans []*androidpublisher.BasePlan) []SubscriptionBasePlan {
	basePlans := make([]SubscriptionBasePlan, 0, len(apiBasePlans))
	for _, apiBasePlan := range apiBasePlans {
		if apiBasePlan == nil {
			continue
		}
		basePlans = append(basePlans, subscriptionBasePlanFromAPI(apiBasePlan))
	}
	return basePlans
}

func subscriptionBasePlanFromAPI(apiBasePlan *androidpublisher.BasePlan) SubscriptionBasePlan {
	basePlan := SubscriptionBasePlan{
		BasePlanID:         apiBasePlan.BasePlanId,
		State:              SubscriptionState(apiBasePlan.State),
		OfferTags:          offerTagsFromAPI(apiBasePlan.OfferTags),
		RegionalConfigs:    subscriptionRegionalConfigsFromAPI(apiBasePlan.RegionalConfigs),
		OtherRegionsConfig: subscriptionOtherRegionsConfigFromAPI(apiBasePlan.OtherRegionsConfig),
	}
	switch {
	case apiBasePlan.AutoRenewingBasePlanType != nil:
		basePlan.Type = SubscriptionBasePlanTypeAutoRenewing
		basePlan.BillingPeriodDuration = apiBasePlan.AutoRenewingBasePlanType.BillingPeriodDuration
		basePlan.GracePeriodDuration = apiBasePlan.AutoRenewingBasePlanType.GracePeriodDuration
		basePlan.AccountHoldDuration = apiBasePlan.AutoRenewingBasePlanType.AccountHoldDuration
		basePlan.LegacyCompatible = apiBasePlan.AutoRenewingBasePlanType.LegacyCompatible
		basePlan.LegacyCompatibleSubscriptionOfferID = apiBasePlan.AutoRenewingBasePlanType.LegacyCompatibleSubscriptionOfferId
		basePlan.ProrationMode = apiBasePlan.AutoRenewingBasePlanType.ProrationMode
		basePlan.ResubscribeState = apiBasePlan.AutoRenewingBasePlanType.ResubscribeState
	case apiBasePlan.PrepaidBasePlanType != nil:
		basePlan.Type = SubscriptionBasePlanTypePrepaid
		basePlan.BillingPeriodDuration = apiBasePlan.PrepaidBasePlanType.BillingPeriodDuration
		basePlan.TimeExtension = apiBasePlan.PrepaidBasePlanType.TimeExtension
	case apiBasePlan.InstallmentsBasePlanType != nil:
		basePlan.Type = SubscriptionBasePlanTypeInstallments
		basePlan.BillingPeriodDuration = apiBasePlan.InstallmentsBasePlanType.BillingPeriodDuration
		basePlan.GracePeriodDuration = apiBasePlan.InstallmentsBasePlanType.GracePeriodDuration
		basePlan.AccountHoldDuration = apiBasePlan.InstallmentsBasePlanType.AccountHoldDuration
		basePlan.CommittedPaymentsCount = apiBasePlan.InstallmentsBasePlanType.CommittedPaymentsCount
		basePlan.ProrationMode = apiBasePlan.InstallmentsBasePlanType.ProrationMode
		basePlan.RenewalType = apiBasePlan.InstallmentsBasePlanType.RenewalType
		basePlan.ResubscribeState = apiBasePlan.InstallmentsBasePlanType.ResubscribeState
	}
	return basePlan
}

func subscriptionRegionalConfigsFromAPI(apiConfigs []*androidpublisher.RegionalBasePlanConfig) []SubscriptionRegionalConfig {
	if len(apiConfigs) == 0 {
		return nil
	}
	configs := make([]SubscriptionRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		configs = append(configs, SubscriptionRegionalConfig{
			RegionCode:                apiConfig.RegionCode,
			NewSubscriberAvailability: apiConfig.NewSubscriberAvailability,
			Price:                     moneyFromAPI(apiConfig.Price),
		})
	}
	return configs
}

func subscriptionOtherRegionsConfigFromAPI(apiConfig *androidpublisher.OtherRegionsBasePlanConfig) *SubscriptionOtherRegionsConfig {
	if apiConfig == nil {
		return nil
	}
	return &SubscriptionOtherRegionsConfig{
		NewSubscriberAvailability: apiConfig.NewSubscriberAvailability,
		USDPrice:                  moneyFromAPI(apiConfig.UsdPrice),
		EURPrice:                  moneyFromAPI(apiConfig.EurPrice),
	}
}

func moneyFromAPI(apiMoney *androidpublisher.Money) *Money {
	if apiMoney == nil {
		return nil
	}
	return &Money{
		CurrencyCode: apiMoney.CurrencyCode,
		Units:        apiMoney.Units,
		Nanos:        apiMoney.Nanos,
	}
}

func offerTagsFromAPI(apiOfferTags []*androidpublisher.OfferTag) []string {
	if len(apiOfferTags) == 0 {
		return nil
	}
	tags := make([]string, 0, len(apiOfferTags))
	for _, apiOfferTag := range apiOfferTags {
		if apiOfferTag == nil {
			continue
		}
		tags = append(tags, apiOfferTag.Tag)
	}
	return tags
}

func restrictedCountriesFromAPI(apiCountries *androidpublisher.RestrictedPaymentCountries) []string {
	if apiCountries == nil {
		return nil
	}
	return apiCountries.RegionCodes
}

func subscriptionOfferListResultFromAPI(options SubscriptionOfferListOptions, response *androidpublisher.ListSubscriptionOffersResponse) SubscriptionOfferListResult {
	result := SubscriptionOfferListResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Offers:      []SubscriptionOffer{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiOffer := range response.SubscriptionOffers {
		result.Offers = append(result.Offers, subscriptionOfferFromAPI(apiOffer))
	}
	return result
}

func subscriptionOfferFromAPI(apiOffer *androidpublisher.SubscriptionOffer) SubscriptionOffer {
	if apiOffer == nil {
		return SubscriptionOffer{RegionalConfigs: []SubscriptionOfferRegionalConfig{}, Phases: []SubscriptionOfferPhase{}}
	}
	return SubscriptionOffer{
		PackageName:        PackageName(apiOffer.PackageName),
		ProductID:          SubscriptionProductID(apiOffer.ProductId),
		BasePlanID:         SubscriptionBasePlanID(apiOffer.BasePlanId),
		OfferID:            SubscriptionOfferID(apiOffer.OfferId),
		State:              SubscriptionOfferState(apiOffer.State),
		OfferTags:          offerTagsFromAPI(apiOffer.OfferTags),
		RegionalConfigs:    subscriptionOfferRegionalConfigsFromAPI(apiOffer.RegionalConfigs),
		OtherRegionsConfig: subscriptionOfferOtherRegionsConfigFromAPI(apiOffer.OtherRegionsConfig),
		Phases:             subscriptionOfferPhasesFromAPI(apiOffer.Phases),
		Targeting:          subscriptionOfferTargetingFromAPI(apiOffer.Targeting),
	}
}

func subscriptionOfferRegionalConfigsFromAPI(apiConfigs []*androidpublisher.RegionalSubscriptionOfferConfig) []SubscriptionOfferRegionalConfig {
	configs := make([]SubscriptionOfferRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		configs = append(configs, SubscriptionOfferRegionalConfig{
			RegionCode:                apiConfig.RegionCode,
			NewSubscriberAvailability: apiConfig.NewSubscriberAvailability,
		})
	}
	return configs
}

func subscriptionOfferOtherRegionsConfigFromAPI(apiConfig *androidpublisher.OtherRegionsSubscriptionOfferConfig) *SubscriptionOfferOtherRegionsConfig {
	if apiConfig == nil {
		return nil
	}
	return &SubscriptionOfferOtherRegionsConfig{NewSubscriberAvailability: apiConfig.OtherRegionsNewSubscriberAvailability}
}

func subscriptionOfferPhasesFromAPI(apiPhases []*androidpublisher.SubscriptionOfferPhase) []SubscriptionOfferPhase {
	phases := make([]SubscriptionOfferPhase, 0, len(apiPhases))
	for _, apiPhase := range apiPhases {
		if apiPhase == nil {
			continue
		}
		phases = append(phases, SubscriptionOfferPhase{
			Duration:           apiPhase.Duration,
			RecurrenceCount:    apiPhase.RecurrenceCount,
			RegionalConfigs:    subscriptionOfferPhaseRegionalConfigsFromAPI(apiPhase.RegionalConfigs),
			OtherRegionsConfig: subscriptionOfferPhaseOtherRegionsConfigFromAPI(apiPhase.OtherRegionsConfig),
		})
	}
	return phases
}

func subscriptionOfferPhaseRegionalConfigsFromAPI(apiConfigs []*androidpublisher.RegionalSubscriptionOfferPhaseConfig) []SubscriptionOfferPhaseRegionalConfig {
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		configs = append(configs, SubscriptionOfferPhaseRegionalConfig{
			RegionCode:       apiConfig.RegionCode,
			Price:            moneyFromAPI(apiConfig.Price),
			AbsoluteDiscount: moneyFromAPI(apiConfig.AbsoluteDiscount),
			RelativeDiscount: apiConfig.RelativeDiscount,
			Free:             apiConfig.Free != nil,
		})
	}
	return configs
}

func subscriptionOfferPhaseOtherRegionsConfigFromAPI(apiConfig *androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig) *SubscriptionOfferPhaseOtherRegionsConfig {
	if apiConfig == nil {
		return nil
	}
	return &SubscriptionOfferPhaseOtherRegionsConfig{
		OtherRegionsPrices: otherRegionsSubscriptionOfferPhasePricesFromAPI(apiConfig.OtherRegionsPrices),
		AbsoluteDiscounts:  otherRegionsSubscriptionOfferPhasePricesFromAPI(apiConfig.AbsoluteDiscounts),
		RelativeDiscount:   apiConfig.RelativeDiscount,
		Free:               apiConfig.Free != nil,
	}
}

func otherRegionsSubscriptionOfferPhasePricesFromAPI(apiPrices *androidpublisher.OtherRegionsSubscriptionOfferPhasePrices) *SubscriptionOfferOtherRegionsPrices {
	if apiPrices == nil {
		return nil
	}
	return &SubscriptionOfferOtherRegionsPrices{
		USDPrice: moneyFromAPI(apiPrices.UsdPrice),
		EURPrice: moneyFromAPI(apiPrices.EurPrice),
	}
}

func subscriptionOfferTargetingFromAPI(apiTargeting *androidpublisher.SubscriptionOfferTargeting) *SubscriptionOfferTargeting {
	if apiTargeting == nil {
		return nil
	}
	return &SubscriptionOfferTargeting{
		Acquisition: subscriptionOfferAcquisitionTargetingFromAPI(apiTargeting.AcquisitionRule),
		Upgrade:     subscriptionOfferUpgradeTargetingFromAPI(apiTargeting.UpgradeRule),
	}
}

func subscriptionOfferAcquisitionTargetingFromAPI(apiRule *androidpublisher.AcquisitionTargetingRule) *SubscriptionOfferAcquisitionTargeting {
	if apiRule == nil {
		return nil
	}
	return &SubscriptionOfferAcquisitionTargeting{Scope: subscriptionOfferTargetingScopeFromAPI(apiRule.Scope)}
}

func subscriptionOfferUpgradeTargetingFromAPI(apiRule *androidpublisher.UpgradeTargetingRule) *SubscriptionOfferUpgradeTargeting {
	if apiRule == nil {
		return nil
	}
	return &SubscriptionOfferUpgradeTargeting{
		Scope:                 subscriptionOfferTargetingScopeFromAPI(apiRule.Scope),
		BillingPeriodDuration: apiRule.BillingPeriodDuration,
		OncePerUser:           apiRule.OncePerUser,
	}
}

func subscriptionOfferTargetingScopeFromAPI(apiScope *androidpublisher.TargetingRuleScope) *SubscriptionOfferTargetingScope {
	if apiScope == nil {
		return nil
	}
	return &SubscriptionOfferTargetingScope{
		AnySubscriptionInApp:      apiScope.AnySubscriptionInApp != nil,
		ThisSubscription:          apiScope.ThisSubscription != nil,
		SpecificSubscriptionInApp: apiScope.SpecificSubscriptionInApp,
	}
}

func productPurchaseFromAPI(options ProductPurchaseOptions, apiPurchase *androidpublisher.ProductPurchaseV2) ProductPurchase {
	if apiPurchase == nil {
		return ProductPurchase{PackageName: options.PackageName, ProductID: options.ProductID, Token: options.Token, LineItems: []ProductPurchaseLineItem{}}
	}
	return ProductPurchase{
		PackageName:                 options.PackageName,
		ProductID:                   firstProductID(options.ProductID, apiPurchase.ProductLineItem),
		Token:                       options.Token,
		OrderID:                     apiPurchase.OrderId,
		PurchaseState:               purchaseStateFromAPI(apiPurchase.PurchaseStateContext),
		PurchaseCompletionTime:      apiPurchase.PurchaseCompletionTime,
		AcknowledgementState:        apiPurchase.AcknowledgementState,
		RegionCode:                  apiPurchase.RegionCode,
		ObfuscatedExternalAccountID: apiPurchase.ObfuscatedExternalAccountId,
		ObfuscatedExternalProfileID: apiPurchase.ObfuscatedExternalProfileId,
		TestPurchase:                apiPurchase.TestPurchaseContext != nil,
		LineItems:                   productPurchaseLineItemsFromAPI(apiPurchase.ProductLineItem),
	}
}

func purchaseStateFromAPI(apiState *androidpublisher.PurchaseStateContext) string {
	if apiState == nil {
		return ""
	}
	return apiState.PurchaseState
}

func firstProductID(fallback InAppProductSKU, apiItems []*androidpublisher.ProductLineItem) InAppProductSKU {
	for _, apiItem := range apiItems {
		if apiItem != nil && apiItem.ProductId != "" {
			return InAppProductSKU(apiItem.ProductId)
		}
	}
	return fallback
}

func productPurchaseLineItemsFromAPI(apiItems []*androidpublisher.ProductLineItem) []ProductPurchaseLineItem {
	items := make([]ProductPurchaseLineItem, 0, len(apiItems))
	for _, apiItem := range apiItems {
		if apiItem == nil {
			continue
		}
		item := ProductPurchaseLineItem{ProductID: apiItem.ProductId}
		if apiItem.ProductOfferDetails != nil {
			item.ConsumptionState = apiItem.ProductOfferDetails.ConsumptionState
			item.PurchaseOptionID = apiItem.ProductOfferDetails.PurchaseOptionId
			item.OfferID = apiItem.ProductOfferDetails.OfferId
			item.OfferToken = apiItem.ProductOfferDetails.OfferToken
			item.OfferTags = apiItem.ProductOfferDetails.OfferTags
			item.Quantity = apiItem.ProductOfferDetails.Quantity
			item.RefundableQuantity = apiItem.ProductOfferDetails.RefundableQuantity
		}
		items = append(items, item)
	}
	return items
}

func subscriptionPurchaseFromAPI(packageName PackageName, token PurchaseToken, apiPurchase *androidpublisher.SubscriptionPurchaseV2) SubscriptionPurchase {
	if apiPurchase == nil {
		return SubscriptionPurchase{PackageName: packageName, Token: token, LineItems: []SubscriptionPurchaseLineItem{}}
	}
	return SubscriptionPurchase{
		PackageName:                 packageName,
		Token:                       token,
		SubscriptionState:           apiPurchase.SubscriptionState,
		AcknowledgementState:        apiPurchase.AcknowledgementState,
		LatestOrderID:               apiPurchase.LatestOrderId,
		LinkedPurchaseToken:         apiPurchase.LinkedPurchaseToken,
		RegionCode:                  apiPurchase.RegionCode,
		StartTime:                   apiPurchase.StartTime,
		LineItems:                   subscriptionPurchaseLineItemsFromAPI(apiPurchase.LineItems),
		ExternalAccountID:           externalAccountIDFromAPI(apiPurchase.ExternalAccountIdentifiers),
		ObfuscatedExternalAccountID: obfuscatedExternalAccountIDFromAPI(apiPurchase.ExternalAccountIdentifiers),
		ObfuscatedExternalProfileID: obfuscatedExternalProfileIDFromAPI(apiPurchase.ExternalAccountIdentifiers),
		TestPurchase:                apiPurchase.TestPurchase != nil,
	}
}

func subscriptionPurchaseLineItemsFromAPI(apiItems []*androidpublisher.SubscriptionPurchaseLineItem) []SubscriptionPurchaseLineItem {
	items := make([]SubscriptionPurchaseLineItem, 0, len(apiItems))
	for _, apiItem := range apiItems {
		if apiItem == nil {
			continue
		}
		item := SubscriptionPurchaseLineItem{
			ProductID:               apiItem.ProductId,
			ExpiryTime:              apiItem.ExpiryTime,
			LatestSuccessfulOrderID: apiItem.LatestSuccessfulOrderId,
		}
		if apiItem.OfferDetails != nil {
			item.BasePlanID = apiItem.OfferDetails.BasePlanId
			item.OfferID = apiItem.OfferDetails.OfferId
			item.OfferTags = apiItem.OfferDetails.OfferTags
		}
		if apiItem.AutoRenewingPlan != nil {
			autoRenewEnabled := apiItem.AutoRenewingPlan.AutoRenewEnabled
			item.AutoRenewEnabled = &autoRenewEnabled
			item.RecurringPrice = moneyFromAPI(apiItem.AutoRenewingPlan.RecurringPrice)
		}
		if apiItem.PrepaidPlan != nil {
			item.Prepaid = true
			item.AllowExtendAfterTime = apiItem.PrepaidPlan.AllowExtendAfterTime
		}
		items = append(items, item)
	}
	return items
}

func externalAccountIDFromAPI(apiIdentifiers *androidpublisher.ExternalAccountIdentifiers) string {
	if apiIdentifiers == nil {
		return ""
	}
	return apiIdentifiers.ExternalAccountId
}

func obfuscatedExternalAccountIDFromAPI(apiIdentifiers *androidpublisher.ExternalAccountIdentifiers) string {
	if apiIdentifiers == nil {
		return ""
	}
	return apiIdentifiers.ObfuscatedExternalAccountId
}

func obfuscatedExternalProfileIDFromAPI(apiIdentifiers *androidpublisher.ExternalAccountIdentifiers) string {
	if apiIdentifiers == nil {
		return ""
	}
	return apiIdentifiers.ObfuscatedExternalProfileId
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
