package play

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

func (p *GooglePublisher) doJSON(req *http.Request, target any) error {
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

func (p *GooglePublisher) doNoContent(req *http.Request) error {
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

func (p *GooglePublisher) InsertEdit(ctx context.Context, packageName PackageName) (Edit, error) {
	edit, err := p.service.Edits.Insert(packageName.String(), &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return Edit{}, fmt.Errorf("insert edit for %s: %w", packageName, err)
	}
	return Edit{ID: edit.Id, ExpiryTimeSeconds: edit.ExpiryTimeSeconds}, nil
}

func (p *GooglePublisher) UploadBundle(ctx context.Context, packageName PackageName, editID string, bundlePath string) (BundleArtifact, error) {
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

func (p *GooglePublisher) UploadAPK(ctx context.Context, packageName PackageName, editID string, apkPath string) (APKArtifact, error) {
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

func (p *GooglePublisher) ValidateEdit(ctx context.Context, packageName PackageName, editID string) error {
	if _, err := p.service.Edits.Validate(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("validate edit %s for %s: %w", editID, packageName, err)
	}
	return nil
}

type CommitEditOptions struct {
	ChangesNotSentForReview bool
}

func (p *GooglePublisher) CommitEdit(ctx context.Context, packageName PackageName, editID string) (Edit, error) {
	return p.CommitEditWithOptions(ctx, packageName, editID, CommitEditOptions{})
}

func (p *GooglePublisher) CommitEditWithOptions(ctx context.Context, packageName PackageName, editID string, opts CommitEditOptions) (Edit, error) {
	call := p.service.Edits.Commit(packageName.String(), editID).Context(ctx)
	if opts.ChangesNotSentForReview {
		call = call.ChangesNotSentForReview(true)
	}
	edit, err := call.Do()
	if err != nil {
		return Edit{}, fmt.Errorf("commit edit %s for %s: %w", editID, packageName, err)
	}
	return Edit{ID: edit.Id, ExpiryTimeSeconds: edit.ExpiryTimeSeconds}, nil
}

func (p *GooglePublisher) DeleteEdit(ctx context.Context, packageName PackageName, editID string) error {
	if err := p.service.Edits.Delete(packageName.String(), editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete edit %s for %s: %w", editID, packageName, err)
	}
	return nil
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
