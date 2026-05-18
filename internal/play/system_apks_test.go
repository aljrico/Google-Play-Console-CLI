package play

import (
	"context"
	"testing"
)

func TestListSystemAPKVariantsPassesOptionsToLister(t *testing.T) {
	lister := &fakeSystemAPKVariantLister{result: SystemAPKVariantListResult{
		PackageName: "com.example.app",
		VersionCode: 42,
		Variants:    []SystemAPKVariant{{VariantID: 7}},
	}}
	options := SystemAPKVariantListOptions{PackageName: "com.example.app", VersionCode: 42}

	result, err := ListSystemAPKVariants(context.Background(), lister, options)
	if err != nil {
		t.Fatalf("ListSystemAPKVariants() error = %v", err)
	}
	if lister.options != options {
		t.Fatalf("options = %#v, want %#v", lister.options, options)
	}
	if len(result.Variants) != 1 || result.Variants[0].VariantID != 7 {
		t.Fatalf("variants = %#v, want variant 7", result.Variants)
	}
}

func TestListSystemAPKVariantsRejectsInvalidOptions(t *testing.T) {
	tests := []SystemAPKVariantListOptions{
		{},
		{PackageName: "bad", VersionCode: 42},
		{PackageName: "com.example.app"},
	}
	for _, options := range tests {
		_, err := ListSystemAPKVariants(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("ListSystemAPKVariants(%#v) expected validation error", options)
		}
	}
}

type fakeSystemAPKVariantLister struct {
	options SystemAPKVariantListOptions
	result  SystemAPKVariantListResult
}

func (l *fakeSystemAPKVariantLister) ListSystemAPKVariants(ctx context.Context, options SystemAPKVariantListOptions) (SystemAPKVariantListResult, error) {
	l.options = options
	return l.result, nil
}
