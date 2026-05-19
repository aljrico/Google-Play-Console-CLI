package play

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/aljrico/Google-Play-Console-CLI/internal/googleclient"
)

func NewPublisherFromActiveProfile(ctx context.Context) (*GooglePublisher, error) {
	httpClient, err := googleclient.ActiveProfileHTTPClient(ctx, androidpublisher.AndroidpublisherScope)
	if err != nil {
		return nil, err
	}
	service, err := androidpublisher.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create Google Play API service: %w", err)
	}
	return &GooglePublisher{service: service, httpClient: httpClient, basePath: service.BasePath}, nil
}

type GooglePublisher struct {
	service    *androidpublisher.Service
	httpClient *http.Client
	basePath   string
}

func (p GooglePublisher) doJSON(req *http.Request, target any) error {
	httpClient := p.httpClient
	if httpClient == nil {
		return fmt.Errorf("Google Play HTTP client is required")
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := googleapi.CheckResponse(response); err != nil {
		return err
	}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func (p GooglePublisher) doNoContent(req *http.Request) error {
	httpClient := p.httpClient
	if httpClient == nil {
		return fmt.Errorf("Google Play HTTP client is required")
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := googleapi.CheckResponse(response); err != nil {
		return err
	}
	return nil
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

func (p GooglePublisher) UploadAPK(ctx context.Context, packageName PackageName, editID string, apkPath string) (APKArtifact, error) {
	file, err := os.Open(apkPath)
	if err != nil {
		return APKArtifact{}, fmt.Errorf("open APK %s: %w", apkPath, err)
	}
	defer file.Close()

	apk, err := p.service.Edits.Apks.Upload(packageName.String(), editID).
		Media(file, googleapi.ContentType("application/octet-stream")).
		Context(ctx).
		Do()
	if err != nil {
		return APKArtifact{}, fmt.Errorf("upload APK %s for %s: %w", apkPath, packageName, err)
	}
	return apkArtifactFromAPI(apk), nil
}

func (p GooglePublisher) UploadInternalSharingAPK(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error) {
	file, err := os.Open(path)
	if err != nil {
		return InternalSharingArtifact{}, fmt.Errorf("open APK %s: %w", path, err)
	}
	defer file.Close()

	artifact, err := p.service.Internalappsharingartifacts.Uploadapk(packageName.String()).
		Media(file, googleapi.ContentType("application/octet-stream")).
		Context(ctx).
		Do()
	if err != nil {
		return InternalSharingArtifact{}, fmt.Errorf("upload internal sharing APK %s for %s: %w", path, packageName, err)
	}
	return internalSharingArtifactFromAPI(artifact), nil
}

func (p GooglePublisher) UploadInternalSharingBundle(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error) {
	file, err := os.Open(path)
	if err != nil {
		return InternalSharingArtifact{}, fmt.Errorf("open bundle %s: %w", path, err)
	}
	defer file.Close()

	artifact, err := p.service.Internalappsharingartifacts.Uploadbundle(packageName.String()).
		Media(file, googleapi.ContentType("application/octet-stream")).
		Context(ctx).
		Do()
	if err != nil {
		return InternalSharingArtifact{}, fmt.Errorf("upload internal sharing bundle %s for %s: %w", path, packageName, err)
	}
	return internalSharingArtifactFromAPI(artifact), nil
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

func (p GooglePublisher) GetTesters(ctx context.Context, packageName PackageName, editID string, track TrackName) (TrackTesters, error) {
	testers, err := p.service.Edits.Testers.Get(packageName.String(), editID, track.String()).Context(ctx).Do()
	if err != nil {
		return TrackTesters{}, fmt.Errorf("get %s testers for %s: %w", track, packageName, err)
	}
	return testersFromAPI(packageName, track, testers), nil
}

func (p GooglePublisher) UpdateTesters(ctx context.Context, packageName PackageName, editID string, track TrackName, googleGroups []TesterGoogleGroup) (TrackTesters, error) {
	apiTesters := testersToAPI(googleGroups)
	updatedTesters, err := p.service.Edits.Testers.Update(packageName.String(), editID, track.String(), apiTesters).Context(ctx).Do()
	if err != nil {
		return TrackTesters{}, fmt.Errorf("update %s testers for %s: %w", track, packageName, err)
	}
	return testersFromAPI(packageName, track, updatedTesters), nil
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

func (p GooglePublisher) ListImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error) {
	response, err := p.service.Edits.Images.List(packageName.String(), editID, language.String(), imageType.String()).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list %s images for %s %s listing: %w", imageType, packageName, language, err)
	}
	images := make([]StoreImage, 0, len(response.Images))
	for _, apiImage := range response.Images {
		if apiImage == nil {
			continue
		}
		images = append(images, imageFromAPI(apiImage))
	}
	return images, nil
}

func (p GooglePublisher) UploadImage(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType, path string) (StoreImage, error) {
	file, err := os.Open(path)
	if err != nil {
		return StoreImage{}, fmt.Errorf("open image %s: %w", path, err)
	}
	defer file.Close()

	response, err := p.service.Edits.Images.Upload(packageName.String(), editID, language.String(), imageType.String()).
		Media(file, googleapi.ContentType(ImageContentType(path))).
		Context(ctx).
		Do()
	if err != nil {
		return StoreImage{}, fmt.Errorf("upload %s image %s for %s %s listing: %w", imageType, path, packageName, language, err)
	}
	if response == nil || response.Image == nil {
		return StoreImage{}, fmt.Errorf("upload %s image %s for %s %s listing produced no image; verify the listing language is supported", imageType, path, packageName, language)
	}
	return imageFromAPI(response.Image), nil
}

func (p GooglePublisher) DeleteImage(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType, imageID string) error {
	if err := p.service.Edits.Images.Delete(packageName.String(), editID, language.String(), imageType.String(), imageID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete %s image %s for %s %s listing: %w", imageType, imageID, packageName, language, err)
	}
	return nil
}

func (p GooglePublisher) DeleteAllImages(ctx context.Context, packageName PackageName, editID string, language ListingLanguage, imageType ImageType) ([]StoreImage, error) {
	response, err := p.service.Edits.Images.Deleteall(packageName.String(), editID, language.String(), imageType.String()).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("delete all %s images for %s %s listing: %w", imageType, packageName, language, err)
	}
	if response == nil {
		return []StoreImage{}, nil
	}
	images := make([]StoreImage, 0, len(response.Deleted))
	for _, apiImage := range response.Deleted {
		if apiImage == nil {
			continue
		}
		images = append(images, imageFromAPI(apiImage))
	}
	return images, nil
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

func (p GooglePublisher) BatchGetInAppProducts(ctx context.Context, options InAppProductBatchGetOptions) (InAppProductBatchGetResult, error) {
	skus := make([]string, 0, len(options.SKUs))
	for _, sku := range options.SKUs {
		skus = append(skus, sku.String())
	}
	response, err := p.service.Inappproducts.BatchGet(options.PackageName.String()).
		Sku(skus...).
		Context(ctx).
		Do()
	if err != nil {
		return InAppProductBatchGetResult{}, fmt.Errorf("batch get in-app products for %s: %w", options.PackageName, err)
	}
	return inAppProductBatchGetResultFromAPI(options, response), nil
}

func (p GooglePublisher) DeleteInAppProduct(ctx context.Context, options InAppProductDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Inappproducts.Delete(options.PackageName.String(), options.SKU.String()).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("delete in-app product %s for %s: %w", options.SKU, options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) BatchDeleteInAppProducts(ctx context.Context, options InAppProductBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	request := &androidpublisher.InappproductsBatchDeleteRequest{
		Requests: make([]*androidpublisher.InappproductsDeleteRequest, 0, len(options.SKUs)),
	}
	for _, sku := range options.SKUs {
		request.Requests = append(request.Requests, &androidpublisher.InappproductsDeleteRequest{
			PackageName:      options.PackageName.String(),
			Sku:              sku.String(),
			LatencyTolerance: latencyTolerance,
		})
	}
	if err := p.service.Inappproducts.BatchDelete(options.PackageName.String(), request).Context(ctx).Do(); err != nil {
		return fmt.Errorf("batch delete in-app products for %s: %w", options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) CreateInAppProduct(ctx context.Context, options InAppProductCreateOptions) (InAppProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return InAppProduct{}, err
	}
	product, err := p.service.Inappproducts.Insert(options.PackageName.String(), inAppProductCreateToAPI(options)).
		AutoConvertMissingPrices(true).
		Context(ctx).
		Do()
	if err != nil {
		return InAppProduct{}, fmt.Errorf("create in-app product %s for %s: %w", options.SKU, options.PackageName, err)
	}
	return inAppProductFromAPI(product), nil
}

func (p GooglePublisher) PatchInAppProduct(ctx context.Context, options InAppProductPatchOptions) (InAppProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return InAppProduct{}, err
	}
	call := p.service.Inappproducts.Patch(options.PackageName.String(), options.SKU.String(), inAppProductPatchToAPI(options))
	if shouldAutoConvertInAppProductPatchPrices(options) {
		call.AutoConvertMissingPrices(true)
	}
	product, err := call.Context(ctx).Do()
	if err != nil {
		return InAppProduct{}, fmt.Errorf("patch in-app product %s for %s: %w", options.SKU, options.PackageName, err)
	}
	return inAppProductFromAPI(product), nil
}

func (p GooglePublisher) ListOneTimeProducts(ctx context.Context, options OneTimeProductListOptions) (OneTimeProductListResult, error) {
	requestURL := googleapi.ResolveRelative(p.basePath, "androidpublisher/v3/applications/{packageName}/oneTimeProducts")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return OneTimeProductListResult{}, fmt.Errorf("create one-time products list request: %w", err)
	}
	googleapi.Expand(req.URL, map[string]string{"packageName": options.PackageName.String()})
	query := req.URL.Query()
	query.Set("alt", "json")
	query.Set("prettyPrint", "false")
	if options.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(options.PageSize, 10))
	}
	if options.PageToken != "" {
		query.Set("pageToken", options.PageToken)
	}
	req.URL.RawQuery = query.Encode()
	var response rawOneTimeProductListResponse
	err = p.doJSON(req, &response)
	if err != nil {
		return OneTimeProductListResult{}, fmt.Errorf("list one-time products for %s: %w", options.PackageName, err)
	}
	return oneTimeProductListResultFromAPI(options, response), nil
}

func (p GooglePublisher) GetOneTimeProduct(ctx context.Context, packageName PackageName, productID OneTimeProductID) (OneTimeProduct, error) {
	requestURL := googleapi.ResolveRelative(p.basePath, "androidpublisher/v3/applications/{packageName}/oneTimeProducts/{productId}")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("create one-time product get request: %w", err)
	}
	googleapi.Expand(req.URL, map[string]string{
		"packageName": packageName.String(),
		"productId":   productID.String(),
	})
	query := req.URL.Query()
	query.Set("alt", "json")
	query.Set("prettyPrint", "false")
	req.URL.RawQuery = query.Encode()
	var product rawOneTimeProduct
	err = p.doJSON(req, &product)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("get one-time product %s for %s: %w", productID, packageName, err)
	}
	return oneTimeProductFromAPI(product), nil
}

func (p GooglePublisher) BatchGetOneTimeProducts(ctx context.Context, options OneTimeProductBatchGetOptions) (OneTimeProductBatchGetResult, error) {
	if err := options.Validate(); err != nil {
		return OneTimeProductBatchGetResult{}, err
	}
	productIDs := make([]string, 0, len(options.ProductIDs))
	for _, productID := range options.ProductIDs {
		productIDs = append(productIDs, productID.String())
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchGet(options.PackageName.String()).
		ProductIds(productIDs...).
		Context(ctx).
		Do()
	if err != nil {
		return OneTimeProductBatchGetResult{}, fmt.Errorf("batch get one-time products for %s: %w", options.PackageName, err)
	}
	return oneTimeProductBatchGetResultFromAPI(options, response)
}

func (p GooglePublisher) CreateOneTimeProduct(ctx context.Context, options OneTimeProductCreateOptions) (OneTimeProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProduct{}, err
	}
	existing, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do()
	if err == nil {
		if existing == nil {
			return OneTimeProduct{}, fmt.Errorf("one-time product %s for %s existence check returned an empty product", options.ProductID, options.PackageName)
		}
		return OneTimeProduct{}, fmt.Errorf("one-time product %s for %s already exists; use patch for updates", options.ProductID, options.PackageName)
	}
	if !isGoogleNotFound(err) {
		return OneTimeProduct{}, fmt.Errorf("check one-time product %s for %s before create: %w", options.ProductID, options.PackageName, err)
	}
	desired := oneTimeProductCreateDesiredProduct(options)
	product, err := p.service.Monetization.Onetimeproducts.Patch(options.PackageName.String(), options.ProductID.String(), oneTimeProductToAPI(desired)).
		AllowMissing(true).
		UpdateMask(oneTimeProductCreateUpdateMask).
		RegionsVersionVersion(options.RegionsVersion).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("create one-time product %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return oneTimeProductFromGeneratedAPI(product)
}

func isGoogleNotFound(err error) bool {
	apiError, ok := err.(*googleapi.Error)
	return ok && apiError.Code == http.StatusNotFound
}

func (p GooglePublisher) PatchOneTimeProduct(ctx context.Context, options OneTimeProductPatchOptions) (OneTimeProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProduct{}, err
	}
	current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("get one-time product %s for %s before patch: %w", options.ProductID, options.PackageName, err)
	}
	currentProduct, err := oneTimeProductFromGeneratedAPI(current)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("decode one-time product %s for %s before patch: %w", options.ProductID, options.PackageName, err)
	}
	mergedListings := mergeOneTimeProductListings(currentProduct.Listings, options)
	if err := validateMergedOneTimeProductListing(options.Listing.LanguageCode, mergedListings); err != nil {
		return OneTimeProduct{}, err
	}
	request := &androidpublisher.OneTimeProduct{
		PackageName: options.PackageName.String(),
		ProductId:   options.ProductID.String(),
		Listings:    oneTimeProductListingsToAPI(mergedListings),
	}
	product, err := p.service.Monetization.Onetimeproducts.Patch(options.PackageName.String(), options.ProductID.String(), request).
		UpdateMask(oneTimeProductPatchUpdateMask).
		RegionsVersionVersion(options.RegionsVersion).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("patch one-time product %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return oneTimeProductFromGeneratedAPI(product)
}

