package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) CreateGrant(ctx context.Context, options GrantCreateOptions) (Grant, error) {
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

func (p *GooglePublisher) PatchGrant(ctx context.Context, options GrantPatchOptions) (Grant, error) {
	apiGrant, err := p.service.Grants.Patch(options.Name.String(), grantToAPI(Grant{
		Name:        options.Name,
		Permissions: options.Permissions,
	})).UpdateMask("appLevelPermissions").Context(ctx).Do()
	if err != nil {
		return Grant{}, fmt.Errorf("patch grant %s: %w", options.Name, err)
	}
	return grantFromAPI(apiGrant), nil
}

func (p *GooglePublisher) DeleteGrant(ctx context.Context, options GrantDeleteOptions) error {
	if err := p.service.Grants.Delete(options.Name.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete grant %s: %w", options.Name, err)
	}
	return nil
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
