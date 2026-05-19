package play

import (
	"context"
	"reflect"
	"testing"
)

func TestListInAppProductsPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeInAppProductClient{
		listResult: InAppProductListResult{
			PackageName: packageName,
			Products:    []InAppProduct{{SKU: "coins_100"}},
		},
	}

	result, err := ListInAppProducts(context.Background(), lister, InAppProductListOptions{
		PackageName: packageName,
		Token:       "next",
	})
	if err != nil {
		t.Fatalf("ListInAppProducts() error = %v", err)
	}
	if len(result.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(result.Products))
	}
	if !reflect.DeepEqual(lister.listOptions, InAppProductListOptions{PackageName: packageName, Token: "next"}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestGetInAppProductPassesSKUToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeInAppProductClient{
		product: InAppProduct{SKU: "coins_100"},
	}

	product, err := GetInAppProduct(context.Background(), getter, InAppProductGetOptions{
		PackageName: packageName,
		SKU:         "coins_100",
	})
	if err != nil {
		t.Fatalf("GetInAppProduct() error = %v", err)
	}
	if product.SKU != "coins_100" {
		t.Fatalf("SKU = %q, want coins_100", product.SKU)
	}
	if getter.sku != "coins_100" {
		t.Fatalf("getter sku = %q, want coins_100", getter.sku)
	}
}

func TestGetInAppProductRejectsMissingSKU(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = GetInAppProduct(context.Background(), nil, InAppProductGetOptions{PackageName: packageName})
	if err == nil {
		t.Fatal("expected SKU validation error")
	}
}

func TestPatchInAppProductDryRunBuildsPlanWithoutPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatusActive,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if result.Applied || !result.DryRun || result.Desired.Status != ProductStatusActive {
		t.Fatalf("result = %#v, want dry-run active patch", result)
	}
}

func TestPatchInAppProductRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = PatchInAppProduct(context.Background(), nil, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatusActive,
	})
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestPatchInAppProductPassesOptionsToPatcher(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	patcher := &fakeInAppProductClient{product: InAppProduct{SKU: "coins_100", Status: ProductStatusInactive}}

	result, err := PatchInAppProduct(context.Background(), patcher, InAppProductPatchOptions{
		PackageName: packageName,
		SKU:         "coins_100",
		Status:      ProductStatusInactive,
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("PatchInAppProduct() error = %v", err)
	}
	if !result.Applied || result.Product == nil || result.Product.Status != ProductStatusInactive {
		t.Fatalf("result = %#v, want applied inactive product", result)
	}
	if patcher.patchOptions.Status != ProductStatusInactive {
		t.Fatalf("patchOptions = %#v", patcher.patchOptions)
	}
}

type fakeInAppProductClient struct {
	listOptions  InAppProductListOptions
	listResult   InAppProductListResult
	patchOptions InAppProductPatchOptions
	sku          InAppProductSKU
	product      InAppProduct
}

func (c *fakeInAppProductClient) ListInAppProducts(ctx context.Context, options InAppProductListOptions) (InAppProductListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeInAppProductClient) GetInAppProduct(ctx context.Context, packageName PackageName, sku InAppProductSKU) (InAppProduct, error) {
	c.sku = sku
	return c.product, nil
}

func (c *fakeInAppProductClient) PatchInAppProduct(ctx context.Context, options InAppProductPatchOptions) (InAppProduct, error) {
	c.patchOptions = options
	return c.product, nil
}
