package play

import (
	"context"
	"reflect"
	"testing"
)

func TestNewOneTimeProductIDValidatesGoogleShape(t *testing.T) {
	valid, err := NewOneTimeProductID("coins_100.v2")
	if err != nil {
		t.Fatalf("NewOneTimeProductID() error = %v", err)
	}
	if valid != "coins_100.v2" {
		t.Fatalf("product ID = %q, want coins_100.v2", valid)
	}

	for _, value := range []string{"", "Coins", "_coins", "coins-100"} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewOneTimeProductID(value); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestListOneTimeProductsPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeOneTimeProductClient{
		listResult: OneTimeProductListResult{
			PackageName: packageName,
			Products:    []OneTimeProduct{{ProductID: "coins_100"}},
		},
	}

	result, err := ListOneTimeProducts(context.Background(), lister, OneTimeProductListOptions{
		PackageName: packageName,
		PageSize:    50,
		PageToken:   "next",
	})
	if err != nil {
		t.Fatalf("ListOneTimeProducts() error = %v", err)
	}
	if len(result.Products) != 1 {
		t.Fatalf("len(Products) = %d, want 1", len(result.Products))
	}
	if !reflect.DeepEqual(lister.listOptions, OneTimeProductListOptions{PackageName: packageName, PageSize: 50, PageToken: "next"}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestListOneTimeProductsRejectsInvalidPageSize(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListOneTimeProducts(context.Background(), nil, OneTimeProductListOptions{
		PackageName: packageName,
		PageSize:    1001,
	})
	if err == nil {
		t.Fatal("expected page size validation error")
	}
}

func TestGetOneTimeProductPassesProductIDToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeOneTimeProductClient{
		product: OneTimeProduct{ProductID: "coins_100"},
	}

	product, err := GetOneTimeProduct(context.Background(), getter, OneTimeProductGetOptions{
		PackageName: packageName,
		ProductID:   "coins_100",
	})
	if err != nil {
		t.Fatalf("GetOneTimeProduct() error = %v", err)
	}
	if product.ProductID != "coins_100" {
		t.Fatalf("ProductID = %q, want coins_100", product.ProductID)
	}
	if getter.productID != "coins_100" {
		t.Fatalf("getter productID = %q, want coins_100", getter.productID)
	}
}

func TestGetOneTimeProductRejectsMissingProductID(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = GetOneTimeProduct(context.Background(), nil, OneTimeProductGetOptions{PackageName: packageName})
	if err == nil {
		t.Fatal("expected product ID validation error")
	}
}

func TestBatchGetOneTimeProductsPassesOptionsToGetter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeOneTimeProductClient{
		batchResult: OneTimeProductBatchGetResult{
			PackageName: packageName,
			Products:    []OneTimeProduct{{ProductID: "coins_100"}, {ProductID: "coins_500"}},
		},
	}
	options := OneTimeProductBatchGetOptions{
		PackageName: packageName,
		ProductIDs:  []OneTimeProductID{"coins_100", "coins_500"},
	}

	result, err := BatchGetOneTimeProducts(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("BatchGetOneTimeProducts() error = %v", err)
	}
	if len(result.Products) != 2 {
		t.Fatalf("len(Products) = %d, want 2", len(result.Products))
	}
	if !reflect.DeepEqual(getter.batchOptions, options) {
		t.Fatalf("batchOptions = %#v, want %#v", getter.batchOptions, options)
	}
}

func TestBatchGetOneTimeProductsRejectsDuplicates(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = BatchGetOneTimeProducts(context.Background(), nil, OneTimeProductBatchGetOptions{
		PackageName: packageName,
		ProductIDs:  []OneTimeProductID{"coins_100", "coins_100"},
	})
	if err == nil {
		t.Fatal("expected duplicate product ID validation error")
	}
}

