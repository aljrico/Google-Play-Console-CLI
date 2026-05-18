package play

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestListGeneratedAPKsPassesOptionsToLister(t *testing.T) {
	lister := &fakeGeneratedAPKLister{result: GeneratedAPKListResult{
		PackageName: "com.example.app",
		VersionCode: 42,
		SigningKeys: []GeneratedAPKSigningKey{
			{CertificateSHA256Hash: "abc", SplitAPKs: []GeneratedSplitAPK{}},
		},
	}}
	options := GeneratedAPKListOptions{PackageName: "com.example.app", VersionCode: 42}

	result, err := ListGeneratedAPKs(context.Background(), lister, options)
	if err != nil {
		t.Fatalf("ListGeneratedAPKs() error = %v", err)
	}
	if len(result.SigningKeys) != 1 {
		t.Fatalf("len(SigningKeys) = %d, want 1", len(result.SigningKeys))
	}
	if !reflect.DeepEqual(lister.options, options) {
		t.Fatalf("options = %#v, want %#v", lister.options, options)
	}
}

func TestListGeneratedAPKsRejectsInvalidOptions(t *testing.T) {
	tests := []GeneratedAPKListOptions{
		{},
		{PackageName: "bad", VersionCode: 42},
		{PackageName: "com.example.app"},
		{PackageName: "com.example.app", VersionCode: -1},
	}
	for _, options := range tests {
		_, err := ListGeneratedAPKs(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("ListGeneratedAPKs(%#v) expected validation error", options)
		}
	}
}

func TestGeneratedAPKJSONPreservesZeroIDs(t *testing.T) {
	payload, err := json.Marshal(GeneratedAPKSigningKey{
		SplitAPKs: []GeneratedSplitAPK{
			{DownloadID: "split-download", VariantID: 0},
		},
		StandaloneAPKs: []GeneratedStandaloneAPK{
			{DownloadID: "standalone-download", VariantID: 0},
		},
		AssetPackSlices: []GeneratedAssetPackSlice{
			{DownloadID: "asset-download", Version: 0},
		},
		RecoveryModules: []GeneratedRecoveryAPK{},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	output := string(payload)
	for _, want := range []string{`"variantId":0`, `"version":0`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}

type fakeGeneratedAPKLister struct {
	options GeneratedAPKListOptions
	result  GeneratedAPKListResult
}

func (l *fakeGeneratedAPKLister) ListGeneratedAPKs(ctx context.Context, options GeneratedAPKListOptions) (GeneratedAPKListResult, error) {
	l.options = options
	return l.result, nil
}
