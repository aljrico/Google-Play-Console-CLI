package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListUsers(ctx context.Context, options UserListOptions) (UserListResult, error) {
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

func (p *GooglePublisher) CreateUser(ctx context.Context, options UserCreateOptions) (User, error) {
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

func (p *GooglePublisher) PatchUser(ctx context.Context, options UserPatchOptions) (User, error) {
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

func (p *GooglePublisher) DeleteUser(ctx context.Context, options UserDeleteOptions) error {
	if err := options.ValidateLive(); err != nil {
		return err
	}
	if err := p.service.Users.Delete(options.Name.String()).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete user %s: %w", options.Name, err)
	}
	return nil
}

func userFractionFromAPI(apiRelease *androidpublisher.TrackRelease) *float64 {
	if apiRelease.UserFraction == 0 {
		return nil
	}
	userFraction := apiRelease.UserFraction
	return &userFraction
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