func TestDeleteOneTimeProductDryRunBuildsPlanWithoutDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := DeleteOneTimeProduct(context.Background(), nil, OneTimeProductDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("DeleteOneTimeProduct() error = %v", err)
	}
	if !result.DryRun || result.Deleted {
		t.Fatalf("result = %#v, want dry-run deletion plan", result)
	}
	if !reflect.DeepEqual(result.Plan.Steps, []string{"delete one-time product"}) {
		t.Fatalf("steps = %#v, want delete step", result.Plan.Steps)
	}
}

func TestDeleteOneTimeProductRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = DeleteOneTimeProduct(context.Background(), nil, OneTimeProductDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestDeleteOneTimeProductPassesOptionsToDeleter(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	deleter := &fakeOneTimeProductClient{}

	result, err := DeleteOneTimeProduct(context.Background(), deleter, OneTimeProductDeleteOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("DeleteOneTimeProduct() error = %v", err)
	}
	if !result.Deleted {
		t.Fatal("Deleted = false, want true")
	}
	if deleter.deleteOptions.ProductID != "coins_100" {
		t.Fatalf("deleteOptions = %#v, want coins_100", deleter.deleteOptions)
	}
}

func TestUpdatePurchaseOptionStateDryRunBuildsPlanWithoutUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := UpdatePurchaseOptionState(context.Background(), nil, PurchaseOptionStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionActivate,
		LatencyTolerance: ProductUpdateLatencyToleranceTolerant,
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("UpdatePurchaseOptionState() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	wantSteps := []string{"plan activate purchase option"}
	if !reflect.DeepEqual(result.Plan.Steps, wantSteps) {
		t.Fatalf("steps = %#v, want %#v", result.Plan.Steps, wantSteps)
	}
}

func TestUpdatePurchaseOptionStateRequiresConfirmOrDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = NewPurchaseOptionStateUpdatePlan(PurchaseOptionStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
	})
	if err == nil {
		t.Fatal("expected confirm or dry-run validation error")
	}
}

func TestUpdatePurchaseOptionStatePassesOptionsToUpdater(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	updater := &fakeOneTimeProductClient{
		product: OneTimeProduct{ProductID: "coins_100"},
	}

	result, err := UpdatePurchaseOptionState(context.Background(), updater, PurchaseOptionStateUpdateOptions{
		PackageName:      packageName,
		ProductID:        "coins_100",
		PurchaseOptionID: "buy",
		Action:           PurchaseOptionStateActionDeactivate,
		LatencyTolerance: ProductUpdateLatencyToleranceSensitive,
		Confirm:          true,
	})
	if err != nil {
		t.Fatalf("UpdatePurchaseOptionState() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if updater.stateOptions.Action != PurchaseOptionStateActionDeactivate {
		t.Fatalf("Action = %q, want deactivate", updater.stateOptions.Action)
	}
}

type fakeOneTimeProductClient struct {
	listOptions   OneTimeProductListOptions
	listResult    OneTimeProductListResult
	batchOptions  OneTimeProductBatchGetOptions
	batchResult   OneTimeProductBatchGetResult
	deleteOptions OneTimeProductDeleteOptions
	productID     OneTimeProductID
	product       OneTimeProduct
	stateOptions  PurchaseOptionStateUpdateOptions
}

func (c *fakeOneTimeProductClient) ListOneTimeProducts(ctx context.Context, options OneTimeProductListOptions) (OneTimeProductListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeOneTimeProductClient) GetOneTimeProduct(ctx context.Context, packageName PackageName, productID OneTimeProductID) (OneTimeProduct, error) {
	c.productID = productID
	return c.product, nil
}

func (c *fakeOneTimeProductClient) BatchGetOneTimeProducts(ctx context.Context, options OneTimeProductBatchGetOptions) (OneTimeProductBatchGetResult, error) {
	c.batchOptions = options
	return c.batchResult, nil
}

func (c *fakeOneTimeProductClient) DeleteOneTimeProduct(ctx context.Context, options OneTimeProductDeleteOptions) error {
	c.deleteOptions = options
	return nil
}

func (c *fakeOneTimeProductClient) UpdatePurchaseOptionState(ctx context.Context, options PurchaseOptionStateUpdateOptions) (OneTimeProduct, error) {
	c.stateOptions = options
	return c.product, nil
}
