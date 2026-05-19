package play

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func (p *GooglePublisher) UploadInternalSharingAPK(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error) {
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

func (p *GooglePublisher) UploadInternalSharingBundle(ctx context.Context, packageName PackageName, path string) (InternalSharingArtifact, error) {
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
