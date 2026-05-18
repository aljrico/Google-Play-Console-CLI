package play

import (
	"context"
	"fmt"
	"strings"
)

type GrantName string

func NewGrantName(value string) (GrantName, error) {
	if value == "" {
		return "", fmt.Errorf("grant name is required")
	}
	segments := strings.Split(value, "/")
	if len(segments) != 6 || segments[0] != "developers" || segments[2] != "users" || segments[4] != "grants" {
		return "", fmt.Errorf("invalid grant name %q", value)
	}
	if _, err := NewDeveloperAccount(segments[1]); err != nil {
		return "", fmt.Errorf("invalid grant developer: %w", err)
	}
	if _, err := NewGrantUserEmail(segments[3]); err != nil {
		return "", fmt.Errorf("invalid grant user: %w", err)
	}
	if segments[5] == "" {
		return "", fmt.Errorf("invalid grant package %q", segments[5])
	}
	return GrantName(value), nil
}

func (g GrantName) String() string {
	return string(g)
}

func (g GrantName) Validate() error {
	_, err := NewGrantName(g.String())
	return err
}

type GrantUserEmail string

func NewGrantUserEmail(value string) (GrantUserEmail, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("user email is required")
	}
	if strings.Contains(value, "/") || !strings.Contains(value, "@") {
		return "", fmt.Errorf("invalid user email %q", value)
	}
	return GrantUserEmail(value), nil
}

func (e GrantUserEmail) String() string {
	return string(e)
}

func (e GrantUserEmail) Validate() error {
	_, err := NewGrantUserEmail(e.String())
	return err
}

type GrantPermission string

const (
	GrantPermissionAccessApp            GrantPermission = "CAN_ACCESS_APP"
	GrantPermissionViewFinancialData    GrantPermission = "CAN_VIEW_FINANCIAL_DATA"
	GrantPermissionManagePermissions    GrantPermission = "CAN_MANAGE_PERMISSIONS"
	GrantPermissionReplyToReviews       GrantPermission = "CAN_REPLY_TO_REVIEWS"
	GrantPermissionManagePublicAPKs     GrantPermission = "CAN_MANAGE_PUBLIC_APKS"
	GrantPermissionManageTrackAPKs      GrantPermission = "CAN_MANAGE_TRACK_APKS"
	GrantPermissionManageTrackUsers     GrantPermission = "CAN_MANAGE_TRACK_USERS"
	GrantPermissionManagePublicListing  GrantPermission = "CAN_MANAGE_PUBLIC_LISTING"
	GrantPermissionManageDraftApps      GrantPermission = "CAN_MANAGE_DRAFT_APPS"
	GrantPermissionManageOrders         GrantPermission = "CAN_MANAGE_ORDERS"
	GrantPermissionManageAppContent     GrantPermission = "CAN_MANAGE_APP_CONTENT"
	GrantPermissionViewNonFinancialData GrantPermission = "CAN_VIEW_NON_FINANCIAL_DATA"
	GrantPermissionViewAppQuality       GrantPermission = "CAN_VIEW_APP_QUALITY"
	GrantPermissionManageDeepLinks      GrantPermission = "CAN_MANAGE_DEEPLINKS"
)

func NewGrantPermission(value string) (GrantPermission, error) {
	permission := GrantPermission(strings.TrimSpace(value))
	switch permission {
	case GrantPermissionAccessApp,
		GrantPermissionViewFinancialData,
		GrantPermissionManagePermissions,
		GrantPermissionReplyToReviews,
		GrantPermissionManagePublicAPKs,
		GrantPermissionManageTrackAPKs,
		GrantPermissionManageTrackUsers,
		GrantPermissionManagePublicListing,
		GrantPermissionManageDraftApps,
		GrantPermissionManageOrders,
		GrantPermissionManageAppContent,
		GrantPermissionViewNonFinancialData,
		GrantPermissionViewAppQuality,
		GrantPermissionManageDeepLinks:
		return permission, nil
	default:
		return "", fmt.Errorf("unsupported grant permission %q", value)
	}
}

func (p GrantPermission) String() string {
	return string(p)
}

func (p GrantPermission) Validate() error {
	_, err := NewGrantPermission(p.String())
	return err
}

type Grant struct {
	Name        GrantName         `json:"name,omitempty"`
	PackageName PackageName       `json:"packageName,omitempty"`
	Permissions []GrantPermission `json:"appLevelPermissions"`
}

type GrantCreateOptions struct {
	Developer   DeveloperAccount  `json:"developer"`
	UserEmail   GrantUserEmail    `json:"userEmail"`
	PackageName PackageName       `json:"packageName"`
	Permissions []GrantPermission `json:"appLevelPermissions"`
	Confirm     bool              `json:"confirm"`
	DryRun      bool              `json:"dryRun"`
}

func (o GrantCreateOptions) Validate() error {
	if err := o.Developer.Validate(); err != nil {
		return err
	}
	if err := o.UserEmail.Validate(); err != nil {
		return err
	}
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	return validateGrantMutation(o.Permissions, o.Confirm, o.DryRun)
}

func (o GrantCreateOptions) Parent() string {
	return o.Developer.ResourceName() + "/users/" + o.UserEmail.String()
}

func (o GrantCreateOptions) GrantName() GrantName {
	return GrantName(o.Parent() + "/grants/" + o.PackageName.String())
}

type GrantPatchOptions struct {
	Name        GrantName         `json:"name"`
	Permissions []GrantPermission `json:"appLevelPermissions"`
	Confirm     bool              `json:"confirm"`
	DryRun      bool              `json:"dryRun"`
}

