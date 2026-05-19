package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) UpdateDataSafety(ctx context.Context, packageName PackageName, safetyLabels string) error {
	request := &androidpublisher.SafetyLabelsUpdateRequest{
		SafetyLabels: safetyLabels,
	}
	if _, err := p.service.Applications.DataSafety(packageName.String(), request).Context(ctx).Do(); err != nil {
		return fmt.Errorf("update data safety for %s: %w", packageName, err)
	}
	return nil
}
