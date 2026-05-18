package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersionJSON(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "--output", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"version"`) {
		t.Fatalf("version output = %s", buf.String())
	}
}

func TestUnknownOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "--output", "yaml"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestVersionRejectsUnexpectedArgs(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{"version", "stray"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCapabilitiesOutputsParityMatrixWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"capabilities",
		"--status",
		"tested",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"status":"tested"`) {
		t.Fatalf("output = %s, want tested status", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestCapabilitiesRejectsUnsupportedStatus(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"capabilities",
		"--status",
		"done-ish",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected unsupported status error")
	}
	if !strings.Contains(err.Error(), "unsupported capability status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestDocsParityOutputsMarkdownWithoutAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"parity",
		"--output",
		"markdown",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "# Parity Matrix") {
		t.Fatalf("output = %s, want parity matrix markdown", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestDocsParityOutputsJSONDocument(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"docs",
		"parity",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{`"name":"parity"`, `"format":"markdown"`, `# Parity Matrix`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

func TestInitDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"init",
		"--directory",
		t.TempDir() + "/.gpc",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"dryRun":true`) {
		t.Fatalf("output = %s, want dryRun true", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestPublishInternalDryRunRejectsInvalidPackage(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"bad",
		"--aab",
		"app-release.aab",
		"--dry-run",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestStatusRejectsInvalidPackageBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"status",
		"--package",
		"bad",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected package validation error")
	}
	if !strings.Contains(err.Error(), "invalid Android package name") {
		t.Fatalf("error = %v, want package validation", err)
	}
}

func TestValidateRejectsInvalidPackageBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"validate",
		"--package",
		"bad",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected package validation error")
	}
	if !strings.Contains(err.Error(), "invalid Android package name") {
		t.Fatalf("error = %v, want package validation", err)
	}
}

func TestUsersListRejectsMissingDeveloperBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"list",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected developer validation error")
	}
	if !strings.Contains(err.Error(), "developer account") {
		t.Fatalf("error = %v, want developer validation", err)
	}
}

func TestUsersListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"users",
		"list",
		"--developer",
		"1234567890",
		"--page-size",
		"-2",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestGrantsCreateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"grants",
		"create",
		"--developer",
		"1234567890",
		"--user-email",
		"user@example.com",
		"--package",
		"com.example.app",
		"--permission",
		"CAN_VIEW_NON_FINANCIAL_DATA",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("output = %s, want dryRun true", buf.String())
	}
	if !strings.Contains(buf.String(), `"target":"developers/1234567890/users/user@example.com/grants/com.example.app"`) {
		t.Fatalf("output = %s, want full grant target", buf.String())
	}
	if !strings.Contains(buf.String(), `"appLevelPermissions":["CAN_VIEW_NON_FINANCIAL_DATA"]`) {
		t.Fatalf("output = %s, want grant permission preview", buf.String())
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestGrantsPatchRequiresConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"grants",
		"patch",
		"--name",
		"developers/123/users/user@example.com/grants/com.example.app",
		"--permission",
		"CAN_REPLY_TO_REVIEWS",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "--confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestInternalSharingUploadDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"internal-sharing",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"kind":"bundle"`) || !strings.Contains(output, `"dryRun":true`) {
		t.Fatalf("output = %s, want bundle dry run", output)
	}
}

func TestInternalSharingUploadRejectsMissingArtifactBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"internal-sharing",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		t.TempDir() + "/missing.apk",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing artifact error")
	}
	if !strings.Contains(err.Error(), "open file") {
		t.Fatalf("error = %v, want local file preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestInternalSharingUploadRejectsDirectoryArtifactBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	artifactPath := t.TempDir() + "/directory.apk"
	if err := os.Mkdir(artifactPath, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"internal-sharing",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		artifactPath,
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected directory artifact error")
	}
	if !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("error = %v, want regular file validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestAppRecoveryListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"app-recovery",
		"list",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected version code validation error")
	}
	if !strings.Contains(err.Error(), "version code") {
		t.Fatalf("error = %v, want version code validation", err)
	}
}

func TestGeneratedAPKsListRejectsMissingVersionCodeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"generated-apks",
		"list",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected version code validation error")
	}
	if !strings.Contains(err.Error(), "version code") {
		t.Fatalf("error = %v, want version code validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestDeviceTierConfigsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"device-tier-configs",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"101",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestDeviceTierConfigsGetRejectsMissingIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"device-tier-configs",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected device tier config ID validation error")
	}
	if !strings.Contains(err.Error(), "device tier config ID") {
		t.Fatalf("error = %v, want device tier config ID validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPublishInternalDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--release-note",
		"en-US=Bug fixes.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("publish dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"releaseNotes":[{"language":"en-US","text":"Bug fixes."}]`) {
		t.Fatalf("publish dry-run output = %s, want release note", buf.String())
	}
}

func TestPublishInternalLiveRejectsMissingBundleBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"com.example.app",
		"--aab",
		t.TempDir() + "/missing.aab",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing bundle error")
	}
	if !strings.Contains(err.Error(), "open bundle") {
		t.Fatalf("error = %v, want bundle preflight", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestPublishInternalLiveRejectsInvalidUserFractionBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bundlePath := writeRootTestFile(t, "app-release.aab")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"publish",
		"internal",
		"--package",
		"com.example.app",
		"--aab",
		bundlePath,
		"--user-fraction",
		"0.25",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected user fraction validation error")
	}
	if !strings.Contains(err.Error(), "user fraction can only be set") {
		t.Fatalf("error = %v, want user fraction validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesUploadRejectsMalformedReleaseNoteBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		"app-release.aab",
		"--release-note",
		"en-US",
		"--dry-run",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected release note validation error")
	}
	if !strings.Contains(err.Error(), "language=text") {
		t.Fatalf("error = %v, want release note format validation", err)
	}
}

func TestReleasesUploadDryRunUsesRequestedTrack(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--track",
		"beta",
		"--aab",
		"app-release.aab",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"track":"beta"`) {
		t.Fatalf("release upload dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"status":"completed"`) {
		t.Fatalf("release upload dry-run output = %s", buf.String())
	}
}

func TestReleasesUploadDryRunSupportsAPK(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--track",
		"internal",
		"--apk",
		"app-release.apk",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"artifact":"apk"`) {
		t.Fatalf("release upload dry-run output = %s, want APK artifact", output)
	}
	if !strings.Contains(output, `"apkPath":"app-release.apk"`) {
		t.Fatalf("release upload dry-run output = %s, want APK path", output)
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestReleasesUploadRejectsMultipleArtifactsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--apk",
		"app-release.apk",
		"--aab",
		"app-release.aab",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected multiple artifact validation error")
	}
	if !strings.Contains(err.Error(), "exactly one of APK path or AAB path") {
		t.Fatalf("error = %v, want artifact validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesUploadLiveRejectsInvalidUserFractionBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	bundlePath := writeRootTestFile(t, "app-release.aab")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"upload",
		"--package",
		"com.example.app",
		"--aab",
		bundlePath,
		"--user-fraction",
		"0.25",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected user fraction validation error")
	}
	if !strings.Contains(err.Error(), "user fraction can only be set") {
		t.Fatalf("error = %v, want user fraction validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReleasesPromoteDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"promote",
		"--package",
		"com.example.app",
		"--from",
		"internal",
		"--to",
		"production",
		"--version-code",
		"42",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"toTrack":"production"`) {
		t.Fatalf("release promote dry-run output = %s", buf.String())
	}
}

func TestReleasesHaltDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"halt",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"halt"`) {
		t.Fatalf("release halt dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"inProgress",
		"--user-fraction",
		"0.25",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"action":"resume"`) {
		t.Fatalf("release resume dry-run output = %s", buf.String())
	}
}

func TestReleasesResumeCompletedDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"releases",
		"resume",
		"--package",
		"com.example.app",
		"--track",
		"production",
		"--version-code",
		"42",
		"--status",
		"completed",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"status":"completed"`) {
		t.Fatalf("release resume completed dry-run output = %s", buf.String())
	}
}

func TestListingsUpdateDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"update",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--title",
		"Example",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"language":"en-US"`) {
		t.Fatalf("listing update dry-run output = %s", buf.String())
	}
}

func TestListingsDeleteDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"delete",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"language":"en-US"`) {
		t.Fatalf("listing delete dry-run output = %s", buf.String())
	}
}

func TestListingsDeleteAllDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"listings",
		"delete-all",
		"--package",
		"com.example.app",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"all":true`) {
		t.Fatalf("listing delete-all dry-run output = %s", buf.String())
	}
}

func TestImagesListRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"images",
		"list",
		"--package",
		"com.example.app",
		"--language",
		"en-US",
		"--type",
		"poster",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected image type validation error")
	}
	if !strings.Contains(err.Error(), "unsupported image type") {
		t.Fatalf("error = %v, want image type validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestDetailsUpdateDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"details",
		"update",
		"--package",
		"com.example.app",
		"--contact-email",
		"support@example.com",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"contactEmail":"support@example.com"`) {
		t.Fatalf("details update dry-run output = %s", buf.String())
	}
}

func TestDataSafetyUpdateDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")
	csvPath := writeRootTestFile(t, "data-safety.csv")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"data-safety",
		"update",
		"--package",
		"com.example.app",
		"--csv",
		csvPath,
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("output = %s, want dryRun true", buf.String())
	}
	if strings.Contains(buf.String(), "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", buf.String())
	}
}

func TestDataSafetyUpdateRequiresConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"data-safety",
		"update",
		"--package",
		"com.example.app",
		"--csv",
		t.TempDir() + "/missing.csv",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
	if !strings.Contains(err.Error(), "--confirm or --dry-run") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func TestReviewsReplyDryRun(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"reply",
		"--package",
		"com.example.app",
		"--review-id",
		"review-123",
		"--text",
		"Thanks for trying the app.",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"reviewId":"review-123"`) {
		t.Fatalf("reviews reply dry-run output = %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"dryRun":true`) {
		t.Fatalf("reviews reply dry-run output = %s", buf.String())
	}
}

func TestReviewsListRejectsInvalidMaxResultsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"list",
		"--package",
		"com.example.app",
		"--max-results",
		"101",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected max results validation error")
	}
	if !strings.Contains(err.Error(), "max results") {
		t.Fatalf("error = %v, want max results validation", err)
	}
}