func (p GooglePublisher) BatchPatchOneTimeProductListings(ctx context.Context, options OneTimeProductBatchPatchListingsOptions) (OneTimeProductBatchPatchListingsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductBatchPatchListingsResult{}, err
	}
	requestsByProduct := oneTimeProductBatchListingPatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductsRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return OneTimeProductBatchPatchListingsResult{}, fmt.Errorf("get one-time product %s for %s before batch patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		currentProduct, err := oneTimeProductFromGeneratedAPI(current)
		if err != nil {
			return OneTimeProductBatchPatchListingsResult{}, fmt.Errorf("decode one-time product %s for %s before batch patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedListings := mergeOneTimeProductListingPatches(currentProduct.Listings, productPatch.Requests)
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductRequest{
			OneTimeProduct: &androidpublisher.OneTimeProduct{
				PackageName: options.PackageName.String(),
				ProductId:   productPatch.ProductID.String(),
				Listings:    oneTimeProductListingsToAPI(mergedListings),
			},
			UpdateMask:       oneTimeProductPatchUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return OneTimeProductBatchPatchListingsResult{}, fmt.Errorf("batch patch one-time product listings for %s: %w", options.PackageName, err)
	}
	products := make([]OneTimeProduct, 0)
	if response != nil {
		products = make([]OneTimeProduct, 0, len(response.OneTimeProducts))
		for _, apiProduct := range response.OneTimeProducts {
			product, err := oneTimeProductFromGeneratedAPI(apiProduct)
			if err != nil {
				return OneTimeProductBatchPatchListingsResult{}, err
			}
			products = append(products, product)
		}
	}
	return OneTimeProductBatchPatchListingsResult{
		PackageName: options.PackageName,
		DryRun:      false,
		Applied:     true,
		Products:    products,
	}, nil
}

type oneTimeProductBatchListingPatchProduct struct {
	ProductID OneTimeProductID
	Requests  []OneTimeProductBatchPatchListingRequest
}

func oneTimeProductBatchListingPatchRequestsByProduct(requests []OneTimeProductBatchPatchListingRequest) []oneTimeProductBatchListingPatchProduct {
	byProduct := map[OneTimeProductID]int{}
	products := make([]oneTimeProductBatchListingPatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, oneTimeProductBatchListingPatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

func (p GooglePublisher) DeleteOneTimeProduct(ctx context.Context, options OneTimeProductDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	err := p.service.Monetization.Onetimeproducts.Delete(
		options.PackageName.String(),
		options.ProductID.String(),
	).LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("delete one-time product %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) BatchDeleteOneTimeProducts(ctx context.Context, options OneTimeProductBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.BatchDeleteOneTimeProductsRequest{
		Requests: make([]*androidpublisher.DeleteOneTimeProductRequest, 0, len(options.ProductIDs)),
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	for _, productID := range options.ProductIDs {
		request.Requests = append(request.Requests, &androidpublisher.DeleteOneTimeProductRequest{
			PackageName:      options.PackageName.String(),
			ProductId:        productID.String(),
			LatencyTolerance: latencyTolerance,
		})
	}
	err := p.service.Monetization.Onetimeproducts.BatchDelete(options.PackageName.String(), request).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("batch delete one-time products for %s: %w", options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) BatchDeletePurchaseOptions(ctx context.Context, options PurchaseOptionBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.BatchDeletePurchaseOptionsRequest{
		Requests: make([]*androidpublisher.DeletePurchaseOptionRequest, 0, len(options.Requests)),
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	for _, item := range options.Requests {
		apiRequest := &androidpublisher.DeletePurchaseOptionRequest{
			PackageName:      options.PackageName.String(),
			ProductId:        item.ProductID.String(),
			PurchaseOptionId: item.PurchaseOptionID.String(),
			LatencyTolerance: latencyTolerance,
		}
		if options.Force {
			apiRequest.Force = true
			apiRequest.ForceSendFields = []string{"Force"}
		}
		request.Requests = append(request.Requests, apiRequest)
	}
	err := p.service.Monetization.Onetimeproducts.PurchaseOptions.BatchDelete(
		options.PackageName.String(),
		options.ParentProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("batch delete purchase options for %s/%s: %w", options.PackageName, options.ParentProductID, err)
	}
	return nil
}

func (p GooglePublisher) BatchPatchPurchaseOptionAvailability(ctx context.Context, options PurchaseOptionBatchPatchAvailabilityOptions) (PurchaseOptionBatchPatchAvailabilityResult, error) {
	if err := options.ValidateLive(); err != nil {
		return PurchaseOptionBatchPatchAvailabilityResult{}, err
	}
	requestsByProduct := purchaseOptionAvailabilityPatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductsRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return PurchaseOptionBatchPatchAvailabilityResult{}, fmt.Errorf("get one-time product %s for %s before purchase option availability patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		currentProduct, err := oneTimeProductFromGeneratedAPI(current)
		if err != nil {
			return PurchaseOptionBatchPatchAvailabilityResult{}, fmt.Errorf("decode one-time product %s for %s before purchase option availability patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedOptions, err := mergePurchaseOptionAvailabilityPatches(currentProduct.PurchaseOptions, productPatch.Requests)
		if err != nil {
			return PurchaseOptionBatchPatchAvailabilityResult{}, err
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductRequest{
			OneTimeProduct: &androidpublisher.OneTimeProduct{
				PackageName:     options.PackageName.String(),
				ProductId:       productPatch.ProductID.String(),
				PurchaseOptions: oneTimeProductPurchaseOptionsToAPI(mergedOptions),
			},
			UpdateMask:       purchaseOptionAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return PurchaseOptionBatchPatchAvailabilityResult{}, fmt.Errorf("batch patch purchase option availability for %s: %w", options.PackageName, err)
	}
	products := make([]OneTimeProduct, 0)
	if response != nil {
		products = make([]OneTimeProduct, 0, len(response.OneTimeProducts))
		for _, apiProduct := range response.OneTimeProducts {
			product, err := oneTimeProductFromGeneratedAPI(apiProduct)
			if err != nil {
				return PurchaseOptionBatchPatchAvailabilityResult{}, err
			}
			products = append(products, product)
		}
	}
	return PurchaseOptionBatchPatchAvailabilityResult{
		PackageName: options.PackageName,
		DryRun:      false,
		Applied:     true,
		Products:    products,
	}, nil
}

func (p GooglePublisher) BatchPatchPurchaseOptionPrices(ctx context.Context, options PurchaseOptionBatchPatchPriceOptions) (PurchaseOptionBatchPatchPriceResult, error) {
	if err := options.ValidateLive(); err != nil {
		return PurchaseOptionBatchPatchPriceResult{}, err
	}
	requestsByProduct := purchaseOptionPricePatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductsRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Onetimeproducts.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return PurchaseOptionBatchPatchPriceResult{}, fmt.Errorf("get one-time product %s for %s before purchase option price patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		currentProduct, err := oneTimeProductFromGeneratedAPI(current)
		if err != nil {
			return PurchaseOptionBatchPatchPriceResult{}, fmt.Errorf("decode one-time product %s for %s before purchase option price patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedOptions, err := mergePurchaseOptionPricePatches(currentProduct.PurchaseOptions, productPatch.Requests)
		if err != nil {
			return PurchaseOptionBatchPatchPriceResult{}, err
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductRequest{
			OneTimeProduct: &androidpublisher.OneTimeProduct{
				PackageName:     options.PackageName.String(),
				ProductId:       productPatch.ProductID.String(),
				PurchaseOptions: oneTimeProductPurchaseOptionsToAPI(mergedOptions),
			},
			UpdateMask:       purchaseOptionAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return PurchaseOptionBatchPatchPriceResult{}, fmt.Errorf("batch patch purchase option prices for %s: %w", options.PackageName, err)
	}
	products := make([]OneTimeProduct, 0)
	if response != nil {
		products = make([]OneTimeProduct, 0, len(response.OneTimeProducts))
		for _, apiProduct := range response.OneTimeProducts {
			product, err := oneTimeProductFromGeneratedAPI(apiProduct)
			if err != nil {
				return PurchaseOptionBatchPatchPriceResult{}, err
			}
			products = append(products, product)
		}
	}
	return PurchaseOptionBatchPatchPriceResult{
		PackageName: options.PackageName,
		DryRun:      false,
		Applied:     true,
		Products:    products,
	}, nil
}

type purchaseOptionAvailabilityPatchProduct struct {
	ProductID OneTimeProductID
	Requests  []PurchaseOptionAvailabilityPatchRequest
}

func purchaseOptionAvailabilityPatchRequestsByProduct(requests []PurchaseOptionAvailabilityPatchRequest) []purchaseOptionAvailabilityPatchProduct {
	byProduct := map[OneTimeProductID]int{}
	products := make([]purchaseOptionAvailabilityPatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, purchaseOptionAvailabilityPatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

type purchaseOptionPricePatchProduct struct {
	ProductID OneTimeProductID
	Requests  []PurchaseOptionPricePatchRequest
}

func purchaseOptionPricePatchRequestsByProduct(requests []PurchaseOptionPricePatchRequest) []purchaseOptionPricePatchProduct {
	byProduct := map[OneTimeProductID]int{}
	products := make([]purchaseOptionPricePatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, purchaseOptionPricePatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

func (p GooglePublisher) UpdatePurchaseOptionState(ctx context.Context, options PurchaseOptionStateUpdateOptions) (OneTimeProduct, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProduct{}, err
	}
	request := &androidpublisher.BatchUpdatePurchaseOptionStatesRequest{
		Requests: []*androidpublisher.UpdatePurchaseOptionStateRequest{
			purchaseOptionStateRequestToAPI(options),
		},
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("%s purchase option %s for %s/%s: %w", options.Action, options.PurchaseOptionID, options.PackageName, options.ProductID, err)
	}
	if response == nil || len(response.OneTimeProducts) == 0 {
		return OneTimeProduct{}, fmt.Errorf("%s purchase option %s for %s/%s: empty response", options.Action, options.PurchaseOptionID, options.PackageName, options.ProductID)
	}
	product, err := oneTimeProductFromGeneratedAPI(response.OneTimeProducts[0])
	if err != nil {
		return OneTimeProduct{}, err
	}
	return product, nil
}

func (p GooglePublisher) ListOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferListOptions) (OneTimeProductOfferListResult, error) {
	call := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.List(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
	).Context(ctx)
	if options.PageSize > 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	response, err := call.Do()
	if err != nil {
		return OneTimeProductOfferListResult{}, fmt.Errorf("list one-time product offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return oneTimeProductOfferListResultFromAPI(options, response), nil
}

func (p GooglePublisher) GetOneTimeProductOffer(ctx context.Context, options OneTimeProductOfferGetOptions) (OneTimeProductOffer, error) {
	request := batchGetOneTimeProductOffersRequestToAPI(options.PackageName, []OneTimeProductOfferBatchGetRequest{{
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		OfferID:          options.OfferID,
	}})
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchGet(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOffer{}, fmt.Errorf("get one-time product offer %s for %s/%s/%s: %w", options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	if response == nil || len(response.OneTimeProductOffers) == 0 {
		return OneTimeProductOffer{}, fmt.Errorf("get one-time product offer %s for %s/%s/%s: empty response", options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID)
	}
	return oneTimeProductOfferFromAPI(response.OneTimeProductOffers[0]), nil
}

func (p GooglePublisher) BatchGetOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchGetOptions) (OneTimeProductOfferBatchGetResult, error) {
	request := batchGetOneTimeProductOffersRequestToAPI(options.PackageName, options.Requests)
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchGet(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchGetResult{}, fmt.Errorf("batch get one-time product offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return oneTimeProductOfferBatchGetResultFromAPI(options, response), nil
}

func (p GooglePublisher) BatchDeleteOneTimeProductOffers(ctx context.Context, options OneTimeProductOfferBatchDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.BatchDeleteOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.DeleteOneTimeProductOfferRequest, 0, len(options.Requests)),
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, &androidpublisher.DeleteOneTimeProductOfferRequest{
			PackageName:      options.PackageName.String(),
			ProductId:        item.ProductID.String(),
			PurchaseOptionId: item.PurchaseOptionID.String(),
			OfferId:          item.OfferID.String(),
			LatencyTolerance: latencyTolerance,
		})
	}
	err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchDelete(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("batch delete one-time product offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return nil
}

func (p GooglePublisher) BatchUpdateOneTimeProductOfferStates(ctx context.Context, options OneTimeProductOfferBatchStateUpdateOptions) (OneTimeProductOfferBatchStateUpdateResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchStateUpdateResult{}, err
	}
	request := &androidpublisher.BatchUpdateOneTimeProductOfferStatesRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferStateRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, oneTimeProductOfferStateRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchStateUpdateResult{}, fmt.Errorf("batch %s one-time product offers for %s/%s/%s: %w", options.Action, options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	offers := make([]OneTimeProductOffer, 0)
	if response != nil {
		offers = make([]OneTimeProductOffer, 0, len(response.OneTimeProductOffers))
		for _, offer := range response.OneTimeProductOffers {
			offers = append(offers, oneTimeProductOfferFromAPI(offer))
		}
	}
	return OneTimeProductOfferBatchStateUpdateResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferBatchMutationRequest(nil), options.Requests...),
		Action:           options.Action,
		Applied:          true,
		Offers:           offers,
	}, nil
}

func (p GooglePublisher) BatchPatchOneTimeProductOfferAvailability(ctx context.Context, options OneTimeProductOfferBatchPatchAvailabilityOptions) (OneTimeProductOfferBatchPatchAvailabilityResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchPatchAvailabilityResult{}, err
	}
	requestsByOffer := oneTimeProductOfferAvailabilityPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.getOneTimeProductOfferForAvailabilityPatch(ctx, options.PackageName, offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		if err != nil {
			return OneTimeProductOfferBatchPatchAvailabilityResult{}, err
		}
		mergedRegionalConfigs := mergeOneTimeProductOfferAvailabilityPatches(oneTimeProductOfferRegionsFromAPI(current.RegionalPricingAndAvailabilityConfigs), offerPatch.Requests)
		if mergedRegionalConfigs == nil {
			return OneTimeProductOfferBatchPatchAvailabilityResult{}, fmt.Errorf("one-time product offer availability patch for %s/%s/%s references a region that is not already configured on the offer; availability-only patches cannot add regional price overrides", offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductOfferRequest{
			OneTimeProductOffer: &androidpublisher.OneTimeProductOffer{
				PackageName:                           options.PackageName.String(),
				ProductId:                             offerPatch.ProductID.String(),
				PurchaseOptionId:                      offerPatch.PurchaseOptionID.String(),
				OfferId:                               offerPatch.OfferID.String(),
				RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       oneTimeProductOfferAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchPatchAvailabilityResult{}, fmt.Errorf("batch patch one-time product offer availability for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return OneTimeProductOfferBatchPatchAvailabilityResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferAvailabilityPatchRequest(nil), options.Requests...),
		Applied:          true,
		Offers:           oneTimeProductOffersFromBatchUpdateResponse(response),
	}, nil
}

func (p GooglePublisher) BatchPatchOneTimeProductOfferRelativeDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchRelativeDiscountsOptions) (OneTimeProductOfferBatchPatchRelativeDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, err
	}
	requestsByOffer := oneTimeProductOfferRelativeDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.getOneTimeProductOfferForRegionalPatch(ctx, options.PackageName, offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID, "relative discount")
		if err != nil {
			return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, err
		}
		mergedRegionalConfigs := mergeOneTimeProductOfferRelativeDiscountPatches(oneTimeProductOfferRegionsFromAPI(current.RegionalPricingAndAvailabilityConfigs), offerPatch.Requests)
		if mergedRegionalConfigs == nil {
			return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, fmt.Errorf("one-time product offer relative discount patch for %s/%s/%s references a region that is not already configured on the offer; relative discount patches cannot add regional price overrides", offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductOfferRequest{
			OneTimeProductOffer: &androidpublisher.OneTimeProductOffer{
				PackageName:                           options.PackageName.String(),
				ProductId:                             offerPatch.ProductID.String(),
				PurchaseOptionId:                      offerPatch.PurchaseOptionID.String(),
				OfferId:                               offerPatch.OfferID.String(),
				RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       oneTimeProductOfferRegionalConfigsUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchPatchRelativeDiscountsResult{}, fmt.Errorf("batch patch one-time product offer relative discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return OneTimeProductOfferBatchPatchRelativeDiscountsResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferRelativeDiscountPatchRequest(nil), options.Requests...),
		Applied:          true,
		Offers:           oneTimeProductOffersFromBatchUpdateResponse(response),
	}, nil
}

func (p GooglePublisher) BatchPatchOneTimeProductOfferAbsoluteDiscounts(ctx context.Context, options OneTimeProductOfferBatchPatchAbsoluteDiscountsOptions) (OneTimeProductOfferBatchPatchAbsoluteDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, err
	}
	requestsByOffer := oneTimeProductOfferAbsoluteDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.UpdateOneTimeProductOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.getOneTimeProductOfferForRegionalPatch(ctx, options.PackageName, offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID, "absolute discount")
		if err != nil {
			return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, err
		}
		mergedRegionalConfigs := mergeOneTimeProductOfferAbsoluteDiscountPatches(oneTimeProductOfferRegionsFromAPI(current.RegionalPricingAndAvailabilityConfigs), offerPatch.Requests)
		if mergedRegionalConfigs == nil {
			return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, fmt.Errorf("one-time product offer absolute discount patch for %s/%s/%s references a region that is not already configured on the offer; absolute discount patches cannot add regional price overrides", offerPatch.ProductID, offerPatch.PurchaseOptionID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateOneTimeProductOfferRequest{
			OneTimeProductOffer: &androidpublisher.OneTimeProductOffer{
				PackageName:                           options.PackageName.String(),
				ProductId:                             offerPatch.ProductID.String(),
				PurchaseOptionId:                      offerPatch.PurchaseOptionID.String(),
				OfferId:                               offerPatch.OfferID.String(),
				RegionalPricingAndAvailabilityConfigs: oneTimeProductOfferRegionsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       oneTimeProductOfferRegionalConfigsUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.PurchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{}, fmt.Errorf("batch patch one-time product offer absolute discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return OneTimeProductOfferBatchPatchAbsoluteDiscountsResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Requests:         append([]OneTimeProductOfferAbsoluteDiscountPatchRequest(nil), options.Requests...),
		Applied:          true,
		Offers:           oneTimeProductOffersFromBatchUpdateResponse(response),
	}, nil
}

func (p GooglePublisher) getOneTimeProductOfferForAvailabilityPatch(ctx context.Context, packageName PackageName, productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, offerID OneTimeProductOfferID) (*androidpublisher.OneTimeProductOffer, error) {
	return p.getOneTimeProductOfferForRegionalPatch(ctx, packageName, productID, purchaseOptionID, offerID, "availability")
}

func (p GooglePublisher) getOneTimeProductOfferForRegionalPatch(ctx context.Context, packageName PackageName, productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, offerID OneTimeProductOfferID, patchKind string) (*androidpublisher.OneTimeProductOffer, error) {
	request := batchGetOneTimeProductOffersRequestToAPI(packageName, []OneTimeProductOfferBatchGetRequest{{
		ProductID:        productID,
		PurchaseOptionID: purchaseOptionID,
		OfferID:          offerID,
	}})
	response, err := p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.BatchGet(
		packageName.String(),
		productID.String(),
		purchaseOptionID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get one-time product offer %s for %s/%s/%s before %s patch: %w", offerID, packageName, productID, purchaseOptionID, patchKind, err)
	}
	if response == nil || len(response.OneTimeProductOffers) == 0 {
		return nil, fmt.Errorf("get one-time product offer %s for %s/%s/%s before %s patch: empty response", offerID, packageName, productID, purchaseOptionID, patchKind)
	}
	return response.OneTimeProductOffers[0], nil
}

func (p GooglePublisher) UpdateOneTimeProductOfferState(ctx context.Context, options OneTimeProductOfferStateUpdateOptions) (OneTimeProductOffer, error) {
	if err := options.ValidateLive(); err != nil {
		return OneTimeProductOffer{}, err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	var (
		offer *androidpublisher.OneTimeProductOffer
		err   error
	)
	switch options.Action {
	case OneTimeProductOfferStateActionActivate:
		offer, err = p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.Activate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.PurchaseOptionID.String(),
			options.OfferID.String(),
			&androidpublisher.ActivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case OneTimeProductOfferStateActionDeactivate:
		offer, err = p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.Deactivate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.PurchaseOptionID.String(),
			options.OfferID.String(),
			&androidpublisher.DeactivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case OneTimeProductOfferStateActionCancel:
		offer, err = p.service.Monetization.Onetimeproducts.PurchaseOptions.Offers.Cancel(
			options.PackageName.String(),
			options.ProductID.String(),
			options.PurchaseOptionID.String(),
			options.OfferID.String(),
			&androidpublisher.CancelOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	default:
		return OneTimeProductOffer{}, fmt.Errorf("unsupported one-time product offer state action %q", options.Action)
	}
	if err != nil {
		return OneTimeProductOffer{}, fmt.Errorf("%s one-time product offer %s for %s/%s/%s: %w", options.Action, options.OfferID, options.PackageName, options.ProductID, options.PurchaseOptionID, err)
	}
	return oneTimeProductOfferFromAPI(offer), nil
}

func oneTimeProductOfferStateRequestToAPI(options OneTimeProductOfferBatchStateUpdateOptions, item OneTimeProductOfferBatchMutationRequest) *androidpublisher.UpdateOneTimeProductOfferStateRequest {
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	switch options.Action {
	case OneTimeProductOfferStateActionActivate:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{
			ActivateOneTimeProductOfferRequest: &androidpublisher.ActivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				PurchaseOptionId: item.PurchaseOptionID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case OneTimeProductOfferStateActionDeactivate:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{
			DeactivateOneTimeProductOfferRequest: &androidpublisher.DeactivateOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				PurchaseOptionId: item.PurchaseOptionID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case OneTimeProductOfferStateActionCancel:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{
			CancelOneTimeProductOfferRequest: &androidpublisher.CancelOneTimeProductOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				PurchaseOptionId: item.PurchaseOptionID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	default:
		return &androidpublisher.UpdateOneTimeProductOfferStateRequest{}
	}
}

type oneTimeProductOfferAvailabilityPatchOffer struct {
	ProductID        OneTimeProductID
	PurchaseOptionID OneTimeProductPurchaseOptionID
	OfferID          OneTimeProductOfferID
	Requests         []OneTimeProductOfferAvailabilityPatchRequest
}

type oneTimeProductOfferRelativeDiscountPatchOffer struct {
	ProductID        OneTimeProductID
	PurchaseOptionID OneTimeProductPurchaseOptionID
	OfferID          OneTimeProductOfferID
	Requests         []OneTimeProductOfferRelativeDiscountPatchRequest
}

type oneTimeProductOfferAbsoluteDiscountPatchOffer struct {
	ProductID        OneTimeProductID
	PurchaseOptionID OneTimeProductPurchaseOptionID
	OfferID          OneTimeProductOfferID
	Requests         []OneTimeProductOfferAbsoluteDiscountPatchRequest
}

func oneTimeProductOfferAvailabilityPatchRequestsByOffer(requests []OneTimeProductOfferAvailabilityPatchRequest) []oneTimeProductOfferAvailabilityPatchOffer {
	byOffer := map[string]int{}
	offers := make([]oneTimeProductOfferAvailabilityPatchOffer, 0)
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, oneTimeProductOfferAvailabilityPatchOffer{
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func oneTimeProductOfferRelativeDiscountPatchRequestsByOffer(requests []OneTimeProductOfferRelativeDiscountPatchRequest) []oneTimeProductOfferRelativeDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]oneTimeProductOfferRelativeDiscountPatchOffer, 0)
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, oneTimeProductOfferRelativeDiscountPatchOffer{
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func oneTimeProductOfferAbsoluteDiscountPatchRequestsByOffer(requests []OneTimeProductOfferAbsoluteDiscountPatchRequest) []oneTimeProductOfferAbsoluteDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]oneTimeProductOfferAbsoluteDiscountPatchOffer, 0)
	for _, request := range requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, oneTimeProductOfferAbsoluteDiscountPatchOffer{
				ProductID:        request.ProductID,
				PurchaseOptionID: request.PurchaseOptionID,
				OfferID:          request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func oneTimeProductOffersFromBatchUpdateResponse(response *androidpublisher.BatchUpdateOneTimeProductOffersResponse) []OneTimeProductOffer {
	if response == nil {
		return []OneTimeProductOffer{}
	}
	offers := make([]OneTimeProductOffer, 0, len(response.OneTimeProductOffers))
	for _, apiOffer := range response.OneTimeProductOffers {
		offers = append(offers, oneTimeProductOfferFromAPI(apiOffer))
	}
	return offers
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

func (p GooglePublisher) CreateSubscription(ctx context.Context, options SubscriptionCreateOptions) (Subscription, error) {
	if err := options.ValidateLive(); err != nil {
		return Subscription{}, err
	}
	subscription, err := p.service.Monetization.Subscriptions.Create(
		options.PackageName.String(),
		subscriptionCreateToAPI(options),
	).
		ProductId(options.ProductID.String()).
		RegionsVersionVersion(options.RegionsVersion).
		Context(ctx).
		Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("create subscription %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p GooglePublisher) BatchGetSubscriptions(ctx context.Context, options SubscriptionBatchGetOptions) (SubscriptionBatchGetResult, error) {
	productIDs := make([]string, 0, len(options.ProductIDs))
	for _, productID := range options.ProductIDs {
		productIDs = append(productIDs, productID.String())
	}
	response, err := p.service.Monetization.Subscriptions.BatchGet(options.PackageName.String()).
		ProductIds(productIDs...).
		Context(ctx).
		Do()
	if err != nil {
		return SubscriptionBatchGetResult{}, fmt.Errorf("batch get subscriptions for %s: %w", options.PackageName, err)
	}
	return subscriptionBatchGetResultFromAPI(options, response), nil
}

func (p GooglePublisher) DeleteSubscription(ctx context.Context, options SubscriptionDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Monetization.Subscriptions.Delete(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete subscription %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) DeleteBasePlan(ctx context.Context, options BasePlanDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Monetization.Subscriptions.BasePlans.Delete(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
	).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete base plan %s for %s/%s: %w", options.BasePlanID, options.PackageName, options.ProductID, err)
	}
	return nil
}

func (p GooglePublisher) PatchSubscription(ctx context.Context, options SubscriptionPatchOptions) (Subscription, error) {
	if err := options.ValidateLive(); err != nil {
		return Subscription{}, err
	}
	current, err := p.service.Monetization.Subscriptions.Get(options.PackageName.String(), options.ProductID.String()).Context(ctx).Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("get subscription %s for %s before patch: %w", options.ProductID, options.PackageName, err)
	}
	mergedListings := mergeSubscriptionListings(subscriptionListingsFromAPI(current.Listings), options)
	request := &androidpublisher.Subscription{
		PackageName: options.PackageName.String(),
		ProductId:   options.ProductID.String(),
		Listings:    subscriptionListingsToAPI(mergedListings),
	}
	subscription, err := p.service.Monetization.Subscriptions.Patch(options.PackageName.String(), options.ProductID.String(), request).
		UpdateMask(subscriptionPatchUpdateMask).
		RegionsVersionVersion(options.RegionsVersion).
		LatencyTolerance(productUpdateLatencyToleranceToAPI(options.LatencyTolerance)).
		Context(ctx).
		Do()
	if err != nil {
		return Subscription{}, fmt.Errorf("patch subscription %s for %s: %w", options.ProductID, options.PackageName, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p GooglePublisher) BatchPatchSubscriptionListings(ctx context.Context, options SubscriptionBatchPatchListingsOptions) (SubscriptionBatchPatchListingsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionBatchPatchListingsResult{}, err
	}
	requestsByProduct := subscriptionBatchListingPatchRequestsByProduct(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionsRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionRequest, 0, len(requestsByProduct)),
	}
	for _, productPatch := range requestsByProduct {
		current, err := p.service.Monetization.Subscriptions.Get(options.PackageName.String(), productPatch.ProductID.String()).Context(ctx).Do()
		if err != nil {
			return SubscriptionBatchPatchListingsResult{}, fmt.Errorf("get subscription %s for %s before batch patch: %w", productPatch.ProductID, options.PackageName, err)
		}
		mergedListings := mergeSubscriptionListingPatches(subscriptionListingsFromAPI(current.Listings), productPatch.Requests)
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionRequest{
			Subscription: &androidpublisher.Subscription{
				PackageName: options.PackageName.String(),
				ProductId:   productPatch.ProductID.String(),
				Listings:    subscriptionListingsToAPI(mergedListings),
			},
			UpdateMask:       subscriptionPatchUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BatchUpdate(options.PackageName.String(), request).Context(ctx).Do()
	if err != nil {
		return SubscriptionBatchPatchListingsResult{}, fmt.Errorf("batch patch subscription listings for %s: %w", options.PackageName, err)
	}
	subscriptions := make([]Subscription, 0)
	if response != nil {
		subscriptions = make([]Subscription, 0, len(response.Subscriptions))
		for _, apiSubscription := range response.Subscriptions {
			subscriptions = append(subscriptions, subscriptionFromAPI(apiSubscription))
		}
	}
	return SubscriptionBatchPatchListingsResult{
		PackageName:   options.PackageName,
		DryRun:        false,
		Applied:       true,
		Subscriptions: subscriptions,
	}, nil
}

type subscriptionBatchListingPatchProduct struct {
	ProductID SubscriptionProductID
	Requests  []SubscriptionBatchPatchListingRequest
}

func subscriptionBatchListingPatchRequestsByProduct(requests []SubscriptionBatchPatchListingRequest) []subscriptionBatchListingPatchProduct {
	byProduct := map[SubscriptionProductID]int{}
	products := make([]subscriptionBatchListingPatchProduct, 0)
	for _, request := range requests {
		index, ok := byProduct[request.ProductID]
		if !ok {
			byProduct[request.ProductID] = len(products)
			products = append(products, subscriptionBatchListingPatchProduct{ProductID: request.ProductID})
			index = len(products) - 1
		}
		products[index].Requests = append(products[index].Requests, request)
	}
	return products
}

func (p GooglePublisher) UpdateBasePlanState(ctx context.Context, options BasePlanStateUpdateOptions) (Subscription, error) {
	if err := options.ValidateLive(); err != nil {
		return Subscription{}, err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	var (
		subscription *androidpublisher.Subscription
		err          error
	)
	switch options.Action {
	case BasePlanStateActionActivate:
		subscription, err = p.service.Monetization.Subscriptions.BasePlans.Activate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			&androidpublisher.ActivateBasePlanRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case BasePlanStateActionDeactivate:
		subscription, err = p.service.Monetization.Subscriptions.BasePlans.Deactivate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			&androidpublisher.DeactivateBasePlanRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	default:
		return Subscription{}, fmt.Errorf("unsupported base plan state action %q", options.Action)
	}
	if err != nil {
		return Subscription{}, fmt.Errorf("%s base plan %s for %s/%s: %w", options.Action, options.BasePlanID, options.PackageName, options.ProductID, err)
	}
	return subscriptionFromAPI(subscription), nil
}

func (p GooglePublisher) BatchUpdateBasePlanStates(ctx context.Context, options BasePlanBatchStateUpdateOptions) (BasePlanBatchStateUpdateResult, error) {
	if err := options.ValidateLive(); err != nil {
		return BasePlanBatchStateUpdateResult{}, err
	}
	request := &androidpublisher.BatchUpdateBasePlanStatesRequest{
		Requests: make([]*androidpublisher.UpdateBasePlanStateRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, basePlanStateRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return BasePlanBatchStateUpdateResult{}, fmt.Errorf("batch %s base plans for %s/%s: %w", options.Action, options.PackageName, options.ProductID, err)
	}
	subscriptions := make([]Subscription, 0)
	if response != nil {
		subscriptions = make([]Subscription, 0, len(response.Subscriptions))
		for _, apiSubscription := range response.Subscriptions {
			subscriptions = append(subscriptions, subscriptionFromAPI(apiSubscription))
		}
	}
	return BasePlanBatchStateUpdateResult{
		PackageName:   options.PackageName,
		ProductID:     options.ProductID,
		Requests:      options.Requests,
		Action:        options.Action,
		DryRun:        false,
		Applied:       true,
		Subscriptions: subscriptions,
	}, nil
}

func (p GooglePublisher) BatchMigrateBasePlanPrices(ctx context.Context, options BasePlanBatchPriceMigrationOptions) (BasePlanBatchPriceMigrationResult, error) {
	if err := options.ValidateLive(); err != nil {
		return BasePlanBatchPriceMigrationResult{}, err
	}
	request := &androidpublisher.BatchMigrateBasePlanPricesRequest{
		Requests: make([]*androidpublisher.MigrateBasePlanPricesRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, basePlanPriceMigrationRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.BatchMigratePrices(
		options.PackageName.String(),
		options.ProductID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return BasePlanBatchPriceMigrationResult{}, fmt.Errorf("batch migrate base plan prices for %s/%s: %w", options.PackageName, options.ProductID, err)
	}
	responses := basePlanPriceMigrationResponsesFromAPI(options, response)
	return BasePlanBatchPriceMigrationResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		DryRun:      false,
		Applied:     true,
		Responses:   responses,
	}, nil
}

func basePlanPriceMigrationRequestToAPI(options BasePlanBatchPriceMigrationOptions, item BasePlanPriceMigrationRequest) *androidpublisher.MigrateBasePlanPricesRequest {
	return &androidpublisher.MigrateBasePlanPricesRequest{
		PackageName:             options.PackageName.String(),
		ProductId:               item.ProductID.String(),
		BasePlanId:              item.BasePlanID.String(),
		RegionsVersion:          &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
		LatencyTolerance:        productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		RegionalPriceMigrations: regionalPriceMigrationsToAPI(item.Regions),
	}
}

func regionalPriceMigrationsToAPI(regions []BasePlanPriceMigrationConfig) []*androidpublisher.RegionalPriceMigrationConfig {
	apiRegions := make([]*androidpublisher.RegionalPriceMigrationConfig, 0, len(regions))
	for _, region := range regions {
		apiRegions = append(apiRegions, &androidpublisher.RegionalPriceMigrationConfig{
			RegionCode:                    region.RegionCode,
			OldestAllowedPriceVersionTime: region.OldestAllowedPriceVersionTime,
			PriceIncreaseType:             basePlanPriceIncreaseTypeToAPI(region.PriceIncreaseType),
		})
	}
	return apiRegions
}

func basePlanPriceIncreaseTypeToAPI(value BasePlanPriceIncreaseType) string {
	switch value {
	case BasePlanPriceIncreaseTypeOptIn:
		return "PRICE_INCREASE_TYPE_OPT_IN"
	case BasePlanPriceIncreaseTypeOptOut:
		return "PRICE_INCREASE_TYPE_OPT_OUT"
	default:
		return ""
	}
}

func basePlanPriceMigrationResponsesFromAPI(options BasePlanBatchPriceMigrationOptions, response *androidpublisher.BatchMigrateBasePlanPricesResponse) []BasePlanPriceMigrationResponse {
	if response == nil {
		return []BasePlanPriceMigrationResponse{}
	}
	count := len(response.Responses)
	if count > len(options.Requests) {
		count = len(options.Requests)
	}
	responses := make([]BasePlanPriceMigrationResponse, 0, count)
	for _, request := range options.Requests[:count] {
		responses = append(responses, BasePlanPriceMigrationResponse{ProductID: request.ProductID, BasePlanID: request.BasePlanID})
	}
	return responses
}

func basePlanStateRequestToAPI(options BasePlanBatchStateUpdateOptions, item BasePlanBatchStateUpdateRequest) *androidpublisher.UpdateBasePlanStateRequest {
	switch options.Action {
	case BasePlanStateActionActivate:
		return &androidpublisher.UpdateBasePlanStateRequest{ActivateBasePlanRequest: activateBasePlanRequestToAPI(options, item)}
	case BasePlanStateActionDeactivate:
		return &androidpublisher.UpdateBasePlanStateRequest{DeactivateBasePlanRequest: deactivateBasePlanRequestToAPI(options, item)}
	default:
		return &androidpublisher.UpdateBasePlanStateRequest{}
	}
}

func activateBasePlanRequestToAPI(options BasePlanBatchStateUpdateOptions, item BasePlanBatchStateUpdateRequest) *androidpublisher.ActivateBasePlanRequest {
	return &androidpublisher.ActivateBasePlanRequest{
		PackageName:      options.PackageName.String(),
		ProductId:        item.ProductID.String(),
		BasePlanId:       item.BasePlanID.String(),
		LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
	}
}

func deactivateBasePlanRequestToAPI(options BasePlanBatchStateUpdateOptions, item BasePlanBatchStateUpdateRequest) *androidpublisher.DeactivateBasePlanRequest {
	return &androidpublisher.DeactivateBasePlanRequest{
		PackageName:      options.PackageName.String(),
		ProductId:        item.ProductID.String(),
		BasePlanId:       item.BasePlanID.String(),
		LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
	}
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

func subscriptionBatchGetResultFromAPI(options SubscriptionBatchGetOptions, response *androidpublisher.BatchGetSubscriptionsResponse) SubscriptionBatchGetResult {
	result := SubscriptionBatchGetResult{
		PackageName:   options.PackageName,
		Subscriptions: []Subscription{},
		Options:       options,
	}
	if response == nil {
		return result
	}
	for _, apiSubscription := range response.Subscriptions {
		result.Subscriptions = append(result.Subscriptions, subscriptionFromAPI(apiSubscription))
	}
	return result
}

func (p GooglePublisher) ListUsers(ctx context.Context, options UserListOptions) (UserListResult, error) {
	if err := options.Validate(); err != nil {
		return UserListResult{}, err
	}
	call := p.service.Users.List(options.Developer.ResourceName()).Context(ctx)
	if options.PageSize != 0 {
		call.PageSize(options.PageSize)
	}
	if options.PageToken != "" {
		call.PageToken(options.PageToken)
	}
	response, err := call.Do()
	if err != nil {
		return UserListResult{}, fmt.Errorf("list users for %s: %w", options.Developer.ResourceName(), err)
	}
	return userListResultFromAPI(options.Developer, response), nil
}

func (p GooglePublisher) CreateUser(ctx context.Context, options UserCreateOptions) (User, error) {
	if err := options.ValidateLive(); err != nil {
		return User{}, err
	}
	apiUser, err := p.service.Users.Create(options.Developer.ResourceName(), userCreateToAPI(options)).
		Context(ctx).
		Do()
	if err != nil {
		return User{}, fmt.Errorf("create user %s under %s: %w", options.UserEmail, options.Developer.ResourceName(), err)
	}
	return userFromAPI(apiUser), nil
}

func (p GooglePublisher) PatchUser(ctx context.Context, options UserPatchOptions) (User, error) {
	if err := options.ValidateLive(); err != nil {
		return User{}, err
	}
	apiUser, err := p.service.Users.Patch(options.Name.String(), userPatchToAPI(options)).
		UpdateMask(options.UpdateMask()).
		Context(ctx).
		Do()
	if err != nil {
		return User{}, fmt.Errorf("patch user %s: %w", options.Name, err)
	}
	return userFromAPI(apiUser), nil
}

func (p GooglePublisher) DeleteUser(ctx context.Context, options UserDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Users.Delete(options.Name.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete user %s: %w", options.Name, err)
	}
	return nil
}

func (p GooglePublisher) ListDeviceTierConfigs(ctx context.Context, options DeviceTierConfigListOptions) (DeviceTierConfigListResult, error) {
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

func (p GooglePublisher) GetDeviceTierConfig(ctx context.Context, options DeviceTierConfigGetOptions) (DeviceTierConfigGetResult, error) {
	apiConfig, err := p.service.Applications.DeviceTierConfigs.Get(options.PackageName.String(), options.DeviceTierConfigID).Context(ctx).Do()
	if err != nil {
		return DeviceTierConfigGetResult{}, fmt.Errorf("get device tier config %d for %s: %w", options.DeviceTierConfigID, options.PackageName, err)
	}
	return DeviceTierConfigGetResult{
		PackageName: options.PackageName,
		Config:      deviceTierConfigFromAPI(apiConfig),
	}, nil
}

func (p GooglePublisher) UpdateDataSafety(ctx context.Context, packageName PackageName, safetyLabels string) error {
	request := &androidpublisher.SafetyLabelsUpdateRequest{
		SafetyLabels: safetyLabels,
	}
	if _, err := p.service.Applications.DataSafety(packageName.String(), request).Context(ctx).Do(); err != nil {
		return fmt.Errorf("update data safety for %s: %w", packageName, err)
	}
	return nil
}

func (p GooglePublisher) CreateGrant(ctx context.Context, options GrantCreateOptions) (Grant, error) {
	apiGrant, err := p.service.Grants.Create(options.Parent(), grantToAPI(Grant{
		Name:        options.GrantName(),
		PackageName: options.PackageName,
		Permissions: options.Permissions,
	})).Context(ctx).Do()
	if err != nil {
		return Grant{}, fmt.Errorf("create grant for %s: %w", options.Parent(), err)
	}
	return grantFromAPI(apiGrant), nil
}

func (p GooglePublisher) PatchGrant(ctx context.Context, options GrantPatchOptions) (Grant, error) {
	apiGrant, err := p.service.Grants.Patch(options.Name.String(), grantToAPI(Grant{
		Name:        options.Name,
		Permissions: options.Permissions,
	})).UpdateMask("appLevelPermissions").Context(ctx).Do()
	if err != nil {
		return Grant{}, fmt.Errorf("patch grant %s: %w", options.Name, err)
	}
	return grantFromAPI(apiGrant), nil
}

func (p GooglePublisher) DeleteGrant(ctx context.Context, options GrantDeleteOptions) error {
	if err := p.service.Grants.Delete(options.Name.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete grant %s: %w", options.Name, err)
	}
	return nil
}

func (p GooglePublisher) GetOrder(ctx context.Context, options OrderGetOptions) (OrderGetResult, error) {
	apiOrder, err := p.service.Orders.Get(options.PackageName.String(), options.OrderID.String()).Context(ctx).Do()
	if err != nil {
		return OrderGetResult{}, fmt.Errorf("get order %s for %s: %w", options.OrderID, options.PackageName, err)
	}
	return OrderGetResult{
		PackageName: options.PackageName,
		OrderID:     options.OrderID,
		Order:       orderFromAPI(apiOrder),
	}, nil
}

func (p GooglePublisher) BatchGetOrders(ctx context.Context, options OrderBatchGetOptions) (OrderBatchGetResult, error) {
	response, err := p.service.Orders.Batchget(options.PackageName.String()).
		OrderIds(orderIDStrings(options.OrderIDs)...).
		Context(ctx).
		Do()
	if err != nil {
		return OrderBatchGetResult{}, fmt.Errorf("batch get orders for %s: %w", options.PackageName, err)
	}
	return orderBatchGetResultFromAPI(options, response), nil
}

func (p GooglePublisher) RefundOrder(ctx context.Context, options OrderRefundOptions) error {
	call := p.service.Orders.Refund(options.PackageName.String(), options.OrderID.String()).Context(ctx)
	if options.Revoke {
		call.Revoke(true)
	}
	if err := call.Do(); err != nil {
		return fmt.Errorf("refund order %s for %s: %w", options.OrderID, options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) ConvertRegionPrices(ctx context.Context, options RegionPriceConversionOptions) (RegionPriceConversionResult, error) {
	request := &androidpublisher.ConvertRegionPricesRequest{
		Price: &androidpublisher.Money{
			CurrencyCode: options.Currency.String(),
			Units:        options.Units,
			Nanos:        options.Nanos,
		},
	}
	response, err := p.service.Monetization.ConvertRegionPrices(options.PackageName.String(), request).
		Context(ctx).
		Do()
	if err != nil {
		return RegionPriceConversionResult{}, fmt.Errorf("convert region prices for %s: %w", options.PackageName, err)
	}
	return regionPriceConversionResultFromAPI(options, response), nil
}

func (p GooglePublisher) ListAppRecoveries(ctx context.Context, options AppRecoveryListOptions) (AppRecoveryListResult, error) {
	response, err := p.service.Apprecovery.List(options.PackageName.String()).
		VersionCode(options.VersionCode).
		Context(ctx).
		Do()
	if err != nil {
		return AppRecoveryListResult{}, fmt.Errorf("list app recoveries for %s version code %d: %w", options.PackageName, options.VersionCode, err)
	}
	return appRecoveryListResultFromAPI(options, response), nil
}

func (p GooglePublisher) CreateAppRecovery(ctx context.Context, options AppRecoveryCreateOptions) (AppRecoveryAction, error) {
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

func (p GooglePublisher) AddAppRecoveryTargeting(ctx context.Context, options AppRecoveryTargetingUpdateOptions) error {
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

func (p GooglePublisher) DeployAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
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

func (p GooglePublisher) CancelAppRecovery(ctx context.Context, options AppRecoveryMutationOptions) error {
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

func (p GooglePublisher) ListGeneratedAPKs(ctx context.Context, options GeneratedAPKListOptions) (GeneratedAPKListResult, error) {
	response, err := p.service.Generatedapks.List(options.PackageName.String(), options.VersionCode).
		Context(ctx).
		Do()
	if err != nil {
		return GeneratedAPKListResult{}, fmt.Errorf("list generated APKs for %s version code %d: %w", options.PackageName, options.VersionCode, err)
	}
	return generatedAPKListResultFromAPI(options, response), nil
}

func (p GooglePublisher) DownloadGeneratedAPK(ctx context.Context, options GeneratedAPKDownloadOptions) (GeneratedAPKDownloadResult, error) {
	if err := options.ValidateLive(); err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	if err := ValidateGeneratedAPKOutputPath(options.OutputPath, options.Force); err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	response, err := p.service.Generatedapks.Download(options.PackageName.String(), options.VersionCode, options.DownloadID.String()).
		Context(ctx).
		Download()
	if err != nil {
		return GeneratedAPKDownloadResult{}, fmt.Errorf("download generated APK %s for %s version code %d: %w", options.DownloadID, options.PackageName, options.VersionCode, err)
	}
	defer response.Body.Close()

	tempFile, err := createGeneratedAPKTempFile(options.OutputPath)
	if err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	tempPath := tempFile.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	bytesWritten, err := io.Copy(tempFile, response.Body)
	if err != nil {
		_ = tempFile.Close()
		return GeneratedAPKDownloadResult{}, fmt.Errorf("write temporary APK %s: %w", tempPath, err)
	}
	if err := tempFile.Close(); err != nil {
		return GeneratedAPKDownloadResult{}, fmt.Errorf("close temporary APK %s: %w", tempPath, err)
	}
	if err := publishGeneratedAPKTempFile(tempPath, options.OutputPath, options.Force); err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	committed = true
	plan, err := NewGeneratedAPKDownloadPlan(options)
	if err != nil {
		return GeneratedAPKDownloadResult{}, err
	}
	return GeneratedAPKDownloadResult{
		PackageName:  options.PackageName,
		VersionCode:  options.VersionCode,
		DownloadID:   options.DownloadID,
		OutputPath:   options.OutputPath,
		DryRun:       false,
		Downloaded:   true,
		BytesWritten: bytesWritten,
		Plan:         plan,
	}, nil
}

func createGeneratedAPKTempFile(outputPath string) (*os.File, error) {
	outputDirectory := filepath.Dir(outputPath)
	outputBase := filepath.Base(outputPath)
	file, err := os.CreateTemp(outputDirectory, "."+outputBase+".tmp-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary APK in %s: %w", outputDirectory, err)
	}
	return file, nil
}

func publishGeneratedAPKTempFile(tempPath string, outputPath string, force bool) error {
	if force {
		if err := os.Rename(tempPath, outputPath); err != nil {
			return fmt.Errorf("replace output APK %s: %w", outputPath, err)
		}
		return nil
	}
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create output APK %s: %w", outputPath, err)
	}
	outputCommitted := false
	defer func() {
		_ = outputFile.Close()
		if !outputCommitted {
			_ = os.Remove(outputPath)
		}
	}()
	tempFile, err := os.Open(tempPath)
	if err != nil {
		return fmt.Errorf("open temporary APK %s: %w", tempPath, err)
	}
	defer tempFile.Close()
	if _, err := io.Copy(outputFile, tempFile); err != nil {
		return fmt.Errorf("copy temporary APK %s to %s: %w", tempPath, outputPath, err)
	}
	if err := outputFile.Close(); err != nil {
		return fmt.Errorf("close output APK %s: %w", outputPath, err)
	}
	outputCommitted = true
	if err := os.Remove(tempPath); err != nil {
		return fmt.Errorf("remove temporary APK %s: %w", tempPath, err)
	}
	return nil
}

func (p GooglePublisher) ListSystemAPKVariants(ctx context.Context, options SystemAPKVariantListOptions) (SystemAPKVariantListResult, error) {
	response, err := p.service.Systemapks.Variants.List(options.PackageName.String(), options.VersionCode).
		Context(ctx).
		Do()
	if err != nil {
		return SystemAPKVariantListResult{}, fmt.Errorf("list system APK variants for %s version code %d: %w", options.PackageName, options.VersionCode, err)
	}
	return systemAPKVariantListResultFromAPI(options, response), nil
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

func (p GooglePublisher) CreateSubscriptionOffer(ctx context.Context, options SubscriptionOfferCreateOptions) (SubscriptionOffer, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOffer{}, err
	}
	offer, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Create(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		subscriptionOfferCreateToAPI(options),
	).
		OfferId(options.OfferID.String()).
		RegionsVersionVersion(options.RegionsVersion).
		Context(ctx).
		Do()
	if err != nil {
		return SubscriptionOffer{}, fmt.Errorf("create subscription offer %s for %s/%s/%s: %w", options.OfferID, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferFromAPI(offer), nil
}

func (p GooglePublisher) DeleteSubscriptionOffer(ctx context.Context, options SubscriptionOfferDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Monetization.Subscriptions.BasePlans.Offers.Delete(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		options.OfferID.String(),
	).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete subscription offer %s for %s/%s/%s: %w", options.OfferID, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return nil
}

func (p GooglePublisher) UpdateSubscriptionOfferState(ctx context.Context, options SubscriptionOfferStateUpdateOptions) (SubscriptionOffer, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOffer{}, err
	}
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	var (
		offer *androidpublisher.SubscriptionOffer
		err   error
	)
	switch options.Action {
	case SubscriptionOfferStateActionActivate:
		offer, err = p.service.Monetization.Subscriptions.BasePlans.Offers.Activate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			options.OfferID.String(),
			&androidpublisher.ActivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	case SubscriptionOfferStateActionDeactivate:
		offer, err = p.service.Monetization.Subscriptions.BasePlans.Offers.Deactivate(
			options.PackageName.String(),
			options.ProductID.String(),
			options.BasePlanID.String(),
			options.OfferID.String(),
			&androidpublisher.DeactivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				BasePlanId:       options.BasePlanID.String(),
				OfferId:          options.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		).Context(ctx).Do()
	default:
		return SubscriptionOffer{}, fmt.Errorf("unsupported subscription offer state action %q", options.Action)
	}
	if err != nil {
		return SubscriptionOffer{}, fmt.Errorf("%s subscription offer %s for %s/%s/%s: %w", options.Action, options.OfferID, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferFromAPI(offer), nil
}

func (p GooglePublisher) BatchUpdateSubscriptionOfferStates(ctx context.Context, options SubscriptionOfferBatchStateUpdateOptions) (SubscriptionOfferBatchStateUpdateResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchStateUpdateResult{}, err
	}
	request := &androidpublisher.BatchUpdateSubscriptionOfferStatesRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferStateRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, subscriptionOfferStateRequestToAPI(options, item))
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdateStates(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchStateUpdateResult{}, fmt.Errorf("batch %s subscription offers for %s/%s/%s: %w", options.Action, options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	offers := subscriptionOffersFromBatchStateUpdateResponse(options, response)
	return SubscriptionOfferBatchStateUpdateResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferBatchMutationRequest(nil), options.Requests...),
		Action:      options.Action,
		Applied:     true,
		Offers:      offers,
	}, nil
}

func (p GooglePublisher) BatchPatchSubscriptionOfferAvailability(ctx context.Context, options SubscriptionOfferBatchPatchAvailabilityOptions) (SubscriptionOfferBatchPatchAvailabilityResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchAvailabilityResult{}, err
	}
	requestsByOffer := subscriptionOfferAvailabilityPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchAvailabilityResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before availability patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		mergedRegionalConfigs := mergeSubscriptionOfferAvailabilityPatches(subscriptionOfferRegionalConfigsFromAPI(current.RegionalConfigs), offerPatch.Requests)
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName:     options.PackageName.String(),
				ProductId:       offerPatch.ProductID.String(),
				BasePlanId:      offerPatch.BasePlanID.String(),
				OfferId:         offerPatch.OfferID.String(),
				RegionalConfigs: subscriptionOfferRegionalConfigsToAPI(mergedRegionalConfigs),
			},
			UpdateMask:       subscriptionOfferAvailabilityUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchAvailabilityResult{}, fmt.Errorf("batch patch subscription offer availability for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchAvailabilityResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferAvailabilityPatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdateResponse(options, response),
	}, nil
}

func (p GooglePublisher) BatchPatchSubscriptionOfferPhaseRelativeDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions) (SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, err
	}
	requestsByOffer := subscriptionOfferPhaseRelativeDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before phase relative discount patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		phases := subscriptionOfferPhasesFromAPI(current.Phases)
		mergedPhases := mergeSubscriptionOfferPhaseRelativeDiscountPatches(phases, offerPatch.Requests)
		if mergedPhases == nil {
			return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, fmt.Errorf("subscription offer phase relative discount patch for %s/%s/%s references a phase or region that is not already configured on the offer", offerPatch.ProductID, offerPatch.BasePlanID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName: options.PackageName.String(),
				ProductId:   offerPatch.ProductID.String(),
				BasePlanId:  offerPatch.BasePlanID.String(),
				OfferId:     offerPatch.OfferID.String(),
				Phases:      subscriptionOfferPhasesToAPI(mergedPhases),
			},
			UpdateMask:       subscriptionOfferPhasesUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{}, fmt.Errorf("batch patch subscription offer phase relative discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchPhaseRelativeDiscountsResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferPhaseRelativeDiscountPatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdatePhaseRelativeDiscountResponse(options, response),
	}, nil
}

func (p GooglePublisher) BatchPatchSubscriptionOfferPhaseAbsoluteDiscounts(ctx context.Context, options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions) (SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, err
	}
	requestsByOffer := subscriptionOfferPhaseAbsoluteDiscountPatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before phase absolute discount patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		phases := subscriptionOfferPhasesFromAPI(current.Phases)
		mergedPhases := mergeSubscriptionOfferPhaseAbsoluteDiscountPatches(phases, offerPatch.Requests)
		if mergedPhases == nil {
			return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, fmt.Errorf("subscription offer phase absolute discount patch for %s/%s/%s references a phase or region that is not already configured on the offer", offerPatch.ProductID, offerPatch.BasePlanID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName: options.PackageName.String(),
				ProductId:   offerPatch.ProductID.String(),
				BasePlanId:  offerPatch.BasePlanID.String(),
				OfferId:     offerPatch.OfferID.String(),
				Phases:      subscriptionOfferPhasesToAPI(mergedPhases),
			},
			UpdateMask:       subscriptionOfferPhasesUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{}, fmt.Errorf("batch patch subscription offer phase absolute discounts for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferPhaseAbsoluteDiscountPatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdatePhaseAbsoluteDiscountResponse(options, response),
	}, nil
}

func (p GooglePublisher) BatchPatchSubscriptionOfferPhasePrices(ctx context.Context, options SubscriptionOfferBatchPatchPhasePricesOptions) (SubscriptionOfferBatchPatchPhasePricesResult, error) {
	if err := options.ValidateLive(); err != nil {
		return SubscriptionOfferBatchPatchPhasePricesResult{}, err
	}
	requestsByOffer := subscriptionOfferPhasePricePatchRequestsByOffer(options.Requests)
	request := &androidpublisher.BatchUpdateSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.UpdateSubscriptionOfferRequest, 0, len(requestsByOffer)),
	}
	for _, offerPatch := range requestsByOffer {
		current, err := p.service.Monetization.Subscriptions.BasePlans.Offers.Get(
			options.PackageName.String(),
			offerPatch.ProductID.String(),
			offerPatch.BasePlanID.String(),
			offerPatch.OfferID.String(),
		).Context(ctx).Do()
		if err != nil {
			return SubscriptionOfferBatchPatchPhasePricesResult{}, fmt.Errorf("get subscription offer %s for %s/%s/%s before phase price patch: %w", offerPatch.OfferID, options.PackageName, offerPatch.ProductID, offerPatch.BasePlanID, err)
		}
		phases := subscriptionOfferPhasesFromAPI(current.Phases)
		mergedPhases := mergeSubscriptionOfferPhasePricePatches(phases, offerPatch.Requests)
		if mergedPhases == nil {
			return SubscriptionOfferBatchPatchPhasePricesResult{}, fmt.Errorf("subscription offer phase price patch for %s/%s/%s references a phase or region that is not already configured on the offer", offerPatch.ProductID, offerPatch.BasePlanID, offerPatch.OfferID)
		}
		request.Requests = append(request.Requests, &androidpublisher.UpdateSubscriptionOfferRequest{
			SubscriptionOffer: &androidpublisher.SubscriptionOffer{
				PackageName: options.PackageName.String(),
				ProductId:   offerPatch.ProductID.String(),
				BasePlanId:  offerPatch.BasePlanID.String(),
				OfferId:     offerPatch.OfferID.String(),
				Phases:      subscriptionOfferPhasesToAPI(mergedPhases),
			},
			UpdateMask:       subscriptionOfferPhasesUpdateMask,
			RegionsVersion:   &androidpublisher.RegionsVersion{Version: options.RegionsVersion},
			LatencyTolerance: productUpdateLatencyToleranceToAPI(options.LatencyTolerance),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchPatchPhasePricesResult{}, fmt.Errorf("batch patch subscription offer phase prices for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return SubscriptionOfferBatchPatchPhasePricesResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Requests:    append([]SubscriptionOfferPhasePricePatchRequest(nil), options.Requests...),
		Applied:     true,
		Offers:      subscriptionOffersFromBatchUpdatePhasePriceResponse(options, response),
	}, nil
}

type subscriptionOfferAvailabilityPatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferAvailabilityPatchRequest
}

type subscriptionOfferPhaseRelativeDiscountPatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferPhaseRelativeDiscountPatchRequest
}

type subscriptionOfferPhaseAbsoluteDiscountPatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest
}

type subscriptionOfferPhasePricePatchOffer struct {
	ProductID  SubscriptionProductID
	BasePlanID SubscriptionBasePlanID
	OfferID    SubscriptionOfferID
	Requests   []SubscriptionOfferPhasePricePatchRequest
}

func subscriptionOfferAvailabilityPatchRequestsByOffer(requests []SubscriptionOfferAvailabilityPatchRequest) []subscriptionOfferAvailabilityPatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferAvailabilityPatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferAvailabilityPatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOfferPhaseRelativeDiscountPatchRequestsByOffer(requests []SubscriptionOfferPhaseRelativeDiscountPatchRequest) []subscriptionOfferPhaseRelativeDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferPhaseRelativeDiscountPatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferPhaseRelativeDiscountPatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOfferPhaseAbsoluteDiscountPatchRequestsByOffer(requests []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) []subscriptionOfferPhaseAbsoluteDiscountPatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferPhaseAbsoluteDiscountPatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferPhaseAbsoluteDiscountPatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOfferPhasePricePatchRequestsByOffer(requests []SubscriptionOfferPhasePricePatchRequest) []subscriptionOfferPhasePricePatchOffer {
	byOffer := map[string]int{}
	offers := make([]subscriptionOfferPhasePricePatchOffer, 0)
	for _, request := range requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		index, ok := byOffer[key]
		if !ok {
			byOffer[key] = len(offers)
			offers = append(offers, subscriptionOfferPhasePricePatchOffer{
				ProductID:  request.ProductID,
				BasePlanID: request.BasePlanID,
				OfferID:    request.OfferID,
			})
			index = len(offers) - 1
		}
		offers[index].Requests = append(offers[index].Requests, request)
	}
	return offers
}

func subscriptionOffersFromBatchUpdateResponse(options SubscriptionOfferBatchPatchAvailabilityOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchUpdatePhaseRelativeDiscountResponse(options SubscriptionOfferBatchPatchPhaseRelativeDiscountsOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchUpdatePhaseAbsoluteDiscountResponse(options SubscriptionOfferBatchPatchPhaseAbsoluteDiscountsOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchUpdatePhasePriceResponse(options SubscriptionOfferBatchPatchPhasePricesOptions, response *androidpublisher.BatchUpdateSubscriptionOffersResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	mutationRequests := make([]SubscriptionOfferBatchMutationRequest, 0, len(options.Requests))
	for _, request := range options.Requests {
		mutationRequests = append(mutationRequests, SubscriptionOfferBatchMutationRequest{
			ProductID:  request.ProductID,
			BasePlanID: request.BasePlanID,
			OfferID:    request.OfferID,
		})
	}
	for _, request := range deduplicateSubscriptionOfferMutationRequests(mutationRequests) {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func subscriptionOffersFromBatchStateUpdateResponse(options SubscriptionOfferBatchStateUpdateOptions, response *androidpublisher.BatchUpdateSubscriptionOfferStatesResponse) []SubscriptionOffer {
	if response == nil {
		return []SubscriptionOffer{}
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	offers := make([]SubscriptionOffer, 0, len(response.SubscriptionOffers))
	for _, request := range options.Requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			offers = append(offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	return append(offers, extras...)
}

func (p GooglePublisher) BatchGetSubscriptionOffers(ctx context.Context, options SubscriptionOfferBatchGetOptions) (SubscriptionOfferBatchGetResult, error) {
	request := &androidpublisher.BatchGetSubscriptionOffersRequest{
		Requests: make([]*androidpublisher.GetSubscriptionOfferRequest, 0, len(options.Requests)),
	}
	for _, item := range options.Requests {
		request.Requests = append(request.Requests, &androidpublisher.GetSubscriptionOfferRequest{
			PackageName: options.PackageName.String(),
			ProductId:   item.ProductID.String(),
			BasePlanId:  item.BasePlanID.String(),
			OfferId:     item.OfferID.String(),
		})
	}
	response, err := p.service.Monetization.Subscriptions.BasePlans.Offers.BatchGet(
		options.PackageName.String(),
		options.ProductID.String(),
		options.BasePlanID.String(),
		request,
	).Context(ctx).Do()
	if err != nil {
		return SubscriptionOfferBatchGetResult{}, fmt.Errorf("batch get subscription offers for %s/%s/%s: %w", options.PackageName, options.ProductID, options.BasePlanID, err)
	}
	return subscriptionOfferBatchGetResultFromAPI(options, response), nil
}

func (p GooglePublisher) GetProductPurchase(ctx context.Context, options ProductPurchaseOptions) (ProductPurchase, error) {
	purchase, err := p.service.Purchases.Productsv2.Getproductpurchasev2(options.PackageName.String(), options.Token.String()).Context(ctx).Do()
	if err != nil {
		return ProductPurchase{}, fmt.Errorf("get product purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return productPurchaseFromAPI(options, purchase), nil
}

func subscriptionOfferStateRequestToAPI(options SubscriptionOfferBatchStateUpdateOptions, item SubscriptionOfferBatchMutationRequest) *androidpublisher.UpdateSubscriptionOfferStateRequest {
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	switch options.Action {
	case SubscriptionOfferStateActionActivate:
		return &androidpublisher.UpdateSubscriptionOfferStateRequest{
			ActivateSubscriptionOfferRequest: &androidpublisher.ActivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				BasePlanId:       item.BasePlanID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case SubscriptionOfferStateActionDeactivate:
		return &androidpublisher.UpdateSubscriptionOfferStateRequest{
			DeactivateSubscriptionOfferRequest: &androidpublisher.DeactivateSubscriptionOfferRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        item.ProductID.String(),
				BasePlanId:       item.BasePlanID.String(),
				OfferId:          item.OfferID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	default:
		return &androidpublisher.UpdateSubscriptionOfferStateRequest{}
	}
}

func subscriptionOfferBatchGetResultFromAPI(options SubscriptionOfferBatchGetOptions, response *androidpublisher.BatchGetSubscriptionOffersResponse) SubscriptionOfferBatchGetResult {
	result := SubscriptionOfferBatchGetResult{
		PackageName: options.PackageName,
		ProductID:   options.ProductID,
		BasePlanID:  options.BasePlanID,
		Offers:      []SubscriptionOffer{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	byKey := make(map[string]SubscriptionOffer, len(response.SubscriptionOffers))
	extras := make([]SubscriptionOffer, 0)
	for _, apiOffer := range response.SubscriptionOffers {
		offer := subscriptionOfferFromAPI(apiOffer)
		key := subscriptionOfferKey(offer.ProductID, offer.BasePlanID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	for _, request := range options.Requests {
		key := subscriptionOfferKey(request.ProductID, request.BasePlanID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			result.Offers = append(result.Offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return subscriptionOfferKey(extras[i].ProductID, extras[i].BasePlanID, extras[i].OfferID) < subscriptionOfferKey(extras[j].ProductID, extras[j].BasePlanID, extras[j].OfferID)
	})
	result.Offers = append(result.Offers, extras...)
	return result
}

func subscriptionOfferKey(productID SubscriptionProductID, basePlanID SubscriptionBasePlanID, offerID SubscriptionOfferID) string {
	return productID.String() + "/" + basePlanID.String() + "/" + offerID.String()
}

func (p GooglePublisher) AcknowledgeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	request := &androidpublisher.ProductPurchasesAcknowledgeRequest{}
	if options.DeveloperPayload != "" {
		request.DeveloperPayload = options.DeveloperPayload
	}
	if err := p.service.Purchases.Products.Acknowledge(options.PackageName.String(), options.ProductID.String(), options.Token.String(), request).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("acknowledge product purchase %s for %s/%s: %w", options.Token, options.PackageName, options.ProductID, err)
	}
	return nil
}

func (p GooglePublisher) ConsumeProductPurchase(ctx context.Context, options ProductPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Purchases.Products.Consume(options.PackageName.String(), options.ProductID.String(), options.Token.String()).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("consume product purchase %s for %s/%s: %w", options.Token, options.PackageName, options.ProductID, err)
	}
	return nil
}

func (p GooglePublisher) GetSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseOptions) (SubscriptionPurchase, error) {
	purchase, err := p.service.Purchases.Subscriptionsv2.Get(options.PackageName.String(), options.Token.String()).Context(ctx).Do()
	if err != nil {
		return SubscriptionPurchase{}, fmt.Errorf("get subscription purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return subscriptionPurchaseFromAPI(options.PackageName, options.Token, purchase), nil
}

func (p GooglePublisher) AcknowledgeSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if options.Action != SubscriptionPurchaseMutationActionAcknowledge {
		return fmt.Errorf("acknowledge subscription purchase requires action %q", SubscriptionPurchaseMutationActionAcknowledge)
	}
	request := &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{}
	if options.DeveloperPayload != "" {
		request.DeveloperPayload = options.DeveloperPayload
	}
	if err := p.service.Purchases.Subscriptions.Acknowledge(options.PackageName.String(), options.SubscriptionID.String(), options.Token.String(), request).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("acknowledge subscription purchase %s for %s/%s: %w", options.Token, options.PackageName, options.SubscriptionID, err)
	}
	return nil
}

func (p GooglePublisher) CancelSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseMutationOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if options.Action != SubscriptionPurchaseMutationActionCancel {
		return fmt.Errorf("cancel subscription purchase requires action %q", SubscriptionPurchaseMutationActionCancel)
	}
	requestBody := subscriptionCancelRequestToAPI(options)
	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("encode subscription cancel request: %w", err)
	}
	requestURL := googleapi.ResolveRelative(p.basePath, "androidpublisher/v3/applications/{packageName}/purchases/subscriptionsv2/tokens/{token}:cancel")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create subscription cancel request: %w", err)
	}
	googleapi.Expand(req.URL, map[string]string{
		"packageName": options.PackageName.String(),
		"token":       options.Token.String(),
	})
	query := req.URL.Query()
	query.Set("alt", "json")
	query.Set("prettyPrint", "false")
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", "application/json")
	if err := p.doNoContent(req); err != nil {
		return fmt.Errorf("cancel subscription purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) RevokeSubscriptionPurchase(ctx context.Context, options SubscriptionPurchaseRevokeOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if _, err := p.service.Purchases.Subscriptionsv2.Revoke(options.PackageName.String(), options.Token.String(), subscriptionRevokeRequestToAPI(options)).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("revoke subscription purchase %s for %s: %w", options.Token, options.PackageName, err)
	}
	return nil
}

func (p GooglePublisher) ListVoidedPurchases(ctx context.Context, options VoidedPurchaseListOptions) (VoidedPurchaseListResult, error) {
	call := p.service.Purchases.Voidedpurchases.List(options.PackageName.String()).Context(ctx)
	if options.MaxResults > 0 {
		call.MaxResults(options.MaxResults)
	}
	if options.StartIndex > 0 {
		call.StartIndex(options.StartIndex)
	}
	if options.Token != "" {
		call.Token(options.Token)
	}
	if options.StartTimeMillis > 0 {
		call.StartTime(options.StartTimeMillis)
	}
	if options.EndTimeMillis > 0 {
		call.EndTime(options.EndTimeMillis)
	}
	if options.Type != VoidedPurchaseTypeProductsOnly {
		call.Type(int64(options.Type))
	}
	if options.IncludeQuantityBasedPartialRefund {
		call.IncludeQuantityBasedPartialRefund(true)
	}
	response, err := call.Do()
	if err != nil {
		return VoidedPurchaseListResult{}, fmt.Errorf("list voided purchases for %s: %w", options.PackageName, err)
	}
	return voidedPurchaseListResultFromAPI(options, response), nil
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

func (p GooglePublisher) PromoteTrackRelease(ctx context.Context, packageName PackageName, editID string, sourceTrack TrackName, targetTrack TrackName, versionCode int64, status ReleaseStatus, userFraction *float64, releaseNotes []ReleaseNote) (TrackRelease, error) {
	source, err := p.service.Edits.Tracks.Get(packageName.String(), editID, sourceTrack.String()).Context(ctx).Do()
	if err != nil {
		return TrackRelease{}, fmt.Errorf("get %s track for %s: %w", sourceTrack, packageName, err)
	}
	apiRelease, err := selectReleaseByVersionCode(source, versionCode)
	if err != nil {
		return TrackRelease{}, fmt.Errorf("find version code %d on %s track for %s: %w", versionCode, sourceTrack, packageName, err)
	}
	setReleaseStatus(apiRelease, status, userFraction)
	if len(releaseNotes) > 0 {
		apiRelease.ReleaseNotes = releaseNotesToAPI(releaseNotes)
	}

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
	apiRelease.ReleaseNotes = releaseNotesToAPI(release.ReleaseNotes)
	return apiRelease
}

func releaseNotesToAPI(notes []ReleaseNote) []*androidpublisher.LocalizedText {
	apiNotes := make([]*androidpublisher.LocalizedText, 0, len(notes))
	for _, note := range notes {
		apiNotes = append(apiNotes, &androidpublisher.LocalizedText{
			Language: note.Language.String(),
			Text:     note.Text,
		})
	}
	return apiNotes
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
		ReleaseNotes: releaseNotesFromAPI(apiRelease.ReleaseNotes),
	}
}

func releaseNotesFromAPI(apiNotes []*androidpublisher.LocalizedText) []ReleaseNote {
	notes := make([]ReleaseNote, 0, len(apiNotes))
	for _, apiNote := range apiNotes {
		if apiNote == nil {
			continue
		}
		notes = append(notes, ReleaseNote{
			Language: ListingLanguage(apiNote.Language),
			Text:     apiNote.Text,
		})
	}
	return notes
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

func imageFromAPI(apiImage *androidpublisher.Image) StoreImage {
	if apiImage == nil {
		return StoreImage{}
	}
	return StoreImage{
		ID:     apiImage.Id,
		URL:    apiImage.Url,
		SHA1:   apiImage.Sha1,
		SHA256: apiImage.Sha256,
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

func inAppProductBatchGetResultFromAPI(options InAppProductBatchGetOptions, response *androidpublisher.InappproductsBatchGetResponse) InAppProductBatchGetResult {
	result := InAppProductBatchGetResult{
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

func inAppProductCreateToAPI(options InAppProductCreateOptions) *androidpublisher.InAppProduct {
	return &androidpublisher.InAppProduct{
		PackageName:     options.PackageName.String(),
		Sku:             options.SKU.String(),
		Status:          options.Status.String(),
		PurchaseType:    ProductPurchaseTypeManagedUser.String(),
		DefaultLanguage: options.DefaultLanguage.String(),
		DefaultPrice:    productPriceToAPI(options.DefaultPrice),
		Listings: map[string]androidpublisher.InAppProductListing{
			options.DefaultLanguage.String(): {
				Title:       options.Listing.Title,
				Description: options.Listing.Description,
				Benefits:    options.Listing.Benefits,
			},
		},
	}
}

func inAppProductPatchToAPI(options InAppProductPatchOptions) *androidpublisher.InAppProduct {
	product := &androidpublisher.InAppProduct{
		PackageName: options.PackageName.String(),
		Sku:         options.SKU.String(),
	}
	if options.Status != "" {
		product.Status = options.Status.String()
	}
	if options.DefaultLanguage != "" {
		product.DefaultLanguage = options.DefaultLanguage.String()
	}
	if options.DefaultPrice != nil {
		product.DefaultPrice = productPriceToAPI(*options.DefaultPrice)
	}
	if len(options.RegionalPrices) > 0 {
		product.Prices = productPricesToAPI(regionalProductPricesToMap(options.RegionalPrices))
	}
	if options.TaxComplianceSettings != nil {
		product.ManagedProductTaxesAndComplianceSettings = managedProductTaxComplianceSettingsToAPI(options.TaxComplianceSettings)
	}
	if options.Listing != nil {
		product.Listings = map[string]androidpublisher.InAppProductListing{
			options.ListingLanguage.String(): {
				Title:       options.Listing.Title,
				Description: options.Listing.Description,
				Benefits:    options.Listing.Benefits,
			},
		}
	}
	return product
}

func productPriceFromAPI(apiPrice *androidpublisher.Price) *ProductPrice {
	if apiPrice == nil {
		return nil
	}
	return &ProductPrice{Currency: apiPrice.Currency, PriceMicros: apiPrice.PriceMicros}
}

func productPriceToAPI(price ProductPrice) *androidpublisher.Price {
	return &androidpublisher.Price{Currency: price.Currency, PriceMicros: price.PriceMicros}
}

func productPricesToAPI(prices map[string]ProductPrice) map[string]androidpublisher.Price {
	if len(prices) == 0 {
		return nil
	}
	apiPrices := make(map[string]androidpublisher.Price, len(prices))
	for region, price := range prices {
		apiPrices[region] = androidpublisher.Price{Currency: price.Currency, PriceMicros: price.PriceMicros}
	}
	return apiPrices
}

func managedProductTaxComplianceSettingsToAPI(settings *ProductTaxComplianceSettings) *androidpublisher.ManagedProductTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	apiSettings := &androidpublisher.ManagedProductTaxAndComplianceSettings{
		EeaWithdrawalRightType: settings.EEAWithdrawalRightType,
	}
	if settings.IsTokenizedDigitalAsset != nil {
		apiSettings.IsTokenizedDigitalAsset = *settings.IsTokenizedDigitalAsset
		apiSettings.ForceSendFields = append(apiSettings.ForceSendFields, "IsTokenizedDigitalAsset")
	}
	if len(settings.TaxRateInfoByRegionCode) > 0 {
		apiSettings.TaxRateInfoByRegionCode = regionalTaxRateInfoToAPI(settings.TaxRateInfoByRegionCode)
	}
	return apiSettings
}

func subscriptionTaxComplianceSettingsToAPI(settings *ProductTaxComplianceSettings) *androidpublisher.SubscriptionTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	apiSettings := &androidpublisher.SubscriptionTaxAndComplianceSettings{
		EeaWithdrawalRightType: settings.EEAWithdrawalRightType,
	}
	if settings.IsTokenizedDigitalAsset != nil {
		apiSettings.IsTokenizedDigitalAsset = *settings.IsTokenizedDigitalAsset
		apiSettings.ForceSendFields = append(apiSettings.ForceSendFields, "IsTokenizedDigitalAsset")
	}
	if len(settings.TaxRateInfoByRegionCode) > 0 {
		apiSettings.TaxRateInfoByRegionCode = regionalTaxRateInfoToAPI(settings.TaxRateInfoByRegionCode)
	}
	return apiSettings
}

func regionalTaxRateInfoToAPI(taxRateInfo map[string]RegionalTaxRateInfo) map[string]androidpublisher.RegionalTaxRateInfo {
	if len(taxRateInfo) == 0 {
		return nil
	}
	apiTaxRateInfo := make(map[string]androidpublisher.RegionalTaxRateInfo, len(taxRateInfo))
	for region, info := range taxRateInfo {
		apiTaxRateInfo[region] = androidpublisher.RegionalTaxRateInfo{
			EligibleForStreamingServiceTaxRate: info.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   info.StreamingTaxType,
			TaxTier:                            info.TaxTier,
		}
	}
	return apiTaxRateInfo
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

type rawOneTimeProductListResponse struct {
	NextPageToken   string              `json:"nextPageToken,omitempty"`
	OneTimeProducts []rawOneTimeProduct `json:"oneTimeProducts,omitempty"`
}

type rawOneTimeProduct struct {
	PackageName                string                                  `json:"packageName,omitempty"`
	ProductID                  string                                  `json:"productId,omitempty"`
	Listings                   []rawOneTimeProductListing              `json:"listings,omitempty"`
	OfferTags                  []rawOfferTag                           `json:"offerTags,omitempty"`
	PurchaseOptions            []rawOneTimeProductPurchaseOption       `json:"purchaseOptions,omitempty"`
	RegionsVersion             *rawRegionsVersion                      `json:"regionsVersion,omitempty"`
	RestrictedPaymentCountries *rawRestrictedPaymentCountries          `json:"restrictedPaymentCountries,omitempty"`
	TaxAndComplianceSettings   *rawOneTimeProductTaxComplianceSettings `json:"taxAndComplianceSettings,omitempty"`
}

type rawOneTimeProductListing struct {
	LanguageCode string `json:"languageCode,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
}

type rawOfferTag struct {
	Tag string `json:"tag,omitempty"`
}

type rawOneTimeProductPurchaseOption struct {
	PurchaseOptionID                      string                                  `json:"purchaseOptionId,omitempty"`
	State                                 string                                  `json:"state,omitempty"`
	BuyOption                             *rawOneTimeProductBuyOption             `json:"buyOption,omitempty"`
	RentOption                            *rawOneTimeProductRentOption            `json:"rentOption,omitempty"`
	OfferTags                             []rawOfferTag                           `json:"offerTags,omitempty"`
	RegionalPricingAndAvailabilityConfigs []rawOneTimeProductRegionalConfig       `json:"regionalPricingAndAvailabilityConfigs,omitempty"`
	NewRegionsConfig                      *rawOneTimeProductNewRegionsConfig      `json:"newRegionsConfig,omitempty"`
	TaxAndComplianceSettings              *rawPurchaseOptionTaxComplianceSettings `json:"taxAndComplianceSettings,omitempty"`
}

type rawOneTimeProductBuyOption struct {
	LegacyCompatible     bool `json:"legacyCompatible,omitempty"`
	MultiQuantityEnabled bool `json:"multiQuantityEnabled,omitempty"`
}

type rawOneTimeProductRentOption struct {
	RentalPeriod     string `json:"rentalPeriod,omitempty"`
	ExpirationPeriod string `json:"expirationPeriod,omitempty"`
}

type rawOneTimeProductRegionalConfig struct {
	RegionCode   string    `json:"regionCode,omitempty"`
	Availability string    `json:"availability,omitempty"`
	Price        *rawMoney `json:"price,omitempty"`
}

type rawOneTimeProductNewRegionsConfig struct {
	Availability string    `json:"availability,omitempty"`
	USDPrice     *rawMoney `json:"usdPrice,omitempty"`
	EURPrice     *rawMoney `json:"eurPrice,omitempty"`
}

type rawPurchaseOptionTaxComplianceSettings struct {
	WithdrawalRightType string `json:"withdrawalRightType,omitempty"`
}

type rawOneTimeProductTaxComplianceSettings struct {
	IsTokenizedDigitalAsset       bool                          `json:"isTokenizedDigitalAsset,omitempty"`
	ProductTaxCategoryCode        string                        `json:"productTaxCategoryCode,omitempty"`
	RegionalProductAgeRatingInfos []rawRegionalProductAgeRating `json:"regionalProductAgeRatingInfos,omitempty"`
	RegionalTaxConfigs            []rawRegionalTaxConfig        `json:"regionalTaxConfigs,omitempty"`
}

type rawRegionalProductAgeRating struct {
	RegionCode           string `json:"regionCode,omitempty"`
	ProductAgeRatingTier string `json:"productAgeRatingTier,omitempty"`
}

type rawRegionalTaxConfig struct {
	RegionCode                         string `json:"regionCode,omitempty"`
	EligibleForStreamingServiceTaxRate bool   `json:"eligibleForStreamingServiceTaxRate,omitempty"`
	StreamingTaxType                   string `json:"streamingTaxType,omitempty"`
	TaxTier                            string `json:"taxTier,omitempty"`
}

type rawRestrictedPaymentCountries struct {
	RegionCodes []string `json:"regionCodes,omitempty"`
}

type rawRegionsVersion struct {
	Version string `json:"version,omitempty"`
}

type rawMoney struct {
	CurrencyCode string        `json:"currencyCode,omitempty"`
	Units        flexibleInt64 `json:"units,omitempty"`
	Nanos        int64         `json:"nanos,omitempty"`
}

type flexibleInt64 int64

func (i *flexibleInt64) UnmarshalJSON(data []byte) error {
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		parsed, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return err
		}
		*i = flexibleInt64(parsed)
		return nil
	}
	var numberValue int64
	if err := json.Unmarshal(data, &numberValue); err != nil {
		return err
	}
	*i = flexibleInt64(numberValue)
	return nil
}

func oneTimeProductListResultFromAPI(options OneTimeProductListOptions, response rawOneTimeProductListResponse) OneTimeProductListResult {
	result := OneTimeProductListResult{
		PackageName: options.PackageName,
		Products:    []OneTimeProduct{},
		Options:     options,
	}
	result.NextPageToken = response.NextPageToken
	for _, apiProduct := range response.OneTimeProducts {
		result.Products = append(result.Products, oneTimeProductFromAPI(apiProduct))
	}
	return result
}

func oneTimeProductBatchGetResultFromAPI(options OneTimeProductBatchGetOptions, response *androidpublisher.BatchGetOneTimeProductsResponse) (OneTimeProductBatchGetResult, error) {
	result := OneTimeProductBatchGetResult{
		PackageName: options.PackageName,
		Products:    []OneTimeProduct{},
		Options:     options,
	}
	if response == nil {
		return result, nil
	}
	for _, apiProduct := range response.OneTimeProducts {
		product, err := oneTimeProductFromGeneratedAPI(apiProduct)
		if err != nil {
			return OneTimeProductBatchGetResult{}, err
		}
		result.Products = append(result.Products, product)
	}
	return result, nil
}

func oneTimeProductFromAPI(apiProduct rawOneTimeProduct) OneTimeProduct {
	if apiProduct.PackageName == "" && apiProduct.ProductID == "" {
		return OneTimeProduct{Listings: []OneTimeProductListing{}, PurchaseOptions: []OneTimeProductPurchaseOption{}}
	}
	return OneTimeProduct{
		PackageName:              PackageName(apiProduct.PackageName),
		ProductID:                OneTimeProductID(apiProduct.ProductID),
		Listings:                 oneTimeProductListingsFromAPI(apiProduct.Listings),
		OfferTags:                rawOfferTagsFromAPI(apiProduct.OfferTags),
		PurchaseOptions:          oneTimeProductPurchaseOptionsFromAPI(apiProduct.PurchaseOptions),
		RegionsVersion:           regionsVersionFromAPI(apiProduct.RegionsVersion),
		RestrictedCountries:      rawRestrictedCountriesFromAPI(apiProduct.RestrictedPaymentCountries),
		TaxAndComplianceSettings: oneTimeProductTaxComplianceSettingsFromAPI(apiProduct.TaxAndComplianceSettings),
	}
}

func oneTimeProductFromGeneratedAPI(apiProduct *androidpublisher.OneTimeProduct) (OneTimeProduct, error) {
	if apiProduct == nil {
		return OneTimeProduct{}, fmt.Errorf("one-time product response is empty")
	}
	content, err := json.Marshal(apiProduct)
	if err != nil {
		return OneTimeProduct{}, fmt.Errorf("marshal one-time product response: %w", err)
	}
	var rawProduct rawOneTimeProduct
	if err := json.Unmarshal(content, &rawProduct); err != nil {
		return OneTimeProduct{}, fmt.Errorf("decode one-time product response: %w", err)
	}
	return oneTimeProductFromAPI(rawProduct), nil
}

func oneTimeProductToAPI(product OneTimeProduct) *androidpublisher.OneTimeProduct {
	return &androidpublisher.OneTimeProduct{
		PackageName:                product.PackageName.String(),
		ProductId:                  product.ProductID.String(),
		Listings:                   oneTimeProductListingsToAPI(product.Listings),
		OfferTags:                  offerTagsToAPI(product.OfferTags),
		PurchaseOptions:            oneTimeProductPurchaseOptionsToAPI(product.PurchaseOptions),
		RestrictedPaymentCountries: restrictedCountriesToAPI(product.RestrictedCountries),
		TaxAndComplianceSettings:   oneTimeProductTaxComplianceSettingsToAPI(product.TaxAndComplianceSettings),
	}
}

func purchaseOptionStateRequestToAPI(options PurchaseOptionStateUpdateOptions) *androidpublisher.UpdatePurchaseOptionStateRequest {
	latencyTolerance := productUpdateLatencyToleranceToAPI(options.LatencyTolerance)
	switch options.Action {
	case PurchaseOptionStateActionActivate:
		return &androidpublisher.UpdatePurchaseOptionStateRequest{
			ActivatePurchaseOptionRequest: &androidpublisher.ActivatePurchaseOptionRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	case PurchaseOptionStateActionDeactivate:
		return &androidpublisher.UpdatePurchaseOptionStateRequest{
			DeactivatePurchaseOptionRequest: &androidpublisher.DeactivatePurchaseOptionRequest{
				PackageName:      options.PackageName.String(),
				ProductId:        options.ProductID.String(),
				PurchaseOptionId: options.PurchaseOptionID.String(),
				LatencyTolerance: latencyTolerance,
			},
		}
	default:
		return &androidpublisher.UpdatePurchaseOptionStateRequest{}
	}
}

func productUpdateLatencyToleranceToAPI(latencyTolerance ProductUpdateLatencyTolerance) string {
	switch latencyTolerance {
	case ProductUpdateLatencyToleranceTolerant:
		return "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_TOLERANT"
	default:
		return "PRODUCT_UPDATE_LATENCY_TOLERANCE_LATENCY_SENSITIVE"
	}
}

func oneTimeProductListingsFromAPI(apiListings []rawOneTimeProductListing) []OneTimeProductListing {
	listings := make([]OneTimeProductListing, 0, len(apiListings))
	for _, apiListing := range apiListings {
		listings = append(listings, OneTimeProductListing{
			LanguageCode: apiListing.LanguageCode,
			Title:        apiListing.Title,
			Description:  apiListing.Description,
		})
	}
	return listings
}

func oneTimeProductListingsToAPI(listings []OneTimeProductListing) []*androidpublisher.OneTimeProductListing {
	apiListings := make([]*androidpublisher.OneTimeProductListing, 0, len(listings))
	for _, listing := range listings {
		apiListings = append(apiListings, oneTimeProductListingToAPI(listing))
	}
	return apiListings
}

func oneTimeProductListingToAPI(listing OneTimeProductListing) *androidpublisher.OneTimeProductListing {
	return &androidpublisher.OneTimeProductListing{
		LanguageCode: listing.LanguageCode,
		Title:        listing.Title,
		Description:  listing.Description,
	}
}

func mergeOneTimeProductListings(current []OneTimeProductListing, options OneTimeProductPatchOptions) []OneTimeProductListing {
	patch := options.Listing
	merged := make([]OneTimeProductListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			merged = append(merged, mergeOneTimeProductListing(listing, options))
			replaced = true
			continue
		}
		merged = append(merged, listing)
	}
	if !replaced {
		merged = append(merged, patch)
	}
	return merged
}

func mergeOneTimeProductListingPatches(current []OneTimeProductListing, patches []OneTimeProductBatchPatchListingRequest) []OneTimeProductListing {
	merged := append([]OneTimeProductListing(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductListingPatch(merged, patch.Listing)
	}
	return merged
}

func mergeOneTimeProductListingPatch(current []OneTimeProductListing, patch OneTimeProductListing) []OneTimeProductListing {
	merged := make([]OneTimeProductListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			merged = append(merged, patch)
			replaced = true
			continue
		}
		merged = append(merged, listing)
	}
	if !replaced {
		merged = append(merged, patch)
	}
	return merged
}

func mergeOneTimeProductListing(current OneTimeProductListing, options OneTimeProductPatchOptions) OneTimeProductListing {
	if options.TitleSet {
		current.Title = options.Listing.Title
	}
	if options.DescriptionSet {
		current.Description = options.Listing.Description
	}
	return current
}

func validateMergedOneTimeProductListing(languageCode string, listings []OneTimeProductListing) error {
	for _, listing := range listings {
		if listing.LanguageCode != languageCode {
			continue
		}
		if strings.TrimSpace(listing.Title) == "" {
			return fmt.Errorf("one-time product listing title is required when adding or replacing listing %s", languageCode)
		}
		if strings.TrimSpace(listing.Description) == "" {
			return fmt.Errorf("one-time product listing description is required when adding or replacing listing %s", languageCode)
		}
		return nil
	}
	return fmt.Errorf("one-time product listing %s was not merged", languageCode)
}

func oneTimeProductPurchaseOptionsFromAPI(apiOptions []rawOneTimeProductPurchaseOption) []OneTimeProductPurchaseOption {
	options := make([]OneTimeProductPurchaseOption, 0, len(apiOptions))
	for _, apiOption := range apiOptions {
		options = append(options, oneTimeProductPurchaseOptionFromAPI(apiOption))
	}
	return options
}

func oneTimeProductPurchaseOptionFromAPI(apiOption rawOneTimeProductPurchaseOption) OneTimeProductPurchaseOption {
	option := OneTimeProductPurchaseOption{
		PurchaseOptionID:         apiOption.PurchaseOptionID,
		State:                    apiOption.State,
		OfferTags:                rawOfferTagsFromAPI(apiOption.OfferTags),
		RegionalConfigs:          oneTimeProductRegionalConfigsFromAPI(apiOption.RegionalPricingAndAvailabilityConfigs),
		NewRegionsConfig:         oneTimeProductNewRegionsConfigFromAPI(apiOption.NewRegionsConfig),
		TaxAndComplianceSettings: oneTimeProductPurchaseOptionTaxComplianceSettingsFromAPI(apiOption.TaxAndComplianceSettings),
	}
	switch {
	case apiOption.BuyOption != nil:
		option.Type = OneTimeProductPurchaseOptionTypeBuy
		option.LegacyCompatible = apiOption.BuyOption.LegacyCompatible
		option.MultiQuantityEnabled = apiOption.BuyOption.MultiQuantityEnabled
	case apiOption.RentOption != nil:
		option.Type = OneTimeProductPurchaseOptionTypeRent
		option.RentalPeriod = apiOption.RentOption.RentalPeriod
		option.ExpirationPeriod = apiOption.RentOption.ExpirationPeriod
	}
	return option
}

func oneTimeProductPurchaseOptionsToAPI(options []OneTimeProductPurchaseOption) []*androidpublisher.OneTimeProductPurchaseOption {
	apiOptions := make([]*androidpublisher.OneTimeProductPurchaseOption, 0, len(options))
	for _, option := range options {
		apiOptions = append(apiOptions, oneTimeProductPurchaseOptionToAPI(option))
	}
	return apiOptions
}

func oneTimeProductPurchaseOptionToAPI(option OneTimeProductPurchaseOption) *androidpublisher.OneTimeProductPurchaseOption {
	apiOption := &androidpublisher.OneTimeProductPurchaseOption{
		PurchaseOptionId:                      option.PurchaseOptionID,
		OfferTags:                             offerTagsToAPI(option.OfferTags),
		RegionalPricingAndAvailabilityConfigs: oneTimeProductRegionalConfigsToAPI(option.RegionalConfigs),
		NewRegionsConfig:                      oneTimeProductNewRegionsConfigToAPI(option.NewRegionsConfig),
		TaxAndComplianceSettings:              oneTimeProductPurchaseOptionTaxComplianceSettingsToAPI(option.TaxAndComplianceSettings),
	}
	switch option.Type {
	case OneTimeProductPurchaseOptionTypeBuy:
		apiOption.BuyOption = &androidpublisher.OneTimeProductBuyPurchaseOption{
			LegacyCompatible:     option.LegacyCompatible,
			MultiQuantityEnabled: option.MultiQuantityEnabled,
		}
		if option.LegacyCompatible {
			apiOption.BuyOption.ForceSendFields = append(apiOption.BuyOption.ForceSendFields, "LegacyCompatible")
		}
		if option.MultiQuantityEnabled {
			apiOption.BuyOption.ForceSendFields = append(apiOption.BuyOption.ForceSendFields, "MultiQuantityEnabled")
		}
	case OneTimeProductPurchaseOptionTypeRent:
		apiOption.RentOption = &androidpublisher.OneTimeProductRentPurchaseOption{
			RentalPeriod:     option.RentalPeriod,
			ExpirationPeriod: option.ExpirationPeriod,
		}
	}
	return apiOption
}

func offerTagsToAPI(tags []string) []*androidpublisher.OfferTag {
	apiTags := make([]*androidpublisher.OfferTag, 0, len(tags))
	for _, tag := range tags {
		apiTags = append(apiTags, &androidpublisher.OfferTag{Tag: tag})
	}
	return apiTags
}

func oneTimeProductRegionalConfigsFromAPI(apiConfigs []rawOneTimeProductRegionalConfig) []OneTimeProductRegionalConfig {
	if len(apiConfigs) == 0 {
		return nil
	}
	configs := make([]OneTimeProductRegionalConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		configs = append(configs, OneTimeProductRegionalConfig{
			RegionCode:   apiConfig.RegionCode,
			Availability: apiConfig.Availability,
			Price:        rawMoneyFromAPI(apiConfig.Price),
		})
	}
	return configs
}

func oneTimeProductRegionalConfigsToAPI(configs []OneTimeProductRegionalConfig) []*androidpublisher.OneTimeProductPurchaseOptionRegionalPricingAndAvailabilityConfig {
	apiConfigs := make([]*androidpublisher.OneTimeProductPurchaseOptionRegionalPricingAndAvailabilityConfig, 0, len(configs))
	for _, config := range configs {
		apiConfigs = append(apiConfigs, &androidpublisher.OneTimeProductPurchaseOptionRegionalPricingAndAvailabilityConfig{
			RegionCode:   config.RegionCode,
			Availability: purchaseOptionAvailabilityToAPI(config.Availability),
			Price:        moneyToAPI(config.Price),
		})
	}
	return apiConfigs
}

func oneTimeProductNewRegionsConfigFromAPI(apiConfig *rawOneTimeProductNewRegionsConfig) *OneTimeProductNewRegionsPricingAndAvailability {
	if apiConfig == nil {
		return nil
	}
	return &OneTimeProductNewRegionsPricingAndAvailability{
		Availability: apiConfig.Availability,
		USDPrice:     rawMoneyFromAPI(apiConfig.USDPrice),
		EURPrice:     rawMoneyFromAPI(apiConfig.EURPrice),
	}
}

func oneTimeProductNewRegionsConfigToAPI(config *OneTimeProductNewRegionsPricingAndAvailability) *androidpublisher.OneTimeProductPurchaseOptionNewRegionsConfig {
	if config == nil {
		return nil
	}
	return &androidpublisher.OneTimeProductPurchaseOptionNewRegionsConfig{
		Availability: purchaseOptionAvailabilityToAPI(config.Availability),
		UsdPrice:     moneyToAPI(config.USDPrice),
		EurPrice:     moneyToAPI(config.EURPrice),
	}
}

func oneTimeProductTaxComplianceSettingsFromAPI(apiSettings *rawOneTimeProductTaxComplianceSettings) *OneTimeProductTaxComplianceSetting {
	if apiSettings == nil {
		return nil
	}
	return &OneTimeProductTaxComplianceSetting{
		IsTokenizedDigitalAsset: apiSettings.IsTokenizedDigitalAsset,
		ProductTaxCategoryCode:  apiSettings.ProductTaxCategoryCode,
		RegionalAgeRatings:      regionalAgeRatingsFromAPI(apiSettings.RegionalProductAgeRatingInfos),
		RegionalTaxConfigs:      regionalTaxConfigsFromAPI(apiSettings.RegionalTaxConfigs),
	}
}

func oneTimeProductTaxComplianceSettingsToAPI(settings *OneTimeProductTaxComplianceSetting) *androidpublisher.OneTimeProductTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	apiSettings := &androidpublisher.OneTimeProductTaxAndComplianceSettings{
		IsTokenizedDigitalAsset: settings.IsTokenizedDigitalAsset,
		RegionalTaxConfigs:      oneTimeProductRegionalTaxConfigsToAPI(settings.RegionalTaxConfigs),
	}
	if settings.IsTokenizedDigitalAsset {
		apiSettings.ForceSendFields = append(apiSettings.ForceSendFields, "IsTokenizedDigitalAsset")
	}
	return apiSettings
}

func regionalAgeRatingsFromAPI(apiRatings []rawRegionalProductAgeRating) []RegionalAgeRating {
	if len(apiRatings) == 0 {
		return nil
	}
	ratings := make([]RegionalAgeRating, 0, len(apiRatings))
	for _, apiRating := range apiRatings {
		ratings = append(ratings, RegionalAgeRating{
			RegionCode:           apiRating.RegionCode,
			ProductAgeRatingTier: apiRating.ProductAgeRatingTier,
		})
	}
	return ratings
}

func regionalTaxConfigsFromAPI(apiConfigs []rawRegionalTaxConfig) []RegionalTaxConfig {
	if len(apiConfigs) == 0 {
		return nil
	}
	configs := make([]RegionalTaxConfig, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		configs = append(configs, RegionalTaxConfig{
			RegionCode:                         apiConfig.RegionCode,
			EligibleForStreamingServiceTaxRate: apiConfig.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   apiConfig.StreamingTaxType,
			TaxTier:                            apiConfig.TaxTier,
		})
	}
	return configs
}

func oneTimeProductRegionalTaxConfigsToAPI(configs []RegionalTaxConfig) []*androidpublisher.RegionalTaxConfig {
	if len(configs) == 0 {
		return nil
	}
	apiConfigs := make([]*androidpublisher.RegionalTaxConfig, 0, len(configs))
	for _, config := range configs {
		apiConfigs = append(apiConfigs, &androidpublisher.RegionalTaxConfig{
			RegionCode:                         config.RegionCode,
			EligibleForStreamingServiceTaxRate: config.EligibleForStreamingServiceTaxRate,
			StreamingTaxType:                   config.StreamingTaxType,
			TaxTier:                            config.TaxTier,
		})
	}
	return apiConfigs
}

func oneTimeProductPurchaseOptionTaxComplianceSettingsFromAPI(apiSettings *rawPurchaseOptionTaxComplianceSettings) *OneTimeProductPurchaseOptionTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	return &OneTimeProductPurchaseOptionTaxComplianceSettings{
		WithdrawalRightType: apiSettings.WithdrawalRightType,
	}
}

func oneTimeProductPurchaseOptionTaxComplianceSettingsToAPI(settings *OneTimeProductPurchaseOptionTaxComplianceSettings) *androidpublisher.PurchaseOptionTaxAndComplianceSettings {
	if settings == nil {
		return nil
	}
	return &androidpublisher.PurchaseOptionTaxAndComplianceSettings{
		WithdrawalRightType: settings.WithdrawalRightType,
	}
}

func mergePurchaseOptionAvailabilityPatches(current []OneTimeProductPurchaseOption, patches []PurchaseOptionAvailabilityPatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := append([]OneTimeProductPurchaseOption(nil), current...)
	for _, patch := range patches {
		next, err := mergePurchaseOptionAvailabilityPatch(merged, patch)
		if err != nil {
			return nil, err
		}
		merged = next
	}
	return merged, nil
}

func mergePurchaseOptionAvailabilityPatch(current []OneTimeProductPurchaseOption, patch PurchaseOptionAvailabilityPatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := make([]OneTimeProductPurchaseOption, 0, len(current))
	for _, option := range current {
		if option.PurchaseOptionID != patch.PurchaseOptionID.String() {
			merged = append(merged, option)
			continue
		}
		regionalConfigs, err := mergePurchaseOptionRegionalAvailabilityPatch(option.RegionalConfigs, patch)
		if err != nil {
			return nil, err
		}
		option.RegionalConfigs = regionalConfigs
		merged = append(merged, option)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("purchase option availability patch for %s/%s references a purchase option that is not already configured", patch.ProductID, patch.PurchaseOptionID)
}

func mergePurchaseOptionRegionalAvailabilityPatch(current []OneTimeProductRegionalConfig, patch PurchaseOptionAvailabilityPatchRequest) ([]OneTimeProductRegionalConfig, error) {
	merged := make([]OneTimeProductRegionalConfig, 0, len(current))
	for _, region := range current {
		if region.RegionCode != patch.RegionCode {
			merged = append(merged, region)
			continue
		}
		if patch.Availability == PurchaseOptionAvailabilityNoLongerAvailable && region.Availability != "AVAILABLE" {
			return nil, fmt.Errorf("purchase option availability patch for %s/%s/%s cannot set noLongerAvailable from current availability %q", patch.ProductID, patch.PurchaseOptionID, patch.RegionCode, region.Availability)
		}
		region.Availability = purchaseOptionAvailabilityToAPI(patch.Availability.String())
		merged = append(merged, region)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("purchase option availability patch for %s/%s/%s references a region that is not already configured; availability-only patches cannot add regional price data", patch.ProductID, patch.PurchaseOptionID, patch.RegionCode)
}

func mergePurchaseOptionPricePatches(current []OneTimeProductPurchaseOption, patches []PurchaseOptionPricePatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := append([]OneTimeProductPurchaseOption(nil), current...)
	for _, patch := range patches {
		next, err := mergePurchaseOptionPricePatch(merged, patch)
		if err != nil {
			return nil, err
		}
		merged = next
	}
	return merged, nil
}

func mergePurchaseOptionPricePatch(current []OneTimeProductPurchaseOption, patch PurchaseOptionPricePatchRequest) ([]OneTimeProductPurchaseOption, error) {
	merged := make([]OneTimeProductPurchaseOption, 0, len(current))
	for _, option := range current {
		if option.PurchaseOptionID != patch.PurchaseOptionID.String() {
			merged = append(merged, option)
			continue
		}
		regionalConfigs, err := mergePurchaseOptionRegionalPricePatch(option.RegionalConfigs, patch)
		if err != nil {
			return nil, err
		}
		option.RegionalConfigs = regionalConfigs
		merged = append(merged, option)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("purchase option price patch for %s/%s references a purchase option that is not already configured", patch.ProductID, patch.PurchaseOptionID)
}

func mergePurchaseOptionRegionalPricePatch(current []OneTimeProductRegionalConfig, patch PurchaseOptionPricePatchRequest) ([]OneTimeProductRegionalConfig, error) {
	merged := make([]OneTimeProductRegionalConfig, 0, len(current))
	for _, region := range current {
		if region.RegionCode != patch.RegionCode {
			merged = append(merged, region)
			continue
		}
		region.Price = &Money{
			CurrencyCode: patch.Price.CurrencyCode,
			Units:        patch.Price.Units,
			Nanos:        patch.Price.Nanos,
		}
		merged = append(merged, region)
		return append(merged, current[len(merged):]...), nil
	}
	return nil, fmt.Errorf("purchase option price patch for %s/%s/%s references a region that is not already configured; price patches cannot add regional availability", patch.ProductID, patch.PurchaseOptionID, patch.RegionCode)
}

func purchaseOptionAvailabilityToAPI(availability string) string {
	switch PurchaseOptionAvailability(availability) {
	case PurchaseOptionAvailabilityAvailable:
		return "AVAILABLE"
	case PurchaseOptionAvailabilityNoLongerAvailable:
		return "NO_LONGER_AVAILABLE"
	case PurchaseOptionAvailabilityAvailableIfReleased:
		return "AVAILABLE_IF_RELEASED"
	case PurchaseOptionAvailabilityAvailableForOffersOnly:
		return "AVAILABLE_FOR_OFFERS_ONLY"
	default:
		return availability
	}
}

func regionsVersionFromAPI(apiVersion *rawRegionsVersion) *RegionsVersion {
	if apiVersion == nil {
		return nil
	}
	return &RegionsVersion{Version: apiVersion.Version}
}

func oneTimeProductOfferListResultFromAPI(options OneTimeProductOfferListOptions, response *androidpublisher.ListOneTimeProductOffersResponse) OneTimeProductOfferListResult {
	result := OneTimeProductOfferListResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Offers:           []OneTimeProductOffer{},
		Options:          options,
	}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiOffer := range response.OneTimeProductOffers {
		result.Offers = append(result.Offers, oneTimeProductOfferFromAPI(apiOffer))
	}
	return result
}

func oneTimeProductOfferBatchGetResultFromAPI(options OneTimeProductOfferBatchGetOptions, response *androidpublisher.BatchGetOneTimeProductOffersResponse) OneTimeProductOfferBatchGetResult {
	result := OneTimeProductOfferBatchGetResult{
		PackageName:      options.PackageName,
		ProductID:        options.ProductID,
		PurchaseOptionID: options.PurchaseOptionID,
		Offers:           []OneTimeProductOffer{},
		Options:          options,
	}
	if response == nil {
		return result
	}
	byKey := make(map[string]OneTimeProductOffer, len(response.OneTimeProductOffers))
	extras := make([]OneTimeProductOffer, 0)
	for _, apiOffer := range response.OneTimeProductOffers {
		offer := oneTimeProductOfferFromAPI(apiOffer)
		key := oneTimeProductOfferKey(offer.ProductID, offer.PurchaseOptionID, offer.OfferID)
		if key == "//" {
			continue
		}
		if _, ok := byKey[key]; ok {
			extras = append(extras, offer)
			continue
		}
		byKey[key] = offer
	}
	for _, request := range options.Requests {
		key := oneTimeProductOfferKey(request.ProductID, request.PurchaseOptionID, request.OfferID)
		if offer, ok := byKey[key]; ok {
			result.Offers = append(result.Offers, offer)
			delete(byKey, key)
		}
	}
	for _, offer := range byKey {
		extras = append(extras, offer)
	}
	sort.Slice(extras, func(i, j int) bool {
		return oneTimeProductOfferKey(extras[i].ProductID, extras[i].PurchaseOptionID, extras[i].OfferID) < oneTimeProductOfferKey(extras[j].ProductID, extras[j].PurchaseOptionID, extras[j].OfferID)
	})
	result.Offers = append(result.Offers, extras...)
	return result
}

func batchGetOneTimeProductOffersRequestToAPI(packageName PackageName, requests []OneTimeProductOfferBatchGetRequest) *androidpublisher.BatchGetOneTimeProductOffersRequest {
	apiRequest := &androidpublisher.BatchGetOneTimeProductOffersRequest{
		Requests: make([]*androidpublisher.GetOneTimeProductOfferRequest, 0, len(requests)),
	}
	for _, request := range requests {
		apiRequest.Requests = append(apiRequest.Requests, &androidpublisher.GetOneTimeProductOfferRequest{
			PackageName:      packageName.String(),
			ProductId:        request.ProductID.String(),
			PurchaseOptionId: request.PurchaseOptionID.String(),
			OfferId:          request.OfferID.String(),
		})
	}
	return apiRequest
}

func oneTimeProductOfferKey(productID OneTimeProductID, purchaseOptionID OneTimeProductPurchaseOptionID, offerID OneTimeProductOfferID) string {
	return productID.String() + "/" + purchaseOptionID.String() + "/" + offerID.String()
}

func oneTimeProductOfferFromAPI(apiOffer *androidpublisher.OneTimeProductOffer) OneTimeProductOffer {
	if apiOffer == nil {
		return OneTimeProductOffer{}
	}
	offer := OneTimeProductOffer{
		PackageName:      PackageName(apiOffer.PackageName),
		ProductID:        OneTimeProductID(apiOffer.ProductId),
		PurchaseOptionID: OneTimeProductPurchaseOptionID(apiOffer.PurchaseOptionId),
		OfferID:          OneTimeProductOfferID(apiOffer.OfferId),
		State:            apiOffer.State,
		OfferTags:        offerTagsFromAPI(apiOffer.OfferTags),
		RegionsVersion:   regionsVersionFromGeneratedAPI(apiOffer.RegionsVersion),
		RegionalConfigs:  oneTimeProductOfferRegionsFromAPI(apiOffer.RegionalPricingAndAvailabilityConfigs),
	}
	switch {
	case apiOffer.DiscountedOffer != nil:
		offer.Type = OneTimeProductOfferTypeDiscounted
		offer.DiscountedOffer = oneTimeProductDiscountedOfferFromAPI(apiOffer.DiscountedOffer)
	case apiOffer.PreOrderOffer != nil:
		offer.Type = OneTimeProductOfferTypePreOrder
		offer.PreOrderOffer = oneTimeProductPreOrderOfferFromAPI(apiOffer.PreOrderOffer)
	}
	return offer
}

func oneTimeProductDiscountedOfferFromAPI(apiOffer *androidpublisher.OneTimeProductDiscountedOffer) *OneTimeProductDiscountedOffer {
	if apiOffer == nil {
		return nil
	}
	return &OneTimeProductDiscountedOffer{
		StartTime:       apiOffer.StartTime,
		EndTime:         apiOffer.EndTime,
		RedemptionLimit: apiOffer.RedemptionLimit,
	}
}

func oneTimeProductPreOrderOfferFromAPI(apiOffer *androidpublisher.OneTimeProductPreOrderOffer) *OneTimeProductPreOrderOffer {
	if apiOffer == nil {
		return nil
	}
	return &OneTimeProductPreOrderOffer{
		StartTime:           apiOffer.StartTime,
		EndTime:             apiOffer.EndTime,
		ReleaseTime:         apiOffer.ReleaseTime,
		PriceChangeBehavior: apiOffer.PriceChangeBehavior,
	}
}

func oneTimeProductOfferRegionsFromAPI(apiConfigs []*androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig) []OneTimeProductOfferRegion {
	if len(apiConfigs) == 0 {
		return nil
	}
	regions := make([]OneTimeProductOfferRegion, 0, len(apiConfigs))
	for _, apiConfig := range apiConfigs {
		if apiConfig == nil {
			continue
		}
		regions = append(regions, OneTimeProductOfferRegion{
			RegionCode:       apiConfig.RegionCode,
			Availability:     apiConfig.Availability,
			AbsoluteDiscount: moneyFromAPI(apiConfig.AbsoluteDiscount),
			RelativeDiscount: apiConfig.RelativeDiscount,
			NoOverride:       apiConfig.NoOverride != nil,
		})
	}
	return regions
}

func oneTimeProductOfferRegionsToAPI(regions []OneTimeProductOfferRegion) []*androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig {
	apiRegions := make([]*androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig, 0, len(regions))
	for _, region := range regions {
		apiRegion := &androidpublisher.OneTimeProductOfferRegionalPricingAndAvailabilityConfig{
			RegionCode:       region.RegionCode,
			Availability:     oneTimeProductOfferAvailabilityToAPI(region.Availability),
			AbsoluteDiscount: moneyToAPI(region.AbsoluteDiscount),
			RelativeDiscount: region.RelativeDiscount,
		}
		if region.NoOverride {
			apiRegion.NoOverride = &androidpublisher.OneTimeProductOfferNoPriceOverrideOptions{}
		}
		if region.RelativeDiscount != 0 {
			apiRegion.ForceSendFields = append(apiRegion.ForceSendFields, "RelativeDiscount")
		}
		apiRegions = append(apiRegions, apiRegion)
	}
	return apiRegions
}

func mergeOneTimeProductOfferAvailabilityPatches(current []OneTimeProductOfferRegion, patches []OneTimeProductOfferAvailabilityPatchRequest) []OneTimeProductOfferRegion {
	merged := append([]OneTimeProductOfferRegion(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductOfferAvailabilityPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeOneTimeProductOfferAvailabilityPatch(current []OneTimeProductOfferRegion, patch OneTimeProductOfferAvailabilityPatchRequest) []OneTimeProductOfferRegion {
	merged := make([]OneTimeProductOfferRegion, 0, len(current)+1)
	replaced := false
	for _, region := range current {
		if region.RegionCode == patch.RegionCode {
			region.Availability = oneTimeProductOfferAvailabilityToAPI(patch.Availability.String())
			merged = append(merged, region)
			replaced = true
			continue
		}
		merged = append(merged, region)
	}
	if !replaced {
		return nil
	}
	return merged
}

func mergeOneTimeProductOfferRelativeDiscountPatches(current []OneTimeProductOfferRegion, patches []OneTimeProductOfferRelativeDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := append([]OneTimeProductOfferRegion(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductOfferRelativeDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeOneTimeProductOfferRelativeDiscountPatch(current []OneTimeProductOfferRegion, patch OneTimeProductOfferRelativeDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := make([]OneTimeProductOfferRegion, 0, len(current)+1)
	replaced := false
	for _, region := range current {
		if region.RegionCode == patch.RegionCode {
			region.RelativeDiscount = patch.RelativeDiscount
			region.AbsoluteDiscount = nil
			region.NoOverride = false
			merged = append(merged, region)
			replaced = true
			continue
		}
		merged = append(merged, region)
	}
	if !replaced {
		return nil
	}
	return merged
}

func mergeOneTimeProductOfferAbsoluteDiscountPatches(current []OneTimeProductOfferRegion, patches []OneTimeProductOfferAbsoluteDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := append([]OneTimeProductOfferRegion(nil), current...)
	for _, patch := range patches {
		merged = mergeOneTimeProductOfferAbsoluteDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeOneTimeProductOfferAbsoluteDiscountPatch(current []OneTimeProductOfferRegion, patch OneTimeProductOfferAbsoluteDiscountPatchRequest) []OneTimeProductOfferRegion {
	merged := make([]OneTimeProductOfferRegion, 0, len(current)+1)
	replaced := false
	for _, region := range current {
		if region.RegionCode == patch.RegionCode {
			absoluteDiscount := patch.AbsoluteDiscount
			region.AbsoluteDiscount = &absoluteDiscount
			region.RelativeDiscount = 0
			region.NoOverride = false
			merged = append(merged, region)
			replaced = true
			continue
		}
		merged = append(merged, region)
	}
	if !replaced {
		return nil
	}
	return merged
}

func oneTimeProductOfferAvailabilityToAPI(availability string) string {
	switch OneTimeProductOfferAvailability(availability) {
	case OneTimeProductOfferAvailabilityAvailable:
		return "AVAILABLE"
	case OneTimeProductOfferAvailabilityNoLongerAvailable:
		return "NO_LONGER_AVAILABLE"
	default:
		return availability
	}
}

func moneyToAPI(money *Money) *androidpublisher.Money {
	if money == nil {
		return nil
	}
	return &androidpublisher.Money{
		CurrencyCode: money.CurrencyCode,
		Units:        money.Units,
		Nanos:        money.Nanos,
	}
}

func regionsVersionFromGeneratedAPI(apiVersion *androidpublisher.RegionsVersion) *RegionsVersion {
	if apiVersion == nil {
		return nil
	}
	return &RegionsVersion{Version: apiVersion.Version}
}

func managedProductTaxComplianceSettingsFromAPI(apiSettings *androidpublisher.ManagedProductTaxAndComplianceSettings) *ProductTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	tokenizedDigitalAsset := apiSettings.IsTokenizedDigitalAsset
	return &ProductTaxComplianceSettings{
		EEAWithdrawalRightType:  apiSettings.EeaWithdrawalRightType,
		IsTokenizedDigitalAsset: &tokenizedDigitalAsset,
		TaxRateInfoByRegionCode: regionalTaxRateInfoFromAPI(apiSettings.TaxRateInfoByRegionCode),
	}
}

func subscriptionTaxComplianceSettingsFromAPI(apiSettings *androidpublisher.SubscriptionTaxAndComplianceSettings) *ProductTaxComplianceSettings {
	if apiSettings == nil {
		return nil
	}
	tokenizedDigitalAsset := apiSettings.IsTokenizedDigitalAsset
	return &ProductTaxComplianceSettings{
		EEAWithdrawalRightType:  apiSettings.EeaWithdrawalRightType,
		IsTokenizedDigitalAsset: &tokenizedDigitalAsset,
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

func subscriptionCreateToAPI(options SubscriptionCreateOptions) *androidpublisher.Subscription {
	subscription := subscriptionCreateDesiredSubscription(options)
	return &androidpublisher.Subscription{
		PackageName:                subscription.PackageName.String(),
		ProductId:                  subscription.ProductID.String(),
		Listings:                   subscriptionListingsToAPI(subscription.Listings),
		BasePlans:                  subscriptionBasePlansToAPI(subscription.BasePlans),
		RestrictedPaymentCountries: restrictedCountriesToAPI(subscription.RestrictedCountries),
		TaxAndComplianceSettings:   subscriptionTaxComplianceSettingsToAPI(subscription.TaxAndComplianceSettings),
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

func subscriptionListingsToAPI(listings []SubscriptionListing) []*androidpublisher.SubscriptionListing {
	apiListings := make([]*androidpublisher.SubscriptionListing, 0, len(listings))
	for _, listing := range listings {
		apiListings = append(apiListings, subscriptionListingToAPI(listing))
	}
	return apiListings
}

func subscriptionListingToAPI(listing SubscriptionListing) *androidpublisher.SubscriptionListing {
	return &androidpublisher.SubscriptionListing{
		LanguageCode: listing.LanguageCode,
		Title:        listing.Title,
		Description:  listing.Description,
		Benefits:     listing.Benefits,
	}
}

func mergeSubscriptionListings(current []SubscriptionListing, options SubscriptionPatchOptions) []SubscriptionListing {
	patch := options.Listing
	merged := make([]SubscriptionListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			merged = append(merged, mergeSubscriptionListing(listing, options))
			replaced = true
			continue
		}
		merged = append(merged, listing)
	}
	if !replaced {
		merged = append(merged, patch)
	}
	return merged
}

func mergeSubscriptionListing(current SubscriptionListing, options SubscriptionPatchOptions) SubscriptionListing {
	current.Title = options.Listing.Title
	if options.DescriptionSet {
		current.Description = options.Listing.Description
	}
	if options.BenefitsSet {
		current.Benefits = options.Listing.Benefits
	}
	return current
}

func mergeSubscriptionListingPatches(current []SubscriptionListing, patches []SubscriptionBatchPatchListingRequest) []SubscriptionListing {
	merged := append([]SubscriptionListing(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionListingPatch(merged, patch.Listing)
	}
	return merged
}

func mergeSubscriptionListingPatch(current []SubscriptionListing, patch SubscriptionListing) []SubscriptionListing {
	merged := make([]SubscriptionListing, 0, len(current)+1)
	replaced := false
	for _, listing := range current {
		if listing.LanguageCode == patch.LanguageCode {
			listing.Title = patch.Title
			listing.Description = patch.Description
			merged = append(merged, listing)
			replaced = true
			continue
		}
		merged = append(merged, listing)
	}
	if !replaced {
		merged = append(merged, patch)
	}
	return merged
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

func subscriptionBasePlansToAPI(basePlans []SubscriptionBasePlan) []*androidpublisher.BasePlan {
	apiBasePlans := make([]*androidpublisher.BasePlan, 0, len(basePlans))
	for _, basePlan := range basePlans {
		apiBasePlans = append(apiBasePlans, subscriptionBasePlanToAPI(basePlan))
	}
	return apiBasePlans
}

func subscriptionBasePlanToAPI(basePlan SubscriptionBasePlan) *androidpublisher.BasePlan {
	apiBasePlan := &androidpublisher.BasePlan{
		BasePlanId:         basePlan.BasePlanID,
		OfferTags:          offerTagsToAPI(basePlan.OfferTags),
		RegionalConfigs:    subscriptionRegionalConfigsToAPI(basePlan.RegionalConfigs),
		OtherRegionsConfig: subscriptionOtherRegionsConfigToAPI(basePlan.OtherRegionsConfig),
	}
	switch basePlan.Type {
	case SubscriptionBasePlanTypeAutoRenewing:
		apiBasePlan.AutoRenewingBasePlanType = &androidpublisher.AutoRenewingBasePlanType{
			BillingPeriodDuration:               basePlan.BillingPeriodDuration,
			GracePeriodDuration:                 basePlan.GracePeriodDuration,
			AccountHoldDuration:                 basePlan.AccountHoldDuration,
			LegacyCompatible:                    basePlan.LegacyCompatible,
			LegacyCompatibleSubscriptionOfferId: basePlan.LegacyCompatibleSubscriptionOfferID,
			ProrationMode:                       basePlan.ProrationMode,
			ResubscribeState:                    basePlan.ResubscribeState,
		}
		if basePlan.LegacyCompatible {
			apiBasePlan.AutoRenewingBasePlanType.ForceSendFields = append(apiBasePlan.AutoRenewingBasePlanType.ForceSendFields, "LegacyCompatible")
		}
	case SubscriptionBasePlanTypePrepaid:
		apiBasePlan.PrepaidBasePlanType = &androidpublisher.PrepaidBasePlanType{
			BillingPeriodDuration: basePlan.BillingPeriodDuration,
			TimeExtension:         basePlan.TimeExtension,
		}
	case SubscriptionBasePlanTypeInstallments:
		apiBasePlan.InstallmentsBasePlanType = &androidpublisher.InstallmentsBasePlanType{
			BillingPeriodDuration:  basePlan.BillingPeriodDuration,
			GracePeriodDuration:    basePlan.GracePeriodDuration,
			AccountHoldDuration:    basePlan.AccountHoldDuration,
			CommittedPaymentsCount: basePlan.CommittedPaymentsCount,
			ProrationMode:          basePlan.ProrationMode,
			RenewalType:            basePlan.RenewalType,
			ResubscribeState:       basePlan.ResubscribeState,
		}
	}
	return apiBasePlan
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

func subscriptionRegionalConfigsToAPI(configs []SubscriptionRegionalConfig) []*androidpublisher.RegionalBasePlanConfig {
	apiConfigs := make([]*androidpublisher.RegionalBasePlanConfig, 0, len(configs))
	for _, config := range configs {
		apiConfig := &androidpublisher.RegionalBasePlanConfig{
			RegionCode:                config.RegionCode,
			NewSubscriberAvailability: config.NewSubscriberAvailability,
			Price:                     moneyToAPI(config.Price),
		}
		apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "NewSubscriberAvailability")
		apiConfigs = append(apiConfigs, apiConfig)
	}
	return apiConfigs
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

func subscriptionOtherRegionsConfigToAPI(config *SubscriptionOtherRegionsConfig) *androidpublisher.OtherRegionsBasePlanConfig {
	if config == nil {
		return nil
	}
	apiConfig := &androidpublisher.OtherRegionsBasePlanConfig{
		NewSubscriberAvailability: config.NewSubscriberAvailability,
		UsdPrice:                  moneyToAPI(config.USDPrice),
		EurPrice:                  moneyToAPI(config.EURPrice),
	}
	apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "NewSubscriberAvailability")
	return apiConfig
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

func rawMoneyFromAPI(apiMoney *rawMoney) *Money {
	if apiMoney == nil {
		return nil
	}
	return &Money{
		CurrencyCode: apiMoney.CurrencyCode,
		Units:        int64(apiMoney.Units),
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

func rawOfferTagsFromAPI(apiOfferTags []rawOfferTag) []string {
	if len(apiOfferTags) == 0 {
		return nil
	}
	tags := make([]string, 0, len(apiOfferTags))
	for _, apiOfferTag := range apiOfferTags {
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

func restrictedCountriesToAPI(countries []string) *androidpublisher.RestrictedPaymentCountries {
	if len(countries) == 0 {
		return nil
	}
	return &androidpublisher.RestrictedPaymentCountries{RegionCodes: append([]string(nil), countries...)}
}

func rawRestrictedCountriesFromAPI(apiCountries *rawRestrictedPaymentCountries) []string {
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

func subscriptionOfferCreateToAPI(options SubscriptionOfferCreateOptions) *androidpublisher.SubscriptionOffer {
	offer := subscriptionOfferCreateDesiredOffer(options)
	return &androidpublisher.SubscriptionOffer{
		PackageName:        offer.PackageName.String(),
		ProductId:          offer.ProductID.String(),
		BasePlanId:         offer.BasePlanID.String(),
		OfferId:            offer.OfferID.String(),
		OfferTags:          offerTagsToAPI(offer.OfferTags),
		RegionalConfigs:    subscriptionOfferRegionalConfigsToAPI(offer.RegionalConfigs),
		OtherRegionsConfig: subscriptionOfferOtherRegionsConfigToAPI(offer.OtherRegionsConfig),
		Phases:             subscriptionOfferPhasesToAPI(offer.Phases),
		Targeting:          subscriptionOfferTargetingToAPI(offer.Targeting),
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

func subscriptionOfferRegionalConfigsToAPI(configs []SubscriptionOfferRegionalConfig) []*androidpublisher.RegionalSubscriptionOfferConfig {
	apiConfigs := make([]*androidpublisher.RegionalSubscriptionOfferConfig, 0, len(configs))
	for _, config := range configs {
		apiConfig := &androidpublisher.RegionalSubscriptionOfferConfig{
			RegionCode:                config.RegionCode,
			NewSubscriberAvailability: config.NewSubscriberAvailability,
		}
		apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "NewSubscriberAvailability")
		apiConfigs = append(apiConfigs, apiConfig)
	}
	return apiConfigs
}

func mergeSubscriptionOfferAvailabilityPatches(current []SubscriptionOfferRegionalConfig, patches []SubscriptionOfferAvailabilityPatchRequest) []SubscriptionOfferRegionalConfig {
	merged := append([]SubscriptionOfferRegionalConfig(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferAvailabilityPatch(merged, patch)
	}
	return merged
}

func mergeSubscriptionOfferAvailabilityPatch(current []SubscriptionOfferRegionalConfig, patch SubscriptionOfferAvailabilityPatchRequest) []SubscriptionOfferRegionalConfig {
	merged := make([]SubscriptionOfferRegionalConfig, 0, len(current)+1)
	replaced := false
	for _, config := range current {
		if config.RegionCode == patch.RegionCode {
			config.NewSubscriberAvailability = patch.Availability
			merged = append(merged, config)
			replaced = true
			continue
		}
		merged = append(merged, config)
	}
	if !replaced {
		merged = append(merged, SubscriptionOfferRegionalConfig{
			RegionCode:                patch.RegionCode,
			NewSubscriberAvailability: patch.Availability,
		})
	}
	return merged
}

func subscriptionOfferOtherRegionsConfigFromAPI(apiConfig *androidpublisher.OtherRegionsSubscriptionOfferConfig) *SubscriptionOfferOtherRegionsConfig {
	if apiConfig == nil {
		return nil
	}
	return &SubscriptionOfferOtherRegionsConfig{NewSubscriberAvailability: apiConfig.OtherRegionsNewSubscriberAvailability}
}

func subscriptionOfferOtherRegionsConfigToAPI(config *SubscriptionOfferOtherRegionsConfig) *androidpublisher.OtherRegionsSubscriptionOfferConfig {
	if config == nil {
		return nil
	}
	apiConfig := &androidpublisher.OtherRegionsSubscriptionOfferConfig{
		OtherRegionsNewSubscriberAvailability: config.NewSubscriberAvailability,
	}
	apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "OtherRegionsNewSubscriberAvailability")
	return apiConfig
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

func subscriptionOfferPhasesToAPI(phases []SubscriptionOfferPhase) []*androidpublisher.SubscriptionOfferPhase {
	apiPhases := make([]*androidpublisher.SubscriptionOfferPhase, 0, len(phases))
	for _, phase := range phases {
		apiPhases = append(apiPhases, &androidpublisher.SubscriptionOfferPhase{
			Duration:           phase.Duration,
			RecurrenceCount:    phase.RecurrenceCount,
			RegionalConfigs:    subscriptionOfferPhaseRegionalConfigsToAPI(phase.RegionalConfigs),
			OtherRegionsConfig: subscriptionOfferPhaseOtherRegionsConfigToAPI(phase.OtherRegionsConfig),
		})
	}
	return apiPhases
}

func subscriptionOfferPhaseRegionalConfigsToAPI(configs []SubscriptionOfferPhaseRegionalConfig) []*androidpublisher.RegionalSubscriptionOfferPhaseConfig {
	apiConfigs := make([]*androidpublisher.RegionalSubscriptionOfferPhaseConfig, 0, len(configs))
	for _, config := range configs {
		apiConfig := &androidpublisher.RegionalSubscriptionOfferPhaseConfig{
			RegionCode:       config.RegionCode,
			Price:            moneyToAPI(config.Price),
			AbsoluteDiscount: moneyToAPI(config.AbsoluteDiscount),
			RelativeDiscount: config.RelativeDiscount,
		}
		if config.Free {
			apiConfig.Free = &androidpublisher.RegionalSubscriptionOfferPhaseFreePriceOverride{}
		}
		if config.RelativeDiscount != 0 {
			apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "RelativeDiscount")
		}
		apiConfigs = append(apiConfigs, apiConfig)
	}
	return apiConfigs
}

func subscriptionOfferPhaseOtherRegionsConfigToAPI(config *SubscriptionOfferPhaseOtherRegionsConfig) *androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig {
	if config == nil {
		return nil
	}
	apiConfig := &androidpublisher.OtherRegionsSubscriptionOfferPhaseConfig{
		OtherRegionsPrices: otherRegionsSubscriptionOfferPhasePricesToAPI(config.OtherRegionsPrices),
		AbsoluteDiscounts:  otherRegionsSubscriptionOfferPhasePricesToAPI(config.AbsoluteDiscounts),
		RelativeDiscount:   config.RelativeDiscount,
	}
	if config.Free {
		apiConfig.Free = &androidpublisher.OtherRegionsSubscriptionOfferPhaseFreePriceOverride{}
	}
	if config.RelativeDiscount != 0 {
		apiConfig.ForceSendFields = append(apiConfig.ForceSendFields, "RelativeDiscount")
	}
	return apiConfig
}

func otherRegionsSubscriptionOfferPhasePricesToAPI(prices *SubscriptionOfferOtherRegionsPrices) *androidpublisher.OtherRegionsSubscriptionOfferPhasePrices {
	if prices == nil {
		return nil
	}
	return &androidpublisher.OtherRegionsSubscriptionOfferPhasePrices{
		UsdPrice: moneyToAPI(prices.USDPrice),
		EurPrice: moneyToAPI(prices.EURPrice),
	}
}

func mergeSubscriptionOfferPhaseRelativeDiscountPatches(current []SubscriptionOfferPhase, patches []SubscriptionOfferPhaseRelativeDiscountPatchRequest) []SubscriptionOfferPhase {
	merged := append([]SubscriptionOfferPhase(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferPhaseRelativeDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeSubscriptionOfferPhaseRelativeDiscountPatch(current []SubscriptionOfferPhase, patch SubscriptionOfferPhaseRelativeDiscountPatchRequest) []SubscriptionOfferPhase {
	if patch.PhaseIndex < 0 || patch.PhaseIndex >= len(current) {
		return nil
	}
	merged := append([]SubscriptionOfferPhase(nil), current...)
	phase := merged[patch.PhaseIndex]
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(phase.RegionalConfigs))
	replaced := false
	for _, config := range phase.RegionalConfigs {
		if config.RegionCode == patch.RegionCode {
			config.RelativeDiscount = patch.RelativeDiscount
			config.Price = nil
			config.AbsoluteDiscount = nil
			config.Free = false
			configs = append(configs, config)
			replaced = true
			continue
		}
		configs = append(configs, config)
	}
	if !replaced {
		return nil
	}
	phase.RegionalConfigs = configs
	merged[patch.PhaseIndex] = phase
	return merged
}

func mergeSubscriptionOfferPhaseAbsoluteDiscountPatches(current []SubscriptionOfferPhase, patches []SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) []SubscriptionOfferPhase {
	merged := append([]SubscriptionOfferPhase(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferPhaseAbsoluteDiscountPatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeSubscriptionOfferPhaseAbsoluteDiscountPatch(current []SubscriptionOfferPhase, patch SubscriptionOfferPhaseAbsoluteDiscountPatchRequest) []SubscriptionOfferPhase {
	if patch.PhaseIndex < 0 || patch.PhaseIndex >= len(current) {
		return nil
	}
	merged := append([]SubscriptionOfferPhase(nil), current...)
	phase := merged[patch.PhaseIndex]
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(phase.RegionalConfigs))
	replaced := false
	for _, config := range phase.RegionalConfigs {
		if config.RegionCode == patch.RegionCode {
			absoluteDiscount := patch.AbsoluteDiscount
			config.AbsoluteDiscount = &absoluteDiscount
			config.Price = nil
			config.RelativeDiscount = 0
			config.Free = false
			configs = append(configs, config)
			replaced = true
			continue
		}
		configs = append(configs, config)
	}
	if !replaced {
		return nil
	}
	phase.RegionalConfigs = configs
	merged[patch.PhaseIndex] = phase
	return merged
}

func mergeSubscriptionOfferPhasePricePatches(current []SubscriptionOfferPhase, patches []SubscriptionOfferPhasePricePatchRequest) []SubscriptionOfferPhase {
	merged := append([]SubscriptionOfferPhase(nil), current...)
	for _, patch := range patches {
		merged = mergeSubscriptionOfferPhasePricePatch(merged, patch)
		if merged == nil {
			return nil
		}
	}
	return merged
}

func mergeSubscriptionOfferPhasePricePatch(current []SubscriptionOfferPhase, patch SubscriptionOfferPhasePricePatchRequest) []SubscriptionOfferPhase {
	if patch.PhaseIndex < 0 || patch.PhaseIndex >= len(current) {
		return nil
	}
	merged := append([]SubscriptionOfferPhase(nil), current...)
	phase := merged[patch.PhaseIndex]
	configs := make([]SubscriptionOfferPhaseRegionalConfig, 0, len(phase.RegionalConfigs))
	replaced := false
	for _, config := range phase.RegionalConfigs {
		if config.RegionCode == patch.RegionCode {
			price := patch.Price
			config.Price = &price
			config.AbsoluteDiscount = nil
			config.RelativeDiscount = 0
			config.Free = false
			configs = append(configs, config)
			replaced = true
			continue
		}
		configs = append(configs, config)
	}
	if !replaced {
		return nil
	}
	phase.RegionalConfigs = configs
	merged[patch.PhaseIndex] = phase
	return merged
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

func subscriptionOfferTargetingToAPI(targeting *SubscriptionOfferTargeting) *androidpublisher.SubscriptionOfferTargeting {
	if targeting == nil {
		return nil
	}
	return &androidpublisher.SubscriptionOfferTargeting{
		AcquisitionRule: subscriptionOfferAcquisitionTargetingToAPI(targeting.Acquisition),
		UpgradeRule:     subscriptionOfferUpgradeTargetingToAPI(targeting.Upgrade),
	}
}

func subscriptionOfferAcquisitionTargetingFromAPI(apiRule *androidpublisher.AcquisitionTargetingRule) *SubscriptionOfferAcquisitionTargeting {
	if apiRule == nil {
		return nil
	}
	return &SubscriptionOfferAcquisitionTargeting{Scope: subscriptionOfferTargetingScopeFromAPI(apiRule.Scope)}
}

func subscriptionOfferAcquisitionTargetingToAPI(rule *SubscriptionOfferAcquisitionTargeting) *androidpublisher.AcquisitionTargetingRule {
	if rule == nil {
		return nil
	}
	return &androidpublisher.AcquisitionTargetingRule{Scope: subscriptionOfferTargetingScopeToAPI(rule.Scope)}
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

func subscriptionOfferUpgradeTargetingToAPI(rule *SubscriptionOfferUpgradeTargeting) *androidpublisher.UpgradeTargetingRule {
	if rule == nil {
		return nil
	}
	apiRule := &androidpublisher.UpgradeTargetingRule{
		Scope:                 subscriptionOfferTargetingScopeToAPI(rule.Scope),
		BillingPeriodDuration: rule.BillingPeriodDuration,
		OncePerUser:           rule.OncePerUser,
	}
	if rule.OncePerUser {
		apiRule.ForceSendFields = append(apiRule.ForceSendFields, "OncePerUser")
	}
	return apiRule
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

func subscriptionOfferTargetingScopeToAPI(scope *SubscriptionOfferTargetingScope) *androidpublisher.TargetingRuleScope {
	if scope == nil {
		return nil
	}
	apiScope := &androidpublisher.TargetingRuleScope{
		SpecificSubscriptionInApp: scope.SpecificSubscriptionInApp,
	}
	if scope.AnySubscriptionInApp {
		apiScope.AnySubscriptionInApp = &androidpublisher.TargetingRuleScopeAnySubscriptionInApp{}
	}
	if scope.ThisSubscription {
		apiScope.ThisSubscription = &androidpublisher.TargetingRuleScopeThisSubscription{}
	}
	return apiScope
}

func userListResultFromAPI(developer DeveloperAccount, response *androidpublisher.ListUsersResponse) UserListResult {
	result := UserListResult{Developer: developer, Users: []User{}}
	if response == nil {
		return result
	}
	result.NextPageToken = response.NextPageToken
	for _, apiUser := range response.Users {
		if apiUser == nil {
			continue
		}
		result.Users = append(result.Users, userFromAPI(apiUser))
	}
	return result
}

func internalSharingArtifactFromAPI(apiArtifact *androidpublisher.InternalAppSharingArtifact) InternalSharingArtifact {
	if apiArtifact == nil {
		return InternalSharingArtifact{}
	}
	return InternalSharingArtifact{
		CertificateFingerprint: apiArtifact.CertificateFingerprint,
		DownloadURL:            apiArtifact.DownloadUrl,
		SHA256:                 apiArtifact.Sha256,
	}
}

func apkArtifactFromAPI(apiAPK *androidpublisher.Apk) APKArtifact {
	if apiAPK == nil {
		return APKArtifact{}
	}
	artifact := APKArtifact{VersionCode: apiAPK.VersionCode}
	if apiAPK.Binary != nil {
		artifact.SHA1 = apiAPK.Binary.Sha1
		artifact.SHA256 = apiAPK.Binary.Sha256
	}
	return artifact
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

func generatedAPKListResultFromAPI(options GeneratedAPKListOptions, response *androidpublisher.GeneratedApksListResponse) GeneratedAPKListResult {
	result := GeneratedAPKListResult{
		PackageName: options.PackageName,
		VersionCode: options.VersionCode,
		SigningKeys: []GeneratedAPKSigningKey{},
	}
	if response == nil {
		return result
	}
	for _, apiSigningKey := range response.GeneratedApks {
		if apiSigningKey == nil {
			continue
		}
		result.SigningKeys = append(result.SigningKeys, generatedAPKSigningKeyFromAPI(apiSigningKey))
	}
	return result
}

func systemAPKVariantListResultFromAPI(options SystemAPKVariantListOptions, response *androidpublisher.SystemApksListResponse) SystemAPKVariantListResult {
	result := SystemAPKVariantListResult{
		PackageName: options.PackageName,
		VersionCode: options.VersionCode,
		Variants:    []SystemAPKVariant{},
	}
	if response == nil {
		return result
	}
	for _, apiVariant := range response.Variants {
		if apiVariant == nil {
			continue
		}
		result.Variants = append(result.Variants, SystemAPKVariant{
			VariantID:  apiVariant.VariantId,
			DeviceSpec: systemAPKDeviceSpecFromAPI(apiVariant.DeviceSpec),
			Options:    systemAPKOptionsFromAPI(apiVariant.Options),
		})
	}
	return result
}

func systemAPKDeviceSpecFromAPI(apiSpec *androidpublisher.DeviceSpec) *SystemAPKDeviceSpec {
	if apiSpec == nil {
		return nil
	}
	return &SystemAPKDeviceSpec{
		ScreenDensity:    apiSpec.ScreenDensity,
		SupportedABIs:    append([]string(nil), apiSpec.SupportedAbis...),
		SupportedLocales: append([]string(nil), apiSpec.SupportedLocales...),
	}
}

func systemAPKOptionsFromAPI(apiOptions *androidpublisher.SystemApkOptions) *SystemAPKOptions {
	if apiOptions == nil {
		return nil
	}
	return &SystemAPKOptions{
		Rotated:                     apiOptions.Rotated,
		UncompressedDexFiles:        apiOptions.UncompressedDexFiles,
		UncompressedNativeLibraries: apiOptions.UncompressedNativeLibraries,
	}
}

func generatedAPKSigningKeyFromAPI(apiSigningKey *androidpublisher.GeneratedApksPerSigningKey) GeneratedAPKSigningKey {
	signingKey := GeneratedAPKSigningKey{
		CertificateSHA256Hash: apiSigningKey.CertificateSha256Hash,
		SplitAPKs:             []GeneratedSplitAPK{},
		StandaloneAPKs:        []GeneratedStandaloneAPK{},
		AssetPackSlices:       []GeneratedAssetPackSlice{},
		RecoveryModules:       []GeneratedRecoveryAPK{},
		UniversalAPK:          generatedUniversalAPKFromAPI(apiSigningKey.GeneratedUniversalApk),
	}
	if apiSigningKey.TargetingInfo != nil {
		signingKey.TargetingPackageName = apiSigningKey.TargetingInfo.PackageName
		signingKey.TargetingInfo = generatedAPKTargetingInfoFromAPI(apiSigningKey.TargetingInfo)
	}
	for _, apiAPK := range apiSigningKey.GeneratedSplitApks {
		if apiAPK == nil {
			continue
		}
		signingKey.SplitAPKs = append(signingKey.SplitAPKs, GeneratedSplitAPK{
			DownloadID: apiAPK.DownloadId,
			ModuleName: apiAPK.ModuleName,
			SplitID:    apiAPK.SplitId,
			VariantID:  apiAPK.VariantId,
		})
	}
	for _, apiAPK := range apiSigningKey.GeneratedStandaloneApks {
		if apiAPK == nil {
			continue
		}
		signingKey.StandaloneAPKs = append(signingKey.StandaloneAPKs, GeneratedStandaloneAPK{
			DownloadID: apiAPK.DownloadId,
			VariantID:  apiAPK.VariantId,
		})
	}
	for _, apiSlice := range apiSigningKey.GeneratedAssetPackSlices {
		if apiSlice == nil {
			continue
		}
		signingKey.AssetPackSlices = append(signingKey.AssetPackSlices, GeneratedAssetPackSlice{
			DownloadID: apiSlice.DownloadId,
			ModuleName: apiSlice.ModuleName,
			SliceID:    apiSlice.SliceId,
			Version:    apiSlice.Version,
		})
	}
	for _, apiRecoveryAPK := range apiSigningKey.GeneratedRecoveryModules {
		if apiRecoveryAPK == nil {
			continue
		}
		signingKey.RecoveryModules = append(signingKey.RecoveryModules, GeneratedRecoveryAPK{
			DownloadID:     apiRecoveryAPK.DownloadId,
			ModuleName:     apiRecoveryAPK.ModuleName,
			RecoveryID:     strconv.FormatInt(apiRecoveryAPK.RecoveryId, 10),
			RecoveryStatus: apiRecoveryAPK.RecoveryStatus,
		})
	}
	return signingKey
}

func generatedUniversalAPKFromAPI(apiAPK *androidpublisher.GeneratedUniversalApk) *GeneratedUniversalAPK {
	if apiAPK == nil {
		return nil
	}
	return &GeneratedUniversalAPK{DownloadID: apiAPK.DownloadId}
}

func generatedAPKTargetingInfoFromAPI(apiTargetingInfo *androidpublisher.TargetingInfo) *GeneratedAPKTargetingInfo {
	if apiTargetingInfo == nil {
		return nil
	}
	targetingInfo := &GeneratedAPKTargetingInfo{
		PackageName:    apiTargetingInfo.PackageName,
		Variants:       []GeneratedAPKTargetingVariant{},
		AssetSliceSets: []GeneratedAPKAssetSliceSet{},
	}
	for _, apiVariant := range apiTargetingInfo.Variant {
		if apiVariant == nil {
			continue
		}
		targetingInfo.Variants = append(targetingInfo.Variants, generatedAPKTargetingVariantFromAPI(apiVariant))
	}
	for _, apiAssetSliceSet := range apiTargetingInfo.AssetSliceSet {
		if apiAssetSliceSet == nil {
			continue
		}
		targetingInfo.AssetSliceSets = append(targetingInfo.AssetSliceSets, generatedAPKAssetSliceSetFromAPI(apiAssetSliceSet))
	}
	return targetingInfo
}

func generatedAPKTargetingVariantFromAPI(apiVariant *androidpublisher.SplitApkVariant) GeneratedAPKTargetingVariant {
	moduleNames := make([]string, 0, len(apiVariant.ApkSet))
	targetedAPKs := []GeneratedAPKTargetedAPK{}
	for _, apiAPKSet := range apiVariant.ApkSet {
		if apiAPKSet == nil || apiAPKSet.ModuleMetadata == nil {
			if apiAPKSet != nil {
				targetedAPKs = append(targetedAPKs, generatedAPKTargetedAPKsFromAPI(apiAPKSet.ApkDescription)...)
			}
			continue
		}
		moduleNames = append(moduleNames, apiAPKSet.ModuleMetadata.Name)
		targetedAPKs = append(targetedAPKs, generatedAPKTargetedAPKsFromAPI(apiAPKSet.ApkDescription)...)
	}
	return GeneratedAPKTargetingVariant{
		VariantNumber: apiVariant.VariantNumber,
		ModuleNames:   moduleNames,
		Targeting:     generatedAPKRawJSONFromAPI(apiVariant.Targeting),
		APKs:          targetedAPKs,
	}
}

func generatedAPKTargetedAPKsFromAPI(apiAPKDescriptions []*androidpublisher.ApkDescription) []GeneratedAPKTargetedAPK {
	targetedAPKs := make([]GeneratedAPKTargetedAPK, 0, len(apiAPKDescriptions))
	for _, apiAPKDescription := range apiAPKDescriptions {
		if apiAPKDescription == nil {
			continue
		}
		targetedAPKs = append(targetedAPKs, GeneratedAPKTargetedAPK{
			Path:      apiAPKDescription.Path,
			Targeting: generatedAPKRawJSONFromAPI(apiAPKDescription.Targeting),
		})
	}
	return targetedAPKs
}

func generatedAPKAssetSliceSetFromAPI(apiAssetSliceSet *androidpublisher.AssetSliceSet) GeneratedAPKAssetSliceSet {
	assetSliceSet := GeneratedAPKAssetSliceSet{APKPaths: []string{}}
	if apiAssetSliceSet.AssetModuleMetadata != nil {
		assetSliceSet.ModuleName = apiAssetSliceSet.AssetModuleMetadata.Name
		assetSliceSet.DeliveryType = apiAssetSliceSet.AssetModuleMetadata.DeliveryType
	}
	for _, apiAPKDescription := range apiAssetSliceSet.ApkDescription {
		if apiAPKDescription == nil || apiAPKDescription.Path == "" {
			continue
		}
		assetSliceSet.APKPaths = append(assetSliceSet.APKPaths, apiAPKDescription.Path)
	}
	return assetSliceSet
}

func generatedAPKRawJSONFromAPI(value any) json.RawMessage {
	if value == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	if string(payload) == "{}" {
		return nil
	}
	return payload
}

func regionPriceConversionResultFromAPI(options RegionPriceConversionOptions, response *androidpublisher.ConvertRegionPricesResponse) RegionPriceConversionResult {
	result := RegionPriceConversionResult{
		PackageName: options.PackageName,
		SourcePrice: Money{
			CurrencyCode: options.Currency.String(),
			Units:        options.Units,
			Nanos:        options.Nanos,
		},
		ConvertedRegionPrices: map[string]ConvertedRegionPrice{},
	}
	if response == nil {
		return result
	}
	if response.RegionVersion != nil {
		result.RegionVersion = response.RegionVersion.Version
	}
	result.ConvertedOtherRegionsPrice = convertedOtherRegionsPriceFromAPI(response.ConvertedOtherRegionsPrice)
	for regionCode, apiPrice := range response.ConvertedRegionPrices {
		result.ConvertedRegionPrices[regionCode] = convertedRegionPriceFromAPI(apiPrice)
	}
	return result
}

func convertedOtherRegionsPriceFromAPI(apiPrice *androidpublisher.ConvertedOtherRegionsPrice) *ConvertedOtherRegionsPrice {
	if apiPrice == nil {
		return nil
	}
	return &ConvertedOtherRegionsPrice{
		USDPrice: moneyFromAPI(apiPrice.UsdPrice),
		EURPrice: moneyFromAPI(apiPrice.EurPrice),
	}
}

func convertedRegionPriceFromAPI(apiPrice androidpublisher.ConvertedRegionPrice) ConvertedRegionPrice {
	return ConvertedRegionPrice{
		RegionCode: apiPrice.RegionCode,
		Price:      moneyFromAPI(apiPrice.Price),
		TaxAmount:  moneyFromAPI(apiPrice.TaxAmount),
	}
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

func deviceUserCountrySetsFromAPI(apiSets []*androidpublisher.UserCountrySet) []DeviceUserCountrySet {
	sets := make([]DeviceUserCountrySet, 0, len(apiSets))
	for _, apiSet := range apiSets {
		if apiSet == nil {
			continue
		}
		sets = append(sets, DeviceUserCountrySet{
			Name:         apiSet.Name,
			CountryCodes: append([]string(nil), apiSet.CountryCodes...),
		})
	}
	return sets
}

func orderBatchGetResultFromAPI(options OrderBatchGetOptions, response *androidpublisher.BatchGetOrdersResponse) OrderBatchGetResult {
	result := OrderBatchGetResult{
		PackageName: options.PackageName,
		OrderIDs:    append([]OrderID(nil), options.OrderIDs...),
		Orders:      []Order{},
	}
	if response == nil {
		return result
	}
	for _, apiOrder := range response.Orders {
		if apiOrder == nil {
			continue
		}
		result.Orders = append(result.Orders, orderFromAPI(apiOrder))
	}
	return result
}

func orderFromAPI(apiOrder *androidpublisher.Order) Order {
	if apiOrder == nil {
		return Order{LineItems: []OrderLineItem{}}
	}
	lineItems := make([]OrderLineItem, 0, len(apiOrder.LineItems))
	for _, apiLineItem := range apiOrder.LineItems {
		if apiLineItem == nil {
			continue
		}
		lineItems = append(lineItems, orderLineItemFromAPI(apiLineItem))
	}
	return Order{
		OrderID:                         apiOrder.OrderId,
		PurchaseToken:                   apiOrder.PurchaseToken,
		State:                           apiOrder.State,
		CreateTime:                      apiOrder.CreateTime,
		LastEventTime:                   apiOrder.LastEventTime,
		BuyerAddress:                    buyerAddressFromAPI(apiOrder.BuyerAddress),
		Total:                           moneyFromAPI(apiOrder.Total),
		Tax:                             moneyFromAPI(apiOrder.Tax),
		DeveloperRevenueInBuyerCurrency: moneyFromAPI(apiOrder.DeveloperRevenueInBuyerCurrency),
		OrderDetails:                    orderDetailsFromAPI(apiOrder.OrderDetails),
		OrderHistory:                    orderHistoryFromAPI(apiOrder.OrderHistory),
		PointsDetails:                   pointsDetailsFromAPI(apiOrder.PointsDetails),
		LineItems:                       lineItems,
	}
}

func orderLineItemFromAPI(apiLineItem *androidpublisher.LineItem) OrderLineItem {
	return OrderLineItem{
		ProductID:              apiLineItem.ProductId,
		ProductTitle:           apiLineItem.ProductTitle,
		ListingPrice:           moneyFromAPI(apiLineItem.ListingPrice),
		Tax:                    moneyFromAPI(apiLineItem.Tax),
		Total:                  moneyFromAPI(apiLineItem.Total),
		OneTimePurchaseDetails: oneTimePurchaseDetailsFromAPI(apiLineItem.OneTimePurchaseDetails),
		PaidAppDetails:         paidAppDetailsFromAPI(apiLineItem.PaidAppDetails),
		SubscriptionDetails:    orderSubscriptionDetailsFromAPI(apiLineItem.SubscriptionDetails),
	}
}

func oneTimePurchaseDetailsFromAPI(apiDetails *androidpublisher.OneTimePurchaseDetails) *OneTimePurchaseDetails {
	if apiDetails == nil {
		return nil
	}
	return &OneTimePurchaseDetails{
		OfferID:  apiDetails.OfferId,
		Quantity: apiDetails.Quantity,
	}
}

func paidAppDetailsFromAPI(apiDetails *androidpublisher.PaidAppDetails) *PaidAppDetails {
	if apiDetails == nil {
		return nil
	}
	return &PaidAppDetails{}
}

func orderSubscriptionDetailsFromAPI(apiDetails *androidpublisher.SubscriptionDetails) *OrderSubscriptionDetails {
	if apiDetails == nil {
		return nil
	}
	return &OrderSubscriptionDetails{
		BasePlanID:             apiDetails.BasePlanId,
		OfferID:                apiDetails.OfferId,
		OfferPhase:             apiDetails.OfferPhase,
		ServicePeriodStartTime: apiDetails.ServicePeriodStartTime,
		ServicePeriodEndTime:   apiDetails.ServicePeriodEndTime,
	}
}

func orderDetailsFromAPI(apiDetails *androidpublisher.OrderDetails) *OrderDetails {
	if apiDetails == nil {
		return nil
	}
	return &OrderDetails{TaxInclusive: apiDetails.TaxInclusive}
}

func orderHistoryFromAPI(apiHistory *androidpublisher.OrderHistory) *OrderHistory {
	if apiHistory == nil {
		return nil
	}
	partialRefundEvents := make([]PartialRefundEvent, 0, len(apiHistory.PartialRefundEvents))
	for _, apiEvent := range apiHistory.PartialRefundEvents {
		if apiEvent == nil {
			continue
		}
		partialRefundEvents = append(partialRefundEvents, partialRefundEventFromAPI(apiEvent))
	}
	return &OrderHistory{
		CancellationEvent:   cancellationEventFromAPI(apiHistory.CancellationEvent),
		ProcessedEvent:      processedEventFromAPI(apiHistory.ProcessedEvent),
		RefundEvent:         orderRefundEventFromAPI(apiHistory.RefundEvent),
		PartialRefundEvents: partialRefundEvents,
	}
}

func cancellationEventFromAPI(apiEvent *androidpublisher.CancellationEvent) *OrderEvent {
	if apiEvent == nil {
		return nil
	}
	return &OrderEvent{EventTime: apiEvent.EventTime}
}

func processedEventFromAPI(apiEvent *androidpublisher.ProcessedEvent) *OrderEvent {
	if apiEvent == nil {
		return nil
	}
	return &OrderEvent{EventTime: apiEvent.EventTime}
}

func orderRefundEventFromAPI(apiEvent *androidpublisher.RefundEvent) *OrderRefundEvent {
	if apiEvent == nil {
		return nil
	}
	return &OrderRefundEvent{
		EventTime:     apiEvent.EventTime,
		RefundReason:  apiEvent.RefundReason,
		RefundDetails: refundDetailsFromAPI(apiEvent.RefundDetails),
	}
}

func partialRefundEventFromAPI(apiEvent *androidpublisher.PartialRefundEvent) PartialRefundEvent {
	return PartialRefundEvent{
		CreateTime:    apiEvent.CreateTime,
		ProcessTime:   apiEvent.ProcessTime,
		State:         apiEvent.State,
		RefundDetails: refundDetailsFromAPI(apiEvent.RefundDetails),
	}
}

func refundDetailsFromAPI(apiDetails *androidpublisher.RefundDetails) *RefundDetails {
	if apiDetails == nil {
		return nil
	}
	return &RefundDetails{
		Tax:   moneyFromAPI(apiDetails.Tax),
		Total: moneyFromAPI(apiDetails.Total),
	}
}

func pointsDetailsFromAPI(apiDetails *androidpublisher.PointsDetails) *PointsDetails {
	if apiDetails == nil {
		return nil
	}
	return &PointsDetails{
		PointsOfferID:            apiDetails.PointsOfferId,
		PointsSpent:              apiDetails.PointsSpent,
		PointsDiscountRateMicros: apiDetails.PointsDiscountRateMicros,
		PointsCouponValue:        moneyFromAPI(apiDetails.PointsCouponValue),
	}
}

func buyerAddressFromAPI(apiAddress *androidpublisher.BuyerAddress) *BuyerAddress {
	if apiAddress == nil {
		return nil
	}
	return &BuyerAddress{
		Country:  apiAddress.BuyerCountry,
		State:    apiAddress.BuyerState,
		Postcode: apiAddress.BuyerPostcode,
	}
}

func orderIDStrings(orderIDs []OrderID) []string {
	values := make([]string, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		values = append(values, orderID.String())
	}
	return values
}

func userFromAPI(apiUser *androidpublisher.User) User {
	if apiUser == nil {
		return User{Grants: []UserGrant{}}
	}
	return User{
		Name:                        apiUser.Name,
		Email:                       apiUser.Email,
		AccessState:                 apiUser.AccessState,
		ExpirationTime:              apiUser.ExpirationTime,
		DeveloperAccountPermissions: userPermissionsFromAPI(apiUser.DeveloperAccountPermissions),
		Partial:                     apiUser.Partial,
		Grants:                      userGrantsFromAPI(apiUser.Grants),
	}
}

func userCreateToAPI(options UserCreateOptions) *androidpublisher.User {
	return &androidpublisher.User{
		Name:                        options.UserName().String(),
		Email:                       options.UserEmail.String(),
		DeveloperAccountPermissions: userPermissionStrings(options.Permissions),
		ExpirationTime:              options.ExpirationTime,
	}
}

func userPatchToAPI(options UserPatchOptions) *androidpublisher.User {
	apiUser := &androidpublisher.User{
		Name:           options.Name.String(),
		ExpirationTime: options.ExpirationTime,
	}
	if len(options.Permissions) > 0 {
		apiUser.DeveloperAccountPermissions = userPermissionStrings(options.Permissions)
	}
	return apiUser
}

func userPermissionsFromAPI(apiPermissions []string) []UserPermission {
	permissions := make([]UserPermission, 0, len(apiPermissions))
	for _, apiPermission := range apiPermissions {
		permissions = append(permissions, UserPermission(apiPermission))
	}
	return permissions
}

func userPermissionStrings(permissions []UserPermission) []string {
	values := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		values = append(values, permission.String())
	}
	return values
}

func userGrantsFromAPI(apiGrants []*androidpublisher.Grant) []UserGrant {
	grants := make([]UserGrant, 0, len(apiGrants))
	for _, apiGrant := range apiGrants {
		if apiGrant == nil {
			continue
		}
		grants = append(grants, UserGrant{
			Name:                apiGrant.Name,
			PackageName:         apiGrant.PackageName,
			AppLevelPermissions: apiGrant.AppLevelPermissions,
		})
	}
	return grants
}

func grantFromAPI(apiGrant *androidpublisher.Grant) Grant {
	if apiGrant == nil {
		return Grant{Permissions: []GrantPermission{}}
	}
	permissions := make([]GrantPermission, 0, len(apiGrant.AppLevelPermissions))
	for _, permission := range apiGrant.AppLevelPermissions {
		permissions = append(permissions, GrantPermission(permission))
	}
	return Grant{
		Name:        GrantName(apiGrant.Name),
		PackageName: PackageName(apiGrant.PackageName),
		Permissions: permissions,
	}
}

func grantToAPI(grant Grant) *androidpublisher.Grant {
	apiGrant := &androidpublisher.Grant{
		Name:        grant.Name.String(),
		PackageName: grant.PackageName.String(),
	}
	for _, permission := range grant.Permissions {
		apiGrant.AppLevelPermissions = append(apiGrant.AppLevelPermissions, permission.String())
	}
	apiGrant.ForceSendFields = append(apiGrant.ForceSendFields, "AppLevelPermissions")
	if grant.Name != "" {
		apiGrant.ForceSendFields = append(apiGrant.ForceSendFields, "Name")
	}
	if grant.PackageName != "" {
		apiGrant.ForceSendFields = append(apiGrant.ForceSendFields, "PackageName")
	}
	return apiGrant
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

func subscriptionRevokeRequestToAPI(options SubscriptionPurchaseRevokeOptions) *androidpublisher.RevokeSubscriptionPurchaseRequest {
	context := &androidpublisher.RevocationContext{}
	switch options.RefundType {
	case SubscriptionRefundTypeFull:
		context.FullRefund = &androidpublisher.RevocationContextFullRefund{}
	case SubscriptionRefundTypeProrated:
		context.ProratedRefund = &androidpublisher.RevocationContextProratedRefund{}
	case SubscriptionRefundTypeItem:
		context.ItemBasedRefund = &androidpublisher.RevocationContextItemBasedRefund{
			ProductId: options.RefundProductID.String(),
		}
	}
	return &androidpublisher.RevokeSubscriptionPurchaseRequest{
		RevocationContext: context,
	}
}

type rawSubscriptionCancelRequest struct {
	CancellationContext rawSubscriptionCancellationContext `json:"cancellationContext"`
}

type rawSubscriptionCancellationContext struct {
	CancellationType string `json:"cancellationType"`
}

func subscriptionCancelRequestToAPI(options SubscriptionPurchaseMutationOptions) rawSubscriptionCancelRequest {
	return rawSubscriptionCancelRequest{
		CancellationContext: rawSubscriptionCancellationContext{
			CancellationType: subscriptionCancellationTypeToAPI(options.CancellationType),
		},
	}
}

func subscriptionCancellationTypeToAPI(cancellationType SubscriptionCancellationType) string {
	switch cancellationType {
	case SubscriptionCancellationTypeUserRequestedStopRenewals:
		return "USER_REQUESTED_STOP_RENEWALS"
	case SubscriptionCancellationTypeDeveloperRequestedStopPayments:
		return "DEVELOPER_REQUESTED_STOP_PAYMENTS"
	default:
		return "CANCELLATION_TYPE_UNSPECIFIED"
	}
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

func voidedPurchaseListResultFromAPI(options VoidedPurchaseListOptions, response *androidpublisher.VoidedPurchasesListResponse) VoidedPurchaseListResult {
	result := VoidedPurchaseListResult{
		PackageName: options.PackageName,
		Options:     options,
		Purchases:   []VoidedPurchase{},
	}
	if response == nil {
		return result
	}
	for _, apiPurchase := range response.VoidedPurchases {
		if apiPurchase == nil {
			continue
		}
		result.Purchases = append(result.Purchases, voidedPurchaseFromAPI(apiPurchase))
	}
	if response.PageInfo != nil {
		result.PageInfo = &VoidedPurchasePageInfo{
			ResultPerPage: response.PageInfo.ResultPerPage,
			StartIndex:    response.PageInfo.StartIndex,
			TotalResults:  response.PageInfo.TotalResults,
		}
	}
	if response.TokenPagination != nil {
		result.Pagination = &VoidedPurchasePagination{
			NextPageToken:     response.TokenPagination.NextPageToken,
			PreviousPageToken: response.TokenPagination.PreviousPageToken,
		}
	}
	return result
}

func voidedPurchaseFromAPI(apiPurchase *androidpublisher.VoidedPurchase) VoidedPurchase {
	return VoidedPurchase{
		OrderID:            apiPurchase.OrderId,
		PurchaseToken:      apiPurchase.PurchaseToken,
		PurchaseTimeMillis: apiPurchase.PurchaseTimeMillis,
		VoidedTimeMillis:   apiPurchase.VoidedTimeMillis,
		VoidedReason:       apiPurchase.VoidedReason,
		VoidedSource:       apiPurchase.VoidedSource,
		VoidedQuantity:     apiPurchase.VoidedQuantity,
	}
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
