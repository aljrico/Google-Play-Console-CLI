package play

import (
	"context"
	"reflect"
	"testing"
)

func TestListDeviceTierConfigsPassesOptionsToLister(t *testing.T) {
	lister := &fakeDeviceTierConfigLister{result: DeviceTierConfigListResult{
		PackageName: "com.example.app",
		Configs: []DeviceTierConfig{
			{ID: "7"},
		},
	}}
	options := DeviceTierConfigListOptions{
		PackageName: "com.example.app",
		PageSize:    25,
		PageToken:   "page",
	}

	result, err := ListDeviceTierConfigs(context.Background(), lister, options)
	if err != nil {
		t.Fatalf("ListDeviceTierConfigs() error = %v", err)
	}
	if len(result.Configs) != 1 {
		t.Fatalf("len(Configs) = %d, want 1", len(result.Configs))
	}
	if !reflect.DeepEqual(lister.options, options) {
		t.Fatalf("options = %#v, want %#v", lister.options, options)
	}
}

func TestGetDeviceTierConfigPassesOptionsToGetter(t *testing.T) {
	getter := &fakeDeviceTierConfigGetter{result: DeviceTierConfigGetResult{
		PackageName: "com.example.app",
		Config:      DeviceTierConfig{ID: "7"},
	}}
	options := DeviceTierConfigGetOptions{PackageName: "com.example.app", DeviceTierConfigID: 7}

	result, err := GetDeviceTierConfig(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("GetDeviceTierConfig() error = %v", err)
	}
	if result.Config.ID != "7" {
		t.Fatalf("ID = %q, want 7", result.Config.ID)
	}
	if !reflect.DeepEqual(getter.options, options) {
		t.Fatalf("options = %#v, want %#v", getter.options, options)
	}
}

func TestDeviceTierConfigOptionsRejectInvalidInputs(t *testing.T) {
	listTests := []DeviceTierConfigListOptions{
		{},
		{PackageName: "bad"},
		{PackageName: "com.example.app", PageSize: -1},
		{PackageName: "com.example.app", PageSize: 101},
	}
	for _, options := range listTests {
		_, err := ListDeviceTierConfigs(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("ListDeviceTierConfigs(%#v) expected validation error", options)
		}
	}

	getTests := []DeviceTierConfigGetOptions{
		{},
		{PackageName: "bad", DeviceTierConfigID: 7},
		{PackageName: "com.example.app"},
	}
	for _, options := range getTests {
		_, err := GetDeviceTierConfig(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("GetDeviceTierConfig(%#v) expected validation error", options)
		}
	}
}

type fakeDeviceTierConfigLister struct {
	options DeviceTierConfigListOptions
	result  DeviceTierConfigListResult
}

func (l *fakeDeviceTierConfigLister) ListDeviceTierConfigs(ctx context.Context, options DeviceTierConfigListOptions) (DeviceTierConfigListResult, error) {
	l.options = options
	return l.result, nil
}

type fakeDeviceTierConfigGetter struct {
	options DeviceTierConfigGetOptions
	result  DeviceTierConfigGetResult
}

func (g *fakeDeviceTierConfigGetter) GetDeviceTierConfig(ctx context.Context, options DeviceTierConfigGetOptions) (DeviceTierConfigGetResult, error) {
	g.options = options
	return g.result, nil
}