func TestReviewsReplyRequiresDryRunOrConfirmBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"reviews",
		"reply",
		"--package",
		"com.example.app",
		"--review-id",
		"review-123",
		"--text",
		"Thanks.",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestInAppProductsGetRejectsMissingSKUBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"in-app-products",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
	if !strings.Contains(err.Error(), "SKU") {
		t.Fatalf("error = %v, want SKU validation", err)
	}
}

func TestSubscriptionsListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"list",
		"--package",
		"com.example.app",
		"--page-size",
		"1001",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestSubscriptionsGetRejectsInvalidProductIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscriptions",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"Premium",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription product ID") {
		t.Fatalf("error = %v, want product ID validation", err)
	}
}

func TestSubscriptionOffersListRejectsInvalidPageSizeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--page-size",
		"1001",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected page size validation error")
	}
	if !strings.Contains(err.Error(), "page size") {
		t.Fatalf("error = %v, want page size validation", err)
	}
}

func TestSubscriptionOffersGetRejectsMissingOfferIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"monthly",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected offer ID validation error")
	}
	if !strings.Contains(err.Error(), "subscription offer ID") {
		t.Fatalf("error = %v, want offer ID validation", err)
	}
}

func TestSubscriptionOffersListAcceptsWildcardsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"list",
		"--package",
		"com.example.app",
		"--product-id",
		"-",
		"--base-plan-id",
		"-",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected auth error after wildcard validation succeeds")
	}
	if strings.Contains(err.Error(), "invalid subscription product ID") || strings.Contains(err.Error(), "base plan") {
		t.Fatalf("error = %v, want auth error after wildcard validation", err)
	}
}

