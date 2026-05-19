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

type fakeOneTimeProductClient struct {
	listOptions OneTimeProductListOptions
	listResult  OneTimeProductListResult
	productID   OneTimeProductID
	product     OneTimeProduct
}

func (c *fakeOneTimeProductClient) ListOneTimeProducts(ctx context.Context, options OneTimeProductListOptions) (OneTimeProductListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeOneTimeProductClient) GetOneTimeProduct(ctx context.Context, packageName PackageName, productID OneTimeProductID) (OneTimeProduct, error) {
	c.productID = productID
	return c.product, nil
}