func (o GrantPatchOptions) Validate() error {
	if err := o.Name.Validate(); err != nil {
		return err
	}
	return validateGrantMutation(o.Permissions, o.Confirm, o.DryRun)
}

type GrantDeleteOptions struct {
	Name    GrantName `json:"name"`
	Confirm bool      `json:"confirm"`
	DryRun  bool      `json:"dryRun"`
}

func (o GrantDeleteOptions) Validate() error {
	if err := o.Name.Validate(); err != nil {
		return err
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("grant deletion requires --confirm or --dry-run")
	}
	return nil
}

func validateGrantMutation(permissions []GrantPermission, confirm bool, dryRun bool) error {
	if len(permissions) == 0 {
		return fmt.Errorf("at least one grant permission is required")
	}
	seen := make(map[GrantPermission]struct{}, len(permissions))
	for _, permission := range permissions {
		if err := permission.Validate(); err != nil {
			return err
		}
		if _, ok := seen[permission]; ok {
			return fmt.Errorf("grant permission %q is duplicated", permission)
		}
		seen[permission] = struct{}{}
	}
	if !dryRun && !confirm {
		return fmt.Errorf("grant update requires --confirm or --dry-run")
	}
	return nil
}

type GrantMutationPlan struct {
	Action      string            `json:"action"`
	Target      string            `json:"target"`
	Permissions []GrantPermission `json:"appLevelPermissions,omitempty"`
	Grant       *Grant            `json:"grant,omitempty"`
	Confirm     bool              `json:"confirm"`
	Steps       []string          `json:"steps"`
}

type GrantMutationResult struct {
	Action  string            `json:"action"`
	DryRun  bool              `json:"dryRun"`
	Applied bool              `json:"applied"`
	Grant   *Grant            `json:"grant,omitempty"`
	Desired *Grant            `json:"desiredGrant,omitempty"`
	Plan    GrantMutationPlan `json:"plan"`
}

type GrantManager interface {
	CreateGrant(ctx context.Context, options GrantCreateOptions) (Grant, error)
	PatchGrant(ctx context.Context, options GrantPatchOptions) (Grant, error)
	DeleteGrant(ctx context.Context, options GrantDeleteOptions) error
}

func CreateGrant(ctx context.Context, manager GrantManager, options GrantCreateOptions) (GrantMutationResult, error) {
	if err := options.Validate(); err != nil {
		return GrantMutationResult{}, err
	}
	desiredGrant := Grant{
		Name:        options.GrantName(),
		PackageName: options.PackageName,
		Permissions: options.Permissions,
	}
	result := GrantMutationResult{
		Action:  "create",
		DryRun:  options.DryRun,
		Desired: &desiredGrant,
		Plan: GrantMutationPlan{
			Action:      "create",
			Target:      options.GrantName().String(),
			Permissions: append([]GrantPermission(nil), options.Permissions...),
			Grant:       &desiredGrant,
			Confirm:     options.Confirm,
			Steps:       grantSteps("create", options.DryRun),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if manager == nil {
		return GrantMutationResult{}, fmt.Errorf("grant manager is required")
	}
	grant, err := manager.CreateGrant(ctx, options)
	if err != nil {
		return GrantMutationResult{}, err
	}
	result.Applied = true
	result.Grant = &grant
	return result, nil
}

func PatchGrant(ctx context.Context, manager GrantManager, options GrantPatchOptions) (GrantMutationResult, error) {
	if err := options.Validate(); err != nil {
		return GrantMutationResult{}, err
	}
	desiredGrant := Grant{
		Name:        options.Name,
		Permissions: options.Permissions,
	}
	result := GrantMutationResult{
		Action:  "patch",
		DryRun:  options.DryRun,
		Desired: &desiredGrant,
		Plan: GrantMutationPlan{
			Action:      "patch",
			Target:      options.Name.String(),
			Permissions: append([]GrantPermission(nil), options.Permissions...),
			Grant:       &desiredGrant,
			Confirm:     options.Confirm,
			Steps:       grantSteps("patch", options.DryRun),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if manager == nil {
		return GrantMutationResult{}, fmt.Errorf("grant manager is required")
	}
	grant, err := manager.PatchGrant(ctx, options)
	if err != nil {
		return GrantMutationResult{}, err
	}
	result.Applied = true
	result.Grant = &grant
	return result, nil
}

func DeleteGrant(ctx context.Context, manager GrantManager, options GrantDeleteOptions) (GrantMutationResult, error) {
	if err := options.Validate(); err != nil {
		return GrantMutationResult{}, err
	}
	result := GrantMutationResult{
		Action: "delete",
		DryRun: options.DryRun,
		Plan: GrantMutationPlan{
			Action:  "delete",
			Target:  options.Name.String(),
			Confirm: options.Confirm,
			Steps:   grantSteps("delete", options.DryRun),
		},
	}
	if options.DryRun {
		return result, nil
	}
	if manager == nil {
		return GrantMutationResult{}, fmt.Errorf("grant manager is required")
	}
	if err := manager.DeleteGrant(ctx, options); err != nil {
		return GrantMutationResult{}, err
	}
	result.Applied = true
	return result, nil
}

func grantSteps(action string, dryRun bool) []string {
	if dryRun {
		return []string{fmt.Sprintf("plan grant %s", action)}
	}
	return []string{fmt.Sprintf("%s grant", action)}
}