func TestSubscriptionOffersGetRejectsWildcardBasePlanBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"subscription-offers",
		"get",
		"--package",
		"com.example.app",
		"--product-id",
		"premium",
		"--base-plan-id",
		"-",
		"--offer-id",
		"intro",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected base plan validation error")
	}
	if !strings.Contains(err.Error(), "subscription base plan ID") {
		t.Fatalf("error = %v, want base plan validation", err)
	}
}

func TestPurchasesProductRejectsMissingTokenBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"--package",
		"com.example.app",
		"--product-id",
		"coins_100",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token validation error")
	}
	if !strings.Contains(err.Error(), "purchase token") {
		t.Fatalf("error = %v, want token validation", err)
	}
}

func TestPurchasesProductAllowsTokenOnlyBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"product",
		"--package",
		"com.example.app",
		"--token",
		"token-123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected auth error")
	}
	if strings.Contains(err.Error(), "product ID") || strings.Contains(err.Error(), "in-app product") {
		t.Fatalf("error = %v, want auth error after token-only validation", err)
	}
}

func TestPurchasesVoidedListRejectsNegativeMaxResultsBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--max-results",
		"-1",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected max results validation error")
	}
	if !strings.Contains(err.Error(), "max results") {
		t.Fatalf("error = %v, want max results validation", err)
	}
}

func TestPurchasesVoidedListRejectsInvalidTypeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--type",
		"2",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected type validation error")
	}
	if !strings.Contains(err.Error(), "voided purchase type") {
		t.Fatalf("error = %v, want type validation", err)
	}
}

func TestPurchasesVoidedListRejectsTokenWithTimeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--token",
		"page",
		"--start-time",
		"1700000000000",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token/time validation error")
	}
	if !strings.Contains(err.Error(), "pagination token") {
		t.Fatalf("error = %v, want token/time validation", err)
	}
}

func TestPurchasesVoidedListRejectsFutureEndTimeBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"voided",
		"list",
		"--package",
		"com.example.app",
		"--end-time",
		"4102444800000",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected future end time validation error")
	}
	if !strings.Contains(err.Error(), "future") {
		t.Fatalf("error = %v, want future end time validation", err)
	}
}

func TestPurchasesSubscriptionRejectsMissingTokenBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"purchases",
		"subscription",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected token validation error")
	}
	if !strings.Contains(err.Error(), "purchase token") {
		t.Fatalf("error = %v, want token validation", err)
	}
}

func TestOrdersGetRejectsMissingOrderIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"get",
		"--package",
		"com.example.app",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected order ID validation error")
	}
	if !strings.Contains(err.Error(), "order ID") {
		t.Fatalf("error = %v, want order ID validation", err)
	}
}

func TestOrdersBatchGetRejectsDuplicateOrderIDBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"orders",
		"batch-get",
		"--package",
		"com.example.app",
		"--order-id",
		"GPA.123",
		"--order-id",
		"GPA.123",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected duplicate order ID validation error")
	}
	if !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("error = %v, want duplicate order ID validation", err)
	}
}

func TestPricingConvertRejectsInvalidPriceBeforeAuth(t *testing.T) {
	t.Setenv("GPC_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"pricing",
		"convert-region-prices",
		"--package",
		"com.example.app",
		"--currency",
		"USD",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected price validation error")
	}
	if !strings.Contains(err.Error(), "price must be greater than 0") {
		t.Fatalf("error = %v, want price validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}

func writeRootTestFile(t *testing.T, name string) string {
	t.Helper()
	path := t.TempDir() + "/" + name
	if err := os.WriteFile(path, []byte("artifact"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
